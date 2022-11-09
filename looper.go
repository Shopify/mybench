package mybench

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type OuterLoopStat struct {
	// The desired wake up time.
	DesiredWakeupTime time.Time

	// The start time of the outer loop iteration and the actual wake up time.
	ActualWakeupTime time.Time

	// Number of Event() calls to make in this outer loop iteration.
	EventBatchSize int64

	// The end time of the outer loop iteration, but doesn't actually include the
	// time taken to calculate the next wakeup time, as that code is assumed to be
	// very fast and negligible.
	EventsEnd time.Time

	// The total time taken to process all Event() calls, including the TraceEvent calls.
	EventsLatency time.Duration

	// The next desired wakeup time
	NextDesiredWakeupTime time.Time

	// The next expected event's activation time. If this timestamp is really
	// behind the actual wakeup time, then the system is very backlogged.
	NextExpectedEventTime time.Time

	// The cumulative number of events executed.
	CumulativeNumberOfEvents int64
}

type EventStat struct {
	TimeTaken time.Duration
}

type LooperType int

const (
	LooperTypeUniform LooperType = iota
	LooperTypePoisson
)

type DiscretizedLooper struct {
	EventRate     float64
	OuterLoopRate float64
	LooperType    LooperType

	Event          func() error
	TraceEvent     func(EventStat)
	TraceOuterLoop func(OuterLoopStat)

	startTime time.Time
}

func (l *DiscretizedLooper) Run(ctx context.Context) error {
	l.startTime = time.Now()

	nextExpectedEventTime := l.startTime
	nextWakeupTime := l.startTime
	eventBatchSize := int64(0)
	executedNumberOfEvents := int64(0)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		lastWakeupTime := nextWakeupTime
		actualWakeupTime := time.Now()

		// Calculate the event batch size for this period
		// ==============================================
		//
		// Each outer loop period contain enough time for the event to trigger N
		// times. The code calculates value of N:

		// First, calculate the next wakeup time based on the outer loop rate. There
		// are a few possibilities:
		//
		// 1. The current time is actually already past the next wakeup time. This
		//    means the loop is very behind. N = 1 in this case. At the end of the
		//    loop, no sleep is needed as we need to catch up as fast as possible
		//    (may not be possible at all, but the system should try nonetheless).
		// 2. The nextExpectedEventTime is actually in the next period.
		// 3. The current time is before the next wakeup time, which means we can
		//    try to compute N. N is computed by simulating the events forward,
		//    until the nextExpectedEventTime is past the nextWakeupTime.
		//    -  There is a special case where the nextExpectedEventTime is in the
		//       past (possibly because the loop is falling behind). In this case,
		//       we are still going to generate a large N. If that causes the loop
		//       to fall further behind, the next iterations will eventually lead to
		//       case 1.
		nextWakeupTime = nextWakeupTime.Add(time.Duration(1.0 / l.OuterLoopRate * 1000000000))
		if actualWakeupTime.Sub(nextWakeupTime) >= 0 {
			// This case is when the looper is really behind, as the next wake up time
			// is behind the current actual wake up time.. so we just execute as
			// fast as possible until we catch up (which might be never). The next
			// wake up time is set to the next event time, which _should_ be in the
			// past, which means no sleeping (unless the Poisson distribution sample
			// something significantly in the future, which then it sleeps)
			eventBatchSize = 1
			nextWakeupTime = nextExpectedEventTime
			nextExpectedEventTime = nextExpectedEventTime.Add(l.interArrivalDuration())
		} else if nextExpectedEventTime.Sub(nextWakeupTime) >= 0 {
			// In this case, the next expected event time is in the next period, so we
			// just skip this period without executing any events.
			eventBatchSize = 0
		} else {
			// nextExpectedEventTime - lastWakeupTime - actualWakeupTime - nextWakeupTime
			// lastWakeupTime - nextExpectedEventTime - actualWakeupTime - nextWakeupTime
			// lastWakeupTime - actualWakeupTime - nextExpectedEventTime - nextWakeupTime
			eventBatchSize = 0 // TODO: Should this be 0 or 1? Is there
			for nextExpectedEventTime.Sub(nextWakeupTime) < 0 {
				nextExpectedEventTime = nextExpectedEventTime.Add(l.interArrivalDuration())
				eventBatchSize++
			}
		}

		// Running the events
		// ==================
		//
		// In an naive implementation, you want to call Event on every loop
		// iteration and wake up at some point later as dictated by the desired loop
		// rate. This does not work because Golang (and non-real-time Linux) cannot
		// keep a constant loop at rates above 100Hz consistently (in other words,
		// requesting for a sleep of 1ms does not mean the goroutine will sleep 1ms.
		// It will usually sleep a lot more than 1ms).
		//
		// Instead, the idea is to batch multiple Event() calls in a single loop
		// iteration. Since there's no sleep in an Event, then the desired rate
		// higher than 100Hz can be achieved. Effectively, this discretizes time
		// into windows due to the limitations imposed by Golang and Linux.
		for i := int64(0); i < eventBatchSize; i++ {
			var eventStart time.Time
			if l.TraceEvent != nil {
				eventStart = time.Now()
			}

			err := l.Event()
			if err != nil {
				return err
			}

			if l.TraceEvent != nil {
				l.TraceEvent(EventStat{
					TimeTaken: time.Since(eventStart),
				})
			}
		}
		executedNumberOfEvents += eventBatchSize
		eventsEnd := time.Now()

		if l.TraceOuterLoop != nil {
			outerLoopStat := OuterLoopStat{
				DesiredWakeupTime:        lastWakeupTime,
				ActualWakeupTime:         actualWakeupTime,
				EventBatchSize:           eventBatchSize,
				EventsEnd:                eventsEnd,
				EventsLatency:            eventsEnd.Sub(actualWakeupTime),
				NextDesiredWakeupTime:    nextWakeupTime,
				NextExpectedEventTime:    nextExpectedEventTime,
				CumulativeNumberOfEvents: executedNumberOfEvents,
			}
			l.TraceOuterLoop(outerLoopStat)
		}

		now := time.Now()
		delta := nextWakeupTime.Sub(now)
		if delta > time.Duration(0) { // May want to have a positive threshold, but requires looking at the top logic calculating the batch size
			time.Sleep(delta)
		}
	}
}

func (l *DiscretizedLooper) interArrivalDuration() time.Duration {
	switch l.LooperType {
	case LooperTypeUniform:
		return l.interArrivalDurationUniform()
	case LooperTypePoisson:
		return l.interArrivalDurationPoisson()
	}

	panic("not implemented")
}

func (l *DiscretizedLooper) interArrivalDurationUniform() time.Duration {
	return time.Duration(1.0 / l.EventRate * 1000000000)
}

func (l *DiscretizedLooper) interArrivalDurationPoisson() time.Duration {
	return time.Duration(-math.Log(1-rand.Float64()) / l.EventRate * 1000000000)
}
