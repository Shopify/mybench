`mybench`
=========

`mybench` is a benchmark authoring library that helps you create your own database benchmark with Golang. The central features of mybench includes:

- A library approach to database benchmarking
- Discretized precise rate control: the rate at which the events run is discretized to a relatively low frequency (default: 50hz), as Linux + Golang cannot reliably maintain 100~1000Hz. The number of events run on each iteration is determined by sampling an uniform or Poisson distribution. The rate control is very precise and have been achieved standard deviations of <0.2% of the desired rate.
- Ability to parallelize a single workload into multiple goroutines, each with its own connection.
- Ability to run multiple workloads simultaneously with data being logged from all workloads.
- Uses [HDR Histogram](https://github.com/HdrHistogram/hdrhistogram-go) to keep track of latency online.
- Web UI for live monitoring throughput and latency of the current benchmark.
- A simple interface for implementing the data loader (which creates the tables and seed it with data) and the benchmark driver.
- A number of built-in data generators, including thread-safe auto incrementing generators.
- Command line wrapper: A wrapper library to help build command line apps for the benchmark.

Design
------

For more details, see [the design doc](./docs/00-mybench-design.rst). Some of the information in this section may eventually move there.

There are a few important structs defined in this library, and they are:

- `Benchmark`: The main "entrypoint" to running a benchmark. This keeps track of multiple `Workload`s and performs data aggregation across all the `Workload`s and their `BenchmarkWorker`s.
- `WorkloadInterface`: an interface that is defined by the end-user who want to create a benchmark. Notably, the end-user will implement an `Event()` function that should be called at a some specified `EventRate` (concurrently with a number of goroutines).
- `Workload`: Responsible for creating and running the workers (goroutines) to call the `Event()` function of the `WorkloadInterface`.
- `BenchmarkWorker`: Responsible for setting up the `Looper` and keeping track of the worker-local statistics (such as the event latency/histograms for the local goroutine).
- `Looper`: Responsible for discretizing the desired event rate into something that's achievable on Linux. Actually calls the `Event()` function. It can also perform complex discretization such as Poisson-distribution based event sampling.
- `BenchmarkDataLoader`: A data loader helper that helps you easily concurrently load data by specifying only a few options, such as the number of rows and the type of data generator for each columns.
- `BenchmarkApp[T]`: A wrapper to help create a command line app for a benchmark.
- `Table`: An object that helps you create the database and track a default set of data generators.

### Data collection and flow

The benchmark system mainly collects data about the throughput and latency of the `Event()` function call, which contains custom logic (usually MySQL calls). Since `Event()` can be called from a large number of `BenchmarkWorker`s, each `BenchmarkWorker` collects its own statistics for performance reasons. The data collected by the `BenchmarkWorker`s are:

- The count and rate of `Event()`
- The latency distribution of `Event()` as tracked via the [HDR Histogram](https://github.com/HdrHistogram/hdrhistogram-go).
- Unimplemented:
  - How long the worker spent in "saturation" (i.e. `Event()` is slower than the requested event rate). This is probably an important metric for later.
  - The amount of time spent sleeping (could be useful to debug saturation problem in case the looper is incorrectly implemented).
  - Everything in `OuterLoopStat`: wakeup latency, event batch size. This is probably less important than the above.

Having all this data in hundreds of independent Goroutines (`BenchmarkWorkers`) is not particularly useful. The data must be aggregated. This data aggregation is done on the workload level by the `Workload`, which is then aggregated at the `Benchmark` level via the data logger. This description may make it sound like the data collection is initiated by the `BenchmarkWorker`s -- it is not. Instead, every few seconds, the data logger calls the appropriate functions to aggregate data. During data collection, a lock taken for each `BenchmarkWorker`, which allows for the safe reading of data. This is fine as each `BenchmarkWorker` has its own mutex and there's never a lot of contention. If this becomes a problem, lockless programming may be a better approach.

Run a benchmark
---------------

- Shopify orders benchmark: `make examplebench && build/examplebench -host mysql-1 -user sys.admin_rw -pass hunter2 -bench -eventrate 3000`
  - Change the host
  - Change the event rate. The command above specifies 3000 events/s.
- Go to https://localhost:8005 to see the monitoring web UI.

Write your own benchmark
------------------------

See [benchmarks](./benchmarks) for examples and read the [docs](./docs).
