package feedback

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SecurityPatternDetector detects security patterns in feedback data
type SecurityPatternDetector struct {
	config *SecurityAnalysisConfig
	logger *zap.Logger
}

// NewSecurityPatternDetector creates a new security pattern detector
func NewSecurityPatternDetector(config *SecurityAnalysisConfig, logger *zap.Logger) *SecurityPatternDetector {
	return &SecurityPatternDetector{
		config: config,
		logger: logger,
	}
}

// DetectPatterns detects security patterns in the provided feedback
func (spd *SecurityPatternDetector) DetectPatterns(ctx context.Context, feedback []*UserFeedback) ([]*SecurityPattern, error) {
	spd.logger.Info("Starting security pattern detection",
		zap.Int("feedback_count", len(feedback)))

	var patterns []*SecurityPattern

	// Detect violation patterns
	violationPatterns, err := spd.detectViolationPatterns(feedback)
	if err != nil {
		spd.logger.Error("Failed to detect violation patterns", zap.Error(err))
	} else {
		patterns = append(patterns, violationPatterns...)
	}

	// Detect source trust patterns
	trustPatterns, err := spd.detectTrustPatterns(feedback)
	if err != nil {
		spd.logger.Error("Failed to detect trust patterns", zap.Error(err))
	} else {
		patterns = append(patterns, trustPatterns...)
	}

	// Detect verification patterns
	verificationPatterns, err := spd.detectVerificationPatterns(feedback)
	if err != nil {
		spd.logger.Error("Failed to detect verification patterns", zap.Error(err))
	} else {
		patterns = append(patterns, verificationPatterns...)
	}

	// Detect performance patterns
	performancePatterns, err := spd.detectPerformancePatterns(feedback)
	if err != nil {
		spd.logger.Error("Failed to detect performance patterns", zap.Error(err))
	} else {
		patterns = append(patterns, performancePatterns...)
	}

	// Filter patterns by confidence and occurrences
	filteredPatterns := spd.filterPatterns(patterns)

	// Sort patterns by severity and confidence
	sort.Slice(filteredPatterns, func(i, j int) bool {
		severityOrder := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}
		if severityOrder[filteredPatterns[i].Severity] == severityOrder[filteredPatterns[j].Severity] {
			return filteredPatterns[i].Confidence > filteredPatterns[j].Confidence
		}
		return severityOrder[filteredPatterns[i].Severity] > severityOrder[filteredPatterns[j].Severity]
	})

	spd.logger.Info("Security pattern detection completed",
		zap.Int("total_patterns", len(patterns)),
		zap.Int("filtered_patterns", len(filteredPatterns)))

	return filteredPatterns, nil
}

// detectViolationPatterns detects security violation patterns
func (spd *SecurityPatternDetector) detectViolationPatterns(feedback []*UserFeedback) ([]*SecurityPattern, error) {
	var patterns []*SecurityPattern

	// Group feedback by violation type
	violationGroups := make(map[string][]*UserFeedback)
	for _, fb := range feedback {
		if fb.FeedbackType == FeedbackTypeSecurityValidation {
			violationType := spd.determineViolationType(fb)
			if violationType != "" {
				violationGroups[violationType] = append(violationGroups[violationType], fb)
			}
		}
	}

	// Create patterns for each violation type
	for violationType, violationFb := range violationGroups {
		if len(violationFb) >= spd.config.PatternMinOccurrences {
			pattern := spd.createViolationPattern(violationType, violationFb)
			if pattern != nil && pattern.Confidence >= spd.config.PatternMinConfidence {
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns, nil
}

// detectTrustPatterns detects trusted source patterns
func (spd *SecurityPatternDetector) detectTrustPatterns(feedback []*UserFeedback) ([]*SecurityPattern, error) {
	var patterns []*SecurityPattern

	// Group feedback by source type
	sourceGroups := make(map[string][]*UserFeedback)
	for _, fb := range feedback {
		if fb.FeedbackType == FeedbackTypeDataSourceTrust {
			sourceType := spd.determineSourceType(fb)
			if sourceType != "" {
				sourceGroups[sourceType] = append(sourceGroups[sourceType], fb)
			}
		}
	}

	// Create patterns for each source type
	for sourceType, sourceFb := range sourceGroups {
		if len(sourceFb) >= spd.config.PatternMinOccurrences {
			pattern := spd.createTrustPattern(sourceType, sourceFb)
			if pattern != nil && pattern.Confidence >= spd.config.PatternMinConfidence {
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns, nil
}

// detectVerificationPatterns detects website verification patterns
func (spd *SecurityPatternDetector) detectVerificationPatterns(feedback []*UserFeedback) ([]*SecurityPattern, error) {
	var patterns []*SecurityPattern

	// Group feedback by verification type
	verificationGroups := make(map[string][]*UserFeedback)
	for _, fb := range feedback {
		if fb.FeedbackType == FeedbackTypeWebsiteVerification {
			verificationType := spd.determineVerificationType(fb)
			if verificationType != "" {
				verificationGroups[verificationType] = append(verificationGroups[verificationType], fb)
			}
		}
	}

	// Create patterns for each verification type
	for verificationType, verificationFb := range verificationGroups {
		if len(verificationFb) >= spd.config.PatternMinOccurrences {
			pattern := spd.createVerificationPattern(verificationType, verificationFb)
			if pattern != nil && pattern.Confidence >= spd.config.PatternMinConfidence {
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns, nil
}

// detectPerformancePatterns detects performance-related patterns
func (spd *SecurityPatternDetector) detectPerformancePatterns(feedback []*UserFeedback) ([]*SecurityPattern, error) {
	var patterns []*SecurityPattern

	// Group feedback by performance characteristics
	performanceGroups := make(map[string][]*UserFeedback)
	for _, fb := range feedback {
		performanceType := spd.determinePerformanceType(fb)
		if performanceType != "" {
			performanceGroups[performanceType] = append(performanceGroups[performanceType], fb)
		}
	}

	// Create patterns for each performance type
	for performanceType, performanceFb := range performanceGroups {
		if len(performanceFb) >= spd.config.PatternMinOccurrences {
			pattern := spd.createPerformancePattern(performanceType, performanceFb)
			if pattern != nil && pattern.Confidence >= spd.config.PatternMinConfidence {
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns, nil
}

// determineViolationType determines the type of security violation
func (spd *SecurityPatternDetector) determineViolationType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.FeedbackText)

	// Check for specific violation types
	if strings.Contains(text, "untrusted") || strings.Contains(text, "unverified") {
		return "untrusted_data_source"
	} else if strings.Contains(text, "malicious") || strings.Contains(text, "attack") {
		return "malicious_input"
	} else if strings.Contains(text, "injection") || strings.Contains(text, "sql") {
		return "injection_attack"
	} else if strings.Contains(text, "xss") || strings.Contains(text, "cross-site") {
		return "xss_attack"
	} else if strings.Contains(text, "phishing") || strings.Contains(text, "fake") {
		return "phishing_attempt"
	} else if strings.Contains(text, "spoofing") || strings.Contains(text, "impersonation") {
		return "spoofing_attempt"
	} else if strings.Contains(text, "data_leak") || strings.Contains(text, "exposure") {
		return "data_exposure"
	} else if strings.Contains(text, "unauthorized") || strings.Contains(text, "access") {
		return "unauthorized_access"
	}

	return ""
}

// determineSourceType determines the type of data source
func (spd *SecurityPatternDetector) determineSourceType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.FeedbackText)

	// Check for specific source types
	if strings.Contains(text, "government") || strings.Contains(text, "sec") || strings.Contains(text, "edgar") {
		return "government_registry"
	} else if strings.Contains(text, "whois") || strings.Contains(text, "domain") {
		return "domain_registry"
	} else if strings.Contains(text, "ssl") || strings.Contains(text, "certificate") {
		return "ssl_certificate"
	} else if strings.Contains(text, "api") || strings.Contains(text, "service") {
		return "external_api"
	} else if strings.Contains(text, "database") || strings.Contains(text, "internal") {
		return "internal_database"
	} else if strings.Contains(text, "website") || strings.Contains(text, "web") {
		return "website_analysis"
	}

	return "unknown_source"
}

// determineVerificationType determines the type of website verification
func (spd *SecurityPatternDetector) determineVerificationType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.FeedbackText)

	// Check for specific verification types
	if strings.Contains(text, "ssl") || strings.Contains(text, "certificate") {
		return "ssl_verification"
	} else if strings.Contains(text, "dns") || strings.Contains(text, "domain") {
		return "dns_verification"
	} else if strings.Contains(text, "whois") || strings.Contains(text, "registration") {
		return "whois_verification"
	} else if strings.Contains(text, "ownership") || strings.Contains(text, "control") {
		return "ownership_verification"
	} else if strings.Contains(text, "content") || strings.Contains(text, "analysis") {
		return "content_verification"
	}

	return "general_verification"
}

// determinePerformanceType determines the type of performance issue
func (spd *SecurityPatternDetector) determinePerformanceType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.FeedbackText)

	// Check for specific performance types
	if strings.Contains(text, "slow") || strings.Contains(text, "timeout") {
		return "slow_response"
	} else if strings.Contains(text, "unavailable") || strings.Contains(text, "down") {
		return "service_unavailable"
	} else if strings.Contains(text, "error") || strings.Contains(text, "failed") {
		return "error_rate_high"
	} else if strings.Contains(text, "memory") || strings.Contains(text, "cpu") {
		return "resource_exhaustion"
	}

	// Check processing time
	if feedback.ProcessingTimeMs > 2000 { // 2 seconds
		return "slow_processing"
	} else if feedback.ProcessingTimeMs > 1000 { // 1 second
		return "moderate_processing"
	}

	return ""
}

// createViolationPattern creates a security violation pattern
func (spd *SecurityPatternDetector) createViolationPattern(violationType string, feedback []*UserFeedback) *SecurityPattern {
	if len(feedback) == 0 {
		return nil
	}

	// Calculate confidence based on frequency and consistency
	confidence := spd.calculatePatternConfidence(feedback)

	// Determine severity
	severity := spd.determineViolationSeverity(violationType, len(feedback))

	// Get time range
	firstSeen, lastSeen := spd.getTimeRange(feedback)

	// Generate description
	description := spd.generateViolationPatternDescription(violationType, len(feedback))

	// Get affected components
	affectedComponents := spd.getAffectedComponents(feedback)

	// Generate recommendations
	recommendations := spd.generateViolationRecommendations(violationType)

	return &SecurityPattern{
		PatternID:          fmt.Sprintf("violation_pattern_%s_%d", violationType, time.Now().Unix()),
		PatternType:        "security_violation",
		Description:        description,
		Confidence:         confidence,
		Occurrences:        len(feedback),
		FirstSeen:          firstSeen,
		LastSeen:           lastSeen,
		AffectedComponents: affectedComponents,
		Severity:           severity,
		PatternData: map[string]interface{}{
			"violation_type": violationType,
			"frequency":      len(feedback),
			"avg_confidence": spd.calculateAverageConfidence(feedback),
		},
		Recommendations: recommendations,
	}
}

// createTrustPattern creates a trusted source pattern
func (spd *SecurityPatternDetector) createTrustPattern(sourceType string, feedback []*UserFeedback) *SecurityPattern {
	if len(feedback) == 0 {
		return nil
	}

	// Calculate confidence based on frequency and consistency
	confidence := spd.calculatePatternConfidence(feedback)

	// Determine severity based on trust issues
	severity := spd.determineTrustSeverity(sourceType, len(feedback))

	// Get time range
	firstSeen, lastSeen := spd.getTimeRange(feedback)

	// Generate description
	description := spd.generateTrustPatternDescription(sourceType, len(feedback))

	// Get affected components
	affectedComponents := spd.getAffectedComponents(feedback)

	// Generate recommendations
	recommendations := spd.generateTrustRecommendations(sourceType)

	return &SecurityPattern{
		PatternID:          fmt.Sprintf("trust_pattern_%s_%d", sourceType, time.Now().Unix()),
		PatternType:        "trusted_source",
		Description:        description,
		Confidence:         confidence,
		Occurrences:        len(feedback),
		FirstSeen:          firstSeen,
		LastSeen:           lastSeen,
		AffectedComponents: affectedComponents,
		Severity:           severity,
		PatternData: map[string]interface{}{
			"source_type":    sourceType,
			"frequency":      len(feedback),
			"avg_confidence": spd.calculateAverageConfidence(feedback),
		},
		Recommendations: recommendations,
	}
}

// createVerificationPattern creates a website verification pattern
func (spd *SecurityPatternDetector) createVerificationPattern(verificationType string, feedback []*UserFeedback) *SecurityPattern {
	if len(feedback) == 0 {
		return nil
	}

	// Calculate confidence based on frequency and consistency
	confidence := spd.calculatePatternConfidence(feedback)

	// Determine severity based on verification issues
	severity := spd.determineVerificationSeverity(verificationType, len(feedback))

	// Get time range
	firstSeen, lastSeen := spd.getTimeRange(feedback)

	// Generate description
	description := spd.generateVerificationPatternDescription(verificationType, len(feedback))

	// Get affected components
	affectedComponents := spd.getAffectedComponents(feedback)

	// Generate recommendations
	recommendations := spd.generateVerificationRecommendations(verificationType)

	return &SecurityPattern{
		PatternID:          fmt.Sprintf("verification_pattern_%s_%d", verificationType, time.Now().Unix()),
		PatternType:        "website_verification",
		Description:        description,
		Confidence:         confidence,
		Occurrences:        len(feedback),
		FirstSeen:          firstSeen,
		LastSeen:           lastSeen,
		AffectedComponents: affectedComponents,
		Severity:           severity,
		PatternData: map[string]interface{}{
			"verification_type": verificationType,
			"frequency":         len(feedback),
			"avg_confidence":    spd.calculateAverageConfidence(feedback),
		},
		Recommendations: recommendations,
	}
}

// createPerformancePattern creates a performance pattern
func (spd *SecurityPatternDetector) createPerformancePattern(performanceType string, feedback []*UserFeedback) *SecurityPattern {
	if len(feedback) == 0 {
		return nil
	}

	// Calculate confidence based on frequency and consistency
	confidence := spd.calculatePatternConfidence(feedback)

	// Determine severity based on performance impact
	severity := spd.determinePerformanceSeverity(performanceType, len(feedback))

	// Get time range
	firstSeen, lastSeen := spd.getTimeRange(feedback)

	// Generate description
	description := spd.generatePerformancePatternDescription(performanceType, len(feedback))

	// Get affected components
	affectedComponents := spd.getAffectedComponents(feedback)

	// Generate recommendations
	recommendations := spd.generatePerformanceRecommendations(performanceType)

	return &SecurityPattern{
		PatternID:          fmt.Sprintf("performance_pattern_%s_%d", performanceType, time.Now().Unix()),
		PatternType:        "performance",
		Description:        description,
		Confidence:         confidence,
		Occurrences:        len(feedback),
		FirstSeen:          firstSeen,
		LastSeen:           lastSeen,
		AffectedComponents: affectedComponents,
		Severity:           severity,
		PatternData: map[string]interface{}{
			"performance_type":    performanceType,
			"frequency":           len(feedback),
			"avg_confidence":      spd.calculateAverageConfidence(feedback),
			"avg_processing_time": spd.calculateAverageProcessingTime(feedback),
		},
		Recommendations: recommendations,
	}
}

// calculatePatternConfidence calculates the confidence of a pattern
func (spd *SecurityPatternDetector) calculatePatternConfidence(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	// Base confidence on frequency
	frequencyConfidence := math.Min(1.0, float64(len(feedback))/10.0)

	// Adjust based on consistency of feedback
	consistencyConfidence := spd.calculateConsistencyConfidence(feedback)

	// Combine confidences
	return (frequencyConfidence + consistencyConfidence) / 2.0
}

// calculateConsistencyConfidence calculates consistency confidence
func (spd *SecurityPatternDetector) calculateConsistencyConfidence(feedback []*UserFeedback) float64 {
	if len(feedback) < 2 {
		return 1.0
	}

	// Calculate variance in confidence scores
	var totalConfidence float64
	for _, fb := range feedback {
		totalConfidence += fb.ConfidenceScore
	}
	meanConfidence := totalConfidence / float64(len(feedback))

	var variance float64
	for _, fb := range feedback {
		diff := fb.ConfidenceScore - meanConfidence
		variance += diff * diff
	}
	variance /= float64(len(feedback))

	// Lower variance = higher consistency
	consistency := math.Max(0.0, 1.0-variance)
	return consistency
}

// calculateAverageConfidence calculates average confidence score
func (spd *SecurityPatternDetector) calculateAverageConfidence(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	var total float64
	for _, fb := range feedback {
		total += fb.ConfidenceScore
	}
	return total / float64(len(feedback))
}

// calculateAverageProcessingTime calculates average processing time
func (spd *SecurityPatternDetector) calculateAverageProcessingTime(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	var total int
	for _, fb := range feedback {
		total += fb.ProcessingTimeMs
	}
	return float64(total) / float64(len(feedback))
}

// getTimeRange gets the time range of feedback
func (spd *SecurityPatternDetector) getTimeRange(feedback []*UserFeedback) (time.Time, time.Time) {
	if len(feedback) == 0 {
		return time.Time{}, time.Time{}
	}

	firstSeen := feedback[0].CreatedAt
	lastSeen := feedback[0].CreatedAt

	for _, fb := range feedback {
		if fb.CreatedAt.Before(firstSeen) {
			firstSeen = fb.CreatedAt
		}
		if fb.CreatedAt.After(lastSeen) {
			lastSeen = fb.CreatedAt
		}
	}

	return firstSeen, lastSeen
}

// getAffectedComponents gets affected components from feedback
func (spd *SecurityPatternDetector) getAffectedComponents(feedback []*UserFeedback) []string {
	components := make(map[string]bool)

	for _, fb := range feedback {
		// Extract components from feedback text
		text := strings.ToLower(fb.FeedbackText)
		if strings.Contains(text, "input") || strings.Contains(text, "validation") {
			components["input_validation"] = true
		}
		if strings.Contains(text, "database") || strings.Contains(text, "storage") {
			components["database"] = true
		}
		if strings.Contains(text, "api") || strings.Contains(text, "service") {
			components["api_service"] = true
		}
		if strings.Contains(text, "website") || strings.Contains(text, "web") {
			components["website_analysis"] = true
		}
		if strings.Contains(text, "ssl") || strings.Contains(text, "certificate") {
			components["ssl_validation"] = true
		}
	}

	var result []string
	for component := range components {
		result = append(result, component)
	}

	return result
}

// Helper methods for determining severity and generating descriptions/recommendations
func (spd *SecurityPatternDetector) determineViolationSeverity(violationType string, frequency int) string {
	baseSeverity := map[string]string{
		"malicious_input":       "high",
		"injection_attack":      "critical",
		"xss_attack":            "critical",
		"phishing_attempt":      "high",
		"spoofing_attempt":      "high",
		"data_exposure":         "critical",
		"unauthorized_access":   "critical",
		"untrusted_data_source": "medium",
	}

	severity := baseSeverity[violationType]
	if severity == "" {
		severity = "medium"
	}

	// Escalate based on frequency
	if frequency > 20 && severity == "medium" {
		severity = "high"
	} else if frequency > 50 && severity == "high" {
		severity = "critical"
	}

	return severity
}

func (spd *SecurityPatternDetector) determineTrustSeverity(sourceType string, frequency int) string {
	// Trust issues are generally medium severity unless frequent
	if frequency > 10 {
		return "high"
	}
	return "medium"
}

func (spd *SecurityPatternDetector) determineVerificationSeverity(verificationType string, frequency int) string {
	// Verification issues are generally medium severity unless frequent
	if frequency > 15 {
		return "high"
	}
	return "medium"
}

func (spd *SecurityPatternDetector) determinePerformanceSeverity(performanceType string, frequency int) string {
	baseSeverity := map[string]string{
		"service_unavailable": "high",
		"resource_exhaustion": "high",
		"slow_response":       "medium",
		"error_rate_high":     "medium",
		"slow_processing":     "low",
	}

	severity := baseSeverity[performanceType]
	if severity == "" {
		severity = "low"
	}

	// Escalate based on frequency
	if frequency > 20 && severity == "low" {
		severity = "medium"
	} else if frequency > 50 && severity == "medium" {
		severity = "high"
	}

	return severity
}

// Description generation methods
func (spd *SecurityPatternDetector) generateViolationPatternDescription(violationType string, frequency int) string {
	descriptions := map[string]string{
		"malicious_input":       "Malicious input detected in classification requests",
		"injection_attack":      "Potential injection attack attempts detected",
		"xss_attack":            "Cross-site scripting attack attempts detected",
		"phishing_attempt":      "Phishing attempts detected in business data",
		"spoofing_attempt":      "Business identity spoofing attempts detected",
		"data_exposure":         "Potential data exposure or leak detected",
		"unauthorized_access":   "Unauthorized access attempts detected",
		"untrusted_data_source": "Untrusted data source usage detected",
	}

	description := descriptions[violationType]
	if description == "" {
		description = "Security violation pattern detected"
	}

	if frequency > 1 {
		description += fmt.Sprintf(" (%d occurrences)", frequency)
	}

	return description
}

func (spd *SecurityPatternDetector) generateTrustPatternDescription(sourceType string, frequency int) string {
	description := fmt.Sprintf("Trust issues with %s data source", sourceType)
	if frequency > 1 {
		description += fmt.Sprintf(" (%d occurrences)", frequency)
	}
	return description
}

func (spd *SecurityPatternDetector) generateVerificationPatternDescription(verificationType string, frequency int) string {
	description := fmt.Sprintf("Website verification issues with %s", verificationType)
	if frequency > 1 {
		description += fmt.Sprintf(" (%d occurrences)", frequency)
	}
	return description
}

func (spd *SecurityPatternDetector) generatePerformancePatternDescription(performanceType string, frequency int) string {
	description := fmt.Sprintf("Performance issues: %s", performanceType)
	if frequency > 1 {
		description += fmt.Sprintf(" (%d occurrences)", frequency)
	}
	return description
}

// Recommendation generation methods
func (spd *SecurityPatternDetector) generateViolationRecommendations(violationType string) []string {
	recommendations := map[string][]string{
		"malicious_input":       {"Implement input sanitization", "Add threat detection rules", "Enhance validation logic"},
		"injection_attack":      {"Implement SQL injection prevention", "Add parameterized queries", "Enhance input validation"},
		"xss_attack":            {"Implement XSS prevention", "Add content security policy", "Sanitize user input"},
		"phishing_attempt":      {"Implement phishing detection", "Add domain reputation checks", "Enhance content analysis"},
		"spoofing_attempt":      {"Implement identity verification", "Add multi-factor authentication", "Enhance business validation"},
		"data_exposure":         {"Implement data encryption", "Add access controls", "Enhance monitoring"},
		"unauthorized_access":   {"Implement access controls", "Add authentication checks", "Enhance authorization logic"},
		"untrusted_data_source": {"Implement source validation", "Add trust scoring", "Enhance source verification"},
	}

	if recs, exists := recommendations[violationType]; exists {
		return recs
	}
	return []string{"Review security measures", "Implement additional validation", "Enhance monitoring"}
}

func (spd *SecurityPatternDetector) generateTrustRecommendations(sourceType string) []string {
	return []string{
		"Review data source reliability",
		"Implement fallback mechanisms",
		"Add source performance monitoring",
		"Update source selection criteria",
	}
}

func (spd *SecurityPatternDetector) generateVerificationRecommendations(verificationType string) []string {
	return []string{
		"Review verification logic",
		"Update verification parameters",
		"Implement additional verification checks",
		"Add verification performance monitoring",
	}
}

func (spd *SecurityPatternDetector) generatePerformanceRecommendations(performanceType string) []string {
	recommendations := map[string][]string{
		"service_unavailable": {"Implement service health checks", "Add fallback services", "Enhance error handling"},
		"resource_exhaustion": {"Optimize resource usage", "Implement resource limits", "Add monitoring"},
		"slow_response":       {"Optimize processing logic", "Add caching", "Implement async processing"},
		"error_rate_high":     {"Review error handling", "Add retry mechanisms", "Enhance logging"},
		"slow_processing":     {"Optimize algorithms", "Add performance monitoring", "Implement parallel processing"},
	}

	if recs, exists := recommendations[performanceType]; exists {
		return recs
	}
	return []string{"Review performance metrics", "Implement optimization", "Add monitoring"}
}

// filterPatterns filters patterns based on confidence and occurrence thresholds
func (spd *SecurityPatternDetector) filterPatterns(patterns []*SecurityPattern) []*SecurityPattern {
	var filtered []*SecurityPattern

	for _, pattern := range patterns {
		if pattern.Confidence >= spd.config.PatternMinConfidence &&
			pattern.Occurrences >= spd.config.PatternMinOccurrences {
			filtered = append(filtered, pattern)
		}
	}

	return filtered
}
