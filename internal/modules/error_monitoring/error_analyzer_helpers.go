package error_monitoring

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// analyzeRootCauses performs root cause analysis for errors
func (ea *ErrorAnalyzer) analyzeRootCauses(errors []ErrorEntry, patterns []*ErrorPattern, processName string) []*RootCauseAnalysis {
	var rootCauses []*RootCauseAnalysis

	// Analyze infrastructure-related root causes
	infraRootCauses := ea.analyzeInfrastructureRootCauses(errors, processName)
	rootCauses = append(rootCauses, infraRootCauses...)

	// Analyze application-related root causes
	appRootCauses := ea.analyzeApplicationRootCauses(errors, processName)
	rootCauses = append(rootCauses, appRootCauses...)

	// Analyze data-related root causes
	dataRootCauses := ea.analyzeDataRootCauses(errors, processName)
	rootCauses = append(rootCauses, dataRootCauses...)

	// Analyze external service root causes
	externalRootCauses := ea.analyzeExternalRootCauses(errors, processName)
	rootCauses = append(rootCauses, externalRootCauses...)

	// Analyze configuration-related root causes
	configRootCauses := ea.analyzeConfigurationRootCauses(errors, processName)
	rootCauses = append(rootCauses, configRootCauses...)

	return rootCauses
}

// analyzeInfrastructureRootCauses analyzes infrastructure-related root causes
func (ea *ErrorAnalyzer) analyzeInfrastructureRootCauses(errors []ErrorEntry, processName string) []*RootCauseAnalysis {
	var rootCauses []*RootCauseAnalysis

	// Check for network connectivity issues
	networkErrors := ea.filterErrorsByType(errors, []string{"network_timeout", "connection_failed", "dns_error"})
	if len(networkErrors) > 0 {
		rootCause := &RootCauseAnalysis{
			ID:                generateRootCauseID(),
			Category:          "infrastructure",
			RootCause:         "Network Connectivity Issues",
			Description:       "Network connectivity problems causing timeouts and connection failures",
			Confidence:        ea.calculateRootCauseConfidence(networkErrors, errors),
			Impact:            ea.determineImpact(networkErrors),
			AffectedProcesses: []string{processName},
			Evidence:          ea.collectEvidence(networkErrors, "network_connectivity"),
			ContributingFactors: []string{
				"Network latency",
				"DNS resolution issues",
				"Firewall restrictions",
				"Load balancer problems",
			},
			Recommendations: []string{
				"Check network connectivity and latency",
				"Verify DNS configuration",
				"Review firewall rules",
				"Monitor load balancer health",
			},
			Timeline: ea.buildTimeline(networkErrors),
		}
		rootCauses = append(rootCauses, rootCause)
	}

	// Check for resource exhaustion
	resourceErrors := ea.filterErrorsByType(errors, []string{"out_of_memory", "disk_full", "cpu_overload"})
	if len(resourceErrors) > 0 {
		rootCause := &RootCauseAnalysis{
			ID:                generateRootCauseID(),
			Category:          "infrastructure",
			RootCause:         "Resource Exhaustion",
			Description:       "System resources (CPU, memory, disk) are being exhausted",
			Confidence:        ea.calculateRootCauseConfidence(resourceErrors, errors),
			Impact:            ea.determineImpact(resourceErrors),
			AffectedProcesses: []string{processName},
			Evidence:          ea.collectEvidence(resourceErrors, "resource_exhaustion"),
			ContributingFactors: []string{
				"Insufficient system resources",
				"Memory leaks",
				"High CPU usage",
				"Disk space issues",
			},
			Recommendations: []string{
				"Monitor resource usage",
				"Scale up infrastructure",
				"Implement resource limits",
				"Optimize application performance",
			},
			Timeline: ea.buildTimeline(resourceErrors),
		}
		rootCauses = append(rootCauses, rootCause)
	}

	return rootCauses
}

// analyzeApplicationRootCauses analyzes application-related root causes
func (ea *ErrorAnalyzer) analyzeApplicationRootCauses(errors []ErrorEntry, processName string) []*RootCauseAnalysis {
	var rootCauses []*RootCauseAnalysis

	// Check for application logic errors
	logicErrors := ea.filterErrorsByType(errors, []string{"validation_error", "business_logic_error", "data_integrity_error"})
	if len(logicErrors) > 0 {
		rootCause := &RootCauseAnalysis{
			ID:                generateRootCauseID(),
			Category:          "application",
			RootCause:         "Application Logic Errors",
			Description:       "Errors in application business logic and validation",
			Confidence:        ea.calculateRootCauseConfidence(logicErrors, errors),
			Impact:            ea.determineImpact(logicErrors),
			AffectedProcesses: []string{processName},
			Evidence:          ea.collectEvidence(logicErrors, "application_logic"),
			ContributingFactors: []string{
				"Invalid input data",
				"Business rule violations",
				"Data validation failures",
				"Logic implementation bugs",
			},
			Recommendations: []string{
				"Review and fix business logic",
				"Improve input validation",
				"Add comprehensive error handling",
				"Implement data integrity checks",
			},
			Timeline: ea.buildTimeline(logicErrors),
		}
		rootCauses = append(rootCauses, rootCause)
	}

	// Check for performance issues
	performanceErrors := ea.filterErrorsByType(errors, []string{"timeout", "slow_response", "deadlock"})
	if len(performanceErrors) > 0 {
		rootCause := &RootCauseAnalysis{
			ID:                generateRootCauseID(),
			Category:          "application",
			RootCause:         "Performance Issues",
			Description:       "Application performance problems causing timeouts and slow responses",
			Confidence:        ea.calculateRootCauseConfidence(performanceErrors, errors),
			Impact:            ea.determineImpact(performanceErrors),
			AffectedProcesses: []string{processName},
			Evidence:          ea.collectEvidence(performanceErrors, "performance"),
			ContributingFactors: []string{
				"Inefficient algorithms",
				"Database query performance",
				"Memory leaks",
				"Concurrency issues",
			},
			Recommendations: []string{
				"Profile application performance",
				"Optimize database queries",
				"Implement caching strategies",
				"Review concurrency patterns",
			},
			Timeline: ea.buildTimeline(performanceErrors),
		}
		rootCauses = append(rootCauses, rootCause)
	}

	return rootCauses
}

// analyzeDataRootCauses analyzes data-related root causes
func (ea *ErrorAnalyzer) analyzeDataRootCauses(errors []ErrorEntry, processName string) []*RootCauseAnalysis {
	var rootCauses []*RootCauseAnalysis

	// Check for data quality issues
	dataQualityErrors := ea.filterErrorsByType(errors, []string{"data_quality_error", "missing_data", "corrupted_data"})
	if len(dataQualityErrors) > 0 {
		rootCause := &RootCauseAnalysis{
			ID:                generateRootCauseID(),
			Category:          "data",
			RootCause:         "Data Quality Issues",
			Description:       "Problems with data quality, completeness, or integrity",
			Confidence:        ea.calculateRootCauseConfidence(dataQualityErrors, errors),
			Impact:            ea.determineImpact(dataQualityErrors),
			AffectedProcesses: []string{processName},
			Evidence:          ea.collectEvidence(dataQualityErrors, "data_quality"),
			ContributingFactors: []string{
				"Missing required data",
				"Data corruption",
				"Inconsistent data formats",
				"Data validation failures",
			},
			Recommendations: []string{
				"Implement data quality checks",
				"Add data validation rules",
				"Monitor data completeness",
				"Establish data governance",
			},
			Timeline: ea.buildTimeline(dataQualityErrors),
		}
		rootCauses = append(rootCauses, rootCause)
	}

	return rootCauses
}

// analyzeExternalRootCauses analyzes external service root causes
func (ea *ErrorAnalyzer) analyzeExternalRootCauses(errors []ErrorEntry, processName string) []*RootCauseAnalysis {
	var rootCauses []*RootCauseAnalysis

	// Check for external API issues
	apiErrors := ea.filterErrorsByType(errors, []string{"external_api_error", "third_party_service_error", "api_rate_limit"})
	if len(apiErrors) > 0 {
		rootCause := &RootCauseAnalysis{
			ID:                generateRootCauseID(),
			Category:          "external",
			RootCause:         "External API Issues",
			Description:       "Problems with external APIs and third-party services",
			Confidence:        ea.calculateRootCauseConfidence(apiErrors, errors),
			Impact:            ea.determineImpact(apiErrors),
			AffectedProcesses: []string{processName},
			Evidence:          ea.collectEvidence(apiErrors, "external_api"),
			ContributingFactors: []string{
				"External service downtime",
				"API rate limiting",
				"Authentication issues",
				"Service contract violations",
			},
			Recommendations: []string{
				"Implement circuit breakers",
				"Add retry mechanisms",
				"Monitor external service health",
				"Establish service level agreements",
			},
			Timeline: ea.buildTimeline(apiErrors),
		}
		rootCauses = append(rootCauses, rootCause)
	}

	return rootCauses
}

// analyzeConfigurationRootCauses analyzes configuration-related root causes
func (ea *ErrorAnalyzer) analyzeConfigurationRootCauses(errors []ErrorEntry, processName string) []*RootCauseAnalysis {
	var rootCauses []*RootCauseAnalysis

	// Check for configuration errors
	configErrors := ea.filterErrorsByType(errors, []string{"configuration_error", "missing_config", "invalid_config"})
	if len(configErrors) > 0 {
		rootCause := &RootCauseAnalysis{
			ID:                generateRootCauseID(),
			Category:          "configuration",
			RootCause:         "Configuration Issues",
			Description:       "Problems with application configuration and settings",
			Confidence:        ea.calculateRootCauseConfidence(configErrors, errors),
			Impact:            ea.determineImpact(configErrors),
			AffectedProcesses: []string{processName},
			Evidence:          ea.collectEvidence(configErrors, "configuration"),
			ContributingFactors: []string{
				"Missing configuration values",
				"Invalid configuration parameters",
				"Environment-specific issues",
				"Configuration deployment problems",
			},
			Recommendations: []string{
				"Validate configuration on startup",
				"Implement configuration management",
				"Use environment-specific configs",
				"Add configuration validation",
			},
			Timeline: ea.buildTimeline(configErrors),
		}
		rootCauses = append(rootCauses, rootCause)
	}

	return rootCauses
}

// analyzeErrorCorrelations analyzes correlations between different error types
func (ea *ErrorAnalyzer) analyzeErrorCorrelations(errors []ErrorEntry, processName string) []*ErrorCorrelation {
	var correlations []*ErrorCorrelation

	// Group errors by type
	errorGroups := ea.groupErrorsByType(errors)
	errorTypes := make([]string, 0, len(errorGroups))
	for errorType := range errorGroups {
		errorTypes = append(errorTypes, errorType)
	}

	// Analyze correlations between all pairs of error types
	for i := 0; i < len(errorTypes); i++ {
		for j := i + 1; j < len(errorTypes); j++ {
			correlation := ea.calculateErrorCorrelation(errorGroups[errorTypes[i]], errorGroups[errorTypes[j]])

			if correlation.Strength >= ea.config.CorrelationThreshold {
				correlations = append(correlations, correlation)
			}
		}
	}

	return correlations
}

// generateRecommendations generates recommendations based on analysis results
func (ea *ErrorAnalyzer) generateRecommendations(patterns []*ErrorPattern, rootCauses []*RootCauseAnalysis, correlations []*ErrorCorrelation) []string {
	var recommendations []string

	// Generate recommendations from patterns
	for _, pattern := range patterns {
		if pattern.Mitigation != nil {
			recommendations = append(recommendations, pattern.Mitigation.Implementation...)
		}
	}

	// Generate recommendations from root causes
	for _, rootCause := range rootCauses {
		recommendations = append(recommendations, rootCause.Recommendations...)
	}

	// Generate recommendations from correlations
	for _, correlation := range correlations {
		if correlation.Strength > 0.8 {
			recommendations = append(recommendations,
				fmt.Sprintf("Investigate causal relationship between %s and %s errors",
					correlation.PrimaryError, correlation.SecondaryError))
		}
	}

	// Remove duplicates
	return ea.removeDuplicateRecommendations(recommendations)
}

// performRiskAssessment performs risk assessment for errors
func (ea *ErrorAnalyzer) performRiskAssessment(errors []ErrorEntry, patterns []*ErrorPattern, rootCauses []*RootCauseAnalysis) *RiskAssessment {
	riskFactors := ea.identifyRiskFactors(errors, patterns, rootCauses)
	riskScore := ea.calculateRiskScore(riskFactors)
	overallRisk := ea.determineOverallRisk(riskScore)
	impactAnalysis := ea.performImpactAnalysis(errors, patterns, rootCauses)
	mitigationPriority := ea.determineMitigationPriority(riskFactors)

	return &RiskAssessment{
		OverallRisk:        overallRisk,
		RiskScore:          riskScore,
		RiskFactors:        riskFactors,
		ImpactAnalysis:     impactAnalysis,
		MitigationPriority: mitigationPriority,
	}
}

// performErrorTrendAnalysis performs trend analysis for errors
func (ea *ErrorAnalyzer) performErrorTrendAnalysis(errors []ErrorEntry, timeRange TimeRange) *ErrorTrendAnalysis {
	overallTrend := ea.calculateOverallTrend(errors, timeRange)
	trendConfidence := ea.calculateTrendConfidence(errors, timeRange)
	seasonalPatterns := ea.detectSeasonalPatterns(errors, timeRange)
	cyclicalPatterns := ea.detectCyclicalPatterns(errors, timeRange)
	predictions := ea.generatePredictions(errors, timeRange)

	return &ErrorTrendAnalysis{
		OverallTrend:     overallTrend,
		TrendConfidence:  trendConfidence,
		SeasonalPatterns: seasonalPatterns,
		CyclicalPatterns: cyclicalPatterns,
		Predictions:      predictions,
	}
}

// Helper methods for pattern detection
func (ea *ErrorAnalyzer) findRepeatingSequences(errors []ErrorEntry) [][]ErrorEntry {
	var sequences [][]ErrorEntry

	// Simple sequence detection - look for repeating patterns
	for i := 0; i < len(errors)-2; i++ {
		for j := i + 2; j < len(errors); j++ {
			sequence := errors[i:j]
			if ea.isRepeatingSequence(sequence, errors) {
				sequences = append(sequences, sequence)
			}
		}
	}

	return sequences
}

func (ea *ErrorAnalyzer) isRepeatingSequence(sequence []ErrorEntry, allErrors []ErrorEntry) bool {
	if len(sequence) < 2 {
		return false
	}

	// Count how many times this sequence appears
	count := 0
	for i := 0; i <= len(allErrors)-len(sequence); i++ {
		if ea.matchesSequence(sequence, allErrors[i:i+len(sequence)]) {
			count++
		}
	}

	return count >= ea.config.PatternDetectionThreshold
}

func (ea *ErrorAnalyzer) matchesSequence(seq1, seq2 []ErrorEntry) bool {
	if len(seq1) != len(seq2) {
		return false
	}

	for i := 0; i < len(seq1); i++ {
		if seq1[i].ErrorType != seq2[i].ErrorType {
			return false
		}
	}

	return true
}

func (ea *ErrorAnalyzer) extractErrorTypes(errors []ErrorEntry) []string {
	types := make(map[string]bool)
	for _, err := range errors {
		types[err.ErrorType] = true
	}

	result := make([]string, 0, len(types))
	for errorType := range types {
		result = append(result, errorType)
	}

	return result
}

func (ea *ErrorAnalyzer) calculateFrequency(sequence []ErrorEntry, allErrors []ErrorEntry) int {
	count := 0
	for i := 0; i <= len(allErrors)-len(sequence); i++ {
		if ea.matchesSequence(sequence, allErrors[i:i+len(sequence)]) {
			count++
		}
	}
	return count
}

func (ea *ErrorAnalyzer) calculatePatternConfidence(sequence []ErrorEntry, allErrors []ErrorEntry) float64 {
	frequency := ea.calculateFrequency(sequence, allErrors)
	totalPossible := len(allErrors) - len(sequence) + 1
	if totalPossible <= 0 {
		return 0.0
	}
	return float64(frequency) / float64(totalPossible)
}

// Helper methods for temporal analysis
func (ea *ErrorAnalyzer) analyzeHourlyDistribution(errors []ErrorEntry) map[int]int {
	distribution := make(map[int]int)

	for _, err := range errors {
		hour := err.Timestamp.Hour()
		distribution[hour]++
	}

	return distribution
}

func (ea *ErrorAnalyzer) analyzeDailyDistribution(errors []ErrorEntry) map[time.Weekday]int {
	distribution := make(map[time.Weekday]int)

	for _, err := range errors {
		day := err.Timestamp.Weekday()
		distribution[day]++
	}

	return distribution
}

func (ea *ErrorAnalyzer) detectPeakHours(hourlyDistribution map[int]int) []int {
	var peakHours []int
	total := 0

	for _, count := range hourlyDistribution {
		total += count
	}

	if total == 0 {
		return peakHours
	}

	average := float64(total) / 24.0
	threshold := average * 1.5 // 50% above average

	for hour, count := range hourlyDistribution {
		if float64(count) > threshold {
			peakHours = append(peakHours, hour)
		}
	}

	sort.Ints(peakHours)
	return peakHours
}

func (ea *ErrorAnalyzer) calculateTemporalFrequency(errors []ErrorEntry, peakHours []int) int {
	count := 0
	for _, err := range errors {
		hour := err.Timestamp.Hour()
		for _, peakHour := range peakHours {
			if hour == peakHour {
				count++
				break
			}
		}
	}
	return count
}

func (ea *ErrorAnalyzer) calculateTemporalConfidence(hourlyDistribution map[int]int, peakHours []int) float64 {
	if len(peakHours) == 0 {
		return 0.0
	}

	total := 0
	peakTotal := 0

	for hour, count := range hourlyDistribution {
		total += count
		for _, peakHour := range peakHours {
			if hour == peakHour {
				peakTotal += count
				break
			}
		}
	}

	if total == 0 {
		return 0.0
	}

	return float64(peakTotal) / float64(total)
}

// Helper methods for dependency analysis
func (ea *ErrorAnalyzer) findErrorDependencies(errors []ErrorEntry) []*ErrorCorrelation {
	var dependencies []*ErrorCorrelation

	// Simple dependency detection - look for errors that occur close in time
	for i := 0; i < len(errors); i++ {
		for j := i + 1; j < len(errors); j++ {
			timeDiff := errors[j].Timestamp.Sub(errors[i].Timestamp)
			if timeDiff > 0 && timeDiff < 5*time.Minute { // Within 5 minutes
				correlation := &ErrorCorrelation{
					ID:              generateCorrelationID(),
					PrimaryError:    errors[i].ErrorType,
					SecondaryError:  errors[j].ErrorType,
					CorrelationType: "temporal",
					Strength:        ea.calculateDependencyStrength(errors[i], errors[j]),
					Confidence:      ea.calculateDependencyConfidence(errors[i], errors[j]),
					Direction:       "forward",
					Evidence:        ea.collectDependencyEvidence(errors[i], errors[j]),
				}
				dependencies = append(dependencies, correlation)
			}
		}
	}

	return dependencies
}

func (ea *ErrorAnalyzer) calculateDependencyStrength(err1, err2 ErrorEntry) float64 {
	timeDiff := err2.Timestamp.Sub(err1.Timestamp)

	// Stronger correlation for closer time differences
	if timeDiff < 1*time.Minute {
		return 0.9
	} else if timeDiff < 2*time.Minute {
		return 0.7
	} else if timeDiff < 5*time.Minute {
		return 0.5
	}

	return 0.3
}

func (ea *ErrorAnalyzer) calculateDependencyConfidence(err1, err2 ErrorEntry) float64 {
	// Simple confidence calculation based on error types
	if err1.ErrorType == err2.ErrorType {
		return 0.8
	}

	// Check if errors are related (e.g., network errors)
	if ea.areErrorsRelated(err1.ErrorType, err2.ErrorType) {
		return 0.7
	}

	return 0.5
}

func (ea *ErrorAnalyzer) areErrorsRelated(type1, type2 string) bool {
	// Define related error types
	relatedGroups := [][]string{
		{"network_timeout", "connection_failed", "dns_error"},
		{"validation_error", "business_logic_error", "data_integrity_error"},
		{"out_of_memory", "cpu_overload", "disk_full"},
	}

	for _, group := range relatedGroups {
		found1, found2 := false, false
		for _, errorType := range group {
			if errorType == type1 {
				found1 = true
			}
			if errorType == type2 {
				found2 = true
			}
		}
		if found1 && found2 {
			return true
		}
	}

	return false
}

// Helper methods for correlation analysis
func (ea *ErrorAnalyzer) calculateErrorCorrelation(errors1, errors2 []ErrorEntry) *ErrorCorrelation {
	if len(errors1) == 0 || len(errors2) == 0 {
		return &ErrorCorrelation{
			ID:              generateCorrelationID(),
			PrimaryError:    "",
			SecondaryError:  "",
			CorrelationType: "none",
			Strength:        0.0,
			Confidence:      0.0,
		}
	}

	// Calculate correlation using time-based analysis
	strength := ea.calculateTimeBasedCorrelation(errors1, errors2)
	confidence := ea.calculateCorrelationConfidence(errors1, errors2)

	return &ErrorCorrelation{
		ID:              generateCorrelationID(),
		PrimaryError:    errors1[0].ErrorType,
		SecondaryError:  errors2[0].ErrorType,
		CorrelationType: "temporal",
		Strength:        strength,
		Confidence:      confidence,
		Direction:       "bidirectional",
		Evidence:        ea.collectCorrelationEvidence(errors1, errors2),
	}
}

func (ea *ErrorAnalyzer) calculateTimeBasedCorrelation(errors1, errors2 []ErrorEntry) float64 {
	// Simple correlation based on temporal proximity
	correlations := 0
	totalComparisons := 0

	for _, err1 := range errors1 {
		for _, err2 := range errors2 {
			totalComparisons++
			timeDiff := math.Abs(err1.Timestamp.Sub(err2.Timestamp).Seconds())
			if timeDiff < 60 { // Within 1 minute
				correlations++
			}
		}
	}

	if totalComparisons == 0 {
		return 0.0
	}

	return float64(correlations) / float64(totalComparisons)
}

func (ea *ErrorAnalyzer) calculateCorrelationConfidence(errors1, errors2 []ErrorEntry) float64 {
	// Confidence based on number of data points
	totalErrors := len(errors1) + len(errors2)
	if totalErrors < 5 {
		return 0.3
	} else if totalErrors < 10 {
		return 0.6
	} else {
		return 0.8
	}
}

// Utility methods
func (ea *ErrorAnalyzer) filterErrorsByType(errors []ErrorEntry, types []string) []ErrorEntry {
	var filtered []ErrorEntry

	typeMap := make(map[string]bool)
	for _, errorType := range types {
		typeMap[errorType] = true
	}

	for _, err := range errors {
		if typeMap[err.ErrorType] {
			filtered = append(filtered, err)
		}
	}

	return filtered
}

func (ea *ErrorAnalyzer) calculateRootCauseConfidence(filteredErrors []ErrorEntry, allErrors []ErrorEntry) float64 {
	if len(allErrors) == 0 {
		return 0.0
	}
	return float64(len(filteredErrors)) / float64(len(allErrors))
}

func (ea *ErrorAnalyzer) determineImpact(errors []ErrorEntry) string {
	if len(errors) == 0 {
		return "low"
	}

	// Simple impact determination based on error count
	if len(errors) > 100 {
		return "high"
	} else if len(errors) > 20 {
		return "medium"
	}
	return "low"
}

func (ea *ErrorAnalyzer) collectEvidence(errors []ErrorEntry, evidenceType string) []EvidenceItem {
	var evidence []EvidenceItem

	for _, err := range errors {
		evidenceItem := EvidenceItem{
			Type:        "error_log",
			Description: fmt.Sprintf("%s error: %s", err.ErrorType, err.ErrorMessage),
			Value:       err.ErrorMessage,
			Confidence:  0.8,
			Timestamp:   err.Timestamp,
			Source:      err.ProcessName,
			Metadata: map[string]interface{}{
				"error_type": err.ErrorType,
				"severity":   err.Severity,
			},
		}
		evidence = append(evidence, evidenceItem)
	}

	return evidence
}

func (ea *ErrorAnalyzer) buildTimeline(errors []ErrorEntry) []TimelineEvent {
	var timeline []TimelineEvent

	for _, err := range errors {
		event := TimelineEvent{
			Timestamp:   err.Timestamp,
			EventType:   "error_occurred",
			Description: err.ErrorMessage,
			Severity:    err.Severity,
			Process:     err.ProcessName,
			ErrorType:   err.ErrorType,
			Context: map[string]interface{}{
				"user_id":    err.UserID,
				"request_id": err.RequestID,
			},
		}
		timeline = append(timeline, event)
	}

	// Sort by timestamp
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Timestamp.Before(timeline[j].Timestamp)
	})

	return timeline
}

func (ea *ErrorAnalyzer) removeDuplicateRecommendations(recommendations []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, rec := range recommendations {
		if !seen[rec] {
			seen[rec] = true
			unique = append(unique, rec)
		}
	}

	return unique
}

func (ea *ErrorAnalyzer) identifyRiskFactors(errors []ErrorEntry, patterns []*ErrorPattern, rootCauses []*RootCauseAnalysis) []RiskFactor {
	var riskFactors []RiskFactor

	// High error rate risk factor
	if len(errors) > 50 {
		riskFactors = append(riskFactors, RiskFactor{
			Factor:      "High Error Rate",
			Description: "Large number of errors detected",
			RiskLevel:   "high",
			Probability: 0.8,
			Impact:      "high",
			Mitigation:  "Implement error prevention strategies",
		})
	}

	// Pattern-based risk factors
	for _, pattern := range patterns {
		if pattern.Frequency > 10 {
			riskFactors = append(riskFactors, RiskFactor{
				Factor:      fmt.Sprintf("Frequent Pattern: %s", pattern.Name),
				Description: pattern.Description,
				RiskLevel:   "medium",
				Probability: pattern.Confidence,
				Impact:      "medium",
				Mitigation:  "Investigate and address root cause",
			})
		}
	}

	// Root cause-based risk factors
	for _, rootCause := range rootCauses {
		if rootCause.Impact == "high" {
			riskFactors = append(riskFactors, RiskFactor{
				Factor:      fmt.Sprintf("Root Cause: %s", rootCause.RootCause),
				Description: rootCause.Description,
				RiskLevel:   "high",
				Probability: rootCause.Confidence,
				Impact:      rootCause.Impact,
				Mitigation:  strings.Join(rootCause.Recommendations, "; "),
			})
		}
	}

	return riskFactors
}

func (ea *ErrorAnalyzer) calculateRiskScore(riskFactors []RiskFactor) float64 {
	if len(riskFactors) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, factor := range riskFactors {
		score := factor.Probability
		if factor.RiskLevel == "high" {
			score *= 1.0
		} else if factor.RiskLevel == "medium" {
			score *= 0.6
		} else {
			score *= 0.3
		}
		totalScore += score
	}

	return math.Min(totalScore/float64(len(riskFactors)), 1.0)
}

func (ea *ErrorAnalyzer) determineOverallRisk(riskScore float64) string {
	if riskScore >= 0.7 {
		return "critical"
	} else if riskScore >= 0.5 {
		return "high"
	} else if riskScore >= 0.3 {
		return "medium"
	}
	return "low"
}

func (ea *ErrorAnalyzer) performImpactAnalysis(errors []ErrorEntry, patterns []*ErrorPattern, rootCauses []*RootCauseAnalysis) *ImpactAnalysis {
	// Simple impact analysis based on error characteristics
	userImpact := "low"
	businessImpact := "low"
	systemImpact := "low"
	financialImpact := "low"
	reputationImpact := "low"

	if len(errors) > 100 {
		userImpact = "high"
		businessImpact = "high"
		systemImpact = "high"
	}

	for _, rootCause := range rootCauses {
		if rootCause.Impact == "high" {
			systemImpact = "high"
			if rootCause.Category == "external" {
				businessImpact = "high"
			}
		}
	}

	return &ImpactAnalysis{
		UserImpact:       userImpact,
		BusinessImpact:   businessImpact,
		SystemImpact:     systemImpact,
		FinancialImpact:  financialImpact,
		ReputationImpact: reputationImpact,
		Details: map[string]interface{}{
			"total_errors": len(errors),
			"patterns":     len(patterns),
			"root_causes":  len(rootCauses),
		},
	}
}

func (ea *ErrorAnalyzer) determineMitigationPriority(riskFactors []RiskFactor) []string {
	var priorities []string

	// Sort risk factors by probability and impact
	sort.Slice(riskFactors, func(i, j int) bool {
		score1 := riskFactors[i].Probability
		score2 := riskFactors[j].Probability

		if riskFactors[i].RiskLevel == "high" {
			score1 *= 1.5
		}
		if riskFactors[j].RiskLevel == "high" {
			score2 *= 1.5
		}

		return score1 > score2
	})

	for _, factor := range riskFactors {
		priorities = append(priorities, factor.Factor)
	}

	return priorities
}

// Trend analysis methods
func (ea *ErrorAnalyzer) calculateOverallTrend(errors []ErrorEntry, timeRange TimeRange) string {
	if len(errors) < 2 {
		return "stable"
	}

	// Simple trend calculation based on error count over time
	midpoint := timeRange.Start.Add(timeRange.End.Sub(timeRange.Start) / 2)

	firstHalf := 0
	secondHalf := 0

	for _, err := range errors {
		if err.Timestamp.Before(midpoint) {
			firstHalf++
		} else {
			secondHalf++
		}
	}

	if secondHalf > int(float64(firstHalf)*1.5) {
		return "degrading"
	} else if firstHalf > int(float64(secondHalf)*1.5) {
		return "improving"
	}

	return "stable"
}

func (ea *ErrorAnalyzer) calculateTrendConfidence(errors []ErrorEntry, timeRange TimeRange) float64 {
	if len(errors) < 5 {
		return 0.3
	} else if len(errors) < 10 {
		return 0.6
	}
	return 0.8
}

func (ea *ErrorAnalyzer) detectSeasonalPatterns(errors []ErrorEntry, timeRange TimeRange) []SeasonalPattern {
	// Simple daily pattern detection
	hourlyDistribution := ea.analyzeHourlyDistribution(errors)
	peakHours := ea.detectPeakHours(hourlyDistribution)

	if len(peakHours) > 0 {
		pattern := SeasonalPattern{
			PatternType: "daily",
			Description: fmt.Sprintf("Peak error hours: %v", peakHours),
			Confidence:  0.7,
		}

		// Convert peak hours to times
		for _, hour := range peakHours {
			peakTime := time.Date(2000, 1, 1, hour, 0, 0, 0, time.UTC)
			pattern.PeakTimes = append(pattern.PeakTimes, peakTime)
		}

		return []SeasonalPattern{pattern}
	}

	return []SeasonalPattern{}
}

func (ea *ErrorAnalyzer) detectCyclicalPatterns(errors []ErrorEntry, timeRange TimeRange) []CyclicalPattern {
	// Simple cyclical pattern detection (e.g., daily cycles)
	if len(errors) > 10 {
		pattern := CyclicalPattern{
			CycleLength: 24 * time.Hour,
			Description: "Daily error cycle",
			Confidence:  0.6,
		}

		return []CyclicalPattern{pattern}
	}

	return []CyclicalPattern{}
}

func (ea *ErrorAnalyzer) generatePredictions(errors []ErrorEntry, timeRange TimeRange) []Prediction {
	var predictions []Prediction

	if len(errors) > 5 {
		// Simple prediction based on current trend
		trend := ea.calculateOverallTrend(errors, timeRange)

		prediction := Prediction{
			PredictionType: "error_rate",
			Value:          trend,
			Confidence:     0.6,
			Timeframe:      24 * time.Hour,
			Description:    fmt.Sprintf("Error rate trend: %s", trend),
		}

		predictions = append(predictions, prediction)
	}

	return predictions
}

// Additional helper methods
func (ea *ErrorAnalyzer) getEarliestTimestamp(errors1, errors2 []ErrorEntry) time.Time {
	var earliest time.Time

	if len(errors1) > 0 {
		earliest = errors1[0].Timestamp
	}
	if len(errors2) > 0 {
		if earliest.IsZero() || errors2[0].Timestamp.Before(earliest) {
			earliest = errors2[0].Timestamp
		}
	}

	return earliest
}

func (ea *ErrorAnalyzer) getLatestTimestamp(errors1, errors2 []ErrorEntry) time.Time {
	var latest time.Time

	if len(errors1) > 0 {
		latest = errors1[len(errors1)-1].Timestamp
	}
	if len(errors2) > 0 {
		if latest.IsZero() || errors2[len(errors2)-1].Timestamp.After(latest) {
			latest = errors2[len(errors2)-1].Timestamp
		}
	}

	return latest
}

func (ea *ErrorAnalyzer) collectDependencyEvidence(err1, err2 ErrorEntry) []EvidenceItem {
	return []EvidenceItem{
		{
			Type: "temporal",
			Description: fmt.Sprintf("Error %s occurred %v before %s",
				err1.ErrorType, err2.Timestamp.Sub(err1.Timestamp), err2.ErrorType),
			Value:      err2.Timestamp.Sub(err1.Timestamp).String(),
			Confidence: 0.8,
			Timestamp:  err2.Timestamp,
			Source:     "dependency_analysis",
		},
	}
}

func (ea *ErrorAnalyzer) collectCorrelationEvidence(errors1, errors2 []ErrorEntry) []EvidenceItem {
	return []EvidenceItem{
		{
			Type: "correlation",
			Description: fmt.Sprintf("Correlation between %s (%d errors) and %s (%d errors)",
				errors1[0].ErrorType, len(errors1), errors2[0].ErrorType, len(errors2)),
			Value:      len(errors1) + len(errors2),
			Confidence: 0.7,
			Timestamp:  time.Now(),
			Source:     "correlation_analysis",
		},
	}
}

// ID generation functions
func generateAnalysisID() string {
	return fmt.Sprintf("analysis_%d", time.Now().Unix())
}

func generatePatternID() string {
	return fmt.Sprintf("pattern_%d", time.Now().UnixNano())
}

func generateRootCauseID() string {
	return fmt.Sprintf("rootcause_%d", time.Now().UnixNano())
}

func generateCorrelationID() string {
	return fmt.Sprintf("correlation_%d", time.Now().UnixNano())
}
