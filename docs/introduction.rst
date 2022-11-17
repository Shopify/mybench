.. _introduction:

================
What is mybench?
================

mybench is a high-performance, massively-parallel library primarily for
building database benchmarks in Go. By default, mybench supports MySQL. While
not done, it should be possible to use it for other databases (or even
non-databases).

The main features of mybench includes:

* **Precision event rate control**: the rate at which events are run is accurately
  and precisely maintained by a discretized-time rate controller.
* **Workload parallelization**: a single workload can be parallelized into multiple
  goroutines, each with its own connection and rate controller.
* **Workload mixing**: multiple workloads can be mixed with arbitrary ratios
  and run as a single benchmark. Data are collected from each workload
  individually can be visualized either in aggregate or individually.
* **Real-time monitoring UI**: mybench exposes a web interface that plots the
  throughput and latency time series while the benchmark is running in
  real-time. This enables faster feedback loop for writing and validating
  benchmarks.
* **SQLite-based data logging**: mybench writes the throughput and latency time
  series into a SQLite database. Multiple benchmark runs can share the same
  database file. The standard and single-file approach to data logging allows
  for better interoperability with other tools as well as for easier data
  sharing.
* **Post-processing scripts**: mybench includes a set of scripts that creates
  standard visualizations with the data logged in the SQLite database file.
  Multiple benchmark runs can be compared with these scripts.
* **Data generators with data sampling**: mybench includes a
  number of built-in random number generators that also have the capability to
  sample from existing data so that the ``WHERE`` clause of ``SELECT``
  statements can be more easily constructed. However, this sampling is not
  perfect for the moment and can result in non-existing data being sampled.

----------------------------
Comparisons with other tools
----------------------------

* `sysbench <https://github.com/akopytov/sysbench>`_ is a similar database and
  system benchmark tool that allows custom benchmarks to be developed with Lua.
  It has a long history in the database benchmark community and is one of the
  de facto standards dating all the way back to 2004.

  * mybench has significantly more accurate event rate control than sysbench
    based on our internal testing.
  * mybench collects more data than sysbench during the benchmark run. In
    addition to being able to visualize these data with plots in real-time,
    mybench also includes more tools to analyze this data after the benchmark
    run is complete.
  * It is easier to define a multi-workload benchmark in mybench than in
    sysbench, as workloads have to be mixed manually with Lua in sysbench.
    Further, mybench also collects workload-specific data, which helps with
    post-benchmark analysis.

* `BenchBase <https://github.com/cmu-db/benchbase>`_ is a multi-DBMS benchmark
  framework written in Java. Using BenchBase, benchmarks can be created for
  multiple databases and comparisons can be made across databases.

  * Both mybench and BenchBase has very accurate rate control, although they
    are implemented differently.
  * Data gathered by mybench can be monitored in real-time via the built-in
    HTTP server whereas data gathered from BenchBase can only be visualized
    during post processing as of this writing.
  * BenchBase is more focused around databases (through heavy use of JDBC)
    while mybench can in principle work with non-database workloads.
  * BenchBase can evolve the ratio of workload mixtures over the duration of
    the benchmark, while mybench cannot do this natively.
