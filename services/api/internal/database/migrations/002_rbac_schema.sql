-- Migration: 002_rbac_schema.sql
-- Description: Add RBAC (Role-Based Access Control) tables and enhance existing tables
-- Created: 2025-01-07

-- Add role assignments table for audit trail and expiration support
CREATE TABLE IF NOT EXISTS role_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('guest', 'user', 'analyst', 'manager', 'admin', 'system')),
    assigned_by UUID NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_role_assignments_user_id ON role_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_role_assignments_role ON role_assignments(role);
CREATE INDEX IF NOT EXISTS idx_role_assignments_active ON role_assignments(is_active);
CREATE INDEX IF NOT EXISTS idx_role_assignments_expires_at ON role_assignments(expires_at);
CREATE INDEX IF NOT EXISTS idx_role_assignments_assigned_by ON role_assignments(assigned_by);

-- Add unique constraint for active role assignments per user
-- Note: Using a simpler index since CURRENT_TIMESTAMP is not immutable
CREATE UNIQUE INDEX IF NOT EXISTS idx_role_assignments_user_active 
ON role_assignments(user_id) 
WHERE is_active = true;

-- Enhance API keys table with role and better permissions structure
ALTER TABLE api_keys 
ADD COLUMN IF NOT EXISTS role VARCHAR(50) CHECK (role IN ('guest', 'user', 'analyst', 'manager', 'admin', 'system'));

-- Update permissions column to be JSON text (if not already)
-- Note: In production, you'd want to migrate existing data
ALTER TABLE api_keys 
ALTER COLUMN permissions TYPE TEXT;

-- Add indexes for API key role-based queries
CREATE INDEX IF NOT EXISTS idx_api_keys_role ON api_keys(role);
CREATE INDEX IF NOT EXISTS idx_api_keys_status_expires ON api_keys(status, expires_at);

-- Add a function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers to automatically update updated_at timestamp
DROP TRIGGER IF EXISTS update_role_assignments_updated_at ON role_assignments;
CREATE TRIGGER update_role_assignments_updated_at 
    BEFORE UPDATE ON role_assignments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;
CREATE TRIGGER update_api_keys_updated_at 
    BEFORE UPDATE ON api_keys 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Ensure users table has proper role constraint and default
ALTER TABLE users 
ALTER COLUMN role SET DEFAULT 'user';

ALTER TABLE users 
ADD CONSTRAINT check_user_role 
CHECK (role IN ('guest', 'user', 'analyst', 'manager', 'admin', 'system'));

-- Create a view for active user roles (combining users.role and role_assignments)
CREATE OR REPLACE VIEW user_effective_roles AS
SELECT 
    u.id as user_id,
    u.email,
    u.username,
    COALESCE(ra.role, u.role) as effective_role,
    u.role as default_role,
    ra.role as assigned_role,
    ra.assigned_by,
    ra.assigned_at,
    ra.expires_at,
    ra.is_active as has_assignment,
    CASE 
        WHEN ra.expires_at IS NOT NULL AND ra.expires_at <= CURRENT_TIMESTAMP THEN false
        WHEN ra.is_active = false THEN false
        ELSE COALESCE(ra.is_active, true)
    END as role_is_active
FROM users u
LEFT JOIN role_assignments ra ON u.id = ra.user_id 
    AND ra.is_active = true 
    AND (ra.expires_at IS NULL OR ra.expires_at > CURRENT_TIMESTAMP);

-- Create indexes on the users table for role queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status_role ON users(status, role);

-- Add some useful functions for role management

-- Function to check if a user has a specific permission
CREATE OR REPLACE FUNCTION user_has_permission(user_id_param VARCHAR(255), permission_param VARCHAR(255))
RETURNS BOOLEAN AS $$
DECLARE
    user_role VARCHAR(50);
    role_permissions TEXT[];
BEGIN
    -- Get effective role for user
    SELECT effective_role INTO user_role 
    FROM user_effective_roles 
    WHERE user_id = user_id_param AND role_is_active = true;
    
    IF user_role IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- Define role permissions (this would ideally be in a separate table)
    CASE user_role
        WHEN 'admin' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:batch', 'classify:view', 'classify:export',
                'risk:assess', 'risk:view', 'risk:export',
                'compliance:check', 'compliance:view', 'compliance:export',
                'users:view', 'users:create', 'users:update', 'users:delete', 'users:manage_roles',
                'api_keys:view', 'api_keys:create', 'api_keys:revoke', 'api_keys:manage',
                'system:metrics', 'system:logs', 'system:config', 'system:backup',
                'audit:view', 'audit:export'
            ];
        WHEN 'manager' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:batch', 'classify:view', 'classify:export',
                'risk:assess', 'risk:view', 'risk:export',
                'compliance:check', 'compliance:view', 'compliance:export',
                'users:view', 'users:create', 'users:update', 'users:manage_roles',
                'api_keys:view', 'api_keys:create', 'api_keys:revoke',
                'system:metrics', 'audit:view', 'audit:export'
            ];
        WHEN 'analyst' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:batch', 'classify:view', 'classify:export',
                'risk:assess', 'risk:view', 'risk:export',
                'compliance:check', 'compliance:view', 'compliance:export',
                'system:metrics', 'audit:view'
            ];
        WHEN 'user' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:view', 'classify:export',
                'risk:assess', 'risk:view', 'risk:export',
                'compliance:check', 'compliance:view', 'compliance:export'
            ];
        WHEN 'guest' THEN
            role_permissions := ARRAY[
                'classify:view', 'risk:view', 'compliance:view'
            ];
        WHEN 'system' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:batch', 'risk:assess', 
                'compliance:check', 'system:metrics'
            ];
        ELSE
            role_permissions := ARRAY[]::TEXT[];
    END CASE;
    
    RETURN permission_param = ANY(role_permissions);
END;
$$ LANGUAGE plpgsql;

-- Function to get all permissions for a user
CREATE OR REPLACE FUNCTION get_user_permissions(user_id_param VARCHAR(255))
RETURNS TEXT[] AS $$
DECLARE
    user_role VARCHAR(50);
    role_permissions TEXT[];
BEGIN
    -- Get effective role for user
    SELECT effective_role INTO user_role 
    FROM user_effective_roles 
    WHERE user_id = user_id_param AND role_is_active = true;
    
    IF user_role IS NULL THEN
        RETURN ARRAY[]::TEXT[];
    END IF;
    
    -- Return permissions based on role
    CASE user_role
        WHEN 'admin' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:batch', 'classify:view', 'classify:export',
                'risk:assess', 'risk:view', 'risk:export',
                'compliance:check', 'compliance:view', 'compliance:export',
                'users:view', 'users:create', 'users:update', 'users:delete', 'users:manage_roles',
                'api_keys:view', 'api_keys:create', 'api_keys:revoke', 'api_keys:manage',
                'system:metrics', 'system:logs', 'system:config', 'system:backup',
                'audit:view', 'audit:export'
            ];
        WHEN 'manager' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:batch', 'classify:view', 'classify:export',
                'risk:assess', 'risk:view', 'risk:export',
                'compliance:check', 'compliance:view', 'compliance:export',
                'users:view', 'users:create', 'users:update', 'users:manage_roles',
                'api_keys:view', 'api_keys:create', 'api_keys:revoke',
                'system:metrics', 'audit:view', 'audit:export'
            ];
        WHEN 'analyst' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:batch', 'classify:view', 'classify:export',
                'risk:assess', 'risk:view', 'risk:export',
                'compliance:check', 'compliance:view', 'compliance:export',
                'system:metrics', 'audit:view'
            ];
        WHEN 'user' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:view', 'classify:export',
                'risk:assess', 'risk:view', 'risk:export',
                'compliance:check', 'compliance:view', 'compliance:export'
            ];
        WHEN 'guest' THEN
            role_permissions := ARRAY[
                'classify:view', 'risk:view', 'compliance:view'
            ];
        WHEN 'system' THEN
            role_permissions := ARRAY[
                'classify:business', 'classify:batch', 'risk:assess', 
                'compliance:check', 'system:metrics'
            ];
        ELSE
            role_permissions := ARRAY[]::TEXT[];
    END CASE;
    
    RETURN role_permissions;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance on new columns and constraints
CREATE INDEX IF NOT EXISTS idx_users_email_status ON users(email, status);
CREATE INDEX IF NOT EXISTS idx_users_username_status ON users(username, status);

-- Add comments for documentation
COMMENT ON TABLE role_assignments IS 'Stores role assignments with audit trail and expiration support';
COMMENT ON COLUMN role_assignments.assigned_by IS 'User ID of the person who assigned this role';
COMMENT ON COLUMN role_assignments.expires_at IS 'Optional expiration timestamp for temporary role assignments';
COMMENT ON VIEW user_effective_roles IS 'Combines user default roles with active role assignments';
COMMENT ON FUNCTION user_has_permission IS 'Checks if a user has a specific permission based on their effective role';
COMMENT ON FUNCTION get_user_permissions IS 'Returns all permissions for a user based on their effective role';
