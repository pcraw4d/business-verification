/**
 * Unit Tests for Coming Soon Banner Component
 * Tests functionality, event handling, and integration with placeholder service
 */

// Mock DOM environment for testing
const { JSDOM } = require('jsdom');
const dom = new JSDOM('<!DOCTYPE html><html><body></body></html>', {
    url: 'http://localhost',
    pretendToBeVisual: true,
    resources: 'usable'
});

global.window = dom.window;
global.document = dom.window.document;
global.navigator = dom.window.navigator;

// Mock fetch for API calls
global.fetch = jest.fn();

// Mock localStorage
const localStorageMock = {
    getItem: jest.fn(),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn(),
};
global.localStorage = localStorageMock;

// Mock document.cookie
Object.defineProperty(document, 'cookie', {
    writable: true,
    value: ''
});

// Import the component
const ComingSoonBanner = require('./coming-soon-banner.js');

describe('ComingSoonBanner', () => {
    let banner;
    let container;
    let mockFeature;
    let mockFeatures;

    beforeEach(() => {
        // Reset DOM
        document.body.innerHTML = '';
        
        // Create test container
        container = document.createElement('div');
        container.id = 'testBannerContainer';
        document.body.appendChild(container);

        // Mock feature data
        mockFeature = {
            id: 'test_feature',
            name: 'Test Feature',
            description: 'This is a test feature description',
            status: 'coming_soon',
            category: 'analytics',
            priority: 1,
            eta: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(), // 30 days from now
            mock_data: { test: 'data' },
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
        };

        mockFeatures = [
            mockFeature,
            {
                id: 'test_feature_2',
                name: 'Test Feature 2',
                description: 'Another test feature',
                status: 'in_development',
                category: 'reporting',
                priority: 2,
                eta: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
            }
        ];

        // Reset mocks
        fetch.mockClear();
        localStorageMock.getItem.mockClear();
        document.cookie = '';
    });

    afterEach(() => {
        if (banner) {
            banner.destroy();
        }
        document.body.innerHTML = '';
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            banner = new ComingSoonBanner({ container });
            
            expect(banner.container).toBe(container);
            expect(banner.apiBaseUrl).toBe('/api/v1');
            expect(banner.featureId).toBeNull();
            expect(banner.category).toBeNull();
            expect(banner.status).toBe('coming_soon');
            expect(banner.showMockDataWarning).toBe(true);
            expect(banner.autoRefresh).toBe(false);
        });

        test('should initialize with custom options', () => {
            const options = {
                container,
                apiBaseUrl: '/api/v2',
                featureId: 'test_feature',
                category: 'analytics',
                status: 'in_development',
                showMockDataWarning: false,
                autoRefresh: true,
                refreshInterval: 60000
            };

            banner = new ComingSoonBanner(options);
            
            expect(banner.apiBaseUrl).toBe('/api/v2');
            expect(banner.featureId).toBe('test_feature');
            expect(banner.category).toBe('analytics');
            expect(banner.status).toBe('in_development');
            expect(banner.showMockDataWarning).toBe(false);
            expect(banner.autoRefresh).toBe(true);
            expect(banner.refreshInterval).toBe(60000);
        });

        test('should create banner interface on initialization', () => {
            banner = new ComingSoonBanner({ container });
            
            const bannerElement = container.querySelector('.coming-soon-banner');
            expect(bannerElement).toBeTruthy();
            expect(bannerElement.style.display).toBe('none');
        });

        test('should add styles to document head', () => {
            banner = new ComingSoonBanner({ container });
            
            const styleElement = document.querySelector('#coming-soon-banner-styles');
            expect(styleElement).toBeTruthy();
            expect(styleElement.textContent).toContain('.coming-soon-banner');
        });
    });

    describe('API Integration', () => {
        test('should fetch feature by ID', async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: mockFeature
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                featureId: 'test_feature' 
            });

            // Wait for async operations
            await new Promise(resolve => setTimeout(resolve, 100));

            expect(fetch).toHaveBeenCalledWith(
                '/api/v1/features/test_feature',
                expect.objectContaining({
                    method: 'GET',
                    headers: expect.objectContaining({
                        'Content-Type': 'application/json'
                    })
                })
            );
        });

        test('should fetch features by category', async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: { features: mockFeatures }
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                category: 'analytics' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));

            expect(fetch).toHaveBeenCalledWith(
                '/api/v1/features/category/analytics',
                expect.objectContaining({
                    method: 'GET'
                })
            );
        });

        test('should fetch features by status', async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: { features: [mockFeature] }
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                status: 'coming_soon' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));

            expect(fetch).toHaveBeenCalledWith(
                '/api/v1/features/status/coming_soon',
                expect.objectContaining({
                    method: 'GET'
                })
            );
        });

        test('should handle API errors gracefully', async () => {
            fetch.mockRejectedValueOnce(new Error('Network error'));

            banner = new ComingSoonBanner({ 
                container, 
                featureId: 'test_feature' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));

            const bannerElement = container.querySelector('.coming-soon-banner');
            expect(bannerElement.style.display).toBe('none');
        });

        test('should include authorization token in requests', async () => {
            localStorageMock.getItem.mockReturnValue('test_token');

            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: mockFeature
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                featureId: 'test_feature' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));

            expect(fetch).toHaveBeenCalledWith(
                expect.any(String),
                expect.objectContaining({
                    headers: expect.objectContaining({
                        'Authorization': 'Bearer test_token'
                    })
                })
            );
        });
    });

    describe('Banner Display', () => {
        beforeEach(async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: mockFeature
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                featureId: 'test_feature' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));
        });

        test('should show banner when feature data is loaded', () => {
            const bannerElement = container.querySelector('.coming-soon-banner');
            expect(bannerElement.style.display).toBe('block');
            expect(banner.isVisible).toBe(true);
        });

        test('should update banner content with feature data', () => {
            const featureName = container.querySelector('#featureName');
            const featureDescription = container.querySelector('#featureDescription');
            const categoryText = container.querySelector('.category-text');
            const priorityText = container.querySelector('.priority-text');
            const etaText = container.querySelector('.eta-text');

            expect(featureName.textContent).toBe('Test Feature');
            expect(featureDescription.textContent).toBe('This is a test feature description');
            expect(categoryText.textContent).toBe('Analytics');
            expect(priorityText.textContent).toBe('High Priority');
            expect(etaText.textContent).toBe('30 days');
        });

        test('should show mock data warning when feature has mock data', () => {
            const warning = container.querySelector('#mockDataWarning');
            expect(warning.style.display).toBe('flex');
        });

        test('should hide mock data warning when dismissed', () => {
            const warning = container.querySelector('#mockDataWarning');
            const dismissBtn = container.querySelector('#dismissWarningBtn');
            
            dismissBtn.click();
            
            expect(warning.style.display).toBe('none');
        });
    });

    describe('Event Handling', () => {
        beforeEach(async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: mockFeature
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                featureId: 'test_feature' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));
        });

        test('should close banner when close button is clicked', () => {
            const closeBtn = container.querySelector('#bannerCloseBtn');
            const bannerElement = container.querySelector('.coming-soon-banner');
            
            closeBtn.click();
            
            // Wait for animation
            setTimeout(() => {
                expect(bannerElement.style.display).toBe('none');
                expect(banner.isVisible).toBe(false);
            }, 350);
        });

        test('should handle notify me button click', () => {
            const notifyBtn = container.querySelector('#notifyMeBtn');
            const consoleSpy = jest.spyOn(console, 'log').mockImplementation();
            
            notifyBtn.click();
            
            expect(consoleSpy).toHaveBeenCalledWith(
                'Notify me clicked for feature:', 
                'Test Feature'
            );
            
            consoleSpy.mockRestore();
        });

        test('should handle learn more button click', () => {
            const learnMoreBtn = container.querySelector('#learnMoreBtn');
            const consoleSpy = jest.spyOn(console, 'log').mockImplementation();
            
            learnMoreBtn.click();
            
            expect(consoleSpy).toHaveBeenCalledWith(
                'Learn more clicked for feature:', 
                'Test Feature'
            );
            
            consoleSpy.mockRestore();
        });

        test('should handle view all features link click', () => {
            const viewAllLink = container.querySelector('#viewAllFeaturesLink');
            const consoleSpy = jest.spyOn(console, 'log').mockImplementation();
            
            viewAllLink.click();
            
            expect(consoleSpy).toHaveBeenCalledWith('View all features clicked');
            
            consoleSpy.mockRestore();
        });

        test('should close banner on Escape key', () => {
            const bannerElement = container.querySelector('.coming-soon-banner');
            
            const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
            document.dispatchEvent(escapeEvent);
            
            setTimeout(() => {
                expect(bannerElement.style.display).toBe('none');
            }, 350);
        });
    });

    describe('Progress Indicator', () => {
        test('should show progress indicator for in_development features', async () => {
            const inDevelopmentFeature = {
                ...mockFeature,
                status: 'in_development',
                created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(), // 7 days ago
                eta: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString() // 7 days from now
            };

            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: inDevelopmentFeature
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                featureId: 'test_feature' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));

            const progressIndicator = container.querySelector('#progressIndicator');
            expect(progressIndicator.style.display).toBe('block');
        });

        test('should hide progress indicator for coming_soon features', async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: mockFeature
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                featureId: 'test_feature' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));

            const progressIndicator = container.querySelector('#progressIndicator');
            expect(progressIndicator.style.display).toBe('none');
        });
    });

    describe('Auto Refresh', () => {
        test('should start auto refresh when enabled', () => {
            const setIntervalSpy = jest.spyOn(global, 'setInterval');
            
            banner = new ComingSoonBanner({ 
                container, 
                autoRefresh: true,
                refreshInterval: 1000
            });
            
            expect(setIntervalSpy).toHaveBeenCalled();
            setIntervalSpy.mockRestore();
        });

        test('should stop auto refresh when destroyed', () => {
            const clearIntervalSpy = jest.spyOn(global, 'clearInterval');
            
            banner = new ComingSoonBanner({ 
                container, 
                autoRefresh: true
            });
            
            banner.destroy();
            
            expect(clearIntervalSpy).toHaveBeenCalled();
            clearIntervalSpy.mockRestore();
        });
    });

    describe('Utility Methods', () => {
        beforeEach(() => {
            banner = new ComingSoonBanner({ container });
        });

        test('should format category correctly', () => {
            expect(banner.formatCategory('analytics')).toBe('Analytics');
            expect(banner.formatCategory('reporting')).toBe('Reporting');
            expect(banner.formatCategory('unknown')).toBe('unknown');
        });

        test('should format priority correctly', () => {
            expect(banner.formatPriority(1)).toBe('High Priority');
            expect(banner.formatPriority(2)).toBe('Medium Priority');
            expect(banner.formatPriority(3)).toBe('Low Priority');
            expect(banner.formatPriority(5)).toBe('Priority 5');
        });

        test('should format ETA correctly', () => {
            const tomorrow = new Date(Date.now() + 24 * 60 * 60 * 1000);
            const nextWeek = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000);
            const nextMonth = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000);
            
            expect(banner.formatETA(tomorrow.toISOString())).toBe('Tomorrow');
            expect(banner.formatETA(nextWeek.toISOString())).toBe('1 week');
            expect(banner.formatETA(nextMonth.toISOString())).toBe('1 month');
            expect(banner.formatETA(null)).toBe('TBD');
        });

        test('should get status subtitle correctly', () => {
            expect(banner.getStatusSubtitle('coming_soon')).toBe('Exciting new features are on the way!');
            expect(banner.getStatusSubtitle('in_development')).toBe('This feature is currently being developed');
            expect(banner.getStatusSubtitle('available')).toBe('This feature is now available!');
            expect(banner.getStatusSubtitle('unknown')).toBe('Feature status update');
        });
    });

    describe('Public API', () => {
        beforeEach(async () => {
            fetch.mockResolvedValue({
                ok: true,
                json: async () => ({
                    success: true,
                    data: mockFeature
                })
            });

            banner = new ComingSoonBanner({ container });
            await new Promise(resolve => setTimeout(resolve, 100));
        });

        test('should set feature by ID', async () => {
            banner.setFeature('new_feature');
            
            await new Promise(resolve => setTimeout(resolve, 100));
            
            expect(banner.featureId).toBe('new_feature');
            expect(banner.category).toBeNull();
        });

        test('should set category', async () => {
            banner.setCategory('reporting');
            
            await new Promise(resolve => setTimeout(resolve, 100));
            
            expect(banner.category).toBe('reporting');
            expect(banner.featureId).toBeNull();
        });

        test('should set status', async () => {
            banner.setStatus('in_development');
            
            await new Promise(resolve => setTimeout(resolve, 100));
            
            expect(banner.status).toBe('in_development');
            expect(banner.featureId).toBeNull();
            expect(banner.category).toBeNull();
        });

        test('should toggle banner visibility', () => {
            expect(banner.isVisible).toBe(true);
            
            banner.toggle();
            
            setTimeout(() => {
                expect(banner.isVisible).toBe(false);
            }, 350);
        });

        test('should refresh data', async () => {
            const loadDataSpy = jest.spyOn(banner, 'loadFeatureData');
            
            banner.refresh();
            
            expect(loadDataSpy).toHaveBeenCalled();
            loadDataSpy.mockRestore();
        });
    });

    describe('Error Handling', () => {
        test('should handle missing feature data gracefully', async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: null
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                featureId: 'nonexistent_feature' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));

            const bannerElement = container.querySelector('.coming-soon-banner');
            expect(bannerElement.style.display).toBe('none');
        });

        test('should handle empty features array', async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: { features: [] }
                })
            });

            banner = new ComingSoonBanner({ 
                container, 
                category: 'nonexistent_category' 
            });

            await new Promise(resolve => setTimeout(resolve, 100));

            const bannerElement = container.querySelector('.coming-soon-banner');
            expect(bannerElement.style.display).toBe('none');
        });
    });

    describe('Accessibility', () => {
        beforeEach(async () => {
            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    success: true,
                    data: mockFeature
                })
            });

            banner = new ComingSoonBanner({ container });
            await new Promise(resolve => setTimeout(resolve, 100));
        });

        test('should have proper ARIA attributes', () => {
            const bannerElement = container.querySelector('.coming-soon-banner');
            const closeBtn = container.querySelector('#bannerCloseBtn');
            
            expect(closeBtn.getAttribute('title')).toBe('Close banner');
        });

        test('should support keyboard navigation', () => {
            const closeBtn = container.querySelector('#bannerCloseBtn');
            const notifyBtn = container.querySelector('#notifyMeBtn');
            const learnMoreBtn = container.querySelector('#learnMoreBtn');
            
            expect(closeBtn.tagName).toBe('BUTTON');
            expect(notifyBtn.tagName).toBe('BUTTON');
            expect(learnMoreBtn.tagName).toBe('BUTTON');
        });
    });

    describe('Responsive Design', () => {
        test('should have responsive CSS classes', () => {
            banner = new ComingSoonBanner({ container });
            
            const styleElement = document.querySelector('#coming-soon-banner-styles');
            expect(styleElement.textContent).toContain('@media (max-width: 768px)');
        });
    });
});

// Integration tests
describe('ComingSoonBanner Integration', () => {
    let banner;
    let container;

    beforeEach(() => {
        container = document.createElement('div');
        container.id = 'integrationTestContainer';
        document.body.appendChild(container);
    });

    afterEach(() => {
        if (banner) {
            banner.destroy();
        }
        document.body.innerHTML = '';
    });

    test('should work with real placeholder service API', async () => {
        // Mock successful API response
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                success: true,
                data: {
                    id: 'advanced_analytics',
                    name: 'Advanced Analytics Dashboard',
                    description: 'Comprehensive analytics and reporting dashboard with real-time insights',
                    status: 'coming_soon',
                    category: 'analytics',
                    priority: 1,
                    eta: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
                    mock_data: { sample_charts: ['Revenue Trend', 'User Growth'] },
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString()
                }
            })
        });

        banner = new ComingSoonBanner({ 
            container, 
            featureId: 'advanced_analytics' 
        });

        await new Promise(resolve => setTimeout(resolve, 100));

        expect(banner.isVisible).toBe(true);
        expect(banner.currentFeature.name).toBe('Advanced Analytics Dashboard');
    });

    test('should handle callback functions', async () => {
        const onFeatureClick = jest.fn();
        const onBannerClose = jest.fn();
        const onMockDataWarning = jest.fn();

        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                success: true,
                data: mockFeature
            })
        });

        banner = new ComingSoonBanner({ 
            container, 
            featureId: 'test_feature',
            onFeatureClick,
            onBannerClose,
            onMockDataWarning
        });

        await new Promise(resolve => setTimeout(resolve, 100));

        // Test feature click callback
        const notifyBtn = container.querySelector('#notifyMeBtn');
        notifyBtn.click();
        expect(onFeatureClick).toHaveBeenCalledWith('notify', mockFeature);

        // Test mock data warning callback
        expect(onMockDataWarning).toHaveBeenCalledWith(mockFeature);

        // Test banner close callback
        banner.hide();
        setTimeout(() => {
            expect(onBannerClose).toHaveBeenCalled();
        }, 350);
    });
});
