package webhooks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SQLWebhookRepository implements WebhookRepository using SQL database
type SQLWebhookRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLWebhookRepository creates a new SQL webhook repository
func NewSQLWebhookRepository(db *sql.DB, logger *zap.Logger) *SQLWebhookRepository {
	return &SQLWebhookRepository{
		db:     db,
		logger: logger,
	}
}

// SaveWebhook saves a webhook to the database
func (r *SQLWebhookRepository) SaveWebhook(ctx context.Context, webhook *Webhook) error {
	r.logger.Info("Saving webhook",
		zap.String("webhook_id", webhook.ID),
		zap.String("tenant_id", webhook.TenantID),
		zap.String("name", webhook.Name))

	// Convert complex fields to JSON
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	retryPolicyJSON, err := json.Marshal(webhook.RetryPolicy)
	if err != nil {
		return fmt.Errorf("failed to marshal retry policy: %w", err)
	}

	rateLimitJSON, err := json.Marshal(webhook.RateLimit)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit: %w", err)
	}

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	filtersJSON, err := json.Marshal(webhook.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	statisticsJSON, err := json.Marshal(webhook.Statistics)
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	metadataJSON, err := json.Marshal(webhook.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check if webhook exists
	existingWebhook, err := r.GetWebhook(ctx, webhook.TenantID, webhook.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing webhook: %w", err)
	}

	if existingWebhook != nil {
		// Update existing webhook
		query := `
			UPDATE webhooks SET
				name = $1, description = $2, url = $3, events = $4, secret = $5, status = $6,
				retry_policy = $7, rate_limit = $8, headers = $9, filters = $10, statistics = $11,
				updated_at = $12, last_triggered_at = $13, metadata = $14
			WHERE id = $15 AND tenant_id = $16
		`
		_, err = r.db.ExecContext(ctx, query,
			webhook.Name, webhook.Description, webhook.URL, eventsJSON, webhook.Secret, webhook.Status,
			retryPolicyJSON, rateLimitJSON, headersJSON, filtersJSON, statisticsJSON,
			webhook.UpdatedAt, webhook.LastTriggeredAt, metadataJSON,
			webhook.ID, webhook.TenantID)
	} else {
		// Insert new webhook
		query := `
			INSERT INTO webhooks (
				id, tenant_id, name, description, url, events, secret, status,
				retry_policy, rate_limit, headers, filters, statistics,
				created_by, created_at, updated_at, last_triggered_at, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
			)
		`
		_, err = r.db.ExecContext(ctx, query,
			webhook.ID, webhook.TenantID, webhook.Name, webhook.Description, webhook.URL, eventsJSON, webhook.Secret, webhook.Status,
			retryPolicyJSON, rateLimitJSON, headersJSON, filtersJSON, statisticsJSON,
			webhook.CreatedBy, webhook.CreatedAt, webhook.UpdatedAt, webhook.LastTriggeredAt, metadataJSON)
	}

	if err != nil {
		return fmt.Errorf("failed to save webhook: %w", err)
	}

	r.logger.Info("Webhook saved successfully",
		zap.String("webhook_id", webhook.ID),
		zap.String("tenant_id", webhook.TenantID))

	return nil
}

// GetWebhook retrieves a webhook by ID
func (r *SQLWebhookRepository) GetWebhook(ctx context.Context, tenantID, webhookID string) (*Webhook, error) {
	r.logger.Debug("Getting webhook",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, name, description, url, events, secret, status,
			   retry_policy, rate_limit, headers, filters, statistics,
			   created_by, created_at, updated_at, last_triggered_at, metadata
		FROM webhooks
		WHERE id = $1 AND tenant_id = $2
	`

	var result struct {
		ID              string     `json:"id"`
		TenantID        string     `json:"tenant_id"`
		Name            string     `json:"name"`
		Description     string     `json:"description"`
		URL             string     `json:"url"`
		Events          string     `json:"events"`
		Secret          string     `json:"secret"`
		Status          string     `json:"status"`
		RetryPolicy     string     `json:"retry_policy"`
		RateLimit       string     `json:"rate_limit"`
		Headers         string     `json:"headers"`
		Filters         string     `json:"filters"`
		Statistics      string     `json:"statistics"`
		CreatedBy       string     `json:"created_by"`
		CreatedAt       time.Time  `json:"created_at"`
		UpdatedAt       time.Time  `json:"updated_at"`
		LastTriggeredAt *time.Time `json:"last_triggered_at"`
		Metadata        string     `json:"metadata"`
	}

	err := r.db.QueryRowContext(ctx, query, webhookID, tenantID).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.Description, &result.URL, &result.Events, &result.Secret, &result.Status,
		&result.RetryPolicy, &result.RateLimit, &result.Headers, &result.Filters, &result.Statistics,
		&result.CreatedBy, &result.CreatedAt, &result.UpdatedAt, &result.LastTriggeredAt, &result.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Webhook not found
		}
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	// Convert the result to Webhook
	webhook, err := r.convertToWebhook(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Webhook retrieved successfully",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	return webhook, nil
}

// ListWebhooks lists webhooks with filters
func (r *SQLWebhookRepository) ListWebhooks(ctx context.Context, filter *WebhookFilter) ([]*Webhook, error) {
	r.logger.Debug("Listing webhooks",
		zap.String("tenant_id", filter.TenantID))

	// Build query with filters
	query := `
		SELECT id, tenant_id, name, description, url, events, secret, status,
			   retry_policy, rate_limit, headers, filters, statistics,
			   created_by, created_at, updated_at, last_triggered_at, metadata
		FROM webhooks
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, filter.TenantID)
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.EventType != "" {
		query += fmt.Sprintf(" AND events::jsonb ? $%d", argIndex)
		args = append(args, string(filter.EventType))
		argIndex++
	}

	if filter.CreatedBy != "" {
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, filter.CreatedBy)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*Webhook
	for rows.Next() {
		var result struct {
			ID              string     `json:"id"`
			TenantID        string     `json:"tenant_id"`
			Name            string     `json:"name"`
			Description     string     `json:"description"`
			URL             string     `json:"url"`
			Events          string     `json:"events"`
			Secret          string     `json:"secret"`
			Status          string     `json:"status"`
			RetryPolicy     string     `json:"retry_policy"`
			RateLimit       string     `json:"rate_limit"`
			Headers         string     `json:"headers"`
			Filters         string     `json:"filters"`
			Statistics      string     `json:"statistics"`
			CreatedBy       string     `json:"created_by"`
			CreatedAt       time.Time  `json:"created_at"`
			UpdatedAt       time.Time  `json:"updated_at"`
			LastTriggeredAt *time.Time `json:"last_triggered_at"`
			Metadata        string     `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.Description, &result.URL, &result.Events, &result.Secret, &result.Status,
			&result.RetryPolicy, &result.RateLimit, &result.Headers, &result.Filters, &result.Statistics,
			&result.CreatedBy, &result.CreatedAt, &result.UpdatedAt, &result.LastTriggeredAt, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		webhook, err := r.convertToWebhook(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Webhooks listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(webhooks)))

	return webhooks, nil
}

// UpdateWebhook updates a webhook
func (r *SQLWebhookRepository) UpdateWebhook(ctx context.Context, webhook *Webhook) error {
	r.logger.Debug("Updating webhook",
		zap.String("webhook_id", webhook.ID),
		zap.String("tenant_id", webhook.TenantID))

	// Convert complex fields to JSON
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	retryPolicyJSON, err := json.Marshal(webhook.RetryPolicy)
	if err != nil {
		return fmt.Errorf("failed to marshal retry policy: %w", err)
	}

	rateLimitJSON, err := json.Marshal(webhook.RateLimit)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit: %w", err)
	}

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	filtersJSON, err := json.Marshal(webhook.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	statisticsJSON, err := json.Marshal(webhook.Statistics)
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	metadataJSON, err := json.Marshal(webhook.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE webhooks SET
			name = $1, description = $2, url = $3, events = $4, secret = $5, status = $6,
			retry_policy = $7, rate_limit = $8, headers = $9, filters = $10, statistics = $11,
			updated_at = $12, last_triggered_at = $13, metadata = $14
		WHERE id = $15 AND tenant_id = $16
	`

	_, err = r.db.ExecContext(ctx, query,
		webhook.Name, webhook.Description, webhook.URL, eventsJSON, webhook.Secret, webhook.Status,
		retryPolicyJSON, rateLimitJSON, headersJSON, filtersJSON, statisticsJSON,
		webhook.UpdatedAt, webhook.LastTriggeredAt, metadataJSON,
		webhook.ID, webhook.TenantID)

	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	r.logger.Debug("Webhook updated successfully",
		zap.String("webhook_id", webhook.ID),
		zap.String("tenant_id", webhook.TenantID))

	return nil
}

// DeleteWebhook deletes a webhook
func (r *SQLWebhookRepository) DeleteWebhook(ctx context.Context, tenantID, webhookID string) error {
	r.logger.Info("Deleting webhook",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	_, err := r.db.ExecContext(ctx, "DELETE FROM webhooks WHERE id = $1 AND tenant_id = $2", webhookID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	r.logger.Info("Webhook deleted successfully",
		zap.String("webhook_id", webhookID),
		zap.String("tenant_id", tenantID))

	return nil
}

// UpdateWebhookStatistics updates webhook statistics
func (r *SQLWebhookRepository) UpdateWebhookStatistics(ctx context.Context, webhookID string, stats *WebhookStatistics) error {
	r.logger.Debug("Updating webhook statistics",
		zap.String("webhook_id", webhookID))

	statisticsJSON, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	query := `UPDATE webhooks SET statistics = $1, updated_at = $2 WHERE id = $3`
	_, err = r.db.ExecContext(ctx, query, statisticsJSON, time.Now(), webhookID)
	if err != nil {
		return fmt.Errorf("failed to update webhook statistics: %w", err)
	}

	r.logger.Debug("Webhook statistics updated successfully",
		zap.String("webhook_id", webhookID))

	return nil
}

// GetWebhookStatistics retrieves webhook statistics
func (r *SQLWebhookRepository) GetWebhookStatistics(ctx context.Context, webhookID string) (*WebhookStatistics, error) {
	r.logger.Debug("Getting webhook statistics",
		zap.String("webhook_id", webhookID))

	query := `SELECT statistics FROM webhooks WHERE id = $1`
	var statisticsJSON string
	err := r.db.QueryRowContext(ctx, query, webhookID).Scan(&statisticsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return &WebhookStatistics{}, nil // Return empty statistics if webhook not found
		}
		return nil, fmt.Errorf("failed to get webhook statistics: %w", err)
	}

	var stats WebhookStatistics
	if err := json.Unmarshal([]byte(statisticsJSON), &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statistics: %w", err)
	}

	r.logger.Debug("Webhook statistics retrieved successfully",
		zap.String("webhook_id", webhookID))

	return &stats, nil
}

// SaveDelivery saves a webhook delivery to the database
func (r *SQLWebhookRepository) SaveDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	r.logger.Debug("Saving webhook delivery",
		zap.String("delivery_id", delivery.ID),
		zap.String("webhook_id", delivery.WebhookID))

	// Convert complex fields to JSON
	payloadJSON, err := json.Marshal(delivery.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	headersJSON, err := json.Marshal(delivery.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	responseHeadersJSON, err := json.Marshal(delivery.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("failed to marshal response headers: %w", err)
	}

	metadataJSON, err := json.Marshal(delivery.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO webhook_deliveries (
			id, webhook_id, tenant_id, event_type, event_id, payload, headers,
			status, attempts, max_attempts, response_code, response_body, response_headers,
			latency, error, next_retry_at, created_at, delivered_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		delivery.ID, delivery.WebhookID, delivery.TenantID, delivery.EventType, delivery.EventID, payloadJSON, headersJSON,
		delivery.Status, delivery.Attempts, delivery.MaxAttempts, delivery.ResponseCode, delivery.ResponseBody, responseHeadersJSON,
		delivery.Latency, delivery.Error, delivery.NextRetryAt, delivery.CreatedAt, delivery.DeliveredAt, metadataJSON)

	if err != nil {
		return fmt.Errorf("failed to save delivery: %w", err)
	}

	r.logger.Debug("Webhook delivery saved successfully",
		zap.String("delivery_id", delivery.ID),
		zap.String("webhook_id", delivery.WebhookID))

	return nil
}

// GetDelivery retrieves a webhook delivery by ID
func (r *SQLWebhookRepository) GetDelivery(ctx context.Context, tenantID, deliveryID string) (*WebhookDelivery, error) {
	r.logger.Debug("Getting webhook delivery",
		zap.String("delivery_id", deliveryID))

	query := `
		SELECT id, webhook_id, tenant_id, event_type, event_id, payload, headers,
			   status, attempts, max_attempts, response_code, response_body, response_headers,
			   latency, error, next_retry_at, created_at, delivered_at, metadata
		FROM webhook_deliveries
		WHERE id = $1
	`

	var result struct {
		ID              string        `json:"id"`
		WebhookID       string        `json:"webhook_id"`
		TenantID        string        `json:"tenant_id"`
		EventType       string        `json:"event_type"`
		EventID         string        `json:"event_id"`
		Payload         string        `json:"payload"`
		Headers         string        `json:"headers"`
		Status          string        `json:"status"`
		Attempts        int           `json:"attempts"`
		MaxAttempts     int           `json:"max_attempts"`
		ResponseCode    int           `json:"response_code"`
		ResponseBody    string        `json:"response_body"`
		ResponseHeaders string        `json:"response_headers"`
		Latency         time.Duration `json:"latency"`
		Error           string        `json:"error"`
		NextRetryAt     *time.Time    `json:"next_retry_at"`
		CreatedAt       time.Time     `json:"created_at"`
		DeliveredAt     *time.Time    `json:"delivered_at"`
		Metadata        string        `json:"metadata"`
	}

	err := r.db.QueryRowContext(ctx, query, deliveryID).Scan(
		&result.ID, &result.WebhookID, &result.TenantID, &result.EventType, &result.EventID, &result.Payload, &result.Headers,
		&result.Status, &result.Attempts, &result.MaxAttempts, &result.ResponseCode, &result.ResponseBody, &result.ResponseHeaders,
		&result.Latency, &result.Error, &result.NextRetryAt, &result.CreatedAt, &result.DeliveredAt, &result.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Delivery not found
		}
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	// Convert the result to WebhookDelivery
	delivery, err := r.convertToWebhookDelivery(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Webhook delivery retrieved successfully",
		zap.String("delivery_id", deliveryID))

	return delivery, nil
}

// ListDeliveries lists webhook deliveries with filters
func (r *SQLWebhookRepository) ListDeliveries(ctx context.Context, filter *DeliveryFilter) ([]*WebhookDelivery, error) {
	r.logger.Debug("Listing webhook deliveries")

	// Build query with filters
	query := `
		SELECT id, webhook_id, tenant_id, event_type, event_id, payload, headers,
			   status, attempts, max_attempts, response_code, response_body, response_headers,
			   latency, error, next_retry_at, created_at, delivered_at, metadata
		FROM webhook_deliveries
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.WebhookID != "" {
		query += fmt.Sprintf(" AND webhook_id = $%d", argIndex)
		args = append(args, filter.WebhookID)
		argIndex++
	}

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, filter.TenantID)
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.EventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, filter.EventType)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*WebhookDelivery
	for rows.Next() {
		var result struct {
			ID              string        `json:"id"`
			WebhookID       string        `json:"webhook_id"`
			TenantID        string        `json:"tenant_id"`
			EventType       string        `json:"event_type"`
			EventID         string        `json:"event_id"`
			Payload         string        `json:"payload"`
			Headers         string        `json:"headers"`
			Status          string        `json:"status"`
			Attempts        int           `json:"attempts"`
			MaxAttempts     int           `json:"max_attempts"`
			ResponseCode    int           `json:"response_code"`
			ResponseBody    string        `json:"response_body"`
			ResponseHeaders string        `json:"response_headers"`
			Latency         time.Duration `json:"latency"`
			Error           string        `json:"error"`
			NextRetryAt     *time.Time    `json:"next_retry_at"`
			CreatedAt       time.Time     `json:"created_at"`
			DeliveredAt     *time.Time    `json:"delivered_at"`
			Metadata        string        `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.WebhookID, &result.TenantID, &result.EventType, &result.EventID, &result.Payload, &result.Headers,
			&result.Status, &result.Attempts, &result.MaxAttempts, &result.ResponseCode, &result.ResponseBody, &result.ResponseHeaders,
			&result.Latency, &result.Error, &result.NextRetryAt, &result.CreatedAt, &result.DeliveredAt, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		delivery, err := r.convertToWebhookDelivery(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		deliveries = append(deliveries, delivery)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Webhook deliveries listed successfully",
		zap.Int("count", len(deliveries)))

	return deliveries, nil
}

// UpdateDelivery updates a webhook delivery
func (r *SQLWebhookRepository) UpdateDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	r.logger.Debug("Updating webhook delivery",
		zap.String("delivery_id", delivery.ID))

	// Convert complex fields to JSON
	_, err := json.Marshal(delivery.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = json.Marshal(delivery.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	responseHeadersJSON, err := json.Marshal(delivery.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("failed to marshal response headers: %w", err)
	}

	metadataJSON, err := json.Marshal(delivery.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE webhook_deliveries SET
			status = $1, attempts = $2, response_code = $3, response_body = $4, response_headers = $5,
			latency = $6, error = $7, next_retry_at = $8, delivered_at = $9, metadata = $10
		WHERE id = $11
	`

	_, err = r.db.ExecContext(ctx, query,
		delivery.Status, delivery.Attempts, delivery.ResponseCode, delivery.ResponseBody, responseHeadersJSON,
		delivery.Latency, delivery.Error, delivery.NextRetryAt, delivery.DeliveredAt, metadataJSON,
		delivery.ID)

	if err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	r.logger.Debug("Webhook delivery updated successfully",
		zap.String("delivery_id", delivery.ID))

	return nil
}

// convertToWebhook converts a database result to Webhook
func (r *SQLWebhookRepository) convertToWebhook(result interface{}) (*Webhook, error) {
	// Type assertion to get the fields
	var id, tenantID, name, description, url, events, secret, status, retryPolicy, rateLimit, headers, filters, statistics, createdBy, metadata string
	var createdAt, updatedAt time.Time
	var lastTriggeredAt *time.Time

	// Use reflection or type assertion based on the actual structure
	switch v := result.(type) {
	case *struct {
		ID              string     `json:"id"`
		TenantID        string     `json:"tenant_id"`
		Name            string     `json:"name"`
		Description     string     `json:"description"`
		URL             string     `json:"url"`
		Events          string     `json:"events"`
		Secret          string     `json:"secret"`
		Status          string     `json:"status"`
		RetryPolicy     string     `json:"retry_policy"`
		RateLimit       string     `json:"rate_limit"`
		Headers         string     `json:"headers"`
		Filters         string     `json:"filters"`
		Statistics      string     `json:"statistics"`
		CreatedBy       string     `json:"created_by"`
		CreatedAt       time.Time  `json:"created_at"`
		UpdatedAt       time.Time  `json:"updated_at"`
		LastTriggeredAt *time.Time `json:"last_triggered_at"`
		Metadata        string     `json:"metadata"`
	}:
		id = v.ID
		tenantID = v.TenantID
		name = v.Name
		description = v.Description
		url = v.URL
		events = v.Events
		secret = v.Secret
		status = v.Status
		retryPolicy = v.RetryPolicy
		rateLimit = v.RateLimit
		headers = v.Headers
		filters = v.Filters
		statistics = v.Statistics
		createdBy = v.CreatedBy
		createdAt = v.CreatedAt
		updatedAt = v.UpdatedAt
		lastTriggeredAt = v.LastTriggeredAt
		metadata = v.Metadata
	default:
		return nil, fmt.Errorf("unsupported result type: %T", result)
	}

	// Parse JSON fields
	var eventsList []WebhookEvent
	if events != "" {
		if err := json.Unmarshal([]byte(events), &eventsList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}
	}

	var retryPolicyObj RetryPolicy
	if retryPolicy != "" {
		if err := json.Unmarshal([]byte(retryPolicy), &retryPolicyObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal retry policy: %w", err)
		}
	}

	var rateLimitObj RateLimitConfig
	if rateLimit != "" {
		if err := json.Unmarshal([]byte(rateLimit), &rateLimitObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rate limit: %w", err)
		}
	}

	var headersMap map[string]string
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &headersMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	}

	var filtersObj WebhookFilters
	if filters != "" {
		if err := json.Unmarshal([]byte(filters), &filtersObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}
	}

	var statisticsObj WebhookStatistics
	if statistics != "" {
		if err := json.Unmarshal([]byte(statistics), &statisticsObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal statistics: %w", err)
		}
	}

	var metadataMap map[string]interface{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	webhook := &Webhook{
		ID:              id,
		TenantID:        tenantID,
		Name:            name,
		Description:     description,
		URL:             url,
		Events:          eventsList,
		Secret:          secret,
		Status:          WebhookStatus(status),
		RetryPolicy:     retryPolicyObj,
		RateLimit:       rateLimitObj,
		Headers:         headersMap,
		Filters:         filtersObj,
		Statistics:      statisticsObj,
		CreatedBy:       createdBy,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		LastTriggeredAt: lastTriggeredAt,
		Metadata:        metadataMap,
	}

	return webhook, nil
}

// convertToWebhookDelivery converts a database result to WebhookDelivery
func (r *SQLWebhookRepository) convertToWebhookDelivery(result interface{}) (*WebhookDelivery, error) {
	// Type assertion to get the fields
	var id, webhookID, tenantID, eventType, eventID, payload, headers, status, responseBody, responseHeaders, error, metadata string
	var attempts, maxAttempts, responseCode int
	var latency time.Duration
	var createdAt time.Time
	var nextRetryAt, deliveredAt *time.Time

	// Use reflection or type assertion based on the actual structure
	switch v := result.(type) {
	case *struct {
		ID              string        `json:"id"`
		WebhookID       string        `json:"webhook_id"`
		TenantID        string        `json:"tenant_id"`
		EventType       string        `json:"event_type"`
		EventID         string        `json:"event_id"`
		Payload         string        `json:"payload"`
		Headers         string        `json:"headers"`
		Status          string        `json:"status"`
		Attempts        int           `json:"attempts"`
		MaxAttempts     int           `json:"max_attempts"`
		ResponseCode    int           `json:"response_code"`
		ResponseBody    string        `json:"response_body"`
		ResponseHeaders string        `json:"response_headers"`
		Latency         time.Duration `json:"latency"`
		Error           string        `json:"error"`
		NextRetryAt     *time.Time    `json:"next_retry_at"`
		CreatedAt       time.Time     `json:"created_at"`
		DeliveredAt     *time.Time    `json:"delivered_at"`
		Metadata        string        `json:"metadata"`
	}:
		id = v.ID
		webhookID = v.WebhookID
		tenantID = v.TenantID
		eventType = v.EventType
		eventID = v.EventID
		payload = v.Payload
		headers = v.Headers
		status = v.Status
		attempts = v.Attempts
		maxAttempts = v.MaxAttempts
		responseCode = v.ResponseCode
		responseBody = v.ResponseBody
		responseHeaders = v.ResponseHeaders
		latency = v.Latency
		error = v.Error
		nextRetryAt = v.NextRetryAt
		createdAt = v.CreatedAt
		deliveredAt = v.DeliveredAt
		metadata = v.Metadata
	default:
		return nil, fmt.Errorf("unsupported result type: %T", result)
	}

	// Parse JSON fields
	var payloadMap map[string]interface{}
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &payloadMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	var headersMap map[string]string
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &headersMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	}

	var responseHeadersMap map[string]string
	if responseHeaders != "" {
		if err := json.Unmarshal([]byte(responseHeaders), &responseHeadersMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response headers: %w", err)
		}
	}

	var metadataMap map[string]interface{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	delivery := &WebhookDelivery{
		ID:              id,
		WebhookID:       webhookID,
		TenantID:        tenantID,
		EventType:       WebhookEvent(eventType),
		EventID:         eventID,
		Payload:         payloadMap,
		Headers:         headersMap,
		Status:          DeliveryStatus(status),
		Attempts:        attempts,
		MaxAttempts:     maxAttempts,
		ResponseCode:    &responseCode,
		ResponseBody:    responseBody,
		ResponseHeaders: responseHeadersMap,
		Latency:         &latency,
		Error:           error,
		NextRetryAt:     nextRetryAt,
		CreatedAt:       createdAt,
		DeliveredAt:     deliveredAt,
		Metadata:        metadataMap,
	}

	return delivery, nil
}

// Placeholder implementations for other methods
// These would follow similar patterns to the webhook methods

func (r *SQLWebhookRepository) SaveTemplate(ctx context.Context, template *WebhookTemplate) error {
	// Implementation for saving webhook template
	return fmt.Errorf("not implemented")
}

func (r *SQLWebhookRepository) GetTemplate(ctx context.Context, templateID string) (*WebhookTemplate, error) {
	// Implementation for getting webhook template
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLWebhookRepository) ListTemplates(ctx context.Context, filter *WebhookTemplateFilter) ([]*WebhookTemplate, error) {
	// Implementation for listing webhook templates
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLWebhookRepository) UpdateTemplate(ctx context.Context, template *WebhookTemplate) error {
	// Implementation for updating webhook template
	return fmt.Errorf("not implemented")
}

func (r *SQLWebhookRepository) DeleteTemplate(ctx context.Context, templateID string) error {
	// Implementation for deleting webhook template
	return fmt.Errorf("not implemented")
}

func (r *SQLWebhookRepository) GetWebhookMetrics(ctx context.Context, tenantID string) (*WebhookMetrics, error) {
	// Implementation for getting webhook metrics
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLWebhookRepository) GetWebhookHealth(ctx context.Context, tenantID, webhookID string) (*WebhookHealth, error) {
	// Implementation for getting webhook health
	return nil, fmt.Errorf("not implemented")
}
