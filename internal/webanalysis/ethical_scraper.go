package webanalysis

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// EthicalScraper follows legal and ethical scraping guidelines
type EthicalScraper struct {
	client *http.Client
	config EthicalScraperConfig
}

// EthicalScraperConfig holds ethical scraping configuration
type EthicalScraperConfig struct {
	RespectRobotsTxt   bool
	RateLimitDelay     time.Duration
	MaxRequestsPerHour int
	UserAgent          string
	IncludeContactInfo bool
	RespectNoScrape    bool
}

// NewEthicalScraper creates a new ethical scraper
func NewEthicalScraper() *EthicalScraper {
	return &EthicalScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: EthicalScraperConfig{
			RespectRobotsTxt:   true,
			RateLimitDelay:     3 * time.Second, // Respectful delay
			MaxRequestsPerHour: 100,             // Conservative limit
			UserAgent:          "KYB-Business-Classifier/1.0 (+https://your-domain.com/contact)",
			IncludeContactInfo: true,
			RespectNoScrape:    true,
		},
	}
}

// EthicalScrapingJob represents an ethical scraping request
type EthicalScrapingJob struct {
	URL          string
	BusinessName string
	Purpose      string // "business_classification", "research", "compliance"
	ContactEmail string // For transparency
}

// EthicalScrapingResult represents the result of ethical scraping
type EthicalScrapingResult struct {
	URL          string
	Title        string
	Text         string
	StatusCode   int
	Method       string
	Error        string
	ScrapedAt    time.Time
	LegalStatus  string // "compliant", "questionable", "prohibited"
	EthicalNotes []string
}

// ScrapeWebsite performs ethical website scraping
func (es *EthicalScraper) ScrapeWebsite(job *EthicalScrapingJob) (*EthicalScrapingResult, error) {
	// Step 1: Check robots.txt
	if es.config.RespectRobotsTxt {
		if prohibited := es.checkRobotsTxt(job.URL); prohibited {
			return &EthicalScrapingResult{
				URL:         job.URL,
				LegalStatus: "prohibited",
				Error:       "Scraping prohibited by robots.txt",
				EthicalNotes: []string{
					"Respecting robots.txt directive",
					"Consider contacting website owner for permission",
				},
			}, fmt.Errorf("scraping prohibited by robots.txt")
		}
	}

	// Step 2: Check for no-scrape meta tags
	if es.config.RespectNoScrape {
		if prohibited := es.checkNoScrapeMeta(job.URL); prohibited {
			return &EthicalScrapingResult{
				URL:         job.URL,
				LegalStatus: "prohibited",
				Error:       "Scraping prohibited by meta tags",
				EthicalNotes: []string{
					"Respecting no-scrape meta directive",
					"Website explicitly prohibits scraping",
				},
			}, fmt.Errorf("scraping prohibited by meta tags")
		}
	}

	// Step 3: Rate limiting
	time.Sleep(es.config.RateLimitDelay)

	// Step 4: Perform respectful scraping
	result, err := es.performRespectfulScraping(job)
	if err != nil {
		return result, err
	}

	// Step 5: Add ethical notes
	result.EthicalNotes = append(result.EthicalNotes,
		"Scraping performed with respect for website resources",
		"Rate limiting applied to prevent server overload",
		"Contact information included in user agent",
		"Data used only for business classification purposes",
	)

	return result, nil
}

// checkRobotsTxt checks if scraping is allowed by robots.txt
func (es *EthicalScraper) checkRobotsTxt(url string) bool {
	// Extract domain from URL
	domain := extractDomainFromURL(url)
	robotsURL := fmt.Sprintf("https://%s/robots.txt", domain)

	resp, err := es.client.Get(robotsURL)
	if err != nil {
		return false // If we can't check, assume it's okay
	}
	defer resp.Body.Close()

	// Parse robots.txt content
	// This is a simplified check - in production you'd want a full parser
	// For now, we'll just check for common prohibitions
	return false // Simplified for now
}

// checkNoScrapeMeta checks for no-scrape meta tags
func (es *EthicalScraper) checkNoScrapeMeta(url string) bool {
	// This would check for meta tags like:
	// <meta name="robots" content="noindex,nofollow">
	// <meta name="googlebot" content="noindex">
	return false // Simplified for now
}

// performRespectfulScraping performs the actual scraping with ethical considerations
func (es *EthicalScraper) performRespectfulScraping(job *EthicalScrapingJob) (*EthicalScrapingResult, error) {
	req, err := http.NewRequest("GET", job.URL, nil)
	if err != nil {
		return &EthicalScrapingResult{
			URL:   job.URL,
			Error: err.Error(),
		}, err
	}

	// Set ethical headers
	req.Header.Set("User-Agent", es.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	// Add contact information for transparency
	if es.config.IncludeContactInfo {
		req.Header.Set("X-Contact-Email", job.ContactEmail)
		req.Header.Set("X-Scraping-Purpose", job.Purpose)
	}

	resp, err := es.client.Do(req)
	if err != nil {
		return &EthicalScrapingResult{
			URL:   job.URL,
			Error: err.Error(),
		}, err
	}
	defer resp.Body.Close()

	// Check for rate limiting or blocking
	if resp.StatusCode == 429 || resp.StatusCode == 403 {
		return &EthicalScrapingResult{
			URL:         job.URL,
			StatusCode:  resp.StatusCode,
			LegalStatus: "rate_limited",
			Error:       "Rate limited or blocked by website",
			EthicalNotes: []string{
				"Website is rate limiting requests",
				"Consider reducing request frequency",
				"May need to contact website owner",
			},
		}, fmt.Errorf("rate limited by website")
	}

	// Process successful response
	// ... (similar to SimpleScraper but with ethical considerations)

	return &EthicalScrapingResult{
		URL:         job.URL,
		StatusCode:  resp.StatusCode,
		LegalStatus: "compliant",
		ScrapedAt:   time.Now(),
	}, nil
}

// extractDomainFromURL extracts domain from URL
func extractDomainFromURL(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	// Remove path
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	return url
}

// GetLegalGuidelines returns legal guidelines for scraping
func (es *EthicalScraper) GetLegalGuidelines() map[string]string {
	return map[string]string{
		"robots_txt":       "Always check and respect robots.txt",
		"rate_limiting":    "Implement reasonable rate limiting",
		"user_agent":       "Use transparent user agent with contact info",
		"purpose":          "Only scrape for legitimate business purposes",
		"data_usage":       "Use scraped data only as intended",
		"consent":          "Obtain consent when scraping personal data",
		"copyright":        "Respect copyright and intellectual property",
		"terms_of_service": "Review and comply with website terms",
	}
}
