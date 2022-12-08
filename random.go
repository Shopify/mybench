package mybench

import (
	"errors"
	"math"
	"math/rand"
	"sort"
	"time"
)

type Rand struct {
	*rand.Rand
}

// Creates a new Rand object
func NewRand() *Rand {
	return &Rand{
		Rand: rand.New(rand.NewSource(time.Now().UnixMicro())),
	}
}

func (r *Rand) UniformFloat(min, max float64) float64 {
	return r.Rand.Float64()*(max-min) + min
}

func (r *Rand) UniformInt(min, max int64) int64 {
	return r.Rand.Int63n(max-min) + min
}

func (r *Rand) NormalFloat(mean, stddev float64) float64 {
	return r.Rand.NormFloat64()*stddev + mean
}

func (r *Rand) NormalInt(mean, stddev int64) int64 {
	return int64(r.NormalFloat(float64(mean), float64(stddev)))
}

func (r *Rand) HistFloat(hist HistogramDistribution) float64 {
	x := r.Rand.Float64()

	v, err := interp1d(hist.cumulativeDistribution, hist.binsEndPoints, x)
	if err != nil {
		panic(err)
	}

	return v
}

func (r *Rand) HistInt(hist HistogramDistribution) int64 {
	return int64(math.Round(r.HistFloat(hist)))
}

// xs and ys must be sorted and must have the same size
func interp1d(xs, ys []float64, x float64) (float64, error) {
	// This is the index to insert x into xs, which means it will be >= the index
	// of the value if the value exists.
	//
	// This means the interpolation range is between i - 1 and i
	i := sort.SearchFloat64s(xs, x)

	if i <= 0 {
		// In this case the search finds we would have to insert x at the beginning
		// of the array, which means the value is out of bounds at the beginning,
		// which means we would need extrapolation.
		return 0.0, errors.New("out of range")
	}

	if i >= len(xs) {
		// In this case the search finds we would have to insert x at the end of
		// the array, which means the value is out of bound at the end, which means
		// we would need extrapolation.
		return 0.0, errors.New("out of range")
	}

	x0, x1 := xs[i-1], xs[i]
	y0, y1 := ys[i-1], ys[i]

	return y0 + (x-x0)*(y1-y0)/(x1-x0), nil
}

// This generates float64 values based on a discrete probability distribution
// (represented via a histogram) via the inverse transform sampling algorithm
// (https://en.wikipedia.org/wiki/Inverse_transform_sampling). Specifically,
// the steps followed are:
//
//  1. Normalize the frequency values of the histogram to values of between 0
//     and 1.
//  2. Compute the cumulative distribution for the normalized histogram such
//     that its output value is also between 0 and 1. So we have the function
//     cdf(bin_value) -> [0, 1].
//  3. Generate a random value, x, between 0 and 1. This value represents a
//     sampled number from the cdf function output. If we compute the inverse
//     function cdf^-1(x) -> bin_value, we will obtain a randomly sampled bin
//     value that will be randomly sampled according to the frequency specified
//     in the histogram.
//  4. The inverse function cdf^-1(x) is calculated via linear interpolation.
//
// Note that the ExistingValue for this distribution is the same as NextValue
// and thus has no memory of past generated values.
type HistogramDistribution struct {
	binsEndPoints          []float64
	cumulativeDistribution []float64
}

// Creates a histogram distribution which is used by Rand to generate random
// numbers.
//
// The value at frequency[i] correspond to the bin starting at [bins[i],
// bins[i+1]). Thus, len(bins) == len(frequency) + 1.
//
// Also, the value in bins must be sorted.
func NewHistogramDistribution(binsEndPoints, frequency []float64) HistogramDistribution {
	if len(binsEndPoints) != len(frequency)+1 {
		panic("len(binsEndPoints) != len(frequency)")
	}

	if !sort.Float64sAreSorted(binsEndPoints) {
		panic("binsEndPoints is not sorted")
	}

	total := 0.0
	for _, v := range frequency {
		total += v
	}

	cumulativeDistribution := make([]float64, len(binsEndPoints))

	cumulativeDistribution[0] = 0.0

	for i, v := range frequency {
		normalizedFrequency := v / total
		cumulativeDistribution[i+1] = cumulativeDistribution[i] + normalizedFrequency
	}

	return HistogramDistribution{
		binsEndPoints:          binsEndPoints,
		cumulativeDistribution: cumulativeDistribution,
	}
}
