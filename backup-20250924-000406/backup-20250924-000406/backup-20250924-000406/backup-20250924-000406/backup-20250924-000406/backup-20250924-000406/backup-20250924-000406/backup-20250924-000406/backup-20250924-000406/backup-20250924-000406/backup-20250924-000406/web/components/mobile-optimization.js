/**
 * Mobile Optimization Component
 * 
 * A comprehensive mobile optimization component that provides:
 * - Progressive enhancement for mobile devices
 * - Touch-friendly interface enhancements
 * - Accessibility compliance for mobile
 * - Performance optimization for mobile
 * 
 * This component follows the professional modular code principles and
 * ensures all new UI components are properly optimized for mobile devices.
 */

class MobileOptimization {
    constructor(options = {}) {
        this.options = {
            enableTouchOptimization: options.enableTouchOptimization !== false,
            enableProgressiveEnhancement: options.enableProgressiveEnhancement !== false,
            enableAccessibility: options.enableAccessibility !== false,
            enablePerformanceOptimization: options.enablePerformanceOptimization !== false,
            touchTargetSize: options.touchTargetSize || 44, // Minimum touch target size in pixels
            ...options
        };
        
        this.isMobile = this.detectMobile();
        this.isTouch = this.detectTouch();
        this.isInitialized = false;
    }

    /**
     * Initialize mobile optimization
     */
    init() {
        if (this.isInitialized) {
            return;
        }

        this.addMobileOptimizationStyles();
        this.enhanceTouchInteractions();
        this.improveAccessibility();
        this.optimizePerformance();
        this.addProgressiveEnhancement();
        
        this.isInitialized = true;
    }

    /**
     * Detect if device is mobile
     */
    detectMobile() {
        return window.innerWidth <= 768 || /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    }

    /**
     * Detect if device supports touch
     */
    detectTouch() {
        return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    }

    /**
     * Add comprehensive mobile optimization styles
     */
    addMobileOptimizationStyles() {
        const style = document.createElement('style');
        style.textContent = `
            /* Mobile Optimization Base Styles */
            .mobile-optimized {
                -webkit-text-size-adjust: 100%;
                -ms-text-size-adjust: 100%;
                text-size-adjust: 100%;
            }

            /* Enhanced Touch Targets */
            .touch-target {
                min-width: ${this.options.touchTargetSize}px;
                min-height: ${this.options.touchTargetSize}px;
                touch-action: manipulation;
                -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
                cursor: pointer;
            }

            .touch-target:active {
                transform: scale(0.98);
                transition: transform 0.1s ease;
            }

            /* Enhanced Form Elements for Mobile */
            .mobile-form input,
            .mobile-form textarea,
            .mobile-form select {
                font-size: 16px; /* Prevents zoom on iOS */
                padding: 12px 16px;
                min-height: ${this.options.touchTargetSize}px;
                border-radius: 8px;
                border: 1px solid #ddd;
                width: 100%;
                box-sizing: border-box;
            }

            .mobile-form input:focus,
            .mobile-form textarea:focus,
            .mobile-form select:focus {
                outline: 2px solid #3498db;
                outline-offset: 2px;
                border-color: #3498db;
            }

            /* Enhanced Button Styles for Mobile */
            .mobile-btn {
                min-height: ${this.options.touchTargetSize}px;
                padding: 12px 20px;
                font-size: 16px;
                border-radius: 8px;
                border: none;
                background: #3498db;
                color: white;
                touch-action: manipulation;
                -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
                cursor: pointer;
                transition: all 0.2s ease;
            }

            .mobile-btn:active {
                transform: scale(0.98);
                background: #2980b9;
            }

            .mobile-btn:focus {
                outline: 2px solid #3498db;
                outline-offset: 2px;
            }

            /* Enhanced Card Styles for Mobile */
            .mobile-card {
                padding: 16px;
                margin: 8px 0;
                border-radius: 12px;
                background: white;
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
                touch-action: manipulation;
            }

            .mobile-card:active {
                transform: scale(0.99);
                transition: transform 0.1s ease;
            }

            /* Enhanced Grid Layouts for Mobile */
            .mobile-grid {
                display: grid;
                gap: 16px;
                padding: 16px;
            }

            .mobile-grid.auto-fit {
                grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            }

            .mobile-grid.single-column {
                grid-template-columns: 1fr;
            }

            /* Enhanced Typography for Mobile */
            .mobile-text {
                line-height: 1.5;
                font-size: 16px;
            }

            .mobile-text.small {
                font-size: 14px;
            }

            .mobile-text.large {
                font-size: 18px;
            }

            /* Enhanced Accessibility for Mobile */
            .sr-only-mobile {
                position: absolute;
                width: 1px;
                height: 1px;
                padding: 0;
                margin: -1px;
                overflow: hidden;
                clip: rect(0, 0, 0, 0);
                white-space: nowrap;
                border: 0;
            }

            .focus-visible {
                outline: 2px solid #3498db;
                outline-offset: 2px;
            }

            /* Enhanced Loading States for Mobile */
            .mobile-loading {
                display: flex;
                align-items: center;
                justify-content: center;
                min-height: 100px;
                font-size: 16px;
                color: #666;
            }

            .mobile-loading::before {
                content: '';
                width: 20px;
                height: 20px;
                border: 2px solid #f3f3f3;
                border-top: 2px solid #3498db;
                border-radius: 50%;
                animation: mobile-spin 1s linear infinite;
                margin-right: 12px;
            }

            @keyframes mobile-spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }

            /* Enhanced Error States for Mobile */
            .mobile-error {
                padding: 16px;
                background: #ffebee;
                color: #c62828;
                border-radius: 8px;
                border-left: 4px solid #c62828;
                margin: 8px 0;
            }

            .mobile-error .error-title {
                font-weight: bold;
                margin-bottom: 8px;
            }

            .mobile-error .error-message {
                font-size: 14px;
                line-height: 1.4;
            }

            /* Enhanced Success States for Mobile */
            .mobile-success {
                padding: 16px;
                background: #e8f5e8;
                color: #2e7d32;
                border-radius: 8px;
                border-left: 4px solid #2e7d32;
                margin: 8px 0;
            }

            .mobile-success .success-title {
                font-weight: bold;
                margin-bottom: 8px;
            }

            .mobile-success .success-message {
                font-size: 14px;
                line-height: 1.4;
            }

            /* Responsive Breakpoints */
            @media (max-width: 768px) {
                .mobile-grid.auto-fit {
                    grid-template-columns: 1fr;
                }
                
                .mobile-card {
                    padding: 12px;
                    margin: 6px 0;
                }
                
                .mobile-text {
                    font-size: 15px;
                }
            }

            @media (max-width: 480px) {
                .mobile-card {
                    padding: 10px;
                    margin: 4px 0;
                }
                
                .mobile-text {
                    font-size: 14px;
                }
                
                .mobile-text.large {
                    font-size: 16px;
                }
            }

            /* Landscape Mobile Optimization */
            @media (max-width: 768px) and (orientation: landscape) {
                .mobile-grid.auto-fit {
                    grid-template-columns: repeat(2, 1fr);
                }
            }
        `;
        
        document.head.appendChild(style);
    }

    /**
     * Enhance touch interactions
     */
    enhanceTouchInteractions() {
        if (!this.isTouch) return;

        // Add touch-friendly classes to interactive elements
        const interactiveElements = document.querySelectorAll('button, a, input[type="button"], input[type="submit"], .clickable, .interactive');
        
        interactiveElements.forEach(element => {
            element.classList.add('touch-target');
            
            // Add touch event listeners for better feedback
            element.addEventListener('touchstart', (e) => {
                element.classList.add('touching');
            }, { passive: true });
            
            element.addEventListener('touchend', (e) => {
                element.classList.remove('touching');
            }, { passive: true });
        });

        // Add touch-friendly styles for touching state
        const touchStyle = document.createElement('style');
        touchStyle.textContent = `
            .touching {
                transform: scale(0.98);
                transition: transform 0.1s ease;
            }
        `;
        document.head.appendChild(touchStyle);
    }

    /**
     * Improve accessibility for mobile
     */
    improveAccessibility() {
        // Add ARIA labels to interactive elements without labels
        const interactiveElements = document.querySelectorAll('button:not([aria-label]):not([aria-labelledby]), a:not([aria-label]):not([aria-labelledby])');
        
        interactiveElements.forEach(element => {
            if (!element.getAttribute('aria-label') && !element.textContent.trim()) {
                const icon = element.querySelector('i[class*="fa-"]');
                if (icon) {
                    const iconClass = icon.className.match(/fa-([a-z-]+)/);
                    if (iconClass) {
                        element.setAttribute('aria-label', iconClass[1].replace('-', ' '));
                    }
                }
            }
        });

        // Add focus management for mobile
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                document.body.classList.add('keyboard-navigation');
            }
        });

        document.addEventListener('mousedown', () => {
            document.body.classList.remove('keyboard-navigation');
        });

        // Add keyboard navigation styles
        const keyboardStyle = document.createElement('style');
        keyboardStyle.textContent = `
            .keyboard-navigation *:focus {
                outline: 2px solid #3498db;
                outline-offset: 2px;
            }
        `;
        document.head.appendChild(keyboardStyle);
    }

    /**
     * Optimize performance for mobile
     */
    optimizePerformance() {
        // Add performance optimization styles
        const performanceStyle = document.createElement('style');
        performanceStyle.textContent = `
            /* Performance optimizations for mobile */
            .mobile-optimized * {
                -webkit-backface-visibility: hidden;
                backface-visibility: hidden;
                -webkit-perspective: 1000;
                perspective: 1000;
            }
            
            .mobile-optimized .animated {
                will-change: transform;
            }
            
            .mobile-optimized .no-animation {
                animation: none !important;
                transition: none !important;
            }
        `;
        document.head.appendChild(performanceStyle);

        // Add reduced motion support
        if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
            document.body.classList.add('no-animation');
        }
    }

    /**
     * Add progressive enhancement
     */
    addProgressiveEnhancement() {
        // Add progressive enhancement classes
        document.body.classList.add('mobile-optimized');
        
        if (this.isMobile) {
            document.body.classList.add('mobile-device');
        }
        
        if (this.isTouch) {
            document.body.classList.add('touch-device');
        }

        // Add viewport meta tag if not present
        if (!document.querySelector('meta[name="viewport"]')) {
            const viewport = document.createElement('meta');
            viewport.name = 'viewport';
            viewport.content = 'width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no';
            document.head.appendChild(viewport);
        }
    }

    /**
     * Apply mobile optimization to a specific element
     */
    optimizeElement(element, options = {}) {
        const defaultOptions = {
            addTouchTarget: true,
            addMobileClasses: true,
            improveAccessibility: true,
            ...options
        };

        if (defaultOptions.addTouchTarget && this.isTouch) {
            element.classList.add('touch-target');
        }

        if (defaultOptions.addMobileClasses) {
            element.classList.add('mobile-optimized');
        }

        if (defaultOptions.improveAccessibility) {
            this.improveElementAccessibility(element);
        }
    }

    /**
     * Improve accessibility for a specific element
     */
    improveElementAccessibility(element) {
        // Add ARIA attributes if missing
        if (element.tagName === 'BUTTON' && !element.getAttribute('aria-label') && !element.textContent.trim()) {
            const icon = element.querySelector('i[class*="fa-"]');
            if (icon) {
                const iconClass = icon.className.match(/fa-([a-z-]+)/);
                if (iconClass) {
                    element.setAttribute('aria-label', iconClass[1].replace('-', ' '));
                }
            }
        }

        // Add role if missing for interactive elements
        if (element.classList.contains('clickable') && !element.getAttribute('role')) {
            element.setAttribute('role', 'button');
        }
    }

    /**
     * Create a mobile-optimized component
     */
    createMobileComponent(tagName, options = {}) {
        const element = document.createElement(tagName);
        
        // Apply default mobile classes
        element.classList.add('mobile-optimized');
        
        if (options.classes) {
            element.classList.add(...options.classes);
        }
        
        if (options.attributes) {
            Object.entries(options.attributes).forEach(([key, value]) => {
                element.setAttribute(key, value);
            });
        }
        
        if (options.content) {
            element.innerHTML = options.content;
        }
        
        // Optimize the element
        this.optimizeElement(element, options.optimization);
        
        return element;
    }

    /**
     * Get mobile optimization status
     */
    getStatus() {
        return {
            isMobile: this.isMobile,
            isTouch: this.isTouch,
            isInitialized: this.isInitialized,
            touchTargetSize: this.options.touchTargetSize,
            features: {
                touchOptimization: this.options.enableTouchOptimization,
                progressiveEnhancement: this.options.enableProgressiveEnhancement,
                accessibility: this.options.enableAccessibility,
                performanceOptimization: this.options.enablePerformanceOptimization
            }
        };
    }
}

// Auto-initialize if in browser environment
if (typeof window !== 'undefined') {
    // Initialize mobile optimization when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.mobileOptimization = new MobileOptimization();
            window.mobileOptimization.init();
        });
    } else {
        window.mobileOptimization = new MobileOptimization();
        window.mobileOptimization.init();
    }
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MobileOptimization;
}
