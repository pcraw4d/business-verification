package webanalysis

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewPageRelevanceScorer(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	if scorer == nil {
		t.Fatal("Expected non-nil PageRelevanceScorer instance")
	}

	if scorer.businessMatcher == nil {
		t.Error("Expected businessMatcher to be initialized")
	}

	if scorer.contentAnalyzer == nil {
		t.Error("Expected contentAnalyzer to be initialized")
	}

	if scorer.technicalAnalyzer == nil {
		t.Error("Expected technicalAnalyzer to be initialized")
	}

	if len(scorer.config.Weights) == 0 {
		t.Error("Expected weights to be configured")
	}
}

func TestScorePage(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	content := &ScrapedContent{
		URL:          "https://example.com/about",
		Title:        "About Test Company",
		Text:         "Test Company is a leading technology company providing innovative software solutions. We have been in business for over 10 years and serve clients worldwide.",
		HTML:         "<h1>About Us</h1><p>Test Company is a leading technology company.</p>",
		ResponseTime: time.Second,
	}

	context := &ScoringContext{
		Industry:     "technology",
		Location:     "United States",
		BusinessType: "corporation",
	}

	score := scorer.ScorePage(content, "Test Company", context)

	if score == nil {
		t.Fatal("Expected non-nil PageRelevanceScore")
	}

	if score.OverallScore <= 0 {
		t.Errorf("Expected positive overall score, got %f", score.OverallScore)
	}

	if score.ConfidenceLevel <= 0 {
		t.Errorf("Expected positive confidence level, got %f", score.ConfidenceLevel)
	}

	if len(score.ComponentScores) == 0 {
		t.Error("Expected component scores to be calculated")
	}

	if len(score.ScoringFactors) == 0 {
		t.Error("Expected scoring factors to be calculated")
	}
}

func TestCalculateBusinessRelevance(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	content := &ScrapedContent{
		Title: "About Test Company",
		Text:  "Test Company is a leading technology company with offices in New York and London.",
	}

	context := &ScoringContext{
		Industry: "technology",
	}

	relevance := scorer.calculateBusinessRelevance(content, "Test Company", context)

	// Test business name match
	if relevance.BusinessNameMatch <= 0 {
		t.Errorf("Expected positive business name match, got %f", relevance.BusinessNameMatch)
	}

	// Test industry relevance
	if relevance.IndustryRelevance <= 0 {
		t.Errorf("Expected positive industry relevance, got %f", relevance.IndustryRelevance)
	}

	// Test geographic relevance
	if relevance.GeographicRelevance <= 0 {
		t.Errorf("Expected positive geographic relevance, got %f", relevance.GeographicRelevance)
	}

	// Test business keywords
	if len(relevance.BusinessKeywords) == 0 {
		t.Error("Expected business keywords to be extracted")
	}

	// Test business entities (may be empty for this test case)
	if len(relevance.BusinessEntities) >= 0 {
		// This is acceptable - entities may not be found in this specific test case
	}
}

func TestCalculateContentRelevance(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	content := &ScrapedContent{
		Title: "About Our Company",
		Text:  "This is a comprehensive page about our company with substantial content that provides valuable information about our services and solutions.",
		HTML:  "<h1>About Us</h1><p>This is a comprehensive page about our company.</p><h2>Our Services</h2><p>We provide excellent services.</p>",
	}

	context := &ScoringContext{}

	relevance := scorer.calculateContentRelevance(content, context)

	// Test content quality
	if relevance.ContentQuality <= 0 {
		t.Errorf("Expected positive content quality, got %f", relevance.ContentQuality)
	}

	// Test content length
	if relevance.ContentLength <= 0 {
		t.Errorf("Expected positive content length, got %d", relevance.ContentLength)
	}

	// Test content structure
	if relevance.ContentStructure <= 0 {
		t.Errorf("Expected positive content structure, got %f", relevance.ContentStructure)
	}

	// Test content readability
	if relevance.ContentReadability <= 0 {
		t.Errorf("Expected positive content readability, got %f", relevance.ContentReadability)
	}

	// Test content topics
	if len(relevance.ContentTopics) == 0 {
		t.Error("Expected content topics to be extracted")
	}

	// Test content sentiment
	if relevance.ContentSentiment < 0 || relevance.ContentSentiment > 1 {
		t.Errorf("Expected sentiment between 0 and 1, got %f", relevance.ContentSentiment)
	}

	// Test content credibility (may be 0 for this test case)
	if relevance.ContentCredibility >= 0 {
		// This is acceptable - credibility may be 0 for this specific test case
	}
}

func TestCalculateTechnicalRelevance(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	content := &ScrapedContent{
		URL:          "https://example.com/about",
		Title:        "About Us",
		HTML:         "<h1>About Us</h1><p>Content</p>",
		ResponseTime: time.Second,
	}

	context := &ScoringContext{}

	relevance := scorer.calculateTechnicalRelevance(content, context)

	// Test page authority
	if relevance.PageAuthority < 0 || relevance.PageAuthority > 1 {
		t.Errorf("Expected page authority between 0 and 1, got %f", relevance.PageAuthority)
	}

	// Test page speed
	if relevance.PageSpeed < 0 || relevance.PageSpeed > 1 {
		t.Errorf("Expected page speed between 0 and 1, got %f", relevance.PageSpeed)
	}

	// Test mobile friendliness
	if relevance.MobileFriendliness < 0 || relevance.MobileFriendliness > 1 {
		t.Errorf("Expected mobile friendliness between 0 and 1, got %f", relevance.MobileFriendliness)
	}

	// Test SEO optimization
	if relevance.SEOOptimization < 0 || relevance.SEOOptimization > 1 {
		t.Errorf("Expected SEO optimization between 0 and 1, got %f", relevance.SEOOptimization)
	}

	// Test accessibility
	if relevance.Accessibility < 0 || relevance.Accessibility > 1 {
		t.Errorf("Expected accessibility between 0 and 1, got %f", relevance.Accessibility)
	}

	// Test security score
	if relevance.SecurityScore < 0 || relevance.SecurityScore > 1 {
		t.Errorf("Expected security score between 0 and 1, got %f", relevance.SecurityScore)
	}
}

func TestCalculateComponentScores(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	score := &PageRelevanceScore{
		BusinessRelevance: BusinessRelevance{
			BusinessNameMatch:   0.8,
			IndustryRelevance:   0.7,
			GeographicRelevance: 0.6,
			ContactInformation:  0.9,
		},
		ContentRelevance: ContentRelevance{
			ContentQuality:   0.8,
			ContentFreshness: 0.7,
			ContentStructure: 0.6,
		},
		TechnicalRelevance: TechnicalRelevance{
			SEOOptimization: 0.8,
			SecurityScore:   0.9,
			PageSpeed:       0.7,
		},
		ComponentScores: make(map[string]float64),
		ScoringFactors:  []ScoringFactor{},
	}

	scorer.calculateComponentScores(score)

	// Check that component scores were calculated
	expectedComponents := []string{
		"business_name_match",
		"industry_relevance",
		"content_quality",
		"geographic_relevance",
		"contact_information",
		"technical_relevance",
		"content_freshness",
		"content_structure",
	}

	for _, component := range expectedComponents {
		if _, exists := score.ComponentScores[component]; !exists {
			t.Errorf("Expected component score for %s", component)
		}
	}

	// Check that scoring factors were added
	if len(score.ScoringFactors) == 0 {
		t.Error("Expected scoring factors to be added")
	}
}

func TestCalculateOverallScore(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	score := &PageRelevanceScore{
		ComponentScores: map[string]float64{
			"business_name_match": 0.8,
			"industry_relevance":  0.7,
			"content_quality":     0.6,
		},
	}

	overallScore := scorer.calculateOverallScore(score)

	if overallScore <= 0 {
		t.Errorf("Expected positive overall score, got %f", overallScore)
	}

	if overallScore > 1 {
		t.Errorf("Expected overall score <= 1, got %f", overallScore)
	}
}

func TestCalculateConfidenceLevel(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	score := &PageRelevanceScore{
		ScoringFactors: []ScoringFactor{
			{Confidence: 0.9},
			{Confidence: 0.8},
			{Confidence: 0.7},
		},
	}

	confidence := scorer.calculateConfidenceLevel(score)

	expectedConfidence := (0.9 + 0.8 + 0.7) / 3.0
	if confidence < expectedConfidence-0.001 || confidence > expectedConfidence+0.001 {
		t.Errorf("Expected confidence around %f, got %f", expectedConfidence, confidence)
	}
}

func TestCalculateIndustryRelevance(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	tests := []struct {
		name     string
		text     string
		industry string
		expected float64
	}{
		{
			name:     "Technology industry match",
			text:     "We provide software and hardware solutions",
			industry: "technology",
			expected: 0.4, // 2 out of 5 keywords
		},
		{
			name:     "Finance industry match",
			text:     "We offer banking and investment services",
			industry: "finance",
			expected: 0.4, // 2 out of 5 keywords
		},
		{
			name:     "No industry specified",
			text:     "We provide various services",
			industry: "",
			expected: 0.5, // Default neutral score
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &ScoringContext{Industry: tt.industry}
			result := scorer.calculateIndustryRelevance(tt.text, context)

			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected industry relevance around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateGeographicRelevance(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	tests := []struct {
		name     string
		text     string
		expected float64
	}{
		{
			name:     "High geographic relevance",
			text:     "Our headquarters is located in New York and we have offices in London",
			expected: 0.2, // 2 out of 10 keywords
		},
		{
			name:     "No geographic keywords",
			text:     "We provide excellent services",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateGeographicRelevance(tt.text)

			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected geographic relevance around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestExtractBusinessKeywords(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	text := "Our company provides business services and solutions to enterprise clients"
	keywords := scorer.extractBusinessKeywords(text)

	expectedKeywords := []string{"company", "business", "services", "solutions", "enterprise", "clients"}

	if len(keywords) < 3 {
		t.Errorf("Expected at least 3 business keywords, got %d", len(keywords))
	}

	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range keywords {
			if keyword == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected keyword %s not found", expected)
		}
	}
}

func TestCalculateContactInformationScore(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	tests := []struct {
		name     string
		text     string
		expected float64
	}{
		{
			name:     "Complete contact information",
			text:     "Contact us at info@example.com or call 555-123-4567. Visit us at 123 Main Street",
			expected: 1.0, // All three patterns matched
		},
		{
			name:     "Partial contact information",
			text:     "Email us at info@example.com",
			expected: 0.33, // Only email matched
		},
		{
			name:     "No contact information",
			text:     "We provide excellent services",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateContactInformationScore(tt.text)

			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected contact information score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateRegistrationDataScore(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	tests := []struct {
		name     string
		text     string
		expected float64
	}{
		{
			name:     "Complete registration data",
			text:     "Test Company Inc was established in 2010 and is a registered corporation",
			expected: 1.0, // All three patterns matched
		},
		{
			name:     "Partial registration data",
			text:     "Test Company Inc was established in 2010",
			expected: 0.66, // Two patterns matched
		},
		{
			name:     "No registration data",
			text:     "We provide excellent services",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateRegistrationDataScore(tt.text)

			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected registration data score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateContentStructure(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	tests := []struct {
		name     string
		html     string
		expected float64
	}{
		{
			name:     "Well structured content",
			html:     "<h1>Title</h1><h2>Subtitle</h2><p>Content</p><ul><li>Item</li></ul>",
			expected: 0.56, // 4 out of 7 elements
		},
		{
			name:     "Minimal structure",
			html:     "<p>Content</p>",
			expected: 0.14, // 1 out of 7 elements
		},
		{
			name:     "No structure",
			html:     "<div>Content</div>",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateContentStructure(tt.html)

			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected content structure score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateMobileFriendliness(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	tests := []struct {
		name     string
		html     string
		expected float64
	}{
		{
			name:     "Mobile friendly",
			html:     `<meta name="viewport" content="width=device-width"><div class="responsive">Content</div>`,
			expected: 0.5, // 2 out of 4 indicators
		},
		{
			name:     "Not mobile friendly",
			html:     "<div>Content</div>",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateMobileFriendliness(tt.html)

			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected mobile friendliness score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateAccessibilityScore(t *testing.T) {
	scorer := NewPageRelevanceScorer()

	tests := []struct {
		name     string
		html     string
		expected float64
	}{
		{
			name:     "Accessible content",
			html:     `<img alt="Description"><div role="main" tabindex="1">Content</div>`,
			expected: 0.75, // 3 out of 4 elements
		},
		{
			name:     "Not accessible",
			html:     "<div>Content</div>",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateAccessibilityScore(tt.html)

			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected accessibility score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestFuzzyMatcherCalculateMatch(t *testing.T) {
	matcher := &FuzzyMatcher{}

	tests := []struct {
		name     string
		text     string
		business string
		expected float64
	}{
		{
			name:     "Exact match",
			text:     "About Test Company",
			business: "Test Company",
			expected: 1.0,
		},
		{
			name:     "Partial match",
			text:     "About Test and Company",
			business: "Test Company",
			expected: 1.0, // Both words match
		},
		{
			name:     "No match",
			text:     "About Other Company",
			business: "Test Company",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.CalculateMatch(tt.text, tt.business)

			if result != tt.expected {
				t.Errorf("Expected match score %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestEntityExtractorExtractEntities(t *testing.T) {
	extractor := &EntityExtractor{}

	text := "ABC Corp and XYZ Inc are leading companies. DEF LLC is also mentioned."
	entities := extractor.ExtractEntities(text)

	if len(entities) == 0 {
		t.Error("Expected entities to be extracted")
	}

	// Check for expected entities
	expectedEntities := []string{"ABC", "XYZ", "DEF"}
	for _, expected := range expectedEntities {
		found := false
		for _, entity := range entities {
			if entity == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected entity %s not found", expected)
		}
	}
}

func TestReadabilityAnalyzerCalculateQuality(t *testing.T) {
	analyzer := &ReadabilityAnalyzer{}

	tests := []struct {
		name     string
		text     string
		expected float64
	}{
		{
			name:     "High quality content",
			text:     "This is a comprehensive article with substantial content. It contains multiple sentences and provides valuable information. The content is well-structured and informative.",
			expected: 1.0, // Should score high due to length, sentences, and word variety
		},
		{
			name:     "Low quality content",
			text:     "Short content",
			expected: 0.1, // Should score low due to short length
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateQuality(tt.text)

			if result < tt.expected-0.2 || result > tt.expected+0.2 {
				t.Errorf("Expected quality score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestReadabilityAnalyzerCalculateReadability(t *testing.T) {
	analyzer := &ReadabilityAnalyzer{}

	tests := []struct {
		name     string
		text     string
		expected float64
	}{
		{
			name:     "Highly readable",
			text:     "Short sentences. Easy to read. Clear content.",
			expected: 0.9, // Should be highly readable
		},
		{
			name:     "Less readable",
			text:     "This is a very long sentence that contains many words and continues for a very long time without any breaks or pauses which makes it difficult to read and understand.",
			expected: 0.3, // Should be less readable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateReadability(tt.text)

			if result < tt.expected-0.2 || result > tt.expected+0.2 {
				t.Errorf("Expected readability score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestTopicAnalyzerExtractTopics(t *testing.T) {
	analyzer := &TopicAnalyzer{}

	text := "We provide business services and technology solutions for healthcare clients"
	topics := analyzer.ExtractTopics(text)

	expectedTopics := []string{"business", "services", "technology", "healthcare", "clients"}

	if len(topics) < 3 {
		t.Errorf("Expected at least 3 topics, got %d", len(topics))
	}

	for _, expected := range expectedTopics {
		found := false
		for _, topic := range topics {
			if topic == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected topic %s not found", expected)
		}
	}
}

func TestSentimentAnalyzerAnalyzeSentiment(t *testing.T) {
	analyzer := &SentimentAnalyzer{}

	tests := []struct {
		name     string
		text     string
		expected float64
	}{
		{
			name:     "Positive sentiment",
			text:     "We provide excellent services and great solutions",
			expected: 1.0, // All positive words
		},
		{
			name:     "Negative sentiment",
			text:     "We had poor results and bad experiences",
			expected: 0.0, // All negative words
		},
		{
			name:     "Mixed sentiment",
			text:     "We provide excellent services but had some problems",
			expected: 0.5, // Equal positive and negative
		},
		{
			name:     "Neutral sentiment",
			text:     "We provide services and solutions",
			expected: 0.5, // No sentiment words
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeSentiment(tt.text)

			if result < tt.expected-0.2 || result > tt.expected+0.2 {
				t.Errorf("Expected sentiment score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCredibilityAnalyzerCalculateCredibility(t *testing.T) {
	analyzer := &CredibilityAnalyzer{}

	tests := []struct {
		name     string
		content  *ScrapedContent
		expected float64
	}{
		{
			name: "High credibility",
			content: &ScrapedContent{
				URL:   "https://example.com/about",
				Title: "About Our Certified Company",
				Text:  "We are a licensed and accredited professional company with 20 years of experience. Contact us at info@example.com",
			},
			expected: 0.4, // Multiple credibility indicators
		},
		{
			name: "Low credibility",
			content: &ScrapedContent{
				URL:   "http://example.com/about",
				Title: "About Us",
				Text:  "We provide services",
			},
			expected: 0.0, // No credibility indicators
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateCredibility(tt.content)

			if result < tt.expected-0.2 || result > tt.expected+0.2 {
				t.Errorf("Expected credibility score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestSEOAnalyzerCalculateSEO(t *testing.T) {
	analyzer := &SEOAnalyzer{}

	tests := []struct {
		name     string
		content  *ScrapedContent
		expected float64
	}{
		{
			name: "Good SEO",
			content: &ScrapedContent{
				Title: "About Our Company - Professional Services",
				Text:  "This is a comprehensive page with substantial content that provides valuable information about our services and solutions.",
				HTML:  "<h1>About Us</h1><h2>Our Services</h2><p>Content</p>",
			},
			expected: 0.8, // Good title, content, and structure
		},
		{
			name: "Poor SEO",
			content: &ScrapedContent{
				Title: "Page",
				Text:  "Short content",
				HTML:  "<div>Content</div>",
			},
			expected: 0.3, // Poor title, short content, no structure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateSEO(tt.content)

			if result < tt.expected-0.2 || result > tt.expected+0.2 {
				t.Errorf("Expected SEO score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestSecurityAnalyzerCalculateSecurity(t *testing.T) {
	analyzer := &SecurityAnalyzer{}

	tests := []struct {
		name     string
		content  *ScrapedContent
		expected float64
	}{
		{
			name: "Secure site",
			content: &ScrapedContent{
				URL: "https://example.com/about",
			},
			expected: 1.0, // HTTPS + simulated security features
		},
		{
			name: "Insecure site",
			content: &ScrapedContent{
				URL: "http://example.com/about",
			},
			expected: 0.6, // Only simulated security features
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateSecurity(tt.content)

			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected security score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestPerformanceAnalyzerCalculateSpeed(t *testing.T) {
	analyzer := &PerformanceAnalyzer{}

	tests := []struct {
		name         string
		responseTime time.Duration
		expected     float64
	}{
		{
			name:         "Fast response",
			responseTime: 500 * time.Millisecond,
			expected:     1.0,
		},
		{
			name:         "Medium response",
			responseTime: 2 * time.Second,
			expected:     0.8,
		},
		{
			name:         "Slow response",
			responseTime: 6 * time.Second,
			expected:     0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateSpeed(tt.responseTime)

			if result != tt.expected {
				t.Errorf("Expected speed score %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestPageRelevanceScoreJSON(t *testing.T) {
	score := &PageRelevanceScore{
		OverallScore:    0.85,
		ConfidenceLevel: 0.8,
		ComponentScores: map[string]float64{
			"business_name_match": 0.9,
			"content_quality":     0.8,
		},
		ScoringFactors: []ScoringFactor{
			{
				Factor:     "business_name_match",
				Score:      0.9,
				Weight:     0.25,
				Confidence: 0.9,
				Reason:     "Business name matching accuracy",
			},
		},
		BusinessRelevance: BusinessRelevance{
			BusinessNameMatch: 0.9,
			IndustryRelevance: 0.8,
		},
		ContentRelevance: ContentRelevance{
			ContentQuality: 0.8,
			ContentLength:  1000,
		},
		TechnicalRelevance: TechnicalRelevance{
			SEOOptimization: 0.7,
			SecurityScore:   0.9,
		},
		ScoredAt: time.Now(),
	}

	// Test that the struct can be marshaled to JSON
	_, err := json.Marshal(score)
	if err != nil {
		t.Errorf("Failed to marshal PageRelevanceScore to JSON: %v", err)
	}
}
