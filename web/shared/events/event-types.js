/**
 * Event Types
 * Standardized event type definitions
 */
export const EventTypes = {
    // Data events
    'risk-data-loaded': 'risk-data-loaded',
    'risk-data-updated': 'risk-data-updated',
    'risk-data-cache-cleared': 'risk-data-cache-cleared',
    'merchant-data-loaded': 'merchant-data-loaded',
    'merchant-data-updated': 'merchant-data-updated',
    'compliance-data-loaded': 'compliance-data-loaded',
    'compliance-data-updated': 'compliance-data-updated',
    
    // UI events
    'alert-created': 'alert-created',
    'alert-acknowledged': 'alert-acknowledged',
    'alert-resolved': 'alert-resolved',
    'export-completed': 'export-completed',
    'chart-updated': 'chart-updated',
    
    // Navigation events
    'navigation-requested': 'navigation-requested',
    'tab-activated': 'tab-activated',
    'page-loaded': 'page-loaded',
    
    // User interaction events
    'recommendation-clicked': 'recommendation-clicked',
    'filter-applied': 'filter-applied',
    'search-performed': 'search-performed'
};

