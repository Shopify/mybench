package main

import "github.com/Shopify/mybench"

type ChirpBench struct {
	*mybench.BenchmarkConfig

	InitialNumRows int64
}

func (b ChirpBench) Name() string {
	return "ChirpBench"
}

func (b ChirpBench) RunLoader() error {
	idGen := mybench.NewAutoIncrementGenerator(1, 0)
	table := NewTableChirps(idGen)
	table.ReloadData(
		b.BenchmarkConfig.DatabaseConfig,
		b.InitialNumRows,
		200,
		b.BenchmarkConfig.RateControlConfig.Concurrency,
	)
	return nil
}

func (b ChirpBench) Workloads() ([]mybench.AbstractWorkload, error) {
	idGen, err := mybench.NewAutoIncrementGeneratorFromDatabase(b.BenchmarkConfig.DatabaseConfig, "chirps", "id")
	if err != nil {
		return nil, err
	}

	table := NewTableChirps(idGen)

	return []mybench.AbstractWorkload{
		NewReadLatestChirps(table),
		NewReadSingleChirp(table),
		NewInsertChirp(table),
	}, nil
}
