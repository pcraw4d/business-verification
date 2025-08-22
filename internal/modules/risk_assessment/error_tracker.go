package risk_assessment

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrorTracker provides error tracking capabilities
type ErrorTracker struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
	mu     sync.RWMutex
	errors map[string]*ErrorStats
}

// ErrorStats contains error statistics
type ErrorStats struct {
	TotalRequests int
	TotalErrors   int
	ErrorRate     float64
	LastErrorTime time.Time
	ErrorTypes    map[string]int
	RecentErrors  []ErrorEntry
	ErrorTrend    string
}

// ErrorEntry contains individual error information
type ErrorEntry struct {
	Timestamp    time.Time
	ErrorType    string
	ErrorMessage string
	Context      string
	Severity     string
}

// ErrorReport contains error tracking report
type ErrorReport struct {
	OverallErrorRate float64
	ErrorRateByType  map[string]float64
	ErrorTrend       string
	Recommendations  []string
	ReportTimestamp  time.Time
}

// NewErrorTracker creates a new error tracker
func NewErrorTracker(config *RiskAssessmentConfig, logger *zap.Logger) *ErrorTracker {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ErrorTracker{
		config: config,
		logger: logger,
		errors: make(map[string]*ErrorStats),
	}
}

// TrackError tracks an error occurrence
func (et *ErrorTracker) TrackError(ctx context.Context, errorType, errorMessage, context string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	now := time.Now()
	stats, exists := et.errors[errorType]
	if !exists {
		stats = &ErrorStats{
			ErrorTypes:   make(map[string]int),
			RecentErrors: make([]ErrorEntry, 0),
		}
		et.errors[errorType] = stats
	}

	// Update statistics
	stats.TotalRequests++
	stats.TotalErrors++
	stats.LastErrorTime = now
	stats.ErrorRate = float64(stats.TotalErrors) / float64(stats.TotalRequests)

	// Track error types
	stats.ErrorTypes[errorMessage]++

	// Add to recent errors
	errorEntry := ErrorEntry{
		Timestamp:    now,
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		Context:      context,
		Severity:     et.determineSeverity(errorType),
	}

	stats.RecentErrors = append(stats.RecentErrors, errorEntry)

	// Keep only recent errors (last 100)
	if len(stats.RecentErrors) > 100 {
		stats.RecentErrors = stats.RecentErrors[1:]
	}

	// Update error trend
	stats.ErrorTrend = et.calculateErrorTrend(stats)

	et.logger.Warn("Error tracked",
		zap.String("error_type", errorType),
		zap.String("error_message", errorMessage),
		zap.String("context", context),
		zap.Float64("error_rate", stats.ErrorRate))
}

// TrackSuccess tracks a successful request
func (et *ErrorTracker) TrackSuccess(ctx context.Context, requestType string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	stats, exists := et.errors[requestType]
	if !exists {
		stats = &ErrorStats{
			ErrorTypes:   make(map[string]int),
			RecentErrors: make([]ErrorEntry, 0),
		}
		et.errors[requestType] = stats
	}

	stats.TotalRequests++
	stats.ErrorRate = float64(stats.TotalErrors) / float64(stats.TotalRequests)
}

// GetErrorReport gets a comprehensive error report
func (et *ErrorTracker) GetErrorReport() *ErrorReport {
	et.mu.RLock()
	defer et.mu.RUnlock()

	report := &ErrorReport{
		ErrorRateByType: make(map[string]float64),
		Recommendations: make([]string, 0),
		ReportTimestamp: time.Now(),
	}

	totalRequests := 0
	totalErrors := 0

	// Calculate overall statistics
	for requestType, stats := range et.errors {
		totalRequests += stats.TotalRequests
		totalErrors += stats.TotalErrors
		report.ErrorRateByType[requestType] = stats.ErrorRate
	}

	if totalRequests > 0 {
		report.OverallErrorRate = float64(totalErrors) / float64(totalRequests)
	}

	// Determine overall error trend
	report.ErrorTrend = et.determineOverallErrorTrend()

	// Generate recommendations
	report.Recommendations = et.generateErrorRecommendations(report)

	return report
}

// IsErrorRateAcceptable checks if the current error rate is acceptable
func (et *ErrorTracker) IsErrorRateAcceptable() bool {
	report := et.GetErrorReport()
	return report.OverallErrorRate <= et.config.MaxErrorRate
}

// GetErrorStats gets error statistics for a specific type
func (et *ErrorTracker) GetErrorStats(errorType string) *ErrorStats {
	et.mu.RLock()
	defer et.mu.RUnlock()

	if stats, exists := et.errors[errorType]; exists {
		return stats
	}
	return nil
}

// ResetErrorStats resets error statistics
func (et *ErrorTracker) ResetErrorStats(errorType string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	if stats, exists := et.errors[errorType]; exists {
		stats.TotalRequests = 0
		stats.TotalErrors = 0
		stats.ErrorRate = 0.0
		stats.ErrorTypes = make(map[string]int)
		stats.RecentErrors = make([]ErrorEntry, 0)
		stats.ErrorTrend = "stable"
	}
}

// determineSeverity determines the severity of an error
func (et *ErrorTracker) determineSeverity(errorType string) string {
	// Define severity levels based on error types
	severityMap := map[string]string{
		"network":        "high",
		"timeout":        "medium",
		"validation":     "low",
		"authentication": "high",
		"authorization":  "high",
		"rate_limit":     "medium",
		"server_error":   "high",
		"client_error":   "low",
	}

	if severity, exists := severityMap[errorType]; exists {
		return severity
	}
	return "medium"
}

// calculateErrorTrend calculates the error trend for a specific type
func (et *ErrorTracker) calculateErrorTrend(stats *ErrorStats) string {
	if len(stats.RecentErrors) < 2 {
		return "stable"
	}

	// Simple trend calculation based on recent errors
	recentCount := 0
	olderCount := 0

	now := time.Now()
	recentThreshold := now.Add(-5 * time.Minute)

	for _, error := range stats.RecentErrors {
		if error.Timestamp.After(recentThreshold) {
			recentCount++
		} else {
			olderCount++
		}
	}

	if recentCount > olderCount {
		return "increasing"
	} else if recentCount < olderCount {
		return "decreasing"
	}
	return "stable"
}

// determineOverallErrorTrend determines the overall error trend
func (et *ErrorTracker) determineOverallErrorTrend() string {
	report := et.GetErrorReport()

	if report.OverallErrorRate > et.config.MaxErrorRate {
		return "critical"
	} else if report.OverallErrorRate > et.config.MaxErrorRate*0.8 {
		return "warning"
	}
	return "stable"
}

// generateErrorRecommendations generates recommendations based on error patterns
func (et *ErrorTracker) generateErrorRecommendations(report *ErrorReport) []string {
	recommendations := make([]string, 0)

	if report.OverallErrorRate > et.config.MaxErrorRate {
		recommendations = append(recommendations, "Error rate exceeds acceptable threshold. Review system health and implement error handling improvements.")
	}

	// Check for specific error types with high rates
	for errorType, rate := range report.ErrorRateByType {
		if rate > 0.1 { // 10% error rate
			recommendations = append(recommendations, "High error rate for "+errorType+". Investigate root cause and implement fixes.")
		}
	}

	if report.ErrorTrend == "increasing" {
		recommendations = append(recommendations, "Error trend is increasing. Monitor system performance and implement preventive measures.")
	}

	return recommendations
}
