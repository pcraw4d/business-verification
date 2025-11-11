# Beta Testing Guide

**Version**: 1.0  
**Date**: 2025-01-27  
**Status**: Ready for Beta Testing

---

## Overview

This guide provides comprehensive instructions for beta testing the KYB Platform. It includes test scenarios, known issues, feedback collection processes, and rollback procedures.

---

## Beta Testing Objectives

1. **Validate Core Functionality**: Ensure all critical features work as expected
2. **Identify Bugs**: Find and document any issues or unexpected behavior
3. **Performance Testing**: Verify system performance under normal usage
4. **User Experience**: Gather feedback on usability and workflow
5. **Security Validation**: Confirm security measures are working correctly

---

## Test Environment

### Access Information
- **API Gateway URL**: `https://api-gateway-production.up.railway.app`
- **Frontend URL**: `https://frontend-production.up.railway.app`
- **Documentation**: See deployment documentation for full URLs

### Credentials
- Beta testers will receive credentials via secure channel
- Use provided test accounts for testing
- Do not use production data

---

## Test Scenarios

### 1. Merchant Verification Flow

**Objective**: Test the complete merchant verification process

**Steps**:
1. Navigate to "Add Merchant" page
2. Fill in merchant information:
   - Business Name: "Test Company Inc"
   - Legal Name: "Test Company Incorporated"
   - Address: "123 Test Street, Test City, TS 12345"
   - Industry: "Retail"
   - Website: "https://testcompany.com"
3. Submit the form
4. Verify redirect to merchant details page
5. Check that all three analyses completed:
   - Business Intelligence
   - Risk Assessment
   - Risk Indicators
6. Verify data persistence (refresh page)

**Expected Results**:
- ✅ Form submits successfully
- ✅ Redirect works correctly
- ✅ All analyses complete within 30 seconds
- ✅ Data persists after page refresh
- ✅ No console errors

**Known Issues**:
- If APIs timeout, data may still be available on refresh
- Some analyses may take up to 35 seconds

---

### 2. Merchant Listing and Filtering

**Objective**: Test merchant listing with pagination, filtering, and sorting

**Steps**:
1. Navigate to merchant list page
2. Test pagination:
   - Click "Next" button
   - Click "Previous" button
   - Verify page numbers update
3. Test filtering:
   - Filter by Portfolio Type
   - Filter by Risk Level
   - Filter by Status
   - Use search query
4. Test sorting:
   - Sort by Name (ascending/descending)
   - Sort by Created Date (ascending/descending)
   - Sort by Risk Level
5. Combine filters and sorting

**Expected Results**:
- ✅ Pagination works correctly
- ✅ Filters apply correctly
- ✅ Sorting works for all fields
- ✅ Combined filters work together
- ✅ Results update without page reload

---

### 3. Risk Assessment

**Objective**: Test risk assessment functionality

**Steps**:
1. Navigate to risk assessment page
2. Enter business information
3. Submit assessment request
4. Wait for assessment to complete
5. Review risk score and factors
6. Check risk recommendations

**Expected Results**:
- ✅ Assessment completes successfully
- ✅ Risk score is calculated correctly
- ✅ Risk factors are displayed
- ✅ Recommendations are provided
- ✅ Response time < 10 seconds

---

### 4. Classification Service

**Objective**: Test business classification

**Steps**:
1. Submit classification request via API or UI
2. Provide business name and description
3. Wait for classification result
4. Verify industry codes (MCC, SIC, NAICS)
5. Check confidence score

**Expected Results**:
- ✅ Classification completes successfully
- ✅ Industry codes are accurate
- ✅ Confidence score is provided
- ✅ Response time < 5 seconds
- ✅ Caching works (second request faster)

---

### 5. API Endpoints

**Objective**: Test all API endpoints

**Endpoints to Test**:
- `GET /health` - Health check
- `POST /api/v1/classify` - Classification
- `GET /api/v1/merchants` - List merchants
- `POST /api/v1/merchants` - Create merchant
- `GET /api/v1/merchants/{id}` - Get merchant
- `POST /api/v1/assess` - Risk assessment

**Test Cases**:
1. **Valid Requests**: Test with valid data
2. **Invalid Requests**: Test with missing/invalid data
3. **Error Handling**: Verify error responses
4. **Rate Limiting**: Test rate limit behavior
5. **Authentication**: Test protected endpoints

**Expected Results**:
- ✅ Valid requests return 200/201
- ✅ Invalid requests return 400 with error details
- ✅ Error responses are consistent
- ✅ Rate limiting works correctly
- ✅ Authentication enforced on protected endpoints

---

### 6. Security Testing

**Objective**: Verify security measures

**Test Cases**:
1. **Input Validation**: Try SQL injection, XSS attempts
2. **Authentication**: Test without tokens, with invalid tokens
3. **CORS**: Test cross-origin requests
4. **Security Headers**: Verify headers are present
5. **Rate Limiting**: Test rate limit enforcement

**Expected Results**:
- ✅ SQL injection attempts are sanitized
- ✅ XSS attempts are blocked
- ✅ Unauthenticated requests are rejected
- ✅ CORS is configured correctly
- ✅ Security headers are present
- ✅ Rate limiting prevents abuse

---

### 7. Performance Testing

**Objective**: Verify system performance

**Test Cases**:
1. **Response Times**: Measure API response times
2. **Concurrent Requests**: Test with multiple simultaneous requests
3. **Large Data Sets**: Test with large merchant lists
4. **Caching**: Verify caching improves performance

**Expected Results**:
- ✅ Health check < 100ms
- ✅ Classification < 5 seconds (first request)
- ✅ Classification < 100ms (cached)
- ✅ Merchant list < 2 seconds
- ✅ System handles 10+ concurrent requests

---

### 8. Error Scenarios

**Objective**: Test error handling

**Test Cases**:
1. **Network Errors**: Simulate network failures
2. **Timeout Errors**: Test with slow responses
3. **Invalid Data**: Submit malformed requests
4. **Missing Data**: Submit requests with missing fields
5. **Service Unavailable**: Test when services are down

**Expected Results**:
- ✅ Errors are handled gracefully
- ✅ Error messages are clear and helpful
- ✅ System recovers from errors
- ✅ No data corruption on errors

---

## Known Issues

### Current Limitations

1. **API Timeouts**:
   - Some API calls may timeout after 35 seconds
   - Workaround: Refresh page, data may be available

2. **Caching**:
   - Classification results cached for 5 minutes
   - May see stale data if updated within cache window

3. **Rate Limiting**:
   - Rate limit: 1000 requests per hour per IP
   - May need to wait if limit exceeded

4. **Browser Compatibility**:
   - Tested on Chrome, Firefox, Safari (latest versions)
   - Older browsers may have issues

---

## Feedback Collection

### How to Report Issues

1. **Bug Reports**:
   - Use provided bug report template
   - Include steps to reproduce
   - Attach screenshots if applicable
   - Include browser/OS information

2. **Feature Requests**:
   - Use feature request template
   - Describe use case
   - Explain expected behavior

3. **Performance Issues**:
   - Include response times
   - Describe conditions (network, load)
   - Include relevant logs if available

### Feedback Channels

- **Email**: beta-feedback@kyb-platform.com
- **Issue Tracker**: [Link to issue tracker]
- **Slack Channel**: #beta-testing

### Feedback Template

```
**Issue Type**: [Bug/Feature Request/Performance]
**Severity**: [Critical/High/Medium/Low]
**Description**: 
**Steps to Reproduce**:
1. 
2. 
3. 
**Expected Behavior**:
**Actual Behavior**:
**Screenshots**: [if applicable]
**Browser/OS**: 
**Additional Notes**:
```

---

## Rollback Plan

### If Critical Issues Found

1. **Immediate Actions**:
   - Notify all beta testers
   - Disable affected features if possible
   - Document issue details

2. **Rollback Procedure**:
   - Revert to previous stable version
   - Restore database backup if needed
   - Verify system stability

3. **Communication**:
   - Send notification to all testers
   - Provide status updates
   - Estimate resolution time

### Rollback Checklist

- [ ] Identify affected services
- [ ] Backup current state
- [ ] Revert code changes
- [ ] Restore database if needed
- [ ] Verify system functionality
- [ ] Notify testers
- [ ] Document rollback reason

---

## Testing Schedule

### Week 1: Core Functionality
- Merchant verification flow
- Basic API endpoints
- Error handling

### Week 2: Advanced Features
- Filtering and sorting
- Performance testing
- Security testing

### Week 3: Edge Cases
- Error scenarios
- Stress testing
- Browser compatibility

### Week 4: Final Review
- Complete all test scenarios
- Submit final feedback
- Review and prioritize issues

---

## Success Criteria

### Beta Testing Success Metrics

1. **Functionality**: 
   - ✅ 95% of test scenarios pass
   - ✅ Critical bugs identified and documented

2. **Performance**:
   - ✅ 90% of requests complete within SLA
   - ✅ No critical performance issues

3. **Stability**:
   - ✅ System uptime > 99%
   - ✅ No data loss incidents

4. **User Experience**:
   - ✅ Positive feedback from 80% of testers
   - ✅ Usability issues documented

---

## Support

### Getting Help

- **Documentation**: See API documentation and deployment guides
- **Slack**: #beta-testing channel
- **Email**: beta-support@kyb-platform.com

### Response Times

- **Critical Issues**: 2 hours
- **High Priority**: 24 hours
- **Medium/Low Priority**: 48 hours

---

## Conclusion

This beta testing phase is crucial for identifying issues and gathering feedback before production launch. Please test thoroughly and report all issues, no matter how minor they may seem.

**Thank you for participating in the beta test!**

---

**Last Updated**: 2025-01-27  
**Version**: 1.0

