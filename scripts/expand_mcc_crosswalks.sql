-- =====================================================
-- Expand MCC Crosswalk Coverage - Phase 2, Task 2.1
-- Purpose: Increase MCC crosswalk coverage from 10.61% to 45%+ (30+ MCC codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 2, Task 2.1
-- =====================================================
-- 
-- This script adds crosswalk data for 30+ MCC codes, linking them to
-- corresponding NAICS and SIC codes using payment processor crosswalk data.
-- Target: 30+ MCC codes with crosswalks (45%+ coverage)
-- =====================================================

-- =====================================================
-- Part 1: Technology MCC Crosswalks
-- =====================================================

-- MCC 5734: Computer Software Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '541512', '454110'],
    'sic', ARRAY['7371', '7372', '5734'],
    'mcc', ARRAY['5734', '5045']
)
WHERE code_type = 'MCC' AND code = '5734';

-- MCC 5735: Record Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451220', '451211'],
    'sic', ARRAY['5735', '5942'],
    'mcc', ARRAY['5735', '5942']
)
WHERE code_type = 'MCC' AND code = '5735';

-- MCC 5045: Computers, Computer Peripheral Equipment, Software
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '541512', '423430', '454110'],
    'sic', ARRAY['7371', '7372', '5045'],
    'mcc', ARRAY['5045', '5734']
)
WHERE code_type = 'MCC' AND code = '5045';

-- MCC 5046: Commercial Equipment, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['423490', '423830'],
    'sic', ARRAY['5046', '5047'],
    'mcc', ARRAY['5046', '5047']
)
WHERE code_type = 'MCC' AND code = '5046';

-- MCC 5047: Medical, Dental, Ophthalmic, and Hospital Equipment and Supplies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['423450', '339112', '621111'],
    'sic', ARRAY['5047', '8011'],
    'mcc', ARRAY['5047', '8011']
)
WHERE code_type = 'MCC' AND code = '5047';

-- MCC 5048: Optical Goods, Photographic Equipment, and Supplies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['423410', '423920', '541922'],
    'sic', ARRAY['5048', '5946'],
    'mcc', ARRAY['5048', '5946']
)
WHERE code_type = 'MCC' AND code = '5048';

-- =====================================================
-- Part 2: Financial Services MCC Crosswalks
-- =====================================================

-- MCC 6010: Financial Institutions - Manual Cash Disbursements
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '522120', '522210'],
    'sic', ARRAY['6021', '6022', '6035'],
    'mcc', ARRAY['6010', '6011', '6012']
)
WHERE code_type = 'MCC' AND code = '6010';

-- MCC 6011: Automated Cash Disbursements
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '522120', '522320'],
    'sic', ARRAY['6021', '6022', '6099'],
    'mcc', ARRAY['6011', '6010', '6012']
)
WHERE code_type = 'MCC' AND code = '6011';

-- MCC 6012: Financial Institutions - Merchandise, Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '522291', '522292'],
    'sic', ARRAY['6021', '6022', '6141'],
    'mcc', ARRAY['6012', '6010', '6011']
)
WHERE code_type = 'MCC' AND code = '6012';

-- MCC 6051: Non-Financial Institutions - Foreign Currency, Money Orders, Travelers Checks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522390', '522320'],
    'sic', ARRAY['6099', '6163'],
    'mcc', ARRAY['6051', '6010']
)
WHERE code_type = 'MCC' AND code = '6051';

-- MCC 6211: Security Brokers/Dealers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['523110', '523120', '523130'],
    'sic', ARRAY['6211', '6221', '6231'],
    'mcc', ARRAY['6211', '6010']
)
WHERE code_type = 'MCC' AND code = '6211';

-- MCC 6300: Insurance Sales, Underwriting, and Premiums
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['524113', '524114', '524210'],
    'sic', ARRAY['6311', '6321', '6331'],
    'mcc', ARRAY['6300', '6010']
)
WHERE code_type = 'MCC' AND code = '6300';

-- =====================================================
-- Part 3: Healthcare MCC Crosswalks
-- =====================================================

-- MCC 8011: Doctors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621111', '621112', '621210'],
    'sic', ARRAY['8011', '8021', '8041'],
    'mcc', ARRAY['8011', '8021', '8041']
)
WHERE code_type = 'MCC' AND code = '8011';

-- MCC 8021: Dentists, Orthodontists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621210', '621310'],
    'sic', ARRAY['8021', '8041'],
    'mcc', ARRAY['8021', '8011', '8041']
)
WHERE code_type = 'MCC' AND code = '8021';

-- MCC 8031: Osteopaths
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621111', '621310'],
    'sic', ARRAY['8011', '8041'],
    'mcc', ARRAY['8031', '8011']
)
WHERE code_type = 'MCC' AND code = '8031';

-- MCC 8041: Chiropractors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621310', '621320'],
    'sic', ARRAY['8041', '8042'],
    'mcc', ARRAY['8041', '8011', '8042']
)
WHERE code_type = 'MCC' AND code = '8041';

-- MCC 8042: Optometrists, Ophthalmologists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621320', '621330'],
    'sic', ARRAY['8042', '8043'],
    'mcc', ARRAY['8042', '8043', '8011']
)
WHERE code_type = 'MCC' AND code = '8042';

-- MCC 8043: Opticians, Optical Goods, Eyeglasses
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['446130', '423410'],
    'sic', ARRAY['8043', '5048'],
    'mcc', ARRAY['8043', '8042', '5048']
)
WHERE code_type = 'MCC' AND code = '8043';

-- MCC 8049: Podiatrists, Chiropodists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621391', '621399'],
    'sic', ARRAY['8049', '8011'],
    'mcc', ARRAY['8049', '8011']
)
WHERE code_type = 'MCC' AND code = '8049';

-- MCC 8050: Nursing and Personal Care Facilities
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['623110', '623210', '623220'],
    'sic', ARRAY['8051', '8052', '8059'],
    'mcc', ARRAY['8050', '8062']
)
WHERE code_type = 'MCC' AND code = '8050';

-- MCC 8062: Hospitals
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['622110', '622210', '622310'],
    'sic', ARRAY['8062', '8063', '8069'],
    'mcc', ARRAY['8062', '8050', '8011']
)
WHERE code_type = 'MCC' AND code = '8062';

-- MCC 8071: Medical and Dental Laboratories
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621511', '621512', '621991'],
    'sic', ARRAY['8071', '8072'],
    'mcc', ARRAY['8071', '8011']
)
WHERE code_type = 'MCC' AND code = '8071';

-- MCC 8099: Medical Services and Health Practitioners, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621999', '621410', '621420'],
    'sic', ARRAY['8099', '8011'],
    'mcc', ARRAY['8099', '8011', '8021']
)
WHERE code_type = 'MCC' AND code = '8099';

-- =====================================================
-- Part 4: Retail MCC Crosswalks
-- =====================================================

-- MCC 5310: Discount Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452112', '452210', '452311'],
    'sic', ARRAY['5331', '5311'],
    'mcc', ARRAY['5310', '5311', '5331']
)
WHERE code_type = 'MCC' AND code = '5310';

-- MCC 5311: Department Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452111', '452112', '452210'],
    'sic', ARRAY['5311', '5331'],
    'mcc', ARRAY['5311', '5310', '5331']
)
WHERE code_type = 'MCC' AND code = '5311';

-- MCC 5331: Variety Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452990', '452210'],
    'sic', ARRAY['5331', '5311'],
    'mcc', ARRAY['5331', '5310', '5311']
)
WHERE code_type = 'MCC' AND code = '5331';

-- MCC 5411: Grocery Stores, Supermarkets
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['445110', '445120', '452210'],
    'sic', ARRAY['5411', '5331'],
    'mcc', ARRAY['5411', '5311']
)
WHERE code_type = 'MCC' AND code = '5411';

-- MCC 5511: Car and Truck Dealers (New and Used)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441110', '441120', '441210'],
    'sic', ARRAY['5511', '5521'],
    'mcc', ARRAY['5511', '5521']
)
WHERE code_type = 'MCC' AND code = '5511';

-- MCC 5521: Automobile and Truck Dealers (Used Only)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441120', '441210'],
    'sic', ARRAY['5521', '5511'],
    'mcc', ARRAY['5521', '5511']
)
WHERE code_type = 'MCC' AND code = '5521';

-- MCC 5531: Auto and Home Supply Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441310', '444110', '444130'],
    'sic', ARRAY['5531', '5719'],
    'mcc', ARRAY['5531', '5712']
)
WHERE code_type = 'MCC' AND code = '5531';

-- MCC 5541: Service Stations (With or Without Ancillary Services)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['447110', '447190'],
    'sic', ARRAY['5541', '5542'],
    'mcc', ARRAY['5541', '5542']
)
WHERE code_type = 'MCC' AND code = '5541';

-- MCC 5542: Automated Fuel Dispensers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['447110', '447190'],
    'sic', ARRAY['5542', '5541'],
    'mcc', ARRAY['5542', '5541']
)
WHERE code_type = 'MCC' AND code = '5542';

-- =====================================================
-- Part 5: Food & Beverage MCC Crosswalks
-- =====================================================

-- MCC 5811: Caterers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722320', '722410'],
    'sic', ARRAY['5812', '5811'],
    'mcc', ARRAY['5811', '5812', '5814']
)
WHERE code_type = 'MCC' AND code = '5811';

-- MCC 5812: Eating Places, Restaurants
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722511', '722513', '722515'],
    'sic', ARRAY['5812', '5813'],
    'mcc', ARRAY['5812', '5811', '5813', '5814']
)
WHERE code_type = 'MCC' AND code = '5812';

-- MCC 5813: Drinking Places (Alcoholic Beverages)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722410', '722511'],
    'sic', ARRAY['5813', '5812'],
    'mcc', ARRAY['5813', '5812', '5921']
)
WHERE code_type = 'MCC' AND code = '5813';

-- MCC 5814: Fast Food Restaurants
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722513', '722515'],
    'sic', ARRAY['5812', '5814'],
    'mcc', ARRAY['5814', '5812', '5811']
)
WHERE code_type = 'MCC' AND code = '5814';

-- =====================================================
-- Part 6: Additional High-Value MCC Crosswalks
-- =====================================================

-- MCC 5912: Drug Stores, Pharmacies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['446110', '446191'],
    'sic', ARRAY['5912', '8011'],
    'mcc', ARRAY['5912', '8011']
)
WHERE code_type = 'MCC' AND code = '5912';

-- MCC 5921: Package Stores - Beer, Wine, and Liquor
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['445310', '722410'],
    'sic', ARRAY['5921', '5813'],
    'mcc', ARRAY['5921', '5813']
)
WHERE code_type = 'MCC' AND code = '5921';

-- MCC 5941: Sporting Goods Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451110', '451120'],
    'sic', ARRAY['5941', '5940'],
    'mcc', ARRAY['5941', '5940']
)
WHERE code_type = 'MCC' AND code = '5941';

-- MCC 5942: Book Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451211', '451212'],
    'sic', ARRAY['5942', '5735'],
    'mcc', ARRAY['5942', '5735']
)
WHERE code_type = 'MCC' AND code = '5942';

-- MCC 5943: Stationery, Office, and School Supply Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453210', '423220'],
    'sic', ARRAY['5943', '5045'],
    'mcc', ARRAY['5943', '5045']
)
WHERE code_type = 'MCC' AND code = '5943';

-- MCC 5944: Jewelry Stores, Watches, Clocks, and Silverware Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['448310', '423940'],
    'sic', ARRAY['5944', '5094'],
    'mcc', ARRAY['5944', '5094']
)
WHERE code_type = 'MCC' AND code = '5944';

-- MCC 5945: Hobby, Toy, and Game Shops
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451120', '451130'],
    'sic', ARRAY['5945', '5941'],
    'mcc', ARRAY['5945', '5941']
)
WHERE code_type = 'MCC' AND code = '5945';

-- MCC 5946: Camera and Photographic Supply Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['423410', '443142'],
    'sic', ARRAY['5946', '5048'],
    'mcc', ARRAY['5946', '5048']
)
WHERE code_type = 'MCC' AND code = '5946';

-- MCC 5947: Gift, Card, Novelty, and Souvenir Shops
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453220', '453310'],
    'sic', ARRAY['5947', '5949'],
    'mcc', ARRAY['5947', '5949']
)
WHERE code_type = 'MCC' AND code = '5947';

-- MCC 5948: Luggage and Leather Goods Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['448320', '316992'],
    'sic', ARRAY['5948', '5699'],
    'mcc', ARRAY['5948', '5699']
)
WHERE code_type = 'MCC' AND code = '5948';

-- MCC 5949: Sewing, Needlework, Fabric, and Piece Goods Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451130', '424310'],
    'sic', ARRAY['5949', '5699'],
    'mcc', ARRAY['5949', '5699']
)
WHERE code_type = 'MCC' AND code = '5949';

-- =====================================================
-- Part 7: Verification Queries
-- =====================================================

-- Count MCC codes with crosswalks
SELECT 
    'MCC Codes with Crosswalks' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'MCC' AND is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 30 THEN '✅ PASS - 30+ MCC codes with crosswalks'
        ELSE '❌ FAIL - Below 30 MCC codes with crosswalks'
    END AS status
FROM code_metadata
WHERE code_type = 'MCC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'naics' OR crosswalk_data ? 'sic');

-- Count crosswalk relationships
SELECT 
    'Total Crosswalk Relationships' AS metric,
    SUM(
        jsonb_array_length(COALESCE(crosswalk_data->'naics', '[]'::jsonb)) +
        jsonb_array_length(COALESCE(crosswalk_data->'sic', '[]'::jsonb)) +
        jsonb_array_length(COALESCE(crosswalk_data->'mcc', '[]'::jsonb))
    ) AS total_relationships
FROM code_metadata
WHERE code_type = 'MCC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb;

-- Sample crosswalk verification
SELECT 
    'Sample MCC Crosswalks' AS example,
    code_type,
    code,
    official_name,
    crosswalk_data->'naics' AS naics_codes,
    crosswalk_data->'sic' AS sic_codes
FROM code_metadata
WHERE code_type = 'MCC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
LIMIT 10;

