package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// APIKeyService handles API key management for integrations
type APIKeyService struct {
	db     database.Database
	logger *observability.Logger
}

// NewAPIKeyService creates a new API key service instance
func NewAPIKeyService(db database.Database, logger *observability.Logger) *APIKeyService {
	return &APIKeyService{
		db:     db,
		logger: logger,
	}
}

// CreateAPIKeyRequest represents a request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string     `json:"name" validate:"required,min=1,max=100"`
	UserID      string     `json:"user_id" validate:"required"`
	Role        Role       `json:"role" validate:"required"`
	Permissions []string   `json:"permissions,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Description string     `json:"description,omitempty"`
}

// CreateAPIKeyResponse represents the response after API key creation
type CreateAPIKeyResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Key         string     `json:"key"` // Only returned on creation
	UserID      string     `json:"user_id"`
	Role        Role       `json:"role"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	Status      string     `json:"status"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateAPIKey creates a new API key for integrations
func (aks *APIKeyService) CreateAPIKey(ctx context.Context, request *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	// Validate the request
	if err := aks.validateCreateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if user exists and has permission to create API keys
	user, err := aks.db.GetUserByID(ctx, request.UserID)
	if err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to get user for API key creation")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate role assignment
	if !CanAssignRole(Role(user.Role), request.Role) {
		aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
			"user_role":   user.Role,
			"target_role": request.Role,
		}).Warn("User attempted to create API key with unauthorized role")
		return nil, fmt.Errorf("insufficient permissions to assign role %s", request.Role)
	}

	// Generate API key
	key, keyHash, err := aks.generateAPIKey()
	if err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to generate API key")
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create API key record
	apiKey := &database.APIKey{
		ID:          generateUUID(),
		Name:        request.Name,
		KeyHash:     keyHash, // Store hash, not the actual key
		UserID:      request.UserID,
		Role:        string(request.Role),
		Permissions: strings.Join(request.Permissions, ","), // Store as comma-separated string
		ExpiresAt:   request.ExpiresAt,
		LastUsedAt:  nil,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store in database
	if err := aks.db.CreateAPIKey(ctx, apiKey); err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to store API key in database")
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Log the creation
	aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
		"api_key_id": apiKey.ID,
		"user_id":    request.UserID,
		"role":       request.Role,
		"name":       request.Name,
	}).Info("API key created successfully")

	// Return response with the actual key (only on creation)
	response := &CreateAPIKeyResponse{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Key:         key, // Return the actual key only once
		UserID:      apiKey.UserID,
		Role:        Role(apiKey.Role),
		Permissions: request.Permissions,
		ExpiresAt:   apiKey.ExpiresAt,
		LastUsedAt:  apiKey.LastUsedAt,
		Status:      apiKey.Status,
		Description: request.Description,
		CreatedAt:   apiKey.CreatedAt,
		UpdatedAt:   apiKey.UpdatedAt,
	}

	return response, nil
}

// ValidateAPIKey validates an API key and returns user context
func (aks *APIKeyService) ValidateAPIKey(ctx context.Context, key string) (*APIKeyContext, error) {
	if key == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Hash the provided key
	keyHash := aks.hashAPIKey(key)

	// Look up the API key in database
	apiKey, err := aks.db.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Warn("Failed to find API key")
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if API key is active
	if apiKey.Status != "active" {
		aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
			"api_key_id": apiKey.ID,
		}).Warn("Attempted to use inactive API key")
		return nil, fmt.Errorf("API key is inactive")
	}

	// Check if API key has expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
			"api_key_id": apiKey.ID,
			"expires_at": apiKey.ExpiresAt,
		}).Warn("Attempted to use expired API key")
		return nil, fmt.Errorf("API key has expired")
	}

	// Update last used timestamp
	now := time.Now()
	if err := aks.db.UpdateAPIKeyLastUsed(ctx, apiKey.ID, now); err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to update API key last used timestamp")
		// Don't fail the request for this error
	}

	// Parse permissions
	permissions := []string{}
	if apiKey.Permissions != "" {
		permissions = strings.Split(apiKey.Permissions, ",")
	}

	// Create API key context
	context := &APIKeyContext{
		APIKeyID:    apiKey.ID,
		UserID:      apiKey.UserID,
		Role:        Role(apiKey.Role),
		Permissions: permissions,
		LastUsedAt:  &now,
	}

	aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
		"api_key_id": apiKey.ID,
		"user_id":    apiKey.UserID,
		"role":       apiKey.Role,
	}).Debug("API key validated successfully")

	return context, nil
}

// APIKeyContext represents the context of a validated API key
type APIKeyContext struct {
	APIKeyID    string     `json:"api_key_id"`
	UserID      string     `json:"user_id"`
	Role        Role       `json:"role"`
	Permissions []string   `json:"permissions"`
	LastUsedAt  *time.Time `json:"last_used_at"`
}

// ListAPIKeysRequest represents a request to list API keys
type ListAPIKeysRequest struct {
	UserID   string `json:"user_id,omitempty"`
	Role     Role   `json:"role,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// ListAPIKeysResponse represents the response for listing API keys
type ListAPIKeysResponse struct {
	APIKeys []*APIKeyResponse `json:"api_keys"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}

// ListAPIKeys lists API keys based on filters
func (aks *APIKeyService) ListAPIKeys(ctx context.Context, request *ListAPIKeysRequest) (*ListAPIKeysResponse, error) {
	// TODO: Implement pagination and filtering
	// For now, return a simple implementation
	apiKeys, err := aks.db.ListAPIKeysByUserID(ctx, request.UserID)
	if err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to list API keys")
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Convert to response format
	responses := make([]*APIKeyResponse, 0, len(apiKeys))
	for _, apiKey := range apiKeys {
		permissions := []string{}
		if apiKey.Permissions != "" {
			permissions = strings.Split(apiKey.Permissions, ",")
		}

		response := &APIKeyResponse{
			ID:          apiKey.ID,
			Name:        apiKey.Name,
			UserID:      apiKey.UserID,
			Role:        Role(apiKey.Role),
			Permissions: permissions,
			ExpiresAt:   apiKey.ExpiresAt,
			LastUsedAt:  apiKey.LastUsedAt,
			IsActive:    apiKey.Status == "active",
			CreatedAt:   apiKey.CreatedAt,
			UpdatedAt:   apiKey.UpdatedAt,
		}
		responses = append(responses, response)
	}

	return &ListAPIKeysResponse{
		APIKeys: responses,
		Total:   len(responses),
		Limit:   request.Limit,
		Offset:  request.Offset,
	}, nil
}

// RevokeAPIKeyRequest represents a request to revoke an API key
type RevokeAPIKeyRequest struct {
	APIKeyID string `json:"api_key_id" validate:"required"`
	UserID   string `json:"user_id" validate:"required"`
	Reason   string `json:"reason,omitempty"`
}

// RevokeAPIKey revokes an API key
func (aks *APIKeyService) RevokeAPIKey(ctx context.Context, request *RevokeAPIKeyRequest) error {
	// Get the API key
	apiKey, err := aks.db.GetAPIKeyByID(ctx, request.APIKeyID)
	if err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to get API key for revocation")
		return fmt.Errorf("API key not found: %w", err)
	}

	// Check if user has permission to revoke this API key
	if apiKey.UserID != request.UserID {
		// TODO: Add admin permission check
		aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
			"api_key_user_id": apiKey.UserID,
			"request_user_id": request.UserID,
		}).Warn("User attempted to revoke API key they don't own")
		return fmt.Errorf("insufficient permissions to revoke this API key")
	}

	// Deactivate the API key
	if err := aks.db.DeactivateAPIKey(ctx, request.APIKeyID); err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to deactivate API key")
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
		"api_key_id": request.APIKeyID,
		"user_id":    request.UserID,
		"reason":     request.Reason,
	}).Info("API key revoked successfully")

	return nil
}

// UpdateAPIKeyRequest represents a request to update an API key
type UpdateAPIKeyRequest struct {
	APIKeyID    string     `json:"api_key_id" validate:"required"`
	UserID      string     `json:"user_id" validate:"required"`
	Name        string     `json:"name,omitempty"`
	Role        Role       `json:"role,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Description string     `json:"description,omitempty"`
}

// UpdateAPIKey updates an API key
func (aks *APIKeyService) UpdateAPIKey(ctx context.Context, request *UpdateAPIKeyRequest) (*APIKeyResponse, error) {
	// Get the API key
	apiKey, err := aks.db.GetAPIKeyByID(ctx, request.APIKeyID)
	if err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to get API key for update")
		return nil, fmt.Errorf("API key not found: %w", err)
	}

	// Check if user has permission to update this API key
	if apiKey.UserID != request.UserID {
		// TODO: Add admin permission check
		aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
			"api_key_user_id": apiKey.UserID,
			"request_user_id": request.UserID,
		}).Warn("User attempted to update API key they don't own")
		return nil, fmt.Errorf("insufficient permissions to update this API key")
	}

	// Update fields if provided
	if request.Name != "" {
		apiKey.Name = request.Name
	}
	if request.Role != "" {
		apiKey.Role = string(request.Role)
	}
	if request.Permissions != nil {
		apiKey.Permissions = strings.Join(request.Permissions, ",")
	}
	if request.ExpiresAt != nil {
		apiKey.ExpiresAt = request.ExpiresAt
	}
	apiKey.UpdatedAt = time.Now()

	// Update in database
	if err := aks.db.UpdateAPIKey(ctx, apiKey); err != nil {
		aks.logger.WithComponent("api_key_service").WithError(err).Error("Failed to update API key in database")
		return nil, fmt.Errorf("failed to update API key: %w", err)
	}

	// Parse permissions for response
	permissions := []string{}
	if apiKey.Permissions != "" {
		permissions = strings.Split(apiKey.Permissions, ",")
	}

	response := &APIKeyResponse{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		UserID:      apiKey.UserID,
		Role:        Role(apiKey.Role),
		Permissions: permissions,
		ExpiresAt:   apiKey.ExpiresAt,
		LastUsedAt:  apiKey.LastUsedAt,
		IsActive:    apiKey.Status == "active",
		CreatedAt:   apiKey.CreatedAt,
		UpdatedAt:   apiKey.UpdatedAt,
	}

	aks.logger.WithComponent("api_key_service").WithFields(map[string]interface{}{
		"api_key_id": request.APIKeyID,
		"user_id":    request.UserID,
	}).Info("API key updated successfully")

	return response, nil
}

// validateCreateRequest validates the API key creation request
func (aks *APIKeyService) validateCreateRequest(request *CreateAPIKeyRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}
	if request.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if !IsValidRole(request.Role) {
		return fmt.Errorf("invalid role: %s", request.Role)
	}

	// Validate permissions if provided
	for _, permission := range request.Permissions {
		if !IsValidPermission(Permission(permission)) {
			return fmt.Errorf("invalid permission: %s", permission)
		}
	}

	return nil
}

// generateAPIKey generates a new API key and its hash
func (aks *APIKeyService) generateAPIKey() (string, string, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Create the API key (format: kyb_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx)
	key := fmt.Sprintf("kyb_%s", hex.EncodeToString(bytes))

	// Hash the key for storage
	hash := aks.hashAPIKey(key)

	return key, hash, nil
}

// hashAPIKey creates a SHA-256 hash of the API key
func (aks *APIKeyService) hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
