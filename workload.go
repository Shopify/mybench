package mybench

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type RateControlConfig struct {
	// The total rate at which Event is called in hz, across all goroutines.
	EventRate float64

	// Number of goroutines to drive the EventRate. Each goroutine will get the
	// same portion of the EventRate. This needs to be increased if a single
	// goroutine cannot drive the EventRate.
	// If not specified, it will be set to an approximately optimal number of MySQL threads (100).
	Concurrency int

	// The maximum event per goroutine. Calculated from EventRate and Concurrency if it is
	// not specified.
	MaxEventRatePerGoroutine int

	// The desired rate of the outer loop that batches events. Default: 50.
	OuterLoopRate float64

	// The type of looper used. Default to Uniform looper.
	LooperType LooperType
}

type VisualizationConfig struct {
	// The min, max, and size of the visualization histogram that tracks the
	// per-event latency. This is not used to actually track the latency as the code
	// internally uses HDR histogram to track a much wider range. Instead, this is
	// used for the final data logging and web UI display.
	// The min and max has units of microseconds and defaults to 0 and 100000.
	LatencyHistMin  int64
	LatencyHistMax  int64
	LatencyHistSize int64
}

// Config used to create the Workload
type WorkloadConfig struct {
	// The name of the workload, for identification purposes only.
	Name string

	// The database config used to create connection objects.
	DatabaseConfig DatabaseConfig

	// scales the workload by the given percentage
	// this currently scales various RateControl parameters
	WorkloadPctScale int

	// Controls the event rate
	RateControl RateControlConfig

	// Configures the visualization for this workload.
	// Some workload may know the latency bounds, and may wish to choose a better scale for the histograms
	Visualization VisualizationConfig
}

func (w WorkloadConfig) Config() WorkloadConfig {
	return w
}

func NewWorkloadConfigWithDefaults(c WorkloadConfig) WorkloadConfig {
	if c.RateControl.EventRate == 0 {
		c.RateControl.EventRate = 1000
	}

	if c.RateControl.Concurrency == 0 {
		c.RateControl.Concurrency = 100
	}

	if c.RateControl.OuterLoopRate == 0 {
		c.RateControl.OuterLoopRate = 50
	}

	if c.WorkloadPctScale > 0 {
		c.RateControl.Concurrency = c.RateControl.Concurrency * c.WorkloadPctScale / 100
		c.RateControl.EventRate = c.RateControl.EventRate * float64(c.WorkloadPctScale) / 100
	}

	// We don't want less than one event/s per worker.
	if c.RateControl.Concurrency > int(c.RateControl.EventRate) {
		c.RateControl.Concurrency = int(c.RateControl.EventRate)
	}

	if c.RateControl.MaxEventRatePerGoroutine == 0 {
		c.RateControl.MaxEventRatePerGoroutine = int(math.Ceil(c.RateControl.EventRate / float64(c.RateControl.Concurrency)))
	}

	if c.Visualization.LatencyHistMin == 0 {
		c.Visualization.LatencyHistMin = 0
	}

	if c.Visualization.LatencyHistMax == 0 {
		c.Visualization.LatencyHistMax = 50000
	}

	if c.Visualization.LatencyHistSize == 0 {
		c.Visualization.LatencyHistSize = 1000
	}

	if c.Name == "" {
		panic("must specify Name")
	}

	return c
}

// An interface for implementing the workload. Although only one Workload struct
// exists, the functions defined in this can be called concurrently from many
// (thousands) goroutines. Care must be taken to ensure no data races (use
// atomic or mutexes, although the latter could be slow as the number of
// goroutines increase)
//
// A example of a workload could be selecting data from a single table. Another
// workload could be inserting into the same table. This is similar in meaning
// as the "class" Section 6.2 in the paper: "OLTP-Bench: An Extensible Testbed
// for Benchmarking Relational Databases"
// https://www.cs.cmu.edu/~pavlo/papers/oltpbench-vldb.pdf.
//
// Each benchmark supports multiple workloads that can be scaled and reported on
// seperately allowing for evolving workload mixtures.
type WorkloadInterface[ContextDataT any] interface {
	// Returns the config for the workload. Should always return the same
	// value.
	//
	// If the struct that implements this interface embeds WorkloadConfig, it
	// does not need to manually define this method as WorkloadConfig already
	// defines this method and structs that embeds WorkloadConfig will inherit
	// that method.
	Config() WorkloadConfig

	// Run the code for a single event. Run your queries here. This function is
	// called repeatedly at the desired rate.
	//
	// This function can be called concurrently from many goroutines. Each
	// goroutine will pass a different connection object to this function, so the
	// connection should be safe to use normally. However, if you rely on any
	// variables set in the Prepare function, make sure it is thread safe but also
	// performant.
	Event(WorkerContext[ContextDataT]) error

	// Since the Event() function can be called from multiple goroutines, there is
	// no convenient place to store goroutine-specific data. Storing it on the
	// WorkloadInterface won't work because that object is shared
	// between all BenchmarkWorkers (and hence goroutines and threads) .
	//
	// Using the WorkerContext.Data field, each BenchmarkWorker can store its own
	// goroutine-local data. This is passed to the Event() function call on that
	// goroutine only, which means use of the data is thread-safe, as it is not
	// accessed from any other threads.
	//
	// This method creates a new context data object and it is called once per
	// goroutine/BenchmarkWorker before the main event loop.
	//
	// If a workload does not need context data, the struct that implements
	// workload interface should simply embed NoContextData, which defines this
	// method.
	NewContextData(*Connection) (ContextDataT, error)
}

// This is a convenience type defined to indicate that there's no context data.
// If you don't need a custom context data for your workload, use this.
//
// Further, the workload should embed this struct so the NewContextData
// function does not need to be defined
type NoContextData struct{}

// Any struct that embeds NoContextData will inherit this method. If this
// struct is also trying to implement WorkloadInterface, then it does not have
// to define this method manually.
func (NoContextData) NewContextData(*Connection) (NoContextData, error) {
	return NoContextData{}, nil
}

// The actual benchmark struct for a single workload.
type Workload[ContextDataT any] struct {
	WorkloadConfig

	workloadIface WorkloadInterface[ContextDataT]
	workers       []*BenchmarkWorker[ContextDataT]

	workersWg *sync.WaitGroup
	logger    logrus.FieldLogger
}

// We want the workload to be templated so the context data can be transparently
// passed from the workload to the Event() function without going through
// runtime type selection.
//
// However, we don't want to store a templated workload in the Benchmark object,
// as it would start to infect every other struct (like DataLogger). Further,
// since the Workload is stored in the Benchmark in a map, if it is templated,
// you cannot have different ContextDataT on a per workload basis if you wanted
// it.
//
// Is this kind of a hack? Maybe. I haven't decided yet.
type AbstractWorkload interface {
	Run(context.Context, time.Time)
	Config() WorkloadConfig

	// The DataLogger need to iterate through all the OnlineHistograms for each
	// worker so it can perform the double buffer swap. Since the DataLogger only
	// have access to a map of AbstractWorkload, it doesn't have access to the
	// underlying BenchmarkWorker array (which are templated). So this
	// AbstractWorkload needs to provide a way to iterate through all the
	// OnlineHistograms. The naive way to implement this would be by returning a
	// slice of OnlineHistograms. This is not a good approach, because it would
	// allocate memory, which can take unbounded time. Since the code that
	// iterates through all the online histograms are in a "critical" section that
	// takes as small amount of a time as possible, this is not acceptable. Thus,
	// basically create an iterator. Since Golang doesn't have a built-in iterator
	// pattern, the functional pattern is chosen.
	// TODO: check that the functional pattern doesn't introduce unnecessary
	// overhead/memory allocations.
	ForEachOnlineHistogram(func(int, *OnlineHistogram))
}

func NewWorkload[ContextDataT any](workloadIface WorkloadInterface[ContextDataT]) (*Workload[ContextDataT], error) {
	c := workloadIface.Config()
	workload := &Workload[ContextDataT]{
		WorkloadConfig: c,
		workloadIface:  workloadIface,
		workersWg:      &sync.WaitGroup{},
		logger:         logrus.WithField("workload", c.Name),
	}

	var err error
	workload.workers = make([]*BenchmarkWorker[ContextDataT], workload.RateControl.Concurrency)
	for i := 0; i < workload.RateControl.Concurrency; i++ {
		workload.workers[i], err = NewBenchmarkWorker(workloadIface)
		if err != nil {
			return nil, err
		}
	}

	return workload, nil
}

func (w *Workload[ContextDataT]) Run(ctx context.Context, startTime time.Time) {
	w.workersWg.Add(w.RateControl.Concurrency)
	w.logger.WithFields(logrus.Fields{"concurrency": w.RateControl.Concurrency, "rate": w.RateControl.EventRate}).Info("starting benchmark workers")
	for i := 0; i < w.RateControl.Concurrency; i++ {
		go func(worker *BenchmarkWorker[ContextDataT]) {
			defer w.workersWg.Done()
			err := worker.Run(ctx, startTime)
			if err != nil {
				w.logger.WithError(err).Panic("failed to run worker")
			}
		}(w.workers[i])
	}

	w.workersWg.Wait()
}

func (w *Workload[ContextDataT]) Config() WorkloadConfig {
	return w.WorkloadConfig
}

func (w *Workload[ContextDataT]) ForEachOnlineHistogram(f func(int, *OnlineHistogram)) {
	for i, worker := range w.workers {
		f(i, worker.onlineHist)
	}
}
