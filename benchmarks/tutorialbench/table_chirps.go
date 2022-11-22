package main

import (
	"github.com/Shopify/mybench"
)

func NewTableChirps(idGen mybench.DataGenerator) mybench.Table {
	return mybench.InitializeTable(mybench.Table{
		Name: "chirps",
		Columns: []*mybench.Column{
			{
				Name:       "id",
				Definition: "BIGINT(20) NOT NULL AUTO_INCREMENT",
				Generator:  idGen,
			},
			{
				Name:       "content",
				Definition: "VARCHAR(140)",
				Generator: mybench.NewHistogramLengthStringGenerator(
					[]float64{-0.5, 20.5, 40.5, 60.5, 80.5, 100.5, 120.5, 140.5},
					[]float64{
						10, // [0, 20)
						10, // [20, 40)
						10, // [40, 60)
						15, // [60, 80)
						15, // [80, 100)
						25, // [100, 120)
						15, // [120, 140)
					},
				),
			},
			{
				Name:       "created_at",
				Definition: "DATETIME",
				Generator:  mybench.NewNowGenerator(),
			},
		},
		Indices: [][]string{
			{"created_at"},
		},
		PrimaryKey: []string{"id"},
	})
}
