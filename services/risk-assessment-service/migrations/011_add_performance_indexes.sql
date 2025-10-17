-- Migration: Add Performance Indexes
-- Description: Add optimized indexes for improved query performance
-- Version: 011
-- Date: 2024-01-15

-- Enable pg_stat_statements extension for query monitoring
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Risk Assessments Table Indexes

-- Index for business lookup with date filtering (most common query pattern)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_business_created 
ON risk_assessments (business_id, created_at DESC);

-- Index for risk level and industry filtering (dashboard queries)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_risk_industry 
ON risk_assessments (risk_level, industry);

-- Index for status filtering (active assessments)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_status 
ON risk_assessments (status) 
WHERE status IN ('pending', 'completed', 'failed');

-- Index for organization-based queries (multi-tenant)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_organization 
ON risk_assessments (organization_id, created_at DESC);

-- Index for risk score range queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_risk_score 
ON risk_assessments (risk_score) 
WHERE risk_score IS NOT NULL;

-- Index for industry-specific queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_industry_created 
ON risk_assessments (industry, created_at DESC);

-- Index for country-based filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_country 
ON risk_assessments (country);

-- Composite index for complex filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_complex_filter 
ON risk_assessments (organization_id, risk_level, industry, created_at DESC);

-- Batch Jobs Table Indexes

-- Index for job processing by status and creation time (highest priority)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_status_created 
ON batch_jobs (status, created_at ASC) 
WHERE status IN ('pending', 'processing');

-- Index for organization-specific job queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_organization_status 
ON batch_jobs (organization_id, status, created_at DESC);

-- Index for job priority scheduling
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_priority_created 
ON batch_jobs (priority DESC, created_at ASC) 
WHERE status = 'pending';

-- Index for job type filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_type_status 
ON batch_jobs (job_type, status);

-- Index for completed jobs cleanup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_completed_created 
ON batch_jobs (created_at) 
WHERE status = 'completed';

-- Index for failed jobs retry
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_failed_retry 
ON batch_jobs (status, retry_count, created_at) 
WHERE status = 'failed' AND retry_count < max_retries;

-- Custom Models Table Indexes

-- Index for active models per organization (most critical)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_models_org_active 
ON custom_models (organization_id, is_active) 
WHERE is_active = true;

-- Index for model type filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_models_type_active 
ON custom_models (model_type, is_active);

-- Index for model versioning and history
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_models_created 
ON custom_models (created_at DESC);

-- Index for model name searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_models_name 
ON custom_models (name);

-- Index for model performance tracking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_models_performance 
ON custom_models (organization_id, model_type, created_at DESC);

-- Batch Results Table Indexes

-- Index for job result lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_results_job_id 
ON batch_results (job_id, request_index);

-- Index for result status filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_results_status 
ON batch_results (status, processed_at DESC);

-- Index for error analysis
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_results_errors 
ON batch_results (status, processed_at) 
WHERE status = 'failed';

-- Risk Factors Table Indexes

-- Index for risk factor lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_factors_assessment 
ON risk_factors (risk_assessment_id, factor_type);

-- Index for factor type analysis
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_factors_type_score 
ON risk_factors (factor_type, score DESC);

-- Index for high-risk factors
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_factors_high_risk 
ON risk_factors (risk_assessment_id, score DESC) 
WHERE score > 0.7;

-- External Data Table Indexes

-- Index for external data lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_external_data_business 
ON external_data (business_id, data_source, created_at DESC);

-- Index for data source filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_external_data_source 
ON external_data (data_source, created_at DESC);

-- Index for data freshness
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_external_data_freshness 
ON external_data (created_at DESC) 
WHERE created_at > NOW() - INTERVAL '30 days';

-- Audit Logs Table Indexes

-- Index for audit trail queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_entity 
ON audit_logs (entity_type, entity_id, created_at DESC);

-- Index for user activity tracking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user 
ON audit_logs (user_id, created_at DESC);

-- Index for action filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_action 
ON audit_logs (action, created_at DESC);

-- Index for compliance reporting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_compliance 
ON audit_logs (entity_type, action, created_at DESC) 
WHERE action IN ('create', 'update', 'delete');

-- Performance Monitoring Table Indexes

-- Index for performance metrics
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_metrics_time 
ON performance_metrics (metric_name, timestamp DESC);

-- Index for alert queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_metrics_alerts 
ON performance_metrics (metric_name, value, timestamp DESC) 
WHERE value > threshold_value;

-- Index for trend analysis
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_metrics_trends 
ON performance_metrics (metric_name, timestamp) 
WHERE timestamp > NOW() - INTERVAL '7 days';

-- Update table statistics for better query planning
ANALYZE risk_assessments;
ANALYZE batch_jobs;
ANALYZE custom_models;
ANALYZE batch_results;
ANALYZE risk_factors;
ANALYZE external_data;
ANALYZE audit_logs;
ANALYZE performance_metrics;

-- Create a function to monitor index usage
CREATE OR REPLACE FUNCTION get_index_usage_stats()
RETURNS TABLE (
    schemaname text,
    tablename text,
    indexname text,
    idx_tup_read bigint,
    idx_tup_fetch bigint,
    idx_scan bigint,
    size_bytes bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.schemaname,
        s.tablename,
        s.indexrelname as indexname,
        s.idx_tup_read,
        s.idx_tup_fetch,
        s.idx_scan,
        pg_relation_size(s.indexrelid) as size_bytes
    FROM pg_stat_user_indexes s
    WHERE s.schemaname = 'public'
    ORDER BY s.idx_scan DESC;
END;
$$ LANGUAGE plpgsql;

-- Create a function to identify unused indexes
CREATE OR REPLACE FUNCTION get_unused_indexes()
RETURNS TABLE (
    schemaname text,
    tablename text,
    indexname text,
    size_bytes bigint,
    definition text
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.schemaname,
        s.tablename,
        s.indexrelname as indexname,
        pg_relation_size(s.indexrelid) as size_bytes,
        i.indexdef as definition
    FROM pg_stat_user_indexes s
    JOIN pg_indexes i ON i.indexname = s.indexrelname
    WHERE s.schemaname = 'public'
    AND s.idx_scan = 0
    AND i.indexdef NOT LIKE '%PRIMARY KEY%'
    AND i.indexdef NOT LIKE '%UNIQUE%'
    ORDER BY pg_relation_size(s.indexrelid) DESC;
END;
$$ LANGUAGE plpgsql;

-- Create a function to get slow query statistics
CREATE OR REPLACE FUNCTION get_slow_queries(threshold_ms integer DEFAULT 1000)
RETURNS TABLE (
    query text,
    calls bigint,
    total_time double precision,
    mean_time double precision,
    rows bigint,
    shared_blks_hit bigint,
    shared_blks_read bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pss.query,
        pss.calls,
        pss.total_exec_time,
        pss.mean_exec_time,
        pss.rows,
        pss.shared_blks_hit,
        pss.shared_blks_read
    FROM pg_stat_statements pss
    WHERE pss.mean_exec_time > threshold_ms
    ORDER BY pss.mean_exec_time DESC
    LIMIT 50;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON INDEX idx_risk_assessments_business_created IS 'Optimizes business lookup queries with date filtering - most common query pattern';
COMMENT ON INDEX idx_risk_assessments_risk_industry IS 'Optimizes risk filtering by level and industry for dashboard queries';
COMMENT ON INDEX idx_risk_assessments_status IS 'Optimizes queries filtering by assessment status';
COMMENT ON INDEX idx_risk_assessments_organization IS 'Optimizes multi-tenant data isolation queries';
COMMENT ON INDEX idx_batch_jobs_status_created IS 'Optimizes job processing queries by status and creation time - highest priority';
COMMENT ON INDEX idx_batch_jobs_organization_status IS 'Optimizes organization-specific job queries';
COMMENT ON INDEX idx_custom_models_org_active IS 'Optimizes active model queries per organization - most critical';
COMMENT ON INDEX idx_custom_models_type_active IS 'Optimizes model queries by type and status';

-- Grant permissions for monitoring functions
GRANT EXECUTE ON FUNCTION get_index_usage_stats() TO risk_assessment_service;
GRANT EXECUTE ON FUNCTION get_unused_indexes() TO risk_assessment_service;
GRANT EXECUTE ON FUNCTION get_slow_queries(integer) TO risk_assessment_service;
