"""
KYB Platform Risk Assessment Service Python SDK Client

This module provides the main client class for interacting with the KYB Platform API.
"""

import json
import time
from typing import Dict, List, Optional, Any, Union
from urllib.parse import urlencode
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

from .exceptions import KYBError, APIError, ValidationError, AuthenticationError
from .exceptions import AuthorizationError, NotFoundError, RateLimitError
from .exceptions import ServiceUnavailableError, TimeoutError, InternalError


class KYBClient:
    """
    KYB Platform Risk Assessment Service Client
    
    This client provides methods to interact with the KYB Platform API for
    risk assessment, compliance checking, and analytics.
    """
    
    def __init__(
        self,
        api_key: str,
        base_url: str = "https://api.kyb-platform.com/v1",
        timeout: int = 30,
        max_retries: int = 3,
        user_agent: str = "kyb-python-client/1.0.0"
    ):
        """
        Initialize the KYB client.
        
        Args:
            api_key: Your KYB Platform API key
            base_url: Base URL for the API (default: https://api.kyb-platform.com/v1)
            timeout: Request timeout in seconds (default: 30)
            max_retries: Maximum number of retries for failed requests (default: 3)
            user_agent: User agent string for requests (default: kyb-python-client/1.0.0)
        """
        if not api_key:
            raise ValueError("API key is required")
        
        self.api_key = api_key
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.user_agent = user_agent
        
        # Setup session with retry strategy
        self.session = requests.Session()
        
        # Configure retry strategy
        retry_strategy = Retry(
            total=max_retries,
            backoff_factor=1,
            status_forcelist=[429, 500, 502, 503, 504],
            allowed_methods=["HEAD", "GET", "POST", "PUT", "DELETE", "OPTIONS", "TRACE"]
        )
        
        adapter = HTTPAdapter(max_retries=retry_strategy)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)
        
        # Set default headers
        self.session.headers.update({
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
            "User-Agent": user_agent
        })
    
    def assess_risk(
        self,
        business_name: str,
        business_address: str,
        industry: str,
        country: str,
        phone: Optional[str] = None,
        email: Optional[str] = None,
        website: Optional[str] = None,
        prediction_horizon: int = 3,
        metadata: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Perform a risk assessment for a business.
        
        Args:
            business_name: Name of the business
            business_address: Address of the business
            industry: Industry of the business
            country: Country code (2-letter ISO)
            phone: Phone number (optional)
            email: Email address (optional)
            website: Website URL (optional)
            prediction_horizon: Prediction horizon in months (0-12, default: 3)
            metadata: Additional metadata (optional)
            
        Returns:
            Dict containing the risk assessment results
            
        Raises:
            ValidationError: If the request data is invalid
            APIError: If the API request fails
        """
        # Validate required fields
        if not business_name:
            raise ValidationError("business_name is required")
        if not business_address:
            raise ValidationError("business_address is required")
        if not industry:
            raise ValidationError("industry is required")
        if not country:
            raise ValidationError("country is required")
        if len(country) != 2:
            raise ValidationError("country must be a 2-letter ISO code")
        if prediction_horizon < 0 or prediction_horizon > 12:
            raise ValidationError("prediction_horizon must be between 0 and 12 months")
        
        # Prepare request data
        data = {
            "business_name": business_name,
            "business_address": business_address,
            "industry": industry,
            "country": country,
            "prediction_horizon": prediction_horizon
        }
        
        if phone:
            data["phone"] = phone
        if email:
            data["email"] = email
        if website:
            data["website"] = website
        if metadata:
            data["metadata"] = metadata
        
        return self._make_request("POST", "/assess", data=data)
    
    def get_risk_assessment(self, assessment_id: str) -> Dict[str, Any]:
        """
        Retrieve a risk assessment by ID.
        
        Args:
            assessment_id: The assessment ID
            
        Returns:
            Dict containing the risk assessment data
            
        Raises:
            ValidationError: If the assessment ID is invalid
            NotFoundError: If the assessment is not found
            APIError: If the API request fails
        """
        if not assessment_id:
            raise ValidationError("assessment_id is required")
        
        return self._make_request("GET", f"/assess/{assessment_id}")
    
    def predict_risk(
        self,
        assessment_id: str,
        horizon_months: int,
        scenarios: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        Perform future risk prediction for a business.
        
        Args:
            assessment_id: The assessment ID
            horizon_months: Prediction horizon in months (1-12)
            scenarios: List of scenarios to analyze (optional)
            
        Returns:
            Dict containing the risk prediction results
            
        Raises:
            ValidationError: If the request data is invalid
            NotFoundError: If the assessment is not found
            APIError: If the API request fails
        """
        if not assessment_id:
            raise ValidationError("assessment_id is required")
        if horizon_months <= 0 or horizon_months > 12:
            raise ValidationError("horizon_months must be between 1 and 12")
        
        data = {"horizon_months": horizon_months}
        if scenarios:
            data["scenarios"] = scenarios
        
        return self._make_request("POST", f"/assess/{assessment_id}/predict", data=data)
    
    def get_risk_history(self, assessment_id: str) -> Dict[str, Any]:
        """
        Retrieve risk assessment history for a business.
        
        Args:
            assessment_id: The assessment ID
            
        Returns:
            Dict containing the risk history data
            
        Raises:
            ValidationError: If the assessment ID is invalid
            NotFoundError: If the assessment is not found
            APIError: If the API request fails
        """
        if not assessment_id:
            raise ValidationError("assessment_id is required")
        
        return self._make_request("GET", f"/assess/{assessment_id}/history")
    
    def check_compliance(
        self,
        business_name: str,
        business_address: str,
        industry: str,
        country: str,
        compliance_types: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        Perform compliance checks for a business.
        
        Args:
            business_name: Name of the business
            business_address: Address of the business
            industry: Industry of the business
            country: Country code (2-letter ISO)
            compliance_types: List of compliance types to check (optional)
            
        Returns:
            Dict containing the compliance check results
            
        Raises:
            ValidationError: If the request data is invalid
            APIError: If the API request fails
        """
        if not business_name:
            raise ValidationError("business_name is required")
        if not business_address:
            raise ValidationError("business_address is required")
        if not industry:
            raise ValidationError("industry is required")
        if not country:
            raise ValidationError("country is required")
        
        data = {
            "business_name": business_name,
            "business_address": business_address,
            "industry": industry,
            "country": country
        }
        
        if compliance_types:
            data["compliance_types"] = compliance_types
        
        return self._make_request("POST", "/compliance/check", data=data)
    
    def screen_sanctions(
        self,
        business_name: str,
        business_address: str,
        country: str
    ) -> Dict[str, Any]:
        """
        Perform sanctions screening for a business.
        
        Args:
            business_name: Name of the business
            business_address: Address of the business
            country: Country code (2-letter ISO)
            
        Returns:
            Dict containing the sanctions screening results
            
        Raises:
            ValidationError: If the request data is invalid
            APIError: If the API request fails
        """
        if not business_name:
            raise ValidationError("business_name is required")
        if not business_address:
            raise ValidationError("business_address is required")
        if not country:
            raise ValidationError("country is required")
        
        data = {
            "business_name": business_name,
            "business_address": business_address,
            "country": country
        }
        
        return self._make_request("POST", "/sanctions/screen", data=data)
    
    def monitor_media(
        self,
        business_name: str,
        business_address: str,
        monitoring_types: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        Set up adverse media monitoring for a business.
        
        Args:
            business_name: Name of the business
            business_address: Address of the business
            monitoring_types: List of monitoring types (optional)
            
        Returns:
            Dict containing the media monitoring setup results
            
        Raises:
            ValidationError: If the request data is invalid
            APIError: If the API request fails
        """
        if not business_name:
            raise ValidationError("business_name is required")
        if not business_address:
            raise ValidationError("business_address is required")
        
        data = {
            "business_name": business_name,
            "business_address": business_address
        }
        
        if monitoring_types:
            data["monitoring_types"] = monitoring_types
        
        return self._make_request("POST", "/media/monitor", data=data)
    
    def get_risk_trends(
        self,
        industry: Optional[str] = None,
        country: Optional[str] = None,
        timeframe: Optional[str] = None,
        limit: Optional[int] = None
    ) -> Dict[str, Any]:
        """
        Retrieve risk trends and analytics.
        
        Args:
            industry: Filter by industry (optional)
            country: Filter by country (optional)
            timeframe: Time period (7d, 30d, 90d, 1y) (optional)
            limit: Number of results (optional)
            
        Returns:
            Dict containing the risk trends data
            
        Raises:
            APIError: If the API request fails
        """
        params = {}
        if industry:
            params["industry"] = industry
        if country:
            params["country"] = country
        if timeframe:
            params["timeframe"] = timeframe
        if limit:
            params["limit"] = limit
        
        return self._make_request("GET", "/analytics/trends", params=params)
    
    def get_risk_insights(
        self,
        industry: Optional[str] = None,
        country: Optional[str] = None,
        risk_level: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Retrieve risk insights and recommendations.
        
        Args:
            industry: Filter by industry (optional)
            country: Filter by country (optional)
            risk_level: Filter by risk level (optional)
            
        Returns:
            Dict containing the risk insights data
            
        Raises:
            APIError: If the API request fails
        """
        params = {}
        if industry:
            params["industry"] = industry
        if country:
            params["country"] = country
        if risk_level:
            params["risk_level"] = risk_level
        
        return self._make_request("GET", "/analytics/insights", params=params)
    
    def _make_request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Make an HTTP request to the API.
        
        Args:
            method: HTTP method
            endpoint: API endpoint
            data: Request data (for POST/PUT requests)
            params: Query parameters
            
        Returns:
            Dict containing the response data
            
        Raises:
            APIError: If the API request fails
        """
        url = self.base_url + endpoint
        
        try:
            if method.upper() == "GET":
                response = self.session.get(url, params=params, timeout=self.timeout)
            elif method.upper() == "POST":
                response = self.session.post(url, json=data, params=params, timeout=self.timeout)
            elif method.upper() == "PUT":
                response = self.session.put(url, json=data, params=params, timeout=self.timeout)
            elif method.upper() == "DELETE":
                response = self.session.delete(url, params=params, timeout=self.timeout)
            else:
                raise APIError(f"Unsupported HTTP method: {method}")
            
            # Check for HTTP errors
            if response.status_code >= 400:
                self._handle_error_response(response)
            
            # Parse response
            try:
                return response.json()
            except json.JSONDecodeError:
                raise APIError(f"Invalid JSON response: {response.text}")
                
        except requests.exceptions.Timeout:
            raise TimeoutError("Request timeout")
        except requests.exceptions.ConnectionError:
            raise ServiceUnavailableError("Service unavailable")
        except requests.exceptions.RequestException as e:
            raise APIError(f"Request failed: {str(e)}")
    
    def _handle_error_response(self, response: requests.Response) -> None:
        """
        Handle error responses from the API.
        
        Args:
            response: The HTTP response object
            
        Raises:
            Appropriate exception based on error type
        """
        try:
            error_data = response.json()
        except json.JSONDecodeError:
            # Fallback to generic error
            raise APIError(f"HTTP {response.status_code}: {response.text}")
        
        error_code = error_data.get("error", {}).get("code", "UNKNOWN_ERROR")
        error_message = error_data.get("error", {}).get("message", "Unknown error")
        
        # Create appropriate exception based on error code
        if error_code == "VALIDATION_ERROR":
            raise ValidationError(error_message, error_data)
        elif error_code == "AUTHENTICATION_ERROR":
            raise AuthenticationError(error_message, error_data)
        elif error_code == "AUTHORIZATION_ERROR":
            raise AuthorizationError(error_message, error_data)
        elif error_code == "NOT_FOUND":
            raise NotFoundError(error_message, error_data)
        elif error_code == "RATE_LIMIT_EXCEEDED":
            raise RateLimitError(error_message, error_data)
        elif error_code == "SERVICE_UNAVAILABLE":
            raise ServiceUnavailableError(error_message, error_data)
        elif error_code == "REQUEST_TIMEOUT":
            raise TimeoutError(error_message, error_data)
        elif error_code == "INTERNAL_ERROR":
            raise InternalError(error_message, error_data)
        else:
            raise APIError(error_message, error_data, response.status_code)
    
    def close(self) -> None:
        """Close the client session."""
        self.session.close()
    
    def __enter__(self):
        """Context manager entry."""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()
