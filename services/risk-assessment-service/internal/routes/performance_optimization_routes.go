package routes

import (
	"net/http"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/handlers"
)

// PerformanceOptimizationRoutes sets up performance optimization routes
func PerformanceOptimizationRoutes(
	performanceOptimizationHandlers *handlers.PerformanceOptimizationHandlers,
	logger *zap.Logger,
) {
	// Performance optimization status
	http.HandleFunc("/api/v1/performance/optimization/status",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			performanceOptimizationHandlers.GetOptimizationStatus(w, r)
		})

	// Trigger immediate optimization
	http.HandleFunc("/api/v1/performance/optimization/optimize",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			performanceOptimizationHandlers.OptimizeNow(w, r)
		})

	// Database optimization
	http.HandleFunc("/api/v1/performance/optimization/database",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				performanceOptimizationHandlers.GetDatabaseOptimization(w, r)
			case http.MethodPost:
				performanceOptimizationHandlers.OptimizeDatabase(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

	// Response time optimization
	http.HandleFunc("/api/v1/performance/optimization/response-time",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				performanceOptimizationHandlers.GetResponseTimeOptimization(w, r)
			case http.MethodPost:
				performanceOptimizationHandlers.OptimizeResponseTime(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

	// Set response time targets
	http.HandleFunc("/api/v1/performance/optimization/targets",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				performanceOptimizationHandlers.GetPerformanceTargets(w, r)
			case http.MethodPut:
				performanceOptimizationHandlers.SetResponseTimeTargets(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

	// Get optimization recommendations
	http.HandleFunc("/api/v1/performance/optimization/recommendations",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			performanceOptimizationHandlers.GetOptimizationRecommendations(w, r)
		})

	// Record response time
	http.HandleFunc("/api/v1/performance/optimization/record-response-time",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			performanceOptimizationHandlers.RecordResponseTime(w, r)
		})

	// Reset optimization data
	http.HandleFunc("/api/v1/performance/optimization/reset",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			performanceOptimizationHandlers.ResetOptimization(w, r)
		})

	logger.Info("Performance optimization routes registered")
}
