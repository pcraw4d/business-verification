package external

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewContactExtractor(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	if extractor == nil {
		t.Fatal("expected contact extractor to be created")
	}

	if extractor.config == nil {
		t.Fatal("expected config to be initialized")
	}

	if extractor.logger != logger {
		t.Fatal("expected logger to be set")
	}
}

func TestNewContactExtractorWithConfig(t *testing.T) {
	logger := zap.NewNop()
	config := &ContactExtractionConfig{
		EnablePhoneExtraction: true,
		EnableEmailExtraction: false,
		MaxExtractionTime:     60 * time.Second,
		ConfidenceThreshold:   0.8,
	}

	extractor := NewContactExtractorWithConfig(config, logger)

	if extractor == nil {
		t.Fatal("expected contact extractor to be created")
	}

	if extractor.config != config {
		t.Fatal("expected custom config to be set")
	}

	if extractor.logger != logger {
		t.Fatal("expected logger to be set")
	}
}

func TestExtractContactInfo_PhoneNumbers(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := `
		Contact us at (555) 123-4567 for support
		Sales: +1-555-987-6543
		Main office: 555-111-2222
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(contactInfo.PhoneNumbers) == 0 {
		t.Fatal("expected phone numbers to be extracted")
	}

	// Check that we have multiple phone numbers
	if len(contactInfo.PhoneNumbers) < 3 {
		t.Errorf("expected at least 3 phone numbers, got %d", len(contactInfo.PhoneNumbers))
	}

	// Check phone number types
	hasSupport := false
	hasSales := false
	hasMain := false

	for _, phone := range contactInfo.PhoneNumbers {
		if phone.Type == "support" {
			hasSupport = true
		}
		if phone.Type == "sales" {
			hasSales = true
		}
		if phone.Type == "main" {
			hasMain = true
		}

		if phone.ConfidenceScore <= 0 || phone.ConfidenceScore > 1 {
			t.Errorf("invalid confidence score: %f", phone.ConfidenceScore)
		}
	}

	if !hasSupport || !hasSales || !hasMain {
		t.Error("expected different phone types to be identified")
	}
}

func TestExtractContactInfo_EmailAddresses(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := `
		Contact us at info@example.com
		Support: support@example.com
		Sales inquiries: sales@example.com
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(contactInfo.EmailAddresses) == 0 {
		t.Fatal("expected email addresses to be extracted")
	}

	// Check that we have multiple email addresses
	if len(contactInfo.EmailAddresses) < 3 {
		t.Errorf("expected at least 3 email addresses, got %d", len(contactInfo.EmailAddresses))
	}

	// Check email types
	hasInfo := false
	hasSupport := false
	hasSales := false

	for _, email := range contactInfo.EmailAddresses {
		if email.Type == "general" && email.Address == "info@example.com" {
			hasInfo = true
		}
		if email.Type == "support" && email.Address == "support@example.com" {
			hasSupport = true
		}
		if email.Type == "sales" && email.Address == "sales@example.com" {
			hasSales = true
		}

		if email.ConfidenceScore <= 0 || email.ConfidenceScore > 1 {
			t.Errorf("invalid confidence score: %f", email.ConfidenceScore)
		}
	}

	if !hasInfo || !hasSupport || !hasSales {
		t.Error("expected different email types to be identified")
	}
}

func TestExtractContactInfo_PhysicalAddresses(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := `
		Our office is located at 123 Main Street, New York, NY 10001
		Branch office: 456 Oak Avenue, Los Angeles, CA 90210
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(contactInfo.PhysicalAddresses) == 0 {
		t.Fatal("expected physical addresses to be extracted")
	}

	// Check that we have at least one address
	if len(contactInfo.PhysicalAddresses) < 1 {
		t.Errorf("expected at least 1 physical address, got %d", len(contactInfo.PhysicalAddresses))
	}

	for _, address := range contactInfo.PhysicalAddresses {
		if address.ConfidenceScore <= 0 || address.ConfidenceScore > 1 {
			t.Errorf("invalid confidence score: %f", address.ConfidenceScore)
		}

		if address.StreetAddress == "" {
			t.Error("expected street address to be extracted")
		}
	}
}

func TestExtractContactInfo_TeamMembers(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := `
		Our team includes:
		John Smith, CEO, john.smith@example.com
		Jane Doe, Marketing Director, jane.doe@example.com
		Bob Johnson, Senior Developer, bob.johnson@example.com
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(contactInfo.TeamMembers) == 0 {
		t.Fatal("expected team members to be extracted")
	}

	// Check that we have multiple team members
	if len(contactInfo.TeamMembers) < 2 {
		t.Errorf("expected at least 2 team members, got %d", len(contactInfo.TeamMembers))
	}

	// Check team member details
	hasCEO := false
	hasMarketing := false
	hasEngineering := false

	for _, member := range contactInfo.TeamMembers {
		if member.Department == "executive" {
			hasCEO = true
		}
		if member.Department == "marketing" {
			hasMarketing = true
		}
		if member.Department == "engineering" {
			hasEngineering = true
		}

		if member.ConfidenceScore <= 0 || member.ConfidenceScore > 1 {
			t.Errorf("invalid confidence score: %f", member.ConfidenceScore)
		}

		if member.Name == "" {
			t.Error("expected team member name to be extracted")
		}

		if member.Title == "" {
			t.Error("expected team member title to be extracted")
		}
	}

	if !hasCEO || !hasMarketing || !hasEngineering {
		t.Error("expected different departments to be identified")
	}
}

func TestExtractContactInfo_ConfidenceScore(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := `
		Contact: (555) 123-4567
		Email: info@example.com
		Address: 123 Main St, New York, NY 10001
		Team: John Smith, CEO
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if contactInfo.ConfidenceScore <= 0 || contactInfo.ConfidenceScore > 1 {
		t.Errorf("invalid overall confidence score: %f", contactInfo.ConfidenceScore)
	}

	// Should have a reasonable confidence score with all data types
	if contactInfo.ConfidenceScore < 0.5 {
		t.Errorf("expected higher confidence score, got %f", contactInfo.ConfidenceScore)
	}
}

func TestExtractContactInfo_DataQuality(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := `
		Contact: (555) 123-4567
		Email: info@example.com
		Address: 123 Main St, New York, NY 10001
		Team: John Smith, CEO
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check data quality metrics
	if contactInfo.DataQuality.Completeness <= 0 || contactInfo.DataQuality.Completeness > 1 {
		t.Errorf("invalid completeness score: %f", contactInfo.DataQuality.Completeness)
	}

	if contactInfo.DataQuality.Accuracy <= 0 || contactInfo.DataQuality.Accuracy > 1 {
		t.Errorf("invalid accuracy score: %f", contactInfo.DataQuality.Accuracy)
	}

	if contactInfo.DataQuality.Timeliness != 1.0 {
		t.Errorf("expected timeliness to be 1.0 for new extraction, got %f", contactInfo.DataQuality.Timeliness)
	}

	if contactInfo.DataQuality.OverallScore <= 0 || contactInfo.DataQuality.OverallScore > 1 {
		t.Errorf("invalid overall quality score: %f", contactInfo.DataQuality.OverallScore)
	}
}

func TestExtractContactInfo_Validation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := `
		Contact: (555) 123-4567
		Email: info@example.com
		Invalid email: not-an-email
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check validation status
	if contactInfo.ValidationStatus.IsValid {
		t.Error("expected validation to fail with invalid email")
	}

	if len(contactInfo.ValidationStatus.ValidationErrors) == 0 {
		t.Error("expected validation errors to be reported")
	}

	if contactInfo.ValidationStatus.LastValidated.IsZero() {
		t.Error("expected last validated timestamp to be set")
	}
}

func TestExtractContactInfo_PrivacyCompliance(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := `
		Contact: (555) 123-4567
		Email: info@example.com
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check privacy compliance
	if !contactInfo.PrivacyCompliance.IsGDPRCompliant {
		t.Error("expected GDPR compliance to be enabled")
	}

	if contactInfo.PrivacyCompliance.ComplianceScore <= 0 || contactInfo.PrivacyCompliance.ComplianceScore > 1 {
		t.Errorf("invalid compliance score: %f", contactInfo.PrivacyCompliance.ComplianceScore)
	}

	if contactInfo.PrivacyCompliance.LastAudit.IsZero() {
		t.Error("expected last audit timestamp to be set")
	}
}

func TestExtractContactInfo_Anonymization(t *testing.T) {
	logger := zap.NewNop()
	config := getDefaultContactExtractionConfig()
	config.EnableAnonymization = true
	extractor := NewContactExtractorWithConfig(config, logger)

	content := `
		Contact: (555) 123-4567
		Email: john.doe@example.com
	`

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that data is anonymized
	if !contactInfo.PrivacyCompliance.IsAnonymized {
		t.Error("expected data to be anonymized")
	}

	// Check phone number anonymization
	for _, phone := range contactInfo.PhoneNumbers {
		if !strings.Contains(phone.Number, "***") {
			t.Error("expected phone number to be anonymized")
		}
	}

	// Check email anonymization
	for _, email := range contactInfo.EmailAddresses {
		if !strings.Contains(email.Address, "***") {
			t.Error("expected email to be anonymized")
		}
	}
}

func TestExtractContactInfo_Timeout(t *testing.T) {
	logger := zap.NewNop()
	config := getDefaultContactExtractionConfig()
	config.MaxExtractionTime = 1 * time.Millisecond
	extractor := NewContactExtractorWithConfig(config, logger)

	// Create a very large content to trigger timeout
	content := strings.Repeat("Contact: (555) 123-4567\n", 10000)

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still return partial results even with timeout
	if contactInfo == nil {
		t.Fatal("expected contact info to be returned even with timeout")
	}
}

func TestExtractContactInfo_EmptyContent(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if contactInfo == nil {
		t.Fatal("expected contact info to be returned even with empty content")
	}

	// Should have zero confidence score with no data
	if contactInfo.ConfidenceScore != 0.0 {
		t.Errorf("expected zero confidence score, got %f", contactInfo.ConfidenceScore)
	}
}

func TestExtractContactInfo_Metadata(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	content := "Contact: (555) 123-4567"

	contactInfo, err := extractor.ExtractContactInfo(context.Background(), "test-business", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check metadata
	if contactInfo.Metadata == nil {
		t.Fatal("expected metadata to be set")
	}

	if _, exists := contactInfo.Metadata["extraction_duration"]; !exists {
		t.Error("expected extraction duration in metadata")
	}

	if _, exists := contactInfo.Metadata["content_length"]; !exists {
		t.Error("expected content length in metadata")
	}

	if _, exists := contactInfo.Metadata["extraction_methods"]; !exists {
		t.Error("expected extraction methods in metadata")
	}
}

func TestGetConfig(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	config := extractor.GetConfig()
	if config == nil {
		t.Fatal("expected config to be returned")
	}

	// Check default values
	if !config.EnablePhoneExtraction {
		t.Error("expected phone extraction to be enabled by default")
	}

	if !config.EnableEmailExtraction {
		t.Error("expected email extraction to be enabled by default")
	}

	if config.MaxExtractionTime != 30*time.Second {
		t.Errorf("expected default extraction time to be 30s, got %v", config.MaxExtractionTime)
	}
}

func TestContactExtractorUpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	newConfig := &ContactExtractionConfig{
		EnablePhoneExtraction: false,
		EnableEmailExtraction: true,
		MaxExtractionTime:     60 * time.Second,
	}

	err := extractor.UpdateConfig(newConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify config was updated
	config := extractor.GetConfig()
	if config.EnablePhoneExtraction {
		t.Error("expected phone extraction to be disabled")
	}

	if !config.EnableEmailExtraction {
		t.Error("expected email extraction to be enabled")
	}

	if config.MaxExtractionTime != 60*time.Second {
		t.Errorf("expected extraction time to be 60s, got %v", config.MaxExtractionTime)
	}
}

func TestUpdateConfig_NilConfig(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewContactExtractor(logger)

	err := extractor.UpdateConfig(nil)
	if err == nil {
		t.Fatal("expected error when updating with nil config")
	}

	if !strings.Contains(err.Error(), "config cannot be nil") {
		t.Errorf("expected specific error message, got: %v", err)
	}
}
