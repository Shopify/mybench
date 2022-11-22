import numpy as np
import matplotlib.axes
import matplotlib.pyplot as plt
from matplotlib.offsetbox import AnchoredText
import pandas

from .utils import data_to_ellipse

class Run(object):
  def __init__(self, load_driver_data: pandas.DataFrame, note: str, remove_data_from_beginning: float = 0.0):
    self.workloads = list(filter(lambda c: c != "__all__", load_driver_data["workload"].unique()))
    self.load_driver_data = load_driver_data
    self.note = note
    self.remove_data_from_beginning = remove_data_from_beginning

    self.time_domain = (0, load_driver_data["seconds_since_start"].iat[-1])

    self.plot_style = {
      "color": "k",
      "linestyle": "-",
    }

  def rate_stddev(self, workload_name="__all__"):
    all_data = self.load_driver_data_for(workload_name, remove_data_from_beginning=self.remove_data_from_beginning)
    return np.std(all_data["rate"])

  def load_driver_data_for(self, workload_name: str, remove_data_from_beginning: float = None) -> pandas.DataFrame:
    # TODO: there are nulls in the data because the uniform_hist feature is not
    # ready yet. Since we need to drop the NaN rows, and pandas treats None as
    # the same as NaN, we need to replace it so it doesn't get accidentally
    # dropped.
    data = self.load_driver_data.where(self.load_driver_data["workload"] == workload_name).replace({None: "Empty"}).dropna()
    if remove_data_from_beginning is None:
      remove_data_from_beginning = self.remove_data_from_beginning

    if remove_data_from_beginning > 0:
      data = data.where(data["seconds_since_start"] >= remove_data_from_beginning).dropna()

    return data

  def standard_figure(self, legend=False, label=True):
    fig = plt.figure(figsize=[14, 8])
    gs = fig.add_gridspec(2, 2)

    xlim = self.time_domain

    # Overall QPS
    ax = fig.add_subplot(gs[0, 0])
    self.plot_overall_qps(ax, style={"color": "k"}, label=label)
    ax.grid()
    ax.set_xlim(xlim)
    ax.set_ylabel("Overall QPS")
    ax.xaxis.set_ticklabels([]) # Remove the xlabel because it'll be the same as the plot below

    ax = fig.add_subplot(gs[0, 1])
    self.plot_overall_all_percentile_latency(ax, label=label)
    ax.grid()
    ax.set_xlim(xlim)
    ax.set_ylabel("All workload latency (ms)")
    ax.xaxis.set_ticklabels([]) # Remove the xlabel because it'll be the same as the plot below

    ax = fig.add_subplot(gs[1, 0])
    self.plot_per_workload_qps(ax, label=label)
    ax.grid()
    ax.set_xlim(xlim)
    ax.set_ylabel("Per workload QPS")
    ax.set_xlabel("Time (s)")

    ax = fig.add_subplot(gs[1, 1])
    self.plot_per_workload_percentile_99_latency(ax, label=label)
    ax.grid()
    ax.set_xlim(xlim)
    ax.set_ylabel("Per workload p99 latency (ms)")
    ax.set_xlabel("Time (s)")

    fig.tight_layout()
    return fig

  def plot_overall_qps(self, ax: matplotlib.axes.Axes, style: dict = {}, **kwargs):
    all_data = self.load_driver_data_for("__all__", remove_data_from_beginning=self.remove_data_from_beginning)
    color = style.get("color", self.plot_style["color"])
    alpha = style.get("alpha", 0.4)

    # Plot scatter plot
    ax.plot(
      all_data["seconds_since_start"], all_data["rate"],
      linestyle="none",
      marker=".",
      color=color,
      alpha=alpha,
    )

    # Plot desired rate
    ax.plot(
      all_data["seconds_since_start"], all_data["desired_rate"],
      linestyle="--",
      color=color,
      alpha=alpha,
    )

    # Plot variance ellipse
    ellipse, line, _ = data_to_ellipse(
      all_data["seconds_since_start"], all_data["rate"],
      edgecolor=color,
      facecolor="None", # string "None" gives empty/transparent
    )

    ax.add_artist(ellipse)
    # ax.add_artist(line)

    if kwargs.get("label", True):
      # Calculate and display the standard deivation
      stddev = np.std(all_data["rate"])
      mean = np.mean(all_data["rate"])
      text = AnchoredText(
        self.note + "\n" + r"$d = {:.0f}$".format(all_data["desired_rate"].iat[0]) + "\n" + r"$\bar{{x}} = {:.0f}$".format(mean) + "\n" + r"$\sigma = {:.0f}$".format(stddev),
        loc="lower center",
        prop=dict(fontsize="xx-small", ha="center"),
        frameon=False,
      )
      ax.add_artist(text)

    if kwargs.get("ylim", True):
      ax.set_ylim(0, ax.get_ylim()[1])

  def plot_per_workload_qps(self, ax: matplotlib.axes.Axes, style: dict = {}, **kwargs):
    for workload_name in self.workloads:
      data = self.load_driver_data_for(workload_name, remove_data_from_beginning=self.remove_data_from_beginning)

      # Plot scatter plot
      l, = ax.plot(
        data["seconds_since_start"], data["rate"],
        linestyle="-",
        marker=".",
        alpha=style.get("alpha", 0.4),
      )

      ax.plot(
        data["seconds_since_start"], data["desired_rate"],
        linestyle="--",
        color=l.get_color(),
      )

      if kwargs.get("label", True):
        # Can be crowded...
        ax.annotate(
          workload_name,
          xy=(data["seconds_since_start"].iat[-1], data["desired_rate"].iat[-1]),
          xytext=(-10, 15),
          color=l.get_color(),
          textcoords="offset pixels",
          va="bottom",
          ha="right",
        )

    if kwargs.get("ylim", True):
      ax.set_ylim(0, ax.get_ylim()[1])

  def plot_per_workload_percentile_99_latency(self, ax: matplotlib.axes.Axes, style: dict = {}, **kwargs):
    for workload_name in self.workloads:
      data = self.load_driver_data_for(workload_name, remove_data_from_beginning=self.remove_data_from_beginning)

      p99 = data["percentile99"] / 1000.0

      l, = ax.plot(
        data["seconds_since_start"], p99,
        linestyle="",
        marker=".",
        alpha=style.get("alpha", 0.4),
      )

      ellipse, line, text = data_to_ellipse(
        data["seconds_since_start"], p99,
        text=workload_name,
        edgecolor=l.get_color(),
        facecolor="None", # string "None" gives empty/transparent
      )

      ax.add_artist(ellipse)
      # ax.add_artist(line)

      if kwargs.get("label", True):
        pass
        # ax.add_artist(text)

    if kwargs.get("ylim", True):
      ax.set_ylim(0, ax.get_ylim()[1])

  def plot_overall_percentile_99(self, ax: matplotlib.axes.Axes, style: dict = {}, **kwargs):
      data = self.load_driver_data_for("__all__", remove_data_from_beginning=self.remove_data_from_beginning)
      p99 = data["percentile99"] / 1000.0

      l, = ax.plot(
        data["seconds_since_start"], p99,
        linestyle="",
        color=style.get("color", None),
        marker=style.get("marker", "."),
        alpha=style.get("alpha", 0.4),
      )

      ellipse, line, text = data_to_ellipse(
        data["seconds_since_start"], p99,
        text=kwargs.get("label", ""),
        edgecolor=l.get_color(),
        facecolor="None", # string "None" gives empty/transparent
      )

      ax.add_artist(ellipse)

      if kwargs.get("label", True):
        ax.add_artist(text)

      if kwargs.get("ylim", True):
        ax.set_ylim(0, ax.get_ylim()[1])

  def plot_overall_all_percentile_latency(self, ax: matplotlib.axes.Axes, style: dict = {}, **kwargs):
    data = self.load_driver_data_for("__all__", remove_data_from_beginning=self.remove_data_from_beginning)

    for key in ["percentile50", "percentile75", "percentile90", "percentile99"]:
      label = f"p{key.split('percentile')[1]}"

      percentile_data = data[key] / 1000.0

      l, = ax.plot(
        data["seconds_since_start"], percentile_data,
        linestyle="-",
        label=label,
      )

      if kwargs.get("label", True):
        ax.annotate(
          label,
          xy=(data["seconds_since_start"].iat[-1], percentile_data.iat[-1]),
          xytext=(5, 0),
          color=l.get_color(),
          textcoords="offset pixels",
          va="center",
        )

    if kwargs.get("ylim", True):
      ax.set_ylim(0, ax.get_ylim()[1])
