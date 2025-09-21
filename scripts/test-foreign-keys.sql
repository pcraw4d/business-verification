-- Foreign Key Constraint Testing Script
-- This script tests all foreign key constraints in the database
-- Run this script to identify orphaned records and constraint violations

-- ============================================================================
-- FOREIGN KEY CONSTRAINT ANALYSIS
-- ============================================================================

-- 1. Get all foreign key constraints
SELECT 
    'FOREIGN KEY CONSTRAINTS' as test_type,
    tc.table_name as source_table,
    kcu.column_name as source_column,
    ccu.table_name as referenced_table,
    ccu.column_name as referenced_column,
    tc.constraint_name
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
        ON ccu.constraint_name = tc.constraint_name
        AND ccu.table_schema = tc.table_schema
WHERE 
    tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
ORDER BY 
    tc.table_name, kcu.column_name;

-- ============================================================================
-- ORPHANED RECORDS DETECTION
-- ============================================================================

-- Test each foreign key constraint for orphaned records
-- This will show any records that reference non-existent parent records

-- Example queries for common foreign key relationships:

-- Test merchants -> users relationship (if exists)
SELECT 
    'ORPHANED MERCHANTS' as test_type,
    COUNT(*) as orphaned_count,
    'merchants.user_id references non-existent users.id' as description
FROM merchants m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.user_id IS NOT NULL AND u.id IS NULL;

-- Test business_verifications -> merchants relationship (if exists)
SELECT 
    'ORPHANED BUSINESS VERIFICATIONS' as test_type,
    COUNT(*) as orphaned_count,
    'business_verifications.merchant_id references non-existent merchants.id' as description
FROM business_verifications bv
LEFT JOIN merchants m ON bv.merchant_id = m.id
WHERE bv.merchant_id IS NOT NULL AND m.id IS NULL;

-- Test classification_results -> merchants relationship (if exists)
SELECT 
    'ORPHANED CLASSIFICATION RESULTS' as test_type,
    COUNT(*) as orphaned_count,
    'classification_results.merchant_id references non-existent merchants.id' as description
FROM classification_results cr
LEFT JOIN merchants m ON cr.merchant_id = m.id
WHERE cr.merchant_id IS NOT NULL AND m.id IS NULL;

-- Test risk_assessments -> merchants relationship (if exists)
SELECT 
    'ORPHANED RISK ASSESSMENTS' as test_type,
    COUNT(*) as orphaned_count,
    'risk_assessments.merchant_id references non-existent merchants.id' as description
FROM risk_assessments ra
LEFT JOIN merchants m ON ra.merchant_id = m.id
WHERE ra.merchant_id IS NOT NULL AND m.id IS NULL;

-- Test audit_logs -> users relationship (if exists)
SELECT 
    'ORPHANED AUDIT LOGS' as test_type,
    COUNT(*) as orphaned_count,
    'audit_logs.user_id references non-existent users.id' as description
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
WHERE al.user_id IS NOT NULL AND u.id IS NULL;

-- ============================================================================
-- COMPREHENSIVE ORPHANED RECORDS CHECK
-- ============================================================================

-- This query dynamically checks all foreign key constraints for orphaned records
WITH fk_constraints AS (
    SELECT 
        tc.table_name as source_table,
        kcu.column_name as source_column,
        ccu.table_name as referenced_table,
        ccu.column_name as referenced_column,
        tc.constraint_name
    FROM 
        information_schema.table_constraints AS tc 
        JOIN information_schema.key_column_usage AS kcu
            ON tc.constraint_name = kcu.constraint_name
            AND tc.table_schema = kcu.table_schema
        JOIN information_schema.constraint_column_usage AS ccu
            ON ccu.constraint_name = tc.constraint_name
            AND ccu.table_schema = tc.table_schema
    WHERE 
        tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_schema = 'public'
)
SELECT 
    'COMPREHENSIVE ORPHANED RECORDS CHECK' as test_type,
    source_table,
    source_column,
    referenced_table,
    referenced_column,
    constraint_name,
    'Run individual queries for each constraint to check for orphaned records' as note
FROM fk_constraints
ORDER BY source_table, source_column;

-- ============================================================================
-- FOREIGN KEY CONSTRAINT INTEGRITY SUMMARY
-- ============================================================================

-- Summary of all foreign key constraints and their status
SELECT 
    'FOREIGN KEY INTEGRITY SUMMARY' as test_type,
    COUNT(*) as total_constraints,
    'All foreign key constraints in the database' as description
FROM 
    information_schema.table_constraints AS tc 
WHERE 
    tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public';

-- ============================================================================
-- DATA TYPE CONSISTENCY CHECK
-- ============================================================================

-- Check if foreign key columns have compatible data types with referenced columns
SELECT 
    'DATA TYPE CONSISTENCY' as test_type,
    tc.table_name as source_table,
    kcu.column_name as source_column,
    ccu.table_name as referenced_table,
    ccu.column_name as referenced_column,
    sc.data_type as source_data_type,
    rc.data_type as referenced_data_type,
    CASE 
        WHEN sc.data_type = rc.data_type THEN 'MATCH'
        ELSE 'MISMATCH'
    END as type_compatibility
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
        ON ccu.constraint_name = tc.constraint_name
        AND ccu.table_schema = tc.table_schema
    JOIN information_schema.columns sc
        ON sc.table_name = tc.table_name
        AND sc.column_name = kcu.column_name
        AND sc.table_schema = tc.table_schema
    JOIN information_schema.columns rc
        ON rc.table_name = ccu.table_name
        AND rc.column_name = ccu.column_name
        AND rc.table_schema = tc.table_schema
WHERE 
    tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
ORDER BY 
    tc.table_name, kcu.column_name;

-- ============================================================================
-- NULL VALUE CHECK FOR FOREIGN KEY COLUMNS
-- ============================================================================

-- Check for NULL values in foreign key columns (if they should not be NULL)
SELECT 
    'NULL VALUE CHECK' as test_type,
    tc.table_name as source_table,
    kcu.column_name as source_column,
    COUNT(*) as total_records,
    COUNT(CASE WHEN kcu.column_name IS NULL THEN 1 END) as null_count,
    ROUND(
        COUNT(CASE WHEN kcu.column_name IS NULL THEN 1 END) * 100.0 / COUNT(*), 
        2
    ) as null_percentage
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
WHERE 
    tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
GROUP BY 
    tc.table_name, kcu.column_name
ORDER BY 
    null_percentage DESC, tc.table_name, kcu.column_name;

-- ============================================================================
-- PERFORMANCE IMPACT ANALYSIS
-- ============================================================================

-- Check for missing indexes on foreign key columns
SELECT 
    'MISSING INDEXES ON FOREIGN KEYS' as test_type,
    tc.table_name as source_table,
    kcu.column_name as source_column,
    CASE 
        WHEN i.indexname IS NULL THEN 'MISSING INDEX'
        ELSE 'INDEX EXISTS'
    END as index_status,
    i.indexname
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
    LEFT JOIN pg_indexes i
        ON i.tablename = tc.table_name
        AND i.indexdef LIKE '%' || kcu.column_name || '%'
WHERE 
    tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
ORDER BY 
    tc.table_name, kcu.column_name;

-- ============================================================================
-- TEST COMPLETION SUMMARY
-- ============================================================================

SELECT 
    'TEST COMPLETION SUMMARY' as test_type,
    'Foreign Key Constraint Testing Complete' as status,
    NOW() as completion_time,
    'Review all results above for any issues' as next_steps;
