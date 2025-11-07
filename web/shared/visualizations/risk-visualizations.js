/**
 * Risk-Specific Visualizations
 * Specialized charts for risk data
 */

class RiskVisualizations {
    constructor() {
        // Chart library will be loaded lazily
        this.chartLibrary = null;
    }
    
    /**
     * Get chart library instance (lazy load)
     */
    getChartLibrary() {
        if (this.chartLibrary) {
            return this.chartLibrary;
        }
        // Try to get from global scope first (for non-module environments)
        if (typeof window !== 'undefined' && window.getChartLibrary) {
            this.chartLibrary = window.getChartLibrary();
            return this.chartLibrary;
        }
        // Try to import if in module environment
        if (typeof import !== 'undefined') {
            import('./chart-library.js').then(chartLib => {
                this.chartLibrary = chartLib.getChartLibrary();
            }).catch(() => {
                console.warn('Chart library not available');
            });
        }
        if (!this.chartLibrary) {
            throw new Error('Chart library not available. Please load chart-library.js first.');
        }
        return this.chartLibrary;
    }
    
    /**
     * Create risk trend chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} riskHistory - Risk history data
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createRiskTrendChart(canvasId, riskHistory, options = {}) {
        if (!riskHistory || !riskHistory.dataPoints || riskHistory.dataPoints.length === 0) {
            console.warn('No risk history data available for trend chart');
            return null;
        }
        
        const data = {
            labels: riskHistory.dataPoints.map(dp => {
                const date = new Date(dp.timestamp);
                return date.toLocaleDateString();
            }),
            datasets: [{
                label: 'Overall Risk Score',
                data: riskHistory.dataPoints.map(dp => dp.overallScore),
                borderColor: '#3b82f6',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                fill: true,
                tension: 0.4,
                borderWidth: 2
            }]
        };
        
        // Add industry average if available
        if (riskHistory.dataPoints[0].industryAverage !== undefined) {
            data.datasets.push({
                label: 'Industry Average',
                data: riskHistory.dataPoints.map(dp => dp.industryAverage || 0),
                borderColor: '#9ca3af',
                backgroundColor: 'rgba(156, 163, 175, 0.1)',
                borderDash: [5, 5],
                fill: false,
                borderWidth: 2
            });
        }
        
        return this.getChartLibrary().createLineChart(canvasId, data, {
            ...options,
            plugins: {
                ...options.plugins,
                title: {
                    display: true,
                    text: options.title || 'Risk Trend Over Time'
                }
            }
        });
    }
    
    /**
     * Create risk radar chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} categoryScores - Risk category scores
     * @param {Object} benchmarks - Industry benchmarks (optional)
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createRiskRadarChart(canvasId, categoryScores, benchmarks = null, options = {}) {
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        const labels = categoryOrder.map(cat => this.formatCategoryName(cat));
        
        const datasets = [{
            label: 'Current Risk Level',
            data: categoryOrder.map(cat => {
                const score = categoryScores[cat]?.score || categoryScores[cat] || 0;
                return score;
            }),
            backgroundColor: 'rgba(59, 130, 246, 0.2)',
            borderColor: 'rgba(59, 130, 246, 1)',
            borderWidth: 2,
            pointBackgroundColor: 'rgba(59, 130, 246, 1)',
            pointBorderColor: '#fff',
            pointHoverBackgroundColor: '#fff',
            pointHoverBorderColor: 'rgba(59, 130, 246, 1)'
        }];
        
        if (benchmarks && benchmarks.averages) {
            datasets.push({
                label: 'Industry Average',
                data: categoryOrder.map(cat => benchmarks.averages[cat] || 0),
                backgroundColor: 'rgba(156, 163, 175, 0.2)',
                borderColor: 'rgba(156, 163, 175, 1)',
                borderWidth: 2,
                borderDash: [5, 5],
                pointBackgroundColor: 'rgba(156, 163, 175, 1)',
                pointBorderColor: '#fff',
                pointHoverBackgroundColor: '#fff',
                pointHoverBorderColor: 'rgba(156, 163, 175, 1)'
            });
        }
        
        const data = {
            labels,
            datasets
        };
        
        return this.getChartLibrary().createRadarChart(canvasId, data, {
            ...options,
            scales: {
                r: {
                    beginAtZero: true,
                    max: options.max || 100,
                    ...options.scales?.r
                }
            }
        });
    }
    
    /**
     * Create risk category bar chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} categoryScores - Risk category scores
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createRiskCategoryBarChart(canvasId, categoryScores, options = {}) {
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        const labels = categoryOrder.map(cat => this.formatCategoryName(cat));
        const data = categoryOrder.map(cat => {
            const score = categoryScores[cat]?.score || categoryScores[cat] || 0;
            return score;
        });
        const colors = data.map(score => this.getRiskColor(score));
        
        const chartData = {
            labels,
            datasets: [{
                label: 'Risk Score',
                data,
                backgroundColor: colors,
                borderColor: colors.map(c => this.darkenColor(c)),
                borderWidth: 1
            }]
        };
        
        return this.getChartLibrary().createBarChart(canvasId, chartData, {
            ...options,
            indexAxis: options.indexAxis || 'y', // Horizontal bars by default
            plugins: {
                ...options.plugins,
                legend: {
                    display: false,
                    ...options.plugins?.legend
                }
            }
        });
    }
    
    /**
     * Format category name
     * @param {string} category - Category key
     * @returns {string} Formatted category name
     */
    formatCategoryName(category) {
        const names = {
            financial: 'Financial',
            operational: 'Operational',
            regulatory: 'Regulatory',
            reputational: 'Reputational',
            cybersecurity: 'Cybersecurity',
            content: 'Content'
        };
        return names[category] || category.charAt(0).toUpperCase() + category.slice(1);
    }
    
    /**
     * Get risk color
     * @param {number} score - Risk score (0-100)
     * @returns {string} Color hex code
     */
    getRiskColor(score) {
        if (score <= 25) return '#10b981'; // Green
        if (score <= 50) return '#eab308'; // Yellow
        if (score <= 75) return '#f97316'; // Orange
        return '#ef4444'; // Red
    }
    
    /**
     * Darken color (simple implementation)
     * @param {string} color - Color hex code
     * @returns {string} Darkened color hex code
     */
    darkenColor(color) {
        // Simple darkening - convert hex to RGB, reduce brightness
        const hex = color.replace('#', '');
        const r = parseInt(hex.substr(0, 2), 16);
        const g = parseInt(hex.substr(2, 2), 16);
        const b = parseInt(hex.substr(4, 2), 16);
        
        const darkenedR = Math.max(0, r - 20).toString(16).padStart(2, '0');
        const darkenedG = Math.max(0, g - 20).toString(16).padStart(2, '0');
        const darkenedB = Math.max(0, b - 20).toString(16).padStart(2, '0');
        
        return `#${darkenedR}${darkenedG}${darkenedB}`;
    }
}

export function getRiskVisualizations() {
    return new RiskVisualizations();
}

export { RiskVisualizations };

// Make available globally for non-module environments
if (typeof window !== 'undefined') {
    window.getRiskVisualizations = getRiskVisualizations;
    window.RiskVisualizations = RiskVisualizations;
}

