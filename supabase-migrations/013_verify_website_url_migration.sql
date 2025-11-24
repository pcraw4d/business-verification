-- Verification Query: Check Website URL Migration Results
-- Purpose: Verify that website URLs have been successfully migrated to contact_info
-- Run this after executing migration 013_migrate_website_urls_to_contact_info.sql

-- Summary Statistics
SELECT 
    COUNT(*) as total_merchants,
    COUNT(contact_info->>'website') as merchants_with_website_in_contact_info,
    COUNT(CASE WHEN contact_website IS NOT NULL AND contact_website != '' THEN 1 END) as merchants_with_legacy_contact_website,
    COUNT(CASE WHEN website_url IS NOT NULL AND website_url != '' THEN 1 END) as merchants_with_legacy_website_url
FROM merchants;

-- Find any merchants that still have website in legacy columns but not in contact_info
-- (These would be edge cases that need manual review)
SELECT 
    id, 
    name, 
    contact_website, 
    website_url, 
    contact_info->>'website' as website_in_contact_info,
    CASE 
        WHEN contact_info->>'website' IS NULL OR contact_info->>'website' = '' THEN 'NEEDS_MIGRATION'
        ELSE 'MIGRATED'
    END as migration_status
FROM merchants
WHERE (contact_website IS NOT NULL OR website_url IS NOT NULL)
  AND (contact_info->>'website' IS NULL OR contact_info->>'website' = '')
ORDER BY created_at DESC;

-- Sample of successfully migrated merchants
SELECT 
    id,
    name,
    contact_info->>'website' as website_url,
    CASE 
        WHEN contact_website IS NOT NULL THEN 'Migrated from contact_website'
        WHEN website_url IS NOT NULL THEN 'Migrated from website_url'
        ELSE 'Already in contact_info'
    END as migration_source
FROM merchants
WHERE contact_info->>'website' IS NOT NULL 
  AND contact_info->>'website' != ''
LIMIT 10;

