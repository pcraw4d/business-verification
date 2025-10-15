#!/bin/bash

# OpenAPI Specification Generation Script
# This script generates OpenAPI specifications from Go code annotations

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$PROJECT_ROOT/api"
TEMP_DIR="/tmp/openapi_generation"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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

# Check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    if ! command -v swag &> /dev/null; then
        missing_deps+=("swag")
    fi
    
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Installing missing dependencies..."
        
        for dep in "${missing_deps[@]}"; do
            case $dep in
                "go")
                    log_error "Go is required but not installed. Please install Go first."
                    exit 1
                    ;;
                "swag")
                    log_info "Installing swag..."
                    go install github.com/swaggo/swag/cmd/swag@latest
                    ;;
                "jq")
                    if [[ "$OSTYPE" == "darwin"* ]]; then
                        brew install jq
                    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
                        sudo apt-get update && sudo apt-get install -y jq
                    else
                        log_error "Please install jq manually for your OS"
                        exit 1
                    fi
                    ;;
            esac
        done
    fi
    
    log_success "All dependencies are available"
}

# Clean up temporary files
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Set up trap for cleanup
trap cleanup EXIT

# Create temporary directory
setup_temp_dir() {
    log_info "Setting up temporary directory..."
    mkdir -p "$TEMP_DIR"
}

# Generate OpenAPI spec from Go code
generate_openapi_spec() {
    log_info "Generating OpenAPI specification from Go code..."
    
    cd "$PROJECT_ROOT"
    
    # Generate swagger docs
    log_info "Running swag init..."
    swag init -g cmd/server/main.go -o "$TEMP_DIR/swagger" --parseDependency --parseInternal
    
    if [ ! -f "$TEMP_DIR/swagger/swagger.json" ]; then
        log_error "Failed to generate swagger.json"
        exit 1
    fi
    
    log_success "Swagger specification generated"
}

# Enhance OpenAPI spec with additional information
enhance_openapi_spec() {
    log_info "Enhancing OpenAPI specification..."
    
    local swagger_file="$TEMP_DIR/swagger/swagger.json"
    local enhanced_file="$TEMP_DIR/openapi_enhanced.json"
    
    # Read the generated swagger file
    local swagger_content=$(cat "$swagger_file")
    
    # Enhance with additional information
    jq --arg version "$(git describe --tags --always)" \
       --arg build_date "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
       --arg git_commit "$(git rev-parse HEAD)" \
       --arg server_url "https://api.kyb-platform.com" \
       '. + {
           "info": (.info + {
               "version": $version,
               "x-build-date": $build_date,
               "x-git-commit": $git_commit,
               "x-generated-by": "automated-pipeline"
           }),
           "servers": [
               {
                   "url": $server_url,
                   "description": "Production server"
               },
               {
                   "url": "https://staging-api.kyb-platform.com",
                   "description": "Staging server"
               },
               {
                   "url": "http://localhost:8080",
                   "description": "Local development server"
               }
           ],
           "x-codegen": {
               "generated": true,
               "generated_at": $build_date,
               "generator_version": "1.0.0"
           }
       }' "$swagger_file" > "$enhanced_file"
    
    log_success "OpenAPI specification enhanced"
}

# Validate OpenAPI specification
validate_openapi_spec() {
    log_info "Validating OpenAPI specification..."
    
    local spec_file="$TEMP_DIR/openapi_enhanced.json"
    
    # Basic JSON validation
    if ! jq empty "$spec_file" 2>/dev/null; then
        log_error "Invalid JSON in OpenAPI specification"
        exit 1
    fi
    
    # Check required fields
    local required_fields=("openapi" "info" "paths" "components")
    for field in "${required_fields[@]}"; do
        if ! jq -e ".$field" "$spec_file" > /dev/null; then
            log_error "Missing required field: $field"
            exit 1
        fi
    done
    
    # Validate OpenAPI version
    local openapi_version=$(jq -r '.openapi' "$spec_file")
    if [[ ! "$openapi_version" =~ ^3\.0\.[0-9]+$ ]]; then
        log_warning "OpenAPI version $openapi_version may not be supported by all tools"
    fi
    
    log_success "OpenAPI specification validation passed"
}

# Generate additional documentation files
generate_documentation_files() {
    log_info "Generating additional documentation files..."
    
    local spec_file="$TEMP_DIR/openapi_enhanced.json"
    local output_dir="$OUTPUT_DIR"
    
    # Create output directory
    mkdir -p "$output_dir"
    
    # Copy enhanced OpenAPI spec
    cp "$spec_file" "$output_dir/openapi.json"
    cp "$spec_file" "$output_dir/openapi.yaml"
    
    # Convert JSON to YAML
    if command -v yq &> /dev/null; then
        yq eval -P "$spec_file" > "$output_dir/openapi.yaml"
    else
        log_warning "yq not found, skipping YAML conversion"
    fi
    
    # Generate API summary
    generate_api_summary "$spec_file" "$output_dir"
    
    # Generate endpoint documentation
    generate_endpoint_docs "$spec_file" "$output_dir"
    
    log_success "Additional documentation files generated"
}

# Generate API summary
generate_api_summary() {
    local spec_file="$1"
    local output_dir="$2"
    
    log_info "Generating API summary..."
    
    local summary_file="$output_dir/API_SUMMARY.md"
    
    cat > "$summary_file" << EOF
# API Summary

Generated on: $(date -u +%Y-%m-%dT%H:%M:%SZ)
OpenAPI Version: $(jq -r '.openapi' "$spec_file")
API Version: $(jq -r '.info.version' "$spec_file")

## Overview

$(jq -r '.info.description' "$spec_file")

## Statistics

- **Total Endpoints**: $(jq '.paths | keys | length' "$spec_file")
- **GET Endpoints**: $(jq '[.paths[] | keys[] | select(. == "get")] | length' "$spec_file")
- **POST Endpoints**: $(jq '[.paths[] | keys[] | select(. == "post")] | length' "$spec_file")
- **PUT Endpoints**: $(jq '[.paths[] | keys[] | select(. == "put")] | length' "$spec_file")
- **DELETE Endpoints**: $(jq '[.paths[] | keys[] | select(. == "delete")] | length' "$spec_file")

## Endpoints

$(jq -r '.paths | keys[]' "$spec_file" | sed 's/^/- /')

## Components

### Schemas
$(jq -r '.components.schemas | keys[]' "$spec_file" | sed 's/^/- /')

### Security Schemes
$(jq -r '.components.securitySchemes | keys[]' "$spec_file" | sed 's/^/- /')

EOF
    
    log_success "API summary generated"
}

# Generate endpoint documentation
generate_endpoint_docs() {
    local spec_file="$1"
    local output_dir="$2"
    
    log_info "Generating endpoint documentation..."
    
    local endpoints_file="$output_dir/ENDPOINTS.md"
    
    cat > "$endpoints_file" << EOF
# API Endpoints

Generated on: $(date -u +%Y-%m-%dT%H:%M:%SZ)

EOF
    
    # Generate documentation for each endpoint
    jq -r '.paths | to_entries[] | "\(.key) \(.value | keys[])"' "$spec_file" | while read -r path method; do
        echo "## $method $path" >> "$endpoints_file"
        echo "" >> "$endpoints_file"
        
        # Get endpoint details
        local summary=$(jq -r ".paths[\"$path\"][\"$method\"].summary // \"No summary\"" "$spec_file")
        local description=$(jq -r ".paths[\"$path\"][\"$method\"].description // \"No description\"" "$spec_file")
        
        echo "**Summary**: $summary" >> "$endpoints_file"
        echo "" >> "$endpoints_file"
        echo "**Description**: $description" >> "$endpoints_file"
        echo "" >> "$endpoints_file"
        
        # Get parameters
        local params=$(jq -r ".paths[\"$path\"][\"$method\"].parameters // []" "$spec_file")
        if [ "$params" != "[]" ]; then
            echo "**Parameters**:"
            echo "" >> "$endpoints_file"
            echo "$params" | jq -r '.[] | "- \(.name) (\(.in)): \(.description // "No description")"' >> "$endpoints_file"
            echo "" >> "$endpoints_file"
        fi
        
        # Get request body
        local request_body=$(jq -r ".paths[\"$path\"][\"$method\"].requestBody // null" "$spec_file")
        if [ "$request_body" != "null" ]; then
            echo "**Request Body**:"
            echo "" >> "$endpoints_file"
            echo "$request_body" | jq -r '.content."application/json".schema."$ref" // "application/json"' >> "$endpoints_file"
            echo "" >> "$endpoints_file"
        fi
        
        # Get responses
        echo "**Responses**:"
        echo "" >> "$endpoints_file"
        jq -r ".paths[\"$path\"][\"$method\"].responses | to_entries[] | \"- \(.key): \(.value.description // \"No description\")\"" "$spec_file" >> "$endpoints_file"
        echo "" >> "$endpoints_file"
        echo "---" >> "$endpoints_file"
        echo "" >> "$endpoints_file"
    done
    
    log_success "Endpoint documentation generated"
}

# Generate SDK documentation
generate_sdk_docs() {
    log_info "Generating SDK documentation..."
    
    local spec_file="$TEMP_DIR/openapi_enhanced.json"
    local sdk_docs_dir="$PROJECT_ROOT/docs/sdks"
    
    mkdir -p "$sdk_docs_dir"
    
    # Generate SDK documentation for each language
    for lang in go python nodejs ruby java php; do
        generate_sdk_doc "$spec_file" "$sdk_docs_dir" "$lang"
    done
    
    log_success "SDK documentation generated"
}

# Generate SDK documentation for a specific language
generate_sdk_doc() {
    local spec_file="$1"
    local output_dir="$2"
    local language="$3"
    
    local doc_file="$output_dir/${language}_sdk.md"
    
    cat > "$doc_file" << EOF
# ${language^} SDK Documentation

Generated on: $(date -u +%Y-%m-%dT%H:%M:%SZ)

## Installation

\`\`\`bash
# Installation instructions for $language
\`\`\`

## Quick Start

\`\`\`$language
# Quick start example for $language
\`\`\`

## API Reference

### Authentication

\`\`\`$language
# Authentication example
\`\`\`

### Risk Assessment

\`\`\`$language
# Risk assessment example
\`\`\`

### Error Handling

\`\`\`$language
# Error handling example
\`\`\`

EOF
    
    log_info "Generated SDK documentation for $language"
}

# Update version information
update_version_info() {
    log_info "Updating version information..."
    
    local version_file="$PROJECT_ROOT/VERSION"
    local version=$(git describe --tags --always)
    
    echo "$version" > "$version_file"
    
    log_success "Version information updated: $version"
}

# Main execution
main() {
    log_info "Starting OpenAPI specification generation..."
    
    check_dependencies
    setup_temp_dir
    generate_openapi_spec
    enhance_openapi_spec
    validate_openapi_spec
    generate_documentation_files
    generate_sdk_docs
    update_version_info
    
    log_success "OpenAPI specification generation completed successfully!"
    log_info "Generated files:"
    log_info "  - $OUTPUT_DIR/openapi.json"
    log_info "  - $OUTPUT_DIR/openapi.yaml"
    log_info "  - $OUTPUT_DIR/API_SUMMARY.md"
    log_info "  - $OUTPUT_DIR/ENDPOINTS.md"
    log_info "  - $PROJECT_ROOT/docs/sdks/"
}

# Run main function
main "$@"
