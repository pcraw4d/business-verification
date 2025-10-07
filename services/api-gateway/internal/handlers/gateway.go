package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/config"
	"kyb-platform/services/api-gateway/internal/supabase"
)

// GatewayHandler handles API Gateway requests
type GatewayHandler struct {
	supabaseClient *supabase.Client
	logger         *zap.Logger
	config         *config.Config
	httpClient     *http.Client
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(supabaseClient *supabase.Client, logger *zap.Logger, cfg *config.Config) *GatewayHandler {
	return &GatewayHandler{
		supabaseClient: supabaseClient,
		logger:         logger,
		config:         cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HealthCheck handles health check requests
func (h *GatewayHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check Supabase connection
	supabaseStatus := "connected"
	if err := h.supabaseClient.HealthCheck(ctx); err != nil {
		supabaseStatus = "disconnected"
		h.logger.Error("Supabase health check failed", zap.Error(err))
	}

	// Get table counts for monitoring
	tableCounts := make(map[string]int)
	tables := []string{"classifications", "merchants", "risk_keywords", "business_risk_assessments"}

	for _, table := range tables {
		if count, err := h.supabaseClient.GetTableCount(ctx, table); err == nil {
			tableCounts[table] = count
		}
	}

	response := map[string]interface{}{
		"status":      "healthy",
		"service":     "api-gateway",
		"version":     "1.0.0",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"environment": h.config.Environment,
		"supabase_status": map[string]interface{}{
			"connected": supabaseStatus == "connected",
			"url":       h.config.Supabase.URL,
		},
		"table_counts": tableCounts,
		"features": map[string]bool{
			"supabase_integration": true,
			"authentication":       true,
			"rate_limiting":        h.config.RateLimit.Enabled,
			"cors_enabled":         true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ProxyToClassification proxies requests to the classification service
func (h *GatewayHandler) ProxyToClassification(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.config.Services.ClassificationURL, "/classify")
}

// ProxyToMerchants proxies requests to the merchant service
func (h *GatewayHandler) ProxyToMerchants(w http.ResponseWriter, r *http.Request) {
	// The route is registered as /merchants in the /api/v1 subrouter
	// So r.URL.Path will be /api/v1/merchants or /api/v1/merchants/{id}
	// We need to pass the full path to the merchant service
	h.proxyRequest(w, r, h.config.Services.MerchantURL, r.URL.Path)
}

// ProxyToClassificationHealth proxies health check requests to the classification service
func (h *GatewayHandler) ProxyToClassificationHealth(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.config.Services.ClassificationURL, "/health")
}

// ProxyToMerchantHealth proxies health check requests to the merchant service
func (h *GatewayHandler) ProxyToMerchantHealth(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.config.Services.MerchantURL, "/health")
}

// proxyRequest proxies a request to a backend service
func (h *GatewayHandler) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL, targetPath string) {
	ctx := r.Context()

	// Log the proxy request
	h.logger.Info("Proxying request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("target", targetURL+targetPath),
		zap.String("user_agent", r.Header.Get("User-Agent")))

	// Create the target URL
	target := targetURL + targetPath
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	// Read the request body
	var body io.Reader
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error("Failed to read request body", zap.Error(err))
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Create the proxy request
	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, target, body)
	if err != nil {
		h.logger.Error("Failed to create proxy request", zap.Error(err))
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request
	for key, values := range r.Header {
		// Skip headers that shouldn't be forwarded
		if key == "Host" || key == "Connection" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Set the Host header to the target
	proxyReq.Host = strings.TrimPrefix(targetURL, "https://")

	// Make the request
	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		h.logger.Error("Proxy request failed", zap.Error(err))
		http.Error(w, "Backend service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		// Skip headers that shouldn't be forwarded
		if key == "Connection" || key == "Transfer-Encoding" {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		h.logger.Error("Failed to copy response body", zap.Error(err))
	}

	h.logger.Info("Proxy request completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", resp.StatusCode))
}

// ProxyToBI proxies requests to the Business Intelligence service
func (h *GatewayHandler) ProxyToBI(w http.ResponseWriter, r *http.Request) {
	// Extract the path after /api/v1/bi/
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/bi")
	if path == "" {
		path = "/"
	}
	
	// Add query parameters if any
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	
	h.proxyRequest(w, r, h.config.Services.BIServiceURL+path)
}
