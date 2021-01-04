package main

import (
	"log"
	"net/http"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	hc = http.Client{}
)

func main() {
	retriever := NewDockerHubRetriever(checkImageLabel, checkImageTag, tokenInfo.Get)

	remainingPullsMetric := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "dockerhub_limit_remaining_requests_total",
		Help: "Docker Hub Rate Limit Remaining Requests",
	})


	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "dockerhub_limit_max_requests_total",
			Help: "Docker Hub Rate Limit Maximum Requests",
		},
		func() float64 {
			limits, err := retriever.GetLimits()
			if err != nil {
				log.Printf("cannot retrieve limits: %s", err.Error())
				return 0
			}

			remainingPullsMetric.Set(float64(limits.Remaining.Limit))

			return float64(limits.Total.Limit)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "runtime",
			Name:      "goroutines_count",
			Help:      "Number of goroutines that currently exist.",
		},
		func() float64 {
			return float64(runtime.NumGoroutine())
		},
	))

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":8881", nil); err != nil {
		log.Panicf("cannot start http server: %s", err.Error())
	}
}
