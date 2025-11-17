/**
 * Admin System Metrics Component
 * Displays system metrics and thresholds
 */
class AdminSystemMetrics {
    constructor() {
        this.metricsData = null;
        this.updateInterval = null;
        this.isUpdating = false;
        this.cache = new Map();
        this.cacheTimeout = 5000; // 5 second cache
    }

    /**
     * Initialize the system metrics component
     */
    init(containerId = 'systemMetrics') {
        this.container = document.getElementById(containerId);
        if (!this.container) {
            console.error('System metrics container not found');
            return;
        }

        this.render();
        this.startAutoRefresh();
    }

    /**
     * Render the system metrics UI
     */
    render() {
        if (!this.container) return;

        this.container.innerHTML = `
            <div class="metrics-grid" style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1rem;">
                <div class="metric-item">
                    <div class="metric-label">CPU Usage</div>
                    <div id="cpuUsage" class="metric-value-small">-</div>
                </div>
                <div class="metric-item">
                    <div class="metric-label">Memory Threshold</div>
                    <div id="memoryThreshold" class="metric-value-small">-</div>
                </div>
                <div class="metric-item">
                    <div class="metric-label">Request Rate</div>
                    <div id="requestRate" class="metric-value-small">-</div>
                </div>
                <div class="metric-item">
                    <div class="metric-label">Error Rate</div>
                    <div id="errorRate" class="metric-value-small">-</div>
                </div>
            </div>
            <div style="margin-top: 1.5rem;">
                <h3 style="margin-bottom: 1rem;">System Thresholds</h3>
                <div id="thresholdsList" class="thresholds-list"></div>
            </div>
        `;
    }

    /**
     * Fetch system metrics
     */
    async fetchSystemMetrics() {
        try {
            const response = await fetch('/api/v1/system', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Error fetching system metrics:', error);
            return null;
        }
    }

    /**
     * Fetch thresholds
     */
    async fetchThresholds() {
        try {
            const response = await fetch('/api/v1/thresholds', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Error fetching thresholds:', error);
            return null;
        }
    }

    /**
     * Update metrics display
     */
    async updateMetrics() {
        const metrics = await this.fetchSystemMetrics();
        const thresholds = await this.fetchThresholds();

        if (metrics) {
            this.updateMetricDisplay('cpuUsage', metrics.cpu_usage || 'N/A', '%');
            this.updateMetricDisplay('memoryThreshold', metrics.memory_threshold || 'N/A', 'MB');
            this.updateMetricDisplay('requestRate', metrics.request_rate || 'N/A', '/sec');
            this.updateMetricDisplay('errorRate', metrics.error_rate || 'N/A', '%');
        }

        if (thresholds) {
            this.updateThresholdsList(thresholds);
        }
    }

    /**
     * Update a metric display element
     */
    updateMetricDisplay(elementId, value, unit = '') {
        const element = document.getElementById(elementId);
        if (element) {
            if (typeof value === 'number') {
                element.textContent = `${value.toFixed(2)} ${unit}`;
            } else {
                element.textContent = `${value} ${unit}`;
            }
        }
    }

    /**
     * Update thresholds list
     */
    updateThresholdsList(thresholds) {
        const container = document.getElementById('thresholdsList');
        if (!container) return;

        const thresholdsList = thresholds.thresholds || thresholds || {};
        
        container.innerHTML = Object.entries(thresholdsList).map(([key, value]) => {
            return `
                <div style="display: flex; justify-content: space-between; padding: 0.75rem; background: #f8f9fa; border-radius: 4px; margin-bottom: 0.5rem;">
                    <span style="font-weight: 500;">${this.formatKey(key)}</span>
                    <span style="color: #4a90e2;">${value}</span>
                </div>
            `;
        }).join('');
    }

    /**
     * Format threshold key for display
     */
    formatKey(key) {
        return key
            .replace(/_/g, ' ')
            .replace(/\b\w/g, l => l.toUpperCase());
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh(interval = 30000) {
        this.updateMetrics();
        
        this.updateInterval = setInterval(() => {
            // Throttle updates
            if (this.isUpdating) return;
            this.isUpdating = true;
            
            requestAnimationFrame(() => {
                this.updateMetrics();
                this.isUpdating = false;
            });
        }, interval);
    }

    /**
     * Stop auto-refresh
     */
    stopAutoRefresh() {
        if (this.updateInterval) {
            clearInterval(this.updateInterval);
            this.updateInterval = null;
        }
    }

    /**
     * Get authentication token
     */
    getAuthToken() {
        const token = localStorage.getItem('auth_token') || localStorage.getItem('access_token');
        if (token) {
            return token;
        }

        const cookies = document.cookie.split(';');
        for (let cookie of cookies) {
            const [name, value] = cookie.trim().split('=');
            if (name === 'auth_token' || name === 'access_token') {
                return value;
            }
        }

        return null;
    }

    /**
     * Destroy the component
     */
    destroy() {
        this.stopAutoRefresh();
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AdminSystemMetrics;
}

