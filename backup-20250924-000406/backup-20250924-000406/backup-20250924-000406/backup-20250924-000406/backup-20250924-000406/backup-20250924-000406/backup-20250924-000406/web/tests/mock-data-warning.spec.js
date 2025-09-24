/**
 * Playwright Tests for Mock Data Warning Component
 * Tests mock data warning functionality, data source information, and user interactions
 */

const { test, expect } = require('@playwright/test');

test.describe('Mock Data Warning Component', () => {
    let page;

    test.beforeEach(async ({ browser }) => {
        page = await browser.newPage();
        
        // Mock the component HTML
        await page.setContent(`
            <!DOCTYPE html>
            <html>
            <head>
                <title>Mock Data Warning Test</title>
                <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
            </head>
            <body>
                <div id="mockDataWarningContainer"></div>
                <script src="/web/components/mock-data-warning.js"></script>
                <script>
                    // Initialize the component
                    document.addEventListener('DOMContentLoaded', () => {
                        window.mockDataWarning = new MockDataWarning({
                            container: document.getElementById('mockDataWarningContainer'),
                            dataSource: 'mock',
                            warningLevel: 'info',
                            showDataSource: true,
                            showDataCount: true
                        });
                    });
                </script>
            </body>
            </html>
        `);
    });

    test.afterEach(async () => {
        await page.close();
    });

    test('should initialize with default options', async () => {
        // Wait for component to initialize
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Check if warning is present
        const warning = await page.locator('#mockDataWarning');
        await expect(warning).toBeVisible();
        
        // Check warning level class
        await expect(warning).toHaveClass(/warning-level-info/);
    });

    test('should display mock data warning content', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Check title and subtitle
        await expect(page.locator('#warningTitle')).toContainText('Mock Data Active');
        await expect(page.locator('#warningSubtitle')).toContainText('This interface is using test data for demonstration purposes');
        
        // Check data source information
        await expect(page.locator('#dataSourceValue')).toContainText('Mock Database');
    });

    test('should show data source information', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Check data source items are visible
        await expect(page.locator('#dataSourceInfo')).toBeVisible();
        await expect(page.locator('#dataCountItem')).toBeVisible();
        await expect(page.locator('#lastUpdatedItem')).toBeVisible();
        await expect(page.locator('#dataQualityItem')).toBeVisible();
    });

    test('should handle close button click', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Click close button
        await page.click('#warningCloseBtn');
        
        // Wait for animation to complete
        await page.waitForTimeout(500);
        
        // Check if warning is hidden
        await expect(page.locator('#mockDataWarning')).not.toBeVisible();
    });

    test('should handle dismiss button click', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Click dismiss button
        await page.click('#dismissWarningBtn');
        
        // Wait for animation to complete
        await page.waitForTimeout(500);
        
        // Check if warning is hidden
        await expect(page.locator('#mockDataWarning')).not.toBeVisible();
    });

    test('should handle view data source button click', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Click view data source button
        await page.click('#viewDataSourceBtn');
        
        // Check if notification appears
        await expect(page.locator('.notification')).toBeVisible();
    });

    test('should support keyboard shortcuts', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Press Escape key
        await page.keyboard.press('Escape');
        
        // Wait for animation to complete
        await page.waitForTimeout(500);
        
        // Check if warning is hidden
        await expect(page.locator('#mockDataWarning')).not.toBeVisible();
    });

    test('should apply different warning levels', async () => {
        // Test error level
        await page.evaluate(() => {
            window.mockDataWarning.setWarningLevel('error');
        });
        
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        await expect(page.locator('#mockDataWarning')).toHaveClass(/warning-level-error/);
        
        // Test warning level
        await page.evaluate(() => {
            window.mockDataWarning.setWarningLevel('warning');
        });
        
        await expect(page.locator('#mockDataWarning')).toHaveClass(/warning-level-warning/);
        
        // Test info level
        await page.evaluate(() => {
            window.mockDataWarning.setWarningLevel('info');
        });
        
        await expect(page.locator('#mockDataWarning')).toHaveClass(/warning-level-info/);
    });

    test('should handle different data sources', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Test staging data source
        await page.evaluate(() => {
            window.mockDataWarning.setDataSource('staging');
        });
        
        await expect(page.locator('#warningTitle')).toContainText('Staging Data Active');
        await expect(page.locator('#warningSubtitle')).toContainText('This interface is using staging environment data');
        
        // Test mock data source
        await page.evaluate(() => {
            window.mockDataWarning.setDataSource('mock');
        });
        
        await expect(page.locator('#warningTitle')).toContainText('Mock Data Active');
        await expect(page.locator('#warningSubtitle')).toContainText('This interface is using test data for demonstration purposes');
    });

    test('should toggle visibility', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Initially visible
        await expect(page.locator('#mockDataWarning')).toBeVisible();
        
        // Toggle to hide
        await page.evaluate(() => {
            window.mockDataWarning.toggle();
        });
        
        await page.waitForTimeout(500);
        await expect(page.locator('#mockDataWarning')).not.toBeVisible();
        
        // Toggle to show
        await page.evaluate(() => {
            window.mockDataWarning.toggle();
        });
        
        await expect(page.locator('#mockDataWarning')).toBeVisible();
    });

    test('should handle auto-hide functionality', async () => {
        // Set auto-hide with short delay
        await page.evaluate(() => {
            window.mockDataWarning.setAutoHide(true, 1000);
            window.mockDataWarning.show();
        });
        
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        await expect(page.locator('#mockDataWarning')).toBeVisible();
        
        // Wait for auto-hide
        await page.waitForTimeout(1500);
        
        // Check if warning is hidden
        await expect(page.locator('#mockDataWarning')).not.toBeVisible();
    });

    test('should be responsive on mobile', async () => {
        // Set mobile viewport
        await page.setViewportSize({ width: 375, height: 667 });
        
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Check if warning is still visible and properly styled
        await expect(page.locator('#mockDataWarning')).toBeVisible();
        
        // Check if buttons are stacked vertically on mobile
        const actionsBottom = page.locator('.warning-actions-bottom');
        const computedStyle = await actionsBottom.evaluate(el => {
            return window.getComputedStyle(el).flexDirection;
        });
        
        // On mobile, buttons should be stacked vertically
        expect(computedStyle).toBe('column');
    });

    test('should support accessibility features', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Check for proper ARIA attributes
        const closeBtn = page.locator('#warningCloseBtn');
        await expect(closeBtn).toHaveAttribute('title', 'Dismiss warning');
        
        // Check if warning is focusable
        await closeBtn.focus();
        await expect(closeBtn).toBeFocused();
        
        // Test tab navigation
        await page.keyboard.press('Tab');
        const viewBtn = page.locator('#viewDataSourceBtn');
        await expect(viewBtn).toBeFocused();
    });

    test('should handle data source info updates', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Mock data source info update
        await page.evaluate(() => {
            window.mockDataWarning.dataInfo = {
                source: 'mock',
                type: 'test_data',
                count: 10000,
                last_updated: new Date().toISOString(),
                quality: 'high',
                description: 'Updated test data'
            };
            window.mockDataWarning.updateWarningContent();
        });
        
        // Check if data count is updated
        await expect(page.locator('#dataCountValue')).toContainText('10.0K records');
        await expect(page.locator('#dataQualityValue')).toContainText('High Quality');
    });

    test('should show notification on actions', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Click view data source button
        await page.click('#viewDataSourceBtn');
        
        // Check if notification appears
        const notification = page.locator('.notification');
        await expect(notification).toBeVisible();
        await expect(notification).toContainText('Data Source Details');
        
        // Wait for notification to disappear
        await page.waitForTimeout(6000);
        await expect(notification).not.toBeVisible();
    });

    test('should handle component destruction', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Destroy component
        await page.evaluate(() => {
            window.mockDataWarning.destroy();
        });
        
        // Check if component is removed
        await expect(page.locator('#mockDataWarning')).not.toBeVisible();
        await expect(page.locator('#mockDataWarningContainer')).toBeEmpty();
    });

    test('should format data correctly', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Test data formatting functions
        const formatResults = await page.evaluate(() => {
            return {
                dataSource: window.mockDataWarning.formatDataSource('mock'),
                dataCount: window.mockDataWarning.formatDataCount(1500),
                dataQuality: window.mockDataWarning.formatDataQuality('high')
            };
        });
        
        expect(formatResults.dataSource).toBe('Mock Database');
        expect(formatResults.dataCount).toBe('1.5K records');
        expect(formatResults.dataQuality).toBe('High Quality');
    });

    test('should handle authentication token', async () => {
        await page.waitForSelector('#mockDataWarning', { timeout: 5000 });
        
        // Test localStorage token
        await page.evaluate(() => {
            localStorage.setItem('auth_token', 'test-token');
        });
        
        const token = await page.evaluate(() => {
            return window.mockDataWarning.getAuthToken();
        });
        
        expect(token).toBe('test-token');
        
        // Test cookie token
        await page.evaluate(() => {
            localStorage.removeItem('auth_token');
            document.cookie = 'auth_token=cookie-token';
        });
        
        const cookieToken = await page.evaluate(() => {
            return window.mockDataWarning.getAuthToken();
        });
        
        expect(cookieToken).toBe('cookie-token');
    });
});
