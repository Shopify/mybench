package mybench

import (
	"context"
	"sync"
	"time"
)

// This is the object type that holds the thread-local context data for each
// benchmark worker. Each benchmark worker has its own copy of this. This
// object is passed by the benchmark worker to the workload Event() function
// through the Event() function argument.
//
// The WorkerContext also holds arbitrary data with user-defined types under
// the Data attribute. This allows the user to store custom thread-local data
// on each benchmark for their workloads. A common use case is to store a
// statement object in an user-defined struct.
type WorkerContext[T any] struct {
	// A worker-specific connection object to the database.
	Conn *Connection

	// A worker-specific Rand object. This is needed because the global rand.*
	// functions in Go uses a global mutex underneath the hood and can cause
	// severe performance problems within mybench due to lock contention.
	Rand *Rand

	// User-defined custom data. If no custom data is needed, set T to be
	// mybench.NoContextData.
	Data T

	// A context.Context object used for tracing only. The tracing used is
	// Golang's runtime/trace package. Each Event executed is already in a task
	// called OuterLoopIteration{WorkloadName}. By the time the Event function is
	// called, it is done so inside a trace region called `Event`. Nested regions
	// can be created by the user.
	TraceCtx context.Context
}

// A single goroutine worker that loops and benchmarks MySQL
type BenchmarkWorker[ContextDataT any] struct {
	workloadIface WorkloadInterface[ContextDataT]
	looper        *DiscretizedLooper
	onlineHist    *OnlineHistogram
	context       WorkerContext[ContextDataT]
}

func NewBenchmarkWorker[ContextDataT any](workloadIface WorkloadInterface[ContextDataT], databaseConfig DatabaseConfig, rateControlConfig RateControlConfig) (*BenchmarkWorker[ContextDataT], error) {
	conn, err := databaseConfig.Connection()
	if err != nil {
		return nil, err
	}

	worker := &BenchmarkWorker[ContextDataT]{
		workloadIface: workloadIface,
		context: WorkerContext[ContextDataT]{
			Conn: conn,
			Rand: NewRand(),
		},
	}

	worker.context.Data, err = workloadIface.NewContextData(conn)
	if err != nil {
		return nil, err
	}

	looper := &DiscretizedLooper{
		EventRate:       rateControlConfig.EventRate / float64(rateControlConfig.Concurrency),
		OuterLoopRate:   rateControlConfig.OuterLoopRate,
		DebugIdentifier: workloadIface.Config().Name,
		Event: func(traceCtx context.Context) error {
			worker.context.TraceCtx = traceCtx
			defer func() { worker.context.TraceCtx = nil }()
			return worker.workloadIface.Event(worker.context)
		},
		TraceEvent:     worker.traceEvent,
		TraceOuterLoop: worker.traceOuterLoop,
	}

	// Set it at the end, because the worker.looper is an interface and we cannot
	// customize it like above anymore.
	worker.looper = looper

	return worker, nil
}

func (b *BenchmarkWorker[ContextDataT]) Run(ctx context.Context, workerInitializationWg *sync.WaitGroup, startTime time.Time) error {
	// TODO: kind of weird that the conn is opened in NewBenchmarkWorker but closed here. This should maybe be fixed
	defer b.context.Conn.Close()
	b.onlineHist = NewOnlineHistogram(startTime)
	workerInitializationWg.Done()
	return b.looper.Run(ctx)
}

func (b *BenchmarkWorker[ContextDataT]) traceEvent(stat EventStat) {
	b.onlineHist.RecordValue(stat.TimeTaken.Microseconds())
}

func (b *BenchmarkWorker[ContextDataT]) traceOuterLoop(stat OuterLoopStat) {
	// TODO: Consider detecting loop overruns and back pressure
}
