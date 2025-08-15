package observability

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// SuccessRateTracker provides comprehensive success rate tracking and analytics
type SuccessRateTracker struct {
	// Core tracking data
	endpointStats   map[string]*EndpointSuccessStats
	userStats       map[string]*UserSuccessStats
	timeWindowStats map[string]*TimeWindowStats
	errorTypeStats  map[string]*ErrorTypeStats
	overallStats    *OverallSuccessStats

	// Configuration
	config SuccessRateTrackerConfig

	// Thread safety
	mu sync.RWMutex
}

// SuccessRateTrackerConfig holds configuration for success rate tracking
type SuccessRateTrackerConfig struct {
	// Tracking intervals
	TrackingWindow    time.Duration `json:"tracking_window"`
	RollingWindowSize time.Duration `json:"rolling_window_size"`
	RetentionPeriod   time.Duration `json:"retention_period"`

	// Thresholds
	CriticalThreshold    float64 `json:"critical_threshold"`
	WarningThreshold     float64 `json:"warning_threshold"`
	DegradationThreshold float64 `json:"degradation_threshold"`

	// Analytics settings
	EnableTrendAnalysis       bool `json:"enable_trend_analysis"`
	EnableAnomalyDetection    bool `json:"enable_anomaly_detection"`
	EnablePredictiveAnalytics bool `json:"enable_predictive_analytics"`

	// Storage settings
	MaxDataPoints     int  `json:"max_data_points"`
	EnableCompression bool `json:"enable_compression"`
	EnableAggregation bool `json:"enable_aggregation"`
}

// EndpointSuccessStats tracks success rates for specific endpoints
type EndpointSuccessStats struct {
	Endpoint           string `json:"endpoint"`
	TotalRequests      int64  `json:"total_requests"`
	SuccessfulRequests int64  `json:"successful_requests"`
	FailedRequests     int64  `json:"failed_requests"`
	TimeoutRequests    int64  `json:"timeout_requests"`

	// Success rates
	OverallSuccessRate  float64 `json:"overall_success_rate"`
	RecentSuccessRate   float64 `json:"recent_success_rate"`
	TrendingSuccessRate float64 `json:"trending_success_rate"`

	// Performance metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`

	// Error breakdown
	ErrorBreakdown map[string]int64 `json:"error_breakdown"`

	// Timestamps
	FirstSeen   time.Time `json:"first_seen"`
	LastUpdated time.Time `json:"last_updated"`

	// Historical data
	HistoricalData []SuccessRateDataPoint `json:"historical_data"`
}

// UserSuccessStats tracks success rates for specific users
type UserSuccessStats struct {
	UserID             string `json:"user_id"`
	TotalRequests      int64  `json:"total_requests"`
	SuccessfulRequests int64  `json:"successful_requests"`
	FailedRequests     int64  `json:"failed_requests"`

	// Success rates
	OverallSuccessRate float64 `json:"overall_success_rate"`
	RecentSuccessRate  float64 `json:"recent_success_rate"`

	// Usage patterns
	MostUsedEndpoints []string  `json:"most_used_endpoints"`
	PeakUsageTime     time.Time `json:"peak_usage_time"`

	// Timestamps
	FirstSeen   time.Time `json:"first_seen"`
	LastUpdated time.Time `json:"last_updated"`
}

// TimeWindowStats tracks success rates across different time windows
type TimeWindowStats struct {
	WindowStart        time.Time `json:"window_start"`
	WindowEnd          time.Time `json:"window_end"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	SuccessRate        float64   `json:"success_rate"`

	// Peak periods
	PeakRequestTime time.Time `json:"peak_request_time"`
	PeakConcurrency int       `json:"peak_concurrency"`

	// Error patterns
	ErrorSpikes []ErrorSpike `json:"error_spikes"`
}

// ErrorTypeStats tracks success rates by error type
type ErrorTypeStats struct {
	ErrorType         string  `json:"error_type"`
	TotalOccurrences  int64   `json:"total_occurrences"`
	RecentOccurrences int64   `json:"recent_occurrences"`
	ImpactScore       float64 `json:"impact_score"`

	// Affected endpoints
	AffectedEndpoints []string `json:"affected_endpoints"`

	// Resolution tracking
	ResolutionTime time.Duration `json:"resolution_time"`
	AutoResolved   bool          `json:"auto_resolved"`

	// Timestamps
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// OverallSuccessStats tracks system-wide success rates
type OverallSuccessStats struct {
	// Current metrics
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`
	TimeoutRequests    int64 `json:"timeout_requests"`

	// Success rates
	OverallSuccessRate  float64 `json:"overall_success_rate"`
	RecentSuccessRate   float64 `json:"recent_success_rate"`
	TrendingSuccessRate float64 `json:"trending_success_rate"`

	// System health
	HealthScore      float64 `json:"health_score"`
	DegradationTrend string  `json:"degradation_trend"` // improving, stable, degrading

	// Performance indicators
	AverageResponseTime time.Duration `json:"average_response_time"`
	PeakThroughput      float64       `json:"peak_throughput"`

	// Timestamps
	LastUpdated time.Time `json:"last_updated"`
}

// SuccessRateDataPoint represents a single data point in success rate history
type SuccessRateDataPoint struct {
	Timestamp          time.Time     `json:"timestamp"`
	SuccessRate        float64       `json:"success_rate"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	ResponseTime       time.Duration `json:"response_time"`
}

// ErrorSpike represents a period of increased error activity
type ErrorSpike struct {
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	ErrorRate         float64   `json:"error_rate"`
	AffectedEndpoints []string  `json:"affected_endpoints"`
	RootCause         string    `json:"root_cause"`
	Resolution        string    `json:"resolution"`
}

// SuccessRateRequest represents a request to track
type SuccessRateRequest struct {
	Endpoint     string        `json:"endpoint"`
	UserID       string        `json:"user_id"`
	Success      bool          `json:"success"`
	ErrorType    string        `json:"error_type"`
	ResponseTime time.Duration `json:"response_time"`
	DataSize     int64         `json:"data_size"`
	Timestamp    time.Time     `json:"timestamp"`
}

// NewSuccessRateTracker creates a new success rate tracker
func NewSuccessRateTracker(config SuccessRateTrackerConfig) *SuccessRateTracker {
	if config.TrackingWindow == 0 {
		config.TrackingWindow = 1 * time.Minute
	}
	if config.RollingWindowSize == 0 {
		config.RollingWindowSize = 24 * time.Hour
	}
	if config.RetentionPeriod == 0 {
		config.RetentionPeriod = 30 * 24 * time.Hour // 30 days
	}
	if config.CriticalThreshold == 0 {
		config.CriticalThreshold = 0.90
	}
	if config.WarningThreshold == 0 {
		config.WarningThreshold = 0.95
	}
	if config.DegradationThreshold == 0 {
		config.DegradationThreshold = 0.98
	}
	if config.MaxDataPoints == 0 {
		config.MaxDataPoints = 10000
	}

	return &SuccessRateTracker{
		endpointStats:   make(map[string]*EndpointSuccessStats),
		userStats:       make(map[string]*UserSuccessStats),
		timeWindowStats: make(map[string]*TimeWindowStats),
		errorTypeStats:  make(map[string]*ErrorTypeStats),
		overallStats:    &OverallSuccessStats{},
		config:          config,
	}
}

// TrackRequest tracks a single request for success rate analysis
func (srt *SuccessRateTracker) TrackRequest(ctx context.Context, req *SuccessRateRequest) error {
	srt.mu.Lock()
	defer srt.mu.Unlock()

	// Update overall stats
	srt.updateOverallStats(req)

	// Update endpoint stats
	srt.updateEndpointStats(req)

	// Update user stats
	srt.updateUserStats(req)

	// Update time window stats
	srt.updateTimeWindowStats(req)

	// Update error type stats if applicable
	if !req.Success && req.ErrorType != "" {
		srt.updateErrorTypeStats(req)
	}

	// Update timestamps
	srt.overallStats.LastUpdated = time.Now()

	return nil
}

// updateOverallStats updates system-wide success rate statistics
func (srt *SuccessRateTracker) updateOverallStats(req *SuccessRateRequest) {
	srt.overallStats.TotalRequests++

	if req.Success {
		srt.overallStats.SuccessfulRequests++
	} else {
		srt.overallStats.FailedRequests++
		if req.ErrorType == "timeout" {
			srt.overallStats.TimeoutRequests++
		}
	}

	// Calculate success rates
	if srt.overallStats.TotalRequests > 0 {
		srt.overallStats.OverallSuccessRate = float64(srt.overallStats.SuccessfulRequests) / float64(srt.overallStats.TotalRequests)
	}

	// Update health score based on success rate
	srt.overallStats.HealthScore = srt.calculateHealthScore()

	// Update degradation trend
	srt.overallStats.DegradationTrend = srt.calculateDegradationTrend()
}

// updateEndpointStats updates endpoint-specific success rate statistics
func (srt *SuccessRateTracker) updateEndpointStats(req *SuccessRateRequest) {
	stats, exists := srt.endpointStats[req.Endpoint]
	if !exists {
		stats = &EndpointSuccessStats{
			Endpoint:       req.Endpoint,
			FirstSeen:      req.Timestamp,
			ErrorBreakdown: make(map[string]int64),
			HistoricalData: make([]SuccessRateDataPoint, 0),
		}
		srt.endpointStats[req.Endpoint] = stats
	}

	stats.TotalRequests++

	if req.Success {
		stats.SuccessfulRequests++
	} else {
		stats.FailedRequests++
		if req.ErrorType != "" {
			stats.ErrorBreakdown[req.ErrorType]++
		}
		if req.ErrorType == "timeout" {
			stats.TimeoutRequests++
		}
	}

	// Calculate success rates
	if stats.TotalRequests > 0 {
		stats.OverallSuccessRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests)
	}

	// Update response time metrics
	srt.updateResponseTimeMetrics(stats, req.ResponseTime)

	// Add to historical data
	srt.addHistoricalDataPoint(stats, req)

	stats.LastUpdated = req.Timestamp
}

// updateUserStats updates user-specific success rate statistics
func (srt *SuccessRateTracker) updateUserStats(req *SuccessRateRequest) {
	stats, exists := srt.userStats[req.UserID]
	if !exists {
		stats = &UserSuccessStats{
			UserID:            req.UserID,
			FirstSeen:         req.Timestamp,
			MostUsedEndpoints: make([]string, 0),
		}
		srt.userStats[req.UserID] = stats
	}

	stats.TotalRequests++

	if req.Success {
		stats.SuccessfulRequests++
	} else {
		stats.FailedRequests++
	}

	// Calculate success rates
	if stats.TotalRequests > 0 {
		stats.OverallSuccessRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests)
	}

	// Update most used endpoints
	srt.updateMostUsedEndpoints(stats, req.Endpoint)

	stats.LastUpdated = req.Timestamp
}

// updateTimeWindowStats updates time window-specific success rate statistics
func (srt *SuccessRateTracker) updateTimeWindowStats(req *SuccessRateRequest) {
	windowKey := srt.getTimeWindowKey(req.Timestamp)

	stats, exists := srt.timeWindowStats[windowKey]
	if !exists {
		windowStart, windowEnd := srt.getTimeWindowBounds(req.Timestamp)
		stats = &TimeWindowStats{
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
			ErrorSpikes: make([]ErrorSpike, 0),
		}
		srt.timeWindowStats[windowKey] = stats
	}

	stats.TotalRequests++

	if req.Success {
		stats.SuccessfulRequests++
	} else {
		stats.FailedRequests++
	}

	// Calculate success rate
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests)
	}

	// Update peak request time
	if stats.PeakRequestTime.IsZero() || req.Timestamp.After(stats.PeakRequestTime) {
		stats.PeakRequestTime = req.Timestamp
	}
}

// updateErrorTypeStats updates error type-specific statistics
func (srt *SuccessRateTracker) updateErrorTypeStats(req *SuccessRateRequest) {
	stats, exists := srt.errorTypeStats[req.ErrorType]
	if !exists {
		stats = &ErrorTypeStats{
			ErrorType:         req.ErrorType,
			FirstSeen:         req.Timestamp,
			AffectedEndpoints: make([]string, 0),
		}
		srt.errorTypeStats[req.ErrorType] = stats
	}

	stats.TotalOccurrences++
	stats.LastSeen = req.Timestamp

	// Update affected endpoints
	srt.updateAffectedEndpoints(stats, req.Endpoint)

	// Calculate impact score
	stats.ImpactScore = srt.calculateErrorImpactScore(stats)
}

// GetEndpointSuccessRate returns success rate for a specific endpoint
func (srt *SuccessRateTracker) GetEndpointSuccessRate(endpoint string) (*EndpointSuccessStats, error) {
	srt.mu.RLock()
	defer srt.mu.RUnlock()

	stats, exists := srt.endpointStats[endpoint]
	if !exists {
		return nil, fmt.Errorf("endpoint %s not found", endpoint)
	}

	return stats, nil
}

// GetUserSuccessRate returns success rate for a specific user
func (srt *SuccessRateTracker) GetUserSuccessRate(userID string) (*UserSuccessStats, error) {
	srt.mu.RLock()
	defer srt.mu.RUnlock()

	stats, exists := srt.userStats[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	return stats, nil
}

// GetOverallSuccessRate returns system-wide success rate statistics
func (srt *SuccessRateTracker) GetOverallSuccessRate() *OverallSuccessStats {
	srt.mu.RLock()
	defer srt.mu.RUnlock()

	return srt.overallStats
}

// GetTopPerformingEndpoints returns the top performing endpoints by success rate
func (srt *SuccessRateTracker) GetTopPerformingEndpoints(limit int) []*EndpointSuccessStats {
	srt.mu.RLock()
	defer srt.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	// Create a slice of all endpoints
	endpoints := make([]*EndpointSuccessStats, 0, len(srt.endpointStats))
	for _, stats := range srt.endpointStats {
		endpoints = append(endpoints, stats)
	}

	// Sort by success rate (descending)
	sortEndpointsBySuccessRate(endpoints)

	// Return top N
	if len(endpoints) > limit {
		return endpoints[:limit]
	}

	return endpoints
}

// GetWorstPerformingEndpoints returns the worst performing endpoints by success rate
func (srt *SuccessRateTracker) GetWorstPerformingEndpoints(limit int) []*EndpointSuccessStats {
	srt.mu.RLock()
	defer srt.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	// Create a slice of all endpoints
	endpoints := make([]*EndpointSuccessStats, 0, len(srt.endpointStats))
	for _, stats := range srt.endpointStats {
		endpoints = append(endpoints, stats)
	}

	// Sort by success rate (ascending)
	sortEndpointsBySuccessRateAscending(endpoints)

	// Return top N
	if len(endpoints) > limit {
		return endpoints[:limit]
	}

	return endpoints
}

// GetSuccessRateTrend returns success rate trend over time
func (srt *SuccessRateTracker) GetSuccessRateTrend(duration time.Duration) ([]SuccessRateDataPoint, error) {
	srt.mu.RLock()
	defer srt.mu.RUnlock()

	startTime := time.Now().Add(-duration)
	var trend []SuccessRateDataPoint

	// Aggregate data from all endpoints
	for _, stats := range srt.endpointStats {
		for _, dataPoint := range stats.HistoricalData {
			if dataPoint.Timestamp.After(startTime) {
				trend = append(trend, dataPoint)
			}
		}
	}

	// Sort by timestamp
	sortSuccessRateDataPoints(trend)

	return trend, nil
}

// GetErrorAnalysis returns detailed error analysis
func (srt *SuccessRateTracker) GetErrorAnalysis() map[string]*ErrorTypeStats {
	srt.mu.RLock()
	defer srt.mu.RUnlock()

	// Create a copy to avoid race conditions
	errorAnalysis := make(map[string]*ErrorTypeStats)
	for errorType, stats := range srt.errorTypeStats {
		errorAnalysis[errorType] = stats
	}

	return errorAnalysis
}

// Helper methods

func (srt *SuccessRateTracker) calculateHealthScore() float64 {
	successRate := srt.overallStats.OverallSuccessRate

	// Health score is based on success rate with additional factors
	baseScore := successRate * 100

	// Penalize for high error rates
	if srt.overallStats.FailedRequests > 0 {
		errorPenalty := float64(srt.overallStats.FailedRequests) / float64(srt.overallStats.TotalRequests) * 20
		baseScore -= errorPenalty
	}

	// Ensure score is between 0 and 100
	if baseScore < 0 {
		return 0
	}
	if baseScore > 100 {
		return 100
	}

	return baseScore
}

func (srt *SuccessRateTracker) calculateDegradationTrend() string {
	// Simple trend calculation based on recent vs overall success rate
	// In a real implementation, this would use more sophisticated trend analysis

	if srt.overallStats.RecentSuccessRate > srt.overallStats.OverallSuccessRate+0.01 {
		return "improving"
	} else if srt.overallStats.RecentSuccessRate < srt.overallStats.OverallSuccessRate-0.01 {
		return "degrading"
	}

	return "stable"
}

func (srt *SuccessRateTracker) updateResponseTimeMetrics(stats *EndpointSuccessStats, responseTime time.Duration) {
	// Update average response time
	if stats.AverageResponseTime == 0 {
		stats.AverageResponseTime = responseTime
	} else {
		// Simple moving average
		stats.AverageResponseTime = (stats.AverageResponseTime + responseTime) / 2
	}

	// For P95 and P99, we'd need to maintain a sorted list of response times
	// This is a simplified implementation
	if responseTime > stats.P95ResponseTime {
		stats.P95ResponseTime = responseTime
	}
	if responseTime > stats.P99ResponseTime {
		stats.P99ResponseTime = responseTime
	}
}

func (srt *SuccessRateTracker) addHistoricalDataPoint(stats *EndpointSuccessStats, req *SuccessRateRequest) {
	dataPoint := SuccessRateDataPoint{
		Timestamp:          req.Timestamp,
		SuccessRate:        stats.OverallSuccessRate,
		TotalRequests:      stats.TotalRequests,
		SuccessfulRequests: stats.SuccessfulRequests,
		FailedRequests:     stats.FailedRequests,
		ResponseTime:       req.ResponseTime,
	}

	stats.HistoricalData = append(stats.HistoricalData, dataPoint)

	// Limit historical data points
	if len(stats.HistoricalData) > srt.config.MaxDataPoints {
		stats.HistoricalData = stats.HistoricalData[1:]
	}
}

func (srt *SuccessRateTracker) updateMostUsedEndpoints(stats *UserSuccessStats, endpoint string) {
	// Simple implementation - in a real system, we'd maintain a proper frequency count
	found := false
	for _, ep := range stats.MostUsedEndpoints {
		if ep == endpoint {
			found = true
			break
		}
	}

	if !found && len(stats.MostUsedEndpoints) < 10 {
		stats.MostUsedEndpoints = append(stats.MostUsedEndpoints, endpoint)
	}
}

func (srt *SuccessRateTracker) updateAffectedEndpoints(stats *ErrorTypeStats, endpoint string) {
	found := false
	for _, ep := range stats.AffectedEndpoints {
		if ep == endpoint {
			found = true
			break
		}
	}

	if !found {
		stats.AffectedEndpoints = append(stats.AffectedEndpoints, endpoint)
	}
}

func (srt *SuccessRateTracker) calculateErrorImpactScore(stats *ErrorTypeStats) float64 {
	// Impact score based on frequency and affected endpoints
	frequencyScore := float64(stats.TotalOccurrences) / 100.0     // Normalize by 100
	endpointScore := float64(len(stats.AffectedEndpoints)) / 10.0 // Normalize by 10

	return (frequencyScore + endpointScore) / 2.0
}

func (srt *SuccessRateTracker) getTimeWindowKey(timestamp time.Time) string {
	// Create a time window key based on tracking window
	windowStart := timestamp.Truncate(srt.config.TrackingWindow)
	return windowStart.Format("2006-01-02T15:04:05")
}

func (srt *SuccessRateTracker) getTimeWindowBounds(timestamp time.Time) (time.Time, time.Time) {
	windowStart := timestamp.Truncate(srt.config.TrackingWindow)
	windowEnd := windowStart.Add(srt.config.TrackingWindow)
	return windowStart, windowEnd
}

// Sorting functions

func sortEndpointsBySuccessRate(endpoints []*EndpointSuccessStats) {
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].OverallSuccessRate > endpoints[j].OverallSuccessRate
	})
}

func sortEndpointsBySuccessRateAscending(endpoints []*EndpointSuccessStats) {
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].OverallSuccessRate < endpoints[j].OverallSuccessRate
	})
}

func sortSuccessRateDataPoints(dataPoints []SuccessRateDataPoint) {
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Timestamp.Before(dataPoints[j].Timestamp)
	})
}
