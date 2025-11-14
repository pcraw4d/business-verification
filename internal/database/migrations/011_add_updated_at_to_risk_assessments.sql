-- Migration: Add updated_at column to risk_assessments table
-- This migration ensures the updated_at column exists for proper tracking of assessment updates

-- Add updated_at column if it doesn't exist
ALTER TABLE risk_assessments 
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Update existing records to have updated_at set to created_at if it's null
UPDATE risk_assessments 
SET updated_at = created_at 
WHERE updated_at IS NULL;

-- Add trigger to automatically update updated_at on row updates
CREATE OR REPLACE FUNCTION update_risk_assessments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop trigger if it exists, then create it
DROP TRIGGER IF EXISTS trigger_update_risk_assessments_updated_at ON risk_assessments;
CREATE TRIGGER trigger_update_risk_assessments_updated_at
    BEFORE UPDATE ON risk_assessments
    FOR EACH ROW
    EXECUTE FUNCTION update_risk_assessments_updated_at();

-- Add comment
COMMENT ON COLUMN risk_assessments.updated_at IS 'Timestamp of last update to the assessment record';

