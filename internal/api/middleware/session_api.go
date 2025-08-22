package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// SessionAPI provides HTTP endpoints for session management
type SessionAPI struct {
	sessionManager *SessionManager
}

// NewSessionAPI creates a new session API
func NewSessionAPI(sessionManager *SessionManager) *SessionAPI {
	return &SessionAPI{
		sessionManager: sessionManager,
	}
}

// SessionResponse represents a session in API responses
type SessionResponse struct {
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
	SessionDuration  time.Duration          `json:"session_duration"`
}

// SessionListResponse represents a list of sessions
type SessionListResponse struct {
	Success  bool              `json:"success"`
	Sessions []SessionResponse `json:"sessions"`
	Count    int               `json:"count"`
	Error    string            `json:"error,omitempty"`
}

// SessionMetricsResponse represents session metrics
type SessionMetricsResponse struct {
	Success bool            `json:"success"`
	Metrics *SessionMetrics `json:"metrics"`
	Error   string          `json:"error,omitempty"`
}

// SessionDetailResponse represents detailed session information
type SessionDetailResponse struct {
	Success bool             `json:"success"`
	Session *SessionResponse `json:"session,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// CreateSessionRequest represents a request to create a session
type CreateSessionRequest struct {
	UserID   string                 `json:"user_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateSessionResponse represents a response to session creation
type CreateSessionResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
}

// GetSessionsHandler returns all active sessions
func (api *SessionAPI) GetSessionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	userID := r.URL.Query().Get("user_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Parse limit and offset
	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get sessions
	var sessions []*UserSession
	if userID != "" {
		sessions = api.sessionManager.GetSessionsByUser(userID)
	} else {
		sessions = api.sessionManager.GetActiveSessions()
	}

	// Apply pagination
	totalCount := len(sessions)
	if offset >= totalCount {
		sessions = []*UserSession{}
	} else {
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		sessions = sessions[offset:end]
	}

	// Convert to response format
	sessionResponses := make([]SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = api.convertToSessionResponse(session)
	}

	response := SessionListResponse{
		Success:  true,
		Sessions: sessionResponses,
		Count:    len(sessionResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSessionHandler returns details for a specific session
func (api *SessionAPI) GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL path
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		response := SessionDetailResponse{
			Success: false,
			Error:   "session_id parameter is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get session
	session, exists := api.sessionManager.GetSession(sessionID)
	if !exists {
		response := SessionDetailResponse{
			Success: false,
			Error:   "session not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert to response format
	sessionResponse := api.convertToSessionResponse(session)

	response := SessionDetailResponse{
		Success: true,
		Session: &sessionResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateSessionHandler creates a new session
func (api *SessionAPI) CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := CreateSessionResponse{
			Success: false,
			Error:   "Invalid JSON in request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create session
	session, err := api.sessionManager.CreateSession(r, req.UserID)
	if err != nil {
		response := CreateSessionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create session: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Add metadata if provided
	if req.Metadata != nil {
		session.mu.Lock()
		for key, value := range req.Metadata {
			session.Metadata[key] = value
		}
		session.mu.Unlock()
	}

	response := CreateSessionResponse{
		Success:   true,
		SessionID: session.ID,
		Message:   "Session created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// DeleteSessionHandler deletes a session
func (api *SessionAPI) DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL query
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		response := map[string]interface{}{
			"success": false,
			"error":   "session_id parameter is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Delete session
	deleted := api.sessionManager.DeleteSession(sessionID)
	if !deleted {
		response := map[string]interface{}{
			"success": false,
			"error":   "session not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Session deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSessionMetricsHandler returns session metrics
func (api *SessionAPI) GetSessionMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get metrics
	metrics := api.sessionManager.GetMetrics()

	response := SessionMetricsResponse{
		Success: true,
		Metrics: metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetCurrentSessionHandler returns the current user's session
func (api *SessionAPI) GetCurrentSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session from context
	session, exists := GetSessionFromContext(r.Context())
	if !exists {
		response := SessionDetailResponse{
			Success: false,
			Error:   "No active session found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert to response format
	sessionResponse := api.convertToSessionResponse(session)

	response := SessionDetailResponse{
		Success: true,
		Session: &sessionResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSessionActivityHandler returns activity log for a session
func (api *SessionAPI) GetSessionActivityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL query
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		// Try to get current session from context
		if session, exists := GetSessionFromContext(r.Context()); exists {
			sessionID = session.ID
		} else {
			response := map[string]interface{}{
				"success": false,
				"error":   "session_id parameter is required or no active session",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Get session
	session, exists := api.sessionManager.GetSession(sessionID)
	if !exists {
		response := map[string]interface{}{
			"success": false,
			"error":   "session not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get activity log with pagination
	session.mu.RLock()
	activityLog := session.ActivityLog
	totalCount := len(activityLog)

	var paginatedActivity []ActivityEntry
	if offset >= totalCount {
		paginatedActivity = []ActivityEntry{}
	} else {
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		paginatedActivity = activityLog[offset:end]
	}
	session.mu.RUnlock()

	response := map[string]interface{}{
		"success":     true,
		"session_id":  sessionID,
		"activity":    paginatedActivity,
		"count":       len(paginatedActivity),
		"total_count": totalCount,
		"offset":      offset,
		"limit":       limit,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSessionStatusHandler returns session management system status
func (api *SessionAPI) GetSessionStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current metrics
	metrics := api.sessionManager.GetMetrics()

	// Get active sessions for additional stats
	activeSessions := api.sessionManager.GetActiveSessions()

	// Calculate additional statistics
	var totalSessionDuration time.Duration
	userSessions := make(map[string]int)
	for _, session := range activeSessions {
		sessionDuration := session.LastAccessTime.Sub(session.CreatedAt)
		totalSessionDuration += sessionDuration

		if session.UserID != "" {
			userSessions[session.UserID]++
		}
	}

	averageSessionDuration := time.Duration(0)
	if len(activeSessions) > 0 {
		averageSessionDuration = totalSessionDuration / time.Duration(len(activeSessions))
	}

	status := map[string]interface{}{
		"success": true,
		"status":  "operational",
		"features": map[string]interface{}{
			"session_management": true,
			"session_tracking":   true,
			"activity_logging":   api.sessionManager.config.TrackUserActivity,
			"metrics_collection": api.sessionManager.config.EnableMetrics,
			"automatic_cleanup":  true,
		},
		"configuration": map[string]interface{}{
			"session_timeout":  api.sessionManager.config.SessionTimeout.String(),
			"cleanup_interval": api.sessionManager.config.CleanupInterval.String(),
			"max_sessions":     api.sessionManager.config.MaxSessions,
			"cookie_name":      api.sessionManager.config.CookieName,
			"cookie_secure":    api.sessionManager.config.CookieSecure,
			"cookie_http_only": api.sessionManager.config.CookieHTTPOnly,
		},
		"current_metrics": map[string]interface{}{
			"active_sessions":          metrics.ActiveSessions,
			"total_sessions":           metrics.TotalSessions,
			"expired_sessions":         metrics.ExpiredSessions,
			"total_requests":           metrics.TotalRequests,
			"requests_per_session":     metrics.RequestsPerSession,
			"peak_sessions":            metrics.PeakSessions,
			"peak_sessions_time":       metrics.PeakSessionsTime.Format(time.RFC3339),
			"average_session_duration": averageSessionDuration.String(),
		},
		"statistics": map[string]interface{}{
			"unique_users":       len(userSessions),
			"anonymous_sessions": int64(len(activeSessions)) - int64(len(userSessions)),
			"sessions_by_hour":   metrics.SessionsByHour,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// convertToSessionResponse converts UserSession to SessionResponse
func (api *SessionAPI) convertToSessionResponse(session *UserSession) SessionResponse {
	session.mu.RLock()
	defer session.mu.RUnlock()

	sessionDuration := session.LastAccessTime.Sub(session.CreatedAt)

	return SessionResponse{
		ID:               session.ID,
		UserID:           session.UserID,
		IPAddress:        session.IPAddress,
		UserAgent:        session.UserAgent,
		CreatedAt:        session.CreatedAt,
		LastAccessTime:   session.LastAccessTime,
		LastActivityTime: session.LastActivityTime,
		RequestCount:     session.RequestCount,
		IsActive:         session.IsActive,
		ExpiresAt:        session.ExpiresAt,
		Metadata:         session.Metadata,
		SessionDuration:  sessionDuration,
	}
}

// RegisterSessionRoutes registers all session management routes
func (api *SessionAPI) RegisterSessionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/sessions", api.GetSessionsHandler)
	mux.HandleFunc("GET /v1/sessions/current", api.GetCurrentSessionHandler)
	mux.HandleFunc("GET /v1/sessions/detail", api.GetSessionHandler)
	mux.HandleFunc("POST /v1/sessions", api.CreateSessionHandler)
	mux.HandleFunc("DELETE /v1/sessions", api.DeleteSessionHandler)
	mux.HandleFunc("GET /v1/sessions/metrics", api.GetSessionMetricsHandler)
	mux.HandleFunc("GET /v1/sessions/activity", api.GetSessionActivityHandler)
	mux.HandleFunc("GET /v1/sessions/status", api.GetSessionStatusHandler)
}
