package classification

import (
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

// TimingPatternGenerator generates human-like timing delays using statistical distributions
type TimingPatternGenerator struct {
	enabled      bool
	rng          *rand.Rand
	domainTiming map[string]*domainTimingState
	timingMutex  sync.RWMutex
}

// domainTimingState tracks timing patterns per domain
type domainTimingState struct {
	lastRequestTime time.Time
	requestCount    int
	avgDelay        time.Duration
}

// NewTimingPatternGenerator creates a new timing pattern generator
func NewTimingPatternGenerator() *TimingPatternGenerator {
	enabled := os.Getenv("SCRAPING_HUMAN_LIKE_TIMING_ENABLED")
	enabledBool := true // Default to enabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	return &TimingPatternGenerator{
		enabled:      enabledBool,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		domainTiming: make(map[string]*domainTimingState),
	}
}

// GetHumanLikeDelay generates a human-like delay based on the base delay and domain
// Uses Weibull distribution for realistic human timing patterns
func (tpg *TimingPatternGenerator) GetHumanLikeDelay(baseDelay time.Duration, domain string) time.Duration {
	if !tpg.enabled {
		// If disabled, return base delay with simple jitter
		jitter := time.Duration(float64(baseDelay) * 0.2 * tpg.rng.Float64())
		return baseDelay + jitter
	}

	// Get or create domain timing state
	tpg.timingMutex.Lock()
	state, exists := tpg.domainTiming[domain]
	if !exists {
		state = &domainTimingState{
			lastRequestTime: time.Now(),
			requestCount:    0,
			avgDelay:        baseDelay,
		}
		tpg.domainTiming[domain] = state
	}
	tpg.timingMutex.Unlock()

	// Calculate delay using Weibull distribution (good for modeling human behavior)
	// Weibull parameters: shape (k) = 1.5 (slightly right-skewed), scale (λ) = baseDelay
	weibullDelay := tpg.weibullDistribution(baseDelay, 1.5)

	// Occasionally add longer pauses (simulating reading/thinking time)
	// 10% chance of adding 2-5x the base delay
	if tpg.rng.Float64() < 0.1 {
		extraDelay := time.Duration(float64(baseDelay) * (2.0 + 3.0*tpg.rng.Float64()))
		weibullDelay += extraDelay
	}

	// Occasionally add shorter bursts (simulating quick navigation)
	// 5% chance of reducing delay by up to 50%
	if tpg.rng.Float64() < 0.05 {
		weibullDelay = time.Duration(float64(weibullDelay) * (0.5 + 0.5*tpg.rng.Float64()))
	}

	// Ensure minimum delay is respected
	if weibullDelay < baseDelay {
		weibullDelay = baseDelay
	}

	// Update domain timing state
	tpg.timingMutex.Lock()
	state.lastRequestTime = time.Now()
	state.requestCount++
	// Update average delay (exponential moving average)
	state.avgDelay = time.Duration(float64(state.avgDelay)*0.9 + float64(weibullDelay)*0.1)
	tpg.timingMutex.Unlock()

	return weibullDelay
}

// weibullDistribution generates a value from a Weibull distribution
// k = shape parameter, lambda = scale parameter
func (tpg *TimingPatternGenerator) weibullDistribution(lambda time.Duration, k float64) time.Duration {
	// Generate uniform random number [0, 1)
	u := tpg.rng.Float64()
	if u == 0 {
		u = 0.0001 // Avoid log(0)
	}

	// Weibull inverse CDF: lambda * (-ln(1-u))^(1/k)
	weibullValue := float64(lambda) * math.Pow(-math.Log(1-u), 1.0/k)
	
	return time.Duration(weibullValue)
}

// GetHumanLikeDelayWithCrawlDelay generates a delay that respects both base delay and robots.txt crawl-delay
func (tpg *TimingPatternGenerator) GetHumanLikeDelayWithCrawlDelay(baseDelay time.Duration, crawlDelay time.Duration, domain string) time.Duration {
	// Use the maximum of base delay and crawl delay as the minimum
	effectiveBaseDelay := baseDelay
	if crawlDelay > baseDelay {
		effectiveBaseDelay = crawlDelay
	}

	return tpg.GetHumanLikeDelay(effectiveBaseDelay, domain)
}

// GetHumanLikeDelay is a convenience function using a default generator
func GetHumanLikeDelay(baseDelay time.Duration, domain string) time.Duration {
	tpg := NewTimingPatternGenerator()
	return tpg.GetHumanLikeDelay(baseDelay, domain)
}

// GetHumanLikeDelayWithCrawlDelay is a convenience function with crawl-delay support
func GetHumanLikeDelayWithCrawlDelay(baseDelay time.Duration, crawlDelay time.Duration, domain string) time.Duration {
	tpg := NewTimingPatternGenerator()
	return tpg.GetHumanLikeDelayWithCrawlDelay(baseDelay, crawlDelay, domain)
}

// CleanupOldDomainTiming removes old domain timing entries to prevent memory leaks
func (tpg *TimingPatternGenerator) CleanupOldDomainTiming(maxAge time.Duration) {
	tpg.timingMutex.Lock()
	defer tpg.timingMutex.Unlock()

	now := time.Now()
	for domain, state := range tpg.domainTiming {
		if now.Sub(state.lastRequestTime) > maxAge {
			delete(tpg.domainTiming, domain)
		}
	}
}

// GetDomainStats returns timing statistics for a domain
func (tpg *TimingPatternGenerator) GetDomainStats(domain string) (avgDelay time.Duration, requestCount int, exists bool) {
	tpg.timingMutex.RLock()
	defer tpg.timingMutex.RUnlock()

	state, exists := tpg.domainTiming[domain]
	if !exists {
		return 0, 0, false
	}

	return state.avgDelay, state.requestCount, true
}

// IsEnabled checks if human-like timing is enabled
func (tpg *TimingPatternGenerator) IsEnabled() bool {
	return tpg.enabled
}

// SetEnabled enables or disables human-like timing
func (tpg *TimingPatternGenerator) SetEnabled(enabled bool) {
	tpg.enabled = enabled
}

