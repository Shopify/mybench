package mybench

import (
	"context"
	"time"
)

type WorkerContext[T any] struct {
	Conn *Connection
	Rand *Rand
	Data T
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
		EventRate:     rateControlConfig.EventRate / float64(rateControlConfig.Concurrency),
		OuterLoopRate: rateControlConfig.OuterLoopRate,
		Event: func() error {
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

func (b *BenchmarkWorker[ContextDataT]) Run(ctx context.Context, startTime time.Time, databaseConfig DatabaseConfig, rateControlConfig RateControlConfig) error {
	// TODO: kind of weird that the conn is opened in NewBenchmarkWorker but closed here. This should maybe be fixed
	defer b.context.Conn.Close()
	b.onlineHist = NewOnlineHistogram(startTime)
	return b.looper.Run(ctx)
}

func (b *BenchmarkWorker[ContextDataT]) traceEvent(stat EventStat) {
	b.onlineHist.RecordValue(stat.TimeTaken.Microseconds())
}

func (b *BenchmarkWorker[ContextDataT]) traceOuterLoop(stat OuterLoopStat) {
	// TODO: Consider detecting loop overruns and back pressure
}
