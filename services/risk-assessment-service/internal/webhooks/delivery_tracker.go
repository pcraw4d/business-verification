package webhooks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// DefaultWebhookDeliveryTracker implements WebhookDeliveryTracker
type DefaultWebhookDeliveryTracker struct {
	repository WebhookRepository
	logger     *zap.Logger
}

// NewDefaultWebhookDeliveryTracker creates a new default webhook delivery tracker
func NewDefaultWebhookDeliveryTracker(repository WebhookRepository, logger *zap.Logger) *DefaultWebhookDeliveryTracker {
	return &DefaultWebhookDeliveryTracker{
		repository: repository,
		logger:     logger,
	}
}

// TrackDelivery tracks a webhook delivery
func (dt *DefaultWebhookDeliveryTracker) TrackDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	dt.logger.Debug("Tracking webhook delivery",
		zap.String("delivery_id", delivery.ID),
		zap.String("webhook_id", delivery.WebhookID),
		zap.String("event_type", string(delivery.EventType)))

	if err := dt.repository.SaveDelivery(ctx, delivery); err != nil {
		return fmt.Errorf("failed to save delivery: %w", err)
	}

	dt.logger.Debug("Webhook delivery tracked successfully",
		zap.String("delivery_id", delivery.ID))

	return nil
}

// UpdateDeliveryStatus updates the status of a webhook delivery
func (dt *DefaultWebhookDeliveryTracker) UpdateDeliveryStatus(ctx context.Context, deliveryID string, status DeliveryStatus, responseCode int, responseBody string, latency time.Duration, error string) error {
	dt.logger.Debug("Updating delivery status",
		zap.String("delivery_id", deliveryID),
		zap.String("status", string(status)))

	// Get delivery
	delivery, err := dt.repository.GetDelivery(ctx, "", deliveryID) // TenantID not needed for delivery lookup
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	if delivery == nil {
		return fmt.Errorf("delivery not found: %s", deliveryID)
	}

	// Update delivery fields
	delivery.Status = status
	delivery.ResponseCode = &responseCode
	delivery.ResponseBody = responseBody
	delivery.Latency = &latency
	delivery.Error = error

	if status == DeliveryStatusDelivered {
		now := time.Now()
		delivery.DeliveredAt = &now
	}

	// Save updated delivery
	if err := dt.repository.UpdateDelivery(ctx, delivery); err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	dt.logger.Debug("Delivery status updated successfully",
		zap.String("delivery_id", deliveryID),
		zap.String("status", string(status)))

	return nil
}

// GetDeliveryHistory retrieves delivery history for a webhook
func (dt *DefaultWebhookDeliveryTracker) GetDeliveryHistory(ctx context.Context, webhookID string, limit int) ([]*WebhookDelivery, error) {
	dt.logger.Debug("Getting delivery history",
		zap.String("webhook_id", webhookID),
		zap.Int("limit", limit))

	filter := &DeliveryFilter{
		WebhookID: webhookID,
		Limit:     limit,
	}

	deliveries, err := dt.repository.ListDeliveries(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}

	dt.logger.Debug("Delivery history retrieved successfully",
		zap.String("webhook_id", webhookID),
		zap.Int("count", len(deliveries)))

	return deliveries, nil
}

// GetFailedDeliveries retrieves failed deliveries for a webhook
func (dt *DefaultWebhookDeliveryTracker) GetFailedDeliveries(ctx context.Context, webhookID string) ([]*WebhookDelivery, error) {
	dt.logger.Debug("Getting failed deliveries",
		zap.String("webhook_id", webhookID))

	filter := &DeliveryFilter{
		WebhookID: webhookID,
		Status:    DeliveryStatusFailed,
	}

	deliveries, err := dt.repository.ListDeliveries(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list failed deliveries: %w", err)
	}

	dt.logger.Debug("Failed deliveries retrieved successfully",
		zap.String("webhook_id", webhookID),
		zap.Int("count", len(deliveries)))

	return deliveries, nil
}

// GetDeliveryStatistics calculates delivery statistics for a webhook
func (dt *DefaultWebhookDeliveryTracker) GetDeliveryStatistics(ctx context.Context, webhookID string) (*WebhookStatistics, error) {
	dt.logger.Debug("Getting delivery statistics",
		zap.String("webhook_id", webhookID))

	// Get all deliveries for the webhook
	filter := &DeliveryFilter{
		WebhookID: webhookID,
	}

	deliveries, err := dt.repository.ListDeliveries(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}

	// Calculate statistics
	stats := &WebhookStatistics{
		TotalDeliveries:      int64(len(deliveries)),
		SuccessfulDeliveries: 0,
		FailedDeliveries:     0,
		SuccessRate:          0,
		AverageLatency:       0,
	}

	var totalLatency time.Duration
	var lastDelivery *time.Time

	for _, delivery := range deliveries {
		if delivery.Status == DeliveryStatusDelivered {
			stats.SuccessfulDeliveries++
		} else {
			stats.FailedDeliveries++
		}

		if delivery.Latency != nil {
			totalLatency += *delivery.Latency
		}

		if lastDelivery == nil || delivery.CreatedAt.After(*lastDelivery) {
			lastDelivery = &delivery.CreatedAt
		}
	}

	// Calculate success rate
	if stats.TotalDeliveries > 0 {
		stats.SuccessRate = float64(stats.SuccessfulDeliveries) / float64(stats.TotalDeliveries) * 100
	}

	// Calculate average latency
	if stats.TotalDeliveries > 0 {
		stats.AverageLatency = float64(totalLatency.Milliseconds()) / float64(stats.TotalDeliveries)
	}

	stats.LastDeliveryAt = lastDelivery

	dt.logger.Debug("Delivery statistics calculated successfully",
		zap.String("webhook_id", webhookID),
		zap.Int64("total_deliveries", stats.TotalDeliveries),
		zap.Float64("success_rate", stats.SuccessRate))

	return stats, nil
}

// GetDeliveryMetrics retrieves delivery metrics for a webhook
func (dt *DefaultWebhookDeliveryTracker) GetDeliveryMetrics(ctx context.Context, webhookID string, startDate, endDate time.Time) (*DeliveryMetrics, error) {
	dt.logger.Debug("Getting delivery metrics",
		zap.String("webhook_id", webhookID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate))

	filter := &DeliveryFilter{
		WebhookID: webhookID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	deliveries, err := dt.repository.ListDeliveries(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}

	// Calculate metrics
	metrics := &DeliveryMetrics{
		TotalDeliveries:      int64(len(deliveries)),
		SuccessfulDeliveries: 0,
		FailedDeliveries:     0,
		RetryingDeliveries:   0,
		SuccessRate:          0,
		AverageLatency:       0,
		DeliveriesByHour:     make(map[string]int64),
		DeliveriesByStatus:   make(map[DeliveryStatus]int64),
		DeliveriesByEvent:    make(map[WebhookEvent]int64),
	}

	var totalLatency time.Duration

	for _, delivery := range deliveries {
		// Count by status
		metrics.DeliveriesByStatus[delivery.Status]++

		// Count by event type
		metrics.DeliveriesByEvent[delivery.EventType]++

		// Count by hour
		hour := delivery.CreatedAt.Format("2006-01-02 15:00")
		metrics.DeliveriesByHour[hour]++

		// Count successful/failed
		if delivery.Status == DeliveryStatusDelivered {
			metrics.SuccessfulDeliveries++
		} else if delivery.Status == DeliveryStatusFailed {
			metrics.FailedDeliveries++
		} else if delivery.Status == DeliveryStatusRetrying {
			metrics.RetryingDeliveries++
		}

		if delivery.Latency != nil {
			totalLatency += *delivery.Latency
		}
	}

	// Calculate success rate
	if metrics.TotalDeliveries > 0 {
		metrics.SuccessRate = float64(metrics.SuccessfulDeliveries) / float64(metrics.TotalDeliveries) * 100
	}

	// Calculate average latency
	if metrics.TotalDeliveries > 0 {
		metrics.AverageLatency = float64(totalLatency.Milliseconds()) / float64(metrics.TotalDeliveries)
	}

	dt.logger.Debug("Delivery metrics calculated successfully",
		zap.String("webhook_id", webhookID),
		zap.Int64("total_deliveries", metrics.TotalDeliveries),
		zap.Float64("success_rate", metrics.SuccessRate))

	return metrics, nil
}

// GetDeliveryTrends retrieves delivery trends for a webhook
func (dt *DefaultWebhookDeliveryTracker) GetDeliveryTrends(ctx context.Context, webhookID string, days int) ([]*DeliveryTrend, error) {
	dt.logger.Debug("Getting delivery trends",
		zap.String("webhook_id", webhookID),
		zap.Int("days", days))

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	filter := &DeliveryFilter{
		WebhookID: webhookID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	deliveries, err := dt.repository.ListDeliveries(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}

	// Group deliveries by date
	deliveriesByDate := make(map[string][]*WebhookDelivery)
	for _, delivery := range deliveries {
		date := delivery.CreatedAt.Format("2006-01-02")
		deliveriesByDate[date] = append(deliveriesByDate[date], delivery)
	}

	// Calculate trends
	var trends []*DeliveryTrend
	for i := 0; i < days; i++ {
		date := endDate.AddDate(0, 0, -i).Format("2006-01-02")
		dayDeliveries := deliveriesByDate[date]

		trend := &DeliveryTrend{
			Date:                 date,
			TotalDeliveries:      int64(len(dayDeliveries)),
			SuccessfulDeliveries: 0,
			FailedDeliveries:     0,
			SuccessRate:          0,
			AverageLatency:       0,
		}

		var totalLatency time.Duration
		for _, delivery := range dayDeliveries {
			if delivery.Status == DeliveryStatusDelivered {
				trend.SuccessfulDeliveries++
			} else {
				trend.FailedDeliveries++
			}
			if delivery.Latency != nil {
				totalLatency += *delivery.Latency
			}
		}

		// Calculate success rate
		if trend.TotalDeliveries > 0 {
			trend.SuccessRate = float64(trend.SuccessfulDeliveries) / float64(trend.TotalDeliveries) * 100
		}

		// Calculate average latency
		if trend.TotalDeliveries > 0 {
			trend.AverageLatency = float64(totalLatency.Milliseconds()) / float64(trend.TotalDeliveries)
		}

		trends = append(trends, trend)
	}

	dt.logger.Debug("Delivery trends calculated successfully",
		zap.String("webhook_id", webhookID),
		zap.Int("days", days),
		zap.Int("trend_points", len(trends)))

	return trends, nil
}

// GetDeliveryHealth calculates delivery health for a webhook
func (dt *DefaultWebhookDeliveryTracker) GetDeliveryHealth(ctx context.Context, webhookID string) (*WebhookHealth, error) {
	dt.logger.Debug("Getting delivery health",
		zap.String("webhook_id", webhookID))

	// Get recent deliveries (last 24 hours)
	endDate := time.Now()
	startDate := endDate.Add(-24 * time.Hour)

	filter := &DeliveryFilter{
		WebhookID: webhookID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	deliveries, err := dt.repository.ListDeliveries(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}

	health := &WebhookHealth{
		WebhookID:        webhookID,
		Status:           "healthy",
		ConsecutiveFails: 0,
		AverageLatency:   0,
		SuccessRate:      0,
	}

	if len(deliveries) == 0 {
		health.Status = "unknown"
		return health, nil
	}

	// Calculate health metrics
	var totalLatency time.Duration
	var successfulDeliveries int64
	var consecutiveFails int
	var lastSuccessful *time.Time
	var lastFailed *time.Time

	// Process deliveries in reverse chronological order
	for i := len(deliveries) - 1; i >= 0; i-- {
		delivery := deliveries[i]

		if delivery.Latency != nil {
			totalLatency += *delivery.Latency
		}

		if delivery.Status == DeliveryStatusDelivered {
			successfulDeliveries++
			lastSuccessful = &delivery.CreatedAt
			consecutiveFails = 0 // Reset consecutive fails
		} else {
			lastFailed = &delivery.CreatedAt
			consecutiveFails++
		}
	}

	// Calculate success rate
	health.SuccessRate = float64(successfulDeliveries) / float64(len(deliveries)) * 100

	// Calculate average latency
	health.AverageLatency = float64(totalLatency.Milliseconds()) / float64(len(deliveries))

	// Set consecutive fails
	health.ConsecutiveFails = consecutiveFails

	// Set last successful/failed times
	health.LastSuccessful = lastSuccessful
	health.LastFailed = lastFailed

	// Determine health status
	if health.SuccessRate >= 95 && health.ConsecutiveFails < 3 {
		health.Status = "healthy"
	} else if health.SuccessRate >= 80 && health.ConsecutiveFails < 5 {
		health.Status = "degraded"
	} else {
		health.Status = "unhealthy"
	}

	dt.logger.Debug("Delivery health calculated successfully",
		zap.String("webhook_id", webhookID),
		zap.String("status", health.Status),
		zap.Float64("success_rate", health.SuccessRate),
		zap.Int("consecutive_fails", health.ConsecutiveFails))

	return health, nil
}

// Additional data structures for delivery tracking

// DeliveryMetrics represents delivery metrics for a webhook
type DeliveryMetrics struct {
	TotalDeliveries      int64                    `json:"total_deliveries"`
	SuccessfulDeliveries int64                    `json:"successful_deliveries"`
	FailedDeliveries     int64                    `json:"failed_deliveries"`
	RetryingDeliveries   int64                    `json:"retrying_deliveries"`
	SuccessRate          float64                  `json:"success_rate"`
	AverageLatency       float64                  `json:"average_latency"`
	DeliveriesByHour     map[string]int64         `json:"deliveries_by_hour"`
	DeliveriesByStatus   map[DeliveryStatus]int64 `json:"deliveries_by_status"`
	DeliveriesByEvent    map[WebhookEvent]int64   `json:"deliveries_by_event"`
}

// DeliveryTrend represents delivery trends for a webhook
type DeliveryTrend struct {
	Date                 string  `json:"date"`
	TotalDeliveries      int64   `json:"total_deliveries"`
	SuccessfulDeliveries int64   `json:"successful_deliveries"`
	FailedDeliveries     int64   `json:"failed_deliveries"`
	SuccessRate          float64 `json:"success_rate"`
	AverageLatency       float64 `json:"average_latency"`
}
