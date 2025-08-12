#!/bin/bash

# KYB Platform - Environment Management Script
# This script helps manage local and Supabase environments

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

# Function to show usage
show_usage() {
    echo "KYB Platform - Environment Management"
    echo "====================================="
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  local     - Start local environment (PostgreSQL + Redis)"
    echo "  supabase  - Start Supabase environment"
    echo "  both      - Start both environments simultaneously"
    echo "  stop      - Stop all environments"
    echo "  status    - Show status of all environments"
    echo "  setup     - Set up Supabase configuration"
    echo "  logs      - Show logs for specified environment"
    echo "  test      - Test both environments"
    echo "  help      - Show this help message"
    echo
    echo "Examples:"
    echo "  $0 local                    # Start local environment"
    echo "  $0 supabase                 # Start Supabase environment"
    echo "  $0 both                     # Start both environments"
    echo "  $0 logs local               # Show local environment logs"
    echo "  $0 test                     # Test both environments"
    echo
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Function to start local environment
start_local() {
    print_status "Starting local environment..."
    
    if [ -f .env.local ]; then
        print_status "Using local environment configuration"
    else
        print_warning "No .env.local file found. Creating default local configuration..."
        create_local_env
    fi
    
    docker-compose up -d
    print_success "Local environment started"
    print_status "Access points:"
    echo "  - Application: http://localhost:8080"
    echo "  - Grafana: http://localhost:3000 (admin/admin)"
    echo "  - Prometheus: http://localhost:9090"
    echo "  - PostgreSQL: localhost:5433"
    echo "  - Redis: localhost:6379"
}

# Function to start Supabase environment
start_supabase() {
    print_status "Starting Supabase environment..."
    
    if [ ! -f .env ]; then
        print_error "No .env file found. Please run '$0 setup' first to configure Supabase."
        exit 1
    fi
    
    # Load environment variables
    source .env
    
    # Check if required Supabase variables are set
    if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_API_KEY" ]; then
        print_error "Supabase configuration incomplete. Please run '$0 setup' first."
        exit 1
    fi
    
    docker-compose -f docker-compose.supabase.yml up -d
    print_success "Supabase environment started"
    print_status "Access points:"
    echo "  - Application: http://localhost:8081"
    echo "  - Grafana: http://localhost:3001 (admin/admin)"
    echo "  - Prometheus: http://localhost:9091"
    echo "  - Supabase Dashboard: $SUPABASE_URL"
}

# Function to start both environments
start_both() {
    print_status "Starting both environments..."
    
    start_local
    echo
    start_supabase
    
    print_success "Both environments started"
    print_status "Access points:"
    echo "  Local Environment:"
    echo "    - Application: http://localhost:8080"
    echo "    - Grafana: http://localhost:3000"
    echo "    - Prometheus: http://localhost:9090"
    echo "  Supabase Environment:"
    echo "    - Application: http://localhost:8081"
    echo "    - Grafana: http://localhost:3001"
    echo "    - Prometheus: http://localhost:9091"
}

# Function to stop all environments
stop_all() {
    print_status "Stopping all environments..."
    
    docker-compose down 2>/dev/null || true
    docker-compose -f docker-compose.supabase.yml down 2>/dev/null || true
    
    print_success "All environments stopped"
}

# Function to show status
show_status() {
    print_status "Environment Status:"
    echo
    
    echo "Local Environment:"
    if docker-compose ps | grep -q "Up"; then
        print_success "  Status: Running"
        echo "  Application: http://localhost:8080"
        echo "  Grafana: http://localhost:3000"
        echo "  Prometheus: http://localhost:9090"
    else
        print_warning "  Status: Stopped"
    fi
    
    echo
    echo "Supabase Environment:"
    if docker-compose -f docker-compose.supabase.yml ps | grep -q "Up"; then
        print_success "  Status: Running"
        echo "  Application: http://localhost:8081"
        echo "  Grafana: http://localhost:3001"
        echo "  Prometheus: http://localhost:9091"
    else
        print_warning "  Status: Stopped"
    fi
    
    echo
    echo "Database Connections:"
    echo "  Local PostgreSQL: localhost:5433"
    echo "  Local Redis: localhost:6379"
    if [ -f .env ]; then
        source .env
        echo "  Supabase: $SUPABASE_URL"
    fi
}

# Function to show logs
show_logs() {
    local env=$1
    
    case $env in
        local)
            print_status "Showing local environment logs..."
            docker-compose logs -f
            ;;
        supabase)
            print_status "Showing Supabase environment logs..."
            docker-compose -f docker-compose.supabase.yml logs -f
            ;;
        *)
            print_error "Invalid environment. Use 'local' or 'supabase'"
            exit 1
            ;;
    esac
}

# Function to set up Supabase
setup_supabase() {
    print_status "Setting up Supabase configuration..."
    
    if [ -f "scripts/setup_supabase.sh" ]; then
        chmod +x scripts/setup_supabase.sh
        ./scripts/setup_supabase.sh
    else
        print_error "Supabase setup script not found"
        exit 1
    fi
}

# Function to test environments
test_environments() {
    print_status "Testing environments..."
    
    echo
    echo "Testing Local Environment:"
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        print_success "  Health check: PASSED"
    else
        print_error "  Health check: FAILED"
    fi
    
    echo
    echo "Testing Supabase Environment:"
    if curl -f http://localhost:8081/health > /dev/null 2>&1; then
        print_success "  Health check: PASSED"
    else
        print_error "  Health check: FAILED"
    fi
    
    echo
    print_status "Environment URLs:"
    echo "  Local: http://localhost:8080"
    echo "  Supabase: http://localhost:8081"
}

# Function to create local environment file
create_local_env() {
    cat > .env.local << EOF
# KYB Platform - Local Environment Configuration
ENV=development

# Provider Configuration
PROVIDER_DATABASE=postgres
PROVIDER_AUTH=local
PROVIDER_CACHE=redis
PROVIDER_STORAGE=local

# Database Configuration (Local PostgreSQL)
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USERNAME=kyb_user
DB_PASSWORD=kyb_password
DB_DATABASE=kyb_platform
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=2
DB_CONN_MAX_LIFETIME=5m
DB_AUTO_MIGRATE=true

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Authentication Configuration
JWT_SECRET=local_jwt_secret_change_in_production
JWT_EXPIRATION=24h
REFRESH_EXPIRATION=168h
MIN_PASSWORD_LENGTH=8
REQUIRE_UPPERCASE=true
REQUIRE_LOWERCASE=true
REQUIRE_NUMBERS=true
REQUIRE_SPECIAL=true
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION=15m

# Server Configuration
PORT=8080
HOST=0.0.0.0
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
IDLE_TIMEOUT=60s

# CORS Configuration
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER=100
RATE_LIMIT_WINDOW_SIZE=60
RATE_LIMIT_BURST_SIZE=200

# Observability Configuration
LOG_LEVEL=debug
LOG_FORMAT=text
METRICS_ENABLED=true
METRICS_PORT=9090
METRICS_PATH=/metrics
TRACING_ENABLED=true
TRACING_URL=http://localhost:14268/api/traces
HEALTH_CHECK_PATH=/health

# Classification Cache Configuration
CLASSIFICATION_CACHE_ENABLED=true
CLASSIFICATION_CACHE_TTL=10m
CLASSIFICATION_CACHE_MAX_ENTRIES=10000
CLASSIFICATION_CACHE_JANITOR_INTERVAL=1m
EOF
    
    print_success "Local environment file created: .env.local"
}

# Main execution
main() {
    check_docker
    
    case "${1:-help}" in
        local)
            start_local
            ;;
        supabase)
            start_supabase
            ;;
        both)
            start_both
            ;;
        stop)
            stop_all
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        setup)
            setup_supabase
            ;;
        test)
            test_environments
            ;;
        help|*)
            show_usage
            ;;
    esac
}

# Run main function
main "$@"
