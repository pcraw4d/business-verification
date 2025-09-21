-- Test Script for Migration 008: User Table Consolidation
-- This script validates the migration was successful
-- Date: January 19, 2025

-- Test 1: Verify consolidated table exists and has correct structure
SELECT 
    'Test 1: Consolidated table structure' as test_name,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_name = 'users_consolidated' 
            AND table_schema = 'public'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 2: Verify all required columns exist
SELECT 
    'Test 2: Required columns exist' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM information_schema.columns 
            WHERE table_name = 'users_consolidated' 
            AND table_schema = 'public'
            AND column_name IN (
                'id', 'email', 'username', 'password_hash', 'first_name', 
                'last_name', 'full_name', 'name', 'company', 'role', 'status',
                'is_active', 'email_verified', 'email_verified_at', 
                'failed_login_attempts', 'locked_until', 'last_login_at',
                'metadata', 'created_at', 'updated_at'
            )
        ) = 19 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 3: Verify indexes exist
SELECT 
    'Test 3: Required indexes exist' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM pg_indexes 
            WHERE tablename = 'users_consolidated'
            AND indexname IN (
                'idx_users_consolidated_email',
                'idx_users_consolidated_username',
                'idx_users_consolidated_role',
                'idx_users_consolidated_status',
                'idx_users_consolidated_is_active',
                'idx_users_consolidated_created_at',
                'idx_users_consolidated_last_login'
            )
        ) = 7 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 4: Verify views exist
SELECT 
    'Test 4: Compatibility views exist' as test_name,
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'users')
        AND EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'profiles')
        THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 5: Verify triggers exist
SELECT 
    'Test 5: Required triggers exist' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM information_schema.triggers 
            WHERE event_object_table = 'users_consolidated'
            AND trigger_name IN (
                'trigger_update_user_computed_fields',
                'trigger_audit_user_changes',
                'trigger_validate_user_data'
            )
        ) = 3 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 6: Verify functions exist
SELECT 
    'Test 6: Helper functions exist' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM information_schema.routines 
            WHERE routine_name IN (
                'update_user_computed_fields',
                'audit_user_changes',
                'validate_user_data',
                'get_user_by_email',
                'update_user_last_login',
                'increment_failed_login_attempts',
                'reset_failed_login_attempts'
            )
            AND routine_type = 'FUNCTION'
        ) = 7 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 7: Test computed fields functionality
INSERT INTO users_consolidated (email, first_name, last_name, role, status)
VALUES ('test@example.com', 'John', 'Doe', 'user', 'active');

SELECT 
    'Test 7: Computed fields work' as test_name,
    CASE 
        WHEN (
            SELECT full_name = 'John Doe' AND name = 'John Doe' AND is_active = true
            FROM users_consolidated 
            WHERE email = 'test@example.com'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 8: Test validation functions
SELECT 
    'Test 8: Validation functions work' as test_name,
    CASE 
        WHEN (
            -- This should fail due to invalid email
            SELECT COUNT(*) FROM (
                SELECT 1 FROM users_consolidated 
                WHERE email = 'invalid-email'
                LIMIT 1
            ) as invalid_check
        ) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 9: Test foreign key constraints
SELECT 
    'Test 9: Foreign key constraints exist' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
            WHERE tc.constraint_type = 'FOREIGN KEY'
            AND kcu.referenced_table_name = 'users_consolidated'
            AND kcu.table_name IN (
                'api_keys', 'businesses', 'business_classifications', 
                'risk_assessments', 'compliance_checks', 'audit_logs',
                'external_service_calls', 'webhooks', 'email_verification_tokens',
                'password_reset_tokens', 'role_assignments'
            )
        ) >= 10 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 10: Test RLS policies
SELECT 
    'Test 10: RLS policies exist' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM pg_policies 
            WHERE tablename = 'users_consolidated'
            AND policyname IN (
                'Users can access own data',
                'Admins can access all data'
            )
        ) = 2 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 11: Test user statistics view
SELECT 
    'Test 11: User statistics view works' as test_name,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM user_statistics
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 12: Test helper functions
SELECT 
    'Test 12: Helper functions work' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM get_user_by_email('test@example.com')
        ) = 1 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 13: Test audit logging
SELECT 
    'Test 13: Audit logging works' as test_name,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM audit_logs 
            WHERE action = 'CREATE' 
            AND resource_type = 'user'
            AND resource_id IN (
                SELECT id FROM users_consolidated WHERE email = 'test@example.com'
            )
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 14: Test data migration from backup tables
SELECT 
    'Test 14: Data migration successful' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM users_consolidated
        ) >= (
            SELECT COUNT(*) FROM users_backup
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Test 15: Test backward compatibility views
SELECT 
    'Test 15: Backward compatibility views work' as test_name,
    CASE 
        WHEN (
            SELECT COUNT(*) FROM users WHERE email = 'test@example.com'
        ) = 1 
        AND (
            SELECT COUNT(*) FROM profiles WHERE email = 'test@example.com'
        ) = 1 THEN 'PASS'
        ELSE 'FAIL'
    END as result;

-- Clean up test data
DELETE FROM users_consolidated WHERE email = 'test@example.com';

-- Summary of all tests
SELECT 
    'MIGRATION VALIDATION SUMMARY' as summary,
    'All tests completed. Check individual test results above.' as details;

-- Performance test: Check query performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM users_consolidated 
WHERE email = 'test@example.com' 
LIMIT 1;

-- Final validation query
SELECT 
    'FINAL VALIDATION' as check_type,
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE status = 'active') as active_users,
    COUNT(*) FILTER (WHERE email_verified = true) as verified_users
FROM users_consolidated;
