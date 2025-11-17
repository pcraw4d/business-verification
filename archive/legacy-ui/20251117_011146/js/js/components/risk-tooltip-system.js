/**
 * Risk Tooltip System
 * Provides interactive tooltips for risk visualizations
 */
class RiskTooltipSystem {
    constructor() {
        this.tooltips = new Map();
        this.activeTooltip = null;
        this.tooltipElement = null;
        this.init();
    }

    /**
     * Initialize the tooltip system
     */
    init() {
        this.createTooltipElement();
        this.bindEvents();
    }

    /**
     * Create tooltip DOM element
     */
    createTooltipElement() {
        this.tooltipElement = document.createElement('div');
        this.tooltipElement.id = 'riskTooltip';
        this.tooltipElement.className = 'risk-tooltip';
        this.tooltipElement.style.cssText = `
            position: absolute;
            background: rgba(0, 0, 0, 0.9);
            color: white;
            padding: 12px 16px;
            border-radius: 6px;
            font-size: 0.875rem;
            pointer-events: none;
            z-index: 10000;
            display: none;
            max-width: 300px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        `;
        // Wait for body to be available before appending
        if (document.body) {
            document.body.appendChild(this.tooltipElement);
        } else {
            // If body not ready, wait for DOMContentLoaded
            if (document.readyState === 'loading') {
                document.addEventListener('DOMContentLoaded', () => {
                    if (document.body && this.tooltipElement) {
                        document.body.appendChild(this.tooltipElement);
                    }
                });
            } else {
                // Fallback: try again after a short delay
                setTimeout(() => {
                    if (document.body && this.tooltipElement) {
                        document.body.appendChild(this.tooltipElement);
                    }
                }, 100);
            }
        }
    }

    /**
     * Bind global events
     */
    bindEvents() {
        document.addEventListener('mousemove', (e) => {
            if (this.activeTooltip) {
                this.updateTooltipPosition(e.clientX, e.clientY);
            }
        });

        document.addEventListener('mouseout', (e) => {
            if (this.activeTooltip && !this.tooltipElement.contains(e.relatedTarget)) {
                this.hideTooltip();
            }
        });
    }

    /**
     * Show tooltip
     * @param {HTMLElement} element - Element to attach tooltip to
     * @param {string|HTMLElement} content - Tooltip content
     * @param {Object} options - Tooltip options
     */
    showTooltip(element, content, options = {}) {
        if (!this.tooltipElement) {
            this.createTooltipElement();
        }

        const tooltipId = options.id || `tooltip-${Date.now()}`;
        
        // Set content
        if (typeof content === 'string') {
            this.tooltipElement.innerHTML = content;
        } else if (content instanceof HTMLElement) {
            this.tooltipElement.innerHTML = '';
            this.tooltipElement.appendChild(content);
        }

        // Store tooltip info
        this.tooltips.set(tooltipId, {
            element,
            content,
            options
        });
        this.activeTooltip = tooltipId;

        // Position tooltip
        const rect = element.getBoundingClientRect();
        const x = options.x !== undefined ? options.x : rect.left + rect.width / 2;
        const y = options.y !== undefined ? options.y : rect.top - 10;

        this.updateTooltipPosition(x, y);
        this.tooltipElement.style.display = 'block';

        // Add custom classes
        if (options.className) {
            this.tooltipElement.className = `risk-tooltip ${options.className}`;
        }
    }

    /**
     * Update tooltip position
     */
    updateTooltipPosition(x, y) {
        if (!this.tooltipElement) return;

        const tooltipRect = this.tooltipElement.getBoundingClientRect();
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;

        // Adjust horizontal position to stay in viewport
        let left = x - tooltipRect.width / 2;
        if (left < 10) left = 10;
        if (left + tooltipRect.width > viewportWidth - 10) {
            left = viewportWidth - tooltipRect.width - 10;
        }

        // Adjust vertical position to stay in viewport
        let top = y - tooltipRect.height - 10;
        if (top < 10) {
            top = y + 20; // Show below if not enough space above
        }
        if (top + tooltipRect.height > viewportHeight - 10) {
            top = viewportHeight - tooltipRect.height - 10;
        }

        this.tooltipElement.style.left = `${left}px`;
        this.tooltipElement.style.top = `${top}px`;
    }

    /**
     * Update tooltip content
     */
    updateTooltip(tooltipId, content) {
        if (!this.activeTooltip || this.activeTooltip !== tooltipId) {
            return;
        }

        if (typeof content === 'string') {
            this.tooltipElement.innerHTML = content;
        } else if (content instanceof HTMLElement) {
            this.tooltipElement.innerHTML = '';
            this.tooltipElement.appendChild(content);
        }

        // Update stored content
        const tooltip = this.tooltips.get(tooltipId);
        if (tooltip) {
            tooltip.content = content;
        }
    }

    /**
     * Hide tooltip
     */
    hideTooltip(tooltipId = null) {
        if (tooltipId && this.activeTooltip !== tooltipId) {
            return;
        }

        if (this.tooltipElement) {
            this.tooltipElement.style.display = 'none';
        }

        this.activeTooltip = null;
    }

    /**
     * Show bar tooltip (for bar charts)
     */
    showBarTooltip(element, data, options = {}) {
        const content = this.createBarTooltipContent(data);
        this.showTooltip(element, content, options);
    }

    /**
     * Update bar tooltip
     */
    updateBarTooltip(tooltipId, data) {
        const content = this.createBarTooltipContent(data);
        this.updateTooltip(tooltipId, content);
    }

    /**
     * Hide bar tooltip
     */
    hideBarTooltip(tooltipId = null) {
        this.hideTooltip(tooltipId);
    }

    /**
     * Create bar tooltip content
     */
    createBarTooltipContent(data) {
        const div = document.createElement('div');
        div.innerHTML = `
            <div style="font-weight: 600; margin-bottom: 8px;">${data.label || 'Risk Indicator'}</div>
            <div style="margin-bottom: 4px;">Value: <strong>${data.value}</strong></div>
            ${data.score ? `<div style="margin-bottom: 4px;">Score: <strong>${data.score}</strong></div>` : ''}
            ${data.description ? `<div style="margin-top: 8px; font-size: 0.8rem; opacity: 0.9;">${data.description}</div>` : ''}
        `;
        return div;
    }

    /**
     * Show data point tooltip
     */
    showDataPointTooltip(element, data, options = {}) {
        const content = this.createDataPointTooltipContent(data);
        this.showTooltip(element, content, options);
    }

    /**
     * Hide data point tooltip
     */
    hideDataPointTooltip(tooltipId = null) {
        this.hideTooltip(tooltipId);
    }

    /**
     * Create data point tooltip content
     */
    createDataPointTooltipContent(data) {
        const div = document.createElement('div');
        div.innerHTML = `
            <div style="font-weight: 600; margin-bottom: 8px;">${data.title || 'Data Point'}</div>
            ${data.x ? `<div>X: <strong>${data.x}</strong></div>` : ''}
            ${data.y ? `<div>Y: <strong>${data.y}</strong></div>` : ''}
            ${data.value ? `<div>Value: <strong>${data.value}</strong></div>` : ''}
            ${data.timestamp ? `<div style="margin-top: 8px; font-size: 0.8rem; opacity: 0.9;">${new Date(data.timestamp).toLocaleString()}</div>` : ''}
        `;
        return div;
    }

    /**
     * Cleanup
     */
    destroy() {
        if (this.tooltipElement && this.tooltipElement.parentNode) {
            this.tooltipElement.parentNode.removeChild(this.tooltipElement);
        }
        this.tooltips.clear();
        this.activeTooltip = null;
    }
}

// Global functions for backward compatibility
function showTooltip(element, content, options) {
    if (!window.riskTooltipSystem) {
        window.riskTooltipSystem = new RiskTooltipSystem();
    }
    return window.riskTooltipSystem.showTooltip(element, content, options);
}

function updateTooltip(tooltipId, content) {
    if (window.riskTooltipSystem) {
        window.riskTooltipSystem.updateTooltip(tooltipId, content);
    }
}

function hideTooltip(tooltipId) {
    if (window.riskTooltipSystem) {
        window.riskTooltipSystem.hideTooltip(tooltipId);
    }
}

function showBarTooltip(element, data, options) {
    if (!window.riskTooltipSystem) {
        window.riskTooltipSystem = new RiskTooltipSystem();
    }
    return window.riskTooltipSystem.showBarTooltip(element, data, options);
}

function updateBarTooltip(tooltipId, data) {
    if (window.riskTooltipSystem) {
        window.riskTooltipSystem.updateBarTooltip(tooltipId, data);
    }
}

function hideBarTooltip(tooltipId) {
    if (window.riskTooltipSystem) {
        window.riskTooltipSystem.hideBarTooltip(tooltipId);
    }
}

function showDataPointTooltip(element, data, options) {
    if (!window.riskTooltipSystem) {
        window.riskTooltipSystem = new RiskTooltipSystem();
    }
    return window.riskTooltipSystem.showDataPointTooltip(element, data, options);
}

function hideDataPointTooltip(tooltipId) {
    if (window.riskTooltipSystem) {
        window.riskTooltipSystem.hideDataPointTooltip(tooltipId);
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.RiskTooltipSystem = RiskTooltipSystem;
    window.riskTooltipSystem = new RiskTooltipSystem();
}

