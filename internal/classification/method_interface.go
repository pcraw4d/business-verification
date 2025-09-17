package classification

import (
	"github.com/pcraw4d/business-verification/internal/classification/methods"
)

// ClassificationMethod is an alias for the methods.ClassificationMethod interface
type ClassificationMethod = methods.ClassificationMethod

// MethodConfig is an alias for the methods.MethodConfig type
type MethodConfig = methods.MethodConfig

// MethodPerformanceMetrics is an alias for the methods.MethodPerformanceMetrics type
type MethodPerformanceMetrics = methods.MethodPerformanceMetrics

// MethodRegistration represents a registered classification method
type MethodRegistration struct {
	Method  ClassificationMethod
	Config  MethodConfig
	Metrics *MethodPerformanceMetrics
}

// NewMethodPerformanceMetrics creates a new MethodPerformanceMetrics instance
func NewMethodPerformanceMetrics() *MethodPerformanceMetrics {
	return methods.NewMethodPerformanceMetrics()
}
