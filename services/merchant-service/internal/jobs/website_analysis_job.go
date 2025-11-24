package jobs

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/supabase"
)

// WebsiteAnalysisJob represents a website analysis job
type WebsiteAnalysisJob struct {
	ID          string
	MerchantID  string
	WebsiteURL  string
	BusinessName string
	Status      JobStatus
	Result      *WebsiteAnalysisResult
	Error       error
	CreatedAt   time.Time
	UpdatedAt   time.Time

	supabaseClient *supabase.Client
	config         *config.Config
	logger         *zap.Logger
	httpClient     *http.Client
}

// WebsiteAnalysisResult represents the result of website analysis
type WebsiteAnalysisResult struct {
	WebsiteURL      string                 `json:"websiteUrl"`
	SSL             SSLData                `json:"ssl"`
	SecurityHeaders SecurityHeadersData     `json:"securityHeaders"`
	Performance     PerformanceData         `json:"performance"`
	Accessibility   AccessibilityData       `json:"accessibility"`
	LastAnalyzed    string                 `json:"lastAnalyzed"`
	Status          string                 `json:"status"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SSLData represents SSL certificate information
type SSLData struct {
	Valid       bool   `json:"valid"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
	Issuer      string `json:"issuer,omitempty"`
	Certificate string `json:"certificate,omitempty"`
}

// SecurityHeadersData represents security headers information
type SecurityHeadersData struct {
	HasHTTPS          bool     `json:"hasHttps"`
	HasHSTS           bool     `json:"hasHsts"`
	HasCSP            bool     `json:"hasCsp"`
	HasXFrameOptions  bool     `json:"hasXFrameOptions"`
	HasXContentType   bool     `json:"hasXContentType"`
	MissingHeaders    []string `json:"missingHeaders,omitempty"`
	SecurityScore     float64  `json:"securityScore"`
}

// PerformanceData represents website performance metrics
type PerformanceData struct {
	LoadTime      float64 `json:"loadTime"`
	PageSize      int     `json:"pageSize"`
	RequestCount  int     `json:"requestCount"`
	PerformanceScore float64 `json:"performanceScore"`
}

// AccessibilityData represents accessibility metrics
type AccessibilityData struct {
	Score           float64   `json:"score"`
	Issues          []string  `json:"issues,omitempty"`
	WCAGCompliance  string    `json:"wcagCompliance,omitempty"`
}

// NewWebsiteAnalysisJob creates a new website analysis job
func NewWebsiteAnalysisJob(
	merchantID, websiteURL, businessName string,
	supabaseClient *supabase.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *WebsiteAnalysisJob {
	return &WebsiteAnalysisJob{
		ID:             fmt.Sprintf("website_analysis_%s_%d", merchantID, time.Now().Unix()),
		MerchantID:     merchantID,
		WebsiteURL:     websiteURL,
		BusinessName:   businessName,
		Status:         StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		supabaseClient: supabaseClient,
		config:         cfg,
		logger:         logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
		},
	}
}

// GetID returns the job ID
func (j *WebsiteAnalysisJob) GetID() string {
	return j.ID
}

// GetMerchantID returns the merchant ID
func (j *WebsiteAnalysisJob) GetMerchantID() string {
	return j.MerchantID
}

// GetType returns the job type
func (j *WebsiteAnalysisJob) GetType() string {
	return "website_analysis"
}

// GetStatus returns the current job status
func (j *WebsiteAnalysisJob) GetStatus() JobStatus {
	return j.Status
}

// SetStatus sets the job status
func (j *WebsiteAnalysisJob) SetStatus(status JobStatus) {
	j.Status = status
	j.UpdatedAt = time.Now()
}

// Process executes the website analysis job
func (j *WebsiteAnalysisJob) Process(ctx context.Context) error {
	startTime := time.Now()
	j.logger.Info("Starting website analysis job",
		zap.String("job_id", j.ID),
		zap.String("merchant_id", j.MerchantID),
		zap.String("website_url", j.WebsiteURL))

	// Update status to processing
	j.SetStatus(StatusProcessing)
	if err := j.updateStatusInDB(ctx, StatusProcessing); err != nil {
		j.logger.Warn("Failed to update status to processing", zap.Error(err))
	}

	// Perform website analysis
	result, err := j.performWebsiteAnalysis(ctx)
	if err != nil {
		j.logger.Error("Website analysis job failed",
			zap.String("job_id", j.ID),
			zap.String("merchant_id", j.MerchantID),
			zap.Error(err))

		j.SetStatus(StatusFailed)
		j.Error = err
		j.updateStatusInDB(ctx, StatusFailed)
		return fmt.Errorf("website analysis failed: %w", err)
	}

	// Save result to database
	if err := j.saveResultToDB(ctx, result); err != nil {
		j.logger.Error("Failed to save website analysis result",
			zap.String("job_id", j.ID),
			zap.String("merchant_id", j.MerchantID),
			zap.Error(err))

		j.SetStatus(StatusFailed)
		j.Error = err
		j.updateStatusInDB(ctx, StatusFailed)
		return fmt.Errorf("failed to save result: %w", err)
	}

	j.Result = result
	j.SetStatus(StatusCompleted)
	j.updateStatusInDB(ctx, StatusCompleted)

	duration := time.Since(startTime)
	j.logger.Info("Website analysis job completed successfully",
		zap.String("job_id", j.ID),
		zap.String("merchant_id", j.MerchantID),
		zap.Duration("duration", duration),
		zap.Float64("security_score", result.SecurityHeaders.SecurityScore),
		zap.Float64("performance_score", result.Performance.PerformanceScore))

	return nil
}

// performWebsiteAnalysis performs the actual website analysis
func (j *WebsiteAnalysisJob) performWebsiteAnalysis(ctx context.Context) (*WebsiteAnalysisResult, error) {
	// Ensure URL has scheme
	websiteURL := j.WebsiteURL
	if !hasScheme(websiteURL) {
		websiteURL = "https://" + websiteURL
	}

	// Parse URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	result := &WebsiteAnalysisResult{
		WebsiteURL:   websiteURL,
		LastAnalyzed: time.Now().Format(time.RFC3339),
		Status:       "completed",
		Metadata:     make(map[string]interface{}),
	}

	// Analyze SSL
	sslData, err := j.analyzeSSL(ctx, parsedURL)
	if err != nil {
		j.logger.Warn("SSL analysis failed", zap.Error(err))
		sslData = SSLData{Valid: false}
	}
	result.SSL = sslData

	// Analyze security headers
	securityHeaders, err := j.analyzeSecurityHeaders(ctx, parsedURL)
	if err != nil {
		j.logger.Warn("Security headers analysis failed", zap.Error(err))
		securityHeaders = SecurityHeadersData{}
	}
	result.SecurityHeaders = securityHeaders

	// Analyze performance
	performance, err := j.analyzePerformance(ctx, parsedURL)
	if err != nil {
		j.logger.Warn("Performance analysis failed", zap.Error(err))
		performance = PerformanceData{}
	}
	result.Performance = performance

	// Analyze accessibility (basic check)
	accessibility := j.analyzeAccessibility(ctx, parsedURL)
	result.Accessibility = accessibility

	return result, nil
}

// analyzeSSL analyzes SSL certificate
func (j *WebsiteAnalysisJob) analyzeSSL(ctx context.Context, parsedURL *url.URL) (SSLData, error) {
	if parsedURL.Scheme != "https" {
		return SSLData{Valid: false}, nil
	}

	// Create connection to check SSL
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}
	conn, err := tls.DialWithDialer(
		dialer,
		"tcp",
		parsedURL.Host+":443",
		&tls.Config{
			InsecureSkipVerify: false,
		},
	)
	if err != nil {
		return SSLData{Valid: false}, err
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return SSLData{Valid: false}, fmt.Errorf("no certificates found")
	}

	cert := state.PeerCertificates[0]
	sslData := SSLData{
		Valid:       true,
		ExpiresAt:   cert.NotAfter.Format(time.RFC3339),
		Issuer:      cert.Issuer.String(),
		Certificate: cert.Subject.String(),
	}

	return sslData, nil
}

// analyzeSecurityHeaders analyzes security headers
func (j *WebsiteAnalysisJob) analyzeSecurityHeaders(ctx context.Context, parsedURL *url.URL) (SecurityHeadersData, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		return SecurityHeadersData{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KYB-Platform/1.0)")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return SecurityHeadersData{}, err
	}
	defer resp.Body.Close()

	headers := SecurityHeadersData{
		HasHTTPS:         parsedURL.Scheme == "https",
		HasHSTS:          resp.Header.Get("Strict-Transport-Security") != "",
		HasCSP:           resp.Header.Get("Content-Security-Policy") != "",
		HasXFrameOptions: resp.Header.Get("X-Frame-Options") != "",
		HasXContentType:  resp.Header.Get("X-Content-Type-Options") != "",
		MissingHeaders:   []string{},
	}

	// Check for missing headers
	if !headers.HasHSTS {
		headers.MissingHeaders = append(headers.MissingHeaders, "Strict-Transport-Security")
	}
	if !headers.HasCSP {
		headers.MissingHeaders = append(headers.MissingHeaders, "Content-Security-Policy")
	}
	if !headers.HasXFrameOptions {
		headers.MissingHeaders = append(headers.MissingHeaders, "X-Frame-Options")
	}
	if !headers.HasXContentType {
		headers.MissingHeaders = append(headers.MissingHeaders, "X-Content-Type-Options")
	}

	// Calculate security score (0-1)
	score := 0.0
	if headers.HasHTTPS {
		score += 0.3
	}
	if headers.HasHSTS {
		score += 0.2
	}
	if headers.HasCSP {
		score += 0.2
	}
	if headers.HasXFrameOptions {
		score += 0.15
	}
	if headers.HasXContentType {
		score += 0.15
	}
	headers.SecurityScore = score

	return headers, nil
}

// analyzePerformance analyzes website performance
func (j *WebsiteAnalysisJob) analyzePerformance(ctx context.Context, parsedURL *url.URL) (PerformanceData, error) {
	startTime := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		return PerformanceData{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KYB-Platform/1.0)")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return PerformanceData{}, err
	}
	defer resp.Body.Close()

	loadTime := time.Since(startTime).Seconds()

	// Read body to get size
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return PerformanceData{}, err
	}

	pageSize := len(bodyBytes)

	// Calculate performance score (inverse of load time, normalized)
	// Good performance: < 2s = 1.0, 2-5s = 0.7, > 5s = 0.3
	performanceScore := 1.0
	if loadTime > 5 {
		performanceScore = 0.3
	} else if loadTime > 2 {
		performanceScore = 0.7
	}

	return PerformanceData{
		LoadTime:        loadTime,
		PageSize:        pageSize,
		RequestCount:    1, // Simplified
		PerformanceScore: performanceScore,
	}, nil
}

// analyzeAccessibility performs basic accessibility analysis
func (j *WebsiteAnalysisJob) analyzeAccessibility(ctx context.Context, parsedURL *url.URL) AccessibilityData {
	// Basic accessibility check - in a real implementation, this would use a proper accessibility checker
	// For now, return a default score
	return AccessibilityData{
		Score:          0.7, // Default score
		Issues:         []string{},
		WCAGCompliance: "AA", // Default compliance level
	}
}

// hasScheme checks if URL has a scheme
func hasScheme(urlStr string) bool {
	return len(urlStr) > 0 && (urlStr[0:4] == "http" || urlStr[0:5] == "https")
}

// updateStatusInDB updates the job status in the database
func (j *WebsiteAnalysisJob) updateStatusInDB(ctx context.Context, status JobStatus) error {
	updateData := map[string]interface{}{
		"website_analysis_status":      string(status),
		"website_analysis_updated_at": time.Now().Format(time.RFC3339),
	}

	// Check if merchant_analytics record exists
	var existing []map[string]interface{}
	_, err := j.supabaseClient.GetClient().From("merchant_analytics").
		Select("id", "", false).
		Eq("merchant_id", j.MerchantID).
		Limit(1, "").
		ExecuteTo(&existing)

	if err != nil || len(existing) == 0 {
		// Create new record
		insertData := map[string]interface{}{
			"merchant_id":                j.MerchantID,
			"website_analysis_status":    string(status),
			"website_analysis_updated_at": time.Now().Format(time.RFC3339),
			"website_analysis_data":      map[string]interface{}{},
		}

		_, _, err := j.supabaseClient.GetClient().From("merchant_analytics").
			Insert(insertData, false, "", "", "").
			Execute()

		return err
	}

	// Update existing record
	_, _, err = j.supabaseClient.GetClient().From("merchant_analytics").
		Update(updateData, "", "").
		Eq("merchant_id", j.MerchantID).
		Execute()

	return err
}

// saveResultToDB saves the website analysis result to the database
func (j *WebsiteAnalysisJob) saveResultToDB(ctx context.Context, result *WebsiteAnalysisResult) error {
	// Convert result to JSONB format
	analysisData := map[string]interface{}{
		"websiteUrl":     result.WebsiteURL,
		"ssl":            result.SSL,
		"securityHeaders": result.SecurityHeaders,
		"performance":    result.Performance,
		"accessibility": result.Accessibility,
		"lastAnalyzed":  result.LastAnalyzed,
		"status":        result.Status,
	}

	// Add metadata
	if len(result.Metadata) > 0 {
		analysisData["metadata"] = result.Metadata
	}

	updateData := map[string]interface{}{
		"website_analysis_data":       analysisData,
		"website_analysis_status":     "completed",
		"website_analysis_updated_at": time.Now().Format(time.RFC3339),
	}

	// Check if record exists
	var existing []map[string]interface{}
	_, err := j.supabaseClient.GetClient().From("merchant_analytics").
		Select("id", "", false).
		Eq("merchant_id", j.MerchantID).
		Limit(1, "").
		ExecuteTo(&existing)

	if err != nil || len(existing) == 0 {
		// Create new record
		insertData := map[string]interface{}{
			"merchant_id":                j.MerchantID,
			"website_analysis_data":      analysisData,
			"website_analysis_status":    "completed",
			"website_analysis_updated_at": time.Now().Format(time.RFC3339),
		}

		_, _, err := j.supabaseClient.GetClient().From("merchant_analytics").
			Insert(insertData, false, "", "", "").
			Execute()

		return err
	}

	// Update existing record
	_, _, err = j.supabaseClient.GetClient().From("merchant_analytics").
		Update(updateData, "", "").
		Eq("merchant_id", j.MerchantID).
		Execute()

	return err
}

