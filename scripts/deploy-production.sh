#!/bin/bash

# Production Deployment Script for V3 API
# This script handles the complete production deployment process

set -e

# Configuration
APP_NAME="business-verification-v3-api"
DEPLOYMENT_ENV="production"
BUILD_DIR="build"
CONFIG_DIR="configs"
BACKUP_DIR="backups"
LOG_DIR="logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go 1.22 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.22"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        error "Go version $GO_VERSION is too old. Please upgrade to Go 1.22 or later."
        exit 1
    fi
    
    # Check if production config exists
    if [ ! -f "$CONFIG_DIR/production.env" ]; then
        error "Production configuration file not found: $CONFIG_DIR/production.env"
        exit 1
    fi
    
    success "Prerequisites check passed"
}

# Create backup
create_backup() {
    log "Creating backup of current deployment..."
    
    if [ -f "$APP_NAME" ]; then
        BACKUP_FILE="$BACKUP_DIR/${APP_NAME}_$(date +%Y%m%d_%H%M%S).backup"
        mkdir -p "$BACKUP_DIR"
        cp "$APP_NAME" "$BACKUP_FILE"
        success "Backup created: $BACKUP_FILE"
    else
        warning "No existing application to backup"
    fi
}

# Build application
build_application() {
    log "Building application for production..."
    
    # Set build flags for production
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
    
    # Build with optimizations
    go build -ldflags="-s -w" -o "$BUILD_DIR/$APP_NAME" ./cmd/api/
    
    if [ $? -eq 0 ]; then
        success "Application built successfully"
    else
        error "Build failed"
        exit 1
    fi
}

# Run tests
run_tests() {
    log "Running tests..."
    
    # Run unit tests
    go test ./... -v
    
    if [ $? -eq 0 ]; then
        success "All tests passed"
    else
        error "Tests failed"
        exit 1
    fi
}

# Validate configuration
validate_configuration() {
    log "Validating production configuration..."
    
    # Load production configuration
    source "$CONFIG_DIR/production.env"
    
    # Check required environment variables
    REQUIRED_VARS=(
        "JWT_SECRET"
        "SUPABASE_URL"
        "SUPABASE_ANON_KEY"
        "SUPABASE_SERVICE_ROLE_KEY"
    )
    
    for var in "${REQUIRED_VARS[@]}"; do
        if [ -z "${!var}" ]; then
            error "Required environment variable not set: $var"
            exit 1
        fi
    done
    
    success "Configuration validation passed"
}

# Deploy application
deploy_application() {
    log "Deploying application..."
    
    # Stop existing application if running
    if pgrep -f "$APP_NAME" > /dev/null; then
        log "Stopping existing application..."
        pkill -f "$APP_NAME" || true
        sleep 2
    fi
    
    # Copy new binary
    cp "$BUILD_DIR/$APP_NAME" .
    chmod +x "$APP_NAME"
    
    success "Application deployed successfully"
}

# Start application
start_application() {
    log "Starting application..."
    
    # Create log directory
    mkdir -p "$LOG_DIR"
    
    # Start application in background
    nohup ./"$APP_NAME" > "$LOG_DIR/app.log" 2>&1 &
    APP_PID=$!
    
    # Wait a moment for startup
    sleep 3
    
    # Check if application is running
    if kill -0 $APP_PID 2>/dev/null; then
        success "Application started successfully (PID: $APP_PID)"
        echo $APP_PID > "$LOG_DIR/app.pid"
    else
        error "Failed to start application"
        exit 1
    fi
}

# Health check
health_check() {
    log "Performing health check..."
    
    # Wait for application to be ready
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            success "Health check passed"
            return 0
        fi
        
        log "Health check attempt $attempt/$max_attempts failed, retrying..."
        sleep 2
        ((attempt++))
    done
    
    error "Health check failed after $max_attempts attempts"
    return 1
}

# Run integration tests
run_integration_tests() {
    log "Running integration tests..."
    
    # Start test server if not already running
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        error "Application is not running. Cannot run integration tests."
        return 1
    fi
    
    # Run integration tests
    if [ -f "scripts/test-v3-api.sh" ]; then
        ./scripts/test-v3-api.sh
        if [ $? -eq 0 ]; then
            success "Integration tests passed"
        else
            error "Integration tests failed"
            return 1
        fi
    else
        warning "Integration test script not found, skipping"
    fi
}

# Setup monitoring
setup_monitoring() {
    log "Setting up monitoring..."
    
    # Create monitoring directory
    mkdir -p monitoring
    
    # Create systemd service file
    cat > monitoring/"$APP_NAME".service << EOF
[Unit]
Description=Business Verification V3 API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/$APP_NAME
Restart=always
RestartSec=5
Environment=NODE_ENV=production

[Install]
WantedBy=multi-user.target
EOF
    
    success "Monitoring setup completed"
}

# Rollback function
rollback() {
    log "Rolling back deployment..."
    
    # Stop current application
    if [ -f "$LOG_DIR/app.pid" ]; then
        PID=$(cat "$LOG_DIR/app.pid")
        kill $PID 2>/dev/null || true
        rm -f "$LOG_DIR/app.pid"
    fi
    
    # Find latest backup
    LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/"$APP_NAME"*.backup 2>/dev/null | head -n1)
    
    if [ -n "$LATEST_BACKUP" ]; then
        cp "$LATEST_BACKUP" "$APP_NAME"
        chmod +x "$APP_NAME"
        success "Rolled back to: $LATEST_BACKUP"
    else
        error "No backup found for rollback"
        exit 1
    fi
    
    # Restart application
    start_application
}

# Cleanup function
cleanup() {
    log "Cleaning up build artifacts..."
    
    # Remove build directory
    rm -rf "$BUILD_DIR"
    
    # Remove old backups (keep last 5)
    if [ -d "$BACKUP_DIR" ]; then
        cd "$BACKUP_DIR"
        ls -t "$APP_NAME"*.backup 2>/dev/null | tail -n +6 | xargs rm -f 2>/dev/null || true
        cd ..
    fi
    
    success "Cleanup completed"
}

# Main deployment function
main() {
    log "Starting production deployment for $APP_NAME"
    
    # Create necessary directories
    mkdir -p "$BUILD_DIR" "$BACKUP_DIR" "$LOG_DIR"
    
    # Run deployment steps
    check_prerequisites
    create_backup
    build_application
    run_tests
    validate_configuration
    deploy_application
    start_application
    
    # Wait for application to start
    sleep 5
    
    # Perform health check
    if health_check; then
        run_integration_tests
        setup_monitoring
        cleanup
        success "Production deployment completed successfully!"
        
        log "Application is running at: http://localhost:8080"
        log "Health check endpoint: http://localhost:8080/health"
        log "API documentation: http://localhost:8080/api/v3/docs"
        
    else
        error "Health check failed, rolling back..."
        rollback
        exit 1
    fi
}

# Handle command line arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "rollback")
        rollback
        ;;
    "health")
        health_check
        ;;
    "stop")
        if [ -f "$LOG_DIR/app.pid" ]; then
            PID=$(cat "$LOG_DIR/app.pid")
            kill $PID 2>/dev/null || true
            rm -f "$LOG_DIR/app.pid"
            success "Application stopped"
        else
            warning "No PID file found"
        fi
        ;;
    "restart")
        if [ -f "$LOG_DIR/app.pid" ]; then
            PID=$(cat "$LOG_DIR/app.pid")
            kill $PID 2>/dev/null || true
            rm -f "$LOG_DIR/app.pid"
        fi
        start_application
        ;;
    *)
        echo "Usage: $0 {deploy|rollback|health|stop|restart}"
        echo "  deploy   - Deploy the application (default)"
        echo "  rollback - Rollback to previous version"
        echo "  health   - Check application health"
        echo "  stop     - Stop the application"
        echo "  restart  - Restart the application"
        exit 1
        ;;
esac
