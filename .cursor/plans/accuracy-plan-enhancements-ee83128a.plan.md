<!-- ee83128a-be09-4118-8410-b27412d1f4a1 136903b8-47b2-4a9c-8c52-657517c27f50 -->
# Accuracy Plan Enhancements Implementation Plan

## Overview

This plan implements the remaining tasks from the accuracy improvement plan, focusing on expanding code metadata coverage, enhancing crosswalk relationships, improving testing infrastructure, and optimizing performance. All implementations include comprehensive unit and integration tests with measurable success criteria.

## Goals and Success Criteria

### Primary Goals

1. **Code Coverage**: Expand from 150 to 500+ codes in `code_metadata` table
2. **Crosswalk Coverage**: Increase from 24.67% to 50%+ (50+ codes with crosswalks)
3. **Hierarchy Coverage**: Increase from 15.33% to 30%+ (30+ codes with hierarchy)
4. **Industry Mapping**: Increase from 59.33% to 80%+ (120+ codes with mappings)
5. **Keyword Coverage**: 90%+ of codes have 15+ keywords
6. **Accuracy Testing**: Comprehensive test suite with 1000+ test cases validating 95% accuracy target
7. **Performance**: Query optimization reducing metadata retrieval time by 30%+

## Phase 1: Database Population Expansion

### Task 1.1: Expand Code Metadata Population

**Goal**: Increase from 150 to 500+ codes covering all major industries

**Targets**:

- 500+ total codes in `code_metadata` table
- 10+ codes per code type per major industry
- 100% of codes have official descriptions
- Focus on high-frequency codes first

**Implementation**:

- Create `scripts/expand_code_metadata_phase1.sql` with 350+ additional codes
- Prioritize codes by usage frequency (most commonly used first)
- Use official sources:
- Census Bureau for NAICS codes
- IRS for SIC codes
- Payment processor data for MCC codes
- Organize by industry sectors: Technology, Healthcare, Financial Services, Retail, Manufacturing, Construction, Transportation, Education, Hospitality, Professional Services

**Files to Create/Modify**:

- `scripts/expand_code_metadata_phase1.sql` - New comprehensive code population script
- `scripts/verify_code_metadata_coverage.sql` - Verification queries

**Testing**:

- **Unit Tests**: `internal/classification/repository/code_metadata_repository_test.go`
- Test `GetCodeMetadata` with new codes
- Test `GetCodeMetadataBatch` with 100+ codes
- Verify all codes have official descriptions
- **Integration Tests**: `test/integration/code_metadata_population_test.go`
- Verify 500+ codes exist
- Verify coverage across all industries
- Verify official descriptions present
- **Success Criteria**: 
- 500+ codes in database
- 100% have official descriptions
- Coverage across 10+ major industries

### Task 1.2: Expand Keyword Population

**Goal**: 90%+ of codes have 15+ keywords

**Targets**:

- Average 15-20 keywords per code
- Include synonyms, related terms, industry-specific terminology
- Automated keyword extraction from official descriptions

**Implementation**:

- Create `scripts/expand_keywords_phase1.sql` to add keywords for all codes
- Implement keyword extraction utility: `internal/classification/keyword_extractor.go`
- Extract keywords from official descriptions
- Include synonyms and related terms
- Use industry-specific terminology dictionaries
- Update `code_keywords` table with expanded keywords

**Files to Create/Modify**:

- `scripts/expand_keywords_phase1.sql` - Keyword population script
- `internal/classification/keyword_extractor.go` - Automated keyword extraction
- `internal/classification/keyword_extractor_test.go` - Unit tests

**Testing**:

- **Unit Tests**: `internal/classification/keyword_extractor_test.go`
- Test keyword extraction from descriptions
- Test synonym expansion
- Test industry-specific terminology
- **Integration Tests**: `test/integration/keyword_coverage_test.go`
- Verify 90%+ of codes have 15+ keywords
- Verify keyword quality (relevance scoring)
- Test keyword matching accuracy
- **Success Criteria**:
- 90%+ codes have 15+ keywords
- Average 18 keywords per code
- Keyword matching accuracy > 85%

## Phase 2: Crosswalk Coverage Expansion (High Priority)

### Task 2.1: Expand MCC Crosswalks

**Goal**: Increase MCC crosswalk coverage from 10.61% to 45%+ (30+ MCC codes)

**Targets**:

- 30+ MCC codes with crosswalks to NAICS/SIC
- Focus on high-value MCC codes (most commonly used)
- Use payment processor crosswalk data

**Implementation**:

- Create `scripts/expand_mcc_crosswalks.sql` with crosswalk data
- Use official crosswalk sources:
- Payment processor MCC ↔ NAICS mappings
- Industry standard crosswalks
- Update `crosswalk_data` JSONB field in `code_metadata` table

**Files to Create/Modify**:

- `scripts/expand_mcc_crosswalks.sql` - MCC crosswalk population
- `internal/classification/repository/code_metadata_repository.go` - Enhance `GetCrosswalkCodes` method
- `scripts/verify_crosswalk_coverage.sql` - Verification queries

**Testing**:

- **Unit Tests**: `internal/classification/repository/code_metadata_repository_test.go`
- Test `GetCrosswalkCodes` for MCC codes
- Verify crosswalk data structure
- Test bidirectional crosswalk retrieval
- **Integration Tests**: `test/integration/crosswalk_accuracy_test.go`
- Verify 30+ MCC codes have crosswalks
- Test crosswalk accuracy (validate against known mappings)
- Test crosswalk completeness (all related codes present)
- **Success Criteria**:
- 30+ MCC codes with crosswalks (45%+ coverage)
- Crosswalk accuracy > 95%
- All crosswalks validated against official sources

### Task 2.2: Expand SIC Crosswalks

**Goal**: Increase SIC crosswalk coverage from 22.86% to 57%+ (20+ SIC codes)

**Targets**:

- 20+ SIC codes with crosswalks to NAICS/MCC
- Use Census Bureau NAICS ↔ SIC crosswalk tables

**Implementation**:

- Create `scripts/expand_sic_crosswalks.sql`
- Use official Census Bureau crosswalk data
- Update `crosswalk_data` JSONB field

**Files to Create/Modify**:

- `scripts/expand_sic_crosswalks.sql` - SIC crosswalk population
- `scripts/verify_crosswalk_coverage.sql` - Update verification queries

**Testing**:

- **Unit Tests**: Test SIC crosswalk retrieval
- **Integration Tests**: Verify 20+ SIC codes with crosswalks
- **Success Criteria**: 20+ SIC codes with crosswalks (57%+ coverage)

### Task 2.3: Enhance NAICS Crosswalks

**Goal**: Increase NAICS crosswalk coverage from 44.90% to 61%+ (30+ NAICS codes)

**Targets**:

- 30+ NAICS codes with crosswalks to SIC/MCC
- Maintain high accuracy (already good coverage)

**Implementation**:

- Create `scripts/expand_naics_crosswalks.sql`
- Use official Census Bureau and payment processor data

**Files to Create/Modify**:

- `scripts/expand_naics_crosswalks.sql` - NAICS crosswalk expansion
- `scripts/verify_crosswalk_coverage.sql` - Update verification

**Testing**:

- **Unit Tests**: Test NAICS crosswalk retrieval
- **Integration Tests**: Verify 30+ NAICS codes with crosswalks
- **Success Criteria**: 30+ NAICS codes with crosswalks (61%+ coverage)

## Phase 3: Hierarchy Data Expansion

### Task 3.1: Expand NAICS Hierarchy

**Goal**: Increase hierarchy coverage from 15.33% to 30%+ (30+ codes with hierarchy)

**Targets**:

- 30+ NAICS codes with parent/child relationships
- Focus on 4-6 digit NAICS codes (most detailed)
- Add parent relationships for all 5-6 digit codes
- Add child relationships for 2-4 digit codes

**Implementation**:

- Create `scripts/expand_naics_hierarchy.sql`
- Use official NAICS hierarchy structure from Census Bureau
- Update `hierarchy` JSONB field in `code_metadata` table

**Files to Create/Modify**:

- `scripts/expand_naics_hierarchy.sql` - NAICS hierarchy population
- `internal/classification/repository/code_metadata_repository.go` - Enhance `GetHierarchyCodes` method
- `scripts/verify_hierarchy_coverage.sql` - Verification queries

**Testing**:

- **Unit Tests**: `internal/classification/repository/code_metadata_repository_test.go`
- Test `GetHierarchyCodes` for parent retrieval
- Test `GetHierarchyCodes` for child retrieval
- Verify hierarchy data structure
- **Integration Tests**: `test/integration/hierarchy_accuracy_test.go`
- Verify 30+ codes have hierarchy data
- Test hierarchy accuracy (validate parent/child relationships)
- Test hierarchy completeness (all levels present)
- **Success Criteria**:
- 30+ codes with hierarchy (30%+ coverage)
- Hierarchy accuracy > 98%
- All parent/child relationships validated

## Phase 4: Industry Mapping Enhancement

### Task 4.1: Expand Industry Mappings

**Goal**: Increase industry mapping coverage from 59.33% to 80%+ (120+ codes with mappings)

**Targets**:

- 80%+ coverage across all code types
- Add primary and secondary industry mappings
- Include industry category mappings (e.g., "Technology", "Healthcare")
- Use industry classification standards (GICS, ICB)

**Implementation**:

- Create `scripts/expand_industry_mappings.sql`
- Map codes to industries using:
- Primary industry (single)
- Secondary industries (array)
- Industry categories (broader classifications)
- Update `industry_mappings` JSONB field

**Files to Create/Modify**:

- `scripts/expand_industry_mappings.sql` - Industry mapping population
- `internal/classification/repository/code_metadata_repository.go` - Enhance `GetCodesByIndustryMapping` method
- `scripts/verify_industry_mapping_coverage.sql` - Verification queries

**Testing**:

- **Unit Tests**: `internal/classification/repository/code_metadata_repository_test.go`
- Test `GetCodesByIndustryMapping` with various industries
- Test primary vs secondary industry matching
- Test industry category matching
- **Integration Tests**: `test/integration/industry_mapping_test.go`
- Verify 80%+ codes have industry mappings
- Test mapping accuracy (validate against known industries)
- Test code selection based on industry
- **Success Criteria**:
- 80%+ codes have industry mappings
- Mapping accuracy > 90%
- All major industries covered

## Phase 5: Comprehensive Accuracy Testing

### Task 5.1: Create Test Dataset

**Goal**: Create comprehensive test dataset with 1000+ known business classifications

**Targets**:

- 1000+ test cases across all major industries
- Known expected classifications (MCC, NAICS, SIC, Industry)
- Diverse business types and sizes
- Edge cases and boundary conditions

**Implementation**:

- Create `scripts/populate_accuracy_test_dataset.sql`
- Create test data structure:
- Business name, description, website URL
- Expected MCC, NAICS, SIC codes
- Expected primary industry
- Test category (e.g., "Technology", "Healthcare", "Edge Cases")
- Create `internal/testing/accuracy_test_dataset.go` for test data management

**Files to Create/Modify**:

- `scripts/populate_accuracy_test_dataset.sql` - Test data population
- `internal/testing/accuracy_test_dataset.go` - Test dataset management
- `internal/testing/accuracy_test_dataset_test.go` - Unit tests

**Testing**:

- **Unit Tests**: Test dataset loading and validation
- **Integration Tests**: Verify 1000+ test cases loaded
- **Success Criteria**: 1000+ test cases across 10+ industries

### Task 5.2: Implement Comprehensive Accuracy Test Suite

**Goal**: Validate 95% accuracy target with comprehensive test suite

**Targets**:

- 95%+ accuracy for industry classification
- 90%+ accuracy for code generation (MCC, NAICS, SIC)
- Test across all major industries
- Validate crosswalk accuracy
- Test hierarchy relationships
- Measure end-to-end accuracy

**Implementation**:

- Create `internal/testing/comprehensive_accuracy_tester.go`
- Load test dataset
- Run classification for each test case
- Compare results with expected values
- Calculate accuracy metrics:
- Industry accuracy
- Code accuracy (per code type)
- Crosswalk accuracy
- Hierarchy accuracy
- End-to-end accuracy
- Create accuracy reporting: `internal/testing/accuracy_report.go`
- Generate detailed accuracy reports
- Identify failure patterns
- Track accuracy over time

**Files to Create/Modify**:

- `internal/testing/comprehensive_accuracy_tester.go` - Main accuracy tester
- `internal/testing/accuracy_report.go` - Accuracy reporting
- `internal/testing/comprehensive_accuracy_tester_test.go` - Unit tests
- `test/integration/comprehensive_accuracy_test.go` - Integration tests
- `cmd/accuracy_test_runner/main.go` - Update to use new tester

**Testing**:

- **Unit Tests**: `internal/testing/comprehensive_accuracy_tester_test.go`
- Test accuracy calculation
- Test metric aggregation
- Test report generation
- **Integration Tests**: `test/integration/comprehensive_accuracy_test.go`
- Run full accuracy test suite
- Verify 95%+ industry accuracy
- Verify 90%+ code accuracy
- Test crosswalk accuracy
- Test hierarchy accuracy
- **Success Criteria**:
- 95%+ industry classification accuracy
- 90%+ code generation accuracy
- All tests passing
- Comprehensive accuracy report generated

## Phase 6: Code Metadata Quality Improvements

### Task 6.1: Enhance Metadata Completeness

**Goal**: Ensure all codes have complete metadata

**Targets**:

- 100% of codes have official descriptions
- 80%+ of codes have additional metadata (aliases, usage frequency, regulatory info)
- All codes validated as active and official

**Implementation**:

- Create `scripts/enhance_code_metadata.sql` to populate `metadata` JSONB field
- Add metadata fields:
- Code aliases and alternative names
- Usage frequency data
- Regulatory information
- Last updated timestamp
- Source information
- Create validation utility: `internal/classification/metadata_validator.go`

**Files to Create/Modify**:

- `scripts/enhance_code_metadata.sql` - Metadata enhancement script
- `internal/classification/metadata_validator.go` - Metadata validation
- `internal/classification/metadata_validator_test.go` - Unit tests

**Testing**:

- **Unit Tests**: Test metadata validation and completeness checks
- **Integration Tests**: Verify 80%+ codes have enhanced metadata
- **Success Criteria**: 80%+ codes have complete metadata

### Task 6.2: Implement Code Validation System

**Goal**: Ensure all codes are active and valid

**Targets**:

- Periodic validation against official code lists
- Flag deprecated codes
- Update codes when new versions released (e.g., NAICS 2027)

**Implementation**:

- Create `internal/classification/code_validator.go`
- Validate codes against official sources
- Flag deprecated codes
- Update code versions
- Create scheduled job: `internal/jobs/code_validation_job.go`
- Run periodic validation
- Update deprecated codes
- Generate validation reports

**Files to Create/Modify**:

- `internal/classification/code_validator.go` - Code validation logic
- `internal/classification/code_validator_test.go` - Unit tests
- `internal/jobs/code_validation_job.go` - Scheduled validation job
- `internal/jobs/code_validation_job_test.go` - Unit tests

**Testing**:

- **Unit Tests**: Test code validation logic
- **Integration Tests**: Test scheduled validation job
- **Success Criteria**: All active codes validated, deprecated codes flagged

## Phase 7: Performance Optimization

### Task 7.1: Query Optimization

**Goal**: Reduce metadata retrieval time by 30%+

**Targets**:

- 30%+ reduction in query execution time
- Optimize crosswalk and hierarchy queries
- Improve JSONB query performance

**Implementation**:

- Create materialized views: `scripts/create_code_metadata_views.sql`
- `code_crosswalk_materialized_view` for crosswalk queries
- `code_hierarchy_materialized_view` for hierarchy queries
- Refresh strategy for materialized views
- Optimize JSONB queries with GIN indexes (already exist, verify)
- Add query caching: `internal/classification/cache/metadata_cache.go`
- Cache frequently accessed metadata
- Cache crosswalk data
- Cache hierarchy data
- Consider denormalization for high-frequency queries

**Files to Create/Modify**:

- `scripts/create_code_metadata_views.sql` - Materialized views
- `internal/classification/cache/metadata_cache.go` - Metadata caching
- `internal/classification/cache/metadata_cache_test.go` - Unit tests
- `scripts/verify_query_performance.sql` - Performance verification

**Testing**:

- **Unit Tests**: Test cache functionality
- **Integration Tests**: `test/integration/metadata_performance_test.go`
- Benchmark query performance
- Verify 30%+ improvement
- Test cache hit rates
- **Success Criteria**:
- 30%+ query performance improvement
- Cache hit rate > 70%
- All queries optimized

## Phase 8: Advanced NLP Features (Optional)

### Task 8.1: Enhanced Word Segmentation

**Goal**: Improve keyword extraction from domains

**Targets**:

- Better compound domain segmentation
- Business-specific dictionary
- Domain name pattern recognition

**Implementation**:

- Enhance `internal/classification/word_segmentation/segmenter.go`
- Add business-specific dictionary
- Implement domain pattern recognition

**Files to Create/Modify**:

- `internal/classification/word_segmentation/segmenter.go` - Enhanced segmentation
- `internal/classification/word_segmentation/segmenter_test.go` - Unit tests

**Testing**:

- **Unit Tests**: Test enhanced segmentation
- **Integration Tests**: Test domain segmentation accuracy
- **Success Criteria**: Improved keyword extraction from domains

### Task 8.2: Advanced Named Entity Recognition

**Goal**: Better entity extraction

**Targets**:

- Industry-specific entity patterns
- Business entity recognition (company types, legal structures)
- Improved location and date extraction

**Implementation**:

- Enhance `internal/classification/nlp/entity_recognizer.go`
- Add industry-specific patterns
- Improve business entity recognition

**Files to Create/Modify**:

- `internal/classification/nlp/entity_recognizer.go` - Enhanced NER
- `internal/classification/nlp/entity_recognizer_test.go` - Unit tests

**Testing**:

- **Unit Tests**: Test entity recognition
- **Integration Tests**: Test entity extraction accuracy
- **Success Criteria**: Improved entity extraction accuracy

### Task 8.3: Advanced Topic Modeling

**Goal**: Better industry classification

**Targets**:

- Industry-specific topic models
- Multi-topic classification support
- Improved topic-to-industry mapping

**Implementation**:

- Enhance `internal/classification/nlp/topic_modeler.go`
- Add industry-specific models
- Implement multi-topic classification

**Files to Create/Modify**:

- `internal/classification/nlp/topic_modeler.go` - Enhanced topic modeling
- `internal/classification/nlp/topic_modeler_test.go` - Unit tests

**Testing**:

- **Unit Tests**: Test topic modeling
- **Integration Tests**: Test topic classification accuracy
- **Success Criteria**: Improved topic classification accuracy

## Testing Strategy

### Unit Testing Requirements

All new code must include comprehensive unit tests:

- Test coverage > 80% for all new code
- Table-driven tests for multiple scenarios
- Mock external dependencies
- Test error handling and edge cases

### Integration Testing Requirements

All features must include integration tests:

- Test with real database
- Test end-to-end workflows
- Validate against success criteria
- Performance benchmarking

### Test Organization

- Unit tests: `internal/classification/**/*_test.go`
- Integration tests: `test/integration/**/*_test.go`
- Test utilities: `internal/classification/testutil/`
- Test data: `test/fixtures/`

## Implementation Order

1. **Phase 1**: Database Population (Tasks 1.1, 1.2) - Foundation
2. **Phase 2**: Crosswalk Coverage (Tasks 2.1, 2.2, 2.3) - High Priority
3. **Phase 3**: Hierarchy Data (Task 3.1) - Medium Priority
4. **Phase 4**: Industry Mappings (Task 4.1) - Medium Priority
5. **Phase 5**: Accuracy Testing (Tasks 5.1, 5.2) - Validation
6. **Phase 6**: Metadata Quality (Tasks 6.1, 6.2) - Quality Assurance
7. **Phase 7**: Performance Optimization (Task 7.1) - Optimization
8. **Phase 8**: Advanced NLP (Tasks 8.1, 8.2, 8.3) - Optional Enhancements

## Success Metrics

### Code Coverage Metrics

- Total codes: 500+ (from 150)
- Codes with keywords (15+): 90%+ (from variable)
- Codes with crosswalks: 50%+ (from 24.67%)
- Codes with hierarchy: 30%+ (from 15.33%)
- Codes with industry mappings: 80%+ (from 59.33%)

### Accuracy Metrics

- Industry classification accuracy: 95%+
- Code generation accuracy: 90%+
- Crosswalk accuracy: 95%+
- Hierarchy accuracy: 98%+

### Performance Metrics

- Query performance improvement: 30%+
- Cache hit rate: 70%+
- Metadata retrieval time: < 50ms (p95)

## Documentation

- Update `docs/plan-completion-status-review.md` after each phase
- Create migration documentation for each SQL script
- Document test data sources and validation methods
- Create accuracy testing guide

## Risk Mitigation

- Incremental implementation (one phase at a time)
- Comprehensive testing at each phase
- Rollback procedures for database changes
- Data validation before and after migrations
- Performance monitoring during implementation

### To-dos

- [ ] Expand code metadata from 150 to 500+ codes covering all major industries. Create expand_code_metadata_phase1.sql with 350+ additional codes from official sources (Census Bureau, IRS, payment processors). Include unit and integration tests verifying 500+ codes exist with official descriptions.
- [ ] Expand keyword population to 90%+ of codes having 15+ keywords. Create expand_keywords_phase1.sql and implement keyword_extractor.go for automated extraction. Include unit and integration tests verifying keyword coverage and matching accuracy > 85%.
- [ ] Expand MCC crosswalk coverage from 10.61% to 45%+ (30+ MCC codes). Create expand_mcc_crosswalks.sql using payment processor data. Include unit and integration tests verifying crosswalk coverage and accuracy > 95%.
- [ ] Expand SIC crosswalk coverage from 22.86% to 57%+ (20+ SIC codes). Create expand_sic_crosswalks.sql using Census Bureau data. Include unit and integration tests verifying crosswalk coverage.
- [ ] Enhance NAICS crosswalk coverage from 44.90% to 61%+ (30+ NAICS codes). Create expand_naics_crosswalks.sql. Include unit and integration tests verifying crosswalk coverage.
- [ ] Expand NAICS hierarchy coverage from 15.33% to 30%+ (30+ codes). Create expand_naics_hierarchy.sql with parent/child relationships. Include unit and integration tests verifying hierarchy accuracy > 98%.
- [ ] Expand industry mapping coverage from 59.33% to 80%+ (120+ codes). Create expand_industry_mappings.sql with primary/secondary industries and categories. Include unit and integration tests verifying mapping accuracy > 90%.
- [ ] Create comprehensive test dataset with 1000+ known business classifications. Create populate_accuracy_test_dataset.sql and accuracy_test_dataset.go for test data management. Include unit and integration tests verifying 1000+ test cases across 10+ industries.
- [ ] Implement comprehensive accuracy test suite validating 95% accuracy target. Create comprehensive_accuracy_tester.go and accuracy_report.go. Include unit and integration tests verifying 95%+ industry accuracy and 90%+ code accuracy.
- [ ] Enhance metadata completeness to 80%+ of codes having additional metadata (aliases, usage frequency, regulatory info). Create enhance_code_metadata.sql and metadata_validator.go. Include unit and integration tests verifying metadata completeness.
- [ ] Implement code validation system for periodic validation against official sources. Create code_validator.go and code_validation_job.go. Include unit and integration tests verifying validation functionality.
- [ ] Optimize queries to reduce metadata retrieval time by 30%+. Create materialized views, implement metadata_cache.go, and optimize JSONB queries. Include integration tests with performance benchmarks verifying 30%+ improvement.
- [ ] Implement advanced NLP features: enhanced word segmentation, advanced NER, and advanced topic modeling. Enhance existing segmenter.go, entity_recognizer.go, and topic_modeler.go. Include unit and integration tests for each enhancement.