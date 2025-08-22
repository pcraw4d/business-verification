package external

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// HTMLParser handles HTML parsing and text extraction
type HTMLParser struct {
	// Configuration for parsing
	removeScripts    bool
	removeStyles     bool
	removeComments   bool
	preserveLinks     bool
	maxTextLength     int
	extractMetaData   bool
	extractStructured bool
}

// NewHTMLParser creates a new HTML parser with default configuration
func NewHTMLParser() *HTMLParser {
	return &HTMLParser{
		removeScripts:    true,
		removeStyles:     true,
		removeComments:   true,
		preserveLinks:     true,
		maxTextLength:     100000, // 100KB limit
		extractMetaData:   true,
		extractStructured: true,
	}
}

// ParsedContent represents the parsed HTML content
type ParsedContent struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords"`
	Text        string            `json:"text"`
	HTML        string            `json:"html"`
	Links       []string          `json:"links"`
	Images      []string          `json:"images"`
	Metadata    map[string]string `json:"metadata"`
	Structured  *StructuredData   `json:"structured,omitempty"`
	Language    string            `json:"language"`
	Encoding    string            `json:"encoding"`
}

// StructuredData represents structured data extracted from the page
type StructuredData struct {
	BusinessName    string   `json:"business_name,omitempty"`
	Address         string   `json:"address,omitempty"`
	Phone           string   `json:"phone,omitempty"`
	Email           string   `json:"email,omitempty"`
	Website         string   `json:"website,omitempty"`
	SocialMedia     []string `json:"social_media,omitempty"`
	BusinessHours   string   `json:"business_hours,omitempty"`
	Services        []string `json:"services,omitempty"`
	TeamMembers     []string `json:"team_members,omitempty"`
	ContactInfo     []string `json:"contact_info,omitempty"`
	LocationInfo    []string `json:"location_info,omitempty"`
}

// ParseHTML parses HTML content and extracts text and metadata
func (p *HTMLParser) ParseHTML(content string) (*ParsedContent, error) {
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := &ParsedContent{
		Metadata: make(map[string]string),
		HTML:     content, // Store original HTML content
	}

	// Extract content
	p.extractTitle(doc, result)
	if p.extractMetaData {
		p.extractMetadata(doc, result)
	}
	p.extractText(doc, result)
	p.extractLinks(doc, result)
	p.extractImages(doc, result)
	p.extractLanguage(doc, result)
	p.extractEncoding(doc, result)

	// Extract structured data if enabled
	if p.extractMetaData {
		result.Structured = p.extractStructuredData(doc)
	}

	// Clean and normalize text
	result.Text = p.cleanText(result.Text)
	result.Description = p.cleanText(result.Description)

	return result, nil
}

// extractTitle extracts the page title
func (p *HTMLParser) extractTitle(doc *html.Node, result *ParsedContent) {
	var findTitle func(*html.Node)
	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				result.Title = strings.TrimSpace(n.FirstChild.Data)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
		}
	}
	findTitle(doc)
}

// extractMetadata extracts meta tags and other metadata
func (p *HTMLParser) extractMetadata(doc *html.Node, result *ParsedContent) {
	var findMeta func(*html.Node)
	findMeta = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name":
					name = attr.Val
				case "property":
					name = attr.Val
				case "content":
					content = attr.Val
				}
			}

			if name != "" && content != "" {
				switch strings.ToLower(name) {
				case "description":
					result.Description = content
				case "keywords":
					result.Keywords = p.parseKeywords(content)
				default:
					result.Metadata[name] = content
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findMeta(c)
		}
	}
	findMeta(doc)
}

// extractText extracts text content from the HTML
func (p *HTMLParser) extractText(doc *html.Node, result *ParsedContent) {
	var buf bytes.Buffer
	p.extractTextNode(doc, &buf)
	result.Text = buf.String()
}

// extractTextNode recursively extracts text from HTML nodes
func (p *HTMLParser) extractTextNode(n *html.Node, buf *bytes.Buffer) {
	switch n.Type {
	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			buf.WriteString(text)
			buf.WriteString(" ")
		}
	case html.ElementNode:
		// Skip certain elements
		if p.shouldSkipElement(n.Data) {
			return
		}

		// Handle special elements
		switch n.Data {
		case "br", "p", "div", "h1", "h2", "h3", "h4", "h5", "h6":
			buf.WriteString("\n")
		case "li":
			buf.WriteString("• ")
		}

		// Process child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			p.extractTextNode(c, buf)
		}

		// Add line break after block elements
		if p.isBlockElement(n.Data) {
			buf.WriteString("\n")
		}
	}

	// Process siblings
	for s := n.NextSibling; s != nil; s = s.NextSibling {
		p.extractTextNode(s, buf)
	}
}

// extractLinks extracts all links from the HTML
func (p *HTMLParser) extractLinks(doc *html.Node, result *ParsedContent) {
	var findLinks func(*html.Node)
	findLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && attr.Val != "" {
					// Skip javascript: and mailto: links
					if !strings.HasPrefix(attr.Val, "javascript:") && 
					   !strings.HasPrefix(attr.Val, "mailto:") {
						result.Links = append(result.Links, attr.Val)
					}
					break
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLinks(c)
		}
	}
	findLinks(doc)
}

// extractImages extracts all image sources from the HTML
func (p *HTMLParser) extractImages(doc *html.Node, result *ParsedContent) {
	var findImages func(*html.Node)
	findImages = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "src" && attr.Val != "" {
					result.Images = append(result.Images, attr.Val)
					break
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findImages(c)
		}
	}
	findImages(doc)
}

// extractLanguage extracts the language from HTML lang attribute
func (p *HTMLParser) extractLanguage(doc *html.Node, result *ParsedContent) {
	var findLang func(*html.Node)
	findLang = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "html" {
			for _, attr := range n.Attr {
				if attr.Key == "lang" {
					result.Language = attr.Val
					return
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLang(c)
		}
	}
	findLang(doc)
}

// extractEncoding extracts the character encoding from meta tags
func (p *HTMLParser) extractEncoding(doc *html.Node, result *ParsedContent) {
	var findEncoding func(*html.Node)
	findEncoding = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var charset string
			for _, attr := range n.Attr {
				if attr.Key == "charset" {
					charset = attr.Val
					break
				}
				if attr.Key == "http-equiv" && strings.ToLower(attr.Val) == "content-type" {
					for _, contentAttr := range n.Attr {
						if contentAttr.Key == "content" {
							if strings.Contains(contentAttr.Val, "charset=") {
								parts := strings.Split(contentAttr.Val, "charset=")
								if len(parts) > 1 {
									charset = parts[1]
								}
							}
							break
						}
					}
				}
			}
			if charset != "" {
				result.Encoding = charset
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findEncoding(c)
		}
	}
	findEncoding(doc)
}

// extractStructuredData extracts structured business information
func (p *HTMLParser) extractStructuredData(doc *html.Node) *StructuredData {
	data := &StructuredData{}

	// Extract from JSON-LD structured data
	p.extractJSONLD(doc, data)

	// Extract from microdata
	p.extractMicrodata(doc, data)

	// Extract from common patterns
	p.extractCommonPatterns(doc, data)

	return data
}

// extractJSONLD extracts structured data from JSON-LD scripts
func (p *HTMLParser) extractJSONLD(doc *html.Node, data *StructuredData) {
	var findJSONLD func(*html.Node)
	findJSONLD = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			var scriptType string
			for _, attr := range n.Attr {
				if attr.Key == "type" {
					scriptType = attr.Val
					break
				}
			}

			if scriptType == "application/ld+json" && n.FirstChild != nil {
				// Parse JSON-LD content
				content := n.FirstChild.Data
				p.parseJSONLDContent(content, data)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findJSONLD(c)
		}
	}
	findJSONLD(doc)
}

// extractMicrodata extracts microdata attributes
func (p *HTMLParser) extractMicrodata(doc *html.Node, data *StructuredData) {
	var findMicrodata func(*html.Node)
	findMicrodata = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if strings.HasPrefix(attr.Key, "itemprop") {
					value := p.getNodeText(n)
					switch attr.Val {
					case "name":
						if data.BusinessName == "" {
							data.BusinessName = value
						}
					case "address":
						if data.Address == "" {
							data.Address = value
						}
					case "telephone":
						if data.Phone == "" {
							data.Phone = value
						}
					case "email":
						if data.Email == "" {
							data.Email = value
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findMicrodata(c)
		}
	}
	findMicrodata(doc)
}

// extractCommonPatterns extracts business information using common patterns
func (p *HTMLParser) extractCommonPatterns(doc *html.Node, data *StructuredData) {
	// Extract phone numbers
	phonePattern := regexp.MustCompile(`(?:\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})`)
	text := p.getNodeText(doc)
	phones := phonePattern.FindAllString(text, -1)
	if len(phones) > 0 && data.Phone == "" {
		data.Phone = phones[0]
	}

	// Extract email addresses
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailPattern.FindAllString(text, -1)
	if len(emails) > 0 && data.Email == "" {
		data.Email = emails[0]
	}

	// Extract social media links
	socialPattern := regexp.MustCompile(`(?:facebook|twitter|linkedin|instagram)\.com/[a-zA-Z0-9._-]+`)
	socialLinks := socialPattern.FindAllString(text, -1)
	data.SocialMedia = append(data.SocialMedia, socialLinks...)
}

// Helper methods
func (p *HTMLParser) shouldSkipElement(tagName string) bool {
	if p.removeScripts && tagName == "script" {
		return true
	}
	if p.removeStyles && (tagName == "style" || tagName == "link") {
		return true
	}
	if p.removeComments {
		return false // Comments are handled differently
	}
	return false
}

func (p *HTMLParser) isBlockElement(tagName string) bool {
	blockElements := []string{"div", "p", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "li", "blockquote", "pre", "table"}
	for _, elem := range blockElements {
		if tagName == elem {
			return true
		}
	}
	return false
}

func (p *HTMLParser) parseKeywords(keywords string) []string {
	parts := strings.Split(keywords, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (p *HTMLParser) cleanText(text string) string {
	// Remove extra whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Limit text length
	if len(text) > p.maxTextLength {
		text = text[:p.maxTextLength] + "..."
	}

	return text
}

func (p *HTMLParser) getNodeText(n *html.Node) string {
	var buf bytes.Buffer
	p.extractTextNode(n, &buf)
	return buf.String()
}

func (p *HTMLParser) parseJSONLDContent(content string, data *StructuredData) {
	// This is a simplified JSON-LD parser
	// In a production environment, you'd want to use a proper JSON parser
	
	// Look for common business properties
	if strings.Contains(content, `"@type":"Organization"`) || 
	   strings.Contains(content, `"@type":"LocalBusiness"`) {
		
		// Extract name
		if nameMatch := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`).FindStringSubmatch(content); len(nameMatch) > 1 {
			data.BusinessName = nameMatch[1]
		}
		
		// Extract address
		if addrMatch := regexp.MustCompile(`"address"\s*:\s*"([^"]+)"`).FindStringSubmatch(content); len(addrMatch) > 1 {
			data.Address = addrMatch[1]
		}
		
		// Extract phone
		if phoneMatch := regexp.MustCompile(`"telephone"\s*:\s*"([^"]+)"`).FindStringSubmatch(content); len(phoneMatch) > 1 {
			data.Phone = phoneMatch[1]
		}
		
		// Extract email
		if emailMatch := regexp.MustCompile(`"email"\s*:\s*"([^"]+)"`).FindStringSubmatch(content); len(emailMatch) > 1 {
			data.Email = emailMatch[1]
		}
	}
}
