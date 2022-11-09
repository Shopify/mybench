package mybench

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// To test the looper, we need a way to reliably cause a certain amount of time
// to pass during a single Event() call. We do not want to use time.Sleep,
// because time.Sleep may put the goroutine to sleep and schedule another
// goroutine. Once the sleep time is done, it then tries to wake up the sleeping
// goroutine. This can introduce significant jitter, especially when the sleep
// duration is low. This jitter may result in unpredicable behavior in test (or
// significant variance in the results), which makes it hard to test without
// being flaky.
//
// Instead, we use a busy sleeper by computing a prime number inefficiently.
// The code starts by calibrating the prime number function and compute a linear
// function (via linear regression) that maps execution time and the number to
// be checked as prime. This is known as the BusySleeper. The BusySleeper.Sleep
// takes a time to sleep, calculates the required value and then call the prime
// number function.
//
// In my limited experimentation, this approach results in low jitter (+/- 30us)
// on a virtual machine.
type BusySleeper struct {
	// v = beta * time + alpha
	alpha float64
	beta  float64
	r2    float64

	maxSleepTimeUs float64
	foundPrime     int64 // Hopefully this means the computePrime function won't be optimized out
}

func NewBusySleeper(maxSleepTimeUs float64) *BusySleeper {
	s := &BusySleeper{
		maxSleepTimeUs: maxSleepTimeUs,
	}

	s.calibrate()
	return s
}

func (s *BusySleeper) Sleep(timeUs float64) {
	sleepTimeRemaining := timeUs
	for sleepTimeRemaining > 0 {
		timeToSleep := sleepTimeRemaining
		if timeToSleep > s.maxSleepTimeUs {
			timeToSleep = s.maxSleepTimeUs
		}

		v := math.Round(s.beta*timeToSleep + s.alpha)
		s.computePrime(int64(v))

		sleepTimeRemaining -= timeToSleep
	}

}

// Very inefficient prime computation to burn time
func (s *BusySleeper) computePrime(v int64) {
	var isPrime bool = true
	for i := int64(1); i < v; i++ {
		if v%i == 0 {
			isPrime = false
		}
	}

	if isPrime {
		s.foundPrime += 1
	}
}

func (s *BusySleeper) calibrate() {
	var currentV int64 = 0
	var currentTimingUs int64 = 0

	diff := int64(200000)

	timings := [][2]int64{}

	for currentTimingUs < int64(s.maxSleepTimeUs) {
		currentV += diff
		start := time.Now()
		s.computePrime(currentV)
		currentTimingUs = time.Since(start).Microseconds()

		row := [2]int64{currentTimingUs, currentV}
		timings = append(timings, row)
		// fmt.Printf("%d, %d\n", row[0], row[1])
	}

	s.alpha, s.beta, s.r2 = simpleLinearRegression(timings)
	logrus.Debugf("v = %.3E * t + %.3f (r^2 = %.6f, n = %d)", s.beta, s.alpha, s.r2, len(timings))
	logrus.Debugf("Ignore me: %d", s.foundPrime) // Hopefully means that computePrime won't be optimized out.
}

// y = alpha + beta * x
func simpleLinearRegression(data [][2]int64) (float64, float64, float64) {
	xbar := 0.0
	ybar := 0.0
	xybar := 0.0
	x2bar := 0.0
	y2bar := 0.0

	for _, row := range data {
		xbar += float64(row[0])
		ybar += float64(row[1])
		xybar += float64(row[0] * row[1])
		x2bar += float64(row[0] * row[0])
		y2bar += float64(row[1] * row[1])
	}

	n := float64(len(data))
	xbar /= n
	ybar /= n
	xybar /= n
	x2bar /= n
	y2bar /= n

	top := 0.0
	bottom := 0.0
	for _, row := range data {
		x := float64(row[0])
		y := float64(row[1])
		top += (x - xbar) * (y - ybar)
		bottom += (x - xbar) * (x - xbar)
	}

	beta := top / bottom
	alpha := ybar - (beta * xbar)

	r := (xybar - xbar*ybar) / math.Sqrt((x2bar-xbar*xbar)*(y2bar-ybar*ybar))

	return alpha, beta, r * r
}

var sleeper *BusySleeper = NewBusySleeper(25000)

func TestLooperNoBackPressure(t *testing.T) {
	numEvents := int64(1000) // 5 seconds

	looper := &DiscretizedLooper{
		EventRate:     200,
		OuterLoopRate: 50,
		LooperType:    LooperTypeUniform,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Don't ant to allocate in the loop, so we allocate plenty of space
	stats := make([]OuterLoopStat, 0, int(looper.OuterLoopRate*float64(numEvents)/looper.EventRate*10))

	looper.Event = func() error {
		sleeper.Sleep(1000)
		return nil
	}

	looper.TraceOuterLoop = func(stat OuterLoopStat) {
		stats = append(stats, stat)
		if stat.CumulativeNumberOfEvents >= numEvents {
			cancel()
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	start := time.Now()
	go func() {
		defer wg.Done()
		looper.Run(ctx)
	}()
	wg.Wait()
	delta := time.Since(start).Microseconds()

	expectedDelta := int64(5000000) // expected 5 seconds
	diff := float64(expectedDelta - delta)
	percent := math.Abs(diff / float64(expectedDelta) * 100.0)

	// The threshold is arbitrarily chosen. Works on my computer.
	require.True(t, percent <= 5, "expected total looper duration is more than 5%% from the actual duration (expected = %d us, actual = %d us, diff = %.2f%%).", expectedDelta, delta, percent)

	lastStat := stats[len(stats)-1]
	require.Equal(t, int64(1000), lastStat.CumulativeNumberOfEvents)

	for _, stat := range stats {
		// Since the EventRate is a multiple of OuterLoopRate, the expected batch
		// size should be constant as there should be no discretization problems.
		require.Equal(t, int64(4), stat.EventBatchSize)
	}

	// Calculate the event rate per segment
	numSegments := 5
	width := len(stats) / numSegments
	// Should have 50 events per segement as that's the outer loop rate
	require.Equal(t, 50, width)

	for i := 0; i < numSegments; i++ {
		stat1 := stats[i*width]
		stat2 := stats[(i+1)*width-1]

		numEvents := stat2.CumulativeNumberOfEvents - stat1.CumulativeNumberOfEvents
		deltaT := stat2.EventsEnd.Sub(stat1.ActualWakeupTime)
		rate := float64(numEvents) / deltaT.Seconds()
		diffPct := math.Abs((rate - looper.EventRate) / looper.EventRate * 100.0)

		require.True(t, diffPct <= 5, "looper rate deviated more than 5%% from the expected (expected = %.1f Hz, actual = %.2f Hz, diff = %.2f%%)", looper.EventRate, rate, diffPct)
	}
}

func TestLooperSignificantBackPressure(t *testing.T) {
	// 5 seconds, 100 events, at a desired rate of 200 means it won't work.
	numEvents := int64(100)

	looper := &DiscretizedLooper{
		EventRate:     200,
		OuterLoopRate: 50,
		LooperType:    LooperTypeUniform,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Don't ant to allocate in the loop, so we allocate plenty of space
	stats := make([]OuterLoopStat, 0, int(looper.OuterLoopRate*float64(numEvents)/looper.EventRate*10))

	looper.Event = func() error {
		// Want 5 seconds with 100 events means 20 events per seconds or 1/20
		// seconds per event This means there's significant back pressure and the
		// looper should switch to the busy loop style, which means the final loop
		// rate should be close to 20Hz.
		sleeper.Sleep(1.0 / 20.0 * 1000000)
		return nil
	}

	looper.TraceOuterLoop = func(stat OuterLoopStat) {
		stats = append(stats, stat)
		if stat.CumulativeNumberOfEvents >= numEvents {
			cancel()
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	start := time.Now()
	go func() {
		defer wg.Done()
		looper.Run(ctx)
	}()
	wg.Wait()
	delta := time.Since(start).Microseconds()

	expectedDelta := int64(5000000) // expected 5 seconds
	diff := float64(expectedDelta - delta)
	percent := math.Abs(diff / float64(expectedDelta) * 100.0)

	// The threshold is arbitrarily chosen. Works on my computer.
	require.True(t, percent <= 5, "expected total looper duration is more than 5%% from the actual duration (expected = %d us, actual = %d us, diff = %.2f%%).", expectedDelta, delta, percent)

	lastStat := stats[len(stats)-1]
	require.Equal(t, int64(100), lastStat.CumulativeNumberOfEvents)

	for i, stat := range stats {
		if i == 0 {
			// Should always be executing only 1 batch, except the first one as the
			// system kind of works out the timing.
			require.Equal(t, int64(4), stat.EventBatchSize)
		} else {
			require.Equal(t, int64(1), stat.EventBatchSize)

			// Check that sleep hasn't occured.
			lastStat := stats[i-1]
			delta := stat.ActualWakeupTime.Sub(lastStat.EventsEnd)

			// Also just picked a random threshold value that worked for me
			require.True(t, delta.Microseconds() < 500, "the loop shouldn't be sleeping at all, but seemed to have waited %d us", delta.Microseconds())
		}
	}

	// Calculate the event rate per segment
	numSegments := 5
	width := len(stats) / numSegments
	require.True(t, width >= 19)
	require.True(t, width <= 21)

	for i := 0; i < numSegments; i++ {
		stat1 := stats[i*width]
		stat2 := stats[(i+1)*width-1]

		numEvents := stat2.CumulativeNumberOfEvents - stat1.CumulativeNumberOfEvents
		deltaT := stat2.EventsEnd.Sub(stat1.ActualWakeupTime)
		rate := float64(numEvents) / deltaT.Seconds()
		expectedRate := 20.0
		diffPct := math.Abs((rate - expectedRate) / expectedRate * 100.0)

		if i == 0 {
			continue // The first data point will be messed up due to how the looper assumes no back pressure to start.
		}

		require.True(t, diffPct <= 8, "looper rate deviated more than 8%% from the expected (expected = %.1f Hz, actual = %.2f Hz, diff = %.2f%%)", looper.EventRate, rate, diffPct)
	}
}

func TestLooperVaryingEventDurationNoBackPressure(t *testing.T) {
	// TODO
	t.Skip()
}

func TestLooperVaryingEventDurationWithSometimesBackPressure(t *testing.T) {
	// TODO
	t.Skip()
}
