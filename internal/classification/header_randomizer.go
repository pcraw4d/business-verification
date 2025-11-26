package classification

import (
	"math/rand"
	"os"
	"strconv"
	"time"
)

var (
	// Accept-Language variants for randomization
	acceptLanguageVariants = []string{
		"en-US,en;q=0.9",
		"en-US,en;q=0.9,fr;q=0.8",
		"en-GB,en;q=0.9",
		"en-US,en;q=0.9,es;q=0.8",
		"en-US,en;q=0.9,de;q=0.8",
		"en-US,en;q=0.9,ja;q=0.8",
		"en-US,en;q=0.9,zh;q=0.8",
		"en-US,en;q=0.9,pt;q=0.8",
		"en-US,en;q=0.9,ru;q=0.8",
		"en-US,en;q=0.9,it;q=0.8",
	}

	// Accept-Encoding variants
	acceptEncodingVariants = []string{
		"gzip, deflate, br",
		"gzip, deflate",
		"gzip, br",
		"deflate, br",
		"gzip",
		"br",
	}

	// Sec-Fetch-Dest variants
	secFetchDestVariants = []string{
		"document",
		"empty",
		"image",
		"script",
		"style",
	}

	// Sec-Fetch-Mode variants
	secFetchModeVariants = []string{
		"navigate",
		"cors",
		"no-cors",
		"same-origin",
	}

	// Sec-Fetch-Site variants
	secFetchSiteVariants = []string{
		"none",
		"same-origin",
		"same-site",
		"cross-site",
	}

	// Sec-Ch-Ua variants (browser versions)
	secChUaVariants = []string{
		`"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
		`"Not_A Brand";v="8", "Chromium";v="121", "Google Chrome";v="121"`,
		`"Not_A Brand";v="8", "Chromium";v="122", "Google Chrome";v="122"`,
		`"Not_A Brand";v="8", "Chromium";v="119", "Google Chrome";v="119"`,
		`"Chromium";v="120", "Google Chrome";v="120", "Not_A Brand";v="8"`,
	}

	// Sec-Ch-Ua-Mobile variants
	secChUaMobileVariants = []string{
		"?0",
		"?1",
	}

	// Sec-Ch-Ua-Platform variants
	secChUaPlatformVariants = []string{
		`"Windows"`,
		`"macOS"`,
		`"Linux"`,
		`"Android"`,
		`"iOS"`,
	}

	// Cache-Control variants
	cacheControlVariants = []string{
		"max-age=0",
		"no-cache",
		"no-cache, no-store, must-revalidate",
		"max-age=3600",
	}
)

// HeaderRandomizer manages header randomization for web scraping
type HeaderRandomizer struct {
	enabled bool
	rng     *rand.Rand
}

// NewHeaderRandomizer creates a new header randomizer
func NewHeaderRandomizer() *HeaderRandomizer {
	enabled := os.Getenv("SCRAPING_HEADER_RANDOMIZATION_ENABLED")
	enabledBool := true // Default to enabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	return &HeaderRandomizer{
		enabled: enabledBool,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetRandomizedHeaders generates randomized but realistic browser headers
// while maintaining the provided base User-Agent (which should be our identifiable one)
func (hr *HeaderRandomizer) GetRandomizedHeaders(baseUserAgent string) map[string]string {
	headers := make(map[string]string)

	// Always set the identifiable User-Agent (never randomize this)
	headers["User-Agent"] = baseUserAgent

	if !hr.enabled {
		// If randomization is disabled, return minimal headers with User-Agent
		headers["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
		headers["Accept-Language"] = "en-US,en;q=0.9"
		headers["Accept-Encoding"] = "gzip, deflate, br"
		headers["Connection"] = "keep-alive"
		headers["Upgrade-Insecure-Requests"] = "1"
		return headers
	}

	// Randomize Accept header with quality values
	acceptVariants := []string{
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
	}
	headers["Accept"] = acceptVariants[hr.rng.Intn(len(acceptVariants))]

	// Randomize Accept-Language
	headers["Accept-Language"] = acceptLanguageVariants[hr.rng.Intn(len(acceptLanguageVariants))]

	// Randomize Accept-Encoding
	headers["Accept-Encoding"] = acceptEncodingVariants[hr.rng.Intn(len(acceptEncodingVariants))]

	// Always include DNT (Do Not Track) - consistent
	headers["DNT"] = "1"

	// Always include Connection - consistent
	headers["Connection"] = "keep-alive"

	// Always include Upgrade-Insecure-Requests - consistent
	headers["Upgrade-Insecure-Requests"] = "1"

	// Randomize Sec-Fetch-* headers (only if not disabled)
	if hr.shouldIncludeSecFetchHeaders() {
		headers["Sec-Fetch-Dest"] = secFetchDestVariants[hr.rng.Intn(len(secFetchDestVariants))]
		headers["Sec-Fetch-Mode"] = secFetchModeVariants[hr.rng.Intn(len(secFetchModeVariants))]
		headers["Sec-Fetch-Site"] = secFetchSiteVariants[hr.rng.Intn(len(secFetchSiteVariants))]
		headers["Sec-Fetch-User"] = "?1" // Usually ?1 for navigation
	}

	// Randomize Sec-Ch-Ua headers (browser client hints)
	if hr.shouldIncludeSecChUaHeaders() {
		headers["Sec-Ch-Ua"] = secChUaVariants[hr.rng.Intn(len(secChUaVariants))]
		headers["Sec-Ch-Ua-Mobile"] = secChUaMobileVariants[hr.rng.Intn(len(secChUaMobileVariants))]
		headers["Sec-Ch-Ua-Platform"] = secChUaPlatformVariants[hr.rng.Intn(len(secChUaPlatformVariants))]
	}

	// Randomize Cache-Control
	headers["Cache-Control"] = cacheControlVariants[hr.rng.Intn(len(cacheControlVariants))]

	return headers
}

// GetRandomizedHeadersWithReferer generates headers with an optional referer
func (hr *HeaderRandomizer) GetRandomizedHeadersWithReferer(baseUserAgent string, referer string) map[string]string {
	headers := hr.GetRandomizedHeaders(baseUserAgent)
	
	// Add referer if provided (for realistic navigation patterns)
	if referer != "" && hr.enabled {
		headers["Referer"] = referer
	}
	
	return headers
}

// shouldIncludeSecFetchHeaders determines if Sec-Fetch-* headers should be included
// Some sites may block requests with these headers, so we make it configurable
func (hr *HeaderRandomizer) shouldIncludeSecFetchHeaders() bool {
	envVal := os.Getenv("SCRAPING_INCLUDE_SEC_FETCH_HEADERS")
	if envVal == "" {
		return true // Default to including them
	}
	val, err := strconv.ParseBool(envVal)
	if err != nil {
		return true
	}
	return val
}

// shouldIncludeSecChUaHeaders determines if Sec-Ch-Ua-* headers should be included
func (hr *HeaderRandomizer) shouldIncludeSecChUaHeaders() bool {
	envVal := os.Getenv("SCRAPING_INCLUDE_SEC_CH_UA_HEADERS")
	if envVal == "" {
		return true // Default to including them
	}
	val, err := strconv.ParseBool(envVal)
	if err != nil {
		return true
	}
	return val
}

// GetRandomizedHeaders is a convenience function that uses a default randomizer
func GetRandomizedHeaders(baseUserAgent string) map[string]string {
	hr := NewHeaderRandomizer()
	return hr.GetRandomizedHeaders(baseUserAgent)
}

// GetRandomizedHeadersWithReferer is a convenience function with referer support
func GetRandomizedHeadersWithReferer(baseUserAgent string, referer string) map[string]string {
	hr := NewHeaderRandomizer()
	return hr.GetRandomizedHeadersWithReferer(baseUserAgent, referer)
}

// IsEnabled checks if header randomization is enabled
func (hr *HeaderRandomizer) IsEnabled() bool {
	return hr.enabled
}

// SetEnabled enables or disables header randomization
func (hr *HeaderRandomizer) SetEnabled(enabled bool) {
	hr.enabled = enabled
}

