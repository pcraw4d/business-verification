package classification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ComprehensiveSecurityMetrics represents comprehensive security metrics
type ComprehensiveSecurityMetrics struct {
	Timestamp           time.Time                           `json:"timestamp"`
	DataSourceTrust     *SecurityDataSourceTrustMetrics     `json:"data_source_trust"`
	WebsiteVerification *SecurityWebsiteVerificationMetrics `json:"website_verification"`
	SecurityViolations  *SecurityViolationTrackingMetrics   `json:"security_violations"`
	ConfidenceIntegrity *ConfidenceIntegrityMetrics         `json:"confidence_integrity"`
	Alerts              []SecurityAlert                     `json:"alerts"`
	Performance         *SecurityPerformanceMetrics         `json:"performance"`
}

// SecurityDataSourceTrustMetrics represents data source trust metrics
type SecurityDataSourceTrustMetrics struct {
	TrustRate        float64   `json:"trust_rate"`
	TrustedCount     int64     `json:"trusted_count"`
	TotalValidations int64     `json:"total_validations"`
	TargetRate       float64   `json:"target_rate"`
	LastUpdated      time.Time `json:"last_updated"`
}

// SecurityWebsiteVerificationMetrics represents website verification metrics
type SecurityWebsiteVerificationMetrics struct {
	SuccessRate   float64   `json:"success_rate"`
	SuccessCount  int64     `json:"success_count"`
	TotalAttempts int64     `json:"total_attempts"`
	TargetRate    float64   `json:"target_rate"`
	LastUpdated   time.Time `json:"last_updated"`
}

// SecurityViolationTrackingMetrics represents security violation metrics
type SecurityViolationTrackingMetrics struct {
	TotalViolations  int64               `json:"total_violations"`
	ViolationsByType map[string]int64    `json:"violations_by_type"`
	RecentViolations []SecurityViolation `json:"recent_violations"`
	LastUpdated      time.Time           `json:"last_updated"`
}

// ConfidenceIntegrityMetrics represents confidence integrity metrics
type ConfidenceIntegrityMetrics struct {
	IntegrityRate float64   `json:"integrity_rate"`
	ValidScores   int64     `json:"valid_scores"`
	TotalScores   int64     `json:"total_scores"`
	Threshold     float64   `json:"threshold"`
	LastUpdated   time.Time `json:"last_updated"`
}

// SecurityPerformanceMetrics represents security monitoring performance metrics
type SecurityPerformanceMetrics struct {
	AverageCollectionTime time.Duration `json:"average_collection_time"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	AverageAlertTime      time.Duration `json:"average_alert_time"`
	TotalCollections      int64         `json:"total_collections"`
	LastUpdated           time.Time     `json:"last_updated"`
}

// SecurityMetricsMonitor provides comprehensive security metrics monitoring for the classification system
type SecurityMetricsMonitor struct {
	logger *zap.Logger
	config *SecurityMetricsConfig

	// Core metrics tracking
	dataSourceTrustTracker     *DataSourceTrustTracker
	websiteVerificationTracker *WebsiteVerificationTracker
	securityViolationTracker   *SecurityViolationTracker
	confidenceIntegrityTracker *ConfidenceIntegrityTracker

	// Alerting system
	alertManager *SecurityAlertManager

	// Performance tracking
	performanceTracker *SecurityPerformanceTracker

	// Thread safety
	mu sync.RWMutex

	// Control
	stopCh           chan struct{}
	monitoringActive bool
	collectionTicker *time.Ticker
}

// SecurityMetricsConfig holds configuration for security metrics monitoring
type SecurityMetricsConfig struct {
	Enabled                      bool                     `json:"enabled"`
	CollectionInterval           time.Duration            `json:"collection_interval"`
	DataSourceTrustTarget        float64                  `json:"data_source_trust_target"`       // 100% target
	WebsiteVerificationTarget    float64                  `json:"website_verification_target"`    // Success rate target
	ConfidenceIntegrityThreshold float64                  `json:"confidence_integrity_threshold"` // Minimum confidence threshold
	AlertingEnabled              bool                     `json:"alerting_enabled"`
	AlertThresholds              *SecurityAlertThresholds `json:"alert_thresholds"`
	RetentionPeriod              time.Duration            `json:"retention_period"`
	MaxMetricsHistory            int                      `json:"max_metrics_history"`
}

// SecurityAlertThresholds defines thresholds for security alerts
type SecurityAlertThresholds struct {
	DataSourceTrustBelow     float64 `json:"data_source_trust_below"`    // Alert if below this %
	WebsiteVerificationBelow float64 `json:"website_verification_below"` // Alert if below this %
	SecurityViolationsAbove  int     `json:"security_violations_above"`  // Alert if above this count
	ConfidenceIntegrityBelow float64 `json:"confidence_integrity_below"` // Alert if below this threshold
}

// DataSourceTrustTracker tracks data source trust rates
type DataSourceTrustTracker struct {
	enabled          bool
	targetRate       float64
	trustedCount     int64
	untrustedCount   int64
	totalValidations int64
	trustRate        float64
	lastUpdated      time.Time
	historicalRates  []TrustRateDataPoint
	mu               sync.RWMutex
}

// TrustRateDataPoint represents a historical trust rate data point
type TrustRateDataPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	TrustRate    float64   `json:"trust_rate"`
	TrustedCount int64     `json:"trusted_count"`
	TotalCount   int64     `json:"total_count"`
}

// WebsiteVerificationTracker tracks website verification success rates
type WebsiteVerificationTracker struct {
	enabled         bool
	targetRate      float64
	successCount    int64
	failureCount    int64
	totalAttempts   int64
	successRate     float64
	lastUpdated     time.Time
	historicalRates []VerificationRateDataPoint
	mu              sync.RWMutex
}

// VerificationRateDataPoint represents a historical verification rate data point
type VerificationRateDataPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	SuccessRate   float64   `json:"success_rate"`
	SuccessCount  int64     `json:"success_count"`
	TotalAttempts int64     `json:"total_attempts"`
}

// SecurityViolationTracker tracks security violations and alerts
type SecurityViolationTracker struct {
	enabled          bool
	alertThreshold   int
	totalViolations  int64
	violationsByType map[string]int64
	recentViolations []SecurityViolation
	lastAlertTime    time.Time
	alertCooldown    time.Duration
	mu               sync.RWMutex
}

// SecurityViolation represents a security violation event
type SecurityViolation struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"` // "low", "medium", "high", "critical"
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Resolved    bool      `json:"resolved"`
}

// ConfidenceIntegrityTracker tracks confidence score integrity
type ConfidenceIntegrityTracker struct {
	enabled         bool
	threshold       float64
	totalScores     int64
	validScores     int64
	invalidScores   int64
	integrityRate   float64
	lastUpdated     time.Time
	historicalRates []ConfidenceIntegrityDataPoint
	mu              sync.RWMutex
}

// ConfidenceIntegrityDataPoint represents a historical confidence integrity data point
type ConfidenceIntegrityDataPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	IntegrityRate float64   `json:"integrity_rate"`
	ValidScores   int64     `json:"valid_scores"`
	TotalScores   int64     `json:"total_scores"`
}

// SecurityAlertManager manages security alerts
type SecurityAlertManager struct {
	enabled        bool
	thresholds     *SecurityAlertThresholds
	alerts         []SecurityAlert
	lastAlertTimes map[string]time.Time
	alertCooldown  time.Duration
	mu             sync.RWMutex
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`
	Severity     string    `json:"severity"`
	Message      string    `json:"message"`
	Metric       string    `json:"metric"`
	Value        float64   `json:"value"`
	Threshold    float64   `json:"threshold"`
	Acknowledged bool      `json:"acknowledged"`
}

// SecurityPerformanceTracker tracks security monitoring performance
type SecurityPerformanceTracker struct {
	enabled         bool
	collectionTimes []time.Duration
	processingTimes []time.Duration
	alertTimes      []time.Duration
	mu              sync.RWMutex
}

// NewSecurityMetricsMonitor creates a new security metrics monitor
func NewSecurityMetricsMonitor(config *SecurityMetricsConfig, logger *zap.Logger) *SecurityMetricsMonitor {
	if config == nil {
		config = &SecurityMetricsConfig{
			Enabled:                      true,
			CollectionInterval:           30 * time.Second,
			DataSourceTrustTarget:        100.0, // 100% target
			WebsiteVerificationTarget:    95.0,  // 95% target
			ConfidenceIntegrityThreshold: 0.8,   // 80% threshold
			AlertingEnabled:              true,
			RetentionPeriod:              24 * time.Hour,
			MaxMetricsHistory:            1000,
		}
	}

	if config.AlertThresholds == nil {
		config.AlertThresholds = &SecurityAlertThresholds{
			DataSourceTrustBelow:     95.0, // Alert if below 95%
			WebsiteVerificationBelow: 90.0, // Alert if below 90%
			SecurityViolationsAbove:  10,   // Alert if more than 10 violations
			ConfidenceIntegrityBelow: 0.7,  // Alert if below 70%
		}
	}

	return &SecurityMetricsMonitor{
		logger: logger,
		config: config,
		dataSourceTrustTracker: &DataSourceTrustTracker{
			enabled:         true,
			targetRate:      config.DataSourceTrustTarget,
			historicalRates: make([]TrustRateDataPoint, 0),
		},
		websiteVerificationTracker: &WebsiteVerificationTracker{
			enabled:         true,
			targetRate:      config.WebsiteVerificationTarget,
			historicalRates: make([]VerificationRateDataPoint, 0),
		},
		securityViolationTracker: &SecurityViolationTracker{
			enabled:          true,
			alertThreshold:   config.AlertThresholds.SecurityViolationsAbove,
			violationsByType: make(map[string]int64),
			recentViolations: make([]SecurityViolation, 0),
			alertCooldown:    5 * time.Minute,
		},
		confidenceIntegrityTracker: &ConfidenceIntegrityTracker{
			enabled:         true,
			threshold:       config.ConfidenceIntegrityThreshold,
			historicalRates: make([]ConfidenceIntegrityDataPoint, 0),
		},
		alertManager: &SecurityAlertManager{
			enabled:        config.AlertingEnabled,
			thresholds:     config.AlertThresholds,
			alerts:         make([]SecurityAlert, 0),
			lastAlertTimes: make(map[string]time.Time),
			alertCooldown:  5 * time.Minute,
		},
		performanceTracker: &SecurityPerformanceTracker{
			enabled:         true,
			collectionTimes: make([]time.Duration, 0),
			processingTimes: make([]time.Duration, 0),
			alertTimes:      make([]time.Duration, 0),
		},
		stopCh: make(chan struct{}),
	}
}

// Start starts the security metrics monitoring
func (smm *SecurityMetricsMonitor) Start() {
	smm.mu.Lock()
	defer smm.mu.Unlock()

	if smm.monitoringActive {
		return
	}

	smm.monitoringActive = true
	smm.collectionTicker = time.NewTicker(smm.config.CollectionInterval)

	// Start background monitoring
	go smm.monitoringLoop()

	smm.logger.Info("Security metrics monitoring started",
		zap.Duration("collection_interval", smm.config.CollectionInterval),
		zap.Float64("data_source_trust_target", smm.config.DataSourceTrustTarget),
		zap.Float64("website_verification_target", smm.config.WebsiteVerificationTarget),
		zap.Float64("confidence_integrity_threshold", smm.config.ConfidenceIntegrityThreshold),
		zap.Bool("alerting_enabled", smm.config.AlertingEnabled))
}

// Stop stops the security metrics monitoring
func (smm *SecurityMetricsMonitor) Stop() {
	smm.mu.Lock()
	defer smm.mu.Unlock()

	if !smm.monitoringActive {
		return
	}

	smm.monitoringActive = false
	if smm.collectionTicker != nil {
		smm.collectionTicker.Stop()
	}
	close(smm.stopCh)

	smm.logger.Info("Security metrics monitoring stopped")
}

// monitoringLoop runs the main monitoring loop
func (smm *SecurityMetricsMonitor) monitoringLoop() {
	for {
		select {
		case <-smm.collectionTicker.C:
			startTime := time.Now()
			smm.collectSecurityMetrics()
			smm.checkSecurityAlerts()
			smm.cleanupOldData()

			// Track performance
			collectionTime := time.Since(startTime)
			smm.performanceTracker.RecordCollectionTime(collectionTime)

		case <-smm.stopCh:
			return
		}
	}
}

// RecordDataSourceTrust records a data source trust validation
func (smm *SecurityMetricsMonitor) RecordDataSourceTrust(ctx context.Context, trusted bool, source string) {
	if !smm.config.Enabled {
		return
	}

	smm.dataSourceTrustTracker.RecordTrust(trusted, source)

	smm.logger.Debug("Data source trust recorded",
		zap.Bool("trusted", trusted),
		zap.String("source", source),
		zap.Float64("current_trust_rate", smm.dataSourceTrustTracker.GetTrustRate()))
}

// RecordWebsiteVerification records a website verification result
func (smm *SecurityMetricsMonitor) RecordWebsiteVerification(ctx context.Context, success bool, domain string, method string) {
	if !smm.config.Enabled {
		return
	}

	smm.websiteVerificationTracker.RecordVerification(success, domain, method)

	smm.logger.Debug("Website verification recorded",
		zap.Bool("success", success),
		zap.String("domain", domain),
		zap.String("method", method),
		zap.Float64("current_success_rate", smm.websiteVerificationTracker.GetSuccessRate()))
}

// RecordSecurityViolation records a security violation
func (smm *SecurityMetricsMonitor) RecordSecurityViolation(ctx context.Context, violationType, severity, description, source string) {
	if !smm.config.Enabled {
		return
	}

	violation := SecurityViolation{
		ID:          fmt.Sprintf("violation_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        violationType,
		Severity:    severity,
		Description: description,
		Source:      source,
		Resolved:    false,
	}

	smm.securityViolationTracker.RecordViolation(violation)

	smm.logger.Warn("Security violation recorded",
		zap.String("type", violationType),
		zap.String("severity", severity),
		zap.String("source", source),
		zap.String("description", description))
}

// RecordConfidenceScore records a confidence score for integrity tracking
func (smm *SecurityMetricsMonitor) RecordConfidenceScore(ctx context.Context, score float64, valid bool, source string) {
	if !smm.config.Enabled {
		return
	}

	smm.confidenceIntegrityTracker.RecordScore(score, valid, source)

	smm.logger.Debug("Confidence score recorded",
		zap.Float64("score", score),
		zap.Bool("valid", valid),
		zap.String("source", source),
		zap.Float64("current_integrity_rate", smm.confidenceIntegrityTracker.GetIntegrityRate()))
}

// GetSecurityMetrics returns current security metrics
func (smm *SecurityMetricsMonitor) GetSecurityMetrics() *ComprehensiveSecurityMetrics {
	smm.mu.RLock()
	defer smm.mu.RUnlock()

	return &ComprehensiveSecurityMetrics{
		Timestamp: time.Now(),
		DataSourceTrust: &SecurityDataSourceTrustMetrics{
			TrustRate:        smm.dataSourceTrustTracker.GetTrustRate(),
			TrustedCount:     smm.dataSourceTrustTracker.GetTrustedCount(),
			TotalValidations: smm.dataSourceTrustTracker.GetTotalValidations(),
			TargetRate:       smm.config.DataSourceTrustTarget,
			LastUpdated:      smm.dataSourceTrustTracker.GetLastUpdated(),
		},
		WebsiteVerification: &SecurityWebsiteVerificationMetrics{
			SuccessRate:   smm.websiteVerificationTracker.GetSuccessRate(),
			SuccessCount:  smm.websiteVerificationTracker.GetSuccessCount(),
			TotalAttempts: smm.websiteVerificationTracker.GetTotalAttempts(),
			TargetRate:    smm.config.WebsiteVerificationTarget,
			LastUpdated:   smm.websiteVerificationTracker.GetLastUpdated(),
		},
		SecurityViolations: &SecurityViolationTrackingMetrics{
			TotalViolations:  smm.securityViolationTracker.GetTotalViolations(),
			ViolationsByType: smm.securityViolationTracker.GetViolationsByType(),
			RecentViolations: smm.securityViolationTracker.GetRecentViolations(10),
			LastUpdated:      smm.securityViolationTracker.GetLastUpdated(),
		},
		ConfidenceIntegrity: &ConfidenceIntegrityMetrics{
			IntegrityRate: smm.confidenceIntegrityTracker.GetIntegrityRate(),
			ValidScores:   smm.confidenceIntegrityTracker.GetValidScores(),
			TotalScores:   smm.confidenceIntegrityTracker.GetTotalScores(),
			Threshold:     smm.config.ConfidenceIntegrityThreshold,
			LastUpdated:   smm.confidenceIntegrityTracker.GetLastUpdated(),
		},
		Alerts:      smm.alertManager.GetRecentAlerts(10),
		Performance: smm.performanceTracker.GetPerformanceMetrics(),
	}
}

// collectSecurityMetrics collects and updates all security metrics
func (smm *SecurityMetricsMonitor) collectSecurityMetrics() {
	startTime := time.Now()

	// Update all tracker metrics
	smm.dataSourceTrustTracker.UpdateMetrics()
	smm.websiteVerificationTracker.UpdateMetrics()
	smm.confidenceIntegrityTracker.UpdateMetrics()

	processingTime := time.Since(startTime)
	smm.performanceTracker.RecordProcessingTime(processingTime)
}

// checkSecurityAlerts checks for security alert conditions
func (smm *SecurityMetricsMonitor) checkSecurityAlerts() {
	if !smm.config.AlertingEnabled {
		return
	}

	startTime := time.Now()

	// Check data source trust rate
	trustRate := smm.dataSourceTrustTracker.GetTrustRate()
	if trustRate < smm.config.AlertThresholds.DataSourceTrustBelow {
		smm.alertManager.CreateAlert("data_source_trust", "high",
			fmt.Sprintf("Data source trust rate %.2f%% is below threshold %.2f%%",
				trustRate, smm.config.AlertThresholds.DataSourceTrustBelow),
			"data_source_trust", trustRate, smm.config.AlertThresholds.DataSourceTrustBelow)
	}

	// Check website verification rate
	verificationRate := smm.websiteVerificationTracker.GetSuccessRate()
	if verificationRate < smm.config.AlertThresholds.WebsiteVerificationBelow {
		smm.alertManager.CreateAlert("website_verification", "high",
			fmt.Sprintf("Website verification rate %.2f%% is below threshold %.2f%%",
				verificationRate, smm.config.AlertThresholds.WebsiteVerificationBelow),
			"website_verification", verificationRate, smm.config.AlertThresholds.WebsiteVerificationBelow)
	}

	// Check security violations
	totalViolations := smm.securityViolationTracker.GetTotalViolations()
	if totalViolations > int64(smm.config.AlertThresholds.SecurityViolationsAbove) {
		smm.alertManager.CreateAlert("security_violations", "critical",
			fmt.Sprintf("Total security violations %d exceeds threshold %d",
				totalViolations, smm.config.AlertThresholds.SecurityViolationsAbove),
			"security_violations", float64(totalViolations), float64(smm.config.AlertThresholds.SecurityViolationsAbove))
	}

	// Check confidence integrity
	integrityRate := smm.confidenceIntegrityTracker.GetIntegrityRate()
	if integrityRate < smm.config.AlertThresholds.ConfidenceIntegrityBelow {
		smm.alertManager.CreateAlert("confidence_integrity", "medium",
			fmt.Sprintf("Confidence integrity rate %.2f is below threshold %.2f",
				integrityRate, smm.config.AlertThresholds.ConfidenceIntegrityBelow),
			"confidence_integrity", integrityRate, smm.config.AlertThresholds.ConfidenceIntegrityBelow)
	}

	alertTime := time.Since(startTime)
	smm.performanceTracker.RecordAlertTime(alertTime)
}

// cleanupOldData cleans up old metrics data
func (smm *SecurityMetricsMonitor) cleanupOldData() {
	cutoffTime := time.Now().Add(-smm.config.RetentionPeriod)

	smm.dataSourceTrustTracker.CleanupOldData(cutoffTime)
	smm.websiteVerificationTracker.CleanupOldData(cutoffTime)
	smm.confidenceIntegrityTracker.CleanupOldData(cutoffTime)
	smm.alertManager.CleanupOldAlerts(cutoffTime)
	smm.performanceTracker.CleanupOldData(cutoffTime)
}
