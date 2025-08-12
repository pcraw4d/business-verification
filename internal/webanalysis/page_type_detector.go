package webanalysis

import (
	"regexp"
	"strings"
	"time"
)

// PageTypeDetection represents the detected type of a web page
type PageTypeDetection struct {
	Type              string    `json:"type"`
	Confidence        float64   `json:"confidence"`
	DetectionMethod   string    `json:"detection_method"`
	Keywords          []string  `json:"keywords"`
	URLPatterns       []string  `json:"url_patterns"`
	ContentIndicators []string  `json:"content_indicators"`
	DetectedAt        time.Time `json:"detected_at"`
}

// PageTypeDetector manages page type detection operations
type PageTypeDetector struct {
	config            PageTypeConfig
	urlAnalyzer       *URLPatternAnalyzer
	contentAnalyzer   *ContentTypeAnalyzer
	structureAnalyzer *PageStructureAnalyzer
}

// PageTypeConfig holds configuration for page type detection
type PageTypeConfig struct {
	ConfidenceThresholds map[string]float64  `json:"confidence_thresholds"`
	URLPatterns          map[string][]string `json:"url_patterns"`
	ContentKeywords      map[string][]string `json:"content_keywords"`
	StructureIndicators  map[string][]string `json:"structure_indicators"`
	DetectionWeights     map[string]float64  `json:"detection_weights"`
}

// NewPageTypeDetector creates a new page type detector
func NewPageTypeDetector(config PageTypeConfig) *PageTypeDetector {
	return &PageTypeDetector{
		config:            config,
		urlAnalyzer:       NewURLPatternAnalyzer(),
		contentAnalyzer:   NewContentTypeAnalyzer(),
		structureAnalyzer: NewPageStructureAnalyzer(),
	}
}

// DetectPageType performs comprehensive page type detection
func (ptd *PageTypeDetector) DetectPageType(content *ScrapedContent) *PageTypeDetection {
	detection := &PageTypeDetection{
		DetectedAt: time.Now(),
	}

	// URL-based detection
	urlType := ptd.urlAnalyzer.AnalyzeURLPattern(content.URL)

	// Content-based detection
	contentType := ptd.contentAnalyzer.AnalyzeContent(content.Text, content.Title)

	// Structure-based detection
	structureType := ptd.structureAnalyzer.AnalyzeStructure(content.HTML)

	// Combine detections and determine final type
	finalType := ptd.combineDetections(urlType, contentType, structureType)

	detection.Type = finalType.Type
	detection.Confidence = finalType.Confidence
	detection.DetectionMethod = finalType.DetectionMethod
	detection.Keywords = finalType.Keywords
	detection.URLPatterns = finalType.URLPatterns
	detection.ContentIndicators = finalType.ContentIndicators

	return detection
}

// combineDetections combines multiple detection methods
func (ptd *PageTypeDetector) combineDetections(urlType, contentType, structureType *PageTypeDetection) *PageTypeDetection {
	// Weight the different detection methods
	urlWeight := ptd.config.DetectionWeights["url"]
	contentWeight := ptd.config.DetectionWeights["content"]
	structureWeight := ptd.config.DetectionWeights["structure"]

	// Create a map to track scores for each page type
	typeScores := make(map[string]float64)
	typeKeywords := make(map[string][]string)
	typeURLPatterns := make(map[string][]string)
	typeContentIndicators := make(map[string][]string)

	// Aggregate URL detection scores
	if urlType != nil && urlType.Confidence > 0 {
		typeScores[urlType.Type] += urlType.Confidence * urlWeight
		typeKeywords[urlType.Type] = append(typeKeywords[urlType.Type], urlType.Keywords...)
		typeURLPatterns[urlType.Type] = append(typeURLPatterns[urlType.Type], urlType.URLPatterns...)
	}

	// Aggregate content detection scores
	if contentType != nil && contentType.Confidence > 0 {
		typeScores[contentType.Type] += contentType.Confidence * contentWeight
		typeKeywords[contentType.Type] = append(typeKeywords[contentType.Type], contentType.Keywords...)
		typeContentIndicators[contentType.Type] = append(typeContentIndicators[contentType.Type], contentType.ContentIndicators...)
	}

	// Aggregate structure detection scores
	if structureType != nil && structureType.Confidence > 0 {
		typeScores[structureType.Type] += structureType.Confidence * structureWeight
		typeKeywords[structureType.Type] = append(typeKeywords[structureType.Type], structureType.Keywords...)
		typeContentIndicators[structureType.Type] = append(typeContentIndicators[structureType.Type], structureType.ContentIndicators...)
	}

	// Find the page type with the highest score
	var bestType string = "unknown"
	var bestScore float64

	for pageType, score := range typeScores {
		if score > bestScore {
			bestScore = score
			bestType = pageType
		}
	}

	// Determine detection method
	detectionMethod := "unknown"
	if urlType != nil && urlType.Confidence > 0 {
		detectionMethod = "url_pattern"
	} else if contentType != nil && contentType.Confidence > 0 {
		detectionMethod = "content_analysis"
	} else if structureType != nil && structureType.Confidence > 0 {
		detectionMethod = "structure_analysis"
	}

	return &PageTypeDetection{
		Type:              bestType,
		Confidence:        bestScore,
		DetectionMethod:   detectionMethod,
		Keywords:          typeKeywords[bestType],
		URLPatterns:       typeURLPatterns[bestType],
		ContentIndicators: typeContentIndicators[bestType],
		DetectedAt:        time.Now(),
	}
}

// URLPatternAnalyzer analyzes URL patterns for page type detection
type URLPatternAnalyzer struct{}

// NewURLPatternAnalyzer creates a new URL pattern analyzer
func NewURLPatternAnalyzer() *URLPatternAnalyzer {
	return &URLPatternAnalyzer{}
}

// AnalyzeURLPattern analyzes URL patterns to detect page types
func (upa *URLPatternAnalyzer) AnalyzeURLPattern(url string) *PageTypeDetection {
	url = strings.ToLower(url)

	// Define URL patterns for different page types
	patterns := map[string][]string{
		"about_us": {
			"about", "about-us", "aboutus", "about_us", "company", "who-we-are", "our-story",
			"about-company", "about-our-company", "about-the-company",
		},
		"mission": {
			"mission", "mission-vision", "vision", "values", "philosophy", "purpose",
			"mission-statement", "vision-statement", "our-mission", "our-vision",
		},
		"products": {
			"products", "product", "solutions", "offerings", "catalog", "inventory",
			"product-catalog", "product-list", "our-products", "product-category",
		},
		"services": {
			"services", "service", "solutions", "offerings", "what-we-do",
			"our-services", "service-list", "service-category", "consulting",
		},
		"contact": {
			"contact", "contact-us", "get-in-touch", "reach-us", "support",
			"contact-information", "get-quote", "request-quote",
		},
		"team": {
			"team", "our-team", "staff", "leadership", "management", "people",
			"meet-the-team", "team-members", "executives",
		},
		"careers": {
			"careers", "jobs", "employment", "work-with-us", "join-us",
			"career-opportunities", "job-openings", "hiring",
		},
		"news": {
			"news", "blog", "articles", "press", "media", "updates",
			"latest-news", "press-releases", "newsroom",
		},
	}

	// Check each pattern
	for pageType, patternList := range patterns {
		for _, pattern := range patternList {
			if strings.Contains(url, pattern) {
				return &PageTypeDetection{
					Type:            pageType,
					Confidence:      0.8,
					DetectionMethod: "url_pattern",
					Keywords:        []string{pattern},
					URLPatterns:     []string{pattern},
					DetectedAt:      time.Now(),
				}
			}
		}
	}

	// Check for path-based patterns
	pathPatterns := map[string]*regexp.Regexp{
		"about_us": regexp.MustCompile(`/(about|company|who-we-are)/?`),
		"mission":  regexp.MustCompile(`/(mission|vision|values|philosophy)/?`),
		"products": regexp.MustCompile(`/(products?|solutions|catalog)/?`),
		"services": regexp.MustCompile(`/(services?|solutions|what-we-do)/?`),
		"contact":  regexp.MustCompile(`/(contact|get-in-touch|support)/?`),
		"team":     regexp.MustCompile(`/(team|staff|leadership|people)/?`),
		"careers":  regexp.MustCompile(`/(careers?|jobs|employment)/?`),
		"news":     regexp.MustCompile(`/(news|blog|articles|press)/?`),
	}

	for pageType, pattern := range pathPatterns {
		if pattern.MatchString(url) {
			return &PageTypeDetection{
				Type:            pageType,
				Confidence:      0.7,
				DetectionMethod: "url_pattern",
				Keywords:        []string{pageType},
				URLPatterns:     []string{pattern.String()},
				DetectedAt:      time.Now(),
			}
		}
	}

	return &PageTypeDetection{
		Type:            "unknown",
		Confidence:      0.0,
		DetectionMethod: "url_pattern",
		DetectedAt:      time.Now(),
	}
}

// ContentTypeAnalyzer analyzes content for page type detection
type ContentTypeAnalyzer struct{}

// NewContentTypeAnalyzer creates a new content type analyzer
func NewContentTypeAnalyzer() *ContentTypeAnalyzer {
	return &ContentTypeAnalyzer{}
}

// AnalyzeContent analyzes content to detect page types
func (cta *ContentTypeAnalyzer) AnalyzeContent(text, title string) *PageTypeDetection {
	combinedText := strings.ToLower(text + " " + title)

	// Define content keywords for different page types
	contentKeywords := map[string][]string{
		"about_us": {
			"about us", "our company", "who we are", "our story", "company history",
			"founded", "established", "since", "heritage", "tradition", "background",
			"company overview", "about our company", "our mission", "our values",
		},
		"mission": {
			"mission", "vision", "values", "philosophy", "purpose", "goals",
			"mission statement", "vision statement", "core values", "principles",
			"our mission", "our vision", "what drives us", "why we exist",
		},
		"products": {
			"products", "product catalog", "our products", "product line",
			"product features", "product specifications", "product details",
			"catalog", "inventory", "offerings", "what we sell",
		},
		"services": {
			"services", "our services", "service offerings", "what we do",
			"service portfolio", "consulting services", "professional services",
			"solutions", "expertise", "capabilities", "service areas",
		},
		"contact": {
			"contact us", "get in touch", "reach us", "contact information",
			"phone number", "email address", "office location", "address",
			"get a quote", "request information", "support",
		},
		"team": {
			"our team", "team members", "staff", "leadership", "management",
			"meet the team", "executives", "professionals", "experts",
			"key personnel", "team profiles",
		},
		"careers": {
			"careers", "job opportunities", "employment", "work with us",
			"join our team", "open positions", "job openings", "hiring",
			"career development", "employee benefits",
		},
		"news": {
			"latest news", "press releases", "news articles", "blog posts",
			"media coverage", "company updates", "announcements",
			"industry news", "press room", "newsroom",
		},
	}

	// Check for keyword matches
	bestType := "unknown"
	bestScore := 0.0
	var matchedKeywords []string

	for pageType, keywords := range contentKeywords {
		score := 0.0
		var foundKeywords []string

		for _, keyword := range keywords {
			if strings.Contains(combinedText, keyword) {
				score += 0.2
				foundKeywords = append(foundKeywords, keyword)
			}
		}

		if score > bestScore {
			bestScore = score
			bestType = pageType
			matchedKeywords = foundKeywords
		}
	}

	// Normalize confidence score
	confidence := bestScore
	if confidence > 1.0 {
		confidence = 1.0
	}

	return &PageTypeDetection{
		Type:              bestType,
		Confidence:        confidence,
		DetectionMethod:   "content_analysis",
		Keywords:          matchedKeywords,
		ContentIndicators: matchedKeywords,
		DetectedAt:        time.Now(),
	}
}

// PageStructureAnalyzer analyzes page structure for type detection
type PageStructureAnalyzer struct{}

// NewPageStructureAnalyzer creates a new page structure analyzer
func NewPageStructureAnalyzer() *PageStructureAnalyzer {
	return &PageStructureAnalyzer{}
}

// AnalyzeStructure analyzes HTML structure to detect page types
func (psa *PageStructureAnalyzer) AnalyzeStructure(html string) *PageTypeDetection {
	html = strings.ToLower(html)

	// Define structure indicators for different page types
	structureIndicators := map[string][]string{
		"about_us": {
			"<h1>about", "<h2>about", "<h3>about", "about us",
			"company history", "our story", "who we are",
		},
		"mission": {
			"<h1>mission", "<h2>mission", "<h3>mission",
			"<h1>vision", "<h2>vision", "<h3>vision",
			"mission statement", "vision statement", "our values",
		},
		"products": {
			"<h1>products", "<h2>products", "<h3>products",
			"product catalog", "our products", "product list",
			"<ul>", "<li>product", "product category",
		},
		"services": {
			"<h1>services", "<h2>services", "<h3>services",
			"our services", "service list", "what we do",
			"<ul>", "<li>service", "service category",
		},
		"contact": {
			"<h1>contact", "<h2>contact", "<h3>contact",
			"contact form", "contact information", "get in touch",
			"phone", "email", "address",
		},
		"team": {
			"<h1>team", "<h2>team", "<h3>team",
			"our team", "team members", "leadership",
			"staff profiles", "meet the team",
		},
		"careers": {
			"<h1>careers", "<h2>careers", "<h3>careers",
			"job opportunities", "open positions", "join us",
			"career", "employment", "work with us",
		},
		"news": {
			"<h1>news", "<h2>news", "<h3>news",
			"blog", "articles", "press releases",
			"latest news", "newsroom", "media",
		},
	}

	// Check for structure indicators
	bestType := "unknown"
	bestScore := 0.0
	var matchedIndicators []string

	for pageType, indicators := range structureIndicators {
		score := 0.0
		var foundIndicators []string

		for _, indicator := range indicators {
			if strings.Contains(html, indicator) {
				score += 0.15
				foundIndicators = append(foundIndicators, indicator)
			}
		}

		if score > bestScore {
			bestScore = score
			bestType = pageType
			matchedIndicators = foundIndicators
		}
	}

	// Normalize confidence score
	confidence := bestScore
	if confidence > 1.0 {
		confidence = 1.0
	}

	return &PageTypeDetection{
		Type:              bestType,
		Confidence:        confidence,
		DetectionMethod:   "structure_analysis",
		Keywords:          matchedIndicators,
		ContentIndicators: matchedIndicators,
		DetectedAt:        time.Now(),
	}
}

// GetPageTypePriority returns the priority score for a page type
func (ptd *PageTypeDetector) GetPageTypePriority(pageType string) float64 {
	priorities := map[string]float64{
		"about_us": 0.9,  // High priority - essential business information
		"mission":  0.8,  // High priority - business purpose and values
		"services": 0.85, // High priority - what the business does
		"products": 0.8,  // High priority - what the business sells
		"contact":  0.7,  // Medium-high priority - contact information
		"team":     0.6,  // Medium priority - team information
		"careers":  0.4,  // Lower priority - job opportunities
		"news":     0.3,  // Lower priority - news and updates
		"unknown":  0.1,  // Lowest priority - unknown page type
	}

	if priority, exists := priorities[pageType]; exists {
		return priority
	}
	return 0.1
}

// IsHighPriorityPageType checks if a page type is considered high priority
func (ptd *PageTypeDetector) IsHighPriorityPageType(pageType string) bool {
	priority := ptd.GetPageTypePriority(pageType)
	return priority >= 0.7
}

// GetPageTypeDescription returns a human-readable description of the page type
func (ptd *PageTypeDetector) GetPageTypeDescription(pageType string) string {
	descriptions := map[string]string{
		"about_us": "About Us/Company Information Page",
		"mission":  "Mission/Vision/Values Page",
		"services": "Services/Solutions Page",
		"products": "Products/Catalog Page",
		"contact":  "Contact Information Page",
		"team":     "Team/Leadership Page",
		"careers":  "Careers/Employment Page",
		"news":     "News/Blog/Media Page",
		"unknown":  "Unknown Page Type",
	}

	if description, exists := descriptions[pageType]; exists {
		return description
	}
	return "Unknown Page Type"
}
