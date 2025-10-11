package models

import (
	"time"
)

// RiskAssessment represents a risk assessment request and response
type RiskAssessment struct {
	ID                string                 `json:"id" db:"id"`
	BusinessID        string                 `json:"business_id" db:"business_id"`
	BusinessName      string                 `json:"business_name" db:"business_name"`
	BusinessAddress   string                 `json:"business_address" db:"business_address"`
	Industry          string                 `json:"industry" db:"industry"`
	Country           string                 `json:"country" db:"country"`
	RiskScore         float64                `json:"risk_score" db:"risk_score"`
	RiskLevel         RiskLevel              `json:"risk_level" db:"risk_level"`
	RiskFactors       []RiskFactor           `json:"risk_factors" db:"risk_factors"`
	PredictionHorizon int                    `json:"prediction_horizon" db:"prediction_horizon"` // months
	ConfidenceScore   float64                `json:"confidence_score" db:"confidence_score"`
	Status            AssessmentStatus       `json:"status" db:"status"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
}

// RiskLevel represents the risk level classification
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// AssessmentStatus represents the status of a risk assessment
type AssessmentStatus string

const (
	StatusPending   AssessmentStatus = "pending"
	StatusCompleted AssessmentStatus = "completed"
	StatusFailed    AssessmentStatus = "failed"
	StatusError     AssessmentStatus = "error"
)

// RiskFactor represents an individual risk factor
type RiskFactor struct {
	Category    RiskCategory `json:"category"`
	Name        string       `json:"name"`
	Score       float64      `json:"score"`
	Weight      float64      `json:"weight"`
	Description string       `json:"description"`
	Source      string       `json:"source"`
	Confidence  float64      `json:"confidence"`
}

// RiskCategory represents the category of risk
type RiskCategory string

const (
	RiskCategoryFinancial     RiskCategory = "financial"
	RiskCategoryOperational   RiskCategory = "operational"
	RiskCategoryCompliance    RiskCategory = "compliance"
	RiskCategoryReputational  RiskCategory = "reputational"
	RiskCategoryRegulatory    RiskCategory = "regulatory"
	RiskCategoryGeopolitical  RiskCategory = "geopolitical"
	RiskCategoryTechnology    RiskCategory = "technology"
	RiskCategoryEnvironmental RiskCategory = "environmental"
)

// RiskAssessmentRequest represents a request for risk assessment
type RiskAssessmentRequest struct {
	BusinessName            string                 `json:"business_name" validate:"required,min=1,max=255"`
	BusinessAddress         string                 `json:"business_address" validate:"required,min=10,max=500"`
	Industry                string                 `json:"industry" validate:"required,min=1,max=100"`
	Country                 string                 `json:"country" validate:"required,len=2"`
	Phone                   string                 `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
	Email                   string                 `json:"email,omitempty" validate:"omitempty,email"`
	Website                 string                 `json:"website,omitempty" validate:"omitempty,url"`
	PredictionHorizon       int                    `json:"prediction_horizon,omitempty" validate:"omitempty,min=1,max=24"`
	ModelType               string                 `json:"model_type,omitempty" validate:"omitempty,oneof=auto xgboost lstm ensemble"`
	IncludeTemporalAnalysis bool                   `json:"include_temporal_analysis,omitempty"`
	Metadata                map[string]interface{} `json:"metadata,omitempty"`
}

// RiskAssessmentResponse represents the response from a risk assessment
type RiskAssessmentResponse struct {
	ID                string                 `json:"id"`
	BusinessID        string                 `json:"business_id"`
	RiskScore         float64                `json:"risk_score"`
	RiskLevel         RiskLevel              `json:"risk_level"`
	RiskFactors       []RiskFactor           `json:"risk_factors"`
	PredictionHorizon int                    `json:"prediction_horizon"`
	ConfidenceScore   float64                `json:"confidence_score"`
	Status            AssessmentStatus       `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// RiskPrediction represents a risk prediction for future time periods
type RiskPrediction struct {
	BusinessID       string             `json:"business_id"`
	PredictionDate   time.Time          `json:"prediction_date"`
	HorizonMonths    int                `json:"horizon_months"`
	PredictedScore   float64            `json:"predicted_score"`
	PredictedLevel   RiskLevel          `json:"predicted_level"`
	ConfidenceScore  float64            `json:"confidence_score"`
	RiskFactors      []RiskFactor       `json:"risk_factors"`
	ScenarioAnalysis []ScenarioAnalysis `json:"scenario_analysis,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
}

// ScenarioAnalysis represents different risk scenarios
type ScenarioAnalysis struct {
	ScenarioName string    `json:"scenario_name"`
	Description  string    `json:"description"`
	RiskScore    float64   `json:"risk_score"`
	RiskLevel    RiskLevel `json:"risk_level"`
	Probability  float64   `json:"probability"`
	Impact       string    `json:"impact"`
}

// ExternalData represents data from external sources
type ExternalData struct {
	Source      string                 `json:"source"`
	SourceType  string                 `json:"source_type"`
	Data        map[string]interface{} `json:"data"`
	Confidence  float64                `json:"confidence"`
	LastUpdated time.Time              `json:"last_updated"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// ComplianceCheck represents a compliance check result
type ComplianceCheck struct {
	CheckType   string                 `json:"check_type"`
	Status      string                 `json:"status"`
	Description string                 `json:"description"`
	RiskLevel   RiskLevel              `json:"risk_level"`
	Details     map[string]interface{} `json:"details"`
	CheckedAt   time.Time              `json:"checked_at"`
	NextCheckAt *time.Time             `json:"next_check_at,omitempty"`
}

// SanctionsCheck represents a sanctions screening result
type SanctionsCheck struct {
	EntityName    string                 `json:"entity_name"`
	EntityType    string                 `json:"entity_type"`
	MatchType     string                 `json:"match_type"`
	MatchScore    float64                `json:"match_score"`
	SanctionsList string                 `json:"sanctions_list"`
	Details       map[string]interface{} `json:"details"`
	CheckedAt     time.Time              `json:"checked_at"`
}

// AdverseMedia represents adverse media monitoring results
type AdverseMedia struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	Sentiment   string    `json:"sentiment"`
	RiskLevel   RiskLevel `json:"risk_level"`
	Summary     string    `json:"summary"`
	Keywords    []string  `json:"keywords"`
}
