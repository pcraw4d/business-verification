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
// Enhanced with 100+ patterns for comprehensive business entity recognition
func (er *EntityRecognizer) loadDefaultPatterns() {
	patterns := []struct {
		pattern     string
		entityType  EntityType
		confidence  float64
		description string
	}{
		// ============================================
		// BUSINESS TYPES (Expanded - 40+ patterns)
		// ============================================
		// Food & Beverage
		{`\b(wine\s+shop|wine\s+store|wine\s+merchant|wine\s+retailer|wine\s+bar|wine\s+cellar)\b`, EntityTypeBusinessType, 0.95, "Wine retail business"},
		{`\b(restaurant|cafe|coffee\s+shop|bistro|diner|tavern|eatery|gastropub)\b`, EntityTypeBusinessType, 0.95, "Restaurant"},
		{`\b(winery|vineyard|vintner|distillery|brewery)\b`, EntityTypeBusinessType, 0.95, "Beverage producer"},
		{`\b(bakery|patisserie|confectionery|dessert\s+shop)\b`, EntityTypeBusinessType, 0.95, "Bakery"},
		{`\b(food\s+truck|mobile\s+kitchen|street\s+food|food\s+cart)\b`, EntityTypeBusinessType, 0.90, "Food truck"},
		{`\b(catering|event\s+catering|catering\s+service)\b`, EntityTypeBusinessType, 0.90, "Catering service"},
		{`\b(pizzeria|pizza\s+restaurant|pizza\s+parlor)\b`, EntityTypeBusinessType, 0.90, "Pizzeria"},
		{`\b(sushi\s+restaurant|sushi\s+bar|japanese\s+restaurant)\b`, EntityTypeBusinessType, 0.90, "Sushi restaurant"},
		{`\b(steakhouse|grill|bbq|barbecue)\b`, EntityTypeBusinessType, 0.90, "Steakhouse/BBQ"},
		{`\b(fast\s+food|quick\s+service|drive\s+thru|drive\s+through)\b`, EntityTypeBusinessType, 0.90, "Fast food"},
		
		// Retail
		{`\b(retail\s+store|retail\s+shop|brick\s+and\s+mortar|physical\s+store|storefront)\b`, EntityTypeBusinessType, 0.95, "Retail store"},
		{`\b(online\s+store|online\s+shop|ecommerce\s+store|web\s+store|digital\s+storefront)\b`, EntityTypeBusinessType, 0.95, "Online store"},
		{`\b(boutique|specialty\s+store|niche\s+retailer)\b`, EntityTypeBusinessType, 0.90, "Boutique"},
		{`\b(department\s+store|big\s+box\s+store|superstore)\b`, EntityTypeBusinessType, 0.90, "Department store"},
		{`\b(grocery\s+store|supermarket|grocery\s+market)\b`, EntityTypeBusinessType, 0.95, "Grocery store"},
		{`\b(convenience\s+store|corner\s+store|convenience\s+shop)\b`, EntityTypeBusinessType, 0.90, "Convenience store"},
		{`\b(pharmacy|drugstore|chemist)\b`, EntityTypeBusinessType, 0.95, "Pharmacy"},
		{`\b(bookstore|book\s+shop|book\s+retailer)\b`, EntityTypeBusinessType, 0.90, "Bookstore"},
		{`\b(jewelry\s+store|jeweler|jewelry\s+shop)\b`, EntityTypeBusinessType, 0.90, "Jewelry store"},
		{`\b(furniture\s+store|furniture\s+retailer|home\s+furnishings)\b`, EntityTypeBusinessType, 0.90, "Furniture store"},
		{`\b(electronics\s+store|electronics\s+retailer|tech\s+store)\b`, EntityTypeBusinessType, 0.90, "Electronics store"},
		{`\b(clothing\s+store|apparel\s+store|fashion\s+retailer)\b`, EntityTypeBusinessType, 0.90, "Clothing store"},
		{`\b(hardware\s+store|home\s+improvement|hardware\s+retailer)\b`, EntityTypeBusinessType, 0.90, "Hardware store"},
		
		// Technology
		{`\b(technology\s+company|tech\s+firm|software\s+company|IT\s+services)\b`, EntityTypeBusinessType, 0.90, "Technology company"},
		{`\b(software\s+development|software\s+company|dev\s+studio)\b`, EntityTypeBusinessType, 0.90, "Software development"},
		{`\b(IT\s+consulting|tech\s+consulting|IT\s+services)\b`, EntityTypeBusinessType, 0.90, "IT consulting"},
		{`\b(cloud\s+services|cloud\s+provider|cloud\s+company)\b`, EntityTypeBusinessType, 0.90, "Cloud services"},
		{`\b(data\s+analytics|analytics\s+company|data\s+services)\b`, EntityTypeBusinessType, 0.90, "Data analytics"},
		{`\b(cybersecurity|security\s+services|cyber\s+security)\b`, EntityTypeBusinessType, 0.90, "Cybersecurity"},
		{`\b(ai\s+company|artificial\s+intelligence|machine\s+learning)\b`, EntityTypeBusinessType, 0.90, "AI/ML company"},
		
		// Financial Services
		{`\b(financial\s+services|bank|credit\s+union|investment\s+firm)\b`, EntityTypeBusinessType, 0.90, "Financial services"},
		{`\b(investment\s+bank|investment\s+firm|asset\s+management)\b`, EntityTypeBusinessType, 0.90, "Investment firm"},
		{`\b(insurance\s+company|insurer|insurance\s+agency)\b`, EntityTypeBusinessType, 0.90, "Insurance company"},
		{`\b(mortgage\s+company|mortgage\s+lender|mortgage\s+broker)\b`, EntityTypeBusinessType, 0.90, "Mortgage company"},
		{`\b(fintech|financial\s+technology|payment\s+processor)\b`, EntityTypeBusinessType, 0.90, "Fintech"},
		
		// Healthcare
		{`\b(healthcare\s+provider|medical\s+clinic|hospital|pharmacy)\b`, EntityTypeBusinessType, 0.90, "Healthcare provider"},
		{`\b(medical\s+practice|doctor\s+office|physician\s+practice)\b`, EntityTypeBusinessType, 0.90, "Medical practice"},
		{`\b(dental\s+clinic|dentist|dental\s+practice)\b`, EntityTypeBusinessType, 0.90, "Dental clinic"},
		{`\b(veterinary|vet\s+clinic|animal\s+hospital)\b`, EntityTypeBusinessType, 0.90, "Veterinary"},
		{`\b(pharmaceutical|pharma\s+company|drug\s+manufacturer)\b`, EntityTypeBusinessType, 0.90, "Pharmaceutical"},
		
		// Real Estate & Construction
		{`\b(real\s+estate|property\s+management|realty|realtor)\b`, EntityTypeBusinessType, 0.90, "Real estate"},
		{`\b(construction\s+company|contractor|builder|construction\s+firm)\b`, EntityTypeBusinessType, 0.90, "Construction company"},
		{`\b(architecture\s+firm|architectural\s+services|architect)\b`, EntityTypeBusinessType, 0.90, "Architecture firm"},
		{`\b(engineering\s+firm|engineering\s+services|engineer)\b`, EntityTypeBusinessType, 0.90, "Engineering firm"},
		{`\b(property\s+development|real\s+estate\s+development)\b`, EntityTypeBusinessType, 0.90, "Property development"},
		
		// Hospitality & Travel
		{`\b(hotel|resort|inn|lodge|motel|hostel)\b`, EntityTypeBusinessType, 0.95, "Hotel"},
		{`\b(travel\s+agency|travel\s+agent|tour\s+operator)\b`, EntityTypeBusinessType, 0.90, "Travel agency"},
		{`\b(airline|air\s+carrier|aviation\s+services)\b`, EntityTypeBusinessType, 0.90, "Airline"},
		{`\b(cruise\s+line|cruise\s+company|cruise\s+ship)\b`, EntityTypeBusinessType, 0.90, "Cruise line"},
		
		// Manufacturing & Industrial
		{`\b(manufacturing|manufacturer|production\s+facility)\b`, EntityTypeBusinessType, 0.90, "Manufacturing"},
		{`\b(industrial|industrial\s+services|industrial\s+equipment)\b`, EntityTypeBusinessType, 0.90, "Industrial"},
		{`\b(automotive|auto\s+manufacturer|car\s+company)\b`, EntityTypeBusinessType, 0.90, "Automotive"},
		
		// Transportation & Logistics
		{`\b(trucking|trucking\s+company|freight\s+company)\b`, EntityTypeBusinessType, 0.90, "Trucking"},
		{`\b(shipping\s+company|freight\s+forwarder|logistics\s+company)\b`, EntityTypeBusinessType, 0.90, "Shipping company"},
		{`\b(courier|delivery\s+service|package\s+delivery)\b`, EntityTypeBusinessType, 0.90, "Courier service"},
		
		// Education
		{`\b(school|university|college|academy|institute)\b`, EntityTypeBusinessType, 0.95, "Educational institution"},
		{`\b(training\s+company|training\s+institute|professional\s+training)\b`, EntityTypeBusinessType, 0.90, "Training company"},
		
		// Entertainment & Media
		{`\b(entertainment|media\s+company|production\s+company)\b`, EntityTypeBusinessType, 0.90, "Entertainment"},
		{`\b(movie\s+theater|cinema|film\s+theater)\b`, EntityTypeBusinessType, 0.90, "Movie theater"},
		{`\b(gaming\s+company|video\s+game\s+company|game\s+developer)\b`, EntityTypeBusinessType, 0.90, "Gaming company"},
		
		// ============================================
		// SERVICES (Expanded - 30+ patterns)
		// ============================================
		{`\b(consulting\s+services|advisory\s+services|consulting\s+firm)\b`, EntityTypeService, 0.85, "Consulting services"},
		{`\b(legal\s+services|law\s+firm|attorney|lawyer)\b`, EntityTypeService, 0.85, "Legal services"},
		{`\b(accounting\s+services|accountant|bookkeeping|tax\s+preparation)\b`, EntityTypeService, 0.85, "Accounting services"},
		{`\b(marketing\s+services|advertising\s+agency|marketing\s+firm)\b`, EntityTypeService, 0.85, "Marketing services"},
		{`\b(web\s+design|web\s+development|software\s+development)\b`, EntityTypeService, 0.85, "Web/software services"},
		{`\b(logistics|shipping|freight|delivery\s+services)\b`, EntityTypeService, 0.85, "Logistics services"},
		{`\b(education\s+services|training|tutoring|educational\s+institution)\b`, EntityTypeService, 0.85, "Education services"},
		{`\b(cleaning\s+services|janitorial|cleaning\s+company)\b`, EntityTypeService, 0.85, "Cleaning services"},
		{`\b(landscaping|lawn\s+care|garden\s+services)\b`, EntityTypeService, 0.85, "Landscaping services"},
		{`\b(plumbing|plumber|plumbing\s+services)\b`, EntityTypeService, 0.85, "Plumbing services"},
		{`\b(electrical\s+services|electrician|electrical\s+contractor)\b`, EntityTypeService, 0.85, "Electrical services"},
		{`\b(hvac|heating|cooling|air\s+conditioning)\b`, EntityTypeService, 0.85, "HVAC services"},
		{`\b(roofing|roofer|roofing\s+services)\b`, EntityTypeService, 0.85, "Roofing services"},
		{`\b(painting\s+services|painter|paint\s+contractor)\b`, EntityTypeService, 0.85, "Painting services"},
		{`\b(security\s+services|security\s+company|security\s+guard)\b`, EntityTypeService, 0.85, "Security services"},
		{`\b(printing\s+services|print\s+shop|printing\s+company)\b`, EntityTypeService, 0.85, "Printing services"},
		{`\b(photography|photographer|photo\s+services)\b`, EntityTypeService, 0.85, "Photography services"},
		{`\b(video\s+production|video\s+services|videography)\b`, EntityTypeService, 0.85, "Video production"},
		{`\b(event\s+planning|event\s+management|event\s+services)\b`, EntityTypeService, 0.85, "Event planning"},
		{`\b(interior\s+design|interior\s+designer|design\s+services)\b`, EntityTypeService, 0.85, "Interior design"},
		{`\b(graphic\s+design|graphic\s+designer|design\s+agency)\b`, EntityTypeService, 0.85, "Graphic design"},
		{`\b(public\s+relations|PR\s+firm|public\s+relations\s+agency)\b`, EntityTypeService, 0.85, "Public relations"},
		{`\b(human\s+resources|HR\s+services|recruiting|recruitment)\b`, EntityTypeService, 0.85, "HR services"},
		{`\b(payroll\s+services|payroll\s+processing|payroll\s+company)\b`, EntityTypeService, 0.85, "Payroll services"},
		{`\b(insurance\s+agency|insurance\s+broker|insurance\s+services)\b`, EntityTypeService, 0.85, "Insurance services"},
		{`\b(real\s+estate\s+services|realty\s+services|property\s+services)\b`, EntityTypeService, 0.85, "Real estate services"},
		{`\b(property\s+management|property\s+manager|property\s+services)\b`, EntityTypeService, 0.85, "Property management"},
		{`\b(waste\s+management|waste\s+services|garbage\s+collection)\b`, EntityTypeService, 0.85, "Waste management"},
		{`\b(utilities|utility\s+services|utility\s+company)\b`, EntityTypeService, 0.85, "Utility services"},
		{`\b(telecommunications|telecom|phone\s+services)\b`, EntityTypeService, 0.85, "Telecommunications"},
		
		// ============================================
		// PRODUCTS (Expanded - 20+ patterns)
		// ============================================
		{`\b(wine|wines|vintage\s+wine|fine\s+wine|premium\s+wine)\b`, EntityTypeProduct, 0.90, "Wine products"},
		{`\b(alcohol|spirits|liquor|beer|beverages)\b`, EntityTypeProduct, 0.85, "Beverage products"},
		{`\b(software|application|app|platform|system)\b`, EntityTypeProduct, 0.80, "Software products"},
		{`\b(consumer\s+goods|products|merchandise|inventory)\b`, EntityTypeProduct, 0.75, "Consumer goods"},
		{`\b(food\s+products|food\s+items|grocery\s+products)\b`, EntityTypeProduct, 0.85, "Food products"},
		{`\b(clothing|apparel|garments|fashion)\b`, EntityTypeProduct, 0.85, "Clothing"},
		{`\b(electronics|electronic\s+devices|tech\s+products)\b`, EntityTypeProduct, 0.85, "Electronics"},
		{`\b(furniture|furnishings|home\s+furniture)\b`, EntityTypeProduct, 0.85, "Furniture"},
		{`\b(automotive\s+parts|car\s+parts|auto\s+accessories)\b`, EntityTypeProduct, 0.85, "Automotive parts"},
		{`\b(medical\s+devices|medical\s+equipment|healthcare\s+products)\b`, EntityTypeProduct, 0.85, "Medical devices"},
		{`\b(pharmaceuticals|drugs|medications|medicine)\b`, EntityTypeProduct, 0.85, "Pharmaceuticals"},
		{`\b(cosmetics|beauty\s+products|personal\s+care)\b`, EntityTypeProduct, 0.85, "Cosmetics"},
		{`\b(toys|games|children\s+products|toy\s+products)\b`, EntityTypeProduct, 0.85, "Toys"},
		{`\b(sports\s+equipment|athletic\s+gear|sporting\s+goods)\b`, EntityTypeProduct, 0.85, "Sports equipment"},
		{`\b(books|publications|printed\s+materials)\b`, EntityTypeProduct, 0.85, "Books"},
		{`\b(machinery|industrial\s+equipment|manufacturing\s+equipment)\b`, EntityTypeProduct, 0.85, "Machinery"},
		{`\b(building\s+materials|construction\s+materials|construction\s+supplies)\b`, EntityTypeProduct, 0.85, "Building materials"},
		{`\b(chemicals|chemical\s+products|industrial\s+chemicals)\b`, EntityTypeProduct, 0.85, "Chemicals"},
		{`\b(energy|power|electricity|renewable\s+energy)\b`, EntityTypeProduct, 0.85, "Energy"},
		{`\b(raw\s+materials|commodities|basic\s+materials)\b`, EntityTypeProduct, 0.85, "Raw materials"},
		
		// ============================================
		// INDUSTRY INDICATORS (Expanded - 25+ patterns)
		// ============================================
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
		{`\b(energy\s+industry|oil\s+and\s+gas|energy\s+sector)\b`, EntityTypeIndustry, 0.90, "Energy industry"},
		{`\b(telecommunications\s+industry|telecom\s+industry|communications)\b`, EntityTypeIndustry, 0.90, "Telecommunications industry"},
		{`\b(media\s+industry|entertainment\s+industry|media\s+sector)\b`, EntityTypeIndustry, 0.90, "Media industry"},
		{`\b(education\s+industry|educational\s+sector|education\s+services)\b`, EntityTypeIndustry, 0.90, "Education industry"},
		{`\b(hospitality\s+industry|tourism|hospitality\s+sector)\b`, EntityTypeIndustry, 0.90, "Hospitality industry"},
		{`\b(automotive\s+industry|auto\s+industry|automotive\s+sector)\b`, EntityTypeIndustry, 0.90, "Automotive industry"},
		{`\b(aerospace\s+industry|aviation\s+industry|aerospace\s+sector)\b`, EntityTypeIndustry, 0.90, "Aerospace industry"},
		{`\b(pharmaceutical\s+industry|pharma\s+industry|pharmaceutical\s+sector)\b`, EntityTypeIndustry, 0.90, "Pharmaceutical industry"},
		{`\b(chemical\s+industry|chemicals\s+sector|chemical\s+manufacturing)\b`, EntityTypeIndustry, 0.90, "Chemical industry"},
		{`\b(textiles\s+industry|textile\s+industry|apparel\s+industry)\b`, EntityTypeIndustry, 0.90, "Textiles industry"},
		{`\b(mining\s+industry|mining\s+sector|extraction\s+industry)\b`, EntityTypeIndustry, 0.90, "Mining industry"},
		{`\b(utilities\s+industry|utility\s+industry|public\s+utilities)\b`, EntityTypeIndustry, 0.90, "Utilities industry"},
		{`\b(waste\s+management\s+industry|waste\s+industry|recycling\s+industry)\b`, EntityTypeIndustry, 0.90, "Waste management industry"},
		{`\b(professional\s+services\s+industry|professional\s+services\s+sector)\b`, EntityTypeIndustry, 0.90, "Professional services industry"},
		{`\b(wholesale\s+trade|wholesale\s+industry|wholesale\s+sector)\b`, EntityTypeIndustry, 0.90, "Wholesale trade"},
		
		// ============================================
		// LOCATION ENTITIES (Expanded - 15+ patterns)
		// ============================================
		{`\b(united\s+states|USA|US|America)\b`, EntityTypeLocation, 0.95, "United States"},
		{`\b(canada|canadian)\b`, EntityTypeLocation, 0.95, "Canada"},
		{`\b(united\s+kingdom|UK|Britain|England)\b`, EntityTypeLocation, 0.95, "United Kingdom"},
		{`\b(australia|australian)\b`, EntityTypeLocation, 0.95, "Australia"},
		{`\b(europe|european|EU)\b`, EntityTypeLocation, 0.90, "Europe"},
		{`\b(california|CA|new\s+york|NY|texas|TX|florida|FL)\b`, EntityTypeLocation, 0.85, "US State"},
		{`\b(london|paris|tokyo|sydney|toronto|vancouver)\b`, EntityTypeLocation, 0.85, "Major city"},
		{`\b(new\s+york\s+city|NYC|los\s+angeles|LA|chicago|san\s+francisco|SF)\b`, EntityTypeLocation, 0.85, "Major US city"},
		{`\b(mexico|mexican)\b`, EntityTypeLocation, 0.90, "Mexico"},
		{`\b(asia|asian|asia\s+pacific|APAC)\b`, EntityTypeLocation, 0.90, "Asia"},
		{`\b(china|chinese)\b`, EntityTypeLocation, 0.90, "China"},
		{`\b(japan|japanese)\b`, EntityTypeLocation, 0.90, "Japan"},
		{`\b(germany|german)\b`, EntityTypeLocation, 0.90, "Germany"},
		{`\b(france|french)\b`, EntityTypeLocation, 0.90, "France"},
		{`\b(italy|italian)\b`, EntityTypeLocation, 0.90, "Italy"},
		{`\b(spain|spanish)\b`, EntityTypeLocation, 0.90, "Spain"},
		
		// ============================================
		// BRAND PATTERNS (Expanded - 10+ patterns)
		// ============================================
		{`\b(inc\.|incorporated|LLC|L\.L\.C\.|Ltd\.|Limited|Corp\.|Corporation)\b`, EntityTypeBrand, 0.80, "Business entity suffix"},
		{`\b(LLP|L\.L\.P\.|limited\s+liability\s+partnership)\b`, EntityTypeBrand, 0.80, "LLP suffix"},
		{`\b(PC|P\.C\.|professional\s+corporation)\b`, EntityTypeBrand, 0.80, "PC suffix"},
		{`\b(PA|P\.A\.|professional\s+association)\b`, EntityTypeBrand, 0.80, "PA suffix"},
		{`\b(co\.|company|co\s+ltd)\b`, EntityTypeBrand, 0.75, "Company suffix"},
		{`\b(group|holdings|enterprises)\b`, EntityTypeBrand, 0.75, "Business group suffix"},
		{`\b(solutions|systems|services|technologies)\b`, EntityTypeBrand, 0.70, "Business descriptor"},
		{`\b(global|international|worldwide)\b`, EntityTypeBrand, 0.70, "Global indicator"},
		{`\b(partners|partnership|associates)\b`, EntityTypeBrand, 0.70, "Partnership indicator"},
		{`\b(industries|industrial|manufacturing)\b`, EntityTypeBrand, 0.70, "Industrial indicator"},
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

