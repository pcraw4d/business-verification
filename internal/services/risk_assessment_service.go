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
