package classification

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

// ScrapingSession manages cookies and session state for web scraping
type ScrapingSession struct {
	domain      string
	cookieJar   *cookiejar.Jar
	referer     string
	createdAt   time.Time
	lastAccess  time.Time
	requestCount int
	mu          sync.RWMutex
}

// ScrapingSessionManager manages multiple scraping sessions
type ScrapingSessionManager struct {
	enabled     bool
	sessions    map[string]*ScrapingSession
	sessionMutex sync.RWMutex
	maxAge      time.Duration
}

// NewScrapingSessionManager creates a new session manager
func NewScrapingSessionManager() *ScrapingSessionManager {
	enabled := os.Getenv("SCRAPING_SESSION_MANAGEMENT_ENABLED")
	enabledBool := true // Default to enabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	maxAgeStr := os.Getenv("SCRAPING_SESSION_MAX_AGE")
	maxAge := 1 * time.Hour // Default 1 hour
	if maxAgeStr != "" {
		if duration, err := time.ParseDuration(maxAgeStr); err == nil {
			maxAge = duration
		}
	}

	return &ScrapingSessionManager{
		enabled:  enabledBool,
		sessions: make(map[string]*ScrapingSession),
		maxAge:   maxAge,
	}
}

// GetOrCreateSession gets an existing session for a domain or creates a new one
func (ssm *ScrapingSessionManager) GetOrCreateSession(domain string) (*ScrapingSession, error) {
	if !ssm.enabled {
		// If disabled, return a temporary session that doesn't persist
		return ssm.createTemporarySession(domain)
	}

	ssm.sessionMutex.Lock()
	defer ssm.sessionMutex.Unlock()

	// Check if session exists and is still valid
	if session, exists := ssm.sessions[domain]; exists {
		if time.Since(session.lastAccess) < ssm.maxAge {
			session.lastAccess = time.Now()
			session.requestCount++
			return session, nil
		}
		// Session expired, remove it
		delete(ssm.sessions, domain)
	}

	// Create new session
	session, err := ssm.createSession(domain)
	if err != nil {
		return nil, err
	}

	ssm.sessions[domain] = session
	return session, nil
}

// createSession creates a new scraping session
func (ssm *ScrapingSessionManager) createSession(domain string) (*ScrapingSession, error) {
	// Create cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &ScrapingSession{
		domain:       domain,
		cookieJar:    jar,
		referer:      "",
		createdAt:    now,
		lastAccess:   now,
		requestCount: 1,
	}, nil
}

// createTemporarySession creates a temporary session that doesn't persist
func (ssm *ScrapingSessionManager) createTemporarySession(domain string) (*ScrapingSession, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &ScrapingSession{
		domain:       domain,
		cookieJar:    jar,
		referer:      "",
		createdAt:    time.Now(),
		lastAccess:   time.Now(),
		requestCount: 1,
	}, nil
}

// GetSession retrieves an existing session without creating a new one
func (ssm *ScrapingSessionManager) GetSession(domain string) (*ScrapingSession, bool) {
	ssm.sessionMutex.RLock()
	defer ssm.sessionMutex.RUnlock()

	session, exists := ssm.sessions[domain]
	if !exists {
		return nil, false
	}

	// Check if session is still valid
	if time.Since(session.lastAccess) >= ssm.maxAge {
		return nil, false
	}

	return session, true
}

// UpdateReferer updates the referer for a session (for realistic navigation)
func (ssm *ScrapingSessionManager) UpdateReferer(domain string, referer string) {
	if !ssm.enabled {
		return
	}

	ssm.sessionMutex.RLock()
	session, exists := ssm.sessions[domain]
	ssm.sessionMutex.RUnlock()

	if exists {
		session.mu.Lock()
		session.referer = referer
		session.mu.Unlock()
	}
}

// GetReferer gets the current referer for a session
func (ssm *ScrapingSessionManager) GetReferer(domain string) string {
	if !ssm.enabled {
		return ""
	}

	ssm.sessionMutex.RLock()
	session, exists := ssm.sessions[domain]
	ssm.sessionMutex.RUnlock()

	if !exists {
		return ""
	}

	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.referer
}

// CleanupExpiredSessions removes expired sessions
func (ssm *ScrapingSessionManager) CleanupExpiredSessions() {
	ssm.sessionMutex.Lock()
	defer ssm.sessionMutex.Unlock()

	now := time.Now()
	for domain, session := range ssm.sessions {
		if now.Sub(session.lastAccess) >= ssm.maxAge {
			delete(ssm.sessions, domain)
		}
	}
}

// GetCookieJar returns the cookie jar for a session
func (ss *ScrapingSession) GetCookieJar() *cookiejar.Jar {
	return ss.cookieJar
}

// GetReferer returns the current referer
func (ss *ScrapingSession) GetReferer() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.referer
}

// SetReferer sets the referer for the session
func (ss *ScrapingSession) SetReferer(referer string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.referer = referer
}

// UpdateAccess updates the last access time
func (ss *ScrapingSession) UpdateAccess() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.lastAccess = time.Now()
	ss.requestCount++
}

// GetDomain returns the domain for this session
func (ss *ScrapingSession) GetDomain() string {
	return ss.domain
}

// GetRequestCount returns the number of requests made in this session
func (ss *ScrapingSession) GetRequestCount() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.requestCount
}

// CreateHTTPClientWithSession creates an HTTP client configured with the session's cookie jar
func CreateHTTPClientWithSession(session *ScrapingSession, timeout time.Duration) *http.Client {
	if session == nil {
		return &http.Client{
			Timeout: timeout,
		}
	}

	return &http.Client{
		Timeout: timeout,
		Jar:     session.GetCookieJar(),
	}
}

// GetCookiesForURL gets cookies for a specific URL from the session
func (ss *ScrapingSession) GetCookiesForURL(urlStr string) []*http.Cookie {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}

	return ss.cookieJar.Cookies(parsedURL)
}

// SetCookiesForURL sets cookies for a specific URL in the session
func (ss *ScrapingSession) SetCookiesForURL(urlStr string, cookies []*http.Cookie) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	ss.cookieJar.SetCookies(parsedURL, cookies)
	return nil
}

// IsEnabled checks if session management is enabled
func (ssm *ScrapingSessionManager) IsEnabled() bool {
	return ssm.enabled
}

// SetEnabled enables or disables session management
func (ssm *ScrapingSessionManager) SetEnabled(enabled bool) {
	ssm.enabled = enabled
}

// GetSessionCount returns the number of active sessions
func (ssm *ScrapingSessionManager) GetSessionCount() int {
	ssm.sessionMutex.RLock()
	defer ssm.sessionMutex.RUnlock()
	return len(ssm.sessions)
}

