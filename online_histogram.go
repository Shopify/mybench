package mybench

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

type IntervalData struct {
	StartTime time.Time
	EndTime   time.Time
	Count     int64
	Delta     float64
	Rate      float64

	// All data in microseconds
	Min          int64
	Mean         float64
	Max          int64
	Percentile25 int64
	Percentile50 int64
	Percentile75 int64
	Percentile90 int64
	Percentile99 int64

	UnderflowCount int64
	OverflowCount  int64

	UniformHist *UniformHistogram
}

// This is a double buffer implemented using a lock. The target usage is as
// follows:
//
// 1. Single consumer single producer.
// 2. The producer goroutine writes data frequently.
// 3. The consumer goroutine reads data infrequently.
// 4. The consumer goroutine will first swap the buffer. It gets the non-active
//    data during the swap. After the swap, it will read the data and then reset
//    the data to 0, so it can be swapped again.
// 5. We can never swap before the non-active data is reset.
// 6. While the producer goroutine is writing to the data, the swap is not
//    allowed to occur.
type LockedDoubleBuffer[T any] struct {
	buf [2]T
	idx int
	mut *sync.Mutex
}

func NewLockedDoubleBuffer[T any](newT func() T) *LockedDoubleBuffer[T] {
	return &LockedDoubleBuffer[T]{
		buf: [2]T{newT(), newT()},
		idx: 0,
		mut: &sync.Mutex{},
	}
}

// Swap the active and non-active data with a lock.
// Returns the non-active data.
func (b *LockedDoubleBuffer[T]) Swap(preSwapCallback func(nonActiveData T)) T {
	preSwapCallback(b.buf[(b.idx+1)%2])
	oldIdx := b.idx

	b.mut.Lock()
	b.idx = (b.idx + 1) % 2
	b.mut.Unlock()

	return b.buf[oldIdx]
}

// Since we need to prevent swapping from happening while writing, we are using
// a lock. Lock-less could work as well but will be more complex and may not be
// necessary.
func (b *LockedDoubleBuffer[T]) SafeActiveWrite(f func(T)) {
	b.mut.Lock()
	defer b.mut.Unlock()

	f(b.buf[b.idx])
}

// This extends the HDR histogram so it can track:
// - Start time
// - Under and overflow counts
type ExtendedHdrHistogram struct {
	hist *hdrhistogram.Histogram

	startTime      time.Time
	overflowCount  int64
	underflowCount int64
}

func NewExtendedHdrHistogram(startTime time.Time) *ExtendedHdrHistogram {
	hist := &ExtendedHdrHistogram{
		hist:           hdrhistogram.New(1, 10000000, 4), // 1us - 10s by default. TODO: make configurable?
		startTime:      startTime,
		overflowCount:  0,
		underflowCount: 0,
	}

	return hist
}

func (h *ExtendedHdrHistogram) RecordValue(v int64) {
	if v > h.hist.HighestTrackableValue() {
		h.overflowCount++
		return
	}

	if v < h.hist.LowestTrackableValue() {
		h.underflowCount++
		return
	}

	h.hist.RecordValue(v)
}

func (h *ExtendedHdrHistogram) ResetDataOnly() {
	h.hist.Reset()
	h.underflowCount = 0
	h.overflowCount = 0
}

func (h *ExtendedHdrHistogram) ResetStartTime(startTime time.Time) {
	h.startTime = startTime
}

// Should only be called from the data logger, after it is copied away from the double buffer.
// Maybe some of these "read" methods should be defined on a different type, so
// it can never be called on an object that could be in the double buffer.
func (h *ExtendedHdrHistogram) Merge(other *ExtendedHdrHistogram) {
	if !h.startTime.Equal(other.startTime) {
		panic(fmt.Sprintf("failed to merge histograms with different start time: %v %v", h.startTime, other.startTime))
	}

	h.underflowCount += other.underflowCount
	h.overflowCount += other.overflowCount

	h.hist.Merge(other.hist)
}

func (h *ExtendedHdrHistogram) IntervalData(endTime time.Time, histMin, histMax, histSize int64) IntervalData {
	data := IntervalData{
		StartTime:      h.startTime,
		EndTime:        endTime,
		Count:          h.hist.TotalCount() + h.underflowCount + h.overflowCount,
		Min:            h.hist.Min(),
		Mean:           h.hist.Mean(),
		Max:            h.hist.Max(),
		UnderflowCount: h.underflowCount,
		OverflowCount:  h.overflowCount,
	}

	percentiles := h.hist.ValueAtPercentiles([]float64{25.0, 50.0, 75.0, 90.0, 99.0})
	data.Delta = data.EndTime.Sub(data.StartTime).Seconds()
	data.Rate = float64(data.Count) / data.Delta
	data.Percentile25 = percentiles[25.0]
	data.Percentile50 = percentiles[50.0]
	data.Percentile75 = percentiles[75.0]
	data.Percentile90 = percentiles[90.0]
	data.Percentile99 = percentiles[99.0]

	data.UniformHist = h.uniformDistribution(histMin, histMax, histSize)
	return data
}

// Just an imperfect approximation for now.
func (h *ExtendedHdrHistogram) uniformDistribution(histMin, histMax, histSize int64) *UniformHistogram {
	hist := NewUniformHistogram(histMin, histMax, histSize)

	for _, bar := range h.hist.Distribution() {
		var v int64
		if bar.From == bar.To {
			v = bar.From
		} else {
			// Attempt to take the median value and record it.
			// How much do we care about integer division, tho?
			v = (bar.To + bar.From) / 2
		}

		hist.RecordValues(v, bar.Count)
	}

	hist.RecordValues(histMin, h.underflowCount)
	hist.RecordValues(histMax, h.overflowCount)

	return hist
}

type OnlineHistogram struct {
	*LockedDoubleBuffer[*ExtendedHdrHistogram]
}

func NewOnlineHistogram(startTime time.Time) *OnlineHistogram {
	return &OnlineHistogram{
		LockedDoubleBuffer: NewLockedDoubleBuffer(func() *ExtendedHdrHistogram {
			return NewExtendedHdrHistogram(startTime)
		}),
	}
}

func (h *OnlineHistogram) RecordValue(v int64) {
	h.SafeActiveWrite(func(hdrHist *ExtendedHdrHistogram) {
		hdrHist.RecordValue(v)
	})
}

func (h *OnlineHistogram) Swap(preSwapCallback func(nonActiveData *ExtendedHdrHistogram)) *ExtendedHdrHistogram {
	return h.LockedDoubleBuffer.Swap(preSwapCallback)
}

type UniformHistogram struct {
	Buckets []hdrhistogram.Bar

	histMin     int64
	histMax     int64
	histSize    int64
	bucketWidth int64
}

func NewUniformHistogram(histMin, histMax, histSize int64) *UniformHistogram {
	h := &UniformHistogram{
		Buckets:     make([]hdrhistogram.Bar, 0, histSize),
		histMin:     histMin,
		histMax:     histMax,
		histSize:    histSize,
		bucketWidth: int64(math.Round(float64(histMax-histMin) / float64(histSize))),
	}

	// TODO: maybe should be an error, but this works for now.
	if h.bucketWidth == 0 {
		panic("bucket width should not be 0, this is likely a programming or config error")
	}

	from := histMin
	for from <= histMax {
		to := from + h.bucketWidth
		h.Buckets = append(h.Buckets, hdrhistogram.Bar{
			From:  from,
			To:    to,
			Count: 0,
		})

		from = to
	}

	return h
}

func (h *UniformHistogram) RecordValues(v int64, count int64) {
	i := h.idx(v)
	h.Buckets[i].Count += count
}

func (h *UniformHistogram) idx(v int64) int64 {
	if v <= h.histMin {
		return 0
	}

	if v >= h.histMax {
		return int64(len(h.Buckets) - 1)
	}

	return (v - h.histMin) / h.bucketWidth
}
