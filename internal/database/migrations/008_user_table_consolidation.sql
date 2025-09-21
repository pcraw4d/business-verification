-- Migration 008: User Table Consolidation
-- This migration consolidates the conflicting user table definitions into a single, comprehensive table
-- Date: January 19, 2025
-- Purpose: Resolve user table conflicts and create unified user management

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Start transaction for atomic migration
BEGIN;

-- Step 1: Create backup tables for existing data
CREATE TABLE IF NOT EXISTS users_backup AS SELECT * FROM users;
CREATE TABLE IF NOT EXISTS profiles_backup AS SELECT * FROM public.profiles;

-- Step 2: Create the consolidated users table with comprehensive schema
-- This combines the best features from all three existing table definitions
CREATE TABLE IF NOT EXISTS users_consolidated (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Authentication fields
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255),
    
    -- Profile information
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    full_name VARCHAR(255), -- Computed field for compatibility
    name VARCHAR(255), -- Single name field for compatibility
    
    -- Business information
    company VARCHAR(255),
    
    -- Role and permissions
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN (
        'user', 'admin', 'compliance_officer', 'risk_manager', 
        'business_analyst', 'developer', 'other'
    )),
    
    -- Account status and security
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN (
        'active', 'inactive', 'suspended', 'pending_verification'
    )),
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Email verification
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    
    -- Security features
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    
    -- Activity tracking
    last_login_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata and extensibility
    metadata JSONB DEFAULT '{}',
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT users_consolidated_email_check CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT users_consolidated_username_check CHECK (username IS NULL OR length(username) >= 3),
    CONSTRAINT users_consolidated_name_check CHECK (
        (first_name IS NOT NULL AND last_name IS NOT NULL) OR 
        (full_name IS NOT NULL) OR 
        (name IS NOT NULL)
    )
);

-- Step 3: Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_consolidated_email ON users_consolidated(email);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_username ON users_consolidated(username) WHERE username IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_consolidated_role ON users_consolidated(role);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_status ON users_consolidated(status);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_is_active ON users_consolidated(is_active);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_created_at ON users_consolidated(created_at);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_last_login ON users_consolidated(last_login_at);

-- Step 4: Create function to populate computed fields
CREATE OR REPLACE FUNCTION update_user_computed_fields()
RETURNS TRIGGER AS $$
BEGIN
    -- Update full_name from first_name and last_name
    IF NEW.first_name IS NOT NULL AND NEW.last_name IS NOT NULL THEN
        NEW.full_name = NEW.first_name || ' ' || NEW.last_name;
    END IF;
    
    -- Update name field for compatibility
    IF NEW.full_name IS NOT NULL THEN
        NEW.name = NEW.full_name;
    ELSIF NEW.first_name IS NOT NULL AND NEW.last_name IS NOT NULL THEN
        NEW.name = NEW.first_name || ' ' || NEW.last_name;
    END IF;
    
    -- Update is_active based on status
    NEW.is_active = (NEW.status = 'active');
    
    -- Update updated_at timestamp
    NEW.updated_at = NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 5: Create trigger for computed fields
DROP TRIGGER IF EXISTS trigger_update_user_computed_fields ON users_consolidated;
CREATE TRIGGER trigger_update_user_computed_fields
    BEFORE INSERT OR UPDATE ON users_consolidated
    FOR EACH ROW
    EXECUTE FUNCTION update_user_computed_fields();

-- Step 6: Migrate data from existing users table (supabase_schema.sql version)
INSERT INTO users_consolidated (
    id, email, name, password_hash, email_verified, email_verified_at,
    created_at, updated_at, last_login_at, is_active, role, metadata
)
SELECT 
    id,
    email,
    name,
    password_hash,
    email_verified,
    email_verified_at,
    created_at,
    updated_at,
    last_login_at,
    is_active,
    role,
    metadata
FROM users_backup
WHERE NOT EXISTS (
    SELECT 1 FROM users_consolidated WHERE users_consolidated.id = users_backup.id
)
ON CONFLICT (id) DO NOTHING;

-- Step 7: Migrate data from profiles table (setup-supabase-schema.sql version)
-- Map profiles to users_consolidated, handling the auth.users reference
INSERT INTO users_consolidated (
    id, email, full_name, role, created_at, updated_at
)
SELECT 
    p.id,
    p.email,
    p.full_name,
    CASE 
        WHEN p.role IN ('compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'other') 
        THEN p.role
        ELSE 'user'
    END,
    p.created_at,
    p.updated_at
FROM profiles_backup p
WHERE NOT EXISTS (
    SELECT 1 FROM users_consolidated WHERE users_consolidated.id = p.id
)
ON CONFLICT (id) DO UPDATE SET
    full_name = EXCLUDED.full_name,
    role = EXCLUDED.role,
    updated_at = NOW();

-- Step 8: Handle users from migration schema (001_initial_schema.sql)
-- This would be needed if that schema was actually deployed
-- For now, we'll create a placeholder for future migration if needed

-- Step 9: Update foreign key references to point to consolidated table
-- First, we need to temporarily disable foreign key constraints

-- Update api_keys table
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_user_id_fkey;
ALTER TABLE api_keys ADD CONSTRAINT api_keys_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update businesses table
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS businesses_user_id_fkey;
ALTER TABLE businesses ADD CONSTRAINT businesses_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update business_classifications table (handle both references)
ALTER TABLE business_classifications DROP CONSTRAINT IF EXISTS business_classifications_user_id_fkey;
ALTER TABLE business_classifications ADD CONSTRAINT business_classifications_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update risk_assessments table
ALTER TABLE risk_assessments DROP CONSTRAINT IF EXISTS risk_assessments_user_id_fkey;
ALTER TABLE risk_assessments ADD CONSTRAINT risk_assessments_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update compliance_checks table
ALTER TABLE compliance_checks DROP CONSTRAINT IF EXISTS compliance_checks_user_id_fkey;
ALTER TABLE compliance_checks ADD CONSTRAINT compliance_checks_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update audit_logs table
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_user_id_fkey;
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update external_service_calls table
ALTER TABLE external_service_calls DROP CONSTRAINT IF EXISTS external_service_calls_user_id_fkey;
ALTER TABLE external_service_calls ADD CONSTRAINT external_service_calls_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update webhooks table
ALTER TABLE webhooks DROP CONSTRAINT IF EXISTS webhooks_user_id_fkey;
ALTER TABLE webhooks ADD CONSTRAINT webhooks_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update email_verification_tokens table
ALTER TABLE email_verification_tokens DROP CONSTRAINT IF EXISTS email_verification_tokens_user_id_fkey;
ALTER TABLE email_verification_tokens ADD CONSTRAINT email_verification_tokens_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update password_reset_tokens table
ALTER TABLE password_reset_tokens DROP CONSTRAINT IF EXISTS password_reset_tokens_user_id_fkey;
ALTER TABLE password_reset_tokens ADD CONSTRAINT password_reset_tokens_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Update role_assignments table
ALTER TABLE role_assignments DROP CONSTRAINT IF EXISTS role_assignments_user_id_fkey;
ALTER TABLE role_assignments ADD CONSTRAINT role_assignments_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_consolidated(id) ON DELETE CASCADE;

-- Step 10: Create view for backward compatibility
CREATE OR REPLACE VIEW users AS
SELECT 
    id,
    email,
    username,
    password_hash,
    first_name,
    last_name,
    full_name as name, -- Map full_name to name for compatibility
    company,
    role,
    status,
    email_verified,
    email_verified_at,
    last_login_at,
    is_active,
    metadata,
    created_at,
    updated_at
FROM users_consolidated;

-- Step 11: Create view for profiles compatibility
CREATE OR REPLACE VIEW profiles AS
SELECT 
    id,
    email,
    full_name,
    role,
    created_at,
    updated_at
FROM users_consolidated;

-- Step 12: Create RLS (Row Level Security) policies if needed
-- Enable RLS on the consolidated table
ALTER TABLE users_consolidated ENABLE ROW LEVEL SECURITY;

-- Create policy for users to access their own data
CREATE POLICY "Users can access own data" ON users_consolidated
    FOR ALL USING (auth.uid() = id);

-- Create policy for admins to access all data
CREATE POLICY "Admins can access all data" ON users_consolidated
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM users_consolidated 
            WHERE id = auth.uid() AND role = 'admin'
        )
    );

-- Step 13: Create audit trigger for user changes
CREATE OR REPLACE FUNCTION audit_user_changes()
RETURNS TRIGGER AS $$
BEGIN
    -- Log user changes to audit_logs if the table exists
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
        VALUES (NEW.id, 'CREATE', 'user', NEW.id, 
                json_build_object('email', NEW.email, 'role', NEW.role)::text, NOW());
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
        VALUES (NEW.id, 'UPDATE', 'user', NEW.id, 
                json_build_object('changes', json_build_object(
                    'email', CASE WHEN OLD.email != NEW.email THEN json_build_object('old', OLD.email, 'new', NEW.email) ELSE NULL END,
                    'role', CASE WHEN OLD.role != NEW.role THEN json_build_object('old', OLD.role, 'new', NEW.role) ELSE NULL END,
                    'status', CASE WHEN OLD.status != NEW.status THEN json_build_object('old', OLD.status, 'new', NEW.status) ELSE NULL END
                ))::text, NOW());
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
        VALUES (OLD.id, 'DELETE', 'user', OLD.id, 
                json_build_object('email', OLD.email, 'role', OLD.role)::text, NOW());
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create audit trigger
DROP TRIGGER IF EXISTS trigger_audit_user_changes ON users_consolidated;
CREATE TRIGGER trigger_audit_user_changes
    AFTER INSERT OR UPDATE OR DELETE ON users_consolidated
    FOR EACH ROW
    EXECUTE FUNCTION audit_user_changes();

-- Step 14: Create helper functions for user management
CREATE OR REPLACE FUNCTION get_user_by_email(user_email TEXT)
RETURNS TABLE (
    id UUID,
    email VARCHAR(255),
    username VARCHAR(100),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    full_name VARCHAR(255),
    role VARCHAR(50),
    status VARCHAR(50),
    is_active BOOLEAN,
    email_verified BOOLEAN,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        u.id, u.email, u.username, u.first_name, u.last_name, u.full_name,
        u.role, u.status, u.is_active, u.email_verified, u.last_login_at, u.created_at
    FROM users_consolidated u
    WHERE u.email = user_email;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_user_last_login(user_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE users_consolidated 
    SET last_login_at = NOW(), updated_at = NOW()
    WHERE id = user_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION increment_failed_login_attempts(user_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE users_consolidated 
    SET 
        failed_login_attempts = failed_login_attempts + 1,
        updated_at = NOW()
    WHERE id = user_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION reset_failed_login_attempts(user_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE users_consolidated 
    SET 
        failed_login_attempts = 0,
        locked_until = NULL,
        updated_at = NOW()
    WHERE id = user_id;
END;
$$ LANGUAGE plpgsql;

-- Step 15: Create validation functions
CREATE OR REPLACE FUNCTION validate_user_data()
RETURNS TRIGGER AS $$
BEGIN
    -- Validate email format
    IF NEW.email !~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$' THEN
        RAISE EXCEPTION 'Invalid email format: %', NEW.email;
    END IF;
    
    -- Validate username if provided
    IF NEW.username IS NOT NULL AND length(NEW.username) < 3 THEN
        RAISE EXCEPTION 'Username must be at least 3 characters long';
    END IF;
    
    -- Validate role
    IF NEW.role NOT IN ('user', 'admin', 'compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'other') THEN
        RAISE EXCEPTION 'Invalid role: %', NEW.role;
    END IF;
    
    -- Validate status
    IF NEW.status NOT IN ('active', 'inactive', 'suspended', 'pending_verification') THEN
        RAISE EXCEPTION 'Invalid status: %', NEW.status;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create validation trigger
DROP TRIGGER IF EXISTS trigger_validate_user_data ON users_consolidated;
CREATE TRIGGER trigger_validate_user_data
    BEFORE INSERT OR UPDATE ON users_consolidated
    FOR EACH ROW
    EXECUTE FUNCTION validate_user_data();

-- Step 16: Create comprehensive user statistics view
CREATE OR REPLACE VIEW user_statistics AS
SELECT 
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE status = 'active') as active_users,
    COUNT(*) FILTER (WHERE status = 'inactive') as inactive_users,
    COUNT(*) FILTER (WHERE status = 'suspended') as suspended_users,
    COUNT(*) FILTER (WHERE status = 'pending_verification') as pending_users,
    COUNT(*) FILTER (WHERE email_verified = true) as verified_users,
    COUNT(*) FILTER (WHERE email_verified = false) as unverified_users,
    COUNT(*) FILTER (WHERE role = 'admin') as admin_users,
    COUNT(*) FILTER (WHERE role = 'compliance_officer') as compliance_officers,
    COUNT(*) FILTER (WHERE role = 'risk_manager') as risk_managers,
    COUNT(*) FILTER (WHERE role = 'business_analyst') as business_analysts,
    COUNT(*) FILTER (WHERE last_login_at > NOW() - INTERVAL '30 days') as active_last_30_days,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '30 days') as new_users_last_30_days
FROM users_consolidated;

-- Step 17: Create migration completion log
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000000'::UUID, -- System user ID
    'MIGRATION',
    'database',
    'users_consolidated',
    'User table consolidation migration completed successfully',
    NOW()
);

-- Commit the transaction
COMMIT;

-- Step 18: Create rollback script (for emergency use)
-- This would be saved as a separate file: 008_user_table_consolidation_rollback.sql
-- For now, we'll create a comment with the rollback steps

/*
ROLLBACK STEPS (Emergency Use Only):
1. DROP VIEW IF EXISTS profiles;
2. DROP VIEW IF EXISTS users;
3. DROP TABLE IF EXISTS users_consolidated CASCADE;
4. Restore foreign key constraints to original tables
5. Restore data from backup tables if needed

BACKUP TABLES CREATED:
- users_backup: Contains original users table data
- profiles_backup: Contains original profiles table data
*/

-- Migration completed successfully
SELECT 'User table consolidation migration completed successfully' as status;
