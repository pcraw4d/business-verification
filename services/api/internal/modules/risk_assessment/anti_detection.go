package risk_assessment

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AntiDetectionService provides protection against web scraping detection
type AntiDetectionService struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// User agent rotation
	userAgents []string
	lastUA     string
	uaIndex    int

	// Request patterns
	requestDelays map[string]time.Time
	delayMutex    sync.RWMutex

	// Proxy management
	proxies      []Proxy
	currentProxy *Proxy
	proxyIndex   int

	// Detection monitoring
	detectionEvents []DetectionEvent
	eventMutex      sync.RWMutex
}

// Proxy represents a proxy server configuration
type Proxy struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Protocol    string  // http, https, socks5
	Location    string  // geographic location
	Speed       int     // speed in ms
	Reliability float64 // reliability score 0-1
	LastUsed    time.Time
	FailCount   int
}

// DetectionEvent represents a detection event
type DetectionEvent struct {
	Timestamp   time.Time
	URL         string
	EventType   DetectionEventType
	Severity    DetectionSeverity
	Description string
	Headers     map[string]string
	Response    *http.Response
	IP          string
	UserAgent   string
}

// DetectionEventType represents the type of detection event
type DetectionEventType string

const (
	DetectionEventBlocked     DetectionEventType = "blocked"
	DetectionEventCaptcha     DetectionEventType = "captcha"
	DetectionEventRateLimited DetectionEventType = "rate_limited"
	DetectionEventSuspicious  DetectionEventType = "suspicious"
	DetectionEventRedirected  DetectionEventType = "redirected"
)

// DetectionSeverity represents the severity of a detection event
type DetectionSeverity string

const (
	DetectionSeverityLow      DetectionSeverity = "low"
	DetectionSeverityMedium   DetectionSeverity = "medium"
	DetectionSeverityHigh     DetectionSeverity = "high"
	DetectionSeverityCritical DetectionSeverity = "critical"
)

// AntiDetectionConfig contains anti-detection configuration
type AntiDetectionConfig struct {
	// User agent rotation
	UserAgentRotationEnabled bool
	UserAgentPool            []string
	UserAgentRotationDelay   time.Duration

	// Request delays
	RequestDelayEnabled bool
	MinDelay            time.Duration
	MaxDelay            time.Duration
	DelayPerDomain      map[string]time.Duration

	// Proxy configuration
	ProxyEnabled          bool
	ProxyPool             []Proxy
	ProxyRotationEnabled  bool
	ProxyRotationInterval time.Duration
	ProxyFailThreshold    int

	// Header customization
	CustomHeadersEnabled bool
	CustomHeaders        map[string]string
	HeaderRandomization  bool

	// Detection monitoring
	DetectionMonitoringEnabled bool
	MaxDetectionEvents         int
	DetectionAlertThreshold    int
}

// NewAntiDetectionService creates a new anti-detection service
func NewAntiDetectionService(config *RiskAssessmentConfig, logger *zap.Logger) *AntiDetectionService {
	if logger == nil {
		logger = zap.NewNop()
	}

	// Default user agents
	defaultUserAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.0.0",
	}

	// Note: Custom headers are configured via AntiDetectionConfig

	// Default proxies (placeholder - in production, use real proxy service)
	defaultProxies := []Proxy{
		{
			Host:        "proxy1.example.com",
			Port:        8080,
			Protocol:    "http",
			Location:    "US",
			Speed:       100,
			Reliability: 0.95,
		},
		{
			Host:        "proxy2.example.com",
			Port:        8080,
			Protocol:    "https",
			Location:    "EU",
			Speed:       150,
			Reliability: 0.90,
		},
	}

	ads := &AntiDetectionService{
		config:          config,
		logger:          logger,
		userAgents:      defaultUserAgents,
		requestDelays:   make(map[string]time.Time),
		proxies:         defaultProxies,
		detectionEvents: make([]DetectionEvent, 0),
	}

	// Start background tasks
	go ads.startBackgroundTasks()

	return ads
}

// CreateHTTPClient creates an HTTP client with anti-detection features
func (ads *AntiDetectionService) CreateHTTPClient(ctx context.Context, targetURL string) (*http.Client, error) {
	ads.mu.RLock()
	defer ads.mu.RUnlock()

	// Create transport with custom settings
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// Add proxy if enabled
	if ads.config.AntiDetectionConfig.ProxyEnabled && len(ads.proxies) > 0 {
		proxy := ads.selectProxy()
		if proxy != nil {
			proxyURL := fmt.Sprintf("%s://%s:%d", proxy.Protocol, proxy.Host, proxy.Port)
			if proxy.Username != "" && proxy.Password != "" {
				proxyURL = fmt.Sprintf("%s://%s:%s@%s:%d", proxy.Protocol, proxy.Username, proxy.Password, proxy.Host, proxy.Port)
			}

			parsedURL, err := url.Parse(proxyURL)
			if err != nil {
				ads.logger.Warn("Failed to parse proxy URL", zap.String("proxy", proxyURL), zap.Error(err))
			} else {
				transport.Proxy = http.ProxyURL(parsedURL)
			}
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client, nil
}

// PrepareRequest prepares an HTTP request with anti-detection headers
func (ads *AntiDetectionService) PrepareRequest(ctx context.Context, req *http.Request, targetURL string) error {
	ads.mu.RLock()
	defer ads.mu.RUnlock()

	// Apply request delay
	if err := ads.applyRequestDelay(ctx, targetURL); err != nil {
		return fmt.Errorf("failed to apply request delay: %w", err)
	}

	// Set user agent
	if ads.config.AntiDetectionConfig.UserAgentRotationEnabled {
		req.Header.Set("User-Agent", ads.getNextUserAgent())
	} else {
		req.Header.Set("User-Agent", ads.userAgents[0])
	}

	// Add custom headers
	if ads.config.AntiDetectionConfig.CustomHeadersEnabled {
		for key, value := range ads.config.AntiDetectionConfig.CustomHeaders {
			if ads.config.AntiDetectionConfig.HeaderRandomization {
				value = ads.randomizeHeaderValue(key, value)
			}
			req.Header.Set(key, value)
		}
	}

	// Add referer if not present
	if req.Header.Get("Referer") == "" {
		parsedURL, err := url.Parse(targetURL)
		if err == nil {
			req.Header.Set("Referer", fmt.Sprintf("%s://%s/", parsedURL.Scheme, parsedURL.Host))
		}
	}

	// Add accept headers
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	}

	return nil
}

// MonitorResponse monitors the response for detection indicators
func (ads *AntiDetectionService) MonitorResponse(ctx context.Context, req *http.Request, resp *http.Response, targetURL string) error {
	ads.eventMutex.Lock()
	defer ads.eventMutex.Unlock()

	// Check for common detection indicators
	detectionEvent := ads.analyzeResponse(req, resp, targetURL)
	if detectionEvent != nil {
		ads.detectionEvents = append(ads.detectionEvents, *detectionEvent)

		// Keep only recent events
		if len(ads.detectionEvents) > ads.config.AntiDetectionConfig.MaxDetectionEvents {
			ads.detectionEvents = ads.detectionEvents[1:]
		}

		// Log detection event
		ads.logger.Warn("Detection event detected",
			zap.String("url", targetURL),
			zap.String("event_type", string(detectionEvent.EventType)),
			zap.String("severity", string(detectionEvent.Severity)),
			zap.String("description", detectionEvent.Description))

		// Trigger alerts if threshold exceeded
		if len(ads.detectionEvents) >= ads.config.AntiDetectionConfig.DetectionAlertThreshold {
			ads.triggerDetectionAlert(ctx, detectionEvent)
		}
	}

	return nil
}

// GetDetectionReport returns a report of recent detection events
func (ads *AntiDetectionService) GetDetectionReport() *DetectionReport {
	ads.eventMutex.RLock()
	defer ads.eventMutex.RUnlock()

	report := &DetectionReport{
		TotalEvents:      len(ads.detectionEvents),
		EventsByType:     make(map[DetectionEventType]int),
		EventsBySeverity: make(map[DetectionSeverity]int),
		RecentEvents:     make([]DetectionEvent, 0),
		ReportTimestamp:  time.Now(),
	}

	// Count events by type and severity
	for _, event := range ads.detectionEvents {
		report.EventsByType[event.EventType]++
		report.EventsBySeverity[event.Severity]++
	}

	// Get recent events (last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, event := range ads.detectionEvents {
		if event.Timestamp.After(cutoff) {
			report.RecentEvents = append(report.RecentEvents, event)
		}
	}

	// Calculate risk score
	report.RiskScore = ads.calculateDetectionRiskScore()

	return report
}

// DetectionReport contains detection monitoring report
type DetectionReport struct {
	TotalEvents      int                        `json:"total_events"`
	EventsByType     map[DetectionEventType]int `json:"events_by_type"`
	EventsBySeverity map[DetectionSeverity]int  `json:"events_by_severity"`
	RecentEvents     []DetectionEvent           `json:"recent_events"`
	RiskScore        float64                    `json:"risk_score"`
	ReportTimestamp  time.Time                  `json:"report_timestamp"`
}

// Helper methods

func (ads *AntiDetectionService) startBackgroundTasks() {
	// Proxy rotation task
	if ads.config.AntiDetectionConfig.ProxyRotationEnabled {
		go func() {
			ticker := time.NewTicker(ads.config.AntiDetectionConfig.ProxyRotationInterval)
			defer ticker.Stop()

			for range ticker.C {
				ads.rotateProxy()
			}
		}()
	}

	// Cleanup old detection events
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			ads.cleanupOldEvents()
		}
	}()
}

func (ads *AntiDetectionService) getNextUserAgent() string {
	if len(ads.userAgents) == 0 {
		return "Mozilla/5.0 (compatible; BusinessVerification/1.0)"
	}

	ads.uaIndex = (ads.uaIndex + 1) % len(ads.userAgents)
	return ads.userAgents[ads.uaIndex]
}

func (ads *AntiDetectionService) applyRequestDelay(ctx context.Context, targetURL string) error {
	if !ads.config.AntiDetectionConfig.RequestDelayEnabled {
		return nil
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	domain := parsedURL.Hostname()

	ads.delayMutex.Lock()
	lastRequest, exists := ads.requestDelays[domain]
	ads.delayMutex.Unlock()

	if exists {
		// Calculate delay based on domain-specific settings
		delay := ads.config.AntiDetectionConfig.MinDelay
		if domainDelay, ok := ads.config.AntiDetectionConfig.DelayPerDomain[domain]; ok {
			delay = domainDelay
		}

		// Add some randomization
		randomDelay := time.Duration(rand.Intn(int(ads.config.AntiDetectionConfig.MaxDelay - ads.config.AntiDetectionConfig.MinDelay)))
		totalDelay := delay + randomDelay

		timeSinceLast := time.Since(lastRequest)
		if timeSinceLast < totalDelay {
			waitTime := totalDelay - timeSinceLast
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				// Continue
			}
		}
	}

	// Update last request time
	ads.delayMutex.Lock()
	ads.requestDelays[domain] = time.Now()
	ads.delayMutex.Unlock()

	return nil
}

func (ads *AntiDetectionService) selectProxy() *Proxy {
	if len(ads.proxies) == 0 {
		return nil
	}

	// Simple round-robin selection
	ads.proxyIndex = (ads.proxyIndex + 1) % len(ads.proxies)
	return &ads.proxies[ads.proxyIndex]
}

func (ads *AntiDetectionService) rotateProxy() {
	ads.mu.Lock()
	defer ads.mu.Unlock()

	if len(ads.proxies) > 1 {
		ads.proxyIndex = (ads.proxyIndex + 1) % len(ads.proxies)
		ads.currentProxy = &ads.proxies[ads.proxyIndex]
		ads.logger.Info("Rotated to new proxy",
			zap.String("proxy", ads.currentProxy.Host),
			zap.String("location", ads.currentProxy.Location))
	}
}

func (ads *AntiDetectionService) randomizeHeaderValue(key, value string) string {
	// Add some randomization to header values
	switch strings.ToLower(key) {
	case "accept-language":
		languages := []string{
			"en-US,en;q=0.9",
			"en-US,en;q=0.8,es;q=0.7",
			"en-GB,en;q=0.9",
			"en-CA,en;q=0.9",
		}
		return languages[rand.Intn(len(languages))]
	case "accept-encoding":
		encodings := []string{
			"gzip, deflate, br",
			"gzip, deflate",
			"gzip",
		}
		return encodings[rand.Intn(len(encodings))]
	default:
		return value
	}
}

func (ads *AntiDetectionService) analyzeResponse(req *http.Request, resp *http.Response, targetURL string) *DetectionEvent {
	// Check status code
	if resp.StatusCode == 403 || resp.StatusCode == 429 {
		return &DetectionEvent{
			Timestamp:   time.Now(),
			URL:         targetURL,
			EventType:   DetectionEventBlocked,
			Severity:    DetectionSeverityHigh,
			Description: fmt.Sprintf("Request blocked with status %d", resp.StatusCode),
			Headers:     ads.extractHeaders(resp.Header),
			Response:    resp,
		}
	}

	// Check for captcha
	body, _ := ads.readResponseBody(resp)
	if strings.Contains(strings.ToLower(body), "captcha") ||
		strings.Contains(strings.ToLower(body), "recaptcha") ||
		strings.Contains(strings.ToLower(body), "cloudflare") {
		return &DetectionEvent{
			Timestamp:   time.Now(),
			URL:         targetURL,
			EventType:   DetectionEventCaptcha,
			Severity:    DetectionSeverityCritical,
			Description: "Captcha detected in response",
			Headers:     ads.extractHeaders(resp.Header),
			Response:    resp,
		}
	}

	// Check for suspicious redirects
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		parsedURL, err := url.Parse(targetURL)
		if err == nil && location != "" && !strings.Contains(location, parsedURL.Hostname()) {
			return &DetectionEvent{
				Timestamp:   time.Now(),
				URL:         targetURL,
				EventType:   DetectionEventRedirected,
				Severity:    DetectionSeverityMedium,
				Description: fmt.Sprintf("Suspicious redirect to %s", location),
				Headers:     ads.extractHeaders(resp.Header),
				Response:    resp,
			}
		}
	}

	// Check for rate limiting headers
	if resp.Header.Get("Retry-After") != "" ||
		resp.Header.Get("X-RateLimit-Remaining") == "0" {
		return &DetectionEvent{
			Timestamp:   time.Now(),
			URL:         targetURL,
			EventType:   DetectionEventRateLimited,
			Severity:    DetectionSeverityMedium,
			Description: "Rate limiting detected",
			Headers:     ads.extractHeaders(resp.Header),
			Response:    resp,
		}
	}

	return nil
}

func (ads *AntiDetectionService) extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

func (ads *AntiDetectionService) readResponseBody(resp *http.Response) (string, error) {
	// This is a simplified version - in production, you'd want to handle large responses
	// and different content types properly
	defer resp.Body.Close()

	// For now, just return a placeholder
	return "", nil
}

func (ads *AntiDetectionService) calculateDetectionRiskScore() float64 {
	if len(ads.detectionEvents) == 0 {
		return 0.0
	}

	// Calculate weighted risk score based on recent events
	score := 0.0
	weight := 1.0

	for i := len(ads.detectionEvents) - 1; i >= 0; i-- {
		event := ads.detectionEvents[i]

		// Weight by severity
		severityWeight := 1.0
		switch event.Severity {
		case DetectionSeverityLow:
			severityWeight = 0.25
		case DetectionSeverityMedium:
			severityWeight = 0.5
		case DetectionSeverityHigh:
			severityWeight = 0.75
		case DetectionSeverityCritical:
			severityWeight = 1.0
		}

		// Weight by recency (more recent events have higher weight)
		age := time.Since(event.Timestamp)
		recencyWeight := 1.0
		if age < time.Hour {
			recencyWeight = 1.0
		} else if age < 24*time.Hour {
			recencyWeight = 0.5
		} else {
			recencyWeight = 0.1
		}

		score += weight * severityWeight * recencyWeight
		weight *= 0.9 // Decay factor
	}

	// Normalize to 0-1 range
	return math.Min(score, 1.0)
}

func (ads *AntiDetectionService) triggerDetectionAlert(ctx context.Context, event *DetectionEvent) {
	ads.logger.Error("Detection alert triggered",
		zap.String("url", event.URL),
		zap.String("event_type", string(event.EventType)),
		zap.String("severity", string(event.Severity)),
		zap.String("description", event.Description))

	// In production, this would trigger notifications, alerts, etc.
}

func (ads *AntiDetectionService) cleanupOldEvents() {
	ads.eventMutex.Lock()
	defer ads.eventMutex.Unlock()

	// Remove events older than 7 days
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	newEvents := make([]DetectionEvent, 0)

	for _, event := range ads.detectionEvents {
		if event.Timestamp.After(cutoff) {
			newEvents = append(newEvents, event)
		}
	}

	ads.detectionEvents = newEvents
}
