package webanalysis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// JavaScriptScraper provides JavaScript-enabled web scraping capabilities
type JavaScriptScraper struct {
	config     JavaScriptScraperConfig
	browserMgr *BrowserManager
	mu         sync.RWMutex
}

// JavaScriptScraperConfig holds configuration for JavaScript scraping
type JavaScriptScraperConfig struct {
	EnableJavaScript       bool          `json:"enable_javascript"`
	WaitForLoadTimeout     time.Duration `json:"wait_for_load_timeout"`
	JavaScriptTimeout      time.Duration `json:"javascript_timeout"`
	EnableFingerprinting   bool          `json:"enable_fingerprinting"`
	EnableSessionMgmt      bool          `json:"enable_session_mgmt"`
	EnableResourceBlocking bool          `json:"enable_resource_blocking"`
	EnableUserInteractions bool          `json:"enable_user_interactions"`
	MaxConcurrentBrowsers  int           `json:"max_concurrent_browsers"`
	ViewportWidth          int           `json:"viewport_width"`
	ViewportHeight         int           `json:"viewport_height"`
	UserDataDir            string        `json:"user_data_dir"`
}

// JavaScriptScrapingJob represents a JavaScript-enabled scraping request
type JavaScriptScrapingJob struct {
	URL                string            `json:"url"`
	Business           string            `json:"business"`
	WaitForSelector    string            `json:"wait_for_selector"`
	WaitForLoadState   string            `json:"wait_for_load_state"`
	ExecuteJavaScript  []string          `json:"execute_javascript"`
	BlockResources     []string          `json:"block_resources"`
	UserInteractions   []UserInteraction `json:"user_interactions"`
	ExtractSelectors   map[string]string `json:"extract_selectors"`
	EnableScreenshots  bool              `json:"enable_screenshots"`
	ScreenshotPath     string            `json:"screenshot_path"`
	Timeout            time.Duration     `json:"timeout"`
	MaxRetries         int               `json:"max_retries"`
	FingerprintProfile string            `json:"fingerprint_profile"`
	SessionID          string            `json:"session_id"`
}

// UserInteraction represents a user interaction to simulate
type UserInteraction struct {
	Type      string `json:"type"`       // "click", "scroll", "type", "hover"
	Selector  string `json:"selector"`   // CSS selector
	Value     string `json:"value"`      // For type interactions
	WaitAfter int    `json:"wait_after"` // Milliseconds to wait after interaction
}

// JavaScriptScrapingResult represents the result of JavaScript-enabled scraping
type JavaScriptScrapingResult struct {
	URL                string            `json:"url"`
	Title              string            `json:"title"`
	HTML               string            `json:"html"`
	Text               string            `json:"text"`
	StatusCode         int               `json:"status_code"`
	ResponseTime       time.Duration     `json:"response_time"`
	JavaScriptExecuted []string          `json:"javascript_executed"`
	ExtractedData      map[string]string `json:"extracted_data"`
	DOMSnapshot        string            `json:"dom_snapshot"`
	ScreenshotPath     string            `json:"screenshot_path"`
	JavaScriptErrors   []string          `json:"javascript_errors"`
	LoadTime           time.Duration     `json:"load_time"`
	NetworkRequests    []NetworkRequest  `json:"network_requests"`
	Error              string            `json:"error,omitempty"`
	ScrapedAt          time.Time         `json:"scraped_at"`
	BrowserFingerprint string            `json:"browser_fingerprint"`
	SessionID          string            `json:"session_id"`
}

// NetworkRequest represents a network request made during scraping
type NetworkRequest struct {
	URL         string `json:"url"`
	Method      string `json:"method"`
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// BrowserManager manages headless browser instances
type BrowserManager struct {
	config      JavaScriptScraperConfig
	browsers    map[string]*BrowserInstance
	mu          sync.RWMutex
	maxBrowsers int
}

// BrowserInstance represents a single browser instance
type BrowserInstance struct {
	ID        string
	Context   context.Context
	Cancel    context.CancelFunc
	SessionID string
	LastUsed  time.Time
	InUse     bool
}

// NewJavaScriptScraper creates a new JavaScript-enabled scraper
func NewJavaScriptScraper(config JavaScriptScraperConfig) *JavaScriptScraper {
	if config.WaitForLoadTimeout == 0 {
		config.WaitForLoadTimeout = 30 * time.Second
	}
	if config.JavaScriptTimeout == 0 {
		config.JavaScriptTimeout = 10 * time.Second
	}
	if config.MaxConcurrentBrowsers == 0 {
		config.MaxConcurrentBrowsers = 5
	}
	if config.ViewportWidth == 0 {
		config.ViewportWidth = 1920
	}
	if config.ViewportHeight == 0 {
		config.ViewportHeight = 1080
	}

	return &JavaScriptScraper{
		config: config,
		browserMgr: &BrowserManager{
			config:      config,
			browsers:    make(map[string]*BrowserInstance),
			maxBrowsers: config.MaxConcurrentBrowsers,
		},
	}
}

// ScrapeWebsite performs JavaScript-enabled website scraping
func (js *JavaScriptScraper) ScrapeWebsite(job *JavaScriptScrapingJob) (*JavaScriptScrapingResult, error) {
	start := time.Now()
	result := &JavaScriptScrapingResult{
		URL:               job.URL,
		ScrapedAt:         time.Now(),
		ExtractedData:     make(map[string]string),
		JavaScriptErrors:  []string{},
		NetworkRequests:   []NetworkRequest{},
		JavaScriptExecuted: []string{},
	}

	// Get or create browser instance
	browser, err := js.browserMgr.GetBrowser(job.SessionID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get browser: %v", err)
		return result, err
	}
	defer js.browserMgr.ReleaseBrowser(browser.ID)

	// Use browser context directly
	ctx := browser.Context

	// Perform the scraping
	err = chromedp.Run(ctx,
		// Set viewport
		chromedp.EmulateViewport(int64(js.config.ViewportWidth), int64(js.config.ViewportHeight)),
		
		// Navigate to URL
		chromedp.Navigate(job.URL),
		
		// Wait for page load
		chromedp.Sleep(2 * time.Second), // Simple wait for page load
		
		// Wait for specific selector if provided
		chromedp.WaitVisible(job.WaitForSelector, chromedp.ByQuery),
		
		// Execute custom JavaScript
		js.executeCustomJavaScript(job.ExecuteJavaScript),
		
		// Perform user interactions
		js.performUserInteractions(job.UserInteractions),
		
		// Extract content
		chromedp.OuterHTML("html", &result.HTML),
		chromedp.Title(&result.Title),
		chromedp.Text("body", &result.Text),
		
		// Take screenshot if enabled
		js.takeScreenshot(job.EnableScreenshots, job.ScreenshotPath, &result.ScreenshotPath),
		
		// Extract data from selectors
		js.extractDataFromSelectors(job.ExtractSelectors, &result.ExtractedData),
	)

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.ResponseTime = time.Since(start)
	result.SessionID = browser.SessionID

	return result, nil
}

// setupBrowserFingerprinting sets up browser fingerprinting
func (js *JavaScriptScraper) setupBrowserFingerprinting(ctx context.Context, profile string) context.Context {
	// Set user agent - this would need to be done at allocator level
	// For now, return the context as-is
	return ctx
}

// executeCustomJavaScript executes custom JavaScript code
func (js *JavaScriptScraper) executeCustomJavaScript(scripts []string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		for _, script := range scripts {
			var result string
			err := chromedp.Evaluate(script, &result).Do(ctx)
			if err != nil {
				log.Printf("JavaScript execution error: %v", err)
				continue
			}
		}
		return nil
	}
}

// performUserInteractions performs user interactions
func (js *JavaScriptScraper) performUserInteractions(interactions []UserInteraction) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		for _, interaction := range interactions {
			switch interaction.Type {
			case "click":
				err := chromedp.Click(interaction.Selector, chromedp.ByQuery).Do(ctx)
				if err != nil {
					log.Printf("Click interaction failed: %v", err)
				}
			case "scroll":
				err := chromedp.Evaluate(fmt.Sprintf("window.scrollTo(0, %s)", interaction.Value), nil).Do(ctx)
				if err != nil {
					log.Printf("Scroll interaction failed: %v", err)
				}
			case "type":
				err := chromedp.SendKeys(interaction.Selector, interaction.Value, chromedp.ByQuery).Do(ctx)
				if err != nil {
					log.Printf("Type interaction failed: %v", err)
				}
			case "hover":
				// Hover is not directly supported in chromedp, use JavaScript instead
				err := chromedp.Evaluate(fmt.Sprintf(`
					const element = document.querySelector('%s');
					if (element) {
						element.dispatchEvent(new MouseEvent('mouseover', {
							bubbles: true,
							cancelable: true,
						}));
					}
				`, interaction.Selector), nil).Do(ctx)
				if err != nil {
					log.Printf("Hover interaction failed: %v", err)
				}
			}

			// Wait after interaction if specified
			if interaction.WaitAfter > 0 {
				time.Sleep(time.Duration(interaction.WaitAfter) * time.Millisecond)
			}
		}
		return nil
	}
}

// takeScreenshot takes a screenshot if enabled
func (js *JavaScriptScraper) takeScreenshot(enabled bool, path string, resultPath *string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		if !enabled {
			return nil
		}

		var buf []byte
		err := chromedp.FullScreenshot(&buf, 90).Do(ctx)
		if err != nil {
			return err
		}

		// Save screenshot to file
		// Implementation for saving file
		*resultPath = path
		return nil
	}
}

// extractDataFromSelectors extracts data from specified selectors
func (js *JavaScriptScraper) extractDataFromSelectors(selectors map[string]string, extractedData *map[string]string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		for key, selector := range selectors {
			var value string
			err := chromedp.Text(selector, &value, chromedp.ByQuery).Do(ctx)
			if err != nil {
				log.Printf("Failed to extract data from selector %s: %v", selector, err)
				continue
			}
			(*extractedData)[key] = strings.TrimSpace(value)
		}
		return nil
	}
}

// getFingerprintUserAgent returns a user agent based on fingerprint profile
func (js *JavaScriptScraper) getFingerprintUserAgent(profile string) string {
	userAgents := map[string]string{
		"chrome_windows":  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"chrome_mac":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"chrome_linux":    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"firefox_windows": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"firefox_mac":     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
		"safari_mac":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	}

	if ua, exists := userAgents[profile]; exists {
		return ua
	}

	// Default to Chrome Windows
	return userAgents["chrome_windows"]
}

// GetBrowser gets or creates a browser instance
func (bm *BrowserManager) GetBrowser(sessionID string) (*BrowserInstance, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Check if we have an available browser
	for _, browser := range bm.browsers {
		if !browser.InUse {
			browser.InUse = true
			browser.LastUsed = time.Now()
			browser.SessionID = sessionID
			return browser, nil
		}
	}

	// Create new browser if under limit
	if len(bm.browsers) < bm.maxBrowsers {
		ctx, cancel := chromedp.NewContext(context.Background())
		browser := &BrowserInstance{
			ID:        fmt.Sprintf("browser_%d", len(bm.browsers)+1),
			Context:   ctx,
			Cancel:    cancel,
			SessionID: sessionID,
			LastUsed:  time.Now(),
			InUse:     true,
		}
		bm.browsers[browser.ID] = browser
		return browser, nil
	}

	return nil, fmt.Errorf("no available browser instances")
}

// ReleaseBrowser releases a browser instance
func (bm *BrowserManager) ReleaseBrowser(browserID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if browser, exists := bm.browsers[browserID]; exists {
		browser.InUse = false
	}
}

// CleanupBrowsers cleans up unused browser instances
func (bm *BrowserManager) CleanupBrowsers() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for id, browser := range bm.browsers {
		if !browser.InUse && time.Since(browser.LastUsed) > 5*time.Minute {
			browser.Cancel()
			delete(bm.browsers, id)
		}
	}
}
