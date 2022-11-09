package main

import (
	"fmt"
	"math"
	"time"

	"github.com/sirupsen/logrus"
)

// Hopefully prevents the compiler optimizer from optimizing out the sumUntil method?
// https://github.com/golang/go/issues/27400
var sum = int64(0)
var twLogger = logrus.WithField("tag", "time_waster")
var timeWaster TimeWaster

func init() {
	var err error
	timeWaster, err = calibrateTiming(0)
	if err != nil {
		panic(err)
	}
	twLogger.Infof("Calibrated time waster: %s", timeWaster.String())

	twLogger.Infof("Testing time waster...")
	for i := 0; i < 30; i++ {
		start := time.Now()
		timeWaster.Spin(time.Duration(i) * time.Millisecond)
		dt := time.Since(start)
		twLogger.Infof("Expected: %dms, Actual: %v, Diff: %v", i, dt, time.Duration(i)*time.Millisecond-dt)
	}
}

func sumUntil(v int64) int64 {
	for i := int64(1); i < v; i++ {
		sum += i
	}

	return sum
}

type TimeWaster struct {
	Slope            float64
	Intercept        float64
	R2               float64
	InterpolationMax int64
}

func (w TimeWaster) Spin(duration time.Duration) {
	dt := float64(duration.Nanoseconds())
	sumUntil(int64(w.Slope*dt + w.Intercept))
}

func (w TimeWaster) String() string {
	return fmt.Sprintf("TimeWaster(v = %.3f * t + %.2f, r^2 = %.3f, interp_t_until = %v)", w.Slope, w.Intercept, w.R2, time.Duration(w.InterpolationMax))
}

func linearRegression(values [][2]int64) (float64, float64, float64) {
	n := float64(len(values))
	xsum := 0.0
	ysum := 0.0
	xxsum := 0.0
	yysum := 0.0
	xysum := 0.0

	for _, value := range values {
		xsum += float64(value[0])
		ysum += float64(value[1])
		xxsum += float64(value[0] * value[0])
		yysum += float64(value[1] * value[1])
		xysum += float64(value[0] * value[1])
	}

	denom := (n*xxsum - xsum*xsum)
	if denom == 0.0 {
		return 0.0, 0.0, 0.0
	}

	slope := (n*xysum - xsum*ysum) / denom
	intercept := ysum/n - slope*xsum/n
	r := (n*xysum - xsum*ysum) / math.Sqrt((n*xxsum-xsum*xsum)*(n*yysum-ysum*ysum))
	return slope, intercept, r * r
}

func calibrateTiming(suggestedMaxValue int64) (TimeWaster, error) {
	twLogger.Infof("Calibrating timing...")

	if suggestedMaxValue == 0 {
		suggestedMaxValue = 10000000
	}

	interval := suggestedMaxValue / 100

	measuredTiming := [][2]int64{}
	maxDt := int64(0)

	// Need to first burn
	benchmark := func(record bool) {
		for v := interval; v < suggestedMaxValue; v += interval {
			start := time.Now()
			sumUntil(v)
			dt := time.Since(start).Nanoseconds()
			if dt > maxDt {
				maxDt = dt
			}
			if record {
				measuredTiming = append(measuredTiming, [2]int64{dt, v})
			}
		}
	}

	// Ensure turbo boost has been engaged
	start := time.Now()
	for time.Since(start).Milliseconds() < 200 {
		benchmark(false)
	}

	for i := 0; i < 3; i++ {
		benchmark(true)
	}

	// Using the sum value prevents compiler optimization of the sumUntil function? See comment on sum global variable
	twLogger.Infof("Timing calibrated! Ignore this value: %d", sum)
	sum = 0

	slope, intercept, r := linearRegression(measuredTiming)
	tw := TimeWaster{
		Slope:            slope,
		Intercept:        intercept,
		R2:               r * r,
		InterpolationMax: maxDt,
	}

	var err error = nil
	if tw.R2 <= 0.99 {
		twLogger.Error("failed to calibrate timing. this suggest inconsistent performance from this machine or perhaps the code is compiled and optimized incorrectly")
		err = fmt.Errorf("calibration failed, r^2 value is less than 0.99, which is not satisfactory: %s", tw.String())
	}

	return tw, err
}
