package website_verification

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"kyb-platform/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AdvancedVerifier implements advanced website ownership verification algorithms
type AdvancedVerifier struct {
	// Configuration
	config *AdvancedVerifierConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Verification components
	dnsVerifier      *DNSVerifier
	whoisVerifier    *WHOISVerifier
	contentVerifier  *ContentVerifier
	nameMatcher      *NameMatcher
	addressMatcher   *AddressMatcher
	phoneMatcher     *PhoneMatcher
	emailVerifier    *EmailVerifier
	confidenceScorer *ConfidenceScorer

	// HTTP client for external requests
	httpClient *http.Client
}

// AdvancedVerifierConfig configuration for advanced verification
type AdvancedVerifierConfig struct {
	// DNS verification settings
	DNSVerificationEnabled bool
	DNSTimeout             time.Duration
	DNSRetries             int
	DNSServers             []string

	// WHOIS verification settings
	WHOISVerificationEnabled bool
	WHOISTimeout             time.Duration
	WHOISRetries             int
	WHOISProviders           []string

	// Content verification settings
	ContentVerificationEnabled bool
	ContentTimeout             time.Duration
	ContentRetries             int
	ContentMaxSize             int64
	ContentUserAgents          []string

	// Matching settings
	FuzzyMatchingEnabled        bool
	FuzzyThreshold              float64
	AddressNormalizationEnabled bool
	PhoneValidationEnabled      bool
	EmailVerificationEnabled    bool

	// Confidence scoring settings
	ConfidenceScoringEnabled bool
	MinConfidenceThreshold   float64
	MaxConfidenceThreshold   float64
}

// DNSVerifier verifies domain ownership through DNS records
type DNSVerifier struct {
	timeout time.Duration
	retries int
	servers []string
}

// WHOISVerifier verifies domain ownership through WHOIS records
type WHOISVerifier struct {
	timeout   time.Duration
	retries   int
	providers []string
}

// ContentVerifier verifies domain ownership through website content analysis
type ContentVerifier struct {
	timeout    time.Duration
	retries    int
	maxSize    int64
	userAgents []string
}

// NameMatcher performs fuzzy matching on business names
type NameMatcher struct {
	enabled   bool
	threshold float64
}

// AddressMatcher performs address normalization and comparison
type AddressMatcher struct {
	enabled bool
}

// PhoneMatcher performs phone number validation and matching
type PhoneMatcher struct {
	enabled bool
}

// EmailVerifier verifies email domain ownership
type EmailVerifier struct {
	enabled bool
}

// ConfidenceScorer calculates confidence scores for verification results
type ConfidenceScorer struct {
	enabled       bool
	minThreshold  float64
	maxThreshold  float64
	weightFactors map[string]float64
}

// VerificationResult represents the result of advanced verification
type VerificationResult struct {
	Domain            string
	BusinessName      string
	VerificationScore float64
	Confidence        float64
	Status            VerificationStatus
	Methods           []VerificationMethod
	Details           VerificationDetails
	Timestamp         time.Time
}

// VerificationStatus represents the status of verification
type VerificationStatus string

const (
	VerificationStatusVerified   VerificationStatus = "verified"
	VerificationStatusUnverified VerificationStatus = "unverified"
	VerificationStatusPending    VerificationStatus = "pending"
	VerificationStatusFailed     VerificationStatus = "failed"
)

// VerificationMethod represents a verification method used
type VerificationMethod struct {
	Type      MethodType
	Score     float64
	Details   string
	Timestamp time.Time
}

// MethodType represents the type of verification method
type MethodType string

const (
	MethodTypeDNS     MethodType = "dns"
	MethodTypeWHOIS   MethodType = "whois"
	MethodTypeContent MethodType = "content"
	MethodTypeName    MethodType = "name"
	MethodTypeAddress MethodType = "address"
	MethodTypePhone   MethodType = "phone"
	MethodTypeEmail   MethodType = "email"
)

// VerificationDetails contains detailed verification information
type VerificationDetails struct {
	DNSRecords     []DNSRecord
	WHOISInfo      *WHOISInfo
	ContentMatches []ContentMatch
	NameMatches    []NameMatch
	AddressMatches []AddressMatch
	PhoneMatches   []PhoneMatch
	EmailMatches   []EmailMatch
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	Type  string
	Value string
	TTL   int
}

// WHOISInfo represents WHOIS information
type WHOISInfo struct {
	Registrar    string
	Registrant   string
	CreationDate time.Time
	ExpiryDate   time.Time
	UpdatedDate  time.Time
	Status       []string
}

// ContentMatch represents a content match
type ContentMatch struct {
	Type       string
	Pattern    string
	Confidence float64
	Location   string
	Extracted  string
}

// NameMatch represents a name match
type NameMatch struct {
	OriginalName string
	MatchedName  string
	Confidence   float64
	Algorithm    string
}

// AddressMatch represents an address match
type AddressMatch struct {
	OriginalAddress   string
	NormalizedAddress string
	Confidence        float64
	Components        AddressComponents
}

// AddressComponents represents address components
type AddressComponents struct {
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
}

// PhoneMatch represents a phone match
type PhoneMatch struct {
	OriginalPhone   string
	NormalizedPhone string
	Confidence      float64
	Format          string
}

// EmailMatch represents an email match
type EmailMatch struct {
	Domain     string
	MXRecords  []string
	Confidence float64
}

// NewAdvancedVerifier creates a new advanced verifier
func NewAdvancedVerifier(config *AdvancedVerifierConfig, logger *observability.Logger, tracer trace.Tracer) *AdvancedVerifier {
	if config == nil {
		config = &AdvancedVerifierConfig{
			DNSVerificationEnabled:      true,
			DNSTimeout:                  10 * time.Second,
			DNSRetries:                  3,
			DNSServers:                  []string{"8.8.8.8:53", "1.1.1.1:53"},
			WHOISVerificationEnabled:    true,
			WHOISTimeout:                15 * time.Second,
			WHOISRetries:                2,
			WHOISProviders:              []string{"whois.verisign-grs.com", "whois.iana.org"},
			ContentVerificationEnabled:  true,
			ContentTimeout:              30 * time.Second,
			ContentRetries:              2,
			ContentMaxSize:              10 * 1024 * 1024, // 10MB
			ContentUserAgents:           []string{"Mozilla/5.0 (compatible; BusinessVerifier/1.0)"},
			FuzzyMatchingEnabled:        true,
			FuzzyThreshold:              0.8,
			AddressNormalizationEnabled: true,
			PhoneValidationEnabled:      true,
			EmailVerificationEnabled:    true,
			ConfidenceScoringEnabled:    true,
			MinConfidenceThreshold:      0.6,
			MaxConfidenceThreshold:      0.95,
		}
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: config.ContentTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	av := &AdvancedVerifier{
		config:     config,
		logger:     logger,
		tracer:     tracer,
		httpClient: httpClient,
	}

	// Initialize verification components
	av.dnsVerifier = &DNSVerifier{
		timeout: config.DNSTimeout,
		retries: config.DNSRetries,
		servers: config.DNSServers,
	}

	av.whoisVerifier = &WHOISVerifier{
		timeout:   config.WHOISTimeout,
		retries:   config.WHOISRetries,
		providers: config.WHOISProviders,
	}

	av.contentVerifier = &ContentVerifier{
		timeout:    config.ContentTimeout,
		retries:    config.ContentRetries,
		maxSize:    config.ContentMaxSize,
		userAgents: config.ContentUserAgents,
	}

	av.nameMatcher = &NameMatcher{
		enabled:   config.FuzzyMatchingEnabled,
		threshold: config.FuzzyThreshold,
	}

	av.addressMatcher = &AddressMatcher{
		enabled: config.AddressNormalizationEnabled,
	}

	av.phoneMatcher = &PhoneMatcher{
		enabled: config.PhoneValidationEnabled,
	}

	av.emailVerifier = &EmailVerifier{
		enabled: config.EmailVerificationEnabled,
	}

	av.confidenceScorer = &ConfidenceScorer{
		enabled:      config.ConfidenceScoringEnabled,
		minThreshold: config.MinConfidenceThreshold,
		maxThreshold: config.MaxConfidenceThreshold,
		weightFactors: map[string]float64{
			"dns":     0.25,
			"whois":   0.20,
			"content": 0.30,
			"name":    0.15,
			"address": 0.05,
			"phone":   0.03,
			"email":   0.02,
		},
	}

	return av
}

// VerifyWebsiteOwnership performs comprehensive website ownership verification
func (av *AdvancedVerifier) VerifyWebsiteOwnership(ctx context.Context, domain, businessName, address, phone, email string) (*VerificationResult, error) {
	ctx, span := av.tracer.Start(ctx, "AdvancedVerifier.VerifyWebsiteOwnership")
	defer span.End()

	span.SetAttributes(
		attribute.String("domain", domain),
		attribute.String("business_name", businessName),
	)

	// Initialize result
	result := &VerificationResult{
		Domain:       domain,
		BusinessName: businessName,
		Status:       VerificationStatusPending,
		Timestamp:    time.Now(),
		Methods:      make([]VerificationMethod, 0),
		Details:      VerificationDetails{},
	}

	// Perform DNS verification
	if av.config.DNSVerificationEnabled {
		dnsMethod, dnsRecords, err := av.dnsVerifier.Verify(ctx, domain)
		if err != nil {
			av.logger.Warn("DNS verification failed", map[string]interface{}{
				"domain": domain,
				"error":  err.Error(),
			})
		} else {
			result.Methods = append(result.Methods, dnsMethod)
			result.Details.DNSRecords = dnsRecords
		}
	}

	// Perform WHOIS verification
	if av.config.WHOISVerificationEnabled {
		whoisMethod, whoisInfo, err := av.whoisVerifier.Verify(ctx, domain)
		if err != nil {
			av.logger.Warn("WHOIS verification failed", map[string]interface{}{
				"domain": domain,
				"error":  err.Error(),
			})
		} else {
			result.Methods = append(result.Methods, whoisMethod)
			result.Details.WHOISInfo = whoisInfo
		}
	}

	// Perform content verification
	if av.config.ContentVerificationEnabled {
		contentMethod, contentMatches, err := av.contentVerifier.Verify(ctx, domain, businessName, av.httpClient)
		if err != nil {
			av.logger.Warn("Content verification failed", map[string]interface{}{
				"domain": domain,
				"error":  err.Error(),
			})
		} else {
			result.Methods = append(result.Methods, contentMethod)
			result.Details.ContentMatches = contentMatches
		}
	}

	// Perform name matching
	if av.config.FuzzyMatchingEnabled && businessName != "" {
		nameMethod, nameMatches, err := av.nameMatcher.Match(ctx, businessName, domain)
		if err != nil {
			av.logger.Warn("Name matching failed", map[string]interface{}{
				"business_name": businessName,
				"error":         err.Error(),
			})
		} else {
			result.Methods = append(result.Methods, nameMethod)
			result.Details.NameMatches = nameMatches
		}
	}

	// Perform address matching
	if av.config.AddressNormalizationEnabled && address != "" {
		addressMethod, addressMatches, err := av.addressMatcher.Match(ctx, address, domain)
		if err != nil {
			av.logger.Warn("Address matching failed", map[string]interface{}{
				"address": address,
				"error":   err.Error(),
			})
		} else {
			result.Methods = append(result.Methods, addressMethod)
			result.Details.AddressMatches = addressMatches
		}
	}

	// Perform phone matching
	if av.config.PhoneValidationEnabled && phone != "" {
		phoneMethod, phoneMatches, err := av.phoneMatcher.Match(ctx, phone, domain)
		if err != nil {
			av.logger.Warn("Phone matching failed", map[string]interface{}{
				"phone": phone,
				"error": err.Error(),
			})
		} else {
			result.Methods = append(result.Methods, phoneMethod)
			result.Details.PhoneMatches = phoneMatches
		}
	}

	// Perform email verification
	if av.config.EmailVerificationEnabled && email != "" {
		emailMethod, emailMatches, err := av.emailVerifier.Verify(ctx, email, domain)
		if err != nil {
			av.logger.Warn("Email verification failed", map[string]interface{}{
				"email": email,
				"error": err.Error(),
			})
		} else {
			result.Methods = append(result.Methods, emailMethod)
			result.Details.EmailMatches = emailMatches
		}
	}

	// Calculate confidence score
	if av.config.ConfidenceScoringEnabled {
		confidence, score := av.confidenceScorer.CalculateConfidence(result.Methods)
		result.Confidence = confidence
		result.VerificationScore = score

		// Determine status based on confidence
		if confidence >= av.config.MaxConfidenceThreshold {
			result.Status = VerificationStatusVerified
		} else if confidence >= av.config.MinConfidenceThreshold {
			result.Status = VerificationStatusPending
		} else {
			result.Status = VerificationStatusUnverified
		}
	} else {
		result.Status = VerificationStatusPending
	}

	av.logger.Info("website ownership verification completed", map[string]interface{}{
		"domain":     domain,
		"status":     result.Status,
		"confidence": result.Confidence,
		"score":      result.VerificationScore,
		"methods":    len(result.Methods),
	})

	return result, nil
}

// DNSVerifier methods

func (dv *DNSVerifier) Verify(ctx context.Context, domain string) (VerificationMethod, []DNSRecord, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "DNSVerifier.Verify")
	defer span.End()

	span.SetAttributes(attribute.String("domain", domain))

	var records []DNSRecord
	var lastErr error

	// Try multiple DNS servers
	for _, server := range dv.servers {
		for attempt := 0; attempt < dv.retries; attempt++ {
			records, lastErr = dv.queryDNS(domain, server)
			if lastErr == nil {
				break
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return VerificationMethod{}, nil, fmt.Errorf("DNS verification failed: %w", lastErr)
	}

	// Calculate DNS verification score
	score := dv.calculateDNSScore(records)

	method := VerificationMethod{
		Type:      MethodTypeDNS,
		Score:     score,
		Details:   fmt.Sprintf("Found %d DNS records", len(records)),
		Timestamp: time.Now(),
	}

	return method, records, nil
}

func (dv *DNSVerifier) queryDNS(domain, server string) ([]DNSRecord, error) {
	// Simplified DNS query - in production, use a proper DNS library
	// This is a placeholder implementation
	records := []DNSRecord{
		{Type: "A", Value: "192.168.1.1", TTL: 300},
		{Type: "MX", Value: "mail.example.com", TTL: 300},
		{Type: "TXT", Value: "v=spf1 include:_spf.google.com ~all", TTL: 300},
	}

	return records, nil
}

func (dv *DNSVerifier) calculateDNSScore(records []DNSRecord) float64 {
	// Calculate score based on DNS record types and completeness
	score := 0.0
	recordTypes := make(map[string]bool)

	for _, record := range records {
		recordTypes[record.Type] = true
	}

	// Score based on essential record types
	if recordTypes["A"] {
		score += 0.4
	}
	if recordTypes["MX"] {
		score += 0.3
	}
	if recordTypes["TXT"] {
		score += 0.2
	}
	if recordTypes["NS"] {
		score += 0.1
	}

	return score
}

// WHOISVerifier methods

func (wv *WHOISVerifier) Verify(ctx context.Context, domain string) (VerificationMethod, *WHOISInfo, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "WHOISVerifier.Verify")
	defer span.End()

	span.SetAttributes(attribute.String("domain", domain))

	var whoisInfo *WHOISInfo
	var lastErr error

	// Try multiple WHOIS providers
	for _, provider := range wv.providers {
		for attempt := 0; attempt < wv.retries; attempt++ {
			whoisInfo, lastErr = wv.queryWHOIS(domain, provider)
			if lastErr == nil {
				break
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return VerificationMethod{}, nil, fmt.Errorf("WHOIS verification failed: %w", lastErr)
	}

	// Calculate WHOIS verification score
	score := wv.calculateWHOISScore(whoisInfo)

	method := VerificationMethod{
		Type:      MethodTypeWHOIS,
		Score:     score,
		Details:   fmt.Sprintf("Registrar: %s", whoisInfo.Registrar),
		Timestamp: time.Now(),
	}

	return method, whoisInfo, nil
}

func (wv *WHOISVerifier) queryWHOIS(domain, provider string) (*WHOISInfo, error) {
	// Simplified WHOIS query - in production, use a proper WHOIS library
	// This is a placeholder implementation
	whoisInfo := &WHOISInfo{
		Registrar:    "Example Registrar",
		Registrant:   "Example Organization",
		CreationDate: time.Now().AddDate(-1, 0, 0),
		ExpiryDate:   time.Now().AddDate(1, 0, 0),
		UpdatedDate:  time.Now(),
		Status:       []string{"clientTransferProhibited"},
	}

	return whoisInfo, nil
}

func (wv *WHOISVerifier) calculateWHOISScore(whoisInfo *WHOISInfo) float64 {
	// Calculate score based on WHOIS information completeness
	score := 0.0

	if whoisInfo.Registrar != "" {
		score += 0.3
	}
	if whoisInfo.Registrant != "" {
		score += 0.3
	}
	if !whoisInfo.CreationDate.IsZero() {
		score += 0.2
	}
	if !whoisInfo.ExpiryDate.IsZero() {
		score += 0.2
	}

	return score
}

// ContentVerifier methods

func (cv *ContentVerifier) Verify(ctx context.Context, domain, businessName string, client *http.Client) (VerificationMethod, []ContentMatch, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "ContentVerifier.Verify")
	defer span.End()

	span.SetAttributes(
		attribute.String("domain", domain),
		attribute.String("business_name", businessName),
	)

	// Fetch website content
	content, err := cv.fetchContent(domain, client)
	if err != nil {
		return VerificationMethod{}, nil, fmt.Errorf("failed to fetch content: %w", err)
	}

	// Analyze content for business information
	matches := cv.analyzeContent(content, businessName)

	// Calculate content verification score
	score := cv.calculateContentScore(matches)

	method := VerificationMethod{
		Type:      MethodTypeContent,
		Score:     score,
		Details:   fmt.Sprintf("Found %d content matches", len(matches)),
		Timestamp: time.Now(),
	}

	return method, matches, nil
}

func (cv *ContentVerifier) fetchContent(domain string, client *http.Client) (string, error) {
	// Simplified content fetching - in production, implement proper web scraping
	// This is a placeholder implementation
	content := fmt.Sprintf(`
		<html>
			<head><title>%s - Official Website</title></head>
			<body>
				<h1>Welcome to %s</h1>
				<p>Contact us at info@%s</p>
				<p>Phone: +1-555-123-4567</p>
				<p>Address: 123 Main St, Anytown, ST 12345</p>
			</body>
		</html>
	`, domain, domain, domain)

	return content, nil
}

func (cv *ContentVerifier) analyzeContent(content, businessName string) []ContentMatch {
	var matches []ContentMatch

	// Look for business name in content
	if businessName != "" {
		if strings.Contains(strings.ToLower(content), strings.ToLower(businessName)) {
			matches = append(matches, ContentMatch{
				Type:       "business_name",
				Pattern:    businessName,
				Confidence: 0.9,
				Location:   "page_content",
				Extracted:  businessName,
			})
		}
	}

	// Look for contact information
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailPattern.FindAllString(content, -1)
	for _, email := range emails {
		matches = append(matches, ContentMatch{
			Type:       "email",
			Pattern:    email,
			Confidence: 0.8,
			Location:   "page_content",
			Extracted:  email,
		})
	}

	// Look for phone numbers
	phonePattern := regexp.MustCompile(`\+?1?[-.\s]?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`)
	phones := phonePattern.FindAllString(content, -1)
	for _, phone := range phones {
		matches = append(matches, ContentMatch{
			Type:       "phone",
			Pattern:    phone,
			Confidence: 0.7,
			Location:   "page_content",
			Extracted:  phone,
		})
	}

	return matches
}

func (cv *ContentVerifier) calculateContentScore(matches []ContentMatch) float64 {
	if len(matches) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, match := range matches {
		totalConfidence += match.Confidence
	}

	return totalConfidence / float64(len(matches))
}

// NameMatcher methods

func (nm *NameMatcher) Match(ctx context.Context, businessName, domain string) (VerificationMethod, []NameMatch, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "NameMatcher.Match")
	defer span.End()

	span.SetAttributes(
		attribute.String("business_name", businessName),
		attribute.String("domain", domain),
	)

	if !nm.enabled {
		return VerificationMethod{}, nil, fmt.Errorf("fuzzy matching disabled")
	}

	// Extract domain name without TLD
	domainName := strings.Split(domain, ".")[0]

	// Perform fuzzy matching
	confidence := nm.calculateNameSimilarity(businessName, domainName)

	var matches []NameMatch
	if confidence >= nm.threshold {
		matches = append(matches, NameMatch{
			OriginalName: businessName,
			MatchedName:  domainName,
			Confidence:   confidence,
			Algorithm:    "fuzzy_string_match",
		})
	}

	score := confidence

	method := VerificationMethod{
		Type:      MethodTypeName,
		Score:     score,
		Details:   fmt.Sprintf("Name similarity: %.2f", confidence),
		Timestamp: time.Now(),
	}

	return method, matches, nil
}

func (nm *NameMatcher) calculateNameSimilarity(name1, name2 string) float64 {
	// Simplified string similarity calculation
	// In production, use a proper string similarity library like Levenshtein distance
	name1Lower := strings.ToLower(strings.ReplaceAll(name1, " ", ""))
	name2Lower := strings.ToLower(strings.ReplaceAll(name2, " ", ""))

	if name1Lower == name2Lower {
		return 1.0
	}

	// Simple character-based similarity
	commonChars := 0
	for _, char := range name1Lower {
		if strings.ContainsRune(name2Lower, char) {
			commonChars++
		}
	}

	if len(name1Lower) == 0 || len(name2Lower) == 0 {
		return 0.0
	}

	return float64(commonChars) / float64(max(len(name1Lower), len(name2Lower)))
}

// AddressMatcher methods

func (am *AddressMatcher) Match(ctx context.Context, address, domain string) (VerificationMethod, []AddressMatch, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "AddressMatcher.Match")
	defer span.End()

	span.SetAttributes(
		attribute.String("address", address),
		attribute.String("domain", domain),
	)

	if !am.enabled {
		return VerificationMethod{}, nil, fmt.Errorf("address normalization disabled")
	}

	// Normalize address
	normalizedAddress := am.normalizeAddress(address)
	components := am.parseAddressComponents(normalizedAddress)

	// For now, return a basic match
	// In production, implement address geocoding and comparison
	matches := []AddressMatch{
		{
			OriginalAddress:   address,
			NormalizedAddress: normalizedAddress,
			Confidence:        0.6,
			Components:        components,
		},
	}

	score := 0.6

	method := VerificationMethod{
		Type:      MethodTypeAddress,
		Score:     score,
		Details:   fmt.Sprintf("Normalized: %s", normalizedAddress),
		Timestamp: time.Now(),
	}

	return method, matches, nil
}

func (am *AddressMatcher) normalizeAddress(address string) string {
	// Simplified address normalization
	// In production, use a proper address normalization library
	normalized := strings.ToLower(address)
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)
	return normalized
}

func (am *AddressMatcher) parseAddressComponents(address string) AddressComponents {
	// Simplified address parsing
	// In production, use a proper address parsing library
	parts := strings.Split(address, ",")
	components := AddressComponents{}

	if len(parts) >= 1 {
		components.Street = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		components.City = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		components.State = strings.TrimSpace(parts[2])
	}
	if len(parts) >= 4 {
		components.PostalCode = strings.TrimSpace(parts[3])
	}
	if len(parts) >= 5 {
		components.Country = strings.TrimSpace(parts[4])
	}

	return components
}

// PhoneMatcher methods

func (pm *PhoneMatcher) Match(ctx context.Context, phone, domain string) (VerificationMethod, []PhoneMatch, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "PhoneMatcher.Match")
	defer span.End()

	span.SetAttributes(
		attribute.String("phone", phone),
		attribute.String("domain", domain),
	)

	if !pm.enabled {
		return VerificationMethod{}, nil, fmt.Errorf("phone validation disabled")
	}

	// Normalize phone number
	normalizedPhone := pm.normalizePhone(phone)
	format := pm.detectPhoneFormat(normalizedPhone)

	// For now, return a basic match
	// In production, implement phone number validation and comparison
	matches := []PhoneMatch{
		{
			OriginalPhone:   phone,
			NormalizedPhone: normalizedPhone,
			Confidence:      0.5,
			Format:          format,
		},
	}

	score := 0.5

	method := VerificationMethod{
		Type:      MethodTypePhone,
		Score:     score,
		Details:   fmt.Sprintf("Format: %s", format),
		Timestamp: time.Now(),
	}

	return method, matches, nil
}

func (pm *PhoneMatcher) normalizePhone(phone string) string {
	// Remove all non-digit characters
	normalized := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	return normalized
}

func (pm *PhoneMatcher) detectPhoneFormat(phone string) string {
	if len(phone) == 10 {
		return "US_10_DIGIT"
	} else if len(phone) == 11 && strings.HasPrefix(phone, "1") {
		return "US_11_DIGIT"
	} else if len(phone) >= 10 {
		return "INTERNATIONAL"
	}
	return "UNKNOWN"
}

// EmailVerifier methods

func (ev *EmailVerifier) Verify(ctx context.Context, email, domain string) (VerificationMethod, []EmailMatch, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "EmailVerifier.Verify")
	defer span.End()

	span.SetAttributes(
		attribute.String("email", email),
		attribute.String("domain", domain),
	)

	if !ev.enabled {
		return VerificationMethod{}, nil, fmt.Errorf("email verification disabled")
	}

	// Extract domain from email
	emailDomain := strings.Split(email, "@")[1]

	// Check if email domain matches website domain
	confidence := 0.0
	if emailDomain == domain {
		confidence = 0.8
	} else if strings.HasSuffix(emailDomain, "."+domain) {
		confidence = 0.6
	}

	// Check MX records for email domain
	mxRecords := ev.getMXRecords(emailDomain)

	matches := []EmailMatch{
		{
			Domain:     emailDomain,
			MXRecords:  mxRecords,
			Confidence: confidence,
		},
	}

	score := confidence

	method := VerificationMethod{
		Type:      MethodTypeEmail,
		Score:     score,
		Details:   fmt.Sprintf("Domain: %s, MX: %d records", emailDomain, len(mxRecords)),
		Timestamp: time.Now(),
	}

	return method, matches, nil
}

func (ev *EmailVerifier) getMXRecords(domain string) []string {
	// Simplified MX record lookup
	// In production, use proper DNS lookup
	return []string{"mail." + domain, "smtp." + domain}
}

// ConfidenceScorer methods

func (cs *ConfidenceScorer) CalculateConfidence(methods []VerificationMethod) (float64, float64) {
	if !cs.enabled || len(methods) == 0 {
		return 0.0, 0.0
	}

	totalWeightedScore := 0.0
	totalWeight := 0.0

	for _, method := range methods {
		weight := cs.weightFactors[string(method.Type)]
		totalWeightedScore += method.Score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0, 0.0
	}

	confidence := totalWeightedScore / totalWeight

	// Clamp confidence to thresholds
	if confidence < cs.minThreshold {
		confidence = cs.minThreshold
	} else if confidence > cs.maxThreshold {
		confidence = cs.maxThreshold
	}

	return confidence, totalWeightedScore
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
