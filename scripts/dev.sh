#!/bin/bash

# KYB Platform - Development Environment Management Script
# Provides easy commands for managing the development environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
COMPOSE_FILE="docker-compose.dev.yml"
SERVICE_NAME="kyb-platform"

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
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  start       - Start the development environment"
    echo "  stop        - Stop the development environment"
    echo "  restart     - Restart the development environment"
    echo "  build       - Build the development environment"
    echo "  logs        - Show logs from all services"
    echo "  logs-app    - Show logs from the application only"
    echo "  shell       - Open shell in the application container"
    echo "  db-shell    - Open shell in the database container"
    echo "  db-reset    - Reset the database (WARNING: destroys all data)"
    echo "  status      - Show status of all services"
    echo "  clean       - Clean up containers, volumes, and images"
    echo "  test        - Run tests in the development environment"
    echo "  migrate     - Run database migrations"
    echo "  seed        - Seed the database with sample data"
    echo "  health      - Check health of all services"
    echo "  urls        - Show URLs for all services"
    echo ""
    echo "Options:"
    echo "  -f, --file FILE    - Use specific compose file (default: docker-compose.dev.yml)"
    echo "  -s, --service NAME - Target specific service"
    echo "  -h, --help         - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start"
    echo "  $0 logs-app"
    echo "  $0 shell"
    echo "  $0 db-reset"
    echo "  $0 -f docker-compose.yml start"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Function to start the development environment
start_dev() {
    print_status "Starting development environment..."
    docker-compose -f $COMPOSE_FILE up -d
    print_success "Development environment started"
    print_status "Waiting for services to be ready..."
    sleep 10
    show_urls
}

# Function to stop the development environment
stop_dev() {
    print_status "Stopping development environment..."
    docker-compose -f $COMPOSE_FILE down
    print_success "Development environment stopped"
}

# Function to restart the development environment
restart_dev() {
    print_status "Restarting development environment..."
    docker-compose -f $COMPOSE_FILE restart
    print_success "Development environment restarted"
}

# Function to build the development environment
build_dev() {
    print_status "Building development environment..."
    docker-compose -f $COMPOSE_FILE build --no-cache
    print_success "Development environment built"
}

# Function to show logs
show_logs() {
    if [ "$1" = "app" ]; then
        print_status "Showing application logs..."
        docker-compose -f $COMPOSE_FILE logs -f $SERVICE_NAME
    else
        print_status "Showing all logs..."
        docker-compose -f $COMPOSE_FILE logs -f
    fi
}

# Function to open shell in container
open_shell() {
    local container=$1
    print_status "Opening shell in $container..."
    docker-compose -f $COMPOSE_FILE exec $container /bin/sh
}

# Function to reset database
reset_database() {
    print_warning "This will destroy all data in the database!"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "Resetting database..."
        docker-compose -f $COMPOSE_FILE down -v
        docker-compose -f $COMPOSE_FILE up -d postgres
        sleep 10
        docker-compose -f $COMPOSE_FILE up -d
        print_success "Database reset complete"
    else
        print_status "Database reset cancelled"
    fi
}

# Function to show status
show_status() {
    print_status "Service status:"
    docker-compose -f $COMPOSE_FILE ps
}

# Function to clean up
clean_up() {
    print_warning "This will remove all containers, volumes, and images!"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "Cleaning up..."
        docker-compose -f $COMPOSE_FILE down -v --rmi all
        docker system prune -f
        print_success "Cleanup complete"
    else
        print_status "Cleanup cancelled"
    fi
}

# Function to run tests
run_tests() {
    print_status "Running tests..."
    docker-compose -f $COMPOSE_FILE exec $SERVICE_NAME go test -v ./...
}

# Function to run migrations
run_migrations() {
    print_status "Running database migrations..."
    docker-compose -f $COMPOSE_FILE exec $SERVICE_NAME go run cmd/migrate/main.go
}

# Function to seed database
seed_database() {
    print_status "Seeding database..."
    docker-compose -f $COMPOSE_FILE exec postgres psql -U kyb_user -d kyb_platform -f /docker-entrypoint-initdb.d/init-db.sql
}

# Function to check health
check_health() {
    print_status "Checking service health..."
    
    # Check application health
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        print_success "Application: Healthy"
    else
        print_error "Application: Unhealthy"
    fi
    
    # Check database health
    if docker-compose -f $COMPOSE_FILE exec -T postgres pg_isready -U kyb_user -d kyb_platform > /dev/null 2>&1; then
        print_success "Database: Healthy"
    else
        print_error "Database: Unhealthy"
    fi
    
    # Check Redis health
    if docker-compose -f $COMPOSE_FILE exec -T redis redis-cli ping > /dev/null 2>&1; then
        print_success "Redis: Healthy"
    else
        print_error "Redis: Unhealthy"
    fi
}

# Function to show URLs
show_urls() {
    echo ""
    print_status "Service URLs:"
    echo "  Application:     http://localhost:8080"
    echo "  API Docs:        http://localhost:8080/docs"
    echo "  Health Check:    http://localhost:8080/health"
    echo "  Metrics:         http://localhost:8080/metrics"
    echo "  Mailhog (SMTP):  http://localhost:8025"
    echo ""
    print_status "Database:"
    echo "  Host: localhost"
    echo "  Port: 5432"
    echo "  Database: kyb_platform"
    echo "  Username: kyb_user"
    echo "  Password: kyb_password"
    echo ""
    print_status "Test Credentials:"
    echo "  Admin: admin@kybplatform.com / admin123"
    echo "  User:  test@kybplatform.com / test123"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--file)
            COMPOSE_FILE="$2"
            shift 2
            ;;
        -s|--service)
            SERVICE_NAME="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        start)
            check_docker
            start_dev
            exit 0
            ;;
        stop)
            check_docker
            stop_dev
            exit 0
            ;;
        restart)
            check_docker
            restart_dev
            exit 0
            ;;
        build)
            check_docker
            build_dev
            exit 0
            ;;
        logs)
            check_docker
            show_logs
            exit 0
            ;;
        logs-app)
            check_docker
            show_logs app
            exit 0
            ;;
        shell)
            check_docker
            open_shell $SERVICE_NAME
            exit 0
            ;;
        db-shell)
            check_docker
            open_shell postgres
            exit 0
            ;;
        db-reset)
            check_docker
            reset_database
            exit 0
            ;;
        status)
            check_docker
            show_status
            exit 0
            ;;
        clean)
            check_docker
            clean_up
            exit 0
            ;;
        test)
            check_docker
            run_tests
            exit 0
            ;;
        migrate)
            check_docker
            run_migrations
            exit 0
            ;;
        seed)
            check_docker
            seed_database
            exit 0
            ;;
        health)
            check_docker
            check_health
            exit 0
            ;;
        urls)
            show_urls
            exit 0
            ;;
        *)
            print_error "Unknown command: $1"
            show_usage
            exit 1
            ;;
    esac
done

# If no command provided, show usage
show_usage
