# Merchant Details: Backend vs Frontend Feature Comparison

## Summary

This document provides a comprehensive comparison of all backend features available for merchant details pages against what is currently implemented in the frontend.

**Related Documents:**
- [Merchant vs Portfolio Level Features](./MERCHANT_VS_PORTFOLIO_LEVEL_FEATURES.md) - Detailed analysis of merchant-level vs portfolio-level features and comparison capabilities

**Last Updated:** 2025-01-27  
**Status:** Comprehensive Review

---

## Comparison Table

| Category | Backend Endpoint/Feature | Frontend Implementation | Status | Notes |
|----------|-------------------------|------------------------|--------|-------|
| **MERCHANT SERVICE - Core** |
| Merchant Data | `GET /api/v1/merchants/{id}` | ✅ `getMerchant()` | ✅ Implemented | Used in MerchantOverviewTab |
| Merchant Analytics | `GET /api/v1/merchants/{id}/analytics` | ✅ `getMerchantAnalytics()` | ✅ Implemented | Used in BusinessAnalyticsTab |
| Website Analysis | `GET /api/v1/merchants/{id}/website-analysis` | ✅ `getWebsiteAnalysis()` | ✅ Implemented | Used in BusinessAnalyticsTab |
| Risk Score | `GET /api/v1/merchants/{id}/risk-score` | ❌ Not implemented | ❌ Missing | Could enhance Overview tab |
| General Analytics | `GET /api/v1/merchants/analytics` | ❌ Not implemented | ⚠️ For comparison | Used by BI Dashboard; needed for merchant comparison |
| Statistics | `GET /api/v1/merchants/statistics` | ❌ Not implemented | ⚠️ For comparison | Used by BI Dashboard; needed for merchant comparison |
| Merchant Search | `POST /api/v1/merchants/search` | ❌ Not implemented | ⚠️ Not needed | Used in merchant list, not details page |
| Portfolio Types | `GET /api/v1/merchants/portfolio-types` | ❌ Not implemented | ⚠️ Not needed | Used in filters, not details page |
| Risk Levels | `GET /api/v1/merchants/risk-levels` | ❌ Not implemented | ⚠️ Not needed | Used in filters, not details page |
| **RISK ASSESSMENT SERVICE - Core** |
| Start Assessment | `POST /api/v1/risk/assess` | ✅ `startRiskAssessment()` | ✅ Implemented | Used in RiskAssessmentTab |
| Get Assessment | `GET /api/v1/risk/assess/{id}` | ✅ `getRiskAssessment()` | ✅ Implemented | Used in RiskAssessmentTab |
| Assessment Status | `GET /api/v1/risk/assess/{id}/status` | ✅ `getAssessmentStatus()` | ✅ Implemented | Used in RiskAssessmentTab |
| Risk History | `GET /api/v1/risk/assess/{id}/history` | ✅ `getRiskHistory()` | ✅ Implemented | Used in RiskAssessmentTab |
| Risk Predictions | `GET /api/v1/risk/predictions/{merchant_id}` | ✅ `getRiskPredictions()` | ✅ Implemented | Used in RiskAssessmentTab |
| Risk Indicators | `GET /api/v1/risk/indicators/{merchantId}` | ✅ `getRiskIndicators()` | ✅ Implemented | Used in RiskIndicatorsTab |
| Risk Alerts | `GET /api/v1/risk/alerts/{merchantId}` | ❌ Not implemented | ❌ Missing | Could enhance RiskIndicatorsTab |
| Risk Benchmarks | `GET /api/v1/risk/benchmarks` | ❌ Not implemented | ⚠️ For comparison | Used by Risk Dashboard; needed for merchant comparison |
| Risk Explainability | `GET /api/v1/risk/explain/{assessmentId}` | ✅ `explainRiskAssessment()` | ✅ Implemented | API function exists but not used in UI |
| Risk Recommendations | `GET /api/v1/risk/recommendations/{merchantId}` | ✅ `getRiskRecommendations()` | ✅ Implemented | API function exists but not used in UI |
| **RISK ASSESSMENT SERVICE - Advanced** |
| Batch Assessment | `POST /api/v1/risk/assess/batch` | ❌ Not implemented | ⚠️ Not needed | Batch operation, not for single merchant |
| Advanced Prediction | `POST /api/v1/risk/predict-advanced` | ❌ Not implemented | ❌ Missing | Could add advanced prediction UI |
| Model Info | `GET /api/v1/models/info` | ❌ Not implemented | ❌ Missing | Could show model information |
| Model Performance | `GET /api/v1/models/performance` | ❌ Not implemented | ❌ Missing | Could show model metrics |
| **RISK ASSESSMENT SERVICE - Compliance** |
| Compliance Check | `POST /api/v1/compliance/check` | ❌ Not implemented | ❌ Missing | Could add compliance tab/section |
| Sanctions Screening | `POST /api/v1/sanctions/screen` | ❌ Not implemented | ❌ Missing | Could add compliance tab/section |
| Adverse Media | `POST /api/v1/media/monitor` | ❌ Not implemented | ❌ Missing | Could add compliance tab/section |
| Compliance Status | `GET /api/v1/compliance/status/{business_id}` | ❌ Not implemented | ❌ Missing | Could add compliance tab/section |
| Compliance Status (Aggregate) | `GET /api/v1/compliance/status/aggregate` | ❌ Not implemented | ⚠️ Dashboard only | Used by Compliance Dashboard, not merchant details |
| **RISK ASSESSMENT SERVICE - Analytics** |
| Risk Trends | `GET /api/v1/analytics/trends` | ❌ Not implemented | ⚠️ For comparison | Used by Risk Dashboard; needed for merchant comparison |
| Risk Insights | `GET /api/v1/analytics/insights` | ❌ Not implemented | ⚠️ For comparison | Used by Risk Dashboard; needed for merchant comparison |
| **RISK ASSESSMENT SERVICE - External Data** |
| Company Data Lookup | `POST /api/v1/external/company-data` | ❌ Not implemented | ❌ Missing | Could add enrichment section |
| External Compliance | `POST /api/v1/external/compliance` | ❌ Not implemented | ❌ Missing | Could add enrichment section |
| External Data Sources | `GET /api/v1/external/sources` | ❌ Not implemented | ❌ Missing | Could show data sources |
| **ENRICHMENT SERVICE** |
| Enrichment Sources | `GET /api/v1/merchants/{id}/enrichment-sources` | ✅ `getEnrichmentSources()` | ✅ Implemented | API function exists but not used in UI |
| Trigger Enrichment | `POST /api/v1/merchants/{id}/enrichment` | ✅ `triggerEnrichment()` | ✅ Implemented | API function exists but not used in UI |
| **WEBSOCKET** |
| Risk WebSocket | `WS /api/v1/risk/ws` | ✅ `RiskWebSocketProvider` | ✅ Implemented | Used in RiskAssessmentTab |
| **DASHBOARD/METRICS** |
| Dashboard Metrics | `GET /api/v1/dashboard/metrics` | ✅ `getDashboardMetrics()` | ✅ Implemented | Not used in merchant details page |
| Risk Metrics | `GET /api/v1/risk/metrics` | ✅ `getRiskMetrics()` | ✅ Implemented | Not used in merchant details page |
| System Metrics | `GET /api/v1/system/metrics` | ✅ `getSystemMetrics()` | ✅ Implemented | Not used in merchant details page |
| Compliance Status | `GET /api/v1/compliance/status` | ✅ `getComplianceStatus()` | ✅ Implemented | Not used in merchant details page |
| Business Intelligence | `GET /api/v1/business-intelligence/metrics` | ✅ `getBusinessIntelligenceMetrics()` | ✅ Implemented | Not used in merchant details page |

---

## Frontend Components Status

### ✅ Fully Implemented Components

1. **MerchantOverviewTab**
   - ✅ Displays merchant basic information
   - ✅ Shows contact information
   - ✅ Shows address
   - ✅ Shows registration & compliance data
   - ✅ Shows metadata (ID, dates)
   - ✅ Uses shadcn UI components
   - ✅ Client-side date formatting (hydration fix)

2. **BusinessAnalyticsTab**
   - ✅ Fetches and displays analytics data
   - ✅ Shows classification information
   - ✅ Shows security metrics
   - ✅ Shows data quality metrics
   - ✅ **NEW:** Industry codes tables (MCC, SIC, NAICS)
   - ✅ **NEW:** Charts (classification confidence, industry distribution, security, data quality)
   - ✅ Fetches website analysis
   - ✅ Uses shadcn UI components
   - ✅ Error handling and loading states

3. **RiskAssessmentTab**
   - ✅ Fetches risk assessment data
   - ✅ Initiates new assessments
   - ✅ Shows assessment status and progress
   - ✅ **NEW:** Risk factors table
   - ✅ **NEW:** Risk history table
   - ✅ **NEW:** Risk score history chart (LineChart)
   - ✅ **NEW:** Risk factors comparison chart (BarChart)
   - ✅ **NEW:** Risk predictions chart (AreaChart)
   - ✅ WebSocket integration for real-time updates
   - ✅ Uses shadcn UI components
   - ✅ Error handling and loading states

4. **RiskIndicatorsTab**
   - ✅ Fetches risk indicators
   - ✅ **NEW:** Risk indicators table with sorting
   - ✅ **NEW:** Severity distribution chart (PieChart)
   - ✅ **NEW:** Grouped by severity view
   - ✅ Uses shadcn UI components
   - ✅ Error handling and loading states

### ⚠️ Partially Implemented (API exists but not used in UI)

1. **Risk Explainability**
   - ✅ API function: `explainRiskAssessment()`
   - ❌ Not displayed in UI
   - **Recommendation:** Add explainability section to RiskAssessmentTab

2. **Risk Recommendations**
   - ✅ API function: `getRiskRecommendations()`
   - ❌ Not displayed in UI
   - **Recommendation:** Add recommendations section to RiskAssessmentTab

3. **Enrichment Features**
   - ✅ API functions: `getEnrichmentSources()`, `triggerEnrichment()`
   - ❌ Not displayed in UI
   - **Recommendation:** Add enrichment button/section to MerchantOverviewTab or header

### ❌ Missing Features (Backend available but not implemented)

1. **Merchant Risk Score**
   - Backend: `GET /api/v1/merchants/{id}/risk-score`
   - **Recommendation:** Add to MerchantOverviewTab or create summary card

2. **Risk Alerts**
   - Backend: `GET /api/v1/risk/alerts/{merchantId}`
   - **Recommendation:** Add to RiskIndicatorsTab or create alerts section

3. **Risk Benchmarks**
   - Backend: `GET /api/v1/risk/benchmarks`
   - **Recommendation:** Add to RiskAssessmentTab to show industry benchmarks

4. **Risk Trends**
   - Backend: `GET /api/v1/analytics/trends`
   - **Recommendation:** Add to RiskAssessmentTab or create trends section

5. **Risk Insights**
   - Backend: `GET /api/v1/analytics/insights`
   - **Recommendation:** Add to RiskAssessmentTab or create insights section

6. **Compliance Features**
   - Backend: Multiple compliance endpoints
   - **Recommendation:** Create new ComplianceTab or add to existing tabs
   - Endpoints:
     - `POST /api/v1/compliance/check`
     - `POST /api/v1/sanctions/screen`
     - `POST /api/v1/media/monitor`
     - `GET /api/v1/compliance/status/{business_id}`

7. **External Data Sources**
   - Backend: Multiple external data endpoints
   - **Recommendation:** Add enrichment section showing data sources
   - Endpoints:
     - `POST /api/v1/external/company-data`
     - `POST /api/v1/external/compliance`
     - `GET /api/v1/external/sources`

8. **Advanced Prediction**
   - Backend: `POST /api/v1/risk/predict-advanced`
   - **Recommendation:** Add advanced prediction options to RiskAssessmentTab

9. **Model Information**
   - Backend: `GET /api/v1/models/info`, `GET /api/v1/models/performance`
   - **Recommendation:** Add model info section to RiskAssessmentTab

---

## Feature Coverage Summary

### ✅ Implemented: 15/35 (43%)
- Core merchant data ✅
- Merchant analytics ✅
- Website analysis ✅
- Risk assessment (basic) ✅
- Risk history ✅
- Risk predictions ✅
- Risk indicators ✅
- WebSocket updates ✅
- Enrichment API functions ✅ (but not used in UI)

### ⚠️ Partially Implemented: 3/35 (9%)
- Risk explainability (API only)
- Risk recommendations (API only)
- Enrichment (API only)

### ❌ Missing: 17/35 (48%)
- Merchant risk score
- Risk alerts
- Risk benchmarks
- Risk trends
- Risk insights
- Compliance features (4 endpoints)
- External data sources (3 endpoints)
- Advanced prediction
- Model information (2 endpoints)

---

## Recommendations

### High Priority (Should implement soon)

1. **Add Risk Alerts to RiskIndicatorsTab**
   - Use `GET /api/v1/risk/alerts/{merchantId}`
   - Show active alerts prominently
   - Add alert notifications

2. **Add Risk Benchmarks to RiskAssessmentTab**
   - Use `GET /api/v1/risk/benchmarks`
   - Show industry benchmarks for comparison
   - Add benchmark visualization

3. **Add Risk Explainability to RiskAssessmentTab**
   - Use existing `explainRiskAssessment()` function
   - Show SHAP values and feature importance
   - Add explainability section

4. **Add Risk Recommendations to RiskAssessmentTab**
   - Use existing `getRiskRecommendations()` function
   - Show actionable recommendations
   - Add recommendations section

5. **Add Enrichment UI**
   - Use existing `getEnrichmentSources()` and `triggerEnrichment()` functions
   - Add "Enrich Data" button to header or MerchantOverviewTab
   - Show enrichment sources and status

### Medium Priority (Nice to have)

6. **Add Compliance Tab**
   - Create new ComplianceTab component
   - Integrate compliance endpoints
   - Show compliance status, sanctions screening, adverse media

7. **Add Risk Trends and Insights**
   - Use `GET /api/v1/analytics/trends` and `GET /api/v1/analytics/insights`
   - Add to RiskAssessmentTab or create separate section
   - Show trend analysis and insights

8. **Add Merchant Risk Score**
   - Use `GET /api/v1/merchants/{id}/risk-score`
   - Add to MerchantOverviewTab as summary card
   - Show quick risk overview

9. **Add External Data Sources**
   - Use external data endpoints
   - Show data sources in enrichment section
   - Display data quality and sources

### Low Priority (Future enhancements)

10. **Advanced Prediction UI**
    - Use `POST /api/v1/risk/predict-advanced`
    - Add advanced options to RiskAssessmentTab
    - Allow custom prediction scenarios

11. **Model Information**
    - Use model info endpoints
    - Show model details and performance
    - Add model selection options

---

## Implementation Notes

### Current Architecture
- Frontend uses Next.js with React components
- All components use shadcn UI for consistency
- API functions are in `frontend/lib/api.ts`
- Components are in `frontend/components/merchant/`
- Lazy loading for tab content
- WebSocket for real-time updates

### Data Flow
1. Merchant details page loads → fetches merchant data
2. Tab components lazy load → fetch their specific data
3. WebSocket connects → provides real-time risk updates
4. Charts and tables render → display processed data

### Missing Integration Points
1. **Enrichment UI** - API functions exist but no UI component
2. **Compliance Tab** - No component for compliance features
3. **Risk Alerts** - Separate endpoint not integrated
4. **Risk Benchmarks** - Endpoint not called
5. **Risk Trends/Insights** - Analytics endpoints not used
6. **External Data Sources** - Endpoints not integrated

---

## Portfolio-Level vs Merchant-Level Architecture

**Note:** For detailed analysis of portfolio-level vs merchant-level features, see [MERCHANT_VS_PORTFOLIO_LEVEL_FEATURES.md](./MERCHANT_VS_PORTFOLIO_LEVEL_FEATURES.md)

### Architecture Overview

The platform uses a two-tier architecture:

1. **Portfolio-Level Pages (Dashboards):**
   - Business Intelligence Dashboard (`/dashboard`) - [View](https://frontend-service-production-b225.up.railway.app/dashboard)
   - Risk Dashboard (`/risk-dashboard`) - [View](https://frontend-service-production-b225.up.railway.app/risk-dashboard)
   - Risk Indicators Dashboard (`/risk-indicators`) - [View](https://frontend-service-production-b225.up.railway.app/risk-indicators)
   - **Purpose:** Display aggregate/portfolio-wide data and statistics
   - **Endpoints Used:** Portfolio-level endpoints (analytics, statistics, trends, insights)

2. **Merchant-Level Page (Details):**
   - Merchant Details Page (`/merchant-details/[id]`)
   - **Purpose:** Display individual merchant data with portfolio comparisons
   - **Endpoints Used:** 
     - Merchant-specific endpoints (primary data)
     - Portfolio-level endpoints (for comparison context)

### Quick Summary
- **Merchant-Level Features (Merchant Details Page):** 43% implemented (15/35 features)
- **Portfolio-Level Features (Dashboard Pages):** Status unknown (needs review)
- **Comparison Features (Merchant Details Page):** 0% implemented (0/5 features)

### Portfolio-Level Endpoints (For Dashboard Pages)
These endpoints should be used by the dashboard pages, not the merchant details page:
- `GET /api/v1/merchants/analytics` - Portfolio-wide analytics → **Business Intelligence Dashboard**
- `GET /api/v1/merchants/statistics` - Portfolio-wide statistics → **Business Intelligence Dashboard**
- `GET /api/v1/analytics/trends` - Portfolio risk trends → **Risk Dashboard**
- `GET /api/v1/analytics/insights` - Portfolio risk insights → **Risk Dashboard**
- `GET /api/v1/risk/metrics` - Portfolio risk metrics → **Risk Dashboard**
- `GET /api/v1/risk/indicators` (aggregate) - Portfolio risk indicators → **Risk Indicators Dashboard**

### Comparison Endpoints (For Merchant Details Page)
These endpoints should be fetched by the merchant details page to enable merchant-to-portfolio comparisons:
- `GET /api/v1/merchants/statistics` - Fetch portfolio averages for comparison
- `GET /api/v1/merchants/analytics` - Fetch portfolio analytics for comparison
- `GET /api/v1/risk/benchmarks?mcc={code}` - Fetch industry benchmarks for comparison
- `GET /api/v1/analytics/trends` - Fetch portfolio trends for context

### Comparison Capabilities Needed in Merchant Details Page
- ✅ Merchant risk score vs portfolio average (from statistics endpoint)
- ✅ Merchant risk score vs industry benchmarks (from benchmarks endpoint)
- ✅ Merchant analytics vs portfolio analytics (from analytics endpoints)
- ✅ Merchant position in portfolio distribution (from statistics endpoint)
- ✅ Percentile rankings (calculated from statistics + merchant data)

---

## Conclusion

**Current Status:** The frontend has good coverage of core merchant details features (43%), with all critical features implemented. However, there are significant opportunities to enhance the page with additional backend features:

1. **Risk Alerts** - High priority, easy to add
2. **Risk Benchmarks** - High priority, adds valuable context
3. **Compliance Features** - Medium priority, could be a new tab
4. **Enrichment UI** - High priority, API already exists
5. **Risk Explainability** - High priority, API already exists
6. **Risk Recommendations** - High priority, API already exists

**Recommendation:** Focus on implementing the "High Priority" items first, as they provide the most value with minimal effort (many API functions already exist).

---

## Dashboard Pages Review Needed

**Action Required:** Review the following dashboard pages to verify they are correctly using portfolio-level endpoints:

1. **Business Intelligence Dashboard** - [View](https://frontend-service-production-b225.up.railway.app/dashboard)
   - Should use: `GET /api/v1/merchants/analytics`
   - Should use: `GET /api/v1/merchants/statistics`
   - Verify: Portfolio-wide analytics are displayed correctly

2. **Risk Dashboard** - [View](https://frontend-service-production-b225.up.railway.app/risk-dashboard)
   - Should use: `GET /api/v1/analytics/trends`
   - Should use: `GET /api/v1/analytics/insights`
   - Should use: `GET /api/v1/risk/metrics`
   - Verify: Portfolio risk trends and distribution are displayed correctly

3. **Risk Indicators Dashboard** - [View](https://frontend-service-production-b225.up.railway.app/risk-indicators)
   - Should use: Aggregate risk indicators endpoints
   - Verify: Portfolio-wide risk indicators are displayed correctly

**Note:** These dashboard pages are the primary consumers of portfolio-level endpoints. The merchant details page should fetch these same endpoints for comparison purposes only.

---

**Document Version:** 1.0  
**Last Updated:** 2025-01-27  
**Next Review:** After implementing high-priority features

