package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/jobs"
	"kyb-platform/internal/models"
)

// RiskAssessmentService provides business logic for risk assessments
type RiskAssessmentService interface {
	StartAssessment(ctx context.Context, merchantID string, options models.AssessmentOptions) (string, error)
	GetAssessmentStatus(ctx context.Context, assessmentID string) (*models.AssessmentStatusResponse, error)
	ProcessAssessment(ctx context.Context, assessmentID string) error
	GetRiskHistory(ctx context.Context, merchantID string, limit, offset int) ([]*models.RiskAssessment, error)
	GetPredictions(ctx context.Context, merchantID string, horizons []int, includeScenarios, includeConfidence bool) (map[string]interface{}, error)
	ExplainAssessment(ctx context.Context, assessmentID string) (map[string]interface{}, error)
	GetRecommendations(ctx context.Context, merchantID string) ([]map[string]interface{}, error)
}

// riskAssessmentService implements RiskAssessmentService
type riskAssessmentService struct {
	repo     *database.RiskAssessmentRepository
	jobQueue *jobs.RiskAssessmentJobProcessor
	logger   *log.Logger
}

// NewRiskAssessmentService creates a new risk assessment service
func NewRiskAssessmentService(
	repo *database.RiskAssessmentRepository,
	jobQueue *jobs.RiskAssessmentJobProcessor,
	logger *log.Logger,
) RiskAssessmentService {
	if logger == nil {
		logger = log.Default()
	}

	return &riskAssessmentService{
		repo:     repo,
		jobQueue: jobQueue,
		logger:   logger,
	}
}

// StartAssessment starts a new risk assessment and returns the assessment ID
func (s *riskAssessmentService) StartAssessment(ctx context.Context, merchantID string, options models.AssessmentOptions) (string, error) {
	s.logger.Printf("Starting risk assessment for merchant: %s", merchantID)

	// Generate assessment ID
	assessmentID := fmt.Sprintf("assess-%d", time.Now().UnixNano())

	// Calculate estimated completion time (5 minutes from now)
	estimatedCompletion := time.Now().Add(5 * time.Minute)

	// Create assessment record
	assessment := &models.RiskAssessment{
		ID:                  assessmentID,
		MerchantID:          merchantID,
		Status:              models.AssessmentStatusPending,
		Options:             options,
		Progress:            0,
		EstimatedCompletion: &estimatedCompletion,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Save to database
	err := s.repo.CreateAssessment(ctx, assessment)
	if err != nil {
		return "", fmt.Errorf("failed to create assessment: %w", err)
	}

	// Enqueue job for processing
	job := jobs.RiskAssessmentJob{
		AssessmentID: assessmentID,
		MerchantID:   merchantID,
		Options:      options,
	}

	err = s.jobQueue.Enqueue(ctx, job)
	if err != nil {
		// If enqueue fails, update status to failed
		s.repo.UpdateAssessmentStatus(ctx, assessmentID, models.AssessmentStatusFailed, 0)
		return "", fmt.Errorf("failed to enqueue assessment: %w", err)
	}

	s.logger.Printf("Risk assessment started: %s", assessmentID)
	return assessmentID, nil
}

// GetAssessmentStatus retrieves the status of a risk assessment
func (s *riskAssessmentService) GetAssessmentStatus(ctx context.Context, assessmentID string) (*models.AssessmentStatusResponse, error) {
	s.logger.Printf("Getting assessment status: %s", assessmentID)

	assessment, err := s.repo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		return nil, fmt.Errorf("assessment not found: %w", err)
	}

	response := &models.AssessmentStatusResponse{
		AssessmentID:        assessment.ID,
		MerchantID:          assessment.MerchantID,
		Status:              string(assessment.Status),
		Progress:            assessment.Progress,
		EstimatedCompletion: assessment.EstimatedCompletion,
		Result:              assessment.Result,
		CompletedAt:         assessment.CompletedAt,
	}

	return response, nil
}

// ProcessAssessment processes a risk assessment (called by job processor)
func (s *riskAssessmentService) ProcessAssessment(ctx context.Context, assessmentID string) error {
	s.logger.Printf("Processing assessment: %s", assessmentID)

	_, err := s.repo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		return fmt.Errorf("assessment not found: %w", err)
	}

	// This method is typically called by the job processor
	// The actual processing happens in the job processor
	// This is here for interface compatibility

	return nil
}

// GetRiskHistory retrieves risk assessment history for a merchant with pagination
func (s *riskAssessmentService) GetRiskHistory(ctx context.Context, merchantID string, limit, offset int) ([]*models.RiskAssessment, error) {
	s.logger.Printf("Getting risk history for merchant: %s (limit: %d, offset: %d)", merchantID, limit, offset)

	// Get all assessments for merchant
	assessments, err := s.repo.GetAssessmentsByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assessments: %w", err)
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(assessments) {
		return []*models.RiskAssessment{}, nil
	}
	if end > len(assessments) {
		end = len(assessments)
	}

	if start < 0 {
		start = 0
	}

	return assessments[start:end], nil
}

// GetPredictions retrieves risk predictions for a merchant
func (s *riskAssessmentService) GetPredictions(ctx context.Context, merchantID string, horizons []int, includeScenarios, includeConfidence bool) (map[string]interface{}, error) {
	s.logger.Printf("Getting risk predictions for merchant: %s", merchantID)

	// Get historical assessments to base predictions on
	assessments, err := s.repo.GetAssessmentsByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assessments: %w", err)
	}

	// Generate predictions based on historical data
	predictions := []map[string]interface{}{}
	for _, horizon := range horizons {
		prediction := map[string]interface{}{
			"horizon": horizon,
			"months":  horizon,
		}

		// Calculate predicted score based on historical trend
		if len(assessments) > 0 {
			// Use most recent assessment as baseline
			latest := assessments[0]
			if latest.Result != nil {
				baseScore := latest.Result.OverallScore
				// Simple trend: assume slight increase over time
				predictedScore := baseScore + (float64(horizon) * 0.01)
				if predictedScore > 1.0 {
					predictedScore = 1.0
				}
				prediction["predictedScore"] = predictedScore
				prediction["riskLevel"] = s.calculateRiskLevel(predictedScore)
			}
		}

		if includeConfidence {
			// Calculate confidence based on data availability
			confidence := 0.7
			if len(assessments) > 3 {
				confidence = 0.9
			} else if len(assessments) > 1 {
				confidence = 0.8
			}
			prediction["confidence"] = confidence
		}

		if includeScenarios {
			predictedScore, _ := prediction["predictedScore"].(float64)
			prediction["scenarios"] = []map[string]interface{}{
				{
					"name":        "baseline",
					"probability": 0.6,
					"score":       predictedScore,
				},
				{
					"name":        "optimistic",
					"probability": 0.2,
					"score":       predictedScore - 0.1,
				},
				{
					"name":        "pessimistic",
					"probability": 0.2,
					"score":       predictedScore + 0.1,
				},
			}
		}

		predictions = append(predictions, prediction)
	}

	return map[string]interface{}{
		"merchantId":        merchantID,
		"horizons":          horizons,
		"predictions":       predictions,
		"includeScenarios":  includeScenarios,
		"includeConfidence": includeConfidence,
	}, nil
}

// ExplainAssessment provides explainability data for a risk assessment
func (s *riskAssessmentService) ExplainAssessment(ctx context.Context, assessmentID string) (map[string]interface{}, error) {
	s.logger.Printf("Getting explanation for assessment: %s", assessmentID)

	assessment, err := s.repo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		return nil, fmt.Errorf("assessment not found: %w", err)
	}

	explanation := map[string]interface{}{
		"assessmentId": assessmentID,
		"factors":      []interface{}{},
		"shapValues":   map[string]float64{},
		"baseValue":    0.5,
		"prediction":   0.0,
	}

	if assessment.Result != nil {
		explanation["prediction"] = assessment.Result.OverallScore

		// Convert factors to explanation format
		factors := []interface{}{}
		shapValues := map[string]float64{}
		for _, factor := range assessment.Result.Factors {
			factors = append(factors, map[string]interface{}{
				"name":   factor.Name,
				"score":  factor.Score,
				"weight": factor.Weight,
			})
			shapValues[factor.Name] = factor.Score * factor.Weight
		}
		explanation["factors"] = factors
		explanation["shapValues"] = shapValues
	}

	return explanation, nil
}

// GetRecommendations retrieves risk mitigation recommendations for a merchant
func (s *riskAssessmentService) GetRecommendations(ctx context.Context, merchantID string) ([]map[string]interface{}, error) {
	s.logger.Printf("Getting risk recommendations for merchant: %s", merchantID)

	// Get latest assessment
	assessments, err := s.repo.GetAssessmentsByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assessments: %w", err)
	}

	recommendations := []map[string]interface{}{}

	if len(assessments) > 0 {
		latest := assessments[0]
		if latest.Result != nil {
			// Generate recommendations based on risk level
			if latest.Result.OverallScore > 0.7 {
				recommendations = append(recommendations, map[string]interface{}{
					"id":          "rec-1",
					"type":        "action",
					"priority":    "high",
					"title":       "High Risk Detected",
					"description": "This merchant shows elevated risk indicators. Consider additional due diligence and monitoring.",
					"actionItems": []string{
						"Conduct enhanced due diligence",
						"Increase monitoring frequency",
						"Review transaction patterns",
					},
				})
			}

			// Generate recommendations based on risk factors
			for _, factor := range latest.Result.Factors {
				if factor.Score > 0.6 {
					recommendations = append(recommendations, map[string]interface{}{
						"id":          fmt.Sprintf("rec-factor-%s", factor.Name),
						"type":        "factor",
						"priority":    "medium",
						"title":       fmt.Sprintf("Address %s Risk", factor.Name),
						"description": fmt.Sprintf("The %s factor shows elevated risk (score: %.2f). Consider mitigation strategies.", factor.Name, factor.Score),
						"actionItems": []string{
							fmt.Sprintf("Review %s indicators", factor.Name),
							"Implement mitigation measures",
						},
					})
				}
			}
		}
	}

	// Default recommendation if no assessments
	if len(recommendations) == 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"id":          "rec-default",
			"type":        "general",
			"priority":    "low",
			"title":       "Regular Monitoring Recommended",
			"description": "Continue regular risk monitoring and assessment.",
			"actionItems": []string{
				"Schedule periodic risk assessments",
				"Monitor for changes in risk profile",
			},
		})
	}

	return recommendations, nil
}

// calculateRiskLevel calculates risk level from score
func (s *riskAssessmentService) calculateRiskLevel(score float64) string {
	if score >= 0.8 {
		return "critical"
	} else if score >= 0.6 {
		return "high"
	} else if score >= 0.4 {
		return "medium"
	} else {
		return "low"
	}
}
