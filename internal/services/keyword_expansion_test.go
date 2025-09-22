// Package services provides tests for keyword expansion functionality
package services

import (
	"context"
	"testing"

	"kyb-platform/internal/models"
)

// MockKeywordRelationshipRepository provides a mock implementation for testing
type MockKeywordRelationshipRepository struct {
	relationships []models.KeywordRelationship
	contexts      []models.KeywordContext
}

func NewMockKeywordRelationshipRepository() *MockKeywordRelationshipRepository {
	return &MockKeywordRelationshipRepository{
		relationships: []models.KeywordRelationship{
			{
				ID:               1,
				PrimaryKeyword:   "software",
				RelatedKeyword:   "application",
				RelationshipType: models.RelationshipTypeSynonym,
				ConfidenceScore:  0.95,
				IsActive:         true,
			},
			{
				ID:               2,
				PrimaryKeyword:   "software",
				RelatedKeyword:   "app",
				RelationshipType: models.RelationshipTypeAbbreviation,
				ConfidenceScore:  0.90,
				IsActive:         true,
			},
			{
				ID:               3,
				PrimaryKeyword:   "technology",
				RelatedKeyword:   "tech",
				RelationshipType: models.RelationshipTypeAbbreviation,
				ConfidenceScore:  0.95,
				IsActive:         true,
			},
			{
				ID:               4,
				PrimaryKeyword:   "medical",
				RelatedKeyword:   "healthcare",
				RelationshipType: models.RelationshipTypeSynonym,
				ConfidenceScore:  0.90,
				IsActive:         true,
			},
			{
				ID:               5,
				PrimaryKeyword:   "banking",
				RelatedKeyword:   "finance",
				RelationshipType: models.RelationshipTypeRelated,
				ConfidenceScore:  0.85,
				IsActive:         true,
			},
		},
		contexts: []models.KeywordContext{
			{
				ID:            1,
				Keyword:       "software",
				IndustryID:    1, // Technology
				ContextType:   models.ContextTypePrimary,
				ContextWeight: 1.5,
				IsActive:      true,
			},
			{
				ID:            2,
				Keyword:       "application",
				IndustryID:    1, // Technology
				ContextType:   models.ContextTypeTechnical,
				ContextWeight: 1.4,
				IsActive:      true,
			},
			{
				ID:            3,
				Keyword:       "medical",
				IndustryID:    2, // Healthcare
				ContextType:   models.ContextTypePrimary,
				ContextWeight: 1.5,
				IsActive:      true,
			},
		},
	}
}

func (m *MockKeywordRelationshipRepository) GetKeywordRelationships(ctx context.Context, primaryKeyword string) ([]models.KeywordRelationship, error) {
	var results []models.KeywordRelationship
	for _, rel := range m.relationships {
		if rel.PrimaryKeyword == primaryKeyword && rel.IsActive {
			results = append(results, rel)
		}
	}
	return results, nil
}

func (m *MockKeywordRelationshipRepository) GetRelatedKeywords(ctx context.Context, primaryKeyword string, relationshipTypes []string) ([]models.ExpandedKeyword, error) {
	var results []models.ExpandedKeyword

	for _, rel := range m.relationships {
		if rel.PrimaryKeyword == primaryKeyword && rel.IsActive {
			// Check if relationship type is in the requested types
			for _, relType := range relationshipTypes {
				if rel.RelationshipType == relType {
					// Find context weight
					contextWeight := 1.0
					industryRelevant := false
					for _, ctx := range m.contexts {
						if ctx.Keyword == rel.RelatedKeyword && ctx.IsActive {
							contextWeight = ctx.ContextWeight
							industryRelevant = true
							break
						}
					}

					results = append(results, models.ExpandedKeyword{
						Keyword:          rel.RelatedKeyword,
						RelationshipType: rel.RelationshipType,
						Confidence:       rel.ConfidenceScore,
						ContextWeight:    contextWeight,
						IndustryRelevant: industryRelevant,
					})
					break
				}
			}
		}
	}

	return results, nil
}

func (m *MockKeywordRelationshipRepository) ExpandKeyword(ctx context.Context, keyword string, industryID int) (*models.KeywordExpansionResult, error) {
	allRelationshipTypes := []string{
		models.RelationshipTypeSynonym,
		models.RelationshipTypeAbbreviation,
		models.RelationshipTypeRelated,
		models.RelationshipTypeVariant,
	}

	expandedKeywords, err := m.GetRelatedKeywords(ctx, keyword, allRelationshipTypes)
	if err != nil {
		return nil, err
	}

	// Calculate confidence
	totalConfidence := 0.0
	for _, ek := range expandedKeywords {
		weight := ek.Confidence * ek.ContextWeight
		if ek.IndustryRelevant {
			weight *= 1.5
		}
		totalConfidence += weight
	}

	confidence := 0.0
	if len(expandedKeywords) > 0 {
		confidence = totalConfidence / float64(len(expandedKeywords))
	}

	return &models.KeywordExpansionResult{
		OriginalKeyword:  keyword,
		ExpandedKeywords: expandedKeywords,
		Confidence:       confidence,
		TotalMatches:     len(expandedKeywords),
	}, nil
}

func (m *MockKeywordRelationshipRepository) GetKeywordContexts(ctx context.Context, keyword string, industryID int) ([]models.KeywordContext, error) {
	var results []models.KeywordContext
	for _, ctx := range m.contexts {
		if ctx.Keyword == keyword && ctx.IsActive && (ctx.IndustryID == industryID || industryID == 0) {
			results = append(results, ctx)
		}
	}
	return results, nil
}

func (m *MockKeywordRelationshipRepository) CreateKeywordRelationship(ctx context.Context, relationship *models.KeywordRelationship) error {
	relationship.ID = len(m.relationships) + 1
	m.relationships = append(m.relationships, *relationship)
	return nil
}

func (m *MockKeywordRelationshipRepository) CreateKeywordContext(ctx context.Context, context *models.KeywordContext) error {
	context.ID = len(m.contexts) + 1
	m.contexts = append(m.contexts, *context)
	return nil
}

func (m *MockKeywordRelationshipRepository) BatchExpandKeywords(ctx context.Context, keywords []string, industryID int) ([]models.KeywordExpansionResult, error) {
	results := make([]models.KeywordExpansionResult, 0, len(keywords))

	for _, keyword := range keywords {
		result, err := m.ExpandKeyword(ctx, keyword, industryID)
		if err == nil && result != nil {
			results = append(results, *result)
		}
	}

	return results, nil
}

// TestKeywordExpansionService tests the keyword expansion service
func TestKeywordExpansionService(t *testing.T) {
	ctx := context.Background()
	mockRepo := NewMockKeywordRelationshipRepository()
	service := NewKeywordExpansionService(mockRepo)

	t.Run("ExpandKeywords", func(t *testing.T) {
		keywords := []string{"software", "technology"}
		industryID := 1 // Technology industry

		expandedKeywords, err := service.ExpandKeywords(ctx, keywords, industryID)
		if err != nil {
			t.Fatalf("ExpandKeywords failed: %v", err)
		}

		// Should include original keywords plus expanded ones
		expectedKeywords := map[string]bool{
			"software":    true,
			"technology":  true,
			"application": true, // synonym of software
			"app":         true, // abbreviation of software
			"tech":        true, // abbreviation of technology
		}

		for _, keyword := range expandedKeywords {
			if !expectedKeywords[keyword] {
				t.Errorf("Unexpected expanded keyword: %s", keyword)
			}
		}

		if len(expandedKeywords) < 4 { // At least original + some expansions
			t.Errorf("Expected at least 4 keywords, got %d", len(expandedKeywords))
		}
	})

	t.Run("ExpandKeywordWithDetails", func(t *testing.T) {
		keyword := "software"
		industryID := 1

		result, err := service.ExpandKeywordWithDetails(ctx, keyword, industryID)
		if err != nil {
			t.Fatalf("ExpandKeywordWithDetails failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if result.OriginalKeyword != keyword {
			t.Errorf("Expected original keyword %s, got %s", keyword, result.OriginalKeyword)
		}

		if len(result.ExpandedKeywords) == 0 {
			t.Error("Expected expanded keywords")
		}

		if result.Confidence <= 0 {
			t.Errorf("Expected positive confidence, got %f", result.Confidence)
		}

		// Check that we have both synonym and abbreviation
		hasSynonym := false
		hasAbbreviation := false
		for _, ek := range result.ExpandedKeywords {
			if ek.RelationshipType == models.RelationshipTypeSynonym {
				hasSynonym = true
			}
			if ek.RelationshipType == models.RelationshipTypeAbbreviation {
				hasAbbreviation = true
			}
		}

		if !hasSynonym {
			t.Error("Expected at least one synonym")
		}
		if !hasAbbreviation {
			t.Error("Expected at least one abbreviation")
		}
	})

	t.Run("GetSynonyms", func(t *testing.T) {
		synonyms, err := service.GetSynonyms(ctx, "software")
		if err != nil {
			t.Fatalf("GetSynonyms failed: %v", err)
		}

		expectedSynonyms := []string{"application"}
		if len(synonyms) != len(expectedSynonyms) {
			t.Errorf("Expected %d synonyms, got %d", len(expectedSynonyms), len(synonyms))
		}

		for _, expected := range expectedSynonyms {
			found := false
			for _, synonym := range synonyms {
				if synonym == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected synonym %s not found", expected)
			}
		}
	})

	t.Run("GetAbbreviations", func(t *testing.T) {
		abbreviations, err := service.GetAbbreviations(ctx, "software")
		if err != nil {
			t.Fatalf("GetAbbreviations failed: %v", err)
		}

		expectedAbbreviations := []string{"app"}
		if len(abbreviations) != len(expectedAbbreviations) {
			t.Errorf("Expected %d abbreviations, got %d", len(expectedAbbreviations), len(abbreviations))
		}

		for _, expected := range expectedAbbreviations {
			found := false
			for _, abbrev := range abbreviations {
				if abbrev == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected abbreviation %s not found", expected)
			}
		}
	})

	t.Run("GetRelatedTerms", func(t *testing.T) {
		relatedTerms, err := service.GetRelatedTerms(ctx, "banking")
		if err != nil {
			t.Fatalf("GetRelatedTerms failed: %v", err)
		}

		expectedRelated := []string{"finance"}
		if len(relatedTerms) != len(expectedRelated) {
			t.Errorf("Expected %d related terms, got %d", len(expectedRelated), len(relatedTerms))
		}

		for _, expected := range expectedRelated {
			found := false
			for _, related := range relatedTerms {
				if related == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected related term %s not found", expected)
			}
		}
	})

	t.Run("ExpandBusinessDescription", func(t *testing.T) {
		description := "We provide software solutions and technology services for businesses"
		industryID := 1

		expandedKeywords, err := service.ExpandBusinessDescription(ctx, description, industryID)
		if err != nil {
			t.Fatalf("ExpandBusinessDescription failed: %v", err)
		}

		// Should extract and expand keywords from description
		expectedKeywords := map[string]bool{
			"software":    true,
			"technology":  true,
			"solutions":   true,
			"services":    true,
			"businesses":  true,
			"application": true, // expanded from software
			"app":         true, // expanded from software
			"tech":        true, // expanded from technology
		}

		foundExpectedCount := 0
		for _, keyword := range expandedKeywords {
			if expectedKeywords[keyword] {
				foundExpectedCount++
			}
		}

		if foundExpectedCount < 5 { // Should find most expected keywords
			t.Errorf("Expected to find at least 5 expected keywords, found %d", foundExpectedCount)
		}
	})

	t.Run("CalculateKeywordRelevance", func(t *testing.T) {
		keywords := []string{"software", "application", "unknown"}
		industryID := 1

		relevanceScores, err := service.CalculateKeywordRelevance(ctx, keywords, industryID)
		if err != nil {
			t.Fatalf("CalculateKeywordRelevance failed: %v", err)
		}

		if len(relevanceScores) != len(keywords) {
			t.Errorf("Expected %d relevance scores, got %d", len(keywords), len(relevanceScores))
		}

		// Software should have high relevance (primary context weight 1.5)
		if relevanceScores["software"] != 1.5 {
			t.Errorf("Expected software relevance 1.5, got %f", relevanceScores["software"])
		}

		// Application should have good relevance (technical context weight 1.4)
		if relevanceScores["application"] != 1.4 {
			t.Errorf("Expected application relevance 1.4, got %f", relevanceScores["application"])
		}

		// Unknown should have default relevance (0.5)
		if relevanceScores["unknown"] != 0.5 {
			t.Errorf("Expected unknown relevance 0.5, got %f", relevanceScores["unknown"])
		}
	})

	t.Run("EmptyInputs", func(t *testing.T) {
		// Test empty keywords list
		expandedKeywords, err := service.ExpandKeywords(ctx, []string{}, 1)
		if err != nil {
			t.Fatalf("ExpandKeywords with empty list failed: %v", err)
		}
		if len(expandedKeywords) != 0 {
			t.Errorf("Expected empty result for empty input, got %d keywords", len(expandedKeywords))
		}

		// Test empty keyword string
		result, err := service.ExpandKeywordWithDetails(ctx, "", 1)
		if err == nil {
			t.Error("Expected error for empty keyword")
		}
		if result != nil {
			t.Error("Expected nil result for empty keyword")
		}

		// Test empty description
		expandedFromDesc, err := service.ExpandBusinessDescription(ctx, "", 1)
		if err != nil {
			t.Fatalf("ExpandBusinessDescription with empty description failed: %v", err)
		}
		if len(expandedFromDesc) != 0 {
			t.Errorf("Expected empty result for empty description, got %d keywords", len(expandedFromDesc))
		}
	})
}

// TestKeywordExpansionIntegration tests integration scenarios
func TestKeywordExpansionIntegration(t *testing.T) {
	ctx := context.Background()
	mockRepo := NewMockKeywordRelationshipRepository()
	service := NewKeywordExpansionService(mockRepo)

	t.Run("TechnologyBusinessExpansion", func(t *testing.T) {
		businessName := "TechCorp Software Solutions"
		description := "We develop mobile applications and web software for enterprises"

		// Extract keywords from business name and description
		allText := businessName + " " + description
		expandedKeywords, err := service.ExpandBusinessDescription(ctx, allText, 1)
		if err != nil {
			t.Fatalf("Business expansion failed: %v", err)
		}

		// Should include expanded keywords
		keywordSet := make(map[string]bool)
		for _, keyword := range expandedKeywords {
			keywordSet[keyword] = true
		}

		// Check for original and expanded terms
		expectedTerms := []string{"software", "application", "app", "tech"}
		foundCount := 0
		for _, term := range expectedTerms {
			if keywordSet[term] {
				foundCount++
			}
		}

		if foundCount < 3 {
			t.Errorf("Expected to find at least 3 technology-related terms, found %d", foundCount)
		}
	})

	t.Run("HealthcareBusinessExpansion", func(t *testing.T) {
		businessName := "MedTech Healthcare Solutions"
		description := "We provide medical technology and healthcare services"

		allText := businessName + " " + description
		expandedKeywords, err := service.ExpandBusinessDescription(ctx, allText, 2)
		if err != nil {
			t.Fatalf("Healthcare business expansion failed: %v", err)
		}

		keywordSet := make(map[string]bool)
		for _, keyword := range expandedKeywords {
			keywordSet[keyword] = true
		}

		// Should find medical and healthcare terms
		if !keywordSet["medical"] || !keywordSet["healthcare"] {
			t.Error("Expected to find both 'medical' and 'healthcare' terms")
		}
	})

	t.Run("CrossIndustryKeywords", func(t *testing.T) {
		// Test keywords that might appear in multiple industries
		keywords := []string{"software", "technology", "solutions", "services"}

		// Technology industry
		techExpanded, err := service.ExpandKeywords(ctx, keywords, 1)
		if err != nil {
			t.Fatalf("Technology expansion failed: %v", err)
		}

		// Healthcare industry (assuming some overlap)
		healthExpanded, err := service.ExpandKeywords(ctx, keywords, 2)
		if err != nil {
			t.Fatalf("Healthcare expansion failed: %v", err)
		}

		// Both should expand, but potentially differently
		if len(techExpanded) == 0 || len(healthExpanded) == 0 {
			t.Error("Expected expansions for both industries")
		}

		// Technology should have more tech-specific expansions
		techSet := make(map[string]bool)
		for _, keyword := range techExpanded {
			techSet[keyword] = true
		}

		if !techSet["application"] || !techSet["app"] {
			t.Error("Expected technology-specific expansions")
		}
	})
}

// BenchmarkKeywordExpansion benchmarks the keyword expansion performance
func BenchmarkKeywordExpansion(b *testing.B) {
	ctx := context.Background()
	mockRepo := NewMockKeywordRelationshipRepository()
	service := NewKeywordExpansionService(mockRepo)

	keywords := []string{"software", "technology", "medical", "banking", "retail"}
	industryID := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ExpandKeywords(ctx, keywords, industryID)
		if err != nil {
			b.Fatalf("ExpandKeywords failed: %v", err)
		}
	}
}

func BenchmarkKeywordExpansionDetails(b *testing.B) {
	ctx := context.Background()
	mockRepo := NewMockKeywordRelationshipRepository()
	service := NewKeywordExpansionService(mockRepo)

	keyword := "software"
	industryID := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ExpandKeywordWithDetails(ctx, keyword, industryID)
		if err != nil {
			b.Fatalf("ExpandKeywordWithDetails failed: %v", err)
		}
	}
}
