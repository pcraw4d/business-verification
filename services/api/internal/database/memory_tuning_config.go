package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
)

// MemoryTuningConfig provides memory optimization configuration for PostgreSQL
// Optimized for classification and risk assessment workloads
type MemoryTuningConfig struct {
	// Core memory settings
	SharedBuffers      string `json:"shared_buffers"`       // 25% of system RAM
	EffectiveCacheSize string `json:"effective_cache_size"` // 75% of system RAM
	WorkMem            string `json:"work_mem"`             // Memory for sorting/hash joins
	MaintenanceWorkMem string `json:"maintenance_work_mem"` // Memory for maintenance ops

	// Query optimization settings
	RandomPageCost         float64 `json:"random_page_cost"`         // SSD optimization
	EffectiveIOConcurrency int     `json:"effective_io_concurrency"` // SSD optimization

	// WAL and checkpoint settings
	WALBuffers                 string  `json:"wal_buffers"`  // WAL buffer size
	MaxWALSize                 string  `json:"max_wal_size"` // Maximum WAL size
	MinWALSize                 string  `json:"min_wal_size"` // Minimum WAL size
	CheckpointCompletionTarget float64 `json:"checkpoint_completion_target"`

	// Parallel query settings
	MaxParallelWorkersPerGather int     `json:"max_parallel_workers_per_gather"`
	MaxParallelWorkers          int     `json:"max_parallel_workers"`
	ParallelTupleCost           float64 `json:"parallel_tuple_cost"`
	ParallelSetupCost           float64 `json:"parallel_setup_cost"`

	// Statistics and planning
	DefaultStatisticsTarget int     `json:"default_statistics_target"`
	CPUTupleCost            float64 `json:"cpu_tuple_cost"`
	CPUIndexTupleCost       float64 `json:"cpu_index_tuple_cost"`
	CPUOperatorCost         float64 `json:"cpu_operator_cost"`
}

// MemoryTuningOptimizer provides memory optimization for PostgreSQL
type MemoryTuningOptimizer struct {
	db     *sql.DB
	logger *log.Logger
}

// NewMemoryTuningOptimizer creates a new memory tuning optimizer
func NewMemoryTuningOptimizer(db *sql.DB, logger *log.Logger) *MemoryTuningOptimizer {
	return &MemoryTuningOptimizer{
		db:     db,
		logger: logger,
	}
}

// GetOptimizedMemoryConfig returns optimized memory configuration
// based on system resources and workload characteristics
func (mto *MemoryTuningOptimizer) GetOptimizedMemoryConfig() *MemoryTuningConfig {
	// Get system memory information
	systemRAM := mto.getSystemRAM()

	// Calculate optimal settings based on system resources
	sharedBuffers := mto.calculateSharedBuffers(systemRAM)
	effectiveCacheSize := mto.calculateEffectiveCacheSize(systemRAM)
	workMem := mto.calculateWorkMem(systemRAM)
	maintenanceWorkMem := mto.calculateMaintenanceWorkMem(systemRAM)

	return &MemoryTuningConfig{
		// Core memory settings
		SharedBuffers:      sharedBuffers,
		EffectiveCacheSize: effectiveCacheSize,
		WorkMem:            workMem,
		MaintenanceWorkMem: maintenanceWorkMem,

		// Query optimization settings
		RandomPageCost:         1.1, // SSD optimized
		EffectiveIOConcurrency: 200, // SSD optimized

		// WAL and checkpoint settings
		WALBuffers:                 "16MB",
		MaxWALSize:                 "1GB",
		MinWALSize:                 "256MB",
		CheckpointCompletionTarget: 0.7,

		// Parallel query settings
		MaxParallelWorkersPerGather: 4,
		MaxParallelWorkers:          8,
		ParallelTupleCost:           0.1,
		ParallelSetupCost:           1000.0,

		// Statistics and planning
		DefaultStatisticsTarget: 100,
		CPUTupleCost:            0.01,
		CPUIndexTupleCost:       0.005,
		CPUOperatorCost:         0.0025,
	}
}

// getSystemRAM estimates system RAM (in MB)
func (mto *MemoryTuningOptimizer) getSystemRAM() int64 {
	// For Supabase, we need to estimate based on plan
	// This is a simplified estimation - in production, you'd get this from Supabase API

	// Default to 1GB for development, adjust based on Supabase plan
	// Pro plan: 1GB, Team plan: 2GB, Enterprise: 4GB+
	defaultRAM := int64(1024) // 1GB in MB

	// Try to get actual memory from system
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// If we can get system memory, use it
	if m.Sys > 0 {
		// Convert bytes to MB
		systemRAM := int64(m.Sys / 1024 / 1024)
		if systemRAM > 0 {
			return systemRAM
		}
	}

	return defaultRAM
}

// calculateSharedBuffers calculates optimal shared buffers (25% of RAM)
func (mto *MemoryTuningOptimizer) calculateSharedBuffers(systemRAM int64) string {
	// 25% of system RAM for shared buffers
	sharedBuffersMB := systemRAM / 4

	// Ensure minimum of 128MB
	if sharedBuffersMB < 128 {
		sharedBuffersMB = 128
	}

	// Cap at 2GB for safety
	if sharedBuffersMB > 2048 {
		sharedBuffersMB = 2048
	}

	return fmt.Sprintf("%dMB", sharedBuffersMB)
}

// calculateEffectiveCacheSize calculates optimal effective cache size (75% of RAM)
func (mto *MemoryTuningOptimizer) calculateEffectiveCacheSize(systemRAM int64) string {
	// 75% of system RAM for effective cache size
	effectiveCacheMB := (systemRAM * 3) / 4

	// Ensure minimum of 1GB
	if effectiveCacheMB < 1024 {
		effectiveCacheMB = 1024
	}

	return fmt.Sprintf("%dMB", effectiveCacheMB)
}

// calculateWorkMem calculates optimal work memory
func (mto *MemoryTuningOptimizer) calculateWorkMem(systemRAM int64) string {
	// For classification and risk assessment workloads:
	// - Complex queries with sorting and hash joins
	// - JSONB operations for risk assessments
	// - Text search operations for classification

	// Base calculation: 1% of system RAM, minimum 4MB, maximum 64MB
	workMemMB := systemRAM / 100

	// Ensure minimum of 4MB
	if workMemMB < 4 {
		workMemMB = 4
	}

	// Cap at 64MB for safety
	if workMemMB > 64 {
		workMemMB = 64
	}

	return fmt.Sprintf("%dMB", workMemMB)
}

// calculateMaintenanceWorkMem calculates optimal maintenance work memory
func (mto *MemoryTuningOptimizer) calculateMaintenanceWorkMem(systemRAM int64) string {
	// For maintenance operations like VACUUM, CREATE INDEX, etc.
	// Base calculation: 5% of system RAM, minimum 64MB, maximum 1GB

	maintenanceWorkMemMB := systemRAM / 20

	// Ensure minimum of 64MB
	if maintenanceWorkMemMB < 64 {
		maintenanceWorkMemMB = 64
	}

	// Cap at 1GB for safety
	if maintenanceWorkMemMB > 1024 {
		maintenanceWorkMemMB = 1024
	}

	return fmt.Sprintf("%dMB", maintenanceWorkMemMB)
}

// ApplyMemoryConfiguration applies the optimized memory configuration to PostgreSQL
func (mto *MemoryTuningOptimizer) ApplyMemoryConfiguration(ctx context.Context, config *MemoryTuningConfig) error {
	mto.logger.Printf("Applying memory configuration: %+v", config)

	// Apply core memory settings
	if err := mto.setPostgreSQLSetting(ctx, "shared_buffers", config.SharedBuffers); err != nil {
		return fmt.Errorf("failed to set shared_buffers: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "effective_cache_size", config.EffectiveCacheSize); err != nil {
		return fmt.Errorf("failed to set effective_cache_size: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "work_mem", config.WorkMem); err != nil {
		return fmt.Errorf("failed to set work_mem: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "maintenance_work_mem", config.MaintenanceWorkMem); err != nil {
		return fmt.Errorf("failed to set maintenance_work_mem: %w", err)
	}

	// Apply query optimization settings
	if err := mto.setPostgreSQLSetting(ctx, "random_page_cost", fmt.Sprintf("%.1f", config.RandomPageCost)); err != nil {
		return fmt.Errorf("failed to set random_page_cost: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "effective_io_concurrency", fmt.Sprintf("%d", config.EffectiveIOConcurrency)); err != nil {
		return fmt.Errorf("failed to set effective_io_concurrency: %w", err)
	}

	// Apply WAL and checkpoint settings
	if err := mto.setPostgreSQLSetting(ctx, "wal_buffers", config.WALBuffers); err != nil {
		return fmt.Errorf("failed to set wal_buffers: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "max_wal_size", config.MaxWALSize); err != nil {
		return fmt.Errorf("failed to set max_wal_size: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "min_wal_size", config.MinWALSize); err != nil {
		return fmt.Errorf("failed to set min_wal_size: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "checkpoint_completion_target", fmt.Sprintf("%.1f", config.CheckpointCompletionTarget)); err != nil {
		return fmt.Errorf("failed to set checkpoint_completion_target: %w", err)
	}

	// Apply parallel query settings
	if err := mto.setPostgreSQLSetting(ctx, "max_parallel_workers_per_gather", fmt.Sprintf("%d", config.MaxParallelWorkersPerGather)); err != nil {
		return fmt.Errorf("failed to set max_parallel_workers_per_gather: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "max_parallel_workers", fmt.Sprintf("%d", config.MaxParallelWorkers)); err != nil {
		return fmt.Errorf("failed to set max_parallel_workers: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "parallel_tuple_cost", fmt.Sprintf("%.1f", config.ParallelTupleCost)); err != nil {
		return fmt.Errorf("failed to set parallel_tuple_cost: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "parallel_setup_cost", fmt.Sprintf("%.1f", config.ParallelSetupCost)); err != nil {
		return fmt.Errorf("failed to set parallel_setup_cost: %w", err)
	}

	// Apply statistics and planning settings
	if err := mto.setPostgreSQLSetting(ctx, "default_statistics_target", fmt.Sprintf("%d", config.DefaultStatisticsTarget)); err != nil {
		return fmt.Errorf("failed to set default_statistics_target: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "cpu_tuple_cost", fmt.Sprintf("%.3f", config.CPUTupleCost)); err != nil {
		return fmt.Errorf("failed to set cpu_tuple_cost: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "cpu_index_tuple_cost", fmt.Sprintf("%.3f", config.CPUIndexTupleCost)); err != nil {
		return fmt.Errorf("failed to set cpu_index_tuple_cost: %w", err)
	}

	if err := mto.setPostgreSQLSetting(ctx, "cpu_operator_cost", fmt.Sprintf("%.4f", config.CPUOperatorCost)); err != nil {
		return fmt.Errorf("failed to set cpu_operator_cost: %w", err)
	}

	// Reload configuration
	if err := mto.reloadPostgreSQLConfiguration(ctx); err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	mto.logger.Printf("Memory configuration applied successfully")
	return nil
}

// setPostgreSQLSetting sets a PostgreSQL configuration parameter
func (mto *MemoryTuningOptimizer) setPostgreSQLSetting(ctx context.Context, setting, value string) error {
	query := fmt.Sprintf("ALTER SYSTEM SET %s = %s", setting, value)

	_, err := mto.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute ALTER SYSTEM SET %s = %s: %w", setting, value, err)
	}

	mto.logger.Printf("Set %s = %s", setting, value)
	return nil
}

// reloadPostgreSQLConfiguration reloads PostgreSQL configuration
func (mto *MemoryTuningOptimizer) reloadPostgreSQLConfiguration(ctx context.Context) error {
	_, err := mto.db.ExecContext(ctx, "SELECT pg_reload_conf()")
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	mto.logger.Printf("PostgreSQL configuration reloaded")
	return nil
}

// ValidateMemoryConfiguration validates that memory configuration is properly applied
func (mto *MemoryTuningOptimizer) ValidateMemoryConfiguration(ctx context.Context, config *MemoryTuningConfig) error {
	// Check key settings
	settings := map[string]string{
		"shared_buffers":           config.SharedBuffers,
		"effective_cache_size":     config.EffectiveCacheSize,
		"work_mem":                 config.WorkMem,
		"maintenance_work_mem":     config.MaintenanceWorkMem,
		"random_page_cost":         fmt.Sprintf("%.1f", config.RandomPageCost),
		"effective_io_concurrency": fmt.Sprintf("%d", config.EffectiveIOConcurrency),
	}

	for setting, expectedValue := range settings {
		actualValue, err := mto.getPostgreSQLSetting(ctx, setting)
		if err != nil {
			return fmt.Errorf("failed to get setting %s: %w", setting, err)
		}

		if !mto.settingsMatch(setting, actualValue, expectedValue) {
			return fmt.Errorf("setting %s mismatch: expected %s, got %s", setting, expectedValue, actualValue)
		}
	}

	mto.logger.Printf("Memory configuration validation passed")
	return nil
}

// getPostgreSQLSetting gets a PostgreSQL configuration parameter value
func (mto *MemoryTuningOptimizer) getPostgreSQLSetting(ctx context.Context, setting string) (string, error) {
	query := "SELECT setting FROM pg_settings WHERE name = $1"

	var value string
	err := mto.db.QueryRowContext(ctx, query, setting).Scan(&value)
	if err != nil {
		return "", fmt.Errorf("failed to query setting %s: %w", setting, err)
	}

	return value, nil
}

// settingsMatch compares setting values, handling different formats
func (mto *MemoryTuningOptimizer) settingsMatch(setting, actual, expected string) bool {
	// Normalize values for comparison
	actual = strings.TrimSpace(strings.ToLower(actual))
	expected = strings.TrimSpace(strings.ToLower(expected))

	// Handle numeric settings
	if mto.isNumericSetting(setting) {
		actualNum, err1 := strconv.ParseFloat(actual, 64)
		expectedNum, err2 := strconv.ParseFloat(expected, 64)

		if err1 == nil && err2 == nil {
			// Allow small differences for floating point values
			diff := actualNum - expectedNum
			if diff < 0 {
				diff = -diff
			}
			return diff < 0.01
		}
	}

	// Handle memory settings (convert to same units)
	if mto.isMemorySetting(setting) {
		actualMB := mto.convertToMB(actual)
		expectedMB := mto.convertToMB(expected)
		return actualMB == expectedMB
	}

	// Direct string comparison
	return actual == expected
}

// isNumericSetting checks if a setting is numeric
func (mto *MemoryTuningOptimizer) isNumericSetting(setting string) bool {
	numericSettings := []string{
		"random_page_cost", "effective_io_concurrency", "checkpoint_completion_target",
		"max_parallel_workers_per_gather", "max_parallel_workers", "parallel_tuple_cost",
		"parallel_setup_cost", "default_statistics_target", "cpu_tuple_cost",
		"cpu_index_tuple_cost", "cpu_operator_cost",
	}

	for _, s := range numericSettings {
		if s == setting {
			return true
		}
	}
	return false
}

// isMemorySetting checks if a setting is a memory setting
func (mto *MemoryTuningOptimizer) isMemorySetting(setting string) bool {
	memorySettings := []string{
		"shared_buffers", "effective_cache_size", "work_mem", "maintenance_work_mem",
		"wal_buffers", "max_wal_size", "min_wal_size",
	}

	for _, s := range memorySettings {
		if s == setting {
			return true
		}
	}
	return false
}

// convertToMB converts memory values to MB for comparison
func (mto *MemoryTuningOptimizer) convertToMB(value string) int64 {
	value = strings.TrimSpace(strings.ToLower(value))

	// Remove common suffixes and convert to MB
	if strings.HasSuffix(value, "kb") {
		num, _ := strconv.ParseInt(strings.TrimSuffix(value, "kb"), 10, 64)
		return num / 1024
	} else if strings.HasSuffix(value, "mb") {
		num, _ := strconv.ParseInt(strings.TrimSuffix(value, "mb"), 10, 64)
		return num
	} else if strings.HasSuffix(value, "gb") {
		num, _ := strconv.ParseInt(strings.TrimSuffix(value, "gb"), 10, 64)
		return num * 1024
	} else if strings.HasSuffix(value, "b") {
		num, _ := strconv.ParseInt(strings.TrimSuffix(value, "b"), 10, 64)
		return num / (1024 * 1024)
	}

	// Try to parse as number (assume MB)
	num, _ := strconv.ParseInt(value, 10, 64)
	return num
}

// GetMemoryUsageStats retrieves current memory usage statistics
func (mto *MemoryTuningOptimizer) GetMemoryUsageStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get shared buffer usage
	var sharedBuffersHit, sharedBuffersRead int64
	err := mto.db.QueryRowContext(ctx, `
		SELECT 
			SUM(blks_hit) as shared_buffers_hit,
			SUM(blks_read) as shared_buffers_read
		FROM pg_stat_database
	`).Scan(&sharedBuffersHit, &sharedBuffersRead)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared buffer stats: %w", err)
	}

	// Calculate hit ratio
	totalReads := sharedBuffersHit + sharedBuffersRead
	hitRatio := float64(0)
	if totalReads > 0 {
		hitRatio = float64(sharedBuffersHit) / float64(totalReads) * 100
	}

	stats["shared_buffers_hit_ratio"] = hitRatio
	stats["shared_buffers_hit"] = sharedBuffersHit
	stats["shared_buffers_read"] = sharedBuffersRead

	// Get current memory settings
	memorySettings := []string{
		"shared_buffers", "effective_cache_size", "work_mem", "maintenance_work_mem",
	}

	for _, setting := range memorySettings {
		value, err := mto.getPostgreSQLSetting(ctx, setting)
		if err != nil {
			mto.logger.Printf("Failed to get setting %s: %v", setting, err)
			continue
		}
		stats[setting] = value
	}

	return stats, nil
}

// OptimizeMemoryForWorkload optimizes memory settings for specific workloads
func (mto *MemoryTuningOptimizer) OptimizeMemoryForWorkload(ctx context.Context, workload string) error {
	config := mto.GetOptimizedMemoryConfig()

	// Adjust settings based on workload
	switch workload {
	case "classification":
		// Increase work_mem for complex classification queries
		config.WorkMem = "32MB"
		config.DefaultStatisticsTarget = 200
		mto.logger.Printf("Optimized memory for classification workload")

	case "risk_assessment":
		// Increase work_mem for JSONB operations
		config.WorkMem = "24MB"
		config.DefaultStatisticsTarget = 150
		mto.logger.Printf("Optimized memory for risk assessment workload")

	case "mixed":
		// Balanced settings for mixed workload
		config.WorkMem = "16MB"
		config.DefaultStatisticsTarget = 100
		mto.logger.Printf("Optimized memory for mixed workload")

	default:
		mto.logger.Printf("Unknown workload type: %s, using default settings", workload)
	}

	// Apply the optimized configuration
	return mto.ApplyMemoryConfiguration(ctx, config)
}
