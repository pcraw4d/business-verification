-- ============================================================================
-- UNIFIED MONITORING SCHEMA ENHANCEMENT
-- ============================================================================
-- This file implements the missing tables for subtask 3.1.2:
-- - unified_performance_reports table
-- - performance_integration_health table
-- 
-- These tables extend the existing consolidated monitoring schema to provide
-- comprehensive reporting and integration health monitoring capabilities.
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- UNIFIED PERFORMANCE REPORTS TABLE
-- ============================================================================
-- This table stores generated performance reports and analytics
-- Supports both automated and manual report generation

CREATE TABLE IF NOT EXISTS unified_performance_reports (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Report identification
    report_name VARCHAR(200) NOT NULL,
    report_type VARCHAR(50) NOT NULL,           -- 'summary', 'detailed', 'trend', 'alert', 'custom'
    report_category VARCHAR(50) NOT NULL,       -- 'performance', 'resource', 'business', 'security'
    
    -- Report scope and filters
    component VARCHAR(100),                     -- Specific component (NULL for system-wide)
    service_name VARCHAR(100),                  -- Specific service (NULL for system-wide)
    time_range_start TIMESTAMP WITH TIME ZONE NOT NULL,
    time_range_end TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Report generation details
    generated_by UUID,                          -- User who generated the report
    generation_method VARCHAR(50) NOT NULL,     -- 'automated', 'manual', 'scheduled'
    report_frequency VARCHAR(20),               -- 'realtime', 'hourly', 'daily', 'weekly', 'monthly'
    
    -- Report content and metadata
    report_data JSONB NOT NULL,                 -- Main report data in structured format
    summary_data JSONB,                         -- Summary statistics and key metrics
    visualizations JSONB,                       -- Chart and graph configurations
    insights JSONB,                             -- AI-generated insights and recommendations
    
    -- Report status and delivery
    status VARCHAR(20) DEFAULT 'generated' NOT NULL, -- 'generating', 'generated', 'delivered', 'failed'
    delivery_method VARCHAR(50),                -- 'email', 'dashboard', 'api', 'file'
    delivery_status VARCHAR(20),                -- 'pending', 'sent', 'delivered', 'failed'
    delivery_recipients TEXT[],                 -- List of recipients
    
    -- Report configuration
    report_config JSONB,                        -- Configuration used to generate report
    filters_applied JSONB,                      -- Filters and parameters applied
    metrics_included TEXT[],                    -- List of metrics included in report
    
    -- Quality and validation
    data_quality_score DECIMAL(3,2),           -- Quality score of underlying data (0.0-1.0)
    completeness_score DECIMAL(3,2),           -- Completeness of data coverage (0.0-1.0)
    validation_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'validated', 'failed'
    
    -- File and storage information
    file_path VARCHAR(500),                     -- Path to stored report file
    file_size_bytes BIGINT,                     -- Size of report file
    file_format VARCHAR(20),                    -- 'json', 'pdf', 'csv', 'xlsx'
    
    -- Metadata and tags
    tags JSONB,
    metadata JSONB,
    
    -- Constraints
    CONSTRAINT valid_report_type CHECK (report_type IN ('summary', 'detailed', 'trend', 'alert', 'custom')),
    CONSTRAINT valid_report_category CHECK (report_category IN ('performance', 'resource', 'business', 'security')),
    CONSTRAINT valid_generation_method CHECK (generation_method IN ('automated', 'manual', 'scheduled')),
    CONSTRAINT valid_status CHECK (status IN ('generating', 'generated', 'delivered', 'failed')),
    CONSTRAINT valid_delivery_method CHECK (delivery_method IN ('email', 'dashboard', 'api', 'file') OR delivery_method IS NULL),
    CONSTRAINT valid_delivery_status CHECK (delivery_status IN ('pending', 'sent', 'delivered', 'failed') OR delivery_status IS NULL),
    CONSTRAINT valid_validation_status CHECK (validation_status IN ('pending', 'validated', 'failed')),
    CONSTRAINT valid_file_format CHECK (file_format IN ('json', 'pdf', 'csv', 'xlsx') OR file_format IS NULL),
    CONSTRAINT valid_time_range CHECK (time_range_end > time_range_start),
    CONSTRAINT valid_quality_scores CHECK (
        (data_quality_score IS NULL OR (data_quality_score >= 0.0 AND data_quality_score <= 1.0)) AND
        (completeness_score IS NULL OR (completeness_score >= 0.0 AND completeness_score <= 1.0))
    )
);

-- ============================================================================
-- PERFORMANCE INTEGRATION HEALTH TABLE
-- ============================================================================
-- This table monitors the health and performance of external integrations
-- and internal service connections

CREATE TABLE IF NOT EXISTS performance_integration_health (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Integration identification
    integration_name VARCHAR(100) NOT NULL,     -- Name of the integration
    integration_type VARCHAR(50) NOT NULL,      -- 'external_api', 'database', 'cache', 'queue', 'storage'
    integration_category VARCHAR(50) NOT NULL,  -- 'classification', 'verification', 'monitoring', 'notification'
    
    -- Service and endpoint details
    service_name VARCHAR(100) NOT NULL,         -- Service using the integration
    endpoint_url VARCHAR(500),                  -- Integration endpoint URL
    connection_string VARCHAR(500),             -- Connection details (encrypted)
    
    -- Health status and metrics
    health_status VARCHAR(20) NOT NULL,         -- 'healthy', 'degraded', 'unhealthy', 'unknown'
    availability_status VARCHAR(20) NOT NULL,   -- 'available', 'unavailable', 'limited', 'unknown'
    performance_status VARCHAR(20) NOT NULL,    -- 'optimal', 'acceptable', 'poor', 'unknown'
    
    -- Response time metrics
    avg_response_time_ms DECIMAL(10,3),         -- Average response time
    min_response_time_ms DECIMAL(10,3),         -- Minimum response time
    max_response_time_ms DECIMAL(10,3),         -- Maximum response time
    p95_response_time_ms DECIMAL(10,3),         -- 95th percentile response time
    p99_response_time_ms DECIMAL(10,3),         -- 99th percentile response time
    
    -- Availability and reliability metrics
    uptime_percentage DECIMAL(5,2),             -- Uptime percentage
    success_rate DECIMAL(5,4),                  -- Success rate (0.0-1.0)
    error_rate DECIMAL(5,4),                    -- Error rate (0.0-1.0)
    timeout_rate DECIMAL(5,4),                  -- Timeout rate (0.0-1.0)
    
    -- Throughput metrics
    requests_per_minute DECIMAL(10,2),          -- Requests per minute
    requests_per_hour DECIMAL(10,2),            -- Requests per hour
    total_requests BIGINT,                      -- Total requests in time period
    successful_requests BIGINT,                 -- Successful requests
    failed_requests BIGINT,                     -- Failed requests
    
    -- Resource utilization
    cpu_usage_percent DECIMAL(5,2),             -- CPU usage percentage
    memory_usage_percent DECIMAL(5,2),          -- Memory usage percentage
    network_usage_mbps DECIMAL(10,2),           -- Network usage in Mbps
    disk_usage_percent DECIMAL(5,2),            -- Disk usage percentage
    
    -- Error and alert information
    active_alerts INTEGER DEFAULT 0,            -- Number of active alerts
    critical_alerts INTEGER DEFAULT 0,          -- Number of critical alerts
    warning_alerts INTEGER DEFAULT 0,           -- Number of warning alerts
    last_error_message TEXT,                    -- Last error message
    last_error_timestamp TIMESTAMP WITH TIME ZONE, -- Last error timestamp
    
    -- Configuration and limits
    timeout_threshold_ms INTEGER,               -- Timeout threshold in milliseconds
    retry_count INTEGER,                        -- Number of retries configured
    rate_limit_per_minute INTEGER,              -- Rate limit per minute
    rate_limit_per_hour INTEGER,                -- Rate limit per hour
    
    -- Health check details
    last_health_check TIMESTAMP WITH TIME ZONE, -- Last health check timestamp
    health_check_interval_seconds INTEGER,      -- Health check interval
    health_check_timeout_seconds INTEGER,       -- Health check timeout
    health_check_endpoint VARCHAR(500),         -- Health check endpoint
    
    -- Authentication and security
    auth_method VARCHAR(50),                    -- Authentication method
    auth_status VARCHAR(20),                    -- Authentication status
    ssl_enabled BOOLEAN,                        -- SSL/TLS enabled
    certificate_valid BOOLEAN,                  -- Certificate validity
    
    -- Data quality and validation
    data_quality_score DECIMAL(3,2),           -- Data quality score (0.0-1.0)
    schema_validation_status VARCHAR(20),       -- Schema validation status
    data_freshness_minutes INTEGER,             -- Data freshness in minutes
    
    -- Cost and usage tracking
    cost_per_request DECIMAL(10,6),             -- Cost per request
    total_cost_period DECIMAL(10,2),            -- Total cost for period
    usage_quota_percent DECIMAL(5,2),           -- Usage quota percentage
    
    -- Metadata and context
    tags JSONB,                                 -- Flexible key-value metadata
    metadata JSONB,                             -- Additional integration-specific data
    configuration JSONB,                        -- Integration configuration
    
    -- Constraints
    CONSTRAINT valid_integration_type CHECK (integration_type IN ('external_api', 'database', 'cache', 'queue', 'storage')),
    CONSTRAINT valid_integration_category CHECK (integration_category IN ('classification', 'verification', 'monitoring', 'notification')),
    CONSTRAINT valid_health_status CHECK (health_status IN ('healthy', 'degraded', 'unhealthy', 'unknown')),
    CONSTRAINT valid_availability_status CHECK (availability_status IN ('available', 'unavailable', 'limited', 'unknown')),
    CONSTRAINT valid_performance_status CHECK (performance_status IN ('optimal', 'acceptable', 'poor', 'unknown')),
    CONSTRAINT valid_auth_method CHECK (auth_method IN ('api_key', 'oauth', 'basic', 'bearer', 'certificate', 'none') OR auth_method IS NULL),
    CONSTRAINT valid_auth_status CHECK (auth_status IN ('authenticated', 'failed', 'expired', 'unknown') OR auth_status IS NULL),
    CONSTRAINT valid_schema_validation_status CHECK (schema_validation_status IN ('valid', 'invalid', 'unknown', 'pending') OR schema_validation_status IS NULL),
    CONSTRAINT valid_rates CHECK (
        success_rate >= 0.0 AND success_rate <= 1.0 AND
        error_rate >= 0.0 AND error_rate <= 1.0 AND
        timeout_rate >= 0.0 AND timeout_rate <= 1.0 AND
        uptime_percentage >= 0.0 AND uptime_percentage <= 100.0
    ),
    CONSTRAINT valid_data_quality CHECK (data_quality_score IS NULL OR (data_quality_score >= 0.0 AND data_quality_score <= 1.0))
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE OPTIMIZATION
-- ============================================================================

-- Unified Performance Reports Indexes
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON unified_performance_reports (created_at);
CREATE INDEX IF NOT EXISTS idx_reports_type ON unified_performance_reports (report_type);
CREATE INDEX IF NOT EXISTS idx_reports_category ON unified_performance_reports (report_category);
CREATE INDEX IF NOT EXISTS idx_reports_component ON unified_performance_reports (component) WHERE component IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reports_service ON unified_performance_reports (service_name) WHERE service_name IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reports_status ON unified_performance_reports (status);
CREATE INDEX IF NOT EXISTS idx_reports_generated_by ON unified_performance_reports (generated_by) WHERE generated_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reports_time_range ON unified_performance_reports (time_range_start, time_range_end);
CREATE INDEX IF NOT EXISTS idx_reports_tags ON unified_performance_reports USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_reports_metadata ON unified_performance_reports USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_reports_data ON unified_performance_reports USING GIN (report_data);

-- Performance Integration Health Indexes
CREATE INDEX IF NOT EXISTS idx_integration_health_timestamp ON performance_integration_health (timestamp);
CREATE INDEX IF NOT EXISTS idx_integration_health_name ON performance_integration_health (integration_name);
CREATE INDEX IF NOT EXISTS idx_integration_health_type ON performance_integration_health (integration_type);
CREATE INDEX IF NOT EXISTS idx_integration_health_category ON performance_integration_health (integration_category);
CREATE INDEX IF NOT EXISTS idx_integration_health_service ON performance_integration_health (service_name);
CREATE INDEX IF NOT EXISTS idx_integration_health_status ON performance_integration_health (health_status);
CREATE INDEX IF NOT EXISTS idx_integration_health_availability ON performance_integration_health (availability_status);
CREATE INDEX IF NOT EXISTS idx_integration_health_performance ON performance_integration_health (performance_status);
CREATE INDEX IF NOT EXISTS idx_integration_health_last_check ON performance_integration_health (last_health_check) WHERE last_health_check IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_integration_health_tags ON performance_integration_health USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_integration_health_metadata ON performance_integration_health USING GIN (metadata);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_reports_component_time ON unified_performance_reports (component, created_at) WHERE component IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reports_service_time ON unified_performance_reports (service_name, created_at) WHERE service_name IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reports_type_category ON unified_performance_reports (report_type, report_category);

CREATE INDEX IF NOT EXISTS idx_integration_health_name_time ON performance_integration_health (integration_name, timestamp);
CREATE INDEX IF NOT EXISTS idx_integration_health_type_category ON performance_integration_health (integration_type, integration_category);
CREATE INDEX IF NOT EXISTS idx_integration_health_service_time ON performance_integration_health (service_name, timestamp);

-- ============================================================================
-- UTILITY FUNCTIONS
-- ============================================================================

-- Function to create a performance report
CREATE OR REPLACE FUNCTION create_performance_report(
    p_report_name VARCHAR(200),
    p_report_type VARCHAR(50),
    p_report_category VARCHAR(50),
    p_time_range_start TIMESTAMP WITH TIME ZONE,
    p_time_range_end TIMESTAMP WITH TIME ZONE,
    p_generated_by UUID,
    p_generation_method VARCHAR(50),
    p_report_data JSONB,
    p_component VARCHAR(100) DEFAULT NULL,
    p_service_name VARCHAR(100) DEFAULT NULL,
    p_report_frequency VARCHAR(20) DEFAULT NULL,
    p_summary_data JSONB DEFAULT NULL,
    p_visualizations JSONB DEFAULT NULL,
    p_insights JSONB DEFAULT NULL,
    p_report_config JSONB DEFAULT NULL,
    p_filters_applied JSONB DEFAULT NULL,
    p_metrics_included TEXT[] DEFAULT NULL,
    p_tags JSONB DEFAULT NULL,
    p_metadata JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    report_id UUID;
BEGIN
    INSERT INTO unified_performance_reports (
        report_name, report_type, report_category,
        time_range_start, time_range_end,
        generated_by, generation_method,
        report_data, component, service_name,
        report_frequency, summary_data, visualizations, insights,
        report_config, filters_applied, metrics_included,
        tags, metadata
    ) VALUES (
        p_report_name, p_report_type, p_report_category,
        p_time_range_start, p_time_range_end,
        p_generated_by, p_generation_method,
        p_report_data, p_component, p_service_name,
        p_report_frequency, p_summary_data, p_visualizations, p_insights,
        p_report_config, p_filters_applied, p_metrics_included,
        p_tags, p_metadata
    ) RETURNING id INTO report_id;
    
    RETURN report_id;
END;
$$ LANGUAGE plpgsql;

-- Function to update integration health
CREATE OR REPLACE FUNCTION update_integration_health(
    p_integration_name VARCHAR(100),
    p_integration_type VARCHAR(50),
    p_integration_category VARCHAR(50),
    p_service_name VARCHAR(100),
    p_health_status VARCHAR(20),
    p_availability_status VARCHAR(20),
    p_performance_status VARCHAR(20),
    p_avg_response_time_ms DECIMAL(10,3) DEFAULT NULL,
    p_success_rate DECIMAL(5,4) DEFAULT NULL,
    p_error_rate DECIMAL(5,4) DEFAULT NULL,
    p_uptime_percentage DECIMAL(5,2) DEFAULT NULL,
    p_requests_per_minute DECIMAL(10,2) DEFAULT NULL,
    p_total_requests BIGINT DEFAULT NULL,
    p_successful_requests BIGINT DEFAULT NULL,
    p_failed_requests BIGINT DEFAULT NULL,
    p_active_alerts INTEGER DEFAULT 0,
    p_critical_alerts INTEGER DEFAULT 0,
    p_warning_alerts INTEGER DEFAULT 0,
    p_endpoint_url VARCHAR(500) DEFAULT NULL,
    p_last_health_check TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    p_data_quality_score DECIMAL(3,2) DEFAULT NULL,
    p_tags JSONB DEFAULT NULL,
    p_metadata JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    health_id UUID;
BEGIN
    INSERT INTO performance_integration_health (
        integration_name, integration_type, integration_category, service_name,
        health_status, availability_status, performance_status,
        avg_response_time_ms, success_rate, error_rate, uptime_percentage,
        requests_per_minute, total_requests, successful_requests, failed_requests,
        active_alerts, critical_alerts, warning_alerts,
        endpoint_url, last_health_check, data_quality_score,
        tags, metadata
    ) VALUES (
        p_integration_name, p_integration_type, p_integration_category, p_service_name,
        p_health_status, p_availability_status, p_performance_status,
        p_avg_response_time_ms, p_success_rate, p_error_rate, p_uptime_percentage,
        p_requests_per_minute, p_total_requests, p_successful_requests, p_failed_requests,
        p_active_alerts, p_critical_alerts, p_warning_alerts,
        p_endpoint_url, p_last_health_check, p_data_quality_score,
        p_tags, p_metadata
    ) RETURNING id INTO health_id;
    
    RETURN health_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- View for recent performance reports summary
CREATE VIEW recent_performance_reports AS
SELECT 
    report_name,
    report_type,
    report_category,
    component,
    service_name,
    status,
    created_at,
    time_range_start,
    time_range_end,
    generation_method,
    data_quality_score,
    completeness_score
FROM unified_performance_reports 
WHERE created_at >= NOW() - INTERVAL '7 days'
ORDER BY created_at DESC;

-- View for integration health summary
CREATE VIEW integration_health_summary AS
SELECT 
    integration_name,
    integration_type,
    integration_category,
    service_name,
    health_status,
    availability_status,
    performance_status,
    avg_response_time_ms,
    success_rate,
    error_rate,
    uptime_percentage,
    active_alerts,
    critical_alerts,
    warning_alerts,
    last_health_check,
    timestamp
FROM performance_integration_health 
WHERE timestamp >= NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC;

-- View for integration health trends
CREATE VIEW integration_health_trends AS
SELECT 
    integration_name,
    integration_type,
    service_name,
    DATE_TRUNC('hour', timestamp) as hour_bucket,
    AVG(avg_response_time_ms) as avg_response_time,
    AVG(success_rate) as avg_success_rate,
    AVG(error_rate) as avg_error_rate,
    AVG(uptime_percentage) as avg_uptime,
    COUNT(*) as health_checks_count
FROM performance_integration_health 
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY integration_name, integration_type, service_name, hour_bucket
ORDER BY hour_bucket DESC;

-- ============================================================================
-- TRIGGERS FOR AUTOMATIC PROCESSING
-- ============================================================================

-- Trigger to automatically validate report data quality
CREATE OR REPLACE FUNCTION validate_report_data_quality()
RETURNS TRIGGER AS $$
DECLARE
    quality_score DECIMAL(3,2) := 1.0;
    completeness_score DECIMAL(3,2) := 1.0;
BEGIN
    -- Calculate data quality score based on report data
    IF NEW.report_data IS NULL OR jsonb_typeof(NEW.report_data) != 'object' THEN
        quality_score := 0.0;
    ELSE
        -- Check for required fields in report data
        IF NOT (NEW.report_data ? 'metrics' AND NEW.report_data ? 'summary') THEN
            quality_score := 0.5;
        END IF;
    END IF;
    
    -- Calculate completeness score based on time range and data coverage
    IF NEW.time_range_end - NEW.time_range_start < INTERVAL '1 minute' THEN
        completeness_score := 0.3;
    ELSIF NEW.time_range_end - NEW.time_range_start < INTERVAL '1 hour' THEN
        completeness_score := 0.7;
    END IF;
    
    -- Update the scores
    NEW.data_quality_score := quality_score;
    NEW.completeness_score := completeness_score;
    
    -- Set validation status
    IF quality_score >= 0.8 AND completeness_score >= 0.8 THEN
        NEW.validation_status := 'validated';
    ELSE
        NEW.validation_status := 'failed';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for report data quality validation
CREATE TRIGGER trigger_validate_report_quality
    BEFORE INSERT OR UPDATE ON unified_performance_reports
    FOR EACH ROW
    EXECUTE FUNCTION validate_report_data_quality();

-- ============================================================================
-- INITIAL DATA AND CONFIGURATION
-- ============================================================================

-- Insert initial integration health records for known integrations
INSERT INTO performance_integration_health (
    integration_name, integration_type, integration_category, service_name,
    health_status, availability_status, performance_status,
    success_rate, error_rate, uptime_percentage,
    active_alerts, critical_alerts, warning_alerts
) VALUES 
    ('supabase-database', 'database', 'monitoring', 'kyb-database', 'healthy', 'available', 'optimal', 0.999, 0.001, 99.9, 0, 0, 0),
    ('redis-cache', 'cache', 'monitoring', 'kyb-cache', 'healthy', 'available', 'optimal', 0.998, 0.002, 99.8, 0, 0, 0),
    ('classification-api', 'external_api', 'classification', 'kyb-classification', 'healthy', 'available', 'optimal', 0.995, 0.005, 99.5, 0, 0, 0),
    ('verification-api', 'external_api', 'verification', 'kyb-verification', 'healthy', 'available', 'optimal', 0.997, 0.003, 99.7, 0, 0, 0);

-- ============================================================================
-- COMMENTS AND DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE unified_performance_reports IS 'Unified table for storing generated performance reports and analytics';
COMMENT ON TABLE performance_integration_health IS 'Comprehensive monitoring of external integrations and internal service connections';

COMMENT ON COLUMN unified_performance_reports.report_data IS 'Main report data in structured JSON format';
COMMENT ON COLUMN unified_performance_reports.summary_data IS 'Summary statistics and key metrics extracted from report data';
COMMENT ON COLUMN unified_performance_reports.visualizations IS 'Chart and graph configurations for report visualization';
COMMENT ON COLUMN unified_performance_reports.insights IS 'AI-generated insights and recommendations based on report data';

COMMENT ON COLUMN performance_integration_health.health_status IS 'Overall health status of the integration';
COMMENT ON COLUMN performance_integration_health.availability_status IS 'Availability status of the integration endpoint';
COMMENT ON COLUMN performance_integration_health.performance_status IS 'Performance status based on response times and throughput';
COMMENT ON COLUMN performance_integration_health.data_quality_score IS 'Quality score of data received from the integration (0.0-1.0)';

-- ============================================================================
-- CLEANUP AND MAINTENANCE FUNCTIONS
-- ============================================================================

-- Function to clean up old reports (retention policy)
CREATE OR REPLACE FUNCTION cleanup_old_reports(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM unified_performance_reports 
    WHERE created_at < NOW() - (retention_days || ' days')::INTERVAL
    AND status IN ('delivered', 'failed');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up old integration health records (retention policy)
CREATE OR REPLACE FUNCTION cleanup_old_integration_health(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM performance_integration_health 
    WHERE timestamp < NOW() - (retention_days || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
