package methods

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"kyb-platform/internal/machine_learning"
	"kyb-platform/internal/machine_learning/infrastructure"
	"kyb-platform/internal/shared"
)

// MLClassificationMethod implements the ClassificationMethod interface for ML-based classification
type MLClassificationMethod struct {
	name            string
	methodType      string
	weight          float64
	enabled         bool
	description     string
	mlClassifier    *machine_learning.ContentClassifier
	pythonMLService interface{} // *infrastructure.PythonMLService - using interface to avoid import cycle
	websiteScraper  WebsiteScraper // Use interface instead of concrete type
	codeGenerator   CodeGenerator  // Use interface instead of concrete type
	logger          *log.Logger
}

// PythonMLServiceInterface defines the interface for Python ML service to avoid import cycles
type PythonMLServiceInterface interface {
	ClassifyEnhanced(ctx context.Context, req interface{}) (interface{}, error)
}

// NewMLClassificationMethod creates a new ML classification method
func NewMLClassificationMethod(
	mlClassifier *machine_learning.ContentClassifier,
	pythonMLService interface{}, // *infrastructure.PythonMLService - using interface to avoid import cycle
	websiteScraper WebsiteScraper, // Use interface instead of concrete type
	codeGenerator CodeGenerator,   // Use interface instead of concrete type
	logger *log.Logger,
) *MLClassificationMethod {
	if logger == nil {
		logger = log.Default()
	}

	return &MLClassificationMethod{
		name:            "ml_classification",
		methodType:      "ml",
		weight:          0.4,
		enabled:         true,
		description:     "Machine learning-based classification using DistilBART models with summarization and explanation",
		mlClassifier:    mlClassifier,
		pythonMLService: pythonMLService,
		websiteScraper:  websiteScraper,
		codeGenerator:   codeGenerator,
		logger:          logger,
	}
}

// GetName returns the unique name of the classification method
func (mlcm *MLClassificationMethod) GetName() string {
	return mlcm.name
}

// GetType returns the type/category of the method
func (mlcm *MLClassificationMethod) GetType() string {
	return mlcm.methodType
}

// GetDescription returns a human-readable description of what this method does
func (mlcm *MLClassificationMethod) GetDescription() string {
	return mlcm.description
}

// GetWeight returns the current weight/importance of this method in the ensemble
func (mlcm *MLClassificationMethod) GetWeight() float64 {
	return mlcm.weight
}

// SetWeight sets the weight/importance of this method in the ensemble
func (mlcm *MLClassificationMethod) SetWeight(weight float64) {
	mlcm.weight = weight
}

// IsEnabled returns whether this method is currently enabled
func (mlcm *MLClassificationMethod) IsEnabled() bool {
	return mlcm.enabled
}

// SetEnabled enables or disables this method
func (mlcm *MLClassificationMethod) SetEnabled(enabled bool) {
	mlcm.enabled = enabled
}

// Classify performs the actual classification using this method
func (mlcm *MLClassificationMethod) Classify(ctx context.Context, businessName, description, websiteURL string) (*shared.ClassificationMethodResult, error) {
	startTime := time.Now()

	// Validate input
	if err := mlcm.ValidateInput(businessName, description, websiteURL); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Perform ML-based classification
	result, err := mlcm.performMLClassification(ctx, businessName, description, websiteURL)
	if err != nil {
		return &shared.ClassificationMethodResult{
			MethodName:     mlcm.name,
			MethodType:     mlcm.methodType,
			Confidence:     0.0,
			ProcessingTime: time.Since(startTime),
			Success:        false,
			Error:          err.Error(),
		}, nil
	}

	// Create evidence from ML result
	evidence := []string{fmt.Sprintf("ML model prediction with confidence %.2f%%", result.ConfidenceScore*100)}

	return &shared.ClassificationMethodResult{
		MethodName:     mlcm.name,
		MethodType:     mlcm.methodType,
		Confidence:     result.ConfidenceScore,
		ProcessingTime: time.Since(startTime),
		Result:         result,
		Evidence:       evidence,
		Success:        true,
	}, nil
}

// GetPerformanceMetrics returns performance metrics for this method
func (mlcm *MLClassificationMethod) GetPerformanceMetrics() interface{} {
	// This would typically be stored and updated by the registry
	// For now, return a basic metrics structure
	return map[string]interface{}{
		"total_requests":        0,
		"successful_requests":   0,
		"failed_requests":       0,
		"average_response_time": 0,
		"accuracy_score":        0.0,
	}
}

// ValidateInput validates the input parameters before classification
func (mlcm *MLClassificationMethod) ValidateInput(businessName, description, websiteURL string) error {
	if strings.TrimSpace(businessName) == "" && strings.TrimSpace(description) == "" {
		return fmt.Errorf("at least one of business name or description must be provided")
	}
	return nil
}

// GetRequiredDependencies returns a list of dependencies this method requires
func (mlcm *MLClassificationMethod) GetRequiredDependencies() []string {
	return []string{"ml_model", "bert"}
}

// Initialize performs any necessary initialization for this method
func (mlcm *MLClassificationMethod) Initialize(ctx context.Context) error {
	if mlcm.mlClassifier == nil {
		return fmt.Errorf("ML classifier is required")
	}
	mlcm.logger.Printf("✅ Initialized ML classification method")
	return nil
}

// Cleanup performs any necessary cleanup when the method is removed
func (mlcm *MLClassificationMethod) Cleanup() error {
	mlcm.logger.Printf("✅ Cleaned up ML classification method")
	return nil
}

// performMLClassification performs the actual ML-based classification
func (mlcm *MLClassificationMethod) performMLClassification(ctx context.Context, businessName, description, websiteURL string) (*shared.IndustryClassification, error) {
	// Try enhanced classification if Python ML service is available and website URL exists
	if mlcm.pythonMLService != nil && websiteURL != "" {
		// Extract website content using existing Go scraper
		websiteContent := mlcm.extractWebsiteContent(ctx, websiteURL)
		// Use enhanced classification with extracted content
		return mlcm.performEnhancedClassification(ctx, businessName, description, websiteURL, websiteContent)
	}

	// Fallback to standard classification
	content := strings.TrimSpace(businessName + " " + description)
	if content == "" {
		return nil, fmt.Errorf("no content to analyze")
	}

	// Use the ML classifier to classify the content
	mlResult, err := mlcm.mlClassifier.ClassifyContent(ctx, content, "")
	if err != nil {
		return nil, fmt.Errorf("failed to classify content with ML: %w", err)
	}

	// Convert ML result to IndustryClassification
	if len(mlResult.Classifications) == 0 {
		return &shared.IndustryClassification{
			IndustryCode:         "unknown",
			IndustryName:         "Unknown",
			ConfidenceScore:      0.0,
			ClassificationMethod: "ml",
			Keywords:             []string{},
		}, nil
	}

	// Get the top classification
	topClassification := mlResult.Classifications[0]

	// Convert to IndustryClassification format
	industryClassification := &shared.IndustryClassification{
		IndustryCode:         topClassification.Label,
		IndustryName:         strings.Title(topClassification.Label),
		ConfidenceScore:      topClassification.Confidence,
		ClassificationMethod: "ml",
		Keywords:             []string{}, // ML doesn't provide specific keywords
	}

	return industryClassification, nil
}

// performEnhancedClassification performs enhanced classification with DistilBART
func (mlcm *MLClassificationMethod) performEnhancedClassification(
	ctx context.Context,
	businessName, description, websiteURL, websiteContent string,
) (*shared.IndustryClassification, error) {
	// Prepare enhanced classification request
	req := &infrastructure.EnhancedClassificationRequest{
		BusinessName:     businessName,
		Description:      description,
		WebsiteURL:       websiteURL,
		WebsiteContent:   websiteContent,
		MaxResults:       5,
		MaxContentLength: 1024,
	}

	// Call enhanced classification endpoint
	// Type assert to get the actual PythonMLService
	pms, ok := mlcm.pythonMLService.(interface {
		ClassifyEnhanced(ctx context.Context, req *infrastructure.EnhancedClassificationRequest) (*infrastructure.EnhancedClassificationResponse, error)
	})
	if !ok {
		mlcm.logger.Printf("⚠️ Python ML service type assertion failed, falling back to standard classification")
		// Fallback to standard classification
		content := strings.TrimSpace(businessName + " " + description)
		if content == "" {
			return nil, fmt.Errorf("no content to analyze and Python ML service unavailable")
		}
		
		mlResult, err2 := mlcm.mlClassifier.ClassifyContent(ctx, content, "")
		if err2 != nil {
			return nil, fmt.Errorf("failed to classify content with ML: %w", err2)
		}
		
		if len(mlResult.Classifications) == 0 {
			return &shared.IndustryClassification{
				IndustryCode:         "unknown",
				IndustryName:         "Unknown",
				ConfidenceScore:      0.0,
				ClassificationMethod: "ml",
				Keywords:             []string{},
			}, nil
		}
		
		topClassification := mlResult.Classifications[0]
		return &shared.IndustryClassification{
			IndustryCode:         topClassification.Label,
			IndustryName:         strings.Title(topClassification.Label),
			ConfidenceScore:      topClassification.Confidence,
			ClassificationMethod: "ml",
			Keywords:             []string{},
		}, nil
	}
	
	enhancedResp, err := pms.ClassifyEnhanced(ctx, req)
	if err != nil {
		mlcm.logger.Printf("⚠️ Enhanced classification failed, falling back to standard: %v", err)
		// Fallback to standard classification (avoid recursion by not checking websiteURL again)
		content := strings.TrimSpace(businessName + " " + description)
		if content == "" {
			return nil, fmt.Errorf("no content to analyze and enhanced classification failed: %w", err)
		}
		
		// Use the ML classifier to classify the content
		mlResult, err2 := mlcm.mlClassifier.ClassifyContent(ctx, content, "")
		if err2 != nil {
			return nil, fmt.Errorf("failed to classify content with ML: %w", err2)
		}
		
		// Convert ML result to IndustryClassification
		if len(mlResult.Classifications) == 0 {
			return &shared.IndustryClassification{
				IndustryCode:         "unknown",
				IndustryName:         "Unknown",
				ConfidenceScore:      0.0,
				ClassificationMethod: "ml",
				Keywords:             []string{},
			}, nil
		}
		
		// Get the top classification
		topClassification := mlResult.Classifications[0]
		
		return &shared.IndustryClassification{
			IndustryCode:         topClassification.Label,
			IndustryName:         strings.Title(topClassification.Label),
			ConfidenceScore:      topClassification.Confidence,
			ClassificationMethod: "ml",
			Keywords:             []string{},
		}, nil
	}

	// Build enhanced result with all required fields
	return mlcm.buildEnhancedResult(ctx, enhancedResp, businessName), nil
}

// buildEnhancedResult builds IndustryClassification from enhanced response
func (mlcm *MLClassificationMethod) buildEnhancedResult(
	ctx context.Context,
	enhancedResp *infrastructure.EnhancedClassificationResponse,
	businessName string,
) *shared.IndustryClassification {
	// Get primary industry and confidence
	primaryIndustry := "Unknown"
	confidence := 0.0
	if len(enhancedResp.Classifications) > 0 {
		primaryIndustry = enhancedResp.Classifications[0].Label
		confidence = enhancedResp.Classifications[0].Confidence
	}

	// Build all industry scores map
	allScores := make(map[string]float64)
	for _, classification := range enhancedResp.Classifications {
		allScores[classification.Label] = classification.Confidence
	}

	// Generate classification codes using existing code generator
	codes := shared.ClassificationCodes{
		MCC:   []shared.MCCCode{},
		SIC:   []shared.SICCode{},
		NAICS: []shared.NAICSCode{},
	}

	if mlcm.codeGenerator != nil {
		// Extract keywords from summary and explanation for code generation
		keywords := mlcm.extractKeywordsFromSummary(enhancedResp.Summary, enhancedResp.Explanation)
		
		// Generate codes using existing ClassificationCodeGenerator
		codeInfo, err := mlcm.codeGenerator.GenerateClassificationCodes(
			ctx,
			keywords,
			primaryIndustry,
			confidence,
		)
		
		if err == nil && codeInfo != nil {
			// Convert ClassificationCodesInfo to shared.ClassificationCodes
			codes.MCC = mlcm.convertMCCCodes(codeInfo.MCC)
			codes.SIC = mlcm.convertSICCodes(codeInfo.SIC)
			codes.NAICS = mlcm.convertNAICSCodes(codeInfo.NAICS)
			mlcm.logger.Printf("✅ Generated %d MCC, %d SIC, %d NAICS codes using code generator",
				len(codes.MCC), len(codes.SIC), len(codes.NAICS))
		} else {
			mlcm.logger.Printf("⚠️ Code generation failed: %v", err)
		}
	} else {
		mlcm.logger.Printf("⚠️ Code generator not available, codes will be empty")
	}

	// Calculate code distribution
	codeDistribution := codes.CalculateCodeDistribution()

	// Determine risk level based on confidence
	riskLevel := "low"
	if confidence < 0.5 {
		riskLevel = "high"
	} else if confidence < 0.7 {
		riskLevel = "medium"
	}

	return &shared.IndustryClassification{
		IndustryCode:         primaryIndustry,
		IndustryName:         primaryIndustry,
		PrimaryIndustry:     primaryIndustry,
		ConfidenceScore:      confidence,
		ClassificationMethod: "ml_distilbart",
		ContentSummary:       enhancedResp.Summary,
		Explanation:          enhancedResp.Explanation,
		AllIndustryScores:    allScores,
		QuantizationEnabled:  enhancedResp.QuantizationEnabled,
		ModelVersion:         enhancedResp.ModelVersion,
		RiskLevel:            riskLevel,
		CodeDistribution:     &codeDistribution,
		Keywords:             []string{},
	}
}

// extractKeywordsFromSummary extracts keywords from summary and explanation for code generation
func (mlcm *MLClassificationMethod) extractKeywordsFromSummary(summary, explanation string) []string {
	var keywords []string
	
	// Simple keyword extraction - split by spaces and filter common words
	allText := strings.ToLower(summary + " " + explanation)
	words := strings.Fields(allText)
	
	// Filter out common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "are": true,
		"was": true, "were": true, "been": true, "be": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"this": true, "that": true, "these": true, "those": true,
	}
	
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// convertMCCCodes converts methods.MCCCode to shared.MCCCode
func (mlcm *MLClassificationMethod) convertMCCCodes(codes []MCCCode) []shared.MCCCode {
	result := make([]shared.MCCCode, len(codes))
	for i, code := range codes {
		result[i] = shared.MCCCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}
	}
	return result
}

// convertSICCodes converts methods.SICCode to shared.SICCode
func (mlcm *MLClassificationMethod) convertSICCodes(codes []SICCode) []shared.SICCode {
	result := make([]shared.SICCode, len(codes))
	for i, code := range codes {
		result[i] = shared.SICCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}
	}
	return result
}

// convertNAICSCodes converts methods.NAICSCode to shared.NAICSCode
func (mlcm *MLClassificationMethod) convertNAICSCodes(codes []NAICSCode) []shared.NAICSCode {
	result := make([]shared.NAICSCode, len(codes))
	for i, code := range codes {
		result[i] = shared.NAICSCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}
	}
	return result
}

// extractWebsiteContent extracts content from website URL using existing Go scraper
func (mlcm *MLClassificationMethod) extractWebsiteContent(ctx context.Context, websiteURL string) string {
	if mlcm.websiteScraper == nil {
		mlcm.logger.Printf("⚠️ Website scraper not available, skipping content extraction")
		return ""
	}

	// Use existing EnhancedWebsiteScraper to extract content
	scrapingResult := mlcm.websiteScraper.ScrapeWebsite(ctx, websiteURL)
	if !scrapingResult.Success {
		mlcm.logger.Printf("⚠️ Website scraping failed for %s: %s", websiteURL, scrapingResult.Error)
		return ""
	}

	// Combine extracted text content
	content := scrapingResult.TextContent
	if scrapingResult.Title != "" {
		content = scrapingResult.Title + " " + content
	}

	mlcm.logger.Printf("✅ Extracted %d characters from website: %s", len(content), websiteURL)
	return content
}
