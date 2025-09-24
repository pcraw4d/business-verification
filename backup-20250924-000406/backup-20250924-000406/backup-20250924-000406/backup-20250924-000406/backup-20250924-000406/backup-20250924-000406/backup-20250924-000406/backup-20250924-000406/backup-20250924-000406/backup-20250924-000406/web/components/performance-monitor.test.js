/**
 * Performance Monitor Unit Tests
 * Tests for performance monitoring, metrics collection, and budget management
 */

const { PerformanceMonitor, PerformanceBudgetManager } = require('./performance-monitor');

// Mock PerformanceObserver
global.PerformanceObserver = jest.fn().mockImplementation((callback) => ({
    observe: jest.fn(),
    disconnect: jest.fn(),
    takeRecords: jest.fn()
}));

// Mock performance API
global.performance = {
    now: jest.fn(() => Date.now()),
    mark: jest.fn(),
    measure: jest.fn(),
    getEntriesByType: jest.fn(() => []),
    getEntriesByName: jest.fn(() => [])
};

// Mock fetch
global.fetch = jest.fn();

describe('PerformanceMonitor', () => {
    let performanceMonitor;

    beforeEach(() => {
        // Reset mocks
        jest.clearAllMocks();
        
        // Mock console methods
        jest.spyOn(console, 'warn').mockImplementation();
        jest.spyOn(console, 'log').mockImplementation();
        
        // Mock process.env
        process.env.NODE_ENV = 'test';
        
        performanceMonitor = new PerformanceMonitor({
            enabled: true,
            sampleRate: 1.0,
            maxMetrics: 100,
            reportInterval: 1000
        });
    });

    afterEach(() => {
        if (performanceMonitor) {
            performanceMonitor.destroy();
        }
        jest.restoreAllMocks();
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            const monitor = new PerformanceMonitor();
            expect(monitor.options.enabled).toBe(true);
            expect(monitor.options.sampleRate).toBe(1.0);
            expect(monitor.options.maxMetrics).toBe(1000);
        });

        test('should initialize with custom options', () => {
            const customOptions = {
                enabled: false,
                sampleRate: 0.5,
                maxMetrics: 500
            };
            const monitor = new PerformanceMonitor(customOptions);
            expect(monitor.options.enabled).toBe(false);
            expect(monitor.options.sampleRate).toBe(0.5);
            expect(monitor.options.maxMetrics).toBe(500);
        });

        test('should not initialize when disabled', () => {
            const monitor = new PerformanceMonitor({ enabled: false });
            expect(monitor.isMonitoring).toBe(false);
        });
    });

    describe('Metric Recording', () => {
        test('should record metric correctly', () => {
            performanceMonitor.recordMetric('test-metric', 100, { type: 'test' });
            
            const metrics = performanceMonitor.getMetrics('test-metric');
            expect(metrics).toHaveLength(1);
            expect(metrics[0].name).toBe('test-metric');
            expect(metrics[0].value).toBe(100);
            expect(metrics[0].metadata.type).toBe('test');
        });

        test('should record custom metric', () => {
            performanceMonitor.recordCustomMetric('custom-metric', 200, { source: 'test' });
            
            const metrics = performanceMonitor.getMetrics('custom-metric');
            expect(metrics).toHaveLength(1);
            expect(metrics[0].value).toBe(200);
            expect(metrics[0].metadata.type).toBe('custom');
            expect(metrics[0].metadata.source).toBe('test');
        });

        test('should limit metrics to maxMetrics', () => {
            const maxMetrics = 5;
            const monitor = new PerformanceMonitor({ maxMetrics });
            
            // Add more metrics than the limit
            for (let i = 0; i < 10; i++) {
                monitor.recordMetric('test-metric', i);
            }
            
            const metrics = monitor.getMetrics('test-metric');
            expect(metrics).toHaveLength(maxMetrics);
            expect(metrics[0].value).toBe(5); // Should keep the last 5
        });

        test('should not record metrics when disabled', () => {
            const monitor = new PerformanceMonitor({ enabled: false });
            monitor.recordMetric('test-metric', 100);
            
            const metrics = monitor.getMetrics('test-metric');
            expect(metrics).toHaveLength(0);
        });
    });

    describe('Performance Budget Checking', () => {
        test('should check performance budget', () => {
            const eventHandler = jest.fn();
            window.addEventListener('performance:budgetExceeded', eventHandler);
            
            // Record metric that exceeds budget
            performanceMonitor.recordMetric('fcp', 2000); // Budget is 1800
            
            expect(eventHandler).toHaveBeenCalled();
            const eventDetail = eventHandler.mock.calls[0][0].detail;
            expect(eventDetail.metric).toBe('fcp');
            expect(eventDetail.value).toBe(2000);
            expect(eventDetail.budget).toBe(1800);
            
            window.removeEventListener('performance:budgetExceeded', eventHandler);
        });

        test('should determine budget severity correctly', () => {
            expect(performanceMonitor.getBudgetSeverity('fcp', 2000, 1800)).toBe('high');
            expect(performanceMonitor.getBudgetSeverity('fcp', 3000, 1800)).toBe('critical');
            expect(performanceMonitor.getBudgetSeverity('fcp', 1900, 1800)).toBe('low');
        });
    });

    describe('Performance Summary', () => {
        test('should generate performance summary', () => {
            // Add some test metrics
            performanceMonitor.recordMetric('test-metric', 100);
            performanceMonitor.recordMetric('test-metric', 200);
            performanceMonitor.recordMetric('test-metric', 300);
            
            const summary = performanceMonitor.getPerformanceSummary();
            
            expect(summary['test-metric']).toBeDefined();
            expect(summary['test-metric'].count).toBe(3);
            expect(summary['test-metric'].min).toBe(100);
            expect(summary['test-metric'].max).toBe(300);
            expect(summary['test-metric'].avg).toBe(200);
        });

        test('should calculate percentiles correctly', () => {
            const values = [10, 20, 30, 40, 50, 60, 70, 80, 90, 100];
            expect(performanceMonitor.percentile(values, 0.5)).toBe(50);
            expect(performanceMonitor.percentile(values, 0.75)).toBe(80);
            expect(performanceMonitor.percentile(values, 0.95)).toBe(100);
        });
    });

    describe('Budget Status', () => {
        test('should get budget status', () => {
            // Add metrics that exceed and meet budgets
            performanceMonitor.recordMetric('fcp', 2000); // Exceeds budget of 1800
            performanceMonitor.recordMetric('lcp', 2000); // Meets budget of 2500
            
            const status = performanceMonitor.getBudgetStatus();
            
            expect(status.fcp.status).toBe('exceeded');
            expect(status.fcp.ratio).toBeGreaterThan(1);
            expect(status.lcp.status).toBe('within-budget');
            expect(status.lcp.ratio).toBeLessThan(1);
        });
    });

    describe('Function Measurement', () => {
        test('should measure function execution time', async () => {
            const testFunction = jest.fn().mockResolvedValue('result');
            const measuredFunction = performanceMonitor.measureFunction('test-fn', testFunction);
            
            const result = await measuredFunction('arg1', 'arg2');
            
            expect(result).toBe('result');
            expect(testFunction).toHaveBeenCalledWith('arg1', 'arg2');
            
            const metrics = performanceMonitor.getMetrics('function-execution-time');
            expect(metrics).toHaveLength(1);
            expect(metrics[0].metadata.functionName).toBe('test-fn');
            expect(metrics[0].metadata.success).toBe(true);
        });

        test('should measure function execution time with error', async () => {
            const testFunction = jest.fn().mockRejectedValue(new Error('Test error'));
            const measuredFunction = performanceMonitor.measureFunction('test-fn', testFunction);
            
            await expect(measuredFunction()).rejects.toThrow('Test error');
            
            const metrics = performanceMonitor.getMetrics('function-execution-time');
            expect(metrics).toHaveLength(1);
            expect(metrics[0].metadata.functionName).toBe('test-fn');
            expect(metrics[0].metadata.success).toBe(false);
            expect(metrics[0].metadata.error).toBe('Test error');
        });
    });

    describe('Component Measurement', () => {
        test('should measure component render time', async () => {
            const renderFunction = jest.fn().mockResolvedValue('<div>Component</div>');
            const measuredRender = performanceMonitor.measureComponentRender('TestComponent', renderFunction);
            
            const result = await measuredRender();
            
            expect(result).toBe('<div>Component</div>');
            
            const metrics = performanceMonitor.getMetrics('component-render-time');
            expect(metrics).toHaveLength(1);
            expect(metrics[0].metadata.componentName).toBe('TestComponent');
            expect(metrics[0].metadata.success).toBe(true);
        });
    });

    describe('Resource Analysis', () => {
        test('should analyze resource type correctly', () => {
            expect(performanceMonitor.getResourceType('script.js')).toBe('script');
            expect(performanceMonitor.getResourceType('style.css')).toBe('stylesheet');
            expect(performanceMonitor.getResourceType('image.png')).toBe('image');
            expect(performanceMonitor.getResourceType('font.woff2')).toBe('font');
            expect(performanceMonitor.getResourceType('unknown.xyz')).toBe('other');
        });

        test('should analyze resource performance', () => {
            const mockEntry = {
                name: 'script.js',
                startTime: 0,
                responseEnd: 100,
                transferSize: 1024,
                decodedBodySize: 1024
            };
            
            performanceMonitor.analyzeResource(mockEntry);
            
            const metrics = performanceMonitor.getMetrics('resource-load-time');
            expect(metrics).toHaveLength(1);
            expect(metrics[0].value).toBe(100);
            expect(metrics[0].metadata.type).toBe('script');
        });
    });

    describe('Recommendations', () => {
        test('should generate optimization recommendations', () => {
            // Add metrics that exceed budgets
            performanceMonitor.recordMetric('fcp', 2000);
            performanceMonitor.recordMetric('bundleSize', 600000);
            
            const recommendations = performanceMonitor.generateRecommendations();
            
            expect(recommendations.length).toBeGreaterThan(0);
            expect(recommendations[0]).toHaveProperty('metric');
            expect(recommendations[0]).toHaveProperty('issue');
            expect(recommendations[0]).toHaveProperty('severity');
            expect(recommendations[0]).toHaveProperty('suggestion');
        });
    });

    describe('Data Export', () => {
        test('should export performance data', () => {
            performanceMonitor.recordMetric('test-metric', 100);
            
            const data = performanceMonitor.exportData();
            
            expect(data).toHaveProperty('metrics');
            expect(data).toHaveProperty('summary');
            expect(data).toHaveProperty('budgetStatus');
            expect(data).toHaveProperty('customMetrics');
            expect(data).toHaveProperty('timestamp');
        });
    });

    describe('Cleanup', () => {
        test('should clear metrics', () => {
            performanceMonitor.recordMetric('test-metric', 100);
            expect(performanceMonitor.getMetrics('test-metric')).toHaveLength(1);
            
            performanceMonitor.clearMetrics();
            expect(performanceMonitor.getMetrics('test-metric')).toHaveLength(0);
        });

        test('should stop monitoring', () => {
            expect(performanceMonitor.isMonitoring).toBe(true);
            
            performanceMonitor.stop();
            expect(performanceMonitor.isMonitoring).toBe(false);
        });

        test('should destroy monitor', () => {
            performanceMonitor.recordMetric('test-metric', 100);
            expect(performanceMonitor.getMetrics('test-metric')).toHaveLength(1);
            
            performanceMonitor.destroy();
            expect(performanceMonitor.getMetrics('test-metric')).toHaveLength(0);
            expect(performanceMonitor.isMonitoring).toBe(false);
        });
    });
});

describe('PerformanceBudgetManager', () => {
    let performanceMonitor;
    let budgetManager;

    beforeEach(() => {
        performanceMonitor = new PerformanceMonitor({ enabled: false });
        budgetManager = new PerformanceBudgetManager(performanceMonitor);
    });

    afterEach(() => {
        if (performanceMonitor) {
            performanceMonitor.destroy();
        }
    });

    describe('Budget Management', () => {
        test('should set performance budget', () => {
            budgetManager.setBudget('test-metric', 1000, {
                unit: 'ms',
                description: 'Test metric budget',
                priority: 'high'
            });
            
            const budget = budgetManager.budgets.get('test-metric');
            expect(budget.value).toBe(1000);
            expect(budget.unit).toBe('ms');
            expect(budget.description).toBe('Test metric budget');
            expect(budget.priority).toBe('high');
        });

        test('should set budget with default options', () => {
            budgetManager.setBudget('test-metric', 1000);
            
            const budget = budgetManager.budgets.get('test-metric');
            expect(budget.value).toBe(1000);
            expect(budget.unit).toBe('ms');
            expect(budget.description).toBe('');
            expect(budget.priority).toBe('medium');
        });
    });

    describe('Budget Checking', () => {
        test('should check budget compliance', () => {
            budgetManager.setBudget('test-metric', 1000);
            
            const result = budgetManager.checkBudget('test-metric', 800);
            expect(result.compliant).toBe(true);
            expect(result.severity).toBe('low');
            
            const result2 = budgetManager.checkBudget('test-metric', 1200);
            expect(result2.compliant).toBe(false);
            expect(result2.severity).toBe('high');
        });

        test('should return null for unknown metric', () => {
            const result = budgetManager.checkBudget('unknown-metric', 100);
            expect(result).toBeNull();
        });
    });

    describe('Severity Levels', () => {
        test('should determine severity correctly', () => {
            expect(budgetManager.getSeverity(0.5)).toBe('low');
            expect(budgetManager.getSeverity(0.9)).toBe('low');
            expect(budgetManager.getSeverity(1.0)).toBe('medium');
            expect(budgetManager.getSeverity(1.1)).toBe('medium');
            expect(budgetManager.getSeverity(1.3)).toBe('high');
            expect(budgetManager.getSeverity(1.6)).toBe('critical');
        });
    });

    describe('Alert Management', () => {
        test('should create alert when budget exceeded', () => {
            const eventHandler = jest.fn();
            window.addEventListener('performance:budgetAlert', eventHandler);
            
            budgetManager.setBudget('test-metric', 1000);
            budgetManager.checkBudget('test-metric', 1200);
            
            expect(eventHandler).toHaveBeenCalled();
            const alert = eventHandler.mock.calls[0][0].detail;
            expect(alert.metric).toBe('test-metric');
            expect(alert.value).toBe(1200);
            expect(alert.severity).toBe('high');
            
            window.removeEventListener('performance:budgetAlert', eventHandler);
        });

        test('should get active alerts', () => {
            budgetManager.setBudget('test-metric', 1000);
            budgetManager.checkBudget('test-metric', 1200);
            budgetManager.checkBudget('test-metric', 1500);
            
            const alerts = budgetManager.getAlerts();
            expect(alerts).toHaveLength(2);
            
            const highAlerts = budgetManager.getAlerts('high');
            expect(highAlerts).toHaveLength(1);
        });

        test('should clear alerts', () => {
            budgetManager.setBudget('test-metric', 1000);
            budgetManager.checkBudget('test-metric', 1200);
            expect(budgetManager.getAlerts()).toHaveLength(1);
            
            budgetManager.clearAlerts();
            expect(budgetManager.getAlerts()).toHaveLength(0);
        });

        test('should clear alerts for specific metric', () => {
            budgetManager.setBudget('metric1', 1000);
            budgetManager.setBudget('metric2', 1000);
            budgetManager.checkBudget('metric1', 1200);
            budgetManager.checkBudget('metric2', 1200);
            expect(budgetManager.getAlerts()).toHaveLength(2);
            
            budgetManager.clearAlerts('metric1');
            expect(budgetManager.getAlerts()).toHaveLength(1);
            expect(budgetManager.getAlerts()[0].metric).toBe('metric2');
        });
    });
});
