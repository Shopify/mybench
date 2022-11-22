package main

import (
	"fmt"

	"github.com/Shopify/mybench"
)

type BulkSelectIndexedFilter struct {
	mybench.WorkloadConfig
	table       *mybench.Table
	filterField string
}

func NewBulkSelectIndexedFilter(config *mybench.BenchmarkConfig, table *mybench.Table, eventRate float64, filterField string) mybench.AbstractWorkload {
	eventRate = eventRate * config.Multiplier
	workloadInterface := &BulkSelectIndexedFilter{
		WorkloadConfig: mybench.NewWorkloadConfigWithDefaults(mybench.WorkloadConfig{
			Name:           "BulkSelectIndexedFilter_" + filterField,
			DatabaseConfig: config.DatabaseConfig,
			RateControl: mybench.RateControlConfig{
				EventRate: eventRate,
			},
		}),
		table:       table,
		filterField: filterField,
	}

	workload, err := mybench.NewWorkload[MicroBenchContextData](workloadInterface)
	if err != nil {
		panic(err)
	}

	return workload
}

func (c *BulkSelectIndexedFilter) Event(ctx mybench.WorkerContext[MicroBenchContextData]) error {
	args := []interface{}{
		c.table.Generate(ctx.Rand, "idx2"),
		c.table.Generate(ctx.Rand, c.filterField), // Should actually be SampleFrom
	}

	_, err := ctx.Data.Statement.Execute(args...)
	return err
}

func (c *BulkSelectIndexedFilter) NewContextData(conn *mybench.Connection) (MicroBenchContextData, error) {
	var err error
	contextData := MicroBenchContextData{}

	query := fmt.Sprintf("SELECT * FROM `%s` WHERE idx2 = ? AND `%s` = ?", c.table.Name, c.filterField)
	contextData.Statement, err = conn.Prepare(query)
	return contextData, err
}
