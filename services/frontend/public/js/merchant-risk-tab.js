/**
 * Merchant Risk Tab Integration
 * 
 * Integrates all risk assessment components into the merchant detail page:
 * - Risk visualization components
 * - SHAP explainability
 * - Scenario analysis
 * - Risk history tracking
 * - Export functionality
 * - WebSocket real-time updates
 */

class MerchantRiskTab {
    constructor() {
        this.components = {
            websocket: null,
            visualization: null,
            explainability: null,
            scenarios: null,
            history: null,
            export: null
        };

        this.currentMerchantId = null;
        this.riskData = null;
        this.isInitialized = false;

        this.init();
    }

    /**
     * Initialize the risk tab
     */
    async init() {
        try {
            // Wait for DOM to be ready
            if (document.readyState === 'loading') {
                document.addEventListener('DOMContentLoaded', () => this.initializeComponents());
            } else {
                this.initializeComponents();
            }
        } catch (error) {
            console.error('Error initializing risk tab:', error);
        }
    }

    /**
     * Initialize all components
     */
    async initializeComponents() {
        // Get merchant ID from URL or page context
        this.currentMerchantId = this.getCurrentMerchantId();
        
        if (!this.currentMerchantId) {
            console.warn('No merchant ID found, using mock data');
            this.currentMerchantId = 'mock-merchant-123';
        }

        // Initialize WebSocket client
        this.components.websocket = new RiskWebSocketClient({
            reconnectInterval: 1000,
            maxReconnectAttempts: 5
        });

        // Initialize visualization component
        this.components.visualization = new RiskVisualization({
            animationDuration: 1000,
            colorScheme: {
                low: '#27ae60',
                medium: '#f39c12',
                high: '#e74c3c',
                critical: '#8e44ad'
            }
        });

        // Initialize explainability component
        this.components.explainability = new RiskExplainability({
            animationDuration: 1000,
            colorScheme: {
                positive: '#27ae60',
                negative: '#e74c3c',
                neutral: '#95a5a6'
            }
        });

        // Initialize scenario analysis component
        this.components.scenarios = new RiskScenarioAnalysis({
            animationDuration: 1000,
            simulationRuns: 1000
        });

        // Initialize history tracking component
        this.components.history = new RiskHistoryTracking({
            animationDuration: 1000,
            defaultTimeRange: 90
        });

        // Initialize export component
        this.components.export = new RiskExport({
            defaultFormat: 'pdf',
            includeCharts: true,
            includeExplanations: true
        });

        // Set up event listeners
        this.setupEventListeners();

        // Load initial data
        await this.loadInitialData();

        // Create UI components
        this.createRiskTabUI();

        this.isInitialized = true;
        console.log('Risk tab initialized successfully');
    }

    /**
     * Get current merchant ID
     */
    getCurrentMerchantId() {
        // Try to get from URL parameters
        const urlParams = new URLSearchParams(window.location.search);
        const merchantId = urlParams.get('id') || urlParams.get('merchantId');
        
        if (merchantId) {
            return merchantId;
        }

        // Try to get from page context
        const merchantElement = document.querySelector('[data-merchant-id]');
        if (merchantElement) {
            return merchantElement.getAttribute('data-merchant-id');
        }

        // Try to get from merchant name element
        const merchantNameElement = document.getElementById('merchantName');
        if (merchantNameElement && merchantNameElement.textContent !== 'Loading...') {
            // Extract ID from merchant name or use a generated one
            return 'merchant-' + merchantNameElement.textContent.toLowerCase().replace(/\s+/g, '-');
        }

        return null;
    }

    /**
     * Set up event listeners
     */
    setupEventListeners() {
        // WebSocket events
        if (this.components.websocket) {
            this.components.websocket.on('riskUpdate', (data) => {
                this.handleRiskUpdate(data);
            });

            this.components.websocket.on('riskPrediction', (data) => {
                this.handleRiskPrediction(data);
            });

            this.components.websocket.on('riskAlert', (data) => {
                this.handleRiskAlert(data);
            });
        }

        // Tab switching events
        document.addEventListener('tabChanged', (event) => {
            if (event.detail.tabName === 'risk') {
                this.onRiskTabActivated();
            }
        });

        // Export events
        document.addEventListener('exportRiskReport', (event) => {
            this.handleExportRequest(event.detail);
        });
    }

    /**
     * Load initial data
     */
    async loadInitialData() {
        try {
            // Load risk assessment data
            this.riskData = await this.loadRiskAssessmentData();
            
            // Subscribe to WebSocket updates
            if (this.components.websocket) {
                this.components.websocket.subscribe(this.currentMerchantId);
            }

        } catch (error) {
            console.error('Error loading initial data:', error);
            // Use mock data for development
            this.riskData = this.generateMockRiskData();
        }
    }

    /**
     * Load risk assessment data
     */
    async loadRiskAssessmentData() {
        try {
            const endpoints = APIConfig.getEndpoints();
            const response = await fetch(endpoints.riskAssess, {
                method: 'POST',
                headers: APIConfig.getHeaders(),
                body: JSON.stringify({ 
                    merchantId: this.currentMerchantId,
                    includePredictions: true,
                    includeExplanations: true
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Error loading risk assessment data:', error);
            return this.generateMockRiskData();
        }
    }

    /**
     * Generate mock risk data for development
     */
    generateMockRiskData() {
        return {
            id: `risk_${this.currentMerchantId}_${Date.now()}`,
            merchantId: this.currentMerchantId,
            overallScore: 7.2,
            trend: 0.3,
            categories: {
                financial: 8.1,
                operational: 6.5,
                compliance: 4.2,
                market: 7.8,
                reputation: 6.9
            },
            predictions: {
                threeMonth: 7.5,
                sixMonth: 7.8,
                twelveMonth: 8.1
            },
            confidence: 0.85,
            lastUpdated: new Date().toISOString(),
            factorContributions: [
                { feature: 'Revenue Growth', contribution: -0.8, reason: 'Declining revenue trend' },
                { feature: 'Market Volatility', contribution: 1.2, reason: 'High market volatility' },
                { feature: 'Operational Efficiency', contribution: 0.5, reason: 'Moderate efficiency' },
                { feature: 'Compliance Score', contribution: -1.1, reason: 'Strong compliance record' },
                { feature: 'Customer Satisfaction', contribution: 0.3, reason: 'Average satisfaction' }
            ],
            shapValues: [
                { name: 'Revenue Growth', shapValue: -0.8, featureValue: -0.05, contribution: -0.8, description: 'Revenue declining at 5% annually' },
                { name: 'Market Volatility', shapValue: 1.2, featureValue: 0.3, contribution: 1.2, description: 'High market volatility detected' },
                { name: 'Operational Efficiency', shapValue: 0.5, featureValue: 0.6, contribution: 0.5, description: 'Below-average operational efficiency' },
                { name: 'Compliance Score', shapValue: -1.1, featureValue: 0.85, contribution: -1.1, description: 'Strong compliance performance' },
                { name: 'Customer Satisfaction', shapValue: 0.3, featureValue: 0.75, contribution: 0.3, description: 'Average customer satisfaction' }
            ]
        };
    }

    /**
     * Create risk tab UI
     */
    createRiskTabUI() {
        const riskContent = document.getElementById('riskAssessmentContent');
        if (!riskContent) return;

        riskContent.innerHTML = `
            <div class="risk-tab-container">
                <!-- Risk Overview Section -->
                <div class="risk-section">
                    <div class="section-header">
                        <h3><i class="fas fa-chart-line"></i> Risk Overview</h3>
                        <div class="section-actions">
                            <button class="btn btn-sm btn-outline" id="refreshRiskBtn">
                                <i class="fas fa-sync-alt"></i> Refresh
                            </button>
                            <button class="btn btn-sm btn-primary" id="exportRiskBtn">
                                <i class="fas fa-download"></i> Export
                            </button>
                        </div>
                    </div>
                    
                    <div class="risk-overview-grid">
                        <div class="risk-gauge-container">
                            <div id="riskGauge" style="height: 300px;"></div>
                        </div>
                        <div class="risk-categories">
                            <div class="category-item">
                                <span class="category-name">Financial Risk</span>
                                <span class="category-score high">8.1</span>
                            </div>
                            <div class="category-item">
                                <span class="category-name">Operational Risk</span>
                                <span class="category-score medium">6.5</span>
                            </div>
                            <div class="category-item">
                                <span class="category-name">Compliance Risk</span>
                                <span class="category-score low">4.2</span>
                            </div>
                            <div class="category-item">
                                <span class="category-name">Market Risk</span>
                                <span class="category-score high">7.8</span>
                            </div>
                            <div class="category-item">
                                <span class="category-name">Reputation Risk</span>
                                <span class="category-score medium">6.9</span>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Risk Trends Section -->
                <div class="risk-section">
                    <div class="section-header">
                        <h3><i class="fas fa-chart-area"></i> Risk Trends</h3>
                    </div>
                    <div id="riskTrendChart" style="height: 300px;"></div>
                </div>

                <!-- Risk Explanation Section -->
                <div class="risk-section">
                    <div class="section-header">
                        <h3><i class="fas fa-question-circle"></i> Risk Explanation</h3>
                    </div>
                    <div class="explanation-tabs">
                        <button class="explanation-tab active" data-tab="shap">SHAP Analysis</button>
                        <button class="explanation-tab" data-tab="importance">Feature Importance</button>
                        <button class="explanation-tab" data-tab="why">Why This Score?</button>
                    </div>
                    <div class="explanation-content">
                        <div id="shapForcePlot" style="height: 400px;"></div>
                        <div id="featureImportance" style="height: 300px; display: none;"></div>
                        <div id="whyScorePanel" style="display: none;"></div>
                    </div>
                </div>

                <!-- Scenario Analysis Section -->
                <div class="risk-section">
                    <div class="section-header">
                        <h3><i class="fas fa-cogs"></i> Scenario Analysis</h3>
                    </div>
                    <div id="scenarioBuilder"></div>
                </div>

                <!-- Risk History Section -->
                <div class="risk-section">
                    <div class="section-header">
                        <h3><i class="fas fa-history"></i> Risk History</h3>
                    </div>
                    <div id="riskHistoryTimeline" style="height: 400px;"></div>
                    <div id="riskTrendSummary"></div>
                </div>
            </div>
        `;

        // Bind UI events
        this.bindUIEvents();

        // Initialize visualizations
        this.initializeVisualizations();
    }

    /**
     * Bind UI events
     */
    bindUIEvents() {
        // Refresh button
        const refreshBtn = document.getElementById('refreshRiskBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.refreshRiskData();
            });
        }

        // Export button
        const exportBtn = document.getElementById('exportRiskBtn');
        if (exportBtn) {
            exportBtn.addEventListener('click', () => {
                this.exportRiskReport();
            });
        }

        // Explanation tabs
        const explanationTabs = document.querySelectorAll('.explanation-tab');
        explanationTabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.switchExplanationTab(e.target.getAttribute('data-tab'));
            });
        });
    }

    /**
     * Initialize visualizations
     */
    initializeVisualizations() {
        if (!this.riskData) return;

        // Create risk gauge
        if (this.components.visualization) {
            this.components.visualization.createRiskGauge('riskGauge', this.riskData.overallScore);
        }

        // Create risk trend chart
        this.createRiskTrendChart();

        // Create SHAP force plot
        if (this.components.explainability && this.riskData.shapValues) {
            this.components.explainability.createSHAPForcePlot('shapForcePlot', {
                features: this.riskData.shapValues
            });
        }

        // Create feature importance chart
        if (this.components.explainability && this.riskData.factorContributions) {
            this.components.explainability.createFeatureImportanceWaterfall('featureImportance', 
                this.riskData.factorContributions.map(f => ({
                    name: f.feature,
                    importance: f.contribution,
                    description: f.reason
                }))
            );
        }

        // Create why score panel
        if (this.components.explainability) {
            this.components.explainability.createWhyScorePanel('whyScorePanel', {
                overallScore: this.riskData.overallScore,
                baseScore: 5.0,
                contributions: this.riskData.factorContributions,
                summary: this.generateRiskSummary(),
                keyFactors: this.extractKeyFactors()
            });
        }

        // Create scenario builder
        if (this.components.scenarios) {
            this.components.scenarios.createScenarioBuilder('scenarioBuilder');
        }

        // Create risk history timeline
        if (this.components.history) {
            this.components.history.createRiskHistoryTimeline('riskHistoryTimeline', []);
            this.components.history.createRiskTrendSummary('riskTrendSummary', {
                currentScore: this.riskData.overallScore,
                change: this.riskData.trend,
                averageScore: 6.8,
                volatility: 1.2,
                trendDirection: this.riskData.trend > 0 ? 'up' : this.riskData.trend < 0 ? 'down' : 'stable',
                trendStrength: 'moderate',
                insights: this.generateRiskInsights()
            });
        }
    }

    /**
     * Create risk trend chart
     */
    createRiskTrendChart() {
        // Generate mock trend data
        const trendData = {
            labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
            historical: [6.8, 7.0, 6.9, 7.1, 7.0, 7.2],
            prediction: [7.2, 7.3, 7.4, 7.5, 7.6, 7.7],
            confidenceUpper: [7.4, 7.5, 7.6, 7.7, 7.8, 7.9],
            confidenceLower: [7.0, 7.1, 7.2, 7.3, 7.4, 7.5]
        };

        if (this.components.visualization) {
            this.components.visualization.createRiskTrendChart('riskTrendChart', trendData);
        }
    }

    /**
     * Generate risk summary
     */
    generateRiskSummary() {
        const score = this.riskData.overallScore;
        const trend = this.riskData.trend;
        
        let summary = `The risk score of ${score.toFixed(1)} indicates `;
        
        if (score <= 3) {
            summary += 'a low risk profile with strong financial stability and compliance.';
        } else if (score <= 6) {
            summary += 'a moderate risk profile with some areas requiring attention.';
        } else if (score <= 8) {
            summary += 'a high risk profile with significant concerns that need immediate attention.';
        } else {
            summary += 'a critical risk profile requiring immediate intervention and monitoring.';
        }

        if (Math.abs(trend) > 0.1) {
            summary += ` The risk has ${trend > 0 ? 'increased' : 'decreased'} by ${Math.abs(trend).toFixed(1)} points recently.`;
        }

        return summary;
    }

    /**
     * Extract key factors
     */
    extractKeyFactors() {
        return this.riskData.factorContributions
            .filter(f => Math.abs(f.contribution) > 0.5)
            .map(f => ({
                name: f.feature,
                description: f.reason,
                impact: f.contribution
            }))
            .sort((a, b) => Math.abs(b.impact) - Math.abs(a.impact))
            .slice(0, 3);
    }

    /**
     * Generate risk insights
     */
    generateRiskInsights() {
        const insights = [];
        
        if (this.riskData.categories.financial > 7) {
            insights.push('Financial risk is elevated due to revenue volatility and cash flow concerns.');
        }
        
        if (this.riskData.categories.compliance < 5) {
            insights.push('Strong compliance record provides risk mitigation benefits.');
        }
        
        if (this.riskData.trend > 0.2) {
            insights.push('Risk trend is increasing, requiring closer monitoring.');
        }
        
        if (this.riskData.confidence < 0.8) {
            insights.push('Prediction confidence is moderate, consider additional data sources.');
        }

        return insights.length > 0 ? insights : ['Risk profile is within normal parameters.'];
    }

    /**
     * Switch explanation tab
     */
    switchExplanationTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.explanation-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // Update content visibility
        document.getElementById('shapForcePlot').style.display = tabName === 'shap' ? 'block' : 'none';
        document.getElementById('featureImportance').style.display = tabName === 'importance' ? 'block' : 'none';
        document.getElementById('whyScorePanel').style.display = tabName === 'why' ? 'block' : 'none';
    }

    /**
     * Handle risk update from WebSocket
     */
    handleRiskUpdate(data) {
        if (data.merchantId === this.currentMerchantId) {
            this.riskData = { ...this.riskData, ...data.riskData };
            this.updateRiskUI();
        }
    }

    /**
     * Handle risk prediction from WebSocket
     */
    handleRiskPrediction(data) {
        if (data.merchantId === this.currentMerchantId) {
            this.riskData.predictions = data.predictions;
            this.riskData.confidence = data.confidence;
            this.updatePredictionUI();
        }
    }

    /**
     * Handle risk alert from WebSocket
     */
    handleRiskAlert(data) {
        if (data.merchantId === this.currentMerchantId) {
            this.showRiskAlert(data);
        }
    }

    /**
     * Update risk UI
     */
    updateRiskUI() {
        // Update risk gauge
        if (this.components.visualization) {
            const gauge = this.components.visualization.d3Visualizations.get('riskGauge');
            if (gauge) {
                gauge.update(this.riskData.overallScore);
            }
        }

        // Update category scores
        Object.entries(this.riskData.categories).forEach(([category, score]) => {
            const element = document.querySelector(`[data-category="${category}"] .category-score`);
            if (element) {
                element.textContent = score.toFixed(1);
                element.className = `category-score ${this.getRiskLevel(score)}`;
            }
        });
    }

    /**
     * Update prediction UI
     */
    updatePredictionUI() {
        // Update prediction displays
        const predictionElements = document.querySelectorAll('[data-prediction]');
        predictionElements.forEach(element => {
            const predictionType = element.getAttribute('data-prediction');
            if (this.riskData.predictions[predictionType]) {
                element.textContent = this.riskData.predictions[predictionType].toFixed(1);
            }
        });
    }

    /**
     * Show risk alert
     */
    showRiskAlert(alertData) {
        const alertContainer = document.getElementById('riskAlerts') || this.createAlertContainer();
        
        const alertElement = document.createElement('div');
        alertElement.className = `risk-alert alert-${alertData.severity}`;
        alertElement.innerHTML = `
            <div class="alert-header">
                <i class="fas fa-exclamation-triangle"></i>
                <span class="alert-type">${alertData.alertType}</span>
                <button class="alert-close" onclick="this.parentElement.parentElement.remove()">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="alert-message">${alertData.message}</div>
            <div class="alert-timestamp">${new Date(alertData.timestamp).toLocaleString()}</div>
        `;

        alertContainer.appendChild(alertElement);

        // Auto-remove after 10 seconds
        setTimeout(() => {
            if (alertElement.parentElement) {
                alertElement.remove();
            }
        }, 10000);
    }

    /**
     * Create alert container
     */
    createAlertContainer() {
        const container = document.createElement('div');
        container.id = 'riskAlerts';
        container.className = 'risk-alerts-container';
        container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 1000;
            max-width: 400px;
        `;
        document.body.appendChild(container);
        return container;
    }

    /**
     * Get risk level
     */
    getRiskLevel(score) {
        if (score <= 3) return 'low';
        if (score <= 6) return 'medium';
        if (score <= 8) return 'high';
        return 'critical';
    }

    /**
     * Refresh risk data
     */
    async refreshRiskData() {
        const refreshBtn = document.getElementById('refreshRiskBtn');
        if (refreshBtn) {
            refreshBtn.disabled = true;
            refreshBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Refreshing...';
        }

        try {
            await this.loadInitialData();
            this.updateRiskUI();
        } finally {
            if (refreshBtn) {
                refreshBtn.disabled = false;
                refreshBtn.innerHTML = '<i class="fas fa-sync-alt"></i> Refresh';
            }
        }
    }

    /**
     * Load risk assessment content into the specified container
     */
    async loadRiskAssessmentContent(container) {
        console.log('🔍 MerchantRiskTab.loadRiskAssessmentContent() called with container:', container);
        if (!container) {
            console.log('❌ No container provided');
            return;
        }

        try {
            // Get merchant ID from URL or global variable
            const merchantId = this.getMerchantId();
            this.currentMerchantId = merchantId;

            // Load the comprehensive risk assessment UI
            container.innerHTML = `
                <div class="risk-content-loaded">
                    <!-- Risk Overview Section -->
                    <div class="risk-overview">
                        <div class="risk-score-card">
                            <div class="risk-score-value" id="overallRiskScore">--</div>
                            <div class="risk-score-label">Overall Risk Score</div>
                            <div class="risk-score-trend" id="riskTrend">
                                <i class="fas fa-minus text-gray-500"></i>
                                <span>Loading...</span>
                            </div>
                        </div>
                        <div class="risk-categories" id="riskCategories">
                            <!-- Risk categories will be populated here -->
                        </div>
                    </div>

                    <!-- Risk Charts Section -->
                    <div class="risk-charts">
                        <div class="chart-container">
                            <h4>Risk Trend (6 months)</h4>
                            <div id="riskTrendChart" style="height: 200px;"></div>
                        </div>
                        <div class="chart-container">
                            <h4>Risk Factor Analysis</h4>
                            <div id="riskFactorChart" style="height: 200px;"></div>
                        </div>
                    </div>

                    <!-- SHAP Explainability Section -->
                    <div class="risk-explainability" id="riskExplainability">
                        <h4>Why this score?</h4>
                        <div id="shapExplanation" class="shap-container">
                            <!-- SHAP explanation will be loaded here -->
                        </div>
                    </div>

                    <!-- Scenario Analysis Section -->
                    <div class="risk-scenarios" id="riskScenarios">
                        <h4>Scenario Analysis</h4>
                        <div id="scenarioAnalysis" class="scenario-container">
                            <!-- Scenario analysis will be loaded here -->
                        </div>
                    </div>

                    <!-- Risk History Section -->
                    <div class="risk-history" id="riskHistory">
                        <h4>Risk History</h4>
                        <div id="riskHistoryChart" class="history-container">
                            <!-- Risk history chart will be loaded here -->
                        </div>
                    </div>

                    <!-- Export Section -->
                    <div class="risk-export" id="riskExport">
                        <h4>Export Reports</h4>
                        <div class="export-buttons">
                            <button class="btn btn-primary" id="exportPDF">
                                <i class="fas fa-file-pdf"></i> Export PDF
                            </button>
                            <button class="btn btn-success" id="exportExcel">
                                <i class="fas fa-file-excel"></i> Export Excel
                            </button>
                            <button class="btn btn-info" id="exportCSV">
                                <i class="fas fa-file-csv"></i> Export CSV
                            </button>
                        </div>
                    </div>
                </div>
            `;

            // Initialize components
            await this.initializeComponents();
            
            // Load initial data
            await this.loadInitialData();
            
            // Update UI with loaded data
            this.updateRiskUI();

        } catch (error) {
            console.error('Error loading risk assessment content:', error);
            container.innerHTML = `
                <div class="error-message">
                    <i class="fas fa-exclamation-triangle"></i>
                    <h4>Error Loading Risk Assessment</h4>
                    <p>Unable to load risk assessment data. Please try again later.</p>
                    <button class="btn btn-primary" onclick="location.reload()">
                        <i class="fas fa-refresh"></i> Retry
                    </button>
                </div>
            `;
        }
    }

    /**
     * Get merchant ID from URL or global variable
     */
    getMerchantId() {
        // Try to get from URL parameter
        const urlParams = new URLSearchParams(window.location.search);
        const merchantId = urlParams.get('merchantId');
        if (merchantId) return merchantId;

        // Try to get from global variable
        if (window.currentMerchantId) return window.currentMerchantId;

        // Try to get from merchant detail page
        const merchantIdElement = document.querySelector('[data-merchant-id]');
        if (merchantIdElement) return merchantIdElement.getAttribute('data-merchant-id');

        // Default fallback
        return 'demo-merchant-001';
    }

    /**
     * Export risk report
     */
    exportRiskReport() {
        if (this.components.export) {
            this.components.export.exportRiskReport({
                merchantId: this.currentMerchantId,
                format: 'pdf',
                template: 'risk-assessment',
                includeCharts: true,
                includeExplanations: true,
                includeScenarios: true
            });
        }
    }

    /**
     * Handle export request
     */
    handleExportRequest(options) {
        if (this.components.export) {
            this.components.export.exportRiskReport({
                merchantId: this.currentMerchantId,
                ...options
            });
        }
    }

    /**
     * On risk tab activated
     */
    onRiskTabActivated() {
        if (!this.isInitialized) {
            this.initializeComponents();
        }

        // Refresh data when tab is activated
        this.refreshRiskData();
    }

    /**
     * Destroy component
     */
    destroy() {
        Object.values(this.components).forEach(component => {
            if (component && component.destroy) {
                component.destroy();
            }
        });

        this.components = {};
        this.riskData = null;
        this.isInitialized = false;
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Only initialize if we're on the merchant detail page
    if (document.getElementById('riskAssessmentContent')) {
        window.merchantRiskTab = new MerchantRiskTab();
    }
});

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MerchantRiskTab;
}

// Make available globally
window.MerchantRiskTab = MerchantRiskTab;
console.log('✅ MerchantRiskTab class loaded and available globally');
