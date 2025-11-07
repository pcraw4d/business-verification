/**
 * Merchant Risk Indicators Tab Controller
 * 
 * Main controller that wires existing risk visualization components together
 * to create a fully functional Risk Indicators tab. Leverages:
 * - Shared Risk Data Service (unified data access)
 * - Shared Risk Visualizations (reusable charts)
 * - RiskExplainability component (SHAP plots, feature importance)
 * - RiskLevelIndicator component (badges, colors, icons)
 * - RiskIndicatorsUITemplate (HTML templates)
 */

class MerchantRiskIndicatorsTab {
    constructor(containerId = 'risk-indicators') {
        this.containerId = containerId;
        this.merchantId = null;
        this.riskData = null;
        this.isInitialized = false;
        
        // Shared components (use if available, fallback to existing)
        this.sharedRiskService = null;
        this.sharedRiskViz = null;
        this.benchmarks = null; // Will be loaded from Business Analytics
        
        // Reuse existing components (initialize later when D3.js is available)
        this.visualization = null;
        this.explainability = new RiskExplainability();
        this.levelIndicator = null; // Initialize later to avoid DOM manipulation
        
        // Use shared data service if available, otherwise fallback to existing
        try {
            // Try to import shared services (will work if modules are loaded)
            if (typeof getRiskDataService !== 'undefined') {
                this.sharedRiskService = getRiskDataService();
            }
            if (typeof getRiskVisualizations !== 'undefined') {
                this.sharedRiskViz = getRiskVisualizations();
            }
        } catch (e) {
            console.warn('Shared services not available, using fallback:', e);
        }
        
        // Fallback to existing data service if shared not available
        if (!this.sharedRiskService) {
            this.dataService = new RiskIndicatorsDataService();
        }
        
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
     * Load data and render the UI
     * Uses shared risk data service if available, otherwise falls back to existing service
     */
    async loadAndRender() {
        try {
            // Show loading state
            this.showLoading();
            
            // Try to use shared risk service if available
            if (this.sharedRiskService) {
                try {
                    // Load risk data with benchmarks
                    const riskData = await this.sharedRiskService.loadRiskData(this.merchantId, {
                        includeHistory: false,
                        includePredictions: false,
                        includeBenchmarks: true,
                        includeExplanations: true,
                        includeRecommendations: true
                    });
                    
                    // Store benchmarks for radar chart
                    if (riskData.benchmarks) {
                        this.benchmarks = riskData.benchmarks;
                    }
                    
                    // Convert shared service format to expected format
                    // The existing code expects riskData with categories, overallRiskScore, etc.
                    // We need to adapt the shared service response
                    if (riskData.current) {
                        // Use existing data service to merge and normalize
                        // This maintains compatibility with existing UI templates
                        const existingData = await this.dataService.loadAllRiskData(this.merchantId);
                        // Merge shared service data with existing structure
                        this.riskData = {
                            ...existingData,
                            // Override with shared service data where available
                            currentAssessment: riskData.current,
                            benchmarks: riskData.benchmarks,
                            predictions: riskData.predictions
                        };
                    } else {
                        // Fallback to existing service if shared service doesn't return expected format
                        this.riskData = await this.dataService.loadAllRiskData(this.merchantId);
                    }
                } catch (sharedError) {
                    console.warn('Shared risk service failed, using fallback:', sharedError);
                    this.riskData = await this.dataService.loadAllRiskData(this.merchantId);
                }
            } else {
                // Use existing data service
                this.riskData = await this.dataService.loadAllRiskData(this.merchantId);
            }
            
            // Render UI
            this.render();
            
            // Initialize visualizations using existing components
            await this.initializeVisualizations();
            
            // Initialize predictive forecast (unique to Risk Indicators tab)
            await this.initializePredictiveForecast();
            
            // Add contextual links to other tabs
            this.addContextualLinks();
            
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
                    
                    <!-- Predictive Risk Forecast Section (Unique to Risk Indicators) -->
                    ${this.uiTemplate.getPredictiveForecastHTML()}
                    
                    <!-- Recommendations Section -->
                    <div id="riskRecommendations" class="mb-6">
                        ${this.uiTemplate.getRecommendationsHTML(this.riskData.recommendations)}
                    </div>
                    
                    <!-- Website Risk Findings Section -->
                    <div id="websiteRiskFindings" class="mb-6">
                        ${this.uiTemplate.getWebsiteRiskFindingsHTML(this.riskData.websiteRisks)}
                    </div>
                    
                    <!-- Contextual Links Section -->
                    <div id="contextualLinks" class="mb-6">
                        <!-- Contextual links will be rendered here -->
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
     * Initialize radar chart using shared RiskVisualizations component
     */
    async initializeRadarChart() {
        try {
            const canvas = document.getElementById('riskRadarChart');
            if (!canvas) {
                console.warn('Radar chart canvas not found');
                return;
            }
            
            // Prepare category scores
            const categories = this.riskData.categories;
            const categoryScores = {};
            const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
            categoryOrder.forEach(cat => {
                categoryScores[cat] = categories[cat] || { score: 0 };
            });
            
            // Load industry benchmarks from Business Analytics (replaces mock data)
            let benchmarks = this.benchmarks;
            if (!benchmarks && this.sharedRiskService) {
                try {
                    benchmarks = await this.sharedRiskService.loadIndustryBenchmarks(this.merchantId);
                    this.benchmarks = benchmarks;
                } catch (error) {
                    console.warn('Failed to load industry benchmarks, using fallback:', error);
                    benchmarks = null;
                }
            }
            
            // Use shared risk visualizations if available
            if (this.sharedRiskViz) {
                this.sharedRiskViz.createRiskRadarChart('riskRadarChart', categoryScores, benchmarks, {
                    max: 100
                });
                console.log('📈 Radar chart initialized with shared component');
            } else if (this.visualization && this.visualization.createRiskRadarChart) {
                // Fallback to existing RiskVisualization component
                const radarData = this.prepareRadarData(benchmarks);
                this.visualization.createRiskRadarChart('riskRadarChart', radarData);
                console.log('📈 Radar chart initialized with existing component');
            } else {
                console.warn('No visualization component available for radar chart');
            }
        } catch (error) {
            console.error('Failed to initialize radar chart:', error);
        }
    }
    
    /**
     * Initialize SHAP analysis - Links to Risk Assessment tab instead of duplicating
     * The Risk Assessment tab has comprehensive SHAP analysis, so we link to it
     */
    async initializeSHAPAnalysis() {
        try {
            // Instead of duplicating SHAP analysis, add a link to Risk Assessment tab
            const shapContainer = document.getElementById('shapForcePlot');
            if (shapContainer) {
                // Add link to Risk Assessment tab for detailed SHAP analysis
                shapContainer.innerHTML = `
                    <div class="text-center p-8 bg-gray-50 rounded-lg border-2 border-dashed border-gray-300">
                        <i class="fas fa-chart-line text-4xl text-blue-600 mb-4"></i>
                        <h3 class="text-lg font-semibold text-gray-900 mb-2">Detailed Risk Explanation</h3>
                        <p class="text-gray-600 mb-4">View comprehensive SHAP analysis and feature importance in the Risk Assessment tab.</p>
                        <a href="#risk-assessment-tab" 
                           onclick="event.preventDefault(); 
                                    const tab = document.querySelector('[data-tab=\\'risk-assessment\\']') || document.querySelector('[href=\\'#risk-assessment-tab\\']');
                                    if (tab) tab.click();
                                    return false;"
                           class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                            <i class="fas fa-arrow-right mr-2"></i>
                            View Risk Assessment
                        </a>
                    </div>
                `;
            }
            
            // Also update "Why this score?" panel
            const whyScoreContainer = document.getElementById('whyScorePanel');
            if (whyScoreContainer) {
                whyScoreContainer.innerHTML = `
                    <div class="p-4 bg-blue-50 rounded-lg border border-blue-200">
                        <p class="text-sm text-blue-800">
                            <i class="fas fa-info-circle mr-2"></i>
                            For detailed risk score explanations, SHAP feature contributions, and scenario analysis, 
                            visit the <a href="#risk-assessment-tab" 
                                        onclick="event.preventDefault(); 
                                                 const tab = document.querySelector('[data-tab=\\'risk-assessment\\']') || document.querySelector('[href=\\'#risk-assessment-tab\\']');
                                                 if (tab) tab.click();
                                                 return false;"
                                        class="underline font-semibold">Risk Assessment tab</a>.
                        </p>
                    </div>
                `;
            }
            
            console.log('🔗 SHAP analysis linked to Risk Assessment tab');
        } catch (error) {
            console.error('Failed to initialize SHAP analysis link:', error);
        }
    }
    
    /**
     * Initialize predictive risk forecast (unique feature for Risk Indicators tab)
     */
    async initializePredictiveForecast() {
        try {
            // Initialize predictive forecast component
            if (typeof PredictiveRiskForecast === 'undefined') {
                console.warn('PredictiveRiskForecast component not available');
                return;
            }
            
            this.predictiveForecast = new PredictiveRiskForecast();
            
            // Load predictions if available from shared service
            let predictions = null;
            if (this.sharedRiskService && this.riskData.predictions) {
                predictions = this.riskData.predictions;
            } else if (this.sharedRiskService) {
                // Try to load predictions
                try {
                    const riskData = await this.sharedRiskService.loadRiskData(this.merchantId, {
                        includePredictions: true
                    });
                    predictions = riskData.predictions;
                } catch (error) {
                    console.warn('Failed to load predictions:', error);
                }
            }
            
            // Initialize forecast component
            await this.predictiveForecast.init(this.merchantId, predictions);
            
            console.log('📊 Predictive forecast initialized');
        } catch (error) {
            console.error('Failed to initialize predictive forecast:', error);
        }
    }
    
    /**
     * Add contextual links to other tabs
     */
    addContextualLinks() {
        try {
            // Use cross-tab navigation if available
            if (typeof getCrossTabNavigation !== 'undefined') {
                const crossTabNav = getCrossTabNavigation();
                const context = {
                    merchantId: this.merchantId,
                    riskLevel: this.riskData.riskLevel,
                    category: this.getHighestRiskCategory()
                };
                crossTabNav.renderContextualLinks('contextualLinks', context);
            } else {
                // Fallback: Add basic contextual links
                const container = document.getElementById('contextualLinks');
                if (container) {
                    container.innerHTML = `
                        <div class="bg-white rounded-lg shadow-lg p-6">
                            <h3 class="text-lg font-semibold text-gray-900 mb-4">Related Information</h3>
                            <div class="space-y-2">
                                <a href="#risk-assessment-tab" 
                                   onclick="event.preventDefault(); 
                                            const tab = document.querySelector('[data-tab=\\'risk-assessment\\']') || document.querySelector('[href=\\'#risk-assessment-tab\\']');
                                            if (tab) tab.click();
                                            return false;"
                                   class="flex items-center p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors">
                                    <i class="fas fa-chart-line mr-3 text-blue-600"></i>
                                    <div class="flex-1">
                                        <div class="font-medium text-gray-900">View Detailed Risk Analysis</div>
                                        <div class="text-sm text-gray-600">See comprehensive risk assessment with trend analysis and SHAP explanations</div>
                                    </div>
                                    <i class="fas fa-chevron-right text-gray-400"></i>
                                </a>
                                <a href="#business-analytics-tab" 
                                   onclick="event.preventDefault(); 
                                            const tab = document.querySelector('[data-tab=\\'business-analytics\\']') || document.querySelector('[href=\\'#business-analytics-tab\\']');
                                            if (tab) tab.click();
                                            return false;"
                                   class="flex items-center p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors">
                                    <i class="fas fa-chart-bar mr-3 text-blue-600"></i>
                                    <div class="flex-1">
                                        <div class="font-medium text-gray-900">See Industry Classification</div>
                                        <div class="text-sm text-gray-600">View MCC, NAICS, and SIC industry codes</div>
                                    </div>
                                    <i class="fas fa-chevron-right text-gray-400"></i>
                                </a>
                            </div>
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Failed to add contextual links:', error);
        }
    }
    
    /**
     * Get highest risk category
     */
    getHighestRiskCategory() {
        if (!this.riskData || !this.riskData.categories) {
            return null;
        }
        
        const categories = this.riskData.categories;
        let highestCategory = null;
        let highestScore = 0;
        
        Object.entries(categories).forEach(([category, data]) => {
            if (data.score > highestScore) {
                highestScore = data.score;
                highestCategory = category;
            }
        });
        
        return highestCategory;
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
     * Prepare radar chart data (fallback method for existing RiskVisualization component)
     * @param {Object} benchmarks - Industry benchmarks (optional)
     * @returns {Object} Radar chart data
     */
    prepareRadarData(benchmarks = null) {
        const categories = this.riskData.categories;
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        
        const datasets = [{
            label: 'Current Risk Level',
            data: categoryOrder.map(cat => categories[cat]?.score || 0),
            backgroundColor: 'rgba(59, 130, 246, 0.2)',
            borderColor: 'rgba(59, 130, 246, 1)',
            borderWidth: 2,
            pointBackgroundColor: 'rgba(59, 130, 246, 1)',
            pointBorderColor: '#fff',
            pointHoverBackgroundColor: '#fff',
            pointHoverBorderColor: 'rgba(59, 130, 246, 1)'
        }];
        
        // Use real industry benchmarks if available, otherwise use fallback
        let industryAverages = [25, 35, 45, 30, 40, 20]; // Fallback mock data
        if (benchmarks && benchmarks.averages) {
            industryAverages = categoryOrder.map(cat => benchmarks.averages[cat] || 0);
        }
        
        datasets.push({
            label: 'Industry Average',
            data: industryAverages,
            backgroundColor: 'rgba(156, 163, 175, 0.2)',
            borderColor: 'rgba(156, 163, 175, 1)',
            borderWidth: 2,
            pointBackgroundColor: 'rgba(156, 163, 175, 1)',
            pointBorderColor: '#fff',
            pointHoverBackgroundColor: '#fff',
            pointHoverBorderColor: 'rgba(156, 163, 175, 1)'
        });
        
        return {
            labels: categoryOrder.map(cat => this.helpers.formatCategoryName(cat)),
            datasets: datasets
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
        this.dataService.clearCache(this.merchantId);
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
            this.dataService.clearCache(this.merchantId);
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
