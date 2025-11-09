/**
 * External Data Sources Component
 * Lists and manages external data sources
 */
class ExternalDataSources {
    constructor(containerId) {
        this.containerId = containerId;
        this.container = document.getElementById(containerId);
        this.sources = [];
    }

    /**
     * Initialize the component
     */
    async init() {
        await this.loadSources();
        this.render();
    }

    /**
     * Load external data sources
     */
    async loadSources() {
        try {
            const response = await fetch('/api/v1/supported', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.sources = data.sources || data || [];
        } catch (error) {
            console.error('Error loading external data sources:', error);
            this.sources = [];
        }
    }

    /**
     * Render the sources list
     */
    render() {
        if (!this.container) return;

        if (this.sources.length === 0) {
            this.container.innerHTML = `
                <div style="text-align: center; padding: 2rem; color: #666;">
                    <i class="fas fa-inbox" style="font-size: 2rem; color: #ccc; margin-bottom: 1rem;"></i>
                    <p>No external data sources available</p>
                </div>
            `;
            return;
        }

        this.container.innerHTML = this.sources.map(source => `
            <div style="padding: 1rem; background: #f8f9fa; border-radius: 4px; margin-bottom: 0.5rem; display: flex; justify-content: space-between; align-items: center;">
                <div>
                    <div style="font-weight: 600; margin-bottom: 0.25rem;">${source.name || source}</div>
                    <div style="font-size: 0.85rem; color: #666;">${source.description || ''}</div>
                </div>
                <div>
                    <span style="padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.75rem; background: ${source.status === 'active' ? '#d4edda' : '#f8d7da'}; color: ${source.status === 'active' ? '#155724' : '#721c24'};">
                        ${source.status || 'unknown'}
                    </span>
                </div>
            </div>
        `).join('');
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
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.ExternalDataSources = ExternalDataSources;
}

