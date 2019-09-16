package resizer

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var inFlightGauge prometheus.Gauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "in_flight_requests",
	Help: "A gauge of requests currently being served by the wrapped handler.",
})

var counter *prometheus.CounterVec = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "A counter for requests to the wrapped handler.",
	},
	[]string{"code", "method"},
)

var duration *prometheus.HistogramVec = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "A histogram of latencies for requests.",
		Buckets: []float64{.025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	},
	[]string{"handler"},
)

var responseSize *prometheus.HistogramVec = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_response_size_bytes",
		Help:    "A histogram of response sizes for requests.",
		Buckets: prometheus.ExponentialBuckets(256, 4, 8),
	},
	[]string{"handler"},
)

var compressionRatio *prometheus.HistogramVec = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "resizer_transform_compresion_ratio",
		Help:    "A histogram of size resulting compressed images (0.5 means the size was halfed)",
		Buckets: prometheus.LinearBuckets(.2, .2, 5),
	},
	[]string{},
)

func init() {
	prometheus.MustRegister(inFlightGauge, counter, duration, responseSize, compressionRatio)
}

func handlerMetric(h http.Handler, handlerName string) http.Handler {
	h = promhttp.InstrumentHandlerResponseSize(responseSize.MustCurryWith(prometheus.Labels{"handler": handlerName}), h)
	h = promhttp.InstrumentHandlerCounter(counter, h)
	h = promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": handlerName}), h)
	return promhttp.InstrumentHandlerInFlight(inFlightGauge, h)
}
