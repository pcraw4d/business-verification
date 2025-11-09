/**
 * KYB Platform Monitoring Dashboard JavaScript
 * Provides real-time monitoring data visualization and alerting
 */

class MonitoringDashboard {
    constructor() {
        this.charts = {};
        this.refreshInterval = null;
        this.isRefreshing = false;
        this.alertSound = null;
        this.lastAlertCheck = Date.now();
        
        this.init();
    }

    /**
     * Initialize the monitoring dashboard
     */
    init() {
        this.setupCharts();
        this.loadInitialData();
        this.setupEventListeners();
        this.startAutoRefresh();
        this.updateTimestamp();
        
        // Initialize alert sound
        this.alertSound = new Audio('data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBSuBzvLZiTYIG2m98OScTgwOUarm7blmGgU7k9n1unEiBC13yO/eizEIHWq+8+OWT');
    }

    /**
     * Setup Chart.js charts for data visualization
     */
    setupCharts() {
        // Request Rate Chart
        const requestRateCtx = document.getElementById('requestRateChart').getContext('2d');
        this.charts.requestRate = new Chart(requestRateCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Requests/sec',
                    data: [],
                    borderColor: '#3498db',
                    backgroundColor: 'rgba(52, 152, 219, 0.1)',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    },
                    x: {
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });

        // Response Time Chart
        const responseTimeCtx = document.getElementById('responseTimeChart').getContext('2d');
        this.charts.responseTime = new Chart(responseTimeCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Response Time (ms)',
                    data: [],
                    borderColor: '#e74c3c',
                    backgroundColor: 'rgba(231, 76, 60, 0.1)',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    },
                    x: {
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Handle window resize
        window.addEventListener('resize', () => {
            Object.values(this.charts).forEach(chart => {
                chart.resize();
            });
        });

        // Handle visibility change to pause/resume auto-refresh
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                this.stopAutoRefresh();
            } else {
                this.startAutoRefresh();
                this.refreshDashboard();
            }
        });
    }

    /**
     * Load initial dashboard data
     */
    async loadInitialData() {
        try {
            await this.fetchMetrics();
            await this.fetchAlerts();
            await this.fetchHealthChecks();
        } catch (error) {
            console.error('Error loading initial data:', error);
            this.showError('Failed to load monitoring data');
        }
    }

    /**
     * Fetch metrics from the API
     */
    async fetchMetrics() {
        try {
            const response = await fetch('/api/v3/monitoring/metrics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.updateMetrics(data);
            this.updateCharts(data);
        } catch (error) {
            console.error('Error fetching metrics:', error);
            // Use mock data for demonstration
            this.updateMetrics(this.getMockMetrics());
            this.updateCharts(this.getMockMetrics());
        }
    }

    /**
     * Fetch alerts from the API
     */
    async fetchAlerts() {
        try {
            const response = await fetch('/api/v3/monitoring/alerts');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.updateAlerts(data);
        } catch (error) {
            console.error('Error fetching alerts:', error);
            // Use mock data for demonstration
            this.updateAlerts(this.getMockAlerts());
        }
    }

    /**
     * Fetch health checks from the API
     */
    async fetchHealthChecks() {
        try {
            const response = await fetch('/api/v3/monitoring/health');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.updateHealthChecks(data);
        } catch (error) {
            console.error('Error fetching health checks:', error);
            // Use mock data for demonstration
            this.updateHealthChecks(this.getMockHealthChecks());
        }
    }

    /**
     * Update metrics display
     */
    updateMetrics(data) {
        document.getElementById('requestRate').textContent = data.requestRate || '0';
        document.getElementById('responseTime').textContent = `${data.responseTime || 0}ms`;
        document.getElementById('errorRate').textContent = `${data.errorRate || 0}%`;
        document.getElementById('activeUsers').textContent = data.activeUsers || '0';
        document.getElementById('memoryUsage').textContent = `${data.memoryUsage || 0}%`;
        document.getElementById('cpuUsage').textContent = `${data.cpuUsage || 0}%`;

        // Update change indicators
        this.updateChangeIndicator('requestRateChange', data.requestRateChange);
        this.updateChangeIndicator('responseTimeChange', data.responseTimeChange);
        this.updateChangeIndicator('errorRateChange', data.errorRateChange);
        this.updateChangeIndicator('activeUsersChange', data.activeUsersChange);
        this.updateChangeIndicator('memoryUsageChange', data.memoryUsageChange);
        this.updateChangeIndicator('cpuUsageChange', data.cpuUsageChange);
    }

    /**
     * Update change indicators
     */
    updateChangeIndicator(elementId, change) {
        const element = document.getElementById(elementId);
        if (!element) return;

        const changeValue = change || 0;
        const parent = element.parentElement;
        
        // Remove existing classes
        parent.classList.remove('positive', 'negative', 'neutral');
        
        // Add appropriate class
        if (changeValue > 0) {
            parent.classList.add('positive');
        } else if (changeValue < 0) {
            parent.classList.add('negative');
        } else {
            parent.classList.add('neutral');
        }
        
        element.textContent = `${changeValue > 0 ? '+' : ''}${changeValue}`;
    }

    /**
     * Update charts with new data
     */
    updateCharts(data) {
        const now = new Date().toLocaleTimeString();
        
        // Update request rate chart
        if (this.charts.requestRate) {
            this.charts.requestRate.data.labels.push(now);
            this.charts.requestRate.data.datasets[0].data.push(data.requestRate || 0);
            
            // Keep only last 20 data points
            if (this.charts.requestRate.data.labels.length > 20) {
                this.charts.requestRate.data.labels.shift();
                this.charts.requestRate.data.datasets[0].data.shift();
            }
            
            this.charts.requestRate.update('none');
        }

        // Update response time chart
        if (this.charts.responseTime) {
            this.charts.responseTime.data.labels.push(now);
            this.charts.responseTime.data.datasets[0].data.push(data.responseTime || 0);
            
            // Keep only last 20 data points
            if (this.charts.responseTime.data.labels.length > 20) {
                this.charts.responseTime.data.labels.shift();
                this.charts.responseTime.data.datasets[0].data.shift();
            }
            
            this.charts.responseTime.update('none');
        }
    }

    /**
     * Update alerts display
     */
    updateAlerts(alerts) {
        const alertsList = document.getElementById('alertsList');
        alertsList.innerHTML = '';

        if (!alerts || alerts.length === 0) {
            alertsList.innerHTML = '<div class="alert-item alert-info"><i class="fas fa-info-circle alert-icon"></i><div class="alert-content"><div class="alert-title">No Active Alerts</div><div class="alert-description">All systems are operating normally</div></div></div>';
            return;
        }

        alerts.forEach(alert => {
            const alertElement = document.createElement('div');
            alertElement.className = `alert-item alert-${alert.severity}`;
            
            const icon = this.getAlertIcon(alert.severity);
            const timeAgo = this.getTimeAgo(alert.timestamp);
            
            alertElement.innerHTML = `
                <i class="fas ${icon} alert-icon"></i>
                <div class="alert-content">
                    <div class="alert-title">${alert.title}</div>
                    <div class="alert-description">${alert.description}</div>
                </div>
                <div class="alert-time">${timeAgo}</div>
            `;
            
            alertsList.appendChild(alertElement);
        });

        // Check for new critical alerts
        this.checkForNewAlerts(alerts);
    }

    /**
     * Update health checks display
     */
    updateHealthChecks(healthChecks) {
        const healthChecksContainer = document.getElementById('healthChecks');
        healthChecksContainer.innerHTML = '';

        if (!healthChecks || healthChecks.length === 0) {
            return;
        }

        healthChecks.forEach(check => {
            const checkElement = document.createElement('div');
            checkElement.className = 'health-check';
            
            const icon = check.status === 'healthy' ? 'fa-check-circle' : 
                        check.status === 'warning' ? 'fa-exclamation-triangle' : 'fa-times-circle';
            const iconColor = check.status === 'healthy' ? '#27ae60' : 
                             check.status === 'warning' ? '#f39c12' : '#e74c3c';
            
            checkElement.innerHTML = `
                <i class="fas ${icon} health-check-icon" style="color: ${iconColor}"></i>
                <div class="health-check-name">${check.name}</div>
            `;
            
            healthChecksContainer.appendChild(checkElement);
        });
    }

    /**
     * Get alert icon based on severity
     */
    getAlertIcon(severity) {
        switch (severity) {
            case 'critical': return 'fa-exclamation-circle';
            case 'warning': return 'fa-exclamation-triangle';
            case 'info': return 'fa-info-circle';
            default: return 'fa-bell';
        }
    }

    /**
     * Get time ago string
     */
    getTimeAgo(timestamp) {
        const now = Date.now();
        const diff = now - timestamp;
        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(diff / 3600000);
        const days = Math.floor(diff / 86400000);

        if (days > 0) return `${days}d ago`;
        if (hours > 0) return `${hours}h ago`;
        if (minutes > 0) return `${minutes}m ago`;
        return 'Just now';
    }

    /**
     * Check for new critical alerts
     */
    checkForNewAlerts(alerts) {
        const criticalAlerts = alerts.filter(alert => 
            alert.severity === 'critical' && 
            alert.timestamp > this.lastAlertCheck
        );

        if (criticalAlerts.length > 0) {
            this.playAlertSound();
            this.showNotification(`New critical alert: ${criticalAlerts[0].title}`);
        }

        this.lastAlertCheck = Date.now();
    }

    /**
     * Play alert sound
     */
    playAlertSound() {
        if (this.alertSound) {
            this.alertSound.play().catch(error => {
                console.log('Could not play alert sound:', error);
            });
        }
    }

    /**
     * Show notification
     */
    showNotification(message) {
        if ('Notification' in window && Notification.permission === 'granted') {
            new Notification('KYB Platform Alert', {
                body: message,
                icon: '/favicon.ico'
            });
        }
    }

    /**
     * Show error message
     */
    showError(message) {
        const alertsList = document.getElementById('alertsList');
        alertsList.innerHTML = `
            <div class="alert-item alert-critical">
                <i class="fas fa-exclamation-circle alert-icon"></i>
                <div class="alert-content">
                    <div class="alert-title">Dashboard Error</div>
                    <div class="alert-description">${message}</div>
                </div>
            </div>
        `;
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
        
        this.refreshInterval = setInterval(() => {
            this.refreshDashboard();
        }, 30000); // Refresh every 30 seconds
    }

    /**
     * Stop auto-refresh
     */
    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    /**
     * Refresh dashboard data
     */
    async refreshDashboard() {
        if (this.isRefreshing) return;
        
        this.isRefreshing = true;
        const refreshButton = document.querySelector('.refresh-button');
        const icon = refreshButton.querySelector('i');
        
        // Show loading state
        icon.className = 'fas fa-sync-alt loading';
        refreshButton.disabled = true;
        
        try {
            await Promise.all([
                this.fetchMetrics(),
                this.fetchAlerts(),
                this.fetchHealthChecks()
            ]);
            this.updateTimestamp();
        } catch (error) {
            console.error('Error refreshing dashboard:', error);
        } finally {
            // Reset button state
            icon.className = 'fas fa-sync-alt';
            refreshButton.disabled = false;
            this.isRefreshing = false;
        }
    }

    /**
     * Update timestamp
     */
    updateTimestamp() {
        const now = new Date().toLocaleString();
        document.getElementById('lastUpdated').textContent = now;
    }

    /**
     * Get mock metrics for demonstration
     */
    getMockMetrics() {
        return {
            requestRate: Math.floor(Math.random() * 100) + 50,
            responseTime: Math.floor(Math.random() * 200) + 100,
            errorRate: (Math.random() * 2).toFixed(1),
            activeUsers: Math.floor(Math.random() * 20) + 5,
            memoryUsage: Math.floor(Math.random() * 30) + 40,
            cpuUsage: Math.floor(Math.random() * 20) + 30,
            requestRateChange: (Math.random() * 10 - 5).toFixed(1),
            responseTimeChange: Math.floor(Math.random() * 20 - 10),
            errorRateChange: (Math.random() * 0.5 - 0.25).toFixed(1),
            activeUsersChange: Math.floor(Math.random() * 6 - 3),
            memoryUsageChange: Math.floor(Math.random() * 4 - 2),
            cpuUsageChange: (Math.random() * 4 - 2).toFixed(1)
        };
    }

    /**
     * Get mock alerts for demonstration
     */
    getMockAlerts() {
        const alerts = [];
        const alertTypes = [
            { severity: 'critical', title: 'High Error Rate', description: 'Error rate exceeded 5% threshold' },
            { severity: 'warning', title: 'High Response Time', description: '95th percentile response time is 2.5s' },
            { severity: 'info', title: 'Scheduled Maintenance', description: 'Database maintenance scheduled for tonight' }
        ];

        // Randomly show 0-2 alerts
        const numAlerts = Math.floor(Math.random() * 3);
        for (let i = 0; i < numAlerts; i++) {
            const alertType = alertTypes[Math.floor(Math.random() * alertTypes.length)];
            alerts.push({
                ...alertType,
                timestamp: Date.now() - Math.random() * 3600000 // Random time in last hour
            });
        }

        return alerts;
    }

    /**
     * Get mock health checks for demonstration
     */
    getMockHealthChecks() {
        return [
            { name: 'API Server', status: 'healthy' },
            { name: 'Database', status: 'healthy' },
            { name: 'Redis Cache', status: 'warning' },
            { name: 'External APIs', status: 'healthy' },
            { name: 'File System', status: 'healthy' },
            { name: 'Memory', status: 'healthy' }
        ];
    }
}

// Global refresh function
function refreshDashboard() {
    if (window.monitoringDashboard) {
        window.monitoringDashboard.refreshDashboard();
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Request notification permission
    if ('Notification' in window && Notification.permission === 'default') {
        Notification.requestPermission();
    }
    
    // Initialize dashboard
    window.monitoringDashboard = new MonitoringDashboard();
});

// Export for testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MonitoringDashboard;
}
