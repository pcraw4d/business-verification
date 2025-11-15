-- Supabase Table Verification Queries
-- Run these queries in Supabase SQL Editor to verify table existence and structure

-- ============================================================================
-- STEP 1: Check if Required Tables Exist
-- ============================================================================

SELECT 
    table_name,
    table_schema,
    CASE 
        WHEN table_name IN (
            'merchant_analytics',
            'merchants',
            'risk_assessments',
            'risk_indicators',
            'enrichment_jobs',
            'enrichment_sources'
        ) THEN '✅ Required for Tests'
        ELSE '⚠️  Additional Table'
    END as test_status
FROM information_schema.tables
WHERE table_schema = 'public'
    AND table_name IN (
        'merchant_analytics',
        'merchants',
        'risk_assessments',
        'risk_indicators',
        'enrichment_jobs',
        'enrichment_sources'
    )
ORDER BY table_name;

-- ============================================================================
-- STEP 2: Check merchant_analytics Table Structure
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

-- Expected columns:
-- - id (UUID)
-- - merchant_id (UUID or VARCHAR)
-- - classification_data (JSONB)
-- - security_data (JSONB)
-- - quality_data (JSONB)
-- - intelligence_data (JSONB)
-- - created_at (TIMESTAMP)
-- - updated_at (TIMESTAMP)

-- ============================================================================
-- STEP 3: Check risk_assessments Table Structure
-- ============================================================================

SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'risk_assessments'
ORDER BY ordinal_position;

-- Expected columns (from migration 010):
-- - id (UUID)
-- - merchant_id (VARCHAR)
-- - status (VARCHAR) - pending, processing, completed, failed
-- - options (JSONB)
-- - result (JSONB)
-- - progress (INTEGER) - 0-100
-- - overall_score (DECIMAL)
-- - risk_level (VARCHAR)
-- - created_at (TIMESTAMP)
-- - updated_at (TIMESTAMP)

-- ============================================================================
-- STEP 4: Check risk_indicators Table Structure
-- ============================================================================

SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'risk_indicators'
ORDER BY ordinal_position;

-- Expected columns:
-- - id (UUID)
-- - merchant_id (VARCHAR)
-- - type (VARCHAR)
-- - name (VARCHAR)
-- - severity (VARCHAR) - low, medium, high, critical
-- - status (VARCHAR) - active, resolved, dismissed
-- - description (TEXT)
-- - detected_at (TIMESTAMP)
-- - score (DECIMAL)

-- ============================================================================
-- STEP 5: Check enrichment Tables
-- ============================================================================

-- Check enrichment_jobs
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'enrichment_jobs'
ORDER BY ordinal_position;

-- Check enrichment_sources
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'enrichment_sources'
ORDER BY ordinal_position;

-- ============================================================================
-- STEP 6: Check Indexes on Test Tables
-- ============================================================================

SELECT 
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
    AND tablename IN (
        'merchant_analytics',
        'merchants',
        'risk_assessments',
        'risk_indicators'
    )
ORDER BY tablename, indexname;

-- ============================================================================
-- STEP 7: Count Records (if tables exist)
-- ============================================================================

-- Count records in each table
SELECT 
    'merchant_analytics' as table_name,
    COUNT(*) as record_count
FROM merchant_analytics
UNION ALL
SELECT 
    'merchants' as table_name,
    COUNT(*) as record_count
FROM merchants
UNION ALL
SELECT 
    'risk_assessments' as table_name,
    COUNT(*) as record_count
FROM risk_assessments
UNION ALL
SELECT 
    'risk_indicators' as table_name,
    COUNT(*) as record_count
FROM risk_indicators;

-- ============================================================================
-- STEP 8: Check Foreign Key Constraints
-- ============================================================================

SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
    AND tc.table_name IN (
        'merchant_analytics',
        'merchants',
        'risk_assessments',
        'risk_indicators'
    )
ORDER BY tc.table_name, kcu.column_name;

