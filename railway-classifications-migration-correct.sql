-- =====================================================
-- Railway Classifications Table Migration (Correct Version)
-- KYB Platform - Railway Server Integration
-- =====================================================
-- 
-- This script creates all required tables with the EXACT schema
-- that the Railway server expects.
--
-- Author: KYB Platform Development Team
-- Date: January 22, 2025
-- Version: 1.4 (Correct)
-- =====================================================

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- DROP AND RECREATE TABLES WITH CORRECT SCHEMA
-- =====================================================

-- Drop existing tables if they exist (to fix any schema issues)
DROP TABLE IF EXISTS classifications CASCADE;
DROP TABLE IF EXISTS merchants CASCADE;
DROP TABLE IF EXISTS mock_merchants CASCADE;

-- Create classifications table with correct schema
CREATE TABLE classifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id VARCHAR(255) NOT NULL,
    business_name VARCHAR(500) NOT NULL,
    description TEXT,
    website_url VARCHAR(1000),
    classification JSONB NOT NULL,
    confidence_score DECIMAL(3,2) NOT NULL CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(business_id)
);

-- Create merchants table with STRING id (not UUID)
CREATE TABLE merchants (
    id VARCHAR(255) PRIMARY KEY,  -- STRING ID, not UUID
    name VARCHAR(500) NOT NULL,
    industry VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    description TEXT,
    website_url VARCHAR(1000),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create mock_merchants table with STRING id (not UUID)
CREATE TABLE mock_merchants (
    id VARCHAR(255) PRIMARY KEY,  -- STRING ID, not UUID
    name VARCHAR(500) NOT NULL,
    industry VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    description TEXT,
    website_url VARCHAR(1000),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================================================
-- CREATE INDEXES
-- =====================================================

-- Classifications indexes
CREATE INDEX idx_classifications_business_id ON classifications(business_id);
CREATE INDEX idx_classifications_business_name ON classifications(business_name);
CREATE INDEX idx_classifications_created_at ON classifications(created_at DESC);
CREATE INDEX idx_classifications_confidence_score ON classifications(confidence_score DESC);

-- Merchants indexes
CREATE INDEX idx_merchants_name ON merchants(name);
CREATE INDEX idx_merchants_industry ON merchants(industry);
CREATE INDEX idx_merchants_status ON merchants(status);
CREATE INDEX idx_merchants_created_at ON merchants(created_at DESC);

-- Mock merchants indexes
CREATE INDEX idx_mock_merchants_name ON mock_merchants(name);
CREATE INDEX idx_mock_merchants_industry ON mock_merchants(industry);
CREATE INDEX idx_mock_merchants_status ON mock_merchants(status);

-- =====================================================
-- SETUP ROW LEVEL SECURITY
-- =====================================================

-- Enable RLS on all tables
ALTER TABLE classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE merchants ENABLE ROW LEVEL SECURITY;
ALTER TABLE mock_merchants ENABLE ROW LEVEL SECURITY;

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
-- CREATE TRIGGERS
-- =====================================================

-- Create or replace the update_updated_at_column function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

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
-- INSERT SAMPLE DATA
-- =====================================================

-- Insert sample merchants with STRING IDs
INSERT INTO merchants (id, name, industry, status, description) VALUES
('merch_1', 'Acme Technology Corp', 'Technology', 'active', 'Leading software development company'),
('merch_2', 'Global Retail Solutions', 'Retail', 'active', 'E-commerce platform provider'),
('merch_3', 'HealthTech Innovations', 'Healthcare', 'active', 'Medical technology solutions'),
('merch_4', 'FinanceFlow Systems', 'Finance', 'inactive', 'Financial services platform');

-- Insert sample mock merchants with STRING IDs
INSERT INTO mock_merchants (id, name, industry, status, description) VALUES
('mock_1', 'Mock Technology Company', 'Technology', 'active', 'Sample technology business'),
('mock_2', 'Mock Retail Store', 'Retail', 'active', 'Sample retail business'),
('mock_3', 'Mock Healthcare Provider', 'Healthcare', 'inactive', 'Sample healthcare business');

-- =====================================================
-- VERIFICATION
-- =====================================================

-- Check table structure
SELECT 
    table_name,
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_name IN ('classifications', 'merchants', 'mock_merchants')
    AND table_schema = 'public'
ORDER BY table_name, ordinal_position;

-- Check sample data
SELECT 'merchants' as table_name, COUNT(*) as record_count FROM merchants
UNION ALL
SELECT 'mock_merchants' as table_name, COUNT(*) as record_count FROM mock_merchants;

-- Test that we can query by string ID
SELECT id, name, industry FROM merchants WHERE id = 'merch_1';
SELECT id, name, industry FROM mock_merchants WHERE id = 'mock_1';

-- =====================================================
-- COMPLETION MESSAGE
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Railway Classifications Migration (Correct) Completed Successfully!';
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Tables Created with Correct Schema:';
    RAISE NOTICE '  ✅ classifications (UUID id for internal use)';
    RAISE NOTICE '  ✅ merchants (VARCHAR id for Railway server)';
    RAISE NOTICE '  ✅ mock_merchants (VARCHAR id for Railway server)';
    RAISE NOTICE '';
    RAISE NOTICE 'Key Fixes Applied:';
    RAISE NOTICE '  ✅ merchants.id is VARCHAR(255) - can store "merch_1"';
    RAISE NOTICE '  ✅ mock_merchants.id is VARCHAR(255) - can store "mock_1"';
    RAISE NOTICE '  ✅ All sample data inserted successfully';
    RAISE NOTICE '  ✅ All indexes and policies created';
    RAISE NOTICE '';
    RAISE NOTICE 'Railway server should now work perfectly!';
    RAISE NOTICE '=====================================================';
END $$;
