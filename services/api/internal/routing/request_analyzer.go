package routing

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"kyb-platform/internal/observability"
	"kyb-platform/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RequestAnalyzer analyzes classification requests to determine optimal routing
type RequestAnalyzer struct {
	logger *observability.Logger
	tracer trace.Tracer
	config RequestAnalyzerConfig
}

// RequestAnalyzerConfig holds configuration for request analysis
type RequestAnalyzerConfig struct {
	EnableComplexityAnalysis bool                 `json:"enable_complexity_analysis"`
	EnablePriorityAssessment bool                 `json:"enable_priority_assessment"`
	MaxRequestSize           int                  `json:"max_request_size"`
	DefaultTimeout           time.Duration        `json:"default_timeout"`
	ComplexityThresholds     ComplexityThresholds `json:"complexity_thresholds"`
	PriorityWeights          PriorityWeights      `json:"priority_weights"`
}

// ComplexityThresholds defines thresholds for complexity analysis
type ComplexityThresholds struct {
	LowComplexity    float64 `json:"low_complexity"`
	MediumComplexity float64 `json:"medium_complexity"`
	HighComplexity   float64 `json:"high_complexity"`
}

// PriorityWeights defines weights for priority assessment
type PriorityWeights struct {
	BusinessNameWeight     float64 `json:"business_name_weight"`
	WebsiteURLWeight       float64 `json:"website_url_weight"`
	KeywordsWeight         float64 `json:"keywords_weight"`
	DescriptionWeight      float64 `json:"description_weight"`
	IndustryWeight         float64 `json:"industry_weight"`
	GeographicRegionWeight float64 `json:"geographic_region_weight"`
}

// RequestAnalysisResult represents the result of request analysis
type RequestAnalysisResult struct {
	RequestID            string                   `json:"request_id"`
	RequestType          RequestType              `json:"request_type"`
	Complexity           ComplexityLevel          `json:"complexity"`
	Priority             PriorityLevel            `json:"priority"`
	ResourceRequirements ResourceRequirements     `json:"resource_requirements"`
	ProcessingTime       time.Duration            `json:"processing_time"`
	AnalysisMetadata     map[string]interface{}   `json:"analysis_metadata"`
	Recommendations      []RoutingRecommendation  `json:"recommendations"`
	ValidationResult     *shared.ValidationResult `json:"validation_result"`
}

// RequestType represents the type of classification request
type RequestType string

const (
	RequestTypeSimple   RequestType = "simple"
	RequestTypeStandard RequestType = "standard"
	RequestTypeComplex  RequestType = "complex"
	RequestTypeBatch    RequestType = "batch"
	RequestTypeUrgent   RequestType = "urgent"
	RequestTypeResearch RequestType = "research"
)

// ComplexityLevel represents the complexity level of a request
type ComplexityLevel string

const (
	ComplexityLevelLow    ComplexityLevel = "low"
	ComplexityLevelMedium ComplexityLevel = "medium"
	ComplexityLevelHigh   ComplexityLevel = "high"
)

// PriorityLevel represents the priority level of a request
type PriorityLevel string

const (
	PriorityLevelLow    PriorityLevel = "low"
	PriorityLevelMedium PriorityLevel = "medium"
	PriorityLevelHigh   PriorityLevel = "high"
	PriorityLevelUrgent PriorityLevel = "urgent"
)

// ResourceRequirements represents the resource requirements for processing
type ResourceRequirements struct {
	CPUIntensity     float64       `json:"cpu_intensity"`
	MemoryIntensity  float64       `json:"memory_intensity"`
	NetworkIntensity float64       `json:"network_intensity"`
	EstimatedTime    time.Duration `json:"estimated_time"`
	Concurrency      int           `json:"concurrency"`
}

// RoutingRecommendation represents a routing recommendation
type RoutingRecommendation struct {
	ModuleType      string                 `json:"module_type"`
	Confidence      float64                `json:"confidence"`
	Reason          string                 `json:"reason"`
	ExpectedLatency time.Duration          `json:"expected_latency"`
	ResourceUsage   ResourceRequirements   `json:"resource_usage"`
	FallbackModules []string               `json:"fallback_modules"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewRequestAnalyzer creates a new request analyzer
func NewRequestAnalyzer(
	logger *observability.Logger,
	tracer trace.Tracer,
	config RequestAnalyzerConfig,
) *RequestAnalyzer {
	// Set default configuration if not provided
	if config.MaxRequestSize == 0 {
		config.MaxRequestSize = 10000
	}
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.ComplexityThresholds.LowComplexity == 0 {
		config.ComplexityThresholds = ComplexityThresholds{
			LowComplexity:    0.3,
			MediumComplexity: 0.7,
			HighComplexity:   1.0,
		}
	}
	if config.PriorityWeights.BusinessNameWeight == 0 {
		config.PriorityWeights = PriorityWeights{
			BusinessNameWeight:     0.3,
			WebsiteURLWeight:       0.25,
			KeywordsWeight:         0.2,
			DescriptionWeight:      0.15,
			IndustryWeight:         0.05,
			GeographicRegionWeight: 0.05,
		}
	}

	return &RequestAnalyzer{
		logger: logger,
		tracer: tracer,
		config: config,
	}
}

// AnalyzeRequest performs comprehensive analysis of a classification request
func (ra *RequestAnalyzer) AnalyzeRequest(ctx context.Context, req *shared.BusinessClassificationRequest) (*RequestAnalysisResult, error) {
	ctx, span := ra.tracer.Start(ctx, "RequestAnalyzer.AnalyzeRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.id", req.ID),
		attribute.String("business.name", req.BusinessName),
	)

	startTime := time.Now()

	// Step 1: Validate request
	validationResult, err := ra.validateRequest(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Step 2: Determine request type
	requestType := ra.determineRequestType(req)

	// Step 3: Analyze complexity
	complexity := ra.analyzeComplexity(req)

	// Step 4: Assess priority
	priority := ra.assessPriority(req)

	// Step 5: Calculate resource requirements
	resourceRequirements := ra.calculateResourceRequirements(req, complexity, requestType)

	// Step 6: Generate routing recommendations
	recommendations := ra.generateRoutingRecommendations(req, requestType, complexity, priority)

	// Step 7: Create analysis result
	result := &RequestAnalysisResult{
		RequestID:            req.ID,
		RequestType:          requestType,
		Complexity:           complexity,
		Priority:             priority,
		ResourceRequirements: resourceRequirements,
		ProcessingTime:       time.Since(startTime),
		AnalysisMetadata:     ra.generateAnalysisMetadata(req),
		Recommendations:      recommendations,
		ValidationResult:     validationResult,
	}

	// Log analysis results
	ra.logger.WithComponent("request_analyzer").Info("request_analysis_completed", map[string]interface{}{
		"request_id":         req.ID,
		"request_type":       requestType,
		"complexity":         complexity,
		"priority":           priority,
		"processing_time_ms": result.ProcessingTime.Milliseconds(),
	})

	return result, nil
}

// validateRequest validates the classification request
func (ra *RequestAnalyzer) validateRequest(ctx context.Context, req *shared.BusinessClassificationRequest) (*shared.ValidationResult, error) {
	ctx, span := ra.tracer.Start(ctx, "RequestAnalyzer.validateRequest")
	defer span.End()

	// Use shared validation
	validationResult, err := shared.ValidateBusinessClassificationRequest(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Additional custom validation
	if !validationResult.IsValid {
		ra.logger.WithComponent("request_analyzer").Warn("request_validation_failed", map[string]interface{}{
			"request_id": req.ID,
			"errors":     validationResult.Errors,
		})
	}

	return validationResult, nil
}

// determineRequestType determines the type of classification request
func (ra *RequestAnalyzer) determineRequestType(req *shared.BusinessClassificationRequest) RequestType {
	// Check for batch request indicators
	if req.Metadata != nil {
		if batchSize, ok := req.Metadata["batch_size"].(int); ok && batchSize > 1 {
			return RequestTypeBatch
		}
		if urgent, ok := req.Metadata["urgent"].(bool); ok && urgent {
			return RequestTypeUrgent
		}
		if research, ok := req.Metadata["research"].(bool); ok && research {
			return RequestTypeResearch
		}
	}

	// Analyze request characteristics
	hasWebsite := req.WebsiteURL != ""
	hasKeywords := len(req.Keywords) > 0
	hasDescription := req.Description != ""

	// Determine type based on available information
	if hasWebsite && hasKeywords && hasDescription {
		return RequestTypeComplex
	} else if hasWebsite || (hasKeywords && hasDescription) {
		return RequestTypeStandard
	} else {
		return RequestTypeSimple
	}
}

// analyzeComplexity analyzes the complexity of the request
func (ra *RequestAnalyzer) analyzeComplexity(req *shared.BusinessClassificationRequest) ComplexityLevel {
	if !ra.config.EnableComplexityAnalysis {
		return ComplexityLevelMedium
	}

	complexityScore := ra.calculateComplexityScore(req)

	if complexityScore <= ra.config.ComplexityThresholds.LowComplexity {
		return ComplexityLevelLow
	} else if complexityScore <= ra.config.ComplexityThresholds.MediumComplexity {
		return ComplexityLevelMedium
	} else {
		return ComplexityLevelHigh
	}
}

// calculateComplexityScore calculates a complexity score for the request
func (ra *RequestAnalyzer) calculateComplexityScore(req *shared.BusinessClassificationRequest) float64 {
	score := 0.0

	// Business name complexity
	if req.BusinessName != "" {
		score += ra.calculateBusinessNameComplexity(req.BusinessName)
	}

	// Website URL complexity
	if req.WebsiteURL != "" {
		score += 0.2 // Base score for having a website
		score += ra.calculateWebsiteComplexity(req.WebsiteURL)
	}

	// Keywords complexity
	if len(req.Keywords) > 0 {
		score += float64(len(req.Keywords)) * 0.05
		score += ra.calculateKeywordsComplexity(req.Keywords)
	}

	// Description complexity
	if req.Description != "" {
		score += ra.calculateDescriptionComplexity(req.Description)
	}

	// Industry complexity
	if req.Industry != "" {
		score += 0.1
	}

	// Geographic region complexity
	if req.GeographicRegion != "" {
		score += 0.1
	}

	// Normalize score to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateBusinessNameComplexity calculates complexity based on business name
func (ra *RequestAnalyzer) calculateBusinessNameComplexity(businessName string) float64 {
	score := 0.0

	// Length complexity
	nameLength := len(businessName)
	if nameLength > 50 {
		score += 0.2
	} else if nameLength > 30 {
		score += 0.1
	}

	// Special characters complexity
	specialCharPattern := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	specialChars := specialCharPattern.FindAllString(businessName, -1)
	score += float64(len(specialChars)) * 0.05

	// Word count complexity
	words := strings.Fields(businessName)
	score += float64(len(words)) * 0.02

	return score
}

// calculateWebsiteComplexity calculates complexity based on website URL
func (ra *RequestAnalyzer) calculateWebsiteComplexity(websiteURL string) float64 {
	score := 0.0

	// URL length complexity
	if len(websiteURL) > 100 {
		score += 0.1
	}

	// Subdomain complexity
	subdomainPattern := regexp.MustCompile(`^https?://[^.]+\.`)
	subdomains := subdomainPattern.FindAllString(websiteURL, -1)
	score += float64(len(subdomains)) * 0.05

	// Path complexity
	if strings.Contains(websiteURL, "/") {
		pathParts := strings.Split(websiteURL, "/")
		score += float64(len(pathParts)) * 0.02
	}

	return score
}

// calculateKeywordsComplexity calculates complexity based on keywords
func (ra *RequestAnalyzer) calculateKeywordsComplexity(keywords []string) float64 {
	score := 0.0

	for _, keyword := range keywords {
		// Keyword length complexity
		if len(keyword) > 20 {
			score += 0.05
		}

		// Special characters in keywords
		specialCharPattern := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
		specialChars := specialCharPattern.FindAllString(keyword, -1)
		score += float64(len(specialChars)) * 0.02
	}

	return score
}

// calculateDescriptionComplexity calculates complexity based on description
func (ra *RequestAnalyzer) calculateDescriptionComplexity(description string) float64 {
	score := 0.0

	// Length complexity
	descLength := len(description)
	if descLength > 1000 {
		score += 0.3
	} else if descLength > 500 {
		score += 0.2
	} else if descLength > 200 {
		score += 0.1
	}

	// Technical terms complexity
	technicalTerms := []string{"software", "technology", "development", "engineering", "consulting", "services"}
	for _, term := range technicalTerms {
		if strings.Contains(strings.ToLower(description), term) {
			score += 0.05
		}
	}

	return score
}

// assessPriority assesses the priority level of the request
func (ra *RequestAnalyzer) assessPriority(req *shared.BusinessClassificationRequest) PriorityLevel {
	if !ra.config.EnablePriorityAssessment {
		return PriorityLevelMedium
	}

	priorityScore := ra.calculatePriorityScore(req)

	if priorityScore >= 0.8 {
		return PriorityLevelUrgent
	} else if priorityScore >= 0.6 {
		return PriorityLevelHigh
	} else if priorityScore >= 0.4 {
		return PriorityLevelMedium
	} else {
		return PriorityLevelLow
	}
}

// calculatePriorityScore calculates a priority score for the request
func (ra *RequestAnalyzer) calculatePriorityScore(req *shared.BusinessClassificationRequest) float64 {
	score := 0.0

	// Business name priority
	if req.BusinessName != "" {
		score += ra.config.PriorityWeights.BusinessNameWeight
	}

	// Website URL priority
	if req.WebsiteURL != "" {
		score += ra.config.PriorityWeights.WebsiteURLWeight
	}

	// Keywords priority
	if len(req.Keywords) > 0 {
		score += ra.config.PriorityWeights.KeywordsWeight
	}

	// Description priority
	if req.Description != "" {
		score += ra.config.PriorityWeights.DescriptionWeight
	}

	// Industry priority
	if req.Industry != "" {
		score += ra.config.PriorityWeights.IndustryWeight
	}

	// Geographic region priority
	if req.GeographicRegion != "" {
		score += ra.config.PriorityWeights.GeographicRegionWeight
	}

	// Metadata-based priority adjustments
	if req.Metadata != nil {
		if urgent, ok := req.Metadata["urgent"].(bool); ok && urgent {
			score += 0.3
		}
		if priority, ok := req.Metadata["priority"].(string); ok {
			switch priority {
			case "high":
				score += 0.2
			case "urgent":
				score += 0.4
			}
		}
	}

	// Normalize score to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateResourceRequirements calculates resource requirements for processing
func (ra *RequestAnalyzer) calculateResourceRequirements(
	req *shared.BusinessClassificationRequest,
	complexity ComplexityLevel,
	requestType RequestType,
) ResourceRequirements {
	requirements := ResourceRequirements{
		CPUIntensity:     0.5,
		MemoryIntensity:  0.5,
		NetworkIntensity: 0.5,
		EstimatedTime:    ra.config.DefaultTimeout,
		Concurrency:      1,
	}

	// Adjust based on complexity
	switch complexity {
	case ComplexityLevelLow:
		requirements.CPUIntensity = 0.3
		requirements.MemoryIntensity = 0.3
		requirements.NetworkIntensity = 0.2
		requirements.EstimatedTime = 5 * time.Second
	case ComplexityLevelMedium:
		requirements.CPUIntensity = 0.6
		requirements.MemoryIntensity = 0.6
		requirements.NetworkIntensity = 0.5
		requirements.EstimatedTime = 15 * time.Second
	case ComplexityLevelHigh:
		requirements.CPUIntensity = 0.9
		requirements.MemoryIntensity = 0.8
		requirements.NetworkIntensity = 0.8
		requirements.EstimatedTime = 30 * time.Second
	}

	// Adjust based on request type
	switch requestType {
	case RequestTypeSimple:
		requirements.Concurrency = 1
	case RequestTypeStandard:
		requirements.Concurrency = 2
	case RequestTypeComplex:
		requirements.Concurrency = 3
	case RequestTypeBatch:
		requirements.Concurrency = 5
		requirements.EstimatedTime *= 2
	case RequestTypeUrgent:
		requirements.Concurrency = 4
		requirements.EstimatedTime = requirements.EstimatedTime / 2
	case RequestTypeResearch:
		requirements.Concurrency = 2
		requirements.EstimatedTime *= 3
	}

	// Adjust based on available information
	if req.WebsiteURL != "" {
		requirements.NetworkIntensity += 0.2
		requirements.EstimatedTime += 10 * time.Second
	}

	if len(req.Keywords) > 5 {
		requirements.CPUIntensity += 0.1
		requirements.EstimatedTime += 5 * time.Second
	}

	return requirements
}

// generateRoutingRecommendations generates routing recommendations
func (ra *RequestAnalyzer) generateRoutingRecommendations(
	req *shared.BusinessClassificationRequest,
	requestType RequestType,
	complexity ComplexityLevel,
	priority PriorityLevel,
) []RoutingRecommendation {
	var recommendations []RoutingRecommendation

	// Generate recommendations based on request characteristics
	if req.WebsiteURL != "" {
		recommendations = append(recommendations, ra.createWebsiteAnalysisRecommendation(req, complexity, priority))
	}

	if req.BusinessName != "" && (len(req.Keywords) > 0 || req.Description != "") {
		recommendations = append(recommendations, ra.createMLClassificationRecommendation(req, complexity, priority))
	}

	if req.BusinessName != "" {
		recommendations = append(recommendations, ra.createWebSearchRecommendation(req, complexity, priority))
	}

	// Always include keyword classification as fallback
	recommendations = append(recommendations, ra.createKeywordClassificationRecommendation(req, complexity, priority))

	// Sort recommendations by confidence
	ra.sortRecommendationsByConfidence(recommendations)

	return recommendations
}

// createWebsiteAnalysisRecommendation creates a recommendation for website analysis
func (ra *RequestAnalyzer) createWebsiteAnalysisRecommendation(
	req *shared.BusinessClassificationRequest,
	complexity ComplexityLevel,
	priority PriorityLevel,
) RoutingRecommendation {
	confidence := 0.9
	expectedLatency := 20 * time.Second

	// Adjust confidence and latency based on complexity
	switch complexity {
	case ComplexityLevelLow:
		confidence = 0.95
		expectedLatency = 15 * time.Second
	case ComplexityLevelHigh:
		confidence = 0.85
		expectedLatency = 30 * time.Second
	}

	return RoutingRecommendation{
		ModuleType:      "website_analysis",
		Confidence:      confidence,
		Reason:          "Website URL provided - website analysis will provide most accurate results",
		ExpectedLatency: expectedLatency,
		ResourceUsage: ResourceRequirements{
			CPUIntensity:     0.7,
			MemoryIntensity:  0.6,
			NetworkIntensity: 0.8,
			EstimatedTime:    expectedLatency,
			Concurrency:      2,
		},
		FallbackModules: []string{"web_search_analysis", "ml_classification"},
		Metadata: map[string]interface{}{
			"website_url": req.WebsiteURL,
			"complexity":  complexity,
		},
	}
}

// createMLClassificationRecommendation creates a recommendation for ML classification
func (ra *RequestAnalyzer) createMLClassificationRecommendation(
	req *shared.BusinessClassificationRequest,
	complexity ComplexityLevel,
	priority PriorityLevel,
) RoutingRecommendation {
	confidence := 0.8
	expectedLatency := 10 * time.Second

	// Adjust confidence based on available information
	if req.Description != "" && len(req.Keywords) > 0 {
		confidence = 0.85
	}

	// Adjust based on complexity
	switch complexity {
	case ComplexityLevelLow:
		confidence = 0.9
		expectedLatency = 8 * time.Second
	case ComplexityLevelHigh:
		confidence = 0.75
		expectedLatency = 15 * time.Second
	}

	return RoutingRecommendation{
		ModuleType:      "ml_classification",
		Confidence:      confidence,
		Reason:          "Rich business information available - ML classification will provide accurate results",
		ExpectedLatency: expectedLatency,
		ResourceUsage: ResourceRequirements{
			CPUIntensity:     0.8,
			MemoryIntensity:  0.7,
			NetworkIntensity: 0.3,
			EstimatedTime:    expectedLatency,
			Concurrency:      1,
		},
		FallbackModules: []string{"keyword_classification", "web_search_analysis"},
		Metadata: map[string]interface{}{
			"has_description": req.Description != "",
			"keywords_count":  len(req.Keywords),
			"complexity":      complexity,
		},
	}
}

// createWebSearchRecommendation creates a recommendation for web search analysis
func (ra *RequestAnalyzer) createWebSearchRecommendation(
	req *shared.BusinessClassificationRequest,
	complexity ComplexityLevel,
	priority PriorityLevel,
) RoutingRecommendation {
	confidence := 0.7
	expectedLatency := 25 * time.Second

	// Adjust confidence based on business name quality
	if len(req.BusinessName) > 10 {
		confidence = 0.75
	}

	return RoutingRecommendation{
		ModuleType:      "web_search_analysis",
		Confidence:      confidence,
		Reason:          "Business name provided - web search will find relevant information",
		ExpectedLatency: expectedLatency,
		ResourceUsage: ResourceRequirements{
			CPUIntensity:     0.6,
			MemoryIntensity:  0.5,
			NetworkIntensity: 0.9,
			EstimatedTime:    expectedLatency,
			Concurrency:      3,
		},
		FallbackModules: []string{"keyword_classification"},
		Metadata: map[string]interface{}{
			"business_name": req.BusinessName,
			"complexity":    complexity,
		},
	}
}

// createKeywordClassificationRecommendation creates a recommendation for keyword classification
func (ra *RequestAnalyzer) createKeywordClassificationRecommendation(
	req *shared.BusinessClassificationRequest,
	complexity ComplexityLevel,
	priority PriorityLevel,
) RoutingRecommendation {
	confidence := 0.6
	expectedLatency := 5 * time.Second

	// Adjust confidence based on available information
	if len(req.Keywords) > 0 {
		confidence = 0.7
	}

	return RoutingRecommendation{
		ModuleType:      "keyword_classification",
		Confidence:      confidence,
		Reason:          "Fast and reliable fallback classification method",
		ExpectedLatency: expectedLatency,
		ResourceUsage: ResourceRequirements{
			CPUIntensity:     0.4,
			MemoryIntensity:  0.3,
			NetworkIntensity: 0.1,
			EstimatedTime:    expectedLatency,
			Concurrency:      1,
		},
		FallbackModules: []string{},
		Metadata: map[string]interface{}{
			"keywords_count": len(req.Keywords),
			"complexity":     complexity,
		},
	}
}

// sortRecommendationsByConfidence sorts recommendations by confidence score
func (ra *RequestAnalyzer) sortRecommendationsByConfidence(recommendations []RoutingRecommendation) {
	// Simple bubble sort for small lists
	for i := 0; i < len(recommendations)-1; i++ {
		for j := 0; j < len(recommendations)-i-1; j++ {
			if recommendations[j].Confidence < recommendations[j+1].Confidence {
				recommendations[j], recommendations[j+1] = recommendations[j+1], recommendations[j]
			}
		}
	}
}

// generateAnalysisMetadata generates metadata for the analysis
func (ra *RequestAnalyzer) generateAnalysisMetadata(req *shared.BusinessClassificationRequest) map[string]interface{} {
	metadata := map[string]interface{}{
		"analyzer_version":      "1.0.0",
		"analysis_timestamp":    time.Now().Unix(),
		"has_website_url":       req.WebsiteURL != "",
		"has_business_name":     req.BusinessName != "",
		"has_keywords":          len(req.Keywords) > 0,
		"has_description":       req.Description != "",
		"has_industry":          req.Industry != "",
		"has_geographic_region": req.GeographicRegion != "",
		"keywords_count":        len(req.Keywords),
		"business_name_length":  len(req.BusinessName),
	}

	// Calculate data quality and completeness scores
	dataQuality := ra.calculateDataQuality(req)
	dataCompleteness := ra.calculateDataCompleteness(req)
	inputComplexity := ra.calculateInputComplexity(req)

	metadata["data_quality"] = dataQuality
	metadata["data_completeness"] = dataCompleteness
	metadata["input_complexity"] = inputComplexity

	if req.Description != "" {
		metadata["description_length"] = len(req.Description)
	}

	if req.WebsiteURL != "" {
		metadata["website_url_length"] = len(req.WebsiteURL)
	}

	return metadata
}

// calculateDataQuality calculates the quality score of the input data
func (ra *RequestAnalyzer) calculateDataQuality(req *shared.BusinessClassificationRequest) float64 {
	score := 0.0
	totalFields := 0

	// Business name quality
	if req.BusinessName != "" {
		totalFields++
		if len(req.BusinessName) >= 3 && len(req.BusinessName) <= 100 {
			score += 0.9 // Good length
		} else if len(req.BusinessName) > 100 {
			score += 0.7 // Too long
		} else {
			score += 0.5 // Too short
		}
	}

	// Website URL quality
	if req.WebsiteURL != "" {
		totalFields++
		if strings.HasPrefix(req.WebsiteURL, "http") {
			score += 0.9 // Valid URL format
		} else {
			score += 0.6 // Invalid URL format
		}
	}

	// Description quality
	if req.Description != "" {
		totalFields++
		if len(req.Description) >= 10 && len(req.Description) <= 1000 {
			score += 0.9 // Good length
		} else if len(req.Description) > 1000 {
			score += 0.7 // Too long
		} else {
			score += 0.5 // Too short
		}
	}

	// Keywords quality
	if len(req.Keywords) > 0 {
		totalFields++
		keywordQuality := 0.0
		for _, keyword := range req.Keywords {
			if len(keyword) >= 2 && len(keyword) <= 50 {
				keywordQuality += 0.9
			} else {
				keywordQuality += 0.5
			}
		}
		score += keywordQuality / float64(len(req.Keywords))
	}

	// Industry quality
	if req.Industry != "" {
		totalFields++
		score += 0.8 // Industry field is generally reliable
	}

	// Geographic region quality
	if req.GeographicRegion != "" {
		totalFields++
		score += 0.8 // Geographic region is generally reliable
	}

	if totalFields == 0 {
		return 0.0
	}

	return score / float64(totalFields)
}

// calculateDataCompleteness calculates the completeness score of the input data
func (ra *RequestAnalyzer) calculateDataCompleteness(req *shared.BusinessClassificationRequest) float64 {
	fields := 0
	totalFields := 6 // business_name, website_url, description, keywords, industry, geographic_region

	if req.BusinessName != "" {
		fields++
	}
	if req.WebsiteURL != "" {
		fields++
	}
	if req.Description != "" {
		fields++
	}
	if len(req.Keywords) > 0 {
		fields++
	}
	if req.Industry != "" {
		fields++
	}
	if req.GeographicRegion != "" {
		fields++
	}

	return float64(fields) / float64(totalFields)
}

// calculateInputComplexity calculates the complexity score of the input data
func (ra *RequestAnalyzer) calculateInputComplexity(req *shared.BusinessClassificationRequest) float64 {
	complexity := 0.0

	// Business name complexity
	if req.BusinessName != "" {
		complexity += 0.2
		if len(req.BusinessName) > 20 {
			complexity += 0.1 // Longer names are more complex
		}
	}

	// Website URL complexity
	if req.WebsiteURL != "" {
		complexity += 0.3 // Website URLs add significant complexity
	}

	// Description complexity
	if req.Description != "" {
		complexity += 0.2
		if len(req.Description) > 100 {
			complexity += 0.1 // Longer descriptions are more complex
		}
	}

	// Keywords complexity
	if len(req.Keywords) > 0 {
		complexity += 0.1
		if len(req.Keywords) > 5 {
			complexity += 0.1 // More keywords add complexity
		}
	}

	// Industry complexity
	if req.Industry != "" {
		complexity += 0.05
	}

	// Geographic region complexity
	if req.GeographicRegion != "" {
		complexity += 0.05
	}

	// Cap complexity at 1.0
	if complexity > 1.0 {
		complexity = 1.0
	}

	return complexity
}
