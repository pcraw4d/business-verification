-- Performance Alerting System for Business Classification Platform
-- This script provides comprehensive alerting for performance degradation across all monitoring systems

-- 1. Create a performance alerts table
CREATE TABLE IF NOT EXISTS performance_alerts (
    id SERIAL PRIMARY KEY,
    alert_id VARCHAR(255) UNIQUE NOT NULL,
    alert_type VARCHAR(100) NOT NULL,
    alert_level VARCHAR(20) NOT NULL,
    alert_category VARCHAR(50) NOT NULL,
    alert_title TEXT NOT NULL,
    alert_message TEXT NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL,
    threshold_value DECIMAL,
    threshold_type VARCHAR(20) NOT NULL,
    severity_score INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    acknowledged_by VARCHAR(255),
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT,
    affected_systems TEXT[],
    recommendations TEXT[],
    escalation_level INTEGER DEFAULT 1,
    escalation_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create a function to generate performance alerts
CREATE OR REPLACE FUNCTION generate_performance_alert(
    p_alert_type VARCHAR(100),
    p_alert_level VARCHAR(20),
    p_alert_category VARCHAR(50),
    p_alert_title TEXT,
    p_alert_message TEXT,
    p_metric_name VARCHAR(100),
    p_metric_value DECIMAL,
    p_threshold_value DECIMAL,
    p_threshold_type VARCHAR(20),
    p_affected_systems TEXT[] DEFAULT NULL,
    p_recommendations TEXT[] DEFAULT NULL
) RETURNS VARCHAR(255) AS $$
DECLARE
    v_alert_id VARCHAR(255);
    v_severity_score INTEGER;
    v_escalation_level INTEGER;
BEGIN
    -- Generate unique alert ID
    v_alert_id := p_alert_type || '_' || EXTRACT(EPOCH FROM NOW())::BIGINT || '_' || RANDOM()::TEXT;
    
    -- Calculate severity score based on alert level
    v_severity_score := CASE p_alert_level
        WHEN 'CRITICAL' THEN 100
        WHEN 'HIGH' THEN 80
        WHEN 'MEDIUM' THEN 60
        WHEN 'LOW' THEN 40
        WHEN 'INFO' THEN 20
        ELSE 0
    END;
    
    -- Determine escalation level
    v_escalation_level := CASE p_alert_level
        WHEN 'CRITICAL' THEN 3
        WHEN 'HIGH' THEN 2
        WHEN 'MEDIUM' THEN 1
        ELSE 0
    END;
    
    -- Insert alert
    INSERT INTO performance_alerts (
        alert_id,
        alert_type,
        alert_level,
        alert_category,
        alert_title,
        alert_message,
        metric_name,
        metric_value,
        threshold_value,
        threshold_type,
        severity_score,
        affected_systems,
        recommendations,
        escalation_level
    ) VALUES (
        v_alert_id,
        p_alert_type,
        p_alert_level,
        p_alert_category,
        p_alert_title,
        p_alert_message,
        p_metric_name,
        p_metric_value,
        p_threshold_value,
        p_threshold_type,
        v_severity_score,
        p_affected_systems,
        p_recommendations,
        v_escalation_level
    );
    
    RETURN v_alert_id;
END;
$$ LANGUAGE plpgsql;

-- 3. Create a function to check database performance alerts
CREATE OR REPLACE FUNCTION check_database_performance_alerts() 
RETURNS TABLE (
    alert_id VARCHAR(255),
    alert_type VARCHAR(100),
    alert_level VARCHAR(20),
    alert_title TEXT,
    alert_message TEXT,
    metric_name VARCHAR(100),
    metric_value DECIMAL,
    threshold_value DECIMAL,
    recommendations TEXT[]
) AS $$
DECLARE
    v_db_size_bytes BIGINT;
    v_active_connections INT;
    v_avg_query_time_ms NUMERIC;
    v_total_queries BIGINT;
    v_connection_utilization NUMERIC;
    v_max_connections INT;
BEGIN
    -- Get current database metrics
    SELECT 
        pg_database_size(current_database()),
        (SELECT count(*)::INT FROM pg_stat_activity WHERE datname = current_database()),
        (SELECT AVG(mean_exec_time) FROM pg_stat_statements WHERE mean_exec_time > 0),
        (SELECT sum(calls) FROM pg_stat_statements),
        (SELECT count(*)::NUMERIC FROM pg_stat_activity WHERE datname = current_database()) * 100 / (SELECT setting::NUMERIC FROM pg_settings WHERE name = 'max_connections'),
        (SELECT setting::INT FROM pg_settings WHERE name = 'max_connections')
    INTO v_db_size_bytes, v_active_connections, v_avg_query_time_ms, v_total_queries, v_connection_utilization, v_max_connections;
    
    -- Check database size alerts (assuming 500MB limit for free tier)
    IF v_db_size_bytes > 500 * 1024 * 1024 THEN
        RETURN QUERY SELECT 
            generate_performance_alert(
                'DATABASE_SIZE',
                'CRITICAL',
                'STORAGE',
                'Database Size Exceeded',
                'Database size has exceeded the free tier limit of 500MB',
                'database_size_bytes',
                v_db_size_bytes::DECIMAL,
                500 * 1024 * 1024::DECIMAL,
                'greater_than',
                ARRAY['database', 'storage'],
                ARRAY['Consider upgrading to paid plan', 'Archive old data', 'Optimize database queries']
            )::VARCHAR(255),
            'DATABASE_SIZE'::VARCHAR(100),
            'CRITICAL'::VARCHAR(20),
            'Database Size Exceeded'::TEXT,
            'Database size has exceeded the free tier limit of 500MB'::TEXT,
            'database_size_bytes'::VARCHAR(100),
            v_db_size_bytes::DECIMAL,
            500 * 1024 * 1024::DECIMAL,
            ARRAY['Consider upgrading to paid plan', 'Archive old data', 'Optimize database queries'];
    END IF;
    
    -- Check connection utilization alerts
    IF v_connection_utilization > 90 THEN
        RETURN QUERY SELECT 
            generate_performance_alert(
                'CONNECTION_UTILIZATION',
                'CRITICAL',
                'CONNECTIONS',
                'High Connection Utilization',
                'Connection utilization is above 90%',
                'connection_utilization_percentage',
                v_connection_utilization,
                90.0,
                'greater_than',
                ARRAY['database', 'connections'],
                ARRAY['Increase max_connections', 'Optimize connection pooling', 'Review long-running queries']
            )::VARCHAR(255),
            'CONNECTION_UTILIZATION'::VARCHAR(100),
            'CRITICAL'::VARCHAR(20),
            'High Connection Utilization'::TEXT,
            'Connection utilization is above 90%'::TEXT,
            'connection_utilization_percentage'::VARCHAR(100),
            v_connection_utilization,
            90.0,
            ARRAY['Increase max_connections', 'Optimize connection pooling', 'Review long-running queries'];
    ELSIF v_connection_utilization > 70 THEN
        RETURN QUERY SELECT 
            generate_performance_alert(
                'CONNECTION_UTILIZATION',
                'HIGH',
                'CONNECTIONS',
                'Moderate Connection Utilization',
                'Connection utilization is above 70%',
                'connection_utilization_percentage',
                v_connection_utilization,
                70.0,
                'greater_than',
                ARRAY['database', 'connections'],
                ARRAY['Monitor connection usage', 'Consider connection optimization']
            )::VARCHAR(255),
            'CONNECTION_UTILIZATION'::VARCHAR(100),
            'HIGH'::VARCHAR(20),
            'Moderate Connection Utilization'::TEXT,
            'Connection utilization is above 70%'::TEXT,
            'connection_utilization_percentage'::VARCHAR(100),
            v_connection_utilization,
            70.0,
            ARRAY['Monitor connection usage', 'Consider connection optimization'];
    END IF;
    
    -- Check query performance alerts
    IF v_avg_query_time_ms > 5000 THEN
        RETURN QUERY SELECT 
            generate_performance_alert(
                'QUERY_PERFORMANCE',
                'CRITICAL',
                'PERFORMANCE',
                'Slow Query Performance',
                'Average query time is above 5000ms',
                'avg_query_time_ms',
                v_avg_query_time_ms,
                5000.0,
                'greater_than',
                ARRAY['database', 'queries'],
                ARRAY['Optimize slow queries', 'Add database indexes', 'Review query patterns']
            )::VARCHAR(255),
            'QUERY_PERFORMANCE'::VARCHAR(100),
            'CRITICAL'::VARCHAR(20),
            'Slow Query Performance'::TEXT,
            'Average query time is above 5000ms'::TEXT,
            'avg_query_time_ms'::VARCHAR(100),
            v_avg_query_time_ms,
            5000.0,
            ARRAY['Optimize slow queries', 'Add database indexes', 'Review query patterns'];
    ELSIF v_avg_query_time_ms > 1000 THEN
        RETURN QUERY SELECT 
            generate_performance_alert(
                'QUERY_PERFORMANCE',
                'HIGH',
                'PERFORMANCE',
                'Moderate Query Performance',
                'Average query time is above 1000ms',
                'avg_query_time_ms',
                v_avg_query_time_ms,
                1000.0,
                'greater_than',
                ARRAY['database', 'queries'],
                ARRAY['Monitor query performance', 'Consider query optimization']
            )::VARCHAR(255),
            'QUERY_PERFORMANCE'::VARCHAR(100),
            'HIGH'::VARCHAR(20),
            'Moderate Query Performance'::TEXT,
            'Average query time is above 1000ms'::TEXT,
            'avg_query_time_ms'::VARCHAR(100),
            v_avg_query_time_ms,
            1000.0,
            ARRAY['Monitor query performance', 'Consider query optimization'];
    END IF;
    
    -- Return empty result if no alerts
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to check classification accuracy alerts
CREATE OR REPLACE FUNCTION check_classification_accuracy_alerts() 
RETURNS TABLE (
    alert_id VARCHAR(255),
    alert_type VARCHAR(100),
    alert_level VARCHAR(20),
    alert_title TEXT,
    alert_message TEXT,
    metric_name VARCHAR(100),
    metric_value DECIMAL,
    threshold_value DECIMAL,
    recommendations TEXT[]
) AS $$
DECLARE
    v_accuracy_percentage NUMERIC;
    v_avg_response_time_ms NUMERIC;
    v_error_rate NUMERIC;
    v_avg_confidence NUMERIC;
BEGIN
    -- Get classification accuracy metrics (assuming classification_accuracy_metrics table exists)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'classification_accuracy_metrics') THEN
        SELECT 
            COALESCE(
                (COUNT(*) FILTER (WHERE is_correct = TRUE)::NUMERIC / 
                 NULLIF(COUNT(*) FILTER (WHERE is_correct IS NOT NULL), 0)) * 100, 
                0
            ),
            COALESCE(AVG(response_time_ms), 0),
            COALESCE(
                (COUNT(*) FILTER (WHERE error_message IS NOT NULL)::NUMERIC / COUNT(*)) * 100, 
                0
            ),
            COALESCE(AVG(predicted_confidence), 0)
        INTO v_accuracy_percentage, v_avg_response_time_ms, v_error_rate, v_avg_confidence
        FROM classification_accuracy_metrics
        WHERE timestamp >= NOW() - INTERVAL '1 hour';
        
        -- Check accuracy alerts
        IF v_accuracy_percentage < 60 THEN
            RETURN QUERY SELECT 
                generate_performance_alert(
                    'CLASSIFICATION_ACCURACY',
                    'CRITICAL',
                    'ACCURACY',
                    'Low Classification Accuracy',
                    'Classification accuracy is below 60%',
                    'accuracy_percentage',
                    v_accuracy_percentage,
                    60.0,
                    'less_than',
                    ARRAY['classification', 'accuracy'],
                    ARRAY['Review classification algorithms', 'Improve training data', 'Check keyword matching']
                )::VARCHAR(255),
                'CLASSIFICATION_ACCURACY'::VARCHAR(100),
                'CRITICAL'::VARCHAR(20),
                'Low Classification Accuracy'::TEXT,
                'Classification accuracy is below 60%'::TEXT,
                'accuracy_percentage'::VARCHAR(100),
                v_accuracy_percentage,
                60.0,
                ARRAY['Review classification algorithms', 'Improve training data', 'Check keyword matching'];
        ELSIF v_accuracy_percentage < 75 THEN
            RETURN QUERY SELECT 
                generate_performance_alert(
                    'CLASSIFICATION_ACCURACY',
                    'HIGH',
                    'ACCURACY',
                    'Moderate Classification Accuracy',
                    'Classification accuracy is below 75%',
                    'accuracy_percentage',
                    v_accuracy_percentage,
                    75.0,
                    'less_than',
                    ARRAY['classification', 'accuracy'],
                    ARRAY['Monitor classification performance', 'Consider algorithm improvements']
                )::VARCHAR(255),
                'CLASSIFICATION_ACCURACY'::VARCHAR(100),
                'HIGH'::VARCHAR(20),
                'Moderate Classification Accuracy'::TEXT,
                    'Classification accuracy is below 75%'::TEXT,
                'accuracy_percentage'::VARCHAR(100),
                v_accuracy_percentage,
                75.0,
                ARRAY['Monitor classification performance', 'Consider algorithm improvements'];
        END IF;
        
        -- Check response time alerts
        IF v_avg_response_time_ms > 5000 THEN
            RETURN QUERY SELECT 
                generate_performance_alert(
                    'CLASSIFICATION_RESPONSE_TIME',
                    'CRITICAL',
                    'PERFORMANCE',
                    'Slow Classification Response',
                    'Classification response time is above 5000ms',
                    'avg_response_time_ms',
                    v_avg_response_time_ms,
                    5000.0,
                    'greater_than',
                    ARRAY['classification', 'performance'],
                    ARRAY['Optimize classification algorithms', 'Review database queries', 'Check external API calls']
                )::VARCHAR(255),
                'CLASSIFICATION_RESPONSE_TIME'::VARCHAR(100),
                'CRITICAL'::VARCHAR(20),
                'Slow Classification Response'::TEXT,
                'Classification response time is above 5000ms'::TEXT,
                'avg_response_time_ms'::VARCHAR(100),
                v_avg_response_time_ms,
                5000.0,
                ARRAY['Optimize classification algorithms', 'Review database queries', 'Check external API calls'];
        ELSIF v_avg_response_time_ms > 2000 THEN
            RETURN QUERY SELECT 
                generate_performance_alert(
                    'CLASSIFICATION_RESPONSE_TIME',
                    'HIGH',
                    'PERFORMANCE',
                    'Moderate Classification Response Time',
                    'Classification response time is above 2000ms',
                    'avg_response_time_ms',
                    v_avg_response_time_ms,
                    2000.0,
                    'greater_than',
                    ARRAY['classification', 'performance'],
                    ARRAY['Monitor response times', 'Consider performance optimization']
                )::VARCHAR(255),
                'CLASSIFICATION_RESPONSE_TIME'::VARCHAR(100),
                'HIGH'::VARCHAR(20),
                'Moderate Classification Response Time'::TEXT,
                'Classification response time is above 2000ms'::TEXT,
                'avg_response_time_ms'::VARCHAR(100),
                v_avg_response_time_ms,
                2000.0,
                ARRAY['Monitor response times', 'Consider performance optimization'];
        END IF;
        
        -- Check error rate alerts
        IF v_error_rate > 20 THEN
            RETURN QUERY SELECT 
                generate_performance_alert(
                    'CLASSIFICATION_ERROR_RATE',
                    'CRITICAL',
                    'ERRORS',
                    'High Classification Error Rate',
                    'Classification error rate is above 20%',
                    'error_rate_percentage',
                    v_error_rate,
                    20.0,
                    'greater_than',
                    ARRAY['classification', 'errors'],
                    ARRAY['Investigate classification errors', 'Check system logs', 'Review error handling']
                )::VARCHAR(255),
                'CLASSIFICATION_ERROR_RATE'::VARCHAR(100),
                'CRITICAL'::VARCHAR(20),
                'High Classification Error Rate'::TEXT,
                'Classification error rate is above 20%'::TEXT,
                'error_rate_percentage'::VARCHAR(100),
                v_error_rate,
                20.0,
                ARRAY['Investigate classification errors', 'Check system logs', 'Review error handling'];
        ELSIF v_error_rate > 10 THEN
            RETURN QUERY SELECT 
                generate_performance_alert(
                    'CLASSIFICATION_ERROR_RATE',
                    'HIGH',
                    'ERRORS',
                    'Moderate Classification Error Rate',
                    'Classification error rate is above 10%',
                    'error_rate_percentage',
                    v_error_rate,
                    10.0,
                    'greater_than',
                    ARRAY['classification', 'errors'],
                    ARRAY['Monitor error rates', 'Review error patterns']
                )::VARCHAR(255),
                'CLASSIFICATION_ERROR_RATE'::VARCHAR(100),
                'HIGH'::VARCHAR(20),
                'Moderate Classification Error Rate'::TEXT,
                'Classification error rate is above 10%'::TEXT,
                'error_rate_percentage'::VARCHAR(100),
                v_error_rate,
                10.0,
                ARRAY['Monitor error rates', 'Review error patterns'];
        END IF;
        
        -- Check confidence alerts
        IF v_avg_confidence < 50 THEN
            RETURN QUERY SELECT 
                generate_performance_alert(
                    'CLASSIFICATION_CONFIDENCE',
                    'CRITICAL',
                    'CONFIDENCE',
                    'Low Classification Confidence',
                    'Average classification confidence is below 50%',
                    'avg_confidence_percentage',
                    v_avg_confidence,
                    50.0,
                    'less_than',
                    ARRAY['classification', 'confidence'],
                    ARRAY['Improve classification algorithms', 'Enhance training data', 'Review keyword matching']
                )::VARCHAR(255),
                'CLASSIFICATION_CONFIDENCE'::VARCHAR(100),
                'CRITICAL'::VARCHAR(20),
                'Low Classification Confidence'::TEXT,
                'Average classification confidence is below 50%'::TEXT,
                'avg_confidence_percentage'::VARCHAR(100),
                v_avg_confidence,
                50.0,
                ARRAY['Improve classification algorithms', 'Enhance training data', 'Review keyword matching'];
        ELSIF v_avg_confidence < 70 THEN
            RETURN QUERY SELECT 
                generate_performance_alert(
                    'CLASSIFICATION_CONFIDENCE',
                    'HIGH',
                    'CONFIDENCE',
                    'Moderate Classification Confidence',
                    'Average classification confidence is below 70%',
                    'avg_confidence_percentage',
                    v_avg_confidence,
                    70.0,
                    'less_than',
                    ARRAY['classification', 'confidence'],
                    ARRAY['Monitor confidence scores', 'Consider algorithm improvements']
                )::VARCHAR(255),
                'CLASSIFICATION_CONFIDENCE'::VARCHAR(100),
                'HIGH'::VARCHAR(20),
                'Moderate Classification Confidence'::TEXT,
                'Average classification confidence is below 70%'::TEXT,
                'avg_confidence_percentage'::VARCHAR(100),
                v_avg_confidence,
                70.0,
                ARRAY['Monitor confidence scores', 'Consider algorithm improvements'];
        END IF;
    END IF;
    
    -- Return empty result if no alerts
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 5. Create a function to check system resource alerts
CREATE OR REPLACE FUNCTION check_system_resource_alerts() 
RETURNS TABLE (
    alert_id VARCHAR(255),
    alert_type VARCHAR(100),
    alert_level VARCHAR(20),
    alert_title TEXT,
    alert_message TEXT,
    metric_name VARCHAR(100),
    metric_value DECIMAL,
    threshold_value DECIMAL,
    recommendations TEXT[]
) AS $$
DECLARE
    v_cpu_usage NUMERIC;
    v_memory_usage NUMERIC;
    v_disk_usage NUMERIC;
    v_network_latency NUMERIC;
BEGIN
    -- Note: These metrics would typically come from system monitoring tools
    -- For now, we'll use placeholder values and check if monitoring data exists
    
    -- Check if we have any system monitoring data
    -- In a real implementation, this would query actual system metrics
    
    -- CPU usage check (placeholder)
    v_cpu_usage := 0; -- Would be actual CPU usage from monitoring system
    
    -- Memory usage check (placeholder)
    v_memory_usage := 0; -- Would be actual memory usage from monitoring system
    
    -- Disk usage check (placeholder)
    v_disk_usage := 0; -- Would be actual disk usage from monitoring system
    
    -- Network latency check (placeholder)
    v_network_latency := 0; -- Would be actual network latency from monitoring system
    
    -- For demonstration, we'll create some example alerts
    -- In a real system, these would be based on actual metrics
    
    -- Example: High CPU usage alert
    IF v_cpu_usage > 90 THEN
        RETURN QUERY SELECT 
            generate_performance_alert(
                'CPU_USAGE',
                'CRITICAL',
                'RESOURCES',
                'High CPU Usage',
                'CPU usage is above 90%',
                'cpu_usage_percentage',
                v_cpu_usage,
                90.0,
                'greater_than',
                ARRAY['system', 'cpu'],
                ARRAY['Scale up resources', 'Optimize application performance', 'Check for resource leaks']
            )::VARCHAR(255),
            'CPU_USAGE'::VARCHAR(100),
            'CRITICAL'::VARCHAR(20),
            'High CPU Usage'::TEXT,
            'CPU usage is above 90%'::TEXT,
            'cpu_usage_percentage'::VARCHAR(100),
            v_cpu_usage,
            90.0,
            ARRAY['Scale up resources', 'Optimize application performance', 'Check for resource leaks'];
    END IF;
    
    -- Example: High memory usage alert
    IF v_memory_usage > 85 THEN
        RETURN QUERY SELECT 
            generate_performance_alert(
                'MEMORY_USAGE',
                'HIGH',
                'RESOURCES',
                'High Memory Usage',
                'Memory usage is above 85%',
                'memory_usage_percentage',
                v_memory_usage,
                85.0,
                'greater_than',
                ARRAY['system', 'memory'],
                ARRAY['Increase memory allocation', 'Check for memory leaks', 'Optimize memory usage']
            )::VARCHAR(255),
            'MEMORY_USAGE'::VARCHAR(100),
            'HIGH'::VARCHAR(20),
            'High Memory Usage'::TEXT,
            'Memory usage is above 85%'::TEXT,
            'memory_usage_percentage'::VARCHAR(100),
            v_memory_usage,
            85.0,
            ARRAY['Increase memory allocation', 'Check for memory leaks', 'Optimize memory usage'];
    END IF;
    
    -- Return empty result if no alerts
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 6. Create a function to get all active alerts
CREATE OR REPLACE FUNCTION get_active_performance_alerts() 
RETURNS TABLE (
    alert_id VARCHAR(255),
    alert_type VARCHAR(100),
    alert_level VARCHAR(20),
    alert_category VARCHAR(50),
    alert_title TEXT,
    alert_message TEXT,
    metric_name VARCHAR(100),
    metric_value DECIMAL,
    threshold_value DECIMAL,
    severity_score INTEGER,
    status VARCHAR(20),
    affected_systems TEXT[],
    recommendations TEXT[],
    escalation_level INTEGER,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pa.alert_id,
        pa.alert_type,
        pa.alert_level,
        pa.alert_category,
        pa.alert_title,
        pa.alert_message,
        pa.metric_name,
        pa.metric_value,
        pa.threshold_value,
        pa.severity_score,
        pa.status,
        pa.affected_systems,
        pa.recommendations,
        pa.escalation_level,
        pa.created_at,
        pa.updated_at
    FROM performance_alerts pa
    WHERE pa.status = 'active'
    ORDER BY pa.severity_score DESC, pa.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 7. Create a function to acknowledge an alert
CREATE OR REPLACE FUNCTION acknowledge_performance_alert(
    p_alert_id VARCHAR(255),
    p_acknowledged_by VARCHAR(255)
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE performance_alerts 
    SET 
        status = 'acknowledged',
        acknowledged_by = p_acknowledged_by,
        acknowledged_at = NOW(),
        updated_at = NOW()
    WHERE alert_id = p_alert_id AND status = 'active';
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- 8. Create a function to resolve an alert
CREATE OR REPLACE FUNCTION resolve_performance_alert(
    p_alert_id VARCHAR(255),
    p_resolution_notes TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE performance_alerts 
    SET 
        status = 'resolved',
        resolved_at = NOW(),
        resolution_notes = p_resolution_notes,
        updated_at = NOW()
    WHERE alert_id = p_alert_id AND status IN ('active', 'acknowledged');
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a function to run all performance checks
CREATE OR REPLACE FUNCTION run_all_performance_checks() 
RETURNS TABLE (
    check_type VARCHAR(100),
    alerts_generated INTEGER,
    check_status VARCHAR(20),
    check_message TEXT
) AS $$
DECLARE
    v_db_alerts INTEGER := 0;
    v_classification_alerts INTEGER := 0;
    v_system_alerts INTEGER := 0;
BEGIN
    -- Run database performance checks
    BEGIN
        SELECT COUNT(*) INTO v_db_alerts FROM check_database_performance_alerts();
        RETURN QUERY SELECT 
            'DATABASE_PERFORMANCE'::VARCHAR(100),
            v_db_alerts,
            'COMPLETED'::VARCHAR(20),
            'Database performance checks completed'::TEXT;
    EXCEPTION WHEN OTHERS THEN
        RETURN QUERY SELECT 
            'DATABASE_PERFORMANCE'::VARCHAR(100),
            0,
            'ERROR'::VARCHAR(20),
            'Database performance checks failed: ' || SQLERRM::TEXT;
    END;
    
    -- Run classification accuracy checks
    BEGIN
        SELECT COUNT(*) INTO v_classification_alerts FROM check_classification_accuracy_alerts();
        RETURN QUERY SELECT 
            'CLASSIFICATION_ACCURACY'::VARCHAR(100),
            v_classification_alerts,
            'COMPLETED'::VARCHAR(20),
            'Classification accuracy checks completed'::TEXT;
    EXCEPTION WHEN OTHERS THEN
        RETURN QUERY SELECT 
            'CLASSIFICATION_ACCURACY'::VARCHAR(100),
            0,
            'ERROR'::VARCHAR(20),
            'Classification accuracy checks failed: ' || SQLERRM::TEXT;
    END;
    
    -- Run system resource checks
    BEGIN
        SELECT COUNT(*) INTO v_system_alerts FROM check_system_resource_alerts();
        RETURN QUERY SELECT 
            'SYSTEM_RESOURCES'::VARCHAR(100),
            v_system_alerts,
            'COMPLETED'::VARCHAR(20),
            'System resource checks completed'::TEXT;
    EXCEPTION WHEN OTHERS THEN
        RETURN QUERY SELECT 
            'SYSTEM_RESOURCES'::VARCHAR(100),
            0,
            'ERROR'::VARCHAR(20),
            'System resource checks failed: ' || SQLERRM::TEXT;
    END;
END;
$$ LANGUAGE plpgsql;

-- 10. Create a function to get alert statistics
CREATE OR REPLACE FUNCTION get_alert_statistics(
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    total_alerts BIGINT,
    active_alerts BIGINT,
    acknowledged_alerts BIGINT,
    resolved_alerts BIGINT,
    critical_alerts BIGINT,
    high_alerts BIGINT,
    medium_alerts BIGINT,
    low_alerts BIGINT,
    alerts_by_category JSONB,
    alerts_by_type JSONB,
    avg_resolution_time_minutes NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_alerts,
        COUNT(*) FILTER (WHERE status = 'active') as active_alerts,
        COUNT(*) FILTER (WHERE status = 'acknowledged') as acknowledged_alerts,
        COUNT(*) FILTER (WHERE status = 'resolved') as resolved_alerts,
        COUNT(*) FILTER (WHERE alert_level = 'CRITICAL') as critical_alerts,
        COUNT(*) FILTER (WHERE alert_level = 'HIGH') as high_alerts,
        COUNT(*) FILTER (WHERE alert_level = 'MEDIUM') as medium_alerts,
        COUNT(*) FILTER (WHERE alert_level = 'LOW') as low_alerts,
        jsonb_object_agg(alert_category, category_count) as alerts_by_category,
        jsonb_object_agg(alert_type, type_count) as alerts_by_type,
        ROUND(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 60), 2) as avg_resolution_time_minutes
    FROM (
        SELECT 
            alert_category,
            alert_type,
            COUNT(*) as category_count,
            COUNT(*) as type_count
        FROM performance_alerts
        WHERE created_at >= NOW() - INTERVAL '1 hour' * p_hours_back
        GROUP BY alert_category, alert_type
    ) stats;
END;
$$ LANGUAGE plpgsql;

-- 11. Create a function to cleanup old alerts
CREATE OR REPLACE FUNCTION cleanup_old_performance_alerts(
    p_days_to_keep INTEGER DEFAULT 30
) RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM performance_alerts
    WHERE created_at < NOW() - INTERVAL '1 day' * p_days_to_keep
    AND status = 'resolved';
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 12. Create a function to validate alerting setup
CREATE OR REPLACE FUNCTION validate_alerting_setup() 
RETURNS TABLE (
    component TEXT,
    status TEXT,
    details TEXT,
    recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check if performance_alerts table exists
    SELECT 
        'Performance Alerts Table' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_alerts') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Table for storing performance alerts' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_alerts') 
            THEN 'Table exists and ready for use' 
            ELSE 'Create performance_alerts table' 
        END as recommendation
    
    UNION ALL
    
    -- Check if alerting functions exist
    SELECT 
        'Alerting Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'generate_performance_alert') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for generating and managing alerts' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'generate_performance_alert') 
            THEN 'All alerting functions are available' 
            ELSE 'Create alerting functions' 
        END as recommendation
    
    UNION ALL
    
    -- Check if monitoring functions exist
    SELECT 
        'Monitoring Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'check_database_performance_alerts') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for checking performance metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'check_database_performance_alerts') 
            THEN 'All monitoring functions are available' 
            ELSE 'Create monitoring functions' 
        END as recommendation;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_performance_alerts_alert_id ON performance_alerts(alert_id);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_status ON performance_alerts(status);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_alert_level ON performance_alerts(alert_level);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_alert_type ON performance_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_created_at ON performance_alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_severity_score ON performance_alerts(severity_score);

-- Create a view for easy alert dashboard access
CREATE OR REPLACE VIEW performance_alert_dashboard AS
SELECT 
    'Active Alerts' as metric_name,
    (SELECT COUNT(*)::TEXT FROM performance_alerts WHERE status = 'active') as current_value,
    '0' as target_value,
    (SELECT 
        CASE 
            WHEN COUNT(*) = 0 THEN 'OK'
            WHEN COUNT(*) FILTER (WHERE alert_level = 'CRITICAL') > 0 THEN 'CRITICAL'
            WHEN COUNT(*) FILTER (WHERE alert_level = 'HIGH') > 0 THEN 'WARNING'
            ELSE 'FAIR'
        END
    FROM performance_alerts WHERE status = 'active') as status
UNION ALL
SELECT 
    'Critical Alerts' as metric_name,
    (SELECT COUNT(*)::TEXT FROM performance_alerts WHERE status = 'active' AND alert_level = 'CRITICAL') as current_value,
    '0' as target_value,
    (SELECT 
        CASE 
            WHEN COUNT(*) = 0 THEN 'OK'
            ELSE 'CRITICAL'
        END
    FROM performance_alerts WHERE status = 'active' AND alert_level = 'CRITICAL') as status
UNION ALL
SELECT 
    'High Alerts' as metric_name,
    (SELECT COUNT(*)::TEXT FROM performance_alerts WHERE status = 'active' AND alert_level = 'HIGH') as current_value,
    '0' as target_value,
    (SELECT 
        CASE 
            WHEN COUNT(*) = 0 THEN 'OK'
            ELSE 'WARNING'
        END
    FROM performance_alerts WHERE status = 'active' AND alert_level = 'HIGH') as status;

-- Grant permissions
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO authenticated;
GRANT SELECT ON performance_alert_dashboard TO authenticated;
GRANT SELECT, INSERT, UPDATE ON performance_alerts TO authenticated;

-- Initial setup message
DO $$
BEGIN
    RAISE NOTICE 'Performance alerting system setup completed successfully!';
    RAISE NOTICE 'Total functions created: 12';
    RAISE NOTICE 'Total tables created: 1';
    RAISE NOTICE 'Total views created: 1';
    RAISE NOTICE 'Total indexes created: 6';
    RAISE NOTICE 'All performance alerting tools are now available.';
    RAISE NOTICE 'Use performance_alert_dashboard view to access current alert status.';
    RAISE NOTICE 'Call run_all_performance_checks() to run all performance checks.';
    RAISE NOTICE 'Call get_active_performance_alerts() to get current active alerts.';
END $$;
