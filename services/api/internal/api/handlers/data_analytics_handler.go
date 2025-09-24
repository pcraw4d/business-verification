package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AnalyticsType represents the type of analytics to perform
type AnalyticsType string

const (
	AnalyticsTypeVerificationTrends AnalyticsType = "verification_trends"
	AnalyticsTypeSuccessRates       AnalyticsType = "success_rates"
	AnalyticsTypeRiskDistribution   AnalyticsType = "risk_distribution"
	AnalyticsTypeIndustryAnalysis   AnalyticsType = "industry_analysis"
	AnalyticsTypeGeographicAnalysis AnalyticsType = "geographic_analysis"
	AnalyticsTypePerformanceMetrics AnalyticsType = "performance_metrics"
	AnalyticsTypeComplianceMetrics  AnalyticsType = "compliance_metrics"
	AnalyticsTypeCustomQuery        AnalyticsType = "custom_query"
	AnalyticsTypePredictiveAnalysis AnalyticsType = "predictive_analysis"
)

// AnalyticsOperation represents the type of analytics operation
type AnalyticsOperation string

const (
	AnalyticsOperationCount            AnalyticsOperation = "count"
	AnalyticsOperationSum              AnalyticsOperation = "sum"
	AnalyticsOperationAverage          AnalyticsOperation = "average"
	AnalyticsOperationMedian           AnalyticsOperation = "median"
	AnalyticsOperationMin              AnalyticsOperation = "min"
	AnalyticsOperationMax              AnalyticsOperation = "max"
	AnalyticsOperationPercentage       AnalyticsOperation = "percentage"
	AnalyticsOperationTrend            AnalyticsOperation = "trend"
	AnalyticsOperationCorrelation      AnalyticsOperation = "correlation"
	AnalyticsOperationPrediction       AnalyticsOperation = "prediction"
	AnalyticsOperationAnomalyDetection AnalyticsOperation = "anomaly_detection"
)

// DataAnalyticsRequest represents a request for data analytics
type DataAnalyticsRequest struct {
	BusinessID          string                 `json:"business_id"`
	AnalyticsType       AnalyticsType          `json:"analytics_type"`
	Operations          []AnalyticsOperation   `json:"operations"`
	Filters             map[string]interface{} `json:"filters,omitempty"`
	TimeRange           *TimeRange             `json:"time_range,omitempty"`
	GroupBy             []string               `json:"group_by,omitempty"`
	OrderBy             []string               `json:"order_by,omitempty"`
	Limit               *int                   `json:"limit,omitempty"`
	Offset              *int                   `json:"offset,omitempty"`
	CustomQuery         string                 `json:"custom_query,omitempty"`
	Parameters          map[string]interface{} `json:"parameters,omitempty"`
	IncludeInsights     bool                   `json:"include_insights"`
	IncludePredictions  bool                   `json:"include_predictions"`
	IncludeTrends       bool                   `json:"include_trends"`
	IncludeCorrelations bool                   `json:"include_correlations"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// DataAnalyticsResponse represents the response from data analytics
type DataAnalyticsResponse struct {
	AnalyticsID    string                 `json:"analytics_id"`
	BusinessID     string                 `json:"business_id"`
	Type           AnalyticsType          `json:"type"`
	Status         string                 `json:"status"`
	IsSuccessful   bool                   `json:"is_successful"`
	Results        []AnalyticsResult      `json:"results"`
	Insights       []AnalyticsInsight     `json:"insights,omitempty"`
	Predictions    []AnalyticsPrediction  `json:"predictions,omitempty"`
	Trends         []AnalyticsTrend       `json:"trends,omitempty"`
	Correlations   []AnalyticsCorrelation `json:"correlations,omitempty"`
	Summary        *AnalyticsSummary      `json:"summary,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	GeneratedAt    time.Time              `json:"generated_at"`
	ProcessingTime string                 `json:"processing_time"`
}

// AnalyticsResult represents a single analytics result
type AnalyticsResult struct {
	Operation  AnalyticsOperation     `json:"operation"`
	Field      string                 `json:"field"`
	Value      interface{}            `json:"value"`
	GroupBy    map[string]interface{} `json:"group_by,omitempty"`
	TimeRange  *TimeRange             `json:"time_range,omitempty"`
	Confidence *float64               `json:"confidence,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsInsight represents an insight from the analytics
type AnalyticsInsight struct {
	Type           string                 `json:"type"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Severity       string                 `json:"severity"`
	Confidence     float64                `json:"confidence"`
	Recommendation string                 `json:"recommendation,omitempty"`
	DataPoints     []interface{}          `json:"data_points,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsPrediction represents a prediction from the analytics
type AnalyticsPrediction struct {
	Field          string                 `json:"field"`
	PredictedValue interface{}            `json:"predicted_value"`
	Confidence     float64                `json:"confidence"`
	TimeHorizon    string                 `json:"time_horizon"`
	Factors        []string               `json:"factors,omitempty"`
	Range          *PredictionRange       `json:"range,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// PredictionRange represents the range for a prediction
type PredictionRange struct {
	Min          interface{} `json:"min"`
	Max          interface{} `json:"max"`
	Percentile25 interface{} `json:"percentile_25,omitempty"`
	Percentile75 interface{} `json:"percentile_75,omitempty"`
}

// AnalyticsTrend represents a trend from the analytics
type AnalyticsTrend struct {
	Field      string                 `json:"field"`
	Direction  string                 `json:"direction"`
	Slope      float64                `json:"slope"`
	Strength   float64                `json:"strength"`
	TimeRange  *TimeRange             `json:"time_range"`
	DataPoints []TrendDataPoint       `json:"data_points"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TrendDataPoint represents a data point in a trend
type TrendDataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     interface{}            `json:"value"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsCorrelation represents a correlation from the analytics
type AnalyticsCorrelation struct {
	Field1       string                 `json:"field1"`
	Field2       string                 `json:"field2"`
	Coefficient  float64                `json:"coefficient"`
	Strength     string                 `json:"strength"`
	Significance float64                `json:"significance"`
	DataPoints   []CorrelationDataPoint `json:"data_points,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CorrelationDataPoint represents a data point in a correlation
type CorrelationDataPoint struct {
	Value1   interface{}            `json:"value1"`
	Value2   interface{}            `json:"value2"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsSummary represents a summary of analytics results
type AnalyticsSummary struct {
	TotalRecords    int                    `json:"total_records"`
	TimeRange       *TimeRange             `json:"time_range,omitempty"`
	KeyMetrics      map[string]interface{} `json:"key_metrics,omitempty"`
	TopInsights     []string               `json:"top_insights,omitempty"`
	Recommendations []string               `json:"recommendations,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsJob represents a background analytics job
type AnalyticsJob struct {
	JobID           string                 `json:"job_id"`
	BusinessID      string                 `json:"business_id"`
	Type            AnalyticsType          `json:"type"`
	Status          JobStatus              `json:"status"`
	Progress        float64                `json:"progress"`
	TotalSteps      int                    `json:"total_steps"`
	CurrentStep     int                    `json:"current_step"`
	StepDescription string                 `json:"step_description"`
	Result          *DataAnalyticsResponse `json:"result,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsSchema represents a pre-configured analytics schema
type AnalyticsSchema struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Type                AnalyticsType          `json:"type"`
	Operations          []AnalyticsOperation   `json:"operations"`
	DefaultFilters      map[string]interface{} `json:"default_filters,omitempty"`
	DefaultGroupBy      []string               `json:"default_group_by,omitempty"`
	DefaultOrderBy      []string               `json:"default_order_by,omitempty"`
	Parameters          map[string]interface{} `json:"parameters,omitempty"`
	IncludeInsights     bool                   `json:"include_insights"`
	IncludePredictions  bool                   `json:"include_predictions"`
	IncludeTrends       bool                   `json:"include_trends"`
	IncludeCorrelations bool                   `json:"include_correlations"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// DataAnalyticsHandler handles data analytics requests
type DataAnalyticsHandler struct {
	logger *zap.Logger
	jobs   map[string]*AnalyticsJob
	mu     sync.RWMutex
}

// NewDataAnalyticsHandler creates a new data analytics handler
func NewDataAnalyticsHandler(logger *zap.Logger) *DataAnalyticsHandler {
	return &DataAnalyticsHandler{
		logger: logger,
		jobs:   make(map[string]*AnalyticsJob),
	}
}

// AnalyzeData performs immediate data analytics
func (h *DataAnalyticsHandler) AnalyzeData(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	// Parse request
	var req DataAnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateAnalyticsRequest(&req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		http.Error(w, fmt.Sprintf("validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Generate analytics ID
	analyticsID := fmt.Sprintf("analytics_%d_%d", time.Now().Unix(), 1)

	// Perform analytics (simulated)
	results, insights, predictions, trends, correlations, summary, err := h.performAnalytics(ctx, &req)
	if err != nil {
		h.logger.Error("analytics failed", zap.Error(err))
		http.Error(w, "analytics processing failed", http.StatusInternalServerError)
		return
	}

	// Create response
	response := &DataAnalyticsResponse{
		AnalyticsID:    analyticsID,
		BusinessID:     req.BusinessID,
		Type:           req.AnalyticsType,
		Status:         "success",
		IsSuccessful:   true,
		Results:        results,
		Insights:       insights,
		Predictions:    predictions,
		Trends:         trends,
		Correlations:   correlations,
		Summary:        summary,
		Metadata:       req.Metadata,
		GeneratedAt:    time.Now(),
		ProcessingTime: time.Since(startTime).String(),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("analytics completed successfully",
		zap.String("analytics_id", analyticsID),
		zap.String("business_id", req.BusinessID),
		zap.String("type", string(req.AnalyticsType)),
		zap.Duration("processing_time", time.Since(startTime)))
}

// CreateAnalyticsJob creates a background analytics job
func (h *DataAnalyticsHandler) CreateAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req DataAnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateAnalyticsRequest(&req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		http.Error(w, fmt.Sprintf("validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Generate job ID
	jobID := fmt.Sprintf("analytics_job_%d_%d", time.Now().Unix(), 1)

	// Create job
	job := &AnalyticsJob{
		JobID:           jobID,
		BusinessID:      req.BusinessID,
		Type:            req.AnalyticsType,
		Status:          JobStatusPending,
		Progress:        0.0,
		TotalSteps:      6,
		CurrentStep:     0,
		StepDescription: "Initializing analytics job",
		CreatedAt:       time.Now(),
		Metadata:        req.Metadata,
	}

	// Store job
	h.mu.Lock()
	h.jobs[jobID] = job
	h.mu.Unlock()

	// Start background processing
	go h.processAnalyticsJob(ctx, job, &req)

	// Return job
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)

	h.logger.Info("analytics job created",
		zap.String("job_id", jobID),
		zap.String("business_id", req.BusinessID),
		zap.String("type", string(req.AnalyticsType)))
}

// GetAnalyticsJob retrieves the status of an analytics job
func (h *DataAnalyticsHandler) GetAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	job, exists := h.jobs[jobID]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "analytics job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}

// ListAnalyticsJobs lists all analytics jobs
func (h *DataAnalyticsHandler) ListAnalyticsJobs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	businessID := r.URL.Query().Get("business_id")
	analyticsType := r.URL.Query().Get("analytics_type")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Filter jobs
	h.mu.RLock()
	var filteredJobs []*AnalyticsJob
	totalCount := 0

	for _, job := range h.jobs {
		// Apply filters
		if status != "" && string(job.Status) != status {
			continue
		}
		if businessID != "" && job.BusinessID != businessID {
			continue
		}
		if analyticsType != "" && string(job.Type) != analyticsType {
			continue
		}

		totalCount++
		if len(filteredJobs) < limit && len(filteredJobs) >= offset {
			filteredJobs = append(filteredJobs, job)
		}
	}
	h.mu.RUnlock()

	// Create response
	response := map[string]interface{}{
		"jobs":        filteredJobs,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAnalyticsSchema retrieves a pre-configured analytics schema
func (h *DataAnalyticsHandler) GetAnalyticsSchema(w http.ResponseWriter, r *http.Request) {
	schemaID := r.URL.Query().Get("schema_id")
	if schemaID == "" {
		http.Error(w, "schema_id is required", http.StatusBadRequest)
		return
	}

	// Get schema (simulated)
	schema := h.getAnalyticsSchema(schemaID)
	if schema == nil {
		http.Error(w, "analytics schema not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schema)
}

// ListAnalyticsSchemas lists all available analytics schemas
func (h *DataAnalyticsHandler) ListAnalyticsSchemas(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	analyticsType := r.URL.Query().Get("analytics_type")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get schemas (simulated)
	schemas := h.getAnalyticsSchemas(analyticsType, limit, offset)

	// Create response
	response := map[string]interface{}{
		"schemas":     schemas,
		"total_count": len(schemas),
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// validateAnalyticsRequest validates the analytics request
func (h *DataAnalyticsHandler) validateAnalyticsRequest(req *DataAnalyticsRequest) error {
	if req.BusinessID == "" {
		return fmt.Errorf("business_id is required")
	}

	if req.AnalyticsType == "" {
		return fmt.Errorf("analytics_type is required")
	}

	if len(req.Operations) == 0 {
		return fmt.Errorf("at least one operation is required")
	}

	// Validate analytics type
	switch req.AnalyticsType {
	case AnalyticsTypeVerificationTrends, AnalyticsTypeSuccessRates, AnalyticsTypeRiskDistribution,
		AnalyticsTypeIndustryAnalysis, AnalyticsTypeGeographicAnalysis, AnalyticsTypePerformanceMetrics,
		AnalyticsTypeComplianceMetrics, AnalyticsTypeCustomQuery, AnalyticsTypePredictiveAnalysis:
	default:
		return fmt.Errorf("invalid analytics_type: %s", req.AnalyticsType)
	}

	// Validate operations
	for _, op := range req.Operations {
		switch op {
		case AnalyticsOperationCount, AnalyticsOperationSum, AnalyticsOperationAverage,
			AnalyticsOperationMedian, AnalyticsOperationMin, AnalyticsOperationMax,
			AnalyticsOperationPercentage, AnalyticsOperationTrend, AnalyticsOperationCorrelation,
			AnalyticsOperationPrediction, AnalyticsOperationAnomalyDetection:
		default:
			return fmt.Errorf("invalid operation: %s", op)
		}
	}

	// Validate custom query for custom query type
	if req.AnalyticsType == AnalyticsTypeCustomQuery && req.CustomQuery == "" {
		return fmt.Errorf("custom_query is required for custom_query analytics type")
	}

	return nil
}

// performAnalytics performs the actual analytics (simulated)
func (h *DataAnalyticsHandler) performAnalytics(ctx context.Context, req *DataAnalyticsRequest) (
	[]AnalyticsResult, []AnalyticsInsight, []AnalyticsPrediction, []AnalyticsTrend, []AnalyticsCorrelation, *AnalyticsSummary, error) {

	// Simulate analytics processing
	results := []AnalyticsResult{
		{
			Operation: AnalyticsOperationCount,
			Field:     "verifications",
			Value:     1500,
			GroupBy:   map[string]interface{}{"status": "completed"},
		},
		{
			Operation:  AnalyticsOperationAverage,
			Field:      "success_rate",
			Value:      0.95,
			Confidence: func() *float64 { v := 0.98; return &v }(),
		},
	}

	insights := []AnalyticsInsight{
		{
			Type:           "trend",
			Title:          "Increasing Verification Success Rate",
			Description:    "Success rate has increased by 5% over the last 30 days",
			Severity:       "low",
			Confidence:     0.85,
			Recommendation: "Continue monitoring the trend and investigate contributing factors",
		},
	}

	predictions := []AnalyticsPrediction{
		{
			Field:          "monthly_verifications",
			PredictedValue: 1800,
			Confidence:     0.92,
			TimeHorizon:    "30_days",
			Factors:        []string{"seasonal_trends", "market_growth"},
			Range: &PredictionRange{
				Min:          1700,
				Max:          1900,
				Percentile25: 1750,
				Percentile75: 1850,
			},
		},
	}

	trends := []AnalyticsTrend{
		{
			Field:     "verification_volume",
			Direction: "increasing",
			Slope:     0.15,
			Strength:  0.78,
			TimeRange: &TimeRange{
				Start: time.Now().AddDate(0, -1, 0),
				End:   time.Now(),
			},
			DataPoints: []TrendDataPoint{
				{Timestamp: time.Now().AddDate(0, -1, 0), Value: 1200},
				{Timestamp: time.Now(), Value: 1500},
			},
		},
	}

	correlations := []AnalyticsCorrelation{
		{
			Field1:       "verification_volume",
			Field2:       "success_rate",
			Coefficient:  0.65,
			Strength:     "moderate",
			Significance: 0.01,
		},
	}

	summary := &AnalyticsSummary{
		TotalRecords: 1500,
		TimeRange: &TimeRange{
			Start: time.Now().AddDate(0, -1, 0),
			End:   time.Now(),
		},
		KeyMetrics: map[string]interface{}{
			"total_verifications":     1500,
			"success_rate":            0.95,
			"average_processing_time": "2.5s",
		},
		TopInsights: []string{
			"Success rate is trending upward",
			"Verification volume is increasing",
		},
		Recommendations: []string{
			"Monitor success rate trends",
			"Consider scaling resources for increased volume",
		},
	}

	return results, insights, predictions, trends, correlations, summary, nil
}

// processAnalyticsJob processes a background analytics job
func (h *DataAnalyticsHandler) processAnalyticsJob(ctx context.Context, job *AnalyticsJob, req *DataAnalyticsRequest) {
	startTime := time.Now()

	// Update job status
	h.updateJobStatus(job, JobStatusProcessing, 0.1, 1, "Validating request parameters")

	// Simulate processing steps
	time.Sleep(100 * time.Millisecond)
	h.updateJobStatus(job, JobStatusProcessing, 0.3, 2, "Collecting data")

	time.Sleep(200 * time.Millisecond)
	h.updateJobStatus(job, JobStatusProcessing, 0.5, 3, "Performing analytics calculations")

	time.Sleep(300 * time.Millisecond)
	h.updateJobStatus(job, JobStatusProcessing, 0.7, 4, "Generating insights and predictions")

	time.Sleep(200 * time.Millisecond)
	h.updateJobStatus(job, JobStatusProcessing, 0.9, 5, "Finalizing results")

	// Perform analytics
	results, insights, predictions, trends, correlations, summary, err := h.performAnalytics(ctx, req)
	if err != nil {
		h.updateJobStatus(job, JobStatusFailed, 1.0, 6, "Analytics processing failed")
		return
	}

	// Create result
	result := &DataAnalyticsResponse{
		AnalyticsID:    job.JobID,
		BusinessID:     job.BusinessID,
		Type:           job.Type,
		Status:         "success",
		IsSuccessful:   true,
		Results:        results,
		Insights:       insights,
		Predictions:    predictions,
		Trends:         trends,
		Correlations:   correlations,
		Summary:        summary,
		Metadata:       job.Metadata,
		GeneratedAt:    time.Now(),
		ProcessingTime: time.Since(startTime).String(),
	}

	// Update job with result
	h.updateJobWithResult(job, result, JobStatusCompleted, 1.0, 6, "Analytics completed successfully")

	h.logger.Info("analytics job completed",
		zap.String("job_id", job.JobID),
		zap.String("business_id", job.BusinessID),
		zap.String("type", string(job.Type)),
		zap.Duration("processing_time", time.Since(startTime)))
}

// updateJobStatus updates the status of a job
func (h *DataAnalyticsHandler) updateJobStatus(job *AnalyticsJob, status JobStatus, progress float64, currentStep int, stepDescription string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	job.Status = status
	job.Progress = progress
	job.CurrentStep = currentStep
	job.StepDescription = stepDescription

	if status == JobStatusProcessing && job.StartedAt == nil {
		now := time.Now()
		job.StartedAt = &now
	}
}

// updateJobWithResult updates a job with its result
func (h *DataAnalyticsHandler) updateJobWithResult(job *AnalyticsJob, result *DataAnalyticsResponse, status JobStatus, progress float64, currentStep int, stepDescription string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	job.Status = status
	job.Progress = progress
	job.CurrentStep = currentStep
	job.StepDescription = stepDescription
	job.Result = result

	now := time.Now()
	job.CompletedAt = &now
}

// getAnalyticsSchema retrieves a pre-configured analytics schema (simulated)
func (h *DataAnalyticsHandler) getAnalyticsSchema(schemaID string) *AnalyticsSchema {
	// Simulated schemas
	schemas := map[string]*AnalyticsSchema{
		"verification_trends_schema": {
			ID:          "verification_trends_schema",
			Name:        "Verification Trends Analysis",
			Description: "Analyze verification trends over time",
			Type:        AnalyticsTypeVerificationTrends,
			Operations:  []AnalyticsOperation{AnalyticsOperationCount, AnalyticsOperationTrend},
			DefaultFilters: map[string]interface{}{
				"status": "completed",
			},
			DefaultGroupBy:  []string{"date"},
			DefaultOrderBy:  []string{"date"},
			IncludeInsights: true,
			IncludeTrends:   true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		"success_rates_schema": {
			ID:                 "success_rates_schema",
			Name:               "Success Rate Analysis",
			Description:        "Analyze verification success rates",
			Type:               AnalyticsTypeSuccessRates,
			Operations:         []AnalyticsOperation{AnalyticsOperationAverage, AnalyticsOperationPercentage},
			DefaultGroupBy:     []string{"industry"},
			IncludeInsights:    true,
			IncludePredictions: true,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}

	return schemas[schemaID]
}

// getAnalyticsSchemas retrieves all analytics schemas (simulated)
func (h *DataAnalyticsHandler) getAnalyticsSchemas(analyticsType string, limit, offset int) []*AnalyticsSchema {
	// Simulated schemas
	allSchemas := []*AnalyticsSchema{
		{
			ID:              "verification_trends_schema",
			Name:            "Verification Trends Analysis",
			Description:     "Analyze verification trends over time",
			Type:            AnalyticsTypeVerificationTrends,
			Operations:      []AnalyticsOperation{AnalyticsOperationCount, AnalyticsOperationTrend},
			IncludeInsights: true,
			IncludeTrends:   true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:                 "success_rates_schema",
			Name:               "Success Rate Analysis",
			Description:        "Analyze verification success rates",
			Type:               AnalyticsTypeSuccessRates,
			Operations:         []AnalyticsOperation{AnalyticsOperationAverage, AnalyticsOperationPercentage},
			IncludeInsights:    true,
			IncludePredictions: true,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		{
			ID:                  "risk_distribution_schema",
			Name:                "Risk Distribution Analysis",
			Description:         "Analyze risk distribution across verifications",
			Type:                AnalyticsTypeRiskDistribution,
			Operations:          []AnalyticsOperation{AnalyticsOperationCount, AnalyticsOperationAverage},
			IncludeInsights:     true,
			IncludeCorrelations: true,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		},
	}

	// Filter by type if specified
	var filteredSchemas []*AnalyticsSchema
	for _, schema := range allSchemas {
		if analyticsType == "" || string(schema.Type) == analyticsType {
			filteredSchemas = append(filteredSchemas, schema)
		}
	}

	// Apply pagination
	start := offset
	end := start + limit
	if start >= len(filteredSchemas) {
		return []*AnalyticsSchema{}
	}
	if end > len(filteredSchemas) {
		end = len(filteredSchemas)
	}

	return filteredSchemas[start:end]
}

// String conversion methods for enums
func (at AnalyticsType) String() string {
	return string(at)
}

func (ao AnalyticsOperation) String() string {
	return string(ao)
}
