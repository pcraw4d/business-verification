#!/bin/bash

# KYB Platform - Railway Startup Script
# This script handles database initialization and application startup

set -e

echo "🚀 KYB Platform Railway Startup"
echo "================================"

# Function to wait for database
wait_for_database() {
    echo "⏳ Waiting for database connection..."
    
    # Try to connect to database
    for i in {1..30}; do
        if pg_isready -h $DATABASE_HOST -p $DATABASE_PORT -U $DATABASE_USER > /dev/null 2>&1; then
            echo "✅ Database connection established"
            return 0
        fi
        echo "⏳ Attempt $i/30: Database not ready yet..."
        sleep 2
    done
    
    echo "❌ Database connection failed after 30 attempts"
    return 1
}

# Function to initialize database
initialize_database() {
    echo "🗄️ Initializing database..."
    
    # Run database migrations if needed
    if [ -f "./migrations/001_initial_schema.sql" ]; then
        echo "📝 Running database migrations..."
        psql $DATABASE_URL -f ./migrations/001_initial_schema.sql
    fi
    
    # Create required tables if they don't exist
    echo "📋 Creating database tables..."
    psql $DATABASE_URL << EOF
-- Create tables if they don't exist
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS business_classifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_name VARCHAR(255) NOT NULL,
    website_url VARCHAR(500),
    primary_industry JSONB,
    secondary_industries JSONB,
    confidence_score DECIMAL(3,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS risk_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID REFERENCES business_classifications(id),
    risk_factors JSONB,
    risk_score DECIMAL(3,2),
    risk_level VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_classifications_business_name ON business_classifications(business_name);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_business_id ON risk_assessments(business_id);
EOF
    
    echo "✅ Database initialization complete"
}

# Function to check required environment variables
check_environment() {
    echo "🔍 Checking environment variables..."
    
    required_vars=(
        "JWT_SECRET"
        "ENCRYPTION_KEY"
        "DATABASE_URL"
    )
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            echo "❌ Required environment variable $var is not set"
            exit 1
        fi
    done
    
    echo "✅ All required environment variables are set"
}

# Function to start the application
start_application() {
    echo "🚀 Starting KYB Platform application..."
    
    # Set proper permissions
    chmod +x ./kyb-platform
    
    # Start the application
    exec ./kyb-platform
}

# Main execution
main() {
    echo "📋 Starting KYB Platform Railway deployment..."
    
    # Check environment variables
    check_environment
    
    # Wait for database if DATABASE_URL is set
    if [ -n "$DATABASE_URL" ]; then
        # Extract database connection details
        export DATABASE_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\).*/\1/p')
        export DATABASE_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
        export DATABASE_USER=$(echo $DATABASE_URL | sed -n 's/.*:\/\/\([^:]*\):.*/\1/p')
        
        # Wait for database
        wait_for_database
        
        # Initialize database
        initialize_database
    else
        echo "⚠️ No DATABASE_URL provided, skipping database initialization"
    fi
    
    # Start the application
    start_application
}

# Run main function
main "$@"
