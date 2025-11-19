# Merchant-Level vs Portfolio-Level Features Comparison

## Summary

This document identifies and compares merchant-specific features vs portfolio/aggregate-level features in the backend, and clarifies which endpoints belong to which pages in the platform architecture.

**Architecture:**
- **Portfolio-Level Pages:** Business Intelligence Dashboard, Risk Dashboard, Risk Indicators Dashboard
- **Merchant-Level Page:** Merchant Details Page (with portfolio comparisons)

**Last Updated:** 2025-01-27  
**Status:** Comprehensive Analysis - Updated for Dashboard Architecture

---

## Architecture Overview

### Portfolio-Level Pages (Dashboards)
These pages display aggregate/portfolio-wide data:
- **Business Intelligence Dashboard** (`/dashboard`) - [View](https://frontend-service-production-b225.up.railway.app/dashboard)
  - Portfolio-wide analytics
  - Business trends
  - Merchant distribution
  - Analytics overview
- **Risk Dashboard** (`/risk-dashboard`) - [View](https://frontend-service-production-b225.up.railway.app/risk-dashboard)
  - Portfolio risk trends
  - Risk distribution
  - Risk assessment details
- **Risk Indicators Dashboard** (`/risk-indicators`) - [View](https://frontend-service-production-b225.up.railway.app/risk-indicators)
  - Overall portfolio risk gauge
  - Portfolio risk trends
  - Risk category breakdown
  - Active risk indicators across portfolio

### Merchant-Level Page (Details with Comparisons)
- **Merchant Details Page** (`/merchant-details/[id]`)
  - Individual merchant data (primary)
  - Merchant-to-portfolio comparisons (context)
  - Merchant-to-industry benchmarks (context)

## Feature Classification

### Merchant-Level Features (Single Merchant)
These endpoints provide data specific to a single merchant identified by `{id}`.
**Used By:** Merchant Details Page (primary data source)

### Portfolio-Level Features (Aggregate/Aggregated)
These endpoints provide aggregated data across all merchants or specific merchant groups.
**Used By:** 
- Dashboard pages (primary data source)
- Merchant Details Page (for comparison context)

### Comparison/Benchmark Features
These endpoints enable comparing merchant data against portfolio averages or industry benchmarks.
**Used By:** Merchant Details Page (to show merchant in context)

---

## Comparison Table

| Feature Category | Merchant-Level Endpoint | Portfolio-Level Endpoint | Used By Dashboard | Used By Merchant Details | Comparison Status |
|-----------------|------------------------|-------------------------|-------------------|------------------------|-------------------|
| **MERCHANT DATA** |
| Basic Information | `GET /api/v1/merchants/{id}` | `GET /api/v1/merchants` (list) | ❌ No | ✅ Yes (primary) | N/A |
| Analytics | `GET /api/v1/merchants/{id}/analytics` | `GET /api/v1/merchants/analytics` | ✅ BI Dashboard | ✅ Yes (primary + comparison) | ❌ Not implemented |
| Statistics | N/A | `GET /api/v1/merchants/statistics` | ✅ BI Dashboard | ⚠️ For comparison only | ❌ Not implemented |
| Risk Score | `GET /api/v1/merchants/{id}/risk-score` | Portfolio avg in statistics | ✅ Risk Dashboard | ⚠️ For comparison only | ❌ Not implemented |
| Website Analysis | `GET /api/v1/merchants/{id}/website-analysis` | N/A | ❌ No | ✅ Yes (primary) | N/A |
| **RISK ASSESSMENT** |
| Risk Assessment | `GET /api/v1/risk/assess/{id}` | N/A | ❌ No | ✅ Yes (primary) | N/A |
| Risk History | `GET /api/v1/risk/assess/{id}/history` | N/A | ❌ No | ✅ Yes (primary) | N/A |
| Risk Predictions | `GET /api/v1/risk/predictions/{merchant_id}` | N/A | ❌ No | ✅ Yes (primary) | N/A |
| Risk Indicators | `GET /api/v1/risk/indicators/{merchantId}` | Aggregate indicators | ✅ Risk Indicators Dashboard | ✅ Yes (primary) | ❌ Not implemented |
| Risk Benchmarks | N/A | `GET /api/v1/risk/benchmarks?mcc={code}` | ⚠️ Risk Dashboard | ⚠️ For comparison only | ❌ Not implemented |
| Risk Trends | N/A | `GET /api/v1/analytics/trends` | ✅ Risk Dashboard | ⚠️ For comparison only | ❌ Not implemented |
| Risk Insights | N/A | `GET /api/v1/analytics/insights` | ✅ Risk Dashboard | ⚠️ For comparison only | ❌ Not implemented |
| Risk Metrics | N/A | `GET /api/v1/risk/metrics` | ✅ Risk Dashboard | ⚠️ For comparison only | ❌ Not implemented |
| **PORTFOLIO METADATA** |
| Portfolio Types | N/A | `GET /api/v1/merchants/portfolio-types` | ✅ BI Dashboard | ❌ No | N/A |
| Risk Levels | N/A | `GET /api/v1/merchants/risk-levels` | ✅ Risk Dashboard | ❌ No | N/A |

---

## Detailed Feature Analysis

### 1. Merchant Analytics vs Portfolio Analytics

#### Merchant-Level: `GET /api/v1/merchants/{id}/analytics`
**Purpose:** Get analytics for a specific merchant
**Returns:**
- Merchant-specific classification data
- Merchant-specific security metrics
- Merchant-specific data quality metrics
**Frontend:** ✅ Implemented in `BusinessAnalyticsTab`

#### Portfolio-Level: `GET /api/v1/merchants/analytics`
**Purpose:** Get aggregate analytics across all merchants
**Returns:**
```json
{
  "total_merchants": 1250,
  "active_merchants": 1180,
  "new_merchants_this_month": 45,
  "merchants_by_risk_level": {
    "low": 850,
    "medium": 320,
    "high": 80
  },
  "merchants_by_portfolio_type": {
    "retail": 450,
    "ecommerce": 380,
    "services": 320,
    "manufacturing": 100
  },
  "revenue_analytics": {
    "total_revenue": 1250000.00,
    "average_revenue_per_merchant": 1000.00,
    "revenue_growth_rate": 15.2
  },
  "performance_metrics": {
    "average_processing_time": "2.5s",
    "success_rate": 99.2,
    "error_rate": 0.8
  }
}
```
**Frontend:** ❌ Not implemented
**Use Case:** Show portfolio-wide metrics for comparison

---

### 2. Merchant Statistics vs Portfolio Statistics

#### Merchant-Level: N/A
**Note:** No merchant-specific statistics endpoint exists. Statistics are embedded in merchant data.

#### Portfolio-Level: `GET /api/v1/merchants/statistics`
**Purpose:** Get comprehensive portfolio statistics
**Returns:**
```json
{
  "overview": {
    "total_merchants": 1250,
    "active_merchants": 1180,
    "inactive_merchants": 70,
    "pending_verification": 25
  },
  "geographic_distribution": {
    "North America": 450,
    "Europe": 380,
    "Asia": 250,
    "Other": 170
  },
  "industry_breakdown": {
    "Retail": 320,
    "Technology": 280,
    "Financial Services": 200,
    "Healthcare": 150,
    "Manufacturing": 100,
    "Other": 200
  },
  "risk_assessment_stats": {
    "low_risk": 850,
    "medium_risk": 320,
    "high_risk": 80,
    "average_risk_score": 0.25
  },
  "verification_stats": {
    "verified": 1150,
    "pending": 25,
    "rejected": 75,
    "verification_success_rate": 93.8
  }
}
```
**Frontend:** ❌ Not implemented
**Use Case:** 
- Show portfolio overview
- Compare merchant risk score vs portfolio average
- Show merchant's position in portfolio distribution

---

### 3. Risk Benchmarks (Industry Comparison)

#### Merchant-Level: N/A
**Note:** Benchmarks are industry-based, not merchant-specific

#### Portfolio-Level: `GET /api/v1/risk/benchmarks?mcc={code}&naics={code}&sic={code}`
**Purpose:** Get industry benchmarks for risk comparison
**Returns:**
```json
{
  "industry_code": "5734",
  "industry_type": "mcc",
  "benchmarks": {
    "average_score": 70.0,
    "median_score": 72.0,
    "percentile_75": 80.0,
    "percentile_90": 85.0
  },
  "last_updated": "2025-01-27T10:00:00Z",
  "is_fallback": false
}
```
**Frontend:** ❌ Not implemented
**Use Case:** 
- Compare merchant risk score vs industry average
- Show merchant percentile ranking
- Display benchmark comparison in RiskAssessmentTab

**How to Use:**
1. Get merchant's industry code (MCC, NAICS, or SIC) from merchant data
2. Call benchmarks endpoint with industry code
3. Compare merchant's risk score against benchmarks
4. Display comparison in UI

---

### 4. Risk Score Comparison

#### Merchant-Level: `GET /api/v1/merchants/{id}/risk-score`
**Purpose:** Get merchant-specific risk score
**Returns:**
```json
{
  "merchant_id": "merchant_123",
  "risk_score": 0.5,
  "risk_level": "medium",
  "factors": [],
  "timestamp": "2025-01-27T10:00:00Z"
}
```
**Frontend:** ❌ Not implemented

#### Portfolio-Level: Available in `GET /api/v1/merchants/statistics`
**Portfolio Average Risk Score:** `risk_assessment_stats.average_risk_score` (0.25 in example)
**Portfolio Distribution:** `risk_assessment_stats.low_risk`, `medium_risk`, `high_risk`

**Comparison Capability:** ⚠️ Manual comparison needed
**Frontend:** ❌ Not implemented
**Use Case:**
- Show merchant risk score vs portfolio average
- Show merchant's position in risk distribution
- Display risk percentile

---

### 5. Portfolio Filtering and Grouping

#### Portfolio Types: `GET /api/v1/merchants/portfolio-types`
**Purpose:** Get list of available portfolio types
**Returns:**
```json
{
  "portfolio_types": ["retail", "ecommerce", "services", "manufacturing"],
  "timestamp": "2025-01-27T10:00:00Z"
}
```
**Frontend:** ❌ Not implemented
**Use Case:** Filter merchants by portfolio type for comparison

#### Risk Levels: `GET /api/v1/merchants/risk-levels`
**Purpose:** Get list of available risk levels
**Returns:**
```json
{
  "risk_levels": ["low", "medium", "high"],
  "timestamp": "2025-01-27T10:00:00Z"
}
```
**Frontend:** ❌ Not implemented
**Use Case:** Filter merchants by risk level for comparison

#### Merchant List with Filters: `GET /api/v1/merchants?portfolio_type={type}&risk_level={level}`
**Purpose:** Get filtered list of merchants
**Frontend:** ✅ Partially implemented (used in merchant list page)
**Use Case:** Compare merchant against similar merchants (same portfolio type, risk level)

---

## Comparison Scenarios

### Scenario 1: Merchant Risk Score vs Portfolio Average
**Current State:** ❌ Not implemented
**Required Endpoints:**
1. `GET /api/v1/merchants/{id}/risk-score` - Merchant risk score
2. `GET /api/v1/merchants/statistics` - Portfolio average risk score

**Implementation:**
- Fetch both endpoints
- Calculate difference: `merchant_score - portfolio_average`
- Display comparison card showing:
  - Merchant score
  - Portfolio average
  - Difference (positive/negative)
  - Percentile position

---

### Scenario 2: Merchant Risk Score vs Industry Benchmarks
**Current State:** ❌ Not implemented
**Required Endpoints:**
1. `GET /api/v1/merchants/{id}` - Get merchant industry code
2. `GET /api/v1/merchants/{id}/risk-score` - Get merchant risk score
3. `GET /api/v1/risk/benchmarks?mcc={code}` - Get industry benchmarks

**Implementation:**
- Extract industry code (MCC, NAICS, or SIC) from merchant data
- Fetch benchmarks for that industry
- Compare merchant score vs:
  - Industry average
  - Industry median
  - 75th percentile
  - 90th percentile
- Display benchmark comparison chart

---

### Scenario 3: Merchant Analytics vs Portfolio Analytics
**Current State:** ⚠️ Partially implemented (merchant analytics only)
**Required Endpoints:**
1. `GET /api/v1/merchants/{id}/analytics` - Merchant analytics
2. `GET /api/v1/merchants/analytics` - Portfolio analytics

**Implementation:**
- Fetch both endpoints
- Compare merchant metrics vs portfolio averages:
  - Classification confidence vs portfolio average
  - Security trust score vs portfolio average
  - Data quality vs portfolio average
- Display side-by-side comparison or overlay charts

---

### Scenario 4: Merchant Position in Portfolio Distribution
**Current State:** ❌ Not implemented
**Required Endpoints:**
1. `GET /api/v1/merchants/{id}/risk-score` - Merchant risk score
2. `GET /api/v1/merchants/statistics` - Portfolio risk distribution

**Implementation:**
- Get merchant risk level (low/medium/high)
- Get portfolio distribution:
  - `low_risk`: 850 merchants
  - `medium_risk`: 320 merchants
  - `high_risk`: 80 merchants
- Calculate merchant's percentile:
  - If low risk: `(850 - merchants_with_lower_risk) / 1250 * 100`
  - If medium risk: `(850 + merchants_with_medium_risk) / 1250 * 100`
  - If high risk: `(850 + 320 + merchants_with_high_risk) / 1250 * 100`
- Display percentile indicator

---

## Frontend Implementation Recommendations

### For Dashboard Pages (Portfolio-Level)

**Note:** These recommendations are for the dashboard pages, not the merchant details page.

1. **Business Intelligence Dashboard** (`/dashboard`)
   - ✅ Already displays portfolio analytics
   - ⚠️ Should use: `GET /api/v1/merchants/analytics`
   - ⚠️ Should use: `GET /api/v1/merchants/statistics`
   - **Status:** Needs review to verify endpoints are correctly integrated

2. **Risk Dashboard** (`/risk-dashboard`)
   - ✅ Already displays risk trends and distribution
   - ⚠️ Should use: `GET /api/v1/analytics/trends`
   - ⚠️ Should use: `GET /api/v1/analytics/insights`
   - ⚠️ Should use: `GET /api/v1/risk/metrics`
   - **Status:** Needs review to verify endpoints are correctly integrated

3. **Risk Indicators Dashboard** (`/risk-indicators`)
   - ✅ Already displays portfolio risk indicators
   - ⚠️ Should use: Aggregate risk indicators endpoint
   - **Status:** Needs review to verify endpoints are correctly integrated

### For Merchant Details Page (Merchant-Level with Comparisons)

**High Priority: Add Portfolio Comparison Features**

1. **Portfolio Statistics Comparison Card**
   - Add to `MerchantOverviewTab` or create new section
   - Fetch: `GET /api/v1/merchants/statistics` (for comparison context)
   - Show portfolio-wide statistics
   - Compare merchant vs portfolio averages
   - Display: "Merchant vs Portfolio Average" comparison

2. **Risk Benchmark Comparison**
   - Add to `RiskAssessmentTab`
   - Fetch: `GET /api/v1/risk/benchmarks?mcc={code}` (for comparison context)
   - Show merchant risk score vs industry benchmarks
   - Display percentile ranking
   - Endpoints:
     - `GET /api/v1/merchants/{id}/risk-score` (merchant data)
     - `GET /api/v1/risk/benchmarks?mcc={code}` (benchmark context)

3. **Portfolio Analytics Comparison**
   - Add to `BusinessAnalyticsTab`
   - Fetch: `GET /api/v1/merchants/analytics` (for comparison context)
   - Compare merchant analytics vs portfolio averages
   - Show side-by-side metrics
   - Endpoints:
     - `GET /api/v1/merchants/{id}/analytics` (already implemented - merchant data)
     - `GET /api/v1/merchants/analytics` (needs implementation - portfolio context)

### Medium Priority: Enhanced Comparison Features

4. **Risk Score Comparison Card**
   - Add to `MerchantOverviewTab`
   - Show merchant risk score vs portfolio average
   - Display risk level distribution
   - Endpoints:
     - `GET /api/v1/merchants/{id}/risk-score`
     - `GET /api/v1/merchants/statistics`

5. **Portfolio Context Indicators**
   - Add badges/indicators showing merchant's position
   - "Above Average", "Below Average", "Top 10%", etc.
   - Use portfolio statistics and benchmarks

---

## API Functions Needed in Frontend

### New API Functions to Add

```typescript
// Portfolio-level analytics
export async function getPortfolioAnalytics(): Promise<PortfolioAnalytics>

// Portfolio-level statistics
export async function getPortfolioStatistics(): Promise<PortfolioStatistics>

// Risk benchmarks (industry comparison)
export async function getRiskBenchmarks(params: {
  mcc?: string;
  naics?: string;
  sic?: string;
}): Promise<RiskBenchmarks>

// Merchant risk score
export async function getMerchantRiskScore(merchantId: string): Promise<MerchantRiskScore>

// Portfolio types (for filtering)
export async function getPortfolioTypes(): Promise<string[]>

// Risk levels (for filtering)
export async function getRiskLevels(): Promise<string[]>
```

---

## UI Components Needed

### 1. Portfolio Comparison Card
**Location:** `MerchantOverviewTab` or new section
**Features:**
- Portfolio statistics summary
- Merchant vs portfolio comparison
- Visual indicators (above/below average)

### 2. Risk Benchmark Comparison
**Location:** `RiskAssessmentTab`
**Features:**
- Industry benchmark chart
- Merchant score vs benchmarks
- Percentile indicator
- Comparison table

### 3. Analytics Comparison
**Location:** `BusinessAnalyticsTab`
**Features:**
- Side-by-side merchant vs portfolio metrics
- Comparison charts
- Difference indicators

### 4. Portfolio Context Badge
**Location:** Header or summary section
**Features:**
- Merchant's position in portfolio
- Risk percentile
- Industry ranking

---

## Data Flow for Comparison Features

### Example: Risk Score Comparison

```
1. User views merchant details page
   ↓
2. Frontend fetches merchant data
   - GET /api/v1/merchants/{id}
   - GET /api/v1/merchants/{id}/risk-score
   ↓
3. Frontend fetches portfolio/benchmark data
   - GET /api/v1/merchants/statistics (for portfolio average)
   - GET /api/v1/risk/benchmarks?mcc={code} (for industry benchmarks)
   ↓
4. Frontend calculates comparisons
   - Merchant score vs portfolio average
   - Merchant score vs industry benchmarks
   - Percentile calculations
   ↓
5. Frontend displays comparison UI
   - Comparison cards
   - Benchmark charts
   - Percentile indicators
```

---

## Summary

### Current State

#### Merchant Details Page (Merchant-Level)
- ✅ **Merchant-level features:** Well implemented (43% coverage - 15/35 features)
- ❌ **Portfolio comparison features:** Not implemented (0% coverage - 0/5 features)
- **Status:** Core merchant data is implemented, but portfolio comparisons are missing

#### Dashboard Pages (Portfolio-Level)
- ⚠️ **Portfolio-level features:** Status unknown (needs review)
- **Pages to Review:**
  - Business Intelligence Dashboard - [View](https://frontend-service-production-b225.up.railway.app/dashboard)
  - Risk Dashboard - [View](https://frontend-service-production-b225.up.railway.app/risk-dashboard)
  - Risk Indicators Dashboard - [View](https://frontend-service-production-b225.up.railway.app/risk-indicators)
- **Action Required:** Verify dashboard pages are correctly using portfolio-level endpoints

### Key Missing Features in Merchant Details Page

1. **Portfolio Statistics Integration** - Fetch portfolio stats for comparison
2. **Portfolio Analytics Integration** - Fetch portfolio analytics for comparison
3. **Risk Benchmarks Integration** - Fetch industry benchmarks for comparison
4. **Merchant Risk Score Integration** - Fetch merchant risk score for comparison
5. **Comparison UI Components** - Display merchant vs portfolio comparisons

### Recommended Implementation Order

#### Phase 1: Review Dashboard Pages
1. Verify Business Intelligence Dashboard uses `GET /api/v1/merchants/analytics` and `GET /api/v1/merchants/statistics`
2. Verify Risk Dashboard uses `GET /api/v1/analytics/trends`, `GET /api/v1/analytics/insights`, and `GET /api/v1/risk/metrics`
3. Verify Risk Indicators Dashboard uses aggregate risk indicators endpoints
4. Document any missing integrations

#### Phase 2: Add Comparison Features to Merchant Details Page
1. Add portfolio statistics fetching (for comparison context)
2. Add risk benchmarks fetching (for comparison context)
3. Add portfolio analytics fetching (for comparison context)
4. Create comparison UI components

#### Phase 3: Enhanced Comparison Features
1. Add percentile calculations
2. Add ranking indicators
3. Add visual comparison charts
4. Add "Above/Below Average" badges

### Architecture Clarity

**Portfolio-Level Endpoints:**
- **Primary Use:** Dashboard pages (Business Intelligence, Risk, Risk Indicators)
- **Secondary Use:** Merchant Details Page (for comparison context only)

**Merchant-Level Endpoints:**
- **Primary Use:** Merchant Details Page
- **Not Used By:** Dashboard pages (except for aggregate calculations)

---

## Architecture Summary Table

| Page Type | Page Name | URL | Primary Purpose | Endpoints Used | Status |
|-----------|-----------|-----|----------------|----------------|--------|
| **Portfolio-Level** | Business Intelligence Dashboard | `/dashboard` | Display portfolio-wide analytics and statistics | `GET /api/v1/merchants/analytics`<br>`GET /api/v1/merchants/statistics` | ⚠️ Needs review |
| **Portfolio-Level** | Risk Dashboard | `/risk-dashboard` | Display portfolio risk trends and distribution | `GET /api/v1/analytics/trends`<br>`GET /api/v1/analytics/insights`<br>`GET /api/v1/risk/metrics` | ⚠️ Needs review |
| **Portfolio-Level** | Risk Indicators Dashboard | `/risk-indicators` | Display portfolio-wide risk indicators | Aggregate risk indicators endpoints | ⚠️ Needs review |
| **Merchant-Level** | Merchant Details | `/merchant-details/[id]` | Display individual merchant data with portfolio comparisons | Merchant-specific endpoints (primary)<br>Portfolio endpoints (for comparison) | ✅ 43% implemented<br>❌ 0% comparison features |

### Endpoint Usage Matrix

| Endpoint | Business Intelligence Dashboard | Risk Dashboard | Risk Indicators Dashboard | Merchant Details Page |
|----------|--------------------------------|----------------|---------------------------|---------------------|
| `GET /api/v1/merchants/{id}` | ❌ | ❌ | ❌ | ✅ Primary |
| `GET /api/v1/merchants/{id}/analytics` | ❌ | ❌ | ❌ | ✅ Primary |
| `GET /api/v1/merchants/analytics` | ✅ Primary | ❌ | ❌ | ⚠️ For comparison |
| `GET /api/v1/merchants/statistics` | ✅ Primary | ❌ | ❌ | ⚠️ For comparison |
| `GET /api/v1/analytics/trends` | ❌ | ✅ Primary | ❌ | ⚠️ For comparison |
| `GET /api/v1/analytics/insights` | ❌ | ✅ Primary | ❌ | ⚠️ For comparison |
| `GET /api/v1/risk/metrics` | ❌ | ✅ Primary | ❌ | ⚠️ For comparison |
| `GET /api/v1/risk/benchmarks` | ❌ | ⚠️ Could use | ❌ | ⚠️ For comparison |
| `GET /api/v1/risk/indicators/{merchantId}` | ❌ | ❌ | ❌ | ✅ Primary |
| Aggregate risk indicators | ❌ | ❌ | ✅ Primary | ❌ |

**Legend:**
- ✅ Primary = Main data source for this page
- ⚠️ For comparison = Used for comparison context only
- ❌ = Not used by this page

---

**Document Version:** 1.1  
**Last Updated:** 2025-01-27  
**Next Review:** After implementing portfolio comparison features

