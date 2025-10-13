-- Performance optimization indexes for 10K concurrent users
-- This migration adds indexes to optimize query performance

-- Risk Assessment indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_tenant_id_created_at 
ON risk_assessments(tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_business_id_status 
ON risk_assessments(business_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_risk_score_created_at 
ON risk_assessments(risk_score, created_at DESC) 
WHERE risk_score IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_industry_created_at 
ON risk_assessments(industry, created_at DESC) 
WHERE industry IS NOT NULL;

-- Batch Job indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_tenant_id_status 
ON batch_jobs(tenant_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_status_created_at 
ON batch_jobs(status, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_tenant_id_created_at 
ON batch_jobs(tenant_id, created_at DESC);

-- Batch Job Results indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_job_results_job_id_status 
ON batch_job_results(job_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_job_results_job_id_created_at 
ON batch_job_results(job_id, created_at DESC);

-- Custom Risk Models indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_risk_models_tenant_id_status 
ON custom_risk_models(tenant_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_risk_models_tenant_id_created_at 
ON custom_risk_models(tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_risk_models_base_model_tenant_id 
ON custom_risk_models(base_model, tenant_id);

-- Dashboards indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dashboards_tenant_id_type 
ON dashboards(tenant_id, type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dashboards_tenant_id_created_at 
ON dashboards(tenant_id, created_at DESC);

-- Reports indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reports_tenant_id_type_status 
ON reports(tenant_id, type, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reports_tenant_id_created_at 
ON reports(tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reports_scheduled_next_run 
ON reports(scheduled_next_run) 
WHERE scheduled_next_run IS NOT NULL;

-- Report Templates indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_report_templates_tenant_id_type 
ON report_templates(tenant_id, type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_report_templates_tenant_id_created_at 
ON report_templates(tenant_id, created_at DESC);

-- BI Gateway Config indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_bi_gateway_configs_tenant_id_type 
ON bi_gateway_configs(tenant_id, type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_bi_gateway_configs_tenant_id_status 
ON bi_gateway_configs(tenant_id, status);

-- Webhooks indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhooks_tenant_id_status 
ON webhooks(tenant_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhooks_tenant_id_created_at 
ON webhooks(tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhooks_events_gin 
ON webhooks USING GIN(events);

-- Webhook Deliveries indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhook_deliveries_webhook_id_status 
ON webhook_deliveries(webhook_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhook_deliveries_webhook_id_created_at 
ON webhook_deliveries(webhook_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhook_deliveries_status_created_at 
ON webhook_deliveries(status, created_at DESC);

-- Webhook Templates indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhook_templates_tenant_id_type 
ON webhook_templates(tenant_id, type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhook_templates_tenant_id_created_at 
ON webhook_templates(tenant_id, created_at DESC);

-- Webhook Rate Limiter States indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhook_rate_limiter_states_webhook_id_updated_at 
ON webhook_rate_limiter_states(webhook_id, updated_at DESC);

-- Webhook Circuit Breaker States indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhook_circuit_breaker_states_webhook_id_updated_at 
ON webhook_circuit_breaker_states(webhook_id, updated_at DESC);

-- Composite indexes for common query patterns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_tenant_industry_score 
ON risk_assessments(tenant_id, industry, risk_score) 
WHERE risk_score IS NOT NULL AND industry IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_tenant_status_created 
ON batch_jobs(tenant_id, status, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhook_deliveries_webhook_status_created 
ON webhook_deliveries(webhook_id, status, created_at DESC);

-- Partial indexes for active records
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_active 
ON risk_assessments(tenant_id, created_at DESC) 
WHERE status = 'completed';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_batch_jobs_active 
ON batch_jobs(tenant_id, created_at DESC) 
WHERE status IN ('pending', 'processing');

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhooks_active 
ON webhooks(tenant_id, created_at DESC) 
WHERE status = 'active';

-- Text search indexes for full-text search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_assessments_business_name_gin 
ON risk_assessments USING GIN(to_tsvector('english', business_name));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_risk_models_name_gin 
ON custom_risk_models USING GIN(to_tsvector('english', name));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dashboards_name_gin 
ON dashboards USING GIN(to_tsvector('english', name));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reports_name_gin 
ON reports USING GIN(to_tsvector('english', name));

-- Statistics update for query planner
ANALYZE risk_assessments;
ANALYZE batch_jobs;
ANALYZE batch_job_results;
ANALYZE custom_risk_models;
ANALYZE dashboards;
ANALYZE reports;
ANALYZE report_templates;
ANALYZE bi_gateway_configs;
ANALYZE webhooks;
ANALYZE webhook_deliveries;
ANALYZE webhook_templates;
ANALYZE webhook_rate_limiter_states;
ANALYZE webhook_circuit_breaker_states;

-- Update table statistics
UPDATE pg_stat_user_tables 
SET n_tup_ins = 0, n_tup_upd = 0, n_tup_del = 0 
WHERE schemaname = 'public';

-- Create function to refresh statistics
CREATE OR REPLACE FUNCTION refresh_table_statistics()
RETURNS void AS $$
BEGIN
    ANALYZE risk_assessments;
    ANALYZE batch_jobs;
    ANALYZE batch_job_results;
    ANALYZE custom_risk_models;
    ANALYZE dashboards;
    ANALYZE reports;
    ANALYZE report_templates;
    ANALYZE bi_gateway_configs;
    ANALYZE webhooks;
    ANALYZE webhook_deliveries;
    ANALYZE webhook_templates;
    ANALYZE webhook_rate_limiter_states;
    ANALYZE webhook_circuit_breaker_states;
END;
$$ LANGUAGE plpgsql;

-- Create function to get index usage statistics
CREATE OR REPLACE FUNCTION get_index_usage_stats()
RETURNS TABLE(
    schemaname text,
    tablename text,
    indexname text,
    idx_tup_read bigint,
    idx_tup_fetch bigint,
    idx_scan bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.schemaname,
        s.tablename,
        s.indexname,
        s.idx_tup_read,
        s.idx_tup_fetch,
        s.idx_scan
    FROM pg_stat_user_indexes s
    WHERE s.schemaname = 'public'
    ORDER BY s.idx_scan DESC;
END;
$$ LANGUAGE plpgsql;

-- Create function to identify unused indexes
CREATE OR REPLACE FUNCTION get_unused_indexes()
RETURNS TABLE(
    schemaname text,
    tablename text,
    indexname text,
    indexsize text
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.schemaname,
        s.tablename,
        s.indexname,
        pg_size_pretty(pg_relation_size(s.indexrelid)) as indexsize
    FROM pg_stat_user_indexes s
    WHERE s.schemaname = 'public'
    AND s.idx_scan = 0
    AND s.indexname NOT LIKE '%_pkey'
    ORDER BY pg_relation_size(s.indexrelid) DESC;
END;
$$ LANGUAGE plpgsql;

-- Create function to get slow queries (requires pg_stat_statements extension)
CREATE OR REPLACE FUNCTION get_slow_queries(limit_count integer DEFAULT 10)
RETURNS TABLE(
    query text,
    calls bigint,
    total_time double precision,
    mean_time double precision,
    rows bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pss.query,
        pss.calls,
        pss.total_time,
        pss.mean_time,
        pss.rows
    FROM pg_stat_statements pss
    WHERE pss.query NOT LIKE '%pg_stat_statements%'
    ORDER BY pss.mean_time DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON INDEX idx_risk_assessments_tenant_id_created_at IS 'Optimizes queries filtering by tenant and ordering by creation date';
COMMENT ON INDEX idx_risk_assessments_business_id_status IS 'Optimizes queries filtering by business ID and status';
COMMENT ON INDEX idx_risk_assessments_risk_score_created_at IS 'Optimizes queries filtering by risk score and ordering by creation date';
COMMENT ON INDEX idx_batch_jobs_tenant_id_status IS 'Optimizes queries filtering batch jobs by tenant and status';
COMMENT ON INDEX idx_webhooks_tenant_id_status IS 'Optimizes queries filtering webhooks by tenant and status';
COMMENT ON INDEX idx_webhook_deliveries_webhook_id_status IS 'Optimizes queries filtering webhook deliveries by webhook ID and status';

-- Create view for performance monitoring
CREATE OR REPLACE VIEW performance_monitoring AS
SELECT 
    'risk_assessments' as table_name,
    COUNT(*) as total_rows,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 hour') as rows_last_hour,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 day') as rows_last_day
FROM risk_assessments
UNION ALL
SELECT 
    'batch_jobs' as table_name,
    COUNT(*) as total_rows,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 hour') as rows_last_hour,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 day') as rows_last_day
FROM batch_jobs
UNION ALL
SELECT 
    'webhooks' as table_name,
    COUNT(*) as total_rows,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 hour') as rows_last_hour,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 day') as rows_last_day
FROM webhooks;

COMMENT ON VIEW performance_monitoring IS 'Provides performance monitoring data for key tables';
