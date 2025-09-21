-- =============================================================================
-- INDEX PERFORMANCE TESTING SCRIPT
-- Subtask 3.2.2: Test Index Performance
-- Supabase Table Improvement Implementation Plan
-- =============================================================================
-- This script tests the performance of all newly created indexes

-- =============================================================================
-- 1. TEST CLASSIFICATION SYSTEM INDEXES
-- =============================================================================

-- Test industries table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industries 
WHERE category = 'traditional' AND is_active = true 
ORDER BY confidence_threshold DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industries 
WHERE name ILIKE '%technology%' 
ORDER BY created_at DESC;

-- Test industry_keywords table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT ik.*, i.name as industry_name 
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE ik.keyword ILIKE '%software%' AND ik.is_active = true
ORDER BY ik.weight DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industry_keywords 
WHERE industry_id = 1 AND is_active = true
ORDER BY weight DESC;

-- Test classification_codes table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM classification_codes 
WHERE code_type = 'MCC' AND is_active = true
ORDER BY code;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM classification_codes 
WHERE description ILIKE '%banking%' AND is_active = true;

-- Test industry_patterns table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industry_patterns 
WHERE industry_id = 1 AND pattern_type = 'phrase' AND is_active = true
ORDER BY confidence_score DESC;

-- Test keyword_weights table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM keyword_weights 
WHERE keyword = 'software' AND industry_id = 1
ORDER BY base_weight DESC;

-- =============================================================================
-- 2. TEST RISK KEYWORDS SYSTEM INDEXES
-- =============================================================================

-- Test risk_keywords table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE risk_category = 'prohibited' AND risk_severity = 'high' AND is_active = true
ORDER BY risk_score_weight DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE keyword ILIKE '%gambling%' AND is_active = true;

-- Test industry_code_crosswalks table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industry_code_crosswalks 
WHERE mcc_code = '6011' AND is_active = true
ORDER BY confidence_score DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industry_code_crosswalks 
WHERE industry_id = 2 AND is_active = true
ORDER BY usage_frequency DESC;

-- Test business_risk_assessments table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM business_risk_assessments 
WHERE business_id = '123e4567-e89b-12d3-a456-426614174000' 
ORDER BY assessment_date DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM business_risk_assessments 
WHERE risk_level IN ('high', 'critical') 
ORDER BY risk_score DESC, assessment_date DESC;

-- Test risk_keyword_relationships table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keyword_relationships 
WHERE parent_keyword_id = 1 AND relationship_type = 'synonym' AND is_active = true;

-- =============================================================================
-- 3. TEST PERFORMANCE MONITORING INDEXES
-- =============================================================================

-- Test classification_performance_metrics table indexes
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM classification_performance_metrics 
WHERE timestamp >= NOW() - INTERVAL '7 days'
ORDER BY timestamp DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM classification_performance_metrics 
WHERE classification_method = 'BERT' AND accuracy_score >= 0.90
ORDER BY timestamp DESC;

-- =============================================================================
-- 4. TEST COMPOSITE INDEXES
-- =============================================================================

-- Test composite indexes for classification system
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industries 
WHERE category = 'traditional' AND confidence_threshold >= 0.80 AND is_active = true
ORDER BY confidence_threshold DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industry_keywords 
WHERE weight >= 0.80 AND industry_id = 1 AND is_active = true
ORDER BY weight DESC;

-- Test composite indexes for risk assessment
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE risk_category = 'prohibited' AND risk_severity = 'high' AND risk_score_weight >= 1.0 AND is_active = true
ORDER BY risk_score_weight DESC;

-- =============================================================================
-- 5. TEST FULL-TEXT SEARCH INDEXES
-- =============================================================================

-- Test full-text search on industries
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industries 
WHERE to_tsvector('english', name || ' ' || COALESCE(description, '')) @@ to_tsquery('english', 'technology & software');

-- Test full-text search on risk keywords
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE to_tsvector('english', keyword || ' ' || COALESCE(description, '')) @@ to_tsquery('english', 'gambling | casino');

-- =============================================================================
-- 6. TEST TRIGRAM INDEXES
-- =============================================================================

-- Test trigram similarity search
EXPLAIN (ANALYZE, BUFFERS) 
SELECT *, similarity(name, 'technolgy') as sim 
FROM industries 
WHERE name % 'technolgy' 
ORDER BY sim DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT *, similarity(keyword, 'sofware') as sim 
FROM industry_keywords 
WHERE keyword % 'sofware' 
ORDER BY sim DESC;

-- =============================================================================
-- 7. TEST PARTIAL INDEXES
-- =============================================================================

-- Test partial indexes for active records
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industries 
WHERE is_active = true 
ORDER BY created_at DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE is_active = true AND risk_severity IN ('high', 'critical');

-- =============================================================================
-- 8. TEST JSONB INDEXES
-- =============================================================================

-- Test JSONB indexes on metadata fields
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM users 
WHERE metadata @> '{"role": "admin"}';

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM businesses 
WHERE address @> '{"country": "US"}';

-- Test JSONB indexes on assessment metadata
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM business_risk_assessments 
WHERE assessment_metadata @> '{"source": "website_analysis"}';

-- =============================================================================
-- 9. TEST ARRAY INDEXES
-- =============================================================================

-- Test GIN indexes on array fields
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE mcc_codes @> ARRAY['6011'];

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM business_risk_assessments 
WHERE detected_keywords @> ARRAY['gambling'];

-- =============================================================================
-- 10. TEST TIME-BASED INDEXES
-- =============================================================================

-- Test time-based queries
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industries 
WHERE created_at >= DATE_TRUNC('month', NOW() - INTERVAL '1 month')
ORDER BY created_at DESC;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM business_risk_assessments 
WHERE assessment_date >= DATE_TRUNC('week', NOW() - INTERVAL '1 week')
ORDER BY assessment_date DESC;

-- =============================================================================
-- 11. TEST FOREIGN KEY INDEXES
-- =============================================================================

-- Test foreign key join performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT i.name, ik.keyword, ik.weight
FROM industries i
JOIN industry_keywords ik ON i.id = ik.industry_id
WHERE i.category = 'traditional' AND ik.is_active = true
ORDER BY ik.weight DESC;

-- =============================================================================
-- 12. TEST UNIQUE CONSTRAINT INDEXES
-- =============================================================================

-- Test unique constraint performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industries 
WHERE name = 'Technology' AND is_active = true;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industry_keywords 
WHERE industry_id = 1 AND keyword = 'software' AND is_active = true;

-- =============================================================================
-- 13. TEST API PERFORMANCE INDEXES
-- =============================================================================

-- Test API-optimized queries
EXPLAIN (ANALYZE, BUFFERS) 
SELECT id, name, category, is_active 
FROM industries 
WHERE is_active = true 
ORDER BY name;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT id, keyword, risk_category, risk_severity 
FROM risk_keywords 
WHERE is_active = true 
ORDER BY keyword;

-- =============================================================================
-- 14. TEST REPORTING INDEXES
-- =============================================================================

-- Test reporting queries
EXPLAIN (ANALYZE, BUFFERS) 
SELECT 
    DATE_TRUNC('month', timestamp) as month,
    classification_method,
    AVG(accuracy_score) as avg_accuracy,
    COUNT(*) as request_count
FROM classification_performance_metrics 
WHERE timestamp >= NOW() - INTERVAL '6 months'
GROUP BY DATE_TRUNC('month', timestamp), classification_method
ORDER BY month DESC, avg_accuracy DESC;

-- =============================================================================
-- 15. TEST SCALABILITY INDEXES
-- =============================================================================

-- Test scalable query patterns
EXPLAIN (ANALYZE, BUFFERS) 
SELECT i.id, i.name, i.category, i.is_active, i.created_at,
       COUNT(ik.id) as keyword_count
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name, i.category, i.is_active, i.created_at
ORDER BY i.created_at DESC;

-- =============================================================================
-- 16. TEST MACHINE LEARNING INDEXES
-- =============================================================================

-- Test ML-optimized queries
EXPLAIN (ANALYZE, BUFFERS) 
SELECT 
    classification_method,
    AVG(accuracy_score) as avg_accuracy,
    AVG(response_time_ms) as avg_response_time,
    COUNT(*) as sample_count
FROM classification_performance_metrics 
WHERE timestamp >= NOW() - INTERVAL '30 days'
GROUP BY classification_method
ORDER BY avg_accuracy DESC;

-- =============================================================================
-- 17. TEST BACKUP AND RECOVERY INDEXES
-- =============================================================================

-- Test backup-optimized queries
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM industries 
WHERE created_at >= NOW() - INTERVAL '1 day'
ORDER BY created_at, updated_at;

-- =============================================================================
-- 18. PERFORMANCE BENCHMARKING QUERIES
-- =============================================================================

-- Benchmark query performance
\timing on

-- Test 1: Simple lookup
SELECT COUNT(*) FROM industries WHERE is_active = true;

-- Test 2: Join query
SELECT COUNT(*) 
FROM industries i
JOIN industry_keywords ik ON i.id = ik.industry_id
WHERE i.is_active = true AND ik.is_active = true;

-- Test 3: Complex filtering
SELECT COUNT(*) 
FROM risk_keywords 
WHERE risk_category = 'prohibited' AND risk_severity = 'high' AND is_active = true;

-- Test 4: Full-text search
SELECT COUNT(*) 
FROM industries 
WHERE to_tsvector('english', name || ' ' || COALESCE(description, '')) @@ to_tsquery('english', 'technology');

-- Test 5: Array operations
SELECT COUNT(*) 
FROM risk_keywords 
WHERE mcc_codes @> ARRAY['6011'];

\timing off

-- =============================================================================
-- 19. INDEX USAGE STATISTICS
-- =============================================================================

-- Check index usage statistics
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as times_used,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size
FROM pg_stat_user_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
ORDER BY idx_scan DESC;

-- =============================================================================
-- 20. INDEX SIZE ANALYSIS
-- =============================================================================

-- Analyze index sizes
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size,
    pg_relation_size(indexname::regclass) as size_bytes
FROM pg_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
ORDER BY pg_relation_size(indexname::regclass) DESC;

-- =============================================================================
-- END OF INDEX PERFORMANCE TESTING
-- =============================================================================
