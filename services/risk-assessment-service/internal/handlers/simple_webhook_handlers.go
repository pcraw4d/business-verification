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

// SimpleWebhookHandlers provides basic webhook functionality
type SimpleWebhookHandlers struct {
	webhookManager webhooks.WebhookManager
	logger         *zap.Logger
}

// NewSimpleWebhookHandlers creates a new SimpleWebhookHandlers instance
func NewSimpleWebhookHandlers(webhookManager webhooks.WebhookManager, logger *zap.Logger) *SimpleWebhookHandlers {
	return &SimpleWebhookHandlers{
		webhookManager: webhookManager,
		logger:         logger,
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

// CreateWebhook creates a new webhook
func (h *SimpleWebhookHandlers) CreateWebhook(w http.ResponseWriter, r *http.Request) {
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
			MaxRetries:      3,
			InitialInterval: 1 * time.Second,
			MaxInterval:     60 * time.Second,
			Multiplier:      2.0,
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
func (h *SimpleWebhookHandlers) GetWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	webhook, err := h.webhookManager.GetWebhook(ctx, tenantID, webhookID)
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
func (h *SimpleWebhookHandlers) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)

	// Parse query parameters
	page, pageSize := parsePaginationParams(r)
	status := r.URL.Query().Get("status")
	eventType := r.URL.Query().Get("event_type")
	search := r.URL.Query().Get("search")

	// Build filters
	filters := &webhooks.WebhookFilter{
		TenantID:  tenantID,
		Status:    status,
		EventType: eventType,
		Search:    search,
		Limit:     pageSize,
		Offset:    (page - 1) * pageSize,
	}

	webhookList, err := h.webhookManager.ListWebhooks(ctx, filters)
	if err != nil {
		h.logger.Error("Failed to list webhooks", zap.Error(err))
		http.Error(w, "Failed to list webhooks", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	webhookResponses := make([]WebhookResponse, len(webhookList.Webhooks))
	for i, webhook := range webhookList.Webhooks {
		webhookResponses[i] = h.webhookToResponse(webhook)
	}

	response := map[string]interface{}{
		"webhooks":  webhookResponses,
		"total":     webhookList.Total,
		"page":      page,
		"page_size": pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateWebhook updates a webhook
func (h *SimpleWebhookHandlers) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode update webhook request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
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

	updatedWebhook, err := h.webhookManager.UpdateWebhook(ctx, tenantID, webhookID, webhookRequest)
	if err != nil {
		h.logger.Error("Failed to update webhook", zap.Error(err), zap.String("webhook_id", webhookID))
		if err == webhooks.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update webhook", http.StatusInternalServerError)
		}
		return
	}

	response := h.webhookToResponse(updatedWebhook)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteWebhook deletes a webhook
func (h *SimpleWebhookHandlers) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	vars := mux.Vars(r)
	webhookID := vars["id"]

	err := h.webhookManager.DeleteWebhook(ctx, tenantID, webhookID)
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

// Helper functions

func (h *SimpleWebhookHandlers) webhookToResponse(webhook *webhooks.WebhookResponse) WebhookResponse {
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
