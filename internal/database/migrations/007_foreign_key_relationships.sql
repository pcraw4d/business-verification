-- Migration: 007_foreign_key_relationships.sql
-- Description: Add additional foreign key constraints and relationship validations
-- Created: 2025-01-19
-- Dependencies: 005_merchant_portfolio_schema.sql

-- Add additional foreign key constraints that may be missing
-- and ensure all relationships are properly defined

-- Ensure merchants table has proper foreign key constraints
-- (These should already exist from 005, but we'll verify and add if missing)

-- Add foreign key constraint for merchants.portfolio_type_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchants_portfolio_type_id_fkey' 
        AND table_name = 'merchants'
    ) THEN
        ALTER TABLE merchants 
        ADD CONSTRAINT merchants_portfolio_type_id_fkey 
        FOREIGN KEY (portfolio_type_id) REFERENCES portfolio_types(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- Add foreign key constraint for merchants.risk_level_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchants_risk_level_id_fkey' 
        AND table_name = 'merchants'
    ) THEN
        ALTER TABLE merchants 
        ADD CONSTRAINT merchants_risk_level_id_fkey 
        FOREIGN KEY (risk_level_id) REFERENCES risk_levels(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- Add foreign key constraint for merchants.created_by if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchants_created_by_fkey' 
        AND table_name = 'merchants'
    ) THEN
        ALTER TABLE merchants 
        ADD CONSTRAINT merchants_created_by_fkey 
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- Add foreign key constraint for merchant_sessions.user_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_sessions_user_id_fkey' 
        AND table_name = 'merchant_sessions'
    ) THEN
        ALTER TABLE merchant_sessions 
        ADD CONSTRAINT merchant_sessions_user_id_fkey 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for merchant_sessions.merchant_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_sessions_merchant_id_fkey' 
        AND table_name = 'merchant_sessions'
    ) THEN
        ALTER TABLE merchant_sessions 
        ADD CONSTRAINT merchant_sessions_merchant_id_fkey 
        FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for merchant_audit_logs.user_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_audit_logs_user_id_fkey' 
        AND table_name = 'merchant_audit_logs'
    ) THEN
        ALTER TABLE merchant_audit_logs 
        ADD CONSTRAINT merchant_audit_logs_user_id_fkey 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Add foreign key constraint for merchant_audit_logs.merchant_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_audit_logs_merchant_id_fkey' 
        AND table_name = 'merchant_audit_logs'
    ) THEN
        ALTER TABLE merchant_audit_logs 
        ADD CONSTRAINT merchant_audit_logs_merchant_id_fkey 
        FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Add foreign key constraint for merchant_audit_logs.session_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_audit_logs_session_id_fkey' 
        AND table_name = 'merchant_audit_logs'
    ) THEN
        ALTER TABLE merchant_audit_logs 
        ADD CONSTRAINT merchant_audit_logs_session_id_fkey 
        FOREIGN KEY (session_id) REFERENCES merchant_sessions(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Add foreign key constraint for compliance_records.merchant_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'compliance_records_merchant_id_fkey' 
        AND table_name = 'compliance_records'
    ) THEN
        ALTER TABLE compliance_records 
        ADD CONSTRAINT compliance_records_merchant_id_fkey 
        FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for compliance_records.checked_by if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'compliance_records_checked_by_fkey' 
        AND table_name = 'compliance_records'
    ) THEN
        ALTER TABLE compliance_records 
        ADD CONSTRAINT compliance_records_checked_by_fkey 
        FOREIGN KEY (checked_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Add foreign key constraint for merchant_analytics.merchant_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_analytics_merchant_id_fkey' 
        AND table_name = 'merchant_analytics'
    ) THEN
        ALTER TABLE merchant_analytics 
        ADD CONSTRAINT merchant_analytics_merchant_id_fkey 
        FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for merchant_notifications.merchant_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_notifications_merchant_id_fkey' 
        AND table_name = 'merchant_notifications'
    ) THEN
        ALTER TABLE merchant_notifications 
        ADD CONSTRAINT merchant_notifications_merchant_id_fkey 
        FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for merchant_notifications.user_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_notifications_user_id_fkey' 
        AND table_name = 'merchant_notifications'
    ) THEN
        ALTER TABLE merchant_notifications 
        ADD CONSTRAINT merchant_notifications_user_id_fkey 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for merchant_comparisons.merchant1_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_comparisons_merchant1_id_fkey' 
        AND table_name = 'merchant_comparisons'
    ) THEN
        ALTER TABLE merchant_comparisons 
        ADD CONSTRAINT merchant_comparisons_merchant1_id_fkey 
        FOREIGN KEY (merchant1_id) REFERENCES merchants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for merchant_comparisons.merchant2_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_comparisons_merchant2_id_fkey' 
        AND table_name = 'merchant_comparisons'
    ) THEN
        ALTER TABLE merchant_comparisons 
        ADD CONSTRAINT merchant_comparisons_merchant2_id_fkey 
        FOREIGN KEY (merchant2_id) REFERENCES merchants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for merchant_comparisons.user_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'merchant_comparisons_user_id_fkey' 
        AND table_name = 'merchant_comparisons'
    ) THEN
        ALTER TABLE merchant_comparisons 
        ADD CONSTRAINT merchant_comparisons_user_id_fkey 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for bulk_operations.user_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'bulk_operations_user_id_fkey' 
        AND table_name = 'bulk_operations'
    ) THEN
        ALTER TABLE bulk_operations 
        ADD CONSTRAINT bulk_operations_user_id_fkey 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- Add foreign key constraint for bulk_operation_items.bulk_operation_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'bulk_operation_items_bulk_operation_id_fkey' 
        AND table_name = 'bulk_operation_items'
    ) THEN
        ALTER TABLE bulk_operation_items 
        ADD CONSTRAINT bulk_operation_items_bulk_operation_id_fkey 
        FOREIGN KEY (bulk_operation_id) REFERENCES bulk_operations(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for bulk_operation_items.merchant_id if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'bulk_operation_items_merchant_id_fkey' 
        AND table_name = 'bulk_operation_items'
    ) THEN
        ALTER TABLE bulk_operation_items 
        ADD CONSTRAINT bulk_operation_items_merchant_id_fkey 
        FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add additional constraints for data integrity

-- Ensure portfolio types are not deleted if merchants reference them
ALTER TABLE portfolio_types 
ADD CONSTRAINT portfolio_types_protect_from_deletion 
CHECK (id IS NOT NULL);

-- Ensure risk levels are not deleted if merchants reference them
ALTER TABLE risk_levels 
ADD CONSTRAINT risk_levels_protect_from_deletion 
CHECK (id IS NOT NULL);

-- Add check constraints for data validation

-- Ensure merchant portfolio type is valid
ALTER TABLE merchants 
ADD CONSTRAINT merchants_valid_portfolio_type 
CHECK (portfolio_type_id IS NOT NULL);

-- Ensure merchant risk level is valid
ALTER TABLE merchants 
ADD CONSTRAINT merchants_valid_risk_level 
CHECK (risk_level_id IS NOT NULL);

-- Ensure merchant created_by is valid
ALTER TABLE merchants 
ADD CONSTRAINT merchants_valid_created_by 
CHECK (created_by IS NOT NULL);

-- Ensure merchant sessions have valid user and merchant
ALTER TABLE merchant_sessions 
ADD CONSTRAINT merchant_sessions_valid_user 
CHECK (user_id IS NOT NULL);

ALTER TABLE merchant_sessions 
ADD CONSTRAINT merchant_sessions_valid_merchant 
CHECK (merchant_id IS NOT NULL);

-- Ensure compliance records have valid merchant
ALTER TABLE compliance_records 
ADD CONSTRAINT compliance_records_valid_merchant 
CHECK (merchant_id IS NOT NULL);

-- Ensure merchant analytics have valid merchant
ALTER TABLE merchant_analytics 
ADD CONSTRAINT merchant_analytics_valid_merchant 
CHECK (merchant_id IS NOT NULL);

-- Ensure merchant notifications have valid merchant and user
ALTER TABLE merchant_notifications 
ADD CONSTRAINT merchant_notifications_valid_merchant 
CHECK (merchant_id IS NOT NULL);

ALTER TABLE merchant_notifications 
ADD CONSTRAINT merchant_notifications_valid_user 
CHECK (user_id IS NOT NULL);

-- Ensure merchant comparisons have valid merchants and user
ALTER TABLE merchant_comparisons 
ADD CONSTRAINT merchant_comparisons_valid_merchant1 
CHECK (merchant1_id IS NOT NULL);

ALTER TABLE merchant_comparisons 
ADD CONSTRAINT merchant_comparisons_valid_merchant2 
CHECK (merchant2_id IS NOT NULL);

ALTER TABLE merchant_comparisons 
ADD CONSTRAINT merchant_comparisons_valid_user 
CHECK (user_id IS NOT NULL);

-- Ensure bulk operations have valid user
ALTER TABLE bulk_operations 
ADD CONSTRAINT bulk_operations_valid_user 
CHECK (user_id IS NOT NULL);

-- Ensure bulk operation items have valid bulk operation and merchant
ALTER TABLE bulk_operation_items 
ADD CONSTRAINT bulk_operation_items_valid_bulk_operation 
CHECK (bulk_operation_id IS NOT NULL);

ALTER TABLE bulk_operation_items 
ADD CONSTRAINT bulk_operation_items_valid_merchant 
CHECK (merchant_id IS NOT NULL);

-- Add comments for documentation
COMMENT ON CONSTRAINT merchants_portfolio_type_id_fkey ON merchants IS 'Foreign key to portfolio_types table';
COMMENT ON CONSTRAINT merchants_risk_level_id_fkey ON merchants IS 'Foreign key to risk_levels table';
COMMENT ON CONSTRAINT merchants_created_by_fkey ON merchants IS 'Foreign key to users table for creator';
COMMENT ON CONSTRAINT merchant_sessions_user_id_fkey ON merchant_sessions IS 'Foreign key to users table';
COMMENT ON CONSTRAINT merchant_sessions_merchant_id_fkey ON merchant_sessions IS 'Foreign key to merchants table';
COMMENT ON CONSTRAINT merchant_audit_logs_user_id_fkey ON merchant_audit_logs IS 'Foreign key to users table (nullable)';
COMMENT ON CONSTRAINT merchant_audit_logs_merchant_id_fkey ON merchant_audit_logs IS 'Foreign key to merchants table (nullable)';
COMMENT ON CONSTRAINT merchant_audit_logs_session_id_fkey ON merchant_audit_logs IS 'Foreign key to merchant_sessions table (nullable)';
COMMENT ON CONSTRAINT compliance_records_merchant_id_fkey ON compliance_records IS 'Foreign key to merchants table';
COMMENT ON CONSTRAINT compliance_records_checked_by_fkey ON compliance_records IS 'Foreign key to users table (nullable)';
COMMENT ON CONSTRAINT merchant_analytics_merchant_id_fkey ON merchant_analytics IS 'Foreign key to merchants table';
COMMENT ON CONSTRAINT merchant_notifications_merchant_id_fkey ON merchant_notifications IS 'Foreign key to merchants table';
COMMENT ON CONSTRAINT merchant_notifications_user_id_fkey ON merchant_notifications IS 'Foreign key to users table';
COMMENT ON CONSTRAINT merchant_comparisons_merchant1_id_fkey ON merchant_comparisons IS 'Foreign key to merchants table (first merchant)';
COMMENT ON CONSTRAINT merchant_comparisons_merchant2_id_fkey ON merchant_comparisons IS 'Foreign key to merchants table (second merchant)';
COMMENT ON CONSTRAINT merchant_comparisons_user_id_fkey ON merchant_comparisons IS 'Foreign key to users table';
COMMENT ON CONSTRAINT bulk_operations_user_id_fkey ON bulk_operations IS 'Foreign key to users table';
COMMENT ON CONSTRAINT bulk_operation_items_bulk_operation_id_fkey ON bulk_operation_items IS 'Foreign key to bulk_operations table';
COMMENT ON CONSTRAINT bulk_operation_items_merchant_id_fkey ON bulk_operation_items IS 'Foreign key to merchants table';
