-- KYB Platform - Supabase Database Migration
-- Run this script in the Supabase SQL Editor to create all necessary tables and data

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create portfolio types table
CREATE TABLE IF NOT EXISTS portfolio_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) UNIQUE NOT NULL CHECK (type IN ('onboarded', 'deactivated', 'prospective', 'pending')),
    description TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create risk levels table
CREATE TABLE IF NOT EXISTS risk_levels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    level VARCHAR(50) UNIQUE NOT NULL CHECK (level IN ('high', 'medium', 'low')),
    description TEXT,
    numeric_value INTEGER NOT NULL,
    color_code VARCHAR(7), -- Hex color code for UI
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create merchants table
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    registration_number VARCHAR(100) UNIQUE NOT NULL,
    tax_id VARCHAR(100),
    industry VARCHAR(100),
    industry_code VARCHAR(20),
    business_type VARCHAR(50),
    founded_date DATE,
    employee_count INTEGER,
    annual_revenue DECIMAL(15,2),
    
    -- Address fields (flattened for better query performance)
    address_street1 VARCHAR(255),
    address_street2 VARCHAR(255),
    address_city VARCHAR(100),
    address_state VARCHAR(100),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100),
    address_country_code VARCHAR(10),
    
    -- Contact info fields (flattened for better query performance)
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),
    contact_website VARCHAR(255),
    contact_primary_contact VARCHAR(255),
    
    -- Portfolio management fields
    portfolio_type_id UUID,
    risk_level_id UUID,
    compliance_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    
    -- Audit fields
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_portfolio_types_updated_at 
    BEFORE UPDATE ON portfolio_types 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_risk_levels_updated_at 
    BEFORE UPDATE ON risk_levels 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_merchants_updated_at 
    BEFORE UPDATE ON merchants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert portfolio types
INSERT INTO portfolio_types (type, description, display_order, is_active) VALUES
('onboarded', 'Fully onboarded and active merchants', 1, true),
('prospective', 'Potential merchants under evaluation', 2, true),
('pending', 'Merchants awaiting approval or processing', 3, true),
('deactivated', 'Deactivated or suspended merchants', 4, true)
ON CONFLICT (type) DO UPDATE SET
    description = EXCLUDED.description,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Insert risk levels
INSERT INTO risk_levels (level, description, numeric_value, color_code, display_order, is_active) VALUES
('low', 'Low risk merchants with established compliance history', 1, '#10B981', 1, true),
('medium', 'Medium risk merchants requiring standard monitoring', 2, '#F59E0B', 2, true),
('high', 'High risk merchants requiring enhanced due diligence', 3, '#EF4444', 3, true)
ON CONFLICT (level) DO UPDATE SET
    description = EXCLUDED.description,
    numeric_value = EXCLUDED.numeric_value,
    color_code = EXCLUDED.color_code,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Insert sample merchants
INSERT INTO merchants (
    id, name, legal_name, registration_number, tax_id, industry, industry_code, business_type,
    founded_date, employee_count, annual_revenue,
    address_street1, address_street2, address_city, address_state, address_postal_code, address_country, address_country_code,
    contact_phone, contact_email, contact_website, contact_primary_contact,
    portfolio_type_id, risk_level_id, compliance_status, status
) VALUES
-- Technology Companies (Onboarded - Low Risk)
('10000000-0000-0000-0000-000000000001', 'TechFlow Solutions', 'TechFlow Solutions Inc.', 'TF-2023-001', '12-3456789', 'Technology', '541511', 'Corporation',
'2020-03-15', 45, 2500000.00,
'123 Innovation Drive', NULL, 'San Francisco', 'CA', '94105', 'United States', 'US',
'+1-415-555-0101', 'info@techflow.com', 'https://techflow.com', 'Sarah Johnson',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'low'), 'compliant', 'active'),

('10000000-0000-0000-0000-000000000002', 'DataSync Analytics', 'DataSync Analytics LLC', 'DS-2022-002', '98-7654321', 'Technology', '541512', 'LLC',
'2019-08-22', 28, 1800000.00,
'456 Data Street', 'Suite 200', 'Austin', 'TX', '78701', 'United States', 'US',
'+1-512-555-0102', 'contact@datasync.com', 'https://datasync.com', 'Michael Chen',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'low'), 'compliant', 'active'),

('10000000-0000-0000-0000-000000000003', 'CloudScale Systems', 'CloudScale Systems Inc.', 'CS-2021-003', '45-6789012', 'Technology', '541511', 'Corporation',
'2018-11-10', 67, 4200000.00,
'789 Cloud Avenue', NULL, 'Seattle', 'WA', '98101', 'United States', 'US',
'+1-206-555-0103', 'hello@cloudscale.com', 'https://cloudscale.com', 'David Rodriguez',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'low'), 'compliant', 'active'),

-- Financial Services (Onboarded - Medium Risk)
('10000000-0000-0000-0000-000000000004', 'Metro Credit Union', 'Metro Credit Union', 'MCU-2020-004', '34-5678901', 'Finance', '522110', 'Credit Union',
'2015-06-30', 125, 8500000.00,
'321 Financial Plaza', 'Floor 15', 'Chicago', 'IL', '60601', 'United States', 'US',
'+1-312-555-0104', 'info@metrocu.org', 'https://metrocu.org', 'Jennifer Williams',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'compliant', 'active'),

('10000000-0000-0000-0000-000000000005', 'Premier Investment Group', 'Premier Investment Group LLC', 'PIG-2019-005', '56-7890123', 'Finance', '523920', 'LLC',
'2017-04-12', 89, 12000000.00,
'654 Investment Way', 'Suite 500', 'New York', 'NY', '10001', 'United States', 'US',
'+1-212-555-0105', 'contact@premierinvest.com', 'https://premierinvest.com', 'Robert Thompson',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'compliant', 'active'),

-- Healthcare (Prospective - Medium Risk)
('10000000-0000-0000-0000-000000000006', 'Wellness Medical Center', 'Wellness Medical Center PLLC', 'WMC-2023-006', '78-9012345', 'Healthcare', '621111', 'PLLC',
'2021-09-05', 156, 9800000.00,
'987 Health Boulevard', NULL, 'Denver', 'CO', '80201', 'United States', 'US',
'+1-303-555-0106', 'info@wellnessmed.com', 'https://wellnessmed.com', 'Dr. Lisa Anderson',
(SELECT id FROM portfolio_types WHERE type = 'prospective'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'pending', 'active'),

('10000000-0000-0000-0000-000000000007', 'Advanced Dental Care', 'Advanced Dental Care PC', 'ADC-2022-007', '90-1234567', 'Healthcare', '621210', 'PC',
'2020-12-18', 34, 2100000.00,
'147 Dental Drive', 'Building A', 'Phoenix', 'AZ', '85001', 'United States', 'US',
'+1-602-555-0107', 'appointments@advanceddental.com', 'https://advanceddental.com', 'Dr. Mark Davis',
(SELECT id FROM portfolio_types WHERE type = 'prospective'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'pending', 'active'),

-- Retail (Pending - Low Risk)
('10000000-0000-0000-0000-000000000008', 'Urban Fashion Co.', 'Urban Fashion Company Inc.', 'UFC-2023-008', '12-3456780', 'Retail', '448140', 'Corporation',
'2022-01-20', 78, 5600000.00,
'258 Fashion Street', NULL, 'Los Angeles', 'CA', '90001', 'United States', 'US',
'+1-213-555-0108', 'orders@urbanfashion.com', 'https://urbanfashion.com', 'Amanda Garcia',
(SELECT id FROM portfolio_types WHERE type = 'pending'), (SELECT id FROM risk_levels WHERE level = 'low'), 'pending', 'active'),

('10000000-0000-0000-0000-000000000009', 'Green Earth Organics', 'Green Earth Organics LLC', 'GEO-2023-009', '23-4567890', 'Retail', '445299', 'LLC',
'2021-07-14', 42, 3200000.00,
'369 Organic Lane', NULL, 'Portland', 'OR', '97201', 'United States', 'US',
'+1-503-555-0109', 'info@greenearth.com', 'https://greenearth.com', 'Elena Martinez',
(SELECT id FROM portfolio_types WHERE type = 'pending'), (SELECT id FROM risk_levels WHERE level = 'low'), 'pending', 'active'),

-- Manufacturing (High Risk - Deactivated)
('10000000-0000-0000-0000-000000000010', 'Precision Manufacturing', 'Precision Manufacturing Corp', 'PMC-2018-010', '34-5678900', 'Manufacturing', '332710', 'Corporation',
'2016-03-08', 234, 18500000.00,
'741 Industrial Park', 'Building 3', 'Detroit', 'MI', '48201', 'United States', 'US',
'+1-313-555-0110', 'info@precisionmfg.com', 'https://precisionmfg.com', 'James Wilson',
(SELECT id FROM portfolio_types WHERE type = 'deactivated'), (SELECT id FROM risk_levels WHERE level = 'high'), 'non_compliant', 'inactive')

ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    legal_name = EXCLUDED.legal_name,
    registration_number = EXCLUDED.registration_number,
    tax_id = EXCLUDED.tax_id,
    industry = EXCLUDED.industry,
    industry_code = EXCLUDED.industry_code,
    business_type = EXCLUDED.business_type,
    founded_date = EXCLUDED.founded_date,
    employee_count = EXCLUDED.employee_count,
    annual_revenue = EXCLUDED.annual_revenue,
    address_street1 = EXCLUDED.address_street1,
    address_street2 = EXCLUDED.address_street2,
    address_city = EXCLUDED.address_city,
    address_state = EXCLUDED.address_state,
    address_postal_code = EXCLUDED.address_postal_code,
    address_country = EXCLUDED.address_country,
    address_country_code = EXCLUDED.address_country_code,
    contact_phone = EXCLUDED.contact_phone,
    contact_email = EXCLUDED.contact_email,
    contact_website = EXCLUDED.contact_website,
    contact_primary_contact = EXCLUDED.contact_primary_contact,
    portfolio_type_id = EXCLUDED.portfolio_type_id,
    risk_level_id = EXCLUDED.risk_level_id,
    compliance_status = EXCLUDED.compliance_status,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_merchants_registration_number ON merchants(registration_number);
CREATE INDEX IF NOT EXISTS idx_merchants_industry ON merchants(industry);
CREATE INDEX IF NOT EXISTS idx_merchants_status ON merchants(status);
CREATE INDEX IF NOT EXISTS idx_merchants_portfolio_type_id ON merchants(portfolio_type_id);
CREATE INDEX IF NOT EXISTS idx_merchants_risk_level_id ON merchants(risk_level_id);
CREATE INDEX IF NOT EXISTS idx_merchants_compliance_status ON merchants(compliance_status);
CREATE INDEX IF NOT EXISTS idx_merchants_created_at ON merchants(created_at);

-- Create search indexes for merchant search functionality
CREATE INDEX IF NOT EXISTS idx_merchants_name_trgm ON merchants USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_merchants_legal_name_trgm ON merchants USING gin(legal_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_merchants_contact_email ON merchants(contact_email);
CREATE INDEX IF NOT EXISTS idx_merchants_contact_phone ON merchants(contact_phone);

-- Verify data insertion
DO $$
DECLARE
    portfolio_count INTEGER;
    risk_count INTEGER;
    merchant_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO portfolio_count FROM portfolio_types;
    SELECT COUNT(*) INTO risk_count FROM risk_levels;
    SELECT COUNT(*) INTO merchant_count FROM merchants;
    
    RAISE NOTICE 'Migration completed successfully:';
    RAISE NOTICE '  Portfolio Types: %', portfolio_count;
    RAISE NOTICE '  Risk Levels: %', risk_count;
    RAISE NOTICE '  Merchants: %', merchant_count;
END $$;
