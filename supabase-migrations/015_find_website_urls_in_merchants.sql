-- Diagnostic Query: Find where website URLs are stored for specific merchants
-- Run this to see all possible locations where website URLs might be stored

-- Check for top-level website column
SELECT 
    id,
    name,
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'website')
        THEN website
        ELSE NULL
    END as website_column,
    contact_info,
    contact_info->>'website' as website_in_contact_info,
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'website_url')
        THEN website_url
        ELSE NULL
    END as website_url_column,
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'contact_website')
        THEN contact_website
        ELSE NULL
    END as contact_website_column
FROM merchants
WHERE 
    LOWER(name) LIKE '%greene%grape%' 
    OR LOWER(name) LIKE '%rei%'
ORDER BY created_at DESC
LIMIT 20;

-- Check all columns in merchants table to see what exists
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'merchants'
  AND (column_name LIKE '%website%' OR column_name LIKE '%contact%' OR column_name LIKE '%url%')
ORDER BY column_name;

-- Check if website data might be in metadata JSONB
SELECT 
    id,
    name,
    metadata,
    metadata->>'website' as website_in_metadata,
    contact_info
FROM merchants
WHERE 
    (LOWER(name) LIKE '%greene%grape%' OR LOWER(name) LIKE '%rei%')
  AND metadata IS NOT NULL
  AND metadata != '{}'::jsonb
ORDER BY created_at DESC;

