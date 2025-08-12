package webanalysis

import (
	"testing"
	"time"
)

func TestPageContentQualityAssessor(t *testing.T) {
	config := ContentQualityConfig{
		Weights: map[string]float64{
			"readability":       0.2,
			"structure":         0.2,
			"completeness":      0.2,
			"business_content":  0.2,
			"technical_content": 0.2,
		},
		QualityThresholds: map[string]float64{
			"min_quality": 0.5,
			"max_quality": 1.0,
		},
	}

	assessor := NewPageContentQualityAssessor(config)

	// Test content with good quality
	goodContent := &ScrapedContent{
		URL:          "https://example.com",
		Title:        "Professional Business Services",
		Text:         "We are a professional business established in 2010, specializing in quality services. Our team of experts provides comprehensive solutions to meet your needs. Contact us at info@example.com or call (555) 123-4567. We are certified and licensed professionals with years of experience in the industry.",
		HTML:         "<html><head><title>Professional Business Services</title><meta name=\"description\" content=\"Quality business services\"></head><body><h1>About Us</h1><p>We are a professional business established in 2010, specializing in quality services.</p><h2>Our Services</h2><ul><li>Consulting</li><li>Training</li><li>Support</li></ul><h2>Contact</h2><p>Contact us at info@example.com or call (555) 123-4567.</p></body></html>",
		ResponseTime: 200 * time.Millisecond,
	}

	quality := assessor.AssessContentQuality(goodContent, "Professional Business Services")

	// Test overall quality
	if quality.OverallQuality < 0.4 {
		t.Errorf("Expected overall quality >= 0.4, got %f", quality.OverallQuality)
	}

	// Test component scores
	if quality.ComponentScores["readability"] < 0.3 {
		t.Errorf("Expected readability score >= 0.3, got %f", quality.ComponentScores["readability"])
	}

	if quality.ComponentScores["structure"] < 0.2 {
		t.Errorf("Expected structure score >= 0.2, got %f", quality.ComponentScores["structure"])
	}

	if quality.ComponentScores["completeness"] < 0.3 {
		t.Errorf("Expected completeness score >= 0.3, got %f", quality.ComponentScores["completeness"])
	}

	if quality.ComponentScores["business_content"] < 0.3 {
		t.Errorf("Expected business content score >= 0.3, got %f", quality.ComponentScores["business_content"])
	}

	// Test quality factors
	if len(quality.QualityFactors) == 0 {
		t.Error("Expected quality factors to be populated")
	}

	// Test metrics
	if quality.ReadabilityMetrics.ReadabilityScore < 0.3 {
		t.Errorf("Expected readability metrics score >= 0.3, got %f", quality.ReadabilityMetrics.ReadabilityScore)
	}

	if quality.StructureMetrics.StructureScore < 0.2 {
		t.Errorf("Expected structure metrics score >= 0.2, got %f", quality.StructureMetrics.StructureScore)
	}

	if quality.CompletenessMetrics.CompletenessScore < 0.3 {
		t.Errorf("Expected completeness metrics score >= 0.3, got %f", quality.CompletenessMetrics.CompletenessScore)
	}

	if quality.BusinessMetrics.BusinessScore < 0.3 {
		t.Errorf("Expected business metrics score >= 0.3, got %f", quality.BusinessMetrics.BusinessScore)
	}

	if quality.TechnicalMetrics.TechnicalScore < 0.2 {
		t.Errorf("Expected technical metrics score >= 0.2, got %f", quality.TechnicalMetrics.TechnicalScore)
	}
}

func TestStructureAnalyzer(t *testing.T) {
	analyzer := NewStructureAnalyzer()

	// Test good structure
	goodHTML := `<html><head><title>Test</title></head><body>
		<h1>Main Title</h1>
		<p>Introduction paragraph with good content.</p>
		<h2>Section 1</h2>
		<p>First section content.</p>
		<ul><li>Item 1</li><li>Item 2</li></ul>
		<h2>Section 2</h2>
		<p>Second section content.</p>
		<nav><a href="/about">About</a><a href="/contact">Contact</a></nav>
		</body></html>`

	goodText := "This is a well-structured document with clear sections. First, we introduce the topic. Then, we provide detailed information. Finally, we conclude with a summary."

	metrics := analyzer.AnalyzeStructure(goodHTML, goodText)

	// Test heading structure
	if metrics.HeadingStructure < 0.6 {
		t.Errorf("Expected heading structure score >= 0.6, got %f", metrics.HeadingStructure)
	}

	// Test paragraph structure
	if metrics.ParagraphStructure < 0.6 {
		t.Errorf("Expected paragraph structure score >= 0.6, got %f", metrics.ParagraphStructure)
	}

	// Test list structure
	if metrics.ListStructure < 0.6 {
		t.Errorf("Expected list structure score >= 0.6, got %f", metrics.ListStructure)
	}

	// Test navigation structure
	if metrics.NavigationStructure < 0.6 {
		t.Errorf("Expected navigation structure score >= 0.6, got %f", metrics.NavigationStructure)
	}

	// Test overall structure score
	if metrics.StructureScore < 0.5 {
		t.Errorf("Expected overall structure score >= 0.5, got %f", metrics.StructureScore)
	}
}

func TestCompletenessAnalyzer(t *testing.T) {
	analyzer := NewCompletenessAnalyzer()

	// Test complete content
	completeContent := &ScrapedContent{
		URL:   "https://example.com",
		Title: "Complete Business Information",
		Text:  "Our company was established in 2010 and has been providing professional services for over 10 years. We specialize in consulting and training services. Our team of certified experts includes 25 professionals with extensive experience. Contact us at info@example.com or call (555) 123-4567. Our address is 123 Business Street, City, State 12345. We offer comprehensive solutions including project management, training programs, and technical support. Our mission is to deliver exceptional value to our clients through innovative approaches and proven methodologies.",
		HTML:  "<html><body><p>Complete content</p></body></html>",
	}

	metrics := analyzer.AnalyzeCompleteness(completeContent)

	// Test content length
	if metrics.ContentLength < 200 {
		t.Errorf("Expected content length >= 200, got %d", metrics.ContentLength)
	}

	// Test information density
	if metrics.InformationDensity < 0.3 {
		t.Errorf("Expected information density >= 0.3, got %f", metrics.InformationDensity)
	}

	// Test factual content
	if metrics.FactualContent < 0.5 {
		t.Errorf("Expected factual content score >= 0.5, got %f", metrics.FactualContent)
	}

	// Test contact information
	if metrics.ContactInformation < 0.5 {
		t.Errorf("Expected contact information score >= 0.5, got %f", metrics.ContactInformation)
	}

	// Test business information
	if metrics.BusinessInformation < 0.5 {
		t.Errorf("Expected business information score >= 0.5, got %f", metrics.BusinessInformation)
	}

	// Test service information
	if metrics.ServiceInformation < 0.5 {
		t.Errorf("Expected service information score >= 0.5, got %f", metrics.ServiceInformation)
	}

	// Test overall completeness score
	if metrics.CompletenessScore < 0.5 {
		t.Errorf("Expected overall completeness score >= 0.5, got %f", metrics.CompletenessScore)
	}
}

func TestBusinessContentAnalyzer(t *testing.T) {
	analyzer := NewBusinessContentAnalyzer()

	// Test business content
	businessContent := &ScrapedContent{
		URL:   "https://example.com",
		Title: "ABC Consulting Services",
		Text:  "ABC Consulting Services was founded in 2010 and has been providing professional consulting services for over 10 years. Our team of certified experts specializes in business transformation and digital innovation. We offer comprehensive solutions including strategic planning, process optimization, and technology implementation. Our CEO, John Smith, leads our team of 25 professionals. We are ISO 9001 certified and have received numerous industry awards. Our clients include Fortune 500 companies and we have a 95% satisfaction rate. Contact us to learn more about our services and case studies.",
		HTML:  "<html><body><p>Business content</p></body></html>",
	}

	metrics := analyzer.AnalyzeBusinessContent(businessContent, "ABC Consulting Services")

	// Test business name presence
	if metrics.BusinessNamePresence < 0.5 {
		t.Errorf("Expected business name presence score >= 0.5, got %f", metrics.BusinessNamePresence)
	}

	// Test service description
	if metrics.ServiceDescription < 0.5 {
		t.Errorf("Expected service description score >= 0.5, got %f", metrics.ServiceDescription)
	}

	// Test company history
	if metrics.CompanyHistory < 0.5 {
		t.Errorf("Expected company history score >= 0.5, got %f", metrics.CompanyHistory)
	}

	// Test team information
	if metrics.TeamInformation < 0.5 {
		t.Errorf("Expected team information score >= 0.5, got %f", metrics.TeamInformation)
	}

	// Test certifications
	if metrics.Certifications < 0.5 {
		t.Errorf("Expected certifications score >= 0.5, got %f", metrics.Certifications)
	}

	// Test testimonials
	if metrics.Testimonials < 0.3 {
		t.Errorf("Expected testimonials score >= 0.3, got %f", metrics.Testimonials)
	}

	// Test case studies
	if metrics.CaseStudies < 0.3 {
		t.Errorf("Expected case studies score >= 0.3, got %f", metrics.CaseStudies)
	}

	// Test overall business score
	if metrics.BusinessScore < 0.5 {
		t.Errorf("Expected overall business score >= 0.5, got %f", metrics.BusinessScore)
	}
}

func TestTechnicalContentAnalyzer(t *testing.T) {
	analyzer := NewTechnicalContentAnalyzer()

	// Test technical content
	technicalContent := &ScrapedContent{
		URL:   "https://example.com",
		Title: "Technical Website",
		HTML:  `<html><head><title>Technical Website</title><meta name="description" content="Technical content"><meta name="keywords" content="technical, web, development"><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head><body><nav><a href="/about" aria-label="About Us">About</a></nav><main><h1>Main Content</h1><p>This is well-structured content with <img src="image.jpg" alt="Description" width="300" height="200"> proper images and <a href="/contact">contact links</a>.</p></main></body></html>`,
	}

	metrics := analyzer.AnalyzeTechnicalContent(technicalContent)

	// Test HTML validity
	if metrics.HTMLValidity < 0.5 {
		t.Errorf("Expected HTML validity score >= 0.5, got %f", metrics.HTMLValidity)
	}

	// Test accessibility score
	if metrics.AccessibilityScore < 0.5 {
		t.Errorf("Expected accessibility score >= 0.5, got %f", metrics.AccessibilityScore)
	}

	// Test mobile optimization
	if metrics.MobileOptimization < 0.5 {
		t.Errorf("Expected mobile optimization score >= 0.5, got %f", metrics.MobileOptimization)
	}

	// Test image optimization
	if metrics.ImageOptimization < 0.5 {
		t.Errorf("Expected image optimization score >= 0.5, got %f", metrics.ImageOptimization)
	}

	// Test link quality
	if metrics.LinkQuality < 0.5 {
		t.Errorf("Expected link quality score >= 0.5, got %f", metrics.LinkQuality)
	}

	// Test meta tag completeness
	if metrics.MetaTagCompleteness < 0.5 {
		t.Errorf("Expected meta tag completeness score >= 0.5, got %f", metrics.MetaTagCompleteness)
	}

	// Test overall technical score
	if metrics.TechnicalScore < 0.5 {
		t.Errorf("Expected overall technical score >= 0.5, got %f", metrics.TechnicalScore)
	}
}

func TestPageContentQualityEdgeCases(t *testing.T) {
	config := ContentQualityConfig{
		Weights: map[string]float64{
			"readability":       0.2,
			"structure":         0.2,
			"completeness":      0.2,
			"business_content":  0.2,
			"technical_content": 0.2,
		},
	}

	assessor := NewPageContentQualityAssessor(config)

	// Test empty content
	emptyContent := &ScrapedContent{
		URL:   "https://example.com",
		Title: "",
		Text:  "",
		HTML:  "",
	}

	quality := assessor.AssessContentQuality(emptyContent, "Test Business")

	// Should handle empty content gracefully
	if quality.OverallQuality < 0 {
		t.Errorf("Expected overall quality >= 0 for empty content, got %f", quality.OverallQuality)
	}

	// Test minimal content
	minimalContent := &ScrapedContent{
		URL:   "https://example.com",
		Title: "Test",
		Text:  "Minimal content.",
		HTML:  "<html><body><p>Minimal</p></body></html>",
	}

	quality = assessor.AssessContentQuality(minimalContent, "Test Business")

	// Should provide reasonable scores for minimal content
	if quality.OverallQuality < 0 {
		t.Errorf("Expected overall quality >= 0 for minimal content, got %f", quality.OverallQuality)
	}
}

func TestQualityFactorDetails(t *testing.T) {
	config := ContentQualityConfig{
		Weights: map[string]float64{
			"readability":       0.2,
			"structure":         0.2,
			"completeness":      0.2,
			"business_content":  0.2,
			"technical_content": 0.2,
		},
	}

	assessor := NewPageContentQualityAssessor(config)

	content := &ScrapedContent{
		URL:   "https://example.com",
		Title: "Quality Test",
		Text:  "This is a test of the quality assessment system with good content structure and business information.",
		HTML:  "<html><head><title>Quality Test</title></head><body><h1>Test</h1><p>Content</p></body></html>",
	}

	quality := assessor.AssessContentQuality(content, "Test Business")

	// Test that quality factors have proper details
	for _, factor := range quality.QualityFactors {
		if factor.Factor == "" {
			t.Error("Quality factor should have a factor name")
		}
		if factor.Reason == "" {
			t.Error("Quality factor should have a reason")
		}
		if factor.Details == "" {
			t.Error("Quality factor should have details")
		}
		if factor.Confidence <= 0 || factor.Confidence > 1 {
			t.Errorf("Quality factor confidence should be between 0 and 1, got %f", factor.Confidence)
		}
	}
}
