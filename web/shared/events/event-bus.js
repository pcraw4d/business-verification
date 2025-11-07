/**
 * Event Bus
 * Centralized event system for component communication
 */
class EventBus {
    constructor() {
        this.listeners = new Map();
    }
    
    /**
     * Subscribe to event
     * @param {string} eventType - Event type
     * @param {Function} callback - Callback function
     * @returns {Function} Unsubscribe function
     */
    on(eventType, callback) {
        if (!this.listeners.has(eventType)) {
            this.listeners.set(eventType, []);
        }
        
        this.listeners.get(eventType).push(callback);
        
        // Return unsubscribe function
        return () => this.off(eventType, callback);
    }
    
    /**
     * Unsubscribe from event
     * @param {string} eventType - Event type
     * @param {Function} callback - Callback function
     */
    off(eventType, callback) {
        if (!this.listeners.has(eventType)) return;
        
        const callbacks = this.listeners.get(eventType);
        const index = callbacks.indexOf(callback);
        if (index > -1) {
            callbacks.splice(index, 1);
        }
    }
    
    /**
     * Emit event
     * @param {string} eventType - Event type
     * @param {*} data - Event data
     */
    emit(eventType, data) {
        if (!this.listeners.has(eventType)) return;
        
        const callbacks = this.listeners.get(eventType);
        callbacks.forEach(callback => {
            try {
                callback(data);
            } catch (error) {
                console.error(`Error in event listener for ${eventType}:`, error);
            }
        });
    }
    
    /**
     * Subscribe to event once
     * @param {string} eventType - Event type
     * @param {Function} callback - Callback function
     */
    once(eventType, callback) {
        const wrappedCallback = (data) => {
            callback(data);
            this.off(eventType, wrappedCallback);
        };
        this.on(eventType, wrappedCallback);
    }
}

// Export singleton instance
let eventBusInstance = null;

export function getEventBus() {
    if (!eventBusInstance) {
        eventBusInstance = new EventBus();
    }
    return eventBusInstance;
}

export { EventBus };

// Make available globally for non-module environments
if (typeof window !== 'undefined') {
    window.getEventBus = getEventBus;
    window.EventBus = EventBus;
}

