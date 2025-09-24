package risk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// BackupHandler handles HTTP requests for risk data backup and restore
type BackupHandler struct {
	logger     *zap.Logger
	backupSvc  *BackupService
	jobManager *BackupJobManager
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(logger *zap.Logger, backupSvc *BackupService, jobManager *BackupJobManager) *BackupHandler {
	return &BackupHandler{
		logger:     logger,
		backupSvc:  backupSvc,
		jobManager: jobManager,
	}
}

// CreateBackupRequest represents the request to create a backup
type CreateBackupRequest struct {
	BusinessID    string                 `json:"business_id,omitempty"`
	BackupType    BackupType             `json:"backup_type" validate:"required"`
	IncludeData   []BackupDataType       `json:"include_data" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Compression   bool                   `json:"compression,omitempty"`
	RetentionDays int                    `json:"retention_days,omitempty"`
}

// CreateBackupResponse represents the response for creating a backup
type CreateBackupResponse struct {
	BackupID  string    `json:"backup_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// CreateBackupJobRequest represents the request to create a backup job
type CreateBackupJobRequest struct {
	BusinessID    string                 `json:"business_id,omitempty"`
	BackupType    BackupType             `json:"backup_type" validate:"required"`
	IncludeData   []BackupDataType       `json:"include_data" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Compression   bool                   `json:"compression,omitempty"`
	RetentionDays int                    `json:"retention_days,omitempty"`
}

// CreateBackupJobResponse represents the response for creating a backup job
type CreateBackupJobResponse struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// GetBackupJobResponse represents the response for getting a backup job
type GetBackupJobResponse struct {
	Job *BackupJob `json:"job"`
}

// ListBackupJobsResponse represents the response for listing backup jobs
type ListBackupJobsResponse struct {
	Jobs  []*BackupJob `json:"jobs"`
	Total int          `json:"total"`
}

// ListBackupsResponse represents the response for listing backups
type ListBackupsResponse struct {
	Backups []*BackupInfo `json:"backups"`
	Total   int           `json:"total"`
}

// GetBackupStatisticsResponse represents the response for getting backup statistics
type GetBackupStatisticsResponse struct {
	Statistics map[string]interface{} `json:"statistics"`
}

// RestoreBackupRequest represents the request to restore a backup
type RestoreBackupRequest struct {
	BackupID    string                 `json:"backup_id" validate:"required"`
	BusinessID  string                 `json:"business_id,omitempty"`
	RestoreType RestoreType            `json:"restore_type" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RestoreBackupResponse represents the response for restoring a backup
type RestoreBackupResponse struct {
	RestoreID   string    `json:"restore_id"`
	BackupID    string    `json:"backup_id"`
	Status      string    `json:"status"`
	RecordCount int       `json:"record_count"`
	StartedAt   time.Time `json:"started_at"`
	Message     string    `json:"message"`
}

// CreateBackupScheduleRequest represents the request to create a backup schedule
type CreateBackupScheduleRequest struct {
	BusinessID    string                 `json:"business_id,omitempty"`
	Name          string                 `json:"name" validate:"required"`
	Description   string                 `json:"description,omitempty"`
	BackupType    BackupType             `json:"backup_type" validate:"required"`
	IncludeData   []BackupDataType       `json:"include_data" validate:"required"`
	Schedule      string                 `json:"schedule" validate:"required"`
	RetentionDays int                    `json:"retention_days,omitempty"`
	Enabled       bool                   `json:"enabled,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateBackupScheduleResponse represents the response for creating a backup schedule
type CreateBackupScheduleResponse struct {
	ScheduleID string    `json:"schedule_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	Message    string    `json:"message"`
}

// ListBackupSchedulesResponse represents the response for listing backup schedules
type ListBackupSchedulesResponse struct {
	Schedules []*BackupSchedule `json:"schedules"`
	Total     int               `json:"total"`
}

// CreateBackup handles POST /api/v1/backup
func (bh *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Parse request body
	var req CreateBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bh.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BackupType == "" {
		http.Error(w, "backup_type is required", http.StatusBadRequest)
		return
	}

	if len(req.IncludeData) == 0 {
		http.Error(w, "include_data is required", http.StatusBadRequest)
		return
	}

	// Create backup request
	backupReq := &BackupRequest{
		BusinessID:    req.BusinessID,
		BackupType:    req.BackupType,
		IncludeData:   req.IncludeData,
		Metadata:      req.Metadata,
		Compression:   req.Compression,
		RetentionDays: req.RetentionDays,
	}

	// Create backup
	response, err := bh.backupSvc.CreateBackup(ctx, backupReq)
	if err != nil {
		bh.logger.Error("Failed to create backup",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", req.BusinessID),
			zap.Error(err))
		http.Error(w, "Failed to create backup", http.StatusInternalServerError)
		return
	}

	// Prepare response
	createResponse := CreateBackupResponse{
		BackupID:  response.BackupID,
		Status:    string(response.Status),
		CreatedAt: response.CreatedAt,
		Message:   "Backup created successfully",
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(createResponse); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("backup_id", response.BackupID),
			zap.Error(err))
	}
}

// CreateBackupJob handles POST /api/v1/backup/jobs
func (bh *BackupHandler) CreateBackupJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Parse request body
	var req CreateBackupJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bh.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BackupType == "" {
		http.Error(w, "backup_type is required", http.StatusBadRequest)
		return
	}

	if len(req.IncludeData) == 0 {
		http.Error(w, "include_data is required", http.StatusBadRequest)
		return
	}

	// Create backup request
	backupReq := &BackupRequest{
		BusinessID:    req.BusinessID,
		BackupType:    req.BackupType,
		IncludeData:   req.IncludeData,
		Metadata:      req.Metadata,
		Compression:   req.Compression,
		RetentionDays: req.RetentionDays,
	}

	// Create backup job
	job, err := bh.jobManager.CreateBackupJob(ctx, backupReq)
	if err != nil {
		bh.logger.Error("Failed to create backup job",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", req.BusinessID),
			zap.Error(err))
		http.Error(w, "Failed to create backup job", http.StatusInternalServerError)
		return
	}

	// Prepare response
	createResponse := CreateBackupJobResponse{
		JobID:     job.ID,
		Status:    string(job.Status),
		CreatedAt: job.CreatedAt,
		Message:   "Backup job created successfully",
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(createResponse); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", job.ID),
			zap.Error(err))
	}
}

// GetBackupJob handles GET /api/v1/backup/jobs/{job_id}
func (bh *BackupHandler) GetBackupJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract job ID from URL path
	jobID := r.URL.Path[len("/api/v1/backup/jobs/"):]
	if jobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	// Get backup job
	job, err := bh.jobManager.GetBackupJob(jobID)
	if err != nil {
		bh.logger.Error("Failed to get backup job",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", jobID),
			zap.Error(err))
		http.Error(w, "Backup job not found", http.StatusNotFound)
		return
	}

	// Prepare response
	response := GetBackupJobResponse{
		Job: job,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", jobID),
			zap.Error(err))
	}
}

// ListBackupJobs handles GET /api/v1/backup/jobs
func (bh *BackupHandler) ListBackupJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract query parameters
	businessID := r.URL.Query().Get("business_id")

	// Get backup jobs
	jobs, err := bh.jobManager.ListBackupJobs(businessID)
	if err != nil {
		bh.logger.Error("Failed to list backup jobs",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", businessID),
			zap.Error(err))
		http.Error(w, "Failed to list backup jobs", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := ListBackupJobsResponse{
		Jobs:  jobs,
		Total: len(jobs),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", businessID),
			zap.Error(err))
	}
}

// CancelBackupJob handles DELETE /api/v1/backup/jobs/{job_id}
func (bh *BackupHandler) CancelBackupJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract job ID from URL path
	jobID := r.URL.Path[len("/api/v1/backup/jobs/"):]
	if jobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	// Cancel backup job
	err := bh.jobManager.CancelBackupJob(jobID)
	if err != nil {
		bh.logger.Error("Failed to cancel backup job",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", jobID),
			zap.Error(err))
		http.Error(w, "Failed to cancel backup job", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send success response
	response := map[string]interface{}{
		"message": "Backup job cancelled successfully",
		"job_id":  jobID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", jobID),
			zap.Error(err))
	}
}

// ListBackups handles GET /api/v1/backup
func (bh *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract query parameters
	businessID := r.URL.Query().Get("business_id")

	// Get backups
	backups, err := bh.backupSvc.ListBackups(businessID)
	if err != nil {
		bh.logger.Error("Failed to list backups",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", businessID),
			zap.Error(err))
		http.Error(w, "Failed to list backups", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := ListBackupsResponse{
		Backups: backups,
		Total:   len(backups),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", businessID),
			zap.Error(err))
	}
}

// RestoreBackup handles POST /api/v1/backup/restore
func (bh *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Parse request body
	var req RestoreBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bh.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BackupID == "" {
		http.Error(w, "backup_id is required", http.StatusBadRequest)
		return
	}

	if req.RestoreType == "" {
		http.Error(w, "restore_type is required", http.StatusBadRequest)
		return
	}

	// Create restore request
	restoreReq := &RestoreRequest{
		BackupID:    req.BackupID,
		BusinessID:  req.BusinessID,
		RestoreType: req.RestoreType,
		Metadata:    req.Metadata,
	}

	// Restore backup
	response, err := bh.backupSvc.RestoreBackup(ctx, restoreReq)
	if err != nil {
		bh.logger.Error("Failed to restore backup",
			zap.String("request_id", requestID.(string)),
			zap.String("backup_id", req.BackupID),
			zap.Error(err))
		http.Error(w, "Failed to restore backup", http.StatusInternalServerError)
		return
	}

	// Prepare response
	restoreResponse := RestoreBackupResponse{
		RestoreID:   response.RestoreID,
		BackupID:    response.BackupID,
		Status:      string(response.Status),
		RecordCount: response.RecordCount,
		StartedAt:   response.StartedAt,
		Message:     "Backup restored successfully",
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(restoreResponse); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("backup_id", req.BackupID),
			zap.Error(err))
	}
}

// DeleteBackup handles DELETE /api/v1/backup/{backup_id}
func (bh *BackupHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Extract backup ID from URL path
	backupID := r.URL.Path[len("/api/v1/backup/"):]
	if backupID == "" {
		http.Error(w, "backup_id is required", http.StatusBadRequest)
		return
	}

	// Delete backup
	err := bh.backupSvc.DeleteBackup(backupID)
	if err != nil {
		bh.logger.Error("Failed to delete backup",
			zap.String("request_id", requestID.(string)),
			zap.String("backup_id", backupID),
			zap.Error(err))
		http.Error(w, "Failed to delete backup", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send success response
	response := map[string]interface{}{
		"message":   "Backup deleted successfully",
		"backup_id": backupID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("backup_id", backupID),
			zap.Error(err))
	}
}

// GetBackupStatistics handles GET /api/v1/backup/statistics
func (bh *BackupHandler) GetBackupStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Get backup statistics
	stats, err := bh.backupSvc.GetBackupStatistics()
	if err != nil {
		bh.logger.Error("Failed to get backup statistics",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Failed to get backup statistics", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := GetBackupStatisticsResponse{
		Statistics: stats,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
	}
}

// CleanupExpiredBackups handles POST /api/v1/backup/cleanup
func (bh *BackupHandler) CleanupExpiredBackups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Cleanup expired backups
	err := bh.backupSvc.CleanupExpiredBackups()
	if err != nil {
		bh.logger.Error("Failed to cleanup expired backups",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Failed to cleanup expired backups", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send success response
	response := map[string]interface{}{
		"message": "Expired backups cleaned up successfully",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
	}
}

// CleanupOldJobs handles POST /api/v1/backup/jobs/cleanup
func (bh *BackupHandler) CleanupOldJobs(w http.ResponseWriter, r *http.Request) {
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
	err = bh.jobManager.CleanupOldJobs(cutoffTime)
	if err != nil {
		bh.logger.Error("Failed to cleanup old jobs",
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
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
	}
}

// CreateBackupSchedule handles POST /api/v1/backup/schedules
func (bh *BackupHandler) CreateBackupSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Parse request body
	var req CreateBackupScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bh.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.BackupType == "" {
		http.Error(w, "backup_type is required", http.StatusBadRequest)
		return
	}

	if len(req.IncludeData) == 0 {
		http.Error(w, "include_data is required", http.StatusBadRequest)
		return
	}

	if req.Schedule == "" {
		http.Error(w, "schedule is required", http.StatusBadRequest)
		return
	}

	// Create backup schedule
	schedule := &BackupSchedule{
		ID:            fmt.Sprintf("schedule_%d", time.Now().UnixNano()),
		BusinessID:    req.BusinessID,
		Name:          req.Name,
		Description:   req.Description,
		BackupType:    req.BackupType,
		IncludeData:   req.IncludeData,
		Schedule:      req.Schedule,
		RetentionDays: req.RetentionDays,
		Enabled:       req.Enabled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      req.Metadata,
	}

	// Add to job manager
	err := bh.jobManager.CreateScheduledBackup(schedule)
	if err != nil {
		bh.logger.Error("Failed to create backup schedule",
			zap.String("request_id", requestID.(string)),
			zap.String("name", req.Name),
			zap.Error(err))
		http.Error(w, "Failed to create backup schedule", http.StatusInternalServerError)
		return
	}

	// Prepare response
	createResponse := CreateBackupScheduleResponse{
		ScheduleID: schedule.ID,
		Name:       schedule.Name,
		CreatedAt:  schedule.CreatedAt,
		Message:    "Backup schedule created successfully",
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(createResponse); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.String("schedule_id", schedule.ID),
			zap.Error(err))
	}
}

// ListBackupSchedules handles GET /api/v1/backup/schedules
func (bh *BackupHandler) ListBackupSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Get backup schedules
	schedules := bh.jobManager.scheduler.ListSchedules()

	// Prepare response
	response := ListBackupSchedulesResponse{
		Schedules: schedules,
		Total:     len(schedules),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		bh.logger.Error("Failed to encode response",
			zap.String("request_id", requestID.(string)),
			zap.Error(err))
	}
}

// RegisterRoutes registers the backup handler routes
func (bh *BackupHandler) RegisterRoutes(mux *http.ServeMux) {
	// Backup management routes
	mux.HandleFunc("POST /api/v1/backup", bh.CreateBackup)
	mux.HandleFunc("GET /api/v1/backup", bh.ListBackups)
	mux.HandleFunc("DELETE /api/v1/backup/{backup_id}", bh.DeleteBackup)
	mux.HandleFunc("GET /api/v1/backup/statistics", bh.GetBackupStatistics)
	mux.HandleFunc("POST /api/v1/backup/cleanup", bh.CleanupExpiredBackups)

	// Backup job management routes
	mux.HandleFunc("POST /api/v1/backup/jobs", bh.CreateBackupJob)
	mux.HandleFunc("GET /api/v1/backup/jobs/{job_id}", bh.GetBackupJob)
	mux.HandleFunc("GET /api/v1/backup/jobs", bh.ListBackupJobs)
	mux.HandleFunc("DELETE /api/v1/backup/jobs/{job_id}", bh.CancelBackupJob)
	mux.HandleFunc("POST /api/v1/backup/jobs/cleanup", bh.CleanupOldJobs)

	// Backup restore routes
	mux.HandleFunc("POST /api/v1/backup/restore", bh.RestoreBackup)

	// Backup schedule routes
	mux.HandleFunc("POST /api/v1/backup/schedules", bh.CreateBackupSchedule)
	mux.HandleFunc("GET /api/v1/backup/schedules", bh.ListBackupSchedules)
}
