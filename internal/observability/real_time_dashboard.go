package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RealTimeDashboard provides real-time performance monitoring and visualization
type RealTimeDashboard struct {
	// Core components
	successRateTracker *SuccessRateTracker
	performanceMonitor *PerformanceMonitor
	metricsCollector   *MetricsCollector

	// Dashboard state
	dashboardState *DashboardState
	config         RealTimeDashboardConfig

	// WebSocket connections for real-time updates
	connections map[string]*DashboardConnection
	connMutex   sync.RWMutex

	// Event channels
	updateChannel chan *DashboardUpdate
	stopChannel   chan struct{}

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *zap.Logger
}

// RealTimeDashboardConfig holds configuration for the real-time dashboard
type RealTimeDashboardConfig struct {
	// Update intervals
	MetricsUpdateInterval time.Duration `json:"metrics_update_interval"`
	DashboardRefreshRate  time.Duration `json:"dashboard_refresh_rate"`
	ConnectionTimeout     time.Duration `json:"connection_timeout"`

	// Display settings
	MaxDataPoints         int  `json:"max_data_points"`
	EnableRealTimeUpdates bool `json:"enable_real_time_updates"`
	EnableHistoricalView  bool `json:"enable_historical_view"`

	// Performance settings
	MaxConcurrentConnections int  `json:"max_concurrent_connections"`
	EnableCompression        bool `json:"enable_compression"`
	EnableCaching            bool `json:"enable_caching"`

	// Security settings
	RequireAuthentication bool     `json:"require_authentication"`
	AllowedOrigins        []string `json:"allowed_origins"`
	APIKeyRequired        bool     `json:"api_key_required"`
}

// DashboardState represents the current state of the dashboard
type DashboardState struct {
	// Real-time metrics
	CurrentMetrics *RealTimeMetrics `json:"current_metrics"`

	// Performance indicators
	PerformanceIndicators *PerformanceIndicators `json:"performance_indicators"`

	// System health
	SystemHealth *SystemHealthStatus `json:"system_health"`

	// Active alerts
	ActiveAlerts []*RealTimeDashboardAlert `json:"active_alerts"`

	// Top performers
	TopPerformingEndpoints   []*EndpointPerformance `json:"top_performing_endpoints"`
	WorstPerformingEndpoints []*EndpointPerformance `json:"worst_performing_endpoints"`

	// User activity
	UserActivity *UserActivityMetrics `json:"user_activity"`

	// Error analysis
	ErrorAnalysis *ErrorAnalysisSummary `json:"error_analysis"`

	// Timestamp
	LastUpdated time.Time `json:"last_updated"`
	UpdateCount int64     `json:"update_count"`
}

// RealTimeMetrics provides real-time performance metrics
type RealTimeMetrics struct {
	// Request metrics
	RequestsPerSecond  float64 `json:"requests_per_second"`
	ActiveConnections  int     `json:"active_connections"`
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`

	// Success rates
	OverallSuccessRate float64 `json:"overall_success_rate"`
	RecentSuccessRate  float64 `json:"recent_success_rate"`

	// Response times
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`

	// Throughput
	DataProcessedPerSecond float64 `json:"data_processed_per_second"`
	PeakThroughput         float64 `json:"peak_throughput"`

	// Resource usage
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIO   float64 `json:"network_io"`
}

// PerformanceIndicators provides key performance indicators
type PerformanceIndicators struct {
	// Health scores
	SystemHealthScore   float64 `json:"system_health_score"`
	APIHealthScore      float64 `json:"api_health_score"`
	DatabaseHealthScore float64 `json:"database_health_score"`

	// Performance trends
	ResponseTimeTrend string `json:"response_time_trend"` // improving, stable, degrading
	SuccessRateTrend  string `json:"success_rate_trend"`
	ThroughputTrend   string `json:"throughput_trend"`

	// Efficiency metrics
	CacheHitRate            float64 `json:"cache_hit_rate"`
	DatabaseQueryEfficiency float64 `json:"database_query_efficiency"`
	ResourceUtilization     float64 `json:"resource_utilization"`
}

// SystemHealthStatus provides system health information
type SystemHealthStatus struct {
	OverallStatus string `json:"overall_status"` // healthy, warning, critical
	StatusMessage string `json:"status_message"`

	// Component health
	APIStatus      string `json:"api_status"`
	DatabaseStatus string `json:"database_status"`
	CacheStatus    string `json:"cache_status"`
	QueueStatus    string `json:"queue_status"`

	// Uptime
	Uptime      time.Duration `json:"uptime"`
	LastRestart time.Time     `json:"last_restart"`

	// Issues
	ActiveIssues   int `json:"active_issues"`
	ResolvedIssues int `json:"resolved_issues"`
}

// RealTimeDashboardAlert represents an alert for the real-time dashboard
type RealTimeDashboardAlert struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	Severity       string     `json:"severity"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Timestamp      time.Time  `json:"timestamp"`
	Status         string     `json:"status"` // active, acknowledged, resolved
	AcknowledgedBy string     `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
}

// EndpointPerformance represents endpoint performance data
type EndpointPerformance struct {
	Endpoint            string        `json:"endpoint"`
	SuccessRate         float64       `json:"success_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	RequestCount        int64         `json:"request_count"`
	ErrorCount          int64         `json:"error_count"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// UserActivityMetrics provides user activity information
type UserActivityMetrics struct {
	ActiveUsers     int `json:"active_users"`
	TotalUsers      int `json:"total_users"`
	NewUsersToday   int `json:"new_users_today"`
	PeakConcurrency int `json:"peak_concurrency"`

	// Usage patterns
	MostActiveUsers        []*UserActivity `json:"most_active_users"`
	UserSessions           int64           `json:"user_sessions"`
	AverageSessionDuration time.Duration   `json:"average_session_duration"`
}

// UserActivity represents individual user activity
type UserActivity struct {
	UserID           string    `json:"user_id"`
	RequestCount     int64     `json:"request_count"`
	LastActivity     time.Time `json:"last_activity"`
	SuccessRate      float64   `json:"success_rate"`
	MostUsedEndpoint string    `json:"most_used_endpoint"`
}

// ErrorAnalysisSummary provides error analysis for the dashboard
type ErrorAnalysisSummary struct {
	TotalErrors int64   `json:"total_errors"`
	ErrorRate   float64 `json:"error_rate"`

	// Error breakdown
	ErrorTypes       map[string]int64 `json:"error_types"`
	MostCommonErrors []*ErrorSummary  `json:"most_common_errors"`

	// Impact analysis
	HighImpactErrors  int `json:"high_impact_errors"`
	ResolvedErrors    int `json:"resolved_errors"`
	PendingResolution int `json:"pending_resolution"`
}

// ErrorSummary represents error summary data
type ErrorSummary struct {
	ErrorType         string    `json:"error_type"`
	Count             int64     `json:"count"`
	ImpactScore       float64   `json:"impact_score"`
	LastOccurrence    time.Time `json:"last_occurrence"`
	AffectedEndpoints []string  `json:"affected_endpoints"`
}

// DashboardConnection represents a WebSocket connection for real-time updates
type DashboardConnection struct {
	ID            string
	Connection    *http.ResponseWriter
	LastActivity  time.Time
	Subscriptions []string
}

// DashboardUpdate represents a dashboard update event
type DashboardUpdate struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// MetricsCollector collects and aggregates metrics
type MetricsCollector struct {
	metrics map[string]interface{}
	mu      sync.RWMutex
}

// NewRealTimeDashboard creates a new real-time dashboard
func NewRealTimeDashboard(
	successRateTracker *SuccessRateTracker,
	performanceMonitor *PerformanceMonitor,
	config RealTimeDashboardConfig,
	logger *zap.Logger,
) *RealTimeDashboard {
	if config.MetricsUpdateInterval == 0 {
		config.MetricsUpdateInterval = 5 * time.Second
	}
	if config.DashboardRefreshRate == 0 {
		config.DashboardRefreshRate = 1 * time.Second
	}
	if config.ConnectionTimeout == 0 {
		config.ConnectionTimeout = 30 * time.Second
	}
	if config.MaxDataPoints == 0 {
		config.MaxDataPoints = 1000
	}
	if config.MaxConcurrentConnections == 0 {
		config.MaxConcurrentConnections = 100
	}

	return &RealTimeDashboard{
		successRateTracker: successRateTracker,
		performanceMonitor: performanceMonitor,
		metricsCollector:   &MetricsCollector{metrics: make(map[string]interface{})},
		dashboardState: &DashboardState{
			CurrentMetrics:           &RealTimeMetrics{},
			PerformanceIndicators:    &PerformanceIndicators{},
			SystemHealth:             &SystemHealthStatus{},
			ActiveAlerts:             make([]*RealTimeDashboardAlert, 0),
			TopPerformingEndpoints:   make([]*EndpointPerformance, 0),
			WorstPerformingEndpoints: make([]*EndpointPerformance, 0),
			UserActivity:             &UserActivityMetrics{},
			ErrorAnalysis:            &ErrorAnalysisSummary{},
			LastUpdated:              time.Now(),
		},
		config:        config,
		connections:   make(map[string]*DashboardConnection),
		updateChannel: make(chan *DashboardUpdate, 100),
		stopChannel:   make(chan struct{}),
		logger:        logger,
	}
}

// Start starts the real-time dashboard
func (rtd *RealTimeDashboard) Start(ctx context.Context) error {
	rtd.logger.Info("Starting real-time dashboard")

	// Start metrics collection
	go rtd.collectMetrics(ctx)

	// Start dashboard updates
	go rtd.updateDashboard(ctx)

	// Start connection management
	go rtd.manageConnections(ctx)

	rtd.logger.Info("Real-time dashboard started successfully")
	return nil
}

// Stop stops the real-time dashboard
func (rtd *RealTimeDashboard) Stop() {
	rtd.logger.Info("Stopping real-time dashboard")
	close(rtd.stopChannel)
}

// GetDashboardState returns the current dashboard state
func (rtd *RealTimeDashboard) GetDashboardState() *DashboardState {
	rtd.mu.RLock()
	defer rtd.mu.RUnlock()

	return rtd.dashboardState
}

// GetRealTimeMetrics returns current real-time metrics
func (rtd *RealTimeDashboard) GetRealTimeMetrics() *RealTimeMetrics {
	rtd.mu.RLock()
	defer rtd.mu.RUnlock()

	return rtd.dashboardState.CurrentMetrics
}

// GetPerformanceIndicators returns current performance indicators
func (rtd *RealTimeDashboard) GetPerformanceIndicators() *PerformanceIndicators {
	rtd.mu.RLock()
	defer rtd.mu.RUnlock()

	return rtd.dashboardState.PerformanceIndicators
}

// GetSystemHealth returns current system health status
func (rtd *RealTimeDashboard) GetSystemHealth() *SystemHealthStatus {
	rtd.mu.RLock()
	defer rtd.mu.RUnlock()

	return rtd.dashboardState.SystemHealth
}

// GetTopPerformingEndpoints returns top performing endpoints
func (rtd *RealTimeDashboard) GetTopPerformingEndpoints(limit int) []*EndpointPerformance {
	if limit <= 0 {
		limit = 10
	}

	topEndpoints := rtd.successRateTracker.GetTopPerformingEndpoints(limit)

	// Convert to dashboard format
	var result []*EndpointPerformance
	for _, endpoint := range topEndpoints {
		result = append(result, &EndpointPerformance{
			Endpoint:            endpoint.Endpoint,
			SuccessRate:         endpoint.OverallSuccessRate,
			AverageResponseTime: endpoint.AverageResponseTime,
			RequestCount:        endpoint.TotalRequests,
			ErrorCount:          endpoint.FailedRequests,
			LastUpdated:         endpoint.LastUpdated,
		})
	}

	return result
}

// GetWorstPerformingEndpoints returns worst performing endpoints
func (rtd *RealTimeDashboard) GetWorstPerformingEndpoints(limit int) []*EndpointPerformance {
	if limit <= 0 {
		limit = 10
	}

	worstEndpoints := rtd.successRateTracker.GetWorstPerformingEndpoints(limit)

	// Convert to dashboard format
	var result []*EndpointPerformance
	for _, endpoint := range worstEndpoints {
		result = append(result, &EndpointPerformance{
			Endpoint:            endpoint.Endpoint,
			SuccessRate:         endpoint.OverallSuccessRate,
			AverageResponseTime: endpoint.AverageResponseTime,
			RequestCount:        endpoint.TotalRequests,
			ErrorCount:          endpoint.FailedRequests,
			LastUpdated:         endpoint.LastUpdated,
		})
	}

	return result
}

// GetErrorAnalysis returns current error analysis
func (rtd *RealTimeDashboard) GetErrorAnalysis() *ErrorAnalysisSummary {
	errorAnalysis := rtd.successRateTracker.GetErrorAnalysis()

	// Calculate totals
	var totalErrors int64
	var mostCommonErrors []*ErrorSummary

	for errorType, stats := range errorAnalysis {
		totalErrors += stats.TotalOccurrences

		mostCommonErrors = append(mostCommonErrors, &ErrorSummary{
			ErrorType:         errorType,
			Count:             stats.TotalOccurrences,
			ImpactScore:       stats.ImpactScore,
			LastOccurrence:    stats.LastSeen,
			AffectedEndpoints: stats.AffectedEndpoints,
		})
	}

	// Sort by count (descending)
	sortErrorSummaries(mostCommonErrors)

	// Limit to top 10
	if len(mostCommonErrors) > 10 {
		mostCommonErrors = mostCommonErrors[:10]
	}

	overallStats := rtd.successRateTracker.GetOverallSuccessRate()
	errorRate := 1.0 - overallStats.OverallSuccessRate

	return &ErrorAnalysisSummary{
		TotalErrors:       totalErrors,
		ErrorRate:         errorRate,
		ErrorTypes:        make(map[string]int64),
		MostCommonErrors:  mostCommonErrors,
		HighImpactErrors:  0, // Would be calculated based on impact scores
		ResolvedErrors:    0, // Would be tracked from error resolution
		PendingResolution: len(errorAnalysis),
	}
}

// collectMetrics collects real-time metrics
func (rtd *RealTimeDashboard) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(rtd.config.MetricsUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rtd.stopChannel:
			return
		case <-ticker.C:
			rtd.updateMetrics()
		}
	}
}

// updateMetrics updates the current metrics
func (rtd *RealTimeDashboard) updateMetrics() {
	rtd.mu.Lock()
	defer rtd.mu.Unlock()

	// Get current performance metrics
	performanceMetrics := rtd.performanceMonitor.GetMetrics()
	overallStats := rtd.successRateTracker.GetOverallSuccessRate()

	// Update real-time metrics
	rtd.dashboardState.CurrentMetrics = &RealTimeMetrics{
		RequestsPerSecond:      performanceMetrics.RequestsPerSecond,
		ActiveConnections:      performanceMetrics.ConcurrentRequests,
		TotalRequests:          overallStats.TotalRequests,
		SuccessfulRequests:     overallStats.SuccessfulRequests,
		FailedRequests:         overallStats.FailedRequests,
		OverallSuccessRate:     overallStats.OverallSuccessRate,
		RecentSuccessRate:      overallStats.RecentSuccessRate,
		AverageResponseTime:    performanceMetrics.AverageResponseTime,
		P95ResponseTime:        performanceMetrics.P95ResponseTime,
		P99ResponseTime:        performanceMetrics.P99ResponseTime,
		DataProcessedPerSecond: float64(performanceMetrics.DataProcessingVolume) / 1024, // KB/s
		PeakThroughput:         performanceMetrics.RequestsPerSecond,
		CPUUsage:               performanceMetrics.CPUUsage,
		MemoryUsage:            performanceMetrics.MemoryUsage,
		DiskUsage:              performanceMetrics.DiskUsage,
		NetworkIO:              performanceMetrics.NetworkIO,
	}

	// Update performance indicators
	rtd.dashboardState.PerformanceIndicators = &PerformanceIndicators{
		SystemHealthScore:       overallStats.HealthScore,
		APIHealthScore:          rtd.calculateAPIHealthScore(),
		DatabaseHealthScore:     rtd.calculateDatabaseHealthScore(),
		ResponseTimeTrend:       overallStats.DegradationTrend,
		SuccessRateTrend:        overallStats.DegradationTrend,
		ThroughputTrend:         rtd.calculateThroughputTrend(),
		CacheHitRate:            0.85, // Mock value - would be calculated from cache metrics
		DatabaseQueryEfficiency: 0.92, // Mock value - would be calculated from DB metrics
		ResourceUtilization:     (performanceMetrics.CPUUsage + performanceMetrics.MemoryUsage) / 2,
	}

	// Update system health
	rtd.dashboardState.SystemHealth = &SystemHealthStatus{
		OverallStatus:  rtd.calculateOverallStatus(),
		StatusMessage:  rtd.generateStatusMessage(),
		APIStatus:      "healthy",
		DatabaseStatus: "healthy",
		CacheStatus:    "healthy",
		QueueStatus:    "healthy",
		Uptime:         time.Since(time.Now().Add(-24 * time.Hour)), // Mock uptime
		LastRestart:    time.Now().Add(-24 * time.Hour),
		ActiveIssues:   len(rtd.dashboardState.ActiveAlerts),
		ResolvedIssues: 0, // Would be tracked from issue resolution
	}

	// Update top/worst performing endpoints
	rtd.dashboardState.TopPerformingEndpoints = rtd.GetTopPerformingEndpoints(5)
	rtd.dashboardState.WorstPerformingEndpoints = rtd.GetWorstPerformingEndpoints(5)

	// Update user activity
	rtd.dashboardState.UserActivity = rtd.getUserActivityMetrics()

	// Update error analysis
	rtd.dashboardState.ErrorAnalysis = rtd.GetErrorAnalysis()

	// Update timestamp and count
	rtd.dashboardState.LastUpdated = time.Now()
	rtd.dashboardState.UpdateCount++

	// Send update to connected clients
	rtd.broadcastUpdate(&DashboardUpdate{
		Type:      "metrics_update",
		Data:      rtd.dashboardState,
		Timestamp: time.Now(),
	})
}

// updateDashboard updates the dashboard state
func (rtd *RealTimeDashboard) updateDashboard(ctx context.Context) {
	ticker := time.NewTicker(rtd.config.DashboardRefreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rtd.stopChannel:
			return
		case <-ticker.C:
			rtd.refreshDashboard()
		}
	}
}

// refreshDashboard refreshes the dashboard display
func (rtd *RealTimeDashboard) refreshDashboard() {
	// This would trigger UI updates for connected clients
	rtd.broadcastUpdate(&DashboardUpdate{
		Type:      "dashboard_refresh",
		Data:      rtd.dashboardState,
		Timestamp: time.Now(),
	})
}

// manageConnections manages WebSocket connections
func (rtd *RealTimeDashboard) manageConnections(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rtd.stopChannel:
			return
		case <-ticker.C:
			rtd.cleanupStaleConnections()
		}
	}
}

// cleanupStaleConnections removes stale connections
func (rtd *RealTimeDashboard) cleanupStaleConnections() {
	rtd.connMutex.Lock()
	defer rtd.connMutex.Unlock()

	cutoff := time.Now().Add(-rtd.config.ConnectionTimeout)

	for id, conn := range rtd.connections {
		if conn.LastActivity.Before(cutoff) {
			delete(rtd.connections, id)
			rtd.logger.Info("Removed stale connection", zap.String("connection_id", id))
		}
	}
}

// broadcastUpdate broadcasts an update to all connected clients
func (rtd *RealTimeDashboard) broadcastUpdate(update *DashboardUpdate) {
	rtd.connMutex.RLock()
	defer rtd.connMutex.RUnlock()

	// Convert update to JSON
	_, err := json.Marshal(update)
	if err != nil {
		rtd.logger.Error("Failed to marshal dashboard update", zap.Error(err))
		return
	}

	// Broadcast to all connections
	for id := range rtd.connections {
		// In a real implementation, this would send via WebSocket
		rtd.logger.Debug("Broadcasting update to connection",
			zap.String("connection_id", id),
			zap.String("update_type", update.Type))
	}
}

// Helper methods

func (rtd *RealTimeDashboard) calculateAPIHealthScore() float64 {
	overallStats := rtd.successRateTracker.GetOverallSuccessRate()

	// API health is based on success rate and response times
	baseScore := overallStats.OverallSuccessRate * 100

	// Penalize for high response times
	if overallStats.AverageResponseTime > 500*time.Millisecond {
		penalty := float64(overallStats.AverageResponseTime.Milliseconds()-500) / 10
		baseScore -= penalty
	}

	if baseScore < 0 {
		return 0
	}
	if baseScore > 100 {
		return 100
	}

	return baseScore
}

func (rtd *RealTimeDashboard) calculateDatabaseHealthScore() float64 {
	// Mock database health score
	// In a real implementation, this would be based on DB metrics
	return 95.0
}

func (rtd *RealTimeDashboard) calculateThroughputTrend() string {
	// Mock throughput trend calculation
	// In a real implementation, this would compare current vs historical throughput
	return "stable"
}

func (rtd *RealTimeDashboard) calculateOverallStatus() string {
	overallStats := rtd.successRateTracker.GetOverallSuccessRate()

	if overallStats.OverallSuccessRate >= 0.95 {
		return "healthy"
	} else if overallStats.OverallSuccessRate >= 0.90 {
		return "warning"
	} else {
		return "critical"
	}
}

func (rtd *RealTimeDashboard) generateStatusMessage() string {
	overallStats := rtd.successRateTracker.GetOverallSuccessRate()

	if overallStats.OverallSuccessRate >= 0.95 {
		return "System operating normally"
	} else if overallStats.OverallSuccessRate >= 0.90 {
		return "System experiencing minor issues"
	} else {
		return "System experiencing critical issues"
	}
}

func (rtd *RealTimeDashboard) getUserActivityMetrics() *UserActivityMetrics {
	// Mock user activity metrics
	// In a real implementation, this would be calculated from user tracking data

	return &UserActivityMetrics{
		ActiveUsers:            25,
		TotalUsers:             150,
		NewUsersToday:          5,
		PeakConcurrency:        50,
		MostActiveUsers:        []*UserActivity{},
		UserSessions:           200,
		AverageSessionDuration: 15 * time.Minute,
	}
}

// Sorting functions

func sortErrorSummaries(summaries []*ErrorSummary) {
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Count > summaries[j].Count
	})
}
