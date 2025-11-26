package nlp

import (
	"strings"
	"testing"
)

func TestNewEntityRecognizer(t *testing.T) {
	er := NewEntityRecognizer()
	if er == nil {
		t.Fatal("NewEntityRecognizer() returned nil")
	}
	if len(er.patterns) == 0 {
		t.Error("EntityRecognizer has no patterns loaded")
	}
}

func TestExtractEntities(t *testing.T) {
	er := NewEntityRecognizer()

	tests := []struct {
		name           string
		text           string
		expectedTypes  []EntityType
		minEntities    int
		expectedTexts  []string
	}{
		{
			name:          "wine shop business",
			text:          "We are a wine shop specializing in fine wines",
			expectedTypes: []EntityType{EntityTypeBusinessType, EntityTypeProduct},
			minEntities:   2,
			expectedTexts: []string{"wine shop", "wine", "wines"},
		},
		{
			name:          "retail store",
			text:          "Our retail store offers a wide variety of products",
			expectedTypes: []EntityType{EntityTypeBusinessType},
			minEntities:   1,
			expectedTexts: []string{"retail store"},
		},
		{
			name:          "technology company",
			text:          "We are a technology company providing software solutions",
			expectedTypes: []EntityType{EntityTypeBusinessType, EntityTypeProduct},
			minEntities:   2,
			expectedTexts: []string{"technology company", "software"},
		},
		{
			name:          "restaurant with location",
			text:          "Our restaurant in California serves fine dining",
			expectedTypes: []EntityType{EntityTypeBusinessType, EntityTypeLocation},
			minEntities:   2,
			expectedTexts: []string{"restaurant", "california"},
		},
		{
			name:          "online store",
			text:          "Visit our online store for the best deals",
			expectedTypes: []EntityType{EntityTypeBusinessType},
			minEntities:   1,
			expectedTexts: []string{"online store"},
		},
		{
			name:          "financial services",
			text:          "We provide financial services and banking solutions",
			expectedTypes: []EntityType{EntityTypeBusinessType, EntityTypeIndustry},
			minEntities:   2,
			expectedTexts: []string{"financial services"},
		},
		{
			name:          "winery",
			text:          "Our winery produces premium wines from local vineyards",
			expectedTypes: []EntityType{EntityTypeBusinessType, EntityTypeProduct},
			minEntities:   2,
			expectedTexts: []string{"winery", "wine", "wines"},
		},
		{
			name:          "empty text",
			text:          "",
			expectedTypes: []EntityType{},
			minEntities:   0,
			expectedTexts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities := er.ExtractEntities(tt.text)

			if len(entities) < tt.minEntities {
				t.Errorf("ExtractEntities() returned %d entities, want at least %d", len(entities), tt.minEntities)
			}

			// Check for expected entity types
			typeMap := make(map[EntityType]bool)
			for _, entity := range entities {
				typeMap[entity.Type] = true
			}

			for _, expectedType := range tt.expectedTypes {
				if !typeMap[expectedType] {
					t.Errorf("ExtractEntities() did not find expected entity type: %s", expectedType)
				}
			}

			// Check for expected texts (case-insensitive)
			textMap := make(map[string]bool)
			for _, entity := range entities {
				textMap[strings.ToLower(entity.Text)] = true
			}

			foundCount := 0
			for _, expectedText := range tt.expectedTexts {
				if textMap[strings.ToLower(expectedText)] {
					foundCount++
				}
			}

			if foundCount == 0 && len(tt.expectedTexts) > 0 {
				t.Logf("Warning: No expected texts found. Got entities: %v", entities)
			}

			// Verify entity properties
			for _, entity := range entities {
				if entity.Text == "" {
					t.Errorf("ExtractEntities() returned entity with empty text")
				}
				if entity.Confidence < 0.0 || entity.Confidence > 1.0 {
					t.Errorf("ExtractEntities() returned entity with invalid confidence: %f", entity.Confidence)
				}
				if entity.Source == "" {
					t.Errorf("ExtractEntities() returned entity with empty source")
				}
			}
		})
	}
}

func TestGetEntitiesByType(t *testing.T) {
	er := NewEntityRecognizer()

	text := "We are a wine shop in California providing retail services"
	entities := er.ExtractEntities(text)

	businessTypes := er.GetEntitiesByType(entities, EntityTypeBusinessType)
	if len(businessTypes) == 0 {
		t.Error("GetEntitiesByType() returned no business type entities")
	}

	locations := er.GetEntitiesByType(entities, EntityTypeLocation)
	if len(locations) == 0 {
		t.Error("GetEntitiesByType() returned no location entities")
	}

	// Verify all returned entities are of the correct type
	for _, entity := range businessTypes {
		if entity.Type != EntityTypeBusinessType {
			t.Errorf("GetEntitiesByType() returned entity with wrong type: %s", entity.Type)
		}
	}
}

func TestGetEntityKeywords(t *testing.T) {
	er := NewEntityRecognizer()

	text := "We are a wine shop specializing in fine wines"
	entities := er.ExtractEntities(text)
	keywords := er.GetEntityKeywords(entities)

	if len(keywords) == 0 {
		t.Error("GetEntityKeywords() returned no keywords")
	}

	// Check that keywords are extracted
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[kw] = true
	}

	expectedKeywords := []string{"wine", "shop", "wines"}
	foundCount := 0
	for _, expected := range expectedKeywords {
		if keywordMap[expected] {
			foundCount++
		}
	}

	if foundCount == 0 {
		t.Errorf("GetEntityKeywords() did not find expected keywords. Got: %v", keywords)
	}

	// Verify keywords are lowercase and have minimum length
	for _, kw := range keywords {
		if kw != strings.ToLower(kw) {
			t.Errorf("GetEntityKeywords() returned non-lowercase keyword: %s", kw)
		}
		if len(kw) < 3 {
			t.Errorf("GetEntityKeywords() returned keyword shorter than 3 characters: %s", kw)
		}
	}
}

func TestDeduplicateEntities(t *testing.T) {
	er := NewEntityRecognizer()

	// Create duplicate entities
	entities := []Entity{
		{Text: "wine shop", Type: EntityTypeBusinessType, Confidence: 0.95, Source: "pattern"},
		{Text: "wine shop", Type: EntityTypeBusinessType, Confidence: 0.90, Source: "pattern"},
		{Text: "retail store", Type: EntityTypeBusinessType, Confidence: 0.95, Source: "pattern"},
	}

	deduplicated := er.deduplicateEntities(entities)

	if len(deduplicated) != 2 {
		t.Errorf("deduplicateEntities() returned %d entities, want 2", len(deduplicated))
	}

	// Verify higher confidence entity is kept
	for _, entity := range deduplicated {
		if entity.Text == "wine shop" && entity.Confidence != 0.95 {
			t.Errorf("deduplicateEntities() kept lower confidence entity: %f", entity.Confidence)
		}
	}
}

func TestAddPattern(t *testing.T) {
	er := NewEntityRecognizer()

	initialCount := len(er.patterns)

	err := er.AddPattern(`\b(test\s+pattern)\b`, EntityTypeProduct, 0.85, "Test pattern")
	if err != nil {
		t.Fatalf("AddPattern() returned error: %v", err)
	}

	if len(er.patterns) != initialCount+1 {
		t.Errorf("AddPattern() did not add pattern. Expected %d patterns, got %d", initialCount+1, len(er.patterns))
	}

	// Test that the pattern works
	entities := er.ExtractEntities("This is a test pattern")
	found := false
	for _, entity := range entities {
		if entity.Text == "test pattern" {
			found = true
			break
		}
	}

	if !found {
		t.Error("AddPattern() added pattern but it doesn't match test text")
	}

	// Test invalid pattern
	err = er.AddPattern(`[invalid regex`, EntityTypeProduct, 0.85, "Invalid")
	if err == nil {
		t.Error("AddPattern() should return error for invalid regex pattern")
	}
}

func BenchmarkExtractEntities(b *testing.B) {
	er := NewEntityRecognizer()
	text := "We are a wine shop in California providing retail services and fine wines to customers across the United States"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		er.ExtractEntities(text)
	}
}

