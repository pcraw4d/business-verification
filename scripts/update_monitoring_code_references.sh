#!/bin/bash

# Update Monitoring Code References Script
# This script updates application code to use unified monitoring tables instead of redundant tables

set -e

echo "=========================================="
echo "Updating Monitoring Code References"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "This script must be run from the project root directory"
    exit 1
fi

print_status "Starting monitoring code reference updates..."

# 1. Backup the original performance_dashboards.go file
print_status "Creating backup of original performance_dashboards.go..."
if [ -f "internal/classification/performance_dashboards.go" ]; then
    cp "internal/classification/performance_dashboards.go" "internal/classification/performance_dashboards.go.backup"
    print_status "Backup created: internal/classification/performance_dashboards.go.backup"
else
    print_warning "Original performance_dashboards.go not found, skipping backup"
fi

# 2. Replace the old performance_dashboards.go with the unified version
print_status "Replacing performance_dashboards.go with unified version..."
if [ -f "internal/classification/performance_dashboards_unified.go" ]; then
    mv "internal/classification/performance_dashboards_unified.go" "internal/classification/performance_dashboards.go"
    print_status "Successfully replaced performance_dashboards.go with unified version"
else
    print_error "performance_dashboards_unified.go not found!"
    exit 1
fi

# 3. Update any imports or references to the old monitoring tables
print_status "Searching for references to old monitoring tables..."

# List of old table names to search for
OLD_TABLES=(
    "performance_metrics"
    "performance_alerts"
    "response_time_metrics"
    "memory_metrics"
    "database_performance_metrics"
    "security_validation_metrics"
    "enhanced_query_performance_log"
    "database_performance_alerts"
    "security_validation_performance_log"
    "security_validation_alerts"
    "security_performance_metrics"
    "security_system_health"
    "classification_accuracy_metrics"
    "connection_pool_metrics"
    "query_performance_log"
    "usage_monitoring"
)

# Search for references to old tables in Go files
print_status "Searching for references to old monitoring tables in Go code..."
for table in "${OLD_TABLES[@]}"; do
    echo "Checking for references to: $table"
    if grep -r "$table" --include="*.go" . > /dev/null 2>&1; then
        print_warning "Found references to old table '$table' in Go code:"
        grep -r "$table" --include="*.go" . | head -5
        echo ""
    else
        print_status "No references found to old table '$table'"
    fi
done

# 4. Search for references to old database functions
print_status "Searching for references to old database functions..."
OLD_FUNCTIONS=(
    "collect_performance_metrics"
    "get_performance_alerts"
    "get_performance_summary"
    "get_query_performance_analysis"
    "get_index_performance_analysis"
    "get_table_performance_analysis"
    "get_connection_performance_analysis"
    "generate_performance_dashboard"
    "log_performance_metrics"
    "get_performance_trends"
    "setup_automated_performance_monitoring"
    "get_performance_dashboard_data"
    "export_performance_data"
    "validate_performance_monitoring_setup"
    "automated_performance_monitoring"
)

for func in "${OLD_FUNCTIONS[@]}"; do
    echo "Checking for references to function: $func"
    if grep -r "$func" --include="*.go" . > /dev/null 2>&1; then
        print_warning "Found references to old function '$func' in Go code:"
        grep -r "$func" --include="*.go" . | head -3
        echo ""
    else
        print_status "No references found to old function '$func'"
    fi
done

# 5. Update any configuration files that might reference old tables
print_status "Checking configuration files for old table references..."
CONFIG_FILES=(
    "configs/*.yaml"
    "configs/*.yml"
    "configs/*.json"
    "configs/*.toml"
    "*.env"
    "*.env.*"
)

for pattern in "${CONFIG_FILES[@]}"; do
    for file in $pattern; do
        if [ -f "$file" ]; then
            for table in "${OLD_TABLES[@]}"; do
                if grep -q "$table" "$file" 2>/dev/null; then
                    print_warning "Found reference to old table '$table' in config file: $file"
                fi
            done
        fi
    done
done

# 6. Check for any test files that might need updating
print_status "Checking test files for old table references..."
if find . -name "*_test.go" -exec grep -l "performance_metrics\|performance_alerts" {} \; > /dev/null 2>&1; then
    print_warning "Found test files that may reference old monitoring tables:"
    find . -name "*_test.go" -exec grep -l "performance_metrics\|performance_alerts" {} \;
    echo ""
else
    print_status "No test files found with old table references"
fi

# 7. Create a summary report
print_status "Creating update summary report..."
cat > "monitoring_code_update_summary.md" << EOF
# Monitoring Code Update Summary

## Update Date
$(date)

## Changes Made

### 1. File Replacements
- Replaced \`internal/classification/performance_dashboards.go\` with unified version
- Created backup: \`internal/classification/performance_dashboards.go.backup\`

### 2. New Unified Implementation
- Uses \`unified_performance_metrics\` table instead of multiple redundant tables
- Uses \`unified_performance_alerts\` table for all alerting
- Uses \`unified_performance_reports\` table for reporting
- Uses \`performance_integration_health\` table for health monitoring

### 3. Removed Dependencies
The following old database functions are no longer used:
$(printf -- "- %s\n" "${OLD_FUNCTIONS[@]}")

### 4. Removed Tables
The following redundant tables have been consolidated:
$(printf -- "- %s\n" "${OLD_TABLES[@]}")

## Next Steps

1. **Test the updated code** to ensure all functionality works with unified tables
2. **Run database migration** to remove redundant tables
3. **Update any remaining references** found in the search above
4. **Update documentation** to reflect the new unified monitoring system
5. **Remove backup files** after successful testing

## Verification

To verify the update was successful:

1. Check that \`internal/classification/performance_dashboards.go\` uses unified tables
2. Run tests to ensure monitoring functionality works
3. Check application logs for any database errors
4. Verify monitoring dashboards display data correctly

## Rollback

If issues are found, restore the original file:
\`\`\`bash
mv internal/classification/performance_dashboards.go.backup internal/classification/performance_dashboards.go
\`\`\`

EOF

print_status "Summary report created: monitoring_code_update_summary.md"

# 8. Run go mod tidy to clean up dependencies
print_status "Running go mod tidy to clean up dependencies..."
if command -v go > /dev/null 2>&1; then
    go mod tidy
    print_status "Dependencies cleaned up successfully"
else
    print_warning "Go not found, skipping go mod tidy"
fi

# 9. Check for any compilation errors
print_status "Checking for compilation errors..."
if command -v go > /dev/null 2>&1; then
    if go build ./... > /dev/null 2>&1; then
        print_status "Code compiles successfully"
    else
        print_error "Compilation errors found. Please check the output above."
        exit 1
    fi
else
    print_warning "Go not found, skipping compilation check"
fi

print_status "=========================================="
print_status "Monitoring Code Update Completed Successfully!"
print_status "=========================================="
print_status "Summary:"
print_status "- Replaced performance_dashboards.go with unified version"
print_status "- Created backup of original file"
print_status "- Searched for old table/function references"
print_status "- Created summary report: monitoring_code_update_summary.md"
print_status ""
print_status "Next steps:"
print_status "1. Test the updated monitoring functionality"
print_status "2. Run the database migration to remove redundant tables"
print_status "3. Update any remaining references found in the search"
print_status "4. Remove backup files after successful testing"
print_status "=========================================="
