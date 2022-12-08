package mybench

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Benchmark struct {
	BenchmarkConfig

	Name        string
	LogInterval time.Duration // TODO: make these two values configurable (move into BenchmarkConfig maybe)
	LogRingSize int

	logger    logrus.FieldLogger
	startTime time.Time

	workloads      map[string]AbstractWorkload
	workloadCtx    context.Context
	workloadCancel context.CancelFunc
	workloadWg     *sync.WaitGroup

	// Auxiliary services are below
	dataLogger       *DataLogger
	dataLoggerCtx    context.Context
	dataLoggerCancel context.CancelFunc
	dataLoggerWg     *sync.WaitGroup

	httpServer *HttpServer
}

func NewBenchmark(benchmarkName string, benchmarkConfig BenchmarkConfig) (*Benchmark, error) {
	b := &Benchmark{
		BenchmarkConfig: benchmarkConfig,
		Name:            benchmarkName,
		LogInterval:     1 * time.Second,
		workloads:       make(map[string]AbstractWorkload),
		logger:          logrus.WithField("tag", "benchmark").WithField("benchmark", benchmarkName),
		workloadWg:      &sync.WaitGroup{},
		dataLoggerWg:    &sync.WaitGroup{},
	}

	b.LogRingSize = int((10*time.Minute)/b.LogInterval) + 1

	b.workloadCtx, b.workloadCancel = context.WithCancel(context.Background())
	b.dataLoggerCtx, b.dataLoggerCancel = context.WithCancel(context.Background())

	var err error
	b.dataLogger, err = NewDataLogger(&DataLogger{
		Interval:       b.LogInterval,
		RingSize:       b.LogRingSize,
		OutputFilename: benchmarkConfig.LogFile,
		TableName:      benchmarkConfig.LogTable,
		Note:           benchmarkConfig.Note,
		Benchmark:      b,
	})

	b.httpServer = NewHttpServer(b, benchmarkConfig.Note, benchmarkConfig.HttpPort)

	return b, err
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
		// Calculate per workload rate control config by scaling it with workload scale
		var perWorkloadRateControlConfig RateControlConfig = b.BenchmarkConfig.RateControlConfig // take a copy
		workloadScale := workload.Config().WorkloadScale

		perWorkloadRateControlConfig.Concurrency = int(math.Ceil(float64(b.BenchmarkConfig.RateControlConfig.Concurrency) * workloadScale))
		perWorkloadRateControlConfig.EventRate = b.BenchmarkConfig.RateControlConfig.EventRate * workloadScale

		// We need to set the RateControlConfig on the workload because the
		// DataLogger needs to know the concurrency and event rate of each workload.
		// We can't pass the RateControlConfig to workload.Run and set it in there,
		// as doing so would introduce a data race: the workload.Run is called in a
		// goroutine and DataLogger.Run is called in another goroutine. It is thus
		// possible that DataLogger.Run will try to fetch the RateControlConfig on
		// the Workload before it is set.
		workload.FinishInitialization(b.BenchmarkConfig.DatabaseConfig, perWorkloadRateControlConfig)

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

	go func() {
		b.httpServer.Run()
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
