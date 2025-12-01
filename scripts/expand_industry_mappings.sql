-- =====================================================
-- Expand Industry Mapping Coverage - Phase 2, Task 2.5
-- Purpose: Increase industry mapping coverage from 59.33% to 80%+ (120+ codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 2, Task 2.5
-- =====================================================
-- 
-- This script adds industry mapping data for 120+ codes, establishing
-- primary and secondary industry classifications for better categorization.
-- Target: 120+ codes with industry mappings (80%+ coverage)
-- Accuracy Target: > 90% accuracy for industry mappings
-- =====================================================

-- =====================================================
-- Part 1: Technology Industry Mappings
-- =====================================================

-- Technology - Software Development
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Technology',
    'secondary_industries', ARRAY['Software', 'IT Services', 'Computer Services', 'Software Development'],
    'industry_category', 'Information Technology'
)
WHERE code_type = 'NAICS' AND code IN ('541511', '541512', '541519')
   OR code_type = 'SIC' AND code IN ('7371', '7372', '7373', '7374', '7375', '7376')
   OR code_type = 'MCC' AND code IN ('5734', '5045', '5735');

-- Technology - Data Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Technology',
    'secondary_industries', ARRAY['Data Services', 'Cloud Services', 'Hosting', 'Internet Services'],
    'industry_category', 'Information Technology'
)
WHERE code_type = 'NAICS' AND code IN ('518210', '518310', '519130');

-- Technology - Engineering Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Technology',
    'secondary_industries', ARRAY['Engineering', 'Consulting', 'Professional Services'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'NAICS' AND code IN ('541330', '541310', '541370')
   OR code_type = 'SIC' AND code IN ('8711', '8712', '8713');

-- =====================================================
-- Part 2: Financial Services Industry Mappings
-- =====================================================

-- Financial Services - Banking
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Financial Services',
    'secondary_industries', ARRAY['Banking', 'Commercial Banking', 'Savings Institutions', 'Credit Services'],
    'industry_category', 'Finance'
)
WHERE code_type = 'NAICS' AND code IN ('522110', '522120', '522210', '522291', '522292')
   OR code_type = 'SIC' AND code IN ('6021', '6022', '6029', '6035', '6036')
   OR code_type = 'MCC' AND code IN ('6010', '6011', '6012');

-- Financial Services - Investment Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Financial Services',
    'secondary_industries', ARRAY['Investment Services', 'Securities', 'Trading', 'Brokerage'],
    'industry_category', 'Finance'
)
WHERE code_type = 'NAICS' AND code IN ('523110', '523120', '523130', '523210')
   OR code_type = 'SIC' AND code IN ('6211', '6221', '6231');

-- =====================================================
-- Part 3: Healthcare Industry Mappings
-- =====================================================

-- Healthcare - Medical Practices
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Healthcare',
    'secondary_industries', ARRAY['Medical Services', 'Physician Services', 'Healthcare Providers'],
    'industry_category', 'Healthcare'
)
WHERE code_type = 'NAICS' AND code IN ('621111', '621112', '621210', '621310', '621320', '621391', '621399')
   OR code_type = 'SIC' AND code IN ('8011', '8021', '8041', '8042', '8043', '8049')
   OR code_type = 'MCC' AND code IN ('8011', '8021', '8041', '8042', '8043', '8049');

-- Healthcare - Hospitals
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Healthcare',
    'secondary_industries', ARRAY['Hospital Services', 'Medical Facilities', 'Healthcare Facilities'],
    'industry_category', 'Healthcare'
)
WHERE code_type = 'NAICS' AND code IN ('622110', '622210', '622310')
   OR code_type = 'SIC' AND code IN ('8062', '8063', '8069')
   OR code_type = 'MCC' AND code IN ('8062', '8050');

-- Healthcare - Medical Laboratories
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Healthcare',
    'secondary_industries', ARRAY['Medical Laboratories', 'Diagnostic Services', 'Testing Services'],
    'industry_category', 'Healthcare'
)
WHERE code_type = 'NAICS' AND code IN ('621511', '621512')
   OR code_type = 'SIC' AND code IN ('8071', '8072')
   OR code_type = 'MCC' AND code IN ('8071');

-- Healthcare - Other Health Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Healthcare',
    'secondary_industries', ARRAY['Health Services', 'Allied Health', 'Health Support Services'],
    'industry_category', 'Healthcare'
)
WHERE code_type = 'NAICS' AND code IN ('621410', '621420', '621999')
   OR code_type = 'SIC' AND code = '8099'
   OR code_type = 'MCC' AND code = '8099';

-- =====================================================
-- Part 4: Retail Industry Mappings
-- =====================================================

-- Retail - Department Stores
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Department Stores', 'Retail Stores', 'General Merchandise'],
    'industry_category', 'Retail'
)
WHERE code_type = 'NAICS' AND code IN ('452111', '452112', '452210')
   OR code_type = 'SIC' AND code IN ('5311', '5331')
   OR code_type = 'MCC' AND code IN ('5310', '5311', '5331');

-- Retail - Grocery Stores
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Grocery Stores', 'Food Retail', 'Supermarkets'],
    'industry_category', 'Retail'
)
WHERE code_type = 'NAICS' AND code IN ('445110', '445120')
   OR code_type = 'SIC' AND code = '5411'
   OR code_type = 'MCC' AND code = '5411';

-- Retail - Specialty Food Stores
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Specialty Food', 'Confectionery', 'Food Retail'],
    'industry_category', 'Retail'
)
WHERE code_type = 'NAICS' AND code = '445292'
   OR code_type = 'SIC' AND code = '5441'
   OR code_type = 'MCC' AND code = '5441';

-- =====================================================
-- Part 5: Food & Beverage Industry Mappings
-- =====================================================

-- Food & Beverage - Restaurants
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Food & Beverage',
    'secondary_industries', ARRAY['Restaurants', 'Food Services', 'Dining Services'],
    'industry_category', 'Food Service'
)
WHERE code_type = 'NAICS' AND code IN ('722511', '722513', '722515')
   OR code_type = 'SIC' AND code = '5812'
   OR code_type = 'MCC' AND code IN ('5811', '5812', '5814');

-- Food & Beverage - Drinking Places
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Food & Beverage',
    'secondary_industries', ARRAY['Beverage Services', 'Bars', 'Nightlife'],
    'industry_category', 'Food Service'
)
WHERE code_type = 'NAICS' AND code = '722410'
   OR code_type = 'SIC' AND code = '5813'
   OR code_type = 'MCC' AND code IN ('5813', '5921');

-- =====================================================
-- Part 6: Construction Industry Mappings
-- =====================================================

-- Construction - Residential Construction
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Construction',
    'secondary_industries', ARRAY['Residential Construction', 'Home Building', 'Housing Construction'],
    'industry_category', 'Construction'
)
WHERE code_type = 'NAICS' AND code IN ('236115', '236116', '236117')
   OR code_type = 'SIC' AND code IN ('1521', '1522', '1531')
   OR code_type = 'MCC' AND code IN ('1521', '1522');

-- Construction - Commercial Construction
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Construction',
    'secondary_industries', ARRAY['Commercial Construction', 'Industrial Construction', 'Building Construction'],
    'industry_category', 'Construction'
)
WHERE code_type = 'NAICS' AND code IN ('236210', '236220')
   OR code_type = 'SIC' AND code IN ('1541', '1542')
   OR code_type = 'MCC' AND code = '1521';

-- Construction - Infrastructure Construction
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Construction',
    'secondary_industries', ARRAY['Infrastructure', 'Highway Construction', 'Public Works'],
    'industry_category', 'Construction'
)
WHERE code_type = 'NAICS' AND code IN ('237310', '237110', '237120', '237130')
   OR code_type = 'SIC' AND code IN ('1611', '1622', '1623')
   OR code_type = 'MCC' AND code = '1521';

-- =====================================================
-- Part 7: Real Estate Industry Mappings
-- =====================================================

-- Real Estate - Property Management
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Real Estate',
    'secondary_industries', ARRAY['Property Management', 'Real Estate Services', 'Property Leasing'],
    'industry_category', 'Real Estate'
)
WHERE code_type = 'NAICS' AND code IN ('531120', '531110')
   OR code_type = 'SIC' AND code IN ('6512', '6513', '6514')
   OR code_type = 'MCC' AND code = '6512';

-- Real Estate - Real Estate Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Real Estate',
    'secondary_industries', ARRAY['Real Estate Services', 'Real Estate Agents', 'Brokerage'],
    'industry_category', 'Real Estate'
)
WHERE code_type = 'NAICS' AND code IN ('531210', '531311', '531312')
   OR code_type = 'SIC' AND code = '6531'
   OR code_type = 'MCC' AND code = '6512';

-- =====================================================
-- Part 8: Professional Services Industry Mappings
-- =====================================================

-- Professional Services - Legal Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Legal Services', 'Law', 'Legal Consulting'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'NAICS' AND code IN ('541110', '541199')
   OR code_type = 'SIC' AND code = '8111'
   OR code_type = 'MCC' AND code = '8111';

-- Professional Services - Accounting Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Accounting', 'Tax Services', 'Financial Consulting'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'NAICS' AND code IN ('541211', '541219', '541214')
   OR code_type = 'SIC' AND code = '8721'
   OR code_type = 'MCC' AND code = '8721';

-- =====================================================
-- Part 9: Education Industry Mappings
-- =====================================================

-- Education - K-12 Schools
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Education',
    'secondary_industries', ARRAY['Elementary Education', 'Secondary Education', 'K-12 Education'],
    'industry_category', 'Education'
)
WHERE code_type = 'NAICS' AND code IN ('611110', '611210')
   OR code_type = 'SIC' AND code IN ('8211', '8222')
   OR code_type = 'MCC' AND code = '8211';

-- Education - Higher Education
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Education',
    'secondary_industries', ARRAY['Higher Education', 'Colleges', 'Universities'],
    'industry_category', 'Education'
)
WHERE code_type = 'NAICS' AND code IN ('611310', '611410', '611420')
   OR code_type = 'SIC' AND code = '8221'
   OR code_type = 'MCC' AND code = '8221';

-- =====================================================
-- Part 10: Transportation Industry Mappings
-- =====================================================

-- Transportation - Air Transportation
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Air Transportation', 'Airlines', 'Aviation'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('481111', '481211', '481212')
   OR code_type = 'SIC' AND code IN ('4512', '4513')
   OR code_type = 'MCC' AND code = '4512';

-- Transportation - Trucking
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Trucking', 'Freight Transportation', 'Logistics'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('484110', '484121', '484122', '484230')
   OR code_type = 'SIC' AND code IN ('4212', '4213', '4214')
   OR code_type = 'MCC' AND code = '4212';

-- Transportation - Courier Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Courier Services', 'Package Delivery', 'Express Delivery'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('492110', '492210')
   OR code_type = 'SIC' AND code IN ('4513', '4215')
   OR code_type = 'MCC' AND code = '4512';

-- Transportation - Public Transit
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Public Transit', 'Urban Transportation', 'Transit Systems'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('485110', '485210', '485310', '485990')
   OR code_type = 'SIC' AND code IN ('4111', '4119')
   OR code_type = 'MCC' AND code = '4111';

-- =====================================================
-- Part 11: Accommodation Industry Mappings
-- =====================================================

-- Accommodation - Hotels
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Hospitality',
    'secondary_industries', ARRAY['Hotels', 'Lodging', 'Accommodation'],
    'industry_category', 'Hospitality'
)
WHERE code_type = 'NAICS' AND code IN ('721110', '721120', '721191')
   OR code_type = 'SIC' AND code IN ('7011', '7012', '7021')
   OR code_type = 'MCC' AND code = '7011';

-- Accommodation - RV Parks and Campgrounds
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Hospitality',
    'secondary_industries', ARRAY['RV Parks', 'Campgrounds', 'Recreation'],
    'industry_category', 'Hospitality'
)
WHERE code_type = 'NAICS' AND code IN ('721211', '721214')
   OR code_type = 'SIC' AND code = '7012'
   OR code_type = 'MCC' AND code = '7011';

-- =====================================================
-- Part 12: Arts and Entertainment Industry Mappings
-- =====================================================

-- Arts and Entertainment - Theater
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Arts & Entertainment',
    'secondary_industries', ARRAY['Theater', 'Performing Arts', 'Entertainment'],
    'industry_category', 'Entertainment'
)
WHERE code_type = 'NAICS' AND code IN ('711110', '711320', '711310')
   OR code_type = 'SIC' AND code IN ('7922', '7929')
   OR code_type = 'MCC' AND code = '7922';

-- Arts and Entertainment - Amusement Parks
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Arts & Entertainment',
    'secondary_industries', ARRAY['Amusement Parks', 'Theme Parks', 'Recreation'],
    'industry_category', 'Entertainment'
)
WHERE code_type = 'NAICS' AND code IN ('713110', '713120', '713210')
   OR code_type = 'SIC' AND code = '7996'
   OR code_type = 'MCC' AND code = '7996';

-- =====================================================
-- Part 13: Manufacturing Industry Mappings
-- =====================================================

-- Manufacturing - Food Manufacturing
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Manufacturing',
    'secondary_industries', ARRAY['Food Manufacturing', 'Food Production', 'Bakery Products'],
    'industry_category', 'Manufacturing'
)
WHERE code_type = 'NAICS' AND code IN ('311811', '311812', '311821')
   OR code_type = 'SIC' AND code IN ('2051', '2052')
   OR code_type = 'MCC' AND code IN ('5462', '5411');

-- =====================================================
-- Verification Queries
-- =====================================================

-- Count codes with industry mappings
SELECT 
    'Codes with Industry Mappings' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 120 THEN '✅ PASS - 120+ codes (80%+)'
        ELSE '❌ FAIL - Below 120 codes'
    END AS status
FROM code_metadata
WHERE is_active = true
  AND industry_mappings != '{}'::jsonb
  AND industry_mappings IS NOT NULL;

-- Industry mapping coverage by code type
SELECT 
    'Industry Mapping Coverage by Code Type' AS metric,
    code_type,
    COUNT(*) FILTER (WHERE industry_mappings != '{}'::jsonb) AS codes_with_mappings,
    COUNT(*) AS total_codes,
    ROUND(COUNT(*) FILTER (WHERE industry_mappings != '{}'::jsonb) * 100.0 / COUNT(*), 2) AS coverage_percentage
FROM code_metadata
WHERE is_active = true
GROUP BY code_type
ORDER BY code_type;

-- Primary industry distribution
SELECT 
    'Primary Industry Distribution' AS metric,
    industry_mappings->>'primary_industry' AS primary_industry,
    COUNT(*) AS code_count
FROM code_metadata
WHERE is_active = true
  AND industry_mappings != '{}'::jsonb
  AND industry_mappings->>'primary_industry' IS NOT NULL
GROUP BY industry_mappings->>'primary_industry'
ORDER BY code_count DESC;

-- Sample industry mapping verification
SELECT 
    'Sample Industry Mappings' AS example,
    code_type,
    code,
    official_name,
    industry_mappings->>'primary_industry' AS primary_industry,
    industry_mappings->'secondary_industries' AS secondary_industries,
    industry_mappings->>'industry_category' AS industry_category
FROM code_metadata
WHERE is_active = true
  AND industry_mappings != '{}'::jsonb
LIMIT 10;

