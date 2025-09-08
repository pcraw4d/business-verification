package classification

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
	"github.com/pcraw4d/business-verification/internal/database"
)

// IntegrationService provides a simple interface for integrating classification services
type IntegrationService struct {
	container *ClassificationContainer
	logger    *log.Logger
}

// NewIntegrationService creates a new integration service
func NewIntegrationService(supabaseClient *database.SupabaseClient, logger *log.Logger) *IntegrationService {
	if logger == nil {
		logger = log.Default()
	}

	container := NewClassificationContainer(supabaseClient, logger)
	return &IntegrationService{
		container: container,
		logger:    logger,
	}
}

// ProcessBusinessClassification processes business classification using the new services
func (s *IntegrationService) ProcessBusinessClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) map[string]interface{} {
	s.logger.Printf("🔍 Processing business classification for: %s", businessName)

	// Get services from container
	industryDetectionService := s.container.GetIndustryDetectionService()
	codeGenerator := s.container.GetCodeGenerator()

	// Perform industry detection
	var industryResult *IndustryDetectionResult
	var err error

	if websiteURL != "" {
		// Use website content analysis
		websiteContent := s.scrapeWebsiteContent(websiteURL)
		industryResult, err = industryDetectionService.DetectIndustryFromContent(ctx, websiteContent)
	} else {
		// Use business information analysis
		industryResult, err = industryDetectionService.DetectIndustryFromBusinessInfo(
			ctx,
			businessName,
			description,
			websiteURL,
		)
	}

	// Handle detection failure
	if err != nil {
		s.logger.Printf("⚠️ Industry detection failed: %v", err)
		industryResult = s.createDefaultResult("Detection failed, using default")
	}

	// Extract keywords for classification codes
	var keywords []string
	if industryResult != nil {
		keywords = industryResult.KeywordsMatched
	}

	// Generate classification codes
	var classificationCodes *ClassificationCodesInfo
	if industryResult != nil {
		classificationCodes, err = codeGenerator.GenerateClassificationCodes(
			ctx,
			keywords,
			industryResult.Industry.Name,
			industryResult.Confidence,
		)
		if err != nil {
			s.logger.Printf("⚠️ Classification code generation failed: %v", err)
			classificationCodes = nil
		}
	}

	// Validate classification codes
	if classificationCodes != nil {
		if err := codeGenerator.ValidateClassificationCodes(classificationCodes, industryResult.Industry.Name); err != nil {
			s.logger.Printf("⚠️ Classification code validation failed: %v", err)
		}
	}

	// Get code statistics
	var codeStats map[string]interface{}
	if classificationCodes != nil {
		codeStats = codeGenerator.GetCodeStatistics(classificationCodes)
	}

	// Build response
	response := map[string]interface{}{
		"success": true,
		"classification_data": map[string]interface{}{
			"industry_detection": map[string]interface{}{
				"detected_industry": industryResult.Industry.Name,
				"confidence":        industryResult.Confidence,
				"keywords_matched":  industryResult.KeywordsMatched,
				"analysis_method":   industryResult.AnalysisMethod,
				"evidence":          industryResult.Evidence,
			},
			"classification_codes": classificationCodes,
			"code_statistics":      codeStats,
		},
		"enhanced_features": map[string]string{
			"database_driven_classification": "active",
			"modular_architecture":           "active",
		},
	}

	return response
}

// GetHealthStatus returns the health status of all classification services
func (s *IntegrationService) GetHealthStatus() map[string]interface{} {
	return s.container.HealthCheck()
}

// Close performs cleanup operations
func (s *IntegrationService) Close() error {
	return s.container.Close()
}

// createDefaultResult creates a default industry detection result
func (s *IntegrationService) createDefaultResult(reason string) *IndustryDetectionResult {
	return &IndustryDetectionResult{
		Industry: &repository.Industry{
			ID:   1,
			Name: "General Business",
		},
		Confidence:      0.5,
		KeywordsMatched: []string{},
		AnalysisMethod:  "Fallback (detection failed)",
		Evidence:        reason,
	}
}

// scrapeWebsiteContent performs basic website content scraping
func (s *IntegrationService) scrapeWebsiteContent(url string) string {
	// Import required packages
	client := &http.Client{Timeout: 10 * time.Second}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logger.Printf("❌ Error creating request for %s: %v", url, err)
		return "Error creating request"
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BusinessVerificationBot/1.0)")
	
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Printf("❌ Error fetching content from %s: %v", url, err)
		return "Error fetching content"
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Printf("❌ Error reading response from %s: %v", url, err)
		return "Error reading response"
	}
	
	content := string(body)
	s.logger.Printf("✅ Successfully scraped %d characters from %s", len(content), url)
	return content
}
