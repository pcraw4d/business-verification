package industry_codes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// CodeDescription represents detailed description information for an industry code
type CodeDescription struct {
	ID               string    `json:"id"`
	CodeID           string    `json:"code_id"`
	ShortDescription string    `json:"short_description"`
	LongDescription  string    `json:"long_description"`
	Examples         []string  `json:"examples,omitempty"`
	Exclusions       []string  `json:"exclusions,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	LastUpdated      time.Time `json:"last_updated"`
	Source           string    `json:"source"`
	Version          string    `json:"version"`
}

// CodeMetadata represents comprehensive metadata for an industry code
type CodeMetadata struct {
	ID               string            `json:"id"`
	CodeID           string            `json:"code_id"`
	Version          string            `json:"version"`
	EffectiveDate    time.Time         `json:"effective_date"`
	ExpirationDate   *time.Time        `json:"expiration_date,omitempty"`
	Source           string            `json:"source"`
	SourceURL        string            `json:"source_url,omitempty"`
	LastUpdated      time.Time         `json:"last_updated"`
	UpdateFrequency  string            `json:"update_frequency"`
	DataQuality      string            `json:"data_quality"`
	Confidence       float64           `json:"confidence"`
	UsageCount       int64             `json:"usage_count"`
	Popularity       string            `json:"popularity"`
	Tags             []string          `json:"tags,omitempty"`
	CustomFields     map[string]string `json:"custom_fields,omitempty"`
	ValidationStatus string            `json:"validation_status"`
	ValidationNotes  string            `json:"validation_notes,omitempty"`
}

// CodeRelationship represents relationships between industry codes
type CodeRelationship struct {
	ID               string    `json:"id"`
	SourceCodeID     string    `json:"source_code_id"`
	TargetCodeID     string    `json:"target_code_id"`
	RelationshipType string    `json:"relationship_type"`
	Confidence       float64   `json:"confidence"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	LastUpdated      time.Time `json:"last_updated"`
}

// RelationshipType constants
const (
	RelationshipTypeParentChild = "parent_child"
	RelationshipTypeCrosswalk   = "crosswalk"
	RelationshipTypeRelated     = "related"
	RelationshipTypeSuperseded  = "superseded"
	RelationshipTypeReplaces    = "replaces"
)

// CodeCrosswalk represents crosswalk mappings between different classification systems
type CodeCrosswalk struct {
	ID          string    `json:"id"`
	SourceCode  string    `json:"source_code"`
	SourceType  CodeType  `json:"source_type"`
	TargetCode  string    `json:"target_code"`
	TargetType  CodeType  `json:"target_type"`
	MappingType string    `json:"mapping_type"`
	Confidence  float64   `json:"confidence"`
	Direction   string    `json:"direction"` // "forward", "reverse", "bidirectional"
	Notes       string    `json:"notes,omitempty"`
	LastUpdated time.Time `json:"last_updated"`
}

// MappingType constants
const (
	MappingTypeExact       = "exact"
	MappingTypeClose       = "close"
	MappingTypePartial     = "partial"
	MappingTypeApproximate = "approximate"
)

// MetadataManager provides comprehensive metadata management for industry codes
type MetadataManager struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(db *sql.DB, logger *zap.Logger) *MetadataManager {
	return &MetadataManager{
		db:     db,
		logger: logger,
	}
}

// Initialize creates the metadata tables if they don't exist
func (mm *MetadataManager) Initialize(ctx context.Context) error {
	queries := []string{
		// Code descriptions table
		`CREATE TABLE IF NOT EXISTS code_descriptions (
			id VARCHAR(50) PRIMARY KEY,
			code_id VARCHAR(50) NOT NULL,
			short_description TEXT NOT NULL,
			long_description TEXT,
			examples TEXT,
			exclusions TEXT,
			notes TEXT,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			source VARCHAR(100) NOT NULL,
			version VARCHAR(20) NOT NULL,
			FOREIGN KEY (code_id) REFERENCES industry_codes(id) ON DELETE CASCADE,
			UNIQUE(code_id, version)
		)`,

		// Code metadata table
		`CREATE TABLE IF NOT EXISTS code_metadata (
			id VARCHAR(50) PRIMARY KEY,
			code_id VARCHAR(50) NOT NULL,
			version VARCHAR(20) NOT NULL,
			effective_date TIMESTAMP NOT NULL,
			expiration_date TIMESTAMP,
			source VARCHAR(100) NOT NULL,
			source_url TEXT,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			update_frequency VARCHAR(50),
			data_quality VARCHAR(20),
			confidence DECIMAL(3,2) DEFAULT 0.00,
			usage_count BIGINT DEFAULT 0,
			popularity VARCHAR(20),
			tags TEXT,
			custom_fields TEXT,
			validation_status VARCHAR(20) DEFAULT 'pending',
			validation_notes TEXT,
			FOREIGN KEY (code_id) REFERENCES industry_codes(id) ON DELETE CASCADE,
			UNIQUE(code_id, version)
		)`,

		// Code relationships table
		`CREATE TABLE IF NOT EXISTS code_relationships (
			id VARCHAR(50) PRIMARY KEY,
			source_code_id VARCHAR(50) NOT NULL,
			target_code_id VARCHAR(50) NOT NULL,
			relationship_type VARCHAR(20) NOT NULL,
			confidence DECIMAL(3,2) DEFAULT 0.00,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (source_code_id) REFERENCES industry_codes(id) ON DELETE CASCADE,
			FOREIGN KEY (target_code_id) REFERENCES industry_codes(id) ON DELETE CASCADE,
			UNIQUE(source_code_id, target_code_id, relationship_type)
		)`,

		// Code crosswalks table
		`CREATE TABLE IF NOT EXISTS code_crosswalks (
			id VARCHAR(50) PRIMARY KEY,
			source_code VARCHAR(20) NOT NULL,
			source_type VARCHAR(10) NOT NULL,
			target_code VARCHAR(20) NOT NULL,
			target_type VARCHAR(10) NOT NULL,
			mapping_type VARCHAR(20) NOT NULL,
			confidence DECIMAL(3,2) DEFAULT 0.00,
			direction VARCHAR(20) NOT NULL,
			notes TEXT,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source_code, source_type, target_code, target_type)
		)`,
	}

	for i, query := range queries {
		_, err := mm.db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create metadata table %d: %w", i+1, err)
		}
	}

	mm.logger.Info("Metadata tables initialized successfully")
	return nil
}

// SaveCodeDescription saves or updates a code description
func (mm *MetadataManager) SaveCodeDescription(ctx context.Context, desc *CodeDescription) error {
	query := `
		INSERT INTO code_descriptions (
			id, code_id, short_description, long_description, examples, exclusions, 
			notes, source, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (code_id, version) DO UPDATE SET
			short_description = EXCLUDED.short_description,
			long_description = EXCLUDED.long_description,
			examples = EXCLUDED.examples,
			exclusions = EXCLUDED.exclusions,
			notes = EXCLUDED.notes,
			last_updated = CURRENT_TIMESTAMP,
			source = EXCLUDED.source
	`

	examplesJSON, _ := json.Marshal(desc.Examples)
	exclusionsJSON, _ := json.Marshal(desc.Exclusions)

	_, err := mm.db.ExecContext(ctx, query,
		desc.ID, desc.CodeID, desc.ShortDescription, desc.LongDescription,
		examplesJSON, exclusionsJSON, desc.Notes, desc.Source, desc.Version)

	if err != nil {
		return fmt.Errorf("failed to save code description: %w", err)
	}

	mm.logger.Debug("Code description saved",
		zap.String("code_id", desc.CodeID),
		zap.String("version", desc.Version))
	return nil
}

// GetCodeDescription retrieves a code description by code ID and version
func (mm *MetadataManager) GetCodeDescription(ctx context.Context, codeID, version string) (*CodeDescription, error) {
	query := `
		SELECT id, code_id, short_description, long_description, examples, exclusions,
		       notes, last_updated, source, version
		FROM code_descriptions
		WHERE code_id = $1 AND version = $2
	`

	var desc CodeDescription
	var examplesJSON, exclusionsJSON []byte

	err := mm.db.QueryRowContext(ctx, query, codeID, version).Scan(
		&desc.ID, &desc.CodeID, &desc.ShortDescription, &desc.LongDescription,
		&examplesJSON, &exclusionsJSON, &desc.Notes, &desc.LastUpdated,
		&desc.Source, &desc.Version)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("code description not found: %s (version: %s)", codeID, version)
		}
		return nil, fmt.Errorf("failed to get code description: %w", err)
	}

	// Parse JSON arrays
	if len(examplesJSON) > 0 {
		json.Unmarshal(examplesJSON, &desc.Examples)
	}
	if len(exclusionsJSON) > 0 {
		json.Unmarshal(exclusionsJSON, &desc.Exclusions)
	}

	return &desc, nil
}

// GetLatestCodeDescription retrieves the latest version of a code description
func (mm *MetadataManager) GetLatestCodeDescription(ctx context.Context, codeID string) (*CodeDescription, error) {
	query := `
		SELECT id, code_id, short_description, long_description, examples, exclusions,
		       notes, last_updated, source, version
		FROM code_descriptions
		WHERE code_id = $1
		ORDER BY version DESC
		LIMIT 1
	`

	var desc CodeDescription
	var examplesJSON, exclusionsJSON []byte

	err := mm.db.QueryRowContext(ctx, query, codeID).Scan(
		&desc.ID, &desc.CodeID, &desc.ShortDescription, &desc.LongDescription,
		&examplesJSON, &exclusionsJSON, &desc.Notes, &desc.LastUpdated,
		&desc.Source, &desc.Version)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("code description not found: %s", codeID)
		}
		return nil, fmt.Errorf("failed to get latest code description: %w", err)
	}

	// Parse JSON arrays
	if len(examplesJSON) > 0 {
		json.Unmarshal(examplesJSON, &desc.Examples)
	}
	if len(exclusionsJSON) > 0 {
		json.Unmarshal(exclusionsJSON, &desc.Exclusions)
	}

	return &desc, nil
}

// SaveCodeMetadata saves or updates code metadata
func (mm *MetadataManager) SaveCodeMetadata(ctx context.Context, metadata *CodeMetadata) error {
	query := `
		INSERT INTO code_metadata (
			id, code_id, version, effective_date, expiration_date, source, source_url,
			update_frequency, data_quality, confidence, usage_count, popularity,
			tags, custom_fields, validation_status, validation_notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (code_id, version) DO UPDATE SET
			effective_date = EXCLUDED.effective_date,
			expiration_date = EXCLUDED.expiration_date,
			source = EXCLUDED.source,
			source_url = EXCLUDED.source_url,
			last_updated = CURRENT_TIMESTAMP,
			update_frequency = EXCLUDED.update_frequency,
			data_quality = EXCLUDED.data_quality,
			confidence = EXCLUDED.confidence,
			usage_count = EXCLUDED.usage_count,
			popularity = EXCLUDED.popularity,
			tags = EXCLUDED.tags,
			custom_fields = EXCLUDED.custom_fields,
			validation_status = EXCLUDED.validation_status,
			validation_notes = EXCLUDED.validation_notes
	`

	tagsJSON, _ := json.Marshal(metadata.Tags)
	customFieldsJSON, _ := json.Marshal(metadata.CustomFields)

	_, err := mm.db.ExecContext(ctx, query,
		metadata.ID, metadata.CodeID, metadata.Version, metadata.EffectiveDate,
		metadata.ExpirationDate, metadata.Source, metadata.SourceURL,
		metadata.UpdateFrequency, metadata.DataQuality, metadata.Confidence,
		metadata.UsageCount, metadata.Popularity, tagsJSON, customFieldsJSON,
		metadata.ValidationStatus, metadata.ValidationNotes)

	if err != nil {
		return fmt.Errorf("failed to save code metadata: %w", err)
	}

	mm.logger.Debug("Code metadata saved",
		zap.String("code_id", metadata.CodeID),
		zap.String("version", metadata.Version))
	return nil
}

// GetCodeMetadata retrieves code metadata by code ID and version
func (mm *MetadataManager) GetCodeMetadata(ctx context.Context, codeID, version string) (*CodeMetadata, error) {
	query := `
		SELECT id, code_id, version, effective_date, expiration_date, source, source_url,
		       last_updated, update_frequency, data_quality, confidence, usage_count,
		       popularity, tags, custom_fields, validation_status, validation_notes
		FROM code_metadata
		WHERE code_id = $1 AND version = $2
	`

	var metadata CodeMetadata
	var tagsJSON, customFieldsJSON []byte

	err := mm.db.QueryRowContext(ctx, query, codeID, version).Scan(
		&metadata.ID, &metadata.CodeID, &metadata.Version, &metadata.EffectiveDate,
		&metadata.ExpirationDate, &metadata.Source, &metadata.SourceURL,
		&metadata.LastUpdated, &metadata.UpdateFrequency, &metadata.DataQuality,
		&metadata.Confidence, &metadata.UsageCount, &metadata.Popularity,
		&tagsJSON, &customFieldsJSON, &metadata.ValidationStatus, &metadata.ValidationNotes)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("code metadata not found: %s (version: %s)", codeID, version)
		}
		return nil, fmt.Errorf("failed to get code metadata: %w", err)
	}

	// Parse JSON objects
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &metadata.Tags)
	}
	if len(customFieldsJSON) > 0 {
		json.Unmarshal(customFieldsJSON, &metadata.CustomFields)
	}

	return &metadata, nil
}

// SaveCodeRelationship saves or updates a code relationship
func (mm *MetadataManager) SaveCodeRelationship(ctx context.Context, rel *CodeRelationship) error {
	query := `
		INSERT INTO code_relationships (
			id, source_code_id, target_code_id, relationship_type, confidence, notes
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (source_code_id, target_code_id, relationship_type) DO UPDATE SET
			confidence = EXCLUDED.confidence,
			notes = EXCLUDED.notes,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := mm.db.ExecContext(ctx, query,
		rel.ID, rel.SourceCodeID, rel.TargetCodeID, rel.RelationshipType,
		rel.Confidence, rel.Notes)

	if err != nil {
		return fmt.Errorf("failed to save code relationship: %w", err)
	}

	mm.logger.Debug("Code relationship saved",
		zap.String("source", rel.SourceCodeID),
		zap.String("target", rel.TargetCodeID),
		zap.String("type", rel.RelationshipType))
	return nil
}

// GetCodeRelationships retrieves relationships for a given code
func (mm *MetadataManager) GetCodeRelationships(ctx context.Context, codeID string, relationshipType string) ([]*CodeRelationship, error) {
	query := `
		SELECT id, source_code_id, target_code_id, relationship_type, confidence,
		       notes, created_at, last_updated
		FROM code_relationships
		WHERE (source_code_id = $1 OR target_code_id = $1)
	`

	args := []interface{}{codeID}
	if relationshipType != "" {
		query += " AND relationship_type = $2"
		args = append(args, relationshipType)
	}

	query += " ORDER BY confidence DESC"

	rows, err := mm.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get code relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*CodeRelationship
	for rows.Next() {
		var rel CodeRelationship
		err := rows.Scan(
			&rel.ID, &rel.SourceCodeID, &rel.TargetCodeID, &rel.RelationshipType,
			&rel.Confidence, &rel.Notes, &rel.CreatedAt, &rel.LastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan code relationship: %w", err)
		}
		relationships = append(relationships, &rel)
	}

	return relationships, nil
}

// SaveCodeCrosswalk saves or updates a code crosswalk mapping
func (mm *MetadataManager) SaveCodeCrosswalk(ctx context.Context, crosswalk *CodeCrosswalk) error {
	query := `
		INSERT INTO code_crosswalks (
			id, source_code, source_type, target_code, target_type, mapping_type,
			confidence, direction, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (source_code, source_type, target_code, target_type) DO UPDATE SET
			mapping_type = EXCLUDED.mapping_type,
			confidence = EXCLUDED.confidence,
			direction = EXCLUDED.direction,
			notes = EXCLUDED.notes,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := mm.db.ExecContext(ctx, query,
		crosswalk.ID, crosswalk.SourceCode, crosswalk.SourceType,
		crosswalk.TargetCode, crosswalk.TargetType, crosswalk.MappingType,
		crosswalk.Confidence, crosswalk.Direction, crosswalk.Notes)

	if err != nil {
		return fmt.Errorf("failed to save code crosswalk: %w", err)
	}

	mm.logger.Debug("Code crosswalk saved",
		zap.String("source", crosswalk.SourceCode),
		zap.String("target", crosswalk.TargetCode))
	return nil
}

// GetCodeCrosswalks retrieves crosswalk mappings for a given code
func (mm *MetadataManager) GetCodeCrosswalks(ctx context.Context, sourceCode string, sourceType CodeType, targetType CodeType) ([]*CodeCrosswalk, error) {
	query := `
		SELECT id, source_code, source_type, target_code, target_type, mapping_type,
		       confidence, direction, notes, last_updated
		FROM code_crosswalks
		WHERE source_code = $1 AND source_type = $2 AND target_type = $3
		ORDER BY confidence DESC
	`

	rows, err := mm.db.QueryContext(ctx, query, sourceCode, sourceType, targetType)
	if err != nil {
		return nil, fmt.Errorf("failed to get code crosswalks: %w", err)
	}
	defer rows.Close()

	var crosswalks []*CodeCrosswalk
	for rows.Next() {
		var crosswalk CodeCrosswalk
		err := rows.Scan(
			&crosswalk.ID, &crosswalk.SourceCode, &crosswalk.SourceType,
			&crosswalk.TargetCode, &crosswalk.TargetType, &crosswalk.MappingType,
			&crosswalk.Confidence, &crosswalk.Direction, &crosswalk.Notes, &crosswalk.LastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan code crosswalk: %w", err)
		}
		crosswalks = append(crosswalks, &crosswalk)
	}

	return crosswalks, nil
}

// UpdateUsageCount increments the usage count for a code
func (mm *MetadataManager) UpdateUsageCount(ctx context.Context, codeID string, increment int64) error {
	query := `
		UPDATE code_metadata
		SET usage_count = usage_count + $1, last_updated = CURRENT_TIMESTAMP
		WHERE code_id = $2
	`

	result, err := mm.db.ExecContext(ctx, query, increment, codeID)
	if err != nil {
		return fmt.Errorf("failed to update usage count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		mm.logger.Warn("No metadata record found for usage count update", zap.String("code_id", codeID))
	}

	return nil
}

// ValidateCodeMetadata validates the quality and completeness of code metadata
func (mm *MetadataManager) ValidateCodeMetadata(ctx context.Context, codeID string) (*ValidationResult, error) {
	// Get the latest metadata
	metadata, err := mm.GetCodeMetadata(ctx, codeID, "latest")
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata for validation: %w", err)
	}

	// Get the latest description
	description, err := mm.GetLatestCodeDescription(ctx, codeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get description for validation: %w", err)
	}

	result := &ValidationResult{
		CodeID:          codeID,
		ValidationDate:  time.Now(),
		OverallScore:    0.0,
		Issues:          []string{},
		Recommendations: []string{},
	}

	// Validate description completeness
	if description.ShortDescription == "" {
		result.Issues = append(result.Issues, "Missing short description")
		result.OverallScore -= 0.2
	}

	if description.LongDescription == "" {
		result.Issues = append(result.Issues, "Missing long description")
		result.OverallScore -= 0.1
	}

	if len(description.Examples) == 0 {
		result.Recommendations = append(result.Recommendations, "Add examples for better understanding")
		result.OverallScore -= 0.05
	}

	// Validate metadata completeness
	if metadata.Source == "" {
		result.Issues = append(result.Issues, "Missing source information")
		result.OverallScore -= 0.15
	}

	if metadata.DataQuality == "" {
		result.Issues = append(result.Issues, "Missing data quality assessment")
		result.OverallScore -= 0.1
	}

	if metadata.Confidence == 0 {
		result.Issues = append(result.Issues, "Missing confidence score")
		result.OverallScore -= 0.1
	}

	// Validate relationships
	relationships, err := mm.GetCodeRelationships(ctx, codeID, "")
	if err == nil && len(relationships) == 0 {
		result.Recommendations = append(result.Recommendations, "Consider adding relationships to related codes")
		result.OverallScore -= 0.05
	}

	// Calculate final score (0.0 to 1.0)
	result.OverallScore = 1.0 + result.OverallScore
	if result.OverallScore < 0.0 {
		result.OverallScore = 0.0
	}

	// Set validation status
	if result.OverallScore >= 0.8 {
		result.Status = "valid"
	} else if result.OverallScore >= 0.6 {
		result.Status = "needs_improvement"
	} else {
		result.Status = "invalid"
	}

	return result, nil
}

// ValidationResult represents the result of metadata validation
type ValidationResult struct {
	CodeID          string    `json:"code_id"`
	ValidationDate  time.Time `json:"validation_date"`
	OverallScore    float64   `json:"overall_score"`
	Status          string    `json:"status"`
	Issues          []string  `json:"issues"`
	Recommendations []string  `json:"recommendations"`
}

// GetMetadataStats returns statistics about the metadata
func (mm *MetadataManager) GetMetadataStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count descriptions
	var descCount int
	err := mm.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM code_descriptions").Scan(&descCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count descriptions: %w", err)
	}
	stats["total_descriptions"] = descCount

	// Count metadata records
	var metadataCount int
	err = mm.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM code_metadata").Scan(&metadataCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count metadata: %w", err)
	}
	stats["total_metadata"] = metadataCount

	// Count relationships
	var relCount int
	err = mm.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM code_relationships").Scan(&relCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count relationships: %w", err)
	}
	stats["total_relationships"] = relCount

	// Count crosswalks
	var crosswalkCount int
	err = mm.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM code_crosswalks").Scan(&crosswalkCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count crosswalks: %w", err)
	}
	stats["total_crosswalks"] = crosswalkCount

	// Average confidence
	var avgConfidence sql.NullFloat64
	err = mm.db.QueryRowContext(ctx, "SELECT AVG(confidence) FROM code_metadata").Scan(&avgConfidence)
	if err != nil {
		return nil, fmt.Errorf("failed to get average confidence: %w", err)
	}
	if avgConfidence.Valid {
		stats["average_confidence"] = avgConfidence.Float64
	}

	return stats, nil
}

// Close closes the database connection
func (mm *MetadataManager) Close() error {
	return mm.db.Close()
}
