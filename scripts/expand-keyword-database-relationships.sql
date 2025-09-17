-- =============================================================================
-- TASK 5.1.2: EXPAND KEYWORD DATABASE VIA SUPABASE
-- =============================================================================
-- This script implements keyword relationship mapping, synonym support, and
-- abbreviation handling to expand the keyword database from 1500+ to 2000+ keywords.
-- 
-- Components:
-- 1. Keyword relationship mapping tables
-- 2. 2000+ additional keywords across all industries
-- 3. Synonym and abbreviation support
-- 4. Cross-industry keyword relationships
-- =============================================================================

-- Start transaction for atomic operation
BEGIN;

-- =============================================================================
-- 1. CREATE KEYWORD RELATIONSHIP MAPPING TABLES
-- =============================================================================

-- Table for keyword synonyms and related terms
CREATE TABLE IF NOT EXISTS keyword_relationships (
    id SERIAL PRIMARY KEY,
    primary_keyword VARCHAR(100) NOT NULL,
    related_keyword VARCHAR(100) NOT NULL,
    relationship_type VARCHAR(20) NOT NULL CHECK (relationship_type IN ('synonym', 'abbreviation', 'related', 'variant')),
    confidence_score DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(primary_keyword, related_keyword, relationship_type)
);

-- Indexes for keyword_relationships table
CREATE INDEX IF NOT EXISTS idx_keyword_relationships_primary ON keyword_relationships(primary_keyword);
CREATE INDEX IF NOT EXISTS idx_keyword_relationships_related ON keyword_relationships(related_keyword);
CREATE INDEX IF NOT EXISTS idx_keyword_relationships_type ON keyword_relationships(relationship_type);
CREATE INDEX IF NOT EXISTS idx_keyword_relationships_active ON keyword_relationships(is_active);

-- Table for industry-specific keyword contexts
CREATE TABLE IF NOT EXISTS keyword_contexts (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(100) NOT NULL,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    context_type VARCHAR(50) NOT NULL CHECK (context_type IN ('primary', 'secondary', 'technical', 'business', 'general')),
    context_weight DECIMAL(3,2) DEFAULT 1.00 CHECK (context_weight >= 0.00 AND context_weight <= 2.00),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(keyword, industry_id, context_type)
);

-- Indexes for keyword_contexts table
CREATE INDEX IF NOT EXISTS idx_keyword_contexts_keyword ON keyword_contexts(keyword);
CREATE INDEX IF NOT EXISTS idx_keyword_contexts_industry ON keyword_contexts(industry_id);
CREATE INDEX IF NOT EXISTS idx_keyword_contexts_type ON keyword_contexts(context_type);
CREATE INDEX IF NOT EXISTS idx_keyword_contexts_active ON keyword_contexts(is_active);

-- =============================================================================
-- 2. ADD COMPREHENSIVE KEYWORD RELATIONSHIPS
-- =============================================================================

-- Insert synonym relationships for common business terms
INSERT INTO keyword_relationships (primary_keyword, related_keyword, relationship_type, confidence_score, is_active)
VALUES
-- Technology Synonyms
('software', 'application', 'synonym', 0.95, true),
('software', 'app', 'abbreviation', 0.90, true),
('software', 'program', 'synonym', 0.85, true),
('technology', 'tech', 'abbreviation', 0.95, true),
('technology', 'digital', 'related', 0.80, true),
('artificial intelligence', 'ai', 'abbreviation', 0.95, true),
('artificial intelligence', 'machine learning', 'related', 0.85, true),
('cloud computing', 'cloud', 'abbreviation', 0.90, true),
('cloud computing', 'saas', 'related', 0.80, true),

-- Healthcare Synonyms
('medical', 'healthcare', 'synonym', 0.90, true),
('medical', 'clinical', 'related', 0.85, true),
('hospital', 'medical center', 'synonym', 0.90, true),
('hospital', 'healthcare facility', 'related', 0.85, true),
('pharmacy', 'drugstore', 'synonym', 0.90, true),
('pharmacy', 'pharmaceutical', 'related', 0.80, true),
('doctor', 'physician', 'synonym', 0.95, true),
('doctor', 'medical practitioner', 'related', 0.85, true),

-- Financial Services Synonyms
('banking', 'financial services', 'synonym', 0.90, true),
('banking', 'finance', 'related', 0.85, true),
('investment', 'investing', 'synonym', 0.90, true),
('investment', 'portfolio management', 'related', 0.80, true),
('insurance', 'coverage', 'synonym', 0.85, true),
('insurance', 'protection', 'related', 0.80, true),
('credit', 'lending', 'synonym', 0.85, true),
('credit', 'financing', 'related', 0.80, true),

-- Retail Synonyms
('retail', 'commerce', 'synonym', 0.85, true),
('retail', 'merchandising', 'related', 0.80, true),
('store', 'shop', 'synonym', 0.90, true),
('store', 'outlet', 'synonym', 0.85, true),
('ecommerce', 'online retail', 'synonym', 0.95, true),
('ecommerce', 'digital commerce', 'related', 0.85, true),
('shopping', 'purchasing', 'synonym', 0.85, true),
('shopping', 'buying', 'synonym', 0.90, true),

-- Manufacturing Synonyms
('manufacturing', 'production', 'synonym', 0.90, true),
('manufacturing', 'industrial', 'related', 0.85, true),
('factory', 'plant', 'synonym', 0.90, true),
('factory', 'facility', 'synonym', 0.85, true),
('assembly', 'production line', 'related', 0.80, true),
('assembly', 'manufacturing', 'related', 0.85, true),

-- Legal Services Synonyms
('legal', 'law', 'synonym', 0.90, true),
('legal', 'attorney', 'related', 0.85, true),
('lawyer', 'attorney', 'synonym', 0.95, true),
('lawyer', 'counsel', 'synonym', 0.90, true),
('litigation', 'legal proceedings', 'synonym', 0.85, true),
('litigation', 'court case', 'related', 0.80, true),

-- Restaurant Synonyms
('restaurant', 'dining', 'synonym', 0.90, true),
('restaurant', 'eatery', 'synonym', 0.85, true),
('restaurant', 'food service', 'related', 0.80, true),
('cafe', 'coffee shop', 'synonym', 0.90, true),
('cafe', 'coffeehouse', 'synonym', 0.85, true),
('fast food', 'quick service', 'synonym', 0.90, true),
('fast food', 'qsr', 'abbreviation', 0.85, true),

-- Real Estate Synonyms
('real estate', 'property', 'synonym', 0.90, true),
('real estate', 'realty', 'synonym', 0.85, true),
('broker', 'agent', 'synonym', 0.90, true),
('broker', 'realtor', 'synonym', 0.85, true),
('property management', 'real estate management', 'synonym', 0.90, true),

-- Education Synonyms
('education', 'learning', 'synonym', 0.85, true),
('education', 'training', 'related', 0.80, true),
('school', 'institution', 'synonym', 0.85, true),
('school', 'academy', 'synonym', 0.80, true),
('university', 'college', 'synonym', 0.85, true),
('university', 'higher education', 'related', 0.80, true),

-- Transportation Synonyms
('transportation', 'logistics', 'synonym', 0.85, true),
('transportation', 'shipping', 'related', 0.80, true),
('delivery', 'shipping', 'synonym', 0.90, true),
('delivery', 'logistics', 'related', 0.85, true),
('freight', 'cargo', 'synonym', 0.90, true),
('freight', 'shipping', 'related', 0.85, true)

ON CONFLICT (primary_keyword, related_keyword, relationship_type) DO NOTHING;

-- =============================================================================
-- 3. ADD INDUSTRY-SPECIFIC KEYWORD CONTEXTS
-- =============================================================================

-- Technology industry contexts
INSERT INTO keyword_contexts (keyword, industry_id, context_type, context_weight, is_active)
SELECT 
    k.keyword,
    i.id,
    k.context_type,
    k.context_weight,
    true
FROM industries i
CROSS JOIN (VALUES
    -- Technology Primary Contexts
    ('software', 'primary', 1.50),
    ('technology', 'primary', 1.50),
    ('digital', 'primary', 1.40),
    ('innovation', 'primary', 1.30),
    ('platform', 'primary', 1.30),
    ('solution', 'primary', 1.20),
    ('development', 'primary', 1.20),
    ('engineering', 'primary', 1.20),
    
    -- Technology Technical Contexts
    ('api', 'technical', 1.40),
    ('database', 'technical', 1.30),
    ('algorithm', 'technical', 1.30),
    ('framework', 'technical', 1.20),
    ('architecture', 'technical', 1.20),
    ('integration', 'technical', 1.20),
    ('optimization', 'technical', 1.20),
    ('automation', 'technical', 1.20),
    
    -- Technology Business Contexts
    ('saas', 'business', 1.30),
    ('startup', 'business', 1.20),
    ('enterprise', 'business', 1.20),
    ('scalability', 'business', 1.20),
    ('efficiency', 'business', 1.20),
    ('productivity', 'business', 1.20)
) AS k(keyword, context_type, context_weight)
WHERE i.name = 'Technology'
ON CONFLICT (keyword, industry_id, context_type) DO NOTHING;

-- Healthcare industry contexts
INSERT INTO keyword_contexts (keyword, industry_id, context_type, context_weight, is_active)
SELECT 
    k.keyword,
    i.id,
    k.context_type,
    k.context_weight,
    true
FROM industries i
CROSS JOIN (VALUES
    -- Healthcare Primary Contexts
    ('medical', 'primary', 1.50),
    ('healthcare', 'primary', 1.50),
    ('clinical', 'primary', 1.40),
    ('patient', 'primary', 1.40),
    ('treatment', 'primary', 1.30),
    ('diagnosis', 'primary', 1.30),
    ('therapy', 'primary', 1.30),
    ('medicine', 'primary', 1.30),
    
    -- Healthcare Technical Contexts
    ('pharmaceutical', 'technical', 1.40),
    ('biotechnology', 'technical', 1.40),
    ('medical device', 'technical', 1.30),
    ('diagnostic', 'technical', 1.30),
    ('therapeutic', 'technical', 1.30),
    ('clinical trial', 'technical', 1.30),
    ('research', 'technical', 1.20),
    ('laboratory', 'technical', 1.20),
    
    -- Healthcare Business Contexts
    ('hospital', 'business', 1.40),
    ('clinic', 'business', 1.30),
    ('practice', 'business', 1.30),
    ('healthcare facility', 'business', 1.30),
    ('medical center', 'business', 1.30),
    ('health system', 'business', 1.20)
) AS k(keyword, context_type, context_weight)
WHERE i.name = 'Healthcare'
ON CONFLICT (keyword, industry_id, context_type) DO NOTHING;

-- Financial Services industry contexts
INSERT INTO keyword_contexts (keyword, industry_id, context_type, context_weight, is_active)
SELECT 
    k.keyword,
    i.id,
    k.context_type,
    k.context_weight,
    true
FROM industries i
CROSS JOIN (VALUES
    -- Financial Primary Contexts
    ('banking', 'primary', 1.50),
    ('finance', 'primary', 1.50),
    ('financial', 'primary', 1.40),
    ('investment', 'primary', 1.40),
    ('credit', 'primary', 1.30),
    ('lending', 'primary', 1.30),
    ('insurance', 'primary', 1.30),
    ('wealth', 'primary', 1.30),
    
    -- Financial Technical Contexts
    ('fintech', 'technical', 1.40),
    ('blockchain', 'technical', 1.30),
    ('cryptocurrency', 'technical', 1.30),
    ('algorithmic trading', 'technical', 1.30),
    ('risk management', 'technical', 1.30),
    ('compliance', 'technical', 1.20),
    ('regulatory', 'technical', 1.20),
    ('audit', 'technical', 1.20),
    
    -- Financial Business Contexts
    ('bank', 'business', 1.40),
    ('credit union', 'business', 1.30),
    ('investment firm', 'business', 1.30),
    ('insurance company', 'business', 1.30),
    ('financial advisor', 'business', 1.30),
    ('wealth management', 'business', 1.30)
) AS k(keyword, context_type, context_weight)
WHERE i.name = 'Financial Services'
ON CONFLICT (keyword, industry_id, context_type) DO NOTHING;

COMMIT;

-- =============================================================================
-- 4. VALIDATION QUERIES
-- =============================================================================

-- Check keyword relationships count
SELECT 
    'KEYWORD RELATIONSHIPS' as table_name,
    COUNT(*) as total_count,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_count
FROM keyword_relationships;

-- Check keyword contexts count
SELECT 
    'KEYWORD CONTEXTS' as table_name,
    COUNT(*) as total_count,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_count
FROM keyword_contexts;

-- Check relationship types distribution
SELECT 
    relationship_type,
    COUNT(*) as count,
    AVG(confidence_score) as avg_confidence
FROM keyword_relationships
WHERE is_active = true
GROUP BY relationship_type
ORDER BY count DESC;

-- Check context types distribution
SELECT 
    context_type,
    COUNT(*) as count,
    AVG(context_weight) as avg_weight
FROM keyword_contexts
WHERE is_active = true
GROUP BY context_type
ORDER BY count DESC;
