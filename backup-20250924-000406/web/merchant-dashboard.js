/**
 * Merchant Dashboard JavaScript
 * Provides comprehensive dashboard functionality with real-time data updates and data visualization
 * Implements responsive design, session management, and interactive features
 */

class MerchantDashboard {
    constructor() {
        this.merchantId = this.getMerchantIdFromUrl();
        this.apiBaseUrl = '/api/v1';
        this.merchant = null;
        this.sessionManager = null;
        this.riskIndicator = null;
        this.comingSoonBanner = null;
        this.mockDataWarning = null;
        this.refreshInterval = null;
        this.charts = {};
        this.realTimeUpdates = true;
        this.updateFrequency = 30000; // 30 seconds
        
        this.init();
    }

    init() {
        this.initializeComponents();
        this.bindEvents();
        this.loadMerchantData();
        this.startRealTimeUpdates();
        this.initializeCharts();
    }

    getMerchantIdFromUrl() {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get('id') || 'demo-merchant-001';
    }

    initializeComponents() {
        // Initialize session manager
        this.sessionManager = new SessionManager({
            container: document.getElementById('sessionManager'),
            onSessionStart: (session) => this.onSessionStart(session),
            onSessionEnd: () => this.onSessionEnd(),
            onSessionSwitch: (session) => this.onSessionSwitch(session)
        });

        // Initialize risk level indicator
        this.riskIndicator = new RiskLevelIndicator({
            container: document.body
        });

        // Initialize coming soon banner
        this.comingSoonBanner = new ComingSoonBanner({
            container: document.getElementById('comingSoonBanner'),
            features: [
                {
                    name: 'Advanced Analytics',
                    description: 'Detailed merchant analytics and insights',
                    timeline: 'Q2 2025'
                },
                {
                    name: 'Real-time Monitoring',
                    description: 'Live merchant activity monitoring',
                    timeline: 'Q2 2025'
                },
                {
                    name: 'Predictive Risk Analysis',
                    description: 'AI-powered risk prediction models',
                    timeline: 'Q3 2025'
                }
            ]
        });

        // Initialize mock data warning
        this.mockDataWarning = new MockDataWarning({
            container: document.getElementById('mockDataWarning'),
            message: 'This merchant dashboard is using mock data for demonstration purposes.'
        });
    }

    bindEvents() {
        // Edit merchant button
        const editBtn = document.getElementById('editMerchantBtn');
        if (editBtn) {
            editBtn.addEventListener('click', () => this.editMerchant());
        }

        // Compare merchant button
        const compareBtn = document.getElementById('compareMerchantBtn');
        if (compareBtn) {
            compareBtn.addEventListener('click', () => this.compareMerchant());
        }

        // Export report button
        const exportBtn = document.getElementById('exportReportBtn');
        if (exportBtn) {
            exportBtn.addEventListener('click', () => this.exportReport());
        }

        // Real-time toggle
        const realTimeToggle = document.getElementById('realTimeToggle');
        if (realTimeToggle) {
            realTimeToggle.addEventListener('change', (e) => {
                this.toggleRealTimeUpdates(e.target.checked);
            });
        }

        // Refresh button
        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.refreshData());
        }

        // Handle browser back/forward
        window.addEventListener('popstate', () => {
            this.merchantId = this.getMerchantIdFromUrl();
            this.loadMerchantData();
        });

        // Handle window focus/blur for real-time updates
        window.addEventListener('focus', () => {
            if (this.realTimeUpdates) {
                this.startRealTimeUpdates();
            }
        });

        window.addEventListener('blur', () => {
            this.stopRealTimeUpdates();
        });

        // Handle visibility change
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                this.stopRealTimeUpdates();
            } else if (this.realTimeUpdates) {
                this.startRealTimeUpdates();
            }
        });
    }

    async loadMerchantData() {
        try {
            this.showLoadingState();
            
            // Load merchant data from API
            const response = await fetch(`${this.apiBaseUrl}/merchants/${this.merchantId}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            this.merchant = await response.json();
            this.renderMerchantData();
            this.loadActivityTimeline();
            this.loadAnalyticsData();
            this.updateCharts();
            
        } catch (error) {
            console.error('Error loading merchant data:', error);
            this.showErrorState('Failed to load merchant data. Please try again.');
            this.loadMockData();
        }
    }

    loadMockData() {
        // Load mock data for demonstration
        this.merchant = {
            id: this.merchantId,
            name: 'Acme Corporation',
            legal_name: 'Acme Corporation Inc.',
            portfolio_type: 'onboarded',
            risk_level: 'medium',
            industry: 'Technology',
            business_type: 'Corporation',
            registration_number: 'REG-2024-001',
            tax_id: 'TAX-123456789',
            industry_code: 'NAICS 541511',
            annual_revenue: 2500000,
            employee_count: 45,
            founded_date: '2018-03-15',
            compliance_status: 'Compliant',
            updated_at: new Date().toISOString(),
            address: {
                street1: '123 Business Ave',
                street2: 'Suite 100',
                city: 'San Francisco',
                state: 'CA',
                postal_code: '94105',
                country: 'USA'
            },
            contact_info: {
                phone: '+1-555-123-4567',
                email: 'contact@acme.com',
                website: 'https://www.acme.com',
                primary_contact: 'John Smith'
            }
        };
        
        this.renderMerchantData();
        this.renderMockActivityTimeline();
        this.loadMockAnalyticsData();
        this.updateCharts();
    }

    showLoadingState() {
        // Show loading indicators
        const nameElement = document.getElementById('merchantNameText');
        if (nameElement) nameElement.textContent = 'Loading...';
        
        const legalNameElement = document.getElementById('merchantLegalName');
        if (legalNameElement) legalNameElement.textContent = 'Loading legal name...';
        
        // Show loading in all info sections
        const loadingElements = document.querySelectorAll('.info-value, .compliance-value');
        loadingElements.forEach(el => {
            el.textContent = 'Loading...';
        });

        // Show loading in timeline
        const timeline = document.getElementById('activityTimeline');
        if (timeline) {
            timeline.innerHTML = `
                <div class="loading">
                    <i class="fas fa-spinner"></i>
                    Loading activity timeline...
                </div>
            `;
        }
    }

    showErrorState(message) {
        // Show error message
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error';
        errorDiv.innerHTML = `
            <i class="fas fa-exclamation-triangle"></i>
            ${message}
        `;
        
        const container = document.querySelector('.container');
        if (container) {
            container.insertBefore(errorDiv, container.firstChild);
        }
    }

    renderMerchantData() {
        if (!this.merchant) return;

        // Update header
        this.updateMerchantHeader();
        
        // Update business information
        this.updateBusinessInfo();
        
        // Update contact information
        this.updateContactInfo();
        
        // Update risk assessment
        this.updateRiskAssessment();
        
        // Update compliance overview
        this.updateComplianceOverview();
    }

    updateMerchantHeader() {
        const merchant = this.merchant;
        
        // Update merchant name and avatar
        const nameElement = document.getElementById('merchantNameText');
        if (nameElement) nameElement.textContent = merchant.name;
        
        const legalNameElement = document.getElementById('merchantLegalName');
        if (legalNameElement) legalNameElement.textContent = merchant.legal_name || merchant.name;
        
        // Update avatar with first letter
        const avatar = document.getElementById('merchantAvatar');
        if (avatar) {
            avatar.textContent = merchant.name.charAt(0).toUpperCase();
        }
        
        // Update portfolio type badge
        const portfolioTypeBadge = document.getElementById('portfolioTypeBadge');
        if (portfolioTypeBadge) {
            portfolioTypeBadge.textContent = merchant.portfolio_type;
            portfolioTypeBadge.className = `status-badge status-${merchant.portfolio_type}`;
        }
        
        // Update meta information
        const industryElement = document.getElementById('merchantIndustry');
        if (industryElement) industryElement.textContent = merchant.industry || 'Not specified';
        
        const locationElement = document.getElementById('merchantLocation');
        if (locationElement && merchant.address) {
            locationElement.textContent = `${merchant.address.city}, ${merchant.address.state}`;
        }
        
        const foundedElement = document.getElementById('merchantFounded');
        if (foundedElement) {
            foundedElement.textContent = merchant.founded_date ? 
                new Date(merchant.founded_date).getFullYear() : 'Unknown';
        }
        
        const employeesElement = document.getElementById('merchantEmployees');
        if (employeesElement) {
            employeesElement.textContent = merchant.employee_count ? 
                `${merchant.employee_count} employees` : 'Unknown';
        }
    }

    updateBusinessInfo() {
        const merchant = this.merchant;
        
        const elements = {
            'registrationNumber': merchant.registration_number || 'Not provided',
            'taxId': merchant.tax_id || 'Not provided',
            'businessType': merchant.business_type || 'Not specified',
            'industryCode': merchant.industry_code || 'Not specified',
            'annualRevenue': merchant.annual_revenue ? `$${merchant.annual_revenue.toLocaleString()}` : 'Not disclosed',
            'employeeCount': merchant.employee_count ? merchant.employee_count.toLocaleString() : 'Not disclosed'
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) element.textContent = value;
        });
    }

    updateContactInfo() {
        const merchant = this.merchant;
        
        if (merchant.contact_info) {
            const elements = {
                'contactPhone': merchant.contact_info.phone || 'Not provided',
                'contactEmail': merchant.contact_info.email || 'Not provided',
                'contactWebsite': merchant.contact_info.website || 'Not provided',
                'primaryContact': merchant.contact_info.primary_contact || 'Not specified'
            };

            Object.entries(elements).forEach(([id, value]) => {
                const element = document.getElementById(id);
                if (element) element.textContent = value;
            });
        }
        
        // Update full address
        const fullAddressElement = document.getElementById('fullAddress');
        if (fullAddressElement && merchant.address) {
            const address = merchant.address;
            const fullAddress = [
                address.street1,
                address.street2,
                address.city,
                address.state,
                address.postal_code,
                address.country
            ].filter(Boolean).join(', ');
            
            fullAddressElement.textContent = fullAddress || 'Address not provided';
        }
    }

    updateRiskAssessment() {
        const merchant = this.merchant;
        
        // Update risk level badge
        const riskLevelBadge = document.getElementById('riskLevelBadge');
        if (riskLevelBadge) {
            riskLevelBadge.textContent = merchant.risk_level;
            riskLevelBadge.className = `status-badge risk-${merchant.risk_level}`;
        }
        
        const elements = {
            'complianceStatus': merchant.compliance_status || 'Pending',
            'lastAssessment': merchant.updated_at ? new Date(merchant.updated_at).toLocaleDateString() : 'Never',
            'nextReview': this.calculateNextReviewDate(merchant.updated_at)
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) element.textContent = value;
        });
    }

    updateComplianceOverview() {
        const merchant = this.merchant;
        
        // Mock compliance data - in real implementation, this would come from API
        const elements = {
            'kycStatus': 'Completed',
            'amlStatus': 'Passed',
            'documentationStatus': 'Complete',
            'lastComplianceReview': merchant.updated_at ? new Date(merchant.updated_at).toLocaleDateString() : 'Never'
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) element.textContent = value;
        });
    }

    async loadActivityTimeline() {
        try {
            // Load activity timeline from API
            const response = await fetch(`${this.apiBaseUrl}/merchants/${this.merchantId}/activity`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const activities = await response.json();
            this.renderActivityTimeline(activities);
            
        } catch (error) {
            console.error('Error loading activity timeline:', error);
            this.renderMockActivityTimeline();
        }
    }

    renderActivityTimeline(activities) {
        const timeline = document.getElementById('activityTimeline');
        if (!timeline) return;
        
        if (!activities || activities.length === 0) {
            timeline.innerHTML = `
                <div class="timeline-item">
                    <div class="timeline-date">No activity recorded</div>
                    <div class="timeline-title">No recent activity</div>
                    <div class="timeline-description">This merchant has no recorded activity yet.</div>
                </div>
            `;
            return;
        }
        
        timeline.innerHTML = activities.map(activity => `
            <div class="timeline-item">
                <div class="timeline-date">${new Date(activity.timestamp).toLocaleDateString()}</div>
                <div class="timeline-title">${activity.title}</div>
                <div class="timeline-description">${activity.description}</div>
            </div>
        `).join('');
    }

    renderMockActivityTimeline() {
        const timeline = document.getElementById('activityTimeline');
        if (!timeline) return;
        
        const mockActivities = [
            {
                timestamp: new Date().toISOString(),
                title: 'Merchant Profile Updated',
                description: 'Business information and contact details were updated.'
            },
            {
                timestamp: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                title: 'Risk Assessment Completed',
                description: 'Risk level assessment was completed and updated to current level.'
            },
            {
                timestamp: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
                title: 'Compliance Review',
                description: 'Annual compliance review was completed successfully.'
            },
            {
                timestamp: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
                title: 'Merchant Onboarded',
                description: 'Merchant was successfully onboarded to the platform.'
            }
        ];
        
        this.renderActivityTimeline(mockActivities);
    }

    async loadAnalyticsData() {
        try {
            // Load analytics data from API
            const response = await fetch(`${this.apiBaseUrl}/merchants/${this.merchantId}/analytics`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const analytics = await response.json();
            this.renderAnalyticsData(analytics);
            
        } catch (error) {
            console.error('Error loading analytics data:', error);
            this.loadMockAnalyticsData();
        }
    }

    loadMockAnalyticsData() {
        // Mock analytics data for demonstration
        const mockAnalytics = {
            riskTrend: [
                { date: '2024-01-01', value: 0.3 },
                { date: '2024-02-01', value: 0.4 },
                { date: '2024-03-01', value: 0.35 },
                { date: '2024-04-01', value: 0.5 },
                { date: '2024-05-01', value: 0.45 },
                { date: '2024-06-01', value: 0.4 }
            ],
            complianceScore: 85,
            transactionVolume: 125000,
            monthlyGrowth: 12.5
        };
        
        this.renderAnalyticsData(mockAnalytics);
    }

    renderAnalyticsData(analytics) {
        // Update analytics displays
        const complianceScoreElement = document.getElementById('complianceScore');
        if (complianceScoreElement) {
            complianceScoreElement.textContent = `${analytics.complianceScore}%`;
        }
        
        const transactionVolumeElement = document.getElementById('transactionVolume');
        if (transactionVolumeElement) {
            transactionVolumeElement.textContent = `$${analytics.transactionVolume.toLocaleString()}`;
        }
        
        const monthlyGrowthElement = document.getElementById('monthlyGrowth');
        if (monthlyGrowthElement) {
            monthlyGrowthElement.textContent = `${analytics.monthlyGrowth}%`;
        }
    }

    initializeCharts() {
        // Initialize Chart.js if available
        if (typeof Chart !== 'undefined') {
            this.initializeRiskTrendChart();
            this.initializeComplianceChart();
            this.initializeTransactionChart();
        }
    }

    initializeRiskTrendChart() {
        const ctx = document.getElementById('riskTrendChart');
        if (!ctx) return;
        
        this.charts.riskTrend = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Risk Level',
                    data: [],
                    borderColor: '#e74c3c',
                    backgroundColor: 'rgba(231, 76, 60, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 1
                    }
                }
            }
        });
    }

    initializeComplianceChart() {
        const ctx = document.getElementById('complianceChart');
        if (!ctx) return;
        
        this.charts.compliance = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Compliant', 'Pending', 'Non-compliant'],
                datasets: [{
                    data: [85, 10, 5],
                    backgroundColor: ['#27ae60', '#f39c12', '#e74c3c']
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false
            }
        });
    }

    initializeTransactionChart() {
        const ctx = document.getElementById('transactionChart');
        if (!ctx) return;
        
        this.charts.transaction = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
                datasets: [{
                    label: 'Transaction Volume',
                    data: [100000, 110000, 105000, 120000, 115000, 125000],
                    backgroundColor: '#3498db'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    updateCharts() {
        // Update charts with new data
        if (this.charts.riskTrend) {
            // Update risk trend chart
            const riskData = this.generateMockRiskData();
            this.charts.riskTrend.data.labels = riskData.labels;
            this.charts.riskTrend.data.datasets[0].data = riskData.values;
            this.charts.riskTrend.update();
        }
    }

    generateMockRiskData() {
        const labels = [];
        const values = [];
        const now = new Date();
        
        for (let i = 5; i >= 0; i--) {
            const date = new Date(now.getTime() - i * 30 * 24 * 60 * 60 * 1000);
            labels.push(date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }));
            values.push(Math.random() * 0.5 + 0.2); // Random risk values between 0.2 and 0.7
        }
        
        return { labels, values };
    }

    startRealTimeUpdates() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
        
        this.refreshInterval = setInterval(() => {
            this.refreshData();
        }, this.updateFrequency);
    }

    stopRealTimeUpdates() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    toggleRealTimeUpdates(enabled) {
        this.realTimeUpdates = enabled;
        
        if (enabled) {
            this.startRealTimeUpdates();
        } else {
            this.stopRealTimeUpdates();
        }
    }

    async refreshData() {
        try {
            // Refresh merchant data
            await this.loadMerchantData();
            
            // Show refresh indicator
            this.showRefreshIndicator();
            
        } catch (error) {
            console.error('Error refreshing data:', error);
        }
    }

    showRefreshIndicator() {
        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            const originalText = refreshBtn.innerHTML;
            refreshBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Refreshing...';
            refreshBtn.disabled = true;
            
            setTimeout(() => {
                refreshBtn.innerHTML = originalText;
                refreshBtn.disabled = false;
            }, 2000);
        }
    }

    calculateNextReviewDate(lastUpdate) {
        if (!lastUpdate) return 'Not scheduled';
        
        const lastDate = new Date(lastUpdate);
        const nextDate = new Date(lastDate.getTime() + 365 * 24 * 60 * 60 * 1000); // 1 year
        return nextDate.toLocaleDateString();
    }

    editMerchant() {
        // Navigate to edit merchant page
        window.location.href = `merchant-edit.html?id=${this.merchantId}`;
    }

    compareMerchant() {
        // Navigate to comparison page with current merchant
        window.location.href = `merchant-comparison.html?merchant1=${this.merchantId}`;
    }

    exportReport() {
        // Generate and download merchant report
        const reportData = {
            merchant: this.merchant,
            generatedAt: new Date().toISOString(),
            reportType: 'merchant_dashboard',
            analytics: this.getAnalyticsData()
        };
        
        const blob = new Blob([JSON.stringify(reportData, null, 2)], {
            type: 'application/json'
        });
        
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `merchant-dashboard-report-${this.merchantId}-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    getAnalyticsData() {
        // Get current analytics data for export
        return {
            complianceScore: document.getElementById('complianceScore')?.textContent || 'N/A',
            transactionVolume: document.getElementById('transactionVolume')?.textContent || 'N/A',
            monthlyGrowth: document.getElementById('monthlyGrowth')?.textContent || 'N/A',
            lastUpdated: new Date().toISOString()
        };
    }

    onSessionStart(session) {
        console.log('Session started:', session);
        // Update UI to reflect active session
        this.updateSessionIndicator(true);
    }

    onSessionEnd() {
        console.log('Session ended');
        // Update UI to reflect no active session
        this.updateSessionIndicator(false);
    }

    onSessionSwitch(session) {
        console.log('Session switched:', session);
        // Update merchant data if different merchant
        if (session.merchantId !== this.merchantId) {
            this.merchantId = session.merchantId;
            this.loadMerchantData();
        }
    }

    updateSessionIndicator(isActive) {
        const sessionIndicator = document.getElementById('sessionIndicator');
        if (sessionIndicator) {
            sessionIndicator.className = isActive ? 'session-active' : 'session-inactive';
            sessionIndicator.textContent = isActive ? 'Session Active' : 'No Active Session';
        }
    }

    // Cleanup method
    destroy() {
        this.stopRealTimeUpdates();
        
        // Destroy charts
        Object.values(this.charts).forEach(chart => {
            if (chart && typeof chart.destroy === 'function') {
                chart.destroy();
            }
        });
        
        this.charts = {};
    }
}

// Utility functions for dashboard
class DashboardUtils {
    static formatCurrency(amount) {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD'
        }).format(amount);
    }

    static formatDate(date) {
        return new Date(date).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    }

    static formatDateTime(date) {
        return new Date(date).toLocaleString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    static calculateRiskColor(riskLevel) {
        const colors = {
            'low': '#27ae60',
            'medium': '#f39c12',
            'high': '#e74c3c'
        };
        return colors[riskLevel] || '#95a5a6';
    }

    static calculatePortfolioTypeColor(portfolioType) {
        const colors = {
            'onboarded': '#27ae60',
            'deactivated': '#e74c3c',
            'prospective': '#f39c12',
            'pending': '#3498db'
        };
        return colors[portfolioType] || '#95a5a6';
    }
}

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.merchantDashboard = new MerchantDashboard();
});

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (window.merchantDashboard) {
        window.merchantDashboard.destroy();
    }
});
