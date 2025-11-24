-- Manual Website URL Update Script
-- Use this to manually update website URLs for merchants that were created before the fix
-- 
-- IMPORTANT: Replace 'MERCHANT_ID' and 'WEBSITE_URL' with actual values
-- Example:
-- UPDATE merchants
-- SET contact_info = COALESCE(contact_info, '{}'::jsonb) || jsonb_build_object('website', 'https://www.rei.com')
-- WHERE id = 'merchant_1763485008187968486';

-- Template for updating a single merchant
UPDATE merchants
SET 
    contact_info = COALESCE(contact_info, '{}'::jsonb) || jsonb_build_object('website', 'WEBSITE_URL'),
    updated_at = NOW()
WHERE id = 'MERCHANT_ID';

-- Bulk update template (update multiple merchants at once)
-- Replace the CASE statements with actual merchant IDs and URLs
UPDATE merchants
SET 
    contact_info = COALESCE(contact_info, '{}'::jsonb) || jsonb_build_object('website', 
        CASE id
            WHEN 'merchant_1763485008187968486' THEN 'https://www.rei.com'
            WHEN 'merchant_1763483634351258512' THEN 'https://www.rei.com'
            WHEN 'merchant_1763944677412641619' THEN 'https://www.rei.com'
            -- Add more merchants here
            -- WHEN 'merchant_xxx' THEN 'https://example.com'
            ELSE contact_info->>'website'
        END
    ),
    updated_at = NOW()
WHERE id IN (
    'merchant_1763485008187968486',
    'merchant_1763483634351258512',
    'merchant_1763944677412641619'
    -- Add more merchant IDs here
);

-- Verify the update
SELECT 
    id,
    name,
    contact_info->>'website' as website_url,
    updated_at
FROM merchants
WHERE id IN (
    'merchant_1763485008187968486',
    'merchant_1763483634351258512',
    'merchant_1763944677412641619'
)
ORDER BY name, created_at DESC;

