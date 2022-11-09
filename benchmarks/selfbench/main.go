package main

import (
	"flag"
	"time"

	"github.com/Shopify/mybench"
)

func main() {
	config := SelfBenchConfig{
		BenchmarkAppConfig: mybench.NewBenchmarkAppConfig(),
	}
	// This benchmark does not need a connection
	// Kind of a hack...
	config.BenchmarkAppConfig.DatabaseConfig.NoConnection = true
	config.BenchmarkAppConfig.DatabaseConfig.Host = "no-host"

	flag.IntVar(&config.Concurrency, "concurrency", 1, "the number of workers to use")
	flag.Float64Var(&config.EventRatePerWorker, "event-rate-per-worker", 100.0, "the number of events per second per worker")
	flag.Float64Var(&config.OuterLoopRate, "outer-loop-rate", 50.0, "the outer loop rate of the DiscretizedLooper")
	flag.DurationVar(&config.SpinTime, "spin-time", 5*time.Millisecond, "the amount of time to waste in each Event() call")
	flag.Parse()

	app, err := mybench.NewBenchmarkApp("SelfBench", config, setupBenchmark, runLoader)
	if err != nil {
		panic(err)
	}

	err = app.Run()
	if err != nil {
		panic(err)
	}
}

func runLoader(app *mybench.BenchmarkApp[SelfBenchConfig]) error {
	// Not needed for this benchmark
	return nil
}

func setupBenchmark(app *mybench.BenchmarkApp[SelfBenchConfig]) error {
	app.AddWorkload(NewSelfBench(app.Config))
	return nil
}
