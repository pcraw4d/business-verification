package webanalysis

import (
	"fmt"
	"time"
)

// HybridScraper combines direct scraping with proxy fallback
type HybridScraper struct {
	simpleScraper *SimpleScraper
	proxyScraper  *WebScraper
	config        HybridScraperConfig
}

// HybridScraperConfig holds configuration for hybrid scraping
type HybridScraperConfig struct {
	DirectScrapingTimeout time.Duration
	ProxyScrapingTimeout  time.Duration
	MaxRetries            int
	EnableProxyFallback   bool
	RateLimitDelay        time.Duration
}

// NewHybridScraper creates a new hybrid scraper
func NewHybridScraper() *HybridScraper {
	return &HybridScraper{
		simpleScraper: NewSimpleScraper(),
		proxyScraper:  NewWebScraper(NewProxyManager()), // Will be nil if no proxies
		config: HybridScraperConfig{
			DirectScrapingTimeout: 15 * time.Second,
			ProxyScrapingTimeout:  30 * time.Second,
			MaxRetries:            3,
			EnableProxyFallback:   true,
			RateLimitDelay:        2 * time.Second,
		},
	}
}

// HybridScrapingJob represents a hybrid scraping request
type HybridScrapingJob struct {
	URL          string
	BusinessName string
	Priority     string // "high", "medium", "low"
	UseProxies   bool   // Force proxy usage
}

// HybridScrapingResult represents the result of hybrid scraping
type HybridScrapingResult struct {
	URL            string
	Title          string
	Text           string
	HTML           string
	StatusCode     int
	Method         string // "direct", "proxy", "failed"
	Error          string
	ScrapedAt      time.Time
	ProcessingTime time.Duration
}

// ScrapeWebsite performs hybrid website scraping
func (hs *HybridScraper) ScrapeWebsite(job *HybridScrapingJob) (*HybridScrapingResult, error) {
	start := time.Now()

	// Strategy 1: Try direct scraping first (faster, cheaper)
	if !job.UseProxies {
		result, err := hs.tryDirectScraping(job)
		if err == nil && result.StatusCode == 200 {
			result.Method = "direct"
			result.ProcessingTime = time.Since(start)
			return result, nil
		}
	}

	// Strategy 2: Fall back to proxy scraping if available
	if hs.config.EnableProxyFallback && hs.proxyScraper != nil {
		result, err := hs.tryProxyScraping(job)
		if err == nil && result.StatusCode == 200 {
			result.Method = "proxy"
			result.ProcessingTime = time.Since(start)
			return result, nil
		}
	}

	// Strategy 3: Return best available result or error
	return &HybridScrapingResult{
		URL:            job.URL,
		Method:         "failed",
		Error:          "all scraping methods failed",
		ScrapedAt:      time.Now(),
		ProcessingTime: time.Since(start),
	}, fmt.Errorf("all scraping methods failed for %s", job.URL)
}

// tryDirectScraping attempts direct scraping
func (hs *HybridScraper) tryDirectScraping(job *HybridScrapingJob) (*HybridScrapingResult, error) {
	simpleJob := &SimpleScrapingJob{
		URL:     job.URL,
		Timeout: hs.config.DirectScrapingTimeout,
	}

	content, err := hs.simpleScraper.ScrapeWebsite(simpleJob)
	if err != nil {
		return &HybridScrapingResult{
			URL:   job.URL,
			Error: err.Error(),
		}, err
	}

	return &HybridScrapingResult{
		URL:        content.URL,
		Title:      content.Title,
		Text:       content.Text,
		HTML:       content.HTML,
		StatusCode: content.StatusCode,
		ScrapedAt:  content.ScrapedAt,
	}, nil
}

// tryProxyScraping attempts proxy scraping
func (hs *HybridScraper) tryProxyScraping(job *HybridScrapingJob) (*HybridScrapingResult, error) {
	proxyJob := &ScrapingJob{
		URL:        job.URL,
		Timeout:    hs.config.ProxyScrapingTimeout,
		MaxRetries: hs.config.MaxRetries,
	}

	content, err := hs.proxyScraper.ScrapeWebsite(proxyJob)
	if err != nil {
		return &HybridScrapingResult{
			URL:   job.URL,
			Error: err.Error(),
		}, err
	}

	return &HybridScrapingResult{
		URL:        content.URL,
		Title:      content.Title,
		Text:       content.Text,
		HTML:       content.HTML,
		StatusCode: 200, // Assume success if no error
		ScrapedAt:  time.Now(),
	}, nil
}

// GetScrapingStats returns statistics about scraping performance
func (hs *HybridScraper) GetScrapingStats() map[string]interface{} {
	return map[string]interface{}{
		"direct_success_rate":   0.0, // TODO: Implement tracking
		"proxy_success_rate":    0.0, // TODO: Implement tracking
		"average_response_time": 0.0, // TODO: Implement tracking
		"total_requests":        0,   // TODO: Implement tracking
	}
}
