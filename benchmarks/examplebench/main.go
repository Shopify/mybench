package main

import (
	"github.com/Shopify/mybench"
	"github.com/jessevdk/go-flags"
)

type ExampleBenchmarkConfig struct {
	*mybench.BenchmarkAppConfig
	InitialNumRows int64 `long:"numrows" description:"the number of rows to load into the database" default:"1000000"`
}

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
				Generator:  mybench.NewTotallyRandomStringGenerator(10, 200),
			},
		},
		PrimaryKey: []string{"id"},
	})
}

func main() {
	config := ExampleBenchmarkConfig{
		BenchmarkAppConfig: mybench.NewBenchmarkAppConfig(),
	}
	_, err := flags.Parse(&config)
	if err != nil {
		panic(err)
	}

	app, err := mybench.NewBenchmarkApp("ExampleBench", config, setupBenchmark, runLoader)
	if err != nil {
		panic(err)
	}

	err = app.Run()
	if err != nil {
		panic(err)
	}
}

func setupBenchmark(app *mybench.BenchmarkApp[ExampleBenchmarkConfig]) error {
	// TODO: once there is a loader, we need to fetch the maximum id before the benchmark here to seed the auto increment generator.
	conn, err := app.Config.DatabaseConfig.Connection()
	if err != nil {
		return err
	}
	defer conn.Close()

	idGen, err := mybench.NewAutoIncrementGeneratorFromDatabase(conn, app.Config.DatabaseConfig.Database, "example_table", "id")
	if err != nil {
		return err
	}

	gen := NewSimpleTable(idGen)

	app.AddWorkload(NewInsertSimpleTable(app.Config, gen))
	app.AddWorkload(NewUpdateSimpleTable(app.Config, gen))
	return nil
}

func runLoader(app *mybench.BenchmarkApp[ExampleBenchmarkConfig]) error {
	NewSimpleTable(mybench.NewAutoIncrementGenerator(0, 0)).ReloadData(app.Config.DatabaseConfig, 1000000, 200, 8)
	return nil
}
