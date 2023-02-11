package main

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var fixNameRe = regexp.MustCompile("[^a-zA-Z0-9_:]")

func fixName(s string) string {
	return fixNameRe.ReplaceAllString(s, "_")
}

func formatLabels(labels prometheus.Labels) string {
	list := make([]string, 0, len(labels))

	for k, v := range labels {
		list = append(list, k+"="+strconv.Quote(v))
	}

	return strings.Join(list, ",")
}

var metrics = sync.Map{}

func setMetric(name string, labels prometheus.Labels, value float64) {
	if !math.IsNaN(value) {
		name = fixName(name)

		var metric prometheus.Gauge

		m, ok := metrics.Load(name)
		if ok {
			metric, _ = m.(prometheus.Gauge)
		} else {
			logger.DEBUG.Printf("creating new metric %s{%s}\n", name, formatLabels(labels))
			metric = prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        name,
				ConstLabels: labels,
			})
			metrics.Store(name, metric)
			prometheus.MustRegister(metric)
		}

		logger.DEBUG.Printf("updating metric %s{%s}: %f\n", name, formatLabels(labels), value)
		metric.Set(value)
	}
}

func RecordMetric(command *ConfigCommand, output Output) {
	prefix := "nagios_check_" + command.Name + "_"

	setMetric(prefix+"check_status", command.Labels, float64(output.Status))
	setMetric(prefix+"check_duration", command.Labels, output.Duration.Seconds())
	setMetric(prefix+"check_run_time", command.Labels, float64(time.Now().Unix()))

	for perfKey, perfData := range output.PerfData {
		setMetric(prefix+perfKey+"_value", command.Labels, perfData.Value)
		setMetric(prefix+perfKey+"_min", command.Labels, perfData.Min)
		setMetric(prefix+perfKey+"_max", command.Labels, perfData.Max)
	}
}
