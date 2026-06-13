package metrics

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	labelNames = []string{"org_id", "project_id", "user_id", "model_name"}

	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talker_requests_total",
			Help: "Total number of HTTP requests made to the API",
		},
		[]string{"method", "path", "status"},
	)

	TokensConsumedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talker_tokens_consumed_total",
			Help: "Total number of tokens consumed by the Talker API",
		},
		labelNames,
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "talker_request_duration_seconds",
			Help:    "Histogram of request latencies",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

// PrometheusMiddleware creates a middleware that records metrics for HTTP requests.
func PrometheusMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			err := next(c)

			_, status := echo.ResolveResponseStatus(c.Response(), err)

			duration := time.Since(start).Seconds()

			method := c.Request().Method
			path := c.Path()
			statusStr := strconv.Itoa(status)

			RequestsTotal.WithLabelValues(method, path, statusStr).Inc()
			RequestDuration.WithLabelValues(method, path, statusStr).Observe(duration)

			return err
		}
	}
}

// ExtractLabels attempts to extract project_id, org_id, and user_id from headers.
// If not found, defaults to "unknown".
func ExtractLabels(c *echo.Context, modelName string) []string {
	orgID := c.Request().Header.Get("X-Org-Id")
	if orgID == "" {
		orgID = "unknown"
	}

	projectID := c.Request().Header.Get("X-Project-Id")
	if projectID == "" {
		projectID = "unknown"
	}

	userID := c.Request().Header.Get("X-User-Id")
	if userID == "" {
		userID = "unknown"
	}

	if modelName == "" {
		modelName = "unknown"
	}

	return []string{orgID, projectID, userID, modelName}
}
