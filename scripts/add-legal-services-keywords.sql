-- =============================================================================
-- TASK 3.2.1: ADD LEGAL SERVICES KEYWORDS
-- =============================================================================
-- This script adds comprehensive legal services keywords for all 4 legal
-- industries to achieve >85% classification accuracy for legal businesses.
-- 
-- Legal Industries Covered:
-- 1. Law Firms (confidence_threshold: 0.80)
-- 2. Legal Consulting (confidence_threshold: 0.75)
-- 3. Legal Services (confidence_threshold: 0.70)
-- 4. Intellectual Property (confidence_threshold: 0.85)
--
-- Keywords Added: 200+ legal-specific keywords with base weights 0.5-1.0
-- =============================================================================

-- Start transaction for atomic operation
BEGIN;

-- =============================================================================
-- 1. LAW FIRMS KEYWORDS (50+ keywords)
-- =============================================================================
INSERT INTO keyword_weights (industry_id, keyword, base_weight, context_multiplier, usage_count, is_active)
SELECT 
  i.id,
  kw.keyword,
  kw.base_weight,
  1.0 as context_multiplier,
  0 as usage_count,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  -- Core legal terms (highest weight)
  ('law firm', 1.000), ('attorney', 0.950), ('lawyer', 0.950), ('legal', 0.900),
  ('law', 0.900), ('counsel', 0.900), ('advocate', 0.850), ('barrister', 0.850),
  
  -- Practice areas (high weight)
  ('litigation', 0.900), ('corporate law', 0.900), ('criminal defense', 0.900),
  ('family law', 0.850), ('personal injury', 0.850), ('real estate law', 0.850),
  ('employment law', 0.800), ('immigration law', 0.800), ('tax law', 0.800),
  ('bankruptcy', 0.800), ('estate planning', 0.800), ('contract law', 0.800),
  
  -- Legal services (medium-high weight)
  ('legal representation', 0.850), ('legal advice', 0.800), ('legal counsel', 0.800),
  ('legal services', 0.800), ('legal assistance', 0.750), ('legal support', 0.750),
  ('case management', 0.750), ('court representation', 0.750), ('legal defense', 0.750),
  
  -- Professional terms (medium weight)
  ('juris doctor', 0.700), ('esquire', 0.700), ('partner', 0.700), ('associate', 0.650),
  ('senior partner', 0.750), ('managing partner', 0.750), ('of counsel', 0.700),
  ('paralegal', 0.600), ('legal secretary', 0.550), ('legal assistant', 0.550),
  
  -- Legal processes (medium weight)
  ('trial', 0.750), ('settlement', 0.750), ('mediation', 0.700), ('arbitration', 0.700),
  ('negotiation', 0.700), ('deposition', 0.650), ('discovery', 0.650), ('pleading', 0.600),
  ('motion', 0.600), ('hearing', 0.600), ('appeal', 0.600), ('verdict', 0.600),
  
  -- Business context (lower weight)
  ('firm', 0.650), ('practice', 0.650), ('office', 0.600), ('chambers', 0.600),
  ('legal department', 0.600), ('law office', 0.600), ('legal team', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Law Firms'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
  base_weight = EXCLUDED.base_weight,
  context_multiplier = EXCLUDED.context_multiplier,
  is_active = EXCLUDED.is_active,
  updated_at = NOW();

-- =============================================================================
-- 2. LEGAL CONSULTING KEYWORDS (50+ keywords)
-- =============================================================================
INSERT INTO keyword_weights (industry_id, keyword, base_weight, context_multiplier, usage_count, is_active)
SELECT 
  i.id,
  kw.keyword,
  kw.base_weight,
  1.0 as context_multiplier,
  0 as usage_count,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  -- Core consulting terms (highest weight)
  ('legal consulting', 1.000), ('legal advisor', 0.950), ('legal consultant', 0.950),
  ('legal advisory', 0.900), ('legal expertise', 0.900), ('legal guidance', 0.900),
  ('legal counsel', 0.850), ('legal support', 0.800), ('legal assistance', 0.800),
  
  -- Advisory services (high weight)
  ('compliance consulting', 0.900), ('regulatory consulting', 0.900), ('risk management', 0.850),
  ('legal strategy', 0.850), ('legal planning', 0.800), ('legal analysis', 0.800),
  ('legal review', 0.800), ('legal assessment', 0.750), ('legal evaluation', 0.750),
  
  -- Specialized consulting (medium-high weight)
  ('corporate consulting', 0.850), ('business consulting', 0.800), ('transaction consulting', 0.800),
  ('due diligence', 0.800), ('legal research', 0.750), ('legal writing', 0.700),
  ('contract review', 0.750), ('policy consulting', 0.700), ('governance consulting', 0.700),
  
  -- Professional services (medium weight)
  ('legal training', 0.700), ('legal education', 0.700), ('legal seminars', 0.650),
  ('legal workshops', 0.650), ('legal coaching', 0.650), ('legal mentoring', 0.600),
  ('legal development', 0.600), ('legal consulting services', 0.750),
  
  -- Industry expertise (medium weight)
  ('industry expertise', 0.700), ('sector knowledge', 0.650), ('domain expertise', 0.650),
  ('specialized knowledge', 0.650), ('professional expertise', 0.600), ('expert advice', 0.600),
  ('specialized consulting', 0.700), ('niche consulting', 0.650),
  
  -- Business context (lower weight)
  ('consulting firm', 0.650), ('advisory firm', 0.650), ('consulting services', 0.600),
  ('advisory services', 0.600), ('professional services', 0.600), ('expert services', 0.600),
  ('consulting practice', 0.600), ('advisory practice', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Legal Consulting'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
  base_weight = EXCLUDED.base_weight,
  context_multiplier = EXCLUDED.context_multiplier,
  is_active = EXCLUDED.is_active,
  updated_at = NOW();

-- =============================================================================
-- 3. LEGAL SERVICES KEYWORDS (50+ keywords)
-- =============================================================================
INSERT INTO keyword_weights (industry_id, keyword, base_weight, context_multiplier, usage_count, is_active)
SELECT 
  i.id,
  kw.keyword,
  kw.base_weight,
  1.0 as context_multiplier,
  0 as usage_count,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  -- Core legal services (highest weight)
  ('legal services', 1.000), ('legal support', 0.900), ('legal assistance', 0.900),
  ('legal help', 0.850), ('legal aid', 0.850), ('legal representation', 0.800),
  ('legal advice', 0.800), ('legal counsel', 0.800), ('legal guidance', 0.750),
  
  -- Support services (high weight)
  ('paralegal services', 0.900), ('legal research', 0.850), ('legal writing', 0.800),
  ('document preparation', 0.800), ('legal documentation', 0.750), ('case preparation', 0.750),
  ('legal filing', 0.700), ('court filing', 0.700), ('legal paperwork', 0.700),
  
  -- Administrative services (medium-high weight)
  ('legal administration', 0.750), ('legal coordination', 0.700), ('legal management', 0.700),
  ('case management', 0.700), ('client services', 0.650), ('legal intake', 0.650),
  ('legal screening', 0.600), ('legal triage', 0.600), ('legal processing', 0.600),
  
  -- Specialized services (medium weight)
  ('notary services', 0.700), ('document notarization', 0.700), ('legal translation', 0.650),
  ('legal transcription', 0.650), ('legal proofreading', 0.600), ('legal editing', 0.600),
  ('legal formatting', 0.550), ('legal indexing', 0.550), ('legal archiving', 0.550),
  
  -- Technology services (medium weight)
  ('legal technology', 0.700), ('legal software', 0.650), ('legal database', 0.600),
  ('legal systems', 0.600), ('legal automation', 0.600), ('legal workflow', 0.600),
  ('legal process', 0.600), ('legal operations', 0.600),
  
  -- Support roles (lower weight)
  ('paralegal', 0.700), ('legal assistant', 0.650), ('legal secretary', 0.600),
  ('legal clerk', 0.600), ('legal intern', 0.550), ('legal trainee', 0.550),
  ('legal staff', 0.600), ('legal personnel', 0.600), ('legal team', 0.600),
  
  -- Business context (lower weight)
  ('legal office', 0.600), ('legal department', 0.600), ('legal center', 0.600),
  ('legal facility', 0.600), ('legal organization', 0.600), ('legal company', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Legal Services'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
  base_weight = EXCLUDED.base_weight,
  context_multiplier = EXCLUDED.context_multiplier,
  is_active = EXCLUDED.is_active,
  updated_at = NOW();

-- =============================================================================
-- 4. INTELLECTUAL PROPERTY KEYWORDS (50+ keywords)
-- =============================================================================
INSERT INTO keyword_weights (industry_id, keyword, base_weight, context_multiplier, usage_count, is_active)
SELECT 
  i.id,
  kw.keyword,
  kw.base_weight,
  1.0 as context_multiplier,
  0 as usage_count,
  true as is_active
FROM industries i
CROSS JOIN (VALUES
  -- Core IP terms (highest weight)
  ('intellectual property', 1.000), ('patent', 0.950), ('trademark', 0.950),
  ('copyright', 0.950), ('IP', 0.900), ('patent attorney', 0.900), ('IP attorney', 0.900),
  ('IP lawyer', 0.900), ('patent lawyer', 0.900), ('trademark attorney', 0.900),
  
  -- Patent services (high weight)
  ('patent prosecution', 0.900), ('patent application', 0.900), ('patent filing', 0.900),
  ('patent search', 0.850), ('patent analysis', 0.850), ('patent examination', 0.850),
  ('patent litigation', 0.850), ('patent infringement', 0.850), ('patent validity', 0.800),
  ('patent portfolio', 0.800), ('patent strategy', 0.800), ('patent licensing', 0.800),
  
  -- Trademark services (high weight)
  ('trademark registration', 0.900), ('trademark filing', 0.900), ('trademark search', 0.850),
  ('trademark clearance', 0.850), ('trademark prosecution', 0.850), ('trademark opposition', 0.800),
  ('trademark cancellation', 0.800), ('trademark enforcement', 0.800), ('trademark monitoring', 0.750),
  ('trademark portfolio', 0.750), ('trademark strategy', 0.750),
  
  -- Copyright services (high weight)
  ('copyright registration', 0.900), ('copyright filing', 0.850), ('copyright protection', 0.850),
  ('copyright enforcement', 0.800), ('copyright infringement', 0.800), ('copyright licensing', 0.750),
  ('copyright assignment', 0.750), ('copyright transfer', 0.750), ('copyright work', 0.700),
  
  -- Trade secrets (medium-high weight)
  ('trade secret', 0.850), ('trade secret protection', 0.800), ('confidentiality', 0.800),
  ('non-disclosure', 0.800), ('NDA', 0.750), ('proprietary information', 0.750),
  ('trade secret litigation', 0.750), ('misappropriation', 0.700),
  
  -- IP transactions (medium weight)
  ('IP licensing', 0.750), ('IP assignment', 0.750), ('IP transfer', 0.750),
  ('IP acquisition', 0.700), ('IP due diligence', 0.700), ('IP valuation', 0.700),
  ('IP audit', 0.700), ('IP portfolio management', 0.700),
  
  -- Technology areas (medium weight)
  ('software patents', 0.750), ('biotech patents', 0.750), ('pharmaceutical patents', 0.750),
  ('mechanical patents', 0.700), ('electrical patents', 0.700), ('chemical patents', 0.700),
  ('design patents', 0.700), ('utility patents', 0.700),
  
  -- Business context (lower weight)
  ('IP law firm', 0.700), ('IP practice', 0.700), ('IP services', 0.700),
  ('IP consulting', 0.700), ('IP advisory', 0.700), ('IP expertise', 0.700),
  ('IP specialist', 0.700), ('IP professional', 0.700)
) AS kw(keyword, base_weight)
WHERE i.name = 'Intellectual Property'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
  base_weight = EXCLUDED.base_weight,
  context_multiplier = EXCLUDED.context_multiplier,
  is_active = EXCLUDED.is_active,
  updated_at = NOW();

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify all legal keywords were added
DO $$
DECLARE
    law_firms_count INTEGER;
    legal_consulting_count INTEGER;
    legal_services_count INTEGER;
    ip_count INTEGER;
    total_count INTEGER;
BEGIN
    -- Count keywords for each legal industry
    SELECT COUNT(*) INTO law_firms_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Law Firms' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO legal_consulting_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Legal Consulting' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO legal_services_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Legal Services' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO ip_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Intellectual Property' AND kw.is_active = true;
    
    total_count := law_firms_count + legal_consulting_count + legal_services_count + ip_count;
    
    RAISE NOTICE 'LEGAL KEYWORDS ADDED:';
    RAISE NOTICE 'Law Firms: % keywords', law_firms_count;
    RAISE NOTICE 'Legal Consulting: % keywords', legal_consulting_count;
    RAISE NOTICE 'Legal Services: % keywords', legal_services_count;
    RAISE NOTICE 'Intellectual Property: % keywords', ip_count;
    RAISE NOTICE 'Total Legal Keywords: %', total_count;
    
    IF total_count >= 200 THEN
        RAISE NOTICE 'SUCCESS: All legal keywords added successfully (>= 200 total)';
    ELSE
        RAISE NOTICE 'WARNING: Expected >= 200 keywords, but found %', total_count;
    END IF;
END $$;

-- Display keyword summary by legal industry
SELECT 
    'LEGAL KEYWORDS SUMMARY' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    i.confidence_threshold,
    COUNT(kw.id) as keyword_count,
    ROUND(AVG(kw.base_weight), 3) as avg_base_weight,
    ROUND(MIN(kw.base_weight), 3) as min_base_weight,
    ROUND(MAX(kw.base_weight), 3) as max_base_weight
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.name IN ('Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property')
GROUP BY i.id, i.name, i.confidence_threshold
ORDER BY i.confidence_threshold DESC;

-- Display sample keywords for each industry
SELECT 
    'SAMPLE LEGAL KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    kw.keyword,
    kw.base_weight,
    CASE 
        WHEN kw.base_weight >= 0.90 THEN 'High'
        WHEN kw.base_weight >= 0.75 THEN 'Medium-High'
        WHEN kw.base_weight >= 0.60 THEN 'Medium'
        ELSE 'Low'
    END as weight_level
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.name IN ('Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property')
AND kw.is_active = true
ORDER BY i.name, kw.base_weight DESC
LIMIT 20;

-- Final verification
SELECT 
    'FINAL VERIFICATION' as summary_type,
    '' as spacer;

SELECT 
    COUNT(DISTINCT i.id) as legal_industries,
    COUNT(kw.id) as total_legal_keywords,
    ROUND(AVG(kw.base_weight), 3) as avg_keyword_weight,
    COUNT(CASE WHEN kw.base_weight >= 0.90 THEN 1 END) as high_weight_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.75 AND kw.base_weight < 0.90 THEN 1 END) as medium_high_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.50 AND kw.base_weight < 0.75 THEN 1 END) as medium_keywords
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.name IN ('Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property')
AND kw.is_active = true;

COMMIT;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TASK 3.2.1: LEGAL SERVICES KEYWORDS COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Legal industries covered: 4';
    RAISE NOTICE 'Keywords added: 200+ legal-specific keywords';
    RAISE NOTICE 'Weight range: 0.50-1.00 (as specified in plan)';
    RAISE NOTICE 'Coverage: Law Firms, Legal Consulting, Legal Services, IP';
    RAISE NOTICE 'Status: Ready for testing and validation';
    RAISE NOTICE 'Next: Task 3.2.2 - Add healthcare keywords';
    RAISE NOTICE '=============================================================================';
END $$;
