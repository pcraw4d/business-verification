# Real Data Integration Guide

## Overview

This guide explains how to integrate the new real-data components that replace mock data with live Supabase API calls. The integration provides comprehensive business intelligence, merchant management, and monitoring capabilities with real-time data.

## Components Overview

### 1. Real Data Integration Core (`web/components/real-data-integration.js`)

The core integration component that handles all API communication with the Supabase backend.

**Key Features:**
- Centralized API client for all data operations
- Automatic caching and cache management
- Error handling and retry logic
- Request/response logging
- Performance monitoring

**Usage:**
```javascript
const dataIntegration = new RealDataIntegration();

// Get merchant data
const merchants = await dataIntegration.getMerchants({
    page: 1,
    limit: 50,
    filters: { status: 'active' }
});

// Get analytics data
const analytics = await dataIntegration.getMerchantAnalytics();
```

### 2. Merchant Dashboard (`web/merchant-dashboard-real-data.js`)

Comprehensive merchant detail dashboard with real-time data.

**Key Features:**
- Real merchant information display
- Live analytics and statistics
- Activity timeline with real events
- Auto-refresh capabilities
- Export functionality

**Integration:**
```html
<!-- Replace the old merchant-dashboard.js script -->
<script src="components/real-data-integration.js"></script>
<script src="merchant-dashboard-real-data.js"></script>
```

### 3. Monitoring Dashboard (`web/monitoring-dashboard-real-data.js`)

System monitoring dashboard with live metrics and alerts.

**Key Features:**
- Real-time system metrics
- Live alert management
- Health check monitoring
- Performance charts
- Auto-refresh every 30 seconds

**Integration:**
```html
<!-- Replace the old monitoring-dashboard.js script -->
<script src="components/real-data-integration.js"></script>
<script src="monitoring-dashboard-real-data.js"></script>
```

### 4. Bulk Operations (`web/merchant-bulk-operations-real-data.js`)

Merchant bulk management with real data operations.

**Key Features:**
- Real merchant listing with pagination
- Bulk status updates
- Bulk export functionality
- Advanced filtering and search
- Real-time selection management

**Integration:**
```html
<!-- Replace the old merchant-bulk-operations.js script -->
<script src="components/real-data-integration.js"></script>
<script src="merchant-bulk-operations-real-data.js"></script>
```

### 5. Main Dashboard (`web/dashboard-real-data.js`)

Business intelligence dashboard with comprehensive real-time data.

**Key Features:**
- Real business intelligence metrics
- Live analytics and statistics
- Recent activity feed
- Performance charts
- Quick action buttons

**Integration:**
```html
<!-- Replace the old dashboard.html script -->
<script src="components/real-data-integration.js"></script>
<script src="dashboard-real-data.js"></script>
```

## Database Schema Requirements

Ensure your Supabase database has the following tables with the schema from `supabase-full-integration-migration.sql`:

### Core Tables
- `merchants` - Merchant information
- `classifications` - Business classifications
- `business_risk_assessments` - Risk assessment data
- `risk_keywords` - Risk keyword definitions
- `industry_code_crosswalks` - Industry code mappings
- `classification_performance_metrics` - Performance tracking

### Sample Data
The migration script includes sample data for:
- Risk keywords (gambling, adult, crypto, etc.)
- Industry code crosswalks (MCC, NAICS, SIC mappings)
- Sample business risk assessments
- Performance metrics

## API Endpoints

The real-data components expect the following API endpoints to be available:

### Merchant Management
- `GET /api/v1/merchants` - List merchants with pagination and filtering
- `GET /api/v1/merchants/{id}` - Get merchant details
- `PUT /api/v1/merchants/{id}` - Update merchant
- `DELETE /api/v1/merchants/{id}` - Delete merchant
- `POST /api/v1/merchants/bulk-update` - Bulk update merchants
- `POST /api/v1/merchants/bulk-export` - Export merchants
- `POST /api/v1/merchants/bulk-delete` - Bulk delete merchants

### Analytics and Statistics
- `GET /api/v1/merchants/analytics` - Get merchant analytics
- `GET /api/v1/merchants/statistics` - Get merchant statistics
- `GET /api/v1/business-intelligence` - Get business intelligence data

### Monitoring
- `GET /api/v1/monitoring/metrics` - Get system metrics
- `GET /api/v1/monitoring/alerts` - Get system alerts
- `GET /api/v1/monitoring/health-checks` - Get health checks
- `GET /api/v1/monitoring/performance` - Get performance metrics
- `POST /api/v1/monitoring/alerts/{id}/acknowledge` - Acknowledge alert

### Activity and Recent Data
- `GET /api/v1/recent-activity` - Get recent activity feed
- `POST /api/v1/bulk-verification` - Run bulk verification

## Environment Configuration

Ensure the following environment variables are set in your Railway deployment:

```bash
# Supabase Configuration
SUPABASE_URL=your_supabase_url
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
SUPABASE_JWT_SECRET=your_jwt_secret

# Feature Flags
ENABLE_REAL_DATA=true
ENABLE_MONITORING=true
ENABLE_ANALYTICS=true
```

## Migration Steps

### 1. Database Migration
```bash
# Run the full integration migration
psql -h your_supabase_host -U postgres -d postgres -f supabase-full-integration-migration.sql
```

### 2. Update HTML Files
Replace the old script references in your HTML files:

**Before:**
```html
<script src="merchant-dashboard.js"></script>
```

**After:**
```html
<script src="components/real-data-integration.js"></script>
<script src="merchant-dashboard-real-data.js"></script>
```

### 3. Update Webpack Configuration
Add the new components to your webpack config:

```javascript
// webpack.config.js
module.exports = {
  entry: {
    'merchant-dashboard': './merchant-dashboard-real-data.js',
    'monitoring-dashboard': './monitoring-dashboard-real-data.js',
    'merchant-bulk-operations': './merchant-bulk-operations-real-data.js',
    'dashboard': './dashboard-real-data.js',
    'real-data-integration': './components/real-data-integration.js'
  }
};
```

### 4. Test Integration
1. Verify database tables are created and populated
2. Test API endpoints are responding correctly
3. Check that UI components load real data
4. Verify auto-refresh functionality
5. Test export and bulk operations

## Performance Considerations

### Caching Strategy
- API responses are cached for 5 minutes by default
- Cache is automatically cleared on data mutations
- Manual cache clearing available via `clearCache()` method

### Auto-Refresh Intervals
- Dashboard: 5 minutes
- Monitoring: 30 seconds
- Merchant operations: 2 minutes
- Adjust intervals based on your needs

### Error Handling
- Automatic retry on network failures
- Graceful degradation to cached data
- User-friendly error messages
- Fallback to mock data if real data unavailable

## Troubleshooting

### Common Issues

1. **No data loading**
   - Check Supabase connection
   - Verify environment variables
   - Check browser console for errors
   - Ensure database tables exist

2. **API errors**
   - Verify API endpoints are implemented
   - Check CORS configuration
   - Validate request/response formats

3. **Performance issues**
   - Reduce auto-refresh intervals
   - Implement pagination for large datasets
   - Use caching effectively
   - Monitor database query performance

### Debug Mode
Enable debug mode by setting:
```javascript
window.DEBUG_MODE = true;
```

This will provide detailed logging of API calls and data flow.

## Security Considerations

### Row Level Security (RLS)
- All tables have RLS enabled
- Public read policies for non-sensitive data
- Service role key for admin operations
- JWT validation for user operations

### Data Validation
- Input validation on all API endpoints
- SQL injection prevention
- XSS protection in UI components
- CSRF protection for state-changing operations

## Monitoring and Observability

### Metrics Tracking
- API response times
- Error rates
- Cache hit rates
- User interaction metrics

### Logging
- Structured logging for all operations
- Error tracking and alerting
- Performance monitoring
- User activity logging

## Future Enhancements

### Planned Features
1. Real-time WebSocket updates
2. Advanced filtering and search
3. Custom dashboard widgets
4. Automated reporting
5. Machine learning insights

### Integration Opportunities
1. External data sources
2. Third-party analytics
3. Notification systems
4. Workflow automation
5. Compliance reporting

## Support and Maintenance

### Regular Maintenance
- Monitor database performance
- Update sample data periodically
- Review and optimize queries
- Update security policies
- Backup and recovery procedures

### Updates and Patches
- Regular dependency updates
- Security patches
- Performance optimizations
- Feature enhancements
- Bug fixes

## Conclusion

The real data integration provides a comprehensive solution for managing business intelligence, merchant operations, and system monitoring with live Supabase data. The modular architecture allows for easy maintenance and future enhancements while providing a robust, scalable foundation for the KYB platform.

For additional support or questions, refer to the individual component documentation or contact the development team.
