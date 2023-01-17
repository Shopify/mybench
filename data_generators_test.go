package mybench

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newRandForTest() *Rand {
	// Fixed seed so the output generated is always the same
	return &Rand{
		Rand: rand.New(rand.NewSource(42)),
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
	const min, max int64 = 30, 35

	gen := NewUniformIntGenerator(min, max)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 500_000

		valuesCount := make(map[int64]int)

		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(int64)
			require.True(t, ok)

			count := valuesCount[v]
			valuesCount[v] = count + 1
		}

		expectedValuesCount := map[int64]int{
			30: 100342,
			31: 99819,
			32: 100050,
			33: 100025,
			34: 99764,
		}

		require.Equal(t, expectedValuesCount, valuesCount)
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
	const allowedDeviation = stddev * 6

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
	const allowedDeviation = stddev * 6.0

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
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000

		valuesCount := map[int64]int{}

		for i := 0; i < n; i++ {
			v, _ := gen.Generate(r).(int64)
			count := valuesCount[v]
			valuesCount[v] = count + 1
		}

		expectedValuesCount := map[int64]int{
			1: 69841,
			2: 30159,
		}

		require.Equal(t, expectedValuesCount, valuesCount)
	})
}

func TestHistogramFloatGenerator(t *testing.T) {
	binsEndPoints := []float64{0.0, 1.0, 2.0, 3.0}
	frequency := []float64{0.7, 0.2, 0.1}

	gen := NewHistogramFloatGenerator(binsEndPoints, frequency)
	r := newRandForTest()

	t.Run("Generate", func(t *testing.T) {
		const n = 100_000

		valuesCount := [3]int{0, 0, 0}

		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(float64)
			require.True(t, ok, "should be generating float64 but didn't")
			var bin int
			if v >= 0 && v < 1 {
				bin = 0
			} else if v >= 1 && v < 2 {
				bin = 1
			} else if v >= 2 && v < 3 {
				bin = 2
			} else {
				require.FailNow(t, "value %v generated is outside the acceptable ranges", v)
			}

			valuesCount[bin]++
		}

		expectedValuesCount := [3]int{69841, 20136, 10023}
		require.Equal(t, expectedValuesCount, valuesCount)
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

		expectedValuesCount := map[string]int{
			"0!cfcd2084": 20052,
			"1!c4ca4238": 19893,
			"2!c81e728d": 20160,
			"3!eccbc87e": 20012,
			"4!a87ff679": 19883,
		}

		require.Equal(t, expectedValuesCount, valuesCount)
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

		expectedValuesCount := map[string]int{
			"1!c4ca4238a0b92": 79785,
			"2!c81e728d9d4c2": 20215,
		}

		require.Equal(t, expectedValuesCount, valuesCount)
	})
}

func TestVariableLengthUniformStringsGenerator(t *testing.T) {
	const minLength, maxLength = 3, 6
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

		expectedLengthCount := map[int]int{
			3: 33020,
			4: 33586,
			5: 33394,
		}

		require.Equal(t, expectedLengthCount, lengthCount)
	})
}

func TestUniformLengthStringsGenerator(t *testing.T) {
	const minLength, maxLength = 3, 3
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

		expectedLengthCount := map[int]int{
			3: 100_000,
		}

		require.Equal(t, expectedLengthCount, lengthCount)
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

		// Amazingly, the seed we chose result in a perfect ratio
		expectedLengthCount := map[int]int{
			10: 80000,
			11: 20000,
		}

		require.Equal(t, expectedLengthCount, lengthCount)
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
		const numGenerate = 5
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
		t.Skip() // not the best test for now so skipped
		now := time.Now().UTC()
		v, ok := gen.Generate(r).(string)
		require.True(t, ok, "should be generating string but is not")
		now2, err := time.Parse("2006-01-02 15:04:05", v)
		require.Nil(t, err)

		diff := math.Abs(now2.Sub(now).Seconds())
		require.True(t, diff <= 2, "should generate now but didn't: expected (%v) actual (%v)", now, now2)
	})

	t.Run("SampleFromExisting", func(t *testing.T) {
		start := time.Now().UTC().Truncate(time.Second)
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

		end := time.Now().UTC().Round(time.Second)

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

		buckets := [2]int{0, 0}

		for i := 0; i < n; i++ {
			v, ok := gen.Generate(r).(string)
			require.True(t, ok, "should be generating string but is not")
			sampledTime, err := time.Parse("2006-01-02 15:04:05", v)
			require.Nil(t, err)

			foundBucket := false
			for j, interval := range intervals {
				if sampledTime.Sub(interval.Start) >= 0 && sampledTime.Sub(interval.End) < 0 {
					foundBucket = true
					buckets[j]++
				}
			}

			if !foundBucket {
				require.FailNow(t, "should be generated between intervals but is not", "sampled = %v, intervals: %v", sampledTime, intervals)
			}
		}

		expectedBuckets := [2]int{49989, 50011}
		require.Equal(t, expectedBuckets, buckets)
	})
}
