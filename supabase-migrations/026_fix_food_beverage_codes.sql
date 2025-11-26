-- =====================================================
-- Migration: Fix Food & Beverage Classification Codes
-- Purpose: Disable incorrect hotel NAICS codes and add correct Food & Beverage codes
-- Industry: Food & Beverage (industry_id=10)
-- =====================================================

-- Step 1: Ensure is_active column exists in classification_codes table
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'classification_codes' 
        AND column_name = 'is_active'
    ) THEN
        ALTER TABLE classification_codes ADD COLUMN is_active BOOLEAN DEFAULT true;
        CREATE INDEX IF NOT EXISTS idx_classification_codes_active ON classification_codes(is_active);
    END IF;
END $$;

-- Step 2: Disable incorrect hotel NAICS codes for Food & Beverage industry
-- These codes (721110, 721120, 721191) are hotel codes, not food/beverage codes
UPDATE classification_codes 
SET is_active = false,
    updated_at = NOW()
WHERE industry_id = (SELECT id FROM industries WHERE name = 'Food & Beverage')
  AND code_type = 'NAICS' 
  AND code IN ('721110', '721120', '721191');

-- Step 3: Add correct Food & Beverage NAICS codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, is_primary, confidence, created_at, updated_at)
SELECT 
    (SELECT id FROM industries WHERE name = 'Food & Beverage'),
    code_type,
    code,
    description,
    true as is_active,
    is_primary,
    confidence,
    NOW() as created_at,
    NOW() as updated_at
FROM (VALUES 
    -- Primary NAICS codes for Food & Beverage
    ('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)', true, 0.95),
    ('NAICS', '445310', 'Beer, Wine, and Liquor Stores', true, 0.95),
    ('NAICS', '722511', 'Full-Service Restaurants', false, 0.90),
    ('NAICS', '722513', 'Limited-Service Restaurants', false, 0.90),
    ('NAICS', '445110', 'Supermarkets and Grocery Stores', false, 0.85),
    ('NAICS', '311111', 'Dog and Cat Food Manufacturing', false, 0.75),
    -- Additional Food & Beverage NAICS codes
    ('NAICS', '722514', 'Cafeterias, Grill Buffets, and Buffets', false, 0.85),
    ('NAICS', '722515', 'Snack and Nonalcoholic Beverage Bars', false, 0.85),
    ('NAICS', '722310', 'Food Service Contractors', false, 0.80),
    ('NAICS', '722320', 'Caterers', false, 0.85),
    ('NAICS', '722330', 'Mobile Food Services', false, 0.85),
    ('NAICS', '445120', 'Convenience Stores', false, 0.80),
    ('NAICS', '445210', 'Meat Markets', false, 0.75),
    ('NAICS', '445220', 'Fish and Seafood Markets', false, 0.75),
    ('NAICS', '445230', 'Fruit and Vegetable Markets', false, 0.75),
    ('NAICS', '445291', 'Baked Goods Stores', false, 0.80),
    ('NAICS', '445292', 'Confectionery and Nut Stores', false, 0.75),
    ('NAICS', '445299', 'All Other Specialty Food Stores', false, 0.70),
    ('NAICS', '312111', 'Soft Drink Manufacturing', false, 0.80),
    ('NAICS', '312112', 'Bottled Water Manufacturing', false, 0.75),
    ('NAICS', '312113', 'Ice Manufacturing', false, 0.70),
    ('NAICS', '312130', 'Wineries', false, 0.85),
    ('NAICS', '312140', 'Distilleries', false, 0.85),
    ('NAICS', '424820', 'Wine and Distilled Alcoholic Beverage Merchant Wholesalers', false, 0.80)
) AS new_codes(code_type, code, description, is_primary, confidence)
ON CONFLICT (code_type, code) DO UPDATE 
SET 
    industry_id = EXCLUDED.industry_id,
    description = EXCLUDED.description,
    is_active = true,
    is_primary = EXCLUDED.is_primary,
    confidence = EXCLUDED.confidence,
    updated_at = NOW();

-- Step 4: Add missing SIC codes for Food & Beverage
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, is_primary, confidence, created_at, updated_at)
SELECT 
    (SELECT id FROM industries WHERE name = 'Food & Beverage'),
    code_type,
    code,
    description,
    true as is_active,
    is_primary,
    confidence,
    NOW() as created_at,
    NOW() as updated_at
FROM (VALUES 
    -- Primary SIC codes for Food & Beverage
    ('SIC', '5812', 'Eating Places', true, 0.95),
    ('SIC', '5921', 'Package Stores (Beer, Wine, Liquor)', true, 0.95),
    ('SIC', '5499', 'Miscellaneous Food Stores', false, 0.85),
    ('SIC', '5813', 'Drinking Places', false, 0.90),
    -- Additional SIC codes for Food & Beverage
    ('SIC', '5814', 'Caterers', false, 0.85),
    ('SIC', '5819', 'Eating and Drinking Places, Not Elsewhere Classified', false, 0.80),
    ('SIC', '5411', 'Grocery Stores', false, 0.85),
    ('SIC', '5421', 'Meat and Fish (Seafood) Markets, Including Freezer Provisioners', false, 0.75),
    ('SIC', '5431', 'Fruit and Vegetable Markets', false, 0.75),
    ('SIC', '5441', 'Candy, Nut, and Confectionery Stores', false, 0.75),
    ('SIC', '5451', 'Dairy Products Stores', false, 0.75),
    ('SIC', '5461', 'Retail Bakeries', false, 0.80),
    ('SIC', '5499', 'Miscellaneous Food Stores', false, 0.70),
    ('SIC', '2084', 'Wines, Brandy, and Brandy Spirits', false, 0.85),
    ('SIC', '2085', 'Distilled and Blended Liquors', false, 0.85),
    ('SIC', '2086', 'Bottled and Canned Soft Drinks and Carbonated Waters', false, 0.80),
    ('SIC', '2087', 'Flavoring Syrup and Concentrate Manufacturing', false, 0.75),
    ('SIC', '2091', 'Canned and Cured Fish and Seafoods', false, 0.70),
    ('SIC', '2092', 'Prepared Fresh or Frozen Fish and Seafoods', false, 0.70),
    ('SIC', '2095', 'Roasted Coffee', false, 0.75),
    ('SIC', '2096', 'Potato Chips, Corn Chips, and Similar Snacks', false, 0.70),
    ('SIC', '2097', 'Manufactured Ice', false, 0.70),
    ('SIC', '2098', 'Macaroni, Spaghetti, Vermicelli, and Noodles', false, 0.70),
    ('SIC', '2099', 'Food Preparations, Not Elsewhere Classified', false, 0.70)
) AS new_codes(code_type, code, description, is_primary, confidence)
ON CONFLICT (code_type, code) DO UPDATE 
SET 
    industry_id = EXCLUDED.industry_id,
    description = EXCLUDED.description,
    is_active = true,
    is_primary = EXCLUDED.is_primary,
    confidence = EXCLUDED.confidence,
    updated_at = NOW();

-- Step 5: Add/Update MCC codes for Food & Beverage
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, is_primary, confidence, created_at, updated_at)
SELECT 
    (SELECT id FROM industries WHERE name = 'Food & Beverage'),
    code_type,
    code,
    description,
    true as is_active,
    is_primary,
    confidence,
    NOW() as created_at,
    NOW() as updated_at
FROM (VALUES 
    -- Primary MCC codes for Food & Beverage
    ('MCC', '5812', 'Eating Places, Restaurants', true, 0.95),
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques', true, 0.95),
    ('MCC', '5814', 'Fast Food Restaurants', false, 0.90),
    ('MCC', '5921', 'Package Stores - Beer, Wine, and Liquor', true, 0.95),
    -- Additional MCC codes for Food & Beverage
    ('MCC', '5411', 'Grocery Stores, Supermarkets', false, 0.85),
    ('MCC', '5422', 'Meat Provisioners - Freezer and Locker', false, 0.75),
    ('MCC', '5441', 'Candy Stores, Nut Stores, and Confectionery Stores', false, 0.75),
    ('MCC', '5451', 'Dairy Products Stores', false, 0.75),
    ('MCC', '5462', 'Bakeries', false, 0.80),
    ('MCC', '5499', 'Miscellaneous Food Stores - Convenience Stores, Markets, Specialty Stores, and Vending Machines', false, 0.80),
    ('MCC', '5541', 'Service Stations (with or without ancillary services)', false, 0.70),
    ('MCC', '5542', 'Automated Fuel Dispensers', false, 0.65)
) AS new_codes(code_type, code, description, is_primary, confidence)
ON CONFLICT (code_type, code) DO UPDATE 
SET 
    industry_id = EXCLUDED.industry_id,
    description = EXCLUDED.description,
    is_active = true,
    is_primary = EXCLUDED.is_primary,
    confidence = EXCLUDED.confidence,
    updated_at = NOW();

-- Step 6: Verification query
DO $$
DECLARE
    disabled_count INTEGER;
    active_naics_count INTEGER;
    active_sic_count INTEGER;
    active_mcc_count INTEGER;
BEGIN
    -- Count disabled hotel codes
    SELECT COUNT(*) INTO disabled_count
    FROM classification_codes
    WHERE industry_id = (SELECT id FROM industries WHERE name = 'Food & Beverage')
      AND code_type = 'NAICS'
      AND code IN ('721110', '721120', '721191')
      AND is_active = false;
    
    -- Count active codes by type
    SELECT COUNT(*) INTO active_naics_count
    FROM classification_codes
    WHERE industry_id = (SELECT id FROM industries WHERE name = 'Food & Beverage')
      AND code_type = 'NAICS'
      AND is_active = true;
    
    SELECT COUNT(*) INTO active_sic_count
    FROM classification_codes
    WHERE industry_id = (SELECT id FROM industries WHERE name = 'Food & Beverage')
      AND code_type = 'SIC'
      AND is_active = true;
    
    SELECT COUNT(*) INTO active_mcc_count
    FROM classification_codes
    WHERE industry_id = (SELECT id FROM industries WHERE name = 'Food & Beverage')
      AND code_type = 'MCC'
      AND is_active = true;
    
    RAISE NOTICE 'Migration completed successfully!';
    RAISE NOTICE 'Disabled hotel codes: %', disabled_count;
    RAISE NOTICE 'Active NAICS codes: %', active_naics_count;
    RAISE NOTICE 'Active SIC codes: %', active_sic_count;
    RAISE NOTICE 'Active MCC codes: %', active_mcc_count;
END $$;

-- =====================================================
-- Migration Complete
-- =====================================================

