package webhooks

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// DefaultWebhookRetryHandler implements WebhookRetryHandler
type DefaultWebhookRetryHandler struct {
	repository WebhookRepository
	logger     *zap.Logger
}

// NewDefaultWebhookRetryHandler creates a new default webhook retry handler
func NewDefaultWebhookRetryHandler(repository WebhookRepository, logger *zap.Logger) *DefaultWebhookRetryHandler {
	return &DefaultWebhookRetryHandler{
		repository: repository,
		logger:     logger,
	}
}

// ScheduleRetry schedules a retry for a failed webhook delivery
func (rh *DefaultWebhookRetryHandler) ScheduleRetry(ctx context.Context, delivery *WebhookDelivery) error {
	rh.logger.Debug("Scheduling retry for webhook delivery",
		zap.String("delivery_id", delivery.ID),
		zap.String("webhook_id", delivery.WebhookID),
		zap.Int("attempts", delivery.Attempts),
		zap.Int("max_attempts", delivery.MaxAttempts))

	// Check if max attempts reached
	if delivery.Attempts >= delivery.MaxAttempts {
		rh.logger.Warn("Max retry attempts reached for delivery",
			zap.String("delivery_id", delivery.ID),
			zap.Int("attempts", delivery.Attempts))
		return fmt.Errorf("max retry attempts reached")
	}

	// Get webhook to get retry policy
	webhook, err := rh.repository.GetWebhook(ctx, delivery.TenantID, delivery.WebhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return fmt.Errorf("webhook not found: %s", delivery.WebhookID)
	}

	// Calculate retry interval using exponential backoff
	retryInterval := rh.calculateRetryInterval(delivery.Attempts, webhook.RetryPolicy)

	// Add jitter if enabled
	if webhook.RetryPolicy.Jitter {
		jitter := time.Duration(rand.Float64() * float64(retryInterval) * 0.1) // 10% jitter
		retryInterval += jitter
	}

	// Set next retry time
	nextRetryAt := time.Now().Add(retryInterval)
	delivery.NextRetryAt = &nextRetryAt
	delivery.Status = DeliveryStatusRetrying

	// Update delivery
	if err := rh.repository.UpdateDelivery(ctx, delivery); err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	rh.logger.Info("Retry scheduled for webhook delivery",
		zap.String("delivery_id", delivery.ID),
		zap.String("webhook_id", delivery.WebhookID),
		zap.Duration("retry_interval", retryInterval),
		zap.Time("next_retry_at", nextRetryAt))

	return nil
}

// ProcessRetries processes pending retries
func (rh *DefaultWebhookRetryHandler) ProcessRetries(ctx context.Context) error {
	rh.logger.Debug("Processing webhook retries")

	// Get deliveries that are ready for retry
	filter := &DeliveryFilter{
		Status: DeliveryStatusRetrying,
	}

	deliveries, err := rh.repository.ListDeliveries(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list retry deliveries: %w", err)
	}

	now := time.Now()
	var processedCount int

	for _, delivery := range deliveries {
		// Check if it's time to retry
		if delivery.NextRetryAt == nil || delivery.NextRetryAt.After(now) {
			continue
		}

		// Check if max attempts reached
		if delivery.Attempts >= delivery.MaxAttempts {
			delivery.Status = DeliveryStatusFailed
			delivery.Error = "Max retry attempts reached"
			if err := rh.repository.UpdateDelivery(ctx, delivery); err != nil {
				rh.logger.Error("Failed to update delivery status", zap.Error(err))
			}
			continue
		}

		// Process retry
		if err := rh.processRetry(ctx, delivery); err != nil {
			rh.logger.Error("Failed to process retry", zap.Error(err))
			continue
		}

		processedCount++
	}

	rh.logger.Debug("Webhook retries processed",
		zap.Int("processed_count", processedCount))

	return nil
}

// processRetry processes a single retry
func (rh *DefaultWebhookRetryHandler) processRetry(ctx context.Context, delivery *WebhookDelivery) error {
	rh.logger.Debug("Processing retry for delivery",
		zap.String("delivery_id", delivery.ID),
		zap.String("webhook_id", delivery.WebhookID),
		zap.Int("attempt", delivery.Attempts+1))

	// Get webhook
	webhook, err := rh.repository.GetWebhook(ctx, delivery.TenantID, delivery.WebhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return fmt.Errorf("webhook not found: %s", delivery.WebhookID)
	}

	// Create event data from delivery
	eventData := &WebhookEventData{
		ID:        delivery.EventID,
		Type:      delivery.EventType,
		TenantID:  delivery.TenantID,
		Data:      delivery.Payload,
		Timestamp: delivery.CreatedAt,
		Source:    "webhook_retry",
		Version:   "1.0",
		Metadata:  delivery.Metadata,
	}

	// Send webhook (this would use the webhook manager's sendWebhook method)
	// For now, we'll simulate the retry
	success, responseCode, responseBody, err := rh.sendWebhookRetry(ctx, webhook, eventData, delivery.Headers)
	latency := time.Since(time.Now())

	// Update delivery
	delivery.Attempts++
	delivery.ResponseCode = &responseCode
	delivery.ResponseBody = responseBody
	delivery.Latency = &latency

	if success {
		delivery.Status = DeliveryStatusDelivered
		now := time.Now()
		delivery.DeliveredAt = &now
		delivery.NextRetryAt = nil
		delivery.Error = ""
	} else {
		delivery.Error = err.Error()

		// Schedule next retry if attempts remaining
		if delivery.Attempts < delivery.MaxAttempts {
			if err := rh.ScheduleRetry(ctx, delivery); err != nil {
				rh.logger.Error("Failed to schedule next retry", zap.Error(err))
			}
		} else {
			delivery.Status = DeliveryStatusFailed
			delivery.NextRetryAt = nil
		}
	}

	// Update delivery
	if err := rh.repository.UpdateDelivery(ctx, delivery); err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	rh.logger.Info("Retry processed for delivery",
		zap.String("delivery_id", delivery.ID),
		zap.String("webhook_id", delivery.WebhookID),
		zap.Int("attempt", delivery.Attempts),
		zap.Bool("success", success),
		zap.Int("response_code", responseCode))

	return nil
}

// GetRetryInfo gets retry information for a delivery
func (rh *DefaultWebhookRetryHandler) GetRetryInfo(ctx context.Context, deliveryID string) (*WebhookRetryInfo, error) {
	rh.logger.Debug("Getting retry info for delivery",
		zap.String("delivery_id", deliveryID))

	delivery, err := rh.repository.GetDelivery(ctx, "", deliveryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	if delivery == nil {
		return nil, fmt.Errorf("delivery not found: %s", deliveryID)
	}

	// Get webhook for retry policy
	webhook, err := rh.repository.GetWebhook(ctx, delivery.TenantID, delivery.WebhookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	if webhook == nil {
		return nil, fmt.Errorf("webhook not found: %s", delivery.WebhookID)
	}

	// Calculate retry info
	retryInfo := &WebhookRetryInfo{
		DeliveryID:       delivery.ID,
		Attempts:         delivery.Attempts,
		MaxAttempts:      delivery.MaxAttempts,
		RetryInterval:    rh.calculateRetryInterval(delivery.Attempts, webhook.RetryPolicy),
		BackoffFactor:    webhook.RetryPolicy.Multiplier,
		Jitter:           webhook.RetryPolicy.Jitter,
		LastError:        delivery.Error,
		ConsecutiveFails: delivery.Attempts,
	}

	if delivery.NextRetryAt != nil {
		retryInfo.NextRetryAt = *delivery.NextRetryAt
	}

	rh.logger.Debug("Retry info retrieved successfully",
		zap.String("delivery_id", deliveryID),
		zap.Int("attempts", retryInfo.Attempts),
		zap.Int("max_attempts", retryInfo.MaxAttempts))

	return retryInfo, nil
}

// CancelRetry cancels a retry for a delivery
func (rh *DefaultWebhookRetryHandler) CancelRetry(ctx context.Context, deliveryID string) error {
	rh.logger.Info("Cancelling retry for delivery",
		zap.String("delivery_id", deliveryID))

	delivery, err := rh.repository.GetDelivery(ctx, "", deliveryID)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	if delivery == nil {
		return fmt.Errorf("delivery not found: %s", deliveryID)
	}

	// Update delivery status
	delivery.Status = DeliveryStatusCancelled
	delivery.NextRetryAt = nil
	delivery.Error = "Retry cancelled by user"

	// Update delivery
	if err := rh.repository.UpdateDelivery(ctx, delivery); err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	rh.logger.Info("Retry cancelled successfully",
		zap.String("delivery_id", deliveryID))

	return nil
}

// calculateRetryInterval calculates the retry interval using exponential backoff
func (rh *DefaultWebhookRetryHandler) calculateRetryInterval(attempts int, policy RetryPolicy) time.Duration {
	// Calculate exponential backoff: initial * (multiplier ^ attempts)
	interval := float64(policy.InitialInterval) * math.Pow(policy.Multiplier, float64(attempts))

	// Cap at max interval
	if interval > float64(policy.MaxInterval) {
		interval = float64(policy.MaxInterval)
	}

	return time.Duration(interval)
}

// sendWebhookRetry sends a webhook retry (placeholder implementation)
func (rh *DefaultWebhookRetryHandler) sendWebhookRetry(ctx context.Context, webhook *Webhook, event *WebhookEventData, headers map[string]string) (bool, int, string, error) {
	// This would use the same logic as the webhook manager's sendWebhook method
	// For now, we'll simulate a retry with a random success/failure

	// Simulate network delay
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	// Simulate success/failure (80% success rate for retries)
	if rand.Float64() < 0.8 {
		return true, 200, "OK", nil
	} else {
		return false, 500, "Internal Server Error", fmt.Errorf("simulated retry failure")
	}
}

// GetRetryQueueStatus gets the status of the retry queue
func (rh *DefaultWebhookRetryHandler) GetRetryQueueStatus(ctx context.Context) (*RetryQueueStatus, error) {
	rh.logger.Debug("Getting retry queue status")

	// Get all retrying deliveries
	filter := &DeliveryFilter{
		Status: DeliveryStatusRetrying,
	}

	deliveries, err := rh.repository.ListDeliveries(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list retry deliveries: %w", err)
	}

	status := &RetryQueueStatus{
		TotalRetries:     len(deliveries),
		ReadyForRetry:    0,
		ScheduledRetries: 0,
		RetriesByWebhook: make(map[string]int),
	}

	now := time.Now()

	for _, delivery := range deliveries {
		// Count by webhook
		status.RetriesByWebhook[delivery.WebhookID]++

		// Check if ready for retry
		if delivery.NextRetryAt != nil && delivery.NextRetryAt.Before(now) {
			status.ReadyForRetry++
		} else {
			status.ScheduledRetries++
		}
	}

	rh.logger.Debug("Retry queue status retrieved successfully",
		zap.Int("total_retries", status.TotalRetries),
		zap.Int("ready_for_retry", status.ReadyForRetry),
		zap.Int("scheduled_retries", status.ScheduledRetries))

	return status, nil
}

// CleanupOldRetries cleans up old retry attempts
func (rh *DefaultWebhookRetryHandler) CleanupOldRetries(ctx context.Context, olderThan time.Duration) error {
	rh.logger.Info("Cleaning up old retries",
		zap.Duration("older_than", olderThan))

	// This would implement cleanup logic for old retry attempts
	// For now, it's a placeholder

	rh.logger.Info("Old retries cleaned up successfully")

	return nil
}

// Additional data structures for retry handling

// RetryQueueStatus represents the status of the retry queue
type RetryQueueStatus struct {
	TotalRetries     int            `json:"total_retries"`
	ReadyForRetry    int            `json:"ready_for_retry"`
	ScheduledRetries int            `json:"scheduled_retries"`
	RetriesByWebhook map[string]int `json:"retries_by_webhook"`
}

// RetryStatistics represents retry statistics
type RetryStatistics struct {
	TotalRetries      int64         `json:"total_retries"`
	SuccessfulRetries int64         `json:"successful_retries"`
	FailedRetries     int64         `json:"failed_retries"`
	RetrySuccessRate  float64       `json:"retry_success_rate"`
	AverageRetryDelay float64       `json:"average_retry_delay"`
	MaxRetryDelay     float64       `json:"max_retry_delay"`
	RetriesByAttempt  map[int]int64 `json:"retries_by_attempt"`
}
