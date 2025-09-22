package classification

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/classification/methods"
)

func TestModularArchitecture_Integration(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)

	// Create and register methods
	descriptionMethod := methods.NewDescriptionClassificationMethod(logger)

	// Register the method
	err := registry.RegisterMethod(descriptionMethod, MethodConfig{
		Name:        "description_classification",
		Type:        "description",
		Weight:      0.5,
		Enabled:     true,
		Description: "Description-based classification",
	})
	if err != nil {
		t.Fatalf("Failed to register method: %v", err)
	}

	// Test that the method is registered
	enabledMethods := registry.GetEnabledMethods()
	if len(enabledMethods) != 1 {
		t.Errorf("Expected 1 enabled method, got %d", len(enabledMethods))
	}

	// Test method configuration
	config, err := registry.GetMethodConfig("description_classification")
	if err != nil {
		t.Errorf("Failed to get method config: %v", err)
	}
	if config.Weight != 0.5 {
		t.Errorf("Expected weight 0.5, got %.2f", config.Weight)
	}

	// Test method classification
	ctx := context.Background()
	businessName := "Test Restaurant"
	description := "A fine dining restaurant serving Italian cuisine"
	websiteURL := "https://testrestaurant.com"

	method := enabledMethods[0]
	result, err := method.Classify(ctx, businessName, description, websiteURL)
	if err != nil {
		t.Errorf("Method classification failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful classification, got error: %s", result.Error)
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Invalid confidence score: %.2f", result.Confidence)
	}

	t.Logf("Classification result: %s (confidence: %.2f%%)", result.Result.IndustryName, result.Confidence*100)

	// Test registry stats
	stats := registry.GetRegistryStats()
	if stats.TotalMethods != 1 {
		t.Errorf("Expected 1 total method in stats, got %d", stats.TotalMethods)
	}
	if stats.EnabledMethods != 1 {
		t.Errorf("Expected 1 enabled method in stats, got %d", stats.EnabledMethods)
	}

	t.Logf("Registry stats: %+v", stats)
}

func TestWeightConfigurationManager_Integration(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := tmpDir + "/test_config.json"

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)
	weightManager := NewWeightConfigurationManager(configFile, registry, logger)

	// Test loading non-existent configuration (should create default)
	err := weightManager.LoadConfiguration()
	if err != nil {
		t.Errorf("LoadConfiguration() error = %v", err)
	}

	// Test setting method weight
	err = weightManager.SetMethodWeight("test_method", 0.8)
	if err != nil {
		t.Errorf("SetMethodWeight() error = %v", err)
	}

	// Verify weight was set
	weight, err := weightManager.GetMethodWeight("test_method")
	if err != nil {
		t.Errorf("GetMethodWeight() error = %v", err)
	}
	if weight != 0.8 {
		t.Errorf("Expected weight 0.8, got %.2f", weight)
	}

	// Test saving configuration
	err = weightManager.SaveConfiguration()
	if err != nil {
		t.Errorf("SaveConfiguration() error = %v", err)
	}

	// Test loading saved configuration
	err = weightManager.LoadConfiguration()
	if err != nil {
		t.Errorf("LoadConfiguration() error = %v", err)
	}

	// Verify configuration was loaded
	configs := weightManager.GetAllConfigurations()
	if len(configs) == 0 {
		t.Error("Expected configuration to be loaded")
	}

	// Test weight summary
	summary := weightManager.GetWeightSummary()
	if summary.TotalMethods == 0 {
		t.Error("Expected methods in weight summary")
	}

	t.Logf("Weight summary: %+v", summary)
}

func TestMethodRegistry_ConcurrentAccess(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)

	// Create and register a method
	descriptionMethod := methods.NewDescriptionClassificationMethod(logger)
	registry.RegisterMethod(descriptionMethod, MethodConfig{
		Name:    "description_classification",
		Type:    "description",
		Weight:  0.5,
		Enabled: true,
	})

	// Test concurrent access
	done := make(chan bool, 10)

	// Start multiple goroutines that access the registry
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Test getting enabled methods
			methods := registry.GetEnabledMethods()
			if len(methods) != 1 {
				t.Errorf("Goroutine %d: Expected 1 enabled method, got %d", id, len(methods))
			}

			// Test getting method config
			config, err := registry.GetMethodConfig("description_classification")
			if err != nil {
				t.Errorf("Goroutine %d: Failed to get method config: %v", id, err)
			}
			if config.Weight != 0.5 {
				t.Errorf("Goroutine %d: Expected weight 0.5, got %.2f", id, config.Weight)
			}

			// Test updating metrics
			err = registry.UpdateMethodMetrics("description_classification", true, 100*time.Millisecond, nil)
			if err != nil {
				t.Errorf("Goroutine %d: Failed to update metrics: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	stats := registry.GetRegistryStats()
	if stats.TotalMethods != 1 {
		t.Errorf("Expected 1 total method, got %d", stats.TotalMethods)
	}

	t.Logf("Concurrent access test completed successfully")
}
