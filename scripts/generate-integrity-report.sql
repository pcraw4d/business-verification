-- Comprehensive Data Integrity Report Generation Script
-- This script generates a comprehensive data integrity report with all validation results

-- ============================================================================
-- DATA INTEGRITY REPORT GENERATION
-- ============================================================================

-- Report Header
SELECT 
    'DATA INTEGRITY REPORT' as report_type,
    NOW() as generated_at,
    current_database() as database_name,
    current_user as generated_by,
    version() as postgresql_version;

-- ============================================================================
-- EXECUTIVE SUMMARY
-- ============================================================================

-- Overall Database Health Summary
WITH integrity_summary AS (
    SELECT 
        -- Foreign Key Integrity
        (SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_type = 'FOREIGN KEY' AND table_schema = 'public') as total_foreign_keys,
        (SELECT COUNT(*) FROM (
            SELECT COUNT(*) FROM merchants m LEFT JOIN users u ON m.user_id = u.id WHERE m.user_id IS NOT NULL AND u.id IS NULL
            UNION ALL
            SELECT COUNT(*) FROM business_verifications bv LEFT JOIN merchants m ON bv.merchant_id = m.id WHERE bv.merchant_id IS NOT NULL AND m.id IS NULL
            UNION ALL
            SELECT COUNT(*) FROM classification_results cr LEFT JOIN merchants m ON cr.merchant_id = m.id WHERE cr.merchant_id IS NOT NULL AND m.id IS NULL
            UNION ALL
            SELECT COUNT(*) FROM risk_assessments ra LEFT JOIN merchants m ON ra.merchant_id = m.id WHERE ra.merchant_id IS NOT NULL AND m.id IS NULL
        ) orphaned_counts) as total_orphaned_records,
        
        -- Data Type Issues
        (SELECT COUNT(*) FROM (
            SELECT COUNT(*) FROM users WHERE email IS NOT NULL AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
            UNION ALL
            SELECT COUNT(*) FROM merchants WHERE created_at > updated_at
            UNION ALL
            SELECT COUNT(*) FROM business_verifications WHERE status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed')
            UNION ALL
            SELECT COUNT(*) FROM classification_results WHERE confidence_score < 0 OR confidence_score > 1
            UNION ALL
            SELECT COUNT(*) FROM risk_assessments WHERE risk_level NOT IN ('low', 'medium', 'high', 'critical')
        ) data_type_issues) as total_data_type_issues,
        
        -- Consistency Issues
        (SELECT COUNT(*) FROM (
            SELECT COUNT(*) FROM users u WHERE NOT EXISTS (SELECT 1 FROM merchants m WHERE m.user_id = u.id) AND EXISTS (SELECT 1 FROM merchants LIMIT 1)
            UNION ALL
            SELECT COUNT(*) FROM merchants m WHERE NOT EXISTS (SELECT 1 FROM business_verifications bv WHERE bv.merchant_id = m.id) AND EXISTS (SELECT 1 FROM business_verifications LIMIT 1)
            UNION ALL
            SELECT COUNT(*) FROM merchants m WHERE NOT EXISTS (SELECT 1 FROM classification_results cr WHERE cr.merchant_id = m.id) AND EXISTS (SELECT 1 FROM classification_results LIMIT 1)
        ) consistency_issues) as total_consistency_issues
)
SELECT 
    'EXECUTIVE SUMMARY' as section,
    total_foreign_keys,
    total_orphaned_records,
    total_data_type_issues,
    total_consistency_issues,
    CASE 
        WHEN total_orphaned_records = 0 AND total_data_type_issues = 0 AND total_consistency_issues = 0 THEN 'EXCELLENT'
        WHEN total_orphaned_records < 10 AND total_data_type_issues < 5 AND total_consistency_issues < 5 THEN 'GOOD'
        WHEN total_orphaned_records < 50 AND total_data_type_issues < 20 AND total_consistency_issues < 20 THEN 'FAIR'
        ELSE 'POOR'
    END as overall_health_status
FROM integrity_summary;

-- ============================================================================
-- TABLE INVENTORY AND STATUS
-- ============================================================================

-- Table Existence and Basic Statistics
SELECT 
    'TABLE INVENTORY' as section,
    t.table_name,
    CASE 
        WHEN t.table_name IS NOT NULL THEN 'EXISTS'
        ELSE 'MISSING'
    END as status,
    COALESCE(s.n_tup_ins, 0) as estimated_rows,
    COALESCE(s.n_tup_upd, 0) as estimated_updates,
    COALESCE(s.n_tup_del, 0) as estimated_deletes
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
LEFT JOIN pg_stat_user_tables s 
    ON expected_tables.table_name = s.relname
ORDER BY expected_tables.table_name;

-- ============================================================================
-- FOREIGN KEY INTEGRITY ANALYSIS
-- ============================================================================

-- Foreign Key Relationships Overview
SELECT 
    'FOREIGN KEY RELATIONSHIPS' as section,
    tc.table_name as child_table,
    kcu.column_name as child_column,
    ccu.table_name as parent_table,
    ccu.column_name as parent_column,
    tc.constraint_name,
    'ACTIVE' as constraint_status
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

-- Foreign Key Integrity Test Results
SELECT 
    'FOREIGN KEY INTEGRITY TEST' as section,
    'merchants.user_id -> users.id' as relationship,
    COUNT(*) as total_merchants,
    COUNT(CASE WHEN u.id IS NULL THEN 1 END) as orphaned_merchants,
    ROUND(COUNT(CASE WHEN u.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2) as orphaned_percentage,
    CASE 
        WHEN COUNT(CASE WHEN u.id IS NULL THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_status
FROM merchants m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.user_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'user_id')

UNION ALL

SELECT 
    'FOREIGN KEY INTEGRITY TEST' as section,
    'business_verifications.merchant_id -> merchants.id' as relationship,
    COUNT(*) as total_verifications,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_verifications,
    ROUND(COUNT(CASE WHEN m.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2) as orphaned_percentage,
    CASE 
        WHEN COUNT(CASE WHEN m.id IS NULL THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_status
FROM business_verifications bv
LEFT JOIN merchants m ON bv.merchant_id = m.id
WHERE bv.merchant_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_verifications' AND column_name = 'merchant_id')

UNION ALL

SELECT 
    'FOREIGN KEY INTEGRITY TEST' as section,
    'classification_results.merchant_id -> merchants.id' as relationship,
    COUNT(*) as total_results,
    COUNT(CASE WHEN m.id IS NULL THEN 1 END) as orphaned_results,
    ROUND(COUNT(CASE WHEN m.id IS NULL THEN 1 END) * 100.0 / COUNT(*), 2) as orphaned_percentage,
    CASE 
        WHEN COUNT(CASE WHEN m.id IS NULL THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_status
FROM classification_results cr
LEFT JOIN merchants m ON cr.merchant_id = m.id
WHERE cr.merchant_id IS NOT NULL
AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'classification_results' AND column_name = 'merchant_id')

ORDER BY relationship;

-- ============================================================================
-- DATA TYPE AND FORMAT VALIDATION
-- ============================================================================

-- Data Type Validation Results
SELECT 
    'DATA TYPE VALIDATION' as section,
    'users.email' as column_reference,
    COUNT(*) as total_records,
    COUNT(CASE WHEN email IS NOT NULL AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' THEN 1 END) as invalid_records,
    ROUND(COUNT(CASE WHEN email IS NOT NULL AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' THEN 1 END) * 100.0 / COUNT(*), 2) as invalid_percentage,
    CASE 
        WHEN COUNT(CASE WHEN email IS NOT NULL AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as validation_status
FROM users
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'email')

UNION ALL

SELECT 
    'DATA TYPE VALIDATION' as section,
    'merchants.created_at vs updated_at' as column_reference,
    COUNT(*) as total_records,
    COUNT(CASE WHEN created_at > updated_at THEN 1 END) as invalid_records,
    ROUND(COUNT(CASE WHEN created_at > updated_at THEN 1 END) * 100.0 / COUNT(*), 2) as invalid_percentage,
    CASE 
        WHEN COUNT(CASE WHEN created_at > updated_at THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as validation_status
FROM merchants
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'created_at')

UNION ALL

SELECT 
    'DATA TYPE VALIDATION' as section,
    'business_verifications.status' as column_reference,
    COUNT(*) as total_records,
    COUNT(CASE WHEN status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed') THEN 1 END) as invalid_records,
    ROUND(COUNT(CASE WHEN status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed') THEN 1 END) * 100.0 / COUNT(*), 2) as invalid_percentage,
    CASE 
        WHEN COUNT(CASE WHEN status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed') THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as validation_status
FROM business_verifications
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'business_verifications' AND column_name = 'status')

UNION ALL

SELECT 
    'DATA TYPE VALIDATION' as section,
    'classification_results.confidence_score' as column_reference,
    COUNT(*) as total_records,
    COUNT(CASE WHEN confidence_score < 0 OR confidence_score > 1 THEN 1 END) as invalid_records,
    ROUND(COUNT(CASE WHEN confidence_score < 0 OR confidence_score > 1 THEN 1 END) * 100.0 / COUNT(*), 2) as invalid_percentage,
    CASE 
        WHEN COUNT(CASE WHEN confidence_score < 0 OR confidence_score > 1 THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as validation_status
FROM classification_results
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'classification_results' AND column_name = 'confidence_score')

ORDER BY column_reference;

-- ============================================================================
-- DATA CONSISTENCY ANALYSIS
-- ============================================================================

-- Data Consistency Test Results
SELECT 
    'DATA CONSISTENCY TEST' as section,
    'User-Merchant Consistency' as test_name,
    COUNT(*) as total_users,
    COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM merchants m WHERE m.user_id = u.id) THEN 1 END) as users_without_merchants,
    ROUND(COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM merchants m WHERE m.user_id = u.id) THEN 1 END) * 100.0 / COUNT(*), 2) as inconsistency_percentage,
    CASE 
        WHEN COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM merchants m WHERE m.user_id = u.id) THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_status
FROM users u
WHERE EXISTS (SELECT 1 FROM merchants LIMIT 1)

UNION ALL

SELECT 
    'DATA CONSISTENCY TEST' as section,
    'Merchant-Verification Consistency' as test_name,
    COUNT(*) as total_merchants,
    COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM business_verifications bv WHERE bv.merchant_id = m.id) THEN 1 END) as merchants_without_verifications,
    ROUND(COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM business_verifications bv WHERE bv.merchant_id = m.id) THEN 1 END) * 100.0 / COUNT(*), 2) as inconsistency_percentage,
    CASE 
        WHEN COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM business_verifications bv WHERE bv.merchant_id = m.id) THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_status
FROM merchants m
WHERE EXISTS (SELECT 1 FROM business_verifications LIMIT 1)

UNION ALL

SELECT 
    'DATA CONSISTENCY TEST' as section,
    'Merchant-Classification Consistency' as test_name,
    COUNT(*) as total_merchants,
    COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM classification_results cr WHERE cr.merchant_id = m.id) THEN 1 END) as merchants_without_classifications,
    ROUND(COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM classification_results cr WHERE cr.merchant_id = m.id) THEN 1 END) * 100.0 / COUNT(*), 2) as inconsistency_percentage,
    CASE 
        WHEN COUNT(CASE WHEN NOT EXISTS (SELECT 1 FROM classification_results cr WHERE cr.merchant_id = m.id) THEN 1 END) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_status
FROM merchants m
WHERE EXISTS (SELECT 1 FROM classification_results LIMIT 1)

ORDER BY test_name;

-- ============================================================================
-- DATA QUALITY METRICS
-- ============================================================================

-- Data Quality Overview
SELECT 
    'DATA QUALITY METRICS' as section,
    'NULL Values in Critical Fields' as metric_name,
    (
        SELECT COUNT(*) FROM users WHERE email IS NULL
    ) +
    (
        SELECT COUNT(*) FROM merchants WHERE name IS NULL
    ) +
    (
        SELECT COUNT(*) FROM business_verifications WHERE status IS NULL
    ) as total_null_values,
    'Lower is better' as interpretation

UNION ALL

SELECT 
    'DATA QUALITY METRICS' as section,
    'Duplicate Records' as metric_name,
    (
        SELECT COUNT(*) - COUNT(DISTINCT user_id) FROM merchants WHERE user_id IS NOT NULL
    ) +
    (
        SELECT COUNT(*) - COUNT(DISTINCT email) FROM users WHERE email IS NOT NULL
    ) as total_duplicates,
    'Should be zero' as interpretation

UNION ALL

SELECT 
    'DATA QUALITY METRICS' as section,
    'Data Length Violations' as metric_name,
    (
        SELECT COUNT(*) FROM users WHERE LENGTH(email) > 255
    ) +
    (
        SELECT COUNT(*) FROM merchants WHERE LENGTH(name) > 255
    ) +
    (
        SELECT COUNT(*) FROM merchants WHERE LENGTH(description) > 1000
    ) as total_length_violations,
    'Should be zero' as interpretation;

-- ============================================================================
-- PERFORMANCE AND INDEX ANALYSIS
-- ============================================================================

-- Index Analysis
SELECT 
    'INDEX ANALYSIS' as section,
    t.table_name,
    t.column_name,
    CASE 
        WHEN i.indexname IS NOT NULL THEN 'INDEXED'
        ELSE 'NOT INDEXED'
    END as index_status,
    CASE 
        WHEN i.indexname IS NOT NULL THEN i.indexname
        ELSE 'MISSING INDEX'
    END as index_name
FROM (
    SELECT tc.table_name, kcu.column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
        ON tc.constraint_name = kcu.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_schema = 'public'
) t
LEFT JOIN pg_indexes i
    ON i.tablename = t.table_name
    AND i.indexdef LIKE '%' || t.column_name || '%'
ORDER BY t.table_name, t.column_name;

-- ============================================================================
-- RECOMMENDATIONS AND ACTION ITEMS
-- ============================================================================

-- Generate Recommendations
SELECT 
    'RECOMMENDATIONS' as section,
    'High Priority' as priority,
    'Fix orphaned records in foreign key relationships' as recommendation,
    'Run cleanup queries to remove orphaned records' as action_required
WHERE EXISTS (
    SELECT 1 FROM merchants m 
    LEFT JOIN users u ON m.user_id = u.id 
    WHERE m.user_id IS NOT NULL AND u.id IS NULL
)

UNION ALL

SELECT 
    'RECOMMENDATIONS' as section,
    'High Priority' as priority,
    'Fix invalid data types and formats' as recommendation,
    'Update invalid records to conform to expected formats' as action_required
WHERE EXISTS (
    SELECT 1 FROM users 
    WHERE email IS NOT NULL AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
)

UNION ALL

SELECT 
    'RECOMMENDATIONS' as section,
    'Medium Priority' as priority,
    'Add missing indexes on foreign key columns' as recommendation,
    'Create indexes to improve query performance' as action_required
WHERE EXISTS (
    SELECT 1 FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
        ON tc.constraint_name = kcu.constraint_name
    LEFT JOIN pg_indexes i
        ON i.tablename = tc.table_name
        AND i.indexdef LIKE '%' || kcu.column_name || '%'
    WHERE tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_schema = 'public'
        AND i.indexname IS NULL
)

UNION ALL

SELECT 
    'RECOMMENDATIONS' as section,
    'Low Priority' as priority,
    'Implement automated data integrity monitoring' as recommendation,
    'Set up regular integrity checks and alerts' as action_required
WHERE NOT EXISTS (
    SELECT 1 FROM merchants m 
    LEFT JOIN users u ON m.user_id = u.id 
    WHERE m.user_id IS NOT NULL AND u.id IS NULL
);

-- ============================================================================
-- REPORT COMPLETION SUMMARY
-- ============================================================================

SELECT 
    'REPORT COMPLETION SUMMARY' as section,
    'Data Integrity Report Generation Complete' as status,
    NOW() as completion_time,
    'Review all sections above for data integrity issues and recommendations' as next_steps;
