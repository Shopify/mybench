from collections.abc import Iterable
from datetime import datetime
import sqlite3
import pandas
import logging
import matplotlib
import matplotlib.gridspec
import matplotlib.axes
import matplotlib.pyplot as plt

from .run import Run


class Database(object):
  def __init__(self, filename: str, benchmark_name: str|None = None):
    self.filename = filename
    self.conn = sqlite3.connect(self.filename)
    self.conn.row_factory = sqlite3.Row
    self.runs_meta = list(self.fetch_runs_meta(benchmark_name))

  def fetch_runs_meta(self, benchmark_name: str|None) -> Iterable[dict]:
    cur = self.conn.cursor()
    query = "SELECT * FROM meta"
    args = []
    if benchmark_name is not None:
      query += " WHERE benchmark_name = ?"
      args.append(benchmark_name)

    for row in cur.execute(query, args):
      yield dict(row)

    cur.close()

  def run_data(self, run_table_name: str, note: str, remove_data_from_beginning: float = 0.0):
    query = f"SELECT * FROM {run_table_name}"

    logging.debug(f"reading data via {query}")
    load_driver_data = pandas.read_sql(query, self.conn)

    # TODO: read prometheus data if applicable

    return Run(load_driver_data, note, remove_data_from_beginning=remove_data_from_beginning)

  def __del__(self):
    self.conn.close()

  def plot_overall_qps_comparison(self, gs: matplotlib.gridspec.GridSpec,
                                        xlim: tuple[float, float],
                                        ylim: tuple[float, float]|None = None,
                                        remove_data_from_beginning: float = 0.0) -> list[matplotlib.axes.Axes]:
    runs = self.time_sorted_runs(remove_data_from_beginning)

    subgs = gs.subgridspec(nrows=1, ncols=len(runs), wspace=0)
    axs = subgs.subplots(sharey=True, squeeze=False)[0]

    for i, ax in enumerate(axs):
      runs[i].plot_overall_qps(ax, ylim=False)
      ax.set_xlim(xlim)
      ax.set_xticks([])

    if ylim is None:
      ylim = (0, axs[0].get_ylim()[1])

    axs[0].set_ylim(ylim)
    axs[0].set_ylabel("Overall QPS")

    return axs

  def plot_overall_latency_percentile_comparison(self, gs: matplotlib.gridspec.GridSpec,
                                                       xlim: tuple[float, float],
                                                       ylim: tuple[float, float]|None = None,
                                                       remove_data_from_beginning: float = 0.0) -> list[matplotlib.axes.Axes]:
    runs = self.time_sorted_runs(remove_data_from_beginning)

    subgs = gs.subgridspec(nrows=1, ncols=len(runs), wspace=0)
    axs = subgs.subplots(sharey=True, squeeze=False)[0]

    for i, ax in enumerate(axs):
      runs[i].plot_overall_all_percentile_latency(ax, ylim=False)
      ax.set_xlim(xlim)
      ax.set_xticks([])

    if ylim is None:
      ylim = (0, axs[0].get_ylim()[1])

    axs[0].set_ylim(ylim)
    axs[0].set_ylabel("Overall latency percentile (ms)")

    return axs

  def standard_figure(self, qps_ylim: tuple[float, float]|None = None,
                            latency_ylim: tuple[float, float]|None = None,
                            title: str|None = None,
                            remove_data_from_beginning: float = 0.0):
    import matplotlib.pyplot as plt

    fig = plt.figure()
    gs = fig.add_gridspec(ncols=1, nrows=2)

    min = float("inf")
    max = float("-inf")
    for run in self.time_sorted_runs(remove_data_from_beginning):
      t = run.load_driver_data_for("__all__")["seconds_since_start"]
      if t.max() > max:
        max = t.max()

      if t.min() < min:
        min = t.min()

    xlim = [min, max]

    self.plot_overall_qps_comparison(gs[0, 0], xlim=xlim, ylim=qps_ylim, remove_data_from_beginning=remove_data_from_beginning)
    self.plot_overall_latency_percentile_comparison(gs[1, 0], xlim=xlim, ylim=latency_ylim, remove_data_from_beginning=remove_data_from_beginning)

    if title:
        fig.suptitle(title)
    fig.tight_layout()
    return fig

  def time_sorted_runs(self, remove_data_from_beginning: float = 0.0) -> list[Run]:
    runs = []
    for meta in sorted(self.runs_meta, key=lambda m: datetime.strptime(m["start_time"], "%Y-%m-%dT%H:%M:%SZ")):
      runs.append(self.run_data(meta["table_name"], meta["note"], remove_data_from_beginning))

    return runs
