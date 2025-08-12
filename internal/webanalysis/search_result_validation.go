package webanalysis

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SearchResultValidator provides comprehensive validation for search results
type SearchResultValidator struct {
	validators []ResultValidator
	config     ValidationConfig
	httpClient *http.Client
	mu         sync.RWMutex
}

// ValidationConfig holds configuration for search result validation
type ValidationConfig struct {
	EnableURLValidation      bool          `json:"enable_url_validation"`
	EnableContentValidation  bool          `json:"enable_content_validation"`
	EnableDomainValidation   bool          `json:"enable_domain_validation"`
	EnableAccessibilityCheck bool          `json:"enable_accessibility_check"`
	EnableSecurityCheck      bool          `json:"enable_security_check"`
	EnableFreshnessCheck     bool          `json:"enable_freshness_check"`
	RequestTimeout           time.Duration `json:"request_timeout"`
	MaxRedirects             int           `json:"max_redirects"`
	UserAgent                string        `json:"user_agent"`
	AllowedStatusCodes       []int         `json:"allowed_status_codes"`
	BlockedDomains           []string      `json:"blocked_domains"`
	BlockedIPRanges          []string      `json:"blocked_ip_ranges"`
	MinContentLength         int           `json:"min_content_length"`
	MaxContentLength         int           `json:"max_content_length"`
	RequiredContentKeywords  []string      `json:"required_content_keywords"`
	BlockedContentKeywords   []string      `json:"blocked_content_keywords"`
	ValidTLDs                []string      `json:"valid_tlds"`
	MaxValidationTime        time.Duration `json:"max_validation_time"`
}

// ResultValidator represents a validator for search results
type ResultValidator interface {
	Validate(ctx context.Context, result *WebSearchResult) (*ValidationResult, error)
	GetName() string
	GetDescription() string
	GetPriority() int
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	ValidatorName  string                 `json:"validator_name"`
	IsValid        bool                   `json:"is_valid"`
	Score          float64                `json:"score"`
	Errors         []string               `json:"errors"`
	Warnings       []string               `json:"warnings"`
	Metadata       map[string]interface{} `json:"metadata"`
	ValidationTime time.Duration          `json:"validation_time"`
}

// ComprehensiveValidationResult represents the overall validation result
type ComprehensiveValidationResult struct {
	Result            *WebSearchResult   `json:"result"`
	OverallValid      bool               `json:"overall_valid"`
	OverallScore      float64            `json:"overall_score"`
	ValidationResults []ValidationResult `json:"validation_results"`
	TotalErrors       int                `json:"total_errors"`
	TotalWarnings     int                `json:"total_warnings"`
	ValidationTime    time.Duration      `json:"validation_time"`
}

// NewSearchResultValidator creates a new search result validator
func NewSearchResultValidator() *SearchResultValidator {
	config := ValidationConfig{
		EnableURLValidation:      true,
		EnableContentValidation:  true,
		EnableDomainValidation:   true,
		EnableAccessibilityCheck: true,
		EnableSecurityCheck:      true,
		EnableFreshnessCheck:     false, // Disabled by default as it requires external data
		RequestTimeout:           time.Second * 10,
		MaxRedirects:             5,
		UserAgent:                "KYB-Platform/1.0 (Search Result Validator)",
		AllowedStatusCodes:       []int{200, 201, 202, 203, 204, 205, 206},
		BlockedDomains: []string{
			"spam.com", "malware.com", "phishing.com", "scam.com",
			"bit.ly", "goo.gl", "tinyurl.com", "t.co",
		},
		BlockedIPRanges: []string{
			"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", // Private IP ranges
		},
		MinContentLength:        50,
		MaxContentLength:        10000,
		RequiredContentKeywords: []string{},
		BlockedContentKeywords: []string{
			"click here", "buy now", "limited time", "act now", "free trial",
			"make money", "earn money", "work from home", "get rich",
			"lose weight", "diet pills", "miracle cure", "100% free",
		},
		ValidTLDs: []string{
			".com", ".org", ".net", ".edu", ".gov", ".mil",
			".co", ".io", ".ai", ".app", ".dev", ".tech",
		},
		MaxValidationTime: time.Second * 30,
	}

	validator := &SearchResultValidator{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= config.MaxRedirects {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}

	// Initialize validators
	validator.initializeValidators()

	return validator
}

// initializeValidators sets up all available validators
func (srv *SearchResultValidator) initializeValidators() {
	srv.validators = []ResultValidator{
		&URLValidator{config: srv.config, httpClient: srv.httpClient},
		&ContentValidator{config: srv.config},
		&DomainValidator{config: srv.config},
		&AccessibilityValidator{config: srv.config, httpClient: srv.httpClient},
		&SecurityValidator{config: srv.config},
		&FreshnessValidator{config: srv.config},
	}
}

// ValidateResult performs comprehensive validation on a search result
func (srv *SearchResultValidator) ValidateResult(ctx context.Context, result *WebSearchResult) (*ComprehensiveValidationResult, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	start := time.Now()

	var validationResults []ValidationResult
	var totalErrors, totalWarnings int
	var overallScore float64

	// Run validators in parallel with timeout
	validationCtx, cancel := context.WithTimeout(ctx, srv.config.MaxValidationTime)
	defer cancel()

	var wg sync.WaitGroup
	resultChan := make(chan ValidationResult, len(srv.validators))
	errorChan := make(chan error, len(srv.validators))

	for _, validator := range srv.validators {
		wg.Add(1)
		go func(v ResultValidator) {
			defer wg.Done()

			validationResult, err := v.Validate(validationCtx, result)
			if err != nil {
				errorChan <- fmt.Errorf("validator %s failed: %w", v.GetName(), err)
				return
			}

			resultChan <- *validationResult
		}(validator)
	}

	// Wait for all validators to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	for validationResult := range resultChan {
		validationResults = append(validationResults, validationResult)
		totalErrors += len(validationResult.Errors)
		totalWarnings += len(validationResult.Warnings)
		overallScore += validationResult.Score
	}

	// Check for validation errors
	for err := range errorChan {
		// Log error but don't fail the entire validation
		fmt.Printf("Validation error: %v\n", err)
	}

	// Calculate overall score
	if len(validationResults) > 0 {
		overallScore = overallScore / float64(len(validationResults))
	}

	// Determine overall validity
	overallValid := overallScore >= 0.7 && totalErrors == 0

	return &ComprehensiveValidationResult{
		Result:            result,
		OverallValid:      overallValid,
		OverallScore:      overallScore,
		ValidationResults: validationResults,
		TotalErrors:       totalErrors,
		TotalWarnings:     totalWarnings,
		ValidationTime:    time.Since(start),
	}, nil
}

// ValidateResults validates multiple search results
func (srv *SearchResultValidator) ValidateResults(ctx context.Context, results []WebSearchResult) ([]*ComprehensiveValidationResult, error) {
	var validatedResults []*ComprehensiveValidationResult

	for _, result := range results {
		validationResult, err := srv.ValidateResult(ctx, &result)
		if err != nil {
			// Log error but continue with other results
			fmt.Printf("Failed to validate result %s: %v\n", result.Title, err)
			continue
		}
		validatedResults = append(validatedResults, validationResult)
	}

	return validatedResults, nil
}

// UpdateConfig updates the validation configuration
func (srv *SearchResultValidator) UpdateConfig(config ValidationConfig) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.config = config
	srv.httpClient.Timeout = config.RequestTimeout
}

// GetConfig returns the current configuration
func (srv *SearchResultValidator) GetConfig() ValidationConfig {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	return srv.config
}

// GetStats returns statistics about validation
func (srv *SearchResultValidator) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_validators": len(srv.validators),
		"config":           srv.config,
	}
}

// URLValidator validates URL accessibility and structure
type URLValidator struct {
	config     ValidationConfig
	httpClient *http.Client
}

func (uv *URLValidator) Validate(ctx context.Context, result *WebSearchResult) (*ValidationResult, error) {
	start := time.Now()

	var errors []string
	var warnings []string
	var metadata = make(map[string]interface{})

	// Parse URL
	parsedURL, err := url.Parse(result.URL)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Invalid URL format: %v", err))
		return &ValidationResult{
			ValidatorName:  uv.GetName(),
			IsValid:        false,
			Score:          0.0,
			Errors:         errors,
			Warnings:       warnings,
			Metadata:       metadata,
			ValidationTime: time.Since(start),
		}, nil
	}

	// Check URL scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		errors = append(errors, "Invalid URL scheme")
	}

	// Check TLD
	if !uv.isValidTLD(parsedURL.Hostname()) {
		warnings = append(warnings, "Suspicious TLD")
	}

	// Check if URL is accessible
	if uv.config.EnableURLValidation {
		statusCode, err := uv.checkURLAccessibility(ctx, result.URL)
		if err != nil {
			errors = append(errors, fmt.Sprintf("URL not accessible: %v", err))
		} else {
			metadata["status_code"] = statusCode
			if !uv.isAllowedStatusCode(statusCode) {
				errors = append(errors, fmt.Sprintf("Unexpected status code: %d", statusCode))
			}
		}
	}

	// Calculate score
	score := 1.0
	if len(errors) > 0 {
		score = 0.0
	} else if len(warnings) > 0 {
		score = 0.7
	}

	return &ValidationResult{
		ValidatorName:  uv.GetName(),
		IsValid:        len(errors) == 0,
		Score:          score,
		Errors:         errors,
		Warnings:       warnings,
		Metadata:       metadata,
		ValidationTime: time.Since(start),
	}, nil
}

func (uv *URLValidator) checkURLAccessibility(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", uv.config.UserAgent)

	resp, err := uv.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (uv *URLValidator) isValidTLD(hostname string) bool {
	for _, tld := range uv.config.ValidTLDs {
		if strings.HasSuffix(hostname, tld) {
			return true
		}
	}
	return false
}

func (uv *URLValidator) isAllowedStatusCode(statusCode int) bool {
	for _, allowed := range uv.config.AllowedStatusCodes {
		if statusCode == allowed {
			return true
		}
	}
	return false
}

func (uv *URLValidator) GetName() string {
	return "URLValidator"
}

func (uv *URLValidator) GetDescription() string {
	return "Validates URL accessibility and structure"
}

func (uv *URLValidator) GetPriority() int {
	return 1
}

// ContentValidator validates content quality and relevance
type ContentValidator struct {
	config ValidationConfig
}

func (cv *ContentValidator) Validate(ctx context.Context, result *WebSearchResult) (*ValidationResult, error) {
	start := time.Now()

	var errors []string
	var warnings []string
	var metadata = make(map[string]interface{})

	content := result.Title + " " + result.Description
	contentLength := len(content)

	metadata["content_length"] = contentLength

	// Check content length
	if contentLength < cv.config.MinContentLength {
		errors = append(errors, fmt.Sprintf("Content too short: %d characters", contentLength))
	}

	if contentLength > cv.config.MaxContentLength {
		warnings = append(warnings, fmt.Sprintf("Content very long: %d characters", contentLength))
	}

	// Check for blocked keywords
	lowerContent := strings.ToLower(content)
	for _, keyword := range cv.config.BlockedContentKeywords {
		if strings.Contains(lowerContent, keyword) {
			errors = append(errors, fmt.Sprintf("Contains blocked keyword: %s", keyword))
		}
	}

	// Check for required keywords (if any)
	if len(cv.config.RequiredContentKeywords) > 0 {
		foundKeywords := 0
		for _, keyword := range cv.config.RequiredContentKeywords {
			if strings.Contains(lowerContent, strings.ToLower(keyword)) {
				foundKeywords++
			}
		}
		if foundKeywords == 0 {
			warnings = append(warnings, "No required keywords found")
		}
		metadata["found_keywords"] = foundKeywords
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		`\$\d+`,             // Dollar amounts
		`\d{3}-\d{3}-\d{4}`, // Phone numbers
		`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, // Email addresses
	}

	for _, pattern := range suspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, content)
		if matched {
			warnings = append(warnings, "Contains suspicious pattern")
			break
		}
	}

	// Calculate score
	score := 1.0
	if len(errors) > 0 {
		score = 0.0
	} else if len(warnings) > 0 {
		score = 0.8
	}

	return &ValidationResult{
		ValidatorName:  cv.GetName(),
		IsValid:        len(errors) == 0,
		Score:          score,
		Errors:         errors,
		Warnings:       warnings,
		Metadata:       metadata,
		ValidationTime: time.Since(start),
	}, nil
}

func (cv *ContentValidator) GetName() string {
	return "ContentValidator"
}

func (cv *ContentValidator) GetDescription() string {
	return "Validates content quality and relevance"
}

func (cv *ContentValidator) GetPriority() int {
	return 2
}

// DomainValidator validates domain reputation and safety
type DomainValidator struct {
	config ValidationConfig
}

func (dv *DomainValidator) Validate(ctx context.Context, result *WebSearchResult) (*ValidationResult, error) {
	start := time.Now()

	var errors []string
	var warnings []string
	var metadata = make(map[string]interface{})

	parsedURL, err := url.Parse(result.URL)
	if err != nil {
		errors = append(errors, "Invalid URL format")
		return &ValidationResult{
			ValidatorName:  dv.GetName(),
			IsValid:        false,
			Score:          0.0,
			Errors:         errors,
			Warnings:       warnings,
			Metadata:       metadata,
			ValidationTime: time.Since(start),
		}, nil
	}

	domain := parsedURL.Hostname()
	metadata["domain"] = domain

	// Check for blocked domains
	for _, blockedDomain := range dv.config.BlockedDomains {
		if strings.Contains(domain, blockedDomain) {
			errors = append(errors, fmt.Sprintf("Domain is blocked: %s", blockedDomain))
		}
	}

	// Check for suspicious domain patterns
	suspiciousPatterns := []string{
		`\d+\.\d+\.\d+\.\d+`,     // IP addresses
		`[a-z]{1,3}\.[a-z]{1,3}`, // Short domains
	}

	for _, pattern := range suspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, domain)
		if matched {
			warnings = append(warnings, "Suspicious domain pattern")
			break
		}
	}

	// Check domain length
	if len(domain) < 5 {
		warnings = append(warnings, "Very short domain name")
	}

	// Calculate score
	score := 1.0
	if len(errors) > 0 {
		score = 0.0
	} else if len(warnings) > 0 {
		score = 0.7
	}

	return &ValidationResult{
		ValidatorName:  dv.GetName(),
		IsValid:        len(errors) == 0,
		Score:          score,
		Errors:         errors,
		Warnings:       warnings,
		Metadata:       metadata,
		ValidationTime: time.Since(start),
	}, nil
}

func (dv *DomainValidator) GetName() string {
	return "DomainValidator"
}

func (dv *DomainValidator) GetDescription() string {
	return "Validates domain reputation and safety"
}

func (dv *DomainValidator) GetPriority() int {
	return 3
}

// AccessibilityValidator checks if the URL is accessible and responsive
type AccessibilityValidator struct {
	config     ValidationConfig
	httpClient *http.Client
}

func (av *AccessibilityValidator) Validate(ctx context.Context, result *WebSearchResult) (*ValidationResult, error) {
	start := time.Now()

	var errors []string
	var warnings []string
	var metadata = make(map[string]interface{})

	if !av.config.EnableAccessibilityCheck {
		return &ValidationResult{
			ValidatorName:  av.GetName(),
			IsValid:        true,
			Score:          1.0,
			Errors:         errors,
			Warnings:       warnings,
			Metadata:       metadata,
			ValidationTime: time.Since(start),
		}, nil
	}

	// Check response time
	req, err := http.NewRequestWithContext(ctx, "GET", result.URL, nil)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to create request: %v", err))
		return &ValidationResult{
			ValidatorName:  av.GetName(),
			IsValid:        false,
			Score:          0.0,
			Errors:         errors,
			Warnings:       warnings,
			Metadata:       metadata,
			ValidationTime: time.Since(start),
		}, nil
	}

	req.Header.Set("User-Agent", av.config.UserAgent)

	responseStart := time.Now()
	resp, err := av.httpClient.Do(req)
	responseTime := time.Since(responseStart)

	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to access URL: %v", err))
	} else {
		defer resp.Body.Close()

		metadata["response_time"] = responseTime.String()
		metadata["status_code"] = resp.StatusCode
		metadata["content_type"] = resp.Header.Get("Content-Type")

		// Check response time
		if responseTime > time.Second*5 {
			warnings = append(warnings, "Slow response time")
		}

		// Check content type
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/json") {
			warnings = append(warnings, "Unexpected content type")
		}
	}

	// Calculate score
	score := 1.0
	if len(errors) > 0 {
		score = 0.0
	} else if len(warnings) > 0 {
		score = 0.8
	}

	return &ValidationResult{
		ValidatorName:  av.GetName(),
		IsValid:        len(errors) == 0,
		Score:          score,
		Errors:         errors,
		Warnings:       warnings,
		Metadata:       metadata,
		ValidationTime: time.Since(start),
	}, nil
}

func (av *AccessibilityValidator) GetName() string {
	return "AccessibilityValidator"
}

func (av *AccessibilityValidator) GetDescription() string {
	return "Checks if the URL is accessible and responsive"
}

func (av *AccessibilityValidator) GetPriority() int {
	return 4
}

// SecurityValidator checks for security-related issues
type SecurityValidator struct {
	config ValidationConfig
}

func (sv *SecurityValidator) Validate(ctx context.Context, result *WebSearchResult) (*ValidationResult, error) {
	start := time.Now()

	var errors []string
	var warnings []string
	var metadata = make(map[string]interface{})

	if !sv.config.EnableSecurityCheck {
		return &ValidationResult{
			ValidatorName:  sv.GetName(),
			IsValid:        true,
			Score:          1.0,
			Errors:         errors,
			Warnings:       warnings,
			Metadata:       metadata,
			ValidationTime: time.Since(start),
		}, nil
	}

	// Check for HTTPS
	if !strings.HasPrefix(result.URL, "https://") {
		warnings = append(warnings, "Not using HTTPS")
	}

	// Check for suspicious URL patterns
	suspiciousPatterns := []string{
		`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`, // IP addresses
		`[a-zA-Z0-9]{32,}`,                 // Long hashes
		`[a-zA-Z0-9]{8,}\.[a-zA-Z0-9]{8,}`, // Suspicious subdomains
	}

	for _, pattern := range suspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, result.URL)
		if matched {
			warnings = append(warnings, "Suspicious URL pattern")
			break
		}
	}

	// Check for common security keywords in content
	securityKeywords := []string{"password", "login", "signin", "admin", "root"}
	content := strings.ToLower(result.Title + " " + result.Description)

	for _, keyword := range securityKeywords {
		if strings.Contains(content, keyword) {
			warnings = append(warnings, "Contains security-sensitive keyword")
			break
		}
	}

	// Calculate score
	score := 1.0
	if len(errors) > 0 {
		score = 0.0
	} else if len(warnings) > 0 {
		score = 0.8
	}

	return &ValidationResult{
		ValidatorName:  sv.GetName(),
		IsValid:        len(errors) == 0,
		Score:          score,
		Errors:         errors,
		Warnings:       warnings,
		Metadata:       metadata,
		ValidationTime: time.Since(start),
	}, nil
}

func (sv *SecurityValidator) GetName() string {
	return "SecurityValidator"
}

func (sv *SecurityValidator) GetDescription() string {
	return "Checks for security-related issues"
}

func (sv *SecurityValidator) GetPriority() int {
	return 5
}

// FreshnessValidator checks content freshness (placeholder for future implementation)
type FreshnessValidator struct {
	config ValidationConfig
}

func (fv *FreshnessValidator) Validate(ctx context.Context, result *WebSearchResult) (*ValidationResult, error) {
	start := time.Now()

	// This is a placeholder for future implementation
	// In a real implementation, this would check:
	// - Last modified date from HTTP headers
	// - Content freshness indicators
	// - Archive.org data
	// - Social media mentions

	return &ValidationResult{
		ValidatorName:  fv.GetName(),
		IsValid:        true,
		Score:          0.5, // Neutral score for placeholder
		Errors:         []string{},
		Warnings:       []string{"Freshness validation not implemented"},
		Metadata:       map[string]interface{}{},
		ValidationTime: time.Since(start),
	}, nil
}

func (fv *FreshnessValidator) GetName() string {
	return "FreshnessValidator"
}

func (fv *FreshnessValidator) GetDescription() string {
	return "Checks content freshness and recency"
}

func (fv *FreshnessValidator) GetPriority() int {
	return 6
}
