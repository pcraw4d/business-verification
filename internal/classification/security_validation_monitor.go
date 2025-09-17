package classification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AdvancedSecurityValidationMonitor provides comprehensive monitoring for security validation processes
type AdvancedSecurityValidationMonitor struct {
	logger             *zap.Logger
	config             *SecurityValidationConfig
	validationStats    map[string]*AdvancedSecurityValidationStats
	securityAlerts     []*AdvancedSecurityValidationAlert
	performanceMetrics []*AdvancedSecurityPerformanceMetric
	mu                 sync.RWMutex

	// Monitoring control
	stopCh           chan struct{}
	monitoringActive bool
	collectionTicker *time.Ticker
}

// SecurityValidationConfig holds configuration for security validation monitoring
type SecurityValidationConfig struct {
	Enabled                      bool          `json:"enabled"`
	CollectionInterval           time.Duration `json:"collection_interval"`
	SlowValidationThreshold      time.Duration `json:"slow_validation_threshold"`
	MaxValidationStats           int           `json:"max_validation_stats"`
	AlertingEnabled              bool          `json:"alerting_enabled"`
	TrackDataSourceValidation    bool          `json:"track_data_source_validation"`
	TrackWebsiteVerification     bool          `json:"track_website_verification"`
	TrackTrustScoreValidation    bool          `json:"track_trust_score_validation"`
	TrackSecurityChecks          bool          `json:"track_security_checks"`
	TrackComplianceValidation    bool          `json:"track_compliance_validation"`
	TrackEncryptionValidation    bool          `json:"track_encryption_validation"`
	TrackAccessControlValidation bool          `json:"track_access_control_validation"`
	TrackAuditLogValidation      bool          `json:"track_audit_log_validation"`
	TrackThreatDetection         bool          `json:"track_threat_detection"`
	TrackVulnerabilityScanning   bool          `json:"track_vulnerability_scanning"`
	TrackSecurityMetrics         bool          `json:"track_security_metrics"`
}

// AdvancedSecurityValidationStats represents statistics for security validation operations
type AdvancedSecurityValidationStats struct {
	ValidationID             string                 `json:"validation_id"`
	ValidationType           string                 `json:"validation_type"`
	ValidationName           string                 `json:"validation_name"`
	ExecutionCount           int64                  `json:"execution_count"`
	TotalExecutionTime       float64                `json:"total_execution_time_ms"`
	AverageExecutionTime     float64                `json:"average_execution_time_ms"`
	MinExecutionTime         float64                `json:"min_execution_time_ms"`
	MaxExecutionTime         float64                `json:"max_execution_time_ms"`
	P50ExecutionTime         float64                `json:"p50_execution_time_ms"`
	P95ExecutionTime         float64                `json:"p95_execution_time_ms"`
	P99ExecutionTime         float64                `json:"p99_execution_time_ms"`
	SuccessCount             int64                  `json:"success_count"`
	FailureCount             int64                  `json:"failure_count"`
	TimeoutCount             int64                  `json:"timeout_count"`
	ErrorCount               int64                  `json:"error_count"`
	SecurityViolationCount   int64                  `json:"security_violation_count"`
	ComplianceViolationCount int64                  `json:"compliance_violation_count"`
	ThreatDetectionCount     int64                  `json:"threat_detection_count"`
	VulnerabilityCount       int64                  `json:"vulnerability_count"`
	TrustScore               float64                `json:"trust_score"`
	ConfidenceLevel          float64                `json:"confidence_level"`
	RiskLevel                string                 `json:"risk_level"`
	PerformanceCategory      string                 `json:"performance_category"`
	SecurityCategory         string                 `json:"security_category"`
	LastExecuted             time.Time              `json:"last_executed"`
	FirstExecuted            time.Time              `json:"first_executed"`
	Metadata                 map[string]interface{} `json:"metadata,omitempty"`
}

// AdvancedSecurityValidationAlert represents a security validation alert
type AdvancedSecurityValidationAlert struct {
	ID              string                 `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	AlertType       string                 `json:"alert_type"`
	Severity        string                 `json:"severity"` // "low", "medium", "high", "critical"
	ValidationType  string                 `json:"validation_type"`
	ValidationID    string                 `json:"validation_id"`
	ValidationName  string                 `json:"validation_name"`
	Threshold       float64                `json:"threshold"`
	ActualValue     float64                `json:"actual_value"`
	Message         string                 `json:"message"`
	SecurityImpact  string                 `json:"security_impact"`
	Recommendations []string               `json:"recommendations"`
	Resolved        bool                   `json:"resolved"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AdvancedSecurityPerformanceMetric represents a security performance metric
type AdvancedSecurityPerformanceMetric struct {
	ID                      string                 `json:"id"`
	Timestamp               time.Time              `json:"timestamp"`
	MetricType              string                 `json:"metric_type"`
	ValidationType          string                 `json:"validation_type"`
	ValidationName          string                 `json:"validation_name"`
	ExecutionTimeMs         float64                `json:"execution_time_ms"`
	SuccessRate             float64                `json:"success_rate"`
	FailureRate             float64                `json:"failure_rate"`
	TimeoutRate             float64                `json:"timeout_rate"`
	ErrorRate               float64                `json:"error_rate"`
	SecurityViolationRate   float64                `json:"security_violation_rate"`
	ComplianceViolationRate float64                `json:"compliance_violation_rate"`
	ThreatDetectionRate     float64                `json:"threat_detection_rate"`
	VulnerabilityRate       float64                `json:"vulnerability_rate"`
	TrustScore              float64                `json:"trust_score"`
	ConfidenceLevel         float64                `json:"confidence_level"`
	RiskLevel               string                 `json:"risk_level"`
	PerformanceScore        float64                `json:"performance_score"`
	SecurityScore           float64                `json:"security_score"`
	OverallScore            float64                `json:"overall_score"`
	Metadata                map[string]interface{} `json:"metadata,omitempty"`
}

// AdvancedSecurityValidationResult represents the result of a security validation operation
type AdvancedSecurityValidationResult struct {
	ValidationID               string                 `json:"validation_id"`
	ValidationType             string                 `json:"validation_type"`
	ValidationName             string                 `json:"validation_name"`
	ExecutionTime              time.Duration          `json:"execution_time"`
	Success                    bool                   `json:"success"`
	Error                      error                  `json:"error,omitempty"`
	SecurityViolation          bool                   `json:"security_violation"`
	ComplianceViolation        bool                   `json:"compliance_violation"`
	ThreatDetected             bool                   `json:"threat_detected"`
	VulnerabilityFound         bool                   `json:"vulnerability_found"`
	TrustScore                 float64                `json:"trust_score"`
	ConfidenceLevel            float64                `json:"confidence_level"`
	RiskLevel                  string                 `json:"risk_level"`
	SecurityRecommendations    []string               `json:"security_recommendations"`
	PerformanceRecommendations []string               `json:"performance_recommendations"`
	Metadata                   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp                  time.Time              `json:"timestamp"`
}

// AdvancedSecuritySystemHealth represents overall security system health
type AdvancedSecuritySystemHealth struct {
	Timestamp               time.Time              `json:"timestamp"`
	OverallSecurityScore    float64                `json:"overall_security_score"`
	OverallPerformanceScore float64                `json:"overall_performance_score"`
	OverallRiskLevel        string                 `json:"overall_risk_level"`
	ActiveThreats           int                    `json:"active_threats"`
	SecurityViolations      int                    `json:"security_violations"`
	ComplianceViolations    int                    `json:"compliance_violations"`
	Vulnerabilities         int                    `json:"vulnerabilities"`
	SlowValidations         int                    `json:"slow_validations"`
	FailedValidations       int                    `json:"failed_validations"`
	HighRiskValidations     int                    `json:"high_risk_validations"`
	TrustScoreAverage       float64                `json:"trust_score_average"`
	ConfidenceLevelAverage  float64                `json:"confidence_level_average"`
	ValidationCount         int                    `json:"validation_count"`
	SuccessRate             float64                `json:"success_rate"`
	FailureRate             float64                `json:"failure_rate"`
	AverageExecutionTime    float64                `json:"average_execution_time_ms"`
	Metadata                map[string]interface{} `json:"metadata,omitempty"`
}

// NewAdvancedSecurityValidationMonitor creates a new security validation monitor
func NewAdvancedSecurityValidationMonitor(logger *zap.Logger, config *SecurityValidationConfig) *AdvancedSecurityValidationMonitor {
	if config == nil {
		config = DefaultSecurityValidationConfig()
	}

	monitor := &AdvancedSecurityValidationMonitor{
		logger:             logger,
		config:             config,
		validationStats:    make(map[string]*AdvancedSecurityValidationStats),
		securityAlerts:     make([]*AdvancedSecurityValidationAlert, 0),
		performanceMetrics: make([]*AdvancedSecurityPerformanceMetric, 0),
		stopCh:             make(chan struct{}),
		monitoringActive:   false,
	}

	// Start monitoring if enabled
	if config.Enabled {
		monitor.Start()
	}

	return monitor
}

// DefaultSecurityValidationConfig returns default configuration
func DefaultSecurityValidationConfig() *SecurityValidationConfig {
	return &SecurityValidationConfig{
		Enabled:                      true,
		CollectionInterval:           30 * time.Second,
		SlowValidationThreshold:      200 * time.Millisecond,
		MaxValidationStats:           1000,
		AlertingEnabled:              true,
		TrackDataSourceValidation:    true,
		TrackWebsiteVerification:     true,
		TrackTrustScoreValidation:    true,
		TrackSecurityChecks:          true,
		TrackComplianceValidation:    true,
		TrackEncryptionValidation:    true,
		TrackAccessControlValidation: true,
		TrackAuditLogValidation:      true,
		TrackThreatDetection:         true,
		TrackVulnerabilityScanning:   true,
		TrackSecurityMetrics:         true,
	}
}

// Start starts the security validation monitoring
func (svm *AdvancedSecurityValidationMonitor) Start() {
	svm.mu.Lock()
	defer svm.mu.Unlock()

	if svm.monitoringActive {
		return
	}

	svm.monitoringActive = true
	svm.collectionTicker = time.NewTicker(svm.config.CollectionInterval)

	// Start background monitoring
	go svm.monitoringLoop()

	svm.logger.Info("Security validation monitoring started",
		zap.Duration("collection_interval", svm.config.CollectionInterval),
		zap.Duration("slow_validation_threshold", svm.config.SlowValidationThreshold),
		zap.Bool("track_data_source_validation", svm.config.TrackDataSourceValidation),
		zap.Bool("track_website_verification", svm.config.TrackWebsiteVerification),
		zap.Bool("track_security_checks", svm.config.TrackSecurityChecks))
}

// Stop stops the security validation monitoring
func (svm *AdvancedSecurityValidationMonitor) Stop() {
	svm.mu.Lock()
	defer svm.mu.Unlock()

	if !svm.monitoringActive {
		return
	}

	svm.monitoringActive = false
	if svm.collectionTicker != nil {
		svm.collectionTicker.Stop()
	}
	close(svm.stopCh)

	svm.logger.Info("Security validation monitoring stopped")
}

// monitoringLoop runs the main monitoring loop
func (svm *AdvancedSecurityValidationMonitor) monitoringLoop() {
	for {
		select {
		case <-svm.collectionTicker.C:
			svm.collectSecurityMetrics()
			svm.analyzeSecurityPerformance()
			svm.generateSecurityAlerts()
		case <-svm.stopCh:
			return
		}
	}
}

// RecordSecurityValidation records a security validation operation
func (svm *AdvancedSecurityValidationMonitor) RecordSecurityValidation(
	ctx context.Context,
	result *AdvancedSecurityValidationResult,
) {
	if !svm.config.Enabled {
		return
	}

	svm.mu.Lock()
	defer svm.mu.Unlock()

	validationKey := svm.generateValidationKey(result.ValidationType, result.ValidationName)

	// Get or create validation stats
	stats, exists := svm.validationStats[validationKey]
	if !exists {
		stats = &AdvancedSecurityValidationStats{
			ValidationID:     result.ValidationID,
			ValidationType:   result.ValidationType,
			ValidationName:   result.ValidationName,
			FirstExecuted:    time.Now(),
			MinExecutionTime: float64(result.ExecutionTime.Milliseconds()),
			Metadata:         make(map[string]interface{}),
		}
		svm.validationStats[validationKey] = stats
	}

	// Update statistics
	stats.ExecutionCount++
	stats.TotalExecutionTime += float64(result.ExecutionTime.Milliseconds())
	stats.AverageExecutionTime = stats.TotalExecutionTime / float64(stats.ExecutionCount)
	stats.LastExecuted = time.Now()

	if float64(result.ExecutionTime.Milliseconds()) > stats.MaxExecutionTime {
		stats.MaxExecutionTime = float64(result.ExecutionTime.Milliseconds())
	}
	if float64(result.ExecutionTime.Milliseconds()) < stats.MinExecutionTime {
		stats.MinExecutionTime = float64(result.ExecutionTime.Milliseconds())
	}

	// Update success/failure counts
	if result.Success {
		stats.SuccessCount++
	} else {
		stats.FailureCount++
	}

	if result.Error != nil {
		stats.ErrorCount++
	}

	// Update security-specific counts
	if result.SecurityViolation {
		stats.SecurityViolationCount++
	}
	if result.ComplianceViolation {
		stats.ComplianceViolationCount++
	}
	if result.ThreatDetected {
		stats.ThreatDetectionCount++
	}
	if result.VulnerabilityFound {
		stats.VulnerabilityCount++
	}

	// Update trust and confidence scores
	stats.TrustScore = result.TrustScore
	stats.ConfidenceLevel = result.ConfidenceLevel
	stats.RiskLevel = result.RiskLevel

	// Calculate performance and security categories
	svm.calculateValidationCategories(stats)

	// Check for alerts
	if svm.config.AlertingEnabled {
		svm.checkSecurityAlerts(stats, result)
	}

	// Clean up old stats if needed
	if len(svm.validationStats) > svm.config.MaxValidationStats {
		svm.cleanupOldValidationStats()
	}
}

// calculateValidationCategories calculates performance and security categories
func (svm *AdvancedSecurityValidationMonitor) calculateValidationCategories(stats *AdvancedSecurityValidationStats) {
	// Calculate performance category
	if stats.AverageExecutionTime < float64(svm.config.SlowValidationThreshold.Milliseconds())/2 {
		stats.PerformanceCategory = "excellent"
	} else if stats.AverageExecutionTime < float64(svm.config.SlowValidationThreshold.Milliseconds()) {
		stats.PerformanceCategory = "good"
	} else if stats.AverageExecutionTime < float64(svm.config.SlowValidationThreshold.Milliseconds())*2 {
		stats.PerformanceCategory = "fair"
	} else if stats.AverageExecutionTime < float64(svm.config.SlowValidationThreshold.Milliseconds())*5 {
		stats.PerformanceCategory = "poor"
	} else {
		stats.PerformanceCategory = "critical"
	}

	// Calculate security category
	if stats.SecurityViolationCount == 0 && stats.ComplianceViolationCount == 0 &&
		stats.ThreatDetectionCount == 0 && stats.VulnerabilityCount == 0 {
		stats.SecurityCategory = "secure"
	} else if stats.SecurityViolationCount == 0 && stats.ComplianceViolationCount == 0 {
		stats.SecurityCategory = "monitored"
	} else if stats.SecurityViolationCount == 0 {
		stats.SecurityCategory = "compliant"
	} else {
		stats.SecurityCategory = "at_risk"
	}
}

// checkSecurityAlerts checks for security validation alerts
func (svm *AdvancedSecurityValidationMonitor) checkSecurityAlerts(
	stats *AdvancedSecurityValidationStats,
	result *AdvancedSecurityValidationResult,
) {
	var alerts []*AdvancedSecurityValidationAlert

	// Check for slow validations
	if stats.AverageExecutionTime > float64(svm.config.SlowValidationThreshold.Milliseconds()) {
		alerts = append(alerts, &AdvancedSecurityValidationAlert{
			ID:             fmt.Sprintf("slow_validation_%s_%d", stats.ValidationID, time.Now().Unix()),
			Timestamp:      time.Now(),
			AlertType:      "slow_validation",
			Severity:       svm.determineSecurityAlertSeverity(stats.AverageExecutionTime, float64(svm.config.SlowValidationThreshold.Milliseconds())),
			ValidationType: stats.ValidationType,
			ValidationID:   stats.ValidationID,
			ValidationName: stats.ValidationName,
			Threshold:      float64(svm.config.SlowValidationThreshold.Milliseconds()),
			ActualValue:    stats.AverageExecutionTime,
			Message:        fmt.Sprintf("Security validation %s is running slowly: %.2fms average", stats.ValidationName, stats.AverageExecutionTime),
			SecurityImpact: "performance_degradation",
			Recommendations: []string{
				"Optimize validation algorithm performance",
				"Consider caching validation results",
				"Review validation logic for efficiency improvements",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"execution_count":      stats.ExecutionCount,
				"performance_category": stats.PerformanceCategory,
			},
		})
	}

	// Check for high failure rate
	if stats.ExecutionCount > 10 && float64(stats.FailureCount)/float64(stats.ExecutionCount) > 0.1 {
		alerts = append(alerts, &AdvancedSecurityValidationAlert{
			ID:             fmt.Sprintf("high_failure_rate_%s_%d", stats.ValidationID, time.Now().Unix()),
			Timestamp:      time.Now(),
			AlertType:      "high_failure_rate",
			Severity:       "high",
			ValidationType: stats.ValidationType,
			ValidationID:   stats.ValidationID,
			ValidationName: stats.ValidationName,
			Threshold:      0.1,
			ActualValue:    float64(stats.FailureCount) / float64(stats.ExecutionCount),
			Message:        fmt.Sprintf("Security validation %s has high failure rate: %.2f%%", stats.ValidationName, float64(stats.FailureCount)/float64(stats.ExecutionCount)*100),
			SecurityImpact: "reliability_concern",
			Recommendations: []string{
				"Investigate validation logic for potential issues",
				"Review input data quality and validation criteria",
				"Consider adding additional error handling",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"execution_count": stats.ExecutionCount,
				"failure_count":   stats.FailureCount,
			},
		})
	}

	// Check for security violations
	if result.SecurityViolation {
		alerts = append(alerts, &AdvancedSecurityValidationAlert{
			ID:             fmt.Sprintf("security_violation_%s_%d", stats.ValidationID, time.Now().Unix()),
			Timestamp:      time.Now(),
			AlertType:      "security_violation",
			Severity:       "critical",
			ValidationType: stats.ValidationType,
			ValidationID:   stats.ValidationID,
			ValidationName: stats.ValidationName,
			Threshold:      0,
			ActualValue:    1,
			Message:        fmt.Sprintf("Security violation detected in validation %s", stats.ValidationName),
			SecurityImpact: "security_breach",
			Recommendations: []string{
				"Immediately investigate the security violation",
				"Review validation criteria and security policies",
				"Consider implementing additional security controls",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"validation_result": result,
			},
		})
	}

	// Check for compliance violations
	if result.ComplianceViolation {
		alerts = append(alerts, &AdvancedSecurityValidationAlert{
			ID:             fmt.Sprintf("compliance_violation_%s_%d", stats.ValidationID, time.Now().Unix()),
			Timestamp:      time.Now(),
			AlertType:      "compliance_violation",
			Severity:       "high",
			ValidationType: stats.ValidationType,
			ValidationID:   stats.ValidationID,
			ValidationName: stats.ValidationName,
			Threshold:      0,
			ActualValue:    1,
			Message:        fmt.Sprintf("Compliance violation detected in validation %s", stats.ValidationName),
			SecurityImpact: "compliance_risk",
			Recommendations: []string{
				"Review compliance requirements and validation criteria",
				"Update validation logic to meet compliance standards",
				"Document compliance violations for audit purposes",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"validation_result": result,
			},
		})
	}

	// Check for threats detected
	if result.ThreatDetected {
		alerts = append(alerts, &AdvancedSecurityValidationAlert{
			ID:             fmt.Sprintf("threat_detected_%s_%d", stats.ValidationID, time.Now().Unix()),
			Timestamp:      time.Now(),
			AlertType:      "threat_detected",
			Severity:       "critical",
			ValidationType: stats.ValidationType,
			ValidationID:   stats.ValidationID,
			ValidationName: stats.ValidationName,
			Threshold:      0,
			ActualValue:    1,
			Message:        fmt.Sprintf("Threat detected in validation %s", stats.ValidationName),
			SecurityImpact: "security_threat",
			Recommendations: []string{
				"Immediately investigate the detected threat",
				"Implement threat response procedures",
				"Review and update threat detection rules",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"validation_result": result,
			},
		})
	}

	// Check for vulnerabilities
	if result.VulnerabilityFound {
		alerts = append(alerts, &AdvancedSecurityValidationAlert{
			ID:             fmt.Sprintf("vulnerability_found_%s_%d", stats.ValidationID, time.Now().Unix()),
			Timestamp:      time.Now(),
			AlertType:      "vulnerability_found",
			Severity:       "high",
			ValidationType: stats.ValidationType,
			ValidationID:   stats.ValidationID,
			ValidationName: stats.ValidationName,
			Threshold:      0,
			ActualValue:    1,
			Message:        fmt.Sprintf("Vulnerability found in validation %s", stats.ValidationName),
			SecurityImpact: "vulnerability_exposure",
			Recommendations: []string{
				"Assess vulnerability severity and impact",
				"Implement appropriate remediation measures",
				"Update security validation procedures",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"validation_result": result,
			},
		})
	}

	// Add alerts to the list
	for _, alert := range alerts {
		svm.securityAlerts = append(svm.securityAlerts, alert)

		// Keep only recent alerts
		if len(svm.securityAlerts) > 1000 {
			svm.securityAlerts = svm.securityAlerts[1:]
		}

		svm.logger.Warn("Security validation alert triggered",
			zap.String("alert_id", alert.ID),
			zap.String("alert_type", alert.AlertType),
			zap.String("severity", alert.Severity),
			zap.String("validation_id", alert.ValidationID),
			zap.String("message", alert.Message))
	}
}

// determineSecurityAlertSeverity determines alert severity based on threshold ratio
func (svm *AdvancedSecurityValidationMonitor) determineSecurityAlertSeverity(actual, threshold float64) string {
	ratio := actual / threshold
	if ratio >= 5.0 {
		return "critical"
	} else if ratio >= 3.0 {
		return "high"
	} else if ratio >= 2.0 {
		return "medium"
	}
	return "low"
}

// generateValidationKey generates a key for validation tracking
func (svm *AdvancedSecurityValidationMonitor) generateValidationKey(validationType, validationName string) string {
	return fmt.Sprintf("%s:%s", validationType, validationName)
}

// cleanupOldValidationStats removes old validation statistics
func (svm *AdvancedSecurityValidationMonitor) cleanupOldValidationStats() {
	// Remove oldest 10% of stats
	removeCount := len(svm.validationStats) / 10
	if removeCount == 0 {
		removeCount = 1
	}

	// Simple cleanup - remove oldest by first executed time
	count := 0
	for key, stats := range svm.validationStats {
		if count >= removeCount {
			break
		}
		if time.Since(stats.FirstExecuted) > 24*time.Hour {
			delete(svm.validationStats, key)
			count++
		}
	}
}

// collectSecurityMetrics collects security performance metrics
func (svm *AdvancedSecurityValidationMonitor) collectSecurityMetrics() {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	// Collect metrics for each validation type
	for _, stats := range svm.validationStats {
		metric := &AdvancedSecurityPerformanceMetric{
			ID:                      fmt.Sprintf("security_metric_%s_%d", stats.ValidationID, time.Now().UnixNano()),
			Timestamp:               time.Now(),
			MetricType:              "security_validation",
			ValidationType:          stats.ValidationType,
			ValidationName:          stats.ValidationName,
			ExecutionTimeMs:         stats.AverageExecutionTime,
			SuccessRate:             float64(stats.SuccessCount) / float64(stats.ExecutionCount) * 100,
			FailureRate:             float64(stats.FailureCount) / float64(stats.ExecutionCount) * 100,
			TimeoutRate:             float64(stats.TimeoutCount) / float64(stats.ExecutionCount) * 100,
			ErrorRate:               float64(stats.ErrorCount) / float64(stats.ExecutionCount) * 100,
			SecurityViolationRate:   float64(stats.SecurityViolationCount) / float64(stats.ExecutionCount) * 100,
			ComplianceViolationRate: float64(stats.ComplianceViolationCount) / float64(stats.ExecutionCount) * 100,
			ThreatDetectionRate:     float64(stats.ThreatDetectionCount) / float64(stats.ExecutionCount) * 100,
			VulnerabilityRate:       float64(stats.VulnerabilityCount) / float64(stats.ExecutionCount) * 100,
			TrustScore:              stats.TrustScore,
			ConfidenceLevel:         stats.ConfidenceLevel,
			RiskLevel:               stats.RiskLevel,
			PerformanceScore:        svm.calculatePerformanceScore(stats),
			SecurityScore:           svm.calculateSecurityScore(stats),
			OverallScore:            0, // Will be calculated
			Metadata:                make(map[string]interface{}),
		}

		// Calculate overall score
		metric.OverallScore = (metric.PerformanceScore + metric.SecurityScore) / 2

		svm.performanceMetrics = append(svm.performanceMetrics, metric)

		// Keep only recent metrics
		if len(svm.performanceMetrics) > 1000 {
			svm.performanceMetrics = svm.performanceMetrics[1:]
		}
	}
}

// calculatePerformanceScore calculates a performance score (0-100)
func (svm *AdvancedSecurityValidationMonitor) calculatePerformanceScore(stats *AdvancedSecurityValidationStats) float64 {
	score := 100.0

	// Deduct points for slow execution
	if stats.AverageExecutionTime > float64(svm.config.SlowValidationThreshold.Milliseconds()) {
		score -= (stats.AverageExecutionTime - float64(svm.config.SlowValidationThreshold.Milliseconds())) / 10
	}

	// Deduct points for high failure rate
	if stats.ExecutionCount > 0 {
		failureRate := float64(stats.FailureCount) / float64(stats.ExecutionCount)
		score -= failureRate * 50
	}

	// Deduct points for high error rate
	if stats.ExecutionCount > 0 {
		errorRate := float64(stats.ErrorCount) / float64(stats.ExecutionCount)
		score -= errorRate * 30
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateSecurityScore calculates a security score (0-100)
func (svm *AdvancedSecurityValidationMonitor) calculateSecurityScore(stats *AdvancedSecurityValidationStats) float64 {
	score := 100.0

	// Deduct points for security violations
	if stats.ExecutionCount > 0 {
		securityViolationRate := float64(stats.SecurityViolationCount) / float64(stats.ExecutionCount)
		score -= securityViolationRate * 100 // Critical impact
	}

	// Deduct points for compliance violations
	if stats.ExecutionCount > 0 {
		complianceViolationRate := float64(stats.ComplianceViolationCount) / float64(stats.ExecutionCount)
		score -= complianceViolationRate * 80 // High impact
	}

	// Deduct points for threats detected
	if stats.ExecutionCount > 0 {
		threatRate := float64(stats.ThreatDetectionCount) / float64(stats.ExecutionCount)
		score -= threatRate * 90 // Critical impact
	}

	// Deduct points for vulnerabilities
	if stats.ExecutionCount > 0 {
		vulnerabilityRate := float64(stats.VulnerabilityCount) / float64(stats.ExecutionCount)
		score -= vulnerabilityRate * 70 // High impact
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// analyzeSecurityPerformance analyzes security performance patterns
func (svm *AdvancedSecurityValidationMonitor) analyzeSecurityPerformance() {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	// Analyze performance patterns and generate insights
	for _, stats := range svm.validationStats {
		if stats.PerformanceCategory == "poor" || stats.PerformanceCategory == "critical" {
			svm.logger.Info("Security validation performance analysis",
				zap.String("validation_id", stats.ValidationID),
				zap.String("validation_name", stats.ValidationName),
				zap.String("performance_category", stats.PerformanceCategory),
				zap.String("security_category", stats.SecurityCategory),
				zap.Float64("average_execution_time", stats.AverageExecutionTime),
				zap.Int64("execution_count", stats.ExecutionCount))
		}
	}
}

// generateSecurityAlerts generates security alerts based on analysis
func (svm *AdvancedSecurityValidationMonitor) generateSecurityAlerts() {
	// This method can be extended to generate additional alerts based on trends
	// and patterns in the security validation data
	svm.logger.Debug("Generating security alerts based on performance analysis")
}

// GetValidationStats returns security validation statistics
func (svm *AdvancedSecurityValidationMonitor) GetValidationStats(limit int) map[string]*AdvancedSecurityValidationStats {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	result := make(map[string]*AdvancedSecurityValidationStats)
	count := 0

	for key, stats := range svm.validationStats {
		if limit > 0 && count >= limit {
			break
		}
		result[key] = stats
		count++
	}

	return result
}

// GetSecurityAlerts returns security validation alerts
func (svm *AdvancedSecurityValidationMonitor) GetSecurityAlerts(resolved bool, limit int) []*AdvancedSecurityValidationAlert {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	var alerts []*AdvancedSecurityValidationAlert
	count := 0

	for _, alert := range svm.securityAlerts {
		if alert.Resolved == resolved {
			if limit > 0 && count >= limit {
				break
			}
			alerts = append(alerts, alert)
			count++
		}
	}

	return alerts
}

// GetPerformanceMetrics returns security performance metrics
func (svm *AdvancedSecurityValidationMonitor) GetPerformanceMetrics(limit int) []*AdvancedSecurityPerformanceMetric {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	var metrics []*AdvancedSecurityPerformanceMetric
	count := 0

	for _, metric := range svm.performanceMetrics {
		if limit > 0 && count >= limit {
			break
		}
		metrics = append(metrics, metric)
		count++
	}

	return metrics
}

// GetSecuritySystemHealth returns overall security system health
func (svm *AdvancedSecurityValidationMonitor) GetSecuritySystemHealth() *AdvancedSecuritySystemHealth {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	health := &AdvancedSecuritySystemHealth{
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Calculate overall scores
	totalValidations := 0
	totalExecutionTime := 0.0
	totalSuccessCount := int64(0)
	totalFailureCount := int64(0)
	totalSecurityViolations := 0
	totalComplianceViolations := 0
	totalThreats := 0
	totalVulnerabilities := 0
	totalSlowValidations := 0
	totalHighRiskValidations := 0
	totalTrustScore := 0.0
	totalConfidenceLevel := 0.0

	for _, stats := range svm.validationStats {
		totalValidations++
		totalExecutionTime += stats.AverageExecutionTime
		totalSuccessCount += stats.SuccessCount
		totalFailureCount += stats.FailureCount
		totalSecurityViolations += int(stats.SecurityViolationCount)
		totalComplianceViolations += int(stats.ComplianceViolationCount)
		totalThreats += int(stats.ThreatDetectionCount)
		totalVulnerabilities += int(stats.VulnerabilityCount)
		totalTrustScore += stats.TrustScore
		totalConfidenceLevel += stats.ConfidenceLevel

		if stats.PerformanceCategory == "poor" || stats.PerformanceCategory == "critical" {
			totalSlowValidations++
		}
		if stats.RiskLevel == "high" || stats.RiskLevel == "critical" {
			totalHighRiskValidations++
		}
	}

	// Calculate averages and rates
	if totalValidations > 0 {
		health.AverageExecutionTime = totalExecutionTime / float64(totalValidations)
		health.TrustScoreAverage = totalTrustScore / float64(totalValidations)
		health.ConfidenceLevelAverage = totalConfidenceLevel / float64(totalValidations)

		totalExecutions := totalSuccessCount + totalFailureCount
		if totalExecutions > 0 {
			health.SuccessRate = float64(totalSuccessCount) / float64(totalExecutions) * 100
			health.FailureRate = float64(totalFailureCount) / float64(totalExecutions) * 100
		}
	}

	// Set counts
	health.ValidationCount = totalValidations
	health.SecurityViolations = totalSecurityViolations
	health.ComplianceViolations = totalComplianceViolations
	health.ActiveThreats = totalThreats
	health.Vulnerabilities = totalVulnerabilities
	health.SlowValidations = totalSlowValidations
	health.FailedValidations = int(totalFailureCount)
	health.HighRiskValidations = totalHighRiskValidations

	// Calculate overall scores
	health.OverallPerformanceScore = svm.calculateOverallPerformanceScore(health)
	health.OverallSecurityScore = svm.calculateOverallSecurityScore(health)
	health.OverallRiskLevel = svm.determineOverallRiskLevel(health)

	return health
}

// calculateOverallPerformanceScore calculates overall performance score
func (svm *AdvancedSecurityValidationMonitor) calculateOverallPerformanceScore(health *AdvancedSecuritySystemHealth) float64 {
	score := 100.0

	// Deduct points for slow validations
	if health.ValidationCount > 0 {
		slowValidationRate := float64(health.SlowValidations) / float64(health.ValidationCount)
		score -= slowValidationRate * 30
	}

	// Deduct points for failed validations
	if health.ValidationCount > 0 {
		failureRate := float64(health.FailedValidations) / float64(health.ValidationCount)
		score -= failureRate * 50
	}

	// Deduct points for slow average execution time
	if health.AverageExecutionTime > float64(svm.config.SlowValidationThreshold.Milliseconds()) {
		score -= (health.AverageExecutionTime - float64(svm.config.SlowValidationThreshold.Milliseconds())) / 10
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateOverallSecurityScore calculates overall security score
func (svm *AdvancedSecurityValidationMonitor) calculateOverallSecurityScore(health *AdvancedSecuritySystemHealth) float64 {
	score := 100.0

	// Deduct points for security violations
	if health.ValidationCount > 0 {
		securityViolationRate := float64(health.SecurityViolations) / float64(health.ValidationCount)
		score -= securityViolationRate * 100 // Critical impact
	}

	// Deduct points for compliance violations
	if health.ValidationCount > 0 {
		complianceViolationRate := float64(health.ComplianceViolations) / float64(health.ValidationCount)
		score -= complianceViolationRate * 80 // High impact
	}

	// Deduct points for active threats
	if health.ValidationCount > 0 {
		threatRate := float64(health.ActiveThreats) / float64(health.ValidationCount)
		score -= threatRate * 90 // Critical impact
	}

	// Deduct points for vulnerabilities
	if health.ValidationCount > 0 {
		vulnerabilityRate := float64(health.Vulnerabilities) / float64(health.ValidationCount)
		score -= vulnerabilityRate * 70 // High impact
	}

	// Deduct points for high risk validations
	if health.ValidationCount > 0 {
		highRiskRate := float64(health.HighRiskValidations) / float64(health.ValidationCount)
		score -= highRiskRate * 60 // High impact
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// determineOverallRiskLevel determines overall risk level
func (svm *AdvancedSecurityValidationMonitor) determineOverallRiskLevel(health *AdvancedSecuritySystemHealth) string {
	if health.SecurityViolations > 0 || health.ActiveThreats > 0 {
		return "critical"
	} else if health.ComplianceViolations > 0 || health.Vulnerabilities > 0 {
		return "high"
	} else if health.HighRiskValidations > 0 {
		return "medium"
	}
	return "low"
}
