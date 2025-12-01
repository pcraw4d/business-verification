-- =====================================================
-- Expand Industry Mapping Coverage - Phase 4 Supplement
-- Purpose: Increase industry mapping coverage from 35.29% to 80%+ (435+ codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 4 (Supplement)
-- =====================================================
-- 
-- Current: 192 codes with industry mappings (35.29%)
-- Target: 435+ codes with industry mappings (80%+)
-- Need: ~243 additional codes with industry mappings
-- =====================================================

-- =====================================================
-- Part 1: Additional Construction Industry Mappings (30 codes)
-- =====================================================

-- Construction - General Contractors
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Construction',
    'secondary_industries', ARRAY['Building Construction', 'General Contracting', 'Residential Construction'],
    'industry_category', 'Construction'
)
WHERE code_type = 'NAICS' AND code IN ('236115', '236116', '236117', '236210', '236220')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Construction - Specialty Trade Contractors
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Construction',
    'secondary_industries', ARRAY['Specialty Trade', 'Contracting', 'Construction Services'],
    'industry_category', 'Construction'
)
WHERE code_type = 'NAICS' AND code IN ('238110', '238120', '238130', '238140', '238150', '238160', '238170', '238210', '238220')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Construction - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Construction',
    'secondary_industries', ARRAY['Construction', 'Contracting', 'Building Services'],
    'industry_category', 'Construction'
)
WHERE code_type = 'SIC' AND code IN ('1711', '1721', '1731', '1741', '1742', '1751', '1752', '1761', '1771', '1791', '1793', '1794', '1795', '1796', '1799')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 2: Additional Financial Services Industry Mappings (25 codes)
-- =====================================================

-- Financial Services - Banking
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Financial Services',
    'secondary_industries', ARRAY['Banking', 'Depository Institutions', 'Credit Services'],
    'industry_category', 'Financial Services'
)
WHERE code_type = 'NAICS' AND code IN ('522110', '522120', '522130', '522190', '522220')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Financial Services - Investment Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Financial Services',
    'secondary_industries', ARRAY['Investment Services', 'Securities', 'Financial Planning'],
    'industry_category', 'Financial Services'
)
WHERE code_type = 'NAICS' AND code IN ('523110', '523120', '523130', '523140', '523910', '523920', '523930')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Financial Services - Insurance
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Financial Services',
    'secondary_industries', ARRAY['Insurance', 'Risk Management', 'Financial Protection'],
    'industry_category', 'Financial Services'
)
WHERE code_type = 'NAICS' AND code IN ('524113', '524114', '524126', '524127', '524128', '524130')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Financial Services - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Financial Services',
    'secondary_industries', ARRAY['Banking', 'Financial Services', 'Credit'],
    'industry_category', 'Financial Services'
)
WHERE code_type = 'SIC' AND code IN ('6099', '6162', '6163')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 3: Additional Technology Industry Mappings (20 codes)
-- =====================================================

-- Technology - Software and IT Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Technology',
    'secondary_industries', ARRAY['Software', 'IT Services', 'Computer Services', 'Technology'],
    'industry_category', 'Information Technology'
)
WHERE code_type = 'NAICS' AND code IN ('541511', '541512', '541519', '541611', '541612', '541613', '541614', '541618', '541621')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Technology - Engineering and Consulting
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Technology',
    'secondary_industries', ARRAY['Engineering', 'Consulting', 'Professional Services', 'Technical Services'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'NAICS' AND code IN ('541310', '541320', '541330', '541370', '541380')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Technology - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Technology',
    'secondary_industries', ARRAY['Computer Services', 'Technology', 'IT Services'],
    'industry_category', 'Information Technology'
)
WHERE code_type = 'SIC' AND code IN ('5047', '5048', '7377', '7378', '7379')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Technology - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Technology',
    'secondary_industries', ARRAY['Computer Equipment', 'Software', 'Technology'],
    'industry_category', 'Information Technology'
)
WHERE code_type = 'MCC' AND code IN ('5045', '5046', '5049', '4814', '4816')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 4: Additional Transportation Industry Mappings (25 codes)
-- =====================================================

-- Transportation - Air Transportation
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Air Transportation', 'Aviation', 'Transportation Services'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('481111', '481112', '481211', '481212')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Transportation - Rail Transportation
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Rail Transportation', 'Railroad', 'Transportation Services'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('482111', '482112')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Transportation - Water Transportation
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Water Transportation', 'Maritime', 'Shipping'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('483111', '483112', '483113')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Transportation - Truck Transportation
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Truck Transportation', 'Freight', 'Logistics'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('484110', '484121', '484122', '484210', '484220', '484230')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Transportation - Bus Transportation
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Bus Transportation', 'Public Transit', 'Transportation Services'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('485210', '485310', '485320', '485410')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Transportation - Couriers and Messengers
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Courier Services', 'Delivery', 'Messenger Services'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'NAICS' AND code IN ('492110', '492210')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Transportation - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Transportation', 'Travel', 'Transportation Services'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'MCC' AND code IN ('4111', '4112', '4119', '4121', '4131', '4215', '4411', '4511', '4722', '4784')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Transportation - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation',
    'secondary_industries', ARRAY['Transportation', 'Shipping', 'Transportation Services'],
    'industry_category', 'Transportation'
)
WHERE code_type = 'SIC' AND code IN ('4119', '4215')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 5: Additional Retail & Commerce Industry Mappings (30 codes)
-- =====================================================

-- Retail - General Merchandise
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Retail', 'General Merchandise', 'Commerce'],
    'industry_category', 'Retail'
)
WHERE code_type = 'NAICS' AND code IN ('452111', '452112', '452990')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Retail - Specialty Stores
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Retail', 'Specialty Stores', 'Commerce'],
    'industry_category', 'Retail'
)
WHERE code_type = 'NAICS' AND code IN ('453110', '453210', '453220', '453310', '453910', '453920', '453930')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Retail - Food Stores
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Retail', 'Food', 'Grocery', 'Commerce'],
    'industry_category', 'Retail'
)
WHERE code_type = 'NAICS' AND code IN ('445210', '445220', '445230', '445291', '445299')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Retail - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Retail', 'Commerce', 'Merchandise'],
    'industry_category', 'Retail'
)
WHERE code_type = 'MCC' AND code IN ('5310', '5311', '5331', '5399', '5451', '5499', '5511', '5521', '5531', '5532', '5541', '5542', '5571', '5651', '5661', '5681', '5697', '5698', '5699', '5712', '5713', '5714', '5719', '5722', '5941', '5942', '5943', '5944', '5945', '5946', '5947', '5948', '5949')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Retail - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Retail', 'Commerce', 'Merchandise'],
    'industry_category', 'Retail'
)
WHERE code_type = 'SIC' AND code IN ('5451', '5499', '5511', '5521', '5531', '5541')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 6: Additional Food & Beverage Industry Mappings (25 codes)
-- =====================================================

-- Food & Beverage - Manufacturing
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Food & Beverage',
    'secondary_industries', ARRAY['Food Manufacturing', 'Beverage Manufacturing', 'Food Processing'],
    'industry_category', 'Food & Beverage'
)
WHERE code_type = 'NAICS' AND code IN ('311113', '311119', '311211', '311212', '311213', '311221', '311225', '311230', '311311', '311312', '311313', '311320', '311330', '311340', '311410', '311411', '311421', '311422', '311423', '311511', '311512', '311513', '311514', '311520', '311615')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Food & Beverage - Retail
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Food & Beverage',
    'secondary_industries', ARRAY['Food Retail', 'Beverage Retail', 'Grocery'],
    'industry_category', 'Food & Beverage'
)
WHERE code_type = 'NAICS' AND code IN ('445210', '445220', '445230', '445291', '445299')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Food & Beverage - Restaurants
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Food & Beverage',
    'secondary_industries', ARRAY['Restaurants', 'Food Service', 'Dining'],
    'industry_category', 'Food & Beverage'
)
WHERE code_type = 'NAICS' AND code IN ('722511', '722513', '722514', '722515', '722320', '722330')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Food & Beverage - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Food & Beverage',
    'secondary_industries', ARRAY['Food Service', 'Restaurants', 'Dining'],
    'industry_category', 'Food & Beverage'
)
WHERE code_type = 'MCC' AND code IN ('5462', '5811', '5812', '5814', '5992')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Food & Beverage - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Food & Beverage',
    'secondary_industries', ARRAY['Food Manufacturing', 'Beverage Manufacturing'],
    'industry_category', 'Food & Beverage'
)
WHERE code_type = 'SIC' AND code IN ('2061', '2062', '2082')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 7: Additional Healthcare Industry Mappings (20 codes)
-- =====================================================

-- Healthcare - Medical Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Healthcare',
    'secondary_industries', ARRAY['Medical Services', 'Healthcare', 'Health Services'],
    'industry_category', 'Healthcare'
)
WHERE code_type = 'NAICS' AND code IN ('621110', '621210', '621310', '621320', '621330', '621340', '621410', '621420', '621491', '621492', '621493', '621498', '621510', '621610', '621910', '621991', '621999', '622110')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Healthcare - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Healthcare',
    'secondary_industries', ARRAY['Medical Services', 'Healthcare', 'Health Services'],
    'industry_category', 'Healthcare'
)
WHERE code_type = 'MCC' AND code IN ('4119', '5047', '5912', '5975', '5976', '5977', '8011', '8021', '8031', '8041', '8042', '8043', '8049', '8062', '8071', '8099')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 8: Additional Education Industry Mappings (15 codes)
-- =====================================================

-- Education - Schools
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Education',
    'secondary_industries', ARRAY['Education', 'Schools', 'Educational Services'],
    'industry_category', 'Education'
)
WHERE code_type = 'NAICS' AND code IN ('611110', '611310', '611410', '611420', '611430', '611511', '611512', '611519', '611610', '611620', '611630', '611691', '611692', '611699')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Education - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Education',
    'secondary_industries', ARRAY['Education', 'Schools', 'Educational Services'],
    'industry_category', 'Education'
)
WHERE code_type = 'MCC' AND code IN ('8211', '8220', '8241', '8244', '8249', '8299')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Education - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Education',
    'secondary_industries', ARRAY['Education', 'Schools', 'Educational Services'],
    'industry_category', 'Education'
)
WHERE code_type = 'SIC' AND code IN ('8243', '8244', '8249', '8299')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 9: Additional Real Estate Industry Mappings (10 codes)
-- =====================================================

-- Real Estate - Property Management
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Real Estate',
    'secondary_industries', ARRAY['Real Estate', 'Property Management', 'Real Estate Services'],
    'industry_category', 'Real Estate'
)
WHERE code_type = 'NAICS' AND code IN ('531110', '531120', '531190', '531210', '531311', '531312', '531320', '531390')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Real Estate - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Real Estate',
    'secondary_industries', ARRAY['Real Estate', 'Property Management'],
    'industry_category', 'Real Estate'
)
WHERE code_type = 'SIC' AND code IN ('6515', '6517', '6541')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 10: Additional Professional Services Industry Mappings (15 codes)
-- =====================================================

-- Professional Services - Legal
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Legal Services', 'Professional Services', 'Law'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'NAICS' AND code IN ('541110', '541199')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Professional Services - Consulting
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Consulting', 'Professional Services', 'Business Services'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'NAICS' AND code IN ('541611', '541612', '541613', '541614', '541618', '541620', '541690', '541214', '541370')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Professional Services - SIC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Professional Services', 'Consulting', 'Business Services'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'SIC' AND code IN ('8711', '8712')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Professional Services - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Professional Services', 'Business Services'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'MCC' AND code IN ('7371', '7372', '7379', '7399', '8111', '8721')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 11: Additional Manufacturing Industry Mappings (15 codes)
-- =====================================================

-- Manufacturing - Food Products
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Manufacturing',
    'secondary_industries', ARRAY['Food Manufacturing', 'Manufacturing', 'Food Processing'],
    'industry_category', 'Manufacturing'
)
WHERE code_type = 'NAICS' AND code IN ('311113', '311119', '311211', '311212', '311213', '311221', '311225', '311230')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Manufacturing - Beverages
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Manufacturing',
    'secondary_industries', ARRAY['Beverage Manufacturing', 'Manufacturing', 'Food & Beverage'],
    'industry_category', 'Manufacturing'
)
WHERE code_type = 'NAICS' AND code IN ('312111', '312112', '312113', '312120', '312130', '312140')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Manufacturing - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Manufacturing',
    'secondary_industries', ARRAY['Manufacturing', 'Industrial', 'Production'],
    'industry_category', 'Manufacturing'
)
WHERE code_type = 'MCC' AND code IN ('5013', '5021', '5039', '5046', '5049', '5051', '5065', '5072', '5074', '5085', '5094', '5099', '5122', '5131', '5137', '5139')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 12: Additional Arts & Entertainment Industry Mappings (10 codes)
-- =====================================================

-- Arts & Entertainment - Performing Arts
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Arts & Entertainment',
    'secondary_industries', ARRAY['Entertainment', 'Arts', 'Performing Arts'],
    'industry_category', 'Arts & Entertainment'
)
WHERE code_type = 'NAICS' AND code IN ('711110', '711120', '711130', '711190', '711211', '711212', '711219', '711310', '711320', '711410', '711510', '712110')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Arts & Entertainment - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Arts & Entertainment',
    'secondary_industries', ARRAY['Entertainment', 'Arts', 'Recreation'],
    'industry_category', 'Arts & Entertainment'
)
WHERE code_type = 'MCC' AND code IN ('5733', '5735', '5971', '5972', '7941', '7999')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 13: Additional Accommodation Industry Mappings (10 codes)
-- =====================================================

-- Accommodation - Hotels and Lodging
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Hospitality',
    'secondary_industries', ARRAY['Hotels', 'Lodging', 'Accommodation', 'Hospitality'],
    'industry_category', 'Hospitality'
)
WHERE code_type = 'NAICS' AND code IN ('721110', '721120', '721191', '721199', '721211', '721214', '721310')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Accommodation - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Hospitality',
    'secondary_industries', ARRAY['Hotels', 'Lodging', 'Accommodation'],
    'industry_category', 'Hospitality'
)
WHERE code_type = 'MCC' AND code IN ('3501', '3502', '3503', '3504', '3505', '3506', '3507', '3508', '3509', '3510', '3511', '3512', '3513', '3514', '3515', '3516', '3517', '3518', '3519', '3520')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 14: Additional Utilities Industry Mappings (5 codes)
-- =====================================================

-- Utilities - Energy
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Utilities',
    'secondary_industries', ARRAY['Utilities', 'Energy', 'Public Utilities'],
    'industry_category', 'Utilities'
)
WHERE code_type = 'NAICS' AND code IN ('221110', '221210', '221310', '221320')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Utilities - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Utilities',
    'secondary_industries', ARRAY['Utilities', 'Energy', 'Public Utilities'],
    'industry_category', 'Utilities'
)
WHERE code_type = 'MCC' AND code IN ('4900')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Part 15: Additional Miscellaneous Industry Mappings (20 codes)
-- =====================================================

-- Miscellaneous - Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Services', 'Business Services', 'Professional Services'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'NAICS' AND code IN ('541199', '541214', '541370', '541410', '541420', '541930')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Miscellaneous - Retail Services
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Retail', 'Services', 'Commerce'],
    'industry_category', 'Retail'
)
WHERE code_type = 'NAICS' AND code IN ('452990', '454110', '454210')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Miscellaneous - MCC Codes
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Retail', 'Services', 'Commerce'],
    'industry_category', 'Retail'
)
WHERE code_type = 'MCC' AND code IN ('5309', '5399', '5961', '5962', '5963', '5970', '5972', '5973', '5974', '5978', '5983', '5993', '5994', '5995', '5996', '5997', '5998', '5999', '6051')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Verification Query
-- =====================================================

-- Verify industry mapping coverage after supplement
SELECT 
    'Industry Mapping Coverage After Supplement' AS metric,
    COUNT(*) AS codes_with_mappings,
    (SELECT COUNT(*) FROM code_metadata WHERE is_active = true) AS total_codes,
    ROUND(COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 0), 2) AS coverage_percentage,
    CASE 
        WHEN COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 0) >= 80.0 THEN '✅ PASS - 80%+ coverage'
        ELSE '❌ FAIL - Below 80% coverage'
    END AS status
FROM code_metadata
WHERE is_active = true
  AND industry_mappings != '{}'::jsonb
  AND industry_mappings IS NOT NULL;

