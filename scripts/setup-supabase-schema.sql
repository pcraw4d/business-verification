-- KYB Platform - Supabase Database Schema Setup
-- Run this script in your Supabase SQL Editor

-- Users table (extends Supabase auth.users)
CREATE TABLE IF NOT EXISTS public.profiles (
    id UUID REFERENCES auth.users(id) PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT,
    role TEXT CHECK (role IN ('compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'other')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Business classifications table
CREATE TABLE IF NOT EXISTS public.business_classifications (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    business_name TEXT NOT NULL,
    website_url TEXT,
    description TEXT,
    primary_industry JSONB,
    secondary_industries JSONB,
    confidence_score DECIMAL(3,2),
    classification_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Risk assessments table
CREATE TABLE IF NOT EXISTS public.risk_assessments (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    business_id UUID REFERENCES public.business_classifications(id),
    risk_factors JSONB,
    risk_score DECIMAL(3,2),
    risk_level TEXT CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    assessment_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Compliance checks table
CREATE TABLE IF NOT EXISTS public.compliance_checks (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    business_id UUID REFERENCES public.business_classifications(id),
    compliance_frameworks JSONB,
    compliance_status JSONB,
    gap_analysis JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Feedback table
CREATE TABLE IF NOT EXISTS public.feedback (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    feedback_type TEXT CHECK (feedback_type IN ('bug', 'feature', 'improvement', 'general')),
    message TEXT NOT NULL,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable Row Level Security (RLS) after all tables are created
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.business_classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.risk_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.compliance_checks ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.feedback ENABLE ROW LEVEL SECURITY;

-- Create indexes for better performance (after tables are created)
CREATE INDEX IF NOT EXISTS idx_business_classifications_user_id ON public.business_classifications(user_id);
CREATE INDEX IF NOT EXISTS idx_business_classifications_created_at ON public.business_classifications(created_at);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_user_id ON public.risk_assessments(user_id);
CREATE INDEX IF NOT EXISTS idx_compliance_checks_user_id ON public.compliance_checks(user_id);
CREATE INDEX IF NOT EXISTS idx_feedback_user_id ON public.feedback(user_id);

-- Insert sample data for testing (optional)
INSERT INTO public.profiles (id, email, full_name, role) 
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'admin@kybplatform.com', 'Admin User', 'compliance_officer')
ON CONFLICT (id) DO NOTHING;

-- Grant necessary permissions
GRANT USAGE ON SCHEMA public TO anon, authenticated;
GRANT ALL ON ALL TABLES IN SCHEMA public TO anon, authenticated;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO anon, authenticated;

-- Create RLS policies (after tables, indexes, and data are created)
-- Drop existing policies if they exist (to avoid conflicts)
DROP POLICY IF EXISTS "Users can view own profile" ON public.profiles;
DROP POLICY IF EXISTS "Users can update own profile" ON public.profiles;
DROP POLICY IF EXISTS "Users can view own classifications" ON public.business_classifications;
DROP POLICY IF EXISTS "Users can insert own classifications" ON public.business_classifications;
DROP POLICY IF EXISTS "Users can view own risk assessments" ON public.risk_assessments;
DROP POLICY IF EXISTS "Users can insert own risk assessments" ON public.risk_assessments;
DROP POLICY IF EXISTS "Users can view own compliance checks" ON public.compliance_checks;
DROP POLICY IF EXISTS "Users can insert own compliance checks" ON public.compliance_checks;
DROP POLICY IF EXISTS "Users can view own feedback" ON public.feedback;
DROP POLICY IF EXISTS "Users can insert own feedback" ON public.feedback;

-- Create new policies
CREATE POLICY "Users can view own profile" ON public.profiles
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON public.profiles
    FOR UPDATE USING (auth.uid() = id);

CREATE POLICY "Users can view own classifications" ON public.business_classifications
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own classifications" ON public.business_classifications
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can view own risk assessments" ON public.risk_assessments
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own risk assessments" ON public.risk_assessments
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can view own compliance checks" ON public.compliance_checks
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own compliance checks" ON public.compliance_checks
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can view own feedback" ON public.feedback
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own feedback" ON public.feedback
    FOR INSERT WITH CHECK (auth.uid() = user_id);
