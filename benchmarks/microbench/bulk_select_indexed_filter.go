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

func NewBulkSelectIndexedFilter(table *mybench.Table, rateScale float64, filterField string) mybench.AbstractWorkload {
	workloadInterface := &BulkSelectIndexedFilter{
		WorkloadConfig: mybench.WorkloadConfig{
			Name:          "BulkSelectIndexedFilter_" + filterField,
			WorkloadScale: rateScale,
		},
		table:       table,
		filterField: filterField,
	}

	return mybench.NewWorkload[MicroBenchContextData](workloadInterface)
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
