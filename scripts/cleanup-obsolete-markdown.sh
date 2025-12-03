#!/bin/bash

# Cleanup Obsolete Markdown Files Script
# Removes obsolete markdown files while preserving important documentation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DRY_RUN=false
VERBOSE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--dry-run] [--verbose]"
            exit 1
            ;;
    esac
done

# Function to log messages
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Files to preserve (important documentation)
PRESERVE_FILES=(
    "README.md"
    "CONTRIBUTING.md"
    "DATABASE_SETUP_GUIDE.md"
    "DEVELOPMENT_WORKFLOW_GUIDE.md"
    "MANUAL_TESTING_PROCEDURES.md"
    "KYB_PLATFORM_COMPREHENSIVE_DOCUMENTATION.md"
    "RAILWAY_SETUP_GUIDE.md"
    "RAILWAY_DEPLOYMENT_CHECKLIST.md"
    "RAILWAY_ENVIRONMENT_VARIABLES.md"
    "REAL_DATA_INTEGRATION_GUIDE.md"
    "supabase-setup-guide.md"
    "SUPABASE_DATABASE_SETUP_INSTRUCTIONS.md"
    "disaster_recovery_design.md"
    "multi_tenant_architecture_design.md"
    "global_deployment_strategy.md"
    "high_volume_processing_plan.md"
    "risk_keywords_documentation.md"
    "database-schema-keyword-classification.md"
    "data_volume_assessment_report.md"
    "redundancy_analysis.md"
    "unified_monitoring_schema_design.md"
    "industry_coverage_dashboard.md"
    "industry_coverage_implementation_plan.md"
)

# Function to check if file should be preserved
should_preserve() {
    local file="$1"
    local basename=$(basename "$file")
    
    # Check preserve list
    for preserve in "${PRESERVE_FILES[@]}"; do
        if [[ "$basename" == "$preserve" ]]; then
            return 0
        fi
    done
    
    # Preserve files in docs directory (they're organized)
    if [[ "$file" == *"/docs/"* ]]; then
        return 0
    fi
    
    return 1
}

# Function to remove files matching patterns
remove_files_by_pattern() {
    local description="$1"
    shift
    local patterns=("$@")
    local count=0
    
    log "Processing: $description"
    
    while IFS= read -r file; do
        # Skip if should be preserved
        if should_preserve "$file"; then
            if [ "$VERBOSE" = true ]; then
                warn "Preserving: $(basename "$file")"
            fi
            continue
        fi
        
        if [ "$VERBOSE" = true ]; then
            echo "  - $(basename "$file")"
        fi
        
        if [ "$DRY_RUN" = false ]; then
            rm -f "$file"
        fi
        ((count++))
    done < <(find "$REPO_ROOT" -maxdepth 1 -name "*.md" -type f 2>/dev/null | while read -r file; do
        basename_file=$(basename "$file")
        for pattern in "${patterns[@]}"; do
            # Convert glob to regex
            regex=$(echo "$pattern" | sed 's/\*/.*/g' | sed 's/\.md$//')
            if echo "$basename_file" | grep -qiE "$regex"; then
                echo "$file"
                break
            fi
        done
    done)
    
    if [ "$DRY_RUN" = true ]; then
        log "Would remove $count file(s)"
    else
        log "Removed $count file(s)"
    fi
}

# Main cleanup function
main() {
    log "Starting obsolete markdown files cleanup"
    if [ "$DRY_RUN" = true ]; then
        warn "DRY RUN MODE - No files will be deleted"
    fi
    echo ""
    
    # Remove obsolete files by category
    remove_files_by_pattern "Old status reports" "*STATUS*.md" "*STATUS_REPORT*.md" "*COMPLETION_STATUS*.md"
    remove_files_by_pattern "Old test results" "*TEST_RESULTS*.md" "*TEST_REPORT*.md" "PHASE_*_TEST*.md" "*TESTING_RESULTS*.md"
    remove_files_by_pattern "Old investigation reports" "*INVESTIGATION*.md" "*ROOT_CAUSE*.md" "*FINDINGS*.md"
    remove_files_by_pattern "Old fix plans" "*FIX_PLAN*.md" "*_FIX.md" "*FIXES*.md" "*REMEDIATION*.md"
    remove_files_by_pattern "Old deployment reports" "DEPLOYMENT_*.md" "*DEPLOYMENT*.md"
    remove_files_by_pattern "Old phase documents" "PHASE_*.md" "phase_*.md"
    remove_files_by_pattern "Old error reports" "ERROR_*.md" "*ERROR*.md"
    remove_files_by_pattern "Old verification reports" "*VERIFICATION*.md" "*VERIFICATION_REPORT*.md"
    remove_files_by_pattern "Old test execution reports" "*TEST_EXECUTION*.md" "*EXECUTION_RESULTS*.md"
    remove_files_by_pattern "Old completion reports" "*COMPLETION*.md" "*COMPLETE*.md"
    remove_files_by_pattern "Old reflection documents" "*reflection*.md" "*REFLECTION*.md"
    remove_files_by_pattern "Old analysis reports" "*ANALYSIS*.md" "*ANALYSIS_REPORT*.md"
    
    # Specific obsolete files
    log "Processing: Specific obsolete files"
    local specific_files=(
        "ADD_MERCHANT_FLOW_FIX_PLAN.md"
        "ADD_MERCHANT_REDIRECT_TEST_RESULTS.md"
        "ADD_MERCHANT_TO_DETAILS_FLOW_ISSUES.md"
        "ADD_MERCHANT_TO_DETAILS_FLOW_TEST_REPORT.md"
        "advanced_feature_planning_6_2_2.md"
        "API_GATEWAY_DEPLOYMENT_ISSUE.md"
        "BETA_TEST_REPORT.md"
        "BETA_TESTING_READY.md"
        "BETA_TESTING_VERIFICATION.md"
        "BI_SERVICE_INVESTIGATION.md"
        "BOT_EVASION_REVIEW.md"
        "BROWSER_TEST_FINDINGS_AND_FIX_PLAN.md"
        "BROWSER_TEST_REPORT_AFTER_FIXES.md"
        "BROWSER_TEST_REPORT_AND_FIX_PLAN.md"
        "BUSINESS_ANALYTICS_CONTENT_VISIBILITY_ROOT_CAUSE.md"
        "CLEANUP_SUCCESS_REPORT.md"
        "CODEBASE_IMPROVEMENT_PLAN_COMPLETION_ANALYSIS.md"
        "codebase_issues_fix_plan.md"
        "COMPLETE_DEPLOYMENT_SUCCESS_REPORT.md"
        "COMPLETE_RAILWAY_SERVICES_INVENTORY.md"
        "COMPLETE_RENDERING_ISSUE_ROOT_CAUSE.md"
        "COMPREHENSIVE_BUILD_REVIEW_PLAN.md"
        "COMPREHENSIVE_CLASSIFICATION_IMPROVEMENT_PLAN_REORGANIZED.md"
        "COMPREHENSIVE_CLASSIFICATION_IMPROVEMENT_PLAN.md"
        "COMPREHENSIVE_FEATURE_ANALYSIS.md"
        "COMPREHENSIVE_REVIEW_COMPLETION_STATUS.md"
        "COMPREHENSIVE_TESTING_RESULTS.md"
        "CRITICAL_FEATURES_IMPLEMENTED.md"
        "CSS_DIAGNOSTIC_TESTING_INSTRUCTIONS.md"
        "DEPLOYMENT_FIX_VERIFICATION.md"
        "DEPLOYMENT_FIXES_STATUS_REPORT.md"
        "DEPLOYMENT_FIXES.md"
        "DEPLOYMENT_INSTRUCTIONS.md"
        "DEPLOYMENT_INVESTIGATION_REPORT.md"
        "DEPLOYMENT_RESULT.md"
        "DEPLOYMENT_STATUS_AND_NEXT_STEPS.md"
        "DEPLOYMENT_STATUS_REPORT.md"
        "DEPLOYMENT_STATUS.md"
        "DEPLOYMENT_STRUCTURE_REFERENCE.md"
        "DEPLOYMENT_SUCCESS_REPORT.md"
        "DEPLOYMENT_SUCCESS.md"
        "DEPLOYMENT_TEST_RESULTS.md"
        "DEPLOYMENT_TESTING_RESULTS.md"
        "DEPLOYMENT_VERIFICATION.md"
        "DIAGNOSTIC_FUNCTION_DEPLOYMENT_STATUS.md"
        "DIAGNOSTIC_OUTPUT_ANALYSIS.md"
        "DOCKERFILE_ISSUE_INVESTIGATION_REPORT.md"
        "E2E_TEST_FIXES_APPLIED.md"
        "E2E_TEST_FIXES_ROUND2.md"
        "E2E_TEST_FIXES_ROUND3.md"
        "E2E_TEST_FIXES_ROUND4.md"
        "E2E_TEST_FIXES_ROUND5.md"
        "E2E_TEST_FIXES_ROUND6.md"
        "ERROR_RESOLUTION_FINAL_STATUS.md"
        "ERROR_RESOLUTION_VERIFICATION.md"
        "ERROR_VERIFICATION_FINAL_REPORT.md"
        "EXPERIMENTAL_SERVICES_CLEANUP_AND_REDIS_ANALYSIS.md"
        "FINAL_DEPLOYMENT_CONFIRMATION.md"
        "FINAL_DEPLOYMENT_STATUS_AND_LEGACY_CLEANUP_RECOMMENDATIONS.md"
        "FINAL_DEPLOYMENT_SUCCESS_REPORT.md"
        "FINAL_ERROR_VERIFICATION.md"
        "FINAL_TEST_EXECUTION_RESULTS.md"
        "FINAL_TESTING_RESULTS.md"
        "FINAL_VERIFICATION_RESULTS.md"
        "FRONTEND_404_INVESTIGATION_REPORT.md"
        "FRONTEND_BACKEND_GAP_IMPLEMENTATION_PLAN.md"
        "frontend_error_review.md"
        "FUNCTIONALITY_TEST_RESULTS.md"
        "IMPLEMENTATION_REVIEW.md"
        "INTEGRATION_TEST_EXECUTION_RESULTS.md"
        "MERCHANT_DETAILS_ERRORS_ANALYSIS.md"
        "MERCHANT_DETAILS_FEATURE_MATRIX.md"
        "MERCHANT_DETAILS_RENDERING_FIX.md"
        "MERCHANT_DETAILS_RENDERING_ISSUE_FIX_PLAN.md"
        "merchant_details_resolution_plan.md"
        "MERCHANT_DETAILS_UI_TEST_REPORT.md"
        "MERCHANT_FORM_ISSUE_ANALYSIS.md"
        "MERCHANT_SERVICE_BUILD_FIX.md"
        "MIGRATION_EXECUTION_COMPLETE.md"
        "MIGRATION_PLAN_RESTORE_FUNCTIONALITY.md"
        "MIGRATION_QUICK_REFERENCE.md"
        "MIGRATION_STATUS.md"
        "monitoring_table_analysis.md"
        "MUTATION_OBSERVER_INFINITE_LOOP_FIX_PLAN.md"
        "MUTATION_OBSERVER_INFINITE_LOOP_FIX.md"
        "NAVIGATION_PAGEMAP_FIX.md"
        "NEXT_STEPS.md"
        "performance_testing_completion.md"
        "POST_MIGRATION_TEST_FINAL_RESULTS.md"
        "POST_MIGRATION_TEST_RESULTS.md"
        "PRODUCTION_DEPLOYMENT_READY.md"
        "PRODUCTION_DEPLOYMENT_STATUS_REPORT.md"
        "PRODUCTION_OPTIMIZATION_IMPLEMENTATION_PLAN.md"
        "PRODUCTION_READINESS_PLAN.md"
        "PROGRESSIVE_DISCLOSURE_DEBUG_ANALYSIS.md"
        "RAILWAY_ARCHITECTURE_REVIEW_REPORT.md"
        "RAILWAY_CLI_AUTHENTICATION.md"
        "RAILWAY_DASHBOARD_ANALYSIS_AND_CLEANUP_PLAN.md"
        "RAILWAY_DEPLOYMENT_FIX_COMPLETION.md"
        "RAILWAY_DEPLOYMENT_FIX_FRONTEND.md"
        "RAILWAY_DEPLOYMENT_FIX.md"
        "RAILWAY_DEPLOYMENT_STATUS_REPORT.md"
        "RAILWAY_DEPLOYMENT_TEST_REPORT.md"
        "RAILWAY_DEPLOYMENT_TESTING_GUIDE.md"
        "RAILWAY_DEPLOYMENT_VERIFICATION_REPORT.md"
        "RAILWAY_LOG_ANALYSIS.md"
        "RAILWAY_VS_LOCALHOST_TESTING_ANALYSIS.md"
        "REFRESH_BUTTON_IMPLEMENTATION.md"
        "REI_403_FINAL_ANALYSIS.md"
        "REI_403_FINDINGS_AND_RECOMMENDATIONS.md"
        "REMAINING_ERRORS_RESOLUTION.md"
        "REMEDIATION_PROGRESS.md"
        "RENDERING_ISSUE_ROOT_CAUSE_FIX.md"
        "RESTORATION_COMPLETE.md"
        "RETEST_RESULTS.md"
        "ROOT_CAUSE_ANALYSIS_AND_NEXT_STEPS.md"
        "SECURITY_IMPROVEMENTS_CLASSIFICATION.md"
        "SERVICE_DEPLOYMENT_AUDIT.md"
        "SESSION_MANAGER_ROOT_CAUSE_FIX.md"
        "SMART_CRAWLER_403_FIX.md"
        "SMART_CRAWLER_BOT_EVASION_TEST_RESULTS.md"
        "subtask_1_3_1_industry_coverage_analysis.md"
        "SUPABASE_CONNECTION_STATUS.md"
        "SUPABASE_CONNECTION_VERIFICATION.md"
        "SUPABASE_TABLE_IMPROVEMENT_IMPLEMENTATION_PLAN.md"
        "SUPABASE_TABLE_RECOVERY_PLAN.md"
        "TAB_SWITCHING_ANALYSIS.md"
        "TAB_SWITCHING_CRITICAL_FINDING.md"
        "TAB_SWITCHING_FIX_APPLIED.md"
        "TAB_SWITCHING_FIX_PLAN.md"
        "TAB_SWITCHING_INVESTIGATION_CONTINUED.md"
        "TAB_SWITCHING_INVESTIGATION_REPORT.md"
        "TAB_SWITCHING_ISSUE_ANALYSIS.md"
        "TAB_SWITCHING_ISSUE_INVESTIGATION.md"
        "TAB_SWITCHING_LOG_ANALYSIS.md"
        "TAB_SWITCHING_ROOT_CAUSE_ANALYSIS.md"
        "TAB_SWITCHING_TEST_AFTER_PROGRESSIVE_DISCLOSURE_FIX.md"
        "TAB_SWITCHING_TEST_REPORT.md"
        "TAB_SWITCHING_TEST_RESULTS.md"
        "TASK_1_1_1_EXECUTION_INSTRUCTIONS.md"
        "task_3_2_industry_analysis.md"
        "task1.11_observability_compilation_fixes_final_progress.md"
        "TECHNICAL_DEBT_ANALYSIS_AND_CLEANUP_PLAN.md"
        "TEST_EXECUTION_COMPLETE.md"
        "TEST_FIXES_APPLIED.md"
        "TEST_FIXES_BULK_OPERATIONS_STATUS.md"
        "TEST_FIXES_COLLAPSIBLE_SECTIONS.md"
        "TEST_FIXES_COMPREHENSIVE.md"
        "TEST_FIXES_FORM_SUBMISSION_DEBUG.md"
        "TEST_FIXES_MERCHANT_FORM_FINAL.md"
        "TEST_FIXES_MERCHANT_FORM_STATUS_FINAL.md"
        "TEST_FIXES_MERCHANT_FORM_STATUS.md"
        "TEST_FIXES_PROGRESS_FINAL.md"
        "TEST_FIXES_PROGRESS_REPORT.md"
        "TEST_FIXES_PROGRESS_UPDATE_2.md"
        "TEST_FIXES_PROGRESS_UPDATE.md"
        "TEST_REPORT_AND_ISSUES.md"
        "TEST_SETUP_COMPLETE.md"
        "TEST_SUITE_AND_COVERAGE_REPORT.md"
        "TESTING_CHECKLIST_RESULTS.md"
        "TESTING_PROCEDURES_MARKET_ANALYSIS_INTERFACE.md"
        "TRAILING_SLASH_FIX.md"
        "UI_FLOW_ANALYSIS.md"
        "UI_FLOW_DIAGRAMS.md"
        "UI_FUNCTIONALITY_TEST_REPORT.md"
        "WEBSITE_KEYWORD_EXTRACTION_ANALYSIS.md"
        "WEBSITE_KEYWORD_EXTRACTION_SOLUTION.md"
        "WEIGHTED_CLASSIFICATION_SYSTEM_IMPROVEMENTS.md"
        "emerging_trends_analysis_2025-09-19.md"
        "industry_coverage_analysis_2025-09-19.md"
        "industry_taxonomy_hierarchy_2025-09-19.md"
    )
    
    local count=0
    for file in "${specific_files[@]}"; do
        local full_path="$REPO_ROOT/$file"
        if [ -f "$full_path" ]; then
            if should_preserve "$full_path"; then
                if [ "$VERBOSE" = true ]; then
                    warn "Preserving: $file"
                fi
                continue
            fi
            
            if [ "$VERBOSE" = true ]; then
                echo "  - $file"
            fi
            
            if [ "$DRY_RUN" = false ]; then
                rm -f "$full_path"
            fi
            ((count++))
        fi
    done
    
    if [ "$DRY_RUN" = true ]; then
        log "Would remove $count specific file(s)"
    else
        log "Removed $count specific file(s)"
    fi
    
    echo ""
    log "=== Cleanup Complete ==="
    if [ "$DRY_RUN" = true ]; then
        warn "This was a DRY RUN. No files were actually deleted."
        warn "Run without --dry-run to perform the actual cleanup."
    else
        log "Obsolete markdown files cleanup completed successfully!"
    fi
}

# Run main function
main

