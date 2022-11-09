package mybench

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Benchmark struct {
	Name        string
	LogInterval time.Duration
	LogRingSize int

	logger    logrus.FieldLogger
	startTime time.Time

	workloads      map[string]AbstractWorkload
	workloadCtx    context.Context
	workloadCancel context.CancelFunc
	workloadWg     *sync.WaitGroup

	dataLogger       *DataLogger
	dataLoggerCtx    context.Context
	dataLoggerCancel context.CancelFunc
	dataLoggerWg     *sync.WaitGroup
}

func NewBenchmark(benchmarkName string, outputFilename string, outputTablename string, note string) (*Benchmark, error) {
	m := &Benchmark{
		Name:         benchmarkName,
		LogInterval:  1 * time.Second,
		workloads:    make(map[string]AbstractWorkload),
		logger:       logrus.WithField("tag", "mcbenchmark"),
		workloadWg:   &sync.WaitGroup{},
		dataLoggerWg: &sync.WaitGroup{},
	}

	m.LogRingSize = int((10*time.Minute)/m.LogInterval) + 1

	m.workloadCtx, m.workloadCancel = context.WithCancel(context.Background())
	m.dataLoggerCtx, m.dataLoggerCancel = context.WithCancel(context.Background())

	var err error
	m.dataLogger, err = NewDataLogger(&DataLogger{
		Interval:       m.LogInterval,
		RingSize:       m.LogRingSize,
		OutputFilename: outputFilename,
		TableName:      outputTablename,
		Note:           note,
		Benchmark:      m,
	})

	return m, err
}

func (b *Benchmark) AddWorkload(workload AbstractWorkload) {
	name := workload.Config().Name
	b.logger.WithField("workload", name).Debug("added workload")
	b.workloads[name] = workload
}

func (b *Benchmark) Start() {
	if !b.startTime.IsZero() {
		panic("already started")
	}

	if len(b.workloads) == 0 {
		panic("no workloads added?")
	}

	b.startTime = time.Now()

	b.workloadWg.Add(len(b.workloads))
	for _, workload := range b.workloads {
		go func(workload AbstractWorkload) {
			defer b.workloadWg.Done()
			workload.Run(b.workloadCtx, b.startTime)
		}(workload)
	}

	b.dataLoggerWg.Add(1)
	go func() {
		defer b.dataLoggerWg.Done()
		b.dataLogger.Run(b.dataLoggerCtx, b.startTime)
	}()
}

func (b *Benchmark) StopAndWait() {
	// Need to stop the benchmark first before cancelling the data logger so the
	// data logger can collect all the data.
	b.workloadCancel()
	b.workloadWg.Wait()

	b.dataLoggerCancel()
	b.dataLoggerWg.Wait()
}

func (b *Benchmark) DataSnapshots() []*DataSnapshot {
	return b.dataLogger.DataSnapshots()
}
