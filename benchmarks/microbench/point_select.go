package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Shopify/mybench"
)

type PointSelect struct {
	mybench.WorkloadConfig
	table *mybench.Table

	batchSize int
}

func NewPointSelect(config *mybench.BenchmarkConfig, table *mybench.Table, eventRate float64, batchSize int) mybench.AbstractWorkload {
	eventRate = eventRate * config.Multiplier
	var workloadInterface mybench.WorkloadInterface[MicroBenchContextData] = &PointSelect{
		WorkloadConfig: mybench.NewWorkloadConfigWithDefaults(mybench.WorkloadConfig{
			Name:           "PointSelect_" + strconv.Itoa(batchSize),
			DatabaseConfig: config.DatabaseConfig,
			RateControl: mybench.RateControlConfig{
				EventRate: eventRate,
			},
		}),
		table:     table,
		batchSize: batchSize,
	}

	workload, err := mybench.NewWorkload(workloadInterface)
	if err != nil {
		panic(err)
	}

	return workload
}

func (c *PointSelect) Event(ctx mybench.WorkerContext[MicroBenchContextData]) error {
	args := make([]interface{}, c.batchSize)
	for i := 0; i < c.batchSize; i++ {
		args[i] = c.table.SampleFromExisting(ctx.Rand, "id")
	}

	_, err := ctx.Data.Statement.Execute(args...)
	return err
}

func (c *PointSelect) NewContextData(conn *mybench.Connection) (MicroBenchContextData, error) {
	var clause string
	if c.batchSize == 1 {
		clause = "id = ?"
	} else {
		questionMarks := strings.Repeat("?,", c.batchSize)
		questionMarks = questionMarks[:len(questionMarks)-1]
		clause = "id IN (" + questionMarks + ")"
	}

	var err error
	contextData := MicroBenchContextData{}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", c.table.Name, clause)
	contextData.Statement, err = conn.Prepare(query)
	return contextData, err
}
