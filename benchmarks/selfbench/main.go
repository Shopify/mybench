package main

import (
	"github.com/Shopify/mybench"
	"github.com/jessevdk/go-flags"
)

func main() {
	config := SelfBenchConfig{
		BenchmarkAppConfig: mybench.NewBenchmarkAppConfig(),
	}
	// This benchmark does not need a connection
	// Kind of a hack...
	config.BenchmarkAppConfig.DatabaseConfig.NoConnection = true
	config.BenchmarkAppConfig.DatabaseConfig.Host = "no-host"

	_, err := flags.Parse(&config)
	if err != nil {
		panic(err)
	}

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
