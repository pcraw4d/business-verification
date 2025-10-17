package ml

import (
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/service"
)

// NewService creates a new ML service instance
func NewService(logger *zap.Logger) *service.MLService {
	return service.NewMLService(logger)
}
