package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/Shopify/mybench"
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat/distuv"
)

func newRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func main() {
	bins := []float64{
		0.0,
		1.5,
		3.0,
		4.5,
		6.0,
	}

	frequency := []float64{
		50.0,
		25.0,
		15.0,
		10.0,
	}

	// Need this for later statistical assertions, but we try the frequency above
	// to make sure the histogram is normalizing internally.
	normalizedFrequency := []float64{
		0.5,
		0.25,
		0.15,
		0.1,
	}

	dist := mybench.NewHistogramFloatDistribution(bins, frequency)

	const n = 100000
	nnn := 0

	for j := 0; j < 5000; j++ {
		r := newRand()

		binsCount := make(map[int]int64) // index of the bin -> count
		for i := 0; i < n; i++ {
			v := dist.NextValue(r)
			idx := sort.SearchFloat64s(bins, v) - 1 // need to minus one to matchup the idx with the frequency array.
			count, _ := binsCount[idx]
			binsCount[idx] = count + 1
		}

		minProb := 1.0

		for bin, observedCount := range binsCount {
			if bin < 0 {
				panic(fmt.Sprintf("bin is out of bound: %d", bin))
			}

			if bin >= len(frequency) {
				panic(fmt.Sprintf("bin is out of bound: %d > %d", bin, len(frequency)))
			}

			expectedCount := normalizedFrequency[bin] * n
			deviation := math.Round(math.Abs(float64(observedCount) - expectedCount))

			// Approximation
			p := normalizedFrequency[bin]
			q := 1 - p
			sigma := math.Sqrt(n * p * q)
			z := deviation / sigma

			binomialDistribution := distuv.Binomial{N: n, P: normalizedFrequency[bin]}

			lower := binomialDistribution.CDF(expectedCount - deviation - 1)
			upper := binomialDistribution.CDF(expectedCount + deviation)

			// Basically this solves:
			//
			// P(deviation >= observedDeviation) = P(count < expectedCount - deviation) + P(count < expectedCount + deviation)
			//
			// If the deviation is unexpected, this number should be very very low.
			deviationProbability := lower + (1 - upper)

			fmt.Printf("%d, %f, %d, %f, %f, %f\n", j, expectedCount, observedCount, deviation, deviationProbability, z)
			if deviationProbability < minProb {
				minProb = deviationProbability
			}
		}

		if minProb < 0.001 {
			nnn++
		}
	}

	logrus.Info(nnn)
}
