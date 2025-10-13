package webhooks

import (
	"context"

	"go.uber.org/zap"
)

// DefaultWebhookEventFilter implements WebhookEventFilter
type DefaultWebhookEventFilter struct {
	logger *zap.Logger
}

// NewDefaultWebhookEventFilter creates a new event filter
func NewDefaultWebhookEventFilter(logger *zap.Logger) *DefaultWebhookEventFilter {
	return &DefaultWebhookEventFilter{
		logger: logger,
	}
}

// ShouldDeliver determines if an event should be delivered for a webhook
func (f *DefaultWebhookEventFilter) ShouldDeliver(ctx context.Context, webhook *Webhook, event *WebhookEventData) (bool, error) {
	// Simple implementation - check if the event type is in the webhook's events list
	for _, webhookEvent := range webhook.Events {
		if webhookEvent == event.Type {
			return true, nil
		}
	}
	return false, nil
}

// ApplyFilters applies filters to an event
func (f *DefaultWebhookEventFilter) ApplyFilters(ctx context.Context, webhook *Webhook, event *WebhookEventData) (*WebhookEventData, error) {
	// Simple implementation - return event as-is
	return event, nil
}
