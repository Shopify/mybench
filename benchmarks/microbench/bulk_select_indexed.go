package main

import (
	"fmt"

	"github.com/Shopify/mybench"
)

type BulkSelectIndexed struct {
	mybench.WorkloadConfig
	table *mybench.Table
}

func NewBulkSelectIndexed(table *mybench.Table, rateScale float64) mybench.AbstractWorkload {
	workloadInterface := &BulkSelectIndexed{
		WorkloadConfig: mybench.WorkloadConfig{
			Name:          "BulkSelectIndexed",
			WorkloadScale: rateScale,
		},
		table: table,
	}

	return mybench.NewWorkload[MicroBenchContextData](workloadInterface)
}

func (c *BulkSelectIndexed) Event(ctx mybench.WorkerContext[MicroBenchContextData]) error {
	args := []interface{}{
		c.table.Generate(ctx.Rand, "idx2"),
	}

	_, err := ctx.Data.Statement.Execute(args...)
	return err
}

func (c *BulkSelectIndexed) NewContextData(conn *mybench.Connection) (MicroBenchContextData, error) {
	var err error
	contextData := MicroBenchContextData{}

	query := fmt.Sprintf("SELECT * FROM `%s` WHERE idx2 = ?", c.table.Name)
	contextData.Statement, err = conn.Prepare(query)
	return contextData, err
}
