-- Migration: 004_supabase_optimizations.sql
-- Description: Supabase-specific optimizations and Row Level Security (RLS) policies
-- Date: 2024-01-01

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- Create cache table for Supabase caching
CREATE TABLE IF NOT EXISTS cache (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for cache table
CREATE INDEX IF NOT EXISTS idx_cache_expires_at ON cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_cache_created_at ON cache(created_at);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_businesses_created_by ON businesses(created_by);
CREATE INDEX IF NOT EXISTS idx_classifications_business_id ON classifications(business_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_business_id ON risk_assessments(business_id);
CREATE INDEX IF NOT EXISTS idx_compliance_status_business_id ON compliance_status(business_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_external_service_calls_user_id ON external_service_calls(user_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_user_id ON webhooks(user_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_webhook_id ON webhook_events(webhook_id);

-- Enable Row Level Security (RLS) on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE businesses ENABLE ROW LEVEL SECURITY;
ALTER TABLE classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE risk_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE compliance_status ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE external_service_calls ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE email_verification_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE password_reset_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE token_blacklist ENABLE ROW LEVEL SECURITY;

-- RLS Policies for users table
CREATE POLICY "Users can view their own profile" ON users
    FOR SELECT USING (auth.uid()::text = id);

CREATE POLICY "Users can update their own profile" ON users
    FOR UPDATE USING (auth.uid()::text = id);

CREATE POLICY "Admins can view all users" ON users
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM role_assignments ra 
            WHERE ra.user_id = auth.uid()::text 
            AND ra.role = 'admin' 
            AND ra.is_active = true
        )
    );

-- RLS Policies for businesses table
CREATE POLICY "Users can view their own businesses" ON businesses
    FOR SELECT USING (created_by = auth.uid()::text);

CREATE POLICY "Users can create their own businesses" ON businesses
    FOR INSERT WITH CHECK (created_by = auth.uid()::text);

CREATE POLICY "Users can update their own businesses" ON businesses
    FOR UPDATE USING (created_by = auth.uid()::text);

CREATE POLICY "Users can delete their own businesses" ON businesses
    FOR DELETE USING (created_by = auth.uid()::text);

CREATE POLICY "Admins can view all businesses" ON businesses
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM role_assignments ra 
            WHERE ra.user_id = auth.uid()::text 
            AND ra.role = 'admin' 
            AND ra.is_active = true
        )
    );

-- RLS Policies for classifications table
CREATE POLICY "Users can view classifications for their businesses" ON classifications
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM businesses b 
            WHERE b.id = business_id 
            AND b.created_by = auth.uid()::text
        )
    );

CREATE POLICY "Users can create classifications for their businesses" ON classifications
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM businesses b 
            WHERE b.id = business_id 
            AND b.created_by = auth.uid()::text
        )
    );

CREATE POLICY "Admins can view all classifications" ON classifications
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM role_assignments ra 
            WHERE ra.user_id = auth.uid()::text 
            AND ra.role = 'admin' 
            AND ra.is_active = true
        )
    );

-- RLS Policies for risk_assessments table
CREATE POLICY "Users can view risk assessments for their businesses" ON risk_assessments
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM businesses b 
            WHERE b.id = business_id 
            AND b.created_by = auth.uid()::text
        )
    );

CREATE POLICY "Users can create risk assessments for their businesses" ON risk_assessments
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM businesses b 
            WHERE b.id = business_id 
            AND b.created_by = auth.uid()::text
        )
    );

CREATE POLICY "Admins can view all risk assessments" ON risk_assessments
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM role_assignments ra 
            WHERE ra.user_id = auth.uid()::text 
            AND ra.role = 'admin' 
            AND ra.is_active = true
        )
    );

-- RLS Policies for compliance_status table
CREATE POLICY "Users can view compliance status for their businesses" ON compliance_status
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM businesses b 
            WHERE b.id = business_id 
            AND b.created_by = auth.uid()::text
        )
    );

CREATE POLICY "Users can create compliance status for their businesses" ON compliance_status
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM businesses b 
            WHERE b.id = business_id 
            AND b.created_by = auth.uid()::text
        )
    );

CREATE POLICY "Admins can view all compliance status" ON compliance_status
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM role_assignments ra 
            WHERE ra.user_id = auth.uid()::text 
            AND ra.role = 'admin' 
            AND ra.is_active = true
        )
    );

-- RLS Policies for audit_logs table
CREATE POLICY "Users can view their own audit logs" ON audit_logs
    FOR SELECT USING (user_id = auth.uid()::text);

CREATE POLICY "Users can create their own audit logs" ON audit_logs
    FOR INSERT WITH CHECK (user_id = auth.uid()::text);

CREATE POLICY "Admins can view all audit logs" ON audit_logs
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM role_assignments ra 
            WHERE ra.user_id = auth.uid()::text 
            AND ra.role = 'admin' 
            AND ra.is_active = true
        )
    );

-- RLS Policies for external_service_calls table
CREATE POLICY "Users can view their own service calls" ON external_service_calls
    FOR SELECT USING (user_id = auth.uid()::text);

CREATE POLICY "Users can create their own service calls" ON external_service_calls
    FOR INSERT WITH CHECK (user_id = auth.uid()::text);

CREATE POLICY "Admins can view all service calls" ON external_service_calls
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM role_assignments ra 
            WHERE ra.user_id = auth.uid()::text 
            AND ra.role = 'admin' 
            AND ra.is_active = true
        )
    );

-- RLS Policies for webhooks table
CREATE POLICY "Users can view their own webhooks" ON webhooks
    FOR SELECT USING (user_id = auth.uid()::text);

CREATE POLICY "Users can create their own webhooks" ON webhooks
    FOR INSERT WITH CHECK (user_id = auth.uid()::text);

CREATE POLICY "Users can update their own webhooks" ON webhooks
    FOR UPDATE USING (user_id = auth.uid()::text);

CREATE POLICY "Users can delete their own webhooks" ON webhooks
    FOR DELETE USING (user_id = auth.uid()::text);

-- RLS Policies for webhook_events table
CREATE POLICY "Users can view events for their webhooks" ON webhook_events
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM webhooks w 
            WHERE w.id = webhook_id 
            AND w.user_id = auth.uid()::text
        )
    );

CREATE POLICY "Users can create events for their webhooks" ON webhook_events
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM webhooks w 
            WHERE w.id = webhook_id 
            AND w.user_id = auth.uid()::text
        )
    );

-- RLS Policies for api_keys table
CREATE POLICY "Users can view their own API keys" ON api_keys
    FOR SELECT USING (user_id = auth.uid()::text);

CREATE POLICY "Users can create their own API keys" ON api_keys
    FOR INSERT WITH CHECK (user_id = auth.uid()::text);

CREATE POLICY "Users can update their own API keys" ON api_keys
    FOR UPDATE USING (user_id = auth.uid()::text);

CREATE POLICY "Users can delete their own API keys" ON api_keys
    FOR DELETE USING (user_id = auth.uid()::text);

-- RLS Policies for role_assignments table
CREATE POLICY "Users can view their own role assignments" ON role_assignments
    FOR SELECT USING (user_id = auth.uid()::text);

CREATE POLICY "Admins can manage all role assignments" ON role_assignments
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM role_assignments ra 
            WHERE ra.user_id = auth.uid()::text 
            AND ra.role = 'admin' 
            AND ra.is_active = true
        )
    );

-- RLS Policies for email_verification_tokens table
CREATE POLICY "Users can view their own verification tokens" ON email_verification_tokens
    FOR SELECT USING (user_id = auth.uid()::text);

CREATE POLICY "Users can create their own verification tokens" ON email_verification_tokens
    FOR INSERT WITH CHECK (user_id = auth.uid()::text);

-- RLS Policies for password_reset_tokens table
CREATE POLICY "Users can view their own reset tokens" ON password_reset_tokens
    FOR SELECT USING (user_id = auth.uid()::text);

CREATE POLICY "Users can create their own reset tokens" ON password_reset_tokens
    FOR INSERT WITH CHECK (user_id = auth.uid()::text);

-- RLS Policies for token_blacklist table
CREATE POLICY "Users can view their own blacklisted tokens" ON token_blacklist
    FOR SELECT USING (user_id = auth.uid()::text);

CREATE POLICY "Users can create their own blacklisted tokens" ON token_blacklist
    FOR INSERT WITH CHECK (user_id = auth.uid()::text);

-- RLS Policies for cache table (read-only for all authenticated users)
CREATE POLICY "Authenticated users can read cache" ON cache
    FOR SELECT USING (auth.role() = 'authenticated');

CREATE POLICY "Authenticated users can write cache" ON cache
    FOR INSERT WITH CHECK (auth.role() = 'authenticated');

CREATE POLICY "Authenticated users can update cache" ON cache
    FOR UPDATE USING (auth.role() = 'authenticated');

CREATE POLICY "Authenticated users can delete cache" ON cache
    FOR DELETE USING (auth.role() = 'authenticated');

-- Create function to automatically set created_by and updated_at
CREATE OR REPLACE FUNCTION set_created_by()
RETURNS TRIGGER AS $$
BEGIN
    NEW.created_by = auth.uid()::text;
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create function to automatically set updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic field updates
CREATE TRIGGER set_businesses_created_by
    BEFORE INSERT ON businesses
    FOR EACH ROW
    EXECUTE FUNCTION set_created_by();

CREATE TRIGGER set_businesses_updated_at
    BEFORE UPDATE ON businesses
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_audit_logs_created_by
    BEFORE INSERT ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION set_created_by();

CREATE TRIGGER set_external_service_calls_created_by
    BEFORE INSERT ON external_service_calls
    FOR EACH ROW
    EXECUTE FUNCTION set_created_by();

CREATE TRIGGER set_webhooks_created_by
    BEFORE INSERT ON webhooks
    FOR EACH ROW
    EXECUTE FUNCTION set_created_by();

CREATE TRIGGER set_webhooks_updated_at
    BEFORE UPDATE ON webhooks
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_api_keys_created_by
    BEFORE INSERT ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION set_created_by();

CREATE TRIGGER set_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Create function to clean up expired cache entries
CREATE OR REPLACE FUNCTION cleanup_expired_cache()
RETURNS void AS $$
BEGIN
    DELETE FROM cache WHERE expires_at <= NOW();
END;
$$ LANGUAGE plpgsql;

-- Create a scheduled job to clean up expired cache entries (runs every hour)
SELECT cron.schedule(
    'cleanup-expired-cache',
    '0 * * * *', -- Every hour
    'SELECT cleanup_expired_cache();'
);

-- Create function to get user's businesses with classification count
CREATE OR REPLACE FUNCTION get_user_businesses_with_stats(user_uuid text)
RETURNS TABLE (
    business_id text,
    business_name text,
    classification_count bigint,
    risk_assessment_count bigint,
    compliance_check_count bigint,
    created_at timestamp with time zone
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        b.id,
        b.name,
        COUNT(DISTINCT c.id) as classification_count,
        COUNT(DISTINCT ra.id) as risk_assessment_count,
        COUNT(DISTINCT cs.id) as compliance_check_count,
        b.created_at
    FROM businesses b
    LEFT JOIN classifications c ON b.id = c.business_id
    LEFT JOIN risk_assessments ra ON b.id = ra.business_id
    LEFT JOIN compliance_status cs ON b.id = cs.business_id
    WHERE b.created_by = user_uuid
    GROUP BY b.id, b.name, b.created_at
    ORDER BY b.created_at DESC;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Grant execute permission on the function
GRANT EXECUTE ON FUNCTION get_user_businesses_with_stats(text) TO authenticated;
