# Error Handling Review - Track 2.1

## Overview

This document reviews error handling in critical paths of the classification service to identify issues contributing to the 67.1% error rate.

## Critical Paths Reviewed

### 1. Website Scraping (`internal/external/website_scraper.go:277-341`)

**Current Implementation:**
- Retry logic with exponential backoff
- Max retries: 3 (configurable)
- Retry delay: 1 second (reduced from 2s)
- Error categorization for retryable vs non-retryable errors

**Issues Identified:**
1. DNS errors may not be retried with fallback DNS servers
2. Network timeouts may not have sufficient retry attempts
3. HTTP 403/429 errors may not be handled appropriately

**Recommendations:**
- Add fallback DNS servers (8.8.8.8, 1.1.1.1, 8.8.4.4) for DNS failures
- Increase retry attempts for network timeouts
- Add specific handling for HTTP 429 (rate limiting) with longer backoff

### 2. Request Processing (`services/classification-service/internal/handlers/classification.go:2201`)

**Current Implementation:**
- Panic recovery in place
- Context expiration handling
- Error responses with frontend-compatible structure

**Issues Identified:**
1. Context cancellations may cause errors instead of graceful handling
2. Errors may not be properly categorized before logging
3. Error metadata may not include sufficient context for debugging

**Recommendations:**
- Improve context cancellation handling to return partial results when possible
- Add error categorization before logging
- Include request context in error metadata

### 3. Code Generation (`internal/classification/classifier.go:256`)

**Current Implementation:**
- Error handling in code generation
- Fallback to empty codes on failure

**Issues Identified:**
1. Code generation errors may be silently ignored
2. Database query errors may not be retried
3. Missing code metadata may not be logged appropriately

**Recommendations:**
- Add retry logic for database queries
- Log code generation failures with more detail
- Add fallback mechanisms for missing code metadata

## Retry Logic Review

### Adaptive Retry Strategy (`internal/classification/retry/adaptive_retry.go:47-107`)

**Current Implementation:**
- DNS errors: Retry with `defaultMaxRetries + 1` attempts
- Network timeouts: Retry with `defaultMaxRetries` attempts
- HTTP 5xx: Retry with `defaultMaxRetries` attempts
- HTTP 4xx (400, 403, 404): No retry (correct)
- HTTP 429: Retry with 5 attempts (good)
- Exponential backoff with jitter

**Issues Identified:**
1. DNS errors don't use fallback DNS servers
2. Network errors may need more retry attempts
3. Context cancellations may not be handled gracefully

**Recommendations:**
1. **DNS Fallback Servers**: Implement fallback DNS resolution using multiple DNS servers:
   - Primary: System DNS
   - Fallback 1: 8.8.8.8 (Google)
   - Fallback 2: 1.1.1.1 (Cloudflare)
   - Fallback 3: 8.8.4.4 (Google secondary)

2. **Network Timeout Retries**: Increase retry attempts for network timeouts from default to default + 2

3. **Context Cancellation**: Check context before each retry attempt and return partial results if context expires

## Error Categories Expected

Based on the investigation plan, we expect to find:

1. **DNS Failures** - Should be retried with fallback servers
2. **Network Timeouts** - Should be retried with exponential backoff
3. **HTTP 4xx Errors** - Should NOT be retried (except 429)
4. **HTTP 5xx Errors** - Should be retried
5. **Context Cancellations** - Should be handled gracefully
6. **Parse Errors** - May need validation improvements

## Next Steps

1. Run error pattern parser on Railway logs
2. Analyze error distribution
3. Implement fixes based on findings
4. Test fixes with 50-sample E2E test

