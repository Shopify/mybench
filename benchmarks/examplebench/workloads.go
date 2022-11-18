package main

import (
	"github.com/Shopify/mybench"
)

type InsertSimpleTable struct {
	mybench.WorkloadConfig
	table mybench.Table
}

func NewInsertSimpleTable(exampleBench ExampleBench, table mybench.Table) mybench.AbstractWorkload {
	var workloadInterface mybench.WorkloadInterface[mybench.NoContextData] = &InsertSimpleTable{
		WorkloadConfig: mybench.NewWorkloadConfigWithDefaults(mybench.WorkloadConfig{
			Name:           "InsertSimpleTable",
			DatabaseConfig: exampleBench.BenchmarkConfig.DatabaseConfig,
			RateControl: mybench.RateControlConfig{
				EventRate: 50 * exampleBench.BenchmarkConfig.Multiplier,
			},
		}),

		table: table,
	}

	workload, err := mybench.NewWorkload(workloadInterface)
	if err != nil {
		panic(err)
	}

	return workload
}

func (r *InsertSimpleTable) Config() mybench.WorkloadConfig {
	return r.WorkloadConfig
}

func (r *InsertSimpleTable) Event(ctx mybench.WorkerContext[mybench.NoContextData]) error {
	query := "INSERT INTO example_table (id, data) VALUES (?, ?)"
	args := make([]interface{}, 2)
	args[0] = r.table.Generate(ctx.Rand, "id")
	args[1] = r.table.Generate(ctx.Rand, "data")

	_, err := ctx.Conn.Execute(query, args...)
	return err
}

func (r *InsertSimpleTable) NewContextData(conn *mybench.Connection) (mybench.NoContextData, error) {
	return mybench.NewNoContextData()
}

type UpdateSimpleTable struct {
	mybench.WorkloadConfig
	table mybench.Table
}

func NewUpdateSimpleTable(exampleBench ExampleBench, table mybench.Table) mybench.AbstractWorkload {
	var workloadInterface mybench.WorkloadInterface[mybench.NoContextData] = &UpdateSimpleTable{
		WorkloadConfig: mybench.NewWorkloadConfigWithDefaults(mybench.WorkloadConfig{
			Name:           "UpdateSimpleTable",
			DatabaseConfig: exampleBench.BenchmarkConfig.DatabaseConfig,
			RateControl: mybench.RateControlConfig{
				EventRate: 50 * exampleBench.BenchmarkConfig.Multiplier,
			},
		}),

		table: table,
	}

	workload, err := mybench.NewWorkload(workloadInterface)
	if err != nil {
		panic(err)
	}

	return workload
}

func (r *UpdateSimpleTable) Config() mybench.WorkloadConfig {
	return r.WorkloadConfig
}

func (r *UpdateSimpleTable) Event(ctx mybench.WorkerContext[mybench.NoContextData]) error {
	query := "UPDATE example_table SET data = ? WHERE id = ?"
	args := make([]interface{}, 2)
	args[0] = r.table.Generate(ctx.Rand, "data")
	args[1] = r.table.SampleFromExisting(ctx.Rand, "id")

	_, err := ctx.Conn.Execute(query, args...)
	return err
}

func (r *UpdateSimpleTable) NewContextData(conn *mybench.Connection) (mybench.NoContextData, error) {
	return mybench.NewNoContextData()
}
