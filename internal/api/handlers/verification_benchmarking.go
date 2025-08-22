package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/external"
	"go.uber.org/zap"
)

// VerificationBenchmarkingHandler handles HTTP requests for verification benchmarking
type VerificationBenchmarkingHandler struct {
	manager *external.VerificationBenchmarkManager
	logger  *zap.Logger
}

// NewVerificationBenchmarkingHandler creates a new verification benchmarking handler
func NewVerificationBenchmarkingHandler(manager *external.VerificationBenchmarkManager, logger *zap.Logger) *VerificationBenchmarkingHandler {
	return &VerificationBenchmarkingHandler{
		manager: manager,
		logger:  logger,
	}
}

// CreateBenchmarkSuiteRequest represents the request to create a benchmark suite
type CreateBenchmarkSuiteRequest struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Category    string                        `json:"category"`
	TestCases   []*external.BenchmarkTestCase `json:"test_cases"`
	Config      map[string]interface{}        `json:"config,omitempty"`
}

// CreateBenchmarkSuiteResponse represents the response from creating a benchmark suite
type CreateBenchmarkSuiteResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// RunBenchmarkRequest represents the request to run a benchmark
type RunBenchmarkRequest struct {
	SuiteID string `json:"suite_id"`
}

// RunBenchmarkResponse represents the response from running a benchmark
type RunBenchmarkResponse struct {
	ID          string                     `json:"id"`
	SuiteID     string                     `json:"suite_id"`
	SuiteName   string                     `json:"suite_name"`
	Status      string                     `json:"status"`
	StartTime   time.Time                  `json:"start_time"`
	EndTime     time.Time                  `json:"end_time"`
	Duration    time.Duration              `json:"duration"`
	TestResults []*external.TestResult     `json:"test_results"`
	Metrics     *external.BenchmarkMetrics `json:"metrics"`
	Summary     string                     `json:"summary"`
}

// GetBenchmarkSuiteResponse represents the response from getting a benchmark suite
type GetBenchmarkSuiteResponse struct {
	Suite *external.BenchmarkSuite `json:"suite"`
}

// ListBenchmarkSuitesResponse represents the response from listing benchmark suites
type ListBenchmarkSuitesResponse struct {
	Suites []*external.BenchmarkSuite `json:"suites"`
	Total  int                        `json:"total"`
}

// GetBenchmarkResultsResponse represents the response from getting benchmark results
type GetBenchmarkResultsResponse struct {
	Results []*external.BenchmarkResult `json:"results"`
	Total   int                         `json:"total"`
}

// CompareBenchmarksRequest represents the request to compare benchmarks
type CompareBenchmarksRequest struct {
	BaselineID   string `json:"baseline_id"`
	ComparisonID string `json:"comparison_id"`
}

// CompareBenchmarksResponse represents the response from comparing benchmarks
type CompareBenchmarksResponse struct {
	Comparison *external.BenchmarkComparison `json:"comparison"`
}

// GetBenchmarkConfigResponse represents the response from getting benchmark configuration
type GetBenchmarkConfigResponse struct {
	Config *external.BenchmarkConfig `json:"config"`
}

// UpdateBenchmarkConfigRequest represents the request to update benchmark configuration
type UpdateBenchmarkConfigRequest struct {
	Config *external.BenchmarkConfig `json:"config"`
}

// UpdateBenchmarkConfigResponse represents the response from updating benchmark configuration
type UpdateBenchmarkConfigResponse struct {
	Message string                    `json:"message"`
	Config  *external.BenchmarkConfig `json:"config"`
}

// RegisterRoutes registers the benchmarking routes
func (h *VerificationBenchmarkingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/benchmarking/suites", h.CreateBenchmarkSuite)
	mux.HandleFunc("GET /api/benchmarking/suites", h.ListBenchmarkSuites)
	mux.HandleFunc("GET /api/benchmarking/suites/{suiteID}", h.GetBenchmarkSuite)
	mux.HandleFunc("POST /api/benchmarking/run", h.RunBenchmark)
	mux.HandleFunc("GET /api/benchmarking/results", h.GetBenchmarkResults)
	mux.HandleFunc("POST /api/benchmarking/compare", h.CompareBenchmarks)
	mux.HandleFunc("GET /api/benchmarking/config", h.GetBenchmarkConfig)
	mux.HandleFunc("PUT /api/benchmarking/config", h.UpdateBenchmarkConfig)
}

// CreateBenchmarkSuite creates a new benchmark suite
func (h *VerificationBenchmarkingHandler) CreateBenchmarkSuite(w http.ResponseWriter, r *http.Request) {
	var req CreateBenchmarkSuiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	// Validate request
	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Suite name is required"})
		return
	}

	if req.Category == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Suite category is required"})
		return
	}

	// Create benchmark suite
	suite := &external.BenchmarkSuite{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		TestCases:   req.TestCases,
		Config:      req.Config,
	}

	if err := h.manager.CreateBenchmarkSuite(suite); err != nil {
		h.logger.Error("Failed to create benchmark suite", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create benchmark suite"})
		return
	}

	response := CreateBenchmarkSuiteResponse{
		ID:        suite.ID,
		Name:      suite.Name,
		Message:   "Benchmark suite created successfully",
		CreatedAt: suite.CreatedAt,
	}

	h.logger.Info("Benchmark suite created",
		zap.String("suite_id", suite.ID),
		zap.String("name", suite.Name))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetBenchmarkSuite retrieves a benchmark suite by ID
func (h *VerificationBenchmarkingHandler) GetBenchmarkSuite(w http.ResponseWriter, r *http.Request) {
	suiteID := r.PathValue("suiteID")
	if suiteID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Suite ID is required"})
		return
	}

	suite, err := h.manager.GetBenchmarkSuite(suiteID)
	if err != nil {
		h.logger.Error("Failed to get benchmark suite", zap.Error(err), zap.String("suite_id", suiteID))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Benchmark suite not found"})
		return
	}

	response := GetBenchmarkSuiteResponse{
		Suite: suite,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListBenchmarkSuites lists all benchmark suites
func (h *VerificationBenchmarkingHandler) ListBenchmarkSuites(w http.ResponseWriter, r *http.Request) {
	suites := h.manager.ListBenchmarkSuites()

	response := ListBenchmarkSuitesResponse{
		Suites: suites,
		Total:  len(suites),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RunBenchmark runs a benchmark suite
func (h *VerificationBenchmarkingHandler) RunBenchmark(w http.ResponseWriter, r *http.Request) {
	var req RunBenchmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if req.SuiteID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Suite ID is required"})
		return
	}

	result, err := h.manager.RunBenchmark(r.Context(), req.SuiteID)
	if err != nil {
		h.logger.Error("Failed to run benchmark", zap.Error(err), zap.String("suite_id", req.SuiteID))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to run benchmark"})
		return
	}

	response := RunBenchmarkResponse{
		ID:          result.ID,
		SuiteID:     result.SuiteID,
		SuiteName:   result.SuiteName,
		Status:      result.Status,
		StartTime:   result.ExecutedAt,
		EndTime:     result.ExecutedAt.Add(result.Duration),
		Duration:    result.Duration,
		TestResults: result.TestResults,
		Metrics:     result.Metrics,
		Summary:     result.Summary,
	}

	h.logger.Info("Benchmark completed",
		zap.String("benchmark_id", result.ID),
		zap.String("suite_id", req.SuiteID),
		zap.String("status", result.Status))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBenchmarkResults retrieves benchmark results
func (h *VerificationBenchmarkingHandler) GetBenchmarkResults(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	results := h.manager.GetBenchmarkResults(limit)

	response := GetBenchmarkResultsResponse{
		Results: results,
		Total:   len(results),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CompareBenchmarks compares two benchmark results
func (h *VerificationBenchmarkingHandler) CompareBenchmarks(w http.ResponseWriter, r *http.Request) {
	var req CompareBenchmarksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if req.BaselineID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Baseline ID is required"})
		return
	}

	if req.ComparisonID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Comparison ID is required"})
		return
	}

	comparison, err := h.manager.CompareBenchmarks(req.BaselineID, req.ComparisonID)
	if err != nil {
		h.logger.Error("Failed to compare benchmarks", zap.Error(err),
			zap.String("baseline_id", req.BaselineID),
			zap.String("comparison_id", req.ComparisonID))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to compare benchmarks"})
		return
	}

	response := CompareBenchmarksResponse{
		Comparison: comparison,
	}

	h.logger.Info("Benchmarks compared",
		zap.String("baseline_id", req.BaselineID),
		zap.String("comparison_id", req.ComparisonID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBenchmarkConfig retrieves the current benchmark configuration
func (h *VerificationBenchmarkingHandler) GetBenchmarkConfig(w http.ResponseWriter, r *http.Request) {
	config := h.manager.GetConfig()

	response := GetBenchmarkConfigResponse{
		Config: config,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateBenchmarkConfig updates the benchmark configuration
func (h *VerificationBenchmarkingHandler) UpdateBenchmarkConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateBenchmarkConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if req.Config == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Config is required"})
		return
	}

	if err := h.manager.UpdateConfig(req.Config); err != nil {
		h.logger.Error("Failed to update config", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to update config: %v", err)})
		return
	}

	response := UpdateBenchmarkConfigResponse{
		Message: "Benchmark configuration updated successfully",
		Config:  req.Config,
	}

	h.logger.Info("Benchmark configuration updated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
