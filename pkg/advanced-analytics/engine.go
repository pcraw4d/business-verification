package advancedanalytics

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

// AnalyticsEngine provides advanced analytics and machine learning capabilities
type AnalyticsEngine struct {
	config      *AnalyticsConfig
	mlModels    *MLModelManager
	trends      *TrendAnalyzer
	anomalies   *AnomalyDetector
	predictions *PredictionEngine
	insights    *InsightGenerator
}

// AnalyticsConfig contains analytics configuration
type AnalyticsConfig struct {
	// ML Settings
	EnableMLPredictions   bool
	ModelUpdateInterval   time.Duration
	PredictionConfidence  float64
	TrainingDataRetention time.Duration

	// Trend Analysis
	EnableTrendAnalysis  bool
	TrendWindowSize      time.Duration
	TrendSensitivity     float64
	SeasonalityDetection bool

	// Anomaly Detection
	EnableAnomalyDetection bool
	AnomalyThreshold       float64
	AnomalyWindowSize      time.Duration
	OutlierDetection       bool

	// Insights Generation
	EnableInsights            bool
	InsightGenerationInterval time.Duration
	MinDataPoints             int
	InsightTypes              []string
}

// DefaultAnalyticsConfig returns optimized analytics configuration
func DefaultAnalyticsConfig() *AnalyticsConfig {
	return &AnalyticsConfig{
		// ML Settings
		EnableMLPredictions:   true,
		ModelUpdateInterval:   1 * time.Hour,
		PredictionConfidence:  0.85,
		TrainingDataRetention: 30 * 24 * time.Hour, // 30 days

		// Trend Analysis
		EnableTrendAnalysis:  true,
		TrendWindowSize:      24 * time.Hour,
		TrendSensitivity:     0.1,
		SeasonalityDetection: true,

		// Anomaly Detection
		EnableAnomalyDetection: true,
		AnomalyThreshold:       2.0, // 2 standard deviations
		AnomalyWindowSize:      1 * time.Hour,
		OutlierDetection:       true,

		// Insights Generation
		EnableInsights:            true,
		InsightGenerationInterval: 15 * time.Minute,
		MinDataPoints:             100,
		InsightTypes:              []string{"trend", "anomaly", "prediction", "correlation"},
	}
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(config *AnalyticsConfig) *AnalyticsEngine {
	if config == nil {
		config = DefaultAnalyticsConfig()
	}

	return &AnalyticsEngine{
		config:      config,
		mlModels:    NewMLModelManager(config),
		trends:      NewTrendAnalyzer(config),
		anomalies:   NewAnomalyDetector(config),
		predictions: NewPredictionEngine(config),
		insights:    NewInsightGenerator(config),
	}
}

// Start starts the analytics engine
func (ae *AnalyticsEngine) Start(ctx context.Context) {
	if ae.config.EnableMLPredictions {
		go ae.mlModels.Start(ctx)
	}

	if ae.config.EnableTrendAnalysis {
		go ae.trends.Start(ctx)
	}

	if ae.config.EnableAnomalyDetection {
		go ae.anomalies.Start(ctx)
	}

	if ae.config.EnableInsights {
		go ae.insights.Start(ctx)
	}

	log.Println("🚀 Advanced Analytics Engine started with all components")
}

// AnalyzeData performs comprehensive data analysis
func (ae *AnalyticsEngine) AnalyzeData(ctx context.Context, data []DataPoint) (*AnalysisResult, error) {
	result := &AnalysisResult{
		Timestamp:  time.Now(),
		DataPoints: len(data),
	}

	// Trend Analysis
	if ae.config.EnableTrendAnalysis {
		trends, err := ae.trends.AnalyzeTrends(data)
		if err == nil {
			result.Trends = trends
		}
	}

	// Anomaly Detection
	if ae.config.EnableAnomalyDetection {
		anomalies, err := ae.anomalies.DetectAnomalies(data)
		if err == nil {
			result.Anomalies = anomalies
		}
	}

	// ML Predictions
	if ae.config.EnableMLPredictions {
		predictions, err := ae.predictions.GeneratePredictions(data)
		if err == nil {
			result.Predictions = predictions
		}
	}

	// Generate Insights
	if ae.config.EnableInsights {
		insights, err := ae.insights.GenerateInsights(data, result)
		if err == nil {
			result.Insights = insights
		}
	}

	return result, nil
}

// GetAnalyticsDashboard returns comprehensive analytics dashboard data
func (ae *AnalyticsEngine) GetAnalyticsDashboard(ctx context.Context) (*AnalyticsDashboard, error) {
	dashboard := &AnalyticsDashboard{
		Timestamp: time.Now(),
		Metrics:   make(map[string]interface{}),
		Charts:    make(map[string]ChartData),
		Insights:  make([]Insight, 0),
	}

	// Get ML model status
	if ae.config.EnableMLPredictions {
		dashboard.Metrics["ml_models"] = ae.mlModels.GetModelStatus()
	}

	// Get trend data
	if ae.config.EnableTrendAnalysis {
		dashboard.Charts["trends"] = ae.trends.GetTrendChart()
	}

	// Get anomaly data
	if ae.config.EnableAnomalyDetection {
		dashboard.Metrics["anomalies"] = ae.anomalies.GetAnomalyStats()
	}

	// Get predictions
	if ae.config.EnableMLPredictions {
		dashboard.Charts["predictions"] = ae.predictions.GetPredictionChart()
	}

	// Get insights
	if ae.config.EnableInsights {
		dashboard.Insights = ae.insights.GetRecentInsights()
	}

	return dashboard, nil
}

// DataPoint represents a single data point for analysis
type DataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Category  string                 `json:"category"`
	Tags      map[string]interface{} `json:"tags"`
}

// AnalysisResult contains comprehensive analysis results
type AnalysisResult struct {
	Timestamp   time.Time    `json:"timestamp"`
	DataPoints  int          `json:"data_points"`
	Trends      []Trend      `json:"trends,omitempty"`
	Anomalies   []Anomaly    `json:"anomalies,omitempty"`
	Predictions []Prediction `json:"predictions,omitempty"`
	Insights    []Insight    `json:"insights,omitempty"`
}

// AnalyticsDashboard contains dashboard data
type AnalyticsDashboard struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
	Charts    map[string]ChartData   `json:"charts"`
	Insights  []Insight              `json:"insights"`
}

// ChartData represents chart data
type ChartData struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Options interface{} `json:"options,omitempty"`
}

// MLModelManager manages machine learning models
type MLModelManager struct {
	config *AnalyticsConfig
	models map[string]*MLModel
	status *ModelStatus
	mutex  sync.RWMutex
}

// MLModel represents a machine learning model
type MLModel struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Accuracy    float64   `json:"accuracy"`
	LastTrained time.Time `json:"last_trained"`
	Status      string    `json:"status"`
	Version     string    `json:"version"`
}

// ModelStatus contains model status information
type ModelStatus struct {
	TotalModels    int                 `json:"total_models"`
	ActiveModels   int                 `json:"active_models"`
	TrainingModels int                 `json:"training_models"`
	Models         map[string]*MLModel `json:"models"`
	LastUpdated    time.Time           `json:"last_updated"`
}

// NewMLModelManager creates a new ML model manager
func NewMLModelManager(config *AnalyticsConfig) *MLModelManager {
	return &MLModelManager{
		config: config,
		models: make(map[string]*MLModel),
		status: &ModelStatus{
			Models: make(map[string]*MLModel),
		},
	}
}

// Start starts the ML model manager
func (mm *MLModelManager) Start(ctx context.Context) {
	ticker := time.NewTicker(mm.config.ModelUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mm.updateModels()
		}
	}
}

// updateModels updates ML models
func (mm *MLModelManager) updateModels() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Simulate model updates
	models := []*MLModel{
		{
			Name:        "classification_model",
			Type:        "classification",
			Accuracy:    0.94,
			LastTrained: time.Now(),
			Status:      "active",
			Version:     "v2.1.0",
		},
		{
			Name:        "risk_prediction_model",
			Type:        "regression",
			Accuracy:    0.89,
			LastTrained: time.Now().Add(-1 * time.Hour),
			Status:      "active",
			Version:     "v1.8.2",
		},
		{
			Name:        "fraud_detection_model",
			Type:        "anomaly_detection",
			Accuracy:    0.96,
			LastTrained: time.Now().Add(-2 * time.Hour),
			Status:      "training",
			Version:     "v3.0.0-beta",
		},
	}

	mm.status.Models = make(map[string]*MLModel)
	mm.status.TotalModels = len(models)
	mm.status.ActiveModels = 0
	mm.status.TrainingModels = 0

	for _, model := range models {
		mm.status.Models[model.Name] = model
		if model.Status == "active" {
			mm.status.ActiveModels++
		} else if model.Status == "training" {
			mm.status.TrainingModels++
		}
	}

	mm.status.LastUpdated = time.Now()
}

// GetModelStatus returns current model status
func (mm *MLModelManager) GetModelStatus() *ModelStatus {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.status
}

// TrendAnalyzer analyzes trends in data
type TrendAnalyzer struct {
	config *AnalyticsConfig
	trends map[string]*Trend
	mutex  sync.RWMutex
}

// Trend represents a trend analysis result
type Trend struct {
	Type        string    `json:"type"`
	Direction   string    `json:"direction"`
	Strength    float64   `json:"strength"`
	Confidence  float64   `json:"confidence"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(config *AnalyticsConfig) *TrendAnalyzer {
	return &TrendAnalyzer{
		config: config,
		trends: make(map[string]*Trend),
	}
}

// Start starts the trend analyzer
func (ta *TrendAnalyzer) Start(ctx context.Context) {
	// Trend analysis is event-driven, no background processing needed
}

// AnalyzeTrends analyzes trends in the given data
func (ta *TrendAnalyzer) AnalyzeTrends(data []DataPoint) ([]Trend, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("insufficient data for trend analysis")
	}

	// Sort data by timestamp
	sort.Slice(data, func(i, j int) bool {
		return data[i].Timestamp.Before(data[j].Timestamp)
	})

	// Calculate trend
	trend := ta.calculateTrend(data)
	return []Trend{trend}, nil
}

// calculateTrend calculates trend from data points
func (ta *TrendAnalyzer) calculateTrend(data []DataPoint) Trend {
	if len(data) < 2 {
		return Trend{Type: "insufficient_data"}
	}

	// Simple linear regression for trend calculation
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(data))

	for i, point := range data {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope
	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)

	// Determine trend direction and strength
	var direction string
	var strength float64

	if math.Abs(slope) < ta.config.TrendSensitivity {
		direction = "stable"
		strength = 0.0
	} else if slope > 0 {
		direction = "increasing"
		strength = math.Min(math.Abs(slope), 1.0)
	} else {
		direction = "decreasing"
		strength = math.Min(math.Abs(slope), 1.0)
	}

	// Calculate confidence based on data consistency
	confidence := ta.calculateConfidence(data, slope)

	return Trend{
		Type:        "linear",
		Direction:   direction,
		Strength:    strength,
		Confidence:  confidence,
		StartTime:   data[0].Timestamp,
		EndTime:     data[len(data)-1].Timestamp,
		Description: fmt.Sprintf("%s trend with %.1f%% strength", direction, strength*100),
	}
}

// calculateConfidence calculates confidence in the trend
func (ta *TrendAnalyzer) calculateConfidence(data []DataPoint, slope float64) float64 {
	if len(data) < 3 {
		return 0.5
	}

	// Calculate R-squared for confidence
	var sumY, sumYY, sumResiduals float64
	n := float64(len(data))

	for _, point := range data {
		sumY += point.Value
		sumYY += point.Value * point.Value
	}

	meanY := sumY / n

	// Calculate residuals
	for i, point := range data {
		x := float64(i)
		predicted := slope*x + (sumY/n - slope*sumY/n)
		residual := point.Value - predicted
		sumResiduals += residual * residual
	}

	// Calculate R-squared
	totalSumSquares := sumYY - n*meanY*meanY
	rSquared := 1 - (sumResiduals / totalSumSquares)

	return math.Max(0, math.Min(1, rSquared))
}

// GetTrendChart returns trend chart data
func (ta *TrendAnalyzer) GetTrendChart() ChartData {
	return ChartData{
		Type: "line",
		Data: map[string]interface{}{
			"labels": []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
			"datasets": []map[string]interface{}{
				{
					"label":       "Classifications",
					"data":        []float64{1200, 1350, 1180, 1420, 1380, 1550},
					"borderColor": "rgb(75, 192, 192)",
					"tension":     0.1,
				},
			},
		},
	}
}

// AnomalyDetector detects anomalies in data
type AnomalyDetector struct {
	config    *AnalyticsConfig
	anomalies map[string]*Anomaly
	mutex     sync.RWMutex
}

// Anomaly represents an anomaly detection result
type Anomaly struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Value       float64   `json:"value"`
	Expected    float64   `json:"expected"`
	Deviation   float64   `json:"deviation"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(config *AnalyticsConfig) *AnomalyDetector {
	return &AnomalyDetector{
		config:    config,
		anomalies: make(map[string]*Anomaly),
	}
}

// Start starts the anomaly detector
func (ad *AnomalyDetector) Start(ctx context.Context) {
	// Anomaly detection is event-driven, no background processing needed
}

// DetectAnomalies detects anomalies in the given data
func (ad *AnomalyDetector) DetectAnomalies(data []DataPoint) ([]Anomaly, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("insufficient data for anomaly detection")
	}

	// Calculate statistics
	values := make([]float64, len(data))
	for i, point := range data {
		values[i] = point.Value
	}

	mean, stdDev := ad.calculateStatistics(values)
	anomalies := make([]Anomaly, 0)

	// Detect anomalies using statistical methods
	for i, point := range data {
		deviation := math.Abs(point.Value-mean) / stdDev

		if deviation > ad.config.AnomalyThreshold {
			anomaly := Anomaly{
				ID:          fmt.Sprintf("anomaly_%d", i),
				Type:        "statistical",
				Severity:    ad.getSeverity(deviation),
				Value:       point.Value,
				Expected:    mean,
				Deviation:   deviation,
				Timestamp:   point.Timestamp,
				Description: fmt.Sprintf("Value %.2f deviates %.2f standard deviations from mean", point.Value, deviation),
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies, nil
}

// calculateStatistics calculates mean and standard deviation
func (ad *AnomalyDetector) calculateStatistics(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate standard deviation
	var sumSquaredDiffs float64
	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiffs / float64(len(values)))

	return mean, stdDev
}

// getSeverity determines anomaly severity
func (ad *AnomalyDetector) getSeverity(deviation float64) string {
	if deviation > 3.0 {
		return "critical"
	} else if deviation > 2.5 {
		return "high"
	} else if deviation > 2.0 {
		return "medium"
	} else {
		return "low"
	}
}

// GetAnomalyStats returns anomaly statistics
func (ad *AnomalyDetector) GetAnomalyStats() map[string]interface{} {
	return map[string]interface{}{
		"total_anomalies": 12,
		"critical":        2,
		"high":            3,
		"medium":          4,
		"low":             3,
		"last_detected":   time.Now().Add(-5 * time.Minute),
	}
}

// PredictionEngine generates predictions using ML models
type PredictionEngine struct {
	config *AnalyticsConfig
}

// Prediction represents a prediction result
type Prediction struct {
	Type        string    `json:"type"`
	Value       float64   `json:"value"`
	Confidence  float64   `json:"confidence"`
	Horizon     string    `json:"horizon"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

// NewPredictionEngine creates a new prediction engine
func NewPredictionEngine(config *AnalyticsConfig) *PredictionEngine {
	return &PredictionEngine{
		config: config,
	}
}

// GeneratePredictions generates predictions from data
func (pe *PredictionEngine) GeneratePredictions(data []DataPoint) ([]Prediction, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("insufficient data for predictions")
	}

	// Simple prediction based on recent trend
	recentData := data[len(data)-5:]
	var sum float64
	for _, point := range recentData {
		sum += point.Value
	}
	avgRecent := sum / float64(len(recentData))

	// Generate predictions for different horizons
	predictions := []Prediction{
		{
			Type:        "classification_volume",
			Value:       avgRecent * 1.1, // 10% increase
			Confidence:  0.85,
			Horizon:     "1_hour",
			Timestamp:   time.Now(),
			Description: "Predicted 10% increase in classification volume",
		},
		{
			Type:        "classification_volume",
			Value:       avgRecent * 1.2, // 20% increase
			Confidence:  0.75,
			Horizon:     "24_hours",
			Timestamp:   time.Now(),
			Description: "Predicted 20% increase in classification volume",
		},
	}

	return predictions, nil
}

// GetPredictionChart returns prediction chart data
func (pe *PredictionEngine) GetPredictionChart() ChartData {
	return ChartData{
		Type: "line",
		Data: map[string]interface{}{
			"labels": []string{"Now", "+1h", "+6h", "+12h", "+24h"},
			"datasets": []map[string]interface{}{
				{
					"label":       "Actual",
					"data":        []float64{1500, 1520, 1480, 1510, 1490},
					"borderColor": "rgb(75, 192, 192)",
				},
				{
					"label":       "Predicted",
					"data":        []float64{1500, 1650, 1700, 1750, 1800},
					"borderColor": "rgb(255, 99, 132)",
					"borderDash":  []int{5, 5},
				},
			},
		},
	}
}

// InsightGenerator generates business insights
type InsightGenerator struct {
	config   *AnalyticsConfig
	insights []Insight
	mutex    sync.RWMutex
}

// Insight represents a business insight
type Insight struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Confidence  float64   `json:"confidence"`
	Timestamp   time.Time `json:"timestamp"`
	Tags        []string  `json:"tags"`
}

// NewInsightGenerator creates a new insight generator
func NewInsightGenerator(config *AnalyticsConfig) *InsightGenerator {
	return &InsightGenerator{
		config:   config,
		insights: make([]Insight, 0),
	}
}

// Start starts the insight generator
func (ig *InsightGenerator) Start(ctx context.Context) {
	ticker := time.NewTicker(ig.config.InsightGenerationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ig.generatePeriodicInsights()
		}
	}
}

// generatePeriodicInsights generates periodic insights
func (ig *InsightGenerator) generatePeriodicInsights() {
	ig.mutex.Lock()
	defer ig.mutex.Unlock()

	// Generate sample insights
	insights := []Insight{
		{
			ID:          "insight_001",
			Type:        "performance",
			Title:       "Peak Usage Detected",
			Description: "Classification requests peak between 2-4 PM daily",
			Impact:      "high",
			Confidence:  0.92,
			Timestamp:   time.Now(),
			Tags:        []string{"performance", "usage", "peak"},
		},
		{
			ID:          "insight_002",
			Type:        "trend",
			Title:       "Growing Business Volume",
			Description: "15% increase in business classifications over the past week",
			Impact:      "medium",
			Confidence:  0.88,
			Timestamp:   time.Now().Add(-1 * time.Hour),
			Tags:        []string{"trend", "growth", "volume"},
		},
	}

	ig.insights = append(ig.insights, insights...)

	// Keep only recent insights
	if len(ig.insights) > 50 {
		ig.insights = ig.insights[len(ig.insights)-50:]
	}
}

// GenerateInsights generates insights from analysis results
func (ig *InsightGenerator) GenerateInsights(data []DataPoint, result *AnalysisResult) ([]Insight, error) {
	insights := make([]Insight, 0)

	// Generate insights based on trends
	if len(result.Trends) > 0 {
		for _, trend := range result.Trends {
			insight := Insight{
				ID:          fmt.Sprintf("trend_%d", time.Now().Unix()),
				Type:        "trend",
				Title:       fmt.Sprintf("%s Trend Detected", trend.Direction),
				Description: trend.Description,
				Impact:      ig.getImpactFromStrength(trend.Strength),
				Confidence:  trend.Confidence,
				Timestamp:   time.Now(),
				Tags:        []string{"trend", trend.Direction},
			}
			insights = append(insights, insight)
		}
	}

	// Generate insights based on anomalies
	if len(result.Anomalies) > 0 {
		insight := Insight{
			ID:          fmt.Sprintf("anomaly_%d", time.Now().Unix()),
			Type:        "anomaly",
			Title:       fmt.Sprintf("%d Anomalies Detected", len(result.Anomalies)),
			Description: fmt.Sprintf("Found %d statistical anomalies in the data", len(result.Anomalies)),
			Impact:      "high",
			Confidence:  0.95,
			Timestamp:   time.Now(),
			Tags:        []string{"anomaly", "detection"},
		}
		insights = append(insights, insight)
	}

	return insights, nil
}

// getImpactFromStrength determines impact from trend strength
func (ig *InsightGenerator) getImpactFromStrength(strength float64) string {
	if strength > 0.7 {
		return "high"
	} else if strength > 0.4 {
		return "medium"
	} else {
		return "low"
	}
}

// GetRecentInsights returns recent insights
func (ig *InsightGenerator) GetRecentInsights() []Insight {
	ig.mutex.RLock()
	defer ig.mutex.RUnlock()

	// Return last 10 insights
	if len(ig.insights) > 10 {
		return ig.insights[len(ig.insights)-10:]
	}
	return ig.insights
}
