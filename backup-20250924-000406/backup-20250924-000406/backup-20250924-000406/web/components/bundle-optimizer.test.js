/**
 * Bundle Optimizer Unit Tests
 * Tests for bundle optimization, code splitting, and module loading
 */

const { BundleOptimizer, DynamicImportManager } = require('./bundle-optimizer');

// Mock fetch for testing
global.fetch = jest.fn();

describe('BundleOptimizer', () => {
    let bundleOptimizer;

    beforeEach(() => {
        // Reset fetch mock
        fetch.mockClear();
        
        // Mock localStorage
        const localStorageMock = {
            getItem: jest.fn(),
            setItem: jest.fn(),
            removeItem: jest.fn(),
            clear: jest.fn(),
            length: 0,
            key: jest.fn()
        };
        Object.defineProperty(window, 'localStorage', {
            value: localStorageMock
        });

        bundleOptimizer = new BundleOptimizer({
            baseUrl: '/test/',
            cachePrefix: 'test-bundle-',
            cacheVersion: '1.0.0'
        });
    });

    afterEach(() => {
        if (bundleOptimizer) {
            bundleOptimizer.destroy();
        }
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            const optimizer = new BundleOptimizer();
            expect(optimizer.options.baseUrl).toBe('/');
            expect(optimizer.options.cachePrefix).toBe('kyb-bundle-');
            expect(optimizer.options.cacheVersion).toBe('1.0.0');
        });

        test('should initialize with custom options', () => {
            const customOptions = {
                baseUrl: '/custom/',
                cachePrefix: 'custom-',
                cacheVersion: '2.0.0'
            };
            const optimizer = new BundleOptimizer(customOptions);
            expect(optimizer.options.baseUrl).toBe('/custom/');
            expect(optimizer.options.cachePrefix).toBe('custom-');
            expect(optimizer.options.cacheVersion).toBe('2.0.0');
        });

        test('should register core modules', () => {
            expect(bundleOptimizer.modules.has('merchant-core')).toBe(true);
            expect(bundleOptimizer.modules.has('merchant-portfolio')).toBe(true);
            expect(bundleOptimizer.modules.has('merchant-dashboard')).toBe(true);
            expect(bundleOptimizer.modules.has('lazy-loader')).toBe(true);
            expect(bundleOptimizer.modules.has('virtual-scroller')).toBe(true);
        });
    });

    describe('Module Registration', () => {
        test('should register a new module', () => {
            bundleOptimizer.registerModule('test-module', {
                path: '/test-module.js',
                dependencies: ['merchant-core'],
                priority: 'high'
            });

            const module = bundleOptimizer.modules.get('test-module');
            expect(module).toBeDefined();
            expect(module.name).toBe('test-module');
            expect(module.path).toBe('/test-module.js');
            expect(module.dependencies).toEqual(['merchant-core']);
            expect(module.priority).toBe('high');
        });

        test('should register module with default values', () => {
            bundleOptimizer.registerModule('simple-module', {
                path: '/simple-module.js'
            });

            const module = bundleOptimizer.modules.get('simple-module');
            expect(module.dependencies).toEqual([]);
            expect(module.priority).toBe('medium');
            expect(module.size).toBe(0);
        });
    });

    describe('Module Loading', () => {
        test('should load module successfully', async () => {
            const mockModuleContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValueOnce({
                ok: true,
                text: () => Promise.resolve(mockModuleContent)
            });

            const result = await bundleOptimizer.loadModule('merchant-core');

            expect(fetch).toHaveBeenCalledWith('/test/merchant-core.js', expect.any(Object));
            expect(result).toBeDefined();
            expect(bundleOptimizer.loadedModules.has('merchant-core')).toBe(true);
        });

        test('should handle module loading error', async () => {
            fetch.mockRejectedValueOnce(new Error('Network error'));

            await expect(bundleOptimizer.loadModule('merchant-core')).rejects.toThrow('Network error');
        });

        test('should handle non-existent module', async () => {
            await expect(bundleOptimizer.loadModule('non-existent')).rejects.toThrow(
                "Module 'non-existent' not found"
            );
        });

        test('should load module with dependencies', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            await bundleOptimizer.loadModule('merchant-portfolio');

            expect(fetch).toHaveBeenCalledWith('/test/merchant-core.js', expect.any(Object));
            expect(fetch).toHaveBeenCalledWith('/test/merchant-portfolio.js', expect.any(Object));
        });

        test('should not load module twice', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValueOnce({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            await bundleOptimizer.loadModule('merchant-core');
            await bundleOptimizer.loadModule('merchant-core');

            expect(fetch).toHaveBeenCalledTimes(1);
        });

        test('should handle concurrent module loading', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            const promises = [
                bundleOptimizer.loadModule('merchant-core'),
                bundleOptimizer.loadModule('merchant-core'),
                bundleOptimizer.loadModule('merchant-core')
            ];

            await Promise.all(promises);

            expect(fetch).toHaveBeenCalledTimes(1);
        });
    });

    describe('Preloading', () => {
        test('should preload high priority modules', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            await bundleOptimizer.preloadModules(['high']);

            expect(bundleOptimizer.loadedModules.has('merchant-core')).toBe(true);
            expect(bundleOptimizer.loadedModules.has('lazy-loader')).toBe(true);
        });

        test('should preload multiple priority levels', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            await bundleOptimizer.preloadModules(['high', 'medium']);

            expect(bundleOptimizer.loadedModules.has('merchant-core')).toBe(true);
            expect(bundleOptimizer.loadedModules.has('merchant-search')).toBe(true);
        });
    });

    describe('Route Loading', () => {
        test('should load modules for merchant-portfolio route', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            await bundleOptimizer.loadRouteModules('merchant-portfolio');

            const expectedModules = ['merchant-core', 'merchant-portfolio', 'merchant-search', 'lazy-loader', 'virtual-scroller'];
            expectedModules.forEach(moduleName => {
                expect(bundleOptimizer.loadedModules.has(moduleName)).toBe(true);
            });
        });

        test('should load modules for merchant-dashboard route', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            await bundleOptimizer.loadRouteModules('merchant-dashboard');

            const expectedModules = ['merchant-core', 'merchant-dashboard', 'lazy-loader'];
            expectedModules.forEach(moduleName => {
                expect(bundleOptimizer.loadedModules.has(moduleName)).toBe(true);
            });
        });

        test('should load default modules for unknown route', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            await bundleOptimizer.loadRouteModules('unknown-route');

            const expectedModules = ['merchant-core', 'utils', 'api-client'];
            expectedModules.forEach(moduleName => {
                expect(bundleOptimizer.loadedModules.has(moduleName)).toBe(true);
            });
        });
    });

    describe('Cache Management', () => {
        test('should cache module content', () => {
            const moduleName = 'test-module';
            const content = { test: 'value' };

            bundleOptimizer.cacheModule(moduleName, content);

            expect(bundleOptimizer.cache.has(moduleName)).toBe(true);
            expect(bundleOptimizer.cache.get(moduleName).content).toEqual(content);
        });

        test('should check cache validity', () => {
            const validCache = {
                content: { test: 'value' },
                timestamp: Date.now(),
                version: '1.0.0'
            };

            const invalidCache = {
                content: { test: 'value' },
                timestamp: Date.now() - 25 * 60 * 60 * 1000, // 25 hours ago
                version: '1.0.0'
            };

            expect(bundleOptimizer.isCacheValid(validCache)).toBe(true);
            expect(bundleOptimizer.isCacheValid(invalidCache)).toBe(false);
        });

        test('should clear cache', () => {
            bundleOptimizer.cache.set('test-module', { content: 'test' });
            expect(bundleOptimizer.cache.size).toBe(1);

            bundleOptimizer.clearCache();
            expect(bundleOptimizer.cache.size).toBe(0);
        });
    });

    describe('Statistics', () => {
        test('should return bundle statistics', () => {
            const stats = bundleOptimizer.getBundleStats();

            expect(stats).toHaveProperty('totalSize');
            expect(stats).toHaveProperty('loadedSize');
            expect(stats).toHaveProperty('moduleCount');
            expect(stats).toHaveProperty('loadedModuleCount');
            expect(stats).toHaveProperty('cacheSize');
            expect(stats).toHaveProperty('cacheEnabled');
            expect(stats).toHaveProperty('averageLoadTime');
        });

        test('should update statistics after loading modules', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            const initialStats = bundleOptimizer.getBundleStats();
            await bundleOptimizer.loadModule('merchant-core');
            const finalStats = bundleOptimizer.getBundleStats();

            expect(finalStats.loadedModuleCount).toBe(initialStats.loadedModuleCount + 1);
            expect(finalStats.loadedSize).toBeGreaterThan(initialStats.loadedSize);
        });
    });

    describe('Event Handling', () => {
        test('should trigger module loaded event', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            const eventHandler = jest.fn();
            window.addEventListener('bundle:moduleLoaded', eventHandler);

            await bundleOptimizer.loadModule('merchant-core');

            expect(eventHandler).toHaveBeenCalled();
            expect(eventHandler.mock.calls[0][0].detail.module).toBe('merchant-core');

            window.removeEventListener('bundle:moduleLoaded', eventHandler);
        });

        test('should trigger module error event', async () => {
            fetch.mockRejectedValueOnce(new Error('Network error'));

            const eventHandler = jest.fn();
            window.addEventListener('bundle:moduleError', eventHandler);

            try {
                await bundleOptimizer.loadModule('merchant-core');
            } catch (error) {
                // Expected error
            }

            expect(eventHandler).toHaveBeenCalled();
            expect(eventHandler.mock.calls[0][0].detail.module).toBe('merchant-core');

            window.removeEventListener('bundle:moduleError', eventHandler);
        });
    });

    describe('Destruction', () => {
        test('should destroy bundle optimizer and clean up', () => {
            bundleOptimizer.loadedModules.add('test-module');
            bundleOptimizer.cache.set('test-module', { content: 'test' });
            bundleOptimizer.loadingPromises.set('test-module', Promise.resolve());

            bundleOptimizer.destroy();

            expect(bundleOptimizer.loadedModules.size).toBe(0);
            expect(bundleOptimizer.cache.size).toBe(0);
            expect(bundleOptimizer.loadingPromises.size).toBe(0);
        });
    });
});

describe('DynamicImportManager', () => {
    let bundleOptimizer;
    let importManager;

    beforeEach(() => {
        bundleOptimizer = new BundleOptimizer({
            baseUrl: '/test/',
            cachePrefix: 'test-bundle-',
            cacheVersion: '1.0.0'
        });
        importManager = new DynamicImportManager(bundleOptimizer);
    });

    afterEach(() => {
        if (bundleOptimizer) {
            bundleOptimizer.destroy();
        }
    });

    describe('Import with Fallback', () => {
        test('should import module successfully', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            const result = await importManager.importWithFallback('merchant-core');

            expect(result).toBeDefined();
        });

        test('should use fallback module on error', async () => {
            fetch.mockRejectedValueOnce(new Error('Network error'));
            
            const mockContent = 'module.exports = { fallback: "value" };';
            fetch.mockResolvedValueOnce({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            const result = await importManager.importWithFallback('merchant-core', 'utils');

            expect(result).toBeDefined();
            expect(fetch).toHaveBeenCalledTimes(2);
        });

        test('should throw error when both module and fallback fail', async () => {
            fetch.mockRejectedValue(new Error('Network error'));

            await expect(importManager.importWithFallback('merchant-core', 'utils')).rejects.toThrow('Network error');
        });
    });

    describe('Conditional Import', () => {
        test('should import module when condition is true', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            const result = await importManager.conditionalImport('merchant-core', () => true);

            expect(result).toBeDefined();
        });

        test('should return null when condition is false', async () => {
            const result = await importManager.conditionalImport('merchant-core', () => false);

            expect(result).toBeNull();
            expect(fetch).not.toHaveBeenCalled();
        });
    });

    describe('Lazy Import with Timeout', () => {
        test('should import module within timeout', async () => {
            const mockContent = 'module.exports = { test: "value" };';
            fetch.mockResolvedValue({
                ok: true,
                text: () => Promise.resolve(mockContent)
            });

            const result = await importManager.lazyImportWithTimeout('merchant-core', 1000);

            expect(result).toBeDefined();
        });

        test('should timeout when module takes too long', async () => {
            fetch.mockImplementation(() => 
                new Promise(resolve => setTimeout(() => resolve({
                    ok: true,
                    text: () => Promise.resolve('module.exports = { test: "value" };')
                }), 2000))
            );

            await expect(importManager.lazyImportWithTimeout('merchant-core', 1000)).rejects.toThrow(
                "Module 'merchant-core' load timeout"
            );
        });
    });
});
