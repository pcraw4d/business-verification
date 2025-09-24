package classification

import (
	"time"
)

// Industry represents an industry in the classification system
type Industry struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Category            string    `json:"category"`
	ConfidenceThreshold float64   `json:"confidence_threshold"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ClassificationCode represents a classification code (MCC, NAICS, SIC)
type ClassificationCode struct {
	ID          int       `json:"id"`
	IndustryID  int       `json:"industry_id"`
	CodeType    string    `json:"code_type"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Confidence  float64   `json:"confidence"`
	IsPrimary   bool      `json:"is_primary"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IndustryKeyword represents a keyword associated with an industry
type IndustryKeyword struct {
	ID         int       `json:"id"`
	IndustryID int       `json:"industry_id"`
	Keyword    string    `json:"keyword"`
	Weight     float64   `json:"weight"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Note: ClassificationCodesInfo, MCCCode, SICCode, and NAICSCode types
// are already defined in other files in the classification package
