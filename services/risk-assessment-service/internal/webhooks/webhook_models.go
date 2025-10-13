package webhooks

import (
	"errors"
	"time"
)

// WebhookEvent represents a webhook event type
type WebhookEvent string

const (
	// Risk Assessment Events
	EventRiskAssessmentStarted   WebhookEvent = "risk_assessment.started"
	EventRiskAssessmentCompleted WebhookEvent = "risk_assessment.completed"
	EventRiskAssessmentFailed    WebhookEvent = "risk_assessment.failed"

	// Risk Prediction Events
	EventRiskPredictionCompleted WebhookEvent = "risk_prediction.completed"
	EventRiskPredictionFailed    WebhookEvent = "risk_prediction.failed"

	// Batch Job Events
	EventBatchJobStarted   WebhookEvent = "batch_job.started"
	EventBatchJobCompleted WebhookEvent = "batch_job.completed"
	EventBatchJobFailed    WebhookEvent = "batch_job.failed"
	EventBatchJobProgress  WebhookEvent = "batch_job.progress"

	// Custom Model Events
	EventCustomModelCreated   WebhookEvent = "custom_model.created"
	EventCustomModelUpdated   WebhookEvent = "custom_model.updated"
	EventCustomModelDeleted   WebhookEvent = "custom_model.deleted"
	EventCustomModelValidated WebhookEvent = "custom_model.validated"

	// Report Events
	EventReportGenerated WebhookEvent = "report.generated"
	EventReportScheduled WebhookEvent = "report.scheduled"
	EventReportFailed    WebhookEvent = "report.failed"

	// Dashboard Events
	EventDashboardCreated WebhookEvent = "dashboard.created"
	EventDashboardUpdated WebhookEvent = "dashboard.updated"
	EventDashboardShared  WebhookEvent = "dashboard.shared"

	// Data Sync Events
	EventDataSyncStarted   WebhookEvent = "data_sync.started"
	EventDataSyncCompleted WebhookEvent = "data_sync.completed"
	EventDataSyncFailed    WebhookEvent = "data_sync.failed"

	// Risk Threshold Events
	EventRiskThresholdExceeded  WebhookEvent = "risk_threshold.exceeded"
	EventRiskThresholdRecovered WebhookEvent = "risk_threshold.recovered"

	// System Events
	EventSystemMaintenance WebhookEvent = "system.maintenance"
	EventSystemError       WebhookEvent = "system.error"
	EventSystemRecovery    WebhookEvent = "system.recovery"
)

// WebhookStatus represents the status of a webhook
type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusInactive WebhookStatus = "inactive"
	WebhookStatusPaused   WebhookStatus = "paused"
	WebhookStatusDisabled WebhookStatus = "disabled"
)

// DeliveryStatus represents the status of a webhook delivery
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSending   DeliveryStatus = "sending"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusRetrying  DeliveryStatus = "retrying"
	DeliveryStatusCancelled DeliveryStatus = "cancelled"
)

// Error types
var (
	ErrWebhookNotFound   = errors.New("webhook not found")
	ErrDeliveryNotFound  = errors.New("delivery not found")
	ErrInvalidWebhookURL = errors.New("invalid webhook URL")
	ErrWebhookDisabled   = errors.New("webhook is disabled")
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID              string                 `json:"id" db:"id"`
	TenantID        string                 `json:"tenant_id" db:"tenant_id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	URL             string                 `json:"url" db:"url"`
	Events          []WebhookEvent         `json:"events" db:"events"`
	Secret          string                 `json:"secret" db:"secret"`
	Status          WebhookStatus          `json:"status" db:"status"`
	RetryPolicy     RetryPolicy            `json:"retry_policy" db:"retry_policy"`
	RateLimit       RateLimitConfig        `json:"rate_limit" db:"rate_limit"`
	Headers         map[string]string      `json:"headers" db:"headers"`
	Filters         WebhookFilters         `json:"filters" db:"filters"`
	Statistics      WebhookStatistics      `json:"statistics" db:"statistics"`
	CreatedBy       string                 `json:"created_by" db:"created_by"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	LastTriggeredAt *time.Time             `json:"last_triggered_at" db:"last_triggered_at"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

// RetryPolicy represents the retry configuration for webhook deliveries
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialInterval time.Duration `json:"initial_interval"`
	MaxInterval     time.Duration `json:"max_interval"`
	Multiplier      float64       `json:"multiplier"`
	Jitter          bool          `json:"jitter"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `json:"enabled"`
	Requests    int           `json:"requests"`
	Window      time.Duration `json:"window"`
	Burst       int           `json:"burst"`
	SkipOnError bool          `json:"skip_on_error"`
}

// WebhookFilters represents filters for webhook events
type WebhookFilters struct {
	EventTypes    []WebhookEvent         `json:"event_types"`
	BusinessIDs   []string               `json:"business_ids"`
	RiskLevels    []string               `json:"risk_levels"`
	Industries    []string               `json:"industries"`
	Countries     []string               `json:"countries"`
	CustomFilters map[string]interface{} `json:"custom_filters"`
}

// WebhookStatistics represents webhook delivery statistics
type WebhookStatistics struct {
	TotalDeliveries      int64      `json:"total_deliveries"`
	SuccessfulDeliveries int64      `json:"successful_deliveries"`
	FailedDeliveries     int64      `json:"failed_deliveries"`
	SuccessRate          float64    `json:"success_rate"`
	AverageLatency       float64    `json:"average_latency"`
	LastDeliveryAt       *time.Time `json:"last_delivery_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID              string                 `json:"id" db:"id"`
	WebhookID       string                 `json:"webhook_id" db:"webhook_id"`
	TenantID        string                 `json:"tenant_id" db:"tenant_id"`
	EventType       WebhookEvent           `json:"event_type" db:"event_type"`
	EventID         string                 `json:"event_id" db:"event_id"`
	Payload         map[string]interface{} `json:"payload" db:"payload"`
	Headers         map[string]string      `json:"headers" db:"headers"`
	Status          DeliveryStatus         `json:"status" db:"status"`
	Attempts        int                    `json:"attempts" db:"attempts"`
	MaxAttempts     int                    `json:"max_attempts" db:"max_attempts"`
	ResponseCode    *int                   `json:"response_code" db:"response_code"`
	ResponseBody    string                 `json:"response_body" db:"response_body"`
	ResponseHeaders map[string]string      `json:"response_headers" db:"response_headers"`
	Latency         *time.Duration         `json:"latency" db:"latency"`
	Error           string                 `json:"error" db:"error"`
	NextRetryAt     *time.Time             `json:"next_retry_at" db:"next_retry_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	DeliveredAt     *time.Time             `json:"delivered_at" db:"delivered_at"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

// WebhookEvent represents an event that triggers webhooks
type WebhookEventData struct {
	ID         string                 `json:"id"`
	Type       WebhookEvent           `json:"type"`
	TenantID   string                 `json:"tenant_id"`
	BusinessID string                 `json:"business_id,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
	Source     string                 `json:"source"`
	Version    string                 `json:"version"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// WebhookRequest represents a request to create/update a webhook
type WebhookRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description,omitempty"`
	URL         string                 `json:"url" validate:"required,url"`
	Events      []WebhookEvent         `json:"events" validate:"required,min=1"`
	Secret      string                 `json:"secret,omitempty"`
	Status      WebhookStatus          `json:"status,omitempty"`
	RetryPolicy RetryPolicy            `json:"retry_policy,omitempty"`
	RateLimit   RateLimitConfig        `json:"rate_limit,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Filters     WebhookFilters         `json:"filters,omitempty"`
	CreatedBy   string                 `json:"created_by" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookResponse represents a response for webhook operations
type WebhookResponse struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	URL             string            `json:"url"`
	Events          []WebhookEvent    `json:"events"`
	Status          WebhookStatus     `json:"status"`
	Statistics      WebhookStatistics `json:"statistics"`
	CreatedBy       string            `json:"created_by"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	LastTriggeredAt *time.Time        `json:"last_triggered_at"`
}

// WebhookListResponse represents a list response for webhooks
type WebhookListResponse struct {
	Webhooks []WebhookResponse `json:"webhooks"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// WebhookFilter represents filters for querying webhooks
type WebhookFilter struct {
	TenantID  string        `json:"tenant_id,omitempty"`
	Status    WebhookStatus `json:"status,omitempty"`
	EventType WebhookEvent  `json:"event_type,omitempty"`
	CreatedBy string        `json:"created_by,omitempty"`
	StartDate *time.Time    `json:"start_date,omitempty"`
	EndDate   *time.Time    `json:"end_date,omitempty"`
	Limit     int           `json:"limit,omitempty"`
	Offset    int           `json:"offset,omitempty"`
}

// DeliveryListResponse represents a list response for webhook deliveries
type DeliveryListResponse struct {
	Deliveries []WebhookDelivery `json:"deliveries"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// DeliveryFilter represents filters for querying webhook deliveries
type DeliveryFilter struct {
	WebhookID string         `json:"webhook_id,omitempty"`
	TenantID  string         `json:"tenant_id,omitempty"`
	Status    DeliveryStatus `json:"status,omitempty"`
	EventType WebhookEvent   `json:"event_type,omitempty"`
	StartDate *time.Time     `json:"start_date,omitempty"`
	EndDate   *time.Time     `json:"end_date,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

// WebhookTestRequest represents a request to test a webhook
type WebhookTestRequest struct {
	EventType WebhookEvent           `json:"event_type" validate:"required"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
}

// WebhookTestResponse represents a response for webhook testing
type WebhookTestResponse struct {
	Success      bool          `json:"success"`
	ResponseCode int           `json:"response_code"`
	ResponseBody string        `json:"response_body"`
	Latency      time.Duration `json:"latency"`
	Error        string        `json:"error,omitempty"`
	DeliveredAt  time.Time     `json:"delivered_at"`
}

// WebhookMetrics represents webhook usage metrics
type WebhookMetrics struct {
	TotalWebhooks        int                      `json:"total_webhooks"`
	ActiveWebhooks       int                      `json:"active_webhooks"`
	TotalDeliveries      int64                    `json:"total_deliveries"`
	SuccessfulDeliveries int64                    `json:"successful_deliveries"`
	FailedDeliveries     int64                    `json:"failed_deliveries"`
	SuccessRate          float64                  `json:"success_rate"`
	AverageLatency       time.Duration            `json:"average_latency"`
	DeliveriesByEvent    map[WebhookEvent]int64   `json:"deliveries_by_event"`
	DeliveriesByStatus   map[DeliveryStatus]int64 `json:"deliveries_by_status"`
	TopWebhooks          []WebhookUsageData       `json:"top_webhooks"`
}

// WebhookUsageData represents webhook usage statistics
type WebhookUsageData struct {
	WebhookID   string    `json:"webhook_id"`
	WebhookName string    `json:"webhook_name"`
	URL         string    `json:"url"`
	Deliveries  int64     `json:"deliveries"`
	SuccessRate float64   `json:"success_rate"`
	LastUsed    time.Time `json:"last_used"`
}

// WebhookHealth represents the health status of a webhook
type WebhookHealth struct {
	WebhookID        string     `json:"webhook_id"`
	Status           string     `json:"status"` // "healthy", "degraded", "unhealthy"
	LastSuccessful   *time.Time `json:"last_successful"`
	LastFailed       *time.Time `json:"last_failed"`
	ConsecutiveFails int        `json:"consecutive_fails"`
	AverageLatency   float64    `json:"average_latency"`
	SuccessRate      float64    `json:"success_rate"`
	NextRetryAt      *time.Time `json:"next_retry_at"`
}

// WebhookSignature represents webhook signature verification
type WebhookSignature struct {
	Algorithm string `json:"algorithm"` // "sha256", "sha1"
	Signature string `json:"signature"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
}

// WebhookPayload represents the payload sent to webhook endpoints
type WebhookPayload struct {
	ID        string                 `json:"id"`
	Type      WebhookEvent           `json:"type"`
	Version   string                 `json:"version"`
	Created   time.Time              `json:"created"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
	Signature WebhookSignature       `json:"signature,omitempty"`
}

// WebhookQueue represents a webhook delivery queue item
type WebhookQueue struct {
	ID          string                 `json:"id"`
	WebhookID   string                 `json:"webhook_id"`
	EventData   WebhookEventData       `json:"event_data"`
	Priority    int                    `json:"priority"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WebhookCircuitBreakerState represents circuit breaker state for webhook delivery
type WebhookCircuitBreakerState struct {
	WebhookID        string        `json:"webhook_id"`
	State            string        `json:"state"` // "closed", "open", "half_open"
	FailureCount     int           `json:"failure_count"`
	LastFailureTime  *time.Time    `json:"last_failure_time"`
	NextAttemptTime  *time.Time    `json:"next_attempt_time"`
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
}

// WebhookRateLimiterState represents rate limiter state for webhook delivery
type WebhookRateLimiterState struct {
	WebhookID    string        `json:"webhook_id"`
	Requests     int           `json:"requests"`
	WindowStart  time.Time     `json:"window_start"`
	WindowSize   time.Duration `json:"window_size"`
	MaxRequests  int           `json:"max_requests"`
	Burst        int           `json:"burst"`
	CurrentBurst int           `json:"current_burst"`
}

// WebhookRetryInfo represents retry information for a webhook delivery
type WebhookRetryInfo struct {
	DeliveryID       string        `json:"delivery_id"`
	Attempts         int           `json:"attempts"`
	MaxAttempts      int           `json:"max_attempts"`
	NextRetryAt      time.Time     `json:"next_retry_at"`
	RetryInterval    time.Duration `json:"retry_interval"`
	BackoffFactor    float64       `json:"backoff_factor"`
	Jitter           bool          `json:"jitter"`
	LastError        string        `json:"last_error"`
	ConsecutiveFails int           `json:"consecutive_fails"`
}

// WebhookEventFilterConfig represents event filtering configuration
type WebhookEventFilterConfig struct {
	EventType  WebhookEvent      `json:"event_type"`
	Conditions []FilterCondition `json:"conditions"`
	Logic      string            `json:"logic"` // "and", "or"
	Enabled    bool              `json:"enabled"`
}

// FilterCondition represents a condition in an event filter
type FilterCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // "eq", "ne", "gt", "lt", "gte", "lte", "in", "not_in", "contains", "regex"
	Value    interface{} `json:"value"`
}

// WebhookTemplate represents a webhook template for common configurations
type WebhookTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Events      []WebhookEvent         `json:"events"`
	URLTemplate string                 `json:"url_template"`
	Headers     map[string]string      `json:"headers"`
	RetryPolicy RetryPolicy            `json:"retry_policy"`
	RateLimit   RateLimitConfig        `json:"rate_limit"`
	Filters     WebhookFilters         `json:"filters"`
	IsPublic    bool                   `json:"is_public"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WebhookTemplateRequest represents a request to create/update a webhook template
type WebhookTemplateRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description,omitempty"`
	Events      []WebhookEvent         `json:"events" validate:"required,min=1"`
	URLTemplate string                 `json:"url_template" validate:"required"`
	Headers     map[string]string      `json:"headers,omitempty"`
	RetryPolicy RetryPolicy            `json:"retry_policy,omitempty"`
	RateLimit   RateLimitConfig        `json:"rate_limit,omitempty"`
	Filters     WebhookFilters         `json:"filters,omitempty"`
	IsPublic    bool                   `json:"is_public,omitempty"`
	CreatedBy   string                 `json:"created_by" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookTemplateResponse represents a response for webhook template operations
type WebhookTemplateResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Events      []WebhookEvent `json:"events"`
	URLTemplate string         `json:"url_template"`
	IsPublic    bool           `json:"is_public"`
	CreatedBy   string         `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// WebhookTemplateListResponse represents a list response for webhook templates
type WebhookTemplateListResponse struct {
	Templates []WebhookTemplateResponse `json:"templates"`
	Total     int                       `json:"total"`
	Page      int                       `json:"page"`
	PageSize  int                       `json:"page_size"`
}

// WebhookTemplateFilter represents filters for querying webhook templates
type WebhookTemplateFilter struct {
	IsPublic  *bool        `json:"is_public,omitempty"`
	EventType WebhookEvent `json:"event_type,omitempty"`
	CreatedBy string       `json:"created_by,omitempty"`
	Limit     int          `json:"limit,omitempty"`
	Offset    int          `json:"offset,omitempty"`
}
