package main

import (
	"fmt"

	"github.com/Shopify/mybench"
)

type InsertChirp struct {
	mybench.WorkloadConfig
	mybench.NoContextData
	table mybench.Table
}

func (w *InsertChirp) Event(ctx mybench.WorkerContext[mybench.NoContextData]) error {
	query := fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?)", w.table.Name)
	id := w.table.Generate(ctx.Rand, "id")
	content := w.table.Generate(ctx.Rand, "content")
	createdAt := w.table.Generate(ctx.Rand, "created_at")

	_, err := ctx.Conn.Execute(query, id, content, createdAt)
	return err
}

func NewInsertChirp(table mybench.Table) mybench.AbstractWorkload {
	workloadInterface := &ReadLatestChirps{
		WorkloadConfig: mybench.WorkloadConfig{
			Name:          "InsertChirp",
			WorkloadScale: 0.05,
		},
		table: table,
	}

	return mybench.NewWorkload[mybench.NoContextData](workloadInterface)
}
