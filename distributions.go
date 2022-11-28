package mybench

import (
	"math"
	"math/rand"
	"sort"

	"go.uber.org/atomic"
	"golang.org/x/exp/constraints"
)

type number interface {
	constraints.Integer | constraints.Float
}

// An interface for a random number distribution.
//
// TODO: having a float as a return is not ideal for high int64s as we may lose
// accuracy, which could be problem for things like auto increment ids
type RandomNumberDistribution[T number] interface {
	// This should generate the next value in the random number distribution
	NextValue(*rand.Rand) T

	// This should sample from an existing value that *probably* has already been
	// generated. Note, this can be a value that has never been generated before.
	//
	// In other words, value returned from this function is generated with a best
	// effort approach and most of the time can be the same as the NextValue
	// function above. It's only in special cases, like the
	// AutoIncrementDistribution, does this need to be different.
	ExistingValue(r *rand.Rand) T
}

// This generates float64 values based on a histogram distribution.
//
// How this generator work is relatively simple:
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
type HistogramFloatDistribution struct {
	bins                   []float64
	cumulativeDistribution []float64
}

// Creates a histogram generator which generates values based on a histogram
// distribution.
//
// The value at frequency[i] correspond to the bin starting at [bins[i],
// bins[i+1]). Thus, len(bins) == len(frequency) + 1.
//
// Also, the value in bins must be sorted.
func NewHistogramFloatDistribution(bins, frequency []float64) *HistogramFloatDistribution {
	if len(bins) != len(frequency)+1 {
		panic("len(bins) != len(frequency) + 1")
	}

	if !sort.Float64sAreSorted(bins) {
		panic("bins is not properly sorted")
	}

	total := 0.0
	for _, v := range frequency {
		total += v
	}

	normalizedFrequency := make([]float64, len(frequency))
	cumulativeDistribution := make([]float64, len(frequency)+1)

	cumulativeDistribution[0] = 0.0
	currentCumulativeDistributionValue := 0.0
	for i, v := range frequency {
		normalizedFrequency[i] = v / total

		cumulativeDistribution[i+1] = currentCumulativeDistributionValue + normalizedFrequency[i]
		currentCumulativeDistributionValue = cumulativeDistribution[i+1]
	}

	return &HistogramFloatDistribution{
		bins:                   bins,
		cumulativeDistribution: cumulativeDistribution,
	}
}

func (d *HistogramFloatDistribution) NextValue(r *rand.Rand) float64 {
	x := r.Float64()

	v, err := interp1d(d.cumulativeDistribution, d.bins, x)
	if err != nil {
		panic(err)
	}

	return v
}

func (d *HistogramFloatDistribution) ExistingValue(r *rand.Rand) float64 {
	return d.NextValue(r)
}

// An integer version of HistogramFloatDistribution.
//
// For the time being, this is implemented by reusing the histogram float
// distribution and thus can suffer from loss of precision.
//
// TODO: fix this at some point
type HistogramIntDistribution struct {
	*HistogramFloatDistribution
}

func NewHistogramIntDistribution(bins []int64, frequency []float64) *HistogramIntDistribution {
	binsFloat := make([]float64, len(bins))
	for i, v := range bins {
		binsFloat[i] = float64(v)
	}
	return &HistogramIntDistribution{
		HistogramFloatDistribution: NewHistogramFloatDistribution(binsFloat, frequency),
	}
}

func (d *HistogramIntDistribution) NextValue(r *rand.Rand) int64 {
	return int64(d.HistogramFloatDistribution.NextValue(r))
}

func (d *HistogramIntDistribution) ExistingValue(r *rand.Rand) int64 {
	return int64(d.HistogramFloatDistribution.ExistingValue(r))
}

// Samples float64 values based on a Gaussian distribution
//
// Note that the ExistingValue for this distribution is the same as NextValue
// and thus has no memory of past generated values.
type GaussianFloatDistribution struct {
	mean   float64
	stddev float64
}

// Creates a new Gaussian distribution with mean and stddev (standard
// deviation) specified.
func NewGaussianFloatDistribution(mean, stddev float64) *GaussianFloatDistribution {
	return &GaussianFloatDistribution{
		mean:   mean,
		stddev: stddev,
	}
}

func (d *GaussianFloatDistribution) NextValue(r *rand.Rand) float64 {
	return r.NormFloat64()*d.stddev + d.mean
}

func (d *GaussianFloatDistribution) ExistingValue(r *rand.Rand) float64 {
	return d.NextValue(r)
}

// The integer version of the GaussianFloatDistribution. Internally, this
// embeds a GaussianFloatDistribution and converts the results from that to an
// int64.
type GaussianIntDistribution struct {
	*GaussianFloatDistribution
}

func NewGaussianIntDistribution(mean, stddev float64) *GaussianIntDistribution {
	return &GaussianIntDistribution{
		GaussianFloatDistribution: NewGaussianFloatDistribution(mean, stddev),
	}
}

func (d *GaussianIntDistribution) NextValue(r *rand.Rand) int64 {
	return int64(math.Round(d.GaussianFloatDistribution.NextValue(r)))
}

func (d *GaussianIntDistribution) ExistingValue(r *rand.Rand) int64 {
	return int64(math.Round(d.GaussianFloatDistribution.ExistingValue(r)))
}

// Samples float64 values between min and max uniformly. Technically, the
// sampling exclude max, so the interval being sampled is [min, max).
//
// Note that the ExistingValue for this distribution is the same as NextValue
// and thus has no memory of past generated values.
type UniformFloatDistribution struct {
	min float64
	max float64
}

func NewUniformFloatDistribution(min, max float64) *UniformFloatDistribution {
	return &UniformFloatDistribution{
		min: min,
		max: max,
	}
}

func (d *UniformFloatDistribution) NextValue(r *rand.Rand) float64 {
	return r.Float64()*(d.max-d.min) + d.min
}

// Samples int64 values between min and max uniformly. Technically, the
// sampling exclude max, so the interval being sampled is [min, max).
//
// Note that the ExistingValue for this distribution is the same as NextValue
// and thus has no memory of past generated values.
//
// Note: this is NOT using UniformFloatDistribution internally and does NOT
// suffer from loss of precision issues.
type UniformIntDistribution struct {
	min int64
	max int64
}

func NewUniformIntDistribution(min, max int64) *UniformIntDistribution {
	return &UniformIntDistribution{
		min: min,
		max: max,
	}
}

func (d *UniformIntDistribution) NextValue(r *rand.Rand) int64 {
	return r.Int63n(d.max-d.min) + d.min
}

func (d *UniformIntDistribution) ExistingValue(r *rand.Rand) int64 {
	return d.NextValue(r)
}

// Generates integer values that are atomically incremented.
//
// ExistingValue is an integer generated by sampling uniformly between the min
// and the current values.
type AutoIncrementDistribution struct {
	min     int64
	current *atomic.Int64
}

func NewAutoIncrementDistribution(min, current int64) *AutoIncrementDistribution {
	return &AutoIncrementDistribution{
		min:     min,
		current: atomic.NewInt64(current),
	}
}

func (d *AutoIncrementDistribution) NextValue(r *rand.Rand) int64 {
	return d.current.Add(1)
}

func (d *AutoIncrementDistribution) ExistingValue(r *rand.Rand) int64 {
	return r.Int63n(d.current.Load()-d.min+1) + d.min
}
