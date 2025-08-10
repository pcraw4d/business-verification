package database

import (
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

func TestNewDatabaseConfig(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "test",
		Password: "test",
		Database: "test_db",
		SSLMode:  "disable",
	}

	dbConfig := NewDatabaseConfig(cfg)
	if dbConfig == nil {
		t.Fatal("Expected database config to be created")
	}

	if dbConfig.Driver != "postgres" {
		t.Errorf("Expected driver 'postgres', got %s", dbConfig.Driver)
	}

	if dbConfig.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got %s", dbConfig.Host)
	}
}

func TestUserModel(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:                  "user-123",
		Email:               "test@example.com",
		Username:            "testuser",
		PasswordHash:        "hashed_password",
		FirstName:           "John",
		LastName:            "Doe",
		Company:             "Test Corp",
		Role:                "user",
		Status:              "active",
		EmailVerified:       true,
		LastLoginAt:         &now,
		FailedLoginAttempts: 0,
		LockedUntil:         nil,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if user.ID != "user-123" {
		t.Errorf("Expected ID 'user-123', got %s", user.ID)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", user.Username)
	}
}

func TestBusinessModel(t *testing.T) {
	now := time.Now()
	foundedDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	annualRevenue := 1000000.0

	business := &Business{
		ID:                 "business-123",
		Name:               "Test Business",
		LegalName:          "Test Business LLC",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		FoundedDate:        &foundedDate,
		EmployeeCount:      50,
		AnnualRevenue:      &annualRevenue,
		Address: Address{
			Street1:     "123 Main St",
			Street2:     "Suite 100",
			City:        "New York",
			State:       "NY",
			PostalCode:  "10001",
			Country:     "United States",
			CountryCode: "US",
		},
		ContactInfo: ContactInfo{
			Phone:          "+1-555-123-4567",
			Email:          "contact@testbusiness.com",
			Website:        "https://testbusiness.com",
			PrimaryContact: "John Smith",
		},
		Status:           "active",
		RiskLevel:        "low",
		ComplianceStatus: "compliant",
		CreatedBy:        "user-123",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if business.ID != "business-123" {
		t.Errorf("Expected ID 'business-123', got %s", business.ID)
	}

	if business.Name != "Test Business" {
		t.Errorf("Expected name 'Test Business', got %s", business.Name)
	}

	if business.RegistrationNumber != "REG123456" {
		t.Errorf("Expected registration number 'REG123456', got %s", business.RegistrationNumber)
	}

	if business.Address.City != "New York" {
		t.Errorf("Expected city 'New York', got %s", business.Address.City)
	}

	if business.ContactInfo.Email != "contact@testbusiness.com" {
		t.Errorf("Expected contact email 'contact@testbusiness.com', got %s", business.ContactInfo.Email)
	}
}

func TestBusinessClassificationModel(t *testing.T) {
	now := time.Now()
	classification := &BusinessClassification{
		ID:                   "classification-123",
		BusinessID:           "business-123",
		IndustryCode:         "541511",
		IndustryName:         "Custom Computer Programming Services",
		ConfidenceScore:      0.95,
		ClassificationMethod: "ml_model",
		Source:               "internal_classifier",
		RawData:              `{"features": ["software", "programming"], "confidence": 0.95}`,
		CreatedAt:            now,
	}

	if classification.ID != "classification-123" {
		t.Errorf("Expected ID 'classification-123', got %s", classification.ID)
	}

	if classification.BusinessID != "business-123" {
		t.Errorf("Expected business ID 'business-123', got %s", classification.BusinessID)
	}

	if classification.IndustryCode != "541511" {
		t.Errorf("Expected industry code '541511', got %s", classification.IndustryCode)
	}

	if classification.ConfidenceScore != 0.95 {
		t.Errorf("Expected confidence score 0.95, got %f", classification.ConfidenceScore)
	}
}

func TestRiskAssessmentModel(t *testing.T) {
	now := time.Now()
	riskFactors := []string{"new_business", "high_revenue", "international_operations"}

	assessment := &RiskAssessment{
		ID:               "assessment-123",
		BusinessID:       "business-123",
		BusinessName:     "Test Business",
		OverallScore:     0.65,
		OverallLevel:     "medium",
		FactorScores:     riskFactors,
		AssessmentMethod: "rule_based",
		Source:           "internal_assessor",
		Metadata:         map[string]interface{}{"factors": riskFactors, "score": 0.65},
		AssessedAt:       now,
		ValidUntil:       now.AddDate(0, 1, 0), // Valid for 1 month
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if assessment.ID != "assessment-123" {
		t.Errorf("Expected ID 'assessment-123', got %s", assessment.ID)
	}

	if assessment.BusinessID != "business-123" {
		t.Errorf("Expected business ID 'business-123', got %s", assessment.BusinessID)
	}

	if assessment.OverallLevel != "medium" {
		t.Errorf("Expected risk level 'medium', got %s", assessment.OverallLevel)
	}

	if assessment.OverallScore != 0.65 {
		t.Errorf("Expected risk score 0.65, got %f", assessment.OverallScore)
	}

	if len(assessment.FactorScores) != 3 {
		t.Errorf("Expected 3 risk factors, got %d", len(assessment.FactorScores))
	}
}

func TestComplianceCheckModel(t *testing.T) {
	now := time.Now()
	requirements := []string{"kyc_verified", "aml_checked", "sanctions_clear"}

	check := &ComplianceCheck{
		ID:             "check-123",
		BusinessID:     "business-123",
		ComplianceType: "kyc_aml",
		Status:         "passed",
		Score:          0.92,
		Requirements:   requirements,
		CheckMethod:    "automated",
		Source:         "compliance_engine",
		RawData:        `{"checks": ["kyc", "aml", "sanctions"], "score": 0.92}`,
		CreatedAt:      now,
	}

	if check.ID != "check-123" {
		t.Errorf("Expected ID 'check-123', got %s", check.ID)
	}

	if check.BusinessID != "business-123" {
		t.Errorf("Expected business ID 'business-123', got %s", check.BusinessID)
	}

	if check.ComplianceType != "kyc_aml" {
		t.Errorf("Expected compliance type 'kyc_aml', got %s", check.ComplianceType)
	}

	if check.Status != "passed" {
		t.Errorf("Expected status 'passed', got %s", check.Status)
	}

	if check.Score != 0.92 {
		t.Errorf("Expected score 0.92, got %f", check.Score)
	}

	if len(check.Requirements) != 3 {
		t.Errorf("Expected 3 requirements, got %d", len(check.Requirements))
	}
}

func TestAPIKeyModel(t *testing.T) {
	now := time.Now()
	permissionsJSON := `{"read","write","admin"}`

	// Database model stores permissions as JSON string
	apiKey := &APIKey{
		ID:          "key-123",
		UserID:      "user-123",
		Name:        "Test API Key",
		KeyHash:     "hashed_key_value",
		Permissions: permissionsJSON,
		Status:      "active",
		LastUsedAt:  &now,
		ExpiresAt:   nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if apiKey.ID != "key-123" {
		t.Errorf("Expected ID 'key-123', got %s", apiKey.ID)
	}

	if apiKey.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got %s", apiKey.UserID)
	}

	if apiKey.Name != "Test API Key" {
		t.Errorf("Expected name 'Test API Key', got %s", apiKey.Name)
	}

	if apiKey.Status != "active" {
		t.Errorf("Expected status 'active', got %s", apiKey.Status)
	}

	if apiKey.Permissions != permissionsJSON {
		t.Errorf("Expected permissions JSON %s, got %s", permissionsJSON, apiKey.Permissions)
	}
}

func TestAuditLogModel(t *testing.T) {
	now := time.Now()

	log := &AuditLog{
		ID:           "log-123",
		UserID:       "user-123",
		Action:       "create_business",
		ResourceType: "business",
		ResourceID:   "business-123",
		Details:      `{"business_name": "Test Business", "registration_number": "REG123456"}`,
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0...",
		RequestID:    "req-123",
		CreatedAt:    now,
	}

	if log.ID != "log-123" {
		t.Errorf("Expected ID 'log-123', got %s", log.ID)
	}

	if log.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got %s", log.UserID)
	}

	if log.Action != "create_business" {
		t.Errorf("Expected action 'create_business', got %s", log.Action)
	}

	if log.ResourceType != "business" {
		t.Errorf("Expected resource type 'business', got %s", log.ResourceType)
	}

	if log.ResourceID != "business-123" {
		t.Errorf("Expected resource ID 'business-123', got %s", log.ResourceID)
	}

	if log.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP address '192.168.1.1', got %s", log.IPAddress)
	}

	if log.RequestID != "req-123" {
		t.Errorf("Expected request ID 'req-123', got %s", log.RequestID)
	}
}

func TestExternalServiceCallModel(t *testing.T) {
	now := time.Now()

	call := &ExternalServiceCall{
		ID:           "call-123",
		UserID:       "user-123",
		ServiceName:  "business_data_api",
		Endpoint:     "/businesses/search",
		Method:       "GET",
		RequestData:  `{"query": "Test Business"}`,
		ResponseData: `{"results": [{"id": "business-123"}]}`,
		StatusCode:   200,
		Duration:     150,
		Error:        "",
		RequestID:    "req-123",
		CreatedAt:    now,
	}

	if call.ID != "call-123" {
		t.Errorf("Expected ID 'call-123', got %s", call.ID)
	}

	if call.ServiceName != "business_data_api" {
		t.Errorf("Expected service name 'business_data_api', got %s", call.ServiceName)
	}

	if call.Endpoint != "/businesses/search" {
		t.Errorf("Expected endpoint '/businesses/search', got %s", call.Endpoint)
	}

	if call.Method != "GET" {
		t.Errorf("Expected method 'GET', got %s", call.Method)
	}

	if call.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", call.StatusCode)
	}

	if call.Duration != 150 {
		t.Errorf("Expected duration 150, got %d", call.Duration)
	}
}

func TestWebhookModel(t *testing.T) {
	now := time.Now()
	events := []string{"business.created", "business.updated", "business.deleted"}

	webhook := &Webhook{
		ID:              "webhook-123",
		UserID:          "user-123",
		Name:            "Test Webhook",
		URL:             "https://example.com/webhook",
		Events:          events,
		Secret:          "webhook_secret",
		Status:          "active",
		LastTriggeredAt: &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if webhook.ID != "webhook-123" {
		t.Errorf("Expected ID 'webhook-123', got %s", webhook.ID)
	}

	if webhook.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got %s", webhook.UserID)
	}

	if webhook.Name != "Test Webhook" {
		t.Errorf("Expected name 'Test Webhook', got %s", webhook.Name)
	}

	if webhook.URL != "https://example.com/webhook" {
		t.Errorf("Expected URL 'https://example.com/webhook', got %s", webhook.URL)
	}

	if webhook.Status != "active" {
		t.Errorf("Expected status 'active', got %s", webhook.Status)
	}

	if len(webhook.Events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(webhook.Events))
	}
}

func TestWebhookEventModel(t *testing.T) {
	now := time.Now()
	responseCode := 200

	event := &WebhookEvent{
		ID:           "event-123",
		WebhookID:    "webhook-123",
		EventType:    "business.created",
		Payload:      `{"business_id": "business-123", "name": "Test Business"}`,
		Status:       "delivered",
		ResponseCode: &responseCode,
		ResponseBody: "OK",
		Attempts:     1,
		NextRetryAt:  nil,
		CreatedAt:    now,
	}

	if event.ID != "event-123" {
		t.Errorf("Expected ID 'event-123', got %s", event.ID)
	}

	if event.WebhookID != "webhook-123" {
		t.Errorf("Expected webhook ID 'webhook-123', got %s", event.WebhookID)
	}

	if event.EventType != "business.created" {
		t.Errorf("Expected event type 'business.created', got %s", event.EventType)
	}

	if event.Status != "delivered" {
		t.Errorf("Expected status 'delivered', got %s", event.Status)
	}

	if *event.ResponseCode != 200 {
		t.Errorf("Expected response code 200, got %d", *event.ResponseCode)
	}

	if event.Attempts != 1 {
		t.Errorf("Expected attempts 1, got %d", event.Attempts)
	}
}
