-- Migration: 008_database_performance_optimization.sql
-- Description: Database performance optimization with additional indexes and query improvements
-- Created: 2025-01-19
-- Dependencies: 005_merchant_portfolio_schema.sql

-- =============================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =============================================================================

-- Additional composite indexes for common query patterns
-- These indexes are designed to optimize the most frequent query patterns

-- Merchants table optimization indexes
-- Composite index for portfolio type + status + created_at (most common filter combination)
CREATE INDEX IF NOT EXISTS idx_merchants_portfolio_status_created 
    ON merchants(portfolio_type_id, status, created_at DESC);

-- Composite index for risk level + compliance status + created_at
CREATE INDEX IF NOT EXISTS idx_merchants_risk_compliance_created 
    ON merchants(risk_level_id, compliance_status, created_at DESC);

-- Composite index for industry + business_type + status
CREATE INDEX IF NOT EXISTS idx_merchants_industry_business_status 
    ON merchants(industry, business_type, status);

-- Composite index for employee_count + annual_revenue + status (for range queries)
CREATE INDEX IF NOT EXISTS idx_merchants_employee_revenue_status 
    ON merchants(employee_count, annual_revenue, status);

-- Composite index for address fields (for location-based queries)
CREATE INDEX IF NOT EXISTS idx_merchants_address_country_state 
    ON merchants(address_country, address_state, address_city);

-- Composite index for contact fields (for contact-based searches)
CREATE INDEX IF NOT EXISTS idx_merchants_contact_email_phone 
    ON merchants(contact_email, contact_phone);

-- Partial indexes for active merchants only (reduces index size)
CREATE INDEX IF NOT EXISTS idx_merchants_active_portfolio_created 
    ON merchants(portfolio_type_id, created_at DESC) 
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_merchants_active_risk_created 
    ON merchants(risk_level_id, created_at DESC) 
    WHERE status = 'active';

-- Text search optimization indexes
-- Full-text search index for merchant names and descriptions
CREATE INDEX IF NOT EXISTS idx_merchants_name_legal_fts 
    ON merchants USING gin(to_tsvector('english', name || ' ' || legal_name));

-- Trigram indexes for fuzzy text search
CREATE INDEX IF NOT EXISTS idx_merchants_industry_trgm 
    ON merchants USING gin(industry gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_merchants_business_type_trgm 
    ON merchants USING gin(business_type gin_trgm_ops);

-- =============================================================================
-- MERCHANT SESSIONS OPTIMIZATION
-- =============================================================================

-- Composite index for active sessions by user
CREATE INDEX IF NOT EXISTS idx_merchant_sessions_user_active_merchant 
    ON merchant_sessions(user_id, is_active, merchant_id) 
    WHERE is_active = true;

-- Index for session cleanup (expired sessions)
CREATE INDEX IF NOT EXISTS idx_merchant_sessions_last_active_cleanup 
    ON merchant_sessions(last_active) 
    WHERE is_active = true;

-- =============================================================================
-- AUDIT LOGS OPTIMIZATION
-- =============================================================================

-- Composite index for audit log queries by merchant and time range
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_merchant_created 
    ON merchant_audit_logs(merchant_id, created_at DESC);

-- Composite index for audit log queries by user and action
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_user_action_created 
    ON merchant_audit_logs(user_id, action, created_at DESC);

-- Partial index for recent audit logs (last 30 days)
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_recent 
    ON merchant_audit_logs(merchant_id, action, created_at DESC) 
    WHERE created_at >= (CURRENT_TIMESTAMP - INTERVAL '30 days');

-- =============================================================================
-- COMPLIANCE RECORDS OPTIMIZATION
-- =============================================================================

-- Composite index for compliance queries by merchant and type
CREATE INDEX IF NOT EXISTS idx_compliance_records_merchant_type_status 
    ON compliance_records(merchant_id, compliance_type, status);

-- Index for expired compliance checks
CREATE INDEX IF NOT EXISTS idx_compliance_records_expires_at_status 
    ON compliance_records(expires_at, status) 
    WHERE expires_at IS NOT NULL;

-- Composite index for compliance score ranges
CREATE INDEX IF NOT EXISTS idx_compliance_records_score_type_checked 
    ON compliance_records(score, compliance_type, checked_at DESC);

-- =============================================================================
-- MERCHANT ANALYTICS OPTIMIZATION
-- =============================================================================

-- Composite index for analytics queries by risk score ranges
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_risk_score_merchant 
    ON merchant_analytics(risk_score, merchant_id);

-- Composite index for analytics queries by compliance score ranges
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_compliance_score_merchant 
    ON merchant_analytics(compliance_score, merchant_id);

-- Index for recent analytics data
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_calculated_recent 
    ON merchant_analytics(calculated_at DESC, merchant_id) 
    WHERE calculated_at >= (CURRENT_TIMESTAMP - INTERVAL '7 days');

-- =============================================================================
-- BULK OPERATIONS OPTIMIZATION
-- =============================================================================

-- Composite index for bulk operations by user and status
CREATE INDEX IF NOT EXISTS idx_bulk_operations_user_status_started 
    ON bulk_operations(user_id, status, started_at DESC);

-- Index for active bulk operations
CREATE INDEX IF NOT EXISTS idx_bulk_operations_active 
    ON bulk_operations(started_at DESC) 
    WHERE status IN ('pending', 'processing');

-- Composite index for bulk operation items by operation and status
CREATE INDEX IF NOT EXISTS idx_bulk_operation_items_operation_status 
    ON bulk_operation_items(bulk_operation_id, status);

-- =============================================================================
-- NOTIFICATIONS OPTIMIZATION
-- =============================================================================

-- Composite index for unread notifications by user and priority
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_user_unread_priority 
    ON merchant_notifications(user_id, is_read, priority, created_at DESC) 
    WHERE is_read = false;

-- Index for notification cleanup (old notifications)
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_created_cleanup 
    ON merchant_notifications(created_at) 
    WHERE created_at < (CURRENT_TIMESTAMP - INTERVAL '90 days');

-- =============================================================================
-- QUERY OPTIMIZATION VIEWS
-- =============================================================================

-- Optimized view for merchant portfolio dashboard
CREATE OR REPLACE VIEW merchant_portfolio_dashboard AS
SELECT 
    m.id,
    m.name,
    m.legal_name,
    m.industry,
    m.business_type,
    pt.type as portfolio_type,
    rl.level as risk_level,
    m.compliance_status,
    m.status,
    m.created_at,
    m.updated_at,
    -- Analytics data
    ma.risk_score,
    ma.compliance_score,
    ma.transaction_volume,
    ma.last_activity,
    -- Notification counts
    COALESCE(unread_notifications.count, 0) as unread_notifications,
    -- Recent audit activity
    COALESCE(recent_audit.count, 0) as recent_audit_count
FROM merchants m
JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
JOIN risk_levels rl ON m.risk_level_id = rl.id
LEFT JOIN merchant_analytics ma ON m.id = ma.merchant_id
LEFT JOIN (
    SELECT merchant_id, COUNT(*) as count
    FROM merchant_notifications 
    WHERE is_read = false 
    GROUP BY merchant_id
) unread_notifications ON m.id = unread_notifications.merchant_id
LEFT JOIN (
    SELECT merchant_id, COUNT(*) as count
    FROM merchant_audit_logs 
    WHERE created_at >= (CURRENT_TIMESTAMP - INTERVAL '7 days')
    GROUP BY merchant_id
) recent_audit ON m.id = recent_audit.merchant_id
WHERE m.status = 'active';

-- Optimized view for merchant search results
CREATE OR REPLACE VIEW merchant_search_results AS
SELECT 
    m.id,
    m.name,
    m.legal_name,
    m.industry,
    m.business_type,
    pt.type as portfolio_type,
    rl.level as risk_level,
    m.compliance_status,
    m.status,
    m.created_at,
    -- Search relevance score (can be used for ranking)
    CASE 
        WHEN m.name ILIKE '%' || $1 || '%' THEN 3
        WHEN m.legal_name ILIKE '%' || $1 || '%' THEN 2
        WHEN m.industry ILIKE '%' || $1 || '%' THEN 1
        ELSE 0
    END as relevance_score
FROM merchants m
JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
JOIN risk_levels rl ON m.risk_level_id = rl.id
WHERE m.status = 'active'
ORDER BY relevance_score DESC, m.created_at DESC;

-- =============================================================================
-- QUERY PERFORMANCE FUNCTIONS
-- =============================================================================

-- Function to get merchant count by filters (optimized)
CREATE OR REPLACE FUNCTION get_merchant_count_by_filters(
    p_portfolio_type VARCHAR(50) DEFAULT NULL,
    p_risk_level VARCHAR(50) DEFAULT NULL,
    p_industry VARCHAR(100) DEFAULT NULL,
    p_status VARCHAR(50) DEFAULT 'active'
) RETURNS INTEGER AS $$
DECLARE
    count_result INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO count_result
    FROM merchants m
    JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
    JOIN risk_levels rl ON m.risk_level_id = rl.id
    WHERE m.status = p_status
    AND (p_portfolio_type IS NULL OR pt.type = p_portfolio_type)
    AND (p_risk_level IS NULL OR rl.level = p_risk_level)
    AND (p_industry IS NULL OR m.industry ILIKE '%' || p_industry || '%');
    
    RETURN count_result;
END;
$$ LANGUAGE plpgsql;

-- Function to get merchants with cursor-based pagination
CREATE OR REPLACE FUNCTION get_merchants_cursor_paginated(
    p_cursor TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    p_limit INTEGER DEFAULT 50,
    p_portfolio_type VARCHAR(50) DEFAULT NULL,
    p_risk_level VARCHAR(50) DEFAULT NULL,
    p_industry VARCHAR(100) DEFAULT NULL,
    p_status VARCHAR(50) DEFAULT 'active'
) RETURNS TABLE (
    id UUID,
    name VARCHAR(255),
    legal_name VARCHAR(255),
    industry VARCHAR(100),
    business_type VARCHAR(50),
    portfolio_type VARCHAR(50),
    risk_level VARCHAR(50),
    compliance_status VARCHAR(50),
    status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        m.id,
        m.name,
        m.legal_name,
        m.industry,
        m.business_type,
        pt.type as portfolio_type,
        rl.level as risk_level,
        m.compliance_status,
        m.status,
        m.created_at,
        m.updated_at
    FROM merchants m
    JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
    JOIN risk_levels rl ON m.risk_level_id = rl.id
    WHERE m.status = p_status
    AND (p_portfolio_type IS NULL OR pt.type = p_portfolio_type)
    AND (p_risk_level IS NULL OR rl.level = p_risk_level)
    AND (p_industry IS NULL OR m.industry ILIKE '%' || p_industry || '%')
    AND (p_cursor IS NULL OR m.created_at < p_cursor)
    ORDER BY m.created_at DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- DATABASE STATISTICS AND MONITORING
-- =============================================================================

-- Function to get query performance statistics
CREATE OR REPLACE FUNCTION get_query_performance_stats() 
RETURNS TABLE (
    table_name TEXT,
    index_name TEXT,
    index_size TEXT,
    index_usage_count BIGINT,
    last_used TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname||'.'||tablename as table_name,
        indexname as index_name,
        pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
        idx_tup_read as index_usage_count,
        last_used
    FROM pg_stat_user_indexes 
    JOIN pg_indexes ON pg_stat_user_indexes.indexrelname = pg_indexes.indexname
    ORDER BY idx_tup_read DESC;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- CLEANUP AND MAINTENANCE
-- =============================================================================

-- Function to clean up old audit logs (keep last 1 year)
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs() 
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM merchant_audit_logs 
    WHERE created_at < (CURRENT_TIMESTAMP - INTERVAL '1 year');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up old notifications (keep last 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_notifications() 
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM merchant_notifications 
    WHERE created_at < (CURRENT_TIMESTAMP - INTERVAL '90 days')
    AND is_read = true;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions() 
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE merchant_sessions 
    SET is_active = false 
    WHERE is_active = true 
    AND last_active < (CURRENT_TIMESTAMP - INTERVAL '24 hours');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- COMMENTS AND DOCUMENTATION
-- =============================================================================

COMMENT ON INDEX idx_merchants_portfolio_status_created IS 'Optimizes queries filtering by portfolio type, status, and ordering by created_at';
COMMENT ON INDEX idx_merchants_risk_compliance_created IS 'Optimizes queries filtering by risk level, compliance status, and ordering by created_at';
COMMENT ON INDEX idx_merchants_industry_business_status IS 'Optimizes queries filtering by industry and business type';
COMMENT ON INDEX idx_merchants_employee_revenue_status IS 'Optimizes range queries on employee count and annual revenue';
COMMENT ON INDEX idx_merchants_address_country_state IS 'Optimizes location-based queries';
COMMENT ON INDEX idx_merchants_contact_email_phone IS 'Optimizes contact-based searches';
COMMENT ON INDEX idx_merchants_active_portfolio_created IS 'Partial index for active merchants only, reduces index size';
COMMENT ON INDEX idx_merchants_name_legal_fts IS 'Full-text search index for merchant names';
COMMENT ON INDEX idx_merchants_industry_trgm IS 'Trigram index for fuzzy industry search';

COMMENT ON VIEW merchant_portfolio_dashboard IS 'Optimized view for merchant portfolio dashboard with aggregated data';
COMMENT ON VIEW merchant_search_results IS 'Optimized view for merchant search with relevance scoring';

COMMENT ON FUNCTION get_merchant_count_by_filters IS 'Optimized function to count merchants by common filter combinations';
COMMENT ON FUNCTION get_merchants_cursor_paginated IS 'Cursor-based pagination function for large merchant datasets';
COMMENT ON FUNCTION get_query_performance_stats IS 'Function to monitor query performance and index usage';
COMMENT ON FUNCTION cleanup_old_audit_logs IS 'Maintenance function to clean up old audit logs';
COMMENT ON FUNCTION cleanup_old_notifications IS 'Maintenance function to clean up old notifications';
COMMENT ON FUNCTION cleanup_expired_sessions IS 'Maintenance function to clean up expired sessions';

-- =============================================================================
-- UPDATE TABLE STATISTICS
-- =============================================================================

-- Update table statistics for better query planning
ANALYZE merchants;
ANALYZE portfolio_types;
ANALYZE risk_levels;
ANALYZE merchant_sessions;
ANALYZE merchant_audit_logs;
ANALYZE compliance_records;
ANALYZE merchant_analytics;
ANALYZE merchant_notifications;
ANALYZE bulk_operations;
ANALYZE bulk_operation_items;
