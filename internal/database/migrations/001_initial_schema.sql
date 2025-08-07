-- Initial database schema for KYB platform
-- Migration: 001_initial_schema.sql

-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    company VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMP,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for fast lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);

-- Businesses table
CREATE TABLE businesses (
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
    
    -- Address fields
    address_street1 VARCHAR(255),
    address_street2 VARCHAR(255),
    address_city VARCHAR(100),
    address_state VARCHAR(100),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100),
    address_country_code VARCHAR(10),
    
    -- Contact info fields
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),
    contact_website VARCHAR(255),
    contact_primary_contact VARCHAR(255),
    
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    risk_level VARCHAR(50) DEFAULT 'unknown',
    compliance_status VARCHAR(50) DEFAULT 'pending',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for businesses table
CREATE INDEX idx_businesses_registration_number ON businesses(registration_number);
CREATE INDEX idx_businesses_tax_id ON businesses(tax_id);
CREATE INDEX idx_businesses_industry ON businesses(industry);
CREATE INDEX idx_businesses_status ON businesses(status);
CREATE INDEX idx_businesses_risk_level ON businesses(risk_level);
CREATE INDEX idx_businesses_created_by ON businesses(created_by);

-- Business classifications table
CREATE TABLE business_classifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    industry_code VARCHAR(20) NOT NULL,
    industry_name VARCHAR(255) NOT NULL,
    confidence_score DECIMAL(5,4) NOT NULL,
    classification_method VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL,
    raw_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for business classifications
CREATE INDEX idx_business_classifications_business_id ON business_classifications(business_id);
CREATE INDEX idx_business_classifications_industry_code ON business_classifications(industry_code);
CREATE INDEX idx_business_classifications_confidence_score ON business_classifications(confidence_score);

-- Risk assessments table
CREATE TABLE risk_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    risk_level VARCHAR(50) NOT NULL,
    risk_score DECIMAL(5,4) NOT NULL,
    risk_factors JSONB,
    assessment_method VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL,
    raw_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for risk assessments
CREATE INDEX idx_risk_assessments_business_id ON risk_assessments(business_id);
CREATE INDEX idx_risk_assessments_risk_level ON risk_assessments(risk_level);
CREATE INDEX idx_risk_assessments_risk_score ON risk_assessments(risk_score);

-- Compliance checks table
CREATE TABLE compliance_checks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    compliance_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    score DECIMAL(5,4) NOT NULL,
    requirements JSONB,
    check_method VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL,
    raw_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for compliance checks
CREATE INDEX idx_compliance_checks_business_id ON compliance_checks(business_id);
CREATE INDEX idx_compliance_checks_compliance_type ON compliance_checks(compliance_type);
CREATE INDEX idx_compliance_checks_status ON compliance_checks(status);

-- API keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    permissions JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for API keys
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_status ON api_keys(status);

-- Audit logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for audit logs
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id);

-- External service calls table
CREATE TABLE external_service_calls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    service_name VARCHAR(100) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    method VARCHAR(10) NOT NULL,
    request_data JSONB,
    response_data JSONB,
    status_code INTEGER NOT NULL,
    duration_ms BIGINT NOT NULL,
    error TEXT,
    request_id VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for external service calls
CREATE INDEX idx_external_service_calls_user_id ON external_service_calls(user_id);
CREATE INDEX idx_external_service_calls_service_name ON external_service_calls(service_name);
CREATE INDEX idx_external_service_calls_status_code ON external_service_calls(status_code);
CREATE INDEX idx_external_service_calls_created_at ON external_service_calls(created_at);
CREATE INDEX idx_external_service_calls_request_id ON external_service_calls(request_id);

-- Webhooks table
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    events JSONB NOT NULL,
    secret VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    last_triggered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for webhooks
CREATE INDEX idx_webhooks_user_id ON webhooks(user_id);
CREATE INDEX idx_webhooks_status ON webhooks(status);

-- Webhook events table
CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    response_code INTEGER,
    response_body TEXT,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for webhook events
CREATE INDEX idx_webhook_events_webhook_id ON webhook_events(webhook_id);
CREATE INDEX idx_webhook_events_status ON webhook_events(status);
CREATE INDEX idx_webhook_events_created_at ON webhook_events(created_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_businesses_updated_at BEFORE UPDATE ON businesses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
