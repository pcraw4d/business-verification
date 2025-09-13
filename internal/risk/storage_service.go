package risk

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RiskStorageService handles all risk data storage operations
type RiskStorageService struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewRiskStorageService creates a new risk storage service
func NewRiskStorageService(db *sql.DB, logger *zap.Logger) *RiskStorageService {
	return &RiskStorageService{
		db:     db,
		logger: logger,
	}
}

// RiskDataStorage represents the database model for risk data storage
type RiskDataStorage struct {
	ID               string                 `json:"id" db:"id"`
	BusinessID       string                 `json:"business_id" db:"business_id"`
	BusinessName     string                 `json:"business_name" db:"business_name"`
	OverallScore     float64                `json:"overall_score" db:"overall_score"`
	OverallLevel     string                 `json:"overall_level" db:"overall_level"`
	CategoryScores   map[string]interface{} `json:"category_scores" db:"category_scores"`
	FactorScores     []RiskScore            `json:"factor_scores" db:"factor_scores"`
	Recommendations  []RiskRecommendation   `json:"recommendations" db:"recommendations"`
	Alerts           []RiskAlert            `json:"alerts" db:"alerts"`
	AssessmentMethod string                 `json:"assessment_method" db:"assessment_method"`
	Source           string                 `json:"source" db:"source"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	AssessedAt       time.Time              `json:"assessed_at" db:"assessed_at"`
	ValidUntil       time.Time              `json:"valid_until" db:"valid_until"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// StoreRiskAssessment stores a risk assessment in the database
func (s *RiskStorageService) StoreRiskAssessment(ctx context.Context, assessment *RiskAssessment) error {
	requestID := "unknown"
	if id := ctx.Value("request_id"); id != nil {
		if str, ok := id.(string); ok {
			requestID = str
		}
	}

	s.logger.Info("Storing risk assessment",
		zap.String("request_id", requestID),
		zap.String("business_id", assessment.BusinessID),
		zap.String("assessment_id", assessment.ID),
		zap.Float64("overall_score", assessment.OverallScore),
		zap.String("overall_level", string(assessment.OverallLevel)),
	)

	// Generate ID if not provided
	if assessment.ID == "" {
		assessment.ID = uuid.New().String()
	}

	// Convert category scores to JSON
	categoryScoresJSON, err := json.Marshal(assessment.CategoryScores)
	if err != nil {
		return fmt.Errorf("failed to marshal category scores: %w", err)
	}

	// Convert factor scores to JSON
	factorScoresJSON, err := json.Marshal(assessment.FactorScores)
	if err != nil {
		return fmt.Errorf("failed to marshal factor scores: %w", err)
	}

	// Convert recommendations to JSON
	recommendationsJSON, err := json.Marshal(assessment.Recommendations)
	if err != nil {
		return fmt.Errorf("failed to marshal recommendations: %w", err)
	}

	// Convert alerts to JSON
	alertsJSON, err := json.Marshal(assessment.Alerts)
	if err != nil {
		return fmt.Errorf("failed to marshal alerts: %w", err)
	}

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(assessment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert risk assessment
	query := `
		INSERT INTO risk_assessments (
			id, business_id, business_name, overall_score, overall_level,
			category_scores, factor_scores, recommendations, alerts,
			assessment_method, source, metadata, assessed_at, valid_until,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	now := time.Now()
	_, err = s.db.ExecContext(ctx, query,
		assessment.ID,
		assessment.BusinessID,
		assessment.BusinessName,
		assessment.OverallScore,
		string(assessment.OverallLevel),
		categoryScoresJSON,
		factorScoresJSON,
		recommendationsJSON,
		alertsJSON,
		"comprehensive", // assessment_method
		"risk_service",  // source
		metadataJSON,
		assessment.AssessedAt,
		assessment.ValidUntil,
		now,
		now,
	)

	if err != nil {
		s.logger.Error("Failed to store risk assessment",
			zap.String("request_id", requestID),
			zap.String("business_id", assessment.BusinessID),
			zap.String("assessment_id", assessment.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to store risk assessment: %w", err)
	}

	s.logger.Info("Risk assessment stored successfully",
		zap.String("request_id", requestID),
		zap.String("business_id", assessment.BusinessID),
		zap.String("assessment_id", assessment.ID),
	)

	return nil
}

// GetRiskAssessment retrieves a risk assessment by ID
func (s *RiskStorageService) GetRiskAssessment(ctx context.Context, id string) (*RiskAssessment, error) {
	requestID := "unknown"
	if rid := ctx.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID = str
		}
	}

	s.logger.Info("Retrieving risk assessment",
		zap.String("request_id", requestID),
		zap.String("assessment_id", id),
	)

	query := `
		SELECT id, business_id, business_name, overall_score, overall_level,
		       category_scores, factor_scores, recommendations, alerts,
		       assessment_method, source, metadata, assessed_at, valid_until,
		       created_at, updated_at
		FROM risk_assessments
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var storage RiskDataStorage
	var categoryScoresStr, factorScoresStr, recommendationsStr, alertsStr, metadataStr string

	err := row.Scan(
		&storage.ID,
		&storage.BusinessID,
		&storage.BusinessName,
		&storage.OverallScore,
		&storage.OverallLevel,
		&categoryScoresStr,
		&factorScoresStr,
		&recommendationsStr,
		&alertsStr,
		&storage.AssessmentMethod,
		&storage.Source,
		&metadataStr,
		&storage.AssessedAt,
		&storage.ValidUntil,
		&storage.CreatedAt,
		&storage.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("Risk assessment not found",
				zap.String("request_id", requestID),
				zap.String("assessment_id", id),
			)
			return nil, fmt.Errorf("risk assessment not found")
		}
		s.logger.Error("Failed to retrieve risk assessment",
			zap.String("request_id", requestID),
			zap.String("assessment_id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve risk assessment: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(categoryScoresStr), &storage.CategoryScores); err != nil {
		storage.CategoryScores = make(map[string]interface{})
	}
	if err := json.Unmarshal([]byte(factorScoresStr), &storage.FactorScores); err != nil {
		storage.FactorScores = []RiskScore{}
	}
	if err := json.Unmarshal([]byte(recommendationsStr), &storage.Recommendations); err != nil {
		storage.Recommendations = []RiskRecommendation{}
	}
	if err := json.Unmarshal([]byte(alertsStr), &storage.Alerts); err != nil {
		storage.Alerts = []RiskAlert{}
	}
	if err := json.Unmarshal([]byte(metadataStr), &storage.Metadata); err != nil {
		storage.Metadata = make(map[string]interface{})
	}

	// Convert to RiskAssessment
	assessment := s.convertStorageToAssessment(&storage)

	s.logger.Info("Risk assessment retrieved successfully",
		zap.String("request_id", requestID),
		zap.String("assessment_id", id),
		zap.String("business_id", assessment.BusinessID),
	)

	return assessment, nil
}

// GetRiskAssessmentsByBusinessID retrieves all risk assessments for a business
func (s *RiskStorageService) GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string, limit, offset int) ([]*RiskAssessment, error) {
	requestID := "unknown"
	if rid := ctx.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID = str
		}
	}

	s.logger.Info("Retrieving risk assessments for business",
		zap.String("request_id", requestID),
		zap.String("business_id", businessID),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	query := `
		SELECT id, business_id, business_name, overall_score, overall_level,
		       category_scores, factor_scores, recommendations, alerts,
		       assessment_method, source, metadata, assessed_at, valid_until,
		       created_at, updated_at
		FROM risk_assessments
		WHERE business_id = $1
		ORDER BY assessed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, businessID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to query risk assessments",
			zap.String("request_id", requestID),
			zap.String("business_id", businessID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to query risk assessments: %w", err)
	}
	defer rows.Close()

	var assessments []*RiskAssessment
	for rows.Next() {
		var storage RiskDataStorage
		var categoryScoresStr, factorScoresStr, recommendationsStr, alertsStr, metadataStr string

		err := rows.Scan(
			&storage.ID,
			&storage.BusinessID,
			&storage.BusinessName,
			&storage.OverallScore,
			&storage.OverallLevel,
			&categoryScoresStr,
			&factorScoresStr,
			&recommendationsStr,
			&alertsStr,
			&storage.AssessmentMethod,
			&storage.Source,
			&metadataStr,
			&storage.AssessedAt,
			&storage.ValidUntil,
			&storage.CreatedAt,
			&storage.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan risk assessment",
				zap.String("request_id", requestID),
				zap.String("business_id", businessID),
				zap.Error(err),
			)
			continue
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(categoryScoresStr), &storage.CategoryScores); err != nil {
			storage.CategoryScores = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(factorScoresStr), &storage.FactorScores); err != nil {
			storage.FactorScores = []RiskScore{}
		}
		if err := json.Unmarshal([]byte(recommendationsStr), &storage.Recommendations); err != nil {
			storage.Recommendations = []RiskRecommendation{}
		}
		if err := json.Unmarshal([]byte(alertsStr), &storage.Alerts); err != nil {
			storage.Alerts = []RiskAlert{}
		}
		if err := json.Unmarshal([]byte(metadataStr), &storage.Metadata); err != nil {
			storage.Metadata = make(map[string]interface{})
		}

		// Convert to RiskAssessment
		assessment := s.convertStorageToAssessment(&storage)
		assessments = append(assessments, assessment)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating risk assessments",
			zap.String("request_id", requestID),
			zap.String("business_id", businessID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("error iterating risk assessments: %w", err)
	}

	s.logger.Info("Risk assessments retrieved successfully",
		zap.String("request_id", requestID),
		zap.String("business_id", businessID),
		zap.Int("count", len(assessments)),
	)

	return assessments, nil
}

// UpdateRiskAssessment updates an existing risk assessment
func (s *RiskStorageService) UpdateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error {
	requestID := "unknown"
	if rid := ctx.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID = str
		}
	}

	s.logger.Info("Updating risk assessment",
		zap.String("request_id", requestID),
		zap.String("business_id", assessment.BusinessID),
		zap.String("assessment_id", assessment.ID),
	)

	// Convert data to JSON
	categoryScoresJSON, err := json.Marshal(assessment.CategoryScores)
	if err != nil {
		return fmt.Errorf("failed to marshal category scores: %w", err)
	}

	factorScoresJSON, err := json.Marshal(assessment.FactorScores)
	if err != nil {
		return fmt.Errorf("failed to marshal factor scores: %w", err)
	}

	recommendationsJSON, err := json.Marshal(assessment.Recommendations)
	if err != nil {
		return fmt.Errorf("failed to marshal recommendations: %w", err)
	}

	alertsJSON, err := json.Marshal(assessment.Alerts)
	if err != nil {
		return fmt.Errorf("failed to marshal alerts: %w", err)
	}

	metadataJSON, err := json.Marshal(assessment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE risk_assessments SET
			business_name = $2,
			overall_score = $3,
			overall_level = $4,
			category_scores = $5,
			factor_scores = $6,
			recommendations = $7,
			alerts = $8,
			metadata = $9,
			assessed_at = $10,
			valid_until = $11,
			updated_at = $12
		WHERE id = $1
	`

	now := time.Now()
	result, err := s.db.ExecContext(ctx, query,
		assessment.ID,
		assessment.BusinessName,
		assessment.OverallScore,
		string(assessment.OverallLevel),
		categoryScoresJSON,
		factorScoresJSON,
		recommendationsJSON,
		alertsJSON,
		metadataJSON,
		assessment.AssessedAt,
		assessment.ValidUntil,
		now,
	)

	if err != nil {
		s.logger.Error("Failed to update risk assessment",
			zap.String("request_id", requestID),
			zap.String("business_id", assessment.BusinessID),
			zap.String("assessment_id", assessment.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update risk assessment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		s.logger.Warn("No risk assessment found to update",
			zap.String("request_id", requestID),
			zap.String("assessment_id", assessment.ID),
		)
		return fmt.Errorf("risk assessment not found")
	}

	s.logger.Info("Risk assessment updated successfully",
		zap.String("request_id", requestID),
		zap.String("business_id", assessment.BusinessID),
		zap.String("assessment_id", assessment.ID),
	)

	return nil
}

// DeleteRiskAssessment deletes a risk assessment
func (s *RiskStorageService) DeleteRiskAssessment(ctx context.Context, id string) error {
	requestID := "unknown"
	if rid := ctx.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID = str
		}
	}

	s.logger.Info("Deleting risk assessment",
		zap.String("request_id", requestID),
		zap.String("assessment_id", id),
	)

	query := `DELETE FROM risk_assessments WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		s.logger.Error("Failed to delete risk assessment",
			zap.String("request_id", requestID),
			zap.String("assessment_id", id),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete risk assessment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		s.logger.Warn("No risk assessment found to delete",
			zap.String("request_id", requestID),
			zap.String("assessment_id", id),
		)
		return fmt.Errorf("risk assessment not found")
	}

	s.logger.Info("Risk assessment deleted successfully",
		zap.String("request_id", requestID),
		zap.String("assessment_id", id),
	)

	return nil
}

// convertStorageToAssessment converts storage model to RiskAssessment
func (s *RiskStorageService) convertStorageToAssessment(storage *RiskDataStorage) *RiskAssessment {
	// Convert category scores from map[string]interface{} to map[RiskCategory]RiskScore
	categoryScores := make(map[RiskCategory]RiskScore)
	for key, value := range storage.CategoryScores {
		if scoreData, ok := value.(map[string]interface{}); ok {
			score := RiskScore{
				FactorID:     getString(scoreData, "factor_id"),
				FactorName:   getString(scoreData, "factor_name"),
				Category:     RiskCategory(key),
				Score:        getFloat64(scoreData, "score"),
				Level:        RiskLevel(getString(scoreData, "level")),
				Confidence:   getFloat64(scoreData, "confidence"),
				Explanation:  getString(scoreData, "explanation"),
				CalculatedAt: storage.AssessedAt,
			}
			if evidence, ok := scoreData["evidence"].([]interface{}); ok {
				for _, e := range evidence {
					if str, ok := e.(string); ok {
						score.Evidence = append(score.Evidence, str)
					}
				}
			}
			categoryScores[RiskCategory(key)] = score
		}
	}

	return &RiskAssessment{
		ID:              storage.ID,
		BusinessID:      storage.BusinessID,
		BusinessName:    storage.BusinessName,
		OverallScore:    storage.OverallScore,
		OverallLevel:    RiskLevel(storage.OverallLevel),
		CategoryScores:  categoryScores,
		FactorScores:    storage.FactorScores,
		Recommendations: storage.Recommendations,
		Alerts:          storage.Alerts,
		AssessedAt:      storage.AssessedAt,
		ValidUntil:      storage.ValidUntil,
		Metadata:        storage.Metadata,
	}
}

// Helper functions for type conversion
func getString(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func getFloat64(data map[string]interface{}, key string) float64 {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0.0
}
