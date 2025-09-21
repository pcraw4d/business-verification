-- Add keywords for industries that are missing or have poor coverage
-- This script addresses the remaining gaps identified in the accuracy tests

-- Breweries keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('brewery', 1.0), ('brewing', 1.0), ('beer', 1.0), ('craft beer', 0.9),
    ('microbrewery', 0.9), ('ale', 0.8), ('lager', 0.8), ('stout', 0.8),
    ('ipa', 0.8), ('pilsner', 0.7), ('wheat beer', 0.7), ('porter', 0.7),
    ('barley', 0.6), ('hops', 0.6), ('fermentation', 0.6), ('tasting room', 0.8),
    ('brewpub', 0.9), ('taproom', 0.8), ('beer garden', 0.7), ('brewmaster', 0.8),
    ('artisanal beer', 0.9), ('local brewery', 0.8), ('beer production', 0.8),
    ('beer distribution', 0.7), ('beer sales', 0.7), ('beer tasting', 0.8)
) AS k(keyword, weight)
WHERE i.name = 'Breweries';

-- Legal Services keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('legal', 1.0), ('law', 1.0), ('attorney', 1.0), ('lawyer', 1.0),
    ('legal services', 1.0), ('law firm', 1.0), ('legal counsel', 0.9),
    ('litigation', 0.9), ('corporate law', 0.9), ('contract law', 0.8),
    ('criminal law', 0.8), ('family law', 0.8), ('real estate law', 0.8),
    ('immigration law', 0.8), ('employment law', 0.8), ('intellectual property', 0.8),
    ('patent law', 0.7), ('trademark', 0.7), ('copyright', 0.7), ('legal advice', 0.9),
    ('legal representation', 0.9), ('court', 0.7), ('legal document', 0.8),
    ('legal consultation', 0.9), ('legal support', 0.8), ('legal assistance', 0.8)
) AS k(keyword, weight)
WHERE i.name = 'Legal Services';

-- Healthcare Technology keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('healthcare technology', 1.0), ('health tech', 1.0), ('medical technology', 1.0),
    ('healthcare software', 0.9), ('medical software', 0.9), ('healthcare it', 0.9),
    ('electronic health records', 0.9), ('ehr', 0.8), ('healthcare data', 0.8),
    ('telemedicine', 0.9), ('digital health', 0.9), ('healthcare innovation', 0.8),
    ('medical devices', 0.8), ('healthcare analytics', 0.8), ('healthcare ai', 0.8),
    ('healthcare automation', 0.7), ('healthcare integration', 0.7), ('healthcare platform', 0.8),
    ('healthcare solutions', 0.9), ('healthcare systems', 0.8), ('healthcare management', 0.7),
    ('healthcare compliance', 0.7), ('healthcare security', 0.7), ('healthcare workflow', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Healthcare Technology';

-- Software Development keywords (enhance existing)
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('software development', 1.0), ('software engineering', 1.0), ('programming', 1.0),
    ('coding', 1.0), ('software solutions', 1.0), ('custom software', 1.0),
    ('enterprise software', 0.9), ('software consulting', 0.9), ('software architecture', 0.9),
    ('software design', 0.8), ('software testing', 0.8), ('software deployment', 0.8),
    ('software maintenance', 0.7), ('software optimization', 0.7), ('software integration', 0.8),
    ('software development lifecycle', 0.8), ('sdlc', 0.7), ('agile development', 0.9),
    ('scrum', 0.8), ('devops', 0.8), ('continuous integration', 0.7),
    ('continuous deployment', 0.7), ('version control', 0.7), ('code review', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Software Development';

-- Fintech keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('fintech', 1.0), ('financial technology', 1.0), ('digital banking', 0.9),
    ('mobile payments', 0.9), ('online banking', 0.9), ('digital wallet', 0.8),
    ('cryptocurrency', 0.8), ('blockchain', 0.8), ('robo advisor', 0.8),
    ('peer to peer lending', 0.8), ('p2p lending', 0.7), ('crowdfunding', 0.7),
    ('insurtech', 0.7), ('regtech', 0.7), ('wealthtech', 0.7),
    ('paytech', 0.7), ('lendtech', 0.7), ('tradetech', 0.7),
    ('financial innovation', 0.8), ('digital finance', 0.8), ('fintech solutions', 0.9),
    ('fintech platform', 0.8), ('fintech services', 0.8), ('fintech startup', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Fintech';

-- Digital Services keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('digital services', 1.0), ('digital marketing', 1.0), ('digital transformation', 1.0),
    ('digital solutions', 1.0), ('digital agency', 0.9), ('digital consulting', 0.9),
    ('digital strategy', 0.9), ('digital innovation', 0.8), ('digital technology', 0.8),
    ('digital platform', 0.8), ('digital tools', 0.7), ('digital automation', 0.7),
    ('digital integration', 0.7), ('digital optimization', 0.7), ('digital analytics', 0.8),
    ('digital content', 0.7), ('digital media', 0.7), ('digital advertising', 0.8),
    ('digital branding', 0.7), ('digital communication', 0.7), ('digital experience', 0.7),
    ('digital workflow', 0.7), ('digital process', 0.7), ('digital business', 0.8)
) AS k(keyword, weight)
WHERE i.name = 'Digital Services';

-- Technology Services keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('technology services', 1.0), ('it services', 1.0), ('tech services', 1.0),
    ('it consulting', 1.0), ('technology consulting', 1.0), ('it support', 0.9),
    ('technical support', 0.9), ('system administration', 0.9), ('network administration', 0.8),
    ('database administration', 0.8), ('cloud services', 0.9), ('managed services', 0.8),
    ('it infrastructure', 0.8), ('system integration', 0.8), ('technology solutions', 0.9),
    ('it solutions', 0.9), ('tech solutions', 0.9), ('technology support', 0.8),
    ('it management', 0.7), ('technology management', 0.7), ('it operations', 0.7),
    ('technology operations', 0.7), ('it maintenance', 0.7), ('technology maintenance', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Technology Services';

-- Healthcare Services keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('healthcare services', 1.0), ('health services', 1.0), ('medical services', 1.0),
    ('healthcare', 1.0), ('healthcare provider', 0.9), ('healthcare facility', 0.8),
    ('healthcare center', 0.8), ('healthcare clinic', 0.8), ('healthcare practice', 0.8),
    ('healthcare management', 0.7), ('healthcare administration', 0.7), ('healthcare operations', 0.7),
    ('healthcare delivery', 0.7), ('healthcare quality', 0.7), ('healthcare safety', 0.7),
    ('healthcare compliance', 0.7), ('healthcare regulation', 0.7), ('healthcare policy', 0.7),
    ('healthcare planning', 0.7), ('healthcare coordination', 0.7), ('healthcare support', 0.7),
    ('healthcare assistance', 0.7), ('healthcare care', 0.8), ('healthcare treatment', 0.8)
) AS k(keyword, weight)
WHERE i.name = 'Healthcare Services';

-- Fast Food keywords (to prevent misclassification)
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
SELECT i.id, k.keyword, k.weight, TRUE
FROM industries i
CROSS JOIN (VALUES
    ('fast food', 1.0), ('quick service', 1.0), ('drive thru', 0.9),
    ('takeout', 0.9), ('delivery', 0.8), ('fast casual', 0.8),
    ('burger', 0.8), ('fries', 0.7), ('sandwich', 0.7), ('pizza', 0.7),
    ('fried chicken', 0.7), ('taco', 0.7), ('hot dog', 0.6), ('nuggets', 0.6),
    ('combo meal', 0.6), ('value meal', 0.6), ('fast service', 0.8),
    ('counter service', 0.7), ('self service', 0.6), ('fast food restaurant', 0.9),
    ('quick food', 0.7), ('instant food', 0.6), ('convenience food', 0.7)
) AS k(keyword, weight)
WHERE i.name = 'Fast Food';
