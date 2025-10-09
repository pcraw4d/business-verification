package external

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// GovernmentClient provides integration with government databases for compliance checks
type GovernmentClient struct {
	*Client
}

// GovernmentResponse represents a response from government databases
type GovernmentResponse struct {
	Results []GovernmentRecord `json:"results"`
	Total   int                `json:"total"`
}

// GovernmentRecord represents a record from government databases
type GovernmentRecord struct {
	ID              string    `json:"id"`
	EntityName      string    `json:"entity_name"`
	EntityType      string    `json:"entity_type"`
	Country         string    `json:"country"`
	Status          string    `json:"status"`
	IssueDate       time.Time `json:"issue_date"`
	ExpiryDate      *time.Time `json:"expiry_date"`
	Description     string    `json:"description"`
	Source          string    `json:"source"`
	Category        string    `json:"category"`
	Severity        string    `json:"severity"`
	ReferenceNumber string    `json:"reference_number"`
	URL             string    `json:"url"`
}

// ComplianceCheckResult represents the result of a compliance check
type ComplianceCheckResult struct {
	BusinessName     string              `json:"business_name"`
	Country          string              `json:"country"`
	Records          []GovernmentRecord  `json:"records"`
	TotalRecords     int                 `json:"total_records"`
	RiskScore        float64             `json:"risk_score"`
	ComplianceStatus string              `json:"compliance_status"`
	Sanctions        []GovernmentRecord  `json:"sanctions"`
	Warnings         []GovernmentRecord  `json:"warnings"`
	LastChecked      time.Time           `json:"last_checked"`
}

// NewGovernmentClient creates a new government database client
func NewGovernmentClient(apiKey string, logger *zap.Logger) *GovernmentClient {
	// Note: This is a mock implementation. Real government APIs would have different endpoints
	config := Config{
		BaseURL:    "https://api.government-db.example.com/v1", // Mock URL
		APIKey:     apiKey,
		Timeout:    20 * time.Second,
		MaxRetries: 3,
	}

	return &GovernmentClient{
		Client: NewClient(config, logger),
	}
}

// CheckSanctions checks if a business is on any sanctions lists
func (c *GovernmentClient) CheckSanctions(ctx context.Context, businessName, country string) (*ComplianceCheckResult, error) {
	c.logger.Info("Checking sanctions lists",
		zap.String("business_name", businessName),
		zap.String("country", country))

	// Mock implementation - in reality, this would call actual government APIs
	// like OFAC (US), EU Sanctions, UN Security Council, etc.
	
	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)

	// Mock data for demonstration
	mockRecords := c.generateMockSanctionsData(businessName, country)
	
	// Analyze the records
	riskScore, complianceStatus := c.analyzeComplianceRisk(mockRecords)
	
	// Categorize records
	sanctions := []GovernmentRecord{}
	warnings := []GovernmentRecord{}
	
	for _, record := range mockRecords {
		switch record.Category {
		case "SANCTIONS", "BLOCKED":
			sanctions = append(sanctions, record)
		case "WARNING", "ALERT":
			warnings = append(warnings, record)
		}
	}

	result := &ComplianceCheckResult{
		BusinessName:     businessName,
		Country:          country,
		Records:          mockRecords,
		TotalRecords:     len(mockRecords),
		RiskScore:        riskScore,
		ComplianceStatus: complianceStatus,
		Sanctions:        sanctions,
		Warnings:         warnings,
		LastChecked:      time.Now(),
	}

	c.logger.Info("Sanctions check completed",
		zap.String("business_name", businessName),
		zap.Int("total_records", len(mockRecords)),
		zap.Int("sanctions", len(sanctions)),
		zap.Int("warnings", len(warnings)),
		zap.Float64("risk_score", riskScore),
		zap.String("compliance_status", complianceStatus))

	return result, nil
}

// CheckRegulatoryCompliance checks regulatory compliance status
func (c *GovernmentClient) CheckRegulatoryCompliance(ctx context.Context, businessName, country, industry string) (*ComplianceCheckResult, error) {
	c.logger.Info("Checking regulatory compliance",
		zap.String("business_name", businessName),
		zap.String("country", country),
		zap.String("industry", industry))

	// Mock implementation - in reality, this would check various regulatory databases
	// like SEC (US), FCA (UK), BaFin (Germany), etc.
	
	// Simulate API call delay
	time.Sleep(150 * time.Millisecond)

	// Mock data for demonstration
	mockRecords := c.generateMockRegulatoryData(businessName, country, industry)
	
	// Analyze the records
	riskScore, complianceStatus := c.analyzeComplianceRisk(mockRecords)
	
	// Categorize records
	sanctions := []GovernmentRecord{}
	warnings := []GovernmentRecord{}
	
	for _, record := range mockRecords {
		switch record.Category {
		case "VIOLATION", "PENALTY", "FINE":
			sanctions = append(sanctions, record)
		case "WARNING", "NOTICE":
			warnings = append(warnings, record)
		}
	}

	result := &ComplianceCheckResult{
		BusinessName:     businessName,
		Country:          country,
		Records:          mockRecords,
		TotalRecords:     len(mockRecords),
		RiskScore:        riskScore,
		ComplianceStatus: complianceStatus,
		Sanctions:        sanctions,
		Warnings:         warnings,
		LastChecked:      time.Now(),
	}

	c.logger.Info("Regulatory compliance check completed",
		zap.String("business_name", businessName),
		zap.Int("total_records", len(mockRecords)),
		zap.Int("violations", len(sanctions)),
		zap.Int("warnings", len(warnings)),
		zap.Float64("risk_score", riskScore),
		zap.String("compliance_status", complianceStatus))

	return result, nil
}

// CheckBusinessRegistration checks business registration status
func (c *GovernmentClient) CheckBusinessRegistration(ctx context.Context, businessName, country string) (*ComplianceCheckResult, error) {
	c.logger.Info("Checking business registration",
		zap.String("business_name", businessName),
		zap.String("country", country))

	// Mock implementation - in reality, this would check business registries
	// like Companies House (UK), SEC EDGAR (US), etc.
	
	// Simulate API call delay
	time.Sleep(200 * time.Millisecond)

	// Mock data for demonstration
	mockRecords := c.generateMockRegistrationData(businessName, country)
	
	// Analyze the records
	riskScore, complianceStatus := c.analyzeComplianceRisk(mockRecords)
	
	// Categorize records
	sanctions := []GovernmentRecord{}
	warnings := []GovernmentRecord{}
	
	for _, record := range mockRecords {
		switch record.Category {
		case "UNREGISTERED", "SUSPENDED", "REVOKED":
			sanctions = append(sanctions, record)
		case "PENDING", "EXPIRED":
			warnings = append(warnings, record)
		}
	}

	result := &ComplianceCheckResult{
		BusinessName:     businessName,
		Country:          country,
		Records:          mockRecords,
		TotalRecords:     len(mockRecords),
		RiskScore:        riskScore,
		ComplianceStatus: complianceStatus,
		Sanctions:        sanctions,
		Warnings:         warnings,
		LastChecked:      time.Now(),
	}

	c.logger.Info("Business registration check completed",
		zap.String("business_name", businessName),
		zap.Int("total_records", len(mockRecords)),
		zap.Float64("risk_score", riskScore),
		zap.String("compliance_status", complianceStatus))

	return result, nil
}

// generateMockSanctionsData generates mock sanctions data for testing
func (c *GovernmentClient) generateMockSanctionsData(businessName, country string) []GovernmentRecord {
	// This is mock data for demonstration purposes
	// In a real implementation, this would be replaced with actual API calls
	
	records := []GovernmentRecord{}
	
	// Simulate some risk scenarios
	if businessName == "High Risk Corp" {
		records = append(records, GovernmentRecord{
			ID:              "SAN001",
			EntityName:      businessName,
			EntityType:      "Corporation",
			Country:         country,
			Status:          "Active",
			IssueDate:       time.Now().AddDate(-1, 0, 0),
			Description:     "Entity listed on OFAC sanctions list",
			Source:          "US Treasury OFAC",
			Category:        "SANCTIONS",
			Severity:        "HIGH",
			ReferenceNumber: "OFAC-2023-001",
		})
	} else if businessName == "Warning Corp" {
		records = append(records, GovernmentRecord{
			ID:              "WARN001",
			EntityName:      businessName,
			EntityType:      "Corporation",
			Country:         country,
			Status:          "Active",
			IssueDate:       time.Now().AddDate(0, -6, 0),
			Description:     "Entity under investigation for potential sanctions violations",
			Source:          "EU Sanctions Office",
			Category:        "WARNING",
			Severity:        "MEDIUM",
			ReferenceNumber: "EU-WARN-2023-001",
		})
	}
	
	return records
}

// generateMockRegulatoryData generates mock regulatory data for testing
func (c *GovernmentClient) generateMockRegulatoryData(businessName, country, industry string) []GovernmentRecord {
	records := []GovernmentRecord{}
	
	// Simulate some regulatory scenarios
	if businessName == "Finance Corp" && industry == "Financial Services" {
		records = append(records, GovernmentRecord{
			ID:              "REG001",
			EntityName:      businessName,
			EntityType:      "Corporation",
			Country:         country,
			Status:          "Active",
			IssueDate:       time.Now().AddDate(0, -3, 0),
			Description:     "Regulatory violation: Inadequate AML procedures",
			Source:          "Financial Conduct Authority",
			Category:        "VIOLATION",
			Severity:        "HIGH",
			ReferenceNumber: "FCA-2023-001",
		})
	}
	
	return records
}

// generateMockRegistrationData generates mock registration data for testing
func (c *GovernmentClient) generateMockRegistrationData(businessName, country string) []GovernmentRecord {
	records := []GovernmentRecord{}
	
	// Simulate some registration scenarios
	if businessName == "Unregistered Corp" {
		records = append(records, GovernmentRecord{
			ID:              "REG002",
			EntityName:      businessName,
			EntityType:      "Corporation",
			Country:         country,
			Status:          "Not Found",
			IssueDate:       time.Now().AddDate(0, -1, 0),
			Description:     "Entity not found in business registry",
			Source:          "Companies House",
			Category:        "UNREGISTERED",
			Severity:        "HIGH",
			ReferenceNumber: "CH-2023-001",
		})
	}
	
	return records
}

// analyzeComplianceRisk analyzes compliance records to determine risk score and status
func (c *GovernmentClient) analyzeComplianceRisk(records []GovernmentRecord) (float64, string) {
	if len(records) == 0 {
		return 0.0, "COMPLIANT"
	}

	riskScore := 0.0
	complianceStatus := "COMPLIANT"

	for _, record := range records {
		switch record.Category {
		case "SANCTIONS", "BLOCKED", "UNREGISTERED", "SUSPENDED", "REVOKED":
			riskScore += 0.8
			complianceStatus = "NON_COMPLIANT"
		case "VIOLATION", "PENALTY", "FINE":
			riskScore += 0.6
			complianceStatus = "VIOLATION"
		case "WARNING", "ALERT", "NOTICE":
			riskScore += 0.3
			if complianceStatus == "COMPLIANT" {
				complianceStatus = "WARNING"
			}
		case "PENDING", "EXPIRED":
			riskScore += 0.2
			if complianceStatus == "COMPLIANT" {
				complianceStatus = "PENDING"
			}
		}
	}

	// Average the risk scores
	riskScore = riskScore / float64(len(records))

	// Ensure risk score is between 0 and 1
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	return riskScore, complianceStatus
}

// IsHealthy checks if the government database service is healthy
func (c *GovernmentClient) IsHealthy(ctx context.Context) error {
	// Mock health check - in reality, this would ping the actual government APIs
	time.Sleep(50 * time.Millisecond)
	return nil
}
