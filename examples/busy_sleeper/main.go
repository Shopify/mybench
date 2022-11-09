package main

import (
	"fmt"
	"math"
	"time"

	"github.com/sirupsen/logrus"
)

// Test program for the busy sleeper that tests the looper

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

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	s := NewBusySleeper(25000)

	start := time.Now()
	s.Sleep(200)
	fmt.Println(time.Since(start).Microseconds())

	start = time.Now()
	s.Sleep(50000)
	fmt.Println(time.Since(start).Microseconds())

	start = time.Now()
	s.Sleep(220000)
	fmt.Println(time.Since(start).Microseconds())
}
