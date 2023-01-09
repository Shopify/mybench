package main

import (
	"flag"

	"github.com/Shopify/mybench"
)

func main() {
	benchmarkInterface := CallaghanBench{
		BenchmarkConfig: mybench.NewBenchmarkConfig(),
	}

	flag.Parse()
	err := mybench.Run(benchmarkInterface)
	if err != nil {
		panic(err)
	}
}
