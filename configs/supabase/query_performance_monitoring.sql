-- Query Performance Monitoring for Business Classification System
-- This script provides comprehensive query performance monitoring and optimization

-- 1. Create a query performance monitoring table
CREATE TABLE IF NOT EXISTS query_performance_log (
    id SERIAL PRIMARY KEY,
    query_id BIGINT,
    query_text TEXT,
    execution_time_ms DECIMAL(10,3),
    rows_returned BIGINT,
    rows_examined BIGINT,
    index_usage_score DECIMAL(5,2),
    cache_hit_ratio DECIMAL(5,2),
    query_complexity_score INTEGER,
    performance_category VARCHAR(20),
    optimization_priority VARCHAR(20),
    recommendations TEXT,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id UUID,
    session_id VARCHAR(255),
    request_id VARCHAR(255)
);

-- 2. Create a function to analyze query performance
CREATE OR REPLACE FUNCTION analyze_query_performance(
    p_query_text TEXT,
    p_execution_time_ms DECIMAL DEFAULT 0,
    p_rows_returned BIGINT DEFAULT 0,
    p_rows_examined BIGINT DEFAULT 0
) RETURNS TABLE (
    query_id BIGINT,
    performance_score DECIMAL,
    performance_category VARCHAR,
    optimization_priority VARCHAR,
    recommendations TEXT,
    index_suggestions TEXT[],
    query_optimization_hints TEXT[]
) AS $$
DECLARE
    v_query_id BIGINT;
    v_performance_score DECIMAL;
    v_performance_category VARCHAR;
    v_optimization_priority VARCHAR;
    v_recommendations TEXT;
    v_index_suggestions TEXT[];
    v_query_optimization_hints TEXT[];
    v_complexity_score INTEGER;
    v_selectivity_score DECIMAL;
    v_join_complexity INTEGER;
BEGIN
    -- Generate a unique query ID based on query text hash
    v_query_id := ('x' || substr(md5(p_query_text), 1, 8))::bit(32)::bigint;
    
    -- Calculate query complexity score
    v_complexity_score := 0;
    
    -- Count SELECT statements
    v_complexity_score := v_complexity_score + (length(p_query_text) - length(replace(p_query_text, 'SELECT', ''))) / 6;
    
    -- Count JOIN statements
    v_join_complexity := (length(p_query_text) - length(replace(p_query_text, 'JOIN', ''))) / 4;
    v_complexity_score := v_complexity_score + v_join_complexity * 2;
    
    -- Count WHERE clauses
    v_complexity_score := v_complexity_score + (length(p_query_text) - length(replace(p_query_text, 'WHERE', ''))) / 5;
    
    -- Count subqueries
    v_complexity_score := v_complexity_score + (length(p_query_text) - length(replace(p_query_text, '(', ''))) * 0.5;
    
    -- Calculate selectivity score
    IF p_rows_examined > 0 THEN
        v_selectivity_score := (p_rows_returned::DECIMAL / p_rows_examined) * 100;
    ELSE
        v_selectivity_score := 100;
    END IF;
    
    -- Calculate performance score (0-100)
    v_performance_score := 100;
    
    -- Deduct points for slow execution
    IF p_execution_time_ms > 1000 THEN
        v_performance_score := v_performance_score - 30;
    ELSIF p_execution_time_ms > 500 THEN
        v_performance_score := v_performance_score - 20;
    ELSIF p_execution_time_ms > 100 THEN
        v_performance_score := v_performance_score - 10;
    END IF;
    
    -- Deduct points for poor selectivity
    IF v_selectivity_score < 10 THEN
        v_performance_score := v_performance_score - 25;
    ELSIF v_selectivity_score < 50 THEN
        v_performance_score := v_performance_score - 15;
    END IF;
    
    -- Deduct points for high complexity
    IF v_complexity_score > 10 THEN
        v_performance_score := v_performance_score - 20;
    ELSIF v_complexity_score > 5 THEN
        v_performance_score := v_performance_score - 10;
    END IF;
    
    -- Determine performance category
    IF v_performance_score >= 80 THEN
        v_performance_category := 'EXCELLENT';
        v_optimization_priority := 'LOW';
        v_recommendations := 'Query performance is excellent. No optimization needed.';
    ELSIF v_performance_score >= 60 THEN
        v_performance_category := 'GOOD';
        v_optimization_priority := 'LOW';
        v_recommendations := 'Query performance is good. Monitor for future optimization opportunities.';
    ELSIF v_performance_score >= 40 THEN
        v_performance_category := 'FAIR';
        v_optimization_priority := 'MEDIUM';
        v_recommendations := 'Query performance is fair. Consider optimization to improve performance.';
    ELSIF v_performance_score >= 20 THEN
        v_performance_category := 'POOR';
        v_optimization_priority := 'HIGH';
        v_recommendations := 'Query performance is poor. Immediate optimization recommended.';
    ELSE
        v_performance_category := 'CRITICAL';
        v_optimization_priority := 'URGENT';
        v_recommendations := 'Query performance is critical. Urgent optimization required.';
    END IF;
    
    -- Generate index suggestions
    v_index_suggestions := ARRAY[]::TEXT[];
    
    -- Suggest indexes for WHERE clauses
    IF p_query_text ~* 'WHERE.*=' THEN
        v_index_suggestions := array_append(v_index_suggestions, 'Consider adding indexes on columns used in WHERE clauses');
    END IF;
    
    -- Suggest indexes for JOIN conditions
    IF p_query_text ~* 'JOIN.*ON' THEN
        v_index_suggestions := array_append(v_index_suggestions, 'Consider adding indexes on columns used in JOIN conditions');
    END IF;
    
    -- Suggest indexes for ORDER BY clauses
    IF p_query_text ~* 'ORDER BY' THEN
        v_index_suggestions := array_append(v_index_suggestions, 'Consider adding indexes on columns used in ORDER BY clauses');
    END IF;
    
    -- Generate query optimization hints
    v_query_optimization_hints := ARRAY[]::TEXT[];
    
    -- Suggest query structure optimization
    IF v_complexity_score > 5 THEN
        v_query_optimization_hints := array_append(v_query_optimization_hints, 'Consider breaking down complex queries into simpler parts');
    END IF;
    
    -- Suggest subquery optimization
    IF p_query_text ~* '\(.*SELECT.*\)' THEN
        v_query_optimization_hints := array_append(v_query_optimization_hints, 'Consider using JOINs instead of subqueries for better performance');
    END IF;
    
    -- Suggest LIMIT optimization
    IF p_query_text ~* 'SELECT.*FROM' AND p_query_text !~* 'LIMIT' THEN
        v_query_optimization_hints := array_append(v_query_optimization_hints, 'Consider adding LIMIT clause to restrict result set size');
    END IF;
    
    -- Return analysis results
    RETURN QUERY
    SELECT 
        v_query_id,
        v_performance_score,
        v_performance_category,
        v_optimization_priority,
        v_recommendations,
        v_index_suggestions,
        v_query_optimization_hints;
END;
$$ LANGUAGE plpgsql;

-- 3. Create a function to log query performance
CREATE OR REPLACE FUNCTION log_query_performance(
    p_query_text TEXT,
    p_execution_time_ms DECIMAL,
    p_rows_returned BIGINT,
    p_rows_examined BIGINT,
    p_user_id UUID DEFAULT NULL,
    p_session_id VARCHAR DEFAULT NULL,
    p_request_id VARCHAR DEFAULT NULL
) RETURNS BIGINT AS $$
DECLARE
    v_query_id BIGINT;
    v_performance_score DECIMAL;
    v_performance_category VARCHAR;
    v_optimization_priority VARCHAR;
    v_recommendations TEXT;
    v_index_suggestions TEXT[];
    v_query_optimization_hints TEXT[];
    v_index_usage_score DECIMAL;
    v_cache_hit_ratio DECIMAL;
    v_complexity_score INTEGER;
    v_log_id BIGINT;
BEGIN
    -- Analyze query performance
    SELECT 
        query_id,
        performance_score,
        performance_category,
        optimization_priority,
        recommendations,
        index_suggestions,
        query_optimization_hints
    INTO 
        v_query_id,
        v_performance_score,
        v_performance_category,
        v_optimization_priority,
        v_recommendations,
        v_index_suggestions,
        v_query_optimization_hints
    FROM analyze_query_performance(p_query_text, p_execution_time_ms, p_rows_returned, p_rows_examined);
    
    -- Calculate index usage score
    IF p_rows_examined > 0 THEN
        v_index_usage_score := (p_rows_returned::DECIMAL / p_rows_examined) * 100;
    ELSE
        v_index_usage_score := 100;
    END IF;
    
    -- Calculate cache hit ratio (simplified)
    v_cache_hit_ratio := 95.0; -- This would be calculated from actual cache statistics
    
    -- Calculate complexity score
    v_complexity_score := 0;
    v_complexity_score := v_complexity_score + (length(p_query_text) - length(replace(p_query_text, 'SELECT', ''))) / 6;
    v_complexity_score := v_complexity_score + (length(p_query_text) - length(replace(p_query_text, 'JOIN', ''))) / 4 * 2;
    v_complexity_score := v_complexity_score + (length(p_query_text) - length(replace(p_query_text, 'WHERE', ''))) / 5;
    v_complexity_score := v_complexity_score + (length(p_query_text) - length(replace(p_query_text, '(', ''))) * 0.5;
    
    -- Insert performance log entry
    INSERT INTO query_performance_log (
        query_id,
        query_text,
        execution_time_ms,
        rows_returned,
        rows_examined,
        index_usage_score,
        cache_hit_ratio,
        query_complexity_score,
        performance_category,
        optimization_priority,
        recommendations,
        user_id,
        session_id,
        request_id
    ) VALUES (
        v_query_id,
        p_query_text,
        p_execution_time_ms,
        p_rows_returned,
        p_rows_examined,
        v_index_usage_score,
        v_cache_hit_ratio,
        v_complexity_score,
        v_performance_category,
        v_optimization_priority,
        v_recommendations,
        p_user_id,
        p_session_id,
        p_request_id
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to get query performance statistics
CREATE OR REPLACE FUNCTION get_query_performance_stats(
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    total_queries BIGINT,
    avg_execution_time_ms DECIMAL,
    max_execution_time_ms DECIMAL,
    min_execution_time_ms DECIMAL,
    total_rows_returned BIGINT,
    total_rows_examined BIGINT,
    avg_index_usage_score DECIMAL,
    avg_cache_hit_ratio DECIMAL,
    avg_complexity_score DECIMAL,
    performance_distribution JSONB,
    top_slow_queries JSONB,
    optimization_opportunities JSONB
) AS $$
DECLARE
    v_total_queries BIGINT;
    v_avg_execution_time_ms DECIMAL;
    v_max_execution_time_ms DECIMAL;
    v_min_execution_time_ms DECIMAL;
    v_total_rows_returned BIGINT;
    v_total_rows_examined BIGINT;
    v_avg_index_usage_score DECIMAL;
    v_avg_cache_hit_ratio DECIMAL;
    v_avg_complexity_score DECIMAL;
    v_performance_distribution JSONB;
    v_top_slow_queries JSONB;
    v_optimization_opportunities JSONB;
BEGIN
    -- Get basic statistics
    SELECT 
        COUNT(*),
        ROUND(AVG(execution_time_ms), 2),
        ROUND(MAX(execution_time_ms), 2),
        ROUND(MIN(execution_time_ms), 2),
        SUM(rows_returned),
        SUM(rows_examined),
        ROUND(AVG(index_usage_score), 2),
        ROUND(AVG(cache_hit_ratio), 2),
        ROUND(AVG(query_complexity_score), 2)
    INTO 
        v_total_queries,
        v_avg_execution_time_ms,
        v_max_execution_time_ms,
        v_min_execution_time_ms,
        v_total_rows_returned,
        v_total_rows_examined,
        v_avg_index_usage_score,
        v_avg_cache_hit_ratio,
        v_avg_complexity_score
    FROM query_performance_log
    WHERE executed_at >= NOW() - INTERVAL '1 hour' * p_hours_back;
    
    -- Get performance distribution
    SELECT jsonb_build_object(
        'excellent', COUNT(CASE WHEN performance_category = 'EXCELLENT' THEN 1 END),
        'good', COUNT(CASE WHEN performance_category = 'GOOD' THEN 1 END),
        'fair', COUNT(CASE WHEN performance_category = 'FAIR' THEN 1 END),
        'poor', COUNT(CASE WHEN performance_category = 'POOR' THEN 1 END),
        'critical', COUNT(CASE WHEN performance_category = 'CRITICAL' THEN 1 END)
    )
    INTO v_performance_distribution
    FROM query_performance_log
    WHERE executed_at >= NOW() - INTERVAL '1 hour' * p_hours_back;
    
    -- Get top slow queries
    SELECT jsonb_agg(
        jsonb_build_object(
            'query_id', query_id,
            'query_text', LEFT(query_text, 100),
            'execution_time_ms', execution_time_ms,
            'performance_category', performance_category,
            'optimization_priority', optimization_priority
        )
    )
    INTO v_top_slow_queries
    FROM (
        SELECT 
            query_id,
            query_text,
            execution_time_ms,
            performance_category,
            optimization_priority
        FROM query_performance_log
        WHERE executed_at >= NOW() - INTERVAL '1 hour' * p_hours_back
        ORDER BY execution_time_ms DESC
        LIMIT 10
    ) slow_queries;
    
    -- Get optimization opportunities
    SELECT jsonb_agg(
        jsonb_build_object(
            'query_id', query_id,
            'query_text', LEFT(query_text, 100),
            'performance_category', performance_category,
            'optimization_priority', optimization_priority,
            'recommendations', recommendations
        )
    )
    INTO v_optimization_opportunities
    FROM (
        SELECT 
            query_id,
            query_text,
            performance_category,
            optimization_priority,
            recommendations
        FROM query_performance_log
        WHERE executed_at >= NOW() - INTERVAL '1 hour' * p_hours_back
        AND optimization_priority IN ('HIGH', 'URGENT')
        ORDER BY 
            CASE optimization_priority 
                WHEN 'URGENT' THEN 1 
                WHEN 'HIGH' THEN 2 
                ELSE 3 
            END,
            execution_time_ms DESC
        LIMIT 20
    ) optimization_queries;
    
    RETURN QUERY
    SELECT 
        v_total_queries,
        v_avg_execution_time_ms,
        v_max_execution_time_ms,
        v_min_execution_time_ms,
        v_total_rows_returned,
        v_total_rows_examined,
        v_avg_index_usage_score,
        v_avg_cache_hit_ratio,
        v_avg_complexity_score,
        v_performance_distribution,
        v_top_slow_queries,
        v_optimization_opportunities;
END;
$$ LANGUAGE plpgsql;

-- 5. Create a function to get query performance trends
CREATE OR REPLACE FUNCTION get_query_performance_trends(
    p_hours_back INTEGER DEFAULT 168 -- 7 days
) RETURNS TABLE (
    hour_bucket TIMESTAMP WITH TIME ZONE,
    total_queries BIGINT,
    avg_execution_time_ms DECIMAL,
    avg_index_usage_score DECIMAL,
    avg_cache_hit_ratio DECIMAL,
    performance_score DECIMAL,
    slow_query_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        date_trunc('hour', executed_at) as hour_bucket,
        COUNT(*) as total_queries,
        ROUND(AVG(execution_time_ms), 2) as avg_execution_time_ms,
        ROUND(AVG(index_usage_score), 2) as avg_index_usage_score,
        ROUND(AVG(cache_hit_ratio), 2) as avg_cache_hit_ratio,
        ROUND(AVG(
            CASE 
                WHEN execution_time_ms < 100 THEN 100
                WHEN execution_time_ms < 500 THEN 80
                WHEN execution_time_ms < 1000 THEN 60
                WHEN execution_time_ms < 2000 THEN 40
                ELSE 20
            END
        ), 2) as performance_score,
        COUNT(CASE WHEN execution_time_ms > 1000 THEN 1 END) as slow_query_count
    FROM query_performance_log
    WHERE executed_at >= NOW() - INTERVAL '1 hour' * p_hours_back
    GROUP BY date_trunc('hour', executed_at)
    ORDER BY hour_bucket DESC;
END;
$$ LANGUAGE plpgsql;

-- 6. Create a function to get query performance alerts
CREATE OR REPLACE FUNCTION get_query_performance_alerts(
    p_hours_back INTEGER DEFAULT 1
) RETURNS TABLE (
    alert_id SERIAL,
    alert_type VARCHAR(50),
    alert_level VARCHAR(20),
    alert_message TEXT,
    query_id BIGINT,
    query_text TEXT,
    execution_time_ms DECIMAL,
    performance_category VARCHAR,
    recommendations TEXT,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    -- Slow query alerts
    SELECT 
        qpl.id as alert_id,
        'SLOW_QUERY' as alert_type,
        CASE 
            WHEN qpl.execution_time_ms > 5000 THEN 'CRITICAL'
            WHEN qpl.execution_time_ms > 2000 THEN 'HIGH'
            WHEN qpl.execution_time_ms > 1000 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'Slow query detected: ' || ROUND(qpl.execution_time_ms, 2) || 'ms execution time' as alert_message,
        qpl.query_id,
        LEFT(qpl.query_text, 200) as query_text,
        qpl.execution_time_ms,
        qpl.performance_category,
        qpl.recommendations,
        qpl.executed_at as created_at
    FROM query_performance_log qpl
    WHERE qpl.executed_at >= NOW() - INTERVAL '1 hour' * p_hours_back
    AND qpl.execution_time_ms > 1000
    
    UNION ALL
    
    -- Poor performance alerts
    SELECT 
        qpl.id as alert_id,
        'POOR_PERFORMANCE' as alert_type,
        CASE 
            WHEN qpl.performance_category = 'CRITICAL' THEN 'CRITICAL'
            WHEN qpl.performance_category = 'POOR' THEN 'HIGH'
            ELSE 'MEDIUM'
        END as alert_level,
        'Poor query performance detected: ' || qpl.performance_category as alert_message,
        qpl.query_id,
        LEFT(qpl.query_text, 200) as query_text,
        qpl.execution_time_ms,
        qpl.performance_category,
        qpl.recommendations,
        qpl.executed_at as created_at
    FROM query_performance_log qpl
    WHERE qpl.executed_at >= NOW() - INTERVAL '1 hour' * p_hours_back
    AND qpl.performance_category IN ('POOR', 'CRITICAL')
    
    UNION ALL
    
    -- High complexity alerts
    SELECT 
        qpl.id as alert_id,
        'HIGH_COMPLEXITY' as alert_type,
        CASE 
            WHEN qpl.query_complexity_score > 20 THEN 'HIGH'
            WHEN qpl.query_complexity_score > 10 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'High query complexity detected: ' || qpl.query_complexity_score || ' complexity score' as alert_message,
        qpl.query_id,
        LEFT(qpl.query_text, 200) as query_text,
        qpl.execution_time_ms,
        qpl.performance_category,
        qpl.recommendations,
        qpl.executed_at as created_at
    FROM query_performance_log qpl
    WHERE qpl.executed_at >= NOW() - INTERVAL '1 hour' * p_hours_back
    AND qpl.query_complexity_score > 10
    
    ORDER BY created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 7. Create a function to get query performance dashboard
CREATE OR REPLACE FUNCTION get_query_performance_dashboard() 
RETURNS TABLE (
    metric_name TEXT,
    current_value TEXT,
    target_value TEXT,
    status TEXT,
    trend TEXT,
    recommendations TEXT
) AS $$
DECLARE
    v_total_queries BIGINT;
    v_avg_execution_time_ms DECIMAL;
    v_slow_query_count BIGINT;
    v_avg_index_usage_score DECIMAL;
    v_avg_cache_hit_ratio DECIMAL;
    v_performance_score DECIMAL;
BEGIN
    -- Get current metrics
    SELECT 
        COUNT(*),
        ROUND(AVG(execution_time_ms), 2),
        COUNT(CASE WHEN execution_time_ms > 1000 THEN 1 END),
        ROUND(AVG(index_usage_score), 2),
        ROUND(AVG(cache_hit_ratio), 2),
        ROUND(AVG(
            CASE 
                WHEN execution_time_ms < 100 THEN 100
                WHEN execution_time_ms < 500 THEN 80
                WHEN execution_time_ms < 1000 THEN 60
                WHEN execution_time_ms < 2000 THEN 40
                ELSE 20
            END
        ), 2)
    INTO 
        v_total_queries,
        v_avg_execution_time_ms,
        v_slow_query_count,
        v_avg_index_usage_score,
        v_avg_cache_hit_ratio,
        v_performance_score
    FROM query_performance_log
    WHERE executed_at >= NOW() - INTERVAL '1 hour';
    
    RETURN QUERY
    -- Total Queries
    SELECT 
        'Total Queries (Last Hour)' as metric_name,
        v_total_queries::TEXT as current_value,
        'N/A' as target_value,
        CASE 
            WHEN v_total_queries > 1000 THEN 'HIGH'
            WHEN v_total_queries > 500 THEN 'MEDIUM'
            ELSE 'LOW'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_total_queries > 1000 THEN 'High query volume - monitor for performance impact'
            WHEN v_total_queries > 500 THEN 'Moderate query volume - monitor performance'
            ELSE 'Low query volume - normal operation'
        END as recommendations
    
    UNION ALL
    
    -- Average Execution Time
    SELECT 
        'Average Execution Time' as metric_name,
        v_avg_execution_time_ms::TEXT || ' ms' as current_value,
        '100 ms' as target_value,
        CASE 
            WHEN v_avg_execution_time_ms > 1000 THEN 'CRITICAL'
            WHEN v_avg_execution_time_ms > 500 THEN 'WARNING'
            WHEN v_avg_execution_time_ms > 100 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_avg_execution_time_ms > 1000 THEN 'Critical: Optimize slow queries immediately'
            WHEN v_avg_execution_time_ms > 500 THEN 'Warning: Review and optimize slow queries'
            WHEN v_avg_execution_time_ms > 100 THEN 'Fair: Monitor query performance'
            ELSE 'Good: Query performance is acceptable'
        END as recommendations
    
    UNION ALL
    
    -- Slow Query Count
    SELECT 
        'Slow Queries (Last Hour)' as metric_name,
        v_slow_query_count::TEXT as current_value,
        '0' as target_value,
        CASE 
            WHEN v_slow_query_count > 50 THEN 'CRITICAL'
            WHEN v_slow_query_count > 20 THEN 'WARNING'
            WHEN v_slow_query_count > 5 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_slow_query_count > 50 THEN 'Critical: Many slow queries detected'
            WHEN v_slow_query_count > 20 THEN 'Warning: Several slow queries detected'
            WHEN v_slow_query_count > 5 THEN 'Fair: Some slow queries detected'
            ELSE 'Good: No slow queries detected'
        END as recommendations
    
    UNION ALL
    
    -- Index Usage Score
    SELECT 
        'Index Usage Score' as metric_name,
        v_avg_index_usage_score::TEXT || '%' as current_value,
        '80%' as target_value,
        CASE 
            WHEN v_avg_index_usage_score < 50 THEN 'CRITICAL'
            WHEN v_avg_index_usage_score < 70 THEN 'WARNING'
            WHEN v_avg_index_usage_score < 80 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_avg_index_usage_score < 50 THEN 'Critical: Poor index usage - add missing indexes'
            WHEN v_avg_index_usage_score < 70 THEN 'Warning: Low index usage - review indexes'
            WHEN v_avg_index_usage_score < 80 THEN 'Fair: Moderate index usage - optimize indexes'
            ELSE 'Good: Index usage is efficient'
        END as recommendations
    
    UNION ALL
    
    -- Cache Hit Ratio
    SELECT 
        'Cache Hit Ratio' as metric_name,
        v_avg_cache_hit_ratio::TEXT || '%' as current_value,
        '95%' as target_value,
        CASE 
            WHEN v_avg_cache_hit_ratio < 80 THEN 'CRITICAL'
            WHEN v_avg_cache_hit_ratio < 90 THEN 'WARNING'
            WHEN v_avg_cache_hit_ratio < 95 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_avg_cache_hit_ratio < 80 THEN 'Critical: Poor cache performance - increase cache size'
            WHEN v_avg_cache_hit_ratio < 90 THEN 'Warning: Low cache hit ratio - optimize cache'
            WHEN v_avg_cache_hit_ratio < 95 THEN 'Fair: Moderate cache performance - monitor cache'
            ELSE 'Good: Cache performance is excellent'
        END as recommendations
    
    UNION ALL
    
    -- Overall Performance Score
    SELECT 
        'Overall Performance Score' as metric_name,
        v_performance_score::TEXT || '%' as current_value,
        '80%' as target_value,
        CASE 
            WHEN v_performance_score < 40 THEN 'CRITICAL'
            WHEN v_performance_score < 60 THEN 'WARNING'
            WHEN v_performance_score < 80 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_performance_score < 40 THEN 'Critical: Overall performance is poor - immediate optimization needed'
            WHEN v_performance_score < 60 THEN 'Warning: Overall performance needs improvement'
            WHEN v_performance_score < 80 THEN 'Fair: Overall performance is acceptable'
            ELSE 'Good: Overall performance is excellent'
        END as recommendations;
END;
$$ LANGUAGE plpgsql;

-- 8. Create a function to cleanup old performance logs
CREATE OR REPLACE FUNCTION cleanup_query_performance_logs(
    p_days_to_keep INTEGER DEFAULT 30
) RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM query_performance_log
    WHERE executed_at < NOW() - INTERVAL '1 day' * p_days_to_keep;
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a function to get query performance insights
CREATE OR REPLACE FUNCTION get_query_performance_insights() 
RETURNS TABLE (
    insight_type TEXT,
    insight_title TEXT,
    insight_description TEXT,
    insight_priority VARCHAR(20),
    insight_recommendations TEXT,
    affected_queries BIGINT,
    potential_improvement DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    -- Slow query insights
    SELECT 
        'SLOW_QUERIES' as insight_type,
        'Slow Query Optimization' as insight_title,
        'Queries with execution time > 1000ms detected' as insight_description,
        'HIGH' as insight_priority,
        'Optimize slow queries by adding indexes, rewriting queries, or using query hints' as insight_recommendations,
        COUNT(*) as affected_queries,
        ROUND(AVG(execution_time_ms - 1000), 2) as potential_improvement
    FROM query_performance_log
    WHERE executed_at >= NOW() - INTERVAL '24 hours'
    AND execution_time_ms > 1000
    
    UNION ALL
    
    -- Index usage insights
    SELECT 
        'INDEX_OPTIMIZATION' as insight_type,
        'Index Usage Optimization' as insight_title,
        'Queries with poor index usage detected' as insight_description,
        'MEDIUM' as insight_priority,
        'Add missing indexes or optimize existing indexes for better performance' as insight_recommendations,
        COUNT(*) as affected_queries,
        ROUND(AVG(80 - index_usage_score), 2) as potential_improvement
    FROM query_performance_log
    WHERE executed_at >= NOW() - INTERVAL '24 hours'
    AND index_usage_score < 80
    
    UNION ALL
    
    -- Query complexity insights
    SELECT 
        'QUERY_COMPLEXITY' as insight_type,
        'Query Complexity Reduction' as insight_title,
        'Highly complex queries detected' as insight_description,
        'MEDIUM' as insight_priority,
        'Simplify complex queries by breaking them into smaller parts or using views' as insight_recommendations,
        COUNT(*) as affected_queries,
        ROUND(AVG(query_complexity_score * 10), 2) as potential_improvement
    FROM query_performance_log
    WHERE executed_at >= NOW() - INTERVAL '24 hours'
    AND query_complexity_score > 10
    
    UNION ALL
    
    -- Cache performance insights
    SELECT 
        'CACHE_OPTIMIZATION' as insight_type,
        'Cache Performance Optimization' as insight_title,
        'Queries with poor cache hit ratio detected' as insight_description,
        'LOW' as insight_priority,
        'Optimize cache configuration or increase cache size for better performance' as insight_recommendations,
        COUNT(*) as affected_queries,
        ROUND(AVG(95 - cache_hit_ratio), 2) as potential_improvement
    FROM query_performance_log
    WHERE executed_at >= NOW() - INTERVAL '24 hours'
    AND cache_hit_ratio < 95;
END;
$$ LANGUAGE plpgsql;

-- 10. Create a function to validate query performance monitoring setup
CREATE OR REPLACE FUNCTION validate_query_performance_monitoring_setup() 
RETURNS TABLE (
    component TEXT,
    status TEXT,
    details TEXT,
    recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check if query_performance_log table exists
    SELECT 
        'Query Performance Log Table' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'query_performance_log') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Table for storing query performance logs' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'query_performance_log') 
            THEN 'Table exists and ready for use' 
            ELSE 'Create query_performance_log table' 
        END as recommendation
    
    UNION ALL
    
    -- Check if performance functions exist
    SELECT 
        'Performance Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'analyze_query_performance') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for analyzing query performance' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'analyze_query_performance') 
            THEN 'All performance functions are available' 
            ELSE 'Create performance analysis functions' 
        END as recommendation
    
    UNION ALL
    
    -- Check if logging function exists
    SELECT 
        'Logging Function' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'log_query_performance') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Function for logging query performance' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'log_query_performance') 
            THEN 'Logging function is ready' 
            ELSE 'Create query performance logging function' 
        END as recommendation;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_query_performance_log_executed_at ON query_performance_log(executed_at);
CREATE INDEX IF NOT EXISTS idx_query_performance_log_query_id ON query_performance_log(query_id);
CREATE INDEX IF NOT EXISTS idx_query_performance_log_performance_category ON query_performance_log(performance_category);
CREATE INDEX IF NOT EXISTS idx_query_performance_log_execution_time ON query_performance_log(execution_time_ms);
CREATE INDEX IF NOT EXISTS idx_query_performance_log_user_id ON query_performance_log(user_id);

-- Create a view for easy query performance dashboard access
CREATE OR REPLACE VIEW query_performance_dashboard AS
SELECT 
    'Query Performance Overview' as metric_name,
    (SELECT COUNT(*)::TEXT FROM query_performance_log WHERE executed_at >= NOW() - INTERVAL '1 hour') as current_value,
    'N/A' as target_value,
    (SELECT 
        CASE 
            WHEN AVG(execution_time_ms) > 1000 THEN 'CRITICAL'
            WHEN AVG(execution_time_ms) > 500 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM query_performance_log WHERE executed_at >= NOW() - INTERVAL '1 hour') as status
UNION ALL
SELECT 
    'Average Execution Time' as metric_name,
    (SELECT ROUND(AVG(execution_time_ms), 2)::TEXT || ' ms' FROM query_performance_log WHERE executed_at >= NOW() - INTERVAL '1 hour') as current_value,
    '100 ms' as target_value,
    (SELECT 
        CASE 
            WHEN AVG(execution_time_ms) > 1000 THEN 'CRITICAL'
            WHEN AVG(execution_time_ms) > 500 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM query_performance_log WHERE executed_at >= NOW() - INTERVAL '1 hour') as status
UNION ALL
SELECT 
    'Slow Queries' as metric_name,
    (SELECT COUNT(*)::TEXT FROM query_performance_log WHERE executed_at >= NOW() - INTERVAL '1 hour' AND execution_time_ms > 1000) as current_value,
    '0' as target_value,
    (SELECT 
        CASE 
            WHEN COUNT(*) > 50 THEN 'CRITICAL'
            WHEN COUNT(*) > 20 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM query_performance_log WHERE executed_at >= NOW() - INTERVAL '1 hour' AND execution_time_ms > 1000) as status;

-- Grant permissions
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO authenticated;
GRANT SELECT ON query_performance_dashboard TO authenticated;
GRANT SELECT, INSERT ON query_performance_log TO authenticated;

-- Initial setup message
DO $$
BEGIN
    RAISE NOTICE 'Query performance monitoring setup completed successfully!';
    RAISE NOTICE 'Total functions created: 10';
    RAISE NOTICE 'Total tables created: 1';
    RAISE NOTICE 'Total views created: 1';
    RAISE NOTICE 'Total indexes created: 5';
    RAISE NOTICE 'All query performance monitoring tools are now available.';
    RAISE NOTICE 'Use query_performance_dashboard view to access current performance metrics.';
    RAISE NOTICE 'Call log_query_performance() to log query performance data.';
    RAISE NOTICE 'Call get_query_performance_stats() to get performance statistics.';
END $$;
