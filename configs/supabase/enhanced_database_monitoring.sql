-- Enhanced Database Performance Monitoring for Business Classification System
-- This script provides comprehensive database query performance monitoring with optimization recommendations

-- 1. Create enhanced query performance monitoring table
CREATE TABLE IF NOT EXISTS enhanced_query_performance_log (
    id SERIAL PRIMARY KEY,
    query_id VARCHAR(255) NOT NULL,
    query_text TEXT NOT NULL,
    query_hash VARCHAR(255) NOT NULL,
    execution_count BIGINT DEFAULT 0,
    total_execution_time_ms DECIMAL(15,4) DEFAULT 0,
    average_execution_time_ms DECIMAL(15,4) DEFAULT 0,
    min_execution_time_ms DECIMAL(15,4) DEFAULT 0,
    max_execution_time_ms DECIMAL(15,4) DEFAULT 0,
    p50_execution_time_ms DECIMAL(15,4),
    p95_execution_time_ms DECIMAL(15,4),
    p99_execution_time_ms DECIMAL(15,4),
    rows_returned BIGINT DEFAULT 0,
    rows_examined BIGINT DEFAULT 0,
    rows_affected BIGINT DEFAULT 0,
    index_usage_score DECIMAL(5,2),
    cache_hit_ratio DECIMAL(5,2),
    performance_category VARCHAR(20),
    optimization_priority VARCHAR(20),
    optimization_score DECIMAL(5,2),
    last_executed TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    first_executed TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    error_count BIGINT DEFAULT 0,
    timeout_count BIGINT DEFAULT 0,
    lock_wait_time_ms DECIMAL(15,4) DEFAULT 0,
    buffer_reads BIGINT DEFAULT 0,
    buffer_hits BIGINT DEFAULT 0,
    temp_files_created BIGINT DEFAULT 0,
    temp_file_size_bytes BIGINT DEFAULT 0,
    sort_operations BIGINT DEFAULT 0,
    hash_operations BIGINT DEFAULT 0,
    join_operations BIGINT DEFAULT 0,
    subquery_operations BIGINT DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Create database performance alerts table
CREATE TABLE IF NOT EXISTS database_performance_alerts (
    id VARCHAR(255) PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    query_id VARCHAR(255),
    query_text TEXT,
    threshold DECIMAL(15,4),
    actual_value DECIMAL(15,4),
    message TEXT,
    recommendations TEXT[],
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create query optimization recommendations table
CREATE TABLE IF NOT EXISTS query_optimization_recommendations (
    id SERIAL PRIMARY KEY,
    query_id VARCHAR(255) NOT NULL,
    query_text TEXT NOT NULL,
    optimization_score DECIMAL(5,2),
    priority VARCHAR(20),
    estimated_improvement_percent DECIMAL(5,2),
    recommendations JSONB,
    index_suggestions JSONB,
    query_rewrite_suggestions JSONB,
    last_analyzed TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Create database system statistics table
CREATE TABLE IF NOT EXISTS database_system_stats (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    connection_count INTEGER,
    active_connections INTEGER,
    idle_connections INTEGER,
    max_connections INTEGER,
    database_size_bytes BIGINT,
    cache_hit_ratio DECIMAL(5,2),
    index_hit_ratio DECIMAL(5,2),
    lock_count INTEGER,
    deadlock_count INTEGER,
    long_running_queries INTEGER,
    slow_queries INTEGER,
    blocked_queries INTEGER,
    temp_files_created BIGINT,
    temp_file_size_bytes BIGINT,
    vacuum_operations BIGINT,
    analyze_operations BIGINT,
    replication_lag_seconds INTEGER,
    uptime_seconds BIGINT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Create indexes for faster querying
CREATE INDEX IF NOT EXISTS idx_eqpl_query_id ON enhanced_query_performance_log (query_id);
CREATE INDEX IF NOT EXISTS idx_eqpl_query_hash ON enhanced_query_performance_log (query_hash);
CREATE INDEX IF NOT EXISTS idx_eqpl_performance_category ON enhanced_query_performance_log (performance_category);
CREATE INDEX IF NOT EXISTS idx_eqpl_optimization_priority ON enhanced_query_performance_log (optimization_priority);
CREATE INDEX IF NOT EXISTS idx_eqpl_last_executed ON enhanced_query_performance_log (last_executed);
CREATE INDEX IF NOT EXISTS idx_eqpl_created_at ON enhanced_query_performance_log (created_at);

CREATE INDEX IF NOT EXISTS idx_dpa_timestamp ON database_performance_alerts (timestamp);
CREATE INDEX IF NOT EXISTS idx_dpa_alert_type ON database_performance_alerts (alert_type);
CREATE INDEX IF NOT EXISTS idx_dpa_severity ON database_performance_alerts (severity);
CREATE INDEX IF NOT EXISTS idx_dpa_resolved ON database_performance_alerts (resolved);
CREATE INDEX IF NOT EXISTS idx_dpa_query_id ON database_performance_alerts (query_id);

CREATE INDEX IF NOT EXISTS idx_qor_query_id ON query_optimization_recommendations (query_id);
CREATE INDEX IF NOT EXISTS idx_qor_priority ON query_optimization_recommendations (priority);
CREATE INDEX IF NOT EXISTS idx_qor_optimization_score ON query_optimization_recommendations (optimization_score);
CREATE INDEX IF NOT EXISTS idx_qor_last_analyzed ON query_optimization_recommendations (last_analyzed);

CREATE INDEX IF NOT EXISTS idx_dss_timestamp ON database_system_stats (timestamp);

-- 6. Create function to analyze query performance and generate recommendations
CREATE OR REPLACE FUNCTION analyze_query_performance_enhanced(
    p_query_text TEXT,
    p_execution_time_ms DECIMAL DEFAULT 0,
    p_rows_returned BIGINT DEFAULT 0,
    p_rows_examined BIGINT DEFAULT 0,
    p_error_occurred BOOLEAN DEFAULT FALSE
) RETURNS TABLE (
    query_id VARCHAR,
    performance_score DECIMAL,
    performance_category VARCHAR,
    optimization_priority VARCHAR,
    optimization_score DECIMAL,
    recommendations TEXT[],
    index_suggestions TEXT[],
    query_rewrite_suggestions TEXT[]
) AS $$
DECLARE
    v_query_id VARCHAR;
    v_query_hash VARCHAR;
    v_performance_score DECIMAL;
    v_performance_category VARCHAR;
    v_optimization_priority VARCHAR;
    v_optimization_score DECIMAL;
    v_recommendations TEXT[];
    v_index_suggestions TEXT[];
    v_query_rewrite_suggestions TEXT[];
    v_efficiency_ratio DECIMAL;
    v_slow_query_threshold DECIMAL := 100.0; -- 100ms threshold
BEGIN
    -- Generate query ID and hash
    v_query_id := 'query_' || extract(epoch from now())::bigint || '_' || abs(hashtext(p_query_text));
    v_query_hash := abs(hashtext(p_query_text))::text;
    
    -- Calculate efficiency ratio
    IF p_rows_examined > 0 THEN
        v_efficiency_ratio := p_rows_returned::DECIMAL / p_rows_examined::DECIMAL;
    ELSE
        v_efficiency_ratio := 1.0;
    END IF;
    
    -- Calculate performance score (0-100, higher is better)
    v_performance_score := 100.0;
    
    -- Deduct points for slow execution
    IF p_execution_time_ms > v_slow_query_threshold THEN
        v_performance_score := v_performance_score - ((p_execution_time_ms - v_slow_query_threshold) / 10);
    END IF;
    
    -- Deduct points for poor efficiency
    IF v_efficiency_ratio < 0.1 THEN
        v_performance_score := v_performance_score - 20;
    ELSIF v_efficiency_ratio < 0.5 THEN
        v_performance_score := v_performance_score - 10;
    END IF;
    
    -- Deduct points for errors
    IF p_error_occurred THEN
        v_performance_score := v_performance_score - 30;
    END IF;
    
    -- Ensure score is between 0 and 100
    v_performance_score := GREATEST(0, LEAST(100, v_performance_score));
    
    -- Determine performance category
    IF v_performance_score >= 90 THEN
        v_performance_category := 'excellent';
    ELSIF v_performance_score >= 75 THEN
        v_performance_category := 'good';
    ELSIF v_performance_score >= 60 THEN
        v_performance_category := 'fair';
    ELSIF v_performance_score >= 40 THEN
        v_performance_category := 'poor';
    ELSE
        v_performance_category := 'critical';
    END IF;
    
    -- Determine optimization priority
    IF v_performance_score < 40 OR p_execution_time_ms > v_slow_query_threshold * 5 THEN
        v_optimization_priority := 'high';
    ELSIF v_performance_score < 60 OR p_execution_time_ms > v_slow_query_threshold * 2 THEN
        v_optimization_priority := 'medium';
    ELSE
        v_optimization_priority := 'low';
    END IF;
    
    -- Set optimization score
    v_optimization_score := v_performance_score;
    
    -- Generate recommendations
    v_recommendations := ARRAY[]::TEXT[];
    
    IF p_execution_time_ms > v_slow_query_threshold THEN
        v_recommendations := array_append(v_recommendations, 'Consider adding indexes to improve query performance');
        v_recommendations := array_append(v_recommendations, 'Review query execution plan for optimization opportunities');
        v_recommendations := array_append(v_recommendations, 'Consider query rewriting or restructuring');
    END IF;
    
    IF v_efficiency_ratio < 0.1 THEN
        v_recommendations := array_append(v_recommendations, 'Add appropriate indexes to reduce rows examined');
        v_recommendations := array_append(v_recommendations, 'Consider adding WHERE clauses to filter data earlier');
    END IF;
    
    IF p_error_occurred THEN
        v_recommendations := array_append(v_recommendations, 'Review query logic for potential issues');
        v_recommendations := array_append(v_recommendations, 'Check for data type mismatches or constraint violations');
    END IF;
    
    -- Generate index suggestions
    v_index_suggestions := ARRAY[]::TEXT[];
    
    IF p_execution_time_ms > v_slow_query_threshold THEN
        v_index_suggestions := array_append(v_index_suggestions, 'CREATE INDEX idx_table_column ON table_name (column_name);');
        v_index_suggestions := array_append(v_index_suggestions, 'CREATE INDEX idx_table_compound ON table_name (col1, col2);');
    END IF;
    
    -- Generate query rewrite suggestions
    v_query_rewrite_suggestions := ARRAY[]::TEXT[];
    
    IF v_efficiency_ratio < 0.1 THEN
        v_query_rewrite_suggestions := array_append(v_query_rewrite_suggestions, 'Add WHERE clauses to filter data earlier in the query');
        v_query_rewrite_suggestions := array_append(v_query_rewrite_suggestions, 'Consider using LIMIT to reduce result set size');
    END IF;
    
    -- Return results
    RETURN QUERY SELECT 
        v_query_id,
        v_performance_score,
        v_performance_category,
        v_optimization_priority,
        v_optimization_score,
        v_recommendations,
        v_index_suggestions,
        v_query_rewrite_suggestions;
END;
$$ LANGUAGE plpgsql;

-- 7. Create function to collect database system statistics
CREATE OR REPLACE FUNCTION collect_database_system_stats()
RETURNS TABLE (
    connection_count INTEGER,
    active_connections INTEGER,
    idle_connections INTEGER,
    max_connections INTEGER,
    database_size_bytes BIGINT,
    cache_hit_ratio DECIMAL,
    index_hit_ratio DECIMAL,
    lock_count INTEGER,
    deadlock_count INTEGER,
    long_running_queries INTEGER,
    slow_queries INTEGER,
    blocked_queries INTEGER,
    temp_files_created BIGINT,
    temp_file_size_bytes BIGINT,
    vacuum_operations BIGINT,
    analyze_operations BIGINT,
    uptime_seconds BIGINT
) AS $$
DECLARE
    v_connection_count INTEGER;
    v_active_connections INTEGER;
    v_idle_connections INTEGER;
    v_max_connections INTEGER;
    v_database_size_bytes BIGINT;
    v_cache_hit_ratio DECIMAL;
    v_index_hit_ratio DECIMAL;
    v_lock_count INTEGER;
    v_deadlock_count INTEGER;
    v_long_running_queries INTEGER;
    v_slow_queries INTEGER;
    v_blocked_queries INTEGER;
    v_temp_files_created BIGINT;
    v_temp_file_size_bytes BIGINT;
    v_vacuum_operations BIGINT;
    v_analyze_operations BIGINT;
    v_uptime_seconds BIGINT;
BEGIN
    -- Get connection statistics
    SELECT 
        count(*)::INTEGER,
        count(*) FILTER (WHERE state = 'active')::INTEGER,
        count(*) FILTER (WHERE state = 'idle')::INTEGER,
        (SELECT setting::INTEGER FROM pg_settings WHERE name = 'max_connections')
    INTO v_connection_count, v_active_connections, v_idle_connections, v_max_connections
    FROM pg_stat_activity;
    
    -- Get database size
    SELECT pg_database_size(current_database()) INTO v_database_size_bytes;
    
    -- Get cache hit ratio
    SELECT 
        CASE 
            WHEN (blks_hit + blks_read) > 0 
            THEN round((blks_hit::DECIMAL / (blks_hit + blks_read)) * 100, 2)
            ELSE 0 
        END
    INTO v_cache_hit_ratio
    FROM pg_stat_database 
    WHERE datname = current_database();
    
    -- Get index hit ratio
    SELECT 
        CASE 
            WHEN (idx_tup_fetch + idx_tup_read) > 0 
            THEN round((idx_tup_fetch::DECIMAL / (idx_tup_fetch + idx_tup_read)) * 100, 2)
            ELSE 0 
        END
    INTO v_index_hit_ratio
    FROM pg_stat_database 
    WHERE datname = current_database();
    
    -- Get lock count
    SELECT count(*)::INTEGER INTO v_lock_count FROM pg_locks;
    
    -- Get deadlock count
    SELECT deadlocks::INTEGER INTO v_deadlock_count FROM pg_stat_database WHERE datname = current_database();
    
    -- Get long running queries (queries running for more than 1 minute)
    SELECT count(*)::INTEGER INTO v_long_running_queries 
    FROM pg_stat_activity 
    WHERE state = 'active' AND now() - query_start > interval '1 minute';
    
    -- Get slow queries (queries running for more than 5 seconds)
    SELECT count(*)::INTEGER INTO v_slow_queries 
    FROM pg_stat_activity 
    WHERE state = 'active' AND now() - query_start > interval '5 seconds';
    
    -- Get blocked queries
    SELECT count(*)::INTEGER INTO v_blocked_queries 
    FROM pg_stat_activity 
    WHERE wait_event_type = 'Lock';
    
    -- Get temp file statistics
    SELECT 
        temp_files::BIGINT,
        temp_bytes::BIGINT
    INTO v_temp_files_created, v_temp_file_size_bytes
    FROM pg_stat_database 
    WHERE datname = current_database();
    
    -- Get vacuum and analyze operations
    SELECT 
        (SELECT count(*)::BIGINT FROM pg_stat_user_tables WHERE n_tup_vacuum > 0),
        (SELECT count(*)::BIGINT FROM pg_stat_user_tables WHERE n_tup_analyze > 0)
    INTO v_vacuum_operations, v_analyze_operations;
    
    -- Get uptime
    SELECT extract(epoch from now() - pg_postmaster_start_time())::BIGINT INTO v_uptime_seconds;
    
    -- Return results
    RETURN QUERY SELECT 
        v_connection_count,
        v_active_connections,
        v_idle_connections,
        v_max_connections,
        v_database_size_bytes,
        v_cache_hit_ratio,
        v_index_hit_ratio,
        v_lock_count,
        v_deadlock_count,
        v_long_running_queries,
        v_slow_queries,
        v_blocked_queries,
        v_temp_files_created,
        v_temp_file_size_bytes,
        v_vacuum_operations,
        v_analyze_operations,
        v_uptime_seconds;
END;
$$ LANGUAGE plpgsql;

-- 8. Create function to get performance dashboard data
CREATE OR REPLACE FUNCTION get_database_performance_dashboard(
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    metric_name TEXT,
    metric_value DECIMAL,
    metric_unit TEXT,
    metric_category TEXT,
    status TEXT,
    trend TEXT,
    recommendations TEXT[]
) AS $$
DECLARE
    v_start_time TIMESTAMP WITH TIME ZONE;
    v_avg_execution_time DECIMAL;
    v_slow_query_count INTEGER;
    v_error_rate DECIMAL;
    v_cache_hit_ratio DECIMAL;
    v_connection_utilization DECIMAL;
    v_database_size_gb DECIMAL;
    v_lock_count INTEGER;
    v_deadlock_count INTEGER;
BEGIN
    v_start_time := now() - (p_hours_back || ' hours')::INTERVAL;
    
    -- Average execution time
    SELECT COALESCE(avg(average_execution_time_ms), 0) INTO v_avg_execution_time
    FROM enhanced_query_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Slow query count
    SELECT count(*)::INTEGER INTO v_slow_query_count
    FROM enhanced_query_performance_log 
    WHERE last_executed >= v_start_time 
    AND average_execution_time_ms > 100;
    
    -- Error rate
    SELECT 
        CASE 
            WHEN sum(execution_count) > 0 
            THEN round((sum(error_count)::DECIMAL / sum(execution_count)) * 100, 2)
            ELSE 0 
        END
    INTO v_error_rate
    FROM enhanced_query_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Cache hit ratio
    SELECT COALESCE(avg(cache_hit_ratio), 0) INTO v_cache_hit_ratio
    FROM database_system_stats 
    WHERE timestamp >= v_start_time;
    
    -- Connection utilization
    SELECT 
        CASE 
            WHEN max_connections > 0 
            THEN round((avg(active_connections)::DECIMAL / max_connections) * 100, 2)
            ELSE 0 
        END
    INTO v_connection_utilization
    FROM database_system_stats 
    WHERE timestamp >= v_start_time;
    
    -- Database size
    SELECT COALESCE(avg(database_size_bytes) / 1024 / 1024 / 1024, 0) INTO v_database_size_gb
    FROM database_system_stats 
    WHERE timestamp >= v_start_time;
    
    -- Lock count
    SELECT COALESCE(avg(lock_count), 0)::INTEGER INTO v_lock_count
    FROM database_system_stats 
    WHERE timestamp >= v_start_time;
    
    -- Deadlock count
    SELECT COALESCE(sum(deadlock_count), 0)::INTEGER INTO v_deadlock_count
    FROM database_system_stats 
    WHERE timestamp >= v_start_time;
    
    -- Return dashboard metrics
    RETURN QUERY
    SELECT 'Average Query Execution Time'::TEXT, v_avg_execution_time, 'ms'::TEXT, 'Performance'::TEXT,
           CASE WHEN v_avg_execution_time < 50 THEN 'OK'::TEXT
                WHEN v_avg_execution_time < 100 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_avg_execution_time > 100 THEN ARRAY['Optimize slow queries', 'Add missing indexes']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Slow Query Count'::TEXT, v_slow_query_count::DECIMAL, 'queries'::TEXT, 'Performance'::TEXT,
           CASE WHEN v_slow_query_count < 10 THEN 'OK'::TEXT
                WHEN v_slow_query_count < 50 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_slow_query_count > 10 THEN ARRAY['Review slow queries', 'Add performance monitoring']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Error Rate'::TEXT, v_error_rate, '%'::TEXT, 'Reliability'::TEXT,
           CASE WHEN v_error_rate < 1 THEN 'OK'::TEXT
                WHEN v_error_rate < 5 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_error_rate > 1 THEN ARRAY['Investigate query errors', 'Improve error handling']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Cache Hit Ratio'::TEXT, v_cache_hit_ratio, '%'::TEXT, 'Performance'::TEXT,
           CASE WHEN v_cache_hit_ratio > 95 THEN 'OK'::TEXT
                WHEN v_cache_hit_ratio > 90 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_cache_hit_ratio < 90 THEN ARRAY['Increase shared_buffers', 'Optimize queries']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Connection Utilization'::TEXT, v_connection_utilization, '%'::TEXT, 'Resources'::TEXT,
           CASE WHEN v_connection_utilization < 70 THEN 'OK'::TEXT
                WHEN v_connection_utilization < 85 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_connection_utilization > 85 THEN ARRAY['Increase max_connections', 'Optimize connection pooling']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Database Size'::TEXT, v_database_size_gb, 'GB'::TEXT, 'Storage'::TEXT,
           CASE WHEN v_database_size_gb < 10 THEN 'OK'::TEXT
                WHEN v_database_size_gb < 50 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_database_size_gb > 50 THEN ARRAY['Archive old data', 'Implement data retention']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Active Locks'::TEXT, v_lock_count::DECIMAL, 'locks'::TEXT, 'Concurrency'::TEXT,
           CASE WHEN v_lock_count < 10 THEN 'OK'::TEXT
                WHEN v_lock_count < 50 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_lock_count > 50 THEN ARRAY['Investigate lock contention', 'Optimize transactions']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Deadlocks'::TEXT, v_deadlock_count::DECIMAL, 'deadlocks'::TEXT, 'Reliability'::TEXT,
           CASE WHEN v_deadlock_count = 0 THEN 'OK'::TEXT
                WHEN v_deadlock_count < 5 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_deadlock_count > 0 THEN ARRAY['Review transaction order', 'Add deadlock detection']::TEXT[]
                ELSE ARRAY[]::TEXT[] END;
END;
$$ LANGUAGE plpgsql;

-- 9. Create view for query performance summary
CREATE OR REPLACE VIEW query_performance_summary AS
SELECT 
    query_id,
    query_text,
    query_hash,
    execution_count,
    average_execution_time_ms,
    performance_category,
    optimization_priority,
    optimization_score,
    last_executed,
    error_count,
    CASE 
        WHEN execution_count > 0 
        THEN round((error_count::DECIMAL / execution_count) * 100, 2)
        ELSE 0 
    END as error_rate_percent,
    CASE 
        WHEN rows_examined > 0 
        THEN round((rows_returned::DECIMAL / rows_examined) * 100, 2)
        ELSE 0 
    END as efficiency_percent
FROM enhanced_query_performance_log
WHERE last_executed >= now() - interval '24 hours'
ORDER BY optimization_score ASC, execution_count DESC;

-- 10. Create view for active performance alerts
CREATE OR REPLACE VIEW active_performance_alerts AS
SELECT 
    id,
    timestamp,
    alert_type,
    severity,
    query_id,
    message,
    recommendations,
    metadata
FROM database_performance_alerts
WHERE resolved = FALSE
ORDER BY 
    CASE severity 
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    timestamp DESC;

-- 11. Create view for optimization recommendations summary
CREATE OR REPLACE VIEW optimization_recommendations_summary AS
SELECT 
    qor.query_id,
    qor.query_text,
    qor.optimization_score,
    qor.priority,
    qor.estimated_improvement_percent,
    qor.last_analyzed,
    eqpl.execution_count,
    eqpl.average_execution_time_ms,
    eqpl.performance_category
FROM query_optimization_recommendations qor
LEFT JOIN enhanced_query_performance_log eqpl ON qor.query_id = eqpl.query_id
WHERE qor.last_analyzed >= now() - interval '7 days'
ORDER BY 
    CASE qor.priority 
        WHEN 'high' THEN 1
        WHEN 'medium' THEN 2
        WHEN 'low' THEN 3
    END,
    qor.optimization_score ASC;

-- 12. Create function to clean up old performance data
CREATE OR REPLACE FUNCTION cleanup_old_performance_data(
    p_days_to_keep INTEGER DEFAULT 30
) RETURNS TABLE (
    table_name TEXT,
    records_deleted BIGINT
) AS $$
DECLARE
    v_cutoff_date TIMESTAMP WITH TIME ZONE;
    v_deleted_count BIGINT;
BEGIN
    v_cutoff_date := now() - (p_days_to_keep || ' days')::INTERVAL;
    
    -- Clean up old query performance logs
    DELETE FROM enhanced_query_performance_log WHERE created_at < v_cutoff_date;
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN QUERY SELECT 'enhanced_query_performance_log'::TEXT, v_deleted_count;
    
    -- Clean up old resolved alerts
    DELETE FROM database_performance_alerts 
    WHERE resolved = TRUE AND resolved_at < v_cutoff_date;
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN QUERY SELECT 'database_performance_alerts'::TEXT, v_deleted_count;
    
    -- Clean up old system stats
    DELETE FROM database_system_stats WHERE timestamp < v_cutoff_date;
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN QUERY SELECT 'database_system_stats'::TEXT, v_deleted_count;
    
    -- Clean up old optimization recommendations
    DELETE FROM query_optimization_recommendations WHERE last_analyzed < v_cutoff_date;
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN QUERY SELECT 'query_optimization_recommendations'::TEXT, v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 13. Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_enhanced_query_performance_log_updated_at
    BEFORE UPDATE ON enhanced_query_performance_log
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_query_optimization_recommendations_updated_at
    BEFORE UPDATE ON query_optimization_recommendations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 14. Create function to get top slow queries
CREATE OR REPLACE FUNCTION get_top_slow_queries(
    p_limit INTEGER DEFAULT 10,
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    query_id VARCHAR,
    query_text TEXT,
    average_execution_time_ms DECIMAL,
    execution_count BIGINT,
    performance_category VARCHAR,
    optimization_priority VARCHAR,
    optimization_score DECIMAL,
    last_executed TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        eqpl.query_id,
        eqpl.query_text,
        eqpl.average_execution_time_ms,
        eqpl.execution_count,
        eqpl.performance_category,
        eqpl.optimization_priority,
        eqpl.optimization_score,
        eqpl.last_executed
    FROM enhanced_query_performance_log eqpl
    WHERE eqpl.last_executed >= now() - (p_hours_back || ' hours')::INTERVAL
    ORDER BY eqpl.average_execution_time_ms DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- 15. Create function to get performance trends
CREATE OR REPLACE FUNCTION get_performance_trends(
    p_hours_back INTEGER DEFAULT 24,
    p_interval_hours INTEGER DEFAULT 1
) RETURNS TABLE (
    time_bucket TIMESTAMP WITH TIME ZONE,
    avg_execution_time_ms DECIMAL,
    query_count BIGINT,
    error_count BIGINT,
    slow_query_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        date_trunc('hour', eqpl.last_executed) + 
        (extract(hour from eqpl.last_executed)::INTEGER / p_interval_hours * p_interval_hours) * interval '1 hour' as time_bucket,
        avg(eqpl.average_execution_time_ms) as avg_execution_time_ms,
        sum(eqpl.execution_count) as query_count,
        sum(eqpl.error_count) as error_count,
        count(*) FILTER (WHERE eqpl.average_execution_time_ms > 100) as slow_query_count
    FROM enhanced_query_performance_log eqpl
    WHERE eqpl.last_executed >= now() - (p_hours_back || ' hours')::INTERVAL
    GROUP BY time_bucket
    ORDER BY time_bucket;
END;
$$ LANGUAGE plpgsql;
