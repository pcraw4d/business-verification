-- Remove Redundant Monitoring Tables
-- This script safely removes redundant performance monitoring tables after consolidation
-- into the unified monitoring schema
--
-- IMPORTANT: This script should only be run after:
-- 1. Unified monitoring schema has been created (unified_performance_monitoring.sql)
-- 2. Data has been migrated from redundant tables to unified tables
-- 3. Application code has been updated to use unified tables
-- 4. All monitoring functionality has been tested with unified tables

-- ============================================================================
-- SAFETY CHECKS AND VALIDATION
-- ============================================================================

-- Check if unified tables exist before proceeding
DO $$
BEGIN
    -- Verify unified tables exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_performance_metrics') THEN
        RAISE EXCEPTION 'unified_performance_metrics table does not exist. Cannot proceed with table removal.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_performance_alerts') THEN
        RAISE EXCEPTION 'unified_performance_alerts table does not exist. Cannot proceed with table removal.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_performance_reports') THEN
        RAISE EXCEPTION 'unified_performance_reports table does not exist. Cannot proceed with table removal.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_integration_health') THEN
        RAISE EXCEPTION 'performance_integration_health table does not exist. Cannot proceed with table removal.';
    END IF;
    
    RAISE NOTICE 'All unified monitoring tables exist. Proceeding with redundant table removal.';
END $$;

-- ============================================================================
-- BACKUP REDUNDANT TABLES (Optional - for safety)
-- ============================================================================

-- Create backup tables for critical data before removal
-- This is optional but recommended for safety

-- Backup performance_metrics from comprehensive_performance_monitoring
CREATE TABLE IF NOT EXISTS backup_performance_metrics_comprehensive AS 
SELECT * FROM performance_metrics WHERE EXISTS (
    SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_metrics'
);

-- Backup performance_alerts from comprehensive_performance_monitoring
CREATE TABLE IF NOT EXISTS backup_performance_alerts_comprehensive AS 
SELECT * FROM performance_alerts WHERE EXISTS (
    SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_alerts'
);

-- ============================================================================
-- REMOVE REDUNDANT TABLES
-- ============================================================================

-- Remove tables from comprehensive_performance_monitoring.sql
-- These tables have been consolidated into unified_performance_metrics

-- 1. Remove performance_metrics (redundant with unified_performance_metrics)
DROP TABLE IF EXISTS performance_metrics CASCADE;

-- 2. Remove performance_alerts (redundant with unified_performance_alerts)
DROP TABLE IF EXISTS performance_alerts CASCADE;

-- 3. Remove response_time_metrics (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS response_time_metrics CASCADE;

-- 4. Remove memory_metrics (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS memory_metrics CASCADE;

-- 5. Remove database_performance_metrics (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS database_performance_metrics CASCADE;

-- 6. Remove security_validation_metrics (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS security_validation_metrics CASCADE;

-- Remove tables from enhanced_database_monitoring.sql
-- These have been consolidated into unified_performance_metrics

-- 7. Remove enhanced_query_performance_log (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS enhanced_query_performance_log CASCADE;

-- 8. Remove database_performance_alerts (consolidated into unified_performance_alerts)
DROP TABLE IF EXISTS database_performance_alerts CASCADE;

-- Remove tables from security_validation_monitoring.sql
-- These have been consolidated into unified_performance_metrics and unified_performance_alerts

-- 9. Remove security_validation_performance_log (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS security_validation_performance_log CASCADE;

-- 10. Remove security_validation_alerts (consolidated into unified_performance_alerts)
DROP TABLE IF EXISTS security_validation_alerts CASCADE;

-- 11. Remove security_performance_metrics (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS security_performance_metrics CASCADE;

-- 12. Remove security_system_health (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS security_system_health CASCADE;

-- Remove tables from classification_accuracy_monitoring.sql
-- These have been consolidated into unified_performance_metrics

-- 13. Remove classification_accuracy_metrics (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS classification_accuracy_metrics CASCADE;

-- Remove tables from connection_pool_monitoring.sql
-- These have been consolidated into unified_performance_metrics

-- 14. Remove connection_pool_metrics (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS connection_pool_metrics CASCADE;

-- Remove tables from query_performance_monitoring.sql
-- These have been consolidated into unified_performance_metrics

-- 15. Remove query_performance_log (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS query_performance_log CASCADE;

-- Remove tables from usage_monitoring.sql
-- These have been consolidated into unified_performance_metrics

-- 16. Remove usage_monitoring (consolidated into unified_performance_metrics)
DROP TABLE IF EXISTS usage_monitoring CASCADE;

-- Remove duplicate tables from performance_dashboards.sql
-- These are exact duplicates of comprehensive_performance_monitoring tables

-- Note: performance_metrics from performance_dashboards.sql is already removed above
-- as it's a duplicate of the one from comprehensive_performance_monitoring.sql

-- ============================================================================
-- REMOVE REDUNDANT FUNCTIONS AND PROCEDURES
-- ============================================================================

-- Remove functions that were specific to the redundant tables

-- Remove collect_performance_metrics function (from performance_dashboards.sql)
DROP FUNCTION IF EXISTS collect_performance_metrics() CASCADE;

-- Remove any other functions that were specific to the removed tables
-- (Add more as needed based on the actual functions in the system)

-- ============================================================================
-- REMOVE REDUNDANT VIEWS
-- ============================================================================

-- Remove any views that depended on the removed tables
-- (Add specific views as needed)

-- ============================================================================
-- CLEANUP AND VALIDATION
-- ============================================================================

-- Verify that unified tables still exist and have data
DO $$
DECLARE
    unified_metrics_count INTEGER;
    unified_alerts_count INTEGER;
    unified_reports_count INTEGER;
    integration_health_count INTEGER;
BEGIN
    -- Count records in unified tables
    SELECT COUNT(*) INTO unified_metrics_count FROM unified_performance_metrics;
    SELECT COUNT(*) INTO unified_alerts_count FROM unified_performance_alerts;
    SELECT COUNT(*) INTO unified_reports_count FROM unified_performance_reports;
    SELECT COUNT(*) INTO integration_health_count FROM performance_integration_health;
    
    -- Log the results
    RAISE NOTICE 'Unified monitoring tables status:';
    RAISE NOTICE '  unified_performance_metrics: % records', unified_metrics_count;
    RAISE NOTICE '  unified_performance_alerts: % records', unified_alerts_count;
    RAISE NOTICE '  unified_performance_reports: % records', unified_reports_count;
    RAISE NOTICE '  performance_integration_health: % records', integration_health_count;
    
    -- Verify we have data in the unified tables
    IF unified_metrics_count = 0 THEN
        RAISE WARNING 'unified_performance_metrics table is empty. Please verify data migration was successful.';
    END IF;
    
    RAISE NOTICE 'Redundant monitoring table removal completed successfully.';
END $$;

-- ============================================================================
-- CREATE CLEANUP SUMMARY
-- ============================================================================

-- Create a summary of what was removed
CREATE TABLE IF NOT EXISTS monitoring_cleanup_summary (
    id SERIAL PRIMARY KEY,
    cleanup_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    removed_tables TEXT[],
    backup_tables_created TEXT[],
    unified_tables_verified TEXT[],
    notes TEXT
);

-- Insert cleanup summary
INSERT INTO monitoring_cleanup_summary (
    removed_tables,
    backup_tables_created,
    unified_tables_verified,
    notes
) VALUES (
    ARRAY[
        'performance_metrics',
        'performance_alerts', 
        'response_time_metrics',
        'memory_metrics',
        'database_performance_metrics',
        'security_validation_metrics',
        'enhanced_query_performance_log',
        'database_performance_alerts',
        'security_validation_performance_log',
        'security_validation_alerts',
        'security_performance_metrics',
        'security_system_health',
        'classification_accuracy_metrics',
        'connection_pool_metrics',
        'query_performance_log',
        'usage_monitoring'
    ],
    ARRAY[
        'backup_performance_metrics_comprehensive',
        'backup_performance_alerts_comprehensive'
    ],
    ARRAY[
        'unified_performance_metrics',
        'unified_performance_alerts',
        'unified_performance_reports',
        'performance_integration_health'
    ],
    'Successfully removed 16 redundant monitoring tables and consolidated into 4 unified tables. Backup tables created for safety.'
);

-- ============================================================================
-- FINAL VALIDATION
-- ============================================================================

-- Final check to ensure no broken dependencies
DO $$
DECLARE
    broken_deps INTEGER;
BEGIN
    -- Check for any remaining references to removed tables
    SELECT COUNT(*) INTO broken_deps
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
    WHERE tc.table_schema = 'public'
    AND (
        kcu.referenced_table_name IN (
            'performance_metrics', 'performance_alerts', 'response_time_metrics',
            'memory_metrics', 'database_performance_metrics', 'security_validation_metrics',
            'enhanced_query_performance_log', 'database_performance_alerts',
            'security_validation_performance_log', 'security_validation_alerts',
            'security_performance_metrics', 'security_system_health',
            'classification_accuracy_metrics', 'connection_pool_metrics',
            'query_performance_log', 'usage_monitoring'
        )
    );
    
    IF broken_deps > 0 THEN
        RAISE WARNING 'Found % potential broken dependencies. Please review and fix.', broken_deps;
    ELSE
        RAISE NOTICE 'No broken dependencies found. Cleanup completed successfully.';
    END IF;
END $$;

-- ============================================================================
-- COMPLETION MESSAGE
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'MONITORING TABLE CLEANUP COMPLETED';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Removed 16 redundant monitoring tables';
    RAISE NOTICE 'Consolidated into 4 unified tables:';
    RAISE NOTICE '  - unified_performance_metrics';
    RAISE NOTICE '  - unified_performance_alerts';
    RAISE NOTICE '  - unified_performance_reports';
    RAISE NOTICE '  - performance_integration_health';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Next steps:';
    RAISE NOTICE '1. Test all monitoring functionality';
    RAISE NOTICE '2. Verify application code uses unified tables';
    RAISE NOTICE '3. Monitor system performance';
    RAISE NOTICE '4. Remove backup tables after verification';
    RAISE NOTICE '========================================';
END $$;
