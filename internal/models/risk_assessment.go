package models

import (
	"encoding/json"
	"time"
)

// RiskAssessment represents an async risk assessment record
type RiskAssessment struct {
	ID              string                 `json:"id" db:"id"`
	MerchantID      string                 `json:"merchantId" db:"merchant_id"`
	Status          AssessmentStatus       `json:"status" db:"status"`
	Options         AssessmentOptions      `json:"options" db:"options"`
	Result          *RiskAssessmentResult  `json:"result,omitempty" db:"result"`
	Progress        int                    `json:"progress" db:"progress"`
	EstimatedCompletion *time.Time         `json:"estimatedCompletion,omitempty" db:"estimated_completion"`
	CreatedAt       time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time              `json:"updatedAt" db:"updated_at"`
	CompletedAt     *time.Time             `json:"completedAt,omitempty" db:"completed_at"`
}

// AssessmentStatus represents the status of a risk assessment
type AssessmentStatus string

const (
	AssessmentStatusPending    AssessmentStatus = "pending"
	AssessmentStatusProcessing AssessmentStatus = "processing"
	AssessmentStatusCompleted  AssessmentStatus = "completed"
	AssessmentStatusFailed     AssessmentStatus = "failed"
)

// AssessmentOptions represents options for a risk assessment
type AssessmentOptions struct {
	IncludeHistory    bool `json:"includeHistory"`
	IncludePredictions bool `json:"includePredictions"`
}

// RiskAssessmentResult represents the result of a completed risk assessment
type RiskAssessmentResult struct {
	OverallScore float64      `json:"overallScore"`
	RiskLevel    string       `json:"riskLevel"`
	Factors      []RiskFactor `json:"factors"`
}

// RiskFactor represents a risk factor in the assessment
type RiskFactor struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
}

// RiskAssessmentRequest represents a request to start a risk assessment
type RiskAssessmentRequest struct {
	MerchantID string             `json:"merchantId"`
	Options    AssessmentOptions   `json:"options"`
}

// RiskAssessmentResponse represents the response when starting an assessment
type RiskAssessmentResponse struct {
	AssessmentID         string     `json:"assessmentId"`
	Status               string     `json:"status"`
	EstimatedCompletion  *time.Time `json:"estimatedCompletion,omitempty"`
}

// AssessmentStatusResponse represents the status of an assessment
type AssessmentStatusResponse struct {
	AssessmentID         string                 `json:"assessmentId"`
	MerchantID           string                 `json:"merchantId"`
	Status               string                 `json:"status"`
	Progress             int                    `json:"progress"`
	EstimatedCompletion  *time.Time             `json:"estimatedCompletion,omitempty"`
	Result               *RiskAssessmentResult  `json:"result,omitempty"`
	CompletedAt          *time.Time              `json:"completedAt,omitempty"`
}

// Scan implements the sql.Scanner interface for AssessmentOptions
func (ao *AssessmentOptions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, ao)
}

// Value implements the driver.Valuer interface for AssessmentOptions
func (ao AssessmentOptions) Value() (interface{}, error) {
	return json.Marshal(ao)
}

// Scan implements the sql.Scanner interface for RiskAssessmentResult
func (rar *RiskAssessmentResult) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, rar)
}

// Value implements the driver.Valuer interface for RiskAssessmentResult
func (rar RiskAssessmentResult) Value() (interface{}, error) {
	return json.Marshal(rar)
}

