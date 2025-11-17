/**
 * Export Button Component
 * Reusable export button for merchant and risk data
 */
class ExportButton {
    constructor(options = {}) {
        this.container = options.container || null;
        this.exportService = null;
        this.dataSource = options.dataSource || null;
        this.exportType = options.exportType || 'merchant'; // 'merchant', 'risk', 'report'
        this.formats = options.formats || ['csv', 'pdf', 'json'];
        this.onExportStart = options.onExportStart || null;
        this.onExportComplete = options.onExportComplete || null;
        this.onExportError = options.onExportError || null;
    }

    /**
     * Initialize the export button
     */
    async init() {
        // Load export service
        await this.loadExportService();
        
        // Create button UI
        this.createButton();
    }

    /**
     * Load export service from shared components
     */
    async loadExportService() {
        try {
            // Ensure shared components are loaded
            if (typeof loadSharedComponents === 'function') {
                await loadSharedComponents();
            }
            
            // Try to get from shared components
            if (typeof getExportService !== 'undefined') {
                this.exportService = getExportService();
            } else if (typeof window !== 'undefined' && window.getExportService) {
                this.exportService = window.getExportService();
            } else if (typeof SharedExportService !== 'undefined') {
                // Fallback: create new instance if getter not available
                this.exportService = new SharedExportService();
            } else {
                // Fallback: try to import
                try {
                    const { getExportService } = await import('../../shared/components/export-service.js');
                    this.exportService = getExportService();
                } catch (error) {
                    console.warn('Could not load export service, using API directly');
                    this.exportService = null;
                }
            }
        } catch (error) {
            console.error('Error loading export service:', error);
            this.exportService = null;
        }
    }

    /**
     * Create the export button UI
     */
    createButton() {
        if (!this.container) {
            console.error('Export button container not found');
            return;
        }

        const buttonHTML = `
            <div class="export-button-container">
                <button class="export-button" id="exportBtn">
                    <i class="fas fa-download"></i>
                    <span>Export</span>
                </button>
                <div class="export-dropdown" id="exportDropdown" style="display: none;">
                    ${this.formats.map(format => `
                        <button class="export-option" data-format="${format}" onclick="exportButton.export('${format}')">
                            <i class="fas fa-file-${this.getFormatIcon(format)}"></i>
                            Export as ${format.toUpperCase()}
                        </button>
                    `).join('')}
                </div>
            </div>
        `;

        this.container.innerHTML = buttonHTML;
        this.bindEvents();
    }

    /**
     * Get icon for format
     */
    getFormatIcon(format) {
        const icons = {
            'csv': 'csv',
            'pdf': 'pdf',
            'json': 'code',
            'excel': 'excel',
            'xlsx': 'excel'
        };
        return icons[format] || 'download';
    }

    /**
     * Bind event handlers
     */
    bindEvents() {
        const button = document.getElementById('exportBtn');
        const dropdown = document.getElementById('exportDropdown');

        if (button) {
            button.addEventListener('click', (e) => {
                e.stopPropagation();
                dropdown.style.display = dropdown.style.display === 'none' ? 'block' : 'none';
            });
        }

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.container.contains(e.target)) {
                dropdown.style.display = 'none';
            }
        });
    }

    /**
     * Export data in specified format
     */
    async export(format) {
        try {
            // Close dropdown
            document.getElementById('exportDropdown').style.display = 'none';

            // Get data
            const data = await this.getExportData();
            if (!data) {
                throw new Error('No data available to export');
            }

            // Call onExportStart callback
            if (this.onExportStart) {
                this.onExportStart(format);
            }

            // Export using service or API
            let result;
            if (this.exportService) {
                result = await this.exportService.exportData(data, {
                    format: format,
                    template: this.getTemplateName(),
                    filename: this.getFilename(format)
                });
            } else {
                result = await this.exportViaAPI(data, format);
            }

            // Call onExportComplete callback
            if (this.onExportComplete) {
                this.onExportComplete(format, result);
            }

            return result;
        } catch (error) {
            console.error('Export error:', error);
            
            // Call onExportError callback
            if (this.onExportError) {
                this.onExportError(format, error);
            } else {
                alert(`Export failed: ${error.message}`);
            }
            
            throw error;
        }
    }

    /**
     * Get data to export
     */
    async getExportData() {
        if (this.dataSource) {
            if (typeof this.dataSource === 'function') {
                return await this.dataSource();
            }
            return this.dataSource;
        }

        // Try to get from current page context
        if (this.exportType === 'merchant') {
            return this.getMerchantData();
        } else if (this.exportType === 'risk') {
            return this.getRiskData();
        }

        return null;
    }

    /**
     * Get merchant data from page
     */
    getMerchantData() {
        // Try to get from merchant detail page
        const merchantId = this.getMerchantId();
        if (!merchantId) {
            return null;
        }

        // Return data structure for export
        return {
            type: 'merchant',
            merchant_id: merchantId,
            timestamp: new Date().toISOString()
        };
    }

    /**
     * Get risk data from page
     */
    getRiskData() {
        const merchantId = this.getMerchantId();
        if (!merchantId) {
            return null;
        }

        return {
            type: 'risk',
            merchant_id: merchantId,
            timestamp: new Date().toISOString()
        };
    }

    /**
     * Get merchant ID from URL or page context
     */
    getMerchantId() {
        // Try URL parameter
        const urlParams = new URLSearchParams(window.location.search);
        const id = urlParams.get('id') || urlParams.get('merchant_id');
        if (id) {
            return id;
        }

        // Try to get from page data
        if (window.currentMerchant && window.currentMerchant.id) {
            return window.currentMerchant.id;
        }

        return null;
    }

    /**
     * Export via API
     */
    async exportViaAPI(data, format) {
        const endpoint = this.exportType === 'risk' 
            ? '/api/v1/export' 
            : '/api/v1/reports/export';

        // Show loading state
        const button = document.getElementById('exportBtn');
        if (button) {
            button.disabled = true;
            button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Exporting...';
        }

        try {
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                },
                body: JSON.stringify({
                    ...data,
                    format: format
                })
            });

            if (!response.ok) {
                throw new Error(`Export failed: ${response.statusText}`);
            }

            const result = await response.json();
            
            // Download file if URL provided
            if (result.file_url) {
                window.open(result.file_url, '_blank');
            } else if (result.download_url) {
                window.open(result.download_url, '_blank');
            }

            return result;
        } finally {
            // Restore button state
            if (button) {
                button.disabled = false;
                button.innerHTML = '<i class="fas fa-download"></i> <span>Export</span>';
            }
        }
    }

    /**
     * Get template name based on export type
     */
    getTemplateName() {
        const templates = {
            'merchant': 'merchantReport',
            'risk': 'riskReport',
            'report': 'complianceReport'
        };
        return templates[this.exportType] || 'default';
    }

    /**
     * Generate filename
     */
    getFilename(format) {
        const timestamp = new Date().toISOString().split('T')[0];
        const type = this.exportType.charAt(0).toUpperCase() + this.exportType.slice(1);
        return `${type}_Export_${timestamp}.${format}`;
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
    window.ExportButton = ExportButton;
}

// Global instance for inline handlers
let exportButton = null;

