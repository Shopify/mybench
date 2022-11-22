const API_STATUS_URL = "/api/status";
const VL_SCHEMA = "https://vega.github.io/schema/vega-lite/v5.json";

// TODO:
// - Consistent colors between histogram and the line graphs

async function get_status() {
  const resp = await fetch(API_STATUS_URL);
  if (!resp.ok) {
    const msg = `failed to get status: ${resp.status} ${resp.statusText}`;
    console.log(msg);
    throw new Error(msg);
  }

  const data = await resp.json();

  // Sort the data so they are always consistent in the UI.
  data.Workloads.sort();
  for (let data_snapshot of data.DataSnapshots) {
    data_snapshot.SortedData = Object.entries(data_snapshot.PerWorkloadData).sort((a, b) => a[0] - b[0]);
  }

  return data;
}

function draw_overall_rate_plot(status_data, time_domain) {
  let vl_data = [];

  for (const data_snapshot of status_data.DataSnapshots) {
    vl_data.push({
      "Time": data_snapshot.Time,
      "Event rate": data_snapshot.AllWorkloadData.Rate,
      "Desired rate": data_snapshot.AllWorkloadData.DesiredRate,
    });
  }

  window.overall_rate_vega_view
    .signal("time_domain", time_domain)
    .change("data", vega.changeset().insert(vl_data).remove(vega.truthy))
    .resize()
    .run();
}

function draw_overall_latency_percentile(status_data, time_domain) {
  let vl_data = [];

  for (const data_snapshot of status_data.DataSnapshots) {
    vl_data.push({
      "Time": data_snapshot.Time,
      "Latency (ms)": data_snapshot.AllWorkloadData.Percentile50 / 1000,
      "Percentile": 50,
    });

    vl_data.push({
      "Time": data_snapshot.Time,
      "Latency (ms)": data_snapshot.AllWorkloadData.Percentile75 / 1000,
      "Percentile": 75,
    });

    vl_data.push({
      "Time": data_snapshot.Time,
      "Latency (ms)": data_snapshot.AllWorkloadData.Percentile90 / 1000,
      "Percentile": 90,
    });

    vl_data.push({
      "Time": data_snapshot.Time,
      "Latency (ms)": data_snapshot.AllWorkloadData.Percentile99 / 1000,
      "Percentile": 99,
    });
  }

  window.latency_percentile_vega_view
    .signal("time_domain", time_domain)
    .change("data", vega.changeset().insert(vl_data).remove(vega.truthy))
    .resize()
    .run();
}

function draw_event_rate_plots(status_data, time_domain) {
  let vl_data = [];

  for (const data_snapshot of status_data.DataSnapshots) {
    for (const [workload_name, workload_snapshot] of data_snapshot.SortedData) {
      vl_data.push({
        "Time": data_snapshot.Time,
        "Workload": workload_name,
        "Event rate": workload_snapshot.Rate,
        "Desired rate": workload_snapshot.DesiredRate,
        "Event rate / desired rate": workload_snapshot.Rate / workload_snapshot.DesiredRate,
      });
    }
  }

  window.event_rate_vega_view
    .signal("time_domain", time_domain)
    .change("data", vega.changeset().insert(vl_data).remove(vega.truthy))
    .resize()
    .run();

  window.event_rate_pct_vega_view
    .signal("time_domain", time_domain)
    .change("data", vega.changeset().insert(vl_data).remove(vega.truthy))
    .resize()
    .run();
}

function draw_latency_plots(status_data, time_domain) {
  let vl_data = [];

  for (const data_snapshots of status_data.DataSnapshots) {
    for (const [workload_name, workload_snapshot] of data_snapshots.SortedData) {
      vl_data.push({
        "Time": data_snapshots.Time,
        "Workload": workload_name,
        "Mean latency (ms)": workload_snapshot.Mean / 1000.0,
        "Max latency (ms)": workload_snapshot.Max / 1000.0,
        "25th Percentile": workload_snapshot.Percentile25 / 1000.0,
        "50th Percentile": workload_snapshot.Percentile50 / 1000.0,
        "75th Percentile": workload_snapshot.Percentile75 / 1000.0,
        "90th Percentile": workload_snapshot.Percentile90 / 1000.0,
        "99th Percentile": workload_snapshot.Percentile99 / 1000.0,
      });
    }
  }

  window.mean_latency_vega_view
    .signal("time_domain", time_domain)
    .change("data", vega.changeset().insert(vl_data).remove(vega.truthy))
    .resize()
    .run();

  window.max_latency_vega_view
    .signal("time_domain", time_domain)
    .change("data", vega.changeset().insert(vl_data).remove(vega.truthy))
    .resize()
    .run();
}

function draw_latency_histogram(status_data, workload_name) {
  if (status_data.DataSnapshots.length == 0) {
    return;
  }

  // Assume they are all the same
  let first_hist_buckets = status_data.DataSnapshots[0].PerWorkloadData[workload_name].UniformHist.Buckets;
  let num_hist_buckets = first_hist_buckets.length;

  let summed_hist_values = [];
  for (let i = 0; i < num_hist_buckets; i++) {
    summed_hist_values.push(0);
  }

  for (const data_snapshot of status_data.DataSnapshots) {
    for (const [i, bar] of data_snapshot.PerWorkloadData[workload_name].UniformHist.Buckets.entries()) {
      summed_hist_values[i] += bar.Count;
    }
  }

  let vl_data = [];
  for (const [i, bar] of first_hist_buckets.entries()) {
    vl_data.push({
      "Latency (ms)": (bar.From + bar.To) / 2 / 1000,
      "Count": summed_hist_values[i],
    });
  }

  window.hist_vega_views[workload_name]
    .signal("latency_min", first_hist_buckets[0].From / 1000)
    .signal("latency_max", first_hist_buckets[first_hist_buckets.length - 1].From / 1000)
    .change("data", vega.changeset().insert(vl_data).remove(vega.truthy))
    .resize()
    .run();
}

function update_plots(status_data) {
  let min_time = 0.0;
  if (status_data.DataSnapshots.length > 0) {
    min_time = status_data.DataSnapshots[0].Time;
  }

  const time_domain = [min_time, status_data.CurrentTime];

  draw_overall_rate_plot(status_data, time_domain);
  draw_overall_latency_percentile(status_data, time_domain);
  draw_event_rate_plots(status_data, time_domain);
  draw_latency_plots(status_data, time_domain);

  for (const workload_name of status_data.Workloads) {
    draw_latency_histogram(status_data, workload_name);
  }
}

async function setup_plots(workloads) {
  // Overall event rate plot
  let overall_rate_vl_schema = {
    $schema: VL_SCHEMA,
    width: "container",
    title: "Overall rate",
    data: { name: "data" },
    layer: [
      {
        mark: {
          type: "line",
          point: true,
        },
        encoding: {
          "x": {
            field: "Time",
            type: "quantitative",
            scale: {
              domain: { signal: "time_domain" },
              nice: false
            },
          },
          "y": { field: "Event rate", type: "quantitative" },
        },
      },
      {
        mark: {
          type: "line",
          strokeWidth: 1,
          strokeDash: [4, 4],
        },
        encoding: {
          "x": {
            field: "Time",
            type: "quantitative",
          },
          "y": { field: "Desired rate", type: "quantitative" },
        },
      },
    ],
  };

  const overall_rate_vega_promise = vegaEmbed("#overall-rate-vis", overall_rate_vl_schema, {
    // Workaround because vega lite can't have signals
    // https://stackoverflow.com/questions/57707494/whats-the-proper-way-to-implement-a-custom-click-handler-in-vega-lite
    patch: (spec) => {
      spec.signals.push({ name: "time_domain" });
      return spec;
    }
  });

  // Percentile plot
  let latency_percentile_vl_schema = {
    $schema: VL_SCHEMA,
    width: "container",
    title: "Overall latency percentile",
    data: { name: "data" },
    signals: [
      { name: "time_domain", value: [0, 1] },
    ],
    layer: [
      {
        mark: {
          type: "line",
          point: true,
        },
        encoding: {
          "x": {
            field: "Time",
            type: "quantitative",
            scale: {
              domain: { signal: "time_domain" },
              nice: false,
            },
          },
          "y": { field: "Latency (ms)", type: "quantitative" },
          "color": {
            field: "Percentile",
            type: "nominal",
            legend: {
              orient: "bottom",
            },
          },
        },
      },
    ]
  };

  const latency_percentile_vega_promise = vegaEmbed("#overall-latency-percentile-vis", latency_percentile_vl_schema, {
    // Workaround because vega lite can't have signals
    // https://stackoverflow.com/questions/57707494/whats-the-proper-way-to-implement-a-custom-click-handler-in-vega-lite
    patch: (spec) => {
      spec.signals.push({ name: "time_domain" });
      return spec;
    }
  });

  // Setup event rate plot
  let event_rate_vl_schema = {
    $schema: VL_SCHEMA,
    width: "container",
    title: "Event rate",
    data: { name: "data" },
    layer: [
      {
        mark: {
          type: "line",
          point: true,
        },
        encoding: {
          "x": {
            field: "Time",
            type: "quantitative",
            scale: {
              domain: { signal: "time_domain" },
              nice: false
            },
          },
          "y": { field: "Event rate", type: "quantitative" },
          "color": {
            field: "Workload",
            type: "nominal",
            legend: {
              orient: "bottom",
            },
          },
        },
      },
      {
        mark: {
          type: "line",
          strokeWidth: 1,
          strokeDash: [4, 4],
        },
        encoding: {
          "x": {
            field: "Time",
            type: "quantitative",
          },
          "y": { field: "Desired rate", type: "quantitative" },
          "color": {
            field: "Workload",
            type: "nominal",
          },
        },
      },
    ],
  };

  const event_rate_vega_promise = vegaEmbed("#event-rate-vis", event_rate_vl_schema, {
    // Workaround because vega lite can't have signals
    // https://stackoverflow.com/questions/57707494/whats-the-proper-way-to-implement-a-custom-click-handler-in-vega-lite
    patch: (spec) => {
      spec.signals.push({ name: "time_domain" });
      return spec;
    }
  });

  let event_rate_pct_vl_spec = { // TODO: merge this into a subplot with the event_rate above to share the X axis
    $schema: VL_SCHEMA,
    width: "container",
    title: "Event rate / desired rate",
    data: { name: "data" },
    layer: [
      {
        mark: {
          type: "line",
          point: true,
        },
        encoding: {
          "x": {
            field: "Time",
            type: "quantitative",
            scale: {
              domain: { signal: "time_domain" },
              nice: false
            },
          },
          "y": { field: "Event rate / desired rate", type: "quantitative" },
          "color": {
            field: "Workload",
            type: "nominal",
            legend: {
              orient: "bottom",
            },
          },
        },
      },
    ],
  };

  const event_rate_pct_vega_promise = vegaEmbed("#event-rate-pct-vis", event_rate_pct_vl_spec, {
    // Workaround because vega lite can't have signals
    // https://stackoverflow.com/questions/57707494/whats-the-proper-way-to-implement-a-custom-click-handler-in-vega-lite
    patch: (spec) => {
      spec.signals.push({ name: "time_domain" });
      return spec;
    }
  });

  let mean_latency_vl_schema = {
    $schema: VL_SCHEMA,
    width: "container",
    title: "Mean event latency",
    data: { name: "data" },
    signals: [
      { name: "time_domain", value: [0, 1] },
    ],
    layer: [
      {
        mark: {
          type: "line",
          point: true,
        },
        encoding: {
          "x": {
            field: "Time",
            type: "quantitative",
            scale: {
              domain: { signal: "time_domain" },
              nice: false,
            },
          },
          "y": { field: "Mean latency (ms)", type: "quantitative" },
          "color": {
            field: "Workload",
            type: "nominal",
            legend: {
              orient: "bottom",
            },
          },
        },
      },
    ]
  };

  const mean_latency_vega_promise = vegaEmbed("#mean-latency-vis", mean_latency_vl_schema, {
    // Workaround because vega lite can't have signals
    // https://stackoverflow.com/questions/57707494/whats-the-proper-way-to-implement-a-custom-click-handler-in-vega-lite
    patch: (spec) => {
      spec.signals.push({ name: "time_domain" });
      return spec;
    }
  });

  let max_latency_vl_schema = {
    $schema: VL_SCHEMA,
    width: "container",
    title: "Max event latency",
    data: { name: "data" },
    signals: [
      { name: "time_domain", value: [0, 1] },
    ],
    layer: [
      {
        mark: {
          type: "line",
          point: true,
        },
        encoding: {
          "x": {
            field: "Time",
            type: "quantitative",
            scale: {
              domain: { signal: "time_domain" },
              nice: false,
            },
          },
          "y": { field: "Max latency (ms)", type: "quantitative" },
          "color": {
            field: "Workload",
            type: "nominal",
            legend: {
              orient: "bottom",
            },
          },
        },
      },
    ]
  };

  const max_latency_vega_promise = vegaEmbed("#max-latency-vis", max_latency_vl_schema, {
    // Workaround because vega lite can't have signals
    // https://stackoverflow.com/questions/57707494/whats-the-proper-way-to-implement-a-custom-click-handler-in-vega-lite
    patch: (spec) => {
      spec.signals.push({ name: "time_domain" });
      return spec;
    }
  });

  let hist_vega_promises = {};

  for (const workload_name of workloads) {
    let id = `hist-${workload_name}`;
    let histograms_div = document.getElementById("histograms");
    let vis_div = document.createElement("div");
    vis_div.id = id;
    vis_div.className = "plot";
    histograms_div.appendChild(vis_div);

    let hist_vl_spec = {
      $schema: VL_SCHEMA,
      width: "container",
      title: `${workload_name} latency histogram`,
      data: { name: "data" },
      layer: [
        {
          mark: { type: "bar" },
          encoding: {
            "x": {
              field: "Latency (ms)",
              type: "quantitative",
              scale: {
                "domainMin": { signal: "latency_min" },
                "domainMax": { signal: "latency_max" },
              },
            },
            "y": { field: "Count", type: "quantitative" },
          },
        },
      ]
    };

    hist_vega_promises[workload_name] = vegaEmbed(vis_div, hist_vl_spec, {
      // Workaround because vega lite can't have signals
      // https://stackoverflow.com/questions/57707494/whats-the-proper-way-to-implement-a-custom-click-handler-in-vega-lite
      patch: (spec) => {
        spec.signals.push({ name: "latency_min" });
        spec.signals.push({ name: "latency_max" });
        return spec;
      }
    });
  }

  let overall_rate_vega_result = await overall_rate_vega_promise;
  window.overall_rate_vega_view = overall_rate_vega_result.view;

  let latency_percentile_vega_result = await latency_percentile_vega_promise;
  window.latency_percentile_vega_view = latency_percentile_vega_result.view;

  let event_rate_vega_result = await event_rate_vega_promise;
  window.event_rate_vega_view = event_rate_vega_result.view;

  let event_rate_pct_vega_result = await event_rate_pct_vega_promise;
  window.event_rate_pct_vega_view = event_rate_pct_vega_result.view;

  let mean_latency_vega_result = await mean_latency_vega_promise;
  window.mean_latency_vega_view = mean_latency_vega_result.view;

  let max_latency_vega_result = await max_latency_vega_promise;
  window.max_latency_vega_view = max_latency_vega_result.view;

  window.hist_vega_views = {};
  for (const workload_name of workloads) {
    let result = await hist_vega_promises[workload_name];
    window.hist_vega_views[workload_name] = result.view;
  }
}

async function main() {
  const status_data = await get_status();
  await setup_plots(status_data.Workloads);
  update_plots(status_data);

  setInterval(async function () {
    const status_data = await get_status();
    document.getElementById("runnote").innerHTML = "";
    if (status_data.Note.length > 0) {
        document.getElementById("runnote").innerHTML = "(" + status_data.Note + ")";
    }
    update_plots(status_data);
  }, 5000);
}

if (document.readyState != "loading") {
  main();
} else {
  document.addEventListener("DOMContentLoaded", main);
}
