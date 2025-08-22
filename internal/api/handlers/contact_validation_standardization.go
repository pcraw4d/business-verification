package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/external"
)

// ContactValidationHandler handles contact validation and standardization requests
type ContactValidationHandler struct {
	standardizer *external.ContactValidationStandardizer
	logger       *zap.Logger
}

// NewContactValidationHandler creates a new contact validation handler
func NewContactValidationHandler(standardizer *external.ContactValidationStandardizer, logger *zap.Logger) *ContactValidationHandler {
	return &ContactValidationHandler{
		standardizer: standardizer,
		logger:       logger,
	}
}

// ValidationRequest represents a validation request
type ValidationRequest struct {
	ContactType string `json:"contact_type"` // "phone", "email", "address"
	Value       string `json:"value"`
}

// ValidationResponse represents a validation response
type ValidationResponse struct {
	Success bool                       `json:"success"`
	Result  *external.ValidationResult `json:"result,omitempty"`
	Error   string                     `json:"error,omitempty"`
	Message string                     `json:"message,omitempty"`
}

// BatchValidationRequest represents a batch validation request
type BatchValidationRequest struct {
	ContactType string   `json:"contact_type"` // "phone", "email", "address"
	Values      []string `json:"values"`
}

// BatchValidationResponse represents a batch validation response
type BatchValidationResponse struct {
	Success bool                            `json:"success"`
	Result  *external.BatchValidationResult `json:"result,omitempty"`
	Error   string                          `json:"error,omitempty"`
	Message string                          `json:"message,omitempty"`
}

// ConfigRequest represents a configuration update request
type ConfigRequest struct {
	Config *external.ContactValidationConfig `json:"config"`
}

// ConfigResponse represents a configuration response
type ConfigResponse struct {
	Success bool                              `json:"success"`
	Config  *external.ContactValidationConfig `json:"config,omitempty"`
	Error   string                            `json:"error,omitempty"`
	Message string                            `json:"message,omitempty"`
}

// StatisticsResponse represents statistics response
type StatisticsResponse struct {
	Success bool                   `json:"success"`
	Stats   map[string]interface{} `json:"stats,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// RegisterRoutes registers the contact validation routes
func (h *ContactValidationHandler) RegisterRoutes(router *mux.Router) {
	// Single validation endpoints
	router.HandleFunc("/validate/phone", h.ValidatePhone).Methods("POST")
	router.HandleFunc("/validate/email", h.ValidateEmail).Methods("POST")
	router.HandleFunc("/validate/address", h.ValidateAddress).Methods("POST")

	// Batch validation endpoint
	router.HandleFunc("/validate/batch", h.ValidateBatch).Methods("POST")

	// Configuration management
	router.HandleFunc("/config", h.GetConfig).Methods("GET")
	router.HandleFunc("/config", h.UpdateConfig).Methods("PUT")

	// Statistics and monitoring
	router.HandleFunc("/stats", h.GetStatistics).Methods("GET")
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
}

// ValidatePhone validates a phone number
func (h *ContactValidationHandler) ValidatePhone(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Value == "" {
		h.sendErrorResponse(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := h.standardizer.ValidatePhoneNumber(ctx, req.Value)
	if err != nil {
		h.logger.Error("phone validation failed", zap.Error(err))
		h.sendErrorResponse(w, "Phone validation failed", http.StatusInternalServerError)
		return
	}

	response := ValidationResponse{
		Success: true,
		Result:  result,
		Message: "Phone validation completed",
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// ValidateEmail validates an email address
func (h *ContactValidationHandler) ValidateEmail(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Value == "" {
		h.sendErrorResponse(w, "Email address is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := h.standardizer.ValidateEmailAddress(ctx, req.Value)
	if err != nil {
		h.logger.Error("email validation failed", zap.Error(err))
		h.sendErrorResponse(w, "Email validation failed", http.StatusInternalServerError)
		return
	}

	response := ValidationResponse{
		Success: true,
		Result:  result,
		Message: "Email validation completed",
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// ValidateAddress validates a physical address
func (h *ContactValidationHandler) ValidateAddress(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Value == "" {
		h.sendErrorResponse(w, "Address is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := h.standardizer.ValidatePhysicalAddress(ctx, req.Value)
	if err != nil {
		h.logger.Error("address validation failed", zap.Error(err))
		h.sendErrorResponse(w, "Address validation failed", http.StatusInternalServerError)
		return
	}

	response := ValidationResponse{
		Success: true,
		Result:  result,
		Message: "Address validation completed",
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// ValidateBatch validates multiple contact items in batch
func (h *ContactValidationHandler) ValidateBatch(w http.ResponseWriter, r *http.Request) {
	var req BatchValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.ContactType == "" {
		h.sendErrorResponse(w, "Contact type is required", http.StatusBadRequest)
		return
	}

	if len(req.Values) == 0 {
		h.sendErrorResponse(w, "At least one value is required", http.StatusBadRequest)
		return
	}

	// Validate contact type
	validTypes := map[string]bool{"phone": true, "email": true, "address": true}
	if !validTypes[strings.ToLower(req.ContactType)] {
		h.sendErrorResponse(w, "Invalid contact type. Must be 'phone', 'email', or 'address'", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := h.standardizer.ValidateBatch(ctx, req.Values, req.ContactType)
	if err != nil {
		h.logger.Error("batch validation failed", zap.Error(err))
		h.sendErrorResponse(w, "Batch validation failed", http.StatusInternalServerError)
		return
	}

	response := BatchValidationResponse{
		Success: true,
		Result:  result,
		Message: fmt.Sprintf("Batch validation completed for %d items", len(req.Values)),
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// GetConfig returns the current configuration
func (h *ContactValidationHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.standardizer.GetConfig()

	response := ConfigResponse{
		Success: true,
		Config:  config,
		Message: "Configuration retrieved successfully",
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// UpdateConfig updates the validation configuration
func (h *ContactValidationHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Config == nil {
		h.sendErrorResponse(w, "Configuration is required", http.StatusBadRequest)
		return
	}

	// Validate configuration
	if err := h.validateConfig(req.Config); err != nil {
		h.sendErrorResponse(w, fmt.Sprintf("Invalid configuration: %s", err.Error()), http.StatusBadRequest)
		return
	}

	err := h.standardizer.UpdateConfig(req.Config)
	if err != nil {
		h.logger.Error("config update failed", zap.Error(err))
		h.sendErrorResponse(w, "Configuration update failed", http.StatusInternalServerError)
		return
	}

	response := ConfigResponse{
		Success: true,
		Config:  req.Config,
		Message: "Configuration updated successfully",
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// GetStatistics returns validation statistics
func (h *ContactValidationHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	config := h.standardizer.GetConfig()

	stats := map[string]interface{}{
		"phone_validation_enabled":    config.EnablePhoneValidation,
		"email_validation_enabled":    config.EnableEmailValidation,
		"address_validation_enabled":  config.EnableAddressValidation,
		"min_confidence_threshold":    config.MinValidationConfidence,
		"max_batch_size":              config.MaxBatchSize,
		"validation_timeout":          config.ValidationTimeout.String(),
		"e164_format_enabled":         config.EnableE164Format,
		"domain_validation_enabled":   config.EnableDomainValidation,
		"mx_validation_enabled":       config.EnableMXValidation,
		"geocoding_enabled":           config.EnableGeocoding,
		"postal_code_validation":      config.EnablePostalCodeValidation,
		"phone_standardization":       config.EnablePhoneStandardization,
		"email_standardization":       config.EnableEmailStandardization,
		"address_standardization":     config.EnableAddressStandardization,
		"fuzzy_matching_enabled":      config.EnableFuzzyMatching,
		"auto_correction_enabled":     config.EnableAutoCorrection,
		"caching_enabled":             config.EnableCaching,
		"blocked_domains_count":       len(config.BlockedDomains),
		"trusted_domains_count":       len(config.TrustedDomains),
		"allowed_country_codes_count": len(config.AllowedCountryCodes),
		"supported_countries_count":   len(config.SupportedCountries),
		"timestamp":                   time.Now().UTC().Format(time.RFC3339),
	}

	response := StatisticsResponse{
		Success: true,
		Stats:   stats,
		Message: "Statistics retrieved successfully",
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// HealthCheck performs a health check
func (h *ContactValidationHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Perform a simple validation to test the system
	testPhone := "+1234567890"
	ctx := r.Context()

	_, err := h.standardizer.ValidatePhoneNumber(ctx, testPhone)
	if err != nil {
		h.logger.Error("health check failed", zap.Error(err))
		h.sendErrorResponse(w, "Health check failed", http.StatusServiceUnavailable)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "contact_validation",
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// validateConfig validates the configuration
func (h *ContactValidationHandler) validateConfig(config *external.ContactValidationConfig) error {
	if config.MinValidationConfidence < 0.0 || config.MinValidationConfidence > 1.0 {
		return fmt.Errorf("min_validation_confidence must be between 0.0 and 1.0")
	}

	if config.MaxBatchSize <= 0 {
		return fmt.Errorf("max_batch_size must be greater than 0")
	}

	if config.ValidationTimeout <= 0 {
		return fmt.Errorf("validation_timeout must be greater than 0")
	}

	if config.DefaultCountryCode != "" {
		// Validate country code format (simple check)
		if len(config.DefaultCountryCode) > 3 {
			return fmt.Errorf("default_country_code must be 1-3 digits")
		}
	}

	return nil
}

// sendJSONResponse sends a JSON response
func (h *ContactValidationHandler) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendErrorResponse sends an error response
func (h *ContactValidationHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := ValidationResponse{
		Success: false,
		Error:   message,
		Message: "Validation failed",
	}

	h.sendJSONResponse(w, response, statusCode)
}
