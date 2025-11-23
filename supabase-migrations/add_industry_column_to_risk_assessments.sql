-- Migration: Add industry column to risk_assessments table
-- Date: 2025-11-23
-- Issue: Column missing in production database causing ERROR #5

-- Add industry column if it doesn't exist
ALTER TABLE risk_assessments 
ADD COLUMN IF NOT EXISTS industry VARCHAR(100);

-- Create indexes for better query performance (if they don't exist)
CREATE INDEX IF NOT EXISTS idx_risk_assessments_risk_industry 
ON risk_assessments (risk_level, industry);

CREATE INDEX IF NOT EXISTS idx_risk_assessments_industry_created 
ON risk_assessments (industry, created_at DESC);

-- Verify the column was added
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'risk_assessments' 
        AND column_name = 'industry'
    ) THEN
        RAISE NOTICE 'Industry column successfully added to risk_assessments table';
    ELSE
        RAISE EXCEPTION 'Failed to add industry column to risk_assessments table';
    END IF;
END $$;

