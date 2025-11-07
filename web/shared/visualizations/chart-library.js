/**
 * Shared Chart Library
 * Provides reusable chart components using Chart.js
 * 
 * This library wraps Chart.js to provide consistent chart creation,
 * management, and styling across all pages.
 */

// Chart.js should be available globally via CDN or module import
function getChartLibrary() {
    if (typeof Chart !== 'undefined') {
        return Chart;
    }
    throw new Error('Chart.js library not loaded. Please include Chart.js before using SharedChartLibrary.');
}

class SharedChartLibrary {
    constructor() {
        this.charts = new Map();
        this.Chart = getChartLibrary();
        this.defaultConfig = {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    enabled: true,
                    mode: 'index',
                    intersect: false
                }
            }
        };
    }
    
    /**
     * Create a line chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} data - Chart data
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createLineChart(canvasId, data, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) {
            throw new Error(`Canvas element not found: ${canvasId}`);
        }
        
        // Destroy existing chart if present
        if (this.charts.has(canvasId)) {
            this.charts.get(canvasId).destroy();
        }
        
        const config = {
            type: 'line',
            data: this.normalizeChartData(data),
            options: {
                ...this.defaultConfig,
                ...options,
                scales: {
                    x: {
                        type: options.scales?.x?.type || 'category',
                        ...options.scales?.x
                    },
                    y: {
                        beginAtZero: true,
                        ...options.scales?.y
                    }
                }
            }
        };
        
        const chart = new this.Chart(canvas, config);
        this.charts.set(canvasId, chart);
        
        return chart;
    }
    
    /**
     * Create a bar chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} data - Chart data
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createBarChart(canvasId, data, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) {
            throw new Error(`Canvas element not found: ${canvasId}`);
        }
        
        if (this.charts.has(canvasId)) {
            this.charts.get(canvasId).destroy();
        }
        
        const config = {
            type: 'bar',
            data: this.normalizeChartData(data),
            options: {
                ...this.defaultConfig,
                ...options,
                scales: {
                    x: {
                        ...options.scales?.x
                    },
                    y: {
                        beginAtZero: true,
                        ...options.scales?.y
                    }
                }
            }
        };
        
        const chart = new this.Chart(canvas, config);
        this.charts.set(canvasId, chart);
        
        return chart;
    }
    
    /**
     * Create a radar chart
     * @param {string} canvasId - Canvas element ID
     * @param {Object} data - Chart data
     * @param {Object} options - Chart options
     * @returns {Chart} Chart instance
     */
    createRadarChart(canvasId, data, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) {
            throw new Error(`Canvas element not found: ${canvasId}`);
        }
        
        if (this.charts.has(canvasId)) {
            this.charts.get(canvasId).destroy();
        }
        
        const config = {
            type: 'radar',
            data: this.normalizeChartData(data),
            options: {
                ...this.defaultConfig,
                ...options,
                scales: {
                    r: {
                        beginAtZero: true,
                        max: options.max || 100,
                        ...options.scales?.r
                    }
                }
            }
        };
        
        const chart = new this.Chart(canvas, config);
        this.charts.set(canvasId, chart);
        
        return chart;
    }
    
    /**
     * Create a gauge chart (semi-circle gauge)
     * @param {string} canvasId - Canvas element ID
     * @param {number} value - Gauge value (0-100)
     * @param {Object} options - Gauge options
     * @returns {Object} Chart reference object
     */
    createGaugeChart(canvasId, value, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) {
            throw new Error(`Canvas element not found: ${canvasId}`);
        }
        
        if (this.charts.has(canvasId)) {
            // Clean up existing gauge if it exists
            const existing = this.charts.get(canvasId);
            if (existing.destroy) {
                existing.destroy();
            }
        }
        
        // Gauge chart implementation using canvas drawing
        const ctx = canvas.getContext('2d');
        const centerX = canvas.width / 2;
        const centerY = canvas.height / 2;
        const radius = Math.min(canvas.width, canvas.height) / 2 - 20;
        
        // Draw gauge arc
        this.drawGaugeArc(ctx, centerX, centerY, radius, value, options);
        
        // Store reference
        const chart = { value, options, canvas, ctx, update: (newValue) => {
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            this.drawGaugeArc(ctx, centerX, centerY, radius, newValue, options);
            chart.value = newValue;
        }};
        this.charts.set(canvasId, chart);
        
        return chart;
    }
    
    /**
     * Update chart data
     * @param {string} canvasId - Canvas element ID
     * @param {Object} newData - New chart data
     */
    updateChart(canvasId, newData) {
        const chart = this.charts.get(canvasId);
        if (!chart) {
            throw new Error(`Chart not found: ${canvasId}`);
        }
        
        if (chart.data) {
            // Standard Chart.js chart
            chart.data = this.normalizeChartData(newData);
            chart.update();
        } else if (chart.update) {
            // Custom chart (like gauge)
            chart.update(newData);
        } else {
            // Recreate chart for non-standard charts
            this.recreateChart(canvasId, newData);
        }
    }
    
    /**
     * Destroy chart
     * @param {string} canvasId - Canvas element ID
     */
    destroyChart(canvasId) {
        const chart = this.charts.get(canvasId);
        if (chart) {
            if (chart.destroy && typeof chart.destroy === 'function') {
                chart.destroy();
            } else if (chart.ctx && chart.canvas) {
                // Clear canvas for custom charts
                const ctx = chart.ctx;
                ctx.clearRect(0, 0, chart.canvas.width, chart.canvas.height);
            }
        }
        this.charts.delete(canvasId);
    }
    
    /**
     * Normalize chart data to Chart.js format
     * @param {Object} data - Raw chart data
     * @returns {Object} Normalized chart data
     */
    normalizeChartData(data) {
        // If already in Chart.js format (has labels and datasets)
        if (data.labels && data.datasets) {
            return data;
        }
        
        // Convert from array format to Chart.js format
        if (Array.isArray(data)) {
            return {
                labels: data.map(item => item.label || item.x || item.name),
                datasets: [{
                    label: 'Data',
                    data: data.map(item => item.value || item.y || item.score)
                }]
            };
        }
        
        // Return as-is if already normalized or in unknown format
        return data;
    }
    
    /**
     * Draw gauge arc
     * @param {CanvasRenderingContext2D} ctx - Canvas context
     * @param {number} centerX - Center X coordinate
     * @param {number} centerY - Center Y coordinate
     * @param {number} radius - Gauge radius
     * @param {number} value - Gauge value (0-100)
     * @param {Object} options - Gauge options
     */
    drawGaugeArc(ctx, centerX, centerY, radius, value, options = {}) {
        const {
            min = 0,
            max = 100,
            color = this.getRiskColor(value),
            backgroundColor = '#e5e7eb',
            lineWidth = 20
        } = options;
        
        const normalizedValue = Math.max(0, Math.min(100, ((value - min) / (max - min)) * 100));
        const startAngle = Math.PI;
        const endAngle = startAngle + (normalizedValue / 100) * Math.PI;
        const fullEndAngle = startAngle + Math.PI;
        
        // Draw background arc
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, startAngle, fullEndAngle);
        ctx.lineWidth = lineWidth;
        ctx.strokeStyle = backgroundColor;
        ctx.lineCap = 'round';
        ctx.stroke();
        
        // Draw value arc
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, startAngle, endAngle);
        ctx.lineWidth = lineWidth;
        ctx.strokeStyle = color;
        ctx.lineCap = 'round';
        ctx.stroke();
        
        // Draw center text if provided
        if (options.showValue !== false) {
            ctx.fillStyle = '#1a202c';
            ctx.font = options.fontSize || '24px Arial';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(Math.round(value).toString(), centerX, centerY - 10);
            
            if (options.label) {
                ctx.font = options.labelFontSize || '12px Arial';
                ctx.fillStyle = '#6b7280';
                ctx.fillText(options.label, centerX, centerY + 15);
            }
        }
    }
    
    /**
     * Get risk color based on value
     * @param {number} value - Risk value (0-100)
     * @returns {string} Color hex code
     */
    getRiskColor(value) {
        if (value <= 25) return '#10b981'; // Green - Low
        if (value <= 50) return '#eab308'; // Yellow - Medium
        if (value <= 75) return '#f97316'; // Orange - High
        return '#ef4444'; // Red - Critical
    }
    
    /**
     * Recreate chart (for non-standard charts that can't be updated)
     * @param {string} canvasId - Canvas element ID
     * @param {Object} newData - New chart data
     */
    recreateChart(canvasId, newData) {
        // This would need to know the chart type to recreate
        // For now, just log a warning
        console.warn(`Cannot update chart ${canvasId}, please destroy and recreate manually`);
    }
    
    /**
     * Get all active charts
     * @returns {Map} Map of all active charts
     */
    getAllCharts() {
        return this.charts;
    }
    
    /**
     * Destroy all charts
     */
    destroyAllCharts() {
        for (const [canvasId, chart] of this.charts.entries()) {
            this.destroyChart(canvasId);
        }
    }
}

// Export singleton instance
let chartLibraryInstance = null;

export function getChartLibrary() {
    if (!chartLibraryInstance) {
        chartLibraryInstance = new SharedChartLibrary();
    }
    return chartLibraryInstance;
}

export { SharedChartLibrary };

// Make available globally for non-module environments
if (typeof window !== 'undefined') {
    window.getChartLibrary = getChartLibrary;
    window.SharedChartLibrary = SharedChartLibrary;
}

