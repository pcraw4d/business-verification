package tenant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// TenantService handles tenant management operations
type TenantService struct {
	repository TenantRepository
	logger     *zap.Logger
}

// TenantRepository defines the interface for tenant data persistence
type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant *Tenant) error
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error)
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	DeleteTenant(ctx context.Context, tenantID string) error
	ListTenants(ctx context.Context, limit, offset int) ([]*Tenant, error)
	GetTenantUsers(ctx context.Context, tenantID string) ([]*TenantUser, error)
	CreateTenantUser(ctx context.Context, user *TenantUser) error
	GetTenantUser(ctx context.Context, tenantID, userID string) (*TenantUser, error)
	UpdateTenantUser(ctx context.Context, user *TenantUser) error
	DeleteTenantUser(ctx context.Context, tenantID, userID string) error
	CreateTenantAPIKey(ctx context.Context, apiKey *TenantAPIKey) error
	GetTenantAPIKey(ctx context.Context, tenantID, keyID string) (*TenantAPIKey, error)
	GetTenantAPIKeyByHash(ctx context.Context, keyHash string) (*TenantAPIKey, error)
	UpdateTenantAPIKey(ctx context.Context, apiKey *TenantAPIKey) error
	DeleteTenantAPIKey(ctx context.Context, tenantID, keyID string) error
	ListTenantAPIKeys(ctx context.Context, tenantID string) ([]*TenantAPIKey, error)
	GetTenantConfiguration(ctx context.Context, tenantID, category, key string) (*TenantConfiguration, error)
	SetTenantConfiguration(ctx context.Context, config *TenantConfiguration) error
	GetTenantUsage(ctx context.Context, tenantID string, period string) (*TenantUsage, error)
	UpdateTenantUsage(ctx context.Context, usage *TenantUsage) error
	GetTenantMetrics(ctx context.Context, tenantID string) (*TenantMetrics, error)
	LogTenantEvent(ctx context.Context, event *TenantEvent) error
}

// NewTenantService creates a new tenant service
func NewTenantService(repository TenantRepository, logger *zap.Logger) *TenantService {
	return &TenantService{
		repository: repository,
		logger:     logger,
	}
}

// CreateTenant creates a new tenant
func (ts *TenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*Tenant, error) {
	// Check if domain is already taken
	if req.Domain != "" {
		existing, err := ts.repository.GetTenantByDomain(ctx, req.Domain)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("domain %s is already taken", req.Domain)
		}
	}

	// Create tenant
	tenant := &Tenant{
		ID:            generateTenantID(),
		Name:          req.Name,
		Domain:        req.Domain,
		Status:        TenantStatusPending,
		Plan:          req.Plan,
		Configuration: req.Configuration,
		Quotas:        DefaultQuotasByPlan[req.Plan],
		Features:      getFeaturesByPlan(req.Plan),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      req.Metadata,
	}

	// Set subscription end date if provided
	if req.SubscriptionEndsAt != nil {
		tenant.SubscriptionEndsAt = req.SubscriptionEndsAt
	}

	// Create tenant in repository
	if err := ts.repository.CreateTenant(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create owner user if provided
	if req.OwnerEmail != "" {
		owner := &TenantUser{
			ID:          generateUserID(),
			TenantID:    tenant.ID,
			UserID:      req.OwnerUserID,
			Email:       req.OwnerEmail,
			Role:        TenantUserRoleOwner,
			Permissions: DefaultPermissionsByRole[TenantUserRoleOwner],
			Status:      TenantUserStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Metadata:    make(map[string]interface{}),
		}

		if err := ts.repository.CreateTenantUser(ctx, owner); err != nil {
			ts.logger.Error("Failed to create owner user", zap.Error(err))
			// Don't fail tenant creation if owner creation fails
		}
	}

	// Log tenant creation event
	event := &TenantEvent{
		ID:        generateEventID(),
		TenantID:  tenant.ID,
		EventType: TenantEventTypeCreated,
		EventData: map[string]interface{}{
			"tenant_name": tenant.Name,
			"domain":      tenant.Domain,
			"plan":        tenant.Plan,
		},
		UserID:    req.CreatedBy,
		CreatedAt: time.Now(),
	}
	ts.repository.LogTenantEvent(ctx, event)

	ts.logger.Info("Tenant created successfully",
		zap.String("tenant_id", tenant.ID),
		zap.String("tenant_name", tenant.Name),
		zap.String("plan", string(tenant.Plan)))

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (ts *TenantService) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	tenant, err := ts.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Update last activity
	now := time.Now()
	tenant.LastActivityAt = &now
	ts.repository.UpdateTenant(ctx, tenant)

	return tenant, nil
}

// UpdateTenant updates a tenant
func (ts *TenantService) UpdateTenant(ctx context.Context, tenantID string, req *UpdateTenantRequest) (*Tenant, error) {
	tenant, err := ts.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Update fields
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Domain != "" {
		// Check if new domain is available
		existing, err := ts.repository.GetTenantByDomain(ctx, req.Domain)
		if err == nil && existing != nil && existing.ID != tenantID {
			return nil, fmt.Errorf("domain %s is already taken", req.Domain)
		}
		tenant.Domain = req.Domain
	}
	if req.Status != "" {
		tenant.Status = req.Status
	}
	if req.Plan != "" {
		oldPlan := tenant.Plan
		tenant.Plan = req.Plan
		tenant.Quotas = DefaultQuotasByPlan[req.Plan]
		tenant.Features = getFeaturesByPlan(req.Plan)

		// Log plan change event
		event := &TenantEvent{
			ID:        generateEventID(),
			TenantID:  tenant.ID,
			EventType: TenantEventTypePlanChanged,
			EventData: map[string]interface{}{
				"old_plan": oldPlan,
				"new_plan": req.Plan,
			},
			UserID:    req.UpdatedBy,
			CreatedAt: time.Now(),
		}
		ts.repository.LogTenantEvent(ctx, event)
	}
	if req.Configuration != nil {
		tenant.Configuration = req.Configuration
	}
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}

	tenant.UpdatedAt = time.Now()

	if err := ts.repository.UpdateTenant(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Log tenant update event
	event := &TenantEvent{
		ID:        generateEventID(),
		TenantID:  tenant.ID,
		EventType: TenantEventTypeUpdated,
		EventData: map[string]interface{}{
			"updated_fields": req,
		},
		UserID:    req.UpdatedBy,
		CreatedAt: time.Now(),
	}
	ts.repository.LogTenantEvent(ctx, event)

	return tenant, nil
}

// CreateTenantUser creates a new user for a tenant
func (ts *TenantService) CreateTenantUser(ctx context.Context, tenantID string, req *CreateTenantUserRequest) (*TenantUser, error) {
	// Verify tenant exists
	tenant, err := ts.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Check user quota
	users, err := ts.repository.GetTenantUsers(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant users: %w", err)
	}

	if tenant.Quotas.MaxUsers > 0 && len(users) >= tenant.Quotas.MaxUsers {
		return nil, fmt.Errorf("user quota exceeded: %d/%d users", len(users), tenant.Quotas.MaxUsers)
	}

	// Create user
	user := &TenantUser{
		ID:          generateUserID(),
		TenantID:    tenantID,
		UserID:      req.UserID,
		Email:       req.Email,
		Role:        req.Role,
		Permissions: DefaultPermissionsByRole[req.Role],
		Status:      TenantUserStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}

	if err := ts.repository.CreateTenantUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create tenant user: %w", err)
	}

	// Log user addition event
	event := &TenantEvent{
		ID:        generateEventID(),
		TenantID:  tenantID,
		EventType: TenantEventTypeUserAdded,
		EventData: map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
			"role":    user.Role,
		},
		UserID:    req.CreatedBy,
		CreatedAt: time.Now(),
	}
	ts.repository.LogTenantEvent(ctx, event)

	return user, nil
}

// CreateTenantAPIKey creates a new API key for a tenant
func (ts *TenantService) CreateTenantAPIKey(ctx context.Context, tenantID string, req *CreateTenantAPIKeyRequest) (*TenantAPIKey, *string, error) {
	// Generate API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to generate API key: %w", err)
	}
	apiKey := hex.EncodeToString(keyBytes)

	// Hash the key for storage
	keyHash := hashAPIKey(apiKey)

	// Create API key record
	apiKeyRecord := &TenantAPIKey{
		ID:          generateAPIKeyID(),
		TenantID:    tenantID,
		Name:        req.Name,
		KeyHash:     keyHash,
		Permissions: req.Permissions,
		RateLimit:   req.RateLimit,
		Status:      APIKeyStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}

	if req.ExpiresAt != nil {
		apiKeyRecord.ExpiresAt = req.ExpiresAt
	}

	if err := ts.repository.CreateTenantAPIKey(ctx, apiKeyRecord); err != nil {
		return nil, nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Log API key creation event
	event := &TenantEvent{
		ID:        generateEventID(),
		TenantID:  tenantID,
		EventType: TenantEventTypeAPIKeyCreated,
		EventData: map[string]interface{}{
			"api_key_id": apiKeyRecord.ID,
			"name":       apiKeyRecord.Name,
		},
		UserID:    req.CreatedBy,
		CreatedAt: time.Now(),
	}
	ts.repository.LogTenantEvent(ctx, event)

	return apiKeyRecord, &apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated tenant context
func (ts *TenantService) ValidateAPIKey(ctx context.Context, apiKey string) (*TenantContext, error) {
	keyHash := hashAPIKey(apiKey)

	apiKeyRecord, err := ts.repository.GetTenantAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// Check if API key is active
	if apiKeyRecord.Status != APIKeyStatusActive {
		return nil, fmt.Errorf("API key is not active")
	}

	// Check if API key is expired
	if apiKeyRecord.ExpiresAt != nil && apiKeyRecord.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Get tenant
	tenant, err := ts.repository.GetTenant(ctx, apiKeyRecord.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Check if tenant is active
	if tenant.Status != TenantStatusActive {
		return nil, fmt.Errorf("tenant is not active")
	}

	// Update last used timestamp
	now := time.Now()
	apiKeyRecord.LastUsedAt = &now
	ts.repository.UpdateTenantAPIKey(ctx, apiKeyRecord)

	// Create tenant context
	tenantCtx := &TenantContext{
		TenantID:    tenant.ID,
		UserID:      "", // API key doesn't have a specific user
		UserRole:    TenantUserRoleAPI,
		Permissions: apiKeyRecord.Permissions,
		APIKeyID:    apiKeyRecord.ID,
		Metadata: map[string]interface{}{
			"api_key_name": apiKeyRecord.Name,
			"rate_limit":   apiKeyRecord.RateLimit,
		},
	}

	return tenantCtx, nil
}

// GetTenantMetrics retrieves metrics for a tenant
func (ts *TenantService) GetTenantMetrics(ctx context.Context, tenantID string) (*TenantMetrics, error) {
	metrics, err := ts.repository.GetTenantMetrics(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant metrics: %w", err)
	}

	return metrics, nil
}

// CheckQuota checks if a tenant has exceeded their quota
func (ts *TenantService) CheckQuota(ctx context.Context, tenantID string, quotaType string, currentUsage int64) error {
	tenant, err := ts.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	var quota int64
	switch quotaType {
	case "assessments_per_month":
		quota = tenant.Quotas.MaxAssessmentsPerMonth
	case "api_requests_per_day":
		quota = tenant.Quotas.MaxAPIRequestsPerDay
	case "concurrent_requests":
		quota = int64(tenant.Quotas.MaxConcurrentRequests)
	default:
		return fmt.Errorf("unknown quota type: %s", quotaType)
	}

	if quota > 0 && currentUsage >= quota {
		// Log quota exceeded event
		event := &TenantEvent{
			ID:        generateEventID(),
			TenantID:  tenantID,
			EventType: TenantEventTypeQuotaExceeded,
			EventData: map[string]interface{}{
				"quota_type":    quotaType,
				"current_usage": currentUsage,
				"quota_limit":   quota,
			},
			CreatedAt: time.Now(),
		}
		ts.repository.LogTenantEvent(ctx, event)

		return fmt.Errorf("quota exceeded for %s: %d/%d", quotaType, currentUsage, quota)
	}

	return nil
}

// Request structures
type CreateTenantRequest struct {
	Name               string                 `json:"name"`
	Domain             string                 `json:"domain"`
	Plan               TenantPlan             `json:"plan"`
	Configuration      map[string]interface{} `json:"configuration"`
	OwnerEmail         string                 `json:"owner_email"`
	OwnerUserID        string                 `json:"owner_user_id"`
	SubscriptionEndsAt *time.Time             `json:"subscription_ends_at"`
	CreatedBy          string                 `json:"created_by"`
	Metadata           map[string]interface{} `json:"metadata"`
}

type UpdateTenantRequest struct {
	Name          string                 `json:"name"`
	Domain        string                 `json:"domain"`
	Status        TenantStatus           `json:"status"`
	Plan          TenantPlan             `json:"plan"`
	Configuration map[string]interface{} `json:"configuration"`
	UpdatedBy     string                 `json:"updated_by"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type CreateTenantUserRequest struct {
	UserID    string                 `json:"user_id"`
	Email     string                 `json:"email"`
	Role      TenantUserRole         `json:"role"`
	CreatedBy string                 `json:"created_by"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type CreateTenantAPIKeyRequest struct {
	Name        string                 `json:"name"`
	Permissions []string               `json:"permissions"`
	RateLimit   int                    `json:"rate_limit"`
	ExpiresAt   *time.Time             `json:"expires_at"`
	CreatedBy   string                 `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Helper functions
func generateTenantID() string {
	return fmt.Sprintf("tenant_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateUserID() string {
	return fmt.Sprintf("user_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateAPIKeyID() string {
	return fmt.Sprintf("key_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateEventID() string {
	return fmt.Sprintf("event_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func hashAPIKey(apiKey string) string {
	// In production, use a proper hash function like SHA-256
	return fmt.Sprintf("hash_%s", apiKey)
}

func getFeaturesByPlan(plan TenantPlan) []string {
	features := map[TenantPlan][]string{
		TenantPlanFree: {
			"basic_assessments",
			"standard_reports",
		},
		TenantPlanBasic: {
			"basic_assessments",
			"standard_reports",
			"api_access",
			"custom_branding",
		},
		TenantPlanProfessional: {
			"basic_assessments",
			"standard_reports",
			"api_access",
			"custom_branding",
			"advanced_analytics",
			"compliance_reporting",
			"priority_support",
		},
		TenantPlanEnterprise: {
			"basic_assessments",
			"standard_reports",
			"api_access",
			"custom_branding",
			"advanced_analytics",
			"compliance_reporting",
			"priority_support",
			"custom_integrations",
			"dedicated_support",
			"sla_guarantee",
		},
	}

	if planFeatures, exists := features[plan]; exists {
		return planFeatures
	}
	return []string{}
}
