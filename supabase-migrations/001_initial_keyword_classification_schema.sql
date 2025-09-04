-- =====================================================
-- Keyword Classification System Schema Migration
-- Supabase Implementation
-- =====================================================

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- 1. industries Table
-- =====================================================
CREATE TABLE industries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('traditional', 'emerging', 'hybrid')),
    parent_industry_id INTEGER REFERENCES industries(id),
    confidence_threshold DECIMAL(3,2) DEFAULT 0.80,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for industries table
CREATE INDEX idx_industries_name ON industries(name);
CREATE INDEX idx_industries_category ON industries(category);
CREATE INDEX idx_industries_active ON industries(is_active);
CREATE INDEX idx_industries_parent ON industries(parent_industry_id);

-- =====================================================
-- 2. industry_keywords Table
-- =====================================================
CREATE TABLE industry_keywords (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    weight DECIMAL(3,2) DEFAULT 1.00 CHECK (weight >= 0.00 AND weight <= 1.00),
    context VARCHAR(50) CHECK (context IN ('business', 'technical', 'general')),
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(industry_id, keyword)
);

-- Indexes for industry_keywords table
CREATE INDEX idx_industry_keywords_industry ON industry_keywords(industry_id);
CREATE INDEX idx_industry_keywords_keyword ON industry_keywords(keyword);
CREATE INDEX idx_industry_keywords_weight ON industry_keywords(weight);
CREATE INDEX idx_industry_keywords_primary ON industry_keywords(is_primary);
CREATE INDEX idx_industry_keywords_context ON industry_keywords(context);

-- =====================================================
-- 3. classification_codes Table
-- =====================================================
CREATE TABLE classification_codes (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('NAICS', 'MCC', 'SIC')),
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    confidence DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence >= 0.00 AND confidence <= 1.00),
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(code_type, code)
);

-- Indexes for classification_codes table
CREATE INDEX idx_classification_codes_industry ON classification_codes(industry_id);
CREATE INDEX idx_classification_codes_type ON classification_codes(code_type);
CREATE INDEX idx_classification_codes_code ON classification_codes(code);
CREATE INDEX idx_classification_codes_primary ON classification_codes(is_primary);

-- =====================================================
-- 4. code_keywords Table
-- =====================================================
CREATE TABLE code_keywords (
    id SERIAL PRIMARY KEY,
    code_id INTEGER NOT NULL REFERENCES classification_codes(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    relevance_score DECIMAL(3,2) DEFAULT 1.00 CHECK (relevance_score >= 0.00 AND relevance_score <= 1.00),
    match_type VARCHAR(20) DEFAULT 'exact' CHECK (match_type IN ('exact', 'partial', 'synonym')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(code_id, keyword)
);

-- Indexes for code_keywords table
CREATE INDEX idx_code_keywords_code ON code_keywords(code_id);
CREATE INDEX idx_code_keywords_keyword ON code_keywords(keyword);
CREATE INDEX idx_code_keywords_relevance ON code_keywords(relevance_score);

-- =====================================================
-- 5. industry_patterns Table
-- =====================================================
CREATE TABLE industry_patterns (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    pattern_type VARCHAR(50) NOT NULL CHECK (pattern_type IN ('regex', 'phrase', 'semantic', 'context')),
    pattern_data TEXT NOT NULL,
    priority INTEGER DEFAULT 1 CHECK (priority >= 1 AND priority <= 10),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for industry_patterns table
CREATE INDEX idx_industry_patterns_industry ON industry_patterns(industry_id);
CREATE INDEX idx_industry_patterns_type ON industry_patterns(pattern_type);
CREATE INDEX idx_industry_patterns_active ON industry_patterns(is_active);
CREATE INDEX idx_industry_patterns_priority ON industry_patterns(priority);

-- =====================================================
-- 6. keyword_weights Table
-- =====================================================
CREATE TABLE keyword_weights (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    base_weight DECIMAL(3,2) DEFAULT 1.00 CHECK (base_weight >= 0.00 AND base_weight <= 1.00),
    context_multiplier DECIMAL(3,2) DEFAULT 1.00 CHECK (context_multiplier >= 0.00 AND context_multiplier <= 2.00),
    frequency_boost DECIMAL(3,2) DEFAULT 1.00 CHECK (frequency_boost >= 0.00 AND frequency_boost <= 2.00),
    recency_factor DECIMAL(3,2) DEFAULT 1.00 CHECK (recency_factor >= 0.00 AND recency_factor <= 2.00),
    calculated_weight DECIMAL(3,2) GENERATED ALWAYS AS (
        base_weight * context_multiplier * frequency_boost * recency_factor
    ) STORED,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(industry_id, keyword)
);

-- Indexes for keyword_weights table
CREATE INDEX idx_keyword_weights_industry ON keyword_weights(industry_id);
CREATE INDEX idx_keyword_weights_calculated ON keyword_weights(calculated_weight);
CREATE INDEX idx_keyword_weights_keyword ON keyword_weights(keyword);

-- =====================================================
-- 7. audit_logs Table
-- =====================================================
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    record_id INTEGER NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values JSONB,
    new_values JSONB,
    user_id VARCHAR(100),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for audit_logs table
CREATE INDEX idx_audit_logs_table ON audit_logs(table_name);
CREATE INDEX idx_audit_logs_record ON audit_logs(record_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);

-- =====================================================
-- Row Level Security (RLS) Policies
-- =====================================================

-- Enable RLS on all tables
ALTER TABLE industries ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_keywords ENABLE ROW LEVEL SECURITY;
ALTER TABLE classification_codes ENABLE ROW LEVEL SECURITY;
ALTER TABLE code_keywords ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_patterns ENABLE ROW LEVEL SECURITY;
ALTER TABLE keyword_weights ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- Create policies for read access (public read, authenticated write)
CREATE POLICY "Allow public read access to industries" ON industries
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to industry_keywords" ON industry_keywords
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to classification_codes" ON classification_codes
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to code_keywords" ON code_keywords
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to industry_patterns" ON industry_patterns
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to keyword_weights" ON keyword_weights
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to audit_logs" ON audit_logs
    FOR SELECT USING (true);

-- Create policies for authenticated write access
CREATE POLICY "Allow authenticated users to manage industries" ON industries
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage industry_keywords" ON industry_keywords
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage classification_codes" ON classification_codes
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage code_keywords" ON code_keywords
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage industry_patterns" ON industry_patterns
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage keyword_weights" ON keyword_weights
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage audit_logs" ON audit_logs
    FOR ALL USING (auth.role() = 'authenticated');

-- =====================================================
-- Triggers for Updated At Timestamps
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_industries_updated_at 
    BEFORE UPDATE ON industries 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_industry_keywords_updated_at 
    BEFORE UPDATE ON industry_keywords 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_classification_codes_updated_at 
    BEFORE UPDATE ON classification_codes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_industry_patterns_updated_at 
    BEFORE UPDATE ON industry_patterns 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_keyword_weights_updated_at 
    BEFORE UPDATE ON keyword_weights 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- Comments for Documentation
-- =====================================================

COMMENT ON TABLE industries IS 'Core industry definitions for business classification';
COMMENT ON TABLE industry_keywords IS 'Keywords associated with each industry for classification';
COMMENT ON TABLE classification_codes IS 'NAICS, MCC, and SIC codes mapped to industries';
COMMENT ON TABLE code_keywords IS 'Keywords specific to each classification code';
COMMENT ON TABLE industry_patterns IS 'Advanced pattern matching rules for industry detection';
COMMENT ON TABLE keyword_weights IS 'Dynamic keyword importance scoring and weighting';
COMMENT ON TABLE audit_logs IS 'Change tracking and versioning for all tables';

-- =====================================================
-- Migration Complete
-- =====================================================

-- Verify schema creation
SELECT 
    table_name, 
    table_type 
FROM information_schema.tables 
WHERE table_schema = 'public' 
    AND table_name IN (
        'industries', 
        'industry_keywords', 
        'classification_codes', 
        'code_keywords', 
        'industry_patterns', 
        'keyword_weights', 
        'audit_logs'
    )
ORDER BY table_name;
