/**
 * Analytics Insights Component
 * Displays analytics insights and trends
 */
class AnalyticsInsights {
    constructor(containerId, options = {}) {
        this.containerId = containerId;
        this.container = document.getElementById(containerId);
        this.options = {
            autoRefresh: true,
            refreshInterval: 60000,
            ...options
        };
        this.insights = null;
        this.updateInterval = null;
        this.isUpdating = false;
    }

    /**
     * Initialize the component
     */
    async init() {
        await this.loadInsights();
        this.render();
        
        if (this.options.autoRefresh) {
            this.startAutoRefresh();
        }
    }

    /**
     * Load insights from API
     */
    async loadInsights() {
        try {
            const response = await fetch('/api/v1/insights', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.insights = await response.json();
        } catch (error) {
            console.error('Error loading insights:', error);
            this.insights = null;
        }
    }

    /**
     * Render insights
     */
    render() {
        if (!this.container) return;

        if (!this.insights) {
            this.container.innerHTML = '<p style="color: #666;">No insights available</p>';
            return;
        }

        const trends = this.insights.trends || [];
        const patterns = this.insights.patterns || [];

        this.container.innerHTML = `
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem;">
                ${trends.map(trend => `
                    <div style="padding: 1rem; background: #f8f9fa; border-radius: 4px;">
                        <div style="font-weight: 600; margin-bottom: 0.5rem;">${trend.label || 'Trend'}</div>
                        <div style="font-size: 1.5rem; color: #4a90e2; font-weight: 600;">${trend.value || 'N/A'}</div>
                        <div style="font-size: 0.85rem; color: #666; margin-top: 0.25rem;">${trend.description || ''}</div>
                    </div>
                `).join('')}
            </div>
            ${patterns.length > 0 ? `
                <div style="margin-top: 2rem;">
                    <h3 style="margin-bottom: 1rem;">Patterns</h3>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1rem;">
                        ${patterns.map(pattern => `
                            <div style="padding: 1rem; background: white; border-radius: 4px; border-left: 3px solid #4a90e2;">
                                <div style="font-weight: 600; margin-bottom: 0.5rem;">${pattern.name || 'Pattern'}</div>
                                <div style="font-size: 0.9rem; color: #666;">${pattern.description || ''}</div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            ` : ''}
        `;
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh() {
        this.updateInterval = setInterval(async () => {
            // Throttle updates
            if (this.isUpdating) return;
            this.isUpdating = true;
            
            try {
                await this.loadInsights();
                this.render();
            } finally {
                this.isUpdating = false;
            }
        }, this.options.refreshInterval);
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
     * Cleanup
     */
    destroy() {
        this.stopAutoRefresh();
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.AnalyticsInsights = AnalyticsInsights;
}

