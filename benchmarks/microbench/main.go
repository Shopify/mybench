package main

import (
	"flag"

	"github.com/Shopify/mybench"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/sirupsen/logrus"
)

type MicroBenchContextData struct {
	Statement *client.Stmt
}

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

func main() {
	benchmarkInterface := MicroBench{
		BenchmarkConfig: mybench.NewBenchmarkConfig(),
	}

	flag.Int64Var(&benchmarkInterface.InitialNumRows, "numrows", 1000000, "the number of rows to load into the database")
	flag.IntVar(&benchmarkInterface.IndexCardinality, "index-cardinality", 100000, "the number of different values to generate for the indexed columns (needed to be the same for both load and bench)")

	flag.Float64Var(&benchmarkInterface.BulkSelectIndexedRate, "bulk-select-indexed", 0.0, "the event rate for bulk insert indexed workload")
	flag.Float64Var(&benchmarkInterface.BulkSelectIndexedOrderIndexedRate, "bulk-select-indexed-order-indexed", 0.0, "the event rate for bulk insert indexed workload with an order by another indexed column")
	flag.Float64Var(&benchmarkInterface.BulkSelectIndexedOrderNonIndexedRate, "bulk-select-indexed-order-non-indexed", 0.0, "the event rate for bulk insert indexed workload with an order by another non-indexed column")
	flag.Float64Var(&benchmarkInterface.BulkSelectIndexedFilterRate, "bulk-select-indexed-filter", 0.0, "the event rate for bulk insert indexed workload but also filter the data after by a non-indexed column")
	flag.Float64Var(&benchmarkInterface.PointSelectRate, "point-select", 0.0, "the event rate for the point select workload")
	flag.Float64Var(&benchmarkInterface.BatchPointSelectRate, "batch-point-select", 0.0, "the event rate for the batch point select workload")

	flag.Parse()

	err := mybench.Run(benchmarkInterface)
	if err != nil {
		panic(err)
	}
}

type MicroBench struct {
	*mybench.BenchmarkConfig

	InitialNumRows   int64
	IndexCardinality int

	BulkSelectIndexedRate                float64
	BulkSelectIndexedOrderIndexedRate    float64
	BulkSelectIndexedOrderNonIndexedRate float64
	BulkSelectIndexedFilterRate          float64
	PointSelectRate                      float64
	BatchPointSelectRate                 float64
}

func (b MicroBench) Name() string {
	return "MicroBench"
}

func (b MicroBench) Workloads() ([]mybench.AbstractWorkload, error) {
	conn, err := b.BenchmarkConfig.DatabaseConfig.Connection()
	if err != nil {
		logrus.WithError(err).Error("cannot connect to database")
		return nil, err
	}
	defer conn.Close()

	idGen, err := mybench.NewAutoIncrementGeneratorFromDatabase(conn, b.BenchmarkConfig.DatabaseConfig.Database, "microbench", "id")
	if err != nil {
		return nil, err
	}

	// TODO: actually should get this value from the database, as opposed to getting it from commandline.
	table := NewMicroBenchTable(idGen, b.IndexCardinality)

	workloads := []mybench.AbstractWorkload{}
	if b.BulkSelectIndexedRate > 0 {
		workloads = append(workloads, NewBulkSelectIndexed(b.BenchmarkConfig, &table, b.BulkSelectIndexedRate))
	}

	if b.BulkSelectIndexedOrderIndexedRate > 0 {
		workloads = append(workloads, NewBulkSelectIndexedOrder(b.BenchmarkConfig, &table, b.BulkSelectIndexedOrderIndexedRate, "idx1"))
	}

	if b.BulkSelectIndexedOrderNonIndexedRate > 0 {
		workloads = append(workloads, NewBulkSelectIndexedOrder(b.BenchmarkConfig, &table, b.BulkSelectIndexedOrderNonIndexedRate, "data1"))
	}

	if b.BulkSelectIndexedFilterRate > 0 {
		workloads = append(workloads, NewBulkSelectIndexedFilter(b.BenchmarkConfig, &table, b.BulkSelectIndexedFilterRate, "b"))
	}

	if b.PointSelectRate > 0 {
		workloads = append(workloads, NewPointSelect(b.BenchmarkConfig, &table, b.PointSelectRate, 1))
	}

	if b.BatchPointSelectRate > 0 {
		workloads = append(workloads, NewPointSelect(b.BenchmarkConfig, &table, b.BatchPointSelectRate, 200))
	}

	return workloads, nil
}

func (b MicroBench) RunLoader() error {
	NewMicroBenchTable(mybench.NewAutoIncrementGenerator(0, 0), b.IndexCardinality).ReloadData(
		b.DatabaseConfig,
		b.InitialNumRows,
		500,
		b.RateControlConfig.Concurrency,
	)
	return nil
}
