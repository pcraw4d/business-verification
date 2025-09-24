-- Database Optimization for Keyword Classification System
-- This file contains recommended database indexes and optimizations for the KYB Platform

-- =============================================================================
-- KEYWORD WEIGHTS TABLE OPTIMIZATIONS
-- =============================================================================

-- Primary index for keyword lookups (most important)
CREATE INDEX IF NOT EXISTS idx_keyword_weights_keyword_active 
ON keyword_weights (keyword, is_active) 
WHERE is_active = true;

-- Index for industry-based keyword queries
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_active 
ON keyword_weights (industry_id, is_active, base_weight DESC) 
WHERE is_active = true;

-- Composite index for keyword search with weight ordering
CREATE INDEX IF NOT EXISTS idx_keyword_weights_search 
ON keyword_weights (is_active, base_weight DESC, keyword) 
WHERE is_active = true;

-- Index for usage count tracking
CREATE INDEX IF NOT EXISTS idx_keyword_weights_usage 
ON keyword_weights (usage_count DESC, last_updated) 
WHERE is_active = true;

-- =============================================================================
-- CLASSIFICATION CODES TABLE OPTIMIZATIONS
-- =============================================================================

-- Primary index for industry-based code lookups
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_active 
ON classification_codes (industry_id, is_active, code_type) 
WHERE is_active = true;

-- Index for code type queries (NAICS, SIC, MCC)
CREATE INDEX IF NOT EXISTS idx_classification_codes_type_active 
ON classification_codes (code_type, is_active, industry_id) 
WHERE is_active = true;

-- Composite index for code lookups with ordering
CREATE INDEX IF NOT EXISTS idx_classification_codes_lookup 
ON classification_codes (is_active, code_type, code, industry_id) 
WHERE is_active = true;

-- =============================================================================
-- INDUSTRIES TABLE OPTIMIZATIONS
-- =============================================================================

-- Primary index for industry lookups
CREATE INDEX IF NOT EXISTS idx_industries_active 
ON industries (id, is_active) 
WHERE is_active = true;

-- Index for industry name searches
CREATE INDEX IF NOT EXISTS idx_industries_name_active 
ON industries (name, is_active) 
WHERE is_active = true;

-- Index for category-based queries
CREATE INDEX IF NOT EXISTS idx_industries_category_active 
ON industries (category, is_active, confidence_threshold) 
WHERE is_active = true;

-- =============================================================================
-- INDUSTRY KEYWORDS TABLE OPTIMIZATIONS
-- =============================================================================

-- Primary index for keyword searches
CREATE INDEX IF NOT EXISTS idx_industry_keywords_search 
ON industry_keywords (keyword, is_active, weight DESC) 
WHERE is_active = true;

-- Index for industry-based keyword queries
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry 
ON industry_keywords (industry_id, is_active, weight DESC) 
WHERE is_active = true;

-- =============================================================================
-- INDUSTRY PATTERNS TABLE OPTIMIZATIONS
-- =============================================================================

-- Index for pattern-based industry searches
CREATE INDEX IF NOT EXISTS idx_industry_patterns_industry_active 
ON industry_patterns (industry_id, is_active, confidence_score DESC) 
WHERE is_active = true;

-- Index for pattern type queries
CREATE INDEX IF NOT EXISTS idx_industry_patterns_type 
ON industry_patterns (pattern_type, is_active, confidence_score DESC) 
WHERE is_active = true;

-- =============================================================================
-- FULL-TEXT SEARCH OPTIMIZATIONS
-- =============================================================================

-- Full-text search index for keywords (if supported)
-- CREATE INDEX IF NOT EXISTS idx_keyword_weights_fts 
-- ON keyword_weights USING gin(to_tsvector('english', keyword)) 
-- WHERE is_active = true;

-- Full-text search index for industry names
-- CREATE INDEX IF NOT EXISTS idx_industries_name_fts 
-- ON industries USING gin(to_tsvector('english', name)) 
-- WHERE is_active = true;

-- Full-text search index for industry descriptions
-- CREATE INDEX IF NOT EXISTS idx_industries_description_fts 
-- ON industries USING gin(to_tsvector('english', description)) 
-- WHERE is_active = true;

-- =============================================================================
-- PARTIAL INDEXES FOR PERFORMANCE
-- =============================================================================

-- Partial index for high-weight keywords (most important for classification)
CREATE INDEX IF NOT EXISTS idx_keyword_weights_high_weight 
ON keyword_weights (keyword, industry_id, base_weight) 
WHERE is_active = true AND base_weight > 0.5;

-- Partial index for frequently used keywords
CREATE INDEX IF NOT EXISTS idx_keyword_weights_frequent 
ON keyword_weights (keyword, industry_id, usage_count DESC) 
WHERE is_active = true AND usage_count > 10;

-- Partial index for recent keywords
CREATE INDEX IF NOT EXISTS idx_keyword_weights_recent 
ON keyword_weights (keyword, industry_id, last_updated DESC) 
WHERE is_active = true AND last_updated > NOW() - INTERVAL '30 days';

-- =============================================================================
-- QUERY OPTIMIZATION STATISTICS
-- =============================================================================

-- Update table statistics for better query planning
ANALYZE keyword_weights;
ANALYZE classification_codes;
ANALYZE industries;
ANALYZE industry_keywords;
ANALYZE industry_patterns;

-- =============================================================================
-- CONNECTION POOLING RECOMMENDATIONS
-- =============================================================================

-- Recommended PostgreSQL configuration for optimal performance:
-- 
-- # Connection Settings
-- max_connections = 200
-- shared_buffers = 256MB
-- effective_cache_size = 1GB
-- 
-- # Query Planning
-- random_page_cost = 1.1
-- effective_io_concurrency = 200
-- 
-- # Memory Settings
-- work_mem = 4MB
-- maintenance_work_mem = 64MB
-- 
-- # Checkpoint Settings
-- checkpoint_completion_target = 0.9
-- wal_buffers = 16MB
-- 
-- # Logging (for monitoring)
-- log_min_duration_statement = 1000
-- log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '

-- =============================================================================
-- MONITORING QUERIES
-- =============================================================================

-- Query to monitor index usage
-- SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
-- FROM pg_stat_user_indexes 
-- WHERE schemaname = 'public' 
-- ORDER BY idx_scan DESC;

-- Query to monitor slow queries
-- SELECT query, calls, total_time, mean_time, rows
-- FROM pg_stat_statements 
-- WHERE mean_time > 1000 
-- ORDER BY mean_time DESC 
-- LIMIT 10;

-- Query to monitor table sizes
-- SELECT schemaname, tablename, 
--        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
-- FROM pg_tables 
-- WHERE schemaname = 'public' 
-- ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
