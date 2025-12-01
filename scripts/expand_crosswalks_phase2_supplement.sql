-- =====================================================
-- Expand Crosswalk Coverage - Phase 2 Supplement
-- Purpose: Increase crosswalk coverage from 33.09% to 50%+ (272+ codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 2 (Supplement)
-- =====================================================
-- 
-- Current: 180 codes with crosswalks (33.09%)
-- Target: 272+ codes with crosswalks (50%+)
-- Need: ~92 additional codes with crosswalks
-- =====================================================

-- =====================================================
-- Part 1: Additional MCC Crosswalks (30 codes)
-- =====================================================

-- MCC 4111: Local and Suburban Passenger Transportation
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485111', '485112', '485113', '485119'],
    'sic', ARRAY['4111', '4119', '4121']
)
WHERE code_type = 'MCC' AND code = '4111'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4112: Passenger Railways
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['482111', '482112'],
    'sic', ARRAY['4011', '4013']
)
WHERE code_type = 'MCC' AND code = '4112'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4119: Local Passenger Transportation, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485210', '485310', '485320'],
    'sic', ARRAY['4119', '4121']
)
WHERE code_type = 'MCC' AND code = '4119'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4121: Taxicabs and Limousines
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485310', '485320'],
    'sic', ARRAY['4121', '4119']
)
WHERE code_type = 'MCC' AND code = '4121'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4131: Bus Lines
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485210', '485410'],
    'sic', ARRAY['4131', '4111']
)
WHERE code_type = 'MCC' AND code = '4131'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4215: Courier Services - Air and Ground
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['492110', '492210'],
    'sic', ARRAY['4215', '4213']
)
WHERE code_type = 'MCC' AND code = '4215'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4411: Steamship and Cruise Lines
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['483111', '483112', '483113'],
    'sic', ARRAY['4411', '4424']
)
WHERE code_type = 'MCC' AND code = '4411'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4457: Boat Dealers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441222', '441228'],
    'sic', ARRAY['5551', '5552']
)
WHERE code_type = 'MCC' AND code = '4457'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4468: Marinas, Service and Supplies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['713930', '441222'],
    'sic', ARRAY['4493', '5551']
)
WHERE code_type = 'MCC' AND code = '4468'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4511: Airlines, Air Carriers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['481111', '481112', '481211', '481212'],
    'sic', ARRAY['4512', '4513']
)
WHERE code_type = 'MCC' AND code = '4511'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4722: Travel Agencies and Tour Operators
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['561510', '561520'],
    'sic', ARRAY['4724', '4725']
)
WHERE code_type = 'MCC' AND code = '4722'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4784: Toll and Bridge Fees
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['488390', '237310'],
    'sic', ARRAY['4789', '1611']
)
WHERE code_type = 'MCC' AND code = '4784'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4814: Telecommunication Equipment and Telephone Sales
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['443142', '517311', '517312'],
    'sic', ARRAY['4812', '4813']
)
WHERE code_type = 'MCC' AND code = '4814'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4816: Computer Network Information Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['518210', '518310', '519130'],
    'sic', ARRAY['7375', '7379']
)
WHERE code_type = 'MCC' AND code = '4816'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 4900: Utilities - Electric, Gas, Water, Sanitary
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['221110', '221210', '221310', '221320'],
    'sic', ARRAY['4911', '4922', '4923', '4924', '4925']
)
WHERE code_type = 'MCC' AND code = '4900'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5013: Motor Vehicle Supplies and New Parts
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441310', '441320'],
    'sic', ARRAY['5013', '5014']
)
WHERE code_type = 'MCC' AND code = '5013'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5021: Office and Commercial Furniture
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['442110', '442210'],
    'sic', ARRAY['5021', '5712']
)
WHERE code_type = 'MCC' AND code = '5021'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5039: Construction Materials, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['423310', '423320', '423330'],
    'sic', ARRAY['5039', '5031']
)
WHERE code_type = 'MCC' AND code = '5039'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5122: Drugs, Drug Proprietaries, and Druggist Sundries
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['446110', '446191'],
    'sic', ARRAY['5912', '5122']
)
WHERE code_type = 'MCC' AND code = '5122'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5131: Piece Goods, Notions, and Other Dry Goods
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['424310', '424320'],
    'sic', ARRAY['5131', '5136']
)
WHERE code_type = 'MCC' AND code = '5131'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5137: Men's, Women's, and Children's Uniforms and Commercial Clothing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['315210', '315220', '315240'],
    'sic', ARRAY['2321', '2322', '2323']
)
WHERE code_type = 'MCC' AND code = '5137'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5139: Commercial Equipment, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['423490', '423830'],
    'sic', ARRAY['5049', '5099']
)
WHERE code_type = 'MCC' AND code = '5139'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5251: Hardware Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['444130'],
    'sic', ARRAY['5251', '5211']
)
WHERE code_type = 'MCC' AND code = '5251'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5309: Duty Free Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452990', '454110'],
    'sic', ARRAY['5311', '5399']
)
WHERE code_type = 'MCC' AND code = '5309'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5310: Discount Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452112', '452910'],
    'sic', ARRAY['5311', '5331']
)
WHERE code_type = 'MCC' AND code = '5310'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5311: Department Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452111', '452112'],
    'sic', ARRAY['5311', '5331']
)
WHERE code_type = 'MCC' AND code = '5311'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5331: Variety Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452990', '452112'],
    'sic', ARRAY['5331', '5399']
)
WHERE code_type = 'MCC' AND code = '5331'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5462: Bakeries
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['311811', '311812', '445291'],
    'sic', ARRAY['5461', '2051']
)
WHERE code_type = 'MCC' AND code = '5462'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5531: Auto Supply Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441310'],
    'sic', ARRAY['5531', '5013']
)
WHERE code_type = 'MCC' AND code = '5531'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5532: Automotive Tire Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441320'],
    'sic', ARRAY['5531', '5014']
)
WHERE code_type = 'MCC' AND code = '5532'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5542: Automated Fuel Dispensers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['447110', '447190'],
    'sic', ARRAY['5541', '5542']
)
WHERE code_type = 'MCC' AND code = '5542'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5571: Motorcycle Shops and Dealers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441221', '441228'],
    'sic', ARRAY['5571', '5551']
)
WHERE code_type = 'MCC' AND code = '5571'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5651: Family Clothing Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['448140', '448150'],
    'sic', ARRAY['5651', '5621']
)
WHERE code_type = 'MCC' AND code = '5651'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5661: Shoe Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['448210'],
    'sic', ARRAY['5661', '5699']
)
WHERE code_type = 'MCC' AND code = '5661'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5681: Furriers and Fur Shops
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['448150', '315210'],
    'sic', ARRAY['5681', '5699']
)
WHERE code_type = 'MCC' AND code = '5681'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5697: Tailors, Alterations
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['812320', '315210'],
    'sic', ARRAY['5699', '7211']
)
WHERE code_type = 'MCC' AND code = '5697'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5698: Wig and Toupee Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['446120', '812112'],
    'sic', ARRAY['5699', '7231']
)
WHERE code_type = 'MCC' AND code = '5698'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5699: Miscellaneous Apparel and Accessory Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['448150', '448190'],
    'sic', ARRAY['5699', '5621']
)
WHERE code_type = 'MCC' AND code = '5699'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5712: Furniture Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['442110', '442210'],
    'sic', ARRAY['5712', '5021']
)
WHERE code_type = 'MCC' AND code = '5712'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5713: Floor Covering Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['442210'],
    'sic', ARRAY['5713', '5719']
)
WHERE code_type = 'MCC' AND code = '5713'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5714: Drapery, Window Covering, and Upholstery Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['442210', '314120'],
    'sic', ARRAY['5714', '5719']
)
WHERE code_type = 'MCC' AND code = '5714'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5719: Miscellaneous Home Furnishing Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['442290', '442110'],
    'sic', ARRAY['5719', '5712']
)
WHERE code_type = 'MCC' AND code = '5719'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5722: Household Appliance Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['443111', '443112'],
    'sic', ARRAY['5722', '5063']
)
WHERE code_type = 'MCC' AND code = '5722'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5811: Caterers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722320', '722310'],
    'sic', ARRAY['5811', '5812']
)
WHERE code_type = 'MCC' AND code = '5811'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5812: Eating Places, Restaurants
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722511', '722513', '722514', '722515'],
    'sic', ARRAY['5812', '5813']
)
WHERE code_type = 'MCC' AND code = '5812'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5814: Fast Food Restaurants
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722513', '722514'],
    'sic', ARRAY['5812', '5813']
)
WHERE code_type = 'MCC' AND code = '5814'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5912: Drug Stores, Pharmacies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['446110', '446191'],
    'sic', ARRAY['5912', '5122']
)
WHERE code_type = 'MCC' AND code = '5912'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5932: Used Merchandise Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453310'],
    'sic', ARRAY['5932', '5933']
)
WHERE code_type = 'MCC' AND code = '5932'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5933: Pawn Shops
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522298', '453310'],
    'sic', ARRAY['5933', '6099']
)
WHERE code_type = 'MCC' AND code = '5933'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5935: Wrecking and Salvage Yards
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['562111', '562112'],
    'sic', ARRAY['5935', '5013']
)
WHERE code_type = 'MCC' AND code = '5935'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5941: Sporting Goods Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451110', '451120'],
    'sic', ARRAY['5941', '5943']
)
WHERE code_type = 'MCC' AND code = '5941'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5942: Book Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451211', '451220'],
    'sic', ARRAY['5942', '5735']
)
WHERE code_type = 'MCC' AND code = '5942'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5943: Office, School, and Stationery Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453210', '453220'],
    'sic', ARRAY['5943', '5111']
)
WHERE code_type = 'MCC' AND code = '5943'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5944: Jewelry Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['448310'],
    'sic', ARRAY['5944', '5094']
)
WHERE code_type = 'MCC' AND code = '5944'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5945: Hobby, Toy, and Game Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451120', '451130'],
    'sic', ARRAY['5945', '5943']
)
WHERE code_type = 'MCC' AND code = '5945'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5946: Camera and Photographic Supply Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['443142'],
    'sic', ARRAY['5946', '5048']
)
WHERE code_type = 'MCC' AND code = '5946'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5947: Gift, Card, Novelty, and Souvenir Shops
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453220', '453310'],
    'sic', ARRAY['5947', '5949']
)
WHERE code_type = 'MCC' AND code = '5947'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5948: Luggage and Leather Goods Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['448320'],
    'sic', ARRAY['5948', '5699']
)
WHERE code_type = 'MCC' AND code = '5948'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5949: Sewing, Needlework, and Piece Goods Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451130', '424310'],
    'sic', ARRAY['5949', '5131']
)
WHERE code_type = 'MCC' AND code = '5949'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5961: Direct Marketing - Inbound Telemarketing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['561422', '561421'],
    'sic', ARRAY['5961', '7389']
)
WHERE code_type = 'MCC' AND code = '5961'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5962: Direct Marketing - Outbound Telemarketing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['561422', '561421'],
    'sic', ARRAY['5962', '7389']
)
WHERE code_type = 'MCC' AND code = '5962'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5963: Direct Marketing - Subscription
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['454110', '454210'],
    'sic', ARRAY['5963', '5961']
)
WHERE code_type = 'MCC' AND code = '5963'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5970: Artist Supply Stores, Craft Shops
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451130', '453220'],
    'sic', ARRAY['5970', '5949']
)
WHERE code_type = 'MCC' AND code = '5970'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5971: Art Dealers and Galleries
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453920', '712110'],
    'sic', ARRAY['5971', '5999']
)
WHERE code_type = 'MCC' AND code = '5971'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5972: Stamp and Coin Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453220', '453310'],
    'sic', ARRAY['5972', '5949']
)
WHERE code_type = 'MCC' AND code = '5972'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5973: Religious Goods Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453220', '453310'],
    'sic', ARRAY['5973', '5999']
)
WHERE code_type = 'MCC' AND code = '5973'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5974: Rubber Stamp Store
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453210', '339950'],
    'sic', ARRAY['5974', '5943']
)
WHERE code_type = 'MCC' AND code = '5974'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5975: Hearing Aids - Sales and Service
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['446199', '621320'],
    'sic', ARRAY['5975', '5047']
)
WHERE code_type = 'MCC' AND code = '5975'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5976: Orthopedic Goods - Prosthetic Devices
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['446199', '339113'],
    'sic', ARRAY['5976', '5047']
)
WHERE code_type = 'MCC' AND code = '5976'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5977: Cosmetic Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['446120', '325620'],
    'sic', ARRAY['5977', '5122']
)
WHERE code_type = 'MCC' AND code = '5977'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5978: Typewriter Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['443142', '423430'],
    'sic', ARRAY['5978', '5045']
)
WHERE code_type = 'MCC' AND code = '5978'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5983: Fuel Dealers (Non-Automotive)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['454312', '454313'],
    'sic', ARRAY['5983', '5984']
)
WHERE code_type = 'MCC' AND code = '5983'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5992: Florists
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453110'],
    'sic', ARRAY['5992', '5261']
)
WHERE code_type = 'MCC' AND code = '5992'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5993: Cigar Stores and Stands
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453991', '312230'],
    'sic', ARRAY['5993', '5994']
)
WHERE code_type = 'MCC' AND code = '5993'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5994: News Dealers and Newsstands
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['451212'],
    'sic', ARRAY['5994', '5999']
)
WHERE code_type = 'MCC' AND code = '5994'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5995: Pet Shops, Pet Food, and Supplies Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453910'],
    'sic', ARRAY['5999', '5995']
)
WHERE code_type = 'MCC' AND code = '5995'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5996: Swimming Pools - Sales, Service, Supplies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['444190', '713940'],
    'sic', ARRAY['5996', '5999']
)
WHERE code_type = 'MCC' AND code = '5996'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5997: Electric Razor Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['443111', '446199'],
    'sic', ARRAY['5997', '5722']
)
WHERE code_type = 'MCC' AND code = '5997'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5998: Tent and Awning Shops
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['314910', '314999'],
    'sic', ARRAY['5998', '2394']
)
WHERE code_type = 'MCC' AND code = '5998'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- MCC 5999: Miscellaneous and Specialty Retail Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453990', '452990'],
    'sic', ARRAY['5999', '5399']
)
WHERE code_type = 'MCC' AND code = '5999'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- =====================================================
-- Part 2: Additional NAICS Crosswalks (30 codes)
-- =====================================================

-- NAICS 238150: Glass and Glazing Contractors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1793', '1799'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'NAICS' AND code = '238150'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 238160: Roofing Contractors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1761', '1799'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'NAICS' AND code = '238160'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 238170: Siding Contractors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1761', '1799'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'NAICS' AND code = '238170'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 238210: Electrical Contractors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1731', '1799'],
    'mcc', ARRAY['4900', '5046']
)
WHERE code_type = 'NAICS' AND code = '238210'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 238220: Plumbing, Heating, and Air-Conditioning Contractors
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['1711', '1799'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'NAICS' AND code = '238220'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311113: Soybean Processing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2074', '2075'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311113'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311119: Other Oilseed Processing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2074', '2075'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311119'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311211: Flour Milling
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2041', '2043'],
    'mcc', ARRAY['5411', '5462']
)
WHERE code_type = 'NAICS' AND code = '311211'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311212: Rice Milling
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2044', '2041'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311212'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311213: Malt Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2082', '2083'],
    'mcc', ARRAY['5182', '5921']
)
WHERE code_type = 'NAICS' AND code = '311213'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311221: Wet Corn Milling
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2046', '2041'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311221'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311225: Fats and Oils Refining and Blending
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2074', '2075'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311225'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311230: Breakfast Cereal Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2043', '2041'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311230'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311311: Sugarcane Mills
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2061', '2062'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311311'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311312: Cane Sugar Refining
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2062', '2061'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311312'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311313: Beet Sugar Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2063', '2062'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311313'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311320: Chocolate and Confectionery Manufacturing from Cacao Beans
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2066', '2067'],
    'mcc', ARRAY['5441', '5499']
)
WHERE code_type = 'NAICS' AND code = '311320'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311330: Confectionery Manufacturing from Purchased Chocolate
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2067', '2066'],
    'mcc', ARRAY['5441', '5499']
)
WHERE code_type = 'NAICS' AND code = '311330'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311340: Nonchocolate Confectionery Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2067', '2066'],
    'mcc', ARRAY['5441', '5499']
)
WHERE code_type = 'NAICS' AND code = '311340'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311410: Frozen Fruit, Juice, and Vegetable Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2037', '2038'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311410'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311411: Frozen Specialty Food Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2037', '2038'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311411'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311421: Fruit and Vegetable Canning
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2033', '2037'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311421'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311422: Specialty Canning
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2033', '2037'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311422'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311423: Dried and Dehydrated Food Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2034', '2037'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311423'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311511: Fluid Milk Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2026', '2021'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311511'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311512: Creamery Butter Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2021', '2026'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311512'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311513: Cheese Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2022', '2021'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311513'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311514: Dry, Condensed, and Evaporated Dairy Product Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2023', '2021'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311514'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311520: Ice Cream and Frozen Dessert Manufacturing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2024', '2021'],
    'mcc', ARRAY['5441', '5499']
)
WHERE code_type = 'NAICS' AND code = '311520'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 311615: Poultry Processing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['2015', '2011'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'NAICS' AND code = '311615'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 452990: All Other General Merchandise Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5399', '5311'],
    'mcc', ARRAY['5310', '5331', '5399']
)
WHERE code_type = 'NAICS' AND code = '452990'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 484122: General Freight Trucking, Long-Distance, Less Than Truckload
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4213', '4214'],
    'mcc', ARRAY['4215', '4784']
)
WHERE code_type = 'NAICS' AND code = '484122'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 484230: Specialized Freight (except Used Goods) Trucking, Long-Distance
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4213', '4214'],
    'mcc', ARRAY['4215', '4784']
)
WHERE code_type = 'NAICS' AND code = '484230'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 492210: Local Messengers and Local Delivery
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4215', '4213'],
    'mcc', ARRAY['4215', '4784']
)
WHERE code_type = 'NAICS' AND code = '492210'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 485210: Interurban and Rural Bus Transportation
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['4131', '4111'],
    'mcc', ARRAY['4131', '4111']
)
WHERE code_type = 'NAICS' AND code = '485210'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 522220: Sales Financing
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['6141', '6153'],
    'mcc', ARRAY['6010', '6011']
)
WHERE code_type = 'NAICS' AND code = '522220'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 541612: Human Resources Consulting Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['7361', '7389'],
    'mcc', ARRAY['7372', '7399']
)
WHERE code_type = 'NAICS' AND code = '541612'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 541621: Testing Laboratories
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8734', '8731'],
    'mcc', ARRAY['8734', '5047']
)
WHERE code_type = 'NAICS' AND code = '541621'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 541199: All Other Legal Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8111', '7389'],
    'mcc', ARRAY['8111', '7399']
)
WHERE code_type = 'NAICS' AND code = '541199'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 541370: Surveying and Mapping (except Geophysical) Services
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8712', '8711'],
    'mcc', ARRAY['7372', '7399']
)
WHERE code_type = 'NAICS' AND code = '541370'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 611620: Sports and Recreation Instruction
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8299', '7911'],
    'mcc', ARRAY['7941', '7999']
)
WHERE code_type = 'NAICS' AND code = '611620'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 611630: Language Schools
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8299', '8244'],
    'mcc', ARRAY['8244', '8299']
)
WHERE code_type = 'NAICS' AND code = '611630'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 611691: Exam Preparation and Tutoring
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8299', '8244'],
    'mcc', ARRAY['8244', '8299']
)
WHERE code_type = 'NAICS' AND code = '611691'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 611692: Automobile Driving Schools
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8299', '8244'],
    'mcc', ARRAY['8244', '8299']
)
WHERE code_type = 'NAICS' AND code = '611692'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- NAICS 611699: All Other Miscellaneous Schools and Instruction
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8299', '8244'],
    'mcc', ARRAY['8244', '8299']
)
WHERE code_type = 'NAICS' AND code = '611699'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- =====================================================
-- Part 3: Additional SIC Crosswalks (32 codes)
-- =====================================================

-- SIC 1711: Plumbing, Heating, and Air-Conditioning
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238220', '238210'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1711'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1721: Painting and Paper Hanging
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238320', '238330'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1721'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1731: Electrical Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238210'],
    'mcc', ARRAY['4900', '5046']
)
WHERE code_type = 'SIC' AND code = '1731'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1741: Masonry, Stone Setting, and Other Stone Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238140', '238150'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1741'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1742: Plastering, Drywall, Acoustical, and Insulation Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238310', '238320'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1742'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1751: Carpentry Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238130', '238310'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1751'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1752: Floor Laying and Other Floor Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238330', '238310'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1752'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1761: Roofing, Siding, and Sheet Metal Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238160', '238170'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1761'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1771: Concrete Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238110', '238120'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1771'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1791: Structural Steel Erection
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238120', '238110'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1791'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1793: Glass and Glazing Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238150'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1793'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1794: Excavation Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238910', '237110'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1794'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1795: Wrecking and Demolition Work
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238910', '562111'],
    'mcc', ARRAY['5935', '5039']
)
WHERE code_type = 'SIC' AND code = '1795'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1796: Installation or Erection of Building Equipment
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238290', '238210'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1796'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 1799: Special Trade Contractors, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['238990', '238390'],
    'mcc', ARRAY['5039', '5046']
)
WHERE code_type = 'SIC' AND code = '1799'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 2061: Raw Cane Sugar
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['311311', '311312'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'SIC' AND code = '2061'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 2062: Cane Sugar Refining
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['311312', '311311'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'SIC' AND code = '2062'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 2082: Malt Beverages
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['312120', '312111'],
    'mcc', ARRAY['5182', '5921']
)
WHERE code_type = 'SIC' AND code = '2082'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 4119: Local Passenger Transportation, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485210', '485310'],
    'mcc', ARRAY['4119', '4121']
)
WHERE code_type = 'SIC' AND code = '4119'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 4215: Courier Services, Except by Air
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['492210', '492110'],
    'mcc', ARRAY['4215', '4784']
)
WHERE code_type = 'SIC' AND code = '4215'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 5047: Medical, Dental, Ophthalmic, and Hospital Equipment and Supplies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['423450', '339112', '621111'],
    'mcc', ARRAY['5047', '8011']
)
WHERE code_type = 'SIC' AND code = '5047'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 5048: Optical Goods, Photographic Equipment, and Supplies
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['443142', '423430'],
    'mcc', ARRAY['5946', '5045']
)
WHERE code_type = 'SIC' AND code = '5048'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 5451: Dairy Products Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['445210', '445220'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'SIC' AND code = '5451'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 5499: Miscellaneous Food Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['445299', '445291'],
    'mcc', ARRAY['5411', '5499']
)
WHERE code_type = 'SIC' AND code = '5499'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 5511: Motor Vehicle Dealers (New and Used)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441110', '441120'],
    'mcc', ARRAY['5511', '5521']
)
WHERE code_type = 'SIC' AND code = '5511'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 5521: Motor Vehicle Dealers (Used Only)
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441120', '453310'],
    'mcc', ARRAY['5521', '5511']
)
WHERE code_type = 'SIC' AND code = '5521'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 5531: Auto and Home Supply Stores
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['441310', '443111'],
    'mcc', ARRAY['5531', '5532']
)
WHERE code_type = 'SIC' AND code = '5531'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 5541: Gasoline Service Stations
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['447110', '447190'],
    'mcc', ARRAY['5541', '5542']
)
WHERE code_type = 'SIC' AND code = '5541'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 6099: Functions Related to Depository Banking
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '522298'],
    'mcc', ARRAY['6010', '6011']
)
WHERE code_type = 'SIC' AND code = '6099'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 6162: Mortgage Bankers and Loan Correspondents
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522310', '522220'],
    'mcc', ARRAY['6010', '6011']
)
WHERE code_type = 'SIC' AND code = '6162'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 6163: Loan Brokers
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522310', '522220'],
    'mcc', ARRAY['6010', '6011']
)
WHERE code_type = 'SIC' AND code = '6163'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 6515: Operators of Residential Mobile Home Sites
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['531110', '531120'],
    'mcc', ARRAY['6513', '6514']
)
WHERE code_type = 'SIC' AND code = '6515'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 6517: Lessors of Railroad Property
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['531120', '531110'],
    'mcc', ARRAY['6513', '6514']
)
WHERE code_type = 'SIC' AND code = '6517'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- SIC 8299: Schools and Educational Services, Not Elsewhere Classified
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['611699', '611691'],
    'mcc', ARRAY['8244', '8299']
)
WHERE code_type = 'SIC' AND code = '8299'
  AND (crosswalk_data = '{}'::jsonb OR crosswalk_data IS NULL);

-- =====================================================
-- Verification Query
-- =====================================================

-- Verify crosswalk coverage after supplement
SELECT 
    'Crosswalk Coverage After Supplement' AS metric,
    COUNT(*) AS codes_with_crosswalks,
    (SELECT COUNT(*) FROM code_metadata WHERE is_active = true) AS total_codes,
    ROUND(COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 0), 2) AS coverage_percentage,
    CASE 
        WHEN COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 0) >= 50.0 THEN '✅ PASS - 50%+ coverage'
        ELSE '❌ FAIL - Below 50% coverage'
    END AS status
FROM code_metadata
WHERE is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'naics' OR crosswalk_data ? 'sic' OR crosswalk_data ? 'mcc');

