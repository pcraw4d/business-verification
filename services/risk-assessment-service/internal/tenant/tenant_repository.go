package tenant

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// TenantRepositoryImpl implements the TenantRepository interface
type TenantRepositoryImpl struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *sql.DB, logger *zap.Logger) TenantRepository {
	return &TenantRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// CreateTenant creates a new tenant
func (r *TenantRepositoryImpl) CreateTenant(ctx context.Context, tenant *Tenant) error {
	query := `
		INSERT INTO tenants (
			id, name, domain, status, plan, configuration, quotas, features,
			created_at, updated_at, last_activity_at, subscription_ends_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	configurationJSON, err := json.Marshal(tenant.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	quotasJSON, err := json.Marshal(tenant.Quotas)
	if err != nil {
		return fmt.Errorf("failed to marshal quotas: %w", err)
	}

	featuresJSON, err := json.Marshal(tenant.Features)
	if err != nil {
		return fmt.Errorf("failed to marshal features: %w", err)
	}

	metadataJSON, err := json.Marshal(tenant.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Domain, tenant.Status, tenant.Plan,
		configurationJSON, quotasJSON, featuresJSON,
		tenant.CreatedAt, tenant.UpdatedAt, tenant.LastActivityAt,
		tenant.SubscriptionEndsAt, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetTenant retrieves a tenant by ID
func (r *TenantRepositoryImpl) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	query := `
		SELECT id, name, domain, status, plan, configuration, quotas, features,
			   created_at, updated_at, last_activity_at, subscription_ends_at, metadata
		FROM tenants
		WHERE id = $1
	`

	var tenant Tenant
	var configurationJSON, quotasJSON, featuresJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.Status, &tenant.Plan,
		&configurationJSON, &quotasJSON, &featuresJSON,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.LastActivityAt,
		&tenant.SubscriptionEndsAt, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(configurationJSON, &tenant.Configuration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := json.Unmarshal(quotasJSON, &tenant.Quotas); err != nil {
		return nil, fmt.Errorf("failed to unmarshal quotas: %w", err)
	}

	if err := json.Unmarshal(featuresJSON, &tenant.Features); err != nil {
		return nil, fmt.Errorf("failed to unmarshal features: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &tenant.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain
func (r *TenantRepositoryImpl) GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	query := `
		SELECT id, name, domain, status, plan, configuration, quotas, features,
			   created_at, updated_at, last_activity_at, subscription_ends_at, metadata
		FROM tenants
		WHERE domain = $1
	`

	var tenant Tenant
	var configurationJSON, quotasJSON, featuresJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, domain).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.Status, &tenant.Plan,
		&configurationJSON, &quotasJSON, &featuresJSON,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.LastActivityAt,
		&tenant.SubscriptionEndsAt, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found for domain: %s", domain)
		}
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(configurationJSON, &tenant.Configuration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := json.Unmarshal(quotasJSON, &tenant.Quotas); err != nil {
		return nil, fmt.Errorf("failed to unmarshal quotas: %w", err)
	}

	if err := json.Unmarshal(featuresJSON, &tenant.Features); err != nil {
		return nil, fmt.Errorf("failed to unmarshal features: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &tenant.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &tenant, nil
}

// UpdateTenant updates a tenant
func (r *TenantRepositoryImpl) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	query := `
		UPDATE tenants SET
			name = $2, domain = $3, status = $4, plan = $5, configuration = $6,
			quotas = $7, features = $8, updated_at = $9, last_activity_at = $10,
			subscription_ends_at = $11, metadata = $12
		WHERE id = $1
	`

	configurationJSON, err := json.Marshal(tenant.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	quotasJSON, err := json.Marshal(tenant.Quotas)
	if err != nil {
		return fmt.Errorf("failed to marshal quotas: %w", err)
	}

	featuresJSON, err := json.Marshal(tenant.Features)
	if err != nil {
		return fmt.Errorf("failed to marshal features: %w", err)
	}

	metadataJSON, err := json.Marshal(tenant.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Domain, tenant.Status, tenant.Plan,
		configurationJSON, quotasJSON, featuresJSON,
		tenant.UpdatedAt, tenant.LastActivityAt, tenant.SubscriptionEndsAt, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

// DeleteTenant deletes a tenant
func (r *TenantRepositoryImpl) DeleteTenant(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenants WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	return nil
}

// ListTenants lists tenants with pagination
func (r *TenantRepositoryImpl) ListTenants(ctx context.Context, limit, offset int) ([]*Tenant, error) {
	query := `
		SELECT id, name, domain, status, plan, configuration, quotas, features,
			   created_at, updated_at, last_activity_at, subscription_ends_at, metadata
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var tenant Tenant
		var configurationJSON, quotasJSON, featuresJSON, metadataJSON []byte

		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.Status, &tenant.Plan,
			&configurationJSON, &quotasJSON, &featuresJSON,
			&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.LastActivityAt,
			&tenant.SubscriptionEndsAt, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(configurationJSON, &tenant.Configuration); err != nil {
			return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
		}

		if err := json.Unmarshal(quotasJSON, &tenant.Quotas); err != nil {
			return nil, fmt.Errorf("failed to unmarshal quotas: %w", err)
		}

		if err := json.Unmarshal(featuresJSON, &tenant.Features); err != nil {
			return nil, fmt.Errorf("failed to unmarshal features: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &tenant.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		tenants = append(tenants, &tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenants: %w", err)
	}

	return tenants, nil
}

// CreateTenantUser creates a new tenant user
func (r *TenantRepositoryImpl) CreateTenantUser(ctx context.Context, user *TenantUser) error {
	query := `
		INSERT INTO tenant_users (
			id, tenant_id, user_id, email, role, permissions, status,
			last_login_at, created_at, updated_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	permissionsJSON, err := json.Marshal(user.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	metadataJSON, err := json.Marshal(user.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		user.ID, user.TenantID, user.UserID, user.Email, user.Role,
		permissionsJSON, user.Status, user.LastLoginAt,
		user.CreatedAt, user.UpdatedAt, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create tenant user: %w", err)
	}

	return nil
}

// GetTenantUser retrieves a tenant user
func (r *TenantRepositoryImpl) GetTenantUser(ctx context.Context, tenantID, userID string) (*TenantUser, error) {
	query := `
		SELECT id, tenant_id, user_id, email, role, permissions, status,
			   last_login_at, created_at, updated_at, metadata
		FROM tenant_users
		WHERE tenant_id = $1 AND user_id = $2
	`

	var user TenantUser
	var permissionsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, userID).Scan(
		&user.ID, &user.TenantID, &user.UserID, &user.Email, &user.Role,
		&permissionsJSON, &user.Status, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant user not found: %s/%s", tenantID, userID)
		}
		return nil, fmt.Errorf("failed to get tenant user: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(permissionsJSON, &user.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &user, nil
}

// GetTenantUsers retrieves all users for a tenant
func (r *TenantRepositoryImpl) GetTenantUsers(ctx context.Context, tenantID string) ([]*TenantUser, error) {
	query := `
		SELECT id, tenant_id, user_id, email, role, permissions, status,
			   last_login_at, created_at, updated_at, metadata
		FROM tenant_users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant users: %w", err)
	}
	defer rows.Close()

	var users []*TenantUser
	for rows.Next() {
		var user TenantUser
		var permissionsJSON, metadataJSON []byte

		err := rows.Scan(
			&user.ID, &user.TenantID, &user.UserID, &user.Email, &user.Role,
			&permissionsJSON, &user.Status, &user.LastLoginAt,
			&user.CreatedAt, &user.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant user: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(permissionsJSON, &user.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenant users: %w", err)
	}

	return users, nil
}

// UpdateTenantUser updates a tenant user
func (r *TenantRepositoryImpl) UpdateTenantUser(ctx context.Context, user *TenantUser) error {
	query := `
		UPDATE tenant_users SET
			email = $3, role = $4, permissions = $5, status = $6,
			last_login_at = $7, updated_at = $8, metadata = $9
		WHERE tenant_id = $1 AND user_id = $2
	`

	permissionsJSON, err := json.Marshal(user.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	metadataJSON, err := json.Marshal(user.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		user.TenantID, user.UserID, user.Email, user.Role,
		permissionsJSON, user.Status, user.LastLoginAt,
		user.UpdatedAt, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant user: %w", err)
	}

	return nil
}

// DeleteTenantUser deletes a tenant user
func (r *TenantRepositoryImpl) DeleteTenantUser(ctx context.Context, tenantID, userID string) error {
	query := `DELETE FROM tenant_users WHERE tenant_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant user not found: %s/%s", tenantID, userID)
	}

	return nil
}

// CreateTenantAPIKey creates a new tenant API key
func (r *TenantRepositoryImpl) CreateTenantAPIKey(ctx context.Context, apiKey *TenantAPIKey) error {
	query := `
		INSERT INTO tenant_api_keys (
			id, tenant_id, name, key_hash, permissions, rate_limit, status,
			last_used_at, expires_at, created_at, updated_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	metadataJSON, err := json.Marshal(apiKey.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		apiKey.ID, apiKey.TenantID, apiKey.Name, apiKey.KeyHash,
		permissionsJSON, apiKey.RateLimit, apiKey.Status,
		apiKey.LastUsedAt, apiKey.ExpiresAt,
		apiKey.CreatedAt, apiKey.UpdatedAt, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create tenant API key: %w", err)
	}

	return nil
}

// GetTenantAPIKey retrieves a tenant API key
func (r *TenantRepositoryImpl) GetTenantAPIKey(ctx context.Context, tenantID, keyID string) (*TenantAPIKey, error) {
	query := `
		SELECT id, tenant_id, name, key_hash, permissions, rate_limit, status,
			   last_used_at, expires_at, created_at, updated_at, metadata
		FROM tenant_api_keys
		WHERE tenant_id = $1 AND id = $2
	`

	var apiKey TenantAPIKey
	var permissionsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, keyID).Scan(
		&apiKey.ID, &apiKey.TenantID, &apiKey.Name, &apiKey.KeyHash,
		&permissionsJSON, &apiKey.RateLimit, &apiKey.Status,
		&apiKey.LastUsedAt, &apiKey.ExpiresAt,
		&apiKey.CreatedAt, &apiKey.UpdatedAt, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant API key not found: %s/%s", tenantID, keyID)
		}
		return nil, fmt.Errorf("failed to get tenant API key: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(permissionsJSON, &apiKey.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &apiKey.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &apiKey, nil
}

// GetTenantAPIKeyByHash retrieves a tenant API key by hash
func (r *TenantRepositoryImpl) GetTenantAPIKeyByHash(ctx context.Context, keyHash string) (*TenantAPIKey, error) {
	query := `
		SELECT id, tenant_id, name, key_hash, permissions, rate_limit, status,
			   last_used_at, expires_at, created_at, updated_at, metadata
		FROM tenant_api_keys
		WHERE key_hash = $1
	`

	var apiKey TenantAPIKey
	var permissionsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, keyHash).Scan(
		&apiKey.ID, &apiKey.TenantID, &apiKey.Name, &apiKey.KeyHash,
		&permissionsJSON, &apiKey.RateLimit, &apiKey.Status,
		&apiKey.LastUsedAt, &apiKey.ExpiresAt,
		&apiKey.CreatedAt, &apiKey.UpdatedAt, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant API key not found for hash")
		}
		return nil, fmt.Errorf("failed to get tenant API key by hash: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(permissionsJSON, &apiKey.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &apiKey.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &apiKey, nil
}

// UpdateTenantAPIKey updates a tenant API key
func (r *TenantRepositoryImpl) UpdateTenantAPIKey(ctx context.Context, apiKey *TenantAPIKey) error {
	query := `
		UPDATE tenant_api_keys SET
			name = $3, permissions = $4, rate_limit = $5, status = $6,
			last_used_at = $7, expires_at = $8, updated_at = $9, metadata = $10
		WHERE tenant_id = $1 AND id = $2
	`

	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	metadataJSON, err := json.Marshal(apiKey.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		apiKey.TenantID, apiKey.ID, apiKey.Name,
		permissionsJSON, apiKey.RateLimit, apiKey.Status,
		apiKey.LastUsedAt, apiKey.ExpiresAt,
		apiKey.UpdatedAt, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant API key: %w", err)
	}

	return nil
}

// DeleteTenantAPIKey deletes a tenant API key
func (r *TenantRepositoryImpl) DeleteTenantAPIKey(ctx context.Context, tenantID, keyID string) error {
	query := `DELETE FROM tenant_api_keys WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.ExecContext(ctx, query, tenantID, keyID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant API key not found: %s/%s", tenantID, keyID)
	}

	return nil
}

// ListTenantAPIKeys lists API keys for a tenant
func (r *TenantRepositoryImpl) ListTenantAPIKeys(ctx context.Context, tenantID string) ([]*TenantAPIKey, error) {
	query := `
		SELECT id, tenant_id, name, key_hash, permissions, rate_limit, status,
			   last_used_at, expires_at, created_at, updated_at, metadata
		FROM tenant_api_keys
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*TenantAPIKey
	for rows.Next() {
		var apiKey TenantAPIKey
		var permissionsJSON, metadataJSON []byte

		err := rows.Scan(
			&apiKey.ID, &apiKey.TenantID, &apiKey.Name, &apiKey.KeyHash,
			&permissionsJSON, &apiKey.RateLimit, &apiKey.Status,
			&apiKey.LastUsedAt, &apiKey.ExpiresAt,
			&apiKey.CreatedAt, &apiKey.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant API key: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(permissionsJSON, &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &apiKey.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		apiKeys = append(apiKeys, &apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenant API keys: %w", err)
	}

	return apiKeys, nil
}

// GetTenantConfiguration retrieves tenant configuration
func (r *TenantRepositoryImpl) GetTenantConfiguration(ctx context.Context, tenantID, category, key string) (*TenantConfiguration, error) {
	query := `
		SELECT id, tenant_id, category, key, value, value_type, description,
			   is_encrypted, created_at, updated_at, updated_by, metadata
		FROM tenant_configurations
		WHERE tenant_id = $1 AND category = $2 AND key = $3
	`

	var config TenantConfiguration
	var valueJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, category, key).Scan(
		&config.ID, &config.TenantID, &config.Category, &config.Key,
		&valueJSON, &config.ValueType, &config.Description,
		&config.IsEncrypted, &config.CreatedAt, &config.UpdatedAt,
		&config.UpdatedBy, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant configuration not found: %s/%s/%s", tenantID, category, key)
		}
		return nil, fmt.Errorf("failed to get tenant configuration: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(valueJSON, &config.Value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &config.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &config, nil
}

// SetTenantConfiguration sets tenant configuration
func (r *TenantRepositoryImpl) SetTenantConfiguration(ctx context.Context, config *TenantConfiguration) error {
	query := `
		INSERT INTO tenant_configurations (
			id, tenant_id, category, key, value, value_type, description,
			is_encrypted, created_at, updated_at, updated_by, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		ON CONFLICT (tenant_id, category, key) DO UPDATE SET
			value = EXCLUDED.value, value_type = EXCLUDED.value_type,
			description = EXCLUDED.description, is_encrypted = EXCLUDED.is_encrypted,
			updated_at = EXCLUDED.updated_at, updated_by = EXCLUDED.updated_by,
			metadata = EXCLUDED.metadata
	`

	valueJSON, err := json.Marshal(config.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	metadataJSON, err := json.Marshal(config.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		config.ID, config.TenantID, config.Category, config.Key,
		valueJSON, config.ValueType, config.Description,
		config.IsEncrypted, config.CreatedAt, config.UpdatedAt,
		config.UpdatedBy, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to set tenant configuration: %w", err)
	}

	return nil
}

// GetTenantUsage retrieves tenant usage statistics
func (r *TenantRepositoryImpl) GetTenantUsage(ctx context.Context, tenantID string, period string) (*TenantUsage, error) {
	query := `
		SELECT id, tenant_id, period, assessments_count, api_requests_count,
			   users_count, data_storage_bytes, audit_logs_count,
			   compliance_reports_count, created_at, updated_at
		FROM tenant_usage
		WHERE tenant_id = $1 AND period = $2
	`

	var usage TenantUsage
	err := r.db.QueryRowContext(ctx, query, tenantID, period).Scan(
		&usage.ID, &usage.TenantID, &usage.Period, &usage.AssessmentsCount,
		&usage.APIRequestsCount, &usage.UsersCount, &usage.DataStorageBytes,
		&usage.AuditLogsCount, &usage.ComplianceReportsCount,
		&usage.CreatedAt, &usage.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant usage not found: %s/%s", tenantID, period)
		}
		return nil, fmt.Errorf("failed to get tenant usage: %w", err)
	}

	return &usage, nil
}

// UpdateTenantUsage updates tenant usage statistics
func (r *TenantRepositoryImpl) UpdateTenantUsage(ctx context.Context, usage *TenantUsage) error {
	query := `
		INSERT INTO tenant_usage (
			id, tenant_id, period, assessments_count, api_requests_count,
			users_count, data_storage_bytes, audit_logs_count,
			compliance_reports_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (tenant_id, period) DO UPDATE SET
			assessments_count = EXCLUDED.assessments_count,
			api_requests_count = EXCLUDED.api_requests_count,
			users_count = EXCLUDED.users_count,
			data_storage_bytes = EXCLUDED.data_storage_bytes,
			audit_logs_count = EXCLUDED.audit_logs_count,
			compliance_reports_count = EXCLUDED.compliance_reports_count,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		usage.ID, usage.TenantID, usage.Period, usage.AssessmentsCount,
		usage.APIRequestsCount, usage.UsersCount, usage.DataStorageBytes,
		usage.AuditLogsCount, usage.ComplianceReportsCount,
		usage.CreatedAt, usage.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant usage: %w", err)
	}

	return nil
}

// GetTenantMetrics retrieves tenant metrics
func (r *TenantRepositoryImpl) GetTenantMetrics(ctx context.Context, tenantID string) (*TenantMetrics, error) {
	// This would typically aggregate data from multiple tables
	// For now, we'll return a mock implementation
	metrics := &TenantMetrics{
		TenantID:            tenantID,
		ActiveUsers:         5,
		TotalAssessments:    1250,
		APIRequestsToday:    450,
		StorageUsed:         1024 * 1024 * 100, // 100MB
		QuotaUtilization:    map[string]float64{"assessments": 0.75, "api_requests": 0.45},
		LastActivityAt:      &[]time.Time{time.Now()}[0],
		HealthScore:         0.95,
		ComplianceScore:     0.92,
		PerformanceMetrics:  map[string]interface{}{"avg_response_time": 150, "error_rate": 0.02},
		ErrorRate:           0.02,
		AverageResponseTime: 150,
	}

	return metrics, nil
}

// LogTenantEvent logs a tenant event
func (r *TenantRepositoryImpl) LogTenantEvent(ctx context.Context, event *TenantEvent) error {
	query := `
		INSERT INTO tenant_events (
			id, tenant_id, event_type, event_data, user_id, ip_address, user_agent, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	eventDataJSON, err := json.Marshal(event.EventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.EventType, eventDataJSON,
		event.UserID, event.IPAddress, event.UserAgent, event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to log tenant event: %w", err)
	}

	return nil
}
