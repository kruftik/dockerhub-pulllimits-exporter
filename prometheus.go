package main

import (
	"log"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	dockerHubMetricsLabels = []string{"interval_sec"}

	dockerHubMaxLimitMetricDesc = prometheus.NewDesc(
		"dockerhub_limit_max_requests_total",
		"Docker Hub maximum requests limit",
		dockerHubMetricsLabels,
		nil,
	)
	dockerHubRemainingLimitMetricDesc = prometheus.NewDesc(
		"dockerhub_limit_remaining_requests_total",
		"Docker Hub remaining requests limit",
		dockerHubMetricsLabels,
		nil,
	)
)

type DockerHubLimitsCollector struct {
	retriever *DockerHubRetriever
}

func (dhlc DockerHubLimitsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(dhlc, ch)
}

func (dhlc DockerHubLimitsCollector) Collect(ch chan<- prometheus.Metric) {
	limits, err := retriever.GetLimits()
	if err != nil {
		log.Printf("cannot retrieve limits: %s", err.Error())
		return
	}

	ch <- prometheus.MustNewConstMetric(
		dockerHubMaxLimitMetricDesc,
		prometheus.GaugeValue,
		float64(limits.Total.Limit),
		strconv.Itoa(int(limits.Total.RefreshDuration/time.Second)),
	)

	ch <- prometheus.MustNewConstMetric(
		dockerHubRemainingLimitMetricDesc,
		prometheus.GaugeValue,
		float64(limits.Remaining.Limit),
		strconv.Itoa(int(limits.Remaining.RefreshDuration/time.Second)),
	)
}

func NewDockerHubLimitsCollector() DockerHubLimitsCollector {
	dhlc := DockerHubLimitsCollector{}

	return dhlc
}

func RegisterCollectors() (*prometheus.Registry, error) {
	reg := prometheus.NewPedanticRegistry()

	reg.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	dhlc := NewDockerHubLimitsCollector()

	prometheus.WrapRegistererWith(prometheus.Labels{}, reg).MustRegister(dhlc)

	return reg, nil
}
