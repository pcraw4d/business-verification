package webanalysis

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// BotDetectionEvasion provides advanced bot detection evasion capabilities
type BotDetectionEvasion struct {
	fingerprintManager *FingerprintManager
	requestRandomizer  *RequestRandomizer
	captchaDetector    *CAPTCHADetector
	behaviorSimulator  *BehaviorSimulator
	config             EvasionConfig
	mu                 sync.RWMutex
}

// EvasionConfig holds configuration for bot detection evasion
type EvasionConfig struct {
	EnableFingerprintRandomization bool          `json:"enable_fingerprint_randomization"`
	EnableRequestRandomization     bool          `json:"enable_request_randomization"`
	EnableCAPTCHADetection         bool          `json:"enable_captcha_detection"`
	EnableBehaviorSimulation       bool          `json:"enable_behavior_simulation"`
	MaxRetries                     int           `json:"max_retries"`
	RetryDelay                     time.Duration `json:"retry_delay"`
	JitterRange                    time.Duration `json:"jitter_range"`
	UserAgentRotation              bool          `json:"user_agent_rotation"`
	HeaderRandomization            bool          `json:"header_randomization"`
	TimingRandomization            bool          `json:"timing_randomization"`
}

// FingerprintManager manages browser fingerprint randomization
type FingerprintManager struct {
	userAgents     []string
	screenResolutions []string
	colorDepths    []int
	timezones      []string
	languages      []string
	platforms      []string
	plugins        []string
	canvasFingerprints []string
	webglFingerprints  []string
	mu             sync.RWMutex
}

// RequestRandomizer manages request pattern randomization
type RequestRandomizer struct {
	headerVariations map[string][]string
	cookieVariations map[string][]string
	requestPatterns  []RequestPattern
	mu               sync.RWMutex
}

// RequestPattern represents a randomized request pattern
type RequestPattern struct {
	Headers     map[string]string
	Cookies     map[string]string
	UserAgent   string
	Accept      string
	AcceptLang  string
	AcceptEnc   string
	Connection  string
	UpgradeInsecure string
}

// CAPTCHADetector detects and handles CAPTCHA challenges
type CAPTCHADetector struct {
	captchaPatterns []string
	solverServices  map[string]CAPTCHASolver
	mu              sync.RWMutex
}

// CAPTCHASolver represents a CAPTCHA solving service
type CAPTCHASolver struct {
	Name     string
	APIKey   string
	Endpoint string
	Timeout  time.Duration
}

// BehaviorSimulator simulates human-like browsing behavior
type BehaviorSimulator struct {
	mouseMovements    []MouseMovement
	keyboardPatterns  []KeyboardPattern
	scrollPatterns    []ScrollPattern
	clickPatterns     []ClickPattern
	mu                sync.RWMutex
}

// MouseMovement represents mouse movement simulation
type MouseMovement struct {
	StartX, StartY int
	EndX, EndY     int
	Duration       time.Duration
	Curve          string // linear, bezier, natural
}

// KeyboardPattern represents keyboard input simulation
type KeyboardPattern struct {
	Text     string
	Delay    time.Duration
	Typos    float64 // probability of typos
	Corrections bool
}

// ScrollPattern represents scroll behavior simulation
type ScrollPattern struct {
	Direction string // up, down, random
	Distance  int
	Duration  time.Duration
	Smooth    bool
}

// ClickPattern represents click behavior simulation
type ClickPattern struct {
	Element   string
	Position  string // center, random, specific
	Delay     time.Duration
	DoubleClick bool
}

// NewBotDetectionEvasion creates a new bot detection evasion system
func NewBotDetectionEvasion() *BotDetectionEvasion {
	config := EvasionConfig{
		EnableFingerprintRandomization: true,
		EnableRequestRandomization:     true,
		EnableCAPTCHADetection:         true,
		EnableBehaviorSimulation:       true,
		MaxRetries:                     3,
		RetryDelay:                     time.Second * 2,
		JitterRange:                    time.Millisecond * 500,
		UserAgentRotation:              true,
		HeaderRandomization:            true,
		TimingRandomization:            true,
	}

	return &BotDetectionEvasion{
		fingerprintManager: NewFingerprintManager(),
		requestRandomizer:  NewRequestRandomizer(),
		captchaDetector:    NewCAPTCHADetector(),
		behaviorSimulator:  NewBehaviorSimulator(),
		config:             config,
	}
}

// NewFingerprintManager creates a new fingerprint manager
func NewFingerprintManager() *FingerprintManager {
	return &FingerprintManager{
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.0.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		},
		screenResolutions: []string{
			"1920x1080", "1366x768", "1440x900", "1536x864", "1280x720",
			"2560x1440", "1600x900", "1024x768", "1280x800", "1920x1200",
		},
		colorDepths: []int{24, 32, 16},
		timezones: []string{
			"America/New_York", "America/Los_Angeles", "Europe/London",
			"Europe/Paris", "Asia/Tokyo", "Australia/Sydney",
		},
		languages: []string{
			"en-US,en;q=0.9", "en-GB,en;q=0.9", "en-CA,en;q=0.9",
			"fr-FR,fr;q=0.9", "de-DE,de;q=0.9", "es-ES,es;q=0.9",
		},
		platforms: []string{
			"Win32", "MacIntel", "Linux x86_64",
		},
		plugins: []string{
			"PDF Viewer", "Chrome PDF Plugin", "Native Client",
			"Widevine Content Decryption Module", "Shockwave Flash",
		},
		canvasFingerprints: []string{
			"canvas-fingerprint-1", "canvas-fingerprint-2", "canvas-fingerprint-3",
		},
		webglFingerprints: []string{
			"webgl-fingerprint-1", "webgl-fingerprint-2", "webgl-fingerprint-3",
		},
	}
}

// NewRequestRandomizer creates a new request randomizer
func NewRequestRandomizer() *RequestRandomizer {
	return &RequestRandomizer{
		headerVariations: map[string][]string{
			"Accept": {
				"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
				"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
				"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			},
			"Accept-Language": {
				"en-US,en;q=0.9", "en-GB,en;q=0.9", "en-CA,en;q=0.9",
				"fr-FR,fr;q=0.9", "de-DE,de;q=0.9", "es-ES,es;q=0.9",
			},
			"Accept-Encoding": {
				"gzip, deflate, br", "gzip, deflate", "gzip",
			},
			"Connection": {
				"keep-alive", "close",
			},
			"Upgrade-Insecure-Requests": {
				"1",
			},
		},
		cookieVariations: map[string][]string{
			"session_id": {"session-1", "session-2", "session-3"},
			"user_pref":  {"pref-1", "pref-2", "pref-3"},
		},
		requestPatterns: []RequestPattern{
			{
				Headers: map[string]string{
					"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
					"Accept-Language": "en-US,en;q=0.9",
					"Accept-Encoding": "gzip, deflate, br",
					"Connection": "keep-alive",
					"Upgrade-Insecure-Requests": "1",
				},
			},
		},
	}
}

// NewCAPTCHADetector creates a new CAPTCHA detector
func NewCAPTCHADetector() *CAPTCHADetector {
	return &CAPTCHADetector{
		captchaPatterns: []string{
			"captcha", "recaptcha", "hcaptcha", "turnstile",
			"cloudflare", "challenge", "verification",
		},
		solverServices: map[string]CAPTCHASolver{
			"2captcha": {
				Name:     "2captcha",
				Endpoint: "https://2captcha.com/in.php",
				Timeout:  time.Second * 30,
			},
			"anticaptcha": {
				Name:     "anticaptcha",
				Endpoint: "https://api.anti-captcha.com/createTask",
				Timeout:  time.Second * 30,
			},
		},
	}
}

// NewBehaviorSimulator creates a new behavior simulator
func NewBehaviorSimulator() *BehaviorSimulator {
	return &BehaviorSimulator{
		mouseMovements: []MouseMovement{
			{StartX: 100, StartY: 100, EndX: 200, EndY: 200, Duration: time.Millisecond * 500, Curve: "natural"},
			{StartX: 200, StartY: 200, EndX: 300, EndY: 150, Duration: time.Millisecond * 300, Curve: "bezier"},
		},
		keyboardPatterns: []KeyboardPattern{
			{Text: "search query", Delay: time.Millisecond * 100, Typos: 0.05, Corrections: true},
		},
		scrollPatterns: []ScrollPattern{
			{Direction: "down", Distance: 500, Duration: time.Millisecond * 1000, Smooth: true},
			{Direction: "up", Distance: 200, Duration: time.Millisecond * 500, Smooth: true},
		},
		clickPatterns: []ClickPattern{
			{Element: "button", Position: "center", Delay: time.Millisecond * 200, DoubleClick: false},
		},
	}
}

// GenerateRandomizedRequest creates a randomized HTTP request with evasion techniques
func (bde *BotDetectionEvasion) GenerateRandomizedRequest(url string, method string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply fingerprint randomization
	if bde.config.EnableFingerprintRandomization {
		bde.applyFingerprintRandomization(req)
	}

	// Apply request randomization
	if bde.config.EnableRequestRandomization {
		bde.applyRequestRandomization(req)
	}

	// Apply timing randomization
	if bde.config.TimingRandomization {
		bde.applyTimingRandomization()
	}

	return req, nil
}

// applyFingerprintRandomization applies browser fingerprint randomization
func (bde *BotDetectionEvasion) applyFingerprintRandomization(req *http.Request) {
	fm := bde.fingerprintManager

	// Randomize User-Agent
	if bde.config.UserAgentRotation {
		req.Header.Set("User-Agent", fm.getRandomUserAgent())
	}

	// Randomize Accept headers
	req.Header.Set("Accept", fm.getRandomAccept())
	req.Header.Set("Accept-Language", fm.getRandomLanguage())
	req.Header.Set("Accept-Encoding", fm.getRandomEncoding())

	// Add platform-specific headers
	req.Header.Set("Sec-Ch-Ua-Platform", fm.getRandomPlatform())
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")

	// Add viewport and screen info
	req.Header.Set("Sec-Ch-Ua-Platform-Version", fm.getRandomPlatformVersion())
	req.Header.Set("Sec-Ch-Ua-Full-Version", fm.getRandomFullVersion())
}

// applyRequestRandomization applies request pattern randomization
func (bde *BotDetectionEvasion) applyRequestRandomization(req *http.Request) {
	rr := bde.requestRandomizer

	// Randomize headers
	for header, variations := range rr.headerVariations {
		if len(variations) > 0 {
			req.Header.Set(header, rr.getRandomVariation(variations))
		}
	}

	// Add random cookies
	cookies := rr.generateRandomCookies()
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	// Add random query parameters
	req.URL.RawQuery = rr.generateRandomQueryParams()
}

// applyTimingRandomization applies timing randomization
func (bde *BotDetectionEvasion) applyTimingRandomization() {
	// Add jitter to timing
	jitter := bde.generateJitter()
	time.Sleep(jitter)
}

// DetectCAPTCHA detects CAPTCHA challenges in response
func (bde *BotDetectionEvasion) DetectCAPTCHA(resp *http.Response) bool {
	if !bde.config.EnableCAPTCHADetection {
		return false
	}

	// Check response body for CAPTCHA patterns
	body := resp.Body
	if body == nil {
		return false
	}

	// Read a portion of the body to check for CAPTCHA
	buffer := make([]byte, 4096)
	n, err := body.Read(buffer)
	if err != nil && n == 0 {
		return false
	}

	content := strings.ToLower(string(buffer[:n]))
	
	for _, pattern := range bde.captchaDetector.captchaPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

// HandleCAPTCHA handles CAPTCHA challenges
func (bde *BotDetectionEvasion) HandleCAPTCHA(resp *http.Response) (*http.Response, error) {
	// This is a placeholder for CAPTCHA solving logic
	// In a real implementation, this would integrate with CAPTCHA solving services
	
	// For now, return the original response
	return resp, nil
}

// SimulateHumanBehavior simulates human-like browsing behavior
func (bde *BotDetectionEvasion) SimulateHumanBehavior() {
	if !bde.config.EnableBehaviorSimulation {
		return
	}

	bs := bde.behaviorSimulator

	// Simulate mouse movements
	bs.simulateMouseMovement()

	// Simulate keyboard input
	bs.simulateKeyboardInput()

	// Simulate scrolling
	bs.simulateScrolling()

	// Simulate clicking
	bs.simulateClicking()
}

// Helper methods for fingerprint manager
func (fm *FingerprintManager) getRandomUserAgent() string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.userAgents[fm.randomIndex(len(fm.userAgents))]
}

func (fm *FingerprintManager) getRandomAccept() string {
	return "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
}

func (fm *FingerprintManager) getRandomLanguage() string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.languages[fm.randomIndex(len(fm.languages))]
}

func (fm *FingerprintManager) getRandomEncoding() string {
	return "gzip, deflate, br"
}

func (fm *FingerprintManager) getRandomPlatform() string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.platforms[fm.randomIndex(len(fm.platforms))]
}

func (fm *FingerprintManager) getRandomPlatformVersion() string {
	return "10.0"
}

func (fm *FingerprintManager) getRandomFullVersion() string {
	return "120.0.6099.109"
}

// Helper methods for request randomizer
func (rr *RequestRandomizer) getRandomVariation(variations []string) string {
	rr.mu.RLock()
	defer rr.mu.RUnlock()
	return variations[rr.randomIndex(len(variations))]
}

func (rr *RequestRandomizer) generateRandomCookies() map[string]string {
	cookies := make(map[string]string)
	
	// Add some random cookies
	for name, variations := range rr.cookieVariations {
		if len(variations) > 0 {
			cookies[name] = variations[rr.randomIndex(len(variations))]
		}
	}
	
	return cookies
}

func (rr *RequestRandomizer) generateRandomQueryParams() string {
	// Add random query parameters to make requests look more natural
	params := []string{
		fmt.Sprintf("_=%d", time.Now().Unix()),
		fmt.Sprintf("v=%d", rr.randomInt(1, 100)),
	}
	
	return strings.Join(params, "&")
}

func (rr *RequestRandomizer) randomInt(min, max int) int {
	delta := max - min
	if delta <= 0 {
		return min
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		return min
	}
	
	return min + int(n.Int64())
}

// Helper methods for behavior simulator
func (bs *BehaviorSimulator) simulateMouseMovement() {
	// Simulate mouse movement with random delays
	time.Sleep(time.Duration(bs.randomInt(50, 200)) * time.Millisecond)
}

func (bs *BehaviorSimulator) simulateKeyboardInput() {
	// Simulate keyboard input with realistic delays
	time.Sleep(time.Duration(bs.randomInt(100, 300)) * time.Millisecond)
}

func (bs *BehaviorSimulator) simulateScrolling() {
	// Simulate scrolling behavior
	time.Sleep(time.Duration(bs.randomInt(200, 500)) * time.Millisecond)
}

func (bs *BehaviorSimulator) simulateClicking() {
	// Simulate clicking behavior
	time.Sleep(time.Duration(bs.randomInt(100, 250)) * time.Millisecond)
}

// Utility methods
func (bde *BotDetectionEvasion) generateJitter() time.Duration {
	jitter := time.Duration(bde.randomInt(0, int(bde.config.JitterRange.Milliseconds()))) * time.Millisecond
	return jitter
}

func (bde *BotDetectionEvasion) randomInt(min, max int) int {
	delta := max - min
	if delta <= 0 {
		return min
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		return min
	}
	
	return min + int(n.Int64())
}

func (fm *FingerprintManager) randomIndex(length int) int {
	if length <= 0 {
		return 0
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return 0
	}
	
	return int(n.Int64())
}

func (rr *RequestRandomizer) randomIndex(length int) int {
	if length <= 0 {
		return 0
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return 0
	}
	
	return int(n.Int64())
}

func (bs *BehaviorSimulator) randomInt(min, max int) int {
	delta := max - min
	if delta <= 0 {
		return min
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		return min
	}
	
	return min + int(n.Int64())
}
