package e2e

import (
	"time"

	"kyb-platform/internal/database"
)

// Handler types and request/response structures for E2E testing

// User Management Types

type CreateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Company   string `json:"company"`
	Role      string `json:"role"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Company   string    `json:"company"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

type InitialSetupRequest struct {
	UserID               string                  `json:"user_id"`
	DefaultRiskLevel     string                  `json:"default_risk_level"`
	NotificationPrefs    NotificationPreferences `json:"notification_preferences"`
	DashboardPreferences DashboardPreferences    `json:"dashboard_preferences"`
}

type NotificationPreferences struct {
	EmailNotifications bool `json:"email_notifications"`
	RiskAlerts         bool `json:"risk_alerts"`
	ComplianceAlerts   bool `json:"compliance_alerts"`
}

type DashboardPreferences struct {
	DefaultView     string `json:"default_view"`
	ItemsPerPage    int    `json:"items_per_page"`
	AutoRefresh     bool   `json:"auto_refresh"`
	RefreshInterval int    `json:"refresh_interval"`
}

type UserDashboardResponse struct {
	UserID               string `json:"user_id"`
	TotalMerchants       int    `json:"total_merchants"`
	PendingVerifications int    `json:"pending_verifications"`
	HighRiskMerchants    int    `json:"high_risk_merchants"`
	ComplianceAlerts     int    `json:"compliance_alerts"`
}

// Merchant Management Types

type CreateMerchantRequest struct {
	Name               string               `json:"name"`
	LegalName          string               `json:"legal_name"`
	RegistrationNumber string               `json:"registration_number"`
	TaxID              string               `json:"tax_id"`
	Industry           string               `json:"industry"`
	IndustryCode       string               `json:"industry_code"`
	BusinessType       string               `json:"business_type"`
	EmployeeCount      int                  `json:"employee_count"`
	Address            database.Address     `json:"address"`
	ContactInfo        database.ContactInfo `json:"contact_info"`
	PortfolioType      string               `json:"portfolio_type"`
	RiskLevel          string               `json:"risk_level"`
}

type MerchantResponse struct {
	ID                 string               `json:"id"`
	Name               string               `json:"name"`
	LegalName          string               `json:"legal_name"`
	RegistrationNumber string               `json:"registration_number"`
	TaxID              string               `json:"tax_id"`
	Industry           string               `json:"industry"`
	IndustryCode       string               `json:"industry_code"`
	BusinessType       string               `json:"business_type"`
	EmployeeCount      int                  `json:"employee_count"`
	Address            database.Address     `json:"address"`
	ContactInfo        database.ContactInfo `json:"contact_info"`
	PortfolioType      string               `json:"portfolio_type"`
	RiskLevel          string               `json:"risk_level"`
	Status             string               `json:"status"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

type UpdateMerchantRequest struct {
	Name               *string               `json:"name,omitempty"`
	LegalName          *string               `json:"legal_name,omitempty"`
	RegistrationNumber *string               `json:"registration_number,omitempty"`
	TaxID              *string               `json:"tax_id,omitempty"`
	Industry           *string               `json:"industry,omitempty"`
	IndustryCode       *string               `json:"industry_code,omitempty"`
	BusinessType       *string               `json:"business_type,omitempty"`
	EmployeeCount      *int                  `json:"employee_count,omitempty"`
	Address            *database.Address     `json:"address,omitempty"`
	ContactInfo        *database.ContactInfo `json:"contact_info,omitempty"`
	PortfolioType      *string               `json:"portfolio_type,omitempty"`
	RiskLevel          *string               `json:"risk_level,omitempty"`
	VerificationStatus *string               `json:"verification_status,omitempty"`
	VerificationScore  *float64              `json:"verification_score,omitempty"`
	LastVerifiedAt     *time.Time            `json:"last_verified_at,omitempty"`
	WebsiteVerified    *bool                 `json:"website_verified,omitempty"`
}

type MerchantListResponse struct {
	Merchants []MerchantResponse `json:"merchants"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

type MerchantDashboardResponse struct {
	MerchantID           string `json:"merchant_id"`
	TotalVerifications   int    `json:"total_verifications"`
	PendingVerifications int    `json:"pending_verifications"`
	RiskAlerts           int    `json:"risk_alerts"`
	ComplianceStatus     string `json:"compliance_status"`
}

// Business Verification Types

type WebsiteScrapingRequest struct {
	MerchantID      string          `json:"merchant_id"`
	WebsiteURL      string          `json:"website_url"`
	ScrapingOptions ScrapingOptions `json:"scraping_options"`
}

type ScrapingOptions struct {
	ExtractBusinessInfo bool `json:"extract_business_info"`
	ExtractContactInfo  bool `json:"extract_contact_info"`
	ExtractProducts     bool `json:"extract_products"`
	ExtractAboutPage    bool `json:"extract_about_page"`
	MaxPages            int  `json:"max_pages"`
	TimeoutSeconds      int  `json:"timeout_seconds"`
}

type WebsiteScrapingResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ScrapingStatusResponse struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
}

type OwnershipVerificationRequest struct {
	MerchantID           string `json:"merchant_id"`
	WebsiteURL           string `json:"website_url"`
	VerificationMethod   string `json:"verification_method"`
	ExpectedBusinessName string `json:"expected_business_name"`
	ExpectedCountry      string `json:"expected_country"`
}

type OwnershipVerificationResponse struct {
	VerificationStatus string                 `json:"verification_status"`
	ConfidenceScore    float64                `json:"confidence_score"`
	MatchedData        map[string]interface{} `json:"matched_data"`
	Discrepancies      []string               `json:"discrepancies"`
}

type DataValidationRequest struct {
	MerchantID      string   `json:"merchant_id"`
	ValidationTypes []string `json:"validation_types"`
}

type DataValidationResponse struct {
	ValidationResults []ValidationResult `json:"validation_results"`
	OverallScore      float64            `json:"overall_score"`
	Status            string             `json:"status"`
}

type ValidationResult struct {
	Type    string                 `json:"type"`
	Status  string                 `json:"status"`
	Score   float64                `json:"score"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

// Classification Types

type ClassificationRequest struct {
	BusinessName          string                `json:"business_name"`
	BusinessDescription   string                `json:"business_description"`
	WebsiteURL            string                `json:"website_url"`
	Country               string                `json:"country"`
	Industry              string                `json:"industry"`
	BusinessType          string                `json:"business_type"`
	EmployeeCount         int                   `json:"employee_count"`
	ClassificationOptions ClassificationOptions `json:"classification_options"`
}

type ClassificationOptions struct {
	UseWebsiteAnalysis  bool    `json:"use_website_analysis"`
	UseMLClassification bool    `json:"use_ml_classification"`
	UseKeywordMatching  bool    `json:"use_keyword_matching"`
	UseWebSearch        bool    `json:"use_web_search"`
	ConfidenceThreshold float64 `json:"confidence_threshold"`
	MaxResults          int     `json:"max_results"`
}

type ClassificationResponse struct {
	ClassificationID  string         `json:"classification_id"`
	BusinessName      string         `json:"business_name"`
	PrimaryIndustry   string         `json:"primary_industry"`
	OverallConfidence float64        `json:"overall_confidence"`
	MethodResults     []MethodResult `json:"method_results"`
	IndustryCodes     []IndustryCode `json:"industry_codes"`
	ProcessingTime    time.Duration  `json:"processing_time"`
	CreatedAt         time.Time      `json:"created_at"`
}

type MethodResult struct {
	Method          string        `json:"method"`
	Confidence      float64       `json:"confidence"`
	PrimaryIndustry string        `json:"primary_industry"`
	ProcessingTime  time.Duration `json:"processing_time"`
}

type IndustryCode struct {
	CodeType    string  `json:"code_type"`
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

type ClassificationHistoryResponse struct {
	Classifications []ClassificationResponse `json:"classifications"`
	Total           int                      `json:"total"`
}

// Risk Assessment Types

type RiskAssessmentRequest struct {
	BusinessID        string                `json:"business_id"`
	BusinessName      string                `json:"business_name"`
	WebsiteURL        string                `json:"website_url"`
	Industry          string                `json:"industry"`
	Country           string                `json:"country"`
	BusinessType      string                `json:"business_type"`
	AssessmentTypes   []string              `json:"assessment_types"`
	AssessmentOptions RiskAssessmentOptions `json:"assessment_options"`
}

type RiskAssessmentOptions struct {
	IncludeDetailedReport   bool    `json:"include_detailed_report"`
	GenerateRecommendations bool    `json:"generate_recommendations"`
	RiskThreshold           float64 `json:"risk_threshold"`
	TimeoutSeconds          int     `json:"timeout_seconds"`
}

type RiskAssessmentResponse struct {
	AssessmentID       string               `json:"assessment_id"`
	BusinessID         string               `json:"business_id"`
	BusinessName       string               `json:"business_name"`
	WebsiteURL         string               `json:"website_url"`
	OverallRiskScore   float64              `json:"overall_risk_score"`
	RiskLevel          string               `json:"risk_level"`
	SecurityAnalysis   *SecurityAnalysis    `json:"security_analysis,omitempty"`
	DomainAnalysis     *DomainAnalysis      `json:"domain_analysis,omitempty"`
	ReputationAnalysis *ReputationAnalysis  `json:"reputation_analysis,omitempty"`
	ComplianceAnalysis *ComplianceAnalysis  `json:"compliance_analysis,omitempty"`
	FinancialAnalysis  *FinancialAnalysis   `json:"financial_analysis,omitempty"`
	Recommendations    []RiskRecommendation `json:"recommendations"`
	AssessmentDate     time.Time            `json:"assessment_date"`
	ProcessingTime     time.Duration        `json:"processing_time"`
}

type SecurityAnalysis struct {
	SSLScore             float64          `json:"ssl_score"`
	TLSScore             float64          `json:"tls_score"`
	SecurityHeaders      []SecurityHeader `json:"security_headers"`
	Vulnerabilities      []Vulnerability  `json:"vulnerabilities"`
	OverallSecurityScore float64          `json:"overall_security_score"`
	Recommendations      []string         `json:"recommendations"`
}

type SecurityHeader struct {
	Name    string `json:"name"`
	Present bool   `json:"present"`
	Value   string `json:"value"`
}

type Vulnerability struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

type DomainAnalysis struct {
	DomainAge          int         `json:"domain_age"`
	Registrar          string      `json:"registrar"`
	RegistrationDate   time.Time   `json:"registration_date"`
	ExpirationDate     time.Time   `json:"expiration_date"`
	DNSSEC             bool        `json:"dnssec"`
	DNSRecords         []DNSRecord `json:"dns_records"`
	OverallDomainScore float64     `json:"overall_domain_score"`
	Recommendations    []string    `json:"recommendations"`
}

type DNSRecord struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

type ReputationAnalysis struct {
	OverallScore        float64               `json:"overall_score"`
	SocialMediaPresence []SocialMediaPresence `json:"social_media_presence"`
	OnlineReviews       []OnlineReview        `json:"online_reviews"`
	BrandMentions       []BrandMention        `json:"brand_mentions"`
	Recommendations     []string              `json:"recommendations"`
}

type SocialMediaPresence struct {
	Platform   string  `json:"platform"`
	Followers  int     `json:"followers"`
	Engagement float64 `json:"engagement"`
}

type OnlineReview struct {
	Platform    string  `json:"platform"`
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"review_count"`
}

type BrandMention struct {
	Source    string    `json:"source"`
	Sentiment string    `json:"sentiment"`
	Date      time.Time `json:"date"`
}

type ComplianceAnalysis struct {
	OverallComplianceScore float64           `json:"overall_compliance_score"`
	ComplianceChecks       []ComplianceCheck `json:"compliance_checks"`
	Certifications         []Certification   `json:"certifications"`
	Recommendations        []string          `json:"recommendations"`
}

type ComplianceCheck struct {
	Type   string  `json:"type"`
	Status string  `json:"status"`
	Score  float64 `json:"score"`
}

type Certification struct {
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	ExpirationDate time.Time `json:"expiration_date"`
}

type FinancialAnalysis struct {
	OverallFinancialScore float64            `json:"overall_financial_score"`
	RevenueIndicators     []RevenueIndicator `json:"revenue_indicators"`
	StabilityMetrics      []StabilityMetric  `json:"stability_metrics"`
	Recommendations       []string           `json:"recommendations"`
}

type RevenueIndicator struct {
	Type       string  `json:"type"`
	Value      float64 `json:"value"`
	Confidence float64 `json:"confidence"`
}

type StabilityMetric struct {
	Type       string  `json:"type"`
	Value      float64 `json:"value"`
	Confidence float64 `json:"confidence"`
}

type RiskRecommendation struct {
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Impact      string `json:"impact"`
}

type RiskAssessmentHistoryResponse struct {
	Assessments []RiskAssessmentResponse `json:"assessments"`
	Total       int                      `json:"total"`
}

// Session Management Types

type SessionResponse struct {
	ID         string     `json:"id"`
	MerchantID string     `json:"merchant_id"`
	UserID     string     `json:"user_id"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Status     string     `json:"status"`
}

// Bulk Operations Types

type BulkUpdateRequest struct {
	MerchantIDs   []string `json:"merchant_ids"`
	PortfolioType string   `json:"portfolio_type,omitempty"`
	RiskLevel     string   `json:"risk_level,omitempty"`
}

type BulkOperationResponse struct {
	SuccessfulUpdates int      `json:"successful_updates"`
	FailedUpdates     int      `json:"failed_updates"`
	Errors            []string `json:"errors"`
}

// Search Types

type SearchMerchantsRequest struct {
	Query    string                 `json:"query"`
	Filters  *MerchantSearchFilters `json:"filters,omitempty"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type MerchantSearchFilters struct {
	PortfolioType string `json:"portfolio_type"`
	RiskLevel     string `json:"risk_level"`
	Industry      string `json:"industry"`
	Country       string `json:"country"`
}

// Comparison Types

type ComparisonRequest struct {
	Merchant1ID    string `json:"merchant1_id"`
	Merchant2ID    string `json:"merchant2_id"`
	ComparisonType string `json:"comparison_type"`
}

type ComparisonResponse struct {
	ID             string           `json:"id"`
	Merchant1ID    string           `json:"merchant1_id"`
	Merchant2ID    string           `json:"merchant2_id"`
	ComparisonType string           `json:"comparison_type"`
	Similarities   []ComparisonItem `json:"similarities"`
	Differences    []ComparisonItem `json:"differences"`
	Score          float64          `json:"score"`
	CreatedAt      time.Time        `json:"created_at"`
}

type ComparisonItem struct {
	Field  string      `json:"field"`
	Value1 interface{} `json:"value1"`
	Value2 interface{} `json:"value2"`
	Weight float64     `json:"weight"`
}

type ComparisonReportRequest struct {
	Merchant1ID   string `json:"merchant1_id"`
	Merchant2ID   string `json:"merchant2_id"`
	ReportType    string `json:"report_type"`
	IncludeCharts bool   `json:"include_charts"`
}

type ComparisonReportResponse struct {
	ReportID string `json:"report_id"`
	Status   string `json:"status"`
	URL      string `json:"url,omitempty"`
}

// Export Types

type PortfolioExportRequest struct {
	Format     string `json:"format"`
	IncludeAll bool   `json:"include_all"`
	DateRange  string `json:"date_range"`
}

type PortfolioExportResponse struct {
	ExportID string `json:"export_id"`
	Status   string `json:"status"`
	URL      string `json:"url,omitempty"`
}

// Compliance Types

type ComplianceDashboardResponse struct {
	TotalMerchants        int `json:"total_merchants"`
	CompliantMerchants    int `json:"compliant_merchants"`
	NonCompliantMerchants int `json:"non_compliant_merchants"`
	PendingReviews        int `json:"pending_reviews"`
}

type ComplianceAlertsResponse struct {
	Alerts []ComplianceAlert `json:"alerts"`
	Total  int               `json:"total"`
}

type ComplianceAlert struct {
	ID         string    `json:"id"`
	MerchantID string    `json:"merchant_id"`
	Type       string    `json:"type"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type ComplianceReportRequest struct {
	ReportType     string `json:"report_type"`
	DateRange      string `json:"date_range"`
	IncludeDetails bool   `json:"include_details"`
}

type ComplianceReportResponse struct {
	ReportID string `json:"report_id"`
	Status   string `json:"status"`
	URL      string `json:"url,omitempty"`
}

// Analytics Types

type AnalyticsDashboardResponse struct {
	TotalMerchants       int            `json:"total_merchants"`
	ActiveMerchants      int            `json:"active_merchants"`
	RiskDistribution     map[string]int `json:"risk_distribution"`
	IndustryDistribution map[string]int `json:"industry_distribution"`
}

// Risk Dashboard Types

type RiskDashboardResponse struct {
	TotalMerchants      int `json:"total_merchants"`
	HighRiskMerchants   int `json:"high_risk_merchants"`
	MediumRiskMerchants int `json:"medium_risk_merchants"`
	LowRiskMerchants    int `json:"low_risk_merchants"`
	RiskAlerts          int `json:"risk_alerts"`
}
