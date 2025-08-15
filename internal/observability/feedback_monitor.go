package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FeedbackMonitor provides comprehensive real-time feedback monitoring for classification features
type FeedbackMonitor struct {
	metrics   *FeedbackMetrics
	alerts    *FeedbackAlertManager
	processor *FeedbackProcessor
	dashboard *FeedbackDashboard
	config    FeedbackMonitorConfig
	mu        sync.RWMutex
}

// FeedbackMonitorConfig holds configuration for feedback monitoring
type FeedbackMonitorConfig struct {
	// Monitoring intervals
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`
	AlertCheckInterval        time.Duration `json:"alert_check_interval"`
	ProcessingInterval        time.Duration `json:"processing_interval"`
	DashboardRefreshInterval  time.Duration `json:"dashboard_refresh_interval"`

	// Feedback thresholds
	FeedbackVolumeThreshold     int           `json:"feedback_volume_threshold"`
	FeedbackSentimentThreshold  float64       `json:"feedback_sentiment_threshold"`
	FeedbackAccuracyThreshold   float64       `json:"feedback_accuracy_threshold"`
	FeedbackProcessingThreshold time.Duration `json:"feedback_processing_threshold"`

	// Alerting settings
	AlertChannels     []string      `json:"alert_channels"`
	EscalationEnabled bool          `json:"escalation_enabled"`
	EscalationDelay   time.Duration `json:"escalation_delay"`

	// Dashboard settings
	HistoricalDataRetention time.Duration `json:"historical_data_retention"`
	RealTimeUpdatesEnabled  bool          `json:"real_time_updates_enabled"`

	// Processing settings
	BatchProcessingEnabled bool          `json:"batch_processing_enabled"`
	BatchSize              int           `json:"batch_size"`
	MaxProcessingTime      time.Duration `json:"max_processing_time"`
}

// FeedbackMetrics tracks comprehensive feedback data
type FeedbackMetrics struct {
	// Overall feedback metrics
	TotalFeedbackCount       int64 `json:"total_feedback_count"`
	PositiveFeedbackCount    int64 `json:"positive_feedback_count"`
	NegativeFeedbackCount    int64 `json:"negative_feedback_count"`
	NeutralFeedbackCount     int64 `json:"neutral_feedback_count"`
	ProcessedFeedbackCount   int64 `json:"processed_feedback_count"`
	UnprocessedFeedbackCount int64 `json:"unprocessed_feedback_count"`

	// Method-specific feedback metrics
	WebsiteAnalysisFeedbackCount  int64 `json:"website_analysis_feedback_count"`
	WebSearchFeedbackCount        int64 `json:"web_search_feedback_count"`
	MLModelFeedbackCount          int64 `json:"ml_model_feedback_count"`
	KeywordBasedFeedbackCount     int64 `json:"keyword_based_feedback_count"`
	FuzzyMatchingFeedbackCount    int64 `json:"fuzzy_matching_feedback_count"`
	CrosswalkMappingFeedbackCount int64 `json:"crosswalk_mapping_feedback_count"`

	// Sentiment metrics
	AverageSentimentScore float64 `json:"average_sentiment_score"`
	PositiveSentimentRate float64 `json:"positive_sentiment_rate"`
	NegativeSentimentRate float64 `json:"negative_sentiment_rate"`
	NeutralSentimentRate  float64 `json:"neutral_sentiment_rate"`

	// Method-specific sentiment metrics
	WebsiteAnalysisSentiment  float64 `json:"website_analysis_sentiment"`
	WebSearchSentiment        float64 `json:"web_search_sentiment"`
	MLModelSentiment          float64 `json:"ml_model_sentiment"`
	KeywordBasedSentiment     float64 `json:"keyword_based_sentiment"`
	FuzzyMatchingSentiment    float64 `json:"fuzzy_matching_sentiment"`
	CrosswalkMappingSentiment float64 `json:"crosswalk_mapping_sentiment"`

	// Accuracy improvement metrics
	AccuracyImprovementRate     float64            `json:"accuracy_improvement_rate"`
	AccuracyImprovementByMethod map[string]float64 `json:"accuracy_improvement_by_method"`
	AccuracyImprovementTrend    float64            `json:"accuracy_improvement_trend"`

	// Processing metrics
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	ProcessingSuccessRate float64       `json:"processing_success_rate"`
	ProcessingErrorRate   float64       `json:"processing_error_rate"`
	ProcessingTimeoutRate float64       `json:"processing_timeout_rate"`

	// Validation metrics
	ValidationSuccessRate float64 `json:"validation_success_rate"`
	ValidationErrorRate   float64 `json:"validation_error_rate"`
	InvalidFeedbackCount  int64   `json:"invalid_feedback_count"`
	ValidFeedbackCount    int64   `json:"valid_feedback_count"`

	// Geographic and industry feedback metrics
	GeographicFeedback map[string]GeographicFeedbackData `json:"geographic_feedback"`
	IndustryFeedback   map[string]IndustryFeedbackData   `json:"industry_feedback"`

	// User satisfaction metrics
	UserSatisfactionScore    float64            `json:"user_satisfaction_score"`
	UserSatisfactionTrend    float64            `json:"user_satisfaction_trend"`
	UserSatisfactionByMethod map[string]float64 `json:"user_satisfaction_by_method"`

	// Quality metrics
	FeedbackQualityScore     float64 `json:"feedback_quality_score"`
	HighQualityFeedbackCount int64   `json:"high_quality_feedback_count"`
	LowQualityFeedbackCount  int64   `json:"low_quality_feedback_count"`

	// Timestamp
	LastUpdated      time.Time     `json:"last_updated"`
	CollectionWindow time.Duration `json:"collection_window"`
}

// GeographicFeedbackData represents feedback data for a specific geographic region
type GeographicFeedbackData struct {
	TotalCount          int64     `json:"total_count"`
	PositiveCount       int64     `json:"positive_count"`
	NegativeCount       int64     `json:"negative_count"`
	NeutralCount        int64     `json:"neutral_count"`
	SentimentScore      float64   `json:"sentiment_score"`
	AccuracyImprovement float64   `json:"accuracy_improvement"`
	UserSatisfaction    float64   `json:"user_satisfaction"`
	LastUpdated         time.Time `json:"last_updated"`
}

// IndustryFeedbackData represents feedback data for a specific industry
type IndustryFeedbackData struct {
	TotalCount          int64     `json:"total_count"`
	PositiveCount       int64     `json:"positive_count"`
	NegativeCount       int64     `json:"negative_count"`
	NeutralCount        int64     `json:"neutral_count"`
	SentimentScore      float64   `json:"sentiment_score"`
	AccuracyImprovement float64   `json:"accuracy_improvement"`
	UserSatisfaction    float64   `json:"user_satisfaction"`
	LastUpdated         time.Time `json:"last_updated"`
}

// FeedbackAlert represents a feedback-related alert
type FeedbackAlert struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`     // volume, sentiment, accuracy, processing, validation
	Severity        string     `json:"severity"` // low, medium, high, critical
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Method          string     `json:"method"` // website_analysis, web_search, ml_model, etc.
	Metric          string     `json:"metric"`
	CurrentValue    float64    `json:"current_value"`
	Threshold       float64    `json:"threshold"`
	Timestamp       time.Time  `json:"timestamp"`
	Status          string     `json:"status"` // active, acknowledged, resolved
	AcknowledgedBy  string     `json:"acknowledged_by,omitempty"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	Recommendations []string   `json:"recommendations"`
	Impact          string     `json:"impact"` // low, medium, high
}

// FeedbackProcessor manages feedback processing
type FeedbackProcessor struct {
	processingQueue chan FeedbackItem
	config          FeedbackMonitorConfig
	mu              sync.RWMutex
}

// FeedbackItem represents a single feedback item
type FeedbackItem struct {
	ID               string    `json:"id"`
	Method           string    `json:"method"`
	Sentiment        string    `json:"sentiment"` // positive, negative, neutral
	SentimentScore   float64   `json:"sentiment_score"`
	AccuracyRating   float64   `json:"accuracy_rating"`
	UserSatisfaction float64   `json:"user_satisfaction"`
	GeographicRegion string    `json:"geographic_region"`
	Industry         string    `json:"industry"`
	QualityScore     float64   `json:"quality_score"`
	Timestamp        time.Time `json:"timestamp"`
	Processed        bool      `json:"processed"`
	Validated        bool      `json:"validated"`
}

// FeedbackDashboard provides real-time feedback visualization
type FeedbackDashboard struct {
	// Real-time metrics
	CurrentMetrics *FeedbackMetrics `json:"current_metrics"`

	// Historical data
	HistoricalMetrics []*FeedbackMetrics `json:"historical_metrics"`

	// Alerts
	ActiveAlerts []*FeedbackAlert `json:"active_alerts"`

	// Method performance
	MethodPerformance map[string]MethodFeedbackData `json:"method_performance"`

	// Geographic performance
	GeographicPerformance map[string]GeographicFeedbackData `json:"geographic_performance"`

	// Industry performance
	IndustryPerformance map[string]IndustryFeedbackData `json:"industry_performance"`

	// Processing status
	ProcessingStatus ProcessingStatusData `json:"processing_status"`

	// Status
	OverallHealth string    `json:"overall_health"`
	LastUpdated   time.Time `json:"last_updated"`
}

// MethodFeedbackData represents feedback data for a specific classification method
type MethodFeedbackData struct {
	TotalCount          int64         `json:"total_count"`
	PositiveCount       int64         `json:"positive_count"`
	NegativeCount       int64         `json:"negative_count"`
	NeutralCount        int64         `json:"neutral_count"`
	SentimentScore      float64       `json:"sentiment_score"`
	AccuracyImprovement float64       `json:"accuracy_improvement"`
	UserSatisfaction    float64       `json:"user_satisfaction"`
	ProcessingTime      time.Duration `json:"processing_time"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// ProcessingStatusData represents feedback processing status
type ProcessingStatusData struct {
	QueueSize             int           `json:"queue_size"`
	ProcessingRate        float64       `json:"processing_rate"`
	SuccessRate           float64       `json:"success_rate"`
	ErrorRate             float64       `json:"error_rate"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	LastUpdated           time.Time     `json:"last_updated"`
}

// FeedbackAlertManager manages feedback alerts
type FeedbackAlertManager struct {
	alerts map[string]*FeedbackAlert
	config FeedbackMonitorConfig
	mu     sync.RWMutex
}

// NewFeedbackMonitor creates a new feedback monitor
func NewFeedbackMonitor(config FeedbackMonitorConfig) *FeedbackMonitor {
	if config.MetricsCollectionInterval == 0 {
		config.MetricsCollectionInterval = 30 * time.Second
	}
	if config.AlertCheckInterval == 0 {
		config.AlertCheckInterval = 60 * time.Second
	}
	if config.ProcessingInterval == 0 {
		config.ProcessingInterval = 10 * time.Second
	}
	if config.DashboardRefreshInterval == 0 {
		config.DashboardRefreshInterval = 15 * time.Second
	}

	monitor := &FeedbackMonitor{
		config: config,
		metrics: &FeedbackMetrics{
			AccuracyImprovementByMethod: make(map[string]float64),
			GeographicFeedback:          make(map[string]GeographicFeedbackData),
			IndustryFeedback:            make(map[string]IndustryFeedbackData),
			UserSatisfactionByMethod:    make(map[string]float64),
		},
		alerts: &FeedbackAlertManager{
			alerts: make(map[string]*FeedbackAlert),
			config: config,
		},
		processor: &FeedbackProcessor{
			processingQueue: make(chan FeedbackItem, 1000),
			config:          config,
		},
		dashboard: &FeedbackDashboard{
			MethodPerformance:     make(map[string]MethodFeedbackData),
			GeographicPerformance: make(map[string]GeographicFeedbackData),
			IndustryPerformance:   make(map[string]IndustryFeedbackData),
		},
	}

	return monitor
}

// Start starts the feedback monitor
func (fm *FeedbackMonitor) Start(ctx context.Context) error {
	// Start metrics collection
	go fm.collectMetrics(ctx)

	// Start alert checking
	go fm.checkAlerts(ctx)

	// Start feedback processing
	go fm.processFeedback(ctx)

	// Start dashboard updates
	go fm.updateDashboard(ctx)

	return nil
}

// collectMetrics collects feedback metrics
func (fm *FeedbackMonitor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(fm.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fm.updateMetrics()
		}
	}
}

// updateMetrics updates the current metrics
func (fm *FeedbackMonitor) updateMetrics() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Update timestamp
	fm.metrics.LastUpdated = time.Now()

	// Calculate derived metrics
	fm.calculateDerivedMetrics()

	// Store historical data
	fm.storeHistoricalData()
}

// calculateDerivedMetrics calculates derived metrics from raw data
func (fm *FeedbackMonitor) calculateDerivedMetrics() {
	// Calculate sentiment rates
	if fm.metrics.TotalFeedbackCount > 0 {
		fm.metrics.PositiveSentimentRate = float64(fm.metrics.PositiveFeedbackCount) / float64(fm.metrics.TotalFeedbackCount)
		fm.metrics.NegativeSentimentRate = float64(fm.metrics.NegativeFeedbackCount) / float64(fm.metrics.TotalFeedbackCount)
		fm.metrics.NeutralSentimentRate = float64(fm.metrics.NeutralFeedbackCount) / float64(fm.metrics.TotalFeedbackCount)
	}

	// Calculate processing rates
	if fm.metrics.TotalFeedbackCount > 0 {
		fm.metrics.ProcessingSuccessRate = float64(fm.metrics.ProcessedFeedbackCount) / float64(fm.metrics.TotalFeedbackCount)
		fm.metrics.ProcessingErrorRate = 1.0 - fm.metrics.ProcessingSuccessRate
	}

	// Calculate validation rates
	if fm.metrics.TotalFeedbackCount > 0 {
		fm.metrics.ValidationSuccessRate = float64(fm.metrics.ValidFeedbackCount) / float64(fm.metrics.TotalFeedbackCount)
		fm.metrics.ValidationErrorRate = float64(fm.metrics.InvalidFeedbackCount) / float64(fm.metrics.TotalFeedbackCount)
	}

	// Calculate user satisfaction
	if fm.metrics.TotalFeedbackCount > 0 {
		totalSatisfaction := fm.metrics.PositiveFeedbackCount*1.0 + fm.metrics.NeutralFeedbackCount*0.5
		fm.metrics.UserSatisfactionScore = float64(totalSatisfaction) / float64(fm.metrics.TotalFeedbackCount)
	}
}

// storeHistoricalData stores historical metrics data
func (fm *FeedbackMonitor) storeHistoricalData() {
	// Create a copy of current metrics
	metricsCopy := *fm.metrics
	fm.dashboard.HistoricalMetrics = append(fm.dashboard.HistoricalMetrics, &metricsCopy)

	// Limit historical data retention
	maxHistoricalData := int(fm.config.HistoricalDataRetention / fm.config.MetricsCollectionInterval)
	if len(fm.dashboard.HistoricalMetrics) > maxHistoricalData {
		fm.dashboard.HistoricalMetrics = fm.dashboard.HistoricalMetrics[1:]
	}
}

// checkAlerts checks for feedback alerts
func (fm *FeedbackMonitor) checkAlerts(ctx context.Context) {
	ticker := time.NewTicker(fm.config.AlertCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fm.checkVolumeAlerts()
			fm.checkSentimentAlerts()
			fm.checkAccuracyAlerts()
			fm.checkProcessingAlerts()
			fm.checkValidationAlerts()
		}
	}
}

// checkVolumeAlerts checks for feedback volume alerts
func (fm *FeedbackMonitor) checkVolumeAlerts() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if fm.metrics.TotalFeedbackCount > int64(fm.config.FeedbackVolumeThreshold) {
		fm.createAlert("volume", "warning", "High Feedback Volume",
			"Feedback volume is above threshold", "total_feedback_count",
			float64(fm.metrics.TotalFeedbackCount), float64(fm.config.FeedbackVolumeThreshold))
	}
}

// checkSentimentAlerts checks for sentiment alerts
func (fm *FeedbackMonitor) checkSentimentAlerts() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if fm.metrics.AverageSentimentScore < fm.config.FeedbackSentimentThreshold {
		fm.createAlert("sentiment", "warning", "Low Feedback Sentiment",
			"Average feedback sentiment is below threshold", "average_sentiment_score",
			fm.metrics.AverageSentimentScore, fm.config.FeedbackSentimentThreshold)
	}
}

// checkAccuracyAlerts checks for accuracy alerts
func (fm *FeedbackMonitor) checkAccuracyAlerts() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if fm.metrics.AccuracyImprovementRate < fm.config.FeedbackAccuracyThreshold {
		fm.createAlert("accuracy", "warning", "Low Accuracy Improvement",
			"Accuracy improvement rate is below threshold", "accuracy_improvement_rate",
			fm.metrics.AccuracyImprovementRate, fm.config.FeedbackAccuracyThreshold)
	}
}

// checkProcessingAlerts checks for processing alerts
func (fm *FeedbackMonitor) checkProcessingAlerts() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if fm.metrics.AverageProcessingTime > fm.config.FeedbackProcessingThreshold {
		fm.createAlert("processing", "warning", "High Feedback Processing Time",
			"Average feedback processing time is above threshold", "average_processing_time",
			float64(fm.metrics.AverageProcessingTime.Milliseconds()),
			float64(fm.config.FeedbackProcessingThreshold.Milliseconds()))
	}
}

// checkValidationAlerts checks for validation alerts
func (fm *FeedbackMonitor) checkValidationAlerts() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if fm.metrics.ValidationSuccessRate < 0.9 {
		fm.createAlert("validation", "warning", "Low Feedback Validation Rate",
			"Feedback validation success rate is below 90%", "validation_success_rate",
			fm.metrics.ValidationSuccessRate, 0.9)
	}
}

// createAlert creates a new feedback alert
func (fm *FeedbackMonitor) createAlert(alertType, severity, title, description, metric string, currentValue, threshold float64) {
	alert := &FeedbackAlert{
		ID:           fmt.Sprintf("feedback_alert_%d", time.Now().Unix()),
		Type:         alertType,
		Severity:     severity,
		Title:        title,
		Description:  description,
		Metric:       metric,
		CurrentValue: currentValue,
		Threshold:    threshold,
		Timestamp:    time.Now(),
		Status:       "active",
		Recommendations: []string{
			"Review feedback processing pipeline",
			"Check system resources",
			"Monitor feedback quality",
		},
		Impact: "medium",
	}

	fm.alerts.alerts[alert.ID] = alert

	// Send alert through configured channels
	fm.sendAlert(alert)
}

// sendAlert sends an alert through configured channels
func (fm *FeedbackMonitor) sendAlert(alert *FeedbackAlert) {
	// Implementation would send alerts through configured channels
	// (email, Slack, webhook, etc.)
	fmt.Printf("Feedback Alert: %s - %s (Severity: %s)\n", alert.Title, alert.Description, alert.Severity)
}

// processFeedback processes feedback items
func (fm *FeedbackMonitor) processFeedback(ctx context.Context) {
	ticker := time.NewTicker(fm.config.ProcessingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fm.processFeedbackBatch()
		}
	}
}

// processFeedbackBatch processes a batch of feedback items
func (fm *FeedbackMonitor) processFeedbackBatch() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Process feedback items from the queue
	processedCount := 0
	startTime := time.Now()

	for i := 0; i < fm.config.BatchSize && len(fm.processor.processingQueue) > 0; i++ {
		select {
		case feedbackItem := <-fm.processor.processingQueue:
			fm.processFeedbackItem(feedbackItem)
			processedCount++
		default:
			break
		}
	}

	// Update processing metrics
	if processedCount > 0 {
		processingTime := time.Since(startTime)
		fm.metrics.AverageProcessingTime = processingTime / time.Duration(processedCount)
		fm.metrics.ProcessedFeedbackCount += int64(processedCount)
	}
}

// processFeedbackItem processes a single feedback item
func (fm *FeedbackMonitor) processFeedbackItem(item FeedbackItem) {
	// Validate feedback item
	if fm.validateFeedbackItem(item) {
		fm.metrics.ValidFeedbackCount++
		item.Validated = true
	} else {
		fm.metrics.InvalidFeedbackCount++
		item.Validated = false
		return
	}

	// Update method-specific metrics
	fm.updateMethodMetrics(item)

	// Update geographic metrics
	fm.updateGeographicMetrics(item)

	// Update industry metrics
	fm.updateIndustryMetrics(item)

	// Update sentiment metrics
	fm.updateSentimentMetrics(item)

	// Mark as processed
	item.Processed = true
	fm.metrics.ProcessedFeedbackCount++
}

// validateFeedbackItem validates a feedback item
func (fm *FeedbackMonitor) validateFeedbackItem(item FeedbackItem) bool {
	// Basic validation checks
	if item.SentimentScore < 0 || item.SentimentScore > 1 {
		return false
	}
	if item.AccuracyRating < 0 || item.AccuracyRating > 1 {
		return false
	}
	if item.UserSatisfaction < 0 || item.UserSatisfaction > 1 {
		return false
	}
	if item.QualityScore < 0 || item.QualityScore > 1 {
		return false
	}

	return true
}

// updateMethodMetrics updates method-specific metrics
func (fm *FeedbackMonitor) updateMethodMetrics(item FeedbackItem) {
	switch item.Method {
	case "website_analysis":
		fm.metrics.WebsiteAnalysisFeedbackCount++
		fm.metrics.WebsiteAnalysisSentiment = item.SentimentScore
	case "web_search":
		fm.metrics.WebSearchFeedbackCount++
		fm.metrics.WebSearchSentiment = item.SentimentScore
	case "ml_model":
		fm.metrics.MLModelFeedbackCount++
		fm.metrics.MLModelSentiment = item.SentimentScore
	case "keyword_based":
		fm.metrics.KeywordBasedFeedbackCount++
		fm.metrics.KeywordBasedSentiment = item.SentimentScore
	case "fuzzy_matching":
		fm.metrics.FuzzyMatchingFeedbackCount++
		fm.metrics.FuzzyMatchingSentiment = item.SentimentScore
	case "crosswalk_mapping":
		fm.metrics.CrosswalkMappingFeedbackCount++
		fm.metrics.CrosswalkMappingSentiment = item.SentimentScore
	}

	// Update accuracy improvement
	if item.AccuracyRating > 0 {
		fm.metrics.AccuracyImprovementByMethod[item.Method] = item.AccuracyRating
	}

	// Update user satisfaction
	if item.UserSatisfaction > 0 {
		fm.metrics.UserSatisfactionByMethod[item.Method] = item.UserSatisfaction
	}
}

// updateGeographicMetrics updates geographic metrics
func (fm *FeedbackMonitor) updateGeographicMetrics(item FeedbackItem) {
	if item.GeographicRegion == "" {
		return
	}

	geoData, exists := fm.metrics.GeographicFeedback[item.GeographicRegion]
	if !exists {
		geoData = GeographicFeedbackData{}
	}

	geoData.TotalCount++
	if item.Sentiment == "positive" {
		geoData.PositiveCount++
	} else if item.Sentiment == "negative" {
		geoData.NegativeCount++
	} else {
		geoData.NeutralCount++
	}

	// Update sentiment score
	totalSentiment := geoData.PositiveCount*1.0 + geoData.NeutralCount*0.5
	geoData.SentimentScore = totalSentiment / float64(geoData.TotalCount)

	// Update accuracy improvement
	if item.AccuracyRating > 0 {
		geoData.AccuracyImprovement = item.AccuracyRating
	}

	// Update user satisfaction
	if item.UserSatisfaction > 0 {
		geoData.UserSatisfaction = item.UserSatisfaction
	}

	geoData.LastUpdated = time.Now()
	fm.metrics.GeographicFeedback[item.GeographicRegion] = geoData
}

// updateIndustryMetrics updates industry metrics
func (fm *FeedbackMonitor) updateIndustryMetrics(item FeedbackItem) {
	if item.Industry == "" {
		return
	}

	industryData, exists := fm.metrics.IndustryFeedback[item.Industry]
	if !exists {
		industryData = IndustryFeedbackData{}
	}

	industryData.TotalCount++
	if item.Sentiment == "positive" {
		industryData.PositiveCount++
	} else if item.Sentiment == "negative" {
		industryData.NegativeCount++
	} else {
		industryData.NeutralCount++
	}

	// Update sentiment score
	totalSentiment := industryData.PositiveCount*1.0 + industryData.NeutralCount*0.5
	industryData.SentimentScore = totalSentiment / float64(industryData.TotalCount)

	// Update accuracy improvement
	if item.AccuracyRating > 0 {
		industryData.AccuracyImprovement = item.AccuracyRating
	}

	// Update user satisfaction
	if item.UserSatisfaction > 0 {
		industryData.UserSatisfaction = item.UserSatisfaction
	}

	industryData.LastUpdated = time.Now()
	fm.metrics.IndustryFeedback[item.Industry] = industryData
}

// updateSentimentMetrics updates sentiment metrics
func (fm *FeedbackMonitor) updateSentimentMetrics(item FeedbackItem) {
	fm.metrics.TotalFeedbackCount++

	if item.Sentiment == "positive" {
		fm.metrics.PositiveFeedbackCount++
	} else if item.Sentiment == "negative" {
		fm.metrics.NegativeFeedbackCount++
	} else {
		fm.metrics.NeutralFeedbackCount++
	}

	// Update average sentiment score
	totalSentiment := fm.metrics.PositiveFeedbackCount*1.0 + fm.metrics.NeutralFeedbackCount*0.5
	fm.metrics.AverageSentimentScore = totalSentiment / float64(fm.metrics.TotalFeedbackCount)
}

// updateDashboard updates the feedback dashboard
func (fm *FeedbackMonitor) updateDashboard(ctx context.Context) {
	ticker := time.NewTicker(fm.config.DashboardRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fm.refreshDashboard()
		}
	}
}

// refreshDashboard refreshes the dashboard data
func (fm *FeedbackMonitor) refreshDashboard() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Update current metrics
	fm.dashboard.CurrentMetrics = fm.metrics

	// Update method performance
	fm.updateMethodPerformance()

	// Update geographic performance
	fm.updateGeographicPerformance()

	// Update industry performance
	fm.updateIndustryPerformance()

	// Update processing status
	fm.updateProcessingStatus()

	// Update overall health
	fm.updateOverallHealth()

	// Update timestamp
	fm.dashboard.LastUpdated = time.Now()
}

// updateMethodPerformance updates method performance data
func (fm *FeedbackMonitor) updateMethodPerformance() {
	methods := []string{"website_analysis", "web_search", "ml_model", "keyword_based", "fuzzy_matching", "crosswalk_mapping"}

	for _, method := range methods {
		var count int64
		var sentiment float64
		var accuracyImprovement float64
		var userSatisfaction float64

		switch method {
		case "website_analysis":
			count = fm.metrics.WebsiteAnalysisFeedbackCount
			sentiment = fm.metrics.WebsiteAnalysisSentiment
		case "web_search":
			count = fm.metrics.WebSearchFeedbackCount
			sentiment = fm.metrics.WebSearchSentiment
		case "ml_model":
			count = fm.metrics.MLModelFeedbackCount
			sentiment = fm.metrics.MLModelSentiment
		case "keyword_based":
			count = fm.metrics.KeywordBasedFeedbackCount
			sentiment = fm.metrics.KeywordBasedSentiment
		case "fuzzy_matching":
			count = fm.metrics.FuzzyMatchingFeedbackCount
			sentiment = fm.metrics.FuzzyMatchingSentiment
		case "crosswalk_mapping":
			count = fm.metrics.CrosswalkMappingFeedbackCount
			sentiment = fm.metrics.CrosswalkMappingSentiment
		}

		accuracyImprovement = fm.metrics.AccuracyImprovementByMethod[method]
		userSatisfaction = fm.metrics.UserSatisfactionByMethod[method]

		fm.dashboard.MethodPerformance[method] = MethodFeedbackData{
			TotalCount:          count,
			SentimentScore:      sentiment,
			AccuracyImprovement: accuracyImprovement,
			UserSatisfaction:    userSatisfaction,
			ProcessingTime:      fm.metrics.AverageProcessingTime,
			LastUpdated:         time.Now(),
		}
	}
}

// updateGeographicPerformance updates geographic performance data
func (fm *FeedbackMonitor) updateGeographicPerformance() {
	for region, data := range fm.metrics.GeographicFeedback {
		fm.dashboard.GeographicPerformance[region] = data
	}
}

// updateIndustryPerformance updates industry performance data
func (fm *FeedbackMonitor) updateIndustryPerformance() {
	for industry, data := range fm.metrics.IndustryFeedback {
		fm.dashboard.IndustryPerformance[industry] = data
	}
}

// updateProcessingStatus updates processing status data
func (fm *FeedbackMonitor) updateProcessingStatus() {
	queueSize := len(fm.processor.processingQueue)
	processingRate := float64(fm.metrics.ProcessedFeedbackCount) / float64(fm.metrics.TotalFeedbackCount)
	successRate := fm.metrics.ProcessingSuccessRate
	errorRate := fm.metrics.ProcessingErrorRate

	fm.dashboard.ProcessingStatus = ProcessingStatusData{
		QueueSize:             queueSize,
		ProcessingRate:        processingRate,
		SuccessRate:           successRate,
		ErrorRate:             errorRate,
		AverageProcessingTime: fm.metrics.AverageProcessingTime,
		LastUpdated:           time.Now(),
	}
}

// updateOverallHealth updates the overall health status
func (fm *FeedbackMonitor) updateOverallHealth() {
	// Determine overall health based on key metrics
	if fm.metrics.UserSatisfactionScore >= 0.8 && fm.metrics.ProcessingSuccessRate >= 0.95 && fm.metrics.ValidationSuccessRate >= 0.9 {
		fm.dashboard.OverallHealth = "healthy"
	} else if fm.metrics.UserSatisfactionScore >= 0.7 && fm.metrics.ProcessingSuccessRate >= 0.9 && fm.metrics.ValidationSuccessRate >= 0.8 {
		fm.dashboard.OverallHealth = "warning"
	} else {
		fm.dashboard.OverallHealth = "critical"
	}
}

// RecordFeedback records a new feedback item
func (fm *FeedbackMonitor) RecordFeedback(method, sentiment string, sentimentScore, accuracyRating, userSatisfaction float64, geographicRegion, industry string, qualityScore float64) {
	feedbackItem := FeedbackItem{
		ID:               fmt.Sprintf("feedback_%d", time.Now().UnixNano()),
		Method:           method,
		Sentiment:        sentiment,
		SentimentScore:   sentimentScore,
		AccuracyRating:   accuracyRating,
		UserSatisfaction: userSatisfaction,
		GeographicRegion: geographicRegion,
		Industry:         industry,
		QualityScore:     qualityScore,
		Timestamp:        time.Now(),
		Processed:        false,
		Validated:        false,
	}

	// Add to processing queue
	select {
	case fm.processor.processingQueue <- feedbackItem:
		// Successfully added to queue
	default:
		// Queue is full, log warning
		fmt.Printf("Warning: Feedback processing queue is full, dropping feedback item\n")
	}
}

// GetMetrics returns the current feedback metrics
func (fm *FeedbackMonitor) GetMetrics() *FeedbackMetrics {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.metrics
}

// GetDashboard returns the current dashboard data
func (fm *FeedbackMonitor) GetDashboard() *FeedbackDashboard {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.dashboard
}

// GetAlerts returns the current alerts
func (fm *FeedbackMonitor) GetAlerts() []*FeedbackAlert {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	alerts := make([]*FeedbackAlert, 0, len(fm.alerts.alerts))
	for _, alert := range fm.alerts.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetProcessingStatus returns the current processing status
func (fm *FeedbackMonitor) GetProcessingStatus() ProcessingStatusData {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.dashboard.ProcessingStatus
}

// GetMethodPerformance returns performance data for a specific method
func (fm *FeedbackMonitor) GetMethodPerformance(method string) (MethodFeedbackData, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	data, exists := fm.dashboard.MethodPerformance[method]
	return data, exists
}

// GetGeographicPerformance returns performance data for a specific geographic region
func (fm *FeedbackMonitor) GetGeographicPerformance(region string) (GeographicFeedbackData, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	data, exists := fm.metrics.GeographicFeedback[region]
	return data, exists
}

// GetIndustryPerformance returns performance data for a specific industry
func (fm *FeedbackMonitor) GetIndustryPerformance(industry string) (IndustryFeedbackData, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	data, exists := fm.metrics.IndustryFeedback[industry]
	return data, exists
}
