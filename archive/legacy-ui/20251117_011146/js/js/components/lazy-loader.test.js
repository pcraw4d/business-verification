/**
 * Lazy Loader Unit Tests
 * Tests for lazy loading functionality and performance optimization
 */

const { LazyLoader, MerchantLazyLoader } = require('./lazy-loader');

describe('LazyLoader', () => {
    let lazyLoader;
    let mockElement;

    beforeEach(() => {
        // Create mock element
        mockElement = document.createElement('div');
        mockElement.classList.add('test-element');
        document.body.appendChild(mockElement);

        // Create lazy loader instance
        lazyLoader = new LazyLoader({
            rootMargin: '50px',
            threshold: 0.1
        });
    });

    afterEach(() => {
        if (lazyLoader) {
            lazyLoader.destroy();
        }
        if (mockElement && mockElement.parentNode) {
            mockElement.parentNode.removeChild(mockElement);
        }
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            const loader = new LazyLoader();
            expect(loader.options.rootMargin).toBe('50px');
            expect(loader.options.threshold).toBe(0.1);
            expect(loader.options.loadingClass).toBe('lazy-loading');
        });

        test('should initialize with custom options', () => {
            const customOptions = {
                rootMargin: '100px',
                threshold: 0.2,
                loadingClass: 'custom-loading'
            };
            const loader = new LazyLoader(customOptions);
            expect(loader.options.rootMargin).toBe('100px');
            expect(loader.options.threshold).toBe(0.2);
            expect(loader.options.loadingClass).toBe('custom-loading');
        });

        test('should create IntersectionObserver when available', () => {
            expect(lazyLoader.observer).toBeDefined();
            expect(lazyLoader.observer).toBeInstanceOf(IntersectionObserver);
        });
    });

    describe('Element Registration', () => {
        test('should register element with loader function', () => {
            const mockLoader = jest.fn().mockResolvedValue('loaded');
            
            lazyLoader.register(mockElement, mockLoader);
            
            expect(lazyLoader.elements.has(mockElement)).toBe(true);
            expect(mockElement.classList.contains('lazy-loading')).toBe(true);
        });

        test('should not register invalid element', () => {
            const mockLoader = jest.fn();
            
            lazyLoader.register(null, mockLoader);
            lazyLoader.register(undefined, mockLoader);
            
            expect(lazyLoader.elements.size).toBe(0);
        });

        test('should not register element with invalid loader', () => {
            lazyLoader.register(mockElement, null);
            lazyLoader.register(mockElement, undefined);
            lazyLoader.register(mockElement, 'not-a-function');
            
            expect(lazyLoader.elements.size).toBe(0);
        });

        test('should add loading class to registered element', () => {
            const mockLoader = jest.fn().mockResolvedValue('loaded');
            
            lazyLoader.register(mockElement, mockLoader);
            
            expect(mockElement.classList.contains('lazy-loading')).toBe(true);
        });
    });

    describe('Element Loading', () => {
        test('should load element successfully', async () => {
            const mockLoader = jest.fn().mockResolvedValue('loaded');
            const loadedEvent = jest.fn();
            
            mockElement.addEventListener('lazyLoaded', loadedEvent);
            lazyLoader.register(mockElement, mockLoader);
            
            await lazyLoader.loadElement(mockElement);
            
            expect(mockLoader).toHaveBeenCalled();
            expect(mockElement.classList.contains('lazy-loaded')).toBe(true);
            expect(mockElement.classList.contains('lazy-loading')).toBe(false);
            expect(loadedEvent).toHaveBeenCalled();
        });

        test('should handle loading errors', async () => {
            const mockLoader = jest.fn().mockRejectedValue(new Error('Loading failed'));
            const errorEvent = jest.fn();
            
            mockElement.addEventListener('lazyError', errorEvent);
            lazyLoader.register(mockElement, mockLoader);
            
            await lazyLoader.loadElement(mockElement);
            
            expect(mockLoader).toHaveBeenCalled();
            expect(mockElement.classList.contains('lazy-error')).toBe(true);
            expect(mockElement.classList.contains('lazy-loading')).toBe(false);
            expect(errorEvent).toHaveBeenCalled();
        });

        test('should not load element twice', async () => {
            const mockLoader = jest.fn().mockResolvedValue('loaded');
            
            lazyLoader.register(mockElement, mockLoader);
            
            await lazyLoader.loadElement(mockElement);
            await lazyLoader.loadElement(mockElement);
            
            expect(mockLoader).toHaveBeenCalledTimes(1);
        });

        test('should not load element while loading', async () => {
            const mockLoader = jest.fn().mockImplementation(() => 
                new Promise(resolve => setTimeout(() => resolve('loaded'), 100))
            );
            
            lazyLoader.register(mockElement, mockLoader);
            
            // Start loading
            const loadPromise1 = lazyLoader.loadElement(mockElement);
            const loadPromise2 = lazyLoader.loadElement(mockElement);
            
            await Promise.all([loadPromise1, loadPromise2]);
            
            expect(mockLoader).toHaveBeenCalledTimes(1);
        });
    });

    describe('Preloading', () => {
        test('should preload registered element', async () => {
            const mockLoader = jest.fn().mockResolvedValue('loaded');
            
            lazyLoader.register(mockElement, mockLoader);
            lazyLoader.preload(mockElement);
            
            expect(mockLoader).toHaveBeenCalled();
        });

        test('should not preload unregistered element', () => {
            const mockLoader = jest.fn();
            
            lazyLoader.preload(mockElement);
            
            expect(mockLoader).not.toHaveBeenCalled();
        });
    });

    describe('Unregistration', () => {
        test('should unregister element', () => {
            const mockLoader = jest.fn().mockResolvedValue('loaded');
            
            lazyLoader.register(mockElement, mockLoader);
            expect(lazyLoader.elements.has(mockElement)).toBe(true);
            
            lazyLoader.unregister(mockElement);
            expect(lazyLoader.elements.has(mockElement)).toBe(false);
        });
    });

    describe('Statistics', () => {
        test('should return correct loading statistics', async () => {
            const mockLoader1 = jest.fn().mockResolvedValue('loaded');
            const mockLoader2 = jest.fn().mockImplementation(() => 
                new Promise(resolve => setTimeout(() => resolve('loaded'), 100))
            );
            const mockLoader3 = jest.fn().mockRejectedValue(new Error('Failed'));
            
            const element2 = document.createElement('div');
            const element3 = document.createElement('div');
            
            lazyLoader.register(mockElement, mockLoader1);
            lazyLoader.register(element2, mockLoader2);
            lazyLoader.register(element3, mockLoader3);
            
            // Load first element
            await lazyLoader.loadElement(mockElement);
            
            // Start loading second element
            lazyLoader.loadElement(element2);
            
            // Load third element (will fail)
            await lazyLoader.loadElement(element3);
            
            const stats = lazyLoader.getStats();
            
            expect(stats.total).toBe(3);
            expect(stats.loaded).toBe(1);
            expect(stats.loading).toBe(1);
            expect(stats.errors).toBe(1);
            expect(stats.pending).toBe(0);
            
            // Clean up
            element2.parentNode?.removeChild(element2);
            element3.parentNode?.removeChild(element3);
        });
    });

    describe('Destruction', () => {
        test('should destroy lazy loader and clean up', () => {
            const mockLoader = jest.fn().mockResolvedValue('loaded');
            
            lazyLoader.register(mockElement, mockLoader);
            expect(lazyLoader.elements.size).toBe(1);
            
            lazyLoader.destroy();
            
            expect(lazyLoader.elements.size).toBe(0);
            expect(lazyLoader.loadingPromises.size).toBe(0);
        });
    });
});

describe('MerchantLazyLoader', () => {
    let merchantLazyLoader;
    let mockElement;

    beforeEach(() => {
        mockElement = document.createElement('div');
        document.body.appendChild(mockElement);
        
        merchantLazyLoader = new MerchantLazyLoader();
    });

    afterEach(() => {
        if (merchantLazyLoader) {
            merchantLazyLoader.destroy();
        }
        if (mockElement && mockElement.parentNode) {
            mockElement.parentNode.removeChild(mockElement);
        }
    });

    describe('Initialization', () => {
        test('should initialize with component loaders', () => {
            expect(merchantLazyLoader.componentLoaders.size).toBe(5);
            expect(merchantLazyLoader.componentLoaders.has('merchant-card')).toBe(true);
            expect(merchantLazyLoader.componentLoaders.has('merchant-chart')).toBe(true);
            expect(merchantLazyLoader.componentLoaders.has('merchant-details')).toBe(true);
            expect(merchantLazyLoader.componentLoaders.has('bulk-operations')).toBe(true);
            expect(merchantLazyLoader.componentLoaders.has('comparison-view')).toBe(true);
        });
    });

    describe('Component Registration', () => {
        test('should register merchant card', () => {
            const merchantData = { id: 'test-1', name: 'Test Merchant' };
            
            merchantLazyLoader.registerMerchantCard(mockElement, merchantData);
            
            expect(merchantLazyLoader.lazyLoader.elements.has(mockElement)).toBe(true);
        });

        test('should register merchant chart', () => {
            const chartConfig = { type: 'bar', data: [] };
            
            merchantLazyLoader.registerMerchantChart(mockElement, chartConfig);
            
            expect(merchantLazyLoader.lazyLoader.elements.has(mockElement)).toBe(true);
        });

        test('should register merchant details', () => {
            const merchantId = 'test-merchant-123';
            
            merchantLazyLoader.registerMerchantDetails(mockElement, merchantId);
            
            expect(merchantLazyLoader.lazyLoader.elements.has(mockElement)).toBe(true);
        });

        test('should register bulk operations', () => {
            const selectedMerchants = ['merchant-1', 'merchant-2'];
            
            merchantLazyLoader.registerBulkOperations(mockElement, selectedMerchants);
            
            expect(merchantLazyLoader.lazyLoader.elements.has(mockElement)).toBe(true);
        });

        test('should register comparison view', () => {
            const merchants = [
                { id: 'merchant-1', name: 'Merchant 1' },
                { id: 'merchant-2', name: 'Merchant 2' }
            ];
            
            merchantLazyLoader.registerComparisonView(mockElement, merchants);
            
            expect(merchantLazyLoader.lazyLoader.elements.has(mockElement)).toBe(true);
        });
    });

    describe('Component Loading', () => {
        test('should load merchant card component', async () => {
            const merchantData = { 
                id: 'test-1', 
                name: 'Test Merchant',
                riskLevel: 'medium',
                industry: 'Technology',
                portfolioType: 'onboarded',
                location: 'San Francisco, CA'
            };
            
            const result = await merchantLazyLoader.componentLoaders.get('merchant-card')(merchantData);
            
            expect(result.type).toBe('merchant-card');
            expect(result.data).toEqual(merchantData);
            expect(result.template).toContain('Test Merchant');
            expect(result.template).toContain('medium');
        });

        test('should load merchant chart component', async () => {
            const chartConfig = { 
                type: 'bar', 
                data: [1, 2, 3],
                options: { responsive: true }
            };
            
            const result = await merchantLazyLoader.componentLoaders.get('merchant-chart')(chartConfig);
            
            expect(result.type).toBe('merchant-chart');
            expect(result.config).toEqual(chartConfig);
            expect(result.chart).toBeDefined();
        });

        test('should load merchant details component', async () => {
            const merchantId = 'test-merchant-123';
            
            const result = await merchantLazyLoader.componentLoaders.get('merchant-details')(merchantId);
            
            expect(result.type).toBe('merchant-details');
            expect(result.merchantId).toBe(merchantId);
            expect(result.details).toBeDefined();
        });

        test('should load bulk operations component', async () => {
            const selectedMerchants = ['merchant-1', 'merchant-2'];
            
            const result = await merchantLazyLoader.componentLoaders.get('bulk-operations')(selectedMerchants);
            
            expect(result.type).toBe('bulk-operations');
            expect(result.selectedMerchants).toEqual(selectedMerchants);
            expect(result.operations).toBeDefined();
            expect(Array.isArray(result.operations)).toBe(true);
        });

        test('should load comparison view component', async () => {
            const merchants = [
                { id: 'merchant-1', name: 'Merchant 1', riskLevel: 'low', portfolioType: 'onboarded', industry: 'Tech', location: 'SF' },
                { id: 'merchant-2', name: 'Merchant 2', riskLevel: 'high', portfolioType: 'pending', industry: 'Finance', location: 'NYC' }
            ];
            
            const result = await merchantLazyLoader.componentLoaders.get('comparison-view')(merchants);
            
            expect(result.type).toBe('comparison-view');
            expect(result.merchants).toEqual(merchants);
            expect(result.comparison).toBeDefined();
            expect(result.comparison.length).toBe(2);
        });
    });

    describe('Statistics', () => {
        test('should return loading statistics', () => {
            const stats = merchantLazyLoader.getStats();
            
            expect(stats).toHaveProperty('total');
            expect(stats).toHaveProperty('loaded');
            expect(stats).toHaveProperty('loading');
            expect(stats).toHaveProperty('errors');
            expect(stats).toHaveProperty('pending');
        });
    });
});
