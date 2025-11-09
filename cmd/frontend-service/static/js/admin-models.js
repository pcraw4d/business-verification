/**
 * Admin Models Management Controller
 * Manages ML model information and performance
 */
class AdminModels {
    constructor() {
        this.modelInfo = null;
        this.modelPerformance = null;
        this.ensembleInfo = null;
        this.performanceChart = null;
    }

    /**
     * Initialize the models dashboard
     */
    async init() {
        document.getElementById('loading').style.display = 'block';
        document.getElementById('modelsDashboard').style.display = 'none';

        try {
            await this.loadModelInfo();
            await this.loadModelPerformance();
            await this.loadEnsembleInfo();
            this.render();
        } catch (error) {
            console.error('Error initializing models dashboard:', error);
            this.showError('Failed to load model information: ' + error.message);
        } finally {
            document.getElementById('loading').style.display = 'none';
            document.getElementById('modelsDashboard').style.display = 'block';
        }
    }

    /**
     * Load model information
     */
    async loadModelInfo() {
        try {
            const response = await fetch('/api/v1/models/info', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.modelInfo = await response.json();
        } catch (error) {
            console.error('Error loading model info:', error);
            this.modelInfo = null;
        }
    }

    /**
     * Load model performance metrics
     */
    async loadModelPerformance() {
        try {
            const response = await fetch('/api/v1/models/performance', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.modelPerformance = await response.json();
        } catch (error) {
            console.error('Error loading model performance:', error);
            this.modelPerformance = null;
        }
    }

    /**
     * Load ensemble information
     */
    async loadEnsembleInfo() {
        try {
            const response = await fetch('/api/v1/ensemble', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.ensembleInfo = await response.json();
        } catch (error) {
            console.error('Error loading ensemble info:', error);
            this.ensembleInfo = null;
        }
    }

    /**
     * Render the models dashboard
     */
    render() {
        this.renderModelInfo();
        this.renderPerformanceMetrics();
        this.updateSummary();
    }

    /**
     * Render model information
     */
    renderModelInfo() {
        const container = document.getElementById('modelInfo');
        if (!container) return;

        if (!this.modelInfo) {
            container.innerHTML = '<p style="color: #666;">No model information available</p>';
            return;
        }

        const models = this.modelInfo.models || [this.modelInfo];
        
        container.innerHTML = models.map(model => `
            <div style="padding: 1rem; background: #f8f9fa; border-radius: 4px; margin-bottom: 1rem;">
                <h3 style="margin-bottom: 0.75rem;">${model.name || 'Model'}</h3>
                <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 0.75rem;">
                    <div>
                        <div style="font-size: 0.85rem; color: #666;">Version</div>
                        <div style="font-weight: 600;">${model.version || 'N/A'}</div>
                    </div>
                    <div>
                        <div style="font-size: 0.85rem; color: #666;">Type</div>
                        <div style="font-weight: 600;">${model.type || 'N/A'}</div>
                    </div>
                    <div>
                        <div style="font-size: 0.85rem; color: #666;">Status</div>
                        <div style="font-weight: 600; color: ${model.status === 'active' ? '#28a745' : '#dc3545'};">${model.status || 'N/A'}</div>
                    </div>
                    <div>
                        <div style="font-size: 0.85rem; color: #666;">Last Updated</div>
                        <div style="font-weight: 600;">${this.formatDate(model.updated_at)}</div>
                    </div>
                </div>
            </div>
        `).join('');
    }

    /**
     * Render performance metrics
     */
    renderPerformanceMetrics() {
        if (!this.modelPerformance) return;

        const accuracy = this.modelPerformance.accuracy || 0;
        const precision = this.modelPerformance.precision || 0;
        const recall = this.modelPerformance.recall || 0;

        // Update summary cards
        const accuracyEl = document.getElementById('modelAccuracy');
        if (accuracyEl) {
            accuracyEl.textContent = `${(accuracy * 100).toFixed(2)}%`;
        }

        const precisionEl = document.getElementById('modelPrecision');
        if (precisionEl) {
            precisionEl.textContent = `${(precision * 100).toFixed(2)}%`;
        }

        const recallEl = document.getElementById('modelRecall');
        if (recallEl) {
            recallEl.textContent = `${(recall * 100).toFixed(2)}%`;
        }

        // Render performance chart
        this.renderPerformanceChart();
    }

    /**
     * Render performance chart
     */
    renderPerformanceChart() {
        const container = document.getElementById('performanceChart');
        if (!container || !this.modelPerformance) return;

        container.innerHTML = '<canvas id="performanceChartCanvas"></canvas>';
        
        const ctx = document.getElementById('performanceChartCanvas').getContext('2d');
        
        this.performanceChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: ['Accuracy', 'Precision', 'Recall', 'F1 Score'],
                datasets: [{
                    label: 'Performance Metrics',
                    data: [
                        (this.modelPerformance.accuracy || 0) * 100,
                        (this.modelPerformance.precision || 0) * 100,
                        (this.modelPerformance.recall || 0) * 100,
                        (this.modelPerformance.f1_score || 0) * 100
                    ],
                    backgroundColor: [
                        '#4a90e2',
                        '#28a745',
                        '#ffc107',
                        '#dc3545'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,
                        title: {
                            display: true,
                            text: 'Percentage (%)'
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }

    /**
     * Update summary cards
     */
    updateSummary() {
        if (this.ensembleInfo) {
            const activeModels = this.ensembleInfo.active_models || 0;
            const totalModels = this.ensembleInfo.total_models || 0;

            const activeEl = document.getElementById('activeModels');
            if (activeEl) {
                activeEl.textContent = activeModels;
            }

            const totalEl = document.getElementById('totalModels');
            if (totalEl) {
                totalEl.textContent = totalModels;
            }
        }
    }

    /**
     * Show error message
     */
    showError(message) {
        const container = document.getElementById('modelsContent');
        if (container) {
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error';
            errorDiv.innerHTML = `<i class="fas fa-exclamation-circle"></i> ${message}`;
            container.insertBefore(errorDiv, container.firstChild);
        }
    }

    /**
     * Format date
     */
    formatDate(dateString) {
        if (!dateString) return 'N/A';
        const date = new Date(dateString);
        return date.toLocaleString();
    }

    /**
     * Get authentication token
     */
    getAuthToken() {
        const token = localStorage.getItem('auth_token') || localStorage.getItem('access_token');
        if (token) {
            return token;
        }

        const cookies = document.cookie.split(';');
        for (let cookie of cookies) {
            const [name, value] = cookie.trim().split('=');
            if (name === 'auth_token' || name === 'access_token') {
                return value;
            }
        }

        return null;
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.AdminModels = AdminModels;
}

