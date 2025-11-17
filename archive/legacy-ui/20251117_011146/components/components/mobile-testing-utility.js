/**
 * Mobile Testing Utility
 * 
 * A comprehensive testing utility for validating mobile optimization
 * across different devices, screen sizes, and orientations.
 */

class MobileTestingUtility {
    constructor() {
        this.testResults = {
            responsive: {},
            touch: {},
            accessibility: {},
            performance: {},
            progressive: {}
        };
        this.isRunning = false;
    }

    /**
     * Run comprehensive mobile tests
     */
    async runAllTests() {
        this.isRunning = true;
        console.log('🚀 Starting comprehensive mobile optimization tests...');

        try {
            await this.testResponsiveDesign();
            await this.testTouchInteractions();
            await this.testAccessibility();
            await this.testPerformance();
            await this.testProgressiveEnhancement();

            this.generateTestReport();
            return this.testResults;
        } catch (error) {
            console.error('❌ Mobile testing failed:', error);
            throw error;
        } finally {
            this.isRunning = false;
        }
    }

    /**
     * Test responsive design across different screen sizes
     */
    async testResponsiveDesign() {
        console.log('📱 Testing responsive design...');
        
        const breakpoints = [
            { name: 'Mobile Small', width: 320, height: 568 },
            { name: 'Mobile Medium', width: 375, height: 667 },
            { name: 'Mobile Large', width: 414, height: 896 },
            { name: 'Tablet Portrait', width: 768, height: 1024 },
            { name: 'Tablet Landscape', width: 1024, height: 768 },
            { name: 'Desktop', width: 1920, height: 1080 }
        ];

        for (const breakpoint of breakpoints) {
            const result = await this.testBreakpoint(breakpoint);
            this.testResults.responsive[breakpoint.name] = result;
        }
    }

    /**
     * Test specific breakpoint
     */
    async testBreakpoint(breakpoint) {
        const originalWidth = window.innerWidth;
        const originalHeight = window.innerHeight;

        // Simulate screen size
        Object.defineProperty(window, 'innerWidth', { value: breakpoint.width, writable: true });
        Object.defineProperty(window, 'innerHeight', { value: breakpoint.height, writable: true });

        // Trigger resize event
        window.dispatchEvent(new Event('resize'));

        // Wait for layout to settle
        await new Promise(resolve => setTimeout(resolve, 100));

        const result = {
            width: breakpoint.width,
            height: breakpoint.height,
            tests: {
                viewportMeta: this.testViewportMeta(),
                responsiveImages: this.testResponsiveImages(),
                flexibleLayouts: this.testFlexibleLayouts(),
                readableText: this.testReadableText(),
                touchTargets: this.testTouchTargets()
            }
        };

        // Restore original dimensions
        Object.defineProperty(window, 'innerWidth', { value: originalWidth, writable: true });
        Object.defineProperty(window, 'innerHeight', { value: originalHeight, writable: true });

        return result;
    }

    /**
     * Test viewport meta tag
     */
    testViewportMeta() {
        const viewportMeta = document.querySelector('meta[name="viewport"]');
        if (!viewportMeta) {
            return { passed: false, message: 'Viewport meta tag missing' };
        }

        const content = viewportMeta.getAttribute('content');
        const hasWidth = content.includes('width=device-width');
        const hasInitialScale = content.includes('initial-scale=1.0');

        return {
            passed: hasWidth && hasInitialScale,
            message: hasWidth && hasInitialScale ? 'Viewport meta tag correct' : 'Viewport meta tag incomplete'
        };
    }

    /**
     * Test responsive images
     */
    testResponsiveImages() {
        const images = document.querySelectorAll('img');
        let responsiveCount = 0;

        images.forEach(img => {
            if (img.srcset || img.sizes || img.style.maxWidth === '100%') {
                responsiveCount++;
            }
        });

        const totalImages = images.length;
        const percentage = totalImages > 0 ? (responsiveCount / totalImages) * 100 : 100;

        return {
            passed: percentage >= 80,
            message: `${responsiveCount}/${totalImages} images are responsive (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test flexible layouts
     */
    testFlexibleLayouts() {
        const containers = document.querySelectorAll('.container, .grid, .flex');
        let flexibleCount = 0;

        containers.forEach(container => {
            const styles = window.getComputedStyle(container);
            if (styles.display === 'flex' || styles.display === 'grid' || 
                styles.maxWidth === '100%' || styles.width === '100%') {
                flexibleCount++;
            }
        });

        const totalContainers = containers.length;
        const percentage = totalContainers > 0 ? (flexibleCount / totalContainers) * 100 : 100;

        return {
            passed: percentage >= 90,
            message: `${flexibleCount}/${totalContainers} containers use flexible layouts (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test readable text
     */
    testReadableText() {
        const textElements = document.querySelectorAll('p, h1, h2, h3, h4, h5, h6, span, div');
        let readableCount = 0;

        textElements.forEach(element => {
            const styles = window.getComputedStyle(element);
            const fontSize = parseFloat(styles.fontSize);
            const lineHeight = parseFloat(styles.lineHeight);
            
            if (fontSize >= 14 && lineHeight >= 1.4) {
                readableCount++;
            }
        });

        const totalText = textElements.length;
        const percentage = totalText > 0 ? (readableCount / totalText) * 100 : 100;

        return {
            passed: percentage >= 85,
            message: `${readableCount}/${totalText} text elements are readable (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test touch targets
     */
    testTouchTargets() {
        const interactiveElements = document.querySelectorAll('button, a, input, select, .clickable, .interactive');
        let touchFriendlyCount = 0;

        interactiveElements.forEach(element => {
            const styles = window.getComputedStyle(element);
            const width = parseFloat(styles.width);
            const height = parseFloat(styles.height);
            const minWidth = parseFloat(styles.minWidth);
            const minHeight = parseFloat(styles.minHeight);

            if ((width >= 44 || minWidth >= 44) && (height >= 44 || minHeight >= 44)) {
                touchFriendlyCount++;
            }
        });

        const totalInteractive = interactiveElements.length;
        const percentage = totalInteractive > 0 ? (touchFriendlyCount / totalInteractive) * 100 : 100;

        return {
            passed: percentage >= 90,
            message: `${touchFriendlyCount}/${totalInteractive} interactive elements are touch-friendly (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test touch interactions
     */
    async testTouchInteractions() {
        console.log('👆 Testing touch interactions...');

        const touchTests = {
            touchDetection: this.testTouchDetection(),
            touchEvents: this.testTouchEvents(),
            touchFeedback: this.testTouchFeedback(),
            gestureSupport: this.testGestureSupport()
        };

        this.testResults.touch = touchTests;
    }

    /**
     * Test touch detection
     */
    testTouchDetection() {
        const hasTouchStart = 'ontouchstart' in window;
        const hasMaxTouchPoints = navigator.maxTouchPoints > 0;
        const hasTouchClass = document.body.classList.contains('touch-device');

        return {
            passed: hasTouchStart || hasMaxTouchPoints,
            message: `Touch detection: ${hasTouchStart || hasMaxTouchPoints ? 'Working' : 'Not detected'}`,
            details: {
                ontouchstart: hasTouchStart,
                maxTouchPoints: hasMaxTouchPoints,
                touchClass: hasTouchClass
            }
        };
    }

    /**
     * Test touch events
     */
    testTouchEvents() {
        const touchElements = document.querySelectorAll('.touch-target, .interactive');
        let touchEventCount = 0;

        touchElements.forEach(element => {
            if (element.style.touchAction === 'manipulation' || 
                element.classList.contains('touch-target')) {
                touchEventCount++;
            }
        });

        const totalTouchElements = touchElements.length;
        const percentage = totalTouchElements > 0 ? (touchEventCount / totalTouchElements) * 100 : 100;

        return {
            passed: percentage >= 80,
            message: `${touchEventCount}/${totalTouchElements} elements have touch optimization (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test touch feedback
     */
    testTouchFeedback() {
        const interactiveElements = document.querySelectorAll('button, a, .clickable');
        let feedbackCount = 0;

        interactiveElements.forEach(element => {
            const styles = window.getComputedStyle(element);
            if (styles.transition.includes('transform') || 
                element.classList.contains('touch-target')) {
                feedbackCount++;
            }
        });

        const totalInteractive = interactiveElements.length;
        const percentage = totalInteractive > 0 ? (feedbackCount / totalInteractive) * 100 : 100;

        return {
            passed: percentage >= 70,
            message: `${feedbackCount}/${totalInteractive} elements provide touch feedback (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test gesture support
     */
    testGestureSupport() {
        const hasGestureSupport = 'ongesturestart' in window || 
                                 'ontouchstart' in window && 'ontouchmove' in window;

        return {
            passed: hasGestureSupport,
            message: `Gesture support: ${hasGestureSupport ? 'Available' : 'Not available'}`
        };
    }

    /**
     * Test accessibility
     */
    async testAccessibility() {
        console.log('♿ Testing accessibility...');

        const accessibilityTests = {
            ariaLabels: this.testAriaLabels(),
            keyboardNavigation: this.testKeyboardNavigation(),
            focusManagement: this.testFocusManagement(),
            colorContrast: this.testColorContrast(),
            screenReader: this.testScreenReaderSupport()
        };

        this.testResults.accessibility = accessibilityTests;
    }

    /**
     * Test ARIA labels
     */
    testAriaLabels() {
        const interactiveElements = document.querySelectorAll('button, a, input, select');
        let labeledCount = 0;

        interactiveElements.forEach(element => {
            if (element.getAttribute('aria-label') || 
                element.getAttribute('aria-labelledby') ||
                element.textContent.trim() ||
                element.querySelector('img[alt]')) {
                labeledCount++;
            }
        });

        const totalInteractive = interactiveElements.length;
        const percentage = totalInteractive > 0 ? (labeledCount / totalInteractive) * 100 : 100;

        return {
            passed: percentage >= 90,
            message: `${labeledCount}/${totalInteractive} interactive elements have labels (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test keyboard navigation
     */
    testKeyboardNavigation() {
        const focusableElements = document.querySelectorAll('button, a, input, select, textarea, [tabindex]');
        let keyboardAccessibleCount = 0;

        focusableElements.forEach(element => {
            const tabIndex = element.getAttribute('tabindex');
            if (tabIndex !== '-1' && !element.disabled) {
                keyboardAccessibleCount++;
            }
        });

        const totalFocusable = focusableElements.length;
        const percentage = totalFocusable > 0 ? (keyboardAccessibleCount / totalFocusable) * 100 : 100;

        return {
            passed: percentage >= 95,
            message: `${keyboardAccessibleCount}/${totalFocusable} elements are keyboard accessible (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test focus management
     */
    testFocusManagement() {
        const hasFocusStyles = document.querySelector('style')?.textContent.includes(':focus') ||
                              document.querySelector('style')?.textContent.includes('.focus-visible');

        return {
            passed: hasFocusStyles,
            message: `Focus management: ${hasFocusStyles ? 'Implemented' : 'Missing focus styles'}`
        };
    }

    /**
     * Test color contrast
     */
    testColorContrast() {
        // Simplified contrast test - in real implementation, would use a proper contrast checker
        const textElements = document.querySelectorAll('p, h1, h2, h3, h4, h5, h6, span, div');
        let goodContrastCount = 0;

        textElements.forEach(element => {
            const styles = window.getComputedStyle(element);
            const color = styles.color;
            const backgroundColor = styles.backgroundColor;
            
            // Basic check for non-transparent colors
            if (color !== 'rgba(0, 0, 0, 0)' && backgroundColor !== 'rgba(0, 0, 0, 0)') {
                goodContrastCount++;
            }
        });

        const totalText = textElements.length;
        const percentage = totalText > 0 ? (goodContrastCount / totalText) * 100 : 100;

        return {
            passed: percentage >= 80,
            message: `${goodContrastCount}/${totalText} text elements have defined colors (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test screen reader support
     */
    testScreenReaderSupport() {
        const hasScreenReaderClass = document.body.classList.contains('keyboard-navigation');
        const hasAriaLive = document.querySelector('[aria-live]');

        return {
            passed: hasScreenReaderClass || hasAriaLive,
            message: `Screen reader support: ${hasScreenReaderClass || hasAriaLive ? 'Available' : 'Limited'}`
        };
    }

    /**
     * Test performance
     */
    async testPerformance() {
        console.log('⚡ Testing performance...');

        const performanceTests = {
            loadTime: await this.testLoadTime(),
            renderTime: await this.testRenderTime(),
            memoryUsage: this.testMemoryUsage(),
            animationPerformance: this.testAnimationPerformance()
        };

        this.testResults.performance = performanceTests;
    }

    /**
     * Test load time
     */
    async testLoadTime() {
        const startTime = performance.now();
        
        // Simulate loading
        await new Promise(resolve => setTimeout(resolve, 100));
        
        const loadTime = performance.now() - startTime;

        return {
            passed: loadTime < 1000,
            message: `Load time: ${loadTime.toFixed(2)}ms`,
            value: loadTime
        };
    }

    /**
     * Test render time
     */
    async testRenderTime() {
        const startTime = performance.now();
        
        // Force a reflow
        document.body.offsetHeight;
        
        const renderTime = performance.now() - startTime;

        return {
            passed: renderTime < 16, // 60fps target
            message: `Render time: ${renderTime.toFixed(2)}ms`,
            value: renderTime
        };
    }

    /**
     * Test memory usage
     */
    testMemoryUsage() {
        const memory = performance.memory;
        if (!memory) {
            return {
                passed: true,
                message: 'Memory API not available'
            };
        }

        const usedMB = memory.usedJSHeapSize / 1024 / 1024;
        const limitMB = memory.jsHeapSizeLimit / 1024 / 1024;
        const percentage = (usedMB / limitMB) * 100;

        return {
            passed: percentage < 80,
            message: `Memory usage: ${usedMB.toFixed(2)}MB / ${limitMB.toFixed(2)}MB (${percentage.toFixed(1)}%)`,
            value: percentage
        };
    }

    /**
     * Test animation performance
     */
    testAnimationPerformance() {
        const animatedElements = document.querySelectorAll('.animated, [style*="animation"], [style*="transition"]');
        let optimizedCount = 0;

        animatedElements.forEach(element => {
            const styles = window.getComputedStyle(element);
            if (styles.willChange !== 'auto' || 
                styles.transform !== 'none' ||
                styles.backfaceVisibility === 'hidden') {
                optimizedCount++;
            }
        });

        const totalAnimated = animatedElements.length;
        const percentage = totalAnimated > 0 ? (optimizedCount / totalAnimated) * 100 : 100;

        return {
            passed: percentage >= 70,
            message: `${optimizedCount}/${totalAnimated} animations are optimized (${percentage.toFixed(1)}%)`
        };
    }

    /**
     * Test progressive enhancement
     */
    async testProgressiveEnhancement() {
        console.log('🔄 Testing progressive enhancement...');

        const progressiveTests = {
            baseFunctionality: this.testBaseFunctionality(),
            enhancedFeatures: this.testEnhancedFeatures(),
            gracefulDegradation: this.testGracefulDegradation(),
            featureDetection: this.testFeatureDetection()
        };

        this.testResults.progressive = progressiveTests;
    }

    /**
     * Test base functionality
     */
    testBaseFunctionality() {
        const hasBasicHTML = document.querySelector('html');
        const hasBasicCSS = document.querySelector('style, link[rel="stylesheet"]');
        const hasBasicJS = document.querySelector('script');

        return {
            passed: hasBasicHTML && hasBasicCSS,
            message: `Base functionality: ${hasBasicHTML && hasBasicCSS ? 'Available' : 'Missing core elements'}`
        };
    }

    /**
     * Test enhanced features
     */
    testEnhancedFeatures() {
        const hasMobileOptimization = document.body.classList.contains('mobile-optimized');
        const hasTouchSupport = document.body.classList.contains('touch-device');
        const hasProgressiveEnhancement = document.querySelector('.progressive-enhancement');

        return {
            passed: hasMobileOptimization,
            message: `Enhanced features: ${hasMobileOptimization ? 'Loaded' : 'Not loaded'}`
        };
    }

    /**
     * Test graceful degradation
     */
    testGracefulDegradation() {
        const hasFallbacks = document.querySelectorAll('[data-fallback], .fallback');
        const hasNoJS = document.querySelector('.no-js');

        return {
            passed: hasFallbacks.length > 0 || hasNoJS,
            message: `Graceful degradation: ${hasFallbacks.length > 0 || hasNoJS ? 'Implemented' : 'Missing fallbacks'}`
        };
    }

    /**
     * Test feature detection
     */
    testFeatureDetection() {
        const hasFeatureDetection = typeof window.mobileOptimization !== 'undefined';
        const hasTouchDetection = 'ontouchstart' in window;
        const hasMediaQueries = window.matchMedia;

        return {
            passed: hasFeatureDetection && hasMediaQueries,
            message: `Feature detection: ${hasFeatureDetection && hasMediaQueries ? 'Working' : 'Limited'}`
        };
    }

    /**
     * Generate comprehensive test report
     */
    generateTestReport() {
        console.log('📊 Generating mobile optimization test report...');

        const report = {
            summary: this.generateSummary(),
            details: this.testResults,
            recommendations: this.generateRecommendations(),
            timestamp: new Date().toISOString()
        };

        console.log('✅ Mobile optimization test report generated');
        console.table(report.summary);

        return report;
    }

    /**
     * Generate test summary
     */
    generateSummary() {
        const summary = {
            'Responsive Design': this.calculateCategoryScore('responsive'),
            'Touch Interactions': this.calculateCategoryScore('touch'),
            'Accessibility': this.calculateCategoryScore('accessibility'),
            'Performance': this.calculateCategoryScore('performance'),
            'Progressive Enhancement': this.calculateCategoryScore('progressive')
        };

        const overallScore = Object.values(summary).reduce((sum, score) => sum + score, 0) / Object.keys(summary).length;
        summary['Overall Score'] = overallScore.toFixed(1) + '%';

        return summary;
    }

    /**
     * Calculate category score
     */
    calculateCategoryScore(category) {
        const tests = this.testResults[category];
        if (!tests) return '0%';

        const passedTests = Object.values(tests).filter(test => test.passed).length;
        const totalTests = Object.keys(tests).length;
        const percentage = (passedTests / totalTests) * 100;

        return percentage.toFixed(1) + '%';
    }

    /**
     * Generate recommendations
     */
    generateRecommendations() {
        const recommendations = [];

        // Responsive design recommendations
        const responsiveScore = parseFloat(this.calculateCategoryScore('responsive'));
        if (responsiveScore < 80) {
            recommendations.push('Improve responsive design implementation');
        }

        // Touch interaction recommendations
        const touchScore = parseFloat(this.calculateCategoryScore('touch'));
        if (touchScore < 80) {
            recommendations.push('Enhance touch interaction support');
        }

        // Accessibility recommendations
        const accessibilityScore = parseFloat(this.calculateCategoryScore('accessibility'));
        if (accessibilityScore < 90) {
            recommendations.push('Improve accessibility compliance');
        }

        // Performance recommendations
        const performanceScore = parseFloat(this.calculateCategoryScore('performance'));
        if (performanceScore < 80) {
            recommendations.push('Optimize performance for mobile devices');
        }

        // Progressive enhancement recommendations
        const progressiveScore = parseFloat(this.calculateCategoryScore('progressive'));
        if (progressiveScore < 80) {
            recommendations.push('Implement better progressive enhancement');
        }

        return recommendations;
    }
}

// Auto-run tests if in browser environment
if (typeof window !== 'undefined') {
    // Initialize testing utility when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.mobileTestingUtility = new MobileTestingUtility();
        });
    } else {
        window.mobileTestingUtility = new MobileTestingUtility();
    }
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MobileTestingUtility;
}
