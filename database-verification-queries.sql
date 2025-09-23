-- =====================================================
-- Database Verification Queries
-- KYB Platform - Post-Migration Verification
-- =====================================================

-- Step 1: Verify all tables exist and show row counts
SELECT 
    schemaname,
    relname as table_name,
    n_tup_ins as total_inserts,
    n_tup_upd as total_updates,
    n_tup_del as total_deletes
FROM pg_stat_user_tables 
WHERE schemaname = 'public'
ORDER BY relname;

-- Step 2: Quick health check - row counts for each table
SELECT 
    'classifications' as table_name, COUNT(*) as row_count FROM classifications
UNION ALL
SELECT 'merchants', COUNT(*) FROM merchants
UNION ALL
SELECT 'mock_merchants', COUNT(*) FROM mock_merchants
UNION ALL
SELECT 'risk_keywords', COUNT(*) FROM risk_keywords
UNION ALL
SELECT 'industry_code_crosswalks', COUNT(*) FROM industry_code_crosswalks
UNION ALL
SELECT 'business_risk_assessments', COUNT(*) FROM business_risk_assessments
UNION ALL
SELECT 'risk_keyword_relationships', COUNT(*) FROM risk_keyword_relationships
UNION ALL
SELECT 'classification_performance_metrics', COUNT(*) FROM classification_performance_metrics;

-- Step 3a: Test risk keywords functionality
SELECT keyword, risk_category, risk_severity, risk_score_weight 
FROM risk_keywords 
ORDER BY risk_severity DESC, risk_score_weight DESC
LIMIT 10;

-- Step 3b: Test industry code crosswalks
SELECT source_system, source_code, target_system, target_code, confidence_score
FROM industry_code_crosswalks 
LIMIT 10;

-- Step 3c: Test business risk assessments
SELECT business_name, risk_level, risk_score, risk_factors
FROM business_risk_assessments;

-- Step 4: Verify table structure and constraints
SELECT 
    table_name,
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_schema = 'public' 
AND table_name IN (
    'classifications',
    'merchants', 
    'mock_merchants',
    'risk_keywords',
    'industry_code_crosswalks',
    'business_risk_assessments',
    'risk_keyword_relationships',
    'classification_performance_metrics'
)
ORDER BY table_name, ordinal_position;

-- Step 5: Verify indexes exist
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE schemaname = 'public'
AND tablename IN (
    'classifications',
    'merchants', 
    'mock_merchants',
    'risk_keywords',
    'industry_code_crosswalks',
    'business_risk_assessments',
    'risk_keyword_relationships',
    'classification_performance_metrics'
)
ORDER BY tablename, indexname;
