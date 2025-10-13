package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WebhookManager manages webhook configurations and deliveries
type WebhookManager interface {
	// Webhook Management
	CreateWebhook(ctx context.Context, request *WebhookRequest) (*WebhookResponse, error)
	GetWebhook(ctx context.Context, tenantID, webhookID string) (*Webhook, error)
	ListWebhooks(ctx context.Context, filter *WebhookFilter) (*WebhookListResponse, error)
	UpdateWebhook(ctx context.Context, tenantID, webhookID string, request *WebhookRequest) (*WebhookResponse, error)
	DeleteWebhook(ctx context.Context, tenantID, webhookID string) error

	// Webhook Operations
	TestWebhook(ctx context.Context, tenantID, webhookID string, request *WebhookTestRequest) (*WebhookTestResponse, error)
	EnableWebhook(ctx context.Context, tenantID, webhookID string) error
	DisableWebhook(ctx context.Context, tenantID, webhookID string) error

	// Event Processing
	ProcessEvent(ctx context.Context, event *WebhookEventData) error
	TriggerWebhook(ctx context.Context, webhookID string, event *WebhookEventData) error

	// Delivery Management
	GetDeliveries(ctx context.Context, filter *DeliveryFilter) (*DeliveryListResponse, error)
	GetDelivery(ctx context.Context, tenantID, deliveryID string) (*WebhookDelivery, error)
	RetryDelivery(ctx context.Context, tenantID, deliveryID string) error
	CancelDelivery(ctx context.Context, tenantID, deliveryID string) error

	// Metrics and Health
	GetWebhookMetrics(ctx context.Context, tenantID string) (*WebhookMetrics, error)
	GetWebhookHealth(ctx context.Context, tenantID, webhookID string) (*WebhookHealth, error)

	// Template Management
	CreateTemplate(ctx context.Context, request *WebhookTemplateRequest) (*WebhookTemplateResponse, error)
	GetTemplate(ctx context.Context, templateID string) (*WebhookTemplate, error)
	ListTemplates(ctx context.Context, filter *WebhookTemplateFilter) (*WebhookTemplateListResponse, error)
	UpdateTemplate(ctx context.Context, templateID string, request *WebhookTemplateRequest) (*WebhookTemplateResponse, error)
	DeleteTemplate(ctx context.Context, templateID string) error
}

// DefaultWebhookManager implements WebhookManager
type DefaultWebhookManager struct {
	repository        WebhookRepository
	deliveryTracker   WebhookDeliveryTracker
	retryHandler      WebhookRetryHandler
	signatureVerifier WebhookSignatureVerifier
	rateLimiter       WebhookRateLimiter
	circuitBreaker    WebhookCircuitBreaker
	eventFilter       WebhookEventFilter
	httpClient        *http.Client
	logger            *zap.Logger
	mu                sync.RWMutex
}

// WebhookRepository defines the interface for webhook data persistence
type WebhookRepository interface {
	// Webhook CRUD
	SaveWebhook(ctx context.Context, webhook *Webhook) error
	GetWebhook(ctx context.Context, tenantID, webhookID string) (*Webhook, error)
	ListWebhooks(ctx context.Context, filter *WebhookFilter) ([]*Webhook, error)
	UpdateWebhook(ctx context.Context, webhook *Webhook) error
	DeleteWebhook(ctx context.Context, tenantID, webhookID string) error

	// Webhook Statistics
	UpdateWebhookStatistics(ctx context.Context, webhookID string, stats *WebhookStatistics) error
	GetWebhookStatistics(ctx context.Context, webhookID string) (*WebhookStatistics, error)

	// Delivery Management
	SaveDelivery(ctx context.Context, delivery *WebhookDelivery) error
	GetDelivery(ctx context.Context, tenantID, deliveryID string) (*WebhookDelivery, error)
	ListDeliveries(ctx context.Context, filter *DeliveryFilter) ([]*WebhookDelivery, error)
	UpdateDelivery(ctx context.Context, delivery *WebhookDelivery) error

	// Template Management
	SaveTemplate(ctx context.Context, template *WebhookTemplate) error
	GetTemplate(ctx context.Context, templateID string) (*WebhookTemplate, error)
	ListTemplates(ctx context.Context, filter *WebhookTemplateFilter) ([]*WebhookTemplate, error)
	UpdateTemplate(ctx context.Context, template *WebhookTemplate) error
	DeleteTemplate(ctx context.Context, templateID string) error

	// Metrics
	GetWebhookMetrics(ctx context.Context, tenantID string) (*WebhookMetrics, error)
	GetWebhookHealth(ctx context.Context, tenantID, webhookID string) (*WebhookHealth, error)
}

// WebhookDeliveryTracker defines the interface for tracking webhook deliveries
type WebhookDeliveryTracker interface {
	TrackDelivery(ctx context.Context, delivery *WebhookDelivery) error
	UpdateDeliveryStatus(ctx context.Context, deliveryID string, status DeliveryStatus, responseCode int, responseBody string, latency time.Duration, error string) error
	GetDeliveryHistory(ctx context.Context, webhookID string, limit int) ([]*WebhookDelivery, error)
	GetFailedDeliveries(ctx context.Context, webhookID string) ([]*WebhookDelivery, error)
}

// WebhookRetryHandler defines the interface for handling webhook retries
type WebhookRetryHandler interface {
	ScheduleRetry(ctx context.Context, delivery *WebhookDelivery) error
	ProcessRetries(ctx context.Context) error
	GetRetryInfo(ctx context.Context, deliveryID string) (*WebhookRetryInfo, error)
	CancelRetry(ctx context.Context, deliveryID string) error
}

// WebhookSignatureVerifier defines the interface for webhook signature verification
type WebhookSignatureVerifier interface {
	GenerateSignature(payload []byte, secret string, timestamp string) (string, error)
	VerifySignature(payload []byte, signature string, secret string, timestamp string) (bool, error)
	ValidateTimestamp(timestamp string, tolerance time.Duration) (bool, error)
}

// WebhookRateLimiter defines the interface for webhook rate limiting
type WebhookRateLimiter interface {
	AllowRequest(ctx context.Context, webhookID string) (bool, error)
	GetRateLimitStatus(ctx context.Context, webhookID string) (*WebhookRateLimiterState, error)
	ResetRateLimit(ctx context.Context, webhookID string) error
}

// WebhookCircuitBreaker defines the interface for webhook circuit breaker
type WebhookCircuitBreaker interface {
	AllowRequest(ctx context.Context, webhookID string) (bool, error)
	RecordSuccess(ctx context.Context, webhookID string) error
	RecordFailure(ctx context.Context, webhookID string) error
	GetCircuitBreakerState(ctx context.Context, webhookID string) (*WebhookCircuitBreakerState, error)
}

// WebhookEventFilter defines the interface for filtering webhook events
type WebhookEventFilter interface {
	ShouldDeliver(ctx context.Context, webhook *Webhook, event *WebhookEventData) (bool, error)
	ApplyFilters(ctx context.Context, webhook *Webhook, event *WebhookEventData) (*WebhookEventData, error)
}

// NewDefaultWebhookManager creates a new default webhook manager
func NewDefaultWebhookManager(
	repository WebhookRepository,
	deliveryTracker WebhookDeliveryTracker,
	retryHandler WebhookRetryHandler,
	signatureVerifier WebhookSignatureVerifier,
	rateLimiter WebhookRateLimiter,
	circuitBreaker WebhookCircuitBreaker,
	eventFilter WebhookEventFilter,
	logger *zap.Logger,
) *DefaultWebhookManager {
	return &DefaultWebhookManager{
		repository:        repository,
		deliveryTracker:   deliveryTracker,
		retryHandler:      retryHandler,
		signatureVerifier: signatureVerifier,
		rateLimiter:       rateLimiter,
		circuitBreaker:    circuitBreaker,
		eventFilter:       eventFilter,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CreateWebhook creates a new webhook
func (wm *DefaultWebhookManager) CreateWebhook(ctx context.Context, request *WebhookRequest) (*WebhookResponse, error) {
	wm.logger.Info("Creating webhook",
		zap.String("name", request.Name),
		zap.String("url", request.URL),
		zap.String("created_by", request.CreatedBy))

	// Generate webhook ID
	webhookID := generateWebhookID()

	// Create webhook
	webhook := &Webhook{
		ID:          webhookID,
		TenantID:    getTenantIDFromContext(ctx),
		Name:        request.Name,
		Description: request.Description,
		URL:         request.URL,
		Events:      request.Events,
		Secret:      request.Secret,
		Status:      WebhookStatusActive,
		RetryPolicy: request.RetryPolicy,
		RateLimit:   request.RateLimit,
		Headers:     request.Headers,
		Filters:     request.Filters,
		Statistics:  WebhookStatistics{},
		CreatedBy:   request.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    request.Metadata,
	}

	// Set default retry policy if not provided
	if webhook.RetryPolicy.MaxRetries == 0 {
		webhook.RetryPolicy = RetryPolicy{
			MaxRetries:      3,
			InitialInterval: 1 * time.Second,
			MaxInterval:     60 * time.Second,
			Multiplier:      2.0,
			Jitter:          true,
		}
	}

	// Set default rate limit if not provided
	if !webhook.RateLimit.Enabled {
		webhook.RateLimit = RateLimitConfig{
			Enabled:     true,
			Requests:    100,
			Window:      1 * time.Minute,
			Burst:       10,
			SkipOnError: false,
		}
	}

	// Save webhook
	if err := wm.repository.SaveWebhook(ctx, webhook); err != nil {
		return nil, fmt.Errorf("failed to save webhook: %w", err)
	}

	response := &WebhookResponse{
		ID:              webhook.ID,
		Name:            webhook.Name,
		Description:     webhook.Description,
		URL:             webhook.URL,
		Events:          webhook.Events,
		Status:          webhook.Status,
		Statistics:      webhook.Statistics,
		CreatedBy:       webhook.CreatedBy,
		CreatedAt:       webhook.CreatedAt,
		UpdatedAt:       webhook.UpdatedAt,
		LastTriggeredAt: webhook.LastTriggeredAt,
	}

	wm.logger.Info("Webhook created successfully",
		zap.String("webhook_id", webhookID),
		zap.String("name", request.Name))

	return response, nil
}

// GetWebhook retrieves a webhook by ID
func (wm *DefaultWebhookManager) GetWebhook(ctx context.Context, tenantID, webhookID string) (*Webhook, error) {
	wm.logger.Debug("Getting webhook",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	webhook, err := wm.repository.GetWebhook(ctx, tenantID, webhookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return nil, fmt.Errorf("webhook not found: %s", webhookID)
	}

	wm.logger.Debug("Webhook retrieved successfully",
		zap.String("webhook_id", webhookID))

	return webhook, nil
}

// ListWebhooks lists webhooks with filters
func (wm *DefaultWebhookManager) ListWebhooks(ctx context.Context, filter *WebhookFilter) (*WebhookListResponse, error) {
	wm.logger.Debug("Listing webhooks",
		zap.String("tenant_id", filter.TenantID))

	webhooks, err := wm.repository.ListWebhooks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	// Convert to response format
	responses := make([]WebhookResponse, len(webhooks))
	for i, webhook := range webhooks {
		responses[i] = WebhookResponse{
			ID:              webhook.ID,
			Name:            webhook.Name,
			Description:     webhook.Description,
			URL:             webhook.URL,
			Events:          webhook.Events,
			Status:          webhook.Status,
			Statistics:      webhook.Statistics,
			CreatedBy:       webhook.CreatedBy,
			CreatedAt:       webhook.CreatedAt,
			UpdatedAt:       webhook.UpdatedAt,
			LastTriggeredAt: webhook.LastTriggeredAt,
		}
	}

	response := &WebhookListResponse{
		Webhooks: responses,
		Total:    len(responses),
		Page:     1, // This would be calculated based on offset/limit
		PageSize: len(responses),
	}

	wm.logger.Debug("Webhooks listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(responses)))

	return response, nil
}

// UpdateWebhook updates an existing webhook
func (wm *DefaultWebhookManager) UpdateWebhook(ctx context.Context, tenantID, webhookID string, request *WebhookRequest) (*WebhookResponse, error) {
	wm.logger.Info("Updating webhook",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	// Get existing webhook
	webhook, err := wm.repository.GetWebhook(ctx, tenantID, webhookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return nil, fmt.Errorf("webhook not found: %s", webhookID)
	}

	// Update webhook fields
	webhook.Name = request.Name
	webhook.Description = request.Description
	webhook.URL = request.URL
	webhook.Events = request.Events
	webhook.Secret = request.Secret
	webhook.RetryPolicy = request.RetryPolicy
	webhook.RateLimit = request.RateLimit
	webhook.Headers = request.Headers
	webhook.Filters = request.Filters
	webhook.Metadata = request.Metadata
	webhook.UpdatedAt = time.Now()

	// Save updated webhook
	if err := wm.repository.UpdateWebhook(ctx, webhook); err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	response := &WebhookResponse{
		ID:              webhook.ID,
		Name:            webhook.Name,
		Description:     webhook.Description,
		URL:             webhook.URL,
		Events:          webhook.Events,
		Status:          webhook.Status,
		Statistics:      webhook.Statistics,
		CreatedBy:       webhook.CreatedBy,
		CreatedAt:       webhook.CreatedAt,
		UpdatedAt:       webhook.UpdatedAt,
		LastTriggeredAt: webhook.LastTriggeredAt,
	}

	wm.logger.Info("Webhook updated successfully",
		zap.String("webhook_id", webhookID))

	return response, nil
}

// DeleteWebhook deletes a webhook
func (wm *DefaultWebhookManager) DeleteWebhook(ctx context.Context, tenantID, webhookID string) error {
	wm.logger.Info("Deleting webhook",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	if err := wm.repository.DeleteWebhook(ctx, tenantID, webhookID); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	wm.logger.Info("Webhook deleted successfully",
		zap.String("webhook_id", webhookID))

	return nil
}

// TestWebhook tests a webhook with a sample payload
func (wm *DefaultWebhookManager) TestWebhook(ctx context.Context, tenantID, webhookID string, request *WebhookTestRequest) (*WebhookTestResponse, error) {
	wm.logger.Info("Testing webhook",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	// Get webhook
	webhook, err := wm.repository.GetWebhook(ctx, tenantID, webhookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return nil, fmt.Errorf("webhook not found: %s", webhookID)
	}

	// Create test event data
	eventData := &WebhookEventData{
		ID:        generateEventID(),
		Type:      request.EventType,
		TenantID:  tenantID,
		Data:      request.Payload,
		Timestamp: time.Now(),
		Source:    "webhook_test",
		Version:   "1.0",
		Metadata:  map[string]interface{}{"test": true},
	}

	// Send test webhook
	start := time.Now()
	success, responseCode, responseBody, err := wm.sendWebhook(ctx, webhook, eventData, request.Headers)
	latency := time.Since(start)

	response := &WebhookTestResponse{
		Success:      success,
		ResponseCode: responseCode,
		ResponseBody: responseBody,
		Latency:      latency,
		DeliveredAt:  time.Now(),
	}

	if err != nil {
		response.Error = err.Error()
	}

	wm.logger.Info("Webhook test completed",
		zap.String("webhook_id", webhookID),
		zap.Bool("success", success),
		zap.Int("response_code", responseCode),
		zap.Duration("latency", latency))

	return response, nil
}

// EnableWebhook enables a webhook
func (wm *DefaultWebhookManager) EnableWebhook(ctx context.Context, tenantID, webhookID string) error {
	wm.logger.Info("Enabling webhook",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	webhook, err := wm.repository.GetWebhook(ctx, tenantID, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	webhook.Status = WebhookStatusActive
	webhook.UpdatedAt = time.Now()

	if err := wm.repository.UpdateWebhook(ctx, webhook); err != nil {
		return fmt.Errorf("failed to enable webhook: %w", err)
	}

	wm.logger.Info("Webhook enabled successfully",
		zap.String("webhook_id", webhookID))

	return nil
}

// DisableWebhook disables a webhook
func (wm *DefaultWebhookManager) DisableWebhook(ctx context.Context, tenantID, webhookID string) error {
	wm.logger.Info("Disabling webhook",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	webhook, err := wm.repository.GetWebhook(ctx, tenantID, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	webhook.Status = WebhookStatusInactive
	webhook.UpdatedAt = time.Now()

	if err := wm.repository.UpdateWebhook(ctx, webhook); err != nil {
		return fmt.Errorf("failed to disable webhook: %w", err)
	}

	wm.logger.Info("Webhook disabled successfully",
		zap.String("webhook_id", webhookID))

	return nil
}

// ProcessEvent processes a webhook event and triggers relevant webhooks
func (wm *DefaultWebhookManager) ProcessEvent(ctx context.Context, event *WebhookEventData) error {
	wm.logger.Info("Processing webhook event",
		zap.String("event_type", string(event.Type)),
		zap.String("event_id", event.ID),
		zap.String("tenant_id", event.TenantID))

	// Find webhooks that should receive this event
	filter := &WebhookFilter{
		TenantID:  event.TenantID,
		Status:    WebhookStatusActive,
		EventType: event.Type,
	}

	webhooks, err := wm.repository.ListWebhooks(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list webhooks for event: %w", err)
	}

	// Process each webhook
	for _, webhook := range webhooks {
		// Check if webhook should receive this event
		shouldDeliver, err := wm.eventFilter.ShouldDeliver(ctx, webhook, event)
		if err != nil {
			wm.logger.Error("Failed to check event filter", zap.Error(err))
			continue
		}

		if !shouldDeliver {
			continue
		}

		// Trigger webhook asynchronously
		go wm.triggerWebhookAsync(context.Background(), webhook, event)
	}

	wm.logger.Info("Webhook event processed",
		zap.String("event_type", string(event.Type)),
		zap.String("event_id", event.ID),
		zap.Int("webhooks_triggered", len(webhooks)))

	return nil
}

// TriggerWebhook triggers a specific webhook
func (wm *DefaultWebhookManager) TriggerWebhook(ctx context.Context, webhookID string, event *WebhookEventData) error {
	wm.logger.Info("Triggering webhook",
		zap.String("webhook_id", webhookID),
		zap.String("event_id", event.ID))

	webhook, err := wm.repository.GetWebhook(ctx, event.TenantID, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	return wm.triggerWebhookAsync(ctx, webhook, event)
}

// triggerWebhookAsync triggers a webhook asynchronously
func (wm *DefaultWebhookManager) triggerWebhookAsync(ctx context.Context, webhook *Webhook, event *WebhookEventData) error {
	wm.logger.Debug("Triggering webhook asynchronously",
		zap.String("webhook_id", webhook.ID),
		zap.String("event_id", event.ID))

	// Check circuit breaker
	allowed, err := wm.circuitBreaker.AllowRequest(ctx, webhook.ID)
	if err != nil {
		wm.logger.Error("Failed to check circuit breaker", zap.Error(err))
		return err
	}

	if !allowed {
		wm.logger.Warn("Webhook request blocked by circuit breaker",
			zap.String("webhook_id", webhook.ID))
		return fmt.Errorf("webhook blocked by circuit breaker")
	}

	// Check rate limit
	allowed, err = wm.rateLimiter.AllowRequest(ctx, webhook.ID)
	if err != nil {
		wm.logger.Error("Failed to check rate limit", zap.Error(err))
		return err
	}

	if !allowed {
		wm.logger.Warn("Webhook request blocked by rate limit",
			zap.String("webhook_id", webhook.ID))
		return fmt.Errorf("webhook blocked by rate limit")
	}

	// Create delivery record
	delivery := &WebhookDelivery{
		ID:          generateDeliveryID(),
		WebhookID:   webhook.ID,
		TenantID:    webhook.TenantID,
		EventType:   event.Type,
		EventID:     event.ID,
		Payload:     event.Data,
		Headers:     webhook.Headers,
		Status:      DeliveryStatusPending,
		Attempts:    0,
		MaxAttempts: webhook.RetryPolicy.MaxRetries,
		CreatedAt:   time.Now(),
		Metadata:    event.Metadata,
	}

	// Save delivery record
	if err := wm.repository.SaveDelivery(ctx, delivery); err != nil {
		wm.logger.Error("Failed to save delivery record", zap.Error(err))
		return err
	}

	// Send webhook
	success, responseCode, responseBody, err := wm.sendWebhook(ctx, webhook, event, webhook.Headers)
	latency := time.Since(delivery.CreatedAt)

	// Update delivery status
	if success {
		delivery.Status = DeliveryStatusDelivered
		delivery.ResponseCode = &responseCode
		delivery.ResponseBody = responseBody
		delivery.Latency = &latency
		delivery.DeliveredAt = &time.Time{}
		*delivery.DeliveredAt = time.Now()

		// Record success in circuit breaker
		wm.circuitBreaker.RecordSuccess(ctx, webhook.ID)
	} else {
		delivery.Status = DeliveryStatusFailed
		delivery.ResponseCode = &responseCode
		delivery.ResponseBody = responseBody
		delivery.Latency = &latency
		delivery.Error = err.Error()

		// Record failure in circuit breaker
		wm.circuitBreaker.RecordFailure(ctx, webhook.ID)

		// Schedule retry if attempts remaining
		if delivery.Attempts < delivery.MaxAttempts {
			delivery.Status = DeliveryStatusRetrying
			if err := wm.retryHandler.ScheduleRetry(ctx, delivery); err != nil {
				wm.logger.Error("Failed to schedule retry", zap.Error(err))
			}
		}
	}

	delivery.Attempts++

	// Update delivery record
	if err := wm.repository.UpdateDelivery(ctx, delivery); err != nil {
		wm.logger.Error("Failed to update delivery record", zap.Error(err))
	}

	// Update webhook statistics
	wm.updateWebhookStatistics(ctx, webhook.ID, success, latency)

	wm.logger.Debug("Webhook delivery completed",
		zap.String("webhook_id", webhook.ID),
		zap.String("delivery_id", delivery.ID),
		zap.Bool("success", success),
		zap.Int("response_code", responseCode),
		zap.Duration("latency", latency))

	return nil
}

// sendWebhook sends the actual HTTP request to the webhook URL
func (wm *DefaultWebhookManager) sendWebhook(ctx context.Context, webhook *Webhook, event *WebhookEventData, headers map[string]string) (bool, int, string, error) {
	// Create payload
	payload := &WebhookPayload{
		ID:       event.ID,
		Type:     event.Type,
		Version:  event.Version,
		Created:  event.Timestamp,
		Data:     event.Data,
		Metadata: event.Metadata,
	}

	// Add signature if secret is provided
	if webhook.Secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		signature, err := wm.signatureVerifier.GenerateSignature(
			[]byte(fmt.Sprintf("%s.%s", timestamp, event.ID)),
			webhook.Secret,
			timestamp,
		)
		if err != nil {
			return false, 0, "", fmt.Errorf("failed to generate signature: %w", err)
		}

		payload.Signature = WebhookSignature{
			Algorithm: "sha256",
			Signature: signature,
			Timestamp: timestamp,
		}
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "KYB-Platform-Webhook/1.0")
	req.Header.Set("X-Webhook-Event", string(event.Type))
	req.Header.Set("X-Webhook-ID", event.ID)
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", event.Timestamp.Unix()))

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := wm.httpClient.Do(req)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, resp.StatusCode, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check if successful (2xx status codes)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return success, resp.StatusCode, string(responseBody), nil
}

// updateWebhookStatistics updates webhook statistics
func (wm *DefaultWebhookManager) updateWebhookStatistics(ctx context.Context, webhookID string, success bool, latency time.Duration) {
	stats, err := wm.repository.GetWebhookStatistics(ctx, webhookID)
	if err != nil {
		wm.logger.Error("Failed to get webhook statistics", zap.Error(err))
		return
	}

	stats.TotalDeliveries++
	if success {
		stats.SuccessfulDeliveries++
	} else {
		stats.FailedDeliveries++
	}

	// Calculate success rate
	if stats.TotalDeliveries > 0 {
		stats.SuccessRate = float64(stats.SuccessfulDeliveries) / float64(stats.TotalDeliveries) * 100
	}

	// Update average latency (simple moving average)
	if stats.AverageLatency == 0 {
		stats.AverageLatency = float64(latency.Milliseconds())
	} else {
		stats.AverageLatency = (stats.AverageLatency + float64(latency.Milliseconds())) / 2
	}

	now := time.Now()
	stats.LastDeliveryAt = &now

	// Save updated statistics
	if err := wm.repository.UpdateWebhookStatistics(ctx, webhookID, stats); err != nil {
		wm.logger.Error("Failed to update webhook statistics", zap.Error(err))
	}
}

// Placeholder implementations for other methods
// These would be implemented following similar patterns

func (wm *DefaultWebhookManager) GetDeliveries(ctx context.Context, filter *DeliveryFilter) (*DeliveryListResponse, error) {
	// Implementation for getting deliveries
	return nil, fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) GetDelivery(ctx context.Context, tenantID, deliveryID string) (*WebhookDelivery, error) {
	// Implementation for getting delivery
	return nil, fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) RetryDelivery(ctx context.Context, tenantID, deliveryID string) error {
	// Implementation for retrying delivery
	return fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) CancelDelivery(ctx context.Context, tenantID, deliveryID string) error {
	// Implementation for cancelling delivery
	return fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) GetWebhookMetrics(ctx context.Context, tenantID string) (*WebhookMetrics, error) {
	// Implementation for getting webhook metrics
	return nil, fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) GetWebhookHealth(ctx context.Context, tenantID, webhookID string) (*WebhookHealth, error) {
	// Implementation for getting webhook health
	return nil, fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) CreateTemplate(ctx context.Context, request *WebhookTemplateRequest) (*WebhookTemplateResponse, error) {
	// Implementation for creating template
	return nil, fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) GetTemplate(ctx context.Context, templateID string) (*WebhookTemplate, error) {
	// Implementation for getting template
	return nil, fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) ListTemplates(ctx context.Context, filter *WebhookTemplateFilter) (*WebhookTemplateListResponse, error) {
	// Implementation for listing templates
	return nil, fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) UpdateTemplate(ctx context.Context, templateID string, request *WebhookTemplateRequest) (*WebhookTemplateResponse, error) {
	// Implementation for updating template
	return nil, fmt.Errorf("not implemented")
}

func (wm *DefaultWebhookManager) DeleteTemplate(ctx context.Context, templateID string) error {
	// Implementation for deleting template
	return fmt.Errorf("not implemented")
}

// Helper functions

func generateWebhookID() string {
	return fmt.Sprintf("wh_%d", time.Now().UnixNano())
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func generateDeliveryID() string {
	return fmt.Sprintf("del_%d", time.Now().UnixNano())
}

func getTenantIDFromContext(ctx context.Context) string {
	// This would extract tenant ID from context
	// Implementation depends on your authentication/authorization system
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}
	return "default"
}
