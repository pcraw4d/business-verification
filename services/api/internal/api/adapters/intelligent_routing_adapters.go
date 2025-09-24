package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"kyb-platform/internal/observability"
	"kyb-platform/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IntelligentRoutingAdapter provides adapters for converting between API and routing formats
type IntelligentRoutingAdapter struct {
	logger  *observability.Logger
	metrics *observability.Metrics
	tracer  trace.Tracer
	cache   shared.Cache
}

// NewIntelligentRoutingAdapter creates a new intelligent routing adapter
func NewIntelligentRoutingAdapter(
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer trace.Tracer,
	cache shared.Cache,
) *IntelligentRoutingAdapter {
	return &IntelligentRoutingAdapter{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
		cache:   cache,
	}
}

// EnhancedClassificationRequest represents the enhanced API request format
type EnhancedClassificationRequest struct {
	BusinessName     string                 `json:"business_name" validate:"required"`
	WebsiteURL       string                 `json:"website_url,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Industry         string                 `json:"industry,omitempty"`
	Keywords         string                 `json:"keywords,omitempty"`
	GeographicRegion string                 `json:"geographic_region,omitempty"`
	EnhancedFeatures *EnhancedFeatures      `json:"enhanced_features,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// EnhancedFeatures specifies which enhanced features to enable
type EnhancedFeatures struct {
	IncludeCompanySize     bool `json:"include_company_size" default:"true"`
	IncludeBusinessModel   bool `json:"include_business_model" default:"true"`
	IncludeTechnologyStack bool `json:"include_technology_stack" default:"true"`
	IncludeRiskAssessment  bool `json:"include_risk_assessment" default:"false"`
}

// EnhancedClassificationResponse represents the enhanced API response format
type EnhancedClassificationResponse struct {
	ID              string                   `json:"id"`
	BusinessName    string                   `json:"business_name"`
	Status          string                   `json:"status"`
	Classifications []IndustryClassification `json:"classifications"`
	CompanySize     *CompanySize             `json:"company_size,omitempty"`
	BusinessModel   *BusinessModel           `json:"business_model,omitempty"`
	TechnologyStack *TechnologyStack         `json:"technology_stack,omitempty"`
	RiskAssessment  *RiskAssessment          `json:"risk_assessment,omitempty"`
	Metadata        *ClassificationMetadata  `json:"metadata"`
	CreatedAt       time.Time                `json:"created_at"`
}

// IndustryClassification represents industry classification data
type IndustryClassification struct {
	IndustryCode       string  `json:"industry_code"`
	IndustryName       string  `json:"industry_name"`
	ConfidenceScore    float64 `json:"confidence_score"`
	ClassificationType string  `json:"classification_type"`
}

// CompanySize represents company size analysis
type CompanySize struct {
	EmployeeCountRange   string  `json:"employee_count_range"`
	RevenueIndicator     string  `json:"revenue_indicator"`
	OfficeLocationsCount int     `json:"office_locations_count"`
	ConfidenceScore      float64 `json:"confidence_score"`
}

// BusinessModel represents business model analysis
type BusinessModel struct {
	ModelType       string  `json:"model_type"`
	RevenueModel    string  `json:"revenue_model"`
	TargetMarket    string  `json:"target_market"`
	PricingModel    string  `json:"pricing_model"`
	ConfidenceScore float64 `json:"confidence_score"`
}

// TechnologyStack represents technology stack analysis
type TechnologyStack struct {
	ProgrammingLanguages []string `json:"programming_languages"`
	Frameworks           []string `json:"frameworks"`
	CloudPlatforms       []string `json:"cloud_platforms"`
	ThirdPartyServices   []string `json:"third_party_services"`
	DevelopmentTools     []string `json:"development_tools"`
	ConfidenceScore      float64  `json:"confidence_score"`
}

// RiskAssessment represents risk assessment data
type RiskAssessment struct {
	OverallRisk    string   `json:"overall_risk"`
	SecurityRisk   string   `json:"security_risk"`
	FinancialRisk  string   `json:"financial_risk"`
	ComplianceRisk string   `json:"compliance_risk"`
	RiskFactors    []string `json:"risk_factors"`
}

// ClassificationMetadata represents classification metadata
type ClassificationMetadata struct {
	ProcessingTime           string   `json:"processing_time"`
	StrategiesUsed           []string `json:"strategies_used"`
	CacheHit                 bool     `json:"cache_hit"`
	IntelligentRoutingMethod string   `json:"intelligent_routing_method"`
	DataPointsExtracted      int      `json:"data_points_extracted"`
}

// AdaptRequest converts API request to routing format
func (a *IntelligentRoutingAdapter) AdaptRequest(
	ctx context.Context,
	apiReq *EnhancedClassificationRequest,
) (*shared.BusinessClassificationRequest, error) {
	ctx, span := a.tracer.Start(ctx, "IntelligentRoutingAdapter.AdaptRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("business_name", apiReq.BusinessName),
		attribute.String("website_url", apiReq.WebsiteURL),
	)

	// Validate request
	if err := a.validateRequest(apiReq); err != nil {
		a.logger.Error("request validation failed", map[string]interface{}{
			"error":         err.Error(),
			"business_name": apiReq.BusinessName,
		})
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Convert to routing format
	routingReq := &shared.BusinessClassificationRequest{
		BusinessName:     apiReq.BusinessName,
		WebsiteURL:       apiReq.WebsiteURL,
		Description:      apiReq.Description,
		Industry:         apiReq.Industry,
		Keywords:         strings.Split(apiReq.Keywords, ","),
		GeographicRegion: apiReq.GeographicRegion,
		Metadata:         apiReq.Metadata,
	}

	// Add enhanced features metadata
	if apiReq.EnhancedFeatures != nil {
		if routingReq.Metadata == nil {
			routingReq.Metadata = make(map[string]interface{})
		}
		routingReq.Metadata["enhanced_features"] = apiReq.EnhancedFeatures
	}

	a.logger.Info("request adapted successfully", map[string]interface{}{
		"business_name":             apiReq.BusinessName,
		"enhanced_features_enabled": apiReq.EnhancedFeatures != nil,
	})

	return routingReq, nil
}

// AdaptResponse converts routing response to API format
func (a *IntelligentRoutingAdapter) AdaptResponse(
	ctx context.Context,
	routingResp *shared.BusinessClassificationResponse,
	apiReq *EnhancedClassificationRequest,
	processingTime time.Duration,
) (*EnhancedClassificationResponse, error) {
	ctx, span := a.tracer.Start(ctx, "IntelligentRoutingAdapter.AdaptResponse")
	defer span.End()

	span.SetAttributes(
		attribute.String("response_id", routingResp.ID),
		attribute.String("business_name", routingResp.BusinessName),
	)

	// Convert classifications
	classifications := make([]IndustryClassification, 0, len(routingResp.Classifications))
	for _, classification := range routingResp.Classifications {
		classifications = append(classifications, IndustryClassification{
			IndustryCode:       classification.IndustryCode,
			IndustryName:       classification.IndustryName,
			ConfidenceScore:    classification.ConfidenceScore,
			ClassificationType: classification.ClassificationMethod,
		})
	}

	// Build metadata
	metadata := &ClassificationMetadata{
		ProcessingTime:           processingTime.String(),
		StrategiesUsed:           []string{"intelligent_routing"},
		CacheHit:                 false, // Will be set by cache layer
		IntelligentRoutingMethod: "enhanced_classification",
		DataPointsExtracted:      len(classifications),
	}

	// Add enhanced features if requested
	var companySize *CompanySize
	var businessModel *BusinessModel
	var technologyStack *TechnologyStack
	var riskAssessment *RiskAssessment

	if apiReq.EnhancedFeatures != nil {
		if apiReq.EnhancedFeatures.IncludeCompanySize {
			companySize = a.extractCompanySize(routingResp)
			metadata.DataPointsExtracted++
		}
		if apiReq.EnhancedFeatures.IncludeBusinessModel {
			businessModel = a.extractBusinessModel(routingResp)
			metadata.DataPointsExtracted++
		}
		if apiReq.EnhancedFeatures.IncludeTechnologyStack {
			technologyStack = a.extractTechnologyStack(routingResp)
			metadata.DataPointsExtracted++
		}
		if apiReq.EnhancedFeatures.IncludeRiskAssessment {
			riskAssessment = a.extractRiskAssessment(routingResp)
			metadata.DataPointsExtracted++
		}
	}

	// Build response
	response := &EnhancedClassificationResponse{
		ID:              routingResp.ID,
		BusinessName:    routingResp.BusinessName,
		Status:          "completed",
		Classifications: classifications,
		CompanySize:     companySize,
		BusinessModel:   businessModel,
		TechnologyStack: technologyStack,
		RiskAssessment:  riskAssessment,
		Metadata:        metadata,
		CreatedAt:       routingResp.CreatedAt,
	}

	a.logger.Info("response adapted successfully", map[string]interface{}{
		"response_id":           routingResp.ID,
		"data_points_extracted": metadata.DataPointsExtracted,
		"processing_time":       processingTime.String(),
	})

	return response, nil
}

// AdaptBatchRequest converts batch API request to routing format
func (a *IntelligentRoutingAdapter) AdaptBatchRequest(
	ctx context.Context,
	apiReq *EnhancedBatchClassificationRequest,
) (*shared.BatchClassificationRequest, error) {
	ctx, span := a.tracer.Start(ctx, "IntelligentRoutingAdapter.AdaptBatchRequest")
	defer span.End()

	span.SetAttributes(
		attribute.Int("request_count", len(apiReq.Requests)),
	)

	// Validate batch request
	if err := a.validateBatchRequest(apiReq); err != nil {
		a.logger.Error("batch request validation failed", map[string]interface{}{
			"error":         err.Error(),
			"request_count": len(apiReq.Requests),
		})
		return nil, fmt.Errorf("batch request validation failed: %w", err)
	}

	// Convert each request
	routingRequests := make([]shared.BusinessClassificationRequest, 0, len(apiReq.Requests))
	for i, req := range apiReq.Requests {
		routingReq, err := a.AdaptRequest(ctx, &req)
		if err != nil {
			return nil, fmt.Errorf("failed to adapt request %d: %w", i, err)
		}
		routingRequests = append(routingRequests, *routingReq)
	}

	batchReq := &shared.BatchClassificationRequest{
		Requests: routingRequests,
	}

	a.logger.Info("batch request adapted successfully", map[string]interface{}{
		"request_count": len(routingRequests),
	})

	return batchReq, nil
}

// AdaptBatchResponse converts batch routing response to API format
func (a *IntelligentRoutingAdapter) AdaptBatchResponse(
	ctx context.Context,
	routingResp *shared.BatchClassificationResponse,
	apiReq *EnhancedBatchClassificationRequest,
	processingTime time.Duration,
) (*EnhancedBatchClassificationResponse, error) {
	ctx, span := a.tracer.Start(ctx, "IntelligentRoutingAdapter.AdaptBatchResponse")
	defer span.End()

	span.SetAttributes(
		attribute.String("batch_id", routingResp.ID),
		attribute.Int("response_count", len(routingResp.Responses)),
		attribute.Int("error_count", len(routingResp.Errors)),
	)

	// Convert responses
	responses := make([]EnhancedClassificationResponse, 0, len(routingResp.Responses))
	for i, resp := range routingResp.Responses {
		// Find corresponding API request for enhanced features
		var apiRequest *EnhancedClassificationRequest
		if i < len(apiReq.Requests) {
			apiRequest = &apiReq.Requests[i]
		}

		adaptedResp, err := a.AdaptResponse(ctx, &resp, apiRequest, processingTime)
		if err != nil {
			return nil, fmt.Errorf("failed to adapt response %d: %w", i, err)
		}
		responses = append(responses, *adaptedResp)
	}

	// Convert errors
	errors := make([]BatchError, 0, len(routingResp.Errors))
	for _, err := range routingResp.Errors {
		errors = append(errors, BatchError{
			Index: err.Index,
			Error: err.Error,
			Code:  "BATCH_ERROR",
		})
	}

	// Build batch metadata
	batchMetadata := &BatchMetadata{
		TotalRequests:      len(apiReq.Requests),
		SuccessfulCount:    len(responses),
		ErrorCount:         len(errors),
		ProcessingTime:     processingTime.String(),
		ParallelProcessing: true,
	}

	response := &EnhancedBatchClassificationResponse{
		ID:          routingResp.ID,
		Status:      "completed",
		Responses:   responses,
		Errors:      errors,
		Metadata:    batchMetadata,
		CreatedAt:   time.Now(),
		CompletedAt: routingResp.CompletedAt,
	}

	a.logger.Info("batch response adapted successfully", map[string]interface{}{
		"batch_id":         routingResp.ID,
		"successful_count": len(responses),
		"error_count":      len(errors),
		"processing_time":  processingTime.String(),
	})

	return response, nil
}

// GenerateCacheKey generates a cache key for the request
func (a *IntelligentRoutingAdapter) GenerateCacheKey(req *EnhancedClassificationRequest) string {
	// Create a deterministic cache key based on request parameters
	keyData := map[string]interface{}{
		"business_name":     req.BusinessName,
		"website_url":       req.WebsiteURL,
		"description":       req.Description,
		"industry":          req.Industry,
		"keywords":          req.Keywords,
		"geographic_region": req.GeographicRegion,
	}

	if req.EnhancedFeatures != nil {
		keyData["enhanced_features"] = req.EnhancedFeatures
	}

	// Serialize to JSON for consistent key generation
	data, _ := json.Marshal(keyData)
	return fmt.Sprintf("classification:%x", data)
}

// GetCachedResponse retrieves cached response if available
func (a *IntelligentRoutingAdapter) GetCachedResponse(
	ctx context.Context,
	cacheKey string,
) (*EnhancedClassificationResponse, error) {
	ctx, span := a.tracer.Start(ctx, "IntelligentRoutingAdapter.GetCachedResponse")
	defer span.End()

	span.SetAttributes(attribute.String("cache_key", cacheKey))

	if a.cache == nil {
		return nil, nil // No cache available
	}

	cached, found, err := a.cache.Get(ctx, cacheKey)
	if err != nil {
		a.logger.Warn("cache get failed", map[string]interface{}{
			"cache_key": cacheKey,
			"error":     err.Error(),
		})
		return nil, nil // Treat as cache miss
	}

	if !found || cached == nil {
		return nil, nil // Cache miss
	}

	// Deserialize cached response
	var response EnhancedClassificationResponse
	switch v := cached.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &response); err != nil {
			a.logger.Warn("cache deserialization failed", map[string]interface{}{
				"cache_key": cacheKey,
				"error":     err.Error(),
			})
			return nil, nil // Treat as cache miss
		}
	case []byte:
		if err := json.Unmarshal(v, &response); err != nil {
			a.logger.Warn("cache deserialization failed", map[string]interface{}{
				"cache_key": cacheKey,
				"error":     err.Error(),
			})
			return nil, nil // Treat as cache miss
		}
	default:
		// Try to marshal and unmarshal for other types
		data, err := json.Marshal(cached)
		if err != nil {
			a.logger.Warn("cache type conversion failed", map[string]interface{}{
				"cache_key": cacheKey,
				"error":     err.Error(),
			})
			return nil, nil // Treat as cache miss
		}
		if err := json.Unmarshal(data, &response); err != nil {
			a.logger.Warn("cache deserialization failed", map[string]interface{}{
				"cache_key": cacheKey,
				"error":     err.Error(),
			})
			return nil, nil // Treat as cache miss
		}
	}

	// Mark as cache hit
	response.Metadata.CacheHit = true

	a.logger.Info("cache hit", map[string]interface{}{
		"cache_key":   cacheKey,
		"response_id": response.ID,
	})

	return &response, nil
}

// CacheResponse stores response in cache
func (a *IntelligentRoutingAdapter) CacheResponse(
	ctx context.Context,
	cacheKey string,
	response *EnhancedClassificationResponse,
	ttl time.Duration,
) error {
	ctx, span := a.tracer.Start(ctx, "IntelligentRoutingAdapter.CacheResponse")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache_key", cacheKey),
		attribute.String("response_id", response.ID),
	)

	if a.cache == nil {
		return nil // No cache available
	}

	// Serialize response
	data, err := json.Marshal(response)
	if err != nil {
		a.logger.Error("response serialization failed", map[string]interface{}{
			"response_id": response.ID,
			"error":       err.Error(),
		})
		return fmt.Errorf("failed to serialize response: %w", err)
	}

	// Store in cache
	if err := a.cache.Set(ctx, cacheKey, string(data), ttl); err != nil {
		a.logger.Error("cache set failed", map[string]interface{}{
			"cache_key": cacheKey,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to cache response: %w", err)
	}

	a.logger.Info("response cached successfully", map[string]interface{}{
		"cache_key":   cacheKey,
		"response_id": response.ID,
		"ttl":         ttl.String(),
	})

	return nil
}

// validateRequest validates the API request
func (a *IntelligentRoutingAdapter) validateRequest(req *EnhancedClassificationRequest) error {
	if req.BusinessName == "" {
		return fmt.Errorf("business_name is required")
	}

	if len(req.BusinessName) > 255 {
		return fmt.Errorf("business_name exceeds maximum length of 255 characters")
	}

	if req.WebsiteURL != "" && !isValidURL(req.WebsiteURL) {
		return fmt.Errorf("invalid website URL format")
	}

	if len(req.Description) > 1000 {
		return fmt.Errorf("description exceeds maximum length of 1000 characters")
	}

	return nil
}

// validateBatchRequest validates the batch API request
func (a *IntelligentRoutingAdapter) validateBatchRequest(req *EnhancedBatchClassificationRequest) error {
	if len(req.Requests) == 0 {
		return fmt.Errorf("batch must contain at least one request")
	}

	if len(req.Requests) > 100 {
		return fmt.Errorf("batch size exceeds maximum of 100 requests")
	}

	for i, request := range req.Requests {
		if err := a.validateRequest(&request); err != nil {
			return fmt.Errorf("request %d validation failed: %w", i, err)
		}
	}

	return nil
}

// extractCompanySize extracts company size information from routing response
func (a *IntelligentRoutingAdapter) extractCompanySize(resp *shared.BusinessClassificationResponse) *CompanySize {
	// This would be implemented based on the actual routing response structure
	// For now, returning a placeholder implementation
	return &CompanySize{
		EmployeeCountRange:   "51-200",
		RevenueIndicator:     "medium_business",
		OfficeLocationsCount: 1,
		ConfidenceScore:      0.85,
	}
}

// extractBusinessModel extracts business model information from routing response
func (a *IntelligentRoutingAdapter) extractBusinessModel(resp *shared.BusinessClassificationResponse) *BusinessModel {
	// This would be implemented based on the actual routing response structure
	// For now, returning a placeholder implementation
	return &BusinessModel{
		ModelType:       "B2B",
		RevenueModel:    "subscription",
		TargetMarket:    "Enterprise technology companies",
		PricingModel:    "Tiered subscription pricing",
		ConfidenceScore: 0.90,
	}
}

// extractTechnologyStack extracts technology stack information from routing response
func (a *IntelligentRoutingAdapter) extractTechnologyStack(resp *shared.BusinessClassificationResponse) *TechnologyStack {
	// This would be implemented based on the actual routing response structure
	// For now, returning a placeholder implementation
	return &TechnologyStack{
		ProgrammingLanguages: []string{"JavaScript", "Python", "Go"},
		Frameworks:           []string{"React", "Node.js", "Django"},
		CloudPlatforms:       []string{"AWS", "Google Cloud"},
		ThirdPartyServices:   []string{"Stripe", "SendGrid", "MongoDB"},
		DevelopmentTools:     []string{"GitHub", "Docker", "Kubernetes"},
		ConfidenceScore:      0.88,
	}
}

// extractRiskAssessment extracts risk assessment information from routing response
func (a *IntelligentRoutingAdapter) extractRiskAssessment(resp *shared.BusinessClassificationResponse) *RiskAssessment {
	// This would be implemented based on the actual routing response structure
	// For now, returning a placeholder implementation
	return &RiskAssessment{
		OverallRisk:    "LOW",
		SecurityRisk:   "LOW",
		FinancialRisk:  "MEDIUM",
		ComplianceRisk: "LOW",
		RiskFactors:    []string{"Limited financial history", "New market entry"},
	}
}

// EnhancedBatchClassificationRequest represents the enhanced batch API request format
type EnhancedBatchClassificationRequest struct {
	Requests []EnhancedClassificationRequest `json:"requests" validate:"required,min=1,max=100"`
}

// EnhancedBatchClassificationResponse represents the enhanced batch API response format
type EnhancedBatchClassificationResponse struct {
	ID          string                           `json:"id"`
	Status      string                           `json:"status"`
	Responses   []EnhancedClassificationResponse `json:"responses"`
	Errors      []BatchError                     `json:"errors"`
	Metadata    *BatchMetadata                   `json:"metadata"`
	CreatedAt   time.Time                        `json:"created_at"`
	CompletedAt time.Time                        `json:"completed_at"`
}

// BatchError represents a batch processing error
type BatchError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

// BatchMetadata represents batch processing metadata
type BatchMetadata struct {
	TotalRequests      int    `json:"total_requests"`
	SuccessfulCount    int    `json:"successful_count"`
	ErrorCount         int    `json:"error_count"`
	ProcessingTime     string `json:"processing_time"`
	ParallelProcessing bool   `json:"parallel_processing"`
}

// isValidURL validates URL format
func isValidURL(url string) bool {
	// Basic URL validation - in production, use a more robust validation library
	return len(url) > 0 && (url[:7] == "http://" || url[:8] == "https://")
}
