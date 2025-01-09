package middlewares

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests.",
		Buckets: prometheus.DefBuckets,
	}, []string{"job", "method", "path"})

	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"job", "method", "path", "status"})
)

func init() {
	prometheus.MustRegister(httpDuration)
	prometheus.MustRegister(httpRequestsTotal)
}

// preprocessPath ensures that the path has exactly one leading slash
func preprocessPath(path string) string {
	trimmedPath := strings.TrimPrefix(path, "/")
	return "/" + strings.Trim(trimmedPath, "/")
}

// normalizePath strips out variable parts of the path (e.g., digits or long strings) and normalizes it
func normalizePath(path string) string {
	path = preprocessPath(path)
	re := regexp.MustCompile(`\b(?:[a-zA-Z]*ID|[a-zA-Z]*id|\d+|[a-zA-Z0-9-_]{20,})\b`)
	return re.ReplaceAllString(path, "id")
}

// stripQueryParams removes the query parameters from the URL path
func stripQueryParams(url string) string {
	parts := strings.Split(url, "?")
	return parts[0]
}

// MetricsMiddleware is the HTTP middleware for tracking metrics
func MetricsMiddleware(job string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(recorder, r)
			duration := time.Since(start).Seconds()
			// Strip query parameters
			strippedPath := stripQueryParams(r.URL.Path)
			normalizedPath := normalizePath(strippedPath)
			statusCode := recorder.statusCode

			httpDuration.WithLabelValues(job, r.Method, normalizedPath).Observe(duration)
			if r.Method == "OPTIONS" || r.Method == "HEAD" {
				return
			}
			httpRequestsTotal.WithLabelValues(job, r.Method, normalizedPath, strconv.Itoa(statusCode)).Inc()
		})
	}
}

// statusRecorder is a custom http.ResponseWriter to capture the response status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	return rec.ResponseWriter.Write(b)
}
