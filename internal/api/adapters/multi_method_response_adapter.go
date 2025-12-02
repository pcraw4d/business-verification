package adapters

import (
	"context"
	"fmt"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/shared"
)

// MultiMethodResponseAdapter adapts classification results to the standard API response format
// Phase 5.1: Updated to use IndustryDetectionService instead of MultiMethodClassifier
type MultiMethodResponseAdapter struct {
	detectionService interface{} // *classification.IndustryDetectionService - using interface to avoid import cycle
}

// NewMultiMethodResponseAdapter creates a new response adapter
func NewMultiMethodResponseAdapter(detectionService interface{}) *MultiMethodResponseAdapter {
	return &MultiMethodResponseAdapter{
		detectionService: detectionService,
	}
}

// AdaptMultiMethodResultToResponse converts an industry detection result to a standard API response
// Phase 5.1: Updated to work with IndustryDetectionResult instead of MultiMethodClassificationResult
func (adapter *MultiMethodResponseAdapter) AdaptMultiMethodResultToResponse(
	ctx context.Context,
	request *shared.BusinessClassificationRequest,
	result interface{}, // *classification.IndustryDetectionResult - using interface for backward compatibility
) (*shared.BusinessClassificationResponse, error) {
	if result == nil {
		return nil, fmt.Errorf("result cannot be nil")
	}

	// Type assert to IndustryDetectionResult
	detectionResult, ok := result.(*classification.IndustryDetectionResult)
	if !ok {
		return nil, fmt.Errorf("invalid result type, expected IndustryDetectionResult")
	}

	// Create the enhanced response
	response := &shared.BusinessClassificationResponse{
		ID:                    request.ID,
		BusinessName:          request.BusinessName,
		DetectedIndustry:      detectionResult.IndustryName,
		Confidence:            detectionResult.Confidence,
		Classifications: []shared.IndustryClassification{
			{
				IndustryName:         detectionResult.IndustryName,
				ConfidenceScore:      detectionResult.Confidence,
				ClassificationMethod: detectionResult.Method,
				Keywords:             detectionResult.Keywords,
			},
		},
		PrimaryClassification: &shared.IndustryClassification{
			IndustryName:         detectionResult.IndustryName,
			ConfidenceScore:      detectionResult.Confidence,
			ClassificationMethod: detectionResult.Method,
			Keywords:             detectionResult.Keywords,
		},
		OverallConfidence:     detectionResult.Confidence,
		ClassificationMethod:  detectionResult.Method,
		ProcessingTime:        detectionResult.ProcessingTime,
		CreatedAt:             detectionResult.CreatedAt,
		Timestamp:             time.Now(),
		ClassificationReasoning: detectionResult.Reasoning,

		// Metadata
		Metadata: map[string]interface{}{
			"method": detectionResult.Method,
		},
	}

	return response, nil
}

// ClassifyWithMultiMethod performs classification and returns a standard response
// Phase 5.1: Updated to use IndustryDetectionService
func (adapter *MultiMethodResponseAdapter) ClassifyWithMultiMethod(
	ctx context.Context,
	request *shared.BusinessClassificationRequest,
) (*shared.BusinessClassificationResponse, error) {
	// Type assert to get IndustryDetectionService
	detectionService, ok := adapter.detectionService.(interface {
		DetectIndustry(ctx context.Context, businessName, description, websiteURL string) (*classification.IndustryDetectionResult, error)
	})
	if !ok {
		return nil, fmt.Errorf("detection service not available or wrong type")
	}

	// Perform classification using detection service
	result, err := detectionService.DetectIndustry(
		ctx,
		request.BusinessName,
		request.Description,
		request.WebsiteURL,
	)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Convert to standard response format
	response := &shared.BusinessClassificationResponse{
		ID:                    request.ID,
		BusinessName:          request.BusinessName,
		DetectedIndustry:      result.IndustryName,
		Confidence:            result.Confidence,
		ClassificationMethod:  result.Method,
		ProcessingTime:        result.ProcessingTime,
		CreatedAt:             result.CreatedAt,
		Timestamp:             time.Now(),
		Classifications: []shared.IndustryClassification{
			{
				IndustryName:         result.IndustryName,
				ConfidenceScore:      result.Confidence,
				ClassificationMethod: result.Method,
				Keywords:             result.Keywords,
			},
		},
		PrimaryClassification: &shared.IndustryClassification{
			IndustryName:         result.IndustryName,
			ConfidenceScore:      result.Confidence,
			ClassificationMethod: result.Method,
			Keywords:             result.Keywords,
		},
		OverallConfidence:     result.Confidence,
		ClassificationReasoning: result.Reasoning,
		Metadata: map[string]interface{}{
			"method": result.Method,
		},
	}

	return response, nil
}

// GetMethodBreakdownSummary returns a summary of the method breakdown for display purposes
func (adapter *MultiMethodResponseAdapter) GetMethodBreakdownSummary(
	methodBreakdown []shared.ClassificationMethodResult,
) map[string]interface{} {
	summary := map[string]interface{}{
		"total_methods":      len(methodBreakdown),
		"successful_methods": adapter.countSuccessfulMethods(methodBreakdown),
		"failed_methods":     len(methodBreakdown) - adapter.countSuccessfulMethods(methodBreakdown),
		"method_types":       make(map[string]int),
		"average_confidence": 0.0,
		"method_details":     make([]map[string]interface{}, 0),
	}

	// Count method types and calculate average confidence
	var totalConfidence float64
	successfulCount := 0

	for _, method := range methodBreakdown {
		// Count method types
		if count, exists := summary["method_types"].(map[string]int)[method.MethodType]; exists {
			summary["method_types"].(map[string]int)[method.MethodType] = count + 1
		} else {
			summary["method_types"].(map[string]int)[method.MethodType] = 1
		}

		// Calculate average confidence for successful methods
		if method.Success {
			totalConfidence += method.Confidence
			successfulCount++
		}

		// Add method details
		methodDetail := map[string]interface{}{
			"method_name":     method.MethodName,
			"method_type":     method.MethodType,
			"success":         method.Success,
			"confidence":      method.Confidence,
			"processing_time": method.ProcessingTime.String(),
			"industry_result": method.Result.IndustryName,
		}

		if !method.Success {
			methodDetail["error"] = method.Error
		}

		summary["method_details"] = append(summary["method_details"].([]map[string]interface{}), methodDetail)
	}

	// Calculate average confidence
	if successfulCount > 0 {
		summary["average_confidence"] = totalConfidence / float64(successfulCount)
	}

	return summary
}

// GetQualityMetricsSummary returns a summary of quality metrics for display purposes
func (adapter *MultiMethodResponseAdapter) GetQualityMetricsSummary(
	qualityMetrics *shared.ClassificationQuality,
) map[string]interface{} {
	if qualityMetrics == nil {
		return map[string]interface{}{
			"overall_quality":     0.0,
			"method_agreement":    0.0,
			"confidence_variance": 1.0,
			"evidence_strength":   0.0,
			"data_completeness":   0.0,
			"quality_grade":       "F",
		}
	}

	// Calculate quality grade
	qualityGrade := adapter.calculateQualityGrade(qualityMetrics.OverallQuality)

	return map[string]interface{}{
		"overall_quality":     qualityMetrics.OverallQuality,
		"method_agreement":    qualityMetrics.MethodAgreement,
		"confidence_variance": qualityMetrics.ConfidenceVariance,
		"evidence_strength":   qualityMetrics.EvidenceStrength,
		"data_completeness":   qualityMetrics.DataCompleteness,
		"quality_grade":       qualityGrade,
		"quality_description": adapter.getQualityDescription(qualityMetrics.OverallQuality),
	}
}

// Helper methods

func (adapter *MultiMethodResponseAdapter) countSuccessfulMethods(
	methodResults []shared.ClassificationMethodResult,
) int {
	count := 0
	for _, method := range methodResults {
		if method.Success {
			count++
		}
	}
	return count
}

func (adapter *MultiMethodResponseAdapter) calculateQualityGrade(overallQuality float64) string {
	switch {
	case overallQuality >= 0.9:
		return "A"
	case overallQuality >= 0.8:
		return "B"
	case overallQuality >= 0.7:
		return "C"
	case overallQuality >= 0.6:
		return "D"
	default:
		return "F"
	}
}

func (adapter *MultiMethodResponseAdapter) getQualityDescription(overallQuality float64) string {
	switch {
	case overallQuality >= 0.9:
		return "Excellent - High confidence in classification accuracy"
	case overallQuality >= 0.8:
		return "Good - Reliable classification with strong evidence"
	case overallQuality >= 0.7:
		return "Fair - Reasonable classification with moderate evidence"
	case overallQuality >= 0.6:
		return "Poor - Low confidence classification with weak evidence"
	default:
		return "Very Poor - Very low confidence classification"
	}
}
