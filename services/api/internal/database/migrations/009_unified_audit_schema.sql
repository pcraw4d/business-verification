-- =============================================================================
-- Migration 009: Unified Audit Schema Consolidation
-- =============================================================================
-- This migration consolidates all audit tables into a single unified schema
-- that supports comprehensive audit logging for all system operations.
-- 
-- Consolidates:
-- - audit_logs (primary schema)
-- - merchant_audit_logs 
-- - audit_logs (legacy schema)
-- - audit_logs (classification schema)
-- =============================================================================

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- 1. Create Unified Audit Logs Table
-- =============================================================================

CREATE TABLE IF NOT EXISTS unified_audit_logs (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- User and Authentication Context
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    
    -- Business Context
    merchant_id UUID REFERENCES merchants(id) ON DELETE SET NULL,
    session_id UUID REFERENCES merchant_sessions(id) ON DELETE SET NULL,
    
    -- Event Classification
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50) NOT NULL DEFAULT 'audit',
    action VARCHAR(100) NOT NULL,
    
    -- Resource Information
    resource_type VARCHAR(100),
    resource_id VARCHAR(100), -- Flexible to handle both UUID and string IDs
    table_name VARCHAR(50), -- For table-level audits
    
    -- Change Tracking
    old_values JSONB,
    new_values JSONB,
    details JSONB,
    
    -- Request Context
    request_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    
    -- Metadata and Timestamps
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_audit_log_action CHECK (action IN (
        'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'READ', 'LOGIN', 'LOGOUT', 
        'ACCESS', 'EXPORT', 'IMPORT', 'VERIFY', 'APPROVE', 'REJECT', 
        'CLASSIFY', 'ASSESS', 'SCAN', 'ANALYZE'
    )),
    CONSTRAINT chk_audit_log_event_category CHECK (event_category IN (
        'audit', 'compliance', 'security', 'business', 'system', 'user', 'merchant'
    )),
    CONSTRAINT chk_audit_log_event_type CHECK (event_type IN (
        'user_action', 'system_event', 'api_call', 'data_change', 'security_event',
        'compliance_check', 'business_operation', 'merchant_operation', 'classification',
        'risk_assessment', 'verification', 'authentication', 'authorization'
    ))
);

-- =============================================================================
-- 2. Create Performance Indexes
-- =============================================================================

-- Primary query indexes
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_user_id ON unified_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_merchant_id ON unified_audit_logs(merchant_id);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_created_at ON unified_audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_action ON unified_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_event_type ON unified_audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_resource_type ON unified_audit_logs(resource_type);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_merchant_created ON unified_audit_logs(merchant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_user_created ON unified_audit_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_action_created ON unified_audit_logs(action, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_event_type_created ON unified_audit_logs(event_type, created_at DESC);

-- JSONB indexes for metadata and details queries
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_metadata_gin ON unified_audit_logs USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_details_gin ON unified_audit_logs USING GIN(details);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_old_values_gin ON unified_audit_logs USING GIN(old_values);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_new_values_gin ON unified_audit_logs USING GIN(new_values);

-- Request tracking indexes
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_request_id ON unified_audit_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_session_id ON unified_audit_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_unified_audit_logs_ip_address ON unified_audit_logs(ip_address);

-- =============================================================================
-- 3. Create Audit Log Migration Functions
-- =============================================================================

-- Function to migrate data from existing audit_logs tables
CREATE OR REPLACE FUNCTION migrate_audit_logs_to_unified()
RETURNS TABLE(
    migrated_count INTEGER,
    error_count INTEGER,
    error_details TEXT[]
) AS $$
DECLARE
    audit_count INTEGER := 0;
    merchant_audit_count INTEGER := 0;
    legacy_audit_count INTEGER := 0;
    classification_audit_count INTEGER := 0;
    total_migrated INTEGER := 0;
    total_errors INTEGER := 0;
    error_list TEXT[] := '{}';
    rec RECORD;
BEGIN
    -- Migrate from primary audit_logs table (supabase_schema.sql)
    BEGIN
        FOR rec IN 
            SELECT 
                id,
                user_id,
                api_key_id,
                event_type,
                resource_type,
                resource_id::VARCHAR(100),
                action,
                details,
                ip_address,
                user_agent,
                created_at,
                metadata
            FROM audit_logs 
            WHERE id NOT IN (SELECT id FROM unified_audit_logs)
        LOOP
            BEGIN
                INSERT INTO unified_audit_logs (
                    id, user_id, api_key_id, event_type, event_category,
                    action, resource_type, resource_id, details,
                    ip_address, user_agent, created_at, metadata
                ) VALUES (
                    rec.id, rec.user_id, rec.api_key_id, rec.event_type, 'audit',
                    rec.action, rec.resource_type, rec.resource_id, rec.details,
                    rec.ip_address, rec.user_agent, rec.created_at, rec.metadata
                );
                audit_count := audit_count + 1;
            EXCEPTION WHEN OTHERS THEN
                total_errors := total_errors + 1;
                error_list := array_append(error_list, 
                    'Primary audit_logs: ' || rec.id || ' - ' || SQLERRM);
            END;
        END LOOP;
    EXCEPTION WHEN OTHERS THEN
        total_errors := total_errors + 1;
        error_list := array_append(error_list, 'Primary audit_logs table: ' || SQLERRM);
    END;

    -- Migrate from merchant_audit_logs table
    BEGIN
        FOR rec IN 
            SELECT 
                id,
                user_id,
                merchant_id,
                action,
                resource_type,
                resource_id,
                details,
                ip_address,
                user_agent,
                request_id,
                session_id,
                created_at
            FROM merchant_audit_logs 
            WHERE id NOT IN (SELECT id FROM unified_audit_logs)
        LOOP
            BEGIN
                INSERT INTO unified_audit_logs (
                    id, user_id, merchant_id, action, event_type, event_category,
                    resource_type, resource_id, details, ip_address, user_agent,
                    request_id, session_id, created_at
                ) VALUES (
                    rec.id, rec.user_id, rec.merchant_id, rec.action, 'merchant_operation', 'merchant',
                    rec.resource_type, rec.resource_id, rec.details, rec.ip_address, rec.user_agent,
                    rec.request_id, rec.session_id, rec.created_at
                );
                merchant_audit_count := merchant_audit_count + 1;
            EXCEPTION WHEN OTHERS THEN
                total_errors := total_errors + 1;
                error_list := array_append(error_list, 
                    'Merchant audit_logs: ' || rec.id || ' - ' || SQLERRM);
            END;
        END LOOP;
    EXCEPTION WHEN OTHERS THEN
        total_errors := total_errors + 1;
        error_list := array_append(error_list, 'Merchant audit_logs table: ' || SQLERRM);
    END;

    -- Migrate from legacy audit_logs table (001_initial_schema.sql)
    BEGIN
        FOR rec IN 
            SELECT 
                id,
                user_id,
                action,
                resource_type,
                resource_id,
                details,
                ip_address,
                user_agent,
                request_id,
                created_at
            FROM audit_logs 
            WHERE id NOT IN (SELECT id FROM unified_audit_logs)
            AND id NOT IN (SELECT id FROM merchant_audit_logs) -- Avoid duplicates
        LOOP
            BEGIN
                INSERT INTO unified_audit_logs (
                    id, user_id, action, event_type, event_category,
                    resource_type, resource_id, details, ip_address, user_agent,
                    request_id, created_at
                ) VALUES (
                    rec.id, rec.user_id, rec.action, 'user_action', 'audit',
                    rec.resource_type, rec.resource_id, rec.details, rec.ip_address, rec.user_agent,
                    rec.request_id, rec.created_at
                );
                legacy_audit_count := legacy_audit_count + 1;
            EXCEPTION WHEN OTHERS THEN
                total_errors := total_errors + 1;
                error_list := array_append(error_list, 
                    'Legacy audit_logs: ' || rec.id || ' - ' || SQLERRM);
            END;
        END LOOP;
    EXCEPTION WHEN OTHERS THEN
        total_errors := total_errors + 1;
        error_list := array_append(error_list, 'Legacy audit_logs table: ' || SQLERRM);
    END;

    -- Migrate from classification audit_logs table
    BEGIN
        FOR rec IN 
            SELECT 
                id::TEXT as id,
                user_id,
                table_name,
                record_id::TEXT as resource_id,
                action,
                old_values,
                new_values,
                timestamp as created_at
            FROM audit_logs 
            WHERE id::TEXT NOT IN (SELECT id FROM unified_audit_logs)
        LOOP
            BEGIN
                INSERT INTO unified_audit_logs (
                    id, user_id, action, event_type, event_category,
                    resource_type, resource_id, old_values, new_values, created_at
                ) VALUES (
                    uuid_generate_v4(), -- Generate new UUID for SERIAL IDs
                    rec.user_id, rec.action, 'data_change', 'system',
                    rec.table_name, rec.resource_id, rec.old_values, rec.new_values, rec.created_at
                );
                classification_audit_count := classification_audit_count + 1;
            EXCEPTION WHEN OTHERS THEN
                total_errors := total_errors + 1;
                error_list := array_append(error_list, 
                    'Classification audit_logs: ' || rec.id || ' - ' || SQLERRM);
            END;
        END LOOP;
    EXCEPTION WHEN OTHERS THEN
        total_errors := total_errors + 1;
        error_list := array_append(error_list, 'Classification audit_logs table: ' || SQLERRM);
    END;

    total_migrated := audit_count + merchant_audit_count + legacy_audit_count + classification_audit_count;

    RETURN QUERY SELECT 
        total_migrated,
        total_errors,
        error_list;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 4. Create Validation Functions
-- =============================================================================

-- Function to validate migration completeness
CREATE OR REPLACE FUNCTION validate_audit_migration()
RETURNS TABLE(
    source_table TEXT,
    source_count BIGINT,
    migrated_count BIGINT,
    is_complete BOOLEAN
) AS $$
BEGIN
    -- Check primary audit_logs
    RETURN QUERY
    SELECT 
        'audit_logs (primary)'::TEXT,
        (SELECT COUNT(*) FROM audit_logs WHERE id IN (
            SELECT id FROM audit_logs 
            EXCEPT 
            SELECT id FROM merchant_audit_logs
            EXCEPT
            SELECT id::TEXT FROM audit_logs WHERE id::TEXT ~ '^[0-9]+$'
        )),
        (SELECT COUNT(*) FROM unified_audit_logs WHERE id IN (
            SELECT id FROM audit_logs 
            EXCEPT 
            SELECT id FROM merchant_audit_logs
            EXCEPT
            SELECT id::TEXT FROM audit_logs WHERE id::TEXT ~ '^[0-9]+$'
        )),
        (SELECT COUNT(*) FROM audit_logs WHERE id IN (
            SELECT id FROM audit_logs 
            EXCEPT 
            SELECT id FROM merchant_audit_logs
            EXCEPT
            SELECT id::TEXT FROM audit_logs WHERE id::TEXT ~ '^[0-9]+$'
        )) = (SELECT COUNT(*) FROM unified_audit_logs WHERE id IN (
            SELECT id FROM audit_logs 
            EXCEPT 
            SELECT id FROM merchant_audit_logs
            EXCEPT
            SELECT id::TEXT FROM audit_logs WHERE id::TEXT ~ '^[0-9]+$'
        ));

    -- Check merchant_audit_logs
    RETURN QUERY
    SELECT 
        'merchant_audit_logs'::TEXT,
        (SELECT COUNT(*) FROM merchant_audit_logs),
        (SELECT COUNT(*) FROM unified_audit_logs WHERE id IN (
            SELECT id FROM merchant_audit_logs
        )),
        (SELECT COUNT(*) FROM merchant_audit_logs) = (SELECT COUNT(*) FROM unified_audit_logs WHERE id IN (
            SELECT id FROM merchant_audit_logs
        ));

    -- Check classification audit_logs (SERIAL IDs)
    RETURN QUERY
    SELECT 
        'audit_logs (classification)'::TEXT,
        (SELECT COUNT(*) FROM audit_logs WHERE id::TEXT ~ '^[0-9]+$'),
        (SELECT COUNT(*) FROM unified_audit_logs WHERE resource_type IN (
            SELECT DISTINCT table_name FROM audit_logs WHERE id::TEXT ~ '^[0-9]+$'
        )),
        (SELECT COUNT(*) FROM audit_logs WHERE id::TEXT ~ '^[0-9]+$') = (SELECT COUNT(*) FROM unified_audit_logs WHERE resource_type IN (
            SELECT DISTINCT table_name FROM audit_logs WHERE id::TEXT ~ '^[0-9]+$'
        ));
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 5. Create Rollback Function
-- =============================================================================

-- Function to rollback the migration if needed
CREATE OR REPLACE FUNCTION rollback_audit_migration()
RETURNS TEXT AS $$
BEGIN
    -- Drop the unified audit logs table
    DROP TABLE IF EXISTS unified_audit_logs CASCADE;
    
    -- Drop migration functions
    DROP FUNCTION IF EXISTS migrate_audit_logs_to_unified();
    DROP FUNCTION IF EXISTS validate_audit_migration();
    DROP FUNCTION IF EXISTS rollback_audit_migration();
    
    RETURN 'Audit migration rollback completed successfully';
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 6. Create Audit Triggers for Automatic Logging
-- =============================================================================

-- Function to automatically log changes to unified_audit_logs
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    old_data JSONB;
    new_data JSONB;
    user_id_val UUID;
BEGIN
    -- Get user_id from session or context (if available)
    user_id_val := current_setting('app.current_user_id', true)::UUID;
    
    -- Convert OLD and NEW records to JSONB
    IF TG_OP = 'DELETE' THEN
        old_data := to_jsonb(OLD);
        new_data := NULL;
    ELSIF TG_OP = 'UPDATE' THEN
        old_data := to_jsonb(OLD);
        new_data := to_jsonb(NEW);
    ELSIF TG_OP = 'INSERT' THEN
        old_data := NULL;
        new_data := to_jsonb(NEW);
    END IF;

    -- Insert audit log
    INSERT INTO unified_audit_logs (
        user_id,
        event_type,
        event_category,
        action,
        resource_type,
        resource_id,
        old_values,
        new_values,
        details,
        metadata
    ) VALUES (
        user_id_val,
        'data_change',
        'system',
        TG_OP,
        TG_TABLE_NAME,
        COALESCE(NEW.id::TEXT, OLD.id::TEXT),
        old_data,
        new_data,
        jsonb_build_object(
            'trigger_name', TG_NAME,
            'trigger_schema', TG_TABLE_SCHEMA,
            'trigger_table', TG_TABLE_NAME
        ),
        jsonb_build_object(
            'trigger_operation', TG_OP,
            'trigger_timestamp', NOW()
        )
    );

    -- Return appropriate record
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 7. Migration Execution Instructions
-- =============================================================================

-- To execute this migration:
-- 1. Run this script to create the unified schema
-- 2. Execute: SELECT * FROM migrate_audit_logs_to_unified();
-- 3. Validate: SELECT * FROM validate_audit_migration();
-- 4. If validation passes, proceed with application code updates
-- 5. If issues found, rollback: SELECT rollback_audit_migration();

-- =============================================================================
-- 8. Comments and Documentation
-- =============================================================================

COMMENT ON TABLE unified_audit_logs IS 'Unified audit logging table consolidating all audit operations across the system';
COMMENT ON COLUMN unified_audit_logs.id IS 'Unique identifier for the audit log entry';
COMMENT ON COLUMN unified_audit_logs.user_id IS 'ID of the user who performed the action (optional)';
COMMENT ON COLUMN unified_audit_logs.api_key_id IS 'ID of the API key used for the action (optional)';
COMMENT ON COLUMN unified_audit_logs.merchant_id IS 'ID of the merchant associated with the action (optional)';
COMMENT ON COLUMN unified_audit_logs.session_id IS 'ID of the session during which the action occurred (optional)';
COMMENT ON COLUMN unified_audit_logs.event_type IS 'Type of event (user_action, system_event, api_call, etc.)';
COMMENT ON COLUMN unified_audit_logs.event_category IS 'Category of event (audit, compliance, security, etc.)';
COMMENT ON COLUMN unified_audit_logs.action IS 'Specific action performed (INSERT, UPDATE, DELETE, etc.)';
COMMENT ON COLUMN unified_audit_logs.resource_type IS 'Type of resource affected (table name, entity type, etc.)';
COMMENT ON COLUMN unified_audit_logs.resource_id IS 'ID of the specific resource affected';
COMMENT ON COLUMN unified_audit_logs.table_name IS 'Name of the database table for table-level audits';
COMMENT ON COLUMN unified_audit_logs.old_values IS 'Previous values before the change (for UPDATE/DELETE)';
COMMENT ON COLUMN unified_audit_logs.new_values IS 'New values after the change (for INSERT/UPDATE)';
COMMENT ON COLUMN unified_audit_logs.details IS 'Additional details about the action';
COMMENT ON COLUMN unified_audit_logs.request_id IS 'ID of the HTTP request that triggered the action';
COMMENT ON COLUMN unified_audit_logs.ip_address IS 'IP address of the client that performed the action';
COMMENT ON COLUMN unified_audit_logs.user_agent IS 'User agent string of the client';
COMMENT ON COLUMN unified_audit_logs.metadata IS 'Additional metadata about the audit event';
COMMENT ON COLUMN unified_audit_logs.created_at IS 'Timestamp when the audit event occurred';

-- =============================================================================
-- Migration Complete
-- =============================================================================
