-- Unified Performance Monitoring for Business Classification System
-- This script provides comprehensive tables for storing unified performance monitoring data

-- 1. Create a unified performance metrics table
CREATE TABLE IF NOT EXISTS unified_performance_metrics (
    id VARCHAR(255) PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    service_name VARCHAR(100) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    version VARCHAR(20) NOT NULL,
    instance_id VARCHAR(100) NOT NULL,
    
    -- System health scores
    system_health_score DECIMAL(5,2),
    overall_performance_score DECIMAL(5,2),
    overall_security_score DECIMAL(5,2),
    
    -- Component health indicators
    response_time_health VARCHAR(20), -- "healthy", "warning", "critical"
    memory_health VARCHAR(20),
    database_health VARCHAR(20),
    security_health VARCHAR(20),
    
    -- Aggregated metrics
    total_requests BIGINT,
    average_response_time_ms DECIMAL(15,4),
    total_memory_usage_mb DECIMAL(15,4),
    total_database_queries BIGINT,
    total_security_validations BIGINT,
    
    -- Performance indicators
    error_rate DECIMAL(5,4),
    throughput_requests_per_second DECIMAL(15,4),
    resource_utilization_percent DECIMAL(5,2),
    
    -- Alert summary
    active_alerts INTEGER,
    critical_alerts INTEGER,
    warning_alerts INTEGER,
    
    -- Component-specific data (JSONB for flexibility)
    response_time_stats JSONB,
    memory_stats JSONB,
    database_stats JSONB,
    security_stats JSONB,
    
    -- Metadata
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Create a table for unified performance alerts
CREATE TABLE IF NOT EXISTS unified_performance_alerts (
    id VARCHAR(255) PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    alert_type VARCHAR(50) NOT NULL, -- "performance", "security", "resource", "system"
    severity VARCHAR(20) NOT NULL, -- "low", "medium", "high", "critical"
    title VARCHAR(255) NOT NULL,
    message TEXT,
    affected_components TEXT[], -- Array of component names
    root_cause TEXT,
    impact TEXT,
    recommendations TEXT[],
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create a table for unified performance reports
CREATE TABLE IF NOT EXISTS unified_performance_reports (
    report_id VARCHAR(255) PRIMARY KEY,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    report_period_seconds BIGINT NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    version VARCHAR(20) NOT NULL,
    instance_id VARCHAR(100) NOT NULL,
    
    -- Executive summary (JSONB for flexibility)
    executive_summary JSONB,
    
    -- Detailed analysis
    component_analysis JSONB,
    trend_analysis JSONB,
    correlation_analysis JSONB,
    
    -- Recommendations
    performance_recommendations TEXT[],
    security_recommendations TEXT[],
    resource_recommendations TEXT[],
    
    -- Report metadata
    report_format VARCHAR(20), -- "json", "html", "pdf"
    storage_location TEXT,
    file_size_bytes BIGINT,
    
    -- Metadata
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Create a table for performance integration health checks
CREATE TABLE IF NOT EXISTS performance_integration_health (
    id VARCHAR(255) PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    service_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL, -- "healthy", "degraded", "unhealthy"
    overall_score DECIMAL(5,2),
    
    -- Component health status
    component_health JSONB, -- Map of component -> health status
    
    -- Issues and recommendations
    active_issues TEXT[],
    recommendations TEXT[],
    
    -- Health check metadata
    check_duration_ms INTEGER,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Create a table for classification operations tracking
CREATE TABLE IF NOT EXISTS classification_operations (
    id VARCHAR(255) PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    service_name VARCHAR(100) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    
    -- Operation timing
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    processing_time_ms DECIMAL(15,4),
    response_time_ms DECIMAL(15,4),
    
    -- Classification results
    confidence_score DECIMAL(5,4),
    keywords_count INTEGER,
    results_count INTEGER,
    cache_hit_ratio DECIMAL(5,4),
    
    -- Operation status
    error_occurred BOOLEAN DEFAULT FALSE,
    error_message TEXT,
    
    -- Associated metrics
    database_queries JSONB, -- Array of database query executions
    security_validations JSONB, -- Array of security validation results
    
    -- Metadata
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Create indexes for faster querying
CREATE INDEX IF NOT EXISTS idx_upm_timestamp ON unified_performance_metrics (timestamp);
CREATE INDEX IF NOT EXISTS idx_upm_service_name ON unified_performance_metrics (service_name);
CREATE INDEX IF NOT EXISTS idx_upm_environment ON unified_performance_metrics (environment);
CREATE INDEX IF NOT EXISTS idx_upm_system_health_score ON unified_performance_metrics (system_health_score);
CREATE INDEX IF NOT EXISTS idx_upm_response_time_health ON unified_performance_metrics (response_time_health);
CREATE INDEX IF NOT EXISTS idx_upm_memory_health ON unified_performance_metrics (memory_health);
CREATE INDEX IF NOT EXISTS idx_upm_database_health ON unified_performance_metrics (database_health);
CREATE INDEX IF NOT EXISTS idx_upm_security_health ON unified_performance_metrics (security_health);

CREATE INDEX IF NOT EXISTS idx_upa_timestamp ON unified_performance_alerts (timestamp);
CREATE INDEX IF NOT EXISTS idx_upa_alert_type ON unified_performance_alerts (alert_type);
CREATE INDEX IF NOT EXISTS idx_upa_severity ON unified_performance_alerts (severity);
CREATE INDEX IF NOT EXISTS idx_upa_resolved ON unified_performance_alerts (resolved);

CREATE INDEX IF NOT EXISTS idx_upr_generated_at ON unified_performance_reports (generated_at);
CREATE INDEX IF NOT EXISTS idx_upr_service_name ON unified_performance_reports (service_name);
CREATE INDEX IF NOT EXISTS idx_upr_environment ON unified_performance_reports (environment);

CREATE INDEX IF NOT EXISTS idx_pih_timestamp ON performance_integration_health (timestamp);
CREATE INDEX IF NOT EXISTS idx_pih_service_name ON performance_integration_health (service_name);
CREATE INDEX IF NOT EXISTS idx_pih_status ON performance_integration_health (status);

CREATE INDEX IF NOT EXISTS idx_co_timestamp ON classification_operations (timestamp);
CREATE INDEX IF NOT EXISTS idx_co_request_id ON classification_operations (request_id);
CREATE INDEX IF NOT EXISTS idx_co_service_name ON classification_operations (service_name);
CREATE INDEX IF NOT EXISTS idx_co_endpoint ON classification_operations (endpoint);
CREATE INDEX IF NOT EXISTS idx_co_error_occurred ON classification_operations (error_occurred);
CREATE INDEX IF NOT EXISTS idx_co_start_time ON classification_operations (start_time);
CREATE INDEX IF NOT EXISTS idx_co_end_time ON classification_operations (end_time);

-- 7. Create views for common queries

-- View for system health trends
CREATE OR REPLACE VIEW system_health_trends AS
SELECT 
    DATE_TRUNC('hour', timestamp) as hour,
    service_name,
    environment,
    AVG(system_health_score) as avg_health_score,
    AVG(overall_performance_score) as avg_performance_score,
    AVG(overall_security_score) as avg_security_score,
    AVG(average_response_time_ms) as avg_response_time,
    AVG(error_rate) as avg_error_rate,
    AVG(throughput_requests_per_second) as avg_throughput,
    COUNT(*) as data_points
FROM unified_performance_metrics
WHERE timestamp >= NOW() - INTERVAL '7 days'
GROUP BY DATE_TRUNC('hour', timestamp), service_name, environment
ORDER BY hour DESC, service_name, environment;

-- View for active alerts summary
CREATE OR REPLACE VIEW active_alerts_summary AS
SELECT 
    alert_type,
    severity,
    COUNT(*) as alert_count,
    MIN(timestamp) as earliest_alert,
    MAX(timestamp) as latest_alert,
    ARRAY_AGG(DISTINCT UNNEST(affected_components)) as all_affected_components
FROM unified_performance_alerts
WHERE resolved = FALSE
GROUP BY alert_type, severity
ORDER BY 
    CASE severity 
        WHEN 'critical' THEN 1 
        WHEN 'high' THEN 2 
        WHEN 'medium' THEN 3 
        WHEN 'low' THEN 4 
    END,
    alert_count DESC;

-- View for component health summary
CREATE OR REPLACE VIEW component_health_summary AS
SELECT 
    service_name,
    environment,
    response_time_health,
    memory_health,
    database_health,
    security_health,
    COUNT(*) as occurrences,
    AVG(system_health_score) as avg_health_score,
    MAX(timestamp) as last_updated
FROM unified_performance_metrics
WHERE timestamp >= NOW() - INTERVAL '1 day'
GROUP BY service_name, environment, response_time_health, memory_health, database_health, security_health
ORDER BY service_name, environment, avg_health_score DESC;

-- View for performance recommendations
CREATE OR REPLACE VIEW performance_recommendations AS
SELECT 
    report_id,
    generated_at,
    service_name,
    environment,
    UNNEST(performance_recommendations) as recommendation,
    'performance' as recommendation_type
FROM unified_performance_reports
WHERE generated_at >= NOW() - INTERVAL '7 days'
UNION ALL
SELECT 
    report_id,
    generated_at,
    service_name,
    environment,
    UNNEST(security_recommendations) as recommendation,
    'security' as recommendation_type
FROM unified_performance_reports
WHERE generated_at >= NOW() - INTERVAL '7 days'
UNION ALL
SELECT 
    report_id,
    generated_at,
    service_name,
    environment,
    UNNEST(resource_recommendations) as recommendation,
    'resource' as recommendation_type
FROM unified_performance_reports
WHERE generated_at >= NOW() - INTERVAL '7 days'
ORDER BY generated_at DESC, service_name, environment;

-- 8. Create functions for data cleanup and maintenance

-- Function to clean up old metrics data
CREATE OR REPLACE FUNCTION cleanup_old_performance_metrics(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM unified_performance_metrics 
    WHERE timestamp < NOW() - (retention_days || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up old alerts data
CREATE OR REPLACE FUNCTION cleanup_old_performance_alerts(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM unified_performance_alerts 
    WHERE timestamp < NOW() - (retention_days || ' days')::INTERVAL
    AND resolved = TRUE;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up old reports data
CREATE OR REPLACE FUNCTION cleanup_old_performance_reports(retention_days INTEGER DEFAULT 365)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM unified_performance_reports 
    WHERE generated_at < NOW() - (retention_days || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up old health check data
CREATE OR REPLACE FUNCTION cleanup_old_health_checks(retention_days INTEGER DEFAULT 7)
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

-- Function to clean up old classification operations data
CREATE OR REPLACE FUNCTION cleanup_old_classification_operations(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM classification_operations 
    WHERE timestamp < NOW() - (retention_days || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a comprehensive cleanup function
CREATE OR REPLACE FUNCTION cleanup_all_performance_data(
    metrics_retention_days INTEGER DEFAULT 30,
    alerts_retention_days INTEGER DEFAULT 90,
    reports_retention_days INTEGER DEFAULT 365,
    health_retention_days INTEGER DEFAULT 7,
    operations_retention_days INTEGER DEFAULT 30
)
RETURNS TABLE(
    table_name TEXT,
    deleted_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 'unified_performance_metrics'::TEXT, cleanup_old_performance_metrics(metrics_retention_days)
    UNION ALL
    SELECT 'unified_performance_alerts'::TEXT, cleanup_old_performance_alerts(alerts_retention_days)
    UNION ALL
    SELECT 'unified_performance_reports'::TEXT, cleanup_old_performance_reports(reports_retention_days)
    UNION ALL
    SELECT 'performance_integration_health'::TEXT, cleanup_old_health_checks(health_retention_days)
    UNION ALL
    SELECT 'classification_operations'::TEXT, cleanup_old_classification_operations(operations_retention_days);
END;
$$ LANGUAGE plpgsql;

-- 10. Create triggers for automatic data cleanup (optional - can be run manually instead)
-- Note: These triggers would run on every insert, which might impact performance
-- Consider running cleanup functions on a schedule instead

-- Example of how to set up a scheduled cleanup (requires pg_cron extension)
-- SELECT cron.schedule('cleanup-performance-data', '0 2 * * *', 'SELECT cleanup_all_performance_data();');

-- 11. Create sample queries for common use cases

-- Query to get current system health
-- SELECT * FROM unified_performance_metrics 
-- WHERE service_name = 'classification_service' 
-- ORDER BY timestamp DESC LIMIT 1;

-- Query to get health trends for the last 24 hours
-- SELECT * FROM system_health_trends 
-- WHERE service_name = 'classification_service' 
-- AND hour >= NOW() - INTERVAL '24 hours'
-- ORDER BY hour DESC;

-- Query to get active alerts
-- SELECT * FROM active_alerts_summary;

-- Query to get component health summary
-- SELECT * FROM component_health_summary 
-- WHERE service_name = 'classification_service'
-- ORDER BY avg_health_score DESC;

-- Query to get recent performance recommendations
-- SELECT * FROM performance_recommendations 
-- WHERE service_name = 'classification_service'
-- ORDER BY generated_at DESC LIMIT 20;
