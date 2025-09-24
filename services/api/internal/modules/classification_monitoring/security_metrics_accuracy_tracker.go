package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SecurityMetricsAccuracyTracker provides comprehensive security metrics tracking with detailed accuracy monitoring
type SecurityMetricsAccuracyTracker struct {
	config *SecurityMetricsConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Core tracking
	trustedDataSources   map[string]*TrustedDataSourceMetrics
	websiteVerifications map[string]*WebsiteVerificationMetrics
	securityViolations   map[string]*SecurityViolationMetrics
	confidenceIntegrity  *ConfidenceIntegrityMetrics

	// Advanced components
	securityAnalyzer  *SecurityAnalyzer
	threatDetector    *ThreatDetector
	complianceMonitor *ComplianceMonitor
	auditLogger       *SecurityAuditLogger
}

// SecurityMetricsConfig configuration for security metrics tracking
type SecurityMetricsConfig struct {
	// Trusted data source settings
	TrustedSourceAccuracyThreshold   float64
	TrustedSourceConfidenceThreshold float64
	TrustedSourceLatencyThreshold    time.Duration

	// Website verification settings
	WebsiteVerificationAccuracyThreshold float64
	WebsiteVerificationTimeoutThreshold  time.Duration

	// Security violation settings
	SecurityViolationThreshold float64
	AnomalyDetectionThreshold  float64
	ThreatDetectionSensitivity float64

	// Compliance settings
	ComplianceAccuracyThreshold float64
	AuditLogRetentionPeriod     time.Duration

	// Alert settings
	SecurityAlertCooldownPeriod time.Duration
	MaxSecurityAlertsPerSource  int
}

// TrustedDataSourceMetrics tracks accuracy and performance for trusted data sources
type TrustedDataSourceMetrics struct {
	SourceName  string
	SourceType  string // "government", "credit_bureau", "business_registry", "financial_institution"
	LastUpdated time.Time

	// Accuracy metrics
	TotalRequests      int64
	SuccessfulRequests int64
	AccuracyScore      float64
	ConfidenceScore    float64
	AverageLatency     time.Duration

	// Security metrics
	TrustLevel                float64 // 0.0 to 1.0
	VerificationRate          float64
	DataIntegrityScore        float64
	AuthenticationSuccessRate float64

	// Historical data
	HistoricalAccuracy   []*SecurityAccuracyDataPoint
	HistoricalConfidence []*SecurityConfidenceDataPoint
	HistoricalLatency    []*SecurityLatencyDataPoint

	// Status and alerts
	Status            string // "active", "degraded", "critical", "suspended"
	LastSecurityCheck time.Time
	SecurityAlerts    []*SecurityAlert
}

// WebsiteVerificationMetrics tracks website verification accuracy
type WebsiteVerificationMetrics struct {
	Domain             string
	VerificationMethod string // "ssl_certificate", "dns_verification", "whois_verification"
	LastUpdated        time.Time

	// Verification metrics
	TotalVerifications      int64
	SuccessfulVerifications int64
	VerificationAccuracy    float64
	AverageVerificationTime time.Duration

	// Security metrics
	SSLValidityScore      float64
	DomainReputationScore float64
	CertificateTrustScore float64

	// Historical data
	HistoricalVerifications []*WebsiteVerificationDataPoint

	// Status
	Status              string // "verified", "pending", "failed", "expired"
	LastVerification    time.Time
	NextVerificationDue time.Time
}

// SecurityViolationMetrics tracks security violations and anomalies
type SecurityViolationMetrics struct {
	ViolationType string // "data_tampering", "unauthorized_access", "suspicious_pattern", "integrity_failure"
	LastUpdated   time.Time

	// Violation metrics
	TotalViolations    int64
	CriticalViolations int64
	ViolationRate      float64
	AverageSeverity    float64

	// Detection metrics
	DetectionAccuracy float64
	FalsePositiveRate float64
	ResponseTime      time.Duration

	// Historical data
	HistoricalViolations []*SecurityViolationDataPoint

	// Status
	Status              string // "monitoring", "investigating", "resolved", "escalated"
	LastViolation       time.Time
	InvestigationStatus string
}

// ConfidenceIntegrityMetrics tracks confidence and data integrity
type ConfidenceIntegrityMetrics struct {
	LastUpdated time.Time

	// Integrity metrics
	OverallIntegrityScore  float64
	DataConsistencyScore   float64
	SourceReliabilityScore float64
	CrossValidationScore   float64

	// Confidence metrics
	AverageConfidence  float64
	ConfidenceVariance float64
	LowConfidenceRate  float64

	// Historical data
	HistoricalIntegrity  []*IntegrityDataPoint
	HistoricalConfidence []*ConfidenceIntegrityDataPoint

	// Status
	IntegrityStatus    string // "high", "medium", "low", "critical"
	LastIntegrityCheck time.Time
}

// SecurityAlert represents a security-related alert
type SecurityAlert struct {
	ID                  string
	AlertType           string // "trusted_source_degradation", "website_verification_failure", "security_violation", "integrity_breach"
	Severity            string // "low", "medium", "high", "critical"
	Source              string
	Message             string
	Timestamp           time.Time
	Resolved            bool
	ResolutionTimestamp *time.Time
	ResolutionNotes     string
}

// Security data point types
type SecurityAccuracyDataPoint struct {
	Timestamp  time.Time
	Accuracy   float64
	Confidence float64
	Source     string
	RequestID  string
}

type SecurityConfidenceDataPoint struct {
	Timestamp  time.Time
	Confidence float64
	Source     string
	RequestID  string
}

type SecurityLatencyDataPoint struct {
	Timestamp time.Time
	Latency   time.Duration
	Source    string
	RequestID string
}

type WebsiteVerificationDataPoint struct {
	Timestamp        time.Time
	Success          bool
	VerificationTime time.Duration
	Method           string
	Domain           string
}

type SecurityViolationDataPoint struct {
	Timestamp     time.Time
	ViolationType string
	Severity      float64
	Source        string
	Description   string
}

type IntegrityDataPoint struct {
	Timestamp        time.Time
	IntegrityScore   float64
	ConsistencyScore float64
	ReliabilityScore float64
}

type ConfidenceIntegrityDataPoint struct {
	Timestamp       time.Time
	Confidence      float64
	Integrity       float64
	CrossValidation float64
}

// DefaultSecurityMetricsConfig returns default configuration
func DefaultSecurityMetricsConfig() *SecurityMetricsConfig {
	return &SecurityMetricsConfig{
		TrustedSourceAccuracyThreshold:       0.95, // 95% accuracy threshold
		TrustedSourceConfidenceThreshold:     0.90, // 90% confidence threshold
		TrustedSourceLatencyThreshold:        5 * time.Second,
		WebsiteVerificationAccuracyThreshold: 0.98, // 98% verification accuracy
		WebsiteVerificationTimeoutThreshold:  10 * time.Second,
		SecurityViolationThreshold:           0.05,                // 5% violation rate threshold
		AnomalyDetectionThreshold:            0.10,                // 10% anomaly threshold
		ThreatDetectionSensitivity:           0.80,                // 80% sensitivity
		ComplianceAccuracyThreshold:          0.95,                // 95% compliance threshold
		AuditLogRetentionPeriod:              90 * 24 * time.Hour, // 90 days
		SecurityAlertCooldownPeriod:          15 * time.Minute,
		MaxSecurityAlertsPerSource:           5,
	}
}

// NewSecurityMetricsAccuracyTracker creates a new security metrics accuracy tracker
func NewSecurityMetricsAccuracyTracker(config *SecurityMetricsConfig, logger *zap.Logger) *SecurityMetricsAccuracyTracker {
	if config == nil {
		config = DefaultSecurityMetricsConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &SecurityMetricsAccuracyTracker{
		config:               config,
		logger:               logger,
		trustedDataSources:   make(map[string]*TrustedDataSourceMetrics),
		websiteVerifications: make(map[string]*WebsiteVerificationMetrics),
		securityViolations:   make(map[string]*SecurityViolationMetrics),
		confidenceIntegrity: &ConfidenceIntegrityMetrics{
			IntegrityStatus: "high",
		},
		securityAnalyzer: &SecurityAnalyzer{
			config: config,
			logger: logger,
		},
		threatDetector: &ThreatDetector{
			config: config,
			logger: logger,
		},
		complianceMonitor: &ComplianceMonitor{
			config: config,
			logger: logger,
		},
		auditLogger: &SecurityAuditLogger{
			config: config,
			logger: logger,
		},
	}
}

// TrackTrustedDataSourceResult tracks a result from a trusted data source
func (smat *SecurityMetricsAccuracyTracker) TrackTrustedDataSourceResult(ctx context.Context, result *ClassificationResult) error {
	smat.mu.Lock()
	defer smat.mu.Unlock()

	// Extract source information from metadata
	sourceName := smat.extractSourceName(result)
	sourceType := smat.extractSourceType(result)

	// Get or create source metrics
	metrics, exists := smat.trustedDataSources[sourceName]
	if !exists {
		metrics = smat.createTrustedDataSourceMetrics(sourceName, sourceType)
		smat.trustedDataSources[sourceName] = metrics
	}

	// Update metrics
	smat.updateTrustedDataSourceMetrics(metrics, result)

	// Perform security analysis
	smat.performSecurityAnalysis(metrics, result)

	// Check for security alerts
	smat.checkSecurityAlerts(metrics)

	// Update audit log
	smat.auditLogger.LogTrustedDataSourceAccess(result)

	return nil
}

// TrackWebsiteVerification tracks a website verification result
func (smat *SecurityMetricsAccuracyTracker) TrackWebsiteVerification(ctx context.Context, domain, method string, success bool, verificationTime time.Duration) error {
	smat.mu.Lock()
	defer smat.mu.Unlock()

	// Get or create verification metrics
	metrics, exists := smat.websiteVerifications[domain]
	if !exists {
		metrics = smat.createWebsiteVerificationMetrics(domain, method)
		smat.websiteVerifications[domain] = metrics
	}

	// Update metrics
	smat.updateWebsiteVerificationMetrics(metrics, success, verificationTime)

	// Check for verification alerts
	smat.checkWebsiteVerificationAlerts(metrics)

	// Update audit log
	smat.auditLogger.LogWebsiteVerification(domain, method, success)

	return nil
}

// TrackSecurityViolation tracks a security violation
func (smat *SecurityMetricsAccuracyTracker) TrackSecurityViolation(ctx context.Context, violationType, source, description string, severity float64) error {
	smat.mu.Lock()
	defer smat.mu.Unlock()

	// Get or create violation metrics
	metrics, exists := smat.securityViolations[violationType]
	if !exists {
		metrics = smat.createSecurityViolationMetrics(violationType)
		smat.securityViolations[violationType] = metrics
	}

	// Update metrics
	smat.updateSecurityViolationMetrics(metrics, source, description, severity)

	// Perform threat detection
	smat.performThreatDetection(metrics)

	// Update audit log
	smat.auditLogger.LogSecurityViolation(violationType, source, severity)

	return nil
}

// GetTrustedDataSourceMetrics returns metrics for a specific trusted data source
func (smat *SecurityMetricsAccuracyTracker) GetTrustedDataSourceMetrics(sourceName string) *TrustedDataSourceMetrics {
	smat.mu.RLock()
	defer smat.mu.RUnlock()

	metrics, exists := smat.trustedDataSources[sourceName]
	if !exists {
		return nil
	}

	return smat.copyTrustedDataSourceMetrics(metrics)
}

// GetAllTrustedDataSourceMetrics returns metrics for all trusted data sources
func (smat *SecurityMetricsAccuracyTracker) GetAllTrustedDataSourceMetrics() map[string]*TrustedDataSourceMetrics {
	smat.mu.RLock()
	defer smat.mu.RUnlock()

	result := make(map[string]*TrustedDataSourceMetrics)
	for name, metrics := range smat.trustedDataSources {
		result[name] = smat.copyTrustedDataSourceMetrics(metrics)
	}

	return result
}

// GetWebsiteVerificationMetrics returns metrics for a specific domain
func (smat *SecurityMetricsAccuracyTracker) GetWebsiteVerificationMetrics(domain string) *WebsiteVerificationMetrics {
	smat.mu.RLock()
	defer smat.mu.RUnlock()

	metrics, exists := smat.websiteVerifications[domain]
	if !exists {
		return nil
	}

	return smat.copyWebsiteVerificationMetrics(metrics)
}

// GetSecurityViolationMetrics returns metrics for a specific violation type
func (smat *SecurityMetricsAccuracyTracker) GetSecurityViolationMetrics(violationType string) *SecurityViolationMetrics {
	smat.mu.RLock()
	defer smat.mu.RUnlock()

	metrics, exists := smat.securityViolations[violationType]
	if !exists {
		return nil
	}

	return smat.copySecurityViolationMetrics(metrics)
}

// GetConfidenceIntegrityMetrics returns confidence and integrity metrics
func (smat *SecurityMetricsAccuracyTracker) GetConfidenceIntegrityMetrics() *ConfidenceIntegrityMetrics {
	smat.mu.RLock()
	defer smat.mu.RUnlock()

	return smat.copyConfidenceIntegrityMetrics(smat.confidenceIntegrity)
}

// GetSecurityAlerts returns all security alerts
func (smat *SecurityMetricsAccuracyTracker) GetSecurityAlerts() []*SecurityAlert {
	smat.mu.RLock()
	defer smat.mu.RUnlock()

	var allAlerts []*SecurityAlert

	// Collect alerts from trusted data sources
	for _, metrics := range smat.trustedDataSources {
		allAlerts = append(allAlerts, metrics.SecurityAlerts...)
	}

	// Collect alerts from website verifications
	for range smat.websiteVerifications {
		// Website verification alerts would be stored here if implemented
	}

	// Collect alerts from security violations
	for range smat.securityViolations {
		// Security violation alerts would be stored here if implemented
	}

	return allAlerts
}

// GetActiveSecurityAlerts returns unresolved security alerts
func (smat *SecurityMetricsAccuracyTracker) GetActiveSecurityAlerts() []*SecurityAlert {
	allAlerts := smat.GetSecurityAlerts()

	var activeAlerts []*SecurityAlert
	for _, alert := range allAlerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// Helper methods

// extractSourceName extracts the source name from classification result metadata
func (smat *SecurityMetricsAccuracyTracker) extractSourceName(result *ClassificationResult) string {
	if source, exists := result.Metadata["trusted_source"]; exists {
		if s, ok := source.(string); ok {
			return s
		}
	}
	return "unknown_source"
}

// extractSourceType extracts the source type from classification result metadata
func (smat *SecurityMetricsAccuracyTracker) extractSourceType(result *ClassificationResult) string {
	if sourceType, exists := result.Metadata["source_type"]; exists {
		if st, ok := sourceType.(string); ok {
			return st
		}
	}
	return "unknown_type"
}

// createTrustedDataSourceMetrics creates new trusted data source metrics
func (smat *SecurityMetricsAccuracyTracker) createTrustedDataSourceMetrics(sourceName, sourceType string) *TrustedDataSourceMetrics {
	return &TrustedDataSourceMetrics{
		SourceName:           sourceName,
		SourceType:           sourceType,
		LastUpdated:          time.Now(),
		HistoricalAccuracy:   make([]*SecurityAccuracyDataPoint, 0),
		HistoricalConfidence: make([]*SecurityConfidenceDataPoint, 0),
		HistoricalLatency:    make([]*SecurityLatencyDataPoint, 0),
		SecurityAlerts:       make([]*SecurityAlert, 0),
		Status:               "active",
		TrustLevel:           1.0, // Start with full trust
	}
}

// createWebsiteVerificationMetrics creates new website verification metrics
func (smat *SecurityMetricsAccuracyTracker) createWebsiteVerificationMetrics(domain, method string) *WebsiteVerificationMetrics {
	return &WebsiteVerificationMetrics{
		Domain:                  domain,
		VerificationMethod:      method,
		LastUpdated:             time.Now(),
		HistoricalVerifications: make([]*WebsiteVerificationDataPoint, 0),
		Status:                  "pending",
	}
}

// createSecurityViolationMetrics creates new security violation metrics
func (smat *SecurityMetricsAccuracyTracker) createSecurityViolationMetrics(violationType string) *SecurityViolationMetrics {
	return &SecurityViolationMetrics{
		ViolationType:        violationType,
		LastUpdated:          time.Now(),
		HistoricalViolations: make([]*SecurityViolationDataPoint, 0),
		Status:               "monitoring",
	}
}

// updateTrustedDataSourceMetrics updates trusted data source metrics
func (smat *SecurityMetricsAccuracyTracker) updateTrustedDataSourceMetrics(metrics *TrustedDataSourceMetrics, result *ClassificationResult) {
	metrics.LastUpdated = time.Now()
	metrics.TotalRequests++

	// Update accuracy
	if result.IsCorrect != nil && *result.IsCorrect {
		metrics.SuccessfulRequests++
	}
	metrics.AccuracyScore = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)

	// Update confidence
	metrics.ConfidenceScore = (metrics.ConfidenceScore*float64(metrics.TotalRequests-1) + result.ConfidenceScore) / float64(metrics.TotalRequests)

	// Update latency (approximate)
	requestLatency := time.Since(result.Timestamp)
	metrics.AverageLatency = (metrics.AverageLatency*time.Duration(metrics.TotalRequests-1) + requestLatency) / time.Duration(metrics.TotalRequests)

	// Add to historical data
	accuracyPoint := &SecurityAccuracyDataPoint{
		Timestamp:  result.Timestamp,
		Accuracy:   metrics.AccuracyScore,
		Confidence: result.ConfidenceScore,
		Source:     metrics.SourceName,
		RequestID:  result.ID,
	}
	metrics.HistoricalAccuracy = append(metrics.HistoricalAccuracy, accuracyPoint)

	confidencePoint := &SecurityConfidenceDataPoint{
		Timestamp:  result.Timestamp,
		Confidence: result.ConfidenceScore,
		Source:     metrics.SourceName,
		RequestID:  result.ID,
	}
	metrics.HistoricalConfidence = append(metrics.HistoricalConfidence, confidencePoint)

	latencyPoint := &SecurityLatencyDataPoint{
		Timestamp: result.Timestamp,
		Latency:   requestLatency,
		Source:    metrics.SourceName,
		RequestID: result.ID,
	}
	metrics.HistoricalLatency = append(metrics.HistoricalLatency, latencyPoint)

	// Maintain window size (keep last 1000 points)
	if len(metrics.HistoricalAccuracy) > 1000 {
		metrics.HistoricalAccuracy = metrics.HistoricalAccuracy[1:]
	}
	if len(metrics.HistoricalConfidence) > 1000 {
		metrics.HistoricalConfidence = metrics.HistoricalConfidence[1:]
	}
	if len(metrics.HistoricalLatency) > 1000 {
		metrics.HistoricalLatency = metrics.HistoricalLatency[1:]
	}
}

// updateWebsiteVerificationMetrics updates website verification metrics
func (smat *SecurityMetricsAccuracyTracker) updateWebsiteVerificationMetrics(metrics *WebsiteVerificationMetrics, success bool, verificationTime time.Duration) {
	metrics.LastUpdated = time.Now()
	metrics.TotalVerifications++

	if success {
		metrics.SuccessfulVerifications++
	}
	metrics.VerificationAccuracy = float64(metrics.SuccessfulVerifications) / float64(metrics.TotalVerifications)
	metrics.AverageVerificationTime = (metrics.AverageVerificationTime*time.Duration(metrics.TotalVerifications-1) + verificationTime) / time.Duration(metrics.TotalVerifications)

	// Add to historical data
	verificationPoint := &WebsiteVerificationDataPoint{
		Timestamp:        time.Now(),
		Success:          success,
		VerificationTime: verificationTime,
		Method:           metrics.VerificationMethod,
		Domain:           metrics.Domain,
	}
	metrics.HistoricalVerifications = append(metrics.HistoricalVerifications, verificationPoint)

	// Maintain window size
	if len(metrics.HistoricalVerifications) > 1000 {
		metrics.HistoricalVerifications = metrics.HistoricalVerifications[1:]
	}

	// Update status
	if success {
		metrics.Status = "verified"
		metrics.LastVerification = time.Now()
		metrics.NextVerificationDue = time.Now().Add(24 * time.Hour) // Re-verify daily
	} else {
		metrics.Status = "failed"
	}
}

// updateSecurityViolationMetrics updates security violation metrics
func (smat *SecurityMetricsAccuracyTracker) updateSecurityViolationMetrics(metrics *SecurityViolationMetrics, source, description string, severity float64) {
	metrics.LastUpdated = time.Now()
	metrics.TotalViolations++

	if severity >= 0.8 { // High severity threshold
		metrics.CriticalViolations++
	}
	metrics.ViolationRate = float64(metrics.TotalViolations) / float64(time.Since(metrics.LastUpdated).Hours())
	metrics.AverageSeverity = (metrics.AverageSeverity*float64(metrics.TotalViolations-1) + severity) / float64(metrics.TotalViolations)

	// Add to historical data
	violationPoint := &SecurityViolationDataPoint{
		Timestamp:     time.Now(),
		ViolationType: metrics.ViolationType,
		Severity:      severity,
		Source:        source,
		Description:   description,
	}
	metrics.HistoricalViolations = append(metrics.HistoricalViolations, violationPoint)

	// Maintain window size
	if len(metrics.HistoricalViolations) > 1000 {
		metrics.HistoricalViolations = metrics.HistoricalViolations[1:]
	}

	metrics.LastViolation = time.Now()
}

// performSecurityAnalysis performs security analysis on trusted data source
func (smat *SecurityMetricsAccuracyTracker) performSecurityAnalysis(metrics *TrustedDataSourceMetrics, result *ClassificationResult) {
	// Update trust level based on performance
	if metrics.AccuracyScore < smat.config.TrustedSourceAccuracyThreshold {
		metrics.TrustLevel *= 0.95 // Reduce trust level
	} else if metrics.AccuracyScore > smat.config.TrustedSourceAccuracyThreshold+0.05 {
		metrics.TrustLevel = math.Min(1.0, metrics.TrustLevel*1.01) // Increase trust level
	}

	// Update data integrity score
	metrics.DataIntegrityScore = smat.securityAnalyzer.CalculateDataIntegrityScore(metrics)

	// Update verification rate
	metrics.VerificationRate = smat.securityAnalyzer.CalculateVerificationRate(metrics)
}

// performThreatDetection performs threat detection on security violations
func (smat *SecurityMetricsAccuracyTracker) performThreatDetection(metrics *SecurityViolationMetrics) {
	// Update detection accuracy
	metrics.DetectionAccuracy = smat.threatDetector.CalculateDetectionAccuracy(metrics)

	// Update false positive rate
	metrics.FalsePositiveRate = smat.threatDetector.CalculateFalsePositiveRate(metrics)
}

// checkSecurityAlerts checks for security alerts on trusted data source
func (smat *SecurityMetricsAccuracyTracker) checkSecurityAlerts(metrics *TrustedDataSourceMetrics) {
	// Check accuracy threshold
	if metrics.AccuracyScore < smat.config.TrustedSourceAccuracyThreshold {
		smat.createSecurityAlert(metrics, "trusted_source_degradation", "high",
			fmt.Sprintf("Trusted source %s accuracy below threshold: %.2f%%", metrics.SourceName, metrics.AccuracyScore*100))
	}

	// Check confidence threshold
	if metrics.ConfidenceScore < smat.config.TrustedSourceConfidenceThreshold {
		smat.createSecurityAlert(metrics, "confidence_degradation", "medium",
			fmt.Sprintf("Trusted source %s confidence below threshold: %.2f%%", metrics.SourceName, metrics.ConfidenceScore*100))
	}

	// Check latency threshold
	if metrics.AverageLatency > smat.config.TrustedSourceLatencyThreshold {
		smat.createSecurityAlert(metrics, "latency_degradation", "medium",
			fmt.Sprintf("Trusted source %s latency above threshold: %v", metrics.SourceName, metrics.AverageLatency))
	}
}

// checkWebsiteVerificationAlerts checks for website verification alerts
func (smat *SecurityMetricsAccuracyTracker) checkWebsiteVerificationAlerts(metrics *WebsiteVerificationMetrics) {
	// Check verification accuracy
	if metrics.VerificationAccuracy < smat.config.WebsiteVerificationAccuracyThreshold {
		smat.createWebsiteVerificationAlert(metrics, "verification_failure", "high",
			fmt.Sprintf("Website verification accuracy below threshold for %s: %.2f%%", metrics.Domain, metrics.VerificationAccuracy*100))
	}
}

// createSecurityAlert creates a security alert for a trusted data source
func (smat *SecurityMetricsAccuracyTracker) createSecurityAlert(metrics *TrustedDataSourceMetrics, alertType, severity, message string) {
	// Check cooldown and limits
	recentAlerts := smat.getRecentAlertsForSource(metrics.SourceName)
	if len(recentAlerts) >= smat.config.MaxSecurityAlertsPerSource {
		return
	}

	alert := &SecurityAlert{
		ID:        fmt.Sprintf("security_%s_%d", metrics.SourceName, time.Now().Unix()),
		AlertType: alertType,
		Severity:  severity,
		Source:    metrics.SourceName,
		Message:   message,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	metrics.SecurityAlerts = append(metrics.SecurityAlerts, alert)
}

// createWebsiteVerificationAlert creates a website verification alert
func (smat *SecurityMetricsAccuracyTracker) createWebsiteVerificationAlert(metrics *WebsiteVerificationMetrics, alertType, severity, message string) {
	// Implementation would be similar to createSecurityAlert
	// For brevity, not implementing here
}

// getRecentAlertsForSource gets recent alerts for a specific source
func (smat *SecurityMetricsAccuracyTracker) getRecentAlertsForSource(sourceName string) []*SecurityAlert {
	var recentAlerts []*SecurityAlert
	cutoff := time.Now().Add(-smat.config.SecurityAlertCooldownPeriod)

	if metrics, exists := smat.trustedDataSources[sourceName]; exists {
		for _, alert := range metrics.SecurityAlerts {
			if alert.Timestamp.After(cutoff) {
				recentAlerts = append(recentAlerts, alert)
			}
		}
	}

	return recentAlerts
}

// Copy methods for thread-safe access

func (smat *SecurityMetricsAccuracyTracker) copyTrustedDataSourceMetrics(metrics *TrustedDataSourceMetrics) *TrustedDataSourceMetrics {
	copy := &TrustedDataSourceMetrics{
		SourceName:                metrics.SourceName,
		SourceType:                metrics.SourceType,
		LastUpdated:               metrics.LastUpdated,
		TotalRequests:             metrics.TotalRequests,
		SuccessfulRequests:        metrics.SuccessfulRequests,
		AccuracyScore:             metrics.AccuracyScore,
		ConfidenceScore:           metrics.ConfidenceScore,
		AverageLatency:            metrics.AverageLatency,
		TrustLevel:                metrics.TrustLevel,
		VerificationRate:          metrics.VerificationRate,
		DataIntegrityScore:        metrics.DataIntegrityScore,
		AuthenticationSuccessRate: metrics.AuthenticationSuccessRate,
		Status:                    metrics.Status,
		LastSecurityCheck:         metrics.LastSecurityCheck,
		HistoricalAccuracy:        make([]*SecurityAccuracyDataPoint, len(metrics.HistoricalAccuracy)),
		HistoricalConfidence:      make([]*SecurityConfidenceDataPoint, len(metrics.HistoricalConfidence)),
		HistoricalLatency:         make([]*SecurityLatencyDataPoint, len(metrics.HistoricalLatency)),
		SecurityAlerts:            make([]*SecurityAlert, len(metrics.SecurityAlerts)),
	}

	// Copy historical data
	for i, point := range metrics.HistoricalAccuracy {
		copy.HistoricalAccuracy[i] = &SecurityAccuracyDataPoint{
			Timestamp:  point.Timestamp,
			Accuracy:   point.Accuracy,
			Confidence: point.Confidence,
			Source:     point.Source,
			RequestID:  point.RequestID,
		}
	}

	for i, point := range metrics.HistoricalConfidence {
		copy.HistoricalConfidence[i] = &SecurityConfidenceDataPoint{
			Timestamp:  point.Timestamp,
			Confidence: point.Confidence,
			Source:     point.Source,
			RequestID:  point.RequestID,
		}
	}

	for i, point := range metrics.HistoricalLatency {
		copy.HistoricalLatency[i] = &SecurityLatencyDataPoint{
			Timestamp: point.Timestamp,
			Latency:   point.Latency,
			Source:    point.Source,
			RequestID: point.RequestID,
		}
	}

	// Copy alerts
	for i, alert := range metrics.SecurityAlerts {
		copy.SecurityAlerts[i] = &SecurityAlert{
			ID:                  alert.ID,
			AlertType:           alert.AlertType,
			Severity:            alert.Severity,
			Source:              alert.Source,
			Message:             alert.Message,
			Timestamp:           alert.Timestamp,
			Resolved:            alert.Resolved,
			ResolutionTimestamp: alert.ResolutionTimestamp,
			ResolutionNotes:     alert.ResolutionNotes,
		}
	}

	return copy
}

func (smat *SecurityMetricsAccuracyTracker) copyWebsiteVerificationMetrics(metrics *WebsiteVerificationMetrics) *WebsiteVerificationMetrics {
	copy := &WebsiteVerificationMetrics{
		Domain:                  metrics.Domain,
		VerificationMethod:      metrics.VerificationMethod,
		LastUpdated:             metrics.LastUpdated,
		TotalVerifications:      metrics.TotalVerifications,
		SuccessfulVerifications: metrics.SuccessfulVerifications,
		VerificationAccuracy:    metrics.VerificationAccuracy,
		AverageVerificationTime: metrics.AverageVerificationTime,
		SSLValidityScore:        metrics.SSLValidityScore,
		DomainReputationScore:   metrics.DomainReputationScore,
		CertificateTrustScore:   metrics.CertificateTrustScore,
		Status:                  metrics.Status,
		LastVerification:        metrics.LastVerification,
		NextVerificationDue:     metrics.NextVerificationDue,
		HistoricalVerifications: make([]*WebsiteVerificationDataPoint, len(metrics.HistoricalVerifications)),
	}

	// Copy historical data
	for i, point := range metrics.HistoricalVerifications {
		copy.HistoricalVerifications[i] = &WebsiteVerificationDataPoint{
			Timestamp:        point.Timestamp,
			Success:          point.Success,
			VerificationTime: point.VerificationTime,
			Method:           point.Method,
			Domain:           point.Domain,
		}
	}

	return copy
}

func (smat *SecurityMetricsAccuracyTracker) copySecurityViolationMetrics(metrics *SecurityViolationMetrics) *SecurityViolationMetrics {
	copy := &SecurityViolationMetrics{
		ViolationType:        metrics.ViolationType,
		LastUpdated:          metrics.LastUpdated,
		TotalViolations:      metrics.TotalViolations,
		CriticalViolations:   metrics.CriticalViolations,
		ViolationRate:        metrics.ViolationRate,
		AverageSeverity:      metrics.AverageSeverity,
		DetectionAccuracy:    metrics.DetectionAccuracy,
		FalsePositiveRate:    metrics.FalsePositiveRate,
		ResponseTime:         metrics.ResponseTime,
		Status:               metrics.Status,
		LastViolation:        metrics.LastViolation,
		InvestigationStatus:  metrics.InvestigationStatus,
		HistoricalViolations: make([]*SecurityViolationDataPoint, len(metrics.HistoricalViolations)),
	}

	// Copy historical data
	for i, point := range metrics.HistoricalViolations {
		copy.HistoricalViolations[i] = &SecurityViolationDataPoint{
			Timestamp:     point.Timestamp,
			ViolationType: point.ViolationType,
			Severity:      point.Severity,
			Source:        point.Source,
			Description:   point.Description,
		}
	}

	return copy
}

func (smat *SecurityMetricsAccuracyTracker) copyConfidenceIntegrityMetrics(metrics *ConfidenceIntegrityMetrics) *ConfidenceIntegrityMetrics {
	copy := &ConfidenceIntegrityMetrics{
		LastUpdated:            metrics.LastUpdated,
		OverallIntegrityScore:  metrics.OverallIntegrityScore,
		DataConsistencyScore:   metrics.DataConsistencyScore,
		SourceReliabilityScore: metrics.SourceReliabilityScore,
		CrossValidationScore:   metrics.CrossValidationScore,
		AverageConfidence:      metrics.AverageConfidence,
		ConfidenceVariance:     metrics.ConfidenceVariance,
		LowConfidenceRate:      metrics.LowConfidenceRate,
		IntegrityStatus:        metrics.IntegrityStatus,
		LastIntegrityCheck:     metrics.LastIntegrityCheck,
		HistoricalIntegrity:    make([]*IntegrityDataPoint, len(metrics.HistoricalIntegrity)),
		HistoricalConfidence:   make([]*ConfidenceIntegrityDataPoint, len(metrics.HistoricalConfidence)),
	}

	// Copy historical data
	for i, point := range metrics.HistoricalIntegrity {
		copy.HistoricalIntegrity[i] = &IntegrityDataPoint{
			Timestamp:        point.Timestamp,
			IntegrityScore:   point.IntegrityScore,
			ConsistencyScore: point.ConsistencyScore,
			ReliabilityScore: point.ReliabilityScore,
		}
	}

	for i, point := range metrics.HistoricalConfidence {
		copy.HistoricalConfidence[i] = &ConfidenceIntegrityDataPoint{
			Timestamp:       point.Timestamp,
			Confidence:      point.Confidence,
			Integrity:       point.Integrity,
			CrossValidation: point.CrossValidation,
		}
	}

	return copy
}

// Supporting component types

// SecurityAnalyzer analyzes security metrics and calculates integrity scores
type SecurityAnalyzer struct {
	config *SecurityMetricsConfig
	logger *zap.Logger
}

// ThreatDetector detects threats and calculates detection metrics
type ThreatDetector struct {
	config *SecurityMetricsConfig
	logger *zap.Logger
}

// ComplianceMonitor monitors compliance with security standards
type ComplianceMonitor struct {
	config *SecurityMetricsConfig
	logger *zap.Logger
}

// SecurityAuditLogger logs security-related events for audit purposes
type SecurityAuditLogger struct {
	config *SecurityMetricsConfig
	logger *zap.Logger
}

// SecurityAnalyzer methods

// CalculateDataIntegrityScore calculates data integrity score for a trusted source
func (sa *SecurityAnalyzer) CalculateDataIntegrityScore(metrics *TrustedDataSourceMetrics) float64 {
	// Base integrity score from accuracy
	integrityScore := metrics.AccuracyScore

	// Adjust based on confidence consistency
	if len(metrics.HistoricalConfidence) > 10 {
		confidenceVariance := sa.calculateConfidenceVariance(metrics.HistoricalConfidence)
		// Lower variance = higher integrity
		integrityScore *= (1.0 - confidenceVariance)
	}

	// Adjust based on latency consistency
	if len(metrics.HistoricalLatency) > 10 {
		latencyVariance := sa.calculateLatencyVariance(metrics.HistoricalLatency)
		// Lower variance = higher integrity
		integrityScore *= (1.0 - latencyVariance)
	}

	return math.Max(0.0, math.Min(1.0, integrityScore))
}

// CalculateVerificationRate calculates verification rate for a trusted source
func (sa *SecurityAnalyzer) CalculateVerificationRate(metrics *TrustedDataSourceMetrics) float64 {
	if metrics.TotalRequests == 0 {
		return 0.0
	}

	// Calculate verification rate based on successful requests and trust level
	baseRate := float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)

	// Adjust based on trust level
	verificationRate := baseRate * metrics.TrustLevel

	return math.Max(0.0, math.Min(1.0, verificationRate))
}

// calculateConfidenceVariance calculates variance in confidence scores
func (sa *SecurityAnalyzer) calculateConfidenceVariance(data []*SecurityConfidenceDataPoint) float64 {
	if len(data) < 2 {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, point := range data {
		sum += point.Confidence
	}
	mean := sum / float64(len(data))

	// Calculate variance
	sumSquaredDiff := 0.0
	for _, point := range data {
		diff := point.Confidence - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(data)-1)
	return math.Sqrt(variance) // Return standard deviation
}

// calculateLatencyVariance calculates variance in latency
func (sa *SecurityAnalyzer) calculateLatencyVariance(data []*SecurityLatencyDataPoint) float64 {
	if len(data) < 2 {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, point := range data {
		sum += float64(point.Latency.Milliseconds())
	}
	mean := sum / float64(len(data))

	// Calculate variance
	sumSquaredDiff := 0.0
	for _, point := range data {
		diff := float64(point.Latency.Milliseconds()) - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(data)-1)
	return math.Sqrt(variance) / mean // Return coefficient of variation
}

// ThreatDetector methods

// CalculateDetectionAccuracy calculates threat detection accuracy
func (td *ThreatDetector) CalculateDetectionAccuracy(metrics *SecurityViolationMetrics) float64 {
	if metrics.TotalViolations == 0 {
		return 1.0 // No violations = perfect detection
	}

	// Base accuracy from critical violations detected
	baseAccuracy := float64(metrics.CriticalViolations) / float64(metrics.TotalViolations)

	// Adjust based on false positive rate
	adjustedAccuracy := baseAccuracy * (1.0 - metrics.FalsePositiveRate)

	return math.Max(0.0, math.Min(1.0, adjustedAccuracy))
}

// CalculateFalsePositiveRate calculates false positive rate for threat detection
func (td *ThreatDetector) CalculateFalsePositiveRate(metrics *SecurityViolationMetrics) float64 {
	if metrics.TotalViolations == 0 {
		return 0.0
	}

	// Estimate false positive rate based on severity distribution
	// Lower average severity might indicate more false positives
	estimatedFalsePositiveRate := math.Max(0.0, 1.0-metrics.AverageSeverity)

	return math.Min(1.0, estimatedFalsePositiveRate)
}

// ComplianceMonitor methods

// CheckCompliance checks compliance with security standards
func (cm *ComplianceMonitor) CheckCompliance(metrics *TrustedDataSourceMetrics) bool {
	// Check accuracy compliance
	if metrics.AccuracyScore < cm.config.ComplianceAccuracyThreshold {
		return false
	}

	// Check confidence compliance
	if metrics.ConfidenceScore < cm.config.TrustedSourceConfidenceThreshold {
		return false
	}

	// Check latency compliance
	if metrics.AverageLatency > cm.config.TrustedSourceLatencyThreshold {
		return false
	}

	// Check trust level
	if metrics.TrustLevel < 0.8 { // Minimum trust level
		return false
	}

	return true
}

// SecurityAuditLogger methods

// LogTrustedDataSourceAccess logs trusted data source access
func (sal *SecurityAuditLogger) LogTrustedDataSourceAccess(result *ClassificationResult) {
	sal.logger.Info("Trusted data source access",
		zap.String("request_id", result.ID),
		zap.String("source", sal.extractSourceName(result)),
		zap.String("method", result.ClassificationMethod),
		zap.Float64("confidence", result.ConfidenceScore),
		zap.Time("timestamp", result.Timestamp),
	)
}

// LogWebsiteVerification logs website verification
func (sal *SecurityAuditLogger) LogWebsiteVerification(domain, method string, success bool) {
	sal.logger.Info("Website verification",
		zap.String("domain", domain),
		zap.String("method", method),
		zap.Bool("success", success),
		zap.Time("timestamp", time.Now()),
	)
}

// LogSecurityViolation logs security violation
func (sal *SecurityAuditLogger) LogSecurityViolation(violationType, source string, severity float64) {
	sal.logger.Warn("Security violation detected",
		zap.String("violation_type", violationType),
		zap.String("source", source),
		zap.Float64("severity", severity),
		zap.Time("timestamp", time.Now()),
	)
}

// extractSourceName extracts source name from result metadata
func (sal *SecurityAuditLogger) extractSourceName(result *ClassificationResult) string {
	if source, exists := result.Metadata["trusted_source"]; exists {
		if s, ok := source.(string); ok {
			return s
		}
	}
	return "unknown_source"
}
