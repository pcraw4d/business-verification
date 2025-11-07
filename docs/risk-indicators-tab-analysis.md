# Risk Indicators Tab - Card Analysis

## Overview

The Risk Indicators tab within the merchant details page provides a comprehensive view of risk assessment across six categories: Financial, Operational, Regulatory, Reputational, Cybersecurity, and Content. This document analyzes each card component, its data sources, and the value it provides to users.

---

## Card 1: Enhanced Risk Level Badges

### Purpose & Content
This card displays the top 4 risk categories as visual badges, providing an at-a-glance overview of the merchant's highest risk areas. Each badge shows:
- **Risk Level**: CRITICAL, HIGH, MEDIUM, or LOW (with color coding)
- **Score**: Numerical risk score (0-100)
- **Trend Indicator**: IMPROVING (↓), STABLE (—), or RISING (↑)
- **Category Icon**: Visual icon representing the risk category

### Source of Information
**Frontend Component**: `RiskIndicatorsUITemplate.getRiskBadgesHTML()`
- **Location**: `web/js/components/risk-indicators-ui-template.js` (lines 16-56)
- **Data Source**: 
  - Primary: `riskData.categories` object from `RiskIndicatorsDataService.loadAllRiskData()`
  - Categories are sorted by score (highest first) and top 4 are displayed
  - Trend calculation compares current score with `previousScore` from historical data
  - Risk level classification uses thresholds:
    - Low: 0-25
    - Medium: 26-50
    - High: 51-75
    - Critical: 76-100

**Backend Data Flow**:
1. `RiskIndicatorsDataService.loadRiskAssessment()` calls `/api/v1/risk/assess` endpoint
2. Backend handler `RiskHandler.AssessRiskHandler()` processes the request
3. Risk service calculates category scores via `RiskScorer.CalculateRiskScore()`
4. Category scores are calculated per category (financial, operational, regulatory, etc.)
5. Data is normalized and merged in `mergeAndNormalize()` method

### Value to Users
- **Quick Risk Identification**: Users can immediately see which risk categories require attention
- **Visual Hierarchy**: Color-coded badges help prioritize risk mitigation efforts
- **Trend Awareness**: Trend indicators show whether risks are improving or worsening over time
- **Benchmark Reference**: The badge display serves as a visual legend for understanding risk levels across the dashboard

---

## Card 2: Risk Heat Map (Legend Section)

### Purpose & Content
This card serves dual purposes:
1. **Risk Intensity Visualization**: Displays a heat map grid showing risk intensity across subcategories within each of the 6 main categories
2. **Legend Reference**: Provides a comprehensive legend explaining:
   - **Risk Level Ranges**: 
     - Low Risk (0-25) - Light Green
     - Medium Risk (26-50) - Yellow
     - High Risk (51-75) - Pink
     - Critical Risk (76-100) - Red
   - **Risk Trend Indicators**: 
     - IMPROVING (↓) - Risk decreasing
     - STABLE (—) - Risk unchanged
     - RISING (↑) - Risk increasing

### Source of Information
**Frontend Component**: `RiskIndicatorsUITemplate.getHeatMapHTML()`
- **Location**: `web/js/components/risk-indicators-ui-template.js` (lines 63-149)
- **Data Source**:
  - Heat map cells generated from `categories[category].subCategories` object
  - Each category has 5 subcategories (e.g., Financial: revenue, cashFlow, debt, credit, market)
  - Subcategory scores are converted to risk levels via `getRiskLevelFromScore()`
  - Legend definitions are static but applied dynamically to actual data

**Backend Data Flow**:
1. Subcategory scores come from `buildRiskCategories()` in `RiskIndicatorsDataService`
2. Default subcategory structure defined in `buildRiskCategories()` (lines 187-256)
3. Real risk assessment data overrides defaults when available
4. Subcategory risk calculations use backend scoring algorithms:
   - `calculateSubcategoryRiskScore()` in `risk_models.go`
   - Category-specific adjustments via `adjustFinancialRiskScore()`, `adjustOperationalRiskScore()`, etc.

### Value to Users
- **Comprehensive Risk Mapping**: Visual representation of all risk subcategories in one view
- **Legend Interpretation**: Provides the key to understanding all color-coding and symbols used throughout the dashboard
- **Drill-Down Capability**: Heat map cells are interactive (hover tooltips) allowing users to explore specific subcategory risks
- **Consistency Reference**: Ensures users understand the same risk level definitions across all cards

---

## Card 3: Enhanced Risk Progress Indicators

### Purpose & Content
This card provides detailed, category-specific risk breakdowns with:
- **Category Name & Icon**: Visual identification of each risk category
- **Score Display**: Current score out of 100 (e.g., "15/100")
- **Risk Level Badge**: Color-coded badge showing LOW, MEDIUM, HIGH, or CRITICAL
- **Progress Bar**: Visual progress bar with animated fill showing risk level
- **Description**: Human-readable description of the risk status (e.g., "Excellent financial health", "Compliance issues detected")
- **Last Updated Timestamp**: ISO timestamp showing when the data was last refreshed

### Source of Information
**Frontend Component**: `RiskIndicatorsUITemplate.getProgressBarsHTML()`
- **Location**: `web/js/components/risk-indicators-ui-template.js` (lines 156-200)
- **Data Source**:
  - Category data from `riskData.categories` object
  - Descriptions generated via `getRiskDescription(category, level)` helper method
  - Last updated timestamp from `category.lastUpdated` (ISO format)
  - Progress bar width calculated as percentage: `(score / 100) * 100%`

**Backend Data Flow**:
1. Category scores calculated by `RiskScorer.CalculateRiskScore()`
2. Each category has dedicated calculation methods:
   - `calculateFinancialCategoryScore()` for financial risk
   - `calculateSecurityCategoryScore()` for cybersecurity risk
   - `calculateComplianceCategoryScore()` for regulatory risk
   - `calculateReputationCategoryScore()` for reputational risk
   - `calculateDomainCategoryScore()` for operational/content risk
3. Scores normalized to 0-100 scale in frontend
4. Risk level determined by `determineRiskLevel()` method using thresholds

**Description Mapping**:
Descriptions are mapped based on category and level:
- Financial: "Excellent financial health" (low), "Some financial concerns" (medium), "Significant financial risks" (high), "Critical financial issues" (critical)
- Operational: "Smooth operations" (low), "Some operational challenges" (medium), "Operational difficulties" (high), "Critical operational issues" (critical)
- Regulatory: "Good compliance standing" (low), "Minor compliance issues" (medium), "Compliance issues detected" (high), "Critical compliance violations" (critical)
- Reputational: "Strong reputation" (low), "Mixed reputation signals" (medium), "Reputation concerns" (high), "Severe reputation damage" (critical)
- Cybersecurity: "Strong security posture" (low), "Some security concerns" (medium), "Security vulnerabilities found" (high), "Critical security breaches" (critical)
- Content: "Clean content profile" (low), "Some content concerns" (medium), "Content risks detected" (high), "High-risk content found" (critical)

### Value to Users
- **Detailed Risk Status**: Provides specific, actionable information about each risk category
- **Contextual Understanding**: Human-readable descriptions help users understand what the scores mean in practical terms
- **Visual Progress Tracking**: Progress bars provide immediate visual feedback on risk levels
- **Data Freshness**: Timestamps ensure users know how current the information is
- **Category Prioritization**: Easy comparison across categories to identify which areas need immediate attention

---

## Card 4: Risk Radar Chart

### Purpose & Content
This card displays a hexagonal radar/spider chart comparing:
- **Current Risk Profile**: Solid blue line showing the merchant's risk scores across all 6 categories
- **Industry Average**: Dashed grey line showing benchmark data for comparison
- **Scale**: 0-10 scale (scores normalized from 0-100 scale)
- **Categories**: All 6 risk categories displayed on the chart axes

### Source of Information
**Frontend Component**: `MerchantRiskIndicatorsTab.initializeRadarChart()`
- **Location**: `web/js/components/merchant-risk-indicators-tab.js` (lines 222-237)
- **Data Preparation**: `prepareRadarData()` method (lines 421-449)
- **Chart Rendering**: Uses `RiskVisualization.createRiskRadarChart()` component (D3.js based)

**Data Source**:
- **Current Risk Data**: 
  - Category scores from `riskData.categories[category].score`
  - Scores normalized from 0-100 to 0-10 scale: `(score / 100) * 10`
  - Category order: ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content']
- **Industry Average**:
  - Currently uses mock data: `[25, 35, 45, 30, 40, 20]` (lines 439)
  - **TODO**: Should be replaced with real industry benchmark data from backend
  - Future implementation should query industry-specific averages from database

**Backend Data Flow**:
1. Category scores come from risk assessment API response
2. Radar chart data prepared in frontend from normalized category scores
3. Industry averages would ideally come from:
   - Aggregated historical risk assessments by industry code (MCC/NAICS/SIC)
   - Stored in database and retrieved via dedicated endpoint
   - Currently not implemented (using mock data)

### Value to Users
- **Visual Comparison**: Immediate visual understanding of how merchant compares to industry peers
- **Risk Profile Overview**: Single chart shows complete risk landscape across all categories
- **Benchmark Context**: Industry average provides context for whether merchant's risks are typical or exceptional
- **Pattern Recognition**: Visual shape of radar chart helps identify risk concentration areas
- **Quick Assessment**: Faster than reading individual scores to understand overall risk posture

---

## Card 5: Risk Category Analysis

### Purpose & Content
This card provides a tabular summary of each risk category with:
- **Category Name**: Formatted category name (e.g., "Financial", "Operational")
- **Score**: Numerical score out of 100
- **Risk Level**: Text badge showing LOW, MEDIUM, HIGH, or CRITICAL
- **Stability Status**: Trend indicator (IMPROVING, STABLE, RISING)

### Source of Information
**Frontend Component**: `MerchantRiskIndicatorsTab.initializeRiskCategoryAnalysis()`
- **Location**: `web/js/components/merchant-risk-indicators-tab.js` (lines 271-312)
- **Data Source**:
  - Same `riskData.categories` object used by other cards
  - Trend calculation via `helpers.calculateTrendDirection(current, previous)`
  - Risk level badges use same classification as other cards

**Backend Data Flow**:
- Same data pipeline as other cards (risk assessment API → data service → normalization)
- Trend data requires historical comparison:
  - Current scores from latest risk assessment
  - Previous scores from historical risk assessments stored in database
  - Trend calculated by comparing current vs. previous (threshold: ±5 points for trend change)

### Value to Users
- **Quick Reference Table**: Easy-to-scan tabular format for comparing all categories
- **Definitive Status**: Provides clear, scorable status for each risk type
- **Trend Tracking**: Shows whether each category is improving, stable, or worsening
- **Reinforcement**: Reinforces data presented in other visual formats (badges, progress bars, radar chart)
- **Print-Friendly**: Tabular format is easier to reference in reports or documentation

---

## Card 6: Risk Summary

### Purpose & Content
This card provides a single, aggregated risk score and summary text including:
- **Overall Risk Score**: Single number out of 100 (e.g., "41/100")
- **Risk Level**: Overall classification (low, medium, high, critical)
- **Summary Text**: Narrative description that includes:
  - Overall risk score and level
  - Count of active alerts (if any)
  - Count of available recommendations (if any)

### Source of Information
**Frontend Component**: `MerchantRiskIndicatorsTab.updateRiskSummary()`
- **Location**: `web/js/components/merchant-risk-indicators-tab.js` (lines 317-337)
- **Data Source**:
  - Overall score: `riskData.overallRiskScore`
  - Risk level: `riskData.riskLevel`
  - Alert count: `riskData.alerts.length`
  - Recommendation count: `riskData.recommendations.length`

**Backend Data Flow**:
1. Overall score calculated in `RiskIndicatorsHelpers.calculateOverallRiskScore(categories)`
2. Calculation method:
   - Weighted average of all 6 category scores
   - Each category may have different weights (currently equal weighting)
   - Formula: `sum(category_score * weight) / sum(weights)`
3. Risk level determined by `getOverallRiskLevel(overallScore)` using same thresholds as individual categories
4. Alerts extracted from `extractAlerts(analytics, risk)` method
5. Recommendations extracted from `extractRecommendations(risk)` method

### Value to Users
- **Executive Summary**: Provides immediate, high-level understanding of merchant's total risk posture
- **Quick Assessment**: Single number allows for rapid prioritization and decision-making
- **Actionable Context**: Summary text highlights if there are alerts or recommendations requiring attention
- **Reporting**: Useful for reports, dashboards, or executive briefings where a single risk metric is needed
- **Comparison Tool**: Overall score enables easy comparison between multiple merchants

---

## Card 7: Recommended Actions

### Purpose & Content
This card provides actionable recommendations based on the risk analysis, including:
- **Recommendation Title**: Clear action item name
- **Priority Badge**: CRITICAL, HIGH, MEDIUM, or LOW priority
- **Description**: Detailed explanation of the recommendation
- **Metadata**: 
  - Impact Score: Expected impact of implementing the recommendation (0-1 scale)
  - Difficulty: Implementation difficulty (Low, Medium, High)
  - Type: Recommendation type (ml_based, manual_verification, document_verification, compliance_check, security_audit)
- **Action Buttons**: 
  - Dismiss: Remove recommendation from view
  - Implement: Start implementation workflow

### Source of Information
**Frontend Component**: `RiskIndicatorsUITemplate.getRecommendationsHTML()`
- **Location**: `web/js/components/risk-indicators-ui-template.js` (lines 298-362)
- **Data Source**: `riskData.recommendations` array from data service

**Backend Data Flow**:
1. Recommendations extracted in `RiskIndicatorsDataService.extractRecommendations(risk)`
2. **Location**: `web/js/components/risk-indicators-data-service.js` (lines 375-428)
3. **Generation Methods**:
   - **ML-Based Recommendations**: Generated from risk assessment API response if `includeRecommendations: true`
   - **Manual Verification Recommendations**: Generated based on high-risk factors:
     - Financial risk → "Financial Document Verification"
     - Regulatory risk → "Regulatory Compliance Verification"
     - Cybersecurity risk → "Security Assessment Required"
4. **Recommendation Structure**:
   ```javascript
   {
     id: string,
     type: string,
     priority: 'critical' | 'high' | 'medium' | 'low',
     title: string,
     description: string,
     impactScore: number (0-1),
     difficulty: string,
     actionRequired: string,
     status: 'pending' | 'in_progress' | 'completed'
   }
   ```

**Backend API**:
- Risk assessment endpoint accepts `includeRecommendations: true` parameter
- Backend risk assessment service generates recommendations based on:
  - Risk factor scores
  - Category risk levels
  - Historical patterns
  - Industry best practices

### Value to Users
- **Actionable Intelligence**: Transforms risk data into concrete action items
- **Prioritization**: Priority badges help users focus on most critical recommendations first
- **Implementation Guidance**: Difficulty and impact scores help users plan resource allocation
- **Workflow Integration**: Action buttons enable users to act directly from the dashboard
- **Risk Mitigation**: Direct path from risk identification to risk mitigation
- **Compliance Support**: Recommendations help ensure regulatory compliance and risk management best practices

**Current Limitation**: 
- Card may display "No specific recommendations available at this time" if:
  - Risk assessment API doesn't return recommendations
  - No high-risk factors trigger manual verification recommendations
  - Recommendation generation logic hasn't been fully implemented

---

## Data Flow Summary

### Complete Data Pipeline

1. **User Action**: User navigates to Risk Indicators tab for a merchant
2. **Frontend Initialization**: `MerchantRiskIndicatorsTab.init(merchantId)` called
3. **Data Loading**: `RiskIndicatorsDataService.loadAllRiskData(merchantId)` aggregates:
   - Merchant data from `/api/v1/merchants/{id}`
   - Stored analytics from business analytics results
   - Live risk assessment from `/api/v1/risk/assess`
4. **Data Normalization**: `mergeAndNormalize()` combines all sources into unified structure
5. **Category Building**: `buildRiskCategories()` creates 6 risk categories with scores and subcategories
6. **UI Rendering**: `RiskIndicatorsUITemplate` methods generate HTML for each card
7. **Visualization**: Charts and interactive elements initialized using D3.js and Chart.js
8. **Data Refresh**: 5-minute cache timeout ensures data freshness

### Key Backend Services

- **Risk Assessment Service**: `internal/api/handlers/risk.go` - `AssessRiskHandler()`
- **Risk Scorer**: `internal/modules/risk_assessment/risk_scorer.go` - `CalculateRiskScore()`
- **Risk Factor Calculator**: `internal/risk/calculation.go` - Category-specific calculations
- **Risk Models**: `services/risk-assessment-service/internal/models/risk_models.go` - Data structures

### Key Frontend Components

- **Main Controller**: `web/js/components/merchant-risk-indicators-tab.js`
- **UI Templates**: `web/js/components/risk-indicators-ui-template.js`
- **Data Service**: `web/js/components/risk-indicators-data-service.js`
- **Helpers**: `web/js/components/risk-indicators-helpers.js` (referenced but not shown in search results)

---

## Conclusion

The Risk Indicators tab provides a comprehensive, multi-faceted view of merchant risk through seven distinct cards, each serving a specific purpose:

1. **Badges** - Quick visual overview of top risks
2. **Heat Map** - Detailed subcategory risk mapping with legend
3. **Progress Indicators** - Detailed category breakdowns with context
4. **Radar Chart** - Visual comparison with industry benchmarks
5. **Category Analysis** - Tabular summary for quick reference
6. **Risk Summary** - Executive-level aggregated score
7. **Recommendations** - Actionable next steps for risk mitigation

All cards share the same underlying data sources but present information in different formats optimized for different use cases, from quick scanning to detailed analysis to actionable planning.

