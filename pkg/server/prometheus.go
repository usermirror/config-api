package server

import (
	"strconv"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var (
	// prometheusMetricsEndpoint is the path scraped for metrics
	prometheusMetricsEndpoint = "/metrics"
)

// newPromExporter creates a new instance for monitoring.
func newPromExporter(subsystem string) *promExporter {
	RequestDurationBucket := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 20, 30, 40, 50, 60}

	exporter := &promExporter{
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Number of HTTP requests",
			},
			[]string{"code", "method"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				Buckets:   RequestDurationBucket,
			},
			[]string{"code"},
		),
	}

	prometheus.MustRegister(
		exporter.requestCount,
		exporter.requestDuration,
	)
	return exporter
}

// promExporter collects metrics about HTTP.
type promExporter struct {
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// SetupMiddleware configures the router for request tracking.
func (e *promExporter) SetupMiddleware(r *fasthttprouter.Router) fasthttp.RequestHandler {
	// handle metrics requests
	r.GET(prometheusMetricsEndpoint, metricsHandler())

	return func(ctx *fasthttp.RequestCtx) {
		// record time
		start := time.Now()

		// handle request
		r.Handler(ctx)

		// don't record stats for metrics
		if string(ctx.Request.URI().Path()) == prometheusMetricsEndpoint {
			return
		}

		// calculate metric values
		status := strconv.Itoa(ctx.Response.StatusCode())
		method := string(ctx.Method())
		elapsed := float64(time.Since(start)) / float64(time.Second)

		// record metrics
		e.requestCount.WithLabelValues(status, method).Inc()
		e.requestDuration.WithLabelValues(status).Observe(elapsed)
	}
}

// metricsHandler returns metrics for Prometheus to scrape
func metricsHandler() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
}
