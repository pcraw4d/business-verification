package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// LogRetentionDashboardHandler handles HTTP requests for log retention and archival
type LogRetentionDashboardHandler struct {
	retentionSystem *observability.LogRetentionSystem
	storageManager  *observability.LogStorageManager
	archiveManager  *observability.LogArchiveManager
	logger          *zap.Logger
}

// NewLogRetentionDashboardHandler creates a new log retention dashboard handler
func NewLogRetentionDashboardHandler(
	retentionSystem *observability.LogRetentionSystem,
	storageManager *observability.LogStorageManager,
	archiveManager *observability.LogArchiveManager,
	logger *zap.Logger,
) *LogRetentionDashboardHandler {
	return &LogRetentionDashboardHandler{
		retentionSystem: retentionSystem,
		storageManager:  storageManager,
		archiveManager:  archiveManager,
		logger:          logger,
	}
}

// GetRetentionMetrics returns current retention metrics
func (h *LogRetentionDashboardHandler) GetRetentionMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.retentionSystem.GetRetentionMetrics(ctx)
	if err != nil {
		h.logger.Error("Failed to get retention metrics", zap.Error(err))
		http.Error(w, "Failed to get retention metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetStorageUsage returns current storage usage statistics
func (h *LogRetentionDashboardHandler) GetStorageUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	usage, err := h.retentionSystem.GetStorageUsage(ctx)
	if err != nil {
		h.logger.Error("Failed to get storage usage", zap.Error(err))
		http.Error(w, "Failed to get storage usage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// GetStorageInfo returns storage information for all providers
func (h *LogRetentionDashboardHandler) GetStorageInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	info, err := h.storageManager.GetAggregatedStorageInfo(ctx)
	if err != nil {
		h.logger.Error("Failed to get storage info", zap.Error(err))
		http.Error(w, "Failed to get storage info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// RunManualCleanup runs a manual cleanup operation
func (h *LogRetentionDashboardHandler) RunManualCleanup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.retentionSystem.RunManualCleanup(ctx); err != nil {
		h.logger.Error("Manual cleanup failed", zap.Error(err))
		http.Error(w, "Manual cleanup failed", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": "Manual cleanup completed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ArchiveLogs archives logs from one storage tier to another
func (h *LogRetentionDashboardHandler) ArchiveLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		SourceTier string `json:"source_tier"`
		DestTier   string `json:"dest_tier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.SourceTier == "" || request.DestTier == "" {
		http.Error(w, "Source tier and destination tier are required", http.StatusBadRequest)
		return
	}

	if err := h.retentionSystem.ArchiveLogs(ctx, request.SourceTier, request.DestTier); err != nil {
		h.logger.Error("Log archival failed",
			zap.String("source_tier", request.SourceTier),
			zap.String("dest_tier", request.DestTier),
			zap.Error(err),
		)
		http.Error(w, "Log archival failed", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":      "success",
		"message":     "Log archival completed successfully",
		"source_tier": request.SourceTier,
		"dest_tier":   request.DestTier,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetArchiveList returns a list of archives
func (h *LogRetentionDashboardHandler) GetArchiveList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get query parameters
	archiverName := r.URL.Query().Get("archiver")
	prefix := r.URL.Query().Get("prefix")

	if archiverName == "" {
		archiverName = "default" // Use default archiver
	}

	archiver, err := h.archiveManager.GetArchiver(archiverName)
	if err != nil {
		h.logger.Error("Failed to get archiver", zap.String("archiver", archiverName), zap.Error(err))
		http.Error(w, "Failed to get archiver", http.StatusBadRequest)
		return
	}

	archives, err := archiver.ListArchives(ctx, prefix)
	if err != nil {
		h.logger.Error("Failed to list archives", zap.Error(err))
		http.Error(w, "Failed to list archives", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(archives)
}

// RestoreArchive restores an archive file
func (h *LogRetentionDashboardHandler) RestoreArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		ArchiverName string `json:"archiver_name"`
		ArchivePath  string `json:"archive_path"`
		DestPath     string `json:"dest_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ArchivePath == "" || request.DestPath == "" {
		http.Error(w, "Archive path and destination path are required", http.StatusBadRequest)
		return
	}

	if request.ArchiverName == "" {
		request.ArchiverName = "default"
	}

	archiver, err := h.archiveManager.GetArchiver(request.ArchiverName)
	if err != nil {
		h.logger.Error("Failed to get archiver", zap.String("archiver", request.ArchiverName), zap.Error(err))
		http.Error(w, "Failed to get archiver", http.StatusBadRequest)
		return
	}

	if err := archiver.Restore(ctx, request.ArchivePath, request.DestPath); err != nil {
		h.logger.Error("Archive restoration failed",
			zap.String("archive_path", request.ArchivePath),
			zap.String("dest_path", request.DestPath),
			zap.Error(err),
		)
		http.Error(w, "Archive restoration failed", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":       "success",
		"message":      "Archive restored successfully",
		"archive_path": request.ArchivePath,
		"dest_path":    request.DestPath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateArchive validates an archive file
func (h *LogRetentionDashboardHandler) ValidateArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		ArchiverName string `json:"archiver_name"`
		ArchivePath  string `json:"archive_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ArchivePath == "" {
		http.Error(w, "Archive path is required", http.StatusBadRequest)
		return
	}

	if request.ArchiverName == "" {
		request.ArchiverName = "default"
	}

	archiver, err := h.archiveManager.GetArchiver(request.ArchiverName)
	if err != nil {
		h.logger.Error("Failed to get archiver", zap.String("archiver", request.ArchiverName), zap.Error(err))
		http.Error(w, "Failed to get archiver", http.StatusBadRequest)
		return
	}

	if err := archiver.Validate(ctx, request.ArchivePath); err != nil {
		response := map[string]interface{}{
			"status":       "error",
			"message":      "Archive validation failed",
			"archive_path": request.ArchivePath,
			"error":        err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"status":       "success",
		"message":      "Archive validation passed",
		"archive_path": request.ArchivePath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteArchive deletes an archive file
func (h *LogRetentionDashboardHandler) DeleteArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		ArchiverName string `json:"archiver_name"`
		ArchivePath  string `json:"archive_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ArchivePath == "" {
		http.Error(w, "Archive path is required", http.StatusBadRequest)
		return
	}

	if request.ArchiverName == "" {
		request.ArchiverName = "default"
	}

	archiver, err := h.archiveManager.GetArchiver(request.ArchiverName)
	if err != nil {
		h.logger.Error("Failed to get archiver", zap.String("archiver", request.ArchiverName), zap.Error(err))
		http.Error(w, "Failed to get archiver", http.StatusBadRequest)
		return
	}

	if err := archiver.DeleteArchive(ctx, request.ArchivePath); err != nil {
		h.logger.Error("Archive deletion failed",
			zap.String("archive_path", request.ArchivePath),
			zap.Error(err),
		)
		http.Error(w, "Archive deletion failed", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":       "success",
		"message":      "Archive deleted successfully",
		"archive_path": request.ArchivePath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetArchiveInfo returns information about an archive file
func (h *LogRetentionDashboardHandler) GetArchiveInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	archiverName := r.URL.Query().Get("archiver")
	archivePath := r.URL.Query().Get("path")

	if archivePath == "" {
		http.Error(w, "Archive path is required", http.StatusBadRequest)
		return
	}

	if archiverName == "" {
		archiverName = "default"
	}

	archiver, err := h.archiveManager.GetArchiver(archiverName)
	if err != nil {
		h.logger.Error("Failed to get archiver", zap.String("archiver", archiverName), zap.Error(err))
		http.Error(w, "Failed to get archiver", http.StatusBadRequest)
		return
	}

	info, err := archiver.GetArchiveInfo(ctx, archivePath)
	if err != nil {
		h.logger.Error("Failed to get archive info",
			zap.String("archive_path", archivePath),
			zap.Error(err),
		)
		http.Error(w, "Failed to get archive info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// BulkArchive archives multiple files
func (h *LogRetentionDashboardHandler) BulkArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		ArchiverName string                       `json:"archiver_name"`
		Files        []string                     `json:"files"`
		DestPath     string                       `json:"dest_path"`
		Config       *observability.ArchiveConfig `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Files) == 0 || request.DestPath == "" {
		http.Error(w, "Files and destination path are required", http.StatusBadRequest)
		return
	}

	if request.ArchiverName == "" {
		request.ArchiverName = "default"
	}

	if request.Config == nil {
		request.Config = &observability.ArchiveConfig{
			CompressionEnabled: true,
			CompressionFormat:  "gzip",
			EncryptionEnabled:  false,
		}
	}

	if err := h.archiveManager.BulkArchive(ctx, request.ArchiverName, request.Files, request.DestPath, request.Config); err != nil {
		h.logger.Error("Bulk archive failed",
			zap.String("archiver", request.ArchiverName),
			zap.Strings("files", request.Files),
			zap.String("dest_path", request.DestPath),
			zap.Error(err),
		)
		http.Error(w, "Bulk archive failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":        "success",
		"message":       "Bulk archive completed successfully",
		"files_count":   len(request.Files),
		"dest_path":     request.DestPath,
		"archiver_name": request.ArchiverName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BulkRestore restores multiple archives
func (h *LogRetentionDashboardHandler) BulkRestore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		ArchiverName string   `json:"archiver_name"`
		Archives     []string `json:"archives"`
		DestPath     string   `json:"dest_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Archives) == 0 || request.DestPath == "" {
		http.Error(w, "Archives and destination path are required", http.StatusBadRequest)
		return
	}

	if request.ArchiverName == "" {
		request.ArchiverName = "default"
	}

	if err := h.archiveManager.BulkRestore(ctx, request.ArchiverName, request.Archives, request.DestPath); err != nil {
		h.logger.Error("Bulk restore failed",
			zap.String("archiver", request.ArchiverName),
			zap.Strings("archives", request.Archives),
			zap.String("dest_path", request.DestPath),
			zap.Error(err),
		)
		http.Error(w, "Bulk restore failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":         "success",
		"message":        "Bulk restore completed successfully",
		"archives_count": len(request.Archives),
		"dest_path":      request.DestPath,
		"archiver_name":  request.ArchiverName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateArchives validates multiple archives
func (h *LogRetentionDashboardHandler) ValidateArchives(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		ArchiverName string   `json:"archiver_name"`
		Archives     []string `json:"archives"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Archives) == 0 {
		http.Error(w, "Archives are required", http.StatusBadRequest)
		return
	}

	if request.ArchiverName == "" {
		request.ArchiverName = "default"
	}

	results, err := h.archiveManager.ValidateArchives(ctx, request.ArchiverName, request.Archives)
	if err != nil {
		h.logger.Error("Archive validation failed",
			zap.String("archiver", request.ArchiverName),
			zap.Strings("archives", request.Archives),
			zap.Error(err),
		)
		http.Error(w, "Archive validation failed", http.StatusInternalServerError)
		return
	}

	// Count validation results
	validCount := 0
	invalidCount := 0
	for _, err := range results {
		if err == nil {
			validCount++
		} else {
			invalidCount++
		}
	}

	response := map[string]interface{}{
		"status":         "success",
		"message":        "Archive validation completed",
		"total_archives": len(request.Archives),
		"valid_count":    validCount,
		"invalid_count":  invalidCount,
		"results":        results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStorageProviders returns a list of registered storage providers
func (h *LogRetentionDashboardHandler) GetStorageProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.storageManager.ListProviders()

	response := map[string]interface{}{
		"providers": providers,
		"count":     len(providers),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetArchivers returns a list of registered archivers
func (h *LogRetentionDashboardHandler) GetArchivers(w http.ResponseWriter, r *http.Request) {
	archivers := h.archiveManager.ListArchivers()

	response := map[string]interface{}{
		"archivers": archivers,
		"count":     len(archivers),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRetentionConfiguration returns the current retention configuration
func (h *LogRetentionDashboardHandler) GetRetentionConfiguration(w http.ResponseWriter, r *http.Request) {
	// This would return the current configuration
	// For now, we'll return a mock configuration
	config := map[string]interface{}{
		"hot_retention_period":     "168h",   // 7 days
		"warm_retention_period":    "720h",   // 30 days
		"cold_retention_period":    "8760h",  // 1 year
		"archive_retention_period": "43800h", // 5 years
		"compression_enabled":      true,
		"compression_format":       "gzip",
		"encryption_enabled":       false,
		"cleanup_interval":         "1h",
		"max_log_file_size":        "104857600", // 100MB
		"max_log_files_per_day":    24,
		"cleanup_batch_size":       100,
		"enable_metrics":           true,
		"enable_health_checks":     true,
		"enable_notifications":     true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// ProcessLogEntry processes a single log entry for retention
func (h *LogRetentionDashboardHandler) ProcessLogEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var entry observability.LogEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid log entry", http.StatusBadRequest)
		return
	}

	if err := h.retentionSystem.ProcessLogEntry(ctx, entry); err != nil {
		h.logger.Error("Failed to process log entry", zap.Error(err))
		http.Error(w, "Failed to process log entry", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": "Log entry processed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
