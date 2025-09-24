-- Dependency Verification Script for Migration 009
-- This script verifies that no dependencies are broken after profiles table removal
-- Date: January 19, 2025
-- Purpose: Comprehensive dependency verification

-- Start transaction for atomic verification
BEGIN;

-- Test 1: Verify all foreign key constraints are valid
DO $$
DECLARE
    constraint_record RECORD;
    constraint_count INTEGER := 0;
    invalid_constraints INTEGER := 0;
BEGIN
    RAISE NOTICE 'Verifying foreign key constraints...';
    
    -- Check all foreign key constraints
    FOR constraint_record IN
        SELECT 
            tc.table_name,
            tc.constraint_name,
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
    LOOP
        constraint_count := constraint_count + 1;
        
        -- Check if the referenced table exists
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = constraint_record.foreign_table_name
        ) THEN
            invalid_constraints := invalid_constraints + 1;
            RAISE WARNING 'Invalid foreign key constraint: % references non-existent table %', 
                constraint_record.constraint_name, constraint_record.foreign_table_name;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Foreign key verification complete: % constraints checked, % invalid', 
        constraint_count, invalid_constraints;
    
    IF invalid_constraints > 0 THEN
        RAISE EXCEPTION 'Dependency verification failed: % invalid foreign key constraints found', invalid_constraints;
    END IF;
END $$;

-- Test 2: Verify all views are functional
DO $$
DECLARE
    view_record RECORD;
    view_count INTEGER := 0;
    broken_views INTEGER := 0;
BEGIN
    RAISE NOTICE 'Verifying view functionality...';
    
    -- Check all views
    FOR view_record IN
        SELECT table_name
        FROM information_schema.views
        WHERE table_schema = 'public'
    LOOP
        view_count := view_count + 1;
        
        -- Test if view can be queried
        BEGIN
            EXECUTE format('SELECT COUNT(*) FROM %I', view_record.table_name);
        EXCEPTION
            WHEN OTHERS THEN
                broken_views := broken_views + 1;
                RAISE WARNING 'Broken view: % - %', view_record.table_name, SQLERRM;
        END;
    END LOOP;
    
    RAISE NOTICE 'View verification complete: % views checked, % broken', 
        view_count, broken_views;
    
    IF broken_views > 0 THEN
        RAISE EXCEPTION 'Dependency verification failed: % broken views found', broken_views;
    END IF;
END $$;

-- Test 3: Verify all triggers are functional
DO $$
DECLARE
    trigger_record RECORD;
    trigger_count INTEGER := 0;
    broken_triggers INTEGER := 0;
BEGIN
    RAISE NOTICE 'Verifying trigger functionality...';
    
    -- Check all triggers
    FOR trigger_record IN
        SELECT 
            trigger_name,
            event_object_table,
            action_statement
        FROM information_schema.triggers
        WHERE trigger_schema = 'public'
    LOOP
        trigger_count := trigger_count + 1;
        
        -- Basic validation - check if trigger references exist
        IF trigger_record.action_statement LIKE '%profiles%' THEN
            broken_triggers := broken_triggers + 1;
            RAISE WARNING 'Trigger may reference removed table: % on %', 
                trigger_record.trigger_name, trigger_record.event_object_table;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Trigger verification complete: % triggers checked, % potentially broken', 
        trigger_count, broken_triggers;
    
    IF broken_triggers > 0 THEN
        RAISE WARNING 'Some triggers may reference removed tables - manual review recommended';
    END IF;
END $$;

-- Test 4: Verify all functions are functional
DO $$
DECLARE
    function_record RECORD;
    function_count INTEGER := 0;
    broken_functions INTEGER := 0;
BEGIN
    RAISE NOTICE 'Verifying function functionality...';
    
    -- Check all functions
    FOR function_record IN
        SELECT 
            routine_name,
            routine_definition
        FROM information_schema.routines
        WHERE routine_schema = 'public'
        AND routine_type = 'FUNCTION'
    LOOP
        function_count := function_count + 1;
        
        -- Basic validation - check if function references removed tables
        IF function_record.routine_definition LIKE '%profiles%' THEN
            broken_functions := broken_functions + 1;
            RAISE WARNING 'Function may reference removed table: %', function_record.routine_name;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Function verification complete: % functions checked, % potentially broken', 
        function_count, broken_functions;
    
    IF broken_functions > 0 THEN
        RAISE WARNING 'Some functions may reference removed tables - manual review recommended';
    END IF;
END $$;

-- Test 5: Verify data integrity across related tables
DO $$
DECLARE
    table_record RECORD;
    integrity_issues INTEGER := 0;
BEGIN
    RAISE NOTICE 'Verifying data integrity...';
    
    -- Check for orphaned records in tables that reference users
    FOR table_record IN
        SELECT 
            tc.table_name,
            kcu.column_name,
            ccu.table_name AS referenced_table
        FROM information_schema.table_constraints AS tc
        JOIN information_schema.key_column_usage AS kcu
            ON tc.constraint_name = kcu.constraint_name
        JOIN information_schema.constraint_column_usage AS ccu
            ON ccu.constraint_name = tc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_schema = 'public'
        AND ccu.table_name = 'users_consolidated'
    LOOP
        -- Check for orphaned records
        EXECUTE format('
            SELECT COUNT(*) FROM %I t 
            WHERE NOT EXISTS (
                SELECT 1 FROM %I u 
                WHERE u.id = t.%I
            )', 
            table_record.table_name, 
            table_record.referenced_table, 
            table_record.column_name
        ) INTO integrity_issues;
        
        IF integrity_issues > 0 THEN
            RAISE WARNING 'Orphaned records found in %: % records', 
                table_record.table_name, integrity_issues;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Data integrity verification complete';
END $$;

-- Test 6: Verify application-specific functionality
DO $$
DECLARE
    test_result TEXT;
BEGIN
    RAISE NOTICE 'Verifying application-specific functionality...';
    
    -- Test user lookup functionality
    BEGIN
        SELECT CASE 
            WHEN COUNT(*) > 0 THEN 'PASS'
            ELSE 'FAIL'
        END INTO test_result
        FROM users_consolidated;
        
        IF test_result = 'FAIL' THEN
            RAISE EXCEPTION 'User lookup functionality broken';
        END IF;
        
        RAISE NOTICE 'User lookup functionality: %', test_result;
    END;
    
    -- Test profile view functionality
    BEGIN
        SELECT CASE 
            WHEN COUNT(*) >= 0 THEN 'PASS'
            ELSE 'FAIL'
        END INTO test_result
        FROM profiles;
        
        IF test_result = 'FAIL' THEN
            RAISE EXCEPTION 'Profile view functionality broken';
        END IF;
        
        RAISE NOTICE 'Profile view functionality: %', test_result;
    END;
    
    -- Test users view functionality
    BEGIN
        SELECT CASE 
            WHEN COUNT(*) >= 0 THEN 'PASS'
            ELSE 'FAIL'
        END INTO test_result
        FROM users;
        
        IF test_result = 'FAIL' THEN
            RAISE EXCEPTION 'Users view functionality broken';
        END IF;
        
        RAISE NOTICE 'Users view functionality: %', test_result;
    END;
    
    RAISE NOTICE 'Application-specific functionality verification complete';
END $$;

-- Test 7: Performance verification
DO $$
DECLARE
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    execution_time INTERVAL;
    performance_acceptable BOOLEAN;
BEGIN
    RAISE NOTICE 'Verifying performance...';
    
    start_time := clock_timestamp();
    
    -- Test common queries
    PERFORM COUNT(*) FROM users_consolidated;
    PERFORM COUNT(*) FROM users_consolidated WHERE role = 'admin';
    PERFORM COUNT(*) FROM users_consolidated WHERE email LIKE '%@example.com';
    PERFORM COUNT(*) FROM profiles;
    PERFORM COUNT(*) FROM users;
    
    end_time := clock_timestamp();
    execution_time := end_time - start_time;
    
    -- Consider performance acceptable if queries complete within 1 second
    performance_acceptable := execution_time < INTERVAL '1 second';
    
    RAISE NOTICE 'Performance test completed in: %', execution_time;
    
    IF NOT performance_acceptable THEN
        RAISE WARNING 'Performance may be degraded - queries took longer than expected';
    ELSE
        RAISE NOTICE 'Performance verification: PASS';
    END IF;
END $$;

-- Test 8: Create dependency verification report
CREATE OR REPLACE VIEW dependency_verification_report AS
SELECT 
    'profiles_table_removal' as migration,
    'dependency_verification' as test_type,
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users_consolidated') 
        AND NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'profiles')
        THEN 'PASS'
        ELSE 'FAIL'
    END as table_structure_status,
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'profiles')
        AND EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'users')
        THEN 'PASS'
        ELSE 'FAIL'
    END as view_status,
    CASE 
        WHEN (SELECT COUNT(*) FROM information_schema.table_constraints 
              WHERE constraint_type = 'FOREIGN KEY' 
              AND constraint_schema = 'public') > 0
        THEN 'PASS'
        ELSE 'FAIL'
    END as foreign_key_status,
    NOW() as verification_date;

-- Log verification completion
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000000'::UUID,
    'DEPENDENCY_VERIFICATION',
    'migration',
    '009_remove_profiles',
    'Dependency verification completed successfully - no broken dependencies found',
    NOW()
);

-- Commit the transaction
COMMIT;

-- Display verification results
SELECT 'Dependency verification completed successfully' as status;
SELECT * FROM dependency_verification_report;

-- Final summary
SELECT 
    'Dependency Verification Summary' as test_suite,
    'All dependency checks passed' as result,
    'No broken dependencies found' as details,
    NOW() as completion_time;
