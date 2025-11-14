#!/bin/bash

# KYB Platform - Local Testing Environment Setup
# This script sets up the local environment for running merchant-details tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}[$(date +'%Y-%m-%d %H:%M:%S')] ${message}${NC}"
}

print_header() {
    echo ""
    print_status $BLUE "=========================================="
    print_status $BLUE "$1"
    print_status $BLUE "=========================================="
    echo ""
}

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    local missing=0
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_status $RED "✗ Go is not installed or not in PATH"
        missing=1
    else
        print_status $GREEN "✓ Go installed: $(go version | awk '{print $3}')"
    fi
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_status $RED "✗ Docker is not installed or not in PATH"
        missing=1
    else
        print_status $GREEN "✓ Docker installed: $(docker --version | awk '{print $3}' | tr -d ',')"
    fi
    
    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_status $RED "✗ Docker Compose is not installed"
        missing=1
    else
        print_status $GREEN "✓ Docker Compose installed"
    fi
    
    # Check if psql is installed (optional, for manual DB access)
    if command -v psql &> /dev/null; then
        print_status $GREEN "✓ psql installed"
    else
        print_status $YELLOW "⚠ psql not installed (optional, for manual database access)"
    fi
    
    if [ $missing -eq 1 ]; then
        print_status $RED "Please install missing prerequisites before continuing"
        exit 1
    fi
    
    print_status $GREEN "✓ All prerequisites met"
}

# Function to create .env.test file
create_test_env() {
    print_header "Creating Test Environment File"
    
    local env_file="$PROJECT_ROOT/.env.test"
    
    if [ -f "$env_file" ]; then
        print_status $YELLOW ".env.test already exists, backing up to .env.test.backup"
        cp "$env_file" "$env_file.backup"
    fi
    
    cat > "$env_file" << 'EOF'
# Local Testing Environment Configuration
# This file is for local testing only

# Test Configuration
TEST_BASE_URL=http://localhost:8080
TEST_AUTH_TOKEN=test-token-local

# Database Configuration (local PostgreSQL)
DATABASE_URL=postgres://kyb_test:kyb_test_password@localhost:5433/kyb_test?sslmode=disable

# Redis Configuration (local Redis)
REDIS_URL=redis://localhost:6379/0

# Application Configuration
ENVIRONMENT=local
LOG_LEVEL=debug
PORT=8080

# CORS Configuration
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*

# Rate Limiting (disabled for testing)
RATE_LIMIT_ENABLED=false

# Feature Flags
FEATURE_BUSINESS_CLASSIFICATION=true
FEATURE_RISK_ASSESSMENT=true
FEATURE_COMPLIANCE_FRAMEWORK=true
EOF

    print_status $GREEN "✓ Created .env.test file"
    print_status $YELLOW "⚠ Please review and update .env.test if needed"
}

# Function to start Docker services
start_docker_services() {
    print_header "Starting Docker Services"
    
    cd "$PROJECT_ROOT"
    
    # Check if Docker is running
    if ! docker info > /dev/null 2>&1; then
        print_status $RED "✗ Docker daemon is not running"
        print_status $YELLOW "Please start Docker Desktop or Docker daemon and run this script again"
        exit 1
    fi
    
    # Create a minimal docker-compose file for testing if it doesn't exist
    local compose_file="$PROJECT_ROOT/docker-compose.test.yml"
    
    # Check if existing file has different configuration
    if [ -f "$compose_file" ]; then
        print_status $YELLOW "docker-compose.test.yml already exists"
        print_status $BLUE "Backing up to docker-compose.test.yml.backup"
        cp "$compose_file" "$compose_file.backup"
    fi
    
    print_status $BLUE "Creating/updating docker-compose.test.yml..."
    cat > "$compose_file" << 'EOF'
services:
  postgres:
    image: postgres:15
    container_name: kyb-test-postgres
    environment:
      POSTGRES_DB: kyb_test
      POSTGRES_USER: kyb_test
      POSTGRES_PASSWORD: kyb_test_password
    ports:
      - "5433:5432"
    volumes:
      - postgres-test-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U kyb_test -d kyb_test"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - kyb-test-network

  redis:
    image: redis:7-alpine
    container_name: kyb-test-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-test-data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    networks:
      - kyb-test-network

volumes:
  postgres-test-data:
  redis-test-data:

networks:
  kyb-test-network:
    driver: bridge
EOF
    print_status $GREEN "✓ Created/updated docker-compose.test.yml"
    
    # Start services
    print_status $BLUE "Starting PostgreSQL and Redis..."
    docker-compose -f "$compose_file" up -d
    
    # Wait for services to be healthy
    print_status $BLUE "Waiting for services to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if docker-compose -f "$compose_file" ps | grep -q "healthy"; then
            print_status $GREEN "✓ Services are healthy"
            break
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    if [ $attempt -eq $max_attempts ]; then
        print_status $YELLOW "⚠ Services may still be starting. Check with: docker-compose -f docker-compose.test.yml ps"
    fi
    
    print_status $GREEN "✓ Docker services started"
}

# Function to run database migrations
run_migrations() {
    print_header "Running Database Migrations"
    
    cd "$PROJECT_ROOT"
    
    # Load environment variables
    if [ -f ".env.test" ]; then
        export $(cat .env.test | grep -v '^#' | xargs)
    fi
    
    # Wait a bit for database to be fully ready
    print_status $BLUE "Waiting for database to be ready..."
    sleep 3
    
    # Check if DATABASE_URL is set
    if [ -z "$DATABASE_URL" ]; then
        print_status $RED "✗ DATABASE_URL not set. Please check .env.test"
        exit 1
    fi
    
    # Run migrations using psql
    print_status $BLUE "Running migrations..."
    
    local migration_dir="$PROJECT_ROOT/internal/database/migrations"
    local migrations=(
        "001_initial_schema.sql"
        "002_rbac_schema.sql"
        "003_performance_indexes.sql"
        "004_enhanced_classification.sql"
        "005_merchant_portfolio_schema.sql"
        "007_foreign_key_relationships.sql"
        "008_additional_performance_indexes.sql"
        "008_enhance_merchants_table.sql"
        "008_unified_compliance_schema.sql"
        "008_user_table_consolidation.sql"
        "009_remove_redundant_profiles_table.sql"
        "009_unified_audit_schema.sql"
        "010_add_async_risk_assessment_columns.sql"
        "011_add_updated_at_to_risk_assessments.sql"
    )
    
    local success=0
    for migration in "${migrations[@]}"; do
        local migration_file="$migration_dir/$migration"
        if [ -f "$migration_file" ]; then
            print_status $BLUE "Running migration: $migration"
            if psql "$DATABASE_URL" -f "$migration_file" > /dev/null 2>&1; then
                print_status $GREEN "✓ $migration"
                success=1
            else
                # Check if error is just "already exists" (migration already run)
                if psql "$DATABASE_URL" -f "$migration_file" 2>&1 | grep -q "already exists\|duplicate"; then
                    print_status $YELLOW "⚠ $migration (already applied)"
                    success=1
                else
                    print_status $YELLOW "⚠ $migration (may have errors, but continuing)"
                fi
            fi
        else
            print_status $YELLOW "⚠ Migration file not found: $migration"
        fi
    done
    
    if [ $success -eq 1 ]; then
        print_status $GREEN "✓ Migrations completed"
    else
        print_status $YELLOW "⚠ Some migrations may have failed. Check database manually."
    fi
}

# Function to verify setup
verify_setup() {
    print_header "Verifying Setup"
    
    # Check database connection
    print_status $BLUE "Testing database connection..."
    if [ -f ".env.test" ]; then
        export $(cat .env.test | grep -v '^#' | xargs)
    fi
    
    if [ -n "$DATABASE_URL" ]; then
        if psql "$DATABASE_URL" -c "SELECT version();" > /dev/null 2>&1; then
            print_status $GREEN "✓ Database connection successful"
        else
            print_status $RED "✗ Database connection failed"
            return 1
        fi
    fi
    
    # Check Redis connection
    print_status $BLUE "Testing Redis connection..."
    if docker exec kyb-test-redis redis-cli ping > /dev/null 2>&1; then
        print_status $GREEN "✓ Redis connection successful"
    else
        print_status $YELLOW "⚠ Redis connection test failed (may still be starting)"
    fi
    
    # Check if test files exist
    print_status $BLUE "Checking test files..."
    if [ -f "$PROJECT_ROOT/test/integration/risk_assessment_integration_test.go" ]; then
        print_status $GREEN "✓ Integration tests found"
    else
        print_status $YELLOW "⚠ Integration tests not found"
    fi
    
    if [ -f "$PROJECT_ROOT/test/e2e/merchant_details_e2e_test.go" ]; then
        print_status $GREEN "✓ E2E tests found"
    else
        print_status $YELLOW "⚠ E2E tests not found"
    fi
    
    print_status $GREEN "✓ Setup verification complete"
}

# Function to show usage instructions
show_usage() {
    print_header "Local Testing Environment Ready!"
    
    echo ""
    print_status $GREEN "Next Steps:"
    echo ""
    echo "1. Load test environment variables:"
    echo "   ${BLUE}source .env.test${NC}"
    echo ""
    echo "2. Or export them manually:"
    echo "   ${BLUE}export TEST_BASE_URL=http://localhost:8080${NC}"
    echo "   ${BLUE}export TEST_AUTH_TOKEN=test-token-local${NC}"
    echo "   ${BLUE}export DATABASE_URL=postgres://kyb_test:kyb_test_password@localhost:5433/kyb_test?sslmode=disable${NC}"
    echo ""
    echo "3. Run integration tests:"
    echo "   ${BLUE}go test -v -tags=integration ./test/integration/risk_assessment_integration_test.go${NC}"
    echo ""
    echo "4. Run E2E tests:"
    echo "   ${BLUE}go test -v -tags=e2e ./test/e2e/merchant_details_e2e_test.go${NC}"
    echo ""
    echo "5. Stop Docker services when done:"
    echo "   ${BLUE}docker-compose -f docker-compose.test.yml down${NC}"
    echo ""
    echo "6. Stop and remove all data:"
    echo "   ${BLUE}docker-compose -f docker-compose.test.yml down -v${NC}"
    echo ""
}

# Main execution
main() {
    print_header "KYB Platform - Local Testing Environment Setup"
    
    check_prerequisites
    create_test_env
    start_docker_services
    run_migrations
    verify_setup
    show_usage
    
    print_status $GREEN "✓ Local testing environment setup complete!"
}

# Run main function
main

