/**
 * Merchant Risk Indicators Tab Controller
 * 
 * Main controller that wires existing risk visualization components together
 * to create a fully functional Risk Indicators tab. Leverages:
 * - RiskVisualization component (radar charts, gauges, trends)
 * - RiskExplainability component (SHAP plots, feature importance)
 * - RiskLevelIndicator component (badges, colors, icons)
 * - RiskIndicatorsUITemplate (HTML templates)
 * - RiskIndicatorsDataService (data aggregation)
 */

class MerchantRiskIndicatorsTab {
    constructor(containerId = 'risk-indicators') {
        this.containerId = containerId;
        this.merchantId = null;
        this.riskData = null;
        this.isInitialized = false;
        
        // Reuse existing components (initialize later when D3.js is available)
        this.visualization = null;
        this.explainability = new RiskExplainability();
        this.levelIndicator = null; // Initialize later to avoid DOM manipulation
        this.dataService = null; // Will be initialized with service getter
        
        // UI template helper
        this.uiTemplate = RiskIndicatorsUITemplate;
        this.helpers = RiskIndicatorsHelpers;
        
        // Event handlers
        this.eventHandlers = new Map();
        
        console.log('🎯 Risk Indicators Tab initialized');
    }
    
    /**
     * Initialize visualization component when D3.js is available
     */
    initializeVisualization() {
        if (typeof d3 !== 'undefined' && !this.visualization) {
            try {
                this.visualization = new RiskVisualization();
                console.log('✅ RiskVisualization component initialized with D3.js');
                return true;
            } catch (error) {
                console.error('❌ Failed to initialize RiskVisualization:', error);
                return false;
            }
        }
        return false;
    }
    
    /**
     * Initialize level indicator component when needed
     */
    initializeLevelIndicator(container) {
        if (!this.levelIndicator && container) {
            try {
                this.levelIndicator = new RiskLevelIndicator({ container: container });
                console.log('✅ RiskLevelIndicator component initialized');
                return true;
            } catch (error) {
                console.error('❌ Failed to initialize RiskLevelIndicator:', error);
                return false;
            }
        }
        return false;
    }
    
    /**
     * Initialize the risk indicators tab
     * @param {string} merchantId - Merchant ID
     */
    async init(merchantId) {
        if (this.isInitialized && this.merchantId === merchantId) {
            console.log('🔄 Risk Indicators Tab already initialized for this merchant');
            return;
        }
        
        // Initialize data service using service getter
        await this.initializeDataService();
        
        this.merchantId = merchantId;
        console.log(`🚀 Initializing Risk Indicators Tab for merchant: ${merchantId}`);
        
        try {
            await this.loadAndRender();
            this.bindEventHandlers();
            this.isInitialized = true;
            console.log('✅ Risk Indicators Tab initialized successfully');
        } catch (error) {
            console.error('❌ Failed to initialize Risk Indicators Tab:', error);
            this.showError(error);
        }
    }
    
    /**
     * Initialize data service using service getter
     */
    async initializeDataService() {
        try {
            // Ensure shared components are loaded
            if (typeof loadSharedComponents === 'function') {
                await loadSharedComponents();
            }
            
            // Try to get risk data service from getter
            if (typeof getRiskDataService !== 'undefined') {
                this.dataService = getRiskDataService();
            } else if (typeof window !== 'undefined' && window.getRiskDataService) {
                this.dataService = window.getRiskDataService();
            } else {
                // Fallback to direct instantiation
                if (typeof RiskIndicatorsDataService !== 'undefined') {
                    this.dataService = new RiskIndicatorsDataService();
                } else {
                    console.warn('RiskIndicatorsDataService not available, using fallback');
                    this.dataService = null;
                }
            }
        } catch (error) {
            console.error('Error initializing data service:', error);
            // Fallback to direct instantiation
            if (typeof RiskIndicatorsDataService !== 'undefined') {
                this.dataService = new RiskIndicatorsDataService();
            }
        }
    }
    
    /**
     * Load data and render the UI
     */
    async loadAndRender() {
        try {
            // Show loading state
            this.showLoading();
            
            // Load all data via data service
            if (!this.dataService) {
                throw new Error('Data service not initialized');
            }
            this.riskData = await this.dataService.loadAllRiskData(this.merchantId);
            
            // Render UI
            this.render();
            
            // Initialize visualizations using existing components
            await this.initializeVisualizations();
            
            // Hide loading state
            this.hideLoading();
            
        } catch (error) {
            console.error('Failed to load and render risk indicators:', error);
            this.showError(error);
        }
    }
    
    /**
     * Render the main UI using templates
     */
    render() {
        const container = document.getElementById(this.containerId);
        if (!container) {
            throw new Error(`Container with ID '${this.containerId}' not found`);
        }
        
        container.innerHTML = `
            <div class="risk-indicators-container">
                <!-- Loading State -->
                <div id="riskIndicatorsLoading" class="loading-state hidden">
                    <div class="flex items-center justify-center py-12">
                        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                        <span class="ml-3 text-gray-600">Loading risk indicators...</span>
                    </div>
                </div>
                
                <!-- Error State -->
                <div id="riskIndicatorsError" class="error-state hidden">
                    <div class="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
                        <i class="fas fa-exclamation-triangle text-red-500 text-2xl mb-3"></i>
                        <h3 class="text-lg font-semibold text-red-800 mb-2">Failed to Load Risk Indicators</h3>
                        <p class="text-red-600 mb-4">There was an error loading the risk data. Please try again.</p>
                        <button onclick="riskIndicatorsTab.retry()" class="btn btn-primary">
                            <i class="fas fa-redo mr-2"></i>
                            Retry
                        </button>
                    </div>
                </div>
                
                <!-- Main Content -->
                <div id="riskIndicatorsContent" class="risk-indicators-content">
                    <!-- Alert Cards at Top -->
                    <div id="riskAlerts" class="mb-6">
                        ${this.uiTemplate.getAlertsHTML(this.riskData.alerts)}
                    </div>
                    
                    <!-- Risk Badges Section -->
                    <div id="riskBadges" class="mb-6">
                        ${this.uiTemplate.getRiskBadgesHTML(this.riskData)}
                    </div>
                    
                    <!-- Heat Map Section -->
                    <div id="riskHeatMap" class="mb-6">
                        ${this.uiTemplate.getHeatMapHTML(this.riskData.categories)}
                    </div>
                    
                    <!-- Progress Bars Section -->
                    <div id="riskProgress" class="mb-6">
                        ${this.uiTemplate.getProgressBarsHTML(this.riskData.categories)}
                    </div>
                    
                    <!-- Radar Chart Section -->
                    <div id="riskRadar" class="mb-6">
                        ${this.uiTemplate.getRadarChartHTML()}
                    </div>
                    
                    <!-- Recommendations Section -->
                    <div id="riskRecommendations" class="mb-6">
                        ${this.uiTemplate.getRecommendationsHTML(this.riskData.recommendations)}
                    </div>
                    
                    <!-- Website Risk Findings Section -->
                    <div id="websiteRiskFindings" class="mb-6">
                        ${this.uiTemplate.getWebsiteRiskFindingsHTML(this.riskData.websiteRisks)}
                    </div>
                </div>
            </div>
        `;
        
        console.log('🎨 Risk Indicators UI rendered');
    }
    
    /**
     * Initialize visualizations using existing components
     */
    async initializeVisualizations() {
        try {
            // Initialize radar chart using existing RiskVisualization component
            await this.initializeRadarChart();
            
            // Initialize SHAP analysis using existing RiskExplainability component
            await this.initializeSHAPAnalysis();
            
            // Initialize risk category analysis
            this.initializeRiskCategoryAnalysis();
            
            // Initialize tooltips and interactive elements
            this.initializeTooltips();
            
            console.log('📊 Visualizations initialized successfully');
            
        } catch (error) {
            console.error('Failed to initialize visualizations:', error);
        }
    }
    
    /**
     * Initialize radar chart using existing RiskVisualization component
     */
    async initializeRadarChart() {
        try {
            const radarData = this.prepareRadarData();
            const canvas = document.getElementById('riskRadarChart');
            
            if (canvas && this.visualization.createRiskRadarChart) {
                // Use existing RiskVisualization component
                this.visualization.createRiskRadarChart('riskRadarChart', radarData);
                console.log('📈 Radar chart initialized');
            } else {
                console.warn('Radar chart canvas not found or RiskVisualization not available');
            }
        } catch (error) {
            console.error('Failed to initialize radar chart:', error);
        }
    }
    
    /**
     * Initialize SHAP analysis using existing RiskExplainability component
     */
    async initializeSHAPAnalysis() {
        try {
            if (this.riskData.shapData && this.explainability) {
                // Create SHAP force plot if container exists
                const shapContainer = document.getElementById('shapForcePlot');
                if (shapContainer && this.explainability.createSHAPForcePlot) {
                    this.explainability.createSHAPForcePlot('shapForcePlot', this.riskData.shapData);
                }
                
                // Create "Why this score?" panel
                const whyScoreContainer = document.getElementById('whyScorePanel');
                if (whyScoreContainer && this.explainability.createWhyScorePanel) {
                    this.explainability.createWhyScorePanel('whyScorePanel', {
                        overallScore: this.riskData.overallRiskScore,
                        contributions: this.riskData.shapData.feature_contributions,
                        recommendations: this.riskData.recommendations
                    });
                }
                
                console.log('🧠 SHAP analysis initialized');
            }
        } catch (error) {
            console.error('Failed to initialize SHAP analysis:', error);
        }
    }
    
    /**
     * Initialize risk category analysis
     */
    initializeRiskCategoryAnalysis() {
        const container = document.getElementById('riskCategoryAnalysis');
        if (!container) return;
        
        const categories = this.riskData.categories;
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        
        container.innerHTML = categoryOrder.map(category => {
            const data = categories[category];
            if (!data) return '';
            
            const trend = this.helpers.calculateTrendDirection(data.score, data.previousScore);
            
            return `
                <div class="risk-category-item p-3 border rounded-lg">
                    <div class="flex items-center justify-between">
                        <div class="flex items-center">
                            <div class="risk-icon risk-icon-${data.level} mr-3">
                                <i class="fas fa-${this.helpers.getRiskIcon(category)} text-xs"></i>
                            </div>
                            <div>
                                <h4 class="font-medium text-gray-900">${this.helpers.formatCategoryName(category)}</h4>
                                <p class="text-sm text-gray-600">Score: ${Math.round(data.score)}/100</p>
                            </div>
                        </div>
                        <div class="text-right">
                            <span class="risk-badge risk-${data.level} px-2 py-1 rounded text-xs font-bold">
                                ${data.level.toUpperCase()}
                            </span>
                            <div class="risk-trend trend-${trend.direction} mt-1">
                                <i class="fas fa-${trend.icon} mr-1"></i>
                                ${trend.label}
                            </div>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
        
        // Update risk summary
        this.updateRiskSummary();
    }
    
    /**
     * Update risk summary text
     */
    updateRiskSummary() {
        const summaryContainer = document.getElementById('riskSummary');
        if (!summaryContainer) return;
        
        const overallScore = this.riskData.overallRiskScore;
        const riskLevel = this.riskData.riskLevel;
        const alertCount = this.riskData.alerts.length;
        const recommendationCount = this.riskData.recommendations.length;
        
        let summary = `Overall risk score: ${Math.round(overallScore)}/100 (${riskLevel} risk). `;
        
        if (alertCount > 0) {
            summary += `${alertCount} active alert${alertCount > 1 ? 's' : ''} require attention. `;
        }
        
        if (recommendationCount > 0) {
            summary += `${recommendationCount} recommendation${recommendationCount > 1 ? 's' : ''} available for risk mitigation.`;
        }
        
        summaryContainer.textContent = summary;
    }
    
    /**
     * Initialize tooltips and interactive elements
     */
    initializeTooltips() {
        // Initialize tooltips for risk badges
        const riskIndicators = document.querySelectorAll('.risk-indicator');
        riskIndicators.forEach(indicator => {
            const tooltip = indicator.querySelector('.risk-tooltip');
            if (tooltip) {
                indicator.addEventListener('mouseenter', () => {
                    tooltip.style.display = 'block';
                });
                indicator.addEventListener('mouseleave', () => {
                    tooltip.style.display = 'none';
                });
            }
        });
        
        // Initialize heat map tooltips
        const heatmapCells = document.querySelectorAll('.heatmap-cell');
        heatmapCells.forEach(cell => {
            const title = cell.getAttribute('title');
            if (title) {
                cell.addEventListener('mouseenter', (e) => {
                    this.showTooltip(e.target, title);
                });
                cell.addEventListener('mouseleave', () => {
                    this.hideTooltip();
                });
            }
        });
    }
    
    /**
     * Show tooltip
     * @param {HTMLElement} element - Target element
     * @param {string} text - Tooltip text
     */
    showTooltip(element, text) {
        // Remove existing tooltip
        this.hideTooltip();
        
        // Create tooltip
        const tooltip = document.createElement('div');
        tooltip.className = 'risk-tooltip-popup';
        tooltip.textContent = text;
        tooltip.style.cssText = `
            position: absolute;
            background: #1f2937;
            color: white;
            padding: 8px 12px;
            border-radius: 6px;
            font-size: 12px;
            z-index: 1000;
            pointer-events: none;
            max-width: 200px;
        `;
        
        document.body.appendChild(tooltip);
        
        // Position tooltip
        const rect = element.getBoundingClientRect();
        tooltip.style.left = `${rect.left + rect.width / 2 - tooltip.offsetWidth / 2}px`;
        tooltip.style.top = `${rect.top - tooltip.offsetHeight - 8}px`;
        
        this.currentTooltip = tooltip;
    }
    
    /**
     * Hide tooltip
     */
    hideTooltip() {
        if (this.currentTooltip) {
            this.currentTooltip.remove();
            this.currentTooltip = null;
        }
    }
    
    /**
     * Prepare radar chart data
     * @returns {Object} Radar chart data
     */
    prepareRadarData() {
        const categories = this.riskData.categories;
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        
        return {
            labels: categoryOrder.map(cat => this.helpers.formatCategoryName(cat)),
            datasets: [{
                label: 'Current Risk Level',
                data: categoryOrder.map(cat => categories[cat]?.score || 0),
                backgroundColor: 'rgba(59, 130, 246, 0.2)',
                borderColor: 'rgba(59, 130, 246, 1)',
                borderWidth: 2,
                pointBackgroundColor: 'rgba(59, 130, 246, 1)',
                pointBorderColor: '#fff',
                pointHoverBackgroundColor: '#fff',
                pointHoverBorderColor: 'rgba(59, 130, 246, 1)'
            }, {
                label: 'Industry Average',
                data: [25, 35, 45, 30, 40, 20], // Mock industry averages
                backgroundColor: 'rgba(156, 163, 175, 0.2)',
                borderColor: 'rgba(156, 163, 175, 1)',
                borderWidth: 2,
                pointBackgroundColor: 'rgba(156, 163, 175, 1)',
                pointBorderColor: '#fff',
                pointHoverBackgroundColor: '#fff',
                pointHoverBorderColor: 'rgba(156, 163, 175, 1)'
            }]
        };
    }
    
    /**
     * Bind event handlers
     */
    bindEventHandlers() {
        // Bind alert action handlers
        this.bindAlertHandlers();
        
        // Bind recommendation action handlers
        this.bindRecommendationHandlers();
        
        // Bind refresh handler
        this.bindRefreshHandler();
        
        console.log('🔗 Event handlers bound');
    }
    
    /**
     * Bind alert action handlers
     */
    bindAlertHandlers() {
        // Acknowledge alert
        window.acknowledgeAlert = (alertId) => {
            console.log(`Acknowledging alert: ${alertId}`);
            // TODO: Implement alert acknowledgment
            this.showToast('Alert acknowledged', 'success');
        };
        
        // Investigate alert
        window.investigateAlert = (alertId) => {
            console.log(`Investigating alert: ${alertId}`);
            // TODO: Implement alert investigation
            this.showToast('Alert investigation started', 'info');
        };
    }
    
    /**
     * Bind recommendation action handlers
     */
    bindRecommendationHandlers() {
        // Dismiss recommendation
        window.dismissRecommendation = (recId) => {
            console.log(`Dismissing recommendation: ${recId}`);
            // TODO: Implement recommendation dismissal
            this.showToast('Recommendation dismissed', 'info');
        };
        
        // Implement recommendation
        window.implementRecommendation = (recId) => {
            console.log(`Implementing recommendation: ${recId}`);
            // TODO: Implement recommendation implementation
            this.showToast('Recommendation implementation started', 'success');
        };
    }
    
    /**
     * Bind refresh handler
     */
    bindRefreshHandler() {
        // Add refresh button if not exists
        const container = document.getElementById(this.containerId);
        if (container && !container.querySelector('.refresh-button')) {
            const refreshBtn = document.createElement('button');
            refreshBtn.className = 'refresh-button btn btn-outline btn-sm absolute top-4 right-4';
            refreshBtn.innerHTML = '<i class="fas fa-sync-alt mr-1"></i>Refresh';
            refreshBtn.onclick = () => this.refresh();
            container.style.position = 'relative';
            container.appendChild(refreshBtn);
        }
    }
    
    /**
     * Show loading state
     */
    showLoading() {
        const loading = document.getElementById('riskIndicatorsLoading');
        const content = document.getElementById('riskIndicatorsContent');
        const error = document.getElementById('riskIndicatorsError');
        
        if (loading) loading.classList.remove('hidden');
        if (content) content.classList.add('hidden');
        if (error) error.classList.add('hidden');
    }
    
    /**
     * Hide loading state
     */
    hideLoading() {
        const loading = document.getElementById('riskIndicatorsLoading');
        const content = document.getElementById('riskIndicatorsContent');
        
        if (loading) loading.classList.add('hidden');
        if (content) content.classList.remove('hidden');
    }
    
    /**
     * Show error state
     * @param {Error} error - Error object
     */
    showError(error) {
        const loading = document.getElementById('riskIndicatorsLoading');
        const content = document.getElementById('riskIndicatorsContent');
        const errorEl = document.getElementById('riskIndicatorsError');
        
        if (loading) loading.classList.add('hidden');
        if (content) content.classList.add('hidden');
        if (errorEl) {
            errorEl.classList.remove('hidden');
            const errorText = errorEl.querySelector('p');
            if (errorText) {
                errorText.textContent = error.message || 'An unexpected error occurred';
            }
        }
    }
    
    /**
     * Retry loading data
     */
    async retry() {
        console.log('🔄 Retrying risk indicators load');
        await this.loadAndRender();
    }
    
    /**
     * Refresh data and re-render
     */
    async refresh() {
        console.log('🔄 Refreshing risk indicators');
        if (this.dataService) {
        this.dataService.clearCache(this.merchantId);
        }
        await this.loadAndRender();
        this.showToast('Risk indicators refreshed', 'success');
    }
    
    /**
     * Show toast notification
     * @param {string} message - Toast message
     * @param {string} type - Toast type (success, error, info, warning)
     */
    showToast(message, type = 'info') {
        // Create toast element
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.innerHTML = `
            <div class="flex items-center">
                <i class="fas fa-${this.getToastIcon(type)} mr-2"></i>
                <span>${message}</span>
            </div>
        `;
        toast.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${this.getToastColor(type)};
            color: white;
            padding: 12px 16px;
            border-radius: 6px;
            z-index: 10000;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
            transform: translateX(100%);
            transition: transform 0.3s ease;
        `;
        
        document.body.appendChild(toast);
        
        // Animate in
        setTimeout(() => {
            toast.style.transform = 'translateX(0)';
        }, 100);
        
        // Remove after 3 seconds
        setTimeout(() => {
            toast.style.transform = 'translateX(100%)';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        }, 3000);
    }
    
    /**
     * Get toast icon for type
     * @param {string} type - Toast type
     * @returns {string} Icon class
     */
    getToastIcon(type) {
        const icons = {
            success: 'check-circle',
            error: 'exclamation-circle',
            info: 'info-circle',
            warning: 'exclamation-triangle'
        };
        return icons[type] || 'info-circle';
    }
    
    /**
     * Get toast color for type
     * @param {string} type - Toast type
     * @returns {string} Color
     */
    getToastColor(type) {
        const colors = {
            success: '#10b981',
            error: '#ef4444',
            info: '#3b82f6',
            warning: '#f59e0b'
        };
        return colors[type] || '#3b82f6';
    }
    
    /**
     * Destroy the component and clean up
     */
    destroy() {
        // Remove event handlers
        this.eventHandlers.forEach((handler, element) => {
            element.removeEventListener(handler.event, handler.function);
        });
        this.eventHandlers.clear();
        
        // Clear cache
        if (this.merchantId) {
            if (this.dataService) {
            this.dataService.clearCache(this.merchantId);
        }
        }
        
        // Hide tooltip
        this.hideTooltip();
        
        // Clear container
        const container = document.getElementById(this.containerId);
        if (container) {
            container.innerHTML = '';
        }
        
        this.isInitialized = false;
        console.log('🧹 Risk Indicators Tab destroyed');
    }
}

// Global instance for event handlers
let riskIndicatorsTab = null;

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MerchantRiskIndicatorsTab;
}
