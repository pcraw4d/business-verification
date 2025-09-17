#!/bin/bash

# =============================================================================
# TASK 3.2.5: EXECUTE MANUFACTURING KEYWORDS IMPLEMENTATION
# =============================================================================
# This script executes the manufacturing keywords implementation for Task 3.2.5
# of the Comprehensive Classification Improvement Plan.
# 
# What this script does:
# 1. Adds comprehensive manufacturing keywords for all 4 manufacturing industries
# 2. Tests the implementation to ensure >85% classification accuracy
# 3. Validates keyword coverage and performance
# 4. Provides detailed reporting on the implementation
# =============================================================================

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_FILE="$PROJECT_ROOT/logs/manufacturing-keywords-$(date +%Y%m%d_%H%M%S).log"

# Create logs directory if it doesn't exist
mkdir -p "$PROJECT_ROOT/logs"

# Database configuration (adjust as needed)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-postgres}"
DB_USER="${DB_USER:-postgres}"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}[$(date '+%Y-%m-%d %H:%M:%S')] ${message}${NC}" | tee -a "$LOG_FILE"
}

# Function to print section headers
print_section() {
    local title=$1
    echo "" | tee -a "$LOG_FILE"
    echo "=============================================================================" | tee -a "$LOG_FILE"
    echo "$title" | tee -a "$LOG_FILE"
    echo "=============================================================================" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
}

# Function to check if database is accessible
check_database() {
    print_status $BLUE "Checking database connectivity..."
    
    if command -v psql >/dev/null 2>&1; then
        if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
            print_status $GREEN "✓ Database connection successful"
            return 0
        else
            print_status $RED "✗ Database connection failed"
            return 1
        fi
    else
        print_status $YELLOW "⚠ psql not found. Please ensure PostgreSQL client is installed."
        return 1
    fi
}

# Function to execute SQL script
execute_sql() {
    local script_file=$1
    local description=$2
    
    print_status $BLUE "Executing: $description"
    
    if [ ! -f "$script_file" ]; then
        print_status $RED "✗ Script file not found: $script_file"
        return 1
    fi
    
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$script_file" >> "$LOG_FILE" 2>&1; then
        print_status $GREEN "✓ $description completed successfully"
        return 0
    else
        print_status $RED "✗ $description failed"
        return 1
    fi
}

# Function to run tests
run_tests() {
    local test_file=$1
    local description=$2
    
    print_status $BLUE "Running tests: $description"
    
    if [ ! -f "$test_file" ]; then
        print_status $RED "✗ Test file not found: $test_file"
        return 1
    fi
    
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$test_file" >> "$LOG_FILE" 2>&1; then
        print_status $GREEN "✓ $description completed successfully"
        return 0
    else
        print_status $RED "✗ $description failed"
        return 1
    fi
}

# Function to validate implementation
validate_implementation() {
    print_section "VALIDATING MANUFACTURING KEYWORDS IMPLEMENTATION"
    
    # Check if all manufacturing industries exist
    print_status $BLUE "Validating manufacturing industries..."
    
    local industry_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) 
        FROM industries 
        WHERE name IN ('Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing')
        AND is_active = true;
    " | tr -d ' ')
    
    if [ "$industry_count" -eq 4 ]; then
        print_status $GREEN "✓ All 4 manufacturing industries found"
    else
        print_status $RED "✗ Expected 4 manufacturing industries, found $industry_count"
        return 1
    fi
    
    # Check keyword counts
    print_status $BLUE "Validating keyword counts..."
    
    local total_keywords=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) 
        FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name IN ('Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing')
        AND kw.is_active = true;
    " | tr -d ' ')
    
    if [ "$total_keywords" -ge 200 ]; then
        print_status $GREEN "✓ Total manufacturing keywords: $total_keywords (>= 200 required)"
    else
        print_status $RED "✗ Insufficient keywords: $total_keywords (>= 200 required)"
        return 1
    fi
    
    # Check weight ranges
    print_status $BLUE "Validating keyword weight ranges..."
    
    local invalid_weights=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) 
        FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name IN ('Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing')
        AND (kw.base_weight < 0.5000 OR kw.base_weight > 1.0000);
    " | tr -d ' ')
    
    if [ "$invalid_weights" -eq 0 ]; then
        print_status $GREEN "✓ All keyword weights are within valid range (0.5000-1.0000)"
    else
        print_status $RED "✗ Found $invalid_weights keywords with invalid weights"
        return 1
    fi
    
    return 0
}

# Function to generate summary report
generate_summary_report() {
    print_section "MANUFACTURING KEYWORDS IMPLEMENTATION SUMMARY"
    
    print_status $BLUE "Generating summary report..."
    
    # Get detailed statistics
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT 
            'MANUFACTURING KEYWORDS SUMMARY' as report_type,
            '' as spacer;
        
        SELECT 
            i.name as industry_name,
            i.confidence_threshold,
            COUNT(kw.keyword) as keyword_count,
            ROUND(MIN(kw.base_weight), 4) as min_weight,
            ROUND(MAX(kw.base_weight), 4) as max_weight,
            ROUND(AVG(kw.base_weight), 4) as avg_weight
        FROM industries i
        LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
        WHERE i.name IN (
            'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing'
        )
        GROUP BY i.name, i.confidence_threshold
        ORDER BY keyword_count DESC;
        
        SELECT 
            'OVERALL STATISTICS' as summary_type,
            '' as spacer;
        
        SELECT 
            COUNT(DISTINCT i.id) as total_industries,
            COUNT(kw.keyword) as total_keywords,
            ROUND(AVG(kw.base_weight), 4) as avg_keyword_weight,
            COUNT(CASE WHEN kw.base_weight >= 0.8000 THEN 1 END) as high_weight_keywords,
            COUNT(CASE WHEN kw.base_weight >= 0.5000 AND kw.base_weight < 0.8000 THEN 1 END) as medium_weight_keywords
        FROM industries i
        JOIN keyword_weights kw ON i.id = kw.industry_id
        WHERE i.name IN (
            'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing'
        )
        AND kw.is_active = true;
    " | tee -a "$LOG_FILE"
}

# Main execution function
main() {
    print_section "TASK 3.2.5: MANUFACTURING KEYWORDS IMPLEMENTATION"
    
    print_status $BLUE "Starting manufacturing keywords implementation..."
    print_status $BLUE "Log file: $LOG_FILE"
    
    # Check database connectivity
    if ! check_database; then
        print_status $RED "Database connectivity check failed. Exiting."
        exit 1
    fi
    
    # Execute manufacturing keywords script
    if ! execute_sql "$SCRIPT_DIR/add-manufacturing-keywords.sql" "Manufacturing Keywords Implementation"; then
        print_status $RED "Manufacturing keywords implementation failed. Exiting."
        exit 1
    fi
    
    # Validate implementation
    if ! validate_implementation; then
        print_status $RED "Implementation validation failed. Exiting."
        exit 1
    fi
    
    # Run comprehensive tests
    if ! run_tests "$SCRIPT_DIR/test-manufacturing-keywords.sql" "Manufacturing Keywords Testing"; then
        print_status $RED "Manufacturing keywords testing failed. Exiting."
        exit 1
    fi
    
    # Generate summary report
    generate_summary_report
    
    print_section "TASK 3.2.5 COMPLETED SUCCESSFULLY"
    
    print_status $GREEN "✓ Manufacturing keywords implementation completed successfully"
    print_status $GREEN "✓ All 4 manufacturing industries have comprehensive keyword coverage"
    print_status $GREEN "✓ 200+ manufacturing keywords added with proper weight ranges"
    print_status $GREEN "✓ Comprehensive testing completed successfully"
    print_status $GREEN "✓ System ready for >85% manufacturing classification accuracy"
    
    print_status $BLUE "Next steps:"
    print_status $BLUE "1. Test manufacturing classification with real business data"
    print_status $BLUE "2. Monitor classification accuracy in production"
    print_status $BLUE "3. Proceed to next Task 3.2 subtask"
    
    print_status $BLUE "Log file saved to: $LOG_FILE"
}

# Run main function
main "$@"
