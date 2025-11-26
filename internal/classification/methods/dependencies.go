package methods

import (
	"context"
)

// WebsiteScraper defines the interface for website scraping functionality
// This interface allows methods package to use website scraping without importing classification package
type WebsiteScraper interface {
	ScrapeWebsite(ctx context.Context, websiteURL string) *ScrapingResult
}

// CodeGenerator defines the interface for classification code generation functionality
// This interface allows methods package to use code generation without importing classification package
type CodeGenerator interface {
	GenerateClassificationCodes(
		ctx context.Context,
		keywords []string,
		detectedIndustry string,
		confidence float64,
		additionalIndustries ...IndustryResult,
	) (*ClassificationCodesInfo, error)
}

// ScrapingResult represents the result of a website scraping operation
type ScrapingResult struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	Content       string            `json:"content"`
	TextContent   string            `json:"text_content"`
	Title         string            `json:"title,omitempty"` // Added for compatibility
	Keywords      []string          `json:"keywords"`
	ContentType   string            `json:"content_type"`
	ContentLength int64             `json:"content_length"`
	Headers       map[string]string `json:"headers"`
	FinalURL      string            `json:"final_url"`
	Success       bool              `json:"success"`
	Error         string            `json:"error,omitempty"`
}

// ClassificationCodesInfo contains the industry classification codes
type ClassificationCodesInfo struct {
	MCC   []MCCCode   `json:"mcc,omitempty"`
	SIC   []SICCode   `json:"sic,omitempty"`
	NAICS []NAICSCode `json:"naics,omitempty"`
}

// MCCCode represents a Merchant Category Code
type MCCCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// SICCode represents a Standard Industrial Classification code
type SICCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// NAICSCode represents a North American Industry Classification System code
type NAICSCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// IndustryResult represents an industry with its confidence score
type IndustryResult struct {
	IndustryName string
	Confidence   float64
}

