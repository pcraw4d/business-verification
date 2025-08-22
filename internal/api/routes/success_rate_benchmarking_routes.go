package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/modules/success_monitoring"
)

// RegisterSuccessRateBenchmarkingRoutes registers all success rate benchmarking routes
func RegisterSuccessRateBenchmarkingRoutes(
	router *mux.Router,
	benchmarkManager *success_monitoring.SuccessRateBenchmarkManager,
	logger *zap.Logger,
) {
	// Create the benchmarking handler
	benchmarkHandler := handlers.NewSuccessRateBenchmarkingHandler(benchmarkManager, logger)

	// Create a subrouter for benchmarking routes
	benchmarkRouter := router.PathPrefix("/api/v3/benchmarking").Subrouter()

	// Benchmark suite management routes
	benchmarkRouter.HandleFunc("/suites", benchmarkHandler.CreateBenchmarkSuite).Methods(http.MethodPost)
	benchmarkRouter.HandleFunc("/suites/{suiteId}/execute", benchmarkHandler.ExecuteBenchmark).Methods(http.MethodPost)
	benchmarkRouter.HandleFunc("/suites/{suiteId}/results", benchmarkHandler.GetBenchmarkResults).Methods(http.MethodGet)
	benchmarkRouter.HandleFunc("/suites/{suiteId}/report", benchmarkHandler.GenerateBenchmarkReport).Methods(http.MethodGet)

	// Baseline management routes
	benchmarkRouter.HandleFunc("/baselines", benchmarkHandler.UpdateBaseline).Methods(http.MethodPost)
	benchmarkRouter.HandleFunc("/baselines/{category}", benchmarkHandler.GetBaselineMetrics).Methods(http.MethodGet)

	// Configuration management routes
	benchmarkRouter.HandleFunc("/config", benchmarkHandler.GetBenchmarkConfiguration).Methods(http.MethodGet)
	benchmarkRouter.HandleFunc("/config", benchmarkHandler.UpdateBenchmarkConfiguration).Methods(http.MethodPut)

	// Add middleware for logging and monitoring
	benchmarkRouter.Use(
		LoggingMiddleware(logger),
		MonitoringMiddleware("success_rate_benchmarking"),
	)

	logger.Info("Success rate benchmarking routes registered",
		zap.String("base_path", "/api/v3/benchmarking"),
		zap.Int("route_count", 7),
	)
}

// LoggingMiddleware adds request logging for benchmarking endpoints
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Benchmarking API request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
			next.ServeHTTP(w, r)
		})
	}
}

// MonitoringMiddleware adds monitoring for benchmarking endpoints
func MonitoringMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add monitoring headers
			w.Header().Set("X-Service", serviceName)
			w.Header().Set("X-API-Version", "v3")

			// Add timing header
			w.Header().Set("X-Request-Start", "t="+string(r.Header.Get("X-Request-Start")))

			next.ServeHTTP(w, r)
		})
	}
}
