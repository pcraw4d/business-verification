-- =============================================================================
-- CRITICAL DATABASE SCHEMA FIXES
-- =============================================================================
-- This script fixes the critical database schema issues that are preventing
-- the classification system from working properly.

-- Fix 1: Add missing is_active column to keyword_weights table
-- =============================================================================
ALTER TABLE keyword_weights ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Update existing records to be active
UPDATE keyword_weights SET is_active = true WHERE is_active IS NULL;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_keyword_weights_active ON keyword_weights(is_active);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_active ON keyword_weights(industry_id, is_active);

-- Fix 2: Add Restaurant Industry (Priority 1)
-- =============================================================================
INSERT INTO industries (name, description, category, confidence_threshold, is_active) VALUES
('Restaurants', 'Food service establishments including fine dining, casual dining, and fast food', 'Food Service', 0.75, true),
('Fast Food', 'Quick service restaurants and fast food chains', 'Food Service', 0.80, true),
('Food & Beverage', 'Food and beverage production, distribution, and retail', 'Food Service', 0.70, true)
ON CONFLICT (name) DO NOTHING;

-- Fix 3: Add comprehensive restaurant keywords
-- =============================================================================
-- Restaurant Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, context_multiplier, usage_count, is_active)
SELECT 
  i.id,
  kw.keyword,
  kw.base_weight,
  1.0 as context_multiplier,
  0 as usage_count,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  -- Core restaurant keywords
  ('restaurant', 1.0), ('dining', 0.95), ('cuisine', 0.90), ('menu', 0.85),
  ('chef', 0.85), ('kitchen', 0.80), ('food', 0.75), ('meal', 0.75),
  
  -- Restaurant types
  ('fine dining', 0.90), ('casual dining', 0.85), ('italian', 0.80),
  ('chinese', 0.80), ('mexican', 0.80), ('american', 0.75), ('seafood', 0.75),
  ('steakhouse', 0.85), ('pizzeria', 0.85), ('cafe', 0.70), ('bistro', 0.80),
  ('grill', 0.75), ('bar', 0.70), ('pub', 0.70), ('tavern', 0.70),
  
  -- Fast food keywords
  ('fast food', 1.0), ('quick service', 0.95), ('drive thru', 0.90),
  ('takeout', 0.85), ('delivery', 0.80), ('burger', 0.85), ('pizza', 0.85),
  ('sandwich', 0.80), ('fries', 0.75), ('chain', 0.70), ('franchise', 0.70),
  
  -- Food & beverage keywords
  ('beverage', 0.90), ('wine', 0.85), ('beer', 0.80), ('spirits', 0.80),
  ('liquor', 0.80), ('alcohol', 0.75), ('cocktail', 0.75), ('brewery', 0.85),
  ('winery', 0.85), ('distillery', 0.85), ('grocery', 0.70), ('market', 0.70),
  ('store', 0.65), ('retail', 0.65)
) AS kw(keyword, base_weight)
WHERE i.name = 'Restaurants'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- Fast Food Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, context_multiplier, usage_count, is_active)
SELECT 
  i.id,
  kw.keyword,
  kw.base_weight,
  1.0 as context_multiplier,
  0 as usage_count,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  ('fast food', 1.0), ('quick service', 0.95), ('drive thru', 0.90),
  ('takeout', 0.85), ('delivery', 0.80), ('burger', 0.85), ('pizza', 0.85),
  ('sandwich', 0.80), ('fries', 0.75), ('chain', 0.70), ('franchise', 0.70),
  ('quick', 0.75), ('service', 0.70), ('restaurant', 0.65)
) AS kw(keyword, base_weight)
WHERE i.name = 'Fast Food'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- Food & Beverage Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, context_multiplier, usage_count, is_active)
SELECT 
  i.id,
  kw.keyword,
  kw.base_weight,
  1.0 as context_multiplier,
  0 as usage_count,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  ('beverage', 0.90), ('wine', 0.85), ('beer', 0.80), ('spirits', 0.80),
  ('liquor', 0.80), ('alcohol', 0.75), ('cocktail', 0.75), ('brewery', 0.85),
  ('winery', 0.85), ('distillery', 0.85), ('grocery', 0.70), ('market', 0.70),
  ('store', 0.65), ('retail', 0.65), ('food', 0.75), ('production', 0.70)
) AS kw(keyword, base_weight)
WHERE i.name = 'Food & Beverage'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- Fix 4: Add restaurant classification codes
-- =============================================================================
-- Restaurant Industry Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active)
SELECT 
  i.id,
  cc.code_type,
  cc.code,
  cc.description,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  -- MCC Codes
  ('MCC', '5812', 'Eating Places and Restaurants'),
  ('MCC', '5813', 'Drinking Places (Alcoholic Beverages)'),
  ('MCC', '5814', 'Fast Food Restaurants'),
  ('MCC', '5499', 'Miscellaneous Food Stores–Convenience Stores, Markets, Specialty Stores, and Vending Machines'),
  
  -- NAICS Codes
  ('NAICS', '722511', 'Full-Service Restaurants'),
  ('NAICS', '722513', 'Limited-Service Restaurants'),
  ('NAICS', '722514', 'Cafeterias, Grill Buffets, and Buffets'),
  ('NAICS', '722515', 'Snack and Nonalcoholic Beverage Bars'),
  ('NAICS', '312130', 'Wineries'),
  ('NAICS', '424820', 'Wine and Distilled Alcoholic Beverage Merchant Wholesalers'),
  ('NAICS', '445320', 'Beer, Wine, and Liquor Retailers'),
  
  -- SIC Codes
  ('SIC', '5812', 'Eating Places'),
  ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
  ('SIC', '5814', 'Fast Food Restaurants'),
  ('SIC', '2084', 'Wines, Brandy, and Brandy Spirits')
) AS cc(code_type, code, description)
WHERE i.name = 'Restaurants'
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- Fast Food Industry Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active)
SELECT 
  i.id,
  cc.code_type,
  cc.code,
  cc.description,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  ('MCC', '5814', 'Fast Food Restaurants'),
  ('NAICS', '722513', 'Limited-Service Restaurants'),
  ('SIC', '5814', 'Fast Food Restaurants')
) AS cc(code_type, code, description)
WHERE i.name = 'Fast Food'
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- Food & Beverage Industry Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active)
SELECT 
  i.id,
  cc.code_type,
  cc.code,
  cc.description,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  ('MCC', '5499', 'Miscellaneous Food Stores–Convenience Stores, Markets, Specialty Stores, and Vending Machines'),
  ('NAICS', '312130', 'Wineries'),
  ('NAICS', '424820', 'Wine and Distilled Alcoholic Beverage Merchant Wholesalers'),
  ('NAICS', '445320', 'Beer, Wine, and Liquor Retailers'),
  ('SIC', '2084', 'Wines, Brandy, and Brandy Spirits')
) AS cc(code_type, code, description)
WHERE i.name = 'Food & Beverage'
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- Fix 5: Verify the fixes
-- =============================================================================
-- Check that the is_active column exists and has data
SELECT 
  'keyword_weights' as table_name,
  COUNT(*) as total_records,
  COUNT(CASE WHEN is_active = true THEN 1 END) as active_records
FROM keyword_weights;

-- Check that restaurant industry was added
SELECT 
  'industries' as table_name,
  COUNT(*) as total_industries,
  COUNT(CASE WHEN name LIKE '%Restaurant%' OR name LIKE '%Food%' THEN 1 END) as food_industries
FROM industries;

-- Check that restaurant keywords were added
SELECT 
  'restaurant_keywords' as table_name,
  COUNT(*) as total_keywords
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Restaurants', 'Fast Food', 'Food & Beverage');

-- Check that restaurant classification codes were added
SELECT 
  'restaurant_codes' as table_name,
  COUNT(*) as total_codes
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name IN ('Restaurants', 'Fast Food', 'Food & Beverage');

-- =============================================================================
-- END OF CRITICAL FIXES
-- =============================================================================
-- After running this script, the classification system should be able to:
-- 1. Build the keyword index successfully (no more "is_active does not exist" error)
-- 2. Classify restaurant businesses correctly
-- 3. Return dynamic confidence scores instead of fixed 0.45
-- 4. Provide relevant keywords for restaurant businesses
-- =============================================================================
