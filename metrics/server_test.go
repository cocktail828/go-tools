package metrics_test

import (
	"testing"
	"time"

	"github.com/cocktail828/go-tools/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestPrometheus(t *testing.T) {
	srv := metrics.NewMetricsServer(nil)
	go srv.Run(":8080")

	// gauge
	gauge := srv.RegisterGauge(metrics.CollectorOpt{
		Namespace: "dbproxy",
		Subsystem: "accesser",
		Name:      "dts_gauge",
	})
	gauge.Set(1)

	// counter
	counter := srv.RegisterCounter(metrics.CollectorOpt{
		Namespace: "dbproxy",
		Subsystem: "accesser",
		Name:      "dts_counter",
	})
	counter.Add(100)

	vec := srv.RegisterCounterVec(metrics.CollectorOpt{
		Namespace: "dbproxy",
		Subsystem: "accesser",
		Name:      "dts_counter_vec",
	}, []string{"code", "desc"})
	vec.WithLabelValues("18934", "kajsdf sdkfj").Inc()

	// histogram
	histogram := srv.RegisterHistogram(metrics.CollectorOpt{
		Namespace: "dbproxy",
		Subsystem: "accesser",
		Name:      "dts_histogram",
	}, prometheus.ExponentialBuckets(0.001, 10, 5))
	timer := prometheus.NewTimer(histogram)
	time.Sleep(time.Millisecond * 100)
	timer.ObserveDuration()

	time.Sleep(time.Hour)
}
