package mybench

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var seed int64

func init() {
	seed = time.Now().UnixMicro()
	logrus.WithField("seed", seed).Info("seeding random number generator for test")
}

func newRandForTest() *Rand {
	return &Rand{
		Rand: rand.New(rand.NewSource(seed)),
	}
}

func TestNullGenerator(t *testing.T) {
	gen := NewNullGenerator()
	r := NewRand()
	t.Run("Generate", func(t *testing.T) {
		require.Equal(t, nil, gen.Generate(r))
	})

	t.Run("SampleFromExisting", func(t *testing.T) {
		require.Equal(t, nil, gen.Generate(r))
	})
}

func TestUniformIntGenerator(t *testing.T) {
	const min, max int64 = 30, 40
	const numValues = max - min

	gen := NewUniformIntGenerator(min, max)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000

		valuesCount := make(map[int64]int)

		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(int64)
			require.True(t, ok)

			count := valuesCount[v]
			valuesCount[v] = count + 1
		}

		require.Equal(t, int(numValues), len(valuesCount))

		for i := int64(min); i < max; i++ {
			_, found := valuesCount[i]
			require.True(t, found, "%d is not generated even though it should have been", i)
		}
	})
}

func TestUniformFloatGenerator(t *testing.T) {
	const min, max float64 = 30.0, 40.0

	gen := NewUniformFloatGenerator(min, max)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000
		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(float64)
			require.True(t, ok, "generated value %v is not float64")
			require.True(t, v >= min, "generated value %v is out of desired range", v)
			require.True(t, v < max, "generated value %v is out of desired range", v)
		}
	})
}

func TestNormalIntGenerator(t *testing.T) {
	const mean, stddev = int64(4), int64(25)
	const allowedDeviation = stddev * 5

	gen := NewNormalIntGenerator(mean, stddev)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 10

		// Probably not the best test, but good enough for now
		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(int64)
			require.True(t, ok, "generated value %v is not int64")
			require.True(t, v <= mean+allowedDeviation, "generated value %v is out of desired range (could happen occasionally due to probabilistic generation)", v)
			require.True(t, v >= mean-allowedDeviation, "generated value %v is out of desired range (could happen occasionally due to probabilistic generation)", v)
		}
	})
}

func TestNormalFloatGenerator(t *testing.T) {
	const mean, stddev = 4.0, 25.0
	const allowedDeviation = stddev * 5.0

	gen := NewNormalFloatGenerator(mean, stddev)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 10

		// Probably not the best test, but good enough for now
		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(float64)
			require.True(t, ok, "generated value %v is not float64")
			require.True(t, v <= mean+allowedDeviation, "generated value %v is out of desired range (could happen occasionally due to probabilistic generation)", v)
			require.True(t, v >= mean-allowedDeviation, "generated value %v is out of desired range (could happen occasionally due to probabilistic generation)", v)
		}
	})
}

func TestHistogramIntGenerator(t *testing.T) {
	binsEndPoints := []float64{0.5, 1.5, 2.5}
	frequency := []float64{0.7, 0.3}

	gen := NewHistogramIntGenerator(binsEndPoints, frequency)
	r := NewRand()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000

		valuesCount := map[int64]int{}

		for i := 0; i < n; i++ {
			v, _ := gen.Generate(r).(int64)
			count := valuesCount[v]
			valuesCount[v] = count + 1
		}

		require.Equal(t, 2, len(valuesCount))

		count, found := valuesCount[1]
		require.True(t, found)

		// Since we only generate two values, the number of values generated for
		// each follows the binomial distribution, which when a large number of
		// values are generated can be approximated with the normal distribution.
		//
		// Note: cannot change this test to generate more than 2 values as then the
		// counts for each value would follow the multinomial distribution, which
		// can only be approximated with a multivariate normal distribution. This
		// would significantly complicate the math here, as you would need to find
		// a region of acceptable values in high-dimensional space.
		p := frequency[0]
		expectedCount := n * p
		standardDeviation := math.Sqrt(n * p * (1 - p))
		deviation := math.Abs(float64(count) - expectedCount)
		require.True(
			t,
			deviation < 4*standardDeviation,
			"generated %d 1's, which is outside the 4 sigma expectations (%d +/- %.2f) based on the bionomial distribution. This test may occasionally (1/15787) fail due to probability.",
			count,
			expectedCount,
			standardDeviation,
		)

		count, found = valuesCount[2]
		require.True(t, found)
		p = frequency[1]
		expectedCount = n * p
		standardDeviation = math.Sqrt(n * p * (1 - p))
		deviation = math.Abs(float64(count) - expectedCount)
		require.True(
			t,
			deviation < 4*standardDeviation,
			"generated %d 2's, which is outside the 4 sigma expectations (%d +/- %.2f) based on the bionomial distribution. This test may occasionally (1/15787) fail due to probability.",
			count,
			expectedCount,
			standardDeviation,
		)
	})
}

func TestHistogramFloatGenerator(t *testing.T) {
	binsEndPoints := []float64{0.0, 1.0, 2.0}
	frequency := []float64{0.7, 0.3}

	gen := NewHistogramFloatGenerator(binsEndPoints, frequency)
	r := NewRand()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000

		valuesCount := map[int]int{}

		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(float64)
			require.True(t, ok, "should be generating float64 but didn't")
			var bin int
			if v >= 0 && v < 1 {
				bin = 0
			} else if v >= 1 && v < 2 {
				bin = 1
			} else {
				require.FailNow(t, "value %v generated is outside the acceptable ranges", v)
			}

			count := valuesCount[bin]
			valuesCount[bin] = count + 1
		}

		require.Equal(t, 2, len(valuesCount), "should have generated two bins of values but found %v", valuesCount)

		count, found := valuesCount[0]
		require.True(t, found)

		// Since we only generate two values, the number of values generated for
		// each follows the binomial distribution, which when a large number of
		// values are generated can be approximated with the normal distribution.
		//
		// Note: cannot change this test to generate more than 2 values as then the
		// counts for each value would follow the multinomial distribution, which
		// can only be approximated with a multivariate normal distribution. This
		// would significantly complicate the math here, as you would need to find
		// a region of acceptable values in high-dimensional space.
		p := frequency[0]
		expectedCount := n * p
		standardDeviation := math.Sqrt(n * p * (1 - p))
		deviation := math.Abs(float64(count) - expectedCount)
		require.True(
			t,
			deviation < 4*standardDeviation,
			"generated %d values in [0, 1), which is outside the 4 sigma expectations (%f +/- %.2f) based on the bionomial distribution. This test may occasionally (1/15787) fail due to probability.",
			count,
			expectedCount,
			standardDeviation,
		)

		count, found = valuesCount[1]
		require.True(t, found)
		p = frequency[1]
		expectedCount = n * p
		standardDeviation = math.Sqrt(n * p * (1 - p))
		deviation = math.Abs(float64(count) - expectedCount)
		require.True(
			t,
			deviation < 4*standardDeviation,
			"generated %d values [1, 2), which is outside the 4 sigma expectations (%f +/- %.2f) based on the bionomial distribution. This test may occasionally (1/15787) fail due to probability.",
			count,
			expectedCount,
			standardDeviation,
		)
	})
}

func TestUniformCardinalityStringGenerator(t *testing.T) {
	const cardinality, length = 5, 10
	gen := NewUniformCardinalityStringGenerator(cardinality, length)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000
		valuesCount := map[string]int{}
		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(string)
			require.True(t, ok, "should be generating string but is not")
			require.Equal(t, length, len(v))
			count := valuesCount[v]
			valuesCount[v] = count + 1
		}

		require.Equal(t, cardinality, len(valuesCount))
	})
}

func TestHistogramCardinalityStringGenerator(t *testing.T) {
	binsEndPoints := []float64{
		0.5,
		1.5,
		2.5,
	}
	frequency := []float64{
		0.8,
		0.2,
	}
	const length = 15

	gen := NewHistogramCardinalityStringGenerator(binsEndPoints, frequency, length)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000
		valuesCount := map[string]int{}
		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(string)
			require.True(t, ok, "should be generating string but is not")
			require.Equal(t, length, len(v))
			count := valuesCount[v]
			valuesCount[v] = count + 1
		}

		require.Equal(t, 2, len(valuesCount))
	})
}

func TestUniformLengthStringGenerator(t *testing.T) {
	const minLength, maxLength = 3, 5
	gen := NewUniformLengthStringGenerator(minLength, maxLength)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000
		lengthCount := map[int]int{}

		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(string)
			require.True(t, ok, "should be generating string but is not")
			count := lengthCount[len(v)]
			lengthCount[len(v)] = count + 1
		}

		require.Equal(t, 2, len(lengthCount), "should have generated strings with two length, but got %v", lengthCount)

		count, found := lengthCount[3]
		require.True(t, found, "didn't generate strings of length 3")
		p := 0.5
		expectedCount := n * p
		standardDeviation := math.Sqrt(n * p * (1 - p))
		deviation := math.Abs(float64(count) - expectedCount)
		require.True(
			t,
			deviation < 4*standardDeviation,
			"generated %d 3 length strings, which is outside the 4 sigma expectations (%f +/- %.2f) based on the bionomial distribution. This test may occasionally (1/15787) fail due to probability.",
		)

		count, found = lengthCount[4]
		require.True(t, found, "didn't generate strings of length 4")
		p = 0.5
		expectedCount = n * p
		standardDeviation = math.Sqrt(n * p * (1 - p))
		deviation = math.Abs(float64(count) - expectedCount)
		require.True(
			t,
			deviation < 4*standardDeviation,
			"generated %d 4 length strings, which is outside the 4 sigma expectations (%f +/- %.2f) based on the bionomial distribution. This test may occasionally (1/15787) fail due to probability.",
		)
	})
}

func TestHistogramLengthStringGenerator(t *testing.T) {
	binsEndPoints := []float64{
		9.5,
		10.5,
		11.5,
	}
	frequency := []float64{
		0.8,
		0.2,
	}

	gen := NewHistogramLengthStringGenerator(binsEndPoints, frequency)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000
		lengthCount := map[int]int{}

		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(string)
			require.True(t, ok, "should be generating string but is not")
			count := lengthCount[len(v)]
			lengthCount[len(v)] = count + 1
		}

		require.Equal(t, len(frequency), len(lengthCount), "should have 2 entries but: %v", lengthCount)
		_, found := lengthCount[10]
		require.True(t, found, "didn't find length of 10, but should have: %v", lengthCount)

		_, found = lengthCount[11]
		require.True(t, found, "didn't find length of 11, but should have: %v", lengthCount)
	})
}

func TestUniqueStringGenerator(t *testing.T) {
	const length = 15
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		gen := NewUniqueStringGenerator(length, 0, 0)
		const n = 100_000 // Generate a large number of unique values
		values := map[string]struct{}{}
		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(string)
			require.True(t, ok, "should be generating string but is not")
			_, found := values[v]
			require.False(t, found, "value %s has already been generated but should not have", v)
			values[v] = struct{}{}
		}

		require.Equal(t, n, len(values), "should have generate %d values but did not: %v", n, values)
	})

	t.Run("SampleFromExisting", func(t *testing.T) {
		gen := NewUniqueStringGenerator(length, 0, 0)
		// Generate a small amount of unique values so when running
		// SampleFromExisting, every value should be generated.
		const numGenerate = 10
		const numSample = 100_000

		values := map[string]struct{}{}
		for i := 0; i < numGenerate; i++ {
			v, ok := gen.Generate(r).(string)
			require.True(t, ok, "should be generating string but is not")
			_, found := values[v]
			require.False(t, found, "value %s has already been generated but should not have", v)
			values[v] = struct{}{}
		}

		samples := map[string]int{}

		for i := 0; i < numSample; i++ {
			v, ok := gen.SampleFromExisting(r).(string)
			require.True(t, ok, "should be generating string but is not")
			_, found := values[v]
			require.True(t, found, "value %s was sampled but wasn't generated: %v", v, values)
			count := samples[v]
			samples[v] = count + 1
		}

		require.Equal(t, numGenerate, len(samples))

		// Check that every value has been generated
		for value := range samples {
			delete(values, value)
		}
		require.Equal(t, 0, len(values))
	})
}

func TestNowGenerator(t *testing.T) {
	gen := NewNowGenerator()
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		now := time.Now()
		v, ok := gen.Generate(r).(string)
		require.True(t, ok, "should be generating string but is not")
		now2, err := time.Parse("2006-01-02 15:04:05", v)
		require.Nil(t, err)

		diff := math.Abs(now2.Sub(now).Seconds())
		require.True(t, diff <= 2, "should generate now but didn't: %v (diff: %v)", v, diff)
	})

	t.Run("SampleFromExisting", func(t *testing.T) {
		start := time.Now().Truncate(time.Second)
		gen.Generate(r)

		time.Sleep(3 * time.Second)
		const n = 10
		values := make([]time.Time, n)
		for i := 0; i < n; i++ {
			v, ok := gen.SampleFromExisting(r).(string)
			require.True(t, ok, "should be generating string but is not")
			sampledTime, err := time.Parse("2006-01-02 15:04:05", v)
			require.Nil(t, err)
			values[i] = sampledTime
		}

		end := time.Now().Round(time.Second)

		for _, value := range values {
			require.True(t, value.Sub(start) >= 0, "should be after start but is not: %v (start = %v)", value, start)
			require.True(t, value.Sub(end) <= 0, "should be before end but is not: %v (end = %v)", value, start)
		}
	})
}

func TestUniformDatetimeGeneratorWithoutGenerateNow(t *testing.T) {
	r := newRandForTest()
	intervals := []DatetimeInterval{
		{
			Start: time.Date(2006, time.November, 10, 11, 45, 0, 0, time.UTC),
			End:   time.Date(2006, time.November, 10, 12, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2007, time.November, 10, 11, 45, 0, 0, time.UTC),
			End:   time.Date(2007, time.November, 10, 12, 0, 0, 0, time.UTC),
		},
	}
	t.Run("Generate", func(t *testing.T) {
		gen := NewUniformDatetimeGenerator(intervals, false)
		const n = 100_000

		buckets := map[int]int{}

		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(string)
			require.True(t, ok, "should be generating string but is not")
			sampledTime, err := time.Parse("2006-01-02 15:04:05", v)
			require.Nil(t, err)

			foundBucket := false
			for j, interval := range intervals {
				if sampledTime.Sub(interval.Start) >= 0 && sampledTime.Sub(interval.End) < 0 {
					foundBucket = true
					count := buckets[j]
					buckets[j] = count + 1
				}
			}

			if !foundBucket {
				require.FailNow(t, "should be generated between intervals but is not", "sampled = %v, intervals: %v", sampledTime, intervals)
			}
		}

		require.Equal(t, len(intervals), len(buckets))
	})
}

func TestGenerateUniqueStringFromInteger(t *testing.T) {
	output := make(map[string]struct{})
	for i := int64(0); i < 1000000; i++ {
		value := generateUniqueStringFromInt(i, 20)
		_, found := output[value]
		require.Equal(t, found, false, fmt.Sprintf("found duplicate values for integer %d with value %s", i, value))
		output[value] = struct{}{}
	}
}
