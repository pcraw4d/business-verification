package feedback

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Note: MockFeedbackRepository is already defined in service_test.go

func TestAdvancedSecurityFeedbackSystem_CollectSecurityFeedback(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AdvancedSecurityConfig{
		MinFeedbackThreshold:          10,
		MaxFeedbackAge:                24 * time.Hour,
		CollectionInterval:            1 * time.Hour,
		BatchProcessingSize:           50,
		SecurityViolationThreshold:    0.1,
		TrustScoreThreshold:           0.8,
		VerificationAccuracyThreshold: 0.9,
		ImprovementInterval:           24 * time.Hour,
		MaxImprovementRuns:            5,
		ImprovementTimeout:            30 * time.Minute,
	}

	mockRepo := &MockFeedbackRepository{}
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 10,
	}, logger)
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          10,
		MaxLearningBatchSize:          100,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	system := NewAdvancedSecurityFeedbackSystem(
		config,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		mockRepo,
	)

	ctx := context.Background()

	// Test security feedback collection
	result, err := system.CollectSecurityFeedback(ctx)
	if err != nil {
		t.Fatalf("Failed to collect security feedback: %v", err)
	}

	// Validate collection result
	if result == nil {
		t.Fatal("Collection result should not be nil")
	}

	if result.TotalProcessed == 0 {
		t.Error("Should have processed some feedback")
	}

	if result.CollectionTime <= 0 {
		t.Error("Collection time should be positive")
	}

	// Check that metrics were updated
	metrics := system.GetSecurityMetrics()
	if metrics.TotalFeedbackCollected == 0 {
		t.Error("Total feedback collected should be updated")
	}

	t.Logf("Security feedback collection completed: %d feedback processed in %v",
		result.TotalProcessed, result.CollectionTime)
}

func TestAdvancedSecurityFeedbackSystem_AnalyzeSecurityFeedback(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AdvancedSecurityConfig{
		MinFeedbackThreshold:          10,
		MaxFeedbackAge:                24 * time.Hour,
		CollectionInterval:            1 * time.Hour,
		BatchProcessingSize:           50,
		SecurityViolationThreshold:    0.1,
		TrustScoreThreshold:           0.8,
		VerificationAccuracyThreshold: 0.9,
		ImprovementInterval:           24 * time.Hour,
		MaxImprovementRuns:            5,
		ImprovementTimeout:            30 * time.Minute,
	}

	mockRepo := &MockFeedbackRepository{}
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 10,
	}, logger)
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          10,
		MaxLearningBatchSize:          100,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	system := NewAdvancedSecurityFeedbackSystem(
		config,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		mockRepo,
	)

	ctx := context.Background()

	// First collect feedback
	collectionResult, err := system.CollectSecurityFeedback(ctx)
	if err != nil {
		t.Fatalf("Failed to collect security feedback: %v", err)
	}

	// Then analyze the collected feedback
	analysisResult, err := system.AnalyzeSecurityFeedback(ctx, collectionResult)
	if err != nil {
		t.Fatalf("Failed to analyze security feedback: %v", err)
	}

	// Validate analysis result
	if analysisResult == nil {
		t.Fatal("Analysis result should not be nil")
	}

	if analysisResult.OverallSecurityScore < 0 || analysisResult.OverallSecurityScore > 1 {
		t.Error("Overall security score should be between 0 and 1")
	}

	if analysisResult.DataQualityScore < 0 || analysisResult.DataQualityScore > 1 {
		t.Error("Data quality score should be between 0 and 1")
	}

	if analysisResult.AnalysisTime <= 0 {
		t.Error("Analysis time should be positive")
	}

	// Check that metrics were updated
	metrics := system.GetSecurityMetrics()
	if metrics.AnalysisRunsCompleted == 0 {
		t.Error("Analysis runs completed should be updated")
	}

	t.Logf("Security feedback analysis completed: security score %.2f, quality score %.2f in %v",
		analysisResult.OverallSecurityScore, analysisResult.DataQualityScore, analysisResult.AnalysisTime)
}

func TestAdvancedSecurityFeedbackSystem_ImproveSecurityAlgorithms(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AdvancedSecurityConfig{
		MinFeedbackThreshold:          10,
		MaxFeedbackAge:                24 * time.Hour,
		CollectionInterval:            1 * time.Hour,
		BatchProcessingSize:           50,
		SecurityViolationThreshold:    0.1,
		TrustScoreThreshold:           0.8,
		VerificationAccuracyThreshold: 0.9,
		ImprovementInterval:           24 * time.Hour,
		MaxImprovementRuns:            5,
		ImprovementTimeout:            30 * time.Minute,
	}

	mockRepo := &MockFeedbackRepository{}
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 10,
	}, logger)
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          10,
		MaxLearningBatchSize:          100,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	system := NewAdvancedSecurityFeedbackSystem(
		config,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		mockRepo,
	)

	ctx := context.Background()

	// Create a mock analysis result that triggers improvement
	analysisResult := &SecurityFeedbackAnalysisResult{
		OverallSecurityScore: 0.05, // Below threshold to trigger improvement
		DataQualityScore:     0.9,  // High quality to allow improvement
		ImprovementOpportunities: []*ImprovementOpportunity{
			{
				OpportunityID:            "test_opportunity",
				AlgorithmType:            "website_verification",
				CurrentPerformance:       0.8,
				PotentialImprovement:     0.2,
				ImplementationComplexity: "medium",
				ExpectedImpact:           "high",
				Priority:                 "high",
			},
		},
	}

	// Test algorithm improvement
	improvementResult, err := system.ImproveSecurityAlgorithms(ctx, analysisResult)
	if err != nil {
		t.Fatalf("Failed to improve security algorithms: %v", err)
	}

	// Validate improvement result
	if improvementResult == nil {
		t.Fatal("Improvement result should not be nil")
	}

	if improvementResult.SuccessRate < 0 || improvementResult.SuccessRate > 1 {
		t.Error("Success rate should be between 0 and 1")
	}

	if improvementResult.ImprovementTime <= 0 {
		t.Error("Improvement time should be positive")
	}

	// Check that metrics were updated
	metrics := system.GetSecurityMetrics()
	if metrics.ImprovementRunsCompleted == 0 {
		t.Error("Improvement runs completed should be updated")
	}

	t.Logf("Security algorithm improvement completed: %d algorithms improved, success rate %.2f in %v",
		len(improvementResult.AlgorithmsImproved), improvementResult.SuccessRate, improvementResult.ImprovementTime)
}

func TestAdvancedSecurityFeedbackSystem_GetSystemHealth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AdvancedSecurityConfig{
		MinFeedbackThreshold:          10,
		MaxFeedbackAge:                24 * time.Hour,
		CollectionInterval:            1 * time.Hour,
		BatchProcessingSize:           50,
		SecurityViolationThreshold:    0.1,
		TrustScoreThreshold:           0.8,
		VerificationAccuracyThreshold: 0.9,
		ImprovementInterval:           24 * time.Hour,
		MaxImprovementRuns:            5,
		ImprovementTimeout:            30 * time.Minute,
	}

	mockRepo := &MockFeedbackRepository{}
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 10,
	}, logger)
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          10,
		MaxLearningBatchSize:          100,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	system := NewAdvancedSecurityFeedbackSystem(
		config,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		mockRepo,
	)

	ctx := context.Background()

	// Test system health
	health := system.GetSystemHealth(ctx)

	// Validate health status
	if health == nil {
		t.Fatal("Health status should not be nil")
	}

	if health["status"] != "healthy" {
		t.Error("System should be healthy")
	}

	// Check required health fields
	requiredFields := []string{
		"total_feedback_collected",
		"analysis_runs_completed",
		"improvement_runs_completed",
		"last_collection_time",
		"last_analysis_time",
		"last_improvement_time",
		"system_uptime",
	}

	for _, field := range requiredFields {
		if _, exists := health[field]; !exists {
			t.Errorf("Health status should contain field: %s", field)
		}
	}

	t.Logf("System health check completed: %+v", health)
}

func TestAdvancedSecurityFeedbackSystem_ConcurrentOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AdvancedSecurityConfig{
		MinFeedbackThreshold:          10,
		MaxFeedbackAge:                24 * time.Hour,
		CollectionInterval:            1 * time.Hour,
		BatchProcessingSize:           50,
		SecurityViolationThreshold:    0.1,
		TrustScoreThreshold:           0.8,
		VerificationAccuracyThreshold: 0.9,
		ImprovementInterval:           24 * time.Hour,
		MaxImprovementRuns:            5,
		ImprovementTimeout:            30 * time.Minute,
	}

	mockRepo := &MockFeedbackRepository{}
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 10,
	}, logger)
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          10,
		MaxLearningBatchSize:          100,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	system := NewAdvancedSecurityFeedbackSystem(
		config,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		mockRepo,
	)

	ctx := context.Background()

	// Test concurrent operations
	done := make(chan bool, 3)
	errors := make(chan error, 3)

	// Concurrent collection
	go func() {
		_, err := system.CollectSecurityFeedback(ctx)
		errors <- err
		done <- true
	}()

	// Concurrent health check
	go func() {
		_ = system.GetSystemHealth(ctx)
		errors <- nil
		done <- true
	}()

	// Concurrent metrics retrieval
	go func() {
		_ = system.GetSecurityMetrics()
		errors <- nil
		done <- true
	}()

	// Wait for all operations to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Check for errors
	for i := 0; i < 3; i++ {
		if err := <-errors; err != nil {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}

	t.Log("Concurrent operations completed successfully")
}

func TestAdvancedSecurityFeedbackSystem_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AdvancedSecurityConfig{
		MinFeedbackThreshold:          1,
		MaxFeedbackAge:                1 * time.Hour,
		CollectionInterval:            1 * time.Hour,
		BatchProcessingSize:           10,
		SecurityViolationThreshold:    0.1,
		TrustScoreThreshold:           0.8,
		VerificationAccuracyThreshold: 0.9,
		ImprovementInterval:           24 * time.Hour,
		MaxImprovementRuns:            5,
		ImprovementTimeout:            30 * time.Minute,
	}

	mockRepo := &MockFeedbackRepository{}
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 1,
	}, logger)
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          1,
		MaxLearningBatchSize:          10,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	system := NewAdvancedSecurityFeedbackSystem(
		config,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		mockRepo,
	)

	ctx := context.Background()

	// Test with empty collection result
	emptyCollectionResult := &SecurityFeedbackCollectionResult{
		CollectedFeedback:         []*UserFeedback{},
		SecurityViolations:        []*SecurityViolation{},
		TrustedSourceIssues:       []*TrustedSourceIssue{},
		WebsiteVerificationIssues: []*WebsiteVerificationIssue{},
		TotalProcessed:            0,
	}

	analysisResult, err := system.AnalyzeSecurityFeedback(ctx, emptyCollectionResult)
	if err != nil {
		t.Fatalf("Failed to analyze empty security feedback: %v", err)
	}

	if analysisResult.OverallSecurityScore != 1.0 {
		t.Error("Overall security score should be 1.0 for empty feedback")
	}

	// Test with high-quality analysis result (should not trigger improvement)
	highQualityAnalysisResult := &SecurityFeedbackAnalysisResult{
		OverallSecurityScore:     0.95, // Above threshold
		DataQualityScore:         0.9,  // High quality
		ImprovementOpportunities: []*ImprovementOpportunity{},
	}

	improvementResult, err := system.ImproveSecurityAlgorithms(ctx, highQualityAnalysisResult)
	if err != nil {
		t.Fatalf("Failed to process high-quality analysis result: %v", err)
	}

	if improvementResult.SuccessRate != 1.0 {
		t.Error("Success rate should be 1.0 when no improvement is needed")
	}

	t.Log("Edge cases handled successfully")
}

func TestAdvancedSecurityFeedbackSystem_ConfigurationValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with nil configuration (should use defaults)
	system := NewAdvancedSecurityFeedbackSystem(
		nil, // nil config
		logger,
		nil, // nil security analyzer
		nil, // nil website verification improver
		nil, // nil repository
	)

	if system == nil {
		t.Fatal("System should be created even with nil configuration")
	}

	if system.config == nil {
		t.Fatal("Default configuration should be created")
	}

	// Validate default configuration values
	if system.config.MinFeedbackThreshold != 50 {
		t.Error("Default MinFeedbackThreshold should be 50")
	}

	if system.config.MaxFeedbackAge != 30*24*time.Hour {
		t.Error("Default MaxFeedbackAge should be 30 days")
	}

	if system.config.SecurityViolationThreshold != 0.1 {
		t.Error("Default SecurityViolationThreshold should be 0.1")
	}

	t.Log("Configuration validation completed successfully")
}

// Benchmark tests
func BenchmarkAdvancedSecurityFeedbackSystem_CollectSecurityFeedback(b *testing.B) {
	logger := zap.NewNop()
	config := &AdvancedSecurityConfig{
		MinFeedbackThreshold:          10,
		MaxFeedbackAge:                24 * time.Hour,
		CollectionInterval:            1 * time.Hour,
		BatchProcessingSize:           50,
		SecurityViolationThreshold:    0.1,
		TrustScoreThreshold:           0.8,
		VerificationAccuracyThreshold: 0.9,
		ImprovementInterval:           24 * time.Hour,
		MaxImprovementRuns:            5,
		ImprovementTimeout:            30 * time.Minute,
	}

	mockRepo := &MockFeedbackRepository{}
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 10,
	}, logger)
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          10,
		MaxLearningBatchSize:          100,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	system := NewAdvancedSecurityFeedbackSystem(
		config,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		mockRepo,
	)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := system.CollectSecurityFeedback(ctx)
		if err != nil {
			b.Fatalf("Failed to collect security feedback: %v", err)
		}
	}
}

func BenchmarkAdvancedSecurityFeedbackSystem_AnalyzeSecurityFeedback(b *testing.B) {
	logger := zap.NewNop()
	config := &AdvancedSecurityConfig{
		MinFeedbackThreshold:          10,
		MaxFeedbackAge:                24 * time.Hour,
		CollectionInterval:            1 * time.Hour,
		BatchProcessingSize:           50,
		SecurityViolationThreshold:    0.1,
		TrustScoreThreshold:           0.8,
		VerificationAccuracyThreshold: 0.9,
		ImprovementInterval:           24 * time.Hour,
		MaxImprovementRuns:            5,
		ImprovementTimeout:            30 * time.Minute,
	}

	mockRepo := &MockFeedbackRepository{}
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 10,
	}, logger)
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          10,
		MaxLearningBatchSize:          100,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	system := NewAdvancedSecurityFeedbackSystem(
		config,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		mockRepo,
	)

	ctx := context.Background()

	// Create a consistent collection result for benchmarking
	collectionResult := &SecurityFeedbackCollectionResult{
		CollectedFeedback: []*UserFeedback{
			{
				ID:                   "test_feedback_1",
				UserID:               "user_123",
				BusinessName:         "Test Business",
				FeedbackType:         FeedbackTypeSecurityValidation,
				FeedbackText:         "Security validation passed",
				ConfidenceScore:      0.95,
				Status:               FeedbackStatusProcessed,
				ProcessingTimeMs:     150,
				ClassificationMethod: MethodSecurity,
				CreatedAt:            time.Now(),
			},
		},
		SecurityViolations:        []*SecurityViolation{},
		TrustedSourceIssues:       []*TrustedSourceIssue{},
		WebsiteVerificationIssues: []*WebsiteVerificationIssue{},
		TotalProcessed:            1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := system.AnalyzeSecurityFeedback(ctx, collectionResult)
		if err != nil {
			b.Fatalf("Failed to analyze security feedback: %v", err)
		}
	}
}
