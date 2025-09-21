#!/bin/bash

# Data Type and Format Validation Script
# This script runs comprehensive data type and format validation tests

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$PROJECT_ROOT/test-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to validate database connection
validate_db_connection() {
    local db_url=$1
    print_status $BLUE "🔍 Validating database connection..."
    
    if command_exists psql; then
        if psql "$db_url" -c "SELECT 1;" >/dev/null 2>&1; then
            print_status $GREEN "✅ Database connection successful"
            return 0
        else
            print_status $RED "❌ Database connection failed"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  psql not found, skipping connection validation"
        return 0
    fi
}

# Function to run SQL tests
run_sql_tests() {
    local db_url=$1
    local output_file="$OUTPUT_DIR/data_type_sql_test_$TIMESTAMP.log"
    
    print_status $BLUE "📊 Running SQL-based data type validation tests..."
    
    if command_exists psql; then
        print_status $BLUE "   Executing SQL test script..."
        if psql "$db_url" -f "$SCRIPT_DIR/validate-data-types.sql" > "$output_file" 2>&1; then
            print_status $GREEN "✅ SQL tests completed successfully"
            print_status $BLUE "   Results saved to: $output_file"
            
            # Check for any failures in the output
            if grep -q "invalid.*count.*[1-9]" "$output_file"; then
                print_status $YELLOW "⚠️  Some data type issues detected - check results"
            fi
            
            return 0
        else
            print_status $RED "❌ SQL tests failed"
            print_status $RED "   Check error log: $output_file"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  psql not found, skipping SQL tests"
        return 0
    fi
}

# Function to run Go tests
run_go_tests() {
    local db_url=$1
    local output_file="$OUTPUT_DIR/data_type_go_test_$TIMESTAMP.log"
    
    print_status $BLUE "🔧 Running Go-based data type validation tests..."
    
    if command_exists go; then
        print_status $BLUE "   Compiling and running Go test program..."
        
        # Change to script directory
        cd "$SCRIPT_DIR"
        
        # Compile the Go program
        if go build -o validate-data-types validate-data-types.go; then
            print_status $GREEN "✅ Go program compiled successfully"
            
            # Run the test
            if ./validate-data-types "$db_url" > "$output_file" 2>&1; then
                print_status $GREEN "✅ Go tests completed successfully"
                print_status $BLUE "   Results saved to: $output_file"
                return 0
            else
                print_status $RED "❌ Go tests failed"
                print_status $RED "   Check error log: $output_file"
                return 1
            fi
        else
            print_status $RED "❌ Failed to compile Go program"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  Go not found, skipping Go tests"
        return 0
    fi
}

# Function to generate summary report
generate_summary_report() {
    local output_file="$OUTPUT_DIR/data_type_test_summary_$TIMESTAMP.md"
    
    print_status $BLUE "📋 Generating summary report..."
    
    cat > "$output_file" << EOF
# Data Type and Format Validation Test Summary

**Test Date:** $(date)
**Timestamp:** $TIMESTAMP

## Test Results

### SQL Tests
EOF

    if [ -f "$OUTPUT_DIR/data_type_sql_test_$TIMESTAMP.log" ]; then
        echo "✅ SQL tests completed" >> "$output_file"
        echo "📁 Results: \`data_type_sql_test_$TIMESTAMP.log\`" >> "$output_file"
    else
        echo "⚠️ SQL tests skipped (psql not available)" >> "$output_file"
    fi

    cat >> "$output_file" << EOF

### Go Tests
EOF

    if [ -f "$OUTPUT_DIR/data_type_go_test_$TIMESTAMP.log" ]; then
        echo "✅ Go tests completed" >> "$output_file"
        echo "📁 Results: \`data_type_go_test_$TIMESTAMP.log\`" >> "$output_file"
    else
        echo "⚠️ Go tests skipped (Go not available)" >> "$output_file"
    fi

    cat >> "$output_file" << EOF

## What Was Tested

### Data Type Validation
- **Email Format**: Validates email addresses using RFC-compliant regex
- **UUID Format**: Validates UUID format for ID columns
- **Phone Format**: Validates phone numbers in E.164 format
- **URL Format**: Validates website URLs
- **Date Format**: Validates date and timestamp formats
- **String Length**: Checks varchar columns against length constraints
- **Numeric Ranges**: Validates numeric data types
- **Boolean Values**: Validates boolean columns
- **JSON Format**: Validates JSON and JSONB columns
- **NULL Constraints**: Checks for NULL values in non-nullable columns

### Format Validation Patterns
- **Email**: \`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$\`
- **UUID**: \`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$\`
- **Phone**: \`^\+?[1-9]\d{1,14}$\`
- **URL**: \`^https?://[^\s/$.?#].[^\s]*$\`
- **Date**: \`^\d{4}-\d{2}-\d{2}$\`
- **Timestamp**: \`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\`

## Next Steps

1. Review the test results in the log files
2. Address any data type or format issues found
3. Fix any constraint violations
4. Re-run tests to verify fixes

## Files Generated

- \`data_type_sql_test_$TIMESTAMP.log\` - SQL test results
- \`data_type_go_test_$TIMESTAMP.log\` - Go test results
- \`data_type_test_summary_$TIMESTAMP.md\` - This summary report

EOF

    print_status $GREEN "✅ Summary report generated: $output_file"
}

# Function to display usage
usage() {
    echo "Usage: $0 <database_url>"
    echo ""
    echo "Arguments:"
    echo "  database_url    PostgreSQL connection string"
    echo ""
    echo "Examples:"
    echo "  $0 'postgresql://user:pass@localhost:5432/dbname'"
    echo "  $0 'postgresql://user:pass@localhost:5432/dbname?sslmode=require'"
    echo ""
    echo "Environment Variables:"
    echo "  DATABASE_URL    Alternative way to specify database URL"
}

# Main execution
main() {
    print_status $BLUE "🚀 Starting Data Type and Format Validation Testing"
    print_status $BLUE "=================================================="
    
    # Get database URL
    local db_url=""
    if [ $# -eq 1 ]; then
        db_url="$1"
    elif [ -n "$DATABASE_URL" ]; then
        db_url="$DATABASE_URL"
    else
        print_status $RED "❌ Database URL required"
        usage
        exit 1
    fi
    
    print_status $BLUE "📊 Database URL: ${db_url%%:*}:****"
    print_status $BLUE "📁 Output directory: $OUTPUT_DIR"
    print_status $BLUE "🕐 Timestamp: $TIMESTAMP"
    echo ""
    
    # Validate database connection
    if ! validate_db_connection "$db_url"; then
        exit 1
    fi
    echo ""
    
    # Run tests
    local sql_result=0
    local go_result=0
    
    run_sql_tests "$db_url" || sql_result=1
    echo ""
    
    run_go_tests "$db_url" || go_result=1
    echo ""
    
    # Generate summary
    generate_summary_report
    echo ""
    
    # Final status
    if [ $sql_result -eq 0 ] && [ $go_result -eq 0 ]; then
        print_status $GREEN "🎉 All data type and format validation tests completed successfully!"
        print_status $BLUE "📁 Check the test results in: $OUTPUT_DIR"
        exit 0
    else
        print_status $YELLOW "⚠️  Some tests failed - check the logs for details"
        print_status $BLUE "📁 Check the test results in: $OUTPUT_DIR"
        exit 1
    fi
}

# Run main function
main "$@"
