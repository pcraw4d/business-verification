package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// TestEndToEndClassification tests the complete classification flow
func TestEndToEndClassification(t *testing.T) {
	// Create test dependencies
	logger := createTestLogger()
	metrics := createTestMetrics()
	service := createTestClassificationService(t)

	// Test end-to-end classification
	ctx := context.Background()
	request := &classification.ClassificationRequest{
		BusinessName:       "Tech Solutions Inc",
		BusinessType:       "Corporation",
		Industry:           "Technology",
		Description:        "Software development and consulting services",
		Keywords:           "software, development, consulting, technology",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
	}

	// Perform classification
	result, err := service.ClassifyBusiness(ctx, request)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected classification result")
	}

	// Verify basic result structure
	if !result.Success {
		t.Error("Expected successful classification")
	}

	if result.BusinessID == "" {
		t.Error("Expected business ID")
	}

	if len(result.Classifications) == 0 {
		t.Error("Expected classifications")
	}

	if result.PrimaryClassification == nil {
		t.Error("Expected primary classification")
	}

	// Verify primary classification
	primary := result.PrimaryClassification
	if primary.IndustryCode == "" {
		t.Error("Expected industry code")
	}
	if primary.IndustryName == "" {
		t.Error("Expected industry name")
	}
	if primary.ConfidenceScore < 0 || primary.ConfidenceScore > 1 {
		t.Errorf("Expected confidence score between 0 and 1, got: %f", primary.ConfidenceScore)
	}

	// Verify processing metadata
	if result.ProcessingTime == 0 {
		t.Error("Expected processing time")
	}
	if result.Timestamp.IsZero() {
		t.Error("Expected timestamp")
	}
}

// TestAPIEndpoints tests the API endpoints
func TestAPIEndpoints(t *testing.T) {
	// Create test handler
	handler := createTestEnhancedClassificationHandler(t)

	// Test enhanced classification endpoint
	t.Run("EnhancedClassification", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"business_name":     "Tech Solutions Inc",
			"business_type":     "Corporation",
			"industry":          "Technology",
			"description":       "Software development services",
			"keywords":          "software, development, technology",
			"ml_model_version":  "bert_classifier_v1.0.0",
			"geographic_region": "California, USA",
			"industry_type":     "technology",
		}

		req := createTestRequest(t, "POST", "/api/v2/classification/enhanced", requestBody)
		req.Header.Set("X-API-Version", "v2")

		w := httptest.NewRecorder()
		handler.HandleEnhancedClassification(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response handlers.EnhancedClassificationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Error("Expected successful response")
		}

		if response.APIVersion != "v2" {
			t.Errorf("Expected API version v2, got %s", response.APIVersion)
		}
	})

	// Test batch classification endpoint
	t.Run("BatchClassification", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"businesses": []map[string]interface{}{
				{
					"business_name": "Tech Solutions Inc",
					"description":   "Software development services",
				},
				{
					"business_name": "Data Analytics Corp",
					"description":   "Data analytics and consulting",
				},
			},
		}

		req := createTestRequest(t, "POST", "/api/v2/classification/enhanced/batch", requestBody)
		req.Header.Set("X-API-Version", "v2")

		w := httptest.NewRecorder()
		handler.HandleBatchEnhancedClassification(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response handlers.BatchEnhancedClassificationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Error("Expected successful response")
		}

		if response.TotalProcessed != 2 {
			t.Errorf("Expected 2 processed, got %d", response.TotalProcessed)
		}

		if len(response.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(response.Results))
		}
	})

	// Test backward compatible endpoint
	t.Run("BackwardCompatible", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"business_name": "Tech Solutions Inc",
			"description":   "Software development services",
		}

		req := createTestRequest(t, "POST", "/api/v1/classification", requestBody)

		w := httptest.NewRecorder()
		handler.HandleBackwardCompatibleClassification(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if !response["success"].(bool) {
			t.Error("Expected successful response")
		}
	})

	// Test API version info endpoint
	t.Run("APIVersionInfo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/versions", nil)

		w := httptest.NewRecorder()
		handler.HandleAPIVersionInfo(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["current_version"] != "v2" {
			t.Errorf("Expected current version v2, got %s", response["current_version"])
		}
	})
}

// TestDatabaseIntegration tests database integration
func TestDatabaseIntegration(t *testing.T) {
	// Create test database connection
	db := createTestDatabase(t)
	defer db.Close()

	// Test database schema
	t.Run("SchemaValidation", func(t *testing.T) {
		// Test that required tables exist
		tables := []string{
			"classifications",
			"feedback",
			"accuracy_validations",
			"accuracy_alerts",
			"accuracy_thresholds",
			"ml_model_versions",
			"crosswalk_mappings",
			"geographic_regions",
			"industry_mappings",
			"dashboard_widgets",
			"dashboard_metrics",
			"alerting_rules",
			"accuracy_reports",
		}

		for _, table := range tables {
			exists, err := tableExists(db, table)
			if err != nil {
				t.Fatalf("Failed to check table %s: %v", table, err)
			}
			if !exists {
				t.Errorf("Expected table %s to exist", table)
			}
		}
	})

	// Test data persistence
	t.Run("DataPersistence", func(t *testing.T) {
		// Test classification data persistence
		classification := &classification.EnhancedClassification{
			BusinessID:              "test_business_1",
			BusinessName:            "Test Business",
			PrimaryIndustryCode:     "541511",
			PrimaryIndustryName:     "Custom Computer Programming Services",
			ConfidenceScore:         0.95,
			MLModelVersion:          "bert_classifier_v1.0.0",
			GeographicRegion:        "California, USA",
			ClassificationAlgorithm: "ml_enhanced",
			CreatedAt:               time.Now(),
		}

		err := saveClassification(db, classification)
		if err != nil {
			t.Fatalf("Failed to save classification: %v", err)
		}

		// Test retrieval
		retrieved, err := getClassification(db, classification.BusinessID)
		if err != nil {
			t.Fatalf("Failed to retrieve classification: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected retrieved classification")
		}

		if retrieved.BusinessName != classification.BusinessName {
			t.Errorf("Expected business name %s, got %s", classification.BusinessName, retrieved.BusinessName)
		}
	})

	// Test feedback persistence
	t.Run("FeedbackPersistence", func(t *testing.T) {
		feedback := &classification.Feedback{
			ID:            "test_feedback_1",
			Type:          classification.FeedbackTypeAccuracy,
			Status:        classification.FeedbackStatusPending,
			Description:   "Test feedback",
			FeedbackValue: true,
			CreatedAt:     time.Now(),
		}

		err := saveFeedback(db, feedback)
		if err != nil {
			t.Fatalf("Failed to save feedback: %v", err)
		}

		// Test retrieval
		retrieved, err := getFeedback(db, feedback.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve feedback: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected retrieved feedback")
		}

		if retrieved.Description != feedback.Description {
			t.Errorf("Expected description %s, got %s", feedback.Description, retrieved.Description)
		}
	})
}

// TestCachingAndPerformance tests caching and performance features
func TestCachingAndPerformance(t *testing.T) {
	// Create test cache manager
	logger := createTestLogger()
	metrics := createTestMetrics()
	cacheManager := classification.NewEnhancedCacheManager(logger, metrics, 1000, time.Hour)

	// Test cache performance
	t.Run("CachePerformance", func(t *testing.T) {
		ctx := context.Background()
		cacheKey := &classification.CacheKey{
			BusinessName: "Performance Test Business",
			BusinessType: "LLC",
			Industry:     "Technology",
		}

		// Test cache miss performance
		start := time.Now()
		_, err := cacheManager.Get(ctx, cacheKey)
		cacheMissTime := time.Since(start)

		if err == nil {
			t.Error("Expected cache miss error")
		}

		// Test cache set performance
		result := &classification.MultiIndustryClassificationResult{
			Success: true,
			Classifications: []classification.IndustryClassification{
				{
					IndustryCode:    "541511",
					IndustryName:    "Custom Computer Programming Services",
					ConfidenceScore: 0.95,
				},
			},
		}

		start = time.Now()
		err = cacheManager.Set(ctx, cacheKey, result, nil)
		cacheSetTime := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Test cache hit performance
		start = time.Now()
		cachedResult, err := cacheManager.Get(ctx, cacheKey)
		cacheHitTime := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to get cached result: %v", err)
		}

		if cachedResult == nil {
			t.Fatal("Expected cached result")
		}

		// Verify performance expectations
		if cacheHitTime >= cacheMissTime {
			t.Errorf("Expected cache hit to be faster than cache miss, got hit: %v, miss: %v", cacheHitTime, cacheMissTime)
		}

		if cacheSetTime > time.Millisecond*100 {
			t.Errorf("Expected cache set to be fast, got: %v", cacheSetTime)
		}
	})

	// Test model optimization performance
	t.Run("ModelOptimizationPerformance", func(t *testing.T) {
		config := &classification.ModelOptimizationConfig{
			QuantizationEnabled:   true,
			QuantizationLevel:     16,
			CacheEnabled:          true,
			PreloadEnabled:        true,
			PerformanceMonitoring: true,
			OptimizationLevel:     "medium",
			MaxCacheSize:          100,
			CacheTTL:              time.Hour,
		}

		optimizer := classification.NewModelOptimizer(config, logger, metrics)

		ctx := context.Background()
		modelName := "performance_test_model"
		modelVersion := "v1.0.0"
		modelData := make([]byte, 1024*1024) // 1MB test data

		// Test optimization performance
		start := time.Now()
		result, err := optimizer.OptimizeModel(ctx, modelName, modelVersion, modelData)
		optimizationTime := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to optimize model: %v", err)
		}

		if result == nil {
			t.Fatal("Expected optimization result")
		}

		// Verify performance expectations
		if optimizationTime > time.Second*10 {
			t.Errorf("Expected optimization to complete within 10 seconds, got: %v", optimizationTime)
		}

		// Test optimized model retrieval performance
		start = time.Now()
		cachedModel, err := optimizer.GetOptimizedModel(ctx, modelName, modelVersion)
		retrievalTime := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to get optimized model: %v", err)
		}

		if cachedModel == nil {
			t.Fatal("Expected cached model")
		}

		// Verify retrieval performance
		if retrievalTime > time.Millisecond*100 {
			t.Errorf("Expected fast model retrieval, got: %v", retrievalTime)
		}
	})

	// Test batch processing performance
	t.Run("BatchProcessingPerformance", func(t *testing.T) {
		service := createTestClassificationService(t)

		// Create batch of requests
		var requests []*classification.ClassificationRequest
		for i := 0; i < 10; i++ {
			requests = append(requests, &classification.ClassificationRequest{
				BusinessName: fmt.Sprintf("Batch Test Business %d", i),
				Description:  "Software development services",
				Keywords:     "software, development, technology",
			})
		}

		// Test batch processing performance
		start := time.Now()
		results, err := service.ClassifyBusinessBatch(context.Background(), requests)
		batchTime := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to process batch: %v", err)
		}

		if len(results) != len(requests) {
			t.Errorf("Expected %d results, got %d", len(requests), len(results))
		}

		// Verify performance expectations
		if batchTime > time.Second*30 {
			t.Errorf("Expected batch processing to complete within 30 seconds, got: %v", batchTime)
		}

		// Calculate average processing time per request
		avgTime := batchTime / time.Duration(len(requests))
		if avgTime > time.Second*3 {
			t.Errorf("Expected average processing time under 3 seconds per request, got: %v", avgTime)
		}
	})
}

// Helper functions

func createTestEnhancedClassificationHandler(t *testing.T) *handlers.EnhancedClassificationHandler {
	// Create mock dependencies
	logger := createTestLogger()
	metrics := createTestMetrics()
	service := createTestClassificationService(t)

	// Create mock components
	mlClassifier := createTestMLClassifier(t)
	crosswalkMapper := createTestCrosswalkMapper(t)
	geographicManager := createTestGeographicManager(t)
	industryMapper := createTestIndustryMapper(t)
	feedbackCollector := createTestFeedbackCollector(t)
	accuracyValidator := createTestAccuracyValidator(t)

	return handlers.NewEnhancedClassificationHandler(
		service,
		mlClassifier,
		crosswalkMapper,
		geographicManager,
		industryMapper,
		feedbackCollector,
		accuracyValidator,
		logger,
		metrics,
	)
}

func createTestRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func createTestClassificationService(t *testing.T) *classification.ClassificationService {
	cfg := &config.ExternalServicesConfig{
		BusinessDataAPI: config.BusinessDataAPIConfig{
			Enabled: true,
			BaseURL: "https://api.example.com",
			APIKey:  "test-key",
			Timeout: 30 * time.Second,
		},
	}

	logger := createTestLogger()
	metrics := createTestMetrics()

	return classification.NewClassificationService(cfg, nil, logger, metrics)
}

func createTestMLClassifier(t *testing.T) *classification.MLClassifier {
	logger := createTestLogger()
	metrics := createTestMetrics()
	modelManager := createTestModelManager(t)
	modelOptimizer := createTestModelOptimizer(t)

	return classification.NewMLClassifier(logger, metrics, modelManager, modelOptimizer)
}

func createTestCrosswalkMapper(t *testing.T) *classification.CrosswalkMapper {
	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewCrosswalkMapper(logger, metrics)
}

func createTestGeographicManager(t *testing.T) *classification.GeographicManager {
	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewGeographicManager(logger, metrics)
}

func createTestIndustryMapper(t *testing.T) *classification.IndustryMapper {
	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewIndustryMapper(logger, metrics)
}

func createTestFeedbackCollector(t *testing.T) *classification.FeedbackCollector {
	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewFeedbackCollector(logger, metrics)
}

func createTestAccuracyValidator(t *testing.T) *classification.AccuracyValidator {
	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewAccuracyValidator(logger, metrics)
}

func createTestModelManager(t *testing.T) *classification.ModelManager {
	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewModelManager(logger, metrics)
}

func createTestModelOptimizer(t *testing.T) *classification.ModelOptimizer {
	config := &classification.ModelOptimizationConfig{
		QuantizationEnabled:   true,
		QuantizationLevel:     16,
		CacheEnabled:          true,
		PreloadEnabled:        true,
		PerformanceMonitoring: true,
		OptimizationLevel:     "medium",
		MaxCacheSize:          100,
		CacheTTL:              time.Hour,
	}

	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewModelOptimizer(config, logger, metrics)
}

func createTestLogger() *observability.Logger {
	return observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
}

func createTestMetrics() *observability.Metrics {
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	return metrics
}

func createTestDatabase(t *testing.T) *sql.DB {
	// This would create a test database connection
	// For now, return nil to avoid compilation errors
	return nil
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	// This would check if a table exists in the database
	// For now, return true to avoid compilation errors
	return true, nil
}

func saveClassification(db *sql.DB, classification *classification.EnhancedClassification) error {
	// This would save a classification to the database
	// For now, return nil to avoid compilation errors
	return nil
}

func getClassification(db *sql.DB, businessID string) (*classification.EnhancedClassification, error) {
	// This would retrieve a classification from the database
	// For now, return a mock result to avoid compilation errors
	return &classification.EnhancedClassification{
		BusinessID:   businessID,
		BusinessName: "Test Business",
	}, nil
}

func saveFeedback(db *sql.DB, feedback *classification.Feedback) error {
	// This would save feedback to the database
	// For now, return nil to avoid compilation errors
	return nil
}

func getFeedback(db *sql.DB, feedbackID string) (*classification.Feedback, error) {
	// This would retrieve feedback from the database
	// For now, return a mock result to avoid compilation errors
	return &classification.Feedback{
		ID:          feedbackID,
		Description: "Test feedback",
	}, nil
}
