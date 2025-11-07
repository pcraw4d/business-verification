/**
 * Predictive Risk Forecast Component
 * 
 * Displays 3, 6, and 12-month risk forecasts with confidence intervals
 * and risk drivers. This is a unique feature for the Risk Indicators tab.
 */

class PredictiveRiskForecast {
    constructor() {
        this.predictions = null;
        this.chart = null;
    }
    
    /**
     * Initialize predictive forecast
     * @param {string} merchantId - Merchant ID
     * @param {Object} predictions - Prediction data from SharedRiskDataService
     */
    async init(merchantId, predictions) {
        this.merchantId = merchantId;
        this.predictions = predictions;
        
        if (!predictions || !predictions.predictions || predictions.predictions.length === 0) {
            this.showNoDataMessage();
            return;
        }
        
        this.render();
        await this.initializeCharts();
        this.renderRiskDrivers();
    }
    
    /**
     * Render the forecast UI
     */
    render() {
        const container = document.getElementById('predictiveRiskForecast');
        if (!container) {
            console.warn('Predictive forecast container not found');
            return;
        }
        
        container.innerHTML = `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <div class="flex items-center justify-between mb-6">
                    <h2 class="text-xl font-bold text-gray-900">
                        <i class="fas fa-chart-line mr-2 text-blue-600"></i>
                        Predictive Risk Forecast
                    </h2>
                    <div class="flex items-center space-x-2">
                        <span class="text-sm text-gray-600">Forecast Horizons:</span>
                        <span class="px-2 py-1 bg-blue-100 text-blue-800 rounded text-sm font-semibold">3M</span>
                        <span class="px-2 py-1 bg-blue-100 text-blue-800 rounded text-sm font-semibold">6M</span>
                        <span class="px-2 py-1 bg-blue-100 text-blue-800 rounded text-sm font-semibold">12M</span>
                    </div>
                </div>
                
                <!-- Forecast Chart -->
                <div class="mb-6">
                    <canvas id="predictiveForecastChart" height="300"></canvas>
                </div>
                
                <!-- Risk Drivers -->
                <div id="riskDriversSection" class="mb-6">
                    <h3 class="text-lg font-semibold text-gray-900 mb-4">Key Risk Drivers</h3>
                    <div id="riskDriversList" class="space-y-3">
                        <!-- Risk drivers will be populated here -->
                    </div>
                </div>
                
                <!-- Scenarios (if available) -->
                <div id="scenariosSection" class="hidden">
                    <h3 class="text-lg font-semibold text-gray-900 mb-4">Scenario Analysis</h3>
                    <div id="scenariosList" class="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <!-- Scenarios will be populated here -->
                    </div>
                </div>
                
                <!-- Confidence Intervals Info -->
                <div class="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
                    <p class="text-sm text-blue-800">
                        <i class="fas fa-info-circle mr-2"></i>
                        Forecasts include 95% confidence intervals. Predictions are based on historical trends, 
                        industry benchmarks, and machine learning models.
                    </p>
                </div>
            </div>
        `;
    }
    
    /**
     * Initialize forecast charts
     */
    async initializeCharts() {
        if (!this.predictions || !this.predictions.predictions) {
            return;
        }
        
        const canvas = document.getElementById('predictiveForecastChart');
        if (!canvas || typeof Chart === 'undefined') {
            console.warn('Chart.js not available or canvas not found');
            return;
        }
        
        // Prepare chart data
        const chartData = this.prepareChartData();
        
        // Destroy existing chart if present
        if (this.chart) {
            this.chart.destroy();
        }
        
        // Create line chart with confidence bands
        const ctx = canvas.getContext('2d');
        this.chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: chartData.labels,
                datasets: [
                    {
                        label: 'Current Risk Score',
                        data: chartData.current,
                        borderColor: '#3b82f6',
                        backgroundColor: 'rgba(59, 130, 246, 0.1)',
                        borderWidth: 3,
                        fill: false,
                        pointRadius: 4
                    },
                    {
                        label: '3-Month Forecast',
                        data: chartData.forecast3M,
                        borderColor: '#10b981',
                        backgroundColor: 'rgba(16, 185, 129, 0.1)',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        fill: false,
                        pointRadius: 3
                    },
                    {
                        label: '6-Month Forecast',
                        data: chartData.forecast6M,
                        borderColor: '#f59e0b',
                        backgroundColor: 'rgba(245, 158, 11, 0.1)',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        fill: false,
                        pointRadius: 3
                    },
                    {
                        label: '12-Month Forecast',
                        data: chartData.forecast12M,
                        borderColor: '#ef4444',
                        backgroundColor: 'rgba(239, 68, 68, 0.1)',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        fill: false,
                        pointRadius: 3
                    },
                    {
                        label: 'Confidence Upper',
                        data: chartData.confidenceUpper,
                        borderColor: 'rgba(239, 68, 68, 0.3)',
                        backgroundColor: 'rgba(239, 68, 68, 0.1)',
                        borderWidth: 1,
                        fill: '+1',
                        pointRadius: 0
                    },
                    {
                        label: 'Confidence Lower',
                        data: chartData.confidenceLower,
                        borderColor: 'rgba(239, 68, 68, 0.3)',
                        backgroundColor: 'rgba(239, 68, 68, 0.1)',
                        borderWidth: 1,
                        fill: false,
                        pointRadius: 0
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
                        intersect: false
                    },
                    title: {
                        display: true,
                        text: 'Risk Score Forecast (0-100)'
                    }
                },
                scales: {
                    x: {
                        title: {
                            display: true,
                            text: 'Time Horizon'
                        }
                    },
                    y: {
                        beginAtZero: true,
                        max: 100,
                        title: {
                            display: true,
                            text: 'Risk Score'
                        }
                    }
                }
            }
        });
    }
    
    /**
     * Prepare chart data from predictions
     */
    prepareChartData() {
        const predictions = this.predictions.predictions || [];
        const currentScore = this.predictions.currentScore || 50;
        
        // Create labels (current, 3M, 6M, 12M)
        const labels = ['Current', '3 Months', '6 Months', '12 Months'];
        
        // Extract forecast values
        const forecast3M = predictions.find(p => p.horizon === 3);
        const forecast6M = predictions.find(p => p.horizon === 6);
        const forecast12M = predictions.find(p => p.horizon === 12);
        
        return {
            labels,
            current: [currentScore, null, null, null],
            forecast3M: [null, forecast3M?.predictedScore || null, null, null],
            forecast6M: [null, null, forecast6M?.predictedScore || null, null],
            forecast12M: [null, null, null, forecast12M?.predictedScore || null],
            confidenceUpper: [
                null,
                forecast3M?.confidenceInterval?.upper || null,
                forecast6M?.confidenceInterval?.upper || null,
                forecast12M?.confidenceInterval?.upper || null
            ],
            confidenceLower: [
                null,
                forecast3M?.confidenceInterval?.lower || null,
                forecast6M?.confidenceInterval?.lower || null,
                forecast12M?.confidenceInterval?.lower || null
            ]
        };
    }
    
    /**
     * Render risk drivers
     */
    renderRiskDrivers() {
        const container = document.getElementById('riskDriversList');
        if (!container || !this.predictions.drivers) {
            return;
        }
        
        const drivers = this.predictions.drivers.slice(0, 5); // Top 5 drivers
        
        container.innerHTML = drivers.map((driver, index) => {
            const impactColor = driver.impact > 0.7 ? 'text-red-600' : driver.impact > 0.4 ? 'text-orange-600' : 'text-yellow-600';
            const impactPercent = Math.round(driver.impact * 100);
            
            return `
                <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg border border-gray-200">
                    <div class="flex items-center flex-1">
                        <span class="text-lg font-bold text-gray-400 mr-3">#${index + 1}</span>
                        <div class="flex-1">
                            <div class="font-semibold text-gray-900">${this.formatCategoryName(driver.category)}: ${driver.factor}</div>
                            <div class="text-sm text-gray-600">${driver.description || 'Risk driver'}</div>
                        </div>
                    </div>
                    <div class="text-right">
                        <div class="text-sm font-semibold ${impactColor}">${impactPercent}% Impact</div>
                        <div class="w-24 h-2 bg-gray-200 rounded-full mt-1">
                            <div class="h-2 rounded-full ${this.getImpactColorClass(driver.impact)}" 
                                 style="width: ${impactPercent}%"></div>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    }
    
    /**
     * Get impact color class
     */
    getImpactColorClass(impact) {
        if (impact > 0.7) return 'bg-red-500';
        if (impact > 0.4) return 'bg-orange-500';
        return 'bg-yellow-500';
    }
    
    /**
     * Format category name
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
        return names[category] || category;
    }
    
    /**
     * Show no data message
     */
    showNoDataMessage() {
        const container = document.getElementById('predictiveRiskForecast');
        if (!container) return;
        
        container.innerHTML = `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <div class="text-center py-8">
                    <i class="fas fa-chart-line text-4xl text-gray-400 mb-4"></i>
                    <h3 class="text-lg font-semibold text-gray-900 mb-2">Predictive Forecasts Not Available</h3>
                    <p class="text-gray-600 mb-4">
                        Risk predictions require historical data and machine learning models. 
                        Check back once sufficient data is available.
                    </p>
                    <a href="#risk-assessment-tab" 
                       onclick="event.preventDefault(); 
                                const tab = document.querySelector('[data-tab=\\'risk-assessment\\']') || document.querySelector('[href=\\'#risk-assessment-tab\\']');
                                if (tab) tab.click();
                                return false;"
                       class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                        <i class="fas fa-arrow-right mr-2"></i>
                        View Historical Trends
                    </a>
                </div>
            </div>
        `;
    }
}

