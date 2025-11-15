package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
)

// RiskAssessmentJob represents a risk assessment job to be processed
type RiskAssessmentJob struct {
	AssessmentID string
	MerchantID   string
	Options      models.AssessmentOptions
}

// RiskAssessmentJobProcessor processes risk assessment jobs asynchronously
type RiskAssessmentJobProcessor struct {
	repo        *database.RiskAssessmentRepository
	riskService RiskAssessmentService
	logger      *log.Logger
	jobQueue    chan RiskAssessmentJob
	workerCount int
	stopChan    chan struct{}
}

// RiskAssessmentService interface for performing actual risk assessment
type RiskAssessmentService interface {
	PerformRiskAssessment(ctx context.Context, merchantID string, options models.AssessmentOptions) (*models.RiskAssessmentResult, error)
}

// NewRiskAssessmentJobProcessor creates a new risk assessment job processor
func NewRiskAssessmentJobProcessor(
	repo *database.RiskAssessmentRepository,
	riskService RiskAssessmentService,
	logger *log.Logger,
	workerCount int,
) *RiskAssessmentJobProcessor {
	if logger == nil {
		logger = log.Default()
	}
	if workerCount <= 0 {
		workerCount = 3 // Default to 3 workers
	}

	processor := &RiskAssessmentJobProcessor{
		repo:        repo,
		riskService: riskService,
		logger:      logger,
		jobQueue:    make(chan RiskAssessmentJob, 100), // Buffer up to 100 jobs
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}

	// Start workers
	processor.startWorkers()

	return processor
}

// Enqueue adds a job to the processing queue
func (p *RiskAssessmentJobProcessor) Enqueue(ctx context.Context, job RiskAssessmentJob) error {
	select {
	case p.jobQueue <- job:
		p.logger.Printf("Enqueued risk assessment job: %s", job.AssessmentID)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("job queue is full")
	}
}

// startWorkers starts the worker goroutines
func (p *RiskAssessmentJobProcessor) startWorkers() {
	for i := 0; i < p.workerCount; i++ {
		go p.worker(i)
	}
	p.logger.Printf("Started %d risk assessment workers", p.workerCount)
}

// worker processes jobs from the queue
func (p *RiskAssessmentJobProcessor) worker(id int) {
	p.logger.Printf("Risk assessment worker %d started", id)

	for {
		select {
		case job := <-p.jobQueue:
			p.processJob(context.Background(), job, id)
		case <-p.stopChan:
			p.logger.Printf("Risk assessment worker %d stopped", id)
			return
		}
	}
}

// Process processes a risk assessment job (public method for direct processing)
func (j *RiskAssessmentJob) Process(ctx context.Context, repo *database.RiskAssessmentRepository, riskService RiskAssessmentService, logger *log.Logger) error {
	if logger == nil {
		logger = log.Default()
	}

	logger.Printf("Processing assessment: %s for merchant: %s", j.AssessmentID, j.MerchantID)

	// Update status to processing
	err := repo.UpdateAssessmentStatus(ctx, j.AssessmentID, models.AssessmentStatusProcessing, 10)
	if err != nil {
		logger.Printf("Error updating status to processing: %v", err)
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Perform the actual risk assessment
	result, err := riskService.PerformRiskAssessment(ctx, j.MerchantID, j.Options)
	if err != nil {
		logger.Printf("Error performing risk assessment: %v", err)
		// Update status to failed
		repo.UpdateAssessmentStatus(ctx, j.AssessmentID, models.AssessmentStatusFailed, 0)
		return fmt.Errorf("risk assessment failed: %w", err)
	}

	// Set assessment ID on result (if result struct supports it)
	// Note: RiskAssessmentResult doesn't have an ID field, so we'll set it in the update

	// Save assessment results to repository
	err = repo.UpdateAssessmentResult(ctx, j.AssessmentID, result)
	if err != nil {
		logger.Printf("Error saving assessment result: %v", err)
		repo.UpdateAssessmentStatus(ctx, j.AssessmentID, models.AssessmentStatusFailed, 0)
		return fmt.Errorf("failed to save result: %w", err)
	}

	// Update status to completed
	completedAt := time.Now()
	err = repo.UpdateAssessmentStatus(ctx, j.AssessmentID, models.AssessmentStatusCompleted, 100)
	if err != nil {
		logger.Printf("Error updating status to completed: %v", err)
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Update completed timestamp if repository supports it
	_ = completedAt // Use completedAt if repository method exists

	logger.Printf("Assessment completed: %s", j.AssessmentID)
	return nil
}

// processJob processes a single risk assessment job (internal method used by processor)
func (p *RiskAssessmentJobProcessor) processJob(ctx context.Context, job RiskAssessmentJob, workerID int) {
	p.logger.Printf("Worker %d processing assessment: %s", workerID, job.AssessmentID)

	// Use the Process method
	err := job.Process(ctx, p.repo, p.riskService, p.logger)
	if err != nil {
		p.logger.Printf("Worker %d failed to process assessment %s: %v", workerID, job.AssessmentID, err)
		return
	}

	p.logger.Printf("Worker %d completed assessment: %s", workerID, job.AssessmentID)
}

// updateProgress simulates progress updates during assessment
func (p *RiskAssessmentJobProcessor) updateProgress(ctx context.Context, assessmentID string) {
	progressValues := []int{20, 40, 60, 80}
	for _, progress := range progressValues {
		select {
		case <-time.After(2 * time.Second):
			p.repo.UpdateAssessmentStatus(ctx, assessmentID, models.AssessmentStatusProcessing, progress)
		case <-ctx.Done():
			return
		}
	}
}

// Stop stops all workers
func (p *RiskAssessmentJobProcessor) Stop() {
	close(p.stopChan)
	p.logger.Printf("Stopping risk assessment job processor")
}

// MockRiskAssessmentService is a mock implementation for testing
type MockRiskAssessmentService struct {
	logger *log.Logger
}

// NewMockRiskAssessmentService creates a new mock risk assessment service
func NewMockRiskAssessmentService(logger *log.Logger) *MockRiskAssessmentService {
	return &MockRiskAssessmentService{
		logger: logger,
	}
}

// PerformRiskAssessment performs a mock risk assessment
func (m *MockRiskAssessmentService) PerformRiskAssessment(ctx context.Context, merchantID string, options models.AssessmentOptions) (*models.RiskAssessmentResult, error) {
	if m.logger != nil {
		m.logger.Printf("Performing mock risk assessment for merchant: %s", merchantID)
	}

	// Simulate processing time
	time.Sleep(5 * time.Second)

	// Generate mock result
	result := &models.RiskAssessmentResult{
		OverallScore: 0.75,
		RiskLevel:    "medium",
		Factors: []models.RiskFactor{
			{
				Name:   "Financial Stability",
				Score:  0.8,
				Weight: 0.3,
			},
			{
				Name:   "Business History",
				Score:  0.7,
				Weight: 0.25,
			},
			{
				Name:   "Compliance",
				Score:  0.75,
				Weight: 0.25,
			},
			{
				Name:   "Industry Risk",
				Score:  0.7,
				Weight: 0.2,
			},
		},
	}

	return result, nil
}
