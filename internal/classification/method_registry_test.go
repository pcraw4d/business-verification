package classification

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification/methods"
	"github.com/pcraw4d/business-verification/internal/classification/repository"
	"github.com/pcraw4d/business-verification/internal/machine_learning"
	"github.com/pcraw4d/business-verification/internal/shared"
)

// MockKeywordRepository for testing
type MockKeywordRepository struct{}

func (m *MockKeywordRepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	return &repository.Industry{ID: id, Name: "Restaurant"}, nil
}

func (m *MockKeywordRepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	return &repository.Industry{ID: 1, Name: name}, nil
}

func (m *MockKeywordRepository) ListIndustries(ctx context.Context, category string) ([]*repository.Industry, error) {
	return []*repository.Industry{{ID: 1, Name: "Restaurant"}}, nil
}

func (m *MockKeywordRepository) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}

func (m *MockKeywordRepository) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}

func (m *MockKeywordRepository) DeleteIndustry(ctx context.Context, id int) error {
	return nil
}

func (m *MockKeywordRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
	return []*repository.IndustryKeyword{}, nil
}

func (m *MockKeywordRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
	return []*repository.IndustryKeyword{}, nil
}

func (m *MockKeywordRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	return nil
}

func (m *MockKeywordRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	return nil
}

func (m *MockKeywordRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	return []*repository.ClassificationCode{}, nil
}

func (m *MockKeywordRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return []*repository.ClassificationCode{}, nil
}

func (m *MockKeywordRepository) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}

func (m *MockKeywordRepository) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}

func (m *MockKeywordRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	return nil
}

func (m *MockKeywordRepository) GetKeywordWeights(ctx context.Context, industryID int) ([]*repository.KeywordWeight, error) {
	return []*repository.KeywordWeight{}, nil
}

func (m *MockKeywordRepository) GetActiveKeywordWeights(ctx context.Context) ([]*repository.KeywordWeight, error) {
	return []*repository.KeywordWeight{}, nil
}

func (m *MockKeywordRepository) BuildKeywordIndex(ctx context.Context) error {
	return nil
}

func (m *MockKeywordRepository) SearchIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	return []*repository.Industry{}, nil
}

func (m *MockKeywordRepository) GetIndustryStatistics(ctx context.Context) (*repository.IndustryStatistics, error) {
	return &repository.IndustryStatistics{}, nil
}

func (m *MockKeywordRepository) GetIndustryKeywords(ctx context.Context, industry string) ([]string, error) {
	return []string{"restaurant", "food", "dining"}, nil
}

func (m *MockKeywordRepository) GetIndustryCodes(ctx context.Context, industry string) ([]shared.IndustryCode, error) {
	return []shared.IndustryCode{
		{Code: "5812", Type: "NAICS", Description: "Restaurants and Other Eating Places"},
	}, nil
}

func (m *MockKeywordRepository) SaveClassification(ctx context.Context, classification *shared.IndustryClassification) error {
	return nil
}

func (m *MockKeywordRepository) GetClassificationHistory(ctx context.Context, businessName string) ([]shared.IndustryClassification, error) {
	return []shared.IndustryClassification{}, nil
}

// MockMLClassifier for testing
type MockMLClassifier struct{}

func (m *MockMLClassifier) ClassifyContent(ctx context.Context, content string, industry string) (*machine_learning.ClassificationResult, error) {
	return &machine_learning.ClassificationResult{
		Classifications: []machine_learning.Classification{
			{
				IndustryName: "Technology",
				Confidence:   0.75,
			},
		},
		Confidence: 0.75,
	}, nil
}

func TestMethodRegistry_RegisterMethod(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)

	// Create mock dependencies
	mockKeywordRepo := &MockKeywordRepository{}
	mockMLClassifier := &MockMLClassifier{}

	// Create methods
	keywordMethod := methods.NewKeywordClassificationMethod(mockKeywordRepo, logger)
	mlMethod := methods.NewMLClassificationMethod(mockMLClassifier, logger)
	descriptionMethod := methods.NewDescriptionClassificationMethod(logger)

	// Test registering methods
	tests := []struct {
		name    string
		method  ClassificationMethod
		config  MethodConfig
		wantErr bool
	}{
		{
			name:   "register keyword method",
			method: keywordMethod,
			config: MethodConfig{
				Name:        "keyword_classification",
				Type:        "keyword",
				Weight:      0.5,
				Enabled:     true,
				Description: "Keyword-based classification",
			},
			wantErr: false,
		},
		{
			name:   "register ML method",
			method: mlMethod,
			config: MethodConfig{
				Name:        "ml_classification",
				Type:        "ml",
				Weight:      0.4,
				Enabled:     true,
				Description: "ML-based classification",
			},
			wantErr: false,
		},
		{
			name:   "register description method",
			method: descriptionMethod,
			config: MethodConfig{
				Name:        "description_classification",
				Type:        "description",
				Weight:      0.1,
				Enabled:     true,
				Description: "Description-based classification",
			},
			wantErr: false,
		},
		{
			name:   "duplicate method name",
			method: keywordMethod,
			config: MethodConfig{
				Name:        "keyword_classification",
				Type:        "keyword",
				Weight:      0.5,
				Enabled:     true,
				Description: "Duplicate keyword method",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterMethod(tt.method, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMethodRegistry_GetEnabledMethods(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)

	// Create and register methods
	mockKeywordRepo := &MockKeywordRepository{}
	mockMLClassifier := &MockMLClassifier{}

	keywordMethod := methods.NewKeywordClassificationMethod(mockKeywordRepo, logger)
	mlMethod := methods.NewMLClassificationMethod(mockMLClassifier, logger)
	descriptionMethod := methods.NewDescriptionClassificationMethod(logger)

	// Register all methods
	registry.RegisterMethod(keywordMethod, MethodConfig{Name: "keyword_classification", Type: "keyword", Weight: 0.5, Enabled: true})
	registry.RegisterMethod(mlMethod, MethodConfig{Name: "ml_classification", Type: "ml", Weight: 0.4, Enabled: true})
	registry.RegisterMethod(descriptionMethod, MethodConfig{Name: "description_classification", Type: "description", Weight: 0.1, Enabled: false})

	// Get enabled methods
	enabledMethods := registry.GetEnabledMethods()

	// Should have 2 enabled methods
	if len(enabledMethods) != 2 {
		t.Errorf("Expected 2 enabled methods, got %d", len(enabledMethods))
	}

	// Check that the correct methods are enabled
	enabledNames := make(map[string]bool)
	for _, method := range enabledMethods {
		enabledNames[method.GetName()] = true
	}

	if !enabledNames["keyword_classification"] {
		t.Error("Expected keyword_classification to be enabled")
	}
	if !enabledNames["ml_classification"] {
		t.Error("Expected ml_classification to be enabled")
	}
	if enabledNames["description_classification"] {
		t.Error("Expected description_classification to be disabled")
	}
}

func TestMethodRegistry_UpdateMethodConfig(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)

	// Create and register a method
	mockKeywordRepo := &MockKeywordRepository{}
	keywordMethod := methods.NewKeywordClassificationMethod(mockKeywordRepo, logger)

	registry.RegisterMethod(keywordMethod, MethodConfig{
		Name:    "keyword_classification",
		Type:    "keyword",
		Weight:  0.5,
		Enabled: true,
	})

	// Update the configuration
	newConfig := MethodConfig{
		Name:    "keyword_classification",
		Type:    "keyword",
		Weight:  0.7,
		Enabled: false,
	}

	err := registry.UpdateMethodConfig("keyword_classification", newConfig)
	if err != nil {
		t.Errorf("UpdateMethodConfig() error = %v", err)
	}

	// Verify the update
	config, err := registry.GetMethodConfig("keyword_classification")
	if err != nil {
		t.Errorf("GetMethodConfig() error = %v", err)
	}

	if config.Weight != 0.7 {
		t.Errorf("Expected weight 0.7, got %.2f", config.Weight)
	}
	if config.Enabled {
		t.Error("Expected method to be disabled")
	}
}

func TestMethodRegistry_UpdateMethodMetrics(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)

	// Create and register a method
	mockKeywordRepo := &MockKeywordRepository{}
	keywordMethod := methods.NewKeywordClassificationMethod(mockKeywordRepo, logger)

	registry.RegisterMethod(keywordMethod, MethodConfig{
		Name:    "keyword_classification",
		Type:    "keyword",
		Weight:  0.5,
		Enabled: true,
	})

	// Update metrics
	err := registry.UpdateMethodMetrics("keyword_classification", true, 100*time.Millisecond, nil)
	if err != nil {
		t.Errorf("UpdateMethodMetrics() error = %v", err)
	}

	// Get metrics
	metrics, err := registry.GetMethodMetrics("keyword_classification")
	if err != nil {
		t.Errorf("GetMethodMetrics() error = %v", err)
	}

	if metrics.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", metrics.SuccessfulRequests)
	}
	if metrics.LastResponseTime != 100*time.Millisecond {
		t.Errorf("Expected 100ms response time, got %v", metrics.LastResponseTime)
	}
}

func TestMethodRegistry_GetRegistryStats(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)

	// Create and register methods
	mockKeywordRepo := &MockKeywordRepository{}
	mockMLClassifier := &MockMLClassifier{}

	keywordMethod := methods.NewKeywordClassificationMethod(mockKeywordRepo, logger)
	mlMethod := methods.NewMLClassificationMethod(mockMLClassifier, logger)

	registry.RegisterMethod(keywordMethod, MethodConfig{Name: "keyword_classification", Type: "keyword", Weight: 0.5, Enabled: true})
	registry.RegisterMethod(mlMethod, MethodConfig{Name: "ml_classification", Type: "ml", Weight: 0.4, Enabled: false})

	// Get stats
	stats := registry.GetRegistryStats()

	if stats.TotalMethods != 2 {
		t.Errorf("Expected 2 total methods, got %d", stats.TotalMethods)
	}
	if stats.EnabledMethods != 1 {
		t.Errorf("Expected 1 enabled method, got %d", stats.EnabledMethods)
	}
	if stats.DisabledMethods != 1 {
		t.Errorf("Expected 1 disabled method, got %d", stats.DisabledMethods)
	}
	if stats.MethodTypes["keyword"] != 1 {
		t.Errorf("Expected 1 keyword method, got %d", stats.MethodTypes["keyword"])
	}
	if stats.MethodTypes["ml"] != 1 {
		t.Errorf("Expected 1 ml method, got %d", stats.MethodTypes["ml"])
	}
}

func TestWeightConfigurationManager_LoadAndSaveConfiguration(t *testing.T) {
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
}

func TestWeightConfigurationManager_SetMethodWeight(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := tmpDir + "/test_config.json"

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)
	weightManager := NewWeightConfigurationManager(configFile, registry, logger)

	// Set method weight
	err := weightManager.SetMethodWeight("test_method", 0.8)
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

	// Test invalid weight
	err = weightManager.SetMethodWeight("test_method", 1.5)
	if err == nil {
		t.Error("Expected error for invalid weight")
	}
}

func TestModularArchitecture_EndToEnd(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	registry := NewMethodRegistry(logger)

	// Create mock dependencies
	mockKeywordRepo := &MockKeywordRepository{}
	mockMLClassifier := &MockMLClassifier{}

	// Create and register methods
	keywordMethod := methods.NewKeywordClassificationMethod(mockKeywordRepo, logger)
	mlMethod := methods.NewMLClassificationMethod(mockMLClassifier, logger)
	descriptionMethod := methods.NewDescriptionClassificationMethod(logger)

	registry.RegisterMethod(keywordMethod, MethodConfig{Name: "keyword_classification", Type: "keyword", Weight: 0.5, Enabled: true})
	registry.RegisterMethod(mlMethod, MethodConfig{Name: "ml_classification", Type: "ml", Weight: 0.4, Enabled: true})
	registry.RegisterMethod(descriptionMethod, MethodConfig{Name: "description_classification", Type: "description", Weight: 0.1, Enabled: true})

	// Test classification with all methods
	ctx := context.Background()
	businessName := "Test Restaurant"
	description := "A fine dining restaurant serving Italian cuisine"
	websiteURL := "https://testrestaurant.com"

	enabledMethods := registry.GetEnabledMethods()
	if len(enabledMethods) != 3 {
		t.Errorf("Expected 3 enabled methods, got %d", len(enabledMethods))
	}

	// Test each method individually
	for _, method := range enabledMethods {
		result, err := method.Classify(ctx, businessName, description, websiteURL)
		if err != nil {
			t.Errorf("Method %s classification failed: %v", method.GetName(), err)
			continue
		}

		if !result.Success {
			t.Errorf("Method %s returned unsuccessful result: %s", method.GetName(), result.Error)
			continue
		}

		if result.Confidence < 0 || result.Confidence > 1 {
			t.Errorf("Method %s returned invalid confidence: %.2f", method.GetName(), result.Confidence)
		}

		t.Logf("Method %s: %s (confidence: %.2f%%)", method.GetName(), result.Result.IndustryName, result.Confidence*100)
	}

	// Test registry stats
	stats := registry.GetRegistryStats()
	if stats.TotalMethods != 3 {
		t.Errorf("Expected 3 total methods in stats, got %d", stats.TotalMethods)
	}
	if stats.EnabledMethods != 3 {
		t.Errorf("Expected 3 enabled methods in stats, got %d", stats.EnabledMethods)
	}

	t.Logf("Registry stats: %+v", stats)
}
