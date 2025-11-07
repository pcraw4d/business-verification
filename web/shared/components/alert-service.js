/**
 * Shared Alert Service
 * Provides unified alert management for all pages
 */

function getEventBusInstance() {
    if (typeof getEventBus !== 'undefined') {
        return getEventBus();
    }
    return {
        emit: () => {},
        on: () => () => {},
        off: () => {},
        once: () => {}
    };
}

class SharedAlertService {
    constructor(config = {}) {
        this.eventBus = config.eventBus || getEventBusInstance();
        this.alerts = new Map();
        this.alertContainer = null;
    }
    
    /**
     * Create an alert
     * @param {Object} alertData - Alert data
     * @returns {string} Alert ID
     */
    createAlert(alertData) {
        const alert = {
            id: alertData.id || this.generateAlertId(),
            type: alertData.type || 'info',
            severity: alertData.severity || 'medium',
            title: alertData.title || 'Alert',
            message: alertData.message || '',
            category: alertData.category || null,
            merchantId: alertData.merchantId || null,
            source: alertData.source || 'system',
            status: 'active',
            createdAt: new Date().toISOString(),
            acknowledgedAt: null,
            resolvedAt: null,
            metadata: alertData.metadata || {}
        };
        
        this.alerts.set(alert.id, alert);
        
        // Emit event
        this.eventBus.emit('alert-created', { alert });
        
        // Display alert if container is set
        if (this.alertContainer) {
            this.displayAlert(alert);
        }
        
        return alert.id;
    }
    
    /**
     * Acknowledge an alert
     * @param {string} alertId - Alert ID
     * @returns {boolean} Success
     */
    acknowledgeAlert(alertId) {
        const alert = this.alerts.get(alertId);
        if (!alert) {
            return false;
        }
        
        alert.status = 'acknowledged';
        alert.acknowledgedAt = new Date().toISOString();
        this.alerts.set(alertId, alert);
        
        // Emit event
        this.eventBus.emit('alert-acknowledged', { alert });
        
        // Update UI
        this.updateAlertDisplay(alert);
        
        return true;
    }
    
    /**
     * Resolve an alert
     * @param {string} alertId - Alert ID
     * @returns {boolean} Success
     */
    resolveAlert(alertId) {
        const alert = this.alerts.get(alertId);
        if (!alert) {
            return false;
        }
        
        alert.status = 'resolved';
        alert.resolvedAt = new Date().toISOString();
        this.alerts.set(alertId, alert);
        
        // Emit event
        this.eventBus.emit('alert-resolved', { alert });
        
        // Update UI
        this.updateAlertDisplay(alert);
        
        return true;
    }
    
    /**
     * Get alerts
     * @param {Object} filters - Filter options
     * @returns {Array} Array of alerts
     */
    getAlerts(filters = {}) {
        let alerts = Array.from(this.alerts.values());
        
        // Apply filters
        if (filters.type) {
            alerts = alerts.filter(a => a.type === filters.type);
        }
        
        if (filters.severity) {
            alerts = alerts.filter(a => a.severity === filters.severity);
        }
        
        if (filters.status) {
            alerts = alerts.filter(a => a.status === filters.status);
        }
        
        if (filters.merchantId) {
            alerts = alerts.filter(a => a.merchantId === filters.merchantId);
        }
        
        if (filters.category) {
            alerts = alerts.filter(a => a.category === filters.category);
        }
        
        // Sort by creation date (newest first)
        alerts.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
        
        return alerts;
    }
    
    /**
     * Display alert in UI
     * @param {Object} alert - Alert object
     */
    displayAlert(alert) {
        if (!this.alertContainer) {
            return;
        }
        
        const alertElement = this.createAlertElement(alert);
        this.alertContainer.appendChild(alertElement);
        
        // Auto-remove after 5 seconds for non-critical alerts
        if (alert.severity !== 'critical') {
            setTimeout(() => {
                if (alertElement.parentNode) {
                    alertElement.remove();
                }
            }, 5000);
        }
    }
    
    /**
     * Create alert DOM element
     * @param {Object} alert - Alert object
     * @returns {HTMLElement} Alert element
     */
    createAlertElement(alert) {
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${alert.severity} alert-${alert.type}`;
        alertDiv.setAttribute('data-alert-id', alert.id);
        
        const severityColors = {
            critical: 'bg-red-100 border-red-500 text-red-800',
            high: 'bg-orange-100 border-orange-500 text-orange-800',
            medium: 'bg-yellow-100 border-yellow-500 text-yellow-800',
            low: 'bg-blue-100 border-blue-500 text-blue-800'
        };
        
        const iconMap = {
            critical: 'exclamation-circle',
            high: 'exclamation-triangle',
            medium: 'info-circle',
            low: 'info-circle'
        };
        
        alertDiv.className = `p-4 mb-2 rounded-lg border-l-4 ${severityColors[alert.severity] || severityColors.medium}`;
        
        alertDiv.innerHTML = `
            <div class="flex items-start justify-between">
                <div class="flex items-start">
                    <i class="fas fa-${iconMap[alert.severity] || iconMap.medium} mt-1 mr-3"></i>
                    <div>
                        <h4 class="font-semibold">${this.escapeHtml(alert.title)}</h4>
                        ${alert.message ? `<p class="text-sm mt-1">${this.escapeHtml(alert.message)}</p>` : ''}
                    </div>
                </div>
                <div class="flex items-center space-x-2 ml-4">
                    ${alert.status === 'active' ? `
                        <button onclick="window.acknowledgeAlert('${alert.id}')" 
                                class="text-xs px-2 py-1 rounded hover:bg-opacity-80">
                            Acknowledge
                        </button>
                        <button onclick="window.resolveAlert('${alert.id}')" 
                                class="text-xs px-2 py-1 rounded hover:bg-opacity-80">
                            Resolve
                        </button>
                    ` : ''}
                    <button onclick="this.parentElement.parentElement.parentElement.remove()" 
                            class="text-xs px-2 py-1 rounded hover:bg-opacity-80">
                        ×
                    </button>
                </div>
            </div>
        `;
        
        return alertDiv;
    }
    
    /**
     * Update alert display
     * @param {Object} alert - Alert object
     */
    updateAlertDisplay(alert) {
        if (!this.alertContainer) {
            return;
        }
        
        const alertElement = this.alertContainer.querySelector(`[data-alert-id="${alert.id}"]`);
        if (alertElement) {
            if (alert.status === 'resolved') {
                alertElement.style.opacity = '0.5';
                alertElement.querySelector('.flex.items-center.space-x-2')?.remove();
            } else if (alert.status === 'acknowledged') {
                alertElement.style.opacity = '0.7';
            }
        }
    }
    
    /**
     * Set alert container
     * @param {string|HTMLElement} container - Container selector or element
     */
    setAlertContainer(container) {
        if (typeof container === 'string') {
            this.alertContainer = document.querySelector(container);
        } else {
            this.alertContainer = container;
        }
    }
    
    /**
     * Generate alert ID
     * @returns {string} Alert ID
     */
    generateAlertId() {
        return `alert_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }
    
    /**
     * Escape HTML
     * @param {string} text - Text to escape
     * @returns {string} Escaped text
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
    
    /**
     * Clear all alerts
     */
    clearAllAlerts() {
        this.alerts.clear();
        if (this.alertContainer) {
            this.alertContainer.innerHTML = '';
        }
    }
    
    /**
     * Get alert count
     * @param {Object} filters - Filter options
     * @returns {number} Alert count
     */
    getAlertCount(filters = {}) {
        return this.getAlerts(filters).length;
    }
}

// Export singleton instance
let alertServiceInstance = null;

export function getAlertService(config) {
    if (!alertServiceInstance) {
        alertServiceInstance = new SharedAlertService(config);
    }
    return alertServiceInstance;
}

export { SharedAlertService };

// Make available globally for non-module environments
if (typeof window !== 'undefined') {
    window.getAlertService = getAlertService;
    window.SharedAlertService = SharedAlertService;
}

