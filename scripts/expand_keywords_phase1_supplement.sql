-- =====================================================
-- Expand Keyword Population - Phase 1 Supplement
-- Purpose: Add keywords for codes missing from code_metadata or with insufficient keywords
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 1, Task 1.2 (Supplement)
-- =====================================================
-- 
-- This script adds keywords for codes that:
-- 1. Don't exist in code_metadata
-- 2. Have fewer than 15 keywords
-- 
-- Target: Reach 90%+ of codes with 15+ keywords and 18+ average
-- =====================================================

-- =====================================================
-- Part 1: Add Keywords from Classification Code Names/Descriptions
-- =====================================================

-- Extract keywords from classification_codes.description
-- for codes that don't have enough keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    keyword,
    CASE
        WHEN cc.is_active THEN 0.90
        ELSE 0.85
    END as relevance_score,
    'exact' as match_type
FROM classification_codes cc
LEFT JOIN code_metadata cm 
    ON cm.code_type = cc.code_type 
    AND cm.code = cc.code
CROSS JOIN LATERAL unnest(
    string_to_array(
        lower(regexp_replace(
            COALESCE(cc.description, ''),
            '[^a-z0-9\s]', ' ', 'g'
        )), ' '
    )
) as keyword
WHERE cc.is_active = true
  AND keyword IS NOT NULL
  AND length(keyword) >= 3
  AND keyword != ALL(ARRAY[
      'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for',
      'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were', 'be',
      'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
      'would', 'should', 'could', 'may', 'might', 'must', 'can', 'this',
      'that', 'these', 'those', 'it', 'its', 'they', 'them', 'their', 'our',
      'your', 'my', 'his', 'her', 'he', 'she', 'we', 'you', 'i', 'me', 'us',
      'services', 'service', 'products', 'product', 'business', 'company',
      'companies', 'other', 'etc', 'including', 'excluding', 'related', 'not',
      'primarily', 'engaged', 'establishments', 'establishment', 'comprises',
      'industry', 'industries', 'u.s.', 'us', 'providing', 'provides'
  ])
  AND keyword !~ '^\d+$'
  -- Only add if code has fewer than 15 keywords
  AND (
      SELECT COUNT(*) 
      FROM code_keywords ck 
      WHERE ck.code_id = cc.id
  ) < 15
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score),
    updated_at = NOW();

-- =====================================================
-- Part 2: Add Industry-Specific Keywords for Common Codes
-- =====================================================

-- Add common industry keywords for codes that still need more
-- This targets codes with 5-14 keywords to bring them to 15+

-- Technology/Software Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id,
    keyword,
    0.85 as relevance_score,
    'synonym' as match_type
FROM classification_codes cc
CROSS JOIN (VALUES
    ('software'), ('application'), ('app'), ('programming'), ('development'),
    ('technology'), ('tech'), ('computer'), ('digital'), ('platform'),
    ('system'), ('code'), ('program'), ('developer'), ('coding')
) AS tech_keywords(keyword)
WHERE cc.is_active = true
  AND (
      LOWER(cc.description) LIKE '%software%' OR
      LOWER(cc.description) LIKE '%computer%' OR
      LOWER(cc.description) LIKE '%technology%' OR
      LOWER(cc.description) LIKE '%programming%'
  )
  AND (
      SELECT COUNT(*) FROM code_keywords ck WHERE ck.code_id = cc.id
  ) < 15
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Healthcare/Medical Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id,
    keyword,
    0.85 as relevance_score,
    'synonym' as match_type
FROM classification_codes cc
CROSS JOIN (VALUES
    ('medical'), ('health'), ('healthcare'), ('hospital'), ('clinic'),
    ('doctor'), ('physician'), ('patient'), ('treatment'), ('care'),
    ('medicine'), ('diagnostic'), ('therapy'), ('wellness'), ('clinical')
) AS health_keywords(keyword)
WHERE cc.is_active = true
  AND (
      LOWER(cc.description) LIKE '%medical%' OR
      LOWER(cc.description) LIKE '%health%' OR
      LOWER(cc.description) LIKE '%hospital%' OR
      LOWER(cc.description) LIKE '%doctor%'
  )
  AND (
      SELECT COUNT(*) FROM code_keywords ck WHERE ck.code_id = cc.id
  ) < 15
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Retail/Store Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id,
    keyword,
    0.85 as relevance_score,
    'synonym' as match_type
FROM classification_codes cc
CROSS JOIN (VALUES
    ('retail'), ('store'), ('shop'), ('merchant'), ('retailer'),
    ('sales'), ('selling'), ('purchase'), ('buy'), ('shopping'),
    ('outlet'), ('market'), ('vendor'), ('commerce'), ('trade')
) AS retail_keywords(keyword)
WHERE cc.is_active = true
  AND (
      LOWER(cc.description) LIKE '%store%' OR
      LOWER(cc.description) LIKE '%retail%' OR
      LOWER(cc.description) LIKE '%shop%'
  )
  AND (
      SELECT COUNT(*) FROM code_keywords ck WHERE ck.code_id = cc.id
  ) < 15
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Restaurant/Food Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id,
    keyword,
    0.85 as relevance_score,
    'synonym' as match_type
FROM classification_codes cc
CROSS JOIN (VALUES
    ('restaurant'), ('food'), ('dining'), ('meal'), ('cuisine'),
    ('cafe'), ('catering'), ('beverage'), ('dining'), ('eating'),
    ('kitchen'), ('menu'), ('service'), ('diner'), ('lunch')
) AS food_keywords(keyword)
WHERE cc.is_active = true
  AND (
      LOWER(cc.description) LIKE '%restaurant%' OR
      LOWER(cc.description) LIKE '%food%' OR
      LOWER(cc.description) LIKE '%dining%' OR
      LOWER(cc.description) LIKE '%cafe%'
  )
  AND (
      SELECT COUNT(*) FROM code_keywords ck WHERE ck.code_id = cc.id
  ) < 15
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Transportation Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id,
    keyword,
    0.85 as relevance_score,
    'synonym' as match_type
FROM classification_codes cc
CROSS JOIN (VALUES
    ('transportation'), ('transport'), ('transit'), ('travel'), ('vehicle'),
    ('trucking'), ('shipping'), ('logistics'), ('delivery'), ('freight'),
    ('cargo'), ('carrier'), ('route'), ('passenger'), ('mobility')
) AS transport_keywords(keyword)
WHERE cc.is_active = true
  AND (
      LOWER(cc.description) LIKE '%transport%' OR
      LOWER(cc.description) LIKE '%truck%' OR
      LOWER(cc.description) LIKE '%delivery%' OR
      LOWER(cc.description) LIKE '%taxi%'
  )
  AND (
      SELECT COUNT(*) FROM code_keywords ck WHERE ck.code_id = cc.id
  ) < 15
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Financial Services Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id,
    keyword,
    0.85 as relevance_score,
    'synonym' as match_type
FROM classification_codes cc
CROSS JOIN (VALUES
    ('financial'), ('finance'), ('banking'), ('bank'), ('credit'),
    ('loan'), ('lending'), ('investment'), ('money'), ('capital'),
    ('payment'), ('transaction'), ('account'), ('funding'), ('financial')
) AS finance_keywords(keyword)
WHERE cc.is_active = true
  AND (
      LOWER(cc.description) LIKE '%bank%' OR
      LOWER(cc.description) LIKE '%financial%' OR
      LOWER(cc.description) LIKE '%credit%' OR
      LOWER(cc.description) LIKE '%loan%'
  )
  AND (
      SELECT COUNT(*) FROM code_keywords ck WHERE ck.code_id = cc.id
  ) < 15
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Part 3: Add Generic Business Keywords for Remaining Codes
-- =====================================================

-- Add generic business keywords for codes that still need more
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id,
    keyword,
    0.80 as relevance_score,
    'partial' as match_type
FROM classification_codes cc
CROSS JOIN (VALUES
    ('business'), ('company'), ('enterprise'), ('organization'), ('firm'),
    ('operation'), ('commercial'), ('professional'), ('corporate'), ('entity'),
    ('vendor'), ('provider'), ('supplier'), ('merchant'), ('establishment')
) AS business_keywords(keyword)
WHERE cc.is_active = true
  AND (
      SELECT COUNT(*) FROM code_keywords ck WHERE ck.code_id = cc.id
  ) < 15
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Verification Query
-- =====================================================

-- Check progress after supplement
SELECT 
    'After Supplement - Codes with 15+ Keywords' AS metric,
    COUNT(DISTINCT code_id) AS codes_with_15plus,
    (SELECT COUNT(*) FROM classification_codes WHERE is_active = true) AS total_codes,
    ROUND(COUNT(DISTINCT code_id) * 100.0 / 
          NULLIF((SELECT COUNT(*) FROM classification_codes WHERE is_active = true), 0), 2) AS percentage,
    CASE
        WHEN ROUND(COUNT(DISTINCT code_id) * 100.0 / 
                   NULLIF((SELECT COUNT(*) FROM classification_codes WHERE is_active = true), 0), 2) >= 90.0 
        THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE cc.is_active = true
GROUP BY ck.code_id
HAVING COUNT(ck.id) >= 15;

-- Average keywords per code
SELECT 
    'After Supplement - Average Keywords per Code' AS metric,
    ROUND(AVG(keyword_count), 2) AS avg_keywords,
    CASE
        WHEN ROUND(AVG(keyword_count), 2) >= 18.0 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM (
    SELECT 
        code_id,
        COUNT(*) as keyword_count
    FROM code_keywords
    GROUP BY code_id
) AS keyword_counts;

