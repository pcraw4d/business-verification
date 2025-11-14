package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"kyb-platform/internal/models"
)

// MerchantPortfolioRepository provides data access operations for merchant portfolio
type MerchantPortfolioRepository struct {
	db     *sql.DB
	logger *log.Logger
}

// NewMerchantPortfolioRepository creates a new merchant portfolio repository
func NewMerchantPortfolioRepository(db *sql.DB, logger *log.Logger) *MerchantPortfolioRepository {
	if logger == nil {
		logger = log.Default()
	}

	return &MerchantPortfolioRepository{
		db:     db,
		logger: logger,
	}
}

// Repository errors
var (
	ErrMerchantNotFound     = errors.New("merchant not found")
	ErrSessionNotFound      = errors.New("session not found")
	ErrAuditLogNotFound     = errors.New("audit log not found")
	ErrNotificationNotFound = errors.New("notification not found")
	ErrComparisonNotFound   = errors.New("comparison not found")
	ErrAnalyticsNotFound    = errors.New("analytics not found")
	ErrDuplicateMerchant    = errors.New("merchant already exists")
	ErrDuplicateSession     = errors.New("session already exists")
	ErrInvalidPagination    = errors.New("invalid pagination parameters")
)

// =============================================================================
// Merchant CRUD Operations
// =============================================================================

// CreateMerchant creates a new merchant in the database
func (r *MerchantPortfolioRepository) CreateMerchant(ctx context.Context, merchant *models.Merchant) error {
	r.logger.Printf("Creating merchant: %s", merchant.ID)

	query := `
		INSERT INTO merchants (
			id, name, legal_name, registration_number, tax_id, industry, industry_code,
			business_type, founded_date, employee_count, annual_revenue,
			address_street1, address_street2, address_city, address_state, 
			address_postal_code, address_country, address_country_code,
			contact_phone, contact_email, contact_website, contact_primary_contact,
			portfolio_type_id, risk_level_id, compliance_status, status,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22,
			$23, $24, $25, $26, $27, $28, $29
		)
	`

	// Get portfolio type and risk level IDs
	portfolioTypeID, err := r.getPortfolioTypeID(ctx, string(merchant.PortfolioType))
	if err != nil {
		return fmt.Errorf("failed to get portfolio type ID: %w", err)
	}

	riskLevelID, err := r.getRiskLevelID(ctx, string(merchant.RiskLevel))
	if err != nil {
		return fmt.Errorf("failed to get risk level ID: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		merchant.ID,
		merchant.Name,
		merchant.LegalName,
		merchant.RegistrationNumber,
		merchant.TaxID,
		merchant.Industry,
		merchant.IndustryCode,
		merchant.BusinessType,
		merchant.FoundedDate,
		merchant.EmployeeCount,
		merchant.AnnualRevenue,
		merchant.Address.Street1,
		merchant.Address.Street2,
		merchant.Address.City,
		merchant.Address.State,
		merchant.Address.PostalCode,
		merchant.Address.Country,
		merchant.Address.CountryCode,
		merchant.ContactInfo.Phone,
		merchant.ContactInfo.Email,
		merchant.ContactInfo.Website,
		merchant.ContactInfo.PrimaryContact,
		portfolioTypeID,
		riskLevelID,
		merchant.ComplianceStatus,
		merchant.Status,
		merchant.CreatedBy,
		merchant.CreatedAt,
		merchant.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return ErrDuplicateMerchant
		}
		return fmt.Errorf("failed to create merchant: %w", err)
	}

	r.logger.Printf("Successfully created merchant: %s", merchant.ID)
	return nil
}

// GetMerchant retrieves a merchant by ID
func (r *MerchantPortfolioRepository) GetMerchant(ctx context.Context, merchantID string) (*models.Merchant, error) {
	r.logger.Printf("Retrieving merchant: %s", merchantID)

	query := `
		SELECT 
			m.id, m.name, m.legal_name, m.registration_number, m.tax_id, m.industry, m.industry_code,
			m.business_type, m.founded_date, m.employee_count, m.annual_revenue,
			m.address_street1, m.address_street2, m.address_city, m.address_state, 
			m.address_postal_code, m.address_country, m.address_country_code,
			m.contact_phone, m.contact_email, m.contact_website, m.contact_primary_contact,
			pt.type as portfolio_type, rl.level as risk_level, m.compliance_status, m.status,
			m.created_by, m.created_at, m.updated_at
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE m.id = $1
	`

	row := r.db.QueryRowContext(ctx, query, merchantID)
	merchant, err := r.scanMerchant(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrMerchantNotFound
		}
		return nil, fmt.Errorf("failed to retrieve merchant: %w", err)
	}

	return merchant, nil
}

// UpdateMerchant updates an existing merchant
func (r *MerchantPortfolioRepository) UpdateMerchant(ctx context.Context, merchant *models.Merchant) error {
	r.logger.Printf("Updating merchant: %s", merchant.ID)

	// Get portfolio type and risk level IDs
	portfolioTypeID, err := r.getPortfolioTypeID(ctx, string(merchant.PortfolioType))
	if err != nil {
		return fmt.Errorf("failed to get portfolio type ID: %w", err)
	}

	riskLevelID, err := r.getRiskLevelID(ctx, string(merchant.RiskLevel))
	if err != nil {
		return fmt.Errorf("failed to get risk level ID: %w", err)
	}

	query := `
		UPDATE merchants SET
			name = $2, legal_name = $3, registration_number = $4, tax_id = $5,
			industry = $6, industry_code = $7, business_type = $8, founded_date = $9,
			employee_count = $10, annual_revenue = $11,
			address_street1 = $12, address_street2 = $13, address_city = $14, address_state = $15,
			address_postal_code = $16, address_country = $17, address_country_code = $18,
			contact_phone = $19, contact_email = $20, contact_website = $21, contact_primary_contact = $22,
			portfolio_type_id = $23, risk_level_id = $24, compliance_status = $25, status = $26,
			updated_at = $27
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		merchant.ID,
		merchant.Name,
		merchant.LegalName,
		merchant.RegistrationNumber,
		merchant.TaxID,
		merchant.Industry,
		merchant.IndustryCode,
		merchant.BusinessType,
		merchant.FoundedDate,
		merchant.EmployeeCount,
		merchant.AnnualRevenue,
		merchant.Address.Street1,
		merchant.Address.Street2,
		merchant.Address.City,
		merchant.Address.State,
		merchant.Address.PostalCode,
		merchant.Address.Country,
		merchant.Address.CountryCode,
		merchant.ContactInfo.Phone,
		merchant.ContactInfo.Email,
		merchant.ContactInfo.Website,
		merchant.ContactInfo.PrimaryContact,
		portfolioTypeID,
		riskLevelID,
		merchant.ComplianceStatus,
		merchant.Status,
		merchant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMerchantNotFound
	}

	r.logger.Printf("Successfully updated merchant: %s", merchant.ID)
	return nil
}

// DeleteMerchant deletes a merchant from the database
func (r *MerchantPortfolioRepository) DeleteMerchant(ctx context.Context, merchantID string) error {
	r.logger.Printf("Deleting merchant: %s", merchantID)

	query := `DELETE FROM merchants WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, merchantID)
	if err != nil {
		return fmt.Errorf("failed to delete merchant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMerchantNotFound
	}

	r.logger.Printf("Successfully deleted merchant: %s", merchantID)
	return nil
}

// =============================================================================
// Merchant Search and Filtering
// =============================================================================

// ListMerchants retrieves merchants with pagination
func (r *MerchantPortfolioRepository) ListMerchants(ctx context.Context, page, pageSize int) ([]*models.Merchant, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, ErrInvalidPagination
	}

	offset := (page - 1) * pageSize
	r.logger.Printf("Listing merchants (page: %d, size: %d, offset: %d)", page, pageSize, offset)

	query := `
		SELECT 
			m.id, m.name, m.legal_name, m.registration_number, m.tax_id, m.industry, m.industry_code,
			m.business_type, m.founded_date, m.employee_count, m.annual_revenue,
			m.address_street1, m.address_street2, m.address_city, m.address_state, 
			m.address_postal_code, m.address_country, m.address_country_code,
			m.contact_phone, m.contact_email, m.contact_website, m.contact_primary_contact,
			pt.type as portfolio_type, rl.level as risk_level, m.compliance_status, m.status,
			m.created_by, m.created_at, m.updated_at
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		ORDER BY m.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list merchants: %w", err)
	}
	defer rows.Close()

	var merchants []*models.Merchant
	for rows.Next() {
		merchant, err := r.scanMerchant(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merchant: %w", err)
		}
		merchants = append(merchants, merchant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating merchants: %w", err)
	}

	r.logger.Printf("Retrieved %d merchants", len(merchants))
	return merchants, nil
}

// SearchMerchants searches merchants with filters
func (r *MerchantPortfolioRepository) SearchMerchants(ctx context.Context, filters *models.MerchantSearchFilters, page, pageSize int) ([]*models.Merchant, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, ErrInvalidPagination
	}

	offset := (page - 1) * pageSize
	r.logger.Printf("Searching merchants with filters (page: %d, size: %d)", page, pageSize)

	// Build dynamic query
	query, args := r.buildSearchQuery(filters, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search merchants: %w", err)
	}
	defer rows.Close()

	var merchants []*models.Merchant
	for rows.Next() {
		merchant, err := r.scanMerchant(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merchant: %w", err)
		}
		merchants = append(merchants, merchant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating merchants: %w", err)
	}

	r.logger.Printf("Found %d merchants matching filters", len(merchants))
	return merchants, nil
}

// CountMerchants counts total merchants matching filters
func (r *MerchantPortfolioRepository) CountMerchants(ctx context.Context, filters *models.MerchantSearchFilters) (int, error) {
	r.logger.Printf("Counting merchants with filters")

	// Build count query
	query, args := r.buildCountQuery(filters)

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count merchants: %w", err)
	}

	r.logger.Printf("Found %d merchants matching filters", count)
	return count, nil
}

// =============================================================================
// Portfolio Type and Risk Level Operations
// =============================================================================

// GetMerchantsByPortfolioType retrieves merchants by portfolio type
func (r *MerchantPortfolioRepository) GetMerchantsByPortfolioType(ctx context.Context, portfolioType models.PortfolioType, page, pageSize int) ([]*models.Merchant, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, ErrInvalidPagination
	}

	offset := (page - 1) * pageSize
	r.logger.Printf("Getting merchants by portfolio type: %s (page: %d, size: %d)", portfolioType, page, pageSize)

	query := `
		SELECT 
			m.id, m.name, m.legal_name, m.registration_number, m.tax_id, m.industry, m.industry_code,
			m.business_type, m.founded_date, m.employee_count, m.annual_revenue,
			m.address_street1, m.address_street2, m.address_city, m.address_state, 
			m.address_postal_code, m.address_country, m.address_country_code,
			m.contact_phone, m.contact_email, m.contact_website, m.contact_primary_contact,
			pt.type as portfolio_type, rl.level as risk_level, m.compliance_status, m.status,
			m.created_by, m.created_at, m.updated_at
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE pt.type = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(portfolioType), pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchants by portfolio type: %w", err)
	}
	defer rows.Close()

	var merchants []*models.Merchant
	for rows.Next() {
		merchant, err := r.scanMerchant(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merchant: %w", err)
		}
		merchants = append(merchants, merchant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating merchants: %w", err)
	}

	r.logger.Printf("Retrieved %d merchants with portfolio type %s", len(merchants), portfolioType)
	return merchants, nil
}

// GetMerchantsByRiskLevel retrieves merchants by risk level
func (r *MerchantPortfolioRepository) GetMerchantsByRiskLevel(ctx context.Context, riskLevel models.RiskLevel, page, pageSize int) ([]*models.Merchant, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, ErrInvalidPagination
	}

	offset := (page - 1) * pageSize
	r.logger.Printf("Getting merchants by risk level: %s (page: %d, size: %d)", riskLevel, page, pageSize)

	query := `
		SELECT 
			m.id, m.name, m.legal_name, m.registration_number, m.tax_id, m.industry, m.industry_code,
			m.business_type, m.founded_date, m.employee_count, m.annual_revenue,
			m.address_street1, m.address_street2, m.address_city, m.address_state, 
			m.address_postal_code, m.address_country, m.address_country_code,
			m.contact_phone, m.contact_email, m.contact_website, m.contact_primary_contact,
			pt.type as portfolio_type, rl.level as risk_level, m.compliance_status, m.status,
			m.created_by, m.created_at, m.updated_at
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE rl.level = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(riskLevel), pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchants by risk level: %w", err)
	}
	defer rows.Close()

	var merchants []*models.Merchant
	for rows.Next() {
		merchant, err := r.scanMerchant(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merchant: %w", err)
		}
		merchants = append(merchants, merchant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating merchants: %w", err)
	}

	r.logger.Printf("Retrieved %d merchants with risk level %s", len(merchants), riskLevel)
	return merchants, nil
}

// =============================================================================
// Session Management
// =============================================================================

// CreateSession creates a new merchant session
func (r *MerchantPortfolioRepository) CreateSession(ctx context.Context, session *models.MerchantSession) error {
	r.logger.Printf("Creating session: %s", session.ID)

	query := `
		INSERT INTO merchant_sessions (
			id, user_id, merchant_id, started_at, last_active, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.MerchantID,
		session.StartedAt,
		session.LastActive,
		session.IsActive,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return ErrDuplicateSession
		}
		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.Printf("Successfully created session: %s", session.ID)
	return nil
}

// GetActiveSessionByUserID retrieves the active session for a user
func (r *MerchantPortfolioRepository) GetActiveSessionByUserID(ctx context.Context, userID string) (*models.MerchantSession, error) {
	r.logger.Printf("Getting active session for user: %s", userID)

	query := `
		SELECT id, user_id, merchant_id, started_at, last_active, is_active, created_at, updated_at
		FROM merchant_sessions 
		WHERE user_id = $1 AND is_active = true
		ORDER BY last_active DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, userID)
	session, err := r.scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	return session, nil
}

// UpdateSession updates a merchant session
func (r *MerchantPortfolioRepository) UpdateSession(ctx context.Context, session *models.MerchantSession) error {
	r.logger.Printf("Updating session: %s", session.ID)

	query := `
		UPDATE merchant_sessions SET
			last_active = $2, is_active = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.LastActive,
		session.IsActive,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	r.logger.Printf("Successfully updated session: %s", session.ID)
	return nil
}

// DeactivateSession deactivates a session
func (r *MerchantPortfolioRepository) DeactivateSession(ctx context.Context, sessionID string) error {
	r.logger.Printf("Deactivating session: %s", sessionID)

	query := `
		UPDATE merchant_sessions SET
			is_active = false, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, sessionID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	r.logger.Printf("Successfully deactivated session: %s", sessionID)
	return nil
}

// =============================================================================
// Audit Logging
// =============================================================================

// CreateAuditLog creates a new audit log entry
func (r *MerchantPortfolioRepository) CreateAuditLog(ctx context.Context, auditLog *models.AuditLog) error {
	r.logger.Printf("Creating audit log: %s", auditLog.ID)

	query := `
		INSERT INTO audit_logs (
			id, user_id, merchant_id, action, resource_type, resource_id,
			details, ip_address, user_agent, request_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		auditLog.ID,
		auditLog.UserID,
		auditLog.MerchantID,
		auditLog.Action,
		auditLog.ResourceType,
		auditLog.ResourceID,
		auditLog.Details,
		auditLog.IPAddress,
		auditLog.UserAgent,
		auditLog.RequestID,
		auditLog.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	r.logger.Printf("Successfully created audit log: %s", auditLog.ID)
	return nil
}

// GetAuditLogsByMerchantID retrieves audit logs for a merchant
func (r *MerchantPortfolioRepository) GetAuditLogsByMerchantID(ctx context.Context, merchantID string, page, pageSize int) ([]*models.AuditLog, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, ErrInvalidPagination
	}

	offset := (page - 1) * pageSize
	r.logger.Printf("Getting audit logs for merchant: %s (page: %d, size: %d)", merchantID, page, pageSize)

	query := `
		SELECT id, user_id, merchant_id, action, resource_type, resource_id,
			   details, ip_address, user_agent, request_id, created_at
		FROM audit_logs 
		WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, merchantID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []*models.AuditLog
	for rows.Next() {
		auditLog, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	r.logger.Printf("Retrieved %d audit logs for merchant %s", len(auditLogs), merchantID)
	return auditLogs, nil
}

// =============================================================================
// Bulk Operations
// =============================================================================

// BulkUpdatePortfolioType updates portfolio type for multiple merchants
func (r *MerchantPortfolioRepository) BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType models.PortfolioType, userID string) error {
	if len(merchantIDs) == 0 {
		return nil
	}

	r.logger.Printf("Bulk updating portfolio type for %d merchants to %s", len(merchantIDs), portfolioType)

	// Build placeholders for IN clause
	placeholders := make([]string, len(merchantIDs))
	args := make([]interface{}, len(merchantIDs)+2)

	for i, id := range merchantIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// Get portfolio type ID
	portfolioTypeID, err := r.getPortfolioTypeID(ctx, string(portfolioType))
	if err != nil {
		return fmt.Errorf("failed to get portfolio type ID: %w", err)
	}

	args[len(merchantIDs)] = portfolioTypeID
	args[len(merchantIDs)+1] = time.Now()

	query := fmt.Sprintf(`
		UPDATE merchants 
		SET portfolio_type_id = $%d, updated_at = $%d
		WHERE id IN (%s)
	`, len(merchantIDs)+1, len(merchantIDs)+2, strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk update portfolio type: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Printf("Successfully updated %d merchants", rowsAffected)
	return nil
}

// BulkUpdateRiskLevel updates risk level for multiple merchants
func (r *MerchantPortfolioRepository) BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel models.RiskLevel, userID string) error {
	if len(merchantIDs) == 0 {
		return nil
	}

	r.logger.Printf("Bulk updating risk level for %d merchants to %s", len(merchantIDs), riskLevel)

	// Build placeholders for IN clause
	placeholders := make([]string, len(merchantIDs))
	args := make([]interface{}, len(merchantIDs)+2)

	for i, id := range merchantIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// Get risk level ID
	riskLevelID, err := r.getRiskLevelID(ctx, string(riskLevel))
	if err != nil {
		return fmt.Errorf("failed to get risk level ID: %w", err)
	}

	args[len(merchantIDs)] = riskLevelID
	args[len(merchantIDs)+1] = time.Now()

	query := fmt.Sprintf(`
		UPDATE merchants 
		SET risk_level_id = $%d, updated_at = $%d
		WHERE id IN (%s)
	`, len(merchantIDs)+1, len(merchantIDs)+2, strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk update risk level: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Printf("Successfully updated %d merchants", rowsAffected)
	return nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// scanMerchant scans a merchant from a database row
func (r *MerchantPortfolioRepository) scanMerchant(scanner interface{}) (*models.Merchant, error) {
	if scanner == nil {
		return nil, errors.New("scanner is nil")
	}

	var merchant models.Merchant
	var portfolioTypeStr, riskLevelStr string

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&merchant.ID,
			&merchant.Name,
			&merchant.LegalName,
			&merchant.RegistrationNumber,
			&merchant.TaxID,
			&merchant.Industry,
			&merchant.IndustryCode,
			&merchant.BusinessType,
			&merchant.FoundedDate,
			&merchant.EmployeeCount,
			&merchant.AnnualRevenue,
			&merchant.Address.Street1,
			&merchant.Address.Street2,
			&merchant.Address.City,
			&merchant.Address.State,
			&merchant.Address.PostalCode,
			&merchant.Address.Country,
			&merchant.Address.CountryCode,
			&merchant.ContactInfo.Phone,
			&merchant.ContactInfo.Email,
			&merchant.ContactInfo.Website,
			&merchant.ContactInfo.PrimaryContact,
			&portfolioTypeStr,
			&riskLevelStr,
			&merchant.ComplianceStatus,
			&merchant.Status,
			&merchant.CreatedBy,
			&merchant.CreatedAt,
			&merchant.UpdatedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&merchant.ID,
			&merchant.Name,
			&merchant.LegalName,
			&merchant.RegistrationNumber,
			&merchant.TaxID,
			&merchant.Industry,
			&merchant.IndustryCode,
			&merchant.BusinessType,
			&merchant.FoundedDate,
			&merchant.EmployeeCount,
			&merchant.AnnualRevenue,
			&merchant.Address.Street1,
			&merchant.Address.Street2,
			&merchant.Address.City,
			&merchant.Address.State,
			&merchant.Address.PostalCode,
			&merchant.Address.Country,
			&merchant.Address.CountryCode,
			&merchant.ContactInfo.Phone,
			&merchant.ContactInfo.Email,
			&merchant.ContactInfo.Website,
			&merchant.ContactInfo.PrimaryContact,
			&portfolioTypeStr,
			&riskLevelStr,
			&merchant.ComplianceStatus,
			&merchant.Status,
			&merchant.CreatedBy,
			&merchant.CreatedAt,
			&merchant.UpdatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		return nil, err
	}

	// Convert string enums to typed enums
	merchant.PortfolioType = models.PortfolioType(portfolioTypeStr)
	merchant.RiskLevel = models.RiskLevel(riskLevelStr)

	return &merchant, nil
}

// scanSession scans a session from a database row
func (r *MerchantPortfolioRepository) scanSession(row *sql.Row) (*models.MerchantSession, error) {
	if row == nil {
		return nil, errors.New("row is nil")
	}

	var session models.MerchantSession

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.MerchantID,
		&session.StartedAt,
		&session.LastActive,
		&session.IsActive,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// scanAuditLog scans an audit log from a database row
func (r *MerchantPortfolioRepository) scanAuditLog(scanner interface{}) (*models.AuditLog, error) {
	if scanner == nil {
		return nil, errors.New("scanner is nil")
	}

	var auditLog models.AuditLog

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&auditLog.ID,
			&auditLog.UserID,
			&auditLog.MerchantID,
			&auditLog.Action,
			&auditLog.ResourceType,
			&auditLog.ResourceID,
			&auditLog.Details,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.RequestID,
			&auditLog.CreatedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&auditLog.ID,
			&auditLog.UserID,
			&auditLog.MerchantID,
			&auditLog.Action,
			&auditLog.ResourceType,
			&auditLog.ResourceID,
			&auditLog.Details,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.RequestID,
			&auditLog.CreatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		return nil, err
	}

	return &auditLog, nil
}

// buildSearchQuery builds a dynamic search query based on filters
func (r *MerchantPortfolioRepository) buildSearchQuery(filters *models.MerchantSearchFilters, limit, offset int) (string, []interface{}) {
	baseQuery := `
		SELECT 
			m.id, m.name, m.legal_name, m.registration_number, m.tax_id, m.industry, m.industry_code,
			m.business_type, m.founded_date, m.employee_count, m.annual_revenue,
			m.address_street1, m.address_street2, m.address_city, m.address_state, 
			m.address_postal_code, m.address_country, m.address_country_code,
			m.contact_phone, m.contact_email, m.contact_website, m.contact_primary_contact,
			pt.type as portfolio_type, rl.level as risk_level, m.compliance_status, m.status,
			m.created_by, m.created_at, m.updated_at
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters != nil {
		if filters.PortfolioType != nil {
			conditions = append(conditions, fmt.Sprintf("pt.type = $%d", argIndex))
			args = append(args, string(*filters.PortfolioType))
			argIndex++
		}

		if filters.RiskLevel != nil {
			conditions = append(conditions, fmt.Sprintf("rl.level = $%d", argIndex))
			args = append(args, string(*filters.RiskLevel))
			argIndex++
		}

		if filters.Industry != "" {
			conditions = append(conditions, fmt.Sprintf("m.industry ILIKE $%d", argIndex))
			args = append(args, "%"+filters.Industry+"%")
			argIndex++
		}

		if filters.Status != "" {
			conditions = append(conditions, fmt.Sprintf("m.status = $%d", argIndex))
			args = append(args, filters.Status)
			argIndex++
		}

		if filters.SearchQuery != "" {
			searchCondition := fmt.Sprintf("(name ILIKE $%d OR legal_name ILIKE $%d OR m.industry ILIKE $%d)", argIndex, argIndex, argIndex)
			conditions = append(conditions, searchCondition)
			args = append(args, "%"+filters.SearchQuery+"%")
			argIndex++
		}

		if filters.CreatedAfter != nil {
			conditions = append(conditions, fmt.Sprintf("m.created_at >= $%d", argIndex))
			args = append(args, *filters.CreatedAfter)
			argIndex++
		}

		if filters.CreatedBefore != nil {
			conditions = append(conditions, fmt.Sprintf("m.created_at <= $%d", argIndex))
			args = append(args, *filters.CreatedBefore)
			argIndex++
		}

		if filters.EmployeeCountMin != nil {
			conditions = append(conditions, fmt.Sprintf("m.employee_count >= $%d", argIndex))
			args = append(args, *filters.EmployeeCountMin)
			argIndex++
		}

		if filters.EmployeeCountMax != nil {
			conditions = append(conditions, fmt.Sprintf("m.employee_count <= $%d", argIndex))
			args = append(args, *filters.EmployeeCountMax)
			argIndex++
		}

		if filters.RevenueMin != nil {
			conditions = append(conditions, fmt.Sprintf("m.annual_revenue >= $%d", argIndex))
			args = append(args, *filters.RevenueMin)
			argIndex++
		}

		if filters.RevenueMax != nil {
			conditions = append(conditions, fmt.Sprintf("m.annual_revenue <= $%d", argIndex))
			args = append(args, *filters.RevenueMax)
			argIndex++
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY m.created_at DESC"
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	return baseQuery, args
}

// buildCountQuery builds a count query based on filters
func (r *MerchantPortfolioRepository) buildCountQuery(filters *models.MerchantSearchFilters) (string, []interface{}) {
	baseQuery := `
		SELECT COUNT(*) 
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters != nil {
		if filters.PortfolioType != nil {
			conditions = append(conditions, fmt.Sprintf("pt.type = $%d", argIndex))
			args = append(args, string(*filters.PortfolioType))
			argIndex++
		}

		if filters.RiskLevel != nil {
			conditions = append(conditions, fmt.Sprintf("rl.level = $%d", argIndex))
			args = append(args, string(*filters.RiskLevel))
			argIndex++
		}

		if filters.Industry != "" {
			conditions = append(conditions, fmt.Sprintf("m.industry ILIKE $%d", argIndex))
			args = append(args, "%"+filters.Industry+"%")
			argIndex++
		}

		if filters.Status != "" {
			conditions = append(conditions, fmt.Sprintf("m.status = $%d", argIndex))
			args = append(args, filters.Status)
			argIndex++
		}

		if filters.SearchQuery != "" {
			searchCondition := fmt.Sprintf("(name ILIKE $%d OR legal_name ILIKE $%d OR m.industry ILIKE $%d)", argIndex, argIndex, argIndex)
			conditions = append(conditions, searchCondition)
			args = append(args, "%"+filters.SearchQuery+"%")
			argIndex++
		}

		if filters.CreatedAfter != nil {
			conditions = append(conditions, fmt.Sprintf("m.created_at >= $%d", argIndex))
			args = append(args, *filters.CreatedAfter)
			argIndex++
		}

		if filters.CreatedBefore != nil {
			conditions = append(conditions, fmt.Sprintf("m.created_at <= $%d", argIndex))
			args = append(args, *filters.CreatedBefore)
			argIndex++
		}

		if filters.EmployeeCountMin != nil {
			conditions = append(conditions, fmt.Sprintf("m.employee_count >= $%d", argIndex))
			args = append(args, *filters.EmployeeCountMin)
			argIndex++
		}

		if filters.EmployeeCountMax != nil {
			conditions = append(conditions, fmt.Sprintf("m.employee_count <= $%d", argIndex))
			args = append(args, *filters.EmployeeCountMax)
			argIndex++
		}

		if filters.RevenueMin != nil {
			conditions = append(conditions, fmt.Sprintf("m.annual_revenue >= $%d", argIndex))
			args = append(args, *filters.RevenueMin)
			argIndex++
		}

		if filters.RevenueMax != nil {
			conditions = append(conditions, fmt.Sprintf("m.annual_revenue <= $%d", argIndex))
			args = append(args, *filters.RevenueMax)
			argIndex++
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	return baseQuery, args
}

// getPortfolioTypeID gets the portfolio type ID by type name
func (r *MerchantPortfolioRepository) getPortfolioTypeID(ctx context.Context, portfolioType string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, "SELECT id FROM portfolio_types WHERE type = $1", portfolioType).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to get portfolio type ID for %s: %w", portfolioType, err)
	}
	return id, nil
}

// getRiskLevelID gets the risk level ID by level name
func (r *MerchantPortfolioRepository) getRiskLevelID(ctx context.Context, riskLevel string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, "SELECT id FROM risk_levels WHERE level = $1", riskLevel).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to get risk level ID for %s: %w", riskLevel, err)
	}
	return id, nil
}

// GetPortfolioDistribution gets count of merchants by portfolio type
func (r *MerchantPortfolioRepository) GetPortfolioDistribution(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT pt.type, COUNT(*) as count
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		GROUP BY pt.type
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query portfolio distribution: %w", err)
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var portfolioType string
		var count int
		if err := rows.Scan(&portfolioType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan portfolio distribution: %w", err)
		}
		dist[portfolioType] = count
	}

	return dist, rows.Err()
}

// GetRiskDistribution gets count of merchants by risk level
func (r *MerchantPortfolioRepository) GetRiskDistribution(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT rl.level, COUNT(*) as count
		FROM merchants m
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		GROUP BY rl.level
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk distribution: %w", err)
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var riskLevel string
		var count int
		if err := rows.Scan(&riskLevel, &count); err != nil {
			return nil, fmt.Errorf("failed to scan risk distribution: %w", err)
		}
		dist[riskLevel] = count
	}

	return dist, rows.Err()
}

// GetIndustryDistribution gets count of merchants by industry
func (r *MerchantPortfolioRepository) GetIndustryDistribution(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT COALESCE(industry, 'Unknown') as industry, COUNT(*) as count
		FROM merchants
		WHERE industry IS NOT NULL AND industry != ''
		GROUP BY industry
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query industry distribution: %w", err)
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var industry string
		var count int
		if err := rows.Scan(&industry, &count); err != nil {
			return nil, fmt.Errorf("failed to scan industry distribution: %w", err)
		}
		dist[industry] = count
	}

	return dist, rows.Err()
}

// GetComplianceDistribution gets count of merchants by compliance status
func (r *MerchantPortfolioRepository) GetComplianceDistribution(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT COALESCE(compliance_status, 'pending') as compliance_status, COUNT(*) as count
		FROM merchants
		GROUP BY compliance_status
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance distribution: %w", err)
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan compliance distribution: %w", err)
		}
		dist[status] = count
	}

	return dist, rows.Err()
}
