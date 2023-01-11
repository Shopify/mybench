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
	// goroutine cannot drive the EventRate.  If not specified, it will be
	// calculated from EventRate and MaxEventRatePerWorker.
	Concurrency int

	// The maximum event rate per goroutine, used to calculate Concurrency.
	// If not specified, it will default to 100.
	MaxEventRatePerWorker float64

	// The desired rate of the outer loop that batches events. Default: 50.
	OuterLoopRate float64

	// The type of looper used. Default to Uniform looper.
	LooperType LooperType

	// If set, per-workload event rate will not be rounded to the nearest
	// integer. Per-workload event rate is rounded because the total event rate
	// as specified by EventRate is multiplied by WorkloadScale. This floating
	// point calculation can result in numbers like 27839.999999999996.
	// Maintaining such a precise value is difficult for the looper, which can
	// cause erratic looking results.
	DoNotRoundEventRate bool
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

	// scales the workload by the given percentage
	// this currently scales various RateControl parameters
	WorkloadScale float64

	// Configures the visualization for this workload.
	// Some workload may know the latency bounds, and may wish to choose a better scale for the histograms
	Visualization VisualizationConfig
}

func (w WorkloadConfig) Config() WorkloadConfig {
	return w
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

	databaseConfig    DatabaseConfig
	rateControlConfig RateControlConfig
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
	Run(context.Context, *sync.WaitGroup, time.Time)
	Config() WorkloadConfig

	// We need to set the RateControlConfig on the Workload object, because the
	// data logger needs to know the concurrency/event rate of the workload.
	// See comments in Benchmark.Start for more details.
	FinishInitialization(DatabaseConfig, RateControlConfig)

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

	// The DataLogger needs the rate control config to make allocations and record
	// desired event rates. See comments in Benchmark.Start for more details.
	RateControlConfig() RateControlConfig
}

func NewWorkload[ContextDataT any](workloadIface WorkloadInterface[ContextDataT]) *Workload[ContextDataT] {
	c := workloadIface.Config()
	if c.Name == "" {
		panic("must specify workload name")
	}

	if c.WorkloadScale < 0 || c.WorkloadScale > 1 {
		panic("WorkloadScale must be between 0 and 1")
	} else if c.WorkloadScale == 0 { // if not specified, default to 1
		c.WorkloadScale = 1
	}

	// Apply Visualization defaults as needed
	if c.Visualization.LatencyHistMin == 0 {
		c.Visualization.LatencyHistMin = 0
	}
	if c.Visualization.LatencyHistMax == 0 {
		c.Visualization.LatencyHistMax = 50000
	}
	if c.Visualization.LatencyHistSize == 0 {
		c.Visualization.LatencyHistSize = 1000
	}

	return &Workload[ContextDataT]{
		WorkloadConfig: c,
		workloadIface:  workloadIface,
		workersWg:      &sync.WaitGroup{},
		logger:         logrus.WithField("workload", c.Name),
	}
}

func (w *Workload[ContextDataT]) FinishInitialization(databaseConfig DatabaseConfig, rateControlConfig RateControlConfig) {
	w.databaseConfig = databaseConfig
	w.rateControlConfig = rateControlConfig

	if !w.rateControlConfig.DoNotRoundEventRate {
		w.rateControlConfig.EventRate = math.RoundToEven(w.rateControlConfig.EventRate)
	}
}

func (w *Workload[ContextDataT]) Run(ctx context.Context, workerInitializationWg *sync.WaitGroup, startTime time.Time) {
	var err error
	w.workers = make([]*BenchmarkWorker[ContextDataT], w.rateControlConfig.Concurrency)
	for i := 0; i < w.rateControlConfig.Concurrency; i++ {
		w.workers[i], err = NewBenchmarkWorker(w.workloadIface, w.databaseConfig, w.rateControlConfig)
		if err != nil {
			panic(err)
		}
	}

	w.workersWg.Add(w.rateControlConfig.Concurrency)
	w.logger.WithFields(logrus.Fields{
		"concurrency": w.rateControlConfig.Concurrency,
		"rate":        w.rateControlConfig.EventRate,
	}).Info("starting benchmark workers")

	for i := 0; i < w.rateControlConfig.Concurrency; i++ {
		go func(worker *BenchmarkWorker[ContextDataT]) {
			defer w.workersWg.Done()
			err := worker.Run(ctx, workerInitializationWg, startTime, w.databaseConfig, w.rateControlConfig)
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

func (w *Workload[ContextDataT]) RateControlConfig() RateControlConfig {
	return w.rateControlConfig
}

func (w *Workload[ContextDataT]) ForEachOnlineHistogram(f func(int, *OnlineHistogram)) {
	for i, worker := range w.workers {
		f(i, worker.onlineHist)
	}
}
