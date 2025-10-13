package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/webhooks"
)

// WebhookHandlers handles webhook-related HTTP requests
type WebhookHandlers struct {
	webhookManager    webhooks.WebhookManager
	deliveryTracker   webhooks.WebhookDeliveryTracker
	retryHandler      webhooks.WebhookRetryHandler
	signatureVerifier webhooks.WebhookSignatureVerifier
	logger            *zap.Logger
}

// NewWebhookHandlers creates a new WebhookHandlers instance
func NewWebhookHandlers(
	webhookManager webhooks.WebhookManager,
	deliveryTracker webhooks.WebhookDeliveryTracker,
	retryHandler webhooks.WebhookRetryHandler,
	signatureVerifier webhooks.WebhookSignatureVerifier,
	logger *zap.Logger,
) *WebhookHandlers {
	return &WebhookHandlers{
		webhookManager:    webhookManager,
		deliveryTracker:   deliveryTracker,
		retryHandler:      retryHandler,
		signatureVerifier: signatureVerifier,
		logger:            logger,
	}
}

// CreateWebhookRequest represents the request to create a webhook
type CreateWebhookRequest struct {
	Name        string                             `json:"name" validate:"required,min=1,max=255"`
	Description string                             `json:"description,omitempty"`
	URL         string                             `json:"url" validate:"required,url"`
	Events      []string                           `json:"events" validate:"required,min=1"`
	Secret      string                             `json:"secret,omitempty"`
	Status      string                             `json:"status,omitempty"`
	RetryPolicy *webhooks.RetryPolicy              `json:"retry_policy,omitempty"`
	RateLimit   *webhooks.RateLimitConfig          `json:"rate_limit,omitempty"`
	Headers     map[string]string                  `json:"headers,omitempty"`
	Filters     *webhooks.WebhookEventFilterConfig `json:"filters,omitempty"`
	Metadata    map[string]interface{}             `json:"metadata,omitempty"`
}

// UpdateWebhookRequest represents the request to update a webhook
type UpdateWebhookRequest struct {
	Name        *string                            `json:"name,omitempty"`
	Description *string                            `json:"description,omitempty"`
	URL         *string                            `json:"url,omitempty"`
	Events      []string                           `json:"events,omitempty"`
	Secret      *string                            `json:"secret,omitempty"`
	Status      *string                            `json:"status,omitempty"`
	RetryPolicy *webhooks.RetryPolicy              `json:"retry_policy,omitempty"`
	RateLimit   *webhooks.RateLimitConfig          `json:"rate_limit,omitempty"`
	Headers     map[string]string                  `json:"headers,omitempty"`
	Filters     *webhooks.WebhookEventFilterConfig `json:"filters,omitempty"`
	Metadata    map[string]interface{}             `json:"metadata,omitempty"`
}

// WebhookResponse represents a webhook response
type WebhookResponse struct {
	ID              string                             `json:"id"`
	TenantID        string                             `json:"tenant_id"`
	Name            string                             `json:"name"`
	Description     string                             `json:"description"`
	URL             string                             `json:"url"`
	Events          []string                           `json:"events"`
	Status          string                             `json:"status"`
	RetryPolicy     *webhooks.RetryPolicy              `json:"retry_policy"`
	RateLimit       *webhooks.RateLimitConfig          `json:"rate_limit"`
	Headers         map[string]string                  `json:"headers"`
	Filters         *webhooks.WebhookEventFilterConfig `json:"filters"`
	Statistics      map[string]interface{}             `json:"statistics"`
	CreatedBy       string                             `json:"created_by"`
	CreatedAt       time.Time                          `json:"created_at"`
	UpdatedAt       time.Time                          `json:"updated_at"`
	LastTriggeredAt *time.Time                         `json:"last_triggered_at,omitempty"`
	Metadata        map[string]interface{}             `json:"metadata"`
}

// WebhookListResponse represents a list of webhooks
type WebhookListResponse struct {
	Webhooks []WebhookResponse `json:"webhooks"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// WebhookDeliveryResponse represents a webhook delivery response
type WebhookDeliveryResponse struct {
	ID              string                 `json:"id"`
	WebhookID       string                 `json:"webhook_id"`
	TenantID        string                 `json:"tenant_id"`
	EventType       string                 `json:"event_type"`
	EventID         string                 `json:"event_id"`
	Status          string                 `json:"status"`
	Attempts        int                    `json:"attempts"`
	MaxAttempts     int                    `json:"max_attempts"`
	ResponseCode    *int                   `json:"response_code,omitempty"`
	ResponseBody    string                 `json:"response_body,omitempty"`
	ResponseHeaders map[string]string      `json:"response_headers,omitempty"`
	Latency         *time.Duration         `json:"latency,omitempty"`
	Error           string                 `json:"error,omitempty"`
	NextRetryAt     *time.Time             `json:"next_retry_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	DeliveredAt     *time.Time             `json:"delivered_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// WebhookDeliveryListResponse represents a list of webhook deliveries
type WebhookDeliveryListResponse struct {
	Deliveries []WebhookDeliveryResponse `json:"deliveries"`
	Total      int                       `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
}

// WebhookTestRequest represents a webhook test request
type WebhookTestRequest struct {
	EventType string                 `json:"event_type" validate:"required"`
	EventID   string                 `json:"event_id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
}

// WebhookTestResponse represents a webhook test response
type WebhookTestResponse struct {
	Success      bool                   `json:"success"`
	DeliveryID   string                 `json:"delivery_id,omitempty"`
	ResponseCode int                    `json:"response_code,omitempty"`
	ResponseBody string                 `json:"response_body,omitempty"`
	Latency      time.Duration          `json:"latency,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookStatsResponse represents webhook statistics
type WebhookStatsResponse struct {
	WebhookID            string                    `json:"webhook_id"`
	TotalDeliveries      int                       `json:"total_deliveries"`
	SuccessfulDeliveries int                       `json:"successful_deliveries"`
	FailedDeliveries     int                       `json:"failed_deliveries"`
	SuccessRate          float64                   `json:"success_rate"`
	AverageLatency       float64                   `json:"average_latency"`
	LastDeliveryAt       *time.Time                `json:"last_delivery_at,omitempty"`
	RecentDeliveries     []WebhookDeliveryResponse `json:"recent_deliveries,omitempty"`
	Metadata             map[string]interface{}    `json:"metadata"`
}

// CreateWebhook creates a new webhook
func (h *WebhookHandlers) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)

	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create webhook request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateCreateWebhookRequest(&req); err != nil {
		h.logger.Error("Invalid create webhook request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set default values
	if req.Status == "" {
		req.Status = "active"
	}
	if req.RetryPolicy == nil {
		req.RetryPolicy = &webhooks.RetryPolicy{
			MaxAttempts: 3,
			BackoffType: "exponential",
			BaseDelay:   1 * time.Second,
			MaxDelay:    60 * time.Second,
		}
	}
	if req.RateLimit == nil {
		req.RateLimit = &webhooks.RateLimitConfig{
			RequestsPerMinute: 60,
			Burst:             10,
		}
	}

	// Create webhook request
	webhookRequest := &webhooks.WebhookRequest{
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Events:      convertStringEventsToWebhookEvents(req.Events),
		Secret:      req.Secret,
		Status:      webhooks.WebhookStatus(req.Status),
		RetryPolicy: *req.RetryPolicy,
		RateLimit:   *req.RateLimit,
		Headers:     req.Headers,
		Filters:     req.Filters,
		Metadata:    req.Metadata,
	}

	createdWebhook, err := h.webhookManager.CreateWebhook(ctx, webhookRequest)
	if err != nil {
		h.logger.Error("Failed to create webhook", zap.Error(err))
		http.Error(w, "Failed to create webhook", http.StatusInternalServerError)
		return
	}

	response := h.webhookToResponse(createdWebhook)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetWebhook retrieves a webhook by ID
func (h *WebhookHandlers) GetWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	webhook, err := h.webhookManager.GetWebhook(ctx, webhookID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get webhook", zap.Error(err), zap.String("webhook_id", webhookID))
		if err == webhooks.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get webhook", http.StatusInternalServerError)
		}
		return
	}

	response := h.webhookToResponse(webhook)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListWebhooks lists webhooks for a tenant
func (h *WebhookHandlers) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)

	// Parse query parameters
	page, pageSize := parsePaginationParams(r)
	status := r.URL.Query().Get("status")
	eventType := r.URL.Query().Get("event_type")
	search := r.URL.Query().Get("search")

	// Build filters
	filters := webhooks.WebhookFilters{
		Status:    status,
		EventType: eventType,
		Search:    search,
	}

	webhooks, total, err := h.webhookManager.ListWebhooks(ctx, tenantID, filters, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list webhooks", zap.Error(err))
		http.Error(w, "Failed to list webhooks", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	webhookResponses := make([]WebhookResponse, len(webhooks))
	for i, webhook := range webhooks {
		webhookResponses[i] = h.webhookToResponse(webhook)
	}

	response := WebhookListResponse{
		Webhooks: webhookResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateWebhook updates a webhook
func (h *WebhookHandlers) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	var req UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode update webhook request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing webhook
	existingWebhook, err := h.webhookManager.GetWebhook(ctx, webhookID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get webhook for update", zap.Error(err), zap.String("webhook_id", webhookID))
		if err == webhooks.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get webhook", http.StatusInternalServerError)
		}
		return
	}

	// Update fields
	if req.Name != nil {
		existingWebhook.Name = *req.Name
	}
	if req.Description != nil {
		existingWebhook.Description = *req.Description
	}
	if req.URL != nil {
		existingWebhook.URL = *req.URL
	}
	if req.Events != nil {
		existingWebhook.Events = req.Events
	}
	if req.Secret != nil {
		existingWebhook.Secret = *req.Secret
	}
	if req.Status != nil {
		existingWebhook.Status = *req.Status
	}
	if req.RetryPolicy != nil {
		existingWebhook.RetryPolicy = *req.RetryPolicy
	}
	if req.RateLimit != nil {
		existingWebhook.RateLimit = *req.RateLimit
	}
	if req.Headers != nil {
		existingWebhook.Headers = req.Headers
	}
	if req.Filters != nil {
		existingWebhook.Filters = req.Filters
	}
	if req.Metadata != nil {
		existingWebhook.Metadata = req.Metadata
	}

	updatedWebhook, err := h.webhookManager.UpdateWebhook(ctx, existingWebhook)
	if err != nil {
		h.logger.Error("Failed to update webhook", zap.Error(err), zap.String("webhook_id", webhookID))
		http.Error(w, "Failed to update webhook", http.StatusInternalServerError)
		return
	}

	response := h.webhookToResponse(updatedWebhook)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteWebhook deletes a webhook
func (h *WebhookHandlers) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	err := h.webhookManager.DeleteWebhook(ctx, webhookID, tenantID)
	if err != nil {
		h.logger.Error("Failed to delete webhook", zap.Error(err), zap.String("webhook_id", webhookID))
		if err == webhooks.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete webhook", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestWebhook tests a webhook with a sample event
func (h *WebhookHandlers) TestWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	var req WebhookTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode test webhook request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateWebhookTestRequest(&req); err != nil {
		h.logger.Error("Invalid test webhook request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get webhook
	webhook, err := h.webhookManager.GetWebhook(ctx, webhookID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get webhook for test", zap.Error(err), zap.String("webhook_id", webhookID))
		if err == webhooks.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get webhook", http.StatusInternalServerError)
		}
		return
	}

	// Create test event
	eventID := req.EventID
	if eventID == "" {
		eventID = fmt.Sprintf("test_%d", time.Now().Unix())
	}

	event := &webhooks.WebhookEvent{
		ID:        eventID,
		TenantID:  tenantID,
		EventType: req.EventType,
		Data:      req.Payload,
		Source:    "test",
		Version:   "1.0",
		Metadata:  req.Metadata,
	}

	// Dispatch test event
	deliveryID, err := h.webhookManager.DispatchEvent(ctx, webhook, event)
	if err != nil {
		h.logger.Error("Failed to dispatch test event", zap.Error(err))
		http.Error(w, "Failed to test webhook", http.StatusInternalServerError)
		return
	}

	// Get delivery result
	delivery, err := h.deliveryTracker.GetDelivery(ctx, deliveryID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get delivery result", zap.Error(err))
		http.Error(w, "Failed to get test result", http.StatusInternalServerError)
		return
	}

	response := WebhookTestResponse{
		Success:      delivery.Status == "delivered",
		DeliveryID:   delivery.ID,
		ResponseCode: delivery.ResponseCode,
		ResponseBody: delivery.ResponseBody,
		Latency:      delivery.Latency,
		Error:        delivery.Error,
		Metadata:     delivery.Metadata,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RetryWebhookDelivery retries a failed webhook delivery
func (h *WebhookHandlers) RetryWebhookDelivery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	deliveryID := vars["delivery_id"]

	// Get delivery
	delivery, err := h.deliveryTracker.GetDelivery(ctx, deliveryID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get delivery for retry", zap.Error(err), zap.String("delivery_id", deliveryID))
		if err == webhooks.ErrDeliveryNotFound {
			http.Error(w, "Delivery not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get delivery", http.StatusInternalServerError)
		}
		return
	}

	// Check if delivery can be retried
	if delivery.Status == "delivered" {
		http.Error(w, "Delivery already successful", http.StatusBadRequest)
		return
	}

	if delivery.Attempts >= delivery.MaxAttempts {
		http.Error(w, "Maximum retry attempts exceeded", http.StatusBadRequest)
		return
	}

	// Retry delivery
	err = h.retryHandler.RetryDelivery(ctx, delivery)
	if err != nil {
		h.logger.Error("Failed to retry delivery", zap.Error(err), zap.String("delivery_id", deliveryID))
		http.Error(w, "Failed to retry delivery", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetWebhookDeliveries lists webhook deliveries
func (h *WebhookHandlers) GetWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	// Parse query parameters
	page, pageSize := parsePaginationParams(r)
	status := r.URL.Query().Get("status")
	eventType := r.URL.Query().Get("event_type")

	// Build filters
	filters := webhooks.DeliveryFilters{
		Status:    status,
		EventType: eventType,
	}

	deliveries, total, err := h.deliveryTracker.ListDeliveries(ctx, webhookID, tenantID, filters, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list webhook deliveries", zap.Error(err))
		http.Error(w, "Failed to list deliveries", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	deliveryResponses := make([]WebhookDeliveryResponse, len(deliveries))
	for i, delivery := range deliveries {
		deliveryResponses[i] = h.deliveryToResponse(delivery)
	}

	response := WebhookDeliveryListResponse{
		Deliveries: deliveryResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWebhookStats gets webhook statistics
func (h *WebhookHandlers) GetWebhookStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	// Get webhook
	webhook, err := h.webhookManager.GetWebhook(ctx, webhookID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get webhook for stats", zap.Error(err), zap.String("webhook_id", webhookID))
		if err == webhooks.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get webhook", http.StatusInternalServerError)
		}
		return
	}

	// Get recent deliveries
	deliveries, _, err := h.deliveryTracker.ListDeliveries(ctx, webhookID, tenantID, webhooks.DeliveryFilters{}, 1, 10)
	if err != nil {
		h.logger.Error("Failed to get recent deliveries", zap.Error(err))
		http.Error(w, "Failed to get webhook stats", http.StatusInternalServerError)
		return
	}

	// Calculate statistics
	stats := calculateWebhookStats(webhook, deliveries)

	response := WebhookStatsResponse{
		WebhookID:            webhook.ID,
		TotalDeliveries:      stats["total_deliveries"].(int),
		SuccessfulDeliveries: stats["successful_deliveries"].(int),
		FailedDeliveries:     stats["failed_deliveries"].(int),
		SuccessRate:          stats["success_rate"].(float64),
		AverageLatency:       stats["average_latency"].(float64),
		LastDeliveryAt:       stats["last_delivery_at"].(*time.Time),
		RecentDeliveries:     h.deliveriesToResponses(deliveries),
		Metadata:             webhook.Metadata,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func (h *WebhookHandlers) webhookToResponse(webhook *webhooks.WebhookResponse) WebhookResponse {
	return WebhookResponse{
		ID:              webhook.ID,
		TenantID:        webhook.TenantID,
		Name:            webhook.Name,
		Description:     webhook.Description,
		URL:             webhook.URL,
		Events:          convertWebhookEventsToStrings(webhook.Events),
		Status:          string(webhook.Status),
		RetryPolicy:     &webhook.RetryPolicy,
		RateLimit:       &webhook.RateLimit,
		Headers:         webhook.Headers,
		Filters:         webhook.Filters,
		Statistics:      webhook.Statistics,
		CreatedBy:       webhook.CreatedBy,
		CreatedAt:       webhook.CreatedAt,
		UpdatedAt:       webhook.UpdatedAt,
		LastTriggeredAt: webhook.LastTriggeredAt,
		Metadata:        webhook.Metadata,
	}
}

func (h *WebhookHandlers) deliveryToResponse(delivery *webhooks.WebhookDelivery) WebhookDeliveryResponse {
	return WebhookDeliveryResponse{
		ID:              delivery.ID,
		WebhookID:       delivery.WebhookID,
		TenantID:        delivery.TenantID,
		EventType:       string(delivery.EventType),
		EventID:         delivery.EventID,
		Status:          string(delivery.Status),
		Attempts:        delivery.Attempts,
		MaxAttempts:     delivery.MaxAttempts,
		ResponseCode:    delivery.ResponseCode,
		ResponseBody:    delivery.ResponseBody,
		ResponseHeaders: delivery.ResponseHeaders,
		Latency:         delivery.Latency,
		Error:           delivery.Error,
		NextRetryAt:     delivery.NextRetryAt,
		CreatedAt:       delivery.CreatedAt,
		DeliveredAt:     delivery.DeliveredAt,
		Metadata:        delivery.Metadata,
	}
}

func (h *WebhookHandlers) deliveriesToResponses(deliveries []*webhooks.WebhookDelivery) []WebhookDeliveryResponse {
	responses := make([]WebhookDeliveryResponse, len(deliveries))
	for i, delivery := range deliveries {
		responses[i] = h.deliveryToResponse(delivery)
	}
	return responses
}

func validateCreateWebhookRequest(req *CreateWebhookRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}
	if req.URL == "" {
		return fmt.Errorf("url is required")
	}
	if !isValidURL(req.URL) {
		return fmt.Errorf("url format is invalid")
	}
	if len(req.Events) == 0 {
		return fmt.Errorf("events are required")
	}
	if req.Status != "" && !isValidWebhookStatus(req.Status) {
		return fmt.Errorf("invalid status: %s", req.Status)
	}
	return nil
}

func validateWebhookTestRequest(req *WebhookTestRequest) error {
	if req.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	return nil
}

func isValidWebhookStatus(status string) bool {
	validStatuses := []string{"active", "inactive", "paused", "disabled"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func calculateWebhookStats(webhook *webhooks.Webhook, deliveries []*webhooks.WebhookDelivery) map[string]interface{} {
	stats := map[string]interface{}{
		"total_deliveries":      0,
		"successful_deliveries": 0,
		"failed_deliveries":     0,
		"success_rate":          0.0,
		"average_latency":       0.0,
		"last_delivery_at":      (*time.Time)(nil),
	}

	if len(deliveries) == 0 {
		return stats
	}

	totalDeliveries := len(deliveries)
	successfulDeliveries := 0
	failedDeliveries := 0
	totalLatency := time.Duration(0)
	latencyCount := 0
	var lastDeliveryAt *time.Time

	for _, delivery := range deliveries {
		if delivery.Status == "delivered" {
			successfulDeliveries++
		} else if delivery.Status == "failed" {
			failedDeliveries++
		}

		if delivery.Latency != nil {
			totalLatency += *delivery.Latency
			latencyCount++
		}

		if lastDeliveryAt == nil || delivery.CreatedAt.After(*lastDeliveryAt) {
			lastDeliveryAt = &delivery.CreatedAt
		}
	}

	successRate := 0.0
	if totalDeliveries > 0 {
		successRate = (float64(successfulDeliveries) / float64(totalDeliveries)) * 100
	}

	averageLatency := 0.0
	if latencyCount > 0 {
		averageLatency = float64(totalLatency.Milliseconds()) / float64(latencyCount)
	}

	stats["total_deliveries"] = totalDeliveries
	stats["successful_deliveries"] = successfulDeliveries
	stats["failed_deliveries"] = failedDeliveries
	stats["success_rate"] = successRate
	stats["average_latency"] = averageLatency
	stats["last_delivery_at"] = lastDeliveryAt

	return stats
}

func parsePaginationParams(r *http.Request) (int, int) {
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	return page, pageSize
}

func getTenantIDFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return "default"
}

func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return "system"
}

func convertStringEventsToWebhookEvents(events []string) []webhooks.WebhookEvent {
	webhookEvents := make([]webhooks.WebhookEvent, len(events))
	for i, event := range events {
		webhookEvents[i] = webhooks.WebhookEvent(event)
	}
	return webhookEvents
}

func convertWebhookEventsToStrings(events []webhooks.WebhookEvent) []string {
	strings := make([]string, len(events))
	for i, event := range events {
		strings[i] = string(event)
	}
	return strings
}
