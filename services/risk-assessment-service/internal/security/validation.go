package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"
)

// SecurityValidator provides comprehensive security validation
type SecurityValidator struct {
	logger *zap.Logger
}

// ValidationResult represents the result of security validation
type ValidationResult struct {
	Valid          bool     `json:"valid"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	SanitizedInput string   `json:"sanitized_input,omitempty"`
	RiskLevel      string   `json:"risk_level,omitempty"`
	SecurityScore  float64  `json:"security_score,omitempty"`
	ValidationTime int64    `json:"validation_time_ms"`
}

// SecurityConfig holds configuration for security validation
type SecurityConfig struct {
	MaxInputLength    int           `json:"max_input_length"`
	MinInputLength    int           `json:"min_input_length"`
	AllowedCharacters string        `json:"allowed_characters"`
	BlockedPatterns   []string      `json:"blocked_patterns"`
	RateLimitWindow   time.Duration `json:"rate_limit_window"`
	MaxRequestsPerIP  int           `json:"max_requests_per_ip"`
	EnableContentScan bool          `json:"enable_content_scan"`
	EnableGeoBlocking bool          `json:"enable_geo_blocking"`
	BlockedCountries  []string      `json:"blocked_countries"`
	EnableHoneypot    bool          `json:"enable_honeypot"`
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(logger *zap.Logger) *SecurityValidator {
	return &SecurityValidator{
		logger: logger,
	}
}

// ValidateInput performs comprehensive security validation on input
func (sv *SecurityValidator) ValidateInput(ctx context.Context, input string, config SecurityConfig) *ValidationResult {
	startTime := time.Now()
	result := &ValidationResult{
		Valid: true,
	}

	// Basic length validation
	if len(input) < config.MinInputLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("input too short, minimum length is %d", config.MinInputLength))
	}

	if len(input) > config.MaxInputLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("input too long, maximum length is %d", config.MaxInputLength))
	}

	// Sanitize input
	sanitized := sv.SanitizeInput(input)
	result.SanitizedInput = sanitized

	// Check for blocked patterns
	riskScore := sv.checkBlockedPatterns(sanitized, config.BlockedPatterns)
	result.SecurityScore = riskScore

	// Determine risk level based on security score
	result.RiskLevel = sv.determineRiskLevel(riskScore)

	// Character validation
	if config.AllowedCharacters != "" {
		if !sv.validateAllowedCharacters(sanitized, config.AllowedCharacters) {
			result.Valid = false
			result.Errors = append(result.Errors, "input contains disallowed characters")
		}
	}

	// Content scanning for malicious patterns
	if config.EnableContentScan {
		contentRisk := sv.scanContent(sanitized)
		result.SecurityScore = (result.SecurityScore + contentRisk) / 2
		result.RiskLevel = sv.determineRiskLevel(result.SecurityScore)
	}

	// Add warnings for high-risk content
	if result.SecurityScore > 0.7 {
		result.Warnings = append(result.Warnings, "input contains potentially risky content")
	}

	result.ValidationTime = time.Since(startTime).Milliseconds()

	sv.logger.Info("security validation completed",
		zap.String("risk_level", result.RiskLevel),
		zap.Float64("security_score", result.SecurityScore),
		zap.Int64("validation_time_ms", result.ValidationTime))

	return result
}

// SanitizeInput removes potentially dangerous content from input
func (sv *SecurityValidator) SanitizeInput(input string) string {
	// Remove HTML tags
	htmlTagRegex := regexp.MustCompile(`<[a-zA-Z][^>]*>`)
	sanitized := htmlTagRegex.ReplaceAllString(input, "")

	// Remove script content
	scriptPatterns := []string{
		`javascript:`,
		`vbscript:`,
		`onload=`,
		`onerror=`,
		`onclick=`,
		`alert\(`,
		`document\.`,
		`window\.`,
		`eval\(`,
		`setTimeout\(`,
		`setInterval\(`,
	}

	for _, pattern := range scriptPatterns {
		regex := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(pattern))
		sanitized = regex.ReplaceAllString(sanitized, "")
	}

	// Remove SQL injection patterns
	sqlPatterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
		`(?i)(script|javascript|vbscript|onload|onerror|onclick)`,
		`['\";]`,
		`--`,
		`/\*`,
		`\*/`,
		`xp_`,
		`sp_`,
	}

	for _, pattern := range sqlPatterns {
		regex := regexp.MustCompile(pattern)
		sanitized = regex.ReplaceAllString(sanitized, "")
	}

	// Remove excessive whitespace
	whitespaceRegex := regexp.MustCompile(`\s+`)
	sanitized = whitespaceRegex.ReplaceAllString(sanitized, " ")

	// Trim leading/trailing whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// checkBlockedPatterns checks input against blocked patterns
func (sv *SecurityValidator) checkBlockedPatterns(input string, blockedPatterns []string) float64 {
	riskScore := 0.0
	matches := 0

	for _, pattern := range blockedPatterns {
		regex := regexp.MustCompile(`(?i)` + pattern)
		if regex.MatchString(input) {
			matches++
		}
	}

	if matches > 0 {
		riskScore = float64(matches) / float64(len(blockedPatterns))
	}

	return riskScore
}

// validateAllowedCharacters checks if input contains only allowed characters
func (sv *SecurityValidator) validateAllowedCharacters(input, allowedChars string) bool {
	allowedRegex := regexp.MustCompile(`^[` + regexp.QuoteMeta(allowedChars) + `]+$`)
	return allowedRegex.MatchString(input)
}

// scanContent performs deep content analysis for malicious patterns
func (sv *SecurityValidator) scanContent(input string) float64 {
	riskScore := 0.0

	// Check for suspicious patterns
	suspiciousPatterns := map[string]float64{
		`(?i)(password|passwd|pwd)`:   0.3,
		`(?i)(admin|administrator)`:   0.4,
		`(?i)(root|system)`:           0.5,
		`(?i)(backdoor|trojan|virus)`: 0.8,
		`(?i)(exploit|hack|crack)`:    0.7,
		`(?i)(injection|sql|xss)`:     0.9,
		`(?i)(phishing|scam|fraud)`:   0.8,
		`(?i)(malware|ransomware)`:    0.9,
		`(?i)(botnet|ddos|attack)`:    0.8,
		`(?i)(keylogger|spyware)`:     0.9,
	}

	for pattern, weight := range suspiciousPatterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(input) {
			riskScore += weight
		}
	}

	// Check for excessive special characters
	specialCharCount := 0
	for _, char := range input {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !unicode.IsSpace(char) {
			specialCharCount++
		}
	}

	if specialCharCount > len(input)/3 {
		riskScore += 0.3
	}

	// Check for repeated patterns (potential obfuscation)
	// Simple check for repeated 3+ character sequences
	if strings.Contains(input, "aaa") || strings.Contains(input, "bbb") || strings.Contains(input, "ccc") {
		riskScore += 0.2
	}

	// Normalize risk score to 0-1 range
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	return riskScore
}

// determineRiskLevel converts security score to risk level
func (sv *SecurityValidator) determineRiskLevel(score float64) string {
	switch {
	case score >= 0.8:
		return "critical"
	case score >= 0.6:
		return "high"
	case score >= 0.4:
		return "medium"
	case score >= 0.2:
		return "low"
	default:
		return "minimal"
	}
}

// ValidateURL performs security validation on URLs
func (sv *SecurityValidator) ValidateURL(inputURL string) *ValidationResult {
	result := &ValidationResult{
		Valid: true,
	}

	// Parse URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "invalid URL format")
		return result
	}

	// Check for suspicious schemes
	suspiciousSchemes := []string{"javascript", "vbscript", "data", "file"}
	for _, scheme := range suspiciousSchemes {
		if strings.ToLower(parsedURL.Scheme) == scheme {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("suspicious URL scheme: %s", scheme))
		}
	}

	// Check for suspicious domains
	suspiciousDomains := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"::1",
	}

	for _, domain := range suspiciousDomains {
		if strings.Contains(parsedURL.Host, domain) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("suspicious domain: %s", domain))
		}
	}

	// Check for IP addresses in host
	ipRegex := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)
	if ipRegex.MatchString(parsedURL.Host) {
		result.Warnings = append(result.Warnings, "URL contains IP address instead of domain name")
	}

	return result
}

// ValidateEmail performs security validation on email addresses
func (sv *SecurityValidator) ValidateEmail(email string) *ValidationResult {
	result := &ValidationResult{
		Valid: true,
	}

	// Basic email format validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		result.Valid = false
		result.Errors = append(result.Errors, "invalid email format")
		return result
	}

	// Check for suspicious email patterns
	suspiciousPatterns := []string{
		`(?i)(admin|administrator)`,
		`(?i)(test|demo|example)`,
		`(?i)(noreply|no-reply)`,
		`(?i)(spam|junk)`,
	}

	for _, pattern := range suspiciousPatterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(email) {
			result.Warnings = append(result.Warnings, "suspicious email pattern detected")
		}
	}

	// Check for disposable email domains
	disposableDomains := []string{
		"10minutemail.com",
		"tempmail.org",
		"guerrillamail.com",
		"mailinator.com",
		"yopmail.com",
	}

	emailParts := strings.Split(email, "@")
	if len(emailParts) == 2 {
		domain := strings.ToLower(emailParts[1])
		for _, disposable := range disposableDomains {
			if domain == disposable {
				result.Warnings = append(result.Warnings, "disposable email domain detected")
				break
			}
		}
	}

	return result
}

// GenerateInputHash generates a secure hash of the input for tracking
func (sv *SecurityValidator) GenerateInputHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// ValidateBusinessData performs comprehensive validation on business data
func (sv *SecurityValidator) ValidateBusinessData(ctx context.Context, data map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid: true,
	}

	// Validate business name
	if name, ok := data["business_name"].(string); ok {
		nameResult := sv.ValidateInput(ctx, name, SecurityConfig{
			MaxInputLength: 255,
			MinInputLength: 1,
			BlockedPatterns: []string{
				`(?i)(test|demo|example|fake)`,
				`(?i)(admin|administrator)`,
				`(?i)(system|root)`,
			},
		})
		if !nameResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, "invalid business name: "+strings.Join(nameResult.Errors, ", "))
		}
		if nameResult.SecurityScore > 0.5 {
			result.Warnings = append(result.Warnings, "business name contains suspicious content")
		}
	}

	// Validate business address
	if address, ok := data["business_address"].(string); ok {
		addressResult := sv.ValidateInput(ctx, address, SecurityConfig{
			MaxInputLength: 500,
			MinInputLength: 10,
			BlockedPatterns: []string{
				`(?i)(test|demo|example|fake)`,
				`(?i)(po box|p\.o\. box)`,
			},
		})
		if !addressResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, "invalid business address: "+strings.Join(addressResult.Errors, ", "))
		}
	}

	// Validate email if present
	if email, ok := data["email"].(string); ok && email != "" {
		emailResult := sv.ValidateEmail(email)
		if !emailResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, "invalid email: "+strings.Join(emailResult.Errors, ", "))
		}
		result.Warnings = append(result.Warnings, emailResult.Warnings...)
	}

	// Validate website if present
	if website, ok := data["website"].(string); ok && website != "" {
		urlResult := sv.ValidateURL(website)
		if !urlResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, "invalid website: "+strings.Join(urlResult.Errors, ", "))
		}
		result.Warnings = append(result.Warnings, urlResult.Warnings...)
	}

	return result
}
