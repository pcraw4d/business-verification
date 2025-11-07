/**
 * Circuit breaker implementation for JavaScript
 * Prevents cascading failures by stopping requests when a service is failing
 */

/**
 * Circuit breaker states
 */
const CircuitState = {
    CLOSED: 'closed',      // Normal operation, requests allowed
    OPEN: 'open',          // Circuit is open, requests immediately rejected
    HALF_OPEN: 'half-open' // Testing if service has recovered
};

/**
 * Circuit breaker configuration
 */
export class CircuitBreakerConfig {
    constructor(options = {}) {
        this.failureThreshold = options.failureThreshold || 5;      // Failures before opening
        this.successThreshold = options.successThreshold || 2;       // Successes to close from half-open
        this.timeout = options.timeout || 30000;                     // Time to wait before half-open (ms)
        this.maxRequests = options.maxRequests || 1;                 // Max requests in half-open
        this.resetTimeout = options.resetTimeout || 60000;            // Time to reset failure count (ms)
    }
}

/**
 * Circuit breaker implementation
 */
export class CircuitBreaker {
    constructor(name, config = new CircuitBreakerConfig()) {
        this.name = name;
        this.config = config;
        this.state = CircuitState.CLOSED;
        this.failureCount = 0;
        this.successCount = 0;
        this.halfOpenCount = 0;
        this.lastFailure = null;
        this.stateChange = Date.now();
    }
    
    /**
     * Execute a function through the circuit breaker
     * 
     * @param {Function} fn - Function to execute (should return a Promise)
     * @returns {Promise} Promise that resolves with function result or rejects with circuit breaker error
     */
    async execute(fn) {
        // Check if request should be allowed
        if (!this.allowRequest()) {
            throw new Error(`Circuit breaker ${this.name} is ${this.state}`);
        }
        
        try {
            const result = await fn();
            this.onSuccess();
            return result;
        } catch (error) {
            this.onFailure();
            throw error;
        }
    }
    
    /**
     * Check if a request should be allowed based on current state
     * @returns {boolean} True if request should be allowed
     */
    allowRequest() {
        switch (this.state) {
            case CircuitState.CLOSED:
                return true;
                
            case CircuitState.OPEN:
                // Check if timeout has elapsed to transition to half-open
                if (Date.now() - this.stateChange >= this.config.timeout) {
                    this.state = CircuitState.HALF_OPEN;
                    this.stateChange = Date.now();
                    this.halfOpenCount = 0;
                    return true;
                }
                return false;
                
            case CircuitState.HALF_OPEN:
                // Allow limited requests in half-open state
                return this.halfOpenCount < this.config.maxRequests;
                
            default:
                return false;
        }
    }
    
    /**
     * Handle successful request
     */
    onSuccess() {
        this.failureCount = 0;
        
        switch (this.state) {
            case CircuitState.HALF_OPEN:
                this.successCount++;
                this.halfOpenCount++;
                // Close circuit if success threshold reached
                if (this.successCount >= this.config.successThreshold) {
                    this.state = CircuitState.CLOSED;
                    this.stateChange = Date.now();
                    this.successCount = 0;
                    this.halfOpenCount = 0;
                }
                break;
                
            case CircuitState.CLOSED:
                // Reset failure count after reset timeout
                if (this.lastFailure && Date.now() - this.lastFailure >= this.config.resetTimeout) {
                    this.failureCount = 0;
                }
                break;
        }
    }
    
    /**
     * Handle failed request
     */
    onFailure() {
        this.failureCount++;
        this.lastFailure = Date.now();
        
        switch (this.state) {
            case CircuitState.CLOSED:
                // Open circuit if failure threshold reached
                if (this.failureCount >= this.config.failureThreshold) {
                    this.state = CircuitState.OPEN;
                    this.stateChange = Date.now();
                }
                break;
                
            case CircuitState.HALF_OPEN:
                // Immediately open circuit on failure in half-open state
                this.state = CircuitState.OPEN;
                this.stateChange = Date.now();
                this.halfOpenCount = 0;
                this.successCount = 0;
                break;
        }
    }
    
    /**
     * Get current state
     * @returns {string} Current circuit breaker state
     */
    getState() {
        return this.state;
    }
    
    /**
     * Get statistics about the circuit breaker
     * @returns {Object} Statistics object
     */
    getStats() {
        return {
            name: this.name,
            state: this.state,
            failureCount: this.failureCount,
            successCount: this.successCount,
            halfOpenCount: this.halfOpenCount,
            lastFailure: this.lastFailure,
            stateChange: this.stateChange
        };
    }
    
    /**
     * Reset the circuit breaker to closed state
     */
    reset() {
        this.state = CircuitState.CLOSED;
        this.failureCount = 0;
        this.successCount = 0;
        this.halfOpenCount = 0;
        this.lastFailure = null;
        this.stateChange = Date.now();
    }
}

/**
 * Create a circuit breaker instance
 * 
 * @param {string} name - Name of the circuit breaker
 * @param {CircuitBreakerConfig|Object} config - Configuration object
 * @returns {CircuitBreaker} Circuit breaker instance
 */
export function createCircuitBreaker(name, config) {
    const breakerConfig = config instanceof CircuitBreakerConfig ? config : new CircuitBreakerConfig(config);
    return new CircuitBreaker(name, breakerConfig);
}

