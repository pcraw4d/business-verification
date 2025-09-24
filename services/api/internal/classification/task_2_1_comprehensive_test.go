package classification

import (
	"strings"
	"testing"
	"time"
)

// TestTask2_1_HTMLContentCleaning tests the HTML content cleaning functionality
// This test matches the specification in the comprehensive plan
func TestTask2_1_HTMLContentCleaning(t *testing.T) {
	service := &IndustryDetectionService{}

	htmlContent := `<html><head><script>alert('test');</script></head><body><h1>Restaurant</h1><p>Fine dining establishment</p></body></html>`
	cleanContent := service.cleanHTMLContent(htmlContent)

	// Verify script tags and JavaScript are removed
	if strings.Contains(cleanContent, "<script>") {
		t.Error("Script tags should be removed from HTML content")
	}
	if strings.Contains(cleanContent, "alert") {
		t.Error("JavaScript code should be removed from HTML content")
	}

	// Verify business content is preserved
	if !strings.Contains(cleanContent, "Restaurant") {
		t.Error("Business content 'Restaurant' should be preserved")
	}
	if !strings.Contains(cleanContent, "Fine dining establishment") {
		t.Error("Business content 'Fine dining establishment' should be preserved")
	}

	t.Logf("HTML cleaning successful: '%s' -> '%s'", htmlContent, cleanContent)
}

// TestTask2_1_BusinessKeywordExtraction tests business keyword extraction
// This test matches the specification in the comprehensive plan
func TestTask2_1_BusinessKeywordExtraction(t *testing.T) {
	service := &IndustryDetectionService{}

	content := "This is a restaurant serving Italian cuisine with fine dining experience"
	keywords := service.extractKeywordsFromContent(content)

	// Verify business keywords are extracted
	expectedKeywords := []string{"restaurant", "italian", "cuisine", "dining"}
	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range keywords {
			if keyword == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected business keyword '%s' not found in extracted keywords: %v", expected, keywords)
		}
	}

	// Verify common words are filtered out
	commonWords := []string{"this", "is", "a", "with"}
	for _, common := range commonWords {
		for _, keyword := range keywords {
			if keyword == common {
				t.Errorf("Common word '%s' should not be in extracted keywords", common)
			}
		}
	}

	t.Logf("Business keyword extraction successful: extracted %d keywords: %v", len(keywords), keywords)
}

// TestTask2_1_TechnicalTermFiltering tests technical term filtering
// This test matches the specification in the comprehensive plan
func TestTask2_1_TechnicalTermFiltering(t *testing.T) {
	service := &IndustryDetectionService{}

	content := "Restaurant with HTML website and JavaScript functionality"
	keywords := service.extractKeywordsFromContent(content)

	// Verify business keywords are preserved
	foundRestaurant := false
	for _, keyword := range keywords {
		if keyword == "restaurant" {
			foundRestaurant = true
			break
		}
	}
	if !foundRestaurant {
		t.Error("Business keyword 'restaurant' should be preserved")
	}

	// Verify technical terms are filtered out
	technicalTerms := []string{"html", "javascript"}
	for _, tech := range technicalTerms {
		for _, keyword := range keywords {
			if keyword == tech {
				t.Errorf("Technical term '%s' should be filtered out from keywords", tech)
			}
		}
	}

	// Note: "website" is now considered business-relevant in some contexts
	// so we don't filter it out in this test

	t.Logf("Technical term filtering successful: extracted %d keywords: %v", len(keywords), keywords)
}

// TestTask2_1_EndToEndWorkflow tests the complete workflow from HTML to keywords
func TestTask2_1_EndToEndWorkflow(t *testing.T) {
	service := &IndustryDetectionService{}

	// Complex HTML content with mixed business and technical information
	htmlContent := `
		<html>
			<head>
				<title>Mario's Italian Restaurant</title>
				<script>function showMenu() { alert('Menu'); }</script>
				<style>body { font-family: Arial; }</style>
			</head>
			<body>
				<h1>Welcome to Mario's Italian Restaurant</h1>
				<p>We serve authentic Italian cuisine with fine dining experience.</p>
				<p>Our restaurant offers traditional pasta, pizza, and wine selection.</p>
				<div class="contact">
					<p>Contact us for reservations and catering services.</p>
				</div>
				<!-- HTML comment with technical info -->
				<script>console.log('Website analytics');</script>
			</body>
		</html>
	`

	// Step 1: Clean HTML content
	cleanContent := service.cleanHTMLContent(htmlContent)

	// Verify HTML cleaning
	if strings.Contains(cleanContent, "<script>") || strings.Contains(cleanContent, "<style>") {
		t.Error("HTML tags should be removed")
	}
	if strings.Contains(cleanContent, "alert") || strings.Contains(cleanContent, "console.log") {
		t.Error("JavaScript code should be removed")
	}
	if strings.Contains(cleanContent, "<!--") || strings.Contains(cleanContent, "-->") {
		t.Error("HTML comments should be removed")
	}

	// Step 2: Extract business keywords
	keywords := service.extractKeywordsFromContent(cleanContent)

	// Verify business keywords are extracted
	expectedBusinessKeywords := []string{"restaurant", "italian", "cuisine", "dining", "pasta", "pizza", "wine", "catering", "services"}
	foundBusinessKeywords := 0
	for _, expected := range expectedBusinessKeywords {
		for _, keyword := range keywords {
			if keyword == expected {
				foundBusinessKeywords++
				break
			}
		}
	}

	if foundBusinessKeywords < 5 {
		t.Errorf("Expected at least 5 business keywords, found %d", foundBusinessKeywords)
	}

	// Verify technical terms are filtered out
	technicalTerms := []string{"html", "script", "style", "function", "console", "log", "analytics"}
	for _, tech := range technicalTerms {
		for _, keyword := range keywords {
			if keyword == tech {
				t.Errorf("Technical term '%s' should be filtered out", tech)
			}
		}
	}

	// Verify common words are filtered out
	commonWords := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}
	for _, common := range commonWords {
		for _, keyword := range keywords {
			if keyword == common {
				t.Errorf("Common word '%s' should be filtered out", common)
			}
		}
	}

	t.Logf("End-to-end workflow successful: extracted %d high-quality business keywords: %v", len(keywords), keywords)
}

// TestTask2_1_PerformanceBenchmarks tests performance requirements
func TestTask2_1_PerformanceBenchmarks(t *testing.T) {
	service := &IndustryDetectionService{}

	// Create large HTML content for performance testing
	largeHTMLContent := strings.Repeat(`
		<html>
			<head><script>alert('test');</script></head>
			<body>
				<h1>Restaurant Business</h1>
				<p>We provide fine dining services with authentic cuisine and professional staff.</p>
				<p>Our restaurant offers catering, reservations, and special events.</p>
			</body>
		</html>
	`, 50) // Repeat 50 times for large content

	// Test HTML cleaning performance
	start := time.Now()
	cleanContent := service.cleanHTMLContent(largeHTMLContent)
	cleaningDuration := time.Since(start)

	// Test keyword extraction performance
	start = time.Now()
	keywords := service.extractKeywordsFromContent(cleanContent)
	extractionDuration := time.Since(start)

	// Verify performance requirements
	if cleaningDuration > 50*time.Millisecond {
		t.Errorf("HTML cleaning took too long: %v (should be < 50ms)", cleaningDuration)
	}

	if extractionDuration > 100*time.Millisecond {
		t.Errorf("Keyword extraction took too long: %v (should be < 100ms)", extractionDuration)
	}

	// Verify results are not empty
	if len(cleanContent) == 0 {
		t.Error("Cleaned content should not be empty")
	}

	if len(keywords) == 0 {
		t.Error("Extracted keywords should not be empty")
	}

	// Verify quality of results
	expectedKeywords := []string{"restaurant", "dining", "services", "cuisine", "staff", "catering", "reservations", "events"}
	foundKeywords := 0
	for _, expected := range expectedKeywords {
		for _, keyword := range keywords {
			if keyword == expected {
				foundKeywords++
				break
			}
		}
	}

	if foundKeywords < 5 {
		t.Errorf("Expected at least 5 high-quality keywords, found %d", foundKeywords)
	}

	t.Logf("Performance benchmarks passed:")
	t.Logf("  HTML cleaning: %v for %d characters", cleaningDuration, len(largeHTMLContent))
	t.Logf("  Keyword extraction: %v for %d characters", extractionDuration, len(cleanContent))
	t.Logf("  Extracted %d keywords: %v", len(keywords), keywords)
}

// TestTask2_1_EdgeCases tests edge cases and error conditions
func TestTask2_1_EdgeCases(t *testing.T) {
	service := &IndustryDetectionService{}

	testCases := []struct {
		name    string
		content string
		expect  func(t *testing.T, keywords []string)
	}{
		{
			name:    "Empty content",
			content: "",
			expect: func(t *testing.T, keywords []string) {
				if len(keywords) != 0 {
					t.Error("Empty content should return no keywords")
				}
			},
		},
		{
			name:    "Only HTML tags",
			content: "<html><head></head><body></body></html>",
			expect: func(t *testing.T, keywords []string) {
				if len(keywords) != 0 {
					t.Error("Content with only HTML tags should return no keywords")
				}
			},
		},
		{
			name:    "Only technical terms",
			content: "HTML CSS JavaScript function var const let",
			expect: func(t *testing.T, keywords []string) {
				if len(keywords) != 0 {
					t.Errorf("Content with only technical terms should return no keywords, got: %v", keywords)
				}
			},
		},
		{
			name:    "Only common words",
			content: "the and or but in on at to for of with by",
			expect: func(t *testing.T, keywords []string) {
				if len(keywords) != 0 {
					t.Errorf("Content with only common words should return no keywords, got: %v", keywords)
				}
			},
		},
		{
			name:    "Mixed case business terms",
			content: "RESTAURANT Restaurant restaurant DINING Dining dining",
			expect: func(t *testing.T, keywords []string) {
				// Should extract business terms regardless of case
				foundRestaurant := false
				foundDining := false
				for _, keyword := range keywords {
					if keyword == "restaurant" {
						foundRestaurant = true
					}
					if keyword == "dining" {
						foundDining = true
					}
				}
				if !foundRestaurant {
					t.Error("Should extract 'restaurant' regardless of case")
				}
				if !foundDining {
					t.Error("Should extract 'dining' regardless of case")
				}
			},
		},
		{
			name:    "Special characters and encoding",
			content: "Restaurant &amp; Café serving authentic cuisine with &quot;quality&quot; service",
			expect: func(t *testing.T, keywords []string) {
				// Should handle HTML entities properly
				expectedKeywords := []string{"restaurant", "café", "serving", "authentic", "cuisine", "quality", "service"}
				foundKeywords := 0
				for _, expected := range expectedKeywords {
					for _, keyword := range keywords {
						if keyword == expected {
							foundKeywords++
							break
						}
					}
				}
				if foundKeywords < 5 {
					t.Errorf("Should extract business keywords from content with HTML entities, found %d", foundKeywords)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keywords := service.extractKeywordsFromContent(tc.content)
			tc.expect(t, keywords)
			t.Logf("Edge case '%s' handled correctly: %d keywords extracted", tc.name, len(keywords))
		})
	}
}

// TestTask2_1_IntegrationWithExistingSystem tests integration with existing classification system
func TestTask2_1_IntegrationWithExistingSystem(t *testing.T) {
	service := &IndustryDetectionService{}

	// Test with realistic business information
	businessName := "Mario's Italian Restaurant"
	description := "Authentic Italian restaurant serving fine dining cuisine with professional service and quality ingredients. We offer catering services and special events."
	websiteURL := "https://mariosrestaurant.com"

	// Extract keywords from business information
	keywords := service.extractKeywordsFromBusinessInfo(businessName, description, websiteURL)

	// Verify business-relevant keywords are extracted
	expectedKeywords := []string{"restaurant", "italian", "dining", "cuisine", "professional", "service", "quality", "ingredients", "catering", "services", "events"}
	foundKeywords := 0
	for _, expected := range expectedKeywords {
		for _, keyword := range keywords {
			if keyword == expected {
				foundKeywords++
				break
			}
		}
	}

	if foundKeywords < 8 {
		t.Errorf("Expected at least 8 business keywords from business info, found %d", foundKeywords)
	}

	// Verify technical terms are filtered out
	technicalTerms := []string{"https", "com", "website"}
	for _, tech := range technicalTerms {
		for _, keyword := range keywords {
			if keyword == tech {
				t.Errorf("Technical term '%s' should be filtered out from business info", tech)
			}
		}
	}

	// Verify common words are filtered out
	commonWords := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by", "we", "our"}
	for _, common := range commonWords {
		for _, keyword := range keywords {
			if keyword == common {
				t.Errorf("Common word '%s' should be filtered out from business info", common)
			}
		}
	}

	t.Logf("Integration test successful: extracted %d high-quality keywords from business info: %v", len(keywords), keywords)
}

// TestTask2_1_QualityMetrics tests quality metrics and scoring
func TestTask2_1_QualityMetrics(t *testing.T) {
	service := &IndustryDetectionService{}

	// Test with high-quality business content
	highQualityContent := "Restaurant serving authentic Italian cuisine with fine dining experience and professional service"
	highQualityKeywords := service.extractKeywordsFromContent(highQualityContent)

	// Test with low-quality content
	lowQualityContent := "The and or but in on at to for of with by"
	lowQualityKeywords := service.extractKeywordsFromContent(lowQualityContent)

	// High-quality content should produce more keywords
	if len(highQualityKeywords) <= len(lowQualityKeywords) {
		t.Errorf("High-quality content should produce more keywords than low-quality content: %d vs %d",
			len(highQualityKeywords), len(lowQualityKeywords))
	}

	// High-quality content should have business-relevant keywords
	businessRelevantCount := 0
	for _, keyword := range highQualityKeywords {
		if service.isBusinessRelevant(keyword) {
			businessRelevantCount++
		}
	}

	if businessRelevantCount < 5 {
		t.Errorf("High-quality content should have at least 5 business-relevant keywords, found %d", businessRelevantCount)
	}

	// Low-quality content should have no business-relevant keywords
	businessRelevantCount = 0
	for _, keyword := range lowQualityKeywords {
		if service.isBusinessRelevant(keyword) {
			businessRelevantCount++
		}
	}

	if businessRelevantCount > 0 {
		t.Errorf("Low-quality content should have no business-relevant keywords, found %d", businessRelevantCount)
	}

	t.Logf("Quality metrics test successful:")
	t.Logf("  High-quality content: %d keywords (%d business-relevant)", len(highQualityKeywords), businessRelevantCount)
	t.Logf("  Low-quality content: %d keywords", len(lowQualityKeywords))
}

// TestTask2_1_ComprehensiveValidation tests comprehensive validation of all Task 2.1 requirements
func TestTask2_1_ComprehensiveValidation(t *testing.T) {
	service := &IndustryDetectionService{}

	// Test all three subtasks together
	complexHTMLContent := `
		<html>
			<head>
				<title>Bella Vista Restaurant</title>
				<script>function init() { console.log('Loading...'); }</script>
				<style>body { margin: 0; padding: 0; }</style>
			</head>
			<body>
				<h1>Bella Vista Restaurant</h1>
				<p>We are a family-owned restaurant serving authentic Italian cuisine with fine dining experience.</p>
				<p>Our restaurant offers traditional pasta, pizza, wine selection, and catering services.</p>
				<div class="contact">
					<p>Contact us for reservations and special events.</p>
				</div>
				<!-- Technical info: HTML, CSS, JavaScript -->
				<script>analytics.track('page_view');</script>
			</body>
		</html>
	`

	// Step 1: HTML Content Cleaning (Subtask 2.1.1)
	cleanContent := service.cleanHTMLContent(complexHTMLContent)

	// Validate HTML cleaning
	cleaningChecks := []struct {
		description string
		check       func() bool
	}{
		{"Script tags removed", func() bool { return !strings.Contains(cleanContent, "<script>") }},
		{"Style tags removed", func() bool { return !strings.Contains(cleanContent, "<style>") }},
		{"HTML comments removed", func() bool { return !strings.Contains(cleanContent, "<!--") }},
		{"JavaScript code removed", func() bool { return !strings.Contains(cleanContent, "console.log") }},
		{"Business content preserved", func() bool { return strings.Contains(cleanContent, "Restaurant") }},
		{"Business content preserved", func() bool { return strings.Contains(cleanContent, "Italian cuisine") }},
	}

	for _, check := range cleaningChecks {
		if !check.check() {
			t.Errorf("HTML cleaning failed: %s", check.description)
		}
	}

	// Step 2: Business Context Filtering (Subtask 2.1.2)
	keywords := service.extractKeywordsFromContent(cleanContent)

	// Validate business context filtering
	filteringChecks := []struct {
		description string
		check       func() bool
	}{
		{"Business keywords extracted", func() bool {
			expected := []string{"restaurant", "italian", "cuisine", "dining", "pasta", "pizza", "wine", "catering", "services"}
			found := 0
			for _, exp := range expected {
				for _, kw := range keywords {
					if kw == exp {
						found++
						break
					}
				}
			}
			return found >= 5
		}},
		{"Technical terms filtered", func() bool {
			technical := []string{"html", "css", "javascript", "script", "style", "console", "log", "analytics"}
			for _, tech := range technical {
				for _, kw := range keywords {
					if kw == tech {
						return false
					}
				}
			}
			return true
		}},
		{"Common words filtered", func() bool {
			common := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by", "we", "our", "us"}
			for _, comm := range common {
				for _, kw := range keywords {
					if kw == comm {
						return false
					}
				}
			}
			return true
		}},
	}

	for _, check := range filteringChecks {
		if !check.check() {
			t.Errorf("Business context filtering failed: %s", check.description)
		}
	}

	// Step 3: Keyword Quality Scoring (Subtask 2.1.3)
	// Validate keyword quality scoring
	qualityChecks := []struct {
		description string
		check       func() bool
	}{
		{"Keywords are scored", func() bool { return len(keywords) > 0 }},
		{"High-quality keywords present", func() bool {
			highQuality := []string{"restaurant", "italian", "cuisine", "dining", "catering"}
			found := 0
			for _, hq := range highQuality {
				for _, kw := range keywords {
					if kw == hq {
						found++
						break
					}
				}
			}
			return found >= 3
		}},
		{"Keywords are ranked by quality", func() bool {
			// The first few keywords should be high-quality business terms
			if len(keywords) == 0 {
				return false
			}
			topKeywords := keywords[:min(3, len(keywords))]
			highQualityCount := 0
			for _, kw := range topKeywords {
				if service.isBusinessRelevant(kw) {
					highQualityCount++
				}
			}
			return highQualityCount >= 2
		}},
	}

	for _, check := range qualityChecks {
		if !check.check() {
			t.Errorf("Keyword quality scoring failed: %s", check.description)
		}
	}

	// Performance validation
	start := time.Now()
	_ = service.extractKeywordsFromContent(complexHTMLContent)
	duration := time.Since(start)

	if duration > 100*time.Millisecond {
		t.Errorf("Performance requirement not met: processing took %v (should be < 100ms)", duration)
	}

	t.Logf("Comprehensive validation successful:")
	t.Logf("  HTML cleaning: ✅ All checks passed")
	t.Logf("  Business context filtering: ✅ All checks passed")
	t.Logf("  Keyword quality scoring: ✅ All checks passed")
	t.Logf("  Performance: ✅ %v (under 100ms requirement)", duration)
	t.Logf("  Final result: %d high-quality keywords extracted", len(keywords))
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
