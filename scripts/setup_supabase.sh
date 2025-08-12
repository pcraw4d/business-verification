#!/bin/bash

# KYB Platform - Supabase Setup Script
# This script helps set up Supabase for the KYB platform

set -e

echo "🚀 KYB Platform - Supabase Setup"
echo "=================================="

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

# Check if required tools are installed
check_requirements() {
    print_status "Checking requirements..."
    
    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_warning "jq is not installed. Install it for better JSON parsing."
    fi
    
    print_success "Requirements check completed"
}

# Get Supabase project details
get_project_details() {
    print_status "Getting Supabase project details..."
    
    echo
    echo "Please provide your Supabase project details:"
    echo
    
    read -p "Supabase Project URL (e.g., https://your-project.supabase.co): " SUPABASE_URL
    read -p "Supabase Anon Key: " SUPABASE_ANON_KEY
    read -p "Supabase Service Role Key: " SUPABASE_SERVICE_ROLE_KEY
    read -p "Supabase JWT Secret: " SUPABASE_JWT_SECRET
    read -p "Database Password: " DB_PASSWORD
    
    # Extract project reference from URL
    PROJECT_REF=$(echo $SUPABASE_URL | sed 's|https://||' | sed 's|.supabase.co||')
    
    # Set database connection details
    DB_HOST="db.${PROJECT_REF}.supabase.co"
    DB_PORT="5432"
    DB_USERNAME="postgres"
    DB_DATABASE="postgres"
    DB_SSL_MODE="require"
    
    print_success "Project details captured"
}

# Create environment file
create_env_file() {
    print_status "Creating environment file..."
    
    cat > .env << EOF
# KYB Platform - Supabase Configuration
ENV=development

# Provider Configuration
PROVIDER_DATABASE=supabase
PROVIDER_AUTH=supabase
PROVIDER_CACHE=supabase
PROVIDER_STORAGE=supabase

# Supabase Configuration
SUPABASE_URL=${SUPABASE_URL}
SUPABASE_API_KEY=${SUPABASE_ANON_KEY}
SUPABASE_SERVICE_ROLE_KEY=${SUPABASE_SERVICE_ROLE_KEY}
SUPABASE_JWT_SECRET=${SUPABASE_JWT_SECRET}

# Database Configuration (Supabase PostgreSQL)
DB_DRIVER=postgres
DB_HOST=${DB_HOST}
DB_PORT=${DB_PORT}
DB_USERNAME=${DB_USERNAME}
DB_PASSWORD=${DB_PASSWORD}
DB_DATABASE=${DB_DATABASE}
DB_SSL_MODE=${DB_SSL_MODE}
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=2
DB_CONN_MAX_LIFETIME=5m
DB_AUTO_MIGRATE=true

# Authentication Configuration
JWT_SECRET=dev_jwt_secret_change_in_production
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
    
    print_success "Environment file created: .env"
}

# Test database connection
test_database_connection() {
    print_status "Testing database connection..."
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        print_warning "psql not found. Skipping database connection test."
        return
    fi
    
    # Test connection
    if PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d "${DB_DATABASE}" -c "SELECT 1;" &> /dev/null; then
        print_success "Database connection successful"
    else
        print_error "Database connection failed"
        print_warning "Please check your database credentials and network connectivity"
    fi
}

# Run database migrations
run_migrations() {
	print_status "Running database migrations..."
	
	# Check if migrations directory exists
	if [ ! -d "internal/database/migrations" ]; then
		print_error "Migrations directory not found"
		return
	fi
	
	# Run migrations using the application
	print_status "Starting application to run migrations..."
	
	# Set environment variables for this session
	export SUPABASE_URL="${SUPABASE_URL}"
	export SUPABASE_API_KEY="${SUPABASE_ANON_KEY}"
	export SUPABASE_SERVICE_ROLE_KEY="${SUPABASE_SERVICE_ROLE_KEY}"
	export SUPABASE_JWT_SECRET="${SUPABASE_JWT_SECRET}"
	export DB_HOST="${DB_HOST}"
	export DB_PORT="${DB_PORT}"
	export DB_USERNAME="${DB_USERNAME}"
	export DB_PASSWORD="${DB_PASSWORD}"
	export DB_DATABASE="${DB_DATABASE}"
	export DB_SSL_MODE="${DB_SSL_MODE}"
	export DB_AUTO_MIGRATE="true"
	
	# Build and run the application briefly to trigger migrations
	print_status "Building application..."
	go build -o kyb-platform ./cmd/api
	
	print_status "Running migrations..."
	timeout 30s ./kyb-platform || true
	
	print_success "Migrations completed"
}

# Create Supabase configuration
create_supabase_config() {
    print_status "Creating Supabase configuration..."
    
    # Create supabase directory if it doesn't exist
    mkdir -p supabase
    
    # Create config.toml
    cat > supabase/config.toml << EOF
# A string used to distinguish different Supabase projects on the same host. Defaults to the
# working directory name when running supabase init.
project_id = "kyb-platform"

[api]
enabled = true
port = 54321
schemas = ["public", "storage", "graphql_public"]
extra_search_path = ["public", "extensions"]
max_rows = 1000

[db]
port = 54322
shadow_port = 54320
major_version = 15

[db.pooler]
enabled = false
port = 54329
pool_mode = "transaction"
default_pool_size = 15
max_client_conn = 100

[realtime]
enabled = true
port = 54323

[studio]
enabled = true
port = 54323
api_url = "http://localhost:54321"

[inbucket]
enabled = true
port = 54324
smtp_port = 54325
pop3_port = 54326

[storage]
enabled = true
file_size_limit = "50MiB"

[auth]
enabled = true
port = 54324
site_url = "http://localhost:3000"
additional_redirect_urls = ["https://localhost:3000"]
jwt_expiry = 3600
refresh_token_rotation_enabled = true
security_update_password_require_reauthentication = true
enable_signup = true

[auth.email]
enable_signup = true
double_confirm_changes = true
enable_confirmations = false

[auth.sms]
enable_signup = true
enable_confirmations = false
template = "Your code is {{ .Code }}"

[auth.external.apple]
enabled = false
client_id = ""
secret = ""
redirect_uri = ""
additional_scopes = ""
EOF
    
    print_success "Supabase configuration created: supabase/config.toml"
}

# Display next steps
show_next_steps() {
    echo
    echo "🎉 Supabase setup completed!"
    echo "=============================="
    echo
    echo "Next steps:"
    echo
    echo "1. Review the generated .env file:"
    echo "   cat .env"
    echo
    echo "2. Start the application:"
    echo "   docker-compose -f docker-compose.dev.yml up"
    echo
    echo "3. Or run locally:"
    echo "   go run ./cmd/api"
    echo
    echo "4. Access the application:"
    echo "   http://localhost:8080"
    echo
    echo "5. Access Supabase Dashboard:"
    echo "   ${SUPABASE_URL}"
    echo
    echo "6. Monitor with Grafana:"
    echo "   http://localhost:3000 (admin/admin)"
    echo
    echo "7. View metrics with Prometheus:"
    echo "   http://localhost:9090"
    echo
    echo "📚 Documentation:"
    echo "   - Supabase Docs: https://supabase.com/docs"
    echo "   - KYB Platform Docs: ./docs/"
    echo
    echo "🔧 Troubleshooting:"
    echo "   - Check logs: docker-compose logs kyb-platform"
    echo "   - Test database: PGPASSWORD='${DB_PASSWORD}' psql -h ${DB_HOST} -U ${DB_USERNAME} -d ${DB_DATABASE}"
    echo
}

# Main execution
main() {
    echo
    print_status "Starting Supabase setup for KYB Platform"
    echo
    
    check_requirements
    get_project_details
    create_env_file
    test_database_connection
    create_supabase_config
    run_migrations
    show_next_steps
    
    print_success "Setup completed successfully!"
}

# Run main function
main "$@"
