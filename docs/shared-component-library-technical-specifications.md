# Shared Component Library - Technical Specifications
## Comprehensive Architecture & Implementation Guide

**Document Version**: 1.0  
**Date**: January 2025  
**Status**: Technical Specification  
**Target**: Unified Component Architecture for KYB Platform

---

## Executive Summary

This document provides detailed technical specifications for the Shared Component Library that will eliminate feature duplication across the KYB Platform. The library provides reusable data services, visualization components, UI components, and navigation utilities that can be leveraged by all pages while maintaining unique value propositions.

**Key Objectives**:
- Eliminate 40%+ code duplication
- Enable 60%+ component reuse
- Provide consistent user experience
- Reduce development time by 40%
- Simplify maintenance and updates

**Architecture Principles**:
1. **Separation of Concerns**: Data services, visualizations, and UI components are separate
2. **Dependency Injection**: Components accept dependencies for testability
3. **Event-Driven**: Components communicate via events, not direct coupling
4. **Progressive Enhancement**: Components work standalone or together
5. **Type Safety**: TypeScript definitions for all components

---

## 1. Architecture Overview

### 1.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Pages                         │
│  (Merchant Details, Compliance, Market Intel, etc.)         │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Shared Component Library                        │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Data Services│  │Visualizations│  │ UI Components│      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Navigation  │  │   Utilities  │  │   Events     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend APIs                             │
│  (Risk, Merchant, Compliance, Market Intelligence)          │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Component Library Structure

```
shared/
├── data-services/          # Data access layer
│   ├── risk-data-service.js
│   ├── merchant-data-service.js
│   ├── compliance-data-service.js
│   ├── market-data-service.js
│   └── analytics-data-service.js
├── visualizations/        # Chart and visualization components
│   ├── chart-library.js
│   ├── risk-visualizations.js
│   ├── benchmark-visualizations.js
│   └── trend-visualizations.js
├── components/            # Reusable UI components
│   ├── export-service.js
│   ├── alert-service.js
│   ├── recommendation-engine.js
│   ├── search-component.js
│   └── filter-component.js
├── navigation/            # Cross-page navigation
│   ├── cross-tab-navigation.js
│   ├── contextual-links.js
│   └── breadcrumb-navigation.js
├── utilities/             # Helper utilities
│   ├── formatters.js
│   ├── validators.js
│   ├── cache-manager.js
│   └── error-handler.js
├── events/                # Event system
│   ├── event-bus.js
│   └── event-types.js
└── types/                 # TypeScript definitions
    ├── data-types.d.ts
    ├── component-types.d.ts
    └── api-types.d.ts
```

---

## 2. Data Services Layer

### 2.1 Shared Risk Data Service

**Purpose**: Unified risk data access for all risk-related pages

**File**: `shared/data-services/risk-data-service.js`

```javascript
/**
 * Shared Risk Data Service
 * Provides unified access to risk data for all pages
 */
class SharedRiskDataService {
    constructor(config = {}) {
        this.apiConfig = config.apiConfig || APIConfig;
        this.cache = new Map();
        this.cacheTimeout = config.cacheTimeout || 5 * 60 * 1000; // 5 minutes
        this.eventBus = config.eventBus || EventBus.getInstance();
    }
    
    /**
     * Load comprehensive risk data for a merchant
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Loading options
     * @returns {Promise<RiskData>} Complete risk data
     */
    async loadRiskData(merchantId, options = {}) {
        const {
            includeHistory = false,
            includePredictions = false,
            includeBenchmarks = false,
            includeExplanations = false,
            includeRecommendations = false
        } = options;
        
        // Check cache
        const cacheKey = this.getCacheKey(merchantId, options);
        const cached = this.getCachedData(cacheKey);
        if (cached) return cached;
        
        // Load data in parallel
        const dataPromises = [
            this.loadCurrentRiskAssessment(merchantId, {
                includeExplanations,
                includeRecommendations
            })
        ];
        
        if (includeHistory) {
            dataPromises.push(this.loadRiskHistory(merchantId));
        }
        
        if (includePredictions) {
            dataPromises.push(this.loadRiskPredictions(merchantId));
        }
        
        if (includeBenchmarks) {
            dataPromises.push(this.loadIndustryBenchmarks(merchantId));
        }
        
        const results = await Promise.allSettled(dataPromises);
        
        // Combine results
        const riskData = this.combineRiskData(results, merchantId);
        
        // Cache result
        this.cacheData(cacheKey, riskData);
        
        // Emit event
        this.eventBus.emit('risk-data-loaded', {
            merchantId,
            riskData
        });
        
        return riskData;
    }
    
    /**
     * Load current risk assessment
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Assessment options
     * @returns {Promise<RiskAssessment>} Risk assessment data
     */
    async loadCurrentRiskAssessment(merchantId, options = {}) {
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.riskAssess, {
            method: 'POST',
            headers: this.apiConfig.getHeaders(),
            body: JSON.stringify({
                merchantId,
                includeTrendAnalysis: true,
                includeRecommendations: options.includeRecommendations || false,
                includeExplanations: options.includeExplanations || false
            })
        });
        
        if (!response.ok) {
            throw new Error(`Risk assessment failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load risk history
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Time range options
     * @returns {Promise<RiskHistory>} Risk history data
     */
    async loadRiskHistory(merchantId, options = {}) {
        const {
            timeRange = '6months',
            granularity = 'daily'
        } = options;
        
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(
            `${endpoints.riskHistory(merchantId)}?timeRange=${timeRange}&granularity=${granularity}`,
            {
                headers: this.apiConfig.getHeaders()
            }
        );
        
        if (!response.ok) {
            throw new Error(`Risk history failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load risk predictions
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Prediction options
     * @returns {Promise<RiskPredictions>} Risk prediction data
     */
    async loadRiskPredictions(merchantId, options = {}) {
        const {
            horizons = [3, 6, 12], // months
            includeScenarios = true,
            includeConfidence = true
        } = options;
        
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(
            `${endpoints.riskPredictions(merchantId)}?horizons=${horizons.join(',')}&includeScenarios=${includeScenarios}&includeConfidence=${includeConfidence}`,
            {
                headers: this.apiConfig.getHeaders()
            }
        );
        
        if (!response.ok) {
            throw new Error(`Risk predictions failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load industry benchmarks
     * @param {string} merchantId - Merchant ID
     * @param {Object} industryCodes - Industry classification codes
     * @returns {Promise<IndustryBenchmarks>} Industry benchmark data
     */
    async loadIndustryBenchmarks(merchantId, industryCodes = null) {
        // If industry codes not provided, fetch from Business Analytics
        if (!industryCodes) {
            industryCodes = await this.getIndustryCodesFromAnalytics(merchantId);
        }
        
        const endpoints = this.apiConfig.getEndpoints();
        const params = new URLSearchParams({
            mcc: industryCodes.mcc || '',
            naics: industryCodes.naics || '',
            sic: industryCodes.sic || ''
        });
        
        const response = await fetch(
            `${endpoints.riskBenchmarks}?${params.toString()}`,
            {
                headers: this.apiConfig.getHeaders()
            }
        );
        
        if (!response.ok) {
            throw new Error(`Industry benchmarks failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Get industry codes from Business Analytics
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<Object>} Industry codes
     */
    async getIndustryCodesFromAnalytics(merchantId) {
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.merchantById(merchantId), {
            headers: this.apiConfig.getHeaders()
        });
        
        if (!response.ok) {
            throw new Error(`Failed to fetch merchant data: ${response.status}`);
        }
        
        const merchantData = await response.json();
        
        return {
            mcc: merchantData.classification?.mcc_codes?.[0]?.code,
            naics: merchantData.classification?.naics_codes?.[0]?.code,
            sic: merchantData.classification?.sic_codes?.[0]?.code
        };
    }
    
    /**
     * Combine risk data from multiple sources
     * @param {Array} results - Promise results
     * @param {string} merchantId - Merchant ID
     * @returns {RiskData} Combined risk data
     */
    combineRiskData(results, merchantId) {
        const [assessment, history, predictions, benchmarks] = results;
        
        return {
            merchantId,
            current: assessment.status === 'fulfilled' ? assessment.value : null,
            history: history?.status === 'fulfilled' ? history.value : null,
            predictions: predictions?.status === 'fulfilled' ? predictions.value : null,
            benchmarks: benchmarks?.status === 'fulfilled' ? benchmarks.value : null,
            lastUpdated: new Date().toISOString(),
            dataSources: {
                assessment: assessment.status === 'fulfilled',
                history: history?.status === 'fulfilled',
                predictions: predictions?.status === 'fulfilled',
                benchmarks: benchmarks?.status === 'fulfilled'
            }
        };
    }
    
    /**
     * Clear cache for a merchant
     * @param {string} merchantId - Merchant ID
     */
    clearCache(merchantId) {
        const keysToDelete = [];
        for (const key of this.cache.keys()) {
            if (key.includes(merchantId)) {
                keysToDelete.push(key);
            }
        }
        keysToDelete.forEach(key => this.cache.delete(key));
        
        this.eventBus.emit('risk-data-cache-cleared', { merchantId });
    }
    
    /**
     * Get cache key
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Options
     * @returns {string} Cache key
     */
    getCacheKey(merchantId, options) {
        return `risk_${merchantId}_${JSON.stringify(options)}`;
    }
    
    /**
     * Get cached data
     * @param {string} key - Cache key
     * @returns {Object|null} Cached data or null
     */
    getCachedData(key) {
        const cached = this.cache.get(key);
        if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
            return cached.data;
        }
        if (cached) {
            this.cache.delete(key);
        }
        return null;
    }
    
    /**
     * Cache data
     * @param {string} key - Cache key
     * @param {Object} data - Data to cache
     */
    cacheData(key, data) {
        this.cache.set(key, {
            data,
            timestamp: Date.now()
        });
    }
}

// Export singleton instance
let riskDataServiceInstance = null;

export function getRiskDataService(config) {
    if (!riskDataServiceInstance) {
        riskDataServiceInstance = new SharedRiskDataService(config);
    }
    return riskDataServiceInstance;
}

export { SharedRiskDataService };
```

### 2.2 Shared Merchant Data Service

**Purpose**: Unified merchant data access

**File**: `shared/data-services/merchant-data-service.js`

```javascript
/**
 * Shared Merchant Data Service
 * Provides unified access to merchant data
 */
class SharedMerchantDataService {
    constructor(config = {}) {
        this.apiConfig = config.apiConfig || APIConfig;
        this.cache = new Map();
        this.cacheTimeout = config.cacheTimeout || 10 * 60 * 1000; // 10 minutes
        this.eventBus = config.eventBus || EventBus.getInstance();
    }
    
    /**
     * Load merchant data
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Loading options
     * @returns {Promise<MerchantData>} Merchant data
     */
    async loadMerchantData(merchantId, options = {}) {
        const {
            includeAnalytics = false,
            includeClassification = false,
            includeRisk = false
        } = options;
        
        // Check cache
        const cacheKey = `merchant_${merchantId}_${JSON.stringify(options)}`;
        const cached = this.getCachedData(cacheKey);
        if (cached) return cached;
        
        // Load base merchant data
        const merchant = await this.loadBaseMerchantData(merchantId);
        
        // Load additional data in parallel if requested
        const additionalPromises = [];
        
        if (includeAnalytics) {
            additionalPromises.push(this.loadMerchantAnalytics(merchantId));
        }
        
        if (includeClassification) {
            additionalPromises.push(this.loadMerchantClassification(merchantId));
        }
        
        if (includeRisk) {
            additionalPromises.push(this.loadMerchantRiskSummary(merchantId));
        }
        
        const additionalResults = await Promise.allSettled(additionalPromises);
        
        // Combine data
        const merchantData = {
            ...merchant,
            analytics: includeAnalytics && additionalResults[0]?.status === 'fulfilled' 
                ? additionalResults[0].value : null,
            classification: includeClassification && additionalResults[1]?.status === 'fulfilled'
                ? additionalResults[1].value : null,
            riskSummary: includeRisk && additionalResults[2]?.status === 'fulfilled'
                ? additionalResults[2].value : null
        };
        
        // Cache result
        this.cacheData(cacheKey, merchantData);
        
        // Emit event
        this.eventBus.emit('merchant-data-loaded', {
            merchantId,
            merchantData
        });
        
        return merchantData;
    }
    
    /**
     * Load base merchant data
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<Merchant>} Base merchant data
     */
    async loadBaseMerchantData(merchantId) {
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.merchantById(merchantId), {
            headers: this.apiConfig.getHeaders()
        });
        
        if (!response.ok) {
            throw new Error(`Failed to load merchant: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load merchant analytics
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<MerchantAnalytics>} Analytics data
     */
    async loadMerchantAnalytics(merchantId) {
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.merchantAnalytics(merchantId), {
            headers: this.apiConfig.getHeaders()
        });
        
        if (!response.ok) {
            throw new Error(`Failed to load analytics: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load merchant classification
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<MerchantClassification>} Classification data
     */
    async loadMerchantClassification(merchantId) {
        // Classification is part of merchant data, extract it
        const merchant = await this.loadBaseMerchantData(merchantId);
        return merchant.classification || null;
    }
    
    /**
     * Load merchant risk summary
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<RiskSummary>} Risk summary
     */
    async loadMerchantRiskSummary(merchantId) {
        const riskService = getRiskDataService();
        const riskData = await riskService.loadRiskData(merchantId, {
            includeHistory: false,
            includePredictions: false,
            includeBenchmarks: false
        });
        
        return {
            overallScore: riskData.current?.overallScore || 0,
            overallLevel: riskData.current?.overallLevel || 'unknown',
            categoryScores: riskData.current?.categoryScores || {}
        };
    }
    
    /**
     * Search merchants
     * @param {Object} searchParams - Search parameters
     * @returns {Promise<MerchantSearchResults>} Search results
     */
    async searchMerchants(searchParams) {
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.merchantSearch, {
            method: 'POST',
            headers: this.apiConfig.getHeaders(),
            body: JSON.stringify(searchParams)
        });
        
        if (!response.ok) {
            throw new Error(`Search failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    // Cache methods (similar to RiskDataService)
    getCachedData(key) { /* ... */ }
    cacheData(key, data) { /* ... */ }
    clearCache(merchantId) { /* ... */ }
}

export function getMerchantDataService(config) {
    if (!merchantDataServiceInstance) {
        merchantDataServiceInstance = new SharedMerchantDataService(config);
    }
    return merchantDataServiceInstance;
}

export { SharedMerchantDataService };
```

### 2.3 Shared Compliance Data Service

**Purpose**: Unified compliance data access

**File**: `shared/data-services/compliance-data-service.js`

```javascript
/**
 * Shared Compliance Data Service
 * Provides unified access to compliance data
 */
class SharedComplianceDataService {
    constructor(config = {}) {
        this.apiConfig = config.apiConfig || APIConfig;
        this.cache = new Map();
        this.cacheTimeout = config.cacheTimeout || 5 * 60 * 1000; // 5 minutes
        this.eventBus = config.eventBus || EventBus.getInstance();
    }
    
    /**
     * Load compliance status
     * @param {string} merchantId - Merchant ID (optional for portfolio view)
     * @param {Object} options - Loading options
     * @returns {Promise<ComplianceData>} Compliance data
     */
    async loadComplianceData(merchantId = null, options = {}) {
        const {
            includeGaps = false,
            includeProgress = false,
            includeAlerts = false,
            frameworks = [] // Empty = all frameworks
        } = options;
        
        // Check cache
        const cacheKey = `compliance_${merchantId || 'portfolio'}_${JSON.stringify(options)}`;
        const cached = this.getCachedData(cacheKey);
        if (cached) return cached;
        
        // Load compliance status
        const status = merchantId
            ? await this.loadMerchantComplianceStatus(merchantId, frameworks)
            : await this.loadPortfolioComplianceStatus(frameworks);
        
        // Load additional data in parallel if requested
        const additionalPromises = [];
        
        if (includeGaps) {
            additionalPromises.push(
                merchantId
                    ? this.loadMerchantComplianceGaps(merchantId, frameworks)
                    : this.loadPortfolioComplianceGaps(frameworks)
            );
        }
        
        if (includeProgress) {
            additionalPromises.push(
                merchantId
                    ? this.loadMerchantComplianceProgress(merchantId, frameworks)
                    : this.loadPortfolioComplianceProgress(frameworks)
            );
        }
        
        if (includeAlerts) {
            additionalPromises.push(
                merchantId
                    ? this.loadMerchantComplianceAlerts(merchantId)
                    : this.loadPortfolioComplianceAlerts()
            );
        }
        
        const additionalResults = await Promise.allSettled(additionalPromises);
        
        // Combine data
        const complianceData = {
            status,
            gaps: includeGaps && additionalResults[0]?.status === 'fulfilled'
                ? additionalResults[0].value : null,
            progress: includeProgress && additionalResults[1]?.status === 'fulfilled'
                ? additionalResults[1].value : null,
            alerts: includeAlerts && additionalResults[2]?.status === 'fulfilled'
                ? additionalResults[2].value : null,
            lastUpdated: new Date().toISOString()
        };
        
        // Cache result
        this.cacheData(cacheKey, complianceData);
        
        // Emit event
        this.eventBus.emit('compliance-data-loaded', {
            merchantId,
            complianceData
        });
        
        return complianceData;
    }
    
    /**
     * Load merchant compliance status
     * @param {string} merchantId - Merchant ID
     * @param {Array} frameworks - Frameworks to include
     * @returns {Promise<ComplianceStatus>} Compliance status
     */
    async loadMerchantComplianceStatus(merchantId, frameworks = []) {
        const endpoints = this.apiConfig.getEndpoints();
        const params = new URLSearchParams();
        if (frameworks.length > 0) {
            params.append('frameworks', frameworks.join(','));
        }
        
        const response = await fetch(
            `${endpoints.complianceStatus(merchantId)}?${params.toString()}`,
            {
                headers: this.apiConfig.getHeaders()
            }
        );
        
        if (!response.ok) {
            throw new Error(`Failed to load compliance status: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load portfolio compliance status
     * @param {Array} frameworks - Frameworks to include
     * @returns {Promise<PortfolioComplianceStatus>} Portfolio compliance status
     */
    async loadPortfolioComplianceStatus(frameworks = []) {
        const endpoints = this.apiConfig.getEndpoints();
        const params = new URLSearchParams();
        if (frameworks.length > 0) {
            params.append('frameworks', frameworks.join(','));
        }
        
        const response = await fetch(
            `${endpoints.complianceDashboard}?${params.toString()}`,
            {
                headers: this.apiConfig.getHeaders()
            }
        );
        
        if (!response.ok) {
            throw new Error(`Failed to load portfolio compliance: ${response.status}`);
        }
        
        return await response.json();
    }
    
    // Additional methods for gaps, progress, alerts...
    async loadMerchantComplianceGaps(merchantId, frameworks) { /* ... */ }
    async loadPortfolioComplianceGaps(frameworks) { /* ... */ }
    async loadMerchantComplianceProgress(merchantId, frameworks) { /* ... */ }
    async loadPortfolioComplianceProgress(frameworks) { /* ... */ }
    async loadMerchantComplianceAlerts(merchantId) { /* ... */ }
    async loadPortfolioComplianceAlerts() { /* ... */ }
    
    // Cache methods
    getCachedData(key) { /* ... */ }
    cacheData(key, data) { /* ... */ }
    clearCache(merchantId) { /* ... */ }
}

export function getComplianceDataService(config) {
    if (!complianceDataServiceInstance) {
        complianceDataServiceInstance = new SharedComplianceDataService(config);
    }
    return complianceDataServiceInstance;
}

export { SharedComplianceDataService };
```

---

## 3. Visualization Components

### 3.1 Shared Chart Library

**Purpose**: Reusable chart components using Chart.js

**File**: `shared/visualizations/chart-library.js`

```javascript
/**
 * Shared Chart Library
 * Provides reusable chart components
 */
class SharedChartLibrary {
    constructor() {
        this.charts = new Map();
        this.defaultConfig = {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    enabled: true,
                    mode: 'index',
                    intersect: false
                }
            }
        };
    }
    
    /**
     * Create a line chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} data - Chart data
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createLineChart(canvasId, data, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) {
            throw new Error(`Canvas element not found: ${canvasId}`);
        }
        
        // Destroy existing chart if present
        if (this.charts.has(canvasId)) {
            this.charts.get(canvasId).destroy();
        }
        
        const config = {
            type: 'line',
            data: this.normalizeChartData(data),
            options: {
                ...this.defaultConfig,
                ...options,
                scales: {
                    x: {
                        type: 'time',
                        time: {
                            unit: 'day'
                        },
                        ...options.scales?.x
                    },
                    y: {
                        beginAtZero: true,
                        ...options.scales?.y
                    }
                }
            }
        };
        
        const chart = new Chart(canvas, config);
        this.charts.set(canvasId, chart);
        
        return chart;
    }
    
    /**
     * Create a bar chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} data - Chart data
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createBarChart(canvasId, data, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) {
            throw new Error(`Canvas element not found: ${canvasId}`);
        }
        
        if (this.charts.has(canvasId)) {
            this.charts.get(canvasId).destroy();
        }
        
        const config = {
            type: 'bar',
            data: this.normalizeChartData(data),
            options: {
                ...this.defaultConfig,
                ...options,
                scales: {
                    x: {
                        ...options.scales?.x
                    },
                    y: {
                        beginAtZero: true,
                        ...options.scales?.y
                    }
                }
            }
        };
        
        const chart = new Chart(canvas, config);
        this.charts.set(canvasId, chart);
        
        return chart;
    }
    
    /**
     * Create a radar chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} data - Chart data
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createRadarChart(canvasId, data, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) {
            throw new Error(`Canvas element not found: ${canvasId}`);
        }
        
        if (this.charts.has(canvasId)) {
            this.charts.get(canvasId).destroy();
        }
        
        const config = {
            type: 'radar',
            data: this.normalizeChartData(data),
            options: {
                ...this.defaultConfig,
                ...options,
                scales: {
                    r: {
                        beginAtZero: true,
                        max: options.max || 100,
                        ...options.scales?.r
                    }
                }
            }
        };
        
        const chart = new Chart(canvas, config);
        this.charts.set(canvasId, chart);
        
        return chart;
    }
    
    /**
     * Create a gauge chart
     * @param {string} canvasId - Canvas element ID
     * @param {number} value - Gauge value (0-100)
     * @param {Object} options - Gauge options
     * @returns {Chart} Chart instance
     */
    createGaugeChart(canvasId, value, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) {
            throw new Error(`Canvas element not found: ${canvasId}`);
        }
        
        if (this.charts.has(canvasId)) {
            this.charts.get(canvasId).destroy();
        }
        
        // Gauge chart implementation using D3.js or custom drawing
        const ctx = canvas.getContext('2d');
        const centerX = canvas.width / 2;
        const centerY = canvas.height / 2;
        const radius = Math.min(canvas.width, canvas.height) / 2 - 20;
        
        // Draw gauge arc
        this.drawGaugeArc(ctx, centerX, centerY, radius, value, options);
        
        // Store reference
        const chart = { value, options, canvas };
        this.charts.set(canvasId, chart);
        
        return chart;
    }
    
    /**
     * Update chart data
     * @param {string} canvasId - Canvas element ID
     * @param {Object} newData - New chart data
     */
    updateChart(canvasId, newData) {
        const chart = this.charts.get(canvasId);
        if (!chart) {
            throw new Error(`Chart not found: ${canvasId}`);
        }
        
        if (chart.data) {
            chart.data = this.normalizeChartData(newData);
            chart.update();
        } else {
            // Recreate chart for non-standard charts
            this.recreateChart(canvasId, newData);
        }
    }
    
    /**
     * Destroy chart
     * @param {string} canvasId - Canvas element ID
     */
    destroyChart(canvasId) {
        const chart = this.charts.get(canvasId);
        if (chart && chart.destroy) {
            chart.destroy();
        }
        this.charts.delete(canvasId);
    }
    
    /**
     * Normalize chart data to Chart.js format
     * @param {Object} data - Raw chart data
     * @returns {Object} Normalized chart data
     */
    normalizeChartData(data) {
        if (data.labels && data.datasets) {
            return data; // Already normalized
        }
        
        // Convert from various formats to Chart.js format
        if (Array.isArray(data)) {
            return {
                labels: data.map(item => item.label || item.x),
                datasets: [{
                    label: 'Data',
                    data: data.map(item => item.value || item.y)
                }]
            };
        }
        
        return data;
    }
    
    /**
     * Draw gauge arc
     * @param {CanvasRenderingContext2D} ctx - Canvas context
     * @param {number} centerX - Center X coordinate
     * @param {number} centerY - Center Y coordinate
     * @param {number} radius - Gauge radius
     * @param {number} value - Gauge value (0-100)
     * @param {Object} options - Gauge options
     */
    drawGaugeArc(ctx, centerX, centerY, radius, value, options = {}) {
        const {
            min = 0,
            max = 100,
            color = this.getRiskColor(value),
            backgroundColor = '#e5e7eb',
            lineWidth = 20
        } = options;
        
        const normalizedValue = ((value - min) / (max - min)) * 100;
        const startAngle = Math.PI;
        const endAngle = startAngle + (normalizedValue / 100) * Math.PI;
        
        // Draw background arc
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, startAngle, startAngle + Math.PI);
        ctx.lineWidth = lineWidth;
        ctx.strokeStyle = backgroundColor;
        ctx.lineCap = 'round';
        ctx.stroke();
        
        // Draw value arc
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, startAngle, endAngle);
        ctx.lineWidth = lineWidth;
        ctx.strokeStyle = color;
        ctx.lineCap = 'round';
        ctx.stroke();
    }
    
    /**
     * Get risk color based on value
     * @param {number} value - Risk value (0-100)
     * @returns {string} Color hex code
     */
    getRiskColor(value) {
        if (value <= 25) return '#10b981'; // Green - Low
        if (value <= 50) return '#eab308'; // Yellow - Medium
        if (value <= 75) return '#f97316'; // Orange - High
        return '#ef4444'; // Red - Critical
    }
}

// Export singleton instance
let chartLibraryInstance = null;

export function getChartLibrary() {
    if (!chartLibraryInstance) {
        chartLibraryInstance = new SharedChartLibrary();
    }
    return chartLibraryInstance;
}

export { SharedChartLibrary };
```

### 3.2 Risk-Specific Visualizations

**Purpose**: Risk-specific chart components

**File**: `shared/visualizations/risk-visualizations.js`

```javascript
/**
 * Risk-Specific Visualizations
 * Specialized charts for risk data
 */
import { getChartLibrary } from './chart-library.js';

class RiskVisualizations {
    constructor() {
        this.chartLibrary = getChartLibrary();
    }
    
    /**
     * Create risk trend chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} riskHistory - Risk history data
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createRiskTrendChart(canvasId, riskHistory, options = {}) {
        const data = {
            labels: riskHistory.dataPoints.map(dp => dp.timestamp),
            datasets: [{
                label: 'Overall Risk Score',
                data: riskHistory.dataPoints.map(dp => dp.overallScore),
                borderColor: '#3b82f6',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                fill: true,
                tension: 0.4
            }, {
                label: 'Industry Average',
                data: riskHistory.dataPoints.map(dp => dp.industryAverage),
                borderColor: '#9ca3af',
                backgroundColor: 'rgba(156, 163, 175, 0.1)',
                borderDash: [5, 5],
                fill: false
            }]
        };
        
        return this.chartLibrary.createLineChart(canvasId, data, {
            ...options,
            plugins: {
                ...options.plugins,
                title: {
                    display: true,
                    text: 'Risk Trend Over Time'
                }
            }
        });
    }
    
    /**
     * Create risk radar chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} categoryScores - Risk category scores
     * @param {Object} benchmarks - Industry benchmarks (optional)
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createRiskRadarChart(canvasId, categoryScores, benchmarks = null, options = {}) {
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        const labels = categoryOrder.map(cat => this.formatCategoryName(cat));
        
        const datasets = [{
            label: 'Current Risk Level',
            data: categoryOrder.map(cat => categoryScores[cat]?.score || 0),
            backgroundColor: 'rgba(59, 130, 246, 0.2)',
            borderColor: 'rgba(59, 130, 246, 1)',
            borderWidth: 2
        }];
        
        if (benchmarks) {
            datasets.push({
                label: 'Industry Average',
                data: categoryOrder.map(cat => benchmarks.averages[cat] || 0),
                backgroundColor: 'rgba(156, 163, 175, 0.2)',
                borderColor: 'rgba(156, 163, 175, 1)',
                borderWidth: 2,
                borderDash: [5, 5]
            });
        }
        
        const data = {
            labels,
            datasets
        };
        
        return this.chartLibrary.createRadarChart(canvasId, data, {
            ...options,
            scales: {
                r: {
                    beginAtZero: true,
                    max: 100,
                    ...options.scales?.r
                }
            }
        });
    }
    
    /**
     * Create risk category bar chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} categoryScores - Risk category scores
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createRiskCategoryBarChart(canvasId, categoryScores, options = {}) {
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        const labels = categoryOrder.map(cat => this.formatCategoryName(cat));
        const data = categoryOrder.map(cat => categoryScores[cat]?.score || 0);
        const colors = categoryOrder.map(cat => this.getRiskColor(categoryScores[cat]?.score || 0));
        
        const chartData = {
            labels,
            datasets: [{
                label: 'Risk Score',
                data,
                backgroundColor: colors,
                borderColor: colors.map(c => this.darkenColor(c)),
                borderWidth: 1
            }]
        };
        
        return this.chartLibrary.createBarChart(canvasId, chartData, {
            ...options,
            indexAxis: 'y', // Horizontal bars
            plugins: {
                ...options.plugins,
                legend: {
                    display: false
                }
            }
        });
    }
    
    /**
     * Format category name
     * @param {string} category - Category key
     * @returns {string} Formatted category name
     */
    formatCategoryName(category) {
        const names = {
            financial: 'Financial',
            operational: 'Operational',
            regulatory: 'Regulatory',
            reputational: 'Reputational',
            cybersecurity: 'Cybersecurity',
            content: 'Content'
        };
        return names[category] || category;
    }
    
    /**
     * Get risk color
     * @param {number} score - Risk score (0-100)
     * @returns {string} Color hex code
     */
    getRiskColor(score) {
        if (score <= 25) return '#10b981'; // Green
        if (score <= 50) return '#eab308'; // Yellow
        if (score <= 75) return '#f97316'; // Orange
        return '#ef4444'; // Red
    }
    
    /**
     * Darken color
     * @param {string} color - Color hex code
     * @returns {string} Darkened color hex code
     */
    darkenColor(color) {
        // Simple color darkening (can be enhanced)
        return color;
    }
}

export function getRiskVisualizations() {
    return new RiskVisualizations();
}

export { RiskVisualizations };
```

---

## 4. UI Components

### 4.1 Shared Export Service

**Purpose**: Unified export functionality

**File**: `shared/components/export-service.js`

```javascript
/**
 * Shared Export Service
 * Provides unified export functionality across all pages
 */
class SharedExportService {
    constructor(config = {}) {
        this.apiConfig = config.apiConfig || APIConfig;
        this.eventBus = config.eventBus || EventBus.getInstance();
    }
    
    /**
     * Export data in specified format
     * @param {string} format - Export format (pdf, excel, csv, json)
     * @param {Object} data - Data to export
     * @param {Object} options - Export options
     * @returns {Promise<Blob>} Exported file blob
     */
    async exportData(format, data, options = {}) {
        const {
            template = 'default',
            filename = `export_${Date.now()}`,
            includeCharts = false,
            includeMetadata = true
        } = options;
        
        switch (format.toLowerCase()) {
            case 'pdf':
                return await this.exportToPDF(data, { template, filename, includeCharts, includeMetadata });
            case 'excel':
            case 'xlsx':
                return await this.exportToExcel(data, { template, filename, includeMetadata });
            case 'csv':
                return await this.exportToCSV(data, { filename, includeMetadata });
            case 'json':
                return await this.exportToJSON(data, { filename, includeMetadata });
            default:
                throw new Error(`Unsupported export format: ${format}`);
        }
    }
    
    /**
     * Export to PDF
     * @param {Object} data - Data to export
     * @param {Object} options - Export options
     * @returns {Promise<Blob>} PDF blob
     */
    async exportToPDF(data, options = {}) {
        const {
            template = 'default',
            filename,
            includeCharts = false,
            includeMetadata = true
        } = options;
        
        // Use jsPDF or similar library
        const { jsPDF } = await import('jspdf');
        const doc = new jsPDF();
        
        // Apply template
        const templateConfig = this.getTemplateConfig(template);
        this.applyPDFTemplate(doc, templateConfig);
        
        // Add content
        this.addPDFContent(doc, data, { includeCharts, includeMetadata });
        
        // Generate blob
        const pdfBlob = doc.output('blob');
        
        // Emit event
        this.eventBus.emit('export-completed', {
            format: 'pdf',
            filename,
            size: pdfBlob.size
        });
        
        return pdfBlob;
    }
    
    /**
     * Export to Excel
     * @param {Object} data - Data to export
     * @param {Object} options - Export options
     * @returns {Promise<Blob>} Excel blob
     */
    async exportToExcel(data, options = {}) {
        const {
            template = 'default',
            filename,
            includeMetadata = true
        } = options;
        
        // Use ExcelJS or similar library
        const ExcelJS = await import('exceljs');
        const workbook = new ExcelJS.Workbook();
        
        // Apply template
        const templateConfig = this.getTemplateConfig(template);
        this.applyExcelTemplate(workbook, templateConfig);
        
        // Add data
        this.addExcelData(workbook, data, { includeMetadata });
        
        // Generate blob
        const buffer = await workbook.xlsx.writeBuffer();
        const blob = new Blob([buffer], {
            type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
        });
        
        // Emit event
        this.eventBus.emit('export-completed', {
            format: 'excel',
            filename,
            size: blob.size
        });
        
        return blob;
    }
    
    /**
     * Export to CSV
     * @param {Object} data - Data to export
     * @param {Object} options - Export options
     * @returns {Promise<Blob>} CSV blob
     */
    async exportToCSV(data, options = {}) {
        const { filename, includeMetadata = true } = options;
        
        // Convert data to CSV
        const csv = this.convertToCSV(data, { includeMetadata });
        
        // Create blob
        const blob = new Blob([csv], { type: 'text/csv' });
        
        // Emit event
        this.eventBus.emit('export-completed', {
            format: 'csv',
            filename,
            size: blob.size
        });
        
        return blob;
    }
    
    /**
     * Export to JSON
     * @param {Object} data - Data to export
     * @param {Object} options - Export options
     * @returns {Promise<Blob>} JSON blob
     */
    async exportToJSON(data, options = {}) {
        const { filename, includeMetadata = true } = options;
        
        const exportData = includeMetadata
            ? {
                metadata: {
                    exportedAt: new Date().toISOString(),
                    version: '1.0',
                    source: 'KYB Platform'
                },
                data
            }
            : data;
        
        const json = JSON.stringify(exportData, null, 2);
        const blob = new Blob([json], { type: 'application/json' });
        
        // Emit event
        this.eventBus.emit('export-completed', {
            format: 'json',
            filename,
            size: blob.size
        });
        
        return blob;
    }
    
    /**
     * Download exported file
     * @param {Blob} blob - File blob
     * @param {string} filename - Filename
     */
    downloadFile(blob, filename) {
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
    }
    
    /**
     * Get template configuration
     * @param {string} templateName - Template name
     * @returns {Object} Template configuration
     */
    getTemplateConfig(templateName) {
        const templates = {
            default: {
                fontSize: 12,
                margin: { top: 20, right: 20, bottom: 20, left: 20 },
                header: { enabled: true, height: 30 },
                footer: { enabled: true, height: 20 }
            },
            riskReport: {
                fontSize: 11,
                margin: { top: 30, right: 20, bottom: 30, left: 20 },
                header: { enabled: true, height: 40, includeLogo: true },
                footer: { enabled: true, height: 25, includePageNumbers: true }
            },
            complianceReport: {
                fontSize: 10,
                margin: { top: 25, right: 20, bottom: 25, left: 20 },
                header: { enabled: true, height: 35 },
                footer: { enabled: true, height: 22, includeTimestamp: true }
            }
        };
        
        return templates[templateName] || templates.default;
    }
    
    // Template application methods
    applyPDFTemplate(doc, config) { /* ... */ }
    applyExcelTemplate(workbook, config) { /* ... */ }
    addPDFContent(doc, data, options) { /* ... */ }
    addExcelData(workbook, data, options) { /* ... */ }
    convertToCSV(data, options) { /* ... */ }
}

export function getExportService(config) {
    if (!exportServiceInstance) {
        exportServiceInstance = new SharedExportService(config);
    }
    return exportServiceInstance;
}

export { SharedExportService };
```

### 4.2 Shared Alert Service

**Purpose**: Unified alert management

**File**: `shared/components/alert-service.js`

```javascript
/**
 * Shared Alert Service
 * Provides unified alert management across all pages
 */
class SharedAlertService {
    constructor(config = {}) {
        this.eventBus = config.eventBus || EventBus.getInstance();
        this.alerts = new Map();
        this.alertRules = new Map();
    }
    
    /**
     * Create alert
     * @param {Object} alertData - Alert data
     * @returns {string} Alert ID
     */
    createAlert(alertData) {
        const {
            type, // 'risk', 'compliance', 'system', etc.
            severity, // 'critical', 'high', 'medium', 'low'
            title,
            message,
            merchantId = null,
            category = null,
            metadata = {}
        } = alertData;
        
        const alert = {
            id: this.generateAlertId(),
            type,
            severity,
            title,
            message,
            merchantId,
            category,
            metadata,
            status: 'active',
            createdAt: new Date().toISOString(),
            acknowledgedAt: null,
            resolvedAt: null
        };
        
        this.alerts.set(alert.id, alert);
        
        // Emit event
        this.eventBus.emit('alert-created', alert);
        
        // Check if alert should be displayed
        if (this.shouldDisplayAlert(alert)) {
            this.displayAlert(alert);
        }
        
        return alert.id;
    }
    
    /**
     * Acknowledge alert
     * @param {string} alertId - Alert ID
     * @param {string} userId - User ID
     */
    acknowledgeAlert(alertId, userId) {
        const alert = this.alerts.get(alertId);
        if (!alert) {
            throw new Error(`Alert not found: ${alertId}`);
        }
        
        alert.status = 'acknowledged';
        alert.acknowledgedAt = new Date().toISOString();
        alert.acknowledgedBy = userId;
        
        this.alerts.set(alertId, alert);
        
        // Emit event
        this.eventBus.emit('alert-acknowledged', alert);
    }
    
    /**
     * Resolve alert
     * @param {string} alertId - Alert ID
     * @param {string} userId - User ID
     * @param {string} resolution - Resolution notes
     */
    resolveAlert(alertId, userId, resolution = '') {
        const alert = this.alerts.get(alertId);
        if (!alert) {
            throw new Error(`Alert not found: ${alertId}`);
        }
        
        alert.status = 'resolved';
        alert.resolvedAt = new Date().toISOString();
        alert.resolvedBy = userId;
        alert.resolution = resolution;
        
        this.alerts.set(alertId, alert);
        
        // Emit event
        this.eventBus.emit('alert-resolved', alert);
    }
    
    /**
     * Get alerts
     * @param {Object} filters - Alert filters
     * @returns {Array} Filtered alerts
     */
    getAlerts(filters = {}) {
        const {
            type = null,
            severity = null,
            status = 'active',
            merchantId = null,
            category = null
        } = filters;
        
        let alerts = Array.from(this.alerts.values());
        
        if (type) {
            alerts = alerts.filter(a => a.type === type);
        }
        
        if (severity) {
            alerts = alerts.filter(a => a.severity === severity);
        }
        
        if (status) {
            alerts = alerts.filter(a => a.status === status);
        }
        
        if (merchantId) {
            alerts = alerts.filter(a => a.merchantId === merchantId);
        }
        
        if (category) {
            alerts = alerts.filter(a => a.category === category);
        }
        
        return alerts.sort((a, b) => {
            // Sort by severity, then by creation date
            const severityOrder = { critical: 0, high: 1, medium: 2, low: 3 };
            const severityDiff = severityOrder[a.severity] - severityOrder[b.severity];
            if (severityDiff !== 0) return severityDiff;
            return new Date(b.createdAt) - new Date(a.createdAt);
        });
    }
    
    /**
     * Display alert in UI
     * @param {Object} alert - Alert data
     */
    displayAlert(alert) {
        // Create alert notification
        const notification = this.createAlertNotification(alert);
        document.body.appendChild(notification);
        
        // Auto-remove after timeout (except critical)
        if (alert.severity !== 'critical') {
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 5000);
        }
    }
    
    /**
     * Create alert notification element
     * @param {Object} alert - Alert data
     * @returns {HTMLElement} Notification element
     */
    createAlertNotification(alert) {
        const notification = document.createElement('div');
        notification.className = `alert-notification alert-${alert.severity}`;
        notification.innerHTML = `
            <div class="alert-icon">
                <i class="fas fa-${this.getAlertIcon(alert.severity)}"></i>
            </div>
            <div class="alert-content">
                <div class="alert-title">${alert.title}</div>
                <div class="alert-message">${alert.message}</div>
            </div>
            <div class="alert-actions">
                <button class="alert-acknowledge" onclick="alertService.acknowledgeAlert('${alert.id}', 'current-user')">
                    <i class="fas fa-check"></i>
                </button>
                <button class="alert-dismiss" onclick="this.parentElement.parentElement.remove()">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;
        
        return notification;
    }
    
    /**
     * Get alert icon for severity
     * @param {string} severity - Alert severity
     * @returns {string} Icon class name
     */
    getAlertIcon(severity) {
        const icons = {
            critical: 'exclamation-circle',
            high: 'exclamation-triangle',
            medium: 'info-circle',
            low: 'info'
        };
        return icons[severity] || 'info';
    }
    
    /**
     * Check if alert should be displayed
     * @param {Object} alert - Alert data
     * @returns {boolean} Should display
     */
    shouldDisplayAlert(alert) {
        // Check alert rules
        const rules = this.alertRules.get(alert.type);
        if (rules && rules.shouldDisplay) {
            return rules.shouldDisplay(alert);
        }
        
        // Default: display all active alerts
        return alert.status === 'active';
    }
    
    /**
     * Generate alert ID
     * @returns {string} Alert ID
     */
    generateAlertId() {
        return `alert_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }
}

export function getAlertService(config) {
    if (!alertServiceInstance) {
        alertServiceInstance = new SharedAlertService(config);
    }
    return alertServiceInstance;
}

export { SharedAlertService };
```

---

## 5. Navigation Components

### 5.1 Cross-Tab Navigation

**Purpose**: Smart navigation between tabs and pages

**File**: `shared/navigation/cross-tab-navigation.js`

```javascript
/**
 * Cross-Tab Navigation
 * Provides smart navigation between tabs and pages
 */
class CrossTabNavigation {
    constructor(config = {}) {
        this.eventBus = config.eventBus || EventBus.getInstance();
        this.navigationHistory = [];
    }
    
    /**
     * Navigate to a specific tab/page
     * @param {Object} navigationTarget - Navigation target
     */
    navigateTo(navigationTarget) {
        const {
            page, // Page name (e.g., 'merchant-details')
            tab = null, // Tab name (e.g., 'risk-assessment')
            merchantId = null, // Merchant ID if applicable
            params = {} // Additional parameters
        } = navigationTarget;
        
        // Build URL
        const url = this.buildNavigationURL(page, tab, merchantId, params);
        
        // Add to history
        this.navigationHistory.push({
            page,
            tab,
            merchantId,
            timestamp: new Date().toISOString()
        });
        
        // Navigate
        if (tab && merchantId) {
            // Navigate to specific tab on merchant details page
            window.location.href = url;
            // After page load, activate the tab
            this.activateTabOnPageLoad(tab);
        } else {
            window.location.href = url;
        }
        
        // Emit event
        this.eventBus.emit('navigation-requested', {
            page,
            tab,
            merchantId,
            params
        });
    }
    
    /**
     * Get contextual links for current page
     * @param {string} currentPage - Current page name
     * @param {Object} contextData - Context data
     * @returns {Array} Array of contextual links
     */
    getContextualLinks(currentPage, contextData) {
        const links = [];
        
        if (currentPage === 'risk-indicators' && contextData.merchantId) {
            // Link to Risk Assessment for detailed analysis
            if (contextData.hasDetailedRisk) {
                links.push({
                    label: 'View Detailed Risk Analysis',
                    icon: 'chart-line',
                    action: () => this.navigateTo({
                        page: 'merchant-details',
                        tab: 'risk-assessment',
                        merchantId: contextData.merchantId
                    })
                });
            }
            
            // Link to Business Analytics for industry context
            if (contextData.industryCodes) {
                links.push({
                    label: 'See Industry Classification',
                    icon: 'industry',
                    action: () => this.navigateTo({
                        page: 'merchant-details',
                        tab: 'business-analytics',
                        merchantId: contextData.merchantId
                    })
                });
            }
            
            // Link to Compliance for regulatory risk details
            if (contextData.regulatoryRisk > 50) {
                links.push({
                    label: 'View Compliance Status',
                    icon: 'clipboard-check',
                    action: () => this.navigateTo({
                        page: 'compliance-dashboard',
                        merchantId: contextData.merchantId
                    })
                });
            }
        }
        
        // Add more contextual link logic for other pages...
        
        return links;
    }
    
    /**
     * Build navigation URL
     * @param {string} page - Page name
     * @param {string} tab - Tab name
     * @param {string} merchantId - Merchant ID
     * @param {Object} params - Additional parameters
     * @returns {string} Navigation URL
     */
    buildNavigationURL(page, tab, merchantId, params) {
        let url = `${page}.html`;
        
        if (merchantId) {
            url += `?merchantId=${merchantId}`;
        }
        
        if (tab) {
            url += merchantId ? `&tab=${tab}` : `?tab=${tab}`;
        }
        
        // Add additional parameters
        const paramEntries = Object.entries(params);
        if (paramEntries.length > 0) {
            const separator = url.includes('?') ? '&' : '?';
            url += separator + paramEntries.map(([key, value]) => `${key}=${encodeURIComponent(value)}`).join('&');
        }
        
        return url;
    }
    
    /**
     * Activate tab after page load
     * @param {string} tabName - Tab name
     */
    activateTabOnPageLoad(tabName) {
        // Wait for page to load
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => {
                this.activateTab(tabName);
            });
        } else {
            this.activateTab(tabName);
        }
    }
    
    /**
     * Activate tab
     * @param {string} tabName - Tab name
     */
    activateTab(tabName) {
        // Find tab button
        const tabButton = document.querySelector(`[data-tab="${tabName}"]`);
        if (tabButton) {
            tabButton.click();
        }
    }
    
    /**
     * Render contextual links in UI
     * @param {HTMLElement} container - Container element
     * @param {Array} links - Contextual links
     */
    renderContextualLinks(container, links) {
        if (!container || links.length === 0) return;
        
        container.innerHTML = `
            <div class="contextual-links">
                <h3 class="contextual-links-title">Related Information</h3>
                <div class="contextual-links-list">
                    ${links.map(link => `
                        <a href="#" class="contextual-link" onclick="event.preventDefault(); ${link.action.toString().replace(/function\s*\(\)\s*\{/, '').replace(/\}$/, '')}">
                            <i class="fas fa-${link.icon}"></i>
                            <span>${link.label}</span>
                            <i class="fas fa-arrow-right"></i>
                        </a>
                    `).join('')}
                </div>
            </div>
        `;
    }
}

export function getCrossTabNavigation(config) {
    if (!crossTabNavigationInstance) {
        crossTabNavigationInstance = new CrossTabNavigation(config);
    }
    return crossTabNavigationInstance;
}

export { CrossTabNavigation };
```

---

## 6. Event System

### 6.1 Event Bus

**Purpose**: Centralized event system for component communication

**File**: `shared/events/event-bus.js`

```javascript
/**
 * Event Bus
 * Centralized event system for component communication
 */
class EventBus {
    constructor() {
        this.listeners = new Map();
    }
    
    /**
     * Subscribe to event
     * @param {string} eventType - Event type
     * @param {Function} callback - Callback function
     * @returns {Function} Unsubscribe function
     */
    on(eventType, callback) {
        if (!this.listeners.has(eventType)) {
            this.listeners.set(eventType, []);
        }
        
        this.listeners.get(eventType).push(callback);
        
        // Return unsubscribe function
        return () => this.off(eventType, callback);
    }
    
    /**
     * Unsubscribe from event
     * @param {string} eventType - Event type
     * @param {Function} callback - Callback function
     */
    off(eventType, callback) {
        if (!this.listeners.has(eventType)) return;
        
        const callbacks = this.listeners.get(eventType);
        const index = callbacks.indexOf(callback);
        if (index > -1) {
            callbacks.splice(index, 1);
        }
    }
    
    /**
     * Emit event
     * @param {string} eventType - Event type
     * @param {*} data - Event data
     */
    emit(eventType, data) {
        if (!this.listeners.has(eventType)) return;
        
        const callbacks = this.listeners.get(eventType);
        callbacks.forEach(callback => {
            try {
                callback(data);
            } catch (error) {
                console.error(`Error in event listener for ${eventType}:`, error);
            }
        });
    }
    
    /**
     * Subscribe to event once
     * @param {string} eventType - Event type
     * @param {Function} callback - Callback function
     */
    once(eventType, callback) {
        const wrappedCallback = (data) => {
            callback(data);
            this.off(eventType, wrappedCallback);
        };
        this.on(eventType, wrappedCallback);
    }
}

// Export singleton instance
let eventBusInstance = null;

export function getEventBus() {
    if (!eventBusInstance) {
        eventBusInstance = new EventBus();
    }
    return eventBusInstance;
}

export { EventBus };
```

### 6.2 Event Types

**Purpose**: Standardized event type definitions

**File**: `shared/events/event-types.js`

```javascript
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
```

---

## 7. TypeScript Definitions

### 7.1 Data Types

**File**: `shared/types/data-types.d.ts`

```typescript
/**
 * Data Type Definitions
 */

export interface RiskData {
    merchantId: string;
    current: RiskAssessment | null;
    history: RiskHistory | null;
    predictions: RiskPredictions | null;
    benchmarks: IndustryBenchmarks | null;
    lastUpdated: string;
    dataSources: {
        assessment: boolean;
        history: boolean;
        predictions: boolean;
        benchmarks: boolean;
    };
}

export interface RiskAssessment {
    id: string;
    merchantId: string;
    overallScore: number;
    overallLevel: RiskLevel;
    categoryScores: Record<string, CategoryScore>;
    riskFactors: RiskFactor[];
    recommendations: Recommendation[];
    explanations?: SHAPExplanation;
    timestamp: string;
}

export interface RiskHistory {
    merchantId: string;
    timeRange: string;
    dataPoints: RiskHistoryDataPoint[];
    trends: RiskTrend[];
}

export interface RiskPredictions {
    merchantId: string;
    horizons: number[]; // months
    predictions: PredictionDataPoint[];
    scenarios: ScenarioData;
    confidence: ConfidenceInterval;
    drivers: RiskDriver[];
}

export interface IndustryBenchmarks {
    industryCodes: {
        mcc?: string;
        naics?: string;
        sic?: string;
    };
    industryName: string;
    averages: Record<string, number>;
    percentiles: {
        p10: Record<string, number>;
        p25: Record<string, number>;
        p50: Record<string, number>;
        p75: Record<string, number>;
        p90: Record<string, number>;
    };
    sampleSize: number;
}

export type RiskLevel = 'low' | 'medium' | 'high' | 'critical';

export interface CategoryScore {
    category: string;
    score: number;
    level: RiskLevel;
    subCategories: Record<string, number>;
    trend?: TrendDirection;
}

export interface TrendDirection {
    direction: 'improving' | 'stable' | 'rising';
    change: number;
    icon: string;
    label: string;
}

export interface MerchantData {
    id: string;
    name: string;
    // ... other merchant fields
    analytics?: MerchantAnalytics;
    classification?: MerchantClassification;
    riskSummary?: RiskSummary;
}

export interface ComplianceData {
    status: ComplianceStatus;
    gaps?: ComplianceGap[];
    progress?: ComplianceProgress;
    alerts?: ComplianceAlert[];
    lastUpdated: string;
}
```

---

## 8. Implementation Plan

### 8.1 Phase 1: Foundation (Weeks 1-2)

**Tasks**:
1. Create shared component library structure
2. Implement Event Bus
3. Implement Shared Risk Data Service
4. Implement Shared Merchant Data Service
5. Create basic TypeScript definitions

**Deliverables**:
- Shared component library folder structure
- Event Bus implementation
- Two core data services
- Basic type definitions

### 8.2 Phase 2: Visualizations (Weeks 3-4)

**Tasks**:
1. Implement Shared Chart Library
2. Implement Risk-Specific Visualizations
3. Create Benchmark Visualizations
4. Create Trend Visualizations

**Deliverables**:
- Reusable chart components
- Risk-specific chart wrappers
- Chart update/destroy methods

### 8.3 Phase 3: UI Components (Weeks 5-6)

**Tasks**:
1. Implement Shared Export Service
2. Implement Shared Alert Service
3. Implement Recommendation Engine
4. Implement Search/Filter Components

**Deliverables**:
- Export functionality (PDF, Excel, CSV, JSON)
- Alert management system
- Recommendation system
- Search/filter components

### 8.4 Phase 4: Navigation (Weeks 7-8)

**Tasks**:
1. Implement Cross-Tab Navigation
2. Implement Contextual Links
3. Create Navigation History
4. Integrate with existing navigation

**Deliverables**:
- Smart navigation system
- Contextual link generation
- Navigation history tracking

### 8.5 Phase 5: Integration (Weeks 9-10)

**Tasks**:
1. Integrate shared components into Risk Indicators tab
2. Integrate into Risk Assessment tab
3. Integrate into Compliance pages
4. Update Merchant Management pages
5. Testing and bug fixes

**Deliverables**:
- All pages using shared components
- No code duplication
- Comprehensive testing

---

## 9. Testing Strategy

### 9.1 Unit Testing

**Framework**: Jest

```javascript
// Example test for SharedRiskDataService
describe('SharedRiskDataService', () => {
    let service;
    let mockApiConfig;
    let mockEventBus;
    
    beforeEach(() => {
        mockApiConfig = {
            getEndpoints: jest.fn(),
            getHeaders: jest.fn()
        };
        mockEventBus = {
            emit: jest.fn()
        };
        service = new SharedRiskDataService({
            apiConfig: mockApiConfig,
            eventBus: mockEventBus
        });
    });
    
    describe('loadRiskData', () => {
        it('should load risk data with all options', async () => {
            // Test implementation
        });
        
        it('should use cache when available', async () => {
            // Test implementation
        });
        
        it('should emit event on data load', async () => {
            // Test implementation
        });
    });
});
```

### 9.2 Integration Testing

Test component interactions and data flow between shared components.

### 9.3 E2E Testing

Test complete user workflows using shared components.

---

## 10. Migration Guide

### 10.1 Migrating Risk Indicators Tab

**Before**:
```javascript
// Old: Direct API calls
const response = await fetch('/api/v1/risk/assess', {
    method: 'POST',
    body: JSON.stringify({ merchantId })
});
const riskData = await response.json();
```

**After**:
```javascript
// New: Using shared service
import { getRiskDataService } from '../shared/data-services/risk-data-service.js';

const riskService = getRiskDataService();
const riskData = await riskService.loadRiskData(merchantId, {
    includePredictions: true,
    includeBenchmarks: true
});
```

### 10.2 Migrating Charts

**Before**:
```javascript
// Old: Direct Chart.js usage
const chart = new Chart(canvas, {
    type: 'line',
    data: { /* ... */ }
});
```

**After**:
```javascript
// New: Using shared chart library
import { getChartLibrary } from '../shared/visualizations/chart-library.js';

const chartLibrary = getChartLibrary();
const chart = chartLibrary.createLineChart('canvasId', data, options);
```

---

## 11. Performance Considerations

### 11.1 Caching Strategy
- Data services implement intelligent caching
- Cache invalidation on data updates
- Configurable cache timeouts

### 11.2 Lazy Loading
- Components load on demand
- Charts render only when visible
- Data fetched only when needed

### 11.3 Bundle Optimization
- Tree-shaking for unused components
- Code splitting by feature
- Dynamic imports for heavy components

---

## 12. Documentation Requirements

### 12.1 Component Documentation
- JSDoc comments for all public methods
- Usage examples for each component
- API reference documentation

### 12.2 Integration Guides
- Step-by-step integration guides
- Migration guides from old implementations
- Best practices documentation

---

**Document Status**: Ready for Implementation  
**Next Steps**: Begin Phase 1 implementation  
**Owner**: Engineering Team

