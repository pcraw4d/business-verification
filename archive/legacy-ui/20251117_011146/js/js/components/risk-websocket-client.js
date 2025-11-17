/**
 * Risk Assessment WebSocket Client
 * 
 * Provides real-time risk assessment updates via WebSocket connection with:
 * - Automatic reconnection with exponential backoff
 * - Message queuing for offline resilience
 * - Event-driven architecture for risk updates
 * - Connection status monitoring
 * - Error handling and retry logic
 */

class RiskWebSocketClient {
    constructor(options = {}) {
        this.options = {
            reconnectInterval: 1000,
            maxReconnectInterval: 30000,
            reconnectDecay: 1.5,
            maxReconnectAttempts: 10,
            heartbeatInterval: 30000,
            messageQueueSize: 100,
            ...options
        };

        this.state = {
            connected: false,
            connecting: false,
            reconnectAttempts: 0,
            lastHeartbeat: null,
            connectionId: null
        };

        this.ws = null;
        this.reconnectTimer = null;
        this.heartbeatTimer = null;
        this.messageQueue = [];
        this.eventListeners = new Map();
        this.subscriptions = new Set();

        this.init();
    }

    /**
     * Initialize the WebSocket client
     */
    init() {
        this.connect();
        this.bindGlobalEvents();
    }

    /**
     * Connect to the WebSocket server
     */
    connect() {
        if (this.state.connecting || this.state.connected) {
            return;
        }

        this.state.connecting = true;
        this.emit('connecting');

        try {
            const endpoints = APIConfig.getEndpoints();
            const wsUrl = endpoints.riskWebSocket;
            
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = (event) => this.handleOpen(event);
            this.ws.onmessage = (event) => this.handleMessage(event);
            this.ws.onclose = (event) => this.handleClose(event);
            this.ws.onerror = (event) => this.handleError(event);

        } catch (error) {
            console.error('WebSocket connection error:', error);
            this.handleConnectionError(error);
        }
    }

    /**
     * Handle WebSocket connection open
     */
    handleOpen(event) {
        console.log('Risk WebSocket connected');
        
        this.state.connected = true;
        this.state.connecting = false;
        this.state.reconnectAttempts = 0;
        
        this.emit('connected', { event });
        this.startHeartbeat();
        this.processMessageQueue();
        this.resubscribe();
    }

    /**
     * Handle incoming WebSocket messages
     */
    handleMessage(event) {
        try {
            const message = JSON.parse(event.data);
            this.state.lastHeartbeat = Date.now();

            switch (message.type) {
                case 'heartbeat':
                    this.handleHeartbeat(message);
                    break;
                case 'risk_update':
                    this.handleRiskUpdate(message);
                    break;
                case 'risk_prediction':
                    this.handleRiskPrediction(message);
                    break;
                case 'risk_alert':
                    this.handleRiskAlert(message);
                    break;
                case 'connection_id':
                    this.state.connectionId = message.connectionId;
                    break;
                default:
                    console.warn('Unknown message type:', message.type);
            }

            this.emit('message', message);

        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
            this.emit('error', { error, message: event.data });
        }
    }

    /**
     * Handle WebSocket connection close
     */
    handleClose(event) {
        console.log('Risk WebSocket disconnected:', event.code, event.reason);
        
        this.state.connected = false;
        this.state.connecting = false;
        
        this.stopHeartbeat();
        this.emit('disconnected', { event });

        // Attempt to reconnect if not a normal closure
        if (event.code !== 1000 && this.state.reconnectAttempts < this.options.maxReconnectAttempts) {
            this.scheduleReconnect();
        }
    }

    /**
     * Handle WebSocket errors
     */
    handleError(event) {
        console.error('Risk WebSocket error:', event);
        this.emit('error', { error: event });
    }

    /**
     * Handle connection errors
     */
    handleConnectionError(error) {
        this.state.connecting = false;
        this.emit('error', { error });
        
        if (this.state.reconnectAttempts < this.options.maxReconnectAttempts) {
            this.scheduleReconnect();
        }
    }

    /**
     * Schedule reconnection with exponential backoff
     */
    scheduleReconnect() {
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
        }

        const delay = Math.min(
            this.options.reconnectInterval * Math.pow(this.options.reconnectDecay, this.state.reconnectAttempts),
            this.options.maxReconnectInterval
        );

        this.state.reconnectAttempts++;
        this.emit('reconnecting', { 
            attempt: this.state.reconnectAttempts, 
            delay 
        });

        this.reconnectTimer = setTimeout(() => {
            this.connect();
        }, delay);
    }

    /**
     * Start heartbeat monitoring
     */
    startHeartbeat() {
        this.stopHeartbeat();
        
        this.heartbeatTimer = setInterval(() => {
            if (this.state.connected) {
                this.send({
                    type: 'heartbeat',
                    timestamp: Date.now()
                });
            }
        }, this.options.heartbeatInterval);
    }

    /**
     * Stop heartbeat monitoring
     */
    stopHeartbeat() {
        if (this.heartbeatTimer) {
            clearInterval(this.heartbeatTimer);
            this.heartbeatTimer = null;
        }
    }

    /**
     * Handle heartbeat response
     */
    handleHeartbeat(message) {
        // Heartbeat received, connection is alive
        this.emit('heartbeat', message);
    }

    /**
     * Handle risk update messages
     */
    handleRiskUpdate(message) {
        const { merchantId, riskData, timestamp } = message;
        
        // Emit risk update event
        this.emit('riskUpdate', {
            merchantId,
            riskData,
            timestamp: new Date(timestamp)
        });

        // Update UI if merchant is currently viewed
        this.updateMerchantRiskUI(merchantId, riskData);
    }

    /**
     * Handle risk prediction messages
     */
    handleRiskPrediction(message) {
        const { merchantId, predictions, confidence, timestamp } = message;
        
        this.emit('riskPrediction', {
            merchantId,
            predictions,
            confidence,
            timestamp: new Date(timestamp)
        });

        this.updateRiskPredictionUI(merchantId, predictions, confidence);
    }

    /**
     * Handle risk alert messages
     */
    handleRiskAlert(message) {
        const { merchantId, alertType, severity, message: alertMessage, timestamp } = message;
        
        this.emit('riskAlert', {
            merchantId,
            alertType,
            severity,
            message: alertMessage,
            timestamp: new Date(timestamp)
        });

        this.showRiskAlert(merchantId, alertType, severity, alertMessage);
    }

    /**
     * Send message to WebSocket server
     */
    send(message) {
        if (this.state.connected && this.ws && this.ws.readyState === WebSocket.OPEN) {
            try {
                this.ws.send(JSON.stringify(message));
                return true;
            } catch (error) {
                console.error('Error sending WebSocket message:', error);
                this.queueMessage(message);
                return false;
            }
        } else {
            this.queueMessage(message);
            return false;
        }
    }

    /**
     * Queue message for later sending
     */
    queueMessage(message) {
        this.messageQueue.push({
            message,
            timestamp: Date.now()
        });

        // Limit queue size
        if (this.messageQueue.length > this.options.messageQueueSize) {
            this.messageQueue.shift();
        }

        this.emit('messageQueued', { message, queueSize: this.messageQueue.length });
    }

    /**
     * Process queued messages
     */
    processMessageQueue() {
        while (this.messageQueue.length > 0 && this.state.connected) {
            const { message } = this.messageQueue.shift();
            this.send(message);
        }
    }

    /**
     * Subscribe to risk updates for a merchant
     */
    subscribe(merchantId) {
        this.subscriptions.add(merchantId);
        
        this.send({
            type: 'subscribe',
            merchantId,
            connectionId: this.state.connectionId
        });

        this.emit('subscribed', { merchantId });
    }

    /**
     * Unsubscribe from risk updates for a merchant
     */
    unsubscribe(merchantId) {
        this.subscriptions.delete(merchantId);
        
        this.send({
            type: 'unsubscribe',
            merchantId,
            connectionId: this.state.connectionId
        });

        this.emit('unsubscribed', { merchantId });
    }

    /**
     * Resubscribe to all previous subscriptions
     */
    resubscribe() {
        this.subscriptions.forEach(merchantId => {
            this.subscribe(merchantId);
        });
    }

    /**
     * Update merchant risk UI
     */
    updateMerchantRiskUI(merchantId, riskData) {
        // Update risk score display
        const riskScoreElement = document.querySelector(`[data-merchant-id="${merchantId}"] .risk-score-value`);
        if (riskScoreElement) {
            riskScoreElement.textContent = riskData.overallScore.toFixed(1);
            this.animateValueChange(riskScoreElement);
        }

        // Update risk categories
        const categoryElements = document.querySelectorAll(`[data-merchant-id="${merchantId}"] .risk-category`);
        categoryElements.forEach(element => {
            const categoryName = element.querySelector('.category-name').textContent.toLowerCase().replace(' risk', '');
            const categoryScore = element.querySelector('.category-score');
            
            if (riskData.categories[categoryName]) {
                categoryScore.textContent = riskData.categories[categoryName].toFixed(1);
                this.updateRiskCategoryClass(categoryScore, riskData.categories[categoryName]);
                this.animateValueChange(categoryScore);
            }
        });

        // Update risk trend
        const trendElement = document.querySelector(`[data-merchant-id="${merchantId}"] .risk-score-trend`);
        if (trendElement && riskData.trend) {
            const trendIcon = trendElement.querySelector('i');
            const trendText = trendElement.querySelector('span');
            
            if (riskData.trend > 0) {
                trendIcon.className = 'fas fa-arrow-up text-red-500';
                trendText.textContent = `+${riskData.trend.toFixed(1)} from last update`;
            } else if (riskData.trend < 0) {
                trendIcon.className = 'fas fa-arrow-down text-green-500';
                trendText.textContent = `${riskData.trend.toFixed(1)} from last update`;
            } else {
                trendIcon.className = 'fas fa-minus text-gray-500';
                trendText.textContent = 'No change from last update';
            }
        }
    }

    /**
     * Update risk prediction UI
     */
    updateRiskPredictionUI(merchantId, predictions, confidence) {
        const predictionContainer = document.querySelector(`[data-merchant-id="${merchantId}"] .risk-predictions`);
        if (predictionContainer) {
            predictionContainer.innerHTML = `
                <div class="prediction-item">
                    <span class="prediction-label">3-Month Forecast:</span>
                    <span class="prediction-value">${predictions.threeMonth.toFixed(1)}</span>
                    <span class="prediction-confidence">(${(confidence * 100).toFixed(0)}% confidence)</span>
                </div>
                <div class="prediction-item">
                    <span class="prediction-label">6-Month Forecast:</span>
                    <span class="prediction-value">${predictions.sixMonth.toFixed(1)}</span>
                    <span class="prediction-confidence">(${(confidence * 100).toFixed(0)}% confidence)</span>
                </div>
            `;
        }
    }

    /**
     * Show risk alert notification
     */
    showRiskAlert(merchantId, alertType, severity, message) {
        const alertContainer = document.getElementById('riskAlerts') || this.createAlertContainer();
        
        const alertElement = document.createElement('div');
        alertElement.className = `risk-alert alert-${severity}`;
        alertElement.innerHTML = `
            <div class="alert-header">
                <i class="fas fa-exclamation-triangle"></i>
                <span class="alert-type">${alertType}</span>
                <button class="alert-close" onclick="this.parentElement.parentElement.remove()">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="alert-message">${message}</div>
            <div class="alert-merchant">Merchant ID: ${merchantId}</div>
        `;

        alertContainer.appendChild(alertElement);

        // Auto-remove after 10 seconds
        setTimeout(() => {
            if (alertElement.parentElement) {
                alertElement.remove();
            }
        }, 10000);
    }

    /**
     * Create alert container if it doesn't exist
     */
    createAlertContainer() {
        const container = document.createElement('div');
        container.id = 'riskAlerts';
        container.className = 'risk-alerts-container';
        container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 1000;
            max-width: 400px;
        `;
        document.body.appendChild(container);
        return container;
    }

    /**
     * Animate value changes
     */
    animateValueChange(element) {
        element.style.transform = 'scale(1.1)';
        element.style.transition = 'transform 0.3s ease';
        
        setTimeout(() => {
            element.style.transform = 'scale(1)';
        }, 300);
    }

    /**
     * Update risk category class based on score
     */
    updateRiskCategoryClass(element, score) {
        element.className = 'category-score';
        
        if (score <= 3) {
            element.classList.add('low');
        } else if (score <= 6) {
            element.classList.add('medium');
        } else if (score <= 8) {
            element.classList.add('high');
        } else {
            element.classList.add('critical');
        }
    }

    /**
     * Add event listener
     */
    on(event, callback) {
        if (!this.eventListeners.has(event)) {
            this.eventListeners.set(event, []);
        }
        this.eventListeners.get(event).push(callback);
    }

    /**
     * Remove event listener
     */
    off(event, callback) {
        if (this.eventListeners.has(event)) {
            const listeners = this.eventListeners.get(event);
            const index = listeners.indexOf(callback);
            if (index > -1) {
                listeners.splice(index, 1);
            }
        }
    }

    /**
     * Emit event to listeners
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
     * Bind global events
     */
    bindGlobalEvents() {
        // Subscribe to current merchant when page loads
        window.addEventListener('load', () => {
            const merchantId = this.getCurrentMerchantId();
            if (merchantId) {
                this.subscribe(merchantId);
            }
        });

        // Handle page visibility changes
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                this.stopHeartbeat();
            } else if (this.state.connected) {
                this.startHeartbeat();
            }
        });

        // Handle beforeunload
        window.addEventListener('beforeunload', () => {
            this.disconnect();
        });
    }

    /**
     * Get current merchant ID from URL or page context
     */
    getCurrentMerchantId() {
        const urlParams = new URLSearchParams(window.location.search);
        const merchantId = urlParams.get('id') || urlParams.get('merchantId');
        
        if (merchantId) {
            return merchantId;
        }

        // Try to get from page context
        const merchantElement = document.querySelector('[data-merchant-id]');
        if (merchantElement) {
            return merchantElement.getAttribute('data-merchant-id');
        }

        return null;
    }

    /**
     * Get connection status
     */
    getStatus() {
        return {
            connected: this.state.connected,
            connecting: this.state.connecting,
            reconnectAttempts: this.state.reconnectAttempts,
            lastHeartbeat: this.state.lastHeartbeat,
            connectionId: this.state.connectionId,
            subscriptions: Array.from(this.subscriptions),
            queueSize: this.messageQueue.length
        };
    }

    /**
     * Disconnect WebSocket
     */
    disconnect() {
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }

        this.stopHeartbeat();

        if (this.ws) {
            this.ws.close(1000, 'Client disconnect');
            this.ws = null;
        }

        this.state.connected = false;
        this.state.connecting = false;
        this.emit('disconnected', { reason: 'Client disconnect' });
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskWebSocketClient;
}
