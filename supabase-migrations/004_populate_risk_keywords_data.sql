-- =====================================================
-- Risk Keywords Data Population Migration
-- Supabase Implementation - Task 1.4.2
-- =====================================================
-- 
-- This script populates the risk_keywords table with comprehensive
-- risk detection data covering all major risk categories:
-- - Card brand prohibited activities (Visa, Mastercard, Amex)
-- - Illegal business activities
-- - High-risk industries
-- - Trade-based money laundering indicators
-- - Sanctions and OFAC-related keywords
-- - Fraud detection patterns
--
-- Author: KYB Platform Development Team
-- Date: January 19, 2025
-- Version: 1.0
-- =====================================================

-- =====================================================
-- 1. ILLEGAL ACTIVITIES (Critical Risk)
-- =====================================================

-- Drug trafficking and illegal substances
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('drug trafficking', 'illegal', 'critical', 'Illegal drug trafficking and distribution', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"drug.*traffick", "traffick.*drug", "narcotic.*smuggl"}', '{"drug dealing", "narcotics trafficking", "drug smuggling", "substance trafficking"}', true),
('cocaine', 'illegal', 'critical', 'Cocaine and cocaine-related activities', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"cocaine", "coke", "crack", "snow"}', '{"coke", "crack", "snow", "blow"}', true),
('heroin', 'illegal', 'critical', 'Heroin and heroin-related activities', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"heroin", "dope", "horse", "smack"}', '{"dope", "horse", "smack", "junk"}', true),
('marijuana', 'illegal', 'critical', 'Marijuana and cannabis-related activities (where illegal)', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"marijuana", "cannabis", "weed", "pot", "hash"}', '{"weed", "pot", "hash", "grass", "dope"}', true),
('methamphetamine', 'illegal', 'critical', 'Methamphetamine and related activities', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"methamphetamine", "meth", "crystal", "ice"}', '{"meth", "crystal", "ice", "crank"}', true),
('ecstasy', 'illegal', 'critical', 'Ecstasy and MDMA-related activities', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"ecstasy", "mdma", "molly", "e"}', '{"mdma", "molly", "e", "x"}', true),
('lsd', 'illegal', 'critical', 'LSD and hallucinogenic substances', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"lsd", "acid", "hallucinogen"}', '{"acid", "hallucinogen", "trip"}', true);

-- Weapons and firearms
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('weapons trafficking', 'illegal', 'critical', 'Illegal weapons trafficking and distribution', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"weapon.*traffick", "arms.*smuggl", "gun.*traffick"}', '{"arms trafficking", "weapon smuggling", "gun running"}', true),
('illegal firearms', 'illegal', 'critical', 'Illegal firearms and weapons sales', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"illegal.*firearm", "unregistered.*gun", "stolen.*weapon"}', '{"illegal guns", "unregistered firearms", "stolen weapons"}', true),
('explosives', 'illegal', 'critical', 'Explosives and bomb-making materials', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"explosive", "bomb.*mak", "dynamite", "tnt"}', '{"bomb making", "dynamite", "tnt", "c4"}', true),
('ammunition', 'illegal', 'critical', 'Illegal ammunition sales and distribution', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"ammunition", "ammo", "bullet", "cartridge"}', '{"ammo", "bullets", "cartridges", "rounds"}', true);

-- Human trafficking
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('human trafficking', 'illegal', 'critical', 'Human trafficking and modern slavery', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"human.*traffick", "sex.*traffick", "labor.*traffick"}', '{"sex trafficking", "labor trafficking", "modern slavery", "forced labor"}', true),
('sex trafficking', 'illegal', 'critical', 'Sex trafficking and exploitation', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"sex.*traffick", "prostitution.*ring", "escort.*service"}', '{"prostitution ring", "escort service", "sex trade"}', true),
('forced labor', 'illegal', 'critical', 'Forced labor and slavery', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"forced.*labor", "slave.*labor", "bonded.*labor"}', '{"slave labor", "bonded labor", "indentured servitude"}', true),
('child exploitation', 'illegal', 'critical', 'Child exploitation and abuse', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"child.*exploit", "minor.*abuse", "underage.*sex"}', '{"minor abuse", "underage sex", "child abuse"}', true);

-- Money laundering and terrorist financing
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('money laundering', 'illegal', 'critical', 'Money laundering and financial crimes', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"money.*launder", "cash.*wash", "dirty.*money"}', '{"cash washing", "dirty money", "financial crime"}', true),
('terrorist financing', 'illegal', 'critical', 'Terrorist financing and support', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"terrorist.*financ", "terror.*fund", "extremist.*fund"}', '{"terror funding", "extremist funding", "terrorist support"}', true),
('hawala', 'illegal', 'critical', 'Hawala and informal money transfer systems', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"hawala", "hundi", "informal.*transfer"}', '{"hundi", "informal transfer", "underground banking"}', true);

-- =====================================================
-- 2. PROHIBITED BY CARD BRANDS (High Risk)
-- =====================================================

-- Adult entertainment
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('adult entertainment', 'prohibited', 'high', 'Adult entertainment and pornography', '{"7273", "7841"}', '{"713290"}', '{"7841"}', '{"Visa", "Mastercard", "American Express"}', '{"adult.*entertain", "pornograph", "xxx", "adult.*content"}', '{"pornography", "xxx", "adult content", "sex industry"}', true),
('strip club', 'prohibited', 'high', 'Strip clubs and adult entertainment venues', '{"7273"}', '{"713290"}', '{"7841"}', '{"Visa", "Mastercard", "American Express"}', '{"strip.*club", "gentlemen.*club", "adult.*club"}', '{"gentlemen club", "adult club", "nude bar"}', true),
('escort service', 'prohibited', 'high', 'Escort services and adult services', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"escort.*service", "adult.*service", "companion.*service"}', '{"adult service", "companion service", "dating service"}', true),
('webcam modeling', 'prohibited', 'high', 'Webcam modeling and adult web services', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"webcam.*model", "cam.*girl", "adult.*webcam"}', '{"cam girl", "adult webcam", "live sex"}', true);

-- Gambling
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('online gambling', 'prohibited', 'high', 'Online gambling and betting', '{"7995"}', '{"713290"}', '{"7995"}', '{"Visa", "Mastercard", "American Express"}', '{"online.*gambl", "internet.*bet", "online.*casino"}', '{"internet betting", "online casino", "digital gambling"}', true),
('sports betting', 'prohibited', 'high', 'Sports betting and wagering', '{"7995"}', '{"713290"}', '{"7995"}', '{"Visa", "Mastercard", "American Express"}', '{"sports.*bet", "sport.*wager", "betting.*sport"}', '{"sport wager", "betting sport", "sports gambling"}', true),
('casino', 'prohibited', 'high', 'Casino and gambling establishments', '{"7995"}', '{"713290"}', '{"7995"}', '{"Visa", "Mastercard", "American Express"}', '{"casino", "gambling.*hall", "betting.*house"}', '{"gambling hall", "betting house", "gaming establishment"}', true),
('poker', 'prohibited', 'high', 'Poker and card games', '{"7995"}', '{"713290"}', '{"7995"}', '{"Visa", "Mastercard", "American Express"}', '{"poker", "card.*game", "poker.*room"}', '{"card game", "poker room", "poker tournament"}', true),
('lottery', 'prohibited', 'high', 'Lottery and numbers games', '{"7995"}', '{"713290"}', '{"7995"}', '{"Visa", "Mastercard", "American Express"}', '{"lottery", "lotto", "numbers.*game"}', '{"lotto", "numbers game", "scratch off"}', true);

-- Cryptocurrency
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('cryptocurrency exchange', 'prohibited', 'high', 'Cryptocurrency exchanges and trading', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"crypto.*exchange", "bitcoin.*exchange", "digital.*currency.*exchange"}', '{"bitcoin exchange", "digital currency exchange", "crypto trading"}', true),
('bitcoin', 'prohibited', 'high', 'Bitcoin and cryptocurrency transactions', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"bitcoin", "btc", "cryptocurrency"}', '{"btc", "cryptocurrency", "digital currency"}', true),
('ethereum', 'prohibited', 'high', 'Ethereum and altcoin transactions', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"ethereum", "eth", "altcoin"}', '{"eth", "altcoin", "ether"}', true),
('crypto mining', 'prohibited', 'high', 'Cryptocurrency mining operations', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"crypto.*min", "bitcoin.*min", "mining.*pool"}', '{"bitcoin mining", "mining pool", "crypto farm"}', true);

-- Tobacco and alcohol
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('tobacco', 'prohibited', 'high', 'Tobacco products and sales', '{"5993"}', '{"312230"}', '{"5993"}', '{"Visa", "Mastercard", "American Express"}', '{"tobacco", "cigarette", "cigar", "smoking"}', '{"cigarette", "cigar", "smoking", "nicotine"}', true),
('alcohol', 'prohibited', 'high', 'Alcohol sales and distribution', '{"5921"}', '{"312120"}', '{"5921"}', '{"Visa", "Mastercard", "American Express"}', '{"alcohol", "liquor", "spirits", "wine"}', '{"liquor", "spirits", "wine", "beer"}', true),
('vaping', 'prohibited', 'high', 'Vaping and e-cigarette products', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"vaping", "e-cigarette", "vape.*pen", "electronic.*cigarette"}', '{"e-cigarette", "vape pen", "electronic cigarette", "e-cig"}', true);

-- Firearms and weapons
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('firearms', 'prohibited', 'high', 'Firearms and weapons sales', '{"5999"}', '{"332992"}', '{"5999"}', '{"Visa", "Mastercard", "American Express"}', '{"firearm", "gun.*shop", "weapon.*store", "ammunition.*store"}', '{"gun shop", "weapon store", "ammunition store", "arms dealer"}', true),
('weapons', 'prohibited', 'high', 'Weapons and military equipment', '{"5999"}', '{"332992"}', '{"5999"}', '{"Visa", "Mastercard", "American Express"}', '{"weapon", "military.*equipment", "tactical.*gear"}', '{"military equipment", "tactical gear", "combat gear"}', true);

-- =====================================================
-- 3. HIGH-RISK INDUSTRIES (Medium-High Risk)
-- =====================================================

-- Money services
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('money services', 'high_risk', 'medium', 'Money services and currency exchange', '{"6012"}', '{"523130"}', '{"6012"}', '{}', '{"money.*service", "currency.*exchange", "money.*transfer"}', '{"currency exchange", "money transfer", "remittance"}', true),
('check cashing', 'high_risk', 'medium', 'Check cashing services', '{"6012"}', '{"523130"}', '{"6012"}', '{}', '{"check.*cash", "cash.*check", "payday.*loan"}', '{"cash check", "payday loan", "quick cash"}', true),
('payday loans', 'high_risk', 'medium', 'Payday loans and short-term lending', '{"6012"}', '{"523130"}', '{"6012"}', '{}', '{"payday.*loan", "short.*term.*loan", "cash.*advance"}', '{"short term loan", "cash advance", "quick loan"}', true),
('money transfer', 'high_risk', 'medium', 'Money transfer and remittance services', '{"6012"}', '{"523130"}', '{"6012"}', '{}', '{"money.*transfer", "remittance", "wire.*transfer"}', '{"remittance", "wire transfer", "international transfer"}', true);

-- Prepaid cards and gift cards
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('prepaid cards', 'high_risk', 'medium', 'Prepaid cards and stored value products', '{}', '{}', '{}', '{}', '{"prepaid.*card", "stored.*value", "gift.*card"}', '{"stored value", "gift card", "prepaid debit"}', true),
('gift cards', 'high_risk', 'medium', 'Gift cards and prepaid products', '{}', '{}', '{}', '{}', '{"gift.*card", "prepaid.*gift", "store.*card"}', '{"prepaid gift", "store card", "retail card"}', true),
('virtual currency', 'high_risk', 'medium', 'Virtual currency and digital assets', '{}', '{}', '{}', '{}', '{"virtual.*currency", "digital.*asset", "token.*sale"}', '{"digital asset", "token sale", "virtual money"}', true);

-- High-risk merchants
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('dating service', 'high_risk', 'medium', 'Dating services and matchmaking', '{}', '{}', '{}', '{}', '{"dating.*service", "matchmaking", "online.*dating"}', '{"matchmaking", "online dating", "relationship service"}', true),
('travel agency', 'high_risk', 'medium', 'Travel agencies and booking services', '{}', '{}', '{}', '{}', '{"travel.*agency", "booking.*service", "vacation.*rental"}', '{"booking service", "vacation rental", "travel booking"}', true),
('telemarketing', 'high_risk', 'medium', 'Telemarketing and call center services', '{}', '{}', '{}', '{}', '{"telemarket", "call.*center", "telephone.*sales"}', '{"call center", "telephone sales", "phone sales"}', true);

-- =====================================================
-- 4. TRADE-BASED MONEY LAUNDERING (TBML)
-- =====================================================

-- Shell companies and front companies
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('shell company', 'tbml', 'high', 'Shell companies and front companies', '{}', '{}', '{}', '{}', '{"shell.*compan", "front.*compan", "paper.*compan"}', '{"front company", "paper company", "dummy company"}', true),
('offshore company', 'tbml', 'high', 'Offshore companies and tax havens', '{}', '{}', '{}', '{}', '{"offshore.*compan", "tax.*haven", "secrecy.*jurisdiction"}', '{"tax haven", "secrecy jurisdiction", "offshore entity"}', true),
('nominee director', 'tbml', 'high', 'Nominee directors and beneficial ownership', '{}', '{}', '{}', '{}', '{"nominee.*director", "beneficial.*owner", "straw.*man"}', '{"beneficial owner", "straw man", "nominee shareholder"}', true),
('bearer shares', 'tbml', 'high', 'Bearer shares and anonymous ownership', '{}', '{}', '{}', '{}', '{"bearer.*share", "anonymous.*owner", "unregistered.*share"}', '{"anonymous owner", "unregistered share", "bearer stock"}', true);

-- Trade finance
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('trade finance', 'tbml', 'high', 'Trade finance and import/export', '{}', '{}', '{}', '{}', '{"trade.*financ", "import.*export", "letter.*credit"}', '{"import export", "letter of credit", "trade credit"}', true),
('over-invoicing', 'tbml', 'high', 'Over-invoicing and trade manipulation', '{}', '{}', '{}', '{}', '{"over.*invoic", "under.*invoic", "false.*invoic"}', '{"under invoicing", "false invoicing", "trade manipulation"}', true),
('commodity trading', 'tbml', 'high', 'Commodity trading and precious metals', '{}', '{}', '{}', '{}', '{"commodit.*trad", "precious.*metal", "gold.*trad"}', '{"precious metal", "gold trading", "commodity exchange"}', true),
('diamond trading', 'tbml', 'high', 'Diamond and gemstone trading', '{}', '{}', '{}', '{}', '{"diamond.*trad", "gemstone.*trad", "precious.*stone"}', '{"gemstone trading", "precious stone", "diamond dealer"}', true);

-- Complex trade structures
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('circular trading', 'tbml', 'high', 'Circular trading and round-trip transactions', '{}', '{}', '{}', '{}', '{"circular.*trad", "round.*trip", "back.*to.*back"}', '{"round trip", "back to back", "circular transaction"}', true),
('layering', 'tbml', 'high', 'Layering and complex transaction structures', '{}', '{}', '{}', '{}', '{"layer", "complex.*structur", "multi.*step"}', '{"complex structure", "multi step", "transaction layering"}', true),
('smurfing', 'tbml', 'high', 'Smurfing and structuring transactions', '{}', '{}', '{}', '{}', '{"smurf", "structur.*transaction", "split.*transaction"}', '{"structuring transaction", "split transaction", "transaction splitting"}', true);

-- =====================================================
-- 5. SANCTIONS AND OFAC-RELATED KEYWORDS
-- =====================================================

-- Sanctions violations
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('sanctions violation', 'sanctions', 'critical', 'Sanctions violations and embargo breaches', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"sanction.*violat", "embargo.*breach", "prohibited.*trade"}', '{"embargo breach", "prohibited trade", "sanctions evasion"}', true),
('ofac violation', 'sanctions', 'critical', 'OFAC violations and sanctions evasion', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"ofac.*violat", "sanction.*evasion", "embargo.*evasion"}', '{"sanctions evasion", "embargo evasion", "ofac breach"}', true),
('embargo', 'sanctions', 'critical', 'Trade embargoes and restrictions', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"embargo", "trade.*restrict", "economic.*sanction"}', '{"trade restriction", "economic sanction", "trade ban"}', true);

-- High-risk countries and entities
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('iran', 'sanctions', 'critical', 'Iran-related transactions and entities', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"iran", "iranian", "persian"}', '{"iranian", "persian", "tehran"}', true),
('north korea', 'sanctions', 'critical', 'North Korea-related transactions and entities', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"north.*korea", "dprk", "pyongyang"}', '{"dprk", "pyongyang", "korean.*north"}', true),
('cuba', 'sanctions', 'critical', 'Cuba-related transactions and entities', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"cuba", "cuban", "havana"}', '{"cuban", "havana", "cuban.*entity"}', true),
('syria', 'sanctions', 'critical', 'Syria-related transactions and entities', '{}', '{}', '{}', '{"Visa", "Mastercard", "American Express"}', '{"syria", "syrian", "damascus"}', '{"syrian", "damascus", "syrian.*entity"}', true);

-- =====================================================
-- 6. FRAUD DETECTION PATTERNS
-- =====================================================

-- Identity fraud
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('identity theft', 'fraud', 'high', 'Identity theft and impersonation', '{}', '{}', '{}', '{}', '{"identity.*theft", "impersonat", "stolen.*identit"}', '{"impersonation", "stolen identity", "identity fraud"}', true),
('fake identity', 'fraud', 'high', 'Fake identities and synthetic fraud', '{}', '{}', '{}', '{}', '{"fake.*identit", "synthetic.*fraud", "false.*identit"}', '{"synthetic fraud", "false identity", "fabricated identity"}', true),
('stolen identity', 'fraud', 'high', 'Stolen identity and account takeover', '{}', '{}', '{}', '{}', '{"stolen.*identit", "account.*takeover", "identity.*fraud"}', '{"account takeover", "identity fraud", "stolen account"}', true);

-- Business fraud
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('fake business', 'fraud', 'high', 'Fake businesses and shell companies', '{}', '{}', '{}', '{}', '{"fake.*business", "fraudulent.*business", "sham.*business"}', '{"fraudulent business", "sham business", "bogus business"}', true),
('business fraud', 'fraud', 'high', 'Business fraud and corporate crime', '{}', '{}', '{}', '{}', '{"business.*fraud", "corporate.*fraud", "company.*fraud"}', '{"corporate fraud", "company fraud", "business crime"}', true),
('ponzi scheme', 'fraud', 'high', 'Ponzi schemes and investment fraud', '{}', '{}', '{}', '{}', '{"ponzi.*scheme", "pyramid.*scheme", "investment.*fraud"}', '{"pyramid scheme", "investment fraud", "securities fraud"}', true);

-- Payment fraud
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('credit card fraud', 'fraud', 'high', 'Credit card fraud and payment fraud', '{}', '{}', '{}', '{}', '{"credit.*card.*fraud", "payment.*fraud", "card.*fraud"}', '{"payment fraud", "card fraud", "transaction fraud"}', true),
('chargeback fraud', 'fraud', 'high', 'Chargeback fraud and friendly fraud', '{}', '{}', '{}', '{}', '{"chargeback.*fraud", "friendly.*fraud", "dispute.*fraud"}', '{"friendly fraud", "dispute fraud", "chargeback abuse"}', true),
('refund fraud', 'fraud', 'high', 'Refund fraud and return fraud', '{}', '{}', '{}', '{}', '{"refund.*fraud", "return.*fraud", "merchandise.*fraud"}', '{"return fraud", "merchandise fraud", "refund abuse"}', true);

-- =====================================================
-- 7. ADDITIONAL RISK PATTERNS
-- =====================================================

-- Geographic risk factors
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('high risk country', 'fraud', 'medium', 'High-risk countries and jurisdictions', '{}', '{}', '{}', '{}', '{"high.*risk.*countr", "risky.*jurisdiction", "offshore.*jurisdiction"}', '{"risky jurisdiction", "offshore jurisdiction", "tax haven"}', true),
('conflict zone', 'fraud', 'medium', 'Conflict zones and unstable regions', '{}', '{}', '{}', '{}', '{"conflict.*zone", "war.*zone", "unstable.*region"}', '{"war zone", "unstable region", "conflict area"}', true);

-- Rapid business changes
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
('rapid business change', 'fraud', 'medium', 'Rapid business changes and high turnover', '{}', '{}', '{}', '{}', '{"rapid.*change", "high.*turnover", "frequent.*change"}', '{"high turnover", "frequent change", "business instability"}', true),
('suspicious activity', 'fraud', 'medium', 'Suspicious business activity patterns', '{}', '{}', '{}', '{}', '{"suspicious.*activ", "unusual.*pattern", "anomalous.*behavior"}', '{"unusual pattern", "anomalous behavior", "suspicious behavior"}', true);

-- =====================================================
-- 8. CREATE INDEXES FOR PERFORMANCE
-- =====================================================

-- Create additional indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_risk_keywords_keyword_trgm ON risk_keywords USING gin(keyword gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_synonyms_trgm ON risk_keywords USING gin(synonyms gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_patterns_trgm ON risk_keywords USING gin(detection_patterns gin_trgm_ops);

-- =====================================================
-- 9. VALIDATION QUERIES
-- =====================================================

-- Verify data insertion
SELECT 
    risk_category,
    risk_severity,
    COUNT(*) as keyword_count
FROM risk_keywords 
WHERE is_active = true
GROUP BY risk_category, risk_severity
ORDER BY risk_category, risk_severity;

-- Check for duplicate keywords
SELECT keyword, COUNT(*) as count
FROM risk_keywords 
WHERE is_active = true
GROUP BY keyword
HAVING COUNT(*) > 1;

-- Verify card brand restrictions
SELECT 
    card_brand_restrictions,
    COUNT(*) as count
FROM risk_keywords 
WHERE is_active = true AND card_brand_restrictions IS NOT NULL
GROUP BY card_brand_restrictions;

-- =====================================================
-- MIGRATION COMPLETION
-- =====================================================

-- Log completion
INSERT INTO migration_log (migration_name, status, completed_at, notes) 
VALUES (
    '004_populate_risk_keywords_data', 
    'completed', 
    NOW(), 
    'Successfully populated risk_keywords table with comprehensive risk detection data covering all major risk categories'
);

-- Update table statistics
ANALYZE risk_keywords;
