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

// MerchantRiskTab v3.0 - Canvas detection fixes
console.log('📦 Loading MerchantRiskTab v3.0 - Canvas detection fixes applied');

class MerchantRiskTab {
    constructor() {
        this.components = {
            websocket: null,
            visualization: null,
            explainability: null,
            scenarios: null,
            history: null,
            export: null,
            tooltip: null,
            scorePanel: null,
            dragDrop: null
        };

        this.currentMerchantId = null;
        this.riskData = null;
        this.isInitialized = false;

        // Don't auto-initialize - wait for explicit call
        // this.init();
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
        if (typeof RiskVisualization !== 'undefined') {
            this.components.visualization = new RiskVisualization({
                animationDuration: 1000,
                colorScheme: {
                    low: '#27ae60',
                    medium: '#f39c12',
                    high: '#e74c3c',
                    critical: '#8e44ad'
                }
            });
        } else {
            console.warn('⚠️ RiskVisualization not available');
        }

        // Initialize explainability component
        if (typeof RiskExplainability !== 'undefined') {
            this.components.explainability = new RiskExplainability({
                animationDuration: 1000,
                colorScheme: {
                    positive: '#27ae60',
                    negative: '#e74c3c',
                    neutral: '#95a5a6'
                }
            });
        } else {
            console.warn('⚠️ RiskExplainability not available');
        }

        // Initialize tooltip system
        if (typeof RiskTooltipSystem !== 'undefined') {
            this.components.tooltip = new RiskTooltipSystem();
            console.log('✅ RiskTooltipSystem initialized');
        } else if (typeof window !== 'undefined' && window.riskTooltipSystem) {
            this.components.tooltip = window.riskTooltipSystem;
        }

        // Initialize score panel
        if (typeof RiskScorePanel !== 'undefined') {
            const scorePanelContainer = document.getElementById('riskScorePanel');
            if (scorePanelContainer) {
                this.components.scorePanel = new RiskScorePanel('riskScorePanel', {
                    collapsed: false,
                    showBreakdown: true,
                    showFactors: true
                });
                this.components.scorePanel.init();
                console.log('✅ RiskScorePanel initialized');
            }
        }

        // Initialize drag and drop
        if (typeof RiskDragDrop !== 'undefined') {
            const riskConfigContainer = document.getElementById('riskConfigContainer');
            if (riskConfigContainer) {
                this.components.dragDrop = new RiskDragDrop('riskConfigContainer', {
                    onDragStart: (element, event) => {
                        console.log('Drag started:', element);
                    },
                    onDrag: (element, event, position) => {
                        // Update position during drag
                    },
                    onDragEnd: (element, event) => {
                        console.log('Drag ended:', element);
                        // Save new configuration
                    }
                });
                this.components.dragDrop.init();
                console.log('✅ RiskDragDrop initialized');
            }
        }

        // Initialize scenario analysis component
        if (typeof RiskScenarioAnalysis !== 'undefined') {
            this.components.scenarios = new RiskScenarioAnalysis({
                animationDuration: 1000,
                simulationRuns: 1000
            });
        } else {
            console.warn('⚠️ RiskScenarioAnalysis not available');
        }

        // Initialize history tracking component
        if (typeof RiskHistoryTracking !== 'undefined') {
            this.components.history = new RiskHistoryTracking({
                animationDuration: 1000,
                defaultTimeRange: 90
            });
        } else {
            console.warn('⚠️ RiskHistoryTracking not available');
        }

        // Initialize export component
        if (typeof RiskExport !== 'undefined') {
            this.components.export = new RiskExport({
                defaultFormat: 'pdf',
                includeCharts: true,
                includeExplanations: true
            });
        } else {
            console.warn('⚠️ RiskExport not available');
        }

        // Set up event listeners
        this.setupEventListeners();

        // Load initial data
        await this.loadInitialData();

        // Create UI components
        this.createRiskTabUI();
        
        // Initialize WebsiteRiskDisplay if available
        if (typeof WebsiteRiskDisplay !== 'undefined') {
            this.initializeWebsiteRiskDisplay();
        }

        this.isInitialized = true;
        console.log('Risk tab initialized successfully');
    }
    
    /**
     * Initialize Website Risk Display component
     */
    initializeWebsiteRiskDisplay() {
        const container = document.getElementById('websiteRiskDisplay');
        if (!container) return;
        
        // This will be populated when risk data is loaded
        // For now, just mark that it's available
        console.log('✅ WebsiteRiskDisplay available for integration');
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
     * Generate mock risk data for development/testing.
     * 
     * FALLBACK BEHAVIOR:
     *   - Used during development when API is unavailable
     *   - Used when risk assessment API call fails
     *   - Should be disabled in production builds
     * 
     * FALLBACK DATA - DO NOT USE AS PRIMARY DATA SOURCE
     * 
     * PRODUCTION SAFETY: In production, mock data is only generated if explicitly allowed.
     * 
     * @returns {Object} Mock risk data
     */
    generateMockRiskData() {
        // Production safety check
        if (typeof APIConfig !== 'undefined' && APIConfig.isProduction && APIConfig.isProduction()) {
            if (!APIConfig.allowMockData || !APIConfig.allowMockData()) {
                throw new Error('Mock data not allowed in production environment');
            }
        }
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
        // This method is deprecated - UI is now created by loadRiskAssessmentContent()
        // Only proceed if riskAssessmentContent exists (old code path)
        const riskContent = document.getElementById('riskAssessmentContent');
        if (!riskContent) {
            // No old container found, UI will be created by loadRiskAssessmentContent()
            return;
        }

        // Only create UI if loadRiskAssessmentContent hasn't been called yet
        if (document.getElementById('riskAssessmentContainer')) {
            // New container exists, don't create old UI
            return;
        }

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

        // Don't initialize visualizations here - they'll be initialized after loadRiskAssessmentContent()
        // This method is called from updateRiskUI() which happens before the HTML is fully set up
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
        // Try to find canvas element directly by ID first
        let canvas = document.getElementById('riskGauge');
        
        // If found element is not a canvas, look for canvas inside it or search by ID again
        if (!canvas) {
            console.error('❌ Risk gauge canvas not found by ID');
            return;
        }
        
        console.log('🔍 Found riskGauge element:', canvas, 'TagName:', canvas.tagName);
        console.log('🔍 Element HTML:', canvas.outerHTML.substring(0, 200));
        
        // If it's not a canvas, it's a container div - find the canvas inside
        if (canvas.tagName !== 'CANVAS') {
            console.log('⚠️ Element is not a canvas, searching for canvas inside...');
            const canvasElement = canvas.querySelector('canvas#riskGauge') || canvas.querySelector('canvas');
            if (canvasElement) {
                canvas = canvasElement;
                console.log('✅ Found canvas element inside container');
            } else {
                console.error('❌ Canvas element not found inside container');
                console.error('Container HTML:', canvas.outerHTML.substring(0, 300));
                console.error('Container children:', Array.from(canvas.children).map(c => `${c.tagName}#${c.id || 'no-id'}`));
                return;
            }
        } else {
            console.log('✅ Element is the canvas itself');
        }
        
        if (!canvas || canvas.tagName !== 'CANVAS') {
            console.error('❌ Failed to get valid canvas element');
            return;
        }

        console.log('🔍 Initializing advanced risk gauge...');
        console.log('🔍 Canvas dimensions:', canvas.width, 'x', canvas.height);
        
        const ctx = canvas.getContext('2d');
        const centerX = canvas.width / 2;
        const centerY = canvas.height / 2 + 10; // Optimized center for better fit
        const radius = 85; // Reduced radius to ensure numbers fit within canvas
        
        // Clear canvas with subtle gradient background
        const bgGradient = ctx.createRadialGradient(centerX, centerY, 0, centerX, centerY, radius + 30);
        bgGradient.addColorStop(0, 'rgba(255, 255, 255, 0.1)');
        bgGradient.addColorStop(1, 'rgba(255, 255, 255, 0.05)');
        ctx.fillStyle = bgGradient;
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        
        // Draw outer glow effect
        ctx.shadowColor = 'rgba(0, 0, 0, 0.1)';
        ctx.shadowBlur = 20;
        ctx.shadowOffsetX = 0;
        ctx.shadowOffsetY = 0;
        
        // Draw background arc with gradient - positioned to wrap around text
        const bgGradient2 = ctx.createLinearGradient(centerX - radius, centerY, centerX + radius, centerY);
        bgGradient2.addColorStop(0, '#f7fafc');
        bgGradient2.addColorStop(0.5, '#edf2f7');
        bgGradient2.addColorStop(1, '#e2e8f0');
        
        ctx.beginPath();
        // Start arc slightly higher to wrap around text better
        ctx.arc(centerX, centerY, radius, Math.PI * 0.8, Math.PI * 2.2);
        ctx.lineWidth = 24; // Slightly thicker for larger gauge
        ctx.strokeStyle = bgGradient2;
        ctx.lineCap = 'round';
        ctx.stroke();
        
        // Reset shadow
        ctx.shadowColor = 'transparent';
        ctx.shadowBlur = 0;
        
        // Draw risk level arc with advanced gradient
        const riskScore = this.riskData?.overallScore || 7.2;
        const angle = (riskScore / 10) * Math.PI * 1.4; // Adjusted for new arc range
        
        // Create dynamic gradient based on risk level
        let progressGradient;
        if (riskScore <= 3) {
            // Low risk: Green gradient
            progressGradient = ctx.createLinearGradient(centerX - radius, centerY, centerX + radius, centerY);
            progressGradient.addColorStop(0, '#48bb78');
            progressGradient.addColorStop(0.5, '#38a169');
            progressGradient.addColorStop(1, '#2f855a');
        } else if (riskScore <= 7) {
            // Medium risk: Yellow/Orange gradient
            progressGradient = ctx.createLinearGradient(centerX - radius, centerY, centerX + radius, centerY);
            progressGradient.addColorStop(0, '#f6ad55');
            progressGradient.addColorStop(0.5, '#ed8936');
            progressGradient.addColorStop(1, '#dd6b20');
        } else {
            // High risk: Red gradient
            progressGradient = ctx.createLinearGradient(centerX - radius, centerY, centerX + radius, centerY);
            progressGradient.addColorStop(0, '#fc8181');
            progressGradient.addColorStop(0.5, '#f56565');
            progressGradient.addColorStop(1, '#e53e3e');
        }
        
        // Draw progress arc with gradient and glow
        ctx.shadowColor = riskScore > 7 ? 'rgba(229, 62, 62, 0.3)' : riskScore > 3 ? 'rgba(237, 137, 54, 0.3)' : 'rgba(56, 161, 105, 0.3)';
        ctx.shadowBlur = 15;
        
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, Math.PI * 0.8, Math.PI * 0.8 + angle);
        ctx.lineWidth = 24; // Match background arc width
        ctx.strokeStyle = progressGradient;
        ctx.lineCap = 'round';
        ctx.stroke();
        
        // Reset shadow
        ctx.shadowColor = 'transparent';
        ctx.shadowBlur = 0;
        
        // Draw inner ring for depth
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius - 15, Math.PI * 0.8, Math.PI * 2.2);
        ctx.lineWidth = 2;
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.8)';
        ctx.stroke();
        
        // Draw advanced tick marks with different sizes
        ctx.strokeStyle = '#a0aec0';
        for (let i = 0; i <= 10; i += 1) {
            const tickAngle = Math.PI * 0.8 + (i / 10) * Math.PI * 1.4;
            const isMajorTick = i % 2 === 0;
            const tickLength = isMajorTick ? 15 : 8;
            const tickWidth = isMajorTick ? 3 : 1.5;
            
            const x1 = centerX + (radius - tickLength) * Math.cos(tickAngle);
            const y1 = centerY + (radius - tickLength) * Math.sin(tickAngle);
            const x2 = centerX + (radius + 5) * Math.cos(tickAngle);
            const y2 = centerY + (radius + 5) * Math.sin(tickAngle);
            
            ctx.lineWidth = tickWidth;
            ctx.beginPath();
            ctx.moveTo(x1, y1);
            ctx.lineTo(x2, y2);
            ctx.stroke();
        }
        
        // Draw labels with better styling
        ctx.fillStyle = '#4a5568';
        ctx.font = 'bold 14px Arial';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        
        for (let i = 0; i <= 10; i += 2) {
            const tickAngle = Math.PI * 0.8 + (i / 10) * Math.PI * 1.4;
            const x = centerX + (radius + 25) * Math.cos(tickAngle);
            const y = centerY + (radius + 25) * Math.sin(tickAngle);
            ctx.fillText(i.toString(), x, y);
        }
        
        // Draw center dot with glow
        ctx.shadowColor = 'rgba(0, 0, 0, 0.2)';
        ctx.shadowBlur = 10;
        ctx.fillStyle = '#ffffff';
        ctx.beginPath();
        ctx.arc(centerX, centerY, 6, 0, 2 * Math.PI);
        ctx.fill();
        
        // Reset shadow
        ctx.shadowColor = 'transparent';
        ctx.shadowBlur = 0;
        
        // Update risk level badge color
        const riskLevelBadge = document.getElementById('riskLevelBadge');
        if (riskLevelBadge) {
            if (riskScore <= 3) {
                riskLevelBadge.style.background = 'linear-gradient(135deg, #48bb78, #38a169)';
                riskLevelBadge.style.color = 'white';
                riskLevelBadge.textContent = 'Low Risk';
            } else if (riskScore <= 7) {
                riskLevelBadge.style.background = 'linear-gradient(135deg, #f6ad55, #ed8936)';
                riskLevelBadge.style.color = 'white';
                riskLevelBadge.textContent = 'Medium Risk';
            } else {
                riskLevelBadge.style.background = 'linear-gradient(135deg, #fc8181, #e53e3e)';
                riskLevelBadge.style.color = 'white';
                riskLevelBadge.textContent = 'High Risk';
            }
        }
        
        console.log('✅ Advanced risk gauge initialized successfully with score:', riskScore);
    }

    /**
     * Initialize risk trend chart
     */
    initializeRiskTrendChart() {
        // Try to find canvas element directly by ID first
        let canvas = document.getElementById('riskTrendChart');
        
        if (!canvas) {
            console.error('❌ Risk trend chart canvas not found by ID');
            return;
        }
        
        console.log('🔍 Initializing risk trend chart...');
        console.log('🔍 Found element:', canvas, 'TagName:', canvas.tagName);
        console.log('🔍 Element HTML:', canvas.outerHTML.substring(0, 200));
        
        // If it's not a canvas, it's a container div - find the canvas inside
        if (canvas.tagName !== 'CANVAS') {
            console.log('⚠️ Element is not a canvas, searching for canvas inside...');
            const canvasElement = canvas.querySelector('canvas#riskTrendChart') || canvas.querySelector('canvas');
            if (canvasElement) {
                canvas = canvasElement;
                console.log('✅ Found canvas element inside container');
            } else {
                console.error('❌ Canvas element not found inside container');
                console.error('Container HTML:', canvas.outerHTML.substring(0, 300));
                console.error('Container children:', Array.from(canvas.children).map(c => `${c.tagName}#${c.id || 'no-id'}`));
                // Create canvas element as fallback
                const container = canvas;
                canvas = document.createElement('canvas');
                canvas.id = 'riskTrendChartCanvas';
                canvas.style.width = '100%';
                canvas.style.height = '100%';
                canvas.style.maxHeight = '200px';
                canvas.style.maxWidth = '100%';
                container.appendChild(canvas);
                console.log('✅ Created new canvas element inside container');
            }
        } else {
            console.log('✅ Element is the canvas itself');
        }
        
        if (!canvas || canvas.tagName !== 'CANVAS') {
            console.error('❌ Failed to get valid canvas element');
            return;
        }
        
        // Destroy existing chart if it exists
        if (window.riskTrendChart && typeof window.riskTrendChart.destroy === 'function') {
            window.riskTrendChart.destroy();
        }
        
        // Create a simple line chart using Chart.js
        const ctx = canvas.getContext('2d');
        
        // Store chart reference globally
        try {
            const trendData = this.riskData?.trendData || {
                labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
                datasets: [{
                    label: 'Risk Score',
                    data: [6.8, 7.0, 6.9, 7.1, 7.0, 7.2],
                    borderColor: '#e53e3e',
                    backgroundColor: 'rgba(229, 62, 62, 0.1)',
                    tension: 0.4
                }]
            };
            
            window.riskTrendChart = new Chart(ctx, {
                type: 'line',
                data: trendData,
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            display: false
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: false,
                            min: 0,
                            max: 10
                        }
                    }
                }
            });
            
            console.log('✅ Risk trend chart initialized successfully');
        } catch (error) {
            console.error('❌ Error creating risk trend chart:', error);
        }
    }

    /**
     * Initialize risk factor chart
     */
    initializeRiskFactorChart() {
        // Try to find canvas element directly by ID first
        let canvas = document.getElementById('riskFactorChart');
        
        if (!canvas) {
            console.error('❌ Risk factor chart canvas not found by ID');
            return;
        }
        
        console.log('🔍 Initializing risk factor chart...');
        console.log('🔍 Found element:', canvas, 'TagName:', canvas.tagName);
        console.log('🔍 Element HTML:', canvas.outerHTML.substring(0, 200));
        
        // If it's not a canvas, it's a container div - find the canvas inside
        if (canvas.tagName !== 'CANVAS') {
            console.log('⚠️ Element is not a canvas, searching for canvas inside...');
            const canvasElement = canvas.querySelector('canvas#riskFactorChart') || canvas.querySelector('canvas');
            if (canvasElement) {
                canvas = canvasElement;
                console.log('✅ Found canvas element inside container');
            } else {
                console.error('❌ Canvas element not found inside container');
                console.error('Container HTML:', canvas.outerHTML.substring(0, 300));
                console.error('Container children:', Array.from(canvas.children).map(c => `${c.tagName}#${c.id || 'no-id'}`));
                // Create canvas element as fallback
                const container = canvas;
                canvas = document.createElement('canvas');
                canvas.id = 'riskFactorChartCanvas';
                canvas.style.width = '100%';
                canvas.style.height = '100%';
                canvas.style.maxHeight = '200px';
                canvas.style.maxWidth = '100%';
                container.appendChild(canvas);
                console.log('✅ Created new canvas element inside container');
            }
        } else {
            console.log('✅ Element is the canvas itself');
        }
        
        if (!canvas || canvas.tagName !== 'CANVAS') {
            console.error('❌ Failed to get valid canvas element');
            return;
        }
        
        // Destroy existing chart if it exists
        if (window.riskFactorChart && typeof window.riskFactorChart.destroy === 'function') {
            window.riskFactorChart.destroy();
        }
        
        const ctx = canvas.getContext('2d');
        
        try {
            const factorData = this.riskData?.riskFactors || [
                { name: 'Financial', score: 8.1 },
                { name: 'Operational', score: 6.5 },
                { name: 'Compliance', score: 4.2 },
                { name: 'Market', score: 7.8 },
                { name: 'Reputation', score: 6.9 }
            ];
            
            window.riskFactorChart = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: factorData.map(f => f.name),
                    datasets: [{
                        label: 'Risk Score',
                        data: factorData.map(f => f.score),
                        backgroundColor: factorData.map(f => 
                            f.score > 7 ? '#e53e3e' : f.score > 4 ? '#f6ad55' : '#48bb78'
                        )
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
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
            
            console.log('✅ Risk factor chart initialized successfully');
        } catch (error) {
            console.error('❌ Error creating risk factor chart:', error);
        }
    }

    /**
     * Initialize SHAP explanation
     */
    initializeSHAPExplanation() {
        const shapContainer = document.getElementById('shapExplanation');
        if (!shapContainer) {
            console.log('❌ SHAP explanation container not found');
            return;
        }

        console.log('🔍 Initializing SHAP explanation...');
        
        // Use explainability component if available
        if (this.components.explainability && this.riskData?.factorContributions) {
            try {
                this.components.explainability.createFeatureImportanceWaterfall('shapExplanation', 
                    this.riskData.factorContributions.map(f => ({
                        name: f.feature || f.name,
                        importance: f.contribution || f.score,
                        description: f.reason || f.description || ''
                    }))
                );
                console.log('✅ SHAP explanation initialized successfully');
            } catch (error) {
                console.error('❌ Error initializing SHAP explanation:', error);
                shapContainer.innerHTML = '<p class="text-gray-500">SHAP explanation will be available once data is loaded.</p>';
            }
        } else {
            shapContainer.innerHTML = '<p class="text-gray-500">SHAP explanation will be available once data is loaded.</p>';
        }
    }

    /**
     * Initialize scenario analysis
     */
    initializeScenarioAnalysis() {
        const scenarioContainer = document.getElementById('scenarioAnalysis');
        if (!scenarioContainer) {
            console.log('❌ Scenario analysis container not found');
            return;
        }

        console.log('🔍 Initializing scenario analysis...');
        
        if (this.components.scenarios) {
            try {
                this.components.scenarios.createScenarioBuilder('scenarioAnalysis');
                console.log('✅ Scenario analysis initialized successfully');
            } catch (error) {
                console.error('❌ Error initializing scenario analysis:', error);
                scenarioContainer.innerHTML = '<p class="text-gray-500">Scenario analysis will be available once data is loaded.</p>';
            }
        } else {
            scenarioContainer.innerHTML = '<p class="text-gray-500">Scenario analysis will be available once data is loaded.</p>';
        }
    }

    /**
     * Initialize risk history chart
     */
    initializeRiskHistoryChart() {
        const historyContainer = document.getElementById('riskHistoryChart');
        if (!historyContainer) {
            console.log('❌ Risk history chart container not found');
            return;
        }

        console.log('🔍 Initializing risk history chart...');
        
        if (this.components.history) {
            try {
                const historyData = this.riskData?.history || [];
                this.components.history.createRiskHistoryTimeline('riskHistoryChart', historyData);
                console.log('✅ Risk history chart initialized successfully');
            } catch (error) {
                console.error('❌ Error initializing risk history chart:', error);
                historyContainer.innerHTML = '<p class="text-gray-500">Risk history will be available once data is loaded.</p>';
            }
        } else {
            historyContainer.innerHTML = '<p class="text-gray-500">Risk history will be available once data is loaded.</p>';
        }
    }

    /**
     * Add export event listeners
     */
    addExportEventListeners() {
        const exportPDF = document.getElementById('exportPDF');
        const exportExcel = document.getElementById('exportExcel');
        const exportCSV = document.getElementById('exportCSV');

        if (exportPDF) {
            exportPDF.addEventListener('click', () => {
                if (this.components.export) {
                    this.components.export.exportRiskReport({
                        merchantId: this.currentMerchantId,
                        format: 'pdf'
                    });
                }
            });
        }

        if (exportExcel) {
            exportExcel.addEventListener('click', () => {
                if (this.components.export) {
                    this.components.export.exportRiskReport({
                        merchantId: this.currentMerchantId,
                        format: 'excel'
                    });
                }
            });
        }

        if (exportCSV) {
            exportCSV.addEventListener('click', () => {
                if (this.components.export) {
                    this.components.export.exportRiskReport({
                        merchantId: this.currentMerchantId,
                        format: 'csv'
                    });
                }
            });
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
                       <div class="risk-overview" style="display: grid; grid-template-columns: 1fr 2fr; gap: 20px; margin: 20px 0;">
                           <div class="risk-score-card" style="background: white; padding: 15px; border-radius: 15px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); text-align: center; position: relative;">
                               <div class="risk-gauge-container" style="position: relative; width: 250px; height: 250px; margin: 0 auto 10px;">
                                   <canvas id="riskGauge" width="250" height="250" style="width: 250px; height: 250px;"></canvas>
                                   <div class="gauge-center-text" style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); text-align: center; z-index: 10;">
                                       <div class="risk-score-value" id="overallRiskScore" style="font-size: 36px; font-weight: 800; color: #1a202c; margin-bottom: 6px; text-shadow: 0 2px 4px rgba(0,0,0,0.1); background: rgba(255,255,255,0.9); padding: 8px 12px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">7.2</div>
                                       <div class="risk-score-label" style="font-size: 14px; color: #4a5568; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 4px;">Overall Risk Score</div>
                                       <div class="risk-level-badge" id="riskLevelBadge" style="padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 1px; display: inline-block;">High Risk</div>
                                   </div>
                               </div>
                               <div class="risk-score-trend" id="riskTrend" style="display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 16px; color: #4a5568;">
                                   <i class="fas fa-chart-line" style="color: #e53e3e;"></i>
                                   <span>Risk Level: High</span>
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

            // Load initial data first (before UI initialization)
            await this.loadInitialData();
            
            // Initialize components (but skip createRiskTabUI since we're using loadRiskAssessmentContent)
            // We'll manually initialize only what we need
            this.isInitialized = true;
            
            // Update UI with loaded data
            this.updateRiskUI();
            
            // Wait a bit longer for DOM to fully render the HTML we just set
            // Use multiple requestAnimationFrame calls to ensure DOM is fully ready
            console.log('⏳ Waiting for DOM to render before initializing visualizations...');
            setTimeout(() => {
                requestAnimationFrame(() => {
                    requestAnimationFrame(() => {
                    // Verify canvas elements exist before initializing
                    const gauge = document.getElementById('riskGauge');
                    const trendChart = document.getElementById('riskTrendChart');
                    const factorChart = document.getElementById('riskFactorChart');
                    
                    console.log('🔍 Canvas elements check before initialization:');
                    console.log('  - riskGauge:', gauge, 'TagName:', gauge?.tagName);
                    if (gauge) {
                        console.log('  - riskGauge HTML:', gauge.outerHTML.substring(0, 150));
                        console.log('  - riskGauge parent:', gauge.parentElement);
                    }
                    console.log('  - riskTrendChart:', trendChart, 'TagName:', trendChart?.tagName);
                    if (trendChart) {
                        console.log('  - riskTrendChart HTML:', trendChart.outerHTML.substring(0, 150));
                        console.log('  - riskTrendChart parent:', trendChart.parentElement);
                    }
                    console.log('  - riskFactorChart:', factorChart, 'TagName:', factorChart?.tagName);
                    if (factorChart) {
                        console.log('  - riskFactorChart HTML:', factorChart.outerHTML.substring(0, 150));
                    }
                    
                    if (!gauge || !trendChart || !factorChart) {
                        console.error('❌ Some canvas elements are missing!');
                        console.error('  - riskGauge missing:', !gauge);
                        console.error('  - riskTrendChart missing:', !trendChart);
                        console.error('  - riskFactorChart missing:', !factorChart);
                        console.error('🔍 Container HTML:', container.innerHTML.substring(0, 500));
                        return;
                    }
                    
                    // Check if they're actually canvas elements
                    if (gauge.tagName !== 'CANVAS') {
                        console.error('❌ riskGauge is not a canvas! It is:', gauge.tagName);
                        console.error('  - riskGauge HTML:', gauge.outerHTML);
                        // Try to find canvas inside if it's a container
                        const canvasInside = gauge.querySelector('canvas');
                        if (canvasInside) {
                            console.log('  ✅ Found canvas inside gauge container');
                        } else {
                            console.error('  ❌ No canvas found inside gauge container');
                            return;
                        }
                    }
                    
                    if (trendChart.tagName !== 'CANVAS') {
                        console.error('❌ riskTrendChart is not a canvas! It is:', trendChart.tagName);
                        console.error('  - riskTrendChart HTML:', trendChart.outerHTML);
                        // Try to find canvas inside if it's a container
                        const canvasInside = trendChart.querySelector('canvas');
                        if (canvasInside) {
                            console.log('  ✅ Found canvas inside trendChart container');
                        } else {
                            console.error('  ❌ No canvas found inside trendChart container');
                            return;
                        }
                    }
                    
                    if (factorChart.tagName !== 'CANVAS') {
                        console.error('❌ riskFactorChart is not a canvas! It is:', factorChart.tagName);
                        console.error('  - riskFactorChart HTML:', factorChart.outerHTML);
                        // Try to find canvas inside if it's a container
                        const canvasInside = factorChart.querySelector('canvas');
                        if (canvasInside) {
                            console.log('  ✅ Found canvas inside factorChart container');
                        } else {
                            console.error('  ❌ No canvas found inside factorChart container');
                            return;
                        }
                    }
                    
                    console.log('✅ All canvas elements validated, initializing visualizations...');
                    this.initializeVisualizations();
                    this.addExportEventListeners();
                    });
                });
            }, 500); // Wait 500ms for DOM to be fully ready

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

// Make available globally for browser FIRST (before any DOMContentLoaded)
if (typeof window !== 'undefined') {
    window.MerchantRiskTab = MerchantRiskTab;
    console.log('✅ MerchantRiskTab class loaded and available globally');
}

// Initialize when DOM is ready (optional - main initialization happens from merchant-details.html)
document.addEventListener('DOMContentLoaded', () => {
    // Only initialize if we're on the merchant detail page with riskAssessmentContainer
    if (document.getElementById('riskAssessmentContainer') || document.getElementById('riskAssessmentContent')) {
        console.log('🔍 Auto-initializing MerchantRiskTab from DOMContentLoaded');
        if (!window.merchantRiskTab) {
            window.merchantRiskTab = new MerchantRiskTab();
        }
    }
});

// Node.js export
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MerchantRiskTab;
}
