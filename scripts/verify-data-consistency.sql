-- Data Consistency Verification Script
-- This script verifies data consistency across related tables in the database

-- ============================================================================
-- DATA CONSISTENCY VERIFICATION ANALYSIS
-- ============================================================================

-- 1. Table Existence and Structure Verification
SELECT 
    'TABLE EXISTENCE CHECK' as test_type,
    table_name,
    CASE 
        WHEN table_name IS NOT NULL THEN 'EXISTS'
        ELSE 'MISSING'
    END as status
FROM (
    SELECT 'users' as table_name
    UNION ALL SELECT 'merchants'
    UNION ALL SELECT 'business_verifications'
    UNION ALL SELECT 'classification_results'
    UNION ALL SELECT 'risk_assessments'
    UNION ALL SELECT 'audit_logs'
    UNION ALL SELECT 'industries'
    UNION ALL SELECT 'industry_keywords'
    UNION ALL SELECT 'risk_keywords'
    UNION ALL SELECT 'business_risk_assessments'
    UNION ALL SELECT 'merchant_audit_logs'
) expected_tables
LEFT JOIN information_schema.tables t 
    ON expected_tables.table_name = t.table_name 
    AND t.table_schema = 'public'
ORDER BY table_name;

-- ============================================================================
-- COUNT CONSISTENCY VERIFICATION
-- ============================================================================

-- 2. User-Merchant Count Consistency
SELECT 
    'COUNT CONSISTENCY - USER-MERCHANT' as test_type,
    'Users without merchants' as description,
    COUNT(*) as inconsistent_count,
    'Users should have at least one merchant if merchants exist' as expected_behavior
FROM users u 
WHERE NOT EXISTS (
    SELECT 1 FROM merchants m WHERE m.user_id = u.id
)
AND EXISTS (SELECT 1 FROM merchants LIMIT 1);

-- 3. Merchant-Verification Count Consistency
SELECT 
    'COUNT CONSISTENCY - MERCHANT-VERIFICATION' as test_type,
    'Merchants without verifications' as description,
    COUNT(*) as inconsistent_count,
    'Merchants should have at least one verification if verifications exist' as expected_behavior
FROM merchants m 
WHERE NOT EXISTS (
    SELECT 1 FROM business_verifications bv WHERE bv.merchant_id = m.id
)
AND EXISTS (SELECT 1 FROM business_verifications LIMIT 1);

-- 4. Merchant-Classification Count Consistency
SELECT 
    'COUNT CONSISTENCY - MERCHANT-CLASSIFICATION' as test_type,
    'Merchants without classification results' as description,
    COUNT(*) as inconsistent_count,
    'Merchants should have classification results if classifications exist' as expected_behavior
FROM merchants m 
WHERE NOT EXISTS (
    SELECT 1 FROM classification_results cr WHERE cr.merchant_id = m.id
)
AND EXISTS (SELECT 1 FROM classification_results LIMIT 1);

-- 5. Merchant-Risk Assessment Count Consistency
SELECT 
    'COUNT CONSISTENCY - MERCHANT-RISK' as test_type,
    'Merchants without risk assessments' as description,
    COUNT(*) as inconsistent_count,
    'Merchants should have risk assessments if assessments exist' as expected_behavior
FROM merchants m 
WHERE NOT EXISTS (
    SELECT 1 FROM risk_assessments ra WHERE ra.merchant_id = m.id
)
AND EXISTS (SELECT 1 FROM risk_assessments LIMIT 1);

-- ============================================================================
-- BUSINESS LOGIC CONSISTENCY VERIFICATION
-- ============================================================================

-- 6. Business Verification Status Consistency
SELECT 
    'BUSINESS LOGIC - VERIFICATION STATUS' as test_type,
    'Invalid verification statuses' as description,
    COUNT(*) as inconsistent_count,
    'Business verifications should have valid status values' as expected_behavior
FROM business_verifications 
WHERE status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed');

-- 7. Classification Confidence Score Consistency
SELECT 
    'BUSINESS LOGIC - CLASSIFICATION CONFIDENCE' as test_type,
    'Invalid confidence scores' as description,
    COUNT(*) as inconsistent_count,
    'Classification results should have confidence scores between 0 and 1' as expected_behavior
FROM classification_results 
WHERE confidence_score < 0 OR confidence_score > 1;

-- 8. Risk Assessment Level Consistency
SELECT 
    'BUSINESS LOGIC - RISK LEVEL' as test_type,
    'Invalid risk levels' as description,
    COUNT(*) as inconsistent_count,
    'Risk assessments should have valid risk levels' as expected_behavior
FROM risk_assessments 
WHERE risk_level NOT IN ('low', 'medium', 'high', 'critical');

-- 9. User Email Format Consistency
SELECT 
    'BUSINESS LOGIC - EMAIL FORMAT' as test_type,
    'Invalid email formats' as description,
    COUNT(*) as inconsistent_count,
    'Users should have valid email formats' as expected_behavior
FROM users 
WHERE email IS NOT NULL 
AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$';

-- 10. Industry Classification Consistency
SELECT 
    'BUSINESS LOGIC - INDUSTRY CLASSIFICATION' as test_type,
    'Invalid industry classifications' as description,
    COUNT(*) as inconsistent_count,
    'Classification results should reference valid industries' as expected_behavior
FROM classification_results cr
LEFT JOIN industries i ON cr.industry_id = i.id
WHERE cr.industry_id IS NOT NULL AND i.id IS NULL;

-- ============================================================================
-- DATA INTEGRITY CONSISTENCY VERIFICATION
-- ============================================================================

-- 11. Date Consistency - Created Before Updated
SELECT 
    'DATA INTEGRITY - DATE CONSISTENCY' as test_type,
    'Created dates after updated dates' as description,
    COUNT(*) as inconsistent_count,
    'Created dates should be before or equal to updated dates' as expected_behavior
FROM merchants 
WHERE created_at > updated_at;

-- 12. Business Verification Date Consistency
SELECT 
    'DATA INTEGRITY - VERIFICATION DATES' as test_type,
    'Missing created dates' as description,
    COUNT(*) as inconsistent_count,
    'Business verifications should have valid created dates' as expected_behavior
FROM business_verifications 
WHERE created_at IS NULL;

-- 13. Classification Timestamp Consistency
SELECT 
    'DATA INTEGRITY - CLASSIFICATION TIMESTAMPS' as test_type,
    'Missing timestamps' as description,
    COUNT(*) as inconsistent_count,
    'Classification results should have valid timestamps' as expected_behavior
FROM classification_results 
WHERE created_at IS NULL OR updated_at IS NULL;

-- 14. Risk Assessment Date Consistency
SELECT 
    'DATA INTEGRITY - RISK ASSESSMENT DATES' as test_type,
    'Missing assessment dates' as description,
    COUNT(*) as inconsistent_count,
    'Risk assessments should have valid assessment dates' as expected_behavior
FROM risk_assessments 
WHERE assessment_date IS NULL;

-- 15. Audit Log Consistency
SELECT 
    'DATA INTEGRITY - AUDIT LOG CONSISTENCY' as test_type,
    'Audit logs without timestamps' as description,
    COUNT(*) as inconsistent_count,
    'Audit logs should have valid timestamps' as expected_behavior
FROM audit_logs 
WHERE created_at IS NULL;

-- ============================================================================
-- REFERENTIAL INTEGRITY CONSISTENCY VERIFICATION
-- ============================================================================

-- 16. Foreign Key Consistency Check
SELECT 
    'REFERENTIAL INTEGRITY - FOREIGN KEYS' as test_type,
    'Orphaned foreign key references' as description,
    (
        SELECT COUNT(*) FROM merchants m
        LEFT JOIN users u ON m.user_id = u.id
        WHERE m.user_id IS NOT NULL AND u.id IS NULL
    ) +
    (
        SELECT COUNT(*) FROM business_verifications bv
        LEFT JOIN merchants m ON bv.merchant_id = m.id
        WHERE bv.merchant_id IS NOT NULL AND m.id IS NULL
    ) +
    (
        SELECT COUNT(*) FROM classification_results cr
        LEFT JOIN merchants m ON cr.merchant_id = m.id
        WHERE cr.merchant_id IS NOT NULL AND m.id IS NULL
    ) +
    (
        SELECT COUNT(*) FROM risk_assessments ra
        LEFT JOIN merchants m ON ra.merchant_id = m.id
        WHERE ra.merchant_id IS NOT NULL AND m.id IS NULL
    ) as inconsistent_count,
    'All foreign key references should point to existing records' as expected_behavior;

-- ============================================================================
-- DATA QUALITY CONSISTENCY VERIFICATION
-- ============================================================================

-- 17. Duplicate Record Consistency
SELECT 
    'DATA QUALITY - DUPLICATE RECORDS' as test_type,
    'Duplicate merchants by user_id' as description,
    COUNT(*) - COUNT(DISTINCT user_id) as inconsistent_count,
    'Each user should have only one merchant record' as expected_behavior
FROM merchants
WHERE user_id IS NOT NULL;

-- 18. NULL Value Consistency
SELECT 
    'DATA QUALITY - NULL VALUES' as test_type,
    'Critical NULL values' as description,
    (
        SELECT COUNT(*) FROM users WHERE email IS NULL
    ) +
    (
        SELECT COUNT(*) FROM merchants WHERE name IS NULL
    ) +
    (
        SELECT COUNT(*) FROM business_verifications WHERE status IS NULL
    ) as inconsistent_count,
    'Critical fields should not be NULL' as expected_behavior;

-- 19. Data Length Consistency
SELECT 
    'DATA QUALITY - DATA LENGTH' as test_type,
    'Oversized string values' as description,
    (
        SELECT COUNT(*) FROM users WHERE LENGTH(email) > 255
    ) +
    (
        SELECT COUNT(*) FROM merchants WHERE LENGTH(name) > 255
    ) +
    (
        SELECT COUNT(*) FROM merchants WHERE LENGTH(description) > 1000
    ) as inconsistent_count,
    'String values should not exceed column length limits' as expected_behavior;

-- ============================================================================
-- BUSINESS RULE CONSISTENCY VERIFICATION
-- ============================================================================

-- 20. Business Rule - Verification Workflow Consistency
SELECT 
    'BUSINESS RULE - VERIFICATION WORKFLOW' as test_type,
    'Invalid verification workflow states' as description,
    COUNT(*) as inconsistent_count,
    'Verifications should follow proper workflow states' as expected_behavior
FROM business_verifications bv
WHERE status = 'completed' 
AND NOT EXISTS (
    SELECT 1 FROM classification_results cr 
    WHERE cr.merchant_id = bv.merchant_id
);

-- 21. Business Rule - Risk Assessment Consistency
SELECT 
    'BUSINESS RULE - RISK ASSESSMENT' as test_type,
    'Missing risk assessments for high-risk merchants' as description,
    COUNT(*) as inconsistent_count,
    'High-risk merchants should have risk assessments' as expected_behavior
FROM merchants m
WHERE m.risk_level = 'high'
AND NOT EXISTS (
    SELECT 1 FROM risk_assessments ra 
    WHERE ra.merchant_id = m.id
);

-- 22. Business Rule - Classification Consistency
SELECT 
    'BUSINESS RULE - CLASSIFICATION' as test_type,
    'Merchants without primary classification' as description,
    COUNT(*) as inconsistent_count,
    'All merchants should have a primary classification' as expected_behavior
FROM merchants m
WHERE NOT EXISTS (
    SELECT 1 FROM classification_results cr 
    WHERE cr.merchant_id = m.id 
    AND cr.is_primary = true
);

-- ============================================================================
-- PERFORMANCE CONSISTENCY VERIFICATION
-- ============================================================================

-- 23. Index Consistency Check
SELECT 
    'PERFORMANCE - INDEX CONSISTENCY' as test_type,
    'Missing indexes on foreign keys' as description,
    COUNT(*) as inconsistent_count,
    'Foreign key columns should be indexed for performance' as expected_behavior
FROM (
    SELECT tc.table_name, kcu.column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
        ON tc.constraint_name = kcu.constraint_name
    LEFT JOIN pg_indexes i
        ON i.tablename = tc.table_name
        AND i.indexdef LIKE '%' || kcu.column_name || '%'
    WHERE tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_schema = 'public'
        AND i.indexname IS NULL
) missing_indexes;

-- ============================================================================
-- COMPREHENSIVE CONSISTENCY SUMMARY
-- ============================================================================

-- 24. Overall Consistency Score
SELECT 
    'COMPREHENSIVE CONSISTENCY SUMMARY' as test_type,
    'Total consistency issues found' as metric,
    (
        -- Count all inconsistencies from above tests
        (SELECT COUNT(*) FROM users u WHERE NOT EXISTS (SELECT 1 FROM merchants m WHERE m.user_id = u.id) AND EXISTS (SELECT 1 FROM merchants LIMIT 1)) +
        (SELECT COUNT(*) FROM merchants m WHERE NOT EXISTS (SELECT 1 FROM business_verifications bv WHERE bv.merchant_id = m.id) AND EXISTS (SELECT 1 FROM business_verifications LIMIT 1)) +
        (SELECT COUNT(*) FROM business_verifications WHERE status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed')) +
        (SELECT COUNT(*) FROM classification_results WHERE confidence_score < 0 OR confidence_score > 1) +
        (SELECT COUNT(*) FROM risk_assessments WHERE risk_level NOT IN ('low', 'medium', 'high', 'critical')) +
        (SELECT COUNT(*) FROM users WHERE email IS NOT NULL AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$') +
        (SELECT COUNT(*) FROM merchants WHERE created_at > updated_at) +
        (SELECT COUNT(*) FROM business_verifications WHERE created_at IS NULL) +
        (SELECT COUNT(*) FROM classification_results WHERE created_at IS NULL OR updated_at IS NULL) +
        (SELECT COUNT(*) FROM risk_assessments WHERE assessment_date IS NULL)
    ) as total_issues,
    'Lower numbers indicate better data consistency' as interpretation;

-- ============================================================================
-- TEST COMPLETION SUMMARY
-- ============================================================================

SELECT 
    'TEST COMPLETION SUMMARY' as test_type,
    'Data Consistency Verification Complete' as status,
    NOW() as completion_time,
    'Review all results above for any data consistency issues' as next_steps;
