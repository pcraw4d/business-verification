package classification

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestNewClassificationService(t *testing.T) {
	cfg := &config.ExternalServicesConfig{
		BusinessDataAPI: config.BusinessDataAPIConfig{
			Enabled: true,
			BaseURL: "https://api.example.com",
			APIKey:  "test-key",
			Timeout: 30 * time.Second,
		},
	}

	// Create mock dependencies
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})

	service := NewClassificationService(cfg, nil, logger, metrics)
	if service == nil {
		t.Fatal("Expected classification service to be created")
	}

	if service.config != cfg {
		t.Error("Expected config to match input config")
	}

	if service.industryData != nil {
		t.Error("Expected industry data to be nil for basic constructor")
	}
}

func TestNewClassificationServiceWithData(t *testing.T) {
	cfg := &config.ExternalServicesConfig{
		BusinessDataAPI: config.BusinessDataAPIConfig{
			Enabled: true,
			BaseURL: "https://api.example.com",
			APIKey:  "test-key",
			Timeout: 30 * time.Second,
		},
	}

	// Create mock dependencies
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})

	// Create mock industry data
	industryData := &IndustryCodeData{
		NAICS: map[string]string{
			"541511": "Custom Computer Programming Services",
		},
		MCC: map[string]string{
			"5045": "Computers, Computer Peripheral Equipment, Software",
		},
		SIC: map[string]string{
			"C-35-357-3571": "Electronic Computers",
		},
	}

	service := NewClassificationServiceWithData(cfg, nil, logger, metrics, industryData)
	if service == nil {
		t.Fatal("Expected classification service to be created")
	}

	if service.config != cfg {
		t.Error("Expected config to match input config")
	}

	if service.industryData != industryData {
		t.Error("Expected industry data to match input data")
	}
}

func TestClassificationRequest(t *testing.T) {
	req := &ClassificationRequest{
		BusinessName:       "Tech Solutions Inc",
		BusinessType:       "LLC",
		Industry:           "Technology",
		Description:        "Software development and consulting services",
		Keywords:           "software, technology, consulting",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
	}

	if req.BusinessName != "Tech Solutions Inc" {
		t.Errorf("Expected business name 'Tech Solutions Inc', got %s", req.BusinessName)
	}

	if req.BusinessType != "LLC" {
		t.Errorf("Expected business type 'LLC', got %s", req.BusinessType)
	}

	if req.Industry != "Technology" {
		t.Errorf("Expected industry 'Technology', got %s", req.Industry)
	}

	if req.Description != "Software development and consulting services" {
		t.Errorf("Expected description 'Software development and consulting services', got %s", req.Description)
	}

	if req.Keywords != "software, technology, consulting" {
		t.Errorf("Expected keywords 'software, technology, consulting', got %s", req.Keywords)
	}

	if req.RegistrationNumber != "REG123456" {
		t.Errorf("Expected registration number 'REG123456', got %s", req.RegistrationNumber)
	}

	if req.TaxID != "TAX123456" {
		t.Errorf("Expected tax ID 'TAX123456', got %s", req.TaxID)
	}
}

func TestIndustryClassification(t *testing.T) {
	classification := &IndustryClassification{
		IndustryCode:         "541511",
		IndustryName:         "Custom Computer Programming Services",
		ConfidenceScore:      0.85,
		ClassificationMethod: "keyword_based",
		Keywords:             []string{"software", "technology"},
		Description:          "Classified based on keywords in business description",
	}

	if classification.IndustryCode != "541511" {
		t.Errorf("Expected industry code '541511', got %s", classification.IndustryCode)
	}

	if classification.IndustryName != "Custom Computer Programming Services" {
		t.Errorf("Expected industry name 'Custom Computer Programming Services', got %s", classification.IndustryName)
	}

	if classification.ConfidenceScore != 0.85 {
		t.Errorf("Expected confidence score 0.85, got %f", classification.ConfidenceScore)
	}

	if classification.ClassificationMethod != "keyword_based" {
		t.Errorf("Expected classification method 'keyword_based', got %s", classification.ClassificationMethod)
	}

	if len(classification.Keywords) != 2 {
		t.Errorf("Expected 2 keywords, got %d", len(classification.Keywords))
	}

	if classification.Description != "Classified based on keywords in business description" {
		t.Errorf("Expected description 'Classified based on keywords in business description', got %s", classification.Description)
	}
}

func TestClassificationResponse(t *testing.T) {
	classifications := []IndustryClassification{
		{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.85,
			ClassificationMethod: "keyword_based",
		},
		{
			IndustryCode:         "541512",
			IndustryName:         "Computer Systems Design Services",
			ConfidenceScore:      0.75,
			ClassificationMethod: "name_based",
		},
	}

	response := &ClassificationResponse{
		BusinessID:            "business_123",
		Classifications:       classifications,
		PrimaryClassification: &classifications[0],
		ConfidenceScore:       0.8,
		ClassificationMethod:  "hybrid",
		ProcessingTime:        150 * time.Millisecond,
		RawData: map[string]interface{}{
			"method": "hybrid_classification",
		},
	}

	if response.BusinessID != "business_123" {
		t.Errorf("Expected business ID 'business_123', got %s", response.BusinessID)
	}

	if len(response.Classifications) != 2 {
		t.Errorf("Expected 2 classifications, got %d", len(response.Classifications))
	}

	if response.PrimaryClassification == nil {
		t.Error("Expected primary classification to be set")
	}

	if response.ConfidenceScore != 0.8 {
		t.Errorf("Expected confidence score 0.8, got %f", response.ConfidenceScore)
	}

	if response.ClassificationMethod != "hybrid" {
		t.Errorf("Expected classification method 'hybrid', got %s", response.ClassificationMethod)
	}

	if response.ProcessingTime != 150*time.Millisecond {
		t.Errorf("Expected processing time 150ms, got %v", response.ProcessingTime)
	}

	if response.RawData["method"] != "hybrid_classification" {
		t.Errorf("Expected raw data method 'hybrid_classification', got %v", response.RawData["method"])
	}
}

func TestBatchClassificationRequest(t *testing.T) {
	businesses := []ClassificationRequest{
		{
			BusinessName: "Tech Solutions Inc",
			BusinessType: "LLC",
			Industry:     "Technology",
		},
		{
			BusinessName: "Financial Services Corp",
			BusinessType: "Corporation",
			Industry:     "Finance",
		},
	}

	req := &BatchClassificationRequest{
		Businesses: businesses,
	}

	if len(req.Businesses) != 2 {
		t.Errorf("Expected 2 businesses, got %d", len(req.Businesses))
	}

	if req.Businesses[0].BusinessName != "Tech Solutions Inc" {
		t.Errorf("Expected first business name 'Tech Solutions Inc', got %s", req.Businesses[0].BusinessName)
	}

	if req.Businesses[1].BusinessName != "Financial Services Corp" {
		t.Errorf("Expected second business name 'Financial Services Corp', got %s", req.Businesses[1].BusinessName)
	}
}

func TestBatchClassificationResponse(t *testing.T) {
	results := []ClassificationResponse{
		{
			BusinessID:           "business_1",
			ConfidenceScore:      0.85,
			ClassificationMethod: "hybrid",
		},
		{
			BusinessID:           "business_2",
			ConfidenceScore:      0.75,
			ClassificationMethod: "hybrid",
		},
	}

	response := &BatchClassificationResponse{
		Results:        results,
		TotalProcessed: 2,
		SuccessCount:   2,
		ErrorCount:     0,
		ProcessingTime: 500 * time.Millisecond,
	}

	if len(response.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(response.Results))
	}

	if response.TotalProcessed != 2 {
		t.Errorf("Expected total processed 2, got %d", response.TotalProcessed)
	}

	if response.SuccessCount != 2 {
		t.Errorf("Expected success count 2, got %d", response.SuccessCount)
	}

	if response.ErrorCount != 0 {
		t.Errorf("Expected error count 0, got %d", response.ErrorCount)
	}

	if response.ProcessingTime != 500*time.Millisecond {
		t.Errorf("Expected processing time 500ms, got %v", response.ProcessingTime)
	}
}

func TestValidateClassificationRequest(t *testing.T) {
	service := &ClassificationService{}

	// Test valid request
	validReq := &ClassificationRequest{
		BusinessName: "Valid Business Name",
		Description:  "Valid description",
	}

	if err := service.validateClassificationRequest(validReq); err != nil {
		t.Errorf("Expected valid request to pass validation, got error: %v", err)
	}

	// Test empty business name
	emptyNameReq := &ClassificationRequest{
		BusinessName: "",
	}

	if err := service.validateClassificationRequest(emptyNameReq); err == nil {
		t.Error("Expected empty business name to fail validation")
	}

	// Test business name too long
	longNameReq := &ClassificationRequest{
		BusinessName: string(make([]byte, 501)), // 501 characters
	}

	if err := service.validateClassificationRequest(longNameReq); err == nil {
		t.Error("Expected long business name to fail validation")
	}

	// Test description too long
	longDescReq := &ClassificationRequest{
		BusinessName: "Valid Business",
		Description:  string(make([]byte, 2001)), // 2001 characters
	}

	if err := service.validateClassificationRequest(longDescReq); err == nil {
		t.Error("Expected long description to fail validation")
	}
}

func TestClassifyByKeywords(t *testing.T) {
	service := &ClassificationService{}

	// Test software business
	softwareReq := &ClassificationRequest{
		BusinessName: "Software Solutions Inc",
		Description:  "We develop custom software solutions",
		Keywords:     "software, technology, programming",
	}

	classifications := service.classifyByKeywords(softwareReq)
	if len(classifications) == 0 {
		t.Error("Expected software business to be classified")
	}

	// Test financial business
	financialReq := &ClassificationRequest{
		BusinessName: "Financial Services Corp",
		Description:  "Banking and financial services",
		Keywords:     "financial, banking, credit",
	}

	classifications = service.classifyByKeywords(financialReq)
	if len(classifications) == 0 {
		t.Error("Expected financial business to be classified")
	}

	// Test business with no keywords
	noKeywordsReq := &ClassificationRequest{
		BusinessName: "Generic Business",
		Description:  "A generic business description",
	}

	classifications = service.classifyByKeywords(noKeywordsReq)
	if len(classifications) != 0 {
		t.Error("Expected business with no keywords to not be classified")
	}
}

func TestClassifyByBusinessType(t *testing.T) {
	service := &ClassificationService{}

	// Test LLC
	llcReq := &ClassificationRequest{
		BusinessName: "Test Business",
		BusinessType: "LLC",
	}

	classifications := service.classifyByBusinessType(llcReq)
	if len(classifications) == 0 {
		t.Error("Expected LLC to be classified")
	}

	// Test nonprofit
	nonprofitReq := &ClassificationRequest{
		BusinessName: "Charity Foundation",
		BusinessType: "nonprofit",
	}

	classifications = service.classifyByBusinessType(nonprofitReq)
	if len(classifications) == 0 {
		t.Error("Expected nonprofit to be classified")
	}

	// Test unknown business type
	unknownReq := &ClassificationRequest{
		BusinessName: "Test Business",
		BusinessType: "unknown_type",
	}

	classifications = service.classifyByBusinessType(unknownReq)
	if len(classifications) != 0 {
		t.Error("Expected unknown business type to not be classified")
	}
}

func TestClassifyByIndustry(t *testing.T) {
	service := &ClassificationService{}

	// Test technology industry
	techReq := &ClassificationRequest{
		BusinessName: "Tech Company",
		Industry:     "technology",
	}

	classifications := service.classifyByIndustry(techReq)
	if len(classifications) == 0 {
		t.Error("Expected technology industry to be classified")
	}

	// Test healthcare industry
	healthReq := &ClassificationRequest{
		BusinessName: "Medical Clinic",
		Industry:     "healthcare",
	}

	classifications = service.classifyByIndustry(healthReq)
	if len(classifications) == 0 {
		t.Error("Expected healthcare industry to be classified")
	}

	// Test unknown industry
	unknownReq := &ClassificationRequest{
		BusinessName: "Test Business",
		Industry:     "unknown_industry",
	}

	classifications = service.classifyByIndustry(unknownReq)
	if len(classifications) != 0 {
		t.Error("Expected unknown industry to not be classified")
	}
}

func TestClassifyByName(t *testing.T) {
	service := &ClassificationService{}

	// Test tech company name
	techReq := &ClassificationRequest{
		BusinessName: "Tech Solutions Inc",
	}

	classifications := service.classifyByName(techReq)
	if len(classifications) == 0 {
		t.Error("Expected tech company name to be classified")
	}

	// Test financial company name
	financialReq := &ClassificationRequest{
		BusinessName: "First National Bank",
	}

	classifications = service.classifyByName(financialReq)
	if len(classifications) == 0 {
		t.Error("Expected financial company name to be classified")
	}

	// Test generic name
	genericReq := &ClassificationRequest{
		BusinessName: "Generic Company",
	}

	classifications = service.classifyByName(genericReq)
	if len(classifications) != 0 {
		t.Error("Expected generic company name to not be classified")
	}
}

func TestClassifyByFuzzy(t *testing.T) {
	// Build minimal industry dataset
	industryData := &IndustryCodeData{
		NAICS: map[string]string{
			"541611": "Administrative Management and General Management Consulting Services",
			"541511": "Custom Computer Programming Services",
			"441110": "New Car Dealers",
		},
		MCC: map[string]string{
			"5812": "Eating places and restaurants",
		},
		SIC: map[string]string{
			"D-50-504-5045": "Computers, Peripheral Equipment, and Software",
		},
	}

	logger := observability.NewLogger(&config.ObservabilityConfig{LogLevel: "error"})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{})
	svc := NewClassificationServiceWithData(&config.ExternalServicesConfig{}, nil, logger, metrics, industryData)

	// Slightly misspelled to trigger fuzzy rather than exact keyword match
	req := &ClassificationRequest{
		BusinessName: "Acme Sofware Solutons",
		Description:  "custom programing and managment consultng",
	}

	got := svc.classifyByFuzzy(req)
	if len(got) == 0 {
		t.Fatal("expected fuzzy classification results, got none")
	}
}

func TestIndustryCodeMappingCrosswalk(t *testing.T) {
	industryData := &IndustryCodeData{
		NAICS: map[string]string{
			"541611": "Administrative Management and General Management Consulting Services",
			"541511": "Custom Computer Programming Services",
		},
		MCC: map[string]string{
			"7392": "Management, Consulting, and Public Relations",
			"5732": "Electronic Sales",
		},
		SIC: map[string]string{
			"I-73-739-7392": "Management Consulting Services",
		},
	}

	mcc, sic := crosswalkFromNAICS("541611", industryData)
	if len(mcc) == 0 && len(sic) == 0 {
		t.Fatal("expected crosswalk to produce related MCC or SIC codes")
	}
}

func TestDeterminePrimaryClassification(t *testing.T) {
	service := &ClassificationService{}

	classifications := []IndustryClassification{
		{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.7,
			ClassificationMethod: "keyword_based",
		},
		{
			IndustryCode:         "541512",
			IndustryName:         "Computer Systems Design Services",
			ConfidenceScore:      0.9,
			ClassificationMethod: "name_based",
		},
		{
			IndustryCode:         "541519",
			IndustryName:         "Other Computer Related Services",
			ConfidenceScore:      0.5,
			ClassificationMethod: "industry_based",
		},
	}

	primary := service.determinePrimaryClassification(classifications)
	if primary == nil {
		t.Fatal("Expected primary classification to be determined")
	}

	if primary.IndustryCode != "541512" {
		t.Errorf("Expected primary classification to have highest confidence score, got %s", primary.IndustryCode)
	}

	if primary.ConfidenceScore != 0.9 {
		t.Errorf("Expected primary classification confidence score 0.9, got %f", primary.ConfidenceScore)
	}
}

func TestCalculateOverallConfidence(t *testing.T) {
	service := &ClassificationService{}

	classifications := []IndustryClassification{
		{
			ConfidenceScore: 0.7,
		},
		{
			ConfidenceScore: 0.9,
		},
		{
			ConfidenceScore: 0.5,
		},
	}

	confidence := service.calculateOverallConfidence(classifications)
	expected := (0.7 + 0.9 + 0.5) / 3.0

	// Use approximate comparison for floating point
	if confidence < expected-0.0001 || confidence > expected+0.0001 {
		t.Errorf("Expected overall confidence %f, got %f", expected, confidence)
	}

	// Test empty classifications
	emptyClassifications := []IndustryClassification{}
	confidence = service.calculateOverallConfidence(emptyClassifications)
	if confidence != 0.0 {
		t.Errorf("Expected confidence 0.0 for empty classifications, got %f", confidence)
	}
}

func TestPostProcessConfidence(t *testing.T) {
	svc := &ClassificationService{}
	in := []IndustryClassification{
		{IndustryCode: "541611", ConfidenceScore: 0.6, ClassificationMethod: "keyword_based"},
		{IndustryCode: "541611", ConfidenceScore: 0.55, ClassificationMethod: "name_pattern_based"},
		{IndustryCode: "541511", ConfidenceScore: 0.7, ClassificationMethod: "keyword_based_naics"},
	}
	out := svc.postProcessConfidence(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 unique codes, got %d", len(out))
	}
	// Ensure the agreed code 541611 received a boost and is preserved once
	found541611 := false
	for _, cl := range out {
		if cl.IndustryCode == "541611" {
			found541611 = true
			if cl.ConfidenceScore <= 0.6 {
				t.Errorf("expected boosted score for 541611, got %f", cl.ConfidenceScore)
			}
		}
	}
	if !found541611 {
		t.Error("expected 541611 to be present after dedup")
	}
}

func TestClassificationCaching(t *testing.T) {
	// Prepare minimal industry data for deterministic result
	industryData := &IndustryCodeData{
		NAICS: map[string]string{
			"541511": "Custom Computer Programming Services",
			"541611": "Administrative Management and General Management Consulting Services",
		},
	}

	logger := observability.NewLogger(&config.ObservabilityConfig{LogLevel: "error"})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{})
	cfg := &config.ExternalServicesConfig{
		ClassificationCache: config.ClassificationCacheConfig{
			Enabled:    true,
			TTL:        time.Minute,
			MaxEntries: 100,
		},
	}
	svc := NewClassificationServiceWithData(cfg, nil, logger, metrics, industryData)

	req := &ClassificationRequest{BusinessName: "Acme Software"}

	// First call: cache miss
	resp1, err := svc.ClassifyBusiness(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error on first classify: %v", err)
	}
	if resp1 == nil || resp1.PrimaryClassification == nil {
		t.Fatal("expected primary classification on first call")
	}

	// Second call: should be cache hit with same classifications but new business ID
	resp2, err := svc.ClassifyBusiness(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error on second classify: %v", err)
	}
	if resp2.BusinessID == resp1.BusinessID {
		t.Error("expected different business IDs across requests even when cached")
	}
	if len(resp2.Classifications) != len(resp1.Classifications) {
		t.Error("expected cached classifications to match in length")
	}
}

func TestGetIndustryName(t *testing.T) {
	service := &ClassificationService{}

	// Test known industry code
	name := service.getIndustryName("541511")
	if name != "Custom Computer Programming Services" {
		t.Errorf("Expected industry name 'Custom Computer Programming Services', got %s", name)
	}

	// Test unknown industry code
	name = service.getIndustryName("999999")
	if name != "Unknown Industry" {
		t.Errorf("Expected industry name 'Unknown Industry', got %s", name)
	}
}

func TestGetDefaultClassification(t *testing.T) {
	service := &ClassificationService{}

	classification := service.getDefaultClassification()
	if classification.IndustryCode != "541611" {
		t.Errorf("Expected default industry code '541611', got %s", classification.IndustryCode)
	}

	if classification.IndustryName != "Administrative Management and General Management Consulting Services" {
		t.Errorf("Expected default industry name 'Administrative Management and General Management Consulting Services', got %s", classification.IndustryName)
	}

	if classification.ConfidenceScore != 0.3 {
		t.Errorf("Expected default confidence score 0.3, got %f", classification.ConfidenceScore)
	}

	if classification.ClassificationMethod != "default" {
		t.Errorf("Expected default classification method 'default', got %s", classification.ClassificationMethod)
	}
}

func TestGenerateBusinessID(t *testing.T) {
	service := &ClassificationService{}

	req := &ClassificationRequest{
		BusinessName: "Test Business",
	}

	id1 := service.generateBusinessID(req)
	id2 := service.generateBusinessID(req)

	if id1 == "" {
		t.Error("Expected business ID to be generated")
	}

	if id2 == "" {
		t.Error("Expected business ID to be generated")
	}

	if id1 == id2 {
		t.Error("Expected business IDs to be unique")
	}

	// Check that IDs start with "business_"
	if !strings.HasPrefix(id1, "business_") {
		t.Errorf("Expected business ID to start with 'business_', got %s", id1)
	}

	if !strings.HasPrefix(id2, "business_") {
		t.Errorf("Expected business ID to start with 'business_', got %s", id2)
	}
}
