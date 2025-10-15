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

        // Initialize WebSocket client (disabled until service is deployed)
        // this.components.websocket = new RiskWebSocketClient({
        //     reconnectInterval: 1000,
        //     maxReconnectAttempts: 5
        // });
        console.log('🔍 WebSocket client disabled (service not deployed yet)');

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
            
            // Subscribe to WebSocket updates (disabled until service is deployed)
            if (this.components.websocket) {
                console.log('🔍 WebSocket disabled (service not deployed yet)');
                // this.components.websocket.subscribe(this.currentMerchantId);
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
            // For now, use mock data since the risk assessment service isn't deployed yet
            console.log('🔍 Using mock risk assessment data (service not deployed yet)');
            return this.generateMockRiskData();
            
            // TODO: Uncomment when risk assessment service is deployed
            /*
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
            */
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
        console.log('🔍 Updating risk UI with data:', this.riskData);
        
        // Update overall risk score
        const overallScoreElement = document.getElementById('overallRiskScore');
        if (overallScoreElement && this.riskData.overallScore) {
            overallScoreElement.textContent = this.riskData.overallScore.toFixed(1);
            overallScoreElement.className = `risk-score-value ${this.getRiskLevel(this.riskData.overallScore)}`;
        }

        // Update risk categories
        if (this.riskData.categories) {
            const categoriesContainer = document.getElementById('riskCategories');
            if (categoriesContainer) {
                categoriesContainer.innerHTML = Object.entries(this.riskData.categories)
                    .map(([category, score]) => `
                        <div class="risk-category">
                            <span class="category-name">${category}</span>
                            <span class="category-score ${this.getRiskLevel(score)}">${score.toFixed(1)}</span>
                        </div>
                    `).join('');
            }
        }

        // Update risk trend
        const trendElement = document.getElementById('riskTrend');
        if (trendElement && this.riskData.trend) {
            const trend = this.riskData.trend;
            const icon = trend.direction === 'up' ? 'fa-arrow-up' : 'fa-arrow-down';
            const color = trend.direction === 'up' ? 'text-red-500' : 'text-green-500';
            trendElement.innerHTML = `
                <i class="fas ${icon} ${color}"></i>
                <span>${trend.change} from last month</span>
            `;
        } else if (trendElement) {
            // Default trend if no data
            trendElement.innerHTML = `
                <i class="fas fa-minus text-gray-500"></i>
                <span>No trend data available</span>
            `;
        }

        // Update risk gauge
        if (this.components.visualization) {
            const gauge = this.components.visualization.d3Visualizations.get('riskGauge');
            if (gauge) {
                gauge.update(this.riskData.overallScore);
            }
        }

        // Update category scores (legacy method)
        Object.entries(this.riskData.categories || {}).forEach(([category, score]) => {
            const element = document.querySelector(`[data-category="${category}"] .category-score`);
            if (element) {
                element.textContent = score.toFixed(1);
                element.className = `category-score ${this.getRiskLevel(score)}`;
            }
        });

        console.log('✅ Risk UI updated successfully');
    }

    /**
     * Initialize visualizations after UI is updated
     */
    initializeVisualizations() {
        console.log('🔍 Initializing visualizations...');
        
        try {
            // Initialize risk gauge
            this.initializeRiskGauge();
            
            // Initialize risk trend chart
            this.initializeRiskTrendChart();
            
            // Initialize risk factor chart
            this.initializeRiskFactorChart();
            
            // Initialize SHAP explanation
            this.initializeSHAPExplanation();
            
            // Initialize scenario analysis
            this.initializeScenarioAnalysis();
            
            // Initialize risk history chart
            this.initializeRiskHistoryChart();
            
            console.log('✅ Visualizations initialized successfully');
        } catch (error) {
            console.error('Error initializing visualizations:', error);
        }
    }

    /**
     * Initialize risk gauge
     */
    initializeRiskGauge() {
        const gaugeContainer = document.getElementById('riskGauge');
        if (!gaugeContainer) {
            console.log('❌ Risk gauge container not found');
            return;
        }

        console.log('🔍 Initializing risk gauge...');
        
        const ctx = gaugeContainer.getContext('2d');
        const centerX = gaugeContainer.width / 2;
        const centerY = gaugeContainer.height / 2;
        const radius = 80;
        
        // Clear canvas
        ctx.clearRect(0, 0, gaugeContainer.width, gaugeContainer.height);
        
        // Draw background arc
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, Math.PI, 2 * Math.PI);
        ctx.lineWidth = 20;
        ctx.strokeStyle = '#e2e8f0';
        ctx.stroke();
        
        // Draw risk level arcs
        const riskScore = this.riskData?.overallScore || 7.2;
        const angle = (riskScore / 10) * Math.PI;
        
        // Low risk (0-3): Green
        if (riskScore <= 3) {
            ctx.strokeStyle = '#38a169';
        }
        // Medium risk (3-7): Yellow/Orange
        else if (riskScore <= 7) {
            ctx.strokeStyle = '#d69e2e';
        }
        // High risk (7-10): Red
        else {
            ctx.strokeStyle = '#e53e3e';
        }
        
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, Math.PI, Math.PI + angle);
        ctx.lineWidth = 20;
        ctx.lineCap = 'round';
        ctx.stroke();
        
        // Draw tick marks
        ctx.strokeStyle = '#718096';
        ctx.lineWidth = 2;
        for (let i = 0; i <= 10; i += 2) {
            const tickAngle = Math.PI + (i / 10) * Math.PI;
            const x1 = centerX + (radius - 10) * Math.cos(tickAngle);
            const y1 = centerY + (radius - 10) * Math.sin(tickAngle);
            const x2 = centerX + (radius + 10) * Math.cos(tickAngle);
            const y2 = centerY + (radius + 10) * Math.sin(tickAngle);
            
            ctx.beginPath();
            ctx.moveTo(x1, y1);
            ctx.lineTo(x2, y2);
            ctx.stroke();
        }
        
        // Draw labels
        ctx.fillStyle = '#718096';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        
        for (let i = 0; i <= 10; i += 2) {
            const tickAngle = Math.PI + (i / 10) * Math.PI;
            const x = centerX + (radius + 25) * Math.cos(tickAngle);
            const y = centerY + (radius + 25) * Math.sin(tickAngle);
            ctx.fillText(i.toString(), x, y);
        }
        
        console.log('✅ Risk gauge initialized successfully');
    }

    /**
     * Initialize risk trend chart
     */
    initializeRiskTrendChart() {
        const chartContainer = document.getElementById('riskTrendChart');
        if (!chartContainer) {
            console.log('❌ Risk trend chart container not found');
            return;
        }

        console.log('🔍 Initializing risk trend chart...');
        console.log('🔍 Chart container dimensions:', chartContainer.offsetWidth, 'x', chartContainer.offsetHeight);
        console.log('🔍 Chart container style:', chartContainer.style.cssText);
        console.log('🔍 Chart container parent:', chartContainer.parentElement);
        
        // Destroy existing chart if it exists
        if (window.riskTrendChart && typeof window.riskTrendChart.destroy === 'function') {
            window.riskTrendChart.destroy();
        }
        
        // Create a simple line chart using Chart.js
        const ctx = chartContainer.getContext('2d');
        
        // Store chart reference globally
        try {
            window.riskTrendChart = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
                    datasets: [{
                        label: 'Risk Score',
                        data: [6.8, 7.1, 6.9, 7.3, 7.0, 7.2],
                        borderColor: 'rgb(75, 192, 192)',
                        backgroundColor: 'rgba(75, 192, 192, 0.2)',
                        tension: 0.1
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        title: {
                            display: true,
                            text: 'Risk Score Trend'
                        },
                        legend: {
                            display: false
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 10
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Error creating risk trend chart:', error);
            return;
        }
        
        console.log('✅ Risk trend chart initialized successfully');
    }

    /**
     * Initialize risk factor chart
     */
    initializeRiskFactorChart() {
        const chartContainer = document.getElementById('riskFactorChart');
        if (!chartContainer) {
            console.log('❌ Risk factor chart container not found');
            return;
        }

        console.log('🔍 Initializing risk factor chart...');
        console.log('🔍 Chart container dimensions:', chartContainer.offsetWidth, 'x', chartContainer.offsetHeight);
        
        // Destroy existing chart if it exists
        if (window.riskFactorChart && typeof window.riskFactorChart.destroy === 'function') {
            window.riskFactorChart.destroy();
        }
        
        // Create a radar chart using Chart.js
        const ctx = chartContainer.getContext('2d');
        // Store chart reference globally
        try {
            window.riskFactorChart = new Chart(ctx, {
                type: 'radar',
                data: {
                    labels: ['Financial', 'Operational', 'Regulatory', 'Reputational', 'Cybersecurity'],
                    datasets: [{
                        label: 'Current Risk',
                        data: [8.1, 6.5, 4.2, 7.8, 6.9],
                        borderColor: 'rgb(255, 99, 132)',
                        backgroundColor: 'rgba(255, 99, 132, 0.2)',
                        pointBackgroundColor: 'rgb(255, 99, 132)',
                        pointBorderColor: '#fff',
                        pointHoverBackgroundColor: '#fff',
                        pointHoverBorderColor: 'rgb(255, 99, 132)'
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        title: {
                            display: true,
                            text: 'Risk Factor Analysis'
                        },
                        legend: {
                            display: false
                        }
                    },
                    scales: {
                        r: {
                            beginAtZero: true,
                            max: 10
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Error creating risk factor chart:', error);
            return;
        }
        
        console.log('✅ Risk factor chart initialized successfully');
    }

    /**
     * Initialize SHAP explanation
     */
    initializeSHAPExplanation() {
        const shapContainer = document.getElementById('shapExplanation');
        if (!shapContainer) return;

        console.log('🔍 Initializing SHAP explanation...');

        shapContainer.innerHTML = `
            <div class="shap-explanation" style="margin-top: 15px;">
                <div class="shap-summary" style="background: #f8f9fa; padding: 15px; border-radius: 8px; margin-bottom: 20px;">
                    <h5 style="margin-bottom: 10px; color: #495057; font-weight: 600;">Risk Score Breakdown</h5>
                    <p style="color: #6c757d; font-size: 14px; margin: 0;">The overall risk score of 7.2 is influenced by the following key factors:</p>
                </div>
                
                <div class="shap-factors" style="display: grid; gap: 12px;">
                    <div class="shap-factor positive" style="display: flex; justify-content: space-between; align-items: center; padding: 12px; background: #fff5f5; border-left: 4px solid #e53e3e; border-radius: 4px;">
                        <div style="display: flex; align-items: center; gap: 10px;">
                            <div style="width: 8px; height: 8px; background: #e53e3e; border-radius: 50%;"></div>
                            <span class="factor-name" style="font-weight: 500; color: #2d3748;">High Transaction Volume</span>
                        </div>
                        <span class="factor-impact" style="background: #fed7d7; color: #c53030; padding: 4px 8px; border-radius: 4px; font-weight: 600; font-size: 14px;">+2.3</span>
                    </div>
                    
                    <div class="shap-factor negative" style="display: flex; justify-content: space-between; align-items: center; padding: 12px; background: #f0fff4; border-left: 4px solid #38a169; border-radius: 4px;">
                        <div style="display: flex; align-items: center; gap: 10px;">
                            <div style="width: 8px; height: 8px; background: #38a169; border-radius: 50%;"></div>
                            <span class="factor-name" style="font-weight: 500; color: #2d3748;">Strong Credit History</span>
                        </div>
                        <span class="factor-impact" style="background: #c6f6d5; color: #2f855a; padding: 4px 8px; border-radius: 4px; font-weight: 600; font-size: 14px;">-1.8</span>
                    </div>
                    
                    <div class="shap-factor positive" style="display: flex; justify-content: space-between; align-items: center; padding: 12px; background: #fff5f5; border-left: 4px solid #e53e3e; border-radius: 4px;">
                        <div style="display: flex; align-items: center; gap: 10px;">
                            <div style="width: 8px; height: 8px; background: #e53e3e; border-radius: 50%;"></div>
                            <span class="factor-name" style="font-weight: 500; color: #2d3748;">Recent Address Change</span>
                        </div>
                        <span class="factor-impact" style="background: #fed7d7; color: #c53030; padding: 4px 8px; border-radius: 4px; font-weight: 600; font-size: 14px;">+1.2</span>
                    </div>
                    
                    <div class="shap-factor positive" style="display: flex; justify-content: space-between; align-items: center; padding: 12px; background: #fff5f5; border-left: 4px solid #e53e3e; border-radius: 4px;">
                        <div style="display: flex; align-items: center; gap: 10px;">
                            <div style="width: 8px; height: 8px; background: #e53e3e; border-radius: 50%;"></div>
                            <span class="factor-name" style="font-weight: 500; color: #2d3748;">High Market Volatility</span>
                        </div>
                        <span class="factor-impact" style="background: #fed7d7; color: #c53030; padding: 4px 8px; border-radius: 4px; font-weight: 600; font-size: 14px;">+0.9</span>
                    </div>
                    
                    <div class="shap-factor negative" style="display: flex; justify-content: space-between; align-items: center; padding: 12px; background: #f0fff4; border-left: 4px solid #38a169; border-radius: 4px;">
                        <div style="display: flex; align-items: center; gap: 10px;">
                            <div style="width: 8px; height: 8px; background: #38a169; border-radius: 50%;"></div>
                            <span class="factor-name" style="font-weight: 500; color: #2d3748;">Stable Business Model</span>
                        </div>
                        <span class="factor-impact" style="background: #c6f6d5; color: #2f855a; padding: 4px 8px; border-radius: 4px; font-weight: 600; font-size: 14px;">-0.7</span>
                    </div>
                </div>
                
                <div class="shap-interactive" style="margin-top: 20px; padding: 15px; background: #f7fafc; border-radius: 8px;">
                    <h6 style="margin-bottom: 10px; color: #4a5568; font-weight: 600;">Interactive Force Plot</h6>
                    <div id="shapForcePlot" style="height: 200px; background: white; border: 1px solid #e2e8f0; border-radius: 4px; position: relative; overflow: hidden;">
                        <canvas id="shapForceCanvas" width="100%" height="200" style="width: 100%; height: 200px; cursor: pointer;"></canvas>
                    </div>
                </div>
            </div>
        `;
        
        // Initialize SHAP force plot after HTML is rendered
        setTimeout(() => {
            this.initializeSHAPForcePlot();
        }, 100);
        
        console.log('✅ SHAP explanation initialized successfully');
    }
    
    /**
     * Initialize SHAP force plot visualization
     */
    initializeSHAPForcePlot() {
        const canvas = document.getElementById('shapForceCanvas');
        if (!canvas) {
            console.log('❌ SHAP force plot canvas not found');
            return;
        }

        console.log('🔍 Initializing SHAP force plot...');
        
        const ctx = canvas.getContext('2d');
        const width = canvas.width = canvas.offsetWidth;
        const height = canvas.height = 200;
        
        // Clear canvas
        ctx.clearRect(0, 0, width, height);
        
        // SHAP values for visualization
        const shapValues = [
            { name: 'High Transaction Volume', value: 2.3, color: '#e53e3e' },
            { name: 'Strong Credit History', value: -1.8, color: '#38a169' },
            { name: 'Recent Address Change', value: 1.2, color: '#e53e3e' },
            { name: 'High Market Volatility', value: 0.9, color: '#e53e3e' },
            { name: 'Stable Business Model', value: -0.7, color: '#38a169' }
        ];
        
        // Calculate positions and draw force plot
        const centerX = width / 2;
        const centerY = height / 2;
        const baseScore = 5.0; // Base score before SHAP contributions
        
        let currentX = centerX - 100; // Start from left
        
        // Draw base score
        ctx.fillStyle = '#4a5568';
        ctx.font = '14px Arial';
        ctx.textAlign = 'center';
        ctx.fillText('Base Score: 5.0', currentX, centerY - 20);
        
        // Draw SHAP contributions
        shapValues.forEach((factor, index) => {
            const barWidth = Math.abs(factor.value) * 20; // Scale factor
            const barHeight = 30;
            const barX = currentX + 20;
            const barY = centerY - barHeight / 2;
            
            // Draw bar
            ctx.fillStyle = factor.color;
            ctx.fillRect(barX, barY, barWidth, barHeight);
            
            // Draw value
            ctx.fillStyle = 'white';
            ctx.font = '12px Arial';
            ctx.textAlign = 'center';
            ctx.fillText(factor.value.toFixed(1), barX + barWidth / 2, centerY + 4);
            
            // Draw factor name
            ctx.fillStyle = '#4a5568';
            ctx.font = '10px Arial';
            ctx.textAlign = 'center';
            ctx.fillText(factor.name, barX + barWidth / 2, centerY + 25);
            
            currentX += barWidth + 30;
        });
        
        // Draw final score
        const finalScore = baseScore + shapValues.reduce((sum, factor) => sum + factor.value, 0);
        ctx.fillStyle = '#2d3748';
        ctx.font = '16px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(`Final Score: ${finalScore.toFixed(1)}`, currentX + 50, centerY - 20);
        
        // Add hover effects
        canvas.addEventListener('mousemove', (e) => {
            const rect = canvas.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;
            
            // Simple hover detection
            if (x > 50 && x < width - 50 && y > centerY - 20 && y < centerY + 20) {
                canvas.style.cursor = 'pointer';
            } else {
                canvas.style.cursor = 'default';
            }
        });
        
        console.log('✅ SHAP force plot initialized successfully');
    }

    /**
     * Initialize scenario analysis
     */
    initializeScenarioAnalysis() {
        const scenarioContainer = document.getElementById('scenarioAnalysis');
        if (!scenarioContainer) return;

        console.log('🔍 Initializing scenario analysis...');

        scenarioContainer.innerHTML = `
            <div class="scenario-analysis" style="margin-top: 15px;">
                <div class="scenario-description" style="background: #f8f9fa; padding: 15px; border-radius: 8px; margin-bottom: 20px;">
                    <h5 style="margin-bottom: 10px; color: #495057; font-weight: 600;">What-If Analysis</h5>
                    <p style="color: #6c757d; font-size: 14px; margin: 0;">Adjust the parameters below to see how different scenarios would affect the risk score:</p>
                </div>
                
                <div class="scenario-controls" style="display: grid; gap: 20px; margin-bottom: 25px;">
                    <div class="scenario-parameter" style="background: white; padding: 15px; border-radius: 8px; border: 1px solid #e2e8f0;">
                        <label style="display: block; margin-bottom: 8px; font-weight: 500; color: #2d3748;">Transaction Volume</label>
                        <div style="display: flex; align-items: center; gap: 15px;">
                            <input type="range" min="0" max="100" value="50" class="scenario-slider" id="transactionVolume" 
                                   style="flex: 1; height: 6px; background: #e2e8f0; border-radius: 3px; outline: none; cursor: pointer;">
                            <span class="scenario-value" style="min-width: 50px; text-align: center; font-weight: 600; color: #4a5568;">50%</span>
                        </div>
                        <div style="display: flex; justify-content: space-between; font-size: 12px; color: #718096; margin-top: 5px;">
                            <span>Low</span>
                            <span>High</span>
                        </div>
                    </div>
                    
                    <div class="scenario-parameter" style="background: white; padding: 15px; border-radius: 8px; border: 1px solid #e2e8f0;">
                        <label style="display: block; margin-bottom: 8px; font-weight: 500; color: #2d3748;">Credit Score</label>
                        <div style="display: flex; align-items: center; gap: 15px;">
                            <input type="range" min="300" max="850" value="650" class="scenario-slider" id="creditScore"
                                   style="flex: 1; height: 6px; background: #e2e8f0; border-radius: 3px; outline: none; cursor: pointer;">
                            <span class="scenario-value" style="min-width: 50px; text-align: center; font-weight: 600; color: #4a5568;">650</span>
                        </div>
                        <div style="display: flex; justify-content: space-between; font-size: 12px; color: #718096; margin-top: 5px;">
                            <span>300</span>
                            <span>850</span>
                        </div>
                    </div>
                    
                    <div class="scenario-parameter" style="background: white; padding: 15px; border-radius: 8px; border: 1px solid #e2e8f0;">
                        <label style="display: block; margin-bottom: 8px; font-weight: 500; color: #2d3748;">Market Conditions</label>
                        <div style="display: flex; align-items: center; gap: 15px;">
                            <input type="range" min="0" max="100" value="30" class="scenario-slider" id="marketConditions"
                                   style="flex: 1; height: 6px; background: #e2e8f0; border-radius: 3px; outline: none; cursor: pointer;">
                            <span class="scenario-value" style="min-width: 50px; text-align: center; font-weight: 600; color: #4a5568;">30%</span>
                        </div>
                        <div style="display: flex; justify-content: space-between; font-size: 12px; color: #718096; margin-top: 5px;">
                            <span>Stable</span>
                            <span>Volatile</span>
                        </div>
                    </div>
                </div>
                
                <div class="scenario-results" style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px;">
                    <div class="scenario-result" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 8px; text-align: center;">
                        <div style="font-size: 14px; opacity: 0.9; margin-bottom: 5px;">Predicted Risk Score</div>
                        <div class="scenario-score" style="font-size: 32px; font-weight: 700; margin-bottom: 5px;">6.8</div>
                        <div style="font-size: 12px; opacity: 0.8;">Current Scenario</div>
                    </div>
                    
                    <div class="scenario-comparison" style="background: #f7fafc; padding: 20px; border-radius: 8px; border: 1px solid #e2e8f0;">
                        <h6 style="margin-bottom: 15px; color: #4a5568; font-weight: 600;">Impact Analysis</h6>
                        <div style="display: grid; gap: 8px;">
                            <div style="display: flex; justify-content: space-between; font-size: 14px;">
                                <span style="color: #718096;">vs. Baseline:</span>
                                <span style="color: #38a169; font-weight: 600;">-0.4</span>
                            </div>
                            <div style="display: flex; justify-content: space-between; font-size: 14px;">
                                <span style="color: #718096;">Confidence:</span>
                                <span style="color: #4a5568; font-weight: 600;">87%</span>
                            </div>
                            <div style="display: flex; justify-content: space-between; font-size: 14px;">
                                <span style="color: #718096;">Risk Level:</span>
                                <span style="color: #d69e2e; font-weight: 600;">Medium</span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="scenario-actions" style="display: flex; gap: 10px; flex-wrap: wrap;">
                    <button class="btn btn-primary" style="background: #4299e1; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-size: 14px;">
                        <i class="fas fa-play"></i> Run Monte Carlo
                    </button>
                    <button class="btn btn-secondary" style="background: #718096; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-size: 14px;">
                        <i class="fas fa-chart-bar"></i> Stress Test
                    </button>
                    <button class="btn btn-outline" style="background: transparent; color: #4a5568; border: 1px solid #e2e8f0; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-size: 14px;">
                        <i class="fas fa-undo"></i> Reset
                    </button>
                </div>
            </div>
        `;
        
        // Add event listeners for sliders
        this.addScenarioEventListeners();
        
        console.log('✅ Scenario analysis initialized successfully');
    }
    
    /**
     * Add event listeners for scenario sliders
     */
    addScenarioEventListeners() {
        const transactionSlider = document.getElementById('transactionVolume');
        const creditSlider = document.getElementById('creditScore');
        const marketSlider = document.getElementById('marketConditions');
        
        if (transactionSlider) {
            transactionSlider.addEventListener('input', (e) => {
                const value = e.target.value;
                const valueSpan = e.target.nextElementSibling;
                if (valueSpan) valueSpan.textContent = value + '%';
                this.updateScenarioScore();
            });
        }
        
        if (creditSlider) {
            creditSlider.addEventListener('input', (e) => {
                const value = e.target.value;
                const valueSpan = e.target.nextElementSibling;
                if (valueSpan) valueSpan.textContent = value;
                this.updateScenarioScore();
            });
        }
        
        if (marketSlider) {
            marketSlider.addEventListener('input', (e) => {
                const value = e.target.value;
                const valueSpan = e.target.nextElementSibling;
                if (valueSpan) valueSpan.textContent = value + '%';
                this.updateScenarioScore();
            });
        }
    }
    
    /**
     * Update scenario score based on slider values
     */
    updateScenarioScore() {
        const transactionValue = document.getElementById('transactionVolume')?.value || 50;
        const creditValue = document.getElementById('creditScore')?.value || 650;
        const marketValue = document.getElementById('marketConditions')?.value || 30;
        
        // Simple calculation for demo purposes
        const baseScore = 7.2;
        const transactionImpact = (transactionValue - 50) * 0.02;
        const creditImpact = (650 - creditValue) * 0.01;
        const marketImpact = (marketValue - 30) * 0.015;
        
        const newScore = Math.max(0, Math.min(10, baseScore + transactionImpact + creditImpact + marketImpact));
        
        const scoreElement = document.querySelector('.scenario-score');
        if (scoreElement) {
            scoreElement.textContent = newScore.toFixed(1);
        }
    }

    /**
     * Initialize risk history chart
     */
    initializeRiskHistoryChart() {
        const historyContainer = document.getElementById('riskHistoryChart');
        if (!historyContainer) return;

        console.log('🔍 Initializing risk history chart...');

        historyContainer.innerHTML = `
            <div class="risk-history-chart" style="margin-top: 15px;">
                <div class="history-summary" style="background: #f8f9fa; padding: 15px; border-radius: 8px; margin-bottom: 20px;">
                    <h5 style="margin-bottom: 10px; color: #495057; font-weight: 600;">Risk Evolution (12 months)</h5>
                    <p style="color: #6c757d; font-size: 14px; margin: 0;">Track how risk scores have changed over time with key events and trends:</p>
                </div>
                
                <div class="history-timeline" style="position: relative; height: 300px; background: white; border: 1px solid #e2e8f0; border-radius: 8px; padding: 20px; margin-bottom: 20px;">
                    <div class="timeline-line" style="position: absolute; top: 50%; left: 5%; right: 5%; height: 2px; background: #e2e8f0; transform: translateY(-50%);"></div>
                    
                    <div class="history-point" style="position: absolute; top: 50%; left: 8%; transform: translateY(-50%);">
                        <div class="history-dot" style="width: 12px; height: 12px; background: #4299e1; border-radius: 50%; border: 3px solid white; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"></div>
                        <div class="history-label" style="position: absolute; top: -40px; left: 50%; transform: translateX(-50%); background: #4299e1; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; white-space: nowrap;">
                            Jan: 6.8
                        </div>
                        <div class="history-event" style="position: absolute; bottom: -30px; left: 50%; transform: translateX(-50%); font-size: 11px; color: #718096; text-align: center; width: 80px;">
                            Business Launch
                        </div>
                    </div>
                    
                    <div class="history-point" style="position: absolute; top: 50%; left: 28%; transform: translateY(-50%);">
                        <div class="history-dot" style="width: 12px; height: 12px; background: #e53e3e; border-radius: 50%; border: 3px solid white; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"></div>
                        <div class="history-label" style="position: absolute; top: -40px; left: 50%; transform: translateX(-50%); background: #e53e3e; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; white-space: nowrap;">
                            Apr: 7.1
                        </div>
                        <div class="history-event" style="position: absolute; bottom: -30px; left: 50%; transform: translateX(-50%); font-size: 11px; color: #718096; text-align: center; width: 80px;">
                            Payment Spike
                        </div>
                    </div>
                    
                    <div class="history-point" style="position: absolute; top: 50%; left: 48%; transform: translateY(-50%);">
                        <div class="history-dot" style="width: 12px; height: 12px; background: #38a169; border-radius: 50%; border: 3px solid white; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"></div>
                        <div class="history-label" style="position: absolute; top: -40px; left: 50%; transform: translateX(-50%); background: #38a169; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; white-space: nowrap;">
                            Aug: 6.9
                        </div>
                        <div class="history-event" style="position: absolute; bottom: -30px; left: 50%; transform: translateX(-50%); font-size: 11px; color: #718096; text-align: center; width: 80px;">
                            Credit Improvement
                        </div>
                    </div>
                    
                    <div class="history-point" style="position: absolute; top: 50%; left: 68%; transform: translateY(-50%);">
                        <div class="history-dot" style="width: 12px; height: 12px; background: #d69e2e; border-radius: 50%; border: 3px solid white; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"></div>
                        <div class="history-label" style="position: absolute; top: -40px; left: 50%; transform: translateX(-50%); background: #d69e2e; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; white-space: nowrap;">
                            Oct: 7.0
                        </div>
                        <div class="history-event" style="position: absolute; bottom: -30px; left: 50%; transform: translateX(-50%); font-size: 11px; color: #718096; text-align: center; width: 80px;">
                            Market Volatility
                        </div>
                    </div>
                    
                    <div class="history-point current" style="position: absolute; top: 50%; left: 88%; transform: translateY(-50%);">
                        <div class="history-dot" style="width: 16px; height: 16px; background: #667eea; border-radius: 50%; border: 4px solid white; box-shadow: 0 4px 8px rgba(0,0,0,0.2);"></div>
                        <div class="history-label" style="position: absolute; top: -40px; left: 50%; transform: translateX(-50%); background: #667eea; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; white-space: nowrap; font-weight: 600;">
                            Dec: 7.2
                        </div>
                        <div class="history-event" style="position: absolute; bottom: -30px; left: 50%; transform: translateX(-50%); font-size: 11px; color: #667eea; text-align: center; width: 80px; font-weight: 600;">
                            Current
                        </div>
                    </div>
                </div>
                
                <div class="history-stats" style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; margin-bottom: 20px;">
                    <div class="stat-card" style="background: #f7fafc; padding: 15px; border-radius: 8px; text-align: center; border: 1px solid #e2e8f0;">
                        <div style="font-size: 24px; font-weight: 700; color: #4a5568; margin-bottom: 5px;">+0.4</div>
                        <div style="font-size: 12px; color: #718096;">12-Month Change</div>
                    </div>
                    <div class="stat-card" style="background: #f7fafc; padding: 15px; border-radius: 8px; text-align: center; border: 1px solid #e2e8f0;">
                        <div style="font-size: 24px; font-weight: 700; color: #4a5568; margin-bottom: 5px;">6.8</div>
                        <div style="font-size: 12px; color: #718096;">Lowest Score</div>
                    </div>
                    <div class="stat-card" style="background: #f7fafc; padding: 15px; border-radius: 8px; text-align: center; border: 1px solid #e2e8f0;">
                        <div style="font-size: 24px; font-weight: 700; color: #4a5568; margin-bottom: 5px;">7.1</div>
                        <div style="font-size: 12px; color: #718096;">Highest Score</div>
                    </div>
                </div>
                
                <div class="history-actions" style="display: flex; gap: 10px; flex-wrap: wrap;">
                    <button class="btn btn-primary" style="background: #4299e1; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-size: 14px;">
                        <i class="fas fa-download"></i> Export History
                    </button>
                    <button class="btn btn-secondary" style="background: #718096; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-size: 14px;">
                        <i class="fas fa-chart-line"></i> View Trends
                    </button>
                    <button class="btn btn-outline" style="background: transparent; color: #4a5568; border: 1px solid #e2e8f0; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-size: 14px;">
                        <i class="fas fa-calendar"></i> Set Alerts
                    </button>
                </div>
            </div>
        `;
        
        console.log('✅ Risk history chart initialized successfully');
    }
    
    /**
     * Add event listeners for export buttons
     */
    addExportEventListeners() {
        const exportPDF = document.getElementById('exportPDF');
        const exportExcel = document.getElementById('exportExcel');
        const exportCSV = document.getElementById('exportCSV');
        
        if (exportPDF) {
            exportPDF.addEventListener('click', () => {
                console.log('🔍 Exporting PDF report...');
                this.exportPDF();
            });
        }
        
        if (exportExcel) {
            exportExcel.addEventListener('click', () => {
                console.log('🔍 Exporting Excel report...');
                this.exportExcel();
            });
        }
        
        if (exportCSV) {
            exportCSV.addEventListener('click', () => {
                console.log('🔍 Exporting CSV report...');
                this.exportCSV();
            });
        }
    }
    
    /**
     * Export PDF report
     */
    exportPDF() {
        // For demo purposes, show a success message
        alert('PDF export functionality would be implemented here. This would generate a comprehensive risk assessment report with charts and data.');
    }
    
    /**
     * Export Excel report
     */
    exportExcel() {
        // For demo purposes, show a success message
        alert('Excel export functionality would be implemented here. This would generate a formatted Excel file with risk data and charts.');
    }
    
    /**
     * Export CSV report
     */
    exportCSV() {
        // For demo purposes, show a success message
        alert('CSV export functionality would be implemented here. This would generate a CSV file with risk assessment data.');
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
                       <div class="risk-overview" style="display: grid; grid-template-columns: 1fr 2fr; gap: 20px; margin: 20px 0;">
                           <div class="risk-score-card" style="background: white; padding: 30px; border-radius: 15px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); text-align: center; position: relative;">
                               <div class="risk-gauge-container" style="position: relative; width: 200px; height: 200px; margin: 0 auto 20px;">
                                   <canvas id="riskGauge" width="200" height="200" style="width: 200px; height: 200px;"></canvas>
                                   <div class="gauge-center-text" style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); text-align: center;">
                                       <div class="risk-score-value" id="overallRiskScore" style="font-size: 36px; font-weight: 700; color: #2d3748; margin-bottom: 5px;">--</div>
                                       <div class="risk-score-label" style="font-size: 14px; color: #718096; font-weight: 500;">Overall Risk Score</div>
                                   </div>
                               </div>
                               <div class="risk-score-trend" id="riskTrend" style="display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 14px;">
                                   <i class="fas fa-minus text-gray-500"></i>
                                   <span>Loading...</span>
                               </div>
                           </div>
                           <div class="risk-categories" id="riskCategories" style="background: white; padding: 20px; border-radius: 15px; box-shadow: 0 4px 20px rgba(0,0,0,0.1);">
                               <h4 style="margin-bottom: 20px; color: #2d3748; font-size: 18px; font-weight: 600;">Risk Categories</h4>
                               <!-- Risk categories will be populated here -->
                           </div>
                       </div>

                       <!-- Risk Charts Section -->
                       <div class="risk-charts" style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin: 20px 0;">
                           <div class="chart-container" style="background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); overflow: hidden; position: relative;">
                               <h4 style="margin-bottom: 15px; color: #333; font-size: 16px; font-weight: 600;">Risk Trend (6 months)</h4>
                               <div style="position: relative; height: 200px; width: 100%;">
                                   <canvas id="riskTrendChart" style="max-height: 200px; max-width: 100%;"></canvas>
                               </div>
                           </div>
                           <div class="chart-container" style="background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); overflow: hidden; position: relative;">
                               <h4 style="margin-bottom: 15px; color: #333; font-size: 16px; font-weight: 600;">Risk Factor Analysis</h4>
                               <div style="position: relative; height: 200px; width: 100%;">
                                   <canvas id="riskFactorChart" style="max-height: 200px; max-width: 100%;"></canvas>
                               </div>
                           </div>
                       </div>

                       <!-- SHAP Explainability Section -->
                       <div class="risk-explainability" id="riskExplainability" style="background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin: 20px 0;">
                           <h4 style="margin-bottom: 15px; color: #333; font-size: 18px; font-weight: 600;">Why this score?</h4>
                           <div id="shapExplanation" class="shap-container">
                               <!-- SHAP explanation will be loaded here -->
                           </div>
                       </div>

                       <!-- Scenario Analysis Section -->
                       <div class="risk-scenarios" id="riskScenarios" style="background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin: 20px 0;">
                           <h4 style="margin-bottom: 15px; color: #333; font-size: 18px; font-weight: 600;">Scenario Analysis</h4>
                           <div id="scenarioAnalysis" class="scenario-container">
                               <!-- Scenario analysis will be loaded here -->
                           </div>
                       </div>

                       <!-- Risk History Section -->
                       <div class="risk-history" id="riskHistory" style="background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin: 20px 0;">
                           <h4 style="margin-bottom: 15px; color: #333; font-size: 18px; font-weight: 600;">Risk History</h4>
                           <div id="riskHistoryChart" class="history-container">
                               <!-- Risk history chart will be loaded here -->
                           </div>
                       </div>

                       <!-- Export Section -->
                       <div class="risk-export" id="riskExport" style="background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin: 20px 0;">
                           <h4 style="margin-bottom: 15px; color: #333; font-size: 18px; font-weight: 600;">Export Reports</h4>
                           <div class="export-buttons" style="display: flex; gap: 10px; flex-wrap: wrap;">
                               <button class="btn btn-primary" id="exportPDF" style="background: #dc3545; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer; display: flex; align-items: center; gap: 8px;">
                                   <i class="fas fa-file-pdf"></i> Export PDF
                               </button>
                               <button class="btn btn-success" id="exportExcel" style="background: #28a745; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer; display: flex; align-items: center; gap: 8px;">
                                   <i class="fas fa-file-excel"></i> Export Excel
                               </button>
                               <button class="btn btn-info" id="exportCSV" style="background: #17a2b8; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer; display: flex; align-items: center; gap: 8px;">
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
            
            // Initialize visualizations after UI is updated with a small delay
            setTimeout(() => {
                this.initializeVisualizations();
                this.addExportEventListeners();
            }, 100);

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
        console.log('🔍 Getting merchant ID...');
        
        // Try to get from URL parameter
        const urlParams = new URLSearchParams(window.location.search);
        const merchantId = urlParams.get('merchantId');
        if (merchantId) {
            console.log('✅ Found merchant ID from URL:', merchantId);
            return merchantId;
        }

        // Try to get from merchant details instance
        console.log('🔍 Checking window.merchantDetailsInstance:', window.merchantDetailsInstance);
        if (window.merchantDetailsInstance?.merchantData?.id) {
            console.log('✅ Found merchant ID from merchantData.id:', window.merchantDetailsInstance.merchantData.id);
            return window.merchantDetailsInstance.merchantData.id;
        }

        // Try to get from merchant details instance business name as fallback
        if (window.merchantDetailsInstance?.merchantData?.businessName) {
            const businessNameId = window.merchantDetailsInstance.merchantData.businessName.toLowerCase().replace(/\s+/g, '-');
            console.log('✅ Using business name as merchant ID:', businessNameId);
            return businessNameId;
        }

        // Try to get from global variable
        if (window.currentMerchantId) {
            console.log('✅ Found merchant ID from global variable:', window.currentMerchantId);
            return window.currentMerchantId;
        }

        // Try to get from merchant detail page
        const merchantIdElement = document.querySelector('[data-merchant-id]');
        if (merchantIdElement) {
            const id = merchantIdElement.getAttribute('data-merchant-id');
            console.log('✅ Found merchant ID from data attribute:', id);
            return id;
        }

        // Default fallback
        console.log('⚠️ No merchant ID found, using default fallback');
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
