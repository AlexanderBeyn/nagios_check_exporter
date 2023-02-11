# nagios_check_exporter

A Nagios check plugin runner and exporter for Prometheus.

## Installing

```
$ go install github.com/AlexanderBeyn/nagios_check_exporter@latest
```

## Configuration

`nagios_check_exporter` requires a configuration file that describes Nagios
checks to run. An example file can be shown with the `-config.example` parameter:

```
$ nagios_check_exporter -config.example
# example configuration file
defaults:
  check_interval: 5m
  timeout: 60s
  labels:
    common_label: some_value

commands:
  - name: http
    command: [/usr/lib/nagios/plugins/check_http, -H, www.example.com]
    timeout: 10s
    labels:
      host: www.example.com
```

| Setting        | Description                                                                                                                                                     |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name           | This value will be used in the final metric name, and in log entries related to the command.                                                                    |
| command        | A list of strings corresponding to the check plugin and all arguments. This will be executed directly, and not passed through `sh -c`                           |
| check_interval | How often to run the check. Valid duration strings are described in Go [time](https://pkg.go.dev/time#ParseDuration) package.                                   |
| timeout        | How long to wait for the check to complete before killing it. Valid duration strings are described in Go [time](https://pkg.go.dev/time#ParseDuration) package. |
| labels         | A mapping of labels to add to all metrics associated with this command.                                                                                         |

The top-level `defaults` section will be applied to all commands, along with the following fall back defaults:

| Setting        | Default |
| -------------- | ------- |
| check_interval | 5m      |
| timeout        | 30s     |

## Running

```
$ nagios_check_exporter -h
Usage of nagios_check_exporter:
  -config.example
        Show example configuration file
  -config.file string
        Configuration file describing Nagios checks (default "config.yaml")
  -listen.address string
        Listening address for metrics (default ":28272")
  -log.level int
        Log level: 0 (ERROR), 1 (INFO), 2 (DEBUG) (default 1)
  -metrics.path string
        Path under which to expose metrics (default "/metrics")
```

## Metrics

`nagios_check_exporter` captures the check plugin's status, along with performance data.

In the table below, _[name]_ is the name defined in the command configuration, and _[perf]_ is the performance data label as presented by the check plugin. If any of the performance data metrics are not returned by the plugin or cannot be parsed, they will not be updated.

| Metric                                  | Description                                                           |
| --------------------------------------- | --------------------------------------------------------------------- |
| nagios_check\__[name]_\_check_status    | Check status. 0 for OK, 1 for WARNING, 2 for CRITICAL, 3 for UNKNOWN. |
| nagios_check\__[name]_\_check_duration  | How long the check plugin took to finish.                             |
| nagios_check\__[name]_\_check_run_time  | The last time this check plugin was executed, in Unix time.           |
| nagios_check\__[name]_\__[perf]_\_value | Current value associated with the performance data.                   |
| nagios_check\__[name]_\__[perf]_\_min   | Minimum value for performance data, as reported by the check plugin.  |
| nagios_check\__[name]_\__[perf]_\_max   | Maximum value for performance data, as reported by the check plugin.  |
