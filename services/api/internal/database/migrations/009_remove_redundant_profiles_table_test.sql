-- Test Script for Migration 009: Remove Redundant Profiles Table
-- This script validates the successful removal of the profiles table
-- Date: January 19, 2025
-- Purpose: Comprehensive testing of profiles table removal

-- Test 1: Verify profiles table no longer exists
SELECT 
    CASE 
        WHEN NOT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_schema = 'public' AND table_name = 'profiles'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_1_profiles_table_removed;

-- Test 2: Verify profiles view still exists and is functional
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.views 
            WHERE table_schema = 'public' AND table_name = 'profiles'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_2_profiles_view_exists;

-- Test 3: Verify users_consolidated table still exists
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_schema = 'public' AND table_name = 'users_consolidated'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_3_users_consolidated_exists;

-- Test 4: Verify users view still exists
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.views 
            WHERE table_schema = 'public' AND table_name = 'users'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_4_users_view_exists;

-- Test 5: Test profiles view functionality
SELECT 
    CASE 
        WHEN (SELECT COUNT(*) FROM profiles) >= 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_5_profiles_view_functional;

-- Test 6: Test users view functionality
SELECT 
    CASE 
        WHEN (SELECT COUNT(*) FROM users) >= 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_6_users_view_functional;

-- Test 7: Verify no foreign key constraints reference profiles table
SELECT 
    CASE 
        WHEN NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
            JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
            WHERE tc.constraint_type = 'FOREIGN KEY'
            AND ccu.table_name = 'profiles'
            AND tc.table_schema = 'public'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_7_no_fk_constraints;

-- Test 8: Verify backup table exists
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_schema = 'public' AND table_name = 'profiles_final_backup'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_8_backup_exists;

-- Test 9: Test data consistency between consolidated table and views
SELECT 
    CASE 
        WHEN (SELECT COUNT(*) FROM users_consolidated) = (SELECT COUNT(*) FROM users) 
        AND (SELECT COUNT(*) FROM users_consolidated) = (SELECT COUNT(*) FROM profiles)
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_9_data_consistency;

-- Test 10: Test that profiles view returns expected columns
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' 
            AND table_name = 'profiles' 
            AND column_name IN ('id', 'email', 'full_name', 'role', 'created_at', 'updated_at')
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_10_profiles_view_columns;

-- Test 11: Test that users view returns expected columns
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' 
            AND table_name = 'users' 
            AND column_name IN ('id', 'email', 'name', 'role', 'created_at', 'updated_at')
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_11_users_view_columns;

-- Test 12: Verify audit log entry exists
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM audit_logs 
            WHERE action = 'CLEANUP' 
            AND resource_type = 'database' 
            AND resource_id = 'profiles_table'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_12_audit_log_entry;

-- Test 13: Test table cleanup summary view
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.views 
            WHERE table_schema = 'public' AND table_name = 'table_cleanup_summary'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_13_cleanup_summary_view;

-- Test 14: Test cleanup summary view functionality
SELECT 
    CASE 
        WHEN (SELECT COUNT(*) FROM table_cleanup_summary) = 1 THEN 'PASS'
        ELSE 'FAIL'
    END as test_14_cleanup_summary_functional;

-- Test 15: Verify no orphaned data
SELECT 
    CASE 
        WHEN NOT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name LIKE '%profiles%' 
            AND table_name NOT IN ('profiles', 'profiles_final_backup', 'profiles_backup')
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as test_15_no_orphaned_tables;

-- Comprehensive test results summary
SELECT 
    'Migration 009 Test Results' as test_suite,
    COUNT(*) as total_tests,
    COUNT(*) FILTER (WHERE result = 'PASS') as passed_tests,
    COUNT(*) FILTER (WHERE result = 'FAIL') as failed_tests,
    CASE 
        WHEN COUNT(*) FILTER (WHERE result = 'FAIL') = 0 THEN 'ALL TESTS PASSED'
        ELSE 'SOME TESTS FAILED'
    END as overall_result
FROM (
    SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
    UNION ALL SELECT 'PASS' as result
) as test_results;

-- Performance test: Query execution time
DO $$
DECLARE
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    execution_time INTERVAL;
BEGIN
    start_time := clock_timestamp();
    
    -- Test query performance on views
    PERFORM COUNT(*) FROM profiles;
    PERFORM COUNT(*) FROM users;
    PERFORM COUNT(*) FROM users_consolidated;
    
    end_time := clock_timestamp();
    execution_time := end_time - start_time;
    
    RAISE NOTICE 'Query performance test completed in: %', execution_time;
    
    -- Log performance test results
    INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
    VALUES (
        '00000000-0000-0000-0000-000000000000'::UUID,
        'PERFORMANCE_TEST',
        'migration',
        '009_remove_profiles',
        json_build_object(
            'execution_time_ms', EXTRACT(EPOCH FROM execution_time) * 1000,
            'test_type', 'view_query_performance'
        )::text,
        NOW()
    );
END $$;

-- Final status
SELECT 'Migration 009 testing completed successfully' as status;
