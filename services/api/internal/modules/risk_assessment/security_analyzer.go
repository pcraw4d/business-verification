package risk_assessment

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SecurityAnalyzer provides security analysis capabilities
type SecurityAnalyzer struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
	client *http.Client
}

// SecurityAnalysisResult represents the result of security analysis
type SecurityAnalysisResult struct {
	WebsiteURL           string                   `json:"website_url"`
	AnalysisTimestamp    time.Time                `json:"analysis_timestamp"`
	OverallSecurityScore float64                  `json:"overall_security_score"`
	SecurityLevel        SecurityLevel            `json:"security_level"`
	SSLInfo              *SSLInfo                 `json:"ssl_info,omitempty"`
	SecurityHeaders      *SecurityHeaders         `json:"security_headers,omitempty"`
	TLSInfo              *TLSInfo                 `json:"tls_info,omitempty"`
	SecurityIssues       []SecurityIssue          `json:"security_issues"`
	Recommendations      []SecurityRecommendation `json:"recommendations"`
	ProcessingTime       time.Duration            `json:"processing_time"`
}

// SecurityLevel represents the security level
type SecurityLevel string

const (
	SecurityLevelExcellent SecurityLevel = "excellent"
	SecurityLevelGood      SecurityLevel = "good"
	SecurityLevelFair      SecurityLevel = "fair"
	SecurityLevelPoor      SecurityLevel = "poor"
	SecurityLevelCritical  SecurityLevel = "critical"
)

// SSLInfo contains SSL certificate information
type SSLInfo struct {
	Valid               bool      `json:"valid"`
	Issuer              string    `json:"issuer"`
	Subject             string    `json:"subject"`
	ValidFrom           time.Time `json:"valid_from"`
	ValidTo             time.Time `json:"valid_to"`
	DaysUntilExpiration int       `json:"days_until_expiration"`
	SignatureAlgorithm  string    `json:"signature_algorithm"`
	KeySize             int       `json:"key_size"`
	CertificateChain    []string  `json:"certificate_chain"`
	RevocationStatus    string    `json:"revocation_status"`
	TrustedByBrowser    bool      `json:"trusted_by_browser"`
	WildcardCertificate bool      `json:"wildcard_certificate"`
	ExtendedValidation  bool      `json:"extended_validation"`
	CertificateScore    float64   `json:"certificate_score"`
	Issues              []string  `json:"issues"`
}

// SecurityHeaders contains security header information
type SecurityHeaders struct {
	HSTS                *HSTSInfo                `json:"hsts,omitempty"`
	CSP                 *CSPInfo                 `json:"csp,omitempty"`
	XFrameOptions       *XFrameOptionsInfo       `json:"x_frame_options,omitempty"`
	XContentTypeOptions *XContentTypeOptionsInfo `json:"x_content_type_options,omitempty"`
	XSSProtection       *XSSProtectionInfo       `json:"xss_protection,omitempty"`
	ReferrerPolicy      *ReferrerPolicyInfo      `json:"referrer_policy,omitempty"`
	PermissionsPolicy   *PermissionsPolicyInfo   `json:"permissions_policy,omitempty"`
	HeadersScore        float64                  `json:"headers_score"`
	MissingHeaders      []string                 `json:"missing_headers"`
	WeakHeaders         []string                 `json:"weak_headers"`
}

// HSTSInfo contains HSTS header information
type HSTSInfo struct {
	Present           bool    `json:"present"`
	MaxAge            int     `json:"max_age"`
	IncludeSubDomains bool    `json:"include_sub_domains"`
	Preload           bool    `json:"preload"`
	Score             float64 `json:"score"`
}

// CSPInfo contains Content Security Policy information
type CSPInfo struct {
	Present         bool    `json:"present"`
	Policy          string  `json:"policy"`
	HasUnsafeInline bool    `json:"has_unsafe_inline"`
	HasUnsafeEval   bool    `json:"has_unsafe_eval"`
	HasNonce        bool    `json:"has_nonce"`
	HasHash         bool    `json:"has_hash"`
	Score           float64 `json:"score"`
}

// XFrameOptionsInfo contains X-Frame-Options information
type XFrameOptionsInfo struct {
	Present bool    `json:"present"`
	Value   string  `json:"value"`
	Score   float64 `json:"score"`
}

// XContentTypeOptionsInfo contains X-Content-Type-Options information
type XContentTypeOptionsInfo struct {
	Present bool    `json:"present"`
	Value   string  `json:"value"`
	Score   float64 `json:"score"`
}

// XSSProtectionInfo contains X-XSS-Protection information
type XSSProtectionInfo struct {
	Present bool    `json:"present"`
	Value   string  `json:"value"`
	Mode    string  `json:"mode"`
	Score   float64 `json:"score"`
}

// ReferrerPolicyInfo contains Referrer-Policy information
type ReferrerPolicyInfo struct {
	Present bool    `json:"present"`
	Value   string  `json:"value"`
	Score   float64 `json:"score"`
}

// PermissionsPolicyInfo contains Permissions-Policy information
type PermissionsPolicyInfo struct {
	Present bool    `json:"present"`
	Value   string  `json:"value"`
	Score   float64 `json:"score"`
}

// TLSInfo contains TLS connection information
type TLSInfo struct {
	Version               string   `json:"version"`
	CipherSuite           string   `json:"cipher_suite"`
	KeyExchangeAlgorithm  string   `json:"key_exchange_algorithm"`
	EncryptionAlgorithm   string   `json:"encryption_algorithm"`
	MACAlgorithm          string   `json:"mac_algorithm"`
	ForwardSecrecy        bool     `json:"forward_secrecy"`
	PerfectForwardSecrecy bool     `json:"perfect_forward_secrecy"`
	SupportedVersions     []string `json:"supported_versions"`
	SupportedCiphers      []string `json:"supported_ciphers"`
	TLSScore              float64  `json:"tls_score"`
	Issues                []string `json:"issues"`
}

// SecurityIssue represents a security issue found during analysis
type SecurityIssue struct {
	Category    string        `json:"category"`
	Severity    SecurityLevel `json:"severity"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Impact      string        `json:"impact"`
	Evidence    string        `json:"evidence"`
	Score       float64       `json:"score"`
}

// SecurityRecommendation represents a security recommendation
type SecurityRecommendation struct {
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
	Timeline    string `json:"timeline"`
	Cost        string `json:"cost"`
}

// NewSecurityAnalyzer creates a new security analyzer
func NewSecurityAnalyzer(config *RiskAssessmentConfig, logger *zap.Logger) *SecurityAnalyzer {
	if logger == nil {
		logger = zap.NewNop()
	}

	// Create HTTP client with custom TLS configuration
	client := &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false, // Always verify SSL certificates
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
			},
		},
	}

	return &SecurityAnalyzer{
		config: config,
		logger: logger,
		client: client,
	}
}

// AnalyzeSecurity performs comprehensive security analysis
func (sa *SecurityAnalyzer) AnalyzeSecurity(ctx context.Context, req *RiskAssessmentRequest) (*SecurityAnalysisResult, error) {
	startTime := time.Now()

	sa.logger.Info("Starting security analysis",
		zap.String("website_url", req.WebsiteURL))

	result := &SecurityAnalysisResult{
		WebsiteURL:        req.WebsiteURL,
		AnalysisTimestamp: time.Now(),
		SecurityIssues:    make([]SecurityIssue, 0),
		Recommendations:   make([]SecurityRecommendation, 0),
	}

	// Parse URL
	parsedURL, err := url.Parse(req.WebsiteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure HTTPS
	if parsedURL.Scheme != "https" {
		result.SecurityIssues = append(result.SecurityIssues, SecurityIssue{
			Category:    "protocol",
			Severity:    SecurityLevelCritical,
			Title:       "Non-HTTPS Connection",
			Description: "Website is not using HTTPS protocol",
			Impact:      "Data transmitted in plain text, vulnerable to interception",
			Evidence:    fmt.Sprintf("URL scheme: %s", parsedURL.Scheme),
			Score:       0.0,
		})
		return result, nil
	}

	// Analyze SSL certificate
	if sa.config.SSLVerificationEnabled {
		sslInfo, err := sa.analyzeSSLCertificate(ctx, parsedURL)
		if err != nil {
			sa.logger.Warn("SSL certificate analysis failed", zap.Error(err))
			result.SecurityIssues = append(result.SecurityIssues, SecurityIssue{
				Category:    "ssl",
				Severity:    SecurityLevelCritical,
				Title:       "SSL Certificate Analysis Failed",
				Description: "Unable to analyze SSL certificate",
				Impact:      "Cannot verify SSL certificate security",
				Evidence:    err.Error(),
				Score:       0.0,
			})
		} else {
			result.SSLInfo = sslInfo
		}
	}

	// Analyze security headers
	if sa.config.SecurityHeadersCheckEnabled {
		headersInfo, err := sa.analyzeSecurityHeaders(ctx, parsedURL)
		if err != nil {
			sa.logger.Warn("Security headers analysis failed", zap.Error(err))
			result.SecurityIssues = append(result.SecurityIssues, SecurityIssue{
				Category:    "headers",
				Severity:    SecurityLevelPoor,
				Title:       "Security Headers Analysis Failed",
				Description: "Unable to analyze security headers",
				Impact:      "Cannot verify security header configuration",
				Evidence:    err.Error(),
				Score:       0.0,
			})
		} else {
			result.SecurityHeaders = headersInfo
		}
	}

	// Analyze TLS configuration
	tlsInfo, err := sa.analyzeTLSConfiguration(ctx, parsedURL)
	if err != nil {
		sa.logger.Warn("TLS configuration analysis failed", zap.Error(err))
		result.SecurityIssues = append(result.SecurityIssues, SecurityIssue{
			Category:    "tls",
			Severity:    SecurityLevelPoor,
			Title:       "TLS Configuration Analysis Failed",
			Description: "Unable to analyze TLS configuration",
			Impact:      "Cannot verify TLS security settings",
			Evidence:    err.Error(),
			Score:       0.0,
		})
	} else {
		result.TLSInfo = tlsInfo
	}

	// Calculate overall security score
	result.OverallSecurityScore = sa.calculateOverallSecurityScore(result)
	result.SecurityLevel = sa.determineSecurityLevel(result.OverallSecurityScore)

	// Generate recommendations
	result.Recommendations = sa.generateSecurityRecommendations(result)

	result.ProcessingTime = time.Since(startTime)

	sa.logger.Info("Security analysis completed",
		zap.String("website_url", req.WebsiteURL),
		zap.Float64("security_score", result.OverallSecurityScore),
		zap.String("security_level", string(result.SecurityLevel)),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// analyzeSSLCertificate analyzes SSL certificate
func (sa *SecurityAnalyzer) analyzeSSLCertificate(ctx context.Context, parsedURL *url.URL) (*SSLInfo, error) {
	// Create TLS connection
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		parsedURL.Host+":443",
		&tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish TLS connection: %w", err)
	}
	defer conn.Close()

	// Get certificate
	cert := conn.ConnectionState().PeerCertificates[0]
	now := time.Now()

	sslInfo := &SSLInfo{
		Valid:               true,
		Issuer:              cert.Issuer.CommonName,
		Subject:             cert.Subject.CommonName,
		ValidFrom:           cert.NotBefore,
		ValidTo:             cert.NotAfter,
		DaysUntilExpiration: int(cert.NotAfter.Sub(now).Hours() / 24),
		SignatureAlgorithm:  cert.SignatureAlgorithm.String(),
		KeySize:             cert.PublicKey.(*rsa.PublicKey).N.BitLen(),
		CertificateChain:    make([]string, 0),
		TrustedByBrowser:    true, // Assume trusted for now
		WildcardCertificate: strings.Contains(cert.Subject.CommonName, "*"),
		ExtendedValidation:  len(cert.Subject.Organization) > 0,
		Issues:              make([]string, 0),
	}

	// Check certificate expiration
	if now.After(cert.NotAfter) {
		sslInfo.Valid = false
		sslInfo.Issues = append(sslInfo.Issues, "Certificate expired")
	} else if now.Before(cert.NotBefore) {
		sslInfo.Valid = false
		sslInfo.Issues = append(sslInfo.Issues, "Certificate not yet valid")
	}

	// Check expiration warning (30 days)
	if sslInfo.DaysUntilExpiration <= 30 {
		sslInfo.Issues = append(sslInfo.Issues, "Certificate expires soon")
	}

	// Check key size
	if sslInfo.KeySize < 2048 {
		sslInfo.Issues = append(sslInfo.Issues, "Weak key size")
	}

	// Build certificate chain
	for _, cert := range conn.ConnectionState().PeerCertificates {
		sslInfo.CertificateChain = append(sslInfo.CertificateChain, cert.Subject.CommonName)
	}

	// Calculate certificate score
	sslInfo.CertificateScore = sa.calculateCertificateScore(sslInfo)

	return sslInfo, nil
}

// analyzeSecurityHeaders analyzes security headers
func (sa *SecurityAnalyzer) analyzeSecurityHeaders(ctx context.Context, parsedURL *url.URL) (*SecurityHeaders, error) {
	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SecurityAnalyzer/1.0)")

	resp, err := sa.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	headers := &SecurityHeaders{
		MissingHeaders: make([]string, 0),
		WeakHeaders:    make([]string, 0),
	}

	// Analyze HSTS
	if hsts := resp.Header.Get("Strict-Transport-Security"); hsts != "" {
		headers.HSTS = sa.parseHSTSHeader(hsts)
	} else {
		headers.MissingHeaders = append(headers.MissingHeaders, "Strict-Transport-Security")
	}

	// Analyze CSP
	if csp := resp.Header.Get("Content-Security-Policy"); csp != "" {
		headers.CSP = sa.parseCSPHeader(csp)
	} else {
		headers.MissingHeaders = append(headers.MissingHeaders, "Content-Security-Policy")
	}

	// Analyze X-Frame-Options
	if xfo := resp.Header.Get("X-Frame-Options"); xfo != "" {
		headers.XFrameOptions = sa.parseXFrameOptionsHeader(xfo)
	} else {
		headers.MissingHeaders = append(headers.MissingHeaders, "X-Frame-Options")
	}

	// Analyze X-Content-Type-Options
	if xcto := resp.Header.Get("X-Content-Type-Options"); xcto != "" {
		headers.XContentTypeOptions = sa.parseXContentTypeOptionsHeader(xcto)
	} else {
		headers.MissingHeaders = append(headers.MissingHeaders, "X-Content-Type-Options")
	}

	// Analyze X-XSS-Protection
	if xxp := resp.Header.Get("X-XSS-Protection"); xxp != "" {
		headers.XSSProtection = sa.parseXSSProtectionHeader(xxp)
	} else {
		headers.MissingHeaders = append(headers.MissingHeaders, "X-XSS-Protection")
	}

	// Analyze Referrer-Policy
	if rp := resp.Header.Get("Referrer-Policy"); rp != "" {
		headers.ReferrerPolicy = sa.parseReferrerPolicyHeader(rp)
	} else {
		headers.MissingHeaders = append(headers.MissingHeaders, "Referrer-Policy")
	}

	// Analyze Permissions-Policy
	if pp := resp.Header.Get("Permissions-Policy"); pp != "" {
		headers.PermissionsPolicy = sa.parsePermissionsPolicyHeader(pp)
	} else {
		headers.MissingHeaders = append(headers.MissingHeaders, "Permissions-Policy")
	}

	// Calculate headers score
	headers.HeadersScore = sa.calculateHeadersScore(headers)

	return headers, nil
}

// analyzeTLSConfiguration analyzes TLS configuration
func (sa *SecurityAnalyzer) analyzeTLSConfiguration(ctx context.Context, parsedURL *url.URL) (*TLSInfo, error) {
	// Create TLS connection
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		parsedURL.Host+":443",
		&tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish TLS connection: %w", err)
	}
	defer conn.Close()

	state := conn.ConnectionState()

	// Get TLS version string
	var versionString string
	switch state.Version {
	case tls.VersionTLS10:
		versionString = "TLS 1.0"
	case tls.VersionTLS11:
		versionString = "TLS 1.1"
	case tls.VersionTLS12:
		versionString = "TLS 1.2"
	case tls.VersionTLS13:
		versionString = "TLS 1.3"
	default:
		versionString = "Unknown"
	}

	// Get cipher suite info
	cipherSuite := tls.CipherSuiteName(state.CipherSuite)
	if cipherSuite == "" {
		cipherSuite = "Unknown"
	}

	// Check if cipher suite supports forward secrecy
	hasForwardSecrecy := sa.hasForwardSecrecy(state.CipherSuite)

	tlsInfo := &TLSInfo{
		Version:               versionString,
		CipherSuite:           cipherSuite,
		KeyExchangeAlgorithm:  cipherSuite,
		EncryptionAlgorithm:   cipherSuite,
		MACAlgorithm:          cipherSuite,
		ForwardSecrecy:        hasForwardSecrecy,
		PerfectForwardSecrecy: hasForwardSecrecy,
		SupportedVersions:     make([]string, 0),
		SupportedCiphers:      make([]string, 0),
		Issues:                make([]string, 0),
	}

	// Check TLS version
	if state.Version < tls.VersionTLS12 {
		tlsInfo.Issues = append(tlsInfo.Issues, "Weak TLS version")
	}

	// Check cipher suite
	if !hasForwardSecrecy {
		tlsInfo.Issues = append(tlsInfo.Issues, "No forward secrecy")
	}

	// Calculate TLS score
	tlsInfo.TLSScore = sa.calculateTLSScore(tlsInfo)

	return tlsInfo, nil
}

// Helper methods for parsing headers and calculating scores
func (sa *SecurityAnalyzer) parseHSTSHeader(header string) *HSTSInfo {
	// Implementation for parsing HSTS header
	return &HSTSInfo{
		Present:           true,
		MaxAge:            31536000, // Default to 1 year
		IncludeSubDomains: strings.Contains(header, "includeSubDomains"),
		Preload:           strings.Contains(header, "preload"),
		Score:             0.9,
	}
}

func (sa *SecurityAnalyzer) parseCSPHeader(header string) *CSPInfo {
	// Implementation for parsing CSP header
	return &CSPInfo{
		Present:         true,
		Policy:          header,
		HasUnsafeInline: strings.Contains(header, "unsafe-inline"),
		HasUnsafeEval:   strings.Contains(header, "unsafe-eval"),
		HasNonce:        strings.Contains(header, "nonce-"),
		HasHash:         strings.Contains(header, "sha256-"),
		Score:           0.8,
	}
}

func (sa *SecurityAnalyzer) parseXFrameOptionsHeader(header string) *XFrameOptionsInfo {
	// Implementation for parsing X-Frame-Options header
	return &XFrameOptionsInfo{
		Present: true,
		Value:   header,
		Score:   0.9,
	}
}

func (sa *SecurityAnalyzer) parseXContentTypeOptionsHeader(header string) *XContentTypeOptionsInfo {
	// Implementation for parsing X-Content-Type-Options header
	return &XContentTypeOptionsInfo{
		Present: true,
		Value:   header,
		Score:   0.9,
	}
}

func (sa *SecurityAnalyzer) parseXSSProtectionHeader(header string) *XSSProtectionInfo {
	// Implementation for parsing X-XSS-Protection header
	return &XSSProtectionInfo{
		Present: true,
		Value:   header,
		Mode:    "block",
		Score:   0.7,
	}
}

func (sa *SecurityAnalyzer) parseReferrerPolicyHeader(header string) *ReferrerPolicyInfo {
	// Implementation for parsing Referrer-Policy header
	return &ReferrerPolicyInfo{
		Present: true,
		Value:   header,
		Score:   0.8,
	}
}

func (sa *SecurityAnalyzer) parsePermissionsPolicyHeader(header string) *PermissionsPolicyInfo {
	// Implementation for parsing Permissions-Policy header
	return &PermissionsPolicyInfo{
		Present: true,
		Value:   header,
		Score:   0.8,
	}
}

// Score calculation methods
func (sa *SecurityAnalyzer) calculateCertificateScore(sslInfo *SSLInfo) float64 {
	score := 1.0

	if !sslInfo.Valid {
		score -= 1.0
	}

	if sslInfo.DaysUntilExpiration <= 30 {
		score -= 0.2
	}

	if sslInfo.KeySize < 2048 {
		score -= 0.3
	}

	if sslInfo.WildcardCertificate {
		score -= 0.1
	}

	if !sslInfo.ExtendedValidation {
		score -= 0.1
	}

	return max(0.0, score)
}

func (sa *SecurityAnalyzer) calculateHeadersScore(headers *SecurityHeaders) float64 {
	score := 0.0
	count := 0

	if headers.HSTS != nil {
		score += headers.HSTS.Score
		count++
	}

	if headers.CSP != nil {
		score += headers.CSP.Score
		count++
	}

	if headers.XFrameOptions != nil {
		score += headers.XFrameOptions.Score
		count++
	}

	if headers.XContentTypeOptions != nil {
		score += headers.XContentTypeOptions.Score
		count++
	}

	if headers.XSSProtection != nil {
		score += headers.XSSProtection.Score
		count++
	}

	if headers.ReferrerPolicy != nil {
		score += headers.ReferrerPolicy.Score
		count++
	}

	if headers.PermissionsPolicy != nil {
		score += headers.PermissionsPolicy.Score
		count++
	}

	if count == 0 {
		return 0.0
	}

	return score / float64(count)
}

func (sa *SecurityAnalyzer) calculateTLSScore(tlsInfo *TLSInfo) float64 {
	score := 1.0

	if tlsInfo.Version < "TLS 1.2" {
		score -= 0.5
	}

	if !tlsInfo.ForwardSecrecy {
		score -= 0.3
	}

	return max(0.0, score)
}

func (sa *SecurityAnalyzer) calculateOverallSecurityScore(result *SecurityAnalysisResult) float64 {
	scores := make([]float64, 0)
	weights := make([]float64, 0)

	// SSL certificate score
	if result.SSLInfo != nil {
		scores = append(scores, result.SSLInfo.CertificateScore)
		weights = append(weights, 0.4)
	}

	// Security headers score
	if result.SecurityHeaders != nil {
		scores = append(scores, result.SecurityHeaders.HeadersScore)
		weights = append(weights, 0.3)
	}

	// TLS configuration score
	if result.TLSInfo != nil {
		scores = append(scores, result.TLSInfo.TLSScore)
		weights = append(weights, 0.3)
	}

	if len(scores) == 0 {
		return 0.0
	}

	// Calculate weighted average
	totalScore := 0.0
	totalWeight := 0.0

	for i, score := range scores {
		totalScore += score * weights[i]
		totalWeight += weights[i]
	}

	return totalScore / totalWeight
}

func (sa *SecurityAnalyzer) determineSecurityLevel(score float64) SecurityLevel {
	switch {
	case score >= 0.9:
		return SecurityLevelExcellent
	case score >= 0.7:
		return SecurityLevelGood
	case score >= 0.5:
		return SecurityLevelFair
	case score >= 0.3:
		return SecurityLevelPoor
	default:
		return SecurityLevelCritical
	}
}

func (sa *SecurityAnalyzer) generateSecurityRecommendations(result *SecurityAnalysisResult) []SecurityRecommendation {
	recommendations := make([]SecurityRecommendation, 0)

	// SSL certificate recommendations
	if result.SSLInfo != nil {
		if !result.SSLInfo.Valid {
			recommendations = append(recommendations, SecurityRecommendation{
				Category:    "ssl",
				Priority:    "critical",
				Title:       "Fix SSL Certificate",
				Description: "SSL certificate is invalid or expired",
				Action:      "Renew SSL certificate immediately",
				Impact:      "Website will be marked as insecure by browsers",
				Effort:      "low",
				Timeline:    "immediate",
				Cost:        "low",
			})
		}

		if result.SSLInfo.DaysUntilExpiration <= 30 {
			recommendations = append(recommendations, SecurityRecommendation{
				Category:    "ssl",
				Priority:    "high",
				Title:       "Renew SSL Certificate",
				Description: "SSL certificate expires soon",
				Action:      "Renew SSL certificate before expiration",
				Impact:      "Prevent certificate expiration",
				Effort:      "low",
				Timeline:    "1 week",
				Cost:        "low",
			})
		}
	}

	// Security headers recommendations
	if result.SecurityHeaders != nil {
		for _, missingHeader := range result.SecurityHeaders.MissingHeaders {
			recommendations = append(recommendations, SecurityRecommendation{
				Category:    "headers",
				Priority:    "medium",
				Title:       "Add " + missingHeader + " Header",
				Description: "Missing security header: " + missingHeader,
				Action:      "Configure web server to include " + missingHeader + " header",
				Impact:      "Improve security posture",
				Effort:      "medium",
				Timeline:    "1 month",
				Cost:        "low",
			})
		}
	}

	return recommendations
}

// hasForwardSecrecy checks if a cipher suite supports forward secrecy
func (sa *SecurityAnalyzer) hasForwardSecrecy(cipherSuite uint16) bool {
	// Common forward secrecy cipher suites
	forwardSecrecyCiphers := map[uint16]bool{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   true,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   true,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: true,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: true,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    true,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  true,
		// Add more as needed
	}

	return forwardSecrecyCiphers[cipherSuite]
}

// Helper function
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
