-- Migration: Add country column to risk_assessments table
-- Date: 2025-11-23
-- Issue: Column missing in production database causing analytics endpoint errors

-- Add country column if it doesn't exist
ALTER TABLE risk_assessments 
ADD COLUMN IF NOT EXISTS country VARCHAR(2);

-- Create index for better query performance (if it doesn't exist)
CREATE INDEX IF NOT EXISTS idx_risk_assessments_country 
ON risk_assessments (country);

CREATE INDEX IF NOT EXISTS idx_risk_assessments_country_created 
ON risk_assessments (country, created_at DESC);

-- Verify the column was added
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'risk_assessments' 
        AND column_name = 'country'
    ) THEN
        RAISE NOTICE 'Country column successfully added to risk_assessments table';
    ELSE
        RAISE EXCEPTION 'Failed to add country column to risk_assessments table';
    END IF;
END $$;

