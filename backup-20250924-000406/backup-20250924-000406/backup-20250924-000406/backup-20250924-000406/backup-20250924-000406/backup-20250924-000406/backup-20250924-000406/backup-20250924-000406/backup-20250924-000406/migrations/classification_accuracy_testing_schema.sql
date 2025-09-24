-- Classification Accuracy Testing Schema
-- This schema supports comprehensive testing of classification accuracy

-- Test samples table for storing known business samples
CREATE TABLE IF NOT EXISTS classification_test_samples (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_name VARCHAR(255) NOT NULL,
    description TEXT,
    website_url VARCHAR(500),
    expected_mcc VARCHAR(10),
    expected_naics VARCHAR(10),
    expected_sic VARCHAR(10),
    expected_industry VARCHAR(255),
    manual_classification JSONB NOT NULL,
    test_category VARCHAR(50) NOT NULL DEFAULT 'primary',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(100),
    notes TEXT
);

-- Accuracy reports table for storing test results
CREATE TABLE IF NOT EXISTS classification_accuracy_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    overall_accuracy DECIMAL(5,4) NOT NULL,
    mcc_accuracy DECIMAL(5,4) NOT NULL,
    naics_accuracy DECIMAL(5,4) NOT NULL,
    sic_accuracy DECIMAL(5,4) NOT NULL,
    industry_accuracy DECIMAL(5,4) NOT NULL,
    confidence_accuracy DECIMAL(5,4) NOT NULL,
    category_metrics JSONB,
    processing_metrics JSONB,
    error_analysis JSONB,
    recommendations JSONB,
    test_duration INTERVAL,
    sample_count INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    test_version VARCHAR(50),
    notes TEXT
);

-- Individual test results table for detailed analysis
CREATE TABLE IF NOT EXISTS classification_test_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    test_sample_id UUID NOT NULL REFERENCES classification_test_samples(id),
    accuracy_report_id UUID NOT NULL REFERENCES classification_accuracy_reports(id),
    business_name VARCHAR(255) NOT NULL,
    expected_mcc VARCHAR(10),
    expected_naics VARCHAR(10),
    expected_sic VARCHAR(10),
    expected_industry VARCHAR(255),
    actual_mcc_results JSONB,
    actual_naics_results JSONB,
    actual_sic_results JSONB,
    actual_industry_results JSONB,
    overall_confidence DECIMAL(3,2),
    processing_time INTERVAL,
    mcc_accuracy DECIMAL(5,4),
    naics_accuracy DECIMAL(5,4),
    sic_accuracy DECIMAL(5,4),
    industry_accuracy DECIMAL(5,4),
    confidence_accuracy DECIMAL(5,4),
    has_errors BOOLEAN DEFAULT false,
    error_types TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Test categories table for organizing test samples
CREATE TABLE IF NOT EXISTS classification_test_categories (
    id SERIAL PRIMARY KEY,
    category_name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    target_accuracy DECIMAL(5,4) DEFAULT 0.95,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default test categories
INSERT INTO classification_test_categories (category_name, description, target_accuracy) VALUES
('primary', 'Primary industry classifications', 0.95),
('edge_case', 'Edge cases and difficult classifications', 0.85),
('high_risk', 'High-risk business classifications', 0.90),
('emerging', 'Emerging industry classifications', 0.80),
('crosswalk', 'MCC/NAICS/SIC crosswalk validation', 0.95),
('confidence', 'Confidence scoring validation', 0.85)
ON CONFLICT (category_name) DO NOTHING;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_test_samples_category ON classification_test_samples(test_category);
CREATE INDEX IF NOT EXISTS idx_test_samples_active ON classification_test_samples(is_active);
CREATE INDEX IF NOT EXISTS idx_test_samples_business_name ON classification_test_samples(business_name);
CREATE INDEX IF NOT EXISTS idx_accuracy_reports_created_at ON classification_accuracy_reports(created_at);
CREATE INDEX IF NOT EXISTS idx_test_results_sample_id ON classification_test_results(test_sample_id);
CREATE INDEX IF NOT EXISTS idx_test_results_report_id ON classification_test_results(accuracy_report_id);
CREATE INDEX IF NOT EXISTS idx_test_results_has_errors ON classification_test_results(has_errors);

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_test_samples_updated_at 
    BEFORE UPDATE ON classification_test_samples 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Views for easy reporting
CREATE OR REPLACE VIEW classification_accuracy_summary AS
SELECT 
    DATE_TRUNC('day', created_at) as test_date,
    COUNT(*) as total_tests,
    AVG(overall_accuracy) as avg_overall_accuracy,
    AVG(mcc_accuracy) as avg_mcc_accuracy,
    AVG(naics_accuracy) as avg_naics_accuracy,
    AVG(sic_accuracy) as avg_sic_accuracy,
    AVG(industry_accuracy) as avg_industry_accuracy,
    AVG(confidence_accuracy) as avg_confidence_accuracy,
    AVG(sample_count) as avg_sample_count
FROM classification_accuracy_reports
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY test_date DESC;

CREATE OR REPLACE VIEW classification_error_analysis AS
SELECT 
    cts.test_category,
    COUNT(*) as total_samples,
    COUNT(CASE WHEN ctr.has_errors THEN 1 END) as error_count,
    ROUND(COUNT(CASE WHEN ctr.has_errors THEN 1 END)::DECIMAL / COUNT(*) * 100, 2) as error_rate,
    AVG(ctr.mcc_accuracy) as avg_mcc_accuracy,
    AVG(ctr.naics_accuracy) as avg_naics_accuracy,
    AVG(ctr.sic_accuracy) as avg_sic_accuracy,
    AVG(ctr.industry_accuracy) as avg_industry_accuracy
FROM classification_test_samples cts
LEFT JOIN classification_test_results ctr ON cts.id = ctr.test_sample_id
WHERE cts.is_active = true
GROUP BY cts.test_category
ORDER BY error_rate DESC;

-- Sample test data for initial testing
INSERT INTO classification_test_samples (
    business_name, description, website_url, expected_mcc, expected_naics, expected_sic, 
    expected_industry, manual_classification, test_category, created_by
) VALUES
-- Technology companies
('Apple Inc.', 'Technology company that designs and manufactures consumer electronics, software, and online services', 'https://apple.com', '5733', '334111', '3571', 'Technology', 
 '{"mcc_code": "5733", "mcc_description": "Computer Software Stores", "naics_code": "334111", "naics_description": "Electronic Computer Manufacturing", "sic_code": "3571", "sic_description": "Electronic Computers", "industry_id": 1, "industry_name": "Technology", "confidence": 0.95, "notes": "Clear technology company", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}', 
 'primary', 'system'),

('Microsoft Corporation', 'Multinational technology corporation that develops, manufactures, licenses, supports and sells computer software, consumer electronics, personal computers, and related services', 'https://microsoft.com', '5733', '541511', '7372', 'Technology',
 '{"mcc_code": "5733", "mcc_description": "Computer Software Stores", "naics_code": "541511", "naics_description": "Custom Computer Programming Services", "sic_code": "7372", "sic_description": "Prepackaged Software", "industry_id": 1, "industry_name": "Technology", "confidence": 0.95, "notes": "Software and technology services", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'primary', 'system'),

-- Financial services
('JPMorgan Chase & Co.', 'Multinational investment bank and financial services holding company', 'https://jpmorganchase.com', '6012', '522110', '6021', 'Financial Services',
 '{"mcc_code": "6012", "mcc_description": "Financial Institutions - Merchandise, Services", "naics_code": "522110", "naics_description": "Commercial Banking", "sic_code": "6021", "sic_description": "National Commercial Banks", "industry_id": 2, "industry_name": "Financial Services", "confidence": 0.98, "notes": "Major commercial bank", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'primary', 'system'),

-- Healthcare
('Johnson & Johnson', 'Multinational corporation founded in 1886 that develops medical devices, pharmaceuticals, and consumer packaged goods', 'https://jnj.com', '5122', '325412', '2834', 'Healthcare',
 '{"mcc_code": "5122", "mcc_description": "Drugs, Drug Proprietaries, and Druggist Sundries", "naics_code": "325412", "naics_description": "Pharmaceutical Preparation Manufacturing", "sic_code": "2834", "sic_description": "Pharmaceutical Preparations", "industry_id": 3, "industry_name": "Healthcare", "confidence": 0.97, "notes": "Pharmaceutical and medical device company", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'primary', 'system'),

-- Retail
('Amazon.com Inc.', 'Multinational technology company which focuses on e-commerce, cloud computing, digital streaming, and artificial intelligence', 'https://amazon.com', '5969', '454110', '5961', 'Retail',
 '{"mcc_code": "5969", "mcc_description": "Miscellaneous and Specialty Retail Stores", "naics_code": "454110", "naics_description": "Electronic Shopping and Mail-Order Houses", "sic_code": "5961", "sic_description": "Catalog and Mail-Order Houses", "industry_id": 4, "industry_name": "Retail", "confidence": 0.96, "notes": "E-commerce and retail platform", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'primary', 'system'),

-- Manufacturing
('General Motors Company', 'Multinational automotive corporation that designs, manufactures, markets, and distributes vehicles and vehicle parts', 'https://gm.com', '5511', '336111', '3711', 'Manufacturing',
 '{"mcc_code": "5511", "mcc_description": "Automotive Dealers (New & Used) Sales, Service, Repairs Parts and Leasing", "naics_code": "336111", "naics_description": "Automobile Manufacturing", "sic_code": "3711", "sic_description": "Motor Vehicles and Passenger Car Bodies", "industry_id": 5, "industry_name": "Manufacturing", "confidence": 0.98, "notes": "Automotive manufacturer", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'primary', 'system'),

-- Edge cases
('Bitcoin Exchange LLC', 'Cryptocurrency exchange platform for buying and selling digital currencies', 'https://bitcoinexchange.com', '5999', '523130', '7389', 'Financial Services',
 '{"mcc_code": "5999", "mcc_description": "Miscellaneous and Specialty Retail Stores", "naics_code": "523130", "naics_description": "Securities and Commodity Exchanges", "sic_code": "7389", "sic_description": "Business Services, Not Elsewhere Classified", "industry_id": 2, "industry_name": "Financial Services", "confidence": 0.85, "notes": "Cryptocurrency exchange - high risk", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'edge_case', 'system'),

-- High-risk examples
('Adult Entertainment Store', 'Adult entertainment and novelty store', 'https://adultstore.com', '7273', '713290', '7999', 'Adult Entertainment',
 '{"mcc_code": "7273", "mcc_description": "Dating Services", "naics_code": "713290", "naics_description": "Other Gambling Industries", "sic_code": "7999", "sic_description": "Amusement and Recreation Services, Not Elsewhere Classified", "industry_id": 6, "industry_name": "Adult Entertainment", "confidence": 0.90, "notes": "Adult entertainment - prohibited by many card brands", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'high_risk', 'system'),

-- Emerging industries
('Green Energy Solutions', 'Renewable energy consulting and solar panel installation services', 'https://greenenergy.com', '1711', '238220', '1629', 'Green Energy',
 '{"mcc_code": "1711", "mcc_description": "Air Conditioning Contractors - Sales and Installation", "naics_code": "238220", "naics_description": "Plumbing, Heating, and Air-Conditioning Contractors", "sic_code": "1629", "sic_description": "Heavy Construction, Not Elsewhere Classified", "industry_id": 7, "industry_name": "Green Energy", "confidence": 0.80, "notes": "Emerging green energy sector", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'emerging', 'system'),

-- Crosswalk validation
('Restaurant Chain', 'Fast food restaurant chain with multiple locations', 'https://restaurantchain.com', '5812', '722513', '5812', 'Food Service',
 '{"mcc_code": "5812", "mcc_description": "Eating Places, Restaurants", "naics_code": "722513", "naics_description": "Limited-Service Restaurants", "sic_code": "5812", "sic_description": "Eating Places", "industry_id": 8, "industry_name": "Food Service", "confidence": 0.95, "notes": "Standard restaurant classification", "classified_by": "expert", "classified_at": "2025-01-19T10:00:00Z"}',
 'crosswalk', 'system');

-- Comments for documentation
COMMENT ON TABLE classification_test_samples IS 'Stores known business samples for testing classification accuracy';
COMMENT ON TABLE classification_accuracy_reports IS 'Stores comprehensive accuracy test results and metrics';
COMMENT ON TABLE classification_test_results IS 'Stores individual test results for detailed analysis';
COMMENT ON TABLE classification_test_categories IS 'Defines test categories and their target accuracy levels';

COMMENT ON COLUMN classification_test_samples.manual_classification IS 'JSONB containing expert manual classification with confidence scores';
COMMENT ON COLUMN classification_accuracy_reports.category_metrics IS 'JSONB containing accuracy metrics by test category';
COMMENT ON COLUMN classification_accuracy_reports.processing_metrics IS 'JSONB containing processing time and performance metrics';
COMMENT ON COLUMN classification_accuracy_reports.error_analysis IS 'JSONB containing detailed error analysis and categorization';
COMMENT ON COLUMN classification_accuracy_reports.recommendations IS 'JSONB array of improvement recommendations';
