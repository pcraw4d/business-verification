-- Add keywords for remaining industries that have no keyword coverage
-- This addresses the 23 industries without keywords identified in validation

-- Agriculture keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('agriculture', 1.0), ('farming', 1.0), ('crop', 1.0), ('livestock', 1.0),
    ('agricultural', 1.0), ('farm', 1.0), ('crops', 0.9), ('farming operation', 0.9),
    ('agricultural production', 0.9), ('crop production', 0.8), ('livestock production', 0.8),
    ('dairy farming', 0.8), ('poultry farming', 0.8), ('cattle farming', 0.8),
    ('grain farming', 0.7), ('vegetable farming', 0.7), ('fruit farming', 0.7),
    ('organic farming', 0.7), ('sustainable agriculture', 0.7), ('agricultural services', 0.8),
    ('farm equipment', 0.6), ('agricultural machinery', 0.6), ('irrigation', 0.6),
    ('soil management', 0.6), ('crop rotation', 0.6)
) AS k(keyword, weight)
WHERE i.name = 'Agriculture';

-- Bars & Pubs keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('bar', 1.0), ('pub', 1.0), ('tavern', 1.0), ('cocktail bar', 0.9),
    ('sports bar', 0.9), ('wine bar', 0.9), ('dive bar', 0.8), ('lounge', 0.8),
    ('nightclub', 0.8), ('bar and grill', 0.9), ('pub food', 0.8), ('bar food', 0.8),
    ('drinks', 0.8), ('alcohol', 0.8), ('beer', 0.8), ('wine', 0.8), ('spirits', 0.8),
    ('cocktails', 0.8), ('happy hour', 0.7), ('bar service', 0.7), ('bartender', 0.7),
    ('bar entertainment', 0.6), ('live music', 0.6), ('bar games', 0.6), ('bar atmosphere', 0.6)
) AS k(keyword, weight)
WHERE i.name = 'Bars & Pubs';

-- Cafes & Coffee Shops keywords (enhance existing)
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('cafe', 1.0), ('coffee shop', 1.0), ('coffee', 1.0), ('espresso', 0.9),
    ('latte', 0.8), ('cappuccino', 0.8), ('coffee house', 0.9), ('coffee bar', 0.8),
    ('specialty coffee', 0.9), ('coffee roasting', 0.7), ('coffee beans', 0.7),
    ('coffee drinks', 0.8), ('coffee service', 0.7), ('coffee culture', 0.6),
    ('coffee experience', 0.6), ('coffee community', 0.6), ('coffee meeting', 0.6),
    ('coffee break', 0.6), ('coffee time', 0.6), ('coffee shop atmosphere', 0.6)
) AS k(keyword, weight)
WHERE i.name = 'Cafes & Coffee Shops';

-- Casual Dining keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('casual dining', 1.0), ('family restaurant', 1.0), ('casual restaurant', 1.0),
    ('dining', 0.9), ('restaurant', 0.9), ('family dining', 0.9), ('casual food', 0.8),
    ('comfort food', 0.8), ('home style cooking', 0.8), ('family style', 0.8),
    ('casual atmosphere', 0.7), ('family friendly', 0.7), ('casual service', 0.7),
    ('dining experience', 0.7), ('casual menu', 0.7), ('family menu', 0.7),
    ('casual dining experience', 0.8), ('family restaurant experience', 0.8)
) AS k(keyword, weight)
WHERE i.name = 'Casual Dining';

-- Catering keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('catering', 1.0), ('catering services', 1.0), ('event catering', 1.0),
    ('catered events', 0.9), ('catered meals', 0.9), ('catered food', 0.9),
    ('wedding catering', 0.9), ('corporate catering', 0.9), ('party catering', 0.8),
    ('catering company', 0.9), ('catering business', 0.8), ('catering menu', 0.8),
    ('catering delivery', 0.7), ('catering setup', 0.7), ('catering staff', 0.7),
    ('catering equipment', 0.6), ('catering supplies', 0.6), ('catering logistics', 0.6)
) AS k(keyword, weight)
WHERE i.name = 'Catering';

-- Consumer Goods keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('consumer goods', 1.0), ('consumer products', 1.0), ('consumer items', 1.0),
    ('retail products', 0.9), ('consumer merchandise', 0.9), ('consumer brands', 0.8),
    ('consumer market', 0.8), ('consumer sales', 0.8), ('consumer distribution', 0.7),
    ('consumer packaging', 0.7), ('consumer quality', 0.7), ('consumer demand', 0.7),
    ('consumer trends', 0.6), ('consumer behavior', 0.6), ('consumer satisfaction', 0.6)
) AS k(keyword, weight)
WHERE i.name = 'Consumer Goods';

-- Consumer Manufacturing keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('consumer manufacturing', 1.0), ('consumer goods manufacturing', 1.0),
    ('consumer products manufacturing', 1.0), ('manufacturing', 0.9),
    ('consumer production', 0.9), ('consumer goods production', 0.9),
    ('consumer assembly', 0.8), ('consumer packaging', 0.8), ('consumer quality control', 0.7),
    ('consumer manufacturing process', 0.8), ('consumer goods production', 0.9)
) AS k(keyword, weight)
WHERE i.name = 'Consumer Manufacturing';

-- Energy Services keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('energy services', 1.0), ('energy', 1.0), ('energy solutions', 1.0),
    ('energy consulting', 0.9), ('energy management', 0.9), ('energy efficiency', 0.9),
    ('energy conservation', 0.8), ('energy optimization', 0.8), ('energy audit', 0.8),
    ('energy planning', 0.7), ('energy strategy', 0.7), ('energy services company', 0.9)
) AS k(keyword, weight)
WHERE i.name = 'Energy Services';

-- Fine Dining keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('fine dining', 1.0), ('upscale dining', 1.0), ('gourmet dining', 1.0),
    ('fine restaurant', 1.0), ('upscale restaurant', 1.0), ('gourmet restaurant', 1.0),
    ('fine cuisine', 0.9), ('gourmet cuisine', 0.9), ('upscale cuisine', 0.9),
    ('fine dining experience', 0.9), ('upscale dining experience', 0.9),
    ('gourmet dining experience', 0.9), ('fine dining atmosphere', 0.7),
    ('upscale atmosphere', 0.7), ('gourmet atmosphere', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Fine Dining';

-- Food & Beverage keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('food and beverage', 1.0), ('food beverage', 1.0), ('f&b', 0.8),
    ('food service', 0.9), ('beverage service', 0.9), ('food industry', 0.9),
    ('beverage industry', 0.9), ('food business', 0.8), ('beverage business', 0.8),
    ('food production', 0.8), ('beverage production', 0.8), ('food distribution', 0.7),
    ('beverage distribution', 0.7), ('food retail', 0.7), ('beverage retail', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Food & Beverage';

-- Food Production keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('food production', 1.0), ('food manufacturing', 1.0), ('food processing', 1.0),
    ('food factory', 0.9), ('food plant', 0.9), ('food facility', 0.9),
    ('food production facility', 0.9), ('food manufacturing facility', 0.9),
    ('food processing facility', 0.9), ('food production line', 0.8),
    ('food manufacturing line', 0.8), ('food processing line', 0.8)
) AS k(keyword, weight)
WHERE i.name = 'Food Production';

-- Food Trucks keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('food truck', 1.0), ('food trucks', 1.0), ('mobile food', 1.0),
    ('street food', 0.9), ('mobile kitchen', 0.9), ('food cart', 0.8),
    ('food trailer', 0.8), ('mobile food service', 0.9), ('street food service', 0.8),
    ('food truck business', 0.8), ('mobile food business', 0.8), ('food truck catering', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Food Trucks';

-- General Business keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('general business', 1.0), ('business', 1.0), ('company', 1.0),
    ('corporation', 0.9), ('enterprise', 0.9), ('organization', 0.9),
    ('business services', 0.8), ('business solutions', 0.8), ('business consulting', 0.8),
    ('business development', 0.7), ('business management', 0.7), ('business operations', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'General Business';

-- Healthcare keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('healthcare', 1.0), ('health care', 1.0), ('health', 1.0),
    ('medical', 0.9), ('healthcare industry', 0.9), ('healthcare sector', 0.9),
    ('healthcare services', 0.9), ('healthcare provider', 0.8), ('healthcare facility', 0.8),
    ('healthcare system', 0.8), ('healthcare delivery', 0.7), ('healthcare management', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Healthcare';

-- Industrial Manufacturing keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('industrial manufacturing', 1.0), ('industrial production', 1.0),
    ('industrial manufacturing facility', 0.9), ('industrial plant', 0.9),
    ('industrial facility', 0.9), ('industrial production facility', 0.9),
    ('industrial manufacturing process', 0.8), ('industrial production process', 0.8),
    ('industrial manufacturing equipment', 0.7), ('industrial production equipment', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Industrial Manufacturing';

-- Intellectual Property keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('intellectual property', 1.0), ('ip', 0.8), ('patents', 0.9),
    ('trademarks', 0.9), ('copyrights', 0.9), ('trade secrets', 0.8),
    ('ip law', 0.8), ('ip rights', 0.8), ('ip protection', 0.8),
    ('ip management', 0.7), ('ip consulting', 0.7), ('ip services', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Intellectual Property';

-- Law Firms keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('law firm', 1.0), ('law firms', 1.0), ('legal firm', 1.0),
    ('attorney firm', 0.9), ('lawyer firm', 0.9), ('legal practice', 0.9),
    ('law practice', 0.9), ('legal office', 0.8), ('law office', 0.8),
    ('legal services firm', 0.8), ('law services firm', 0.8)
) AS k(keyword, weight)
WHERE i.name = 'Law Firms';

-- Legal Consulting keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('legal consulting', 1.0), ('legal consultant', 1.0), ('legal advice', 1.0),
    ('legal counsel', 0.9), ('legal guidance', 0.9), ('legal expertise', 0.9),
    ('legal consultation', 0.9), ('legal support', 0.8), ('legal assistance', 0.8),
    ('legal advisory', 0.8), ('legal consulting services', 0.9)
) AS k(keyword, weight)
WHERE i.name = 'Legal Consulting';

-- Quick Service keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('quick service', 1.0), ('fast service', 1.0), ('quick food', 1.0),
    ('fast food', 0.9), ('quick dining', 0.9), ('fast dining', 0.9),
    ('quick service restaurant', 0.9), ('fast service restaurant', 0.9),
    ('quick service food', 0.8), ('fast service food', 0.8)
) AS k(keyword, weight)
WHERE i.name = 'Quick Service';

-- Retail keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('retail', 1.0), ('retail store', 1.0), ('retail business', 1.0),
    ('retail sales', 0.9), ('retail shop', 0.9), ('retail outlet', 0.9),
    ('retail store', 0.9), ('retail location', 0.8), ('retail space', 0.8),
    ('retail customer', 0.7), ('retail market', 0.7), ('retail industry', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Retail';

-- Technology keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('technology', 1.0), ('tech', 1.0), ('it', 0.9), ('information technology', 0.9),
    ('tech company', 0.9), ('technology company', 0.9), ('tech business', 0.8),
    ('technology business', 0.8), ('tech industry', 0.8), ('technology industry', 0.8),
    ('tech solutions', 0.8), ('technology solutions', 0.8), ('tech services', 0.8),
    ('technology services', 0.8), ('tech innovation', 0.7), ('technology innovation', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Technology';

-- Wholesale keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('wholesale', 1.0), ('wholesaler', 1.0), ('wholesale business', 1.0),
    ('wholesale distribution', 0.9), ('wholesale sales', 0.9), ('wholesale trade', 0.9),
    ('wholesale supplier', 0.8), ('wholesale vendor', 0.8), ('wholesale dealer', 0.8),
    ('wholesale market', 0.7), ('wholesale industry', 0.7), ('wholesale operations', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Wholesale';

-- Wineries keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('winery', 1.0), ('wineries', 1.0), ('wine production', 1.0),
    ('wine making', 1.0), ('wine', 0.9), ('wine cellar', 0.9),
    ('wine tasting', 0.9), ('wine vineyard', 0.8), ('wine grapes', 0.8),
    ('wine fermentation', 0.8), ('wine bottling', 0.7), ('wine distribution', 0.7),
    ('wine sales', 0.7), ('wine business', 0.7), ('wine industry', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Wineries';
