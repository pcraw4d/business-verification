/**
 * Monitoring Dashboard with Real Data Integration
 * Replaces mock data with real Supabase API calls
 * Provides comprehensive system monitoring with live data
 */

class MonitoringDashboardRealData {
    constructor() {
        this.dataIntegration = new RealDataIntegration();
        this.metrics = null;
        this.alerts = null;
        this.healthChecks = null;
        this.performanceData = null;
        this.refreshInterval = null;
        this.isLoading = false;
        this.charts = {};
        
        this.init();
    }

    /**
     * Initialize the dashboard
     */
    async init() {
        try {
            this.showLoadingState();
            await this.loadMonitoringData();
            this.initializeCharts();
            this.bindEvents();
            this.startAutoRefresh();
        } catch (error) {
            console.error('Failed to initialize monitoring dashboard:', error);
            this.showErrorState(error.message);
        }
    }

    /**
     * Load monitoring data from Supabase
     */
    async loadMonitoringData() {
        try {
            this.isLoading = true;
            
            // Load all monitoring data in parallel
            const [metrics, alerts, healthChecks, performanceData] = await Promise.all([
                this.dataIntegration.getSystemMetrics(),
                this.dataIntegration.getSystemAlerts(),
                this.dataIntegration.getHealthChecks(),
                this.dataIntegration.getPerformanceMetrics()
            ]);
            
            this.metrics = metrics;
            this.alerts = alerts;
            this.healthChecks = healthChecks;
            this.performanceData = performanceData;
            
            // Render all data
            this.renderMetricsData();
            this.renderAlertsData();
            this.renderHealthChecksData();
            this.renderPerformanceData();
            this.updateSystemStatus();
            
            this.hideLoadingState();
        } catch (error) {
            console.error('Failed to load monitoring data:', error);
            this.showErrorState('Failed to load monitoring data');
        } finally {
            this.isLoading = false;
        }
    }

    /**
     * Render metrics data
     */
    renderMetricsData() {
        if (!this.metrics) return;

        // System metrics
        this.updateElement('cpuUsage', `${this.metrics.cpu_usage || 0}%`);
        this.updateElement('memoryUsage', `${this.metrics.memory_usage || 0}%`);
        this.updateElement('diskUsage', `${this.metrics.disk_usage || 0}%`);
        this.updateElement('networkLatency', `${this.metrics.network_latency || 0}ms`);
        
        // Application metrics
        this.updateElement('requestRate', `${this.metrics.request_rate || 0}/min`);
        this.updateElement('errorRate', `${this.metrics.error_rate || 0}%`);
        this.updateElement('responseTime', `${this.metrics.response_time || 0}ms`);
        this.updateElement('activeConnections', this.metrics.active_connections || 0);
        
        // Database metrics
        this.updateElement('dbConnections', this.metrics.db_connections || 0);
        this.updateElement('dbQueryTime', `${this.metrics.db_query_time || 0}ms`);
        this.updateElement('dbCacheHitRate', `${this.metrics.db_cache_hit_rate || 0}%`);
        
        // Update metric indicators
        this.updateMetricIndicators();
    }

    /**
     * Render alerts data
     */
    renderAlertsData() {
        if (!this.alerts) return;

        const alertsContainer = document.getElementById('alertsList');
        if (!alertsContainer) return;

        // Sort alerts by severity and timestamp
        const sortedAlerts = this.alerts.sort((a, b) => {
            const severityOrder = { 'critical': 0, 'high': 1, 'medium': 2, 'low': 3 };
            if (severityOrder[a.severity] !== severityOrder[b.severity]) {
                return severityOrder[a.severity] - severityOrder[b.severity];
            }
            return new Date(b.timestamp) - new Date(a.timestamp);
        });

        alertsContainer.innerHTML = sortedAlerts.map(alert => `
            <div class="alert-item ${alert.severity}">
                <div class="alert-icon">
                    <i class="fas fa-${this.getAlertIcon(alert.severity)}"></i>
                </div>
                <div class="alert-content">
                    <div class="alert-title">${alert.title}</div>
                    <div class="alert-description">${alert.description}</div>
                    <div class="alert-meta">
                        <span class="alert-source">${alert.source}</span>
                        <span class="alert-time">${this.formatTimeAgo(new Date(alert.timestamp))}</span>
                    </div>
                </div>
                <div class="alert-actions">
                    <button class="btn btn-sm btn-outline" onclick="monitoringDashboard.acknowledgeAlert('${alert.id}')">
                        Acknowledge
                    </button>
                </div>
            </div>
        `).join('');

        // Update alert counters
        this.updateAlertCounters();
    }

    /**
     * Render health checks data
     */
    renderHealthChecksData() {
        if (!this.healthChecks) return;

        const healthChecksContainer = document.getElementById('healthChecksList');
        if (!healthChecksContainer) return;

        healthChecksContainer.innerHTML = this.healthChecks.map(check => `
            <div class="health-check-item ${check.status}">
                <div class="health-check-icon">
                    <i class="fas fa-${check.status === 'healthy' ? 'check-circle' : 'exclamation-triangle'}"></i>
                </div>
                <div class="health-check-content">
                    <div class="health-check-name">${check.name}</div>
                    <div class="health-check-description">${check.description}</div>
                    <div class="health-check-meta">
                        <span class="health-check-response-time">${check.response_time}ms</span>
                        <span class="health-check-last-check">${this.formatTimeAgo(new Date(check.last_check))}</span>
                    </div>
                </div>
                <div class="health-check-status">
                    <span class="status-badge ${check.status}">${check.status}</span>
                </div>
            </div>
        `).join('');

        // Update health check summary
        this.updateHealthCheckSummary();
    }

    /**
     * Render performance data
     */
    renderPerformanceData() {
        if (!this.performanceData) return;

        // Update performance charts
        this.updatePerformanceCharts();
        
        // Update performance metrics
        this.updateElement('avgResponseTime', `${this.performanceData.avg_response_time || 0}ms`);
        this.updateElement('p95ResponseTime', `${this.performanceData.p95_response_time || 0}ms`);
        this.updateElement('p99ResponseTime', `${this.performanceData.p99_response_time || 0}ms`);
        this.updateElement('throughput', `${this.performanceData.throughput || 0} req/s`);
    }

    /**
     * Update system status
     */
    updateSystemStatus() {
        const statusElement = document.getElementById('systemStatus');
        if (!statusElement) return;

        // Determine overall system status
        let overallStatus = 'healthy';
        let statusMessage = 'All systems operational';

        if (this.alerts && this.alerts.some(alert => alert.severity === 'critical' && !alert.acknowledged)) {
            overallStatus = 'critical';
            statusMessage = 'Critical alerts detected';
        } else if (this.alerts && this.alerts.some(alert => alert.severity === 'high' && !alert.acknowledged)) {
            overallStatus = 'warning';
            statusMessage = 'High priority alerts detected';
        } else if (this.healthChecks && this.healthChecks.some(check => check.status !== 'healthy')) {
            overallStatus = 'warning';
            statusMessage = 'Some health checks are failing';
        }

        statusElement.className = `system-status ${overallStatus}`;
        statusElement.innerHTML = `
            <div class="status-icon">
                <i class="fas fa-${this.getStatusIcon(overallStatus)}"></i>
            </div>
            <div class="status-content">
                <div class="status-title">${statusMessage}</div>
                <div class="status-subtitle">Last updated: ${new Date().toLocaleTimeString()}</div>
            </div>
        `;
    }

    /**
     * Update metric indicators
     */
    updateMetricIndicators() {
        if (!this.metrics) return;

        // CPU usage indicator
        this.updateMetricIndicator('cpuIndicator', this.metrics.cpu_usage, 80, 90);
        
        // Memory usage indicator
        this.updateMetricIndicator('memoryIndicator', this.metrics.memory_usage, 80, 90);
        
        // Disk usage indicator
        this.updateMetricIndicator('diskIndicator', this.metrics.disk_usage, 85, 95);
        
        // Error rate indicator
        this.updateMetricIndicator('errorRateIndicator', this.metrics.error_rate, 5, 10);
    }

    /**
     * Update individual metric indicator
     */
    updateMetricIndicator(elementId, value, warningThreshold, criticalThreshold) {
        const element = document.getElementById(elementId);
        if (!element) return;

        let status = 'healthy';
        if (value >= criticalThreshold) {
            status = 'critical';
        } else if (value >= warningThreshold) {
            status = 'warning';
        }

        element.className = `metric-indicator ${status}`;
        element.style.width = `${Math.min(value, 100)}%`;
    }

    /**
     * Update alert counters
     */
    updateAlertCounters() {
        if (!this.alerts) return;

        const counters = {
            critical: 0,
            high: 0,
            medium: 0,
            low: 0,
            total: 0
        };

        this.alerts.forEach(alert => {
            if (!alert.acknowledged) {
                counters[alert.severity]++;
                counters.total++;
            }
        });

        // Update counter elements
        Object.entries(counters).forEach(([severity, count]) => {
            this.updateElement(`${severity}Alerts`, count);
        });
    }

    /**
     * Update health check summary
     */
    updateHealthCheckSummary() {
        if (!this.healthChecks) return;

        const total = this.healthChecks.length;
        const healthy = this.healthChecks.filter(check => check.status === 'healthy').length;
        const unhealthy = total - healthy;

        this.updateElement('totalHealthChecks', total);
        this.updateElement('healthyChecks', healthy);
        this.updateElement('unhealthyChecks', unhealthy);
        
        // Update health percentage
        const healthPercentage = total > 0 ? Math.round((healthy / total) * 100) : 100;
        this.updateElement('healthPercentage', `${healthPercentage}%`);
    }

    /**
     * Initialize charts
     */
    initializeCharts() {
        // Initialize performance charts if Chart.js is available
        if (typeof Chart !== 'undefined') {
            this.initializeResponseTimeChart();
            this.initializeThroughputChart();
            this.initializeErrorRateChart();
        }
    }

    /**
     * Initialize response time chart
     */
    initializeResponseTimeChart() {
        const ctx = document.getElementById('responseTimeChart');
        if (!ctx) return;

        this.charts.responseTime = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Response Time (ms)',
                    data: [],
                    borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    /**
     * Initialize throughput chart
     */
    initializeThroughputChart() {
        const ctx = document.getElementById('throughputChart');
        if (!ctx) return;

        this.charts.throughput = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    label: 'Requests/sec',
                    data: [],
                    backgroundColor: 'rgba(54, 162, 235, 0.8)'
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    /**
     * Initialize error rate chart
     */
    initializeErrorRateChart() {
        const ctx = document.getElementById('errorRateChart');
        if (!ctx) return;

        this.charts.errorRate = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Error Rate (%)',
                    data: [],
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100
                    }
                }
            }
        });
    }

    /**
     * Update performance charts
     */
    updatePerformanceCharts() {
        if (!this.performanceData || !this.performanceData.time_series) return;

        const timeSeries = this.performanceData.time_series;
        const labels = timeSeries.map(point => new Date(point.timestamp).toLocaleTimeString());
        
        // Update response time chart
        if (this.charts.responseTime) {
            this.charts.responseTime.data.labels = labels;
            this.charts.responseTime.data.datasets[0].data = timeSeries.map(point => point.response_time);
            this.charts.responseTime.update();
        }
        
        // Update throughput chart
        if (this.charts.throughput) {
            this.charts.throughput.data.labels = labels;
            this.charts.throughput.data.datasets[0].data = timeSeries.map(point => point.throughput);
            this.charts.throughput.update();
        }
        
        // Update error rate chart
        if (this.charts.errorRate) {
            this.charts.errorRate.data.labels = labels;
            this.charts.errorRate.data.datasets[0].data = timeSeries.map(point => point.error_rate);
            this.charts.errorRate.update();
        }
    }

    /**
     * Get alert icon based on severity
     */
    getAlertIcon(severity) {
        const icons = {
            critical: 'exclamation-triangle',
            high: 'exclamation-circle',
            medium: 'info-circle',
            low: 'info'
        };
        return icons[severity] || 'info';
    }

    /**
     * Get status icon
     */
    getStatusIcon(status) {
        const icons = {
            healthy: 'check-circle',
            warning: 'exclamation-triangle',
            critical: 'times-circle'
        };
        return icons[status] || 'question-circle';
    }

    /**
     * Update element content
     */
    updateElement(id, content) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = content;
        }
    }

    /**
     * Format time ago
     */
    formatTimeAgo(date) {
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) return 'Just now';
        if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
        if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
        if (diffInSeconds < 2592000) return `${Math.floor(diffInSeconds / 86400)}d ago`;
        
        return date.toLocaleDateString();
    }

    /**
     * Show loading state
     */
    showLoadingState() {
        const loadingElement = document.getElementById('loadingIndicator');
        if (loadingElement) {
            loadingElement.style.display = 'block';
        }
    }

    /**
     * Hide loading state
     */
    hideLoadingState() {
        const loadingElement = document.getElementById('loadingIndicator');
        if (loadingElement) {
            loadingElement.style.display = 'none';
        }
    }

    /**
     * Show error state
     */
    showErrorState(message) {
        const errorElement = document.getElementById('errorMessage');
        if (errorElement) {
            errorElement.textContent = message;
            errorElement.style.display = 'block';
        }
        
        this.hideLoadingState();
    }

    /**
     * Bind event handlers
     */
    bindEvents() {
        // Refresh button
        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.refreshData());
        }
        
        // Export button
        const exportBtn = document.getElementById('exportBtn');
        if (exportBtn) {
            exportBtn.addEventListener('click', () => this.exportData());
        }
    }

    /**
     * Refresh data
     */
    async refreshData() {
        if (this.isLoading) return;
        
        try {
            this.dataIntegration.clearCache();
            await this.loadMonitoringData();
        } catch (error) {
            console.error('Failed to refresh data:', error);
            this.showErrorState('Failed to refresh data');
        }
    }

    /**
     * Export data
     */
    exportData() {
        const exportData = {
            metrics: this.metrics,
            alerts: this.alerts,
            healthChecks: this.healthChecks,
            performanceData: this.performanceData,
            exported_at: new Date().toISOString()
        };
        
        const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `monitoring-data-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    /**
     * Acknowledge alert
     */
    async acknowledgeAlert(alertId) {
        try {
            await this.dataIntegration.acknowledgeAlert(alertId);
            await this.refreshData();
        } catch (error) {
            console.error('Failed to acknowledge alert:', error);
            this.showErrorState('Failed to acknowledge alert');
        }
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh() {
        // Refresh every 30 seconds for monitoring data
        this.refreshInterval = setInterval(() => {
            if (!this.isLoading) {
                this.refreshData();
            }
        }, 30 * 1000);
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
     * Cleanup
     */
    destroy() {
        this.stopAutoRefresh();
        this.dataIntegration.clearCache();
        
        // Destroy charts
        Object.values(this.charts).forEach(chart => {
            if (chart && typeof chart.destroy === 'function') {
                chart.destroy();
            }
        });
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.monitoringDashboard = new MonitoringDashboardRealData();
});

// Export for use in other components
window.MonitoringDashboardRealData = MonitoringDashboardRealData;
