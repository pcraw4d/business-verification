#!/bin/bash

# KYB Platform Production Optimization Script
# This script runs the complete production optimization process

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
CONFIG_FILE="$PROJECT_ROOT/config/optimization.yaml"
LOG_FILE="$PROJECT_ROOT/logs/optimization-$(date +%Y%m%d-%H%M%S).log"

# Default values
PHASE="all"
DRY_RUN=false
VERBOSE=false
SKIP_CHECKS=false

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
    cat << EOF
Usage: $0 [OPTIONS]

KYB Platform Production Optimization Script

OPTIONS:
    -p, --phase PHASE       Optimization phase to run (database, cache, performance, monitoring, all)
    -d, --dry-run          Run in dry-run mode (no actual changes)
    -v, --verbose          Enable verbose logging
    -s, --skip-checks      Skip pre-optimization checks
    -c, --config FILE      Configuration file path (default: config/optimization.yaml)
    -h, --help             Show this help message

EXAMPLES:
    $0 --phase database --dry-run
    $0 --phase all --verbose
    $0 --phase cache --config custom-config.yaml

PHASES:
    database      - Database optimization (indexing, connection pooling, query optimization)
    cache         - Cache optimization (Redis configuration, multi-level caching)
    performance   - API performance optimization (compression, async processing)
    monitoring    - Monitoring and alerting optimization
    all           - Run all optimization phases (default)

EOF
}

# Function to parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--phase)
                PHASE="$2"
                shift 2
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -s|--skip-checks)
                SKIP_CHECKS=true
                shift
                ;;
            -c|--config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Function to validate phase
validate_phase() {
    case $PHASE in
        database|cache|performance|monitoring|all)
            return 0
            ;;
        *)
            print_error "Invalid phase: $PHASE"
            print_error "Valid phases: database, cache, performance, monitoring, all"
            exit 1
            ;;
    esac
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.22"
    if ! printf '%s\n%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
        print_error "Go version $GO_VERSION is below required version $REQUIRED_VERSION"
        exit 1
    fi
    
    # Check if configuration file exists
    if [[ ! -f "$CONFIG_FILE" ]]; then
        print_error "Configuration file not found: $CONFIG_FILE"
        exit 1
    fi
    
    # Check if required environment variables are set
    if [[ -z "$DATABASE_URL" ]]; then
        print_warning "DATABASE_URL not set, using default"
    fi
    
    if [[ -z "$REDIS_URL" ]]; then
        print_warning "REDIS_URL not set, using default"
    fi
    
    print_success "Prerequisites check completed"
}

# Function to create necessary directories
create_directories() {
    print_status "Creating necessary directories..."
    
    mkdir -p "$PROJECT_ROOT/logs"
    mkdir -p "$PROJECT_ROOT/reports"
    mkdir -p "$PROJECT_ROOT/tmp"
    
    print_success "Directories created"
}

# Function to build the optimization binary
build_optimization_binary() {
    print_status "Building optimization binary..."
    
    cd "$PROJECT_ROOT"
    
    # Build the optimization binary
    if go build -o bin/optimization cmd/optimization/main.go; then
        print_success "Optimization binary built successfully"
    else
        print_error "Failed to build optimization binary"
        exit 1
    fi
}

# Function to run optimization
run_optimization() {
    print_status "Running production optimization..."
    print_status "Phase: $PHASE"
    print_status "Dry Run: $DRY_RUN"
    print_status "Config: $CONFIG_FILE"
    print_status "Log File: $LOG_FILE"
    
    cd "$PROJECT_ROOT"
    
    # Prepare command arguments
    ARGS=(
        "-config" "$CONFIG_FILE"
        "-phase" "$PHASE"
    )
    
    if [[ "$DRY_RUN" == true ]]; then
        ARGS+=("-dry-run")
    fi
    
    if [[ "$VERBOSE" == true ]]; then
        ARGS+=("-verbose")
    fi
    
    # Run optimization
    if ./bin/optimization "${ARGS[@]}" 2>&1 | tee "$LOG_FILE"; then
        print_success "Optimization completed successfully"
    else
        print_error "Optimization failed. Check log file: $LOG_FILE"
        exit 1
    fi
}

# Function to generate optimization report
generate_report() {
    print_status "Generating optimization report..."
    
    REPORT_FILE="$PROJECT_ROOT/reports/optimization-report-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# KYB Platform Production Optimization Report

**Generated**: $(date)
**Phase**: $PHASE
**Dry Run**: $DRY_RUN
**Configuration**: $CONFIG_FILE

## Summary

This report summarizes the production optimization process for the KYB Platform.

## Optimization Details

- **Phase**: $PHASE
- **Mode**: $([ "$DRY_RUN" == true ] && echo "Dry Run" || echo "Production")
- **Configuration File**: $CONFIG_FILE
- **Log File**: $LOG_FILE

## Results

The optimization process has been completed. Please review the log file for detailed results.

## Next Steps

1. Review the optimization results
2. Monitor system performance
3. Adjust configuration as needed
4. Schedule regular optimization runs

## Files Generated

- Log File: $LOG_FILE
- Report File: $REPORT_FILE

EOF
    
    print_success "Optimization report generated: $REPORT_FILE"
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up temporary files..."
    
    # Remove temporary files if any
    rm -f "$PROJECT_ROOT/tmp/optimization-*"
    
    print_success "Cleanup completed"
}

# Function to show final summary
show_summary() {
    echo
    print_success "🎉 Production optimization completed successfully!"
    echo
    echo "Summary:"
    echo "  Phase: $PHASE"
    echo "  Mode: $([ "$DRY_RUN" == true ] && echo "Dry Run" || echo "Production")"
    echo "  Log File: $LOG_FILE"
    echo "  Report File: $PROJECT_ROOT/reports/optimization-report-$(date +%Y%m%d-%H%M%S).md"
    echo
    echo "Next steps:"
    echo "  1. Review the optimization results"
    echo "  2. Monitor system performance"
    echo "  3. Adjust configuration as needed"
    echo "  4. Schedule regular optimization runs"
    echo
}

# Main function
main() {
    echo "🚀 KYB Platform Production Optimization"
    echo "========================================"
    echo
    
    # Parse command line arguments
    parse_arguments "$@"
    
    # Validate phase
    validate_phase
    
    # Check prerequisites
    if [[ "$SKIP_CHECKS" != true ]]; then
        check_prerequisites
    fi
    
    # Create necessary directories
    create_directories
    
    # Build optimization binary
    build_optimization_binary
    
    # Run optimization
    run_optimization
    
    # Generate report
    generate_report
    
    # Cleanup
    cleanup
    
    # Show final summary
    show_summary
}

# Trap to handle script interruption
trap 'print_error "Script interrupted"; exit 1' INT TERM

# Run main function
main "$@"
