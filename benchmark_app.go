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

type BenchmarkConfig struct {
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

func NewBenchmarkConfig() *BenchmarkConfig {
	config := &BenchmarkConfig{}

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

func (c BenchmarkConfig) Config() BenchmarkConfig {
	return c
}

func (c BenchmarkConfig) Validate() error {
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

// This is the interface that the benchmark application needs to implement
type BenchmarkInterface interface {
	// Returns the name of the benchmark
	Name() string

	// Returns a list of constructed workloads (or an error) for this benchmark.
	Workloads() ([]AbstractWorkload, error)

	// This function is called if the benchmark ran with the --load flag. This
	// should load the database when called.
	RunLoader() error

	// Returns the benchmark app configuration. The configuration returned must
	// have already been validated with defaults filled in. Make sure to call
	// .Validate() on it during the construction of the object that implements
	// this interface.
	//
	// If you implement BenchmarkInterface with BenchmarkConfig embedded in it,
	// you won't need to define this method as the BenchmarkConfig object already
	// defines this method for you, and embedding it in your struct will cause
	// your struct to inherit the method implemented on BenchmarkConfig.
	Config() BenchmarkConfig
}

// Runs a custom defined benchmark that implements the BenchmarkInterface.
func Run(benchmarkInterface BenchmarkInterface) error {
	config := benchmarkInterface.Config()
	err := config.Validate()
	if err != nil {
		return err
	}

	// Creates the database if needed
	if !config.DatabaseConfig.NoConnection {
		err := config.DatabaseConfig.CreateDatabaseIfNeeded()
		if err != nil {
			logrus.WithError(err).Error("cannot create new database")
			return err
		}
	}

	// Run the loader if configured to do so
	if config.Load {
		return benchmarkInterface.RunLoader()
	}

	// Construct and run the benchmark
	benchmark, err := NewBenchmark(
		benchmarkInterface.Name(),
		config.LogFile,
		config.LogTable,
		config.Note,
		config.HttpPort,
	)
	if err != nil {
		return err
	}

	workloads, err := benchmarkInterface.Workloads()
	if err != nil {
		return err
	}

	for _, workload := range workloads {
		benchmark.AddWorkload(workload)
	}

	benchmark.Start()

	// Handle stopping of the benchmarks via either signals or timers (if duration
	// is configured).
	quitCh := make(chan struct{})

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

		s := <-c
		logrus.WithField("signal", s.String()).Warn("received termination signal")
		benchmark.StopAndWait()
		logrus.Info("benchmark stopped")
		close(quitCh)
	}()

	if config.Duration > time.Duration(0) {
		logrus.Infof("running benchmark for %v", config.Duration)
		go func() {
			time.Sleep(config.Duration)
			benchmark.StopAndWait()
			close(quitCh)
		}()
	} else {
		logrus.Info("running benchmark indefinitely")
	}

	<-quitCh

	return nil
}
