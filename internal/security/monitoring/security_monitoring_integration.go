package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kyb-platform/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SecurityMonitoringIntegration integrates advanced security monitoring with the classification system
type SecurityMonitoringIntegration struct {
	// Core components
	advancedMonitor *AdvancedSecurityMonitor
	securityMonitor *SecurityMonitor

	// Configuration
	config *SecurityMonitoringIntegrationConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Integration state
	integrated bool
	mux        sync.RWMutex

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// SecurityMonitoringIntegrationConfig configuration for security monitoring integration
type SecurityMonitoringIntegrationConfig struct {
	// Integration settings
	EnableAdvancedMonitoring    bool          `json:"enable_advanced_monitoring" yaml:"enable_advanced_monitoring"`
	EnableSecurityEventLogging  bool          `json:"enable_security_event_logging" yaml:"enable_security_event_logging"`
	EnablePerformanceMonitoring bool          `json:"enable_performance_monitoring" yaml:"enable_performance_monitoring"`
	IntegrationCheckInterval    time.Duration `json:"integration_check_interval" yaml:"integration_check_interval"`

	// Data source trust settings
	TrustedDataSources       []string `json:"trusted_data_sources" yaml:"trusted_data_sources"`
	UntrustedDataSources     []string `json:"untrusted_data_sources" yaml:"untrusted_data_sources"`
	DataSourceTrustThreshold float64  `json:"data_source_trust_threshold" yaml:"data_source_trust_threshold"`

	// Website verification settings
	WebsiteVerificationEnabled bool          `json:"website_verification_enabled" yaml:"website_verification_enabled"`
	WebsiteVerificationTimeout time.Duration `json:"website_verification_timeout" yaml:"website_verification_timeout"`
	WebsiteVerificationRetries int           `json:"website_verification_retries" yaml:"website_verification_retries"`

	// Confidence score settings
	ConfidenceScoreValidationEnabled bool    `json:"confidence_score_validation_enabled" yaml:"confidence_score_validation_enabled"`
	ConfidenceScoreThreshold         float64 `json:"confidence_score_threshold" yaml:"confidence_score_threshold"`
	ConfidenceScoreAnomalyThreshold  float64 `json:"confidence_score_anomaly_threshold" yaml:"confidence_score_anomaly_threshold"`

	// Alerting settings
	EnableRealTimeAlerts     bool          `json:"enable_real_time_alerts" yaml:"enable_real_time_alerts"`
	AlertProcessingInterval  time.Duration `json:"alert_processing_interval" yaml:"alert_processing_interval"`
	AlertEscalationThreshold int           `json:"alert_escalation_threshold" yaml:"alert_escalation_threshold"`

	// External integrations
	WebhookURL     string        `json:"webhook_url" yaml:"webhook_url"`
	WebhookTimeout time.Duration `json:"webhook_timeout" yaml:"webhook_timeout"`
	SlackWebhook   string        `json:"slack_webhook" yaml:"slack_webhook"`
	EmailAlerts    []string      `json:"email_alerts" yaml:"email_alerts"`
}

// ClassificationSecurityContext represents security context for classification operations
type ClassificationSecurityContext struct {
	RequestID             string                 `json:"request_id"`
	UserID                string                 `json:"user_id,omitempty"`
	IPAddress             string                 `json:"ip_address,omitempty"`
	UserAgent             string                 `json:"user_agent,omitempty"`
	BusinessName          string                 `json:"business_name"`
	BusinessDescription   string                 `json:"business_description"`
	WebsiteURL            string                 `json:"website_url,omitempty"`
	DataSources           []string               `json:"data_sources"`
	ClassificationMethods []string               `json:"classification_methods"`
	ConfidenceScores      map[string]float64     `json:"confidence_scores"`
	SecurityFlags         map[string]bool        `json:"security_flags"`
	Timestamp             time.Time              `json:"timestamp"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// SecurityValidationResult represents the result of security validation
type SecurityValidationResult struct {
	Valid               bool                     `json:"valid"`
	TrustedDataSources  bool                     `json:"trusted_data_sources"`
	WebsiteVerified     bool                     `json:"website_verified"`
	ConfidenceIntegrity bool                     `json:"confidence_integrity"`
	SecurityScore       float64                  `json:"security_score"`
	Violations          []SecurityViolation      `json:"violations"`
	Warnings            []SecurityWarning        `json:"warnings"`
	Recommendations     []SecurityRecommendation `json:"recommendations"`
	Timestamp           time.Time                `json:"timestamp"`
}

// SecurityViolation represents a security violation
type SecurityViolation struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SecurityWarning represents a security warning
type SecurityWarning struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SecurityRecommendation represents a security recommendation
type SecurityRecommendation struct {
	Type        string                 `json:"type"`
	Priority    string                 `json:"priority"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewSecurityMonitoringIntegration creates a new security monitoring integration
func NewSecurityMonitoringIntegration(
	config *SecurityMonitoringIntegrationConfig,
	advancedMonitor *AdvancedSecurityMonitor,
	securityMonitor *SecurityMonitor,
	logger *observability.Logger,
	tracer trace.Tracer,
) *SecurityMonitoringIntegration {
	if config == nil {
		config = &SecurityMonitoringIntegrationConfig{
			EnableAdvancedMonitoring:    true,
			EnableSecurityEventLogging:  true,
			EnablePerformanceMonitoring: true,
			IntegrationCheckInterval:    30 * time.Second,

			TrustedDataSources:       []string{"supabase", "government_apis", "verified_sources"},
			UntrustedDataSources:     []string{"unverified_apis", "suspicious_sources"},
			DataSourceTrustThreshold: 95.0,

			WebsiteVerificationEnabled: true,
			WebsiteVerificationTimeout: 10 * time.Second,
			WebsiteVerificationRetries: 3,

			ConfidenceScoreValidationEnabled: true,
			ConfidenceScoreThreshold:         0.8,
			ConfidenceScoreAnomalyThreshold:  0.1,

			EnableRealTimeAlerts:     true,
			AlertProcessingInterval:  1 * time.Minute,
			AlertEscalationThreshold: 5,

			WebhookTimeout: 10 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	integration := &SecurityMonitoringIntegration{
		advancedMonitor: advancedMonitor,
		securityMonitor: securityMonitor,
		config:          config,
		logger:          logger,
		tracer:          tracer,
		integrated:      false,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start integration workers
	integration.startIntegrationWorkers()

	return integration
}

// startIntegrationWorkers starts background integration workers
func (smi *SecurityMonitoringIntegration) startIntegrationWorkers() {
	// Integration monitoring worker
	go smi.integrationMonitoringWorker()

	// Alert processing worker
	go smi.alertProcessingWorker()

	// Security validation worker
	go smi.securityValidationWorker()
}

// integrationMonitoringWorker monitors integration health
func (smi *SecurityMonitoringIntegration) integrationMonitoringWorker() {
	ticker := time.NewTicker(smi.config.IntegrationCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-smi.ctx.Done():
			return
		case <-ticker.C:
			smi.monitorIntegrationHealth()
		}
	}
}

// alertProcessingWorker processes security alerts
func (smi *SecurityMonitoringIntegration) alertProcessingWorker() {
	ticker := time.NewTicker(smi.config.AlertProcessingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-smi.ctx.Done():
			return
		case <-ticker.C:
			smi.processSecurityAlerts()
		}
	}
}

// securityValidationWorker performs security validation
func (smi *SecurityMonitoringIntegration) securityValidationWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-smi.ctx.Done():
			return
		case <-ticker.C:
			smi.performSecurityValidation()
		}
	}
}

// ValidateClassificationSecurity validates security for a classification operation
func (smi *SecurityMonitoringIntegration) ValidateClassificationSecurity(ctx context.Context, securityContext *ClassificationSecurityContext) (*SecurityValidationResult, error) {
	_, span := smi.tracer.Start(ctx, "SecurityMonitoringIntegration.ValidateClassificationSecurity")
	defer span.End()

	result := &SecurityValidationResult{
		Valid:               true,
		TrustedDataSources:  true,
		WebsiteVerified:     true,
		ConfidenceIntegrity: true,
		SecurityScore:       100.0,
		Violations:          make([]SecurityViolation, 0),
		Warnings:            make([]SecurityWarning, 0),
		Recommendations:     make([]SecurityRecommendation, 0),
		Timestamp:           time.Now(),
	}

	// Validate data sources
	if err := smi.validateDataSources(securityContext, result); err != nil {
		smi.logger.Error("data source validation failed", map[string]interface{}{
			"error":      err.Error(),
			"request_id": securityContext.RequestID,
		})
	}

	// Validate website verification
	if err := smi.validateWebsiteVerification(securityContext, result); err != nil {
		smi.logger.Error("website verification validation failed", map[string]interface{}{
			"error":      err.Error(),
			"request_id": securityContext.RequestID,
		})
	}

	// Validate confidence scores
	if err := smi.validateConfidenceScores(securityContext, result); err != nil {
		smi.logger.Error("confidence score validation failed", map[string]interface{}{
			"error":      err.Error(),
			"request_id": securityContext.RequestID,
		})
	}

	// Calculate overall security score
	smi.calculateSecurityScore(result)

	// Record security validation event
	smi.recordSecurityValidationEvent(securityContext, result)

	span.SetAttributes(
		attribute.String("request_id", securityContext.RequestID),
		attribute.Bool("valid", result.Valid),
		attribute.Float64("security_score", result.SecurityScore),
		attribute.Int("violations", len(result.Violations)),
		attribute.Int("warnings", len(result.Warnings)),
	)

	return result, nil
}

// validateDataSources validates data source trust
func (smi *SecurityMonitoringIntegration) validateDataSources(securityContext *ClassificationSecurityContext, result *SecurityValidationResult) error {
	for _, dataSource := range securityContext.DataSources {
		// Check if data source is in untrusted list
		for _, untrustedSource := range smi.config.UntrustedDataSources {
			if dataSource == untrustedSource {
				violation := SecurityViolation{
					Type:        "untrusted_data_source",
					Severity:    "high",
					Description: fmt.Sprintf("Untrusted data source detected: %s", dataSource),
					Details: map[string]interface{}{
						"data_source": dataSource,
						"request_id":  securityContext.RequestID,
					},
					Timestamp: time.Now(),
				}
				result.Violations = append(result.Violations, violation)
				result.TrustedDataSources = false
				result.Valid = false

				// Record in advanced monitor
				smi.advancedMonitor.RecordDataSourceRequest(dataSource, dataSource, false)
				smi.advancedMonitor.RecordSecurityViolation(
					ViolationTypeUntrustedDataSource,
					ViolationSeverityHigh,
					"security_validation",
					fmt.Sprintf("Untrusted data source: %s", dataSource),
					map[string]interface{}{
						"request_id":  securityContext.RequestID,
						"data_source": dataSource,
					},
				)
			}
		}

		// Check if data source is in trusted list
		trusted := false
		for _, trustedSource := range smi.config.TrustedDataSources {
			if dataSource == trustedSource {
				trusted = true
				break
			}
		}

		if trusted {
			smi.advancedMonitor.RecordDataSourceRequest(dataSource, dataSource, true)
		} else {
			// Unknown data source - add warning
			warning := SecurityWarning{
				Type:        "unknown_data_source",
				Description: fmt.Sprintf("Unknown data source: %s", dataSource),
				Details: map[string]interface{}{
					"data_source": dataSource,
					"request_id":  securityContext.RequestID,
				},
				Timestamp: time.Now(),
			}
			result.Warnings = append(result.Warnings, warning)
		}
	}

	return nil
}

// validateWebsiteVerification validates website verification
func (smi *SecurityMonitoringIntegration) validateWebsiteVerification(securityContext *ClassificationSecurityContext, result *SecurityValidationResult) error {
	if !smi.config.WebsiteVerificationEnabled || securityContext.WebsiteURL == "" {
		return nil
	}

	// Extract domain from URL
	domain := smi.extractDomainFromURL(securityContext.WebsiteURL)
	if domain == "" {
		result.Warnings = append(result.Warnings, SecurityWarning{
			Type:        "invalid_website_url",
			Description: fmt.Sprintf("Invalid website URL: %s", securityContext.WebsiteURL),
			Details: map[string]interface{}{
				"website_url": securityContext.WebsiteURL,
				"request_id":  securityContext.RequestID,
			},
			Timestamp: time.Now(),
		})
		return nil
	}

	// Perform website verification
	verified, err := smi.performWebsiteVerification(domain)
	if err != nil {
		smi.logger.Error("website verification failed", map[string]interface{}{
			"error":      err.Error(),
			"domain":     domain,
			"request_id": securityContext.RequestID,
		})
		return err
	}

	// Record verification result
	smi.advancedMonitor.RecordWebsiteVerification(domain, verified)

	if !verified {
		violation := SecurityViolation{
			Type:        "website_verification_failure",
			Severity:    "medium",
			Description: fmt.Sprintf("Website verification failed for domain: %s", domain),
			Details: map[string]interface{}{
				"domain":      domain,
				"website_url": securityContext.WebsiteURL,
				"request_id":  securityContext.RequestID,
			},
			Timestamp: time.Now(),
		}
		result.Violations = append(result.Violations, violation)
		result.WebsiteVerified = false
		result.Valid = false

		// Record in advanced monitor
		smi.advancedMonitor.RecordSecurityViolation(
			ViolationTypeWebsiteVerificationFailure,
			ViolationSeverityMedium,
			"website_verification",
			fmt.Sprintf("Website verification failed: %s", domain),
			map[string]interface{}{
				"request_id":  securityContext.RequestID,
				"domain":      domain,
				"website_url": securityContext.WebsiteURL,
			},
		)
	}

	return nil
}

// validateConfidenceScores validates confidence score integrity
func (smi *SecurityMonitoringIntegration) validateConfidenceScores(securityContext *ClassificationSecurityContext, result *SecurityValidationResult) error {
	if !smi.config.ConfidenceScoreValidationEnabled {
		return nil
	}

	for method, score := range securityContext.ConfidenceScores {
		// Check if score is within valid range
		if score < 0.0 || score > 1.0 {
			violation := SecurityViolation{
				Type:        "invalid_confidence_score",
				Severity:    "high",
				Description: fmt.Sprintf("Invalid confidence score for method %s: %.3f", method, score),
				Details: map[string]interface{}{
					"method":     method,
					"score":      score,
					"request_id": securityContext.RequestID,
				},
				Timestamp: time.Now(),
			}
			result.Violations = append(result.Violations, violation)
			result.ConfidenceIntegrity = false
			result.Valid = false

			// Record in advanced monitor
			smi.advancedMonitor.RecordConfidenceIntegrityEvent(
				ConfidenceEventTypeOutOfRange,
				ConfidenceEventSeverityHigh,
				securityContext.RequestID,
				0.5, // Expected score
				score,
				map[string]interface{}{
					"method":     method,
					"request_id": securityContext.RequestID,
				},
			)
		}

		// Check for confidence score anomalies
		expectedScore := smi.calculateExpectedConfidenceScore(method, securityContext)
		scoreDifference := score - expectedScore
		if scoreDifference < 0 {
			scoreDifference = -scoreDifference
		}

		if scoreDifference > smi.config.ConfidenceScoreAnomalyThreshold {
			violation := SecurityViolation{
				Type:        "confidence_score_anomaly",
				Severity:    "medium",
				Description: fmt.Sprintf("Confidence score anomaly for method %s: expected %.3f, got %.3f", method, expectedScore, score),
				Details: map[string]interface{}{
					"method":           method,
					"expected_score":   expectedScore,
					"actual_score":     score,
					"score_difference": scoreDifference,
					"request_id":       securityContext.RequestID,
				},
				Timestamp: time.Now(),
			}
			result.Violations = append(result.Violations, violation)
			result.ConfidenceIntegrity = false

			// Record in advanced monitor
			smi.advancedMonitor.RecordConfidenceIntegrityEvent(
				ConfidenceEventTypeAnomaly,
				ConfidenceEventSeverityMedium,
				securityContext.RequestID,
				expectedScore,
				score,
				map[string]interface{}{
					"method":     method,
					"request_id": securityContext.RequestID,
				},
			)
		}
	}

	return nil
}

// calculateSecurityScore calculates the overall security score
func (smi *SecurityMonitoringIntegration) calculateSecurityScore(result *SecurityValidationResult) {
	score := 100.0

	// Deduct points for violations
	for _, violation := range result.Violations {
		switch violation.Severity {
		case "critical":
			score -= 25.0
		case "high":
			score -= 15.0
		case "medium":
			score -= 10.0
		case "low":
			score -= 5.0
		}
	}

	// Deduct points for warnings
	for range result.Warnings {
		score -= 2.0
	}

	// Ensure score doesn't go below 0
	if score < 0.0 {
		score = 0.0
	}

	result.SecurityScore = score
}

// recordSecurityValidationEvent records the security validation event
func (smi *SecurityMonitoringIntegration) recordSecurityValidationEvent(securityContext *ClassificationSecurityContext, result *SecurityValidationResult) {
	// Record in security monitor
	event := &SecurityEvent{
		ID:        generateEventID(),
		Type:      EventTypeDataAccess,
		Severity:  smi.calculateEventSeverity(result),
		Source:    "security_validation",
		UserID:    securityContext.UserID,
		IPAddress: securityContext.IPAddress,
		UserAgent: securityContext.UserAgent,
		Endpoint:  "/api/classify",
		Method:    "POST",
		Details: map[string]interface{}{
			"request_id":             securityContext.RequestID,
			"business_name":          securityContext.BusinessName,
			"website_url":            securityContext.WebsiteURL,
			"data_sources":           securityContext.DataSources,
			"classification_methods": securityContext.ClassificationMethods,
			"confidence_scores":      securityContext.ConfidenceScores,
			"security_score":         result.SecurityScore,
			"violations":             len(result.Violations),
			"warnings":               len(result.Warnings),
			"valid":                  result.Valid,
		},
		Timestamp: time.Now(),
		Resolved:  false,
	}

	smi.securityMonitor.RecordEvent(event)
}

// Helper methods

func (smi *SecurityMonitoringIntegration) calculateEventSeverity(result *SecurityValidationResult) SecurityEventSeverity {
	if result.SecurityScore >= 90.0 {
		return SeverityInfo
	} else if result.SecurityScore >= 75.0 {
		return SeverityLow
	} else if result.SecurityScore >= 50.0 {
		return SeverityMedium
	} else if result.SecurityScore >= 25.0 {
		return SeverityHigh
	} else {
		return SeverityCritical
	}
}

func (smi *SecurityMonitoringIntegration) extractDomainFromURL(url string) string {
	// Simplified domain extraction - in production, use proper URL parsing
	// This is a placeholder implementation
	if len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://") {
		// Remove protocol
		if url[:7] == "http://" {
			url = url[7:]
		} else {
			url = url[8:]
		}
		// Find first slash or end of string
		for i, char := range url {
			if char == '/' {
				return url[:i]
			}
		}
		return url
	}
	return ""
}

func (smi *SecurityMonitoringIntegration) performWebsiteVerification(domain string) (bool, error) {
	// Simplified website verification - in production, implement proper verification
	// This is a placeholder implementation
	time.Sleep(100 * time.Millisecond) // Simulate verification delay

	// For testing purposes, consider certain domains as verified
	verifiedDomains := []string{"example.com", "google.com", "github.com", "stackoverflow.com"}
	for _, verifiedDomain := range verifiedDomains {
		if domain == verifiedDomain {
			return true, nil
		}
	}

	// For testing purposes, consider certain domains as unverified
	unverifiedDomains := []string{"suspicious-site.com", "malicious-domain.net", "fake-website.org"}
	for _, unverifiedDomain := range unverifiedDomains {
		if domain == unverifiedDomain {
			return false, nil
		}
	}

	// Default to verified for unknown domains in testing
	return true, nil
}

func (smi *SecurityMonitoringIntegration) calculateExpectedConfidenceScore(method string, securityContext *ClassificationSecurityContext) float64 {
	// Simplified expected confidence score calculation
	// In production, this would be based on historical data and method characteristics

	baseScore := 0.8

	// Adjust based on business name length
	if len(securityContext.BusinessName) > 10 {
		baseScore += 0.1
	}

	// Adjust based on description length
	if len(securityContext.BusinessDescription) > 50 {
		baseScore += 0.1
	}

	// Adjust based on method
	switch method {
	case "keyword_matching":
		baseScore = 0.7
	case "ml_classification":
		baseScore = 0.85
	case "description_analysis":
		baseScore = 0.75
	default:
		baseScore = 0.8
	}

	// Ensure score is within valid range
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	if baseScore < 0.0 {
		baseScore = 0.0
	}

	return baseScore
}

// Monitoring worker implementations

func (smi *SecurityMonitoringIntegration) monitorIntegrationHealth() {
	// Implementation would monitor integration health
	// For now, this is a placeholder
}

func (smi *SecurityMonitoringIntegration) processSecurityAlerts() {
	// Implementation would process security alerts
	// For now, this is a placeholder
}

func (smi *SecurityMonitoringIntegration) performSecurityValidation() {
	// Implementation would perform periodic security validation
	// For now, this is a placeholder
}

// Getter methods

func (smi *SecurityMonitoringIntegration) GetAdvancedMonitor() *AdvancedSecurityMonitor {
	return smi.advancedMonitor
}

func (smi *SecurityMonitoringIntegration) GetSecurityMonitor() *SecurityMonitor {
	return smi.securityMonitor
}

func (smi *SecurityMonitoringIntegration) GetConfig() *SecurityMonitoringIntegrationConfig {
	return smi.config
}

func (smi *SecurityMonitoringIntegration) IsIntegrated() bool {
	smi.mux.RLock()
	defer smi.mux.RUnlock()
	return smi.integrated
}

// Shutdown shuts down the security monitoring integration
func (smi *SecurityMonitoringIntegration) Shutdown() {
	smi.cancel()
	smi.logger.Info("security monitoring integration shutting down", map[string]interface{}{})
}
