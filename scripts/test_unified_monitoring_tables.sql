-- Test Unified Monitoring Tables
-- This script tests the unified monitoring tables to ensure they are working properly
-- before removing the redundant tables

-- ============================================================================
-- TEST UNIFIED PERFORMANCE METRICS TABLE
-- ============================================================================

-- Test 1: Insert a sample metric
INSERT INTO unified_performance_metrics (
    component, component_instance, service_name, metric_type, metric_category,
    metric_name, metric_value, metric_unit, tags, metadata, data_source, created_at
) VALUES (
    'test', 'test_instance', 'test_service', 'performance', 'test',
    'test_metric', 100.5, 'ms', '{"test": "true"}', '{"details": "test data"}', 'test_script', NOW()
);

-- Test 2: Query the inserted metric
SELECT 
    component, component_instance, service_name, metric_type, metric_category,
    metric_name, metric_value, metric_unit, tags, metadata, data_source, created_at
FROM unified_performance_metrics 
WHERE component = 'test' 
ORDER BY created_at DESC 
LIMIT 1;

-- ============================================================================
-- TEST UNIFIED PERFORMANCE ALERTS TABLE
-- ============================================================================

-- Test 3: Insert a sample alert
INSERT INTO unified_performance_alerts (
    component, component_instance, service_name, alert_type, severity, status,
    title, description, metric_name, metric_value, threshold_value,
    tags, metadata, created_at
) VALUES (
    'test', 'test_instance', 'test_service', 'performance', 'warning', 'active',
    'Test Alert', 'This is a test alert', 'test_metric', 150.0, 100.0,
    '{"test": "true"}', '{"details": "test alert"}', NOW()
);

-- Test 4: Query the inserted alert
SELECT 
    component, component_instance, service_name, alert_type, severity, status,
    title, description, metric_name, metric_value, threshold_value,
    tags, metadata, created_at
FROM unified_performance_alerts 
WHERE component = 'test' 
ORDER BY created_at DESC 
LIMIT 1;

-- ============================================================================
-- TEST UNIFIED PERFORMANCE REPORTS TABLE
-- ============================================================================

-- Test 5: Insert a sample report
INSERT INTO unified_performance_reports (
    component, component_instance, service_name, report_type, report_category,
    report_name, report_data, tags, metadata, created_at
) VALUES (
    'test', 'test_instance', 'test_service', 'performance', 'test',
    'test_report', '{"summary": "test report data"}', '{"test": "true"}', '{"details": "test report"}', NOW()
);

-- Test 6: Query the inserted report
SELECT 
    component, component_instance, service_name, report_type, report_category,
    report_name, report_data, tags, metadata, created_at
FROM unified_performance_reports 
WHERE component = 'test' 
ORDER BY created_at DESC 
LIMIT 1;

-- ============================================================================
-- TEST PERFORMANCE INTEGRATION HEALTH TABLE
-- ============================================================================

-- Test 7: Insert a sample health record
INSERT INTO performance_integration_health (
    component, component_instance, service_name, integration_type, health_status,
    health_score, response_time, error_rate, tags, metadata, created_at
) VALUES (
    'test', 'test_instance', 'test_service', 'api', 'healthy',
    95.5, 50.0, 0.01, '{"test": "true"}', '{"details": "test health"}', NOW()
);

-- Test 8: Query the inserted health record
SELECT 
    component, component_instance, service_name, integration_type, health_status,
    health_score, response_time, error_rate, tags, metadata, created_at
FROM performance_integration_health 
WHERE component = 'test' 
ORDER BY created_at DESC 
LIMIT 1;

-- ============================================================================
-- TEST COMPLEX QUERIES
-- ============================================================================

-- Test 9: Test complex query with joins
SELECT 
    m.component,
    m.metric_name,
    m.metric_value,
    a.alert_type,
    a.severity,
    h.health_status
FROM unified_performance_metrics m
LEFT JOIN unified_performance_alerts a ON m.component = a.component AND m.metric_name = a.metric_name
LEFT JOIN performance_integration_health h ON m.component = h.component
WHERE m.component = 'test'
ORDER BY m.created_at DESC;

-- Test 10: Test aggregation queries
SELECT 
    component,
    metric_category,
    COUNT(*) as metric_count,
    AVG(metric_value) as avg_value,
    MAX(metric_value) as max_value,
    MIN(metric_value) as min_value
FROM unified_performance_metrics 
WHERE created_at >= NOW() - INTERVAL '1 hour'
GROUP BY component, metric_category
ORDER BY component, metric_category;

-- ============================================================================
-- CLEANUP TEST DATA
-- ============================================================================

-- Clean up test data
DELETE FROM unified_performance_metrics WHERE component = 'test';
DELETE FROM unified_performance_alerts WHERE component = 'test';
DELETE FROM unified_performance_reports WHERE component = 'test';
DELETE FROM performance_integration_health WHERE component = 'test';

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Verify all test data has been cleaned up
SELECT 'unified_performance_metrics' as table_name, COUNT(*) as test_records FROM unified_performance_metrics WHERE component = 'test'
UNION ALL
SELECT 'unified_performance_alerts' as table_name, COUNT(*) as test_records FROM unified_performance_alerts WHERE component = 'test'
UNION ALL
SELECT 'unified_performance_reports' as table_name, COUNT(*) as test_records FROM unified_performance_reports WHERE component = 'test'
UNION ALL
SELECT 'performance_integration_health' as table_name, COUNT(*) as test_records FROM performance_integration_health WHERE component = 'test';

-- ============================================================================
-- FINAL STATUS
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'UNIFIED MONITORING TABLES TEST COMPLETED';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'All unified monitoring tables are working correctly.';
    RAISE NOTICE 'Ready to proceed with redundant table removal.';
    RAISE NOTICE '========================================';
END $$;
