package risk_assessment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RiskAssessmentService provides comprehensive risk assessment capabilities
type RiskAssessmentService struct {
	config             *RiskAssessmentConfig
	logger             *zap.Logger
	securityAnalyzer   *SecurityAnalyzer
	domainAnalyzer     *DomainAnalyzer
	reputationAnalyzer *ReputationAnalyzer
	complianceAnalyzer *ComplianceAnalyzer
	financialAnalyzer  *FinancialAnalyzer
	riskScorer         *RiskScorer
	rateLimiter        *RateLimiter
	errorTracker       *ErrorTracker
}

// RiskAssessmentConfig contains configuration for risk assessment
type RiskAssessmentConfig struct {
	SecurityAnalysisEnabled     bool                `json:"security_analysis_enabled"`
	DomainAnalysisEnabled       bool                `json:"domain_analysis_enabled"`
	ReputationAnalysisEnabled   bool                `json:"reputation_analysis_enabled"`
	ComplianceAnalysisEnabled   bool                `json:"compliance_analysis_enabled"`
	FinancialAnalysisEnabled    bool                `json:"financial_analysis_enabled"`
	MaxConcurrentRequests       int                 `json:"max_concurrent_requests"`
	RequestTimeout              time.Duration       `json:"request_timeout"`
	RateLimitPerMinute          int                 `json:"rate_limit_per_minute"`
	MaxErrorRate                float64             `json:"max_error_rate"`
	UserAgentRotationEnabled    bool                `json:"user_agent_rotation_enabled"`
	ProxySupportEnabled         bool                `json:"proxy_support_enabled"`
	SSLVerificationEnabled      bool                `json:"ssl_verification_enabled"`
	SecurityHeadersCheckEnabled bool                `json:"security_headers_check_enabled"`
	WHOISLookupEnabled          bool                `json:"whois_lookup_enabled"`
	SocialMediaAnalysisEnabled  bool                `json:"social_media_analysis_enabled"`
	ReviewAnalysisEnabled       bool                `json:"review_analysis_enabled"`
	SentimentAnalysisEnabled    bool                `json:"sentiment_analysis_enabled"`
	ComplianceCheckEnabled      bool                `json:"compliance_check_enabled"`
	FinancialDataEnabled        bool                `json:"financial_data_enabled"`
	AntiDetectionConfig         AntiDetectionConfig `json:"anti_detection_config"`
}

// RiskAssessmentRequest represents a risk assessment request
type RiskAssessmentRequest struct {
	BusinessName    string            `json:"business_name"`
	WebsiteURL      string            `json:"website_url"`
	DomainName      string            `json:"domain_name"`
	Industry        string            `json:"industry"`
	BusinessType    string            `json:"business_type"`
	AdditionalData  map[string]string `json:"additional_data"`
	AnalysisOptions *AnalysisOptions  `json:"analysis_options"`
}

// AnalysisOptions specifies which analyses to perform
type AnalysisOptions struct {
	SecurityAnalysis     bool `json:"security_analysis"`
	DomainAnalysis       bool `json:"domain_analysis"`
	ReputationAnalysis   bool `json:"reputation_analysis"`
	ComplianceAnalysis   bool `json:"compliance_analysis"`
	FinancialAnalysis    bool `json:"financial_analysis"`
	ComprehensiveScoring bool `json:"comprehensive_scoring"`
}

// RiskAssessmentResult represents the complete risk assessment result
type RiskAssessmentResult struct {
	RequestID           string                    `json:"request_id"`
	BusinessName        string                    `json:"business_name"`
	WebsiteURL          string                    `json:"website_url"`
	DomainName          string                    `json:"domain_name"`
	AssessmentTimestamp time.Time                 `json:"assessment_timestamp"`
	ProcessingTime      time.Duration             `json:"processing_time"`
	OverallRiskScore    float64                   `json:"overall_risk_score"`
	RiskLevel           RiskLevel                 `json:"risk_level"`
	RiskCategory        RiskCategory              `json:"risk_category"`
	SecurityAnalysis    *SecurityAnalysisResult   `json:"security_analysis,omitempty"`
	DomainAnalysis      *DomainAnalysisResult     `json:"domain_analysis,omitempty"`
	ReputationAnalysis  *ReputationAnalysisResult `json:"reputation_analysis,omitempty"`
	ComplianceAnalysis  *ComplianceAnalysisResult `json:"compliance_analysis,omitempty"`
	FinancialAnalysis   *FinancialAnalysisResult  `json:"financial_analysis,omitempty"`
	RiskFactors         []RiskFactor              `json:"risk_factors"`
	Recommendations     []Recommendation          `json:"recommendations"`
	ConfidenceScore     float64                   `json:"confidence_score"`
	ErrorRate           float64                   `json:"error_rate"`
}

// RiskLevel represents the overall risk level
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// RiskCategory represents the type of risk
type RiskCategory string

const (
	RiskCategorySecurity    RiskCategory = "security"
	RiskCategoryReputation  RiskCategory = "reputation"
	RiskCategoryCompliance  RiskCategory = "compliance"
	RiskCategoryFinancial   RiskCategory = "financial"
	RiskCategoryOperational RiskCategory = "operational"
)

// RiskFactor represents a specific risk factor
type RiskFactor struct {
	Category    RiskCategory `json:"category"`
	Factor      string       `json:"factor"`
	Description string       `json:"description"`
	Severity    RiskLevel    `json:"severity"`
	Score       float64      `json:"score"`
	Evidence    string       `json:"evidence"`
	Impact      string       `json:"impact"`
}

// Recommendation represents a risk mitigation recommendation
type Recommendation struct {
	Category    RiskCategory `json:"category"`
	Priority    string       `json:"priority"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Action      string       `json:"action"`
	Impact      string       `json:"impact"`
	Effort      string       `json:"effort"`
	Timeline    string       `json:"timeline"`
	Cost        string       `json:"cost"`
}

// DefaultRiskAssessmentConfig returns default configuration for risk assessment
func DefaultRiskAssessmentConfig() *RiskAssessmentConfig {
	return &RiskAssessmentConfig{
		SecurityAnalysisEnabled:     true,
		DomainAnalysisEnabled:       true,
		ReputationAnalysisEnabled:   true,
		ComplianceAnalysisEnabled:   true,
		FinancialAnalysisEnabled:    false, // Disabled by default due to data availability
		MaxConcurrentRequests:       10,
		RequestTimeout:              30 * time.Second,
		RateLimitPerMinute:          60,
		MaxErrorRate:                0.05, // 5%
		UserAgentRotationEnabled:    true,
		ProxySupportEnabled:         false,
		SSLVerificationEnabled:      true,
		SecurityHeadersCheckEnabled: true,
		WHOISLookupEnabled:          true,
		SocialMediaAnalysisEnabled:  true,
		ReviewAnalysisEnabled:       true,
		SentimentAnalysisEnabled:    true,
		ComplianceCheckEnabled:      true,
		FinancialDataEnabled:        false,
	}
}

// NewRiskAssessmentService creates a new risk assessment service
func NewRiskAssessmentService(config *RiskAssessmentConfig, logger *zap.Logger) *RiskAssessmentService {
	if config == nil {
		config = DefaultRiskAssessmentConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	service := &RiskAssessmentService{
		config: config,
		logger: logger,
	}

	// Initialize analyzers
	if config.SecurityAnalysisEnabled {
		service.securityAnalyzer = NewSecurityAnalyzer(config, logger)
	}
	if config.DomainAnalysisEnabled {
		service.domainAnalyzer = NewDomainAnalyzer(config, logger)
	}
	if config.ReputationAnalysisEnabled {
		service.reputationAnalyzer = NewReputationAnalyzer(config, logger)
	}
	if config.ComplianceAnalysisEnabled {
		service.complianceAnalyzer = NewComplianceAnalyzer(config, logger)
	}
	if config.FinancialAnalysisEnabled {
		service.financialAnalyzer = NewFinancialAnalyzer(config, logger)
	}

	// Initialize core components
	service.riskScorer = NewRiskScorer(config, logger)
	service.rateLimiter = NewRateLimiter(config, logger)
	service.errorTracker = NewErrorTracker(config, logger)

	return service
}

// AssessRisk performs comprehensive risk assessment
func (ras *RiskAssessmentService) AssessRisk(ctx context.Context, req *RiskAssessmentRequest) (*RiskAssessmentResult, error) {
	startTime := time.Now()
	requestID := generateRequestID()

	ras.logger.Info("Starting risk assessment",
		zap.String("request_id", requestID),
		zap.String("business_name", req.BusinessName),
		zap.String("website_url", req.WebsiteURL))

	// Validate request
	if err := ras.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check rate limits
	rateLimitResult, err := ras.rateLimiter.CheckRateLimit(ctx, "risk_assessment")
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}
	if !rateLimitResult.Allowed {
		return nil, fmt.Errorf("rate limit exceeded, retry after %v", rateLimitResult.RetryAfter)
	}

	// Initialize result
	result := &RiskAssessmentResult{
		RequestID:           requestID,
		BusinessName:        req.BusinessName,
		WebsiteURL:          req.WebsiteURL,
		DomainName:          req.DomainName,
		AssessmentTimestamp: time.Now(),
		RiskFactors:         make([]RiskFactor, 0),
		Recommendations:     make([]Recommendation, 0),
	}

	// Determine analysis options
	options := ras.determineAnalysisOptions(req)

	// Perform security analysis
	if options.SecurityAnalysis && ras.securityAnalyzer != nil {
		securityResult, err := ras.securityAnalyzer.AnalyzeSecurity(ctx, req)
		if err != nil {
			ras.logger.Warn("Security analysis failed", zap.Error(err))
			ras.errorTracker.TrackError(ctx, "security_analysis", err.Error(), "security_analyzer")
		} else {
			result.SecurityAnalysis = securityResult
			ras.addRiskFactorsFromSecurity(result, securityResult)
		}
	}

	// Perform domain analysis
	if options.DomainAnalysis && ras.domainAnalyzer != nil {
		domainName := req.DomainName
		if domainName == "" && req.WebsiteURL != "" {
			domainName = extractDomainFromURL(req.WebsiteURL)
		}
		if domainName != "" {
			domainResult, err := ras.domainAnalyzer.AnalyzeDomain(ctx, domainName)
			if err != nil {
				ras.logger.Warn("Domain analysis failed", zap.Error(err))
				ras.errorTracker.TrackError(ctx, "domain_analysis", err.Error(), "domain_analyzer")
			} else {
				result.DomainAnalysis = domainResult
				ras.addRiskFactorsFromDomain(result, domainResult)
			}
		}
	}

	// Perform reputation analysis
	if options.ReputationAnalysis && ras.reputationAnalyzer != nil {
		reputationResult, err := ras.reputationAnalyzer.AnalyzeReputation(ctx, req.BusinessName, req.WebsiteURL)
		if err != nil {
			ras.logger.Warn("Reputation analysis failed", zap.Error(err))
			ras.errorTracker.TrackError(ctx, "reputation_analysis", err.Error(), "reputation_analyzer")
		} else {
			result.ReputationAnalysis = reputationResult
			ras.addRiskFactorsFromReputation(result, reputationResult)
		}
	}

	// Perform compliance analysis
	if options.ComplianceAnalysis && ras.complianceAnalyzer != nil {
		complianceResult, err := ras.complianceAnalyzer.AnalyzeCompliance(ctx, req.BusinessName, req.WebsiteURL, req.Industry)
		if err != nil {
			ras.logger.Warn("Compliance analysis failed", zap.Error(err))
			ras.errorTracker.TrackError(ctx, "compliance_analysis", err.Error(), "compliance_analyzer")
		} else {
			result.ComplianceAnalysis = complianceResult
			ras.addRiskFactorsFromCompliance(result, complianceResult)
		}
	}

	// Perform financial analysis
	if options.FinancialAnalysis && ras.financialAnalyzer != nil {
		financialResult, err := ras.financialAnalyzer.AnalyzeFinancial(ctx, req.BusinessName, req.Industry)
		if err != nil {
			ras.logger.Warn("Financial analysis failed", zap.Error(err))
			ras.errorTracker.TrackError(ctx, "financial_analysis", err.Error(), "financial_analyzer")
		} else {
			result.FinancialAnalysis = financialResult
			ras.addRiskFactorsFromFinancial(result, financialResult)
		}
	}

	// Calculate overall risk score and level
	if options.ComprehensiveScoring && ras.riskScorer != nil {
		riskScore, err := ras.riskScorer.CalculateRiskScore(ctx, result)
		if err != nil {
			ras.logger.Warn("Risk score calculation failed", zap.Error(err))
		} else {
			result.OverallRiskScore = riskScore.OverallScore
			result.RiskLevel = riskScore.RiskLevel
			result.ConfidenceScore = riskScore.ConfidenceLevel
		}
	}

	// Calculate processing time and error rate
	result.ProcessingTime = time.Since(startTime)
	errorReport := ras.errorTracker.GetErrorReport()
	result.ErrorRate = errorReport.OverallErrorRate

	ras.logger.Info("Risk assessment completed",
		zap.String("request_id", requestID),
		zap.Float64("risk_score", result.OverallRiskScore),
		zap.String("risk_level", string(result.RiskLevel)),
		zap.Duration("processing_time", result.ProcessingTime),
		zap.Float64("error_rate", result.ErrorRate))

	return result, nil
}

// extractDomainFromURL extracts domain name from a URL
func extractDomainFromURL(url string) string {
	// Remove protocol if present
	domain := url
	if strings.HasPrefix(domain, "http://") {
		domain = strings.TrimPrefix(domain, "http://")
	} else if strings.HasPrefix(domain, "https://") {
		domain = strings.TrimPrefix(domain, "https://")
	}

	// Remove path and query parameters
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove www. prefix
	domain = strings.TrimPrefix(domain, "www.")

	return strings.ToLower(domain)
}

// validateRequest validates the risk assessment request
func (ras *RiskAssessmentService) validateRequest(req *RiskAssessmentRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.BusinessName == "" && req.WebsiteURL == "" && req.DomainName == "" {
		return fmt.Errorf("at least one of business_name, website_url, or domain_name must be provided")
	}

	if req.WebsiteURL != "" {
		if !isValidURL(req.WebsiteURL) {
			return fmt.Errorf("invalid website URL format")
		}
	}

	return nil
}

// determineAnalysisOptions determines which analyses to perform
func (ras *RiskAssessmentService) determineAnalysisOptions(req *RiskAssessmentRequest) *AnalysisOptions {
	if req.AnalysisOptions != nil {
		return req.AnalysisOptions
	}

	// Default to all enabled analyses
	return &AnalysisOptions{
		SecurityAnalysis:     ras.config.SecurityAnalysisEnabled,
		DomainAnalysis:       ras.config.DomainAnalysisEnabled,
		ReputationAnalysis:   ras.config.ReputationAnalysisEnabled,
		ComplianceAnalysis:   ras.config.ComplianceAnalysisEnabled,
		FinancialAnalysis:    ras.config.FinancialAnalysisEnabled,
		ComprehensiveScoring: true,
	}
}

// calculateOverallRisk calculates the overall risk score and level
func (ras *RiskAssessmentService) calculateOverallRisk(result *RiskAssessmentResult) {
	if ras.riskScorer != nil {
		riskScore, err := ras.riskScorer.CalculateRiskScore(context.Background(), result)
		if err != nil {
			ras.logger.Warn("Risk score calculation failed", zap.Error(err))
		} else {
			result.OverallRiskScore = riskScore.OverallScore
			result.RiskLevel = riskScore.RiskLevel
			result.ConfidenceScore = riskScore.ConfidenceLevel
		}
	}
}

// generateRecommendations generates risk mitigation recommendations
func (ras *RiskAssessmentService) generateRecommendations(result *RiskAssessmentResult) {
	// Implementation will be added in the risk scorer
	// For now, add basic recommendations based on risk level
	if result.RiskLevel == RiskLevelHigh || result.RiskLevel == RiskLevelCritical {
		result.Recommendations = append(result.Recommendations, Recommendation{
			Category:    "general",
			Title:       "High Risk Detected",
			Description: "This business shows elevated risk indicators. Consider additional due diligence.",
			Priority:    "high",
		})
	}
}

// Helper methods for adding risk factors from different analyses
func (ras *RiskAssessmentService) addRiskFactorsFromSecurity(result *RiskAssessmentResult, securityResult *SecurityAnalysisResult) {
	// Implementation will be added when security analysis is implemented
}

func (ras *RiskAssessmentService) addRiskFactorsFromDomain(result *RiskAssessmentResult, domainResult *DomainAnalysisResult) {
	// Implementation will be added when domain analysis is implemented
}

func (ras *RiskAssessmentService) addRiskFactorsFromReputation(result *RiskAssessmentResult, reputationResult *ReputationAnalysisResult) {
	// Implementation will be added when reputation analysis is implemented
}

func (ras *RiskAssessmentService) addRiskFactorsFromCompliance(result *RiskAssessmentResult, complianceResult *ComplianceAnalysisResult) {
	// Implementation will be added when compliance analysis is implemented
}

func (ras *RiskAssessmentService) addRiskFactorsFromFinancial(result *RiskAssessmentResult, financialResult *FinancialAnalysisResult) {
	// Implementation will be added when financial analysis is implemented
}

// GetErrorRate returns the current error rate
func (ras *RiskAssessmentService) GetErrorRate() float64 {
	if ras.errorTracker != nil {
		errorReport := ras.errorTracker.GetErrorReport()
		return errorReport.OverallErrorRate
	}
	return 0.0
}

// GetRateLimitStatus returns current rate limit status
func (ras *RiskAssessmentService) GetRateLimitStatus() interface{} {
	if ras.rateLimiter != nil {
		// For now, return a simple status structure
		return map[string]interface{}{
			"current_usage": 0,
			"limit":         ras.config.RateLimitPerMinute,
			"reset_time":    time.Now().Add(time.Minute),
		}
	}
	return nil
}

// Helper function to generate request ID
func generateRequestID() string {
	return fmt.Sprintf("risk_%d", time.Now().UnixNano())
}

// Helper function to validate URL
func isValidURL(url string) bool {
	// Basic URL validation - can be enhanced
	return len(url) > 0 && (url[:7] == "http://" || url[:8] == "https://")
}
