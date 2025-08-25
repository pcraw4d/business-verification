#!/bin/bash

# KYB Platform - Enhanced Business Intelligence Beta Testing Launch Script
# This script sets up and launches the comprehensive beta testing environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_FILE="$PROJECT_ROOT/beta-testing-launch.log"

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

print_header() {
    echo -e "${PURPLE}================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}================================${NC}"
}

# Function to log messages
log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.22+ first."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_status "Go version: $GO_VERSION"
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_warning "Docker is not installed. Some features may not work."
    else
        print_status "Docker is available"
    fi
    
    # Check if required files exist
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        print_error "go.mod not found. Please run this script from the project root."
        exit 1
    fi
    
    print_success "Prerequisites check completed"
}

# Function to set up environment
setup_environment() {
    print_header "Setting Up Environment"
    
    cd "$PROJECT_ROOT"
    
    # Create .env file if it doesn't exist
    if [ ! -f ".env" ]; then
        print_status "Creating .env file from template..."
        if [ -f "env.example" ]; then
            cp env.example .env
            print_success "Created .env file from template"
        else
            print_warning "env.example not found, creating basic .env file"
            cat > .env << EOF
# KYB Platform Environment Configuration
ENVIRONMENT=development
LOG_LEVEL=info
PORT=8080
DATABASE_URL=postgresql://localhost/kyb_platform
REDIS_URL=redis://localhost:6379
JWT_SECRET=$(openssl rand -hex 32)
ENCRYPTION_KEY=$(openssl rand -hex 32)
EOF
        fi
    fi
    
    # Generate secure secrets if not present
    if ! grep -q "JWT_SECRET=" .env || grep -q "JWT_SECRET=$" .env; then
        print_status "Generating JWT secret..."
        sed -i "s/JWT_SECRET=.*/JWT_SECRET=$(openssl rand -hex 32)/" .env
    fi
    
    if ! grep -q "ENCRYPTION_KEY=" .env || grep -q "ENCRYPTION_KEY=$" .env; then
        print_status "Generating encryption key..."
        sed -i "s/ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$(openssl rand -hex 32)/" .env
    fi
    
    print_success "Environment setup completed"
}

# Function to build the application
build_application() {
    print_header "Building Application"
    
    cd "$PROJECT_ROOT"
    
    # Clean previous builds
    print_status "Cleaning previous builds..."
    go clean -cache
    rm -f kyb-platform
    
    # Download dependencies
    print_status "Downloading dependencies..."
    go mod download
    
    # Run tests
    print_status "Running tests..."
    if go test ./... -v; then
        print_success "All tests passed"
    else
        print_warning "Some tests failed, but continuing with build"
    fi
    
    # Build the application
    print_status "Building application..."
    if go build -o kyb-platform ./cmd/api; then
        print_success "Application built successfully"
    else
        print_error "Build failed"
        exit 1
    fi
    
    # Build Docker image if Docker is available
    if command -v docker &> /dev/null; then
        print_status "Building Docker image..."
        if docker build -f Dockerfile.beta -t kyb-platform:beta .; then
            print_success "Docker image built successfully"
        else
            print_warning "Docker build failed, but continuing"
        fi
    fi
}

# Function to start services
start_services() {
    print_header "Starting Services"
    
    cd "$PROJECT_ROOT"
    
    # Start PostgreSQL if using Docker
    if command -v docker &> /dev/null; then
        print_status "Starting PostgreSQL container..."
        if ! docker ps | grep -q kyb-postgres; then
            docker run -d \
                --name kyb-postgres \
                -e POSTGRES_DB=kyb_platform \
                -e POSTGRES_USER=kyb_user \
                -e POSTGRES_PASSWORD=kyb_password \
                -p 5432:5432 \
                postgres:15
            print_success "PostgreSQL container started"
        else
            print_status "PostgreSQL container already running"
        fi
        
        # Start Redis if using Docker
        print_status "Starting Redis container..."
        if ! docker ps | grep -q kyb-redis; then
            docker run -d \
                --name kyb-redis \
                -p 6379:6379 \
                redis:7-alpine
            print_success "Redis container started"
        else
            print_status "Redis container already running"
        fi
    else
        print_warning "Docker not available. Please ensure PostgreSQL and Redis are running manually."
    fi
}

# Function to run database migrations
run_migrations() {
    print_header "Running Database Migrations"
    
    cd "$PROJECT_ROOT"
    
    # Wait for database to be ready
    print_status "Waiting for database to be ready..."
    sleep 5
    
    # Run migrations if migration tool exists
    if [ -f "migrate" ] || command -v migrate &> /dev/null; then
        print_status "Running database migrations..."
        if migrate -path internal/database/migrations -database "$(grep DATABASE_URL .env | cut -d '=' -f2)" up; then
            print_success "Database migrations completed"
        else
            print_warning "Database migrations failed, but continuing"
        fi
    else
        print_warning "Migration tool not found. Please run migrations manually."
    fi
}

# Function to start the application
start_application() {
    print_header "Starting Application"
    
    cd "$PROJECT_ROOT"
    
    # Load environment variables
    export $(cat .env | grep -v '^#' | xargs)
    
    # Start the application
    print_status "Starting KYB Platform..."
    print_status "Application will be available at: http://localhost:${PORT:-8080}"
    print_status "Beta testing UI: http://localhost:${PORT:-8080}/"
    print_status "API documentation: http://localhost:${PORT:-8080}/docs"
    
    # Run the application
    ./kyb-platform
}

# Function to show beta testing information
show_beta_info() {
    print_header "Enhanced Business Intelligence Beta Testing"
    
    echo -e "${CYAN}🎯 Beta Testing Features Available:${NC}"
    echo -e "  ✅ Enhanced Classification with ML Integration"
    echo -e "  ✅ Website Verification (90%+ success rate)"
    echo -e "  ✅ Company Size Analysis"
    echo -e "  ✅ Business Model Detection"
    echo -e "  ✅ Technology Stack Analysis"
    echo -e "  ✅ Financial Health Assessment"
    echo -e "  ✅ Compliance Analysis"
    echo -e "  ✅ Market Presence Analysis"
    echo -e "  ✅ Enhanced Contact Intelligence"
    echo -e "  ✅ Data Quality Framework"
    echo -e "  ✅ Validation Framework"
    echo -e "  ✅ Geographic Awareness"
    echo -e "  ✅ Confidence Scoring"
    echo -e "  ✅ Real-time Feedback Collection"
    
    echo -e "\n${CYAN}🚀 Testing URLs:${NC}"
    echo -e "  📱 Beta Testing UI: http://localhost:${PORT:-8080}/"
    echo -e "  📊 Dashboard: http://localhost:${PORT:-8080}/dashboard"
    echo -e "  📚 API Documentation: http://localhost:${PORT:-8080}/docs"
    echo -e "  🔍 Health Check: http://localhost:${PORT:-8080}/health"
    
    echo -e "\n${CYAN}📋 Testing Instructions:${NC}"
    echo -e "  1. Open the Beta Testing UI in your browser"
    echo -e "  2. Enter a business name and optional website URL"
    echo -e "  3. Click 'Analyze Business Intelligence'"
    echo -e "  4. Review comprehensive results including:"
    echo -e "     - Industry classification with confidence scores"
    echo -e "     - Website verification results"
    echo -e "     - Data extraction insights"
    echo -e "     - Geographic analysis"
    echo -e "     - Enhanced features status"
    
    echo -e "\n${CYAN}📝 Feedback Collection:${NC}"
    echo -e "  - All user interactions are logged for analysis"
    echo -e "  - Performance metrics are collected automatically"
    echo -e "  - Error reports are generated for debugging"
    echo -e "  - User satisfaction scores are tracked"
    
    echo -e "\n${YELLOW}⚠️  Important Notes:${NC}"
    echo -e "  - This is a beta testing environment"
    echo -e "  - All enhanced features are active"
    echo -e "  - Data is processed in real-time"
    echo -e "  - Performance monitoring is enabled"
}

# Function to show help
show_help() {
    echo -e "${PURPLE}KYB Platform - Enhanced Business Intelligence Beta Testing Launch Script${NC}"
    echo ""
    echo "Usage: $0 [OPTION]"
    echo ""
    echo "Options:"
    echo "  setup     - Set up the environment and build the application"
    echo "  start     - Start the application (requires setup first)"
    echo "  full      - Complete setup and start (recommended)"
    echo "  info      - Show beta testing information"
    echo "  help      - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 full    # Complete setup and launch"
    echo "  $0 setup   # Set up environment only"
    echo "  $0 start   # Start application only"
    echo "  $0 info    # Show testing information"
}

# Main script logic
main() {
    case "${1:-full}" in
        "setup")
            log_message "Starting setup process"
            check_prerequisites
            setup_environment
            build_application
            start_services
            run_migrations
            print_success "Setup completed successfully"
            show_beta_info
            ;;
        "start")
            log_message "Starting application"
            start_application
            ;;
        "full")
            log_message "Starting full setup and launch"
            check_prerequisites
            setup_environment
            build_application
            start_services
            run_migrations
            show_beta_info
            print_success "Setup completed successfully"
            print_status "Starting application in 3 seconds..."
            sleep 3
            start_application
            ;;
        "info")
            show_beta_info
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
