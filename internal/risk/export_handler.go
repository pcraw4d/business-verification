package risk

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// ExportHandler handles HTTP requests for risk data export
type ExportHandler struct {
	logger     *zap.Logger
	exportSvc  *ExportService
	jobManager *ExportJobManager
}

// NewExportHandler creates a new export handler
func NewExportHandler(logger *zap.Logger, exportSvc *ExportService, jobManager *ExportJobManager) *ExportHandler {
	return &ExportHandler{
		logger:     logger,
		exportSvc:  exportSvc,
		jobManager: jobManager,
	}
}

// CreateExportJobRequest represents the request to create an export job
type CreateExportJobRequest struct {
	BusinessID string                 `json:"business_id" validate:"required"`
	ExportType ExportType             `json:"export_type" validate:"required"`
	Format     ExportFormat           `json:"format" validate:"required"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CreateExportJobResponse represents the response for creating an export job
type CreateExportJobResponse struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// GetExportJobResponse represents the response for getting an export job
type GetExportJobResponse struct {
	Job *ExportJob `json:"job"`
}

// ListExportJobsResponse represents the response for listing export jobs
type ListExportJobsResponse struct {
	Jobs  []*ExportJob `json:"jobs"`
	Total int          `json:"total"`
}

// GetExportJobStatisticsResponse represents the response for getting job statistics
type GetExportJobStatisticsResponse struct {
	Statistics map[string]interface{} `json:"statistics"`
}

// CreateExportJob handles POST /api/v1/export/jobs
func (eh *ExportHandler) CreateExportJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Parse request body
	var req CreateExportJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessID == "" {
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	if req.ExportType == "" {
		http.Error(w, "export_type is required", http.StatusBadRequest)
		return
	}

	if req.Format == "" {
		http.Error(w, "format is required", http.StatusBadRequest)
		return
	}

	// Create export request
	exportReq := &ExportRequest{
		BusinessID: req.BusinessID,
		ExportType: req.ExportType,
		Format:     req.Format,
		Metadata:   req.Metadata,
	}

	// Create export job
	job, err := eh.jobManager.CreateExportJob(ctx, exportReq)
	if err != nil {
		eh.logger.Error("Failed to create export job",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", req.BusinessID),
			zap.Error(err))
		http.Error(w, "Failed to create export job", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := CreateExportJobResponse{
		JobID:     job.ID,
		Status:    job.Status,
		CreatedAt: job.CreatedAt,
		Message:   "Export job created successfully",
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", job.ID),
			zap.Error(err))
	}
}

// GetExportJob handles GET /api/v1/export/jobs/{job_id}
func (eh *ExportHandler) GetExportJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract job ID from URL path
	jobID := r.URL.Path[len("/api/v1/export/jobs/"):]
	if jobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	// Get export job
	job, err := eh.jobManager.GetExportJob(jobID)
	if err != nil {
		eh.logger.Error("Failed to get export job",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", jobID),
			zap.Error(err))
		http.Error(w, "Export job not found", http.StatusNotFound)
		return
	}

	// Prepare response
	response := GetExportJobResponse{
		Job: job,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", jobID),
			zap.Error(err))
	}
}

// ListExportJobs handles GET /api/v1/export/jobs
func (eh *ExportHandler) ListExportJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract query parameters
	businessID := r.URL.Query().Get("business_id")
	if businessID == "" {
		http.Error(w, "business_id query parameter is required", http.StatusBadRequest)
		return
	}

	// Get export jobs
	jobs, err := eh.jobManager.ListExportJobs(businessID)
	if err != nil {
		eh.logger.Error("Failed to list export jobs",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", businessID),
			zap.Error(err))
		http.Error(w, "Failed to list export jobs", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := ListExportJobsResponse{
		Jobs:  jobs,
		Total: len(jobs),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", businessID),
			zap.Error(err))
	}
}

// CancelExportJob handles DELETE /api/v1/export/jobs/{job_id}
func (eh *ExportHandler) CancelExportJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract job ID from URL path
	jobID := r.URL.Path[len("/api/v1/export/jobs/"):]
	if jobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	// Cancel export job
	err := eh.jobManager.CancelExportJob(jobID)
	if err != nil {
		eh.logger.Error("Failed to cancel export job",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", jobID),
			zap.Error(err))
		http.Error(w, "Failed to cancel export job", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send success response
	response := map[string]interface{}{
		"message": "Export job cancelled successfully",
		"job_id":  jobID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", jobID),
			zap.Error(err))
	}
}

// GetExportJobStatistics handles GET /api/v1/export/jobs/statistics
func (eh *ExportHandler) GetExportJobStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Get job statistics
	stats := eh.jobManager.GetJobStatistics()

	// Prepare response
	response := GetExportJobStatisticsResponse{
		Statistics: stats,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
	}
}

// CleanupOldJobs handles POST /api/v1/export/jobs/cleanup
func (eh *ExportHandler) CleanupOldJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract query parameters
	hoursStr := r.URL.Query().Get("hours")
	if hoursStr == "" {
		hoursStr = "24" // Default to 24 hours
	}

	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 0 {
		http.Error(w, "Invalid hours parameter", http.StatusBadRequest)
		return
	}

	// Calculate cutoff time
	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	// Cleanup old jobs
	err = eh.jobManager.CleanupOldJobs(cutoffTime)
	if err != nil {
		eh.logger.Error("Failed to cleanup old jobs",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Failed to cleanup old jobs", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send success response
	response := map[string]interface{}{
		"message":     "Old jobs cleaned up successfully",
		"cutoff_time": cutoffTime,
		"hours":       hours,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
	}
}

// ExportData handles POST /api/v1/export/data
func (eh *ExportHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Parse request body
	var req CreateExportJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessID == "" {
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	if req.ExportType == "" {
		http.Error(w, "export_type is required", http.StatusBadRequest)
		return
	}

	if req.Format == "" {
		http.Error(w, "format is required", http.StatusBadRequest)
		return
	}

	// Create export request
	exportReq := &ExportRequest{
		BusinessID: req.BusinessID,
		ExportType: req.ExportType,
		Format:     req.Format,
		Metadata:   req.Metadata,
	}

	// Perform immediate export (synchronous)
	var response *ExportResponse
	var err error

	switch req.ExportType {
	case ExportTypeAssessments:
		// In a real implementation, this would query the database
		assessments := []*RiskAssessment{
			{
				ID:           "mock-assessment-1",
				BusinessID:   req.BusinessID,
				BusinessName: "Mock Business",
				OverallScore: 75.5,
				OverallLevel: RiskLevelHigh,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
				AlertLevel:   RiskLevelMedium,
			},
		}
		response, err = eh.exportSvc.ExportRiskAssessments(ctx, assessments, req.Format)
	case ExportTypeFactors:
		// In a real implementation, this would query the database
		factors := []RiskScore{
			{
				FactorID:     "mock-factor-1",
				FactorName:   "Mock Financial Risk",
				Category:     RiskCategoryFinancial,
				Score:        80.0,
				Level:        RiskLevelHigh,
				Confidence:   0.9,
				Explanation:  "Mock financial risk explanation",
				CalculatedAt: time.Now(),
			},
		}
		response, err = eh.exportSvc.ExportRiskFactors(ctx, factors, req.Format)
	case ExportTypeTrends:
		// In a real implementation, this would query the database
		trends := []RiskTrend{
			{
				BusinessID:   req.BusinessID,
				Category:     RiskCategoryFinancial,
				Score:        75.0,
				Level:        RiskLevelHigh,
				RecordedAt:   time.Now(),
				ChangeFrom:   5.0,
				ChangePeriod: "1 month",
			},
		}
		response, err = eh.exportSvc.ExportRiskTrends(ctx, trends, req.Format)
	case ExportTypeAlerts:
		// In a real implementation, this would query the database
		alerts := []RiskAlert{
			{
				ID:             "mock-alert-1",
				BusinessID:     req.BusinessID,
				RiskFactor:     "mock-risk-factor",
				Level:          RiskLevelHigh,
				Message:        "Mock alert message",
				Score:          85.0,
				Threshold:      80.0,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
				AcknowledgedAt: nil,
			},
		}
		response, err = eh.exportSvc.ExportRiskAlerts(ctx, alerts, req.Format)
	default:
		http.Error(w, "Unsupported export type", http.StatusBadRequest)
		return
	}

	if err != nil {
		eh.logger.Error("Failed to export data",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", req.BusinessID),
			zap.String("export_type", string(req.ExportType)),
			zap.Error(err))
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", req.BusinessID),
			zap.Error(err))
	}
}

// RegisterRoutes registers the export handler routes
func (eh *ExportHandler) RegisterRoutes(mux *http.ServeMux) {
	// Export job management routes
	mux.HandleFunc("POST /api/v1/export/jobs", eh.CreateExportJob)
	mux.HandleFunc("GET /api/v1/export/jobs/{job_id}", eh.GetExportJob)
	mux.HandleFunc("GET /api/v1/export/jobs", eh.ListExportJobs)
	mux.HandleFunc("DELETE /api/v1/export/jobs/{job_id}", eh.CancelExportJob)
	mux.HandleFunc("GET /api/v1/export/jobs/statistics", eh.GetExportJobStatistics)
	mux.HandleFunc("POST /api/v1/export/jobs/cleanup", eh.CleanupOldJobs)

	// Direct export routes
	mux.HandleFunc("POST /api/v1/export/data", eh.ExportData)
}
