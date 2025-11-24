-- Migration: Add Analytics Status Tracking
-- Date: January 2025
-- Purpose: Add status columns to merchant_analytics table for tracking classification and website analysis processing

BEGIN;

-- Add status columns to merchant_analytics table
ALTER TABLE merchant_analytics 
ADD COLUMN IF NOT EXISTS classification_status VARCHAR(50) DEFAULT 'pending' CHECK (classification_status IN ('pending', 'processing', 'completed', 'failed')),
ADD COLUMN IF NOT EXISTS classification_updated_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS website_analysis_status VARCHAR(50) DEFAULT 'pending' CHECK (website_analysis_status IN ('pending', 'processing', 'completed', 'failed', 'skipped')),
ADD COLUMN IF NOT EXISTS website_analysis_data JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS website_analysis_updated_at TIMESTAMP WITH TIME ZONE;

-- Create indexes for status queries
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_classification_status ON merchant_analytics(classification_status);
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_website_analysis_status ON merchant_analytics(website_analysis_status);
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_website_analysis_data ON merchant_analytics USING GIN (website_analysis_data);

-- Update existing records to have 'pending' status if NULL
UPDATE merchant_analytics 
SET classification_status = 'pending' 
WHERE classification_status IS NULL;

UPDATE merchant_analytics 
SET website_analysis_status = 'pending' 
WHERE website_analysis_status IS NULL;

COMMIT;

