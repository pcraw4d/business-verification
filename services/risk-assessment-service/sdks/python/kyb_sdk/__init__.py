"""
KYB Platform Risk Assessment Service Python SDK

This SDK provides a Python interface to the KYB Platform Risk Assessment Service API.
"""

from .client import KYBClient
from .exceptions import KYBError, ValidationError, AuthenticationError, AuthorizationError
from .exceptions import NotFoundError, RateLimitError, ServiceUnavailableError, TimeoutError
from .exceptions import InternalError, APIError

__version__ = "1.0.0"
__author__ = "KYB Platform"
__email__ = "support@kyb-platform.com"

__all__ = [
    "KYBClient",
    "KYBError",
    "ValidationError",
    "AuthenticationError",
    "AuthorizationError",
    "NotFoundError",
    "RateLimitError",
    "ServiceUnavailableError",
    "TimeoutError",
    "InternalError",
    "APIError",
]
