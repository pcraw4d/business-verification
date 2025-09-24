package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/observability"
)

// TransformationType represents the type of transformation to perform
type TransformationType string

const (
	TransformationTypeDataCleaning  TransformationType = "data_cleaning"
	TransformationTypeNormalization TransformationType = "normalization"
	TransformationTypeEnrichment    TransformationType = "enrichment"
	TransformationTypeAggregation   TransformationType = "aggregation"
	TransformationTypeFiltering     TransformationType = "filtering"
	TransformationTypeMapping       TransformationType = "mapping"
	TransformationTypeCustom        TransformationType = "custom"
	TransformationTypeAll           TransformationType = "all"
)

// TransformationOperation represents a specific transformation operation
type TransformationOperation string

const (
	TransformationOperationTrim      TransformationOperation = "trim"
	TransformationOperationToLower   TransformationOperation = "to_lower"
	TransformationOperationToUpper   TransformationOperation = "to_upper"
	TransformationOperationReplace   TransformationOperation = "replace"
	TransformationOperationExtract   TransformationOperation = "extract"
	TransformationOperationFormat    TransformationOperation = "format"
	TransformationOperationValidate  TransformationOperation = "validate"
	TransformationOperationEnrich    TransformationOperation = "enrich"
	TransformationOperationAggregate TransformationOperation = "aggregate"
	TransformationOperationFilter    TransformationOperation = "filter"
	TransformationOperationMap       TransformationOperation = "map"
	TransformationOperationCustom    TransformationOperation = "custom"
)

// DataTransformationRule represents a transformation rule
type DataTransformationRule struct {
	Field       string                  `json:"field"`
	Operation   TransformationOperation `json:"operation"`
	Parameters  map[string]interface{}  `json:"parameters"`
	Condition   string                  `json:"condition,omitempty"`
	Description string                  `json:"description"`
	Enabled     bool                    `json:"enabled"`
	Order       int                     `json:"order"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
}

// DataTransformationRequest represents a request to transform data
type DataTransformationRequest struct {
	BusinessID         string                   `json:"business_id,omitempty"`
	TransformationType TransformationType       `json:"transformation_type"`
	Data               interface{}              `json:"data"`
	Rules              []DataTransformationRule `json:"rules,omitempty"`
	SchemaID           string                   `json:"schema_id,omitempty"`
	ValidateBefore     bool                     `json:"validate_before"`
	ValidateAfter      bool                     `json:"validate_after"`
	IncludeMetadata    bool                     `json:"include_metadata"`
	Metadata           map[string]interface{}   `json:"metadata,omitempty"`
}

// DataTransformationResponse represents the response from a data transformation request
type DataTransformationResponse struct {
	TransformationID   string                   `json:"transformation_id"`
	BusinessID         string                   `json:"business_id,omitempty"`
	TransformationType TransformationType       `json:"transformation_type"`
	Status             string                   `json:"status"` // success, partial, failed
	IsSuccessful       bool                     `json:"is_successful"`
	OriginalData       interface{}              `json:"original_data,omitempty"`
	TransformedData    interface{}              `json:"transformed_data"`
	AppliedRules       []DataTransformationRule `json:"applied_rules"`
	SkippedRules       []DataTransformationRule `json:"skipped_rules,omitempty"`
	FailedRules        []DataTransformationRule `json:"failed_rules,omitempty"`
	ValidationBefore   interface{}              `json:"validation_before,omitempty"`
	ValidationAfter    interface{}              `json:"validation_after,omitempty"`
	Summary            map[string]interface{}   `json:"summary"`
	TransformedAt      time.Time                `json:"transformed_at"`
	ProcessingTime     time.Duration            `json:"processing_time"`
	Metadata           map[string]interface{}   `json:"metadata,omitempty"`
}

// TransformationJob represents a background transformation job
type TransformationJob struct {
	ID                 string                      `json:"id"`
	BusinessID         string                      `json:"business_id,omitempty"`
	TransformationType TransformationType          `json:"transformation_type"`
	Status             string                      `json:"status"` // pending, processing, completed, failed
	Progress           int                         `json:"progress"`
	CreatedAt          time.Time                   `json:"created_at"`
	StartedAt          *time.Time                  `json:"started_at,omitempty"`
	CompletedAt        *time.Time                  `json:"completed_at,omitempty"`
	Result             *DataTransformationResponse `json:"result,omitempty"`
	Error              string                      `json:"error,omitempty"`
	Metadata           map[string]interface{}      `json:"metadata,omitempty"`
}

// TransformationSchema represents a transformation schema
type TransformationSchema struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Type        TransformationType       `json:"type"`
	Rules       []DataTransformationRule `json:"rules"`
	Version     string                   `json:"version"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// DataTransformationHandler handles data transformation API endpoints
type DataTransformationHandler struct {
	logger             *zap.Logger
	metrics            *observability.Metrics
	transformationJobs map[string]*TransformationJob
	jobMutex           sync.RWMutex
	jobCounter         int
	schemas            map[string]*TransformationSchema
	schemaMutex        sync.RWMutex
}

// NewDataTransformationHandler creates a new data transformation handler
func NewDataTransformationHandler(
	logger *zap.Logger,
	metrics *observability.Metrics,
) *DataTransformationHandler {
	handler := &DataTransformationHandler{
		logger:             logger,
		metrics:            metrics,
		transformationJobs: make(map[string]*TransformationJob),
		schemas:            make(map[string]*TransformationSchema),
		jobCounter:         0,
	}

	// Initialize default schemas
	handler.initializeDefaultSchemas()

	return handler
}

// TransformDataHandler handles immediate data transformation requests
func (h *DataTransformationHandler) TransformDataHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse request
	var req DataTransformationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode transformation request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateTransformationRequest(&req); err != nil {
		h.logger.Error("Invalid transformation request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract business ID from header if not provided
	if req.BusinessID == "" {
		req.BusinessID = r.Header.Get("X-Business-ID")
	}

	// Transform data
	result, err := h.transformData(r.Context(), &req)
	if err != nil {
		h.logger.Error("Data transformation failed", zap.Error(err))
		http.Error(w, fmt.Sprintf("Transformation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Transformation-ID", result.TransformationID)
	w.Header().Set("X-Processing-Time", result.ProcessingTime.String())

	// Return response
	json.NewEncoder(w).Encode(result)

	// Log metrics
	h.logger.Info("Data transformation completed",
		zap.String("transformation_id", result.TransformationID),
		zap.String("business_id", req.BusinessID),
		zap.String("type", string(req.TransformationType)),
		zap.Bool("success", result.IsSuccessful),
		zap.Duration("processing_time", time.Since(startTime)),
	)
}

// CreateTransformationJobHandler handles background transformation job creation
func (h *DataTransformationHandler) CreateTransformationJobHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req DataTransformationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode transformation job request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateTransformationRequest(&req); err != nil {
		h.logger.Error("Invalid transformation job request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract business ID from header if not provided
	if req.BusinessID == "" {
		req.BusinessID = r.Header.Get("X-Business-ID")
	}

	// Create transformation job
	job, err := h.createTransformationJob(&req)
	if err != nil {
		h.logger.Error("Failed to create transformation job", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create job: %v", err), http.StatusInternalServerError)
		return
	}

	// Start background processing
	go h.processTransformationJob(job)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Job-ID", job.ID)

	// Return response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":     job.ID,
		"status":     job.Status,
		"created_at": job.CreatedAt,
		"message":    "Transformation job created successfully",
	})

	h.logger.Info("Transformation job created",
		zap.String("job_id", job.ID),
		zap.String("business_id", req.BusinessID),
		zap.String("type", string(req.TransformationType)),
	)
}

// GetTransformationJobHandler handles transformation job status retrieval
func (h *DataTransformationHandler) GetTransformationJobHandler(w http.ResponseWriter, r *http.Request) {
	jobID := extractPathParam(r, "job_id")
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	h.jobMutex.RLock()
	job, exists := h.transformationJobs[jobID]
	h.jobMutex.RUnlock()

	if !exists {
		http.Error(w, "Transformation job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// ListTransformationJobsHandler handles transformation job listing
func (h *DataTransformationHandler) ListTransformationJobsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	businessID := r.URL.Query().Get("business_id")
	status := r.URL.Query().Get("status")
	transformationType := r.URL.Query().Get("type")

	// Get jobs
	h.jobMutex.RLock()
	var jobs []*TransformationJob
	for _, job := range h.transformationJobs {
		// Apply filters
		if businessID != "" && job.BusinessID != businessID {
			continue
		}
		if status != "" && job.Status != status {
			continue
		}
		if transformationType != "" && string(job.TransformationType) != transformationType {
			continue
		}
		jobs = append(jobs, job)
	}
	h.jobMutex.RUnlock()

	// Calculate pagination
	total := len(jobs)
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	// Apply pagination
	paginatedJobs := jobs[start:end]

	response := map[string]interface{}{
		"jobs": paginatedJobs,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTransformationSchemaHandler handles transformation schema retrieval
func (h *DataTransformationHandler) GetTransformationSchemaHandler(w http.ResponseWriter, r *http.Request) {
	schemaID := extractPathParam(r, "schema_id")
	if schemaID == "" {
		http.Error(w, "Schema ID is required", http.StatusBadRequest)
		return
	}

	h.schemaMutex.RLock()
	schema, exists := h.schemas[schemaID]
	h.schemaMutex.RUnlock()

	if !exists {
		http.Error(w, "Transformation schema not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// ListTransformationSchemasHandler handles transformation schema listing
func (h *DataTransformationHandler) ListTransformationSchemasHandler(w http.ResponseWriter, r *http.Request) {
	transformationType := r.URL.Query().Get("type")

	h.schemaMutex.RLock()
	var schemas []*TransformationSchema
	for _, schema := range h.schemas {
		if transformationType != "" && string(schema.Type) != transformationType {
			continue
		}
		schemas = append(schemas, schema)
	}
	h.schemaMutex.RUnlock()

	response := map[string]interface{}{
		"schemas": schemas,
		"total":   len(schemas),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func (h *DataTransformationHandler) validateTransformationRequest(req *DataTransformationRequest) error {
	if req.TransformationType == "" {
		return fmt.Errorf("transformation_type is required")
	}

	if !h.isValidTransformationType(req.TransformationType) {
		return fmt.Errorf("invalid transformation_type: %s", req.TransformationType)
	}

	if req.Data == nil {
		return fmt.Errorf("data is required")
	}

	// Validate schema if provided
	if req.SchemaID != "" {
		h.schemaMutex.RLock()
		_, exists := h.schemas[req.SchemaID]
		h.schemaMutex.RUnlock()
		if !exists {
			return fmt.Errorf("schema not found: %s", req.SchemaID)
		}
	}

	return nil
}

func (h *DataTransformationHandler) isValidTransformationType(transformationType TransformationType) bool {
	validTypes := []TransformationType{
		TransformationTypeDataCleaning,
		TransformationTypeNormalization,
		TransformationTypeEnrichment,
		TransformationTypeAggregation,
		TransformationTypeFiltering,
		TransformationTypeMapping,
		TransformationTypeCustom,
		TransformationTypeAll,
	}

	for _, validType := range validTypes {
		if transformationType == validType {
			return true
		}
	}
	return false
}

func (h *DataTransformationHandler) transformData(ctx context.Context, req *DataTransformationRequest) (*DataTransformationResponse, error) {
	startTime := time.Now()

	// Get transformation rules
	rules := req.Rules
	if req.SchemaID != "" {
		h.schemaMutex.RLock()
		schema, exists := h.schemas[req.SchemaID]
		h.schemaMutex.RUnlock()
		if exists {
			rules = schema.Rules
		}
	}

	// Perform pre-transformation validation if requested
	var validationBefore interface{}
	if req.ValidateBefore {
		validationBefore = h.performPreTransformationValidation(req.Data)
	}

	// Apply transformation rules
	transformedData := req.Data
	var appliedRules, skippedRules, failedRules []DataTransformationRule

	for _, rule := range rules {
		if !rule.Enabled {
			skippedRules = append(skippedRules, rule)
			continue
		}

		// Check condition if specified
		if rule.Condition != "" && !h.evaluateCondition(transformedData, rule.Condition) {
			skippedRules = append(skippedRules, rule)
			continue
		}

		// Apply transformation
		transformed, err := h.applyTransformationRule(transformedData, rule)
		if err != nil {
			h.logger.Error("Transformation rule failed",
				zap.String("field", rule.Field),
				zap.String("operation", string(rule.Operation)),
				zap.Error(err),
			)
			failedRules = append(failedRules, rule)
			continue
		}

		transformedData = transformed
		appliedRules = append(appliedRules, rule)
	}

	// Perform post-transformation validation if requested
	var validationAfter interface{}
	if req.ValidateAfter {
		validationAfter = h.performPostTransformationValidation(transformedData)
	}

	// Calculate processing time
	processingTime := time.Since(startTime)

	// Determine status
	status := "success"
	if len(failedRules) > 0 {
		if len(appliedRules) > 0 {
			status = "partial"
		} else {
			status = "failed"
		}
	}

	// Create summary
	summary := map[string]interface{}{
		"total_rules":     len(rules),
		"applied_rules":   len(appliedRules),
		"skipped_rules":   len(skippedRules),
		"failed_rules":    len(failedRules),
		"success_rate":    float64(len(appliedRules)) / float64(len(rules)),
		"processing_time": processingTime.String(),
	}

	// Create response
	response := &DataTransformationResponse{
		TransformationID:   h.generateTransformationID(),
		BusinessID:         req.BusinessID,
		TransformationType: req.TransformationType,
		Status:             status,
		IsSuccessful:       status == "success",
		OriginalData:       req.Data,
		TransformedData:    transformedData,
		AppliedRules:       appliedRules,
		SkippedRules:       skippedRules,
		FailedRules:        failedRules,
		ValidationBefore:   validationBefore,
		ValidationAfter:    validationAfter,
		Summary:            summary,
		TransformedAt:      time.Now(),
		ProcessingTime:     processingTime,
		Metadata:           req.Metadata,
	}

	return response, nil
}

func (h *DataTransformationHandler) createTransformationJob(req *DataTransformationRequest) (*TransformationJob, error) {
	h.jobMutex.Lock()
	defer h.jobMutex.Unlock()

	h.jobCounter++
	jobID := fmt.Sprintf("transform_%d_%d", time.Now().Unix(), h.jobCounter)

	job := &TransformationJob{
		ID:                 jobID,
		BusinessID:         req.BusinessID,
		TransformationType: req.TransformationType,
		Status:             "pending",
		Progress:           0,
		CreatedAt:          time.Now(),
		Metadata:           req.Metadata,
	}

	h.transformationJobs[jobID] = job
	return job, nil
}

func (h *DataTransformationHandler) processTransformationJob(job *TransformationJob) {
	startTime := time.Now()

	// Update job status
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
	job.Result = &DataTransformationResponse{
		TransformationID:   job.ID,
		BusinessID:         job.BusinessID,
		TransformationType: job.TransformationType,
		Status:             "success",
		IsSuccessful:       true,
		TransformedData:    map[string]interface{}{"transformed": true},
		AppliedRules:       []DataTransformationRule{},
		Summary: map[string]interface{}{
			"total_rules":   0,
			"applied_rules": 0,
			"success_rate":  1.0,
		},
		TransformedAt:  time.Now(),
		ProcessingTime: time.Since(startTime),
		Metadata:       job.Metadata,
	}
	h.jobMutex.Unlock()

	h.logger.Info("Transformation job completed",
		zap.String("job_id", job.ID),
		zap.Duration("processing_time", time.Since(startTime)),
	)
}

func (h *DataTransformationHandler) applyTransformationRule(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// This is a simplified implementation
	// In a real implementation, you would have more sophisticated transformation logic

	switch rule.Operation {
	case TransformationOperationTrim:
		return h.applyTrimTransformation(data, rule)
	case TransformationOperationToLower:
		return h.applyToLowerTransformation(data, rule)
	case TransformationOperationToUpper:
		return h.applyToUpperTransformation(data, rule)
	case TransformationOperationReplace:
		return h.applyReplaceTransformation(data, rule)
	case TransformationOperationExtract:
		return h.applyExtractTransformation(data, rule)
	case TransformationOperationFormat:
		return h.applyFormatTransformation(data, rule)
	case TransformationOperationValidate:
		return h.applyValidateTransformation(data, rule)
	case TransformationOperationEnrich:
		return h.applyEnrichTransformation(data, rule)
	case TransformationOperationAggregate:
		return h.applyAggregateTransformation(data, rule)
	case TransformationOperationFilter:
		return h.applyFilterTransformation(data, rule)
	case TransformationOperationMap:
		return h.applyMapTransformation(data, rule)
	case TransformationOperationCustom:
		return h.applyCustomTransformation(data, rule)
	default:
		return data, fmt.Errorf("unsupported transformation operation: %s", rule.Operation)
	}
}

func (h *DataTransformationHandler) applyTrimTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyToLowerTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyToUpperTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyReplaceTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyExtractTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyFormatTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyValidateTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyEnrichTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyAggregateTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyFilterTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyMapTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) applyCustomTransformation(data interface{}, rule DataTransformationRule) (interface{}, error) {
	// Mock implementation
	return data, nil
}

func (h *DataTransformationHandler) evaluateCondition(data interface{}, condition string) bool {
	// Mock implementation - always return true
	return true
}

func (h *DataTransformationHandler) performPreTransformationValidation(data interface{}) interface{} {
	// Mock implementation
	return map[string]interface{}{
		"valid":  true,
		"issues": []interface{}{},
	}
}

func (h *DataTransformationHandler) performPostTransformationValidation(data interface{}) interface{} {
	// Mock implementation
	return map[string]interface{}{
		"valid":  true,
		"issues": []interface{}{},
	}
}

func (h *DataTransformationHandler) generateTransformationID() string {
	return fmt.Sprintf("transform_%d", time.Now().UnixNano())
}

func (h *DataTransformationHandler) initializeDefaultSchemas() {
	// Data cleaning schema
	dataCleaningSchema := &TransformationSchema{
		ID:          "data_cleaning_default",
		Name:        "Default Data Cleaning",
		Description: "Standard data cleaning transformations",
		Type:        TransformationTypeDataCleaning,
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rules: []DataTransformationRule{
			{
				Field:       "business_name",
				Operation:   TransformationOperationTrim,
				Parameters:  map[string]interface{}{},
				Description: "Trim whitespace from business name",
				Enabled:     true,
				Order:       1,
			},
			{
				Field:       "email",
				Operation:   TransformationOperationToLower,
				Parameters:  map[string]interface{}{},
				Description: "Convert email to lowercase",
				Enabled:     true,
				Order:       2,
			},
		},
	}

	// Normalization schema
	normalizationSchema := &TransformationSchema{
		ID:          "normalization_default",
		Name:        "Default Normalization",
		Description: "Standard data normalization transformations",
		Type:        TransformationTypeNormalization,
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rules: []DataTransformationRule{
			{
				Field:       "phone",
				Operation:   TransformationOperationFormat,
				Parameters:  map[string]interface{}{"format": "E.164"},
				Description: "Format phone number to E.164",
				Enabled:     true,
				Order:       1,
			},
		},
	}

	h.schemaMutex.Lock()
	h.schemas[dataCleaningSchema.ID] = dataCleaningSchema
	h.schemas[normalizationSchema.ID] = normalizationSchema
	h.schemaMutex.Unlock()
}

func extractPathParam(r *http.Request, param string) string {
	// Extract parameter from URL path
	// This is a simplified implementation
	path := r.URL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == param && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
