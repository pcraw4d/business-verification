/**
 * Risk History Tracking Component
 * 
 * Provides comprehensive risk history and trend visualization:
 * - Time-series charts with zoom and pan (D3.js)
 * - Risk score evolution over time
 * - Prediction accuracy tracking
 * - Historical event annotations
 * - Export functionality for reports
 */

class RiskHistoryTracking {
    constructor(options = {}) {
        this.options = {
            animationDuration: 1000,
            defaultTimeRange: 90, // days
            colorScheme: {
                riskScore: '#3498db',
                prediction: '#e74c3c',
                actual: '#27ae60',
                events: '#f39c12'
            },
            ...options
        };

        this.historyData = null;
        this.visualizations = new Map();
        this.currentMerchantId = null;
        this.timeRange = this.options.defaultTimeRange;

        this.init();
    }

    /**
     * Initialize the risk history component
     */
    init() {
        this.setupEventListeners();
        this.initializeTimeScales();
    }

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Listen for merchant changes
        document.addEventListener('merchantChanged', (event) => {
            this.currentMerchantId = event.detail.merchantId;
            this.loadHistoryData(event.detail.merchantId);
        });

        // Listen for time range changes
        document.addEventListener('timeRangeChanged', (event) => {
            this.timeRange = event.detail.timeRange;
            this.updateTimeRange();
        });
    }

    /**
     * Initialize D3 time scales
     */
    initializeTimeScales() {
        this.timeScale = d3.scaleTime();
        this.riskScoreScale = d3.scaleLinear().domain([0, 10]);
        this.predictionAccuracyScale = d3.scaleLinear().domain([0, 1]);
    }

    /**
     * Create risk history timeline (D3.js)
     */
    createRiskHistoryTimeline(containerId, historyData) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const width = container.clientWidth || 800;
        const height = container.clientHeight || 500;
        const margin = { top: 20, right: 80, bottom: 60, left: 60 };

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

        // Set up scales
        this.timeScale.range([0, plotWidth]);
        this.riskScoreScale.range([plotHeight, 0]);

        // Parse dates and set domain
        const dates = historyData.map(d => new Date(d.date));
        this.timeScale.domain(d3.extent(dates));
        this.riskScoreScale.domain([0, 10]);

        // Create line generator
        const line = d3.line()
            .x(d => this.timeScale(new Date(d.date)))
            .y(d => this.riskScoreScale(d.riskScore))
            .curve(d3.curveMonotoneX);

        // Create area generator for confidence band
        const area = d3.area()
            .x(d => this.timeScale(new Date(d.date)))
            .y0(d => this.riskScoreScale(d.confidenceLower))
            .y1(d => this.riskScoreScale(d.confidenceUpper))
            .curve(d3.curveMonotoneX);

        // Add confidence band
        g.append('path')
            .datum(historyData)
            .attr('fill', this.options.colorScheme.prediction)
            .attr('fill-opacity', 0.1)
            .attr('d', area);

        // Add main risk score line
        g.append('path')
            .datum(historyData)
            .attr('fill', 'none')
            .attr('stroke', this.options.colorScheme.riskScore)
            .attr('stroke-width', 3)
            .attr('d', line);

        // Add prediction line if available
        if (historyData.some(d => d.predictedScore !== undefined)) {
            const predictionLine = d3.line()
                .x(d => this.timeScale(new Date(d.date)))
                .y(d => this.riskScoreScale(d.predictedScore))
                .curve(d3.curveMonotoneX);

            g.append('path')
                .datum(historyData.filter(d => d.predictedScore !== undefined))
                .attr('fill', 'none')
                .attr('stroke', this.options.colorScheme.prediction)
                .attr('stroke-width', 2)
                .attr('stroke-dasharray', '5,5')
                .attr('d', predictionLine);
        }

        // Add data points
        const points = g.selectAll('.data-point')
            .data(historyData)
            .enter().append('circle')
            .attr('class', 'data-point')
            .attr('cx', d => this.timeScale(new Date(d.date)))
            .attr('cy', d => this.riskScoreScale(d.riskScore))
            .attr('r', 4)
            .attr('fill', this.options.colorScheme.riskScore)
            .attr('stroke', '#fff')
            .attr('stroke-width', 2)
            .style('cursor', 'pointer')
            .on('mouseover', function(event, d) {
                showDataPointTooltip(event, d);
            })
            .on('mouseout', function() {
                hideDataPointTooltip();
            });

        // Add event annotations
        this.addEventAnnotations(g, historyData, plotHeight);

        // Add axes
        this.addAxes(g, plotWidth, plotHeight);

        // Add zoom behavior
        this.addZoomBehavior(svg, g, plotWidth, plotHeight);

        // Create tooltip
        const tooltip = d3.select('body').append('div')
            .attr('class', 'risk-history-tooltip')
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

        function showDataPointTooltip(event, d) {
            tooltip.html(`
                <div style="font-weight: bold; margin-bottom: 4px;">${new Date(d.date).toLocaleDateString()}</div>
                <div>Risk Score: <span style="color: #3498db">${d.riskScore.toFixed(2)}</span></div>
                ${d.predictedScore ? `<div>Predicted: <span style="color: #e74c3c">${d.predictedScore.toFixed(2)}</span></div>` : ''}
                ${d.accuracy ? `<div>Prediction Accuracy: <span style="color: #27ae60">${(d.accuracy * 100).toFixed(1)}%</span></div>` : ''}
                ${d.events && d.events.length > 0 ? `
                    <div style="margin-top: 8px;">
                        <strong>Events:</strong>
                        ${d.events.map(event => `<div style="font-size: 10px; color: #ccc;">• ${event}</div>`).join('')}
                    </div>
                ` : ''}
            `)
            .style('opacity', 1);
        }

        function hideDataPointTooltip() {
            tooltip.style('opacity', 0);
        }

        // Store visualization reference
        const visualization = {
            svg,
            g,
            line,
            area,
            points,
            tooltip,
            update: (newData) => this.updateRiskHistoryTimeline(visualization, newData)
        };

        this.visualizations.set(containerId, visualization);
        return visualization;
    }

    /**
     * Add event annotations to timeline
     */
    addEventAnnotations(g, historyData, plotHeight) {
        const events = historyData.filter(d => d.events && d.events.length > 0);
        
        events.forEach(d => {
            const x = this.timeScale(new Date(d.date));
            const y = this.riskScoreScale(d.riskScore);

            // Add event marker
            g.append('circle')
                .attr('cx', x)
                .attr('cy', y)
                .attr('r', 6)
                .attr('fill', this.options.colorScheme.events)
                .attr('stroke', '#fff')
                .attr('stroke-width', 2);

            // Add event line
            g.append('line')
                .attr('x1', x)
                .attr('y1', y)
                .attr('x2', x)
                .attr('y2', plotHeight + 10)
                .attr('stroke', this.options.colorScheme.events)
                .attr('stroke-width', 1)
                .attr('stroke-dasharray', '3,3');

            // Add event label
            g.append('text')
                .attr('x', x)
                .attr('y', plotHeight + 25)
                .attr('text-anchor', 'middle')
                .attr('font-size', '10px')
                .attr('fill', this.options.colorScheme.events)
                .text(d.events[0].substring(0, 15) + (d.events[0].length > 15 ? '...' : ''));
        });
    }

    /**
     * Add axes to timeline
     */
    addAxes(g, plotWidth, plotHeight) {
        // X-axis (time)
        g.append('g')
            .attr('transform', `translate(0, ${plotHeight})`)
            .call(d3.axisBottom(this.timeScale)
                .tickFormat(d3.timeFormat('%m/%d')))
            .append('text')
            .attr('x', plotWidth / 2)
            .attr('y', 40)
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .attr('font-weight', 'bold')
            .attr('fill', '#2c3e50')
            .text('Date');

        // Y-axis (risk score)
        g.append('g')
            .call(d3.axisLeft(this.riskScoreScale)
                .tickFormat(d => d.toFixed(1)))
            .append('text')
            .attr('transform', 'rotate(-90)')
            .attr('y', -40)
            .attr('x', -plotHeight / 2)
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .attr('font-weight', 'bold')
            .attr('fill', '#2c3e50')
            .text('Risk Score');

        // Add risk level bands
        const riskLevels = [
            { level: 'Low', min: 0, max: 3, color: '#27ae60', opacity: 0.1 },
            { level: 'Medium', min: 3, max: 6, color: '#f39c12', opacity: 0.1 },
            { level: 'High', min: 6, max: 8, color: '#e74c3c', opacity: 0.1 },
            { level: 'Critical', min: 8, max: 10, color: '#8e44ad', opacity: 0.1 }
        ];

        riskLevels.forEach(level => {
            g.append('rect')
                .attr('x', 0)
                .attr('y', this.riskScoreScale(level.max))
                .attr('width', plotWidth)
                .attr('height', this.riskScoreScale(level.min) - this.riskScoreScale(level.max))
                .attr('fill', level.color)
                .attr('opacity', level.opacity);
        });
    }

    /**
     * Add zoom behavior
     */
    addZoomBehavior(svg, g, plotWidth, plotHeight) {
        const zoom = d3.zoom()
            .scaleExtent([0.5, 10])
            .on('zoom', (event) => {
                const { transform } = event;
                
                // Update scales
                const newTimeScale = transform.rescaleX(this.timeScale);
                const newRiskScoreScale = transform.rescaleY(this.riskScoreScale);

                // Update axes
                g.select('.x-axis').call(d3.axisBottom(newTimeScale));
                g.select('.y-axis').call(d3.axisLeft(newRiskScoreScale));

                // Update lines and areas
                const newLine = d3.line()
                    .x(d => newTimeScale(new Date(d.date)))
                    .y(d => newRiskScoreScale(d.riskScore))
                    .curve(d3.curveMonotoneX);

                const newArea = d3.area()
                    .x(d => newTimeScale(new Date(d.date)))
                    .y0(d => newRiskScoreScale(d.confidenceLower))
                    .y1(d => newRiskScoreScale(d.confidenceUpper))
                    .curve(d3.curveMonotoneX);

                g.selectAll('path').attr('d', (d, i) => {
                    if (i === 0) return newArea(d);
                    return newLine(d);
                });

                // Update points
                g.selectAll('.data-point')
                    .attr('cx', d => newTimeScale(new Date(d.date)))
                    .attr('cy', d => newRiskScoreScale(d.riskScore));
            });

        svg.call(zoom);
    }

    /**
     * Create prediction accuracy chart
     */
    createPredictionAccuracyChart(containerId, accuracyData) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const width = container.clientWidth || 600;
        const height = container.clientHeight || 300;
        const margin = { top: 20, right: 20, bottom: 40, left: 60 };

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

        // Set up scales
        const xScale = d3.scaleTime()
            .domain(d3.extent(accuracyData, d => new Date(d.date)))
            .range([0, plotWidth]);

        const yScale = d3.scaleLinear()
            .domain([0, 1])
            .range([plotHeight, 0]);

        // Create line generator
        const line = d3.line()
            .x(d => xScale(new Date(d.date)))
            .y(d => yScale(d.accuracy))
            .curve(d3.curveMonotoneX);

        // Add accuracy line
        g.append('path')
            .datum(accuracyData)
            .attr('fill', 'none')
            .attr('stroke', this.options.colorScheme.actual)
            .attr('stroke-width', 3)
            .attr('d', line);

        // Add data points
        g.selectAll('.accuracy-point')
            .data(accuracyData)
            .enter().append('circle')
            .attr('class', 'accuracy-point')
            .attr('cx', d => xScale(new Date(d.date)))
            .attr('cy', d => yScale(d.accuracy))
            .attr('r', 4)
            .attr('fill', this.options.colorScheme.actual)
            .attr('stroke', '#fff')
            .attr('stroke-width', 2);

        // Add accuracy threshold line
        g.append('line')
            .attr('x1', 0)
            .attr('y1', yScale(0.8))
            .attr('x2', plotWidth)
            .attr('y2', yScale(0.8))
            .attr('stroke', '#e74c3c')
            .attr('stroke-width', 2)
            .attr('stroke-dasharray', '5,5');

        // Add axes
        g.append('g')
            .attr('transform', `translate(0, ${plotHeight})`)
            .call(d3.axisBottom(xScale)
                .tickFormat(d3.timeFormat('%m/%d')));

        g.append('g')
            .call(d3.axisLeft(yScale)
                .tickFormat(d3.format('.0%')));

        // Add labels
        g.append('text')
            .attr('x', plotWidth / 2)
            .attr('y', plotHeight + 35)
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .attr('font-weight', 'bold')
            .attr('fill', '#2c3e50')
            .text('Date');

        g.append('text')
            .attr('transform', 'rotate(-90)')
            .attr('y', -40)
            .attr('x', -plotHeight / 2)
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .attr('font-weight', 'bold')
            .attr('fill', '#2c3e50')
            .text('Prediction Accuracy');

        // Store visualization reference
        const visualization = {
            svg,
            g,
            line,
            update: (newData) => this.updatePredictionAccuracyChart(visualization, newData)
        };

        this.visualizations.set(containerId, visualization);
        return visualization;
    }

    /**
     * Create risk trend summary
     */
    createRiskTrendSummary(containerId, trendData) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        container.innerHTML = `
            <div class="risk-trend-summary">
                <div class="trend-header">
                    <h4>Risk Trend Analysis</h4>
                    <div class="time-range-selector">
                        <select id="timeRangeSelector">
                            <option value="30">Last 30 days</option>
                            <option value="90" selected>Last 90 days</option>
                            <option value="180">Last 6 months</option>
                            <option value="365">Last year</option>
                        </select>
                    </div>
                </div>

                <div class="trend-metrics">
                    <div class="metric-card">
                        <div class="metric-label">Current Score</div>
                        <div class="metric-value current">${trendData.currentScore.toFixed(1)}</div>
                        <div class="metric-change ${trendData.change >= 0 ? 'positive' : 'negative'}">
                            ${trendData.change >= 0 ? '+' : ''}${trendData.change.toFixed(1)} from last period
                        </div>
                    </div>

                    <div class="metric-card">
                        <div class="metric-label">Average Score</div>
                        <div class="metric-value">${trendData.averageScore.toFixed(1)}</div>
                        <div class="metric-subtitle">${this.timeRange} day period</div>
                    </div>

                    <div class="metric-card">
                        <div class="metric-label">Volatility</div>
                        <div class="metric-value">${trendData.volatility.toFixed(2)}</div>
                        <div class="metric-subtitle">Standard deviation</div>
                    </div>

                    <div class="metric-card">
                        <div class="metric-label">Trend Direction</div>
                        <div class="metric-value trend-${trendData.trendDirection}">
                            <i class="fas fa-arrow-${trendData.trendDirection === 'up' ? 'up' : trendData.trendDirection === 'down' ? 'down' : 'right'}"></i>
                        </div>
                        <div class="metric-subtitle">${trendData.trendStrength}</div>
                    </div>
                </div>

                <div class="trend-insights">
                    <h5>Key Insights</h5>
                    <ul>
                        ${trendData.insights.map(insight => `<li>${insight}</li>`).join('')}
                    </ul>
                </div>
            </div>
        `;

        // Bind time range selector
        const timeRangeSelector = container.querySelector('#timeRangeSelector');
        timeRangeSelector.addEventListener('change', (e) => {
            this.timeRange = parseInt(e.target.value);
            document.dispatchEvent(new CustomEvent('timeRangeChanged', {
                detail: { timeRange: this.timeRange }
            }));
        });

        this.addTrendSummaryStyles();
    }

    /**
     * Load history data for a merchant
     */
    async loadHistoryData(merchantId) {
        try {
            const endpoints = APIConfig.getEndpoints();
            const response = await fetch(endpoints.riskHistory(merchantId), {
                headers: APIConfig.getHeaders()
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const historyData = await response.json();
            this.historyData = historyData;
            return historyData;

        } catch (error) {
            console.error('Error loading history data:', error);
            // Return mock data for development
            return this.generateMockHistoryData();
        }
    }

    /**
     * Generate mock history data for development
     */
    generateMockHistoryData() {
        const data = [];
        const startDate = new Date();
        startDate.setDate(startDate.getDate() - this.timeRange);

        for (let i = 0; i < this.timeRange; i++) {
            const date = new Date(startDate);
            date.setDate(date.getDate() + i);
            
            // Generate realistic risk score with some trend
            const baseScore = 5 + Math.sin(i / 30) * 2;
            const randomFactor = (Math.random() - 0.5) * 1;
            const riskScore = Math.max(0, Math.min(10, baseScore + randomFactor));

            const dataPoint = {
                date: date.toISOString().split('T')[0],
                riskScore: riskScore,
                confidenceLower: Math.max(0, riskScore - 0.5),
                confidenceUpper: Math.min(10, riskScore + 0.5)
            };

            // Add predictions for some data points
            if (i > 7 && Math.random() > 0.3) {
                dataPoint.predictedScore = riskScore + (Math.random() - 0.5) * 0.8;
                dataPoint.accuracy = 0.7 + Math.random() * 0.3;
            }

            // Add events occasionally
            if (Math.random() > 0.9) {
                const events = [
                    'Payment spike detected',
                    'New compliance requirement',
                    'Market volatility increase',
                    'Customer complaint received',
                    'Operational efficiency improvement'
                ];
                dataPoint.events = [events[Math.floor(Math.random() * events.length)]];
            }

            data.push(dataPoint);
        }

        return data;
    }

    /**
     * Update time range
     */
    updateTimeRange() {
        if (this.currentMerchantId) {
            this.loadHistoryData(this.currentMerchantId).then(data => {
                this.visualizations.forEach(visualization => {
                    if (visualization.update) {
                        visualization.update(data);
                    }
                });
            });
        }
    }

    /**
     * Update risk history timeline
     */
    updateRiskHistoryTimeline(visualization, newData) {
        // Recreate the timeline with new data
        const containerId = visualization.svg.node().parentElement.id;
        this.visualizations.delete(containerId);
        return this.createRiskHistoryTimeline(containerId, newData);
    }

    /**
     * Update prediction accuracy chart
     */
    updatePredictionAccuracyChart(visualization, newData) {
        // Recreate the chart with new data
        const containerId = visualization.svg.node().parentElement.id;
        this.visualizations.delete(containerId);
        return this.createPredictionAccuracyChart(containerId, newData);
    }

    /**
     * Add trend summary styles
     */
    addTrendSummaryStyles() {
        if (document.getElementById('trend-summary-styles')) return;

        const style = document.createElement('style');
        style.id = 'trend-summary-styles';
        style.textContent = `
            .risk-trend-summary {
                background: #fff;
                border-radius: 10px;
                padding: 20px;
                box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            }

            .trend-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 20px;
                padding-bottom: 15px;
                border-bottom: 2px solid #ecf0f1;
            }

            .time-range-selector select {
                padding: 8px 12px;
                border: 1px solid #ddd;
                border-radius: 5px;
                background: white;
            }

            .trend-metrics {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                gap: 15px;
                margin-bottom: 20px;
            }

            .metric-card {
                background: #f8f9fa;
                padding: 15px;
                border-radius: 8px;
                text-align: center;
                border-left: 4px solid #3498db;
            }

            .metric-label {
                font-size: 0.9em;
                color: #666;
                margin-bottom: 5px;
            }

            .metric-value {
                font-size: 2em;
                font-weight: bold;
                color: #2c3e50;
                margin-bottom: 5px;
            }

            .metric-value.current {
                color: #3498db;
            }

            .metric-value.trend-up {
                color: #e74c3c;
            }

            .metric-value.trend-down {
                color: #27ae60;
            }

            .metric-value.trend-stable {
                color: #f39c12;
            }

            .metric-change {
                font-size: 0.8em;
                font-weight: bold;
            }

            .metric-change.positive {
                color: #e74c3c;
            }

            .metric-change.negative {
                color: #27ae60;
            }

            .metric-subtitle {
                font-size: 0.8em;
                color: #666;
            }

            .trend-insights {
                background: #f8f9fa;
                padding: 15px;
                border-radius: 8px;
            }

            .trend-insights h5 {
                margin-bottom: 10px;
                color: #2c3e50;
            }

            .trend-insights ul {
                list-style: none;
                padding: 0;
            }

            .trend-insights li {
                padding: 5px 0;
                border-bottom: 1px solid #eee;
            }

            .trend-insights li:last-child {
                border-bottom: none;
            }

            @media (max-width: 768px) {
                .trend-header {
                    flex-direction: column;
                    gap: 15px;
                    align-items: stretch;
                }

                .trend-metrics {
                    grid-template-columns: repeat(2, 1fr);
                }
            }
        `;

        document.head.appendChild(style);
    }

    /**
     * Export history data
     */
    exportHistoryData(format = 'csv') {
        if (!this.historyData) return;

        let content;
        let filename;
        let mimeType;

        if (format === 'csv') {
            const headers = ['Date', 'Risk Score', 'Predicted Score', 'Accuracy', 'Events'];
            const rows = this.historyData.map(d => [
                d.date,
                d.riskScore.toFixed(2),
                d.predictedScore ? d.predictedScore.toFixed(2) : '',
                d.accuracy ? (d.accuracy * 100).toFixed(1) + '%' : '',
                d.events ? d.events.join('; ') : ''
            ]);

            content = [headers, ...rows].map(row => row.join(',')).join('\n');
            filename = `risk-history-${this.currentMerchantId}-${new Date().toISOString().split('T')[0]}.csv`;
            mimeType = 'text/csv';
        } else if (format === 'json') {
            content = JSON.stringify(this.historyData, null, 2);
            filename = `risk-history-${this.currentMerchantId}-${new Date().toISOString().split('T')[0]}.json`;
            mimeType = 'application/json';
        }

        const blob = new Blob([content], { type: mimeType });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        a.click();
        window.URL.revokeObjectURL(url);
    }

    /**
     * Destroy component
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
        this.historyData = null;
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskHistoryTracking;
}
