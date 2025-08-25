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

	"github.com/pcraw4d/business-verification/internal/observability"
)

// AggregationType represents the type of aggregation to perform
type AggregationType string

const (
	AggregationTypeBusinessMetrics      AggregationType = "business_metrics"
	AggregationTypeRiskAssessments      AggregationType = "risk_assessments"
	AggregationTypeComplianceReports    AggregationType = "compliance_reports"
	AggregationTypePerformanceAnalytics AggregationType = "performance_analytics"
	AggregationTypeTrendAnalysis        AggregationType = "trend_analysis"
	AggregationTypeCustom               AggregationType = "custom"
	AggregationTypeAll                  AggregationType = "all"
)

// AggregationOperation represents a specific aggregation operation
type AggregationOperation string

const (
	AggregationOperationCount      AggregationOperation = "count"
	AggregationOperationSum        AggregationOperation = "sum"
	AggregationOperationAverage    AggregationOperation = "average"
	AggregationOperationMin        AggregationOperation = "min"
	AggregationOperationMax        AggregationOperation = "max"
	AggregationOperationMedian     AggregationOperation = "median"
	AggregationOperationPercentile AggregationOperation = "percentile"
	AggregationOperationGroupBy    AggregationOperation = "group_by"
	AggregationOperationPivot      AggregationOperation = "pivot"
	AggregationOperationCustom     AggregationOperation = "custom"
)

// DataAggregationRule represents an aggregation rule
type DataAggregationRule struct {
	Field       string                 `json:"field"`
	Operation   AggregationOperation   `json:"operation"`
	Parameters  map[string]interface{} `json:"parameters"`
	Condition   string                 `json:"condition,omitempty"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Order       int                    `json:"order"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DataAggregationRequest represents a request to aggregate data
type DataAggregationRequest struct {
	BusinessID      string                 `json:"business_id,omitempty"`
	AggregationType AggregationType        `json:"aggregation_type"`
	Data            interface{}            `json:"data"`
	Rules           []DataAggregationRule  `json:"rules,omitempty"`
	SchemaID        string                 `json:"schema_id,omitempty"`
	GroupBy         []string               `json:"group_by,omitempty"`
	Filters         map[string]interface{} `json:"filters,omitempty"`
	TimeRange       *TimeRange             `json:"time_range,omitempty"`
	IncludeMetadata bool                   `json:"include_metadata"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TimeRange represents a time range for aggregation
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// DataAggregationResponse represents the response from a data aggregation request
type DataAggregationResponse struct {
	AggregationID   string                 `json:"aggregation_id"`
	BusinessID      string                 `json:"business_id,omitempty"`
	AggregationType AggregationType        `json:"aggregation_type"`
	Status          string                 `json:"status"` // success, partial, failed
	IsSuccessful    bool                   `json:"is_successful"`
	OriginalData    interface{}            `json:"original_data,omitempty"`
	AggregatedData  interface{}            `json:"aggregated_data"`
	AppliedRules    []DataAggregationRule  `json:"applied_rules"`
	SkippedRules    []DataAggregationRule  `json:"skipped_rules,omitempty"`
	FailedRules     []DataAggregationRule  `json:"failed_rules,omitempty"`
	Summary         map[string]interface{} `json:"summary"`
	GroupedResults  map[string]interface{} `json:"grouped_results,omitempty"`
	TimeRange       *TimeRange             `json:"time_range,omitempty"`
	AggregatedAt    time.Time              `json:"aggregated_at"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AggregationJob represents a background aggregation job
type AggregationJob struct {
	ID              string                   `json:"id"`
	BusinessID      string                   `json:"business_id,omitempty"`
	AggregationType AggregationType          `json:"aggregation_type"`
	Status          string                   `json:"status"` // pending, processing, completed, failed
	Progress        int                      `json:"progress"`
	CreatedAt       time.Time                `json:"created_at"`
	StartedAt       *time.Time               `json:"started_at,omitempty"`
	CompletedAt     *time.Time               `json:"completed_at,omitempty"`
	Result          *DataAggregationResponse `json:"result,omitempty"`
	Error           string                   `json:"error,omitempty"`
	Metadata        map[string]interface{}   `json:"metadata,omitempty"`
}

// AggregationSchema represents an aggregation schema
type AggregationSchema struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        AggregationType        `json:"type"`
	Rules       []DataAggregationRule  `json:"rules"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DataAggregationHandler handles data aggregation API endpoints
type DataAggregationHandler struct {
	logger          *zap.Logger
	metrics         *observability.Metrics
	aggregationJobs map[string]*AggregationJob
	jobMutex        sync.RWMutex
	jobCounter      int
	schemas         map[string]*AggregationSchema
	schemaMutex     sync.RWMutex
}

// NewDataAggregationHandler creates a new data aggregation handler
func NewDataAggregationHandler(
	logger *zap.Logger,
	metrics *observability.Metrics,
) *DataAggregationHandler {
	handler := &DataAggregationHandler{
		logger:          logger,
		metrics:         metrics,
		aggregationJobs: make(map[string]*AggregationJob),
		schemas:         make(map[string]*AggregationSchema),
		jobCounter:      0,
	}

	// Initialize default schemas
	handler.initializeDefaultSchemas()

	return handler
}

// AggregateDataHandler handles immediate data aggregation requests
func (h *DataAggregationHandler) AggregateDataHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse request
	var req DataAggregationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode aggregation request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateAggregationRequest(&req); err != nil {
		h.logger.Error("Invalid aggregation request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract business ID from header if not provided
	if req.BusinessID == "" {
		req.BusinessID = r.Header.Get("X-Business-ID")
	}

	// Aggregate data
	result, err := h.aggregateData(r.Context(), &req)
	if err != nil {
		h.logger.Error("Data aggregation failed", zap.Error(err))
		http.Error(w, fmt.Sprintf("Aggregation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Aggregation-ID", result.AggregationID)
	w.Header().Set("X-Processing-Time", result.ProcessingTime.String())

	// Return response
	json.NewEncoder(w).Encode(result)

	// Log metrics
	h.logger.Info("Data aggregation completed",
		zap.String("aggregation_id", result.AggregationID),
		zap.String("business_id", req.BusinessID),
		zap.String("type", string(req.AggregationType)),
		zap.Bool("success", result.IsSuccessful),
		zap.Duration("processing_time", time.Since(startTime)),
	)
}

// CreateAggregationJobHandler handles background aggregation job creation
func (h *DataAggregationHandler) CreateAggregationJobHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req DataAggregationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode aggregation job request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateAggregationRequest(&req); err != nil {
		h.logger.Error("Invalid aggregation job request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract business ID from header if not provided
	if req.BusinessID == "" {
		req.BusinessID = r.Header.Get("X-Business-ID")
	}

	// Create aggregation job
	job, err := h.createAggregationJob(&req)
	if err != nil {
		h.logger.Error("Failed to create aggregation job", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create job: %v", err), http.StatusInternalServerError)
		return
	}

	// Start background processing
	go h.processAggregationJob(job)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Job-ID", job.ID)

	// Return response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":     job.ID,
		"status":     job.Status,
		"created_at": job.CreatedAt,
		"message":    "Aggregation job created successfully",
	})

	h.logger.Info("Aggregation job created",
		zap.String("job_id", job.ID),
		zap.String("business_id", req.BusinessID),
		zap.String("type", string(req.AggregationType)),
	)
}

// GetAggregationJobHandler handles aggregation job status retrieval
func (h *DataAggregationHandler) GetAggregationJobHandler(w http.ResponseWriter, r *http.Request) {
	jobID := extractPathParam(r, "job_id")
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	h.jobMutex.RLock()
	job, exists := h.aggregationJobs[jobID]
	h.jobMutex.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// ListAggregationJobsHandler handles aggregation job listing
func (h *DataAggregationHandler) ListAggregationJobsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	businessID := query.Get("business_id")
	status := query.Get("status")
	aggregationType := query.Get("aggregation_type")
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Filter jobs
	h.jobMutex.RLock()
	var filteredJobs []*AggregationJob
	for _, job := range h.aggregationJobs {
		if businessID != "" && job.BusinessID != businessID {
			continue
		}
		if status != "" && job.Status != status {
			continue
		}
		if aggregationType != "" && string(job.AggregationType) != aggregationType {
			continue
		}
		filteredJobs = append(filteredJobs, job)
	}
	h.jobMutex.RUnlock()

	// Paginate results
	total := len(filteredJobs)
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	var paginatedJobs []*AggregationJob
	if start < total {
		paginatedJobs = filteredJobs[start:end]
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs": paginatedJobs,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetAggregationSchemaHandler handles aggregation schema retrieval
func (h *DataAggregationHandler) GetAggregationSchemaHandler(w http.ResponseWriter, r *http.Request) {
	schemaID := extractPathParam(r, "schema_id")
	if schemaID == "" {
		http.Error(w, "Schema ID is required", http.StatusBadRequest)
		return
	}

	h.schemaMutex.RLock()
	schema, exists := h.schemas[schemaID]
	h.schemaMutex.RUnlock()

	if !exists {
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// ListAggregationSchemasHandler handles aggregation schema listing
func (h *DataAggregationHandler) ListAggregationSchemasHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	aggregationType := query.Get("type")

	// Filter schemas
	h.schemaMutex.RLock()
	var filteredSchemas []*AggregationSchema
	for _, schema := range h.schemas {
		if aggregationType != "" && string(schema.Type) != aggregationType {
			continue
		}
		filteredSchemas = append(filteredSchemas, schema)
	}
	h.schemaMutex.RUnlock()

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schemas": filteredSchemas,
		"total":   len(filteredSchemas),
	})
}

// Helper functions

func (h *DataAggregationHandler) validateAggregationRequest(req *DataAggregationRequest) error {
	if req.AggregationType == "" {
		return fmt.Errorf("aggregation_type is required")
	}

	if !h.isValidAggregationType(req.AggregationType) {
		return fmt.Errorf("invalid aggregation_type: %s", req.AggregationType)
	}

	if req.Data == nil {
		return fmt.Errorf("data is required")
	}

	return nil
}

func (h *DataAggregationHandler) isValidAggregationType(aggType AggregationType) bool {
	validTypes := []AggregationType{
		AggregationTypeBusinessMetrics,
		AggregationTypeRiskAssessments,
		AggregationTypeComplianceReports,
		AggregationTypePerformanceAnalytics,
		AggregationTypeTrendAnalysis,
		AggregationTypeCustom,
		AggregationTypeAll,
	}

	for _, validType := range validTypes {
		if aggType == validType {
			return true
		}
	}
	return false
}

func (h *DataAggregationHandler) aggregateData(ctx context.Context, req *DataAggregationRequest) (*DataAggregationResponse, error) {
	startTime := time.Now()

	// Load schema if provided
	var rules []DataAggregationRule
	if req.SchemaID != "" {
		h.schemaMutex.RLock()
		schema, exists := h.schemas[req.SchemaID]
		h.schemaMutex.RUnlock()

		if exists {
			rules = schema.Rules
		}
	}

	// Use provided rules if no schema
	if len(rules) == 0 {
		rules = req.Rules
	}

	// Apply aggregation rules
	aggregatedData := req.Data
	appliedRules := []DataAggregationRule{}
	skippedRules := []DataAggregationRule{}
	failedRules := []DataAggregationRule{}

	for _, rule := range rules {
		if !rule.Enabled {
			skippedRules = append(skippedRules, rule)
			continue
		}

		// Check condition if provided
		if rule.Condition != "" {
			if !h.evaluateCondition(aggregatedData, rule.Condition) {
				skippedRules = append(skippedRules, rule)
				continue
			}
		}

		// Apply aggregation rule
		result, err := h.applyAggregationRule(aggregatedData, rule)
		if err != nil {
			failedRules = append(failedRules, rule)
			h.logger.Error("Failed to apply aggregation rule",
				zap.String("field", rule.Field),
				zap.String("operation", string(rule.Operation)),
				zap.Error(err),
			)
			continue
		}

		aggregatedData = result
		appliedRules = append(appliedRules, rule)
	}

	// Determine status
	status := "success"
	if len(failedRules) > 0 {
		if len(appliedRules) == 0 {
			status = "failed"
		} else {
			status = "partial"
		}
	}

	// Create summary
	summary := map[string]interface{}{
		"total_rules":     len(rules),
		"applied_rules":   len(appliedRules),
		"skipped_rules":   len(skippedRules),
		"failed_rules":    len(failedRules),
		"success_rate":    float64(len(appliedRules)) / float64(len(rules)),
		"data_count":      h.getDataCount(aggregatedData),
		"processing_time": time.Since(startTime).String(),
	}

	// Create response
	response := &DataAggregationResponse{
		AggregationID:   h.generateAggregationID(),
		BusinessID:      req.BusinessID,
		AggregationType: req.AggregationType,
		Status:          status,
		IsSuccessful:    status == "success",
		OriginalData:    req.Data,
		AggregatedData:  aggregatedData,
		AppliedRules:    appliedRules,
		SkippedRules:    skippedRules,
		FailedRules:     failedRules,
		Summary:         summary,
		TimeRange:       req.TimeRange,
		AggregatedAt:    time.Now(),
		ProcessingTime:  time.Since(startTime),
		Metadata:        req.Metadata,
	}

	return response, nil
}

func (h *DataAggregationHandler) createAggregationJob(req *DataAggregationRequest) (*AggregationJob, error) {
	h.jobMutex.Lock()
	defer h.jobMutex.Unlock()

	h.jobCounter++
	jobID := fmt.Sprintf("agg_job_%d_%d", time.Now().Unix(), h.jobCounter)

	job := &AggregationJob{
		ID:              jobID,
		BusinessID:      req.BusinessID,
		AggregationType: req.AggregationType,
		Status:          "pending",
		Progress:        0,
		CreatedAt:       time.Now(),
		Metadata:        req.Metadata,
	}

	h.aggregationJobs[jobID] = job
	return job, nil
}

func (h *DataAggregationHandler) processAggregationJob(job *AggregationJob) {
	h.jobMutex.Lock()
	job.Status = "processing"
	now := time.Now()
	job.StartedAt = &now
	h.jobMutex.Unlock()

	// Simulate processing time
	time.Sleep(2 * time.Second)

	// Update progress
	h.jobMutex.Lock()
	job.Progress = 50
	h.jobMutex.Unlock()

	// Simulate more processing
	time.Sleep(1 * time.Second)

	// Complete job
	h.jobMutex.Lock()
	job.Status = "completed"
	job.Progress = 100
	now = time.Now()
	job.CompletedAt = &now

	// Create mock result
	job.Result = &DataAggregationResponse{
		AggregationID:   h.generateAggregationID(),
		BusinessID:      job.BusinessID,
		AggregationType: job.AggregationType,
		Status:          "success",
		IsSuccessful:    true,
		AggregatedData: map[string]interface{}{
			"total_count":   100,
			"average_score": 0.85,
			"success_rate":  0.92,
		},
		Summary: map[string]interface{}{
			"total_rules":   5,
			"applied_rules": 5,
			"success_rate":  1.0,
		},
		AggregatedAt:   time.Now(),
		ProcessingTime: time.Since(*job.StartedAt),
		Metadata:       job.Metadata,
	}
	h.jobMutex.Unlock()

	h.logger.Info("Aggregation job completed",
		zap.String("job_id", job.ID),
		zap.String("business_id", job.BusinessID),
		zap.String("status", job.Status),
	)
}

func (h *DataAggregationHandler) applyAggregationRule(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation - in a real system, this would apply actual aggregation logic
	switch rule.Operation {
	case AggregationOperationCount:
		return h.applyCountAggregation(data, rule)
	case AggregationOperationSum:
		return h.applySumAggregation(data, rule)
	case AggregationOperationAverage:
		return h.applyAverageAggregation(data, rule)
	case AggregationOperationMin:
		return h.applyMinAggregation(data, rule)
	case AggregationOperationMax:
		return h.applyMaxAggregation(data, rule)
	case AggregationOperationMedian:
		return h.applyMedianAggregation(data, rule)
	case AggregationOperationPercentile:
		return h.applyPercentileAggregation(data, rule)
	case AggregationOperationGroupBy:
		return h.applyGroupByAggregation(data, rule)
	case AggregationOperationPivot:
		return h.applyPivotAggregation(data, rule)
	case AggregationOperationCustom:
		return h.applyCustomAggregation(data, rule)
	default:
		return nil, fmt.Errorf("unsupported aggregation operation: %s", rule.Operation)
	}
}

// Mock aggregation implementations
func (h *DataAggregationHandler) applyCountAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field":  rule.Field,
		"count":  100,
		"result": "count_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applySumAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field":  rule.Field,
		"sum":    1500.50,
		"result": "sum_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applyAverageAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field":   rule.Field,
		"average": 75.25,
		"result":  "average_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applyMinAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field":  rule.Field,
		"min":    10.0,
		"result": "min_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applyMaxAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field":  rule.Field,
		"max":    95.0,
		"result": "max_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applyMedianAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field":  rule.Field,
		"median": 78.5,
		"result": "median_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applyPercentileAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	percentile := 90.0
	if p, ok := rule.Parameters["percentile"]; ok {
		if pFloat, ok := p.(float64); ok {
			percentile = pFloat
		}
	}

	return map[string]interface{}{
		"field":      rule.Field,
		"percentile": percentile,
		"value":      88.0,
		"result":     "percentile_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applyGroupByAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field": rule.Field,
		"groups": map[string]interface{}{
			"group_1": 25,
			"group_2": 30,
			"group_3": 45,
		},
		"result": "group_by_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applyPivotAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field": rule.Field,
		"pivot": map[string]interface{}{
			"rows": []string{"row_1", "row_2", "row_3"},
			"cols": []string{"col_1", "col_2", "col_3"},
			"data": [][]interface{}{
				{10, 20, 30},
				{15, 25, 35},
				{20, 30, 40},
			},
		},
		"result": "pivot_aggregation",
	}, nil
}

func (h *DataAggregationHandler) applyCustomAggregation(data interface{}, rule DataAggregationRule) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"field":  rule.Field,
		"custom": "custom_aggregation_result",
		"result": "custom_aggregation",
	}, nil
}

func (h *DataAggregationHandler) evaluateCondition(data interface{}, condition string) bool {
	// Mock implementation - in a real system, this would evaluate the condition
	return true
}

func (h *DataAggregationHandler) getDataCount(data interface{}) int {
	// Mock implementation - in a real system, this would count the actual data
	return 100
}

func (h *DataAggregationHandler) generateAggregationID() string {
	return fmt.Sprintf("agg_%d_%d", time.Now().Unix(), h.jobCounter)
}

func (h *DataAggregationHandler) initializeDefaultSchemas() {
	// Business Metrics Schema
	businessMetricsSchema := &AggregationSchema{
		ID:          "business_metrics_default",
		Name:        "Default Business Metrics",
		Description: "Default schema for business metrics aggregation",
		Type:        AggregationTypeBusinessMetrics,
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rules: []DataAggregationRule{
			{
				Field:       "verification_count",
				Operation:   AggregationOperationCount,
				Parameters:  map[string]interface{}{},
				Description: "Count total verifications",
				Enabled:     true,
				Order:       1,
			},
			{
				Field:       "success_rate",
				Operation:   AggregationOperationAverage,
				Parameters:  map[string]interface{}{},
				Description: "Calculate average success rate",
				Enabled:     true,
				Order:       2,
			},
		},
	}

	// Risk Assessment Schema
	riskAssessmentSchema := &AggregationSchema{
		ID:          "risk_assessment_default",
		Name:        "Default Risk Assessment",
		Description: "Default schema for risk assessment aggregation",
		Type:        AggregationTypeRiskAssessments,
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rules: []DataAggregationRule{
			{
				Field:       "risk_score",
				Operation:   AggregationOperationAverage,
				Parameters:  map[string]interface{}{},
				Description: "Calculate average risk score",
				Enabled:     true,
				Order:       1,
			},
			{
				Field:       "high_risk_count",
				Operation:   AggregationOperationCount,
				Parameters:  map[string]interface{}{"threshold": 0.8},
				Description: "Count high-risk assessments",
				Enabled:     true,
				Order:       2,
			},
		},
	}

	h.schemaMutex.Lock()
	h.schemas[businessMetricsSchema.ID] = businessMetricsSchema
	h.schemas[riskAssessmentSchema.ID] = riskAssessmentSchema
	h.schemaMutex.Unlock()
}
