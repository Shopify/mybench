package main

import (
	"flag"

	"github.com/Shopify/mybench"
)

func NewSimpleTable(idGen *mybench.AutoIncrementGenerator) mybench.Table {
	return mybench.InitializeTable(mybench.Table{
		Name: "example_table",
		Columns: []*mybench.Column{
			{
				Name:       "id",
				Definition: "BIGINT(20) NOT NULL AUTO_INCREMENT",
				Generator:  idGen,
			},
			{
				Name:       "data",
				Definition: "VARCHAR(255)",
				Generator:  mybench.NewUniformLengthStringGenerator(10, 200),
			},
		},
		PrimaryKey: []string{"id"},
	})
}

func main() {
	benchmarkInterface := ExampleBench{
		BenchmarkConfig: mybench.NewBenchmarkConfig(),
	}
	flag.Int64Var(&benchmarkInterface.InitialNumRows, "numrows", 1000000, "the number of rows to load into the database")
	flag.Parse()

	err := mybench.Run(benchmarkInterface)
	if err != nil {
		panic(err)
	}
}

type ExampleBench struct {
	*mybench.BenchmarkConfig
	InitialNumRows int64
}

func (b ExampleBench) Name() string {
	return "ExampleBench"
}

func (b ExampleBench) Workloads() ([]mybench.AbstractWorkload, error) {
	idGen, err := mybench.NewAutoIncrementGeneratorFromDatabase(b.BenchmarkConfig.DatabaseConfig, "example_table", "id")
	if err != nil {
		return nil, err
	}

	table := NewSimpleTable(idGen)

	return []mybench.AbstractWorkload{
		NewInsertSimpleTable(table),
		NewUpdateSimpleTable(table),
	}, nil
}

func (b ExampleBench) RunLoader() error {
	NewSimpleTable(mybench.NewAutoIncrementGenerator(0, 0)).ReloadData(b.BenchmarkConfig.DatabaseConfig, b.InitialNumRows, 200, 8)
	return nil
}
