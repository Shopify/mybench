package main

import (
	"fmt"

	"github.com/Shopify/mybench"
	"github.com/go-mysql-org/go-mysql/client"
)

type ReadSingleChirpContext struct {
	stmt *client.Stmt
}

type ReadSingleChirp struct {
	mybench.WorkloadConfig
	table mybench.Table
}

func (w *ReadSingleChirp) Event(ctx mybench.WorkerContext[ReadSingleChirpContext]) error {
	id := w.table.SampleFromExisting(ctx.Rand, "id")
	_, err := ctx.Data.stmt.Execute(id)
	return err
}

func (w *ReadSingleChirp) NewContextData(conn *mybench.Connection) (ReadSingleChirpContext, error) {
	var err error
	ctx := ReadSingleChirpContext{}
	ctx.stmt, err = conn.Prepare(fmt.Sprintf("SELECT * FROM `%s` WHERE id = ?", w.table.Name))
	return ctx, err
}

func NewReadSingleChirp(table mybench.Table) mybench.AbstractWorkload {
	workloadInterface := &ReadSingleChirp{
		WorkloadConfig: mybench.WorkloadConfig{
			Name:          "ReadSingleChirp",
			WorkloadScale: 0.2,
		},
		table: table,
	}

	return mybench.NewWorkload[ReadSingleChirpContext](workloadInterface)
}
