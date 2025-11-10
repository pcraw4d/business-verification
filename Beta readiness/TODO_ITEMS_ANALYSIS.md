# TODO Items Analysis

**Date**: 2025-11-10  
**Status**: Analysis Complete

---

## Summary

Analysis of TODO items across the codebase to identify which can be fixed programmatically vs require manual work or are acceptable for beta.

---

## TODO Items by Category

### ✅ Acceptable for Beta (Can Defer)

#### Risk Assessment Service - Not Implemented Endpoints
These endpoints return `501 Not Implemented` which is acceptable for beta:

1. **HandleComplianceCheck** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:577`)
   - Status: Returns 501 Not Implemented
   - Impact: Low - Feature not required for beta
   - Action: Document as future feature

2. **HandleSanctionsScreening** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:583`)
   - Status: Returns 501 Not Implemented
   - Impact: Low - Feature not required for beta
   - Action: Document as future feature

3. **HandleAdverseMediaMonitoring** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:589`)
   - Status: Returns 501 Not Implemented
   - Impact: Low - Feature not required for beta
   - Action: Document as future feature

4. **HandleRiskTrends** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:595`)
   - Status: Returns 501 Not Implemented
   - Impact: Low - Feature not required for beta
   - Action: Document as future feature

5. **HandleRiskInsights** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:601`)
   - Status: Returns 501 Not Implemented
   - Impact: Low - Feature not required for beta
   - Action: Document as future feature

6. **HandleRiskHistory** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:285`)
   - Status: Returns 501 Not Implemented
   - Impact: Low - Feature not required for beta
   - Action: Document as future feature

#### Frontend - UI Features
7. **Alert Acknowledgment** (`services/frontend/public/js/components/merchant-risk-indicators-tab.js:513`)
   - Status: TODO comment
   - Impact: Low - UI enhancement
   - Action: Document as future feature

8. **Alert Investigation** (`services/frontend/public/js/components/merchant-risk-indicators-tab.js:520`)
   - Status: TODO comment
   - Impact: Low - UI enhancement
   - Action: Document as future feature

9. **Recommendation Dismissal** (`services/frontend/public/js/components/merchant-risk-indicators-tab.js:532`)
   - Status: TODO comment
   - Impact: Low - UI enhancement
   - Action: Document as future feature

10. **Recommendation Implementation** (`services/frontend/public/js/components/merchant-risk-indicators-tab.js:539`)
    - Status: TODO comment
    - Impact: Low - UI enhancement
    - Action: Document as future feature

#### External Integrations
11. **Thomson Reuters Client** (`services/risk-assessment-service/internal/external/thomson_reuters/client.go`)
    - Status: Multiple TODO comments - Real API not implemented
    - Impact: Medium - External integration
    - Action: Document as future integration, mock client is working

12. **World-Check Client** (`services/risk-assessment-service/internal/external/thomson_reuters/client.go`)
    - Status: Multiple TODO comments - Real API not implemented
    - Impact: Medium - External integration
    - Action: Document as future integration, mock client is working

---

### ⚠️ Should Address (Medium Priority)

#### Risk Assessment Service - Monitoring
1. **Monitoring Configuration Loading** (`services/risk-assessment-service/cmd/main.go:300`)
   - Status: TODO - Disabled for service startup
   - Impact: Medium - Limited observability
   - Action: Implement proper configuration loading

2. **Alert Rules Configuration** (`services/risk-assessment-service/cmd/main.go:313`)
   - Status: TODO - Disabled for service startup
   - Impact: Medium - Limited alerting
   - Action: Implement proper alert rules configuration

3. **Monitoring Config Structure** (`services/risk-assessment-service/cmd/main.go:963`)
   - Status: TODO - Fix monitoring config structure
   - Impact: Medium - Configuration issues
   - Action: Fix config structure

4. **Interface Adapters** (`services/risk-assessment-service/cmd/main.go:953`)
   - Status: TODO - Implement proper interface adapters
   - Impact: Medium - Code quality
   - Action: Implement adapters for cache, pool, and query components

#### Risk Assessment Service - Data
5. **Get Risk Assessment by ID** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:173`)
   - Status: TODO - Not implemented
   - Impact: Medium - Missing functionality
   - Action: Implement if needed for beta

6. **Data Points Count** (`services/risk-assessment-service/internal/handlers/risk_assessment.go:563`)
   - Status: TODO - Get actual count from database
   - Impact: Low - Currently returns 0
   - Action: Implement database query

#### Merchant Service
7. **CreatedBy Field** (`services/merchant-service/internal/handlers/merchant.go:340`)
   - Status: TODO - Get from auth context
   - Impact: Low - Currently hardcoded to "system"
   - Action: Extract from auth context

8. **Save to Supabase** (`services/merchant-service/internal/handlers/merchant.go:345`)
   - Status: TODO comment
   - Impact: Unknown - Need to verify if actually saving
   - Action: Verify implementation

---

### 🔴 High Priority (Should Fix)

None identified - All high-priority items have been addressed.

---

## Recommendations

### For Beta Release
1. **Document Not Implemented Features**: Create a "Known Limitations" document
2. **Accept 501 Responses**: These are acceptable for beta
3. **Monitor TODO Items**: Track for post-beta implementation

### Post-Beta
1. **Implement Monitoring Configuration**: Improve observability
2. **Complete External Integrations**: Thomson Reuters, World-Check
3. **Implement Missing Handlers**: Risk history, compliance check, etc.

---

## Action Items

1. ✅ Document acceptable TODOs for beta
2. ⏳ Create "Known Limitations" document
3. ⏳ Implement monitoring configuration loading
4. ⏳ Fix monitoring config structure
5. ⏳ Verify merchant service Supabase saving

---

**Last Updated**: 2025-11-10

