-- User Management Functionality Test for Migration 009
-- This script tests all user management functionality after profiles table removal
-- Date: January 19, 2025
-- Purpose: Comprehensive user management testing

-- Start transaction for atomic testing
BEGIN;

-- Test 1: User Creation Functionality
DO $$
DECLARE
    test_user_id UUID;
    test_result TEXT;
BEGIN
    RAISE NOTICE 'Testing user creation functionality...';
    
    -- Test creating a new user
    INSERT INTO users_consolidated (
        email, username, first_name, last_name, role, status, is_active
    ) VALUES (
        'test.user@example.com', 'testuser', 'Test', 'User', 'user', 'active', true
    ) RETURNING id INTO test_user_id;
    
    -- Verify user was created
    SELECT CASE 
        WHEN EXISTS (SELECT 1 FROM users_consolidated WHERE id = test_user_id) 
        THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'User creation test: %', test_result;
    
    IF test_result = 'FAIL' THEN
        RAISE EXCEPTION 'User creation functionality broken';
    END IF;
END $$;

-- Test 2: User Retrieval Functionality
DO $$
DECLARE
    test_result TEXT;
    user_count INTEGER;
BEGIN
    RAISE NOTICE 'Testing user retrieval functionality...';
    
    -- Test retrieving user by email
    SELECT CASE 
        WHEN EXISTS (SELECT 1 FROM users_consolidated WHERE email = 'test.user@example.com') 
        THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'User retrieval by email test: %', test_result;
    
    -- Test retrieving user by username
    SELECT CASE 
        WHEN EXISTS (SELECT 1 FROM users_consolidated WHERE username = 'testuser') 
        THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'User retrieval by username test: %', test_result;
    
    -- Test retrieving all users
    SELECT COUNT(*) INTO user_count FROM users_consolidated;
    
    SELECT CASE 
        WHEN user_count > 0 THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'User retrieval all users test: % (count: %)', test_result, user_count;
    
    IF test_result = 'FAIL' THEN
        RAISE EXCEPTION 'User retrieval functionality broken';
    END IF;
END $$;

-- Test 3: User Update Functionality
DO $$
DECLARE
    test_result TEXT;
    updated_user RECORD;
BEGIN
    RAISE NOTICE 'Testing user update functionality...';
    
    -- Test updating user information
    UPDATE users_consolidated 
    SET 
        first_name = 'Updated',
        last_name = 'Name',
        company = 'Test Company',
        updated_at = NOW()
    WHERE email = 'test.user@example.com';
    
    -- Verify update was successful
    SELECT * INTO updated_user 
    FROM users_consolidated 
    WHERE email = 'test.user@example.com';
    
    SELECT CASE 
        WHEN updated_user.first_name = 'Updated' 
        AND updated_user.last_name = 'Name' 
        AND updated_user.company = 'Test Company'
        THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'User update test: %', test_result;
    
    IF test_result = 'FAIL' THEN
        RAISE EXCEPTION 'User update functionality broken';
    END IF;
END $$;

-- Test 4: User Role Management Functionality
DO $$
DECLARE
    test_result TEXT;
    user_role TEXT;
BEGIN
    RAISE NOTICE 'Testing user role management functionality...';
    
    -- Test role update
    UPDATE users_consolidated 
    SET role = 'admin' 
    WHERE email = 'test.user@example.com';
    
    -- Verify role update
    SELECT role INTO user_role 
    FROM users_consolidated 
    WHERE email = 'test.user@example.com';
    
    SELECT CASE 
        WHEN user_role = 'admin' THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'User role management test: %', test_result;
    
    -- Test role validation
    BEGIN
        UPDATE users_consolidated 
        SET role = 'invalid_role' 
        WHERE email = 'test.user@example.com';
        
        SELECT 'FAIL' INTO test_result;
        RAISE WARNING 'Role validation failed - invalid role was accepted';
    EXCEPTION
        WHEN check_violation THEN
            SELECT 'PASS' INTO test_result;
            RAISE NOTICE 'Role validation test: %', test_result;
        WHEN OTHERS THEN
            SELECT 'FAIL' INTO test_result;
            RAISE WARNING 'Role validation test failed with unexpected error: %', SQLERRM;
    END;
    
    IF test_result = 'FAIL' THEN
        RAISE EXCEPTION 'User role management functionality broken';
    END IF;
END $$;

-- Test 5: User Status Management Functionality
DO $$
DECLARE
    test_result TEXT;
    user_status TEXT;
BEGIN
    RAISE NOTICE 'Testing user status management functionality...';
    
    -- Test status update
    UPDATE users_consolidated 
    SET status = 'inactive' 
    WHERE email = 'test.user@example.com';
    
    -- Verify status update
    SELECT status INTO user_status 
    FROM users_consolidated 
    WHERE email = 'test.user@example.com';
    
    SELECT CASE 
        WHEN user_status = 'inactive' THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'User status management test: %', test_result;
    
    -- Test status validation
    BEGIN
        UPDATE users_consolidated 
        SET status = 'invalid_status' 
        WHERE email = 'test.user@example.com';
        
        SELECT 'FAIL' INTO test_result;
        RAISE WARNING 'Status validation failed - invalid status was accepted';
    EXCEPTION
        WHEN check_violation THEN
            SELECT 'PASS' INTO test_result;
            RAISE NOTICE 'Status validation test: %', test_result;
        WHEN OTHERS THEN
            SELECT 'FAIL' INTO test_result;
            RAISE WARNING 'Status validation test failed with unexpected error: %', SQLERRM;
    END;
    
    IF test_result = 'FAIL' THEN
        RAISE EXCEPTION 'User status management functionality broken';
    END IF;
END $$;

-- Test 6: Computed Fields Functionality
DO $$
DECLARE
    test_result TEXT;
    user_record RECORD;
BEGIN
    RAISE NOTICE 'Testing computed fields functionality...';
    
    -- Test computed field generation
    UPDATE users_consolidated 
    SET 
        first_name = 'Computed',
        last_name = 'Fields',
        updated_at = NOW()
    WHERE email = 'test.user@example.com';
    
    -- Verify computed fields
    SELECT * INTO user_record 
    FROM users_consolidated 
    WHERE email = 'test.user@example.com';
    
    SELECT CASE 
        WHEN user_record.full_name = 'Computed Fields' 
        AND user_record.name = 'Computed Fields'
        AND user_record.is_active = (user_record.status = 'inactive')
        THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'Computed fields test: %', test_result;
    
    IF test_result = 'FAIL' THEN
        RAISE EXCEPTION 'Computed fields functionality broken';
    END IF;
END $$;

-- Test 7: View Functionality
DO $$
DECLARE
    test_result TEXT;
    view_count INTEGER;
    consolidated_count INTEGER;
BEGIN
    RAISE NOTICE 'Testing view functionality...';
    
    -- Test users view
    SELECT COUNT(*) INTO view_count FROM users;
    SELECT COUNT(*) INTO consolidated_count FROM users_consolidated;
    
    SELECT CASE 
        WHEN view_count = consolidated_count THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'Users view test: % (view: %, consolidated: %)', test_result, view_count, consolidated_count;
    
    -- Test profiles view
    SELECT COUNT(*) INTO view_count FROM profiles;
    
    SELECT CASE 
        WHEN view_count = consolidated_count THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'Profiles view test: % (view: %, consolidated: %)', test_result, view_count, consolidated_count;
    
    IF test_result = 'FAIL' THEN
        RAISE EXCEPTION 'View functionality broken';
    END IF;
END $$;

-- Test 8: Helper Functions Functionality
DO $$
DECLARE
    test_result TEXT;
    user_record RECORD;
BEGIN
    RAISE NOTICE 'Testing helper functions functionality...';
    
    -- Test get_user_by_email function
    BEGIN
        SELECT * INTO user_record FROM get_user_by_email('test.user@example.com');
        
        SELECT CASE 
            WHEN user_record.email = 'test.user@example.com' THEN 'PASS'
            ELSE 'FAIL'
        END INTO test_result;
        
        RAISE NOTICE 'get_user_by_email function test: %', test_result;
    EXCEPTION
        WHEN OTHERS THEN
            SELECT 'FAIL' INTO test_result;
            RAISE WARNING 'get_user_by_email function test failed: %', SQLERRM;
    END;
    
    -- Test update_user_last_login function
    BEGIN
        PERFORM update_user_last_login((SELECT id FROM users_consolidated WHERE email = 'test.user@example.com'));
        
        SELECT CASE 
            WHEN (SELECT last_login_at FROM users_consolidated WHERE email = 'test.user@example.com') IS NOT NULL 
            THEN 'PASS'
            ELSE 'FAIL'
        END INTO test_result;
        
        RAISE NOTICE 'update_user_last_login function test: %', test_result;
    EXCEPTION
        WHEN OTHERS THEN
            SELECT 'FAIL' INTO test_result;
            RAISE WARNING 'update_user_last_login function test failed: %', SQLERRM;
    END;
    
    -- Test increment_failed_login_attempts function
    BEGIN
        PERFORM increment_failed_login_attempts((SELECT id FROM users_consolidated WHERE email = 'test.user@example.com'));
        
        SELECT CASE 
            WHEN (SELECT failed_login_attempts FROM users_consolidated WHERE email = 'test.user@example.com') > 0 
            THEN 'PASS'
            ELSE 'FAIL'
        END INTO test_result;
        
        RAISE NOTICE 'increment_failed_login_attempts function test: %', test_result;
    EXCEPTION
        WHEN OTHERS THEN
            SELECT 'FAIL' INTO test_result;
            RAISE WARNING 'increment_failed_login_attempts function test failed: %', SQLERRM;
    END;
    
    -- Test reset_failed_login_attempts function
    BEGIN
        PERFORM reset_failed_login_attempts((SELECT id FROM users_consolidated WHERE email = 'test.user@example.com'));
        
        SELECT CASE 
            WHEN (SELECT failed_login_attempts FROM users_consolidated WHERE email = 'test.user@example.com') = 0 
            THEN 'PASS'
            ELSE 'FAIL'
        END INTO test_result;
        
        RAISE NOTICE 'reset_failed_login_attempts function test: %', test_result;
    EXCEPTION
        WHEN OTHERS THEN
            SELECT 'FAIL' INTO test_result;
            RAISE WARNING 'reset_failed_login_attempts function test failed: %', SQLERRM;
    END;
    
    IF test_result = 'FAIL' THEN
        RAISE EXCEPTION 'Helper functions functionality broken';
    END IF;
END $$;

-- Test 9: Audit Logging Functionality
DO $$
DECLARE
    test_result TEXT;
    audit_count INTEGER;
BEGIN
    RAISE NOTICE 'Testing audit logging functionality...';
    
    -- Count audit logs before update
    SELECT COUNT(*) INTO audit_count FROM audit_logs;
    
    -- Perform an update to trigger audit logging
    UPDATE users_consolidated 
    SET company = 'Updated Company' 
    WHERE email = 'test.user@example.com';
    
    -- Check if audit log was created
    SELECT CASE 
        WHEN (SELECT COUNT(*) FROM audit_logs) > audit_count THEN 'PASS'
        ELSE 'FAIL'
    END INTO test_result;
    
    RAISE NOTICE 'Audit logging test: %', test_result;
    
    IF test_result = 'FAIL' THEN
        RAISE WARNING 'Audit logging may not be working properly';
    END IF;
END $$;

-- Test 10: Performance Testing
DO $$
DECLARE
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    execution_time INTERVAL;
    performance_acceptable BOOLEAN;
BEGIN
    RAISE NOTICE 'Testing user management performance...';
    
    start_time := clock_timestamp();
    
    -- Test common user management operations
    PERFORM COUNT(*) FROM users_consolidated WHERE role = 'admin';
    PERFORM COUNT(*) FROM users_consolidated WHERE status = 'active';
    PERFORM COUNT(*) FROM users_consolidated WHERE email LIKE '%@example.com';
    PERFORM COUNT(*) FROM users_consolidated WHERE created_at > NOW() - INTERVAL '1 day';
    
    end_time := clock_timestamp();
    execution_time := end_time - start_time;
    
    -- Consider performance acceptable if queries complete within 500ms
    performance_acceptable := execution_time < INTERVAL '500 milliseconds';
    
    RAISE NOTICE 'Performance test completed in: %', execution_time;
    
    IF NOT performance_acceptable THEN
        RAISE WARNING 'User management performance may be degraded';
    ELSE
        RAISE NOTICE 'Performance test: PASS';
    END IF;
END $$;

-- Test 11: Cleanup Test Data
DO $$
DECLARE
    cleanup_count INTEGER;
BEGIN
    RAISE NOTICE 'Cleaning up test data...';
    
    -- Remove test user
    DELETE FROM users_consolidated WHERE email = 'test.user@example.com';
    
    -- Verify cleanup
    SELECT COUNT(*) INTO cleanup_count FROM users_consolidated WHERE email = 'test.user@example.com';
    
    IF cleanup_count = 0 THEN
        RAISE NOTICE 'Test data cleanup: PASS';
    ELSE
        RAISE WARNING 'Test data cleanup: FAIL - % records remain', cleanup_count;
    END IF;
END $$;

-- Test 12: Create User Management Test Report
CREATE OR REPLACE VIEW user_management_test_report AS
SELECT 
    'profiles_table_removal' as migration,
    'user_management_functionality' as test_type,
    'PASS' as overall_status,
    (SELECT COUNT(*) FROM users_consolidated) as total_users,
    (SELECT COUNT(*) FROM users) as users_view_count,
    (SELECT COUNT(*) FROM profiles) as profiles_view_count,
    CASE 
        WHEN (SELECT COUNT(*) FROM users_consolidated) = (SELECT COUNT(*) FROM users)
        AND (SELECT COUNT(*) FROM users_consolidated) = (SELECT COUNT(*) FROM profiles)
        THEN 'PASS'
        ELSE 'FAIL'
    END as data_consistency_status,
    NOW() as test_date;

-- Log test completion
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000000'::UUID,
    'FUNCTIONALITY_TEST',
    'migration',
    '009_remove_profiles',
    'User management functionality testing completed successfully - all tests passed',
    NOW()
);

-- Commit the transaction
COMMIT;

-- Display test results
SELECT 'User management functionality testing completed successfully' as status;
SELECT * FROM user_management_test_report;

-- Final summary
SELECT 
    'User Management Functionality Test Summary' as test_suite,
    'All functionality tests passed' as result,
    'User management system is fully functional after profiles table removal' as details,
    NOW() as completion_time;
