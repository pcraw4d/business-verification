package webanalysis

import (
	"math"
	"regexp"
	"strings"
)

// StructureAnalyzer analyzes content structure and organization
type StructureAnalyzer struct{}

// NewStructureAnalyzer creates a new structure analyzer
func NewStructureAnalyzer() *StructureAnalyzer {
	return &StructureAnalyzer{}
}

// AnalyzeStructure analyzes the structure and organization of content
func (sa *StructureAnalyzer) AnalyzeStructure(html, text string) StructureMetrics {
	metrics := StructureMetrics{}

	// Heading structure analysis
	metrics.HeadingStructure = sa.analyzeHeadingStructure(html)

	// Paragraph structure analysis
	metrics.ParagraphStructure = sa.analyzeParagraphStructure(html, text)

	// List structure analysis
	metrics.ListStructure = sa.analyzeListStructure(html)

	// Table structure analysis
	metrics.TableStructure = sa.analyzeTableStructure(html)

	// Navigation structure analysis
	metrics.NavigationStructure = sa.analyzeNavigationStructure(html)

	// Content organization analysis
	metrics.ContentOrganization = sa.analyzeContentOrganization(text)

	// Logical flow analysis
	metrics.LogicalFlow = sa.analyzeLogicalFlow(text)

	// Calculate overall structure score
	metrics.StructureScore = (metrics.HeadingStructure + metrics.ParagraphStructure +
		metrics.ListStructure + metrics.TableStructure + metrics.NavigationStructure +
		metrics.ContentOrganization + metrics.LogicalFlow) / 7.0

	return metrics
}

// analyzeHeadingStructure analyzes heading hierarchy and structure
func (sa *StructureAnalyzer) analyzeHeadingStructure(html string) float64 {
	score := 0.0

	// Check for H1 presence
	if strings.Contains(html, "<h1>") {
		score += 0.2
	}

	// Check for H2 presence
	if strings.Contains(html, "<h2>") {
		score += 0.15
	}

	// Check for H3 presence
	if strings.Contains(html, "<h3>") {
		score += 0.1
	}

	// Check for proper heading hierarchy
	h1Count := strings.Count(html, "<h1>")
	h2Count := strings.Count(html, "<h2>")
	h3Count := strings.Count(html, "<h3>")

	if h1Count > 0 && h2Count > 0 {
		score += 0.15
	}
	if h2Count > 0 && h3Count > 0 {
		score += 0.1
	}

	// Check for reasonable heading distribution
	if h1Count <= 2 && h2Count <= 10 && h3Count <= 20 {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// analyzeParagraphStructure analyzes paragraph organization
func (sa *StructureAnalyzer) analyzeParagraphStructure(html, text string) float64 {
	score := 0.0

	// Count paragraphs
	paragraphCount := strings.Count(html, "<p>")

	// Check for reasonable paragraph count
	if paragraphCount >= 3 && paragraphCount <= 50 {
		score += 0.3
	}

	// Check for paragraph length variety
	paragraphs := regexp.MustCompile(`<p[^>]*>(.*?)</p>`).FindAllString(html, -1)
	if len(paragraphs) > 0 {
		avgLength := 0
		for _, p := range paragraphs {
			avgLength += len(p)
		}
		avgLength /= len(paragraphs)

		if avgLength >= 50 && avgLength <= 500 {
			score += 0.3
		}
	}

	// Check for logical paragraph breaks
	if strings.Contains(text, ". ") {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// analyzeListStructure analyzes list organization
func (sa *StructureAnalyzer) analyzeListStructure(html string) float64 {
	score := 0.0

	// Check for unordered lists
	if strings.Contains(html, "<ul>") {
		score += 0.3
	}

	// Check for ordered lists
	if strings.Contains(html, "<ol>") {
		score += 0.3
	}

	// Check for list items
	listItemCount := strings.Count(html, "<li>")
	if listItemCount >= 3 && listItemCount <= 20 {
		score += 0.2
	}

	// Check for nested lists
	if strings.Contains(html, "<ul><ul>") || strings.Contains(html, "<ol><ol>") {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// analyzeTableStructure analyzes table organization
func (sa *StructureAnalyzer) analyzeTableStructure(html string) float64 {
	score := 0.0

	// Check for table presence
	if strings.Contains(html, "<table>") {
		score += 0.4
	}

	// Check for table headers
	if strings.Contains(html, "<th>") {
		score += 0.3
	}

	// Check for table rows
	rowCount := strings.Count(html, "<tr>")
	if rowCount >= 2 && rowCount <= 20 {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// analyzeNavigationStructure analyzes navigation elements
func (sa *StructureAnalyzer) analyzeNavigationStructure(html string) float64 {
	score := 0.0

	// Check for navigation elements
	if strings.Contains(html, "<nav>") {
		score += 0.4
	}

	// Check for menu elements
	if strings.Contains(html, "menu") || strings.Contains(html, "navigation") {
		score += 0.3
	}

	// Check for breadcrumbs
	if strings.Contains(html, "breadcrumb") {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// analyzeContentOrganization analyzes overall content organization
func (sa *StructureAnalyzer) analyzeContentOrganization(text string) float64 {
	score := 0.0

	// Check for introduction
	if strings.Contains(strings.ToLower(text), "introduction") ||
		strings.Contains(strings.ToLower(text), "about") {
		score += 0.3
	}

	// Check for conclusion
	if strings.Contains(strings.ToLower(text), "conclusion") ||
		strings.Contains(strings.ToLower(text), "summary") {
		score += 0.3
	}

	// Check for logical sections
	sections := []string{"services", "products", "contact", "team", "history"}
	for _, section := range sections {
		if strings.Contains(strings.ToLower(text), section) {
			score += 0.1
		}
	}

	return math.Min(score, 1.0)
}

// analyzeLogicalFlow analyzes logical content flow
func (sa *StructureAnalyzer) analyzeLogicalFlow(text string) float64 {
	score := 0.0

	// Check for transition words
	transitions := []string{"first", "second", "third", "next", "then", "finally", "however", "therefore", "in addition"}
	for _, transition := range transitions {
		if strings.Contains(strings.ToLower(text), transition) {
			score += 0.1
		}
	}

	// Check for logical connectors
	connectors := []string{"because", "since", "as a result", "consequently", "for example"}
	for _, connector := range connectors {
		if strings.Contains(strings.ToLower(text), connector) {
			score += 0.1
		}
	}

	return math.Min(score, 1.0)
}

// CompletenessAnalyzer analyzes content completeness
type CompletenessAnalyzer struct{}

// NewCompletenessAnalyzer creates a new completeness analyzer
func NewCompletenessAnalyzer() *CompletenessAnalyzer {
	return &CompletenessAnalyzer{}
}

// AnalyzeCompleteness analyzes content completeness and information density
func (ca *CompletenessAnalyzer) AnalyzeCompleteness(content *ScrapedContent) CompletenessMetrics {
	metrics := CompletenessMetrics{}

	// Content length
	metrics.ContentLength = len(content.Text)

	// Information density
	metrics.InformationDensity = ca.calculateInformationDensity(content.Text)

	// Factual content
	metrics.FactualContent = ca.analyzeFactualContent(content.Text)

	// Descriptive content
	metrics.DescriptiveContent = ca.analyzeDescriptiveContent(content.Text)

	// Contact information
	metrics.ContactInformation = ca.analyzeContactInformation(content.Text)

	// Business information
	metrics.BusinessInformation = ca.analyzeBusinessInformation(content.Text)

	// Service information
	metrics.ServiceInformation = ca.analyzeServiceInformation(content.Text)

	// Calculate overall completeness score
	metrics.CompletenessScore = (metrics.InformationDensity + metrics.FactualContent +
		metrics.DescriptiveContent + metrics.ContactInformation +
		metrics.BusinessInformation + metrics.ServiceInformation) / 6.0

	return metrics
}

// calculateInformationDensity calculates information density
func (ca *CompletenessAnalyzer) calculateInformationDensity(text string) float64 {
	if len(text) == 0 {
		return 0.0
	}

	// Count sentences
	sentences := regexp.MustCompile(`[.!?]+`).Split(text, -1)
	sentenceCount := len(sentences)

	// Count words
	words := regexp.MustCompile(`\s+`).Split(strings.TrimSpace(text), -1)
	wordCount := len(words)

	// Calculate information density
	if sentenceCount > 0 {
		return math.Min(float64(wordCount)/float64(sentenceCount)/20.0, 1.0)
	}
	return 0.0
}

// analyzeFactualContent analyzes factual information presence
func (ca *CompletenessAnalyzer) analyzeFactualContent(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for factual indicators
	factualIndicators := []string{"established", "founded", "since", "years", "experience", "certified", "licensed"}
	for _, indicator := range factualIndicators {
		if strings.Contains(text, indicator) {
			score += 0.15
		}
	}

	// Check for numbers and statistics
	numberPattern := regexp.MustCompile(`\d+`)
	if numberPattern.MatchString(text) {
		score += 0.2
	}

	// Check for specific dates
	datePattern := regexp.MustCompile(`\d{4}`)
	if datePattern.MatchString(text) {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// analyzeDescriptiveContent analyzes descriptive content
func (ca *CompletenessAnalyzer) analyzeDescriptiveContent(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for descriptive adjectives
	descriptiveWords := []string{"professional", "expert", "quality", "reliable", "trusted", "innovative", "leading"}
	for _, word := range descriptiveWords {
		if strings.Contains(text, word) {
			score += 0.1
		}
	}

	// Check for detailed descriptions
	if strings.Contains(text, "specializing in") || strings.Contains(text, "focusing on") {
		score += 0.3
	}

	// Check for benefit statements
	if strings.Contains(text, "benefit") || strings.Contains(text, "advantage") {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// analyzeContactInformation analyzes contact information presence
func (ca *CompletenessAnalyzer) analyzeContactInformation(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for email
	if strings.Contains(text, "@") {
		score += 0.3
	}

	// Check for phone
	if strings.Contains(text, "phone") || strings.Contains(text, "tel") {
		score += 0.3
	}

	// Check for address
	if strings.Contains(text, "address") || strings.Contains(text, "street") {
		score += 0.2
	}

	// Check for contact form
	if strings.Contains(text, "contact") || strings.Contains(text, "form") {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// analyzeBusinessInformation analyzes business information presence
func (ca *CompletenessAnalyzer) analyzeBusinessInformation(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for company information
	businessIndicators := []string{"company", "corporation", "llc", "inc", "ltd", "business"}
	for _, indicator := range businessIndicators {
		if strings.Contains(text, indicator) {
			score += 0.2
		}
	}

	// Check for industry information
	if strings.Contains(text, "industry") || strings.Contains(text, "sector") {
		score += 0.2
	}

	// Check for mission statement
	if strings.Contains(text, "mission") || strings.Contains(text, "vision") {
		score += 0.2
	}

	// Check for values
	if strings.Contains(text, "values") || strings.Contains(text, "principles") {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// analyzeServiceInformation analyzes service information presence
func (ca *CompletenessAnalyzer) analyzeServiceInformation(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for services
	if strings.Contains(text, "services") || strings.Contains(text, "offerings") {
		score += 0.3
	}

	// Check for products
	if strings.Contains(text, "products") || strings.Contains(text, "solutions") {
		score += 0.3
	}

	// Check for expertise areas
	if strings.Contains(text, "expertise") || strings.Contains(text, "specialization") {
		score += 0.2
	}

	// Check for capabilities
	if strings.Contains(text, "capabilities") || strings.Contains(text, "features") {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// BusinessContentAnalyzer analyzes business-specific content quality
type BusinessContentAnalyzer struct{}

// NewBusinessContentAnalyzer creates a new business content analyzer
func NewBusinessContentAnalyzer() *BusinessContentAnalyzer {
	return &BusinessContentAnalyzer{}
}

// AnalyzeBusinessContent analyzes business-specific content quality
func (bca *BusinessContentAnalyzer) AnalyzeBusinessContent(content *ScrapedContent, business string) BusinessContentMetrics {
	metrics := BusinessContentMetrics{}

	// Business name presence
	metrics.BusinessNamePresence = bca.analyzeBusinessNamePresence(content.Text, business)

	// Service description
	metrics.ServiceDescription = bca.analyzeServiceDescription(content.Text)

	// Company history
	metrics.CompanyHistory = bca.analyzeCompanyHistory(content.Text)

	// Team information
	metrics.TeamInformation = bca.analyzeTeamInformation(content.Text)

	// Certifications
	metrics.Certifications = bca.analyzeCertifications(content.Text)

	// Testimonials
	metrics.Testimonials = bca.analyzeTestimonials(content.Text)

	// Case studies
	metrics.CaseStudies = bca.analyzeCaseStudies(content.Text)

	// Calculate overall business score
	metrics.BusinessScore = (metrics.BusinessNamePresence + metrics.ServiceDescription +
		metrics.CompanyHistory + metrics.TeamInformation + metrics.Certifications +
		metrics.Testimonials + metrics.CaseStudies) / 7.0

	return metrics
}

// analyzeBusinessNamePresence analyzes business name presence
func (bca *BusinessContentAnalyzer) analyzeBusinessNamePresence(text, business string) float64 {
	score := 0.0
	text = strings.ToLower(text)
	business = strings.ToLower(business)

	// Exact match
	if strings.Contains(text, business) {
		score += 0.5
	}

	// Partial match
	words := strings.Fields(business)
	for _, word := range words {
		if len(word) > 3 && strings.Contains(text, word) {
			score += 0.2
		}
	}

	return math.Min(score, 1.0)
}

// analyzeServiceDescription analyzes service description quality
func (bca *BusinessContentAnalyzer) analyzeServiceDescription(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for service descriptions
	serviceIndicators := []string{"we provide", "we offer", "our services", "specializing in", "focusing on"}
	for _, indicator := range serviceIndicators {
		if strings.Contains(text, indicator) {
			score += 0.2
		}
	}

	// Check for detailed service descriptions
	if strings.Contains(text, "service") && len(text) > 500 {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// analyzeCompanyHistory analyzes company history presence
func (bca *BusinessContentAnalyzer) analyzeCompanyHistory(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for history indicators
	historyIndicators := []string{"founded", "established", "since", "history", "heritage", "tradition"}
	for _, indicator := range historyIndicators {
		if strings.Contains(text, indicator) {
			score += 0.2
		}
	}

	// Check for founding year
	yearPattern := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	if yearPattern.MatchString(text) {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// analyzeTeamInformation analyzes team information presence
func (bca *BusinessContentAnalyzer) analyzeTeamInformation(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for team indicators
	teamIndicators := []string{"team", "staff", "employees", "professionals", "experts"}
	for _, indicator := range teamIndicators {
		if strings.Contains(text, indicator) {
			score += 0.2
		}
	}

	// Check for leadership information
	leadershipIndicators := []string{"ceo", "president", "director", "manager", "founder"}
	for _, indicator := range leadershipIndicators {
		if strings.Contains(text, indicator) {
			score += 0.2
		}
	}

	return math.Min(score, 1.0)
}

// analyzeCertifications analyzes certification presence
func (bca *BusinessContentAnalyzer) analyzeCertifications(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for certification indicators
	certIndicators := []string{"certified", "accredited", "licensed", "iso", "certification"}
	for _, indicator := range certIndicators {
		if strings.Contains(text, indicator) {
			score += 0.2
		}
	}

	return math.Min(score, 1.0)
}

// analyzeTestimonials analyzes testimonial presence
func (bca *BusinessContentAnalyzer) analyzeTestimonials(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for testimonial indicators
	testimonialIndicators := []string{"testimonial", "review", "feedback", "client", "customer"}
	for _, indicator := range testimonialIndicators {
		if strings.Contains(text, indicator) {
			score += 0.2
		}
	}

	return math.Min(score, 1.0)
}

// analyzeCaseStudies analyzes case study presence
func (bca *BusinessContentAnalyzer) analyzeCaseStudies(text string) float64 {
	score := 0.0
	text = strings.ToLower(text)

	// Check for case study indicators
	caseStudyIndicators := []string{"case study", "success story", "project", "portfolio"}
	for _, indicator := range caseStudyIndicators {
		if strings.Contains(text, indicator) {
			score += 0.3
		}
	}

	return math.Min(score, 1.0)
}

// TechnicalContentAnalyzer analyzes technical content quality
type TechnicalContentAnalyzer struct{}

// NewTechnicalContentAnalyzer creates a new technical content analyzer
func NewTechnicalContentAnalyzer() *TechnicalContentAnalyzer {
	return &TechnicalContentAnalyzer{}
}

// AnalyzeTechnicalContent analyzes technical content quality
func (tca *TechnicalContentAnalyzer) AnalyzeTechnicalContent(content *ScrapedContent) TechnicalContentMetrics {
	metrics := TechnicalContentMetrics{}

	// HTML validity
	metrics.HTMLValidity = tca.analyzeHTMLValidity(content.HTML)

	// Accessibility score
	metrics.AccessibilityScore = tca.analyzeAccessibility(content.HTML)

	// Mobile optimization
	metrics.MobileOptimization = tca.analyzeMobileOptimization(content.HTML)

	// Image optimization
	metrics.ImageOptimization = tca.analyzeImageOptimization(content.HTML)

	// Link quality
	metrics.LinkQuality = tca.analyzeLinkQuality(content.HTML)

	// Meta tag completeness
	metrics.MetaTagCompleteness = tca.analyzeMetaTagCompleteness(content.HTML)

	// Calculate overall technical score
	metrics.TechnicalScore = (metrics.HTMLValidity + metrics.AccessibilityScore +
		metrics.MobileOptimization + metrics.ImageOptimization +
		metrics.LinkQuality + metrics.MetaTagCompleteness) / 6.0

	return metrics
}

// analyzeHTMLValidity analyzes HTML validity
func (tca *TechnicalContentAnalyzer) analyzeHTMLValidity(html string) float64 {
	score := 0.0

	// Check for proper HTML structure
	if strings.Contains(html, "<html>") && strings.Contains(html, "</html>") {
		score += 0.3
	}

	// Check for proper head and body
	if strings.Contains(html, "<head>") && strings.Contains(html, "<body>") {
		score += 0.3
	}

	// Check for proper closing tags
	openTags := []string{"<div>", "<p>", "<h1>", "<h2>", "<h3>", "<ul>", "<ol>", "<li>"}
	closeTags := []string{"</div>", "</p>", "</h1>", "</h2>", "</h3>", "</ul>", "</ol>", "</li>"}

	for i, openTag := range openTags {
		openCount := strings.Count(html, openTag)
		closeCount := strings.Count(html, closeTags[i])
		if openCount == closeCount {
			score += 0.1
		}
	}

	return math.Min(score, 1.0)
}

// analyzeAccessibility analyzes accessibility features
func (tca *TechnicalContentAnalyzer) analyzeAccessibility(html string) float64 {
	score := 0.0

	// Check for alt attributes on images
	if strings.Contains(html, "alt=") {
		score += 0.3
	}

	// Check for ARIA labels
	if strings.Contains(html, "aria-label") || strings.Contains(html, "aria-labelledby") {
		score += 0.3
	}

	// Check for semantic HTML
	semanticTags := []string{"<nav>", "<main>", "<article>", "<section>", "<header>", "<footer>"}
	for _, tag := range semanticTags {
		if strings.Contains(html, tag) {
			score += 0.1
		}
	}

	return math.Min(score, 1.0)
}

// analyzeMobileOptimization analyzes mobile optimization
func (tca *TechnicalContentAnalyzer) analyzeMobileOptimization(html string) float64 {
	score := 0.0

	// Check for viewport meta tag
	if strings.Contains(html, "viewport") {
		score += 0.4
	}

	// Check for responsive design indicators
	responsiveIndicators := []string{"media", "responsive", "mobile"}
	for _, indicator := range responsiveIndicators {
		if strings.Contains(html, indicator) {
			score += 0.2
		}
	}

	return math.Min(score, 1.0)
}

// analyzeImageOptimization analyzes image optimization
func (tca *TechnicalContentAnalyzer) analyzeImageOptimization(html string) float64 {
	score := 0.0

	// Check for image tags
	if strings.Contains(html, "<img") {
		score += 0.3
	}

	// Check for alt attributes
	if strings.Contains(html, "alt=") {
		score += 0.3
	}

	// Check for image dimensions
	if strings.Contains(html, "width=") || strings.Contains(html, "height=") {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// analyzeLinkQuality analyzes link quality
func (tca *TechnicalContentAnalyzer) analyzeLinkQuality(html string) float64 {
	score := 0.0

	// Check for internal links
	if strings.Contains(html, "href=") {
		score += 0.4
	}

	// Check for descriptive link text
	if strings.Contains(html, "<a") && len(html) > 100 {
		score += 0.3
	}

	// Check for proper link structure
	if strings.Contains(html, "</a>") {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// analyzeMetaTagCompleteness analyzes meta tag completeness
func (tca *TechnicalContentAnalyzer) analyzeMetaTagCompleteness(html string) float64 {
	score := 0.0

	// Check for title tag
	if strings.Contains(html, "<title>") {
		score += 0.3
	}

	// Check for meta description
	if strings.Contains(html, "meta name=\"description\"") {
		score += 0.3
	}

	// Check for meta keywords
	if strings.Contains(html, "meta name=\"keywords\"") {
		score += 0.2
	}

	// Check for charset declaration
	if strings.Contains(html, "charset") {
		score += 0.2
	}

	return math.Min(score, 1.0)
}
