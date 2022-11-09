package main

import (
	"flag"

	"github.com/Shopify/mybench"
	"github.com/go-mysql-org/go-mysql/client"
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
	InitialNumRows   int64
	IndexCardinality int

	BulkSelectIndexedRate                float64
	BulkSelectIndexedOrderIndexedRate    float64
	BulkSelectIndexedOrderNonIndexedRate float64
	BulkSelectIndexedFilterRate          float64
	PointSelectRate                      float64
	BatchPointSelectRate                 float64
}

type MicroBenchContextData struct {
	Statement *client.Stmt
}

func main() {
	config := MicroBenchConfig{
		BenchmarkAppConfig: mybench.NewBenchmarkAppConfig(),
	}

	flag.Int64Var(&config.InitialNumRows, "numrows", 1000000, "the number of rows to load into the database")
	flag.IntVar(&config.IndexCardinality, "index-cardinality", 100000, "the number of different values to generate for the indexed columns (needed to be the same for both load and bench)")

	flag.Float64Var(&config.BulkSelectIndexedRate, "bulk-select-indexed", 0.0, "the event rate for bulk insert indexed workload")
	flag.Float64Var(&config.BulkSelectIndexedOrderIndexedRate, "bulk-select-indexed-order-indexed", 0.0, "the event rate for bulk insert indexed workload with an order by another indexed column")
	flag.Float64Var(&config.BulkSelectIndexedOrderNonIndexedRate, "bulk-select-indexed-order-non-indexed", 0.0, "the event rate for bulk insert indexed workload with an order by another non-indexed column")
	flag.Float64Var(&config.BulkSelectIndexedFilterRate, "bulk-select-indexed-filter", 0.0, "the event rate for bulk insert indexed workload but also filter the data after by a non-indexed column")
	flag.Float64Var(&config.PointSelectRate, "point-select", 0.0, "the event rate for the point select workload")
	flag.Float64Var(&config.BatchPointSelectRate, "batch-point-select", 0.0, "the event rate for the batch point select workload")

	flag.Parse()

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
