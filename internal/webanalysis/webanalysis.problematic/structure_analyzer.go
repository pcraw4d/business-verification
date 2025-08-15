package webanalysis

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// WebsiteStructureAnalyzer analyzes website structure and navigation
type WebsiteStructureAnalyzer struct {
	navigationAnalyzer    *NavigationAnalyzer
	contentExtractor      *ContentExtractor
	contentAggregator     *ContentAggregator
	relevanceScorer       *StructureRelevanceScorer
	businessInfoExtractor *BusinessInfoExtractor
}

// StructureAnalysisResult represents website structure analysis results
type StructureAnalysisResult struct {
	WebsiteURL          string                 `json:"website_url"`
	NavigationStructure *NavigationStructure   `json:"navigation_structure"`
	BusinessInformation *BusinessInformation   `json:"business_information"`
	ContentAggregation  *ContentAggregation    `json:"content_aggregation"`
	StructureRelevance  *StructureRelevance    `json:"structure_relevance"`
	AnalysisTime        time.Time              `json:"analysis_time"`
	AnalysisMetadata    map[string]interface{} `json:"analysis_metadata"`
}

// NavigationStructure represents website navigation structure
type NavigationStructure struct {
	MainNavigation   []NavigationItem `json:"main_navigation"`
	FooterNavigation []NavigationItem `json:"footer_navigation"`
	HeaderElements   []HeaderElement  `json:"header_elements"`
	Breadcrumbs      []BreadcrumbItem `json:"breadcrumbs"`
	SiteMap          []SiteMapItem    `json:"site_map"`
	NavigationDepth  int              `json:"navigation_depth"`
	TotalPages       int              `json:"total_pages"`
	StructureQuality float64          `json:"structure_quality"`
}

// NavigationItem represents a navigation menu item
type NavigationItem struct {
	Text        string           `json:"text"`
	URL         string           `json:"url"`
	Type        string           `json:"type"` // main, footer, header, etc.
	Priority    float64          `json:"priority"`
	Relevance   float64          `json:"relevance"`
	HasChildren bool             `json:"has_children"`
	Children    []NavigationItem `json:"children,omitempty"`
}

// HeaderElement represents header elements
type HeaderElement struct {
	Type      string  `json:"type"` // logo, contact, social, etc.
	Content   string  `json:"content"`
	URL       string  `json:"url,omitempty"`
	Relevance float64 `json:"relevance"`
}

// BreadcrumbItem represents breadcrumb navigation
type BreadcrumbItem struct {
	Text     string `json:"text"`
	URL      string `json:"url"`
	Position int    `json:"position"`
}

// SiteMapItem represents sitemap structure
type SiteMapItem struct {
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Type         string    `json:"type"`
	Priority     float64   `json:"priority"`
	LastModified time.Time `json:"last_modified,omitempty"`
}

// BusinessInformation represents extracted business information
type BusinessInformation struct {
	CompanyName     string            `json:"company_name"`
	Address         string            `json:"address"`
	Phone           string            `json:"phone"`
	Email           string            `json:"email"`
	SocialMedia     map[string]string `json:"social_media"`
	BusinessHours   string            `json:"business_hours"`
	Services        []string          `json:"services"`
	Products        []string          `json:"products"`
	AboutInfo       string            `json:"about_info"`
	ContactInfo     string            `json:"contact_info"`
	ExtractionScore float64           `json:"extraction_score"`
}

// ContentAggregation represents aggregated content from multiple pages
type ContentAggregation struct {
	MainContent     string            `json:"main_content"`
	AboutContent    string            `json:"about_content"`
	ServicesContent string            `json:"services_content"`
	ContactContent  string            `json:"contact_content"`
	FooterContent   string            `json:"footer_content"`
	HeaderContent   string            `json:"header_content"`
	MetaContent     map[string]string `json:"meta_content"`
	AggregatedText  string            `json:"aggregated_text"`
	ContentQuality  float64           `json:"content_quality"`
	PageCount       int               `json:"page_count"`
}

// StructureRelevance represents structure-based relevance scoring
type StructureRelevance struct {
	OverallRelevance    float64            `json:"overall_relevance"`
	NavigationRelevance float64            `json:"navigation_relevance"`
	ContentRelevance    float64            `json:"content_relevance"`
	BusinessRelevance   float64            `json:"business_relevance"`
	StructureQuality    float64            `json:"structure_quality"`
	RelevanceFactors    map[string]float64 `json:"relevance_factors"`
	ConfidenceScore     float64            `json:"confidence_score"`
}

// NewWebsiteStructureAnalyzer creates a new website structure analyzer
func NewWebsiteStructureAnalyzer() *WebsiteStructureAnalyzer {
	return &WebsiteStructureAnalyzer{
		navigationAnalyzer:    NewNavigationAnalyzer(),
		contentExtractor:      NewContentExtractor(),
		contentAggregator:     NewContentAggregator(),
		relevanceScorer:       NewStructureRelevanceScorer(),
		businessInfoExtractor: NewBusinessInfoExtractor(),
	}
}

// AnalyzeWebsiteStructure performs comprehensive website structure analysis
func (sa *WebsiteStructureAnalyzer) AnalyzeWebsiteStructure(ctx context.Context, website string, pages []*ScrapedContent) (*StructureAnalysisResult, error) {
	startTime := time.Now()

	// Step 1: Analyze navigation structure
	navigationStructure, err := sa.navigationAnalyzer.AnalyzeNavigation(website, pages)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze navigation: %w", err)
	}

	// Step 2: Extract business information
	businessInformation, err := sa.businessInfoExtractor.ExtractBusinessInfo(pages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract business information: %w", err)
	}

	// Step 3: Aggregate content from multiple pages
	contentAggregation, err := sa.contentAggregator.AggregateContent(pages)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate content: %w", err)
	}

	// Step 4: Calculate structure-based relevance
	structureRelevance, err := sa.relevanceScorer.CalculateRelevance(navigationStructure, businessInformation, contentAggregation)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate relevance: %w", err)
	}

	// Create analysis metadata
	metadata := map[string]interface{}{
		"website":           website,
		"pages_analyzed":    len(pages),
		"analysis_duration": time.Since(startTime).String(),
		"navigation_depth":  navigationStructure.NavigationDepth,
		"total_pages":       navigationStructure.TotalPages,
		"structure_quality": navigationStructure.StructureQuality,
	}

	result := &StructureAnalysisResult{
		WebsiteURL:          website,
		NavigationStructure: navigationStructure,
		BusinessInformation: businessInformation,
		ContentAggregation:  contentAggregation,
		StructureRelevance:  structureRelevance,
		AnalysisTime:        time.Now(),
		AnalysisMetadata:    metadata,
	}

	return result, nil
}

// NavigationAnalyzer analyzes website navigation structure
type NavigationAnalyzer struct {
	navPatterns    []*regexp.Regexp
	headerPatterns []*regexp.Regexp
	footerPatterns []*regexp.Regexp
}

// NewNavigationAnalyzer creates a new navigation analyzer
func NewNavigationAnalyzer() *NavigationAnalyzer {
	na := &NavigationAnalyzer{}
	na.initializePatterns()
	return na
}

// AnalyzeNavigation analyzes website navigation structure
func (na *NavigationAnalyzer) AnalyzeNavigation(website string, pages []*ScrapedContent) (*NavigationStructure, error) {
	var mainNavigation []NavigationItem
	var footerNavigation []NavigationItem
	var headerElements []HeaderElement
	var breadcrumbs []BreadcrumbItem
	var siteMap []SiteMapItem

	// Analyze each page for navigation elements
	for _, page := range pages {
		// Extract main navigation
		mainNav := na.extractMainNavigation(page)
		mainNavigation = append(mainNavigation, mainNav...)

		// Extract footer navigation
		footerNav := na.extractFooterNavigation(page)
		footerNavigation = append(footerNavigation, footerNav...)

		// Extract header elements
		headerElems := na.extractHeaderElements(page)
		headerElements = append(headerElements, headerElems...)

		// Extract breadcrumbs
		breadcrumb := na.extractBreadcrumbs(page)
		breadcrumbs = append(breadcrumbs, breadcrumb...)

		// Add to sitemap
		siteMapItem := SiteMapItem{
			URL:   page.URL,
			Title: page.Title,
			Type:  na.determinePageType(page.URL),
		}
		siteMap = append(siteMap, siteMapItem)
	}

	// Remove duplicates
	mainNavigation = na.removeDuplicateNavigation(mainNavigation)
	footerNavigation = na.removeDuplicateNavigation(footerNavigation)
	headerElements = na.removeDuplicateHeaderElements(headerElements)

	// Calculate navigation depth
	navigationDepth := na.calculateNavigationDepth(mainNavigation)

	// Calculate structure quality
	structureQuality := na.calculateStructureQuality(mainNavigation, footerNavigation, headerElements)

	structure := &NavigationStructure{
		MainNavigation:   mainNavigation,
		FooterNavigation: footerNavigation,
		HeaderElements:   headerElements,
		Breadcrumbs:      breadcrumbs,
		SiteMap:          siteMap,
		NavigationDepth:  navigationDepth,
		TotalPages:       len(pages),
		StructureQuality: structureQuality,
	}

	return structure, nil
}

// initializePatterns initializes navigation patterns
func (na *NavigationAnalyzer) initializePatterns() {
	// Navigation patterns
	na.navPatterns = []*regexp.Regexp{
		regexp.MustCompile(`<nav[^>]*>(.*?)</nav>`),
		regexp.MustCompile(`<ul[^>]*class="[^"]*nav[^"]*"[^>]*>(.*?)</ul>`),
		regexp.MustCompile(`<div[^>]*class="[^"]*menu[^"]*"[^>]*>(.*?)</div>`),
	}

	// Header patterns
	na.headerPatterns = []*regexp.Regexp{
		regexp.MustCompile(`<header[^>]*>(.*?)</header>`),
		regexp.MustCompile(`<div[^>]*class="[^"]*header[^"]*"[^>]*>(.*?)</div>`),
	}

	// Footer patterns
	na.footerPatterns = []*regexp.Regexp{
		regexp.MustCompile(`<footer[^>]*>(.*?)</footer>`),
		regexp.MustCompile(`<div[^>]*class="[^"]*footer[^"]*"[^>]*>(.*?)</div>`),
	}
}

// extractMainNavigation extracts main navigation items
func (na *NavigationAnalyzer) extractMainNavigation(page *ScrapedContent) []NavigationItem {
	var items []NavigationItem

	for _, pattern := range na.navPatterns {
		matches := pattern.FindAllStringSubmatch(page.HTML, -1)
		for _, match := range matches {
			if len(match) > 1 {
				navItems := na.parseNavigationItems(match[1], "main")
				items = append(items, navItems...)
			}
		}
	}

	return items
}

// extractFooterNavigation extracts footer navigation items
func (na *NavigationAnalyzer) extractFooterNavigation(page *ScrapedContent) []NavigationItem {
	var items []NavigationItem

	for _, pattern := range na.footerPatterns {
		matches := pattern.FindAllStringSubmatch(page.HTML, -1)
		for _, match := range matches {
			if len(match) > 1 {
				navItems := na.parseNavigationItems(match[1], "footer")
				items = append(items, navItems...)
			}
		}
	}

	return items
}

// extractHeaderElements extracts header elements
func (na *NavigationAnalyzer) extractHeaderElements(page *ScrapedContent) []HeaderElement {
	var elements []HeaderElement

	for _, pattern := range na.headerPatterns {
		matches := pattern.FindAllStringSubmatch(page.HTML, -1)
		for _, match := range matches {
			if len(match) > 1 {
				headerElems := na.parseHeaderElements(match[1])
				elements = append(elements, headerElems...)
			}
		}
	}

	return elements
}

// extractBreadcrumbs extracts breadcrumb navigation
func (na *NavigationAnalyzer) extractBreadcrumbs(page *ScrapedContent) []BreadcrumbItem {
	var breadcrumbs []BreadcrumbItem

	// Look for breadcrumb patterns
	breadcrumbPattern := regexp.MustCompile(`<nav[^>]*class="[^"]*breadcrumb[^"]*"[^>]*>(.*?)</nav>`)
	matches := breadcrumbPattern.FindAllStringSubmatch(page.HTML, -1)

	for _, match := range matches {
		if len(match) > 1 {
			items := na.parseBreadcrumbItems(match[1])
			breadcrumbs = append(breadcrumbs, items...)
		}
	}

	return breadcrumbs
}

// parseNavigationItems parses navigation items from HTML
func (na *NavigationAnalyzer) parseNavigationItems(html string, navType string) []NavigationItem {
	var items []NavigationItem

	// Extract links
	linkPattern := regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	matches := linkPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			text := strings.TrimSpace(match[2])
			url := match[1]

			// Clean HTML tags from text
			text = na.cleanHTMLTags(text)

			if text != "" && url != "" {
				item := NavigationItem{
					Text:      text,
					URL:       url,
					Type:      navType,
					Priority:  na.calculateNavigationPriority(text, navType),
					Relevance: na.calculateNavigationRelevance(text),
				}
				items = append(items, item)
			}
		}
	}

	return items
}

// parseHeaderElements parses header elements from HTML
func (na *NavigationAnalyzer) parseHeaderElements(html string) []HeaderElement {
	var elements []HeaderElement

	// Extract logo
	logoPattern := regexp.MustCompile(`<img[^>]*alt="[^"]*logo[^"]*"[^>]*>`)
	if logoPattern.MatchString(html) {
		elements = append(elements, HeaderElement{
			Type:      "logo",
			Content:   "Logo found",
			Relevance: 0.8,
		})
	}

	// Extract contact information
	contactPattern := regexp.MustCompile(`<[^>]*class="[^"]*contact[^"]*"[^>]*>(.*?)</[^>]*>`)
	matches := contactPattern.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			content := na.cleanHTMLTags(match[1])
			if content != "" {
				elements = append(elements, HeaderElement{
					Type:      "contact",
					Content:   content,
					Relevance: 0.7,
				})
			}
		}
	}

	return elements
}

// parseBreadcrumbItems parses breadcrumb items from HTML
func (na *NavigationAnalyzer) parseBreadcrumbItems(html string) []BreadcrumbItem {
	var items []BreadcrumbItem

	linkPattern := regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	matches := linkPattern.FindAllStringSubmatch(html, -1)

	for i, match := range matches {
		if len(match) > 2 {
			text := na.cleanHTMLTags(match[2])
			url := match[1]

			if text != "" {
				items = append(items, BreadcrumbItem{
					Text:     text,
					URL:      url,
					Position: i,
				})
			}
		}
	}

	return items
}

// cleanHTMLTags removes HTML tags from text
func (na *NavigationAnalyzer) cleanHTMLTags(text string) string {
	// Remove HTML tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	text = tagPattern.ReplaceAllString(text, "")

	// Remove extra whitespace
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	return text
}

// calculateNavigationPriority calculates priority for navigation items
func (na *NavigationAnalyzer) calculateNavigationPriority(text string, navType string) float64 {
	priority := 0.5

	// Higher priority for main navigation
	if navType == "main" {
		priority += 0.3
	}

	// Higher priority for important pages
	importantPages := []string{"about", "services", "products", "contact", "home"}
	textLower := strings.ToLower(text)
	for _, important := range importantPages {
		if strings.Contains(textLower, important) {
			priority += 0.2
			break
		}
	}

	return priority
}

// calculateNavigationRelevance calculates relevance for navigation items
func (na *NavigationAnalyzer) calculateNavigationRelevance(text string) float64 {
	relevance := 0.5

	// Higher relevance for business-related terms
	businessTerms := []string{"about", "services", "products", "contact", "company", "business"}
	textLower := strings.ToLower(text)
	for _, term := range businessTerms {
		if strings.Contains(textLower, term) {
			relevance += 0.1
		}
	}

	if relevance > 1.0 {
		relevance = 1.0
	}

	return relevance
}

// determinePageType determines the type of a page based on URL
func (na *NavigationAnalyzer) determinePageType(url string) string {
	urlLower := strings.ToLower(url)

	if strings.Contains(urlLower, "about") {
		return "about"
	} else if strings.Contains(urlLower, "services") {
		return "services"
	} else if strings.Contains(urlLower, "products") {
		return "products"
	} else if strings.Contains(urlLower, "contact") {
		return "contact"
	} else if strings.Contains(urlLower, "home") || urlLower == "/" {
		return "home"
	}

	return "other"
}

// calculateNavigationDepth calculates the depth of navigation
func (na *NavigationAnalyzer) calculateNavigationDepth(navigation []NavigationItem) int {
	maxDepth := 0

	for _, item := range navigation {
		depth := na.calculateItemDepth(item)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// calculateItemDepth calculates the depth of a navigation item
func (na *NavigationAnalyzer) calculateItemDepth(item NavigationItem) int {
	if len(item.Children) == 0 {
		return 1
	}

	maxChildDepth := 0
	for _, child := range item.Children {
		childDepth := na.calculateItemDepth(child)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return 1 + maxChildDepth
}

// calculateStructureQuality calculates the quality of navigation structure
func (na *NavigationAnalyzer) calculateStructureQuality(mainNav, footerNav []NavigationItem, headerElems []HeaderElement) float64 {
	quality := 0.0

	// Quality based on main navigation
	if len(mainNav) > 0 {
		quality += 0.4
	}

	// Quality based on footer navigation
	if len(footerNav) > 0 {
		quality += 0.2
	}

	// Quality based on header elements
	if len(headerElems) > 0 {
		quality += 0.2
	}

	// Quality based on navigation depth
	if len(mainNav) > 3 {
		quality += 0.2
	}

	return quality
}

// removeDuplicateNavigation removes duplicate navigation items
func (na *NavigationAnalyzer) removeDuplicateNavigation(items []NavigationItem) []NavigationItem {
	seen := make(map[string]bool)
	var unique []NavigationItem

	for _, item := range items {
		key := item.Text + "|" + item.URL
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}

	return unique
}

// removeDuplicateHeaderElements removes duplicate header elements
func (na *NavigationAnalyzer) removeDuplicateHeaderElements(elements []HeaderElement) []HeaderElement {
	seen := make(map[string]bool)
	var unique []HeaderElement

	for _, element := range elements {
		key := element.Type + "|" + element.Content
		if !seen[key] {
			seen[key] = true
			unique = append(unique, element)
		}
	}

	return unique
}

// ContentExtractor extracts content from specific page sections
type ContentExtractor struct {
	sectionPatterns map[string]*regexp.Regexp
}

// NewContentExtractor creates a new content extractor
func NewContentExtractor() *ContentExtractor {
	ce := &ContentExtractor{}
	ce.initializePatterns()
	return ce
}

// initializePatterns initializes content extraction patterns
func (ce *ContentExtractor) initializePatterns() {
	ce.sectionPatterns = map[string]*regexp.Regexp{
		"header":  regexp.MustCompile(`<header[^>]*>(.*?)</header>`),
		"footer":  regexp.MustCompile(`<footer[^>]*>(.*?)</footer>`),
		"main":    regexp.MustCompile(`<main[^>]*>(.*?)</main>`),
		"article": regexp.MustCompile(`<article[^>]*>(.*?)</article>`),
		"section": regexp.MustCompile(`<section[^>]*>(.*?)</section>`),
		"aside":   regexp.MustCompile(`<aside[^>]*>(.*?)</aside>`),
	}
}

// ExtractContent extracts content from specific sections
func (ce *ContentExtractor) ExtractContent(page *ScrapedContent, section string) string {
	pattern, exists := ce.sectionPatterns[section]
	if !exists {
		return ""
	}

	matches := pattern.FindAllStringSubmatch(page.HTML, -1)
	if len(matches) == 0 {
		return ""
	}

	var content strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			content.WriteString(match[1])
			content.WriteString(" ")
		}
	}

	return strings.TrimSpace(content.String())
}

// ContentAggregator aggregates content from multiple pages
type ContentAggregator struct {
	contentExtractor *ContentExtractor
}

// NewContentAggregator creates a new content aggregator
func NewContentAggregator() *ContentAggregator {
	return &ContentAggregator{
		contentExtractor: NewContentExtractor(),
	}
}

// AggregateContent aggregates content from multiple pages
func (ca *ContentAggregator) AggregateContent(pages []*ScrapedContent) (*ContentAggregation, error) {
	var mainContent, aboutContent, servicesContent, contactContent, footerContent, headerContent strings.Builder
	metaContent := make(map[string]string)

	// Aggregate content from each page
	for _, page := range pages {
		// Main content
		mainContent.WriteString(page.Text)
		mainContent.WriteString(" ")

		// Section-specific content
		headerContent.WriteString(ca.contentExtractor.ExtractContent(page, "header"))
		headerContent.WriteString(" ")
		footerContent.WriteString(ca.contentExtractor.ExtractContent(page, "footer"))
		footerContent.WriteString(" ")

		// Page-specific content based on URL
		pageType := ca.determinePageType(page.URL)
		switch pageType {
		case "about":
			aboutContent.WriteString(page.Text)
			aboutContent.WriteString(" ")
		case "services":
			servicesContent.WriteString(page.Text)
			servicesContent.WriteString(" ")
		case "contact":
			contactContent.WriteString(page.Text)
			contactContent.WriteString(" ")
		}

		// Extract meta content
		metaContent[page.URL] = page.Title
	}

	// Calculate content quality
	contentQuality := ca.calculateContentQuality(pages)

	aggregation := &ContentAggregation{
		MainContent:     strings.TrimSpace(mainContent.String()),
		AboutContent:    strings.TrimSpace(aboutContent.String()),
		ServicesContent: strings.TrimSpace(servicesContent.String()),
		ContactContent:  strings.TrimSpace(contactContent.String()),
		FooterContent:   strings.TrimSpace(footerContent.String()),
		HeaderContent:   strings.TrimSpace(headerContent.String()),
		MetaContent:     metaContent,
		AggregatedText:  strings.TrimSpace(mainContent.String()),
		ContentQuality:  contentQuality,
		PageCount:       len(pages),
	}

	return aggregation, nil
}

// determinePageType determines the type of a page based on URL
func (ca *ContentAggregator) determinePageType(url string) string {
	urlLower := strings.ToLower(url)

	if strings.Contains(urlLower, "about") {
		return "about"
	} else if strings.Contains(urlLower, "services") {
		return "services"
	} else if strings.Contains(urlLower, "contact") {
		return "contact"
	}

	return "other"
}

// calculateContentQuality calculates the quality of aggregated content
func (ca *ContentAggregator) calculateContentQuality(pages []*ScrapedContent) float64 {
	if len(pages) == 0 {
		return 0.0
	}

	totalLength := 0
	totalQuality := 0.0

	for _, page := range pages {
		length := len(page.Text)
		totalLength += length

		// Quality based on content length
		if length > 1000 {
			totalQuality += 0.8
		} else if length > 500 {
			totalQuality += 0.6
		} else if length > 100 {
			totalQuality += 0.4
		} else {
			totalQuality += 0.2
		}
	}

	return totalQuality / float64(len(pages))
}

// BusinessInfoExtractor extracts business information from pages
type BusinessInfoExtractor struct {
	patterns map[string]*regexp.Regexp
}

// NewBusinessInfoExtractor creates a new business info extractor
func NewBusinessInfoExtractor() *BusinessInfoExtractor {
	bie := &BusinessInfoExtractor{}
	bie.initializePatterns()
	return bie
}

// initializePatterns initializes business information extraction patterns
func (bie *BusinessInfoExtractor) initializePatterns() {
	bie.patterns = map[string]*regexp.Regexp{
		"phone":   regexp.MustCompile(`\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}`),
		"email":   regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		"address": regexp.MustCompile(`\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter),\s*[A-Za-z\s]+,\s*[A-Z]{2}\s+\d{5}`),
		"hours":   regexp.MustCompile(`(?i)(?:hours?|open|closed|monday|tuesday|wednesday|thursday|friday|saturday|sunday)[\s:]*[a-zA-Z0-9\s,:-]+`),
	}
}

// ExtractBusinessInfo extracts business information from pages
func (bie *BusinessInfoExtractor) ExtractBusinessInfo(pages []*ScrapedContent) (*BusinessInformation, error) {
	info := &BusinessInformation{
		SocialMedia: make(map[string]string),
		Services:    []string{},
		Products:    []string{},
	}

	// Extract information from each page
	for _, page := range pages {
		bie.extractFromPage(page, info)
	}

	// Calculate extraction score
	info.ExtractionScore = bie.calculateExtractionScore(info)

	return info, nil
}

// extractFromPage extracts business information from a single page
func (bie *BusinessInfoExtractor) extractFromPage(page *ScrapedContent, info *BusinessInformation) {
	content := page.Text + " " + page.HTML

	// Extract phone numbers
	if info.Phone == "" {
		phones := bie.patterns["phone"].FindAllString(content, -1)
		if len(phones) > 0 {
			info.Phone = phones[0]
		}
	}

	// Extract email addresses
	if info.Email == "" {
		emails := bie.patterns["email"].FindAllString(content, -1)
		if len(emails) > 0 {
			info.Email = emails[0]
		}
	}

	// Extract addresses
	if info.Address == "" {
		addresses := bie.patterns["address"].FindAllString(content, -1)
		if len(addresses) > 0 {
			info.Address = addresses[0]
		}
	}

	// Extract business hours
	if info.BusinessHours == "" {
		hours := bie.patterns["hours"].FindAllString(content, -1)
		if len(hours) > 0 {
			info.BusinessHours = hours[0]
		}
	}

	// Extract company name (from title or content)
	if info.CompanyName == "" {
		info.CompanyName = bie.extractCompanyName(page)
	}

	// Extract social media links
	bie.extractSocialMedia(content, info)

	// Extract services and products
	bie.extractServicesAndProducts(content, info)
}

// extractCompanyName extracts company name from page
func (bie *BusinessInfoExtractor) extractCompanyName(page *ScrapedContent) string {
	// Try to extract from title first
	if page.Title != "" {
		// Remove common suffixes
		title := strings.TrimSpace(page.Title)
		suffixes := []string{" - Home", " | Home", " - Welcome", " | Welcome"}
		for _, suffix := range suffixes {
			title = strings.TrimSuffix(title, suffix)
		}
		return title
	}

	// Try to extract from content
	content := page.Text
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 5 && len(line) < 100 {
			// Check if line looks like a company name
			if strings.Contains(line, "Inc") || strings.Contains(line, "LLC") || strings.Contains(line, "Corp") {
				return line
			}
		}
	}

	return ""
}

// extractSocialMedia extracts social media links
func (bie *BusinessInfoExtractor) extractSocialMedia(content string, info *BusinessInformation) {
	socialPatterns := map[string]*regexp.Regexp{
		"facebook":  regexp.MustCompile(`https?://(?:www\.)?facebook\.com/[^\s"']+`),
		"twitter":   regexp.MustCompile(`https?://(?:www\.)?twitter\.com/[^\s"']+`),
		"linkedin":  regexp.MustCompile(`https?://(?:www\.)?linkedin\.com/[^\s"']+`),
		"instagram": regexp.MustCompile(`https?://(?:www\.)?instagram\.com/[^\s"']+`),
	}

	for platform, pattern := range socialPatterns {
		matches := pattern.FindAllString(content, -1)
		if len(matches) > 0 {
			info.SocialMedia[platform] = matches[0]
		}
	}
}

// extractServicesAndProducts extracts services and products
func (bie *BusinessInfoExtractor) extractServicesAndProducts(content string, info *BusinessInformation) {
	// Simple extraction based on keywords
	serviceKeywords := []string{"service", "consulting", "solutions", "support", "maintenance"}
	productKeywords := []string{"product", "software", "hardware", "equipment", "tool"}

	contentLower := strings.ToLower(content)

	for _, keyword := range serviceKeywords {
		if strings.Contains(contentLower, keyword) {
			info.Services = append(info.Services, keyword)
		}
	}

	for _, keyword := range productKeywords {
		if strings.Contains(contentLower, keyword) {
			info.Products = append(info.Products, keyword)
		}
	}
}

// calculateExtractionScore calculates the extraction score
func (bie *BusinessInfoExtractor) calculateExtractionScore(info *BusinessInformation) float64 {
	score := 0.0

	if info.CompanyName != "" {
		score += 0.2
	}
	if info.Address != "" {
		score += 0.2
	}
	if info.Phone != "" {
		score += 0.15
	}
	if info.Email != "" {
		score += 0.15
	}
	if info.BusinessHours != "" {
		score += 0.1
	}
	if len(info.SocialMedia) > 0 {
		score += 0.1
	}
	if len(info.Services) > 0 {
		score += 0.05
	}
	if len(info.Products) > 0 {
		score += 0.05
	}

	return score
}

// StructureRelevanceScorer calculates structure-based relevance
type StructureRelevanceScorer struct {
	weights map[string]float64
}

// NewStructureRelevanceScorer creates a new structure relevance scorer
func NewStructureRelevanceScorer() *StructureRelevanceScorer {
	return &StructureRelevanceScorer{
		weights: map[string]float64{
			"navigation": 0.3,
			"content":    0.3,
			"business":   0.4,
		},
	}
}

// CalculateRelevance calculates structure-based relevance
func (srs *StructureRelevanceScorer) CalculateRelevance(
	navigation *NavigationStructure,
	business *BusinessInformation,
	content *ContentAggregation) (*StructureRelevance, error) {

	// Calculate navigation relevance
	navigationRelevance := srs.calculateNavigationRelevance(navigation)

	// Calculate content relevance
	contentRelevance := srs.calculateContentRelevance(content)

	// Calculate business relevance
	businessRelevance := srs.calculateBusinessRelevance(business)

	// Calculate overall relevance
	overallRelevance := navigationRelevance*srs.weights["navigation"] +
		contentRelevance*srs.weights["content"] +
		businessRelevance*srs.weights["business"]

	// Calculate structure quality
	structureQuality := navigation.StructureQuality

	// Calculate confidence score
	confidenceScore := (overallRelevance + structureQuality) / 2.0

	// Create relevance factors
	relevanceFactors := map[string]float64{
		"navigation_depth":  float64(navigation.NavigationDepth),
		"total_pages":       float64(navigation.TotalPages),
		"content_quality":   content.ContentQuality,
		"extraction_score":  business.ExtractionScore,
		"structure_quality": structureQuality,
	}

	relevance := &StructureRelevance{
		OverallRelevance:    overallRelevance,
		NavigationRelevance: navigationRelevance,
		ContentRelevance:    contentRelevance,
		BusinessRelevance:   businessRelevance,
		StructureQuality:    structureQuality,
		RelevanceFactors:    relevanceFactors,
		ConfidenceScore:     confidenceScore,
	}

	return relevance, nil
}

// calculateNavigationRelevance calculates navigation relevance
func (srs *StructureRelevanceScorer) calculateNavigationRelevance(navigation *NavigationStructure) float64 {
	relevance := 0.0

	// Relevance based on navigation depth
	if navigation.NavigationDepth >= 3 {
		relevance += 0.3
	} else if navigation.NavigationDepth >= 2 {
		relevance += 0.2
	} else if navigation.NavigationDepth >= 1 {
		relevance += 0.1
	}

	// Relevance based on total pages
	if navigation.TotalPages >= 10 {
		relevance += 0.3
	} else if navigation.TotalPages >= 5 {
		relevance += 0.2
	} else if navigation.TotalPages >= 3 {
		relevance += 0.1
	}

	// Relevance based on structure quality
	relevance += navigation.StructureQuality * 0.4

	return relevance
}

// calculateContentRelevance calculates content relevance
func (srs *StructureRelevanceScorer) calculateContentRelevance(content *ContentAggregation) float64 {
	relevance := 0.0

	// Relevance based on content quality
	relevance += content.ContentQuality * 0.4

	// Relevance based on page count
	if content.PageCount >= 5 {
		relevance += 0.3
	} else if content.PageCount >= 3 {
		relevance += 0.2
	} else if content.PageCount >= 1 {
		relevance += 0.1
	}

	// Relevance based on content length
	contentLength := len(content.AggregatedText)
	if contentLength >= 10000 {
		relevance += 0.3
	} else if contentLength >= 5000 {
		relevance += 0.2
	} else if contentLength >= 1000 {
		relevance += 0.1
	}

	return relevance
}

// calculateBusinessRelevance calculates business relevance
func (srs *StructureRelevanceScorer) calculateBusinessRelevance(business *BusinessInformation) float64 {
	relevance := 0.0

	// Relevance based on extraction score
	relevance += business.ExtractionScore * 0.6

	// Relevance based on completeness
	completeness := 0.0
	if business.CompanyName != "" {
		completeness += 0.2
	}
	if business.Address != "" {
		completeness += 0.2
	}
	if business.Phone != "" {
		completeness += 0.15
	}
	if business.Email != "" {
		completeness += 0.15
	}
	if len(business.SocialMedia) > 0 {
		completeness += 0.1
	}
	if len(business.Services) > 0 || len(business.Products) > 0 {
		completeness += 0.2
	}

	relevance += completeness * 0.4

	return relevance
}
