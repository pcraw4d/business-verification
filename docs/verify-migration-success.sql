-- Quick Verification: Check if Migration 011 Tables Were Created
-- Run this after executing migration 011_create_test_tables.sql

-- ============================================================================
-- Step 1: Verify All 6 Required Tables Exist
-- ============================================================================

SELECT 
    table_name,
    CASE 
        WHEN table_name IN ('merchants', 'risk_assessments', 'merchant_analytics', 
                           'risk_indicators', 'enrichment_jobs', 'enrichment_sources') 
        THEN '✅ Required'
        ELSE '⚠️ Optional'
    END as status
FROM information_schema.tables
WHERE table_schema = 'public'
    AND table_name IN (
        'merchants',
        'risk_assessments',
        'merchant_analytics',
        'risk_indicators',
        'enrichment_jobs',
        'enrichment_sources'
    )
ORDER BY 
    CASE table_name
        WHEN 'merchants' THEN 1
        WHEN 'risk_assessments' THEN 2
        WHEN 'merchant_analytics' THEN 3
        WHEN 'risk_indicators' THEN 4
        WHEN 'enrichment_jobs' THEN 5
        WHEN 'enrichment_sources' THEN 6
    END;

-- ============================================================================
-- Step 2: Verify merchant_analytics Table Structure
-- ============================================================================

SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'merchant_analytics'
ORDER BY ordinal_position;

-- ============================================================================
-- Step 3: Verify risk_indicators Table Structure
-- ============================================================================

SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'risk_indicators'
ORDER BY ordinal_position;

-- ============================================================================
-- Step 4: Verify enrichment_jobs Table Structure
-- ============================================================================

SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'enrichment_jobs'
ORDER BY ordinal_position;

-- ============================================================================
-- Step 5: Verify enrichment_sources Table Structure and Default Data
-- ============================================================================

-- Check structure
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'enrichment_sources'
ORDER BY ordinal_position;

-- Check default data was inserted
SELECT 
    source_id,
    name,
    description,
    enabled
FROM enrichment_sources
ORDER BY source_id;

-- ============================================================================
-- Step 6: Verify Indexes Were Created
-- ============================================================================

SELECT 
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
    AND tablename IN (
        'merchant_analytics',
        'risk_indicators',
        'enrichment_jobs',
        'enrichment_sources'
    )
ORDER BY tablename, indexname;

