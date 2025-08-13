package webanalysis

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// SimpleScraper provides basic website scraping without proxies
type SimpleScraper struct {
	client *http.Client
}

// NewSimpleScraper creates a new simple scraper
func NewSimpleScraper() *SimpleScraper {
	return &SimpleScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SimpleScrapingJob represents a simple scraping request
type SimpleScrapingJob struct {
	URL     string
	Timeout time.Duration
}

// SimpleScrapedContent represents scraped content
type SimpleScrapedContent struct {
	URL        string
	Title      string
	Text       string
	HTML       string
	StatusCode int
	Error      string
	ScrapedAt  time.Time
}

// ScrapeWebsite performs simple website scraping
func (ss *SimpleScraper) ScrapeWebsite(job *SimpleScrapingJob) (*SimpleScrapedContent, error) {
	// Create request with realistic headers
	req, err := http.NewRequest("GET", job.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set realistic headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Perform request
	resp, err := ss.client.Do(req)
	if err != nil {
		return &SimpleScrapedContent{
			URL:       job.URL,
			Error:     err.Error(),
			ScrapedAt: time.Now(),
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &SimpleScrapedContent{
			URL:       job.URL,
			Error:     err.Error(),
			ScrapedAt: time.Now(),
		}, err
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return &SimpleScrapedContent{
			URL:       job.URL,
			Error:     err.Error(),
			ScrapedAt: time.Now(),
		}, err
	}

	// Extract title and text
	title := extractTitle(doc)
	text := extractText(doc)

	return &SimpleScrapedContent{
		URL:        job.URL,
		Title:      title,
		Text:       text,
		HTML:       string(body),
		StatusCode: resp.StatusCode,
		ScrapedAt:  time.Now(),
	}, nil
}

// extractTitle extracts the page title
func extractTitle(doc *html.Node) string {
	var title string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				title = n.FirstChild.Data
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return title
}

// extractText extracts readable text from HTML
func extractText(doc *html.Node) string {
	var text strings.Builder
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return strings.TrimSpace(text.String())
}
