/**
 * Merchant Bulk Operations JavaScript
 * Provides comprehensive bulk operations interface with progress tracking and pause/resume functionality
 * Integrates with existing merchant portfolio system for seamless operations
 * 
 * Features:
 * - Multiple operation types (portfolio updates, risk level changes, exports, notifications, etc.)
 * - Real-time progress tracking with pause/resume capabilities
 * - Merchant selection with filtering and bulk selection
 * - Operation logging with detailed status updates
 * - Export functionality for operation results
 * - Responsive design with mobile support
 * - Integration with existing components
 * - Mock data support for MVP testing
 */

class MerchantBulkOperations {
    constructor() {
        this.apiBaseUrl = '/api/v1';
        this.currentOperation = null;
        this.selectedMerchants = new Set();
        this.merchants = [];
        this.operationState = {
            status: 'ready', // ready, running, paused, completed, failed
            progress: 0,
            completed: 0,
            failed: 0,
            total: 0,
            currentIndex: 0,
            operationId: null
        };
        this.operationLog = [];
        this.batchSize = 10; // Process merchants in batches
        this.batchDelay = 1000; // Delay between batches (ms)
        this.isProcessing = false;
        this.shouldPause = false;
        
        this.init();
    }

    /**
     * Initialize the bulk operations interface
     */
    init() {
        this.bindEvents();
        this.loadMerchants();
        this.updateUI();
        this.logOperation('Bulk operations interface initialized', 'info');
    }

    /**
     * Bind event listeners
     */
    bindEvents() {
        // Operation selection
        document.querySelectorAll('.operation-card').forEach(card => {
            card.addEventListener('click', (e) => this.selectOperation(e.currentTarget.dataset.operation));
        });

        // Merchant selection
        document.getElementById('selectAll').addEventListener('click', () => this.selectAllMerchants());
        document.getElementById('selectNone').addEventListener('click', () => this.selectNoMerchants());
        document.getElementById('selectByFilter').addEventListener('click', () => this.selectByFilter());
        document.getElementById('loadMore').addEventListener('click', () => this.loadMoreMerchants());

        // Operation controls
        document.getElementById('startOperation').addEventListener('click', () => this.startOperation());
        document.getElementById('pauseOperation').addEventListener('click', () => this.pauseOperation());
        document.getElementById('resumeOperation').addEventListener('click', () => this.resumeOperation());
        document.getElementById('stopOperation').addEventListener('click', () => this.stopOperation());

        // Export results
        document.getElementById('exportResults').addEventListener('click', () => this.exportResults());
    }

    /**
     * Load merchants for selection
     */
    async loadMerchants() {
        try {
            this.logOperation('Loading merchants...', 'info');
            
            // For MVP, use mock data
            const mockMerchants = this.generateMockMerchants(50);
            this.merchants = mockMerchants;
            
            this.renderMerchantList();
            this.updateSelectionStats();
            this.logOperation(`Loaded ${this.merchants.length} merchants`, 'success');
            
        } catch (error) {
            this.logOperation(`Failed to load merchants: ${error.message}`, 'error');
            console.error('Error loading merchants:', error);
        }
    }

    /**
     * Generate mock merchants for testing
     */
    generateMockMerchants(count) {
        const merchants = [];
        const businessTypes = ['Restaurant', 'Retail Store', 'Online Shop', 'Service Provider', 'Manufacturing', 'Consulting', 'Healthcare', 'Technology'];
        const portfolioTypes = ['onboarded', 'pending', 'deactivated', 'prospective'];
        const riskLevels = ['low', 'medium', 'high'];
        const cities = ['New York', 'Los Angeles', 'Chicago', 'Houston', 'Phoenix', 'Philadelphia', 'San Antonio', 'San Diego'];

        for (let i = 1; i <= count; i++) {
            const businessType = businessTypes[Math.floor(Math.random() * businessTypes.length)];
            const portfolioType = portfolioTypes[Math.floor(Math.random() * portfolioTypes.length)];
            const riskLevel = riskLevels[Math.floor(Math.random() * riskLevels.length)];
            const city = cities[Math.floor(Math.random() * cities.length)];

            merchants.push({
                id: `merchant_${i}`,
                name: `${businessType} ${i}`,
                businessType: businessType,
                address: `${Math.floor(Math.random() * 9999) + 1} Main St, ${city}, ST ${Math.floor(Math.random() * 99999) + 10000}`,
                portfolioType: portfolioType,
                riskLevel: riskLevel,
                email: `contact@${businessType.toLowerCase().replace(/\s+/g, '')}${i}.com`,
                phone: `+1-${Math.floor(Math.random() * 900) + 100}-${Math.floor(Math.random() * 900) + 100}-${Math.floor(Math.random() * 9000) + 1000}`,
                website: `https://www.${businessType.toLowerCase().replace(/\s+/g, '')}${i}.com`,
                createdAt: new Date(Date.now() - Math.random() * 365 * 24 * 60 * 60 * 1000).toISOString(),
                lastUpdated: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString()
            });
        }

        return merchants;
    }

    /**
     * Render merchant list
     */
    renderMerchantList() {
        const merchantList = document.getElementById('merchantList');
        merchantList.innerHTML = '';

        this.merchants.forEach(merchant => {
            const merchantItem = this.createMerchantItem(merchant);
            merchantList.appendChild(merchantItem);
        });
    }

    /**
     * Create merchant list item
     */
    createMerchantItem(merchant) {
        const item = document.createElement('div');
        item.className = 'merchant-item';
        item.dataset.merchantId = merchant.id;

        const initials = merchant.name.split(' ').map(word => word[0]).join('').toUpperCase();
        const portfolioBadge = this.createBadge(merchant.portfolioType, 'portfolio');
        const riskBadge = this.createBadge(merchant.riskLevel, 'risk');

        item.innerHTML = `
            <input type="checkbox" class="merchant-checkbox" data-merchant-id="${merchant.id}">
            <div class="merchant-info">
                <div class="merchant-avatar">${initials}</div>
                <div class="merchant-details">
                    <h4>${merchant.name}</h4>
                    <p>${merchant.businessType} • ${merchant.address}</p>
                </div>
                <div class="merchant-badges">
                    ${portfolioBadge}
                    ${riskBadge}
                </div>
            </div>
        `;

        // Bind checkbox event
        const checkbox = item.querySelector('.merchant-checkbox');
        checkbox.addEventListener('change', (e) => this.toggleMerchantSelection(merchant.id, e.target.checked));

        return item;
    }

    /**
     * Create badge element
     */
    createBadge(value, type) {
        const badgeClass = type === 'portfolio' ? `badge-${value}` : `badge-${value}-risk`;
        const displayValue = value.charAt(0).toUpperCase() + value.slice(1);
        return `<span class="badge ${badgeClass}">${displayValue}</span>`;
    }

    /**
     * Toggle merchant selection
     */
    toggleMerchantSelection(merchantId, selected) {
        if (selected) {
            this.selectedMerchants.add(merchantId);
        } else {
            this.selectedMerchants.delete(merchantId);
        }
        
        this.updateSelectionStats();
        this.updateOperationControls();
    }

    /**
     * Select all merchants
     */
    selectAllMerchants() {
        this.selectedMerchants.clear();
        this.merchants.forEach(merchant => {
            this.selectedMerchants.add(merchant.id);
        });
        
        this.updateCheckboxes();
        this.updateSelectionStats();
        this.updateOperationControls();
        this.logOperation(`Selected all ${this.merchants.length} merchants`, 'info');
    }

    /**
     * Select no merchants
     */
    selectNoMerchants() {
        this.selectedMerchants.clear();
        this.updateCheckboxes();
        this.updateSelectionStats();
        this.updateOperationControls();
        this.logOperation('Deselected all merchants', 'info');
    }

    /**
     * Select merchants by filter
     */
    selectByFilter() {
        // For MVP, select merchants with specific criteria
        this.selectedMerchants.clear();
        this.merchants.forEach(merchant => {
            if (merchant.portfolioType === 'pending' || merchant.riskLevel === 'high') {
                this.selectedMerchants.add(merchant.id);
            }
        });
        
        this.updateCheckboxes();
        this.updateSelectionStats();
        this.updateOperationControls();
        this.logOperation(`Selected ${this.selectedMerchants.size} merchants by filter`, 'info');
    }

    /**
     * Load more merchants
     */
    async loadMoreMerchants() {
        try {
            this.logOperation('Loading more merchants...', 'info');
            
            // Generate additional mock merchants
            const additionalMerchants = this.generateMockMerchants(25);
            this.merchants.push(...additionalMerchants);
            
            this.renderMerchantList();
            this.updateSelectionStats();
            this.logOperation(`Loaded ${additionalMerchants.length} additional merchants`, 'success');
            
        } catch (error) {
            this.logOperation(`Failed to load more merchants: ${error.message}`, 'error');
            console.error('Error loading more merchants:', error);
        }
    }

    /**
     * Update checkboxes to reflect selection state
     */
    updateCheckboxes() {
        document.querySelectorAll('.merchant-checkbox').forEach(checkbox => {
            const merchantId = checkbox.dataset.merchantId;
            checkbox.checked = this.selectedMerchants.has(merchantId);
        });
    }

    /**
     * Update selection statistics
     */
    updateSelectionStats() {
        document.getElementById('totalMerchants').textContent = this.merchants.length;
        document.getElementById('selectedMerchants').textContent = this.selectedMerchants.size;
    }

    /**
     * Select operation type
     */
    selectOperation(operationType) {
        // Remove previous selection
        document.querySelectorAll('.operation-card').forEach(card => {
            card.classList.remove('selected');
        });

        // Select new operation
        const selectedCard = document.querySelector(`[data-operation="${operationType}"]`);
        if (selectedCard) {
            selectedCard.classList.add('selected');
            this.currentOperation = operationType;
            this.showOperationConfig(operationType);
            this.updateOperationControls();
            this.logOperation(`Selected operation: ${operationType}`, 'info');
        }
    }

    /**
     * Show operation configuration
     */
    showOperationConfig(operationType) {
        const configSection = document.getElementById('operationConfig');
        const configContent = document.getElementById('configContent');
        
        let configHTML = '';
        
        switch (operationType) {
            case 'update-portfolio':
                configHTML = this.getPortfolioUpdateConfig();
                break;
            case 'update-risk':
                configHTML = this.getRiskUpdateConfig();
                break;
            case 'export-data':
                configHTML = this.getExportConfig();
                break;
            case 'send-notifications':
                configHTML = this.getNotificationConfig();
                break;
            case 'schedule-review':
                configHTML = this.getReviewScheduleConfig();
                break;
            case 'bulk-deactivate':
                configHTML = this.getDeactivationConfig();
                break;
        }
        
        configContent.innerHTML = configHTML;
        configSection.style.display = 'block';
    }

    /**
     * Get portfolio update configuration
     */
    getPortfolioUpdateConfig() {
        return `
            <div class="config-group">
                <label for="newPortfolioType">New Portfolio Type:</label>
                <select id="newPortfolioType" class="form-control">
                    <option value="onboarded">Onboarded</option>
                    <option value="pending">Pending</option>
                    <option value="deactivated">Deactivated</option>
                    <option value="prospective">Prospective</option>
                </select>
            </div>
            <div class="config-group">
                <label for="updateReason">Reason for Update:</label>
                <textarea id="updateReason" class="form-control" placeholder="Enter reason for portfolio type change..."></textarea>
            </div>
        `;
    }

    /**
     * Get risk update configuration
     */
    getRiskUpdateConfig() {
        return `
            <div class="config-group">
                <label for="newRiskLevel">New Risk Level:</label>
                <select id="newRiskLevel" class="form-control">
                    <option value="low">Low Risk</option>
                    <option value="medium">Medium Risk</option>
                    <option value="high">High Risk</option>
                </select>
            </div>
            <div class="config-group">
                <label for="riskAssessment">Risk Assessment Notes:</label>
                <textarea id="riskAssessment" class="form-control" placeholder="Enter risk assessment details..."></textarea>
            </div>
        `;
    }

    /**
     * Get export configuration
     */
    getExportConfig() {
        return `
            <div class="config-group">
                <label for="exportFormat">Export Format:</label>
                <select id="exportFormat" class="form-control">
                    <option value="csv">CSV</option>
                    <option value="excel">Excel</option>
                    <option value="json">JSON</option>
                </select>
            </div>
            <div class="config-group">
                <label for="exportFields">Fields to Export:</label>
                <div class="checkbox-group">
                    <label><input type="checkbox" checked> Basic Info</label>
                    <label><input type="checkbox" checked> Portfolio Type</label>
                    <label><input type="checkbox" checked> Risk Level</label>
                    <label><input type="checkbox"> Contact Details</label>
                    <label><input type="checkbox"> Audit Log</label>
                </div>
            </div>
        `;
    }

    /**
     * Get notification configuration
     */
    getNotificationConfig() {
        return `
            <div class="config-group">
                <label for="notificationType">Notification Type:</label>
                <select id="notificationType" class="form-control">
                    <option value="email">Email</option>
                    <option value="sms">SMS</option>
                    <option value="both">Email & SMS</option>
                </select>
            </div>
            <div class="config-group">
                <label for="notificationMessage">Message:</label>
                <textarea id="notificationMessage" class="form-control" placeholder="Enter notification message..."></textarea>
            </div>
        `;
    }

    /**
     * Get review schedule configuration
     */
    getReviewScheduleConfig() {
        return `
            <div class="config-group">
                <label for="reviewDate">Review Date:</label>
                <input type="date" id="reviewDate" class="form-control">
            </div>
            <div class="config-group">
                <label for="reviewType">Review Type:</label>
                <select id="reviewType" class="form-control">
                    <option value="compliance">Compliance Review</option>
                    <option value="risk">Risk Assessment</option>
                    <option value="kyc">KYC Verification</option>
                </select>
            </div>
        `;
    }

    /**
     * Get deactivation configuration
     */
    getDeactivationConfig() {
        return `
            <div class="config-group">
                <label for="deactivationReason">Deactivation Reason:</label>
                <select id="deactivationReason" class="form-control">
                    <option value="compliance">Compliance Issues</option>
                    <option value="risk">High Risk</option>
                    <option value="request">Merchant Request</option>
                    <option value="other">Other</option>
                </select>
            </div>
            <div class="config-group">
                <label for="deactivationNotes">Additional Notes:</label>
                <textarea id="deactivationNotes" class="form-control" placeholder="Enter additional notes..."></textarea>
            </div>
            <div class="config-group">
                <label>
                    <input type="checkbox" id="sendNotification"> Send notification to merchants
                </label>
            </div>
        `;
    }

    /**
     * Update operation controls based on current state
     */
    updateOperationControls() {
        const hasSelection = this.selectedMerchants.size > 0;
        const hasOperation = this.currentOperation !== null;
        const canStart = hasSelection && hasOperation && this.operationState.status === 'ready';
        const canPause = this.operationState.status === 'running';
        const canResume = this.operationState.status === 'paused';
        const canStop = ['running', 'paused'].includes(this.operationState.status);

        document.getElementById('startOperation').disabled = !canStart;
        document.getElementById('pauseOperation').disabled = !canPause;
        document.getElementById('resumeOperation').disabled = !canResume;
        document.getElementById('stopOperation').disabled = !canStop;
    }

    /**
     * Start bulk operation
     */
    async startOperation() {
        if (!this.currentOperation || this.selectedMerchants.size === 0) {
            this.logOperation('Cannot start operation: No operation selected or no merchants selected', 'error');
            return;
        }

        try {
            this.operationState = {
                status: 'running',
                progress: 0,
                completed: 0,
                failed: 0,
                total: this.selectedMerchants.size,
                currentIndex: 0,
                operationId: `op_${Date.now()}`
            };

            this.isProcessing = true;
            this.shouldPause = false;

            this.updateUI();
            this.logOperation(`Started ${this.currentOperation} operation on ${this.selectedMerchants.size} merchants`, 'info');

            // Process merchants in batches
            await this.processMerchantsInBatches();

            if (this.operationState.status === 'running') {
                this.operationState.status = 'completed';
                this.logOperation('Operation completed successfully', 'success');
            }

        } catch (error) {
            this.operationState.status = 'failed';
            this.logOperation(`Operation failed: ${error.message}`, 'error');
            console.error('Operation error:', error);
        } finally {
            this.isProcessing = false;
            this.updateUI();
        }
    }

    /**
     * Process merchants in batches
     */
    async processMerchantsInBatches() {
        const merchantIds = Array.from(this.selectedMerchants);
        const batches = this.chunkArray(merchantIds, this.batchSize);

        for (let i = 0; i < batches.length; i++) {
            if (this.shouldPause) {
                this.operationState.status = 'paused';
                this.logOperation('Operation paused', 'warning');
                break;
            }

            const batch = batches[i];
            await this.processBatch(batch, i + 1, batches.length);
            
            // Delay between batches
            if (i < batches.length - 1) {
                await this.delay(this.batchDelay);
            }
        }
    }

    /**
     * Process a batch of merchants
     */
    async processBatch(merchantIds, batchNumber, totalBatches) {
        this.logOperation(`Processing batch ${batchNumber}/${totalBatches} (${merchantIds.length} merchants)`, 'info');

        for (const merchantId of merchantIds) {
            if (this.shouldPause) {
                break;
            }

            try {
                await this.processMerchant(merchantId);
                this.operationState.completed++;
            } catch (error) {
                this.operationState.failed++;
                this.logOperation(`Failed to process merchant ${merchantId}: ${error.message}`, 'error');
            }

            this.operationState.currentIndex++;
            this.operationState.progress = (this.operationState.currentIndex / this.operationState.total) * 100;
            this.updateProgressUI();
        }
    }

    /**
     * Process individual merchant
     */
    async processMerchant(merchantId) {
        const merchant = this.merchants.find(m => m.id === merchantId);
        if (!merchant) {
            throw new Error('Merchant not found');
        }

        // Simulate processing time
        await this.delay(Math.random() * 500 + 200);

        // Simulate operation based on type
        switch (this.currentOperation) {
            case 'update-portfolio':
                await this.simulatePortfolioUpdate(merchant);
                break;
            case 'update-risk':
                await this.simulateRiskUpdate(merchant);
                break;
            case 'export-data':
                await this.simulateDataExport(merchant);
                break;
            case 'send-notifications':
                await this.simulateNotification(merchant);
                break;
            case 'schedule-review':
                await this.simulateReviewSchedule(merchant);
                break;
            case 'bulk-deactivate':
                await this.simulateDeactivation(merchant);
                break;
        }

        this.logOperation(`Processed merchant: ${merchant.name}`, 'success');
    }

    /**
     * Simulate portfolio update
     */
    async simulatePortfolioUpdate(merchant) {
        const newPortfolioType = document.getElementById('newPortfolioType')?.value || 'onboarded';
        merchant.portfolioType = newPortfolioType;
        merchant.lastUpdated = new Date().toISOString();
    }

    /**
     * Simulate risk update
     */
    async simulateRiskUpdate(merchant) {
        const newRiskLevel = document.getElementById('newRiskLevel')?.value || 'medium';
        merchant.riskLevel = newRiskLevel;
        merchant.lastUpdated = new Date().toISOString();
    }

    /**
     * Simulate data export
     */
    async simulateDataExport(merchant) {
        // Simulate export processing
        const exportFormat = document.getElementById('exportFormat')?.value || 'csv';
        // In real implementation, this would generate actual export data
    }

    /**
     * Simulate notification sending
     */
    async simulateNotification(merchant) {
        const notificationType = document.getElementById('notificationType')?.value || 'email';
        const message = document.getElementById('notificationMessage')?.value || 'Default notification';
        // In real implementation, this would send actual notifications
    }

    /**
     * Simulate review scheduling
     */
    async simulateReviewSchedule(merchant) {
        const reviewDate = document.getElementById('reviewDate')?.value || new Date().toISOString().split('T')[0];
        const reviewType = document.getElementById('reviewType')?.value || 'compliance';
        // In real implementation, this would schedule actual reviews
    }

    /**
     * Simulate deactivation
     */
    async simulateDeactivation(merchant) {
        const reason = document.getElementById('deactivationReason')?.value || 'other';
        merchant.portfolioType = 'deactivated';
        merchant.lastUpdated = new Date().toISOString();
    }

    /**
     * Pause operation
     */
    pauseOperation() {
        this.shouldPause = true;
        this.logOperation('Operation pause requested', 'warning');
    }

    /**
     * Resume operation
     */
    async resumeOperation() {
        this.shouldPause = false;
        this.operationState.status = 'running';
        this.updateUI();
        this.logOperation('Operation resumed', 'info');

        // Continue processing from where we left off
        if (this.operationState.currentIndex < this.operationState.total) {
            await this.processMerchantsInBatches();
        }
    }

    /**
     * Stop operation
     */
    stopOperation() {
        this.isProcessing = false;
        this.shouldPause = true;
        this.operationState.status = 'failed';
        this.updateUI();
        this.logOperation('Operation stopped by user', 'warning');
    }

    /**
     * Update progress UI
     */
    updateProgressUI() {
        document.getElementById('progressCompleted').textContent = this.operationState.completed;
        document.getElementById('progressFailed').textContent = this.operationState.failed;
        document.getElementById('progressPercentage').textContent = `${Math.round(this.operationState.progress)}%`;
        document.getElementById('progressBarFill').style.width = `${this.operationState.progress}%`;
        
        const progressText = this.operationState.status === 'running' 
            ? `Processing ${this.operationState.currentIndex}/${this.operationState.total} merchants`
            : this.operationState.status === 'paused'
            ? 'Operation paused'
            : this.operationState.status === 'completed'
            ? 'Operation completed'
            : this.operationState.status === 'failed'
            ? 'Operation failed'
            : 'Ready to start';
            
        document.getElementById('progressText').textContent = progressText;
    }

    /**
     * Update overall UI
     */
    updateUI() {
        this.updateProgressUI();
        this.updateOperationControls();
        this.updateStatusIndicator();
    }

    /**
     * Update status indicator
     */
    updateStatusIndicator() {
        const statusIndicator = document.getElementById('operationStatus');
        const status = this.operationState.status;
        
        statusIndicator.className = `status-indicator status-${status}`;
        
        const statusText = {
            'ready': 'Ready',
            'running': 'Running',
            'paused': 'Paused',
            'completed': 'Completed',
            'failed': 'Failed'
        };
        
        statusIndicator.innerHTML = `<i class="fas fa-circle"></i> ${statusText[status]}`;
    }

    /**
     * Log operation event
     */
    logOperation(message, type = 'info') {
        const timestamp = new Date().toLocaleString();
        const logEntry = {
            timestamp,
            message,
            type
        };
        
        this.operationLog.push(logEntry);
        this.renderLogEntry(logEntry);
    }

    /**
     * Render log entry
     */
    renderLogEntry(logEntry) {
        const logContainer = document.getElementById('operationLog');
        const logItem = document.createElement('div');
        logItem.className = 'log-entry';
        
        const iconClass = {
            'success': 'fas fa-check',
            'error': 'fas fa-times',
            'warning': 'fas fa-exclamation-triangle',
            'info': 'fas fa-info'
        }[logEntry.type] || 'fas fa-info';
        
        logItem.innerHTML = `
            <div class="log-icon ${logEntry.type}">
                <i class="${iconClass}"></i>
            </div>
            <div class="log-content">
                <div class="log-message">${logEntry.message}</div>
                <div class="log-timestamp">${logEntry.timestamp}</div>
            </div>
        `;
        
        logContainer.appendChild(logItem);
        logContainer.scrollTop = logContainer.scrollHeight;
    }

    /**
     * Export operation results
     */
    exportResults() {
        if (this.operationLog.length === 0) {
            this.logOperation('No results to export', 'warning');
            return;
        }

        try {
            const results = {
                operation: this.currentOperation,
                timestamp: new Date().toISOString(),
                summary: {
                    total: this.operationState.total,
                    completed: this.operationState.completed,
                    failed: this.operationState.failed,
                    status: this.operationState.status
                },
                log: this.operationLog
            };

            const dataStr = JSON.stringify(results, null, 2);
            const dataBlob = new Blob([dataStr], { type: 'application/json' });
            
            const link = document.createElement('a');
            link.href = URL.createObjectURL(dataBlob);
            link.download = `bulk-operation-results-${Date.now()}.json`;
            link.click();
            
            this.logOperation('Results exported successfully', 'success');
            
        } catch (error) {
            this.logOperation(`Failed to export results: ${error.message}`, 'error');
            console.error('Export error:', error);
        }
    }

    /**
     * Utility function to chunk array
     */
    chunkArray(array, chunkSize) {
        const chunks = [];
        for (let i = 0; i < array.length; i += chunkSize) {
            chunks.push(array.slice(i, i + chunkSize));
        }
        return chunks;
    }

    /**
     * Utility function to delay execution
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    new MerchantBulkOperations();
});
