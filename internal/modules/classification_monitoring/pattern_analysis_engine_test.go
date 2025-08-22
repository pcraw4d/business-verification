package classification_monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPatternAnalysisEngine(t *testing.T) {
	logger := zap.NewNop()
	config := &PatternAnalysisConfig{
		EnableDeepAnalysis:             true,
		EnablePredictiveAnalysis:       true,
		EnableRootCauseAnalysis:        true,
		PatternRetentionPeriod:         24 * time.Hour,
		AnalysisWindowSize:             1 * time.Hour,
		MinPatternOccurrences:          5,
		ConfidenceThreshold:            0.7,
		EnableRealTimeAnalysis:         true,
		MaxPatternsPerCategory:         10,
		EnableCrossDimensionalAnalysis: true,
	}

	engine := NewPatternAnalysisEngine(config, logger)

	assert.NotNil(t, engine)
	assert.Equal(t, config, engine.config)
	assert.Equal(t, logger, engine.logger)
	assert.NotNil(t, engine.patterns)
	assert.NotNil(t, engine.patternHistory)
	assert.NotNil(t, engine.rootCauseAnalyzer)
	assert.NotNil(t, engine.recommendationEngine)
	assert.NotNil(t, engine.predictiveAnalyzer)
}

func TestPatternAnalysisEngine_AnalyzeMisclassifications(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Create test misclassifications
	misclassifications := []*MisclassificationRecord{
		{
			ID:                     "test-1",
			Timestamp:              time.Now(),
			BusinessName:           "Test Business 1",
			ExpectedClassification: "Technology",
			ActualClassification:   "Finance",
			ConfidenceScore:        0.9,
			ClassificationMethod:   "ml",
			InputData:              map[string]interface{}{"text": "software company"},
			ErrorType:              "misclassification",
			Severity:               "high",
		},
		{
			ID:                     "test-2",
			Timestamp:              time.Now(),
			BusinessName:           "Test Business 2",
			ExpectedClassification: "Technology",
			ActualClassification:   "Finance",
			ConfidenceScore:        0.85,
			ClassificationMethod:   "ml",
			InputData:              map[string]interface{}{"text": "tech startup"},
			ErrorType:              "misclassification",
			Severity:               "high",
		},
		{
			ID:                     "test-3",
			Timestamp:              time.Now(),
			BusinessName:           "Test Business 3",
			ExpectedClassification: "Technology",
			ActualClassification:   "Finance",
			ConfidenceScore:        0.88,
			ClassificationMethod:   "keyword",
			InputData:              map[string]interface{}{"text": "software development"},
			ErrorType:              "misclassification",
			Severity:               "high",
		},
	}

	ctx := context.Background()
	result, err := engine.AnalyzeMisclassifications(ctx, misclassifications)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(misclassifications), result.PatternsFound)
	assert.Greater(t, result.NewPatterns, 0)
	assert.NotEmpty(t, result.Recommendations)
	assert.NotNil(t, result.Summary)
}

func TestPatternAnalysisEngine_GetPatterns(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Add some test patterns
	pattern := &MisclassificationPattern{
		ID:              "test-pattern",
		Name:            "Test Pattern",
		Description:     "Test pattern description",
		PatternType:     PatternTypeConfidence,
		Category:        PatternCategoryModelPerformance,
		Severity:        PatternSeverityHigh,
		Confidence:      0.8,
		OccurrenceCount: 10,
		FirstSeen:       time.Now(),
		LastSeen:        time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	engine.patterns["test-pattern"] = pattern

	patterns := engine.GetPatterns()

	assert.Len(t, patterns, 1)
	assert.Contains(t, patterns, "test-pattern")
	assert.Equal(t, pattern, patterns["test-pattern"])
}

func TestPatternAnalysisEngine_GetPatternsByType(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Add test patterns of different types
	confidencePattern := &MisclassificationPattern{
		ID:          "confidence-pattern",
		PatternType: PatternTypeConfidence,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	temporalPattern := &MisclassificationPattern{
		ID:          "temporal-pattern",
		PatternType: PatternTypeTemporal,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	engine.patterns["confidence-pattern"] = confidencePattern
	engine.patterns["temporal-pattern"] = temporalPattern

	confidencePatterns := engine.GetPatternsByType(PatternTypeConfidence)
	temporalPatterns := engine.GetPatternsByType(PatternTypeTemporal)

	assert.Len(t, confidencePatterns, 1)
	assert.Len(t, temporalPatterns, 1)
	assert.Equal(t, "confidence-pattern", confidencePatterns[0].ID)
	assert.Equal(t, "temporal-pattern", temporalPatterns[0].ID)
}

func TestPatternAnalysisEngine_GetPatternsBySeverity(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Add test patterns of different severities
	highSeverityPattern := &MisclassificationPattern{
		ID:        "high-severity-pattern",
		Severity:  PatternSeverityHigh,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	criticalSeverityPattern := &MisclassificationPattern{
		ID:        "critical-severity-pattern",
		Severity:  PatternSeverityCritical,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	engine.patterns["high-severity-pattern"] = highSeverityPattern
	engine.patterns["critical-severity-pattern"] = criticalSeverityPattern

	highSeverityPatterns := engine.GetPatternsBySeverity(PatternSeverityHigh)
	criticalSeverityPatterns := engine.GetPatternsBySeverity(PatternSeverityCritical)

	assert.Len(t, highSeverityPatterns, 1)
	assert.Len(t, criticalSeverityPatterns, 1)
	assert.Equal(t, "high-severity-pattern", highSeverityPatterns[0].ID)
	assert.Equal(t, "critical-severity-pattern", criticalSeverityPatterns[0].ID)
}

func TestPatternAnalysisEngine_GetPatternHistory(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Add test history
	historyResult := &PatternAnalysisResult{
		ID:            "test-history",
		AnalysisTime:  time.Now(),
		PatternsFound: 5,
		NewPatterns:   3,
	}

	engine.patternHistory = append(engine.patternHistory, historyResult)

	history := engine.GetPatternHistory()

	assert.Len(t, history, 1)
	assert.Equal(t, "test-history", history[0].ID)
	assert.Equal(t, 5, history[0].PatternsFound)
	assert.Equal(t, 3, history[0].NewPatterns)
}

func TestPatternAnalysisEngine_analyzeTemporalPatterns(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Create misclassifications at different times
	now := time.Now()
	misclassifications := []*MisclassificationRecord{
		{
			ID:                   "test-1",
			Timestamp:            now,
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-2",
			Timestamp:            now.Add(1 * time.Hour),
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-3",
			Timestamp:            now.Add(2 * time.Hour),
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
	}

	patterns, err := engine.analyzeTemporalPatterns(misclassifications)

	require.NoError(t, err)
	assert.NotNil(t, patterns)
	// Should detect temporal patterns if they meet the minimum occurrence threshold
}

func TestPatternAnalysisEngine_analyzeConfidencePatterns(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Create misclassifications with high confidence
	misclassifications := []*MisclassificationRecord{
		{
			ID:                   "test-1",
			ConfidenceScore:      0.9,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-2",
			ConfidenceScore:      0.85,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-3",
			ConfidenceScore:      0.88,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-4",
			ConfidenceScore:      0.92,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-5",
			ConfidenceScore:      0.87,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-6",
			ConfidenceScore:      0.89,
			ClassificationMethod: "ml",
		},
	}

	patterns, err := engine.analyzeConfidencePatterns(misclassifications)

	require.NoError(t, err)
	assert.NotNil(t, patterns)
	// Should detect high confidence misclassification patterns
	assert.Greater(t, len(patterns), 0)
}

func TestPatternAnalysisEngine_analyzeInputPatterns(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Create misclassifications with different input lengths
	misclassifications := []*MisclassificationRecord{
		{
			ID:                   "test-1",
			InputData:            map[string]interface{}{"text": "short"},
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-2",
			InputData:            map[string]interface{}{"text": "short"},
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-3",
			InputData:            map[string]interface{}{"text": "short"},
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-4",
			InputData:            map[string]interface{}{"text": "short"},
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-5",
			InputData:            map[string]interface{}{"text": "short"},
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-6",
			InputData:            map[string]interface{}{"text": "short"},
			ConfidenceScore:      0.8,
			ClassificationMethod: "ml",
		},
	}

	patterns, err := engine.analyzeInputPatterns(misclassifications)

	require.NoError(t, err)
	assert.NotNil(t, patterns)
	// Should detect input length patterns
	assert.Greater(t, len(patterns), 0)
}

func TestPatternAnalysisEngine_analyzeCrossDimensionalPatterns(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Create misclassifications with different method-confidence combinations
	misclassifications := []*MisclassificationRecord{
		{
			ID:                   "test-1",
			ConfidenceScore:      0.9,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-2",
			ConfidenceScore:      0.85,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-3",
			ConfidenceScore:      0.88,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-4",
			ConfidenceScore:      0.92,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-5",
			ConfidenceScore:      0.87,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-6",
			ConfidenceScore:      0.89,
			ClassificationMethod: "ml",
		},
	}

	patterns, err := engine.analyzeCrossDimensionalPatterns(misclassifications)

	require.NoError(t, err)
	assert.NotNil(t, patterns)
	// Should detect cross-dimensional patterns
	assert.Greater(t, len(patterns), 0)
}

func TestPatternAnalysisEngine_analyzeRootCauses(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Create misclassifications with high confidence
	misclassifications := []*MisclassificationRecord{
		{
			ID:                   "test-1",
			ConfidenceScore:      0.9,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-2",
			ConfidenceScore:      0.85,
			ClassificationMethod: "ml",
		},
		{
			ID:                   "test-3",
			ConfidenceScore:      0.88,
			ClassificationMethod: "ml",
		},
	}

	rootCauses := engine.analyzeRootCauses(misclassifications)

	assert.NotNil(t, rootCauses)
	assert.Greater(t, len(rootCauses), 0)

	// Check that root causes have required fields
	for _, rootCause := range rootCauses {
		assert.NotEmpty(t, rootCause.ID)
		assert.NotEmpty(t, rootCause.Type)
		assert.NotEmpty(t, rootCause.Description)
		assert.GreaterOrEqual(t, rootCause.Confidence, 0.0)
		assert.LessOrEqual(t, rootCause.Confidence, 1.0)
	}
}

func TestPatternAnalysisEngine_HelperMethods(t *testing.T) {
	logger := zap.NewNop()
	engine := NewPatternAnalysisEngine(nil, logger)

	// Test extractInputText
	inputData := map[string]interface{}{
		"text": "test text",
		"name": "test name",
	}
	text := engine.extractInputText(inputData)
	assert.Equal(t, "test text", text)

	// Test extractKeywords
	keywords := engine.extractKeywords("This is a test text with important keywords")
	assert.NotEmpty(t, keywords)
	assert.Contains(t, keywords, "test")
	assert.Contains(t, keywords, "text")
	assert.Contains(t, keywords, "important")
	assert.Contains(t, keywords, "keywords")

	// Test extractPhrases
	phrases := engine.extractPhrases("This is a test")
	assert.Len(t, phrases, 3)
	assert.Contains(t, phrases, "This is")
	assert.Contains(t, phrases, "is a")
	assert.Contains(t, phrases, "a test")

	// Test getConfidenceLevel
	assert.Equal(t, "high", engine.getConfidenceLevel(0.9))
	assert.Equal(t, "medium", engine.getConfidenceLevel(0.7))
	assert.Equal(t, "low", engine.getConfidenceLevel(0.3))

	// Test calculateSeverity
	assert.Equal(t, PatternSeverityCritical, engine.calculateSeverity(30, 100))
	assert.Equal(t, PatternSeverityHigh, engine.calculateSeverity(20, 100))
	assert.Equal(t, PatternSeverityMedium, engine.calculateSeverity(10, 100))
	assert.Equal(t, PatternSeverityLow, engine.calculateSeverity(3, 100))
}

func TestRecommendationEngine_GenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	engine := NewRecommendationEngine(logger)

	patterns := []*MisclassificationPattern{
		{
			PatternType: PatternTypeConfidence,
			Severity:    PatternSeverityCritical,
		},
		{
			PatternType: PatternTypeTemporal,
			Severity:    PatternSeverityHigh,
		},
		{
			PatternType: PatternTypeSemantic,
			Severity:    PatternSeverityMedium,
		},
	}

	recommendations := engine.GenerateRecommendations(patterns)

	assert.NotEmpty(t, recommendations)
	assert.Len(t, recommendations, 3)

	// Check that recommendations have required fields
	for _, rec := range recommendations {
		assert.NotEmpty(t, rec.ID)
		assert.NotEmpty(t, rec.Type)
		assert.NotEmpty(t, rec.Priority)
		assert.NotEmpty(t, rec.Title)
		assert.NotEmpty(t, rec.Description)
		assert.NotEmpty(t, rec.Actions)
		assert.NotEmpty(t, rec.Impact)
		assert.NotEmpty(t, rec.Effort)
	}
}

func TestPredictiveAnalyzer_NewPredictiveAnalyzer(t *testing.T) {
	logger := zap.NewNop()
	config := &PatternAnalysisConfig{
		EnablePredictiveAnalysis: true,
	}

	analyzer := NewPredictiveAnalyzer(config, logger)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config, analyzer.config)
	assert.Equal(t, logger, analyzer.logger)
}
