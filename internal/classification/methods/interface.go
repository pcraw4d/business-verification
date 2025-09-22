package methods

import (
	"context"
	"time"

	"kyb-platform/internal/shared"
)

// ClassificationMethod defines the interface that all classification methods must implement
type ClassificationMethod interface {
	// GetName returns the unique name of the classification method
	GetName() string

	// GetType returns the type/category of the method (e.g., "keyword", "ml", "external_api")
	GetType() string

	// GetDescription returns a human-readable description of what this method does
	GetDescription() string

	// GetWeight returns the current weight/importance of this method in the ensemble
	GetWeight() float64

	// SetWeight sets the weight/importance of this method in the ensemble
	SetWeight(weight float64)

	// IsEnabled returns whether this method is currently enabled
	IsEnabled() bool

	// SetEnabled enables or disables this method
	SetEnabled(enabled bool)

	// Classify performs the actual classification using this method
	Classify(ctx context.Context, businessName, description, websiteURL string) (*shared.ClassificationMethodResult, error)

	// GetPerformanceMetrics returns performance metrics for this method
	GetPerformanceMetrics() interface{}

	// ValidateInput validates the input parameters before classification
	ValidateInput(businessName, description, websiteURL string) error

	// GetRequiredDependencies returns a list of dependencies this method requires
	GetRequiredDependencies() []string

	// Initialize performs any necessary initialization for this method
	Initialize(ctx context.Context) error

	// Cleanup performs any necessary cleanup when the method is removed
	Cleanup() error
}

// MethodConfig represents configuration for a classification method
type MethodConfig struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Weight      float64                `json:"weight"`
	Enabled     bool                   `json:"enabled"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// MethodPerformanceMetrics tracks performance metrics for a classification method
type MethodPerformanceMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastResponseTime    time.Duration `json:"last_response_time"`
	AccuracyScore       float64       `json:"accuracy_score"`
	LastAccuracyUpdate  time.Time     `json:"last_accuracy_update"`
	ErrorRate           float64       `json:"error_rate"`
	LastError           string        `json:"last_error,omitempty"`
	LastErrorTime       *time.Time    `json:"last_error_time,omitempty"`
}

// NewMethodPerformanceMetrics creates a new MethodPerformanceMetrics instance
func NewMethodPerformanceMetrics() *MethodPerformanceMetrics {
	return &MethodPerformanceMetrics{
		TotalRequests:       0,
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		LastResponseTime:    0,
		AccuracyScore:       0.0,
		LastAccuracyUpdate:  time.Now(),
		ErrorRate:           0.0,
	}
}

// UpdateMetrics updates the performance metrics for a method
func (m *MethodPerformanceMetrics) UpdateMetrics(success bool, responseTime time.Duration, err error) {
	m.TotalRequests++
	m.LastResponseTime = responseTime

	if success {
		m.SuccessfulRequests++
	} else {
		m.FailedRequests++
		if err != nil {
			now := time.Now()
			m.LastError = err.Error()
			m.LastErrorTime = &now
		}
	}

	// Update average response time
	if m.TotalRequests == 1 {
		m.AverageResponseTime = responseTime
	} else {
		// Calculate running average
		m.AverageResponseTime = time.Duration(
			(int64(m.AverageResponseTime)*int64(m.TotalRequests-1) + int64(responseTime)) / int64(m.TotalRequests),
		)
	}

	// Update error rate
	m.ErrorRate = float64(m.FailedRequests) / float64(m.TotalRequests)
}

// UpdateAccuracy updates the accuracy score for the method
func (m *MethodPerformanceMetrics) UpdateAccuracy(accuracy float64) {
	m.AccuracyScore = accuracy
	m.LastAccuracyUpdate = time.Now()
}
