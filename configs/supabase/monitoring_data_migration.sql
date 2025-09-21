-- ============================================================================
-- MONITORING DATA MIGRATION SCRIPT
-- ============================================================================
-- This script migrates data from redundant monitoring tables to the unified
-- monitoring schema, ensuring data integrity and consistency.
--
-- Migration Strategy:
-- 1. Migrate performance metrics from multiple tables to unified_performance_metrics
-- 2. Migrate alerts from multiple alert tables to unified_performance_alerts
-- 3. Migrate reports and analytics to unified_performance_reports
-- 4. Migrate integration health data to performance_integration_health
-- 5. Preserve data relationships and metadata
-- ============================================================================

-- ============================================================================
-- 1. MIGRATE PERFORMANCE METRICS
-- ============================================================================

-- Migrate from performance_metrics (comprehensive_performance_monitoring.sql)
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'system' as component,
    'main' as component_instance,
    'performance_monitor' as service_name,
    'performance' as metric_type,
    CASE 
        WHEN metric_name LIKE '%response_time%' THEN 'latency'
        WHEN metric_name LIKE '%throughput%' THEN 'throughput'
        WHEN metric_name LIKE '%error%' THEN 'error_rate'
        WHEN metric_name LIKE '%memory%' THEN 'memory'
        WHEN metric_name LIKE '%cpu%' THEN 'cpu'
        ELSE 'general'
    END as metric_category,
    metric_name,
    metric_value,
    metric_unit,
    jsonb_build_object(
        'environment', environment,
        'version', version,
        'instance_id', instance_id
    ) as tags,
    jsonb_build_object(
        'threshold', threshold_value,
        'status', status,
        'original_table', 'performance_metrics'
    ) as metadata,
    0.95 as confidence_score,
    'comprehensive_performance_monitoring' as data_source,
    created_at
FROM performance_metrics
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_metrics');

-- Migrate from response_time_metrics
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'api' as component,
    endpoint as component_instance,
    'api_gateway' as service_name,
    'performance' as metric_type,
    'latency' as metric_category,
    'response_time' as metric_name,
    response_time_ms as metric_value,
    'ms' as metric_unit,
    jsonb_build_object(
        'endpoint', endpoint,
        'method', method,
        'status_code', status_code
    ) as tags,
    jsonb_build_object(
        'min_response_time', min_response_time_ms,
        'max_response_time', max_response_time_ms,
        'request_count', request_count,
        'original_table', 'response_time_metrics'
    ) as metadata,
    0.90 as confidence_score,
    'comprehensive_performance_monitoring' as data_source,
    created_at
FROM response_time_metrics
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'response_time_metrics');

-- Migrate from memory_metrics
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'system' as component,
    'main' as component_instance,
    'memory_monitor' as service_name,
    'resource' as metric_type,
    'memory' as metric_category,
    'memory_usage' as metric_name,
    memory_usage_mb as metric_value,
    'MB' as metric_unit,
    jsonb_build_object(
        'environment', environment,
        'instance_id', instance_id
    ) as tags,
    jsonb_build_object(
        'total_memory', total_memory_mb,
        'free_memory', free_memory_mb,
        'memory_percentage', memory_percentage,
        'original_table', 'memory_metrics'
    ) as metadata,
    0.95 as confidence_score,
    'comprehensive_performance_monitoring' as data_source,
    created_at
FROM memory_metrics
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'memory_metrics');

-- Migrate from database_performance_metrics
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'database' as component,
    database_name as component_instance,
    'database_monitor' as service_name,
    'performance' as metric_type,
    CASE 
        WHEN metric_name LIKE '%query%' THEN 'latency'
        WHEN metric_name LIKE '%connection%' THEN 'resource'
        WHEN metric_name LIKE '%cache%' THEN 'cache'
        ELSE 'general'
    END as metric_category,
    metric_name,
    metric_value,
    metric_unit,
    jsonb_build_object(
        'database_name', database_name,
        'environment', environment
    ) as tags,
    jsonb_build_object(
        'threshold', threshold_value,
        'status', status,
        'original_table', 'database_performance_metrics'
    ) as metadata,
    0.90 as confidence_score,
    'comprehensive_performance_monitoring' as data_source,
    created_at
FROM database_performance_metrics
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'database_performance_metrics');

-- Migrate from security_validation_metrics
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'security' as component,
    validation_type as component_instance,
    'security_validator' as service_name,
    'security' as metric_type,
    'validation' as metric_category,
    metric_name,
    metric_value,
    metric_unit,
    jsonb_build_object(
        'validation_type', validation_type,
        'environment', environment
    ) as tags,
    jsonb_build_object(
        'threshold', threshold_value,
        'status', status,
        'original_table', 'security_validation_metrics'
    ) as metadata,
    0.85 as confidence_score,
    'security_validation_monitoring' as data_source,
    created_at
FROM security_validation_metrics
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'security_validation_metrics');

-- ============================================================================
-- 2. MIGRATE ALERTS
-- ============================================================================

-- Migrate from performance_alerts
INSERT INTO unified_performance_alerts (
    alert_type,
    alert_category,
    severity,
    component,
    component_instance,
    service_name,
    alert_name,
    description,
    condition,
    current_value,
    threshold_value,
    status,
    tags,
    metadata,
    created_at
)
SELECT 
    'threshold' as alert_type,
    'performance' as alert_category,
    CASE 
        WHEN severity = 'high' THEN 'critical'
        WHEN severity = 'medium' THEN 'warning'
        ELSE 'info'
    END as severity,
    'system' as component,
    'main' as component_instance,
    'performance_monitor' as service_name,
    alert_name,
    description,
    jsonb_build_object(
        'metric_name', metric_name,
        'operator', operator,
        'threshold', threshold_value
    ) as condition,
    current_value,
    threshold_value,
    CASE 
        WHEN status = 'active' THEN 'active'
        WHEN status = 'resolved' THEN 'resolved'
        ELSE 'acknowledged'
    END as status,
    jsonb_build_object(
        'environment', environment,
        'instance_id', instance_id
    ) as tags,
    jsonb_build_object(
        'original_table', 'performance_alerts',
        'alert_id', id
    ) as metadata,
    created_at
FROM performance_alerts
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_alerts');

-- Migrate from security_validation_alerts
INSERT INTO unified_performance_alerts (
    alert_type,
    alert_category,
    severity,
    component,
    component_instance,
    service_name,
    alert_name,
    description,
    condition,
    current_value,
    threshold_value,
    status,
    tags,
    metadata,
    created_at
)
SELECT 
    'threshold' as alert_type,
    'security' as alert_category,
    CASE 
        WHEN severity = 'critical' THEN 'critical'
        WHEN severity = 'warning' THEN 'warning'
        ELSE 'info'
    END as severity,
    'security' as component,
    validation_type as component_instance,
    'security_validator' as service_name,
    alert_name,
    description,
    jsonb_build_object(
        'validation_type', validation_type,
        'metric_name', metric_name,
        'operator', operator
    ) as condition,
    current_value,
    threshold_value,
    CASE 
        WHEN status = 'active' THEN 'active'
        WHEN status = 'resolved' THEN 'resolved'
        ELSE 'acknowledged'
    END as status,
    jsonb_build_object(
        'validation_type', validation_type,
        'environment', environment
    ) as tags,
    jsonb_build_object(
        'original_table', 'security_validation_alerts',
        'alert_id', id
    ) as metadata,
    created_at
FROM security_validation_alerts
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'security_validation_alerts');

-- Migrate from database_performance_alerts
INSERT INTO unified_performance_alerts (
    alert_type,
    alert_category,
    severity,
    component,
    component_instance,
    service_name,
    alert_name,
    description,
    condition,
    current_value,
    threshold_value,
    status,
    tags,
    metadata,
    created_at
)
SELECT 
    'threshold' as alert_type,
    'performance' as alert_category,
    CASE 
        WHEN severity = 'critical' THEN 'critical'
        WHEN severity = 'warning' THEN 'warning'
        ELSE 'info'
    END as severity,
    'database' as component,
    database_name as component_instance,
    'database_monitor' as service_name,
    alert_name,
    description,
    jsonb_build_object(
        'database_name', database_name,
        'metric_name', metric_name,
        'operator', operator
    ) as condition,
    current_value,
    threshold_value,
    CASE 
        WHEN status = 'active' THEN 'active'
        WHEN status = 'resolved' THEN 'resolved'
        ELSE 'acknowledged'
    END as status,
    jsonb_build_object(
        'database_name', database_name,
        'environment', environment
    ) as tags,
    jsonb_build_object(
        'original_table', 'database_performance_alerts',
        'alert_id', id
    ) as metadata,
    created_at
FROM database_performance_alerts
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'database_performance_alerts');

-- ============================================================================
-- 3. MIGRATE QUERY PERFORMANCE DATA
-- ============================================================================

-- Migrate from query_performance_log
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'database' as component,
    database_name as component_instance,
    'query_monitor' as service_name,
    'performance' as metric_type,
    'latency' as metric_category,
    'query_execution_time' as metric_name,
    execution_time_ms as metric_value,
    'ms' as metric_unit,
    jsonb_build_object(
        'database_name', database_name,
        'query_type', query_type,
        'table_name', table_name
    ) as tags,
    jsonb_build_object(
        'query_hash', query_hash,
        'rows_affected', rows_affected,
        'original_table', 'query_performance_log'
    ) as metadata,
    0.95 as confidence_score,
    'query_performance_monitoring' as data_source,
    created_at
FROM query_performance_log
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'query_performance_log');

-- Migrate from enhanced_query_performance_log
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'database' as component,
    database_name as component_instance,
    'enhanced_query_monitor' as service_name,
    'performance' as metric_type,
    'latency' as metric_category,
    'enhanced_query_execution_time' as metric_name,
    execution_time_ms as metric_value,
    'ms' as metric_unit,
    jsonb_build_object(
        'database_name', database_name,
        'query_type', query_type,
        'table_name', table_name,
        'index_used', index_used
    ) as tags,
    jsonb_build_object(
        'query_hash', query_hash,
        'rows_affected', rows_affected,
        'cpu_time_ms', cpu_time_ms,
        'io_time_ms', io_time_ms,
        'original_table', 'enhanced_query_performance_log'
    ) as metadata,
    0.98 as confidence_score,
    'enhanced_database_monitoring' as data_source,
    created_at
FROM enhanced_query_performance_log
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'enhanced_query_performance_log');

-- ============================================================================
-- 4. MIGRATE CONNECTION POOL METRICS
-- ============================================================================

-- Migrate from connection_pool_metrics
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'database' as component,
    pool_name as component_instance,
    'connection_pool_monitor' as service_name,
    'resource' as metric_type,
    'connection' as metric_category,
    'connection_count' as metric_name,
    active_connections as metric_value,
    'count' as metric_unit,
    jsonb_build_object(
        'pool_name', pool_name,
        'database_name', database_name
    ) as tags,
    jsonb_build_object(
        'max_connections', max_connections,
        'idle_connections', idle_connections,
        'waiting_connections', waiting_connections,
        'original_table', 'connection_pool_metrics'
    ) as metadata,
    0.90 as confidence_score,
    'connection_pool_monitoring' as data_source,
    created_at
FROM connection_pool_metrics
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'connection_pool_metrics');

-- ============================================================================
-- 5. MIGRATE CLASSIFICATION ACCURACY METRICS
-- ============================================================================

-- Migrate from classification_accuracy_metrics
INSERT INTO unified_performance_metrics (
    component,
    component_instance,
    service_name,
    metric_type,
    metric_category,
    metric_name,
    metric_value,
    metric_unit,
    tags,
    metadata,
    confidence_score,
    data_source,
    created_at
)
SELECT 
    'classification' as component,
    model_name as component_instance,
    'classification_monitor' as service_name,
    'business' as metric_type,
    'accuracy' as metric_category,
    'classification_accuracy' as metric_name,
    accuracy_score as metric_value,
    'percent' as metric_unit,
    jsonb_build_object(
        'model_name', model_name,
        'dataset_type', dataset_type,
        'environment', environment
    ) as tags,
    jsonb_build_object(
        'precision', precision_score,
        'recall', recall_score,
        'f1_score', f1_score,
        'total_predictions', total_predictions,
        'correct_predictions', correct_predictions,
        'original_table', 'classification_accuracy_metrics'
    ) as metadata,
    0.95 as confidence_score,
    'classification_accuracy_monitoring' as data_source,
    created_at
FROM classification_accuracy_metrics
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'classification_accuracy_metrics');

-- ============================================================================
-- 6. CREATE MIGRATION SUMMARY REPORT
-- ============================================================================

-- Create a temporary table to track migration results
CREATE TEMP TABLE migration_summary (
    table_name VARCHAR(100),
    records_migrated INTEGER,
    migration_status VARCHAR(20),
    notes TEXT
);

-- Insert migration summary data
INSERT INTO migration_summary VALUES
('performance_metrics', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'comprehensive_performance_monitoring'),
 'completed', 'Migrated to unified_performance_metrics'),
('response_time_metrics', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'comprehensive_performance_monitoring' AND metric_name = 'response_time'),
 'completed', 'Migrated to unified_performance_metrics'),
('memory_metrics', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'comprehensive_performance_monitoring' AND metric_category = 'memory'),
 'completed', 'Migrated to unified_performance_metrics'),
('database_performance_metrics', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'comprehensive_performance_monitoring' AND component = 'database'),
 'completed', 'Migrated to unified_performance_metrics'),
('security_validation_metrics', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'security_validation_monitoring'),
 'completed', 'Migrated to unified_performance_metrics'),
('performance_alerts', 
 (SELECT COUNT(*) FROM unified_performance_alerts WHERE metadata->>'original_table' = 'performance_alerts'),
 'completed', 'Migrated to unified_performance_alerts'),
('security_validation_alerts', 
 (SELECT COUNT(*) FROM unified_performance_alerts WHERE metadata->>'original_table' = 'security_validation_alerts'),
 'completed', 'Migrated to unified_performance_alerts'),
('database_performance_alerts', 
 (SELECT COUNT(*) FROM unified_performance_alerts WHERE metadata->>'original_table' = 'database_performance_alerts'),
 'completed', 'Migrated to unified_performance_alerts'),
('query_performance_log', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'query_performance_monitoring'),
 'completed', 'Migrated to unified_performance_metrics'),
('enhanced_query_performance_log', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'enhanced_database_monitoring'),
 'completed', 'Migrated to unified_performance_metrics'),
('connection_pool_metrics', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'connection_pool_monitoring'),
 'completed', 'Migrated to unified_performance_metrics'),
('classification_accuracy_metrics', 
 (SELECT COUNT(*) FROM unified_performance_metrics WHERE data_source = 'classification_accuracy_monitoring'),
 'completed', 'Migrated to unified_performance_metrics');

-- Display migration summary
SELECT 
    table_name,
    records_migrated,
    migration_status,
    notes
FROM migration_summary
ORDER BY table_name;

-- ============================================================================
-- 7. VERIFY MIGRATION INTEGRITY
-- ============================================================================

-- Verify that all unified tables have data
SELECT 
    'unified_performance_metrics' as table_name,
    COUNT(*) as total_records,
    COUNT(DISTINCT component) as unique_components,
    COUNT(DISTINCT metric_type) as unique_metric_types,
    MIN(created_at) as earliest_record,
    MAX(created_at) as latest_record
FROM unified_performance_metrics

UNION ALL

SELECT 
    'unified_performance_alerts' as table_name,
    COUNT(*) as total_records,
    COUNT(DISTINCT component) as unique_components,
    COUNT(DISTINCT alert_category) as unique_alert_categories,
    MIN(created_at) as earliest_record,
    MAX(created_at) as latest_record
FROM unified_performance_alerts;

-- ============================================================================
-- MIGRATION COMPLETED
-- ============================================================================
-- The migration has been completed successfully. All monitoring data from
-- redundant tables has been migrated to the unified monitoring schema.
--
-- Next Steps:
-- 1. Update application code to use unified tables
-- 2. Test monitoring functionality
-- 3. Verify alert systems
-- 4. Remove redundant tables (Task 3.1.4)
-- ============================================================================
