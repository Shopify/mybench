package main

import (
	"runtime/trace"

	"github.com/Shopify/mybench"
)

type InsertSimpleTable struct {
	mybench.WorkloadConfig
	mybench.NoContextData
	table mybench.Table
}

func NewInsertSimpleTable(table mybench.Table) mybench.AbstractWorkload {
	workloadInterface := &InsertSimpleTable{
		WorkloadConfig: mybench.WorkloadConfig{
			Name:          "InsertSimpleTable",
			WorkloadScale: 0.5,
		},

		table: table,
	}

	return mybench.NewWorkload[mybench.NoContextData](workloadInterface)
}

func (r *InsertSimpleTable) Event(ctx mybench.WorkerContext[mybench.NoContextData]) error {
	query := "INSERT INTO example_table (id, data) VALUES (?, ?)"
	args := make([]interface{}, 2)
	args[0] = r.table.Generate(ctx.Rand, "id")
	args[1] = r.table.Generate(ctx.Rand, "data")

	_, err := ctx.Conn.Execute(query, args...)
	return err
}

type UpdateSimpleTable struct {
	mybench.WorkloadConfig
	mybench.NoContextData
	table mybench.Table
}

func NewUpdateSimpleTable(table mybench.Table) mybench.AbstractWorkload {
	workloadInterface := &UpdateSimpleTable{
		WorkloadConfig: mybench.WorkloadConfig{
			Name:          "UpdateSimpleTable",
			WorkloadScale: 0.5,
		},

		table: table,
	}

	return mybench.NewWorkload[mybench.NoContextData](workloadInterface)
}

func (r *UpdateSimpleTable) Event(ctx mybench.WorkerContext[mybench.NoContextData]) error {
	query := "UPDATE example_table SET data = ? WHERE id = ?"
	args := make([]interface{}, 2)
	args[0] = r.table.Generate(ctx.Rand, "data")
	args[1] = r.table.SampleFromExisting(ctx.Rand, "id")

	defer trace.StartRegion(ctx.TraceCtx, "UpdateSimpleTableQuery").End()
	_, err := ctx.Conn.Execute(query, args...)
	return err
}
