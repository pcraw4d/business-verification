package main

import (
	"kyb-platform/services/risk-assessment-service/internal/api/handlers"
	"kyb-platform/services/risk-assessment-service/internal/ml/service"

	"go.uber.org/zap"
)

func main() {
	logger := zap.NewNop()
	mlService := &service.MLService{}
	_ = handlers.NewExplainabilityHandlers(mlService, logger)
}
