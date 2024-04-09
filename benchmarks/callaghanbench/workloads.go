package main

import (
	"bytes"
	"fmt"

	"github.com/Shopify/mybench"
)

type Insert struct {
	mybench.WorkloadConfig
	mybench.NoContextData

	batchSize int
	table     mybench.Table
}

func (w *Insert) Event(ctx mybench.WorkerContext[mybench.NoContextData]) error {
	// https://github.com/mdcallag/mytools/blob/3c57ee97431112bdc167fc9c0ef032b24cb3c485/bench/ibench/iibench.py#L707-L731
	var valuesBuf bytes.Buffer
	for i := 0; i < w.batchSize; i++ {
		valuesBuf.WriteByte('(')

		dateandtime := w.table.Generate(ctx.Rand, "dateandtime").(string)
		valuesBuf.WriteString(fmt.Sprintf("'%s'", dateandtime))
		valuesBuf.WriteByte(',')

		cashregisterid := w.table.Generate(ctx.Rand, "cashregisterid").(int64)
		valuesBuf.WriteString(fmt.Sprintf("%d", cashregisterid))
		valuesBuf.WriteByte(',')

		customerid := w.table.Generate(ctx.Rand, "customerid").(int64)
		valuesBuf.WriteString(fmt.Sprintf("%d", customerid))
		valuesBuf.WriteByte(',')

		productid := w.table.Generate(ctx.Rand, "productid").(int64)
		valuesBuf.WriteString(fmt.Sprintf("%d", productid))
		valuesBuf.WriteByte(',')

		price := w.table.Generate(ctx.Rand, "price").(float64)
		valuesBuf.WriteString(fmt.Sprintf("%f", price))
		valuesBuf.WriteByte(',')

		data := w.table.Generate(ctx.Rand, "data").(string)
		valuesBuf.WriteString(fmt.Sprintf("'%s'", data))

		valuesBuf.WriteByte(')')
		if i != w.batchSize-1 {
			valuesBuf.WriteByte(',')
		}
	}

	query := fmt.Sprintf("insert into purchase_index (dateandtime, cashregisterid, customerid, productid, price, data) values %s", valuesBuf.String())
	_, err := ctx.Conn.Execute(query)
	return err
}

func NewInsert(table mybench.Table) mybench.AbstractWorkload {
	workloadInterface := &Insert{
		WorkloadConfig: mybench.WorkloadConfig{
			Name:          "Insert",
			WorkloadScale: 1.0,
		},
		batchSize: 1000,
		table:     table,
	}

	return mybench.NewWorkload[mybench.NoContextData](workloadInterface)
}
