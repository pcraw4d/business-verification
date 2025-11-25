# Testing Guide for Hybrid Code Generation

## Overview

This document describes the test suite for the hybrid code generation feature, which combines industry-based and keyword-based code matching.

## Test Files Created

### 1. Repository Tests
**File**: `internal/classification/repository/supabase_repository_keyword_codes_test.go`

Tests the new `GetClassificationCodesByKeywords()` repository method:
- Keyword matching with various keyword sets
- Relevance score filtering
- Code type filtering (MCC, SIC, NAICS)
- Empty keyword array handling
- Case-insensitive matching

### 2. Classifier Hybrid Tests
**File**: `internal/classification/classifier_hybrid_test.go`

Tests the hybrid code generation logic:
- `generateCodesFromKeywords()` - keyword-based code generation
- `mergeCodeResults()` - merging industry and keyword results
- Confidence filtering and ranking
- Top-N limiting
- Multi-industry code generation
- Code deduplication across sources

### 3. Integration Tests
**File**: `services/classification-service/test/integration/hybrid_code_generation_test.go`

Template for integration tests (requires database setup):
- End-to-end hybrid code generation flow
- Fallback to industry-only when code_keywords is empty
- Performance testing with large keyword sets
- Multi-industry code generation
- Confidence filtering

## Running Tests

### Unit Tests

```bash
# Run repository keyword codes tests
go test ./internal/classification/repository -run TestGetClassificationCodesByKeywords -v

# Run hybrid classifier tests
go test ./internal/classification -run TestGenerateCodesFromKeywords -v
go test ./internal/classification -run TestMergeCodeResults -v
go test ./internal/classification -run TestGenerateCodesForMultipleIndustries -v

# Run all classifier tests
go test ./internal/classification -v
```

### Integration Tests

```bash
# Set environment variable to enable integration tests
export INTEGRATION_TESTS=true

# Run integration tests
go test ./services/classification-service/test/integration -v
```

## Test Coverage

### Repository Layer
- ✅ Keyword-to-code matching via `code_keywords` table
- ✅ Relevance score filtering
- ✅ Code type filtering
- ✅ Empty result handling

### Classifier Layer
- ✅ Keyword-based code generation
- ✅ Industry-based code generation
- ✅ Hybrid merging with confidence weighting
- ✅ Code deduplication
- ✅ Confidence threshold filtering
- ✅ Top-N limiting
- ✅ Multi-industry support

### Integration Layer
- ⚠️ End-to-end flow (requires database setup)
- ⚠️ Performance benchmarks (requires database setup)
- ⚠️ Fallback behavior (requires database setup)

## Mock Repository Updates

The `MockKeywordRepository` in test files needs to implement the new `GetClassificationCodesByKeywords()` method. This has been added to:
- `internal/classification/service_test.go`
- `internal/classification/method_registry_test.go`

**Note**: Some test files may have duplicate `MockKeywordRepository` definitions. These should be consolidated in a future refactoring.

## Test Data Requirements

For integration tests, the following database tables should be populated:
- `code_keywords` - Keyword-to-code mappings with relevance scores
- `classification_codes` - Industry classification codes
- `industries` - Industry definitions

## Known Issues

1. **Duplicate Mock Definitions**: Multiple test files define `MockKeywordRepository`, causing compilation conflicts. These should be consolidated.

2. **Integration Test Setup**: Integration tests require a test database to be configured. The current tests are templates that need database connection setup.

3. **Existing Test Updates**: Some existing tests in `classifier_test.go` need updates to use the new `generateCodesInParallel()` signature that accepts `[]IndustryResult` instead of individual parameters.

## Future Improvements

1. Consolidate all `MockKeywordRepository` definitions into a single shared test utility
2. Set up test database fixtures for integration tests
3. Add performance benchmarks with realistic data volumes
4. Add test coverage for edge cases (empty tables, malformed data, etc.)
5. Add tests for the API response enhancement (Phase 5)

