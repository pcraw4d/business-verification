-- Migration: Add async risk assessment columns to risk_assessments table
-- This migration adds support for async risk assessment processing

-- Add new columns for async assessment support
ALTER TABLE risk_assessments 
ADD COLUMN IF NOT EXISTS merchant_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
ADD COLUMN IF NOT EXISTS options JSONB,
ADD COLUMN IF NOT EXISTS result JSONB,
ADD COLUMN IF NOT EXISTS progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
ADD COLUMN IF NOT EXISTS estimated_completion TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP WITH TIME ZONE;

-- Create index on merchant_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_risk_assessments_merchant_id ON risk_assessments(merchant_id);

-- Create index on status for filtering by status
CREATE INDEX IF NOT EXISTS idx_risk_assessments_status ON risk_assessments(status);

-- Create index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_risk_assessments_created_at ON risk_assessments(created_at);

-- Update existing records to have default status
UPDATE risk_assessments 
SET status = 'completed' 
WHERE status IS NULL;

-- Add comment to table
COMMENT ON TABLE risk_assessments IS 'Risk assessment records with support for async processing';
COMMENT ON COLUMN risk_assessments.status IS 'Assessment status: pending, processing, completed, failed';
COMMENT ON COLUMN risk_assessments.options IS 'Assessment options including includeHistory and includePredictions';
COMMENT ON COLUMN risk_assessments.result IS 'Final assessment result with overallScore, riskLevel, and factors';
COMMENT ON COLUMN risk_assessments.progress IS 'Progress percentage (0-100)';
COMMENT ON COLUMN risk_assessments.estimated_completion IS 'Estimated completion time for pending/processing assessments';

