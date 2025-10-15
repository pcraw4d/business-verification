/**
 * Risk Visualization Component
 * 
 * Provides advanced risk visualization capabilities using D3.js and Chart.js:
 * - Real-time risk gauge (D3.js radial gauge)
 * - Risk trend line chart with prediction bands
 * - Risk category radar chart
 * - Risk distribution sunburst chart (D3.js)
 * - Force-directed graph for risk factor relationships
 * - Network graph for inter-merchant risk correlations
 */

class RiskVisualization {
    constructor(options = {}) {
        this.options = {
            animationDuration: 1000,
            colorScheme: {
                low: '#27ae60',
                medium: '#f39c12',
                high: '#e74c3c',
                critical: '#8e44ad'
            },
            ...options
        };

        this.charts = new Map();
        this.d3Visualizations = new Map();
        this.animationQueue = [];
        this.isAnimating = false;

        this.init();
    }

    /**
     * Initialize the visualization component
     */
    init() {
        this.setupEventListeners();
        this.initializeColorScales();
    }

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Listen for window resize events
        window.addEventListener('resize', () => {
            this.debounce(() => this.resizeAllCharts(), 250);
        });

        // Listen for theme changes
        document.addEventListener('themeChange', (event) => {
            this.updateColorScheme(event.detail.theme);
        });
    }

    /**
     * Initialize D3 color scales
     */
    initializeColorScales() {
        this.colorScale = d3.scaleOrdinal()
            .domain(['low', 'medium', 'high', 'critical'])
            .range([this.options.colorScheme.low, this.options.colorScheme.medium, 
                   this.options.colorScheme.high, this.options.colorScheme.critical]);

        this.riskScoreScale = d3.scaleLinear()
            .domain([0, 10])
            .range([0, 1]);
    }

    /**
     * Create real-time risk gauge (D3.js)
     */
    createRiskGauge(containerId, initialValue = 0) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const width = container.clientWidth || 300;
        const height = container.clientHeight || 300;
        const radius = Math.min(width, height) / 2 - 20;

        // Clear existing content
        container.innerHTML = '';

        const svg = d3.select(container)
            .append('svg')
            .attr('width', width)
            .attr('height', height);

        const g = svg.append('g')
            .attr('transform', `translate(${width / 2}, ${height / 2})`);

        // Create arc generator
        const arc = d3.arc()
            .innerRadius(radius * 0.6)
            .outerRadius(radius)
            .startAngle(0)
            .endAngle(d => d.endAngle);

        // Create background arc
        const backgroundArc = d3.arc()
            .innerRadius(radius * 0.6)
            .outerRadius(radius)
            .startAngle(0)
            .endAngle(2 * Math.PI);

        // Add background
        g.append('path')
            .datum({ endAngle: 2 * Math.PI })
            .attr('d', backgroundArc)
            .attr('fill', '#f0f0f0')
            .attr('stroke', '#ddd')
            .attr('stroke-width', 2);

        // Create risk level arcs
        const riskLevels = [
            { level: 'low', start: 0, end: Math.PI * 0.5, color: this.options.colorScheme.low },
            { level: 'medium', start: Math.PI * 0.5, end: Math.PI * 0.75, color: this.options.colorScheme.medium },
            { level: 'high', start: Math.PI * 0.75, end: Math.PI * 0.9, color: this.options.colorScheme.high },
            { level: 'critical', start: Math.PI * 0.9, end: Math.PI, color: this.options.colorScheme.critical }
        ];

        riskLevels.forEach(level => {
            g.append('path')
                .datum({ endAngle: level.end })
                .attr('d', arc)
                .attr('fill', level.color)
                .attr('opacity', 0.3)
                .attr('transform', `rotate(${level.start * 180 / Math.PI - 90})`);
        });

        // Create value arc
        const valueArc = g.append('path')
            .attr('fill', 'none')
            .attr('stroke', '#3498db')
            .attr('stroke-width', 8)
            .attr('stroke-linecap', 'round')
            .attr('opacity', 0);

        // Create needle
        const needle = g.append('g')
            .attr('class', 'needle');

        needle.append('line')
            .attr('x1', 0)
            .attr('y1', 0)
            .attr('x2', 0)
            .attr('y2', -radius * 0.8)
            .attr('stroke', '#2c3e50')
            .attr('stroke-width', 3)
            .attr('stroke-linecap', 'round');

        needle.append('circle')
            .attr('cx', 0)
            .attr('cy', 0)
            .attr('r', 8)
            .attr('fill', '#2c3e50');

        // Add center text
        const centerText = g.append('g')
            .attr('class', 'center-text');

        centerText.append('text')
            .attr('text-anchor', 'middle')
            .attr('dy', '-0.5em')
            .attr('font-size', '2em')
            .attr('font-weight', 'bold')
            .attr('fill', '#2c3e50')
            .text('0.0');

        centerText.append('text')
            .attr('text-anchor', 'middle')
            .attr('dy', '1.5em')
            .attr('font-size', '0.8em')
            .attr('fill', '#7f8c8d')
            .text('Risk Score');

        // Store chart reference
        const chart = {
            svg,
            g,
            valueArc,
            needle,
            centerText,
            radius,
            update: (value) => this.updateRiskGauge(chart, value)
        };

        this.d3Visualizations.set(containerId, chart);

        // Initial animation
        this.animateRiskGauge(chart, initialValue);

        return chart;
    }

    /**
     * Update risk gauge value
     */
    updateRiskGauge(chart, value) {
        const angle = (value / 10) * Math.PI - Math.PI / 2;
        
        // Update needle rotation
        chart.needle
            .transition()
            .duration(this.options.animationDuration)
            .attr('transform', `rotate(${angle * 180 / Math.PI})`);

        // Update center text
        chart.centerText.select('text')
            .transition()
            .duration(this.options.animationDuration)
            .tween('text', function() {
                const current = parseFloat(this.textContent) || 0;
                const target = value;
                const interpolate = d3.interpolate(current, target);
                return function(t) {
                    this.textContent = interpolate(t).toFixed(1);
                };
            });

        // Update value arc
        const arc = d3.arc()
            .innerRadius(chart.radius * 0.6)
            .outerRadius(chart.radius)
            .startAngle(-Math.PI / 2)
            .endAngle(angle);

        chart.valueArc
            .datum({ endAngle: angle })
            .transition()
            .duration(this.options.animationDuration)
            .attr('d', arc)
            .attr('opacity', 1);
    }

    /**
     * Animate risk gauge on initial load
     */
    animateRiskGauge(chart, value) {
        chart.valueArc
            .datum({ endAngle: -Math.PI / 2 })
            .attr('d', d3.arc()
                .innerRadius(chart.radius * 0.6)
                .outerRadius(chart.radius)
                .startAngle(-Math.PI / 2)
                .endAngle(-Math.PI / 2))
            .attr('opacity', 0);

        setTimeout(() => {
            this.updateRiskGauge(chart, value);
        }, 100);
    }

    /**
     * Create risk trend chart with prediction bands (Chart.js)
     */
    createRiskTrendChart(containerId, data) {
        const ctx = document.getElementById(containerId);
        if (!ctx) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: data.labels || [],
                datasets: [
                    {
                        label: 'Historical Risk Score',
                        data: data.historical || [],
                        borderColor: '#3498db',
                        backgroundColor: 'rgba(52, 152, 219, 0.1)',
                        borderWidth: 3,
                        fill: false,
                        tension: 0.4
                    },
                    {
                        label: 'Prediction',
                        data: data.prediction || [],
                        borderColor: '#e74c3c',
                        backgroundColor: 'rgba(231, 76, 60, 0.1)',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        fill: false,
                        tension: 0.4
                    },
                    {
                        label: 'Confidence Band (Upper)',
                        data: data.confidenceUpper || [],
                        borderColor: 'rgba(231, 76, 60, 0.3)',
                        backgroundColor: 'rgba(231, 76, 60, 0.1)',
                        borderWidth: 1,
                        fill: '+1',
                        pointRadius: 0,
                        tension: 0.4
                    },
                    {
                        label: 'Confidence Band (Lower)',
                        data: data.confidenceLower || [],
                        borderColor: 'rgba(231, 76, 60, 0.3)',
                        backgroundColor: 'rgba(231, 76, 60, 0.1)',
                        borderWidth: 1,
                        fill: false,
                        pointRadius: 0,
                        tension: 0.4
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false,
                        callbacks: {
                            label: function(context) {
                                let label = context.dataset.label || '';
                                if (label) {
                                    label += ': ';
                                }
                                label += context.parsed.y.toFixed(2);
                                return label;
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        display: true,
                        title: {
                            display: true,
                            text: 'Time'
                        },
                        grid: {
                            display: false
                        }
                    },
                    y: {
                        display: true,
                        title: {
                            display: true,
                            text: 'Risk Score'
                        },
                        min: 0,
                        max: 10,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    }
                },
                interaction: {
                    mode: 'nearest',
                    axis: 'x',
                    intersect: false
                }
            }
        });

        this.charts.set(containerId, chart);
        return chart;
    }

    /**
     * Create risk category radar chart (Chart.js)
     */
    createRiskRadarChart(containerId, data) {
        const ctx = document.getElementById(containerId);
        if (!ctx) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const chart = new Chart(ctx, {
            type: 'radar',
            data: {
                labels: data.labels || [],
                datasets: [
                    {
                        label: 'Current Risk',
                        data: data.current || [],
                        borderColor: '#3498db',
                        backgroundColor: 'rgba(52, 152, 219, 0.2)',
                        borderWidth: 3,
                        pointBackgroundColor: '#3498db',
                        pointBorderColor: '#fff',
                        pointHoverBackgroundColor: '#fff',
                        pointHoverBorderColor: '#3498db'
                    },
                    {
                        label: 'Industry Average',
                        data: data.industryAverage || [],
                        borderColor: '#95a5a6',
                        backgroundColor: 'rgba(149, 165, 166, 0.1)',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        pointBackgroundColor: '#95a5a6',
                        pointBorderColor: '#fff',
                        pointHoverBackgroundColor: '#fff',
                        pointHoverBorderColor: '#95a5a6'
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'top'
                    }
                },
                scales: {
                    r: {
                        beginAtZero: true,
                        max: 10,
                        ticks: {
                            stepSize: 2
                        },
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        },
                        pointLabels: {
                            font: {
                                size: 12
                            }
                        }
                    }
                }
            }
        });

        this.charts.set(containerId, chart);
        return chart;
    }

    /**
     * Create risk distribution sunburst chart (D3.js)
     */
    createRiskSunburstChart(containerId, data) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const width = container.clientWidth || 400;
        const height = container.clientHeight || 400;
        const radius = Math.min(width, height) / 2;

        // Clear existing content
        container.innerHTML = '';

        const svg = d3.select(container)
            .append('svg')
            .attr('width', width)
            .attr('height', height);

        const g = svg.append('g')
            .attr('transform', `translate(${width / 2}, ${height / 2})`);

        // Create partition layout
        const partition = d3.partition()
            .size([2 * Math.PI, radius]);

        const root = d3.hierarchy(data)
            .sum(d => d.value);

        partition(root);

        // Create arcs
        const arc = d3.arc()
            .startAngle(d => d.x0)
            .endAngle(d => d.x1)
            .innerRadius(d => d.y0)
            .outerRadius(d => d.y1);

        // Add arcs
        g.selectAll('path')
            .data(root.descendants())
            .enter().append('path')
            .attr('d', arc)
            .attr('fill', d => {
                if (d.depth === 0) return '#f0f0f0';
                if (d.depth === 1) return this.colorScale(d.data.name);
                return d3.color(this.colorScale(d.parent.data.name)).brighter(0.5);
            })
            .attr('stroke', '#fff')
            .attr('stroke-width', 2)
            .on('mouseover', function(event, d) {
                // Show tooltip
                const tooltip = d3.select('body').append('div')
                    .attr('class', 'sunburst-tooltip')
                    .style('position', 'absolute')
                    .style('background', 'rgba(0, 0, 0, 0.8)')
                    .style('color', 'white')
                    .style('padding', '8px')
                    .style('border-radius', '4px')
                    .style('font-size', '12px')
                    .style('pointer-events', 'none')
                    .style('z-index', '1000');

                tooltip.html(`
                    <strong>${d.data.name}</strong><br/>
                    Value: ${d.value}<br/>
                    Percentage: ${((d.value / root.value) * 100).toFixed(1)}%
                `);

                d3.select(this)
                    .attr('opacity', 0.8)
                    .attr('stroke-width', 4);
            })
            .on('mousemove', function(event) {
                d3.select('.sunburst-tooltip')
                    .style('left', (event.pageX + 10) + 'px')
                    .style('top', (event.pageY - 10) + 'px');
            })
            .on('mouseout', function() {
                d3.select('.sunburst-tooltip').remove();
                d3.select(this)
                    .attr('opacity', 1)
                    .attr('stroke-width', 2);
            });

        // Store chart reference
        const chart = {
            svg,
            g,
            root,
            update: (newData) => this.updateRiskSunburstChart(chart, newData)
        };

        this.d3Visualizations.set(containerId, chart);
        return chart;
    }

    /**
     * Create force-directed graph for risk factor relationships (D3.js)
     */
    createRiskFactorNetwork(containerId, data) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container ${containerId} not found`);
            return null;
        }

        const width = container.clientWidth || 600;
        const height = container.clientHeight || 400;

        // Clear existing content
        container.innerHTML = '';

        const svg = d3.select(container)
            .append('svg')
            .attr('width', width)
            .attr('height', height);

        // Create force simulation
        const simulation = d3.forceSimulation(data.nodes)
            .force('link', d3.forceLink(data.links).id(d => d.id).distance(100))
            .force('charge', d3.forceManyBody().strength(-300))
            .force('center', d3.forceCenter(width / 2, height / 2));

        // Create links
        const link = svg.append('g')
            .selectAll('line')
            .data(data.links)
            .enter().append('line')
            .attr('stroke', '#999')
            .attr('stroke-opacity', 0.6)
            .attr('stroke-width', d => Math.sqrt(d.strength * 3));

        // Create nodes
        const node = svg.append('g')
            .selectAll('circle')
            .data(data.nodes)
            .enter().append('circle')
            .attr('r', d => Math.sqrt(d.importance) * 5 + 8)
            .attr('fill', d => this.colorScale(d.riskLevel))
            .attr('stroke', '#fff')
            .attr('stroke-width', 2)
            .call(d3.drag()
                .on('start', dragstarted)
                .on('drag', dragged)
                .on('end', dragended));

        // Add labels
        const label = svg.append('g')
            .selectAll('text')
            .data(data.nodes)
            .enter().append('text')
            .text(d => d.name)
            .attr('font-size', '10px')
            .attr('text-anchor', 'middle')
            .attr('dy', '0.35em');

        // Update positions on tick
        simulation.on('tick', () => {
            link
                .attr('x1', d => d.source.x)
                .attr('y1', d => d.source.y)
                .attr('x2', d => d.target.x)
                .attr('y2', d => d.target.y);

            node
                .attr('cx', d => d.x)
                .attr('cy', d => d.y);

            label
                .attr('x', d => d.x)
                .attr('y', d => d.y);
        });

        // Drag functions
        function dragstarted(event, d) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
        }

        function dragged(event, d) {
            d.fx = event.x;
            d.fy = event.y;
        }

        function dragended(event, d) {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
        }

        // Store chart reference
        const chart = {
            svg,
            simulation,
            node,
            link,
            label,
            update: (newData) => this.updateRiskFactorNetwork(chart, newData)
        };

        this.d3Visualizations.set(containerId, chart);
        return chart;
    }

    /**
     * Update risk trend chart
     */
    updateRiskTrendChart(containerId, newData) {
        const chart = this.charts.get(containerId);
        if (!chart) return;

        chart.data.labels = newData.labels;
        chart.data.datasets[0].data = newData.historical;
        chart.data.datasets[1].data = newData.prediction;
        chart.data.datasets[2].data = newData.confidenceUpper;
        chart.data.datasets[3].data = newData.confidenceLower;

        chart.update('active');
    }

    /**
     * Update risk radar chart
     */
    updateRiskRadarChart(containerId, newData) {
        const chart = this.charts.get(containerId);
        if (!chart) return;

        chart.data.labels = newData.labels;
        chart.data.datasets[0].data = newData.current;
        chart.data.datasets[1].data = newData.industryAverage;

        chart.update('active');
    }

    /**
     * Update sunburst chart
     */
    updateRiskSunburstChart(chart, newData) {
        // Recreate the chart with new data
        const containerId = chart.svg.node().parentElement.id;
        this.d3Visualizations.delete(containerId);
        return this.createRiskSunburstChart(containerId, newData);
    }

    /**
     * Update network chart
     */
    updateRiskFactorNetwork(chart, newData) {
        // Update simulation with new data
        chart.simulation.nodes(newData.nodes);
        chart.simulation.force('link').links(newData.links);
        chart.simulation.alpha(0.3).restart();

        // Update visual elements
        chart.node.data(newData.nodes);
        chart.link.data(newData.links);
        chart.label.data(newData.nodes);
    }

    /**
     * Resize all charts
     */
    resizeAllCharts() {
        // Resize Chart.js charts
        this.charts.forEach(chart => {
            if (chart.resize) {
                chart.resize();
            }
        });

        // Resize D3.js visualizations
        this.d3Visualizations.forEach((chart, containerId) => {
            const container = document.getElementById(containerId);
            if (container) {
                const width = container.clientWidth;
                const height = container.clientHeight;
                
                if (chart.svg) {
                    chart.svg
                        .attr('width', width)
                        .attr('height', height);
                }
            }
        });
    }

    /**
     * Update color scheme
     */
    updateColorScheme(theme) {
        // Update color scale
        this.initializeColorScales();

        // Update existing charts
        this.charts.forEach(chart => {
            if (chart.options && chart.options.plugins && chart.options.plugins.legend) {
                chart.update('none');
            }
        });

        this.d3Visualizations.forEach(chart => {
            if (chart.update) {
                // Trigger a redraw with current data
                chart.update();
            }
        });
    }

    /**
     * Debounce function
     */
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    /**
     * Destroy all charts
     */
    destroy() {
        // Destroy Chart.js charts
        this.charts.forEach(chart => {
            if (chart.destroy) {
                chart.destroy();
            }
        });
        this.charts.clear();

        // Clear D3.js visualizations
        this.d3Visualizations.forEach(chart => {
            if (chart.svg) {
                chart.svg.remove();
            }
        });
        this.d3Visualizations.clear();
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskVisualization;
}
