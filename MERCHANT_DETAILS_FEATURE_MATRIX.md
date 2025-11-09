# Merchant Details Feature Matrix

## Overview
This document provides a comprehensive comparison of all features across the four merchant detail pages to guide consolidation into a single unified `merchant-details.html` page.

## Pages Analyzed
1. `services/frontend/public/merchant-details.html` (Base - 4 tabs)
2. `services/frontend/public/merchant-detail.html` (5 tabs)
3. `services/frontend/public/merchant-details-new.html` (4 tabs - similar to base)
4. `services/frontend/public/merchant-details-old.html` (4 tabs - similar to base)

---

## Tab Structure Comparison

### merchant-details.html (Base)
- ✅ Merchant Details
- ✅ Business Analytics
- ✅ Risk Assessment
- ✅ Risk Indicators

### merchant-detail.html
- ✅ Overview
- ✅ Contact
- ✅ Financial
- ✅ Risk Assessment
- ✅ Compliance

### merchant-details-new.html
- ✅ Merchant Details
- ✅ Business Analytics
- ✅ Risk Assessment
- ✅ Risk Indicators

### merchant-details-old.html
- ✅ Merchant Details
- ✅ Business Analytics
- ✅ Risk Assessment
- ✅ Risk Indicators

**Consolidated Tab Structure (Target):**
1. Merchant Details
2. Business Analytics
3. Risk Assessment
4. Risk Indicators
5. Overview (from merchant-detail.html)
6. Contact (from merchant-detail.html)
7. Financial (from merchant-detail.html)
8. Compliance (from merchant-detail.html)

---

## Feature Inventory

### 1. Merchant Details Tab

#### Source: merchant-details.html, merchant-details-new.html, merchant-details-old.html
**Features:**
- Business Information Display
  - Business Name
  - Website URL
  - Business Description
  - Registration Number
- Address Information Display
  - Street Address
  - City
  - State/Province
  - Postal Code
  - Country
- Contact Information Display
  - Phone Number
  - Email Address

**Dependencies:**
- `components/navigation.js`
- `components/merchant-context.js`
- Data from `sessionStorage.getItem('merchantData')`

**Priority:** Critical

---

### 2. Business Analytics Tab

#### Source: merchant-details.html, merchant-details-new.html, merchant-details-old.html
**Features:**
- Core Classification Results
  - Primary Industry Display
  - Industry Code Display
  - Confidence Score Display
  - Risk Level Display
  - Expandable Details Section
- Enhanced Classification Results
  - Top 3 MCC Codes (with descriptions and confidence)
  - Top 3 SIC Codes (with descriptions and confidence)
  - Top 3 NAICS Codes (with descriptions and confidence)
  - Method Breakdown Section
  - Website Keywords Section (toggleable)
  - Classification Reasoning Section
- Security & Trust Indicators
  - Data Source Trust Indicator
  - Website Verification Indicator
  - Security Validation Indicator
  - Trust Score Display
  - Expandable Security Details
- Data Quality Metrics
  - Overall Quality Grade
  - Evidence Strength
  - Data Completeness Percentage
  - Agreement Score
  - Consistency Score
  - Expandable Quality Details
- Risk Assessment Summary
  - Overall Risk Score
  - Compliance Risk
  - Financial Risk
  - Operational Risk
  - Expandable Risk Details
- Business Intelligence
  - Employee Count Range
  - Revenue Range
  - Founded Year
  - Business Location
  - Expandable Intelligence Details
- Verification Status
  - Verification Status
  - Processing Time
  - Data Sources Count
  - Last Updated Timestamp
  - Expandable Verification Details

**Dependencies:**
- `components/navigation.js`
- `components/merchant-context.js`
- Data from `sessionStorage.getItem('merchantApiResults')`
- API: Business Intelligence API

**Priority:** Critical

---

### 3. Risk Assessment Tab

#### Source: merchant-details.html, merchant-detail.html
**Features:**
- Risk Score Panel
  - Overall Risk Score Display
  - Risk Score Trend Indicator
  - Risk Category Breakdown
- Risk Visualization
  - D3.js Risk Gauge
  - Risk Trend Charts (Chart.js)
  - Risk Category Visualizations
- Risk Explainability
  - SHAP Values Display
  - Feature Importance Charts
  - Risk Factor Explanations
- Risk Scenario Analysis
  - Scenario Testing Interface
  - What-If Analysis
  - Risk Sensitivity Analysis
- Risk History Tracking
  - Historical Risk Scores
  - Risk Trend Over Time
  - Risk Event Timeline
- Risk Export Functionality
  - PDF Export
  - Excel Export
  - CSV Export
- Website Risk Display
  - Website Security Assessment
  - SSL Certificate Status
  - Domain Risk Indicators
- Real-Time WebSocket Updates
  - Live Risk Score Updates
  - Real-Time Risk Alerts
  - Connection Status Indicator
- Risk Configuration (Drag & Drop)
  - Customizable Risk Factors
  - Risk Weight Configuration
  - Drag-and-Drop Interface

**Dependencies:**
- `js/components/risk-websocket-client.js`
- `js/components/risk-visualization.js`
- `js/components/risk-explainability.js`
- `js/components/risk-scenarios.js`
- `js/components/risk-history.js`
- `js/components/risk-export.js`
- `js/components/risk-tooltip-system.js`
- `js/components/risk-score-panel.js`
- `js/components/risk-drag-drop.js`
- `js/components/website-risk-display.js`
- `js/merchant-risk-tab.js`
- D3.js v7: `https://d3js.org/d3.v7.min.js`
- Chart.js: `https://cdn.jsdelivr.net/npm/chart.js`
- API: Risk Assessment API

**Priority:** Critical

---

### 4. Risk Indicators Tab

#### Source: merchant-details.html, merchant-details-new.html, merchant-details-old.html
**Features:**
- Risk Indicators Display
  - Risk Level Indicators
  - Risk Badge System
  - Risk Trend Indicators
- Risk Visualization
  - D3.js Visualizations
  - Risk Gauge Charts
- Risk Keywords Section
  - Detected Keywords Display
  - Keyword Severity Indicators
  - Keyword Categories
  - Keyword Highlighting
- Risk Tooltip System
  - Interactive Tooltips
  - Mobile-Friendly Tooltips
- Risk Indicators Data Service
  - Real-Time Data Loading
  - API Integration

**Dependencies:**
- `js/api-config.js`
- `js/components/real-data-integration.js`
- `js/components/risk-visualization.js`
- `js/components/risk-explainability.js`
- `js/components/risk-level-indicator.js`
- `js/utils/risk-indicators-helpers.js`
- `js/components/risk-indicators-ui-template.js`
- `js/components/risk-indicators-data-service.js`
- `js/components/website-risk-display.js`
- `js/components/merchant-risk-indicators-tab.js`
- D3.js v7: `https://d3js.org/d3.v7.min.js`
- `css/risk-indicators.css`
- API: Risk Indicators API

**Priority:** Critical

---

### 5. Overview Tab

#### Source: merchant-detail.html
**Features:**
- Business Overview Card
  - Business ID
  - Industry
  - Portfolio Type
  - Risk Level
  - Status
  - Created Date
- Recent Activity Card
  - Activity Timeline
  - Activity Types (Transaction, Verification, Update, Alert)
  - Activity Descriptions
  - Activity Dates
  - Activity Amounts (if applicable)

**Dependencies:**
- `components/navigation.js`
- `components/session-manager.js`
- `components/coming-soon-banner.js`
- `components/mock-data-warning.js`
- `components/real-data-integration.js`
- `merchant-dashboard-real-data.js`
- API: Merchant Details API (`/api/v1/merchants/{id}`)

**Priority:** High

---

### 6. Contact Tab

#### Source: merchant-detail.html
**Features:**
- Contact Information Card
  - Address
  - Phone
  - Email
  - Website (with link)

**Dependencies:**
- `components/navigation.js`
- `components/session-manager.js`
- API: Merchant Details API

**Priority:** High

---

### 7. Financial Tab

#### Source: merchant-detail.html
**Features:**
- Financial Information Card
  - Annual Revenue
  - Monthly Volume
  - Transaction Count
  - Average Transaction Amount
  - Employee Count
  - Founded Year
  - Compliance Score

**Dependencies:**
- `components/navigation.js`
- `components/session-manager.js`
- API: Merchant Details API

**Priority:** High

---

### 8. Compliance Tab

#### Source: merchant-detail.html
**Features:**
- Compliance Status Card
  - KYB Status (with badge)
  - Last Verification Date
  - Compliance Score
  - Next Review Date

**Dependencies:**
- `components/navigation.js`
- `components/session-manager.js`
- API: Merchant Details API

**Priority:** High

---

## Additional Features from merchant-detail.html

### Header Features
- Merchant Avatar (Initial-based)
- Merchant Name Display
- Merchant Industry Display
- Merchant Badges
  - Portfolio Type Badge
  - Risk Level Badge
  - Status Badge
- Header Actions
  - Export Button (with dropdown: CSV, PDF, JSON)
  - Enrich Data Button
  - Edit Merchant Button
  - View Portfolio Button

**Dependencies:**
- `js/components/export-button.js`
- `js/components/data-enrichment.js`
- `js/components/external-data-sources.js`
- API: Data Enrichment API

**Priority:** High

---

## JavaScript Components Inventory

### Core Components (All Pages)
- `components/navigation.js` - Navigation system
- `components/merchant-context.js` - Merchant context management

### Risk Assessment Components
- `js/components/risk-websocket-client.js` - Real-time WebSocket client
- `js/components/risk-visualization.js` - D3.js visualizations
- `js/components/risk-explainability.js` - SHAP explainability
- `js/components/risk-scenarios.js` - Scenario analysis
- `js/components/risk-history.js` - Historical tracking
- `js/components/risk-export.js` - Export functionality
- `js/components/risk-tooltip-system.js` - Tooltip system
- `js/components/risk-score-panel.js` - Score panel
- `js/components/risk-drag-drop.js` - Drag and drop
- `js/components/website-risk-display.js` - Website risk
- `js/merchant-risk-tab.js` - Main risk tab controller

### Risk Indicators Components
- `js/api-config.js` - API configuration
- `js/components/real-data-integration.js` - Real data integration
- `js/components/risk-level-indicator.js` - Level indicators
- `js/utils/risk-indicators-helpers.js` - Helper functions
- `js/components/risk-indicators-ui-template.js` - UI templates
- `js/components/risk-indicators-data-service.js` - Data service
- `js/components/merchant-risk-indicators-tab.js` - Indicators tab controller

### Additional Components (merchant-detail.html)
- `components/session-manager.js` - Session management
- `components/coming-soon-banner.js` - Coming soon banner
- `components/mock-data-warning.js` - Mock data warning
- `components/real-data-integration.js` - Real data integration
- `merchant-dashboard-real-data.js` - Dashboard real data
- `js/components/export-button.js` - Export button component
- `js/components/data-enrichment.js` - Data enrichment
- `js/components/external-data-sources.js` - External data sources

---

## External Libraries

### Visualization Libraries
- **D3.js v7**: `https://d3js.org/d3.v7.min.js`
  - Used for: Risk gauges, risk visualizations, risk indicators
- **Chart.js**: `https://cdn.jsdelivr.net/npm/chart.js`
  - Used for: Risk trend charts, risk category charts

### CSS Frameworks
- **Tailwind CSS**: `https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css`
- **Font Awesome**: `https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css`
- **Custom CSS**: `css/risk-indicators.css`

---

## API Endpoints

### Business Intelligence API
- Endpoint: `/api/v1/business-intelligence`
- Used by: Business Analytics tab

### Risk Assessment API
- Endpoint: `/api/v1/risk-assessment`
- Used by: Risk Assessment tab

### Risk Indicators API
- Endpoint: `/api/v1/risk-indicators`
- Used by: Risk Indicators tab

### Merchant Details API
- Endpoint: `/api/v1/merchants/{id}`
- Used by: Overview, Contact, Financial, Compliance tabs

### Data Enrichment API
- Endpoint: `/api/v1/enrichment`
- Used by: Data enrichment feature

### WebSocket Endpoint
- Endpoint: `ws://` or `wss://` (for real-time risk updates)
- Used by: Risk Assessment tab (real-time updates)

---

## Data Sources

### Session Storage
- `merchantData` - Merchant form data
- `merchantApiResults` - API results from verification

### URL Parameters
- `merchantId` or `id` - Merchant identifier

### Local Storage
- Not used in current implementation

---

## Feature Priority Classification

### Critical Priority (Must Have)
1. ✅ Merchant Details Tab - Core business information
2. ✅ Business Analytics Tab - Classification results and analytics
3. ✅ Risk Assessment Tab - Comprehensive risk analysis
4. ✅ Risk Indicators Tab - Risk indicators and visualizations
5. ✅ Tab Navigation System - All tabs must be functional
6. ✅ Data Loading from sessionStorage - Core data flow
7. ✅ API Integration - All API endpoints must work

### High Priority (Should Have)
1. ✅ Overview Tab - Business overview and activity
2. ✅ Contact Tab - Contact information
3. ✅ Financial Tab - Financial information
4. ✅ Compliance Tab - Compliance status
5. ✅ Export Functionality - PDF, Excel, CSV export
6. ✅ Data Enrichment - Enrich data button
7. ✅ Real-Time WebSocket Updates - Live risk updates
8. ✅ Header Actions - Export, Enrich, Edit, View Portfolio

### Medium Priority (Nice to Have)
1. ✅ Risk Configuration (Drag & Drop) - Customizable risk factors
2. ✅ Website Risk Display - Website security assessment
3. ✅ External Data Sources - External data integration
4. ✅ Progressive Disclosure - Expandable sections
5. ✅ Skeleton Loading - Loading states

### Low Priority (Future Enhancements)
1. ✅ Mobile Tooltip System - Enhanced mobile experience
2. ✅ Keyword Highlighting - Advanced keyword features
3. ✅ Risk Tooltip System - Enhanced tooltips

---

## Feature Conflicts and Overlaps

### Overlapping Features
1. **Risk Assessment Tab** - Exists in both merchant-details.html and merchant-detail.html
   - **Resolution**: Use the more comprehensive implementation from merchant-detail.html (includes WebSocket, export, scenarios, history)

2. **Merchant Information Display** - Exists in Merchant Details tab and Overview/Contact tabs
   - **Resolution**: Keep detailed view in Merchant Details tab, summary in Overview/Contact tabs

3. **Financial Information** - Exists in Business Analytics tab and Financial tab
   - **Resolution**: Keep summary in Business Analytics, detailed view in Financial tab

### Conflicting Implementations
1. **Tab Navigation System**
   - merchant-details.html: Simple tab switching
   - merchant-detail.html: More robust TabManager class with keyboard navigation
   - **Resolution**: Use TabManager from merchant-detail.html for better UX

2. **Risk Assessment Initialization**
   - merchant-details.html: Basic initialization
   - merchant-detail.html: Comprehensive initialization with loading states
   - **Resolution**: Use comprehensive initialization from merchant-detail.html

---

## Missing Features in merchant-details.html

### From merchant-detail.html
1. ❌ Overview Tab - Missing
2. ❌ Contact Tab - Missing
3. ❌ Financial Tab - Missing
4. ❌ Compliance Tab - Missing
5. ❌ Export Button Component - Missing
6. ❌ Data Enrichment Component - Missing
7. ❌ External Data Sources Component - Missing
8. ❌ TabManager Class - Missing (better tab navigation)
9. ❌ Session Manager - Missing
10. ❌ Coming Soon Banner - Missing
11. ❌ Mock Data Warning - Missing
12. ❌ Merchant Avatar - Missing
13. ❌ Merchant Badges - Missing
14. ❌ Header Actions - Missing

---

## Dependencies Summary

### Required JavaScript Files
1. Core:
   - `components/navigation.js`
   - `components/merchant-context.js`
   - `components/session-manager.js`

2. Risk Assessment:
   - `js/components/risk-websocket-client.js`
   - `js/components/risk-visualization.js`
   - `js/components/risk-explainability.js`
   - `js/components/risk-scenarios.js`
   - `js/components/risk-history.js`
   - `js/components/risk-export.js`
   - `js/components/risk-tooltip-system.js`
   - `js/components/risk-score-panel.js`
   - `js/components/risk-drag-drop.js`
   - `js/components/website-risk-display.js`
   - `js/merchant-risk-tab.js`

3. Risk Indicators:
   - `js/api-config.js`
   - `js/components/real-data-integration.js`
   - `js/components/risk-level-indicator.js`
   - `js/utils/risk-indicators-helpers.js`
   - `js/components/risk-indicators-ui-template.js`
   - `js/components/risk-indicators-data-service.js`
   - `js/components/merchant-risk-indicators-tab.js`

4. Additional:
   - `js/components/export-button.js`
   - `js/components/data-enrichment.js`
   - `js/components/external-data-sources.js`
   - `components/coming-soon-banner.js`
   - `components/mock-data-warning.js`

### Required CSS Files
1. `css/risk-indicators.css`

### Required External Libraries
1. D3.js v7: `https://d3js.org/d3.v7.min.js`
2. Chart.js: `https://cdn.jsdelivr.net/npm/chart.js`
3. Tailwind CSS: `https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css`
4. Font Awesome: `https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css`

---

## Implementation Notes

### Tab Navigation
- Use TabManager class from merchant-detail.html for better keyboard navigation and accessibility
- Ensure all 8 tabs can be switched smoothly
- Maintain active state management

### Data Loading
- Primary source: sessionStorage (`merchantData`, `merchantApiResults`)
- Fallback: URL parameters (`merchantId` or `id`)
- API calls for real-time data updates

### Initialization Order
1. Load core components (navigation, merchant-context)
2. Load merchant data from sessionStorage
3. Initialize tab navigation
4. Initialize risk components (after D3.js and Chart.js load)
5. Initialize WebSocket client (if needed)
6. Load API data

### Error Handling
- Graceful degradation if components fail to load
- Fallback UI for missing data
- Error messages for API failures

---

## Success Criteria

✅ All 8 tabs functional and displaying correct data
✅ All features from all 4 pages preserved and working
✅ No JavaScript errors in console
✅ All visualizations render correctly
✅ Export functionality works for all formats
✅ Real-time updates via WebSocket functional
✅ Responsive design works on all screen sizes
✅ Redirects from old URLs work correctly
✅ No broken links in navigation
✅ Documentation updated

---

## Next Steps

1. ✅ Phase 1 Complete: Feature audit and matrix creation
2. ⏭️ Phase 2: Base consolidation - Use merchant-details.html as base, add missing tabs
3. ⏭️ Phase 3: Feature integration - Integrate all components
4. ⏭️ Phase 4: Testing & validation
5. ⏭️ Phase 5: Cleanup & redirects

---

**Document Version:** 1.0  
**Last Updated:** 2025-01-27  
**Status:** Phase 1 Complete - Ready for Phase 2
