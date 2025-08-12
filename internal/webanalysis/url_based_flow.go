package webanalysis

import (
	"context"
	"fmt"
	"time"
)

// URLBasedFlow handles URL-based classification using website scraping
type URLBasedFlow struct {
	scraper             *WebScraper
	classifier          *IndustryClassifier
	riskDetector        *RiskDetector
	connectionValidator ConnectionValidatorInterface
	config              URLFlowConfig
}

// URLFlowConfig holds configuration for URL-based flow
type URLFlowConfig struct {
	MaxScrapingDepth           int           `json:"max_scraping_depth"`
	ScrapingTimeout            time.Duration `json:"scraping_timeout"`
	MinContentLength           int           `json:"min_content_length"`
	MaxPagesToScrape           int           `json:"max_pages_to_scrape"`
	EnableRiskAnalysis         bool          `json:"enable_risk_analysis"`
	EnableConnectionValidation bool          `json:"enable_connection_validation"`
}

// URLFlowResult represents the result of URL-based classification
type URLFlowResult struct {
	WebsiteData          *WebsiteAnalysis
	Industries           []IndustryClassification
	Confidence           float64
	RiskAssessment       *RiskAssessment
	ConnectionValidation *ConnectionValidation
	ProcessingTime       time.Duration
	Errors               []string
}

// NewURLBasedFlow creates a new URL-based flow
func NewURLBasedFlow() *URLBasedFlow {
	config := URLFlowConfig{
		MaxScrapingDepth:           3,
		ScrapingTimeout:            time.Second * 30,
		MinContentLength:           100,
		MaxPagesToScrape:           10,
		EnableRiskAnalysis:         true,
		EnableConnectionValidation: true,
	}

	return &URLBasedFlow{
		scraper:             NewWebScraper(NewProxyManager()),
		classifier:          NewIndustryClassifier(),
		riskDetector:        NewRiskDetector(),
		connectionValidator: NewConnectionValidatorAdapter(),
		config:              config,
	}
}

// Execute performs URL-based classification
func (uf *URLBasedFlow) Execute(ctx context.Context, req *ClassificationRequest) (*URLFlowResult, error) {
	start := time.Now()

	if req.WebsiteURL == "" {
		return nil, fmt.Errorf("website URL is required for URL-based flow")
	}

	// Step 1: Scrape website content
	websiteData, err := uf.scrapeWebsite(ctx, req.WebsiteURL)
	if err != nil {
		return &URLFlowResult{
			Errors: []string{fmt.Sprintf("Website scraping failed: %v", err)},
		}, fmt.Errorf("website scraping failed: %w", err)
	}

	// Step 2: Classify industries
	industries, err := uf.classifyIndustries(ctx, websiteData, req)
	if err != nil {
		return &URLFlowResult{
			WebsiteData: websiteData,
			Errors:      []string{fmt.Sprintf("Industry classification failed: %v", err)},
		}, fmt.Errorf("industry classification failed: %w", err)
	}

	// Step 3: Calculate overall confidence
	confidence := uf.calculateConfidence(industries, websiteData)

	// Step 4: Risk analysis (if enabled)
	var riskAssessment *RiskAssessment
	if uf.config.EnableRiskAnalysis && req.IncludeRiskAnalysis {
		riskAssessment, err = uf.analyzeRisks(ctx, websiteData, req)
		if err != nil {
			// Log error but don't fail the entire flow
			fmt.Printf("Risk analysis failed: %v", err)
		}
	}

	// Step 5: Connection validation (if enabled)
	var connectionValidation *ConnectionValidation
	if uf.config.EnableConnectionValidation && req.IncludeConnectionValidation {
		connectionValidation, err = uf.validateConnection(ctx, websiteData, req)
		if err != nil {
			// Log error but don't fail the entire flow
			fmt.Printf("Connection validation failed: %v", err)
		}
	}

	return &URLFlowResult{
		WebsiteData:          websiteData,
		Industries:           industries,
		Confidence:           confidence,
		RiskAssessment:       riskAssessment,
		ConnectionValidation: connectionValidation,
		ProcessingTime:       time.Since(start),
	}, nil
}

// scrapeWebsite scrapes the website content
func (uf *URLBasedFlow) scrapeWebsite(ctx context.Context, url string) (*WebsiteAnalysis, error) {
	// Create scraping job
	job := &ScrapingJob{
		URL:        url,
		Timeout:    uf.config.ScrapingTimeout,
		MaxRetries: 3,
	}

	// Scrape the website
	content, err := uf.scraper.ScrapeWebsite(job)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape website: %w", err)
	}

	// Create website analysis
	websiteData := &WebsiteAnalysis{
		URL:           url,
		Title:         content.Title,
		Content:       content.Text,
		ExtractedData: content.ExtractedData,
		PageCount:     1, // For now, just the main page
		ScrapingDepth: 1,
		QualityScore:  uf.calculateQualityScore(content),
	}

	return websiteData, nil
}

// classifyIndustries performs industry classification
func (uf *URLBasedFlow) classifyIndustries(ctx context.Context, websiteData *WebsiteAnalysis, req *ClassificationRequest) ([]IndustryClassification, error) {
	// Use the industry classifier to analyze the content
	classifications, err := uf.classifier.ClassifyContent(ctx, websiteData.Content, req.MaxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to classify industries: %w", err)
	}

	return classifications, nil
}

// analyzeRisks performs risk analysis
func (uf *URLBasedFlow) analyzeRisks(ctx context.Context, websiteData *WebsiteAnalysis, req *ClassificationRequest) (*RiskAssessment, error) {
	// Use the risk detector to analyze the content
	riskAssessment, err := uf.riskDetector.AnalyzeContent(ctx, websiteData.Content, req.BusinessName)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze risks: %w", err)
	}

	return riskAssessment, nil
}

// validateConnection validates business-website connection
func (uf *URLBasedFlow) validateConnection(ctx context.Context, websiteData *WebsiteAnalysis, req *ClassificationRequest) (*ConnectionValidation, error) {
	// Use the connection validator to validate the connection
	connectionValidation, err := uf.connectionValidator.ValidateConnection(ctx, websiteData, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate connection: %w", err)
	}

	return connectionValidation, nil
}

// calculateConfidence calculates overall confidence score
func (uf *URLBasedFlow) calculateConfidence(industries []IndustryClassification, websiteData *WebsiteAnalysis) float64 {
	if len(industries) == 0 {
		return 0.0
	}

	// Calculate average confidence of top industries
	totalConfidence := 0.0
	count := 0

	for i, industry := range industries {
		if i >= 3 { // Only consider top 3
			break
		}
		totalConfidence += industry.Confidence
		count++
	}

	if count == 0 {
		return 0.0
	}

	avgConfidence := totalConfidence / float64(count)

	// Adjust confidence based on website quality
	qualityAdjustment := websiteData.QualityScore * 0.2

	return avgConfidence + qualityAdjustment
}

// calculateQualityScore calculates the quality score of scraped content
func (uf *URLBasedFlow) calculateQualityScore(content *ScrapedContent) float64 {
	score := 0.0

	// Content length score
	if len(content.Text) > uf.config.MinContentLength {
		score += 0.3
	}

	// Title presence score
	if content.Title != "" {
		score += 0.2
	}

	// Extracted data score
	if len(content.ExtractedData) > 0 {
		score += 0.3
	}

	// Status code score
	if content.StatusCode == 200 {
		score += 0.2
	}

	return score
}

// GetStats returns statistics about URL-based flow usage
func (uf *URLBasedFlow) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_requests":          0,     // TODO: Implement counter
		"successful_scrapes":      0,     // TODO: Implement counter
		"average_confidence":      0.0,   // TODO: Implement calculation
		"average_processing_time": "0ms", // TODO: Implement calculation
	}
}
