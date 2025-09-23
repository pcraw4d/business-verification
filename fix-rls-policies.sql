-- Fix RLS policies for classifications table
-- This script fixes the Row Level Security policies to allow the Railway server to insert data

-- Drop existing restrictive policies
DROP POLICY IF EXISTS "Enable read access for all users" ON classifications;
DROP POLICY IF EXISTS "Enable insert access for all users" ON classifications;
DROP POLICY IF EXISTS "Enable update access for all users" ON classifications;
DROP POLICY IF EXISTS "Enable delete access for all users" ON classifications;

-- Create new permissive policies
CREATE POLICY "Enable read access for all users" ON classifications FOR SELECT USING (true);
CREATE POLICY "Enable insert access for all users" ON classifications FOR INSERT WITH CHECK (true);
CREATE POLICY "Enable update access for all users" ON classifications FOR UPDATE USING (true) WITH CHECK (true);
CREATE POLICY "Enable delete access for all users" ON classifications FOR DELETE USING (true);

-- Also fix policies for business_risk_assessments table
DROP POLICY IF EXISTS "Enable read access for all users" ON business_risk_assessments;
DROP POLICY IF EXISTS "Enable insert access for all users" ON business_risk_assessments;
DROP POLICY IF EXISTS "Enable update access for all users" ON business_risk_assessments;
DROP POLICY IF EXISTS "Enable delete access for all users" ON business_risk_assessments;

CREATE POLICY "Enable read access for all users" ON business_risk_assessments FOR SELECT USING (true);
CREATE POLICY "Enable insert access for all users" ON business_risk_assessments FOR INSERT WITH CHECK (true);
CREATE POLICY "Enable update access for all users" ON business_risk_assessments FOR UPDATE USING (true) WITH CHECK (true);
CREATE POLICY "Enable delete access for all users" ON business_risk_assessments FOR DELETE USING (true);

-- Verify the policies are working
SELECT schemaname, tablename, policyname, permissive, roles, cmd, qual 
FROM pg_policies 
WHERE tablename IN ('classifications', 'business_risk_assessments')
ORDER BY tablename, policyname;
