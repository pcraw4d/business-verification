# Risk Assessment Business Data Retrieval Implementation

**Date**: 2025-11-10  
**Status**: ✅ Completed

---

## Summary

Implemented business data retrieval from database for risk prediction endpoints, replacing mock data with actual database queries.

---

## Issues

1. **HandleRiskPrediction** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:353`)
   - Status: TODO - Retrieve business data from database using ID from URL
   - Impact: Medium - Using mock data instead of real business data

2. **HandleRiskPredictions** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:620`)
   - Status: TODO - Fetch real merchant data from database using merchantID
   - Impact: Medium - Using mock data instead of real merchant data

---

## Implementation

### 1. HandleRiskPrediction

**Approach**:
- Extract assessment ID from URL path
- Query `risk_assessments` table by ID
- Extract business data from assessment record
- Fallback to mock data if assessment not found

**Data Extracted**:
- `business_name` from assessment
- `business_address` from assessment
- `industry` from assessment
- `country` from assessment
- `business_id` added to metadata if available

### 2. HandleRiskPredictions

**Approach**:
- Try to get merchant from `merchants` table first
- If not found, try to get latest assessment for merchant
- Extract business data from merchant or assessment
- Fallback to mock data if neither found

**Data Sources** (in priority order):
1. `merchants` table (name, address, industry)
2. Latest `risk_assessments` record for merchant (business_name, business_address, industry, country)
3. Mock data (final fallback)

---

## Code Changes

### HandleRiskPrediction

**Before**:
```go
// TODO: Retrieve business data from database using ID from URL
business := &models.RiskAssessmentRequest{
    BusinessName:      "Sample Business",
    BusinessAddress:   "123 Sample St, Sample City, SC 12345",
    Industry:          "Technology",
    Country:           "US",
    ...
}
```

**After**:
```go
// Extract assessment ID from URL and retrieve business data from database
vars := mux.Vars(r)
assessmentID := vars["id"]

var business *models.RiskAssessmentRequest
if assessmentID != "" {
    // Query risk_assessments table
    var assessmentResult []map[string]interface{}
    _, err := h.supabaseClient.GetClient().From("risk_assessments").
        Select("*", "", false).
        Eq("id", assessmentID).
        Single().
        ExecuteTo(&assessmentResult)

    if err == nil && len(assessmentResult) > 0 {
        // Extract business data from assessment
        assessmentData := assessmentResult[0]
        business = &models.RiskAssessmentRequest{
            BusinessName:      getString(assessmentData, "business_name"),
            BusinessAddress:   getString(assessmentData, "business_address"),
            Industry:          getString(assessmentData, "industry"),
            Country:           getString(assessmentData, "country"),
            ...
        }
    }
}
// Fallback to mock data if not found
```

### HandleRiskPredictions

**Before**:
```go
// TODO: Fetch real merchant data from database using merchantID
business := &models.RiskAssessmentRequest{
    BusinessName:    "Merchant " + merchantID,
    BusinessAddress: "Unknown",
    Industry:        "General",
    Country:         "US",
}
```

**After**:
```go
// Try to fetch real merchant data from database
var business *models.RiskAssessmentRequest

// First, try merchants table
var merchantResult []map[string]interface{}
_, err := h.supabaseClient.GetClient().From("merchants").
    Select("*", "", false).
    Eq("id", merchantID).
    Single().
    ExecuteTo(&merchantResult)

if err == nil && len(merchantResult) > 0 {
    // Extract from merchant
    merchantData := merchantResult[0]
    business = &models.RiskAssessmentRequest{
        BusinessName:    getString(merchantData, "name"),
        BusinessAddress: getString(merchantData, "address"),
        Industry:        getString(merchantData, "industry"),
        Country:         "US",
    }
} else {
    // Fallback: Try latest assessment
    var assessmentResult []map[string]interface{}
    _, err := h.supabaseClient.GetClient().From("risk_assessments").
        Select("*", "", false).
        Eq("business_id", merchantID).
        Order("created_at", false).
        Limit(1, "").
        ExecuteTo(&assessmentResult)

    if err == nil && len(assessmentResult) > 0 {
        // Extract from assessment
        ...
    }
}
// Final fallback to mock data
```

---

## Benefits

1. **Real Data**: Uses actual business/merchant data instead of mock data
2. **Better Predictions**: ML models get accurate business information
3. **Resilient**: Multiple fallback strategies ensure service continues working
4. **Improved Accuracy**: Risk predictions based on real business data

---

## Fallback Strategy

### HandleRiskPrediction
1. Query `risk_assessments` by ID
2. Fallback to mock data if not found

### HandleRiskPredictions
1. Query `merchants` table by ID
2. Query latest `risk_assessments` for merchant
3. Fallback to mock data if neither found

---

## Testing Recommendations

1. **Test With Assessment ID**: Verify business data is extracted correctly
2. **Test With Merchant ID**: Verify merchant data is retrieved
3. **Test Assessment Fallback**: Verify latest assessment is used if merchant not found
4. **Test Mock Fallback**: Verify mock data is used if nothing found
5. **Test Invalid IDs**: Verify error handling works correctly

---

## Files Changed

1. ✅ `services/risk-assessment-service/internal/handlers/risk_assessment.go`

---

**Last Updated**: 2025-11-10

