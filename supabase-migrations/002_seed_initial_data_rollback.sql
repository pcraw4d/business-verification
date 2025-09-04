-- =====================================================
-- Rollback Initial Data Seeding
-- Supabase Implementation
-- =====================================================

-- This script removes the seeded data from the initial data migration
-- Use this if you need to reset the database to its initial state

-- =====================================================
-- 1. Remove Seeded Data in Reverse Order
-- =====================================================

-- Remove keyword weights
DELETE FROM keyword_weights WHERE created_at >= (
    SELECT created_at FROM migrations WHERE version = '002' LIMIT 1
);

-- Remove industry patterns
DELETE FROM industry_patterns WHERE created_at >= (
    SELECT created_at FROM migrations WHERE version = '002' LIMIT 1
);

-- Remove classification codes
DELETE FROM classification_codes WHERE created_at >= (
    SELECT created_at FROM migrations WHERE version = '002' LIMIT 1
);

-- Remove industry keywords
DELETE FROM industry_keywords WHERE created_at >= (
    SELECT created_at FROM migrations WHERE version = '002' LIMIT 1
);

-- Remove industries (except General Business which is the default)
DELETE FROM industries WHERE name != 'General Business' AND created_at >= (
    SELECT created_at FROM migrations WHERE version = '002' LIMIT 1
);

-- Remove audit logs for this migration
DELETE FROM audit_logs WHERE table_name = 'migrations' AND record_id = 2;

-- Remove migration record
DELETE FROM migrations WHERE version = '002';

-- =====================================================
-- 2. Reset Auto-increment Counters
-- =====================================================

-- Reset industry ID sequence (keep General Business as ID 1)
SELECT setval('industries_id_seq', 1);

-- Reset other sequences
SELECT setval('industry_keywords_id_seq', 1);
SELECT setval('classification_codes_id_seq', 1);
SELECT setval('industry_patterns_id_seq', 1);
SELECT setval('keyword_weights_id_seq', 1);
SELECT setval('audit_logs_id_seq', 1);

-- =====================================================
-- 3. Verify Rollback
-- =====================================================

-- Check remaining data
SELECT 'industries' as table_name, COUNT(*) as record_count FROM industries
UNION ALL
SELECT 'industry_keywords' as table_name, COUNT(*) as record_count FROM industry_keywords
UNION ALL
SELECT 'classification_codes' as table_name, COUNT(*) as record_count FROM classification_codes
UNION ALL
SELECT 'industry_patterns' as table_name, COUNT(*) as record_count FROM industry_patterns
UNION ALL
SELECT 'keyword_weights' as table_name, COUNT(*) as record_count FROM keyword_weights;

-- =====================================================
-- Rollback Complete
-- =====================================================

-- The database has been reset to its initial state with:
-- ✅ All seeded data removed
-- ✅ Only General Business industry remains (default)
-- ✅ Auto-increment sequences reset
-- ✅ Migration record removed
-- ✅ Audit trail cleaned up

-- You can now re-run the seeding migration or start fresh.
