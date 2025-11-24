-- Query to check for specific merchants by name
-- Use this to verify if merchants exist in the database

-- Search for "Greene Grape" or "REI" merchants
SELECT 
    id,
    name,
    legal_name,
    contact_info->>'website' as website_in_contact_info,
    contact_info->>'email' as email_in_contact_info,
    contact_info->>'phone' as phone_in_contact_info,
    created_at,
    updated_at,
    status,
    risk_level
FROM merchants
WHERE 
    LOWER(name) LIKE '%greene%grape%' 
    OR LOWER(name) LIKE '%rei%'
    OR LOWER(legal_name) LIKE '%greene%grape%'
    OR LOWER(legal_name) LIKE '%rei%'
ORDER BY created_at DESC;

-- Alternative: List all merchants created in the last 24 hours
SELECT 
    id,
    name,
    legal_name,
    contact_info->>'website' as website_url,
    created_at,
    status
FROM merchants
WHERE created_at >= NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC
LIMIT 50;

-- Check if merchants have contact_info at all
SELECT 
    id,
    name,
    contact_info IS NOT NULL as has_contact_info,
    contact_info,
    CASE 
        WHEN contact_info IS NULL THEN 'NULL'
        WHEN contact_info = '{}'::jsonb THEN 'EMPTY'
        ELSE 'HAS_DATA'
    END as contact_info_status
FROM merchants
WHERE 
    LOWER(name) LIKE '%greene%grape%' 
    OR LOWER(name) LIKE '%rei%'
ORDER BY created_at DESC;

