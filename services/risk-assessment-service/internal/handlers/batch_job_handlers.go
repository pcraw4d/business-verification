package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/batch"
)

// BatchJobHandler handles batch job API requests
type BatchJobHandler struct {
	jobManager JobManager
	logger     *zap.Logger
}

// JobManager interface for batch job management
type JobManager interface {
	SubmitBatchJob(ctx context.Context, request *batch.BatchJobRequest) (*batch.BatchJobResponse, error)
	GetBatchJobStatus(ctx context.Context, tenantID, jobID string) (*batch.BatchJobStatus, error)
	GetBatchJobResults(ctx context.Context, tenantID, jobID string) (*batch.BatchJobResults, error)
	CancelBatchJob(ctx context.Context, tenantID, jobID string) error
	ResumeBatchJob(ctx context.Context, tenantID, jobID string) error
	ListBatchJobs(ctx context.Context, filter *batch.BatchJobFilter) ([]*batch.BatchJob, error)
	GetBatchJobMetrics(ctx context.Context, tenantID string, startDate, endDate time.Time) (*batch.BatchJobMetrics, error)
}

// NewBatchJobHandler creates a new batch job handler
func NewBatchJobHandler(jobManager JobManager, logger *zap.Logger) *BatchJobHandler {
	return &BatchJobHandler{
		jobManager: jobManager,
		logger:     logger,
	}
}

// HandleSubmitBatchJob handles POST /api/v1/assess/batch/async
func (h *BatchJobHandler) HandleSubmitBatchJob(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling submit batch job request")

	// Parse request body
	var request batch.BatchJobRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateBatchJobRequest(&request); err != nil {
		h.logger.Error("Invalid batch job request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID from context (this would be implemented based on your auth system)
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Set tenant ID in request context
	ctx := r.Context()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)

	// Submit batch job
	response, err := h.jobManager.SubmitBatchJob(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to submit batch job", zap.Error(err))
		http.Error(w, "Failed to submit batch job", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetBatchJobStatus handles GET /api/v1/assess/batch/{job_id}
func (h *BatchJobHandler) HandleGetBatchJobStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]

	h.logger.Info("Handling get batch job status request",
		zap.String("job_id", jobID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get batch job status
	status, err := h.jobManager.GetBatchJobStatus(r.Context(), tenantID, jobID)
	if err != nil {
		h.logger.Error("Failed to get batch job status", zap.Error(err))
		http.Error(w, "Failed to get batch job status", http.StatusInternalServerError)
		return
	}

	if status == nil {
		http.Error(w, "Batch job not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetBatchJobResults handles GET /api/v1/assess/batch/{job_id}/results
func (h *BatchJobHandler) HandleGetBatchJobResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]

	h.logger.Info("Handling get batch job results request",
		zap.String("job_id", jobID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get batch job results
	results, err := h.jobManager.GetBatchJobResults(r.Context(), tenantID, jobID)
	if err != nil {
		h.logger.Error("Failed to get batch job results", zap.Error(err))
		http.Error(w, "Failed to get batch job results", http.StatusInternalServerError)
		return
	}

	if results == nil {
		http.Error(w, "Batch job not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleCancelBatchJob handles DELETE /api/v1/assess/batch/{job_id}
func (h *BatchJobHandler) HandleCancelBatchJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]

	h.logger.Info("Handling cancel batch job request",
		zap.String("job_id", jobID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Cancel batch job
	if err := h.jobManager.CancelBatchJob(r.Context(), tenantID, jobID); err != nil {
		h.logger.Error("Failed to cancel batch job", zap.Error(err))
		http.Error(w, "Failed to cancel batch job", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// HandleResumeBatchJob handles POST /api/v1/assess/batch/{job_id}/resume
func (h *BatchJobHandler) HandleResumeBatchJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]

	h.logger.Info("Handling resume batch job request",
		zap.String("job_id", jobID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Resume batch job
	if err := h.jobManager.ResumeBatchJob(r.Context(), tenantID, jobID); err != nil {
		h.logger.Error("Failed to resume batch job", zap.Error(err))
		http.Error(w, "Failed to resume batch job", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "Batch job resumed successfully",
		"job_id":  jobID,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleListBatchJobs handles GET /api/v1/assess/batch
func (h *BatchJobHandler) HandleListBatchJobs(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list batch jobs request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseBatchJobFilter(r, tenantID)

	// List batch jobs
	jobs, err := h.jobManager.ListBatchJobs(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list batch jobs", zap.Error(err))
		http.Error(w, "Failed to list batch jobs", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetBatchJobMetrics handles GET /api/v1/assess/batch/metrics
func (h *BatchJobHandler) HandleGetBatchJobMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get batch job metrics request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse date range from query parameters
	startDate, endDate, err := h.parseDateRange(r)
	if err != nil {
		h.logger.Error("Invalid date range", zap.Error(err))
		http.Error(w, "Invalid date range", http.StatusBadRequest)
		return
	}

	// Get batch job metrics
	metrics, err := h.jobManager.GetBatchJobMetrics(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get batch job metrics", zap.Error(err))
		http.Error(w, "Failed to get batch job metrics", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// Helper functions

// validateBatchJobRequest validates a batch job request
func validateBatchJobRequest(request *batch.BatchJobRequest) error {
	if request.JobType == "" {
		return fmt.Errorf("job_type is required")
	}

	if len(request.Requests) == 0 {
		return fmt.Errorf("requests cannot be empty")
	}

	if len(request.Requests) > 10000 {
		return fmt.Errorf("too many requests: %d exceeds maximum 10000", len(request.Requests))
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate job type
	validJobTypes := map[string]bool{
		"risk_assessment":   true,
		"compliance_check":  true,
		"custom_model_test": true,
	}

	if !validJobTypes[request.JobType] {
		return fmt.Errorf("invalid job_type: %s", request.JobType)
	}

	return nil
}

// extractTenantID extracts tenant ID from request context
func (h *BatchJobHandler) extractTenantID(r *http.Request) string {
	// This would be implemented based on your authentication/authorization system
	// For now, return a default tenant ID
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	return "default"
}

// parseBatchJobFilter parses query parameters into a batch job filter
func (h *BatchJobHandler) parseBatchJobFilter(r *http.Request, tenantID string) *batch.BatchJobFilter {
	filter := &batch.BatchJobFilter{
		TenantID: tenantID,
	}

	// Parse status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = batch.JobStatus(status)
	}

	// Parse job type filter
	if jobType := r.URL.Query().Get("job_type"); jobType != "" {
		filter.JobType = jobType
	}

	// Parse created by filter
	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		filter.CreatedBy = createdBy
	}

	// Parse date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}

// parseDateRange parses start and end dates from query parameters
func (h *BatchJobHandler) parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	// Default to last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// Parse start date
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		} else {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format: %s", startDateStr)
		}
	}

	// Parse end date
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		} else {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format: %s", endDateStr)
		}
	}

	// Validate date range
	if startDate.After(endDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date cannot be after end_date")
	}

	return startDate, endDate, nil
}
