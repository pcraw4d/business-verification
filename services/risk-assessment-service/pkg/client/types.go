package client

import (
	"time"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// RequestOptions provides additional options for API requests
type RequestOptions struct {
	Headers map[string]string
	Timeout time.Duration
}

// RiskPredictionRequest represents a risk prediction request
type RiskPredictionRequest struct {
	HorizonMonths int      `json:"horizon_months"`
	Scenarios     []string `json:"scenarios,omitempty"`
}

// RiskHistoryResponse represents a risk history response
type RiskHistoryResponse struct {
	BusinessID  string                          `json:"business_id"`
	Assessments []models.RiskAssessmentResponse `json:"assessments"`
	Trends      RiskTrends                      `json:"trends"`
}

// RiskTrends represents risk trends
type RiskTrends struct {
	ScoreTrend      string `json:"score_trend"`
	LevelTrend      string `json:"level_trend"`
	ConfidenceTrend string `json:"confidence_trend"`
}

// ComplianceCheckRequest represents a compliance check request
type ComplianceCheckRequest struct {
	BusinessName    string   `json:"business_name"`
	BusinessAddress string   `json:"business_address"`
	Industry        string   `json:"industry"`
	Country         string   `json:"country"`
	ComplianceTypes []string `json:"compliance_types"`
}

// ComplianceCheckResponse represents a compliance check response
type ComplianceCheckResponse struct {
	BusinessID       string            `json:"business_id"`
	ComplianceStatus string            `json:"compliance_status"`
	Checks           []ComplianceCheck `json:"checks"`
	CreatedAt        time.Time         `json:"created_at"`
}

// ComplianceCheck represents a compliance check result
type ComplianceCheck struct {
	Type    string  `json:"type"`
	Status  string  `json:"status"`
	Score   float64 `json:"score"`
	Details string  `json:"details"`
}

// SanctionsScreeningRequest represents a sanctions screening request
type SanctionsScreeningRequest struct {
	BusinessName    string `json:"business_name"`
	BusinessAddress string `json:"business_address"`
	Country         string `json:"country"`
}

// SanctionsScreeningResponse represents a sanctions screening response
type SanctionsScreeningResponse struct {
	BusinessID      string           `json:"business_id"`
	SanctionsStatus string           `json:"sanctions_status"`
	Matches         []SanctionsMatch `json:"matches"`
	ScreeningDate   time.Time        `json:"screening_date"`
}

// SanctionsMatch represents a sanctions match
type SanctionsMatch struct {
	Source    string    `json:"source"`
	Score     float64   `json:"score"`
	Details   string    `json:"details"`
	MatchDate time.Time `json:"match_date"`
}

// MediaMonitoringRequest represents a media monitoring request
type MediaMonitoringRequest struct {
	BusinessName    string   `json:"business_name"`
	BusinessAddress string   `json:"business_address"`
	MonitoringTypes []string `json:"monitoring_types"`
}

// MediaMonitoringResponse represents a media monitoring response
type MediaMonitoringResponse struct {
	BusinessID   string       `json:"business_id"`
	MonitoringID string       `json:"monitoring_id"`
	Status       string       `json:"status"`
	Alerts       []MediaAlert `json:"alerts"`
	CreatedAt    time.Time    `json:"created_at"`
}

// MediaAlert represents a media alert
type MediaAlert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	Severity    string    `json:"severity"`
	CreatedAt   time.Time `json:"created_at"`
}

// RiskTrendsOptions provides options for risk trends requests
type RiskTrendsOptions struct {
	Industry  string
	Country   string
	Timeframe string
	Limit     int
}

// RiskTrendsResponse represents a risk trends response
type RiskTrendsResponse struct {
	Trends  []RiskTrend  `json:"trends"`
	Summary TrendSummary `json:"summary"`
}

// RiskTrend represents a risk trend
type RiskTrend struct {
	Industry         string  `json:"industry"`
	Country          string  `json:"country"`
	AverageRiskScore float64 `json:"average_risk_score"`
	TrendDirection   string  `json:"trend_direction"`
	ChangePercentage float64 `json:"change_percentage"`
	SampleSize       int     `json:"sample_size"`
}

// TrendSummary represents a trend summary
type TrendSummary struct {
	TotalAssessments   int     `json:"total_assessments"`
	AverageRiskScore   float64 `json:"average_risk_score"`
	HighRiskPercentage float64 `json:"high_risk_percentage"`
}

// RiskInsightsOptions provides options for risk insights requests
type RiskInsightsOptions struct {
	Industry  string
	Country   string
	RiskLevel string
}

// RiskInsightsResponse represents a risk insights response
type RiskInsightsResponse struct {
	Insights        []RiskInsight    `json:"insights"`
	Recommendations []Recommendation `json:"recommendations"`
}

// RiskInsight represents a risk insight
type RiskInsight struct {
	Type           string `json:"type"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Impact         string `json:"impact"`
	Recommendation string `json:"recommendation"`
}

// Recommendation represents a recommendation
type Recommendation struct {
	Category string `json:"category"`
	Action   string `json:"action"`
	Priority string `json:"priority"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
	Method    string      `json:"method"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	Field      string            `json:"field,omitempty"`
	Validation []ValidationError `json:"validation,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// APIError represents an API error
type APIError struct {
	StatusCode int
	Message    string
	Response   *ErrorResponse
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Response != nil {
		return e.Response.Error.Message
	}
	return e.Message
}

// IsValidationError checks if the error is a validation error
func (e *APIError) IsValidationError() bool {
	return e.Response != nil && e.Response.Error.Code == "VALIDATION_ERROR"
}

// IsAuthenticationError checks if the error is an authentication error
func (e *APIError) IsAuthenticationError() bool {
	return e.Response != nil && e.Response.Error.Code == "AUTHENTICATION_ERROR"
}

// IsAuthorizationError checks if the error is an authorization error
func (e *APIError) IsAuthorizationError() bool {
	return e.Response != nil && e.Response.Error.Code == "AUTHORIZATION_ERROR"
}

// IsRateLimitError checks if the error is a rate limit error
func (e *APIError) IsRateLimitError() bool {
	return e.Response != nil && e.Response.Error.Code == "RATE_LIMIT_EXCEEDED"
}

// IsNotFoundError checks if the error is a not found error
func (e *APIError) IsNotFoundError() bool {
	return e.Response != nil && e.Response.Error.Code == "NOT_FOUND"
}

// IsServiceUnavailableError checks if the error is a service unavailable error
func (e *APIError) IsServiceUnavailableError() bool {
	return e.Response != nil && e.Response.Error.Code == "SERVICE_UNAVAILABLE"
}

// IsTimeoutError checks if the error is a timeout error
func (e *APIError) IsTimeoutError() bool {
	return e.Response != nil && e.Response.Error.Code == "REQUEST_TIMEOUT"
}

// IsInternalError checks if the error is an internal error
func (e *APIError) IsInternalError() bool {
	return e.Response != nil && e.Response.Error.Code == "INTERNAL_ERROR"
}

// GetValidationErrors returns validation errors if any
func (e *APIError) GetValidationErrors() []ValidationError {
	if e.Response != nil {
		return e.Response.Error.Validation
	}
	return nil
}

// GetRequestID returns the request ID if available
func (e *APIError) GetRequestID() string {
	if e.Response != nil {
		return e.Response.RequestID
	}
	return ""
}

// GetTimestamp returns the error timestamp if available
func (e *APIError) GetTimestamp() string {
	if e.Response != nil {
		return e.Response.Timestamp
	}
	return ""
}

// GetPath returns the request path if available
func (e *APIError) GetPath() string {
	if e.Response != nil {
		return e.Response.Path
	}
	return ""
}

// GetMethod returns the request method if available
func (e *APIError) GetMethod() string {
	if e.Response != nil {
		return e.Response.Method
	}
	return ""
}
