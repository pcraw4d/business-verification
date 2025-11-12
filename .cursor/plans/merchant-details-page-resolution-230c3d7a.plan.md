<!-- 230c3d7a-d657-487d-a2ce-ebb7b6e9f4a0 f031fa91-3e89-4924-b2f0-5560cebde2a0 -->
# Merchant Details Page - Comprehensive Implementation Plan

## Overview

This plan addresses all issues identified in the UI review summary and resolution plan documents. The implementation is organized into 4 phases with detailed, actionable todos.

**Target File:** `cmd/frontend-service/static/merchant-details.html`
**Navigation Component:** `cmd/frontend-service/static/components/navigation.js`
**Session Manager:** `cmd/frontend-service/static/components/session-manager.js`

---

## Phase 1: Critical Fixes (Week 1)

### 1.1 Enable Left Navigation Sidebar

**File:** `cmd/frontend-service/static/components/navigation.js`

- Remove `'merchant-details'` from `skipNavigationPages` array (line 212)
- Verify navigation component initializes correctly on page load
- Test sidebar toggle button functionality
- Adjust main content area margin-left to accommodate sidebar (280px)
- Test responsive behavior on mobile devices (< 768px)
- Verify sidebar doesn't conflict with existing page layout
- Ensure sidebar state (open/closed) persists appropriately

**Success Criteria:** Sidebar visible, toggle works, no layout conflicts

---

### 1.2 Backend Data Flow Investigation

**Files to Review:**

- Backend API endpoints for merchant data
- Database schema and queries
- Form submission handlers in `add-merchant.html`

- Trace complete data flow: add-merchant form → API → database → merchant-details display
- Document all API endpoints used by merchant-details page:
- `/api/v1/merchants/{merchantId}` - Base merchant data
- Business Intelligence endpoints
- Risk Assessment endpoints
- Analytics endpoints
- Verify data persistence from form submission
- Check merchant ID generation and passing between pages
- Review backend logs for errors in data processing
- Create data mapping document: form fields → database columns → display components
- Test each endpoint with sample merchant data
- Document endpoint response structures

**Deliverables:** Data flow diagram, API endpoint inventory, error log analysis, data mapping document

---

### 1.3 Fix Data Loading and Population

**File:** `cmd/frontend-service/static/merchant-details.html`

- Fix data loading issues identified in backend investigation
- Ensure all cards populate with real data when available from form input
- Implement mock data fallback system:
- Create mock data repository/service
- Define fallback logic: Real Data → Mock Data → Empty State
- Ensure mock data structure matches real data structure
- Implement data source tracking (real/mock/empty) for each card
- Test with various form input combinations
- Verify data appears in correct format and location

**Success Criteria:** All cards show real data when available, mock data appears when real data unavailable

---

### 1.4 Implement Responsive Tab Navigation

**File:** `cmd/frontend-service/static/merchant-details.html`

- Implement responsive tab design:
- **Desktop (>1024px):** Show all 4 tabs horizontally
- **Tablet (768px-1024px):** Collapse overflow into "More" dropdown with chevron icon
- **Mobile (<768px):** Horizontal scroll with visible scroll indicators
- Rename "Merchant Details" tab button to "Overview" (line 411-414)
- Remove Overview, Contact, Financial, Compliance as separate top-level tabs (lines 427-442)
- Update tab button IDs and data-tab attributes
- Update tab content container IDs
- Add "More" dropdown for overflow tabs on tablet
- Implement smooth transitions between breakpoints
- Add active tab highlighting
- Ensure keyboard navigation and ARIA labels for accessibility
- Test across all standard viewport sizes

**Success Criteria:** All tabs accessible on all screen sizes, no horizontal scrolling on desktop, smooth responsive transitions

---

### 1.5 Fix Tab Switching and Unique Content

**File:** `cmd/frontend-service/static/merchant-details.html`

- Verify Business Analytics tab loads unique analytics content
- Verify Risk Assessment tab loads unique risk assessment content
- Verify Risk Indicators tab loads unique risk indicators content
- Fix tab click handlers in `switchTab()` method (around line 1930)
- Ensure each tab triggers correct data loading from backend
- Remove duplicate content across tabs
- Map backend endpoints to specific tabs:
- Business Analytics → Analytics API endpoints
- Risk Assessment → Risk Assessment API endpoints
- Risk Indicators → Risk Indicators API endpoints
- Test tab switching in isolation
- Verify route parameters are passed correctly

**Success Criteria:** Each tab displays unique content, tab switching works smoothly, no duplicate content

---

## Phase 2: High Priority (Week 2)

### 2.1 Restructure Overview Tab with Cards

**File:** `cmd/frontend-service/static/merchant-details.html`

- Update Overview tab content (currently `#merchant-details` tab, line 449)
- Create card layout with optimized spacing:
- **Overview Card:** High-level merchant summary (Business ID, Industry, Status, Key metrics)
- **Contact Card:** Contact information (Address, Phone, Email, Website)
- **Financial Card:** Financial data (Revenue, Employee Count, Founded Year, Transactions)
- **Compliance Card:** Compliance status (KYB Status, Verification Date, Compliance Score, Certifications)
- Implement flexible grid layout:
- Desktop: 2 cards in first row, 2 cards in second row (or best-fit arrangement)
- Tablet: 2 cards per row
- Mobile: 1 card per row
- Ensure cards are visually distinct with proper spacing and styling
- Add loading states for each card
- Implement error states for cards
- Maintain existing data population logic from `loadOverviewTab()`, `loadContactTab()`, `loadFinancialTab()`, `loadComplianceTab()` methods
- Update tab switching logic to handle renamed tab

**Success Criteria:** Four distinct cards in optimized layout, all cards populate with data, responsive on all devices

---

### 2.2 Implement Mock Data Tooltip System

**File:** `cmd/frontend-service/static/merchant-details.html`

- Create reusable tooltip component with:
- Auto-placement algorithm (smart positioning to avoid viewport edges)
- Hover trigger functionality
- Minimal, professional styling
- Create tooltip content templates for each card type:
- Identify which specific data points are mock
- Explain why mock data is used (e.g., "External data source not yet integrated")
- Add tooltips to all cards using mock data
- Implement data source tracking system
- Add subtle visual indicators (info icon) for cards with mock data
- Remove or hide global "Mock Data Warning" banner (if exists)
- Test tooltip appearance and behavior
- Ensure tooltips don't interfere with user experience

**Success Criteria:** Tooltips appear on all mock data cards, content is clear and informative, no global warning banner

---

### 2.3 Verify Tab Content Uniqueness

**File:** `cmd/frontend-service/static/merchant-details.html`

- Audit content for each tab:
- Business Analytics tab content
- Risk Assessment tab content
- Risk Indicators tab content
- Verify each tab has dedicated backend data source
- Test each tab independently
- Verify no content duplication across tabs
- Document content mapping for each tab
- Test data loading performance for each tab

**Success Criteria:** Each tab shows unique content, no content duplication, all tabs load correctly

---

### 2.4 Comprehensive Navigation and Data Flow Testing

- Test complete flow: Add merchant → View merchant details → Verify all tabs show correct data
- Test with various form inputs from add-merchant form
- Verify data appears correctly in all cards
- Test mock data fallback scenarios
- Verify tooltips appear correctly on cards with mock data
- Test all navigation paths (sidebar, tabs, buttons)
- Test tab switching functionality
- Test responsive navigation on all devices
- Test error scenarios:
- Missing data
- API failures
- Network errors
- Verify error handling and loading states

**Success Criteria:** All navigation paths work, data flows correctly, error handling works, no critical bugs

---

## Phase 3: Medium Priority (Week 3-4)

### 3.1 Implement Fixed Footer with Session Buttons

**File:** `cmd/frontend-service/static/merchant-details.html`

- Create fixed footer component positioned at bottom of viewport
- Implement enterprise SaaS-style design:
- Subtle background (semi-transparent or light)
- Clear button styling
- Proper spacing and padding
- Professional appearance
- Move History and End Session buttons from sidebar to footer
- Enable both buttons (currently disabled in session-manager.js)
- Implement button functionality:
- **History Button:** Navigate to history page (session history view)
- **End Session Button:** Clear current session data and redirect to dashboard
- Ensure footer stays above content but below modals (z-index management)
- Implement responsive design for mobile devices
- Ensure accessibility compliance (keyboard navigation, screen readers)
- Test on various mobile devices

**Success Criteria:** Footer always visible, buttons work correctly, professional appearance, mobile-friendly

---

### 3.2 Integrate Data Enrichment Component UI

**File:** `cmd/frontend-service/static/merchant-details.html`

- Determine optimal placement (Overview tab or Business Analytics tab)
- Create container element for Data Enrichment UI
- Initialize Data Enrichment component with UI (currently initialized but no UI, line 2239)
- Display available enrichment sources (Thomson Reuters, etc.)
- Show enrichment status
- Add button to trigger enrichment
- Display enrichment results
- Connect to enrichment API endpoints
- Handle enrichment requests and display progress
- Add proper styling and spacing

**Success Criteria:** UI is visible and accessible, enrichment functionality works, results display correctly

---

### 3.3 Integrate External Data Sources Component

**File:** `cmd/frontend-service/static/merchant-details.html`

- Create container element `#externalDataSourcesContainer`
- Add to appropriate tab (Business Analytics or Overview)
- Initialize component: `new ExternalDataSources('externalDataSourcesContainer')`
- Verify component loads sources correctly
- Display sources list with status (active/inactive)
- Handle loading and error states
- Add proper styling
- Make responsive

**Success Criteria:** Component is visible, sources load correctly, status displays accurately

---

### 3.4 Add Risk Configuration Controls

**File:** `cmd/frontend-service/static/merchant-details.html`

- Add "Configure Risk Factors" button in Risk Assessment tab
- Create toggle functionality to show/hide `#riskConfigContainer` (currently hidden, line 989)
- Remove `display: none` from container
- Control visibility via toggle button
- Add smooth show/hide animation
- Verify RiskDragDrop component initializes correctly
- Test drag-and-drop functionality
- Ensure configuration saves correctly
- Add visual feedback during drag operations
- Make button discoverable and accessible

**Success Criteria:** Button is visible and discoverable, drag-drop works correctly, configuration saves

---

### 3.5 Verify and Populate Risk Score Panel & Website Risk Display

**File:** `cmd/frontend-service/static/merchant-details.html`

- Verify RiskScorePanel component initialization (container exists at line 983)
- Check data is being passed to panel
- Verify panel displays risk score breakdown with factors
- Add loading state if needed
- Verify WebsiteRiskDisplay component initialization (container exists at line 986)
- Check if website data is available for current merchant
- Add placeholder/loading state if data missing
- Ensure both components receive merchant ID correctly
- Check API endpoints are called correctly
- Ensure data structure matches component expectations
- Add error handling for both components

**Success Criteria:** Both components display data, loading states work, error handling implemented

---

### 3.6 Add Export Buttons to All Tabs

**File:** `cmd/frontend-service/static/merchant-details.html`

- Add export buttons to Business Analytics tab
- Add export buttons to Risk Indicators tab
- Verify existing Risk Assessment export buttons work (initialized at line 2218)
- Ensure consistent styling across tabs
- Verify export functionality works for all data types:
- CSV export
- PDF export
- JSON export
- Excel export
- Add export progress indicators
- Ensure each tab prepares data correctly for export
- Format data appropriately for each export type
- Include all relevant data in exports
- Handle export errors gracefully

**Success Criteria:** All tabs have export buttons, exports work correctly, progress indicators show

---

### 3.7 Integrate Merchant Context Component

**File:** `cmd/frontend-service/static/merchant-details.html`

- Verify MerchantContext component is being initialized (loaded at line 14)
- Check if context elements are being created
- Verify context doesn't conflict with page layout
- Ensure context is visible and useful
- Integrate context with page layout
- Add breadcrumbs if applicable
- Display merchant info in header/sidebar if needed
- Test context display and updates
- Ensure context is accessible

**Success Criteria:** Context is visible, no layout conflicts, context is useful

---

### 3.8 Load and Integrate Session Manager UI

**File:** `cmd/frontend-service/static/merchant-details.html`

- Review `session-manager-ui.js` component (exists but not loaded)
- Determine if it provides needed functionality beyond `session-manager.js`
- If needed, load component in HTML (add script tag)
- Initialize component appropriately
- Verify session management UI works
- Test session history display
- Ensure session controls function correctly
- Verify no conflicts with other components

**Success Criteria:** Session management works, UI is functional if needed, no conflicts

---

### 3.9 Research and Document Risk WebSocket Service

**File:** `cmd/frontend-service/static/merchant-details.html`

- Research service purpose and benefits (already documented in ui_review_summary.md)
- Document implementation steps:
- Backend verification: `/ws/risk-assessment/{assessment_id}` endpoint
- Frontend integration: Enable WebSocket connection in MerchantRiskTab (uncomment lines 67-71)
- UI indicators: Connection status, "Real-time updates" badge
- Event handling: riskUpdate, riskPrediction, riskAlert events
- Create decision document with effort estimate (15-23 hours)
- If approved, implement:
- Enable WebSocket connection
- Add connection status indicator
- Show "Real-time updates" badge when connected
- Add reconnection status messages
- Test real-time updates
- Verify backend service is accessible

**Success Criteria:** Service purpose documented, implementation plan complete, decision made, implementation working (if approved)

---

### 3.10 Enhance Tooltip System for All Cards

**File:** `cmd/frontend-service/static/merchant-details.html`

- Enhance tooltip component for card-level use
- Implement auto-placement algorithm
- Add hover trigger
- Ensure minimal styling
- Create tooltip content for each card type:
- Overview Card tooltips
- Contact Card tooltips
- Financial Card tooltips
- Compliance Card tooltips
- Business Analytics card tooltips
- Risk Assessment card tooltips
- Risk Indicators card tooltips
- Add tooltips to all cards using mock data
- Implement data source tracking
- Add visual indicators (subtle info icons)
- Test tooltip placement and content
- Verify tooltips don't interfere with UX

**Success Criteria:** Tooltips appear correctly, content is clear, placement is optimal, no UX interference

---

### 3.11 Populate All Expandable Sections

**File:** `cmd/frontend-service/static/merchant-details.html`

- Verify all expandable sections have content when expanded:
- `#coreDetails` (Core Classification Results)
- `#securityDetails` (Security & Trust Indicators)
- `#qualityDetails` (Data Quality Metrics)
- `#riskDetails` (Risk Assessment)
- `#intelligenceDetails` (Business Intelligence)
- `#verificationDetails` (Verification Status)
- Populate sections with real or mock data (with tooltips)
- Add loading states for sections that fetch data on expand
- Ensure smooth expand/collapse animations
- Verify "Show Details" buttons work correctly
- Test expand/collapse functionality
- Verify content loads correctly with real and mock data

**Success Criteria:** All sections have content, expand/collapse works smoothly, loading states work, animations are smooth

---

## Phase 4: Polish and Testing (Ongoing)

### 4.1 Accessibility Audit

- Conduct comprehensive accessibility audit (WCAG 2.1 compliance)
- Fix identified accessibility issues
- Add ARIA labels where needed
- Test with screen readers (NVDA, JAWS, VoiceOver)
- Ensure keyboard navigation works for all interactive elements
- Verify color contrast ratios meet WCAG standards
- Add skip navigation links
- Ensure focus indicators are visible

**Success Criteria:** WCAG 2.1 AA compliance, all interactive elements keyboard accessible

---

### 4.2 Mobile Responsive Design Improvements

- Conduct mobile device testing on:
- iOS devices (iPhone, iPad)
- Android devices (various screen sizes)
- Fix responsive design issues identified
- Optimize touch interactions (minimum 44x44px touch targets)
- Test on various mobile devices and orientations
- Ensure all features work on mobile
- Optimize images for mobile
- Test performance on mobile networks

**Success Criteria:** All features work on mobile, touch interactions optimized, performance acceptable

---

### 4.3 Export Functionality Testing

- Test all export formats:
- CSV export
- PDF export
- JSON export
- Excel export
- Verify export data accuracy
- Test export with various data scenarios:
- Full data
- Partial data
- Empty data
- Large datasets
- Add export error handling
- Optimize export performance
- Test export file downloads
- Verify exported files open correctly

**Success Criteria:** All export formats work, data accuracy verified, error handling implemented

---

### 4.4 Performance Optimization

- Measure page load times (target: < 3 seconds)
- Optimize API calls (batch requests where possible)
- Implement data caching where appropriate
- Optimize image loading (lazy loading, WebP format)
- Reduce JavaScript bundle size if needed
- Minimize CSS
- Implement code splitting if beneficial
- Profile and identify bottlenecks
- Optimize database queries (backend)

**Success Criteria:** Page load < 3 seconds, API calls optimized, images optimized

---

### 4.5 Cross-Browser Testing

- Test in all major browsers:
- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)
- Fix browser-specific issues
- Ensure consistent appearance across browsers
- Test functionality across browsers
- Verify CSS compatibility
- Test JavaScript compatibility
- Verify polyfills where needed

**Success Criteria:** Consistent appearance and functionality across all browsers

---

## Implementation Notes

### Key Files to Modify

1. **`cmd/frontend-service/static/merchant-details.html`** - Main page file
2. **`cmd/frontend-service/static/components/navigation.js`** - Navigation component
3. **`cmd/frontend-service/static/components/session-manager.js`** - Session manager component

### Testing Strategy

- Unit tests for individual components
- Integration tests for data flow
- End-to-end tests for user workflows
- Cross-browser testing
- Mobile device testing
- Accessibility testing
- Performance testing

### Success Metrics

- All Phase 1 tasks completed
- All Phase 2 tasks completed
- All Phase 3 tasks completed
- Beta-ready MVP showcasing all capabilities
- All features functional and tested
- Professional appearance maintained
- Ready for beta tester feedback

### To-dos

- [ ] Enable left navigation sidebar: Remove merchant-details from skipNavigationPages array in navigation.js, verify initialization, test toggle functionality, adjust layout margins, test responsive behavior
- [ ] Backend data flow investigation: Trace data flow from add-merchant form to merchant-details, document all API endpoints, verify data persistence, check merchant ID passing, review error logs, create data mapping document
- [ ] Fix data loading and population: Fix identified issues, ensure real data populates from form, implement mock data fallback system with data source tracking, test with various form inputs
- [ ] Implement responsive tab navigation: Desktop show all tabs, tablet collapse overflow into More dropdown, mobile horizontal scroll, rename Merchant Details to Overview, remove Overview/Contact/Financial/Compliance as separate tabs
- [ ] Fix tab switching and unique content: Verify each tab loads unique content, fix tab click handlers, map backend endpoints to tabs, remove duplicate content, test tab switching
- [ ] Restructure Overview tab with cards: Create Overview, Contact, Financial, Compliance cards in optimized grid layout, maintain data population logic, add loading/error states, ensure responsive design
- [ ] Implement mock data tooltip system: Create reusable tooltip component with auto-placement, add tooltips to all cards with mock data, create content templates, remove global warning banner
- [ ] Verify tab content uniqueness: Audit content for each tab, verify dedicated backend sources, test independently, document content mapping, verify no duplication
- [ ] Comprehensive navigation and data flow testing: Test complete add-merchant to merchant-details flow, test with various inputs, test navigation paths, test error scenarios, verify error handling
- [ ] Implement fixed footer with session buttons: Create enterprise SaaS-style fixed footer, move History and End Session buttons from sidebar, enable buttons, implement functionality, ensure responsive and accessible
- [ ] Integrate Data Enrichment component UI: Determine placement, create container, initialize component with UI, display sources and status, add trigger button, connect to API, handle requests
- [ ] Integrate External Data Sources component: Create container, initialize component, display sources list with status, handle loading/error states, add styling, make responsive
- [ ] Add Risk Configuration controls: Add Configure Risk Factors button, create toggle for riskConfigContainer, verify drag-drop functionality, ensure configuration saves, add visual feedback
- [ ] Verify and populate Risk Score Panel and Website Risk Display: Verify initialization, check data passing, add loading states, ensure merchant ID received, add error handling
- [ ] Add export buttons to all tabs: Add to Business Analytics and Risk Indicators tabs, verify functionality for all formats (CSV, PDF, JSON, Excel), add progress indicators, handle errors
- [ ] Integrate Merchant Context component: Verify initialization, check context elements, ensure no layout conflicts, add breadcrumbs if needed, test context display and updates
- [ ] Load and integrate Session Manager UI: Review component, determine if needed, load if required, initialize appropriately, verify functionality, test session history
- [ ] Research and document Risk WebSocket service: Document service purpose, create implementation plan, create decision document with effort estimate, implement if approved
- [ ] Enhance tooltip system for all cards: Enhance component, implement auto-placement, create content for all card types, add to all mock data cards, test placement and content
- [ ] Populate all expandable sections: Verify all sections have content, populate with real/mock data, add loading states, ensure smooth animations, test expand/collapse functionality
- [ ] Conduct accessibility audit: WCAG 2.1 compliance check, fix issues, add ARIA labels, test with screen readers, ensure keyboard navigation, verify color contrast
- [ ] Mobile responsive design improvements: Test on iOS and Android devices, fix responsive issues, optimize touch interactions, test all features on mobile, optimize images
- [ ] Export functionality testing: Test all export formats, verify data accuracy, test various scenarios, add error handling, optimize performance, test file downloads
- [ ] Performance optimization: Measure page load times, optimize API calls, implement caching, optimize images, reduce bundle size, profile bottlenecks, optimize queries
- [ ] Cross-browser testing: Test in Chrome, Firefox, Safari, Edge, fix browser-specific issues, ensure consistent appearance, test functionality, verify compatibility