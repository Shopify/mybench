package mybench

import (
	"context"
	"math/rand"
	"time"
)

type WorkerContext[T any] struct {
	Conn *Connection
	Rand *rand.Rand
	Data T
}

// A single goroutine worker that loops and benchmarks MySQL
type BenchmarkWorker[ContextDataT any] struct {
	workloadIface WorkloadInterface[ContextDataT]
	looper        *DiscretizedLooper
	onlineHist    *OnlineHistogram
	context       WorkerContext[ContextDataT]
}

func NewBenchmarkWorker[ContextDataT any](workloadIface WorkloadInterface[ContextDataT]) (*BenchmarkWorker[ContextDataT], error) {
	config := workloadIface.Config()
	conn, err := config.DatabaseConfig.Connection()
	if err != nil {
		return nil, err
	}

	worker := &BenchmarkWorker[ContextDataT]{
		workloadIface: workloadIface,
		context: WorkerContext[ContextDataT]{
			Conn: conn,
			Rand: rand.New(rand.NewSource(time.Now().Unix())),
		},
	}

	worker.context.Data, err = workloadIface.NewContextData(conn)
	if err != nil {
		return nil, err
	}

	looper := &DiscretizedLooper{
		EventRate:     config.RateControlConfig.EventRate / float64(config.RateControlConfig.Concurrency),
		OuterLoopRate: config.RateControlConfig.OuterLoopRate,
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

func (b *BenchmarkWorker[ContextDataT]) Run(ctx context.Context, startTime time.Time) error {
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
