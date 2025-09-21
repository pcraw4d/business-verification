#!/bin/bash

# Backup and Recovery Testing Script
# This script runs the comprehensive backup and recovery tests for the Supabase Table Improvement project

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
BACKUP_DIR="/tmp/backup_recovery_test_$(date +%Y%m%d_%H%M%S)"

# Default configuration
SUPABASE_URL="${SUPABASE_URL:-postgresql://postgres:password@localhost:5432/kyb_platform}"
TEST_DATABASE_URL="${TEST_DATABASE_URL:-postgresql://postgres:password@localhost:5432/kyb_platform_test}"
TEST_DATA_SIZE="${TEST_DATA_SIZE:-1000}"
RECOVERY_TIMEOUT="${RECOVERY_TIMEOUT:-10m}"
VALIDATION_RETRIES="${VALIDATION_RETRIES:-3}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go to run the tests."
        exit 1
    fi
    
    # Check if pg_dump is available
    if ! command -v pg_dump &> /dev/null; then
        log_error "pg_dump is not available. Please install PostgreSQL client tools."
        exit 1
    fi
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        log_error "psql is not available. Please install PostgreSQL client tools."
        exit 1
    fi
    
    # Check if backup directory is writable
    if ! mkdir -p "$BACKUP_DIR" 2>/dev/null; then
        log_error "Cannot create backup directory: $BACKUP_DIR"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

setup_environment() {
    log_info "Setting up test environment..."
    
    # Export environment variables
    export SUPABASE_URL
    export TEST_DATABASE_URL
    export BACKUP_DIRECTORY="$BACKUP_DIR"
    export TEST_DATA_SIZE
    export RECOVERY_TIMEOUT
    export VALIDATION_RETRIES
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    
    log_success "Environment setup completed"
    log_info "Configuration:"
    log_info "  Supabase URL: $SUPABASE_URL"
    log_info "  Test Database URL: $TEST_DATABASE_URL"
    log_info "  Backup Directory: $BACKUP_DIR"
    log_info "  Test Data Size: $TEST_DATA_SIZE"
    log_info "  Recovery Timeout: $RECOVERY_TIMEOUT"
    log_info "  Validation Retries: $VALIDATION_RETRIES"
}

run_tests() {
    log_info "Running backup and recovery tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run individual tests
    log_info "Running backup procedures test..."
    if go test -v -run TestBackupProceduresOnly ./internal/testing; then
        log_success "Backup procedures test passed"
    else
        log_error "Backup procedures test failed"
        return 1
    fi
    
    log_info "Running recovery scenarios test..."
    if go test -v -run TestRecoveryScenariosOnly ./internal/testing; then
        log_success "Recovery scenarios test passed"
    else
        log_error "Recovery scenarios test failed"
        return 1
    fi
    
    log_info "Running data restoration validation..."
    if go test -v -run TestDataRestorationOnly ./internal/testing; then
        log_success "Data restoration validation passed"
    else
        log_error "Data restoration validation failed"
        return 1
    fi
    
    log_info "Running point-in-time recovery test..."
    if go test -v -run TestPointInTimeRecoveryOnly ./internal/testing; then
        log_success "Point-in-time recovery test passed"
    else
        log_error "Point-in-time recovery test failed"
        return 1
    fi
    
    # Run complete integration test
    log_info "Running complete integration test..."
    if go test -v -run TestBackupRecoveryIntegration ./internal/testing; then
        log_success "Complete integration test passed"
    else
        log_error "Complete integration test failed"
        return 1
    fi
    
    log_success "All tests completed successfully"
}

run_benchmarks() {
    log_info "Running backup and recovery benchmarks..."
    
    cd "$PROJECT_ROOT"
    
    log_info "Benchmarking backup procedures..."
    go test -bench=BenchmarkBackupProcedures ./internal/testing
    
    log_info "Benchmarking recovery procedures..."
    go test -bench=BenchmarkRecoveryProcedures ./internal/testing
    
    log_success "Benchmarks completed"
}

show_results() {
    log_info "Test results and reports:"
    
    if [ -f "$BACKUP_DIR/backup_recovery_test_report.json" ]; then
        log_success "JSON report: $BACKUP_DIR/backup_recovery_test_report.json"
    fi
    
    if [ -f "$BACKUP_DIR/backup_recovery_test_summary.txt" ]; then
        log_success "Human-readable summary: $BACKUP_DIR/backup_recovery_test_summary.txt"
        echo ""
        log_info "Summary content:"
        cat "$BACKUP_DIR/backup_recovery_test_summary.txt"
    fi
}

cleanup() {
    log_info "Cleaning up test environment..."
    
    # Remove backup directory if requested
    if [ "${CLEANUP:-true}" = "true" ]; then
        rm -rf "$BACKUP_DIR"
        log_success "Cleanup completed"
    else
        log_info "Backup directory preserved: $BACKUP_DIR"
    fi
}

show_help() {
    echo "Backup and Recovery Testing Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -b, --benchmarks        Run benchmarks only"
    echo "  -t, --tests             Run tests only (default)"
    echo "  -a, --all               Run tests and benchmarks"
    echo "  --no-cleanup            Don't clean up backup directory"
    echo "  --backup-dir DIR        Use custom backup directory"
    echo ""
    echo "Environment Variables:"
    echo "  SUPABASE_URL            Main database connection string"
    echo "  TEST_DATABASE_URL       Test database connection string"
    echo "  TEST_DATA_SIZE          Number of test records to create"
    echo "  RECOVERY_TIMEOUT        Timeout for recovery operations"
    echo "  VALIDATION_RETRIES      Number of validation retries"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run all tests"
    echo "  $0 --benchmarks         # Run benchmarks only"
    echo "  $0 --no-cleanup         # Keep backup files"
    echo "  $0 --backup-dir /tmp/my_backup  # Use custom backup directory"
}

# Main execution
main() {
    local run_tests_flag=true
    local run_benchmarks_flag=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -b|--benchmarks)
                run_tests_flag=false
                run_benchmarks_flag=true
                shift
                ;;
            -t|--tests)
                run_tests_flag=true
                run_benchmarks_flag=false
                shift
                ;;
            -a|--all)
                run_tests_flag=true
                run_benchmarks_flag=true
                shift
                ;;
            --no-cleanup)
                export CLEANUP=false
                shift
                ;;
            --backup-dir)
                BACKUP_DIR="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Set up trap for cleanup
    trap cleanup EXIT
    
    # Run the tests
    check_prerequisites
    setup_environment
    
    if [ "$run_tests_flag" = true ]; then
        run_tests
    fi
    
    if [ "$run_benchmarks_flag" = true ]; then
        run_benchmarks
    fi
    
    show_results
    
    log_success "Backup and recovery testing completed successfully!"
}

# Run main function
main "$@"
