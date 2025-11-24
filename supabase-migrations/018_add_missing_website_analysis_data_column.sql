-- Migration: Add Missing website_analysis_data Column
-- Date: January 2025
-- Purpose: Add website_analysis_data column that was missing from merchant_analytics table
-- This fixes the PGRST204 error: "Could not find the 'website_analysis_data' column"

BEGIN;

-- Add website_analysis_data column if it doesn't exist
ALTER TABLE merchant_analytics 
ADD COLUMN IF NOT EXISTS website_analysis_data JSONB DEFAULT '{}';

-- Verify the column was added
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'merchant_analytics' 
        AND column_name = 'website_analysis_data'
    ) THEN
        RAISE EXCEPTION 'Failed to add website_analysis_data column';
    ELSE
        RAISE NOTICE 'Successfully added website_analysis_data column';
    END IF;
END $$;

-- Create GIN index for JSONB queries (if not exists)
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_website_analysis_data 
ON merchant_analytics USING GIN (website_analysis_data)
WHERE website_analysis_data IS NOT NULL AND website_analysis_data != '{}'::jsonb;

-- Add comment for documentation
COMMENT ON COLUMN merchant_analytics.website_analysis_data IS 
'JSONB field storing website analysis results including performance, security, accessibility metrics';

COMMIT;

-- Verification query (run separately to confirm)
-- SELECT column_name, data_type, column_default
-- FROM information_schema.columns
-- WHERE table_name = 'merchant_analytics'
--   AND column_name IN (
--     'website_analysis_data',
--     'website_analysis_status',
--     'classification_data',
--     'classification_status'
--   )
-- ORDER BY column_name;

