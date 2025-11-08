# Unused Features Analysis

**Date**: November 7, 2025  
**Analysis Type**: Comprehensive Review of Unused API Endpoints and Frontend Components  
**Total Unused Endpoints**: 37  
**Total Unused Components**: 72

---

## Executive Summary

This analysis categorizes and evaluates 37 unused API endpoints and 72 potentially unused frontend components identified in the KYB Platform codebase. The analysis provides:

1. **Categorization** by feature type and purpose
2. **Purpose Analysis** for each category
3. **Recommendations** for handling each feature
4. **Priority Assessment** for integration or removal
5. **Risk Analysis** for removal decisions

---

## Part 1: Unused API Endpoints Analysis

### Category 1: Beta Testing & Feedback Management (2 endpoints)

#### Endpoints
- `/invites` (POST) - Send beta tester invites
- `/bugs` (POST) - Submit bug reports

#### Purpose Analysis
These endpoints are part of a beta testing infrastructure designed to:
- Manage beta tester invitations and onboarding
- Collect bug reports and feedback from beta testers
- Support controlled feature rollouts

**Implementation Status**: Fully implemented in `services/risk-assessment-service/beta-testing/`

#### Recommendations

**Option A: Keep for Future Use (Recommended)**
- **Priority**: Medium
- **Action**: Document as "Beta Testing Infrastructure - Reserved for Future Use"
- **Rationale**: 
  - Complete implementation exists
  - Useful for controlled feature rollouts
  - Low maintenance cost (already implemented)
  - May be needed for future beta programs

**Option B: Remove**
- **Priority**: Low
- **Action**: Remove if no beta testing program is planned within 12 months
- **Rationale**: Reduces codebase complexity if truly unused

**Decision Matrix**:
- Keep if: Beta testing program planned within 12 months OR feature flags needed
- Remove if: No beta testing planned AND code maintenance is a concern

---

### Category 2: System Monitoring & Performance (12 endpoints)

#### Endpoints
- `/v1/memory/profile` (GET)
- `/v1/memory/profile/history` (GET)
- `/v1/memory/pools` (GET, POST)
- `/v1/memory/optimize` (POST)
- `/v1/thresholds` (GET, PUT)
- `/v1/risk/thresholds` (GET)
- `/system` (GET)
- `/query` (GET) - Prometheus query endpoint

#### Purpose Analysis
These endpoints provide:
- **Memory Management**: Profiling, optimization, leak detection, pooling
- **Performance Monitoring**: System metrics, resource thresholds
- **Operational Control**: Manual optimization triggers, threshold configuration

**Implementation Status**: Fully implemented in `internal/api/middleware/`

#### Recommendations

**Option A: Integrate into Admin Dashboard (High Priority)**
- **Priority**: High
- **Action**: Create admin/ops dashboard to expose these endpoints
- **Benefits**:
  - Critical for production operations
  - Enables proactive performance management
  - Supports troubleshooting and capacity planning
- **Implementation**:
  - Create `/admin` route in frontend
  - Build monitoring dashboard UI
  - Add authentication/authorization (admin-only)
  - Integrate with Grafana for visualization

**Option B: Expose via API Documentation**
- **Priority**: Medium
- **Action**: Document in OpenAPI spec and API docs
- **Benefits**: Allows external monitoring tools to integrate

**Option C: Remove (NOT Recommended)**
- **Risk**: High - These are critical operational endpoints
- **Action**: Only remove if alternative monitoring solution is in place

**Decision**: **KEEP and INTEGRATE** - These are essential for production operations

---

### Category 3: Session Management (3 endpoints)

#### Endpoints
- `GET /v1/sessions`
- `POST /v1/sessions`
- `DELETE /v1/sessions`

#### Purpose Analysis
Session management endpoints for:
- Creating and managing user sessions
- Tracking session activity and metrics
- Session lifecycle management

**Implementation Status**: Fully implemented in `internal/api/middleware/session_api.go`

#### Recommendations

**Option A: Integrate for Multi-User Support (High Priority)**
- **Priority**: High
- **Action**: Integrate session management into frontend
- **Use Cases**:
  - Multi-user dashboard support
  - Session persistence across devices
  - Activity tracking and audit logs
- **Implementation**:
  - Add session management UI
  - Integrate with authentication system
  - Add session switching capability

**Option B: Keep for Future Multi-Tenancy**
- **Priority**: Medium
- **Action**: Document as "Reserved for Multi-Tenancy Features"
- **Rationale**: Essential for future multi-user/multi-tenant support

**Decision**: **KEEP and PLAN INTEGRATION** - Critical for multi-user features

---

### Category 4: ML Model Management (3 endpoints)

#### Endpoints
- `/models/info` (GET)
- `/ml-models` (GET/POST)
- `/ensemble` (GET/POST)

#### Purpose Analysis
Machine learning model management:
- Model information and metadata
- Model versioning and deployment
- Ensemble model management

**Implementation Status**: Partially implemented in `services/risk-assessment-service/internal/handlers/`

#### Recommendations

**Option A: Integrate into Model Management UI (Medium Priority)**
- **Priority**: Medium
- **Action**: Create model management dashboard
- **Use Cases**:
  - View model performance metrics
  - Deploy/rollback model versions
  - Monitor model accuracy
- **Implementation**:
  - Create `/admin/models` route
  - Build model management interface
  - Add model performance visualization

**Option B: Keep for Future ML Ops**
- **Priority**: Low
- **Action**: Document as "ML Operations - Future Feature"
- **Rationale**: May be needed for advanced ML operations

**Decision**: **KEEP** - Useful for ML operations, integrate when ML features expand

---

### Category 5: External Data Integration (2 endpoints)

#### Endpoints
- `/thomson-reuters` (POST)
- `/supported` (GET) - List supported external APIs

#### Purpose Analysis
External data provider integration:
- Thomson Reuters data integration
- Supported external API listing
- Third-party data enrichment

**Implementation Status**: Implemented in `services/risk-assessment-service/internal/handlers/`

#### Recommendations

**Option A: Integrate for Data Enrichment (Medium Priority)**
- **Priority**: Medium
- **Action**: Add external data integration UI
- **Use Cases**:
  - Merchant data enrichment
  - Additional risk indicators
  - Compliance data verification
- **Implementation**:
  - Add "Enrich Data" button in merchant detail view
  - Show available external data sources
  - Display enrichment results

**Option B: Keep for Future Integration**
- **Priority**: Low
- **Action**: Document as "External Data Integration - Reserved"
- **Rationale**: May be needed for compliance or data enrichment features

**Decision**: **KEEP** - Valuable for data enrichment features

---

### Category 6: Reporting & Export (3 endpoints)

#### Endpoints
- `/reports` (GET/POST)
- `/reports/export` (GET/POST)
- `/export` (GET/POST)

#### Purpose Analysis
Report generation and data export:
- Custom report generation
- Data export in various formats
- Scheduled report delivery

**Implementation Status**: Partially implemented

#### Recommendations

**Option A: Integrate Export Functionality (High Priority)**
- **Priority**: High
- **Action**: Add export buttons to relevant views
- **Use Cases**:
  - Export merchant data
  - Export risk assessments
  - Generate compliance reports
- **Implementation**:
  - Add "Export" button to merchant list
  - Add "Export Report" to risk assessment view
  - Support CSV, PDF, JSON formats

**Option B: Keep for Future Reporting Features**
- **Priority**: Medium
- **Action**: Document as "Reporting Infrastructure - Future Feature"

**Decision**: **KEEP and INTEGRATE** - High user value for data export

---

### Category 7: Industry & Classification (2 endpoints)

#### Endpoints
- `/industry` (POST) - Industry experiment creation
- `/supported` (GET) - Supported features/industries

#### Purpose Analysis
Industry-specific features:
- Industry experiment management
- Supported industry listing
- Industry-specific configurations

**Implementation Status**: Implemented in `services/risk-assessment-service/internal/api/routes/`

#### Recommendations

**Option A: Integrate for Industry Management (Low Priority)**
- **Priority**: Low
- **Action**: Add industry management UI (if needed)
- **Use Cases**: Industry-specific risk models, configurations

**Option B: Remove if Not Needed**
- **Priority**: Low
- **Action**: Remove if industry experiments are not used
- **Rationale**: May be experimental feature not in production use

**Decision**: **EVALUATE USAGE** - Check if industry experiments are actively used

---

### Category 8: Testing & Development (5 endpoints)

#### Endpoints
- `/ping` (GET)
- `/live` (GET)
- `/filesystem` (GET/POST)
- `/memory` (GET)
- `/history` (GET)

#### Purpose Analysis
Development and testing utilities:
- Health checks and liveness probes
- File system operations (development)
- Memory inspection (debugging)
- History/audit trail access

**Implementation Status**: Mixed - some are health checks, some are dev tools

#### Recommendations

**Option A: Keep Health Checks, Remove Dev Tools**
- **Priority**: Medium
- **Action**: 
  - Keep `/ping` and `/live` (standard health checks)
  - Remove `/filesystem` and `/memory` (dev tools)
  - Keep `/history` if used for audit trails
- **Rationale**: 
  - Health checks are standard infrastructure
  - Dev tools should not be in production codebase

**Option B: Move Dev Tools to Separate Service**
- **Priority**: Low
- **Action**: Create separate dev-tools service
- **Rationale**: Keeps production code clean

**Decision**: **KEEP HEALTH CHECKS, REMOVE DEV TOOLS** - Clean separation of concerns

---

### Category 9: Advanced Features (7 endpoints)

#### Endpoints
- `/suites` (GET/POST)
- `/suites/{suiteId}/results` (GET)
- `/suites/{suiteId}/report` (GET)
- `/insights` (GET)
- `/queue` (GET/POST)
- `/events` (GET/POST)
- `/self-driving` (GET)

#### Purpose Analysis
Advanced/experimental features:
- Test suite management
- Analytics insights
- Queue management
- Event streaming
- Self-driving/automation features

**Implementation Status**: Mixed - some implemented, some experimental

#### Recommendations

**Option A: Evaluate Each Feature Individually**
- **Priority**: Varies by feature
- **Action**: Review each endpoint's implementation and usage
- **Analysis**:
  - `/suites` - Test suite management (likely for QA/testing)
  - `/insights` - Analytics insights (may be valuable)
  - `/queue` - Job queue management (operational)
  - `/events` - Event streaming (may be valuable)
  - `/self-driving` - Automation (experimental)

**Option B: Remove Experimental Features**
- **Priority**: Low
- **Action**: Remove `/self-driving` if not in active use
- **Rationale**: Experimental features add complexity

**Decision**: **EVALUATE INDIVIDUALLY** - Each feature needs separate assessment

---

### Category 10: User Management (1 endpoint)

#### Endpoints
- `/register` (POST)

#### Purpose Analysis
User registration endpoint:
- New user account creation
- User onboarding

**Implementation Status**: Likely implemented but not used in frontend

#### Recommendations

**Option A: Integrate Registration UI (High Priority)**
- **Priority**: High
- **Action**: Add user registration form to frontend
- **Use Cases**: 
  - New user signup
  - Account creation
  - User onboarding flow
- **Implementation**:
  - Create registration page
  - Add registration form
  - Integrate with authentication system

**Option B: Remove if Using External Auth**
- **Priority**: Low
- **Action**: Remove if using Supabase Auth or external provider
- **Rationale**: Redundant if external auth handles registration

**Decision**: **EVALUATE AUTH STRATEGY** - Keep if custom auth, remove if external

---

## Part 2: Unused Frontend Components Analysis

### Category 1: Utility Functions & Helpers (15 components)

#### Components
- `asyncGeneratorStep`, `ownKeys`, `executeModuleContent` - Internal JS utilities
- `getBrowserPrefix`, `supportsCSSFeature`, `waitForBrowserRendering` - Browser compatibility
- `logHeader`, `logSuccess`, `logError`, `logWarning`, `logInfo` - Logging utilities
- `createOutputDirectories`, `checkPrerequisites` - Test utilities
- `return` - False positive (keyword, not component)

#### Purpose Analysis
These are utility functions that may be:
- Used indirectly through other functions
- Part of build/test tooling
- Internal implementation details

#### Recommendations

**Option A: Keep - Likely Used Indirectly**
- **Priority**: Low
- **Action**: Verify actual usage with better static analysis
- **Rationale**: Many utilities are used indirectly

**Option B: Remove True Unused Utilities**
- **Priority**: Low
- **Action**: Remove only after confirming no indirect usage
- **Rationale**: Reduces codebase size

**Decision**: **KEEP FOR NOW** - Verify with better analysis tools

---

### Category 2: Service Getters & Factories (8 components)

#### Components
- `getEventBus`, `getAlertService`, `getExportService`
- `getMerchantDataService`, `getRiskDataService`
- `getChartLibrary`, `getRiskVisualizations`
- `loadSharedComponents`

#### Purpose Analysis
Service factory/getter functions:
- Dependency injection pattern
- Service initialization
- Shared component loading

**Implementation Status**: Implemented in shared component library

#### Recommendations

**Option A: Integrate Service Getters (High Priority)**
- **Priority**: High
- **Action**: Use these getters in frontend components
- **Benefits**:
  - Proper service initialization
  - Consistent service access
  - Better code organization
- **Implementation**:
  - Replace direct service instantiation with getters
  - Use `getEventBus()` for event handling
  - Use `getExportService()` for export functionality

**Option B: Keep for Future Integration**
- **Priority**: Medium
- **Action**: Document as "Service Factory Pattern - For Integration"

**Decision**: **INTEGRATE** - These are architectural patterns that should be used

---

### Category 3: UI Components & Templates (4 components)

#### Components
- `APIConfig` - API configuration class
- `RiskIndicatorsUITemplate` - Risk indicators UI template
- `WebsiteRiskDisplay` - Website risk display component
- `RiskIndicatorsHelpers` - Risk indicators helper functions

#### Purpose Analysis
UI components for risk assessment:
- Risk visualization
- UI templates
- Display components

**Implementation Status**: Implemented but not instantiated

#### Recommendations

**Option A: Integrate UI Components (High Priority)**
- **Priority**: High
- **Action**: Use these components in risk assessment views
- **Use Cases**:
  - Display risk indicators
  - Show website risk analysis
  - Render risk visualizations
- **Implementation**:
  - Add `RiskIndicatorsUITemplate` to merchant detail page
  - Use `WebsiteRiskDisplay` for website analysis
  - Integrate `RiskIndicatorsHelpers` for risk calculations

**Option B: Remove if Redundant**
- **Priority**: Low
- **Action**: Remove if replaced by newer components

**Decision**: **INTEGRATE** - These appear to be valuable UI components

---

### Category 4: Dashboard & Navigation (3 components)

#### Components
- `DashboardUtils` - Dashboard utility functions
- `refreshDashboard` - Dashboard refresh function
- `getCrossTabNavigation` - Cross-tab navigation

#### Purpose Analysis
Dashboard management:
- Dashboard utilities
- Refresh functionality
- Multi-tab navigation

#### Recommendations

**Option A: Integrate Dashboard Features (Medium Priority)**
- **Priority**: Medium
- **Action**: Use dashboard utilities in main dashboard
- **Use Cases**:
  - Auto-refresh dashboard data
  - Cross-tab synchronization
  - Dashboard state management

**Option B: Keep for Future Dashboard Enhancements**
- **Priority**: Low
- **Action**: Document as "Dashboard Utilities - For Integration"

**Decision**: **INTEGRATE** - Useful for dashboard functionality

---

### Category 5: Testing & Simulation (20 components)

#### Components
- `globalSetup`, `globalTeardown` - Test setup/teardown
- `simulateHover`, `simulateFocus`, `simulateClick`, `simulateTouch` - Interaction simulation
- `simulateKeyboardNavigation` - Keyboard navigation testing
- `waitForAnimation`, `waitForElementStable` - Async waiting utilities
- `captureTooltip`, `showTooltip`, `updateTooltip`, `hideTooltip` - Tooltip testing
- `setLoadingState`, `setErrorState`, `setEmptyState` - State simulation
- `simulateFormValidation`, `simulateNetworkDelay` - Form/network testing
- `clearAllStates`, `waitForPageStable` - State management
- `setViewportSize`, `navigateToDashboard` - Viewport/navigation testing
- `takeScreenshot`, `waitForCharts` - Visual testing

#### Purpose Analysis
Testing and simulation utilities:
- Browser automation testing
- UI interaction simulation
- Visual regression testing
- State management for testing

**Implementation Status**: Test utilities, likely in test files

#### Recommendations

**Option A: Keep - Test Utilities (Low Priority)**
- **Priority**: Low
- **Action**: Keep if used in test files
- **Rationale**: Test utilities are intentionally not used in production code

**Option B: Remove if Not in Test Files**
- **Priority**: Low
- **Action**: Remove if these are in production code (not test files)
- **Rationale**: Test utilities should be in test directories

**Decision**: **VERIFY LOCATION** - Keep if in test files, remove if in production code

---

### Category 6: Risk Assessment Components (8 components)

#### Components
- `setRiskState`, `waitForCharts` - Risk state management
- `showFeatureDetails`, `showBarTooltip`, `updateBarTooltip`, `hideBarTooltip` - Tooltip management
- `toggleWhyScorePanel` - Score explanation panel
- `showDataPointTooltip`, `hideDataPointTooltip` - Data point tooltips
- `dragstarted`, `dragged`, `dragended` - Drag and drop handlers

#### Purpose Analysis
Risk assessment UI interactions:
- Risk state management
- Interactive tooltips
- Score explanation
- Drag and drop functionality

#### Recommendations

**Option A: Integrate Risk Components (High Priority)**
- **Priority**: High
- **Action**: Use these components in risk assessment views
- **Use Cases**:
  - Interactive risk visualizations
  - Tooltip explanations
  - Score breakdown panels
  - Drag-and-drop risk configuration
- **Implementation**:
  - Add tooltip system to risk charts
  - Integrate "Why Score" panel
  - Add drag-and-drop for risk configuration

**Option B: Keep for Future Risk Features**
- **Priority**: Medium
- **Action**: Document as "Risk UI Components - For Integration"

**Decision**: **INTEGRATE** - These enhance risk assessment UX

---

## Part 3: Priority Matrix & Action Plan

### High Priority - Immediate Integration (12 endpoints, 15 components)

#### Endpoints
1. `/v1/memory/*` - System monitoring (admin dashboard)
2. `/v1/sessions/*` - Session management
3. `/reports/export` - Data export
4. `/export` - Export functionality
5. `/register` - User registration (if custom auth)

#### Components
1. Service getters (`getEventBus`, `getExportService`, etc.)
2. UI components (`RiskIndicatorsUITemplate`, `WebsiteRiskDisplay`)
3. Risk components (tooltips, score panels, drag-drop)

**Timeline**: 4-6 weeks  
**Effort**: Medium-High  
**Value**: High

---

### Medium Priority - Plan Integration (15 endpoints, 10 components)

#### Endpoints
1. `/models/info` - ML model management
2. `/thomson-reuters` - External data integration
3. `/supported` - Supported features listing
4. `/insights` - Analytics insights
5. `/queue` - Queue management
6. `/events` - Event streaming

#### Components
1. Dashboard utilities
2. Navigation components

**Timeline**: 8-12 weeks  
**Effort**: Medium  
**Value**: Medium-High

---

### Low Priority - Evaluate or Keep (10 endpoints, 47 components)

#### Endpoints
1. `/invites`, `/bugs` - Beta testing (keep if needed)
2. `/industry` - Industry experiments (evaluate usage)
3. `/ping`, `/live` - Health checks (keep)
4. `/filesystem`, `/memory` - Dev tools (remove)
5. `/suites` - Test suites (evaluate)
6. `/self-driving` - Experimental (remove if unused)

#### Components
1. Test utilities (keep if in test files)
2. Utility functions (verify indirect usage)

**Timeline**: As needed  
**Effort**: Low  
**Value**: Low-Medium

---

## Part 4: Risk Analysis

### High Risk - Do Not Remove
- `/v1/memory/*` - Critical for operations
- `/v1/sessions/*` - Needed for multi-user
- `/v1/thresholds` - Performance monitoring
- Service getters - Architectural patterns

### Medium Risk - Evaluate Before Removing
- `/models/info` - May be needed for ML ops
- `/thomson-reuters` - Data enrichment value
- UI components - May be used in HTML templates

### Low Risk - Safe to Remove
- `/filesystem` - Dev tool
- `/self-driving` - Experimental
- Test utilities in production code

---

## Part 5: Recommendations Summary

### Immediate Actions (Next 2 Weeks)
1. ✅ **Integrate Export Functionality** - High user value
2. ✅ **Create Admin Dashboard** - For system monitoring endpoints
3. ✅ **Integrate Service Getters** - Improve code architecture
4. ✅ **Add Risk UI Components** - Enhance risk assessment UX

### Short-term Actions (Next 1-2 Months)
1. ✅ **Integrate Session Management** - For multi-user support
2. ✅ **Add Model Management UI** - For ML operations
3. ✅ **Integrate External Data Sources** - For data enrichment

### Long-term Actions (3+ Months)
1. ⚠️ **Evaluate Beta Testing Features** - Keep if needed
2. ⚠️ **Review Experimental Features** - Remove if unused
3. ⚠️ **Clean Up Dev Tools** - Remove from production code

### Documentation Actions
1. ✅ **Document Reserved Endpoints** - Mark as "Future Use"
2. ✅ **Update API Documentation** - Include all endpoints
3. ✅ **Create Integration Guide** - For unused components

---

## Part 6: Metrics & Success Criteria

### Integration Success Metrics
- **Endpoint Usage**: Track API calls to newly integrated endpoints
- **Component Usage**: Monitor component instantiation
- **User Engagement**: Measure feature adoption
- **Error Rates**: Monitor integration-related errors

### Removal Success Criteria
- **No Breaking Changes**: Verify no dependencies before removal
- **Test Coverage**: Ensure tests pass after removal
- **Documentation Updated**: Update API docs and guides

---

## Conclusion

The analysis reveals:
- **37 unused endpoints** across 10 categories
- **72 unused components** across 6 categories
- **High-value features** ready for integration
- **Low-risk removal candidates** for code cleanup

**Recommended Approach**:
1. **Integrate high-priority features** (12 endpoints, 15 components)
2. **Plan medium-priority integration** (15 endpoints, 10 components)
3. **Evaluate low-priority features** (10 endpoints, 47 components)
4. **Remove only after careful evaluation** and dependency checking

**Estimated Impact**:
- **User Value**: High (export, monitoring, risk visualization)
- **Code Quality**: Medium (service patterns, architecture)
- **Maintenance**: Low (removal of unused code)

---

**Next Steps**:
1. Review this analysis with team
2. Prioritize integration tasks
3. Create tickets for high-priority items
4. Schedule evaluation for low-priority items
5. Update documentation with reserved endpoints

