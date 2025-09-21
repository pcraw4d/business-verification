-- KYB Platform - Supabase Database Schema Setup
-- Run this script in your Supabase SQL Editor

-- Drop existing tables if they exist (for clean setup)
DROP TABLE IF EXISTS public.feedback CASCADE;
DROP TABLE IF EXISTS public.compliance_checks CASCADE;
DROP TABLE IF EXISTS public.risk_assessments CASCADE;
DROP TABLE IF EXISTS public.business_classifications CASCADE;
DROP TABLE IF EXISTS public.users_consolidated CASCADE;

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Consolidated Users table (extends Supabase auth.users)
CREATE TABLE IF NOT EXISTS public.users_consolidated (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Authentication fields
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255),
    
    -- Profile information
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    full_name VARCHAR(255), -- Computed field for compatibility
    name VARCHAR(255), -- Single name field for compatibility
    
    -- Business information
    company VARCHAR(255),
    
    -- Role and permissions
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN (
        'user', 'admin', 'compliance_officer', 'risk_manager', 
        'business_analyst', 'developer', 'other'
    )),
    
    -- Account status and security
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN (
        'active', 'inactive', 'suspended', 'pending_verification'
    )),
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Email verification
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    
    -- Security features
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    
    -- Activity tracking
    last_login_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata and extensibility
    metadata JSONB DEFAULT '{}',
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT users_consolidated_email_check CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT users_consolidated_username_check CHECK (username IS NULL OR length(username) >= 3),
    CONSTRAINT users_consolidated_name_check CHECK (
        (first_name IS NOT NULL AND last_name IS NOT NULL) OR 
        (full_name IS NOT NULL) OR 
        (name IS NOT NULL)
    )
);

-- Business classifications table
CREATE TABLE IF NOT EXISTS public.business_classifications (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.users_consolidated(id) NOT NULL,
    business_name TEXT NOT NULL,
    website_url TEXT,
    description TEXT,
    primary_industry JSONB,
    secondary_industries JSONB,
    confidence_score DECIMAL(3,2) CHECK (confidence_score >= 0 AND confidence_score <= 1),
    classification_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Risk assessments table
CREATE TABLE IF NOT EXISTS public.risk_assessments (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.users_consolidated(id) NOT NULL,
    business_id UUID REFERENCES public.business_classifications(id) ON DELETE CASCADE NOT NULL,
    risk_factors JSONB,
    risk_score DECIMAL(3,2) CHECK (risk_score >= 0 AND risk_score <= 1),
    risk_level TEXT CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    assessment_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Compliance checks table
CREATE TABLE IF NOT EXISTS public.compliance_checks (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.users_consolidated(id) NOT NULL,
    business_id UUID REFERENCES public.business_classifications(id) ON DELETE CASCADE NOT NULL,
    compliance_frameworks JSONB,
    compliance_status JSONB,
    gap_analysis JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Feedback table
CREATE TABLE IF NOT EXISTS public.feedback (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.users_consolidated(id) NOT NULL,
    feedback_type TEXT CHECK (feedback_type IN ('bug', 'feature', 'improvement', 'general')),
    message TEXT NOT NULL,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable Row Level Security (RLS) after all tables are created
ALTER TABLE public.users_consolidated ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.business_classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.risk_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.compliance_checks ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.feedback ENABLE ROW LEVEL SECURITY;

-- Create indexes for better performance (after tables are created)
CREATE INDEX IF NOT EXISTS idx_business_classifications_user_id ON public.business_classifications(user_id);
CREATE INDEX IF NOT EXISTS idx_business_classifications_created_at ON public.business_classifications(created_at);
CREATE INDEX IF NOT EXISTS idx_business_classifications_confidence_score ON public.business_classifications(confidence_score);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_user_id ON public.risk_assessments(user_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_business_id ON public.risk_assessments(business_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_risk_score ON public.risk_assessments(risk_score);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_risk_level ON public.risk_assessments(risk_level);
CREATE INDEX IF NOT EXISTS idx_compliance_checks_user_id ON public.compliance_checks(user_id);
CREATE INDEX IF NOT EXISTS idx_compliance_checks_business_id ON public.compliance_checks(business_id);
CREATE INDEX IF NOT EXISTS idx_feedback_user_id ON public.feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_feedback_status ON public.feedback(status);
CREATE INDEX IF NOT EXISTS idx_feedback_created_at ON public.feedback(created_at);

-- Create composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_business_classifications_user_created ON public.business_classifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_user_business ON public.risk_assessments(user_id, business_id);
CREATE INDEX IF NOT EXISTS idx_compliance_checks_user_business ON public.compliance_checks(user_id, business_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_users_consolidated_updated_at BEFORE UPDATE ON public.users_consolidated
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_business_classifications_updated_at BEFORE UPDATE ON public.business_classifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_risk_assessments_updated_at BEFORE UPDATE ON public.risk_assessments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_compliance_checks_updated_at BEFORE UPDATE ON public.compliance_checks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_feedback_updated_at BEFORE UPDATE ON public.feedback
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create JSONB validation functions
CREATE OR REPLACE FUNCTION validate_industry_jsonb()
RETURNS TRIGGER AS $$
BEGIN
    -- Validate primary_industry structure
    IF NEW.primary_industry IS NOT NULL THEN
        IF NOT (NEW.primary_industry ? 'industry_code' AND NEW.primary_industry ? 'industry_name') THEN
            RAISE EXCEPTION 'primary_industry must contain industry_code and industry_name';
        END IF;
    END IF;
    
    -- Validate secondary_industries structure
    IF NEW.secondary_industries IS NOT NULL THEN
        IF jsonb_typeof(NEW.secondary_industries) != 'array' THEN
            RAISE EXCEPTION 'secondary_industries must be an array';
        END IF;
        
        -- Validate each secondary industry
        FOR i IN 0..jsonb_array_length(NEW.secondary_industries) - 1 LOOP
            IF NOT (NEW.secondary_industries->i ? 'industry_code' AND NEW.secondary_industries->i ? 'industry_name') THEN
                RAISE EXCEPTION 'secondary_industries[%] must contain industry_code and industry_name', i;
            END IF;
        END LOOP;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION validate_risk_factors_jsonb()
RETURNS TRIGGER AS $$
BEGIN
    -- Validate risk_factors structure
    IF NEW.risk_factors IS NOT NULL THEN
        IF jsonb_typeof(NEW.risk_factors) != 'object' THEN
            RAISE EXCEPTION 'risk_factors must be an object';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION validate_compliance_frameworks_jsonb()
RETURNS TRIGGER AS $$
BEGIN
    -- Validate compliance_frameworks structure
    IF NEW.compliance_frameworks IS NOT NULL THEN
        IF jsonb_typeof(NEW.compliance_frameworks) != 'array' THEN
            RAISE EXCEPTION 'compliance_frameworks must be an array';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create validation triggers
CREATE TRIGGER validate_business_classifications_industry BEFORE INSERT OR UPDATE ON public.business_classifications
    FOR EACH ROW EXECUTE FUNCTION validate_industry_jsonb();

CREATE TRIGGER validate_risk_assessments_factors BEFORE INSERT OR UPDATE ON public.risk_assessments
    FOR EACH ROW EXECUTE FUNCTION validate_risk_factors_jsonb();

CREATE TRIGGER validate_compliance_checks_frameworks BEFORE INSERT OR UPDATE ON public.compliance_checks
    FOR EACH ROW EXECUTE FUNCTION validate_compliance_frameworks_jsonb();

-- Grant necessary permissions
GRANT USAGE ON SCHEMA public TO anon, authenticated;
GRANT ALL ON ALL TABLES IN SCHEMA public TO anon, authenticated;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO anon, authenticated;

-- Create RLS policies (after tables, indexes, and data are created)
-- Drop existing policies if they exist (to avoid conflicts)
DROP POLICY IF EXISTS "Users can view own profile" ON public.users_consolidated;
DROP POLICY IF EXISTS "Users can update own profile" ON public.users_consolidated;
DROP POLICY IF EXISTS "Users can insert own profile" ON public.users_consolidated;
DROP POLICY IF EXISTS "Users can delete own profile" ON public.users_consolidated;

DROP POLICY IF EXISTS "Users can view own classifications" ON public.business_classifications;
DROP POLICY IF EXISTS "Users can insert own classifications" ON public.business_classifications;
DROP POLICY IF EXISTS "Users can update own classifications" ON public.business_classifications;
DROP POLICY IF EXISTS "Users can delete own classifications" ON public.business_classifications;

DROP POLICY IF EXISTS "Users can view own risk assessments" ON public.risk_assessments;
DROP POLICY IF EXISTS "Users can insert own risk assessments" ON public.risk_assessments;
DROP POLICY IF EXISTS "Users can update own risk assessments" ON public.risk_assessments;
DROP POLICY IF EXISTS "Users can delete own risk assessments" ON public.risk_assessments;

DROP POLICY IF EXISTS "Users can view own compliance checks" ON public.compliance_checks;
DROP POLICY IF EXISTS "Users can insert own compliance checks" ON public.compliance_checks;
DROP POLICY IF EXISTS "Users can update own compliance checks" ON public.compliance_checks;
DROP POLICY IF EXISTS "Users can delete own compliance checks" ON public.compliance_checks;

DROP POLICY IF EXISTS "Users can view own feedback" ON public.feedback;
DROP POLICY IF EXISTS "Users can insert own feedback" ON public.feedback;
DROP POLICY IF EXISTS "Users can update own feedback" ON public.feedback;
DROP POLICY IF EXISTS "Users can delete own feedback" ON public.feedback;

-- Create comprehensive policies for users_consolidated
CREATE POLICY "Users can view own profile" ON public.users_consolidated
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can insert own profile" ON public.users_consolidated
    FOR INSERT WITH CHECK (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON public.users_consolidated
    FOR UPDATE USING (auth.uid() = id);

CREATE POLICY "Users can delete own profile" ON public.users_consolidated
    FOR DELETE USING (auth.uid() = id);

-- Create comprehensive policies for business_classifications
CREATE POLICY "Users can view own classifications" ON public.business_classifications
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own classifications" ON public.business_classifications
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own classifications" ON public.business_classifications
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own classifications" ON public.business_classifications
    FOR DELETE USING (auth.uid() = user_id);

-- Create comprehensive policies for risk_assessments
CREATE POLICY "Users can view own risk assessments" ON public.risk_assessments
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own risk assessments" ON public.risk_assessments
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own risk assessments" ON public.risk_assessments
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own risk assessments" ON public.risk_assessments
    FOR DELETE USING (auth.uid() = user_id);

-- Create comprehensive policies for compliance_checks
CREATE POLICY "Users can view own compliance checks" ON public.compliance_checks
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own compliance checks" ON public.compliance_checks
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own compliance checks" ON public.compliance_checks
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own compliance checks" ON public.compliance_checks
    FOR DELETE USING (auth.uid() = user_id);

-- Create comprehensive policies for feedback
CREATE POLICY "Users can view own feedback" ON public.feedback
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own feedback" ON public.feedback
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own feedback" ON public.feedback
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own feedback" ON public.feedback
    FOR DELETE USING (auth.uid() = user_id);

-- Create materialized view for business classification statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS public.business_classification_stats AS
SELECT 
    user_id,
    COUNT(*) as total_classifications,
    AVG(confidence_score) as avg_confidence_score,
    COUNT(CASE WHEN confidence_score >= 0.8 THEN 1 END) as high_confidence_count,
    COUNT(CASE WHEN confidence_score < 0.5 THEN 1 END) as low_confidence_count,
    MAX(created_at) as last_classification_date
FROM public.business_classifications
GROUP BY user_id;

-- Create index on materialized view
CREATE INDEX IF NOT EXISTS idx_business_classification_stats_user_id ON public.business_classification_stats(user_id);

-- Create function to refresh materialized view
CREATE OR REPLACE FUNCTION refresh_business_classification_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW public.business_classification_stats;
END;
$$ language 'plpgsql';

-- Create trigger to refresh materialized view when business_classifications changes
CREATE OR REPLACE FUNCTION trigger_refresh_business_classification_stats()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM refresh_business_classification_stats();
    RETURN NULL;
END;
$$ language 'plpgsql';

CREATE TRIGGER refresh_business_classification_stats_trigger
    AFTER INSERT OR UPDATE OR DELETE ON public.business_classifications
    FOR EACH STATEMENT EXECUTE FUNCTION trigger_refresh_business_classification_stats();

-- Note: Sample data insertion removed to avoid foreign key constraint violations
-- To create a test user:
-- 1. Go to Authentication > Users in your Supabase dashboard
-- 2. Create a new user or use an existing one
-- 3. Then manually insert a profile record for that user:
--    INSERT INTO public.users_consolidated (id, email, full_name, role) 
--    VALUES ('actual-user-uuid-from-auth', 'user@example.com', 'Test User', 'compliance_officer');

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_consolidated_email ON users_consolidated(email);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_username ON users_consolidated(username) WHERE username IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_consolidated_role ON users_consolidated(role);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_status ON users_consolidated(status);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_is_active ON users_consolidated(is_active);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_created_at ON users_consolidated(created_at);
CREATE INDEX IF NOT EXISTS idx_users_consolidated_last_login ON users_consolidated(last_login_at);

-- Create compatibility views for backward compatibility
CREATE OR REPLACE VIEW users AS
SELECT 
    id,
    email,
    username,
    password_hash,
    first_name,
    last_name,
    full_name as name, -- Map full_name to name for compatibility
    company,
    role,
    status,
    email_verified,
    email_verified_at,
    last_login_at,
    is_active,
    metadata,
    created_at,
    updated_at
FROM users_consolidated;

CREATE OR REPLACE VIEW profiles AS
SELECT 
    id,
    email,
    full_name,
    role,
    created_at,
    updated_at
FROM users_consolidated;
