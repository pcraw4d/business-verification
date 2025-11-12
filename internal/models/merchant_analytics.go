package models

import (
	"time"
)

// AnalyticsData represents comprehensive analytics data for a merchant
type AnalyticsData struct {
	MerchantID     string              `json:"merchantId"`
	Classification ClassificationData  `json:"classification"`
	Security       SecurityData        `json:"security"`
	Quality        QualityData         `json:"quality"`
	Intelligence   IntelligenceData    `json:"intelligence,omitempty"`
	Timestamp      time.Time           `json:"timestamp"`
}

// ClassificationData represents classification information for a merchant
type ClassificationData struct {
	PrimaryIndustry string        `json:"primaryIndustry"`
	ConfidenceScore float64       `json:"confidenceScore"`
	RiskLevel       string        `json:"riskLevel"`
	MCCCodes        []IndustryCode `json:"mccCodes,omitempty"`
	SICCodes        []IndustryCode `json:"sicCodes,omitempty"`
	NAICSCodes      []IndustryCode `json:"naicsCodes,omitempty"`
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// SecurityData represents security metrics for a merchant
type SecurityData struct {
	TrustScore      float64         `json:"trustScore"`
	SSLValid        bool            `json:"sslValid"`
	SSLExpiryDate   *time.Time      `json:"sslExpiryDate,omitempty"`
	SecurityHeaders []SecurityHeader `json:"securityHeaders,omitempty"`
}

// SecurityHeader represents a security header check
type SecurityHeader struct {
	Header  string `json:"header"`
	Present bool   `json:"present"`
	Value   string `json:"value,omitempty"`
}

// QualityData represents data quality metrics
type QualityData struct {
	CompletenessScore float64  `json:"completenessScore"`
	DataPoints        int      `json:"dataPoints"`
	MissingFields     []string `json:"missingFields,omitempty"`
}

// IntelligenceData represents business intelligence data
type IntelligenceData struct {
	BusinessAge   *int     `json:"businessAge,omitempty"`
	EmployeeCount *int     `json:"employeeCount,omitempty"`
	AnnualRevenue *float64 `json:"annualRevenue,omitempty"`
}

// WebsiteAnalysisData represents website analysis results
type WebsiteAnalysisData struct {
	MerchantID      string            `json:"merchantId"`
	WebsiteURL      string            `json:"websiteUrl"`
	SSL             SSLData           `json:"ssl"`
	SecurityHeaders []SecurityHeader  `json:"securityHeaders"`
	Performance     PerformanceData   `json:"performance"`
	Accessibility   AccessibilityData `json:"accessibility"`
	LastAnalyzed    time.Time         `json:"lastAnalyzed"`
}

// SSLData represents SSL certificate information
type SSLData struct {
	Valid      bool       `json:"valid"`
	ExpiryDate *time.Time `json:"expiryDate,omitempty"`
	Issuer     string     `json:"issuer,omitempty"`
	Grade      string     `json:"grade,omitempty"`
}

// PerformanceData represents website performance metrics
type PerformanceData struct {
	LoadTime float64 `json:"loadTime"`
	PageSize int     `json:"pageSize"`
	Requests int     `json:"requests"`
	Score    int     `json:"score"`
}

// AccessibilityData represents website accessibility metrics
type AccessibilityData struct {
	Score  float64  `json:"score"`
	Issues []string `json:"issues,omitempty"`
}

