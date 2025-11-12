package feedback

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SupabaseFeedbackRepository implements the FeedbackRepository interface using Supabase
type SupabaseFeedbackRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSupabaseFeedbackRepository creates a new Supabase feedback repository
func NewSupabaseFeedbackRepository(db *sql.DB, logger *zap.Logger) *SupabaseFeedbackRepository {
	return &SupabaseFeedbackRepository{
		db:     db,
		logger: logger,
	}
}

// SaveUserFeedback saves user feedback to the database
func (r *SupabaseFeedbackRepository) SaveUserFeedback(ctx context.Context, feedback UserFeedback) error {
	query := `
		INSERT INTO user_feedback (
			id, user_id, business_name, original_classification_id, feedback_type,
			feedback_value, feedback_text, suggested_classification_id, confidence_score,
			status, processing_time_ms, model_version_id, classification_method,
			ensemble_weight, created_at, processed_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	// TODO: Refactor - UserFeedback doesn't have these fields, should use ClassificationClassificationUserFeedback
	// Stub values for now to allow compilation
	_, err := r.db.ExecContext(ctx, query,
		feedback.ID,
		feedback.UserID,
		"",                   // feedback.BusinessName - stub
		"",                   // feedback.OriginalClassificationID - stub
		"",                   // string(feedback.FeedbackType) - stub
		nil,                  // feedback.FeedbackValue - stub
		feedback.Comments,    // Use Comments as FeedbackText substitute
		"",                   // feedback.SuggestedClassificationID - stub
		0.0,                  // feedback.ConfidenceScore - stub
		"",                   // string(feedback.Status) - stub
		0,                    // feedback.ProcessingTimeMs - stub
		"",                   // feedback.ModelVersionID - stub
		"",                   // string(feedback.ClassificationMethod) - stub
		0.0,                  // feedback.EnsembleWeight - stub
		feedback.SubmittedAt, // Use SubmittedAt as CreatedAt substitute
		time.Time{},          // feedback.ProcessedAt - stub (not in UserFeedback)
		feedback.Metadata,
	)

	if err != nil {
		r.logger.Error("failed to save user feedback",
			zap.String("feedback_id", feedback.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to save user feedback: %w", err)
	}

	r.logger.Info("user feedback saved successfully",
		zap.String("feedback_id", feedback.ID.String()),
		zap.String("user_id", feedback.UserID))

	return nil
}

// GetUserFeedback retrieves user feedback by ID
func (r *SupabaseFeedbackRepository) GetUserFeedback(ctx context.Context, id string) (*UserFeedback, error) {
	query := `
		SELECT id, user_id, business_name, original_classification_id, feedback_type,
			   feedback_value, feedback_text, suggested_classification_id, confidence_score,
			   status, processing_time_ms, model_version_id, classification_method,
			   ensemble_weight, created_at, processed_at, metadata
		FROM user_feedback
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var feedback UserFeedback
	var feedbackTypeStr, statusStr, classificationMethodStr string
	// Temporary variables for fields that don't exist in UserFeedback
	var businessName, originalClassificationID, feedbackText, suggestedClassificationID, modelVersionID string
	var feedbackValue map[string]interface{}
	var confidenceScore, ensembleWeight float64
	var processingTimeMs int
	var createdAt, processedAt time.Time

	err := row.Scan(
		&feedback.ID,
		&feedback.UserID,
		&businessName,             // Stub - not in UserFeedback
		&originalClassificationID, // Stub - not in UserFeedback
		&feedbackTypeStr,
		&feedbackValue,             // Stub - not in UserFeedback
		&feedbackText,              // Stub - not in UserFeedback
		&suggestedClassificationID, // Stub - not in UserFeedback
		&confidenceScore,           // Stub - not in UserFeedback
		&statusStr,
		&processingTimeMs, // Stub - not in UserFeedback
		&modelVersionID,   // Stub - not in UserFeedback
		&classificationMethodStr,
		&ensembleWeight, // Stub - not in UserFeedback
		&createdAt,      // Stub - not in UserFeedback
		&processedAt,    // Stub - not in UserFeedback
		&feedback.Metadata,
	)

	// Map stub values to available UserFeedback fields where possible
	feedback.Comments = feedbackText // Use Comments as FeedbackText substitute
	feedback.SubmittedAt = createdAt // Use SubmittedAt as CreatedAt substitute

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user feedback not found: %s", id)
		}
		r.logger.Error("failed to get user feedback",
			zap.String("feedback_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user feedback: %w", err)
	}

	// TODO: UserFeedback doesn't have FeedbackType, Status, or ClassificationMethod fields
	// These would need to be stored in Metadata or the struct needs to be refactored
	// Stub - skip assignment

	return &feedback, nil
}

// GetUserFeedbackByClassificationID retrieves user feedback by classification ID
func (r *SupabaseFeedbackRepository) GetUserFeedbackByClassificationID(ctx context.Context, classificationID string) ([]UserFeedback, error) {
	query := `
		SELECT id, user_id, business_name, original_classification_id, feedback_type,
			   feedback_value, feedback_text, suggested_classification_id, confidence_score,
			   status, processing_time_ms, model_version_id, classification_method,
			   ensemble_weight, created_at, processed_at, metadata
		FROM user_feedback
		WHERE original_classification_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, classificationID)
	if err != nil {
		r.logger.Error("failed to query user feedback by classification ID",
			zap.String("classification_id", classificationID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query user feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []UserFeedback

	for rows.Next() {
		var feedback UserFeedback
		var feedbackTypeStr, statusStr, classificationMethodStr string
		// Temporary variables for fields that don't exist in UserFeedback
		var businessName, originalClassificationID, feedbackText, suggestedClassificationID, modelVersionID string
		var feedbackValue map[string]interface{}
		var confidenceScore, ensembleWeight float64
		var processingTimeMs int
		var createdAt, processedAt time.Time

		err := rows.Scan(
			&feedback.ID,
			&feedback.UserID,
			&businessName,             // Stub - not in UserFeedback
			&originalClassificationID, // Stub - not in UserFeedback
			&feedbackTypeStr,
			&feedbackValue,             // Stub - not in UserFeedback
			&feedbackText,              // Stub - not in UserFeedback
			&suggestedClassificationID, // Stub - not in UserFeedback
			&confidenceScore,           // Stub - not in UserFeedback
			&statusStr,
			&processingTimeMs, // Stub - not in UserFeedback
			&modelVersionID,   // Stub - not in UserFeedback
			&classificationMethodStr,
			&ensembleWeight, // Stub - not in UserFeedback
			&createdAt,      // Stub - not in UserFeedback
			&processedAt,    // Stub - not in UserFeedback
			&feedback.Metadata,
		)

		// Map stub values to available UserFeedback fields where possible
		feedback.Comments = feedbackText // Use Comments as FeedbackText substitute
		feedback.SubmittedAt = createdAt // Use SubmittedAt as CreatedAt substitute

		if err != nil {
			r.logger.Error("failed to scan user feedback row",
				zap.Error(err))
			continue
		}

		// TODO: UserFeedback doesn't have FeedbackType, Status, or ClassificationMethod fields
		// These would need to be stored in Metadata or the struct needs to be refactored
		// Stub - skip assignment

		feedbacks = append(feedbacks, feedback)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating user feedback rows",
			zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return feedbacks, nil
}

// UpdateUserFeedbackStatus updates the status of user feedback
func (r *SupabaseFeedbackRepository) UpdateUserFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error {
	query := `
		UPDATE user_feedback
		SET status = $1, processed_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, string(status), now, id)

	if err != nil {
		r.logger.Error("failed to update user feedback status",
			zap.String("feedback_id", id),
			zap.String("status", string(status)),
			zap.Error(err))
		return fmt.Errorf("failed to update user feedback status: %w", err)
	}

	r.logger.Info("user feedback status updated",
		zap.String("feedback_id", id),
		zap.String("status", string(status)))

	return nil
}

// SaveMLModelFeedback saves ML model feedback to the database
func (r *SupabaseFeedbackRepository) SaveMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error {
	query := `
		INSERT INTO ml_model_feedback (
			id, model_version_id, model_type, classification_method, prediction_id,
			actual_result, predicted_result, accuracy_score, confidence_score,
			processing_time_ms, error_type, error_description, status,
			created_at, processed_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		feedback.ID,
		feedback.ModelVersionID,
		feedback.ModelType,
		string(feedback.ClassificationMethod),
		feedback.PredictionID,
		feedback.ActualResult,
		feedback.PredictedResult,
		feedback.AccuracyScore,
		feedback.ConfidenceScore,
		feedback.ProcessingTimeMs,
		feedback.ErrorType,
		feedback.ErrorDescription,
		string(feedback.Status),
		feedback.CreatedAt,
		feedback.ProcessedAt,
		feedback.Metadata,
	)

	if err != nil {
		r.logger.Error("failed to save ML model feedback",
			zap.String("feedback_id", feedback.ID),
			zap.Error(err))
		return fmt.Errorf("failed to save ML model feedback: %w", err)
	}

	r.logger.Info("ML model feedback saved successfully",
		zap.String("feedback_id", feedback.ID),
		zap.String("model_type", feedback.ModelType))

	return nil
}

// GetMLModelFeedback retrieves ML model feedback by ID
func (r *SupabaseFeedbackRepository) GetMLModelFeedback(ctx context.Context, id string) (*MLModelFeedback, error) {
	query := `
		SELECT id, model_version_id, model_type, classification_method, prediction_id,
			   actual_result, predicted_result, accuracy_score, confidence_score,
			   processing_time_ms, error_type, error_description, status,
			   created_at, processed_at, metadata
		FROM ml_model_feedback
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var feedback MLModelFeedback
	var classificationMethodStr, statusStr string

	err := row.Scan(
		&feedback.ID,
		&feedback.ModelVersionID,
		&feedback.ModelType,
		&classificationMethodStr,
		&feedback.PredictionID,
		&feedback.ActualResult,
		&feedback.PredictedResult,
		&feedback.AccuracyScore,
		&feedback.ConfidenceScore,
		&feedback.ProcessingTimeMs,
		&feedback.ErrorType,
		&feedback.ErrorDescription,
		&statusStr,
		&feedback.CreatedAt,
		&feedback.ProcessedAt,
		&feedback.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ML model feedback not found: %s", id)
		}
		r.logger.Error("failed to get ML model feedback",
			zap.String("feedback_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get ML model feedback: %w", err)
	}

	feedback.ClassificationMethod = ClassificationMethod(classificationMethodStr)
	feedback.Status = FeedbackStatus(statusStr)

	return &feedback, nil
}

// GetMLModelFeedbackByModelVersion retrieves ML model feedback by model version
func (r *SupabaseFeedbackRepository) GetMLModelFeedbackByModelVersion(ctx context.Context, modelVersion string) ([]MLModelFeedback, error) {
	query := `
		SELECT id, model_version_id, model_type, classification_method, prediction_id,
			   actual_result, predicted_result, accuracy_score, confidence_score,
			   processing_time_ms, error_type, error_description, status,
			   created_at, processed_at, metadata
		FROM ml_model_feedback
		WHERE model_version_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, modelVersion)
	if err != nil {
		r.logger.Error("failed to query ML model feedback by model version",
			zap.String("model_version", modelVersion),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query ML model feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []MLModelFeedback

	for rows.Next() {
		var feedback MLModelFeedback
		var classificationMethodStr, statusStr string

		err := rows.Scan(
			&feedback.ID,
			&feedback.ModelVersionID,
			&feedback.ModelType,
			&classificationMethodStr,
			&feedback.PredictionID,
			&feedback.ActualResult,
			&feedback.PredictedResult,
			&feedback.AccuracyScore,
			&feedback.ConfidenceScore,
			&feedback.ProcessingTimeMs,
			&feedback.ErrorType,
			&feedback.ErrorDescription,
			&statusStr,
			&feedback.CreatedAt,
			&feedback.ProcessedAt,
			&feedback.Metadata,
		)

		if err != nil {
			r.logger.Error("failed to scan ML model feedback row",
				zap.Error(err))
			continue
		}

		feedback.ClassificationMethod = ClassificationMethod(classificationMethodStr)
		feedback.Status = FeedbackStatus(statusStr)

		feedbacks = append(feedbacks, feedback)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating ML model feedback rows",
			zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return feedbacks, nil
}

// UpdateMLModelFeedbackStatus updates the status of ML model feedback
func (r *SupabaseFeedbackRepository) UpdateMLModelFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error {
	query := `
		UPDATE ml_model_feedback
		SET status = $1, processed_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, string(status), now, id)

	if err != nil {
		r.logger.Error("failed to update ML model feedback status",
			zap.String("feedback_id", id),
			zap.String("status", string(status)),
			zap.Error(err))
		return fmt.Errorf("failed to update ML model feedback status: %w", err)
	}

	r.logger.Info("ML model feedback status updated",
		zap.String("feedback_id", id),
		zap.String("status", string(status)))

	return nil
}

// SaveSecurityValidationFeedback saves security validation feedback to the database
func (r *SupabaseFeedbackRepository) SaveSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error {
	query := `
		INSERT INTO security_validation_feedback (
			id, validation_type, data_source_type, website_url, validation_result,
			trust_score, verification_status, security_violations, processing_time_ms,
			status, created_at, processed_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		feedback.ID,
		feedback.ValidationType,
		feedback.DataSourceType,
		feedback.WebsiteURL,
		feedback.ValidationResult,
		feedback.TrustScore,
		feedback.VerificationStatus,
		feedback.SecurityViolations,
		feedback.ProcessingTimeMs,
		string(feedback.Status),
		feedback.CreatedAt,
		feedback.ProcessedAt,
		feedback.Metadata,
	)

	if err != nil {
		r.logger.Error("failed to save security validation feedback",
			zap.String("feedback_id", feedback.ID),
			zap.Error(err))
		return fmt.Errorf("failed to save security validation feedback: %w", err)
	}

	r.logger.Info("security validation feedback saved successfully",
		zap.String("feedback_id", feedback.ID),
		zap.String("validation_type", feedback.ValidationType))

	return nil
}

// GetSecurityValidationFeedback retrieves security validation feedback by ID
func (r *SupabaseFeedbackRepository) GetSecurityValidationFeedback(ctx context.Context, id string) (*SecurityValidationFeedback, error) {
	query := `
		SELECT id, validation_type, data_source_type, website_url, validation_result,
			   trust_score, verification_status, security_violations, processing_time_ms,
			   status, created_at, processed_at, metadata
		FROM security_validation_feedback
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var feedback SecurityValidationFeedback
	var statusStr string

	err := row.Scan(
		&feedback.ID,
		&feedback.ValidationType,
		&feedback.DataSourceType,
		&feedback.WebsiteURL,
		&feedback.ValidationResult,
		&feedback.TrustScore,
		&feedback.VerificationStatus,
		&feedback.SecurityViolations,
		&feedback.ProcessingTimeMs,
		&statusStr,
		&feedback.CreatedAt,
		&feedback.ProcessedAt,
		&feedback.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("security validation feedback not found: %s", id)
		}
		r.logger.Error("failed to get security validation feedback",
			zap.String("feedback_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get security validation feedback: %w", err)
	}

	feedback.Status = FeedbackStatus(statusStr)

	return &feedback, nil
}

// GetSecurityValidationFeedbackByType retrieves security validation feedback by type
func (r *SupabaseFeedbackRepository) GetSecurityValidationFeedbackByType(ctx context.Context, validationType string) ([]SecurityValidationFeedback, error) {
	query := `
		SELECT id, validation_type, data_source_type, website_url, validation_result,
			   trust_score, verification_status, security_violations, processing_time_ms,
			   status, created_at, processed_at, metadata
		FROM security_validation_feedback
		WHERE validation_type = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, validationType)
	if err != nil {
		r.logger.Error("failed to query security validation feedback by type",
			zap.String("validation_type", validationType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query security validation feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []SecurityValidationFeedback

	for rows.Next() {
		var feedback SecurityValidationFeedback
		var statusStr string

		err := rows.Scan(
			&feedback.ID,
			&feedback.ValidationType,
			&feedback.DataSourceType,
			&feedback.WebsiteURL,
			&feedback.ValidationResult,
			&feedback.TrustScore,
			&feedback.VerificationStatus,
			&feedback.SecurityViolations,
			&feedback.ProcessingTimeMs,
			&statusStr,
			&feedback.CreatedAt,
			&feedback.ProcessedAt,
			&feedback.Metadata,
		)

		if err != nil {
			r.logger.Error("failed to scan security validation feedback row",
				zap.Error(err))
			continue
		}

		feedback.Status = FeedbackStatus(statusStr)

		feedbacks = append(feedbacks, feedback)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating security validation feedback rows",
			zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return feedbacks, nil
}

// UpdateSecurityValidationFeedbackStatus updates the status of security validation feedback
func (r *SupabaseFeedbackRepository) UpdateSecurityValidationFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error {
	query := `
		UPDATE security_validation_feedback
		SET status = $1, processed_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, string(status), now, id)

	if err != nil {
		r.logger.Error("failed to update security validation feedback status",
			zap.String("feedback_id", id),
			zap.String("status", string(status)),
			zap.Error(err))
		return fmt.Errorf("failed to update security validation feedback status: %w", err)
	}

	r.logger.Info("security validation feedback status updated",
		zap.String("feedback_id", id),
		zap.String("status", string(status)))

	return nil
}

// GetFeedbackTrends retrieves feedback trends for analysis
func (r *SupabaseFeedbackRepository) GetFeedbackTrends(ctx context.Context, request FeedbackAnalysisRequest) ([]FeedbackTrend, error) {
	query := `
		SELECT method, time_window, total_feedback, positive_feedback, negative_feedback,
			   average_accuracy, average_confidence, average_processing_time, error_rate,
			   security_violation_rate, created_at, updated_at
		FROM feedback_trends
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if request.Method != "" {
		query += fmt.Sprintf(" AND method = $%d", argIndex)
		args = append(args, string(request.Method))
		argIndex++
	}

	if request.TimeWindow != "" {
		query += fmt.Sprintf(" AND time_window = $%d", argIndex)
		args = append(args, request.TimeWindow)
		argIndex++
	}

	if request.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *request.StartDate)
		argIndex++
	}

	if request.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *request.EndDate)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to query feedback trends",
			zap.Error(err))
		return nil, fmt.Errorf("failed to query feedback trends: %w", err)
	}
	defer rows.Close()

	var trends []FeedbackTrend

	for rows.Next() {
		var trend FeedbackTrend
		var methodStr string

		err := rows.Scan(
			&methodStr,
			&trend.TimeWindow,
			&trend.TotalFeedback,
			&trend.PositiveFeedback,
			&trend.NegativeFeedback,
			&trend.AverageAccuracy,
			&trend.AverageConfidence,
			&trend.AverageProcessingTime,
			&trend.ErrorRate,
			&trend.SecurityViolationRate,
			&trend.CreatedAt,
			&trend.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("failed to scan feedback trend row",
				zap.Error(err))
			continue
		}

		trend.Method = ClassificationMethod(methodStr)
		trends = append(trends, trend)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating feedback trend rows",
			zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return trends, nil
}

// GetFeedbackStatistics retrieves aggregated feedback statistics
func (r *SupabaseFeedbackRepository) GetFeedbackStatistics(ctx context.Context, request FeedbackAnalysisRequest) (*FeedbackAnalysisResponse, error) {
	// Get trends first
	trends, err := r.GetFeedbackTrends(ctx, request)
	if err != nil {
		return nil, err
	}

	// Calculate aggregated statistics
	var totalFeedback int
	var totalAccuracy, totalConfidence, totalErrorRate, totalSecurityViolationRate float64

	for _, trend := range trends {
		totalFeedback += trend.TotalFeedback
		totalAccuracy += trend.AverageAccuracy * float64(trend.TotalFeedback)
		totalConfidence += trend.AverageConfidence * float64(trend.TotalFeedback)
		totalErrorRate += trend.ErrorRate * float64(trend.TotalFeedback)
		totalSecurityViolationRate += trend.SecurityViolationRate * float64(trend.TotalFeedback)
	}

	var averageAccuracy, averageConfidence, errorRate, securityViolationRate float64
	if totalFeedback > 0 {
		averageAccuracy = totalAccuracy / float64(totalFeedback)
		averageConfidence = totalConfidence / float64(totalFeedback)
		errorRate = totalErrorRate / float64(totalFeedback)
		securityViolationRate = totalSecurityViolationRate / float64(totalFeedback)
	}

	return &FeedbackAnalysisResponse{
		Trends:                trends,
		TotalFeedback:         totalFeedback,
		AverageAccuracy:       averageAccuracy,
		AverageConfidence:     averageConfidence,
		ErrorRate:             errorRate,
		SecurityViolationRate: securityViolationRate,
		AnalysisTime:          time.Now(),
	}, nil
}

// Batch operations

// SaveBatchUserFeedback saves multiple user feedback records in batch
func (r *SupabaseFeedbackRepository) SaveBatchUserFeedback(ctx context.Context, feedbacks []UserFeedback) error {
	if len(feedbacks) == 0 {
		return nil
	}

	query := `
		INSERT INTO user_feedback (
			id, user_id, business_name, original_classification_id, feedback_type,
			feedback_value, feedback_text, suggested_classification_id, confidence_score,
			status, processing_time_ms, model_version_id, classification_method,
			ensemble_weight, created_at, processed_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, feedback := range feedbacks {
		// TODO: Refactor - UserFeedback doesn't have these fields, should use ClassificationClassificationUserFeedback
		// Stub values for now to allow compilation
		_, err := stmt.ExecContext(ctx,
			feedback.ID,
			feedback.UserID,
			"",                   // feedback.BusinessName - stub
			"",                   // feedback.OriginalClassificationID - stub
			"",                   // string(feedback.FeedbackType) - stub
			nil,                  // feedback.FeedbackValue - stub
			feedback.Comments,    // Use Comments as FeedbackText substitute
			"",                   // feedback.SuggestedClassificationID - stub
			0.0,                  // feedback.ConfidenceScore - stub
			"",                   // string(feedback.Status) - stub
			0,                    // feedback.ProcessingTimeMs - stub
			"",                   // feedback.ModelVersionID - stub
			"",                   // string(feedback.ClassificationMethod) - stub
			0.0,                  // feedback.EnsembleWeight - stub
			feedback.SubmittedAt, // Use SubmittedAt as CreatedAt substitute
			time.Time{},          // feedback.ProcessedAt - stub
			feedback.Metadata,
		)

		if err != nil {
			r.logger.Error("failed to execute batch user feedback insert",
				zap.String("feedback_id", feedback.ID.String()),
				zap.Error(err))
			return fmt.Errorf("failed to insert user feedback: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("batch user feedback saved successfully",
		zap.Int("count", len(feedbacks)))

	return nil
}

// SaveBatchMLModelFeedback saves multiple ML model feedback records in batch
func (r *SupabaseFeedbackRepository) SaveBatchMLModelFeedback(ctx context.Context, feedbacks []MLModelFeedback) error {
	if len(feedbacks) == 0 {
		return nil
	}

	query := `
		INSERT INTO ml_model_feedback (
			id, model_version_id, model_type, classification_method, prediction_id,
			actual_result, predicted_result, accuracy_score, confidence_score,
			processing_time_ms, error_type, error_description, status,
			created_at, processed_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, feedback := range feedbacks {
		_, err := stmt.ExecContext(ctx,
			feedback.ID,
			feedback.ModelVersionID,
			feedback.ModelType,
			string(feedback.ClassificationMethod),
			feedback.PredictionID,
			feedback.ActualResult,
			feedback.PredictedResult,
			feedback.AccuracyScore,
			feedback.ConfidenceScore,
			feedback.ProcessingTimeMs,
			feedback.ErrorType,
			feedback.ErrorDescription,
			string(feedback.Status),
			feedback.CreatedAt,
			feedback.ProcessedAt,
			feedback.Metadata,
		)

		if err != nil {
			r.logger.Error("failed to execute batch ML model feedback insert",
				zap.String("feedback_id", feedback.ID),
				zap.Error(err))
			return fmt.Errorf("failed to insert ML model feedback: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("batch ML model feedback saved successfully",
		zap.Int("count", len(feedbacks)))

	return nil
}

// SaveBatchSecurityValidationFeedback saves multiple security validation feedback records in batch
func (r *SupabaseFeedbackRepository) SaveBatchSecurityValidationFeedback(ctx context.Context, feedbacks []SecurityValidationFeedback) error {
	if len(feedbacks) == 0 {
		return nil
	}

	query := `
		INSERT INTO security_validation_feedback (
			id, validation_type, data_source_type, website_url, validation_result,
			trust_score, verification_status, security_violations, processing_time_ms,
			status, created_at, processed_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, feedback := range feedbacks {
		_, err := stmt.ExecContext(ctx,
			feedback.ID,
			feedback.ValidationType,
			feedback.DataSourceType,
			feedback.WebsiteURL,
			feedback.ValidationResult,
			feedback.TrustScore,
			feedback.VerificationStatus,
			feedback.SecurityViolations,
			feedback.ProcessingTimeMs,
			string(feedback.Status),
			feedback.CreatedAt,
			feedback.ProcessedAt,
			feedback.Metadata,
		)

		if err != nil {
			r.logger.Error("failed to execute batch security validation feedback insert",
				zap.String("feedback_id", feedback.ID),
				zap.Error(err))
			return fmt.Errorf("failed to insert security validation feedback: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("batch security validation feedback saved successfully",
		zap.Int("count", len(feedbacks)))

	return nil
}
