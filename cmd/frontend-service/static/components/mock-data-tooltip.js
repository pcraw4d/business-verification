/**
 * Mock Data Tooltip Component
 * Provides subtle, informative tooltips for cards displaying mock data
 * Auto-places tooltips to avoid viewport edges
 */
class MockDataTooltip {
    constructor(options = {}) {
        this.options = {
            placement: options.placement || 'auto', // auto, top, bottom, left, right
            offset: options.offset || 8,
            maxWidth: options.maxWidth || 280,
            delay: options.delay || 200,
            hideDelay: options.hideDelay || 100,
            className: options.className || 'mock-data-tooltip',
            ...options
        };
        
        this.tooltipElement = null;
        this.currentTarget = null;
        this.showTimer = null;
        this.hideTimer = null;
        this.isVisible = false;
        
        this.init();
    }
    
    init() {
        this.createTooltipElement();
        this.addStyles();
    }
    
    createTooltipElement() {
        if (this.tooltipElement) return;
        
        this.tooltipElement = document.createElement('div');
        this.tooltipElement.className = this.options.className;
        this.tooltipElement.setAttribute('role', 'tooltip');
        this.tooltipElement.setAttribute('aria-hidden', 'true');
        this.tooltipElement.style.cssText = `
            position: absolute;
            z-index: 10000;
            pointer-events: none;
            opacity: 0;
            transition: opacity 0.2s ease-in-out;
            max-width: ${this.options.maxWidth}px;
        `;
        
        document.body.appendChild(this.tooltipElement);
    }
    
    addStyles() {
        if (document.getElementById('mock-data-tooltip-styles')) return;
        
        const styles = document.createElement('style');
        styles.id = 'mock-data-tooltip-styles';
        styles.textContent = `
            .mock-data-tooltip {
                background: rgba(0, 0, 0, 0.9);
                color: white;
                padding: 10px 14px;
                border-radius: 6px;
                font-size: 12px;
                line-height: 1.5;
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
                word-wrap: break-word;
            }
            
            .mock-data-tooltip::before {
                content: '';
                position: absolute;
                width: 0;
                height: 0;
                border: 6px solid transparent;
            }
            
            /* Arrow positioning for different placements */
            .mock-data-tooltip.placement-top::before {
                bottom: -12px;
                left: 50%;
                transform: translateX(-50%);
                border-top-color: rgba(0, 0, 0, 0.9);
            }
            
            .mock-data-tooltip.placement-bottom::before {
                top: -12px;
                left: 50%;
                transform: translateX(-50%);
                border-bottom-color: rgba(0, 0, 0, 0.9);
            }
            
            .mock-data-tooltip.placement-left::before {
                right: -12px;
                top: 50%;
                transform: translateY(-50%);
                border-left-color: rgba(0, 0, 0, 0.9);
            }
            
            .mock-data-tooltip.placement-right::before {
                left: -12px;
                top: 50%;
                transform: translateY(-50%);
                border-right-color: rgba(0, 0, 0, 0.9);
            }
            
            .mock-data-tooltip-header {
                font-weight: 600;
                margin-bottom: 6px;
                display: flex;
                align-items: center;
                gap: 6px;
            }
            
            .mock-data-tooltip-icon {
                font-size: 14px;
            }
            
            .mock-data-tooltip-content {
                margin: 0;
            }
            
            .mock-data-tooltip-footer {
                margin-top: 6px;
                padding-top: 6px;
                border-top: 1px solid rgba(255, 255, 255, 0.2);
                font-size: 11px;
                opacity: 0.8;
            }
            
            /* Mock data indicator icon */
            .mock-data-indicator {
                display: inline-flex;
                align-items: center;
                justify-content: center;
                width: 18px;
                height: 18px;
                border-radius: 50%;
                background: rgba(255, 193, 7, 0.2);
                color: #ffc107;
                font-size: 10px;
                cursor: help;
                margin-left: 6px;
                vertical-align: middle;
                transition: all 0.2s ease;
            }
            
            .mock-data-indicator:hover {
                background: rgba(255, 193, 7, 0.3);
                transform: scale(1.1);
            }
            
            /* Accessibility */
            @media (prefers-reduced-motion: reduce) {
                .mock-data-tooltip {
                    transition: none;
                }
                
                .mock-data-indicator {
                    transition: none;
                }
            }
        `;
        
        document.head.appendChild(styles);
    }
    
    /**
     * Show tooltip for an element
     * @param {HTMLElement} element - Target element
     * @param {string|Object} content - Tooltip content (string or object with header/content/footer)
     */
    show(element, content) {
        if (!element || !this.tooltipElement) return;
        
        // Clear any existing timers
        this.clearTimers();
        
        // Set content
        this.setContent(content);
        
        // Calculate position
        const position = this.calculatePosition(element);
        this.updatePosition(position);
        
        // Show tooltip
        this.currentTarget = element;
        this.tooltipElement.setAttribute('aria-hidden', 'false');
        this.tooltipElement.style.opacity = '1';
        this.isVisible = true;
    }
    
    /**
     * Hide tooltip
     */
    hide() {
        if (!this.tooltipElement) return;
        
        this.clearTimers();
        
        this.tooltipElement.style.opacity = '0';
        this.tooltipElement.setAttribute('aria-hidden', 'true');
        this.isVisible = false;
        this.currentTarget = null;
    }
    
    /**
     * Show tooltip with delay
     */
    showDelayed(element, content) {
        this.clearTimers();
        this.currentTarget = element;
        
        this.showTimer = setTimeout(() => {
            this.show(element, content);
        }, this.options.delay);
    }
    
    /**
     * Hide tooltip with delay
     */
    hideDelayed() {
        this.clearTimers();
        
        this.hideTimer = setTimeout(() => {
            this.hide();
        }, this.options.hideDelay);
    }
    
    /**
     * Set tooltip content
     */
    setContent(content) {
        if (!this.tooltipElement) return;
        
        if (typeof content === 'string') {
            this.tooltipElement.innerHTML = `
                <div class="mock-data-tooltip-content">${this.escapeHtml(content)}</div>
            `;
        } else if (typeof content === 'object') {
            const header = content.header ? `
                <div class="mock-data-tooltip-header">
                    <i class="fas fa-info-circle mock-data-tooltip-icon" aria-hidden="true"></i>
                    <span>${this.escapeHtml(content.header)}</span>
                </div>
            ` : '';
            
            const body = content.content ? `
                <div class="mock-data-tooltip-content">${this.escapeHtml(content.content)}</div>
            ` : '';
            
            const footer = content.footer ? `
                <div class="mock-data-tooltip-footer">${this.escapeHtml(content.footer)}</div>
            ` : '';
            
            this.tooltipElement.innerHTML = header + body + footer;
        }
    }
    
    /**
     * Calculate optimal tooltip position
     */
    calculatePosition(element) {
        const rect = element.getBoundingClientRect();
        const tooltipRect = this.tooltipElement.getBoundingClientRect();
        const viewport = {
            width: window.innerWidth,
            height: window.innerHeight
        };
        
        let placement = this.options.placement;
        
        // Auto-placement logic
        if (placement === 'auto') {
            const spaceTop = rect.top;
            const spaceBottom = viewport.height - rect.bottom;
            const spaceLeft = rect.left;
            const spaceRight = viewport.width - rect.right;
            
            // Prefer top or bottom based on available space
            if (spaceBottom >= tooltipRect.height + this.options.offset) {
                placement = 'bottom';
            } else if (spaceTop >= tooltipRect.height + this.options.offset) {
                placement = 'top';
            } else if (spaceRight >= tooltipRect.width + this.options.offset) {
                placement = 'right';
            } else if (spaceLeft >= tooltipRect.width + this.options.offset) {
                placement = 'left';
            } else {
                // Default to bottom if no space available
                placement = 'bottom';
            }
        }
        
        // Calculate position based on placement
        let x, y;
        const offset = this.options.offset;
        
        switch (placement) {
            case 'top':
                x = rect.left + (rect.width / 2) - (tooltipRect.width / 2);
                y = rect.top - tooltipRect.height - offset;
                break;
            case 'bottom':
                x = rect.left + (rect.width / 2) - (tooltipRect.width / 2);
                y = rect.bottom + offset;
                break;
            case 'left':
                x = rect.left - tooltipRect.width - offset;
                y = rect.top + (rect.height / 2) - (tooltipRect.height / 2);
                break;
            case 'right':
                x = rect.right + offset;
                y = rect.top + (rect.height / 2) - (tooltipRect.height / 2);
                break;
            default:
                x = rect.left + (rect.width / 2) - (tooltipRect.width / 2);
                y = rect.bottom + offset;
        }
        
        // Keep tooltip within viewport
        x = Math.max(10, Math.min(x, viewport.width - tooltipRect.width - 10));
        y = Math.max(10, Math.min(y, viewport.height - tooltipRect.height - 10));
        
        return { x, y, placement };
    }
    
    /**
     * Update tooltip position
     */
    updatePosition(position) {
        if (!this.tooltipElement) return;
        
        // Remove all placement classes
        this.tooltipElement.classList.remove('placement-top', 'placement-bottom', 'placement-left', 'placement-right');
        
        // Add current placement class
        this.tooltipElement.classList.add(`placement-${position.placement}`);
        
        // Set position
        this.tooltipElement.style.left = `${position.x}px`;
        this.tooltipElement.style.top = `${position.y}px`;
    }
    
    /**
     * Clear timers
     */
    clearTimers() {
        if (this.showTimer) {
            clearTimeout(this.showTimer);
            this.showTimer = null;
        }
        if (this.hideTimer) {
            clearTimeout(this.hideTimer);
            this.hideTimer = null;
        }
    }
    
    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
    
    /**
     * Attach tooltip to an element
     */
    attach(element, content, options = {}) {
        if (!element) return;
        
        // Check if indicator already exists to avoid duplicates
        const existingIndicator = element.querySelector('.mock-data-indicator');
        if (existingIndicator) {
            // Update existing indicator's event listeners
            this.updateIndicatorListeners(existingIndicator, element, content);
            return;
        }
        
        const indicator = this.createIndicator();
        
        // Try to insert after the title text, or append to element
        const titleText = element.querySelector('h3, h2, h4');
        if (titleText) {
            // Insert indicator after title text
            titleText.appendChild(indicator);
        } else {
            // Fallback: append to element
            element.appendChild(indicator);
        }
        
        // Add event listeners
        this.setupIndicatorListeners(indicator, element, content);
    }
    
    /**
     * Setup event listeners for indicator
     */
    setupIndicatorListeners(indicator, element, content) {
        indicator.addEventListener('mouseenter', () => {
            this.showDelayed(element, content);
        });
        
        indicator.addEventListener('mouseleave', () => {
            this.hideDelayed();
        });
        
        indicator.addEventListener('focus', () => {
            this.showDelayed(element, content);
        });
        
        indicator.addEventListener('blur', () => {
            this.hideDelayed();
        });
        
        // Prevent tooltip from closing when moving from indicator to tooltip
        indicator.addEventListener('mouseenter', (e) => {
            e.stopPropagation();
        });
    }
    
    /**
     * Update existing indicator's event listeners
     */
    updateIndicatorListeners(indicator, element, content) {
        // Remove old listeners by cloning the element
        const newIndicator = indicator.cloneNode(true);
        indicator.parentNode.replaceChild(newIndicator, indicator);
        
        // Setup new listeners
        this.setupIndicatorListeners(newIndicator, element, content);
    }
    
    /**
     * Create mock data indicator icon
     */
    createIndicator() {
        const indicator = document.createElement('span');
        indicator.className = 'mock-data-indicator';
        indicator.setAttribute('role', 'button');
        indicator.setAttribute('aria-label', 'Mock data indicator - hover for details');
        indicator.setAttribute('tabindex', '0');
        indicator.innerHTML = '<i class="fas fa-info-circle" aria-hidden="true"></i>';
        return indicator;
    }
    
    /**
     * Destroy tooltip instance
     */
    destroy() {
        this.hide();
        this.clearTimers();
        if (this.tooltipElement) {
            this.tooltipElement.remove();
            this.tooltipElement = null;
        }
    }
}

// Create global instance
if (typeof window !== 'undefined') {
    window.MockDataTooltip = MockDataTooltip;
    window.mockDataTooltip = new MockDataTooltip();
}

