package mybench

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/trace"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

const createMetaTableStatement = `
CREATE TABLE IF NOT EXISTS meta (
	table_name TEXT PRIMARY KEY,
	benchmark_name TEXT,
	mybench_version TEXT,
	note TEXT,
	start_time TEXT,
	end_time TEXT
)
`

var VersionString = "1.0"

var insertMetaStatement = "INSERT INTO meta (table_name, note, benchmark_name, mybench_version, start_time) VALUES (?, ?, ?, '" + VersionString + "', ?)"

const updateMetaEndTimeStatement = `
UPDATE meta SET end_time = ? WHERE table_name = ?
`

const createTableStatement = `
CREATE TABLE %s (
	id INTEGER PRIMARY KEY AUTOINCREMENT, -- Using auto increment allows the id to be strictly monotonic
	workload TEXT,
	seconds_since_start REAL,
	interval_start TEXT,
	interval_end TEXT,
	desired_rate REAL,
	count INTEGER,
	delta REAL,
	rate REAL,
	min INTEGER,
	mean INTEGER,
	max INTEGER,
	underflow_count INTEGER,
	overflow_count INTEGER,
	percentile25 INTEGER,
	percentile50 INTEGER,
	percentile75 INTEGER,
	percentile90 INTEGER,
	percentile99 INTEGER,
	uniform_hist TEXT
);
CREATE INDEX %s_workload ON %s(workload);
`

// TODO: insert uniform_hist  as well
const insertQuery = `
INSERT INTO %s (
	workload,
	seconds_since_start,
	interval_start,
	interval_end,
	desired_rate,
	count,
	delta,
	rate,
	min,
	mean,
	max,
	underflow_count,
	overflow_count,
	percentile25,
	percentile50,
	percentile75,
	percentile90,
	percentile99
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

// Merges the IntervalData with other data.
// Represents all the stats collected for a single workload.
type WorkloadDataSnapshot struct {
	// Throughput and latency data
	IntervalData

	// The desired throughput
	DesiredRate float64
}

func (s WorkloadDataSnapshot) queryArgs(workload string, secondsSinceStart float64) []interface{} {
	return []interface{}{
		workload,
		secondsSinceStart,
		s.StartTime.Format(time.RFC3339),
		s.EndTime.Format(time.RFC3339),
		s.DesiredRate,
		s.Count,
		s.Delta,
		s.Rate,
		s.Min,
		s.Mean,
		s.Max,
		s.UnderflowCount,
		s.OverflowCount,
		s.Percentile25,
		s.Percentile50,
		s.Percentile75,
		s.Percentile90,
		s.Percentile99,
	}
}

type DataSnapshot struct {
	// Time since start of the test.
	Time float64

	// The throughput and latency data for all monitored benchmarks merged
	// together.
	AllWorkloadData WorkloadDataSnapshot

	// The throughput and latency data for individual monitored benchmarks,
	// indexed by the workload name.
	PerWorkloadData map[string]WorkloadDataSnapshot
}

type DataLogger struct {
	Interval       time.Duration
	RingSize       int
	OutputFilename string
	TableName      string
	Note           string
	Benchmark      *Benchmark

	logger    logrus.FieldLogger
	startTime time.Time
	db        *sql.DB
	dataRing  *Ring[*DataSnapshot]
}

func NewDataLogger(dataLogger *DataLogger) (*DataLogger, error) {
	dataLogger.logger = logrus.WithField("tag", "data_logger")

	if dataLogger.OutputFilename == "" {
		return nil, errors.New("must specify output filename for data logger")
	}

	return dataLogger, nil
}

func (d *DataLogger) Run(ctx context.Context, startTime time.Time) {
	d.startTime = startTime
	if d.TableName == "" {
		d.TableName = "T" + startTime.Format(time.RFC3339) // Need to start table name with a character.
		d.TableName = strings.Replace(d.TableName, ":", "_", -1)
		d.TableName = strings.Replace(d.TableName, "-", "_", -1)
	}

	err := d.initializeLogDatabase()
	if err != nil {
		logrus.WithError(err).Panic("failed to initialize log database")
	}
	defer d.closeLogDatabase()

	// Initialize the ring
	d.dataRing = NewRing[*DataSnapshot](d.RingSize)

	// Start collecting data!
	nextWakeupTime := time.Now().Add(d.Interval)
	for {
		now := time.Now()
		delta := nextWakeupTime.Sub(now)
		if delta <= time.Duration(0) {
			d.logger.WithField(
				"delta", delta,
			).Error("data logger not logging with sufficient performance, check if the CollectData region is taking too long and increase the data logging interval")
		}

		select {
		case <-ctx.Done():
			// Some data might be lost at the end, but this is OK because during
			// termination, the goroutines may terminate over a significant period of
			// time, significantly skewing the rate calculations.
			return
		case <-time.After(delta):
			d.collectAndLogData()
		}

		nextWakeupTime = nextWakeupTime.Add(d.Interval)
	}
}

func (d *DataLogger) DataSnapshots() []*DataSnapshot {
	return d.dataRing.ReadAllOrdered()
}

func (d *DataLogger) initializeLogDatabase() error {
	var err error
	d.db, err = sql.Open("sqlite3", d.OutputFilename)
	if err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(createMetaTableStatement)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(fmt.Sprintf(createTableStatement, d.TableName, d.TableName, d.TableName))
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(insertMetaStatement, d.TableName, d.Note, d.Benchmark.Name, d.startTime.Format(time.RFC3339))
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	d.logger.Infof("using log data file at %s with table %s", d.OutputFilename, d.TableName)
	return nil
}

func (d *DataLogger) closeLogDatabase() error {
	defer d.db.Close()
	_, err := d.db.Exec(updateMetaEndTimeStatement, time.Now().Format(time.RFC3339), d.TableName)
	return err
}

func (d *DataLogger) collectAndLogData() {
	dataSnapshot := d.collectData()
	d.logData(dataSnapshot)
}

func (d *DataLogger) collectData() *DataSnapshot {
	ctx, task := trace.NewTask(context.Background(), "CollectData")
	defer task.End()

	now := time.Now()
	dataSnapshot := &DataSnapshot{
		Time:            now.Sub(d.startTime).Seconds(),
		PerWorkloadData: make(map[string]WorkloadDataSnapshot),
	}

	// We need to take a snapshot of all current histograms. A snapshot implies it
	// must be done extremely quickly. Specifically, from the moment we get the
	// histogram from the first worker and the moment we get the histogram from
	// the last worker must be very quick to avoid smearing effects when
	// calculating the rates (~10ms at most, if 40000 events/s is expected, 10ms
	// would be 400 events).
	//
	// This is why we use a double buffer swap. Specifically, we first swap, which
	// should be quick. Then we read, which may take a lot longer, but it is okay
	// as we have a snapshot of the data. Lastly, we reset the histogram here, so
	// it can be swapped again for the next time and the workers will write to a 0
	// initialized histogram again.
	//
	// In this architecture, the swapping of the double buffer must ONLY be done
	// here. If something else swapped again while the reading is occuring, the
	// pointer to the histogram we keep in this method would be written to again,
	// which would be bad as the histogram might not be reset, and there would be
	// data race, and a lot of other problems. So don't do it!
	//
	// Another thing to be very careful about is the time stamps. Ideally, this
	// function controls all timestamps to ensure calculations are not smeared.
	//
	// To ensure the data swapping occur as quickly as possible, the array memory
	// allocation is done outside the actual swap loop. It probably doesn't
	// matter, as testing shows that it is about 5-10% the time it takes to swap
	// the data.
	region := trace.StartRegion(ctx, "AllocateSliceForData")
	histograms := make(map[string][]*ExtendedHdrHistogram)
	for _, workload := range d.Benchmark.workloads {
		config := workload.Config()
		histograms[config.Name] = make([]*ExtendedHdrHistogram, config.RateControl.Concurrency)
	}

	// Anonymous function declaration likely needs an allocation too (to capture
	// variables via closure), so might as well do it early.
	resetStartTime := func(h *ExtendedHdrHistogram) {
		h.ResetStartTime(now)
	}
	region.End()

	// Actually swap the data. When the data is swapped, the benchmark workers
	// will write into a new histogram with the 0 values and the start time equal
	// to the `now` variable defined here. The SwapData region should be very
	// fast. Confirm with pprof. In the current experiment, I see a worst case
	// of around 60us.
	region = trace.StartRegion(ctx, "SwapData")
	for _, workload := range d.Benchmark.workloads {
		config := workload.Config()
		workload.ForEachOnlineHistogram(func(i int, onlineHist *OnlineHistogram) {
			histograms[config.Name][i] = onlineHist.Swap(resetStartTime)
		})
	}
	region.End()

	// Read and merge the data. This takes a long time and we don't want to block data collection.
	region = trace.StartRegion(ctx, "MergeData")

	// Make sure we get the last start time. The start time should all be the same
	// (Merge will panic if it is not the same)
	var lastStartTime time.Time
	for _, hists := range histograms {
		lastStartTime = hists[0].startTime
		break
	}

	allWorkloadsMergedHistogram := NewExtendedHdrHistogram(lastStartTime)
	for workloadName, hists := range histograms {
		perWorkloadMergedHistogram := NewExtendedHdrHistogram(lastStartTime)
		for _, hist := range hists {
			perWorkloadMergedHistogram.Merge(hist)
		}

		// TODO: perhaps this is not the best way to get the LatencyHistMin and Max...
		workload := d.Benchmark.workloads[workloadName]
		config := workload.Config()

		dataSnapshot.PerWorkloadData[workloadName] = WorkloadDataSnapshot{
			IntervalData: perWorkloadMergedHistogram.IntervalData(
				now,
				config.Visualization.LatencyHistMin,
				config.Visualization.LatencyHistMax,
				config.Visualization.LatencyHistSize,
			),
			DesiredRate: config.RateControl.EventRate,
		}

		allWorkloadsMergedHistogram.Merge(perWorkloadMergedHistogram)
		dataSnapshot.AllWorkloadData.DesiredRate += config.RateControl.EventRate
	}

	dataSnapshot.AllWorkloadData.IntervalData = allWorkloadsMergedHistogram.IntervalData(now, 1, 300000, 1000) // TODO: configurable
	region.End()

	// Reset all the histograms so it can be swapped again
	region = trace.StartRegion(ctx, "ResetData")
	for _, hists := range histograms {
		for _, hist := range hists {
			hist.ResetDataOnly()
		}
	}
	region.End()

	return dataSnapshot
}

func (d *DataLogger) logData(dataSnapshot *DataSnapshot) {
	_, task := trace.NewTask(context.Background(), "LogData")
	defer task.End()

	d.dataRing.Push(dataSnapshot)

	args := dataSnapshot.AllWorkloadData.queryArgs("__all__", dataSnapshot.Time)
	_, err := d.db.Exec(fmt.Sprintf(insertQuery, d.TableName), args...)
	if err != nil {
		d.logger.WithError(err).Panic("failed to write data")
	}
	d.logger.Debug(args)

	for workloadName, workloadSnapshot := range dataSnapshot.PerWorkloadData {
		args := workloadSnapshot.queryArgs(workloadName, dataSnapshot.Time)
		_, err := d.db.Exec(fmt.Sprintf(insertQuery, d.TableName), args...)
		if err != nil {
			d.logger.WithError(err).Panic("failed to write data")
		}
		d.logger.Debug(args)
	}

}
