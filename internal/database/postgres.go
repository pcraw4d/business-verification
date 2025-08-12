package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"encoding/json"

	_ "github.com/lib/pq"
)

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	db     *sql.DB
	config *DatabaseConfig
	tx     *sql.Tx
}

// NewPostgresDB creates a new PostgreSQL database instance
func NewPostgresDB(cfg *DatabaseConfig) *PostgresDB {
	return &PostgresDB{
		config: cfg,
	}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresDB) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.Username,
		p.config.Password,
		p.config.Database,
		p.config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(p.config.MaxOpenConns)
	db.SetMaxIdleConns(p.config.MaxIdleConns)
	db.SetConnMaxLifetime(p.config.ConnMaxLifetime)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db
	return nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping tests the database connection
func (p *PostgresDB) Ping(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("database not connected")
	}
	return p.db.PingContext(ctx)
}

// BeginTx starts a new transaction
func (p *PostgresDB) BeginTx(ctx context.Context) (Database, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PostgresDB{
		db:     p.db,
		config: p.config,
		tx:     tx,
	}, nil
}

// Commit commits the current transaction
func (p *PostgresDB) Commit() error {
	if p.tx == nil {
		return fmt.Errorf("no active transaction")
	}
	return p.tx.Commit()
}

// Rollback rolls back the current transaction
func (p *PostgresDB) Rollback() error {
	if p.tx == nil {
		return fmt.Errorf("no active transaction")
	}
	return p.tx.Rollback()
}

// getDB returns the appropriate database connection (transaction or regular)
func (p *PostgresDB) getDB() interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
} {
	if p.tx != nil {
		return p.tx
	}
	return p.db
}

// GetDB returns the underlying *sql.DB for migration purposes
func (p *PostgresDB) GetDB() *sql.DB {
	return p.db
}

// CreateUser creates a new user
func (p *PostgresDB) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, first_name, last_name, company, role, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := p.getDB().ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.FirstName, user.LastName, user.Company, user.Role, user.Status)

	return err
}

// GetUserByID retrieves a user by ID
func (p *PostgresDB) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT * FROM users WHERE id = $1`

	row := p.getDB().QueryRowContext(ctx, query, id)

	var user User
	err := row.Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Company, &user.Role, &user.Status,
		&user.EmailVerified, &user.LastLoginAt, &user.FailedLoginAttempts,
		&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (p *PostgresDB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT * FROM users WHERE email = $1`

	row := p.getDB().QueryRowContext(ctx, query, email)

	var user User
	err := row.Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Company, &user.Role, &user.Status,
		&user.EmailVerified, &user.LastLoginAt, &user.FailedLoginAttempts,
		&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (p *PostgresDB) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users 
		SET email = $2, username = $3, password_hash = $4, first_name = $5, 
		    last_name = $6, company = $7, role = $8, status = $9, 
		    email_verified = $10, last_login_at = $11, failed_login_attempts = $12,
		    locked_until = $13, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := p.getDB().ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.FirstName, user.LastName, user.Company, user.Role, user.Status,
		user.EmailVerified, user.LastLoginAt, user.FailedLoginAttempts,
		user.LockedUntil)

	return err
}

// DeleteUser deletes a user by ID
func (p *PostgresDB) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := p.getDB().ExecContext(ctx, query, id)
	return err
}

// ListUsers retrieves a list of users with pagination
func (p *PostgresDB) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := p.getDB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.PasswordHash,
			&user.FirstName, &user.LastName, &user.Company, &user.Role, &user.Status,
			&user.EmailVerified, &user.LastLoginAt, &user.FailedLoginAttempts,
			&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// CreateBusiness creates a new business
func (p *PostgresDB) CreateBusiness(ctx context.Context, business *Business) error {
	query := `
		INSERT INTO businesses (
			id, name, legal_name, registration_number, tax_id, industry, industry_code,
			business_type, founded_date, employee_count, annual_revenue,
			address_street1, address_street2, address_city, address_state,
			address_postal_code, address_country, address_country_code,
			contact_phone, contact_email, contact_website, contact_primary_contact,
			status, risk_level, compliance_status, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		)
	`

	_, err := p.getDB().ExecContext(ctx, query,
		business.ID, business.Name, business.LegalName, business.RegistrationNumber,
		business.TaxID, business.Industry, business.IndustryCode, business.BusinessType,
		business.FoundedDate, business.EmployeeCount, business.AnnualRevenue,
		business.Address.Street1, business.Address.Street2, business.Address.City,
		business.Address.State, business.Address.PostalCode, business.Address.Country,
		business.Address.CountryCode, business.ContactInfo.Phone, business.ContactInfo.Email,
		business.ContactInfo.Website, business.ContactInfo.PrimaryContact,
		business.Status, business.RiskLevel, business.ComplianceStatus, business.CreatedBy)

	return err
}

// GetBusinessByID retrieves a business by ID
func (p *PostgresDB) GetBusinessByID(ctx context.Context, id string) (*Business, error) {
	query := `SELECT * FROM businesses WHERE id = $1`

	row := p.getDB().QueryRowContext(ctx, query, id)

	var business Business
	err := row.Scan(
		&business.ID, &business.Name, &business.LegalName, &business.RegistrationNumber,
		&business.TaxID, &business.Industry, &business.IndustryCode, &business.BusinessType,
		&business.FoundedDate, &business.EmployeeCount, &business.AnnualRevenue,
		&business.Address.Street1, &business.Address.Street2, &business.Address.City,
		&business.Address.State, &business.Address.PostalCode, &business.Address.Country,
		&business.Address.CountryCode, &business.ContactInfo.Phone, &business.ContactInfo.Email,
		&business.ContactInfo.Website, &business.ContactInfo.PrimaryContact,
		&business.Status, &business.RiskLevel, &business.ComplianceStatus, &business.CreatedBy,
		&business.CreatedAt, &business.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &business, nil
}

// GetBusinessByRegistrationNumber retrieves a business by registration number
func (p *PostgresDB) GetBusinessByRegistrationNumber(ctx context.Context, regNumber string) (*Business, error) {
	query := `SELECT * FROM businesses WHERE registration_number = $1`

	row := p.getDB().QueryRowContext(ctx, query, regNumber)

	var business Business
	err := row.Scan(
		&business.ID, &business.Name, &business.LegalName, &business.RegistrationNumber,
		&business.TaxID, &business.Industry, &business.IndustryCode, &business.BusinessType,
		&business.FoundedDate, &business.EmployeeCount, &business.AnnualRevenue,
		&business.Address.Street1, &business.Address.Street2, &business.Address.City,
		&business.Address.State, &business.Address.PostalCode, &business.Address.Country,
		&business.Address.CountryCode, &business.ContactInfo.Phone, &business.ContactInfo.Email,
		&business.ContactInfo.Website, &business.ContactInfo.PrimaryContact,
		&business.Status, &business.RiskLevel, &business.ComplianceStatus, &business.CreatedBy,
		&business.CreatedAt, &business.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &business, nil
}

// UpdateBusiness updates an existing business
func (p *PostgresDB) UpdateBusiness(ctx context.Context, business *Business) error {
	query := `
		UPDATE businesses 
		SET name = $2, legal_name = $3, registration_number = $4, tax_id = $5,
		    industry = $6, industry_code = $7, business_type = $8, founded_date = $9,
		    employee_count = $10, annual_revenue = $11,
		    address_street1 = $12, address_street2 = $13, address_city = $14,
		    address_state = $15, address_postal_code = $16, address_country = $17,
		    address_country_code = $18, contact_phone = $19, contact_email = $20,
		    contact_website = $21, contact_primary_contact = $22,
		    status = $23, risk_level = $24, compliance_status = $25,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := p.getDB().ExecContext(ctx, query,
		business.ID, business.Name, business.LegalName, business.RegistrationNumber,
		business.TaxID, business.Industry, business.IndustryCode, business.BusinessType,
		business.FoundedDate, business.EmployeeCount, business.AnnualRevenue,
		business.Address.Street1, business.Address.Street2, business.Address.City,
		business.Address.State, business.Address.PostalCode, business.Address.Country,
		business.Address.CountryCode, business.ContactInfo.Phone, business.ContactInfo.Email,
		business.ContactInfo.Website, business.ContactInfo.PrimaryContact,
		business.Status, business.RiskLevel, business.ComplianceStatus)

	return err
}

// DeleteBusiness deletes a business by ID
func (p *PostgresDB) DeleteBusiness(ctx context.Context, id string) error {
	query := `DELETE FROM businesses WHERE id = $1`

	_, err := p.getDB().ExecContext(ctx, query, id)
	return err
}

// ListBusinesses retrieves a list of businesses with pagination
func (p *PostgresDB) ListBusinesses(ctx context.Context, limit, offset int) ([]*Business, error) {
	query := `SELECT * FROM businesses ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := p.getDB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []*Business
	for rows.Next() {
		var business Business
		err := rows.Scan(
			&business.ID, &business.Name, &business.LegalName, &business.RegistrationNumber,
			&business.TaxID, &business.Industry, &business.IndustryCode, &business.BusinessType,
			&business.FoundedDate, &business.EmployeeCount, &business.AnnualRevenue,
			&business.Address.Street1, &business.Address.Street2, &business.Address.City,
			&business.Address.State, &business.Address.PostalCode, &business.Address.Country,
			&business.Address.CountryCode, &business.ContactInfo.Phone, &business.ContactInfo.Email,
			&business.ContactInfo.Website, &business.ContactInfo.PrimaryContact,
			&business.Status, &business.RiskLevel, &business.ComplianceStatus, &business.CreatedBy,
			&business.CreatedAt, &business.UpdatedAt)
		if err != nil {
			return nil, err
		}
		businesses = append(businesses, &business)
	}

	return businesses, nil
}

// SearchBusinesses searches businesses by query
func (p *PostgresDB) SearchBusinesses(ctx context.Context, query string, limit, offset int) ([]*Business, error) {
	searchQuery := `
		SELECT * FROM businesses 
		WHERE name ILIKE $1 OR legal_name ILIKE $1 OR registration_number ILIKE $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	searchTerm := "%" + query + "%"
	rows, err := p.getDB().QueryContext(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []*Business
	for rows.Next() {
		var business Business
		err := rows.Scan(
			&business.ID, &business.Name, &business.LegalName, &business.RegistrationNumber,
			&business.TaxID, &business.Industry, &business.IndustryCode, &business.BusinessType,
			&business.FoundedDate, &business.EmployeeCount, &business.AnnualRevenue,
			&business.Address.Street1, &business.Address.Street2, &business.Address.City,
			&business.Address.State, &business.Address.PostalCode, &business.Address.Country,
			&business.Address.CountryCode, &business.ContactInfo.Phone, &business.ContactInfo.Email,
			&business.ContactInfo.Website, &business.ContactInfo.PrimaryContact,
			&business.Status, &business.RiskLevel, &business.ComplianceStatus, &business.CreatedBy,
			&business.CreatedAt, &business.UpdatedAt)
		if err != nil {
			return nil, err
		}
		businesses = append(businesses, &business)
	}

	return businesses, nil
}

// Placeholder implementations for other methods
// These would be implemented similarly to the above methods

func (p *PostgresDB) CreateBusinessClassification(ctx context.Context, classification *BusinessClassification) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetBusinessClassificationByID(ctx context.Context, id string) (*BusinessClassification, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetBusinessClassificationsByBusinessID(ctx context.Context, businessID string) ([]*BusinessClassification, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) UpdateBusinessClassification(ctx context.Context, classification *BusinessClassification) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) DeleteBusinessClassification(ctx context.Context, id string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) CreateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetRiskAssessmentByID(ctx context.Context, id string) (*RiskAssessment, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string) ([]*RiskAssessment, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) UpdateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) DeleteRiskAssessment(ctx context.Context, id string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) CreateComplianceCheck(ctx context.Context, check *ComplianceCheck) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetComplianceCheckByID(ctx context.Context, id string) (*ComplianceCheck, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetComplianceChecksByBusinessID(ctx context.Context, businessID string) ([]*ComplianceCheck, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) UpdateComplianceCheck(ctx context.Context, check *ComplianceCheck) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) DeleteComplianceCheck(ctx context.Context, id string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) CreateAPIKey(ctx context.Context, apiKey *APIKey) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetAPIKeyByID(ctx context.Context, id string) (*APIKey, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) UpdateAPIKey(ctx context.Context, apiKey *APIKey) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) DeleteAPIKey(ctx context.Context, id string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) ListAPIKeysByUserID(ctx context.Context, userID string) ([]*APIKey, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*AuditLog, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetAuditLogsByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*AuditLog, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) CreateExternalServiceCall(ctx context.Context, call *ExternalServiceCall) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetExternalServiceCallsByUserID(ctx context.Context, userID string, limit, offset int) ([]*ExternalServiceCall, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetExternalServiceCallsByService(ctx context.Context, serviceName string, limit, offset int) ([]*ExternalServiceCall, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) CreateWebhook(ctx context.Context, webhook *Webhook) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetWebhookByID(ctx context.Context, id string) (*Webhook, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetWebhooksByUserID(ctx context.Context, userID string) ([]*Webhook, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) UpdateWebhook(ctx context.Context, webhook *Webhook) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) DeleteWebhook(ctx context.Context, id string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) CreateWebhookEvent(ctx context.Context, event *WebhookEvent) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetWebhookEventByID(ctx context.Context, id string) (*WebhookEvent, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) GetWebhookEventsByWebhookID(ctx context.Context, webhookID string, limit, offset int) ([]*WebhookEvent, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (p *PostgresDB) UpdateWebhookEvent(ctx context.Context, event *WebhookEvent) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (p *PostgresDB) DeleteWebhookEvent(ctx context.Context, id string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

// Email verification token methods
func (p *PostgresDB) CreateEmailVerificationToken(ctx context.Context, token *EmailVerificationToken) error {
	query := `
		INSERT INTO email_verification_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := p.getDB().ExecContext(ctx, query,
		token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	return err
}

func (p *PostgresDB) GetEmailVerificationToken(ctx context.Context, token string) (*EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM email_verification_tokens
		WHERE token = $1
	`
	var verificationToken EmailVerificationToken
	err := p.getDB().QueryRowContext(ctx, query, token).Scan(
		&verificationToken.ID, &verificationToken.UserID, &verificationToken.Token,
		&verificationToken.ExpiresAt, &verificationToken.UsedAt, &verificationToken.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &verificationToken, nil
}

func (p *PostgresDB) MarkEmailVerificationTokenUsed(ctx context.Context, token string) error {
	query := `
		UPDATE email_verification_tokens
		SET used_at = CURRENT_TIMESTAMP
		WHERE token = $1
	`
	_, err := p.getDB().ExecContext(ctx, query, token)
	return err
}

func (p *PostgresDB) DeleteExpiredEmailVerificationTokens(ctx context.Context) error {
	query := `
		DELETE FROM email_verification_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
	`
	_, err := p.getDB().ExecContext(ctx, query)
	return err
}

// Password reset token methods
func (p *PostgresDB) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := p.getDB().ExecContext(ctx, query,
		token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	return err
}

func (p *PostgresDB) GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`
	var resetToken PasswordResetToken
	err := p.getDB().QueryRowContext(ctx, query, token).Scan(
		&resetToken.ID, &resetToken.UserID, &resetToken.Token,
		&resetToken.ExpiresAt, &resetToken.UsedAt, &resetToken.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

func (p *PostgresDB) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = CURRENT_TIMESTAMP
		WHERE token = $1
	`
	_, err := p.getDB().ExecContext(ctx, query, token)
	return err
}

func (p *PostgresDB) DeleteExpiredPasswordResetTokens(ctx context.Context) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
	`
	_, err := p.getDB().ExecContext(ctx, query)
	return err
}

// Token blacklist methods
func (p *PostgresDB) CreateTokenBlacklist(ctx context.Context, blacklist *TokenBlacklist) error {
	query := `
		INSERT INTO token_blacklist (id, token_id, user_id, expires_at, blacklisted_at, reason)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := p.getDB().ExecContext(ctx, query,
		blacklist.ID, blacklist.TokenID, blacklist.UserID,
		blacklist.ExpiresAt, blacklist.BlacklistedAt, blacklist.Reason)
	return err
}

func (p *PostgresDB) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM token_blacklist
			WHERE token_id = $1 AND expires_at > CURRENT_TIMESTAMP
		)
	`
	var exists bool
	err := p.getDB().QueryRowContext(ctx, query, tokenID).Scan(&exists)
	return exists, err
}

func (p *PostgresDB) DeleteExpiredTokenBlacklist(ctx context.Context) error {
	query := `
		DELETE FROM token_blacklist
		WHERE expires_at < CURRENT_TIMESTAMP
	`
	_, err := p.getDB().ExecContext(ctx, query)
	return err
}

// Role assignment methods
func (p *PostgresDB) CreateRoleAssignment(ctx context.Context, assignment *RoleAssignment) error {
	query := `
		INSERT INTO role_assignments (id, user_id, role, assigned_by, assigned_at, expires_at, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := p.getDB().ExecContext(ctx, query,
		assignment.ID, assignment.UserID, assignment.Role, assignment.AssignedBy,
		assignment.AssignedAt, assignment.ExpiresAt, assignment.IsActive,
		assignment.CreatedAt, assignment.UpdatedAt)
	return err
}

func (p *PostgresDB) GetRoleAssignmentByID(ctx context.Context, id string) (*RoleAssignment, error) {
	query := `
		SELECT id, user_id, role, assigned_by, assigned_at, expires_at, is_active, created_at, updated_at
		FROM role_assignments
		WHERE id = $1
	`
	var assignment RoleAssignment
	err := p.getDB().QueryRowContext(ctx, query, id).Scan(
		&assignment.ID, &assignment.UserID, &assignment.Role, &assignment.AssignedBy,
		&assignment.AssignedAt, &assignment.ExpiresAt, &assignment.IsActive,
		&assignment.CreatedAt, &assignment.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (p *PostgresDB) GetActiveRoleAssignmentByUserID(ctx context.Context, userID string) (*RoleAssignment, error) {
	query := `
		SELECT id, user_id, role, assigned_by, assigned_at, expires_at, is_active, created_at, updated_at
		FROM role_assignments
		WHERE user_id = $1 AND is_active = true 
		AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		ORDER BY created_at DESC
		LIMIT 1
	`
	var assignment RoleAssignment
	err := p.getDB().QueryRowContext(ctx, query, userID).Scan(
		&assignment.ID, &assignment.UserID, &assignment.Role, &assignment.AssignedBy,
		&assignment.AssignedAt, &assignment.ExpiresAt, &assignment.IsActive,
		&assignment.CreatedAt, &assignment.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (p *PostgresDB) GetRoleAssignmentsByUserID(ctx context.Context, userID string) ([]*RoleAssignment, error) {
	query := `
		SELECT id, user_id, role, assigned_by, assigned_at, expires_at, is_active, created_at, updated_at
		FROM role_assignments
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := p.getDB().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*RoleAssignment
	for rows.Next() {
		var assignment RoleAssignment
		err := rows.Scan(
			&assignment.ID, &assignment.UserID, &assignment.Role, &assignment.AssignedBy,
			&assignment.AssignedAt, &assignment.ExpiresAt, &assignment.IsActive,
			&assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, &assignment)
	}

	return assignments, rows.Err()
}

func (p *PostgresDB) UpdateRoleAssignment(ctx context.Context, assignment *RoleAssignment) error {
	query := `
		UPDATE role_assignments
		SET role = $2, assigned_by = $3, assigned_at = $4, expires_at = $5, 
		    is_active = $6, updated_at = $7
		WHERE id = $1
	`
	_, err := p.getDB().ExecContext(ctx, query,
		assignment.ID, assignment.Role, assignment.AssignedBy, assignment.AssignedAt,
		assignment.ExpiresAt, assignment.IsActive, assignment.UpdatedAt)
	return err
}

func (p *PostgresDB) DeactivateRoleAssignment(ctx context.Context, id string) error {
	query := `
		UPDATE role_assignments
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := p.getDB().ExecContext(ctx, query, id)
	return err
}

func (p *PostgresDB) DeleteExpiredRoleAssignments(ctx context.Context) error {
	query := `
		UPDATE role_assignments
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE expires_at < CURRENT_TIMESTAMP AND is_active = true
	`
	_, err := p.getDB().ExecContext(ctx, query)
	return err
}

// Enhanced API key management with RBAC
func (p *PostgresDB) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	query := `
		UPDATE api_keys
		SET last_used_at = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := p.getDB().ExecContext(ctx, query, id, lastUsed)
	return err
}

func (p *PostgresDB) GetActiveAPIKeysByRole(ctx context.Context, role string) ([]*APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, role, permissions, status, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE role = $1 AND status = 'active' 
		AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		ORDER BY created_at DESC
	`
	rows, err := p.getDB().QueryContext(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		var apiKey APIKey
		err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyHash,
			&apiKey.Role, &apiKey.Permissions, &apiKey.Status,
			&apiKey.LastUsedAt, &apiKey.ExpiresAt, &apiKey.CreatedAt, &apiKey.UpdatedAt)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, &apiKey)
	}

	return apiKeys, rows.Err()
}

func (p *PostgresDB) DeactivateAPIKey(ctx context.Context, id string) error {
	query := `
		UPDATE api_keys
		SET status = 'inactive', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := p.getDB().ExecContext(ctx, query, id)
	return err
}

// GetRiskAssessmentHistory retrieves risk assessment history for a business
func (p *PostgresDB) GetRiskAssessmentHistory(ctx context.Context, businessID string, limit, offset int) ([]*RiskAssessment, error) {
	query := `
		SELECT id, business_id, business_name, overall_score, overall_level, 
		       category_scores, factor_scores, recommendations, predictions, alerts,
		       assessment_method, source, metadata, assessed_at, valid_until, 
		       created_at, updated_at
		FROM risk_assessments 
		WHERE business_id = $1 
		ORDER BY assessed_at DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := p.getDB().QueryContext(ctx, query, businessID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk assessment history: %w", err)
	}
	defer rows.Close()

	var assessments []*RiskAssessment
	for rows.Next() {
		assessment := &RiskAssessment{}
		var categoryScoresStr, factorScoresStr, recommendationsStr, predictionsStr, alertsStr, metadataStr string

		err := rows.Scan(
			&assessment.ID,
			&assessment.BusinessID,
			&assessment.BusinessName,
			&assessment.OverallScore,
			&assessment.OverallLevel,
			&categoryScoresStr,
			&factorScoresStr,
			&recommendationsStr,
			&predictionsStr,
			&alertsStr,
			&assessment.AssessmentMethod,
			&assessment.Source,
			&metadataStr,
			&assessment.AssessedAt,
			&assessment.ValidUntil,
			&assessment.CreatedAt,
			&assessment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk assessment: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(categoryScoresStr), &assessment.CategoryScores); err != nil {
			assessment.CategoryScores = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(factorScoresStr), &assessment.FactorScores); err != nil {
			assessment.FactorScores = []string{}
		}
		if err := json.Unmarshal([]byte(recommendationsStr), &assessment.Recommendations); err != nil {
			assessment.Recommendations = []string{}
		}
		if err := json.Unmarshal([]byte(predictionsStr), &assessment.Predictions); err != nil {
			assessment.Predictions = []string{}
		}
		if err := json.Unmarshal([]byte(alertsStr), &assessment.Alerts); err != nil {
			assessment.Alerts = []string{}
		}
		if err := json.Unmarshal([]byte(metadataStr), &assessment.Metadata); err != nil {
			assessment.Metadata = make(map[string]interface{})
		}

		assessments = append(assessments, assessment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk assessment history: %w", err)
	}

	return assessments, nil
}

// GetRiskAssessmentHistoryByDateRange retrieves risk assessment history within a date range
func (p *PostgresDB) GetRiskAssessmentHistoryByDateRange(ctx context.Context, businessID string, startDate, endDate time.Time) ([]*RiskAssessment, error) {
	query := `
		SELECT id, business_id, business_name, overall_score, overall_level, 
		       category_scores, factor_scores, recommendations, predictions, alerts,
		       assessment_method, source, metadata, assessed_at, valid_until, 
		       created_at, updated_at
		FROM risk_assessments 
		WHERE business_id = $1 AND assessed_at >= $2 AND assessed_at <= $3
		ORDER BY assessed_at DESC
	`

	rows, err := p.getDB().QueryContext(ctx, query, businessID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk assessment history by date range: %w", err)
	}
	defer rows.Close()

	var assessments []*RiskAssessment
	for rows.Next() {
		assessment := &RiskAssessment{}
		var categoryScoresStr, factorScoresStr, recommendationsStr, predictionsStr, alertsStr, metadataStr string

		err := rows.Scan(
			&assessment.ID,
			&assessment.BusinessID,
			&assessment.BusinessName,
			&assessment.OverallScore,
			&assessment.OverallLevel,
			&categoryScoresStr,
			&factorScoresStr,
			&recommendationsStr,
			&predictionsStr,
			&alertsStr,
			&assessment.AssessmentMethod,
			&assessment.Source,
			&metadataStr,
			&assessment.AssessedAt,
			&assessment.ValidUntil,
			&assessment.CreatedAt,
			&assessment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk assessment: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(categoryScoresStr), &assessment.CategoryScores); err != nil {
			assessment.CategoryScores = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(factorScoresStr), &assessment.FactorScores); err != nil {
			assessment.FactorScores = []string{}
		}
		if err := json.Unmarshal([]byte(recommendationsStr), &assessment.Recommendations); err != nil {
			assessment.Recommendations = []string{}
		}
		if err := json.Unmarshal([]byte(predictionsStr), &assessment.Predictions); err != nil {
			assessment.Predictions = []string{}
		}
		if err := json.Unmarshal([]byte(alertsStr), &assessment.Alerts); err != nil {
			assessment.Alerts = []string{}
		}
		if err := json.Unmarshal([]byte(metadataStr), &assessment.Metadata); err != nil {
			assessment.Metadata = make(map[string]interface{})
		}

		assessments = append(assessments, assessment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk assessment history by date range: %w", err)
	}

	return assessments, nil
}

// GetLatestRiskAssessment retrieves the most recent risk assessment for a business
func (p *PostgresDB) GetLatestRiskAssessment(ctx context.Context, businessID string) (*RiskAssessment, error) {
	query := `
		SELECT id, business_id, business_name, overall_score, overall_level, 
		       category_scores, factor_scores, recommendations, predictions, alerts,
		       assessment_method, source, metadata, assessed_at, valid_until, 
		       created_at, updated_at
		FROM risk_assessments 
		WHERE business_id = $1 
		ORDER BY assessed_at DESC 
		LIMIT 1
	`

	row := p.getDB().QueryRowContext(ctx, query, businessID)

	assessment := &RiskAssessment{}
	var categoryScoresStr, factorScoresStr, recommendationsStr, predictionsStr, alertsStr, metadataStr string

	err := row.Scan(
		&assessment.ID,
		&assessment.BusinessID,
		&assessment.BusinessName,
		&assessment.OverallScore,
		&assessment.OverallLevel,
		&categoryScoresStr,
		&factorScoresStr,
		&recommendationsStr,
		&predictionsStr,
		&alertsStr,
		&assessment.AssessmentMethod,
		&assessment.Source,
		&metadataStr,
		&assessment.AssessedAt,
		&assessment.ValidUntil,
		&assessment.CreatedAt,
		&assessment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No assessment found
		}
		return nil, fmt.Errorf("failed to scan latest risk assessment: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(categoryScoresStr), &assessment.CategoryScores); err != nil {
		assessment.CategoryScores = make(map[string]interface{})
	}
	if err := json.Unmarshal([]byte(factorScoresStr), &assessment.FactorScores); err != nil {
		assessment.FactorScores = []string{}
	}
	if err := json.Unmarshal([]byte(recommendationsStr), &assessment.Recommendations); err != nil {
		assessment.Recommendations = []string{}
	}
	if err := json.Unmarshal([]byte(predictionsStr), &assessment.Predictions); err != nil {
		assessment.Predictions = []string{}
	}
	if err := json.Unmarshal([]byte(alertsStr), &assessment.Alerts); err != nil {
		assessment.Alerts = []string{}
	}
	if err := json.Unmarshal([]byte(metadataStr), &assessment.Metadata); err != nil {
		assessment.Metadata = make(map[string]interface{})
	}

	return assessment, nil
}

// GetRiskAssessmentTrends retrieves risk assessment trends for a business
func (p *PostgresDB) GetRiskAssessmentTrends(ctx context.Context, businessID string, days int) ([]*RiskAssessment, error) {
	query := `
		SELECT id, business_id, business_name, overall_score, overall_level, 
		       category_scores, factor_scores, recommendations, predictions, alerts,
		       assessment_method, source, metadata, assessed_at, valid_until, 
		       created_at, updated_at
		FROM risk_assessments 
		WHERE business_id = $1 AND assessed_at >= NOW() - INTERVAL '1 day' * $2
		ORDER BY assessed_at DESC
	`

	rows, err := p.getDB().QueryContext(ctx, query, businessID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk assessment trends: %w", err)
	}
	defer rows.Close()

	var assessments []*RiskAssessment
	for rows.Next() {
		assessment := &RiskAssessment{}
		var categoryScoresStr, factorScoresStr, recommendationsStr, predictionsStr, alertsStr, metadataStr string

		err := rows.Scan(
			&assessment.ID,
			&assessment.BusinessID,
			&assessment.BusinessName,
			&assessment.OverallScore,
			&assessment.OverallLevel,
			&categoryScoresStr,
			&factorScoresStr,
			&recommendationsStr,
			&predictionsStr,
			&alertsStr,
			&assessment.AssessmentMethod,
			&assessment.Source,
			&metadataStr,
			&assessment.AssessedAt,
			&assessment.ValidUntil,
			&assessment.CreatedAt,
			&assessment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk assessment trend: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(categoryScoresStr), &assessment.CategoryScores); err != nil {
			assessment.CategoryScores = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(factorScoresStr), &assessment.FactorScores); err != nil {
			assessment.FactorScores = []string{}
		}
		if err := json.Unmarshal([]byte(recommendationsStr), &assessment.Recommendations); err != nil {
			assessment.Recommendations = []string{}
		}
		if err := json.Unmarshal([]byte(predictionsStr), &assessment.Predictions); err != nil {
			assessment.Predictions = []string{}
		}
		if err := json.Unmarshal([]byte(alertsStr), &assessment.Alerts); err != nil {
			assessment.Alerts = []string{}
		}
		if err := json.Unmarshal([]byte(metadataStr), &assessment.Metadata); err != nil {
			assessment.Metadata = make(map[string]interface{})
		}

		assessments = append(assessments, assessment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk assessment trends: %w", err)
	}

	return assessments, nil
}

// GetRiskAssessmentsByLevel retrieves risk assessments by risk level
func (p *PostgresDB) GetRiskAssessmentsByLevel(ctx context.Context, businessID string, riskLevel string) ([]*RiskAssessment, error) {
	query := `
		SELECT id, business_id, business_name, overall_score, overall_level, 
		       category_scores, factor_scores, recommendations, predictions, alerts,
		       assessment_method, source, metadata, assessed_at, valid_until, 
		       created_at, updated_at
		FROM risk_assessments 
		WHERE business_id = $1 AND overall_level = $2
		ORDER BY assessed_at DESC
	`

	rows, err := p.getDB().QueryContext(ctx, query, businessID, riskLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk assessments by level: %w", err)
	}
	defer rows.Close()

	var assessments []*RiskAssessment
	for rows.Next() {
		assessment := &RiskAssessment{}
		var categoryScoresStr, factorScoresStr, recommendationsStr, predictionsStr, alertsStr, metadataStr string

		err := rows.Scan(
			&assessment.ID,
			&assessment.BusinessID,
			&assessment.BusinessName,
			&assessment.OverallScore,
			&assessment.OverallLevel,
			&categoryScoresStr,
			&factorScoresStr,
			&recommendationsStr,
			&predictionsStr,
			&alertsStr,
			&assessment.AssessmentMethod,
			&assessment.Source,
			&metadataStr,
			&assessment.AssessedAt,
			&assessment.ValidUntil,
			&assessment.CreatedAt,
			&assessment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk assessment by level: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(categoryScoresStr), &assessment.CategoryScores); err != nil {
			assessment.CategoryScores = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(factorScoresStr), &assessment.FactorScores); err != nil {
			assessment.FactorScores = []string{}
		}
		if err := json.Unmarshal([]byte(recommendationsStr), &assessment.Recommendations); err != nil {
			assessment.Recommendations = []string{}
		}
		if err := json.Unmarshal([]byte(predictionsStr), &assessment.Predictions); err != nil {
			assessment.Predictions = []string{}
		}
		if err := json.Unmarshal([]byte(alertsStr), &assessment.Alerts); err != nil {
			assessment.Alerts = []string{}
		}
		if err := json.Unmarshal([]byte(metadataStr), &assessment.Metadata); err != nil {
			assessment.Metadata = make(map[string]interface{})
		}

		assessments = append(assessments, assessment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk assessments by level: %w", err)
	}

	return assessments, nil
}

// GetRiskAssessmentsByScoreRange retrieves risk assessments by score range
func (p *PostgresDB) GetRiskAssessmentsByScoreRange(ctx context.Context, businessID string, minScore, maxScore float64) ([]*RiskAssessment, error) {
	query := `
		SELECT id, business_id, business_name, overall_score, overall_level, 
		       category_scores, factor_scores, recommendations, predictions, alerts,
		       assessment_method, source, metadata, assessed_at, valid_until, 
		       created_at, updated_at
		FROM risk_assessments 
		WHERE business_id = $1 AND overall_score >= $2 AND overall_score <= $3
		ORDER BY assessed_at DESC
	`

	rows, err := p.getDB().QueryContext(ctx, query, businessID, minScore, maxScore)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk assessments by score range: %w", err)
	}
	defer rows.Close()

	var assessments []*RiskAssessment
	for rows.Next() {
		assessment := &RiskAssessment{}
		var categoryScoresStr, factorScoresStr, recommendationsStr, predictionsStr, alertsStr, metadataStr string

		err := rows.Scan(
			&assessment.ID,
			&assessment.BusinessID,
			&assessment.BusinessName,
			&assessment.OverallScore,
			&assessment.OverallLevel,
			&categoryScoresStr,
			&factorScoresStr,
			&recommendationsStr,
			&predictionsStr,
			&alertsStr,
			&assessment.AssessmentMethod,
			&assessment.Source,
			&metadataStr,
			&assessment.AssessedAt,
			&assessment.ValidUntil,
			&assessment.CreatedAt,
			&assessment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk assessment by score range: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(categoryScoresStr), &assessment.CategoryScores); err != nil {
			assessment.CategoryScores = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(factorScoresStr), &assessment.FactorScores); err != nil {
			assessment.FactorScores = []string{}
		}
		if err := json.Unmarshal([]byte(recommendationsStr), &assessment.Recommendations); err != nil {
			assessment.Recommendations = []string{}
		}
		if err := json.Unmarshal([]byte(predictionsStr), &assessment.Predictions); err != nil {
			assessment.Predictions = []string{}
		}
		if err := json.Unmarshal([]byte(alertsStr), &assessment.Alerts); err != nil {
			assessment.Alerts = []string{}
		}
		if err := json.Unmarshal([]byte(metadataStr), &assessment.Metadata); err != nil {
			assessment.Metadata = make(map[string]interface{})
		}

		assessments = append(assessments, assessment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk assessments by score range: %w", err)
	}

	return assessments, nil
}

// GetRiskAssessmentStatistics retrieves risk assessment statistics for a business
func (p *PostgresDB) GetRiskAssessmentStatistics(ctx context.Context, businessID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_assessments,
			AVG(overall_score) as average_score,
			MIN(overall_score) as min_score,
			MAX(overall_score) as max_score,
			COUNT(CASE WHEN overall_level = 'low' THEN 1 END) as low_count,
			COUNT(CASE WHEN overall_level = 'medium' THEN 1 END) as medium_count,
			COUNT(CASE WHEN overall_level = 'high' THEN 1 END) as high_count,
			COUNT(CASE WHEN overall_level = 'critical' THEN 1 END) as critical_count,
			MIN(assessed_at) as first_assessment,
			MAX(assessed_at) as last_assessment
		FROM risk_assessments 
		WHERE business_id = $1
	`

	row := p.getDB().QueryRowContext(ctx, query, businessID)

	var stats struct {
		TotalAssessments int       `db:"total_assessments"`
		AverageScore     float64   `db:"average_score"`
		MinScore         float64   `db:"min_score"`
		MaxScore         float64   `db:"max_score"`
		LowCount         int       `db:"low_count"`
		MediumCount      int       `db:"medium_count"`
		HighCount        int       `db:"high_count"`
		CriticalCount    int       `db:"critical_count"`
		FirstAssessment  time.Time `db:"first_assessment"`
		LastAssessment   time.Time `db:"last_assessment"`
	}

	err := row.Scan(
		&stats.TotalAssessments,
		&stats.AverageScore,
		&stats.MinScore,
		&stats.MaxScore,
		&stats.LowCount,
		&stats.MediumCount,
		&stats.HighCount,
		&stats.CriticalCount,
		&stats.FirstAssessment,
		&stats.LastAssessment,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan risk assessment statistics: %w", err)
	}

	result := map[string]interface{}{
		"total_assessments": stats.TotalAssessments,
		"average_score":     stats.AverageScore,
		"min_score":         stats.MinScore,
		"max_score":         stats.MaxScore,
		"level_distribution": map[string]int{
			"low":      stats.LowCount,
			"medium":   stats.MediumCount,
			"high":     stats.HighCount,
			"critical": stats.CriticalCount,
		},
		"first_assessment": stats.FirstAssessment,
		"last_assessment":  stats.LastAssessment,
	}

	return result, nil
}
