package enrichment

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// DataFreshnessTracker provides advanced data freshness and update frequency tracking
type DataFreshnessTracker struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *DataFreshnessConfig

	// Tracking data
	mu               sync.RWMutex
	freshnessHistory map[string]*FreshnessRecord
	updatePatterns   map[string]*UpdatePattern
	stalenessAlerts  map[string]*StalenessAlert
	lastCleanup      time.Time
}

// DataFreshnessConfig contains configuration for freshness tracking
type DataFreshnessConfig struct {
	// Tracking settings
	EnableTracking          bool `json:"enable_tracking"`           // Enable freshness tracking
	EnableUpdatePatterns    bool `json:"enable_update_patterns"`    // Enable update pattern analysis
	EnableStalenessAlerts   bool `json:"enable_staleness_alerts"`   // Enable staleness alerts
	EnablePredictiveScoring bool `json:"enable_predictive_scoring"` // Enable predictive freshness scoring

	// Thresholds
	StalenessThreshold         time.Duration `json:"staleness_threshold"`          // When data is considered stale
	CriticalStalenessThreshold time.Duration `json:"critical_staleness_threshold"` // Critical staleness threshold
	UpdateFrequencyThreshold   time.Duration `json:"update_frequency_threshold"`   // Expected update frequency

	// Scoring weights
	AgeWeight             float64 `json:"age_weight"`              // Weight for data age
	UpdateFrequencyWeight float64 `json:"update_frequency_weight"` // Weight for update frequency
	PredictiveWeight      float64 `json:"predictive_weight"`       // Weight for predictive scoring
	ConsistencyWeight     float64 `json:"consistency_weight"`      // Weight for update consistency

	// History settings
	MaxHistorySize         int           `json:"max_history_size"`         // Maximum history records per data source
	HistoryRetentionPeriod time.Duration `json:"history_retention_period"` // How long to keep history
	CleanupInterval        time.Duration `json:"cleanup_interval"`         // How often to cleanup old records

	// Alert settings
	AlertCooldownPeriod time.Duration `json:"alert_cooldown_period"` // Minimum time between alerts
	MaxAlertsPerSource  int           `json:"max_alerts_per_source"` // Maximum alerts per data source
}

// FreshnessRecord represents a single freshness tracking record
type FreshnessRecord struct {
	DataID             string                 `json:"data_id"`
	DataType           string                 `json:"data_type"`
	Source             string                 `json:"source"`
	Timestamp          time.Time              `json:"timestamp"`
	Age                time.Duration          `json:"age"`
	FreshnessScore     float64                `json:"freshness_score"`
	UpdateFrequency    time.Duration          `json:"update_frequency"`
	IsStale            bool                   `json:"is_stale"`
	IsCriticalStale    bool                   `json:"is_critical_stale"`
	LastUpdated        time.Time              `json:"last_updated"`
	NextExpectedUpdate time.Time              `json:"next_expected_update"`
	UpdateCount        int                    `json:"update_count"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// UpdatePattern represents the update pattern for a data source
type UpdatePattern struct {
	Source             string        `json:"source"`
	DataType           string        `json:"data_type"`
	AverageInterval    time.Duration `json:"average_interval"`
	MinInterval        time.Duration `json:"min_interval"`
	MaxInterval        time.Duration `json:"max_interval"`
	StandardDeviation  time.Duration `json:"standard_deviation"`
	UpdateCount        int           `json:"update_count"`
	LastUpdate         time.Time     `json:"last_update"`
	NextExpectedUpdate time.Time     `json:"next_expected_update"`
	ConsistencyScore   float64       `json:"consistency_score"`
	PatternType        string        `json:"pattern_type"` // "regular", "irregular", "sporadic"
	Confidence         float64       `json:"confidence"`
}

// StalenessAlert represents a staleness alert
type StalenessAlert struct {
	DataID        string                 `json:"data_id"`
	Source        string                 `json:"source"`
	DataType      string                 `json:"data_type"`
	AlertType     string                 `json:"alert_type"` // "staleness", "critical_staleness", "update_frequency"
	Severity      string                 `json:"severity"`   // "low", "medium", "high", "critical"
	Message       string                 `json:"message"`
	CreatedAt     time.Time              `json:"created_at"`
	LastTriggered time.Time              `json:"last_triggered"`
	TriggerCount  int                    `json:"trigger_count"`
	IsActive      bool                   `json:"is_active"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// FreshnessAnalysisResult contains comprehensive freshness analysis
type FreshnessAnalysisResult struct {
	// Current freshness
	CurrentFreshness *FreshnessRecord `json:"current_freshness"`
	OverallScore     float64          `json:"overall_score"`
	FreshnessLevel   string           `json:"freshness_level"` // "fresh", "aging", "stale", "critical"

	// Update patterns
	UpdatePattern        *UpdatePattern `json:"update_pattern"`
	UpdateFrequencyScore float64        `json:"update_frequency_score"`
	ConsistencyScore     float64        `json:"consistency_score"`

	// Predictive analysis
	PredictiveScore      float64   `json:"predictive_score"`
	NextUpdatePrediction time.Time `json:"next_update_prediction"`
	StalenessRisk        float64   `json:"staleness_risk"`

	// Alerts and recommendations
	ActiveAlerts    []*StalenessAlert `json:"active_alerts"`
	Recommendations []string          `json:"recommendations"`
	PriorityActions []string          `json:"priority_actions"`

	// Historical analysis
	HistoricalTrend []FreshnessRecord `json:"historical_trend"`
	TrendDirection  string            `json:"trend_direction"` // "improving", "stable", "declining"
	TrendConfidence float64           `json:"trend_confidence"`

	// Metadata
	AnalyzedAt     time.Time     `json:"analyzed_at"`
	ProcessingTime time.Duration `json:"processing_time"`
	DataPoints     int           `json:"data_points"`
}

// NewDataFreshnessTracker creates a new data freshness tracker
func NewDataFreshnessTracker(logger *zap.Logger, config *DataFreshnessConfig) *DataFreshnessTracker {
	if config == nil {
		config = getDefaultDataFreshnessConfig()
	}

	return &DataFreshnessTracker{
		logger:           logger,
		tracer:           trace.NewNoopTracerProvider().Tracer("data_freshness_tracker"),
		config:           config,
		freshnessHistory: make(map[string]*FreshnessRecord),
		updatePatterns:   make(map[string]*UpdatePattern),
		stalenessAlerts:  make(map[string]*StalenessAlert),
		lastCleanup:      time.Now(),
	}
}

// TrackFreshness tracks the freshness of a data item
func (dft *DataFreshnessTracker) TrackFreshness(ctx context.Context, data interface{}, dataID, dataType, source string) (*FreshnessRecord, error) {
	ctx, span := dft.tracer.Start(ctx, "data_freshness_tracker.track",
		trace.WithAttributes(
			attribute.String("data_id", dataID),
			attribute.String("data_type", dataType),
			attribute.String("source", source),
		))
	defer span.End()

	dft.logger.Info("Tracking data freshness",
		zap.String("data_id", dataID),
		zap.String("data_type", dataType),
		zap.String("source", source))

	// Extract timestamp from data
	timestamp := dft.extractTimestamp(data)
	if timestamp.IsZero() {
		timestamp = time.Now() // Use current time if no timestamp available
	}

	// Calculate age
	age := time.Since(timestamp)

	// Calculate freshness score
	freshnessScore := dft.calculateFreshnessScore(age)

	// Check staleness
	isStale := age > dft.config.StalenessThreshold
	isCriticalStale := age > dft.config.CriticalStalenessThreshold

	// Create freshness record
	record := &FreshnessRecord{
		DataID:          dataID,
		DataType:        dataType,
		Source:          source,
		Timestamp:       timestamp,
		Age:             age,
		FreshnessScore:  freshnessScore,
		IsStale:         isStale,
		IsCriticalStale: isCriticalStale,
		LastUpdated:     time.Now(),
		UpdateCount:     1,
		Metadata:        make(map[string]interface{}),
	}

	// Update history
	dft.mu.Lock()
	defer dft.mu.Unlock()

	// Check if we have previous record for this data
	key := fmt.Sprintf("%s:%s:%s", dataID, dataType, source)
	if existingRecord, exists := dft.freshnessHistory[key]; exists {
		// Calculate update frequency
		updateInterval := time.Since(existingRecord.LastUpdated)
		record.UpdateFrequency = updateInterval
		record.UpdateCount = existingRecord.UpdateCount + 1
		record.NextExpectedUpdate = dft.predictNextUpdate(existingRecord, updateInterval)
	} else {
		// First time tracking this data
		record.UpdateFrequency = 0
		record.NextExpectedUpdate = time.Now().Add(dft.config.UpdateFrequencyThreshold)
	}

	// Store the record
	dft.freshnessHistory[key] = record

	// Update patterns if enabled
	if dft.config.EnableUpdatePatterns {
		dft.updatePatternAnalysis(key, record)
	}

	// Check for staleness alerts if enabled
	if dft.config.EnableStalenessAlerts {
		dft.checkStalenessAlerts(record)
	}

	// Cleanup old records periodically
	dft.cleanupIfNeeded()

	dft.logger.Info("Freshness tracking completed",
		zap.String("data_id", dataID),
		zap.Duration("age", age),
		zap.Float64("freshness_score", freshnessScore),
		zap.Bool("is_stale", isStale))

	return record, nil
}

// AnalyzeFreshness performs comprehensive freshness analysis
func (dft *DataFreshnessTracker) AnalyzeFreshness(ctx context.Context, dataID, dataType, source string) (*FreshnessAnalysisResult, error) {
	ctx, span := dft.tracer.Start(ctx, "data_freshness_tracker.analyze",
		trace.WithAttributes(
			attribute.String("data_id", dataID),
			attribute.String("data_type", dataType),
			attribute.String("source", source),
		))
	defer span.End()

	startTime := time.Now()

	dft.logger.Info("Starting freshness analysis",
		zap.String("data_id", dataID),
		zap.String("data_type", dataType),
		zap.String("source", source))

	result := &FreshnessAnalysisResult{
		ActiveAlerts:    []*StalenessAlert{},
		Recommendations: []string{},
		PriorityActions: []string{},
		HistoricalTrend: []FreshnessRecord{},
		AnalyzedAt:      time.Now(),
	}

	dft.mu.RLock()
	defer dft.mu.RUnlock()

	// Get current freshness record
	key := fmt.Sprintf("%s:%s:%s", dataID, dataType, source)
	if record, exists := dft.freshnessHistory[key]; exists {
		result.CurrentFreshness = record
		result.DataPoints = 1
	} else {
		// No history available
		result.CurrentFreshness = &FreshnessRecord{
			DataID:         dataID,
			DataType:       dataType,
			Source:         source,
			FreshnessScore: 0.5, // Default score
		}
	}

	// Get update pattern
	if pattern, exists := dft.updatePatterns[fmt.Sprintf("%s:%s", dataType, source)]; exists {
		result.UpdatePattern = pattern
		result.UpdateFrequencyScore = dft.calculateUpdateFrequencyScore(pattern)
		result.ConsistencyScore = pattern.ConsistencyScore
		result.NextUpdatePrediction = pattern.NextExpectedUpdate
	}

	// Calculate overall score
	result.OverallScore = dft.calculateOverallFreshnessScore(result)

	// Determine freshness level
	result.FreshnessLevel = dft.determineFreshnessLevel(result.OverallScore)

	// Calculate predictive score if enabled
	if dft.config.EnablePredictiveScoring {
		result.PredictiveScore = dft.calculatePredictiveScore(result)
		result.StalenessRisk = dft.calculateStalenessRisk(result)
	}

	// Get active alerts
	result.ActiveAlerts = dft.getActiveAlerts(dataID, dataType, source)

	// Generate recommendations
	result.Recommendations = dft.generateRecommendations(result)
	result.PriorityActions = dft.generatePriorityActions(result)

	// Analyze historical trend
	result.HistoricalTrend = dft.getHistoricalTrend(key)
	result.TrendDirection = dft.analyzeTrendDirection(result.HistoricalTrend)
	result.TrendConfidence = dft.calculateTrendConfidence(result.HistoricalTrend)

	result.ProcessingTime = time.Since(startTime)

	dft.logger.Info("Freshness analysis completed",
		zap.String("data_id", dataID),
		zap.Float64("overall_score", result.OverallScore),
		zap.String("freshness_level", result.FreshnessLevel),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// GetStalenessAlerts returns all active staleness alerts
func (dft *DataFreshnessTracker) GetStalenessAlerts(ctx context.Context) ([]*StalenessAlert, error) {
	dft.mu.RLock()
	defer dft.mu.RUnlock()

	alerts := []*StalenessAlert{}
	for _, alert := range dft.stalenessAlerts {
		if alert.IsActive {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// GetUpdatePatterns returns all update patterns
func (dft *DataFreshnessTracker) GetUpdatePatterns(ctx context.Context) ([]*UpdatePattern, error) {
	dft.mu.RLock()
	defer dft.mu.RUnlock()

	patterns := []*UpdatePattern{}
	for _, pattern := range dft.updatePatterns {
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// Helper methods

func (dft *DataFreshnessTracker) extractTimestamp(data interface{}) time.Time {
	// Try to extract timestamp from various data structures
	if data == nil {
		return time.Time{}
	}

	// Use reflection to check for common timestamp fields
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		// Check for common timestamp field names
		timestampFields := []string{"UpdatedAt", "updated_at", "LastUpdated", "last_updated", "ModifiedAt", "modified_at", "Timestamp", "timestamp"}

		for _, fieldName := range timestampFields {
			if field := v.FieldByName(fieldName); field.IsValid() && !field.IsZero() {
				if field.Type() == reflect.TypeOf(time.Time{}) {
					return field.Interface().(time.Time)
				}
			}
		}

		// Check for CreatedAt as fallback
		if field := v.FieldByName("CreatedAt"); field.IsValid() && !field.IsZero() {
			if field.Type() == reflect.TypeOf(time.Time{}) {
				return field.Interface().(time.Time)
			}
		}
	}

	// If no timestamp found, return zero time to use current time
	return time.Time{}
}

func (dft *DataFreshnessTracker) calculateFreshnessScore(age time.Duration) float64 {
	thresholdHours := dft.config.StalenessThreshold.Hours()
	ageHours := age.Hours()

	if ageHours <= thresholdHours {
		return 1.0 // Fresh data
	}

	// Exponential decay for older data
	decay := math.Exp(-ageHours / thresholdHours)
	return math.Max(0.1, decay) // Minimum score of 0.1
}

func (dft *DataFreshnessTracker) predictNextUpdate(record *FreshnessRecord, updateInterval time.Duration) time.Time {
	if updateInterval == 0 {
		return time.Now().Add(dft.config.UpdateFrequencyThreshold)
	}

	// Simple prediction based on last update interval
	return time.Now().Add(updateInterval)
}

func (dft *DataFreshnessTracker) updatePatternAnalysis(key string, record *FreshnessRecord) {
	patternKey := fmt.Sprintf("%s:%s", record.DataType, record.Source)

	if pattern, exists := dft.updatePatterns[patternKey]; exists {
		// Update existing pattern
		pattern.UpdateCount++
		pattern.LastUpdate = time.Now()

		// Update average interval
		if record.UpdateFrequency > 0 {
			oldAvg := pattern.AverageInterval
			pattern.AverageInterval = (oldAvg*time.Duration(pattern.UpdateCount-1) + record.UpdateFrequency) / time.Duration(pattern.UpdateCount)
		}

		// Update min/max intervals
		if record.UpdateFrequency < pattern.MinInterval || pattern.MinInterval == 0 {
			pattern.MinInterval = record.UpdateFrequency
		}
		if record.UpdateFrequency > pattern.MaxInterval {
			pattern.MaxInterval = record.UpdateFrequency
		}

		// Update consistency score
		pattern.ConsistencyScore = dft.calculateConsistencyScore(pattern)
		pattern.NextExpectedUpdate = dft.predictNextUpdate(record, pattern.AverageInterval)
	} else {
		// Create new pattern
		pattern := &UpdatePattern{
			Source:           record.Source,
			DataType:         record.DataType,
			AverageInterval:  record.UpdateFrequency,
			MinInterval:      record.UpdateFrequency,
			MaxInterval:      record.UpdateFrequency,
			UpdateCount:      1,
			LastUpdate:       time.Now(),
			ConsistencyScore: 1.0,
			PatternType:      "regular",
			Confidence:       0.5,
		}
		dft.updatePatterns[patternKey] = pattern
	}
}

func (dft *DataFreshnessTracker) checkStalenessAlerts(record *FreshnessRecord) {
	alertKey := fmt.Sprintf("%s:%s:%s", record.DataID, record.DataType, record.Source)

	// Check if alert already exists and is within cooldown
	if existingAlert, exists := dft.stalenessAlerts[alertKey]; exists {
		if time.Since(existingAlert.LastTriggered) < dft.config.AlertCooldownPeriod {
			return // Still in cooldown
		}
	}

	var alert *StalenessAlert

	if record.IsCriticalStale {
		alert = &StalenessAlert{
			DataID:        record.DataID,
			Source:        record.Source,
			DataType:      record.DataType,
			AlertType:     "critical_staleness",
			Severity:      "critical",
			Message:       fmt.Sprintf("Data is critically stale (age: %v)", record.Age),
			CreatedAt:     time.Now(),
			LastTriggered: time.Now(),
			TriggerCount:  1,
			IsActive:      true,
			Metadata:      make(map[string]interface{}),
		}
	} else if record.IsStale {
		alert = &StalenessAlert{
			DataID:        record.DataID,
			Source:        record.Source,
			DataType:      record.DataType,
			AlertType:     "staleness",
			Severity:      "high",
			Message:       fmt.Sprintf("Data is stale (age: %v)", record.Age),
			CreatedAt:     time.Now(),
			LastTriggered: time.Now(),
			TriggerCount:  1,
			IsActive:      true,
			Metadata:      make(map[string]interface{}),
		}
	}

	if alert != nil {
		if existingAlert, exists := dft.stalenessAlerts[alertKey]; exists {
			alert.TriggerCount = existingAlert.TriggerCount + 1
		}
		dft.stalenessAlerts[alertKey] = alert
	}
}

func (dft *DataFreshnessTracker) cleanupIfNeeded() {
	if time.Since(dft.lastCleanup) < dft.config.CleanupInterval {
		return
	}

	dft.lastCleanup = time.Now()
	cutoff := time.Now().Add(-dft.config.HistoryRetentionPeriod)

	// Cleanup old freshness records
	for key, record := range dft.freshnessHistory {
		if record.LastUpdated.Before(cutoff) {
			delete(dft.freshnessHistory, key)
		}
	}

	// Cleanup old alerts
	for key, alert := range dft.stalenessAlerts {
		if alert.LastTriggered.Before(cutoff) {
			delete(dft.stalenessAlerts, key)
		}
	}
}

func (dft *DataFreshnessTracker) calculateUpdateFrequencyScore(pattern *UpdatePattern) float64 {
	if pattern == nil {
		return 0.5
	}

	// Score based on how close the actual interval is to expected
	expectedInterval := dft.config.UpdateFrequencyThreshold
	actualInterval := pattern.AverageInterval

	if actualInterval == 0 {
		return 0.5
	}

	ratio := float64(expectedInterval) / float64(actualInterval)
	if ratio > 1 {
		ratio = 1 / ratio // More frequent updates are better
	}

	return math.Min(1.0, ratio)
}

func (dft *DataFreshnessTracker) calculateConsistencyScore(pattern *UpdatePattern) float64 {
	if pattern.UpdateCount < 2 {
		return 1.0
	}

	// Calculate coefficient of variation (lower is more consistent)
	mean := float64(pattern.AverageInterval)
	stdDev := float64(pattern.StandardDeviation)

	if mean == 0 {
		return 1.0
	}

	cv := stdDev / mean
	consistency := math.Max(0.0, 1.0-cv)

	return consistency
}

func (dft *DataFreshnessTracker) calculateOverallFreshnessScore(result *FreshnessAnalysisResult) float64 {
	score := 0.0
	totalWeight := 0.0

	// Current freshness score
	if result.CurrentFreshness != nil {
		score += result.CurrentFreshness.FreshnessScore * dft.config.AgeWeight
		totalWeight += dft.config.AgeWeight
	}

	// Update frequency score
	score += result.UpdateFrequencyScore * dft.config.UpdateFrequencyWeight
	totalWeight += dft.config.UpdateFrequencyWeight

	// Consistency score
	score += result.ConsistencyScore * dft.config.ConsistencyWeight
	totalWeight += dft.config.ConsistencyWeight

	if totalWeight == 0 {
		return 0.5
	}

	return score / totalWeight
}

func (dft *DataFreshnessTracker) determineFreshnessLevel(score float64) string {
	if score >= 0.8 {
		return "fresh"
	} else if score >= 0.6 {
		return "aging"
	} else if score >= 0.4 {
		return "stale"
	} else {
		return "critical"
	}
}

func (dft *DataFreshnessTracker) calculatePredictiveScore(result *FreshnessAnalysisResult) float64 {
	// Simple predictive scoring based on current trends
	score := result.OverallScore

	// Adjust based on update pattern consistency
	if result.UpdatePattern != nil {
		score = (score + result.UpdatePattern.ConsistencyScore) / 2.0
	}

	// Adjust based on trend direction
	if result.TrendDirection == "improving" {
		score += 0.1
	} else if result.TrendDirection == "declining" {
		score -= 0.1
	}

	return math.Max(0.0, math.Min(1.0, score))
}

func (dft *DataFreshnessTracker) calculateStalenessRisk(result *FreshnessAnalysisResult) float64 {
	risk := 0.0

	// Base risk from current freshness
	if result.CurrentFreshness != nil {
		if result.CurrentFreshness.IsCriticalStale {
			risk += 0.8
		} else if result.CurrentFreshness.IsStale {
			risk += 0.5
		}
	}

	// Risk from update patterns
	if result.UpdatePattern != nil && result.UpdatePattern.ConsistencyScore < 0.5 {
		risk += 0.3
	}

	// Risk from trend
	if result.TrendDirection == "declining" {
		risk += 0.2
	}

	return math.Min(1.0, risk)
}

func (dft *DataFreshnessTracker) getActiveAlerts(dataID, dataType, source string) []*StalenessAlert {
	alerts := []*StalenessAlert{}

	for _, alert := range dft.stalenessAlerts {
		if alert.IsActive &&
			(alert.DataID == dataID || dataID == "") &&
			(alert.DataType == dataType || dataType == "") &&
			(alert.Source == source || source == "") {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

func (dft *DataFreshnessTracker) generateRecommendations(result *FreshnessAnalysisResult) []string {
	recommendations := []string{}

	if result.CurrentFreshness != nil {
		if result.CurrentFreshness.IsCriticalStale {
			recommendations = append(recommendations, "Immediate data refresh required")
		} else if result.CurrentFreshness.IsStale {
			recommendations = append(recommendations, "Data refresh recommended")
		}
	}

	if result.UpdatePattern != nil && result.UpdatePattern.ConsistencyScore < 0.7 {
		recommendations = append(recommendations, "Improve update frequency consistency")
	}

	if result.StalenessRisk > 0.7 {
		recommendations = append(recommendations, "High staleness risk - implement proactive refresh")
	}

	return recommendations
}

func (dft *DataFreshnessTracker) generatePriorityActions(result *FreshnessAnalysisResult) []string {
	actions := []string{}

	if result.CurrentFreshness != nil && result.CurrentFreshness.IsCriticalStale {
		actions = append(actions, "URGENT: Refresh data immediately")
	}

	if len(result.ActiveAlerts) > 0 {
		actions = append(actions, "Review and address active staleness alerts")
	}

	if result.UpdatePattern != nil && result.UpdatePattern.ConsistencyScore < 0.5 {
		actions = append(actions, "Investigate update frequency inconsistencies")
	}

	return actions
}

func (dft *DataFreshnessTracker) getHistoricalTrend(key string) []FreshnessRecord {
	// This would return historical records for trend analysis
	// For now, return empty slice
	return []FreshnessRecord{}
}

func (dft *DataFreshnessTracker) analyzeTrendDirection(trend []FreshnessRecord) string {
	if len(trend) < 2 {
		return "stable"
	}

	// Simple trend analysis
	firstScore := trend[0].FreshnessScore
	lastScore := trend[len(trend)-1].FreshnessScore

	if lastScore > firstScore+0.1 {
		return "improving"
	} else if lastScore < firstScore-0.1 {
		return "declining"
	} else {
		return "stable"
	}
}

func (dft *DataFreshnessTracker) calculateTrendConfidence(trend []FreshnessRecord) float64 {
	if len(trend) < 3 {
		return 0.3
	}

	// Calculate confidence based on trend consistency
	// For now, return a simple confidence score
	return math.Min(1.0, float64(len(trend))/10.0)
}

// getDefaultDataFreshnessConfig returns default configuration for data freshness tracking
func getDefaultDataFreshnessConfig() *DataFreshnessConfig {
	return &DataFreshnessConfig{
		// Enable all features
		EnableTracking:          true,
		EnableUpdatePatterns:    true,
		EnableStalenessAlerts:   true,
		EnablePredictiveScoring: true,

		// Thresholds
		StalenessThreshold:         24 * time.Hour,     // 24 hours
		CriticalStalenessThreshold: 7 * 24 * time.Hour, // 1 week
		UpdateFrequencyThreshold:   6 * time.Hour,      // 6 hours

		// Scoring weights
		AgeWeight:             0.4,
		UpdateFrequencyWeight: 0.3,
		PredictiveWeight:      0.2,
		ConsistencyWeight:     0.1,

		// History settings
		MaxHistorySize:         1000,
		HistoryRetentionPeriod: 30 * 24 * time.Hour, // 30 days
		CleanupInterval:        1 * time.Hour,       // 1 hour

		// Alert settings
		AlertCooldownPeriod: 1 * time.Hour, // 1 hour
		MaxAlertsPerSource:  10,
	}
}
