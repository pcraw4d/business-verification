/**
 * Performance Monitor
 * Comprehensive performance monitoring for merchant UI components
 * Tracks metrics, performance budgets, and provides optimization insights
 */

class PerformanceMonitor {
    constructor(options = {}) {
        this.options = {
            enabled: true,
            sampleRate: 1.0,
            maxMetrics: 1000,
            reportInterval: 30000, // 30 seconds
            performanceBudget: {
                fcp: 1800, // First Contentful Paint
                lcp: 2500, // Largest Contentful Paint
                fid: 100,  // First Input Delay
                cls: 0.1,  // Cumulative Layout Shift
                ttfb: 600, // Time to First Byte
                bundleSize: 500000, // 500KB
                loadTime: 3000 // 3 seconds
            },
            ...options
        };

        this.metrics = new Map();
        this.observers = new Map();
        this.performanceEntries = [];
        this.budgets = this.options.performanceBudget;
        this.isMonitoring = false;
        this.reportTimer = null;
        this.customMetrics = new Map();

        this.init();
    }

    /**
     * Initialize performance monitoring
     */
    init() {
        if (!this.options.enabled) return;

        this.setupCoreWebVitals();
        this.setupResourceMonitoring();
        this.setupNavigationMonitoring();
        this.setupCustomMetrics();
        this.startReporting();
        this.isMonitoring = true;
    }

    /**
     * Setup Core Web Vitals monitoring
     */
    setupCoreWebVitals() {
        // First Contentful Paint (FCP)
        this.observePerformanceEntry('paint', (entries) => {
            entries.forEach(entry => {
                if (entry.name === 'first-contentful-paint') {
                    this.recordMetric('fcp', entry.startTime);
                }
            });
        });

        // Largest Contentful Paint (LCP)
        this.observePerformanceEntry('largest-contentful-paint', (entries) => {
            const lastEntry = entries[entries.length - 1];
            this.recordMetric('lcp', lastEntry.startTime);
        });

        // First Input Delay (FID)
        this.observePerformanceEntry('first-input', (entries) => {
            entries.forEach(entry => {
                this.recordMetric('fid', entry.processingStart - entry.startTime);
            });
        });

        // Cumulative Layout Shift (CLS)
        this.observePerformanceEntry('layout-shift', (entries) => {
            let clsValue = 0;
            entries.forEach(entry => {
                if (!entry.hadRecentInput) {
                    clsValue += entry.value;
                }
            });
            this.recordMetric('cls', clsValue);
        });
    }

    /**
     * Setup resource monitoring
     */
    setupResourceMonitoring() {
        this.observePerformanceEntry('resource', (entries) => {
            entries.forEach(entry => {
                this.analyzeResource(entry);
            });
        });
    }

    /**
     * Setup navigation monitoring
     */
    setupNavigationMonitoring() {
        this.observePerformanceEntry('navigation', (entries) => {
            entries.forEach(entry => {
                this.analyzeNavigation(entry);
            });
        });
    }

    /**
     * Setup custom metrics
     */
    setupCustomMetrics() {
        // Bundle loading metrics
        this.recordCustomMetric('bundle-load-start', performance.now());
        
        // Component render metrics
        this.setupComponentMetrics();
        
        // API call metrics
        this.setupAPIMetrics();
    }

    /**
     * Setup component-specific metrics
     */
    setupComponentMetrics() {
        // Monitor component render times
        const originalCreateElement = document.createElement;
        document.createElement = function(tagName) {
            const startTime = performance.now();
            const element = originalCreateElement.call(this, tagName);
            
            // Track element creation time
            setTimeout(() => {
                const endTime = performance.now();
                const renderTime = endTime - startTime;
                
                if (window.performanceMonitor) {
                    window.performanceMonitor.recordCustomMetric('element-render-time', renderTime, {
                        tagName,
                        elementId: element.id,
                        className: element.className
                    });
                }
            }, 0);
            
            return element;
        };
    }

    /**
     * Setup API call metrics
     */
    setupAPIMetrics() {
        const originalFetch = window.fetch;
        window.fetch = async function(url, options) {
            const startTime = performance.now();
            
            try {
                const response = await originalFetch.call(this, url, options);
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                if (window.performanceMonitor) {
                    window.performanceMonitor.recordCustomMetric('api-call-duration', duration, {
                        url: url.toString(),
                        method: options?.method || 'GET',
                        status: response.status,
                        statusText: response.statusText
                    });
                }
                
                return response;
            } catch (error) {
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                if (window.performanceMonitor) {
                    window.performanceMonitor.recordCustomMetric('api-call-error', duration, {
                        url: url.toString(),
                        method: options?.method || 'GET',
                        error: error.message
                    });
                }
                
                throw error;
            }
        };
    }

    /**
     * Observe performance entries
     * @param {string} type - Entry type
     * @param {Function} callback - Callback function
     */
    observePerformanceEntry(type, callback) {
        if (!('PerformanceObserver' in window)) return;

        try {
            const observer = new PerformanceObserver((list) => {
                const entries = list.getEntries();
                if (entries.length > 0) {
                    callback(entries);
                }
            });

            observer.observe({ entryTypes: [type] });
            this.observers.set(type, observer);
        } catch (error) {
            console.warn(`Failed to observe ${type} entries:`, error);
        }
    }

    /**
     * Record a performance metric
     * @param {string} name - Metric name
     * @param {number} value - Metric value
     * @param {Object} metadata - Additional metadata
     */
    recordMetric(name, value, metadata = {}) {
        if (!this.options.enabled) return;

        const metric = {
            name,
            value,
            timestamp: Date.now(),
            metadata
        };

        if (!this.metrics.has(name)) {
            this.metrics.set(name, []);
        }

        const metricArray = this.metrics.get(name);
        metricArray.push(metric);

        // Keep only the most recent metrics
        if (metricArray.length > this.options.maxMetrics) {
            metricArray.splice(0, metricArray.length - this.options.maxMetrics);
        }

        // Check performance budget
        this.checkPerformanceBudget(name, value);

        // Trigger metric recorded event
        this.triggerEvent('metricRecorded', { name, value, metadata });
    }

    /**
     * Record custom metric
     * @param {string} name - Metric name
     * @param {number} value - Metric value
     * @param {Object} metadata - Additional metadata
     */
    recordCustomMetric(name, value, metadata = {}) {
        this.recordMetric(name, value, { ...metadata, type: 'custom' });
    }

    /**
     * Analyze resource performance
     * @param {PerformanceResourceTiming} entry - Resource timing entry
     */
    analyzeResource(entry) {
        const resourceType = this.getResourceType(entry.name);
        const duration = entry.responseEnd - entry.startTime;
        const size = entry.transferSize || 0;

        this.recordCustomMetric('resource-load-time', duration, {
            type: resourceType,
            url: entry.name,
            size,
            cached: entry.transferSize === 0 && entry.decodedBodySize > 0
        });

        // Track bundle sizes
        if (resourceType === 'script' || resourceType === 'stylesheet') {
            this.recordCustomMetric('bundle-size', size, {
                type: resourceType,
                url: entry.name
            });
        }
    }

    /**
     * Analyze navigation performance
     * @param {PerformanceNavigationTiming} entry - Navigation timing entry
     */
    analyzeNavigation(entry) {
        const metrics = {
            'ttfb': entry.responseStart - entry.requestStart,
            'dom-content-loaded': entry.domContentLoadedEventEnd - entry.navigationStart,
            'load-complete': entry.loadEventEnd - entry.navigationStart,
            'dns-lookup': entry.domainLookupEnd - entry.domainLookupStart,
            'tcp-connect': entry.connectEnd - entry.connectStart,
            'ssl-negotiate': entry.secureConnectionStart > 0 ? entry.connectEnd - entry.secureConnectionStart : 0
        };

        Object.entries(metrics).forEach(([name, value]) => {
            if (value > 0) {
                this.recordMetric(name, value, { type: 'navigation' });
            }
        });
    }

    /**
     * Get resource type from URL
     * @param {string} url - Resource URL
     */
    getResourceType(url) {
        const extension = url.split('.').pop().toLowerCase();
        const typeMap = {
            'js': 'script',
            'css': 'stylesheet',
            'png': 'image',
            'jpg': 'image',
            'jpeg': 'image',
            'gif': 'image',
            'svg': 'image',
            'webp': 'image',
            'woff': 'font',
            'woff2': 'font',
            'ttf': 'font',
            'eot': 'font'
        };
        return typeMap[extension] || 'other';
    }

    /**
     * Check performance budget
     * @param {string} metricName - Metric name
     * @param {number} value - Metric value
     */
    checkPerformanceBudget(metricName, value) {
        const budget = this.budgets[metricName];
        if (budget && value > budget) {
            this.triggerEvent('budgetExceeded', {
                metric: metricName,
                value,
                budget,
                severity: this.getBudgetSeverity(metricName, value, budget)
            });
        }
    }

    /**
     * Get budget severity level
     * @param {string} metricName - Metric name
     * @param {number} value - Metric value
     * @param {number} budget - Budget limit
     */
    getBudgetSeverity(metricName, value, budget) {
        const ratio = value / budget;
        if (ratio > 2) return 'critical';
        if (ratio > 1.5) return 'high';
        if (ratio > 1.2) return 'medium';
        return 'low';
    }

    /**
     * Get performance metrics
     * @param {string} name - Metric name (optional)
     */
    getMetrics(name = null) {
        if (name) {
            return this.metrics.get(name) || [];
        }
        return Object.fromEntries(this.metrics);
    }

    /**
     * Get performance summary
     */
    getPerformanceSummary() {
        const summary = {};
        
        this.metrics.forEach((values, name) => {
            if (values.length === 0) return;
            
            const numericValues = values.map(v => v.value).filter(v => typeof v === 'number');
            if (numericValues.length === 0) return;
            
            summary[name] = {
                count: numericValues.length,
                min: Math.min(...numericValues),
                max: Math.max(...numericValues),
                avg: numericValues.reduce((a, b) => a + b, 0) / numericValues.length,
                p50: this.percentile(numericValues, 0.5),
                p75: this.percentile(numericValues, 0.75),
                p95: this.percentile(numericValues, 0.95),
                p99: this.percentile(numericValues, 0.99)
            };
        });
        
        return summary;
    }

    /**
     * Calculate percentile
     * @param {Array} values - Array of values
     * @param {number} percentile - Percentile (0-1)
     */
    percentile(values, percentile) {
        const sorted = [...values].sort((a, b) => a - b);
        const index = Math.ceil(sorted.length * percentile) - 1;
        return sorted[Math.max(0, index)];
    }

    /**
     * Get performance budget status
     */
    getBudgetStatus() {
        const status = {};
        const summary = this.getPerformanceSummary();
        
        Object.keys(this.budgets).forEach(metric => {
            const budget = this.budgets[metric];
            const metricSummary = summary[metric];
            
            if (metricSummary) {
                const avgValue = metricSummary.avg;
                const ratio = avgValue / budget;
                
                status[metric] = {
                    budget,
                    average: avgValue,
                    ratio,
                    status: ratio > 1 ? 'exceeded' : 'within-budget',
                    severity: this.getBudgetSeverity(metric, avgValue, budget)
                };
            }
        });
        
        return status;
    }

    /**
     * Start performance reporting
     */
    startReporting() {
        if (this.reportTimer) {
            clearInterval(this.reportTimer);
        }
        
        this.reportTimer = setInterval(() => {
            this.generateReport();
        }, this.options.reportInterval);
    }

    /**
     * Generate performance report
     */
    generateReport() {
        const report = {
            timestamp: Date.now(),
            summary: this.getPerformanceSummary(),
            budgetStatus: this.getBudgetStatus(),
            customMetrics: Object.fromEntries(this.customMetrics),
            recommendations: this.generateRecommendations()
        };
        
        this.triggerEvent('performanceReport', report);
        
        // Log to console in development
        if (process.env.NODE_ENV === 'development') {
            console.log('Performance Report:', report);
        }
    }

    /**
     * Generate performance recommendations
     */
    generateRecommendations() {
        const recommendations = [];
        const budgetStatus = this.getBudgetStatus();
        
        Object.entries(budgetStatus).forEach(([metric, status]) => {
            if (status.status === 'exceeded') {
                recommendations.push({
                    metric,
                    issue: `${metric} exceeds budget by ${Math.round((status.ratio - 1) * 100)}%`,
                    severity: status.severity,
                    suggestion: this.getOptimizationSuggestion(metric, status.ratio)
                });
            }
        });
        
        return recommendations;
    }

    /**
     * Get optimization suggestion for metric
     * @param {string} metric - Metric name
     * @param {number} ratio - Budget ratio
     */
    getOptimizationSuggestion(metric, ratio) {
        const suggestions = {
            fcp: 'Optimize critical rendering path, reduce render-blocking resources',
            lcp: 'Optimize largest contentful paint element, use image optimization',
            fid: 'Reduce JavaScript execution time, use code splitting',
            cls: 'Avoid layout shifts, reserve space for dynamic content',
            ttfb: 'Optimize server response time, use CDN',
            bundleSize: 'Implement code splitting, remove unused code',
            loadTime: 'Optimize resource loading, use lazy loading'
        };
        
        return suggestions[metric] || 'Consider performance optimization';
    }

    /**
     * Measure function execution time
     * @param {string} name - Function name
     * @param {Function} fn - Function to measure
     */
    measureFunction(name, fn) {
        return async (...args) => {
            const startTime = performance.now();
            
            try {
                const result = await fn(...args);
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                this.recordCustomMetric('function-execution-time', duration, {
                    functionName: name,
                    success: true
                });
                
                return result;
            } catch (error) {
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                this.recordCustomMetric('function-execution-time', duration, {
                    functionName: name,
                    success: false,
                    error: error.message
                });
                
                throw error;
            }
        };
    }

    /**
     * Measure component render time
     * @param {string} componentName - Component name
     * @param {Function} renderFn - Render function
     */
    measureComponentRender(componentName, renderFn) {
        return async (...args) => {
            const startTime = performance.now();
            
            try {
                const result = await renderFn(...args);
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                this.recordCustomMetric('component-render-time', duration, {
                    componentName,
                    success: true
                });
                
                return result;
            } catch (error) {
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                this.recordCustomMetric('component-render-time', duration, {
                    componentName,
                    success: false,
                    error: error.message
                });
                
                throw error;
            }
        };
    }

    /**
     * Trigger custom event
     * @param {string} eventName - Event name
     * @param {Object} detail - Event detail
     */
    triggerEvent(eventName, detail) {
        const event = new CustomEvent(`performance:${eventName}`, { detail });
        window.dispatchEvent(event);
    }

    /**
     * Export performance data
     */
    exportData() {
        return {
            metrics: this.getMetrics(),
            summary: this.getPerformanceSummary(),
            budgetStatus: this.getBudgetStatus(),
            customMetrics: Object.fromEntries(this.customMetrics),
            timestamp: Date.now()
        };
    }

    /**
     * Clear all metrics
     */
    clearMetrics() {
        this.metrics.clear();
        this.customMetrics.clear();
        this.performanceEntries = [];
    }

    /**
     * Stop monitoring
     */
    stop() {
        this.isMonitoring = false;
        
        if (this.reportTimer) {
            clearInterval(this.reportTimer);
            this.reportTimer = null;
        }
        
        this.observers.forEach(observer => observer.disconnect());
        this.observers.clear();
    }

    /**
     * Destroy performance monitor
     */
    destroy() {
        this.stop();
        this.clearMetrics();
    }
}

/**
 * Performance Budget Manager
 * Manages performance budgets and alerts
 */
class PerformanceBudgetManager {
    constructor(performanceMonitor) {
        this.performanceMonitor = performanceMonitor;
        this.budgets = new Map();
        this.alerts = [];
        this.alertThresholds = {
            low: 0.8,
            medium: 1.0,
            high: 1.2,
            critical: 1.5
        };
    }

    /**
     * Set performance budget
     * @param {string} metric - Metric name
     * @param {number} budget - Budget value
     * @param {Object} options - Budget options
     */
    setBudget(metric, budget, options = {}) {
        this.budgets.set(metric, {
            value: budget,
            unit: options.unit || 'ms',
            description: options.description || '',
            priority: options.priority || 'medium'
        });
    }

    /**
     * Check budget compliance
     * @param {string} metric - Metric name
     * @param {number} value - Current value
     */
    checkBudget(metric, value) {
        const budget = this.budgets.get(metric);
        if (!budget) return null;

        const ratio = value / budget.value;
        const severity = this.getSeverity(ratio);
        
        if (severity !== 'low') {
            this.createAlert(metric, value, budget, ratio, severity);
        }
        
        return {
            metric,
            value,
            budget: budget.value,
            ratio,
            severity,
            compliant: ratio <= 1.0
        };
    }

    /**
     * Get severity level
     * @param {number} ratio - Budget ratio
     */
    getSeverity(ratio) {
        if (ratio >= this.alertThresholds.critical) return 'critical';
        if (ratio >= this.alertThresholds.high) return 'high';
        if (ratio >= this.alertThresholds.medium) return 'medium';
        return 'low';
    }

    /**
     * Create performance alert
     * @param {string} metric - Metric name
     * @param {number} value - Current value
     * @param {Object} budget - Budget configuration
     * @param {number} ratio - Budget ratio
     * @param {string} severity - Severity level
     */
    createAlert(metric, value, budget, ratio, severity) {
        const alert = {
            id: `${metric}-${Date.now()}`,
            metric,
            value,
            budget: budget.value,
            ratio,
            severity,
            timestamp: Date.now(),
            description: budget.description,
            priority: budget.priority
        };
        
        this.alerts.push(alert);
        
        // Trigger alert event
        this.performanceMonitor.triggerEvent('budgetAlert', alert);
    }

    /**
     * Get active alerts
     * @param {string} severity - Filter by severity
     */
    getAlerts(severity = null) {
        if (severity) {
            return this.alerts.filter(alert => alert.severity === severity);
        }
        return this.alerts;
    }

    /**
     * Clear alerts
     * @param {string} metric - Filter by metric
     */
    clearAlerts(metric = null) {
        if (metric) {
            this.alerts = this.alerts.filter(alert => alert.metric !== metric);
        } else {
            this.alerts = [];
        }
    }
}

// Global performance monitor instance
if (typeof window !== 'undefined') {
    window.performanceMonitor = new PerformanceMonitor();
    window.performanceBudgetManager = new PerformanceBudgetManager(window.performanceMonitor);
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { PerformanceMonitor, PerformanceBudgetManager };
}
