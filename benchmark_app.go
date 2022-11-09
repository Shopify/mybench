package mybench

import (
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type BenchmarkAppConfig struct {
	DatabaseConfig DatabaseConfig

	Bench bool
	Load  bool

	LoadConcurrency int
	HttpPort        int
	Duration        time.Duration
	Multiplier      float64
	LogFile         string
	LogTable        string
	Note            string
}

type ConfigType interface {
	// Have to do the following because of this bug: https://github.com/golang/go/issues/48522
	// Perhaps this code shouldn't even use generic and it can just use regular interface..., but what's the fun in that?! /s
	GetCommonBenchmarkConfig() BenchmarkAppConfig
	Validate() error
}

func NewBenchmarkAppConfig() *BenchmarkAppConfig {
	config := &BenchmarkAppConfig{}

	flag.StringVar(&config.DatabaseConfig.Host, "host", "", "database host name")
	flag.IntVar(&config.DatabaseConfig.Port, "port", 3306, "database port (default: 3306)")
	flag.StringVar(&config.DatabaseConfig.User, "user", "root", "database user (default: root)")
	flag.StringVar(&config.DatabaseConfig.Pass, "pass", "", "database password (default: empty)")
	flag.StringVar(&config.DatabaseConfig.Database, "db", "mybench", "database name (default: mybench)")

	flag.BoolVar(&config.Load, "load", false, "load the data before the benchmark")
	flag.BoolVar(&config.Bench, "bench", false, "run the benchmark")
	flag.IntVar(&config.HttpPort, "httpport", 8005, "port of the monitoring UI")
	flag.DurationVar(&config.Duration, "duration", 0, "duration of the benchmark")
	flag.Float64Var(&config.Multiplier, "multiplier", 1.0, "multiplier of the benchmark")
	flag.StringVar(&config.LogFile, "log", "data.sqlite", "the path to the log file")
	flag.StringVar(&config.LogTable, "logtable", "", "the table name in the sqlite file to record to (default: based on the start time in RFC3399)")
	flag.StringVar(&config.Note, "note", "", "a note to include in the meta table entry for this run")

	flag.IntVar(&config.LoadConcurrency, "load-concurrency", 16, "the concurrency to use during the load")

	return config
}

func (c BenchmarkAppConfig) GetCommonBenchmarkConfig() BenchmarkAppConfig {
	return c
}

func (c BenchmarkAppConfig) Validate() error {
	if c.Bench == c.Load {
		return errors.New("must only specify one of -bench or -load")
	}

	if c.DatabaseConfig.Host == "" {
		return errors.New("must specify -host")
	}

	if c.LogFile == "" {
		return errors.New("must specify log filename")
	}

	return nil
}

// Defines a benchmark application with common command-line flags.
//
// This application is a generic that takes two different types:
//
//  1. ConfigT: this is a type for a custom configuration object, where custom
//     configuration data can be defined for the entire benchmark application.
//     An example of data stored in this type could be a configuration variable
//     that controls the event rate of a particular workload (if the benchmark
//     application defines multiple workloads).
//  2. ContextDataT: this is a type for a custom goroutine-specific data object,
//     where custom data can be stored. This data is created once per goroutine
//     (via the WorkloadInterface.NewContextData method) and can be
//     accessed in the WorkerContext object passed to the
//     WorkloadInterface.Event method. An example of data stored in this
//     type could be a prepared statement. Prepared statements must not be
//     reused across multiple threads, so each goroutine must create its own
//     prepared statements.
type BenchmarkApp[ConfigT ConfigType] struct {
	Config ConfigT

	// Custom benchmark setup code
	// This is where you create the workloads.
	setupBenchmark func(*BenchmarkApp[ConfigT]) error

	runLoader func(*BenchmarkApp[ConfigT]) error

	benchmark  *Benchmark
	httpServer *HttpServer
}

func NewBenchmarkApp[ConfigT ConfigType](benchmarkName string, config ConfigT, setupBenchmark func(*BenchmarkApp[ConfigT]) error, runLoader func(*BenchmarkApp[ConfigT]) error) (*BenchmarkApp[ConfigT], error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	if !config.GetCommonBenchmarkConfig().DatabaseConfig.NoConnection {
		err = config.GetCommonBenchmarkConfig().DatabaseConfig.CreateDatabaseIfNeeded()
		if err != nil {
			logrus.WithError(err).Panic("cannot create new database")
			return nil, err
		}
	}

	// The code constructing the Benchmark is not ideal (messy) but works.
	benchmark, err := NewBenchmark(
		benchmarkName,
		config.GetCommonBenchmarkConfig().LogFile,
		config.GetCommonBenchmarkConfig().LogTable,
		config.GetCommonBenchmarkConfig().Note,
	)
	if err != nil {
		return nil, err
	}

	return &BenchmarkApp[ConfigT]{
		Config:         config,
		setupBenchmark: setupBenchmark,
		runLoader:      runLoader,
		benchmark:      benchmark,
		httpServer:     NewHttpServer(benchmark, config.GetCommonBenchmarkConfig().HttpPort),
	}, nil
}

func (a *BenchmarkApp[ConfigT]) AddWorkload(workload AbstractWorkload) {
	a.benchmark.AddWorkload(workload)
}

func (a *BenchmarkApp[ConfigT]) Run() error {
	commonConfig := a.Config.GetCommonBenchmarkConfig()
	if commonConfig.Load {
		return a.RunLoader()
	}

	return a.RunBenchmark(commonConfig.Duration)
}

func (a *BenchmarkApp[ConfigT]) RunLoader() error {
	return a.runLoader(a)
}

func (a *BenchmarkApp[ConfigT]) RunBenchmark(duration time.Duration) error {
	err := a.setupBenchmark(a)
	if err != nil {
		return err
	}

	a.benchmark.Start()

	go func() {
		a.httpServer.Run()
	}()

	quitCh := make(chan struct{})

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

		s := <-c
		logrus.WithField("signal", s.String()).Warn("received termination signal")
		a.benchmark.StopAndWait()
		logrus.Info("benchmark stopped")
		close(quitCh)
	}()

	if duration > time.Duration(0) {
		logrus.Infof("running benchmark for %v", duration)
		go func() {
			time.Sleep(duration)
			a.benchmark.StopAndWait()
			close(quitCh)
		}()
	} else {
		logrus.Info("running benchmark indefinitely")
	}

	<-quitCh
	return nil
}
