/**
 * Session Manager UI Component
 * Provides UI for managing user sessions
 */
class SessionManagerUI {
    constructor() {
        this.sessions = [];
        this.currentSession = null;
        this.metrics = null;
        this.updateInterval = null;
        this.isUpdating = false;
    }

    /**
     * Initialize the session manager UI
     */
    async init() {
        document.getElementById('loading').style.display = 'block';
        document.getElementById('sessionsDashboard').style.display = 'none';

        try {
            await this.loadSessions();
            await this.loadMetrics();
            this.render();
            this.startAutoRefresh();
        } catch (error) {
            console.error('Error initializing session manager UI:', error);
            this.showError('Failed to load sessions: ' + error.message);
        } finally {
            document.getElementById('loading').style.display = 'none';
            document.getElementById('sessionsDashboard').style.display = 'block';
        }
    }

    /**
     * Load sessions from API
     */
    async loadSessions() {
        try {
            const response = await fetch('/api/v1/sessions', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.sessions = data.sessions || data || [];
        } catch (error) {
            console.error('Error loading sessions:', error);
            this.sessions = [];
        }
    }

    /**
     * Load session metrics
     */
    async loadMetrics() {
        try {
            const response = await fetch('/api/v1/sessions/metrics', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.metrics = await response.json();
        } catch (error) {
            console.error('Error loading session metrics:', error);
            this.metrics = null;
        }
    }

    /**
     * Render the sessions UI
     */
    render() {
        this.renderSessionsList();
        this.renderMetrics();
        this.updateSummary();
    }

    /**
     * Render sessions list
     */
    renderSessionsList() {
        const container = document.getElementById('sessionsList');
        if (!container) return;

        if (this.sessions.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-inbox"></i>
                    <p>No active sessions</p>
                </div>
            `;
            return;
        }

        container.innerHTML = this.sessions.map(session => `
            <div class="session-card ${session.is_current ? 'active' : ''}">
                <div class="session-card-header">
                    <div class="session-card-title">
                        ${session.device || 'Unknown Device'}
                        ${session.is_current ? '<span style="color: #4a90e2;">(Current)</span>' : ''}
                    </div>
                    <span class="session-badge ${session.is_active ? 'active' : 'inactive'}">
                        ${session.is_active ? 'Active' : 'Inactive'}
                    </span>
                </div>
                <div class="session-info">
                    <div class="session-info-item">
                        <span class="session-info-label">IP Address</span>
                        <span class="session-info-value">${session.ip_address || 'N/A'}</span>
                    </div>
                    <div class="session-info-item">
                        <span class="session-info-label">User Agent</span>
                        <span class="session-info-value" style="font-size: 0.8rem;">${this.truncate(session.user_agent || 'N/A', 50)}</span>
                    </div>
                    <div class="session-info-item">
                        <span class="session-info-label">Created</span>
                        <span class="session-info-value">${this.formatDate(session.created_at)}</span>
                    </div>
                    <div class="session-info-item">
                        <span class="session-info-label">Last Activity</span>
                        <span class="session-info-value">${this.formatDate(session.last_activity)}</span>
                    </div>
                </div>
                <div class="session-actions">
                    ${!session.is_current ? `
                        <button class="btn btn-outline" onclick="sessionManagerUI.switchSession('${session.id}')">
                            <i class="fas fa-exchange-alt"></i> Switch
                        </button>
                    ` : ''}
                    <button class="btn btn-danger" onclick="sessionManagerUI.terminateSession('${session.id}')">
                        <i class="fas fa-times"></i> Terminate
                    </button>
                </div>
            </div>
        `).join('');
    }

    /**
     * Render session metrics
     */
    renderMetrics() {
        const container = document.getElementById('sessionMetrics');
        if (!container) return;

        if (!this.metrics) {
            container.innerHTML = '<p style="color: #666;">No metrics available</p>';
            return;
        }

        container.innerHTML = `
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem;">
                <div>
                    <div style="font-size: 0.9rem; color: #666; margin-bottom: 0.5rem;">Total Sessions</div>
                    <div style="font-size: 1.5rem; font-weight: 600; color: #4a90e2;">${this.metrics.total_sessions || 0}</div>
                </div>
                <div>
                    <div style="font-size: 0.9rem; color: #666; margin-bottom: 0.5rem;">Active Sessions</div>
                    <div style="font-size: 1.5rem; font-weight: 600; color: #28a745;">${this.metrics.active_sessions || 0}</div>
                </div>
                <div>
                    <div style="font-size: 0.9rem; color: #666; margin-bottom: 0.5rem;">Avg Session Duration</div>
                    <div style="font-size: 1.5rem; font-weight: 600; color: #4a90e2;">${this.formatDuration(this.metrics.avg_duration)}</div>
                </div>
            </div>
        `;
    }

    /**
     * Update summary cards
     */
    updateSummary() {
        const activeCount = this.sessions.filter(s => s.is_active).length;
        const totalCount = this.sessions.length;
        const lastActivity = this.sessions.length > 0 
            ? this.sessions.sort((a, b) => new Date(b.last_activity) - new Date(a.last_activity))[0].last_activity
            : null;

        const activeCountEl = document.getElementById('activeSessionsCount');
        if (activeCountEl) {
            activeCountEl.textContent = activeCount;
        }

        const totalCountEl = document.getElementById('totalSessionsCount');
        if (totalCountEl) {
            totalCountEl.textContent = totalCount;
        }

        const lastActivityEl = document.getElementById('lastActivity');
        if (lastActivityEl) {
            lastActivityEl.textContent = lastActivity ? this.formatDate(lastActivity) : 'Never';
        }
    }

    /**
     * Create a new session
     */
    async createSession() {
        try {
            const response = await fetch('/api/v1/sessions', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    device: navigator.userAgent,
                    ip_address: await this.getIPAddress()
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            alert('New session created successfully!');
            await this.loadSessions();
            this.render();
        } catch (error) {
            console.error('Error creating session:', error);
            alert('Failed to create session: ' + error.message);
        }
    }

    /**
     * Switch to a different session
     */
    async switchSession(sessionId) {
        if (!confirm('Switch to this session?')) {
            return;
        }

        try {
            // Note: This would typically be handled by the backend
            // For now, we'll just reload the page
            alert('Session switching functionality will be implemented in the backend');
            // window.location.reload();
        } catch (error) {
            console.error('Error switching session:', error);
            alert('Failed to switch session: ' + error.message);
        }
    }

    /**
     * Terminate a session
     */
    async terminateSession(sessionId) {
        if (!confirm('Are you sure you want to terminate this session?')) {
            return;
        }

        try {
            const response = await fetch(`/api/v1/sessions?id=${sessionId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            alert('Session terminated successfully');
            await this.loadSessions();
            this.render();
        } catch (error) {
            console.error('Error terminating session:', error);
            alert('Failed to terminate session: ' + error.message);
        }
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh(interval = 30000) {
        this.updateInterval = setInterval(async () => {
            // Throttle updates
            if (this.isUpdating) return;
            this.isUpdating = true;
            
            try {
                await this.loadSessions();
                this.render();
            } finally {
                this.isUpdating = false;
            }
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
     * Show error message
     */
    showError(message) {
        const container = document.getElementById('sessionsContent');
        if (container) {
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error';
            errorDiv.innerHTML = `<i class="fas fa-exclamation-circle"></i> ${message}`;
            container.insertBefore(errorDiv, container.firstChild);
        }
    }

    /**
     * Format date
     */
    formatDate(dateString) {
        if (!dateString) return 'N/A';
        const date = new Date(dateString);
        return date.toLocaleString();
    }

    /**
     * Format duration
     */
    formatDuration(seconds) {
        if (!seconds) return 'N/A';
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        if (hours > 0) {
            return `${hours}h ${minutes}m`;
        }
        return `${minutes}m`;
    }

    /**
     * Truncate string
     */
    truncate(str, maxLength) {
        if (!str || str.length <= maxLength) return str;
        return str.substring(0, maxLength) + '...';
    }

    /**
     * Get IP address (mock)
     */
    async getIPAddress() {
        // In a real implementation, this would be handled by the backend
        return '127.0.0.1';
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
    window.SessionManagerUI = SessionManagerUI;
}

