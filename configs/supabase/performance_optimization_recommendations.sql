-- Performance Optimization Recommendations for Business Classification Platform
-- This script provides comprehensive performance optimization recommendations based on monitoring data

-- 1. Create a performance recommendations table
CREATE TABLE IF NOT EXISTS performance_recommendations (
    id SERIAL PRIMARY KEY,
    recommendation_id VARCHAR(255) UNIQUE NOT NULL,
    recommendation_type VARCHAR(100) NOT NULL,
    recommendation_category VARCHAR(50) NOT NULL,
    recommendation_title TEXT NOT NULL,
    recommendation_description TEXT NOT NULL,
    recommendation_priority VARCHAR(20) NOT NULL,
    recommendation_impact VARCHAR(20) NOT NULL,
    recommendation_effort VARCHAR(20) NOT NULL,
    recommendation_benefit TEXT NOT NULL,
    recommendation_implementation TEXT NOT NULL,
    recommendation_validation TEXT NOT NULL,
    affected_systems TEXT[],
    related_metrics TEXT[],
    estimated_improvement_percentage NUMERIC,
    estimated_implementation_time_hours NUMERIC,
    prerequisites TEXT[],
    dependencies TEXT[],
    status VARCHAR(20) DEFAULT 'pending',
    assigned_to VARCHAR(255),
    assigned_at TIMESTAMPTZ,
    implemented_at TIMESTAMPTZ,
    implementation_notes TEXT,
    validation_results TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create a function to generate database performance recommendations
CREATE OR REPLACE FUNCTION generate_database_performance_recommendations() 
RETURNS TABLE (
    recommendation_id VARCHAR(255),
    recommendation_type VARCHAR(100),
    recommendation_category VARCHAR(50),
    recommendation_title TEXT,
    recommendation_description TEXT,
    recommendation_priority VARCHAR(20),
    recommendation_impact VARCHAR(20),
    recommendation_effort VARCHAR(20),
    recommendation_benefit TEXT,
    recommendation_implementation TEXT,
    recommendation_validation TEXT,
    affected_systems TEXT[],
    related_metrics TEXT[],
    estimated_improvement_percentage NUMERIC,
    estimated_implementation_time_hours NUMERIC,
    prerequisites TEXT[],
    dependencies TEXT[]
) AS $$
DECLARE
    v_db_size_bytes BIGINT;
    v_active_connections INT;
    v_avg_query_time_ms NUMERIC;
    v_connection_utilization NUMERIC;
    v_max_connections INT;
    v_slow_queries_count INT;
    v_unused_indexes_count INT;
    v_table_bloat_percentage NUMERIC;
BEGIN
    -- Get current database metrics
    SELECT 
        pg_database_size(current_database()),
        (SELECT count(*)::INT FROM pg_stat_activity WHERE datname = current_database()),
        (SELECT AVG(mean_exec_time) FROM pg_stat_statements WHERE mean_exec_time > 0),
        (SELECT count(*)::NUMERIC FROM pg_stat_activity WHERE datname = current_database()) * 100 / (SELECT setting::NUMERIC FROM pg_settings WHERE name = 'max_connections'),
        (SELECT setting::INT FROM pg_settings WHERE name = 'max_connections'),
        (SELECT count(*)::INT FROM pg_stat_statements WHERE mean_exec_time > 1000),
        (SELECT count(*)::INT FROM pg_stat_user_indexes WHERE idx_scan = 0),
        (SELECT AVG(bloat_percentage) FROM (
            SELECT 
                schemaname,
                tablename,
                ROUND(
                    (pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename))::NUMERIC / 
                    NULLIF(pg_total_relation_size(schemaname||'.'||tablename), 0) * 100, 2
                ) as bloat_percentage
            FROM pg_tables 
            WHERE schemaname = 'public'
        ) bloat_analysis)
    INTO v_db_size_bytes, v_active_connections, v_avg_query_time_ms, v_connection_utilization, v_max_connections, v_slow_queries_count, v_unused_indexes_count, v_table_bloat_percentage;
    
    -- Database size optimization recommendations
    IF v_db_size_bytes > 400 * 1024 * 1024 THEN -- 400MB threshold
        RETURN QUERY SELECT 
            'DB_SIZE_OPT_001'::VARCHAR(255),
            'DATABASE_SIZE'::VARCHAR(100),
            'STORAGE'::VARCHAR(50),
            'Optimize Database Storage Usage'::TEXT,
            'Database size is approaching free tier limits. Implement storage optimization strategies.'::TEXT,
            'HIGH'::VARCHAR(20),
            'HIGH'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'Reduce storage costs and improve performance'::TEXT,
            '1. Archive old classification data older than 6 months
2. Implement data compression for historical data
3. Remove unused tables and columns
4. Optimize data types and indexes
5. Consider upgrading to paid plan for larger storage'::TEXT,
            'Monitor database size reduction and query performance improvement'::TEXT,
            ARRAY['database', 'storage']::TEXT[],
            ARRAY['database_size_bytes', 'storage_usage_percentage']::TEXT[],
            25.0::NUMERIC,
            8.0::NUMERIC,
            ARRAY['Database backup', 'Data analysis tools']::TEXT[],
            ARRAY['Data archiving system', 'Compression tools']::TEXT[];
    END IF;
    
    -- Connection pool optimization recommendations
    IF v_connection_utilization > 70 THEN
        RETURN QUERY SELECT 
            'DB_CONN_OPT_001'::VARCHAR(255),
            'CONNECTION_POOL'::VARCHAR(100),
            'CONNECTIONS'::VARCHAR(50),
            'Optimize Database Connection Pool'::TEXT,
            'Connection utilization is high. Optimize connection pooling and query patterns.'::TEXT,
            'HIGH'::VARCHAR(20),
            'HIGH'::VARCHAR(20),
            'LOW'::VARCHAR(20),
            'Improve connection efficiency and reduce connection overhead'::TEXT,
            '1. Implement connection pooling in application
2. Optimize long-running queries
3. Use prepared statements
4. Implement connection timeouts
5. Monitor and limit concurrent connections'::TEXT,
            'Monitor connection utilization and query performance'::TEXT,
            ARRAY['database', 'connections', 'application']::TEXT[],
            ARRAY['connection_utilization_percentage', 'active_connections']::TEXT[],
            30.0::NUMERIC,
            4.0::NUMERIC,
            ARRAY['Connection pool library', 'Query analysis tools']::TEXT[],
            ARRAY['Application configuration', 'Database monitoring']::TEXT[];
    END IF;
    
    -- Query performance optimization recommendations
    IF v_avg_query_time_ms > 500 THEN
        RETURN QUERY SELECT 
            'DB_QUERY_OPT_001'::VARCHAR(255),
            'QUERY_PERFORMANCE'::VARCHAR(100),
            'PERFORMANCE'::VARCHAR(50),
            'Optimize Database Query Performance'::TEXT,
            'Average query time is high. Implement query optimization strategies.'::TEXT,
            'CRITICAL'::VARCHAR(20),
            'HIGH'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'Significantly improve query response times and user experience'::TEXT,
            '1. Analyze and optimize slow queries
2. Add missing database indexes
3. Use query execution plans to identify bottlenecks
4. Implement query caching
5. Consider database partitioning for large tables'::TEXT,
            'Monitor query execution times and database performance metrics'::TEXT,
            ARRAY['database', 'queries', 'indexes']::TEXT[],
            ARRAY['avg_query_time_ms', 'slow_queries_count']::TEXT[],
            50.0::NUMERIC,
            12.0::NUMERIC,
            ARRAY['Query analysis tools', 'Database monitoring']::TEXT[],
            ARRAY['Index optimization', 'Query caching system']::TEXT[];
    END IF;
    
    -- Index optimization recommendations
    IF v_unused_indexes_count > 5 THEN
        RETURN QUERY SELECT 
            'DB_INDEX_OPT_001'::VARCHAR(255),
            'INDEX_OPTIMIZATION'::VARCHAR(100),
            'PERFORMANCE'::VARCHAR(50),
            'Optimize Database Indexes'::TEXT,
            'Multiple unused indexes detected. Remove unused indexes and add missing ones.'::TEXT,
            'MEDIUM'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'LOW'::VARCHAR(20),
            'Reduce storage overhead and improve write performance'::TEXT,
            '1. Identify and remove unused indexes
2. Analyze query patterns to add missing indexes
3. Optimize index column order
4. Consider partial indexes for filtered queries
5. Monitor index usage after changes'::TEXT,
            'Monitor index usage statistics and query performance'::TEXT,
            ARRAY['database', 'indexes']::TEXT[],
            ARRAY['unused_indexes_count', 'index_usage_percentage']::TEXT[],
            15.0::NUMERIC,
            6.0::NUMERIC,
            ARRAY['Index analysis tools', 'Query pattern analysis']::TEXT[],
            ARRAY['Database monitoring', 'Query optimization']::TEXT[];
    END IF;
    
    -- Table bloat optimization recommendations
    IF v_table_bloat_percentage > 20 THEN
        RETURN QUERY SELECT 
            'DB_BLOAT_OPT_001'::VARCHAR(255),
            'TABLE_BLOAT'::VARCHAR(100),
            'STORAGE'::VARCHAR(50),
            'Optimize Table Bloat'::TEXT,
            'Table bloat is high. Implement table maintenance and optimization.'::TEXT,
            'MEDIUM'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'Reduce storage usage and improve query performance'::TEXT,
            '1. Run VACUUM FULL on bloated tables
2. Implement regular VACUUM and ANALYZE
3. Consider table partitioning
4. Optimize UPDATE and DELETE patterns
5. Monitor table bloat regularly'::TEXT,
            'Monitor table bloat percentage and storage usage'::TEXT,
            ARRAY['database', 'tables', 'storage']::TEXT[],
            ARRAY['table_bloat_percentage', 'storage_usage']::TEXT[],
            20.0::NUMERIC,
            4.0::NUMERIC,
            ARRAY['Database maintenance tools', 'Monitoring system']::TEXT[],
            ARRAY['Regular maintenance schedule', 'Storage monitoring']::TEXT[];
    END IF;
    
    -- Return empty result if no recommendations
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 3. Create a function to generate classification performance recommendations
CREATE OR REPLACE FUNCTION generate_classification_performance_recommendations() 
RETURNS TABLE (
    recommendation_id VARCHAR(255),
    recommendation_type VARCHAR(100),
    recommendation_category VARCHAR(50),
    recommendation_title TEXT,
    recommendation_description TEXT,
    recommendation_priority VARCHAR(20),
    recommendation_impact VARCHAR(20),
    recommendation_effort VARCHAR(20),
    recommendation_benefit TEXT,
    recommendation_implementation TEXT,
    recommendation_validation TEXT,
    affected_systems TEXT[],
    related_metrics TEXT[],
    estimated_improvement_percentage NUMERIC,
    estimated_implementation_time_hours NUMERIC,
    prerequisites TEXT[],
    dependencies TEXT[]
) AS $$
DECLARE
    v_accuracy_percentage NUMERIC;
    v_avg_response_time_ms NUMERIC;
    v_error_rate NUMERIC;
    v_avg_confidence NUMERIC;
    v_keyword_coverage NUMERIC;
    v_algorithm_efficiency NUMERIC;
BEGIN
    -- Get classification performance metrics (assuming classification_accuracy_metrics table exists)
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
            COALESCE(AVG(predicted_confidence), 0),
            COALESCE(
                (COUNT(DISTINCT keyword)::NUMERIC / 
                 NULLIF((SELECT COUNT(*) FROM keywords), 0)) * 100, 
                0
            ),
            COALESCE(AVG(processing_time_ms), 0)
        INTO v_accuracy_percentage, v_avg_response_time_ms, v_error_rate, v_avg_confidence, v_keyword_coverage, v_algorithm_efficiency
        FROM classification_accuracy_metrics
        WHERE timestamp >= NOW() - INTERVAL '24 hours';
        
        -- Classification accuracy optimization recommendations
        IF v_accuracy_percentage < 80 THEN
            RETURN QUERY SELECT 
                'CLASS_ACC_OPT_001'::VARCHAR(255),
                'CLASSIFICATION_ACCURACY'::VARCHAR(100),
                'ACCURACY'::VARCHAR(50),
                'Improve Classification Accuracy'::TEXT,
                'Classification accuracy is below optimal levels. Implement accuracy improvement strategies.'::TEXT,
                'HIGH'::VARCHAR(20),
                'HIGH'::VARCHAR(20),
                'HIGH'::VARCHAR(20),
                'Significantly improve classification accuracy and user satisfaction'::TEXT,
                '1. Analyze misclassified cases and identify patterns
2. Improve keyword matching algorithms
3. Add more training data and examples
4. Implement ensemble methods for better accuracy
5. Regular accuracy monitoring and feedback loops'::TEXT,
                'Monitor classification accuracy metrics and user feedback'::TEXT,
                ARRAY['classification', 'algorithms', 'accuracy']::TEXT[],
                ARRAY['accuracy_percentage', 'misclassification_rate']::TEXT[],
                25.0::NUMERIC,
                16.0::NUMERIC,
                ARRAY['Accuracy analysis tools', 'Training data']::TEXT[],
                ARRAY['Algorithm improvements', 'Data quality enhancement']::TEXT[];
        END IF;
        
        -- Response time optimization recommendations
        IF v_avg_response_time_ms > 2000 THEN
            RETURN QUERY SELECT 
                'CLASS_RESP_OPT_001'::VARCHAR(255),
                'CLASSIFICATION_RESPONSE_TIME'::VARCHAR(100),
                'PERFORMANCE'::VARCHAR(50),
                'Optimize Classification Response Time'::TEXT,
                'Classification response time is high. Implement performance optimization strategies.'::TEXT,
                'HIGH'::VARCHAR(20),
                'HIGH'::VARCHAR(20),
                'MEDIUM'::VARCHAR(20),
                'Improve user experience with faster response times'::TEXT,
                '1. Optimize classification algorithms for speed
2. Implement result caching for common queries
3. Use asynchronous processing for batch operations
4. Optimize database queries in classification pipeline
5. Consider algorithm parallelization'::TEXT,
                'Monitor response time metrics and user satisfaction'::TEXT,
                ARRAY['classification', 'performance', 'algorithms']::TEXT[],
                ARRAY['avg_response_time_ms', 'processing_time_ms']::TEXT[],
                40.0::NUMERIC,
                12.0::NUMERIC,
                ARRAY['Performance profiling tools', 'Caching system']::TEXT[],
                ARRAY['Algorithm optimization', 'Caching implementation']::TEXT[];
        END IF;
        
        -- Error rate optimization recommendations
        IF v_error_rate > 5 THEN
            RETURN QUERY SELECT 
                'CLASS_ERR_OPT_001'::VARCHAR(255),
                'CLASSIFICATION_ERROR_RATE'::VARCHAR(100),
                'RELIABILITY'::VARCHAR(50),
                'Reduce Classification Error Rate'::TEXT,
                'Classification error rate is high. Implement error reduction strategies.'::TEXT,
                'CRITICAL'::VARCHAR(20),
                'HIGH'::VARCHAR(20),
                'HIGH'::VARCHAR(20),
                'Improve system reliability and user trust'::TEXT,
                '1. Implement comprehensive error handling
2. Add input validation and sanitization
3. Implement retry mechanisms for transient failures
4. Add circuit breakers for external dependencies
5. Improve logging and monitoring for error analysis'::TEXT,
                'Monitor error rates and system reliability metrics'::TEXT,
                ARRAY['classification', 'reliability', 'error_handling']::TEXT[],
                ARRAY['error_rate_percentage', 'system_reliability']::TEXT[],
                60.0::NUMERIC,
                8.0::NUMERIC,
                ARRAY['Error monitoring tools', 'Logging system']::TEXT[],
                ARRAY['Error handling framework', 'Monitoring system']::TEXT[];
        END IF;
        
        -- Confidence optimization recommendations
        IF v_avg_confidence < 70 THEN
            RETURN QUERY SELECT 
                'CLASS_CONF_OPT_001'::VARCHAR(255),
                'CLASSIFICATION_CONFIDENCE'::VARCHAR(100),
                'CONFIDENCE'::VARCHAR(50),
                'Improve Classification Confidence'::TEXT,
                'Average classification confidence is low. Implement confidence improvement strategies.'::TEXT,
                'MEDIUM'::VARCHAR(20),
                'MEDIUM'::VARCHAR(20),
                'MEDIUM'::VARCHAR(20),
                'Improve classification reliability and user trust'::TEXT,
                '1. Improve keyword matching algorithms
2. Add more comprehensive training data
3. Implement confidence calibration
4. Use ensemble methods for better confidence
5. Add uncertainty quantification'::TEXT,
                'Monitor confidence scores and classification quality'::TEXT,
                ARRAY['classification', 'confidence', 'algorithms']::TEXT[],
                ARRAY['avg_confidence_percentage', 'confidence_distribution']::TEXT[],
                20.0::NUMERIC,
                10.0::NUMERIC,
                ARRAY['Confidence analysis tools', 'Training data']::TEXT[],
                ARRAY['Algorithm improvements', 'Data quality enhancement']::TEXT[];
        END IF;
        
        -- Keyword coverage optimization recommendations
        IF v_keyword_coverage < 80 THEN
            RETURN QUERY SELECT 
                'CLASS_KEY_OPT_001'::VARCHAR(255),
                'KEYWORD_COVERAGE'::VARCHAR(100),
                'COVERAGE'::VARCHAR(50),
                'Improve Keyword Coverage'::TEXT,
                'Keyword coverage is low. Implement keyword expansion strategies.'::TEXT,
                'MEDIUM'::VARCHAR(20),
                'MEDIUM'::VARCHAR(20),
                'LOW'::VARCHAR(20),
                'Improve classification coverage and accuracy'::TEXT,
                '1. Analyze uncovered business types and add keywords
2. Implement keyword suggestion system
3. Add industry-specific keyword sets
4. Use machine learning for keyword discovery
5. Regular keyword coverage analysis'::TEXT,
                'Monitor keyword coverage and classification success rates'::TEXT,
                ARRAY['classification', 'keywords', 'coverage']::TEXT[],
                ARRAY['keyword_coverage_percentage', 'classification_success_rate']::TEXT[],
                15.0::NUMERIC,
                6.0::NUMERIC,
                ARRAY['Keyword analysis tools', 'Industry data']::TEXT[],
                ARRAY['Keyword management system', 'Data sources']::TEXT[];
        END IF;
    END IF;
    
    -- Return empty result if no recommendations
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to generate system resource recommendations
CREATE OR REPLACE FUNCTION generate_system_resource_recommendations() 
RETURNS TABLE (
    recommendation_id VARCHAR(255),
    recommendation_type VARCHAR(100),
    recommendation_category VARCHAR(50),
    recommendation_title TEXT,
    recommendation_description TEXT,
    recommendation_priority VARCHAR(20),
    recommendation_impact VARCHAR(20),
    recommendation_effort VARCHAR(20),
    recommendation_benefit TEXT,
    recommendation_implementation TEXT,
    recommendation_validation TEXT,
    affected_systems TEXT[],
    related_metrics TEXT[],
    estimated_improvement_percentage NUMERIC,
    estimated_implementation_time_hours NUMERIC,
    prerequisites TEXT[],
    dependencies TEXT[]
) AS $$
DECLARE
    v_cpu_usage NUMERIC;
    v_memory_usage NUMERIC;
    v_disk_usage NUMERIC;
    v_network_latency NUMERIC;
BEGIN
    -- Note: These metrics would typically come from system monitoring tools
    -- For now, we'll use placeholder values and check if monitoring data exists
    
    -- CPU usage optimization recommendations
    v_cpu_usage := 0; -- Would be actual CPU usage from monitoring system
    
    IF v_cpu_usage > 80 THEN
        RETURN QUERY SELECT 
            'SYS_CPU_OPT_001'::VARCHAR(255),
            'CPU_OPTIMIZATION'::VARCHAR(100),
            'RESOURCES'::VARCHAR(50),
            'Optimize CPU Usage'::TEXT,
            'CPU usage is high. Implement CPU optimization strategies.'::TEXT,
            'HIGH'::VARCHAR(20),
            'HIGH'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'Improve system performance and reduce resource contention'::TEXT,
            '1. Optimize application algorithms and data structures
2. Implement CPU profiling and optimization
3. Use asynchronous processing for I/O operations
4. Consider horizontal scaling
5. Optimize database queries and operations'::TEXT,
            'Monitor CPU usage and application performance metrics'::TEXT,
            ARRAY['system', 'cpu', 'application']::TEXT[],
            ARRAY['cpu_usage_percentage', 'application_performance']::TEXT[],
            30.0::NUMERIC,
            8.0::NUMERIC,
            ARRAY['CPU profiling tools', 'Performance monitoring']::TEXT[],
            ARRAY['Algorithm optimization', 'Performance monitoring']::TEXT[];
    END IF;
    
    -- Memory usage optimization recommendations
    v_memory_usage := 0; -- Would be actual memory usage from monitoring system
    
    IF v_memory_usage > 85 THEN
        RETURN QUERY SELECT 
            'SYS_MEM_OPT_001'::VARCHAR(255),
            'MEMORY_OPTIMIZATION'::VARCHAR(100),
            'RESOURCES'::VARCHAR(50),
            'Optimize Memory Usage'::TEXT,
            'Memory usage is high. Implement memory optimization strategies.'::TEXT,
            'HIGH'::VARCHAR(20),
            'HIGH'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'Improve system stability and performance'::TEXT,
            '1. Implement memory profiling and leak detection
2. Optimize data structures and algorithms
3. Use memory pooling and caching strategies
4. Implement garbage collection optimization
5. Consider memory-efficient data formats'::TEXT,
            'Monitor memory usage and system stability'::TEXT,
            ARRAY['system', 'memory', 'application']::TEXT[],
            ARRAY['memory_usage_percentage', 'memory_leaks']::TEXT[],
            25.0::NUMERIC,
            10.0::NUMERIC,
            ARRAY['Memory profiling tools', 'System monitoring']::TEXT[],
            ARRAY['Memory optimization', 'Performance monitoring']::TEXT[];
    END IF;
    
    -- Disk usage optimization recommendations
    v_disk_usage := 0; -- Would be actual disk usage from monitoring system
    
    IF v_disk_usage > 90 THEN
        RETURN QUERY SELECT 
            'SYS_DISK_OPT_001'::VARCHAR(255),
            'DISK_OPTIMIZATION'::VARCHAR(100),
            'STORAGE'::VARCHAR(50),
            'Optimize Disk Usage'::TEXT,
            'Disk usage is high. Implement disk optimization strategies.'::TEXT,
            'MEDIUM'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'LOW'::VARCHAR(20),
            'Prevent disk space issues and improve performance'::TEXT,
            '1. Implement log rotation and cleanup
2. Archive old data and files
3. Use disk compression for large files
4. Implement disk usage monitoring
5. Consider disk space expansion'::TEXT,
            'Monitor disk usage and system performance'::TEXT,
            ARRAY['system', 'disk', 'storage']::TEXT[],
            ARRAY['disk_usage_percentage', 'disk_performance']::TEXT[],
            20.0::NUMERIC,
            4.0::NUMERIC,
            ARRAY['Disk monitoring tools', 'Cleanup scripts']::TEXT[],
            ARRAY['Storage management', 'Monitoring system']::TEXT[];
    END IF;
    
    -- Network latency optimization recommendations
    v_network_latency := 0; -- Would be actual network latency from monitoring system
    
    IF v_network_latency > 100 THEN
        RETURN QUERY SELECT 
            'SYS_NET_OPT_001'::VARCHAR(255),
            'NETWORK_OPTIMIZATION'::VARCHAR(100),
            'NETWORK'::VARCHAR(50),
            'Optimize Network Performance'::TEXT,
            'Network latency is high. Implement network optimization strategies.'::TEXT,
            'MEDIUM'::VARCHAR(20),
            'MEDIUM'::VARCHAR(20),
            'LOW'::VARCHAR(20),
            'Improve network performance and user experience'::TEXT,
            '1. Optimize network configuration
2. Implement connection pooling
3. Use CDN for static content
4. Optimize data transfer formats
5. Implement network monitoring'::TEXT,
            'Monitor network latency and performance metrics'::TEXT,
            ARRAY['system', 'network', 'performance']::TEXT[],
            ARRAY['network_latency_ms', 'network_throughput']::TEXT[],
            15.0::NUMERIC,
            6.0::NUMERIC,
            ARRAY['Network monitoring tools', 'CDN service']::TEXT[],
            ARRAY['Network optimization', 'Performance monitoring']::TEXT[];
    END IF;
    
    -- Return empty result if no recommendations
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 5. Create a function to get all performance recommendations
CREATE OR REPLACE FUNCTION get_all_performance_recommendations() 
RETURNS TABLE (
    recommendation_id VARCHAR(255),
    recommendation_type VARCHAR(100),
    recommendation_category VARCHAR(50),
    recommendation_title TEXT,
    recommendation_description TEXT,
    recommendation_priority VARCHAR(20),
    recommendation_impact VARCHAR(20),
    recommendation_effort VARCHAR(20),
    recommendation_benefit TEXT,
    recommendation_implementation TEXT,
    recommendation_validation TEXT,
    affected_systems TEXT[],
    related_metrics TEXT[],
    estimated_improvement_percentage NUMERIC,
    estimated_implementation_time_hours NUMERIC,
    prerequisites TEXT[],
    dependencies TEXT[]
) AS $$
BEGIN
    -- Get database performance recommendations
    RETURN QUERY SELECT * FROM generate_database_performance_recommendations();
    
    -- Get classification performance recommendations
    RETURN QUERY SELECT * FROM generate_classification_performance_recommendations();
    
    -- Get system resource recommendations
    RETURN QUERY SELECT * FROM generate_system_resource_recommendations();
END;
$$ LANGUAGE plpgsql;

-- 6. Create a function to save recommendations to the database
CREATE OR REPLACE FUNCTION save_performance_recommendations() 
RETURNS INTEGER AS $$
DECLARE
    v_saved_count INTEGER := 0;
    rec RECORD;
BEGIN
    -- Clear existing recommendations
    DELETE FROM performance_recommendations WHERE status = 'pending';
    
    -- Insert new recommendations
    FOR rec IN SELECT * FROM get_all_performance_recommendations() LOOP
        INSERT INTO performance_recommendations (
            recommendation_id,
            recommendation_type,
            recommendation_category,
            recommendation_title,
            recommendation_description,
            recommendation_priority,
            recommendation_impact,
            recommendation_effort,
            recommendation_benefit,
            recommendation_implementation,
            recommendation_validation,
            affected_systems,
            related_metrics,
            estimated_improvement_percentage,
            estimated_implementation_time_hours,
            prerequisites,
            dependencies
        ) VALUES (
            rec.recommendation_id,
            rec.recommendation_type,
            rec.recommendation_category,
            rec.recommendation_title,
            rec.recommendation_description,
            rec.recommendation_priority,
            rec.recommendation_impact,
            rec.recommendation_effort,
            rec.recommendation_benefit,
            rec.recommendation_implementation,
            rec.recommendation_validation,
            rec.affected_systems,
            rec.related_metrics,
            rec.estimated_improvement_percentage,
            rec.estimated_implementation_time_hours,
            rec.prerequisites,
            rec.dependencies
        );
        
        v_saved_count := v_saved_count + 1;
    END LOOP;
    
    RETURN v_saved_count;
END;
$$ LANGUAGE plpgsql;

-- 7. Create a function to get recommendations by priority
CREATE OR REPLACE FUNCTION get_recommendations_by_priority(
    p_priority VARCHAR(20)
) RETURNS TABLE (
    recommendation_id VARCHAR(255),
    recommendation_type VARCHAR(100),
    recommendation_category VARCHAR(50),
    recommendation_title TEXT,
    recommendation_description TEXT,
    recommendation_priority VARCHAR(20),
    recommendation_impact VARCHAR(20),
    recommendation_effort VARCHAR(20),
    recommendation_benefit TEXT,
    recommendation_implementation TEXT,
    recommendation_validation TEXT,
    affected_systems TEXT[],
    related_metrics TEXT[],
    estimated_improvement_percentage NUMERIC,
    estimated_implementation_time_hours NUMERIC,
    prerequisites TEXT[],
    dependencies TEXT[],
    status VARCHAR(20),
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pr.recommendation_id,
        pr.recommendation_type,
        pr.recommendation_category,
        pr.recommendation_title,
        pr.recommendation_description,
        pr.recommendation_priority,
        pr.recommendation_impact,
        pr.recommendation_effort,
        pr.recommendation_benefit,
        pr.recommendation_implementation,
        pr.recommendation_validation,
        pr.affected_systems,
        pr.related_metrics,
        pr.estimated_improvement_percentage,
        pr.estimated_implementation_time_hours,
        pr.prerequisites,
        pr.dependencies,
        pr.status,
        pr.created_at
    FROM performance_recommendations pr
    WHERE pr.recommendation_priority = p_priority
    ORDER BY pr.estimated_improvement_percentage DESC, pr.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 8. Create a function to get recommendations by category
CREATE OR REPLACE FUNCTION get_recommendations_by_category(
    p_category VARCHAR(50)
) RETURNS TABLE (
    recommendation_id VARCHAR(255),
    recommendation_type VARCHAR(100),
    recommendation_category VARCHAR(50),
    recommendation_title TEXT,
    recommendation_description TEXT,
    recommendation_priority VARCHAR(20),
    recommendation_impact VARCHAR(20),
    recommendation_effort VARCHAR(20),
    recommendation_benefit TEXT,
    recommendation_implementation TEXT,
    recommendation_validation TEXT,
    affected_systems TEXT[],
    related_metrics TEXT[],
    estimated_improvement_percentage NUMERIC,
    estimated_implementation_time_hours NUMERIC,
    prerequisites TEXT[],
    dependencies TEXT[],
    status VARCHAR(20),
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pr.recommendation_id,
        pr.recommendation_type,
        pr.recommendation_category,
        pr.recommendation_title,
        pr.recommendation_description,
        pr.recommendation_priority,
        pr.recommendation_impact,
        pr.recommendation_effort,
        pr.recommendation_benefit,
        pr.recommendation_implementation,
        pr.recommendation_validation,
        pr.affected_systems,
        pr.related_metrics,
        pr.estimated_improvement_percentage,
        pr.estimated_implementation_time_hours,
        pr.prerequisites,
        pr.dependencies,
        pr.status,
        pr.created_at
    FROM performance_recommendations pr
    WHERE pr.recommendation_category = p_category
    ORDER BY pr.recommendation_priority DESC, pr.estimated_improvement_percentage DESC;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a function to implement a recommendation
CREATE OR REPLACE FUNCTION implement_recommendation(
    p_recommendation_id VARCHAR(255),
    p_implemented_by VARCHAR(255),
    p_implementation_notes TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE performance_recommendations 
    SET 
        status = 'implemented',
        assigned_to = p_implemented_by,
        implemented_at = NOW(),
        implementation_notes = p_implementation_notes,
        updated_at = NOW()
    WHERE recommendation_id = p_recommendation_id AND status = 'pending';
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- 10. Create a function to get recommendation statistics
CREATE OR REPLACE FUNCTION get_recommendation_statistics() 
RETURNS TABLE (
    total_recommendations BIGINT,
    pending_recommendations BIGINT,
    implemented_recommendations BIGINT,
    critical_recommendations BIGINT,
    high_recommendations BIGINT,
    medium_recommendations BIGINT,
    low_recommendations BIGINT,
    recommendations_by_category JSONB,
    recommendations_by_type JSONB,
    avg_implementation_time_hours NUMERIC,
    total_estimated_improvement NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_recommendations,
        COUNT(*) FILTER (WHERE status = 'pending') as pending_recommendations,
        COUNT(*) FILTER (WHERE status = 'implemented') as implemented_recommendations,
        COUNT(*) FILTER (WHERE recommendation_priority = 'CRITICAL') as critical_recommendations,
        COUNT(*) FILTER (WHERE recommendation_priority = 'HIGH') as high_recommendations,
        COUNT(*) FILTER (WHERE recommendation_priority = 'MEDIUM') as medium_recommendations,
        COUNT(*) FILTER (WHERE recommendation_priority = 'LOW') as low_recommendations,
        jsonb_object_agg(pr.recommendation_category, category_count) as recommendations_by_category,
        jsonb_object_agg(pr.recommendation_type, type_count) as recommendations_by_type,
        ROUND(AVG(pr.estimated_implementation_time_hours), 2) as avg_implementation_time_hours,
        ROUND(SUM(pr.estimated_improvement_percentage), 2) as total_estimated_improvement
    FROM (
        SELECT 
            recommendation_category,
            recommendation_type,
            COUNT(*) as category_count,
            COUNT(*) as type_count
        FROM performance_recommendations
        GROUP BY recommendation_category, recommendation_type
    ) pr;
END;
$$ LANGUAGE plpgsql;

-- 11. Create a function to validate recommendations setup
CREATE OR REPLACE FUNCTION validate_recommendations_setup() 
RETURNS TABLE (
    component TEXT,
    status TEXT,
    details TEXT,
    recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check if performance_recommendations table exists
    SELECT 
        'Performance Recommendations Table' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_recommendations') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Table for storing performance recommendations' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_recommendations') 
            THEN 'Table exists and ready for use' 
            ELSE 'Create performance_recommendations table' 
        END as recommendation
    
    UNION ALL
    
    -- Check if recommendation functions exist
    SELECT 
        'Recommendation Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'generate_database_performance_recommendations') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for generating performance recommendations' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'generate_database_performance_recommendations') 
            THEN 'All recommendation functions are available' 
            ELSE 'Create recommendation functions' 
        END as recommendation
    
    UNION ALL
    
    -- Check if monitoring functions exist
    SELECT 
        'Monitoring Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'get_database_stats') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for monitoring performance metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'get_database_stats') 
            THEN 'All monitoring functions are available' 
            ELSE 'Create monitoring functions' 
        END as recommendation;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_performance_recommendations_recommendation_id ON performance_recommendations(recommendation_id);
CREATE INDEX IF NOT EXISTS idx_performance_recommendations_priority ON performance_recommendations(recommendation_priority);
CREATE INDEX IF NOT EXISTS idx_performance_recommendations_category ON performance_recommendations(recommendation_category);
CREATE INDEX IF NOT EXISTS idx_performance_recommendations_type ON performance_recommendations(recommendation_type);
CREATE INDEX IF NOT EXISTS idx_performance_recommendations_status ON performance_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_performance_recommendations_created_at ON performance_recommendations(created_at);

-- Create a view for easy recommendations dashboard access
CREATE OR REPLACE VIEW performance_recommendations_dashboard AS
SELECT 
    'Total Recommendations' as metric_name,
    (SELECT COUNT(*)::TEXT FROM performance_recommendations) as current_value,
    '0' as target_value,
    (SELECT 
        CASE 
            WHEN COUNT(*) = 0 THEN 'OK'
            WHEN COUNT(*) FILTER (WHERE recommendation_priority = 'CRITICAL') > 0 THEN 'CRITICAL'
            WHEN COUNT(*) FILTER (WHERE recommendation_priority = 'HIGH') > 0 THEN 'WARNING'
            ELSE 'FAIR'
        END
    FROM performance_recommendations WHERE status = 'pending') as status
UNION ALL
SELECT 
    'Critical Recommendations' as metric_name,
    (SELECT COUNT(*)::TEXT FROM performance_recommendations WHERE recommendation_priority = 'CRITICAL' AND status = 'pending') as current_value,
    '0' as target_value,
    (SELECT 
        CASE 
            WHEN COUNT(*) = 0 THEN 'OK'
            ELSE 'CRITICAL'
        END
    FROM performance_recommendations WHERE recommendation_priority = 'CRITICAL' AND status = 'pending') as status
UNION ALL
SELECT 
    'High Priority Recommendations' as metric_name,
    (SELECT COUNT(*)::TEXT FROM performance_recommendations WHERE recommendation_priority = 'HIGH' AND status = 'pending') as current_value,
    '0' as target_value,
    (SELECT 
        CASE 
            WHEN COUNT(*) = 0 THEN 'OK'
            ELSE 'WARNING'
        END
    FROM performance_recommendations WHERE recommendation_priority = 'HIGH' AND status = 'pending') as status;

-- Grant permissions
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO authenticated;
GRANT SELECT ON performance_recommendations_dashboard TO authenticated;
GRANT SELECT, INSERT, UPDATE ON performance_recommendations TO authenticated;

-- Initial setup message
DO $$
BEGIN
    RAISE NOTICE 'Performance optimization recommendations system setup completed successfully!';
    RAISE NOTICE 'Total functions created: 11';
    RAISE NOTICE 'Total tables created: 1';
    RAISE NOTICE 'Total views created: 1';
    RAISE NOTICE 'Total indexes created: 6';
    RAISE NOTICE 'All performance optimization recommendation tools are now available.';
    RAISE NOTICE 'Use performance_recommendations_dashboard view to access current recommendations status.';
    RAISE NOTICE 'Call save_performance_recommendations() to generate and save recommendations.';
    RAISE NOTICE 'Call get_all_performance_recommendations() to get current recommendations.';
END $$;
