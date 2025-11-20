# Frontend Documentation

**Date:** 2025-01-27  
**Version:** 1.0.0  
**Purpose:** Comprehensive documentation of all frontend components, component props, usage examples, state management, data flow, caching strategy, and error handling patterns.

---

## Table of Contents

1. [Overview](#overview)
2. [New Components](#new-components)
3. [Component Props](#component-props)
4. [Usage Examples](#usage-examples)
5. [State Management](#state-management)
6. [Data Flow](#data-flow)
7. [Caching Strategy](#caching-strategy)
8. [Error Handling Patterns](#error-handling-patterns)
9. [API Client Functions](#api-client-functions)
10. [Type Definitions](#type-definitions)

---

## Overview

This documentation covers all new frontend components created for the merchant details page portfolio comparison features, risk assessment features, and dashboard enhancements.

**Component Categories:**
- **Portfolio Comparison Components:** Compare merchant data against portfolio averages
- **Risk Assessment Components:** Display risk scores, alerts, explainability, and recommendations
- **Analytics Components:** Display analytics comparisons
- **Enrichment Components:** Trigger and display data enrichment status

---

## New Components

### 1. PortfolioComparisonCard

**File:** `frontend/components/merchant/PortfolioComparisonCard.tsx`

**Purpose:** Displays a comparison of a specific merchant's statistics against the overall portfolio average.

**Props:**
```typescript
interface PortfolioComparisonCardProps {
  merchantId: string;
  merchantRiskLevel?: 'low' | 'medium' | 'high' | 'critical';
}
```

**Features:**
- Fetches portfolio statistics and merchant risk score
- Calculates percentile ranking
- Displays position indicators (Above Average, Below Average, etc.)
- Shows comparison metrics with visual indicators

**Usage:**
```tsx
<PortfolioComparisonCard 
  merchantId="merchant-123" 
  merchantRiskLevel="medium" 
/>
```

**Data Sources:**
- `getPortfolioStatistics()` - Portfolio statistics
- `getMerchantRiskScore(merchantId)` - Merchant risk score

---

### 2. RiskScoreCard

**File:** `frontend/components/merchant/RiskScoreCard.tsx`

**Purpose:** Displays a specific merchant's risk score and level.

**Props:**
```typescript
interface RiskScoreCardProps {
  merchantId: string;
}
```

**Features:**
- Fetches merchant risk score
- Displays risk score (0-100)
- Shows risk level with color-coded badge
- Displays confidence score
- Shows assessment date

**Usage:**
```tsx
<RiskScoreCard merchantId="merchant-123" />
```

**Data Sources:**
- `getMerchantRiskScore(merchantId)` - Merchant risk score

---

### 3. PortfolioContextBadge

**File:** `frontend/components/merchant/PortfolioContextBadge.tsx`

**Purpose:** Displays a badge indicating a merchant's risk position relative to the overall portfolio.

**Props:**
```typescript
interface PortfolioContextBadgeProps {
  merchantId: string;
  variant?: 'default' | 'compact' | 'detailed';
}
```

**Features:**
- Fetches portfolio statistics and merchant risk score
- Calculates percentile and position
- Displays position indicators (Top 10%, Above Average, etc.)
- Supports multiple variants (default, compact, detailed)

**Usage:**
```tsx
<PortfolioContextBadge 
  merchantId="merchant-123" 
  variant="detailed" 
/>
```

**Data Sources:**
- `getPortfolioStatistics()` - Portfolio statistics
- `getMerchantRiskScore(merchantId)` - Merchant risk score

---

### 4. RiskBenchmarkComparison

**File:** `frontend/components/merchant/RiskBenchmarkComparison.tsx`

**Purpose:** Displays a comparison of a specific merchant's risk score against industry benchmarks.

**Props:**
```typescript
interface RiskBenchmarkComparisonProps {
  merchantId: string;
}
```

**Features:**
- Extracts industry code from merchant analytics (MCC, NAICS, or SIC)
- Fetches industry benchmarks
- Calculates percentile and position
- Displays comparison chart
- Shows detailed benchmark statistics

**Usage:**
```tsx
<RiskBenchmarkComparison merchantId="merchant-123" />
```

**Data Sources:**
- `getMerchantAnalytics(merchantId)` - Merchant analytics (for industry codes)
- `getRiskBenchmarks({ mcc, naics, sic })` - Industry benchmarks
- `getMerchantRiskScore(merchantId)` - Merchant risk score

---

### 5. AnalyticsComparison

**File:** `frontend/components/merchant/AnalyticsComparison.tsx`

**Purpose:** Displays a comparison of a specific merchant's analytics metrics against the overall portfolio averages.

**Props:**
```typescript
interface AnalyticsComparisonProps {
  merchantId: string;
  merchantAnalytics?: AnalyticsData; // Optional - will fetch if not provided
}
```

**Features:**
- Accepts merchant analytics as prop or fetches if not provided
- Fetches portfolio analytics
- Compares classification confidence, security trust score, and data quality
- Displays comparison charts for each metric
- Shows difference percentages

**Usage:**
```tsx
<AnalyticsComparison 
  merchantId="merchant-123" 
  merchantAnalytics={analytics} 
/>
```

**Data Sources:**
- `getMerchantAnalytics(merchantId)` - Merchant analytics (if not provided as prop)
- `getPortfolioAnalytics()` - Portfolio analytics

---

### 6. RiskAlertsSection

**File:** `frontend/components/merchant/RiskAlertsSection.tsx`

**Purpose:** Displays active risk alerts for a merchant, grouped by severity.

**Props:**
```typescript
interface RiskAlertsSectionProps {
  merchantId: string;
  autoRefresh?: boolean; // Default: true
  refreshInterval?: number; // Default: 30000 (30 seconds)
}
```

**Features:**
- Fetches active risk alerts
- Groups alerts by severity (critical, high, medium, low)
- Displays alerts in collapsible sections
- Shows toast notifications for critical and high severity alerts
- Auto-refreshes alerts periodically
- Includes loading and error states

**Usage:**
```tsx
<RiskAlertsSection 
  merchantId="merchant-123" 
  autoRefresh={true}
  refreshInterval={30000}
/>
```

**Data Sources:**
- `getRiskAlerts(merchantId, severity?, status?)` - Risk alerts

---

### 7. RiskExplainabilitySection

**File:** `frontend/components/merchant/RiskExplainabilitySection.tsx`

**Purpose:** Displays the explainability of a merchant's risk assessment, including SHAP values and feature importance.

**Props:**
```typescript
interface RiskExplainabilitySectionProps {
  merchantId: string;
}
```

**Features:**
- Fetches current risk assessment to get assessment ID
- Fetches risk explanation (SHAP values, feature importance)
- Displays SHAP values chart (top 10 features)
- Displays feature importance chart
- Shows risk factors table with scores, weights, and impacts
- Includes loading and error states with retry

**Usage:**
```tsx
<RiskExplainabilitySection merchantId="merchant-123" />
```

**Data Sources:**
- `getRiskAssessment(merchantId)` - Current risk assessment (for assessment ID)
- `explainRiskAssessment(assessmentId)` - Risk explanation

---

### 8. RiskRecommendationsSection

**File:** `frontend/components/merchant/RiskRecommendationsSection.tsx`

**Purpose:** Displays actionable risk recommendations for a merchant, grouped by priority.

**Props:**
```typescript
interface RiskRecommendationsSectionProps {
  merchantId: string;
}
```

**Features:**
- Fetches risk recommendations
- Groups recommendations by priority (high, medium, low)
- Displays recommendations in collapsible sections
- Shows action items for each recommendation
- Includes loading and error states

**Usage:**
```tsx
<RiskRecommendationsSection merchantId="merchant-123" />
```

**Data Sources:**
- `getRiskRecommendations(merchantId)` - Risk recommendations

---

### 9. EnrichmentButton

**File:** `frontend/components/merchant/EnrichmentButton.tsx`

**Purpose:** Provides a UI for triggering data enrichment for a merchant.

**Props:**
```typescript
interface EnrichmentButtonProps {
  merchantId: string;
}
```

**Features:**
- Fetches available enrichment sources
- Displays sources in a dialog
- Allows enabling/disabling sources
- Triggers enrichment job
- Shows job status (pending, processing, completed, failed)
- Displays badge with number of enabled sources
- Includes error handling and loading states

**Usage:**
```tsx
<EnrichmentButton merchantId="merchant-123" />
```

**Data Sources:**
- `getEnrichmentSources(merchantId)` - Available enrichment sources
- `triggerEnrichment(merchantId, sources)` - Trigger enrichment

---

## Component Props

### Common Props

All components share common patterns:

**Loading State:**
- Components show loading skeletons while fetching data
- Loading states are managed internally with `useState`

**Error Handling:**
- Components display error messages when API calls fail
- Error states include retry functionality
- Errors are logged to console for debugging

**Data Fetching:**
- Components fetch data on mount using `useEffect`
- Data is cached using `APICache` (5-7 minute TTL)
- Requests are deduplicated using `RequestDeduplicator`

---

## Usage Examples

### Merchant Overview Tab

```tsx
import { PortfolioComparisonCard } from '@/components/merchant/PortfolioComparisonCard';
import { RiskScoreCard } from '@/components/merchant/RiskScoreCard';
import { PortfolioContextBadge } from '@/components/merchant/PortfolioContextBadge';

export function MerchantOverviewTab({ merchantId }: { merchantId: string }) {
  return (
    <div className="space-y-6">
      <PortfolioContextBadge merchantId={merchantId} variant="detailed" />
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <RiskScoreCard merchantId={merchantId} />
        <PortfolioComparisonCard 
          merchantId={merchantId} 
          merchantRiskLevel={merchant.riskLevel}
        />
      </div>
    </div>
  );
}
```

### Risk Assessment Tab

```tsx
import { RiskBenchmarkComparison } from '@/components/merchant/RiskBenchmarkComparison';
import { RiskExplainabilitySection } from '@/components/merchant/RiskExplainabilitySection';
import { RiskRecommendationsSection } from '@/components/merchant/RiskRecommendationsSection';

export function RiskAssessmentTab({ merchantId }: { merchantId: string }) {
  return (
    <div className="space-y-6">
      <RiskBenchmarkComparison merchantId={merchantId} />
      <RiskExplainabilitySection merchantId={merchantId} />
      <RiskRecommendationsSection merchantId={merchantId} />
    </div>
  );
}
```

### Business Analytics Tab

```tsx
import { AnalyticsComparison } from '@/components/merchant/AnalyticsComparison';

export function BusinessAnalyticsTab({ merchantId, analytics }: Props) {
  return (
    <div className="space-y-6">
      <AnalyticsComparison 
        merchantId={merchantId} 
        merchantAnalytics={analytics}
      />
    </div>
  );
}
```

### Risk Indicators Tab

```tsx
import { RiskAlertsSection } from '@/components/merchant/RiskAlertsSection';

export function RiskIndicatorsTab({ merchantId }: { merchantId: string }) {
  return (
    <div className="space-y-6">
      <RiskAlertsSection 
        merchantId={merchantId}
        autoRefresh={true}
        refreshInterval={30000}
      />
    </div>
  );
}
```

---

## State Management

### State Management Pattern

The frontend uses **React's built-in state management** with `useState` and `useEffect`:

**Pattern:**
```typescript
const [data, setData] = useState<DataType | null>(null);
const [loading, setLoading] = useState(true);
const [error, setError] = useState<string | null>(null);

useEffect(() => {
  async function fetchData() {
    try {
      setLoading(true);
      setError(null);
      const result = await apiFunction();
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  }
  
  fetchData();
}, [dependencies]);
```

**Key Points:**
- ✅ **NOT using React Query** - Codebase uses `useState` + `useEffect` pattern
- ✅ **Caching handled by `APICache`** - In-memory cache with TTL
- ✅ **Request deduplication handled by `RequestDeduplicator`** - Prevents duplicate concurrent requests
- ✅ **State is component-local** - No global state management library

---

## Data Flow

### Data Flow Pattern

**1. Component Mount:**
```
Component mounts → useEffect triggers → API function called
```

**2. API Function:**
```
API function → Check cache → If cached, return cached data
           → If not cached, check deduplicator
           → If duplicate request in progress, wait for it
           → If new request, make fetch call
           → Cache response → Return data
```

**3. Component Update:**
```
API function returns → setData() → Component re-renders → Display data
```

### Example Data Flow

**PortfolioComparisonCard:**
```
1. Component mounts
2. useEffect triggers
3. Parallel fetch:
   - getPortfolioStatistics()
   - getMerchantRiskScore(merchantId)
4. Both functions check cache
5. If not cached, make fetch requests
6. Responses cached (5-7 min TTL)
7. Calculate comparison metrics
8. setData() → Component displays comparison
```

---

## Caching Strategy

### APICache Class

**File:** `frontend/lib/api-cache.ts`

**Purpose:** In-memory cache for API responses with TTL (time-to-live).

**Features:**
- In-memory storage using `Map`
- TTL-based expiration
- Automatic cleanup of expired entries
- Thread-safe (single-threaded JavaScript)

**Usage:**
```typescript
const apiCache = new APICache(5 * 60 * 1000); // 5 minutes TTL

// Get from cache
const cached = apiCache.get<DataType>('cache-key');
if (cached) {
  return cached;
}

// Set in cache
apiCache.set('cache-key', data, 7 * 60 * 1000); // 7 minutes TTL
```

### Cache TTL by Endpoint

| Endpoint | Cache TTL | Reason |
|----------|-----------|--------|
| Portfolio Analytics | 7 minutes | Changes infrequently, high read volume |
| Portfolio Statistics | 5 minutes | Changes infrequently, high read volume |
| Risk Trends | 5 minutes | Time-series data, updates periodically |
| Risk Insights | 5 minutes | Insights change periodically |
| Risk Benchmarks | 10 minutes | Industry benchmarks change rarely |
| Merchant Risk Score | 2 minutes | Risk scores update more frequently |
| Merchant Analytics | 5 minutes | Analytics update periodically |
| Risk Alerts | 30 seconds | Alerts need near real-time updates |

### Cache Key Format

Cache keys follow this pattern:
- `portfolio-analytics`
- `portfolio-statistics`
- `risk-trends-{timeframe}-{industry}-{country}`
- `risk-insights-{industry}-{country}-{risk_level}`
- `risk-benchmarks-{mcc}-{naics}-{sic}`
- `merchant-risk-score-{merchantId}`
- `merchant-analytics-{merchantId}`

---

## Error Handling Patterns

### Error Handling Strategy

**1. API Function Level:**
```typescript
try {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`API Error ${response.status}`);
  }
  const data = await response.json();
  return data;
} catch (error) {
  // Log error
  console.error('API call failed:', error);
  // Re-throw for component to handle
  throw error;
}
```

**2. Component Level:**
```typescript
useEffect(() => {
  async function fetchData() {
    try {
      setError(null);
      const data = await apiFunction();
      setData(data);
    } catch (err) {
      const errorMessage = err instanceof Error 
        ? err.message 
        : 'Failed to load data';
      setError(errorMessage);
      toast.error('Failed to load data', {
        description: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  }
  
  fetchData();
}, []);
```

**3. Error Display:**
```tsx
{error && (
  <div className="text-destructive">
    Error: {error}
    <Button onClick={retry}>Retry</Button>
  </div>
)}
```

### Error Types

**Network Errors:**
- Connection refused
- Timeout
- Network unavailable

**API Errors:**
- 400 Bad Request
- 401 Unauthorized
- 404 Not Found
- 500 Internal Server Error
- 502 Bad Gateway
- 503 Service Unavailable

**Data Errors:**
- Invalid response format
- Missing required fields
- Type mismatches

---

## API Client Functions

### New API Functions

**File:** `frontend/lib/api.ts`

#### Portfolio-Level Functions

**getPortfolioAnalytics():**
```typescript
export async function getPortfolioAnalytics(): Promise<PortfolioAnalytics>
```
- **Cache Key:** `portfolio-analytics`
- **Cache TTL:** 7 minutes
- **Endpoint:** `GET /api/v1/merchants/analytics`

**getPortfolioStatistics():**
```typescript
export async function getPortfolioStatistics(): Promise<PortfolioStatistics>
```
- **Cache Key:** `portfolio-statistics`
- **Cache TTL:** 5 minutes
- **Endpoint:** `GET /api/v1/merchants/statistics`

**getRiskTrends(params?):**
```typescript
export async function getRiskTrends(params?: {
  timeframe?: string;
  industry?: string;
  country?: string;
  limit?: number;
}): Promise<RiskTrends>
```
- **Cache Key:** `risk-trends-{timeframe}-{industry}-{country}`
- **Cache TTL:** 5 minutes
- **Endpoint:** `GET /api/v1/analytics/trends`

**getRiskInsights(params?):**
```typescript
export async function getRiskInsights(params?: {
  industry?: string;
  country?: string;
  risk_level?: string;
}): Promise<RiskInsights>
```
- **Cache Key:** `risk-insights-{industry}-{country}-{risk_level}`
- **Cache TTL:** 5 minutes
- **Endpoint:** `GET /api/v1/analytics/insights`

#### Comparison Functions

**getRiskBenchmarks(params):**
```typescript
export async function getRiskBenchmarks(params: {
  mcc?: string;
  naics?: string;
  sic?: string;
}): Promise<RiskBenchmarks>
```
- **Cache Key:** `risk-benchmarks-{mcc}-{naics}-{sic}`
- **Cache TTL:** 10 minutes
- **Endpoint:** `GET /api/v1/risk/benchmarks`

**getMerchantRiskScore(merchantId):**
```typescript
export async function getMerchantRiskScore(
  merchantId: string
): Promise<MerchantRiskScore>
```
- **Cache Key:** `merchant-risk-score-{merchantId}`
- **Cache TTL:** 2 minutes
- **Endpoint:** `GET /api/v1/merchants/{id}/risk-score`

#### Risk Assessment Functions

**getRiskAlerts(merchantId, severity?, status?):**
```typescript
export async function getRiskAlerts(
  merchantId: string,
  severity?: string,
  status?: string
): Promise<RiskIndicatorsData>
```
- **Cache Key:** `risk-alerts-{merchantId}-{severity}-{status}`
- **Cache TTL:** 30 seconds (short TTL for near real-time)
- **Endpoint:** `GET /api/v1/risk/indicators/{id}?status=active`

**explainRiskAssessment(assessmentId):**
```typescript
export async function explainRiskAssessment(
  assessmentId: string
): Promise<RiskExplanationResponse>
```
- **Cache Key:** `risk-explanation-{assessmentId}`
- **Cache TTL:** 5 minutes
- **Endpoint:** `GET /api/v1/risk/explain/{assessmentId}`

**getRiskRecommendations(merchantId):**
```typescript
export async function getRiskRecommendations(
  merchantId: string
): Promise<RiskRecommendationsResponse>
```
- **Cache Key:** `risk-recommendations-{merchantId}`
- **Cache TTL:** 5 minutes
- **Endpoint:** `GET /api/v1/merchants/{id}/risk-recommendations`

---

## Type Definitions

### Type Definitions File

**File:** `frontend/types/merchant.ts`

### Portfolio-Level Types

**PortfolioAnalytics:**
```typescript
interface PortfolioAnalytics {
  totalMerchants: number;
  averageRiskScore: number;
  averageClassificationConfidence: number;
  averageSecurityTrustScore: number;
  averageDataQuality: number;
  riskDistribution: {
    low: number;
    medium: number;
    high: number;
  };
  industryDistribution: Record<string, number>;
  countryDistribution: Record<string, number>;
  timestamp: string;
}
```

**PortfolioStatistics:**
```typescript
interface PortfolioStatistics {
  totalMerchants: number;
  totalAssessments: number;
  averageRiskScore: number;
  riskDistribution: {
    low: number;
    medium: number;
    high: number;
  };
  industryBreakdown: Array<{
    industry: string;
    count: number;
    averageRiskScore: number;
  }>;
  countryBreakdown: Array<{
    country: string;
    count: number;
    averageRiskScore: number;
  }>;
  timestamp: string;
}
```

### Risk Types

**RiskTrends:**
```typescript
interface RiskTrends {
  trends: RiskTrend[];
  summary: TrendSummary;
}

interface RiskTrend {
  industry: string;
  country: string;
  average_risk_score: number;
  trend_direction: 'improving' | 'worsening' | 'stable';
  change_percentage: number;
  sample_size: number;
}
```

**RiskInsights:**
```typescript
interface RiskInsights {
  insights: RiskInsight[];
  recommendations: Recommendation[];
}
```

**RiskBenchmarks:**
```typescript
interface RiskBenchmarks {
  industry_code: string;
  industry_type: 'mcc' | 'naics' | 'sic';
  average_risk_score: number;
  median_risk_score: number;
  percentile_25: number;
  percentile_75: number;
  percentile_90: number;
  sample_size: number;
}
```

**MerchantRiskScore:**
```typescript
interface MerchantRiskScore {
  merchant_id: string;
  risk_score: number;
  risk_level: 'low' | 'medium' | 'high';
  confidence_score: number;
  assessment_date: string;
  factors: Array<{
    category: string;
    score: number;
    weight: number;
  }>;
}
```

### Comparison Types

**PortfolioComparison:**
```typescript
interface PortfolioComparison {
  merchantScore: number;
  portfolioAverage: number;
  portfolioMedian: number;
  percentile: number;
  position: 'above_average' | 'below_average' | 'average';
  difference: number;
  differencePercentage: number;
}
```

**BenchmarkComparison:**
```typescript
interface BenchmarkComparison {
  merchantScore: number;
  industryAverage: number;
  industryMedian: number;
  industryPercentile75: number;
  industryPercentile90: number;
  percentile: number;
  position: 'top_10' | 'top_25' | 'average' | 'bottom_25' | 'bottom_10';
  difference: number;
  differencePercentage: number;
}
```

**AnalyticsComparison:**
```typescript
interface AnalyticsComparison {
  merchant: {
    classificationConfidence: number;
    securityTrustScore: number;
    dataQuality: number;
  };
  portfolio: {
    averageClassificationConfidence: number;
    averageSecurityTrustScore: number;
    averageDataQuality: number;
  };
  differences: {
    classificationConfidence: number;
    securityTrustScore: number;
    dataQuality: number;
  };
  percentages: {
    classificationConfidence: number;
    securityTrustScore: number;
    dataQuality: number;
  };
}
```

---

## Best Practices

### Component Development

1. **Always use TypeScript interfaces for props**
2. **Handle loading and error states**
3. **Use APICache for data caching**
4. **Use RequestDeduplicator for request deduplication**
5. **Display user-friendly error messages**
6. **Include retry functionality for failed requests**

### Data Fetching

1. **Fetch data in parallel when possible**
2. **Use Promise.allSettled for independent requests**
3. **Cache responses with appropriate TTL**
4. **Deduplicate concurrent requests**
5. **Handle errors gracefully**

### State Management

1. **Use useState for component-local state**
2. **Use useEffect for data fetching**
3. **Clean up effects to prevent memory leaks**
4. **Don't use global state for component data**

---

## Conclusion

This documentation covers all new frontend components, their props, usage examples, state management patterns, data flow, caching strategy, and error handling.

**Last Updated:** 2025-01-27  
**Version:** 1.0.0  
**Status:** ✅ Complete

