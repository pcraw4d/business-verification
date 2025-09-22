// Package repository provides data access layer for the KYB platform
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"kyb-platform/internal/models"
)

// KeywordRelationshipRepository defines the interface for keyword relationship operations
type KeywordRelationshipRepository interface {
	// GetKeywordRelationships retrieves all relationships for a primary keyword
	GetKeywordRelationships(ctx context.Context, primaryKeyword string) ([]models.KeywordRelationship, error)

	// GetRelatedKeywords gets all related keywords for a primary keyword
	GetRelatedKeywords(ctx context.Context, primaryKeyword string, relationshipTypes []string) ([]models.ExpandedKeyword, error)

	// ExpandKeyword expands a keyword using all relationship types
	ExpandKeyword(ctx context.Context, keyword string, industryID int) (*models.KeywordExpansionResult, error)

	// GetKeywordContexts retrieves context information for a keyword
	GetKeywordContexts(ctx context.Context, keyword string, industryID int) ([]models.KeywordContext, error)

	// CreateKeywordRelationship creates a new keyword relationship
	CreateKeywordRelationship(ctx context.Context, relationship *models.KeywordRelationship) error

	// CreateKeywordContext creates a new keyword context
	CreateKeywordContext(ctx context.Context, context *models.KeywordContext) error

	// BatchExpandKeywords expands multiple keywords efficiently
	BatchExpandKeywords(ctx context.Context, keywords []string, industryID int) ([]models.KeywordExpansionResult, error)
}

// keywordRelationshipRepository implements KeywordRelationshipRepository
type keywordRelationshipRepository struct {
	db *sql.DB
}

// NewKeywordRelationshipRepository creates a new keyword relationship repository
func NewKeywordRelationshipRepository(db *sql.DB) KeywordRelationshipRepository {
	return &keywordRelationshipRepository{
		db: db,
	}
}

// GetKeywordRelationships retrieves all relationships for a primary keyword
func (r *keywordRelationshipRepository) GetKeywordRelationships(ctx context.Context, primaryKeyword string) ([]models.KeywordRelationship, error) {
	query := `
		SELECT id, primary_keyword, related_keyword, relationship_type, 
		       confidence_score, is_active, created_at, updated_at
		FROM keyword_relationships 
		WHERE primary_keyword = $1 AND is_active = true
		ORDER BY confidence_score DESC, relationship_type
	`

	rows, err := r.db.QueryContext(ctx, query, primaryKeyword)
	if err != nil {
		return nil, fmt.Errorf("failed to query keyword relationships: %w", err)
	}
	defer rows.Close()

	var relationships []models.KeywordRelationship
	for rows.Next() {
		var rel models.KeywordRelationship
		err := rows.Scan(
			&rel.ID, &rel.PrimaryKeyword, &rel.RelatedKeyword, &rel.RelationshipType,
			&rel.ConfidenceScore, &rel.IsActive, &rel.CreatedAt, &rel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan keyword relationship: %w", err)
		}
		relationships = append(relationships, rel)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating keyword relationships: %w", err)
	}

	return relationships, nil
}

// GetRelatedKeywords gets all related keywords for a primary keyword
func (r *keywordRelationshipRepository) GetRelatedKeywords(ctx context.Context, primaryKeyword string, relationshipTypes []string) ([]models.ExpandedKeyword, error) {
	// Build placeholders for relationship types
	placeholders := make([]string, len(relationshipTypes))
	args := []interface{}{primaryKeyword}

	for i, relType := range relationshipTypes {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, relType)
	}

	query := fmt.Sprintf(`
		SELECT kr.related_keyword, kr.relationship_type, kr.confidence_score,
		       COALESCE(kc.context_weight, 1.0) as context_weight,
		       CASE WHEN kc.id IS NOT NULL THEN true ELSE false END as industry_relevant
		FROM keyword_relationships kr
		LEFT JOIN keyword_contexts kc ON kr.related_keyword = kc.keyword
		WHERE kr.primary_keyword = $1 
		  AND kr.relationship_type IN (%s)
		  AND kr.is_active = true
		ORDER BY kr.confidence_score DESC, kc.context_weight DESC
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query related keywords: %w", err)
	}
	defer rows.Close()

	var expandedKeywords []models.ExpandedKeyword
	for rows.Next() {
		var ek models.ExpandedKeyword
		err := rows.Scan(
			&ek.Keyword, &ek.RelationshipType, &ek.Confidence,
			&ek.ContextWeight, &ek.IndustryRelevant,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expanded keyword: %w", err)
		}
		expandedKeywords = append(expandedKeywords, ek)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expanded keywords: %w", err)
	}

	return expandedKeywords, nil
}

// ExpandKeyword expands a keyword using all relationship types
func (r *keywordRelationshipRepository) ExpandKeyword(ctx context.Context, keyword string, industryID int) (*models.KeywordExpansionResult, error) {
	// Get all relationship types
	allRelationshipTypes := []string{
		models.RelationshipTypeSynonym,
		models.RelationshipTypeAbbreviation,
		models.RelationshipTypeRelated,
		models.RelationshipTypeVariant,
	}

	// Get related keywords
	expandedKeywords, err := r.GetRelatedKeywords(ctx, keyword, allRelationshipTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get related keywords: %w", err)
	}

	// Calculate overall confidence based on relationships and context
	totalConfidence := 0.0
	totalWeight := 0.0

	for _, ek := range expandedKeywords {
		weight := ek.Confidence * ek.ContextWeight
		if ek.IndustryRelevant {
			weight *= 1.5 // Boost industry-relevant keywords
		}
		totalConfidence += weight
		totalWeight += 1.0
	}

	confidence := 0.0
	if totalWeight > 0 {
		confidence = totalConfidence / totalWeight
	}

	result := &models.KeywordExpansionResult{
		OriginalKeyword:  keyword,
		ExpandedKeywords: expandedKeywords,
		Confidence:       confidence,
		TotalMatches:     len(expandedKeywords),
	}

	return result, nil
}

// GetKeywordContexts retrieves context information for a keyword
func (r *keywordRelationshipRepository) GetKeywordContexts(ctx context.Context, keyword string, industryID int) ([]models.KeywordContext, error) {
	query := `
		SELECT id, keyword, industry_id, context_type, context_weight, 
		       is_active, created_at, updated_at
		FROM keyword_contexts 
		WHERE keyword = $1 AND (industry_id = $2 OR industry_id IS NULL) AND is_active = true
		ORDER BY context_weight DESC, context_type
	`

	rows, err := r.db.QueryContext(ctx, query, keyword, industryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query keyword contexts: %w", err)
	}
	defer rows.Close()

	var contexts []models.KeywordContext
	for rows.Next() {
		var ctx models.KeywordContext
		err := rows.Scan(
			&ctx.ID, &ctx.Keyword, &ctx.IndustryID, &ctx.ContextType,
			&ctx.ContextWeight, &ctx.IsActive, &ctx.CreatedAt, &ctx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan keyword context: %w", err)
		}
		contexts = append(contexts, ctx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating keyword contexts: %w", err)
	}

	return contexts, nil
}

// CreateKeywordRelationship creates a new keyword relationship
func (r *keywordRelationshipRepository) CreateKeywordRelationship(ctx context.Context, relationship *models.KeywordRelationship) error {
	query := `
		INSERT INTO keyword_relationships (primary_keyword, related_keyword, relationship_type, confidence_score, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		relationship.PrimaryKeyword,
		relationship.RelatedKeyword,
		relationship.RelationshipType,
		relationship.ConfidenceScore,
		relationship.IsActive,
	).Scan(&relationship.ID, &relationship.CreatedAt, &relationship.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create keyword relationship: %w", err)
	}

	return nil
}

// CreateKeywordContext creates a new keyword context
func (r *keywordRelationshipRepository) CreateKeywordContext(ctx context.Context, context *models.KeywordContext) error {
	query := `
		INSERT INTO keyword_contexts (keyword, industry_id, context_type, context_weight, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		context.Keyword,
		context.IndustryID,
		context.ContextType,
		context.ContextWeight,
		context.IsActive,
	).Scan(&context.ID, &context.CreatedAt, &context.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create keyword context: %w", err)
	}

	return nil
}

// BatchExpandKeywords expands multiple keywords efficiently
func (r *keywordRelationshipRepository) BatchExpandKeywords(ctx context.Context, keywords []string, industryID int) ([]models.KeywordExpansionResult, error) {
	if len(keywords) == 0 {
		return []models.KeywordExpansionResult{}, nil
	}

	results := make([]models.KeywordExpansionResult, 0, len(keywords))

	// Process keywords in batches to avoid too many database queries
	for _, keyword := range keywords {
		result, err := r.ExpandKeyword(ctx, keyword, industryID)
		if err != nil {
			// Log error but continue with other keywords
			continue
		}
		if result != nil {
			results = append(results, *result)
		}
	}

	return results, nil
}
