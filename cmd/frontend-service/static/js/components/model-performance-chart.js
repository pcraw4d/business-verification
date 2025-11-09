/**
 * Model Performance Chart Component
 * Displays ML model performance visualization
 */
class ModelPerformanceChart {
    constructor(containerId, options = {}) {
        this.containerId = containerId;
        this.container = document.getElementById(containerId);
        this.options = {
            chartType: 'bar',
            ...options
        };
        this.chart = null;
    }

    /**
     * Initialize the chart
     */
    init() {
        if (!this.container) {
            console.error(`Chart container ${this.containerId} not found`);
            return;
        }

        this.createChart();
    }

    /**
     * Create the chart
     */
    createChart() {
        const canvas = document.createElement('canvas');
        this.container.innerHTML = '';
        this.container.appendChild(canvas);

        const ctx = canvas.getContext('2d');
        
        this.chart = new Chart(ctx, {
            type: this.options.chartType,
            data: {
                labels: [],
                datasets: []
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100
                    }
                }
            }
        });
    }

    /**
     * Update chart with performance data
     */
    updateChart(performanceData) {
        if (!this.chart || !performanceData) return;

        const labels = performanceData.labels || ['Accuracy', 'Precision', 'Recall', 'F1 Score'];
        const data = performanceData.data || [];

        this.chart.data.labels = labels;
        this.chart.data.datasets = [{
            label: 'Performance Metrics',
            data: data,
            backgroundColor: [
                '#4a90e2',
                '#28a745',
                '#ffc107',
                '#dc3545'
            ]
        }];

        this.chart.update();
    }

    /**
     * Destroy the chart
     */
    destroy() {
        if (this.chart) {
            this.chart.destroy();
            this.chart = null;
        }
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.ModelPerformanceChart = ModelPerformanceChart;
}

