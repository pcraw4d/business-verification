package feedback

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SecurityFeedbackAnalyzer analyzes security-related feedback
type SecurityFeedbackAnalyzer struct {
	config *MLAnalysisConfig
	logger *zap.Logger
}

// NewSecurityFeedbackAnalyzer creates a new security feedback analyzer
func NewSecurityFeedbackAnalyzer(config *MLAnalysisConfig, logger *zap.Logger) *SecurityFeedbackAnalyzer {
	return &SecurityFeedbackAnalyzer{
		config: config,
		logger: logger,
	}
}

// AnalyzeSecurityFeedback analyzes security-related feedback
func (sfa *SecurityFeedbackAnalyzer) AnalyzeSecurityFeedback(ctx context.Context, feedback []*UserFeedback) (*SecurityFeedbackAnalysis, error) {
	if len(feedback) < sfa.config.MinFeedbackThreshold {
		return &SecurityFeedbackAnalysis{
			OverallSecurityScore: 1.0, // Assume perfect security with no feedback
		}, nil
	}

	analysis := &SecurityFeedbackAnalysis{
		SecurityViolations:        []*SecurityViolation{},
		TrustedSourceIssues:       []*TrustedSourceIssue{},
		WebsiteVerificationIssues: []*WebsiteVerificationIssue{},
		Recommendations:           []*SecurityRecommendation{},
	}

	// Analyze security violations
	violations := sfa.analyzeSecurityViolations(feedback)
	analysis.SecurityViolations = violations

	// Analyze trusted source issues
	trustedSourceIssues := sfa.analyzeTrustedSourceIssues(feedback)
	analysis.TrustedSourceIssues = trustedSourceIssues

	// Analyze website verification issues
	websiteIssues := sfa.analyzeWebsiteVerificationIssues(feedback)
	analysis.WebsiteVerificationIssues = websiteIssues

	// Calculate overall security score
	analysis.OverallSecurityScore = sfa.calculateOverallSecurityScore(analysis)

	// Generate security recommendations
	recommendations := sfa.generateSecurityRecommendations(analysis)
	analysis.Recommendations = recommendations

	return analysis, nil
}

// AnalyzeTrustedSourceFeedback analyzes trusted data source feedback
func (sfa *SecurityFeedbackAnalyzer) AnalyzeTrustedSourceFeedback(ctx context.Context, feedback []*UserFeedback) (*TrustedSourceAnalysis, error) {
	if len(feedback) < sfa.config.MinFeedbackThreshold {
		return &TrustedSourceAnalysis{
			OverallTrustScore: 1.0, // Assume perfect trust with no feedback
		}, nil
	}

	analysis := &TrustedSourceAnalysis{
		SourceReliability: make(map[string]float64),
		SourceAccuracy:    make(map[string]float64),
		SourcePerformance: make(map[string]float64),
		Recommendations:   []string{},
	}

	// Group feedback by source type
	sourceFeedback := make(map[string][]*UserFeedback)
	for _, fb := range feedback {
		sourceType := sfa.determineSourceType(fb)
		sourceFeedback[sourceType] = append(sourceFeedback[sourceType], fb)
	}

	// Analyze each source type
	for sourceType, sourceFb := range sourceFeedback {
		if len(sourceFb) >= sfa.config.MinFeedbackThreshold {
			reliability := sfa.calculateSourceReliability(sourceFb)
			accuracy := sfa.calculateSourceAccuracy(sourceFb)
			performance := sfa.calculateSourcePerformance(sourceFb)

			analysis.SourceReliability[sourceType] = reliability
			analysis.SourceAccuracy[sourceType] = accuracy
			analysis.SourcePerformance[sourceType] = performance
		}
	}

	// Calculate overall trust score
	analysis.OverallTrustScore = sfa.calculateOverallTrustScore(analysis)

	// Generate recommendations
	recommendations := sfa.generateTrustedSourceRecommendations(analysis)
	analysis.Recommendations = recommendations

	return analysis, nil
}

// analyzeSecurityViolations analyzes security violations in feedback
func (sfa *SecurityFeedbackAnalyzer) analyzeSecurityViolations(feedback []*UserFeedback) []*SecurityViolation {
	var violations []*SecurityViolation

	// Group violations by type
	// Stub: UserFeedback doesn't have FeedbackType field - needs refactoring
	violationGroups := make(map[string][]*UserFeedback)
	for _, fb := range feedback {
		// TODO: Refactor to use ClassificationClassificationUserFeedback which has FeedbackType
		// For now, skip security validation filtering
		violationType := sfa.determineViolationType(fb)
		if violationType != "" {
			violationGroups[violationType] = append(violationGroups[violationType], fb)
		}
	}

	// Create violation records
	for violationType, violationFb := range violationGroups {
		if len(violationFb) >= sfa.config.MinFeedbackThreshold/2 { // Lower threshold for violations
			violation := sfa.createSecurityViolation(violationType, violationFb)
			if violation != nil {
				violations = append(violations, violation)
			}
		}
	}

	// Sort by severity and frequency
	sort.Slice(violations, func(i, j int) bool {
		severityOrder := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}
		if severityOrder[violations[i].Severity] == severityOrder[violations[j].Severity] {
			return violations[i].DetectionTime.After(violations[j].DetectionTime)
		}
		return severityOrder[violations[i].Severity] > severityOrder[violations[j].Severity]
	})

	return violations
}

// determineViolationType determines the type of security violation
// Stub: UserFeedback doesn't have FeedbackText or FeedbackValue fields - needs refactoring
func (sfa *SecurityFeedbackAnalyzer) determineViolationType(feedback *UserFeedback) string {
	// Use Comments as FeedbackText substitute
	text := strings.ToLower(feedback.Comments)

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

	// TODO: Refactor to use ClassificationClassificationUserFeedback which has FeedbackValue
	// Stub - skip FeedbackValue check

	return ""
}

// createSecurityViolation creates a security violation record
func (sfa *SecurityFeedbackAnalyzer) createSecurityViolation(violationType string, feedback []*UserFeedback) *SecurityViolation {
	if len(feedback) == 0 {
		return nil
	}

	// Determine severity based on violation type and frequency
	severity := sfa.determineViolationSeverity(violationType, len(feedback))

	// Get affected data
	affectedData := sfa.extractAffectedData(feedback)

	// Get detection time (most recent feedback)
	var latestTime time.Time
	for _, fb := range feedback {
		if fb.SubmittedAt.After(latestTime) {
			latestTime = fb.SubmittedAt
		}
	}

	// Determine resolution status
	resolutionStatus := sfa.determineResolutionStatus(feedback)

	return &SecurityViolation{
		ViolationID:      fmt.Sprintf("violation_%s_%d", violationType, latestTime.Unix()),
		ViolationType:    violationType,
		Severity:         severity,
		Description:      sfa.generateViolationDescription(violationType, len(feedback)),
		AffectedData:     affectedData,
		DetectionTime:    latestTime,
		ResolutionStatus: resolutionStatus,
	}
}

// determineViolationSeverity determines the severity of a violation
func (sfa *SecurityFeedbackAnalyzer) determineViolationSeverity(violationType string, frequency int) string {
	// Base severity on violation type
	baseSeverity := map[string]string{
		"malicious_input":            "high",
		"injection_attack":           "critical",
		"xss_attack":                 "critical",
		"phishing_attempt":           "high",
		"spoofing_attempt":           "high",
		"data_exposure":              "critical",
		"unauthorized_access":        "critical",
		"untrusted_data_source":      "medium",
		"general_security_violation": "medium",
	}

	severity := baseSeverity[violationType]
	if severity == "" {
		severity = "low"
	}

	// Escalate severity based on frequency
	if frequency > 20 {
		if severity == "low" {
			severity = "medium"
		} else if severity == "medium" {
			severity = "high"
		}
	} else if frequency > 50 {
		if severity == "medium" {
			severity = "high"
		} else if severity == "high" {
			severity = "critical"
		}
	}

	return severity
}

// extractAffectedData extracts affected data from feedback
func (sfa *SecurityFeedbackAnalyzer) extractAffectedData(feedback []*UserFeedback) []string {
	var affectedData []string
	dataTypes := make(map[string]bool)

	for _, fb := range feedback {
		// Extract data types from feedback
		// Stub: UserFeedback doesn't have BusinessName or FeedbackText fields
		// TODO: Refactor to use ClassificationClassificationUserFeedback
		// For now, use Comments as FeedbackText substitute
		if strings.Contains(strings.ToLower(fb.Comments), "business") || strings.Contains(strings.ToLower(fb.Comments), "name") {
			dataTypes["business_name"] = true
		}
		if strings.Contains(strings.ToLower(fb.Comments), "email") {
			dataTypes["email"] = true
		}
		if strings.Contains(strings.ToLower(fb.Comments), "website") {
			dataTypes["website"] = true
		}
		if strings.Contains(strings.ToLower(fb.Comments), "phone") {
			dataTypes["phone"] = true
		}
		if strings.Contains(strings.ToLower(fb.Comments), "address") {
			dataTypes["address"] = true
		}
		if strings.Contains(strings.ToLower(fb.Comments), "description") {
			dataTypes["description"] = true
		}
	}

	for dataType := range dataTypes {
		affectedData = append(affectedData, dataType)
	}

	return affectedData
}

// determineResolutionStatus determines the resolution status of a violation
func (sfa *SecurityFeedbackAnalyzer) determineResolutionStatus(feedback []*UserFeedback) string {
	// Check if any feedback indicates resolution
	for _, fb := range feedback {
		text := strings.ToLower(fb.Comments)
		if strings.Contains(text, "resolved") || strings.Contains(text, "fixed") {
			return "resolved"
		} else if strings.Contains(text, "investigating") || strings.Contains(text, "pending") {
			return "investigating"
		}
	}

	// Check feedback age
	now := time.Now()
	for _, fb := range feedback {
		if now.Sub(fb.SubmittedAt) > 24*time.Hour {
			return "investigating"
		}
	}

	return "pending"
}

// generateViolationDescription generates a description for a security violation
func (sfa *SecurityFeedbackAnalyzer) generateViolationDescription(violationType string, frequency int) string {
	descriptions := map[string]string{
		"malicious_input":            "Malicious input detected in classification requests",
		"injection_attack":           "Potential injection attack attempt detected",
		"xss_attack":                 "Cross-site scripting attack attempt detected",
		"phishing_attempt":           "Phishing attempt detected in business data",
		"spoofing_attempt":           "Business identity spoofing attempt detected",
		"data_exposure":              "Potential data exposure or leak detected",
		"unauthorized_access":        "Unauthorized access attempt detected",
		"untrusted_data_source":      "Untrusted data source used in classification",
		"general_security_violation": "General security violation detected",
	}

	description := descriptions[violationType]
	if description == "" {
		description = "Security violation detected"
	}

	if frequency > 1 {
		description += fmt.Sprintf(" (%d occurrences)", frequency)
	}

	return description
}

// analyzeTrustedSourceIssues analyzes issues with trusted data sources
func (sfa *SecurityFeedbackAnalyzer) analyzeTrustedSourceIssues(feedback []*UserFeedback) []*TrustedSourceIssue {
	var issues []*TrustedSourceIssue

	// Group issues by source type
	// Stub: UserFeedback doesn't have FeedbackType field - needs refactoring
	issueGroups := make(map[string][]*UserFeedback)
	for _, fb := range feedback {
		// TODO: Refactor to use ClassificationClassificationUserFeedback which has FeedbackType
		// For now, include all feedback
		sourceType := sfa.determineSourceType(fb)
		issueType := sfa.determineSourceIssueType(fb)
		key := fmt.Sprintf("%s_%s", sourceType, issueType)
		issueGroups[key] = append(issueGroups[key], fb)
	}

	// Create issue records
	for key, issueFb := range issueGroups {
		if len(issueFb) >= sfa.config.MinFeedbackThreshold/3 { // Lower threshold for source issues
			parts := strings.Split(key, "_")
			if len(parts) >= 2 {
				sourceType := parts[0]
				issueType := strings.Join(parts[1:], "_")
				issue := sfa.createTrustedSourceIssue(sourceType, issueType, issueFb)
				if issue != nil {
					issues = append(issues, issue)
				}
			}
		}
	}

	return issues
}

// determineSourceType determines the type of data source
func (sfa *SecurityFeedbackAnalyzer) determineSourceType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.Comments)

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

	// Default based on feedback type
	// Stub: UserFeedback doesn't have FeedbackType field - needs refactoring
	// TODO: Refactor to use ClassificationClassificationUserFeedback which has FeedbackType
	return "unknown_source" // Default stub value
}

// determineSourceIssueType determines the type of source issue
func (sfa *SecurityFeedbackAnalyzer) determineSourceIssueType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.Comments)

	// Check for specific issue types
	if strings.Contains(text, "unavailable") || strings.Contains(text, "down") {
		return "source_unavailable"
	} else if strings.Contains(text, "slow") || strings.Contains(text, "timeout") {
		return "performance_issue"
	} else if strings.Contains(text, "invalid") || strings.Contains(text, "error") {
		return "invalid_response"
	} else if strings.Contains(text, "outdated") || strings.Contains(text, "stale") {
		return "outdated_data"
	} else if strings.Contains(text, "incomplete") || strings.Contains(text, "missing") {
		return "incomplete_data"
	} else if strings.Contains(text, "untrusted") || strings.Contains(text, "suspicious") {
		return "trust_issue"
	}

	return "general_issue"
}

// createTrustedSourceIssue creates a trusted source issue record
func (sfa *SecurityFeedbackAnalyzer) createTrustedSourceIssue(sourceType, issueType string, feedback []*UserFeedback) *TrustedSourceIssue {
	if len(feedback) == 0 {
		return nil
	}

	// Calculate affected requests
	affectedRequests := len(feedback)

	// Determine resolution status
	resolutionStatus := sfa.determineResolutionStatus(feedback)

	// Generate description
	description := sfa.generateSourceIssueDescription(sourceType, issueType, affectedRequests)

	return &TrustedSourceIssue{
		IssueID:          fmt.Sprintf("source_issue_%s_%s_%d", sourceType, issueType, time.Now().Unix()),
		SourceType:       sourceType,
		IssueType:        issueType,
		Description:      description,
		AffectedRequests: affectedRequests,
		ResolutionStatus: resolutionStatus,
	}
}

// generateSourceIssueDescription generates a description for a source issue
func (sfa *SecurityFeedbackAnalyzer) generateSourceIssueDescription(sourceType, issueType string, affectedRequests int) string {
	descriptions := map[string]string{
		"source_unavailable": "Data source is currently unavailable",
		"performance_issue":  "Data source experiencing performance issues",
		"invalid_response":   "Data source returning invalid responses",
		"outdated_data":      "Data source providing outdated information",
		"incomplete_data":    "Data source providing incomplete information",
		"trust_issue":        "Trust issues with data source",
		"general_issue":      "General issue with data source",
	}

	description := descriptions[issueType]
	if description == "" {
		description = "Issue with data source"
	}

	description += fmt.Sprintf(" (%s, %d affected requests)", sourceType, affectedRequests)

	return description
}

// analyzeWebsiteVerificationIssues analyzes website verification issues
func (sfa *SecurityFeedbackAnalyzer) analyzeWebsiteVerificationIssues(feedback []*UserFeedback) []*WebsiteVerificationIssue {
	var issues []*WebsiteVerificationIssue

	// Group issues by verification type
	// Stub: UserFeedback doesn't have FeedbackType field - needs refactoring
	issueGroups := make(map[string][]*UserFeedback)
	for _, fb := range feedback {
		// TODO: Refactor to use ClassificationClassificationUserFeedback which has FeedbackType
		// For now, include all feedback
		verificationType := sfa.determineVerificationType(fb)
		issueType := sfa.determineWebsiteIssueType(fb)
		key := fmt.Sprintf("%s_%s", verificationType, issueType)
		issueGroups[key] = append(issueGroups[key], fb)
	}

	// Create issue records
	for key, issueFb := range issueGroups {
		if len(issueFb) >= sfa.config.MinFeedbackThreshold/4 { // Lower threshold for website issues
			parts := strings.Split(key, "_")
			if len(parts) >= 2 {
				verificationType := parts[0]
				issueType := strings.Join(parts[1:], "_")
				issue := sfa.createWebsiteVerificationIssue(verificationType, issueType, issueFb)
				if issue != nil {
					issues = append(issues, issue)
				}
			}
		}
	}

	return issues
}

// determineVerificationType determines the type of website verification
func (sfa *SecurityFeedbackAnalyzer) determineVerificationType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.Comments)

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

// determineWebsiteIssueType determines the type of website issue
func (sfa *SecurityFeedbackAnalyzer) determineWebsiteIssueType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.Comments)

	// Check for specific issue types
	if strings.Contains(text, "invalid") || strings.Contains(text, "error") {
		return "verification_failure"
	} else if strings.Contains(text, "expired") || strings.Contains(text, "invalid") {
		return "certificate_issue"
	} else if strings.Contains(text, "mismatch") || strings.Contains(text, "different") {
		return "domain_mismatch"
	} else if strings.Contains(text, "unreachable") || strings.Contains(text, "down") {
		return "website_unreachable"
	} else if strings.Contains(text, "suspicious") || strings.Contains(text, "malicious") {
		return "security_concern"
	}

	return "general_issue"
}

// createWebsiteVerificationIssue creates a website verification issue record
func (sfa *SecurityFeedbackAnalyzer) createWebsiteVerificationIssue(verificationType, issueType string, feedback []*UserFeedback) *WebsiteVerificationIssue {
	if len(feedback) == 0 {
		return nil
	}

	// Extract affected websites
	affectedWebsites := sfa.extractAffectedWebsites(feedback)

	// Determine resolution status
	resolutionStatus := sfa.determineResolutionStatus(feedback)

	// Generate description
	description := sfa.generateWebsiteIssueDescription(verificationType, issueType, len(affectedWebsites))

	return &WebsiteVerificationIssue{
		IssueID:          fmt.Sprintf("website_issue_%s_%s_%d", verificationType, issueType, time.Now().Unix()),
		VerificationType: verificationType,
		IssueType:        issueType,
		Description:      description,
		AffectedWebsites: affectedWebsites,
		ResolutionStatus: resolutionStatus,
	}
}

// extractAffectedWebsites extracts affected websites from feedback
func (sfa *SecurityFeedbackAnalyzer) extractAffectedWebsites(feedback []*UserFeedback) []string {
	var websites []string
	websiteSet := make(map[string]bool)

	for _, fb := range feedback {
		// Extract website from feedback text
		text := strings.ToLower(fb.Comments)
		// Simple extraction - look for common website patterns
		words := strings.Fields(text)
		for _, word := range words {
			if strings.Contains(word, ".com") || strings.Contains(word, ".org") || strings.Contains(word, ".net") {
				website := strings.Trim(word, ".,!?")
				if !websiteSet[website] {
					websiteSet[website] = true
					websites = append(websites, website)
				}
			}
		}
	}

	return websites
}

// generateWebsiteIssueDescription generates a description for a website issue
func (sfa *SecurityFeedbackAnalyzer) generateWebsiteIssueDescription(verificationType, issueType string, affectedCount int) string {
	descriptions := map[string]string{
		"verification_failure": "Website verification failed",
		"certificate_issue":    "SSL certificate issue detected",
		"domain_mismatch":      "Domain name mismatch detected",
		"website_unreachable":  "Website is unreachable",
		"security_concern":     "Security concern with website",
		"general_issue":        "General website verification issue",
	}

	description := descriptions[issueType]
	if description == "" {
		description = "Website verification issue"
	}

	description += fmt.Sprintf(" (%s, %d affected websites)", verificationType, affectedCount)

	return description
}

// calculateOverallSecurityScore calculates the overall security score
func (sfa *SecurityFeedbackAnalyzer) calculateOverallSecurityScore(analysis *SecurityFeedbackAnalysis) float64 {
	baseScore := 1.0

	// Penalize security violations
	violationPenalty := 0.0
	for _, violation := range analysis.SecurityViolations {
		severityPenalty := map[string]float64{
			"critical": 0.2,
			"high":     0.1,
			"medium":   0.05,
			"low":      0.02,
		}
		violationPenalty += severityPenalty[violation.Severity]
	}

	// Penalize trusted source issues
	sourcePenalty := float64(len(analysis.TrustedSourceIssues)) * 0.05

	// Penalize website verification issues
	websitePenalty := float64(len(analysis.WebsiteVerificationIssues)) * 0.03

	// Calculate final score
	finalScore := baseScore - violationPenalty - sourcePenalty - websitePenalty

	// Ensure score is between 0 and 1
	if finalScore < 0 {
		finalScore = 0
	} else if finalScore > 1 {
		finalScore = 1
	}

	return finalScore
}

// calculateSourceReliability calculates reliability for a data source
func (sfa *SecurityFeedbackAnalyzer) calculateSourceReliability(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 1.0
	}

	var reliabilityScore float64
	var totalWeight float64

	for _, fb := range feedback {
		// Base reliability on feedback type and content
		weight := 1.0
		score := 0.8 // Default score

		// Stub: UserFeedback doesn't have FeedbackType field - needs refactoring
		// TODO: Refactor to use ClassificationClassificationUserFeedback which has FeedbackType
		// For now, skip type-specific logic
		if false { // fb.FeedbackType == FeedbackTypeDataSourceTrust {
			// Check for reliability indicators in feedback
			text := strings.ToLower(fb.Comments)
			if strings.Contains(text, "reliable") || strings.Contains(text, "trusted") {
				score = 1.0
			} else if strings.Contains(text, "unreliable") || strings.Contains(text, "untrusted") {
				score = 0.0
			} else if strings.Contains(text, "slow") || strings.Contains(text, "timeout") {
				score = 0.6
			} else if strings.Contains(text, "error") || strings.Contains(text, "invalid") {
				score = 0.3
			}
		}

		reliabilityScore += score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.8 // Default reliability
	}

	return reliabilityScore / totalWeight
}

// calculateSourceAccuracy calculates accuracy for a data source
func (sfa *SecurityFeedbackAnalyzer) calculateSourceAccuracy(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 1.0
	}

	var accuracyScore float64
	var totalWeight float64

	for _, fb := range feedback {
		weight := 1.0
		score := 0.8 // Default score

		// Stub: UserFeedback doesn't have FeedbackType or FeedbackValue fields - needs refactoring
		// TODO: Refactor to use ClassificationClassificationUserFeedback which has these fields
		// For now, skip type-specific logic
		if false { // fb.FeedbackType == FeedbackTypeAccuracy {
			if false { // accuracy, ok := fb.FeedbackValue["accuracy"].(string); ok && accuracy == "correct" {
				score = 1.0
			} else {
				score = 0.0
			}
		} else {
			// Infer accuracy from feedback text
			text := strings.ToLower(fb.Comments)
			if strings.Contains(text, "accurate") || strings.Contains(text, "correct") {
				score = 1.0
			} else if strings.Contains(text, "inaccurate") || strings.Contains(text, "wrong") {
				score = 0.0
			} else if strings.Contains(text, "outdated") || strings.Contains(text, "stale") {
				score = 0.5
			}
		}

		accuracyScore += score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.8 // Default accuracy
	}

	return accuracyScore / totalWeight
}

// calculateSourcePerformance calculates performance for a data source
func (sfa *SecurityFeedbackAnalyzer) calculateSourcePerformance(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 1.0
	}

	var performanceScore float64
	var totalWeight float64

	for _, fb := range feedback {
		weight := 1.0
		score := 0.8 // Default score

		// Check processing time
		// Stub: UserFeedback doesn't have ProcessingTimeMs field
		// TODO: Refactor to use ClassificationClassificationUserFeedback
		// For now, skip processing time checks
		processingTimeMs := 0 // Stub value
		if processingTimeMs > 0 {
			if processingTimeMs < 500 {
				score = 1.0 // Fast
			} else if processingTimeMs < 2000 {
				score = 0.7 // Medium
			} else {
				score = 0.3 // Slow
			}
		}

		// Adjust based on feedback text
		text := strings.ToLower(fb.Comments)
		if strings.Contains(text, "fast") || strings.Contains(text, "quick") {
			score = 1.0
		} else if strings.Contains(text, "slow") || strings.Contains(text, "timeout") {
			score = 0.2
		} else if strings.Contains(text, "unavailable") || strings.Contains(text, "down") {
			score = 0.0
		}

		performanceScore += score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.8 // Default performance
	}

	return performanceScore / totalWeight
}

// calculateOverallTrustScore calculates the overall trust score for data sources
func (sfa *SecurityFeedbackAnalyzer) calculateOverallTrustScore(analysis *TrustedSourceAnalysis) float64 {
	if len(analysis.SourceReliability) == 0 {
		return 1.0
	}

	var totalScore float64
	var totalWeight float64

	// Weight different metrics
	reliabilityWeight := 0.4
	accuracyWeight := 0.4
	performanceWeight := 0.2

	for sourceType := range analysis.SourceReliability {
		reliability := analysis.SourceReliability[sourceType]
		accuracy := analysis.SourceAccuracy[sourceType]
		performance := analysis.SourcePerformance[sourceType]

		// Calculate weighted score for this source
		sourceScore := reliability*reliabilityWeight + accuracy*accuracyWeight + performance*performanceWeight

		// Weight by number of feedback items (more feedback = higher weight)
		weight := 1.0 // Could be based on feedback count if available

		totalScore += sourceScore * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.8 // Default trust score
	}

	return totalScore / totalWeight
}

// generateSecurityRecommendations generates security recommendations
func (sfa *SecurityFeedbackAnalyzer) generateSecurityRecommendations(analysis *SecurityFeedbackAnalysis) []*SecurityRecommendation {
	var recommendations []*SecurityRecommendation

	// Generate recommendations based on security violations
	for _, violation := range analysis.SecurityViolations {
		rec := sfa.generateViolationRecommendation(violation)
		if rec != nil {
			recommendations = append(recommendations, rec)
		}
	}

	// Generate recommendations based on trusted source issues
	for _, issue := range analysis.TrustedSourceIssues {
		rec := sfa.generateSourceIssueRecommendation(issue)
		if rec != nil {
			recommendations = append(recommendations, rec)
		}
	}

	// Generate recommendations based on website verification issues
	for _, issue := range analysis.WebsiteVerificationIssues {
		rec := sfa.generateWebsiteIssueRecommendation(issue)
		if rec != nil {
			recommendations = append(recommendations, rec)
		}
	}

	// Sort by priority
	sort.Slice(recommendations, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
	})

	return recommendations
}

// generateViolationRecommendation generates a recommendation for a security violation
func (sfa *SecurityFeedbackAnalyzer) generateViolationRecommendation(violation *SecurityViolation) *SecurityRecommendation {
	priority := "medium"
	if violation.Severity == "critical" {
		priority = "high"
	} else if violation.Severity == "low" {
		priority = "low"
	}

	description := fmt.Sprintf("Address %s security violation", violation.ViolationType)

	implementationSteps := []string{
		"Investigate the root cause of the violation",
		"Implement additional validation checks",
		"Update security monitoring rules",
		"Test the fix with similar inputs",
	}

	validationCriteria := []string{
		"Violation no longer occurs with similar inputs",
		"Security monitoring detects and blocks similar attempts",
		"System performance is not significantly impacted",
	}

	return &SecurityRecommendation{
		RecommendationID:    fmt.Sprintf("security_violation_%s", violation.ViolationID),
		SecurityType:        "violation_prevention",
		Priority:            priority,
		Description:         description,
		AffectedComponents:  []string{"input_validation", "security_monitoring", "classification_engine"},
		ImplementationSteps: implementationSteps,
		ValidationCriteria:  validationCriteria,
	}
}

// generateSourceIssueRecommendation generates a recommendation for a trusted source issue
func (sfa *SecurityFeedbackAnalyzer) generateSourceIssueRecommendation(issue *TrustedSourceIssue) *SecurityRecommendation {
	priority := "medium"
	if issue.IssueType == "source_unavailable" || issue.IssueType == "trust_issue" {
		priority = "high"
	}

	description := fmt.Sprintf("Resolve %s issue with %s", issue.IssueType, issue.SourceType)

	implementationSteps := []string{
		"Contact the data source provider",
		"Implement fallback mechanisms",
		"Update data source configuration",
		"Monitor source availability",
	}

	validationCriteria := []string{
		"Data source is available and responding correctly",
		"Fallback mechanisms work when source is unavailable",
		"Data quality meets expected standards",
	}

	return &SecurityRecommendation{
		RecommendationID:    fmt.Sprintf("source_issue_%s", issue.IssueID),
		SecurityType:        "data_source_improvement",
		Priority:            priority,
		Description:         description,
		AffectedComponents:  []string{"data_source_integration", "fallback_mechanisms"},
		ImplementationSteps: implementationSteps,
		ValidationCriteria:  validationCriteria,
	}
}

// generateWebsiteIssueRecommendation generates a recommendation for a website verification issue
func (sfa *SecurityFeedbackAnalyzer) generateWebsiteIssueRecommendation(issue *WebsiteVerificationIssue) *SecurityRecommendation {
	priority := "medium"
	if issue.IssueType == "security_concern" || issue.IssueType == "verification_failure" {
		priority = "high"
	}

	description := fmt.Sprintf("Resolve %s issue with %s", issue.IssueType, issue.VerificationType)

	implementationSteps := []string{
		"Review website verification logic",
		"Update verification parameters",
		"Implement additional verification checks",
		"Test with affected websites",
	}

	validationCriteria := []string{
		"Website verification works correctly",
		"False positives are minimized",
		"Security concerns are properly flagged",
	}

	return &SecurityRecommendation{
		RecommendationID:    fmt.Sprintf("website_issue_%s", issue.IssueID),
		SecurityType:        "website_verification_improvement",
		Priority:            priority,
		Description:         description,
		AffectedComponents:  []string{"website_verification", "ssl_validation", "domain_analysis"},
		ImplementationSteps: implementationSteps,
		ValidationCriteria:  validationCriteria,
	}
}

// generateTrustedSourceRecommendations generates recommendations for trusted source improvements
func (sfa *SecurityFeedbackAnalyzer) generateTrustedSourceRecommendations(analysis *TrustedSourceAnalysis) []string {
	var recommendations []string

	// Check for low-performing sources
	for sourceType, reliability := range analysis.SourceReliability {
		if reliability < 0.7 {
			recommendations = append(recommendations, fmt.Sprintf("Improve reliability of %s data source (current: %.2f)", sourceType, reliability))
		}
	}

	for sourceType, accuracy := range analysis.SourceAccuracy {
		if accuracy < 0.8 {
			recommendations = append(recommendations, fmt.Sprintf("Improve accuracy of %s data source (current: %.2f)", sourceType, accuracy))
		}
	}

	for sourceType, performance := range analysis.SourcePerformance {
		if performance < 0.6 {
			recommendations = append(recommendations, fmt.Sprintf("Improve performance of %s data source (current: %.2f)", sourceType, performance))
		}
	}

	// General recommendations
	if analysis.OverallTrustScore < 0.8 {
		recommendations = append(recommendations, "Overall data source trust score is below threshold - review all sources")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "All data sources are performing well")
	}

	return recommendations
}
