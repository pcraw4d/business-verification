/**
 * Merchant Comparison Component
 * Handles side-by-side merchant comparison with exportable reports and data visualization
 */

class MerchantComparison {
    constructor() {
        this.merchant1 = null;
        this.merchant2 = null;
        this.comparisonData = null;
        this.currentView = 'side-by-side'; // 'side-by-side' or 'summary'
        this.riskChart = null;
        
        this.initializeEventListeners();
        this.loadMerchantOptions();
        this.initializeChart();
    }

    /**
     * Initialize event listeners for the comparison interface
     */
    initializeEventListeners() {
        // Merchant selection
        document.getElementById('merchant1Select').addEventListener('change', (e) => {
            this.handleMerchantSelection(1, e.target.value);
        });

        document.getElementById('merchant2Select').addEventListener('change', (e) => {
            this.handleMerchantSelection(2, e.target.value);
        });

        // Search functionality
        document.getElementById('merchant1Search').addEventListener('input', (e) => {
            this.filterMerchantOptions(1, e.target.value);
        });

        document.getElementById('merchant2Search').addEventListener('input', (e) => {
            this.filterMerchantOptions(2, e.target.value);
        });

        // Action buttons
        document.getElementById('exportReportBtn').addEventListener('click', () => {
            this.exportComparisonReport();
        });

        document.getElementById('clearComparisonBtn').addEventListener('click', () => {
            this.clearComparison();
        });

        document.getElementById('toggleViewBtn').addEventListener('click', () => {
            this.toggleView();
        });

        document.getElementById('printReportBtn').addEventListener('click', () => {
            this.printReport();
        });

        // View merchant details
        window.viewMerchantDetail = (merchantNumber) => {
            this.viewMerchantDetail(merchantNumber);
        };
    }

    /**
     * Load merchant options from the API
     */
    async loadMerchantOptions() {
        try {
            const response = await fetch('/api/v1/merchants?page_size=1000');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            this.populateMerchantSelects(data.merchants);
        } catch (error) {
            console.error('Error loading merchants:', error);
            this.showError('Failed to load merchant options');
        }
    }

    /**
     * Populate merchant select dropdowns
     */
    populateMerchantSelects(merchants) {
        const select1 = document.getElementById('merchant1Select');
        const select2 = document.getElementById('merchant2Select');
        
        // Clear existing options (except the first placeholder)
        select1.innerHTML = '<option value="">Choose a merchant...</option>';
        select2.innerHTML = '<option value="">Choose a merchant...</option>';
        
        merchants.forEach(merchant => {
            const option1 = document.createElement('option');
            option1.value = merchant.id;
            option1.textContent = `${merchant.name} (${merchant.portfolio_type})`;
            option1.dataset.merchant = JSON.stringify(merchant);
            select1.appendChild(option1);
            
            const option2 = document.createElement('option');
            option2.value = merchant.id;
            option2.textContent = `${merchant.name} (${merchant.portfolio_type})`;
            option2.dataset.merchant = JSON.stringify(merchant);
            select2.appendChild(option2);
        });
    }

    /**
     * Filter merchant options based on search input
     */
    filterMerchantOptions(merchantNumber, searchTerm) {
        const select = document.getElementById(`merchant${merchantNumber}Select`);
        const options = select.querySelectorAll('option');
        
        options.forEach(option => {
            if (option.value === '') return; // Skip placeholder option
            
            const merchantData = JSON.parse(option.dataset.merchant);
            const searchText = `${merchantData.name} ${merchantData.industry} ${merchantData.portfolio_type}`.toLowerCase();
            
            if (searchText.includes(searchTerm.toLowerCase())) {
                option.style.display = 'block';
            } else {
                option.style.display = 'none';
            }
        });
    }

    /**
     * Handle merchant selection
     */
    async handleMerchantSelection(merchantNumber, merchantId) {
        if (!merchantId) {
            this.hideMerchantPreview(merchantNumber);
            return;
        }

        try {
            const response = await fetch(`/api/v1/merchants/${merchantId}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const merchant = await response.json();
            this.setMerchant(merchantNumber, merchant);
            this.showMerchantPreview(merchantNumber, merchant);
            
            // Check if both merchants are selected
            if (this.merchant1 && this.merchant2) {
                await this.performComparison();
            }
        } catch (error) {
            console.error(`Error loading merchant ${merchantNumber}:`, error);
            this.showError(`Failed to load merchant ${merchantNumber}`);
        }
    }

    /**
     * Set merchant data
     */
    setMerchant(merchantNumber, merchant) {
        if (merchantNumber === 1) {
            this.merchant1 = merchant;
        } else {
            this.merchant2 = merchant;
        }
    }

    /**
     * Show merchant preview
     */
    showMerchantPreview(merchantNumber, merchant) {
        const preview = document.getElementById(`merchant${merchantNumber}Preview`);
        const name = document.getElementById(`merchant${merchantNumber}Name`);
        const type = document.getElementById(`merchant${merchantNumber}Type`);
        const industry = document.getElementById(`merchant${merchantNumber}Industry`);
        const location = document.getElementById(`merchant${merchantNumber}Location`);
        
        name.textContent = merchant.name;
        type.textContent = merchant.portfolio_type;
        type.className = `portfolio-type ${merchant.portfolio_type}`;
        industry.textContent = merchant.industry || 'Industry not specified';
        location.textContent = this.formatAddress(merchant.address);
        
        preview.style.display = 'block';
    }

    /**
     * Hide merchant preview
     */
    hideMerchantPreview(merchantNumber) {
        const preview = document.getElementById(`merchant${merchantNumber}Preview`);
        preview.style.display = 'none';
        
        if (merchantNumber === 1) {
            this.merchant1 = null;
        } else {
            this.merchant2 = null;
        }
    }

    /**
     * Format address for display
     */
    formatAddress(address) {
        if (!address) return 'Address not available';
        
        const parts = [
            address.street1,
            address.city,
            address.state,
            address.country
        ].filter(part => part && part.trim());
        
        return parts.join(', ');
    }

    /**
     * Perform comparison between two merchants
     */
    async performComparison() {
        if (!this.merchant1 || !this.merchant2) {
            return;
        }

        this.showLoadingState();
        
        try {
            // Generate comparison data
            this.comparisonData = this.generateComparisonData();
            
            // Populate comparison views
            this.populateSideBySideView();
            this.populateSummaryView();
            
            // Show comparison results
            this.showComparisonResults();
            
            // Enable export button
            document.getElementById('exportReportBtn').disabled = false;
            
        } catch (error) {
            console.error('Error performing comparison:', error);
            this.showError('Failed to perform comparison');
        } finally {
            this.hideLoadingState();
        }
    }

    /**
     * Generate comparison data
     */
    generateComparisonData() {
        const comparison = {
            merchant1: this.merchant1,
            merchant2: this.merchant2,
            differences: this.findDifferences(),
            riskComparison: this.compareRiskLevels(),
            recommendations: this.generateRecommendations(),
            summary: this.generateSummary()
        };
        
        return comparison;
    }

    /**
     * Find key differences between merchants
     */
    findDifferences() {
        const differences = [];
        
        // Portfolio type comparison
        if (this.merchant1.portfolio_type !== this.merchant2.portfolio_type) {
            differences.push({
                category: 'Portfolio Type',
                merchant1: this.merchant1.portfolio_type,
                merchant2: this.merchant2.portfolio_type,
                impact: 'medium'
            });
        }
        
        // Risk level comparison
        if (this.merchant1.risk_level !== this.merchant2.risk_level) {
            differences.push({
                category: 'Risk Level',
                merchant1: this.merchant1.risk_level,
                merchant2: this.merchant2.risk_level,
                impact: 'high'
            });
        }
        
        // Industry comparison
        if (this.merchant1.industry !== this.merchant2.industry) {
            differences.push({
                category: 'Industry',
                merchant1: this.merchant1.industry,
                merchant2: this.merchant2.industry,
                impact: 'medium'
            });
        }
        
        // Employee count comparison
        if (this.merchant1.employee_count !== this.merchant2.employee_count) {
            differences.push({
                category: 'Employee Count',
                merchant1: this.merchant1.employee_count,
                merchant2: this.merchant2.employee_count,
                impact: 'low'
            });
        }
        
        // Annual revenue comparison
        if (this.merchant1.annual_revenue !== this.merchant2.annual_revenue) {
            differences.push({
                category: 'Annual Revenue',
                merchant1: this.formatCurrency(this.merchant1.annual_revenue),
                merchant2: this.formatCurrency(this.merchant2.annual_revenue),
                impact: 'medium'
            });
        }
        
        return differences;
    }

    /**
     * Compare risk levels
     */
    compareRiskLevels() {
        const riskValues = {
            'low': 1,
            'medium': 2,
            'high': 3
        };
        
        return {
            merchant1: {
                level: this.merchant1.risk_level,
                value: riskValues[this.merchant1.risk_level] || 0
            },
            merchant2: {
                level: this.merchant2.risk_level,
                value: riskValues[this.merchant2.risk_level] || 0
            }
        };
    }

    /**
     * Generate recommendations based on comparison
     */
    generateRecommendations() {
        const recommendations = [];
        
        // Risk-based recommendations
        if (this.merchant1.risk_level === 'high' && this.merchant2.risk_level !== 'high') {
            recommendations.push({
                type: 'risk',
                priority: 'high',
                message: `${this.merchant1.name} requires enhanced monitoring due to high risk level`,
                action: 'Implement additional compliance checks and regular reviews'
            });
        }
        
        if (this.merchant2.risk_level === 'high' && this.merchant1.risk_level !== 'high') {
            recommendations.push({
                type: 'risk',
                priority: 'high',
                message: `${this.merchant2.name} requires enhanced monitoring due to high risk level`,
                action: 'Implement additional compliance checks and regular reviews'
            });
        }
        
        // Portfolio type recommendations
        if (this.merchant1.portfolio_type === 'prospective' && this.merchant2.portfolio_type === 'onboarded') {
            recommendations.push({
                type: 'onboarding',
                priority: 'medium',
                message: `Consider accelerating onboarding process for ${this.merchant1.name}`,
                action: 'Review onboarding requirements and expedite approval process'
            });
        }
        
        if (this.merchant2.portfolio_type === 'prospective' && this.merchant1.portfolio_type === 'onboarded') {
            recommendations.push({
                type: 'onboarding',
                priority: 'medium',
                message: `Consider accelerating onboarding process for ${this.merchant2.name}`,
                action: 'Review onboarding requirements and expedite approval process'
            });
        }
        
        // Industry recommendations
        if (this.merchant1.industry !== this.merchant2.industry) {
            recommendations.push({
                type: 'diversification',
                priority: 'low',
                message: 'Different industries provide portfolio diversification',
                action: 'Monitor industry-specific risks and compliance requirements'
            });
        }
        
        return recommendations;
    }

    /**
     * Generate comparison summary
     */
    generateSummary() {
        const riskComparison = this.compareRiskLevels();
        const differences = this.findDifferences();
        
        return {
            totalDifferences: differences.length,
            highImpactDifferences: differences.filter(d => d.impact === 'high').length,
            riskAdvantage: riskComparison.merchant1.value < riskComparison.merchant2.value ? 
                this.merchant1.name : this.merchant2.name,
            portfolioAdvantage: this.merchant1.portfolio_type === 'onboarded' && 
                this.merchant2.portfolio_type !== 'onboarded' ? 
                this.merchant1.name : this.merchant2.name
        };
    }

    /**
     * Populate side-by-side comparison view
     */
    populateSideBySideView() {
        // Populate merchant 1 data
        this.populateMerchantColumn(1, this.merchant1);
        
        // Populate merchant 2 data
        this.populateMerchantColumn(2, this.merchant2);
    }

    /**
     * Populate a merchant column in the comparison view
     */
    populateMerchantColumn(merchantNumber, merchant) {
        const prefix = `compareMerchant${merchantNumber}`;
        
        // Basic information
        document.getElementById(`${prefix}Name`).textContent = merchant.name;
        document.getElementById(`${prefix}Type`).textContent = merchant.portfolio_type;
        document.getElementById(`${prefix}Type`).className = `portfolio-type ${merchant.portfolio_type}`;
        document.getElementById(`${prefix}Risk`).textContent = merchant.risk_level;
        document.getElementById(`${prefix}Risk`).className = `risk-level ${merchant.risk_level}`;
        
        document.getElementById(`${prefix}BusinessName`).textContent = merchant.name;
        document.getElementById(`${prefix}Industry`).textContent = merchant.industry || 'Not specified';
        document.getElementById(`${prefix}Address`).textContent = this.formatAddress(merchant.address);
        document.getElementById(`${prefix}Phone`).textContent = merchant.contact_info?.phone || 'Not specified';
        document.getElementById(`${prefix}Email`).textContent = merchant.contact_info?.email || 'Not specified';
        document.getElementById(`${prefix}Website`).textContent = merchant.contact_info?.website || 'Not specified';
        
        // Financial information (mock data for MVP)
        document.getElementById(`${prefix}Revenue`).textContent = this.formatCurrency(merchant.annual_revenue);
        document.getElementById(`${prefix}Volume`).textContent = this.formatCurrency(merchant.annual_revenue * 0.1); // Mock transaction volume
        document.getElementById(`${prefix}AvgTransaction`).textContent = this.formatCurrency(merchant.annual_revenue * 0.001); // Mock avg transaction
        document.getElementById(`${prefix}ChargebackRate`).textContent = '0.5%'; // Mock chargeback rate
        
        // Risk assessment (mock data for MVP)
        const riskScore = this.calculateRiskScore(merchant.risk_level);
        document.getElementById(`${prefix}RiskScore`).textContent = `${(riskScore * 100).toFixed(1)}%`;
        document.getElementById(`${prefix}RiskScore`).className = `metric-value risk-${merchant.risk_level}`;
        
        // Risk breakdown (mock data)
        this.populateRiskBreakdown(merchantNumber, merchant.risk_level);
        
        // Compliance status (mock data for MVP)
        this.populateComplianceStatus(merchantNumber, merchant);
    }

    /**
     * Calculate risk score based on risk level
     */
    calculateRiskScore(riskLevel) {
        const scores = {
            'low': 0.2,
            'medium': 0.5,
            'high': 0.8
        };
        return scores[riskLevel] || 0.5;
    }

    /**
     * Populate risk breakdown
     */
    populateRiskBreakdown(merchantNumber, riskLevel) {
        const baseScore = this.calculateRiskScore(riskLevel);
        const variations = [0.1, -0.05, 0.15]; // Small variations for different risk categories
        
        const categories = ['Compliance', 'Financial', 'Operational'];
        categories.forEach((category, index) => {
            const score = Math.max(0, Math.min(1, baseScore + variations[index]));
            const element = document.getElementById(`compareMerchant${merchantNumber}${category}Risk`);
            if (element) {
                element.style.width = `${score * 100}%`;
                element.className = `risk-fill risk-${riskLevel}`;
            }
        });
    }

    /**
     * Populate compliance status
     */
    populateComplianceStatus(merchantNumber, merchant) {
        const complianceChecks = ['KYC', 'AML', 'Sanctions', 'PEP'];
        
        complianceChecks.forEach(check => {
            const element = document.getElementById(`compareMerchant${merchantNumber}${check}`);
            const statusElement = element.querySelector('.status');
            
            // Mock compliance status based on portfolio type and risk level
            let status = 'pending';
            if (merchant.portfolio_type === 'onboarded') {
                status = 'completed';
            } else if (merchant.risk_level === 'high') {
                status = 'failed';
            }
            
            statusElement.textContent = status;
            statusElement.className = `status ${status}`;
        });
    }

    /**
     * Populate summary view
     */
    populateSummaryView() {
        // Key differences
        this.populateKeyDifferences();
        
        // Risk comparison chart
        this.updateRiskComparisonChart();
        
        // Recommendations
        this.populateRecommendations();
    }

    /**
     * Populate key differences
     */
    populateKeyDifferences() {
        const container = document.getElementById('keyDifferences');
        container.innerHTML = '';
        
        this.comparisonData.differences.forEach(diff => {
            const diffElement = document.createElement('div');
            diffElement.className = `difference-item impact-${diff.impact}`;
            diffElement.innerHTML = `
                <div class="difference-category">${diff.category}</div>
                <div class="difference-values">
                    <div class="merchant1-value">
                        <strong>${this.merchant1.name}:</strong> ${diff.merchant1}
                    </div>
                    <div class="merchant2-value">
                        <strong>${this.merchant2.name}:</strong> ${diff.merchant2}
                    </div>
                </div>
            `;
            container.appendChild(diffElement);
        });
    }

    /**
     * Update risk comparison chart
     */
    updateRiskComparisonChart() {
        if (!this.riskChart) {
            this.initializeChart();
        }
        
        const riskComparison = this.comparisonData.riskComparison;
        
        this.riskChart.data.datasets[0].data = [
            riskComparison.merchant1.value,
            riskComparison.merchant2.value
        ];
        
        this.riskChart.update();
    }

    /**
     * Initialize risk comparison chart
     */
    initializeChart() {
        const ctx = document.getElementById('riskComparisonChart');
        if (!ctx) return;
        
        this.riskChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [this.merchant1?.name || 'Merchant 1', this.merchant2?.name || 'Merchant 2'],
                datasets: [{
                    label: 'Risk Level',
                    data: [0, 0],
                    backgroundColor: ['#dc3545', '#ffc107'],
                    borderColor: ['#dc3545', '#ffc107'],
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 3,
                        ticks: {
                            stepSize: 1,
                            callback: function(value) {
                                const labels = ['', 'Low', 'Medium', 'High'];
                                return labels[value] || '';
                            }
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
     * Populate recommendations
     */
    populateRecommendations() {
        const container = document.getElementById('recommendations');
        container.innerHTML = '';
        
        this.comparisonData.recommendations.forEach(rec => {
            const recElement = document.createElement('div');
            recElement.className = `recommendation-item priority-${rec.priority}`;
            recElement.innerHTML = `
                <div class="recommendation-header">
                    <i class="fas fa-${this.getRecommendationIcon(rec.type)}"></i>
                    <span class="recommendation-type">${rec.type.toUpperCase()}</span>
                    <span class="recommendation-priority">${rec.priority.toUpperCase()}</span>
                </div>
                <div class="recommendation-message">${rec.message}</div>
                <div class="recommendation-action">${rec.action}</div>
            `;
            container.appendChild(recElement);
        });
    }

    /**
     * Get icon for recommendation type
     */
    getRecommendationIcon(type) {
        const icons = {
            'risk': 'exclamation-triangle',
            'onboarding': 'user-plus',
            'diversification': 'chart-pie'
        };
        return icons[type] || 'info-circle';
    }

    /**
     * Show comparison results
     */
    showComparisonResults() {
        document.getElementById('comparisonResults').style.display = 'block';
        document.getElementById('emptyState').style.display = 'none';
    }

    /**
     * Toggle between side-by-side and summary views
     */
    toggleView() {
        const sideBySideView = document.getElementById('sideBySideView');
        const summaryView = document.getElementById('summaryView');
        const toggleBtn = document.getElementById('toggleViewBtn');
        
        if (this.currentView === 'side-by-side') {
            sideBySideView.style.display = 'none';
            summaryView.style.display = 'block';
            toggleBtn.innerHTML = '<i class="fas fa-columns"></i> Side-by-Side View';
            this.currentView = 'summary';
        } else {
            sideBySideView.style.display = 'block';
            summaryView.style.display = 'none';
            toggleBtn.innerHTML = '<i class="fas fa-th-large"></i> Summary View';
            this.currentView = 'side-by-side';
        }
    }

    /**
     * Export comparison report
     */
    exportComparisonReport() {
        if (!this.comparisonData) {
            this.showError('No comparison data to export');
            return;
        }
        
        const report = this.generateExportReport();
        this.downloadReport(report);
    }

    /**
     * Generate export report
     */
    generateExportReport() {
        const report = {
            title: 'Merchant Comparison Report',
            generatedAt: new Date().toISOString(),
            merchants: {
                merchant1: this.merchant1,
                merchant2: this.merchant2
            },
            comparison: this.comparisonData,
            summary: {
                totalDifferences: this.comparisonData.differences.length,
                highImpactDifferences: this.comparisonData.differences.filter(d => d.impact === 'high').length,
                recommendations: this.comparisonData.recommendations.length
            }
        };
        
        return report;
    }

    /**
     * Download report as JSON file
     */
    downloadReport(report) {
        const dataStr = JSON.stringify(report, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(dataBlob);
        
        const link = document.createElement('a');
        link.href = url;
        link.download = `merchant-comparison-${Date.now()}.json`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        URL.revokeObjectURL(url);
    }

    /**
     * Print report
     */
    printReport() {
        if (!this.comparisonData) {
            this.showError('No comparison data to print');
            return;
        }
        
        // Create print-friendly version
        const printWindow = window.open('', '_blank');
        const printContent = this.generatePrintContent();
        
        printWindow.document.write(printContent);
        printWindow.document.close();
        printWindow.print();
    }

    /**
     * Generate print content
     */
    generatePrintContent() {
        return `
            <!DOCTYPE html>
            <html>
            <head>
                <title>Merchant Comparison Report</title>
                <style>
                    body { font-family: Arial, sans-serif; margin: 20px; }
                    .header { text-align: center; margin-bottom: 30px; }
                    .merchant-section { margin-bottom: 30px; }
                    .comparison-table { width: 100%; border-collapse: collapse; }
                    .comparison-table th, .comparison-table td { border: 1px solid #ddd; padding: 8px; text-align: left; }
                    .comparison-table th { background-color: #f2f2f2; }
                    .recommendations { margin-top: 30px; }
                    .recommendation-item { margin-bottom: 15px; padding: 10px; border-left: 4px solid #007bff; }
                </style>
            </head>
            <body>
                <div class="header">
                    <h1>Merchant Comparison Report</h1>
                    <p>Generated on: ${new Date().toLocaleDateString()}</p>
                </div>
                
                <div class="merchant-section">
                    <h2>Merchants Compared</h2>
                    <table class="comparison-table">
                        <tr>
                            <th>Attribute</th>
                            <th>${this.merchant1.name}</th>
                            <th>${this.merchant2.name}</th>
                        </tr>
                        <tr><td>Portfolio Type</td><td>${this.merchant1.portfolio_type}</td><td>${this.merchant2.portfolio_type}</td></tr>
                        <tr><td>Risk Level</td><td>${this.merchant1.risk_level}</td><td>${this.merchant2.risk_level}</td></tr>
                        <tr><td>Industry</td><td>${this.merchant1.industry}</td><td>${this.merchant2.industry}</td></tr>
                        <tr><td>Employee Count</td><td>${this.merchant1.employee_count}</td><td>${this.merchant2.employee_count}</td></tr>
                        <tr><td>Annual Revenue</td><td>${this.formatCurrency(this.merchant1.annual_revenue)}</td><td>${this.formatCurrency(this.merchant2.annual_revenue)}</td></tr>
                    </table>
                </div>
                
                <div class="recommendations">
                    <h2>Recommendations</h2>
                    ${this.comparisonData.recommendations.map(rec => `
                        <div class="recommendation-item">
                            <strong>${rec.type.toUpperCase()} - ${rec.priority.toUpperCase()}</strong><br>
                            ${rec.message}<br>
                            <em>Action: ${rec.action}</em>
                        </div>
                    `).join('')}
                </div>
            </body>
            </html>
        `;
    }

    /**
     * Clear comparison
     */
    clearComparison() {
        this.merchant1 = null;
        this.merchant2 = null;
        this.comparisonData = null;
        
        // Reset form
        document.getElementById('merchant1Select').value = '';
        document.getElementById('merchant2Select').value = '';
        document.getElementById('merchant1Search').value = '';
        document.getElementById('merchant2Search').value = '';
        
        // Hide previews
        this.hideMerchantPreview(1);
        this.hideMerchantPreview(2);
        
        // Hide comparison results
        document.getElementById('comparisonResults').style.display = 'none';
        document.getElementById('emptyState').style.display = 'block';
        
        // Disable export button
        document.getElementById('exportReportBtn').disabled = true;
        
        // Reset view
        this.currentView = 'side-by-side';
        document.getElementById('sideBySideView').style.display = 'block';
        document.getElementById('summaryView').style.display = 'none';
        document.getElementById('toggleViewBtn').innerHTML = '<i class="fas fa-th-large"></i> Summary View';
    }

    /**
     * View merchant detail
     */
    viewMerchantDetail(merchantNumber) {
        const merchant = merchantNumber === 1 ? this.merchant1 : this.merchant2;
        if (merchant) {
            // Navigate to merchant detail page
            window.location.href = `merchant-detail.html?id=${merchant.id}`;
        }
    }

    /**
     * Show loading state
     */
    showLoadingState() {
        document.getElementById('loadingState').style.display = 'block';
    }

    /**
     * Hide loading state
     */
    hideLoadingState() {
        document.getElementById('loadingState').style.display = 'none';
    }

    /**
     * Show error message
     */
    showError(message) {
        // Create or update error message
        let errorElement = document.getElementById('errorMessage');
        if (!errorElement) {
            errorElement = document.createElement('div');
            errorElement.id = 'errorMessage';
            errorElement.className = 'error-message';
            document.querySelector('.main-content').insertBefore(errorElement, document.querySelector('.main-content').firstChild);
        }
        
        errorElement.innerHTML = `
            <div class="error-content">
                <i class="fas fa-exclamation-triangle"></i>
                <span>${message}</span>
                <button onclick="this.parentElement.parentElement.remove()" class="error-close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;
        
        // Auto-hide after 5 seconds
        setTimeout(() => {
            if (errorElement && errorElement.parentElement) {
                errorElement.remove();
            }
        }, 5000);
    }

    /**
     * Format currency for display
     */
    formatCurrency(amount) {
        if (!amount) return 'Not specified';
        
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0
        }).format(amount);
    }
}

// Initialize the merchant comparison component when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new MerchantComparison();
});
