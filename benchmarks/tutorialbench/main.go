package main

import (
	"flag"

	"github.com/Shopify/mybench"
)

func main() {
	benchmarkInterface := ChirpBench{
		BenchmarkConfig: mybench.NewBenchmarkConfig(),
	}
	flag.Int64Var(&benchmarkInterface.InitialNumRows, "numrows", 10_000_000, "the number of rows to load into the database")
	flag.Parse()

	err := mybench.Run(benchmarkInterface)
	if err != nil {
		panic(err)
	}
}
