package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/mybench"
)

const loopDuration = 30
const outerLoopRate = 50

var foundPrime = 0

// Very inefficient prime computation to burn time
func computePrime(v int64) {
	var isPrime bool = true
	for i := int64(1); i < v; i++ {
		if v%i == 0 {
			isPrime = false
		}
	}

	if isPrime {
		foundPrime += 1
	}
}

func runLoop(v int64) []mybench.OuterLoopStat {
	looper := &mybench.DiscretizedLooper{
		EventRate:     1001,
		OuterLoopRate: outerLoopRate,
		LooperType:    mybench.LooperTypePoisson,
	}

	// Should have at most loopDuration * outerLoopRate data points if the loop is
	// executing normally. However, if it is falling behind, the looper will busy
	// loop which will significantly increase the size of the logs.
	stats := make([]mybench.OuterLoopStat, 0, loopDuration*outerLoopRate*50)

	looper.Event = func() error {
		computePrime(v)
		return nil
	}

	looper.TraceOuterLoop = func(stat mybench.OuterLoopStat) {
		stats = append(stats, stat)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer wg.Done()
		looper.Run(ctx)
	}()

	time.Sleep(loopDuration * time.Second)
	cancel()
	wg.Wait()

	return stats
}

func main() {
	// On my test setup:
	// prime = 10000000 -> 29.6ms (33.62Hz)
	// prime = 1000000 -> 2.97ms (336.03Hz)
	// prime = 340000 -> 1.04ms (960Hz)
	// prime = 100000 -> 0.3ms (3.3kHz)
	timing := flag.Int64("timing", 0, "check the timing of a particular prime number on your computer")
	prime := flag.Int64("prime", 1337, "prime number to compute")
	flag.Parse()

	if *timing != int64(0) {
		attempts := 10
		total := time.Duration(0)
		for i := 0; i < attempts; i++ {
			now := time.Now()
			computePrime(*timing)
			duration := time.Since(now)
			total += duration
			fmt.Printf("Try %d: %v\n", i, duration)
		}

		total = total / time.Duration(attempts)
		fmt.Printf("Average: %v | Max rate: %.2f Hz\n", total, 1.0/total.Seconds())
	} else {
		stats := runLoop(*prime)
		for i, stat := range stats {
			fmt.Printf(
				"%d,%d,%d,%d,%d,%d,%d,%d,%d\n",
				i+1,
				stat.DesiredWakeupTime.UnixMicro(),
				stat.ActualWakeupTime.UnixMicro(),
				stat.EventBatchSize,
				stat.EventsEnd.UnixMicro(),
				stat.EventsLatency.Microseconds(),
				stat.NextDesiredWakeupTime.UnixMicro(),
				stat.NextExpectedEventTime.UnixMicro(),
				stat.CumulativeNumberOfEvents,
			)
		}
	}
}
