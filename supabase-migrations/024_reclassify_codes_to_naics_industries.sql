-- Migration: Reclassify Codes to NAICS-Aligned Industries
-- This migration reclassifies existing classification codes from "General Business"
-- to the appropriate NAICS-aligned industries based on their NAICS code prefixes

-- =====================================================
-- Step 0: Fix any existing NULL industry_id values
-- =====================================================

-- Set any NULL industry_id values back to General Business
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'General Business')
WHERE industry_id IS NULL
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'General Business');

-- =====================================================
-- Step 1: Reclassify NAICS Codes Based on 2-Digit Prefix
-- =====================================================

-- NAICS 11: Agriculture, Forestry, Fishing and Hunting
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Agriculture, Forestry, Fishing and Hunting')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '11'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Agriculture, Forestry, Fishing and Hunting');

-- NAICS 21: Mining
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Mining, Quarrying, and Oil and Gas Extraction')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '21'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Mining, Quarrying, and Oil and Gas Extraction');

-- NAICS 22: Utilities
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Utilities')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '22'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Utilities');

-- NAICS 23: Construction (already exists, but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Construction')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '23'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Construction');

-- NAICS 31-33: Manufacturing (already exists, but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Manufacturing')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) IN ('31', '32', '33')
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Manufacturing');

-- NAICS 42: Wholesale Trade
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Wholesale Trade')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '42'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Wholesale Trade');

-- NAICS 44-45: Retail Trade (already exists as "Retail", but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Retail')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) IN ('44', '45')
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Retail');

-- NAICS 48-49: Transportation and Warehousing (already exists, but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Transportation')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) IN ('48', '49')
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Transportation');

-- NAICS 51: Information (already exists as "Technology", but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Technology')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '51'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Technology');

-- NAICS 52: Finance and Insurance (already exists as "Finance", but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Finance')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '52'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Finance');

-- NAICS 53: Real Estate and Rental and Leasing
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Real Estate and Rental and Leasing')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '53'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Real Estate and Rental and Leasing');

-- NAICS 54: Professional, Scientific, and Technical Services
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Professional, Scientific, and Technical Services')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '54'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Professional, Scientific, and Technical Services');

-- NAICS 55: Management of Companies and Enterprises
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Management of Companies and Enterprises')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '55'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Management of Companies and Enterprises');

-- NAICS 56: Administrative and Support Services
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Administrative and Support Services')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '56'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Administrative and Support Services');

-- NAICS 61: Educational Services (already exists as "Education", but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Education')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '61'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Education');

-- NAICS 62: Health Care and Social Assistance (already exists as "Healthcare", but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Healthcare')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '62'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Healthcare');

-- NAICS 71: Arts, Entertainment, and Recreation
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Arts, Entertainment, and Recreation')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '71'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Arts, Entertainment, and Recreation');

-- NAICS 72: Accommodation and Food Services (already exists as "Food & Beverage", but update if needed)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Food & Beverage')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '72'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Food & Beverage');

-- NAICS 81: Other Services
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Other Services')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '81'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Other Services');

-- NAICS 92: Public Administration
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Public Administration')
WHERE code_type = 'NAICS'
  AND LEFT(code, 2) = '92'
  AND industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Public Administration');

-- =====================================================
-- Step 2: Reclassify Codes Based on Description Keywords
-- =====================================================

-- For codes that don't have NAICS codes or need additional classification
-- based on description matching

-- Professional Services (NAICS 54)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Professional, Scientific, and Technical Services')
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Professional, Scientific, and Technical Services')
  AND (
    description ILIKE '%consulting%' OR
    description ILIKE '%legal%' OR
    description ILIKE '%attorney%' OR
    description ILIKE '%lawyer%' OR
    description ILIKE '%accounting%' OR
    description ILIKE '%accountant%' OR
    description ILIKE '%engineering%' OR
    description ILIKE '%architect%' OR
    description ILIKE '%scientific%' OR
    description ILIKE '%research%'
  );

-- Real Estate (NAICS 53)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Real Estate and Rental and Leasing')
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Real Estate and Rental and Leasing')
  AND (
    description ILIKE '%real estate%' OR
    description ILIKE '%property%' OR
    description ILIKE '%realtor%' OR
    description ILIKE '%broker%' OR
    description ILIKE '%leasing%' OR
    description ILIKE '%rental%'
  );

-- Wholesale Trade (NAICS 42)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Wholesale Trade')
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Wholesale Trade')
  AND (
    description ILIKE '%wholesale%' OR
    description ILIKE '%distributor%' OR
    description ILIKE '%b2b%' OR
    description ILIKE '%business to business%'
  );

-- Administrative and Support Services (NAICS 56)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Administrative and Support Services')
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Administrative and Support Services')
  AND (
    description ILIKE '%administrative%' OR
    description ILIKE '%support services%' OR
    description ILIKE '%facilities management%' OR
    description ILIKE '%employment services%' OR
    description ILIKE '%staffing%' OR
    description ILIKE '%waste management%'
  );

-- Arts, Entertainment, and Recreation (NAICS 71)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Arts, Entertainment, and Recreation')
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Arts, Entertainment, and Recreation')
  AND (
    description ILIKE '%arts%' OR
    description ILIKE '%entertainment%' OR
    description ILIKE '%recreation%' OR
    description ILIKE '%museum%' OR
    description ILIKE '%theater%' OR
    description ILIKE '%sports%' OR
    description ILIKE '%fitness%'
  );

-- Other Services (NAICS 81)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Other Services')
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Other Services')
  AND (
    description ILIKE '%repair%' OR
    description ILIKE '%maintenance%' OR
    description ILIKE '%personal services%' OR
    description ILIKE '%laundry%' OR
    description ILIKE '%funeral%' OR
    description ILIKE '%religious%'
  );

-- Public Administration (NAICS 92)
UPDATE classification_codes 
SET industry_id = (SELECT id FROM industries WHERE name = 'Public Administration')
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business')
  AND EXISTS (SELECT 1 FROM industries WHERE name = 'Public Administration')
  AND (
    description ILIKE '%government%' OR
    description ILIKE '%public administration%' OR
    description ILIKE '%federal%' OR
    description ILIKE '%state%' OR
    description ILIKE '%municipal%' OR
    description ILIKE '%regulatory%'
  );

-- =====================================================
-- Step 3: Display Reclassification Summary
-- =====================================================

SELECT 
    'Reclassification Summary' as status,
    COUNT(*) as total_codes_reclassified,
    COUNT(DISTINCT industry_id) as industries_affected
FROM classification_codes
WHERE industry_id != (SELECT id FROM industries WHERE name = 'General Business');

SELECT 
    i.name as industry,
    COUNT(cc.id) as code_count,
    COUNT(DISTINCT cc.code_type) as code_types
FROM classification_codes cc
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name != 'General Business'
GROUP BY i.name
ORDER BY code_count DESC;

-- Show remaining codes in General Business (should be minimal)
SELECT 
    'Remaining in General Business' as status,
    COUNT(*) as code_count,
    COUNT(DISTINCT code_type) as code_types,
    STRING_AGG(DISTINCT code_type, ', ') as types
FROM classification_codes
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business');

