package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsInitOnce sync.Once

	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge
)

// Health response structure
type Health struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// statusRecorder wraps http.ResponseWriter to capture status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (recorder *statusRecorder) WriteHeader(code int) {
	recorder.statusCode = code
	recorder.ResponseWriter.WriteHeader(code)
}

// InitMetrics initializes Prometheus metrics (thread-safe)
func InitMetrics() {
	metricsInitOnce.Do(func() {
		httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "configuration_service_http_requests_total",
			Help: "Total number of HTTP requests handled by configuration-service.",
		}, []string{"method", "route", "status"})

		httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "configuration_service_http_request_duration_seconds",
			Help:    "Duration of HTTP requests handled by configuration-service.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route", "status"})

		httpRequestsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "configuration_service_http_in_flight_requests",
			Help: "Current number of in-flight HTTP requests.",
		})

		prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpRequestsInFlight)
	})
}

// Middleware returns a middleware that instruments HTTP requests with metrics
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		next.ServeHTTP(recorder, r)

		route := resolveRouteLabel(r)
		statusCode := strconv.Itoa(recorder.statusCode)

		httpRequestsTotal.WithLabelValues(r.Method, route, statusCode).Inc()
		httpRequestDuration.WithLabelValues(r.Method, route, statusCode).Observe(time.Since(start).Seconds())
	})
}

// resolveRouteLabel extracts the route template from the mux router
func resolveRouteLabel(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if template, err := route.GetPathTemplate(); err == nil {
			return template
		}
	}
	return "unknown"
}

// HealthHandler responds to health checks
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(Health{
		Status:    "OK",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
