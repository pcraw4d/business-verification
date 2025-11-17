/**
 * Event Streaming Component
 * Handles real-time event streaming via WebSocket or SSE
 */
class EventStream {
    constructor(options = {}) {
        this.options = {
            url: options.url || '/api/v1/events/stream',
            useWebSocket: options.useWebSocket !== false,
            reconnectInterval: options.reconnectInterval || 5000,
            maxReconnectAttempts: options.maxReconnectAttempts || 10,
            ...options
        };
        this.ws = null;
        this.eventSource = null;
        this.reconnectAttempts = 0;
        this.listeners = new Map();
        this.isConnected = false;
    }

    /**
     * Initialize the event stream
     */
    async init() {
        if (this.options.useWebSocket) {
            await this.initWebSocket();
        } else {
            await this.initEventSource();
        }
    }

    /**
     * Initialize WebSocket connection
     */
    async initWebSocket() {
        try {
            const wsUrl = this.options.url.replace('http://', 'ws://').replace('https://', 'wss://');
            
            // Add connection timeout
            const connectionTimeout = setTimeout(() => {
                if (this.ws && this.ws.readyState === WebSocket.CONNECTING) {
                    this.ws.close();
                    this.emit('error', new Error('WebSocket connection timeout'));
                }
            }, 10000); // 10 second timeout
            
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                clearTimeout(connectionTimeout);
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.emit('connected');
                console.log('Event stream connected via WebSocket');
            };

            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleEvent(data);
                } catch (error) {
                    console.error('Error parsing event data:', error);
                }
            };

            this.ws.onerror = (error) => {
                clearTimeout(connectionTimeout);
                console.error('WebSocket error:', error);
                this.emit('error', error);
            };

            this.ws.onclose = () => {
                this.isConnected = false;
                this.emit('disconnected');
                this.attemptReconnect();
            };
        } catch (error) {
            console.error('Error initializing WebSocket:', error);
            this.emit('error', error);
        }
    }

    /**
     * Initialize EventSource (SSE) connection
     */
    async initEventSource() {
        try {
            this.eventSource = new EventSource(this.options.url);

            this.eventSource.onopen = () => {
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.emit('connected');
                console.log('Event stream connected via SSE');
            };

            this.eventSource.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleEvent(data);
                } catch (error) {
                    console.error('Error parsing event data:', error);
                }
            };

            this.eventSource.onerror = (error) => {
                console.error('EventSource error:', error);
                this.isConnected = false;
                this.emit('error', error);
                this.attemptReconnect();
            };
        } catch (error) {
            console.error('Error initializing EventSource:', error);
            this.emit('error', error);
        }
    }

    /**
     * Handle incoming event
     */
    handleEvent(data) {
        const eventType = data.type || 'message';
        this.emit(eventType, data);
        this.emit('*', data); // Emit to all listeners
    }

    /**
     * Subscribe to events
     */
    on(eventType, callback) {
        if (!this.listeners.has(eventType)) {
            this.listeners.set(eventType, []);
        }
        this.listeners.get(eventType).push(callback);
    }

    /**
     * Unsubscribe from events
     */
    off(eventType, callback) {
        if (!this.listeners.has(eventType)) return;
        
        const callbacks = this.listeners.get(eventType);
        const index = callbacks.indexOf(callback);
        if (index > -1) {
            callbacks.splice(index, 1);
        if (callbacks.length === 0) {
            this.listeners.delete(eventType);
        }
    }

    /**
     * Emit event to listeners
     */
    emit(eventType, data) {
        if (this.listeners.has(eventType)) {
            this.listeners.get(eventType).forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    console.error('Error in event listener:', error);
                }
            });
        }
    }

    /**
     * Attempt to reconnect
     */
    attemptReconnect() {
        if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
            console.error('Max reconnect attempts reached');
            this.emit('reconnect_failed');
            return;
        }

        this.reconnectAttempts++;
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.options.maxReconnectAttempts})...`);

        setTimeout(() => {
            if (this.options.useWebSocket) {
                this.initWebSocket();
            } else {
                this.initEventSource();
            }
        }, this.options.reconnectInterval);
    }

    /**
     * Close the connection
     */
    close() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }
        this.isConnected = false;
    }

    /**
     * Send message (WebSocket only)
     */
    send(data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data));
        } else {
            console.warn('WebSocket not connected');
        }
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.EventStream = EventStream;
}

