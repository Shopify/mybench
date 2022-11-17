package main

import (
	"github.com/Shopify/mybench"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

func NewMicroBenchTable(idGen *mybench.AutoIncrementGenerator, indexCardinality int) mybench.Table {
	return mybench.InitializeTable(mybench.Table{
		Name: "microbench",
		Columns: []*mybench.Column{
			{
				Name:       "id",
				Definition: "BIGINT(20) NOT NULL AUTO_INCREMENT",
				Generator:  idGen,
			},
			{
				Name:       "idx1",
				Definition: "VARCHAR(255)",
				Generator:  mybench.NewBoundedCardinalityStringGenerator(indexCardinality, 30),
			},
			{
				Name:       "idx2",
				Definition: "BIGINT(20)",
				Generator:  mybench.NewUniformIntGenerator(1, int64(indexCardinality)),
			},
			{
				Name:       "data1",
				Definition: "VARCHAR(255)",
				Generator:  mybench.NewTotallyRandomStringGenerator(10, 200),
			},
			{
				Name:       "data2",
				Definition: "BIGINT(20)",
				Generator:  mybench.NewUniformIntGenerator(1, 20000),
			},
			{
				Name:       "b",
				Definition: "TINYINT(1)",
				Generator:  mybench.NewUniformIntGenerator(0, 1),
			},
		},
		PrimaryKey: []string{"id"},
		Indices: [][]string{
			{"idx1"},
			{"idx2"},
		},
	})
}

type MicroBenchConfig struct {
	*mybench.BenchmarkAppConfig
	InitialNumRows   int64 `long:"numrows" description:"number of rows to load into the database" default:"1000000"`
	IndexCardinality int   `long:"index-cardinality" description:"number of different values to generate for the indexed columns (this needs to be specified for both --bench and --load)" default:"100000"`

	BulkSelectIndexedRate                float64 `long:"bulk-select-indexed" default:"0.0"`
	BulkSelectIndexedOrderIndexedRate    float64 `long:"bulk-select-indexed-order-indexed" default:"0.0"`
	BulkSelectIndexedOrderNonIndexedRate float64 `long:"bulk-select-indexed-order-non-indexed" default:"0.0"`
	BulkSelectIndexedFilterRate          float64 `long:"bulk-select-indexed-filter" default:"0.0"`
	PointSelectRate                      float64 `long:"point-select" default:"0.0"`
	BatchPointSelectRate                 float64 `long:"batch-point-select" default:"0.0"`
}

type MicroBenchContextData struct {
	Statement *client.Stmt
}

func main() {
	config := MicroBenchConfig{
		BenchmarkAppConfig: mybench.NewBenchmarkAppConfig(),
	}
	_, err := flags.Parse(&config)
	if err != nil {
		panic(err)
	}

	app, err := mybench.NewBenchmarkApp("MicroBench", config, setupBenchmark, runLoader)
	if err != nil {
		panic(err)
	}

	err = app.Run()
	if err != nil {
		panic(err)
	}
}

func setupBenchmark(app *mybench.BenchmarkApp[MicroBenchConfig]) error {
	conn, err := app.Config.DatabaseConfig.Connection()
	if err != nil {
		logrus.WithError(err).Error("cannot connect to database")
		return err
	}
	defer conn.Close()

	idGen, err := mybench.NewAutoIncrementGeneratorFromDatabase(conn, app.Config.DatabaseConfig.Database, "microbench", "id")
	if err != nil {
		return err
	}

	// TODO: actually should get this value from the database, as opposed to getting it from commandline.
	table := NewMicroBenchTable(idGen, app.Config.IndexCardinality)

	if app.Config.BulkSelectIndexedRate > 0 {
		app.AddWorkload(NewBulkSelectIndexed(app.Config, &table, app.Config.BulkSelectIndexedRate))
	}

	if app.Config.BulkSelectIndexedOrderIndexedRate > 0 {
		app.AddWorkload(NewBulkSelectIndexedOrder(app.Config, &table, app.Config.BulkSelectIndexedOrderIndexedRate, "idx1"))
	}

	if app.Config.BulkSelectIndexedOrderNonIndexedRate > 0 {
		app.AddWorkload(NewBulkSelectIndexedOrder(app.Config, &table, app.Config.BulkSelectIndexedOrderNonIndexedRate, "data1"))
	}

	if app.Config.BulkSelectIndexedFilterRate > 0 {
		app.AddWorkload(NewBulkSelectIndexedFilter(app.Config, &table, app.Config.BulkSelectIndexedFilterRate, "b"))
	}

	if app.Config.PointSelectRate > 0 {
		app.AddWorkload(NewPointSelect(app.Config, &table, app.Config.PointSelectRate, 1))
	}

	if app.Config.BatchPointSelectRate > 0 {
		app.AddWorkload(NewPointSelect(app.Config, &table, app.Config.BatchPointSelectRate, 200))
	}

	return nil
}

func runLoader(app *mybench.BenchmarkApp[MicroBenchConfig]) error {
	NewMicroBenchTable(mybench.NewAutoIncrementGenerator(0, 0), app.Config.IndexCardinality).ReloadData(
		app.Config.DatabaseConfig,
		app.Config.InitialNumRows,
		500,
		app.Config.LoadConcurrency,
	)
	return nil
}
