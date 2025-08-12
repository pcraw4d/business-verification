package webanalysis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// WebsiteValidator provides comprehensive website validation and verification
type WebsiteValidator struct {
	authenticityChecker *AuthenticityChecker
	trafficAnalyzer     *TrafficAnalyzer
	domainReputation    *DomainReputationChecker
	sslValidator        *SSLValidator
	contentQuality      *ContentQualityAssessor
	config              ValidatorConfig
	mu                  sync.RWMutex
}

// ValidatorConfig holds configuration for website validation
type ValidatorConfig struct {
	EnableAuthenticityCheck   bool          `json:"enable_authenticity_check"`
	EnableTrafficAnalysis     bool          `json:"enable_traffic_analysis"`
	EnableDomainReputation    bool          `json:"enable_domain_reputation"`
	EnableSSLValidation       bool          `json:"enable_ssl_validation"`
	EnableContentQualityCheck bool          `json:"enable_content_quality_check"`
	Timeout                   time.Duration `json:"timeout"`
	MaxRedirects              int           `json:"max_redirects"`
	UserAgent                 string        `json:"user_agent"`
	FollowRedirects           bool          `json:"follow_redirects"`
	CheckDNS                  bool          `json:"check_dns"`
	CheckWHOIS                bool          `json:"check_whois"`
	MinContentLength          int           `json:"min_content_length"`
	MaxContentLength          int           `json:"max_content_length"`
}

// WebsiteValidationResult represents the result of website validation
type WebsiteValidationResult struct {
	URL               string                  `json:"url"`
	IsValid           bool                    `json:"is_valid"`
	AuthenticityScore float64                 `json:"authenticity_score"`
	TrafficAnalysis   *TrafficAnalysisResult  `json:"traffic_analysis"`
	DomainReputation  *DomainReputationResult `json:"domain_reputation"`
	SSLValidation     *SSLValidationResult    `json:"ssl_validation"`
	ContentQuality    *ContentQualityResult   `json:"content_quality"`
	OverallScore      float64                 `json:"overall_score"`
	Warnings          []string                `json:"warnings"`
	Errors            []string                `json:"errors"`
	ValidationTime    time.Duration           `json:"validation_time"`
	LastChecked       time.Time               `json:"last_checked"`
}

// AuthenticityChecker validates website authenticity
type AuthenticityChecker struct {
	suspiciousPatterns []*regexp.Regexp
	legitimatePatterns []*regexp.Regexp
	blacklistDomains   map[string]bool
	whitelistDomains   map[string]bool
	mu                 sync.RWMutex
}

// TrafficAnalyzer analyzes website traffic patterns
type TrafficAnalyzer struct {
	botPatterns    []*regexp.Regexp
	humanPatterns  []*regexp.Regexp
	suspiciousIPs  map[string]bool
	geographicData map[string]string
	mu             sync.RWMutex
}

// TrafficAnalysisResult represents traffic analysis results
type TrafficAnalysisResult struct {
	IsBotTraffic           bool               `json:"is_bot_traffic"`
	HumanTrafficPercentage float64            `json:"human_traffic_percentage"`
	SuspiciousActivity     bool               `json:"suspicious_activity"`
	GeographicDistribution map[string]float64 `json:"geographic_distribution"`
	TrafficVolume          string             `json:"traffic_volume"` // low, medium, high
	ResponseTime           float64            `json:"response_time"`
	Uptime                 float64            `json:"uptime"`
}

// DomainReputationChecker checks domain reputation
type DomainReputationChecker struct {
	reputationAPIs   map[string]ReputationAPI
	blacklistSources []string
	whitelistSources []string
	reputationCache  map[string]DomainReputationResult
	cacheExpiry      time.Duration
	mu               sync.RWMutex
}

// DomainReputationResult represents domain reputation results
type DomainReputationResult struct {
	Domain           string    `json:"domain"`
	Age              int       `json:"age_days"`
	ReputationScore  float64   `json:"reputation_score"`
	IsBlacklisted    bool      `json:"is_blacklisted"`
	IsWhitelisted    bool      `json:"is_whitelisted"`
	BlacklistSources []string  `json:"blacklist_sources"`
	WhitelistSources []string  `json:"whitelist_sources"`
	RegistrationDate time.Time `json:"registration_date"`
	ExpirationDate   time.Time `json:"expiration_date"`
	Registrar        string    `json:"registrar"`
	Country          string    `json:"country"`
	LastChecked      time.Time `json:"last_checked"`
}

// ReputationAPI represents a reputation checking API
type ReputationAPI struct {
	Name     string
	Endpoint string
	APIKey   string
	Timeout  time.Duration
}

// SSLValidator validates SSL certificates
type SSLValidator struct {
	trustedCAs      *x509.CertPool
	sslConfig       *tls.Config
	checkRevocation bool
	mu              sync.RWMutex
}

// SSLValidationResult represents SSL validation results
type SSLValidationResult struct {
	IsValid            bool      `json:"is_valid"`
	CertificateValid   bool      `json:"certificate_valid"`
	ChainValid         bool      `json:"chain_valid"`
	RevocationValid    bool      `json:"revocation_valid"`
	Issuer             string    `json:"issuer"`
	Subject            string    `json:"subject"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
	SerialNumber       string    `json:"serial_number"`
	SignatureAlgorithm string    `json:"signature_algorithm"`
	KeyUsage           []string  `json:"key_usage"`
	ExtendedKeyUsage   []string  `json:"extended_key_usage"`
	Warnings           []string  `json:"warnings"`
	Errors             []string  `json:"errors"`
}

// ContentQualityAssessor assesses website content quality
type ContentQualityAssessor struct {
	qualityMetrics    map[string]QualityMetric
	spamPatterns      []*regexp.Regexp
	legitimateContent []*regexp.Regexp
	mu                sync.RWMutex
}

// ContentQualityResult represents content quality assessment results
type ContentQualityResult struct {
	OverallScore     float64            `json:"overall_score"`
	ReadabilityScore float64            `json:"readability_score"`
	SpamScore        float64            `json:"spam_score"`
	ContentLength    int                `json:"content_length"`
	WordCount        int                `json:"word_count"`
	UniqueWords      int                `json:"unique_words"`
	GrammarErrors    int                `json:"grammar_errors"`
	SpellingErrors   int                `json:"spelling_errors"`
	LinkCount        int                `json:"link_count"`
	ImageCount       int                `json:"image_count"`
	IsSpam           bool               `json:"is_spam"`
	QualityMetrics   map[string]float64 `json:"quality_metrics"`
	Warnings         []string           `json:"warnings"`
}

// QualityMetric represents a content quality metric
type QualityMetric struct {
	Name        string
	Weight      float64
	Description string
	Calculator  func(content string) float64
}

// NewWebsiteValidator creates a new website validator
func NewWebsiteValidator() *WebsiteValidator {
	config := ValidatorConfig{
		EnableAuthenticityCheck:   true,
		EnableTrafficAnalysis:     true,
		EnableDomainReputation:    true,
		EnableSSLValidation:       true,
		EnableContentQualityCheck: true,
		Timeout:                   time.Second * 30,
		MaxRedirects:              10,
		UserAgent:                 "Mozilla/5.0 (compatible; KYB-Validator/1.0)",
		FollowRedirects:           true,
		CheckDNS:                  true,
		CheckWHOIS:                true,
		MinContentLength:          100,
		MaxContentLength:          1000000,
	}

	return &WebsiteValidator{
		authenticityChecker: NewAuthenticityChecker(),
		trafficAnalyzer:     NewTrafficAnalyzer(),
		domainReputation:    NewDomainReputationChecker(),
		sslValidator:        NewSSLValidator(),
		contentQuality:      NewContentQualityAssessor(),
		config:              config,
	}
}

// NewAuthenticityChecker creates a new authenticity checker
func NewAuthenticityChecker() *AuthenticityChecker {
	return &AuthenticityChecker{
		suspiciousPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)phishing|scam|fraud|fake`),
			regexp.MustCompile(`(?i)click\s*here\s*to\s*claim`),
			regexp.MustCompile(`(?i)urgent\s*action\s*required`),
			regexp.MustCompile(`(?i)limited\s*time\s*offer`),
			regexp.MustCompile(`(?i)you\s*have\s*won`),
		},
		legitimatePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)about\s*us|contact|privacy\s*policy`),
			regexp.MustCompile(`(?i)terms\s*of\s*service|legal`),
			regexp.MustCompile(`(?i)company|business|corporate`),
		},
		blacklistDomains: make(map[string]bool),
		whitelistDomains: make(map[string]bool),
	}
}

// NewTrafficAnalyzer creates a new traffic analyzer
func NewTrafficAnalyzer() *TrafficAnalyzer {
	return &TrafficAnalyzer{
		botPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)bot|crawler|spider`),
			regexp.MustCompile(`(?i)automated|script`),
		},
		humanPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)user|human|browser`),
		},
		suspiciousIPs:  make(map[string]bool),
		geographicData: make(map[string]string),
	}
}

// NewDomainReputationChecker creates a new domain reputation checker
func NewDomainReputationChecker() *DomainReputationChecker {
	return &DomainReputationChecker{
		reputationAPIs: map[string]ReputationAPI{
			"virustotal": {
				Name:     "VirusTotal",
				Endpoint: "https://www.virustotal.com/vtapi/v2/url/report",
				Timeout:  time.Second * 30,
			},
			"phishtank": {
				Name:     "PhishTank",
				Endpoint: "https://checkurl.phishtank.com/checkurl/",
				Timeout:  time.Second * 30,
			},
		},
		blacklistSources: []string{
			"https://urlhaus.abuse.ch/downloads/text/",
			"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		},
		whitelistSources: []string{
			"https://raw.githubusercontent.com/anudeepND/blacklist/master/adslist.txt",
		},
		reputationCache: make(map[string]DomainReputationResult),
		cacheExpiry:     time.Hour * 24,
	}
}

// NewSSLValidator creates a new SSL validator
func NewSSLValidator() *SSLValidator {
	return &SSLValidator{
		trustedCAs:      x509.NewCertPool(),
		sslConfig:       &tls.Config{},
		checkRevocation: true,
	}
}

// NewContentQualityAssessor creates a new content quality assessor
func NewContentQualityAssessor() *ContentQualityAssessor {
	return &ContentQualityAssessor{
		qualityMetrics: map[string]QualityMetric{
			"readability": {
				Name:        "Readability",
				Weight:      0.3,
				Description: "Text readability score",
				Calculator:  calculateReadability,
			},
			"content_length": {
				Name:        "Content Length",
				Weight:      0.2,
				Description: "Content length score",
				Calculator:  calculateContentLength,
			},
			"spam_detection": {
				Name:        "Spam Detection",
				Weight:      0.3,
				Description: "Spam detection score",
				Calculator:  calculateSpamScore,
			},
			"grammar": {
				Name:        "Grammar",
				Weight:      0.2,
				Description: "Grammar and spelling score",
				Calculator:  calculateGrammarScore,
			},
		},
		spamPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)buy\s*now|act\s*now|limited\s*time`),
			regexp.MustCompile(`(?i)click\s*here|free\s*money|earn\s*money`),
			regexp.MustCompile(`(?i)urgent|immediate|instant`),
		},
		legitimateContent: []*regexp.Regexp{
			regexp.MustCompile(`(?i)about|contact|services|products`),
			regexp.MustCompile(`(?i)company|business|corporate`),
		},
	}
}

// ValidateWebsite performs comprehensive website validation
func (wv *WebsiteValidator) ValidateWebsite(urlStr string) (*WebsiteValidationResult, error) {
	start := time.Now()

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	result := &WebsiteValidationResult{
		URL:         urlStr,
		LastChecked: time.Now(),
	}

	// Perform authenticity check
	if wv.config.EnableAuthenticityCheck {
		authenticityScore, err := wv.authenticityChecker.CheckAuthenticity(urlStr)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Authenticity check failed: %v", err))
		} else {
			result.AuthenticityScore = authenticityScore
		}
	}

	// Perform traffic analysis
	if wv.config.EnableTrafficAnalysis {
		trafficResult, err := wv.trafficAnalyzer.AnalyzeTraffic(urlStr)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Traffic analysis failed: %v", err))
		} else {
			result.TrafficAnalysis = trafficResult
		}
	}

	// Check domain reputation
	if wv.config.EnableDomainReputation {
		reputationResult, err := wv.domainReputation.CheckReputation(parsedURL.Host)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Domain reputation check failed: %v", err))
		} else {
			result.DomainReputation = &reputationResult
		}
	}

	// Validate SSL certificate
	if wv.config.EnableSSLValidation {
		sslResult, err := wv.sslValidator.ValidateSSL(urlStr)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("SSL validation failed: %v", err))
		} else {
			result.SSLValidation = &sslResult
		}
	}

	// Assess content quality
	if wv.config.EnableContentQualityCheck {
		contentResult, err := wv.contentQuality.AssessQuality(urlStr)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Content quality assessment failed: %v", err))
		} else {
			result.ContentQuality = contentResult
		}
	}

	// Calculate overall score
	result.OverallScore = wv.calculateOverallScore(result)
	result.IsValid = result.OverallScore >= 0.7
	result.ValidationTime = time.Since(start)

	return result, nil
}

// CheckAuthenticity checks website authenticity
func (ac *AuthenticityChecker) CheckAuthenticity(urlStr string) (float64, error) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Fetch website content
	content, err := ac.fetchContent(urlStr)
	if err != nil {
		return 0.0, fmt.Errorf("failed to fetch content: %w", err)
	}

	score := 1.0

	// Check for suspicious patterns
	for _, pattern := range ac.suspiciousPatterns {
		if pattern.MatchString(content) {
			score -= 0.2
		}
	}

	// Check for legitimate patterns
	for _, pattern := range ac.legitimatePatterns {
		if pattern.MatchString(content) {
			score += 0.1
		}
	}

	// Check domain blacklist/whitelist
	parsedURL, _ := url.Parse(urlStr)
	if ac.blacklistDomains[parsedURL.Host] {
		score = 0.0
	} else if ac.whitelistDomains[parsedURL.Host] {
		score = 1.0
	}

	// Ensure score is between 0 and 1
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}

	return score, nil
}

// AnalyzeTraffic analyzes website traffic patterns
func (ta *TrafficAnalyzer) AnalyzeTraffic(urlStr string) (*TrafficAnalysisResult, error) {
	ta.mu.RLock()
	defer ta.mu.RUnlock()

	result := &TrafficAnalysisResult{
		GeographicDistribution: make(map[string]float64),
	}

	// Simulate traffic analysis (in a real implementation, this would use actual traffic data)
	result.IsBotTraffic = ta.detectBotTraffic(urlStr)
	result.HumanTrafficPercentage = ta.calculateHumanTraffic(urlStr)
	result.SuspiciousActivity = ta.detectSuspiciousActivity(urlStr)
	result.TrafficVolume = ta.assessTrafficVolume(urlStr)
	result.ResponseTime = ta.measureResponseTime(urlStr)
	result.Uptime = ta.calculateUptime(urlStr)

	return result, nil
}

// CheckReputation checks domain reputation
func (drc *DomainReputationChecker) CheckReputation(domain string) (DomainReputationResult, error) {
	drc.mu.Lock()
	defer drc.mu.Unlock()

	// Check cache first
	if cached, exists := drc.reputationCache[domain]; exists {
		if time.Since(cached.LastChecked) < drc.cacheExpiry {
			return cached, nil
		}
	}

	result := DomainReputationResult{
		Domain:           domain,
		LastChecked:      time.Now(),
		BlacklistSources: []string{},
		WhitelistSources: []string{},
	}

	// Check domain age
	result.Age = drc.getDomainAge(domain)

	// Check blacklists
	result.IsBlacklisted = drc.checkBlacklists(domain, &result.BlacklistSources)

	// Check whitelists
	result.IsWhitelisted = drc.checkWhitelists(domain, &result.WhitelistSources)

	// Calculate reputation score
	result.ReputationScore = drc.calculateReputationScore(result)

	// Cache the result
	drc.reputationCache[domain] = result

	return result, nil
}

// ValidateSSL validates SSL certificate
func (sv *SSLValidator) ValidateSSL(urlStr string) (SSLValidationResult, error) {
	sv.mu.RLock()
	defer sv.mu.RUnlock()

	result := SSLValidationResult{
		Warnings: []string{},
		Errors:   []string{},
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid URL: %v", err))
		return result, err
	}

	// Ensure HTTPS
	if parsedURL.Scheme != "https" {
		result.Errors = append(result.Errors, "URL is not HTTPS")
		return result, nil
	}

	// Connect to server
	conn, err := tls.Dial("tcp", parsedURL.Host+":443", sv.sslConfig)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("TLS connection failed: %v", err))
		return result, err
	}
	defer conn.Close()

	// Get certificate
	cert := conn.ConnectionState().PeerCertificates[0]
	result.Subject = cert.Subject.CommonName
	result.Issuer = cert.Issuer.CommonName
	result.ValidFrom = cert.NotBefore
	result.ValidUntil = cert.NotAfter
	result.SerialNumber = cert.SerialNumber.String()
	result.SignatureAlgorithm = cert.SignatureAlgorithm.String()

	// Check certificate validity
	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		result.CertificateValid = false
		result.Errors = append(result.Errors, "Certificate is not valid")
	} else {
		result.CertificateValid = true
	}

	// Check certificate chain
	if len(conn.ConnectionState().PeerCertificates) > 1 {
		result.ChainValid = true
	} else {
		result.ChainValid = false
		result.Warnings = append(result.Warnings, "Certificate chain validation failed")
	}

	// Overall SSL validity
	result.IsValid = result.CertificateValid && result.ChainValid

	return result, nil
}

// AssessQuality assesses content quality
func (cqa *ContentQualityAssessor) AssessQuality(urlStr string) (*ContentQualityResult, error) {
	cqa.mu.RLock()
	defer cqa.mu.RUnlock()

	result := &ContentQualityResult{
		QualityMetrics: make(map[string]float64),
		Warnings:       []string{},
	}

	// Fetch content
	content, err := cqa.fetchContent(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}

	// Calculate metrics
	for name, metric := range cqa.qualityMetrics {
		score := metric.Calculator(content)
		result.QualityMetrics[name] = score
	}

	// Calculate overall score
	result.OverallScore = cqa.calculateOverallScore(result.QualityMetrics)

	// Analyze content
	result.ContentLength = len(content)
	result.WordCount = cqa.countWords(content)
	result.UniqueWords = cqa.countUniqueWords(content)
	result.LinkCount = cqa.countLinks(content)
	result.ImageCount = cqa.countImages(content)

	// Check for spam
	result.IsSpam = cqa.detectSpam(content)
	if result.IsSpam {
		result.SpamScore = 0.0
	} else {
		result.SpamScore = 1.0
	}

	return result, nil
}

// Helper methods for AuthenticityChecker
func (ac *AuthenticityChecker) fetchContent(urlStr string) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Helper methods for TrafficAnalyzer
func (ta *TrafficAnalyzer) detectBotTraffic(urlStr string) bool {
	// Simulate bot detection
	return false
}

func (ta *TrafficAnalyzer) calculateHumanTraffic(urlStr string) float64 {
	// Simulate human traffic calculation
	return 85.0
}

func (ta *TrafficAnalyzer) detectSuspiciousActivity(urlStr string) bool {
	// Simulate suspicious activity detection
	return false
}

func (ta *TrafficAnalyzer) assessTrafficVolume(urlStr string) string {
	// Simulate traffic volume assessment
	return "medium"
}

func (ta *TrafficAnalyzer) measureResponseTime(urlStr string) float64 {
	// Simulate response time measurement
	return 150.0
}

func (ta *TrafficAnalyzer) calculateUptime(urlStr string) float64 {
	// Simulate uptime calculation
	return 99.5
}

// Helper methods for DomainReputationChecker
func (drc *DomainReputationChecker) getDomainAge(domain string) int {
	// Simulate domain age calculation
	return 365
}

func (drc *DomainReputationChecker) checkBlacklists(domain string, sources *[]string) bool {
	// Simulate blacklist checking
	return false
}

func (drc *DomainReputationChecker) checkWhitelists(domain string, sources *[]string) bool {
	// Simulate whitelist checking
	return false
}

func (drc *DomainReputationChecker) calculateReputationScore(result DomainReputationResult) float64 {
	score := 1.0

	if result.IsBlacklisted {
		score = 0.0
	} else if result.IsWhitelisted {
		score = 1.0
	} else {
		// Calculate based on age and other factors
		if result.Age > 365 {
			score += 0.2
		}
	}

	return score
}

// Helper methods for ContentQualityAssessor
func (cqa *ContentQualityAssessor) fetchContent(urlStr string) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (cqa *ContentQualityAssessor) calculateOverallScore(metrics map[string]float64) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	for name, score := range metrics {
		if metric, exists := cqa.qualityMetrics[name]; exists {
			totalScore += score * metric.Weight
			totalWeight += metric.Weight
		}
	}

	if totalWeight > 0 {
		return totalScore / totalWeight
	}
	return 0.0
}

func (cqa *ContentQualityAssessor) countWords(content string) int {
	words := strings.Fields(content)
	return len(words)
}

func (cqa *ContentQualityAssessor) countUniqueWords(content string) int {
	words := strings.Fields(content)
	unique := make(map[string]bool)
	for _, word := range words {
		unique[strings.ToLower(word)] = true
	}
	return len(unique)
}

func (cqa *ContentQualityAssessor) countLinks(content string) int {
	linkPattern := regexp.MustCompile(`<a[^>]+href=`)
	matches := linkPattern.FindAllString(content, -1)
	return len(matches)
}

func (cqa *ContentQualityAssessor) countImages(content string) int {
	imgPattern := regexp.MustCompile(`<img[^>]+>`)
	matches := imgPattern.FindAllString(content, -1)
	return len(matches)
}

func (cqa *ContentQualityAssessor) detectSpam(content string) bool {
	for _, pattern := range cqa.spamPatterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}

// Quality metric calculators
func calculateReadability(content string) float64 {
	// Simple readability calculation
	words := strings.Fields(content)
	sentences := strings.Split(content, ".")

	if len(sentences) == 0 || len(words) == 0 {
		return 0.0
	}

	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))

	// Flesch Reading Ease approximation
	if avgWordsPerSentence <= 10 {
		return 1.0
	} else if avgWordsPerSentence <= 15 {
		return 0.8
	} else if avgWordsPerSentence <= 20 {
		return 0.6
	} else {
		return 0.4
	}
}

func calculateContentLength(content string) float64 {
	length := len(content)
	if length >= 1000 {
		return 1.0
	} else if length >= 500 {
		return 0.8
	} else if length >= 200 {
		return 0.6
	} else if length >= 100 {
		return 0.4
	} else {
		return 0.2
	}
}

func calculateSpamScore(content string) float64 {
	spamPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)buy\s*now|act\s*now|limited\s*time`),
		regexp.MustCompile(`(?i)click\s*here|free\s*money|earn\s*money`),
		regexp.MustCompile(`(?i)urgent|immediate|instant`),
	}

	spamCount := 0
	for _, pattern := range spamPatterns {
		if pattern.MatchString(content) {
			spamCount++
		}
	}

	if spamCount == 0 {
		return 1.0
	} else if spamCount == 1 {
		return 0.7
	} else if spamCount == 2 {
		return 0.4
	} else {
		return 0.1
	}
}

func calculateGrammarScore(content string) float64 {
	// Simple grammar check (in a real implementation, this would use a proper grammar checker)
	words := strings.Fields(content)
	if len(words) == 0 {
		return 0.0
	}

	// Simulate grammar score based on content length and structure
	if len(words) > 50 {
		return 0.9
	} else if len(words) > 20 {
		return 0.8
	} else if len(words) > 10 {
		return 0.7
	} else {
		return 0.6
	}
}

// Helper method for WebsiteValidator
func (wv *WebsiteValidator) calculateOverallScore(result *WebsiteValidationResult) float64 {
	score := 0.0
	weight := 0.0

	// Authenticity score (30% weight)
	if wv.config.EnableAuthenticityCheck {
		score += result.AuthenticityScore * 0.3
		weight += 0.3
	}

	// Domain reputation score (25% weight)
	if wv.config.EnableDomainReputation && result.DomainReputation != nil {
		score += result.DomainReputation.ReputationScore * 0.25
		weight += 0.25
	}

	// SSL validation score (20% weight)
	if wv.config.EnableSSLValidation && result.SSLValidation != nil {
		if result.SSLValidation.IsValid {
			score += 1.0 * 0.2
		}
		weight += 0.2
	}

	// Content quality score (25% weight)
	if wv.config.EnableContentQualityCheck && result.ContentQuality != nil {
		score += result.ContentQuality.OverallScore * 0.25
		weight += 0.25
	}

	if weight > 0 {
		return score / weight
	}
	return 0.0
}
