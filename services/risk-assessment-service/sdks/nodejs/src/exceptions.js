/**
 * KYB Platform Risk Assessment Service Node.js SDK Exceptions
 * 
 * This module defines custom exceptions for the KYB Platform SDK.
 */

/**
 * Base exception for all KYB SDK errors.
 */
class KYBError extends Error {
    constructor(message, errorData = {}) {
        super(message);
        this.name = this.constructor.name;
        this.message = message;
        this.errorData = errorData;
        
        // Maintains proper stack trace for where our error was thrown (only available on V8)
        if (Error.captureStackTrace) {
            Error.captureStackTrace(this, this.constructor);
        }
    }

    /**
     * Get the request ID from the error data.
     * @returns {string|null} The request ID
     */
    getRequestId() {
        return this.errorData.request_id || null;
    }

    /**
     * Get the timestamp from the error data.
     * @returns {string|null} The timestamp
     */
    getTimestamp() {
        return this.errorData.timestamp || null;
    }

    /**
     * Get the request path from the error data.
     * @returns {string|null} The request path
     */
    getPath() {
        return this.errorData.path || null;
    }

    /**
     * Get the request method from the error data.
     * @returns {string|null} The request method
     */
    getMethod() {
        return this.errorData.method || null;
    }
}

/**
 * Exception raised for API errors.
 */
class APIError extends KYBError {
    constructor(message, errorData = {}, statusCode = null) {
        super(message, errorData);
        this.statusCode = statusCode;
    }

    /**
     * Get validation errors if any.
     * @returns {Array} Array of validation errors
     */
    getValidationErrors() {
        return this.errorData.error?.validation || [];
    }
}

/**
 * Exception raised for validation errors.
 */
class ValidationError extends APIError {
    constructor(message, errorData = {}) {
        super(message, errorData, 400);
    }

    /**
     * Get detailed validation errors.
     * @returns {Array} Array of validation errors
     */
    getValidationErrors() {
        return this.errorData.error?.validation || [];
    }
}

/**
 * Exception raised for authentication errors.
 */
class AuthenticationError extends APIError {
    constructor(message, errorData = {}) {
        super(message, errorData, 401);
    }
}

/**
 * Exception raised for authorization errors.
 */
class AuthorizationError extends APIError {
    constructor(message, errorData = {}) {
        super(message, errorData, 403);
    }
}

/**
 * Exception raised when a resource is not found.
 */
class NotFoundError extends APIError {
    constructor(message, errorData = {}) {
        super(message, errorData, 404);
    }
}

/**
 * Exception raised when rate limit is exceeded.
 */
class RateLimitError extends APIError {
    constructor(message, errorData = {}) {
        super(message, errorData, 429);
    }
}

/**
 * Exception raised when the service is unavailable.
 */
class ServiceUnavailableError extends APIError {
    constructor(message, errorData = {}) {
        super(message, errorData, 503);
    }
}

/**
 * Exception raised when a request times out.
 */
class TimeoutError extends APIError {
    constructor(message, errorData = {}) {
        super(message, errorData, 408);
    }
}

/**
 * Exception raised for internal server errors.
 */
class InternalError extends APIError {
    constructor(message, errorData = {}) {
        super(message, errorData, 500);
    }
}

module.exports = {
    KYBError,
    APIError,
    ValidationError,
    AuthenticationError,
    AuthorizationError,
    NotFoundError,
    RateLimitError,
    ServiceUnavailableError,
    TimeoutError,
    InternalError
};
