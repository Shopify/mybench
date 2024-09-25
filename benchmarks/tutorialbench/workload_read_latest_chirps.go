package main

import (
	"fmt"

	"github.com/Shopify/mybench"
)

type ReadLatestChirps struct {
	mybench.WorkloadConfig
	mybench.NoContextData
	table mybench.Table
}

func (w *ReadLatestChirps) Event(ctx mybench.WorkerContext[mybench.NoContextData]) error {
	query := fmt.Sprintf("SELECT * FROM `%s` ORDER BY created_at DESC LIMIT 200", w.table.Name)
	_, err := ctx.Conn.Execute(query)
	return err
}

func NewReadLatestChirps(table mybench.Table) mybench.AbstractWorkload {
	workloadInterface := &ReadLatestChirps{
		WorkloadConfig: mybench.WorkloadConfig{
			Name:          "ReadLatestChirps",
			WorkloadScale: 0.75,
		},
		table: table,
	}

	return mybench.NewWorkload[mybench.NoContextData](workloadInterface)
}
