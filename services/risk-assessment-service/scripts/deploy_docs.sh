#!/bin/bash

# Documentation Deployment Script
# This script deploys documentation to various platforms

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCS_DIR="$PROJECT_ROOT/docs"
BUILD_DIR="$PROJECT_ROOT/docs-site"
TEMP_DIR="/tmp/docs_deployment"

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
    
    if ! command -v node &> /dev/null; then
        missing_deps+=("node")
    fi
    
    if ! command -v npm &> /dev/null; then
        missing_deps+=("npm")
    fi
    
    if ! command -v git &> /dev/null; then
        missing_deps+=("git")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
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

# Build documentation site
build_docs_site() {
    log_info "Building documentation site..."
    
    # Create build directory
    mkdir -p "$BUILD_DIR"
    
    # Initialize Docusaurus if not exists
    if [ ! -f "$BUILD_DIR/package.json" ]; then
        log_info "Initializing Docusaurus..."
        cd "$BUILD_DIR"
        npx create-docusaurus@latest . classic --typescript
        cd "$PROJECT_ROOT"
    fi
    
    # Copy documentation files
    log_info "Copying documentation files..."
    cp -r "$DOCS_DIR"/* "$BUILD_DIR/docs/"
    
    # Generate API documentation
    generate_api_docs
    
    # Build the site
    log_info "Building Docusaurus site..."
    cd "$BUILD_DIR"
    npm run build
    cd "$PROJECT_ROOT"
    
    log_success "Documentation site built successfully"
}

# Generate API documentation
generate_api_docs() {
    log_info "Generating API documentation..."
    
    local api_docs_dir="$BUILD_DIR/docs/api"
    mkdir -p "$api_docs_dir"
    
    # Copy OpenAPI specs
    if [ -f "$PROJECT_ROOT/api/openapi.yaml" ]; then
        cp "$PROJECT_ROOT/api/openapi.yaml" "$api_docs_dir/"
        cp "$PROJECT_ROOT/api/openapi.json" "$api_docs_dir/"
    fi
    
    # Generate API reference
    cat > "$api_docs_dir/README.md" << EOF
# API Reference

This section contains the complete API reference for the Risk Assessment Service.

## OpenAPI Specification

- [OpenAPI YAML](openapi.yaml)
- [OpenAPI JSON](openapi.json)

## Interactive Documentation

You can explore the API interactively using the embedded Swagger UI below:

<iframe src="/swagger-ui/" width="100%" height="800px" frameborder="0"></iframe>

## SDK Documentation

- [Go SDK](sdks/go-sdk)
- [Python SDK](sdks/python-sdk)
- [Node.js SDK](sdks/nodejs-sdk)
- [Ruby SDK](sdks/ruby-sdk)
- [Java SDK](sdks/java-sdk)
- [PHP SDK](sdks/php-sdk)

## Authentication

All API requests require authentication using JWT tokens. See the [Authentication Guide](authentication) for details.

## Rate Limiting

API requests are rate limited. See the [Rate Limiting Guide](rate-limiting) for details.

## Error Handling

The API uses standard HTTP status codes and returns detailed error information. See the [Error Handling Guide](error-handling) for details.

EOF
    
    log_success "API documentation generated"
}

# Deploy to GitHub Pages
deploy_to_github_pages() {
    log_info "Deploying to GitHub Pages..."
    
    local repo_url="${GITHUB_REPOSITORY:-kyb-platform/risk-assessment-service}"
    local gh_pages_branch="gh-pages"
    
    # Check if we're in a GitHub Actions environment
    if [ -n "$GITHUB_ACTIONS" ]; then
        # Configure git for GitHub Actions
        git config --global user.name "GitHub Actions"
        git config --global user.email "actions@github.com"
        
        # Set up GitHub Pages deployment
        cd "$BUILD_DIR"
        
        # Install gh-pages package
        npm install --save-dev gh-pages
        
        # Deploy to GitHub Pages
        npx gh-pages -d build -b "$gh_pages_branch" -r "https://github.com/$repo_url.git"
        
        log_success "Deployed to GitHub Pages: https://$(echo "$repo_url" | cut -d'/' -f1).github.io/$(echo "$repo_url" | cut -d'/' -f2)"
    else
        log_warning "Not in GitHub Actions environment, skipping GitHub Pages deployment"
    fi
}

# Deploy to Netlify
deploy_to_netlify() {
    log_info "Deploying to Netlify..."
    
    if [ -n "$NETLIFY_SITE_ID" ] && [ -n "$NETLIFY_AUTH_TOKEN" ]; then
        # Install Netlify CLI
        if ! command -v netlify &> /dev/null; then
            npm install -g netlify-cli
        fi
        
        # Deploy to Netlify
        cd "$BUILD_DIR"
        netlify deploy --prod --dir=build --site="$NETLIFY_SITE_ID"
        
        log_success "Deployed to Netlify"
    else
        log_warning "Netlify credentials not provided, skipping Netlify deployment"
    fi
}

# Deploy to Vercel
deploy_to_vercel() {
    log_info "Deploying to Vercel..."
    
    if [ -n "$VERCEL_TOKEN" ]; then
        # Install Vercel CLI
        if ! command -v vercel &> /dev/null; then
            npm install -g vercel
        fi
        
        # Deploy to Vercel
        cd "$BUILD_DIR"
        vercel --prod --token="$VERCEL_TOKEN"
        
        log_success "Deployed to Vercel"
    else
        log_warning "Vercel token not provided, skipping Vercel deployment"
    fi
}

# Deploy to AWS S3
deploy_to_s3() {
    log_info "Deploying to AWS S3..."
    
    if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ] && [ -n "$S3_BUCKET" ]; then
        # Install AWS CLI
        if ! command -v aws &> /dev/null; then
            log_error "AWS CLI not found. Please install AWS CLI first."
            exit 1
        fi
        
        # Sync to S3
        aws s3 sync "$BUILD_DIR/build" "s3://$S3_BUCKET" --delete
        
        # Set up CloudFront invalidation if distribution ID is provided
        if [ -n "$CLOUDFRONT_DISTRIBUTION_ID" ]; then
            aws cloudfront create-invalidation --distribution-id "$CLOUDFRONT_DISTRIBUTION_ID" --paths "/*"
        fi
        
        log_success "Deployed to AWS S3: https://$S3_BUCKET.s3-website.amazonaws.com"
    else
        log_warning "AWS credentials or S3 bucket not provided, skipping S3 deployment"
    fi
}

# Generate sitemap
generate_sitemap() {
    log_info "Generating sitemap..."
    
    local sitemap_file="$BUILD_DIR/build/sitemap.xml"
    local base_url="${SITE_URL:-https://docs.kyb-platform.com}"
    
    cat > "$sitemap_file" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>$base_url</loc>
        <lastmod>$(date -u +%Y-%m-%d)</lastmod>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>
    <url>
        <loc>$base_url/docs</loc>
        <lastmod>$(date -u +%Y-%m-%d)</lastmod>
        <changefreq>weekly</changefreq>
        <priority>0.8</priority>
    </url>
    <url>
        <loc>$base_url/api</loc>
        <lastmod>$(date -u +%Y-%m-%d)</lastmod>
        <changefreq>weekly</changefreq>
        <priority>0.8</priority>
    </url>
</urlset>
EOF
    
    log_success "Sitemap generated"
}

# Validate documentation
validate_docs() {
    log_info "Validating documentation..."
    
    local validation_errors=0
    
    # Check for broken links
    if command -v linkchecker &> /dev/null; then
        log_info "Checking for broken links..."
        if ! linkchecker "$BUILD_DIR/build" --check-extern; then
            validation_errors=$((validation_errors + 1))
        fi
    else
        log_warning "linkchecker not found, skipping link validation"
    fi
    
    # Check for missing images
    log_info "Checking for missing images..."
    find "$BUILD_DIR/build" -name "*.html" -exec grep -l "img src" {} \; | while read -r file; do
        grep -o 'src="[^"]*"' "$file" | sed 's/src="//;s/"//' | while read -r img_path; do
            if [[ "$img_path" =~ ^http ]]; then
                continue  # Skip external URLs
            fi
            
            local full_path="$BUILD_DIR/build/$img_path"
            if [ ! -f "$full_path" ]; then
                log_error "Missing image: $img_path in $file"
                validation_errors=$((validation_errors + 1))
            fi
        done
    done
    
    # Check for missing documentation files
    log_info "Checking for missing documentation files..."
    local required_docs=(
        "README.md"
        "API_DOCUMENTATION.md"
        "GETTING_STARTED.md"
        "TROUBLESHOOTING.md"
    )
    
    for doc in "${required_docs[@]}"; do
        if [ ! -f "$DOCS_DIR/$doc" ]; then
            log_error "Missing required documentation: $doc"
            validation_errors=$((validation_errors + 1))
        fi
    done
    
    if [ $validation_errors -eq 0 ]; then
        log_success "Documentation validation passed"
    else
        log_error "Documentation validation failed with $validation_errors errors"
        exit 1
    fi
}

# Send deployment notification
send_notification() {
    local deployment_url="$1"
    local status="$2"
    
    if [ -n "$SLACK_WEBHOOK_URL" ]; then
        local color="good"
        if [ "$status" != "success" ]; then
            color="danger"
        fi
        
        local message="Documentation deployment $status"
        if [ -n "$deployment_url" ]; then
            message="$message: $deployment_url"
        fi
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"$message\", \"color\":\"$color\"}" \
            "$SLACK_WEBHOOK_URL"
    fi
    
    if [ -n "$DISCORD_WEBHOOK_URL" ]; then
        local color="3066993"  # Green
        if [ "$status" != "success" ]; then
            color="15158332"  # Red
        fi
        
        local message="Documentation deployment $status"
        if [ -n "$deployment_url" ]; then
            message="$message: $deployment_url"
        fi
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"embeds\":[{\"title\":\"Documentation Deployment\",\"description\":\"$message\",\"color\":$color}]}" \
            "$DISCORD_WEBHOOK_URL"
    fi
}

# Main execution
main() {
    log_info "Starting documentation deployment..."
    
    check_dependencies
    setup_temp_dir
    
    # Build documentation site
    build_docs_site
    
    # Generate sitemap
    generate_sitemap
    
    # Validate documentation
    validate_docs
    
    # Deploy to various platforms
    local deployment_url=""
    local deployment_status="success"
    
    # Deploy to GitHub Pages
    if [ "${DEPLOY_GITHUB_PAGES:-false}" = "true" ]; then
        deploy_to_github_pages
        deployment_url="https://$(echo "${GITHUB_REPOSITORY:-kyb-platform/risk-assessment-service}" | cut -d'/' -f1).github.io/$(echo "${GITHUB_REPOSITORY:-kyb-platform/risk-assessment-service}" | cut -d'/' -f2)"
    fi
    
    # Deploy to Netlify
    if [ "${DEPLOY_NETLIFY:-false}" = "true" ]; then
        deploy_to_netlify
    fi
    
    # Deploy to Vercel
    if [ "${DEPLOY_VERCEL:-false}" = "true" ]; then
        deploy_to_vercel
    fi
    
    # Deploy to AWS S3
    if [ "${DEPLOY_S3:-false}" = "true" ]; then
        deploy_to_s3
        deployment_url="https://${S3_BUCKET}.s3-website.amazonaws.com"
    fi
    
    # Send notification
    send_notification "$deployment_url" "$deployment_status"
    
    log_success "Documentation deployment completed successfully!"
    if [ -n "$deployment_url" ]; then
        log_info "Documentation available at: $deployment_url"
    fi
}

# Run main function
main "$@"
