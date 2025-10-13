package custom

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/industry"
)

// CustomModelRepository interface defines methods for custom model storage
type CustomModelRepository interface {
	SaveCustomModel(ctx context.Context, model *CustomRiskModel) error
	GetCustomModel(ctx context.Context, modelID string) (*CustomRiskModel, error)
	ListCustomModels(ctx context.Context, tenantID string, limit, offset int) ([]*CustomRiskModel, error)
	DeleteCustomModel(ctx context.Context, modelID string) error
	GetActiveCustomModel(ctx context.Context, tenantID string, baseModel string) (*CustomRiskModel, error)
	CountCustomModels(ctx context.Context, tenantID string) (int, error)
	UpdateModelStatus(ctx context.Context, modelID string, isActive bool) error
	GetModelVersions(ctx context.Context, modelID string) ([]*CustomRiskModel, error)
}

// SQLCustomModelRepository implements CustomModelRepository using SQL database
type SQLCustomModelRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewCustomModelRepository creates a new custom model repository
func NewCustomModelRepository(db *sql.DB, logger *zap.Logger) CustomModelRepository {
	return &SQLCustomModelRepository{
		db:     db,
		logger: logger,
	}
}

// NewSQLCustomModelRepository creates a new SQL custom model repository
func NewSQLCustomModelRepository(db *sql.DB, logger *zap.Logger) *SQLCustomModelRepository {
	return &SQLCustomModelRepository{
		db:     db,
		logger: logger,
	}
}

// SaveCustomModel saves a custom risk model to the database
func (cmr *SQLCustomModelRepository) SaveCustomModel(ctx context.Context, model *CustomRiskModel) error {
	cmr.logger.Info("Saving custom risk model",
		zap.String("model_id", model.ID),
		zap.String("tenant_id", model.TenantID))

	// Serialize custom factors
	customFactorsJSON, err := json.Marshal(model.CustomFactors)
	if err != nil {
		return fmt.Errorf("failed to serialize custom factors: %w", err)
	}

	// Serialize factor weights
	factorWeightsJSON, err := json.Marshal(model.FactorWeights)
	if err != nil {
		return fmt.Errorf("failed to serialize factor weights: %w", err)
	}

	// Serialize thresholds
	thresholdsJSON, err := json.Marshal(model.Thresholds)
	if err != nil {
		return fmt.Errorf("failed to serialize thresholds: %w", err)
	}

	// Serialize validation rules
	validationRulesJSON, err := json.Marshal(model.ValidationRules)
	if err != nil {
		return fmt.Errorf("failed to serialize validation rules: %w", err)
	}

	// Serialize metadata
	metadataJSON, err := json.Marshal(model.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	query := `
		INSERT INTO custom_risk_models (
			id, tenant_id, name, description, base_model, custom_factors,
			factor_weights, thresholds, validation_rules, is_active, version,
			created_at, updated_at, created_by, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			base_model = EXCLUDED.base_model,
			custom_factors = EXCLUDED.custom_factors,
			factor_weights = EXCLUDED.factor_weights,
			thresholds = EXCLUDED.thresholds,
			validation_rules = EXCLUDED.validation_rules,
			is_active = EXCLUDED.is_active,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at,
			created_by = EXCLUDED.created_by,
			metadata = EXCLUDED.metadata
	`

	_, err = cmr.db.ExecContext(ctx, query,
		model.ID,
		model.TenantID,
		model.Name,
		model.Description,
		string(model.BaseModel),
		customFactorsJSON,
		factorWeightsJSON,
		thresholdsJSON,
		validationRulesJSON,
		model.IsActive,
		model.Version,
		model.CreatedAt,
		model.UpdatedAt,
		model.CreatedBy,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to save custom model: %w", err)
	}

	cmr.logger.Info("Custom risk model saved successfully",
		zap.String("model_id", model.ID))

	return nil
}

// GetCustomModel retrieves a custom risk model by ID
func (cmr *SQLCustomModelRepository) GetCustomModel(ctx context.Context, modelID string) (*CustomRiskModel, error) {
	cmr.logger.Info("Retrieving custom risk model",
		zap.String("model_id", modelID))

	query := `
		SELECT id, tenant_id, name, description, base_model, custom_factors,
			   factor_weights, thresholds, validation_rules, is_active, version,
			   created_at, updated_at, created_by, metadata
		FROM custom_risk_models
		WHERE id = $1
	`

	row := cmr.db.QueryRowContext(ctx, query, modelID)

	var model CustomRiskModel
	var baseModelStr string
	var customFactorsJSON, factorWeightsJSON, thresholdsJSON, validationRulesJSON, metadataJSON []byte

	err := row.Scan(
		&model.ID,
		&model.TenantID,
		&model.Name,
		&model.Description,
		&baseModelStr,
		&customFactorsJSON,
		&factorWeightsJSON,
		&thresholdsJSON,
		&validationRulesJSON,
		&model.IsActive,
		&model.Version,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.CreatedBy,
		&metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("custom model not found: %s", modelID)
		}
		return nil, fmt.Errorf("failed to retrieve custom model: %w", err)
	}

	// Deserialize the JSON fields
	model.BaseModel = industry.IndustryType(baseModelStr)

	if err := json.Unmarshal(customFactorsJSON, &model.CustomFactors); err != nil {
		return nil, fmt.Errorf("failed to deserialize custom factors: %w", err)
	}

	if err := json.Unmarshal(factorWeightsJSON, &model.FactorWeights); err != nil {
		return nil, fmt.Errorf("failed to deserialize factor weights: %w", err)
	}

	if err := json.Unmarshal(thresholdsJSON, &model.Thresholds); err != nil {
		return nil, fmt.Errorf("failed to deserialize thresholds: %w", err)
	}

	if err := json.Unmarshal(validationRulesJSON, &model.ValidationRules); err != nil {
		return nil, fmt.Errorf("failed to deserialize validation rules: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &model.Metadata); err != nil {
		return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
	}

	cmr.logger.Info("Custom risk model retrieved successfully",
		zap.String("model_id", model.ID))

	return &model, nil
}

// ListCustomModels retrieves custom risk models for a tenant
func (cmr *SQLCustomModelRepository) ListCustomModels(ctx context.Context, tenantID string, limit, offset int) ([]*CustomRiskModel, error) {
	cmr.logger.Info("Listing custom risk models",
		zap.String("tenant_id", tenantID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	query := `
		SELECT id, tenant_id, name, description, base_model, custom_factors,
			   factor_weights, thresholds, validation_rules, is_active, version,
			   created_at, updated_at, created_by, metadata
		FROM custom_risk_models
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := cmr.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom models: %w", err)
	}
	defer rows.Close()

	var models []*CustomRiskModel
	for rows.Next() {
		var model CustomRiskModel
		var baseModelStr string
		var customFactorsJSON, factorWeightsJSON, thresholdsJSON, validationRulesJSON, metadataJSON []byte

		err := rows.Scan(
			&model.ID,
			&model.TenantID,
			&model.Name,
			&model.Description,
			&baseModelStr,
			&customFactorsJSON,
			&factorWeightsJSON,
			&thresholdsJSON,
			&validationRulesJSON,
			&model.IsActive,
			&model.Version,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.CreatedBy,
			&metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan custom model: %w", err)
		}

		// Deserialize the JSON fields
		model.BaseModel = industry.IndustryType(baseModelStr)

		if err := json.Unmarshal(customFactorsJSON, &model.CustomFactors); err != nil {
			return nil, fmt.Errorf("failed to deserialize custom factors: %w", err)
		}

		if err := json.Unmarshal(factorWeightsJSON, &model.FactorWeights); err != nil {
			return nil, fmt.Errorf("failed to deserialize factor weights: %w", err)
		}

		if err := json.Unmarshal(thresholdsJSON, &model.Thresholds); err != nil {
			return nil, fmt.Errorf("failed to deserialize thresholds: %w", err)
		}

		if err := json.Unmarshal(validationRulesJSON, &model.ValidationRules); err != nil {
			return nil, fmt.Errorf("failed to deserialize validation rules: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &model.Metadata); err != nil {
			return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
		}

		models = append(models, &model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating custom models: %w", err)
	}

	cmr.logger.Info("Custom risk models listed successfully",
		zap.String("tenant_id", tenantID),
		zap.Int("count", len(models)))

	return models, nil
}

// DeleteCustomModel deletes a custom risk model
func (cmr *SQLCustomModelRepository) DeleteCustomModel(ctx context.Context, modelID string) error {
	cmr.logger.Info("Deleting custom risk model",
		zap.String("model_id", modelID))

	query := `DELETE FROM custom_risk_models WHERE id = $1`

	result, err := cmr.db.ExecContext(ctx, query, modelID)
	if err != nil {
		return fmt.Errorf("failed to delete custom model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("custom model not found: %s", modelID)
	}

	cmr.logger.Info("Custom risk model deleted successfully",
		zap.String("model_id", modelID))

	return nil
}

// GetActiveCustomModel retrieves the active custom model for a tenant and base model
func (cmr *SQLCustomModelRepository) GetActiveCustomModel(ctx context.Context, tenantID string, baseModel string) (*CustomRiskModel, error) {
	cmr.logger.Info("Retrieving active custom risk model",
		zap.String("tenant_id", tenantID),
		zap.String("base_model", baseModel))

	query := `
		SELECT id, tenant_id, name, description, base_model, custom_factors,
			   factor_weights, thresholds, validation_rules, is_active, version,
			   created_at, updated_at, created_by, metadata
		FROM custom_risk_models
		WHERE tenant_id = $1 AND base_model = $2 AND is_active = true
		ORDER BY updated_at DESC
		LIMIT 1
	`

	row := cmr.db.QueryRowContext(ctx, query, tenantID, baseModel)

	var model CustomRiskModel
	var baseModelStr string
	var customFactorsJSON, factorWeightsJSON, thresholdsJSON, validationRulesJSON, metadataJSON []byte

	err := row.Scan(
		&model.ID,
		&model.TenantID,
		&model.Name,
		&model.Description,
		&baseModelStr,
		&customFactorsJSON,
		&factorWeightsJSON,
		&thresholdsJSON,
		&validationRulesJSON,
		&model.IsActive,
		&model.Version,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.CreatedBy,
		&metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active custom model found for tenant %s and base model %s", tenantID, baseModel)
		}
		return nil, fmt.Errorf("failed to retrieve active custom model: %w", err)
	}

	// Deserialize the JSON fields
	model.BaseModel = industry.IndustryType(baseModelStr)

	if err := json.Unmarshal(customFactorsJSON, &model.CustomFactors); err != nil {
		return nil, fmt.Errorf("failed to deserialize custom factors: %w", err)
	}

	if err := json.Unmarshal(factorWeightsJSON, &model.FactorWeights); err != nil {
		return nil, fmt.Errorf("failed to deserialize factor weights: %w", err)
	}

	if err := json.Unmarshal(thresholdsJSON, &model.Thresholds); err != nil {
		return nil, fmt.Errorf("failed to deserialize thresholds: %w", err)
	}

	if err := json.Unmarshal(validationRulesJSON, &model.ValidationRules); err != nil {
		return nil, fmt.Errorf("failed to deserialize validation rules: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &model.Metadata); err != nil {
		return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
	}

	cmr.logger.Info("Active custom risk model retrieved successfully",
		zap.String("model_id", model.ID))

	return &model, nil
}

// CountCustomModels counts the number of custom models for a tenant
func (cmr *SQLCustomModelRepository) CountCustomModels(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COUNT(*) FROM custom_risk_models WHERE tenant_id = $1`

	var count int
	err := cmr.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count custom models: %w", err)
	}

	return count, nil
}

// UpdateModelStatus updates the active status of a custom model
func (cmr *SQLCustomModelRepository) UpdateModelStatus(ctx context.Context, modelID string, isActive bool) error {
	cmr.logger.Info("Updating custom model status",
		zap.String("model_id", modelID),
		zap.Bool("is_active", isActive))

	query := `
		UPDATE custom_risk_models 
		SET is_active = $1, updated_at = $2 
		WHERE id = $3
	`

	result, err := cmr.db.ExecContext(ctx, query, isActive, time.Now(), modelID)
	if err != nil {
		return fmt.Errorf("failed to update model status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("custom model not found: %s", modelID)
	}

	cmr.logger.Info("Custom model status updated successfully",
		zap.String("model_id", modelID))

	return nil
}

// GetModelVersions retrieves all versions of a custom model
func (cmr *SQLCustomModelRepository) GetModelVersions(ctx context.Context, modelID string) ([]*CustomRiskModel, error) {
	cmr.logger.Info("Retrieving model versions",
		zap.String("model_id", modelID))

	query := `
		SELECT id, tenant_id, name, description, base_model, custom_factors,
			   factor_weights, thresholds, validation_rules, is_active, version,
			   created_at, updated_at, created_by, metadata
		FROM custom_risk_models
		WHERE id = $1
		ORDER BY created_at DESC
	`

	rows, err := cmr.db.QueryContext(ctx, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve model versions: %w", err)
	}
	defer rows.Close()

	var models []*CustomRiskModel
	for rows.Next() {
		var model CustomRiskModel
		var baseModelStr string
		var customFactorsJSON, factorWeightsJSON, thresholdsJSON, validationRulesJSON, metadataJSON []byte

		err := rows.Scan(
			&model.ID,
			&model.TenantID,
			&model.Name,
			&model.Description,
			&baseModelStr,
			&customFactorsJSON,
			&factorWeightsJSON,
			&thresholdsJSON,
			&validationRulesJSON,
			&model.IsActive,
			&model.Version,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.CreatedBy,
			&metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan model version: %w", err)
		}

		// Deserialize the JSON fields
		model.BaseModel = industry.IndustryType(baseModelStr)

		if err := json.Unmarshal(customFactorsJSON, &model.CustomFactors); err != nil {
			return nil, fmt.Errorf("failed to deserialize custom factors: %w", err)
		}

		if err := json.Unmarshal(factorWeightsJSON, &model.FactorWeights); err != nil {
			return nil, fmt.Errorf("failed to deserialize factor weights: %w", err)
		}

		if err := json.Unmarshal(thresholdsJSON, &model.Thresholds); err != nil {
			return nil, fmt.Errorf("failed to deserialize thresholds: %w", err)
		}

		if err := json.Unmarshal(validationRulesJSON, &model.ValidationRules); err != nil {
			return nil, fmt.Errorf("failed to deserialize validation rules: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &model.Metadata); err != nil {
			return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
		}

		models = append(models, &model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating model versions: %w", err)
	}

	cmr.logger.Info("Model versions retrieved successfully",
		zap.String("model_id", modelID),
		zap.Int("version_count", len(models)))

	return models, nil
}
