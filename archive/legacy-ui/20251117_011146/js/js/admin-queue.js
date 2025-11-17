/**
 * Admin Queue Management Controller
 * Manages background job queue
 */
class AdminQueue {
    constructor() {
        this.queueStatus = null;
        this.jobs = [];
        this.updateInterval = null;
    }

    /**
     * Initialize the queue dashboard
     */
    async init() {
        document.getElementById('loading').style.display = 'block';
        document.getElementById('queueDashboard').style.display = 'none';

        try {
            await this.loadQueueStatus();
            await this.loadJobs();
            this.render();
            this.startAutoRefresh();
        } catch (error) {
            console.error('Error initializing queue dashboard:', error);
            this.showError('Failed to load queue status: ' + error.message);
        } finally {
            document.getElementById('loading').style.display = 'none';
            document.getElementById('queueDashboard').style.display = 'block';
        }
    }

    /**
     * Load queue status
     */
    async loadQueueStatus() {
        try {
            const response = await fetch('/api/v1/queue', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.queueStatus = await response.json();
        } catch (error) {
            console.error('Error loading queue status:', error);
            this.queueStatus = null;
        }
    }

    /**
     * Load queue jobs
     */
    async loadJobs() {
        try {
            const response = await fetch('/api/v1/queue/jobs', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.jobs = data.jobs || data || [];
        } catch (error) {
            console.error('Error loading jobs:', error);
            this.jobs = [];
        }
    }

    /**
     * Render the queue dashboard
     */
    render() {
        this.renderSummary();
        this.renderJobs();
    }

    /**
     * Render summary cards
     */
    renderSummary() {
        if (!this.queueStatus) return;

        const pending = this.queueStatus.pending || 0;
        const processing = this.queueStatus.processing || 0;
        const completed = this.queueStatus.completed || 0;
        const failed = this.queueStatus.failed || 0;

        const pendingEl = document.getElementById('pendingJobs');
        if (pendingEl) pendingEl.textContent = pending;

        const processingEl = document.getElementById('processingJobs');
        if (processingEl) processingEl.textContent = processing;

        const completedEl = document.getElementById('completedJobs');
        if (completedEl) completedEl.textContent = completed;

        const failedEl = document.getElementById('failedJobs');
        if (failedEl) failedEl.textContent = failed;
    }

    /**
     * Render jobs list
     */
    renderJobs() {
        const container = document.getElementById('queueJobs');
        if (!container) return;

        if (this.jobs.length === 0) {
            container.innerHTML = '<p style="color: #666;">No jobs in queue</p>';
            return;
        }

        container.innerHTML = this.jobs.map(job => `
            <div class="job-item">
                <div>
                    <div style="font-weight: 600; margin-bottom: 0.25rem;">${job.name || job.id || 'Job'}</div>
                    <div style="font-size: 0.85rem; color: #666;">${job.description || ''}</div>
                    <div style="font-size: 0.8rem; color: #999; margin-top: 0.25rem;">Created: ${this.formatDate(job.created_at)}</div>
                </div>
                <div style="display: flex; align-items: center; gap: 1rem;">
                    <span class="job-status ${job.status || 'pending'}">${job.status || 'pending'}</span>
                    ${job.status === 'processing' || job.status === 'pending' ? `
                        <button class="btn btn-danger" onclick="adminQueue.cancelJob('${job.id}')">
                            <i class="fas fa-times"></i> Cancel
                        </button>
                    ` : ''}
                    ${job.status === 'failed' ? `
                        <button class="btn btn-primary" onclick="adminQueue.retryJob('${job.id}')">
                            <i class="fas fa-redo"></i> Retry
                        </button>
                    ` : ''}
                </div>
            </div>
        `).join('');
    }

    /**
     * Cancel a job
     */
    async cancelJob(jobId) {
        if (!confirm('Cancel this job?')) {
            return;
        }

        try {
            const response = await fetch(`/api/v1/queue/jobs/${jobId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            alert('Job cancelled successfully');
            await this.loadJobs();
            await this.loadQueueStatus();
            this.render();
        } catch (error) {
            console.error('Error cancelling job:', error);
            alert('Failed to cancel job: ' + error.message);
        }
    }

    /**
     * Retry a failed job
     */
    async retryJob(jobId) {
        try {
            const response = await fetch(`/api/v1/queue/jobs/${jobId}/retry`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            alert('Job retried successfully');
            await this.loadJobs();
            await this.loadQueueStatus();
            this.render();
        } catch (error) {
            console.error('Error retrying job:', error);
            alert('Failed to retry job: ' + error.message);
        }
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh(interval = 10000) {
        this.updateInterval = setInterval(async () => {
            await this.loadQueueStatus();
            await this.loadJobs();
            this.render();
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
        const container = document.getElementById('queueContent');
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
    window.AdminQueue = AdminQueue;
    window.adminQueue = null;
}

// Initialize global instance
document.addEventListener('DOMContentLoaded', () => {
    if (typeof AdminQueue !== 'undefined') {
        window.adminQueue = new AdminQueue();
    }
});

