package classification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// MultiIndustryClassification represents a top-3 industry classification result
type MultiIndustryClassification struct {
	Classifications      []IndustryClassification `json:"classifications"`
	PrimaryIndustry      IndustryClassification   `json:"primary_industry"`
	SecondaryIndustry    *IndustryClassification  `json:"secondary_industry,omitempty"`
	TertiaryIndustry     *IndustryClassification  `json:"tertiary_industry,omitempty"`
	OverallConfidence    float64                  `json:"overall_confidence"`
	ClassificationMethod string                   `json:"classification_method"`
	ProcessingTime       time.Duration            `json:"processing_time"`
	ValidationScore      float64                  `json:"validation_score,omitempty"`
}

// MultiIndustryService provides enhanced multi-industry classification functionality
type MultiIndustryService struct {
	baseService *ClassificationService
	logger      *observability.Logger
	metrics     *observability.Metrics

	// Configuration
	minConfidenceThreshold float64
	maxClassifications     int
	confidenceWeighting    map[string]float64

	// Enhanced ranking engine
	rankingEngine *ConfidenceRankingEngine

	// Top-3 selection engine
	top3SelectionEngine *Top3SelectionEngine
}

// NewMultiIndustryService creates a new multi-industry classification service
func NewMultiIndustryService(baseService *ClassificationService, logger *observability.Logger, metrics *observability.Metrics) *MultiIndustryService {
	return &MultiIndustryService{
		baseService: baseService,
		logger:      logger,
		metrics:     metrics,

		// Configuration
		minConfidenceThreshold: 0.1, // Minimum confidence to include in results
		maxClassifications:     3,   // Maximum number of classifications to return
		confidenceWeighting: map[string]float64{
			"keyword_match":     0.3,
			"description_match": 0.25,
			"business_type":     0.2,
			"industry_hint":     0.15,
			"fuzzy_match":       0.1,
		},

		// Enhanced ranking engine
		rankingEngine: NewConfidenceRankingEngine(),

		// Top-3 selection engine
		top3SelectionEngine: NewTop3SelectionEngine(logger, metrics),
	}
}

// ClassifyBusinessMultiIndustry performs multi-industry classification with top-3 results
func (m *MultiIndustryService) ClassifyBusinessMultiIndustry(ctx context.Context, req *ClassificationRequest) (*MultiIndustryClassification, error) {
	start := time.Now()

	// Log multi-industry classification start
	m.logger.WithComponent("multi_industry_classification").LogBusinessEvent(ctx, "multi_industry_classification_started", "", map[string]interface{}{
		"business_name": req.BusinessName,
		"business_type": req.BusinessType,
		"industry":      req.Industry,
	})

	// Perform base classification
	baseResponse, err := m.baseService.ClassifyBusiness(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("base classification failed: %w", err)
	}

	// Generate additional classifications using different methods
	allClassifications := m.generateMultiIndustryClassifications(ctx, req, baseResponse)

	// Apply enhanced confidence-based ranking and filtering
	rankedClassifications := m.rankAndFilterClassifications(allClassifications)

	// Select top-3 classifications using enhanced selection engine
	top3Result := m.top3SelectionEngine.SelectTop3Classifications(ctx, rankedClassifications)
	topClassifications := top3Result.AllClassifications

	// Calculate overall confidence
	overallConfidence := m.calculateOverallConfidence(topClassifications)

	// Create multi-industry response
	result := &MultiIndustryClassification{
		Classifications:      topClassifications,
		PrimaryIndustry:      top3Result.PrimaryIndustry,
		SecondaryIndustry:    top3Result.SecondaryIndustry,
		TertiaryIndustry:     top3Result.TertiaryIndustry,
		OverallConfidence:    overallConfidence,
		ClassificationMethod: "multi_industry_enhanced",
		ProcessingTime:       time.Since(start),
	}

	// Calculate validation score using top-3 selection metrics
	if top3Result.SelectionMetrics != nil {
		result.ValidationScore = top3Result.SelectionMetrics.SelectionQuality
	} else {
		result.ValidationScore = m.calculateValidationScore(result)
	}

	// Log completion
	m.logger.WithComponent("multi_industry_classification").LogBusinessEvent(ctx, "multi_industry_classification_completed", baseResponse.BusinessID, map[string]interface{}{
		"business_name":         req.BusinessName,
		"primary_industry_code": result.PrimaryIndustry.IndustryCode,
		"primary_industry_name": result.PrimaryIndustry.IndustryName,
		"overall_confidence":    overallConfidence,
		"processing_time_ms":    time.Since(start).Milliseconds(),
		"total_classifications": len(topClassifications),
		"validation_score":      result.ValidationScore,
	})

	// Record metrics
	m.metrics.RecordBusinessClassification("multi_industry_success", fmt.Sprintf("%.2f", overallConfidence))

	return result, nil
}

// generateMultiIndustryClassifications generates classifications using multiple methods
func (m *MultiIndustryService) generateMultiIndustryClassifications(ctx context.Context, req *ClassificationRequest, baseResponse *ClassificationResponse) []IndustryClassification {
	var allClassifications []IndustryClassification

	// Add base classifications
	allClassifications = append(allClassifications, baseResponse.Classifications...)

	// Generate keyword-based classifications
	keywordClassifications := m.generateKeywordBasedClassifications(req)
	allClassifications = append(allClassifications, keywordClassifications...)

	// Generate description-based classifications
	descriptionClassifications := m.generateDescriptionBasedClassifications(req)
	allClassifications = append(allClassifications, descriptionClassifications...)

	// Generate business-type-based classifications
	businessTypeClassifications := m.generateBusinessTypeClassifications(req)
	allClassifications = append(allClassifications, businessTypeClassifications...)

	// Generate industry-hint-based classifications
	industryHintClassifications := m.generateIndustryHintClassifications(req)
	allClassifications = append(allClassifications, industryHintClassifications...)

	// Generate fuzzy-match classifications
	fuzzyClassifications := m.generateFuzzyMatchClassifications(req)
	allClassifications = append(allClassifications, fuzzyClassifications...)

	return allClassifications
}

// generateKeywordBasedClassifications generates classifications based on keyword analysis
func (m *MultiIndustryService) generateKeywordBasedClassifications(req *ClassificationRequest) []IndustryClassification {
	var classifications []IndustryClassification

	if req.Keywords == "" {
		return classifications
	}

	keywords := strings.Split(req.Keywords, ",")
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}

		// Find industries that match this keyword
		matchedIndustries := m.findIndustriesByKeyword(keyword)
		for _, industry := range matchedIndustries {
			classification := IndustryClassification{
				IndustryCode:         industry.Code,
				IndustryName:         industry.Name,
				ConfidenceScore:      m.calculateKeywordConfidence(keyword, industry),
				ClassificationMethod: "keyword_match",
				Keywords:             []string{keyword},
				Description:          fmt.Sprintf("Matched by keyword: %s", keyword),
			}
			classifications = append(classifications, classification)
		}
	}

	return classifications
}

// generateDescriptionBasedClassifications generates classifications based on business description
func (m *MultiIndustryService) generateDescriptionBasedClassifications(req *ClassificationRequest) []IndustryClassification {
	var classifications []IndustryClassification

	if req.Description == "" {
		return classifications
	}

	// Extract key terms from description
	terms := m.extractKeyTerms(req.Description)

	for _, term := range terms {
		matchedIndustries := m.findIndustriesByKeyword(term)
		for _, industry := range matchedIndustries {
			classification := IndustryClassification{
				IndustryCode:         industry.Code,
				IndustryName:         industry.Name,
				ConfidenceScore:      m.calculateDescriptionConfidence(term, industry, req.Description),
				ClassificationMethod: "description_match",
				Keywords:             []string{term},
				Description:          fmt.Sprintf("Matched by description term: %s", term),
			}
			classifications = append(classifications, classification)
		}
	}

	return classifications
}

// generateBusinessTypeClassifications generates classifications based on business type
func (m *MultiIndustryService) generateBusinessTypeClassifications(req *ClassificationRequest) []IndustryClassification {
	var classifications []IndustryClassification

	if req.BusinessType == "" {
		return classifications
	}

	// Map business types to likely industries
	businessTypeIndustries := m.mapBusinessTypeToIndustries(req.BusinessType)

	for _, industry := range businessTypeIndustries {
		classification := IndustryClassification{
			IndustryCode:         industry.Code,
			IndustryName:         industry.Name,
			ConfidenceScore:      m.calculateBusinessTypeConfidence(req.BusinessType, industry),
			ClassificationMethod: "business_type",
			Keywords:             []string{req.BusinessType},
			Description:          fmt.Sprintf("Matched by business type: %s", req.BusinessType),
		}
		classifications = append(classifications, classification)
	}

	return classifications
}

// generateIndustryHintClassifications generates classifications based on industry hints
func (m *MultiIndustryService) generateIndustryHintClassifications(req *ClassificationRequest) []IndustryClassification {
	var classifications []IndustryClassification

	if req.Industry == "" {
		return classifications
	}

	// Find industries that match the provided industry hint
	matchedIndustries := m.findIndustriesByKeyword(req.Industry)

	for _, industry := range matchedIndustries {
		classification := IndustryClassification{
			IndustryCode:         industry.Code,
			IndustryName:         industry.Name,
			ConfidenceScore:      m.calculateIndustryHintConfidence(req.Industry, industry),
			ClassificationMethod: "industry_hint",
			Keywords:             []string{req.Industry},
			Description:          fmt.Sprintf("Matched by industry hint: %s", req.Industry),
		}
		classifications = append(classifications, classification)
	}

	return classifications
}

// generateFuzzyMatchClassifications generates classifications using fuzzy matching
func (m *MultiIndustryService) generateFuzzyMatchClassifications(req *ClassificationRequest) []IndustryClassification {
	var classifications []IndustryClassification

	// Use fuzzy matching on business name
	matchedIndustries := m.findIndustriesByFuzzyMatch(req.BusinessName)

	for _, industry := range matchedIndustries {
		classification := IndustryClassification{
			IndustryCode:         industry.Code,
			IndustryName:         industry.Name,
			ConfidenceScore:      m.calculateFuzzyMatchConfidence(req.BusinessName, industry),
			ClassificationMethod: "fuzzy_match",
			Keywords:             []string{req.BusinessName},
			Description:          fmt.Sprintf("Matched by fuzzy matching: %s", req.BusinessName),
		}
		classifications = append(classifications, classification)
	}

	return classifications
}

// rankAndFilterClassifications ranks and filters classifications using enhanced confidence ranking
func (m *MultiIndustryService) rankAndFilterClassifications(classifications []IndustryClassification) []IndustryClassification {
	// Filter by minimum confidence threshold
	var filtered []IndustryClassification
	for _, classification := range classifications {
		if classification.ConfidenceScore >= m.minConfidenceThreshold {
			filtered = append(filtered, classification)
		}
	}

	// Use enhanced confidence ranking engine
	rankedClassifications := m.rankingEngine.RankClassifications(filtered)

	return rankedClassifications
}

// selectTopClassifications selects the top N classifications
func (m *MultiIndustryService) selectTopClassifications(classifications []IndustryClassification) []IndustryClassification {
	if len(classifications) <= m.maxClassifications {
		return classifications
	}
	return classifications[:m.maxClassifications]
}

// calculateOverallConfidence calculates the overall confidence score
func (m *MultiIndustryService) calculateOverallConfidence(classifications []IndustryClassification) float64 {
	if len(classifications) == 0 {
		return 0.0
	}

	// Weight primary classification more heavily
	if len(classifications) == 1 {
		return classifications[0].ConfidenceScore
	}

	// Calculate weighted average
	totalWeight := 0.0
	totalScore := 0.0

	for i, classification := range classifications {
		weight := 1.0 / float64(i+1) // Decreasing weight for each position
		totalWeight += weight
		totalScore += classification.ConfidenceScore * weight
	}

	return totalScore / totalWeight
}

// calculateValidationScore calculates a validation score for the multi-industry result
func (m *MultiIndustryService) calculateValidationScore(result *MultiIndustryClassification) float64 {
	score := 0.0
	factors := 0

	// Factor 1: Primary confidence score
	if result.PrimaryIndustry.ConfidenceScore > 0 {
		score += result.PrimaryIndustry.ConfidenceScore * 0.4
		factors++
	}

	// Factor 2: Consistency between classifications
	if len(result.Classifications) > 1 {
		consistency := m.calculateClassificationConsistency(result.Classifications)
		score += consistency * 0.3
		factors++
	}

	// Factor 3: Method diversity
	diversity := m.calculateMethodDiversity(result.Classifications)
	score += diversity * 0.2
	factors++

	// Factor 4: Overall confidence
	if result.OverallConfidence > 0 {
		score += result.OverallConfidence * 0.1
		factors++
	}

	if factors == 0 {
		return 0.0
	}

	return score / float64(factors)
}

// calculateClassificationConsistency calculates consistency between classifications
func (m *MultiIndustryService) calculateClassificationConsistency(classifications []IndustryClassification) float64 {
	if len(classifications) < 2 {
		return 1.0
	}

	// Check if classifications are in related industries
	relatedCount := 0
	totalComparisons := 0

	for i := 0; i < len(classifications); i++ {
		for j := i + 1; j < len(classifications); j++ {
			if m.areIndustriesRelated(classifications[i], classifications[j]) {
				relatedCount++
			}
			totalComparisons++
		}
	}

	if totalComparisons == 0 {
		return 1.0
	}

	return float64(relatedCount) / float64(totalComparisons)
}

// calculateMethodDiversity calculates the diversity of classification methods
func (m *MultiIndustryService) calculateMethodDiversity(classifications []IndustryClassification) float64 {
	if len(classifications) == 0 {
		return 0.0
	}

	methods := make(map[string]bool)
	for _, classification := range classifications {
		methods[classification.ClassificationMethod] = true
	}

	// More diverse methods = higher score
	return float64(len(methods)) / float64(len(classifications))
}

// Helper methods (to be implemented based on existing classification logic)
func (m *MultiIndustryService) findIndustriesByKeyword(keyword string) []IndustryCode {
	// Implementation would use existing industry data
	return []IndustryCode{}
}

func (m *MultiIndustryService) findIndustriesByFuzzyMatch(businessName string) []IndustryCode {
	// Implementation would use fuzzy matching logic
	return []IndustryCode{}
}

func (m *MultiIndustryService) mapBusinessTypeToIndustries(businessType string) []IndustryCode {
	// Implementation would map business types to industries
	return []IndustryCode{}
}

func (m *MultiIndustryService) extractKeyTerms(description string) []string {
	// Implementation would extract key terms from description
	return []string{}
}

func (m *MultiIndustryService) areIndustriesRelated(industry1, industry2 IndustryClassification) bool {
	// Implementation would check if industries are related
	return false
}

// Confidence calculation methods
func (m *MultiIndustryService) calculateKeywordConfidence(keyword string, industry IndustryCode) float64 {
	// Implementation would calculate confidence based on keyword match
	return 0.0
}

func (m *MultiIndustryService) calculateDescriptionConfidence(term string, industry IndustryCode, description string) float64 {
	// Implementation would calculate confidence based on description match
	return 0.0
}

func (m *MultiIndustryService) calculateBusinessTypeConfidence(businessType string, industry IndustryCode) float64 {
	// Implementation would calculate confidence based on business type
	return 0.0
}

func (m *MultiIndustryService) calculateIndustryHintConfidence(hint string, industry IndustryCode) float64 {
	// Implementation would calculate confidence based on industry hint
	return 0.0
}

func (m *MultiIndustryService) calculateFuzzyMatchConfidence(businessName string, industry IndustryCode) float64 {
	// Implementation would calculate confidence based on fuzzy match
	return 0.0
}

// IndustryCode represents industry code data
type IndustryCode struct {
	Code string
	Name string
	// Add other fields as needed
}
