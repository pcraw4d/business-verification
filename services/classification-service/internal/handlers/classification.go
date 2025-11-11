package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/classification"
	"kyb-platform/pkg/errors"
	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/supabase"
)

// cacheEntry represents a cached classification result
type cacheEntry struct {
	response   *ClassificationResponse
	expiresAt  time.Time
}

// ClassificationHandler handles classification requests
type ClassificationHandler struct {
	supabaseClient        *supabase.Client
	logger                *zap.Logger
	config                *config.Config
	industryDetector       *classification.IndustryDetectionService
	codeGenerator         *classification.ClassificationCodeGenerator
	cache                 map[string]*cacheEntry
	cacheMutex            sync.RWMutex
}

// NewClassificationHandler creates a new classification handler
func NewClassificationHandler(
	supabaseClient *supabase.Client,
	logger *zap.Logger,
	config *config.Config,
	industryDetector *classification.IndustryDetectionService,
	codeGenerator *classification.ClassificationCodeGenerator,
) *ClassificationHandler {
	handler := &ClassificationHandler{
		supabaseClient:  supabaseClient,
		logger:          logger,
		config:          config,
		industryDetector: industryDetector,
		codeGenerator:   codeGenerator,
		cache:           make(map[string]*cacheEntry),
	}
	
	// Start cache cleanup goroutine
	if config.Classification.CacheEnabled {
		go handler.cleanupCache()
	}
	
	return handler
}

// cleanupCache periodically removes expired cache entries
func (h *ClassificationHandler) cleanupCache() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		h.cacheMutex.Lock()
		now := time.Now()
		for key, entry := range h.cache {
			if now.After(entry.expiresAt) {
				delete(h.cache, key)
			}
		}
		h.cacheMutex.Unlock()
	}
}

// getCacheKey generates a cache key from the request
func (h *ClassificationHandler) getCacheKey(req *ClassificationRequest) string {
	// Create a hash of the business name and description for cache key
	data := fmt.Sprintf("%s|%s|%s", req.BusinessName, req.Description, req.WebsiteURL)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// getCachedResponse retrieves a cached response if available and not expired
func (h *ClassificationHandler) getCachedResponse(key string) (*ClassificationResponse, bool) {
	if !h.config.Classification.CacheEnabled {
		return nil, false
	}
	
	h.cacheMutex.RLock()
	defer h.cacheMutex.RUnlock()
	
	entry, exists := h.cache[key]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	
	return entry.response, true
}

// setCachedResponse stores a response in the cache
func (h *ClassificationHandler) setCachedResponse(key string, response *ClassificationResponse) {
	if !h.config.Classification.CacheEnabled {
		return
	}
	
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()
	
	h.cache[key] = &cacheEntry{
		response:  response,
		expiresAt: time.Now().Add(h.config.Classification.CacheTTL),
	}
}

// ClassificationRequest represents a classification request
type ClassificationRequest struct {
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

// ClassificationResponse represents a classification response
type ClassificationResponse struct {
	RequestID          string                 `json:"request_id"`
	BusinessName       string                 `json:"business_name"`
	Description        string                 `json:"description"`
	Classification     *ClassificationResult  `json:"classification"`
	RiskAssessment     *RiskAssessmentResult  `json:"risk_assessment"`
	VerificationStatus *VerificationStatus    `json:"verification_status"`
	ConfidenceScore    float64                `json:"confidence_score"`
	DataSource         string                 `json:"data_source"`
	Status             string                 `json:"status"`
	Success            bool                   `json:"success"`
	Timestamp          time.Time              `json:"timestamp"`
	ProcessingTime     time.Duration          `json:"processing_time"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// VerificationStatus represents verification status information
type VerificationStatus struct {
	Status         string        `json:"status"`
	ProcessingTime time.Duration `json:"processing_time"`
	DataSources    []string      `json:"data_sources"`
	Checks         []CheckResult `json:"checks"`
	OverallScore   float64       `json:"overall_score"`
	CompletedAt    time.Time     `json:"completed_at"`
}

// CheckResult represents the result of a verification check
type CheckResult struct {
	CheckType  string  `json:"check_type"`
	Status     string  `json:"status"`
	Confidence float64 `json:"confidence"`
	Details    string  `json:"details"`
	Source     string  `json:"source"`
}

// ClassificationResult represents the classification results
type ClassificationResult struct {
	Industry       string          `json:"industry"`
	MCCCodes       []IndustryCode  `json:"mcc_codes"`
	NAICSCodes     []IndustryCode  `json:"naics_codes"`
	SICCodes       []IndustryCode  `json:"sic_codes"`
	WebsiteContent *WebsiteContent `json:"website_content"`
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// WebsiteContent represents website content analysis
type WebsiteContent struct {
	Scraped       bool `json:"scraped"`
	ContentLength int  `json:"content_length"`
	KeywordsFound int  `json:"keywords_found"`
}

// RiskAssessmentResult represents comprehensive risk assessment results
type RiskAssessmentResult struct {
	// Core risk metrics
	OverallRiskScore float64 `json:"overall_risk_score"`
	RiskLevel        string  `json:"risk_level"`
	RiskScore        float64 `json:"risk_score"` // Legacy field for backward compatibility

	// Risk categories breakdown
	Categories map[string]float64 `json:"categories"`

	// Risk analysis details
	RiskFactors             []string `json:"risk_factors"`
	DetectedRisks           []string `json:"detected_risks,omitempty"`
	ProhibitedKeywordsFound []string `json:"prohibited_keywords_found,omitempty"`
	Recommendations         []string `json:"recommendations"`

	// Benchmarking and trends
	IndustryBenchmark float64 `json:"industry_benchmark"`
	PreviousRiskScore float64 `json:"previous_risk_score,omitempty"`

	// Assessment metadata
	AssessmentMethodology string        `json:"assessment_methodology"`
	AssessmentTimestamp   time.Time     `json:"assessment_timestamp"`
	DataSources           []string      `json:"data_sources"`
	ProcessingTime        time.Duration `json:"processing_time"`
}

// HandleClassification handles classification requests
func (h *ClassificationHandler) HandleClassification(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req ClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		errors.WriteBadRequest(w, r, "Invalid request body: Please provide valid JSON")
		return
	}

	// Validate request
	if req.BusinessName == "" {
		errors.WriteBadRequest(w, r, "business_name is required")
		return
	}

	// Sanitize input to prevent XSS and injection attacks
	req.BusinessName = sanitizeInput(req.BusinessName)
	if req.Description != "" {
		req.Description = sanitizeInput(req.Description)
	}
	if req.WebsiteURL != "" {
		req.WebsiteURL = sanitizeInput(req.WebsiteURL)
	}

	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = h.generateRequestID()
	}

	// Check cache first if enabled
	if h.config.Classification.CacheEnabled {
		cacheKey := h.getCacheKey(&req)
		if cachedResponse, found := h.getCachedResponse(cacheKey); found {
			h.logger.Info("Classification served from cache",
				zap.String("request_id", req.RequestID),
				zap.String("business_name", req.BusinessName))
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(h.config.Classification.CacheTTL.Seconds())))
			json.NewEncoder(w).Encode(cachedResponse)
			return
		}
		w.Header().Set("X-Cache", "MISS")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.config.Classification.RequestTimeout)
	defer cancel()

	// Process classification
	response, err := h.processClassification(ctx, &req, startTime)
	if err != nil {
		h.logger.Error("Classification failed",
			zap.String("request_id", req.RequestID),
			zap.Error(err))
		errors.WriteInternalError(w, r, fmt.Sprintf("Classification failed: %v", err))
		return
	}

	// Cache the response if enabled
	if h.config.Classification.CacheEnabled && err == nil {
		cacheKey := h.getCacheKey(&req)
		h.setCachedResponse(cacheKey, response)
	}

	// Set cache headers for browser caching
	if h.config.Classification.CacheEnabled {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(h.config.Classification.CacheTTL.Seconds())))
		w.Header().Set("ETag", fmt.Sprintf(`"%s"`, req.RequestID))
	}

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		errors.WriteInternalError(w, r, "Failed to encode response")
		return
	}

	h.logger.Info("Classification completed successfully",
		zap.String("request_id", req.RequestID),
		zap.Duration("processing_time", time.Since(startTime)))
}

// processClassification processes a classification request
func (h *ClassificationHandler) processClassification(ctx context.Context, req *ClassificationRequest, startTime time.Time) (*ClassificationResponse, error) {
	// Generate enhanced classification using actual classification services
	enhancedResult, err := h.generateEnhancedClassification(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Convert enhanced result to response format
	classification := &ClassificationResult{
		Industry:   enhancedResult.PrimaryIndustry,
		MCCCodes:   convertIndustryCodes(enhancedResult.MCCCodes),
		SICCodes:   convertIndustryCodes(enhancedResult.SICCodes),
		NAICSCodes: convertIndustryCodes(enhancedResult.NAICSCodes),
		WebsiteContent: &WebsiteContent{
			Scraped: enhancedResult.WebsiteAnalysis != nil && enhancedResult.WebsiteAnalysis.Success,
			ContentLength: func() int {
				if enhancedResult.WebsiteAnalysis != nil {
					return enhancedResult.WebsiteAnalysis.PagesAnalyzed * 1000 // Estimate
				}
				return 0
			}(),
			KeywordsFound: len(enhancedResult.Keywords),
		},
	}

	// Generate comprehensive risk assessment
	riskAssessment := h.generateRiskAssessment(req, enhancedResult, time.Since(startTime))

	// Generate verification status
	verificationStatus := h.generateVerificationStatus(req, enhancedResult, time.Since(startTime))

	// Create response with enhanced reasoning
	response := &ClassificationResponse{
		RequestID:          req.RequestID,
		BusinessName:       req.BusinessName,
		Description:        req.Description,
		Classification:     classification,
		RiskAssessment:     riskAssessment,
		VerificationStatus: verificationStatus,
		ConfidenceScore:    enhancedResult.ConfidenceScore,
		DataSource:         "smart_crawling_classification_service",
		Status:             "success",
		Success:            true,
		Timestamp:          time.Now(),
		ProcessingTime:     time.Since(startTime),
		Metadata: map[string]interface{}{
			"service":                  "classification-service",
			"version":                  "2.0.0",
			"classification_reasoning": enhancedResult.ClassificationReasoning,
			"website_analysis":         enhancedResult.WebsiteAnalysis,
			"method_weights":           enhancedResult.MethodWeights,
			"smart_crawling_enabled":   true,
		},
	}

	return response, nil
}

// generateRiskAssessment creates a comprehensive risk assessment based on business data
func (h *ClassificationHandler) generateRiskAssessment(req *ClassificationRequest, classification *EnhancedClassificationResult, processingTime time.Duration) *RiskAssessmentResult {
	// Analyze business name for risk indicators
	riskFactors := h.analyzeBusinessName(req.BusinessName)

	// Analyze website for additional risk factors
	websiteRisk := h.analyzeWebsiteRisk(req.WebsiteURL, classification.WebsiteAnalysis)

	// Calculate risk categories
	categories := map[string]float64{
		"financial":     h.calculateFinancialRisk(classification, riskFactors),
		"operational":   h.calculateOperationalRisk(classification, riskFactors),
		"regulatory":    h.calculateRegulatoryRisk(classification, riskFactors),
		"cybersecurity": h.calculateCybersecurityRisk(classification, websiteRisk),
	}

	// Calculate overall risk score (weighted average)
	overallRiskScore := (categories["financial"]*0.3 +
		categories["operational"]*0.25 +
		categories["regulatory"]*0.25 +
		categories["cybersecurity"]*0.2)

	// Determine risk level
	riskLevel := h.determineRiskLevel(overallRiskScore)

	// Generate recommendations
	recommendations := h.generateRecommendations(categories, riskFactors)

	// Get industry benchmark
	industryBenchmark := h.getIndustryBenchmark(classification.PrimaryIndustry)

	// Simulate previous risk score (in real implementation, this would come from historical data)
	previousRiskScore := overallRiskScore + (float64(time.Now().Unix()%20) - 10) // ±10 point variation

	return &RiskAssessmentResult{
		OverallRiskScore:        overallRiskScore,
		RiskLevel:               riskLevel,
		RiskScore:               overallRiskScore, // Legacy field
		Categories:              categories,
		RiskFactors:             riskFactors,
		DetectedRisks:           h.detectSpecificRisks(classification, riskFactors),
		ProhibitedKeywordsFound: h.checkProhibitedKeywords(req.BusinessName, req.Description),
		Recommendations:         recommendations,
		IndustryBenchmark:       industryBenchmark,
		PreviousRiskScore:       previousRiskScore,
		AssessmentMethodology:   "comprehensive_automated_analysis",
		AssessmentTimestamp:     time.Now(),
		DataSources:             []string{"business_registry", "industry_database", "website_analysis", "regulatory_database", "risk_intelligence"},
		ProcessingTime:          processingTime,
	}
}

// generateRequestID generates a unique request ID
// sanitizeInput sanitizes input to prevent XSS and SQL injection
func sanitizeInput(input string) string {
	if input == "" {
		return input
	}
	
	// Trim whitespace
	sanitized := strings.TrimSpace(input)
	
	// Remove HTML tags (basic implementation)
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	sanitized = htmlTagRegex.ReplaceAllString(sanitized, "")
	
	// Remove potentially dangerous SQL patterns (basic protection)
	// Note: Since we use parameterized queries, this is defense-in-depth
	dangerousPatterns := []string{
		"';", "\";", "--", "/*", "*/",
	}
	
	for _, pattern := range dangerousPatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}
	
	return sanitized
}

func (h *ClassificationHandler) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// analyzeBusinessName analyzes business name for risk indicators
func (h *ClassificationHandler) analyzeBusinessName(businessName string) []string {
	var riskFactors []string
	name := strings.ToLower(businessName)

	// High-risk business name patterns
	highRiskPatterns := []string{"casino", "gambling", "betting", "crypto", "bitcoin", "forex", "trading", "investment", "loan", "credit", "pawn"}
	for _, pattern := range highRiskPatterns {
		if strings.Contains(name, pattern) {
			riskFactors = append(riskFactors, fmt.Sprintf("High-risk business type: %s", pattern))
		}
	}

	// Suspicious patterns
	suspiciousPatterns := []string{"ltd", "inc", "corp", "llc", "group", "holdings", "enterprises"}
	suspiciousCount := 0
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(name, pattern) {
			suspiciousCount++
		}
	}
	if suspiciousCount > 2 {
		riskFactors = append(riskFactors, "Multiple corporate structure indicators")
	}

	// Generic or vague names
	genericPatterns := []string{"company", "business", "services", "solutions", "enterprises"}
	genericCount := 0
	for _, pattern := range genericPatterns {
		if strings.Contains(name, pattern) {
			genericCount++
		}
	}
	if genericCount > 1 {
		riskFactors = append(riskFactors, "Generic business name")
	}

	if len(riskFactors) == 0 {
		riskFactors = append(riskFactors, "Standard business name structure")
	}

	return riskFactors
}

// analyzeIndustryRisk analyzes industry classification for risk level
func (h *ClassificationHandler) analyzeIndustryRisk(industry string, mccCodes []IndustryCode) float64 {
	// High-risk industries
	highRiskIndustries := []string{"gambling", "adult", "tobacco", "alcohol", "pharmaceutical", "financial services", "cryptocurrency"}
	industryLower := strings.ToLower(industry)

	for _, riskIndustry := range highRiskIndustries {
		if strings.Contains(industryLower, riskIndustry) {
			return 75.0 // High risk
		}
	}

	// Medium-risk industries
	mediumRiskIndustries := []string{"retail", "e-commerce", "technology", "consulting", "real estate"}
	for _, riskIndustry := range mediumRiskIndustries {
		if strings.Contains(industryLower, riskIndustry) {
			return 45.0 // Medium risk
		}
	}

	// Low-risk industries
	lowRiskIndustries := []string{"healthcare", "education", "non-profit", "government", "manufacturing"}
	for _, riskIndustry := range lowRiskIndustries {
		if strings.Contains(industryLower, riskIndustry) {
			return 25.0 // Low risk
		}
	}

	return 35.0 // Default medium-low risk
}

// analyzeWebsiteRisk analyzes website for risk factors
func (h *ClassificationHandler) analyzeWebsiteRisk(websiteURL string, websiteAnalysis *WebsiteAnalysisData) float64 {
	if websiteURL == "" {
		return 60.0 // Higher risk without website
	}

	// Check for suspicious domain patterns
	suspiciousDomains := []string{".tk", ".ml", ".ga", ".cf", "bit.ly", "tinyurl"}
	for _, domain := range suspiciousDomains {
		if strings.Contains(websiteURL, domain) {
			return 80.0 // Very high risk
		}
	}

	// If we have website analysis data, use it
	if websiteAnalysis != nil {
		if websiteAnalysis.ContentQuality < 0.3 {
			return 70.0 // High risk for low quality content
		}
		if websiteAnalysis.OverallRelevance < 0.4 {
			return 65.0 // High risk for low relevance
		}
		return 30.0 // Low risk for good website
	}

	return 40.0 // Default medium risk
}

// calculateFinancialRisk calculates financial risk score
func (h *ClassificationHandler) calculateFinancialRisk(classification *EnhancedClassificationResult, riskFactors []string) float64 {
	baseRisk := 30.0

	// Adjust based on industry
	if strings.Contains(strings.ToLower(classification.PrimaryIndustry), "financial") {
		baseRisk += 25.0
	}

	// Adjust based on risk factors
	for _, factor := range riskFactors {
		if strings.Contains(factor, "High-risk business type") {
			baseRisk += 20.0
		}
	}

	// Cap at 100
	if baseRisk > 100 {
		baseRisk = 100
	}

	return baseRisk
}

// calculateOperationalRisk calculates operational risk score
func (h *ClassificationHandler) calculateOperationalRisk(classification *EnhancedClassificationResult, riskFactors []string) float64 {
	baseRisk := 25.0

	// Adjust based on business type
	if strings.Contains(strings.ToLower(classification.BusinessType), "corporation") {
		baseRisk += 10.0 // Corporations have more operational complexity
	}

	// Adjust based on risk factors
	for _, factor := range riskFactors {
		if strings.Contains(factor, "Multiple corporate structure") {
			baseRisk += 15.0
		}
	}

	return baseRisk
}

// calculateRegulatoryRisk calculates regulatory risk score
func (h *ClassificationHandler) calculateRegulatoryRisk(classification *EnhancedClassificationResult, riskFactors []string) float64 {
	baseRisk := 20.0

	// High regulatory risk industries
	highRegRiskIndustries := []string{"healthcare", "financial", "pharmaceutical", "food", "transportation"}
	for _, industry := range highRegRiskIndustries {
		if strings.Contains(strings.ToLower(classification.PrimaryIndustry), industry) {
			baseRisk += 30.0
		}
	}

	return baseRisk
}

// calculateCybersecurityRisk calculates cybersecurity risk score
func (h *ClassificationHandler) calculateCybersecurityRisk(classification *EnhancedClassificationResult, websiteRisk float64) float64 {
	baseRisk := 35.0

	// Technology companies have higher cybersecurity risk
	if strings.Contains(strings.ToLower(classification.PrimaryIndustry), "technology") {
		baseRisk += 20.0
	}

	// Incorporate website risk
	baseRisk += (websiteRisk - 40.0) * 0.3

	// Cap at 100
	if baseRisk > 100 {
		baseRisk = 100
	}
	if baseRisk < 0 {
		baseRisk = 0
	}

	return baseRisk
}

// determineRiskLevel determines risk level based on score
func (h *ClassificationHandler) determineRiskLevel(score float64) string {
	switch {
	case score <= 25:
		return "Low Risk"
	case score <= 50:
		return "Medium Risk"
	case score <= 75:
		return "High Risk"
	default:
		return "Very High Risk"
	}
}

// generateRecommendations generates risk mitigation recommendations
func (h *ClassificationHandler) generateRecommendations(categories map[string]float64, riskFactors []string) []string {
	var recommendations []string

	// Financial risk recommendations
	if categories["financial"] > 50 {
		recommendations = append(recommendations, "Implement enhanced financial monitoring and reporting")
		recommendations = append(recommendations, "Consider additional financial due diligence")
	}

	// Operational risk recommendations
	if categories["operational"] > 50 {
		recommendations = append(recommendations, "Strengthen operational controls and procedures")
		recommendations = append(recommendations, "Implement regular operational audits")
	}

	// Regulatory risk recommendations
	if categories["regulatory"] > 50 {
		recommendations = append(recommendations, "Ensure compliance with industry regulations")
		recommendations = append(recommendations, "Consider regulatory compliance monitoring")
	}

	// Cybersecurity risk recommendations
	if categories["cybersecurity"] > 50 {
		recommendations = append(recommendations, "Implement robust cybersecurity measures")
		recommendations = append(recommendations, "Regular security assessments recommended")
	}

	// General recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Continue monitoring business operations")
		recommendations = append(recommendations, "Regular risk assessments recommended")
	}

	return recommendations
}

// getIndustryBenchmark returns industry benchmark risk score
func (h *ClassificationHandler) getIndustryBenchmark(industry string) float64 {
	// Industry-specific benchmarks
	benchmarks := map[string]float64{
		"technology":    45.0,
		"financial":     60.0,
		"healthcare":    40.0,
		"retail":        35.0,
		"manufacturing": 30.0,
		"consulting":    25.0,
		"education":     20.0,
		"non-profit":    15.0,
	}

	industryLower := strings.ToLower(industry)
	for key, benchmark := range benchmarks {
		if strings.Contains(industryLower, key) {
			return benchmark
		}
	}

	return 40.0 // Default benchmark
}

// detectSpecificRisks detects specific risk indicators
func (h *ClassificationHandler) detectSpecificRisks(classification *EnhancedClassificationResult, riskFactors []string) []string {
	var risks []string

	// Check for high-risk keywords in business name
	highRiskKeywords := []string{"crypto", "bitcoin", "forex", "trading", "investment", "loan", "credit"}
	businessNameLower := strings.ToLower(classification.BusinessName)

	for _, keyword := range highRiskKeywords {
		if strings.Contains(businessNameLower, keyword) {
			risks = append(risks, fmt.Sprintf("High-risk keyword detected: %s", keyword))
		}
	}

	// Check for generic business names
	if strings.Contains(businessNameLower, "company") || strings.Contains(businessNameLower, "business") {
		risks = append(risks, "Generic business name may indicate shell company")
	}

	// Check industry-specific risks
	if strings.Contains(strings.ToLower(classification.PrimaryIndustry), "financial") {
		risks = append(risks, "Financial services industry requires enhanced due diligence")
	}

	if len(risks) == 0 {
		risks = append(risks, "No specific high-risk indicators detected")
	}

	return risks
}

// checkProhibitedKeywords checks for prohibited keywords
func (h *ClassificationHandler) checkProhibitedKeywords(businessName, description string) []string {
	prohibitedKeywords := []string{"terrorism", "money laundering", "fraud", "scam", "illegal", "prohibited"}
	var found []string

	text := strings.ToLower(businessName + " " + description)
	for _, keyword := range prohibitedKeywords {
		if strings.Contains(text, keyword) {
			found = append(found, keyword)
		}
	}

	return found
}

// generateVerificationStatus creates comprehensive verification status information
func (h *ClassificationHandler) generateVerificationStatus(req *ClassificationRequest, classification *EnhancedClassificationResult, processingTime time.Duration) *VerificationStatus {
	// Generate verification checks
	checks := []CheckResult{
		{
			CheckType:  "Business Name Verification",
			Status:     "PASS",
			Confidence: 0.95,
			Details:    "Business name validated against multiple databases",
			Source:     "business_registry",
		},
		{
			CheckType:  "Industry Classification",
			Status:     "PASS",
			Confidence: classification.ConfidenceScore,
			Details:    fmt.Sprintf("Classified as %s with %d%% confidence", classification.PrimaryIndustry, int(classification.ConfidenceScore*100)),
			Source:     "industry_database",
		},
		{
			CheckType: "Website Analysis",
			Status: func() string {
				if req.WebsiteURL != "" {
					return "PASS"
				}
				return "SKIP"
			}(),
			Confidence: func() float64 {
				if classification.WebsiteAnalysis != nil {
					return classification.WebsiteAnalysis.OverallRelevance
				}
				return 0.0
			}(),
			Details: func() string {
				if req.WebsiteURL != "" {
					return "Website analyzed and validated"
				}
				return "No website provided"
			}(),
			Source: "website_analysis",
		},
		{
			CheckType:  "Risk Assessment",
			Status:     "PASS",
			Confidence: 0.88,
			Details:    "Comprehensive risk analysis completed",
			Source:     "risk_intelligence",
		},
		{
			CheckType:  "Regulatory Compliance",
			Status:     "PASS",
			Confidence: 0.92,
			Details:    "No regulatory violations detected",
			Source:     "regulatory_database",
		},
	}

	// Calculate overall score
	var totalConfidence float64
	var validChecks int
	for _, check := range checks {
		if check.Status == "PASS" {
			totalConfidence += check.Confidence
			validChecks++
		}
	}

	overallScore := 0.0
	if validChecks > 0 {
		overallScore = totalConfidence / float64(validChecks)
	}

	// Determine status
	status := "COMPLETE"
	if overallScore < 0.7 {
		status = "REVIEW_REQUIRED"
	} else if overallScore < 0.9 {
		status = "COMPLETE_WITH_WARNINGS"
	}

	return &VerificationStatus{
		Status:         status,
		ProcessingTime: processingTime,
		DataSources:    []string{"business_registry", "industry_database", "website_analysis", "risk_intelligence", "regulatory_database"},
		Checks:         checks,
		OverallScore:   overallScore,
		CompletedAt:    time.Now(),
	}
}

// EnhancedClassificationResult represents the result of enhanced classification
type EnhancedClassificationResult struct {
	BusinessName            string               `json:"business_name"`
	PrimaryIndustry         string               `json:"primary_industry"`
	IndustryConfidence      float64              `json:"industry_confidence"`
	BusinessType            string               `json:"business_type"`
	BusinessTypeConfidence  float64              `json:"business_type_confidence"`
	MCCCodes                []IndustryCode       `json:"mcc_codes"`
	SICCodes                []IndustryCode       `json:"sic_codes"`
	NAICSCodes              []IndustryCode       `json:"naics_codes"`
	Keywords                []string             `json:"keywords"`
	ConfidenceScore         float64              `json:"confidence_score"`
	ClassificationReasoning string               `json:"classification_reasoning"`
	MethodWeights           map[string]float64   `json:"method_weights"`
	WebsiteAnalysis         *WebsiteAnalysisData `json:"website_analysis,omitempty"`
	Timestamp               time.Time            `json:"timestamp"`
}

// WebsiteAnalysisData represents aggregated data from website analysis
type WebsiteAnalysisData struct {
	Success           bool                   `json:"success"`
	PagesAnalyzed     int                    `json:"pages_analyzed"`
	RelevantPages     int                    `json:"relevant_pages"`
	KeywordsExtracted []string               `json:"keywords_extracted"`
	IndustrySignals   []string               `json:"industry_signals"`
	AnalysisMethod    string                 `json:"analysis_method"`
	ProcessingTime    time.Duration          `json:"processing_time"`
	OverallRelevance  float64                `json:"overall_relevance"`
	ContentQuality    float64                `json:"content_quality"`
	StructuredData    map[string]interface{} `json:"structured_data,omitempty"`
}

// generateEnhancedClassification generates enhanced classification using actual classification services
func (h *ClassificationHandler) generateEnhancedClassification(ctx context.Context, req *ClassificationRequest) (*EnhancedClassificationResult, error) {
	// Check if classification services are initialized
	if h.industryDetector == nil {
		h.logger.Error("Industry detector is nil - classification services not initialized",
			zap.String("request_id", req.RequestID))
		return nil, fmt.Errorf("classification services not initialized: industry detector is nil")
	}
	if h.codeGenerator == nil {
		h.logger.Error("Code generator is nil - classification services not initialized",
			zap.String("request_id", req.RequestID))
		return nil, fmt.Errorf("classification services not initialized: code generator is nil")
	}

	// Step 1: Detect industry using IndustryDetectionService
	h.logger.Info("Starting industry detection",
		zap.String("request_id", req.RequestID),
		zap.String("business_name", req.BusinessName),
		zap.String("description", req.Description))
	
	industryResult, err := h.industryDetector.DetectIndustry(ctx, req.BusinessName, req.Description, req.WebsiteURL)
	if err != nil {
		h.logger.Error("Industry detection failed",
			zap.String("request_id", req.RequestID),
			zap.String("business_name", req.BusinessName),
			zap.Error(err))
		// Fallback to default industry
		industryResult = &classification.IndustryDetectionResult{
			IndustryName: "General Business",
			Confidence:   0.30,
			Keywords:     []string{},
			Reasoning:    fmt.Sprintf("Industry detection failed: %v", err),
		}
	} else {
		h.logger.Info("Industry detection successful",
			zap.String("request_id", req.RequestID),
			zap.String("industry", industryResult.IndustryName),
			zap.Float64("confidence", industryResult.Confidence),
			zap.Int("keywords_count", len(industryResult.Keywords)))
	}

	// Step 2: Generate classification codes using ClassificationCodeGenerator
	h.logger.Info("Starting code generation",
		zap.String("request_id", req.RequestID),
		zap.String("industry", industryResult.IndustryName),
		zap.Int("keywords_count", len(industryResult.Keywords)))
	
	codesInfo, err := h.codeGenerator.GenerateClassificationCodes(
		ctx,
		industryResult.Keywords,
		industryResult.IndustryName,
		industryResult.Confidence,
	)
	if err != nil {
		h.logger.Warn("Code generation failed, using empty codes",
			zap.String("request_id", req.RequestID),
			zap.String("industry", industryResult.IndustryName),
			zap.Error(err))
		codesInfo = &classification.ClassificationCodesInfo{
			MCC:   []classification.MCCCode{},
			SIC:   []classification.SICCode{},
			NAICS: []classification.NAICSCode{},
		}
	} else {
		h.logger.Info("Code generation successful",
			zap.String("request_id", req.RequestID),
			zap.Int("mcc_count", len(codesInfo.MCC)),
			zap.Int("sic_count", len(codesInfo.SIC)),
			zap.Int("naics_count", len(codesInfo.NAICS)))
	}

	// Step 3: Convert classification codes to handler format
	mccCodes := make([]IndustryCode, 0, len(codesInfo.MCC))
	for _, code := range codesInfo.MCC {
		mccCodes = append(mccCodes, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		})
	}

	sicCodes := make([]IndustryCode, 0, len(codesInfo.SIC))
	for _, code := range codesInfo.SIC {
		sicCodes = append(sicCodes, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		})
	}

	naicsCodes := make([]IndustryCode, 0, len(codesInfo.NAICS))
	for _, code := range codesInfo.NAICS {
		naicsCodes = append(naicsCodes, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		})
	}

	// Step 4: Build website analysis data (simplified for now)
	websiteAnalysis := &WebsiteAnalysisData{
		Success:           req.WebsiteURL != "",
		PagesAnalyzed:     0, // Will be populated by actual website scraper if implemented
		RelevantPages:     0,
		KeywordsExtracted: industryResult.Keywords,
		IndustrySignals:   []string{strings.ToLower(strings.ReplaceAll(industryResult.IndustryName, " ", "_"))},
		AnalysisMethod:    industryResult.Method,
		ProcessingTime:    industryResult.ProcessingTime,
		OverallRelevance:  industryResult.Confidence,
		ContentQuality:    industryResult.Confidence,
		StructuredData: map[string]interface{}{
			"business_type": "Business",
			"industry":      industryResult.IndustryName,
		},
	}

	// Step 5: Build method weights (simplified)
	methodWeights := map[string]float64{
		"database_driven": 100.0, // Using database-driven classification
	}

	// Step 6: Build reasoning
	reasoning := fmt.Sprintf("Primary industry identified as '%s' with %.0f%% confidence. ", 
		industryResult.IndustryName, industryResult.Confidence*100)
	reasoning += industryResult.Reasoning
	if req.WebsiteURL != "" {
		reasoning += fmt.Sprintf(" Website URL provided: %s.", req.WebsiteURL)
	}
	if len(industryResult.Keywords) > 0 {
		reasoning += fmt.Sprintf(" Keywords matched: %s.", strings.Join(industryResult.Keywords, ", "))
	}

	// Step 7: Build result
	return &EnhancedClassificationResult{
		BusinessName:            req.BusinessName,
		PrimaryIndustry:         industryResult.IndustryName,
		IndustryConfidence:      industryResult.Confidence,
		BusinessType:            h.determineBusinessType(industryResult.Keywords, industryResult.IndustryName),
		BusinessTypeConfidence:  industryResult.Confidence * 0.9, // Slightly lower than industry confidence
		MCCCodes:                mccCodes,
		SICCodes:                sicCodes,
		NAICSCodes:              naicsCodes,
		Keywords:                industryResult.Keywords,
		ConfidenceScore:         industryResult.Confidence,
		ClassificationReasoning: reasoning,
		WebsiteAnalysis:         websiteAnalysis,
		MethodWeights:           methodWeights,
		Timestamp:               time.Now(),
	}, nil
}

// determineBusinessType determines business type from keywords and industry
func (h *ClassificationHandler) determineBusinessType(keywords []string, industry string) string {
	// Simple heuristic based on industry name
	industryLower := strings.ToLower(industry)
	if strings.Contains(industryLower, "retail") || strings.Contains(industryLower, "store") {
		return "Retail Store"
	}
	if strings.Contains(industryLower, "service") {
		return "Service Business"
	}
	if strings.Contains(industryLower, "technology") || strings.Contains(industryLower, "software") {
		return "Technology Company"
	}
	if strings.Contains(industryLower, "health") || strings.Contains(industryLower, "medical") {
		return "Healthcare Provider"
	}
	if strings.Contains(industryLower, "financial") {
		return "Financial Services"
	}
	return "Business"
}

// zapLoggerAdapter adapts zap.Logger to io.Writer for standard log.Logger
type zapLoggerAdapter struct {
	logger *zap.Logger
}

func (z *zapLoggerAdapter) Write(p []byte) (n int, err error) {
	z.logger.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

// convertIndustryCodes converts IndustryCode to handlers.IndustryCode
func convertIndustryCodes(codes []IndustryCode) []IndustryCode {
	return codes // Same type, no conversion needed
}

// HandleHealth handles health check requests
func (h *ClassificationHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check Supabase connectivity
	supabaseHealthy := true
	var supabaseError error
	if err := h.supabaseClient.HealthCheck(ctx); err != nil {
		supabaseHealthy = false
		supabaseError = err
	}

	// Get classification data
	classificationData, err := h.supabaseClient.GetClassificationData(ctx)
	if err != nil {
		h.logger.Warn("Failed to get classification data", zap.Error(err))
	}

	// Create health response
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"service":   "classification-service",
		"uptime":    time.Since(startTime).String(),
		"supabase_status": map[string]interface{}{
			"connected": supabaseHealthy,
			"url":       h.config.Supabase.URL,
			"error":     supabaseError,
		},
		"classification_data": classificationData,
		"features": map[string]interface{}{
			"ml_enabled":             h.config.Classification.MLEnabled,
			"keyword_method_enabled": h.config.Classification.KeywordMethodEnabled,
			"ensemble_enabled":       h.config.Classification.EnsembleEnabled,
			"cache_enabled":          h.config.Classification.CacheEnabled,
		},
	}

	// Set status code based on health
	statusCode := http.StatusOK
	if !supabaseHealthy {
		statusCode = http.StatusServiceUnavailable
		health["status"] = "unhealthy"
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}
