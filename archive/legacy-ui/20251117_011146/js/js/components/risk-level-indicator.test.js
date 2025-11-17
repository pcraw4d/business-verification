/**
 * Unit Tests for Risk Level Indicator Component
 * Tests risk level visualization, filtering, and interaction functionality
 */

// Mock DOM environment for testing
const { JSDOM } = require('jsdom');
const fs = require('fs');
const path = require('path');

// Setup DOM environment
const dom = new JSDOM(`
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
</head>
<body>
    <div id="testContainer"></div>
</body>
</html>
`, {
    url: 'http://localhost',
    pretendToBeVisual: true,
    resources: 'usable'
});

global.window = dom.window;
global.document = dom.window.document;
global.navigator = dom.window.navigator;

// Load the component
const componentPath = path.join(__dirname, 'risk-level-indicator.js');
const componentCode = fs.readFileSync(componentPath, 'utf8');
eval(componentCode);

describe('RiskLevelIndicator Component', () => {
    let container;
    let indicator;

    beforeEach(() => {
        // Create fresh container for each test
        container = document.getElementById('testContainer');
        container.innerHTML = '';
        
        // Reset any global state
        window.riskLevelIndicator = null;
    });

    afterEach(() => {
        if (indicator) {
            indicator.destroy();
        }
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            indicator = new RiskLevelIndicator({
                container: container
            });

            expect(indicator.mode).toBe('single');
            expect(indicator.allowClear).toBe(true);
            expect(indicator.showCounts).toBe(true);
            expect(indicator.allowAll).toBe(true);
            expect(indicator.compactMode).toBe(false);
            expect(indicator.selectedValues).toBe(null);
        });

        test('should initialize with custom options', () => {
            indicator = new RiskLevelIndicator({
                container: container,
                mode: 'multiple',
                allowClear: false,
                showCounts: false,
                compactMode: true,
                initialValue: 'low'
            });

            expect(indicator.mode).toBe('multiple');
            expect(indicator.allowClear).toBe(false);
            expect(indicator.showCounts).toBe(false);
            expect(indicator.compactMode).toBe(true);
            expect(indicator.selectedValues).toEqual(['low']);
        });

        test('should create proper DOM structure', () => {
            indicator = new RiskLevelIndicator({
                container: container
            });

            expect(container.querySelector('.risk-level-indicator-container')).toBeTruthy();
            expect(container.querySelector('.indicator-header')).toBeTruthy();
            expect(container.querySelector('.indicator-label')).toBeTruthy();
            expect(container.querySelector('.risk-indicator-dropdown')).toBeTruthy();
        });

        test('should create compact view when compactMode is true', () => {
            indicator = new RiskLevelIndicator({
                container: container,
                compactMode: true
            });

            expect(container.querySelector('.risk-level-badges')).toBeTruthy();
            expect(container.querySelector('.risk-badge')).toBeTruthy();
            expect(container.querySelector('.risk-indicator-dropdown')).toBeFalsy();
        });
    });

    describe('Risk Level Definitions', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container
            });
        });

        test('should have all required risk levels', () => {
            expect(indicator.riskLevels.low).toBeDefined();
            expect(indicator.riskLevels.medium).toBeDefined();
            expect(indicator.riskLevels.high).toBeDefined();
        });

        test('should have proper risk level properties', () => {
            const lowRisk = indicator.riskLevels.low;
            expect(lowRisk.label).toBe('Low Risk');
            expect(lowRisk.description).toBe('Minimal risk factors identified');
            expect(lowRisk.icon).toBe('fas fa-shield-alt');
            expect(lowRisk.color).toBe('#27ae60');
            expect(lowRisk.priority).toBe(1);
        });

        test('should have proper color coding', () => {
            expect(indicator.riskLevels.low.color).toBe('#27ae60'); // Green
            expect(indicator.riskLevels.medium.color).toBe('#f39c12'); // Orange
            expect(indicator.riskLevels.high.color).toBe('#e74c3c'); // Red
        });
    });

    describe('Single Selection Mode', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container,
                mode: 'single'
            });
        });

        test('should select single risk level', () => {
            indicator.selectOption('medium');
            expect(indicator.selectedValues).toBe('medium');
        });

        test('should clear selection when selecting empty value', () => {
            indicator.selectOption('medium');
            indicator.selectOption('');
            expect(indicator.selectedValues).toBe(null);
        });

        test('should update UI when selection changes', () => {
            const onChange = jest.fn();
            indicator.onChange = onChange;

            indicator.selectOption('high');
            expect(onChange).toHaveBeenCalledWith('high');
        });

        test('should close dropdown after selection in single mode', () => {
            indicator.toggleDropdown();
            expect(indicator.isExpanded).toBe(true);

            indicator.selectOption('low');
            expect(indicator.isExpanded).toBe(false);
        });
    });

    describe('Multiple Selection Mode', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container,
                mode: 'multiple'
            });
        });

        test('should select multiple risk levels', () => {
            indicator.selectOption('low');
            indicator.selectOption('high');
            expect(indicator.selectedValues).toEqual(['low', 'high']);
        });

        test('should toggle selection when clicking same option', () => {
            indicator.selectOption('medium');
            expect(indicator.selectedValues).toEqual(['medium']);

            indicator.selectOption('medium');
            expect(indicator.selectedValues).toEqual([]);
        });

        test('should select all risk levels', () => {
            indicator.selectAll();
            expect(indicator.selectedValues).toEqual(['low', 'medium', 'high']);
        });

        test('should clear all selections', () => {
            indicator.selectOption('low');
            indicator.selectOption('high');
            indicator.clearAll();
            expect(indicator.selectedValues).toEqual([]);
        });

        test('should not close dropdown after selection in multiple mode', () => {
            indicator.toggleDropdown();
            expect(indicator.isExpanded).toBe(true);

            indicator.selectOption('low');
            expect(indicator.isExpanded).toBe(true);
        });
    });

    describe('UI Updates', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container
            });
        });

        test('should update trigger text for single selection', () => {
            indicator.selectOption('medium');
            const triggerText = container.querySelector('#riskIndicatorText');
            expect(triggerText.textContent).toBe('Medium Risk');
        });

        test('should update trigger text for multiple selection', () => {
            indicator = new RiskLevelIndicator({
                container: container,
                mode: 'multiple'
            });
            indicator.selectOption('low');
            indicator.selectOption('high');

            const triggerText = container.querySelector('#riskIndicatorText');
            expect(triggerText.textContent).toBe('2 selected');
        });

        test('should show clear button when selection exists', () => {
            indicator.selectOption('low');
            const clearBtn = container.querySelector('#clearRiskIndicator');
            expect(clearBtn.style.display).toBe('block');
        });

        test('should hide clear button when no selection', () => {
            const clearBtn = container.querySelector('#clearRiskIndicator');
            expect(clearBtn.style.display).toBe('none');
        });

        test('should update option states in dropdown', () => {
            indicator.toggleDropdown();
            indicator.selectOption('high');

            const highOption = container.querySelector('[data-value="high"]');
            expect(highOption.classList.contains('selected')).toBe(true);
        });
    });

    describe('Compact Mode', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container,
                compactMode: true
            });
        });

        test('should create risk badges', () => {
            const badges = container.querySelectorAll('.risk-badge');
            expect(badges.length).toBe(3); // low, medium, high
        });

        test('should select badge when clicked', () => {
            const lowBadge = container.querySelector('[data-value="low"]');
            lowBadge.click();

            expect(indicator.selectedValues).toBe('low');
            expect(lowBadge.classList.contains('selected')).toBe(true);
        });

        test('should update badge states', () => {
            indicator.selectOption('medium');
            const mediumBadge = container.querySelector('[data-value="medium"]');
            expect(mediumBadge.classList.contains('selected')).toBe(true);
        });
    });

    describe('Event Handling', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container
            });
        });

        test('should handle dropdown toggle', () => {
            const trigger = container.querySelector('#riskIndicatorTrigger');
            trigger.click();

            expect(indicator.isExpanded).toBe(true);
            expect(container.querySelector('.risk-indicator-dropdown').classList.contains('open')).toBe(true);
        });

        test('should close dropdown when clicking outside', () => {
            indicator.toggleDropdown();
            expect(indicator.isExpanded).toBe(true);

            // Simulate click outside
            document.body.click();
            expect(indicator.isExpanded).toBe(false);
        });

        test('should handle keyboard navigation', () => {
            const trigger = container.querySelector('#riskIndicatorTrigger');
            
            // Test Enter key
            const enterEvent = new KeyboardEvent('keydown', { key: 'Enter' });
            trigger.dispatchEvent(enterEvent);
            expect(indicator.isExpanded).toBe(true);

            // Test Escape key
            const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
            trigger.dispatchEvent(escapeEvent);
            expect(indicator.isExpanded).toBe(false);
        });

        test('should trigger risk level click callback', () => {
            const onRiskLevelClick = jest.fn();
            indicator.onRiskLevelClick = onRiskLevelClick;

            indicator.selectOption('high');
            expect(onRiskLevelClick).toHaveBeenCalledWith({
                value: 'high',
                level: indicator.riskLevels.high,
                selected: true
            });
        });
    });

    describe('API Integration', () => {
        beforeEach(() => {
            // Mock fetch
            global.fetch = jest.fn();
            indicator = new RiskLevelIndicator({
                container: container,
                updateCountsOnInit: true
            });
        });

        afterEach(() => {
            global.fetch.mockRestore();
        });

        test('should update counts from API', async () => {
            const mockResponse = {
                ok: true,
                json: () => Promise.resolve({
                    risk_levels: {
                        low: 150,
                        medium: 75,
                        high: 25
                    },
                    total: 250
                })
            };

            global.fetch.mockResolvedValue(mockResponse);

            await indicator.updateCounts();

            expect(indicator.riskLevels.low.count).toBe(150);
            expect(indicator.riskLevels.medium.count).toBe(75);
            expect(indicator.riskLevels.high.count).toBe(25);
        });

        test('should handle API errors gracefully', async () => {
            global.fetch.mockRejectedValue(new Error('API Error'));

            // Should not throw
            await expect(indicator.updateCounts()).resolves.toBeUndefined();
        });
    });

    describe('Public Methods', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container
            });
        });

        test('should get current value', () => {
            indicator.selectOption('medium');
            expect(indicator.getValue()).toBe('medium');
        });

        test('should set value programmatically', () => {
            indicator.setValue('high');
            expect(indicator.selectedValues).toBe('high');
        });

        test('should get risk level info', () => {
            const info = indicator.getRiskLevelInfo('low');
            expect(info.label).toBe('Low Risk');
            expect(info.color).toBe('#27ae60');
        });

        test('should get selected risk levels', () => {
            indicator = new RiskLevelIndicator({
                container: container,
                mode: 'multiple'
            });
            indicator.selectOption('low');
            indicator.selectOption('high');

            const selected = indicator.getSelectedRiskLevels();
            expect(selected).toHaveLength(2);
            expect(selected[0].value).toBe('low');
            expect(selected[1].value).toBe('high');
        });

        test('should set disabled state', () => {
            indicator.setDisabled(true);
            expect(indicator.disabled).toBe(true);

            const trigger = container.querySelector('#riskIndicatorTrigger');
            expect(trigger.disabled).toBe(true);
        });

        test('should reset to initial state', () => {
            indicator.selectOption('high');
            indicator.reset();
            expect(indicator.selectedValues).toBe(null);
        });

        test('should switch to compact mode', () => {
            indicator.setCompactMode(true);
            expect(indicator.compactMode).toBe(true);
            expect(container.querySelector('.risk-level-badges')).toBeTruthy();
        });
    });

    describe('Filter Integration', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container
            });
        });

        test('should trigger filter callback', () => {
            const onFilter = jest.fn();
            indicator.onFilter = onFilter;

            indicator.selectOption('medium');
            expect(onFilter).toHaveBeenCalledWith({
                risk_level: 'medium',
                mode: 'single'
            });
        });

        test('should trigger change callback', () => {
            const onChange = jest.fn();
            indicator.onChange = onChange;

            indicator.selectOption('high');
            expect(onChange).toHaveBeenCalledWith('high');
        });

        test('should trigger clear callback', () => {
            const onClear = jest.fn();
            indicator.onClear = onClear;

            indicator.selectOption('low');
            indicator.clearSelection();
            expect(onClear).toHaveBeenCalled();
        });
    });

    describe('Accessibility', () => {
        beforeEach(() => {
            indicator = new RiskLevelIndicator({
                container: container
            });
        });

        test('should have proper ARIA attributes', () => {
            const trigger = container.querySelector('#riskIndicatorTrigger');
            expect(trigger).toBeTruthy();
        });

        test('should support keyboard navigation', () => {
            const trigger = container.querySelector('#riskIndicatorTrigger');
            
            // Focus should work
            trigger.focus();
            expect(document.activeElement).toBe(trigger);
        });

        test('should have proper labels', () => {
            const label = container.querySelector('.indicator-label');
            expect(label.textContent).toContain('Risk Level');
        });
    });

    describe('Error Handling', () => {
        test('should handle invalid risk level values', () => {
            indicator = new RiskLevelIndicator({
                container: container
            });

            // Should not throw when selecting invalid value
            expect(() => indicator.selectOption('invalid')).not.toThrow();
        });

        test('should handle missing container', () => {
            // Should not throw when container is null
            expect(() => new RiskLevelIndicator({ container: null })).not.toThrow();
        });
    });

    describe('Performance', () => {
        test('should handle rapid selection changes', () => {
            indicator = new RiskLevelIndicator({
                container: container,
                mode: 'multiple'
            });

            // Rapid selection changes should not cause issues
            for (let i = 0; i < 100; i++) {
                indicator.selectOption('low');
                indicator.selectOption('medium');
                indicator.selectOption('high');
            }

            expect(indicator.selectedValues).toEqual(['low', 'medium', 'high']);
        });
    });
});

// Run tests if this file is executed directly
if (require.main === module) {
    const { runTests } = require('jest');
    runTests();
}
