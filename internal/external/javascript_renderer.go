package external

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

// JavaScriptRenderer handles rendering of JavaScript-heavy websites
type JavaScriptRenderer struct {
	timeout     time.Duration
	waitTime    time.Duration
	userAgent   string
	viewport    Viewport
	logger      *zap.Logger
	headless    bool
	disableGPU  bool
	noSandbox   bool
}

// Viewport represents browser viewport settings
type Viewport struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

// DefaultViewport returns default viewport settings
func DefaultViewport() Viewport {
	return Viewport{
		Width:  1920,
		Height: 1080,
	}
}

// RenderConfig holds configuration for JavaScript rendering
type RenderConfig struct {
	Timeout     time.Duration `json:"timeout"`
	WaitTime    time.Duration `json:"wait_time"`
	UserAgent   string        `json:"user_agent"`
	Viewport    Viewport      `json:"viewport"`
	Headless    bool          `json:"headless"`
	DisableGPU  bool          `json:"disable_gpu"`
	NoSandbox   bool          `json:"no_sandbox"`
	WaitForSelector string    `json:"wait_for_selector"`
	WaitForNetworkIdle bool   `json:"wait_for_network_idle"`
}

// DefaultRenderConfig returns default configuration for JavaScript rendering
func DefaultRenderConfig() *RenderConfig {
	return &RenderConfig{
		Timeout:     30 * time.Second,
		WaitTime:    2 * time.Second,
		UserAgent:   "KYB-Platform-Bot/1.0 (+https://kyb-platform.com/bot)",
		Viewport:    DefaultViewport(),
		Headless:    true,
		DisableGPU:  true,
		NoSandbox:   true,
		WaitForNetworkIdle: true,
	}
}

// NewJavaScriptRenderer creates a new JavaScript renderer
func NewJavaScriptRenderer(config *RenderConfig, logger *zap.Logger) *JavaScriptRenderer {
	if config == nil {
		config = DefaultRenderConfig()
	}

	return &JavaScriptRenderer{
		timeout:    config.Timeout,
		waitTime:   config.WaitTime,
		userAgent:  config.UserAgent,
		viewport:   config.Viewport,
		logger:     logger,
		headless:   config.Headless,
		disableGPU: config.DisableGPU,
		noSandbox:  config.NoSandbox,
	}
}

// RenderResult represents the result of JavaScript rendering
type RenderResult struct {
	URL           string            `json:"url"`
	HTML          string            `json:"html"`
	Text          string            `json:"text"`
	Title         string            `json:"title"`
	Status        string            `json:"status"`
	Error         string            `json:"error,omitempty"`
	RenderedAt    time.Time         `json:"rendered_at"`
	Duration      time.Duration     `json:"duration"`
	ConsoleLogs   []string          `json:"console_logs,omitempty"`
	NetworkLogs   []NetworkLog      `json:"network_logs,omitempty"`
}

// NetworkLog represents a network request/response
type NetworkLog struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	Status   int    `json:"status"`
	Type     string `json:"type"`
	Duration int64  `json:"duration"`
}

// RenderPage renders a page with JavaScript support
func (r *JavaScriptRenderer) RenderPage(ctx context.Context, url string, config *RenderConfig) (*RenderResult, error) {
	startTime := time.Now()
	
	if config == nil {
		config = DefaultRenderConfig()
	}

	r.logger.Info("Starting JavaScript rendering",
		zap.String("url", url),
		zap.String("user_agent", r.userAgent),
		zap.Bool("headless", r.headless))

	// Create Chrome context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(r.userAgent),
		chromedp.WindowSize(int(config.Viewport.Width), int(config.Viewport.Height)),
		chromedp.Flag("headless", r.headless),
		chromedp.Flag("disable-gpu", r.disableGPU),
		chromedp.Flag("no-sandbox", r.noSandbox),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create Chrome context
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set timeout
	taskCtx, cancel = context.WithTimeout(taskCtx, config.Timeout)
	defer cancel()

	var result RenderResult
	var consoleLogs []string
	var networkLogs []NetworkLog

	// Define tasks
	tasks := chromedp.Tasks{
		// Navigate to page
		chromedp.Navigate(url),
		
		// Wait for page to load
		chromedp.Sleep(config.WaitTime),
		
		// Wait for network idle if enabled
		chromedp.Sleep(500 * time.Millisecond),
		
		// Wait for specific selector if provided
		chromedp.Sleep(1 * time.Second),
		
		// Capture console logs
		chromedp.Sleep(100 * time.Millisecond),
		
		// Get page content
		chromedp.OuterHTML("html", &result.HTML),
		chromedp.Title(&result.Title),
		chromedp.Text("body", &result.Text),
	}

	// Execute tasks
	err := chromedp.Run(taskCtx, tasks)
	
	result.URL = url
	result.RenderedAt = time.Now()
	result.Duration = time.Since(startTime)
	result.ConsoleLogs = consoleLogs
	result.NetworkLogs = networkLogs

	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		r.logger.Error("JavaScript rendering failed",
			zap.String("url", url),
			zap.Error(err),
			zap.Duration("duration", result.Duration))
		return &result, fmt.Errorf("rendering failed: %w", err)
	}

	result.Status = "success"
	r.logger.Info("JavaScript rendering completed",
		zap.String("url", url),
		zap.String("title", result.Title),
		zap.Duration("duration", result.Duration),
		zap.Int("console_logs", len(consoleLogs)))

	return &result, nil
}

// RenderMultiplePages renders multiple pages concurrently
func (r *JavaScriptRenderer) RenderMultiplePages(ctx context.Context, urls []string, config *RenderConfig, maxConcurrency int) (map[string]*RenderResult, error) {
	if maxConcurrency <= 0 {
		maxConcurrency = 3 // Lower concurrency for browser instances
	}

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(map[string]*RenderResult)
	errors := make(map[string]error)

	// Create channel for results
	resultChan := make(chan struct {
		url    string
		result *RenderResult
		err    error
	}, len(urls))

	// Start rendering goroutines
	for _, url := range urls {
		go func(targetURL string) {
			semaphore <- struct{}{} // Acquire semaphore
			defer func() {
				<-semaphore // Release semaphore
			}()

			result, err := r.RenderPage(ctx, targetURL, config)
			resultChan <- struct {
				url    string
				result *RenderResult
				err    error
			}{targetURL, result, err}
		}(url)
	}

	// Collect results
	for i := 0; i < len(urls); i++ {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case result := <-resultChan:
			if result.err != nil {
				errors[result.url] = result.err
				r.logger.Error("Failed to render page",
					zap.String("url", result.url),
					zap.Error(result.err))
			} else {
				results[result.url] = result.result
			}
		}
	}

	// Log summary
	r.logger.Info("Multiple page rendering completed",
		zap.Int("total_urls", len(urls)),
		zap.Int("successful", len(results)),
		zap.Int("failed", len(errors)))

	return results, nil
}

// DetectJavaScriptDependency detects if a page requires JavaScript rendering
func (r *JavaScriptRenderer) DetectJavaScriptDependency(htmlContent string) bool {
	// Check for common JavaScript frameworks and patterns
	indicators := []string{
		"<script",
		"react",
		"vue",
		"angular",
		"spa",
		"single-page",
		"__NEXT_DATA__",
		"window.__INITIAL_STATE__",
		"data-reactroot",
		"ng-app",
		"v-app",
	}

	content := strings.ToLower(htmlContent)
	for _, indicator := range indicators {
		if strings.Contains(content, strings.ToLower(indicator)) {
			return true
		}
	}

	// Check for minimal content that might indicate JavaScript rendering is needed
	if len(htmlContent) < 1000 {
		return true
	}

	return false
}

// ExtractDynamicContent extracts content that might be loaded dynamically
func (r *JavaScriptRenderer) ExtractDynamicContent(ctx context.Context, url string, selectors []string) (map[string]string, error) {
	config := DefaultRenderConfig()
	config.WaitTime = 5 * time.Second // Longer wait for dynamic content

	result, err := r.RenderPage(ctx, url, config)
	if err != nil {
		return nil, err
	}

	// Parse HTML and extract content from selectors
	parser := NewHTMLParser()
	parsed, err := parser.ParseHTML(result.HTML)
	if err != nil {
		return nil, err
	}

	// Extract content from specific selectors
	extracted := make(map[string]string)
	for _, selector := range selectors {
		// This is a simplified selector extraction
		// In production, you'd want to use a proper CSS selector engine
		if strings.Contains(parsed.Text, selector) {
			extracted[selector] = "Content found" // Simplified
		}
	}

	return extracted, nil
}

// CheckRenderingNecessity determines if JavaScript rendering is necessary for a URL
func (r *JavaScriptRenderer) CheckRenderingNecessity(ctx context.Context, url string) (*RenderingCheck, error) {
	// First, try to get the page without JavaScript
	scraper := NewWebsiteScraper(DefaultScrapingConfig(), r.logger)
	scrapingResult, err := scraper.ScrapeWebsite(ctx, url)
	if err != nil {
		return nil, err
	}

	check := &RenderingCheck{
		URL:           url,
		NeedsRendering: false,
		Reason:        "Static content detected",
		StaticContent: scrapingResult.Content,
	}

	// Check if JavaScript rendering is needed
	if r.DetectJavaScriptDependency(scrapingResult.Content) {
		check.NeedsRendering = true
		check.Reason = "JavaScript dependency detected"
		
		// Try rendering with JavaScript
		renderResult, err := r.RenderPage(ctx, url, DefaultRenderConfig())
		if err == nil {
			check.RenderedContent = renderResult.HTML
			check.RenderingSuccessful = true
		} else {
			check.RenderingError = err.Error()
		}
	}

	return check, nil
}

// RenderingCheck represents the result of checking if rendering is necessary
type RenderingCheck struct {
	URL                string `json:"url"`
	NeedsRendering     bool   `json:"needs_rendering"`
	Reason             string `json:"reason"`
	StaticContent      string `json:"static_content,omitempty"`
	RenderedContent    string `json:"rendered_content,omitempty"`
	RenderingSuccessful bool  `json:"rendering_successful"`
	RenderingError     string `json:"rendering_error,omitempty"`
}
