package webanalysis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBetaTestingFramework(t *testing.T) {
	logger := zap.NewNop()
	scraper := &JavaScriptScraper{}
	basicScraper := &BasicScraper{}
	performanceTracker := &SuccessRateTracker{}

	framework := NewBetaTestingFramework(logger, scraper, basicScraper, performanceTracker)

	assert.NotNil(t, framework)
	assert.Equal(t, logger, framework.logger)
	assert.Equal(t, scraper, framework.scraper)
	assert.Equal(t, basicScraper, framework.basicScraper)
	assert.Equal(t, performanceTracker, framework.performanceTracker)
	assert.NotNil(t, framework.feedbackCollector)
	assert.NotNil(t, framework.abTestManager)
	assert.NotNil(t, framework.metrics)
}

func TestCreateBetaTestScenario(t *testing.T) {
	framework := createTestFramework(t)

	scenario := &BetaTestScenario{
		ID:          "test_scenario_1",
		Name:        "Test Scenario 1",
		Description: "A test scenario for enhanced scraping",
		URLs:        []string{"https://example.com", "https://test.com"},
		Parameters: map[string]interface{}{
			"timeout": 30,
			"retries": 3,
		},
		ExpectedOutcome: "Enhanced scraping should outperform basic scraping",
		Priority:        1,
	}

	err := framework.CreateBetaTestScenario(context.Background(), scenario)
	require.NoError(t, err)

	assert.NotZero(t, scenario.CreatedAt)
}

func TestRunABTest(t *testing.T) {
	framework := createTestFramework(t)

	// Mock scraping results
	framework.scraper = &MockJavaScriptScraper{
		shouldSucceed: true,
		result: &ScrapingResult{
			BusinessName: "Test Business",
			Industry:     "Technology",
			Address:      "123 Test St",
			Phone:        "555-1234",
			Email:        "test@example.com",
			Website:      "https://example.com",
		},
	}

	framework.basicScraper = &MockBasicScraper{
		shouldSucceed: true,
		result: &ScrapingResult{
			BusinessName: "Test Business",
			Industry:     "Tech",
			Address:      "123 Test Street",
			Phone:        "555-1234",
			Email:        "",
			Website:      "https://example.com",
		},
	}

	result, err := framework.RunABTest(context.Background(), "https://example.com", "test_1")
	require.NoError(t, err)

	assert.NotNil(t, result)
	assert.Equal(t, "test_1", result.TestID)
	assert.Equal(t, "https://example.com", result.URL)
	assert.Equal(t, "enhanced", result.Method)
	assert.True(t, result.Success)
	assert.Greater(t, result.ResponseTime, time.Duration(0))
	assert.Greater(t, result.Accuracy, 0.0)
	assert.Greater(t, result.DataQuality, 0.0)
}

func TestRunBetaTestScenario(t *testing.T) {
	framework := createTestFramework(t)

	// Mock scraping results
	framework.scraper = &MockJavaScriptScraper{
		shouldSucceed: true,
		result: &ScrapingResult{
			BusinessName: "Test Business",
			Industry:     "Technology",
			Address:      "123 Test St",
			Phone:        "555-1234",
			Email:        "test@example.com",
			Website:      "https://example.com",
		},
	}

	framework.basicScraper = &MockBasicScraper{
		shouldSucceed: true,
		result: &ScrapingResult{
			BusinessName: "Test Business",
			Industry:     "Tech",
			Address:      "123 Test Street",
			Phone:        "555-1234",
			Email:        "",
			Website:      "https://example.com",
		},
	}

	scenario := &BetaTestScenario{
		ID:   "test_scenario_2",
		Name: "Test Scenario 2",
		URLs: []string{"https://example1.com", "https://example2.com"},
	}

	results, err := framework.RunBetaTestScenario(context.Background(), scenario)
	require.NoError(t, err)

	assert.Len(t, results, 2)
	for _, result := range results {
		assert.Equal(t, "test_scenario_2", result.TestID)
		assert.True(t, result.Success)
		assert.Greater(t, result.ResponseTime, time.Duration(0))
	}
}

func TestCollectUserFeedback(t *testing.T) {
	framework := createTestFramework(t)

	feedback := &BetaFeedback{
		UserID:       "user_1",
		TestID:       "test_1",
		URL:          "https://example.com",
		Method:       "enhanced",
		Satisfaction: 4,
		Accuracy:     5,
		Speed:        4,
		Comments:     "Great results with enhanced scraping",
	}

	err := framework.CollectUserFeedback(context.Background(), feedback)
	require.NoError(t, err)

	assert.NotZero(t, feedback.Timestamp)
}

func TestGetPerformanceComparison(t *testing.T) {
	framework := createTestFramework(t)

	// Mock performance tracker
	framework.performanceTracker = &MockSuccessRateTracker{
		basicMetrics: &ScrapingMetrics{
			SuccessRate:         0.8,
			AverageResponseTime: 2 * time.Second,
			DataQuality:         0.7,
		},
		enhancedMetrics: &ScrapingMetrics{
			SuccessRate:         0.95,
			AverageResponseTime: 1.5 * time.Second,
			DataQuality:         0.9,
		},
	}

	comparison, err := framework.GetPerformanceComparison(context.Background(), 24*time.Hour)
	require.NoError(t, err)

	assert.NotNil(t, comparison)
	assert.Equal(t, 24*time.Hour, comparison.TimeRange)
	assert.NotZero(t, comparison.Generated)
	assert.NotNil(t, comparison.BasicMetrics)
	assert.NotNil(t, comparison.EnhancedMetrics)
	assert.Greater(t, comparison.SuccessRateImprovement, 0.0)
	assert.Greater(t, comparison.ResponseTimeImprovement, time.Duration(0))
	assert.Greater(t, comparison.DataQualityImprovement, 0.0)
}

func TestGetBetaMetrics(t *testing.T) {
	framework := createTestFramework(t)

	// Add some test data
	framework.metrics.RecordABTestResult(&ABTestResult{
		TestID:       "test_1",
		URL:          "https://example.com",
		Method:       "enhanced",
		Success:      true,
		ResponseTime: 1 * time.Second,
		Accuracy:     0.9,
		DataQuality:  0.85,
		Timestamp:    time.Now(),
	})

	framework.metrics.RecordUserFeedback(&BetaFeedback{
		UserID:       "user_1",
		TestID:       "test_1",
		Method:       "enhanced",
		Satisfaction: 4,
		Accuracy:     5,
		Speed:        4,
		Timestamp:    time.Now(),
	})

	metrics, err := framework.GetBetaMetrics(context.Background())
	require.NoError(t, err)

	assert.NotNil(t, metrics)
	assert.Greater(t, metrics.TotalTests, 0)
	assert.Greater(t, metrics.EnhancedSuccessRate, 0.0)
	assert.Greater(t, metrics.AverageSatisfaction, 0.0)
	assert.Greater(t, metrics.AverageAccuracy, 0.0)
	assert.Greater(t, metrics.AverageSpeed, 0.0)
}

func TestCalculateAccuracy(t *testing.T) {
	framework := createTestFramework(t)

	// Test with complete data
	result := &ScrapingResult{
		BusinessName: "Test Business",
		Industry:     "Technology",
		Address:      "123 Test St",
		Phone:        "555-1234",
		Email:        "test@example.com",
		Website:      "https://example.com",
	}

	accuracy := framework.calculateAccuracy(result)
	assert.Equal(t, 1.0, accuracy)

	// Test with partial data
	result = &ScrapingResult{
		BusinessName: "Test Business",
		Industry:     "Technology",
		Address:      "",
		Phone:        "555-1234",
		Email:        "",
		Website:      "",
	}

	accuracy = framework.calculateAccuracy(result)
	assert.Equal(t, 0.5, accuracy) // 3 out of 6 fields populated

	// Test with nil result
	accuracy = framework.calculateAccuracy(nil)
	assert.Equal(t, 0.0, accuracy)
}

func TestCalculateDataQuality(t *testing.T) {
	framework := createTestFramework(t)

	// Test with high-quality data
	result := &ScrapingResult{
		BusinessName: "Test Business Inc.",
		Industry:     "Technology",
		Address:      "123 Main Street, City, State 12345",
		Phone:        "555-123-4567",
		Email:        "contact@testbusiness.com",
		Website:      "https://testbusiness.com",
	}

	quality := framework.calculateDataQuality(result)
	assert.Greater(t, quality, 0.8)

	// Test with low-quality data
	result = &ScrapingResult{
		BusinessName: "A",
		Industry:     "",
		Address:      "Short",
		Phone:        "123",
		Email:        "invalid",
		Website:      "",
	}

	quality = framework.calculateDataQuality(result)
	assert.Less(t, quality, 0.5)

	// Test with nil result
	quality = framework.calculateDataQuality(nil)
	assert.Equal(t, 0.0, quality)
}

func TestAssessFieldQuality(t *testing.T) {
	framework := createTestFramework(t)

	// Test good quality field
	quality := framework.assessFieldQuality("Test Business Inc.")
	assert.Greater(t, quality, 0.8)

	// Test very short field
	quality = framework.assessFieldQuality("A")
	assert.Less(t, quality, 0.6)

	// Test very long field (likely garbage)
	longField := string(make([]byte, 600))
	quality = framework.assessFieldQuality(longField)
	assert.Less(t, quality, 0.9)

	// Test empty field
	quality = framework.assessFieldQuality("")
	assert.Equal(t, 0.0, quality)
}

func TestAssessIndustryQuality(t *testing.T) {
	framework := createTestFramework(t)

	// Test valid industry code
	quality := framework.assessIndustryQuality("541511")
	assert.Greater(t, quality, 1.0)

	// Test meaningful industry name
	quality = framework.assessIndustryQuality("Technology")
	assert.Greater(t, quality, 1.0)

	// Test poor quality
	quality = framework.assessIndustryQuality("")
	assert.Equal(t, 1.0, quality)
}

func TestAssessAddressQuality(t *testing.T) {
	framework := createTestFramework(t)

	// Test good address
	quality := framework.assessAddressQuality("123 Main Street, City, State 12345")
	assert.Greater(t, quality, 1.0)

	// Test short address
	quality = framework.assessAddressQuality("Short")
	assert.Less(t, quality, 0.8)

	// Test empty address
	quality = framework.assessAddressQuality("")
	assert.Equal(t, 1.0, quality)
}

func TestAssessPhoneQuality(t *testing.T) {
	framework := createTestFramework(t)

	// Test valid phone
	quality := framework.assessPhoneQuality("555-123-4567")
	assert.Greater(t, quality, 1.0)

	// Test short phone
	quality = framework.assessPhoneQuality("123")
	assert.Equal(t, 1.0, quality)

	// Test empty phone
	quality = framework.assessPhoneQuality("")
	assert.Equal(t, 1.0, quality)
}

func TestAssessEmailQuality(t *testing.T) {
	framework := createTestFramework(t)

	// Test valid email
	quality := framework.assessEmailQuality("test@example.com")
	assert.Greater(t, quality, 1.0)

	// Test invalid email
	quality = framework.assessEmailQuality("invalid")
	assert.Equal(t, 1.0, quality)

	// Test empty email
	quality = framework.assessEmailQuality("")
	assert.Equal(t, 1.0, quality)
}

// Helper functions

func createTestFramework(t *testing.T) *BetaTestingFramework {
	logger := zap.NewNop()
	scraper := &JavaScriptScraper{}
	basicScraper := &BasicScraper{}
	performanceTracker := &SuccessRateTracker{}

	return NewBetaTestingFramework(logger, scraper, basicScraper, performanceTracker)
}

// Mock implementations for testing

type MockJavaScriptScraper struct {
	shouldSucceed bool
	result        *ScrapingResult
	err           error
}

func (m *MockJavaScriptScraper) Scrape(ctx context.Context, url string) (*ScrapingResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.shouldSucceed {
		return nil, assert.AnError
	}
	return m.result, nil
}

type MockBasicScraper struct {
	shouldSucceed bool
	result        *ScrapingResult
	err           error
}

func (m *MockBasicScraper) Scrape(ctx context.Context, url string) (*ScrapingResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.shouldSucceed {
		return nil, assert.AnError
	}
	return m.result, nil
}

type MockSuccessRateTracker struct {
	basicMetrics    *ScrapingMetrics
	enhancedMetrics *ScrapingMetrics
	err             error
}

func (m *MockSuccessRateTracker) GetMetrics(ctx context.Context, method string, timeRange time.Duration) (*ScrapingMetrics, error) {
	if m.err != nil {
		return nil, m.err
	}
	if method == "basic" {
		return m.basicMetrics, nil
	}
	return m.enhancedMetrics, nil
}
