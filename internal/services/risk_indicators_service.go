package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
)

// RiskIndicatorsService provides business logic for risk indicators
type RiskIndicatorsService interface {
	GetRiskIndicators(ctx context.Context, merchantID string, filters *database.RiskIndicatorFilters) (*models.RiskIndicatorsData, error)
}

// riskIndicatorsService implements RiskIndicatorsService
type riskIndicatorsService struct {
	indicatorsRepo *database.RiskIndicatorsRepository
	logger         *log.Logger
}

// NewRiskIndicatorsService creates a new risk indicators service
func NewRiskIndicatorsService(
	indicatorsRepo *database.RiskIndicatorsRepository,
	logger *log.Logger,
) RiskIndicatorsService {
	if logger == nil {
		logger = log.Default()
	}

	return &riskIndicatorsService{
		indicatorsRepo: indicatorsRepo,
		logger:         logger,
	}
}

// GetRiskIndicators retrieves risk indicators for a merchant
func (s *riskIndicatorsService) GetRiskIndicators(ctx context.Context, merchantID string, filters *database.RiskIndicatorFilters) (*models.RiskIndicatorsData, error) {
	s.logger.Printf("Getting risk indicators for merchant: %s", merchantID)

	// Get indicators from repository
	indicators, err := s.indicatorsRepo.GetByMerchantID(ctx, merchantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get indicators: %w", err)
	}

	// Calculate overall score
	overallScore := s.calculateOverallScore(indicators)

	// Assemble response
	data := &models.RiskIndicatorsData{
		MerchantID:   merchantID,
		OverallScore: overallScore,
		Indicators:   indicators,
		LastUpdated:  time.Now(),
	}

	return data, nil
}

// calculateOverallScore calculates the overall risk score from indicators
func (s *riskIndicatorsService) calculateOverallScore(indicators []models.RiskIndicator) float64 {
	if len(indicators) == 0 {
		return 0.0
	}

	var totalScore float64
	var totalWeight float64

	for _, indicator := range indicators {
		// Weight by severity
		weight := 1.0
		switch indicator.Severity {
		case "critical":
			weight = 4.0
		case "high":
			weight = 3.0
		case "medium":
			weight = 2.0
		case "low":
			weight = 1.0
		}

		totalScore += indicator.Score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

