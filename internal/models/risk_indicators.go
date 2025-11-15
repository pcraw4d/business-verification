package models

import "time"

// RiskIndicatorsData represents risk indicators data for a merchant
type RiskIndicatorsData struct {
	MerchantID    string          `json:"merchantId"`
	OverallScore  float64         `json:"overallScore"`
	Indicators    []RiskIndicator `json:"indicators"`
	LastUpdated   time.Time       `json:"lastUpdated"`
}

// RiskIndicator represents a single risk indicator
type RiskIndicator struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Severity    string    `json:"severity"` // low, medium, high, critical
	Status      string    `json:"status"`   // active, resolved, dismissed
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detectedAt"`
	Score       float64   `json:"score"`
}

