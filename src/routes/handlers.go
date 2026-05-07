package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/eclipse-xfsc/configuration-service/src/metrics"
	"github.com/eclipse-xfsc/configuration-service/src/telemetry"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// ConfigLoadFn allows test injection
var ConfigLoadFn func() (string, string, error)

// ConfigMapFn allows test injection
var ConfigMapFn func(name, namespace string) (map[string]string, error)

// ConfigurationHandler handles GET /configuration requests
func ConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	configmapName, configmapNamespace, err := ConfigLoadFn()
	if err != nil {
		telemetry.Logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	data, err := ConfigMapFn(configmapName, configmapNamespace)
	if err != nil {
		telemetry.Logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// IsAliveHandler handles GET /isAlive requests
func IsAliveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// RequestLogger logs HTTP requests
func RequestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		telemetry.Logger.Infow("",
			zap.String("method", string(r.Method)),
			zap.String("uri", string(r.RequestURI)),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

// Register builds and returns the HTTP router with all handlers
func Register() *mux.Router {
	metrics.InitMetrics()

	router := mux.NewRouter().StrictSlash(true)
	router.Use(metrics.Middleware)
	router.Use(RequestLogger)

	router.HandleFunc("/configuration", ConfigurationHandler).Methods("GET")
	router.HandleFunc("/health", metrics.HealthHandler).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	router.HandleFunc("/isAlive", IsAliveHandler).Methods("GET")

	return router
}

// Start starts the HTTP server on the given port
func Start(port int) {
	router := Register()
	portString := ":" + strconv.Itoa(port)
	log.Fatal(http.ListenAndServe(portString, router))
}
