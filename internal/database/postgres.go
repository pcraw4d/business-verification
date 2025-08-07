package database

import (
	"context"
	"database/sql"
	"fmt"

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
