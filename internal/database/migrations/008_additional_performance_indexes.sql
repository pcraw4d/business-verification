-- Migration: 008_additional_performance_indexes.sql
-- Description: Add additional indexes for complex query patterns and performance optimization
-- Created: 2025-01-19
-- Dependencies: 005_merchant_portfolio_schema.sql

-- Additional composite indexes for complex queries

-- Index for merchant search with multiple filters
CREATE INDEX IF NOT EXISTS idx_merchants_search_composite ON merchants 
USING btree (status, compliance_status, created_at DESC);

-- Index for portfolio type + risk level + status combination
CREATE INDEX IF NOT EXISTS idx_merchants_portfolio_risk_status ON merchants 
USING btree (portfolio_type_id, risk_level_id, status);

-- Index for industry + business type filtering
CREATE INDEX IF NOT EXISTS idx_merchants_industry_business_type ON merchants 
USING btree (industry, business_type);

-- Index for address-based searches
CREATE INDEX IF NOT EXISTS idx_merchants_address_country_state ON merchants 
USING btree (address_country, address_state);

-- Index for contact information searches
CREATE INDEX IF NOT EXISTS idx_merchants_contact_composite ON merchants 
USING btree (contact_email, contact_phone);

-- Index for financial filtering
CREATE INDEX IF NOT EXISTS idx_merchants_financial_composite ON merchants 
USING btree (employee_count, annual_revenue);

-- Index for date range queries
CREATE INDEX IF NOT EXISTS idx_merchants_date_range ON merchants 
USING btree (created_at, updated_at);

-- Index for user-specific merchant queries
CREATE INDEX IF NOT EXISTS idx_merchants_user_created ON merchants 
USING btree (created_by, created_at DESC);

-- Additional audit log indexes for complex queries

-- Index for audit logs by user and time range
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_time_range ON merchant_audit_logs 
USING btree (user_id, created_at DESC);

-- Index for audit logs by merchant and action type
CREATE INDEX IF NOT EXISTS idx_audit_logs_merchant_action ON merchant_audit_logs 
USING btree (merchant_id, action, created_at DESC);

-- Index for audit logs by resource type and ID
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON merchant_audit_logs 
USING btree (resource_type, resource_id, created_at DESC);

-- Index for session management queries

-- Index for active sessions by user
CREATE INDEX IF NOT EXISTS idx_sessions_user_active_time ON merchant_sessions 
USING btree (user_id, is_active, last_active DESC);

-- Index for session cleanup (expired sessions)
CREATE INDEX IF NOT EXISTS idx_sessions_cleanup ON merchant_sessions 
USING btree (is_active, last_active) WHERE is_active = true;

-- Additional compliance indexes

-- Index for compliance records by merchant and type
CREATE INDEX IF NOT EXISTS idx_compliance_merchant_type ON compliance_records 
USING btree (merchant_id, compliance_type, checked_at DESC);

-- Index for compliance records by status and expiration
CREATE INDEX IF NOT EXISTS idx_compliance_status_expiry ON compliance_records 
USING btree (status, expires_at) WHERE expires_at IS NOT NULL;

-- Index for compliance score ranges
CREATE INDEX IF NOT EXISTS idx_compliance_score_range ON compliance_records 
USING btree (score, status);

-- Additional analytics indexes

-- Index for analytics by merchant and calculation date
CREATE INDEX IF NOT EXISTS idx_analytics_merchant_calculated ON merchant_analytics 
USING btree (merchant_id, calculated_at DESC);

-- Index for analytics by risk score ranges
CREATE INDEX IF NOT EXISTS idx_analytics_risk_score_range ON merchant_analytics 
USING btree (risk_score, compliance_score);

-- Index for analytics flags
CREATE INDEX IF NOT EXISTS idx_analytics_flags ON merchant_analytics 
USING gin (flags);

-- Additional notification indexes

-- Index for notifications by user and read status
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON merchant_notifications 
USING btree (user_id, is_read, created_at DESC);

-- Index for notifications by merchant and type
CREATE INDEX IF NOT EXISTS idx_notifications_merchant_type ON merchant_notifications 
USING btree (merchant_id, type, created_at DESC);

-- Index for notifications by priority and read status
CREATE INDEX IF NOT EXISTS idx_notifications_priority_read ON merchant_notifications 
USING btree (priority, is_read, created_at DESC);

-- Additional bulk operation indexes

-- Index for bulk operations by user and status
CREATE INDEX IF NOT EXISTS idx_bulk_ops_user_status ON bulk_operations 
USING btree (user_id, status, started_at DESC);

-- Index for bulk operations by type and status
CREATE INDEX IF NOT EXISTS idx_bulk_ops_type_status ON bulk_operations 
USING btree (operation_type, status, started_at DESC);

-- Index for bulk operation items by operation and status
CREATE INDEX IF NOT EXISTS idx_bulk_items_operation_status ON bulk_operation_items 
USING btree (bulk_operation_id, status, processed_at);

-- Additional comparison indexes

-- Index for comparisons by user and creation date
CREATE INDEX IF NOT EXISTS idx_comparisons_user_created ON merchant_comparisons 
USING btree (user_id, created_at DESC);

-- Index for comparisons by merchant pairs
CREATE INDEX IF NOT EXISTS idx_comparisons_merchant_pairs ON merchant_comparisons 
USING btree (merchant1_id, merchant2_id);

-- Partial indexes for common query patterns

-- Index for active merchants only
CREATE INDEX IF NOT EXISTS idx_merchants_active_only ON merchants 
USING btree (portfolio_type_id, risk_level_id, created_at DESC) 
WHERE status = 'active';

-- Index for high-risk merchants only
CREATE INDEX IF NOT EXISTS idx_merchants_high_risk ON merchants 
USING btree (portfolio_type_id, compliance_status, created_at DESC) 
WHERE risk_level_id = (SELECT id FROM risk_levels WHERE level = 'high');

-- Index for pending compliance merchants
CREATE INDEX IF NOT EXISTS idx_merchants_pending_compliance ON merchants 
USING btree (portfolio_type_id, risk_level_id, created_at DESC) 
WHERE compliance_status = 'pending';

-- Index for recent audit logs (last 30 days)
CREATE INDEX IF NOT EXISTS idx_audit_logs_recent ON merchant_audit_logs 
USING btree (merchant_id, action, created_at DESC) 
WHERE created_at >= CURRENT_DATE - INTERVAL '30 days';

-- Index for unread notifications
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON merchant_notifications 
USING btree (user_id, priority, created_at DESC) 
WHERE is_read = false;

-- Index for active bulk operations
CREATE INDEX IF NOT EXISTS idx_bulk_ops_active ON bulk_operations 
USING btree (user_id, operation_type, started_at DESC) 
WHERE status IN ('pending', 'processing');

-- Expression indexes for computed values

-- Index for merchant name length (for filtering by name length)
CREATE INDEX IF NOT EXISTS idx_merchants_name_length ON merchants 
USING btree (length(name));

-- Index for merchant age (days since creation)
CREATE INDEX IF NOT EXISTS idx_merchants_age ON merchants 
USING btree ((CURRENT_DATE - created_at::date));

-- Index for compliance score categories
CREATE INDEX IF NOT EXISTS idx_compliance_score_category ON compliance_records 
USING btree (
    CASE 
        WHEN score >= 0.8 THEN 'high'
        WHEN score >= 0.6 THEN 'medium'
        ELSE 'low'
    END
);

-- Index for risk level numeric ordering
CREATE INDEX IF NOT EXISTS idx_merchants_risk_numeric ON merchants 
USING btree (risk_level_id, portfolio_type_id) 
INCLUDE (id, name, compliance_status);

-- Covering indexes for common queries

-- Covering index for merchant list queries
CREATE INDEX IF NOT EXISTS idx_merchants_list_covering ON merchants 
USING btree (created_at DESC) 
INCLUDE (id, name, portfolio_type_id, risk_level_id, status, compliance_status);

-- Covering index for merchant search results
CREATE INDEX IF NOT EXISTS idx_merchants_search_covering ON merchants 
USING btree (portfolio_type_id, risk_level_id, status) 
INCLUDE (id, name, industry, compliance_status, created_at);

-- Covering index for audit log queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_covering ON merchant_audit_logs 
USING btree (merchant_id, created_at DESC) 
INCLUDE (id, user_id, action, resource_type, resource_id);

-- Covering index for session queries
CREATE INDEX IF NOT EXISTS idx_sessions_covering ON merchant_sessions 
USING btree (user_id, is_active, last_active DESC) 
INCLUDE (id, merchant_id, started_at);

-- Statistics and maintenance

-- Update table statistics for better query planning
ANALYZE merchants;
ANALYZE merchant_audit_logs;
ANALYZE merchant_sessions;
ANALYZE compliance_records;
ANALYZE merchant_analytics;
ANALYZE merchant_notifications;
ANALYZE bulk_operations;
ANALYZE bulk_operation_items;
ANALYZE merchant_comparisons;

-- Add comments for documentation
COMMENT ON INDEX idx_merchants_search_composite IS 'Composite index for merchant search with multiple filters';
COMMENT ON INDEX idx_merchants_portfolio_risk_status IS 'Index for portfolio type + risk level + status combination';
COMMENT ON INDEX idx_merchants_industry_business_type IS 'Index for industry + business type filtering';
COMMENT ON INDEX idx_merchants_address_country_state IS 'Index for address-based searches';
COMMENT ON INDEX idx_merchants_contact_composite IS 'Index for contact information searches';
COMMENT ON INDEX idx_merchants_financial_composite IS 'Index for financial filtering';
COMMENT ON INDEX idx_merchants_date_range IS 'Index for date range queries';
COMMENT ON INDEX idx_merchants_user_created IS 'Index for user-specific merchant queries';
COMMENT ON INDEX idx_audit_logs_user_time_range IS 'Index for audit logs by user and time range';
COMMENT ON INDEX idx_audit_logs_merchant_action IS 'Index for audit logs by merchant and action type';
COMMENT ON INDEX idx_audit_logs_resource IS 'Index for audit logs by resource type and ID';
COMMENT ON INDEX idx_sessions_user_active_time IS 'Index for active sessions by user';
COMMENT ON INDEX idx_sessions_cleanup IS 'Index for session cleanup (expired sessions)';
COMMENT ON INDEX idx_compliance_merchant_type IS 'Index for compliance records by merchant and type';
COMMENT ON INDEX idx_compliance_status_expiry IS 'Index for compliance records by status and expiration';
COMMENT ON INDEX idx_compliance_score_range IS 'Index for compliance score ranges';
COMMENT ON INDEX idx_analytics_merchant_calculated IS 'Index for analytics by merchant and calculation date';
COMMENT ON INDEX idx_analytics_risk_score_range IS 'Index for analytics by risk score ranges';
COMMENT ON INDEX idx_analytics_flags IS 'Index for analytics flags using GIN';
COMMENT ON INDEX idx_notifications_user_read IS 'Index for notifications by user and read status';
COMMENT ON INDEX idx_notifications_merchant_type IS 'Index for notifications by merchant and type';
COMMENT ON INDEX idx_notifications_priority_read IS 'Index for notifications by priority and read status';
COMMENT ON INDEX idx_bulk_ops_user_status IS 'Index for bulk operations by user and status';
COMMENT ON INDEX idx_bulk_ops_type_status IS 'Index for bulk operations by type and status';
COMMENT ON INDEX idx_bulk_items_operation_status IS 'Index for bulk operation items by operation and status';
COMMENT ON INDEX idx_comparisons_user_created IS 'Index for comparisons by user and creation date';
COMMENT ON INDEX idx_comparisons_merchant_pairs IS 'Index for comparisons by merchant pairs';
COMMENT ON INDEX idx_merchants_active_only IS 'Partial index for active merchants only';
COMMENT ON INDEX idx_merchants_high_risk IS 'Partial index for high-risk merchants only';
COMMENT ON INDEX idx_merchants_pending_compliance IS 'Partial index for pending compliance merchants';
COMMENT ON INDEX idx_audit_logs_recent IS 'Partial index for recent audit logs (last 30 days)';
COMMENT ON INDEX idx_notifications_unread IS 'Partial index for unread notifications';
COMMENT ON INDEX idx_bulk_ops_active IS 'Partial index for active bulk operations';
COMMENT ON INDEX idx_merchants_name_length IS 'Expression index for merchant name length';
COMMENT ON INDEX idx_merchants_age IS 'Expression index for merchant age (days since creation)';
COMMENT ON INDEX idx_compliance_score_category IS 'Expression index for compliance score categories';
COMMENT ON INDEX idx_merchants_risk_numeric IS 'Index for risk level numeric ordering with covering columns';
COMMENT ON INDEX idx_merchants_list_covering IS 'Covering index for merchant list queries';
COMMENT ON INDEX idx_merchants_search_covering IS 'Covering index for merchant search results';
COMMENT ON INDEX idx_audit_logs_covering IS 'Covering index for audit log queries';
COMMENT ON INDEX idx_sessions_covering IS 'Covering index for session queries';
