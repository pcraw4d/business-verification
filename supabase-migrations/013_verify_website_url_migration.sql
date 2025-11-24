-- Verification Query: Check Website URL Migration Results
-- Purpose: Verify that website URLs have been successfully migrated to contact_info
-- Run this after executing migration 013_migrate_website_urls_to_contact_info.sql

-- Summary Statistics
-- Note: This query safely handles cases where legacy columns may not exist
-- Uses dynamic SQL to avoid column reference errors

-- First, check which columns exist
DO $$
DECLARE
    has_contact_website BOOLEAN;
    has_website_url BOOLEAN;
    contact_website_count INTEGER := 0;
    website_url_count INTEGER := 0;
BEGIN
    -- Check if columns exist
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'merchants' AND column_name = 'contact_website'
    ) INTO has_contact_website;
    
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'merchants' AND column_name = 'website_url'
    ) INTO has_website_url;
    
    -- Count from legacy columns only if they exist
    IF has_contact_website THEN
        EXECUTE 'SELECT COUNT(*) FROM merchants WHERE contact_website IS NOT NULL AND contact_website != '''''
        INTO contact_website_count;
    END IF;
    
    IF has_website_url THEN
        EXECUTE 'SELECT COUNT(*) FROM merchants WHERE website_url IS NOT NULL AND website_url != '''''
        INTO website_url_count;
    END IF;
    
    -- Display results
    RAISE NOTICE '=== Migration Verification Results ===';
    RAISE NOTICE 'contact_website column exists: %', has_contact_website;
    RAISE NOTICE 'website_url column exists: %', has_website_url;
    RAISE NOTICE 'Merchants with website in contact_website: %', contact_website_count;
    RAISE NOTICE 'Merchants with website in website_url: %', website_url_count;
END $$;

-- Summary Statistics (safe query that doesn't reference legacy columns)
SELECT 
    COUNT(*) as total_merchants,
    COUNT(contact_info->>'website') as merchants_with_website_in_contact_info,
    COUNT(CASE WHEN contact_info->>'website' IS NOT NULL AND contact_info->>'website' != '' THEN 1 END) as merchants_with_website_count
FROM merchants;

-- Find any merchants that still have website in legacy columns but not in contact_info
-- (These would be edge cases that need manual review)
-- Note: This query only runs if legacy columns exist
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'contact_website')
       OR EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'website_url')
    THEN
        RAISE NOTICE 'Legacy columns exist, checking for unmigrated data...';
    ELSE
        RAISE NOTICE 'No legacy columns found, all website URLs should be in contact_info';
    END IF;
END $$;

-- Query for unmigrated merchants (only if columns exist)
-- Run this manually if legacy columns exist:
-- SELECT 
--     id, 
--     name, 
--     CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'contact_website') 
--          THEN contact_website ELSE NULL END as contact_website, 
--     CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'website_url') 
--          THEN website_url ELSE NULL END as website_url, 
--     contact_info->>'website' as website_in_contact_info,
--     CASE 
--         WHEN contact_info->>'website' IS NULL OR contact_info->>'website' = '' THEN 'NEEDS_MIGRATION'
--         ELSE 'MIGRATED'
--     END as migration_status
-- FROM merchants
-- WHERE (
--     (EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'contact_website') 
--      AND contact_website IS NOT NULL)
--     OR
--     (EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'website_url') 
--      AND website_url IS NOT NULL)
-- )
--   AND (contact_info->>'website' IS NULL OR contact_info->>'website' = '')
-- ORDER BY created_at DESC;

-- Sample of successfully migrated merchants
SELECT 
    id,
    name,
    contact_info->>'website' as website_url,
    'In contact_info' as migration_source
FROM merchants
WHERE contact_info->>'website' IS NOT NULL 
  AND contact_info->>'website' != ''
LIMIT 10;

