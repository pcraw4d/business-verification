package shared

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Core Models Tests
// =============================================================================

func TestBusinessClassificationRequest(t *testing.T) {
	req := &BusinessClassificationRequest{
		ID:                 "test-123",
		BusinessName:       "Test Business",
		BusinessType:       "LLC",
		Industry:           "Technology",
		Description:        "A test business",
		Keywords:           []string{"test", "business"},
		WebsiteURL:         "https://testbusiness.com",
		RegistrationNumber: "123456789",
		TaxID:              "12-3456789",
		Address:            "123 Test St, Test City, TS 12345",
		GeographicRegion:   "North America",
		Metadata:           map[string]interface{}{"source": "test"},
		RequestedAt:        time.Now(),
	}

	assert.NotNil(t, req)
	assert.Equal(t, "test-123", req.ID)
	assert.Equal(t, "Test Business", req.BusinessName)
	assert.Equal(t, "LLC", req.BusinessType)
	assert.Equal(t, "Technology", req.Industry)
	assert.Len(t, req.Keywords, 2)
	assert.Equal(t, "https://testbusiness.com", req.WebsiteURL)
}

func TestBusinessClassificationResponse(t *testing.T) {
	classification := &IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "ml",
		Keywords:             []string{"software", "technology"},
		Description:          "Software publishing industry",
		Evidence:             "ML model prediction",
		ProcessingTime:       1500 * time.Millisecond,
		Metadata:             map[string]interface{}{"model_version": "1.0"},
	}

	resp := &BusinessClassificationResponse{
		ID:                    "resp-123",
		BusinessName:          "Test Business",
		Classifications:       []IndustryClassification{*classification},
		PrimaryClassification: classification,
		OverallConfidence:     0.85,
		ClassificationMethod:  "ensemble",
		ProcessingTime:        2000 * time.Millisecond,
		ModuleResults: map[string]ModuleResult{
			"ml_module": {
				ModuleID:        "ml_module",
				ModuleType:      "ml",
				Success:         true,
				Classifications: []IndustryClassification{*classification},
				ProcessingTime:  1500 * time.Millisecond,
				Confidence:      0.85,
				RawData:         map[string]interface{}{"model_predictions": []interface{}{}},
				Metadata:        map[string]interface{}{"model_version": "1.0"},
			},
		},
		RawData:   map[string]interface{}{"raw_data": "test"},
		CreatedAt: time.Now(),
		Metadata:  map[string]interface{}{"source": "test"},
	}

	assert.NotNil(t, resp)
	assert.Equal(t, "resp-123", resp.ID)
	assert.Equal(t, "Test Business", resp.BusinessName)
	assert.Len(t, resp.Classifications, 1)
	assert.NotNil(t, resp.PrimaryClassification)
	assert.Equal(t, 0.85, resp.OverallConfidence)
	assert.Equal(t, "ensemble", resp.ClassificationMethod)
	assert.Len(t, resp.ModuleResults, 1)
}

func TestIndustryClassification(t *testing.T) {
	classification := &IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "ml",
		Keywords:             []string{"software", "technology"},
		Description:          "Software publishing industry",
		Evidence:             "ML model prediction",
		ProcessingTime:       1500 * time.Millisecond,
		Metadata:             map[string]interface{}{"model_version": "1.0"},
	}

	assert.NotNil(t, classification)
	assert.Equal(t, "511210", classification.IndustryCode)
	assert.Equal(t, "Software Publishers", classification.IndustryName)
	assert.Equal(t, 0.85, classification.ConfidenceScore)
	assert.Equal(t, "ml", classification.ClassificationMethod)
	assert.Len(t, classification.Keywords, 2)
	assert.Equal(t, "Software publishing industry", classification.Description)
	assert.Equal(t, "ML model prediction", classification.Evidence)
	assert.Equal(t, 1500*time.Millisecond, classification.ProcessingTime)
}

func TestModuleResult(t *testing.T) {
	classification := &IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "ml",
	}

	result := &ModuleResult{
		ModuleID:        "ml_module",
		ModuleType:      "ml",
		Success:         true,
		Classifications: []IndustryClassification{*classification},
		ProcessingTime:  1500 * time.Millisecond,
		Confidence:      0.85,
		RawData:         map[string]interface{}{"model_predictions": []interface{}{}},
		Metadata:        map[string]interface{}{"model_version": "1.0"},
	}

	assert.NotNil(t, result)
	assert.Equal(t, "ml_module", result.ModuleID)
	assert.Equal(t, "ml", result.ModuleType)
	assert.True(t, result.Success)
	assert.Len(t, result.Classifications, 1)
	assert.Equal(t, 1500*time.Millisecond, result.ProcessingTime)
	assert.Equal(t, 0.85, result.Confidence)
}

// =============================================================================
// Batch Processing Tests
// =============================================================================

func TestBatchClassificationRequest(t *testing.T) {
	requests := []BusinessClassificationRequest{
		{
			ID:           "req-1",
			BusinessName: "Business 1",
			BusinessType: "LLC",
		},
		{
			ID:           "req-2",
			BusinessName: "Business 2",
			BusinessType: "Corporation",
		},
	}

	batchReq := &BatchClassificationRequest{
		ID:          "batch-123",
		Requests:    requests,
		BatchSize:   10,
		Concurrency: 5,
		Timeout:     30 * time.Second,
		Metadata:    map[string]interface{}{"source": "batch_test"},
		RequestedAt: time.Now(),
	}

	assert.NotNil(t, batchReq)
	assert.Equal(t, "batch-123", batchReq.ID)
	assert.Len(t, batchReq.Requests, 2)
	assert.Equal(t, 10, batchReq.BatchSize)
	assert.Equal(t, 5, batchReq.Concurrency)
	assert.Equal(t, 30*time.Second, batchReq.Timeout)
}

func TestBatchClassificationResponse(t *testing.T) {
	responses := []BusinessClassificationResponse{
		{
			ID:                "resp-1",
			BusinessName:      "Business 1",
			OverallConfidence: 0.85,
		},
		{
			ID:                "resp-2",
			BusinessName:      "Business 2",
			OverallConfidence: 0.90,
		},
	}

	errors := []BatchError{
		{
			Index:        0,
			BusinessName: "Failed Business",
			Error:        "Processing failed",
			ModuleID:     "ml_module",
		},
	}

	batchResp := &BatchClassificationResponse{
		ID:             "batch-resp-123",
		Responses:      responses,
		TotalCount:     3,
		SuccessCount:   2,
		ErrorCount:     1,
		ProcessingTime: 5000 * time.Millisecond,
		Errors:         errors,
		Metadata:       map[string]interface{}{"source": "batch_test"},
		CompletedAt:    time.Now(),
	}

	assert.NotNil(t, batchResp)
	assert.Equal(t, "batch-resp-123", batchResp.ID)
	assert.Len(t, batchResp.Responses, 2)
	assert.Equal(t, 3, batchResp.TotalCount)
	assert.Equal(t, 2, batchResp.SuccessCount)
	assert.Equal(t, 1, batchResp.ErrorCount)
	assert.Equal(t, 5000*time.Millisecond, batchResp.ProcessingTime)
	assert.Len(t, batchResp.Errors, 1)
}

// =============================================================================
// Enhanced Classification Tests
// =============================================================================

func TestEnhancedClassification(t *testing.T) {
	now := time.Now()
	mlModelVersion := "1.0.0"
	mlConfidenceScore := 0.85
	geographicRegion := "North America"
	regionConfidenceScore := 0.90
	classificationAlgorithm := "ensemble"
	processingTimeMS := 1500

	enhanced := &EnhancedClassification{
		ID:                   uuid.New(),
		BusinessName:         "Test Business",
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "ensemble",
		Description:          "A test business",
		CreatedAt:            now,
		UpdatedAt:            now,

		// Enhanced fields
		MLModelVersion:          &mlModelVersion,
		MLConfidenceScore:       &mlConfidenceScore,
		CrosswalkMappings:       map[string]interface{}{"naics_to_sic": "5112"},
		GeographicRegion:        &geographicRegion,
		RegionConfidenceScore:   &regionConfidenceScore,
		IndustrySpecificData:    map[string]interface{}{"software_type": "enterprise"},
		ClassificationAlgorithm: &classificationAlgorithm,
		ValidationRulesApplied:  map[string]interface{}{"rule_1": "passed"},
		ProcessingTimeMS:        &processingTimeMS,
		EnhancedMetadata:        map[string]interface{}{"source": "enhanced_test"},
	}

	assert.NotNil(t, enhanced)
	assert.Equal(t, "Test Business", enhanced.BusinessName)
	assert.Equal(t, "511210", enhanced.IndustryCode)
	assert.Equal(t, "Software Publishers", enhanced.IndustryName)
	assert.Equal(t, 0.85, enhanced.ConfidenceScore)
	assert.Equal(t, "ensemble", enhanced.ClassificationMethod)
	assert.Equal(t, now, enhanced.CreatedAt)
	assert.Equal(t, now, enhanced.UpdatedAt)
	assert.Equal(t, "1.0.0", *enhanced.MLModelVersion)
	assert.Equal(t, 0.85, *enhanced.MLConfidenceScore)
	assert.Equal(t, "North America", *enhanced.GeographicRegion)
	assert.Equal(t, 0.90, *enhanced.RegionConfidenceScore)
	assert.Equal(t, "ensemble", *enhanced.ClassificationAlgorithm)
	assert.Equal(t, 1500, *enhanced.ProcessingTimeMS)
}

// =============================================================================
// ML Classification Tests
// =============================================================================

func TestMLClassificationRequest(t *testing.T) {
	req := &MLClassificationRequest{
		BusinessName:        "Test Business",
		BusinessDescription: "A technology company",
		Keywords:            []string{"software", "technology"},
		WebsiteContent:      "We develop software solutions",
		IndustryHints:       []string{"technology", "software"},
		GeographicRegion:    "North America",
		BusinessType:        "LLC",
		Metadata:            map[string]interface{}{"source": "ml_test"},
	}

	assert.NotNil(t, req)
	assert.Equal(t, "Test Business", req.BusinessName)
	assert.Equal(t, "A technology company", req.BusinessDescription)
	assert.Len(t, req.Keywords, 2)
	assert.Equal(t, "We develop software solutions", req.WebsiteContent)
	assert.Len(t, req.IndustryHints, 2)
	assert.Equal(t, "North America", req.GeographicRegion)
	assert.Equal(t, "LLC", req.BusinessType)
}

func TestMLClassificationResult(t *testing.T) {
	predictions := []ModelPrediction{
		{
			ModelID:         "bert-model-1",
			ModelType:       ModelTypeBERT,
			IndustryCode:    "511210",
			IndustryName:    "Software Publishers",
			ConfidenceScore: 0.85,
			RawScore:        0.87,
		},
	}

	featureImportance := map[string]float64{
		"business_name": 0.3,
		"description":   0.4,
		"keywords":      0.3,
	}

	result := &MLClassificationResult{
		IndustryCode:       "511210",
		IndustryName:       "Software Publishers",
		ConfidenceScore:    0.85,
		ModelType:          ModelTypeBERT,
		ModelVersion:       "1.0.0",
		InferenceTime:      1500 * time.Millisecond,
		ModelPredictions:   predictions,
		EnsembleScore:      0.87,
		FeatureImportance:  featureImportance,
		ProcessingMetadata: map[string]interface{}{"model_version": "1.0.0"},
	}

	assert.NotNil(t, result)
	assert.Equal(t, "511210", result.IndustryCode)
	assert.Equal(t, "Software Publishers", result.IndustryName)
	assert.Equal(t, 0.85, result.ConfidenceScore)
	assert.Equal(t, ModelTypeBERT, result.ModelType)
	assert.Equal(t, "1.0.0", result.ModelVersion)
	assert.Equal(t, 1500*time.Millisecond, result.InferenceTime)
	assert.Len(t, result.ModelPredictions, 1)
	assert.Equal(t, 0.87, result.EnsembleScore)
	assert.Len(t, result.FeatureImportance, 3)
}

func TestModelPrediction(t *testing.T) {
	prediction := &ModelPrediction{
		ModelID:         "bert-model-1",
		ModelType:       ModelTypeBERT,
		IndustryCode:    "511210",
		IndustryName:    "Software Publishers",
		ConfidenceScore: 0.85,
		RawScore:        0.87,
	}

	assert.NotNil(t, prediction)
	assert.Equal(t, "bert-model-1", prediction.ModelID)
	assert.Equal(t, ModelTypeBERT, prediction.ModelType)
	assert.Equal(t, "511210", prediction.IndustryCode)
	assert.Equal(t, "Software Publishers", prediction.IndustryName)
	assert.Equal(t, 0.85, prediction.ConfidenceScore)
	assert.Equal(t, 0.87, prediction.RawScore)
}

func TestModelInfo(t *testing.T) {
	performance := &ModelPerformance{
		Accuracy:        0.85,
		Precision:       0.87,
		Recall:          0.83,
		F1Score:         0.85,
		InferenceTime:   150.0,
		Throughput:      100.0,
		MemoryUsage:     512.0,
		LastEvaluated:   time.Now(),
		EvaluationCount: 1000,
	}

	config := &ModelConfig{
		ModelType:         ModelTypeBERT,
		MaxSequenceLength: 512,
		BatchSize:         32,
		LearningRate:      0.001,
		Epochs:            10,
		ValidationSplit:   0.2,
		Hyperparameters:   map[string]interface{}{"dropout": 0.1},
	}

	modelInfo := &ModelInfo{
		ID:          "bert-model-1",
		Name:        "BERT Classification Model",
		Type:        ModelTypeBERT,
		Version:     "1.0.0",
		Status:      ModelStatusReady,
		LoadedAt:    time.Now(),
		LastUsed:    time.Now(),
		UsageCount:  1000,
		Performance: performance,
		Config:      config,
		Metadata:    map[string]interface{}{"source": "model_test"},
	}

	assert.NotNil(t, modelInfo)
	assert.Equal(t, "bert-model-1", modelInfo.ID)
	assert.Equal(t, "BERT Classification Model", modelInfo.Name)
	assert.Equal(t, ModelTypeBERT, modelInfo.Type)
	assert.Equal(t, "1.0.0", modelInfo.Version)
	assert.Equal(t, ModelStatusReady, modelInfo.Status)
	assert.Equal(t, int64(1000), modelInfo.UsageCount)
	assert.NotNil(t, modelInfo.Performance)
	assert.NotNil(t, modelInfo.Config)
}

// =============================================================================
// Website Analysis Tests
// =============================================================================

func TestWebsiteAnalysisRequest(t *testing.T) {
	req := &WebsiteAnalysisRequest{
		BusinessName:          "Test Business",
		WebsiteURL:            "https://testbusiness.com",
		MaxPages:              5,
		IncludeMeta:           true,
		IncludeStructuredData: true,
		Metadata:              map[string]interface{}{"source": "website_test"},
	}

	assert.NotNil(t, req)
	assert.Equal(t, "Test Business", req.BusinessName)
	assert.Equal(t, "https://testbusiness.com", req.WebsiteURL)
	assert.Equal(t, 5, req.MaxPages)
	assert.True(t, req.IncludeMeta)
	assert.True(t, req.IncludeStructuredData)
}

func TestWebsiteAnalysisResult(t *testing.T) {
	connectionValidation := &ConnectionValidationResult{
		IsValid:          true,
		Confidence:       0.90,
		ValidationMethod: "content_analysis",
		BusinessMatch:    true,
		DomainAge:        365,
		SSLValid:         true,
		ValidationErrors: []string{},
	}

	contentAnalysis := &ContentAnalysisResult{
		ContentQuality:     0.85,
		ContentLength:      5000,
		MetaTags:           map[string]string{"title": "Test Business"},
		StructuredData:     map[string]interface{}{"@type": "Organization"},
		IndustryIndicators: []string{"software", "technology"},
		BusinessKeywords:   []string{"test", "business"},
		ContentType:        "business_website",
	}

	semanticAnalysis := &SemanticAnalysisResult{
		SemanticScore:    0.80,
		TopicModeling:    map[string]float64{"technology": 0.6, "business": 0.4},
		SentimentScore:   0.7,
		KeyPhrases:       []string{"software development", "business solutions"},
		EntityExtraction: map[string]string{"company": "Test Business"},
	}

	industryClassification := []IndustryClassificationResult{
		{
			IndustryCode: "511210",
			IndustryName: "Software Publishers",
			Confidence:   0.85,
			Keywords:     []string{"software", "technology"},
			Evidence:     "Website content analysis",
		},
	}

	pageAnalysis := []PageAnalysisResult{
		{
			URL:            "https://testbusiness.com",
			PageType:       "homepage",
			ContentQuality: 0.85,
			Relevance:      0.90,
			Priority:       1,
		},
	}

	result := &WebsiteAnalysisResult{
		WebsiteURL:             "https://testbusiness.com",
		BusinessName:           "Test Business",
		ConnectionValidation:   connectionValidation,
		ContentAnalysis:        contentAnalysis,
		SemanticAnalysis:       semanticAnalysis,
		IndustryClassification: industryClassification,
		PageAnalysis:           pageAnalysis,
		OverallConfidence:      0.85,
		AnalysisTime:           time.Now(),
		AnalysisMetadata:       map[string]interface{}{"source": "website_test"},
	}

	assert.NotNil(t, result)
	assert.Equal(t, "https://testbusiness.com", result.WebsiteURL)
	assert.Equal(t, "Test Business", result.BusinessName)
	assert.NotNil(t, result.ConnectionValidation)
	assert.NotNil(t, result.ContentAnalysis)
	assert.NotNil(t, result.SemanticAnalysis)
	assert.Len(t, result.IndustryClassification, 1)
	assert.Len(t, result.PageAnalysis, 1)
	assert.Equal(t, 0.85, result.OverallConfidence)
}

// =============================================================================
// Web Search Analysis Tests
// =============================================================================

func TestWebSearchAnalysisRequest(t *testing.T) {
	req := &WebSearchAnalysisRequest{
		BusinessName:  "Test Business",
		SearchQuery:   "Test Business software company",
		BusinessType:  "LLC",
		Industry:      "Technology",
		Address:       "123 Test St, Test City, TS 12345",
		MaxResults:    10,
		SearchEngines: []string{"google", "bing", "duckduckgo"},
		Metadata:      map[string]interface{}{"source": "web_search_test"},
	}

	assert.NotNil(t, req)
	assert.Equal(t, "Test Business", req.BusinessName)
	assert.Equal(t, "Test Business software company", req.SearchQuery)
	assert.Equal(t, "LLC", req.BusinessType)
	assert.Equal(t, "Technology", req.Industry)
	assert.Equal(t, "123 Test St, Test City, TS 12345", req.Address)
	assert.Equal(t, 10, req.MaxResults)
	assert.Len(t, req.SearchEngines, 3)
}

func TestWebSearchAnalysisResult(t *testing.T) {
	searchResults := []SearchResult{
		{
			Title:          "Test Business - Official Website",
			URL:            "https://testbusiness.com",
			Description:    "Official website of Test Business",
			Content:        "We provide software solutions",
			RelevanceScore: 0.95,
			Rank:           1,
			Source:         "google",
			PublishedDate:  &time.Time{},
			Metadata:       map[string]string{"domain": "testbusiness.com"},
		},
	}

	analysisResults := &SearchAnalysisResults{
		TotalResults:       10,
		FilteredResults:    8,
		AverageRelevance:   0.85,
		TopKeywords:        []string{"test", "business", "software"},
		SpamDetected:       0,
		DuplicatesRemoved:  2,
		ContentQuality:     0.80,
		SourceDistribution: map[string]int{"google": 5, "bing": 3, "duckduckgo": 2},
	}

	industryClassification := []IndustryClassificationResult{
		{
			IndustryCode: "511210",
			IndustryName: "Software Publishers",
			Confidence:   0.85,
			Keywords:     []string{"software", "technology"},
			Evidence:     "Search result analysis",
		},
	}

	businessExtraction := &BusinessExtractionResult{
		BusinessName:    "Test Business",
		WebsiteURL:      "https://testbusiness.com",
		PhoneNumber:     "555-123-4567",
		EmailAddress:    "contact@testbusiness.com",
		Address:         "123 Test St, Test City, TS 12345",
		SocialMedia:     map[string]string{"linkedin": "linkedin.com/company/testbusiness"},
		Confidence:      0.90,
		ExtractedFields: map[string]string{"phone": "555-123-4567"},
	}

	result := &WebSearchAnalysisResult{
		SearchQuery:            "Test Business software company",
		BusinessName:           "Test Business",
		SearchResults:          searchResults,
		AnalysisResults:        analysisResults,
		IndustryClassification: industryClassification,
		BusinessExtraction:     businessExtraction,
		OverallConfidence:      0.85,
		SearchTime:             time.Now(),
		AnalysisMetadata:       map[string]interface{}{"source": "web_search_test"},
	}

	assert.NotNil(t, result)
	assert.Equal(t, "Test Business software company", result.SearchQuery)
	assert.Equal(t, "Test Business", result.BusinessName)
	assert.Len(t, result.SearchResults, 1)
	assert.NotNil(t, result.AnalysisResults)
	assert.Len(t, result.IndustryClassification, 1)
	assert.NotNil(t, result.BusinessExtraction)
	assert.Equal(t, 0.85, result.OverallConfidence)
}

// =============================================================================
// Utility Functions Tests
// =============================================================================

func TestGetConfidenceLevel(t *testing.T) {
	assert.Equal(t, ConfidenceLevelHigh, GetConfidenceLevel(0.9))
	assert.Equal(t, ConfidenceLevelHigh, GetConfidenceLevel(0.8))
	assert.Equal(t, ConfidenceLevelMedium, GetConfidenceLevel(0.7))
	assert.Equal(t, ConfidenceLevelMedium, GetConfidenceLevel(0.5))
	assert.Equal(t, ConfidenceLevelLow, GetConfidenceLevel(0.4))
	assert.Equal(t, ConfidenceLevelLow, GetConfidenceLevel(0.1))
}

func TestIsValidClassificationMethod(t *testing.T) {
	assert.True(t, IsValidClassificationMethod(ClassificationMethodKeyword))
	assert.True(t, IsValidClassificationMethod(ClassificationMethodML))
	assert.True(t, IsValidClassificationMethod(ClassificationMethodWebsite))
	assert.True(t, IsValidClassificationMethod(ClassificationMethodWebSearch))
	assert.True(t, IsValidClassificationMethod(ClassificationMethodEnsemble))
	assert.True(t, IsValidClassificationMethod(ClassificationMethodHybrid))
	assert.False(t, IsValidClassificationMethod("invalid_method"))
}

func TestIsValidIndustryType(t *testing.T) {
	assert.True(t, IsValidIndustryType(IndustryTypeAgriculture))
	assert.True(t, IsValidIndustryType(IndustryTypeRetail))
	assert.True(t, IsValidIndustryType(IndustryTypeFood))
	assert.True(t, IsValidIndustryType(IndustryTypeManufacturing))
	assert.True(t, IsValidIndustryType(IndustryTypeTechnology))
	assert.True(t, IsValidIndustryType(IndustryTypeFinance))
	assert.True(t, IsValidIndustryType(IndustryTypeHealthcare))
	assert.True(t, IsValidIndustryType(IndustryTypeOther))
	assert.False(t, IsValidIndustryType("invalid_type"))
}

func TestIsValidBusinessName(t *testing.T) {
	assert.True(t, IsValidBusinessName("Test Business"))
	assert.True(t, IsValidBusinessName("Test-Business"))
	assert.True(t, IsValidBusinessName("Test & Business"))
	assert.True(t, IsValidBusinessName("Test Business (LLC)"))
	assert.False(t, IsValidBusinessName(""))
	assert.False(t, IsValidBusinessName("Test Business with invalid characters @#$%"))
}

func TestIsValidURL(t *testing.T) {
	assert.True(t, IsValidURL("https://testbusiness.com"))
	assert.True(t, IsValidURL("http://testbusiness.com"))
	assert.True(t, IsValidURL("https://www.testbusiness.com/path"))
	assert.False(t, IsValidURL(""))
	assert.False(t, IsValidURL("invalid-url"))
	assert.False(t, IsValidURL("ftp://testbusiness.com"))
}

func TestIsValidEmail(t *testing.T) {
	assert.True(t, IsValidEmail("test@testbusiness.com"))
	assert.True(t, IsValidEmail("test.email@testbusiness.com"))
	assert.True(t, IsValidEmail("test+email@testbusiness.com"))
	assert.False(t, IsValidEmail(""))
	assert.False(t, IsValidEmail("invalid-email"))
	assert.False(t, IsValidEmail("test@"))
}

func TestIsValidPhoneNumber(t *testing.T) {
	assert.True(t, IsValidPhoneNumber("555-123-4567"))
	assert.True(t, IsValidPhoneNumber("555.123.4567"))
	assert.True(t, IsValidPhoneNumber("5551234567"))
	assert.False(t, IsValidPhoneNumber(""))
	assert.False(t, IsValidPhoneNumber("invalid-phone"))
	assert.False(t, IsValidPhoneNumber("555-123"))
}

func TestIsValidIndustryCode(t *testing.T) {
	assert.True(t, IsValidIndustryCode("511210")) // NAICS
	assert.True(t, IsValidIndustryCode("5112"))   // SIC
	assert.True(t, IsValidIndustryCode("5411"))   // MCC
	assert.False(t, IsValidIndustryCode(""))
	assert.False(t, IsValidIndustryCode("invalid"))
	assert.False(t, IsValidIndustryCode("12345")) // 5 digits
}

func TestIsValidConfidenceScore(t *testing.T) {
	assert.True(t, IsValidConfidenceScore(0.0))
	assert.True(t, IsValidConfidenceScore(0.5))
	assert.True(t, IsValidConfidenceScore(1.0))
	assert.False(t, IsValidConfidenceScore(-0.1))
	assert.False(t, IsValidConfidenceScore(1.1))
}

func TestIsValidProcessingTime(t *testing.T) {
	assert.True(t, IsValidProcessingTime(0))
	assert.True(t, IsValidProcessingTime(1*time.Second))
	assert.True(t, IsValidProcessingTime(1*time.Hour))
	assert.True(t, IsValidProcessingTime(24*time.Hour))
	assert.False(t, IsValidProcessingTime(-1*time.Second))
	assert.False(t, IsValidProcessingTime(25*time.Hour))
}

func TestIsValidModelType(t *testing.T) {
	assert.True(t, IsValidModelType(ModelTypeBERT))
	assert.True(t, IsValidModelType(ModelTypeEnsemble))
	assert.True(t, IsValidModelType(ModelTypeTransformer))
	assert.True(t, IsValidModelType(ModelTypeCustom))
	assert.False(t, IsValidModelType("invalid_type"))
}

func TestIsValidModelStatus(t *testing.T) {
	assert.True(t, IsValidModelStatus(ModelStatusLoading))
	assert.True(t, IsValidModelStatus(ModelStatusReady))
	assert.True(t, IsValidModelStatus(ModelStatusError))
	assert.True(t, IsValidModelStatus(ModelStatusUpdating))
	assert.True(t, IsValidModelStatus(ModelStatusDeprecated))
	assert.False(t, IsValidModelStatus("invalid_status"))
}

// =============================================================================
// Sanitization Tests
// =============================================================================

func TestSanitizeBusinessName(t *testing.T) {
	assert.Equal(t, "Test Business", SanitizeBusinessName("  Test   Business  "))
	assert.Equal(t, "Test-Business", SanitizeBusinessName("Test-Business"))
	assert.Equal(t, "Test Business with  invalid chars", SanitizeBusinessName("Test Business with @#$% invalid chars"))
	assert.Equal(t, "Test Business (LLC)", SanitizeBusinessName("Test Business (LLC)"))
}

func TestSanitizeURL(t *testing.T) {
	assert.Equal(t, "https://testbusiness.com", SanitizeURL("testbusiness.com"))
	assert.Equal(t, "https://testbusiness.com", SanitizeURL("https://testbusiness.com"))
	assert.Equal(t, "http://testbusiness.com", SanitizeURL("http://testbusiness.com"))
	assert.Equal(t, "https://testbusiness.com", SanitizeURL("  https://testbusiness.com  "))
}

func TestSanitizeEmail(t *testing.T) {
	assert.Equal(t, "test@testbusiness.com", SanitizeEmail("  TEST@TESTBUSINESS.COM  "))
	assert.Equal(t, "test.email@testbusiness.com", SanitizeEmail("test.email@testbusiness.com"))
	assert.Equal(t, "test@testbusiness.comwithinvalidchars", SanitizeEmail("test@testbusiness.com with invalid chars"))
}

func TestSanitizePhoneNumber(t *testing.T) {
	assert.Equal(t, "555-123-4567", SanitizePhoneNumber("555-123-4567"))
	assert.Equal(t, "555-123-4567", SanitizePhoneNumber("555.123.4567"))
	assert.Equal(t, "555-123-4567", SanitizePhoneNumber("5551234567"))
	assert.Equal(t, "555-123-4567", SanitizePhoneNumber("(555) 123-4567"))
}

func TestSanitizeIndustryCode(t *testing.T) {
	assert.Equal(t, "511210", SanitizeIndustryCode("511210"))
	assert.Equal(t, "005112", SanitizeIndustryCode("5112"))
	assert.Equal(t, "511210", SanitizeIndustryCode("511210abc"))
	assert.Equal(t, "511210", SanitizeIndustryCode("511210123"))
}
