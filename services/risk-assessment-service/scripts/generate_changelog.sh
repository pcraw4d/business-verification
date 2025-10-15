#!/bin/bash

# Changelog Generation Script
# This script generates changelog from git commits using conventional commits

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHANGELOG_FILE="$PROJECT_ROOT/CHANGELOG.md"
TEMP_DIR="/tmp/changelog_generation"

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
    
    if ! command -v git &> /dev/null; then
        missing_deps+=("git")
    fi
    
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
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

# Get the latest version tag
get_latest_version() {
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -z "$latest_tag" ]; then
        echo "v1.0.0"
    else
        echo "$latest_tag"
    fi
}

# Get the next version based on conventional commits
get_next_version() {
    local current_version="$1"
    local commits_since_tag=$(git log --oneline "$current_version..HEAD" 2>/dev/null || git log --oneline)
    
    local major_bump=false
    local minor_bump=false
    local patch_bump=false
    
    while IFS= read -r commit; do
        local commit_type=$(echo "$commit" | cut -d' ' -f2 | cut -d':' -f1)
        local is_breaking=$(echo "$commit" | grep -q "BREAKING CHANGE\|!" && echo "true" || echo "false")
        
        case "$commit_type" in
            "feat")
                if [ "$is_breaking" = "true" ]; then
                    major_bump=true
                else
                    minor_bump=true
                fi
                ;;
            "fix")
                patch_bump=true
                ;;
            "perf"|"refactor")
                patch_bump=true
                ;;
            "docs"|"style"|"test"|"chore")
                # No version bump for these
                ;;
            *)
                # Unknown commit type, assume patch
                patch_bump=true
                ;;
        esac
    done <<< "$commits_since_tag"
    
    # Determine version bump
    if [ "$major_bump" = "true" ]; then
        bump_major_version "$current_version"
    elif [ "$minor_bump" = "true" ]; then
        bump_minor_version "$current_version"
    elif [ "$patch_bump" = "true" ]; then
        bump_patch_version "$current_version"
    else
        echo "$current_version"
    fi
}

# Bump major version
bump_major_version() {
    local version="$1"
    local major=$(echo "$version" | sed 's/v//' | cut -d'.' -f1)
    local new_major=$((major + 1))
    echo "v${new_major}.0.0"
}

# Bump minor version
bump_minor_version() {
    local version="$1"
    local major=$(echo "$version" | sed 's/v//' | cut -d'.' -f1)
    local minor=$(echo "$version" | sed 's/v//' | cut -d'.' -f2)
    local new_minor=$((minor + 1))
    echo "v${major}.${new_minor}.0"
}

# Bump patch version
bump_patch_version() {
    local version="$1"
    local major=$(echo "$version" | sed 's/v//' | cut -d'.' -f1)
    local minor=$(echo "$version" | sed 's/v//' | cut -d'.' -f2)
    local patch=$(echo "$version" | sed 's/v//' | cut -d'.' -f3)
    local new_patch=$((patch + 1))
    echo "v${major}.${minor}.${new_patch}"
}

# Get commits since last tag
get_commits_since_tag() {
    local tag="$1"
    local commits_file="$TEMP_DIR/commits.json"
    
    # Get commits in JSON format
    git log --pretty=format:'{"hash":"%H","short_hash":"%h","author":"%an","date":"%ad","subject":"%s","body":"%b"}' \
        --date=iso \
        "$tag..HEAD" 2>/dev/null > "$commits_file" || \
    git log --pretty=format:'{"hash":"%H","short_hash":"%h","author":"%an","date":"%ad","subject":"%s","body":"%b"}' \
        --date=iso > "$commits_file"
    
    echo "$commits_file"
}

# Parse conventional commits
parse_commits() {
    local commits_file="$1"
    local parsed_file="$TEMP_DIR/parsed_commits.json"
    
    # Initialize arrays for different types
    local features=()
    local fixes=()
    local breaking_changes=()
    local performance=()
    local refactors=()
    local docs=()
    local other=()
    
    # Process each commit
    while IFS= read -r commit_json; do
        local hash=$(echo "$commit_json" | jq -r '.hash')
        local short_hash=$(echo "$commit_json" | jq -r '.short_hash')
        local author=$(echo "$commit_json" | jq -r '.author')
        local date=$(echo "$commit_json" | jq -r '.date')
        local subject=$(echo "$commit_json" | jq -r '.subject')
        local body=$(echo "$commit_json" | jq -r '.body')
        
        # Parse conventional commit
        local commit_type=$(echo "$subject" | cut -d' ' -f1 | cut -d':' -f1)
        local scope=$(echo "$subject" | cut -d' ' -f1 | cut -d':' -f2)
        local description=$(echo "$subject" | cut -d' ' -f2-)
        local is_breaking=$(echo "$subject" | grep -q "!" && echo "true" || echo "false")
        
        # Create commit object
        local commit_obj=$(jq -n \
            --arg hash "$hash" \
            --arg short_hash "$short_hash" \
            --arg author "$author" \
            --arg date "$date" \
            --arg type "$commit_type" \
            --arg scope "$scope" \
            --arg description "$description" \
            --argjson breaking "$is_breaking" \
            --arg body "$body" \
            '{
                hash: $hash,
                short_hash: $short_hash,
                author: $author,
                date: $date,
                type: $type,
                scope: $scope,
                description: $description,
                breaking: $breaking,
                body: $body
            }')
        
        # Categorize commit
        case "$commit_type" in
            "feat")
                if [ "$is_breaking" = "true" ]; then
                    breaking_changes+=("$commit_obj")
                else
                    features+=("$commit_obj")
                fi
                ;;
            "fix")
                fixes+=("$commit_obj")
                ;;
            "perf")
                performance+=("$commit_obj")
                ;;
            "refactor")
                refactors+=("$commit_obj")
                ;;
            "docs")
                docs+=("$commit_obj")
                ;;
            *)
                other+=("$commit_obj")
                ;;
        esac
    done < "$commits_file"
    
    # Create categorized commits object
    local categorized=$(jq -n \
        --argjson features "$(printf '%s\n' "${features[@]}" | jq -s .)" \
        --argjson fixes "$(printf '%s\n' "${fixes[@]}" | jq -s .)" \
        --argjson breaking_changes "$(printf '%s\n' "${breaking_changes[@]}" | jq -s .)" \
        --argjson performance "$(printf '%s\n' "${performance[@]}" | jq -s .)" \
        --argjson refactors "$(printf '%s\n' "${refactors[@]}" | jq -s .)" \
        --argjson docs "$(printf '%s\n' "${docs[@]}" | jq -s .)" \
        --argjson other "$(printf '%s\n' "${other[@]}" | jq -s .)" \
        '{
            features: $features,
            fixes: $fixes,
            breaking_changes: $breaking_changes,
            performance: $performance,
            refactors: $refactors,
            docs: $docs,
            other: $other
        }')
    
    echo "$categorized" > "$parsed_file"
    echo "$parsed_file"
}

# Generate changelog entry
generate_changelog_entry() {
    local version="$1"
    local parsed_file="$2"
    local entry_file="$TEMP_DIR/changelog_entry.md"
    
    local date=$(date -u +%Y-%m-%d)
    
    cat > "$entry_file" << EOF
## [$version] - $date

EOF
    
    # Add breaking changes
    local breaking_count=$(jq '.breaking_changes | length' "$parsed_file")
    if [ "$breaking_count" -gt 0 ]; then
        echo "### Breaking Changes" >> "$entry_file"
        echo "" >> "$entry_file"
        
        jq -r '.breaking_changes[] | "- **\(.scope)**: \(.description) (\(.short_hash))"' "$parsed_file" >> "$entry_file"
        echo "" >> "$entry_file"
    fi
    
    # Add new features
    local features_count=$(jq '.features | length' "$parsed_file")
    if [ "$features_count" -gt 0 ]; then
        echo "### Added" >> "$entry_file"
        echo "" >> "$entry_file"
        
        jq -r '.features[] | "- **\(.scope)**: \(.description) (\(.short_hash))"' "$parsed_file" >> "$entry_file"
        echo "" >> "$entry_file"
    fi
    
    # Add bug fixes
    local fixes_count=$(jq '.fixes | length' "$parsed_file")
    if [ "$fixes_count" -gt 0 ]; then
        echo "### Fixed" >> "$entry_file"
        echo "" >> "$entry_file"
        
        jq -r '.fixes[] | "- **\(.scope)**: \(.description) (\(.short_hash))"' "$parsed_file" >> "$entry_file"
        echo "" >> "$entry_file"
    fi
    
    # Add performance improvements
    local perf_count=$(jq '.performance | length' "$parsed_file")
    if [ "$perf_count" -gt 0 ]; then
        echo "### Performance" >> "$entry_file"
        echo "" >> "$entry_file"
        
        jq -r '.performance[] | "- **\(.scope)**: \(.description) (\(.short_hash))"' "$parsed_file" >> "$entry_file"
        echo "" >> "$entry_file"
    fi
    
    # Add refactoring
    local refactor_count=$(jq '.refactors | length' "$parsed_file")
    if [ "$refactor_count" -gt 0 ]; then
        echo "### Changed" >> "$entry_file"
        echo "" >> "$entry_file"
        
        jq -r '.refactors[] | "- **\(.scope)**: \(.description) (\(.short_hash))"' "$parsed_file" >> "$entry_file"
        echo "" >> "$entry_file"
    fi
    
    # Add documentation
    local docs_count=$(jq '.docs | length' "$parsed_file")
    if [ "$docs_count" -gt 0 ]; then
        echo "### Documentation" >> "$entry_file"
        echo "" >> "$entry_file"
        
        jq -r '.docs[] | "- **\(.scope)**: \(.description) (\(.short_hash))"' "$parsed_file" >> "$entry_file"
        echo "" >> "$entry_file"
    fi
    
    # Add other changes
    local other_count=$(jq '.other | length' "$parsed_file")
    if [ "$other_count" -gt 0 ]; then
        echo "### Other" >> "$entry_file"
        echo "" >> "$entry_file"
        
        jq -r '.other[] | "- **\(.scope)**: \(.description) (\(.short_hash))"' "$parsed_file" >> "$entry_file"
        echo "" >> "$entry_file"
    fi
    
    echo "$entry_file"
}

# Update main changelog file
update_changelog() {
    local version="$1"
    local entry_file="$2"
    
    log_info "Updating changelog file..."
    
    # Create backup
    cp "$CHANGELOG_FILE" "$CHANGELOG_FILE.backup.$(date +%Y%m%d_%H%M%S)"
    
    # Read current changelog
    local changelog_content=$(cat "$CHANGELOG_FILE")
    local new_entry=$(cat "$entry_file")
    
    # Find the position to insert new entry (after [Unreleased] section)
    local unreleased_section=$(echo "$changelog_content" | grep -n "## \[Unreleased\]" | cut -d: -f1)
    
    if [ -n "$unreleased_section" ]; then
        # Insert after [Unreleased] section
        local before_unreleased=$(echo "$changelog_content" | head -n "$unreleased_section")
        local after_unreleased=$(echo "$changelog_content" | tail -n +$((unreleased_section + 1)))
        
        # Create new changelog
        cat > "$CHANGELOG_FILE" << EOF
$before_unreleased

$new_entry
$after_unreleased
EOF
    else
        # No [Unreleased] section, prepend to file
        cat > "$CHANGELOG_FILE" << EOF
$new_entry

$changelog_content
EOF
    fi
    
    log_success "Changelog updated successfully"
}

# Generate release notes
generate_release_notes() {
    local version="$1"
    local parsed_file="$2"
    local release_notes_file="$PROJECT_ROOT/RELEASE_NOTES.md"
    
    log_info "Generating release notes..."
    
    local date=$(date -u +%Y-%m-%d)
    
    cat > "$release_notes_file" << EOF
# Release Notes - $version

**Release Date**: $date

## Summary

This release includes $(jq '.features | length' "$parsed_file") new features, $(jq '.fixes | length' "$parsed_file") bug fixes, and $(jq '.breaking_changes | length' "$parsed_file") breaking changes.

## What's New

EOF
    
    # Add features
    local features_count=$(jq '.features | length' "$parsed_file")
    if [ "$features_count" -gt 0 ]; then
        echo "### 🚀 New Features" >> "$release_notes_file"
        echo "" >> "$release_notes_file"
        
        jq -r '.features[] | "- \(.description)"' "$parsed_file" >> "$release_notes_file"
        echo "" >> "$release_notes_file"
    fi
    
    # Add breaking changes
    local breaking_count=$(jq '.breaking_changes | length' "$parsed_file")
    if [ "$breaking_count" -gt 0 ]; then
        echo "### ⚠️ Breaking Changes" >> "$release_notes_file"
        echo "" >> "$release_notes_file"
        
        jq -r '.breaking_changes[] | "- \(.description)"' "$parsed_file" >> "$release_notes_file"
        echo "" >> "$release_notes_file"
    fi
    
    # Add bug fixes
    local fixes_count=$(jq '.fixes | length' "$parsed_file")
    if [ "$fixes_count" -gt 0 ]; then
        echo "### 🐛 Bug Fixes" >> "$release_notes_file"
        echo "" >> "$release_notes_file"
        
        jq -r '.fixes[] | "- \(.description)"' "$parsed_file" >> "$release_notes_file"
        echo "" >> "$release_notes_file"
    fi
    
    # Add performance improvements
    local perf_count=$(jq '.performance | length' "$parsed_file")
    if [ "$perf_count" -gt 0 ]; then
        echo "### ⚡ Performance Improvements" >> "$release_notes_file"
        echo "" >> "$release_notes_file"
        
        jq -r '.performance[] | "- \(.description)"' "$parsed_file" >> "$release_notes_file"
        echo "" >> "$release_notes_file"
    fi
    
    echo "## Migration Guide" >> "$release_notes_file"
    echo "" >> "$release_notes_file"
    echo "Please refer to the [Migration Guide](docs/MIGRATION_GUIDES.md) for detailed instructions on upgrading to this version." >> "$release_notes_file"
    echo "" >> "$release_notes_file"
    
    echo "## Full Changelog" >> "$release_notes_file"
    echo "" >> "$release_notes_file"
    echo "See the [full changelog](CHANGELOG.md) for a complete list of changes." >> "$release_notes_file"
    
    log_success "Release notes generated"
}

# Create git tag
create_git_tag() {
    local version="$1"
    
    log_info "Creating git tag: $version"
    
    if git tag -l | grep -q "^$version$"; then
        log_warning "Tag $version already exists"
        return
    fi
    
    git tag -a "$version" -m "Release $version"
    log_success "Git tag created: $version"
}

# Main execution
main() {
    log_info "Starting changelog generation..."
    
    check_dependencies
    setup_temp_dir
    
    # Get current and next version
    local current_version=$(get_latest_version)
    local next_version=$(get_next_version "$current_version")
    
    log_info "Current version: $current_version"
    log_info "Next version: $next_version"
    
    # Get and parse commits
    local commits_file=$(get_commits_since_tag "$current_version")
    local parsed_file=$(parse_commits "$commits_file")
    
    # Generate changelog entry
    local entry_file=$(generate_changelog_entry "$next_version" "$parsed_file")
    
    # Update changelog
    update_changelog "$next_version" "$entry_file"
    
    # Generate release notes
    generate_release_notes "$next_version" "$parsed_file"
    
    # Create git tag (optional)
    if [ "${CREATE_TAG:-false}" = "true" ]; then
        create_git_tag "$next_version"
    fi
    
    log_success "Changelog generation completed successfully!"
    log_info "Generated files:"
    log_info "  - $CHANGELOG_FILE"
    log_info "  - $PROJECT_ROOT/RELEASE_NOTES.md"
    
    if [ "${CREATE_TAG:-false}" = "true" ]; then
        log_info "  - Git tag: $next_version"
    fi
}

# Run main function
main "$@"
