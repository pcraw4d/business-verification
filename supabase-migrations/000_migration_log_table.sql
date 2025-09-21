-- =====================================================
-- Migration Log Table
-- Supabase Implementation
-- =====================================================
-- 
-- This script creates a migration log table to track
-- all database migrations and their status.
--
-- Author: KYB Platform Development Team
-- Date: January 19, 2025
-- Version: 1.0
-- =====================================================

-- Create migration_log table if it doesn't exist
CREATE TABLE IF NOT EXISTS migration_log (
    id SERIAL PRIMARY KEY,
    migration_name VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_migration_log_name ON migration_log(migration_name);
CREATE INDEX IF NOT EXISTS idx_migration_log_status ON migration_log(status);
CREATE INDEX IF NOT EXISTS idx_migration_log_created_at ON migration_log(created_at);

-- Insert initial migration log entry
INSERT INTO migration_log (migration_name, status, completed_at, notes) 
VALUES (
    '000_migration_log_table', 
    'completed', 
    NOW(), 
    'Created migration log table for tracking database migrations'
) ON CONFLICT (migration_name) DO NOTHING;
