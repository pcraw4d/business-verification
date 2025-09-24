# Merchant-Centric UI Implementation Task List
## KYB Platform - Transform Dashboard-Centric to Merchant-Centric Architecture

**Document Version**: 1.0  
**Created**: January 2025  
**Status**: Planning Phase  
**Target**: Merchant-Centric UI with Portfolio Management

---

## Relevant Files

- `internal/api/handlers/merchant_portfolio_handler.go` - New API handler for merchant portfolio management
- `internal/api/handlers/merchant_portfolio_handler_test.go` - Unit tests for merchant portfolio handler
- `internal/services/merchant_portfolio_service.go` - Core business logic for merchant portfolio management
- `internal/services/merchant_portfolio_service_test.go` - Unit tests for merchant portfolio service
- `internal/database/merchant_portfolio_repository.go` - Data access layer for merchant portfolios
- `internal/database/merchant_portfolio_repository_test.go` - Unit tests for merchant portfolio repository
- `internal/database/migrations/005_merchant_portfolio_schema.sql` - Database schema for merchant portfolios
- `internal/database/migrations/006_mock_data_seed.sql` - Mock data seed for merchant portfolios
- `internal/database/migrations/006_mock_data_seed_test.go` - Unit tests for mock data seed
- `internal/database/migrations/007_foreign_key_relationships.sql` - Foreign key constraints and relationship validations
- `internal/database/migrations/007_foreign_key_relationships_test.go` - Unit tests for foreign key relationships
- `web/merchant-portfolio.html` - Main merchant portfolio dashboard
- `web/merchant-portfolio.js` - JavaScript for merchant portfolio functionality with real-time search, bulk operations, pagination, and export capabilities
- `web/merchant-portfolio.test.js` - Unit tests for merchant portfolio JavaScript functionality
- `web/merchant-detail.html` - Individual merchant holistic dashboard with comprehensive merchant information display
- `web/merchant-dashboard.html` - Individual merchant holistic dashboard
- `web/merchant-dashboard.js` - JavaScript for merchant dashboard functionality with real-time updates, data visualization, and comprehensive dashboard features
- `web/merchant-dashboard.test.js` - Unit tests for merchant dashboard JavaScript functionality
- `web/merchant-hub-integration.html` - Merchant hub integration interface with unified navigation and merchant context switching
- `web/merchant-hub-integration.test.js` - Frontend integration tests for merchant hub integration functionality
- `web/merchant-comparison.html` - 2-merchant comparison interface with side-by-side view and report generation capabilities
- `web/styles/merchant-comparison.css` - CSS styles for merchant comparison interface with responsive design and visual indicators
- `web/components/merchant-search.js` - Merchant search and filtering component with real-time search, debouncing, and comprehensive filtering
- `web/components/merchant-search.test.js` - Unit tests for merchant search component
- `web/components/portfolio-type-filter.js` - Portfolio type filtering component with visual indicators and multiple selection modes
- `web/components/portfolio-type-filter.test.js` - Unit tests for portfolio type filter component
- `web/components/risk-level-indicator.js` - Risk level visualization component with color coding, icons, and filtering capabilities
- `web/components/risk-level-indicator.test.js` - Unit tests for risk level indicator component
- `web/components/merchant-comparison.js` - Merchant comparison component with side-by-side comparison, exportable reports, and data visualization
- `web/components/merchant-comparison.test.js` - Unit tests for merchant comparison component
- `web/components/bulk-operations.js` - Bulk operations interface component
- `internal/api/routes/merchant_routes.go` - API routes for merchant management
- `internal/api/middleware/merchant_auth.go` - Authentication middleware for merchant operations
- `internal/modules/merchant_analytics/` - New module for merchant analytics
- `internal/modules/bulk_operations/` - New module for bulk merchant operations
- `internal/shared/models/merchant_models.go` - Shared merchant data models
- `internal/shared/interfaces/merchant_interfaces.go` - Shared merchant interfaces
- `internal/placeholders/placeholder_service.go` - Placeholder service for coming soon features
- `internal/placeholders/placeholder_service_test.go` - Unit tests for placeholder service
- `internal/api/handlers/placeholder_handler.go` - API handler for placeholder features
- `internal/api/handlers/placeholder_handler_test.go` - Unit tests for placeholder handler
- `internal/database/mock_merchant_database.go` - Mock merchant database for MVP testing
- `internal/database/mock_merchant_database_test.go` - Unit tests for mock merchant database
- `internal/models/merchant_portfolio.go` - Merchant portfolio data models
- `internal/models/merchant_portfolio_test.go` - Unit tests for merchant portfolio models
- `web/components/coming-soon-banner.js` - Coming soon banner component
- `web/components/mock-data-warning.js` - Mock data warning component
- `web/components/portfolio-type-filter.js` - Portfolio type filtering component
- `web/components/risk-level-indicator.js` - Risk level indicator component
- `web/components/session-manager.js` - Session management component with single merchant session management, state persistence, and session switching with overview reset
- `web/components/session-manager.test.js` - Unit tests for session manager component
- `web/components/bulk-progress-tracker.js` - Bulk operations progress tracker
- `web/components/merchant-comparison.js` - 2-merchant comparison component
- `web/components/audit-log-viewer.js` - Audit log viewer component
- `web/components/aml-compliance-tracker.js` - AML compliance tracking component
- `test/integration/merchant_portfolio_test.go` - Integration tests for merchant portfolio
- `test/e2e/merchant_workflow_test.go` - End-to-end tests for merchant workflows
- `test/mock_data/merchant_test_data.go` - Mock merchant data for testing
- `test/placeholders/placeholder_test.go` - Placeholder functionality tests
- `docs/user-guides/README.md` - User documentation index and overview
- `docs/user-guides/quick-start-guide.md` - Quick start guide for new users
- `docs/user-guides/merchant-portfolio-user-guide.md` - Comprehensive user guide for merchant portfolio management
- `docs/user-guides/merchant-ui-features.md` - Complete feature documentation for all merchant-centric UI components
- `docs/user-guides/troubleshooting-guide.md` - Troubleshooting guide for common user issues
- `docs/developer-guides/README.md` - Developer documentation index and overview
- `docs/developer-guides/architecture.md` - System architecture, design decisions, and technical overview
- `docs/developer-guides/api-development.md` - RESTful API development, endpoints, and best practices
- `docs/developer-guides/testing.md` - Comprehensive testing strategies, unit tests, integration tests, and E2E testing
- `docs/developer-guides/deployment.md` - Deployment procedures for different environments and platforms
- `docs/developer-guides/contributing.md` - Contribution guidelines, coding standards, and development workflow

### Notes

- Unit tests should be placed alongside the code files they are testing
- Integration tests should be in the `test/integration/` directory
- End-to-end tests should be in the `test/e2e/` directory
- Use `go test ./...` to run all tests, or `go test ./internal/services/` for specific packages
- Frontend tests should use the existing Playwright framework in `test/playwright/`
- Mock data should be realistic and cover various business types and scenarios
- Placeholders should clearly indicate "Coming Soon" status with expected timeline
- All merchant operations should include audit logging for compliance
- Bulk operations should handle 1000s of merchants with progress tracking
- Comparison should be limited to 2 merchants with exportable reports
- Session management should ensure only one merchant active at a time
- Hub integration should maintain backwards compatibility
- Portfolio types: onboarded, deactivated, prospective, pending
- Risk levels: high, medium, low
- MVP supports 20 concurrent users, scalable to 1000s
- AML compliance tracking following FATF recommendations

## Tasks

- [ ] 1.0 Backend Foundation - Merchant Portfolio Management APIs
- [ ] 2.0 Database Schema and Data Models
- [ ] 3.0 Placeholder System Implementation
- [ ] 4.0 Frontend Foundation - Merchant-Centric UI Components
- [x] 5.0 Unified Merchant Dashboard Implementation
- [ ] 6.0 Advanced Features - Bulk Operations and Comparison
- [ ] 7.0 Integration and Testing
- [ ] 8.0 Performance Optimization and Monitoring
- [x] 9.0 Documentation and Deployment

I have generated the high-level tasks based on the PRD and current codebase analysis, including a dedicated placeholder system for managing MVP functionality with mock data and coming soon features. Ready to generate the sub-tasks? Respond with 'Go' to proceed.

---

## Detailed Sub-Tasks

### 1.0 Backend Foundation - Merchant Portfolio Management APIs

#### 1.1 Core Merchant Portfolio Service
- [x] **1.1.1** Create `internal/services/merchant_portfolio_service.go` with core business logic
  - Implement merchant CRUD operations
  - Add portfolio type management (onboarded, deactivated, prospective, pending)
  - Add risk level assignment (high, medium, low)
  - Implement session management (single merchant active at a time)
  - Add audit logging for all merchant operations
  - **Testing**: Unit tests with 90%+ coverage
  - **Dependencies**: None

- [x] **1.1.2** Create `internal/models/merchant_portfolio.go` with data models
  - Define Merchant struct with all required fields
  - Define PortfolioType enum (onboarded, deactivated, prospective, pending)
  - Define RiskLevel enum (high, medium, low)
  - Define MerchantSession struct for session management
  - Define AuditLog struct for compliance tracking
  - **Testing**: Unit tests for all structs and validation
  - **Dependencies**: None

- [x] **1.1.3** Create `internal/database/merchant_portfolio_repository.go`
  - Implement database operations for merchant portfolio
  - Add pagination support for large merchant lists (1000s)
  - Implement search and filtering capabilities
  - Add bulk operation support
  - **Testing**: Integration tests with test database
  - **Dependencies**: 1.1.2

#### 1.2 Mock Merchant Database
- [x] **1.2.1** Create `internal/database/mock_merchant_database.go`
  - Generate 5000+ realistic mock merchants
  - Include diverse business types and industries
  - Add realistic business data (names, addresses, websites, etc.)
  - Implement portfolio type distribution
  - Implement risk level distribution
  - **Testing**: Unit tests for data generation
  - **Dependencies**: 1.1.2

- [x] **1.2.2** Create `test/mock_data/merchant_test_data.go`
  - Define test data sets for different scenarios
  - Include edge cases and boundary conditions
  - Add performance test data (large datasets)
  - **Testing**: Data validation tests
  - **Dependencies**: 1.2.1

#### 1.3 API Handlers
- [x] **1.3.1** Create `internal/api/handlers/merchant_portfolio_handler.go`
  - Implement REST API endpoints for merchant portfolio
  - Add merchant search and filtering endpoints
  - Implement bulk operations endpoints
  - Add session management endpoints
  - **Testing**: Unit tests with mocked dependencies
  - **Dependencies**: 1.1.1, 1.1.3

- [x] **1.3.2** Create `internal/api/routes/merchant_routes.go`
  - Define API routes for merchant portfolio
  - Add middleware for authentication and logging
  - Implement rate limiting for bulk operations
  - **Testing**: Integration tests for all routes
  - **Dependencies**: 1.3.1

#### 1.4 Compliance and Audit
- [x] **1.4.1** Create `internal/services/audit_service.go`
  - Implement audit logging for all merchant operations
  - Add AML compliance tracking
  - Implement FATF recommendation compliance
  - **Testing**: Unit tests for audit functionality
  - **Dependencies**: 1.1.2

- [x] **1.4.2** Create `internal/services/compliance_service.go`
  - Implement compliance checking for merchant operations
  - Add regulatory requirement validation
  - Implement compliance reporting
  - **Testing**: Unit tests for compliance logic
  - **Dependencies**: 1.4.1

### 2.0 Database Schema and Data Models

#### 2.1 Database Schema
- [x] **2.1.1** Create `internal/database/migrations/005_merchant_portfolio_schema.sql`
  - Define merchants table with all required fields
  - Add portfolio_types table
  - Add risk_levels table
  - Add merchant_sessions table
  - Add audit_logs table
  - Add compliance_records table
  - **Testing**: Migration tests
  - **Dependencies**: None

- [x] **2.1.2** Create `internal/database/migrations/006_mock_data_seed.sql`
  - Seed database with mock merchant data
  - Add portfolio type and risk level data
  - Include test data for development
  - **Testing**: Data validation tests
  - **Dependencies**: 2.1.1

#### 2.2 Data Relationships
- [x] **2.2.1** Implement foreign key relationships
  - Link merchants to users
  - Link merchants to portfolio types
  - Link merchants to risk levels
  - Link audit logs to merchants
  - **Testing**: Relationship validation tests
  - **Dependencies**: 2.1.1

- [x] **2.2.2** Add database indexes for performance
  - Index on merchant search fields
  - Index on portfolio type and risk level
  - Index on audit log timestamps
  - **Testing**: Performance tests
  - **Dependencies**: 2.2.1

### 3.0 Placeholder System Implementation

#### 3.1 Placeholder Service
- [x] **3.1.1** Create `internal/placeholders/placeholder_service.go`
  - Implement placeholder management for coming soon features
  - Add feature status tracking (coming soon, in development, available)
  - Implement mock data integration for testing
  - **Testing**: Unit tests for placeholder functionality
  - **Dependencies**: None

- [x] **3.1.2** Create `internal/placeholders/placeholder_config.go`
  - Define configuration for placeholder features
  - Add feature descriptions and timelines
  - Implement environment-specific configurations
  - **Testing**: Configuration validation tests
  - **Dependencies**: 3.1.1

#### 3.2 Placeholder API
- [x] **3.2.1** Create `internal/api/handlers/placeholder_handler.go`
  - Implement API endpoints for placeholder features
  - Add feature status endpoints
  - Implement mock data endpoints
  - **Testing**: Unit tests for placeholder API
  - **Dependencies**: 3.1.1

### 4.0 Frontend Foundation - Merchant-Centric UI Components

#### 4.1 Core UI Components
- [x] **4.1.1** Create `web/components/merchant-search.js`
  - Implement merchant search functionality
  - Add filtering by portfolio type and risk level
  - Implement real-time search with debouncing
  - **Testing**: Frontend unit tests
  - **Dependencies**: 1.3.1

- [x] **4.1.2** Create `web/components/portfolio-type-filter.js`
  - Implement portfolio type filtering
  - Add visual indicators for each type
  - Implement filter state management
  - **Testing**: Frontend unit tests
  - **Dependencies**: 4.1.1

- [x] **4.1.3** Create `web/components/risk-level-indicator.js`
  - Implement risk level visualization
  - Add color coding and icons
  - Implement risk level filtering
  - **Testing**: Frontend unit tests
  - **Dependencies**: 4.1.1

#### 4.2 Session Management
- [x] **4.2.1** Create `web/components/session-manager.js`
  - Implement single merchant session management
  - Add session state persistence
  - Implement session switching with overview reset
  - **Testing**: Frontend unit tests
  - **Dependencies**: 4.1.1

- [x] **4.2.2** Create `web/components/merchant-navigation.js`
  - Implement merchant navigation between different merchants
  - Add breadcrumb navigation
  - Implement quick merchant switching
  - **Testing**: Frontend unit tests
  - **Dependencies**: 4.2.1

#### 4.3 Placeholder Components
- [x] **4.3.1** Create `web/components/coming-soon-banner.js`
  - Implement coming soon feature indicators
  - Add feature descriptions and timelines
  - Implement mock data warnings
  - **Testing**: Frontend unit tests
  - **Dependencies**: 3.2.1

- [x] **4.3.2** Create `web/components/mock-data-warning.js`
  - Implement mock data warnings
  - Add clear indicators for test data
  - Implement data source information
  - **Testing**: Frontend unit tests
  - **Dependencies**: 4.3.1

### 5.0 Unified Merchant Dashboard Implementation

#### 5.1 Merchant Detail Dashboard
- [x] **5.1.1** Create `web/merchant-detail.html`
  - Implement holistic merchant view
  - Add all merchant information in single view
  - Implement responsive design
  - **Testing**: Frontend integration tests
  - **Dependencies**: 4.1.1, 4.2.1

- [x] **5.1.2** Create `web/merchant-dashboard.js`
  - Implement dashboard functionality
  - Add real-time data updates
  - Implement data visualization
  - **Testing**: Frontend unit tests
  - **Dependencies**: 5.1.1

#### 5.2 Portfolio Management
- [x] **5.2.1** Create `web/merchant-portfolio.html`
  - Implement merchant portfolio list view
  - Add pagination for large merchant lists
  - Implement search and filtering
  - **Testing**: Frontend integration tests
  - **Dependencies**: 4.1.1, 4.1.2

- [x] **5.2.2** Create `web/merchant-portfolio.js`
  - Implement portfolio management functionality
  - Add bulk selection capabilities
  - Implement portfolio type management
  - **Testing**: Frontend unit tests
  - **Dependencies**: 5.2.1

#### 5.3 Hub Integration
- [x] **5.3.1** Integrate with existing hub navigation
  - Add merchant portfolio to main navigation
  - Implement backwards compatibility
  - Add merchant context to existing dashboards
  - **Testing**: Integration tests
  - **Dependencies**: 5.1.1, 5.2.1

- [x] **5.3.2** Create `web/merchant-hub-integration.html`
  - Implement hub integration interface
  - Add merchant context switching
  - Implement unified navigation
  - **Testing**: Frontend integration tests
  - **Dependencies**: 5.3.1

### 6.0 Advanced Features - Bulk Operations and Comparison

#### 6.1 Bulk Operations
- [x] **6.1.1** Create `web/merchant-bulk-operations.html`
  - Implement bulk operations interface
  - Add progress tracking for large operations
  - Implement pause/resume functionality
  - **Testing**: Frontend integration tests
  - **Dependencies**: 5.2.1

- [x] **6.1.2** Create `web/components/bulk-progress-tracker.js`
  - Implement progress tracking component
  - Add real-time progress updates
  - Implement operation status management
  - **Testing**: Frontend unit tests
  - **Dependencies**: 6.1.1

- [x] **6.1.3** Create `internal/services/bulk_operations_service.go`
  - Implement bulk operations business logic
  - Add progress tracking and status management
  - Implement pause/resume functionality
  - **Testing**: Unit tests with mocked operations
  - **Dependencies**: 1.1.1

#### 6.2 Merchant Comparison
- [x] **6.2.1** Create `web/merchant-comparison.html`
  - Implement 2-merchant comparison interface
  - Add side-by-side comparison view
  - Implement comparison report generation
  - **Testing**: Frontend integration tests
  - **Dependencies**: 5.1.1

- [x] **6.2.2** Create `web/components/merchant-comparison.js`
  - Implement comparison functionality
  - Add exportable report generation
  - Implement comparison data visualization
  - **Testing**: Frontend unit tests
  - **Dependencies**: 6.2.1

- [x] **6.2.3** Create `internal/services/comparison_service.go`
  - Implement comparison business logic
  - Add report generation functionality
  - Implement export capabilities
  - **Testing**: Unit tests for comparison logic
  - **Dependencies**: 1.1.1

### 7.0 Integration and Testing

#### 7.1 Backend Testing
- [x] **7.1.1** Create comprehensive unit tests
  - Test all service layer functions
  - Test all repository functions
  - Test all API handlers
  - **Target**: 90%+ code coverage
  - **Dependencies**: All backend components

- [x] **7.1.2** Create integration tests
  - Test database operations
  - Test API endpoints
  - Test service integrations
  - **Target**: All critical paths covered
  - **Dependencies**: 7.1.1

- [x] **7.1.3** Create end-to-end tests
  - Test complete merchant workflows
  - Test bulk operations
  - Test comparison functionality
  - **Target**: All user journeys covered
  - **Dependencies**: 7.1.2

#### 7.2 Frontend Testing
- [x] **7.2.1** Create Playwright tests
  - Test merchant portfolio functionality
  - Test merchant detail views
  - Test bulk operations
  - **Target**: All UI interactions covered
  - **Dependencies**: All frontend components

- [x] **7.2.2** Create component tests
  - Test individual UI components
  - Test component interactions
  - Test responsive design
  - **Target**: All components tested
  - **Dependencies**: 7.2.1

#### 7.3 Performance Testing
- [x] **7.3.1** Create performance tests
  - Test with 1000s of merchants
  - Test bulk operations performance
  - Test concurrent user scenarios
  - **Target**: 20 concurrent users for MVP
  - **Dependencies**: 7.1.3, 7.2.2

### 8.0 Performance Optimization and Monitoring

#### 8.1 Performance Optimization
- [x] **8.1.1** Optimize database queries
  - Add proper indexing
  - Optimize pagination queries
  - Implement query caching
  - **Target**: Sub-second response times
  - **Dependencies**: 7.3.1

- [x] **8.1.2** Implement caching strategies
  - Add Redis caching for frequently accessed data
  - Implement cache invalidation
  - Add cache monitoring
  - **Target**: Reduced database load
  - **Dependencies**: 8.1.1

- [x] **8.1.3** Optimize frontend performance
  - Implement lazy loading
  - Add virtual scrolling for large lists
  - Optimize bundle size
  - **Target**: Fast page load times
  - **Dependencies**: 8.1.2

#### 8.2 Monitoring and Observability
- [x] **8.2.1** Implement application monitoring
  - Add performance metrics
  - Implement error tracking
  - Add user behavior analytics
  - **Target**: Full observability
  - **Dependencies**: 8.1.3

- [x] **8.2.2** Create monitoring dashboards
  - Add performance dashboards
  - Implement alerting
  - Add health checks
  - **Target**: Proactive monitoring
  - **Dependencies**: 8.2.1

### 9.0 Documentation and Deployment

#### 9.1 Documentation
- [x] **9.1.1** Create API documentation
  - Document all API endpoints
  - Add request/response examples
  - Create integration guides
  - **Target**: Complete API documentation
  - **Dependencies**: All API components

- [x] **9.1.2** Create user documentation
  - Create user guides
  - Add feature documentation
  - Create troubleshooting guides
  - **Target**: Complete user documentation
  - **Dependencies**: All frontend components

- [x] **9.1.3** Create developer documentation
  - Document architecture decisions
  - Add deployment guides
  - Create contribution guidelines
  - **Target**: Complete developer documentation
  - **Dependencies**: 9.1.1, 9.1.2

#### 9.2 Deployment
- [x] **9.2.1** Create deployment scripts
  - Create Docker configurations
  - Add deployment automation
  - Implement rollback capabilities
  - **Target**: Automated deployment
  - **Dependencies**: 9.1.3

- [x] **9.2.2** Create production configuration
  - Configure production environment
  - Add security configurations
  - Implement monitoring setup
  - **Target**: Production-ready deployment
  - **Dependencies**: 9.2.1

- [x] **9.2.3** Create rollback procedures
  - Implement rollback scripts
  - Add rollback testing
  - Create rollback documentation
  - **Target**: Safe rollback capabilities
  - **Dependencies**: 9.2.2

---

## Phase Reflection Tasks

### Phase 1 Reflection (After Backend Foundation)
- [x] **1.R.1** Review backend implementation
  - Assess code quality and architecture
  - Identify any gaps or improvements needed
  - Review test coverage and quality
  - **Deliverable**: Backend implementation review report

### Phase 2 Reflection (After Database Schema)
- [x] **2.R.1** Review database design
  - Assess schema design and relationships
  - Identify performance optimization opportunities
  - Review data integrity and constraints
  - **Deliverable**: Database design review report

### Phase 3 Reflection (After Placeholder System)
- [x] **3.R.1** Review placeholder implementation
  - Assess placeholder system effectiveness
  - Identify improvements for coming soon features
  - Review mock data quality and coverage
  - **Deliverable**: Placeholder system review report

### Phase 4 Reflection (After Frontend Foundation)
- [x] **4.R.1** Review frontend components
  - Assess component design and reusability
  - Identify UI/UX improvements
  - Review component testing coverage
  - **Deliverable**: Frontend component review report

### Phase 5 Reflection (After Unified Dashboard)
- [x] **5.R.1** Review dashboard implementation
  - Assess user experience and navigation
  - Identify performance improvements
  - Review integration with existing hub
  - **Deliverable**: Dashboard implementation review report

### Phase 6 Reflection (After Advanced Features)
- [x] **6.R.1** Review advanced features
  - Assess bulk operations performance
  - Identify comparison feature improvements
  - Review feature completeness
  - **Deliverable**: Advanced features review report

### Phase 7 Reflection (After Integration and Testing)
- [x] **7.R.1** Review testing implementation
  - Assess test coverage and quality
  - Identify testing gaps
  - Review test automation effectiveness
  - **Deliverable**: Testing implementation review report

### Phase 8 Reflection (After Performance Optimization)
- [x] **8.R.1** Review performance optimization
  - Assess performance improvements
  - Identify monitoring effectiveness
  - Review scalability readiness
  - **Deliverable**: Performance optimization review report

### Phase 9 Reflection (After Documentation and Deployment)
- [x] **9.R.1** Review documentation and deployment
  - Assess documentation completeness
  - Identify deployment process improvements
  - Review production readiness
  - **Deliverable**: Final implementation review report

---

## Success Criteria

### MVP Success Criteria
- [ ] **MVP.1** Support 20 concurrent users
- [ ] **MVP.2** Handle 1000s of merchants in portfolio
- [ ] **MVP.3** Single merchant session management
- [ ] **MVP.4** 2-merchant comparison functionality
- [ ] **MVP.5** Bulk operations with progress tracking
- [ ] **MVP.6** Portfolio type and risk level management
- [ ] **MVP.7** Mock data integration for testing
- [ ] **MVP.8** Coming soon feature placeholders
- [ ] **MVP.9** Hub integration with backwards compatibility
- [ ] **MVP.10** AML compliance tracking

### Post-MVP Success Criteria
- [ ] **POST.1** Scale to 1000s of concurrent users
- [ ] **POST.2** Real-time data updates
- [ ] **POST.3** Advanced security and access controls
- [ ] **POST.4** External API integrations
- [ ] **POST.5** Advanced analytics and reporting
- [ ] **POST.6** Multi-tenant support
- [ ] **POST.7** Advanced compliance features
- [ ] **POST.8** Performance optimization
- [ ] **POST.9** Advanced monitoring and alerting
- [ ] **POST.10** Production deployment with rollback
