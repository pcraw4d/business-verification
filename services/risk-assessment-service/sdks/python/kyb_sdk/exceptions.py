"""
KYB Platform Risk Assessment Service Python SDK Exceptions

This module defines custom exceptions for the KYB Platform SDK.
"""

from typing import Dict, Any, Optional


class KYBError(Exception):
    """Base exception for all KYB SDK errors."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message)
        self.message = message
        self.error_data = error_data or {}
    
    def get_request_id(self) -> Optional[str]:
        """Get the request ID from the error data."""
        return self.error_data.get("request_id")
    
    def get_timestamp(self) -> Optional[str]:
        """Get the timestamp from the error data."""
        return self.error_data.get("timestamp")
    
    def get_path(self) -> Optional[str]:
        """Get the request path from the error data."""
        return self.error_data.get("path")
    
    def get_method(self) -> Optional[str]:
        """Get the request method from the error data."""
        return self.error_data.get("method")


class APIError(KYBError):
    """Exception raised for API errors."""
    
    def __init__(
        self,
        message: str,
        error_data: Optional[Dict[str, Any]] = None,
        status_code: Optional[int] = None
    ):
        super().__init__(message, error_data)
        self.status_code = status_code
    
    def get_validation_errors(self) -> list:
        """Get validation errors if any."""
        return self.error_data.get("error", {}).get("validation", [])


class ValidationError(APIError):
    """Exception raised for validation errors."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message, error_data, 400)
    
    def get_validation_errors(self) -> list:
        """Get detailed validation errors."""
        return self.error_data.get("error", {}).get("validation", [])


class AuthenticationError(APIError):
    """Exception raised for authentication errors."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message, error_data, 401)


class AuthorizationError(APIError):
    """Exception raised for authorization errors."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message, error_data, 403)


class NotFoundError(APIError):
    """Exception raised when a resource is not found."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message, error_data, 404)


class RateLimitError(APIError):
    """Exception raised when rate limit is exceeded."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message, error_data, 429)


class ServiceUnavailableError(APIError):
    """Exception raised when the service is unavailable."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message, error_data, 503)


class TimeoutError(APIError):
    """Exception raised when a request times out."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message, error_data, 408)


class InternalError(APIError):
    """Exception raised for internal server errors."""
    
    def __init__(self, message: str, error_data: Optional[Dict[str, Any]] = None):
        super().__init__(message, error_data, 500)
