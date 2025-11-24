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

-- Bulk update: Update all merchants with their respective website URLs
-- REI merchants: https://www.rei.com
-- The Greene Grape merchants: https://www.thegreenegrape.com (verify this URL)
UPDATE merchants
SET 
    contact_info = COALESCE(contact_info, '{}'::jsonb) || jsonb_build_object('website', 
        CASE 
            -- REI merchants
            WHEN id = 'merchant_1763944677412641619' THEN 'https://www.rei.com'
            WHEN id = 'merchant_1763485008187968486' THEN 'https://www.rei.com'
            WHEN id = 'merchant_1763483634351258512' THEN 'https://www.rei.com'
            -- The Greene Grape merchants
            WHEN id = 'merchant_1763614602674531538' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763613565911879706' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763563857044968893' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763484568986421177' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763482640279067685' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763446101089334931' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763445358519853000' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763444719973539182' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763442916779721163' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763442790929915478' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763441957515262613' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763440529018776390' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763440473453556339' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763439674803108078' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763438595091459560' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763438592983375065' THEN 'https://www.thegreenegrape.com'
            WHEN id = 'merchant_1763438591852141334' THEN 'https://www.thegreenegrape.com'
            ELSE contact_info->>'website'
        END
    ),
    updated_at = NOW()
WHERE id IN (
    -- REI merchants
    'merchant_1763944677412641619',
    'merchant_1763485008187968486',
    'merchant_1763483634351258512',
    -- The Greene Grape merchants
    'merchant_1763614602674531538',
    'merchant_1763613565911879706',
    'merchant_1763563857044968893',
    'merchant_1763484568986421177',
    'merchant_1763482640279067685',
    'merchant_1763446101089334931',
    'merchant_1763445358519853000',
    'merchant_1763444719973539182',
    'merchant_1763442916779721163',
    'merchant_1763442790929915478',
    'merchant_1763441957515262613',
    'merchant_1763440529018776390',
    'merchant_1763440473453556339',
    'merchant_1763439674803108078',
    'merchant_1763438595091459560',
    'merchant_1763438592983375065',
    'merchant_1763438591852141334'
);

-- Verify the update
SELECT 
    id,
    name,
    contact_info->>'website' as website_url,
    updated_at,
    CASE 
        WHEN contact_info->>'website' IS NOT NULL AND contact_info->>'website' != '' THEN '✅ Updated'
        ELSE '❌ Missing'
    END as status
FROM merchants
WHERE id IN (
    -- REI merchants
    'merchant_1763944677412641619',
    'merchant_1763485008187968486',
    'merchant_1763483634351258512',
    -- The Greene Grape merchants
    'merchant_1763614602674531538',
    'merchant_1763613565911879706',
    'merchant_1763563857044968893',
    'merchant_1763484568986421177',
    'merchant_1763482640279067685',
    'merchant_1763446101089334931',
    'merchant_1763445358519853000',
    'merchant_1763444719973539182',
    'merchant_1763442916779721163',
    'merchant_1763442790929915478',
    'merchant_1763441957515262613',
    'merchant_1763440529018776390',
    'merchant_1763440473453556339',
    'merchant_1763439674803108078',
    'merchant_1763438595091459560',
    'merchant_1763438592983375065',
    'merchant_1763438591852141334'
)
ORDER BY name, created_at DESC;

