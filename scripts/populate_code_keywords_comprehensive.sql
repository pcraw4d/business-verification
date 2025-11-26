-- =====================================================
-- Comprehensive Code Keywords Population Script
-- Goal: 15-20 keywords per classification code
-- Current: ~4 keywords per code
-- Target: 15-20 keywords per code (3-5x increase)
-- =====================================================

-- This script will:
-- 1. Extract comprehensive keywords from code descriptions
-- 2. Add industry-specific keywords
-- 3. Add synonyms and variations
-- 4. Add related terms and context keywords
-- 5. Add code-type-specific keywords
-- 6. Ensure 15-20 keywords per code

-- =====================================================
-- Step 1: Enhanced keyword extraction from descriptions
-- =====================================================

-- Extract all meaningful words from descriptions (improved extraction)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    trim(lower(keyword)) as keyword,
    CASE 
        -- Higher relevance for longer, more specific words
        WHEN length(trim(keyword)) >= 8 THEN 0.90
        WHEN length(trim(keyword)) >= 6 THEN 0.85
        ELSE 0.80
    END as relevance_score,
    'exact' as match_type
FROM classification_codes cc,
LATERAL unnest(
    string_to_array(
        regexp_replace(
            lower(cc.description),
            '[^a-z0-9\s-]',
            ' ',
            'g'
        ),
        ' '
    )
) as keyword
WHERE cc.is_active = true
  AND trim(keyword) != ''
  AND length(trim(keyword)) >= 3
  AND trim(keyword) NOT IN (
    -- Common stop words
    'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for',
    'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were', 'be',
    'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
    'would', 'should', 'could', 'may', 'might', 'must', 'can', 'this',
    'that', 'these', 'those', 'it', 'its', 'they', 'them', 'their', 'our',
    'your', 'my', 'his', 'her', 'he', 'she', 'we', 'you', 'i', 'me', 'us',
    'services', 'service', 'products', 'product', 'business', 'company',
    'companies', 'other', 'etc', 'including', 'excluding', 'related', 'not',
    'all', 'any', 'each', 'every', 'some', 'many', 'most', 'more', 'less',
    'than', 'then', 'when', 'where', 'what', 'which', 'who', 'whom', 'how',
    'why', 'while', 'during', 'through', 'until', 'unless', 'since', 'ago',
    'about', 'above', 'below', 'between', 'among', 'within', 'without',
    'into', 'onto', 'upon', 'toward', 'towards', 'against', 'beside',
    'besides', 'except', 'beyond', 'across', 'around', 'along', 'after',
    'before', 'behind', 'beneath', 'beside', 'besides', 'between', 'beyond'
  )
  AND trim(keyword) !~ '^\d+$'  -- Exclude pure numbers
  AND trim(keyword) !~ '^[a-z]$'  -- Exclude single letters
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Extract compound words and phrases (2-word combinations)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
WITH word_arrays AS (
    SELECT 
        cc.id as code_id,
        regexp_split_to_array(
            regexp_replace(
                lower(cc.description),
                '[^a-z0-9\s-]',
                ' ',
                'g'
            ),
            '\s+'
        ) as words
    FROM classification_codes cc
    WHERE cc.is_active = true
)
SELECT DISTINCT
    wa.code_id,
    trim(lower(wa.words[i] || ' ' || wa.words[i+1])) as keyword,
    0.95 as relevance_score,
    'exact' as match_type
FROM word_arrays wa
CROSS JOIN generate_series(1, array_length(wa.words, 1) - 1) as i
WHERE wa.words[i] IS NOT NULL
  AND wa.words[i+1] IS NOT NULL
  AND length(trim(wa.words[i] || ' ' || wa.words[i+1])) >= 6
  AND length(trim(wa.words[i] || ' ' || wa.words[i+1])) <= 50
  AND trim(wa.words[i] || ' ' || wa.words[i+1]) !~ '^\d+'
  AND trim(wa.words[i]) NOT IN ('the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were')
  AND trim(wa.words[i+1]) NOT IN ('the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- =====================================================
-- Step 2: Industry-specific comprehensive keywords
-- =====================================================

-- Technology & Software (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('software'), ('technology'), ('tech'), ('digital'), ('computer'), ('programming'),
    ('development'), ('coding'), ('app'), ('application'), ('platform'), ('system'),
    ('saas'), ('cloud'), ('api'), ('web'), ('mobile'), ('ios'), ('android'),
    ('devops'), ('infrastructure'), ('server'), ('database'), ('backend'), ('frontend'),
    ('code'), ('developer'), ('engineer'), ('it'), ('information'), ('data'),
    ('automation'), ('integration'), ('solution'), ('service'), ('tool'), ('framework'),
    ('algorithm'), ('analytics'), ('ai'), ('artificial'), ('intelligence'), ('machine'),
    ('learning'), ('ml'), ('neural'), ('network'), ('blockchain'), ('crypto'),
    ('cybersecurity'), ('security'), ('encryption'), ('authentication'), ('authorization'),
    ('network'), ('networking'), ('protocol'), ('http'), ('https'), ('tcp'), ('ip'),
    ('server'), ('hosting'), ('domain'), ('dns'), ('ssl'), ('tls'), ('certificate')
) AS tech_keywords(keyword)
WHERE i.name IN ('Technology', 'Software Development')
  AND (lower(cc.description) LIKE '%' || tech_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || tech_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Healthcare & Medical (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('healthcare'), ('medical'), ('health'), ('hospital'), ('clinic'), ('doctor'),
    ('physician'), ('patient'), ('medicine'), ('pharmaceutical'), ('drug'), ('therapy'),
    ('treatment'), ('diagnosis'), ('surgery'), ('nursing'), ('wellness'), ('fitness'),
    ('dental'), ('vision'), ('mental'), ('psychology'), ('counseling'), ('rehabilitation'),
    ('emergency'), ('ambulance'), ('pharmacy'), ('laboratory'), ('radiology'), ('imaging'),
    ('cardiology'), ('oncology'), ('neurology'), ('orthopedic'), ('pediatric'), ('geriatric'),
    ('surgical'), ('anesthesia'), ('pathology'), ('dermatology'), ('gynecology'), ('urology'),
    ('pulmonology'), ('gastroenterology'), ('endocrinology'), ('rheumatology'), ('immunology'),
    ('vaccine'), ('prescription'), ('medication'), ('diagnostic'), ('therapeutic'), ('clinical')
) AS health_keywords(keyword)
WHERE i.name IN ('Healthcare', 'Medical Technology')
  AND (lower(cc.description) LIKE '%' || health_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || health_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Finance & Banking (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('finance'), ('financial'), ('banking'), ('bank'), ('investment'), ('investing'),
    ('trading'), ('stock'), ('market'), ('fund'), ('capital'), ('credit'), ('loan'),
    ('mortgage'), ('insurance'), ('payment'), ('transaction'), ('card'), ('debit'),
    ('accounting'), ('accountant'), ('audit'), ('tax'), ('account'), ('savings'),
    ('checking'), ('wealth'), ('asset'), ('portfolio'), ('broker'), ('advisor'),
    ('securities'), ('bond'), ('equity'), ('derivative'), ('futures'), ('options'),
    ('forex'), ('currency'), ('exchange'), ('trading'), ('trader'), ('analyst'),
    ('fintech'), ('cryptocurrency'), ('bitcoin'), ('blockchain'), ('digital'), ('wallet'),
    ('payment'), ('processing'), ('gateway'), ('merchant'), ('acquirer'), ('issuer')
) AS finance_keywords(keyword)
WHERE i.name IN ('Finance', 'Fintech', 'Insurance')
  AND (lower(cc.description) LIKE '%' || finance_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || finance_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Retail & E-commerce (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('retail'), ('store'), ('shop'), ('shopping'), ('merchant'), ('commerce'),
    ('ecommerce'), ('online'), ('marketplace'), ('product'), ('goods'), ('sale'),
    ('selling'), ('buying'), ('purchase'), ('customer'), ('consumer'), ('grocery'),
    ('supermarket'), ('department'), ('boutique'), ('outlet'), ('warehouse'), ('wholesale'),
    ('fashion'), ('apparel'), ('clothing'), ('accessory'), ('jewelry'), ('footwear'),
    ('electronics'), ('appliance'), ('furniture'), ('home'), ('garden'), ('automotive'),
    ('beauty'), ('cosmetic'), ('personal'), ('care'), ('health'), ('wellness'),
    ('sports'), ('recreation'), ('outdoor'), ('camping'), ('hunting'), ('fishing')
) AS retail_keywords(keyword)
WHERE i.name IN ('Retail', 'E-commerce', 'Fashion & Apparel')
  AND (lower(cc.description) LIKE '%' || retail_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || retail_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Manufacturing & Industrial (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('manufacturing'), ('production'), ('factory'), ('industrial'), ('machinery'),
    ('equipment'), ('assembly'), ('fabrication'), ('processing'), ('machining'),
    ('welding'), ('cutting'), ('molding'), ('casting'), ('packaging'), ('quality'),
    ('control'), ('automation'), ('robotics'), ('cnc'), ('metal'), ('plastic'),
    ('wood'), ('textile'), ('chemical'), ('material'), ('component'), ('part'),
    ('automotive'), ('aerospace'), ('defense'), ('electronics'), ('semiconductor'),
    ('pharmaceutical'), ('food'), ('beverage'), ('packaging'), ('printing'), ('publishing')
) AS manufacturing_keywords(keyword)
WHERE i.name = 'Manufacturing'
  AND (lower(cc.description) LIKE '%' || manufacturing_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || manufacturing_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Food & Beverage (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('restaurant'), ('food'), ('dining'), ('cafe'), ('bistro'), ('bar'), ('grill'),
    ('catering'), ('bakery'), ('pizza'), ('burger'), ('cuisine'), ('menu'), ('chef'),
    ('kitchen'), ('beverage'), ('drink'), ('coffee'), ('tea'), ('alcohol'), ('wine'),
    ('beer'), ('liquor'), ('dining'), ('meal'), ('breakfast'), ('lunch'), ('dinner'),
    ('fast'), ('casual'), ('fine'), ('gourmet'), ('organic'), ('vegan'), ('vegetarian'),
    ('seafood'), ('steakhouse'), ('italian'), ('mexican'), ('asian'), ('chinese'), ('japanese')
) AS food_keywords(keyword)
WHERE i.name = 'Food & Beverage'
  AND (lower(cc.description) LIKE '%' || food_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || food_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Transportation & Logistics (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('transportation'), ('transport'), ('logistics'), ('shipping'), ('delivery'),
    ('freight'), ('cargo'), ('trucking'), ('truck'), ('vehicle'), ('fleet'),
    ('warehouse'), ('warehousing'), ('distribution'), ('supply'), ('chain'), ('storage'),
    ('courier'), ('express'), ('parcel'), ('package'), ('mail'), ('postal'),
    ('airline'), ('aviation'), ('airport'), ('aircraft'), ('railroad'), ('railway'),
    ('shipping'), ('maritime'), ('port'), ('harbor'), ('trucking'), ('trucking')
) AS transport_keywords(keyword)
WHERE i.name = 'Transportation'
  AND (lower(cc.description) LIKE '%' || transport_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || transport_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Education (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('education'), ('school'), ('university'), ('college'), ('learning'), ('teaching'),
    ('training'), ('course'), ('class'), ('student'), ('teacher'), ('instructor'),
    ('academic'), ('curriculum'), ('degree'), ('certificate'), ('diploma'), ('tutoring'),
    ('online'), ('distance'), ('e-learning'), ('edtech'), ('educational'), ('institution'),
    ('elementary'), ('secondary'), ('high'), ('middle'), ('preschool'), ('kindergarten'),
    ('vocational'), ('technical'), ('professional'), ('graduate'), ('undergraduate')
) AS education_keywords(keyword)
WHERE i.name = 'Education'
  AND (lower(cc.description) LIKE '%' || education_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || education_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Real Estate (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('real'), ('estate'), ('property'), ('rental'), ('leasing'), ('landlord'),
    ('tenant'), ('residential'), ('commercial'), ('office'), ('retail'), ('industrial'),
    ('apartment'), ('condo'), ('house'), ('home'), ('building'), ('construction'),
    ('development'), ('broker'), ('agent'), ('realtor'), ('mortgage'), ('loan'),
    ('appraisal'), ('inspection'), ('title'), ('escrow'), ('closing'), ('transaction')
) AS realestate_keywords(keyword)
WHERE i.name = 'Real Estate and Rental and Leasing'
  AND (lower(cc.description) LIKE '%' || realestate_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || realestate_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Construction (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('construction'), ('building'), ('contractor'), ('contracting'), ('remodeling'),
    ('renovation'), ('renovation'), ('demolition'), ('excavation'), ('foundation'),
    ('framing'), ('roofing'), ('plumbing'), ('electrical'), ('hvac'), ('heating'),
    ('cooling'), ('ventilation'), ('carpentry'), ('masonry'), ('concrete'), ('steel'),
    ('architecture'), ('engineering'), ('design'), ('planning'), ('project'), ('management')
) AS construction_keywords(keyword)
WHERE i.name = 'Construction'
  AND (lower(cc.description) LIKE '%' || construction_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || construction_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Agriculture (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('agriculture'), ('farming'), ('farm'), ('crop'), ('livestock'), ('cattle'),
    ('poultry'), ('dairy'), ('grain'), ('wheat'), ('corn'), ('soybean'),
    ('cotton'), ('fruit'), ('vegetable'), ('organic'), ('sustainable'), ('irrigation'),
    ('harvest'), ('cultivation'), ('ranching'), ('fishing'), ('forestry'), ('timber')
) AS agriculture_keywords(keyword)
WHERE i.name = 'Agriculture, Forestry, Fishing and Hunting'
  AND (lower(cc.description) LIKE '%' || agriculture_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || agriculture_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Mining & Energy (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('mining'), ('quarrying'), ('extraction'), ('oil'), ('gas'), ('petroleum'),
    ('coal'), ('mineral'), ('metal'), ('copper'), ('gold'), ('silver'),
    ('drilling'), ('exploration'), ('production'), ('refining'), ('processing'),
    ('energy'), ('power'), ('electric'), ('utility'), ('solar'), ('wind'),
    ('renewable'), ('nuclear'), ('hydroelectric'), ('generation'), ('transmission')
) AS mining_keywords(keyword)
WHERE i.name IN ('Mining, Quarrying, and Oil and Gas Extraction', 'Utilities')
  AND (lower(cc.description) LIKE '%' || mining_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || mining_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Professional Services (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('legal'), ('law'), ('attorney'), ('lawyer'), ('law firm'), ('litigation'),
    ('accounting'), ('accountant'), ('cpa'), ('audit'), ('bookkeeping'), ('tax'),
    ('consulting'), ('consultant'), ('advisory'), ('strategy'), ('management'),
    ('engineering'), ('architect'), ('architecture'), ('design'), ('planning'),
    ('marketing'), ('advertising'), ('public'), ('relations'), ('pr'), ('media')
) AS professional_keywords(keyword)
WHERE i.name = 'Professional, Scientific, and Technical Services'
  AND (lower(cc.description) LIKE '%' || professional_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || professional_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- =====================================================
-- Step 3: Comprehensive synonym expansion
-- =====================================================

-- Technology synonyms (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT ck.code_id, synonym, ck.relevance_score * 0.85, 'synonym'
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
CROSS JOIN (VALUES
    ('software', 'app'), ('software', 'application'), ('software', 'program'), ('software', 'code'),
    ('technology', 'tech'), ('technology', 'digital'), ('technology', 'it'), ('technology', 'information'),
    ('development', 'dev'), ('development', 'programming'), ('development', 'coding'),
    ('platform', 'system'), ('platform', 'framework'), ('platform', 'solution'),
    ('cloud', 'hosting'), ('cloud', 'infrastructure'), ('cloud', 'saas'),
    ('data', 'information'), ('data', 'database'), ('data', 'analytics'),
    ('computer', 'pc'), ('computer', 'desktop'), ('computer', 'laptop'),
    ('server', 'host'), ('server', 'hosting'), ('server', 'infrastructure'),
    ('network', 'networking'), ('network', 'connectivity'), ('network', 'internet'),
    ('security', 'cybersecurity'), ('security', 'protection'), ('security', 'safety')
) AS tech_synonyms(original, synonym)
WHERE ck.keyword = tech_synonyms.original 
  AND ck.match_type = 'exact'
  AND i.name IN ('Technology', 'Software Development')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Healthcare synonyms (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT ck.code_id, synonym, ck.relevance_score * 0.85, 'synonym'
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
CROSS JOIN (VALUES
    ('healthcare', 'medical'), ('healthcare', 'health'), ('healthcare', 'hospital'),
    ('doctor', 'physician'), ('doctor', 'md'), ('doctor', 'practitioner'),
    ('medicine', 'drug'), ('medicine', 'pharmaceutical'), ('medicine', 'medication'),
    ('therapy', 'treatment'), ('therapy', 'rehabilitation'), ('therapy', 'counseling'),
    ('surgery', 'surgical'), ('surgery', 'operation'), ('surgery', 'procedure'),
    ('patient', 'client'), ('patient', 'individual'), ('patient', 'person'),
    ('clinic', 'medical'), ('clinic', 'facility'), ('clinic', 'center')
) AS health_synonyms(original, synonym)
WHERE ck.keyword = health_synonyms.original 
  AND ck.match_type = 'exact'
  AND i.name IN ('Healthcare', 'Medical Technology')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Finance synonyms (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT ck.code_id, synonym, ck.relevance_score * 0.85, 'synonym'
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
CROSS JOIN (VALUES
    ('finance', 'financial'), ('finance', 'banking'), ('finance', 'investment'),
    ('bank', 'banking'), ('bank', 'financial'), ('bank', 'institution'),
    ('payment', 'transaction'), ('payment', 'processing'), ('payment', 'card'),
    ('loan', 'credit'), ('loan', 'lending'), ('loan', 'mortgage'),
    ('insurance', 'coverage'), ('insurance', 'policy'), ('insurance', 'protection'),
    ('investment', 'investing'), ('investment', 'portfolio'), ('investment', 'asset'),
    ('trading', 'trade'), ('trading', 'exchange'), ('trading', 'market'),
    ('account', 'accounting'), ('account', 'banking'), ('account', 'financial')
) AS finance_synonyms(original, synonym)
WHERE ck.keyword = finance_synonyms.original 
  AND ck.match_type = 'exact'
  AND i.name IN ('Finance', 'Fintech', 'Insurance')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- Retail synonyms (expanded)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT ck.code_id, synonym, ck.relevance_score * 0.85, 'synonym'
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
CROSS JOIN (VALUES
    ('retail', 'store'), ('retail', 'shop'), ('retail', 'merchant'),
    ('ecommerce', 'online'), ('ecommerce', 'digital'), ('ecommerce', 'internet'),
    ('product', 'goods'), ('product', 'item'), ('product', 'merchandise'),
    ('sale', 'selling'), ('sale', 'transaction'), ('sale', 'purchase'),
    ('customer', 'consumer'), ('customer', 'buyer'), ('customer', 'shopper'),
    ('shopping', 'buying'), ('shopping', 'purchasing'), ('shopping', 'retail')
) AS retail_synonyms(original, synonym)
WHERE ck.keyword = retail_synonyms.original 
  AND ck.match_type = 'exact'
  AND i.name IN ('Retail', 'E-commerce', 'Fashion & Apparel')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- =====================================================
-- Step 4: Code-type-specific keywords
-- =====================================================

-- MCC (Merchant Category Code) specific keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.80, 'partial'
FROM classification_codes cc
CROSS JOIN (VALUES
    ('merchant'), ('payment'), ('transaction'), ('card'), ('credit'), ('debit'),
    ('point'), ('sale'), ('pos'), ('terminal'), ('processing'), ('gateway'),
    ('processor'), ('acquirer'), ('issuer'), ('network'), ('visa'), ('mastercard'),
    ('amex'), ('discover'), ('chargeback'), ('refund'), ('authorization'),
    ('settlement'), ('clearing'), ('reconciliation'), ('fee'), ('charge')
) AS mcc_keywords(keyword)
WHERE cc.code_type = 'MCC'
  AND cc.is_active = true
  AND (lower(cc.description) LIKE '%' || mcc_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || mcc_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- NAICS (North American Industry Classification System) specific keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.85, 'partial'
FROM classification_codes cc
CROSS JOIN (VALUES
    ('industry'), ('sector'), ('establishment'), ('enterprise'), ('business'),
    ('activity'), ('operation'), ('production'), ('manufacturing'), ('service'),
    ('classification'), ('category'), ('group'), ('division'), ('subsector'),
    ('naics'), ('north'), ('american'), ('classification'), ('system')
) AS naics_keywords(keyword)
WHERE cc.code_type = 'NAICS'
  AND cc.is_active = true
  AND (lower(cc.description) LIKE '%' || naics_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || naics_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- SIC (Standard Industrial Classification) specific keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.80, 'partial'
FROM classification_codes cc
CROSS JOIN (VALUES
    ('standard'), ('industry'), ('classification'), ('division'), ('major'),
    ('group'), ('establishment'), ('activity'), ('code'), ('category'),
    ('sic'), ('standard'), ('industrial'), ('classification'), ('system')
) AS sic_keywords(keyword)
WHERE cc.code_type = 'SIC'
  AND cc.is_active = true
  AND (lower(cc.description) LIKE '%' || sic_keywords.keyword || '%'
       OR lower(cc.code) LIKE '%' || sic_keywords.keyword || '%')
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- =====================================================
-- Step 5: Common business terms (applied to all codes)
-- =====================================================

-- General business keywords that apply to many codes
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.75, 'partial'
FROM classification_codes cc
CROSS JOIN (VALUES
    ('business'), ('company'), ('corporation'), ('firm'), ('organization'),
    ('enterprise'), ('establishment'), ('operation'), ('service'), ('provider'),
    ('vendor'), ('supplier'), ('contractor'), ('consultant'), ('agency'),
    ('management'), ('administration'), ('support'), ('maintenance'), ('repair'),
    ('sales'), ('marketing'), ('customer'), ('client'), ('professional')
) AS business_keywords(keyword)
WHERE cc.is_active = true
  AND lower(cc.description) LIKE '%' || business_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score);

-- =====================================================
-- Step 6: Fill gaps to ensure 15-20 keywords per code
-- =====================================================

-- For codes with fewer than 15 keywords, add related terms from industry keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    ik.keyword,
    0.75 as relevance_score,
    'partial' as match_type
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
INNER JOIN industry_keywords ik ON ik.industry_id = i.id
WHERE cc.is_active = true
  AND cc.id NOT IN (
    -- Exclude codes that already have 15+ keywords
    SELECT code_id
    FROM code_keywords
    GROUP BY code_id
    HAVING COUNT(*) >= 15
  )
  AND NOT EXISTS (
    -- Don't add if keyword already exists for this code
    SELECT 1 FROM code_keywords ck
    WHERE ck.code_id = cc.id AND ck.keyword = ik.keyword
  )
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 7: Final statistics and verification
-- =====================================================

DO $$
DECLARE
    total_keywords INTEGER;
    total_codes INTEGER;
    codes_with_keywords INTEGER;
    avg_keywords_per_code NUMERIC;
    codes_below_15 INTEGER;
    codes_15_to_20 INTEGER;
    codes_above_20 INTEGER;
BEGIN
    -- Overall statistics
    SELECT COUNT(*) INTO total_keywords FROM code_keywords;
    SELECT COUNT(*) INTO total_codes FROM classification_codes WHERE is_active = true;
    SELECT COUNT(DISTINCT code_id) INTO codes_with_keywords FROM code_keywords;
    SELECT ROUND(COUNT(*)::NUMERIC / NULLIF(COUNT(DISTINCT code_id), 0), 2) INTO avg_keywords_per_code FROM code_keywords;
    
    -- Distribution by keyword count
    SELECT COUNT(*) INTO codes_below_15
    FROM (
        SELECT code_id, COUNT(*) as kw_count
        FROM code_keywords
        GROUP BY code_id
        HAVING COUNT(*) < 15
    ) sub;
    
    SELECT COUNT(*) INTO codes_15_to_20
    FROM (
        SELECT code_id, COUNT(*) as kw_count
        FROM code_keywords
        GROUP BY code_id
        HAVING COUNT(*) >= 15 AND COUNT(*) <= 20
    ) sub;
    
    SELECT COUNT(*) INTO codes_above_20
    FROM (
        SELECT code_id, COUNT(*) as kw_count
        FROM code_keywords
        GROUP BY code_id
        HAVING COUNT(*) > 20
    ) sub;
    
    RAISE NOTICE '====================================================';
    RAISE NOTICE 'CODE KEYWORDS POPULATION SUMMARY';
    RAISE NOTICE '====================================================';
    RAISE NOTICE 'Total Keywords: %', total_keywords;
    RAISE NOTICE 'Total Classification Codes: %', total_codes;
    RAISE NOTICE 'Codes with Keywords: %', codes_with_keywords;
    RAISE NOTICE 'Average Keywords per Code: %', avg_keywords_per_code;
    RAISE NOTICE '';
    RAISE NOTICE 'Distribution:';
    RAISE NOTICE '  Codes with < 15 keywords: %', codes_below_15;
    RAISE NOTICE '  Codes with 15-20 keywords: %', codes_15_to_20;
    RAISE NOTICE '  Codes with > 20 keywords: %', codes_above_20;
    RAISE NOTICE '====================================================';
END $$;

-- Display breakdown by code type
SELECT 
    cc.code_type,
    COUNT(DISTINCT ck.code_id) as codes_with_keywords,
    COUNT(ck.id) as total_keywords,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance,
    ROUND(COUNT(ck.id)::NUMERIC / NULLIF(COUNT(DISTINCT ck.code_id), 0), 2) as avg_keywords_per_code
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE cc.is_active = true
GROUP BY cc.code_type
ORDER BY cc.code_type;

-- Display breakdown by match type
SELECT 
    match_type,
    COUNT(*) as count,
    ROUND(AVG(relevance_score), 2) as avg_relevance,
    ROUND(MIN(relevance_score), 2) as min_relevance,
    ROUND(MAX(relevance_score), 2) as max_relevance
FROM code_keywords
GROUP BY match_type
ORDER BY match_type;

