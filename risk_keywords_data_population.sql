-- ====================================================================
-- Risk Keywords Data Population Script
-- ====================================================================
-- This script populates the risk_keywords table with comprehensive
-- risk detection data for merchant verification and compliance.
-- 
-- Categories covered:
-- 1. Illegal Activities (Critical Risk)
-- 2. Prohibited by Card Brands (High Risk) 
-- 3. High-Risk Industries (Medium-High Risk)
-- 4. Trade-Based Money Laundering (TBML)
-- 5. Fraud Indicators (Medium Risk)
-- 6. Sanctions and OFAC Violations
-- ====================================================================

-- Clear existing data (for fresh population)
TRUNCATE TABLE risk_keywords RESTART IDENTITY CASCADE;

-- ====================================================================
-- 1. ILLEGAL ACTIVITIES (Critical Risk)
-- ====================================================================

-- Drug Trafficking and Illegal Substances
INSERT INTO risk_keywords (
    keyword, risk_category, risk_severity, description,
    mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
    detection_patterns, synonyms
) VALUES 
-- Drug trafficking keywords
('drug trafficking', 'illegal', 'critical', 'Illegal drug trafficking activities', 
 ARRAY['7995'], ARRAY['621999'], ARRAY['7999'], 
 ARRAY['Visa', 'Mastercard', 'Amex'], 
 ARRAY['(?i)(drug|trafficking|dealing|distribution)', '(?i)(cocaine|heroin|marijuana|methamphetamine)'],
 ARRAY['drug dealing', 'drug distribution', 'narcotics trafficking', 'substance trafficking']),

('cocaine', 'illegal', 'critical', 'Cocaine-related illegal activities',
 ARRAY['7995'], ARRAY['621999'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(cocaine|coke|blow|snow)'],
 ARRAY['coke', 'blow', 'snow', 'white powder']),

('heroin', 'illegal', 'critical', 'Heroin-related illegal activities',
 ARRAY['7995'], ARRAY['621999'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(heroin|dope|smack|horse)'],
 ARRAY['dope', 'smack', 'horse', 'brown sugar']),

('marijuana', 'illegal', 'critical', 'Marijuana trafficking (where illegal)',
 ARRAY['7995'], ARRAY['621999'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(marijuana|weed|cannabis|pot|hash)'],
 ARRAY['weed', 'cannabis', 'pot', 'hash', 'ganja', 'mary jane']),

-- Weapons and Arms
('weapons trafficking', 'illegal', 'critical', 'Illegal weapons trafficking',
 ARRAY['5999'], ARRAY['332999'], ARRAY['3489'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(weapon|firearm|gun|rifle|pistol)', '(?i)(trafficking|smuggling|illegal)'],
 ARRAY['arms trafficking', 'gun smuggling', 'firearm dealing', 'weapon dealing']),

('illegal firearms', 'illegal', 'critical', 'Illegal firearms sales',
 ARRAY['5999'], ARRAY['332999'], ARRAY['3489'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(illegal|unlicensed|straw|purchase)', '(?i)(firearm|gun|weapon)'],
 ARRAY['unlicensed firearms', 'straw purchase', 'illegal guns']),

-- Human Trafficking
('human trafficking', 'illegal', 'critical', 'Human trafficking activities',
 ARRAY['7999'], ARRAY['624190'], ARRAY['7299'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(human|trafficking|sex|slave)', '(?i)(forced|labor|prostitution)'],
 ARRAY['sex trafficking', 'forced labor', 'modern slavery', 'human smuggling']),

('sex trafficking', 'illegal', 'critical', 'Sex trafficking activities',
 ARRAY['7999'], ARRAY['624190'], ARRAY['7299'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(sex|trafficking|escort|prostitution)', '(?i)(forced|coerced|underage)'],
 ARRAY['forced prostitution', 'escort trafficking', 'sex slavery']),

-- Money Laundering
('money laundering', 'illegal', 'critical', 'Money laundering activities',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(money|laundering|dirty|money)', '(?i)(clean|wash|structure)'],
 ARRAY['dirty money', 'money washing', 'financial crime', 'laundering']),

('terrorist financing', 'illegal', 'critical', 'Terrorist financing activities',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(terrorist|terrorism|financing|funding)', '(?i)(extremist|radical)'],
 ARRAY['terrorism funding', 'extremist financing', 'terrorist funding']);

-- ====================================================================
-- 2. PROHIBITED BY CARD BRANDS (High Risk)
-- ====================================================================

-- Adult Entertainment
INSERT INTO risk_keywords (
    keyword, risk_category, risk_severity, description,
    mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
    detection_patterns, synonyms
) VALUES 
('adult entertainment', 'prohibited', 'high', 'Adult entertainment services',
 ARRAY['7273', '7841'], ARRAY['713290'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(adult|porn|sex|strip|escort)', '(?i)(entertainment|club|service)'],
 ARRAY['pornography', 'strip club', 'escort service', 'adult club']),

('pornography', 'prohibited', 'high', 'Pornographic content and services',
 ARRAY['7273', '7841'], ARRAY['713290'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(porn|xxx|adult|sex)', '(?i)(video|content|site|service)'],
 ARRAY['xxx', 'adult content', 'sex site', 'porn site']),

-- Gambling
('gambling', 'prohibited', 'high', 'Gambling and betting services',
 ARRAY['7995'], ARRAY['713290'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(gambling|betting|casino|poker)', '(?i)(online|sports|lottery)'],
 ARRAY['betting', 'casino', 'poker', 'sports betting', 'lottery']),

('online gambling', 'prohibited', 'high', 'Online gambling platforms',
 ARRAY['7995'], ARRAY['713290'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(online|internet|web)', '(?i)(gambling|betting|casino)'],
 ARRAY['internet gambling', 'web betting', 'online casino']),

-- Cryptocurrency (High Risk)
('cryptocurrency', 'prohibited', 'high', 'Cryptocurrency trading and services',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(crypto|bitcoin|ethereum|digital)', '(?i)(currency|coin|token|trading)'],
 ARRAY['bitcoin', 'ethereum', 'digital currency', 'crypto trading']),

('bitcoin', 'prohibited', 'high', 'Bitcoin trading and services',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(bitcoin|btc)', '(?i)(trading|exchange|mining|wallet)'],
 ARRAY['btc', 'bitcoin trading', 'crypto exchange']),

-- Tobacco and Alcohol
('tobacco', 'prohibited', 'high', 'Tobacco products and sales',
 ARRAY['5993'], ARRAY['453991'], ARRAY['5993'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(tobacco|cigarette|cigar|smoke)', '(?i)(sale|retail|wholesale)'],
 ARRAY['cigarettes', 'cigars', 'smoking products', 'tobacco retail']),

('alcohol', 'prohibited', 'high', 'Alcohol sales and distribution',
 ARRAY['5921'], ARRAY['445310'], ARRAY['5921'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(alcohol|beer|wine|spirits)', '(?i)(sale|retail|distribution)'],
 ARRAY['beer', 'wine', 'spirits', 'liquor', 'alcoholic beverages']);

-- ====================================================================
-- 3. HIGH-RISK INDUSTRIES (Medium-High Risk)
-- ====================================================================

INSERT INTO risk_keywords (
    keyword, risk_category, risk_severity, description,
    mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
    detection_patterns, synonyms
) VALUES 
-- Money Services
('money services', 'high_risk', 'medium', 'Money transfer and financial services',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(money|transfer|remittance|wire)', '(?i)(service|business|center)'],
 ARRAY['money transfer', 'remittance', 'wire transfer', 'money order']),

('check cashing', 'high_risk', 'medium', 'Check cashing services',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(check|cashing)', '(?i)(service|center|business)'],
 ARRAY['check cashing service', 'cash check', 'check exchange']),

-- Prepaid Cards and Gift Cards
('prepaid cards', 'high_risk', 'medium', 'Prepaid card services',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(prepaid|gift|card)', '(?i)(service|business|retail)'],
 ARRAY['gift cards', 'prepaid services', 'card services']),

('gift cards', 'high_risk', 'medium', 'Gift card services',
 ARRAY['5999'], ARRAY['453220'], ARRAY['5999'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(gift|card)', '(?i)(service|retail|business)'],
 ARRAY['gift card service', 'card retail', 'gift services']),

-- Cryptocurrency Exchanges
('crypto exchange', 'high_risk', 'medium', 'Cryptocurrency exchange services',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(crypto|exchange)', '(?i)(trading|platform|service)'],
 ARRAY['cryptocurrency exchange', 'crypto trading', 'digital exchange']),

-- High-Risk Merchants
('dating services', 'high_risk', 'medium', 'Online dating and relationship services',
 ARRAY['7273'], ARRAY['713290'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(dating|match|relationship)', '(?i)(service|site|app|platform)'],
 ARRAY['matchmaking', 'relationship service', 'dating site']),

('travel services', 'high_risk', 'medium', 'Travel booking and services',
 ARRAY['4722'], ARRAY['561510'], ARRAY['4722'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(travel|booking|reservation)', '(?i)(service|agency|platform)'],
 ARRAY['travel booking', 'reservation service', 'travel agency']);

-- ====================================================================
-- 4. TRADE-BASED MONEY LAUNDERING (TBML)
-- ====================================================================

INSERT INTO risk_keywords (
    keyword, risk_category, risk_severity, description,
    mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
    detection_patterns, synonyms
) VALUES 
-- Shell Companies
('shell company', 'tbml', 'high', 'Shell company indicators',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(shell|company|corporation)', '(?i)(front|dummy|paper)'],
 ARRAY['front company', 'dummy corporation', 'paper company', 'shell corporation']),

('front company', 'tbml', 'high', 'Front company indicators',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(front|company)', '(?i)(shell|dummy|paper)'],
 ARRAY['shell company', 'dummy company', 'paper company']),

-- Trade Finance
('trade finance', 'tbml', 'medium', 'Trade finance and import/export',
 ARRAY['6012'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(trade|finance)', '(?i)(import|export|letter|credit)'],
 ARRAY['import finance', 'export finance', 'trade credit', 'letter of credit']),

('import export', 'tbml', 'medium', 'Import/export businesses',
 ARRAY['5999'], ARRAY['423990'], ARRAY['5999'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(import|export)', '(?i)(business|company|trading)'],
 ARRAY['international trade', 'global trading', 'cross-border trade']),

-- Commodity Trading
('commodity trading', 'tbml', 'medium', 'Commodity trading businesses',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(commodity|trading)', '(?i)(precious|metal|oil|gas)'],
 ARRAY['precious metals', 'oil trading', 'gas trading', 'commodity exchange']),

('precious metals', 'tbml', 'medium', 'Precious metals trading',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(precious|metal)', '(?i)(gold|silver|platinum|trading)'],
 ARRAY['gold trading', 'silver trading', 'metal exchange', 'bullion trading']),

-- Complex Trade Structures
('complex trade', 'tbml', 'medium', 'Complex trade structures',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(complex|trade)', '(?i)(structure|arrangement|scheme)'],
 ARRAY['trade structure', 'complex arrangement', 'trade scheme']);

-- ====================================================================
-- 5. FRAUD INDICATORS (Medium Risk)
-- ====================================================================

INSERT INTO risk_keywords (
    keyword, risk_category, risk_severity, description,
    mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
    detection_patterns, synonyms
) VALUES 
-- Fake Business Names
('fake business', 'fraud', 'medium', 'Fake business name indicators',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(fake|false|dummy)', '(?i)(business|company|name)'],
 ARRAY['false business', 'dummy business', 'fake company']),

('stolen identity', 'fraud', 'medium', 'Stolen identity indicators',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(stolen|identity)', '(?i)(theft|fraud|fake)'],
 ARRAY['identity theft', 'identity fraud', 'stolen identity']),

-- Rapid Business Changes
('rapid changes', 'fraud', 'medium', 'Rapid business changes indicators',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(rapid|quick|frequent)', '(?i)(change|turnover|modification)'],
 ARRAY['quick changes', 'frequent turnover', 'rapid modification']),

('high turnover', 'fraud', 'medium', 'High business turnover indicators',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(high|turnover)', '(?i)(business|company|staff)'],
 ARRAY['business turnover', 'company turnover', 'staff turnover']),

-- Unusual Transaction Patterns
('unusual patterns', 'fraud', 'medium', 'Unusual transaction patterns',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(unusual|suspicious)', '(?i)(pattern|transaction|activity)'],
 ARRAY['suspicious patterns', 'unusual activity', 'suspicious transactions']),

-- Geographic Risk Factors
('high risk country', 'fraud', 'medium', 'High-risk geographic locations',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(high|risk)', '(?i)(country|region|location)'],
 ARRAY['sanctioned country', 'embargoed region', 'high-risk location']);

-- ====================================================================
-- 6. SANCTIONS AND OFAC VIOLATIONS
-- ====================================================================

INSERT INTO risk_keywords (
    keyword, risk_category, risk_severity, description,
    mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
    detection_patterns, synonyms
) VALUES 
-- OFAC Violations
('ofac violation', 'sanctions', 'critical', 'OFAC sanctions violations',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(ofac|sanctions)', '(?i)(violation|breach|non-compliance)'],
 ARRAY['sanctions violation', 'ofac breach', 'sanctions non-compliance']),

('sanctions list', 'sanctions', 'critical', 'Sanctions list entities',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(sanctions|list)', '(?i)(entity|person|organization)'],
 ARRAY['sanctioned entity', 'blocked person', 'prohibited entity']),

-- Terrorist Organizations
('terrorist organization', 'sanctions', 'critical', 'Terrorist organization indicators',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(terrorist|terrorism)', '(?i)(organization|group|network)'],
 ARRAY['terrorist group', 'terrorist network', 'extremist organization']),

-- Embargoed Countries
('embargoed country', 'sanctions', 'critical', 'Embargoed country indicators',
 ARRAY['5999'], ARRAY['523110'], ARRAY['6099'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(embargoed|sanctioned)', '(?i)(country|nation|state)'],
 ARRAY['sanctioned country', 'embargoed nation', 'prohibited country']);

-- ====================================================================
-- 7. ADDITIONAL HIGH-RISK MCC CODES
-- ====================================================================

-- Insert specific prohibited MCC codes
INSERT INTO risk_keywords (
    keyword, risk_category, risk_severity, description,
    mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
    detection_patterns, synonyms
) VALUES 
-- Prohibited MCC 7995 - Gambling
('mcc 7995', 'prohibited', 'high', 'Gambling establishments (MCC 7995)',
 ARRAY['7995'], ARRAY['713290'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(gambling|casino|betting)'],
 ARRAY['gambling mcc', 'casino mcc', 'betting mcc']),

-- Prohibited MCC 7273 - Dating Services
('mcc 7273', 'prohibited', 'high', 'Dating services (MCC 7273)',
 ARRAY['7273'], ARRAY['713290'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard'],
 ARRAY['(?i)(dating|match|relationship)'],
 ARRAY['dating mcc', 'matchmaking mcc', 'relationship mcc']),

-- Prohibited MCC 7841 - Video Entertainment
('mcc 7841', 'prohibited', 'high', 'Video entertainment (MCC 7841)',
 ARRAY['7841'], ARRAY['713290'], ARRAY['7999'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(video|entertainment|adult)'],
 ARRAY['video mcc', 'entertainment mcc', 'adult mcc']),

-- Prohibited MCC 5993 - Cigar Stores
('mcc 5993', 'prohibited', 'high', 'Cigar stores and stands (MCC 5993)',
 ARRAY['5993'], ARRAY['453991'], ARRAY['5993'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(cigar|tobacco|smoke)'],
 ARRAY['cigar mcc', 'tobacco mcc', 'smoke mcc']),

-- Prohibited MCC 5921 - Package Stores
('mcc 5921', 'prohibited', 'high', 'Package stores - beer, wine, liquor (MCC 5921)',
 ARRAY['5921'], ARRAY['445310'], ARRAY['5921'],
 ARRAY['Visa', 'Mastercard', 'Amex'],
 ARRAY['(?i)(package|store|beer|wine|liquor)'],
 ARRAY['package store mcc', 'liquor mcc', 'alcohol mcc']);

-- ====================================================================
-- 8. CREATE INDEXES FOR PERFORMANCE
-- ====================================================================

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_risk_keywords_category ON risk_keywords(risk_category);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_severity ON risk_keywords(risk_severity);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_keyword ON risk_keywords USING gin(to_tsvector('english', keyword));
CREATE INDEX IF NOT EXISTS idx_risk_keywords_mcc_codes ON risk_keywords USING gin(mcc_codes);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_synonyms ON risk_keywords USING gin(synonyms);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_active ON risk_keywords(is_active) WHERE is_active = true;

-- ====================================================================
-- 9. VALIDATION QUERIES
-- ====================================================================

-- Verify data population
SELECT 
    risk_category,
    risk_severity,
    COUNT(*) as keyword_count
FROM risk_keywords 
WHERE is_active = true
GROUP BY risk_category, risk_severity
ORDER BY risk_category, risk_severity;

-- Check for any data integrity issues
SELECT 
    'Total Keywords' as metric,
    COUNT(*) as count
FROM risk_keywords
UNION ALL
SELECT 
    'Active Keywords' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE is_active = true
UNION ALL
SELECT 
    'Critical Risk Keywords' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE risk_severity = 'critical' AND is_active = true
UNION ALL
SELECT 
    'High Risk Keywords' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE risk_severity = 'high' AND is_active = true;

-- ====================================================================
-- END OF SCRIPT
-- ====================================================================
