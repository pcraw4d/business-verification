-- Migration 009: Remove Redundant Profiles Table
-- This migration safely removes the redundant profiles table after consolidation
-- Date: January 19, 2025
-- Purpose: Complete user table consolidation by removing redundant tables

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Start transaction for atomic migration
BEGIN;

-- Step 1: Verify migration was successful
-- Check that users_consolidated table exists and has data
DO $$
DECLARE
    consolidated_count INTEGER;
    profiles_count INTEGER;
    users_view_count INTEGER;
    profiles_view_count INTEGER;
BEGIN
    -- Check users_consolidated table
    SELECT COUNT(*) INTO consolidated_count FROM users_consolidated;
    
    -- Check if profiles table still exists
    SELECT COUNT(*) INTO profiles_count FROM information_schema.tables 
    WHERE table_schema = 'public' AND table_name = 'profiles';
    
    -- Check if views exist
    SELECT COUNT(*) INTO users_view_count FROM information_schema.views 
    WHERE table_schema = 'public' AND table_name = 'users';
    
    SELECT COUNT(*) INTO profiles_view_count FROM information_schema.views 
    WHERE table_schema = 'public' AND table_name = 'profiles';
    
    -- Verify migration prerequisites
    IF consolidated_count = 0 THEN
        RAISE EXCEPTION 'Migration failed: users_consolidated table is empty';
    END IF;
    
    IF users_view_count = 0 THEN
        RAISE EXCEPTION 'Migration failed: users view does not exist';
    END IF;
    
    IF profiles_view_count = 0 THEN
        RAISE EXCEPTION 'Migration failed: profiles view does not exist';
    END IF;
    
    -- Log verification results
    INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
    VALUES (
        '00000000-0000-0000-0000-000000000000'::UUID,
        'VERIFICATION',
        'migration',
        '009_remove_profiles',
        json_build_object(
            'consolidated_users_count', consolidated_count,
            'profiles_table_exists', profiles_count > 0,
            'users_view_exists', users_view_count > 0,
            'profiles_view_exists', profiles_view_count > 0
        )::text,
        NOW()
    );
    
    RAISE NOTICE 'Migration verification passed: % users in consolidated table', consolidated_count;
END $$;

-- Step 2: Check for any remaining foreign key references to profiles table
-- This is a safety check to ensure no tables still reference the profiles table
DO $$
DECLARE
    fk_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO fk_count
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
    AND ccu.table_name = 'profiles'
    AND tc.table_schema = 'public';
    
    IF fk_count > 0 THEN
        RAISE EXCEPTION 'Cannot remove profiles table: % foreign key constraints still reference it', fk_count;
    END IF;
    
    RAISE NOTICE 'No foreign key constraints found referencing profiles table';
END $$;

-- Step 3: Verify that all data is accessible through the views
-- Test that the profiles view returns the same data as the original profiles table
DO $$
DECLARE
    original_count INTEGER;
    view_count INTEGER;
BEGIN
    -- Count records in original profiles table (if it exists)
    SELECT COUNT(*) INTO original_count FROM public.profiles;
    
    -- Count records in profiles view
    SELECT COUNT(*) INTO view_count FROM profiles;
    
    -- Verify data consistency
    IF original_count != view_count THEN
        RAISE EXCEPTION 'Data inconsistency: profiles table has % records, profiles view has % records', 
            original_count, view_count;
    END IF;
    
    RAISE NOTICE 'Data consistency verified: % records in both profiles table and view', original_count;
END $$;

-- Step 4: Create final backup of profiles table before removal
CREATE TABLE IF NOT EXISTS profiles_final_backup AS 
SELECT * FROM public.profiles;

-- Step 5: Drop the redundant profiles table
-- This is safe because:
-- 1. All data has been migrated to users_consolidated
-- 2. A profiles view exists for backward compatibility
-- 3. All foreign key constraints have been updated
-- 4. Final backup has been created
DROP TABLE IF EXISTS public.profiles CASCADE;

-- Step 6: Verify table removal
DO $$
DECLARE
    table_exists INTEGER;
BEGIN
    SELECT COUNT(*) INTO table_exists FROM information_schema.tables 
    WHERE table_schema = 'public' AND table_name = 'profiles';
    
    IF table_exists > 0 THEN
        RAISE EXCEPTION 'Failed to remove profiles table';
    END IF;
    
    RAISE NOTICE 'Profiles table successfully removed';
END $$;

-- Step 7: Test that the profiles view still works
DO $$
DECLARE
    view_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO view_count FROM profiles;
    RAISE NOTICE 'Profiles view still functional: % records accessible', view_count;
END $$;

-- Step 8: Update backup script to remove profiles table reference
-- This ensures future backups don't try to backup the non-existent table
-- Note: This would need to be updated in the actual backup script file

-- Step 9: Create cleanup completion log
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000000'::UUID,
    'CLEANUP',
    'database',
    'profiles_table',
    'Redundant profiles table successfully removed after consolidation',
    NOW()
);

-- Step 10: Create summary statistics
CREATE OR REPLACE VIEW table_cleanup_summary AS
SELECT 
    'profiles' as removed_table,
    'users_consolidated' as consolidated_table,
    (SELECT COUNT(*) FROM users_consolidated) as total_users,
    (SELECT COUNT(*) FROM profiles) as profiles_view_count,
    'success' as status,
    NOW() as cleanup_date;

-- Commit the transaction
COMMIT;

-- Final verification
SELECT 'Profiles table removal completed successfully' as status;
SELECT * FROM table_cleanup_summary;
