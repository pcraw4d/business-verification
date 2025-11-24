-- Migration: Consolidate Website URLs into contact_info JSONB Field
-- Date: January 2025
-- Purpose: Move website URLs from legacy columns (contact_website, website_url) 
--          into contact_info["website"] for consistency and to enable website analysis

BEGIN;

-- Step 1: Ensure contact_info exists for all merchants (set to empty object if null)
UPDATE merchants
SET contact_info = '{}'::jsonb
WHERE contact_info IS NULL;

-- Step 2: Migrate from contact_website column (if exists)
-- Only migrate if contact_info doesn't already have a website
UPDATE merchants
SET contact_info = contact_info || jsonb_build_object('website', contact_website)
WHERE contact_website IS NOT NULL 
  AND contact_website != ''
  AND (contact_info->>'website' IS NULL OR contact_info->>'website' = '');

-- Step 3: Migrate from website_url column (if exists)
-- Only migrate if contact_info doesn't already have a website
UPDATE merchants
SET contact_info = contact_info || jsonb_build_object('website', website_url)
WHERE website_url IS NOT NULL 
  AND website_url != ''
  AND (contact_info->>'website' IS NULL OR contact_info->>'website' = '');

-- Step 4: Add index for performance (if not exists)
CREATE INDEX IF NOT EXISTS idx_merchants_contact_info_website 
ON merchants USING GIN ((contact_info->'website'))
WHERE contact_info->>'website' IS NOT NULL;

-- Step 5: Add comment for documentation
COMMENT ON INDEX idx_merchants_contact_info_website IS 
'Index for querying merchants by website URL in contact_info JSONB field';

COMMIT;

