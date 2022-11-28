package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/Shopify/mybench"
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

	for j := 0; j < 5000; j++ {
		r := newRand()

		binsCount := make(map[int]int64) // index of the bin -> count
		for i := 0; i < n; i++ {
			v := dist.NextValue(r)
			idx := sort.SearchFloat64s(bins, v) - 1 // need to minus one to matchup the idx with the frequency array.
			count, _ := binsCount[idx]
			binsCount[idx] = count + 1
		}

		for bin, observedCount := range binsCount {
			if bin < 0 {
				panic(fmt.Sprintf("bin is out of bound: %d", bin))
			}

			if bin >= len(frequency) {
				panic(fmt.Sprintf("bin is out of bound: %d > %d", bin, len(frequency)))
			}

			expectedCount := normalizedFrequency[bin] * n
			deviation := math.Round(math.Abs(float64(observedCount) - expectedCount))

			binomialDistribution := distuv.Binomial{N: n, P: normalizedFrequency[bin]}

			lower := binomialDistribution.CDF(expectedCount - deviation)
			upper := binomialDistribution.CDF(expectedCount + deviation)

			// Basically this solves:
			//
			// P(deviation >= observedDeviation) = P(count < expectedCount - deviation) + P(count < expectedCount + deviation)
			//
			// If the deviation is unexpected, this number should be very very low.
			deviationProbability := lower + (1 - upper)

			deviationProbability2 := binomialDistribution.Prob(float64(observedCount))

			fmt.Printf("%d, %f, %d, %f, %f, %f\n", j, expectedCount, observedCount, deviation, deviationProbability, deviationProbability2)
		}
	}
}

// func newRandForTest() *rand.Rand {
// 	return rand.New(rand.NewSource(randomSeed))
// }
//
// }
//
//
//
// 	t.Run("NextValue follows histogram distribution", func(t *testing.T) {
// 		const n = 100000
//
// 		r := newRandForTest()
// 		binsCount := make(map[int]int64) // index of the bin -> count
//
// 		for i := 0; i < n; i++ {
// 			v := dist.NextValue(r)
// 			idx := sort.SearchFloat64s(bins, v) - 1 // need to minus one to matchup the idx with the frequency array.
// 			count, _ := binsCount[idx]
// 			binsCount[idx] = count + 1
// 		}
//
// 		for bin, observedCount := range binsCount {
// 			require.True(t, bin >= 0, "bin is out of bound: %d", bin)
// 			require.True(t, bin < len(frequency), "bin is out of bound: %d > %d", bin, len(frequency))
//
// 			expectedCount := normalizedFrequency[bin] * n
// 			deviation := math.Round(math.Abs(float64(observedCount) - expectedCount))
//
// 			binomialDistribution := distuv.Binomial{N: n, P: normalizedFrequency[bin]}
//
// 			lower := binomialDistribution.CDF(expectedCount - deviation)
// 			upper := binomialDistribution.CDF(expectedCount + deviation)
//
// 			// Basically this solves:
// 			//
// 			// P(deviation >= observedDeviation) = P(count < expectedCount - deviation) + P(count < expectedCount + deviation)
// 			//
// 			// If the deviation is unexpected, this number should be very very low.
// 			deviationProbability := lower + (1 - upper)
//
// 			require.True(
// 				t,
// 				deviationProbability > (1.0/1000.0),
// 				"deviation detected for histogram generator. this should happen 1/1000 tests. expected = %v; observed = %v; probability %v",
// 				expectedCount,
// 				observedCount,
// 				deviationProbability,
// 			)
// 		}
// 	})
