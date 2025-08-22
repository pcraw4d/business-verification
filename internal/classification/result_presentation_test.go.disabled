package classification

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewResultPresentationEngine(t *testing.T) {
	// Arrange & Act
	engine := NewResultPresentationEngine(nil, nil)

	// Assert
	assert.NotNil(t, engine)
	assert.True(t, engine.includeConfidenceBreakdown)
	assert.True(t, engine.includeReliabilityFactors)
	assert.True(t, engine.includeUncertaintyFactors)
	assert.True(t, engine.includeProcessingMetrics)
	assert.True(t, engine.includeRecommendations)
	assert.Equal(t, "detailed", engine.formatOutput)
}

func TestPresentClassificationResult_Basic(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &ClassificationResponse{
		PrimaryClassification: &IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
			Keywords:             []string{"software", "publishing"},
		},
		ConfidenceScore: 0.85,
		ProcessingTime:  100 * time.Millisecond,
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
		Keywords:     "software,technology",
		Description:  "Software development company",
		BusinessType: "Corporation",
		Industry:     "Technology",
	}

	// Act
	enhancedResult := engine.PresentClassificationResult(context.Background(), result, request)

	// Assert
	assert.NotNil(t, enhancedResult)
	assert.Equal(t, "Test Software Company", enhancedResult.BusinessName)
	assert.NotEmpty(t, enhancedResult.RequestID)
	assert.Equal(t, 0.85, enhancedResult.OverallConfidence)
	assert.Equal(t, "high", enhancedResult.ConfidenceLevel)
	assert.NotNil(t, enhancedResult.PrimaryClassification)
	assert.Equal(t, "511210", enhancedResult.PrimaryClassification.IndustryCode)
	assert.Equal(t, "Software Publishers", enhancedResult.PrimaryClassification.IndustryName)
	assert.Equal(t, 0.85, enhancedResult.PrimaryClassification.ConfidenceScore)
	assert.Equal(t, "high", enhancedResult.PrimaryClassification.ConfidenceLevel)
}

func TestPresentClassificationResult_WithAnalysis(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	engine.SetIncludeRecommendations(true)
	result := &ClassificationResponse{
		PrimaryClassification: &IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		ConfidenceScore: 0.85,
		ProcessingTime:  100 * time.Millisecond,
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
		Keywords:     "software,technology",
	}

	// Act
	enhancedResult := engine.PresentClassificationResult(context.Background(), result, request)

	// Assert
	assert.NotNil(t, enhancedResult)
	assert.NotNil(t, enhancedResult.Analysis)
	assert.Equal(t, 0.85, enhancedResult.Analysis.QualityScore)
	assert.Equal(t, 1.0, enhancedResult.Analysis.ConsistencyScore)
	assert.Equal(t, 1.0, enhancedResult.Analysis.DiversityScore)
	assert.Equal(t, 0.85, enhancedResult.Analysis.ReliabilityScore)
}

func TestPresentClassificationResult_WithProcessingMetrics(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	engine.SetIncludeProcessingMetrics(true)
	result := &ClassificationResponse{
		PrimaryClassification: &IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		ConfidenceScore: 0.85,
		ProcessingTime:  100 * time.Millisecond,
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
	}

	// Act
	enhancedResult := engine.PresentClassificationResult(context.Background(), result, request)

	// Assert
	assert.NotNil(t, enhancedResult)
	assert.NotNil(t, enhancedResult.ProcessingMetrics)
	assert.Equal(t, 100*time.Millisecond, enhancedResult.ProcessingMetrics.TotalProcessingTime)
	assert.Equal(t, 1, enhancedResult.ProcessingMetrics.ClassificationsSelected)
	assert.Equal(t, 0.85, enhancedResult.ProcessingMetrics.AverageConfidence)
	assert.Equal(t, 0.0, enhancedResult.ProcessingMetrics.ConfidenceSpread)
}

func TestPresentMultiIndustryResult_Basic(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		PrimaryIndustry: IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		SecondaryIndustry: &IndustryClassification{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.75,
			ClassificationMethod: "description_match",
		},
		Classifications: []IndustryClassification{
			{
				IndustryCode:         "511210",
				IndustryName:         "Software Publishers",
				ConfidenceScore:      0.85,
				ClassificationMethod: "keyword_match",
			},
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.75,
				ClassificationMethod: "description_match",
			},
		},
		OverallConfidence:    0.8,
		ValidationScore:      0.75,
		ClassificationMethod: "multi_industry_enhanced",
		ProcessingTime:       150 * time.Millisecond,
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
		Keywords:     "software,technology",
		Description:  "Software development company",
	}

	// Act
	enhancedResult := engine.PresentMultiIndustryResult(context.Background(), result, request)

	// Assert
	assert.NotNil(t, enhancedResult)
	assert.Equal(t, "Test Software Company", enhancedResult.BusinessName)
	assert.Equal(t, 0.8, enhancedResult.OverallConfidence)
	assert.Equal(t, "high", enhancedResult.ConfidenceLevel)
	assert.NotNil(t, enhancedResult.MultiIndustryResult)
	assert.NotNil(t, enhancedResult.MultiIndustryResult.PrimaryIndustry)
	assert.NotNil(t, enhancedResult.MultiIndustryResult.SecondaryIndustry)
	assert.Equal(t, 2, len(enhancedResult.MultiIndustryResult.AllClassifications))
}

func TestPresentMultiIndustryResult_WithAnalysis(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	engine.SetIncludeRecommendations(true)
	result := &MultiIndustryClassification{
		PrimaryIndustry: IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		Classifications: []IndustryClassification{
			{
				IndustryCode:         "511210",
				IndustryName:         "Software Publishers",
				ConfidenceScore:      0.85,
				ClassificationMethod: "keyword_match",
			},
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.75,
				ClassificationMethod: "description_match",
			},
		},
		OverallConfidence:    0.8,
		ValidationScore:      0.75,
		ClassificationMethod: "multi_industry_enhanced",
		ProcessingTime:       150 * time.Millisecond,
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
	}

	// Act
	enhancedResult := engine.PresentMultiIndustryResult(context.Background(), result, request)

	// Assert
	assert.NotNil(t, enhancedResult)
	assert.NotNil(t, enhancedResult.Analysis)
	assert.Equal(t, 0.775, enhancedResult.Analysis.QualityScore) // (0.8 + 0.75) / 2
	assert.Greater(t, enhancedResult.Analysis.ConsistencyScore, 0.0)
	assert.Greater(t, enhancedResult.Analysis.DiversityScore, 0.0)
	assert.Equal(t, 0.8, enhancedResult.Analysis.ReliabilityScore)
}

func TestPresentMultiIndustryResult_WithProcessingMetrics(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	engine.SetIncludeProcessingMetrics(true)
	result := &MultiIndustryClassification{
		PrimaryIndustry: IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		Classifications: []IndustryClassification{
			{
				IndustryCode:         "511210",
				IndustryName:         "Software Publishers",
				ConfidenceScore:      0.85,
				ClassificationMethod: "keyword_match",
			},
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.75,
				ClassificationMethod: "description_match",
			},
		},
		OverallConfidence:    0.8,
		ValidationScore:      0.75,
		ClassificationMethod: "multi_industry_enhanced",
		ProcessingTime:       150 * time.Millisecond,
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
	}

	// Act
	enhancedResult := engine.PresentMultiIndustryResult(context.Background(), result, request)

	// Assert
	assert.NotNil(t, enhancedResult)
	assert.NotNil(t, enhancedResult.ProcessingMetrics)
	assert.Equal(t, 150*time.Millisecond, enhancedResult.ProcessingMetrics.TotalProcessingTime)
	assert.Equal(t, 2, enhancedResult.ProcessingMetrics.ClassificationsSelected)
	assert.Equal(t, 0.8, enhancedResult.ProcessingMetrics.AverageConfidence)
	assert.InDelta(t, 0.1, enhancedResult.ProcessingMetrics.ConfidenceSpread, 0.001) // 0.85 - 0.75
	assert.Equal(t, 2, len(enhancedResult.ProcessingMetrics.MethodDistribution))
}

func TestCreateEnhancedIndustryResult(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "keyword_match",
		Keywords:             []string{"software", "publishing"},
		Description:          "Software publishing industry",
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
		Keywords:     "software,technology",
	}

	// Act
	enhanced := engine.createEnhancedIndustryResult(classification, request)

	// Assert
	assert.NotNil(t, enhanced)
	assert.Equal(t, "511210", enhanced.IndustryCode)
	assert.Equal(t, "Software Publishers", enhanced.IndustryName)
	assert.Equal(t, 0.85, enhanced.ConfidenceScore)
	assert.Equal(t, "keyword_match", enhanced.ClassificationMethod)
	assert.Equal(t, "high", enhanced.ConfidenceLevel)
	assert.Equal(t, []string{"software", "publishing"}, enhanced.Keywords)
	assert.Equal(t, "Software publishing industry", enhanced.Description)
	assert.Equal(t, "51", enhanced.IndustryCategory)
	assert.Equal(t, "5112", enhanced.IndustrySubcategory)
	assert.Equal(t, "enhanced_confidence_scoring", enhanced.ScoringMethod)
}

func TestCreateEnhancedIndustryResult_WithConfidenceBreakdown(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	engine.SetIncludeConfidenceBreakdown(true)
	classification := IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "keyword_match",
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
	}

	// Act
	enhanced := engine.createEnhancedIndustryResult(classification, request)

	// Assert
	assert.NotNil(t, enhanced)
	assert.NotNil(t, enhanced.ConfidenceBreakdown)
	assert.Equal(t, 0.85, enhanced.ConfidenceBreakdown.OverallScore)
	assert.Equal(t, 0.85, enhanced.ConfidenceBreakdown.BaseConfidence)
	assert.NotNil(t, enhanced.ConfidenceBreakdown.Weights)
	assert.NotNil(t, enhanced.ConfidenceBreakdown.ComponentContributions)
}

func TestCreateEnhancedMultiIndustryResult(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		PrimaryIndustry: IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		SecondaryIndustry: &IndustryClassification{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.75,
			ClassificationMethod: "description_match",
		},
		Classifications: []IndustryClassification{
			{
				IndustryCode:         "511210",
				IndustryName:         "Software Publishers",
				ConfidenceScore:      0.85,
				ClassificationMethod: "keyword_match",
			},
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.75,
				ClassificationMethod: "description_match",
			},
		},
		OverallConfidence:    0.8,
		ValidationScore:      0.75,
		ClassificationMethod: "multi_industry_enhanced",
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
	}

	// Act
	enhanced := engine.createEnhancedMultiIndustryResult(result, request)

	// Assert
	assert.NotNil(t, enhanced)
	assert.NotNil(t, enhanced.PrimaryIndustry)
	assert.NotNil(t, enhanced.SecondaryIndustry)
	assert.Nil(t, enhanced.TertiaryIndustry)
	assert.Equal(t, 2, len(enhanced.AllClassifications))
	assert.Equal(t, "multi_industry_enhanced", enhanced.SelectionMethod)
	assert.Equal(t, 0.75, enhanced.ValidationScore)
	assert.Equal(t, 0.8, enhanced.OverallConfidence)
	assert.Equal(t, "high", enhanced.ConfidenceLevel)
}

func TestCreateConfidenceBreakdown(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "keyword_match",
	}

	// Act
	breakdown := engine.createConfidenceBreakdown(classification)

	// Assert
	assert.NotNil(t, breakdown)
	assert.Equal(t, 0.85, breakdown.OverallScore)
	assert.Equal(t, 0.85, breakdown.BaseConfidence)
	assert.Equal(t, 0.0, breakdown.KeywordScore)
	assert.Equal(t, 0.0, breakdown.DescriptionScore)
	assert.NotNil(t, breakdown.Weights)
	assert.NotNil(t, breakdown.ComponentContributions)
	assert.Equal(t, 0.2125, breakdown.ComponentContributions["base_confidence"]) // 0.85 * 0.25
}

func TestCreateClassificationAnalysis(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &ClassificationResponse{
		PrimaryClassification: &IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		ConfidenceScore: 0.85,
		ProcessingTime:  100 * time.Millisecond,
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
		Keywords:     "software,technology",
	}

	// Act
	analysis := engine.createClassificationAnalysis(result, request)

	// Assert
	assert.NotNil(t, analysis)
	assert.Equal(t, 0.85, analysis.QualityScore)
	assert.Equal(t, 1.0, analysis.ConsistencyScore)
	assert.Equal(t, 1.0, analysis.DiversityScore)
	assert.Equal(t, 0.85, analysis.ReliabilityScore)
	assert.NotNil(t, analysis.QualityFactors)
	assert.NotNil(t, analysis.QualityIssues)
	assert.NotNil(t, analysis.ConsistencyFactors)
	assert.NotNil(t, analysis.DiversityFactors)
	assert.NotNil(t, analysis.ReliabilityFactors)
	assert.NotNil(t, analysis.Recommendations)
	assert.NotNil(t, analysis.NextSteps)
}

func TestCreateMultiIndustryAnalysis(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		PrimaryIndustry: IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		Classifications: []IndustryClassification{
			{
				IndustryCode:         "511210",
				IndustryName:         "Software Publishers",
				ConfidenceScore:      0.85,
				ClassificationMethod: "keyword_match",
			},
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.75,
				ClassificationMethod: "description_match",
			},
		},
		OverallConfidence:    0.8,
		ValidationScore:      0.75,
		ClassificationMethod: "multi_industry_enhanced",
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
	}

	// Act
	analysis := engine.createMultiIndustryAnalysis(result, request)

	// Assert
	assert.NotNil(t, analysis)
	assert.Equal(t, 0.775, analysis.QualityScore) // (0.8 + 0.75) / 2
	assert.Greater(t, analysis.ConsistencyScore, 0.0)
	assert.Greater(t, analysis.DiversityScore, 0.0)
	assert.Equal(t, 0.8, analysis.ReliabilityScore)
	assert.NotNil(t, analysis.QualityFactors)
	assert.NotNil(t, analysis.QualityIssues)
	assert.NotNil(t, analysis.ConsistencyFactors)
	assert.NotNil(t, analysis.DiversityFactors)
	assert.NotNil(t, analysis.ReliabilityFactors)
	assert.NotNil(t, analysis.Recommendations)
	assert.NotNil(t, analysis.NextSteps)
}

func TestCreateProcessingMetrics(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &ClassificationResponse{
		PrimaryClassification: &IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		ConfidenceScore: 0.85,
		ProcessingTime:  100 * time.Millisecond,
	}
	start := time.Now()

	// Act
	metrics := engine.createProcessingMetrics(result, start)

	// Assert
	assert.NotNil(t, metrics)
	assert.Equal(t, 100*time.Millisecond, metrics.TotalProcessingTime)
	assert.Equal(t, 70*time.Millisecond, metrics.ClassificationTime)
	assert.Equal(t, 20*time.Millisecond, metrics.ScoringTime)
	assert.Equal(t, 5*time.Millisecond, metrics.SelectionTime)
	assert.Greater(t, metrics.PresentationTime, 0*time.Millisecond)
	assert.Equal(t, 1, metrics.ClassificationsGenerated)
	assert.Equal(t, 0, metrics.ClassificationsFiltered)
	assert.Equal(t, 1, metrics.ClassificationsSelected)
	assert.Equal(t, 0.85, metrics.AverageConfidence)
	assert.Equal(t, 0.0, metrics.ConfidenceSpread)
	assert.Equal(t, 1, metrics.MethodDistribution["keyword_match"])
}

func TestCreateMultiIndustryProcessingMetrics(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		PrimaryIndustry: IndustryClassification{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_match",
		},
		Classifications: []IndustryClassification{
			{
				IndustryCode:         "511210",
				IndustryName:         "Software Publishers",
				ConfidenceScore:      0.85,
				ClassificationMethod: "keyword_match",
			},
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.75,
				ClassificationMethod: "description_match",
			},
		},
		OverallConfidence:    0.8,
		ValidationScore:      0.75,
		ClassificationMethod: "multi_industry_enhanced",
		ProcessingTime:       150 * time.Millisecond,
	}
	start := time.Now()

	// Act
	metrics := engine.createMultiIndustryProcessingMetrics(result, start)

	// Assert
	assert.NotNil(t, metrics)
	assert.Equal(t, 150*time.Millisecond, metrics.TotalProcessingTime)
	assert.Equal(t, 90*time.Millisecond, metrics.ClassificationTime)
	assert.Equal(t, 37*time.Millisecond+500*time.Microsecond, metrics.ScoringTime)
	assert.Equal(t, 15*time.Millisecond, metrics.SelectionTime)
	assert.Greater(t, metrics.PresentationTime, 0*time.Millisecond)
	assert.Equal(t, 4, metrics.ClassificationsGenerated)
	assert.Equal(t, 2, metrics.ClassificationsFiltered)
	assert.Equal(t, 2, metrics.ClassificationsSelected)
	assert.Equal(t, 0.8, metrics.AverageConfidence)
	assert.InDelta(t, 0.1, metrics.ConfidenceSpread, 0.001)
	assert.Equal(t, 1, metrics.MethodDistribution["keyword_match"])
	assert.Equal(t, 1, metrics.MethodDistribution["description_match"])
}

func TestResultPresentationEngine_DetermineConfidenceLevel(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, "very_high", engine.determineConfidenceLevel(0.95))
	assert.Equal(t, "high", engine.determineConfidenceLevel(0.85))
	assert.Equal(t, "medium_high", engine.determineConfidenceLevel(0.75))
	assert.Equal(t, "medium", engine.determineConfidenceLevel(0.65))
	assert.Equal(t, "medium_low", engine.determineConfidenceLevel(0.55))
	assert.Equal(t, "low", engine.determineConfidenceLevel(0.35))
	assert.Equal(t, "very_low", engine.determineConfidenceLevel(0.15))
}

func TestExtractIndustryCategory(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, "51", engine.extractIndustryCategory("511210"))
	assert.Equal(t, "54", engine.extractIndustryCategory("541511"))
	assert.Equal(t, "", engine.extractIndustryCategory("5"))
	assert.Equal(t, "", engine.extractIndustryCategory(""))
}

func TestExtractIndustrySubcategory(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, "5112", engine.extractIndustrySubcategory("511210"))
	assert.Equal(t, "5415", engine.extractIndustrySubcategory("541511"))
	assert.Equal(t, "", engine.extractIndustrySubcategory("511"))
	assert.Equal(t, "", engine.extractIndustrySubcategory(""))
}

func TestCalculateQualityScore(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &ClassificationResponse{
		ConfidenceScore: 0.85,
	}

	// Act
	score := engine.calculateQualityScore(result)

	// Assert
	assert.Equal(t, 0.85, score)
}

func TestCalculateMultiIndustryQualityScore(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		OverallConfidence: 0.8,
		ValidationScore:   0.75,
	}

	// Act
	score := engine.calculateMultiIndustryQualityScore(result)

	// Assert
	assert.Equal(t, 0.775, score) // (0.8 + 0.75) / 2
}

func TestResultPresentationEngine_CalculateConsistencyScore(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		OverallConfidence: 0.8,
		Classifications: []IndustryClassification{
			{ConfidenceScore: 0.85},
			{ConfidenceScore: 0.75},
		},
	}

	// Act
	score := engine.calculateConsistencyScore(result)

	// Assert
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestResultPresentationEngine_CalculateConsistencyScore_SingleClassification(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		OverallConfidence: 0.8,
		Classifications: []IndustryClassification{
			{ConfidenceScore: 0.85},
		},
	}

	// Act
	score := engine.calculateConsistencyScore(result)

	// Assert
	assert.Equal(t, 1.0, score)
}

func TestResultPresentationEngine_CalculateDiversityScore(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		Classifications: []IndustryClassification{
			{IndustryCode: "511210"},
			{IndustryCode: "541511"},
		},
	}

	// Act
	score := engine.calculateDiversityScore(result)

	// Assert
	assert.Equal(t, 1.0, score) // Different major categories (51 vs 54)
}

func TestCalculateDiversityScore_SameCategory(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		Classifications: []IndustryClassification{
			{IndustryCode: "511210"},
			{IndustryCode: "511211"},
		},
	}

	// Act
	score := engine.calculateDiversityScore(result)

	// Assert
	assert.Equal(t, 0.5, score) // Same major category (51)
}

func TestResultPresentationEngine_CalculateDiversityScore_SingleClassification(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		Classifications: []IndustryClassification{
			{IndustryCode: "511210"},
		},
	}

	// Act
	score := engine.calculateDiversityScore(result)

	// Assert
	assert.Equal(t, 1.0, score)
}

func TestCalculateConfidenceSpread(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	confidences := []float64{0.85, 0.75, 0.90}

	// Act
	spread := engine.calculateConfidenceSpread(confidences)

	// Assert
	assert.InDelta(t, 0.15, spread, 0.001) // 0.90 - 0.75
}

func TestCalculateConfidenceSpread_SingleValue(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	confidences := []float64{0.85}

	// Act
	spread := engine.calculateConfidenceSpread(confidences)

	// Assert
	assert.Equal(t, 0.0, spread)
}

func TestGenerateRequestID(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)

	// Act
	id1 := engine.generateRequestID()
	id2 := engine.generateRequestID()

	// Assert
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "req_")
	assert.Contains(t, id2, "req_")
}

func TestConfigurationMethods(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)

	// Act & Assert
	engine.SetFormat("summary")
	assert.Equal(t, "summary", engine.formatOutput)

	engine.SetIncludeConfidenceBreakdown(false)
	assert.False(t, engine.includeConfidenceBreakdown)

	engine.SetIncludeReliabilityFactors(false)
	assert.False(t, engine.includeReliabilityFactors)

	engine.SetIncludeUncertaintyFactors(false)
	assert.False(t, engine.includeUncertaintyFactors)

	engine.SetIncludeProcessingMetrics(false)
	assert.False(t, engine.includeProcessingMetrics)

	engine.SetIncludeRecommendations(false)
	assert.False(t, engine.includeRecommendations)
}

func TestCreateMetadata(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &ClassificationResponse{
		PrimaryClassification: &IndustryClassification{
			IndustryCode: "511210",
		},
	}
	request := &ClassificationRequest{
		BusinessName: "Test Company",
	}

	// Act
	metadata := engine.createMetadata(result, request)

	// Assert
	assert.NotNil(t, metadata)
	assert.Equal(t, "detailed", metadata["presentation_format"])
	assert.Equal(t, "1.0", metadata["version"])
	assert.NotNil(t, metadata["timestamp"])
}

func TestCreateMultiIndustryMetadata(t *testing.T) {
	// Arrange
	engine := NewResultPresentationEngine(nil, nil)
	result := &MultiIndustryClassification{
		Classifications: []IndustryClassification{
			{IndustryCode: "511210"},
			{IndustryCode: "541511"},
		},
	}
	request := &ClassificationRequest{
		BusinessName: "Test Company",
	}

	// Act
	metadata := engine.createMultiIndustryMetadata(result, request)

	// Assert
	assert.NotNil(t, metadata)
	assert.Equal(t, "detailed", metadata["presentation_format"])
	assert.Equal(t, "1.0", metadata["version"])
	assert.NotNil(t, metadata["timestamp"])
	assert.Equal(t, 2, metadata["num_classifications"])
}
