package intelligent_routing

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RequestAnalyzerConfig represents the configuration for the request analyzer
type RequestAnalyzerConfig struct {
	EnableMLClassification   bool              `json:"enable_ml_classification"`
	EnableIndustryDetection  bool              `json:"enable_industry_detection"`
	EnableGeographicAnalysis bool              `json:"enable_geographic_analysis"`
	DefaultComplexity        RequestComplexity `json:"default_complexity"`
	DefaultPriority          RequestPriority   `json:"default_priority"`
	AnalysisTimeout          time.Duration     `json:"analysis_timeout"`
	ConfidenceThreshold      float64           `json:"confidence_threshold"`
}

// requestAnalyzer implements the RequestAnalyzer interface
type requestAnalyzer struct {
	config *RequestAnalyzerConfig
	logger *zap.Logger

	// Industry detection patterns
	industryPatterns map[string]*regexp.Regexp

	// Geographic region patterns
	geographicPatterns map[string]*regexp.Regexp

	// Business size patterns
	businessSizePatterns map[string]*regexp.Regexp

	// Risk level indicators
	riskIndicators map[string]float64
}

// NewRequestAnalyzer creates a new request analyzer instance
func NewRequestAnalyzer(config *RequestAnalyzerConfig, logger *zap.Logger) RequestAnalyzer {
	if config == nil {
		config = &RequestAnalyzerConfig{
			EnableMLClassification:   true,
			EnableIndustryDetection:  true,
			EnableGeographicAnalysis: true,
			DefaultComplexity:        ComplexityModerate,
			DefaultPriority:          PriorityNormal,
			AnalysisTimeout:          30 * time.Second,
			ConfidenceThreshold:      0.7,
		}
	}

	analyzer := &requestAnalyzer{
		config: config,
		logger: logger,
	}

	// Initialize patterns
	analyzer.initializePatterns()

	return analyzer
}

// AnalyzeRequest performs comprehensive analysis of a verification request
func (ra *requestAnalyzer) AnalyzeRequest(ctx context.Context, request *VerificationRequest) (*RequestAnalysis, error) {
	ra.logger.Info("Starting request analysis",
		zap.String("request_id", request.ID),
		zap.String("business_name", request.BusinessName))

	// Create analysis context with timeout
	analysisCtx, cancel := context.WithTimeout(ctx, ra.config.AnalysisTimeout)
	defer cancel()

	analysis := &RequestAnalysis{
		RequestID:  request.ID,
		AnalysisID: generateAnalysisID(),
		CreatedAt:  time.Now(),
	}

	// Perform classification
	classification, err := ra.ClassifyRequest(analysisCtx, request)
	if err != nil {
		ra.logger.Error("Failed to classify request",
			zap.String("request_id", request.ID),
			zap.Error(err))
		return nil, fmt.Errorf("classification failed: %w", err)
	}
	analysis.Classification = classification

	// Assess complexity
	complexity, err := ra.AssessComplexity(analysisCtx, request)
	if err != nil {
		ra.logger.Error("Failed to assess complexity",
			zap.String("request_id", request.ID),
			zap.Error(err))
		complexity = ra.config.DefaultComplexity
	}
	analysis.Complexity = complexity

	// Determine priority
	priority, err := ra.DeterminePriority(analysisCtx, request)
	if err != nil {
		ra.logger.Error("Failed to determine priority",
			zap.String("request_id", request.ID),
			zap.Error(err))
		priority = ra.config.DefaultPriority
	}
	analysis.Priority = priority

	// Calculate resource needs
	analysis.ResourceNeeds = ra.calculateResourceNeeds(request, classification, complexity)

	// Identify risk factors
	analysis.RiskFactors = ra.identifyRiskFactors(request, classification)

	// Calculate overall confidence
	analysis.Confidence = ra.calculateConfidence(classification, complexity, priority)

	ra.logger.Info("Request analysis completed",
		zap.String("request_id", request.ID),
		zap.String("complexity", string(complexity)),
		zap.String("priority", string(priority)),
		zap.Float64("confidence", analysis.Confidence))

	return analysis, nil
}

// ClassifyRequest classifies a verification request based on its characteristics
func (ra *requestAnalyzer) ClassifyRequest(ctx context.Context, request *VerificationRequest) (*RequestClassification, error) {
	classification := &RequestClassification{
		Confidence: 0.0,
	}

	// Determine request type
	requestType, typeConfidence := ra.determineRequestType(request)
	classification.RequestType = requestType

	// Detect industry
	if ra.config.EnableIndustryDetection {
		industry, industryConfidence := ra.detectIndustry(request)
		classification.Industry = industry
		classification.Confidence += industryConfidence * 0.3
	}

	// Analyze geographic region
	if ra.config.EnableGeographicAnalysis {
		region, regionConfidence := ra.detectGeographicRegion(request)
		classification.GeographicRegion = region
		classification.Confidence += regionConfidence * 0.2
	}

	// Determine business size
	businessSize, sizeConfidence := ra.determineBusinessSize(request)
	classification.BusinessSize = businessSize
	classification.Confidence += sizeConfidence * 0.2

	// Assess compliance level
	complianceLevel, complianceConfidence := ra.assessComplianceLevel(request)
	classification.ComplianceLevel = complianceLevel
	classification.Confidence += complianceConfidence * 0.15

	// Determine risk level
	riskLevel, riskConfidence := ra.determineRiskLevel(request)
	classification.RiskLevel = riskLevel
	classification.Confidence += riskConfidence * 0.15

	// Add type confidence
	classification.Confidence += typeConfidence * 0.2

	// Ensure confidence is within bounds
	if classification.Confidence > 1.0 {
		classification.Confidence = 1.0
	}

	return classification, nil
}

// AssessComplexity determines the complexity level of a request
func (ra *requestAnalyzer) AssessComplexity(ctx context.Context, request *VerificationRequest) (RequestComplexity, error) {
	complexityScore := 0.0

	// Analyze business name complexity
	complexityScore += ra.analyzeNameComplexity(request.BusinessName)

	// Analyze address complexity
	complexityScore += ra.analyzeAddressComplexity(request.BusinessAddress)

	// Analyze industry complexity
	complexityScore += ra.analyzeIndustryComplexity(request.Industry)

	// Analyze metadata complexity
	complexityScore += ra.analyzeMetadataComplexity(request.Metadata)

	// Determine complexity level based on score
	switch {
	case complexityScore < 0.3:
		return ComplexitySimple, nil
	case complexityScore < 0.6:
		return ComplexityModerate, nil
	case complexityScore < 0.8:
		return ComplexityComplex, nil
	default:
		return ComplexityAdvanced, nil
	}
}

// DeterminePriority determines the priority level of a request
func (ra *requestAnalyzer) DeterminePriority(ctx context.Context, request *VerificationRequest) (RequestPriority, error) {
	priorityScore := 0.0

	// Check if priority is already set
	if request.Priority != "" {
		return request.Priority, nil
	}

	// Check for urgent indicators
	if ra.hasUrgentIndicators(request) {
		return PriorityUrgent, nil
	}

	// Analyze deadline
	if request.Deadline != nil {
		timeUntilDeadline := request.Deadline.Sub(time.Now())
		if timeUntilDeadline < 1*time.Hour {
			priorityScore += 0.8
		} else if timeUntilDeadline < 24*time.Hour {
			priorityScore += 0.6
		} else if timeUntilDeadline < 7*24*time.Hour {
			priorityScore += 0.4
		}
	}

	// Analyze client priority
	priorityScore += ra.analyzeClientPriority(request.ClientID)

	// Analyze business characteristics
	priorityScore += ra.analyzeBusinessPriority(request)

	// Determine priority based on score
	switch {
	case priorityScore < 0.3:
		return PriorityLow, nil
	case priorityScore < 0.6:
		return PriorityNormal, nil
	default:
		return PriorityHigh, nil
	}
}

// Helper methods for request analysis

func (ra *requestAnalyzer) determineRequestType(request *VerificationRequest) (RequestType, float64) {
	// Check if request type is already specified
	if request.RequestType != "" {
		return request.RequestType, 1.0
	}

	// Analyze business characteristics to determine type
	indicators := make(map[RequestType]float64)

	// Basic verification indicators
	if ra.isBasicVerification(request) {
		indicators[RequestTypeBasic] += 0.8
	}

	// Enhanced verification indicators
	if ra.isEnhancedVerification(request) {
		indicators[RequestTypeEnhanced] += 0.7
	}

	// Compliance verification indicators
	if ra.isComplianceVerification(request) {
		indicators[RequestTypeCompliance] += 0.9
	}

	// Risk assessment indicators
	if ra.isRiskAssessment(request) {
		indicators[RequestTypeRisk] += 0.8
	}

	// Find the type with highest score
	var bestType RequestType
	var bestScore float64

	for reqType, score := range indicators {
		if score > bestScore {
			bestType = reqType
			bestScore = score
		}
	}

	// Default to basic if no clear indicators
	if bestScore == 0 {
		return RequestTypeBasic, 0.5
	}

	return bestType, bestScore
}

func (ra *requestAnalyzer) detectIndustry(request *VerificationRequest) (string, float64) {
	// Check if industry is already specified
	if request.Industry != "" {
		return request.Industry, 0.9
	}

	// Analyze business name for industry indicators
	text := strings.ToLower(request.BusinessName + " " + request.BusinessAddress)

	for industry, pattern := range ra.industryPatterns {
		if pattern.MatchString(text) {
			return industry, 0.8
		}
	}

	return "unknown", 0.3
}

func (ra *requestAnalyzer) detectGeographicRegion(request *VerificationRequest) (string, float64) {
	text := strings.ToLower(request.BusinessAddress)

	for region, pattern := range ra.geographicPatterns {
		if pattern.MatchString(text) {
			return region, 0.8
		}
	}

	return "unknown", 0.3
}

func (ra *requestAnalyzer) determineBusinessSize(request *VerificationRequest) (string, float64) {
	text := strings.ToLower(request.BusinessName + " " + request.BusinessAddress)

	for size, pattern := range ra.businessSizePatterns {
		if pattern.MatchString(text) {
			return size, 0.7
		}
	}

	// Default to medium size
	return "medium", 0.5
}

func (ra *requestAnalyzer) assessComplianceLevel(request *VerificationRequest) (string, float64) {
	// Check metadata for compliance indicators
	if complianceLevel, exists := request.Metadata["compliance_level"]; exists {
		return complianceLevel, 0.9
	}

	// Analyze business characteristics for compliance needs
	text := strings.ToLower(request.BusinessName + " " + request.BusinessAddress)

	// High compliance indicators
	highComplianceKeywords := []string{"bank", "financial", "insurance", "healthcare", "pharmaceutical"}
	for _, keyword := range highComplianceKeywords {
		if strings.Contains(text, keyword) {
			return "high", 0.8
		}
	}

	// Medium compliance indicators
	mediumComplianceKeywords := []string{"legal", "accounting", "consulting", "real estate"}
	for _, keyword := range mediumComplianceKeywords {
		if strings.Contains(text, keyword) {
			return "medium", 0.7
		}
	}

	return "low", 0.6
}

func (ra *requestAnalyzer) determineRiskLevel(request *VerificationRequest) (string, float64) {
	riskScore := 0.0

	// Analyze business name for risk indicators
	text := strings.ToLower(request.BusinessName + " " + request.BusinessAddress)

	for indicator, weight := range ra.riskIndicators {
		if strings.Contains(text, indicator) {
			riskScore += weight
		}
	}

	// Determine risk level based on score
	switch {
	case riskScore < 0.3:
		return "low", 0.8
	case riskScore < 0.6:
		return "medium", 0.7
	default:
		return "high", 0.9
	}
}

func (ra *requestAnalyzer) calculateResourceNeeds(request *VerificationRequest, classification *RequestClassification, complexity RequestComplexity) ResourceNeeds {
	needs := ResourceNeeds{
		CPUUsage:       0.1,
		MemoryUsage:    0.1,
		NetworkUsage:   0.1,
		ProcessingTime: 5 * time.Second,
		Concurrency:    1,
	}

	// Adjust based on complexity
	switch complexity {
	case ComplexitySimple:
		needs.CPUUsage = 0.1
		needs.MemoryUsage = 0.1
		needs.ProcessingTime = 3 * time.Second
	case ComplexityModerate:
		needs.CPUUsage = 0.3
		needs.MemoryUsage = 0.3
		needs.ProcessingTime = 8 * time.Second
	case ComplexityComplex:
		needs.CPUUsage = 0.6
		needs.MemoryUsage = 0.6
		needs.ProcessingTime = 15 * time.Second
		needs.Concurrency = 2
	case ComplexityAdvanced:
		needs.CPUUsage = 0.9
		needs.MemoryUsage = 0.9
		needs.ProcessingTime = 30 * time.Second
		needs.Concurrency = 3
	}

	// Adjust based on request type
	switch classification.RequestType {
	case RequestTypeEnhanced:
		needs.ProcessingTime = time.Duration(float64(needs.ProcessingTime) * 1.5)
		needs.NetworkUsage = 0.5
	case RequestTypeCompliance:
		needs.ProcessingTime = time.Duration(float64(needs.ProcessingTime) * 2.0)
		needs.NetworkUsage = 0.7
	case RequestTypeRisk:
		needs.ProcessingTime = time.Duration(float64(needs.ProcessingTime) * 1.8)
		needs.NetworkUsage = 0.6
	}

	return needs
}

func (ra *requestAnalyzer) identifyRiskFactors(request *VerificationRequest, classification *RequestClassification) []string {
	var riskFactors []string

	// Check for high-risk industries
	if classification.Industry == "financial" || classification.Industry == "gambling" {
		riskFactors = append(riskFactors, "high-risk_industry")
	}

	// Check for high-risk regions
	if classification.GeographicRegion == "high_risk_region" {
		riskFactors = append(riskFactors, "high_risk_region")
	}

	// Check for compliance requirements
	if classification.ComplianceLevel == "high" {
		riskFactors = append(riskFactors, "high_compliance_requirements")
	}

	// Check for business size indicators
	if classification.BusinessSize == "large" {
		riskFactors = append(riskFactors, "large_business_scale")
	}

	// Check for urgent processing
	if request.Priority == PriorityUrgent {
		riskFactors = append(riskFactors, "urgent_processing_required")
	}

	return riskFactors
}

func (ra *requestAnalyzer) calculateConfidence(classification *RequestClassification, complexity RequestComplexity, priority RequestPriority) float64 {
	confidence := classification.Confidence

	// Adjust confidence based on complexity
	switch complexity {
	case ComplexitySimple:
		confidence += 0.1
	case ComplexityModerate:
		confidence += 0.0
	case ComplexityComplex:
		confidence -= 0.1
	case ComplexityAdvanced:
		confidence -= 0.2
	}

	// Adjust confidence based on priority
	switch priority {
	case PriorityLow:
		confidence += 0.05
	case PriorityNormal:
		confidence += 0.0
	case PriorityHigh:
		confidence -= 0.05
	case PriorityUrgent:
		confidence -= 0.1
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// Helper methods for complexity analysis

func (ra *requestAnalyzer) analyzeNameComplexity(name string) float64 {
	complexity := 0.0

	// Length complexity
	if len(name) > 50 {
		complexity += 0.3
	} else if len(name) > 30 {
		complexity += 0.2
	}

	// Special characters complexity
	specialCharCount := 0
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == ' ') {
			specialCharCount++
		}
	}
	complexity += float64(specialCharCount) * 0.1

	// Word count complexity
	words := strings.Fields(name)
	if len(words) > 5 {
		complexity += 0.2
	}

	return complexity
}

func (ra *requestAnalyzer) analyzeAddressComplexity(address string) float64 {
	complexity := 0.0

	// Length complexity
	if len(address) > 100 {
		complexity += 0.3
	} else if len(address) > 50 {
		complexity += 0.2
	}

	// International address complexity
	if strings.Contains(strings.ToLower(address), "international") ||
		strings.Contains(strings.ToLower(address), "foreign") {
		complexity += 0.4
	}

	// PO Box complexity
	if strings.Contains(strings.ToUpper(address), "PO BOX") ||
		strings.Contains(strings.ToUpper(address), "P.O. BOX") {
		complexity += 0.2
	}

	return complexity
}

func (ra *requestAnalyzer) analyzeIndustryComplexity(industry string) float64 {
	// High complexity industries
	highComplexityIndustries := []string{"financial", "healthcare", "pharmaceutical", "technology"}
	for _, highComplexity := range highComplexityIndustries {
		if strings.Contains(strings.ToLower(industry), highComplexity) {
			return 0.4
		}
	}

	// Medium complexity industries
	mediumComplexityIndustries := []string{"legal", "accounting", "consulting", "real estate"}
	for _, mediumComplexity := range mediumComplexityIndustries {
		if strings.Contains(strings.ToLower(industry), mediumComplexity) {
			return 0.2
		}
	}

	return 0.1
}

func (ra *requestAnalyzer) analyzeMetadataComplexity(metadata map[string]string) float64 {
	complexity := 0.0

	// More metadata means more complexity
	if len(metadata) > 10 {
		complexity += 0.3
	} else if len(metadata) > 5 {
		complexity += 0.2
	}

	// Specific metadata types that indicate complexity
	complexityIndicators := []string{"compliance_level", "risk_assessment", "regulatory_requirements"}
	for _, indicator := range complexityIndicators {
		if _, exists := metadata[indicator]; exists {
			complexity += 0.2
		}
	}

	return complexity
}

// Helper methods for priority analysis

func (ra *requestAnalyzer) hasUrgentIndicators(request *VerificationRequest) bool {
	// Check metadata for urgent indicators
	if urgent, exists := request.Metadata["urgent"]; exists && urgent == "true" {
		return true
	}

	// Check for immediate deadline
	if request.Deadline != nil && request.Deadline.Sub(time.Now()) < 30*time.Minute {
		return true
	}

	// Check for urgent keywords in business name
	urgentKeywords := []string{"emergency", "urgent", "immediate", "critical"}
	text := strings.ToLower(request.BusinessName)
	for _, keyword := range urgentKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

func (ra *requestAnalyzer) analyzeClientPriority(clientID string) float64 {
	// This would typically query a client database
	// For now, use a simple heuristic based on client ID
	if strings.Contains(clientID, "premium") {
		return 0.4
	} else if strings.Contains(clientID, "enterprise") {
		return 0.3
	} else if strings.Contains(clientID, "standard") {
		return 0.1
	}
	return 0.0
}

func (ra *requestAnalyzer) analyzeBusinessPriority(request *VerificationRequest) float64 {
	priority := 0.0

	// Large businesses get higher priority
	text := strings.ToLower(request.BusinessName)
	largeBusinessKeywords := []string{"corporation", "corp", "inc", "llc", "ltd", "company"}
	for _, keyword := range largeBusinessKeywords {
		if strings.Contains(text, keyword) {
			priority += 0.2
			break
		}
	}

	// Financial businesses get higher priority
	financialKeywords := []string{"bank", "financial", "credit", "investment"}
	for _, keyword := range financialKeywords {
		if strings.Contains(text, keyword) {
			priority += 0.3
			break
		}
	}

	return priority
}

// Helper methods for request type detection

func (ra *requestAnalyzer) isBasicVerification(request *VerificationRequest) bool {
	// Basic verification for simple business names
	text := strings.ToLower(request.BusinessName)
	simpleKeywords := []string{"shop", "store", "market", "cafe", "restaurant"}
	for _, keyword := range simpleKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func (ra *requestAnalyzer) isEnhancedVerification(request *VerificationRequest) bool {
	// Enhanced verification for medium complexity businesses
	text := strings.ToLower(request.BusinessName)
	enhancedKeywords := []string{"services", "consulting", "solutions", "group"}
	for _, keyword := range enhancedKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func (ra *requestAnalyzer) isComplianceVerification(request *VerificationRequest) bool {
	// Compliance verification for regulated industries
	text := strings.ToLower(request.BusinessName + " " + request.BusinessAddress)
	complianceKeywords := []string{"bank", "financial", "insurance", "healthcare", "legal"}
	for _, keyword := range complianceKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func (ra *requestAnalyzer) isRiskAssessment(request *VerificationRequest) bool {
	// Risk assessment for high-risk businesses
	text := strings.ToLower(request.BusinessName + " " + request.BusinessAddress)
	riskKeywords := []string{"gambling", "casino", "adult", "weapon", "chemical"}
	for _, keyword := range riskKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// initializePatterns initializes the pattern matching for various analyses
func (ra *requestAnalyzer) initializePatterns() {
	// Industry patterns
	ra.industryPatterns = map[string]*regexp.Regexp{
		"financial":     regexp.MustCompile(`(?i)(bank|financial|credit|investment|insurance)`),
		"healthcare":    regexp.MustCompile(`(?i)(hospital|clinic|medical|healthcare|pharmacy)`),
		"technology":    regexp.MustCompile(`(?i)(tech|software|it|computer|digital)`),
		"retail":        regexp.MustCompile(`(?i)(shop|store|market|retail|supermarket)`),
		"manufacturing": regexp.MustCompile(`(?i)(manufacturing|factory|industrial|production)`),
		"legal":         regexp.MustCompile(`(?i)(law|legal|attorney|lawyer|legal services)`),
		"real_estate":   regexp.MustCompile(`(?i)(real estate|property|realtor|housing)`),
	}

	// Geographic patterns
	ra.geographicPatterns = map[string]*regexp.Regexp{
		"north_america": regexp.MustCompile(`(?i)(usa|united states|canada|mexico)`),
		"europe":        regexp.MustCompile(`(?i)(uk|united kingdom|germany|france|spain|italy)`),
		"asia":          regexp.MustCompile(`(?i)(china|japan|india|singapore|hong kong)`),
		"australia":     regexp.MustCompile(`(?i)(australia|new zealand)`),
	}

	// Business size patterns
	ra.businessSizePatterns = map[string]*regexp.Regexp{
		"small":  regexp.MustCompile(`(?i)(shop|store|cafe|restaurant|small)`),
		"medium": regexp.MustCompile(`(?i)(services|consulting|group|company)`),
		"large":  regexp.MustCompile(`(?i)(corporation|corp|inc|enterprise|multinational)`),
	}

	// Risk indicators
	ra.riskIndicators = map[string]float64{
		"gambling":      0.8,
		"casino":        0.8,
		"adult":         0.7,
		"weapon":        0.9,
		"chemical":      0.6,
		"financial":     0.5,
		"international": 0.4,
	}
}

// generateAnalysisID generates a unique analysis ID
func generateAnalysisID() string {
	return fmt.Sprintf("analysis_%d", time.Now().UnixNano())
}
