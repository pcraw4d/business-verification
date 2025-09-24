package risk

import (
	"time"
)

// EnhancedRiskAssessmentRequest represents a comprehensive risk assessment request
type EnhancedRiskAssessmentRequest struct {
	AssessmentID               string                 `json:"assessment_id"`
	BusinessID                 string                 `json:"business_id"`
	RiskFactorInputs           []RiskFactorInput      `json:"risk_factor_inputs"`
	IncludeTrendAnalysis       bool                   `json:"include_trend_analysis"`
	IncludeCorrelationAnalysis bool                   `json:"include_correlation_analysis"`
	TimeRange                  *TimeRange             `json:"time_range,omitempty"`
	CustomWeights              map[string]float64     `json:"custom_weights,omitempty"`
	Metadata                   map[string]interface{} `json:"metadata,omitempty"`
}

// EnhancedRiskAssessmentResponse represents the response from enhanced risk assessment
type EnhancedRiskAssessmentResponse struct {
	AssessmentID     string                 `json:"assessment_id"`
	BusinessID       string                 `json:"business_id"`
	Timestamp        time.Time              `json:"timestamp"`
	OverallRiskScore float64                `json:"overall_risk_score"`
	OverallRiskLevel RiskLevel              `json:"overall_risk_level"`
	RiskFactors      []RiskFactorDetail     `json:"risk_factors"`
	Recommendations  []RecommendationDetail `json:"recommendations"`
	TrendData        *RiskTrendData         `json:"trend_data,omitempty"`
	CorrelationData  map[string]float64     `json:"correlation_data,omitempty"`
	Alerts           []AlertDetail          `json:"alerts"`
	ConfidenceScore  float64                `json:"confidence_score"`
	ProcessingTimeMs int64                  `json:"processing_time_ms"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RiskFactorDetail represents detailed information about a risk factor
type RiskFactorDetail struct {
	FactorType          string                 `json:"factor_type"`
	Score               float64                `json:"score"`
	RiskLevel           RiskLevel              `json:"risk_level"`
	Confidence          float64                `json:"confidence"`
	Weight              float64                `json:"weight"`
	Description         string                 `json:"description"`
	ContributingFactors []string               `json:"contributing_factors"`
	LastUpdated         time.Time              `json:"last_updated"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// RiskTrendData represents trend analysis data
// Note: This type is already defined in trend_analysis_service.go
// type RiskTrendData struct {
//	BusinessID   string      `json:"business_id"`
//	TimeRange    *TimeRange  `json:"time_range"`
//	Trends       []RiskTrend `json:"trends"`
//	LastAnalyzed time.Time   `json:"last_analyzed"`
//	DataPoints   int         `json:"data_points"`
//	TrendSummary string      `json:"trend_summary"`
// }

// RiskTrend represents a single risk trend
// Note: This type is already defined in models.go
// type RiskTrend struct {
//	FactorType  string           `json:"factor_type"`
//	Direction   TrendDirection   `json:"direction"`
//	Magnitude   float64          `json:"magnitude"`
//	Confidence  float64          `json:"confidence"`
//	Timeframe   string           `json:"timeframe"`
//	Description string           `json:"description"`
//	DataPoints  []TrendDataPoint `json:"data_points"`
// }

// TrendDirection represents the direction of a trend
type TrendDirection string

const (
	TrendDirectionImproving     TrendDirection = "improving"
	TrendDirectionDeteriorating TrendDirection = "deteriorating"
	TrendDirectionStable        TrendDirection = "stable"
	TrendDirectionUnknown       TrendDirection = "unknown"
)

// TrendDataPoint represents a single data point in a trend
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Score     float64   `json:"score"`
}

// RecommendationDetail represents detailed recommendation information
type RecommendationDetail struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Priority        PriorityLevel          `json:"priority"`
	Category        string                 `json:"category"`
	Impact          string                 `json:"impact"`
	Effort          string                 `json:"effort"`
	Timeline        string                 `json:"timeline"`
	Cost            *CostEstimate          `json:"cost,omitempty"`
	Resources       []string               `json:"resources"`
	Prerequisites   []string               `json:"prerequisites"`
	ExpectedOutcome string                 `json:"expected_outcome"`
	SuccessMetrics  []string               `json:"success_metrics"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PriorityLevel represents the priority level of a recommendation
type PriorityLevel string

const (
	PriorityCritical PriorityLevel = "critical"
	PriorityHigh     PriorityLevel = "high"
	PriorityMedium   PriorityLevel = "medium"
	PriorityLow      PriorityLevel = "low"
)

// CostEstimate represents cost estimation for a recommendation
type CostEstimate struct {
	MinCost    float64 `json:"min_cost"`
	MaxCost    float64 `json:"max_cost"`
	Currency   string  `json:"currency"`
	Timeframe  string  `json:"timeframe"`
	Confidence float64 `json:"confidence"`
}

// AlertDetail represents detailed alert information
type AlertDetail struct {
	ID             string                 `json:"id"`
	BusinessID     string                 `json:"business_id"`
	AlertType      string                 `json:"alert_type"`
	Severity       AlertSeverity          `json:"severity"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	RiskFactor     string                 `json:"risk_factor"`
	Threshold      float64                `json:"threshold"`
	CurrentValue   float64                `json:"current_value"`
	Status         AlertStatus            `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time             `json:"resolved_at,omitempty"`
	AcknowledgedBy *string                `json:"acknowledged_by,omitempty"`
	ResolvedBy     *string                `json:"resolved_by,omitempty"`
	Notes          string                 `json:"notes"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Note: AlertSeverity and AlertStatus are defined in alert_system.go

// RiskHistoryEntry represents a historical risk assessment entry
// Note: This type is already defined in automated_alerts_stub.go
// type RiskHistoryEntry struct {
//	ID           string                 `json:"id"`
//	BusinessID   string                 `json:"business_id"`
//	AssessmentID string                 `json:"assessment_id"`
//	Timestamp    time.Time              `json:"timestamp"`
//	RiskScore    float64                `json:"risk_score"`
//	RiskLevel    RiskLevel              `json:"risk_level"`
//	FactorScores map[string]float64     `json:"factor_scores"`
//	Confidence   float64                `json:"confidence"`
//	Metadata     map[string]interface{} `json:"metadata,omitempty"`
// }

// TimeRange represents a time range for analysis
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  string    `json:"duration"`
}

// RiskFactorInput represents input data for risk factor calculation
// Note: This type is already defined in calculation.go
// type RiskFactorInput struct {
//	FactorType string                 `json:"factor_type"`
//	Data       map[string]interface{} `json:"data"`
//	Weight     float64                `json:"weight"`
//	Metadata   map[string]interface{} `json:"metadata,omitempty"`
// }

// EnhancedRiskFactorResult represents the result of enhanced risk factor calculation
// Note: This type is already defined in enhanced_calculation.go
// type EnhancedRiskFactorResult struct {
//	FactorType          string                 `json:"factor_type"`
//	Score               float64                `json:"score"`
//	RiskLevel           RiskLevel              `json:"risk_level"`
//	Confidence          float64                `json:"confidence"`
//	Description         string                 `json:"description"`
//	ContributingFactors []string               `json:"contributing_factors"`
//	Metadata            map[string]interface{} `json:"metadata,omitempty"`
// }

// RiskTrendAnalysisResponse represents the response from trend analysis
// Note: This type is already defined in trend_analysis_service.go
// type RiskTrendAnalysisResponse struct {
//	BusinessID   string      `json:"business_id"`
//	TimeRange    *TimeRange  `json:"time_range"`
//	Trends       []RiskTrend `json:"trends"`
//	LastAnalyzed time.Time   `json:"last_analyzed"`
//	DataPoints   int         `json:"data_points"`
//	TrendSummary string      `json:"trend_summary"`
// }

// EnhancedAssessmentSummary contains summary information
type EnhancedAssessmentSummary struct {
	TotalFactors         int       `json:"total_factors"`
	HighRiskFactors      int       `json:"high_risk_factors"`
	CriticalRiskFactors  int       `json:"critical_risk_factors"`
	RecommendationsCount int       `json:"recommendations_count"`
	ActiveAlertsCount    int       `json:"active_alerts_count"`
	DataQualityScore     float64   `json:"data_quality_score"`
	AssessmentConfidence float64   `json:"assessment_confidence"`
	LastUpdated          time.Time `json:"last_updated"`
}
