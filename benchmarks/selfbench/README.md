`selfbench`
===========

This benchmark tests the mybench performance with respect to different
concurrency control parameters. To do this, instead of query MySQL in the
`Event` function of the `BenchmarkClass`, we call a function that have a known
run time. By doing this, we know what the latency should be to a high-degree of
certainty. By comparing the resulting distribution logged by mybench with this
known quantity, we can reveal additional latencies caused by factors outside of
mybench code that may cause latency, such as in the operating system, Golang
runtime, and hardware resource constraints.

Function with known runtime
---------------------------

We need to create a function with a known run-time. `Sleep` does not work
because it usually result in a system call (or a call to the Golang runtime)
that pauses the thread. The wakeup of the thread is not guaranteed at the end of
the sleep duration. This wakeup latency is not what we want to measure. Instead,
we need a function that burns the CPU for a fixed amount of time, ideally with
durations of a few hundred microseconds for this benchmark to work. To do this,
we can define the function `sumUntil`:

```math
\mathrm{sumUntil}(v) = \sum_{i = 0}^{v} i
```

This function has a runtime of $av + b$, where $a$ and $b$ are unknown
constants. By measuring the excution duration with different values of $v$
during the benchmark run (on init), we can solve for $a$ and $b$ with linear
regression. Note that $b$ should be very close to 0, but we solve for it anyway
to ensure better fit with the linear regression algorithm.
