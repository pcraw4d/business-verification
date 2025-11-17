/**
 * Risk Explainability Component
 * 
 * Provides SHAP-based risk factor explanation capabilities:
 * - Interactive SHAP force plot visualization (D3.js)
 * - Feature importance waterfall charts
 * - Risk factor contribution breakdown
 * - Tooltips with detailed explanations
 * - "Why this score?" expandable panels
 */

class RiskExplainability {
    constructor(options = {}) {
        this.options = {
            animationDuration: 1000,
            colorScheme: {
                positive: '#27ae60',
                negative: '#e74c3c',
                neutral: '#95a5a6'
            },
            ...options
        };

        this.visualizations = new Map();
        this.explanations = new Map();
        this.currentMerchantId = null;

        this.init();
    }

    /**
     * Initialize the explainability component
     */
    init() {
        this.setupEventListeners();
        this.initializeColorScales();
    }

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Listen for risk updates
        document.addEventListener('riskUpdate', (event) => {
            this.updateExplanations(event.detail.merchantId, event.detail.riskData);
        });

        // Listen for merchant changes
        document.addEventListener('merchantChanged', (event) => {
            this.currentMerchantId = event.detail.merchantId;
            this.loadExplanations(event.detail.merchantId);
        });
    }

    /**
     * Initialize D3 color scales
     */
    initializeColorScales() {
        this.contributionScale = d3.scaleLinear()
            .domain([-1, 0, 1])
            .range([this.options.colorScheme.negative, this.options.colorScheme.neutral, this.options.colorScheme.positive]);

        this.importanceScale = d3.scaleLinear()
            .domain([0, 1])
            .range([0.3, 1]);
    }

    /**
     * Create SHAP force plot visualization (D3.js)
     */
    createSHAPForcePlot(containerId, shapData) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const width = container.clientWidth || 800;
        const height = container.clientHeight || 400;
        const margin = { top: 20, right: 20, bottom: 20, left: 20 };

        // Clear existing content
        container.innerHTML = '';

        const svg = d3.select(container)
            .append('svg')
            .attr('width', width)
            .attr('height', height);

        const g = svg.append('g')
            .attr('transform', `translate(${margin.left}, ${margin.top})`);

        const plotWidth = width - margin.left - margin.right;
        const plotHeight = height - margin.top - margin.bottom;

        // Create force simulation
        const simulation = d3.forceSimulation(shapData.features)
            .force('x', d3.forceX(d => {
                // Position based on SHAP value
                return (d.shapValue + 1) * plotWidth / 2;
            }).strength(1))
            .force('y', d3.forceY(plotHeight / 2).strength(0.1))
            .force('collide', d3.forceCollide().radius(d => Math.abs(d.shapValue) * 20 + 10))
            .stop();

        // Run simulation
        for (let i = 0; i < 100; ++i) simulation.tick();

        // Create background line
        g.append('line')
            .attr('x1', plotWidth / 2)
            .attr('y1', 0)
            .attr('x2', plotWidth / 2)
            .attr('y2', plotHeight)
            .attr('stroke', '#ddd')
            .attr('stroke-width', 2)
            .attr('stroke-dasharray', '5,5');

        // Create feature circles
        const circles = g.selectAll('.feature-circle')
            .data(shapData.features)
            .enter().append('circle')
            .attr('class', 'feature-circle')
            .attr('cx', d => d.x)
            .attr('cy', d => d.y)
            .attr('r', d => Math.abs(d.shapValue) * 15 + 8)
            .attr('fill', d => this.contributionScale(d.shapValue))
            .attr('stroke', '#fff')
            .attr('stroke-width', 2)
            .attr('opacity', 0.8)
            .style('cursor', 'pointer')
            .on('mouseover', function(event, d) {
                showTooltip(event, d);
                d3.select(this)
                    .attr('opacity', 1)
                    .attr('stroke-width', 3);
            })
            .on('mousemove', function(event, d) {
                updateTooltip(event, d);
            })
            .on('mouseout', function() {
                hideTooltip();
                d3.select(this)
                    .attr('opacity', 0.8)
                    .attr('stroke-width', 2);
            })
            .on('click', function(event, d) {
                showFeatureDetails(d);
            });

        // Add feature labels
        const labels = g.selectAll('.feature-label')
            .data(shapData.features)
            .enter().append('text')
            .attr('class', 'feature-label')
            .attr('x', d => d.x)
            .attr('y', d => d.y)
            .attr('text-anchor', 'middle')
            .attr('dy', '0.35em')
            .attr('font-size', '10px')
            .attr('font-weight', 'bold')
            .attr('fill', '#fff')
            .text(d => d.name.length > 8 ? d.name.substring(0, 8) + '...' : d.name);

        // Create tooltip
        const tooltip = d3.select('body').append('div')
            .attr('class', 'shap-tooltip')
            .style('position', 'absolute')
            .style('background', 'rgba(0, 0, 0, 0.9)')
            .style('color', 'white')
            .style('padding', '12px')
            .style('border-radius', '6px')
            .style('font-size', '12px')
            .style('pointer-events', 'none')
            .style('z-index', '1000')
            .style('opacity', 0)
            .style('transition', 'opacity 0.3s');

        function showTooltip(event, d) {
            tooltip.html(`
                <div style="font-weight: bold; margin-bottom: 4px;">${d.name}</div>
                <div>SHAP Value: <span style="color: ${d.shapValue > 0 ? '#27ae60' : '#e74c3c'}">${d.shapValue > 0 ? '+' : ''}${d.shapValue.toFixed(3)}</span></div>
                <div>Feature Value: ${d.featureValue}</div>
                <div>Contribution: ${d.contribution.toFixed(3)}</div>
                <div style="margin-top: 4px; font-size: 10px; color: #ccc;">${d.description}</div>
            `)
            .style('opacity', 1);
        }

        function updateTooltip(event, d) {
            tooltip
                .style('left', (event.pageX + 10) + 'px')
                .style('top', (event.pageY - 10) + 'px');
        }

        function hideTooltip() {
            tooltip.style('opacity', 0);
        }

        function showFeatureDetails(d) {
            // Create modal or expandable panel with detailed feature information
            createFeatureDetailModal(d);
        }

        // Add axis labels
        g.append('text')
            .attr('x', 10)
            .attr('y', 20)
            .attr('font-size', '12px')
            .attr('font-weight', 'bold')
            .attr('fill', '#e74c3c')
            .text('Decreases Risk');

        g.append('text')
            .attr('x', plotWidth - 10)
            .attr('y', 20)
            .attr('text-anchor', 'end')
            .attr('font-size', '12px')
            .attr('font-weight', 'bold')
            .attr('fill', '#27ae60')
            .text('Increases Risk');

        // Store visualization reference
        const visualization = {
            svg,
            g,
            circles,
            labels,
            tooltip,
            update: (newData) => this.updateSHAPForcePlot(visualization, newData)
        };

        this.visualizations.set(containerId, visualization);
        return visualization;
    }

    /**
     * Create feature importance waterfall chart (D3.js)
     */
    createFeatureImportanceWaterfall(containerId, importanceData) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const width = container.clientWidth || 600;
        const height = container.clientHeight || 400;
        const margin = { top: 20, right: 20, bottom: 60, left: 80 };

        // Clear existing content
        container.innerHTML = '';

        const svg = d3.select(container)
            .append('svg')
            .attr('width', width)
            .attr('height', height);

        const g = svg.append('g')
            .attr('transform', `translate(${margin.left}, ${margin.top})`);

        const plotWidth = width - margin.left - margin.right;
        const plotHeight = height - margin.top - margin.bottom;

        // Sort data by importance
        const sortedData = importanceData.sort((a, b) => Math.abs(b.importance) - Math.abs(a.importance));

        // Create scales
        const xScale = d3.scaleBand()
            .domain(sortedData.map(d => d.name))
            .range([0, plotWidth])
            .padding(0.1);

        const yScale = d3.scaleLinear()
            .domain([0, d3.max(sortedData, d => Math.abs(d.importance))])
            .range([plotHeight, 0]);

        // Create bars
        const bars = g.selectAll('.importance-bar')
            .data(sortedData)
            .enter().append('rect')
            .attr('class', 'importance-bar')
            .attr('x', d => xScale(d.name))
            .attr('y', d => d.importance > 0 ? yScale(d.importance) : yScale(0))
            .attr('width', xScale.bandwidth())
            .attr('height', d => Math.abs(yScale(d.importance) - yScale(0)))
            .attr('fill', d => this.contributionScale(d.importance > 0 ? 1 : -1))
            .attr('opacity', 0.8)
            .style('cursor', 'pointer')
            .on('mouseover', function(event, d) {
                showBarTooltip(event, d);
                d3.select(this)
                    .attr('opacity', 1)
                    .attr('stroke', '#fff')
                    .attr('stroke-width', 2);
            })
            .on('mousemove', function(event, d) {
                updateBarTooltip(event, d);
            })
            .on('mouseout', function() {
                hideBarTooltip();
                d3.select(this)
                    .attr('opacity', 0.8)
                    .attr('stroke', 'none');
            });

        // Add value labels on bars
        g.selectAll('.bar-label')
            .data(sortedData)
            .enter().append('text')
            .attr('class', 'bar-label')
            .attr('x', d => xScale(d.name) + xScale.bandwidth() / 2)
            .attr('y', d => d.importance > 0 ? yScale(d.importance) - 5 : yScale(0) + 15)
            .attr('text-anchor', 'middle')
            .attr('font-size', '10px')
            .attr('font-weight', 'bold')
            .attr('fill', '#2c3e50')
            .text(d => d.importance.toFixed(3));

        // Add x-axis
        g.append('g')
            .attr('transform', `translate(0, ${plotHeight})`)
            .call(d3.axisBottom(xScale))
            .selectAll('text')
            .attr('transform', 'rotate(-45)')
            .attr('text-anchor', 'end')
            .attr('dx', '-0.5em')
            .attr('dy', '0.5em');

        // Add y-axis
        g.append('g')
            .call(d3.axisLeft(yScale))
            .append('text')
            .attr('transform', 'rotate(-90)')
            .attr('y', -50)
            .attr('x', -plotHeight / 2)
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .attr('font-weight', 'bold')
            .attr('fill', '#2c3e50')
            .text('Feature Importance');

        // Create bar tooltip
        const barTooltip = d3.select('body').append('div')
            .attr('class', 'waterfall-tooltip')
            .style('position', 'absolute')
            .style('background', 'rgba(0, 0, 0, 0.9)')
            .style('color', 'white')
            .style('padding', '12px')
            .style('border-radius', '6px')
            .style('font-size', '12px')
            .style('pointer-events', 'none')
            .style('z-index', '1000')
            .style('opacity', 0)
            .style('transition', 'opacity 0.3s');

        function showBarTooltip(event, d) {
            barTooltip.html(`
                <div style="font-weight: bold; margin-bottom: 4px;">${d.name}</div>
                <div>Importance: <span style="color: ${d.importance > 0 ? '#27ae60' : '#e74c3c'}">${d.importance > 0 ? '+' : ''}${d.importance.toFixed(3)}</span></div>
                <div>Percentage: ${((Math.abs(d.importance) / d3.sum(sortedData.map(x => Math.abs(x.importance)))) * 100).toFixed(1)}%</div>
                <div style="margin-top: 4px; font-size: 10px; color: #ccc;">${d.description}</div>
            `)
            .style('opacity', 1);
        }

        function updateBarTooltip(event, d) {
            barTooltip
                .style('left', (event.pageX + 10) + 'px')
                .style('top', (event.pageY - 10) + 'px');
        }

        function hideBarTooltip() {
            barTooltip.style('opacity', 0);
        }

        // Store visualization reference
        const visualization = {
            svg,
            g,
            bars,
            update: (newData) => this.updateFeatureImportanceWaterfall(visualization, newData)
        };

        this.visualizations.set(containerId, visualization);
        return visualization;
    }

    /**
     * Create "Why this score?" expandable panel
     */
    createWhyScorePanel(containerId, explanationData) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        container.innerHTML = `
            <div class="why-score-panel">
                <div class="panel-header" onclick="toggleWhyScorePanel()">
                    <h3>
                        <i class="fas fa-question-circle"></i>
                        Why is the risk score ${explanationData.overallScore.toFixed(1)}?
                    </h3>
                    <i class="fas fa-chevron-down panel-toggle"></i>
                </div>
                <div class="panel-content" style="display: none;">
                    <div class="score-breakdown">
                        <div class="base-score">
                            <span class="label">Base Score:</span>
                            <span class="value">${explanationData.baseScore.toFixed(1)}</span>
                        </div>
                        <div class="contributions">
                            ${explanationData.contributions.map(contrib => `
                                <div class="contribution-item ${contrib.impact > 0 ? 'positive' : 'negative'}">
                                    <span class="feature-name">${contrib.feature}</span>
                                    <span class="impact-value">${contrib.impact > 0 ? '+' : ''}${contrib.impact.toFixed(2)}</span>
                                    <span class="impact-reason">${contrib.reason}</span>
                                </div>
                            `).join('')}
                        </div>
                        <div class="total-score">
                            <span class="label">Total Score:</span>
                            <span class="value">${explanationData.overallScore.toFixed(1)}</span>
                        </div>
                    </div>
                    <div class="explanation-text">
                        <p>${explanationData.summary}</p>
                        <div class="key-factors">
                            <h4>Key Risk Factors:</h4>
                            <ul>
                                ${explanationData.keyFactors.map(factor => `
                                    <li>
                                        <strong>${factor.name}:</strong> ${factor.description}
                                        <span class="factor-impact ${factor.impact > 0 ? 'positive' : 'negative'}">
                                            (${factor.impact > 0 ? '+' : ''}${factor.impact.toFixed(2)})
                                        </span>
                                    </li>
                                `).join('')}
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        `;

        // Add CSS styles
        this.addWhyScorePanelStyles();

        // Store reference
        this.explanations.set(containerId, {
            container,
            data: explanationData,
            update: (newData) => this.updateWhyScorePanel(containerId, newData)
        });

        return this.explanations.get(containerId);
    }

    /**
     * Add CSS styles for why score panel
     */
    addWhyScorePanelStyles() {
        if (document.getElementById('why-score-panel-styles')) return;

        const style = document.createElement('style');
        style.id = 'why-score-panel-styles';
        style.textContent = `
            .why-score-panel {
                background: #fff;
                border-radius: 10px;
                box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
                margin: 20px 0;
                overflow: hidden;
            }

            .panel-header {
                background: linear-gradient(135deg, #3498db, #2980b9);
                color: white;
                padding: 15px 20px;
                cursor: pointer;
                display: flex;
                justify-content: space-between;
                align-items: center;
                transition: background 0.3s ease;
            }

            .panel-header:hover {
                background: linear-gradient(135deg, #2980b9, #1f4e79);
            }

            .panel-header h3 {
                margin: 0;
                font-size: 1.1rem;
                display: flex;
                align-items: center;
                gap: 10px;
            }

            .panel-toggle {
                transition: transform 0.3s ease;
            }

            .panel-toggle.rotated {
                transform: rotate(180deg);
            }

            .panel-content {
                padding: 20px;
                animation: slideDown 0.3s ease;
            }

            @keyframes slideDown {
                from { opacity: 0; transform: translateY(-10px); }
                to { opacity: 1; transform: translateY(0); }
            }

            .score-breakdown {
                margin-bottom: 20px;
            }

            .base-score, .total-score {
                display: flex;
                justify-content: space-between;
                padding: 10px 0;
                border-bottom: 1px solid #eee;
                font-weight: bold;
            }

            .total-score {
                border-top: 2px solid #3498db;
                background: #f8f9fa;
                margin-top: 10px;
                padding: 15px;
                border-radius: 5px;
            }

            .contributions {
                margin: 15px 0;
            }

            .contribution-item {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 8px 0;
                border-bottom: 1px solid #f0f0f0;
            }

            .contribution-item.positive .impact-value {
                color: #27ae60;
            }

            .contribution-item.negative .impact-value {
                color: #e74c3c;
            }

            .feature-name {
                font-weight: 600;
                flex: 1;
            }

            .impact-value {
                font-weight: bold;
                margin: 0 10px;
            }

            .impact-reason {
                font-size: 0.9em;
                color: #666;
                flex: 2;
            }

            .explanation-text p {
                margin-bottom: 15px;
                line-height: 1.6;
            }

            .key-factors h4 {
                margin-bottom: 10px;
                color: #2c3e50;
            }

            .key-factors ul {
                list-style: none;
                padding: 0;
            }

            .key-factors li {
                padding: 8px 0;
                border-bottom: 1px solid #f0f0f0;
            }

            .factor-impact {
                font-weight: bold;
                margin-left: 10px;
            }

            .factor-impact.positive {
                color: #27ae60;
            }

            .factor-impact.negative {
                color: #e74c3c;
            }
        `;

        document.head.appendChild(style);
    }

    /**
     * Update SHAP force plot
     */
    updateSHAPForcePlot(visualization, newData) {
        // Update data
        visualization.circles.data(newData.features);
        visualization.labels.data(newData.features);

        // Update positions and attributes
        visualization.circles
            .transition()
            .duration(this.options.animationDuration)
            .attr('cx', d => d.x)
            .attr('cy', d => d.y)
            .attr('r', d => Math.abs(d.shapValue) * 15 + 8)
            .attr('fill', d => this.contributionScale(d.shapValue));

        visualization.labels
            .transition()
            .duration(this.options.animationDuration)
            .attr('x', d => d.x)
            .attr('y', d => d.y)
            .text(d => d.name.length > 8 ? d.name.substring(0, 8) + '...' : d.name);
    }

    /**
     * Update feature importance waterfall
     */
    updateFeatureImportanceWaterfall(visualization, newData) {
        // Recreate the chart with new data
        const containerId = visualization.svg.node().parentElement.id;
        this.visualizations.delete(containerId);
        return this.createFeatureImportanceWaterfall(containerId, newData);
    }

    /**
     * Update why score panel
     */
    updateWhyScorePanel(containerId, newData) {
        const explanation = this.explanations.get(containerId);
        if (explanation) {
            explanation.data = newData;
            this.createWhyScorePanel(containerId, newData);
        }
    }

    /**
     * Load explanations for a merchant
     */
    async loadExplanations(merchantId) {
        try {
            const endpoints = APIConfig.getEndpoints();
            const response = await fetch(endpoints.riskExplain(merchantId), {
                headers: APIConfig.getHeaders()
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const explanationData = await response.json();
            this.explanations.set(merchantId, explanationData);
            return explanationData;

        } catch (error) {
            console.error('Error loading explanations:', error);
            return null;
        }
    }

    /**
     * Update explanations when risk data changes
     */
    updateExplanations(merchantId, riskData) {
        if (this.explanations.has(merchantId)) {
            const currentExplanation = this.explanations.get(merchantId);
            // Update with new risk data
            this.updateExplanationData(currentExplanation, riskData);
        }
    }

    /**
     * Update explanation data with new risk information
     */
    updateExplanationData(explanation, riskData) {
        // Update overall score
        explanation.overallScore = riskData.overallScore;
        
        // Update contributions based on new risk factors
        explanation.contributions = riskData.factorContributions || explanation.contributions;
        
        // Update summary
        explanation.summary = this.generateSummary(riskData);
        
        // Update key factors
        explanation.keyFactors = this.extractKeyFactors(riskData);
    }

    /**
     * Generate explanation summary
     */
    generateSummary(riskData) {
        const score = riskData.overallScore;
        const trend = riskData.trend || 0;
        
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
     * Extract key factors from risk data
     */
    extractKeyFactors(riskData) {
        const factors = [];
        
        if (riskData.categories) {
            Object.entries(riskData.categories).forEach(([category, score]) => {
                if (Math.abs(score - 5) > 1) { // Significant deviation from neutral
                    factors.push({
                        name: category.charAt(0).toUpperCase() + category.slice(1) + ' Risk',
                        description: this.getCategoryDescription(category, score),
                        impact: score - 5
                    });
                }
            });
        }

        return factors.sort((a, b) => Math.abs(b.impact) - Math.abs(a.impact)).slice(0, 5);
    }

    /**
     * Get category description
     */
    getCategoryDescription(category, score) {
        const descriptions = {
            financial: score > 5 ? 'Financial instability and cash flow concerns' : 'Strong financial position and stable cash flow',
            operational: score > 5 ? 'Operational inefficiencies and process issues' : 'Efficient operations and well-defined processes',
            compliance: score > 5 ? 'Compliance gaps and regulatory concerns' : 'Strong compliance record and regulatory adherence',
            market: score > 5 ? 'Market volatility and competitive pressures' : 'Stable market position and competitive advantage',
            reputation: score > 5 ? 'Reputation concerns and negative sentiment' : 'Strong reputation and positive market sentiment'
        };

        return descriptions[category] || 'Standard risk level for this category';
    }

    /**
     * Destroy all visualizations
     */
    destroy() {
        this.visualizations.forEach(visualization => {
            if (visualization.svg) {
                visualization.svg.remove();
            }
            if (visualization.tooltip) {
                visualization.tooltip.remove();
            }
        });
        this.visualizations.clear();
        this.explanations.clear();
    }
}

// Global function for panel toggle
function toggleWhyScorePanel() {
    const panel = document.querySelector('.why-score-panel');
    const content = panel.querySelector('.panel-content');
    const toggle = panel.querySelector('.panel-toggle');
    
    if (content.style.display === 'none') {
        content.style.display = 'block';
        toggle.classList.add('rotated');
    } else {
        content.style.display = 'none';
        toggle.classList.remove('rotated');
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskExplainability;
}
