package adapters

import (
	"context"
	"fmt"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/shared"
)

// MultiMethodResponseAdapter adapts multi-method classification results to the standard API response format
type MultiMethodResponseAdapter struct {
	multiMethodClassifier *classification.MultiMethodClassifier
}

// NewMultiMethodResponseAdapter creates a new multi-method response adapter
func NewMultiMethodResponseAdapter(multiMethodClassifier *classification.MultiMethodClassifier) *MultiMethodResponseAdapter {
	return &MultiMethodResponseAdapter{
		multiMethodClassifier: multiMethodClassifier,
	}
}

// AdaptMultiMethodResultToResponse converts a multi-method classification result to a standard API response
func (adapter *MultiMethodResponseAdapter) AdaptMultiMethodResultToResponse(
	ctx context.Context,
	request *shared.BusinessClassificationRequest,
	multiMethodResult *classification.MultiMethodClassificationResult,
) (*shared.BusinessClassificationResponse, error) {
	if multiMethodResult == nil {
		return nil, fmt.Errorf("multi-method result cannot be nil")
	}

	// Convert method results to shared format
	var methodBreakdown []shared.ClassificationMethodResult
	for _, method := range multiMethodResult.MethodResults {
		methodBreakdown = append(methodBreakdown, shared.ClassificationMethodResult{
			MethodName:     method.MethodName,
			MethodType:     method.MethodType,
			Confidence:     method.Confidence,
			ProcessingTime: method.ProcessingTime,
			Result:         method.Result,
			Evidence:       method.Evidence,
			Keywords:       method.Keywords,
			Error:          method.Error,
			Success:        method.Success,
		})
	}

	// Create the enhanced response
	response := &shared.BusinessClassificationResponse{
		ID:                    request.ID,
		BusinessName:          request.BusinessName,
		DetectedIndustry:      multiMethodResult.PrimaryClassification.IndustryName,
		Confidence:            multiMethodResult.EnsembleConfidence,
		Classifications:       []shared.IndustryClassification{*multiMethodResult.PrimaryClassification},
		PrimaryClassification: multiMethodResult.PrimaryClassification,
		OverallConfidence:     multiMethodResult.EnsembleConfidence,
		ClassificationMethod:  "multi_method_ensemble",
		ProcessingTime:        multiMethodResult.ProcessingTime,
		CreatedAt:             multiMethodResult.CreatedAt,
		Timestamp:             time.Now(),

		// Enhanced multi-method fields
		MethodBreakdown:         methodBreakdown,
		EnsembleConfidence:      multiMethodResult.EnsembleConfidence,
		ClassificationReasoning: multiMethodResult.ClassificationReasoning,
		QualityMetrics:          multiMethodResult.QualityMetrics,

		// Metadata
		Metadata: map[string]interface{}{
			"multi_method_enabled": true,
			"method_count":         len(multiMethodResult.MethodResults),
			"successful_methods":   adapter.countSuccessfulMethods(multiMethodResult.MethodResults),
			"ensemble_method":      "weighted_average",
			"quality_score":        multiMethodResult.QualityMetrics.OverallQuality,
		},
	}

	// Add classification codes if available
	if multiMethodResult.PrimaryClassification.Metadata != nil {
		if codes, exists := multiMethodResult.PrimaryClassification.Metadata["classification_codes"]; exists {
			if classificationCodes, ok := codes.(shared.ClassificationCodes); ok {
				response.ClassificationCodes = classificationCodes
			}
		}
	}

	return response, nil
}

// ClassifyWithMultiMethod performs classification using multiple methods and returns a standard response
func (adapter *MultiMethodResponseAdapter) ClassifyWithMultiMethod(
	ctx context.Context,
	request *shared.BusinessClassificationRequest,
) (*shared.BusinessClassificationResponse, error) {
	// Perform multi-method classification
	multiMethodResult, err := adapter.multiMethodClassifier.ClassifyWithMultipleMethods(
		ctx,
		request.BusinessName,
		request.Description,
		request.WebsiteURL,
	)
	if err != nil {
		return nil, fmt.Errorf("multi-method classification failed: %w", err)
	}

	// Adapt to standard response format
	response, err := adapter.AdaptMultiMethodResultToResponse(ctx, request, multiMethodResult)
	if err != nil {
		return nil, fmt.Errorf("failed to adapt multi-method result: %w", err)
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
