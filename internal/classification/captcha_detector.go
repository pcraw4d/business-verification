package classification

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// CAPTCHADetector detects CAPTCHA challenges in HTTP responses
type CAPTCHADetector struct {
	enabled bool
	patterns []string
}

// CAPTCHAType represents the type of CAPTCHA detected
type CAPTCHAType string

const (
	CAPTCHATypeNone        CAPTCHAType = "none"
	CAPTCHATypeReCAPTCHA   CAPTCHAType = "recaptcha"
	CAPTCHATypeHCaptcha    CAPTCHAType = "hcaptcha"
	CAPTCHATypeCloudflare  CAPTCHAType = "cloudflare"
	CAPTCHATypeGeneric     CAPTCHAType = "generic"
	CAPTCHATypeTurnstile   CAPTCHAType = "turnstile"
)

// CAPTCHAResult contains the result of CAPTCHA detection
type CAPTCHAResult struct {
	Detected bool
	Type     CAPTCHAType
	Message  string
}

// NewCAPTCHADetector creates a new CAPTCHA detector
func NewCAPTCHADetector() *CAPTCHADetector {
	enabled := os.Getenv("SCRAPING_CAPTCHA_DETECTION_ENABLED")
	enabledBool := true // Default to enabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	patterns := []string{
		// reCAPTCHA patterns
		"recaptcha",
		"g-recaptcha",
		"google.com/recaptcha",
		"recaptcha/api.js",
		"recaptcha/enterprise",
		
		// hCaptcha patterns
		"hcaptcha",
		"hcaptcha.com",
		"hcaptcha/api.js",
		
		// Cloudflare patterns
		"cloudflare",
		"cf-browser-verification",
		"cf-challenge",
		"checking your browser",
		"just a moment",
		"ray id",
		"cf-ray",
		
		// Turnstile (Cloudflare's new CAPTCHA)
		"turnstile",
		"challenges.cloudflare.com",
		
		// Generic CAPTCHA patterns
		"captcha",
		"challenge",
		"verify you are human",
		"prove you are not a robot",
		"human verification",
		"security check",
		"verification required",
		"access denied",
		"blocked",
	}

	return &CAPTCHADetector{
		enabled:  enabledBool,
		patterns: patterns,
	}
}

// DetectCAPTCHA detects CAPTCHA in an HTTP response
// Returns a CAPTCHAResult indicating if CAPTCHA was detected and its type
func (cd *CAPTCHADetector) DetectCAPTCHA(resp *http.Response, body []byte) CAPTCHAResult {
	if !cd.enabled {
		return CAPTCHAResult{Detected: false, Type: CAPTCHATypeNone}
	}

	// Check response headers for CAPTCHA indicators
	if cd.checkHeaders(resp) {
		return cd.identifyCAPTCHAType(resp, body)
	}

	// Check response body for CAPTCHA patterns
	if len(body) > 0 {
		bodyLower := strings.ToLower(string(body))
		return cd.checkBody(bodyLower, resp)
	}

	// If body wasn't provided, try to read it (with size limit)
	if resp.Body != nil {
		limitedReader := io.LimitReader(resp.Body, 50*1024) // Read up to 50KB
		bodyBytes, err := io.ReadAll(limitedReader)
		if err == nil && len(bodyBytes) > 0 {
			bodyLower := strings.ToLower(string(bodyBytes))
			return cd.checkBody(bodyLower, resp)
		}
	}

	return CAPTCHAResult{Detected: false, Type: CAPTCHATypeNone}
}

// checkHeaders checks response headers for CAPTCHA indicators
func (cd *CAPTCHADetector) checkHeaders(resp *http.Response) bool {
	// Check for Cloudflare challenge headers
	if resp.Header.Get("cf-challenge") != "" {
		return true
	}
	if resp.Header.Get("cf-ray") != "" && resp.StatusCode == 403 {
		return true
	}

	// Check for CAPTCHA-related headers
	for key, values := range resp.Header {
		keyLower := strings.ToLower(key)
		for _, value := range values {
			valueLower := strings.ToLower(value)
			if strings.Contains(keyLower, "captcha") || strings.Contains(valueLower, "captcha") {
				return true
			}
			if strings.Contains(keyLower, "challenge") || strings.Contains(valueLower, "challenge") {
				return true
			}
		}
	}

	return false
}

// checkBody checks the response body for CAPTCHA patterns
func (cd *CAPTCHADetector) checkBody(bodyLower string, resp *http.Response) CAPTCHAResult {
	// Check for specific CAPTCHA types first (more specific patterns)
	
	// reCAPTCHA
	if strings.Contains(bodyLower, "recaptcha") || 
	   strings.Contains(bodyLower, "g-recaptcha") ||
	   strings.Contains(bodyLower, "google.com/recaptcha") {
		return CAPTCHAResult{
			Detected: true,
			Type:     CAPTCHATypeReCAPTCHA,
			Message:  "reCAPTCHA detected",
		}
	}

	// hCaptcha
	if strings.Contains(bodyLower, "hcaptcha") || 
	   strings.Contains(bodyLower, "hcaptcha.com") {
		return CAPTCHAResult{
			Detected: true,
			Type:     CAPTCHATypeHCaptcha,
			Message:  "hCaptcha detected",
		}
	}

	// Cloudflare
	if strings.Contains(bodyLower, "cloudflare") ||
	   strings.Contains(bodyLower, "cf-browser-verification") ||
	   strings.Contains(bodyLower, "checking your browser") ||
	   strings.Contains(bodyLower, "just a moment") ||
	   strings.Contains(bodyLower, "cf-challenge") {
		return CAPTCHAResult{
			Detected: true,
			Type:     CAPTCHATypeCloudflare,
			Message:  "Cloudflare challenge detected",
		}
	}

	// Turnstile (Cloudflare's new CAPTCHA)
	if strings.Contains(bodyLower, "turnstile") ||
	   strings.Contains(bodyLower, "challenges.cloudflare.com") {
		return CAPTCHAResult{
			Detected: true,
			Type:     CAPTCHATypeTurnstile,
			Message:  "Cloudflare Turnstile detected",
		}
	}

	// Generic CAPTCHA patterns
	for _, pattern := range cd.patterns {
		if strings.Contains(bodyLower, pattern) {
			// Check if it's a generic pattern (not already matched)
			if pattern == "captcha" || pattern == "challenge" ||
			   pattern == "verify you are human" ||
			   pattern == "prove you are not a robot" {
				return CAPTCHAResult{
					Detected: true,
					Type:     CAPTCHATypeGeneric,
					Message:  "Generic CAPTCHA detected",
				}
			}
		}
	}

	// Check status code + content combination
	if resp.StatusCode == 403 {
		// 403 with certain content might indicate CAPTCHA
		if strings.Contains(bodyLower, "access denied") ||
		   strings.Contains(bodyLower, "blocked") ||
		   strings.Contains(bodyLower, "forbidden") {
			// Could be CAPTCHA or just blocking - check for CAPTCHA keywords
			if strings.Contains(bodyLower, "captcha") || 
			   strings.Contains(bodyLower, "challenge") ||
			   strings.Contains(bodyLower, "verify") {
				return CAPTCHAResult{
					Detected: true,
					Type:     CAPTCHATypeGeneric,
					Message:  "Possible CAPTCHA (403 with verification content)",
				}
			}
		}
	}

	return CAPTCHAResult{Detected: false, Type: CAPTCHATypeNone}
}

// identifyCAPTCHAType identifies the specific type of CAPTCHA based on headers and body
func (cd *CAPTCHADetector) identifyCAPTCHAType(resp *http.Response, body []byte) CAPTCHAResult {
	// Check headers first
	if resp.Header.Get("cf-challenge") != "" || resp.Header.Get("cf-ray") != "" {
		return CAPTCHAResult{
			Detected: true,
			Type:     CAPTCHATypeCloudflare,
			Message:  "Cloudflare challenge detected via headers",
		}
	}

	// Check body if available
	if len(body) > 0 {
		bodyLower := strings.ToLower(string(body))
		return cd.checkBody(bodyLower, resp)
	}

	return CAPTCHAResult{
		Detected: true,
		Type:     CAPTCHATypeGeneric,
		Message:  "CAPTCHA detected via headers",
	}
}

// DetectCAPTCHA is a convenience function using a default detector
func DetectCAPTCHA(resp *http.Response, body []byte) CAPTCHAResult {
	detector := NewCAPTCHADetector()
	return detector.DetectCAPTCHA(resp, body)
}

// IsEnabled checks if CAPTCHA detection is enabled
func (cd *CAPTCHADetector) IsEnabled() bool {
	return cd.enabled
}

// SetEnabled enables or disables CAPTCHA detection
func (cd *CAPTCHADetector) SetEnabled(enabled bool) {
	cd.enabled = enabled
}

