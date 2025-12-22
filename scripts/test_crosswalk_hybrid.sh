#!/bin/bash
# Test script for hybrid crosswalk approach
# Tests both structured table and JSONB fallback

set -e

echo "=========================================="
echo "HYBRID CROSSWALK APPROACH TEST"
echo "=========================================="
echo ""

# Load environment variables
if [ -f railway.env ]; then
    source railway.env
fi

DATABASE_URL="${DATABASE_URL:-}"

if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL not set"
    exit 1
fi

echo "📊 Testing Structured Table (industry_code_crosswalks)"
echo "---------------------------------------------------"

# Test 1: MCC 5819 -> NAICS (should use structured table)
echo ""
echo "Test 1: MCC 5819 -> NAICS"
psql "$DATABASE_URL" -c "
SELECT 
    source_system,
    source_code,
    target_system,
    target_code,
    confidence_score,
    mapping_type
FROM industry_code_crosswalks
WHERE source_system = 'MCC' 
  AND source_code = '5819' 
  AND target_system = 'NAICS' 
  AND is_active = true
ORDER BY confidence_score DESC
LIMIT 5;
" 2>&1 | grep -v "rows)" | grep -v "^$" || echo "   No results found"

# Test 2: MCC 5819 -> SIC
echo ""
echo "Test 2: MCC 5819 -> SIC"
psql "$DATABASE_URL" -c "
SELECT 
    source_system,
    source_code,
    target_system,
    target_code,
    confidence_score
FROM industry_code_crosswalks
WHERE source_system = 'MCC' 
  AND source_code = '5819' 
  AND target_system = 'SIC' 
  AND is_active = true
ORDER BY confidence_score DESC
LIMIT 5;
" 2>&1 | grep -v "rows)" | grep -v "^$" || echo "   No results found"

# Test 3: NAICS -> MCC (reverse mapping)
echo ""
echo "Test 3: NAICS 445120 -> MCC (reverse mapping)"
psql "$DATABASE_URL" -c "
SELECT 
    source_system,
    source_code,
    target_system,
    target_code,
    confidence_score
FROM industry_code_crosswalks
WHERE source_system = 'NAICS' 
  AND source_code = '445120' 
  AND target_system = 'MCC' 
  AND is_active = true
ORDER BY confidence_score DESC
LIMIT 5;
" 2>&1 | grep -v "rows)" | grep -v "^$" || echo "   No results found"

echo ""
echo "=========================================="
echo "Testing JSONB Fallback (code_metadata)"
echo "=========================================="
echo ""

# Test 4: Check code_metadata for same code
echo "Test 4: code_metadata.crosswalk_data for MCC 5819"
psql "$DATABASE_URL" -c "
SELECT 
    code_type,
    code,
    crosswalk_data->'naics' as naics_codes,
    crosswalk_data->'sic' as sic_codes
FROM code_metadata
WHERE code_type = 'MCC' 
  AND code = '5819' 
  AND is_active = true
LIMIT 1;
" 2>&1 | grep -v "rows)" | grep -v "^$" || echo "   No results found"

echo ""
echo "=========================================="
echo "Performance Test"
echo "=========================================="
echo ""

# Performance test: Compare query times
echo "Testing query performance..."
time psql "$DATABASE_URL" -c "
SELECT COUNT(*) 
FROM industry_code_crosswalks
WHERE source_system = 'MCC' 
  AND source_code = '5819' 
  AND is_active = true;
" > /dev/null 2>&1

echo ""
echo "✅ Crosswalk tests complete"

