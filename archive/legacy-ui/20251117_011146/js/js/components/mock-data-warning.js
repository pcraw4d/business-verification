/**
 * Mock Data Warning Component
 * Displays clear indicators for test data and mock data sources
 * Integrates with placeholder service to show data source information
 */
class MockDataWarning {
    constructor(options = {}) {
        this.container = options.container || document.body;
        this.apiBaseUrl = options.apiBaseUrl || '/api/v1';
        this.dataSource = options.dataSource || 'mock';
        this.warningLevel = options.warningLevel || 'info'; // info, warning, error
        this.showDataSource = options.showDataSource !== false;
        this.showDataCount = options.showDataCount !== false;
        this.autoHide = options.autoHide || false;
        this.autoHideDelay = options.autoHideDelay || 10000; // 10 seconds
        this.isVisible = false;
        this.dataInfo = null;
        this.hideTimer = null;
        
        // Event callbacks
        this.onWarningClick = options.onWarningClick || null;
        this.onWarningDismiss = options.onWarningDismiss || null;
        this.onDataSourceClick = options.onDataSourceClick || null;
        
        this.init();
    }

    init() {
        this.createWarningInterface();
        this.bindEvents();
        this.loadDataSourceInfo();
        
        if (this.autoHide) {
            this.startAutoHide();
        }
    }

    createWarningInterface() {
        const warningHTML = `
            <div class="mock-data-warning" id="mockDataWarning" style="display: none;">
                <div class="warning-content">
                    <div class="warning-header">
                        <div class="warning-icon">
                            <i class="fas fa-database"></i>
                        </div>
                        <div class="warning-title-section">
                            <h4 class="warning-title" id="warningTitle">Mock Data Active</h4>
                            <p class="warning-subtitle" id="warningSubtitle">This interface is using test data for demonstration purposes</p>
                        </div>
                        <div class="warning-actions">
                            <button class="warning-close-btn" id="warningCloseBtn" title="Dismiss warning">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                    </div>
                    
                    <div class="warning-body">
                        <div class="data-source-info" id="dataSourceInfo">
                            <div class="data-source-item">
                                <div class="data-source-label">
                                    <i class="fas fa-info-circle"></i>
                                    <span>Data Source:</span>
                                </div>
                                <div class="data-source-value" id="dataSourceValue">Loading...</div>
                            </div>
                            
                            <div class="data-source-item" id="dataCountItem" style="display: none;">
                                <div class="data-source-label">
                                    <i class="fas fa-list"></i>
                                    <span>Records:</span>
                                </div>
                                <div class="data-source-value" id="dataCountValue">Loading...</div>
                            </div>
                            
                            <div class="data-source-item" id="lastUpdatedItem" style="display: none;">
                                <div class="data-source-label">
                                    <i class="fas fa-clock"></i>
                                    <span>Last Updated:</span>
                                </div>
                                <div class="data-source-value" id="lastUpdatedValue">Loading...</div>
                            </div>
                            
                            <div class="data-source-item" id="dataQualityItem" style="display: none;">
                                <div class="data-source-label">
                                    <i class="fas fa-check-circle"></i>
                                    <span>Data Quality:</span>
                                </div>
                                <div class="data-source-value" id="dataQualityValue">Loading...</div>
                            </div>
                        </div>
                        
                        <div class="warning-actions-bottom">
                            <button class="btn btn-outline" id="viewDataSourceBtn">
                                <i class="fas fa-eye"></i>
                                View Data Source
                            </button>
                            <button class="btn btn-primary" id="switchToRealDataBtn" style="display: none;">
                                <i class="fas fa-sync"></i>
                                Switch to Real Data
                            </button>
                            <button class="btn btn-secondary" id="dismissWarningBtn">
                                <i class="fas fa-check"></i>
                                Dismiss
                            </button>
                        </div>
                        
                        <div class="warning-footer">
                            <div class="warning-note">
                                <i class="fas fa-lightbulb"></i>
                                <span>This is a demonstration environment. All data shown is for testing purposes only.</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.container.innerHTML = warningHTML;
        this.addStyles();
    }

    addStyles() {
        const styles = `
            <style>
                .mock-data-warning {
                    position: fixed;
                    top: 20px;
                    left: 20px;
                    width: 400px;
                    max-width: calc(100vw - 40px);
                    background: linear-gradient(135deg, #ff9800 0%, #f57c00 100%);
                    border-radius: 12px;
                    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
                    z-index: 9999;
                    animation: slideInLeft 0.5s ease-out;
                    backdrop-filter: blur(10px);
                    border: 1px solid rgba(255, 255, 255, 0.1);
                }

                .warning-content {
                    padding: 20px;
                    color: white;
                }

                .warning-header {
                    display: flex;
                    align-items: flex-start;
                    gap: 12px;
                    margin-bottom: 16px;
                }

                .warning-icon {
                    width: 40px;
                    height: 40px;
                    background: rgba(255, 255, 255, 0.2);
                    border-radius: 10px;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 1.2rem;
                    flex-shrink: 0;
                    color: #fff3cd;
                }

                .warning-title-section {
                    flex: 1;
                }

                .warning-title {
                    margin: 0 0 4px 0;
                    font-size: 1.1rem;
                    font-weight: 600;
                    color: white;
                }

                .warning-subtitle {
                    margin: 0;
                    font-size: 0.85rem;
                    color: rgba(255, 255, 255, 0.9);
                    line-height: 1.4;
                }

                .warning-actions {
                    display: flex;
                    gap: 8px;
                }

                .warning-close-btn {
                    width: 28px;
                    height: 28px;
                    background: rgba(255, 255, 255, 0.1);
                    border: none;
                    border-radius: 6px;
                    color: white;
                    cursor: pointer;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    transition: all 0.3s ease;
                    font-size: 0.8rem;
                }

                .warning-close-btn:hover {
                    background: rgba(255, 255, 255, 0.2);
                    transform: scale(1.1);
                }

                .warning-body {
                    margin-bottom: 16px;
                }

                .data-source-info {
                    background: rgba(255, 255, 255, 0.1);
                    border-radius: 10px;
                    padding: 16px;
                    margin-bottom: 16px;
                }

                .data-source-item {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 12px;
                }

                .data-source-item:last-child {
                    margin-bottom: 0;
                }

                .data-source-label {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    font-size: 0.85rem;
                    color: rgba(255, 255, 255, 0.8);
                    font-weight: 500;
                }

                .data-source-label i {
                    font-size: 0.75rem;
                    opacity: 0.7;
                }

                .data-source-value {
                    font-size: 0.85rem;
                    color: white;
                    font-weight: 600;
                    text-align: right;
                    max-width: 60%;
                    word-break: break-word;
                }

                .warning-actions-bottom {
                    display: flex;
                    gap: 8px;
                    margin-bottom: 16px;
                    flex-wrap: wrap;
                }

                .btn {
                    padding: 8px 12px;
                    border: none;
                    border-radius: 6px;
                    font-size: 0.8rem;
                    font-weight: 600;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    display: flex;
                    align-items: center;
                    gap: 4px;
                    text-decoration: none;
                    flex: 1;
                    justify-content: center;
                    min-width: 0;
                }

                .btn:disabled {
                    opacity: 0.6;
                    cursor: not-allowed;
                }

                .btn-primary {
                    background: rgba(255, 255, 255, 0.2);
                    color: white;
                    border: 1px solid rgba(255, 255, 255, 0.3);
                }

                .btn-primary:hover:not(:disabled) {
                    background: rgba(255, 255, 255, 0.3);
                    transform: translateY(-1px);
                }

                .btn-outline {
                    background: transparent;
                    color: white;
                    border: 1px solid rgba(255, 255, 255, 0.3);
                }

                .btn-outline:hover:not(:disabled) {
                    background: rgba(255, 255, 255, 0.1);
                    transform: translateY(-1px);
                }

                .btn-secondary {
                    background: rgba(255, 255, 255, 0.1);
                    color: white;
                    border: 1px solid rgba(255, 255, 255, 0.2);
                }

                .btn-secondary:hover:not(:disabled) {
                    background: rgba(255, 255, 255, 0.2);
                    transform: translateY(-1px);
                }

                .warning-footer {
                    border-top: 1px solid rgba(255, 255, 255, 0.1);
                    padding-top: 12px;
                }

                .warning-note {
                    display: flex;
                    align-items: flex-start;
                    gap: 8px;
                    font-size: 0.75rem;
                    color: rgba(255, 255, 255, 0.8);
                    line-height: 1.4;
                }

                .warning-note i {
                    font-size: 0.7rem;
                    margin-top: 2px;
                    opacity: 0.7;
                }

                /* Warning level variations */
                .mock-data-warning.warning-level-error {
                    background: linear-gradient(135deg, #f44336 0%, #d32f2f 100%);
                }

                .mock-data-warning.warning-level-error .warning-icon {
                    color: #ffcdd2;
                }

                .mock-data-warning.warning-level-warning {
                    background: linear-gradient(135deg, #ff9800 0%, #f57c00 100%);
                }

                .mock-data-warning.warning-level-warning .warning-icon {
                    color: #fff3cd;
                }

                .mock-data-warning.warning-level-info {
                    background: linear-gradient(135deg, #2196f3 0%, #1976d2 100%);
                }

                .mock-data-warning.warning-level-info .warning-icon {
                    color: #bbdefb;
                }

                /* Animations */
                @keyframes slideInLeft {
                    from {
                        opacity: 0;
                        transform: translateX(-100%);
                    }
                    to {
                        opacity: 1;
                        transform: translateX(0);
                    }
                }

                @keyframes slideOutLeft {
                    from {
                        opacity: 1;
                        transform: translateX(0);
                    }
                    to {
                        opacity: 0;
                        transform: translateX(-100%);
                    }
                }

                .warning-closing {
                    animation: slideOutLeft 0.3s ease-in forwards;
                }

                /* Responsive Design */
                @media (max-width: 768px) {
                    .mock-data-warning {
                        top: 10px;
                        left: 10px;
                        right: 10px;
                        width: auto;
                        max-width: none;
                    }

                    .warning-content {
                        padding: 16px;
                    }

                    .warning-header {
                        gap: 10px;
                    }

                    .warning-icon {
                        width: 36px;
                        height: 36px;
                        font-size: 1rem;
                    }

                    .warning-title {
                        font-size: 1rem;
                    }

                    .warning-subtitle {
                        font-size: 0.8rem;
                    }

                    .data-source-item {
                        flex-direction: column;
                        align-items: flex-start;
                        gap: 4px;
                    }

                    .data-source-value {
                        text-align: left;
                        max-width: 100%;
                    }

                    .warning-actions-bottom {
                        flex-direction: column;
                    }

                    .btn {
                        width: 100%;
                    }
                }

                /* Dark mode support */
                @media (prefers-color-scheme: dark) {
                    .mock-data-warning {
                        background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
                    }
                }

                /* High contrast mode */
                @media (prefers-contrast: high) {
                    .mock-data-warning {
                        border: 2px solid white;
                    }

                    .warning-close-btn {
                        border: 1px solid white;
                    }
                }

                /* Reduced motion */
                @media (prefers-reduced-motion: reduce) {
                    .mock-data-warning {
                        animation: none;
                    }

                    .warning-closing {
                        animation: none;
                    }

                    .btn:hover:not(:disabled) {
                        transform: none;
                    }

                    .warning-close-btn:hover {
                        transform: none;
                    }
                }
            </style>
        `;

        // Add styles to head if not already added
        if (!document.querySelector('#mock-data-warning-styles')) {
            const styleElement = document.createElement('style');
            styleElement.id = 'mock-data-warning-styles';
            styleElement.textContent = styles;
            document.head.appendChild(styleElement);
        }
    }

    bindEvents() {
        const warningCloseBtn = document.getElementById('warningCloseBtn');
        const viewDataSourceBtn = document.getElementById('viewDataSourceBtn');
        const switchToRealDataBtn = document.getElementById('switchToRealDataBtn');
        const dismissWarningBtn = document.getElementById('dismissWarningBtn');

        // Close warning
        warningCloseBtn.addEventListener('click', () => {
            this.hide();
        });

        // View data source button
        viewDataSourceBtn.addEventListener('click', () => {
            this.handleViewDataSource();
        });

        // Switch to real data button
        switchToRealDataBtn.addEventListener('click', () => {
            this.handleSwitchToRealData();
        });

        // Dismiss warning
        dismissWarningBtn.addEventListener('click', () => {
            this.handleDismissWarning();
        });

        // Click outside to close
        document.addEventListener('click', (e) => {
            const warning = document.getElementById('mockDataWarning');
            if (this.isVisible && warning && !warning.contains(e.target)) {
                // Don't auto-close on outside click for better UX
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (this.isVisible) {
                if (e.key === 'Escape') {
                    this.hide();
                }
            }
        });
    }

    async loadDataSourceInfo() {
        try {
            const dataInfo = await this.getDataSourceInfo();
            if (dataInfo) {
                this.dataInfo = dataInfo;
                this.updateWarningContent();
                this.show();
            }
        } catch (error) {
            console.error('Error loading data source info:', error);
            // Show with default info
            this.updateWarningContent();
            this.show();
        }
    }

    async getDataSourceInfo() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/data-source/info`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            return data.success ? data.data : null;
        } catch (error) {
            console.error('Error fetching data source info:', error);
            // Return default mock data info
            return {
                source: 'mock',
                type: 'test_data',
                count: 5000,
                last_updated: new Date().toISOString(),
                quality: 'high',
                description: 'Generated test data for demonstration purposes'
            };
        }
    }

    updateWarningContent() {
        const warning = document.getElementById('mockDataWarning');
        
        // Set warning level class
        warning.className = `mock-data-warning warning-level-${this.warningLevel}`;
        
        // Update title and subtitle based on data source
        const title = document.getElementById('warningTitle');
        const subtitle = document.getElementById('warningSubtitle');
        
        if (this.dataSource === 'mock') {
            title.textContent = 'Mock Data Active';
            subtitle.textContent = 'This interface is using test data for demonstration purposes';
        } else if (this.dataSource === 'staging') {
            title.textContent = 'Staging Data Active';
            subtitle.textContent = 'This interface is using staging environment data';
        } else {
            title.textContent = 'Test Data Active';
            subtitle.textContent = 'This interface is using test data';
        }
        
        // Update data source information
        if (this.dataInfo) {
            document.getElementById('dataSourceValue').textContent = this.formatDataSource(this.dataInfo.source);
            
            if (this.showDataCount && this.dataInfo.count) {
                document.getElementById('dataCountItem').style.display = 'flex';
                document.getElementById('dataCountValue').textContent = this.formatDataCount(this.dataInfo.count);
            }
            
            if (this.dataInfo.last_updated) {
                document.getElementById('lastUpdatedItem').style.display = 'flex';
                document.getElementById('lastUpdatedValue').textContent = this.formatLastUpdated(this.dataInfo.last_updated);
            }
            
            if (this.dataInfo.quality) {
                document.getElementById('dataQualityItem').style.display = 'flex';
                document.getElementById('dataQualityValue').textContent = this.formatDataQuality(this.dataInfo.quality);
            }
        }
        
        // Show/hide switch to real data button based on environment
        const switchBtn = document.getElementById('switchToRealDataBtn');
        if (this.dataSource === 'mock' && this.canSwitchToRealData()) {
            switchBtn.style.display = 'flex';
        } else {
            switchBtn.style.display = 'none';
        }
    }

    formatDataSource(source) {
        const sources = {
            'mock': 'Mock Database',
            'staging': 'Staging Database',
            'test': 'Test Database',
            'demo': 'Demo Database',
            'production': 'Production Database'
        };
        return sources[source] || source;
    }

    formatDataCount(count) {
        if (count >= 1000000) {
            return `${(count / 1000000).toFixed(1)}M records`;
        } else if (count >= 1000) {
            return `${(count / 1000).toFixed(1)}K records`;
        } else {
            return `${count} records`;
        }
    }

    formatLastUpdated(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diffTime = now.getTime() - date.getTime();
        const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));
        
        if (diffDays === 0) {
            return 'Today';
        } else if (diffDays === 1) {
            return 'Yesterday';
        } else if (diffDays < 7) {
            return `${diffDays} days ago`;
        } else {
            return date.toLocaleDateString();
        }
    }

    formatDataQuality(quality) {
        const qualities = {
            'high': 'High Quality',
            'medium': 'Medium Quality',
            'low': 'Low Quality',
            'unknown': 'Unknown Quality'
        };
        return qualities[quality] || quality;
    }

    canSwitchToRealData() {
        // Check if user has permission to switch to real data
        // This would typically check user roles and environment settings
        return false; // For safety, default to false
    }

    show() {
        const warning = document.getElementById('mockDataWarning');
        if (warning) {
            warning.style.display = 'block';
            this.isVisible = true;
            
            // Add entrance animation
            warning.style.animation = 'slideInLeft 0.5s ease-out';
            
            // Start auto-hide timer if enabled
            if (this.autoHide) {
                this.startAutoHide();
            }
        }
    }

    hide() {
        const warning = document.getElementById('mockDataWarning');
        if (warning) {
            warning.classList.add('warning-closing');
            
            setTimeout(() => {
                warning.style.display = 'none';
                warning.classList.remove('warning-closing');
                this.isVisible = false;
                
                // Clear auto-hide timer
                this.stopAutoHide();
                
                // Call callback if provided
                if (this.onWarningDismiss) {
                    this.onWarningDismiss();
                }
            }, 300);
        }
    }

    startAutoHide() {
        this.stopAutoHide(); // Clear any existing timer
        this.hideTimer = setTimeout(() => {
            this.hide();
        }, this.autoHideDelay);
    }

    stopAutoHide() {
        if (this.hideTimer) {
            clearTimeout(this.hideTimer);
            this.hideTimer = null;
        }
    }

    handleViewDataSource() {
        console.log('View data source clicked');
        
        // Call callback if provided
        if (this.onDataSourceClick) {
            this.onDataSourceClick(this.dataInfo);
        }
        
        // Show data source details
        this.showDataSourceDetails();
    }

    handleSwitchToRealData() {
        console.log('Switch to real data clicked');
        
        // This would typically trigger a confirmation dialog
        // and then switch the data source
        this.showNotification('Switching to real data requires admin approval', 'warning');
    }

    handleDismissWarning() {
        console.log('Dismiss warning clicked');
        this.hide();
    }

    showDataSourceDetails() {
        const details = `
            Data Source Details:
            • Type: ${this.dataInfo?.type || 'Unknown'}
            • Records: ${this.dataInfo?.count || 'Unknown'}
            • Last Updated: ${this.dataInfo?.last_updated || 'Unknown'}
            • Quality: ${this.dataInfo?.quality || 'Unknown'}
            • Description: ${this.dataInfo?.description || 'No description available'}
        `;
        
        this.showNotification(details, 'info');
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${type === 'success' ? '#4CAF50' : type === 'error' ? '#f44336' : type === 'warning' ? '#ff9800' : '#2196F3'};
            color: white;
            padding: 16px 24px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
            z-index: 10001;
            animation: slideInRight 0.3s ease-out;
            max-width: 400px;
            text-align: left;
            white-space: pre-line;
            font-size: 0.9rem;
        `;
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }

    getAuthToken() {
        // Get auth token from localStorage or cookie
        return localStorage.getItem('auth_token') || 
               document.cookie.split('; ').find(row => row.startsWith('auth_token='))?.split('=')[1] || 
               '';
    }

    // Public methods for external control
    setDataSource(source) {
        this.dataSource = source;
        this.updateWarningContent();
    }

    setWarningLevel(level) {
        this.warningLevel = level;
        this.updateWarningContent();
    }

    setAutoHide(enabled, delay = null) {
        this.autoHide = enabled;
        if (delay !== null) {
            this.autoHideDelay = delay;
        }
        
        if (enabled && this.isVisible) {
            this.startAutoHide();
        } else {
            this.stopAutoHide();
        }
    }

    refresh() {
        this.loadDataSourceInfo();
    }

    toggle() {
        if (this.isVisible) {
            this.hide();
        } else {
            this.show();
        }
    }

    destroy() {
        this.stopAutoHide();
        if (this.container) {
            this.container.innerHTML = '';
        }
        this.isVisible = false;
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MockDataWarning;
}

// Auto-initialize if container is found
document.addEventListener('DOMContentLoaded', () => {
    const warningContainer = document.getElementById('mockDataWarningContainer');
    if (warningContainer && !window.mockDataWarning) {
        window.mockDataWarning = new MockDataWarning({
            container: warningContainer
        });
    }
});
