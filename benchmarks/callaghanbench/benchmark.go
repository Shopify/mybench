package main

import "github.com/Shopify/mybench"

// https://github.com/mdcallag/mytools/blob/3c57ee97431112bdc167fc9c0ef032b24cb3c485/bench/ibench/iibench.py#L357-L390
// https://github.com/mdcallag/mytools/blob/3c57ee97431112bdc167fc9c0ef032b24cb3c485/bench/ibench/iibench.py#L517
func NewTablePurchaseIndex(idGen mybench.DataGenerator) mybench.Table {

	return mybench.InitializeTable(mybench.Table{
		Name: "purchase_index",
		Columns: []*mybench.Column{
			{
				Name:       "transactionid",
				Definition: "bigint not null auto_increment",
				Generator:  idGen,
			},
			{
				Name:       "dateandtime",
				Definition: "datetime",
				Generator:  mybench.NewNowGenerator(),
			},
			{
				Name:       "cashregisterid",
				Definition: "int not null",
				// https://github.com/mdcallag/mytools/blob/3c57ee97431112bdc167fc9c0ef032b24cb3c485/bench/ibench/iibench.py#L496
				Generator: mybench.NewUniformIntGenerator(0, 1000),
			},
			{
				Name:       "customerid",
				Definition: "int not null",
				// https://github.com/mdcallag/mytools/blob/3c57ee97431112bdc167fc9c0ef032b24cb3c485/bench/ibench/iibench.py#L498
				Generator: mybench.NewUniformIntGenerator(0, 100000),
			},
			{
				Name:       "productid",
				Definition: "int not null",
				// https://github.com/mdcallag/mytools/blob/3c57ee97431112bdc167fc9c0ef032b24cb3c485/bench/ibench/iibench.py#L497
				Generator: mybench.NewUniformIntGenerator(0, 10000),
			},
			{
				Name:       "price",
				Definition: "float not null",
				// https://github.com/mdcallag/mytools/blob/3c57ee97431112bdc167fc9c0ef032b24cb3c485/bench/ibench/iibench.py#L499
				// This is not implemented to be accurate. The generator implemented by
				// Mark Callaghan is dependent on customer id. Specifically it's:
				//
				// (UniformFloat(0, 500) + customerid) / 100.0
				//
				// TODO: mybench should have a way to generate values dependent on
				// another value, but then the generators may have to be specified in a
				// DAG, which can be a pain.
				Generator: mybench.NewUniformFloatGenerator(0, 5),
			},
			{
				Name:       "data",
				Definition: "varchar(4000)",
				// https://github.com/mdcallag/mytools/blob/3c57ee97431112bdc167fc9c0ef032b24cb3c485/bench/ibench/iibench.py#L500
				// TODO: this is not accurately translated.
				Generator: mybench.NewUniformLengthStringGenerator(10, 11),
			},
		},
		PrimaryKey: []string{"transactionid"},
	})
}

type CallaghanBench struct {
	*mybench.BenchmarkConfig

	InitialNumRows int64
}

func (b CallaghanBench) Name() string {
	return "CallaghanBench"
}

func (b CallaghanBench) RunLoader() error {
	idGen := mybench.NewNullGenerator()
	table := NewTablePurchaseIndex(idGen)
	table.ReloadData(
		b.BenchmarkConfig.DatabaseConfig,
		b.InitialNumRows,
		200,
		b.BenchmarkConfig.RateControlConfig.Concurrency,
	)
	return nil
}

func (b CallaghanBench) Workloads() ([]mybench.AbstractWorkload, error) {
	idGen := mybench.NewNullGenerator()
	table := NewTablePurchaseIndex(idGen)

	return []mybench.AbstractWorkload{
		NewInsert(table),
	}, nil
}
