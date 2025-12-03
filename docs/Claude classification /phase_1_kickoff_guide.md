# Phase 1 Kick-Off Guide: Fix the Foundation
## Weeks 1-2: Scraping → 95% Success Rate

**Goal:** Fix scraping to provide high-quality content to your classifier. This is the root cause of your <5% accuracy.

---

## Overview

**Current State:**
- Scrape success rate: ~50%
- Content quality: Poor (often no output)
- Downstream accuracy: <5%

**Target State:**
- Scrape success rate: ≥95%
- Content quality score: ≥0.7 for 90%+ of scrapes
- Expected accuracy jump: 50-60% (from better input alone)

**Key Changes:**
1. Enhanced content extraction (structured data)
2. Multi-tier fallback scraping strategies
3. Playwright service for JS-heavy sites
4. Content quality validation
5. Comprehensive logging

---

## Week 1: Enhance Content Extraction + Logging

### Task 1: Add Comprehensive Logging (Day 1)

**File:** `internal/external/website_scraper.go`

Add detailed logging at every step to understand failure modes:

```go
import (
    "log/slog"
    "time"
)

func (s *WebsiteScraper) Scrape(url string) (*ScrapedContent, error) {
    startTime := time.Now()
    
    slog.Info("Starting scrape",
        "url", url,
        "timestamp", startTime)
    
    // DNS/Connection Phase
    resp, err := s.httpClient.Get(url)
    if err != nil {
        slog.Error("HTTP request failed",
            "url", url,
            "error", err,
            "stage", "http_request",
            "duration_ms", time.Since(startTime).Milliseconds())
        return nil, err
    }
    defer resp.Body.Close()
    
    slog.Info("HTTP request succeeded",
        "url", url,
        "status_code", resp.StatusCode,
        "content_type", resp.Header.Get("Content-Type"),
        "duration_ms", time.Since(startTime).Milliseconds())
    
    // Check for non-200 responses
    if resp.StatusCode != 200 {
        slog.Warn("Non-200 status code",
            "url", url,
            "status_code", resp.StatusCode)
        // Continue anyway - may still get useful content
    }
    
    // Read body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        slog.Error("Failed to read response body",
            "url", url,
            "error", err,
            "stage", "body_read")
        return nil, err
    }
    
    slog.Info("Response body read",
        "url", url,
        "body_size_bytes", len(body))
    
    // Parse HTML
    doc, err := html.Parse(bytes.NewReader(body))
    if err != nil {
        slog.Error("HTML parsing failed",
            "url", url,
            "error", err,
            "stage", "html_parse",
            "body_preview", string(body[:min(100, len(body))]))
        return nil, err
    }
    
    // Extract structured content
    content := s.extractStructuredContent(doc, string(body))
    
    // Calculate quality score
    qualityScore := calculateContentQuality(content)
    
    slog.Info("Scrape completed",
        "url", url,
        "word_count", content.WordCount,
        "has_title", content.Title != "",
        "has_meta_desc", content.MetaDesc != "",
        "num_headings", len(content.Headings),
        "has_about", content.AboutText != "",
        "quality_score", qualityScore,
        "total_duration_ms", time.Since(startTime).Milliseconds())
    
    return content, nil
}
```

**Why This Matters:**
- You'll see exactly where scraping fails (HTTP? Parsing? Content extraction?)
- Quality score helps identify borderline cases
- Duration tracking identifies performance bottlenecks

---

### Task 2: Extract Structured Content (Day 2-3)

**File:** `internal/external/website_scraper.go`

Replace simple text extraction with structured extraction:

```go
type ScrapedContent struct {
    // Existing fields
    RawHTML     string
    PlainText   string
    
    // NEW: High-signal structured content
    Title       string    `json:"title"`
    MetaDesc    string    `json:"meta_description"`
    Headings    []string  `json:"headings"`        // H1, H2, H3
    NavMenu     []string  `json:"navigation"`      // Nav items (business areas)
    AboutText   string    `json:"about_text"`      // About/Company section
    ProductList []string  `json:"products"`        // Products/services
    ContactInfo string    `json:"contact"`         // Contact page content
    
    // Quality metrics
    WordCount   int       `json:"word_count"`
    Language    string    `json:"language"`
    HasLogo     bool      `json:"has_logo"`
    QualityScore float64  `json:"quality_score"`
    
    // Metadata
    Domain      string    `json:"domain"`
    ScrapedAt   time.Time `json:"scraped_at"`
}

func (s *WebsiteScraper) extractStructuredContent(doc *html.Node, rawHTML string) *ScrapedContent {
    content := &ScrapedContent{
        RawHTML:   rawHTML,
        ScrapedAt: time.Now(),
    }
    
    // Extract domain
    content.Domain = extractDomain(rawHTML)
    
    // Extract title
    content.Title = extractTitle(doc)
    
    // Extract meta description
    content.MetaDesc = extractMetaDescription(doc)
    
    // Extract all headings (H1-H3) - high signal for classification
    content.Headings = extractHeadings(doc)
    
    // Extract navigation menu - indicates business areas
    content.NavMenu = extractNavigation(doc)
    
    // Extract "About" section - highest quality content
    content.AboutText = extractAboutSection(doc)
    
    // Extract product/service listings
    content.ProductList = extractProductsServices(doc)
    
    // Extract contact page content
    content.ContactInfo = extractContactInfo(doc)
    
    // Detect language
    content.Language = detectLanguage(doc)
    
    // Check for logo (indicator of legitimate business)
    content.HasLogo = hasLogo(doc)
    
    // Combine all text with priority weighting
    content.PlainText = combineTextWithWeights(content)
    
    // Count words
    content.WordCount = len(strings.Fields(content.PlainText))
    
    // Calculate quality score
    content.QualityScore = calculateContentQuality(content)
    
    return content
}

// Extract title
func extractTitle(doc *html.Node) string {
    if title := findNode(doc, "title"); title != nil {
        return extractText(title)
    }
    
    // Fallback: look for h1
    if h1 := findNodeByTag(doc, "h1"); h1 != nil {
        return extractText(h1)
    }
    
    return ""
}

// Extract meta description
func extractMetaDescription(doc *html.Node) string {
    var description string
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "meta" {
            var isDescription bool
            var content string
            
            for _, attr := range n.Attr {
                if attr.Key == "name" && attr.Val == "description" {
                    isDescription = true
                }
                if attr.Key == "content" {
                    content = attr.Val
                }
            }
            
            if isDescription && content != "" {
                description = content
                return
            }
        }
        
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
    
    return description
}

// Extract headings (H1-H3)
func extractHeadings(doc *html.Node) []string {
    headings := []string{}
    
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode {
            if n.Data == "h1" || n.Data == "h2" || n.Data == "h3" {
                text := extractText(n)
                if text != "" && len(text) < 200 { // Reasonable heading length
                    headings = append(headings, text)
                }
            }
        }
        
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
    
    return headings
}

// Extract navigation menu
func extractNavigation(doc *html.Node) []string {
    navItems := []string{}
    
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "nav" {
            // Found nav element, extract all links
            extractLinks(n, &navItems)
        }
        
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
    
    return navItems
}

// Extract "About" section - HIGHEST PRIORITY
func extractAboutSection(doc *html.Node) string {
    aboutText := ""
    
    // Strategy 1: Look for section/div with id/class containing "about"
    aboutSection := findNodeWithIdentifier(doc, []string{"about", "about-us", "company", "who-we-are"})
    if aboutSection != nil {
        aboutText = extractText(aboutSection)
        if len(aboutText) > 100 { // Substantial content
            return aboutText
        }
    }
    
    // Strategy 2: Look for <a> tag with "about" in href, then find that page content
    // (In a real implementation, you might crawl the about page separately)
    
    return aboutText
}

// Extract products/services
func extractProductsServices(doc *html.Node) []string {
    products := []string{}
    
    // Look for common product/service listing patterns
    productSection := findNodeWithIdentifier(doc, []string{
        "products", "services", "menu", "offerings", "solutions",
    })
    
    if productSection != nil {
        // Extract list items
        var f func(*html.Node)
        f = func(n *html.Node) {
            if n.Type == html.ElementNode && (n.Data == "li" || n.Data == "h3" || n.Data == "h4") {
                text := extractText(n)
                if text != "" && len(text) < 100 { // Reasonable product name
                    products = append(products, text)
                }
            }
            
            for c := n.FirstChild; c != nil; c = c.NextSibling {
                f(c)
            }
        }
        f(productSection)
    }
    
    return products
}

// Combine text with priority weighting
func combineTextWithWeights(content *ScrapedContent) string {
    parts := []string{}
    
    // Title (highest weight - repeat 3x)
    if content.Title != "" {
        parts = append(parts, content.Title, content.Title, content.Title)
    }
    
    // Meta description (high weight - repeat 2x)
    if content.MetaDesc != "" {
        parts = append(parts, content.MetaDesc, content.MetaDesc)
    }
    
    // About text (high weight - repeat 2x)
    if content.AboutText != "" {
        parts = append(parts, content.AboutText, content.AboutText)
    }
    
    // Headings (medium weight - repeat 1x)
    if len(content.Headings) > 0 {
        parts = append(parts, strings.Join(content.Headings, ". "))
    }
    
    // Navigation (medium weight)
    if len(content.NavMenu) > 0 {
        parts = append(parts, strings.Join(content.NavMenu, ". "))
    }
    
    // Products (lower weight)
    if len(content.ProductList) > 0 {
        parts = append(parts, strings.Join(content.ProductList, ". "))
    }
    
    return strings.Join(parts, ". ")
}

// Calculate content quality score (0.0 - 1.0)
func calculateContentQuality(content *ScrapedContent) float64 {
    score := 0.0
    
    // Has title? +0.15
    if content.Title != "" {
        score += 0.15
    }
    
    // Has meta description? +0.15
    if content.MetaDesc != "" {
        score += 0.15
    }
    
    // Has headings? +0.15
    if len(content.Headings) > 0 {
        score += 0.15
    }
    
    // Has about section? +0.20 (most important)
    if content.AboutText != "" && len(content.AboutText) > 100 {
        score += 0.20
    }
    
    // Sufficient word count? +0.15
    if content.WordCount >= 200 {
        score += 0.15
    }
    
    // Has navigation? +0.10
    if len(content.NavMenu) > 0 {
        score += 0.10
    }
    
    // Has logo? +0.10
    if content.HasLogo {
        score += 0.10
    }
    
    return math.Min(score, 1.0)
}

// Helper: Find node by tag
func findNodeByTag(n *html.Node, tag string) *html.Node {
    if n.Type == html.ElementNode && n.Data == tag {
        return n
    }
    
    for c := n.FirstChild; c != nil; c = c.NextSibling {
        if result := findNodeByTag(c, tag); result != nil {
            return result
        }
    }
    
    return nil
}

// Helper: Find node with identifier in id/class
func findNodeWithIdentifier(n *html.Node, identifiers []string) *html.Node {
    if n.Type == html.ElementNode {
        for _, attr := range n.Attr {
            if attr.Key == "id" || attr.Key == "class" {
                for _, identifier := range identifiers {
                    if strings.Contains(strings.ToLower(attr.Val), identifier) {
                        return n
                    }
                }
            }
        }
    }
    
    for c := n.FirstChild; c != nil; c = c.NextSibling {
        if result := findNodeWithIdentifier(c, identifiers); result != nil {
            return result
        }
    }
    
    return nil
}

// Helper: Extract text from node
func extractText(n *html.Node) string {
    var text strings.Builder
    
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.TextNode {
            text.WriteString(n.Data)
        }
        
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(n)
    
    // Clean whitespace
    result := strings.TrimSpace(text.String())
    result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
    
    return result
}
```

**Test After This:**
```bash
go test ./internal/external -v -run TestWebsiteScraper
```

Pick 5-10 websites from your test set and verify:
- Quality scores are ≥0.7
- Structured content is extracted (title, headings, about)
- Word counts are reasonable (≥200)

---

## Week 2: Multi-Tier Scraping + Playwright Service

### Task 3: Implement Scraper Strategies (Day 4-5)

**File:** `internal/external/website_scraper.go`

Add strategy pattern for fallback scraping:

```go
// ScraperStrategy interface
type ScraperStrategy interface {
    Scrape(url string) (*ScrapedContent, error)
    Name() string
}

// Strategy 1: Simple HTTP client (fastest)
type SimpleHTTPScraper struct {
    client *http.Client
}

func (s *SimpleHTTPScraper) Name() string {
    return "simple_http"
}

func (s *SimpleHTTPScraper) Scrape(url string) (*ScrapedContent, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    // Basic headers
    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; YourBot/1.0)")
    
    resp, err := s.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Read and parse
    body, _ := io.ReadAll(resp.Body)
    doc, err := html.Parse(bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    
    content := extractStructuredContent(doc, string(body))
    return content, nil
}

// Strategy 2: HTTP with realistic browser headers
type BrowserHeadersScraper struct {
    client *http.Client
}

func (s *BrowserHeadersScraper) Name() string {
    return "browser_headers"
}

func (s *BrowserHeadersScraper) Scrape(url string) (*ScrapedContent, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    // Realistic browser headers to avoid bot detection
    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
    req.Header.Set("Accept-Language", "en-US,en;q=0.5")
    req.Header.Set("Accept-Encoding", "gzip, deflate, br")
    req.Header.Set("DNT", "1")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Upgrade-Insecure-Requests", "1")
    
    resp, err := s.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Handle gzip
    var reader io.Reader = resp.Body
    if resp.Header.Get("Content-Encoding") == "gzip" {
        gzipReader, err := gzip.NewReader(resp.Body)
        if err != nil {
            return nil, err
        }
        defer gzipReader.Close()
        reader = gzipReader
    }
    
    body, _ := io.ReadAll(reader)
    doc, err := html.Parse(bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    
    content := extractStructuredContent(doc, string(body))
    return content, nil
}

// Strategy 3: Playwright service (for JS-heavy sites)
type PlaywrightScraper struct {
    serviceURL string
    client     *http.Client
}

func (s *PlaywrightScraper) Name() string {
    return "playwright"
}

func (s *PlaywrightScraper) Scrape(url string) (*ScrapedContent, error) {
    // Call Playwright service
    reqBody, _ := json.Marshal(map[string]string{
        "url": url,
    })
    
    resp, err := s.client.Post(
        s.serviceURL+"/scrape",
        "application/json",
        bytes.NewReader(reqBody),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        HTML  string `json:"html"`
        Error string `json:"error"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    if result.Error != "" {
        return nil, fmt.Errorf("playwright error: %s", result.Error)
    }
    
    // Parse HTML from Playwright
    doc, err := html.Parse(strings.NewReader(result.HTML))
    if err != nil {
        return nil, err
    }
    
    content := extractStructuredContent(doc, result.HTML)
    return content, nil
}

// Main WebsiteScraper with strategy fallback
type WebsiteScraper struct {
    strategies []ScraperStrategy
}

func NewWebsiteScraper(playwrightServiceURL string) *WebsiteScraper {
    httpClient := &http.Client{
        Timeout: 15 * time.Second,
    }
    
    return &WebsiteScraper{
        strategies: []ScraperStrategy{
            &SimpleHTTPScraper{client: httpClient},
            &BrowserHeadersScraper{client: httpClient},
            &PlaywrightScraper{
                serviceURL: playwrightServiceURL,
                client: &http.Client{Timeout: 30 * time.Second},
            },
        },
    }
}

func (s *WebsiteScraper) Scrape(url string) (*ScrapedContent, error) {
    var lastErr error
    
    for i, strategy := range s.strategies {
        slog.Info("Attempting scrape strategy",
            "strategy", strategy.Name(),
            "url", url,
            "attempt", i+1)
        
        content, err := strategy.Scrape(url)
        
        if err == nil && isContentValid(content) {
            slog.Info("Strategy succeeded",
                "strategy", strategy.Name(),
                "quality_score", content.QualityScore)
            return content, nil
        }
        
        lastErr = err
        slog.Warn("Strategy failed, trying next",
            "strategy", strategy.Name(),
            "error", err,
            "quality_score", func() float64 {
                if content != nil {
                    return content.QualityScore
                }
                return 0.0
            }())
    }
    
    return nil, fmt.Errorf("all scraping strategies failed: %w", lastErr)
}

// Validate content quality
func isContentValid(content *ScrapedContent) bool {
    if content == nil {
        return false
    }
    
    // Minimum word count
    if content.WordCount < 50 {
        slog.Debug("Content invalid: insufficient word count", "word_count", content.WordCount)
        return false
    }
    
    // Must have basic metadata
    if content.Title == "" && content.MetaDesc == "" {
        slog.Debug("Content invalid: no title or meta description")
        return false
    }
    
    // Check for error pages
    if containsErrorIndicators(content.PlainText) {
        slog.Debug("Content invalid: contains error indicators")
        return false
    }
    
    // Quality score threshold
    if content.QualityScore < 0.5 {
        slog.Debug("Content invalid: quality score too low", "score", content.QualityScore)
        return false
    }
    
    return true
}

func containsErrorIndicators(text string) bool {
    lowerText := strings.ToLower(text)
    errorIndicators := []string{
        "404", "not found", "page not found",
        "403", "access denied", "forbidden",
        "500", "internal server error",
        "503", "service unavailable",
        "error", "oops",
    }
    
    for _, indicator := range errorIndicators {
        if strings.Contains(lowerText, indicator) {
            return true
        }
    }
    
    return false
}
```

---

### Task 4: Create Playwright Service (Day 6-7)

**Create new Railway service:**

**File:** `services/playwright-scraper/package.json`
```json
{
  "name": "playwright-scraper",
  "version": "1.0.0",
  "main": "index.js",
  "dependencies": {
    "express": "^4.18.2",
    "playwright": "^1.40.0"
  },
  "scripts": {
    "start": "node index.js"
  }
}
```

**File:** `services/playwright-scraper/index.js`
```javascript
const express = require('express');
const { chromium } = require('playwright');

const app = express();
app.use(express.json());

// Health check
app.get('/health', (req, res) => {
    res.json({ status: 'ok', service: 'playwright-scraper' });
});

// Main scraping endpoint
app.post('/scrape', async (req, res) => {
    const { url } = req.body;
    
    if (!url) {
        return res.status(400).json({ error: 'URL is required' });
    }
    
    console.log(`Scraping: ${url}`);
    
    let browser;
    try {
        // Launch browser
        browser = await chromium.launch({
            headless: true,
            args: ['--no-sandbox', '--disable-setuid-sandbox']
        });
        
        const context = await browser.newContext({
            userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
            viewport: { width: 1920, height: 1080 }
        });
        
        const page = await context.newPage();
        
        // Set timeout
        page.setDefaultTimeout(15000);
        
        // Navigate
        await page.goto(url, { 
            waitUntil: 'networkidle',
            timeout: 15000 
        });
        
        // Wait a bit for any dynamic content
        await page.waitForTimeout(1000);
        
        // Get full HTML
        const html = await page.content();
        
        // Close
        await browser.close();
        
        console.log(`Success: ${url} (${html.length} bytes)`);
        
        res.json({
            html: html,
            success: true
        });
        
    } catch (error) {
        console.error(`Error scraping ${url}:`, error.message);
        
        if (browser) {
            await browser.close();
        }
        
        res.status(500).json({
            error: error.message,
            success: false
        });
    }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Playwright scraper listening on port ${PORT}`);
});
```

**File:** `services/playwright-scraper/Dockerfile`
```dockerfile
FROM mcr.microsoft.com/playwright:v1.40.0-focal

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install --production

# Copy application
COPY . .

# Expose port
EXPOSE 3000

# Start service
CMD ["npm", "start"]
```

**Deploy to Railway:**
```bash
# In Railway dashboard:
# 1. Create new service: "playwright-scraper"
# 2. Connect to GitHub repo
# 3. Set root directory: services/playwright-scraper
# 4. Railway will auto-detect Dockerfile and deploy
# 5. Note the service URL (e.g., https://playwright-scraper-production.up.railway.app)
```

**Update Go service environment:**
```bash
# Add to your classification service env vars:
PLAYWRIGHT_SERVICE_URL=https://playwright-scraper-production.up.railway.app
```

---

### Task 5: Test with Your Test Set (Day 7)

**Run classification on your existing test set:**

```bash
# Assuming you have test URLs in Supabase
# Create a test script or use your existing test runner

# Expected results after Phase 1:
# - Scrape success rate: ≥95%
# - Average quality score: ≥0.7
# - Average word count: ≥200
# - Accuracy improvement: 10x (from <5% to 50-60%)
```

**Create a test report:**
```bash
# Track these metrics:
# 1. Scrape success rate by strategy
#    - SimpleHTTP: X%
#    - BrowserHeaders: X%
#    - Playwright: X%
# 2. Content quality scores
# 3. Classification accuracy (if ground truth available)
# 4. Performance (avg scrape time by strategy)
```

---

## Success Criteria for Phase 1

Before moving to Phase 2, verify:

- [ ] ✅ Scrape success rate ≥95%
- [ ] ✅ Content quality score ≥0.7 for 90%+ of successful scrapes
- [ ] ✅ Average word count ≥200 words
- [ ] ✅ No "no output" errors
- [ ] ✅ Playwright service deployed and working
- [ ] ✅ Strategy fallback working (logs show which strategy succeeded)
- [ ] ✅ Comprehensive logging in place
- [ ] ✅ Classification accuracy improved (even if still < 80%)

**Key Log Lines to Look For:**
```
INFO: Starting scrape url=https://example.com
INFO: Attempting scrape strategy strategy=simple_http attempt=1
WARN: Strategy failed, trying next strategy=simple_http
INFO: Attempting scrape strategy strategy=browser_headers attempt=2
INFO: Strategy succeeded strategy=browser_headers quality_score=0.85
INFO: Scrape completed word_count=342 quality_score=0.85 total_duration_ms=245
```

---

## Troubleshooting

**Issue: Playwright service crashes**
- Check Railway logs
- Ensure enough memory (512MB minimum)
- Verify Dockerfile is correct

**Issue: Still getting poor quality scores**
- Check specific extraction functions (title, about, etc.)
- Verify HTML structure of failing sites
- Add more fallback strategies for extraction

**Issue: Simple HTTP works but returns 403**
- Bot detection - this is why we have BrowserHeaders and Playwright
- Should automatically fallback to next strategy

**Issue: Scraping is slow**
- Most requests should hit SimpleHTTP or BrowserHeaders (<2s)
- Playwright only for 10-20% of requests
- Check timeouts aren't too aggressive

---

## Next Steps: Phase 2

Once Phase 1 is complete, you'll move to Phase 2 (Weeks 3-4):
- Return top 3 codes (not just 1)
- Improve confidence calibration
- Add fast path (<100ms)
- Generate explanations

**Expected Accuracy After Phase 2:** 80-85%

---

## Questions During Phase 1?

If you hit any issues:
1. Check Railway logs for both services
2. Review slog output for failure patterns
3. Test individual strategies manually
4. Verify Playwright service is accessible from classification service

You've got this! 🚀
