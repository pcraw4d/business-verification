package custom

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// CustomModelRepository defines the interface for custom model data access
type CustomModelRepository interface {
	// SaveCustomModel saves a custom risk model
	SaveCustomModel(ctx context.Context, model *CustomRiskModel) error

	// GetCustomModel retrieves a custom risk model by ID
	GetCustomModel(ctx context.Context, tenantID, modelID string) (*CustomRiskModel, error)

	// ListCustomModels lists custom models for a tenant
	ListCustomModels(ctx context.Context, tenantID string, limit, offset int) ([]*CustomRiskModel, error)

	// DeleteCustomModel deletes a custom risk model
	DeleteCustomModel(ctx context.Context, tenantID, modelID string) error

	// GetCustomModelVersions gets all versions of a custom model
	GetCustomModelVersions(ctx context.Context, tenantID, modelID string) ([]*CustomRiskModel, error)

	// ActivateCustomModel activates a custom model
	ActivateCustomModel(ctx context.Context, tenantID, modelID string) error

	// DeactivateCustomModel deactivates a custom model
	DeactivateCustomModel(ctx context.Context, tenantID, modelID string) error
}

// SQLCustomModelRepository implements CustomModelRepository using SQL database
type SQLCustomModelRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLCustomModelRepository creates a new SQL custom model repository
func NewSQLCustomModelRepository(db *sql.DB, logger *zap.Logger) *SQLCustomModelRepository {
	return &SQLCustomModelRepository{
		db:     db,
		logger: logger,
	}
}

// SaveCustomModel saves a custom risk model to the database
func (r *SQLCustomModelRepository) SaveCustomModel(ctx context.Context, model *CustomRiskModel) error {
	r.logger.Info("Saving custom risk model",
		zap.String("model_id", model.ID),
		zap.String("tenant_id", model.TenantID),
		zap.String("name", model.Name))

	// Convert custom factors to JSON
	customFactorsJSON, err := json.Marshal(model.CustomFactors)
	if err != nil {
		return fmt.Errorf("failed to marshal custom factors: %w", err)
	}

	// Convert factor weights to JSON
	factorWeightsJSON, err := json.Marshal(model.FactorWeights)
	if err != nil {
		return fmt.Errorf("failed to marshal factor weights: %w", err)
	}

	// Convert thresholds to JSON
	thresholdsJSON, err := json.Marshal(model.Thresholds)
	if err != nil {
		return fmt.Errorf("failed to marshal thresholds: %w", err)
	}

	// Convert validation rules to JSON
	validationRulesJSON, err := json.Marshal(model.ValidationRules)
	if err != nil {
		return fmt.Errorf("failed to marshal validation rules: %w", err)
	}

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(model.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check if model exists
	existingModel, err := r.GetCustomModel(ctx, model.TenantID, model.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing model: %w", err)
	}

	if existingModel != nil {
		// Update existing model
		query := `
			UPDATE custom_risk_models SET
				name = $1, description = $2, base_model = $3, custom_factors = $4,
				factor_weights = $5, thresholds = $6, validation_rules = $7,
				is_active = $8, version = $9, updated_at = $10, updated_by = $11, metadata = $12
			WHERE id = $13 AND tenant_id = $14
		`
		_, err = r.db.ExecContext(ctx, query,
			model.Name, model.Description, model.BaseModel, customFactorsJSON,
			factorWeightsJSON, thresholdsJSON, validationRulesJSON,
			model.IsActive, model.Version, model.UpdatedAt, model.UpdatedBy, metadataJSON,
			model.ID, model.TenantID)
	} else {
		// Insert new model
		query := `
			INSERT INTO custom_risk_models (
				id, tenant_id, name, description, base_model, custom_factors,
				factor_weights, thresholds, validation_rules, is_active, version,
				created_at, updated_at, created_by, updated_by, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
			)
		`
		_, err = r.db.ExecContext(ctx, query,
			model.ID, model.TenantID, model.Name, model.Description, model.BaseModel,
			customFactorsJSON, factorWeightsJSON, thresholdsJSON, validationRulesJSON,
			model.IsActive, model.Version, model.CreatedAt, model.UpdatedAt,
			model.CreatedBy, model.UpdatedBy, metadataJSON)
	}

	if err != nil {
		return fmt.Errorf("failed to save custom model: %w", err)
	}

	r.logger.Info("Custom risk model saved successfully",
		zap.String("model_id", model.ID),
		zap.String("tenant_id", model.TenantID))

	return nil
}

// GetCustomModel retrieves a custom risk model by ID
func (r *SQLCustomModelRepository) GetCustomModel(ctx context.Context, tenantID, modelID string) (*CustomRiskModel, error) {
	r.logger.Debug("Getting custom risk model",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, name, description, base_model, custom_factors,
			   factor_weights, thresholds, validation_rules, is_active, version,
			   created_at, updated_at, created_by, updated_by, metadata
		FROM custom_risk_models
		WHERE id = $1 AND tenant_id = $2
	`

	var result struct {
		ID              string    `json:"id"`
		TenantID        string    `json:"tenant_id"`
		Name            string    `json:"name"`
		Description     string    `json:"description"`
		BaseModel       string    `json:"base_model"`
		CustomFactors   string    `json:"custom_factors"`
		FactorWeights   string    `json:"factor_weights"`
		Thresholds      string    `json:"thresholds"`
		ValidationRules string    `json:"validation_rules"`
		IsActive        bool      `json:"is_active"`
		Version         int       `json:"version"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
		CreatedBy       string    `json:"created_by"`
		UpdatedBy       string    `json:"updated_by"`
		Metadata        string    `json:"metadata"`
	}

	err := r.db.QueryRowContext(ctx, query, modelID, tenantID).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.Description, &result.BaseModel,
		&result.CustomFactors, &result.FactorWeights, &result.Thresholds, &result.ValidationRules,
		&result.IsActive, &result.Version, &result.CreatedAt, &result.UpdatedAt,
		&result.CreatedBy, &result.UpdatedBy, &result.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Model not found
		}
		return nil, fmt.Errorf("failed to get custom model: %w", err)
	}

	// Convert the result to CustomRiskModel
	model, err := r.convertToCustomRiskModel(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Custom risk model retrieved successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	return model, nil
}

// ListCustomModels lists custom models for a tenant
func (r *SQLCustomModelRepository) ListCustomModels(ctx context.Context, tenantID string, limit, offset int) ([]*CustomRiskModel, error) {
	r.logger.Debug("Listing custom risk models",
		zap.String("tenant_id", tenantID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	query := `
		SELECT id, tenant_id, name, description, base_model, custom_factors,
			   factor_weights, thresholds, validation_rules, is_active, version,
			   created_at, updated_at, created_by, updated_by, metadata
		FROM custom_risk_models
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom models: %w", err)
	}
	defer rows.Close()

	var models []*CustomRiskModel
	for rows.Next() {
		var result struct {
			ID              string    `json:"id"`
			TenantID        string    `json:"tenant_id"`
			Name            string    `json:"name"`
			Description     string    `json:"description"`
			BaseModel       string    `json:"base_model"`
			CustomFactors   string    `json:"custom_factors"`
			FactorWeights   string    `json:"factor_weights"`
			Thresholds      string    `json:"thresholds"`
			ValidationRules string    `json:"validation_rules"`
			IsActive        bool      `json:"is_active"`
			Version         int       `json:"version"`
			CreatedAt       time.Time `json:"created_at"`
			UpdatedAt       time.Time `json:"updated_at"`
			CreatedBy       string    `json:"created_by"`
			UpdatedBy       string    `json:"updated_by"`
			Metadata        string    `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.Description, &result.BaseModel,
			&result.CustomFactors, &result.FactorWeights, &result.Thresholds, &result.ValidationRules,
			&result.IsActive, &result.Version, &result.CreatedAt, &result.UpdatedAt,
			&result.CreatedBy, &result.UpdatedBy, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		model, err := r.convertToCustomRiskModel(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		models = append(models, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Custom risk models listed successfully",
		zap.String("tenant_id", tenantID),
		zap.Int("count", len(models)))

	return models, nil
}

// DeleteCustomModel deletes a custom risk model
func (r *SQLCustomModelRepository) DeleteCustomModel(ctx context.Context, tenantID, modelID string) error {
	r.logger.Info("Deleting custom risk model",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	query := `DELETE FROM custom_risk_models WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.ExecContext(ctx, query, modelID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to delete custom model: %w", err)
	}

	r.logger.Info("Custom risk model deleted successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetCustomModelVersions gets all versions of a custom model
func (r *SQLCustomModelRepository) GetCustomModelVersions(ctx context.Context, tenantID, modelID string) ([]*CustomRiskModel, error) {
	r.logger.Debug("Getting custom risk model versions",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	// For now, return empty slice - versioning can be implemented later
	return []*CustomRiskModel{}, nil
}

// ActivateCustomModel activates a custom model
func (r *SQLCustomModelRepository) ActivateCustomModel(ctx context.Context, tenantID, modelID string) error {
	r.logger.Info("Activating custom risk model",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	query := `UPDATE custom_risk_models SET is_active = true, updated_at = $1 WHERE id = $2 AND tenant_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), modelID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to activate custom model: %w", err)
	}

	r.logger.Info("Custom risk model activated successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	return nil
}

// DeactivateCustomModel deactivates a custom model
func (r *SQLCustomModelRepository) DeactivateCustomModel(ctx context.Context, tenantID, modelID string) error {
	r.logger.Info("Deactivating custom risk model",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	query := `UPDATE custom_risk_models SET is_active = false, updated_at = $1 WHERE id = $2 AND tenant_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), modelID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to deactivate custom model: %w", err)
	}

	r.logger.Info("Custom risk model deactivated successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))

	return nil
}

// convertToCustomRiskModel converts a database result to CustomRiskModel
func (r *SQLCustomModelRepository) convertToCustomRiskModel(result interface{}) (*CustomRiskModel, error) {
	// Type assertion to get the fields
	var id, tenantID, name, description, baseModel, customFactors, factorWeights, thresholds, validationRules, createdBy, updatedBy, metadata string
	var isActive bool
	var version int
	var createdAt, updatedAt time.Time

	// Use reflection or type assertion based on the actual structure
	// For now, assuming we have access to the fields directly
	switch v := result.(type) {
	case *struct {
		ID              string    `json:"id"`
		TenantID        string    `json:"tenant_id"`
		Name            string    `json:"name"`
		Description     string    `json:"description"`
		BaseModel       string    `json:"base_model"`
		CustomFactors   string    `json:"custom_factors"`
		FactorWeights   string    `json:"factor_weights"`
		Thresholds      string    `json:"thresholds"`
		ValidationRules string    `json:"validation_rules"`
		IsActive        bool      `json:"is_active"`
		Version         int       `json:"version"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
		CreatedBy       string    `json:"created_by"`
		UpdatedBy       string    `json:"updated_by"`
		Metadata        string    `json:"metadata"`
	}:
		id = v.ID
		tenantID = v.TenantID
		name = v.Name
		description = v.Description
		baseModel = v.BaseModel
		customFactors = v.CustomFactors
		factorWeights = v.FactorWeights
		thresholds = v.Thresholds
		validationRules = v.ValidationRules
		isActive = v.IsActive
		version = v.Version
		createdAt = v.CreatedAt
		updatedAt = v.UpdatedAt
		createdBy = v.CreatedBy
		updatedBy = v.UpdatedBy
		metadata = v.Metadata
	default:
		return nil, fmt.Errorf("unsupported result type: %T", result)
	}

	// Parse custom factors
	var customFactorsList []CustomRiskFactor
	if customFactors != "" {
		if err := json.Unmarshal([]byte(customFactors), &customFactorsList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom factors: %w", err)
		}
	}

	// Parse factor weights
	var factorWeightsMap map[string]float64
	if factorWeights != "" {
		if err := json.Unmarshal([]byte(factorWeights), &factorWeightsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal factor weights: %w", err)
		}
	}

	// Parse thresholds
	var thresholdsMap map[string]float64
	if thresholds != "" {
		if err := json.Unmarshal([]byte(thresholds), &thresholdsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal thresholds: %w", err)
		}
	}

	// Convert string keys to RiskLevel
	riskLevelThresholds := make(map[models.RiskLevel]float64)
	for key, value := range thresholdsMap {
		riskLevelThresholds[models.RiskLevel(key)] = value
	}

	// Parse validation rules
	var validationRulesList []ValidationRule
	if validationRules != "" {
		if err := json.Unmarshal([]byte(validationRules), &validationRulesList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validation rules: %w", err)
		}
	}

	// Parse metadata
	var metadataMap map[string]interface{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	model := &CustomRiskModel{
		ID:              id,
		TenantID:        tenantID,
		Name:            name,
		Description:     description,
		BaseModel:       baseModel,
		CustomFactors:   customFactorsList,
		FactorWeights:   factorWeightsMap,
		Thresholds:      riskLevelThresholds,
		ValidationRules: validationRulesList,
		IsActive:        isActive,
		Version:         version,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
		Metadata:        metadataMap,
	}

	return model, nil
}
