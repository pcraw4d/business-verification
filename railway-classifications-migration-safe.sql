-- =====================================================
-- Railway Classifications Table Migration (Safe Version)
-- KYB Platform - Railway Server Integration
-- =====================================================
-- 
-- This script creates the classifications table that the Railway server
-- expects to exist for storing business classification results.
-- This version handles existing objects gracefully.
--
-- Author: KYB Platform Development Team
-- Date: January 22, 2025
-- Version: 1.1 (Safe)
-- 
-- Dependencies:
-- - Supabase database with uuid-ossp extension
-- =====================================================

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- CLASSIFICATIONS TABLE
-- =====================================================
-- Create the classifications table that Railway server expects

CREATE TABLE IF NOT EXISTS classifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id VARCHAR(255) NOT NULL,
    business_name VARCHAR(500) NOT NULL,
    description TEXT,
    website_url VARCHAR(1000),
    classification JSONB NOT NULL, -- Stores the full classification object
    confidence_score DECIMAL(3,2) NOT NULL CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure unique business_id
    UNIQUE(business_id)
);

-- =====================================================
-- MERCHANTS TABLE (for Railway server merchant endpoints)
-- =====================================================
-- Create merchants table for Railway server merchant endpoints

CREATE TABLE IF NOT EXISTS merchants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    industry VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    description TEXT,
    website_url VARCHAR(1000),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================================================
-- MOCK_MERCHANTS TABLE (fallback for Railway server)
-- =====================================================
-- Create mock_merchants table for Railway server fallback

CREATE TABLE IF NOT EXISTS mock_merchants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    industry VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    description TEXT,
    website_url VARCHAR(1000),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Classifications indexes
CREATE INDEX IF NOT EXISTS idx_classifications_business_id ON classifications(business_id);
CREATE INDEX IF NOT EXISTS idx_classifications_business_name ON classifications(business_name);
CREATE INDEX IF NOT EXISTS idx_classifications_created_at ON classifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_classifications_confidence_score ON classifications(confidence_score DESC);

-- Merchants indexes
CREATE INDEX IF NOT EXISTS idx_merchants_name ON merchants(name);
CREATE INDEX IF NOT EXISTS idx_merchants_industry ON merchants(industry);
CREATE INDEX IF NOT EXISTS idx_merchants_status ON merchants(status);
CREATE INDEX IF NOT EXISTS idx_merchants_created_at ON merchants(created_at DESC);

-- Mock merchants indexes
CREATE INDEX IF NOT EXISTS idx_mock_merchants_name ON mock_merchants(name);
CREATE INDEX IF NOT EXISTS idx_mock_merchants_industry ON mock_merchants(industry);
CREATE INDEX IF NOT EXISTS idx_mock_merchants_status ON mock_merchants(status);

-- =====================================================
-- ROW LEVEL SECURITY (RLS) POLICIES
-- =====================================================

-- Enable RLS on all tables
ALTER TABLE classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE merchants ENABLE ROW LEVEL SECURITY;
ALTER TABLE mock_merchants ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist (to avoid conflicts)
DROP POLICY IF EXISTS "Allow public read access to classifications" ON classifications;
DROP POLICY IF EXISTS "Allow public read access to merchants" ON merchants;
DROP POLICY IF EXISTS "Allow public read access to mock_merchants" ON mock_merchants;
DROP POLICY IF EXISTS "Allow authenticated users to manage classifications" ON classifications;
DROP POLICY IF EXISTS "Allow authenticated users to manage merchants" ON merchants;
DROP POLICY IF EXISTS "Allow authenticated users to manage mock_merchants" ON mock_merchants;

-- Create policies for public read access
CREATE POLICY "Allow public read access to classifications" ON classifications
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to merchants" ON merchants
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to mock_merchants" ON mock_merchants
    FOR SELECT USING (true);

-- Create policies for authenticated write access
CREATE POLICY "Allow authenticated users to manage classifications" ON classifications
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage merchants" ON merchants
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage mock_merchants" ON mock_merchants
    FOR ALL USING (auth.role() = 'authenticated');

-- =====================================================
-- TRIGGERS FOR UPDATED AT TIMESTAMPS
-- =====================================================

-- Create or replace the update_updated_at_column function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Drop existing triggers if they exist (to avoid conflicts)
DROP TRIGGER IF EXISTS update_classifications_updated_at ON classifications;
DROP TRIGGER IF EXISTS update_merchants_updated_at ON merchants;
DROP TRIGGER IF EXISTS update_mock_merchants_updated_at ON mock_merchants;

-- Create triggers for updated_at columns
CREATE TRIGGER update_classifications_updated_at 
    BEFORE UPDATE ON classifications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_merchants_updated_at 
    BEFORE UPDATE ON merchants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mock_merchants_updated_at 
    BEFORE UPDATE ON mock_merchants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- SAMPLE DATA INSERTION
-- =====================================================

-- Insert sample merchants
INSERT INTO merchants (id, name, industry, status, description) VALUES
('merch_1', 'Acme Technology Corp', 'Technology', 'active', 'Leading software development company'),
('merch_2', 'Global Retail Solutions', 'Retail', 'active', 'E-commerce platform provider'),
('merch_3', 'HealthTech Innovations', 'Healthcare', 'active', 'Medical technology solutions'),
('merch_4', 'FinanceFlow Systems', 'Finance', 'inactive', 'Financial services platform')
ON CONFLICT (id) DO NOTHING;

-- Insert sample mock merchants
INSERT INTO mock_merchants (id, name, industry, status, description) VALUES
('mock_1', 'Mock Technology Company', 'Technology', 'active', 'Sample technology business'),
('mock_2', 'Mock Retail Store', 'Retail', 'active', 'Sample retail business'),
('mock_3', 'Mock Healthcare Provider', 'Healthcare', 'inactive', 'Sample healthcare business')
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- COMMENTS FOR DOCUMENTATION
-- =====================================================

COMMENT ON TABLE classifications IS 'Business classification results stored by Railway server with full classification objects and confidence scores';
COMMENT ON TABLE merchants IS 'Merchant data for Railway server merchant endpoints';
COMMENT ON TABLE mock_merchants IS 'Mock merchant data for Railway server fallback functionality';

-- Column comments
COMMENT ON COLUMN classifications.classification IS 'Full classification object with industry, codes, and metadata in JSONB format';
COMMENT ON COLUMN classifications.confidence_score IS 'Confidence score for the classification (0.00-1.00)';

-- =====================================================
-- VERIFICATION QUERIES
-- =====================================================

-- Verify table creation
SELECT 
    table_name, 
    table_type,
    CASE 
        WHEN table_name IN ('classifications', 'merchants', 'mock_merchants') 
        THEN '✅ Created'
        ELSE '❌ Missing'
    END as status
FROM information_schema.tables 
WHERE table_schema = 'public' 
    AND table_name IN ('classifications', 'merchants', 'mock_merchants')
ORDER BY table_name;

-- Verify sample data
SELECT 'merchants' as table_name, COUNT(*) as record_count FROM merchants
UNION ALL
SELECT 'mock_merchants' as table_name, COUNT(*) as record_count FROM mock_merchants;

-- =====================================================
-- COMPLETION MESSAGE
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Railway Classifications Migration (Safe) Completed Successfully!';
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Tables Created/Updated:';
    RAISE NOTICE '  ✅ classifications (for Railway server classification storage)';
    RAISE NOTICE '  ✅ merchants (for Railway server merchant endpoints)';
    RAISE NOTICE '  ✅ mock_merchants (for Railway server fallback)';
    RAISE NOTICE '';
    RAISE NOTICE 'Features Added/Updated:';
    RAISE NOTICE '  ✅ Comprehensive indexing for performance';
    RAISE NOTICE '  ✅ Row Level Security (RLS) policies';
    RAISE NOTICE '  ✅ Updated_at triggers (safely handled existing ones)';
    RAISE NOTICE '  ✅ Sample data for testing';
    RAISE NOTICE '';
    RAISE NOTICE 'Railway server should now be able to store and retrieve classifications!';
    RAISE NOTICE '=====================================================';
END $$;
