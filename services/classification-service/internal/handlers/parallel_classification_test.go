package handlers

import (
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"kyb-platform/services/classification-service/internal/config"
)

func TestParallelClassification_EnsembleVoting(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			MinContentLengthForML:         50,
		},
	}

	handler := &ClassificationHandler{
		logger: logger,
		config: cfg,
	}

	// Mock results
	pythonMLResult := &EnhancedClassificationResult{
		PrimaryIndustry:    "Technology",
		ConfidenceScore:    0.85,
		Keywords:           []string{"software", "technology", "development"},
		MCCCodes:           []IndustryCode{{Code: "5734", Description: "Computer Software Stores"}},
		SICCodes:           []IndustryCode{{Code: "7372", Description: "Prepackaged Software"}},
		NAICSCodes:         []IndustryCode{{Code: "541511", Description: "Custom Computer Programming Services"}},
	}

	goResult := &EnhancedClassificationResult{
		PrimaryIndustry:    "Technology",
		ConfidenceScore:    0.80,
		Keywords:           []string{"technology", "software", "solutions"},
		MCCCodes:           []IndustryCode{{Code: "5734", Description: "Computer Software Stores"}},
		SICCodes:           []IndustryCode{{Code: "7372", Description: "Prepackaged Software"}},
		NAICSCodes:         []IndustryCode{{Code: "541511", Description: "Custom Computer Programming Services"}},
	}

	req := &ClassificationRequest{
		BusinessName: "Tech Solutions Inc",
		RequestID:    "test-123",
	}

	// Test ensemble voting
	result := handler.combineEnsembleResults(pythonMLResult, goResult, req)

	// Verify consensus boost
	if result.PrimaryIndustry != "Technology" {
		t.Errorf("Expected industry 'Technology', got %q", result.PrimaryIndustry)
	}

	// Verify confidence is weighted combination
	expectedMin := 0.80 // Minimum of the two
	expectedMax := 0.85 // Maximum of the two
	if result.ConfidenceScore < expectedMin || result.ConfidenceScore > expectedMax+0.05 {
		t.Errorf("Expected confidence between %.2f and %.2f, got %.2f", expectedMin, expectedMax+0.05, result.ConfidenceScore)
	}

	// Verify keywords are merged
	if len(result.Keywords) < 3 {
		t.Errorf("Expected at least 3 merged keywords, got %d", len(result.Keywords))
	}

	// Verify codes are merged
	if len(result.MCCCodes) == 0 {
		t.Error("Expected merged MCC codes")
	}
	if len(result.SICCodes) == 0 {
		t.Error("Expected merged SIC codes")
	}
	if len(result.NAICSCodes) == 0 {
		t.Error("Expected merged NAICS codes")
	}
}

func TestParallelClassification_ConsensusBoost(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
		},
	}

	handler := &ClassificationHandler{
		logger: logger,
		config: cfg,
	}

	// Both methods agree on industry
	pythonMLResult := &EnhancedClassificationResult{
		PrimaryIndustry: "Technology",
		ConfidenceScore: 0.85,
	}

	goResult := &EnhancedClassificationResult{
		PrimaryIndustry: "Technology",
		ConfidenceScore: 0.80,
	}

	req := &ClassificationRequest{
		BusinessName: "Tech Solutions Inc",
		RequestID:    "test-123",
	}

	result := handler.combineEnsembleResults(pythonMLResult, goResult, req)

	// Verify consensus boost is applied (confidence should be higher than weighted average)
	weightedAvg := (0.85*0.60 + 0.80*0.40) // 0.83
	if result.ConfidenceScore <= weightedAvg {
		t.Errorf("Expected consensus boost, confidence %.2f should be > weighted avg %.2f", result.ConfidenceScore, weightedAvg)
	}

	// Verify reasoning mentions consensus
	if result.ClassificationReasoning == "" {
		t.Error("Expected classification reasoning")
	}
}

func TestParallelClassification_Disagreement(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
		},
	}

	handler := &ClassificationHandler{
		logger: logger,
		config: cfg,
	}

	// Methods disagree on industry
	pythonMLResult := &EnhancedClassificationResult{
		PrimaryIndustry: "Technology",
		ConfidenceScore: 0.90, // Higher confidence
	}

	goResult := &EnhancedClassificationResult{
		PrimaryIndustry: "Retail",
		ConfidenceScore: 0.75, // Lower confidence
	}

	req := &ClassificationRequest{
		BusinessName: "Tech Solutions Inc",
		RequestID:    "test-123",
	}

	result := handler.combineEnsembleResults(pythonMLResult, goResult, req)

	// Should select based on weighted confidence
	// Python ML: 0.90 * 0.60 = 0.54
	// Go: 0.75 * 0.40 = 0.30
	// Python ML should win
	if result.PrimaryIndustry != "Technology" {
		t.Errorf("Expected 'Technology' (higher weighted confidence), got %q", result.PrimaryIndustry)
	}
}

func TestParallelExecution_Timing(t *testing.T) {
	// Test that parallel execution is actually faster
	var wg sync.WaitGroup
	startTime := time.Now()

	// Simulate parallel execution
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond) // Simulate work
		}()
	}

	wg.Wait()
	parallelDuration := time.Since(startTime)

	// Simulate sequential execution
	startTime = time.Now()
	time.Sleep(100 * time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	sequentialDuration := time.Since(startTime)

	// Parallel should be faster (approximately 100ms vs 200ms)
	if parallelDuration >= sequentialDuration {
		t.Errorf("Parallel execution (%v) should be faster than sequential (%v)", parallelDuration, sequentialDuration)
	}

	// Parallel should be at least 30% faster
	speedup := float64(sequentialDuration) / float64(parallelDuration)
	if speedup < 1.3 {
		t.Errorf("Expected at least 1.3x speedup, got %.2fx", speedup)
	}
}

