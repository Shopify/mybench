package main

import (
	"runtime"
	"time"

	"github.com/Shopify/mybench"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/constraints"
)

type SelfBenchConfig struct {
	*mybench.BenchmarkAppConfig
	EventRatePerWorker float64
	Concurrency        int
	OuterLoopRate      float64
	SpinTime           time.Duration
}

type SelfBench struct {
	mybench.WorkloadConfig
	config SelfBenchConfig
}

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func NewSelfBench(config SelfBenchConfig) mybench.AbstractWorkload {
	// TODO: make these configurable:
	latencyHistMin := max(0, config.SpinTime.Microseconds()-1000)
	latencyHistMax := config.SpinTime.Microseconds() + 1000

	maxEventRatePerWorker := 1.0 / config.SpinTime.Seconds()
	maxTotalEventRate := maxEventRatePerWorker * float64(runtime.NumCPU())

	logrus.Warnf("With a spin time of %v, the max event rate per worker is %v ev/s, the max total event rate is %v ev/s", config.SpinTime, maxEventRatePerWorker, maxTotalEventRate)

	if config.EventRatePerWorker > maxEventRatePerWorker {
		logrus.Errorf("expected event rate per worker (%v ev/s) is higher than the theoretical maximum possible event rate (%v ev/s), suboptimal results are expected", config.EventRatePerWorker, maxEventRatePerWorker)
	}

	totalEventRate := config.EventRatePerWorker * float64(config.Concurrency)
	if totalEventRate > maxTotalEventRate {
		logrus.Errorf("expected total event rate (%v ev/s) is higher than the theoretical maximum possible event rate (%v ev/s), suboptimal results are expected", totalEventRate, maxTotalEventRate)
	}

	logrus.Warnf(
		"minimum CPU utilization is expected to be %.2f%% of all CPUs or %.2f cores (%v/%v ev/s)",
		totalEventRate/maxTotalEventRate*100.0,
		totalEventRate/maxTotalEventRate*float64(runtime.NumCPU()),
		totalEventRate,
		maxTotalEventRate,
	)

	var workloadInterface mybench.WorkloadInterface[mybench.NoContextData] = &SelfBench{
		WorkloadConfig: mybench.NewWorkloadConfigWithDefaults(mybench.WorkloadConfig{
			Name:           "SelfBench",
			DatabaseConfig: config.DatabaseConfig,
			RateControl: mybench.RateControlConfig{
				EventRate:     totalEventRate,
				Concurrency:   config.Concurrency,
				OuterLoopRate: config.OuterLoopRate,
			},
			Visualization: mybench.VisualizationConfig{
				LatencyHistMin: latencyHistMin,
				LatencyHistMax: latencyHistMax,
			},
		}),
		config: config,
	}

	workload, err := mybench.NewWorkload(workloadInterface)
	if err != nil {
		panic(err)
	}

	return workload

}

func (r *SelfBench) Config() mybench.WorkloadConfig {
	return r.WorkloadConfig
}

func (r *SelfBench) Event(ctx mybench.WorkerContext[mybench.NoContextData]) error {
	timeWaster.Spin(r.config.SpinTime)
	return nil
}

func (r *SelfBench) NewContextData(conn *mybench.Connection) (mybench.NoContextData, error) {
	return mybench.NewNoContextData()
}
