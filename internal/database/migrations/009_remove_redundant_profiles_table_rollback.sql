-- Rollback Script for Migration 009: Remove Redundant Profiles Table
-- This script restores the profiles table from backup in case of issues
-- Date: January 19, 2025
-- Purpose: Emergency rollback for profiles table removal

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Start transaction for atomic rollback
BEGIN;

-- Step 1: Verify rollback prerequisites
DO $$
DECLARE
    backup_exists INTEGER;
    consolidated_exists INTEGER;
BEGIN
    -- Check if backup table exists
    SELECT COUNT(*) INTO backup_exists FROM information_schema.tables 
    WHERE table_schema = 'public' AND table_name = 'profiles_final_backup';
    
    -- Check if consolidated table exists
    SELECT COUNT(*) INTO consolidated_exists FROM information_schema.tables 
    WHERE table_schema = 'public' AND table_name = 'users_consolidated';
    
    IF backup_exists = 0 THEN
        RAISE EXCEPTION 'Rollback failed: profiles_final_backup table not found';
    END IF;
    
    IF consolidated_exists = 0 THEN
        RAISE EXCEPTION 'Rollback failed: users_consolidated table not found';
    END IF;
    
    RAISE NOTICE 'Rollback prerequisites verified';
END $$;

-- Step 2: Check if profiles table already exists
DO $$
DECLARE
    table_exists INTEGER;
BEGIN
    SELECT COUNT(*) INTO table_exists FROM information_schema.tables 
    WHERE table_schema = 'public' AND table_name = 'profiles';
    
    IF table_exists > 0 THEN
        RAISE EXCEPTION 'Profiles table already exists, cannot rollback';
    END IF;
    
    RAISE NOTICE 'Profiles table does not exist, proceeding with rollback';
END $$;

-- Step 3: Restore profiles table from backup
CREATE TABLE public.profiles (
    id UUID REFERENCES auth.users(id) PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT,
    role TEXT CHECK (role IN ('compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'other')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Step 4: Restore data from backup
INSERT INTO public.profiles (id, email, full_name, role, created_at, updated_at)
SELECT id, email, full_name, role, created_at, updated_at
FROM profiles_final_backup;

-- Step 5: Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_profiles_email ON public.profiles(email);
CREATE INDEX IF NOT EXISTS idx_profiles_role ON public.profiles(role);
CREATE INDEX IF NOT EXISTS idx_profiles_created_at ON public.profiles(created_at);

-- Step 6: Restore foreign key constraints
-- Note: This would need to be updated based on the original schema
-- For now, we'll create the basic structure

-- Step 7: Verify data restoration
DO $$
DECLARE
    backup_count INTEGER;
    restored_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO backup_count FROM profiles_final_backup;
    SELECT COUNT(*) INTO restored_count FROM public.profiles;
    
    IF backup_count != restored_count THEN
        RAISE EXCEPTION 'Data restoration failed: backup has % records, restored has % records', 
            backup_count, restored_count;
    END IF;
    
    RAISE NOTICE 'Data restoration verified: % records restored', restored_count;
END $$;

-- Step 8: Drop the profiles view to avoid conflicts
DROP VIEW IF EXISTS profiles CASCADE;

-- Step 9: Create rollback completion log
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000000'::UUID,
    'ROLLBACK',
    'database',
    'profiles_table',
    'Profiles table restored from backup due to rollback',
    NOW()
);

-- Step 10: Create rollback summary
CREATE OR REPLACE VIEW rollback_summary AS
SELECT 
    'profiles' as restored_table,
    (SELECT COUNT(*) FROM public.profiles) as restored_records,
    'rollback_complete' as status,
    NOW() as rollback_date;

-- Commit the transaction
COMMIT;

-- Final verification
SELECT 'Profiles table rollback completed successfully' as status;
SELECT * FROM rollback_summary;

-- Important Notes:
-- 1. After rollback, you may need to update foreign key constraints
-- 2. Application code may need to be updated to use profiles table again
-- 3. Consider running the consolidation migration again after fixing issues
-- 4. The profiles_final_backup table can be dropped after successful rollback
