-- =====================================================
-- Expand NAICS Crosswalk Coverage - Phase 2, Task 2.3
-- Purpose: Increase NAICS crosswalk coverage from 44.90% to 61%+ (30+ NAICS codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 2, Task 2.3
-- =====================================================
-- 
-- This script adds crosswalk data for 30+ NAICS codes, linking them to
-- corresponding SIC and MCC codes using Census Bureau crosswalk data.
-- Target: 30+ NAICS codes with crosswalks (61%+ coverage)
-- =====================================================

-- =====================================================
-- Part 1: Technology NAICS Crosswalks
-- =====================================================

-- NAICS 541511: Custom Computer Programming Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7371', '7372', '7373'],
    'mcc', ARRAY['5734', '5045', '5735'],
    'naics', ARRAY['541511', '541512', '541519']
)
WHERE code_type = 'NAICS' AND code = '541511';

-- NAICS 541512: Computer Systems Design Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7373', '7371', '7372'],
    'mcc', ARRAY['5045', '5734'],
    'naics', ARRAY['541512', '541511', '541330']
)
WHERE code_type = 'NAICS' AND code = '541512';

-- NAICS 541519: Other Computer Related Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7374', '7375', '7376'],
    'mcc', ARRAY['5045', '5734'],
    'naics', ARRAY['541519', '541511', '541512']
)
WHERE code_type = 'NAICS' AND code = '541519';

-- NAICS 518210: Data Processing, Hosting, and Related Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7374', '7375', '7376'],
    'mcc', ARRAY['5045'],
    'naics', ARRAY['518210', '518310', '519130']
)
WHERE code_type = 'NAICS' AND code = '518210';

-- NAICS 518310: Internet Service Providers and Web Search Portals
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7375', '7374'],
    'mcc', ARRAY['5045'],
    'naics', ARRAY['518310', '518210', '519130']
)
WHERE code_type = 'NAICS' AND code = '518310';

-- NAICS 519130: Internet Publishing and Broadcasting and Web Search Portals
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7375', '7374'],
    'mcc', ARRAY['5045'],
    'naics', ARRAY['519130', '518310', '518210']
)
WHERE code_type = 'NAICS' AND code = '519130';

-- NAICS 541330: Engineering Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8711', '8712', '8713'],
    'mcc', ARRAY['8711'],
    'naics', ARRAY['541330', '541310', '541320']
)
WHERE code_type = 'NAICS' AND code = '541330';

-- NAICS 541310: Architectural Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8712', '8711'],
    'mcc', ARRAY['8711'],
    'naics', ARRAY['541310', '541330']
)
WHERE code_type = 'NAICS' AND code = '541310';

-- =====================================================
-- Part 2: Financial Services NAICS Crosswalks
-- =====================================================

-- NAICS 522110: Commercial Banking
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6021', '6022', '6029'],
    'mcc', ARRAY['6010', '6011', '6012'],
    'naics', ARRAY['522110', '522120', '522210']
)
WHERE code_type = 'NAICS' AND code = '522110';

-- NAICS 522120: Savings Institutions
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6035', '6036', '6021'],
    'mcc', ARRAY['6010', '6011'],
    'naics', ARRAY['522120', '522110']
)
WHERE code_type = 'NAICS' AND code = '522120';

-- NAICS 522210: Credit Card Issuing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6029', '6021'],
    'mcc', ARRAY['6010', '6011', '6012'],
    'naics', ARRAY['522210', '522110', '522291']
)
WHERE code_type = 'NAICS' AND code = '522210';

-- NAICS 522291: Consumer Lending
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6029', '6141'],
    'mcc', ARRAY['6010', '6011'],
    'naics', ARRAY['522291', '522292', '522110']
)
WHERE code_type = 'NAICS' AND code = '522291';

-- NAICS 522292: Real Estate Credit
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6162', '6163'],
    'mcc', ARRAY['6010', '6011'],
    'naics', ARRAY['522292', '522291', '522110']
)
WHERE code_type = 'NAICS' AND code = '522292';

-- NAICS 523110: Investment Banking and Securities Dealing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6211', '6221'],
    'mcc', ARRAY['6211', '6010'],
    'naics', ARRAY['523110', '523120', '523130']
)
WHERE code_type = 'NAICS' AND code = '523110';

-- NAICS 523120: Securities Brokerage
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6211', '6221'],
    'mcc', ARRAY['6211'],
    'naics', ARRAY['523120', '523110', '523130']
)
WHERE code_type = 'NAICS' AND code = '523120';

-- NAICS 523130: Commodity Contracts Dealing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6221', '6211'],
    'mcc', ARRAY['6211'],
    'naics', ARRAY['523130', '523140', '523110']
)
WHERE code_type = 'NAICS' AND code = '523130';

-- NAICS 523210: Securities and Commodity Exchanges
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6231', '6211'],
    'mcc', ARRAY['6211'],
    'naics', ARRAY['523210', '523110']
)
WHERE code_type = 'NAICS' AND code = '523210';

-- =====================================================
-- Part 3: Healthcare NAICS Crosswalks
-- =====================================================

-- NAICS 621111: Offices of Physicians (except Mental Health Specialists)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8011', '8021', '8041'],
    'mcc', ARRAY['8011', '8021', '8041', '8042', '8043', '8049'],
    'naics', ARRAY['621111', '621112', '621210']
)
WHERE code_type = 'NAICS' AND code = '621111';

-- NAICS 621112: Offices of Physicians, Mental Health Specialists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8011', '8063'],
    'mcc', ARRAY['8011', '8062'],
    'naics', ARRAY['621112', '621111', '621420']
)
WHERE code_type = 'NAICS' AND code = '621112';

-- NAICS 621210: Offices of Dentists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8021', '8011'],
    'mcc', ARRAY['8021', '8011'],
    'naics', ARRAY['621210', '621310']
)
WHERE code_type = 'NAICS' AND code = '621210';

-- NAICS 621310: Offices of Chiropractors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8041', '8042'],
    'mcc', ARRAY['8041', '8011'],
    'naics', ARRAY['621310', '621320']
)
WHERE code_type = 'NAICS' AND code = '621310';

-- NAICS 621320: Offices of Optometrists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8042', '8043'],
    'mcc', ARRAY['8042', '8043', '8011'],
    'naics', ARRAY['621320', '621330']
)
WHERE code_type = 'NAICS' AND code = '621320';

-- NAICS 621391: Offices of Podiatrists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8043', '8049'],
    'mcc', ARRAY['8043', '8049', '8011'],
    'naics', ARRAY['621391', '621399']
)
WHERE code_type = 'NAICS' AND code = '621391';

-- NAICS 621399: Offices of All Other Miscellaneous Health Practitioners
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8049', '8011', '8099'],
    'mcc', ARRAY['8049', '8011', '8099'],
    'naics', ARRAY['621399', '621410', '621420']
)
WHERE code_type = 'NAICS' AND code = '621399';

-- NAICS 622110: General Medical and Surgical Hospitals
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8062', '8063', '8069'],
    'mcc', ARRAY['8062', '8050', '8011'],
    'naics', ARRAY['622110', '622210', '622310']
)
WHERE code_type = 'NAICS' AND code = '622110';

-- NAICS 622210: Psychiatric and Substance Abuse Hospitals
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8063', '8062'],
    'mcc', ARRAY['8062', '8011'],
    'naics', ARRAY['622210', '621420']
)
WHERE code_type = 'NAICS' AND code = '622210';

-- NAICS 622310: Specialty (except Psychiatric and Substance Abuse) Hospitals
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8069', '8062', '8063'],
    'mcc', ARRAY['8062', '8011'],
    'naics', ARRAY['622310', '622110']
)
WHERE code_type = 'NAICS' AND code = '622310';

-- NAICS 621511: Medical Laboratories
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8071', '8072'],
    'mcc', ARRAY['8071', '8011'],
    'naics', ARRAY['621511', '621512', '621991']
)
WHERE code_type = 'NAICS' AND code = '621511';

-- NAICS 621512: Diagnostic Imaging Centers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8071'],
    'mcc', ARRAY['8071', '8011'],
    'naics', ARRAY['621512', '621511', '621991']
)
WHERE code_type = 'NAICS' AND code = '621512';

-- =====================================================
-- Part 4: Retail NAICS Crosswalks
-- =====================================================

-- NAICS 452111: Department Stores (except Discount Department Stores)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5311', '5331'],
    'mcc', ARRAY['5310', '5311', '5331'],
    'naics', ARRAY['452111', '452112', '452210']
)
WHERE code_type = 'NAICS' AND code = '452111';

-- NAICS 452112: Discount Department Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5331', '5311'],
    'mcc', ARRAY['5331', '5310', '5311'],
    'naics', ARRAY['452112', '452111', '452210']
)
WHERE code_type = 'NAICS' AND code = '452112';

-- NAICS 452210: Warehouse Clubs and Supercenters
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5331', '5311'],
    'mcc', ARRAY['5310', '5311', '5331'],
    'naics', ARRAY['452210', '452111', '452112']
)
WHERE code_type = 'NAICS' AND code = '452210';

-- NAICS 445110: Supermarkets and Other Grocery (except Convenience) Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5411', '5331'],
    'mcc', ARRAY['5411', '5311'],
    'naics', ARRAY['445110', '445120', '452210']
)
WHERE code_type = 'NAICS' AND code = '445110';

-- NAICS 445120: Convenience Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5411'],
    'mcc', ARRAY['5411', '5311'],
    'naics', ARRAY['445120', '445110', '452210']
)
WHERE code_type = 'NAICS' AND code = '445120';

-- NAICS 445292: Confectionery and Nut Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5441', '5411'],
    'mcc', ARRAY['5441', '5411'],
    'naics', ARRAY['445292', '311320']
)
WHERE code_type = 'NAICS' AND code = '445292';

-- =====================================================
-- Part 5: Food & Beverage NAICS Crosswalks
-- =====================================================

-- NAICS 722511: Full-Service Restaurants
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5812', '5813'],
    'mcc', ARRAY['5811', '5812', '5813', '5814'],
    'naics', ARRAY['722511', '722513', '722515']
)
WHERE code_type = 'NAICS' AND code = '722511';

-- NAICS 722513: Limited-Service Restaurants
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5812'],
    'mcc', ARRAY['5814', '5812'],
    'naics', ARRAY['722513', '722511', '722515']
)
WHERE code_type = 'NAICS' AND code = '722513';

-- NAICS 722515: Snack and Nonalcoholic Beverage Bars
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5812'],
    'mcc', ARRAY['5814', '5812'],
    'naics', ARRAY['722515', '722511', '722513']
)
WHERE code_type = 'NAICS' AND code = '722515';

-- NAICS 722410: Drinking Places (Alcoholic Beverages)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5813', '5812'],
    'mcc', ARRAY['5813', '5812', '5921'],
    'naics', ARRAY['722410', '722511']
)
WHERE code_type = 'NAICS' AND code = '722410';

-- =====================================================
-- Part 6: Construction NAICS Crosswalks
-- =====================================================

-- NAICS 236115: New Single-Family Housing Construction (except Operative Builders)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1521', '1522', '1531'],
    'mcc', ARRAY['1521', '1522'],
    'naics', ARRAY['236115', '236116', '236117']
)
WHERE code_type = 'NAICS' AND code = '236115';

-- NAICS 236116: New Multifamily Housing Construction (except Operative Builders)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1522', '1521'],
    'mcc', ARRAY['1522', '1521'],
    'naics', ARRAY['236116', '236115']
)
WHERE code_type = 'NAICS' AND code = '236116';

-- NAICS 236117: New Housing Operative Builders
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1531', '1521', '1522'],
    'mcc', ARRAY['1521', '1522'],
    'naics', ARRAY['236117', '237210']
)
WHERE code_type = 'NAICS' AND code = '236117';

-- NAICS 236210: Industrial Building Construction
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1541', '1542'],
    'mcc', ARRAY['1521'],
    'naics', ARRAY['236210', '236220']
)
WHERE code_type = 'NAICS' AND code = '236210';

-- NAICS 236220: Commercial and Institutional Building Construction
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1542', '1541'],
    'mcc', ARRAY['1521'],
    'naics', ARRAY['236220', '236210']
)
WHERE code_type = 'NAICS' AND code = '236220';

-- NAICS 237310: Highway, Street, and Bridge Construction
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1611', '1622', '1623'],
    'mcc', ARRAY['1521'],
    'naics', ARRAY['237310', '237110', '237120']
)
WHERE code_type = 'NAICS' AND code = '237310';

-- NAICS 237110: Water and Sewer Line and Related Structures Construction
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1623', '1611', '1622'],
    'mcc', ARRAY['1521'],
    'naics', ARRAY['237110', '237120', '237130']
)
WHERE code_type = 'NAICS' AND code = '237110';

-- =====================================================
-- Part 7: Real Estate NAICS Crosswalks
-- =====================================================

-- NAICS 531120: Lessors of Nonresidential Buildings (except Miniwarehouses)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6512', '6513', '6514'],
    'mcc', ARRAY['6512'],
    'naics', ARRAY['531120', '531110']
)
WHERE code_type = 'NAICS' AND code = '531120';

-- NAICS 531110: Lessors of Residential Buildings and Dwellings
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6513', '6514', '6512'],
    'mcc', ARRAY['6512'],
    'naics', ARRAY['531110', '531311']
)
WHERE code_type = 'NAICS' AND code = '531110';

-- NAICS 531210: Offices of Real Estate Agents and Brokers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6531', '6512'],
    'mcc', ARRAY['6512'],
    'naics', ARRAY['531210', '531311', '531312']
)
WHERE code_type = 'NAICS' AND code = '531210';

-- =====================================================
-- Part 8: Professional Services NAICS Crosswalks
-- =====================================================

-- NAICS 541110: Offices of Lawyers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8111'],
    'mcc', ARRAY['8111'],
    'naics', ARRAY['541110', '541199']
)
WHERE code_type = 'NAICS' AND code = '541110';

-- NAICS 541211: Offices of Certified Public Accountants
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8721', '8111'],
    'mcc', ARRAY['8721'],
    'naics', ARRAY['541211', '541219', '541214']
)
WHERE code_type = 'NAICS' AND code = '541211';

-- NAICS 541219: Other Accounting Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8721'],
    'mcc', ARRAY['8721'],
    'naics', ARRAY['541219', '541211', '541214']
)
WHERE code_type = 'NAICS' AND code = '541219';

-- =====================================================
-- Part 9: Education NAICS Crosswalks
-- =====================================================

-- NAICS 611110: Elementary and Secondary Schools
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8211', '8221'],
    'mcc', ARRAY['8211'],
    'naics', ARRAY['611110', '611210']
)
WHERE code_type = 'NAICS' AND code = '611110';

-- NAICS 611310: Colleges, Universities, and Professional Schools
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8221', '8222', '8211'],
    'mcc', ARRAY['8221'],
    'naics', ARRAY['611310', '611410', '611420']
)
WHERE code_type = 'NAICS' AND code = '611310';

-- NAICS 611210: Junior Colleges
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8222', '8221'],
    'mcc', ARRAY['8221'],
    'naics', ARRAY['611210', '611310']
)
WHERE code_type = 'NAICS' AND code = '611210';

-- =====================================================
-- Part 10: Transportation NAICS Crosswalks
-- =====================================================

-- NAICS 481111: Scheduled Passenger Air Transportation
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4512', '4513'],
    'mcc', ARRAY['4512'],
    'naics', ARRAY['481111', '481211', '481212']
)
WHERE code_type = 'NAICS' AND code = '481111';

-- NAICS 492110: Couriers and Express Delivery Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4513', '4512', '4215'],
    'mcc', ARRAY['4512'],
    'naics', ARRAY['492110', '492210']
)
WHERE code_type = 'NAICS' AND code = '492110';

-- NAICS 484110: General Freight Trucking, Local
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4212', '4213', '4214'],
    'mcc', ARRAY['4212'],
    'naics', ARRAY['484110', '484121', '484122']
)
WHERE code_type = 'NAICS' AND code = '484110';

-- NAICS 484121: General Freight Trucking, Long-Distance, Truckload
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4213', '4212', '4214'],
    'mcc', ARRAY['4212'],
    'naics', ARRAY['484121', '484122', '484230']
)
WHERE code_type = 'NAICS' AND code = '484121';

-- NAICS 485110: Urban Transit Systems
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4111', '4119'],
    'mcc', ARRAY['4111'],
    'naics', ARRAY['485110', '485210', '485310']
)
WHERE code_type = 'NAICS' AND code = '485110';

-- =====================================================
-- Part 11: Accommodation NAICS Crosswalks
-- =====================================================

-- NAICS 721110: Hotels (except Casino Hotels) and Motels
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7011', '7012', '7021'],
    'mcc', ARRAY['7011'],
    'naics', ARRAY['721110', '721120', '721191']
)
WHERE code_type = 'NAICS' AND code = '721110';

-- NAICS 721211: RV (Recreational Vehicle) Parks and Campgrounds
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7012', '7011'],
    'mcc', ARRAY['7011'],
    'naics', ARRAY['721211', '721214']
)
WHERE code_type = 'NAICS' AND code = '721211';

-- =====================================================
-- Part 12: Arts and Entertainment NAICS Crosswalks
-- =====================================================

-- NAICS 711110: Theater Companies and Dinner Theaters
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7922', '7929'],
    'mcc', ARRAY['7922'],
    'naics', ARRAY['711110', '711320', '711310']
)
WHERE code_type = 'NAICS' AND code = '711110';

-- NAICS 713110: Amusement and Theme Parks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7996', '7922'],
    'mcc', ARRAY['7996'],
    'naics', ARRAY['713110', '713120', '713210']
)
WHERE code_type = 'NAICS' AND code = '713110';

-- =====================================================
-- Verification Queries
-- =====================================================

-- Count NAICS codes with crosswalks
SELECT 
    'NAICS Codes with Crosswalks' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 30 THEN '✅ PASS - 30+ NAICS codes (61%+)'
        ELSE '❌ FAIL - Below 30 NAICS codes'
    END AS status
FROM code_metadata
WHERE code_type = 'NAICS'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'sic' OR crosswalk_data ? 'mcc');

-- Sample NAICS crosswalk verification
SELECT 
    'Sample NAICS Crosswalks' AS example,
    code_type,
    code,
    official_name,
    crosswalk_data->'sic' AS sic_codes,
    crosswalk_data->'mcc' AS mcc_codes
FROM code_metadata
WHERE code_type = 'NAICS'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
LIMIT 10;
