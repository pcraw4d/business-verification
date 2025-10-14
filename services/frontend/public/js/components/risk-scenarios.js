/**
 * Risk Scenario Analysis Component
 * 
 * Provides interactive scenario analysis capabilities:
 * - Interactive scenario builder with parameter sliders
 * - Monte Carlo simulation results visualization
 * - Side-by-side scenario comparison
 * - What-if analysis for business changes
 * - Stress testing visualizations
 */

class RiskScenarioAnalysis {
    constructor(options = {}) {
        this.options = {
            animationDuration: 1000,
            simulationRuns: 1000,
            colorScheme: {
                baseline: '#3498db',
                optimistic: '#27ae60',
                pessimistic: '#e74c3c',
                stress: '#8e44ad'
            },
            ...options
        };

        this.scenarios = new Map();
        this.currentScenario = null;
        this.simulationResults = null;
        this.visualizations = new Map();

        this.init();
    }

    /**
     * Initialize the scenario analysis component
     */
    init() {
        this.setupEventListeners();
        this.initializeDefaultScenarios();
    }

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Listen for scenario parameter changes
        document.addEventListener('scenarioParameterChanged', (event) => {
            this.updateScenario(event.detail.scenarioId, event.detail.parameter, event.detail.value);
        });

        // Listen for scenario selection
        document.addEventListener('scenarioSelected', (event) => {
            this.selectScenario(event.detail.scenarioId);
        });
    }

    /**
     * Initialize default scenarios
     */
    initializeDefaultScenarios() {
        const defaultScenarios = {
            baseline: {
                id: 'baseline',
                name: 'Baseline Scenario',
                description: 'Current business conditions and risk factors',
                parameters: {
                    revenueGrowth: { value: 0.05, min: -0.2, max: 0.3, step: 0.01, label: 'Revenue Growth Rate' },
                    marketVolatility: { value: 0.15, min: 0.05, max: 0.5, step: 0.01, label: 'Market Volatility' },
                    operationalEfficiency: { value: 0.8, min: 0.5, max: 1.0, step: 0.01, label: 'Operational Efficiency' },
                    complianceScore: { value: 0.85, min: 0.3, max: 1.0, step: 0.01, label: 'Compliance Score' },
                    customerSatisfaction: { value: 0.75, min: 0.3, max: 1.0, step: 0.01, label: 'Customer Satisfaction' },
                    employeeRetention: { value: 0.7, min: 0.3, max: 1.0, step: 0.01, label: 'Employee Retention' }
                },
                color: this.options.colorScheme.baseline
            },
            optimistic: {
                id: 'optimistic',
                name: 'Optimistic Scenario',
                description: 'Favorable market conditions and improved performance',
                parameters: {
                    revenueGrowth: { value: 0.15, min: -0.2, max: 0.3, step: 0.01, label: 'Revenue Growth Rate' },
                    marketVolatility: { value: 0.08, min: 0.05, max: 0.5, step: 0.01, label: 'Market Volatility' },
                    operationalEfficiency: { value: 0.95, min: 0.5, max: 1.0, step: 0.01, label: 'Operational Efficiency' },
                    complianceScore: { value: 0.95, min: 0.3, max: 1.0, step: 0.01, label: 'Compliance Score' },
                    customerSatisfaction: { value: 0.9, min: 0.3, max: 1.0, step: 0.01, label: 'Customer Satisfaction' },
                    employeeRetention: { value: 0.85, min: 0.3, max: 1.0, step: 0.01, label: 'Employee Retention' }
                },
                color: this.options.colorScheme.optimistic
            },
            pessimistic: {
                id: 'pessimistic',
                name: 'Pessimistic Scenario',
                description: 'Challenging market conditions and operational difficulties',
                parameters: {
                    revenueGrowth: { value: -0.05, min: -0.2, max: 0.3, step: 0.01, label: 'Revenue Growth Rate' },
                    marketVolatility: { value: 0.3, min: 0.05, max: 0.5, step: 0.01, label: 'Market Volatility' },
                    operationalEfficiency: { value: 0.6, min: 0.5, max: 1.0, step: 0.01, label: 'Operational Efficiency' },
                    complianceScore: { value: 0.6, min: 0.3, max: 1.0, step: 0.01, label: 'Compliance Score' },
                    customerSatisfaction: { value: 0.5, min: 0.3, max: 1.0, step: 0.01, label: 'Customer Satisfaction' },
                    employeeRetention: { value: 0.4, min: 0.3, max: 1.0, step: 0.01, label: 'Employee Retention' }
                },
                color: this.options.colorScheme.pessimistic
            },
            stress: {
                id: 'stress',
                name: 'Stress Test Scenario',
                description: 'Extreme adverse conditions for stress testing',
                parameters: {
                    revenueGrowth: { value: -0.15, min: -0.2, max: 0.3, step: 0.01, label: 'Revenue Growth Rate' },
                    marketVolatility: { value: 0.45, min: 0.05, max: 0.5, step: 0.01, label: 'Market Volatility' },
                    operationalEfficiency: { value: 0.5, min: 0.5, max: 1.0, step: 0.01, label: 'Operational Efficiency' },
                    complianceScore: { value: 0.4, min: 0.3, max: 1.0, step: 0.01, label: 'Compliance Score' },
                    customerSatisfaction: { value: 0.3, min: 0.3, max: 1.0, step: 0.01, label: 'Customer Satisfaction' },
                    employeeRetention: { value: 0.3, min: 0.3, max: 1.0, step: 0.01, label: 'Employee Retention' }
                },
                color: this.options.colorScheme.stress
            }
        };

        Object.values(defaultScenarios).forEach(scenario => {
            this.scenarios.set(scenario.id, scenario);
        });

        this.currentScenario = 'baseline';
    }

    /**
     * Create scenario builder interface
     */
    createScenarioBuilder(containerId) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        container.innerHTML = `
            <div class="scenario-builder">
                <div class="scenario-header">
                    <h3>Scenario Analysis</h3>
                    <div class="scenario-controls">
                        <select id="scenarioSelector" class="scenario-selector">
                            ${Array.from(this.scenarios.values()).map(scenario => 
                                `<option value="${scenario.id}">${scenario.name}</option>`
                            ).join('')}
                        </select>
                        <button id="runSimulationBtn" class="btn btn-primary">
                            <i class="fas fa-play"></i>
                            Run Simulation
                        </button>
                        <button id="compareScenariosBtn" class="btn btn-outline">
                            <i class="fas fa-balance-scale"></i>
                            Compare Scenarios
                        </button>
                    </div>
                </div>
                
                <div class="scenario-content">
                    <div class="scenario-description">
                        <p id="scenarioDescription">Select a scenario to view its parameters and description.</p>
                    </div>
                    
                    <div class="scenario-parameters" id="scenarioParameters">
                        <!-- Parameters will be populated here -->
                    </div>
                    
                    <div class="scenario-results" id="scenarioResults" style="display: none;">
                        <!-- Simulation results will be displayed here -->
                    </div>
                </div>
            </div>
        `;

        this.bindScenarioBuilderEvents(containerId);
        this.updateScenarioParameters();
        this.addScenarioBuilderStyles();

        return container;
    }

    /**
     * Bind scenario builder events
     */
    bindScenarioBuilderEvents(containerId) {
        const container = document.getElementById(containerId);
        
        // Scenario selector change
        const scenarioSelector = container.querySelector('#scenarioSelector');
        scenarioSelector.addEventListener('change', (e) => {
            this.selectScenario(e.target.value);
        });

        // Run simulation button
        const runSimulationBtn = container.querySelector('#runSimulationBtn');
        runSimulationBtn.addEventListener('click', () => {
            this.runSimulation();
        });

        // Compare scenarios button
        const compareScenariosBtn = container.querySelector('#compareScenariosBtn');
        compareScenariosBtn.addEventListener('click', () => {
            this.showScenarioComparison();
        });
    }

    /**
     * Update scenario parameters display
     */
    updateScenarioParameters() {
        const container = document.getElementById('scenarioParameters');
        if (!container) return;

        const scenario = this.scenarios.get(this.currentScenario);
        if (!scenario) return;

        container.innerHTML = `
            <div class="parameters-grid">
                ${Object.entries(scenario.parameters).map(([key, param]) => `
                    <div class="parameter-item">
                        <label class="parameter-label">
                            ${param.label}
                            <span class="parameter-value">${(param.value * 100).toFixed(1)}%</span>
                        </label>
                        <div class="parameter-control">
                            <input type="range" 
                                   class="parameter-slider" 
                                   data-parameter="${key}"
                                   min="${param.min}" 
                                   max="${param.max}" 
                                   step="${param.step}" 
                                   value="${param.value}"
                                   style="--color: ${scenario.color}">
                            <div class="parameter-range">
                                <span>${(param.min * 100).toFixed(0)}%</span>
                                <span>${(param.max * 100).toFixed(0)}%</span>
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;

        // Bind slider events
        container.querySelectorAll('.parameter-slider').forEach(slider => {
            slider.addEventListener('input', (e) => {
                const parameter = e.target.getAttribute('data-parameter');
                const value = parseFloat(e.target.value);
                this.updateScenarioParameter(parameter, value);
            });
        });

        // Update scenario description
        const descriptionElement = document.getElementById('scenarioDescription');
        if (descriptionElement) {
            descriptionElement.textContent = scenario.description;
        }
    }

    /**
     * Update scenario parameter
     */
    updateScenarioParameter(parameter, value) {
        const scenario = this.scenarios.get(this.currentScenario);
        if (scenario && scenario.parameters[parameter]) {
            scenario.parameters[parameter].value = value;
            
            // Update display
            const slider = document.querySelector(`[data-parameter="${parameter}"]`);
            if (slider) {
                const valueDisplay = slider.parentElement.querySelector('.parameter-value');
                if (valueDisplay) {
                    valueDisplay.textContent = `${(value * 100).toFixed(1)}%`;
                }
            }

            // Emit event
            document.dispatchEvent(new CustomEvent('scenarioParameterChanged', {
                detail: {
                    scenarioId: this.currentScenario,
                    parameter,
                    value
                }
            }));
        }
    }

    /**
     * Select scenario
     */
    selectScenario(scenarioId) {
        if (this.scenarios.has(scenarioId)) {
            this.currentScenario = scenarioId;
            this.updateScenarioParameters();
            
            // Update selector
            const selector = document.getElementById('scenarioSelector');
            if (selector) {
                selector.value = scenarioId;
            }

            // Emit event
            document.dispatchEvent(new CustomEvent('scenarioSelected', {
                detail: { scenarioId }
            }));
        }
    }

    /**
     * Run Monte Carlo simulation
     */
    async runSimulation() {
        const runBtn = document.getElementById('runSimulationBtn');
        const resultsContainer = document.getElementById('scenarioResults');
        
        // Show loading state
        runBtn.disabled = true;
        runBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Running Simulation...';
        
        resultsContainer.style.display = 'block';
        resultsContainer.innerHTML = `
            <div class="simulation-loading">
                <div class="loading-spinner"></div>
                <p>Running Monte Carlo simulation with ${this.options.simulationRuns} iterations...</p>
            </div>
        `;

        try {
            // Simulate API call delay
            await new Promise(resolve => setTimeout(resolve, 2000));

            // Generate simulation results
            const results = this.generateSimulationResults();
            this.simulationResults = results;

            // Display results
            this.displaySimulationResults(results);

        } catch (error) {
            console.error('Simulation error:', error);
            resultsContainer.innerHTML = `
                <div class="simulation-error">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error running simulation: ${error.message}</p>
                </div>
            `;
        } finally {
            // Reset button
            runBtn.disabled = false;
            runBtn.innerHTML = '<i class="fas fa-play"></i> Run Simulation';
        }
    }

    /**
     * Generate simulation results
     */
    generateSimulationResults() {
        const scenario = this.scenarios.get(this.currentScenario);
        const results = {
            scenario: scenario.name,
            iterations: this.options.simulationRuns,
            riskScores: [],
            outcomes: {
                low: 0,
                medium: 0,
                high: 0,
                critical: 0
            },
            statistics: {
                mean: 0,
                median: 0,
                stdDev: 0,
                percentile95: 0,
                percentile5: 0
            }
        };

        // Generate risk scores based on scenario parameters
        for (let i = 0; i < this.options.simulationRuns; i++) {
            const riskScore = this.calculateRiskScore(scenario.parameters);
            results.riskScores.push(riskScore);

            // Categorize outcome
            if (riskScore <= 3) results.outcomes.low++;
            else if (riskScore <= 6) results.outcomes.medium++;
            else if (riskScore <= 8) results.outcomes.high++;
            else results.outcomes.critical++;
        }

        // Calculate statistics
        results.riskScores.sort((a, b) => a - b);
        results.statistics.mean = results.riskScores.reduce((sum, score) => sum + score, 0) / results.riskScores.length;
        results.statistics.median = results.riskScores[Math.floor(results.riskScores.length / 2)];
        results.statistics.stdDev = Math.sqrt(results.riskScores.reduce((sum, score) => sum + Math.pow(score - results.statistics.mean, 2), 0) / results.riskScores.length);
        results.statistics.percentile95 = results.riskScores[Math.floor(results.riskScores.length * 0.95)];
        results.statistics.percentile5 = results.riskScores[Math.floor(results.riskScores.length * 0.05)];

        return results;
    }

    /**
     * Calculate risk score from parameters
     */
    calculateRiskScore(parameters) {
        // Add some randomness to simulate Monte Carlo
        const randomFactor = (Math.random() - 0.5) * 0.2;
        
        // Weighted calculation of risk score
        const weights = {
            revenueGrowth: 0.25,
            marketVolatility: 0.2,
            operationalEfficiency: 0.2,
            complianceScore: 0.15,
            customerSatisfaction: 0.1,
            employeeRetention: 0.1
        };

        let riskScore = 5; // Base score

        Object.entries(parameters).forEach(([key, param]) => {
            const weight = weights[key] || 0;
            const impact = (param.value - 0.5) * 10; // Convert to -5 to +5 range
            riskScore += impact * weight;
        });

        // Add random factor
        riskScore += randomFactor;

        // Clamp to 0-10 range
        return Math.max(0, Math.min(10, riskScore));
    }

    /**
     * Display simulation results
     */
    displaySimulationResults(results) {
        const container = document.getElementById('scenarioResults');
        
        container.innerHTML = `
            <div class="simulation-results">
                <div class="results-header">
                    <h4>Simulation Results: ${results.scenario}</h4>
                    <div class="results-summary">
                        <span class="iterations">${results.iterations.toLocaleString()} iterations</span>
                    </div>
                </div>

                <div class="results-grid">
                    <div class="results-statistics">
                        <h5>Risk Score Statistics</h5>
                        <div class="stat-item">
                            <span class="stat-label">Mean:</span>
                            <span class="stat-value">${results.statistics.mean.toFixed(2)}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Median:</span>
                            <span class="stat-value">${results.statistics.median.toFixed(2)}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Std Dev:</span>
                            <span class="stat-value">${results.statistics.stdDev.toFixed(2)}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">95th Percentile:</span>
                            <span class="stat-value">${results.statistics.percentile95.toFixed(2)}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">5th Percentile:</span>
                            <span class="stat-value">${results.statistics.percentile5.toFixed(2)}</span>
                        </div>
                    </div>

                    <div class="results-outcomes">
                        <h5>Risk Level Distribution</h5>
                        <div class="outcome-chart">
                            <div class="outcome-bar low" style="width: ${(results.outcomes.low / results.iterations) * 100}%">
                                <span class="outcome-label">Low (${((results.outcomes.low / results.iterations) * 100).toFixed(1)}%)</span>
                            </div>
                            <div class="outcome-bar medium" style="width: ${(results.outcomes.medium / results.iterations) * 100}%">
                                <span class="outcome-label">Medium (${((results.outcomes.medium / results.iterations) * 100).toFixed(1)}%)</span>
                            </div>
                            <div class="outcome-bar high" style="width: ${(results.outcomes.high / results.iterations) * 100}%">
                                <span class="outcome-label">High (${((results.outcomes.high / results.iterations) * 100).toFixed(1)}%)</span>
                            </div>
                            <div class="outcome-bar critical" style="width: ${(results.outcomes.critical / results.iterations) * 100}%">
                                <span class="outcome-label">Critical (${((results.outcomes.critical / results.iterations) * 100).toFixed(1)}%)</span>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="results-chart">
                    <h5>Risk Score Distribution</h5>
                    <div id="riskDistributionChart" style="height: 300px;"></div>
                </div>
            </div>
        `;

        // Create distribution chart
        this.createRiskDistributionChart('riskDistributionChart', results.riskScores);
    }

    /**
     * Create risk distribution chart
     */
    createRiskDistributionChart(containerId, riskScores) {
        const container = document.getElementById(containerId);
        if (!container) return;

        // Create histogram data
        const bins = 20;
        const binSize = 10 / bins;
        const histogram = new Array(bins).fill(0);

        riskScores.forEach(score => {
            const binIndex = Math.min(Math.floor(score / binSize), bins - 1);
            histogram[binIndex]++;
        });

        const maxCount = Math.max(...histogram);
        const normalizedHistogram = histogram.map(count => count / maxCount);

        // Create D3 visualization
        const width = container.clientWidth || 400;
        const height = container.clientHeight || 300;
        const margin = { top: 20, right: 20, bottom: 40, left: 40 };

        const svg = d3.select(container)
            .append('svg')
            .attr('width', width)
            .attr('height', height);

        const g = svg.append('g')
            .attr('transform', `translate(${margin.left}, ${margin.top})`);

        const plotWidth = width - margin.left - margin.right;
        const plotHeight = height - margin.top - margin.bottom;

        const xScale = d3.scaleLinear()
            .domain([0, 10])
            .range([0, plotWidth]);

        const yScale = d3.scaleLinear()
            .domain([0, 1])
            .range([plotHeight, 0]);

        // Create bars
        g.selectAll('.histogram-bar')
            .data(normalizedHistogram)
            .enter().append('rect')
            .attr('class', 'histogram-bar')
            .attr('x', (d, i) => xScale(i * binSize))
            .attr('y', d => yScale(d))
            .attr('width', xScale(binSize))
            .attr('height', d => plotHeight - yScale(d))
            .attr('fill', '#3498db')
            .attr('opacity', 0.7);

        // Add axes
        g.append('g')
            .attr('transform', `translate(0, ${plotHeight})`)
            .call(d3.axisBottom(xScale))
            .append('text')
            .attr('x', plotWidth / 2)
            .attr('y', 30)
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .text('Risk Score');

        g.append('g')
            .call(d3.axisLeft(yScale))
            .append('text')
            .attr('transform', 'rotate(-90)')
            .attr('y', -30)
            .attr('x', -plotHeight / 2)
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .text('Relative Frequency');
    }

    /**
     * Show scenario comparison
     */
    showScenarioComparison() {
        // Create comparison modal
        const modal = document.createElement('div');
        modal.className = 'scenario-comparison-modal';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Scenario Comparison</h3>
                    <button class="modal-close" onclick="this.closest('.scenario-comparison-modal').remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    <div class="comparison-grid">
                        ${Array.from(this.scenarios.values()).map(scenario => `
                            <div class="comparison-scenario">
                                <h4 style="color: ${scenario.color}">${scenario.name}</h4>
                                <p>${scenario.description}</p>
                                <div class="scenario-parameters-compact">
                                    ${Object.entries(scenario.parameters).map(([key, param]) => `
                                        <div class="parameter-compact">
                                            <span class="param-name">${param.label}:</span>
                                            <span class="param-value">${(param.value * 100).toFixed(1)}%</span>
                                        </div>
                                    `).join('')}
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(modal);
        this.addComparisonModalStyles();
    }

    /**
     * Add scenario builder styles
     */
    addScenarioBuilderStyles() {
        if (document.getElementById('scenario-builder-styles')) return;

        const style = document.createElement('style');
        style.id = 'scenario-builder-styles';
        style.textContent = `
            .scenario-builder {
                background: #fff;
                border-radius: 10px;
                padding: 20px;
                box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            }

            .scenario-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 20px;
                padding-bottom: 15px;
                border-bottom: 2px solid #ecf0f1;
            }

            .scenario-controls {
                display: flex;
                gap: 10px;
                align-items: center;
            }

            .scenario-selector {
                padding: 8px 12px;
                border: 1px solid #ddd;
                border-radius: 5px;
                background: white;
            }

            .parameters-grid {
                display: grid;
                gap: 20px;
                margin: 20px 0;
            }

            .parameter-item {
                background: #f8f9fa;
                padding: 15px;
                border-radius: 8px;
                border-left: 4px solid var(--color, #3498db);
            }

            .parameter-label {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 10px;
                font-weight: 600;
                color: #2c3e50;
            }

            .parameter-value {
                background: var(--color, #3498db);
                color: white;
                padding: 4px 8px;
                border-radius: 4px;
                font-size: 0.9em;
            }

            .parameter-control {
                position: relative;
            }

            .parameter-slider {
                width: 100%;
                height: 6px;
                border-radius: 3px;
                background: #ddd;
                outline: none;
                -webkit-appearance: none;
            }

            .parameter-slider::-webkit-slider-thumb {
                -webkit-appearance: none;
                appearance: none;
                width: 20px;
                height: 20px;
                border-radius: 50%;
                background: var(--color, #3498db);
                cursor: pointer;
                border: 2px solid white;
                box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
            }

            .parameter-slider::-moz-range-thumb {
                width: 20px;
                height: 20px;
                border-radius: 50%;
                background: var(--color, #3498db);
                cursor: pointer;
                border: 2px solid white;
                box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
            }

            .parameter-range {
                display: flex;
                justify-content: space-between;
                margin-top: 5px;
                font-size: 0.8em;
                color: #666;
            }

            .simulation-loading {
                text-align: center;
                padding: 40px;
            }

            .loading-spinner {
                width: 40px;
                height: 40px;
                border: 4px solid #f3f3f3;
                border-top: 4px solid #3498db;
                border-radius: 50%;
                animation: spin 1s linear infinite;
                margin: 0 auto 20px;
            }

            @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }

            .simulation-results {
                margin-top: 20px;
            }

            .results-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 20px;
                padding-bottom: 10px;
                border-bottom: 1px solid #eee;
            }

            .results-grid {
                display: grid;
                grid-template-columns: 1fr 1fr;
                gap: 20px;
                margin-bottom: 20px;
            }

            .results-statistics, .results-outcomes {
                background: #f8f9fa;
                padding: 15px;
                border-radius: 8px;
            }

            .stat-item {
                display: flex;
                justify-content: space-between;
                padding: 5px 0;
                border-bottom: 1px solid #eee;
            }

            .stat-label {
                font-weight: 600;
            }

            .stat-value {
                color: #3498db;
                font-weight: bold;
            }

            .outcome-chart {
                display: flex;
                height: 30px;
                border-radius: 5px;
                overflow: hidden;
                margin-top: 10px;
            }

            .outcome-bar {
                display: flex;
                align-items: center;
                justify-content: center;
                color: white;
                font-size: 0.8em;
                font-weight: bold;
            }

            .outcome-bar.low { background: #27ae60; }
            .outcome-bar.medium { background: #f39c12; }
            .outcome-bar.high { background: #e74c3c; }
            .outcome-bar.critical { background: #8e44ad; }

            .results-chart {
                background: #f8f9fa;
                padding: 15px;
                border-radius: 8px;
            }

            .simulation-error {
                text-align: center;
                padding: 40px;
                color: #e74c3c;
            }

            .simulation-error i {
                font-size: 2em;
                margin-bottom: 10px;
            }

            @media (max-width: 768px) {
                .scenario-header {
                    flex-direction: column;
                    gap: 15px;
                    align-items: stretch;
                }

                .scenario-controls {
                    justify-content: center;
                }

                .results-grid {
                    grid-template-columns: 1fr;
                }
            }
        `;

        document.head.appendChild(style);
    }

    /**
     * Add comparison modal styles
     */
    addComparisonModalStyles() {
        if (document.getElementById('comparison-modal-styles')) return;

        const style = document.createElement('style');
        style.id = 'comparison-modal-styles';
        style.textContent = `
            .scenario-comparison-modal {
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: rgba(0, 0, 0, 0.5);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 1000;
            }

            .modal-content {
                background: white;
                border-radius: 10px;
                max-width: 90%;
                max-height: 90%;
                overflow: auto;
                box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
            }

            .modal-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 20px;
                border-bottom: 1px solid #eee;
            }

            .modal-close {
                background: none;
                border: none;
                font-size: 1.5em;
                cursor: pointer;
                color: #666;
            }

            .modal-body {
                padding: 20px;
            }

            .comparison-grid {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
                gap: 20px;
            }

            .comparison-scenario {
                background: #f8f9fa;
                padding: 15px;
                border-radius: 8px;
                border-left: 4px solid #3498db;
            }

            .scenario-parameters-compact {
                margin-top: 10px;
            }

            .parameter-compact {
                display: flex;
                justify-content: space-between;
                padding: 3px 0;
                font-size: 0.9em;
            }

            .param-name {
                color: #666;
            }

            .param-value {
                font-weight: bold;
                color: #2c3e50;
            }
        `;

        document.head.appendChild(style);
    }

    /**
     * Update scenario
     */
    updateScenario(scenarioId, parameter, value) {
        const scenario = this.scenarios.get(scenarioId);
        if (scenario && scenario.parameters[parameter]) {
            scenario.parameters[parameter].value = value;
        }
    }

    /**
     * Get scenario
     */
    getScenario(scenarioId) {
        return this.scenarios.get(scenarioId);
    }

    /**
     * Get all scenarios
     */
    getAllScenarios() {
        return Array.from(this.scenarios.values());
    }

    /**
     * Destroy component
     */
    destroy() {
        this.scenarios.clear();
        this.visualizations.clear();
        this.simulationResults = null;
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskScenarioAnalysis;
}
