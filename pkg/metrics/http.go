package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request latencies",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"path", "method", "status"},
	)
)

func init() {
	// Multiple registrations are safe; prometheus ignores duplicates.
	prometheus.MustRegister(httpDuration)
	prometheus.MustRegister(httpRequests)
}

// Middleware instruments HTTP handlers with Prometheus metrics.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		labels := prometheus.Labels{
			"path":   path,
			"method": c.Request.Method,
			"status": strconv.Itoa(c.Writer.Status()),
		}
		elapsed := time.Since(start).Seconds()
		httpDuration.With(labels).Observe(elapsed)
		httpRequests.With(labels).Inc()
	}
}
