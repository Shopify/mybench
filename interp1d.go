package mybench

import (
	"errors"
	"sort"
)

func interpolate(x0, y0, x1, y1, x float64) float64 {
	return y0 + (x-x0)*(y1-y0)/(x1-x0)
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

	return interpolate(x0, y0, x1, y1, x), nil
}
