-- Orphaned Records Detection Script
-- This script checks for orphaned records across all relationships in the database

-- ============================================================================
-- ORPHANED RECORDS DETECTION ANALYSIS
-- ============================================================================

-- 1. Get all foreign key relationships
SELECT 
    'FOREIGN KEY RELATIONSHIPS' as test_type,
    tc.table_name as child_table,
    kcu.column_name as child_column,
    ccu.table_name as parent_table,
    ccu.column_name as parent_column,
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
-- FOREIGN KEY ORPHANED RECORDS DETECTION
-- ============================================================================

-- Check for orphaned records in foreign key relationships
-- This query dynamically checks all foreign key constraints

-- Example: Check merchants -> users relationship
SELECT 
    'ORPHANED MERCHANTS' as test_type,
    'merchants.user_id -> users.id' as relationship,
    COUNT(*) as total_merchants,
    COUNT(CASE WHEN u.id IS NULL THEN 1 END) as orphaned_merchants,
    ROUND(
        COUNT(CASE WHEN u.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM merchants m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.user_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'user_id');

-- Example: Check business_verifications -> merchants relationship
SELECT 
    'ORPHANED BUSINESS VERIFICATIONS' as test_type,
    'business_verifications.merchant_id -> merchants.id' as relationship,
    COUNT(*) as total_verifications,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_verifications,
    ROUND(
        COUNT(CASE WHEN m.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM business_verifications bv
LEFT JOIN merchants m ON bv.merchant_id = m.id
WHERE bv.merchant_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_verifications' AND column_name = 'merchant_id');

-- Example: Check classification_results -> merchants relationship
SELECT 
    'ORPHANED CLASSIFICATION RESULTS' as test_type,
    'classification_results.merchant_id -> merchants.id' as relationship,
    COUNT(*) as total_results,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_results,
    ROUND(
        COUNT(CASE WHEN m.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM classification_results cr
LEFT JOIN merchants m ON cr.merchant_id = m.id
WHERE cr.merchant_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'classification_results' AND column_name = 'merchant_id');

-- Example: Check risk_assessments -> merchants relationship
SELECT 
    'ORPHANED RISK ASSESSMENTS' as test_type,
    'risk_assessments.merchant_id -> merchants.id' as relationship,
    COUNT(*) as total_assessments,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_assessments,
    ROUND(
        COUNT(CASE WHEN m.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM risk_assessments ra
LEFT JOIN merchants m ON ra.merchant_id = m.id
WHERE ra.merchant_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'risk_assessments' AND column_name = 'merchant_id');

-- Example: Check audit_logs -> users relationship
SELECT 
    'ORPHANED AUDIT LOGS' as test_type,
    'audit_logs.user_id -> users.id' as relationship,
    COUNT(*) as total_logs,
    COUNT(CASE WHEN u.id IS NULL THEN 1 END) as orphaned_logs,
    ROUND(
        COUNT(CASE WHEN u.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
WHERE al.user_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'audit_logs' AND column_name = 'user_id');

-- ============================================================================
-- LOGICAL RELATIONSHIP ORPHANED RECORDS DETECTION
-- ============================================================================

-- Check logical business relationships that should exist

-- Check business_verifications -> users relationship
SELECT 
    'ORPHANED BUSINESS VERIFICATIONS (USER)' as test_type,
    'business_verifications.user_id -> users.id' as relationship,
    COUNT(*) as total_verifications,
    COUNT(CASE WHEN u.id IS NULL THEN 1 END) as orphaned_verifications,
    ROUND(
        COUNT(CASE WHEN u.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM business_verifications bv
LEFT JOIN users u ON bv.user_id = u.id
WHERE bv.user_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_verifications' AND column_name = 'user_id');

-- Check merchant_audit_logs -> merchants relationship
SELECT 
    'ORPHANED MERCHANT AUDIT LOGS' as test_type,
    'merchant_audit_logs.merchant_id -> merchants.id' as relationship,
    COUNT(*) as total_logs,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_logs,
    ROUND(
        COUNT(CASE WHEN m.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM merchant_audit_logs mal
LEFT JOIN merchants m ON mal.merchant_id = m.id
WHERE mal.merchant_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchant_audit_logs' AND column_name = 'merchant_id');

-- Check industry_keywords -> industries relationship
SELECT 
    'ORPHANED INDUSTRY KEYWORDS' as test_type,
    'industry_keywords.industry_id -> industries.id' as relationship,
    COUNT(*) as total_keywords,
    COUNT(CASE WHEN i.id IS NULL THEN 1 END) as orphaned_keywords,
    ROUND(
        COUNT(CASE WHEN i.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM industry_keywords ik
LEFT JOIN industries i ON ik.industry_id = i.id
WHERE ik.industry_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'industry_keywords' AND column_name = 'industry_id');

-- Check business_risk_assessments -> merchants relationship
SELECT 
    'ORPHANED BUSINESS RISK ASSESSMENTS' as test_type,
    'business_risk_assessments.business_id -> merchants.id' as relationship,
    COUNT(*) as total_assessments,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_assessments,
    ROUND(
        COUNT(CASE WHEN m.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM business_risk_assessments bra
LEFT JOIN merchants m ON bra.business_id = m.id
WHERE bra.business_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_risk_assessments' AND column_name = 'business_id');

-- Check business_risk_assessments -> risk_keywords relationship
SELECT 
    'ORPHANED RISK ASSESSMENTS (KEYWORDS)' as test_type,
    'business_risk_assessments.risk_keyword_id -> risk_keywords.id' as relationship,
    COUNT(*) as total_assessments,
    COUNT(CASE WHEN rk.id IS NULL THEN 1 END) as orphaned_assessments,
    ROUND(
        COUNT(CASE WHEN rk.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as orphaned_percentage
FROM business_risk_assessments bra
LEFT JOIN risk_keywords rk ON bra.risk_keyword_id = rk.id
WHERE bra.risk_keyword_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_risk_assessments' AND column_name = 'risk_keyword_id');

-- ============================================================================
-- COMPREHENSIVE ORPHANED RECORDS CHECK
-- ============================================================================

-- This query provides a comprehensive overview of all potential orphaned records
WITH fk_constraints AS (
    SELECT 
        tc.table_name as child_table,
        kcu.column_name as child_column,
        ccu.table_name as parent_table,
        ccu.column_name as parent_column,
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
    child_table,
    child_column,
    parent_table,
    parent_column,
    constraint_name,
    'Run individual queries for each relationship to check for orphaned records' as note
FROM fk_constraints
ORDER BY child_table, child_column;

-- ============================================================================
-- ORPHANED RECORDS SUMMARY BY TABLE
-- ============================================================================

-- Summary of orphaned records by child table
SELECT 
    'ORPHANED RECORDS SUMMARY' as test_type,
    'merchants' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN u.id IS NULL THEN 1 END) as orphaned_records,
    'user_id references' as relationship_type
FROM merchants m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.user_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'user_id')

UNION ALL

SELECT 
    'ORPHANED RECORDS SUMMARY' as test_type,
    'business_verifications' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_records,
    'merchant_id references' as relationship_type
FROM business_verifications bv
LEFT JOIN merchants m ON bv.merchant_id = m.id
WHERE bv.merchant_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_verifications' AND column_name = 'merchant_id')

UNION ALL

SELECT 
    'ORPHANED RECORDS SUMMARY' as test_type,
    'classification_results' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_records,
    'merchant_id references' as relationship_type
FROM classification_results cr
LEFT JOIN merchants m ON cr.merchant_id = m.id
WHERE cr.merchant_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'classification_results' AND column_name = 'merchant_id')

ORDER BY orphaned_records DESC, table_name;

-- ============================================================================
-- ORPHANED RECORDS IMPACT ANALYSIS
-- ============================================================================

-- Analyze the impact of orphaned records on data integrity
SELECT 
    'ORPHANED RECORDS IMPACT ANALYSIS' as test_type,
    'Total orphaned records across all relationships' as metric,
    (
        SELECT COUNT(*)
        FROM merchants m
        LEFT JOIN users u ON m.user_id = u.id
        WHERE m.user_id IS NOT NULL AND u.id IS NULL
        AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'user_id')
    ) +
    (
        SELECT COUNT(*)
        FROM business_verifications bv
        LEFT JOIN merchants m ON bv.merchant_id = m.id
        WHERE bv.merchant_id IS NOT NULL AND m.id IS NULL
        AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_verifications' AND column_name = 'merchant_id')
    ) +
    (
        SELECT COUNT(*)
        FROM classification_results cr
        LEFT JOIN merchants m ON cr.merchant_id = m.id
        WHERE cr.merchant_id IS NOT NULL AND m.id IS NULL
        AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'classification_results' AND column_name = 'merchant_id')
    ) as total_orphaned_records;

-- ============================================================================
-- ORPHANED RECORDS CLEANUP RECOMMENDATIONS
-- ============================================================================

-- Provide recommendations for cleaning up orphaned records
SELECT 
    'ORPHANED RECORDS CLEANUP RECOMMENDATIONS' as test_type,
    'merchants' as table_name,
    'user_id' as column_name,
    'users' as referenced_table,
    'id' as referenced_column,
    'DELETE FROM merchants WHERE user_id NOT IN (SELECT id FROM users)' as cleanup_query,
    'WARNING: This will delete merchants with invalid user references' as warning
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'user_id')

UNION ALL

SELECT 
    'ORPHANED RECORDS CLEANUP RECOMMENDATIONS' as test_type,
    'business_verifications' as table_name,
    'merchant_id' as column_name,
    'merchants' as referenced_table,
    'id' as referenced_column,
    'DELETE FROM business_verifications WHERE merchant_id NOT IN (SELECT id FROM merchants)' as cleanup_query,
    'WARNING: This will delete verifications with invalid merchant references' as warning
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_verifications' AND column_name = 'merchant_id')

UNION ALL

SELECT 
    'ORPHANED RECORDS CLEANUP RECOMMENDATIONS' as test_type,
    'classification_results' as table_name,
    'merchant_id' as column_name,
    'merchants' as referenced_table,
    'id' as referenced_column,
    'DELETE FROM classification_results WHERE merchant_id NOT IN (SELECT id FROM merchants)' as cleanup_query,
    'WARNING: This will delete classification results with invalid merchant references' as warning
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'classification_results' AND column_name = 'merchant_id')

ORDER BY table_name, column_name;

-- ============================================================================
-- TEST COMPLETION SUMMARY
-- ============================================================================

SELECT 
    'TEST COMPLETION SUMMARY' as test_type,
    'Orphaned Records Detection Complete' as status,
    NOW() as completion_time,
    'Review all results above for any orphaned records that need cleanup' as next_steps;
