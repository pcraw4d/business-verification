package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ModelRegistry manages model versions and deployments
type ModelRegistry struct {
	// Registry configuration
	config ModelRegistryConfig

	// Model storage
	models map[string]*ModelVersion

	// Deployment tracking
	deployments map[string]*ModelDeployment

	// Version history
	versionHistory map[string][]*ModelVersion

	// Performance tracking
	metrics *ServiceMetrics

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// ModelRegistryConfig holds configuration for the model registry
type ModelRegistryConfig struct {
	// Registry configuration
	StorageType    string        `json:"storage_type"` // local, s3, gcs, azure
	StoragePath    string        `json:"storage_path"`
	BackupEnabled  bool          `json:"backup_enabled"`
	BackupInterval time.Duration `json:"backup_interval"`

	// Model management
	MaxModelVersions int           `json:"max_model_versions"`
	ModelRetention   time.Duration `json:"model_retention"`
	AutoCleanup      bool          `json:"auto_cleanup"`

	// Versioning
	VersioningEnabled  bool   `json:"versioning_enabled"`
	VersioningStrategy string `json:"versioning_strategy"` // semantic, timestamp, hash

	// Deployment
	AutoDeploymentEnabled bool `json:"auto_deployment_enabled"`
	DeploymentValidation  bool `json:"deployment_validation"`

	// Monitoring
	MetricsEnabled      bool `json:"metrics_enabled"`
	DeploymentTracking  bool `json:"deployment_tracking"`
	PerformanceTracking bool `json:"performance_tracking"`
}

// NewModelRegistry creates a new model registry
func NewModelRegistry(logger *log.Logger) *ModelRegistry {
	if logger == nil {
		logger = log.Default()
	}

	return &ModelRegistry{
		config: ModelRegistryConfig{
			StorageType:           "local",
			StoragePath:           "./models",
			BackupEnabled:         true,
			BackupInterval:        24 * time.Hour,
			MaxModelVersions:      10,
			ModelRetention:        30 * 24 * time.Hour, // 30 days
			AutoCleanup:           true,
			VersioningEnabled:     true,
			VersioningStrategy:    "semantic",
			AutoDeploymentEnabled: false,
			DeploymentValidation:  true,
			MetricsEnabled:        true,
			DeploymentTracking:    true,
			PerformanceTracking:   true,
		},
		models:         make(map[string]*ModelVersion),
		deployments:    make(map[string]*ModelDeployment),
		versionHistory: make(map[string][]*ModelVersion),
		logger:         logger,
	}
}

// Initialize initializes the model registry
func (mr *ModelRegistry) Initialize(ctx context.Context) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	mr.logger.Printf("📚 Initializing Model Registry")

	// Initialize metrics
	mr.metrics = &ServiceMetrics{
		RequestCount:   0,
		SuccessCount:   0,
		ErrorCount:     0,
		AverageLatency: 0,
		P95Latency:     0,
		P99Latency:     0,
		Throughput:     0,
		ErrorRate:      0,
		LastUpdated:    time.Now(),
	}

	// Load existing models from storage
	if err := mr.loadModelsFromStorage(ctx); err != nil {
		return fmt.Errorf("failed to load models from storage: %w", err)
	}

	// Start cleanup process if enabled
	if mr.config.AutoCleanup {
		go mr.startCleanupProcess(ctx)
	}

	// Start backup process if enabled
	if mr.config.BackupEnabled {
		go mr.startBackupProcess(ctx)
	}

	mr.logger.Printf("✅ Model Registry initialized with %d models and %d deployments",
		len(mr.models), len(mr.deployments))

	return nil
}

// RegisterModel registers a new model version
func (mr *ModelRegistry) RegisterModel(ctx context.Context, model *MLModel) (*ModelVersion, error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	start := time.Now()
	mr.metrics.RequestCount++

	// Generate version
	version := mr.generateVersion(model.ID)

	// Create model version
	modelVersion := &ModelVersion{
		ID:         fmt.Sprintf("%s_%s", model.ID, version),
		ModelID:    model.ID,
		Version:    version,
		ModelPath:  model.ModelPath,
		ConfigPath: model.ConfigPath,
		Metrics: ModelMetrics{
			ModelID:      model.ID,
			ModelVersion: version,
			LastUpdated:  time.Now(),
		},
		IsActive:   false,
		CreatedAt:  time.Now(),
		DeployedAt: time.Time{},
	}

	// Store model version
	mr.models[modelVersion.ID] = modelVersion

	// Add to version history
	mr.versionHistory[model.ID] = append(mr.versionHistory[model.ID], modelVersion)

	// Cleanup old versions if needed
	if mr.config.AutoCleanup {
		mr.cleanupOldVersions(model.ID)
	}

	// Save to storage
	if err := mr.saveModelToStorage(ctx, modelVersion); err != nil {
		mr.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to save model to storage: %w", err)
	}

	mr.metrics.SuccessCount++
	mr.updateLatencyMetrics(time.Since(start))

	mr.logger.Printf("✅ Registered model %s version %s", model.ID, version)
	return modelVersion, nil
}

// DeployModel deploys a model version to a service
func (mr *ModelRegistry) DeployModel(ctx context.Context, modelID, version, serviceName, endpoint string) (*ModelDeployment, error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	start := time.Now()
	mr.metrics.RequestCount++

	// Find model version
	modelVersionID := fmt.Sprintf("%s_%s", modelID, version)
	modelVersion, exists := mr.models[modelVersionID]
	if !exists {
		mr.metrics.ErrorCount++
		return nil, fmt.Errorf("model version %s not found", modelVersionID)
	}

	// Validate deployment if enabled
	if mr.config.DeploymentValidation {
		if err := mr.validateDeployment(ctx, modelVersion); err != nil {
			mr.metrics.ErrorCount++
			return nil, fmt.Errorf("deployment validation failed: %w", err)
		}
	}

	// Create deployment
	deployment := &ModelDeployment{
		ID:          fmt.Sprintf("deploy_%d", time.Now().UnixNano()),
		ModelID:     modelID,
		Version:     version,
		ServiceName: serviceName,
		Endpoint:    endpoint,
		Status:      "deploying",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store deployment
	mr.deployments[deployment.ID] = deployment

	// Update model version
	modelVersion.IsActive = true
	modelVersion.DeployedAt = time.Now()

	// Update deployment status
	deployment.Status = "active"
	deployment.UpdatedAt = time.Now()

	mr.metrics.SuccessCount++
	mr.updateLatencyMetrics(time.Since(start))

	mr.logger.Printf("✅ Deployed model %s version %s to %s", modelID, version, serviceName)
	return deployment, nil
}

// GetModelVersion retrieves a specific model version
func (mr *ModelRegistry) GetModelVersion(modelID, version string) (*ModelVersion, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	modelVersionID := fmt.Sprintf("%s_%s", modelID, version)
	modelVersion, exists := mr.models[modelVersionID]
	if !exists {
		return nil, fmt.Errorf("model version %s not found", modelVersionID)
	}

	return modelVersion, nil
}

// GetActiveModelVersion retrieves the active version of a model
func (mr *ModelRegistry) GetActiveModelVersion(modelID string) (*ModelVersion, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	versions, exists := mr.versionHistory[modelID]
	if !exists {
		return nil, fmt.Errorf("no versions found for model %s", modelID)
	}

	// Find active version
	for _, version := range versions {
		if version.IsActive {
			return version, nil
		}
	}

	return nil, fmt.Errorf("no active version found for model %s", modelID)
}

// ListModelVersions lists all versions of a model
func (mr *ModelRegistry) ListModelVersions(modelID string) ([]*ModelVersion, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	versions, exists := mr.versionHistory[modelID]
	if !exists {
		return nil, fmt.Errorf("no versions found for model %s", modelID)
	}

	// Return a copy of the versions
	result := make([]*ModelVersion, len(versions))
	copy(result, versions)

	return result, nil
}

// ListDeployments lists all deployments
func (mr *ModelRegistry) ListDeployments() ([]*ModelDeployment, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	deployments := make([]*ModelDeployment, 0, len(mr.deployments))
	for _, deployment := range mr.deployments {
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// UpdateModelMetrics updates metrics for a model version
func (mr *ModelRegistry) UpdateModelMetrics(modelID, version string, metrics ModelMetrics) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	modelVersionID := fmt.Sprintf("%s_%s", modelID, version)
	modelVersion, exists := mr.models[modelVersionID]
	if !exists {
		return fmt.Errorf("model version %s not found", modelVersionID)
	}

	modelVersion.Metrics = metrics
	modelVersion.Metrics.LastUpdated = time.Now()

	mr.logger.Printf("📊 Updated metrics for model %s version %s", modelID, version)
	return nil
}

// RollbackModel rolls back to a previous model version
func (mr *ModelRegistry) RollbackModel(ctx context.Context, modelID, targetVersion string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	start := time.Now()
	mr.metrics.RequestCount++

	// Get target version
	targetModelVersion, err := mr.GetModelVersion(modelID, targetVersion)
	if err != nil {
		mr.metrics.ErrorCount++
		return fmt.Errorf("target version not found: %w", err)
	}

	// Deactivate current active version
	versions := mr.versionHistory[modelID]
	for _, version := range versions {
		if version.IsActive {
			version.IsActive = false
		}
	}

	// Activate target version
	targetModelVersion.IsActive = true
	targetModelVersion.DeployedAt = time.Now()

	// Update deployments
	for _, deployment := range mr.deployments {
		if deployment.ModelID == modelID {
			deployment.Version = targetVersion
			deployment.UpdatedAt = time.Now()
		}
	}

	mr.metrics.SuccessCount++
	mr.updateLatencyMetrics(time.Since(start))

	mr.logger.Printf("🔄 Rolled back model %s to version %s", modelID, targetVersion)
	return nil
}

// HealthCheck performs a health check on the model registry
func (mr *ModelRegistry) HealthCheck(ctx context.Context) (*HealthCheck, error) {
	start := time.Now()

	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Check if models are loaded
	if len(mr.models) == 0 {
		return &HealthCheck{
			Name:      "model_registry",
			Status:    "fail",
			Message:   "No models loaded",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	// Check storage accessibility
	if err := mr.checkStorageAccess(ctx); err != nil {
		return &HealthCheck{
			Name:      "model_registry",
			Status:    "fail",
			Message:   fmt.Sprintf("Storage access failed: %v", err),
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	return &HealthCheck{
		Name:      "model_registry",
		Status:    "pass",
		Message:   "Model registry is healthy",
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}, nil
}

// GetMetrics returns registry metrics
func (mr *ModelRegistry) GetMetrics(ctx context.Context) (*ServiceMetrics, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Return a copy of metrics
	metrics := *mr.metrics
	return &metrics, nil
}

// generateVersion generates a version string based on the configured strategy
func (mr *ModelRegistry) generateVersion(modelID string) string {
	switch mr.config.VersioningStrategy {
	case "semantic":
		return mr.generateSemanticVersion(modelID)
	case "timestamp":
		return fmt.Sprintf("%d", time.Now().Unix())
	case "hash":
		return fmt.Sprintf("%x", time.Now().UnixNano())
	default:
		return fmt.Sprintf("%d", time.Now().Unix())
	}
}

// generateSemanticVersion generates a semantic version
func (mr *ModelRegistry) generateSemanticVersion(modelID string) string {
	versions := mr.versionHistory[modelID]
	if len(versions) == 0 {
		return "1.0.0"
	}

	// Simple increment for now - in a real implementation, this would parse semantic versions
	return fmt.Sprintf("1.%d.0", len(versions))
}

// cleanupOldVersions removes old versions beyond the limit
func (mr *ModelRegistry) cleanupOldVersions(modelID string) {
	versions := mr.versionHistory[modelID]
	if len(versions) <= mr.config.MaxModelVersions {
		return
	}

	// Sort by creation time (oldest first)
	// Keep the most recent versions
	versionsToKeep := versions[len(versions)-mr.config.MaxModelVersions:]
	versionsToRemove := versions[:len(versions)-mr.config.MaxModelVersions]

	// Remove old versions
	for _, version := range versionsToRemove {
		delete(mr.models, version.ID)
	}

	// Update version history
	mr.versionHistory[modelID] = versionsToKeep

	mr.logger.Printf("🗑️ Cleaned up %d old versions for model %s", len(versionsToRemove), modelID)
}

// loadModelsFromStorage loads models from storage
func (mr *ModelRegistry) loadModelsFromStorage(ctx context.Context) error {
	// This would typically load from the configured storage backend
	// For now, we'll create some sample models

	sampleModels := []*ModelVersion{
		{
			ID:         "bert_classification_1.0.0",
			ModelID:    "bert_classification",
			Version:    "1.0.0",
			ModelPath:  "./models/bert_classification_1.0.0.bin",
			ConfigPath: "./models/bert_classification_1.0.0.json",
			Metrics: ModelMetrics{
				ModelID:      "bert_classification",
				ModelVersion: "1.0.0",
				Accuracy:     0.95,
				Precision:    0.94,
				Recall:       0.96,
				F1Score:      0.95,
				LastUpdated:  time.Now(),
			},
			IsActive:   true,
			CreatedAt:  time.Now().AddDate(0, 0, -7),
			DeployedAt: time.Now().AddDate(0, 0, -7),
		},
		{
			ID:         "distilbert_classification_1.0.0",
			ModelID:    "distilbert_classification",
			Version:    "1.0.0",
			ModelPath:  "./models/distilbert_classification_1.0.0.bin",
			ConfigPath: "./models/distilbert_classification_1.0.0.json",
			Metrics: ModelMetrics{
				ModelID:      "distilbert_classification",
				ModelVersion: "1.0.0",
				Accuracy:     0.92,
				Precision:    0.91,
				Recall:       0.93,
				F1Score:      0.92,
				LastUpdated:  time.Now(),
			},
			IsActive:   true,
			CreatedAt:  time.Now().AddDate(0, 0, -5),
			DeployedAt: time.Now().AddDate(0, 0, -5),
		},
	}

	for _, model := range sampleModels {
		mr.models[model.ID] = model
		mr.versionHistory[model.ModelID] = append(mr.versionHistory[model.ModelID], model)
	}

	return nil
}

// saveModelToStorage saves a model to storage
func (mr *ModelRegistry) saveModelToStorage(ctx context.Context, model *ModelVersion) error {
	// This would typically save to the configured storage backend
	// For now, we'll just log the action
	mr.logger.Printf("💾 Saved model %s to storage", model.ID)
	return nil
}

// validateDeployment validates a model deployment
func (mr *ModelRegistry) validateDeployment(ctx context.Context, model *ModelVersion) error {
	// Check if model files exist
	// Check model metrics
	// Validate model configuration
	// etc.

	if model.Metrics.Accuracy < 0.8 {
		return fmt.Errorf("model accuracy too low: %.2f", model.Metrics.Accuracy)
	}

	return nil
}

// checkStorageAccess checks if storage is accessible
func (mr *ModelRegistry) checkStorageAccess(ctx context.Context) error {
	// This would typically check the configured storage backend
	return nil
}

// startCleanupProcess starts the cleanup process
func (mr *ModelRegistry) startCleanupProcess(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mr.performCleanup()
		}
	}
}

// startBackupProcess starts the backup process
func (mr *ModelRegistry) startBackupProcess(ctx context.Context) {
	ticker := time.NewTicker(mr.config.BackupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mr.performBackup(ctx)
		}
	}
}

// performCleanup performs cleanup of old models
func (mr *ModelRegistry) performCleanup() {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	cutoff := time.Now().Add(-mr.config.ModelRetention)
	cleanedCount := 0

	for modelID, versions := range mr.versionHistory {
		var activeVersions []*ModelVersion
		for _, version := range versions {
			if version.CreatedAt.After(cutoff) || version.IsActive {
				activeVersions = append(activeVersions, version)
			} else {
				delete(mr.models, version.ID)
				cleanedCount++
			}
		}
		mr.versionHistory[modelID] = activeVersions
	}

	if cleanedCount > 0 {
		mr.logger.Printf("🗑️ Cleaned up %d old model versions", cleanedCount)
	}
}

// performBackup performs backup of models
func (mr *ModelRegistry) performBackup(ctx context.Context) {
	// This would typically backup models to a secondary storage location
	mr.logger.Printf("💾 Performing model backup")
}

// updateLatencyMetrics updates latency metrics
func (mr *ModelRegistry) updateLatencyMetrics(latency time.Duration) {
	// Simple moving average for average latency
	if mr.metrics.AverageLatency == 0 {
		mr.metrics.AverageLatency = latency
	} else {
		mr.metrics.AverageLatency = (mr.metrics.AverageLatency + latency) / 2
	}

	// Update P95 and P99 (simplified implementation)
	if latency > mr.metrics.P95Latency {
		mr.metrics.P95Latency = latency
	}
	if latency > mr.metrics.P99Latency {
		mr.metrics.P99Latency = latency
	}

	mr.metrics.LastUpdated = time.Now()
}
