.. _eventrate-and-concurrency-control:

==================================
Event rate and concurrency control
==================================

To achieve high total event rate, mybench utilizes a large number of concurrent
benchmark workers (via goroutines). To achieve thousands of events per second
per benchmark worker, the events are executed in batches in a relatively-slow
"outer" loop (See rate control section in :ref:`detailed-design-doc` for more
details). The number of benchmark workers and the event execution loops can be
controlled via the following parameters:

* ``-eventrate``: specifies the total event rate of the benchmark in Hz.
* ``-concurrency``: specifies the number of benchmark workers to use.
* ``-workermaxrate``: specifies the max event rate of a worker so that
  ``-concurrency`` can be automatically calculated if not specified.
* ``-outerlooprate``: specifies the rate at which the "outer" loop runs in Hz.


If none of these arguments are specified, the defaults are: ``-eventrate`` is
1000; ``-concurrency`` is automatically determined based on ``-workermaxrate``,
which by default is 100; and ``-outerlooprate`` is 50.

This document will go through a few common examples of how to tune these
parameters.

------------------------------
Specifying only ``-eventrate``
------------------------------

In most cases, only ``-eventrate`` is needed. The number of benchmark workers
is automatically determined based on the value of ``-workermaxrate``, which is
by default 100, with the formula ``concurrency = ceiling(eventrate /
workermaxrate)``. Since the total event rate is spread between multiple
workloads according to the specified ``WorkloadConfig.WorkloadScale``, the
calculated concurrency number may be greater than the value above if there are
multiple workloads, as all workloads must have at least 1 benchmark worker, and
the each workload will perform its own round-up operation when determining its
own concurrency.

For illustrative purposes, for a single-workload benchmark, here are some
examples of:

+----------------+--------------------+------------------------------------+
| ``-eventrate`` | ``-workermaxrate`` | ``-concurrency`` (auto calculated) |
+================+====================+====================================+
| 10000          | default (100)      | 100                                |
+----------------+--------------------+------------------------------------+
| 35000          | default (100)      | 350                                |
+----------------+--------------------+------------------------------------+
| 35001          | 500                | 71                                 |
+----------------+--------------------+------------------------------------+

Adjustment to ``-workermaxrate`` may be important if the default value of 100
event/s/worker is not sufficient to fully utilize the target database.

----------------------------------------------
Specifying ``-eventrate`` and ``-concurrency``
----------------------------------------------

In some situations, it may be desirable to fix the number of benchmark workers
to use. For example, one may wish to fix the number of concurrent connections
used against the database. In these cases, ``-eventrate`` and ``-concurrency``
can be both specified. If ``-concurrency`` is specified, the value of
``-workermaxrate`` is ignored.

-------------------------------------
Advanced: changing ``-outerlooprate``
-------------------------------------

Mybench batches events in a slow-running outer loop to maximize throughput. By
default, the ``-outerlooprate`` is set to 50. If a benchmark worker's event
rate is set to 100, then 100 / 50 = 2 ``Event()`` calls are made per outer-loop
iteration. If ``-outerlooprate`` is changed to 25, then 4 ``Event()`` calls are
made per outer-loop iteration.

Tuning this variable may be useful if the behavior of the looper is erratic
independent of the database, although it is not guaranteed that this is the
root of the problem. Generally this value should not exceed 100-200, as Linux
and Golang's scheduler is unlikely to be able to maintain a constant 100-200
Hz. The per-benchmark-worker event rate (calculated from the total
``-eventrate`` by mybench) should be divisible by the ``-outerlooprate``. If it
is not divisible, the looper may not run steadily and may experience
oscillations around the desired rates due to discretization errors.
