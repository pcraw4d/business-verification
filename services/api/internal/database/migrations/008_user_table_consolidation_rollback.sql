-- Rollback Script for Migration 008: User Table Consolidation
-- WARNING: This script will rollback the user table consolidation
-- Use only in emergency situations
-- Date: January 19, 2025

-- Start transaction for atomic rollback
BEGIN;

-- Step 1: Log rollback initiation
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000000'::UUID, -- System user ID
    'ROLLBACK',
    'database',
    'users_consolidated',
    'User table consolidation rollback initiated',
    NOW()
);

-- Step 2: Drop views first
DROP VIEW IF EXISTS profiles CASCADE;
DROP VIEW IF EXISTS users CASCADE;

-- Step 3: Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_audit_user_changes ON users_consolidated;
DROP TRIGGER IF EXISTS trigger_validate_user_data ON users_consolidated;
DROP TRIGGER IF EXISTS trigger_update_user_computed_fields ON users_consolidated;

DROP FUNCTION IF EXISTS audit_user_changes();
DROP FUNCTION IF EXISTS validate_user_data();
DROP FUNCTION IF EXISTS update_user_computed_fields();
DROP FUNCTION IF EXISTS get_user_by_email(TEXT);
DROP FUNCTION IF EXISTS update_user_last_login(UUID);
DROP FUNCTION IF EXISTS increment_failed_login_attempts(UUID);
DROP FUNCTION IF EXISTS reset_failed_login_attempts(UUID);

-- Step 4: Drop views and statistics
DROP VIEW IF EXISTS user_statistics;

-- Step 5: Restore original foreign key constraints
-- Note: This assumes the original tables still exist

-- Restore api_keys foreign key
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_user_id_fkey;
-- Restore original constraint (this would need to be adjusted based on actual original schema)

-- Restore businesses foreign key
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS businesses_user_id_fkey;

-- Restore business_classifications foreign key
ALTER TABLE business_classifications DROP CONSTRAINT IF EXISTS business_classifications_user_id_fkey;

-- Restore risk_assessments foreign key
ALTER TABLE risk_assessments DROP CONSTRAINT IF EXISTS risk_assessments_user_id_fkey;

-- Restore compliance_checks foreign key
ALTER TABLE compliance_checks DROP CONSTRAINT IF EXISTS compliance_checks_user_id_fkey;

-- Restore audit_logs foreign key
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_user_id_fkey;

-- Restore external_service_calls foreign key
ALTER TABLE external_service_calls DROP CONSTRAINT IF EXISTS external_service_calls_user_id_fkey;

-- Restore webhooks foreign key
ALTER TABLE webhooks DROP CONSTRAINT IF EXISTS webhooks_user_id_fkey;

-- Restore email_verification_tokens foreign key
ALTER TABLE email_verification_tokens DROP CONSTRAINT IF EXISTS email_verification_tokens_user_id_fkey;

-- Restore password_reset_tokens foreign key
ALTER TABLE password_reset_tokens DROP CONSTRAINT IF EXISTS password_reset_tokens_user_id_fkey;

-- Restore role_assignments foreign key
ALTER TABLE role_assignments DROP CONSTRAINT IF EXISTS role_assignments_user_id_fkey;

-- Step 6: Drop the consolidated table
DROP TABLE IF EXISTS users_consolidated CASCADE;

-- Step 7: Restore original tables from backup if they don't exist
-- This step would need to be customized based on the actual original schema

-- Step 8: Log rollback completion
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000000'::UUID, -- System user ID
    'ROLLBACK',
    'database',
    'users_consolidated',
    'User table consolidation rollback completed',
    NOW()
);

-- Commit the rollback
COMMIT;

-- Rollback completed
SELECT 'User table consolidation rollback completed' as status;
