package sanitizer

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Sanitizer provides comprehensive input sanitization functionality
type Sanitizer struct {
	// Compiled regex patterns for performance
	htmlTagPattern       *regexp.Regexp
	sqlInjectionPattern  *regexp.Regexp
	xssPattern           *regexp.Regexp
	pathTraversalPattern *regexp.Regexp
	scriptPattern        *regexp.Regexp
	urlPattern           *regexp.Regexp
	emailPattern         *regexp.Regexp
	phonePattern         *regexp.Regexp
	uuidPattern          *regexp.Regexp
}

// SanitizationConfig holds configuration for sanitization
type SanitizationConfig struct {
	RemoveHTMLTags       bool `json:"remove_html_tags"`
	RemoveScripts        bool `json:"remove_scripts"`
	NormalizeUnicode     bool `json:"normalize_unicode"`
	TrimWhitespace       bool `json:"trim_whitespace"`
	RemoveNullBytes      bool `json:"remove_null_bytes"`
	NormalizeLineEndings bool `json:"normalize_line_endings"`
	MaxLength            int  `json:"max_length"`
	AllowHTML            bool `json:"allow_html"`
	StrictMode           bool `json:"strict_mode"`
}

// NewSanitizer creates a new sanitizer with default configuration
func NewSanitizer() *Sanitizer {
	return NewSanitizerWithConfig(&SanitizationConfig{
		RemoveHTMLTags:       true,
		RemoveScripts:        true,
		NormalizeUnicode:     true,
		TrimWhitespace:       true,
		RemoveNullBytes:      true,
		NormalizeLineEndings: true,
		MaxLength:            10000,
		AllowHTML:            false,
		StrictMode:           false,
	})
}

// NewSanitizerWithConfig creates a new sanitizer with custom configuration
func NewSanitizerWithConfig(config *SanitizationConfig) *Sanitizer {
	// Compile regex patterns for performance
	htmlTagPattern := regexp.MustCompile(`<[^>]*>`)
	sqlInjectionPattern := regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`)
	xssPattern := regexp.MustCompile(`(?i)(javascript:|vbscript:|onload=|onerror=|onclick=|<script)`)
	pathTraversalPattern := regexp.MustCompile(`\.\./|\.\.\\|/etc/|/proc/|/sys/|/dev/`)
	scriptPattern := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phonePattern := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	return &Sanitizer{
		htmlTagPattern:       htmlTagPattern,
		sqlInjectionPattern:  sqlInjectionPattern,
		xssPattern:           xssPattern,
		pathTraversalPattern: pathTraversalPattern,
		scriptPattern:        scriptPattern,
		urlPattern:           urlPattern,
		emailPattern:         emailPattern,
		phonePattern:         phonePattern,
		uuidPattern:          uuidPattern,
	}
}

// SanitizeString sanitizes a string input
func (s *Sanitizer) SanitizeString(input string, config *SanitizationConfig) string {
	if input == "" {
		return input
	}

	if config == nil {
		config = &SanitizationConfig{
			RemoveHTMLTags:       true,
			RemoveScripts:        true,
			NormalizeUnicode:     true,
			TrimWhitespace:       true,
			RemoveNullBytes:      true,
			NormalizeLineEndings: true,
			MaxLength:            10000,
			AllowHTML:            false,
			StrictMode:           false,
		}
	}

	result := input

	// Remove null bytes
	if config.RemoveNullBytes {
		result = strings.ReplaceAll(result, "\x00", "")
	}

	// Normalize Unicode
	if config.NormalizeUnicode {
		result = norm.NFC.String(result)
	}

	// Remove scripts
	if config.RemoveScripts {
		result = s.removeScripts(result)
	}

	// Remove HTML tags
	if config.RemoveHTMLTags && !config.AllowHTML {
		result = s.removeHTMLTags(result)
	}

	// Normalize line endings
	if config.NormalizeLineEndings {
		result = strings.ReplaceAll(result, "\r\n", "\n")
		result = strings.ReplaceAll(result, "\r", "\n")
	}

	// Trim whitespace
	if config.TrimWhitespace {
		result = strings.TrimSpace(result)
	}

	// Apply length limit
	if config.MaxLength > 0 && len(result) > config.MaxLength {
		result = result[:config.MaxLength]
	}

	// Strict mode: additional sanitization
	if config.StrictMode {
		result = s.strictSanitize(result)
	}

	return result
}

// SanitizeEmail sanitizes and validates an email address
func (s *Sanitizer) SanitizeEmail(email string) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	// Basic sanitization
	sanitized := s.SanitizeString(email, &SanitizationConfig{
		RemoveHTMLTags:       true,
		RemoveScripts:        true,
		NormalizeUnicode:     true,
		TrimWhitespace:       true,
		RemoveNullBytes:      true,
		NormalizeLineEndings: false,
		MaxLength:            254, // RFC 5321 limit
		AllowHTML:            false,
		StrictMode:           true,
	})

	// Validate email format
	if !s.emailPattern.MatchString(sanitized) {
		return "", fmt.Errorf("invalid email format: %s", sanitized)
	}

	// Check for common email issues
	if strings.Contains(sanitized, "..") {
		return "", fmt.Errorf("email contains consecutive dots")
	}

	if strings.HasPrefix(sanitized, ".") || strings.HasSuffix(sanitized, ".") {
		return "", fmt.Errorf("email cannot start or end with a dot")
	}

	return sanitized, nil
}

// SanitizeURL sanitizes and validates a URL
func (s *Sanitizer) SanitizeURL(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	// Basic sanitization
	sanitized := s.SanitizeString(url, &SanitizationConfig{
		RemoveHTMLTags:       true,
		RemoveScripts:        true,
		NormalizeUnicode:     true,
		TrimWhitespace:       true,
		RemoveNullBytes:      true,
		NormalizeLineEndings: false,
		MaxLength:            2048, // Common URL length limit
		AllowHTML:            false,
		StrictMode:           true,
	})

	// Validate URL format
	if !s.urlPattern.MatchString(sanitized) {
		return "", fmt.Errorf("invalid URL format: %s", sanitized)
	}

	// Check for dangerous protocols
	lowerURL := strings.ToLower(sanitized)
	if strings.HasPrefix(lowerURL, "javascript:") ||
		strings.HasPrefix(lowerURL, "vbscript:") ||
		strings.HasPrefix(lowerURL, "data:") {
		return "", fmt.Errorf("dangerous URL protocol: %s", sanitized)
	}

	return sanitized, nil
}

// SanitizePhone sanitizes and validates a phone number
func (s *Sanitizer) SanitizePhone(phone string) (string, error) {
	if phone == "" {
		return "", fmt.Errorf("phone number cannot be empty")
	}

	// Basic sanitization
	sanitized := s.SanitizeString(phone, &SanitizationConfig{
		RemoveHTMLTags:       true,
		RemoveScripts:        true,
		NormalizeUnicode:     true,
		TrimWhitespace:       true,
		RemoveNullBytes:      true,
		NormalizeLineEndings: false,
		MaxLength:            20,
		AllowHTML:            false,
		StrictMode:           true,
	})

	// Validate phone format (E.164)
	if !s.phonePattern.MatchString(sanitized) {
		return "", fmt.Errorf("invalid phone number format (E.164 required): %s", sanitized)
	}

	return sanitized, nil
}

// SanitizeUUID sanitizes and validates a UUID
func (s *Sanitizer) SanitizeUUID(uuid string) (string, error) {
	if uuid == "" {
		return "", fmt.Errorf("UUID cannot be empty")
	}

	// Basic sanitization
	sanitized := s.SanitizeString(uuid, &SanitizationConfig{
		RemoveHTMLTags:       true,
		RemoveScripts:        true,
		NormalizeUnicode:     true,
		TrimWhitespace:       true,
		RemoveNullBytes:      true,
		NormalizeLineEndings: false,
		MaxLength:            36,
		AllowHTML:            false,
		StrictMode:           true,
	})

	// Validate UUID format
	if !s.uuidPattern.MatchString(strings.ToLower(sanitized)) {
		return "", fmt.Errorf("invalid UUID format: %s", sanitized)
	}

	return strings.ToLower(sanitized), nil
}

// SanitizeHTML sanitizes HTML content while preserving safe tags
func (s *Sanitizer) SanitizeHTML(html string) string {
	if html == "" {
		return html
	}

	// Remove scripts
	html = s.removeScripts(html)

	// Remove dangerous attributes
	html = s.removeDangerousAttributes(html)

	// Remove dangerous tags
	html = s.removeDangerousTags(html)

	// Normalize Unicode
	html = norm.NFC.String(html)

	// Remove null bytes
	html = strings.ReplaceAll(html, "\x00", "")

	return html
}

// SanitizeFilename sanitizes a filename for safe file operations
func (s *Sanitizer) SanitizeFilename(filename string) string {
	if filename == "" {
		return filename
	}

	// Remove path traversal attempts
	filename = s.pathTraversalPattern.ReplaceAllString(filename, "")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Remove dangerous characters
	dangerousChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = dangerousChars.ReplaceAllString(filename, "")

	// Trim whitespace and dots
	filename = strings.TrimSpace(filename)
	filename = strings.Trim(filename, ".")

	// Limit length
	if len(filename) > 255 {
		filename = filename[:255]
	}

	return filename
}

// SanitizeSQL sanitizes input for SQL queries (use parameterized queries instead)
func (s *Sanitizer) SanitizeSQL(input string) string {
	if input == "" {
		return input
	}

	// Remove SQL injection patterns
	result := s.sqlInjectionPattern.ReplaceAllString(input, "")

	// Remove comments
	commentPattern := regexp.MustCompile(`--.*$|/\*.*?\*/`)
	result = commentPattern.ReplaceAllString(result, "")

	// Remove multiple spaces
	spacePattern := regexp.MustCompile(`\s+`)
	result = spacePattern.ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}

// GenerateSecureToken generates a cryptographically secure random token
func (s *Sanitizer) GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be positive")
	}

	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64
	token := base64.URLEncoding.EncodeToString(bytes)

	// Trim to exact length
	if len(token) > length {
		token = token[:length]
	}

	return token, nil
}

// ValidateInput validates input without sanitizing
func (s *Sanitizer) ValidateInput(input string, inputType string) error {
	if input == "" {
		return fmt.Errorf("input cannot be empty")
	}

	switch inputType {
	case "email":
		if !s.emailPattern.MatchString(input) {
			return fmt.Errorf("invalid email format")
		}
	case "url":
		if !s.urlPattern.MatchString(input) {
			return fmt.Errorf("invalid URL format")
		}
	case "phone":
		if !s.phonePattern.MatchString(input) {
			return fmt.Errorf("invalid phone number format")
		}
	case "uuid":
		if !s.uuidPattern.MatchString(strings.ToLower(input)) {
			return fmt.Errorf("invalid UUID format")
		}
	default:
		return fmt.Errorf("unknown input type: %s", inputType)
	}

	return nil
}

// removeScripts removes script tags and content
func (s *Sanitizer) removeScripts(input string) string {
	// Remove script tags
	result := s.scriptPattern.ReplaceAllString(input, "")

	// Remove script attributes
	scriptAttrPattern := regexp.MustCompile(`(?i)on\w+\s*=\s*["'][^"']*["']`)
	result = scriptAttrPattern.ReplaceAllString(result, "")

	// Remove javascript: URLs
	jsUrlPattern := regexp.MustCompile(`(?i)javascript:[^"'\s]*`)
	result = jsUrlPattern.ReplaceAllString(result, "")

	return result
}

// removeHTMLTags removes HTML tags
func (s *Sanitizer) removeHTMLTags(input string) string {
	return s.htmlTagPattern.ReplaceAllString(input, "")
}

// removeDangerousAttributes removes dangerous HTML attributes
func (s *Sanitizer) removeDangerousAttributes(input string) string {
	dangerousAttrPattern := regexp.MustCompile(`(?i)\s+(on\w+|javascript:|vbscript:)\s*=\s*["'][^"']*["']`)
	return dangerousAttrPattern.ReplaceAllString(input, "")
}

// removeDangerousTags removes dangerous HTML tags
func (s *Sanitizer) removeDangerousTags(input string) string {
	dangerousTagPattern := regexp.MustCompile(`(?i)</?(script|object|embed|applet|form|input|textarea|select|button|iframe|frame|frameset|noframes|noscript)[^>]*>`)
	return dangerousTagPattern.ReplaceAllString(input, "")
}

// strictSanitize applies strict sanitization rules
func (s *Sanitizer) strictSanitize(input string) string {
	// Remove all non-printable characters except newlines and tabs
	var result bytes.Buffer
	for _, r := range input {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			result.WriteRune(r)
		}
	}

	// Remove control characters
	controlPattern := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	resultStr := controlPattern.ReplaceAllString(result.String(), "")

	// Normalize whitespace
	whitespacePattern := regexp.MustCompile(`\s+`)
	resultStr = whitespacePattern.ReplaceAllString(resultStr, " ")

	return strings.TrimSpace(resultStr)
}

// IsValidUTF8 checks if a string contains valid UTF-8
func (s *Sanitizer) IsValidUTF8(input string) bool {
	return utf8.ValidString(input)
}

// NormalizeUnicode normalizes Unicode characters
func (s *Sanitizer) NormalizeUnicode(input string) string {
	return norm.NFC.String(input)
}

// EscapeHTML escapes HTML special characters
func (s *Sanitizer) EscapeHTML(input string) string {
	return html.EscapeString(input)
}

// UnescapeHTML unescapes HTML entities
func (s *Sanitizer) UnescapeHTML(input string) string {
	return html.UnescapeString(input)
}

// RemoveControlCharacters removes control characters from a string
func (s *Sanitizer) RemoveControlCharacters(input string) string {
	var result bytes.Buffer
	for _, r := range input {
		if !unicode.IsControl(r) || r == '\n' || r == '\t' || r == '\r' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// SanitizeJSON sanitizes JSON string input
func (s *Sanitizer) SanitizeJSON(input string) string {
	if input == "" {
		return input
	}

	// Remove null bytes
	result := strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newlines and tabs
	result = s.RemoveControlCharacters(result)

	// Normalize Unicode
	result = norm.NFC.String(result)

	// Trim whitespace
	result = strings.TrimSpace(result)

	return result
}

// SanitizeXML sanitizes XML string input
func (s *Sanitizer) SanitizeXML(input string) string {
	if input == "" {
		return input
	}

	// Remove null bytes
	result := strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newlines and tabs
	result = s.RemoveControlCharacters(result)

	// Remove dangerous XML entities
	dangerousEntityPattern := regexp.MustCompile(`&(?:[a-zA-Z0-9]+|#[0-9]+|#x[0-9a-fA-F]+);`)
	result = dangerousEntityPattern.ReplaceAllString(result, "")

	// Normalize Unicode
	result = norm.NFC.String(result)

	// Trim whitespace
	result = strings.TrimSpace(result)

	return result
}
