-- E2E Test Database Initialization Script
-- This script sets up the database schema and test data for E2E tests

-- Create test database if it doesn't exist
CREATE DATABASE IF NOT EXISTS kyb_e2e_test;

-- Use the test database
\c kyb_e2e_test;

-- Create test tables
CREATE TABLE IF NOT EXISTS test_businesses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    industry VARCHAR(100),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS test_classifications (
    id SERIAL PRIMARY KEY,
    business_id INTEGER REFERENCES test_businesses(id),
    industry_code VARCHAR(20),
    industry_name VARCHAR(255),
    confidence_score DECIMAL(5,4),
    classification_method VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS test_keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(100) NOT NULL,
    industry_code VARCHAR(20),
    weight DECIMAL(5,4) DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert test data
INSERT INTO test_businesses (name, industry, description) VALUES
('Tech Solutions Inc', 'Technology', 'Software development and consulting services'),
('Acme Corporation', 'Manufacturing', 'Industrial equipment manufacturing'),
('Global Services Ltd', 'Services', 'Business consulting and advisory services'),
('Digital Marketing Co', 'Marketing', 'Digital marketing and advertising services'),
('Healthcare Partners', 'Healthcare', 'Medical equipment and healthcare services');

INSERT INTO test_keywords (keyword, industry_code, weight) VALUES
('software', '541511', 0.9),
('development', '541511', 0.8),
('programming', '541511', 0.9),
('technology', '541511', 0.7),
('consulting', '541611', 0.8),
('advisory', '541611', 0.7),
('manufacturing', '336111', 0.9),
('equipment', '336111', 0.8),
('marketing', '541810', 0.9),
('advertising', '541810', 0.8),
('healthcare', '621111', 0.9),
('medical', '621111', 0.8);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_test_businesses_name ON test_businesses(name);
CREATE INDEX IF NOT EXISTS idx_test_businesses_industry ON test_businesses(industry);
CREATE INDEX IF NOT EXISTS idx_test_classifications_business_id ON test_classifications(business_id);
CREATE INDEX IF NOT EXISTS idx_test_classifications_industry_code ON test_classifications(industry_code);
CREATE INDEX IF NOT EXISTS idx_test_keywords_keyword ON test_keywords(keyword);
CREATE INDEX IF NOT EXISTS idx_test_keywords_industry_code ON test_keywords(industry_code);

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO kyb_test;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO kyb_test;
