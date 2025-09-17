package integrations

import (
	"context"
)

// FreeDataValidationServiceInterface defines the interface for free data validation services
type FreeDataValidationServiceInterface interface {
	// ValidateBusinessData validates business data using free government APIs
	ValidateBusinessData(ctx context.Context, data BusinessDataForValidation) (*ValidationResult, error)

	// GetValidationStats returns statistics about validation performance
	GetValidationStats() map[string]interface{}
}

// Ensure FreeDataValidationService implements the interface
var _ FreeDataValidationServiceInterface = (*FreeDataValidationService)(nil)
