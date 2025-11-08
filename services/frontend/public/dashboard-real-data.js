/**
 * Main Dashboard with Real Data Integration
 * Replaces mock data with real Supabase API calls
 * Provides comprehensive business intelligence with live data
 */

class DashboardRealData {
    constructor() {
        this.dataIntegration = new RealDataIntegration();
        this.businessIntelligence = null;
        this.analytics = null;
        this.statistics = null;
        this.recentActivity = null;
        this.refreshInterval = null;
        this.isLoading = false;
        this.charts = {};
        this.crossTabNavigation = null;
        this.dashboardUtils = null;
        
        this.init();
    }

    /**
     * Initialize the dashboard
     */
    async init() {
        try {
            this.showLoadingState();
            await this.loadDashboardData();
            this.initializeCharts();
            this.bindEvents();
            this.startAutoRefresh();
            await this.initializeDashboardUtilities();
        } catch (error) {
            console.error('Failed to initialize dashboard:', error);
            this.showErrorState(error.message);
        }
    }

    /**
     * Load dashboard data from Supabase
     */
    async loadDashboardData() {
        try {
            this.isLoading = true;
            
            // Load all dashboard data in parallel
            const [businessIntelligence, analytics, statistics, recentActivity] = await Promise.all([
                this.dataIntegration.getBusinessIntelligence(),
                this.dataIntegration.getMerchantAnalytics(),
                this.dataIntegration.getMerchantStatistics(),
                this.dataIntegration.getRecentActivity()
            ]);
            
            this.businessIntelligence = businessIntelligence;
            this.analytics = analytics;
            this.statistics = statistics;
            this.recentActivity = recentActivity;
            
            // Render all data
            this.renderBusinessIntelligence();
            this.renderAnalyticsData();
            this.renderStatisticsData();
            this.renderRecentActivity();
            this.updateQuickStats();
            
            this.hideLoadingState();
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            this.showErrorState('Failed to load dashboard data');
        } finally {
            this.isLoading = false;
        }
    }

    /**
     * Render business intelligence data
     */
    renderBusinessIntelligence() {
        if (!this.businessIntelligence) return;

        // Key metrics
        this.updateElement('totalMerchants', this.businessIntelligence.total_merchants || 0);
        this.updateElement('activeMerchants', this.businessIntelligence.active_merchants || 0);
        this.updateElement('pendingMerchants', this.businessIntelligence.pending_merchants || 0);
        this.updateElement('suspendedMerchants', this.businessIntelligence.suspended_merchants || 0);
        
        // Revenue metrics
        // Use DashboardUtils if available
        const formatCurrency = this.dashboardUtils?.formatCurrency || this.formatCurrency.bind(this);
        this.updateElement('totalRevenue', formatCurrency(this.businessIntelligence.total_revenue || 0));
        this.updateElement('averageRevenue', formatCurrency(this.businessIntelligence.average_revenue || 0));
        this.updateElement('monthlyGrowth', `${this.businessIntelligence.monthly_growth || 0}%`);
        
        // Risk metrics
        this.updateElement('highRiskMerchants', this.businessIntelligence.high_risk_merchants || 0);
        this.updateElement('mediumRiskMerchants', this.businessIntelligence.medium_risk_merchants || 0);
        this.updateElement('lowRiskMerchants', this.businessIntelligence.low_risk_merchants || 0);
        
        // Performance metrics
        this.updateElement('verificationRate', `${this.businessIntelligence.verification_rate || 0}%`);
        this.updateElement('complianceScore', `${this.businessIntelligence.compliance_score || 0}%`);
        this.updateElement('averageProcessingTime', `${this.businessIntelligence.average_processing_time || 0} hours`);
    }

    /**
     * Render analytics data
     */
    renderAnalyticsData() {
        if (!this.analytics) return;

        // Portfolio distribution
        if (this.analytics.portfolio_distribution) {
            this.renderPortfolioDistribution(this.analytics.portfolio_distribution);
        }
        
        // Risk distribution
        if (this.analytics.risk_distribution) {
            this.renderRiskDistribution(this.analytics.risk_distribution);
        }
        
        // Industry distribution
        if (this.analytics.industry_distribution) {
            this.renderIndustryDistribution(this.analytics.industry_distribution);
        }
        
        // Geographic distribution
        if (this.analytics.geographic_distribution) {
            this.renderGeographicDistribution(this.analytics.geographic_distribution);
        }
    }

    /**
     * Render statistics data
     */
    renderStatisticsData() {
        if (!this.statistics) return;

        // Update statistics cards
        this.updateElement('totalClassifications', this.statistics.total_classifications || 0);
        this.updateElement('successfulClassifications', this.statistics.successful_classifications || 0);
        this.updateElement('failedClassifications', this.statistics.failed_classifications || 0);
        this.updateElement('averageConfidence', `${this.statistics.average_confidence || 0}%`);
        
        // Update performance metrics
        this.updateElement('systemUptime', `${this.statistics.system_uptime || 0}%`);
        this.updateElement('averageResponseTime', `${this.statistics.average_response_time || 0}ms`);
        this.updateElement('errorRate', `${this.statistics.error_rate || 0}%`);
    }

    /**
     * Render recent activity
     */
    renderRecentActivity() {
        if (!this.recentActivity) return;

        const activityContainer = document.getElementById('recentActivityList');
        if (!activityContainer) return;

        activityContainer.innerHTML = this.recentActivity.map(activity => `
            <div class="activity-item">
                <div class="activity-icon ${activity.type}">
                    <i class="fas fa-${this.getActivityIcon(activity.type)}"></i>
                </div>
                <div class="activity-content">
                    <div class="activity-title">${activity.title}</div>
                    <div class="activity-description">${activity.description}</div>
                    <div class="activity-meta">
                        <span class="activity-source">${activity.source}</span>
                        <span class="activity-time">${this.formatTimeAgo(new Date(activity.timestamp))}</span>
                    </div>
                </div>
            </div>
        `).join('');
    }

    /**
     * Update quick stats
     */
    updateQuickStats() {
        // Update quick stat cards with real data
        const quickStats = [
            { id: 'quickStat1', label: 'Active Merchants', value: this.businessIntelligence?.active_merchants || 0 },
            { id: 'quickStat2', label: 'Total Revenue', value: this.formatCurrency(this.businessIntelligence?.total_revenue || 0) },
            { id: 'quickStat3', label: 'Verification Rate', value: `${this.businessIntelligence?.verification_rate || 0}%` },
            { id: 'quickStat4', label: 'System Health', value: `${this.statistics?.system_uptime || 0}%` }
        ];

        quickStats.forEach(stat => {
            const element = document.getElementById(stat.id);
            if (element) {
                element.innerHTML = `
                    <div class="quick-stat-value">${stat.value}</div>
                    <div class="quick-stat-label">${stat.label}</div>
                `;
            }
        });
    }

    /**
     * Render portfolio distribution chart
     */
    renderPortfolioDistribution(distribution) {
        const container = document.getElementById('portfolioDistribution');
        if (!container) return;

        const total = Object.values(distribution).reduce((sum, count) => sum + count, 0);
        
        container.innerHTML = Object.entries(distribution).map(([type, count]) => {
            const percentage = total > 0 ? (count / total * 100).toFixed(1) : 0;
            return `
                <div class="distribution-item">
                    <div class="distribution-label">${type}</div>
                    <div class="distribution-bar">
                        <div class="distribution-fill" style="width: ${percentage}%"></div>
                    </div>
                    <div class="distribution-value">${count} (${percentage}%)</div>
                </div>
            `;
        }).join('');
    }

    /**
     * Render risk distribution chart
     */
    renderRiskDistribution(distribution) {
        const container = document.getElementById('riskDistribution');
        if (!container) return;

        const total = Object.values(distribution).reduce((sum, count) => sum + count, 0);
        
        container.innerHTML = Object.entries(distribution).map(([level, count]) => {
            const percentage = total > 0 ? (count / total * 100).toFixed(1) : 0;
            return `
                <div class="distribution-item">
                    <div class="distribution-label risk-${level.toLowerCase()}">${level}</div>
                    <div class="distribution-bar">
                        <div class="distribution-fill risk-${level.toLowerCase()}" style="width: ${percentage}%"></div>
                    </div>
                    <div class="distribution-value">${count} (${percentage}%)</div>
                </div>
            `;
        }).join('');
    }

    /**
     * Render industry distribution chart
     */
    renderIndustryDistribution(distribution) {
        const container = document.getElementById('industryDistribution');
        if (!container) return;

        const total = Object.values(distribution).reduce((sum, count) => sum + count, 0);
        
        container.innerHTML = Object.entries(distribution).map(([industry, count]) => {
            const percentage = total > 0 ? (count / total * 100).toFixed(1) : 0;
            return `
                <div class="distribution-item">
                    <div class="distribution-label">${industry}</div>
                    <div class="distribution-bar">
                        <div class="distribution-fill" style="width: ${percentage}%"></div>
                    </div>
                    <div class="distribution-value">${count} (${percentage}%)</div>
                </div>
            `;
        }).join('');
    }

    /**
     * Render geographic distribution chart
     */
    renderGeographicDistribution(distribution) {
        const container = document.getElementById('geographicDistribution');
        if (!container) return;

        const total = Object.values(distribution).reduce((sum, count) => sum + count, 0);
        
        container.innerHTML = Object.entries(distribution).map(([location, count]) => {
            const percentage = total > 0 ? (count / total * 100).toFixed(1) : 0;
            return `
                <div class="distribution-item">
                    <div class="distribution-label">${location}</div>
                    <div class="distribution-bar">
                        <div class="distribution-fill" style="width: ${percentage}%"></div>
                    </div>
                    <div class="distribution-value">${count} (${percentage}%)</div>
                </div>
            `;
        }).join('');
    }

    /**
     * Initialize charts
     */
    initializeCharts() {
        // Initialize charts if Chart.js is available
        if (typeof Chart !== 'undefined') {
            this.initializeRevenueChart();
            this.initializeMerchantGrowthChart();
            this.initializeRiskTrendChart();
        }
    }

    /**
     * Initialize revenue chart
     */
    initializeRevenueChart() {
        const ctx = document.getElementById('revenueChart');
        if (!ctx) return;

        this.charts.revenue = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Monthly Revenue',
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
                        beginAtZero: true,
                        ticks: {
                            callback: function(value) {
                                return '$' + value.toLocaleString();
                            }
                        }
                    }
                }
            }
        });
    }

    /**
     * Initialize merchant growth chart
     */
    initializeMerchantGrowthChart() {
        const ctx = document.getElementById('merchantGrowthChart');
        if (!ctx) return;

        this.charts.merchantGrowth = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    label: 'New Merchants',
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
     * Initialize risk trend chart
     */
    initializeRiskTrendChart() {
        const ctx = document.getElementById('riskTrendChart');
        if (!ctx) return;

        this.charts.riskTrend = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [
                    {
                        label: 'High Risk',
                        data: [],
                        borderColor: 'rgb(255, 99, 132)',
                        backgroundColor: 'rgba(255, 99, 132, 0.2)',
                        tension: 0.1
                    },
                    {
                        label: 'Medium Risk',
                        data: [],
                        borderColor: 'rgb(255, 205, 86)',
                        backgroundColor: 'rgba(255, 205, 86, 0.2)',
                        tension: 0.1
                    },
                    {
                        label: 'Low Risk',
                        data: [],
                        borderColor: 'rgb(75, 192, 192)',
                        backgroundColor: 'rgba(75, 192, 192, 0.2)',
                        tension: 0.1
                    }
                ]
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
     * Update charts with real data
     */
    updateCharts() {
        if (!this.businessIntelligence || !this.businessIntelligence.time_series) return;

        const timeSeries = this.businessIntelligence.time_series;
        const labels = timeSeries.map(point => new Date(point.timestamp).toLocaleDateString());
        
        // Update revenue chart
        if (this.charts.revenue) {
            this.charts.revenue.data.labels = labels;
            this.charts.revenue.data.datasets[0].data = timeSeries.map(point => point.revenue);
            this.charts.revenue.update();
        }
        
        // Update merchant growth chart
        if (this.charts.merchantGrowth) {
            this.charts.merchantGrowth.data.labels = labels;
            this.charts.merchantGrowth.data.datasets[0].data = timeSeries.map(point => point.new_merchants);
            this.charts.merchantGrowth.update();
        }
        
        // Update risk trend chart
        if (this.charts.riskTrend) {
            this.charts.riskTrend.data.labels = labels;
            this.charts.riskTrend.data.datasets[0].data = timeSeries.map(point => point.high_risk);
            this.charts.riskTrend.data.datasets[1].data = timeSeries.map(point => point.medium_risk);
            this.charts.riskTrend.data.datasets[2].data = timeSeries.map(point => point.low_risk);
            this.charts.riskTrend.update();
        }
    }

    /**
     * Get activity icon based on type
     */
    getActivityIcon(type) {
        const icons = {
            'merchant_created': 'user-plus',
            'merchant_updated': 'user-edit',
            'verification_completed': 'check-circle',
            'risk_assessment': 'shield-alt',
            'system_alert': 'exclamation-triangle',
            'compliance_check': 'clipboard-check'
        };
        return icons[type] || 'info-circle';
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
     * Format currency values
     */
    formatCurrency(amount) {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0
        }).format(amount);
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
        
        // Quick action buttons
        const quickActionBtns = document.querySelectorAll('.quick-action-btn');
        quickActionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const action = e.target.dataset.action;
                this.handleQuickAction(action);
            });
        });
    }

    /**
     * Handle quick actions
     */
    handleQuickAction(action) {
        switch (action) {
            case 'view_merchants':
                window.location.href = 'merchant-bulk-operations.html';
                break;
            case 'view_monitoring':
                window.location.href = 'monitoring-dashboard.html';
                break;
            case 'run_verification':
                this.runBulkVerification();
                break;
            case 'export_data':
                this.exportData();
                break;
        }
    }

    /**
     * Run bulk verification
     */
    async runBulkVerification() {
        try {
            this.showLoadingState();
            
            const result = await this.dataIntegration.runBulkVerification();
            
            this.showNotification(`Bulk verification completed. ${result.processed} merchants processed.`, 'success');
            await this.loadDashboardData();
        } catch (error) {
            console.error('Failed to run bulk verification:', error);
            this.showNotification('Failed to run bulk verification', 'error');
        } finally {
            this.hideLoadingState();
        }
    }

    /**
     * Refresh data
     */
    async refreshData() {
        if (this.isLoading) return;
        
        try {
            this.dataIntegration.clearCache();
            await this.loadDashboardData();
            this.updateCharts();
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
            businessIntelligence: this.businessIntelligence,
            analytics: this.analytics,
            statistics: this.statistics,
            recentActivity: this.recentActivity,
            exported_at: new Date().toISOString()
        };
        
        const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `dashboard-data-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    /**
     * Show notification
     */
    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh() {
        // Refresh every 5 minutes
        this.refreshInterval = setInterval(() => {
            if (!this.isLoading) {
                this.refreshData();
            }
        }, 5 * 60 * 1000);
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
     * Initialize dashboard utilities
     */
    async initializeDashboardUtilities() {
        try {
            // Load shared components
            if (typeof loadSharedComponents === 'function') {
                await loadSharedComponents();
            }

            // Initialize cross-tab navigation
            if (typeof getCrossTabNavigation !== 'undefined') {
                this.crossTabNavigation = getCrossTabNavigation();
            } else if (typeof window !== 'undefined' && window.getCrossTabNavigation) {
                this.crossTabNavigation = window.getCrossTabNavigation();
            }

            // Initialize dashboard utils
            if (typeof DashboardUtils !== 'undefined') {
                this.dashboardUtils = DashboardUtils;
            } else if (typeof window !== 'undefined' && window.DashboardUtils) {
                this.dashboardUtils = window.DashboardUtils;
            }

            // Make refreshDashboard available globally
            window.refreshDashboard = () => {
                this.loadDashboardData();
            };

            console.log('✅ Dashboard utilities initialized');
        } catch (error) {
            console.error('Error initializing dashboard utilities:', error);
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
    window.dashboard = new DashboardRealData();
});

// Export for use in other components
window.DashboardRealData = DashboardRealData;
