-- =====================================================
-- Risk Keywords System Schema Rollback Migration
-- Supabase Implementation - Task 1.4.1 Rollback
-- =====================================================

-- =====================================================
-- 1. Drop Triggers First
-- =====================================================

DROP TRIGGER IF EXISTS validate_risk_keyword_trigger ON risk_keywords;
DROP TRIGGER IF EXISTS update_risk_keywords_updated_at ON risk_keywords;
DROP TRIGGER IF EXISTS update_industry_code_crosswalks_updated_at ON industry_code_crosswalks;
DROP TRIGGER IF EXISTS update_business_risk_assessments_updated_at ON business_risk_assessments;
DROP TRIGGER IF EXISTS update_risk_keyword_relationships_updated_at ON risk_keyword_relationships;

-- =====================================================
-- 2. Drop Functions
-- =====================================================

DROP FUNCTION IF EXISTS validate_risk_keyword();

-- =====================================================
-- 3. Drop Tables (in reverse dependency order)
-- =====================================================

-- Drop tables that reference other tables first
DROP TABLE IF EXISTS risk_keyword_relationships CASCADE;
DROP TABLE IF EXISTS business_risk_assessments CASCADE;
DROP TABLE IF EXISTS industry_code_crosswalks CASCADE;
DROP TABLE IF EXISTS risk_keywords CASCADE;

-- =====================================================
-- 4. Clean Up Audit Logs
-- =====================================================

-- Remove risk-related audit log entries
DELETE FROM audit_logs 
WHERE table_name IN (
    'risk_keywords', 
    'industry_code_crosswalks', 
    'business_risk_assessments', 
    'risk_keyword_relationships'
);

-- =====================================================
-- 5. Verification
-- =====================================================

-- Verify tables are dropped
SELECT 
    table_name, 
    CASE 
        WHEN table_name IS NULL THEN '✅ Dropped'
        ELSE '❌ Still exists'
    END as status
FROM information_schema.tables 
WHERE table_schema = 'public' 
    AND table_name IN (
        'risk_keywords', 
        'industry_code_crosswalks', 
        'business_risk_assessments', 
        'risk_keyword_relationships'
    )
ORDER BY table_name;

-- =====================================================
-- Rollback Complete
-- =====================================================
