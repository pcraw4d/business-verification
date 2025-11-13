package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/models"
)

// RiskAssessmentRepository provides data access operations for risk assessments
type RiskAssessmentRepository struct {
	db     *sql.DB
	logger *log.Logger
}

// NewRiskAssessmentRepository creates a new risk assessment repository
func NewRiskAssessmentRepository(db *sql.DB, logger *log.Logger) *RiskAssessmentRepository {
	if logger == nil {
		logger = log.Default()
	}

	return &RiskAssessmentRepository{
		db:     db,
		logger: logger,
	}
}

// Repository errors
var (
	ErrAssessmentNotFound = errors.New("assessment not found")
)

// CreateAssessment creates a new risk assessment record
func (r *RiskAssessmentRepository) CreateAssessment(ctx context.Context, assessment *models.RiskAssessment) error {
	r.logger.Printf("Creating risk assessment: %s for merchant: %s", assessment.ID, assessment.MerchantID)

	// Serialize options and result to JSON
	optionsJSON, err := json.Marshal(assessment.Options)
	if err != nil {
		return fmt.Errorf("failed to marshal options: %w", err)
	}

	var resultJSON []byte
	if assessment.Result != nil {
		resultJSON, err = json.Marshal(assessment.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	// Check if async columns exist, use appropriate query
	// For now, assume migration has been run - if not, this will fail and migration needs to be run
	query := `
		INSERT INTO risk_assessments (
			id, merchant_id, status, options, result, progress,
			estimated_completion, created_at, updated_at, completed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			options = EXCLUDED.options,
			result = EXCLUDED.result,
			progress = EXCLUDED.progress,
			estimated_completion = EXCLUDED.estimated_completion,
			updated_at = EXCLUDED.updated_at,
			completed_at = EXCLUDED.completed_at
	`

	_, err = r.db.ExecContext(ctx, query,
		assessment.ID,
		assessment.MerchantID,
		string(assessment.Status),
		optionsJSON,
		resultJSON,
		assessment.Progress,
		assessment.EstimatedCompletion,
		assessment.CreatedAt,
		assessment.UpdatedAt,
		assessment.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create assessment: %w", err)
	}

	r.logger.Printf("Successfully created risk assessment: %s", assessment.ID)
	return nil
}

// GetAssessmentByID retrieves a risk assessment by ID
func (r *RiskAssessmentRepository) GetAssessmentByID(ctx context.Context, assessmentID string) (*models.RiskAssessment, error) {
	r.logger.Printf("Retrieving risk assessment: %s", assessmentID)

	query := `
		SELECT 
			id, merchant_id, status, options, result, progress,
			estimated_completion, created_at, updated_at, completed_at
		FROM risk_assessments
		WHERE id = $1
	`

	var assessment models.RiskAssessment
	var statusStr string
	var optionsJSON, resultJSON []byte
	var estimatedCompletion, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, assessmentID).Scan(
		&assessment.ID,
		&assessment.MerchantID,
		&statusStr,
		&optionsJSON,
		&resultJSON,
		&assessment.Progress,
		&estimatedCompletion,
		&assessment.CreatedAt,
		&assessment.UpdatedAt,
		&completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAssessmentNotFound
		}
		return nil, fmt.Errorf("failed to retrieve assessment: %w", err)
	}

	// Parse status
	assessment.Status = models.AssessmentStatus(statusStr)

	// Parse options
	if len(optionsJSON) > 0 {
		if err := json.Unmarshal(optionsJSON, &assessment.Options); err != nil {
			r.logger.Printf("Warning: failed to unmarshal options: %v", err)
		}
	}

	// Parse result
	if len(resultJSON) > 0 {
		var result models.RiskAssessmentResult
		if err := json.Unmarshal(resultJSON, &result); err != nil {
			r.logger.Printf("Warning: failed to unmarshal result: %v", err)
		} else {
			assessment.Result = &result
		}
	}

	// Parse timestamps
	if estimatedCompletion.Valid {
		assessment.EstimatedCompletion = &estimatedCompletion.Time
	}
	if completedAt.Valid {
		assessment.CompletedAt = &completedAt.Time
	}

	return &assessment, nil
}

// UpdateAssessmentStatus updates the status and progress of an assessment
func (r *RiskAssessmentRepository) UpdateAssessmentStatus(ctx context.Context, assessmentID string, status models.AssessmentStatus, progress int) error {
	r.logger.Printf("Updating assessment status: %s to %s (progress: %d%%)", assessmentID, status, progress)

	query := `
		UPDATE risk_assessments
		SET status = $1, progress = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, string(status), progress, time.Now(), assessmentID)
	if err != nil {
		return fmt.Errorf("failed to update assessment status: %w", err)
	}

	return nil
}

// UpdateAssessmentResult updates the assessment with the final result
func (r *RiskAssessmentRepository) UpdateAssessmentResult(ctx context.Context, assessmentID string, result *models.RiskAssessmentResult) error {
	r.logger.Printf("Updating assessment result: %s", assessmentID)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	query := `
		UPDATE risk_assessments
		SET 
			result = $1,
			status = $2,
			progress = 100,
			completed_at = $3,
			updated_at = $3
		WHERE id = $4
	`

	_, err = r.db.ExecContext(ctx, query,
		resultJSON,
		string(models.AssessmentStatusCompleted),
		time.Now(),
		assessmentID,
	)

	if err != nil {
		return fmt.Errorf("failed to update assessment result: %w", err)
	}

	return nil
}

// GetAssessmentsByMerchantID retrieves all assessments for a merchant
func (r *RiskAssessmentRepository) GetAssessmentsByMerchantID(ctx context.Context, merchantID string) ([]*models.RiskAssessment, error) {
	r.logger.Printf("Retrieving assessments for merchant: %s", merchantID)

	query := `
		SELECT 
			id, merchant_id, status, options, result, progress,
			estimated_completion, created_at, updated_at, completed_at
		FROM risk_assessments
		WHERE merchant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query assessments: %w", err)
	}
	defer rows.Close()

	assessments := []*models.RiskAssessment{}
	for rows.Next() {
		var assessment models.RiskAssessment
		var statusStr string
		var optionsJSON, resultJSON []byte
		var estimatedCompletion, completedAt sql.NullTime

		err := rows.Scan(
			&assessment.ID,
			&assessment.MerchantID,
			&statusStr,
			&optionsJSON,
			&resultJSON,
			&assessment.Progress,
			&estimatedCompletion,
			&assessment.CreatedAt,
			&assessment.UpdatedAt,
			&completedAt,
		)

		if err != nil {
			r.logger.Printf("Error scanning assessment: %v", err)
			continue
		}

		assessment.Status = models.AssessmentStatus(statusStr)

		if len(optionsJSON) > 0 {
			if err := json.Unmarshal(optionsJSON, &assessment.Options); err != nil {
				r.logger.Printf("Warning: Failed to unmarshal assessment options JSON for assessment %s: %v", assessment.ID, err)
				// Continue with zero value for Options rather than failing silently
			}
		}

		if len(resultJSON) > 0 {
			var result models.RiskAssessmentResult
			if err := json.Unmarshal(resultJSON, &result); err == nil {
				assessment.Result = &result
			}
		}

		if estimatedCompletion.Valid {
			assessment.EstimatedCompletion = &estimatedCompletion.Time
		}
		if completedAt.Valid {
			assessment.CompletedAt = &completedAt.Time
		}

		assessments = append(assessments, &assessment)
	}

	return assessments, nil
}
