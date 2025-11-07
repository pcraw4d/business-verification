/**
 * Retry utilities for API calls with exponential backoff
 */

// Make available globally for non-module environments
if (typeof window !== 'undefined') {
    window.retryWithBackoff = retryWithBackoff;
    window.retrySimple = retrySimple;
}

/**
 * Retry a function with exponential backoff
 * 
 * @param {Function} fn - Function to retry (should return a Promise)
 * @param {Object} options - Retry options
 * @param {number} options.maxAttempts - Maximum number of retry attempts (default: 3)
 * @param {number} options.initialDelay - Initial delay in milliseconds (default: 100)
 * @param {number} options.maxDelay - Maximum delay in milliseconds (default: 5000)
 * @param {number} options.multiplier - Exponential backoff multiplier (default: 2.0)
 * @param {boolean} options.jitter - Enable jitter to prevent thundering herd (default: true)
 * @returns {Promise} Promise that resolves with the function result or rejects after all attempts fail
 */
export async function retryWithBackoff(fn, options = {}) {
    const {
        maxAttempts = 3,
        initialDelay = 100,
        maxDelay = 5000,
        multiplier = 2.0,
        jitter = true
    } = options;
    
    let lastError;
    
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
        try {
            const result = await fn();
            return result;
        } catch (error) {
            lastError = error;
            
            // Don't wait after the last attempt
            if (attempt < maxAttempts - 1) {
                // Calculate delay with exponential backoff
                let delay = initialDelay * Math.pow(multiplier, attempt);
                
                // Cap at max delay
                if (delay > maxDelay) {
                    delay = maxDelay;
                }
                
                // Add jitter if enabled (up to 25% of delay)
                if (jitter) {
                    const jitterAmount = Math.random() * 0.25 * delay;
                    delay += jitterAmount;
                }
                
                // Wait before retry
                await new Promise(resolve => setTimeout(resolve, delay));
            }
        }
    }
    
    // All attempts failed
    throw new Error(`Operation failed after ${maxAttempts} attempts: ${lastError.message}`);
}

/**
 * Retry a function with simple configuration
 * 
 * @param {Function} fn - Function to retry (should return a Promise)
 * @param {number} maxAttempts - Maximum number of retry attempts (default: 3)
 * @param {number} initialDelay - Initial delay in milliseconds (default: 100)
 * @returns {Promise} Promise that resolves with the function result or rejects after all attempts fail
 */
export async function retrySimple(fn, maxAttempts = 3, initialDelay = 100) {
    return retryWithBackoff(fn, { maxAttempts, initialDelay });
}

