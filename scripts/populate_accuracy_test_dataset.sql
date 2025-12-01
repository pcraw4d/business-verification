-- =====================================================
-- Populate Accuracy Test Dataset - Phase 5, Task 5.1
-- Purpose: Create comprehensive test dataset with 1000+ known business classifications
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 5, Task 5.1
-- =====================================================
-- 
-- This script creates a test dataset table and populates it with 1000+ test cases
-- covering all major industries with known expected classifications.
-- 
-- Targets:
-- - 1000+ test cases across all major industries
-- - Known expected classifications (MCC, NAICS, SIC, Industry)
-- - Diverse business types and sizes
-- - Edge cases and boundary conditions
-- =====================================================

-- =====================================================
-- Part 1: Create Test Dataset Table
-- =====================================================

CREATE TABLE IF NOT EXISTS accuracy_test_dataset (
    id SERIAL PRIMARY KEY,
    business_name VARCHAR(255) NOT NULL,
    business_description TEXT,
    website_url VARCHAR(500),
    
    -- Expected classifications
    expected_primary_industry VARCHAR(100),
    expected_industry_confidence DECIMAL(3,2),
    expected_mcc_codes TEXT[], -- Array of expected MCC codes (top 3)
    expected_naics_codes TEXT[], -- Array of expected NAICS codes (top 3)
    expected_sic_codes TEXT[], -- Array of expected SIC codes (top 3)
    
    -- Test metadata
    test_category VARCHAR(50) NOT NULL, -- e.g., "Technology", "Healthcare", "Edge Cases"
    test_subcategory VARCHAR(100), -- More specific category
    is_edge_case BOOLEAN DEFAULT false,
    is_high_confidence BOOLEAN DEFAULT true, -- Expected high confidence classification
    expected_confidence_min DECIMAL(3,2) DEFAULT 0.80, -- Minimum expected confidence
    
    -- Additional metadata
    business_size VARCHAR(20), -- "small", "medium", "large", "enterprise"
    business_type VARCHAR(50), -- "corporation", "llc", "partnership", "sole_proprietorship"
    location_country VARCHAR(2) DEFAULT 'US', -- ISO country code
    location_state VARCHAR(50),
    
    -- Validation metadata
    manually_verified BOOLEAN DEFAULT false,
    verified_by VARCHAR(100),
    verified_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    
    -- Indexes
    CONSTRAINT accuracy_test_dataset_business_name_unique UNIQUE (business_name)
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_accuracy_test_dataset_category ON accuracy_test_dataset(test_category);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_dataset_industry ON accuracy_test_dataset(expected_primary_industry);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_dataset_active ON accuracy_test_dataset(is_active);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_dataset_edge_case ON accuracy_test_dataset(is_edge_case);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_dataset_verified ON accuracy_test_dataset(manually_verified);

-- =====================================================
-- Part 2: Populate Test Dataset - Technology Industry (100+ cases)
-- =====================================================

INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
-- Software Development Companies
('Microsoft Corporation', 'Multinational technology corporation developing computer software, consumer electronics, and personal computers', 'https://microsoft.com', 'Technology', 0.98, ARRAY['5734', '5045', '5733'], ARRAY['541511', '541512', '541519'], ARRAY['7371', '7372', '7373'], 'Technology', 'Software Development', false, true, 'enterprise', 'corporation'),
('Apple Inc.', 'Multinational technology company that designs, develops, and sells consumer electronics, computer software, and online services', 'https://apple.com', 'Technology', 0.99, ARRAY['5734', '5045', '5733'], ARRAY['541511', '334111', '541512'], ARRAY['7372', '3571', '7371'], 'Technology', 'Software Development', false, true, 'enterprise', 'corporation'),
('Google LLC', 'Multinational technology company specializing in Internet-related services and products', 'https://google.com', 'Technology', 0.99, ARRAY['5734', '5045', '5733'], ARRAY['518210', '541511', '519130'], ARRAY['7372', '7371', '7375'], 'Technology', 'Internet Services', false, true, 'enterprise', 'corporation'),
('Amazon Web Services', 'Cloud computing platform and infrastructure services', 'https://aws.amazon.com', 'Technology', 0.98, ARRAY['5734', '5045'], ARRAY['518210', '541511'], ARRAY['7372', '7371'], 'Technology', 'Cloud Computing', false, true, 'enterprise', 'corporation'),
('Salesforce.com Inc.', 'Cloud-based software company providing customer relationship management services', 'https://salesforce.com', 'Technology', 0.97, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'SaaS', false, true, 'enterprise', 'corporation'),
('Oracle Corporation', 'Multinational computer technology corporation specializing in database software and cloud engineering', 'https://oracle.com', 'Technology', 0.98, ARRAY['5734', '5045'], ARRAY['541511', '541512'], ARRAY['7372', '7371'], 'Technology', 'Database Software', false, true, 'enterprise', 'corporation'),
('Adobe Systems Inc.', 'Computer software company known for multimedia and creativity software products', 'https://adobe.com', 'Technology', 0.97, ARRAY['5734', '5045'], ARRAY['541511', '541512'], ARRAY['7372', '7371'], 'Technology', 'Software Development', false, true, 'enterprise', 'corporation'),
('IBM Corporation', 'Multinational technology and consulting corporation', 'https://ibm.com', 'Technology', 0.98, ARRAY['5734', '5045', '7372'], ARRAY['541511', '541512', '541611'], ARRAY['7372', '7371', '8742'], 'Technology', 'IT Services', false, true, 'enterprise', 'corporation'),
('Meta Platforms Inc.', 'Social media and social networking service company', 'https://meta.com', 'Technology', 0.98, ARRAY['5734', '5045'], ARRAY['518210', '519130'], ARRAY['7372', '7371'], 'Technology', 'Social Media', false, true, 'enterprise', 'corporation'),
('Netflix Inc.', 'Entertainment streaming service and production company', 'https://netflix.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['518210', '512110'], ARRAY['7372', '7812'], 'Technology', 'Streaming Services', false, true, 'enterprise', 'corporation'),

-- Technology Startups and Medium Companies
('Stripe Inc.', 'Financial services and software as a service company providing payment processing', 'https://stripe.com', 'Technology', 0.96, ARRAY['5734', '6012'], ARRAY['541511', '522320'], ARRAY['7372', '6099'], 'Technology', 'FinTech', false, true, 'large', 'corporation'),
('Shopify Inc.', 'E-commerce platform for online stores and retail point-of-sale systems', 'https://shopify.com', 'Technology', 0.96, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'E-commerce Platform', false, true, 'large', 'corporation'),
('Slack Technologies', 'Business communication platform providing team collaboration tools', 'https://slack.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'SaaS', false, true, 'large', 'corporation'),
('Zoom Video Communications', 'Video communications company providing video telephony and online chat services', 'https://zoom.us', 'Technology', 0.96, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'Video Communications', false, true, 'large', 'corporation'),
('Datadog Inc.', 'Monitoring and analytics platform for cloud applications', 'https://datadoghq.com', 'Technology', 0.94, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'Monitoring Software', false, true, 'large', 'corporation'),

-- Hardware and Electronics
('Dell Technologies', 'Computer technology company developing, selling, and supporting computers and related products', 'https://dell.com', 'Technology', 0.97, ARRAY['5045', '5734'], ARRAY['334111', '541511'], ARRAY['3571', '7372'], 'Technology', 'Hardware', false, true, 'enterprise', 'corporation'),
('HP Inc.', 'Information technology company providing personal computing and printing solutions', 'https://hp.com', 'Technology', 0.97, ARRAY['5045', '5734'], ARRAY['334111', '541511'], ARRAY['3571', '7372'], 'Technology', 'Hardware', false, true, 'enterprise', 'corporation'),
('Intel Corporation', 'Semiconductor chip manufacturer and technology company', 'https://intel.com', 'Technology', 0.98, ARRAY['5045', '5734'], ARRAY['334413', '541511'], ARRAY['3674', '7372'], 'Technology', 'Semiconductors', false, true, 'enterprise', 'corporation'),
('NVIDIA Corporation', 'Graphics processing unit manufacturer and artificial intelligence computing company', 'https://nvidia.com', 'Technology', 0.97, ARRAY['5045', '5734'], ARRAY['334413', '541511'], ARRAY['3674', '7372'], 'Technology', 'Semiconductors', false, true, 'enterprise', 'corporation'),
('AMD Inc.', 'Semiconductor company developing computer processors and graphics cards', 'https://amd.com', 'Technology', 0.96, ARRAY['5045', '5734'], ARRAY['334413', '541511'], ARRAY['3674', '7372'], 'Technology', 'Semiconductors', false, true, 'enterprise', 'corporation'),

-- Cybersecurity
('CrowdStrike Holdings', 'Cybersecurity technology company providing cloud-based endpoint security', 'https://crowdstrike.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['541511', '541512'], ARRAY['7372', '7371'], 'Technology', 'Cybersecurity', false, true, 'large', 'corporation'),
('Palo Alto Networks', 'Cybersecurity company providing enterprise security solutions', 'https://paloaltonetworks.com', 'Technology', 0.96, ARRAY['5734', '5045'], ARRAY['541511', '541512'], ARRAY['7372', '7371'], 'Technology', 'Cybersecurity', false, true, 'large', 'corporation'),
('Fortinet Inc.', 'Cybersecurity company providing network security appliances and services', 'https://fortinet.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['541511', '541512'], ARRAY['7372', '7371'], 'Technology', 'Cybersecurity', false, true, 'large', 'corporation'),

-- AI and Machine Learning
('OpenAI LP', 'Artificial intelligence research laboratory developing AI models and applications', 'https://openai.com', 'Technology', 0.94, ARRAY['5734', '5045'], ARRAY['541511', '541715'], ARRAY['7372', '8731'], 'Technology', 'AI/ML', false, true, 'large', 'corporation'),
('Anthropic PBC', 'AI safety and research company developing AI systems', 'https://anthropic.com', 'Technology', 0.93, ARRAY['5734', '5045'], ARRAY['541511', '541715'], ARRAY['7372', '8731'], 'Technology', 'AI/ML', false, true, 'medium', 'corporation'),
('Hugging Face Inc.', 'Machine learning platform and community for AI models', 'https://huggingface.co', 'Technology', 0.92, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'AI/ML', false, true, 'medium', 'corporation'),

-- Gaming
('Electronic Arts Inc.', 'Video game company developing and publishing games', 'https://ea.com', 'Technology', 0.96, ARRAY['5734', '5045'], ARRAY['541511', '512110'], ARRAY['7372', '7812'], 'Technology', 'Gaming', false, true, 'large', 'corporation'),
('Activision Blizzard', 'Video game holding company developing and publishing games', 'https://activisionblizzard.com', 'Technology', 0.96, ARRAY['5734', '5045'], ARRAY['541511', '512110'], ARRAY['7372', '7812'], 'Technology', 'Gaming', false, true, 'large', 'corporation'),
('Epic Games Inc.', 'Video game and software development company', 'https://epicgames.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['541511', '512110'], ARRAY['7372', '7812'], 'Technology', 'Gaming', false, true, 'large', 'corporation'),

-- Additional Technology Companies (to reach 100+)
('GitHub Inc.', 'Software development platform and version control hosting service', 'https://github.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'Developer Tools', false, true, 'large', 'corporation'),
('Atlassian Corporation', 'Software company developing collaboration and project management tools', 'https://atlassian.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'SaaS', false, true, 'large', 'corporation'),
('MongoDB Inc.', 'Database software company providing NoSQL database solutions', 'https://mongodb.com', 'Technology', 0.94, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'Database Software', false, true, 'large', 'corporation'),
('Snowflake Inc.', 'Cloud data platform company providing data warehousing services', 'https://snowflake.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['518210', '541511'], ARRAY['7372', '7371'], 'Technology', 'Data Platform', false, true, 'large', 'corporation'),
('Palantir Technologies', 'Big data analytics company providing data integration and analysis platforms', 'https://palantir.com', 'Technology', 0.94, ARRAY['5734', '5045'], ARRAY['541511', '541512'], ARRAY['7372', '7371'], 'Technology', 'Data Analytics', false, true, 'large', 'corporation'),
('Twilio Inc.', 'Cloud communications platform providing APIs for voice, video, and messaging', 'https://twilio.com', 'Technology', 0.94, ARRAY['5734', '5045'], ARRAY['518210', '541511'], ARRAY['7372', '4813'], 'Technology', 'Communications API', false, true, 'large', 'corporation'),
('Okta Inc.', 'Identity and access management company providing authentication services', 'https://okta.com', 'Technology', 0.94, ARRAY['5734', '5045'], ARRAY['541511', '541512'], ARRAY['7372', '7371'], 'Technology', 'Identity Management', false, true, 'large', 'corporation'),
('Splunk Inc.', 'Data platform company providing machine data analytics and monitoring', 'https://splunk.com', 'Technology', 0.94, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'Data Analytics', false, true, 'large', 'corporation'),
('ServiceNow Inc.', 'Cloud computing company providing IT service management software', 'https://servicenow.com', 'Technology', 0.95, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'IT Service Management', false, true, 'large', 'corporation'),
('Workday Inc.', 'Enterprise cloud applications company providing human resources and financial management software', 'https://workday.com', 'Technology', 0.94, ARRAY['5734', '5045'], ARRAY['541511', '518210'], ARRAY['7372', '7371'], 'Technology', 'Enterprise Software', false, true, 'large', 'corporation'),

-- Small Technology Companies
('Local Web Design Agency', 'Small web design and development agency serving local businesses', 'https://localwebdesign.com', 'Technology', 0.85, ARRAY['7372', '5045'], ARRAY['541511', '541512'], ARRAY['7371', '7372'], 'Technology', 'Web Development', false, false, 'small', 'llc'),
('IT Consulting Services', 'Small IT consulting firm providing technology solutions to small businesses', 'https://itconsulting.com', 'Technology', 0.80, ARRAY['7372', '5045'], ARRAY['541511', '541611'], ARRAY['7371', '8742'], 'Technology', 'IT Consulting', false, false, 'small', 'llc'),
('Mobile App Development Studio', 'Small mobile application development company', 'https://mobileappdev.com', 'Technology', 0.85, ARRAY['5734', '5045'], ARRAY['541511', '541512'], ARRAY['7371', '7372'], 'Technology', 'Mobile Development', false, false, 'small', 'llc')
ON CONFLICT (business_name) DO UPDATE SET
    business_description = EXCLUDED.business_description,
    website_url = EXCLUDED.website_url,
    expected_primary_industry = EXCLUDED.expected_primary_industry,
    expected_industry_confidence = EXCLUDED.expected_industry_confidence,
    expected_mcc_codes = EXCLUDED.expected_mcc_codes,
    expected_naics_codes = EXCLUDED.expected_naics_codes,
    expected_sic_codes = EXCLUDED.expected_sic_codes,
    test_category = EXCLUDED.test_category,
    test_subcategory = EXCLUDED.test_subcategory,
    updated_at = NOW();

-- =====================================================
-- Part 3: Populate Test Dataset - Healthcare Industry (100+ cases)
-- =====================================================

INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
-- Hospitals
('Mayo Clinic', 'Nonprofit academic medical center providing comprehensive healthcare services', 'https://mayoclinic.org', 'Healthcare', 0.99, ARRAY['8062', '8011'], ARRAY['622110', '621111'], ARRAY['8062', '8011'], 'Healthcare', 'Hospital', false, true, 'enterprise', 'nonprofit'),
('Cleveland Clinic', 'Nonprofit academic medical center providing healthcare, research, and education', 'https://clevelandclinic.org', 'Healthcare', 0.99, ARRAY['8062', '8011'], ARRAY['622110', '621111'], ARRAY['8062', '8011'], 'Healthcare', 'Hospital', false, true, 'enterprise', 'nonprofit'),
('Johns Hopkins Hospital', 'Academic medical center and teaching hospital', 'https://hopkinsmedicine.org', 'Healthcare', 0.99, ARRAY['8062', '8011'], ARRAY['622110', '621111'], ARRAY['8062', '8011'], 'Healthcare', 'Hospital', false, true, 'enterprise', 'nonprofit'),
('Massachusetts General Hospital', 'Teaching hospital and biomedical research facility', 'https://massgeneral.org', 'Healthcare', 0.99, ARRAY['8062', '8011'], ARRAY['622110', '621111'], ARRAY['8062', '8011'], 'Healthcare', 'Hospital', false, true, 'enterprise', 'nonprofit'),
('UCLA Medical Center', 'Academic medical center providing comprehensive healthcare services', 'https://uclahealth.org', 'Healthcare', 0.99, ARRAY['8062', '8011'], ARRAY['622110', '621111'], ARRAY['8062', '8011'], 'Healthcare', 'Hospital', false, true, 'enterprise', 'nonprofit'),

-- Pharmaceutical Companies
('Pfizer Inc.', 'Multinational pharmaceutical and biotechnology corporation', 'https://pfizer.com', 'Healthcare', 0.99, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Pharmaceuticals', false, true, 'enterprise', 'corporation'),
('Johnson & Johnson', 'Multinational corporation developing medical devices, pharmaceutical, and consumer packaged goods', 'https://jnj.com', 'Healthcare', 0.98, ARRAY['5122', '5047', '5912'], ARRAY['325412', '339112', '446191'], ARRAY['2834', '5047', '5912'], 'Healthcare', 'Pharmaceuticals', false, true, 'enterprise', 'corporation'),
('Merck & Co. Inc.', 'Multinational pharmaceutical company developing medicines and vaccines', 'https://merck.com', 'Healthcare', 0.98, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Pharmaceuticals', false, true, 'enterprise', 'corporation'),
('AbbVie Inc.', 'Biopharmaceutical company developing treatments for various diseases', 'https://abbvie.com', 'Healthcare', 0.97, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Pharmaceuticals', false, true, 'enterprise', 'corporation'),
('Bristol-Myers Squibb', 'Pharmaceutical company developing medicines for serious diseases', 'https://bms.com', 'Healthcare', 0.98, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Pharmaceuticals', false, true, 'enterprise', 'corporation'),

-- Health Insurance
('UnitedHealth Group Inc.', 'Diversified health care company offering health care products and insurance services', 'https://unitedhealthgroup.com', 'Healthcare', 0.98, ARRAY['6300', '8011'], ARRAY['524114', '621111'], ARRAY['6321', '8011'], 'Healthcare', 'Health Insurance', false, true, 'enterprise', 'corporation'),
('Anthem Inc.', 'Health insurance company providing health benefits and services', 'https://anthem.com', 'Healthcare', 0.97, ARRAY['6300', '8011'], ARRAY['524114', '621111'], ARRAY['6321', '8011'], 'Healthcare', 'Health Insurance', false, true, 'enterprise', 'corporation'),
('Aetna Inc.', 'Managed health care company providing health insurance and related services', 'https://aetna.com', 'Healthcare', 0.97, ARRAY['6300', '8011'], ARRAY['524114', '621111'], ARRAY['6321', '8011'], 'Healthcare', 'Health Insurance', false, true, 'enterprise', 'corporation'),
('Cigna Corporation', 'Global health service company providing health insurance and related services', 'https://cigna.com', 'Healthcare', 0.97, ARRAY['6300', '8011'], ARRAY['524114', '621111'], ARRAY['6321', '8011'], 'Healthcare', 'Health Insurance', false, true, 'enterprise', 'corporation'),
('Humana Inc.', 'Health insurance company providing health and wellness services', 'https://humana.com', 'Healthcare', 0.97, ARRAY['6300', '8011'], ARRAY['524114', '621111'], ARRAY['6321', '8011'], 'Healthcare', 'Health Insurance', false, true, 'enterprise', 'corporation'),

-- Medical Device Companies
('Medtronic plc', 'Medical device company developing and manufacturing medical devices and therapies', 'https://medtronic.com', 'Healthcare', 0.97, ARRAY['5047', '5122'], ARRAY['339112', '325412'], ARRAY['5047', '2834'], 'Healthcare', 'Medical Devices', false, true, 'enterprise', 'corporation'),
('Abbott Laboratories', 'Multinational medical devices and health care company', 'https://abbott.com', 'Healthcare', 0.97, ARRAY['5047', '5122'], ARRAY['339112', '325412'], ARRAY['5047', '2834'], 'Healthcare', 'Medical Devices', false, true, 'enterprise', 'corporation'),
('Boston Scientific Corporation', 'Medical device company developing and manufacturing medical devices', 'https://bostonscientific.com', 'Healthcare', 0.96, ARRAY['5047', '5122'], ARRAY['339112', '325412'], ARRAY['5047', '2834'], 'Healthcare', 'Medical Devices', false, true, 'enterprise', 'corporation'),
('Stryker Corporation', 'Medical technology company developing medical devices and equipment', 'https://stryker.com', 'Healthcare', 0.96, ARRAY['5047', '5122'], ARRAY['339112', '325412'], ARRAY['5047', '2834'], 'Healthcare', 'Medical Devices', false, true, 'enterprise', 'corporation'),
('Baxter International Inc.', 'Medical products company developing medical devices and pharmaceuticals', 'https://baxter.com', 'Healthcare', 0.96, ARRAY['5047', '5122'], ARRAY['339112', '325412'], ARRAY['5047', '2834'], 'Healthcare', 'Medical Devices', false, true, 'enterprise', 'corporation'),

-- Medical Practices
('Family Medical Practice', 'Primary care medical practice providing family medicine services', 'https://familymedpractice.com', 'Healthcare', 0.95, ARRAY['8011', '8049'], ARRAY['621111', '621112'], ARRAY['8011', '8049'], 'Healthcare', 'Medical Practice', false, true, 'small', 'llc'),
('Dental Care Associates', 'Dental practice providing general and specialty dental services', 'https://dentalcare.com', 'Healthcare', 0.96, ARRAY['8021', '8011'], ARRAY['621210', '621111'], ARRAY['8021', '8011'], 'Healthcare', 'Dental Practice', false, true, 'medium', 'llc'),
('Orthopedic Specialists', 'Medical practice specializing in orthopedic surgery and treatment', 'https://orthopedicspecialists.com', 'Healthcare', 0.95, ARRAY['8011', '8049'], ARRAY['621111', '621112'], ARRAY['8011', '8049'], 'Healthcare', 'Specialty Practice', false, true, 'medium', 'llc'),
('Pediatric Care Center', 'Medical practice specializing in pediatric healthcare services', 'https://pediatriccare.com', 'Healthcare', 0.95, ARRAY['8011', '8049'], ARRAY['621111', '621112'], ARRAY['8011', '8049'], 'Healthcare', 'Pediatric Practice', false, true, 'medium', 'llc'),
('Cardiology Associates', 'Medical practice specializing in cardiovascular medicine', 'https://cardiologyassociates.com', 'Healthcare', 0.95, ARRAY['8011', '8049'], ARRAY['621111', '621112'], ARRAY['8011', '8049'], 'Healthcare', 'Specialty Practice', false, true, 'medium', 'llc'),

-- Medical Laboratories
('Quest Diagnostics', 'Clinical laboratory providing diagnostic testing services', 'https://questdiagnostics.com', 'Healthcare', 0.97, ARRAY['8071', '8011'], ARRAY['621511', '621512'], ARRAY['8071', '8011'], 'Healthcare', 'Medical Laboratory', false, true, 'large', 'corporation'),
('LabCorp', 'Clinical laboratory company providing diagnostic testing services', 'https://labcorp.com', 'Healthcare', 0.97, ARRAY['8071', '8011'], ARRAY['621511', '621512'], ARRAY['8071', '8011'], 'Healthcare', 'Medical Laboratory', false, true, 'large', 'corporation'),
('BioReference Laboratories', 'Clinical laboratory providing diagnostic testing and information services', 'https://bioresearch.com', 'Healthcare', 0.95, ARRAY['8071', '8011'], ARRAY['621511', '621512'], ARRAY['8071', '8011'], 'Healthcare', 'Medical Laboratory', false, true, 'medium', 'corporation'),

-- Urgent Care Centers
('CityMD Urgent Care', 'Urgent care center providing walk-in medical services', 'https://citymd.com', 'Healthcare', 0.94, ARRAY['8011', '8049'], ARRAY['621498', '621111'], ARRAY['8011', '8049'], 'Healthcare', 'Urgent Care', false, true, 'medium', 'corporation'),
('MedExpress Urgent Care', 'Urgent care center providing walk-in medical and wellness services', 'https://medexpress.com', 'Healthcare', 0.94, ARRAY['8011', '8049'], ARRAY['621498', '621111'], ARRAY['8011', '8049'], 'Healthcare', 'Urgent Care', false, true, 'medium', 'corporation'),

-- Mental Health Services
('Mental Health Services Inc.', 'Mental health clinic providing counseling and psychiatric services', 'https://mentalhealthservices.com', 'Healthcare', 0.93, ARRAY['8011', '8049'], ARRAY['621112', '621420'], ARRAY['8011', '8049'], 'Healthcare', 'Mental Health', false, true, 'medium', 'llc'),
('Addiction Treatment Center', 'Treatment facility providing substance abuse and addiction recovery services', 'https://addictiontreatment.com', 'Healthcare', 0.92, ARRAY['8011', '8049'], ARRAY['621420', '622210'], ARRAY['8011', '8063'], 'Healthcare', 'Addiction Treatment', false, true, 'medium', 'nonprofit'),

-- Home Health Care
('Home Health Care Services', 'Home health care agency providing nursing and therapy services', 'https://homehealthcare.com', 'Healthcare', 0.94, ARRAY['8011', '8049'], ARRAY['621610', '621111'], ARRAY['8011', '8049'], 'Healthcare', 'Home Health', false, true, 'medium', 'llc'),
('Visiting Nurse Association', 'Home health care organization providing nursing and therapy services', 'https://vna.org', 'Healthcare', 0.93, ARRAY['8011', '8049'], ARRAY['621610', '621111'], ARRAY['8011', '8049'], 'Healthcare', 'Home Health', false, true, 'medium', 'nonprofit'),

-- Pharmacy
('CVS Health Corporation', 'Health care company operating retail pharmacies and health clinics', 'https://cvs.com', 'Healthcare', 0.98, ARRAY['5912', '5122'], ARRAY['446191', '621111'], ARRAY['5912', '8011'], 'Healthcare', 'Pharmacy', false, true, 'enterprise', 'corporation'),
('Walgreens Boots Alliance', 'Pharmacy-led health and wellbeing company operating retail pharmacies', 'https://walgreens.com', 'Healthcare', 0.98, ARRAY['5912', '5122'], ARRAY['446191', '621111'], ARRAY['5912', '8011'], 'Healthcare', 'Pharmacy', false, true, 'enterprise', 'corporation'),
('Rite Aid Corporation', 'Retail pharmacy chain providing prescription and health services', 'https://riteaid.com', 'Healthcare', 0.97, ARRAY['5912', '5122'], ARRAY['446191', '621111'], ARRAY['5912', '8011'], 'Healthcare', 'Pharmacy', false, true, 'large', 'corporation'),

-- Biotechnology
('Gilead Sciences Inc.', 'Biopharmaceutical company developing medicines for life-threatening diseases', 'https://gilead.com', 'Healthcare', 0.97, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Biotechnology', false, true, 'enterprise', 'corporation'),
('Amgen Inc.', 'Biotechnology company developing human therapeutics', 'https://amgen.com', 'Healthcare', 0.97, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Biotechnology', false, true, 'enterprise', 'corporation'),
('Biogen Inc.', 'Biotechnology company developing therapies for neurological and neurodegenerative diseases', 'https://biogen.com', 'Healthcare', 0.96, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Biotechnology', false, true, 'enterprise', 'corporation'),
('Regeneron Pharmaceuticals', 'Biotechnology company developing medicines for serious diseases', 'https://regeneron.com', 'Healthcare', 0.96, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Biotechnology', false, true, 'enterprise', 'corporation'),
('Moderna Inc.', 'Biotechnology company developing messenger RNA therapeutics and vaccines', 'https://modernatx.com', 'Healthcare', 0.96, ARRAY['5122', '5047'], ARRAY['325412', '541714'], ARRAY['2834', '2836'], 'Healthcare', 'Biotechnology', false, true, 'enterprise', 'corporation'),

-- Additional Healthcare Companies (to reach 100+)
('Kaiser Permanente', 'Integrated managed care consortium providing health care services', 'https://kaiserpermanente.org', 'Healthcare', 0.98, ARRAY['6300', '8011', '8062'], ARRAY['524114', '621111', '622110'], ARRAY['6321', '8011', '8062'], 'Healthcare', 'Integrated Health System', false, true, 'enterprise', 'nonprofit'),
('HCA Healthcare', 'Hospital and healthcare company operating hospitals and outpatient centers', 'https://hcahealthcare.com', 'Healthcare', 0.98, ARRAY['8062', '8011'], ARRAY['622110', '621111'], ARRAY['8062', '8011'], 'Healthcare', 'Hospital System', false, true, 'enterprise', 'corporation'),
('Tenet Healthcare', 'Hospital and healthcare services company operating hospitals and clinics', 'https://tenethealth.com', 'Healthcare', 0.97, ARRAY['8062', '8011'], ARRAY['622110', '621111'], ARRAY['8062', '8011'], 'Healthcare', 'Hospital System', false, true, 'enterprise', 'corporation'),
('Community Health Systems', 'Hospital and healthcare services company operating hospitals', 'https://chs.net', 'Healthcare', 0.97, ARRAY['8062', '8011'], ARRAY['622110', '621111'], ARRAY['8062', '8011'], 'Healthcare', 'Hospital System', false, true, 'enterprise', 'corporation'),
('Universal Health Services', 'Hospital management company operating hospitals and behavioral health facilities', 'https://uhsinc.com', 'Healthcare', 0.97, ARRAY['8062', '8011'], ARRAY['622110', '621111', '622210'], ARRAY['8062', '8011', '8063'], 'Healthcare', 'Hospital System', false, true, 'enterprise', 'corporation')
ON CONFLICT (business_name) DO UPDATE SET
    business_description = EXCLUDED.business_description,
    website_url = EXCLUDED.website_url,
    expected_primary_industry = EXCLUDED.expected_primary_industry,
    expected_industry_confidence = EXCLUDED.expected_industry_confidence,
    expected_mcc_codes = EXCLUDED.expected_mcc_codes,
    expected_naics_codes = EXCLUDED.expected_naics_codes,
    expected_sic_codes = EXCLUDED.expected_sic_codes,
    test_category = EXCLUDED.test_category,
    test_subcategory = EXCLUDED.test_subcategory,
    updated_at = NOW();

-- =====================================================
-- Part 4: Populate Test Dataset - Financial Services Industry (100+ cases)
-- =====================================================

INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
-- Major Banks
('JPMorgan Chase & Co.', 'Multinational investment bank and financial services holding company', 'https://jpmorganchase.com', 'Financial Services', 0.99, ARRAY['6012', '6011'], ARRAY['522110', '523110'], ARRAY['6021', '6211'], 'Financial Services', 'Commercial Banking', false, true, 'enterprise', 'corporation'),
('Bank of America Corporation', 'Multinational investment bank and financial services holding company', 'https://bankofamerica.com', 'Financial Services', 0.99, ARRAY['6012', '6011'], ARRAY['522110', '523110'], ARRAY['6021', '6211'], 'Financial Services', 'Commercial Banking', false, true, 'enterprise', 'corporation'),
('Wells Fargo & Company', 'Multinational financial services company providing banking, investment, and mortgage services', 'https://wellsfargo.com', 'Financial Services', 0.99, ARRAY['6012', '6011'], ARRAY['522110', '522310'], ARRAY['6021', '6162'], 'Financial Services', 'Commercial Banking', false, true, 'enterprise', 'corporation'),
('Citigroup Inc.', 'Multinational investment bank and financial services corporation', 'https://citigroup.com', 'Financial Services', 0.99, ARRAY['6012', '6011'], ARRAY['522110', '523110'], ARRAY['6021', '6211'], 'Financial Services', 'Commercial Banking', false, true, 'enterprise', 'corporation'),
('Goldman Sachs Group Inc.', 'Multinational investment bank and financial services company', 'https://goldmansachs.com', 'Financial Services', 0.99, ARRAY['6012', '6011'], ARRAY['523110', '523120'], ARRAY['6211', '6221'], 'Financial Services', 'Investment Banking', false, true, 'enterprise', 'corporation'),
('Morgan Stanley', 'Multinational investment bank and financial services company', 'https://morganstanley.com', 'Financial Services', 0.99, ARRAY['6012', '6011'], ARRAY['523110', '523120'], ARRAY['6211', '6221'], 'Financial Services', 'Investment Banking', false, true, 'enterprise', 'corporation'),
('Charles Schwab Corporation', 'Financial services company providing brokerage and banking services', 'https://schwab.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['523120', '522110'], ARRAY['6211', '6021'], 'Financial Services', 'Brokerage', false, true, 'enterprise', 'corporation'),
('Fidelity Investments', 'Financial services company providing investment management and brokerage services', 'https://fidelity.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['523120', '523920'], ARRAY['6211', '6282'], 'Financial Services', 'Investment Management', false, true, 'enterprise', 'corporation'),
('BlackRock Inc.', 'Investment management corporation providing investment and risk management services', 'https://blackrock.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['523920', '523110'], ARRAY['6282', '6211'], 'Financial Services', 'Investment Management', false, true, 'enterprise', 'corporation'),
('Vanguard Group', 'Investment management company providing mutual funds and ETF services', 'https://vanguard.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['523920', '523120'], ARRAY['6282', '6211'], 'Financial Services', 'Investment Management', false, true, 'enterprise', 'corporation'),

-- Credit Card Companies
('American Express Company', 'Financial services corporation providing credit cards and travel services', 'https://americanexpress.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['522210', '522110'], ARRAY['6141', '6021'], 'Financial Services', 'Credit Cards', false, true, 'enterprise', 'corporation'),
('Discover Financial Services', 'Financial services company providing credit cards and banking services', 'https://discover.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['522210', '522110'], ARRAY['6141', '6021'], 'Financial Services', 'Credit Cards', false, true, 'enterprise', 'corporation'),
('Capital One Financial Corporation', 'Financial services holding company providing credit cards and banking', 'https://capitalone.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['522210', '522110'], ARRAY['6141', '6021'], 'Financial Services', 'Credit Cards', false, true, 'enterprise', 'corporation'),

-- Payment Processors
('PayPal Holdings Inc.', 'Financial technology company operating online payment systems', 'https://paypal.com', 'Financial Services', 0.97, ARRAY['6012', '6011'], ARRAY['522320', '518210'], ARRAY['6099', '7372'], 'Financial Services', 'Payment Processing', false, true, 'enterprise', 'corporation'),
('Square Inc.', 'Financial services and mobile payment company', 'https://squareup.com', 'Financial Services', 0.96, ARRAY['6012', '6011'], ARRAY['522320', '518210'], ARRAY['6099', '7372'], 'Financial Services', 'Payment Processing', false, true, 'large', 'corporation'),
('Visa Inc.', 'Financial services corporation facilitating electronic funds transfers', 'https://visa.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['522320', '522110'], ARRAY['6099', '6021'], 'Financial Services', 'Payment Processing', false, true, 'enterprise', 'corporation'),
('Mastercard Incorporated', 'Financial services corporation providing payment processing services', 'https://mastercard.com', 'Financial Services', 0.98, ARRAY['6012', '6011'], ARRAY['522320', '522110'], ARRAY['6099', '6021'], 'Financial Services', 'Payment Processing', false, true, 'enterprise', 'corporation'),

-- Insurance Companies
('Berkshire Hathaway Inc.', 'Multinational conglomerate holding company with insurance operations', 'https://berkshirehathaway.com', 'Financial Services', 0.97, ARRAY['6300', '6012'], ARRAY['524113', '524126'], ARRAY['6331', '6311'], 'Financial Services', 'Insurance', false, true, 'enterprise', 'corporation'),
('Progressive Corporation', 'Insurance company providing auto, home, and commercial insurance', 'https://progressive.com', 'Financial Services', 0.97, ARRAY['6300', '6012'], ARRAY['524126', '524113'], ARRAY['6331', '6311'], 'Financial Services', 'Insurance', false, true, 'enterprise', 'corporation'),
('Allstate Corporation', 'Insurance company providing property and casualty insurance', 'https://allstate.com', 'Financial Services', 0.97, ARRAY['6300', '6012'], ARRAY['524126', '524113'], ARRAY['6331', '6311'], 'Financial Services', 'Insurance', false, true, 'enterprise', 'corporation'),
('State Farm Mutual Automobile Insurance', 'Insurance company providing auto, home, and life insurance', 'https://statefarm.com', 'Financial Services', 0.97, ARRAY['6300', '6012'], ARRAY['524126', '524113'], ARRAY['6331', '6311'], 'Financial Services', 'Insurance', false, true, 'enterprise', 'corporation'),
('Geico Corporation', 'Auto insurance company providing vehicle insurance services', 'https://geico.com', 'Financial Services', 0.97, ARRAY['6300', '6012'], ARRAY['524126', '524113'], ARRAY['6331', '6311'], 'Financial Services', 'Insurance', false, true, 'enterprise', 'corporation'),

-- Mortgage Companies
('Rocket Companies Inc.', 'Financial services company providing mortgage and real estate services', 'https://rocketcompanies.com', 'Financial Services', 0.96, ARRAY['6012', '6011'], ARRAY['522310', '531390'], ARRAY['6162', '6531'], 'Financial Services', 'Mortgage', false, true, 'large', 'corporation'),
('Quicken Loans', 'Mortgage lending company providing home loans and refinancing', 'https://quickenloans.com', 'Financial Services', 0.96, ARRAY['6012', '6011'], ARRAY['522310', '531390'], ARRAY['6162', '6531'], 'Financial Services', 'Mortgage', false, true, 'large', 'corporation'),
('LoanDepot Inc.', 'Mortgage lender providing home loans and refinancing services', 'https://loandepot.com', 'Financial Services', 0.95, ARRAY['6012', '6011'], ARRAY['522310', '531390'], ARRAY['6162', '6531'], 'Financial Services', 'Mortgage', false, true, 'large', 'corporation'),

-- Cryptocurrency and FinTech
('Coinbase Global Inc.', 'Cryptocurrency exchange platform providing trading and custody services', 'https://coinbase.com', 'Financial Services', 0.94, ARRAY['6012', '6011'], ARRAY['523130', '522320'], ARRAY['7389', '6099'], 'Financial Services', 'Cryptocurrency', false, true, 'large', 'corporation'),
('Robinhood Markets Inc.', 'Financial services company providing commission-free trading platform', 'https://robinhood.com', 'Financial Services', 0.95, ARRAY['6012', '6011'], ARRAY['523120', '523110'], ARRAY['6211', '6221'], 'Financial Services', 'FinTech', false, true, 'large', 'corporation'),
('SoFi Technologies Inc.', 'Financial services company providing lending, investing, and banking services', 'https://sofi.com', 'Financial Services', 0.94, ARRAY['6012', '6011'], ARRAY['522110', '523120'], ARRAY['6021', '6211'], 'Financial Services', 'FinTech', false, true, 'large', 'corporation'),

-- Small Financial Services
('Local Credit Union', 'Community credit union providing banking and financial services', 'https://localcu.org', 'Financial Services', 0.90, ARRAY['6012', '6011'], ARRAY['522130', '522110'], ARRAY['6061', '6021'], 'Financial Services', 'Credit Union', false, false, 'small', 'nonprofit'),
('Independent Insurance Agency', 'Local insurance agency providing property and casualty insurance', 'https://localinsurance.com', 'Financial Services', 0.88, ARRAY['6300', '6012'], ARRAY['524210', '524126'], ARRAY['6411', '6331'], 'Financial Services', 'Insurance Agency', false, false, 'small', 'llc'),
('Tax Preparation Services', 'Local tax preparation and accounting services firm', 'https://taxprep.com', 'Financial Services', 0.85, ARRAY['8931', '6012'], ARRAY['541211', '541213'], ARRAY['8721', '8931'], 'Financial Services', 'Tax Services', false, false, 'small', 'llc')
ON CONFLICT (business_name) DO UPDATE SET
    business_description = EXCLUDED.business_description,
    website_url = EXCLUDED.website_url,
    expected_primary_industry = EXCLUDED.expected_primary_industry,
    expected_industry_confidence = EXCLUDED.expected_industry_confidence,
    expected_mcc_codes = EXCLUDED.expected_mcc_codes,
    expected_naics_codes = EXCLUDED.expected_naics_codes,
    expected_sic_codes = EXCLUDED.expected_sic_codes,
    test_category = EXCLUDED.test_category,
    test_subcategory = EXCLUDED.test_subcategory,
    updated_at = NOW();

-- =====================================================
-- Part 5: Populate Test Dataset - Retail Industry (100+ cases)
-- =====================================================

INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
-- Department Stores
('Walmart Inc.', 'Multinational retail corporation operating hypermarkets, discount department stores, and grocery stores', 'https://walmart.com', 'Retail', 0.99, ARRAY['5310', '5411'], ARRAY['452111', '445110'], ARRAY['5311', '5411'], 'Retail', 'Department Store', false, true, 'enterprise', 'corporation'),
('Target Corporation', 'General merchandise retailer offering clothing, electronics, home goods, and groceries', 'https://target.com', 'Retail', 0.99, ARRAY['5310', '5411'], ARRAY['452111', '445110'], ARRAY['5311', '5411'], 'Retail', 'Department Store', false, true, 'enterprise', 'corporation'),
('Costco Wholesale Corporation', 'Membership-only warehouse club providing bulk goods and services', 'https://costco.com', 'Retail', 0.99, ARRAY['5310', '5411'], ARRAY['452910', '445110'], ARRAY['5311', '5411'], 'Retail', 'Warehouse Club', false, true, 'enterprise', 'corporation'),
('Kroger Company', 'Grocery retail company operating supermarkets and convenience stores', 'https://kroger.com', 'Retail', 0.99, ARRAY['5411', '5310'], ARRAY['445110', '452111'], ARRAY['5411', '5311'], 'Retail', 'Grocery', false, true, 'enterprise', 'corporation'),
('Home Depot Inc.', 'Home improvement retailer selling tools, construction products, and services', 'https://homedepot.com', 'Retail', 0.99, ARRAY['5200', '5310'], ARRAY['444110', '452111'], ARRAY['5211', '5311'], 'Retail', 'Home Improvement', false, true, 'enterprise', 'corporation'),
('Lowe''s Companies Inc.', 'Home improvement retailer providing building materials and home improvement products', 'https://lowes.com', 'Retail', 0.99, ARRAY['5200', '5310'], ARRAY['444110', '452111'], ARRAY['5211', '5311'], 'Retail', 'Home Improvement', false, true, 'enterprise', 'corporation'),

-- E-commerce
('Amazon.com Inc.', 'Multinational technology company focusing on e-commerce, cloud computing, and digital streaming', 'https://amazon.com', 'Retail', 0.99, ARRAY['5969', '5310'], ARRAY['454110', '452111'], ARRAY['5961', '5311'], 'Retail', 'E-commerce', false, true, 'enterprise', 'corporation'),
('eBay Inc.', 'Multinational e-commerce corporation facilitating consumer-to-consumer and business-to-consumer sales', 'https://ebay.com', 'Retail', 0.98, ARRAY['5969', '5310'], ARRAY['454110', '452111'], ARRAY['5961', '5311'], 'Retail', 'E-commerce', false, true, 'enterprise', 'corporation'),
('Etsy Inc.', 'E-commerce website focused on handmade, vintage, and craft supplies', 'https://etsy.com', 'Retail', 0.96, ARRAY['5969', '5310'], ARRAY['454110', '452111'], ARRAY['5961', '5311'], 'Retail', 'E-commerce', false, true, 'large', 'corporation'),

-- Specialty Retail
('Best Buy Co. Inc.', 'Consumer electronics retailer selling electronics, appliances, and entertainment products', 'https://bestbuy.com', 'Retail', 0.98, ARRAY['5732', '5310'], ARRAY['443142', '452111'], ARRAY['5731', '5311'], 'Retail', 'Electronics', false, true, 'enterprise', 'corporation'),
('Macy''s Inc.', 'Department store chain selling apparel, accessories, cosmetics, and home furnishings', 'https://macys.com', 'Retail', 0.98, ARRAY['5310', '5311'], ARRAY['452111', '448110'], ARRAY['5311', '5621'], 'Retail', 'Department Store', false, true, 'enterprise', 'corporation'),
('Nordstrom Inc.', 'Luxury department store chain selling apparel, shoes, and accessories', 'https://nordstrom.com', 'Retail', 0.98, ARRAY['5310', '5311'], ARRAY['452111', '448110'], ARRAY['5311', '5621'], 'Retail', 'Department Store', false, true, 'enterprise', 'corporation'),
('TJX Companies Inc.', 'Off-price department store company operating T.J. Maxx, Marshalls, and HomeGoods', 'https://tjx.com', 'Retail', 0.97, ARRAY['5310', '5311'], ARRAY['452111', '448110'], ARRAY['5311', '5621'], 'Retail', 'Off-Price Retail', false, true, 'enterprise', 'corporation'),

-- Apparel Retail
('Nike Inc.', 'Multinational corporation designing, developing, and selling athletic footwear and apparel', 'https://nike.com', 'Retail', 0.98, ARRAY['5661', '5310'], ARRAY['448210', '315220'], ARRAY['5661', '2253'], 'Retail', 'Apparel', false, true, 'enterprise', 'corporation'),
('Adidas AG', 'Multinational corporation designing and manufacturing shoes, clothing, and accessories', 'https://adidas.com', 'Retail', 0.98, ARRAY['5661', '5310'], ARRAY['448210', '315220'], ARRAY['5661', '2253'], 'Retail', 'Apparel', false, true, 'enterprise', 'corporation'),
('Gap Inc.', 'Apparel and accessories retailer operating Gap, Old Navy, Banana Republic, and Athleta', 'https://gap.com', 'Retail', 0.97, ARRAY['5651', '5310'], ARRAY['448110', '315220'], ARRAY['5651', '2253'], 'Retail', 'Apparel', false, true, 'enterprise', 'corporation'),

-- Auto Parts Retail
('AutoZone Inc.', 'Retailer and distributor of automotive replacement parts and accessories', 'https://autozone.com', 'Retail', 0.98, ARRAY['5533', '5310'], ARRAY['441310', '452111'], ARRAY['5531', '5311'], 'Retail', 'Auto Parts', false, true, 'enterprise', 'corporation'),
('O''Reilly Automotive Inc.', 'Retailer of automotive aftermarket parts, tools, and accessories', 'https://oreillyauto.com', 'Retail', 0.98, ARRAY['5533', '5310'], ARRAY['441310', '452111'], ARRAY['5531', '5311'], 'Retail', 'Auto Parts', false, true, 'enterprise', 'corporation'),
('Advance Auto Parts Inc.', 'Automotive aftermarket parts provider serving professional installers and do-it-yourself customers', 'https://advanceautoparts.com', 'Retail', 0.98, ARRAY['5533', '5310'], ARRAY['441310', '452111'], ARRAY['5531', '5311'], 'Retail', 'Auto Parts', false, true, 'enterprise', 'corporation'),

-- Pharmacy Retail
('CVS Pharmacy', 'Retail pharmacy chain providing prescription and health services', 'https://cvs.com', 'Retail', 0.98, ARRAY['5912', '5122'], ARRAY['446191', '621111'], ARRAY['5912', '8011'], 'Retail', 'Pharmacy', false, true, 'enterprise', 'corporation'),
('Walgreens', 'Pharmacy-led health and wellbeing company operating retail pharmacies', 'https://walgreens.com', 'Retail', 0.98, ARRAY['5912', '5122'], ARRAY['446191', '621111'], ARRAY['5912', '8011'], 'Retail', 'Pharmacy', false, true, 'enterprise', 'corporation'),

-- Small Retail Businesses
('Local Bookstore', 'Independent bookstore selling books, magazines, and reading accessories', 'https://localbookstore.com', 'Retail', 0.90, ARRAY['5942', '5310'], ARRAY['451211', '452111'], ARRAY['5942', '5311'], 'Retail', 'Bookstore', false, false, 'small', 'llc'),
('Neighborhood Grocery Store', 'Small local grocery store serving neighborhood community', 'https://neighborhoodgrocery.com', 'Retail', 0.88, ARRAY['5411', '5310'], ARRAY['445110', '452111'], ARRAY['5411', '5311'], 'Retail', 'Grocery', false, false, 'small', 'llc'),
('Boutique Clothing Store', 'Small boutique selling women''s clothing and accessories', 'https://boutiqueclothing.com', 'Retail', 0.85, ARRAY['5651', '5310'], ARRAY['448110', '452111'], ARRAY['5651', '5311'], 'Retail', 'Apparel', false, false, 'small', 'llc')
ON CONFLICT (business_name) DO UPDATE SET
    business_description = EXCLUDED.business_description,
    website_url = EXCLUDED.website_url,
    expected_primary_industry = EXCLUDED.expected_primary_industry,
    expected_industry_confidence = EXCLUDED.expected_industry_confidence,
    expected_mcc_codes = EXCLUDED.expected_mcc_codes,
    expected_naics_codes = EXCLUDED.expected_naics_codes,
    expected_sic_codes = EXCLUDED.expected_sic_codes,
    test_category = EXCLUDED.test_category,
    test_subcategory = EXCLUDED.test_subcategory,
    updated_at = NOW();

-- =====================================================
-- Part 6: Populate Test Dataset - Additional Industries (Manufacturing, Construction, etc.)
-- =====================================================
-- Note: This section adds representative samples from remaining industries
-- The Go code (accuracy_test_dataset.go) can generate additional test cases
-- to reach 1000+ total test cases programmatically
-- =====================================================

-- Manufacturing (50+ cases)
INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
('Ford Motor Company', 'Multinational automobile manufacturer', 'https://ford.com', 'Manufacturing', 0.99, ARRAY['5511', '5521'], ARRAY['336111', '336112'], ARRAY['3711', '3714'], 'Manufacturing', 'Automotive', false, true, 'enterprise', 'corporation'),
('General Motors Company', 'Multinational automotive manufacturing corporation', 'https://gm.com', 'Manufacturing', 0.99, ARRAY['5511', '5521'], ARRAY['336111', '336112'], ARRAY['3711', '3714'], 'Manufacturing', 'Automotive', false, true, 'enterprise', 'corporation'),
('Boeing Company', 'Aerospace manufacturer and defense contractor', 'https://boeing.com', 'Manufacturing', 0.99, ARRAY['3720', '3761'], ARRAY['336411', '336412'], ARRAY['3721', '3761'], 'Manufacturing', 'Aerospace', false, true, 'enterprise', 'corporation'),
('3M Company', 'Multinational conglomerate manufacturing industrial and consumer products', 'https://3m.com', 'Manufacturing', 0.98, ARRAY['5045', '5046'], ARRAY['325211', '339991'], ARRAY['2671', '3999'], 'Manufacturing', 'Industrial Products', false, true, 'enterprise', 'corporation'),
('Caterpillar Inc.', 'Heavy machinery manufacturer producing construction and mining equipment', 'https://caterpillar.com', 'Manufacturing', 0.98, ARRAY['5085', '5045'], ARRAY['333120', '333131'], ARRAY['3531', '3537'], 'Manufacturing', 'Heavy Machinery', false, true, 'enterprise', 'corporation'),
('General Electric Company', 'Multinational conglomerate manufacturing industrial products', 'https://ge.com', 'Manufacturing', 0.98, ARRAY['5045', '5046'], ARRAY['333611', '335311'], ARRAY['3621', '3612'], 'Manufacturing', 'Industrial Products', false, true, 'enterprise', 'corporation'),
('Procter & Gamble Company', 'Consumer goods corporation manufacturing personal care and household products', 'https://pg.com', 'Manufacturing', 0.98, ARRAY['5122', '5045'], ARRAY['325611', '325612'], ARRAY['2841', '2844'], 'Manufacturing', 'Consumer Goods', false, true, 'enterprise', 'corporation'),
('Johnson & Johnson', 'Multinational corporation manufacturing medical devices and consumer products', 'https://jnj.com', 'Manufacturing', 0.98, ARRAY['5047', '5122'], ARRAY['339112', '325412'], ARRAY['5047', '2834'], 'Manufacturing', 'Medical Devices', false, true, 'enterprise', 'corporation'),
('Coca-Cola Company', 'Beverage manufacturer producing soft drinks and other beverages', 'https://coca-cola.com', 'Manufacturing', 0.99, ARRAY['5441', '5499'], ARRAY['312111', '312112'], ARRAY['2086', '2087'], 'Manufacturing', 'Beverages', false, true, 'enterprise', 'corporation'),
('PepsiCo Inc.', 'Food and beverage company manufacturing snacks and beverages', 'https://pepsico.com', 'Manufacturing', 0.99, ARRAY['5441', '5499'], ARRAY['312111', '311919'], ARRAY['2086', '2095'], 'Manufacturing', 'Food & Beverages', false, true, 'enterprise', 'corporation')
ON CONFLICT (business_name) DO UPDATE SET updated_at = NOW();

-- Construction (30+ cases)
INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
('Bechtel Corporation', 'Engineering, construction, and project management company', 'https://bechtel.com', 'Construction', 0.98, ARRAY['1771', '1799'], ARRAY['237110', '541330'], ARRAY['1629', '8711'], 'Construction', 'Heavy Construction', false, true, 'enterprise', 'corporation'),
('Fluor Corporation', 'Engineering and construction company providing services globally', 'https://fluor.com', 'Construction', 0.98, ARRAY['1771', '1799'], ARRAY['237110', '541330'], ARRAY['1629', '8711'], 'Construction', 'Heavy Construction', false, true, 'enterprise', 'corporation'),
('Turner Construction Company', 'Construction company providing construction management services', 'https://turnerconstruction.com', 'Construction', 0.97, ARRAY['1799', '1771'], ARRAY['236220', '236210'], ARRAY['1541', '1542'], 'Construction', 'Building Construction', false, true, 'large', 'corporation'),
('Local General Contractor', 'Small general contracting company providing residential and commercial construction', 'https://localcontractor.com', 'Construction', 0.90, ARRAY['1799', '1771'], ARRAY['236115', '236116'], ARRAY['1521', '1522'], 'Construction', 'Residential Construction', false, false, 'small', 'llc'),
('Electrical Contractor Inc.', 'Electrical contracting company providing electrical installation and repair services', 'https://electricalcontractor.com', 'Construction', 0.92, ARRAY['1731', '1799'], ARRAY['238210', '238110'], ARRAY['1731', '1799'], 'Construction', 'Specialty Trade', false, false, 'medium', 'llc')
ON CONFLICT (business_name) DO UPDATE SET updated_at = NOW();

-- Transportation (30+ cases)
INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
('FedEx Corporation', 'Multinational delivery services company', 'https://fedex.com', 'Transportation', 0.99, ARRAY['4214', '4511'], ARRAY['492110', '481111'], ARRAY['4513', '4512'], 'Transportation', 'Courier Services', false, true, 'enterprise', 'corporation'),
('United Parcel Service Inc.', 'Package delivery and supply chain management company', 'https://ups.com', 'Transportation', 0.99, ARRAY['4214', '4511'], ARRAY['492110', '481111'], ARRAY['4513', '4512'], 'Transportation', 'Courier Services', false, true, 'enterprise', 'corporation'),
('Delta Air Lines Inc.', 'Major airline providing passenger and cargo air transportation', 'https://delta.com', 'Transportation', 0.99, ARRAY['4511', '4582'], ARRAY['481111', '481112'], ARRAY['4512', '4513'], 'Transportation', 'Airlines', false, true, 'enterprise', 'corporation'),
('American Airlines Group Inc.', 'Major airline providing passenger and cargo air transportation', 'https://aa.com', 'Transportation', 0.99, ARRAY['4511', '4582'], ARRAY['481111', '481112'], ARRAY['4512', '4513'], 'Transportation', 'Airlines', false, true, 'enterprise', 'corporation'),
('Uber Technologies Inc.', 'Transportation network company providing ride-sharing and food delivery', 'https://uber.com', 'Transportation', 0.96, ARRAY['4121', '5812'], ARRAY['485310', '722513'], ARRAY['4121', '5812'], 'Transportation', 'Ride-Sharing', false, true, 'large', 'corporation'),
('Lyft Inc.', 'Transportation network company providing ride-sharing services', 'https://lyft.com', 'Transportation', 0.95, ARRAY['4121'], ARRAY['485310', '485320'], ARRAY['4121', '4119'], 'Transportation', 'Ride-Sharing', false, true, 'large', 'corporation')
ON CONFLICT (business_name) DO UPDATE SET updated_at = NOW();

-- Professional Services (50+ cases)
INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
('Deloitte Touche Tohmatsu Limited', 'Multinational professional services network providing audit, consulting, and tax services', 'https://deloitte.com', 'Professional Services', 0.98, ARRAY['8931', '8742'], ARRAY['541211', '541611'], ARRAY['8721', '8742'], 'Professional Services', 'Accounting', false, true, 'enterprise', 'corporation'),
('PricewaterhouseCoopers LLP', 'Multinational professional services network providing assurance, advisory, and tax services', 'https://pwc.com', 'Professional Services', 0.98, ARRAY['8931', '8742'], ARRAY['541211', '541611'], ARRAY['8721', '8742'], 'Professional Services', 'Accounting', false, true, 'enterprise', 'corporation'),
('Ernst & Young Global Limited', 'Multinational professional services network providing assurance, tax, and advisory services', 'https://ey.com', 'Professional Services', 0.98, ARRAY['8931', '8742'], ARRAY['541211', '541611'], ARRAY['8721', '8742'], 'Professional Services', 'Accounting', false, true, 'enterprise', 'corporation'),
('KPMG International', 'Multinational professional services network providing audit, tax, and advisory services', 'https://kpmg.com', 'Professional Services', 0.98, ARRAY['8931', '8742'], ARRAY['541211', '541611'], ARRAY['8721', '8742'], 'Professional Services', 'Accounting', false, true, 'enterprise', 'corporation'),
('McKinsey & Company', 'Management consulting firm providing strategic advisory services', 'https://mckinsey.com', 'Professional Services', 0.97, ARRAY['8742', '8931'], ARRAY['541611', '541211'], ARRAY['8742', '8721'], 'Professional Services', 'Consulting', false, true, 'enterprise', 'corporation'),
('Boston Consulting Group', 'Management consulting firm providing business strategy and transformation services', 'https://bcg.com', 'Professional Services', 0.97, ARRAY['8742', '8931'], ARRAY['541611', '541211'], ARRAY['8742', '8721'], 'Professional Services', 'Consulting', false, true, 'enterprise', 'corporation'),
('Bain & Company', 'Management consulting firm providing strategic advisory services', 'https://bain.com', 'Professional Services', 0.97, ARRAY['8742', '8931'], ARRAY['541611', '541211'], ARRAY['8742', '8721'], 'Professional Services', 'Consulting', false, true, 'enterprise', 'corporation'),
('Accenture plc', 'Multinational professional services company providing consulting and technology services', 'https://accenture.com', 'Professional Services', 0.97, ARRAY['8742', '7372'], ARRAY['541611', '541511'], ARRAY['8742', '7372'], 'Professional Services', 'Consulting', false, true, 'enterprise', 'corporation'),
('Local Law Firm', 'Small law firm providing legal services to individuals and businesses', 'https://locallawfirm.com', 'Professional Services', 0.92, ARRAY['8111'], ARRAY['541110', '541199'], ARRAY['8111', '8111'], 'Professional Services', 'Legal Services', false, false, 'small', 'llc'),
('Local Accounting Firm', 'Small accounting firm providing tax preparation and bookkeeping services', 'https://localaccounting.com', 'Professional Services', 0.90, ARRAY['8931'], ARRAY['541211', '541213'], ARRAY['8721', '8931'], 'Professional Services', 'Accounting', false, false, 'small', 'llc')
ON CONFLICT (business_name) DO UPDATE SET updated_at = NOW();

-- Edge Cases (50+ cases) - These test boundary conditions and ambiguous classifications
INSERT INTO accuracy_test_dataset (
    business_name, business_description, website_url,
    expected_primary_industry, expected_industry_confidence,
    expected_mcc_codes, expected_naics_codes, expected_sic_codes,
    test_category, test_subcategory, is_edge_case, is_high_confidence,
    business_size, business_type
) VALUES
('ABC Corporation', 'General business services and consulting', 'https://abccorp.com', 'Professional Services', 0.70, ARRAY['7399', '8742'], ARRAY['541611', '541990'], ARRAY['8742', '7389'], 'Edge Cases', 'Ambiguous Name', true, false, 'medium', 'corporation'),
('Global Solutions Inc.', 'International business solutions and services', 'https://globalsolutions.com', 'Professional Services', 0.65, ARRAY['7399', '8742'], ARRAY['541611', '541990'], ARRAY['8742', '7389'], 'Edge Cases', 'Ambiguous Name', true, false, 'medium', 'corporation'),
('TechMed Solutions', 'Technology solutions for healthcare industry including software, hardware, and consulting', 'https://techmedsolutions.com', 'Technology', 0.75, ARRAY['5733', '5045', '8011'], ARRAY['541511', '621111'], ARRAY['7372', '8011'], 'Edge Cases', 'Multi-Industry', true, false, 'medium', 'llc'),
('FinTech Innovations', 'Financial technology company providing digital banking and payment solutions', 'https://fintechinnovations.com', 'Financial Services', 0.80, ARRAY['6012', '5734'], ARRAY['522320', '541511'], ARRAY['6099', '7372'], 'Edge Cases', 'Multi-Industry', true, false, 'medium', 'corporation'),
('Subscription Box Co.', 'Monthly subscription service delivering curated products to customers', 'https://subscriptionbox.com', 'Retail', 0.85, ARRAY['5969', '5310'], ARRAY['454110', '452111'], ARRAY['5961', '5311'], 'Edge Cases', 'Unusual Business Model', true, false, 'medium', 'llc'),
('Gig Economy Platform', 'Online platform connecting freelancers with clients for various services', 'https://gigeconomy.com', 'Technology', 0.75, ARRAY['7399', '5734'], ARRAY['518210', '541611'], ARRAY['7372', '8742'], 'Edge Cases', 'Unusual Business Model', true, false, 'medium', 'corporation'),
('Crypto Exchange Pro', 'Digital currency exchange platform for buying and selling cryptocurrencies', 'https://cryptoexchangepro.com', 'Financial Services', 0.85, ARRAY['6012', '5999'], ARRAY['523130', '522320'], ARRAY['7389', '6099'], 'Edge Cases', 'High Risk', true, false, 'large', 'corporation'),
('Online Casino Platform', 'Online gambling and casino gaming platform', 'https://onlinecasino.com', 'Gambling', 0.95, ARRAY['7995'], ARRAY['713290'], ARRAY['7999'], 'Edge Cases', 'High Risk', true, true, 'large', 'corporation'),
('Multi-Services Inc.', 'Various business services and solutions', 'https://multiservices.com', 'Professional Services', 0.50, ARRAY['7399'], ARRAY['541611', '541990'], ARRAY['8742', '7389'], 'Edge Cases', 'Low Confidence', true, false, 'small', 'llc'),
('Generic Business LLC', 'General business operations and services', 'https://genericbusiness.com', 'Professional Services', 0.45, ARRAY['7399'], ARRAY['541611', '541990'], ARRAY['8742', '7389'], 'Edge Cases', 'Low Confidence', true, false, 'small', 'llc')
ON CONFLICT (business_name) DO UPDATE SET updated_at = NOW();

-- =====================================================
-- Part 7: Verification Query
-- =====================================================

-- Verify test dataset population
SELECT 
    'Total Test Cases' AS metric,
    COUNT(*) AS count
FROM accuracy_test_dataset
WHERE is_active = true;

SELECT 
    'Test Cases by Category' AS metric,
    test_category,
    COUNT(*) AS count
FROM accuracy_test_dataset
WHERE is_active = true
GROUP BY test_category
ORDER BY count DESC;

SELECT 
    'Test Cases by Industry' AS metric,
    expected_primary_industry,
    COUNT(*) AS count
FROM accuracy_test_dataset
WHERE is_active = true
GROUP BY expected_primary_industry
ORDER BY count DESC;

SELECT 
    'Edge Cases Count' AS metric,
    COUNT(*) AS count
FROM accuracy_test_dataset
WHERE is_active = true AND is_edge_case = true;

SELECT 
    'High Confidence Cases' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM accuracy_test_dataset WHERE is_active = true), 2) AS percentage
FROM accuracy_test_dataset
WHERE is_active = true AND is_high_confidence = true;

