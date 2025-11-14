-- Migration: 012_create_risk_thresholds_table.sql
-- Description: Create risk_thresholds table for persistent threshold configuration storage
-- Created: 2025-01-19
-- Dependencies: 001_initial_schema.sql

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create risk_thresholds table
CREATE TABLE IF NOT EXISTS risk_thresholds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Threshold identification
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Threshold categorization
    category VARCHAR(50) NOT NULL CHECK (category IN (
        'financial', 'operational', 'regulatory', 'reputational', 
        'cybersecurity', 'compliance', 'strategic', 'other'
    )),
    industry_code VARCHAR(20),
    business_type VARCHAR(100),
    
    -- Risk level thresholds (stored as JSONB for flexibility)
    risk_levels JSONB NOT NULL,
    
    -- Configuration flags
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    
    -- Metadata for extensibility
    metadata JSONB,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL DEFAULT 'system',
    last_modified_by VARCHAR(255) NOT NULL DEFAULT 'system',
    
    -- Constraints
    -- Ensure risk_levels is not empty (it's a JSONB object, not an array)
    CONSTRAINT risk_levels_not_empty CHECK (risk_levels::text != '{}' AND risk_levels::text != 'null')
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_risk_thresholds_category ON risk_thresholds(category);
CREATE INDEX IF NOT EXISTS idx_risk_thresholds_industry_code ON risk_thresholds(industry_code) WHERE industry_code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_risk_thresholds_is_active ON risk_thresholds(is_active);
CREATE INDEX IF NOT EXISTS idx_risk_thresholds_is_default ON risk_thresholds(is_default);
CREATE INDEX IF NOT EXISTS idx_risk_thresholds_category_active ON risk_thresholds(category, is_active);
CREATE INDEX IF NOT EXISTS idx_risk_thresholds_priority ON risk_thresholds(priority DESC);

-- Create unique constraint for default thresholds per category
CREATE UNIQUE INDEX IF NOT EXISTS idx_risk_thresholds_unique_default 
    ON risk_thresholds(category) 
    WHERE is_default = true AND is_active = true;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_risk_thresholds_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_risk_thresholds_updated_at
    BEFORE UPDATE ON risk_thresholds
    FOR EACH ROW
    EXECUTE FUNCTION update_risk_thresholds_updated_at();

-- Add comments for documentation
COMMENT ON TABLE risk_thresholds IS 'Stores risk threshold configurations for different categories, industries, and business types';
COMMENT ON COLUMN risk_thresholds.risk_levels IS 'JSONB object mapping risk levels (low, medium, high, critical) to threshold values';
COMMENT ON COLUMN risk_thresholds.category IS 'Risk category this threshold applies to';
COMMENT ON COLUMN risk_thresholds.industry_code IS 'Optional industry code for industry-specific thresholds';
COMMENT ON COLUMN risk_thresholds.is_default IS 'Whether this is the default threshold for the category';
COMMENT ON COLUMN risk_thresholds.priority IS 'Priority for matching (higher priority = better match)';

