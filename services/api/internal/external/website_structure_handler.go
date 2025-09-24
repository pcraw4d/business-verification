package external

import (
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// WebsiteStructureHandler handles different website structures and formats
type WebsiteStructureHandler struct {
	logger *zap.Logger
}

// NewWebsiteStructureHandler creates a new website structure handler
func NewWebsiteStructureHandler(logger *zap.Logger) *WebsiteStructureHandler {
	return &WebsiteStructureHandler{
		logger: logger,
	}
}

// WebsiteStructure represents the detected structure of a website
type WebsiteStructure struct {
	Type         string            `json:"type"`
	Framework    string            `json:"framework,omitempty"`
	Template     string            `json:"template,omitempty"`
	Layout       string            `json:"layout,omitempty"`
	ContentAreas []ContentArea     `json:"content_areas,omitempty"`
	Navigation   []NavigationItem  `json:"navigation,omitempty"`
	Forms        []Form            `json:"forms,omitempty"`
	Tables       []Table           `json:"tables,omitempty"`
	Lists        []List            `json:"lists,omitempty"`
	Confidence   float64           `json:"confidence"`
	Metadata     map[string]string `json:"metadata"`
}

// ContentArea represents a content area on the page
type ContentArea struct {
	Type       string `json:"type"` // header, footer, main, sidebar, etc.
	Selector   string `json:"selector,omitempty"`
	Content    string `json:"content,omitempty"`
	Importance int    `json:"importance"` // 1-10 scale
}

// NavigationItem represents a navigation item
type NavigationItem struct {
	Text     string           `json:"text"`
	URL      string           `json:"url,omitempty"`
	Children []NavigationItem `json:"children,omitempty"`
}

// Form represents a form on the page
type Form struct {
	Action  string      `json:"action,omitempty"`
	Method  string      `json:"method,omitempty"`
	Fields  []FormField `json:"fields,omitempty"`
	Purpose string      `json:"purpose,omitempty"`
}

// FormField represents a form field
type FormField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Label       string `json:"label,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Required    bool   `json:"required"`
}

// Table represents a table on the page
type Table struct {
	Headers []string   `json:"headers,omitempty"`
	Rows    [][]string `json:"rows,omitempty"`
	Purpose string     `json:"purpose,omitempty"`
}

// List represents a list on the page
type List struct {
	Type    string   `json:"type"` // ul, ol, dl
	Items   []string `json:"items,omitempty"`
	Purpose string   `json:"purpose,omitempty"`
}

// AnalyzeWebsiteStructure analyzes the structure of a website
func (h *WebsiteStructureHandler) AnalyzeWebsiteStructure(content *ParsedContent) (*WebsiteStructure, error) {
	structure := &WebsiteStructure{
		ContentAreas: []ContentArea{},
		Navigation:   []NavigationItem{},
		Forms:        []Form{},
		Tables:       []Table{},
		Lists:        []List{},
		Metadata:     make(map[string]string),
	}

	// Detect website type
	structure.Type = h.detectWebsiteType(content)

	// Detect framework
	structure.Framework = h.detectFramework(content)

	// Detect template/layout
	structure.Template = h.detectTemplate(content)

	// Analyze content areas
	structure.ContentAreas = h.analyzeContentAreas(content)

	// Analyze navigation
	structure.Navigation = h.analyzeNavigation(content)

	// Analyze forms
	structure.Forms = h.analyzeForms(content)

	// Analyze tables
	structure.Tables = h.analyzeTables(content)

	// Analyze lists
	structure.Lists = h.analyzeLists(content)

	// Calculate confidence
	structure.Confidence = h.calculateStructureConfidence(structure)

	return structure, nil
}

// detectWebsiteType detects the type of website
func (h *WebsiteStructureHandler) detectWebsiteType(content *ParsedContent) string {
	text := strings.ToLower(content.Text)

	// E-commerce indicators
	ecommercePatterns := []string{
		"add to cart", "shopping cart", "checkout", "buy now", "add to basket",
		"product", "price", "sale", "discount", "shipping", "payment",
		"order", "purchase", "shop", "store", "buy", "sell",
	}

	for _, pattern := range ecommercePatterns {
		if strings.Contains(text, pattern) {
			return "e-commerce"
		}
	}

	// Corporate/Business indicators
	corporatePatterns := []string{
		"about us", "our company", "corporate", "business", "enterprise",
		"services", "solutions", "consulting", "professional", "expertise",
		"team", "leadership", "careers", "contact us", "get in touch",
	}

	for _, pattern := range corporatePatterns {
		if strings.Contains(text, pattern) {
			return "corporate"
		}
	}

	// Blog/News indicators
	blogPatterns := []string{
		"blog", "news", "article", "post", "read more", "published",
		"author", "date", "category", "tags", "comments", "subscribe",
	}

	for _, pattern := range blogPatterns {
		if strings.Contains(text, pattern) {
			return "blog/news"
		}
	}

	// Portfolio indicators
	portfolioPatterns := []string{
		"portfolio", "work", "projects", "case studies", "showcase",
		"gallery", "examples", "samples", "previous work",
	}

	for _, pattern := range portfolioPatterns {
		if strings.Contains(text, pattern) {
			return "portfolio"
		}
	}

	// Landing page indicators
	landingPatterns := []string{
		"get started", "sign up", "free trial", "download", "learn more",
		"hero", "call to action", "cta", "conversion", "lead generation",
	}

	for _, pattern := range landingPatterns {
		if strings.Contains(text, pattern) {
			return "landing_page"
		}
	}

	return "general"
}

// detectFramework detects the web framework used
func (h *WebsiteStructureHandler) detectFramework(content *ParsedContent) string {
	html := strings.ToLower(content.HTML)

	// React indicators
	if strings.Contains(html, "react") || strings.Contains(html, "data-reactroot") {
		return "React"
	}

	// Vue indicators
	if strings.Contains(html, "vue") || strings.Contains(html, "v-app") {
		return "Vue.js"
	}

	// Angular indicators
	if strings.Contains(html, "angular") || strings.Contains(html, "ng-app") {
		return "Angular"
	}

	// WordPress indicators
	if strings.Contains(html, "wp-content") || strings.Contains(html, "wordpress") {
		return "WordPress"
	}

	// Shopify indicators
	if strings.Contains(html, "shopify") || strings.Contains(html, "cart.js") {
		return "Shopify"
	}

	// Wix indicators
	if strings.Contains(html, "wix") || strings.Contains(html, "wixsite") {
		return "Wix"
	}

	// Squarespace indicators
	if strings.Contains(html, "squarespace") {
		return "Squarespace"
	}

	// Bootstrap indicators
	if strings.Contains(html, "bootstrap") || strings.Contains(html, "bs-") {
		return "Bootstrap"
	}

	// Foundation indicators
	if strings.Contains(html, "foundation") || strings.Contains(html, "foundation-") {
		return "Foundation"
	}

	return "Unknown"
}

// detectTemplate detects the template/layout used
func (h *WebsiteStructureHandler) detectTemplate(content *ParsedContent) string {
	html := strings.ToLower(content.HTML)

	// Common template patterns
	templates := map[string][]string{
		"single_page": {"single-page", "spa", "single page"},
		"multi_page":  {"multi-page", "traditional", "static"},
		"blog":        {"blog", "article", "post"},
		"ecommerce":   {"shop", "store", "product"},
		"landing":     {"landing", "hero", "conversion"},
	}

	for template, patterns := range templates {
		for _, pattern := range patterns {
			if strings.Contains(html, pattern) {
				return template
			}
		}
	}

	return "custom"
}

// analyzeContentAreas analyzes content areas on the page
func (h *WebsiteStructureHandler) analyzeContentAreas(content *ParsedContent) []ContentArea {
	var areas []ContentArea

	// Look for header content
	if headerContent := h.extractHeaderContent(content); headerContent != "" {
		areas = append(areas, ContentArea{
			Type:       "header",
			Content:    headerContent,
			Importance: 8,
		})
	}

	// Look for main content
	if mainContent := h.extractMainContent(content); mainContent != "" {
		areas = append(areas, ContentArea{
			Type:       "main",
			Content:    mainContent,
			Importance: 10,
		})
	}

	// Look for footer content
	if footerContent := h.extractFooterContent(content); footerContent != "" {
		areas = append(areas, ContentArea{
			Type:       "footer",
			Content:    footerContent,
			Importance: 6,
		})
	}

	// Look for sidebar content
	if sidebarContent := h.extractSidebarContent(content); sidebarContent != "" {
		areas = append(areas, ContentArea{
			Type:       "sidebar",
			Content:    sidebarContent,
			Importance: 5,
		})
	}

	return areas
}

// analyzeNavigation analyzes navigation structure
func (h *WebsiteStructureHandler) analyzeNavigation(content *ParsedContent) []NavigationItem {
	var navigation []NavigationItem

	// Extract navigation from links
	for _, link := range content.Links {
		// Look for common navigation patterns
		if h.isNavigationLink(link) {
			text := h.extractLinkText(link, content)
			if text != "" {
				navigation = append(navigation, NavigationItem{
					Text: text,
					URL:  link,
				})
			}
		}
	}

	return navigation
}

// analyzeForms analyzes forms on the page
func (h *WebsiteStructureHandler) analyzeForms(content *ParsedContent) []Form {
	var forms []Form

	// Look for form patterns in HTML
	formPatterns := []string{
		`<form[^>]*action="([^"]*)"[^>]*method="([^"]*)"[^>]*>`,
		`<input[^>]*name="([^"]*)"[^>]*type="([^"]*)"[^>]*>`,
		`<textarea[^>]*name="([^"]*)"[^>]*>`,
		`<select[^>]*name="([^"]*)"[^>]*>`,
	}

	for _, pattern := range formPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content.HTML, -1)
		for _, match := range matches {
			if len(match) > 1 {
				form := Form{
					Action: match[1],
					Method: match[2],
				}
				forms = append(forms, form)
			}
		}
	}

	return forms
}

// analyzeTables analyzes tables on the page
func (h *WebsiteStructureHandler) analyzeTables(content *ParsedContent) []Table {
	var tables []Table

	// Look for table patterns
	tablePattern := regexp.MustCompile(`<table[^>]*>(.*?)</table>`)
	matches := tablePattern.FindAllStringSubmatch(content.HTML, -1)

	for _, match := range matches {
		if len(match) > 1 {
			table := h.parseTable(match[1])
			tables = append(tables, table)
		}
	}

	return tables
}

// analyzeLists analyzes lists on the page
func (h *WebsiteStructureHandler) analyzeLists(content *ParsedContent) []List {
	var lists []List

	// Look for list patterns
	listPatterns := map[string]string{
		"ul": `<ul[^>]*>(.*?)</ul>`,
		"ol": `<ol[^>]*>(.*?)</ol>`,
		"dl": `<dl[^>]*>(.*?)</dl>`,
	}

	for listType, pattern := range listPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content.HTML, -1)

		for _, match := range matches {
			if len(match) > 1 {
				list := h.parseList(match[1], listType)
				lists = append(lists, list)
			}
		}
	}

	return lists
}

// Helper methods for content extraction
func (h *WebsiteStructureHandler) extractHeaderContent(content *ParsedContent) string {
	// Look for header patterns
	headerPatterns := []string{
		`(?i)(?:header|navigation|nav|menu)[^.]*`,
		`(?i)(?:logo|brand|company name)[^.]*`,
	}

	for _, pattern := range headerPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindString(content.Text)
		if matches != "" {
			return matches
		}
	}

	return ""
}

func (h *WebsiteStructureHandler) extractMainContent(content *ParsedContent) string {
	// Look for main content patterns
	mainPatterns := []string{
		`(?i)(?:main|content|body|article)[^.]*`,
		`(?i)(?:about|services|products|solutions)[^.]*`,
	}

	for _, pattern := range mainPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindString(content.Text)
		if matches != "" {
			return matches
		}
	}

	// Fallback to first 500 characters
	if len(content.Text) > 500 {
		return content.Text[:500]
	}

	return content.Text
}

func (h *WebsiteStructureHandler) extractFooterContent(content *ParsedContent) string {
	// Look for footer patterns
	footerPatterns := []string{
		`(?i)(?:footer|contact|address|phone|email)[^.]*`,
		`(?i)(?:copyright|all rights reserved)[^.]*`,
	}

	for _, pattern := range footerPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindString(content.Text)
		if matches != "" {
			return matches
		}
	}

	return ""
}

func (h *WebsiteStructureHandler) extractSidebarContent(content *ParsedContent) string {
	// Look for sidebar patterns
	sidebarPatterns := []string{
		`(?i)(?:sidebar|side panel|side menu)[^.]*`,
		`(?i)(?:related|popular|featured)[^.]*`,
	}

	for _, pattern := range sidebarPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindString(content.Text)
		if matches != "" {
			return matches
		}
	}

	return ""
}

func (h *WebsiteStructureHandler) isNavigationLink(link string) bool {
	// Common navigation link patterns
	navPatterns := []string{
		"about", "services", "products", "contact", "home", "blog", "news",
		"company", "team", "careers", "support", "help", "faq",
	}

	linkLower := strings.ToLower(link)
	for _, pattern := range navPatterns {
		if strings.Contains(linkLower, pattern) {
			return true
		}
	}

	return false
}

func (h *WebsiteStructureHandler) extractLinkText(link string, content *ParsedContent) string {
	// Look for link text in HTML
	linkPattern := regexp.MustCompile(fmt.Sprintf(`<a[^>]*href="%s"[^>]*>(.*?)</a>`, regexp.QuoteMeta(link)))
	matches := linkPattern.FindStringSubmatch(content.HTML)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

func (h *WebsiteStructureHandler) parseTable(tableHTML string) Table {
	table := Table{}

	// Extract headers
	headerPattern := regexp.MustCompile(`<th[^>]*>(.*?)</th>`)
	headers := headerPattern.FindAllStringSubmatch(tableHTML, -1)
	for _, header := range headers {
		if len(header) > 1 {
			table.Headers = append(table.Headers, strings.TrimSpace(header[1]))
		}
	}

	// Extract rows
	rowPattern := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
	rows := rowPattern.FindAllStringSubmatch(tableHTML, -1)
	for _, row := range rows {
		if len(row) > 1 {
			cellPattern := regexp.MustCompile(`<td[^>]*>(.*?)</td>`)
			cells := cellPattern.FindAllStringSubmatch(row[1], -1)
			var rowData []string
			for _, cell := range cells {
				if len(cell) > 1 {
					rowData = append(rowData, strings.TrimSpace(cell[1]))
				}
			}
			if len(rowData) > 0 {
				table.Rows = append(table.Rows, rowData)
			}
		}
	}

	return table
}

func (h *WebsiteStructureHandler) parseList(listHTML string, listType string) List {
	list := List{Type: listType}

	// Extract list items
	itemPattern := regexp.MustCompile(`<li[^>]*>(.*?)</li>`)
	items := itemPattern.FindAllStringSubmatch(listHTML, -1)
	for _, item := range items {
		if len(item) > 1 {
			list.Items = append(list.Items, strings.TrimSpace(item[1]))
		}
	}

	return list
}

// calculateStructureConfidence calculates confidence score for structure analysis
func (h *WebsiteStructureHandler) calculateStructureConfidence(structure *WebsiteStructure) float64 {
	score := 0.0
	total := 0.0

	// Website type detection (20% weight)
	if structure.Type != "general" {
		score += 20.0
	}
	total += 20.0

	// Framework detection (15% weight)
	if structure.Framework != "Unknown" {
		score += 15.0
	}
	total += 15.0

	// Content areas (25% weight)
	if len(structure.ContentAreas) > 0 {
		score += 25.0
	}
	total += 25.0

	// Navigation (15% weight)
	if len(structure.Navigation) > 0 {
		score += 15.0
	}
	total += 15.0

	// Forms/Tables/Lists (15% weight)
	if len(structure.Forms) > 0 || len(structure.Tables) > 0 || len(structure.Lists) > 0 {
		score += 15.0
	}
	total += 15.0

	// Template detection (10% weight)
	if structure.Template != "custom" {
		score += 10.0
	}
	total += 10.0

	if total == 0 {
		return 0.0
	}

	return (score / total) * 100.0
}
