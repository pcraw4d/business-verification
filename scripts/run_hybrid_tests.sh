#!/bin/bash

# Script to run hybrid code generation tests with proper configuration
# This script automatically sources credentials from existing .env files

set -e

echo "🧪 Running Hybrid Code Generation Tests"
echo "========================================"
echo ""

# Source environment variables from config files
# Try to load from setup script first, or source directly
if [ -f "scripts/setup_test_env.sh" ]; then
    echo "📋 Loading environment from config files..."
    source scripts/setup_test_env.sh
    echo ""
fi

# Check if Supabase is configured
if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_ANON_KEY" ]; then
    echo "⚠️  Supabase not configured. Running unit tests only..."
    echo ""
    echo "To run integration tests, set:"
    echo "  export SUPABASE_URL=your_supabase_url"
    echo "  export SUPABASE_ANON_KEY=your_anon_key"
    echo ""
    
    # Run unit tests only
    echo "📦 Running unit tests..."
    go test ./internal/classification/testutil -v -run TestHybridCodeGeneration_Integration
    go test ./internal/classification -v -run TestGenerateCodesFromKeywords
    go test ./internal/classification -v -run TestMergeCodeResults
    go test ./internal/classification -v -run TestGenerateCodesForMultipleIndustries
else
    echo "✅ Supabase configured. Running all tests..."
    echo ""
    
    # Run all tests including integration
    echo "📦 Running unit tests..."
    go test ./internal/classification/testutil -v
    
    echo ""
    echo "🔗 Running integration tests..."
    go test ./internal/classification/testutil -v -run TestHybridCodeGeneration_WithRealRepository
    
    echo ""
    echo "📊 Running benchmarks..."
    go test ./internal/classification -bench=BenchmarkHybrid -benchmem -run=^$
fi

echo ""
echo "✅ Tests completed!"

