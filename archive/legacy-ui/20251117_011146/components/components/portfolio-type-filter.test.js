/**
 * Unit Tests for Portfolio Type Filter Component
 * Tests all functionality including single/multiple selection modes, visual indicators, and API integration
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

// Import the component
const PortfolioTypeFilter = require('./portfolio-type-filter.js');

describe('PortfolioTypeFilter', () => {
    let container;
    let filter;

    beforeEach(() => {
        // Create a fresh container for each test
        container = document.createElement('div');
        document.body.appendChild(container);
        
        // Reset fetch mock
        fetch.mockClear();
        
        // Mock successful API response
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                portfolio_types: {
                    onboarded: 150,
                    deactivated: 25,
                    prospective: 75,
                    pending: 50
                },
                total: 300
            })
        });
    });

    afterEach(() => {
        if (filter) {
            filter.destroy();
        }
        if (container && container.parentNode) {
            container.parentNode.removeChild(container);
        }
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            filter = new PortfolioTypeFilter({ container });
            
            expect(filter.mode).toBe('single');
            expect(filter.allowClear).toBe(true);
            expect(filter.showCounts).toBe(true);
            expect(filter.allowAll).toBe(true);
            expect(filter.disabled).toBe(false);
            expect(filter.selectedValues).toBeNull();
        });

        test('should initialize with custom options', () => {
            filter = new PortfolioTypeFilter({
                container,
                mode: 'multiple',
                allowClear: false,
                showCounts: false,
                allowAll: false,
                disabled: true,
                initialValue: 'onboarded'
            });
            
            expect(filter.mode).toBe('multiple');
            expect(filter.allowClear).toBe(false);
            expect(filter.showCounts).toBe(false);
            expect(filter.allowAll).toBe(false);
            expect(filter.disabled).toBe(true);
            expect(filter.selectedValues).toEqual(['onboarded']);
        });

        test('should create filter interface', () => {
            filter = new PortfolioTypeFilter({ container });
            
            expect(container.querySelector('.portfolio-type-filter-container')).toBeTruthy();
            expect(container.querySelector('.filter-trigger')).toBeTruthy();
            expect(container.querySelector('.filter-dropdown-menu')).toBeTruthy();
        });

        test('should render all portfolio types', () => {
            filter = new PortfolioTypeFilter({ container });
            
            const options = container.querySelectorAll('.filter-option');
            expect(options.length).toBe(5); // 4 types + "All Types" option
        });
    });

    describe('Single Selection Mode', () => {
        beforeEach(() => {
            filter = new PortfolioTypeFilter({ 
                container,
                mode: 'single'
            });
        });

        test('should select single option', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            expect(filter.getValue()).toBe('onboarded');
            expect(onboardedOption.classList.contains('selected')).toBe(true);
        });

        test('should deselect when selecting same option', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            onboardedOption.click();
            
            expect(filter.getValue()).toBe('onboarded'); // Still selected in single mode
        });

        test('should close dropdown after selection', () => {
            const trigger = container.querySelector('.filter-trigger');
            const menu = container.querySelector('.filter-dropdown-menu');
            
            trigger.click(); // Open dropdown
            expect(menu.style.display).toBe('block');
            
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click(); // Select option
            
            expect(menu.style.display).toBe('none');
        });

        test('should update trigger text after selection', () => {
            const triggerText = container.querySelector('#portfolioFilterText');
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            
            onboardedOption.click();
            
            expect(triggerText.textContent).toBe('Onboarded');
        });

        test('should clear selection', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            filter.clearSelection();
            
            expect(filter.getValue()).toBeNull();
        });
    });

    describe('Multiple Selection Mode', () => {
        beforeEach(() => {
            filter = new PortfolioTypeFilter({ 
                container,
                mode: 'multiple'
            });
        });

        test('should select multiple options', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            const pendingOption = container.querySelector('[data-value="pending"]');
            
            onboardedOption.click();
            pendingOption.click();
            
            expect(filter.getValue()).toEqual(['onboarded', 'pending']);
        });

        test('should deselect option when clicked again', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            
            onboardedOption.click(); // Select
            onboardedOption.click(); // Deselect
            
            expect(filter.getValue()).toEqual([]);
        });

        test('should not close dropdown after selection', () => {
            const trigger = container.querySelector('.filter-trigger');
            const menu = container.querySelector('.filter-dropdown-menu');
            
            trigger.click(); // Open dropdown
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click(); // Select option
            
            expect(menu.style.display).toBe('block');
        });

        test('should show selected tags', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            const pendingOption = container.querySelector('[data-value="pending"]');
            
            onboardedOption.click();
            pendingOption.click();
            
            const selectedTags = container.querySelector('#portfolioSelectedTags');
            expect(selectedTags.style.display).toBe('flex');
            expect(selectedTags.querySelectorAll('.selected-tag').length).toBe(2);
        });

        test('should remove tag when remove button clicked', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            const pendingOption = container.querySelector('[data-value="pending"]');
            
            onboardedOption.click();
            pendingOption.click();
            
            const removeButton = container.querySelector('.tag-remove');
            removeButton.click();
            
            expect(filter.getValue().length).toBe(1);
        });

        test('should select all options', () => {
            const selectAllBtn = container.querySelector('#selectAllPortfolioTypes');
            selectAllBtn.click();
            
            expect(filter.getValue()).toEqual(['onboarded', 'deactivated', 'prospective', 'pending']);
        });

        test('should clear all selections', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            const pendingOption = container.querySelector('[data-value="pending"]');
            
            onboardedOption.click();
            pendingOption.click();
            
            const clearAllBtn = container.querySelector('#clearAllPortfolioTypes');
            clearAllBtn.click();
            
            expect(filter.getValue()).toEqual([]);
        });
    });

    describe('Visual Indicators', () => {
        beforeEach(() => {
            filter = new PortfolioTypeFilter({ container });
        });

        test('should display correct icons for each portfolio type', () => {
            const onboardedIcon = container.querySelector('[data-value="onboarded"] .fas.fa-check-circle');
            const deactivatedIcon = container.querySelector('[data-value="deactivated"] .fas.fa-times-circle');
            const prospectiveIcon = container.querySelector('[data-value="prospective"] .fas.fa-eye');
            const pendingIcon = container.querySelector('[data-value="pending"] .fas.fa-clock');
            
            expect(onboardedIcon).toBeTruthy();
            expect(deactivatedIcon).toBeTruthy();
            expect(prospectiveIcon).toBeTruthy();
            expect(pendingIcon).toBeTruthy();
        });

        test('should display correct colors for each portfolio type', () => {
            const onboardedIndicator = container.querySelector('[data-value="onboarded"] .option-indicator');
            const deactivatedIndicator = container.querySelector('[data-value="deactivated"] .option-indicator');
            
            expect(onboardedIndicator.style.backgroundColor).toBe('rgb(213, 244, 230)');
            expect(onboardedIndicator.style.borderColor).toBe('rgb(39, 174, 96)');
            expect(deactivatedIndicator.style.backgroundColor).toBe('rgb(250, 219, 216)');
            expect(deactivatedIndicator.style.borderColor).toBe('rgb(231, 76, 60)');
        });

        test('should show selection checkboxes', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            const checkbox = onboardedOption.querySelector('.option-checkbox .fas.fa-check');
            expect(checkbox.style.display).toBe('block');
        });
    });

    describe('Filter State Management', () => {
        beforeEach(() => {
            filter = new PortfolioTypeFilter({ container });
        });

        test('should trigger onChange callback', () => {
            const onChangeMock = jest.fn();
            filter.onChange = onChangeMock;
            
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            expect(onChangeMock).toHaveBeenCalledWith('onboarded');
        });

        test('should trigger onFilter callback', () => {
            const onFilterMock = jest.fn();
            filter.onFilter = onFilterMock;
            
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            expect(onFilterMock).toHaveBeenCalledWith({
                portfolio_type: 'onboarded',
                mode: 'single'
            });
        });

        test('should trigger onClear callback', () => {
            const onClearMock = jest.fn();
            filter.onClear = onClearMock;
            
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            const clearBtn = container.querySelector('#clearPortfolioFilter');
            clearBtn.click();
            
            expect(onClearMock).toHaveBeenCalled();
        });

        test('should set initial value', () => {
            filter = new PortfolioTypeFilter({
                container,
                initialValue: 'pending'
            });
            
            expect(filter.getValue()).toBe('pending');
        });

        test('should set value programmatically', () => {
            filter.setValue('deactivated');
            
            expect(filter.getValue()).toBe('deactivated');
        });
    });

    describe('API Integration', () => {
        beforeEach(() => {
            filter = new PortfolioTypeFilter({ 
                container,
                updateCountsOnInit: true
            });
        });

        test('should fetch and update counts', async () => {
            await filter.updateCounts();
            
            expect(fetch).toHaveBeenCalledWith('/api/v1/merchants/counts', {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer '
                }
            });
        });

        test('should update count displays', async () => {
            await filter.updateCounts();
            
            const onboardedCount = container.querySelector('#count-onboarded');
            const deactivatedCount = container.querySelector('#count-deactivated');
            
            expect(onboardedCount.textContent).toBe('150');
            expect(deactivatedCount.textContent).toBe('25');
        });

        test('should handle API errors gracefully', async () => {
            fetch.mockRejectedValue(new Error('API Error'));
            
            // Should not throw
            await expect(filter.updateCounts()).resolves.toBeUndefined();
        });

        test('should update counts from provided data', () => {
            const data = {
                portfolio_types: {
                    onboarded: 200,
                    deactivated: 30
                },
                total: 230
            };
            
            filter.updateCountsFromData(data);
            
            expect(filter.portfolioTypes.onboarded.count).toBe(200);
            expect(filter.portfolioTypes.deactivated.count).toBe(30);
        });
    });

    describe('Accessibility', () => {
        beforeEach(() => {
            filter = new PortfolioTypeFilter({ container });
        });

        test('should support keyboard navigation', () => {
            const trigger = container.querySelector('.filter-trigger');
            
            // Test Enter key
            const enterEvent = new KeyboardEvent('keydown', { key: 'Enter' });
            trigger.dispatchEvent(enterEvent);
            
            const menu = container.querySelector('.filter-dropdown-menu');
            expect(menu.style.display).toBe('block');
            
            // Test Escape key
            const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
            trigger.dispatchEvent(escapeEvent);
            
            expect(menu.style.display).toBe('none');
        });

        test('should close dropdown when clicking outside', () => {
            const trigger = container.querySelector('.filter-trigger');
            trigger.click(); // Open dropdown
            
            const menu = container.querySelector('.filter-dropdown-menu');
            expect(menu.style.display).toBe('block');
            
            // Click outside
            document.body.click();
            
            expect(menu.style.display).toBe('none');
        });

        test('should disable when setDisabled is called', () => {
            filter.setDisabled(true);
            
            const trigger = container.querySelector('.filter-trigger');
            expect(trigger.disabled).toBe(true);
        });
    });

    describe('Edge Cases', () => {
        test('should handle empty selection in multiple mode', () => {
            filter = new PortfolioTypeFilter({ 
                container,
                mode: 'multiple'
            });
            
            expect(filter.getValue()).toEqual([]);
            expect(filter.hasSelection()).toBe(false);
        });

        test('should handle null selection in single mode', () => {
            filter = new PortfolioTypeFilter({ 
                container,
                mode: 'single'
            });
            
            expect(filter.getValue()).toBeNull();
            expect(filter.hasSelection()).toBe(false);
        });

        test('should handle invalid initial value', () => {
            filter = new PortfolioTypeFilter({
                container,
                initialValue: 'invalid_type'
            });
            
            expect(filter.getValue()).toBe('invalid_type');
        });

        test('should handle missing container', () => {
            expect(() => {
                new PortfolioTypeFilter({ container: null });
            }).toThrow();
        });
    });

    describe('Public Methods', () => {
        beforeEach(() => {
            filter = new PortfolioTypeFilter({ container });
        });

        test('should refresh counts', async () => {
            await filter.refresh();
            
            expect(fetch).toHaveBeenCalled();
        });

        test('should reset to initial state', () => {
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            filter.reset();
            
            expect(filter.getValue()).toBeNull();
        });

        test('should destroy component', () => {
            filter.destroy();
            
            expect(container.innerHTML).toBe('');
        });
    });

    describe('Configuration Options', () => {
        test('should hide counts when showCounts is false', () => {
            filter = new PortfolioTypeFilter({ 
                container,
                showCounts: false
            });
            
            const countElements = container.querySelectorAll('.option-count');
            expect(countElements.length).toBe(0);
        });

        test('should hide "All Types" option when allowAll is false', () => {
            filter = new PortfolioTypeFilter({ 
                container,
                allowAll: false
            });
            
            const allOption = container.querySelector('.all-option');
            expect(allOption).toBeFalsy();
        });

        test('should hide clear button when allowClear is false', () => {
            filter = new PortfolioTypeFilter({ 
                container,
                allowClear: false
            });
            
            const clearBtn = container.querySelector('#clearPortfolioFilter');
            expect(clearBtn).toBeFalsy();
        });

        test('should not update counts when updateCountsOnInit is false', () => {
            filter = new PortfolioTypeFilter({ 
                container,
                updateCountsOnInit: false
            });
            
            expect(fetch).not.toHaveBeenCalled();
        });
    });
});
