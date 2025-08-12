package webanalysis

import (
	"testing"
)

func TestPageTypeDetector(t *testing.T) {
	config := PageTypeConfig{
		DetectionWeights: map[string]float64{
			"url":       0.3,
			"content":   0.4,
			"structure": 0.3,
		},
		ConfidenceThresholds: map[string]float64{
			"min_confidence": 0.5,
			"max_confidence": 1.0,
		},
	}

	detector := NewPageTypeDetector(config)

	// Test about us page
	aboutUsContent := &ScrapedContent{
		URL:   "https://example.com/about-us",
		Title: "About Our Company",
		Text:  "We are a company founded in 2010. Our story begins with a mission to provide quality services. Learn more about our company history and who we are.",
		HTML:  "<html><body><h1>About Us</h1><p>We are a company founded in 2010.</p><h2>Our Story</h2><p>Our story begins with a mission.</p></body></html>",
	}

	detection := detector.DetectPageType(aboutUsContent)

	// Test detection results
	if detection.Type != "about_us" {
		t.Errorf("Expected page type 'about_us', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.5 {
		t.Errorf("Expected confidence >= 0.5, got %f", detection.Confidence)
	}

	if detection.DetectionMethod == "" {
		t.Error("Expected detection method to be set")
	}

	if len(detection.Keywords) == 0 {
		t.Error("Expected keywords to be populated")
	}

	// Test services page
	servicesContent := &ScrapedContent{
		URL:   "https://example.com/services",
		Title: "Our Services",
		Text:  "We provide comprehensive consulting services and professional solutions. Our service offerings include strategic planning and implementation.",
		HTML:  "<html><body><h1>Our Services</h1><ul><li>Consulting Services</li><li>Strategic Planning</li></ul></body></html>",
	}

	detection = detector.DetectPageType(servicesContent)

	if detection.Type != "services" {
		t.Errorf("Expected page type 'services', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.5 {
		t.Errorf("Expected confidence >= 0.5, got %f", detection.Confidence)
	}
}

func TestURLPatternAnalyzer(t *testing.T) {
	analyzer := NewURLPatternAnalyzer()

	// Test about us URL patterns
	testCases := []struct {
		url          string
		expectedType string
		description  string
	}{
		{"https://example.com/about", "about_us", "Simple about URL"},
		{"https://example.com/about-us", "about_us", "Hyphenated about URL"},
		{"https://example.com/company", "about_us", "Company URL"},
		{"https://example.com/who-we-are", "about_us", "Who we are URL"},
		{"https://example.com/mission", "mission", "Mission URL"},
		{"https://example.com/vision", "mission", "Vision URL"},
		{"https://example.com/products", "products", "Products URL"},
		{"https://example.com/services", "services", "Services URL"},
		{"https://example.com/contact", "contact", "Contact URL"},
		{"https://example.com/team", "team", "Team URL"},
		{"https://example.com/careers", "careers", "Careers URL"},
		{"https://example.com/news", "news", "News URL"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			detection := analyzer.AnalyzeURLPattern(tc.url)
			if detection.Type != tc.expectedType {
				t.Errorf("Expected type '%s', got '%s' for URL '%s'", tc.expectedType, detection.Type, tc.url)
			}
			if detection.Confidence < 0.5 {
				t.Errorf("Expected confidence >= 0.5, got %f for URL '%s'", detection.Confidence, tc.url)
			}
		})
	}

	// Test unknown URL
	unknownDetection := analyzer.AnalyzeURLPattern("https://example.com/random-page")
	if unknownDetection.Type != "unknown" {
		t.Errorf("Expected type 'unknown', got '%s'", unknownDetection.Type)
	}
	if unknownDetection.Confidence != 0.0 {
		t.Errorf("Expected confidence 0.0, got %f", unknownDetection.Confidence)
	}
}

func TestContentTypeAnalyzer(t *testing.T) {
	analyzer := NewContentTypeAnalyzer()

	// Test about us content
	aboutUsText := "We are a company founded in 2010. Our story begins with a mission to provide quality services. Learn more about our company history and who we are."
	aboutUsTitle := "About Our Company"

	detection := analyzer.AnalyzeContent(aboutUsText, aboutUsTitle)

	if detection.Type != "about_us" {
		t.Errorf("Expected type 'about_us', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.3 {
		t.Errorf("Expected confidence >= 0.3, got %f", detection.Confidence)
	}

	if len(detection.Keywords) == 0 {
		t.Error("Expected keywords to be populated")
	}

	// Test services content
	servicesText := "We provide comprehensive consulting services and professional solutions. Our service offerings include strategic planning and implementation."
	servicesTitle := "Our Services"

	detection = analyzer.AnalyzeContent(servicesText, servicesTitle)

	if detection.Type != "services" {
		t.Errorf("Expected type 'services', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.3 {
		t.Errorf("Expected confidence >= 0.3, got %f", detection.Confidence)
	}

	// Test mission content
	missionText := "Our mission is to deliver exceptional value. Our vision drives everything we do. We are guided by our core values and principles."
	missionTitle := "Mission and Vision"

	detection = analyzer.AnalyzeContent(missionText, missionTitle)

	if detection.Type != "mission" {
		t.Errorf("Expected type 'mission', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.3 {
		t.Errorf("Expected confidence >= 0.3, got %f", detection.Confidence)
	}
}

func TestPageStructureAnalyzer(t *testing.T) {
	analyzer := NewPageStructureAnalyzer()

	// Test about us structure
	aboutUsHTML := `<html><body>
		<h1>About Us</h1>
		<p>We are a company founded in 2010.</p>
		<h2>Our Story</h2>
		<p>Our story begins with a mission.</p>
		<h3>Company History</h3>
		<p>Learn more about our company history.</p>
	</body></html>`

	detection := analyzer.AnalyzeStructure(aboutUsHTML)

	if detection.Type != "about_us" {
		t.Errorf("Expected type 'about_us', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.3 {
		t.Errorf("Expected confidence >= 0.3, got %f", detection.Confidence)
	}

	// Test services structure
	servicesHTML := `<html><body>
		<h1>Our Services</h1>
		<p>We provide comprehensive services.</p>
		<h2>Service List</h2>
		<ul>
			<li>Consulting Services</li>
			<li>Strategic Planning</li>
			<li>Implementation</li>
		</ul>
		<h3>What We Do</h3>
		<p>We specialize in business solutions.</p>
	</body></html>`

	detection = analyzer.AnalyzeStructure(servicesHTML)

	if detection.Type != "services" {
		t.Errorf("Expected type 'services', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.3 {
		t.Errorf("Expected confidence >= 0.3, got %f", detection.Confidence)
	}

	// Test products structure
	productsHTML := `<html><body>
		<h1>Our Products</h1>
		<p>We offer a wide range of products.</p>
		<h2>Product Catalog</h2>
		<ul>
			<li>Product A</li>
			<li>Product B</li>
			<li>Product C</li>
		</ul>
	</body></html>`

	detection = analyzer.AnalyzeStructure(productsHTML)

	if detection.Type != "products" {
		t.Errorf("Expected type 'products', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.3 {
		t.Errorf("Expected confidence >= 0.3, got %f", detection.Confidence)
	}
}

func TestPageTypePriority(t *testing.T) {
	config := PageTypeConfig{}
	detector := NewPageTypeDetector(config)

	// Test priority scores
	testCases := []struct {
		pageType         string
		expectedPriority float64
		description      string
	}{
		{"about_us", 0.9, "About us should have high priority"},
		{"mission", 0.8, "Mission should have high priority"},
		{"services", 0.85, "Services should have high priority"},
		{"products", 0.8, "Products should have high priority"},
		{"contact", 0.7, "Contact should have medium-high priority"},
		{"team", 0.6, "Team should have medium priority"},
		{"careers", 0.4, "Careers should have lower priority"},
		{"news", 0.3, "News should have lower priority"},
		{"unknown", 0.1, "Unknown should have lowest priority"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			priority := detector.GetPageTypePriority(tc.pageType)
			if priority != tc.expectedPriority {
				t.Errorf("Expected priority %f, got %f for page type '%s'", tc.expectedPriority, priority, tc.pageType)
			}
		})
	}
}

func TestHighPriorityPageTypes(t *testing.T) {
	config := PageTypeConfig{}
	detector := NewPageTypeDetector(config)

	// Test high priority page types
	highPriorityTypes := []string{"about_us", "mission", "services", "products"}
	lowPriorityTypes := []string{"contact", "team", "careers", "news", "unknown"}

	for _, pageType := range highPriorityTypes {
		if !detector.IsHighPriorityPageType(pageType) {
			t.Errorf("Expected '%s' to be high priority", pageType)
		}
	}

	for _, pageType := range lowPriorityTypes {
		if detector.IsHighPriorityPageType(pageType) {
			t.Errorf("Expected '%s' to be low priority", pageType)
		}
	}
}

func TestPageTypeDescription(t *testing.T) {
	config := PageTypeConfig{}
	detector := NewPageTypeDetector(config)

	// Test descriptions
	testCases := []struct {
		pageType            string
		expectedDescription string
	}{
		{"about_us", "About Us/Company Information Page"},
		{"mission", "Mission/Vision/Values Page"},
		{"services", "Services/Solutions Page"},
		{"products", "Products/Catalog Page"},
		{"contact", "Contact Information Page"},
		{"team", "Team/Leadership Page"},
		{"careers", "Careers/Employment Page"},
		{"news", "News/Blog/Media Page"},
		{"unknown", "Unknown Page Type"},
	}

	for _, tc := range testCases {
		description := detector.GetPageTypeDescription(tc.pageType)
		if description != tc.expectedDescription {
			t.Errorf("Expected description '%s', got '%s' for page type '%s'", tc.expectedDescription, description, tc.pageType)
		}
	}
}

func TestPageTypeDetectionEdgeCases(t *testing.T) {
	config := PageTypeConfig{
		DetectionWeights: map[string]float64{
			"url":       0.3,
			"content":   0.4,
			"structure": 0.3,
		},
	}

	detector := NewPageTypeDetector(config)

	// Test empty content
	emptyContent := &ScrapedContent{
		URL:   "https://example.com",
		Title: "",
		Text:  "",
		HTML:  "",
	}

	detection := detector.DetectPageType(emptyContent)

	// Should handle empty content gracefully
	if detection.Type != "unknown" {
		t.Errorf("Expected page type 'unknown' for empty content, got '%s'", detection.Type)
	}

	// Test minimal content
	minimalContent := &ScrapedContent{
		URL:   "https://example.com",
		Title: "Test",
		Text:  "Minimal content.",
		HTML:  "<html><body><p>Minimal</p></body></html>",
	}

	detection = detector.DetectPageType(minimalContent)

	// Should provide reasonable detection for minimal content
	if detection.Type == "" {
		t.Error("Expected page type to be set for minimal content")
	}

	if detection.Confidence < 0 {
		t.Errorf("Expected confidence >= 0, got %f", detection.Confidence)
	}
}

func TestDetectionMethodCombination(t *testing.T) {
	config := PageTypeConfig{
		DetectionWeights: map[string]float64{
			"url":       0.3,
			"content":   0.4,
			"structure": 0.3,
		},
	}

	detector := NewPageTypeDetector(config)

	// Test content that matches multiple detection methods
	multiMethodContent := &ScrapedContent{
		URL:   "https://example.com/about-us",
		Title: "About Our Company",
		Text:  "We are a company founded in 2010. Our story begins with a mission to provide quality services. Learn more about our company history and who we are.",
		HTML:  "<html><body><h1>About Us</h1><p>We are a company founded in 2010.</p><h2>Our Story</h2><p>Our story begins with a mission.</p></body></html>",
	}

	detection := detector.DetectPageType(multiMethodContent)

	// Should detect about_us with high confidence due to multiple methods
	if detection.Type != "about_us" {
		t.Errorf("Expected type 'about_us', got '%s'", detection.Type)
	}

	if detection.Confidence < 0.6 {
		t.Errorf("Expected high confidence >= 0.6 due to multiple detection methods, got %f", detection.Confidence)
	}

	// Should have multiple detection indicators
	if len(detection.Keywords) == 0 {
		t.Error("Expected keywords to be populated")
	}

	if len(detection.URLPatterns) == 0 {
		t.Error("Expected URL patterns to be populated")
	}

	if len(detection.ContentIndicators) == 0 {
		t.Error("Expected content indicators to be populated")
	}
}
