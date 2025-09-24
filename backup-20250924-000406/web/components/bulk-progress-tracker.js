/**
 * Bulk Progress Tracker Component
 * 
 * Provides real-time progress tracking for bulk operations with:
 * - Progress bar visualization
 * - Operation status management
 * - Real-time updates via WebSocket or polling
 * - Pause/resume functionality
 * - Error handling and retry logic
 * - Export capabilities for operation results
 */

class BulkProgressTracker {
    constructor(options = {}) {
        this.options = {
            updateInterval: 1000, // Update every second
            maxRetries: 3,
            retryDelay: 2000,
            autoRefresh: true,
            showDetailedLog: true,
            enablePauseResume: true,
            enableExport: true,
            ...options
        };

        this.state = {
            operationId: null,
            status: 'ready', // ready, running, paused, completed, failed, cancelled
            progress: 0,
            totalItems: 0,
            completedItems: 0,
            failedItems: 0,
            currentItem: null,
            startTime: null,
            endTime: null,
            estimatedTimeRemaining: null,
            errors: [],
            results: []
        };

        this.elements = {};
        this.updateTimer = null;
        this.retryCount = 0;
        this.eventListeners = new Map();

        this.init();
    }

    /**
     * Initialize the progress tracker
     */
    init() {
        this.bindElements();
        this.bindEvents();
        this.updateDisplay();
        this.log('Progress tracker initialized', 'info');
    }

    /**
     * Bind DOM elements
     */
    bindElements() {
        this.elements = {
            progressBar: document.getElementById('progressBarFill'),
            progressText: document.getElementById('progressText'),
            progressPercentage: document.getElementById('progressPercentage'),
            progressCompleted: document.getElementById('progressCompleted'),
            progressFailed: document.getElementById('progressFailed'),
            operationStatus: document.getElementById('operationStatus'),
            startButton: document.getElementById('startOperation'),
            pauseButton: document.getElementById('pauseOperation'),
            resumeButton: document.getElementById('resumeOperation'),
            stopButton: document.getElementById('stopOperation'),
            exportButton: document.getElementById('exportResults'),
            operationLog: document.getElementById('operationLog')
        };

        // Validate required elements
        const requiredElements = ['progressBar', 'progressText', 'progressPercentage', 'operationStatus'];
        for (const elementName of requiredElements) {
            if (!this.elements[elementName]) {
                console.warn(`Required element not found: ${elementName}`);
            }
        }
    }

    /**
     * Bind event listeners
     */
    bindEvents() {
        // Control buttons
        if (this.elements.startButton) {
            this.elements.startButton.addEventListener('click', () => this.startOperation());
        }
        
        if (this.elements.pauseButton) {
            this.elements.pauseButton.addEventListener('click', () => this.pauseOperation());
        }
        
        if (this.elements.resumeButton) {
            this.elements.resumeButton.addEventListener('click', () => this.resumeOperation());
        }
        
        if (this.elements.stopButton) {
            this.elements.stopButton.addEventListener('click', () => this.stopOperation());
        }
        
        if (this.elements.exportButton) {
            this.elements.exportButton.addEventListener('click', () => this.exportResults());
        }

        // Custom events
        this.addEventListener('progressUpdate', (data) => this.handleProgressUpdate(data));
        this.addEventListener('operationComplete', (data) => this.handleOperationComplete(data));
        this.addEventListener('operationError', (data) => this.handleOperationError(data));
    }

    /**
     * Start a new bulk operation
     */
    async startOperation(operationData = {}) {
        try {
            this.log('Starting bulk operation', 'info');
            
            // Reset state
            this.resetState();
            
            // Set operation data
            this.state.operationId = operationData.operationId || this.generateOperationId();
            this.state.totalItems = operationData.totalItems || 0;
            this.state.status = 'running';
            this.state.startTime = new Date();
            
            // Update UI
            this.updateDisplay();
            this.updateControlButtons();
            
            // Start progress monitoring
            this.startProgressMonitoring();
            
            // Emit start event
            this.emit('operationStart', {
                operationId: this.state.operationId,
                totalItems: this.state.totalItems,
                startTime: this.state.startTime
            });
            
            this.log(`Operation started: ${this.state.operationId}`, 'success');
            
        } catch (error) {
            this.handleError('Failed to start operation', error);
        }
    }

    /**
     * Pause the current operation
     */
    pauseOperation() {
        if (this.state.status !== 'running') {
            this.log('Cannot pause operation - not currently running', 'warning');
            return;
        }

        this.state.status = 'paused';
        this.updateDisplay();
        this.updateControlButtons();
        this.log('Operation paused', 'warning');
        
        this.emit('operationPause', {
            operationId: this.state.operationId,
            progress: this.state.progress
        });
    }

    /**
     * Resume the paused operation
     */
    resumeOperation() {
        if (this.state.status !== 'paused') {
            this.log('Cannot resume operation - not currently paused', 'warning');
            return;
        }

        this.state.status = 'running';
        this.updateDisplay();
        this.updateControlButtons();
        this.log('Operation resumed', 'info');
        
        this.emit('operationResume', {
            operationId: this.state.operationId,
            progress: this.state.progress
        });
    }

    /**
     * Stop the current operation
     */
    stopOperation() {
        if (this.state.status === 'ready' || this.state.status === 'completed') {
            this.log('No operation to stop', 'warning');
            return;
        }

        this.state.status = 'cancelled';
        this.state.endTime = new Date();
        this.updateDisplay();
        this.updateControlButtons();
        this.stopProgressMonitoring();
        
        this.log('Operation cancelled', 'warning');
        
        this.emit('operationCancel', {
            operationId: this.state.operationId,
            progress: this.state.progress,
            endTime: this.state.endTime
        });
    }

    /**
     * Update progress for the current operation
     */
    updateProgress(progressData) {
        const {
            completed = 0,
            failed = 0,
            currentItem = null,
            errors = []
        } = progressData;

        // Update state
        this.state.completedItems = completed;
        this.state.failedItems = failed;
        this.state.currentItem = currentItem;
        this.state.errors = [...this.state.errors, ...errors];
        
        // Calculate progress percentage
        const totalProcessed = completed + failed;
        this.state.progress = this.state.totalItems > 0 ? 
            Math.round((totalProcessed / this.state.totalItems) * 100) : 0;
        
        // Calculate estimated time remaining
        this.calculateEstimatedTime();
        
        // Update display
        this.updateDisplay();
        
        // Check if operation is complete
        if (totalProcessed >= this.state.totalItems && this.state.status === 'running') {
            this.completeOperation();
        }
        
        this.emit('progressUpdate', {
            operationId: this.state.operationId,
            progress: this.state.progress,
            completed,
            failed,
            currentItem
        });
    }

    /**
     * Complete the current operation
     */
    completeOperation() {
        this.state.status = 'completed';
        this.state.endTime = new Date();
        this.state.progress = 100;
        
        this.updateDisplay();
        this.updateControlButtons();
        this.stopProgressMonitoring();
        
        const duration = this.state.endTime - this.state.startTime;
        this.log(`Operation completed in ${this.formatDuration(duration)}`, 'success');
        
        this.emit('operationComplete', {
            operationId: this.state.operationId,
            duration,
            completed: this.state.completedItems,
            failed: this.state.failedItems,
            results: this.state.results
        });
    }

    /**
     * Handle operation error
     */
    handleOperationError(error) {
        this.state.status = 'failed';
        this.state.endTime = new Date();
        this.state.errors.push(error);
        
        this.updateDisplay();
        this.updateControlButtons();
        this.stopProgressMonitoring();
        
        this.log(`Operation failed: ${error.message}`, 'error');
        
        this.emit('operationError', {
            operationId: this.state.operationId,
            error,
            progress: this.state.progress
        });
    }

    /**
     * Start progress monitoring
     */
    startProgressMonitoring() {
        if (this.updateTimer) {
            clearInterval(this.updateTimer);
        }
        
        this.updateTimer = setInterval(() => {
            this.fetchProgressUpdate();
        }, this.options.updateInterval);
    }

    /**
     * Stop progress monitoring
     */
    stopProgressMonitoring() {
        if (this.updateTimer) {
            clearInterval(this.updateTimer);
            this.updateTimer = null;
        }
    }

    /**
     * Fetch progress update from server
     */
    async fetchProgressUpdate() {
        if (!this.state.operationId || this.state.status !== 'running') {
            return;
        }

        try {
            const response = await fetch(`/api/bulk-operations/${this.state.operationId}/progress`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.updateProgress(data);
            this.retryCount = 0; // Reset retry count on success
            
        } catch (error) {
            this.handleProgressError(error);
        }
    }

    /**
     * Handle progress fetch error
     */
    handleProgressError(error) {
        this.retryCount++;
        
        if (this.retryCount <= this.options.maxRetries) {
            this.log(`Progress fetch failed, retrying (${this.retryCount}/${this.options.maxRetries})`, 'warning');
            setTimeout(() => {
                this.fetchProgressUpdate();
            }, this.options.retryDelay);
        } else {
            this.handleOperationError(new Error('Failed to fetch progress updates'));
        }
    }

    /**
     * Calculate estimated time remaining
     */
    calculateEstimatedTime() {
        if (!this.state.startTime || this.state.completedItems === 0) {
            this.state.estimatedTimeRemaining = null;
            return;
        }

        const elapsed = Date.now() - this.state.startTime.getTime();
        const rate = this.state.completedItems / elapsed; // items per millisecond
        const remaining = this.state.totalItems - this.state.completedItems - this.state.failedItems;
        
        if (rate > 0 && remaining > 0) {
            this.state.estimatedTimeRemaining = Math.round(remaining / rate);
        }
    }

    /**
     * Update the display
     */
    updateDisplay() {
        // Update progress bar
        if (this.elements.progressBar) {
            this.elements.progressBar.style.width = `${this.state.progress}%`;
        }
        
        // Update progress text
        if (this.elements.progressText) {
            this.elements.progressText.textContent = this.getProgressText();
        }
        
        // Update progress percentage
        if (this.elements.progressPercentage) {
            this.elements.progressPercentage.textContent = `${this.state.progress}%`;
        }
        
        // Update completed count
        if (this.elements.progressCompleted) {
            this.elements.progressCompleted.textContent = this.state.completedItems;
        }
        
        // Update failed count
        if (this.elements.progressFailed) {
            this.elements.progressFailed.textContent = this.state.failedItems;
        }
        
        // Update operation status
        if (this.elements.operationStatus) {
            this.elements.operationStatus.className = `status-indicator status-${this.state.status}`;
            this.elements.operationStatus.innerHTML = `
                <i class="fas fa-circle"></i>
                ${this.getStatusText()}
            `;
        }
        
        // Update control buttons
        this.updateControlButtons();
    }

    /**
     * Update control buttons based on current state
     */
    updateControlButtons() {
        const buttons = {
            start: this.elements.startButton,
            pause: this.elements.pauseButton,
            resume: this.elements.resumeButton,
            stop: this.elements.stopButton,
            export: this.elements.exportButton
        };

        // Reset all buttons
        Object.values(buttons).forEach(btn => {
            if (btn) {
                btn.disabled = true;
                btn.style.display = 'none';
            }
        });

        // Configure buttons based on status
        switch (this.state.status) {
            case 'ready':
                if (buttons.start) {
                    buttons.start.disabled = false;
                    buttons.start.style.display = 'inline-flex';
                }
                break;
                
            case 'running':
                if (buttons.pause && this.options.enablePauseResume) {
                    buttons.pause.disabled = false;
                    buttons.pause.style.display = 'inline-flex';
                }
                if (buttons.stop) {
                    buttons.stop.disabled = false;
                    buttons.stop.style.display = 'inline-flex';
                }
                break;
                
            case 'paused':
                if (buttons.resume && this.options.enablePauseResume) {
                    buttons.resume.disabled = false;
                    buttons.resume.style.display = 'inline-flex';
                }
                if (buttons.stop) {
                    buttons.stop.disabled = false;
                    buttons.stop.style.display = 'inline-flex';
                }
                break;
                
            case 'completed':
            case 'failed':
            case 'cancelled':
                if (buttons.export && this.options.enableExport) {
                    buttons.export.disabled = false;
                    buttons.export.style.display = 'inline-flex';
                }
                break;
        }
    }

    /**
     * Get progress text based on current state
     */
    getProgressText() {
        switch (this.state.status) {
            case 'ready':
                return 'Ready to start';
            case 'running':
                if (this.state.currentItem) {
                    return `Processing: ${this.state.currentItem}`;
                }
                return `Processing ${this.state.completedItems + this.state.failedItems} of ${this.state.totalItems}`;
            case 'paused':
                return 'Operation paused';
            case 'completed':
                return `Completed: ${this.state.completedItems} items processed`;
            case 'failed':
                return `Failed: ${this.state.errors.length} errors`;
            case 'cancelled':
                return 'Operation cancelled';
            default:
                return 'Unknown status';
        }
    }

    /**
     * Get status text
     */
    getStatusText() {
        const statusMap = {
            'ready': 'Ready',
            'running': 'Running',
            'paused': 'Paused',
            'completed': 'Completed',
            'failed': 'Failed',
            'cancelled': 'Cancelled'
        };
        return statusMap[this.state.status] || 'Unknown';
    }

    /**
     * Export operation results
     */
    async exportResults() {
        if (!this.state.operationId) {
            this.log('No operation results to export', 'warning');
            return;
        }

        try {
            this.log('Exporting operation results', 'info');
            
            const response = await fetch(`/api/bulk-operations/${this.state.operationId}/export`);
            if (!response.ok) {
                throw new Error(`Export failed: ${response.statusText}`);
            }
            
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `bulk-operation-${this.state.operationId}-results.csv`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            this.log('Results exported successfully', 'success');
            
        } catch (error) {
            this.handleError('Failed to export results', error);
        }
    }

    /**
     * Reset state for new operation
     */
    resetState() {
        this.state = {
            operationId: null,
            status: 'ready',
            progress: 0,
            totalItems: 0,
            completedItems: 0,
            failedItems: 0,
            currentItem: null,
            startTime: null,
            endTime: null,
            estimatedTimeRemaining: null,
            errors: [],
            results: []
        };
        this.retryCount = 0;
    }

    /**
     * Generate unique operation ID
     */
    generateOperationId() {
        return `bulk-op-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    }

    /**
     * Format duration in human-readable format
     */
    formatDuration(milliseconds) {
        const seconds = Math.floor(milliseconds / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        
        if (hours > 0) {
            return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
        } else if (minutes > 0) {
            return `${minutes}m ${seconds % 60}s`;
        } else {
            return `${seconds}s`;
        }
    }

    /**
     * Add log entry
     */
    log(message, type = 'info') {
        if (!this.options.showDetailedLog) return;
        
        const timestamp = new Date().toLocaleString();
        const logEntry = {
            message,
            type,
            timestamp
        };
        
        // Add to operation log if available
        if (this.elements.operationLog) {
            this.addLogEntry(logEntry);
        }
        
        // Emit log event
        this.emit('log', logEntry);
    }

    /**
     * Add log entry to the UI
     */
    addLogEntry(logEntry) {
        const logContainer = this.elements.operationLog;
        if (!logContainer) return;
        
        const entryElement = document.createElement('div');
        entryElement.className = 'log-entry';
        
        const iconClass = {
            'success': 'fas fa-check',
            'error': 'fas fa-times',
            'warning': 'fas fa-exclamation-triangle',
            'info': 'fas fa-info'
        }[logEntry.type] || 'fas fa-info';
        
        entryElement.innerHTML = `
            <div class="log-icon ${logEntry.type}">
                <i class="${iconClass}"></i>
            </div>
            <div class="log-content">
                <div class="log-message">${logEntry.message}</div>
                <div class="log-timestamp">${logEntry.timestamp}</div>
            </div>
        `;
        
        logContainer.appendChild(entryElement);
        logContainer.scrollTop = logContainer.scrollHeight;
    }

    /**
     * Handle errors
     */
    handleError(message, error) {
        console.error(message, error);
        this.log(`${message}: ${error.message}`, 'error');
        this.emit('error', { message, error });
    }

    /**
     * Add event listener
     */
    addEventListener(event, callback) {
        if (!this.eventListeners.has(event)) {
            this.eventListeners.set(event, []);
        }
        this.eventListeners.get(event).push(callback);
    }

    /**
     * Remove event listener
     */
    removeEventListener(event, callback) {
        if (this.eventListeners.has(event)) {
            const listeners = this.eventListeners.get(event);
            const index = listeners.indexOf(callback);
            if (index > -1) {
                listeners.splice(index, 1);
            }
        }
    }

    /**
     * Emit event
     */
    emit(event, data) {
        if (this.eventListeners.has(event)) {
            this.eventListeners.get(event).forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    console.error(`Error in event listener for ${event}:`, error);
                }
            });
        }
    }

    /**
     * Get current state
     */
    getState() {
        return { ...this.state };
    }

    /**
     * Update options
     */
    updateOptions(newOptions) {
        this.options = { ...this.options, ...newOptions };
    }

    /**
     * Destroy the component
     */
    destroy() {
        this.stopProgressMonitoring();
        this.eventListeners.clear();
        this.log('Progress tracker destroyed', 'info');
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = BulkProgressTracker;
} else if (typeof window !== 'undefined') {
    window.BulkProgressTracker = BulkProgressTracker;
}