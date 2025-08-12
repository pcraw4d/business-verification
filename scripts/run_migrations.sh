#!/bin/bash

# KYB Platform - Database Migration Script
# This script runs database migrations for the KYB platform

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

# Function to check prerequisites
check_prerequisites() {
    print_status $BLUE "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_status $RED "Error: Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in the project root
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        print_status $RED "Error: Not in project root directory"
        exit 1
    fi
    
    # Check if .env file exists
    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        print_status $RED "Error: .env file not found. Please run setup_supabase.sh first"
        exit 1
    fi
    
    print_status $GREEN "✓ Prerequisites check passed"
}

# Function to run migrations
run_migrations() {
    print_status $BLUE "Running database migrations..."
    
    cd "$PROJECT_ROOT"
    
    # Set environment variables
    export $(cat .env | grep -v '^#' | xargs)
    
    # Run migrations using Go
    print_status $BLUE "Executing migration system..."
    
    # Create a temporary migration runner
    cat > /tmp/migrate_runner.go << 'EOF'
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/pcraw4d/business-verification/internal/config"
    "github.com/pcraw4d/business-verification/internal/database"
    "github.com/pcraw4d/business-verification/internal/observability"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // Initialize logger
    logger, err := observability.NewLogger(&cfg.Observability)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    
    // Initialize database
    db, err := database.NewDatabase(&cfg.Database, logger)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()
    
    // Get migration system
    migrationSystem := database.NewMigrationSystem(db.DB, &cfg.Database)
    
    // Initialize migration table
    ctx := context.Background()
    err = migrationSystem.InitializeMigrationTable(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize migration table: %v", err)
    }
    
    // Run migrations
    err = migrationSystem.RunMigrations(ctx)
    if err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }
    
    // Get migration status
    status, err := migrationSystem.GetMigrationStatus(ctx)
    if err != nil {
        log.Fatalf("Failed to get migration status: %v", err)
    }
    
    logger.Info("Migrations completed successfully", "status", status)
    println("✅ Migrations completed successfully!")
}
EOF
    
    # Run the migration
    cd /tmp
    go mod init migrate_runner
    go mod tidy
    go run migrate_runner.go
    
    # Cleanup
    rm -rf /tmp/migrate_runner.go /tmp/go.mod /tmp/go.sum
    
    print_status $GREEN "✓ Migrations completed successfully"
}

# Function to verify migrations
verify_migrations() {
    print_status $BLUE "Verifying migrations..."
    
    # Test database connection
    print_status $BLUE "Testing database connection..."
    
    # Create a simple connection test
    cat > /tmp/connection_test.go << 'EOF'
package main

import (
    "database/sql"
    "log"
    "os"
    
    _ "github.com/lib/pq"
)

func main() {
    // Get database connection string from environment
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        // Construct from individual env vars
        host := os.Getenv("DB_HOST")
        port := os.Getenv("DB_PORT")
        user := os.Getenv("DB_USERNAME")
        password := os.Getenv("DB_PASSWORD")
        database := os.Getenv("DB_DATABASE")
        sslMode := os.Getenv("DB_SSL_MODE")
        
        dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", 
            user, password, host, port, database, sslMode)
    }
    
    // Connect to database
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Test connection
    err = db.Ping()
    if err != nil {
        log.Fatalf("Failed to ping database: %v", err)
    }
    
    // Check if migrations table exists
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&count)
    if err != nil {
        log.Fatalf("Failed to query migrations table: %v", err)
    }
    
    println("✅ Database connection successful!")
    println("✅ Migrations table found with", count, "migrations")
}
EOF
    
    cd /tmp
    go mod init connection_test
    go mod tidy
    go run connection_test.go
    
    # Cleanup
    rm -rf /tmp/connection_test.go /tmp/go.mod /tmp/go.sum
    
    print_status $GREEN "✓ Database verification completed"
}

# Main execution
main() {
    print_status $BLUE "🚀 KYB Platform - Database Migration Runner"
    print_status $BLUE "============================================="
    
    check_prerequisites
    run_migrations
    verify_migrations
    
    print_status $GREEN "🎉 Migration process completed successfully!"
    print_status $BLUE "Next steps:"
    print_status $BLUE "1. Start the application: go run ./cmd/api"
    print_status $BLUE "2. Test the API endpoints"
    print_status $BLUE "3. Check the Supabase dashboard for data"
}

# Run main function
main "$@"
