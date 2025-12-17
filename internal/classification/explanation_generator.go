package classification

import (
	"fmt"
	"strings"
	"time"
)

// ClassificationExplanation represents a structured explanation for a classification (Phase 2)
type ClassificationExplanation struct {
	PrimaryReason     string             `json:"primary_reason"`
	SupportingFactors []string           `json:"supporting_factors"`
	KeyTermsFound     []string           `json:"key_terms_found"`
	ConfidenceFactors map[string]float64 `json:"confidence_factors,omitempty"`
	MethodUsed        string             `json:"method_used"`
	ProcessingPath    string             `json:"processing_path"` // "fast_path", "full_strategy", "ml_validated"
	// Phase 5: Enhanced explanation fields
	LayerUsed         string             `json:"layer_used,omitempty"`         // "layer1", "layer2", "layer3"
	FromCache         bool               `json:"from_cache,omitempty"`         // Indicates if result came from cache
	CachedAt          *string            `json:"cached_at,omitempty"`          // When result was cached (ISO 8601)
	ProcessingTimeMs  int                `json:"processing_time_ms,omitempty"`  // Processing time in milliseconds
}

// ExplanationGenerator generates human-readable explanations for classifications (Phase 2)
type ExplanationGenerator struct{}

// NewExplanationGenerator creates a new explanation generator
func NewExplanationGenerator() *ExplanationGenerator {
	return &ExplanationGenerator{}
}

// GenerateExplanation generates a structured explanation for a classification result (Phase 2)
func (g *ExplanationGenerator) GenerateExplanation(
	result *MultiStrategyResult,
	codes *ClassificationCodesInfo,
	contentQuality float64,
) *ClassificationExplanation {
	exp := &ClassificationExplanation{
		MethodUsed:        result.Method,
		KeyTermsFound:     result.Keywords,
		ConfidenceFactors: g.extractConfidenceFactors(result),
		ProcessingPath:    g.determineProcessingPath(result),
	}

	// Generate primary reason based on method
	exp.PrimaryReason = g.generatePrimaryReason(result)

	// Generate supporting factors
	exp.SupportingFactors = g.generateSupportingFactors(result, codes, contentQuality)

	return exp
}

// GenerateExplanationWithPhase5 generates a structured explanation with Phase 5 enhancements
func (g *ExplanationGenerator) GenerateExplanationWithPhase5(
	result *MultiStrategyResult,
	codes *ClassificationCodesInfo,
	contentQuality float64,
	layerUsed string,
	fromCache bool,
	cachedAt *time.Time,
	processingTimeMs int,
) *ClassificationExplanation {
	exp := g.GenerateExplanation(result, codes, contentQuality)
	
	// Phase 5: Add layer and cache information
	exp.LayerUsed = layerUsed
	exp.FromCache = fromCache
	if cachedAt != nil {
		cachedAtStr := cachedAt.Format(time.RFC3339)
		exp.CachedAt = &cachedAtStr
	}
	exp.ProcessingTimeMs = processingTimeMs
	
	// Enhance primary reason with layer information
	if layerUsed != "" {
		layerInfo := g.getLayerDescription(layerUsed)
		if !fromCache {
			exp.PrimaryReason = fmt.Sprintf("%s (Processed via %s)", exp.PrimaryReason, layerInfo)
		} else {
			exp.PrimaryReason = fmt.Sprintf("%s (Retrieved from cache, originally processed via %s)", exp.PrimaryReason, layerInfo)
		}
	}
	
	// Add cache status to supporting factors
	if fromCache {
		exp.SupportingFactors = append([]string{"Result retrieved from cache (30-day TTL)"}, exp.SupportingFactors...)
	} else {
		exp.SupportingFactors = append([]string{fmt.Sprintf("Processed via %s", g.getLayerDescription(layerUsed))}, exp.SupportingFactors...)
	}
	
	// Add processing time information
	if processingTimeMs > 0 {
		if processingTimeMs < 100 {
			exp.SupportingFactors = append(exp.SupportingFactors, fmt.Sprintf("Fast processing time: %dms", processingTimeMs))
		} else if processingTimeMs < 500 {
			exp.SupportingFactors = append(exp.SupportingFactors, fmt.Sprintf("Processing time: %dms", processingTimeMs))
		}
	}
	
	return exp
}

// getLayerDescription returns a human-readable description of the layer
func (g *ExplanationGenerator) getLayerDescription(layerUsed string) string {
	switch layerUsed {
	case "layer1":
		return "Layer 1 (Keyword-based classification)"
	case "layer2":
		return "Layer 2 (Embedding-based classification)"
	case "layer3":
		return "Layer 3 (LLM-based classification)"
	default:
		return "Multi-layer classification"
	}
}

// determineProcessingPath determines the processing path used (Phase 2)
func (g *ExplanationGenerator) determineProcessingPath(result *MultiStrategyResult) string {
	if result.Method != "" {
		// Check for fast_path in method name (e.g., "fast_path_keyword")
		if strings.Contains(result.Method, "fast_path") {
			return "fast_path"
		}
		// Check reasoning for fast path indicators
		if strings.Contains(result.Reasoning, "Fast path") {
			return "fast_path"
		}
		if strings.Contains(result.Method, "ml") {
			return "ml_validated"
		}
	}
	return "full_strategy"
}

// extractConfidenceFactors extracts confidence scores from strategies (Phase 2)
func (g *ExplanationGenerator) extractConfidenceFactors(result *MultiStrategyResult) map[string]float64 {
	factors := make(map[string]float64)
	
	// If strategies are available, use them
	if len(result.Strategies) > 0 {
		for _, strategy := range result.Strategies {
			factors[strategy.StrategyName] = strategy.Score
		}
	} else {
		// Fallback: infer confidence factors from available data
		// Use confidence score as overall factor
		factors["overall_confidence"] = result.Confidence
		
		// Infer strategy scores from method and keywords
		if len(result.Keywords) > 0 {
			// Keyword-based classification
			factors["keyword"] = result.Confidence * 0.9
		}
		if result.Method != "" {
			if strings.Contains(result.Method, "fast_path") {
				factors["fast_path"] = result.Confidence
			} else if strings.Contains(result.Method, "ml") {
				factors["ml"] = result.Confidence * 0.85
			} else {
				factors["multi_strategy"] = result.Confidence
			}
		}
	}
	
	return factors
}

// generatePrimaryReason generates the primary reason for classification (Phase 2)
func (g *ExplanationGenerator) generatePrimaryReason(result *MultiStrategyResult) string {
	// Check method first
	if strings.Contains(result.Method, "fast_path") {
		if len(result.Keywords) > 0 {
			return fmt.Sprintf(
				"Strong match based on clear industry indicator '%s' found in business information",
				result.Keywords[0],
			)
		}
		return "Clear industry classification from business information"
	}

	// Check strategy scores to determine dominant strategy
	keywordScore := 0.0
	entityScore := 0.0
	topicScore := 0.0

	// If strategies are available, use them
	if len(result.Strategies) > 0 {
		for _, strategy := range result.Strategies {
			switch strategy.StrategyName {
			case "keyword":
				keywordScore = strategy.Score
			case "entity":
				entityScore = strategy.Score
			case "topic":
				topicScore = strategy.Score
			}
		}
	} else {
		// Fallback: infer strategy scores from available data
		if len(result.Keywords) > 0 {
			keywordScore = result.Confidence * 0.9
		}
		if len(result.Entities) > 0 {
			entityScore = result.Confidence * 0.8
		}
		if len(result.TopicScores) > 0 {
			topicScore = result.Confidence * 0.75
		}
	}

	// Generate reason based on dominant strategy
	if keywordScore > 0.85 {
		if len(result.Keywords) > 0 {
			keywordStr := strings.Join(result.Keywords[:minIntForExplanation(3, len(result.Keywords))], ", ")
			return fmt.Sprintf(
				"Classified as '%s' based on strong keyword matches: %s",
				result.PrimaryIndustry,
				keywordStr,
			)
		}
		return fmt.Sprintf(
			"Classified as '%s' based on strong keyword matching",
			result.PrimaryIndustry,
		)
	}

	if entityScore > 0.80 {
		return fmt.Sprintf(
			"Business entities and services indicate '%s' industry",
			result.PrimaryIndustry,
		)
	}

	if topicScore > 0.75 {
		return fmt.Sprintf(
			"Website content and topic analysis indicates '%s' sector",
			result.PrimaryIndustry,
		)
	}

	// Multi-strategy ensemble or fallback
	if len(result.Keywords) > 0 {
		return fmt.Sprintf(
			"Classification as '%s' based on keyword analysis and business information",
			result.PrimaryIndustry,
		)
	}

	return fmt.Sprintf(
		"Classification as '%s' based on multiple indicators from business information",
		result.PrimaryIndustry,
	)
}

// generateSupportingFactors generates supporting factors for the explanation (Phase 2)
func (g *ExplanationGenerator) generateSupportingFactors(
	result *MultiStrategyResult,
	codes *ClassificationCodesInfo,
	contentQuality float64,
) []string {
	factors := []string{}

	// Factor: Content quality
	if contentQuality > 0.8 {
		factors = append(factors,
			"High-quality business information with comprehensive details")
	} else if contentQuality > 0.6 {
		factors = append(factors,
			"Business information provides good context")
	}

	// Factor: Multiple strategy agreement
	agreementCount := 0
	for _, strategy := range result.Strategies {
		if strategy.Score > 0.70 {
			agreementCount++
		}
	}

	if agreementCount >= 3 {
		factors = append(factors,
			fmt.Sprintf("Multiple classification strategies agree (%d/4 with high confidence)",
				agreementCount))
	} else if agreementCount >= 2 {
		factors = append(factors,
			fmt.Sprintf("Multiple classification strategies support this result (%d strategies)",
				agreementCount))
	}

	// Factor: Code validation
	if codes != nil {
		if len(codes.MCC) > 0 && len(codes.SIC) > 0 && len(codes.NAICS) > 0 {
			// Check if top codes have high confidence
			highConfidenceCodes := 0
			if len(codes.MCC) > 0 && codes.MCC[0].Confidence > 0.85 {
				highConfidenceCodes++
			}
			if len(codes.SIC) > 0 && codes.SIC[0].Confidence > 0.85 {
				highConfidenceCodes++
			}
			if len(codes.NAICS) > 0 && codes.NAICS[0].Confidence > 0.85 {
				highConfidenceCodes++
			}

			if highConfidenceCodes >= 3 {
				factors = append(factors,
					"Industry codes (MCC/SIC/NAICS) strongly align and validate classification")
			} else if highConfidenceCodes >= 2 {
				factors = append(factors,
					"Industry codes identified across multiple classification systems")
			} else {
				factors = append(factors,
					"Industry codes identified for classification validation")
			}
		}
	}

	// Factor: Keywords found
	if len(result.Keywords) > 0 {
		if len(result.Keywords) <= 3 {
			factors = append(factors,
				fmt.Sprintf("Industry-specific terms found: %s",
					strings.Join(result.Keywords, ", ")))
		} else {
			factors = append(factors,
				fmt.Sprintf("Multiple industry-specific terms found (%d total)",
					len(result.Keywords)))
		}
	} else {
		// Even if no keywords, add a factor about business information
		factors = append(factors,
			"Business information analyzed for classification")
	}

	// Factor: Entities found
	if len(result.Entities) > 0 {
		factors = append(factors,
			fmt.Sprintf("Business entities identified (%d entities)",
				len(result.Entities)))
	}

	// Factor: High confidence
	if result.Confidence > 0.90 {
		factors = append(factors,
			fmt.Sprintf("Very high confidence score (%.0f%%)",
				result.Confidence*100))
	} else if result.Confidence > 0.80 {
		factors = append(factors,
			fmt.Sprintf("High confidence score (%.0f%%)",
				result.Confidence*100))
	}

	// Limit to top 5 factors
	if len(factors) > 5 {
		factors = factors[:5]
	}

	return factors
}

// Helper function (renamed to avoid conflict with existing min/minInt functions)
func minIntForExplanation(a, b int) int {
	if a < b {
		return a
	}
	return b
}
