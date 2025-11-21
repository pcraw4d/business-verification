# Phase 2 Testing Plan

## Overview
This document outlines the testing strategy for Phase 2: Fix Missing Data Issues and Improve Error Handling.

## Test Categories

### 1. Error Code Display Tests
- [ ] Verify error codes are displayed in all error messages
- [ ] Verify error code format: "Error CODE-XXX: [message]"
- [ ] Test all error code categories:
  - PortfolioComparison (PC-001 to PC-005)
  - RiskScore (RS-001 to RS-003)
  - AnalyticsComparison (AC-001 to AC-005)
  - RiskBenchmark (RB-001 to RB-005)

### 2. Error Handling Tests
- [ ] Test missing risk score scenario
- [ ] Test missing portfolio stats scenario
- [ ] Test missing both data scenarios
- [ ] Test invalid data scenarios
- [ ] Test API fetch failures

### 3. CTA (Call-to-Action) Button Tests
- [ ] "Run Risk Assessment" button appears when risk score missing
- [ ] "Refresh Data" button appears when portfolio stats missing
- [ ] "Enrich Data" button appears when industry code missing
- [ ] "Retry" button appears on errors
- [ ] All buttons are clickable and functional

### 4. Loading State Tests
- [ ] Loading messages are descriptive ("Loading portfolio comparison..." not just "Loading...")
- [ ] Loading states show appropriate skeletons
- [ ] Loading states transition correctly to content or error

### 5. Partial Data Display Tests
- [ ] Component shows merchant score only when portfolio stats missing
- [ ] Component shows portfolio average when risk score missing
- [ ] Component shows full comparison when both available
- [ ] No errors thrown with partial data

### 6. Type Guard and Validation Tests
- [ ] Invalid risk scores are caught and handled
- [ ] Invalid portfolio stats are caught and handled
- [ ] Type guards prevent runtime errors
- [ ] Validation provides helpful error messages

### 7. Development Logging Tests
- [ ] Console logs appear in development mode
- [ ] Logs include API response structures
- [ ] Logs include field availability
- [ ] No logs in production mode

## Test Execution Plan

### Manual Testing Checklist
1. **PortfolioComparisonCard**
   - [ ] Test with missing risk score
   - [ ] Test with missing portfolio stats
   - [ ] Test with both missing
   - [ ] Test with invalid data
   - [ ] Verify error codes displayed
   - [ ] Verify CTAs work

2. **RiskScoreCard**
   - [ ] Test with no risk score
   - [ ] Test with invalid risk score
   - [ ] Test API failure
   - [ ] Verify error codes displayed
   - [ ] Verify "Start Risk Assessment" button

3. **AnalyticsComparison**
   - [ ] Test with missing merchant analytics
   - [ ] Test with missing portfolio analytics
   - [ ] Test with both missing
   - [ ] Verify error codes displayed
   - [ ] Verify retry functionality

4. **RiskBenchmarkComparison**
   - [ ] Test with missing industry code
   - [ ] Test with unavailable benchmarks
   - [ ] Test with missing risk score
   - [ ] Verify error codes displayed
   - [ ] Verify "Enrich Data" button

### Automated Testing
- Run existing unit tests
- Verify all tests pass
- Check test coverage for new code

## Success Criteria
- All error messages include error codes
- All error states have actionable CTAs
- Loading states are descriptive
- Partial data scenarios handled gracefully
- Type guards prevent runtime errors
- No console errors in browser
- All automated tests pass

