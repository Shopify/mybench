package mybench

import (
	"errors"
	"flag"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type BenchmarkConfig struct {
	Bench    bool
	Load     bool
	Duration time.Duration
	LogFile  string
	LogTable string
	Note     string

	DatabaseConfig DatabaseConfig

	RateControlConfig RateControlConfig

	HttpPort int
}

func NewBenchmarkConfig() *BenchmarkConfig {
	config := &BenchmarkConfig{}

	flag.BoolVar(&config.Load, "load", false, "load the data before the benchmark")
	flag.BoolVar(&config.Bench, "bench", false, "run the benchmark")
	flag.DurationVar(&config.Duration, "duration", 0, "duration of the benchmark")
	flag.StringVar(&config.LogFile, "log", "data.sqlite", "the path to the log file")
	flag.StringVar(&config.LogTable, "logtable", "", "the table name in the sqlite file to record to (default: based on the start time in RFC3399)")
	flag.StringVar(&config.Note, "note", "", "a note to include in the meta table entry for this run")

	flag.StringVar(&config.DatabaseConfig.Host, "host", "", "database host name")
	flag.IntVar(&config.DatabaseConfig.Port, "port", 3306, "database port (default: 3306)")
	flag.StringVar(&config.DatabaseConfig.User, "user", "root", "database user (default: root)")
	flag.StringVar(&config.DatabaseConfig.Pass, "pass", "", "database password (default: empty)")
	flag.StringVar(&config.DatabaseConfig.Database, "db", "mybench", "database name (default: mybench)")
	flag.IntVar(&config.DatabaseConfig.ConnectionMultiplier, "connectionmultiplier", 1, "number of database connections per parallel worker (default: 1)")
	flag.BoolVar(&config.DatabaseConfig.ClientMultiStatements, "clientmultistatements", false, "Enable support for CLIENT_MULTI_STATEMENTS on the connection")

	flag.Float64Var(&config.RateControlConfig.EventRate, "eventrate", 1000, "target event rate of the benchmark in requests per second (default: 1000)")
	flag.IntVar(&config.RateControlConfig.Concurrency, "concurrency", 0, "number of parallel workers to use during the benchmark (default: auto)")
	flag.Float64Var(&config.RateControlConfig.MaxEventRatePerWorker, "workermaxrate", 100, "maximum event rate per worker (default: 100)")
	flag.Float64Var(&config.RateControlConfig.OuterLoopRate, "outerlooprate", 50, "desired rate of outer loop that batches events -- advanced option (default: 50)")

	flag.IntVar(&config.HttpPort, "httpport", 8005, "port of the monitoring UI")

	return config
}

func (c *BenchmarkConfig) Config() BenchmarkConfig {
	return *c
}

func (c *BenchmarkConfig) ValidateAndSetDefaults() error {
	if c.Bench == c.Load {
		return errors.New("must only specify one of -bench or -load")
	}

	if c.DatabaseConfig.Host == "" {
		return errors.New("must specify -host")
	}

	if c.LogFile == "" {
		return errors.New("must specify log filename")
	}

	if c.DatabaseConfig.ConnectionMultiplier != 1 && c.RateControlConfig.Concurrency == 0 {
		return errors.New("must specify -concurrency if -connectionmultiplier is specified")
	}

	// Apply RateControlConfig defaults as needed
	if c.RateControlConfig.EventRate == 0 {
		c.RateControlConfig.EventRate = 1000
	}

	if c.RateControlConfig.MaxEventRatePerWorker == 0 {
		c.RateControlConfig.MaxEventRatePerWorker = 100
	}

	if c.RateControlConfig.OuterLoopRate == 0 {
		c.RateControlConfig.OuterLoopRate = 50
	}

	// If Concurrency is not specified, calculate it from EventRate and MaxEventRatePerWorker
	if c.RateControlConfig.Concurrency == 0 {
		if c.DatabaseConfig.ConnectionMultiplier > 1 {
			return errors.New("if ConnectionMultiplier is specified, Concurrency must be specified")
		}

		c.RateControlConfig.Concurrency = int(math.Ceil(c.RateControlConfig.EventRate / float64(c.RateControlConfig.MaxEventRatePerWorker)))
	}

	if c.RateControlConfig.Concurrency > int(c.RateControlConfig.EventRate) {
		c.RateControlConfig.Concurrency = int(c.RateControlConfig.EventRate)
		logrus.Warnf("Concurrency is too high for the given EventRate. Reducing Concurrency to %d", c.RateControlConfig.Concurrency)
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
	err := config.ValidateAndSetDefaults()
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
	benchmark, err := NewBenchmark(benchmarkInterface.Name(), config)
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
