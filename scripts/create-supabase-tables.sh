#!/bin/bash

# KYB Platform - Create Supabase Tables Script
# This script creates the necessary tables in Supabase using the REST API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required environment variables are set
if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    print_error "Required Supabase environment variables not set"
    print_error "Need: SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY"
    exit 1
fi

print_status "🚀 Creating Supabase tables for KYB Platform"
print_status "============================================="

# Create portfolio types table
print_status "Creating portfolio_types table..."
curl -X POST \
    -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Content-Type: application/json" \
    -H "Prefer: return=minimal" \
    -d '{
        "query": "CREATE TABLE IF NOT EXISTS portfolio_types (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), type VARCHAR(50) UNIQUE NOT NULL CHECK (type IN ('"'"'onboarded'"'"', '"'"'deactivated'"'"', '"'"'prospective'"'"', '"'"'pending'"'"')), description TEXT, display_order INTEGER NOT NULL DEFAULT 0, is_active BOOLEAN NOT NULL DEFAULT true, created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP);"
    }' \
    "$SUPABASE_URL/rest/v1/rpc/exec" \
    || print_warning "Failed to create portfolio_types table"

# Create risk levels table
print_status "Creating risk_levels table..."
curl -X POST \
    -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Content-Type: application/json" \
    -H "Prefer: return=minimal" \
    -d '{
        "query": "CREATE TABLE IF NOT EXISTS risk_levels (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), level VARCHAR(50) UNIQUE NOT NULL CHECK (level IN ('"'"'high'"'"', '"'"'medium'"'"', '"'"'low'"'"')), description TEXT, numeric_value INTEGER NOT NULL, color_code VARCHAR(7), display_order INTEGER NOT NULL DEFAULT 0, is_active BOOLEAN NOT NULL DEFAULT true, created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP);"
    }' \
    "$SUPABASE_URL/rest/v1/rpc/exec" \
    || print_warning "Failed to create risk_levels table"

# Create merchants table
print_status "Creating merchants table..."
curl -X POST \
    -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Content-Type: application/json" \
    -H "Prefer: return=minimal" \
    -d '{
        "query": "CREATE TABLE IF NOT EXISTS merchants (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), name VARCHAR(255) NOT NULL, legal_name VARCHAR(255) NOT NULL, registration_number VARCHAR(100) UNIQUE NOT NULL, tax_id VARCHAR(100), industry VARCHAR(100), industry_code VARCHAR(20), business_type VARCHAR(50), founded_date DATE, employee_count INTEGER, annual_revenue DECIMAL(15,2), address_street1 VARCHAR(255), address_street2 VARCHAR(255), address_city VARCHAR(100), address_state VARCHAR(100), address_postal_code VARCHAR(20), address_country VARCHAR(100), address_country_code VARCHAR(10), contact_phone VARCHAR(50), contact_email VARCHAR(255), contact_website VARCHAR(255), contact_primary_contact VARCHAR(255), portfolio_type_id UUID, risk_level_id UUID, compliance_status VARCHAR(50) NOT NULL DEFAULT '"'"'pending'"'"', status VARCHAR(50) NOT NULL DEFAULT '"'"'active'"'"', created_by UUID, created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP);"
    }' \
    "$SUPABASE_URL/rest/v1/rpc/exec" \
    || print_warning "Failed to create merchants table"

print_success "Table creation completed!"
print_status "Note: If table creation failed, you may need to run the SQL migration manually in the Supabase SQL Editor"
print_status "SQL file available at: supabase-migration.sql"
