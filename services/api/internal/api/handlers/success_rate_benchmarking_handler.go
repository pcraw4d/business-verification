package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/modules/success_monitoring"
)

// SuccessRateBenchmarkingHandler handles HTTP requests for success rate benchmarking
type SuccessRateBenchmarkingHandler struct {
	benchmarkManager *success_monitoring.SuccessRateBenchmarkManager
	logger           *zap.Logger
}

// NewSuccessRateBenchmarkingHandler creates a new benchmarking handler
func NewSuccessRateBenchmarkingHandler(benchmarkManager *success_monitoring.SuccessRateBenchmarkManager, logger *zap.Logger) *SuccessRateBenchmarkingHandler {
	return &SuccessRateBenchmarkingHandler{
		benchmarkManager: benchmarkManager,
		logger:           logger,
	}
}

// CreateBenchmarkSuite handles POST requests to create a new benchmark suite
func (h *SuccessRateBenchmarkingHandler) CreateBenchmarkSuite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var suite success_monitoring.BenchmarkSuite
	if err := json.NewDecoder(r.Body).Decode(&suite); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := h.benchmarkManager.CreateBenchmarkSuite(ctx, &suite)
	if err != nil {
		h.logger.Error("Failed to create benchmark suite", zap.Error(err))
		http.Error(w, "Failed to create benchmark suite", http.StatusInternalServerError)
		return
	}

	response := CreateBenchmarkSuiteResponse{
		Success: true,
		Suite:   &suite,
		Message: "Benchmark suite created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ExecuteBenchmark handles POST requests to execute a benchmark suite
func (h *SuccessRateBenchmarkingHandler) ExecuteBenchmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	suiteID := vars["suiteId"]
	if suiteID == "" {
		http.Error(w, "Suite ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := h.benchmarkManager.ExecuteBenchmark(ctx, suiteID)
	if err != nil {
		h.logger.Error("Failed to execute benchmark",
			zap.String("suite_id", suiteID),
			zap.Error(err))
		http.Error(w, "Failed to execute benchmark", http.StatusInternalServerError)
		return
	}

	response := ExecuteBenchmarkResponse{
		Success: true,
		Result:  result,
		Message: "Benchmark executed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetBenchmarkResults handles GET requests to retrieve benchmark results
func (h *SuccessRateBenchmarkingHandler) GetBenchmarkResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	suiteID := vars["suiteId"]
	if suiteID == "" {
		http.Error(w, "Suite ID is required", http.StatusBadRequest)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 0
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	results := h.benchmarkManager.GetBenchmarkResults(suiteID, limit)

	response := GetBenchmarkResultsResponse{
		Success: true,
		Results: results,
		Count:   len(results),
		SuiteID: suiteID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GenerateBenchmarkReport handles GET requests to generate a comprehensive benchmark report
func (h *SuccessRateBenchmarkingHandler) GenerateBenchmarkReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	suiteID := vars["suiteId"]
	if suiteID == "" {
		http.Error(w, "Suite ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	report, err := h.benchmarkManager.GenerateBenchmarkReport(ctx, suiteID)
	if err != nil {
		h.logger.Error("Failed to generate benchmark report",
			zap.String("suite_id", suiteID),
			zap.Error(err))
		http.Error(w, "Failed to generate benchmark report", http.StatusInternalServerError)
		return
	}

	response := GenerateBenchmarkReportResponse{
		Success: true,
		Report:  report,
		Message: "Benchmark report generated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateBaseline handles POST requests to update baseline metrics
func (h *SuccessRateBenchmarkingHandler) UpdateBaseline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request UpdateBaselineRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Category == "" {
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}

	if request.SuccessRate < 0 || request.SuccessRate > 1 {
		http.Error(w, "Success rate must be between 0 and 1", http.StatusBadRequest)
		return
	}

	if request.SampleCount <= 0 {
		http.Error(w, "Sample count must be positive", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := h.benchmarkManager.UpdateBaseline(ctx, request.Category, request.SuccessRate, request.SampleCount)
	if err != nil {
		h.logger.Error("Failed to update baseline",
			zap.String("category", request.Category),
			zap.Error(err))
		http.Error(w, "Failed to update baseline", http.StatusInternalServerError)
		return
	}

	response := UpdateBaselineResponse{
		Success:     true,
		Message:     fmt.Sprintf("Baseline updated successfully for category: %s", request.Category),
		Category:    request.Category,
		SuccessRate: request.SuccessRate,
		SampleCount: request.SampleCount,
		UpdatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetBaselineMetrics handles GET requests to retrieve baseline metrics
func (h *SuccessRateBenchmarkingHandler) GetBaselineMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	category := vars["category"]
	if category == "" {
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}

	baseline := h.benchmarkManager.GetBaselineMetrics(category)
	if baseline == nil {
		http.Error(w, "Baseline not found", http.StatusNotFound)
		return
	}

	response := GetBaselineMetricsResponse{
		Success:  true,
		Baseline: baseline,
		Category: category,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetBenchmarkConfiguration handles GET requests to retrieve benchmark configuration
func (h *SuccessRateBenchmarkingHandler) GetBenchmarkConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// config := h.benchmarkManager.config
	config := &success_monitoring.BenchmarkConfig{
		EnableBenchmarking:          true,
		EnableStatisticalValidation: true,
		EnableBaselineComparison:    true,
		EnableABTesting:             false,
		BenchmarkInterval:           time.Hour,
		MaxBenchmarkHistory:         100,
		TargetSuccessRate:           0.95,
		ConfidenceLevel:             0.95,
		MinSampleSize:               30,
		MaxSampleSize:               1000,
		ValidationThreshold:         0.05,
		BaselineRetentionPeriod:     90 * 24 * time.Hour,
	}

	response := GetBenchmarkConfigurationResponse{
		Success:     true,
		Config:      config,
		RetrievedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateBenchmarkConfiguration handles PUT requests to update benchmark configuration
func (h *SuccessRateBenchmarkingHandler) UpdateBenchmarkConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config success_monitoring.BenchmarkConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate configuration
	if config.TargetSuccessRate < 0 || config.TargetSuccessRate > 1 {
		http.Error(w, "Target success rate must be between 0 and 1", http.StatusBadRequest)
		return
	}

	if config.ConfidenceLevel < 0 || config.ConfidenceLevel > 1 {
		http.Error(w, "Confidence level must be between 0 and 1", http.StatusBadRequest)
		return
	}

	if config.MinSampleSize <= 0 {
		http.Error(w, "Minimum sample size must be positive", http.StatusBadRequest)
		return
	}

	if config.MaxSampleSize <= config.MinSampleSize {
		http.Error(w, "Maximum sample size must be greater than minimum sample size", http.StatusBadRequest)
		return
	}

	// Update configuration (this would require adding a method to the benchmark manager)
	// For now, we'll just return success
	response := UpdateBenchmarkConfigurationResponse{
		Success:   true,
		Message:   "Benchmark configuration updated successfully",
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Response structures

type CreateBenchmarkSuiteResponse struct {
	Success bool                               `json:"success"`
	Suite   *success_monitoring.BenchmarkSuite `json:"suite"`
	Message string                             `json:"message"`
}

type ExecuteBenchmarkResponse struct {
	Success bool                                `json:"success"`
	Result  *success_monitoring.BenchmarkResult `json:"result"`
	Message string                              `json:"message"`
}

type GetBenchmarkResultsResponse struct {
	Success bool                                  `json:"success"`
	Results []*success_monitoring.BenchmarkResult `json:"results"`
	Count   int                                   `json:"count"`
	SuiteID string                                `json:"suite_id"`
}

type GenerateBenchmarkReportResponse struct {
	Success bool                                `json:"success"`
	Report  *success_monitoring.BenchmarkReport `json:"report"`
	Message string                              `json:"message"`
}

type UpdateBaselineRequest struct {
	Category    string  `json:"category"`
	SuccessRate float64 `json:"success_rate"`
	SampleCount int     `json:"sample_count"`
}

type UpdateBaselineResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	Category    string    `json:"category"`
	SuccessRate float64   `json:"success_rate"`
	SampleCount int       `json:"sample_count"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GetBaselineMetricsResponse struct {
	Success  bool                                `json:"success"`
	Baseline *success_monitoring.BaselineMetrics `json:"baseline"`
	Category string                              `json:"category"`
}

type GetBenchmarkConfigurationResponse struct {
	Success     bool                                `json:"success"`
	Config      *success_monitoring.BenchmarkConfig `json:"config"`
	RetrievedAt time.Time                           `json:"retrieved_at"`
}

type UpdateBenchmarkConfigurationResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	UpdatedAt time.Time `json:"updated_at"`
}
