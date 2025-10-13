-- Migration: Create business intelligence tables
-- Description: Creates tables for BI data synchronization, queries, dashboards, and exports

-- Create data syncs table
CREATE TABLE IF NOT EXISTS data_syncs (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    data_source_type VARCHAR(50) NOT NULL CHECK (data_source_type IN ('risk_assessment', 'batch_job', 'report', 'dashboard', 'custom_model', 'webhook', 'performance')),
    source_config JSONB NOT NULL,
    destination_config JSONB NOT NULL,
    sync_schedule JSONB NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'syncing', 'completed', 'failed', 'paused')),
    last_sync_at TIMESTAMP WITH TIME ZONE,
    next_sync_at TIMESTAMP WITH TIME ZONE,
    records_synced BIGINT DEFAULT 0,
    records_failed BIGINT DEFAULT 0,
    error TEXT,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    
    -- Indexes for performance
    INDEX idx_data_syncs_tenant_id (tenant_id),
    INDEX idx_data_syncs_data_source_type (data_source_type),
    INDEX idx_data_syncs_status (status),
    INDEX idx_data_syncs_next_sync_at (next_sync_at),
    INDEX idx_data_syncs_created_by (created_by),
    INDEX idx_data_syncs_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_data_syncs_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create data exports table
CREATE TABLE IF NOT EXISTS data_exports (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    data_source_type VARCHAR(50) NOT NULL CHECK (data_source_type IN ('risk_assessment', 'batch_job', 'report', 'dashboard', 'custom_model', 'webhook', 'performance')),
    source_config JSONB NOT NULL,
    format VARCHAR(20) NOT NULL CHECK (format IN ('json', 'csv', 'excel', 'parquet', 'avro')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'exporting', 'completed', 'failed', 'expired')),
    file_size BIGINT DEFAULT 0,
    download_url TEXT,
    records_exported BIGINT DEFAULT 0,
    error TEXT,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_data_exports_tenant_id (tenant_id),
    INDEX idx_data_exports_data_source_type (data_source_type),
    INDEX idx_data_exports_format (format),
    INDEX idx_data_exports_status (status),
    INDEX idx_data_exports_created_by (created_by),
    INDEX idx_data_exports_created_at (created_at),
    INDEX idx_data_exports_expires_at (expires_at),
    
    -- Constraints
    CONSTRAINT fk_data_exports_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI queries table
CREATE TABLE IF NOT EXISTS bi_queries (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query JSONB NOT NULL,
    parameters JSONB DEFAULT '[]',
    is_public BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_bi_queries_tenant_id (tenant_id),
    INDEX idx_bi_queries_is_public (is_public),
    INDEX idx_bi_queries_created_by (created_by),
    INDEX idx_bi_queries_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_bi_queries_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI dashboards table
CREATE TABLE IF NOT EXISTS bi_dashboards (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    layout JSONB NOT NULL,
    widgets JSONB NOT NULL,
    filters JSONB DEFAULT '[]',
    is_public BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_bi_dashboards_tenant_id (tenant_id),
    INDEX idx_bi_dashboards_is_public (is_public),
    INDEX idx_bi_dashboards_is_default (is_default),
    INDEX idx_bi_dashboards_created_by (created_by),
    INDEX idx_bi_dashboards_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_bi_dashboards_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI query results table for caching
CREATE TABLE IF NOT EXISTS bi_query_results (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    query_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    parameters JSONB,
    result_data JSONB NOT NULL,
    columns JSONB NOT NULL,
    row_count BIGINT NOT NULL,
    execution_time_ms INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Indexes
    INDEX idx_bi_query_results_query_id (query_id),
    INDEX idx_bi_query_results_tenant_id (tenant_id),
    INDEX idx_bi_query_results_created_at (created_at),
    INDEX idx_bi_query_results_expires_at (expires_at),
    
    -- Constraints
    CONSTRAINT fk_bi_query_results_query FOREIGN KEY (query_id) REFERENCES bi_queries(id) ON DELETE CASCADE,
    CONSTRAINT fk_bi_query_results_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI dashboard views table for tracking usage
CREATE TABLE IF NOT EXISTS bi_dashboard_views (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    dashboard_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    session_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_bi_dashboard_views_dashboard_id (dashboard_id),
    INDEX idx_bi_dashboard_views_tenant_id (tenant_id),
    INDEX idx_bi_dashboard_views_user_id (user_id),
    INDEX idx_bi_dashboard_views_viewed_at (viewed_at),
    
    -- Constraints
    CONSTRAINT fk_bi_dashboard_views_dashboard FOREIGN KEY (dashboard_id) REFERENCES bi_dashboards(id) ON DELETE CASCADE,
    CONSTRAINT fk_bi_dashboard_views_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI query executions table for tracking usage
CREATE TABLE IF NOT EXISTS bi_query_executions (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    query_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    parameters JSONB,
    execution_time_ms INTEGER NOT NULL,
    row_count BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'error', 'timeout')),
    error_message TEXT,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_bi_query_executions_query_id (query_id),
    INDEX idx_bi_query_executions_tenant_id (tenant_id),
    INDEX idx_bi_query_executions_user_id (user_id),
    INDEX idx_bi_query_executions_executed_at (executed_at),
    INDEX idx_bi_query_executions_status (status),
    
    -- Constraints
    CONSTRAINT fk_bi_query_executions_query FOREIGN KEY (query_id) REFERENCES bi_queries(id) ON DELETE CASCADE,
    CONSTRAINT fk_bi_query_executions_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI data sources table for managing data source configurations
CREATE TABLE IF NOT EXISTS bi_data_sources (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('risk_assessment', 'batch_job', 'report', 'dashboard', 'custom_model', 'webhook', 'performance')),
    config JSONB NOT NULL,
    schema JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_bi_data_sources_tenant_id (tenant_id),
    INDEX idx_bi_data_sources_type (type),
    INDEX idx_bi_data_sources_is_active (is_active),
    INDEX idx_bi_data_sources_created_by (created_by),
    INDEX idx_bi_data_sources_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_bi_data_sources_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI sync logs table for detailed sync tracking
CREATE TABLE IF NOT EXISTS bi_sync_logs (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    sync_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('started', 'in_progress', 'completed', 'failed', 'cancelled')),
    records_processed BIGINT DEFAULT 0,
    records_synced BIGINT DEFAULT 0,
    records_failed BIGINT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_bi_sync_logs_sync_id (sync_id),
    INDEX idx_bi_sync_logs_tenant_id (tenant_id),
    INDEX idx_bi_sync_logs_status (status),
    INDEX idx_bi_sync_logs_started_at (started_at),
    
    -- Constraints
    CONSTRAINT fk_bi_sync_logs_sync FOREIGN KEY (sync_id) REFERENCES data_syncs(id) ON DELETE CASCADE,
    CONSTRAINT fk_bi_sync_logs_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI export logs table for detailed export tracking
CREATE TABLE IF NOT EXISTS bi_export_logs (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    export_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('started', 'in_progress', 'completed', 'failed', 'cancelled')),
    records_processed BIGINT DEFAULT 0,
    records_exported BIGINT DEFAULT 0,
    file_size BIGINT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_bi_export_logs_export_id (export_id),
    INDEX idx_bi_export_logs_tenant_id (tenant_id),
    INDEX idx_bi_export_logs_status (status),
    INDEX idx_bi_export_logs_started_at (started_at),
    
    -- Constraints
    CONSTRAINT fk_bi_export_logs_export FOREIGN KEY (export_id) REFERENCES data_exports(id) ON DELETE CASCADE,
    CONSTRAINT fk_bi_export_logs_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create BI favorites table
CREATE TABLE IF NOT EXISTS bi_favorites (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(20) NOT NULL CHECK (resource_type IN ('query', 'dashboard', 'data_sync', 'data_export')),
    resource_id VARCHAR(255) NOT NULL,
    favorited_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_bi_favorites_tenant_id (tenant_id),
    INDEX idx_bi_favorites_user_id (user_id),
    INDEX idx_bi_favorites_resource_type (resource_type),
    INDEX idx_bi_favorites_resource_id (resource_id),
    
    -- Constraints
    CONSTRAINT fk_bi_favorites_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(tenant_id, user_id, resource_type, resource_id)
);

-- Create BI shares table for sharing resources
CREATE TABLE IF NOT EXISTS bi_shares (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(20) NOT NULL CHECK (resource_type IN ('query', 'dashboard', 'data_sync', 'data_export')),
    resource_id VARCHAR(255) NOT NULL,
    shared_with_tenant_id VARCHAR(255),
    shared_with_user_id VARCHAR(255),
    permission VARCHAR(20) DEFAULT 'read' CHECK (permission IN ('read', 'write', 'admin')),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_bi_shares_tenant_id (tenant_id),
    INDEX idx_bi_shares_resource_type (resource_type),
    INDEX idx_bi_shares_resource_id (resource_id),
    INDEX idx_bi_shares_shared_with (shared_with_tenant_id, shared_with_user_id),
    INDEX idx_bi_shares_expires_at (expires_at),
    
    -- Constraints
    CONSTRAINT fk_bi_shares_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_bi_shares_shared_tenant FOREIGN KEY (shared_with_tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_data_syncs_tenant_type ON data_syncs(tenant_id, data_source_type);
CREATE INDEX IF NOT EXISTS idx_data_syncs_tenant_status ON data_syncs(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_data_syncs_tenant_created_by ON data_syncs(tenant_id, created_by);
CREATE INDEX IF NOT EXISTS idx_data_syncs_tenant_created_at ON data_syncs(tenant_id, created_at);

CREATE INDEX IF NOT EXISTS idx_data_exports_tenant_type ON data_exports(tenant_id, data_source_type);
CREATE INDEX IF NOT EXISTS idx_data_exports_tenant_status ON data_exports(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_data_exports_tenant_created_by ON data_exports(tenant_id, created_by);
CREATE INDEX IF NOT EXISTS idx_data_exports_tenant_created_at ON data_exports(tenant_id, created_at);

CREATE INDEX IF NOT EXISTS idx_bi_queries_tenant_public ON bi_queries(tenant_id, is_public);
CREATE INDEX IF NOT EXISTS idx_bi_queries_tenant_created_by ON bi_queries(tenant_id, created_by);

CREATE INDEX IF NOT EXISTS idx_bi_dashboards_tenant_public ON bi_dashboards(tenant_id, is_public);
CREATE INDEX IF NOT EXISTS idx_bi_dashboards_tenant_default ON bi_dashboards(tenant_id, is_default);
CREATE INDEX IF NOT EXISTS idx_bi_dashboards_tenant_created_by ON bi_dashboards(tenant_id, created_by);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_bi_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_data_syncs_updated_at
    BEFORE UPDATE ON data_syncs
    FOR EACH ROW
    EXECUTE FUNCTION update_bi_updated_at();

CREATE TRIGGER update_bi_queries_updated_at
    BEFORE UPDATE ON bi_queries
    FOR EACH ROW
    EXECUTE FUNCTION update_bi_updated_at();

CREATE TRIGGER update_bi_dashboards_updated_at
    BEFORE UPDATE ON bi_dashboards
    FOR EACH ROW
    EXECUTE FUNCTION update_bi_updated_at();

CREATE TRIGGER update_bi_data_sources_updated_at
    BEFORE UPDATE ON bi_data_sources
    FOR EACH ROW
    EXECUTE FUNCTION update_bi_updated_at();

-- Create function to clean up expired query results
CREATE OR REPLACE FUNCTION cleanup_expired_bi_query_results()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM bi_query_results 
    WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up expired data exports
CREATE OR REPLACE FUNCTION cleanup_expired_bi_data_exports()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM data_exports 
    WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up expired shares
CREATE OR REPLACE FUNCTION cleanup_expired_bi_shares()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM bi_shares 
    WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old sync logs (older than 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_bi_sync_logs()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM bi_sync_logs 
    WHERE started_at < CURRENT_TIMESTAMP - INTERVAL '30 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old export logs (older than 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_bi_export_logs()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM bi_export_logs 
    WHERE started_at < CURRENT_TIMESTAMP - INTERVAL '30 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old query executions (older than 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_bi_query_executions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM bi_query_executions 
    WHERE executed_at < CURRENT_TIMESTAMP - INTERVAL '90 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old dashboard views (older than 1 year)
CREATE OR REPLACE FUNCTION cleanup_old_bi_dashboard_views()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM bi_dashboard_views 
    WHERE viewed_at < CURRENT_TIMESTAMP - INTERVAL '1 year';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Insert sample BI data sources for testing
INSERT INTO bi_data_sources (
    id, tenant_id, name, type, is_active, created_by
) VALUES (
    'risk_assessments_source', 'default', 'Risk Assessments', 'risk_assessment', true, 'system'
), (
    'batch_jobs_source', 'default', 'Batch Jobs', 'batch_job', true, 'system'
), (
    'reports_source', 'default', 'Reports', 'report', true, 'system'
), (
    'dashboards_source', 'default', 'Dashboards', 'dashboard', true, 'system'
), (
    'performance_source', 'default', 'Performance Metrics', 'performance', true, 'system'
) ON CONFLICT (id) DO NOTHING;

-- Insert sample BI queries for testing
INSERT INTO bi_queries (
    id, tenant_id, name, description, is_public, is_default, created_by
) VALUES (
    'risk_overview_query', 'default', 'Risk Overview', 'High-level risk assessment overview with key metrics', true, true, 'system'
), (
    'batch_performance_query', 'default', 'Batch Performance', 'Batch job performance and completion statistics', true, true, 'system'
), (
    'report_usage_query', 'default', 'Report Usage', 'Report generation and usage analytics', true, true, 'system'
), (
    'dashboard_analytics_query', 'default', 'Dashboard Analytics', 'Dashboard view and interaction analytics', true, true, 'system'
) ON CONFLICT (id) DO NOTHING;

-- Insert sample BI dashboards for testing
INSERT INTO bi_dashboards (
    id, tenant_id, name, description, is_public, is_default, created_by
) VALUES (
    'executive_dashboard', 'default', 'Executive Dashboard', 'High-level executive dashboard with key business metrics', true, true, 'system'
), (
    'operational_dashboard', 'default', 'Operational Dashboard', 'Operational metrics and performance monitoring', true, true, 'system'
), (
    'analytics_dashboard', 'default', 'Analytics Dashboard', 'Detailed analytics and reporting dashboard', true, true, 'system'
) ON CONFLICT (id) DO NOTHING;

-- Add comments to tables for documentation
COMMENT ON TABLE data_syncs IS 'Stores data synchronization configurations and status';
COMMENT ON TABLE data_exports IS 'Stores data export jobs and results';
COMMENT ON TABLE bi_queries IS 'Stores business intelligence queries and configurations';
COMMENT ON TABLE bi_dashboards IS 'Stores business intelligence dashboards and layouts';
COMMENT ON TABLE bi_query_results IS 'Caches query results for performance optimization';
COMMENT ON TABLE bi_dashboard_views IS 'Tracks dashboard view statistics and user interactions';
COMMENT ON TABLE bi_query_executions IS 'Tracks query execution history and performance metrics';
COMMENT ON TABLE bi_data_sources IS 'Manages data source configurations and schemas';
COMMENT ON TABLE bi_sync_logs IS 'Detailed logs for data synchronization operations';
COMMENT ON TABLE bi_export_logs IS 'Detailed logs for data export operations';
COMMENT ON TABLE bi_favorites IS 'User favorites for BI resources';
COMMENT ON TABLE bi_shares IS 'Resource sharing permissions and access control';
