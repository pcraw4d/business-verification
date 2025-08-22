package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pcraw4d/business-verification/internal/external"
	"go.uber.org/zap"
)

// FallbackStrategiesHandler handles fallback strategy API endpoints
type FallbackStrategiesHandler struct {
	manager *external.FallbackStrategyManager
	logger  *zap.Logger
}

// NewFallbackStrategiesHandler creates a new fallback strategies handler
func NewFallbackStrategiesHandler(manager *external.FallbackStrategyManager, logger *zap.Logger) *FallbackStrategiesHandler {
	return &FallbackStrategiesHandler{
		manager: manager,
		logger:  logger,
	}
}

// ExecuteFallbackStrategiesRequest represents the request for executing fallback strategies
type ExecuteFallbackStrategiesRequest struct {
	URL           string `json:"url" validate:"required,url"`
	OriginalError string `json:"original_error" validate:"required"`
}

// ExecuteFallbackStrategiesResponse represents the response from executing fallback strategies
type ExecuteFallbackStrategiesResponse struct {
	Success bool                     `json:"success"`
	Result  *external.FallbackResult `json:"result,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// ExecuteFallbackStrategies executes fallback strategies for a blocked website
func (h *FallbackStrategiesHandler) ExecuteFallbackStrategies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req ExecuteFallbackStrategiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Validate request
	if req.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "url is required"})
		return
	}

	if req.OriginalError == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "original_error is required"})
		return
	}

	// Execute fallback strategies
	result, err := h.manager.ExecuteFallbackStrategies(r.Context(), req.URL, &external.BlockingError{Message: req.OriginalError})
	if err != nil {
		h.logger.Error("Failed to execute fallback strategies",
			zap.String("url", req.URL),
			zap.Error(err))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ExecuteFallbackStrategiesResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ExecuteFallbackStrategiesResponse{
		Success: true,
		Result:  result,
	})
}

// GetConfigRequest represents the request for getting fallback configuration
type GetConfigRequest struct{}

// GetConfigResponse represents the response for getting fallback configuration
type GetConfigResponse struct {
	Success bool                     `json:"success"`
	Config  *external.FallbackConfig `json:"config,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// GetConfig returns the current fallback configuration
func (h *FallbackStrategiesHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	config := h.manager.GetConfig()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetConfigResponse{
		Success: true,
		Config:  config,
	})
}

// UpdateFallbackConfigRequest represents the request for updating fallback configuration
type UpdateFallbackConfigRequest struct {
	Config *external.FallbackConfig `json:"config" validate:"required"`
}

// UpdateFallbackConfigResponse represents the response for updating fallback configuration
type UpdateFallbackConfigResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// UpdateConfig updates the fallback configuration
func (h *FallbackStrategiesHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req UpdateFallbackConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Validate request
	if req.Config == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "config is required"})
		return
	}

	// Update configuration
	h.manager.UpdateConfig(req.Config)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UpdateFallbackConfigResponse{
		Success: true,
	})
}

// AddProxyRequest represents the request for adding a proxy
type AddProxyRequest struct {
	Proxy *external.Proxy `json:"proxy" validate:"required"`
}

// AddProxyResponse represents the response for adding a proxy
type AddProxyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// AddProxy adds a proxy to the proxy pool
func (h *FallbackStrategiesHandler) AddProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req AddProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Validate request
	if req.Proxy == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "proxy is required"})
		return
	}

	if req.Proxy.Host == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "proxy host is required"})
		return
	}

	if req.Proxy.Port <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "proxy port must be greater than 0"})
		return
	}

	// Add proxy
	h.manager.AddProxy(*req.Proxy)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AddProxyResponse{
		Success: true,
	})
}

// RemoveProxyRequest represents the request for removing a proxy
type RemoveProxyRequest struct {
	Host string `json:"host" validate:"required"`
	Port int    `json:"port" validate:"required,min=1,max=65535"`
}

// RemoveProxyResponse represents the response for removing a proxy
type RemoveProxyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RemoveProxy removes a proxy from the proxy pool
func (h *FallbackStrategiesHandler) RemoveProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req RemoveProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Validate request
	if req.Host == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "host is required"})
		return
	}

	if req.Port <= 0 || req.Port > 65535 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "port must be between 1 and 65535"})
		return
	}

	// Remove proxy
	h.manager.RemoveProxy(req.Host, req.Port)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RemoveProxyResponse{
		Success: true,
	})
}

// TestFallbackStrategyRequest represents the request for testing a fallback strategy
type TestFallbackStrategyRequest struct {
	URL      string `json:"url" validate:"required,url"`
	Strategy string `json:"strategy" validate:"required,oneof=user_agent_rotation header_customization proxy_rotation alternative_sources"`
}

// TestFallbackStrategyResponse represents the response for testing a fallback strategy
type TestFallbackStrategyResponse struct {
	Success bool                     `json:"success"`
	Result  *external.FallbackResult `json:"result,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// TestFallbackStrategy tests a specific fallback strategy
func (h *FallbackStrategiesHandler) TestFallbackStrategy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req TestFallbackStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Validate request
	if req.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "url is required"})
		return
	}

	if req.Strategy == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "strategy is required"})
		return
	}

	// Test specific strategy
	var result *external.FallbackResult
	var err error

	switch req.Strategy {
	case "user_agent_rotation":
		result = h.manager.TryUserAgentRotation(r.Context(), req.URL)
	case "header_customization":
		result = h.manager.TryHeaderCustomization(r.Context(), req.URL)
	case "proxy_rotation":
		result = h.manager.TryProxyRotation(r.Context(), req.URL)
	case "alternative_sources":
		result = h.manager.TryAlternativeDataSources(r.Context(), req.URL)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid strategy"})
		return
	}

	if err != nil {
		h.logger.Error("Failed to test fallback strategy",
			zap.String("url", req.URL),
			zap.String("strategy", req.Strategy),
			zap.Error(err))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TestFallbackStrategyResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TestFallbackStrategyResponse{
		Success: true,
		Result:  result,
	})
}

// RegisterRoutes registers the fallback strategies routes
func (h *FallbackStrategiesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/fallback/execute", h.ExecuteFallbackStrategies)
	mux.HandleFunc("GET /api/v1/fallback/config", h.GetConfig)
	mux.HandleFunc("PUT /api/v1/fallback/config", h.UpdateConfig)
	mux.HandleFunc("POST /api/v1/fallback/proxy", h.AddProxy)
	mux.HandleFunc("DELETE /api/v1/fallback/proxy", h.RemoveProxy)
	mux.HandleFunc("POST /api/v1/fallback/test", h.TestFallbackStrategy)
}
