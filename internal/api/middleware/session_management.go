package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SessionManager manages user sessions for concurrent request handling
type SessionManager struct {
	sessions    map[string]*UserSession
	mu          sync.RWMutex
	config      *SessionConfig
	metrics     *SessionMetrics
	cleanupDone chan struct{}
}

// SessionConfig holds configuration for session management
type SessionConfig struct {
	SessionTimeout    time.Duration // Session timeout duration
	CleanupInterval   time.Duration // Cleanup interval for expired sessions
	MaxSessions       int           // Maximum number of concurrent sessions
	SessionIDLength   int           // Length of session ID
	CookieName        string        // Cookie name for session ID
	CookieSecure      bool          // Whether to use secure cookies
	CookieHTTPOnly    bool          // Whether to use HTTP-only cookies
	CookieSameSite    http.SameSite // SameSite cookie attribute
	EnableMetrics     bool          // Enable detailed metrics collection
	TrackUserActivity bool          // Track detailed user activity
}

// UserSession represents a user session
type UserSession struct {
	ID               string                 `json:"id"`
	UserID           string                 `json:"user_id,omitempty"`
	IPAddress        string                 `json:"ip_address"`
	UserAgent        string                 `json:"user_agent"`
	CreatedAt        time.Time              `json:"created_at"`
	LastAccessTime   time.Time              `json:"last_access_time"`
	LastActivityTime time.Time              `json:"last_activity_time"`
	RequestCount     int64                  `json:"request_count"`
	IsActive         bool                   `json:"is_active"`
	ExpiresAt        time.Time              `json:"expires_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ActivityLog      []ActivityEntry        `json:"activity_log,omitempty"`
	mu               sync.RWMutex
}

// ActivityEntry represents a user activity entry
type ActivityEntry struct {
	Timestamp    time.Time              `json:"timestamp"`
	Action       string                 `json:"action"`
	Endpoint     string                 `json:"endpoint"`
	Method       string                 `json:"method"`
	StatusCode   int                    `json:"status_code"`
	Duration     time.Duration          `json:"duration"`
	RequestSize  int64                  `json:"request_size"`
	ResponseSize int64                  `json:"response_size"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SessionMetrics tracks session-related metrics
type SessionMetrics struct {
	TotalSessions        int64         `json:"total_sessions"`
	ActiveSessions       int64         `json:"active_sessions"`
	ExpiredSessions      int64         `json:"expired_sessions"`
	AverageSessionLength time.Duration `json:"average_session_length"`
	TotalRequests        int64         `json:"total_requests"`
	RequestsPerSession   float64       `json:"requests_per_session"`
	LastUpdated          time.Time     `json:"last_updated"`
	PeakSessions         int64         `json:"peak_sessions"`
	PeakSessionsTime     time.Time     `json:"peak_sessions_time"`
	SessionsByHour       map[int]int64 `json:"sessions_by_hour"`
	mu                   sync.RWMutex
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		SessionTimeout:    24 * time.Hour,          // 24 hour session timeout
		CleanupInterval:   1 * time.Hour,           // Cleanup every hour
		MaxSessions:       1000,                    // Support 1000 concurrent sessions
		SessionIDLength:   32,                      // 32 character session ID
		CookieName:        "kyb_session_id",        // Cookie name
		CookieSecure:      false,                   // Allow non-HTTPS for development
		CookieHTTPOnly:    true,                    // HTTP-only cookies
		CookieSameSite:    http.SameSiteStrictMode, // Strict SameSite
		EnableMetrics:     true,                    // Enable metrics
		TrackUserActivity: true,                    // Track user activity
	}
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *SessionConfig) *SessionManager {
	if config == nil {
		config = DefaultSessionConfig()
	}

	sm := &SessionManager{
		sessions: make(map[string]*UserSession),
		config:   config,
		metrics: &SessionMetrics{
			SessionsByHour: make(map[int]int64),
			LastUpdated:    time.Now(),
		},
		cleanupDone: make(chan struct{}),
	}

	// Start cleanup goroutine
	go sm.cleanupExpiredSessions()

	return sm
}

// generateSessionID generates a cryptographically secure session ID
func (sm *SessionManager) generateSessionID() (string, error) {
	bytes := make([]byte, sm.config.SessionIDLength/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new user session
func (sm *SessionManager) CreateSession(r *http.Request, userID string) (*UserSession, error) {
	// Check session limits
	sm.mu.RLock()
	sessionCount := len(sm.sessions)
	sm.mu.RUnlock()

	if sessionCount >= sm.config.MaxSessions {
		return nil, fmt.Errorf("maximum session limit reached: %d", sm.config.MaxSessions)
	}

	// Generate session ID
	sessionID, err := sm.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Extract client information
	ipAddress := sm.getClientIP(r)
	userAgent := r.UserAgent()

	// Create session
	now := time.Now()
	session := &UserSession{
		ID:               sessionID,
		UserID:           userID,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		CreatedAt:        now,
		LastAccessTime:   now,
		LastActivityTime: now,
		RequestCount:     0,
		IsActive:         true,
		ExpiresAt:        now.Add(sm.config.SessionTimeout),
		Metadata:         make(map[string]interface{}),
		ActivityLog:      make([]ActivityEntry, 0),
	}

	// Store session
	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	// Update metrics
	sm.updateMetrics()

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*UserSession, bool) {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		sm.DeleteSession(sessionID)
		return nil, false
	}

	return session, true
}

// UpdateSessionActivity updates session activity
func (sm *SessionManager) UpdateSessionActivity(sessionID string, activity ActivityEntry) error {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// Update session metadata
	session.LastAccessTime = time.Now()
	session.LastActivityTime = activity.Timestamp
	session.RequestCount++
	session.ExpiresAt = time.Now().Add(sm.config.SessionTimeout)

	// Add activity log entry if tracking is enabled
	if sm.config.TrackUserActivity {
		session.ActivityLog = append(session.ActivityLog, activity)

		// Limit activity log size to prevent memory bloat
		maxLogEntries := 1000
		if len(session.ActivityLog) > maxLogEntries {
			session.ActivityLog = session.ActivityLog[len(session.ActivityLog)-maxLogEntries:]
		}
	}

	// Update metrics
	sm.updateMetrics()

	return nil
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.mu.Lock()
		session.IsActive = false
		session.mu.Unlock()

		delete(sm.sessions, sessionID)
		sm.updateMetricsLocked()
		return true
	}

	return false
}

// GetActiveSessions returns all active sessions
func (sm *SessionManager) GetActiveSessions() []*UserSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*UserSession, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		if session.IsActive && time.Now().Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// GetSessionsByUser returns all sessions for a specific user
func (sm *SessionManager) GetSessionsByUser(userID string) []*UserSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*UserSession, 0)
	for _, session := range sm.sessions {
		if session.UserID == userID && session.IsActive && time.Now().Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// GetMetrics returns current session metrics
func (sm *SessionManager) GetMetrics() *SessionMetrics {
	sm.metrics.mu.RLock()
	defer sm.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metricsCopy := &SessionMetrics{
		TotalSessions:        sm.metrics.TotalSessions,
		ActiveSessions:       sm.metrics.ActiveSessions,
		ExpiredSessions:      sm.metrics.ExpiredSessions,
		AverageSessionLength: sm.metrics.AverageSessionLength,
		TotalRequests:        sm.metrics.TotalRequests,
		RequestsPerSession:   sm.metrics.RequestsPerSession,
		LastUpdated:          sm.metrics.LastUpdated,
		PeakSessions:         sm.metrics.PeakSessions,
		PeakSessionsTime:     sm.metrics.PeakSessionsTime,
		SessionsByHour:       make(map[int]int64),
	}

	// Copy sessions by hour map
	for hour, count := range sm.metrics.SessionsByHour {
		metricsCopy.SessionsByHour[hour] = count
	}

	return metricsCopy
}

// updateMetrics updates session metrics
func (sm *SessionManager) updateMetrics() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sm.updateMetricsLocked()
}

// updateMetricsLocked updates metrics (assumes mutex is already locked)
func (sm *SessionManager) updateMetricsLocked() {
	sm.metrics.mu.Lock()
	defer sm.metrics.mu.Unlock()

	now := time.Now()
	activeCount := int64(0)
	totalRequests := int64(0)
	totalSessionLength := time.Duration(0)

	// Calculate metrics
	for _, session := range sm.sessions {
		if session.IsActive && now.Before(session.ExpiresAt) {
			activeCount++
			totalRequests += session.RequestCount
			sessionLength := session.LastAccessTime.Sub(session.CreatedAt)
			totalSessionLength += sessionLength
		}
	}

	// Update metrics
	sm.metrics.ActiveSessions = activeCount
	sm.metrics.TotalRequests = totalRequests
	sm.metrics.LastUpdated = now

	// Update peak sessions
	if activeCount > sm.metrics.PeakSessions {
		sm.metrics.PeakSessions = activeCount
		sm.metrics.PeakSessionsTime = now
	}

	// Calculate averages
	if activeCount > 0 {
		sm.metrics.AverageSessionLength = totalSessionLength / time.Duration(activeCount)
		sm.metrics.RequestsPerSession = float64(totalRequests) / float64(activeCount)
	}

	// Update sessions by hour
	hour := now.Hour()
	sm.metrics.SessionsByHour[hour] = activeCount
}

// getClientIP extracts the client IP address from the request
func (sm *SessionManager) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// cleanupExpiredSessions runs a background cleanup process
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(sm.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.performCleanup()
		case <-sm.cleanupDone:
			return
		}
	}
}

// performCleanup removes expired sessions
func (sm *SessionManager) performCleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	expiredCount := int64(0)

	for sessionID, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			session.mu.Lock()
			session.IsActive = false
			session.mu.Unlock()

			delete(sm.sessions, sessionID)
			expiredCount++
		}
	}

	// Update expired sessions metric
	sm.metrics.mu.Lock()
	sm.metrics.ExpiredSessions += expiredCount
	sm.metrics.mu.Unlock()

	if expiredCount > 0 {
		fmt.Printf("Cleaned up %d expired sessions\n", expiredCount)
	}
}

// Shutdown gracefully shuts down the session manager
func (sm *SessionManager) Shutdown() {
	close(sm.cleanupDone)

	// Clean up all sessions
	sm.mu.Lock()
	for sessionID, session := range sm.sessions {
		session.mu.Lock()
		session.IsActive = false
		session.mu.Unlock()
		delete(sm.sessions, sessionID)
	}
	sm.mu.Unlock()
}

// SessionMiddleware creates middleware for session management
func (sm *SessionManager) SessionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Get or create session
			sessionID := sm.getSessionIDFromRequest(r)
			var session *UserSession
			var exists bool

			if sessionID != "" {
				session, exists = sm.GetSession(sessionID)
			}

			// If no valid session exists, create a new one
			if !exists {
				var err error
				session, err = sm.CreateSession(r, "") // Empty user ID for anonymous sessions
				if err != nil {
					http.Error(w, "Failed to create session", http.StatusInternalServerError)
					return
				}

				// Set session cookie
				sm.setSessionCookie(w, session.ID)
			}

			// Add session to request context
			ctx := context.WithValue(r.Context(), "session", session)
			r = r.WithContext(ctx)

			// Create response recorder to capture response details
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				responseSize:   0,
			}

			// Process request
			next.ServeHTTP(recorder, r)

			// Record activity
			activity := ActivityEntry{
				Timestamp:    startTime,
				Action:       "request",
				Endpoint:     r.URL.Path,
				Method:       r.Method,
				StatusCode:   recorder.statusCode,
				Duration:     time.Since(startTime),
				RequestSize:  r.ContentLength,
				ResponseSize: recorder.responseSize,
				Metadata: map[string]interface{}{
					"user_agent": r.UserAgent(),
					"referer":    r.Referer(),
				},
			}

			sm.UpdateSessionActivity(session.ID, activity)
		})
	}
}

// getSessionIDFromRequest extracts session ID from request
func (sm *SessionManager) getSessionIDFromRequest(r *http.Request) string {
	// Try to get from cookie first
	if cookie, err := r.Cookie(sm.config.CookieName); err == nil {
		return cookie.Value
	}

	// Try to get from Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		if len(auth) > 7 && auth[:7] == "Bearer " {
			return auth[7:]
		}
	}

	// Try to get from X-Session-ID header
	return r.Header.Get("X-Session-ID")
}

// setSessionCookie sets the session cookie
func (sm *SessionManager) setSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(sm.config.SessionTimeout),
		Secure:   sm.config.CookieSecure,
		HttpOnly: sm.config.CookieHTTPOnly,
		SameSite: sm.config.CookieSameSite,
	}

	http.SetCookie(w, cookie)
}

// responseRecorder captures response details
type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(data []byte) (int, error) {
	size, err := rr.ResponseWriter.Write(data)
	rr.responseSize += int64(size)
	return size, err
}

// GetSessionFromContext retrieves session from request context
func GetSessionFromContext(ctx context.Context) (*UserSession, bool) {
	session, ok := ctx.Value("session").(*UserSession)
	return session, ok
}
