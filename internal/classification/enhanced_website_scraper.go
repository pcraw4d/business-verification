package classification

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// EnhancedWebsiteScraper provides advanced website scraping capabilities
type EnhancedWebsiteScraper struct {
	logger *log.Logger
	client *http.Client
}

// ScrapingResult represents the result of a website scraping operation
type ScrapingResult struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	Content       string            `json:"content"`
	TextContent   string            `json:"text_content"`
	Keywords      []string          `json:"keywords"`
	ContentType   string            `json:"content_type"`
	ContentLength int64             `json:"content_length"`
	Headers       map[string]string `json:"headers"`
	FinalURL      string            `json:"final_url"`
	ScrapedAt     time.Time         `json:"scraped_at"`
	Duration      time.Duration     `json:"duration"`
	Error         string            `json:"error,omitempty"`
	Success       bool              `json:"success"`
}

// NewEnhancedWebsiteScraper creates a new enhanced website scraper
func NewEnhancedWebsiteScraper(logger *log.Logger) *EnhancedWebsiteScraper {
	return &EnhancedWebsiteScraper{
		logger: logger,
		client: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 2,
			},
		},
	}
}

// ScrapeWebsite performs enhanced website scraping with comprehensive error handling
func (ews *EnhancedWebsiteScraper) ScrapeWebsite(ctx context.Context, websiteURL string) *ScrapingResult {
	startTime := time.Now()
	result := &ScrapingResult{
		URL:       websiteURL,
		ScrapedAt: time.Now(),
		Success:   false,
	}

	ews.logger.Printf("🌐 [Enhanced] Starting enhanced website scraping for: %s", websiteURL)

	// Validate and normalize URL
	normalizedURL, err := ews.normalizeURL(websiteURL)
	if err != nil {
		result.Error = fmt.Sprintf("URL validation failed: %v", err)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}
	result.URL = normalizedURL

	// Create request with enhanced headers
	req, err := ews.createRequest(ctx, normalizedURL)
	if err != nil {
		result.Error = fmt.Sprintf("Request creation failed: %v", err)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}

	// Make HTTP request
	resp, err := ews.client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}
	defer resp.Body.Close()

	// Process response
	result.StatusCode = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")
	result.FinalURL = resp.Request.URL.String()

	// Extract headers
	result.Headers = make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
		}
	}

	ews.logger.Printf("📊 [Enhanced] Response received - Status: %d, Content-Type: %s, Final URL: %s", 
		result.StatusCode, result.ContentType, result.FinalURL)

	// Check for errors
	if resp.StatusCode >= 400 {
		result.Error = fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}

	// Read response body
	body, err := ews.readResponseBody(resp)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response body: %v", err)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}

	result.Content = string(body)
	result.ContentLength = int64(len(body))

	// Extract text content
	result.TextContent = ews.extractTextFromHTML(result.Content)
	ews.logger.Printf("🧹 [Enhanced] Extracted %d characters of text content", len(result.TextContent))

	// Extract keywords
	result.Keywords = ews.extractBusinessKeywords(result.TextContent)
	ews.logger.Printf("🔍 [Enhanced] Extracted %d keywords: %v", len(result.Keywords), result.Keywords)

	result.Duration = time.Since(startTime)
	result.Success = true

	ews.logger.Printf("✅ [Enhanced] Website scraping completed for %s in %v", normalizedURL, result.Duration)
	return result
}

// normalizeURL validates and normalizes the URL
func (ews *EnhancedWebsiteScraper) normalizeURL(websiteURL string) (string, error) {
	// Add scheme if missing
	if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
		websiteURL = "https://" + websiteURL
		ews.logger.Printf("🔧 [Enhanced] Added HTTPS scheme: %s", websiteURL)
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Ensure we have a valid host
	if parsedURL.Host == "" {
		return "", fmt.Errorf("missing host in URL")
	}

	return parsedURL.String(), nil
}

// createRequest creates an HTTP request with enhanced headers
func (ews *EnhancedWebsiteScraper) createRequest(ctx context.Context, websiteURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
	if err != nil {
		return nil, err
	}

	// Set comprehensive headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Sec-Ch-Ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)

	return req, nil
}

// readResponseBody reads the response body with size limiting
func (ews *EnhancedWebsiteScraper) readResponseBody(resp *http.Response) ([]byte, error) {
	maxSize := int64(10 * 1024 * 1024) // 10MB limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		return nil, err
	}

	ews.logger.Printf("📄 [Enhanced] Read %d bytes from response", len(body))
	return body, nil
}

// extractTextFromHTML extracts clean text content from HTML
func (ews *EnhancedWebsiteScraper) extractTextFromHTML(htmlContent string) string {
	// Remove script and style tags completely
	htmlContent = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?i)<noscript[^>]*>.*?</noscript>`).ReplaceAllString(htmlContent, "")
	
	// Remove HTML comments
	htmlContent = regexp.MustCompile(`<!--.*?-->`).ReplaceAllString(htmlContent, "")
	
	// Remove HTML tags
	htmlContent = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlContent, " ")
	
	// Decode HTML entities (basic ones)
	htmlContent = strings.ReplaceAll(htmlContent, "&amp;", "&")
	htmlContent = strings.ReplaceAll(htmlContent, "&lt;", "<")
	htmlContent = strings.ReplaceAll(htmlContent, "&gt;", ">")
	htmlContent = strings.ReplaceAll(htmlContent, "&quot;", "\"")
	htmlContent = strings.ReplaceAll(htmlContent, "&#39;", "'")
	htmlContent = strings.ReplaceAll(htmlContent, "&nbsp;", " ")
	
	// Clean up whitespace
	htmlContent = regexp.MustCompile(`\s+`).ReplaceAllString(htmlContent, " ")
	
	return strings.TrimSpace(htmlContent)
}

// extractBusinessKeywords extracts business-relevant keywords from text content
func (ews *EnhancedWebsiteScraper) extractBusinessKeywords(textContent string) []string {
	var keywords []string
	
	// Convert to lowercase for processing
	text := strings.ToLower(textContent)
	
	// Enhanced business-relevant keyword patterns
	businessPatterns := []string{
		// Food & Beverage
		`\b(restaurant|cafe|coffee|food|dining|kitchen|catering|bakery|bar|pub|brewery|winery|wine|beer|cocktail|menu|chef|cook|cuisine|delivery|takeout)\b`,
		// Technology
		`\b(technology|software|tech|app|digital|web|mobile|cloud|ai|ml|data|cyber|security|programming|development|IT|computer|internet|online|platform|api|database)\b`,
		// Healthcare
		`\b(healthcare|medical|clinic|hospital|doctor|dentist|therapy|wellness|pharmacy|medicine|patient|treatment|health|care|nurse|physician)\b`,
		// Legal
		`\b(legal|law|attorney|lawyer|court|litigation|patent|trademark|copyright|legal services|advocacy|justice|legal advice|law firm)\b`,
		// Retail & E-commerce
		`\b(retail|store|shop|ecommerce|online|fashion|clothing|electronics|beauty|products|merchandise|selling|commerce|shopping|marketplace)\b`,
		// Finance
		`\b(finance|banking|investment|insurance|accounting|tax|financial|credit|loan|money|capital|funding|payment|transaction|wealth)\b`,
		// Real Estate
		`\b(real estate|property|construction|building|architecture|design|interior|home|house|apartment|rental|mortgage|property management)\b`,
		// Education
		`\b(education|school|university|training|learning|course|academy|institute|student|teacher|teaching|academic|degree|certification)\b`,
		// Consulting & Professional Services
		`\b(consulting|advisory|strategy|management|business|corporate|professional|services|expert|specialist|consultant|advisory)\b`,
		// Manufacturing
		`\b(manufacturing|production|factory|industrial|automotive|machinery|equipment|assembly|production|industrial|machinery)\b`,
		// Transportation & Logistics
		`\b(transportation|logistics|shipping|delivery|freight|warehouse|supply chain|trucking|logistics|shipping|delivery)\b`,
		// Entertainment & Media
		`\b(entertainment|media|marketing|advertising|design|creative|art|music|film|television|broadcasting|publishing|content)\b`,
		// Energy & Utilities
		`\b(energy|utilities|renewable|solar|wind|oil|gas|power|electricity|energy|utilities|renewable|solar|wind)\b`,
		// Agriculture
		`\b(agriculture|farming|food production|crop|livestock|organic|sustainable|agriculture|farming|food production|crop)\b`,
		// Travel & Hospitality
		`\b(travel|tourism|hospitality|hotel|accommodation|vacation|booking|trip|travel|tourism|hospitality|hotel|accommodation)\b`,
	}
	
	// Extract keywords using patterns
	for _, pattern := range businessPatterns {
		matches := regexp.MustCompile(pattern).FindAllString(text, -1)
		for _, match := range matches {
			// Remove duplicates and add to keywords
			if !ews.containsKeyword(keywords, match) {
				keywords = append(keywords, match)
			}
		}
	}
	
	// Also extract common business words
	commonBusinessWords := []string{
		"service", "services", "company", "business", "corp", "corporation", "inc", "llc", "ltd",
		"enterprise", "solutions", "systems", "group", "associates", "partners", "consulting",
		"management", "development", "production", "distribution", "marketing", "sales",
		"customer", "clients", "professional", "expert", "specialist", "quality", "premium",
		"innovative", "leading", "trusted", "reliable", "experienced", "established",
		"support", "help", "contact", "about", "team", "staff", "employees", "office",
		"location", "address", "phone", "email", "website", "online", "digital",
	}
	
	for _, word := range commonBusinessWords {
		if strings.Contains(text, word) && !ews.containsKeyword(keywords, word) {
			keywords = append(keywords, word)
		}
	}
	
	// Limit to top 25 keywords to avoid noise
	if len(keywords) > 25 {
		keywords = keywords[:25]
	}
	
	return keywords
}

// containsKeyword checks if a keyword already exists in the slice
func (ews *EnhancedWebsiteScraper) containsKeyword(keywords []string, keyword string) bool {
	for _, k := range keywords {
		if k == keyword {
			return true
		}
	}
	return false
}
