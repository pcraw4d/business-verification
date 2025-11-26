package nlp

import (
	"regexp"
	"strings"
	"sync"
)

// EntityType represents the type of entity extracted
type EntityType string

const (
	// Business entity types
	EntityTypeBusinessType EntityType = "BUSINESS_TYPE"
	EntityTypeService     EntityType = "SERVICE"
	EntityTypeProduct     EntityType = "PRODUCT"
	EntityTypeIndustry    EntityType = "INDUSTRY"
	EntityTypeLocation    EntityType = "LOCATION"
	EntityTypeBrand       EntityType = "BRAND"
)

// Entity represents a named entity extracted from text
type Entity struct {
	Text       string    // The extracted entity text
	Type       EntityType // The type of entity
	Confidence float64   // Confidence score (0.0 to 1.0)
	Source     string    // "pattern" or "nlp" - indicates extraction method
	Start      int       // Start position in original text
	End        int       // End position in original text
}

// EntityPattern represents a pattern for matching entities
type EntityPattern struct {
	Pattern     *regexp.Regexp
	Type        EntityType
	Confidence  float64
	Description string
}

// EntityRecognizer performs named entity recognition using pattern-based and library-based approaches
type EntityRecognizer struct {
	patterns   []EntityPattern
	patternsMu sync.RWMutex
	// Library-based NER can be added later if needed
	// For now, we focus on pattern-based approach for performance and reliability
}

// NewEntityRecognizer creates a new entity recognizer with default patterns
func NewEntityRecognizer() *EntityRecognizer {
	er := &EntityRecognizer{
		patterns: make([]EntityPattern, 0),
	}
	er.loadDefaultPatterns()
	return er
}

// ExtractEntities extracts named entities from text using pattern-based and library-based methods
func (er *EntityRecognizer) ExtractEntities(text string) []Entity {
	if text == "" {
		return []Entity{}
	}

	// Normalize text for processing
	normalizedText := strings.ToLower(text)
	entities := []Entity{}

	// Pattern-based extraction (fast, high precision)
	patternEntities := er.extractWithPatterns(normalizedText, text)
	entities = append(entities, patternEntities...)

	// Library-based extraction can be added here if needed
	// For now, pattern-based is sufficient for business classification

	// Deduplicate entities
	entities = er.deduplicateEntities(entities)

	return entities
}

// extractWithPatterns extracts entities using regex patterns
func (er *EntityRecognizer) extractWithPatterns(normalizedText, originalText string) []Entity {
	er.patternsMu.RLock()
	patterns := er.patterns
	er.patternsMu.RUnlock()

	entities := []Entity{}
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		matches := pattern.Pattern.FindAllStringSubmatchIndex(normalizedText, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				start := match[0]
				end := match[1]
				entityText := normalizedText[start:end]

				// Create unique key for deduplication
				key := strings.ToLower(entityText) + ":" + string(pattern.Type)
				if !seen[key] {
					seen[key] = true
					entities = append(entities, Entity{
						Text:       entityText,
						Type:       pattern.Type,
						Confidence: pattern.Confidence,
						Source:     "pattern",
						Start:      start,
						End:        end,
					})
				}
			}
		}
	}

	return entities
}

// deduplicateEntities removes duplicate entities, keeping the one with higher confidence
func (er *EntityRecognizer) deduplicateEntities(entities []Entity) []Entity {
	if len(entities) == 0 {
		return entities
	}

	// Map to track best entity for each text+type combination
	entityMap := make(map[string]Entity)

	for _, entity := range entities {
		key := strings.ToLower(entity.Text) + ":" + string(entity.Type)
		if existing, exists := entityMap[key]; !exists || entity.Confidence > existing.Confidence {
			entityMap[key] = entity
		}
	}

	// Convert map back to slice
	result := make([]Entity, 0, len(entityMap))
	for _, entity := range entityMap {
		result = append(result, entity)
	}

	return result
}

// loadDefaultPatterns loads default entity recognition patterns
func (er *EntityRecognizer) loadDefaultPatterns() {
	patterns := []struct {
		pattern     string
		entityType  EntityType
		confidence  float64
		description string
	}{
		// Business Types
		{`\b(wine\s+shop|wine\s+store|wine\s+merchant|wine\s+retailer|wine\s+bar|wine\s+cellar)\b`, EntityTypeBusinessType, 0.95, "Wine retail business"},
		{`\b(retail\s+store|retail\s+shop|brick\s+and\s+mortar|physical\s+store|storefront)\b`, EntityTypeBusinessType, 0.95, "Retail store"},
		{`\b(online\s+store|online\s+shop|ecommerce\s+store|web\s+store|digital\s+storefront)\b`, EntityTypeBusinessType, 0.95, "Online store"},
		{`\b(restaurant|cafe|coffee\s+shop|bistro|diner|tavern|eatery|gastropub)\b`, EntityTypeBusinessType, 0.95, "Restaurant"},
		{`\b(winery|vineyard|vintner|distillery|brewery)\b`, EntityTypeBusinessType, 0.95, "Beverage producer"},
		{`\b(technology\s+company|tech\s+firm|software\s+company|IT\s+services)\b`, EntityTypeBusinessType, 0.90, "Technology company"},
		{`\b(financial\s+services|bank|credit\s+union|investment\s+firm)\b`, EntityTypeBusinessType, 0.90, "Financial services"},
		{`\b(healthcare\s+provider|medical\s+clinic|hospital|pharmacy)\b`, EntityTypeBusinessType, 0.90, "Healthcare provider"},
		{`\b(real\s+estate|property\s+management|realty|realtor)\b`, EntityTypeBusinessType, 0.90, "Real estate"},
		{`\b(construction\s+company|contractor|builder|construction\s+firm)\b`, EntityTypeBusinessType, 0.90, "Construction company"},

		// Services
		{`\b(consulting\s+services|advisory\s+services|consulting\s+firm)\b`, EntityTypeService, 0.85, "Consulting services"},
		{`\b(legal\s+services|law\s+firm|attorney|lawyer)\b`, EntityTypeService, 0.85, "Legal services"},
		{`\b(accounting\s+services|accountant|bookkeeping|tax\s+preparation)\b`, EntityTypeService, 0.85, "Accounting services"},
		{`\b(marketing\s+services|advertising\s+agency|marketing\s+firm)\b`, EntityTypeService, 0.85, "Marketing services"},
		{`\b(web\s+design|web\s+development|software\s+development)\b`, EntityTypeService, 0.85, "Web/software services"},
		{`\b(logistics|shipping|freight|delivery\s+services)\b`, EntityTypeService, 0.85, "Logistics services"},
		{`\b(education\s+services|training|tutoring|educational\s+institution)\b`, EntityTypeService, 0.85, "Education services"},

		// Products
		{`\b(wine|wines|vintage\s+wine|fine\s+wine|premium\s+wine)\b`, EntityTypeProduct, 0.90, "Wine products"},
		{`\b(alcohol|spirits|liquor|beer|beverages)\b`, EntityTypeProduct, 0.85, "Beverage products"},
		{`\b(software|application|app|platform|system)\b`, EntityTypeProduct, 0.80, "Software products"},
		{`\b(consumer\s+goods|products|merchandise|inventory)\b`, EntityTypeProduct, 0.75, "Consumer goods"},

		// Industry Indicators
		{`\b(food\s+and\s+beverage|F&B|foodservice|hospitality)\b`, EntityTypeIndustry, 0.95, "Food & Beverage industry"},
		{`\b(retail\s+industry|retail\s+sector|commerce|merchandising)\b`, EntityTypeIndustry, 0.95, "Retail industry"},
		{`\b(technology\s+industry|tech\s+sector|IT\s+industry|software\s+industry)\b`, EntityTypeIndustry, 0.95, "Technology industry"},
		{`\b(financial\s+services|banking|finance|fintech)\b`, EntityTypeIndustry, 0.95, "Financial services industry"},
		{`\b(healthcare\s+industry|medical\s+industry|health\s+services)\b`, EntityTypeIndustry, 0.95, "Healthcare industry"},
		{`\b(manufacturing|production|factory|industrial)\b`, EntityTypeIndustry, 0.90, "Manufacturing industry"},
		{`\b(construction\s+industry|building|construction\s+sector)\b`, EntityTypeIndustry, 0.90, "Construction industry"},
		{`\b(agriculture|farming|agricultural)\b`, EntityTypeIndustry, 0.90, "Agriculture industry"},
		{`\b(transportation|logistics\s+industry|shipping\s+industry)\b`, EntityTypeIndustry, 0.90, "Transportation industry"},
		{`\b(real\s+estate\s+industry|property\s+industry)\b`, EntityTypeIndustry, 0.90, "Real estate industry"},

		// Location Entities (for regional classification)
		{`\b(united\s+states|USA|US|America)\b`, EntityTypeLocation, 0.95, "United States"},
		{`\b(canada|canadian)\b`, EntityTypeLocation, 0.95, "Canada"},
		{`\b(united\s+kingdom|UK|Britain|England)\b`, EntityTypeLocation, 0.95, "United Kingdom"},
		{`\b(australia|australian)\b`, EntityTypeLocation, 0.95, "Australia"},
		{`\b(europe|european|EU)\b`, EntityTypeLocation, 0.90, "Europe"},
		{`\b(california|CA|new\s+york|NY|texas|TX|florida|FL)\b`, EntityTypeLocation, 0.85, "US State"},
		{`\b(london|paris|tokyo|sydney|toronto|vancouver)\b`, EntityTypeLocation, 0.85, "Major city"},

		// Brand Patterns (common business suffixes)
		{`\b(inc\.|incorporated|LLC|L\.L\.C\.|Ltd\.|Limited|Corp\.|Corporation)\b`, EntityTypeBrand, 0.80, "Business entity suffix"},
	}

	er.patternsMu.Lock()
	defer er.patternsMu.Unlock()

	er.patterns = make([]EntityPattern, 0, len(patterns))
	for _, p := range patterns {
		regex, err := regexp.Compile(p.pattern)
		if err == nil {
			er.patterns = append(er.patterns, EntityPattern{
				Pattern:     regex,
				Type:        p.entityType,
				Confidence:  p.confidence,
				Description: p.description,
			})
		}
	}
}

// AddPattern adds a custom pattern to the recognizer
func (er *EntityRecognizer) AddPattern(pattern string, entityType EntityType, confidence float64, description string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	er.patternsMu.Lock()
	defer er.patternsMu.Unlock()

	er.patterns = append(er.patterns, EntityPattern{
		Pattern:     regex,
		Type:        entityType,
		Confidence:  confidence,
		Description: description,
	})

	return nil
}

// GetEntitiesByType returns entities filtered by type
func (er *EntityRecognizer) GetEntitiesByType(entities []Entity, entityType EntityType) []Entity {
	result := []Entity{}
	for _, entity := range entities {
		if entity.Type == entityType {
			result = append(result, entity)
		}
	}
	return result
}

// GetEntityKeywords extracts keywords from entities for classification
func (er *EntityRecognizer) GetEntityKeywords(entities []Entity) []string {
	keywords := make([]string, 0, len(entities))
	seen := make(map[string]bool)

	for _, entity := range entities {
		// Extract individual words from entity text
		words := strings.Fields(entity.Text)
		for _, word := range words {
			word = strings.ToLower(strings.Trim(word, ".,!?;:\"'()[]{}"))
			if len(word) >= 3 && !seen[word] {
				seen[word] = true
				keywords = append(keywords, word)
			}
		}
	}

	return keywords
}

