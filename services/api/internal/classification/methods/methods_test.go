package methods

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func TestDescriptionClassificationMethod(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	method := NewDescriptionClassificationMethod(logger)

	// Test basic properties
	if method.GetName() != "description_classification" {
		t.Errorf("Expected name 'description_classification', got '%s'", method.GetName())
	}

	if method.GetType() != "description" {
		t.Errorf("Expected type 'description', got '%s'", method.GetType())
	}

	// Test weight management
	originalWeight := method.GetWeight()
	method.SetWeight(0.8)
	if method.GetWeight() != 0.8 {
		t.Errorf("Expected weight 0.8, got %.2f", method.GetWeight())
	}
	method.SetWeight(originalWeight)

	// Test enabled state
	if !method.IsEnabled() {
		t.Error("Expected method to be enabled by default")
	}

	method.SetEnabled(false)
	if method.IsEnabled() {
		t.Error("Expected method to be disabled")
	}
	method.SetEnabled(true)

	// Test initialization
	ctx := context.Background()
	if err := method.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Test classification
	result, err := method.Classify(ctx, "Medical Clinic", "Healthcare services and medical treatment", "https://clinic.com")
	if err != nil {
		t.Errorf("Classification failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful classification, got error: %s", result.Error)
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Invalid confidence score: %.2f", result.Confidence)
	}

	t.Logf("Description method result: %s (confidence: %.2f%%)", result.Result.IndustryName, result.Confidence*100)

	// Test cleanup
	if err := method.Cleanup(); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
}

func TestMethodPerformanceMetrics(t *testing.T) {
	metrics := NewMethodPerformanceMetrics()

	// Test initial state
	if metrics.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests, got %d", metrics.TotalRequests)
	}

	// Test successful request
	metrics.UpdateMetrics(true, 100*time.Millisecond, nil)
	if metrics.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", metrics.SuccessfulRequests)
	}
	if metrics.FailedRequests != 0 {
		t.Errorf("Expected 0 failed requests, got %d", metrics.FailedRequests)
	}

	// Test failed request
	metrics.UpdateMetrics(false, 200*time.Millisecond, nil)
	if metrics.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", metrics.SuccessfulRequests)
	}
	if metrics.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", metrics.FailedRequests)
	}

	// Test accuracy update
	metrics.UpdateAccuracy(0.85)
	if metrics.AccuracyScore != 0.85 {
		t.Errorf("Expected accuracy 0.85, got %.2f", metrics.AccuracyScore)
	}
}
