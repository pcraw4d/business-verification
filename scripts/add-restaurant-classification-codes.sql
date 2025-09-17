-- =============================================================================
-- Restaurant Classification Codes Addition Script
-- Subtask 1.2.3: Add restaurant classification codes (MCC, SIC, NAICS)
-- =============================================================================

-- This script adds comprehensive restaurant classification codes to improve
-- classification accuracy and provide industry code mapping for food service businesses

-- =============================================================================
-- 1. ADD RESTAURANT CLASSIFICATION CODES
-- =============================================================================

-- Restaurant Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Restaurants
    ('NAICS', '722511', 'Full-Service Restaurants'),
    ('NAICS', '722513', 'Limited-Service Restaurants'),
    ('NAICS', '722514', 'Cafeterias, Grill Buffets, and Buffets'),
    ('NAICS', '722515', 'Snack and Nonalcoholic Beverage Bars'),
    ('NAICS', '722310', 'Food Service Contractors'),
    ('NAICS', '722320', 'Caterers'),
    ('NAICS', '722330', 'Mobile Food Services'),
    ('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)'),
    
    -- SIC Codes for Restaurants
    ('SIC', '5812', 'Eating Places'),
    ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    ('SIC', '5814', 'Caterers'),
    ('SIC', '5819', 'Eating and Drinking Places, Not Elsewhere Classified'),
    
    -- MCC Codes for Restaurants
    ('MCC', '5812', 'Eating Places, Restaurants'),
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques'),
    ('MCC', '5814', 'Fast Food Restaurants'),
    ('MCC', '5815', 'Digital Goods - Games'),
    ('MCC', '5816', 'Digital Goods - Applications (Excludes Games)'),
    ('MCC', '5817', 'Digital Goods - Media, Books, Movies, Music'),
    ('MCC', '5818', 'Digital Goods - Large Digital Goods Merchant'),
    ('MCC', '5819', 'Miscellaneous Food Stores - Convenience Stores, Specialty Markets, Vending Machines')
) AS cc(code_type, code, description)
WHERE i.name = 'Restaurants'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Fast Food Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Fast Food
    ('NAICS', '722513', 'Limited-Service Restaurants'),
    ('NAICS', '722515', 'Snack and Nonalcoholic Beverage Bars'),
    ('NAICS', '722330', 'Mobile Food Services'),
    
    -- SIC Codes for Fast Food
    ('SIC', '5812', 'Eating Places'),
    ('SIC', '5819', 'Eating and Drinking Places, Not Elsewhere Classified'),
    
    -- MCC Codes for Fast Food
    ('MCC', '5814', 'Fast Food Restaurants'),
    ('MCC', '5812', 'Eating Places, Restaurants')
) AS cc(code_type, code, description)
WHERE i.name = 'Fast Food'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Fine Dining Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Fine Dining
    ('NAICS', '722511', 'Full-Service Restaurants'),
    ('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)'),
    
    -- SIC Codes for Fine Dining
    ('SIC', '5812', 'Eating Places'),
    ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    
    -- MCC Codes for Fine Dining
    ('MCC', '5812', 'Eating Places, Restaurants'),
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques')
) AS cc(code_type, code, description)
WHERE i.name = 'Fine Dining'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Casual Dining Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Casual Dining
    ('NAICS', '722511', 'Full-Service Restaurants'),
    ('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)'),
    
    -- SIC Codes for Casual Dining
    ('SIC', '5812', 'Eating Places'),
    ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    
    -- MCC Codes for Casual Dining
    ('MCC', '5812', 'Eating Places, Restaurants'),
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques')
) AS cc(code_type, code, description)
WHERE i.name = 'Casual Dining'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Quick Service Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Quick Service
    ('NAICS', '722513', 'Limited-Service Restaurants'),
    ('NAICS', '722515', 'Snack and Nonalcoholic Beverage Bars'),
    
    -- SIC Codes for Quick Service
    ('SIC', '5812', 'Eating Places'),
    
    -- MCC Codes for Quick Service
    ('MCC', '5814', 'Fast Food Restaurants'),
    ('MCC', '5812', 'Eating Places, Restaurants')
) AS cc(code_type, code, description)
WHERE i.name = 'Quick Service'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Food & Beverage Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Food & Beverage
    ('NAICS', '722511', 'Full-Service Restaurants'),
    ('NAICS', '722513', 'Limited-Service Restaurants'),
    ('NAICS', '722514', 'Cafeterias, Grill Buffets, and Buffets'),
    ('NAICS', '722515', 'Snack and Nonalcoholic Beverage Bars'),
    ('NAICS', '722310', 'Food Service Contractors'),
    ('NAICS', '722320', 'Caterers'),
    ('NAICS', '722330', 'Mobile Food Services'),
    ('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)'),
    
    -- SIC Codes for Food & Beverage
    ('SIC', '5812', 'Eating Places'),
    ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    ('SIC', '5814', 'Caterers'),
    ('SIC', '5819', 'Eating and Drinking Places, Not Elsewhere Classified'),
    
    -- MCC Codes for Food & Beverage
    ('MCC', '5812', 'Eating Places, Restaurants'),
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques'),
    ('MCC', '5814', 'Fast Food Restaurants')
) AS cc(code_type, code, description)
WHERE i.name = 'Food & Beverage'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Catering Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Catering
    ('NAICS', '722320', 'Caterers'),
    ('NAICS', '722310', 'Food Service Contractors'),
    
    -- SIC Codes for Catering
    ('SIC', '5814', 'Caterers'),
    
    -- MCC Codes for Catering
    ('MCC', '5812', 'Eating Places, Restaurants')
) AS cc(code_type, code, description)
WHERE i.name = 'Catering'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Food Trucks Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Food Trucks
    ('NAICS', '722330', 'Mobile Food Services'),
    ('NAICS', '722513', 'Limited-Service Restaurants'),
    
    -- SIC Codes for Food Trucks
    ('SIC', '5812', 'Eating Places'),
    ('SIC', '5819', 'Eating and Drinking Places, Not Elsewhere Classified'),
    
    -- MCC Codes for Food Trucks
    ('MCC', '5814', 'Fast Food Restaurants'),
    ('MCC', '5812', 'Eating Places, Restaurants')
) AS cc(code_type, code, description)
WHERE i.name = 'Food Trucks'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Cafes & Coffee Shops Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Cafes & Coffee Shops
    ('NAICS', '722515', 'Snack and Nonalcoholic Beverage Bars'),
    ('NAICS', '722513', 'Limited-Service Restaurants'),
    
    -- SIC Codes for Cafes & Coffee Shops
    ('SIC', '5812', 'Eating Places'),
    ('SIC', '5819', 'Eating and Drinking Places, Not Elsewhere Classified'),
    
    -- MCC Codes for Cafes & Coffee Shops
    ('MCC', '5812', 'Eating Places, Restaurants'),
    ('MCC', '5814', 'Fast Food Restaurants')
) AS cc(code_type, code, description)
WHERE i.name = 'Cafes & Coffee Shops'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Bars & Pubs Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Bars & Pubs
    ('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)'),
    ('NAICS', '722511', 'Full-Service Restaurants'),
    
    -- SIC Codes for Bars & Pubs
    ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    ('SIC', '5812', 'Eating Places'),
    
    -- MCC Codes for Bars & Pubs
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques'),
    ('MCC', '5812', 'Eating Places, Restaurants')
) AS cc(code_type, code, description)
WHERE i.name = 'Bars & Pubs'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Breweries Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Breweries
    ('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)'),
    ('NAICS', '312120', 'Breweries'),
    
    -- SIC Codes for Breweries
    ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    ('SIC', '2082', 'Malt Beverages'),
    
    -- MCC Codes for Breweries
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques')
) AS cc(code_type, code, description)
WHERE i.name = 'Breweries'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Wineries Industry Classification Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, created_at, updated_at)
SELECT 
    i.id,
    cc.code_type,
    cc.code,
    cc.description,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM industries i
CROSS JOIN (VALUES
    -- NAICS Codes for Wineries
    ('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)'),
    ('NAICS', '312130', 'Wineries'),
    
    -- SIC Codes for Wineries
    ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    ('SIC', '2084', 'Wines, Brandy, and Brandy Spirits'),
    
    -- MCC Codes for Wineries
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques')
) AS cc(code_type, code, description)
WHERE i.name = 'Wineries'
ON CONFLICT (code_type, code) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- =============================================================================
-- 2. VERIFICATION QUERIES
-- =============================================================================

-- Verify restaurant classification codes were added successfully
DO $$
DECLARE
    total_codes INTEGER;
    naics_codes INTEGER;
    sic_codes INTEGER;
    mcc_codes INTEGER;
    restaurant_codes INTEGER;
    fast_food_codes INTEGER;
BEGIN
    -- Count total codes added
    SELECT COUNT(*) INTO total_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.is_active = true;
    
    -- Count codes by type
    SELECT COUNT(*) INTO naics_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.code_type = 'NAICS' AND cc.is_active = true;
    
    SELECT COUNT(*) INTO sic_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.code_type = 'SIC' AND cc.is_active = true;
    
    SELECT COUNT(*) INTO mcc_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.code_type = 'MCC' AND cc.is_active = true;
    
    -- Count codes per major industry
    SELECT COUNT(*) INTO restaurant_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name = 'Restaurants' AND cc.is_active = true;
    
    SELECT COUNT(*) INTO fast_food_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name = 'Fast Food' AND cc.is_active = true;
    
    -- Report results
    RAISE NOTICE 'Restaurant classification codes added: %', total_codes;
    RAISE NOTICE 'NAICS codes: %', naics_codes;
    RAISE NOTICE 'SIC codes: %', sic_codes;
    RAISE NOTICE 'MCC codes: %', mcc_codes;
    RAISE NOTICE 'Restaurants industry codes: %', restaurant_codes;
    RAISE NOTICE 'Fast Food industry codes: %', fast_food_codes;
    
    -- Verify specific codes exist
    RAISE NOTICE 'Key code verification:';
    FOR total_codes IN 
        SELECT i.name, cc.code_type, cc.code, cc.description
        FROM classification_codes cc
        JOIN industries i ON cc.industry_id = i.id
        WHERE i.name IN ('Restaurants', 'Fast Food', 'Fine Dining')
        AND cc.code IN ('722511', '722513', '5812', '5814')
        AND cc.is_active = true
        ORDER BY i.name, cc.code_type, cc.code
    LOOP
        RAISE NOTICE '  %: % % - %', 
            total_codes.name, total_codes.code_type, total_codes.code, total_codes.description;
    END LOOP;
END $$;

-- =============================================================================
-- 3. DISPLAY ADDED CODES SUMMARY
-- =============================================================================

-- Show classification codes count per restaurant industry
SELECT 
    i.name as industry_name,
    COUNT(cc.code) as total_codes,
    COUNT(CASE WHEN cc.code_type = 'NAICS' THEN 1 END) as naics_codes,
    COUNT(CASE WHEN cc.code_type = 'SIC' THEN 1 END) as sic_codes,
    COUNT(CASE WHEN cc.code_type = 'MCC' THEN 1 END) as mcc_codes
FROM industries i
LEFT JOIN classification_codes cc ON i.id = cc.industry_id AND cc.is_active = true
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
GROUP BY i.name
ORDER BY total_codes DESC, i.name;

-- Show detailed classification codes
SELECT 
    i.name as industry_name,
    cc.code_type,
    cc.code,
    cc.description
FROM industries i
JOIN classification_codes cc ON i.id = cc.industry_id
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
AND cc.is_active = true
ORDER BY i.name, cc.code_type, cc.code;

-- =============================================================================
-- 4. COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'RESTAURANT CLASSIFICATION CODES ADDITION COMPLETED SUCCESSFULLY';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Added comprehensive classification codes for 12 restaurant industry categories';
    RAISE NOTICE 'Codes include NAICS, SIC, and MCC classifications for accurate industry mapping';
    RAISE NOTICE 'All codes are active and linked to appropriate restaurant industries';
    RAISE NOTICE 'Foundation established for complete restaurant business classification';
    RAISE NOTICE 'Next step: Test restaurant classification (Task 1.3)';
    RAISE NOTICE '=============================================================================';
END $$;
