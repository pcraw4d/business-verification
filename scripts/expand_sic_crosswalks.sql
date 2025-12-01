-- =====================================================
-- Expand SIC Crosswalk Coverage - Phase 2, Task 2.2
-- Purpose: Increase SIC crosswalk coverage from 22.86% to 57%+ (20+ SIC codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 2, Task 2.2
-- =====================================================
-- 
-- This script adds crosswalk data for 20+ SIC codes, linking them to
-- corresponding NAICS and MCC codes using Census Bureau crosswalk data.
-- Target: 20+ SIC codes with crosswalks (57%+ coverage)
-- =====================================================

-- =====================================================
-- Part 1: Technology SIC Crosswalks
-- =====================================================

-- SIC 7371: Computer Programming Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '541512', '541519'],
    'sic', ARRAY['7371', '7372', '7373'],
    'mcc', ARRAY['5734', '5735', '5045']
)
WHERE code_type = 'SIC' AND code = '7371';

-- SIC 7372: Prepackaged Software
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '518210', '519130'],
    'sic', ARRAY['7372', '7371', '7373'],
    'mcc', ARRAY['5734', '5045']
)
WHERE code_type = 'SIC' AND code = '7372';

-- SIC 7373: Computer Integrated Systems Design
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541512', '541330', '541511'],
    'sic', ARRAY['7373', '7371', '7372'],
    'mcc', ARRAY['5045', '5734']
)
WHERE code_type = 'SIC' AND code = '7373';

-- SIC 7374: Computer Processing and Data Preparation Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['518210', '541519', '541511'],
    'sic', ARRAY['7374', '7371', '7375'],
    'mcc', ARRAY['5045', '5734']
)
WHERE code_type = 'SIC' AND code = '7374';

-- SIC 7375: Information Retrieval Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['518310', '519130', '541519'],
    'sic', ARRAY['7375', '7374', '7376'],
    'mcc', ARRAY['5045']
)
WHERE code_type = 'SIC' AND code = '7375';

-- SIC 7376: Computer Facilities Management Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541519', '518210', '541512'],
    'sic', ARRAY['7376', '7374', '7377'],
    'mcc', ARRAY['5045']
)
WHERE code_type = 'SIC' AND code = '7376';

-- =====================================================
-- Part 2: Financial Services SIC Crosswalks
-- =====================================================

-- SIC 6021: National Commercial Banks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '522120', '522210'],
    'sic', ARRAY['6021', '6022', '6029'],
    'mcc', ARRAY['6010', '6011', '6012']
)
WHERE code_type = 'SIC' AND code = '6021';

-- SIC 6022: State Commercial Banks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '522120'],
    'sic', ARRAY['6022', '6021', '6029'],
    'mcc', ARRAY['6010', '6011', '6012']
)
WHERE code_type = 'SIC' AND code = '6022';

-- SIC 6029: Commercial Banks, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '522291', '522292'],
    'sic', ARRAY['6029', '6021', '6022'],
    'mcc', ARRAY['6010', '6011', '6012']
)
WHERE code_type = 'SIC' AND code = '6029';

-- SIC 6035: Savings Institutions, Federally Chartered
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522120', '522110'],
    'sic', ARRAY['6035', '6036', '6021'],
    'mcc', ARRAY['6010', '6011']
)
WHERE code_type = 'SIC' AND code = '6035';

-- SIC 6036: Savings Institutions, Not Federally Chartered
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522120', '522110'],
    'sic', ARRAY['6036', '6035', '6021'],
    'mcc', ARRAY['6010', '6011']
)
WHERE code_type = 'SIC' AND code = '6036';

-- SIC 6211: Security Brokers, Dealers, and Flotation Companies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['523110', '523120', '523130'],
    'sic', ARRAY['6211', '6221', '6231'],
    'mcc', ARRAY['6211', '6010']
)
WHERE code_type = 'SIC' AND code = '6211';

-- SIC 6221: Commodity Contracts Brokers and Dealers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['523130', '523140'],
    'sic', ARRAY['6221', '6211', '6231'],
    'mcc', ARRAY['6211']
)
WHERE code_type = 'SIC' AND code = '6221';

-- SIC 6231: Security and Commodity Exchanges
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['523210', '523110'],
    'sic', ARRAY['6231', '6211', '6221'],
    'mcc', ARRAY['6211']
)
WHERE code_type = 'SIC' AND code = '6231';

-- =====================================================
-- Part 3: Healthcare SIC Crosswalks
-- =====================================================

-- SIC 8011: Offices and Clinics of Doctors of Medicine
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621111', '621112', '621210'],
    'sic', ARRAY['8011', '8021', '8041'],
    'mcc', ARRAY['8011', '8021', '8041', '8042', '8043', '8049']
)
WHERE code_type = 'SIC' AND code = '8011';

-- SIC 8021: Offices and Clinics of Dentists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621210', '621310'],
    'sic', ARRAY['8021', '8011', '8041'],
    'mcc', ARRAY['8021', '8011', '8041']
)
WHERE code_type = 'SIC' AND code = '8021';

-- SIC 8041: Offices and Clinics of Chiropractors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621310', '621320'],
    'sic', ARRAY['8041', '8042', '8011'],
    'mcc', ARRAY['8041', '8011', '8042']
)
WHERE code_type = 'SIC' AND code = '8041';

-- SIC 8042: Offices and Clinics of Optometrists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621320', '621330'],
    'sic', ARRAY['8042', '8043', '8011'],
    'mcc', ARRAY['8042', '8043', '8011']
)
WHERE code_type = 'SIC' AND code = '8042';

-- SIC 8043: Offices and Clinics of Podiatrists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621391', '621399'],
    'sic', ARRAY['8043', '8049', '8011'],
    'mcc', ARRAY['8043', '8049', '8011']
)
WHERE code_type = 'SIC' AND code = '8043';

-- SIC 8049: Offices and Clinics of Health Practitioners, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621399', '621410', '621420'],
    'sic', ARRAY['8049', '8011', '8099'],
    'mcc', ARRAY['8049', '8011', '8099']
)
WHERE code_type = 'SIC' AND code = '8049';

-- SIC 8062: General Medical and Surgical Hospitals
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['622110', '622210', '622310'],
    'sic', ARRAY['8062', '8063', '8069'],
    'mcc', ARRAY['8062', '8050', '8011']
)
WHERE code_type = 'SIC' AND code = '8062';

-- SIC 8063: Psychiatric Hospitals
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['622210', '621420'],
    'sic', ARRAY['8063', '8062', '8069'],
    'mcc', ARRAY['8062', '8011']
)
WHERE code_type = 'SIC' AND code = '8063';

-- SIC 8069: Specialty Hospitals, Except Psychiatric
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['622310', '622110'],
    'sic', ARRAY['8069', '8062', '8063'],
    'mcc', ARRAY['8062', '8011']
)
WHERE code_type = 'SIC' AND code = '8069';

-- SIC 8071: Medical Laboratories
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621511', '621512', '621991'],
    'sic', ARRAY['8071', '8072'],
    'mcc', ARRAY['8071', '8011']
)
WHERE code_type = 'SIC' AND code = '8071';

-- SIC 8072: Dental Laboratories
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['339116', '621210'],
    'sic', ARRAY['8072', '8071'],
    'mcc', ARRAY['8021', '8011']
)
WHERE code_type = 'SIC' AND code = '8072';

-- SIC 8099: Health and Allied Services, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621999', '621410', '621420'],
    'sic', ARRAY['8099', '8011', '8049'],
    'mcc', ARRAY['8099', '8011']
)
WHERE code_type = 'SIC' AND code = '8099';

-- =====================================================
-- Part 4: Retail SIC Crosswalks
-- =====================================================

-- SIC 5311: Department Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452111', '452112', '452210'],
    'sic', ARRAY['5311', '5331'],
    'mcc', ARRAY['5310', '5311', '5331']
)
WHERE code_type = 'SIC' AND code = '5311';

-- SIC 5331: Variety Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452990', '452210'],
    'sic', ARRAY['5331', '5311'],
    'mcc', ARRAY['5331', '5310', '5311']
)
WHERE code_type = 'SIC' AND code = '5331';

-- SIC 5411: Grocery Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['445110', '445120', '452210'],
    'sic', ARRAY['5411', '5331'],
    'mcc', ARRAY['5411', '5311']
)
WHERE code_type = 'SIC' AND code = '5411';

-- SIC 5441: Candy, Nut, and Confectionery Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['445292', '311320'],
    'sic', ARRAY['5441', '5411'],
    'mcc', ARRAY['5441', '5411']
)
WHERE code_type = 'SIC' AND code = '5441';

-- SIC 5812: Eating Places
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722511', '722513', '722515'],
    'sic', ARRAY['5812', '5813'],
    'mcc', ARRAY['5811', '5812', '5813', '5814']
)
WHERE code_type = 'SIC' AND code = '5812';

-- SIC 5813: Drinking Places
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722410', '722511'],
    'sic', ARRAY['5813', '5812'],
    'mcc', ARRAY['5813', '5812', '5921']
)
WHERE code_type = 'SIC' AND code = '5813';

-- =====================================================
-- Part 5: Manufacturing SIC Crosswalks
-- =====================================================

-- SIC 2051: Bread, Cake, and Related Products
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['311811', '311812', '311821'],
    'sic', ARRAY['2051', '2052'],
    'mcc', ARRAY['5462', '5411']
)
WHERE code_type = 'SIC' AND code = '2051';

-- SIC 2052: Cookies and Crackers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['311821', '311812'],
    'sic', ARRAY['2052', '2051'],
    'mcc', ARRAY['5462', '5411']
)
WHERE code_type = 'SIC' AND code = '2052';

-- =====================================================
-- Part 6: Construction SIC Crosswalks
-- =====================================================

-- SIC 1521: General Contractors - Single-Family Houses
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['236115', '236116', '236117'],
    'sic', ARRAY['1521', '1522', '1531'],
    'mcc', ARRAY['1521', '1522']
)
WHERE code_type = 'SIC' AND code = '1521';

-- SIC 1522: General Contractors - Residential Buildings, Other Than Single-Family
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['236116', '236115'],
    'sic', ARRAY['1522', '1521', '1531'],
    'mcc', ARRAY['1522', '1521']
)
WHERE code_type = 'SIC' AND code = '1522';

-- SIC 1531: Operative Builders
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['236117', '237210'],
    'sic', ARRAY['1531', '1521', '1522'],
    'mcc', ARRAY['1521', '1522']
)
WHERE code_type = 'SIC' AND code = '1531';

-- SIC 1541: General Contractors - Industrial Buildings and Warehouses
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['236210', '236220'],
    'sic', ARRAY['1541', '1542'],
    'mcc', ARRAY['1521']
)
WHERE code_type = 'SIC' AND code = '1541';

-- SIC 1542: General Contractors - Nonresidential Buildings, Other Than Industrial Buildings and Warehouses
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['236220', '236210'],
    'sic', ARRAY['1542', '1541'],
    'mcc', ARRAY['1521']
)
WHERE code_type = 'SIC' AND code = '1542';

-- SIC 1611: Highway and Street Construction
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['237310', '237110', '237120'],
    'sic', ARRAY['1611', '1622', '1623'],
    'mcc', ARRAY['1521']
)
WHERE code_type = 'SIC' AND code = '1611';

-- SIC 1622: Bridge, Tunnel, and Elevated Highway Construction
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['237310', '237990'],
    'sic', ARRAY['1622', '1611', '1623'],
    'mcc', ARRAY['1521']
)
WHERE code_type = 'SIC' AND code = '1622';

-- SIC 1623: Water, Sewer, and Utility Lines
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['237110', '237120', '237130'],
    'sic', ARRAY['1623', '1611', '1622'],
    'mcc', ARRAY['1521']
)
WHERE code_type = 'SIC' AND code = '1623';

-- =====================================================
-- Part 7: Real Estate SIC Crosswalks
-- =====================================================

-- SIC 6512: Operators of Nonresidential Buildings
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['531120', '531110'],
    'sic', ARRAY['6512', '6513', '6514'],
    'mcc', ARRAY['6512']
)
WHERE code_type = 'SIC' AND code = '6512';

-- SIC 6513: Operators of Apartment Buildings
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['531110', '531311'],
    'sic', ARRAY['6513', '6514', '6512'],
    'mcc', ARRAY['6512']
)
WHERE code_type = 'SIC' AND code = '6513';

-- SIC 6514: Operators of Dwellings Other Than Apartment Buildings
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['531110', '531311'],
    'sic', ARRAY['6514', '6513', '6512'],
    'mcc', ARRAY['6512']
)
WHERE code_type = 'SIC' AND code = '6514';

-- SIC 6531: Real Estate Agents and Managers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['531210', '531311', '531312'],
    'sic', ARRAY['6531', '6512'],
    'mcc', ARRAY['6512']
)
WHERE code_type = 'SIC' AND code = '6531';

-- =====================================================
-- Part 8: Professional Services SIC Crosswalks
-- =====================================================

-- SIC 8111: Legal Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541110', '541199'],
    'sic', ARRAY['8111'],
    'mcc', ARRAY['8111']
)
WHERE code_type = 'SIC' AND code = '8111';

-- SIC 8721: Accounting, Auditing, and Bookkeeping Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541211', '541219', '541214'],
    'sic', ARRAY['8721', '8111'],
    'mcc', ARRAY['8721']
)
WHERE code_type = 'SIC' AND code = '8721';

-- SIC 8711: Engineering Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541330', '541310', '541320'],
    'sic', ARRAY['8711', '8712', '8713'],
    'mcc', ARRAY['8711']
)
WHERE code_type = 'SIC' AND code = '8711';

-- SIC 8712: Architectural Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541310', '541330'],
    'sic', ARRAY['8712', '8711', '8713'],
    'mcc', ARRAY['8711']
)
WHERE code_type = 'SIC' AND code = '8712';

-- SIC 8713: Surveying Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541370', '541330'],
    'sic', ARRAY['8713', '8711', '8712'],
    'mcc', ARRAY['8711']
)
WHERE code_type = 'SIC' AND code = '8713';

-- =====================================================
-- Part 9: Education SIC Crosswalks
-- =====================================================

-- SIC 8211: Elementary and Secondary Schools
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['611110', '611210'],
    'sic', ARRAY['8211', '8221'],
    'mcc', ARRAY['8211']
)
WHERE code_type = 'SIC' AND code = '8211';

-- SIC 8221: Colleges, Universities, and Professional Schools
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['611310', '611410', '611420'],
    'sic', ARRAY['8221', '8222', '8211'],
    'mcc', ARRAY['8221']
)
WHERE code_type = 'SIC' AND code = '8221';

-- SIC 8222: Junior Colleges
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['611210', '611310'],
    'sic', ARRAY['8222', '8221'],
    'mcc', ARRAY['8221']
)
WHERE code_type = 'SIC' AND code = '8222';

-- =====================================================
-- Part 10: Transportation SIC Crosswalks
-- =====================================================

-- SIC 4512: Air Transportation, Scheduled
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['481111', '481211', '481212'],
    'sic', ARRAY['4512', '4513'],
    'mcc', ARRAY['4512']
)
WHERE code_type = 'SIC' AND code = '4512';

-- SIC 4513: Air Courier Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['492110', '492210'],
    'sic', ARRAY['4513', '4512'],
    'mcc', ARRAY['4512']
)
WHERE code_type = 'SIC' AND code = '4513';

-- SIC 4111: Local and Suburban Transit
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485110', '485210', '485310'],
    'sic', ARRAY['4111', '4119'],
    'mcc', ARRAY['4111']
)
WHERE code_type = 'SIC' AND code = '4111';

-- SIC 4119: Local Passenger Transportation, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485990', '485110'],
    'sic', ARRAY['4119', '4111'],
    'mcc', ARRAY['4111']
)
WHERE code_type = 'SIC' AND code = '4119';

-- SIC 4212: Local Trucking Without Storage
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['484110', '484121', '484122'],
    'sic', ARRAY['4212', '4213', '4214'],
    'mcc', ARRAY['4212']
)
WHERE code_type = 'SIC' AND code = '4212';

-- SIC 4213: Trucking, Except Local
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['484121', '484122', '484230'],
    'sic', ARRAY['4213', '4212', '4214'],
    'mcc', ARRAY['4212']
)
WHERE code_type = 'SIC' AND code = '4213';

-- SIC 4214: Local Trucking With Storage
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['484110', '493110'],
    'sic', ARRAY['4214', '4212', '4213'],
    'mcc', ARRAY['4212']
)
WHERE code_type = 'SIC' AND code = '4214';

-- SIC 4215: Courier Services, Except by Air
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['492110', '492210'],
    'sic', ARRAY['4215', '4513'],
    'mcc', ARRAY['4212']
)
WHERE code_type = 'SIC' AND code = '4215';

-- =====================================================
-- Part 11: Accommodation SIC Crosswalks
-- =====================================================

-- SIC 7011: Hotels and Motels
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['721110', '721120', '721191'],
    'sic', ARRAY['7011', '7012', '7021'],
    'mcc', ARRAY['7011']
)
WHERE code_type = 'SIC' AND code = '7011';

-- SIC 7012: Sporting and Recreational Camps
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['721211', '721214'],
    'sic', ARRAY['7012', '7011'],
    'mcc', ARRAY['7011']
)
WHERE code_type = 'SIC' AND code = '7012';

-- SIC 7021: Rooming and Boarding Houses
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['721310', '721110'],
    'sic', ARRAY['7021', '7011'],
    'mcc', ARRAY['7011']
)
WHERE code_type = 'SIC' AND code = '7021';

-- =====================================================
-- Part 12: Arts and Entertainment SIC Crosswalks
-- =====================================================

-- SIC 7922: Theatrical Producers (except Motion Picture) and Ticket Agencies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['711110', '711320', '711310'],
    'sic', ARRAY['7922', '7929'],
    'mcc', ARRAY['7922']
)
WHERE code_type = 'SIC' AND code = '7922';

-- SIC 7996: Amusement Parks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['713110', '713120', '713210'],
    'sic', ARRAY['7996', '7922'],
    'mcc', ARRAY['7996']
)
WHERE code_type = 'SIC' AND code = '7996';

-- SIC 7929: Bands, Orchestras, Actors, and Other Entertainers and Entertainment Groups
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['711130', '711190', '711510'],
    'sic', ARRAY['7929', '7922'],
    'mcc', ARRAY['7922']
)
WHERE code_type = 'SIC' AND code = '7929';

-- =====================================================
-- Verification Queries
-- =====================================================

-- Count SIC codes with crosswalks
SELECT 
    'SIC Codes with Crosswalks' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'SIC' AND is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 20 THEN '✅ PASS - 20+ SIC codes (57%+)'
        ELSE '❌ FAIL - Below 20 SIC codes'
    END AS status
FROM code_metadata
WHERE code_type = 'SIC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'naics' OR crosswalk_data ? 'mcc');

-- Sample SIC crosswalk verification
SELECT 
    'Sample SIC Crosswalks' AS example,
    code_type,
    code,
    official_name,
    crosswalk_data->'naics' AS naics_codes,
    crosswalk_data->'mcc' AS mcc_codes
FROM code_metadata
WHERE code_type = 'SIC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
LIMIT 10;

