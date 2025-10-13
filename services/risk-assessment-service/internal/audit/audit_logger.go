package audit

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AuditLogger handles audit logging operations
type AuditLogger struct {
	config     *AuditConfig
	logger     *zap.Logger
	repository AuditRepository
	buffer     []AuditEvent
	bufferMux  sync.RWMutex
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// AuditRepository defines the interface for audit data persistence
type AuditRepository interface {
	SaveAuditEvent(ctx context.Context, event *AuditEvent) error
	SaveAuditLog(ctx context.Context, log *AuditLog) error
	GetAuditEvents(ctx context.Context, query AuditQuery) ([]AuditEvent, error)
	GetAuditStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (*AuditStats, error)
	GetAuditLog(ctx context.Context, eventID string) (*AuditLog, error)
	DeleteExpiredLogs(ctx context.Context, before time.Time) error
	GetAuditLogsByHash(ctx context.Context, hash string) ([]AuditLog, error)
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(config *AuditConfig, repository AuditRepository, logger *zap.Logger) *AuditLogger {
	al := &AuditLogger{
		config:     config,
		logger:     logger,
		repository: repository,
		buffer:     make([]AuditEvent, 0, config.BatchSize),
		stopChan:   make(chan struct{}),
	}

	// Start background flush routine
	al.wg.Add(1)
	go al.flushRoutine()

	return al
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) error {
	if !al.config.Enabled {
		return nil
	}

	// Generate event ID if not provided
	if event.ID == "" {
		event.ID = generateEventID()
	}

	// Set timestamps
	now := time.Now()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	event.UpdatedAt = now

	// Generate hash for integrity
	if al.config.EnableHashing {
		event.Hash = al.generateEventHash(event)
	}

	// Add to buffer
	al.bufferMux.Lock()
	al.buffer = append(al.buffer, *event)
	al.bufferMux.Unlock()

	// Flush if buffer is full
	if len(al.buffer) >= al.config.BatchSize {
		al.flushBuffer(ctx)
	}

	return nil
}

// LogRequest logs an HTTP request
func (al *AuditLogger) LogRequest(ctx context.Context, req *AuditEvent) error {
	// Set default values
	if req.Action == "" {
		req.Action = "http_request"
	}
	if req.Resource == "" {
		req.Resource = "api"
	}

	return al.LogEvent(ctx, req)
}

// LogDataAccess logs data access events
func (al *AuditLogger) LogDataAccess(ctx context.Context, tenantID, userID, resource, resourceID, action string, metadata map[string]interface{}) error {
	event := &AuditEvent{
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Metadata:   metadata,
	}

	return al.LogEvent(ctx, event)
}

// LogSecurityEvent logs security-related events
func (al *AuditLogger) LogSecurityEvent(ctx context.Context, tenantID, userID, action string, metadata map[string]interface{}) error {
	event := &AuditEvent{
		TenantID: tenantID,
		UserID:   userID,
		Action:   action,
		Resource: "security",
		Metadata: metadata,
	}

	return al.LogEvent(ctx, event)
}

// LogAdminAction logs administrative actions
func (al *AuditLogger) LogAdminAction(ctx context.Context, tenantID, userID, action, resource, resourceID string, metadata map[string]interface{}) error {
	event := &AuditEvent{
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Metadata:   metadata,
	}

	return al.LogEvent(ctx, event)
}

// GetAuditEvents retrieves audit events based on query
func (al *AuditLogger) GetAuditEvents(ctx context.Context, query AuditQuery) ([]AuditEvent, error) {
	return al.repository.GetAuditEvents(ctx, query)
}

// GetAuditStats retrieves audit statistics
func (al *AuditLogger) GetAuditStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (*AuditStats, error) {
	return al.repository.GetAuditStats(ctx, tenantID, startDate, endDate)
}

// GetAuditLog retrieves an immutable audit log entry
func (al *AuditLogger) GetAuditLog(ctx context.Context, eventID string) (*AuditLog, error) {
	return al.repository.GetAuditLog(ctx, eventID)
}

// VerifyAuditIntegrity verifies the integrity of audit logs
func (al *AuditLogger) VerifyAuditIntegrity(ctx context.Context, eventID string) (bool, error) {
	log, err := al.repository.GetAuditLog(ctx, eventID)
	if err != nil {
		return false, err
	}

	// Deserialize event data
	var event AuditEvent
	if err := json.Unmarshal(log.EventData, &event); err != nil {
		return false, fmt.Errorf("failed to deserialize event data: %w", err)
	}

	// Verify hash
	expectedHash := al.generateEventHash(&event)
	if log.Hash != expectedHash {
		return false, fmt.Errorf("hash mismatch for event %s", eventID)
	}

	return true, nil
}

// flushBuffer flushes the audit buffer to storage
func (al *AuditLogger) flushBuffer(ctx context.Context) {
	al.bufferMux.Lock()
	if len(al.buffer) == 0 {
		al.bufferMux.Unlock()
		return
	}

	// Copy buffer and clear it
	events := make([]AuditEvent, len(al.buffer))
	copy(events, al.buffer)
	al.buffer = al.buffer[:0]
	al.bufferMux.Unlock()

	// Process events in background
	go func() {
		for _, event := range events {
			if err := al.saveEvent(ctx, &event); err != nil {
				al.logger.Error("Failed to save audit event",
					zap.String("event_id", event.ID),
					zap.Error(err))
			}
		}
	}()
}

// saveEvent saves a single audit event
func (al *AuditLogger) saveEvent(ctx context.Context, event *AuditEvent) error {
	// Save to audit events table
	if err := al.repository.SaveAuditEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save audit event: %w", err)
	}

	// Create immutable audit log entry
	auditLog, err := al.createAuditLog(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// Save to audit logs table
	if err := al.repository.SaveAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	return nil
}

// createAuditLog creates an immutable audit log entry
func (al *AuditLogger) createAuditLog(ctx context.Context, event *AuditEvent) (*AuditLog, error) {
	// Serialize event data
	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize event data: %w", err)
	}

	// Generate hash
	hash := al.generateDataHash(eventData)

	// Get previous hash (for blockchain-like integrity)
	prevHash := ""
	if logs, err := al.repository.GetAuditLogsByHash(ctx, ""); err == nil && len(logs) > 0 {
		// Get the most recent log entry
		prevHash = logs[len(logs)-1].Hash
	}

	auditLog := &AuditLog{
		ID:        generateLogID(),
		EventID:   event.ID,
		TenantID:  event.TenantID,
		EventData: eventData,
		Hash:      hash,
		PrevHash:  prevHash,
		CreatedAt: time.Now(),
	}

	return auditLog, nil
}

// flushRoutine runs the background flush routine
func (al *AuditLogger) flushRoutine() {
	defer al.wg.Done()

	ticker := time.NewTicker(al.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			al.flushBuffer(context.Background())
		case <-al.stopChan:
			// Final flush on shutdown
			al.flushBuffer(context.Background())
			return
		}
	}
}

// Stop stops the audit logger
func (al *AuditLogger) Stop() {
	close(al.stopChan)
	al.wg.Wait()
}

// generateEventHash generates a hash for an audit event
func (al *AuditLogger) generateEventHash(event *AuditEvent) string {
	// Create a copy without the hash field for hashing
	eventCopy := *event
	eventCopy.Hash = ""

	data, _ := json.Marshal(eventCopy)
	return al.generateDataHash(data)
}

// generateDataHash generates a SHA-256 hash of data
func (al *AuditLogger) generateDataHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// generateLogID generates a unique log ID
func generateLogID() string {
	return fmt.Sprintf("log_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// AuditMiddleware creates middleware for automatic request logging
func AuditMiddleware(auditLogger *AuditLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract request information
			event := &AuditEvent{
				Method:      r.Method,
				Endpoint:    r.URL.Path,
				IPAddress:   getClientIP(r),
				UserAgent:   r.UserAgent(),
				RequestID:   getRequestID(r),
				RequestSize: r.ContentLength,
			}

			// Extract tenant and user information from context
			if tenantID := r.Context().Value("tenant_id"); tenantID != nil {
				event.TenantID = tenantID.(string)
			}
			if userID := r.Context().Value("user_id"); userID != nil {
				event.UserID = userID.(string)
			}
			if sessionID := r.Context().Value("session_id"); sessionID != nil {
				event.SessionID = sessionID.(string)
			}

			// Wrap response writer to capture status and size
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Complete audit event
			event.Status = wrapped.statusCode
			event.Duration = time.Since(start).Milliseconds()
			event.ResponseSize = wrapped.size

			// Log the event
			if err := auditLogger.LogRequest(r.Context(), event); err != nil {
				// Log error but don't fail the request
				auditLogger.logger.Error("Failed to log audit event", zap.Error(err))
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += int64(n)
	return n, err
}

// Helper functions
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
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

func getRequestID(r *http.Request) string {
	if rid := r.Header.Get("X-Request-ID"); rid != "" {
		return rid
	}
	return generateEventID()
}
