-- =====================================================
-- Keyword Classification System Schema Rollback
-- Supabase Implementation
-- =====================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_keyword_weights_updated_at ON keyword_weights;
DROP TRIGGER IF EXISTS update_industry_patterns_updated_at ON industry_patterns;
DROP TRIGGER IF EXISTS update_classification_codes_updated_at ON classification_codes;
DROP TRIGGER IF EXISTS update_industry_keywords_updated_at ON industry_keywords;
DROP TRIGGER IF EXISTS update_industries_updated_at ON industries;

-- Drop the update function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS keyword_weights CASCADE;
DROP TABLE IF EXISTS industry_patterns CASCADE;
DROP TABLE IF EXISTS code_keywords CASCADE;
DROP TABLE IF EXISTS classification_codes CASCADE;
DROP TABLE IF EXISTS industry_keywords CASCADE;
DROP TABLE IF EXISTS industries CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";

-- =====================================================
-- Rollback Complete
-- =====================================================

-- Verify tables are dropped
SELECT 
    table_name, 
    table_type 
FROM information_schema.tables 
WHERE table_schema = 'public' 
    AND table_name IN (
        'industries', 
        'industry_keywords', 
        'classification_codes', 
        'code_keywords', 
        'industry_patterns', 
        'keyword_weights', 
        'audit_logs'
    )
ORDER BY table_name;
