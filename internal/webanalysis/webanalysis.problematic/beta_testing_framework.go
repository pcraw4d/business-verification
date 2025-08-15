package webanalysis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BetaTestingFramework manages beta testing for enhanced scraping features
type BetaTestingFramework struct {
	logger             *zap.Logger
	scraper            *JavaScriptScraper
	basicScraper       *BasicScraper
	performanceTracker *SuccessRateTracker
	feedbackCollector  *BetaFeedbackCollector
	abTestManager      *ABTestManager
	metrics            *BetaMetricsCollector
	mu                 sync.RWMutex
}

// BetaTestScenario defines a test scenario for beta testing
type BetaTestScenario struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	URLs            []string               `json:"urls"`
	Parameters      map[string]interface{} `json:"parameters"`
	ExpectedOutcome string                 `json:"expected_outcome"`
	Priority        int                    `json:"priority"`
	CreatedAt       time.Time              `json:"created_at"`
}

// ABTestResult represents the result of an A/B test
type ABTestResult struct {
	TestID       string        `json:"test_id"`
	URL          string        `json:"url"`
	Method       string        `json:"method"` // "basic" or "enhanced"
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	Accuracy     float64       `json:"accuracy"`
	DataQuality  float64       `json:"data_quality"`
	Timestamp    time.Time     `json:"timestamp"`
}

// BetaFeedback represents user feedback for scraping features
type BetaFeedback struct {
	UserID       string    `json:"user_id"`
	TestID       string    `json:"test_id"`
	URL          string    `json:"url"`
	Method       string    `json:"method"`
	Satisfaction int       `json:"satisfaction"` // 1-5 scale
	Accuracy     int       `json:"accuracy"`     // 1-5 scale
	Speed        int       `json:"speed"`        // 1-5 scale
	Comments     string    `json:"comments"`
	Timestamp    time.Time `json:"timestamp"`
}

// NewBetaTestingFramework creates a new beta testing framework
func NewBetaTestingFramework(
	logger *zap.Logger,
	scraper *JavaScriptScraper,
	basicScraper *BasicScraper,
	performanceTracker *SuccessRateTracker,
) *BetaTestingFramework {
	return &BetaTestingFramework{
		logger:             logger,
		scraper:            scraper,
		basicScraper:       basicScraper,
		performanceTracker: performanceTracker,
		feedbackCollector:  NewBetaFeedbackCollector(logger),
		abTestManager:      NewABTestManager(logger),
		metrics:            NewBetaMetricsCollector(logger),
	}
}

// CreateBetaTestScenario creates a new beta test scenario
func (btf *BetaTestingFramework) CreateBetaTestScenario(ctx context.Context, scenario *BetaTestScenario) error {
	btf.mu.Lock()
	defer btf.mu.Unlock()

	scenario.CreatedAt = time.Now()

	btf.logger.Info("Created beta test scenario",
		zap.String("scenario_id", scenario.ID),
		zap.String("name", scenario.Name),
		zap.Int("url_count", len(scenario.URLs)),
		zap.Int("priority", scenario.Priority),
	)

	return nil
}

// RunABTest performs A/B testing between basic and enhanced scraping
func (btf *BetaTestingFramework) RunABTest(ctx context.Context, url string, testID string) (*ABTestResult, error) {
	btf.logger.Info("Starting A/B test",
		zap.String("test_id", testID),
		zap.String("url", url),
	)

	// Test basic scraping
	basicStart := time.Now()
	basicResult, basicErr := btf.basicScraper.Scrape(ctx, url)
	basicDuration := time.Since(basicStart)

	// Test enhanced scraping
	enhancedStart := time.Now()
	enhancedResult, enhancedErr := btf.scraper.Scrape(ctx, url)
	enhancedDuration := time.Since(enhancedStart)

	// Compare results
	basicSuccess := basicErr == nil && basicResult != nil
	enhancedSuccess := enhancedErr == nil && enhancedResult != nil

	// Calculate accuracy and data quality
	basicAccuracy := btf.calculateAccuracy(basicResult)
	enhancedAccuracy := btf.calculateAccuracy(enhancedResult)
	basicQuality := btf.calculateDataQuality(basicResult)
	enhancedQuality := btf.calculateDataQuality(enhancedResult)

	// Create results
	basicABResult := &ABTestResult{
		TestID:       testID,
		URL:          url,
		Method:       "basic",
		Success:      basicSuccess,
		ResponseTime: basicDuration,
		Accuracy:     basicAccuracy,
		DataQuality:  basicQuality,
		Timestamp:    time.Now(),
	}

	enhancedABResult := &ABTestResult{
		TestID:       testID,
		URL:          url,
		Method:       "enhanced",
		Success:      enhancedSuccess,
		ResponseTime: enhancedDuration,
		Accuracy:     enhancedAccuracy,
		DataQuality:  enhancedQuality,
		Timestamp:    time.Now(),
	}

	// Store results
	btf.abTestManager.StoreResult(basicABResult)
	btf.abTestManager.StoreResult(enhancedABResult)

	// Track metrics
	btf.metrics.RecordABTestResult(basicABResult)
	btf.metrics.RecordABTestResult(enhancedABResult)

	btf.logger.Info("A/B test completed",
		zap.String("test_id", testID),
		zap.String("url", url),
		zap.Bool("basic_success", basicSuccess),
		zap.Bool("enhanced_success", enhancedSuccess),
		zap.Duration("basic_time", basicDuration),
		zap.Duration("enhanced_time", enhancedDuration),
		zap.Float64("basic_accuracy", basicAccuracy),
		zap.Float64("enhanced_accuracy", enhancedAccuracy),
	)

	return enhancedABResult, nil
}

// RunBetaTestScenario executes a complete beta test scenario
func (btf *BetaTestingFramework) RunBetaTestScenario(ctx context.Context, scenario *BetaTestScenario) ([]*ABTestResult, error) {
	btf.logger.Info("Running beta test scenario",
		zap.String("scenario_id", scenario.ID),
		zap.String("name", scenario.Name),
		zap.Int("url_count", len(scenario.URLs)),
	)

	var results []*ABTestResult
	var wg sync.WaitGroup
	resultChan := make(chan *ABTestResult, len(scenario.URLs))

	// Run A/B tests for each URL in parallel
	for _, url := range scenario.URLs {
		wg.Add(1)
		go func(testURL string) {
			defer wg.Done()

			result, err := btf.RunABTest(ctx, testURL, scenario.ID)
			if err != nil {
				btf.logger.Error("Failed to run A/B test",
					zap.String("url", testURL),
					zap.String("scenario_id", scenario.ID),
					zap.Error(err),
				)
				return
			}

			resultChan <- result
		}(url)
	}

	// Wait for all tests to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}

	btf.logger.Info("Beta test scenario completed",
		zap.String("scenario_id", scenario.ID),
		zap.Int("total_results", len(results)),
	)

	return results, nil
}

// CollectUserFeedback collects feedback from beta users
func (btf *BetaTestingFramework) CollectUserFeedback(ctx context.Context, feedback *BetaFeedback) error {
	feedback.Timestamp = time.Now()

	err := btf.feedbackCollector.StoreFeedback(feedback)
	if err != nil {
		return fmt.Errorf("failed to store feedback: %w", err)
	}

	btf.metrics.RecordUserFeedback(feedback)

	btf.logger.Info("User feedback collected",
		zap.String("user_id", feedback.UserID),
		zap.String("test_id", feedback.TestID),
		zap.String("method", feedback.Method),
		zap.Int("satisfaction", feedback.Satisfaction),
		zap.Int("accuracy", feedback.Accuracy),
		zap.Int("speed", feedback.Speed),
	)

	return nil
}

// GetPerformanceComparison returns performance comparison between basic and enhanced scraping
func (btf *BetaTestingFramework) GetPerformanceComparison(ctx context.Context, timeRange time.Duration) (*PerformanceComparison, error) {
	btf.mu.RLock()
	defer btf.mu.RUnlock()

	comparison := &PerformanceComparison{
		TimeRange: timeRange,
		Generated: time.Now(),
	}

	// Get basic scraping metrics
	basicMetrics, err := btf.performanceTracker.GetMetrics(ctx, "basic", timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic metrics: %w", err)
	}

	// Get enhanced scraping metrics
	enhancedMetrics, err := btf.performanceTracker.GetMetrics(ctx, "enhanced", timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get enhanced metrics: %w", err)
	}

	comparison.BasicMetrics = basicMetrics
	comparison.EnhancedMetrics = enhancedMetrics

	// Calculate improvements
	comparison.SuccessRateImprovement = enhancedMetrics.SuccessRate - basicMetrics.SuccessRate
	comparison.ResponseTimeImprovement = basicMetrics.AverageResponseTime - enhancedMetrics.AverageResponseTime
	comparison.DataQualityImprovement = enhancedMetrics.DataQuality - basicMetrics.DataQuality

	return comparison, nil
}

// GetBetaMetrics returns comprehensive beta testing metrics
func (btf *BetaTestingFramework) GetBetaMetrics(ctx context.Context) (*BetaMetrics, error) {
	return btf.metrics.GetMetrics(ctx)
}

// calculateAccuracy calculates the accuracy of scraping results
func (btf *BetaTestingFramework) calculateAccuracy(result *ScrapingResult) float64 {
	if result == nil {
		return 0.0
	}

	// Simple accuracy calculation based on data completeness
	totalFields := 0
	populatedFields := 0

	if result.BusinessName != "" {
		populatedFields++
	}
	totalFields++

	if result.Industry != "" {
		populatedFields++
	}
	totalFields++

	if result.Address != "" {
		populatedFields++
	}
	totalFields++

	if result.Phone != "" {
		populatedFields++
	}
	totalFields++

	if result.Email != "" {
		populatedFields++
	}
	totalFields++

	if result.Website != "" {
		populatedFields++
	}
	totalFields++

	if totalFields == 0 {
		return 0.0
	}

	return float64(populatedFields) / float64(totalFields)
}

// calculateDataQuality calculates the quality of scraped data
func (btf *BetaTestingFramework) calculateDataQuality(result *ScrapingResult) float64 {
	if result == nil {
		return 0.0
	}

	quality := 0.0
	count := 0

	// Check business name quality
	if result.BusinessName != "" {
		quality += btf.assessFieldQuality(result.BusinessName)
		count++
	}

	// Check industry classification quality
	if result.Industry != "" {
		quality += btf.assessIndustryQuality(result.Industry)
		count++
	}

	// Check address quality
	if result.Address != "" {
		quality += btf.assessAddressQuality(result.Address)
		count++
	}

	// Check contact information quality
	if result.Phone != "" {
		quality += btf.assessPhoneQuality(result.Phone)
		count++
	}

	if result.Email != "" {
		quality += btf.assessEmailQuality(result.Email)
		count++
	}

	if count == 0 {
		return 0.0
	}

	return quality / float64(count)
}

// assessFieldQuality assesses the quality of a text field
func (btf *BetaTestingFramework) assessFieldQuality(field string) float64 {
	if field == "" {
		return 0.0
	}

	quality := 1.0

	// Penalize very short fields
	if len(field) < 2 {
		quality -= 0.5
	}

	// Penalize fields with only special characters
	specialCharCount := 0
	for _, char := range field {
		if char < 32 || char > 126 {
			specialCharCount++
		}
	}

	if float64(specialCharCount)/float64(len(field)) > 0.5 {
		quality -= 0.3
	}

	// Penalize fields that are too long (likely garbage)
	if len(field) > 500 {
		quality -= 0.2
	}

	return quality
}

// assessIndustryQuality assesses the quality of industry classification
func (btf *BetaTestingFramework) assessIndustryQuality(industry string) float64 {
	// Industry-specific quality assessment
	quality := 1.0

	// Check if it's a recognized industry code
	if btf.isValidIndustryCode(industry) {
		quality += 0.2
	}

	// Check if it's a meaningful industry name
	if btf.isMeaningfulIndustryName(industry) {
		quality += 0.1
	}

	return quality
}

// assessAddressQuality assesses the quality of address data
func (btf *BetaTestingFramework) assessAddressQuality(address string) float64 {
	quality := 1.0

	// Check for common address patterns
	if btf.containsAddressPatterns(address) {
		quality += 0.2
	}

	// Check for reasonable length
	if len(address) < 10 {
		quality -= 0.3
	}

	return quality
}

// assessPhoneQuality assesses the quality of phone number
func (btf *BetaTestingFramework) assessPhoneQuality(phone string) float64 {
	quality := 1.0

	// Check for valid phone number format
	if btf.isValidPhoneFormat(phone) {
		quality += 0.3
	}

	return quality
}

// assessEmailQuality assesses the quality of email address
func (btf *BetaTestingFramework) assessEmailQuality(email string) float64 {
	quality := 1.0

	// Check for valid email format
	if btf.isValidEmailFormat(email) {
		quality += 0.3
	}

	return quality
}

// Helper methods for quality assessment
func (btf *BetaTestingFramework) isValidIndustryCode(code string) bool {
	// Implementation would check against known industry codes
	return len(code) > 0 && len(code) < 10
}

func (btf *BetaTestingFramework) isMeaningfulIndustryName(name string) bool {
	// Implementation would check against known industry names
	return len(name) > 2 && len(name) < 100
}

func (btf *BetaTestingFramework) containsAddressPatterns(address string) bool {
	// Implementation would check for address patterns
	return len(address) > 0
}

func (btf *BetaTestingFramework) isValidPhoneFormat(phone string) bool {
	// Implementation would validate phone number format
	return len(phone) >= 10
}

func (btf *BetaTestingFramework) isValidEmailFormat(email string) bool {
	// Implementation would validate email format
	return len(email) > 0 && len(email) < 100
}

// PerformanceComparison represents performance comparison between scraping methods
type PerformanceComparison struct {
	TimeRange               time.Duration    `json:"time_range"`
	Generated               time.Time        `json:"generated"`
	BasicMetrics            *ScrapingMetrics `json:"basic_metrics"`
	EnhancedMetrics         *ScrapingMetrics `json:"enhanced_metrics"`
	SuccessRateImprovement  float64          `json:"success_rate_improvement"`
	ResponseTimeImprovement time.Duration    `json:"response_time_improvement"`
	DataQualityImprovement  float64          `json:"data_quality_improvement"`
}

// BetaMetrics represents comprehensive beta testing metrics
type BetaMetrics struct {
	TotalTests          int       `json:"total_tests"`
	EnhancedSuccessRate float64   `json:"enhanced_success_rate"`
	BasicSuccessRate    float64   `json:"basic_success_rate"`
	AverageSatisfaction float64   `json:"average_satisfaction"`
	AverageAccuracy     float64   `json:"average_accuracy"`
	AverageSpeed        float64   `json:"average_speed"`
	UserFeedbackCount   int       `json:"user_feedback_count"`
	Generated           time.Time `json:"generated"`
}
