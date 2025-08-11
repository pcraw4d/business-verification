package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// BackupHandler handles backup-related HTTP requests
type BackupHandler struct {
	backupService *database.BackupService
	logger        *observability.Logger
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(backupService *database.BackupService, logger *observability.Logger) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
		logger:        logger,
	}
}

// BackupRequest represents a backup creation request
type BackupRequest struct {
	Compression bool   `json:"compression"`
	Encryption  bool   `json:"encryption"`
	CrossRegion bool   `json:"cross_region"`
	Description string `json:"description"`
}

// BackupResponse represents a backup response
type BackupResponse struct {
	BackupID    string    `json:"backup_id"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    string    `json:"duration"`
	Success     bool      `json:"success"`
	Compressed  bool      `json:"compressed"`
	Encrypted   bool      `json:"encrypted"`
	CrossRegion bool      `json:"cross_region"`
	Error       string    `json:"error,omitempty"`
}

// BackupListResponse represents a list of backups
type BackupListResponse struct {
	Backups []BackupResponse `json:"backups"`
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

// BackupStatsResponse represents backup statistics
type BackupStatsResponse struct {
	TotalBackups      int        `json:"total_backups"`
	SuccessfulBackups int        `json:"successful_backups"`
	FailedBackups     int        `json:"failed_backups"`
	SuccessRate       float64    `json:"success_rate"`
	TotalSize         int64      `json:"total_size"`
	AverageSize       int64      `json:"average_size"`
	OldestBackup      *time.Time `json:"oldest_backup,omitempty"`
	NewestBackup      *time.Time `json:"newest_backup,omitempty"`
}

// CreateBackup handles POST /v1/backup
func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode backup request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Creating backup",
		"compression", req.Compression,
		"encryption", req.Encryption,
		"cross_region", req.CrossRegion,
	)

	// Create backup
	result, err := h.backupService.CreateBackup(ctx)
	if err != nil {
		h.logger.Error("Failed to create backup", "error", err)
		http.Error(w, fmt.Sprintf("Failed to create backup: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := BackupResponse{
		BackupID:    result.BackupID,
		Filename:    result.Filename,
		Size:        result.Size,
		Checksum:    result.Checksum,
		StartTime:   result.StartTime,
		EndTime:     result.EndTime,
		Duration:    result.Duration.String(),
		Success:     result.Success,
		Compressed:  result.Compressed,
		Encrypted:   result.Encrypted,
		CrossRegion: result.CrossRegion,
	}

	if result.Error != nil {
		response.Error = result.Error.Error()
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode backup response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Backup created successfully", "backup_id", result.BackupID)
}

// ListBackups handles GET /v1/backup
func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	h.logger.Info("Listing backups", "page", page, "limit", limit)

	// Get backups
	backups, err := h.backupService.ListBackups(ctx)
	if err != nil {
		h.logger.Error("Failed to list backups", "error", err)
		http.Error(w, fmt.Sprintf("Failed to list backups: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var responseBackups []BackupResponse
	for _, backup := range backups {
		responseBackup := BackupResponse{
			BackupID:    backup.BackupID,
			Filename:    backup.Filename,
			Size:        backup.Size,
			Checksum:    backup.Checksum,
			StartTime:   backup.StartTime,
			EndTime:     backup.EndTime,
			Duration:    backup.Duration.String(),
			Success:     backup.Success,
			Compressed:  backup.Compressed,
			Encrypted:   backup.Encrypted,
			CrossRegion: backup.CrossRegion,
		}

		if backup.Error != nil {
			responseBackup.Error = backup.Error.Error()
		}

		responseBackups = append(responseBackups, responseBackup)
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit
	if start >= len(responseBackups) {
		start = len(responseBackups)
	}
	if end > len(responseBackups) {
		end = len(responseBackups)
	}

	response := BackupListResponse{
		Backups: responseBackups[start:end],
		Total:   len(backups),
		Page:    page,
		Limit:   limit,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode backup list response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Backup list retrieved successfully", "total", len(backups))
}

// GetBackup handles GET /v1/backup/{backup_id}
func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract backup ID from URL
	backupID := r.PathValue("backup_id")
	if backupID == "" {
		http.Error(w, "Backup ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting backup details", "backup_id", backupID)

	// Get backup metadata
	backups, err := h.backupService.ListBackups(ctx)
	if err != nil {
		h.logger.Error("Failed to get backup", "error", err, "backup_id", backupID)
		http.Error(w, fmt.Sprintf("Failed to get backup: %v", err), http.StatusInternalServerError)
		return
	}

	// Find the specific backup
	var targetBackup *database.BackupResult
	for _, backup := range backups {
		if backup.BackupID == backupID {
			targetBackup = backup
			break
		}
	}

	if targetBackup == nil {
		h.logger.Warn("Backup not found", "backup_id", backupID)
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	// Convert to response format
	response := BackupResponse{
		BackupID:    targetBackup.BackupID,
		Filename:    targetBackup.Filename,
		Size:        targetBackup.Size,
		Checksum:    targetBackup.Checksum,
		StartTime:   targetBackup.StartTime,
		EndTime:     targetBackup.EndTime,
		Duration:    targetBackup.Duration.String(),
		Success:     targetBackup.Success,
		Compressed:  targetBackup.Compressed,
		Encrypted:   targetBackup.Encrypted,
		CrossRegion: targetBackup.CrossRegion,
	}

	if targetBackup.Error != nil {
		response.Error = targetBackup.Error.Error()
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode backup response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Backup details retrieved successfully", "backup_id", backupID)
}

// RestoreBackup handles POST /v1/backup/{backup_id}/restore
func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract backup ID from URL
	backupID := r.PathValue("backup_id")
	if backupID == "" {
		http.Error(w, "Backup ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Restoring backup", "backup_id", backupID)

	// Restore backup
	if err := h.backupService.RestoreBackup(ctx, backupID); err != nil {
		h.logger.Error("Failed to restore backup", "error", err, "backup_id", backupID)
		http.Error(w, fmt.Sprintf("Failed to restore backup: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":   "Backup restored successfully",
		"backup_id": backupID,
		"timestamp": time.Now().UTC(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode restore response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Backup restored successfully", "backup_id", backupID)
}

// ValidateBackup handles POST /v1/backup/{backup_id}/validate
func (h *BackupHandler) ValidateBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract backup ID from URL
	backupID := r.PathValue("backup_id")
	if backupID == "" {
		http.Error(w, "Backup ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Validating backup", "backup_id", backupID)

	// Validate backup
	if err := h.backupService.ValidateBackup(ctx, backupID); err != nil {
		h.logger.Error("Backup validation failed", "error", err, "backup_id", backupID)
		http.Error(w, fmt.Sprintf("Backup validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":   "Backup validation successful",
		"backup_id": backupID,
		"timestamp": time.Now().UTC(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode validation response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Backup validation successful", "backup_id", backupID)
}

// CleanupBackups handles POST /v1/backup/cleanup
func (h *BackupHandler) CleanupBackups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Cleaning up old backups")

	// Cleanup old backups
	if err := h.backupService.CleanupOldBackups(ctx); err != nil {
		h.logger.Error("Failed to cleanup backups", "error", err)
		http.Error(w, fmt.Sprintf("Failed to cleanup backups: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":   "Backup cleanup completed successfully",
		"timestamp": time.Now().UTC(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode cleanup response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Backup cleanup completed successfully")
}

// GetBackupStats handles GET /v1/backup/stats
func (h *BackupHandler) GetBackupStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Getting backup statistics")

	// Get backups
	backups, err := h.backupService.ListBackups(ctx)
	if err != nil {
		h.logger.Error("Failed to get backup statistics", "error", err)
		http.Error(w, fmt.Sprintf("Failed to get backup statistics: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate statistics
	var totalSize int64
	var successfulBackups, failedBackups int
	var oldestBackup, newestBackup *time.Time

	for _, backup := range backups {
		totalSize += backup.Size

		if backup.Success {
			successfulBackups++
		} else {
			failedBackups++
		}

		// Track oldest and newest backups
		if oldestBackup == nil || backup.StartTime.Before(*oldestBackup) {
			oldestBackup = &backup.StartTime
		}
		if newestBackup == nil || backup.StartTime.After(*newestBackup) {
			newestBackup = &backup.StartTime
		}
	}

	totalBackups := len(backups)
	var successRate float64
	if totalBackups > 0 {
		successRate = float64(successfulBackups) / float64(totalBackups) * 100
	}

	var averageSize int64
	if totalBackups > 0 {
		averageSize = totalSize / int64(totalBackups)
	}

	response := BackupStatsResponse{
		TotalBackups:      totalBackups,
		SuccessfulBackups: successfulBackups,
		FailedBackups:     failedBackups,
		SuccessRate:       successRate,
		TotalSize:         totalSize,
		AverageSize:       averageSize,
		OldestBackup:      oldestBackup,
		NewestBackup:      newestBackup,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode backup stats response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Backup statistics retrieved successfully", "total_backups", totalBackups)
}

// TestBackup handles POST /v1/backup/test
func (h *BackupHandler) TestBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Testing backup system")

	// Create a test backup
	result, err := h.backupService.CreateBackup(ctx)
	if err != nil {
		h.logger.Error("Backup test failed", "error", err)
		http.Error(w, fmt.Sprintf("Backup test failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Validate the test backup
	if err := h.backupService.ValidateBackup(ctx, result.BackupID); err != nil {
		h.logger.Error("Backup validation test failed", "error", err, "backup_id", result.BackupID)
		http.Error(w, fmt.Sprintf("Backup validation test failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":   "Backup test completed successfully",
		"backup_id": result.BackupID,
		"size":      result.Size,
		"duration":  result.Duration.String(),
		"timestamp": time.Now().UTC(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode test response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Backup test completed successfully", "backup_id", result.BackupID)
}
