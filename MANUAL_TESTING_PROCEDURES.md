# Manual Workflow Testing Procedures

## Overview
This document provides comprehensive manual testing procedures for the business intelligence system. These procedures should be followed step-by-step to ensure thorough testing of all workflows and functionality.

## Testing Environment
- **API Base URL**: http://localhost:8080
- **UI Base URL**: http://localhost:8081
- **Test Date**: September 11, 2025

## Prerequisites
- API server running on port 8080
- UI server running on port 8081
- Web browser for UI testing
- API testing tool (curl, Postman, or similar)

## Testing Workflows

### 1. Market Analysis Workflow Testing

#### Step 1: Test Market Analysis API Endpoint
Test the market analysis API endpoint with the following curl command:

```bash
curl -X POST http://localhost:8080/v2/business-intelligence/market-analysis \
  -H 'Content-Type: application/json' \
  -d '{
    "business_id": "test-business-123",
    "time_range": {
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-12-31T23:59:59Z"
    },
    "options": {
      "include_competitors": true,
      "include_trends": true,
      "include_forecasts": true
    }
  }'
```

**Expected Results:**
- HTTP status code: 200 (if implemented) or 501 (not implemented)
- Response should contain market analysis data
- Response time should be under 5 seconds

#### Step 2: Test Market Analysis UI
Open your web browser and navigate to:
`http://localhost:8081/market-analysis-dashboard.html`

**Manual UI Testing Checklist:**
- [ ] Page loads successfully
- [ ] All form fields are visible and functional
- [ ] Date pickers work correctly
- [ ] Submit button is clickable
- [ ] Error messages display appropriately
- [ ] Loading states are shown during processing
- [ ] Results are displayed in a readable format

#### Step 3: Test Market Analysis Job Creation
Test the job creation endpoint:

```bash
curl -X POST http://localhost:8080/v2/business-intelligence/market-analysis/jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "business_id": "test-business-123",
    "time_range": {
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-12-31T23:59:59Z"
    }
  }'
```

#### Step 4: Test Market Analysis Job Status
Check job status using the job ID from the previous response:

```bash
curl -X GET http://localhost:8080/v2/business-intelligence/market-analysis/jobs/{job_id}
```

#### Step 5: Test Market Analysis Results Retrieval
Retrieve analysis results:

```bash
curl -X GET http://localhost:8080/v2/business-intelligence/market-analysis/{analysis_id}
```

**Document your findings:**
- Record any errors or unexpected behavior
- Note response times and data accuracy
- Document UI usability issues

### 2. Competitive Analysis Workflow Testing

#### Step 1: Test Competitive Analysis API Endpoint
Test the competitive analysis API endpoint:

```bash
curl -X POST http://localhost:8080/v2/business-intelligence/competitive-analysis \
  -H 'Content-Type: application/json' \
  -d '{
    "business_id": "test-business-123",
    "competitors": ["competitor1", "competitor2", "competitor3"],
    "time_range": {
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-12-31T23:59:59Z"
    },
    "options": {
      "include_market_share": true,
      "include_pricing": true,
      "include_features": true
    }
  }'
```

#### Step 2: Test Competitive Analysis UI
Open your web browser and navigate to:
`http://localhost:8081/competitive-analysis-dashboard.html`

**Manual UI Testing Checklist:**
- [ ] Page loads successfully
- [ ] Competitor selection interface works
- [ ] Analysis options are configurable
- [ ] Results display competitor comparisons
- [ ] Charts and graphs render correctly
- [ ] Export functionality works (if available)

#### Step 3: Test Competitive Analysis Job Workflow
Follow the same job creation and status checking pattern as market analysis.

**Document your findings:**
- Record competitor data accuracy
- Note analysis depth and insights
- Document UI responsiveness and usability

### 3. Growth Analytics Workflow Testing

#### Step 1: Test Growth Analytics API Endpoint
Test the growth analytics API endpoint:

```bash
curl -X POST http://localhost:8080/v2/business-intelligence/growth-analytics \
  -H 'Content-Type: application/json' \
  -d '{
    "business_id": "test-business-123",
    "time_range": {
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-12-31T23:59:59Z"
    },
    "options": {
      "include_revenue": true,
      "include_customers": true,
      "include_metrics": true
    }
  }'
```

#### Step 2: Test Growth Analytics UI
Open your web browser and navigate to:
`http://localhost:8081/business-growth-analytics.html`

**Manual UI Testing Checklist:**
- [ ] Page loads successfully
- [ ] Growth metrics are displayed clearly
- [ ] Time series charts render correctly
- [ ] Trend analysis is accurate
- [ ] Forecasting features work
- [ ] Data export functionality works

#### Step 3: Test Growth Analytics Job Workflow
Follow the same job creation and status checking pattern.

**Document your findings:**
- Record growth metric accuracy
- Note trend analysis quality
- Document forecasting reliability

### 4. Error Handling and Edge Cases Testing

#### Step 1: Test Invalid Input Handling
Test with invalid JSON:

```bash
curl -X POST http://localhost:8080/v2/business-intelligence/market-analysis \
  -H 'Content-Type: application/json' \
  -d 'invalid json'
```

**Expected:** HTTP 400 Bad Request

#### Step 2: Test Missing Required Fields
Test with missing business_id:

```bash
curl -X POST http://localhost:8080/v2/business-intelligence/market-analysis \
  -H 'Content-Type: application/json' \
  -d '{
    "time_range": {
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-12-31T23:59:59Z"
    }
  }'
```

**Expected:** HTTP 400 Bad Request with validation error

#### Step 3: Test Invalid Date Ranges
Test with invalid date range:

```bash
curl -X POST http://localhost:8080/v2/business-intelligence/market-analysis \
  -H 'Content-Type: application/json' \
  -d '{
    "business_id": "test-business-123",
    "time_range": {
      "start_date": "2024-12-31T23:59:59Z",
      "end_date": "2024-01-01T00:00:00Z"
    }
  }'
```

**Expected:** HTTP 400 Bad Request with date validation error

#### Step 4: Test Non-existent Resource Access
Test accessing non-existent analysis:

```bash
curl -X GET http://localhost:8080/v2/business-intelligence/market-analysis/non-existent-id
```

**Expected:** HTTP 404 Not Found

#### Step 5: Test Rate Limiting
Send multiple rapid requests to test rate limiting:

```bash
for i in {1..10}; do
  curl -X POST http://localhost:8080/v2/business-intelligence/market-analysis \
    -H 'Content-Type: application/json' \
    -d "{\"business_id\": \"test-$i\"}" &
done
wait
```

**Expected:** Some requests should be rate limited (HTTP 429)

**Document your findings:**
- Record error message quality and helpfulness
- Note error handling consistency across endpoints
- Document any security vulnerabilities found

### 5. User Interface Navigation Testing

#### Step 1: Test Main Dashboard Navigation
Open your web browser and navigate to:
`http://localhost:8081/dashboard.html`

**Navigation Testing Checklist:**
- [ ] Main dashboard loads successfully
- [ ] Navigation menu is visible and functional
- [ ] All dashboard sections are accessible
- [ ] Links to business intelligence modules work
- [ ] Responsive design works on different screen sizes

#### Step 2: Test Business Intelligence Module Navigation
Test navigation between BI modules:

1. **Market Analysis Dashboard**: `http://localhost:8081/market-analysis-dashboard.html`
2. **Competitive Analysis Dashboard**: `http://localhost:8081/competitive-analysis-dashboard.html`
3. **Growth Analytics Dashboard**: `http://localhost:8081/business-growth-analytics.html`

**Navigation Testing Checklist:**
- [ ] All modules load without errors
- [ ] Navigation between modules is smooth
- [ ] Breadcrumbs or back navigation works
- [ ] Page titles and headers are correct
- [ ] Loading states are appropriate

#### Step 3: Test Form Interactions
Test form functionality across all modules:

**Form Testing Checklist:**
- [ ] All input fields accept data correctly
- [ ] Date pickers work and validate dates
- [ ] Dropdown selections work properly
- [ ] Checkboxes and radio buttons function
- [ ] Form validation provides clear feedback
- [ ] Submit buttons trigger appropriate actions
- [ ] Reset/clear functionality works

#### Step 4: Test Data Display and Visualization
Test how data is presented to users:

**Data Display Testing Checklist:**
- [ ] Tables display data clearly and are sortable
- [ ] Charts and graphs render correctly
- [ ] Data is formatted appropriately (dates, numbers, currency)
- [ ] Empty states are handled gracefully
- [ ] Loading states are shown during data fetching
- [ ] Error states display helpful messages

#### Step 5: Test Accessibility
Test accessibility features:

**Accessibility Testing Checklist:**
- [ ] Page can be navigated using keyboard only
- [ ] Screen reader compatibility (if available)
- [ ] Color contrast is sufficient
- [ ] Text is readable at different zoom levels
- [ ] Form labels are properly associated
- [ ] Error messages are accessible

**Document your findings:**
- Record any navigation issues or confusion
- Note usability problems or improvements needed
- Document accessibility barriers
- Record performance issues or slow loading

## Testing Report Template

### Issues Found
(To be filled in during manual testing)

### Recommendations
(To be filled in during manual testing)

### Overall Assessment
(To be filled in after completing all manual tests)

### Next Steps
1. Complete all manual testing procedures
2. Document all findings and issues
3. Prioritize issues for resolution
4. Plan improvements based on findings
5. Schedule follow-up testing after fixes

## Testing Completion Checklist

- [ ] Market Analysis Workflow Testing completed
- [ ] Competitive Analysis Workflow Testing completed
- [ ] Growth Analytics Workflow Testing completed
- [ ] Error Handling and Edge Cases Testing completed
- [ ] User Interface Navigation Testing completed
- [ ] All findings documented
- [ ] Issues prioritized for resolution
- [ ] Recommendations provided
- [ ] Overall assessment completed

## Notes
- Record all observations, issues, and recommendations during testing
- Take screenshots of any UI issues or unexpected behavior
- Document response times and performance observations
- Note any security concerns or vulnerabilities
- Provide specific examples of any problems encountered
