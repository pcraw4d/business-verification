"""
Merchant Portfolio API SDK - Python

A comprehensive SDK for interacting with the Merchant Portfolio Management API

Version: 1.0.0
Author: KYB Platform Team
"""

import requests
import time
import json
from typing import Dict, List, Optional, Any, Union
from dataclasses import dataclass
from datetime import datetime, timedelta
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@dataclass
class APIError(Exception):
    """Custom API Error class"""
    message: str
    status: int
    code: str
    details: Optional[str] = None
    
    def __str__(self):
        return f"APIError({self.status}): {self.message}"


class MerchantPortfolioAPI:
    """Merchant Portfolio API Client"""
    
    def __init__(self, config: Dict[str, Any] = None):
        """
        Initialize the API client
        
        Args:
            config: Configuration dictionary with baseURL, apiKey, timeout, etc.
        """
        config = config or {}
        self.base_url = config.get('baseURL', 'https://api.kyb-platform.com/v1')
        self.api_key = config.get('apiKey')
        self.timeout = config.get('timeout', 30)
        self.retry_attempts = config.get('retryAttempts', 3)
        self.retry_delay = config.get('retryDelay', 1)
        
        # Initialize session
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {self.api_key}',
            'Content-Type': 'application/json',
            'User-Agent': 'MerchantPortfolioSDK/1.0.0'
        })
        
        # Setup logging
        self.logger = logger
    
    def _make_request(self, method: str, endpoint: str, data: Dict = None, params: Dict = None) -> Dict:
        """
        Make HTTP request with retry logic
        
        Args:
            method: HTTP method
            endpoint: API endpoint
            data: Request body data
            params: Query parameters
            
        Returns:
            Response data as dictionary
            
        Raises:
            APIError: If request fails
        """
        url = f"{self.base_url}{endpoint}"
        
        for attempt in range(self.retry_attempts):
            try:
                self.logger.info(f"[API] {method.upper()} {endpoint}")
                
                response = self.session.request(
                    method=method,
                    url=url,
                    json=data,
                    params=params,
                    timeout=self.timeout
                )
                
                self.logger.info(f"[API] {response.status_code} {endpoint}")
                
                if response.status_code < 400:
                    return response.json() if response.content else {}
                
                # Handle error responses
                error_data = response.json() if response.content else {}
                error_info = error_data.get('error', {})
                
                raise APIError(
                    message=error_info.get('message', 'API request failed'),
                    status=response.status_code,
                    code=error_info.get('code', 'UNKNOWN_ERROR'),
                    details=error_info.get('details')
                )
                
            except requests.exceptions.RequestException as e:
                if attempt < self.retry_attempts - 1:
                    delay = self.retry_delay * (2 ** attempt)
                    self.logger.warning(f"[API] Request failed, retrying in {delay}s (attempt {attempt + 1})")
                    time.sleep(delay)
                    continue
                else:
                    raise APIError(
                        message=f"Network error: {str(e)}",
                        status=0,
                        code='NETWORK_ERROR'
                    )
    
    # ============================================================================
    # MERCHANT CRUD OPERATIONS
    # ============================================================================
    
    def create_merchant(self, merchant_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Create a new merchant
        
        Args:
            merchant_data: Merchant data dictionary
            
        Returns:
            Created merchant data
            
        Raises:
            APIError: If creation fails
        """
        try:
            return self._make_request('POST', '/merchants', data=merchant_data)
        except APIError as e:
            raise APIError(f"Failed to create merchant: {e.message}", e.status, e.code, e.details)
    
    def get_merchant(self, merchant_id: str) -> Dict[str, Any]:
        """
        Get a merchant by ID
        
        Args:
            merchant_id: Merchant ID
            
        Returns:
            Merchant data
            
        Raises:
            APIError: If retrieval fails
        """
        try:
            return self._make_request('GET', f'/merchants/{merchant_id}')
        except APIError as e:
            raise APIError(f"Failed to get merchant: {e.message}", e.status, e.code, e.details)
    
    def update_merchant(self, merchant_id: str, update_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Update a merchant
        
        Args:
            merchant_id: Merchant ID
            update_data: Update data dictionary
            
        Returns:
            Updated merchant data
            
        Raises:
            APIError: If update fails
        """
        try:
            return self._make_request('PUT', f'/merchants/{merchant_id}', data=update_data)
        except APIError as e:
            raise APIError(f"Failed to update merchant: {e.message}", e.status, e.code, e.details)
    
    def delete_merchant(self, merchant_id: str) -> None:
        """
        Delete a merchant
        
        Args:
            merchant_id: Merchant ID
            
        Raises:
            APIError: If deletion fails
        """
        try:
            self._make_request('DELETE', f'/merchants/{merchant_id}')
        except APIError as e:
            raise APIError(f"Failed to delete merchant: {e.message}", e.status, e.code, e.details)
    
    # ============================================================================
    # SEARCH AND LISTING
    # ============================================================================
    
    def list_merchants(self, filters: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        List merchants with optional filters
        
        Args:
            filters: Search filters dictionary
            
        Returns:
            Paginated merchant list
            
        Raises:
            APIError: If listing fails
        """
        try:
            return self._make_request('GET', '/merchants', params=filters or {})
        except APIError as e:
            raise APIError(f"Failed to list merchants: {e.message}", e.status, e.code, e.details)
    
    def search_merchants(self, search_criteria: Dict[str, Any]) -> Dict[str, Any]:
        """
        Advanced search for merchants
        
        Args:
            search_criteria: Search criteria dictionary
            
        Returns:
            Search results
            
        Raises:
            APIError: If search fails
        """
        try:
            return self._make_request('POST', '/merchants/search', data=search_criteria)
        except APIError as e:
            raise APIError(f"Failed to search merchants: {e.message}", e.status, e.code, e.details)
    
    def get_all_merchants(self, filters: Dict[str, Any] = None) -> List[Dict[str, Any]]:
        """
        Get all merchants with automatic pagination
        
        Args:
            filters: Search filters dictionary
            
        Returns:
            List of all merchants
            
        Raises:
            APIError: If retrieval fails
        """
        all_merchants = []
        page = 1
        has_more = True
        
        while has_more:
            response = self.list_merchants({
                **(filters or {}),
                'page': page,
                'page_size': 100
            })
            
            all_merchants.extend(response['merchants'])
            has_more = response['has_more']
            page += 1
            
            # Add small delay to respect rate limits
            if has_more:
                time.sleep(0.1)
        
        return all_merchants
    
    # ============================================================================
    # BULK OPERATIONS
    # ============================================================================
    
    def bulk_update_portfolio_type(self, merchant_ids: List[str], portfolio_type: str) -> Dict[str, Any]:
        """
        Bulk update portfolio type
        
        Args:
            merchant_ids: List of merchant IDs
            portfolio_type: New portfolio type
            
        Returns:
            Bulk operation result
            
        Raises:
            APIError: If bulk update fails
        """
        try:
            return self._make_request('POST', '/merchants/bulk/portfolio-type', data={
                'merchant_ids': merchant_ids,
                'portfolio_type': portfolio_type
            })
        except APIError as e:
            raise APIError(f"Failed to bulk update portfolio type: {e.message}", e.status, e.code, e.details)
    
    def bulk_update_risk_level(self, merchant_ids: List[str], risk_level: str) -> Dict[str, Any]:
        """
        Bulk update risk level
        
        Args:
            merchant_ids: List of merchant IDs
            risk_level: New risk level
            
        Returns:
            Bulk operation result
            
        Raises:
            APIError: If bulk update fails
        """
        try:
            return self._make_request('POST', '/merchants/bulk/risk-level', data={
                'merchant_ids': merchant_ids,
                'risk_level': risk_level
            })
        except APIError as e:
            raise APIError(f"Failed to bulk update risk level: {e.message}", e.status, e.code, e.details)
    
    def get_bulk_operation_status(self, operation_id: str) -> Dict[str, Any]:
        """
        Get bulk operation status
        
        Args:
            operation_id: Operation ID
            
        Returns:
            Operation status
            
        Raises:
            APIError: If status retrieval fails
        """
        try:
            return self._make_request('GET', f'/merchants/bulk/{operation_id}')
        except APIError as e:
            raise APIError(f"Failed to get bulk operation status: {e.message}", e.status, e.code, e.details)
    
    def wait_for_bulk_operation(self, operation_id: str, poll_interval: int = 2) -> Dict[str, Any]:
        """
        Wait for bulk operation to complete
        
        Args:
            operation_id: Operation ID
            poll_interval: Polling interval in seconds
            
        Returns:
            Final operation result
            
        Raises:
            APIError: If operation fails
        """
        while True:
            status = self.get_bulk_operation_status(operation_id)
            
            if status['status'] in ['completed', 'failed']:
                return status
            
            time.sleep(poll_interval)
    
    # ============================================================================
    # SESSION MANAGEMENT
    # ============================================================================
    
    def start_merchant_session(self, merchant_id: str) -> Dict[str, Any]:
        """
        Start a merchant session
        
        Args:
            merchant_id: Merchant ID
            
        Returns:
            Session data
            
        Raises:
            APIError: If session start fails
        """
        try:
            return self._make_request('POST', f'/merchants/{merchant_id}/session')
        except APIError as e:
            raise APIError(f"Failed to start merchant session: {e.message}", e.status, e.code, e.details)
    
    def end_merchant_session(self, merchant_id: str) -> None:
        """
        End a merchant session
        
        Args:
            merchant_id: Merchant ID
            
        Raises:
            APIError: If session end fails
        """
        try:
            self._make_request('DELETE', f'/merchants/{merchant_id}/session')
        except APIError as e:
            raise APIError(f"Failed to end merchant session: {e.message}", e.status, e.code, e.details)
    
    def get_active_session(self) -> Optional[Dict[str, Any]]:
        """
        Get active merchant session
        
        Returns:
            Active session data or None
            
        Raises:
            APIError: If session retrieval fails
        """
        try:
            return self._make_request('GET', '/merchants/session/active')
        except APIError as e:
            if e.status == 404:
                return None
            raise APIError(f"Failed to get active session: {e.message}", e.status, e.code, e.details)
    
    # ============================================================================
    # ANALYTICS AND REPORTING
    # ============================================================================
    
    def get_analytics(self, filters: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Get merchant analytics
        
        Args:
            filters: Analytics filters dictionary
            
        Returns:
            Analytics data
            
        Raises:
            APIError: If analytics retrieval fails
        """
        try:
            return self._make_request('GET', '/merchants/analytics', params=filters or {})
        except APIError as e:
            raise APIError(f"Failed to get analytics: {e.message}", e.status, e.code, e.details)
    
    def get_portfolio_types(self) -> Dict[str, Any]:
        """
        Get portfolio types
        
        Returns:
            Portfolio types list
            
        Raises:
            APIError: If retrieval fails
        """
        try:
            return self._make_request('GET', '/merchants/portfolio-types')
        except APIError as e:
            raise APIError(f"Failed to get portfolio types: {e.message}", e.status, e.code, e.details)
    
    def get_risk_levels(self) -> Dict[str, Any]:
        """
        Get risk levels
        
        Returns:
            Risk levels list
            
        Raises:
            APIError: If retrieval fails
        """
        try:
            return self._make_request('GET', '/merchants/risk-levels')
        except APIError as e:
            raise APIError(f"Failed to get risk levels: {e.message}", e.status, e.code, e.details)
    
    def get_statistics(self) -> Dict[str, Any]:
        """
        Get merchant statistics
        
        Returns:
            Statistics data
            
        Raises:
            APIError: If statistics retrieval fails
        """
        try:
            return self._make_request('GET', '/merchants/statistics')
        except APIError as e:
            raise APIError(f"Failed to get statistics: {e.message}", e.status, e.code, e.details)
    
    # ============================================================================
    # UTILITY METHODS
    # ============================================================================
    
    def validate_merchant_data(self, data: Dict[str, Any]) -> List[str]:
        """
        Validate merchant data
        
        Args:
            data: Merchant data dictionary
            
        Returns:
            List of validation errors
        """
        errors = []
        
        if not data.get('name') or not data['name'].strip():
            errors.append('Name is required')
        
        valid_portfolio_types = ['onboarded', 'deactivated', 'prospective', 'pending']
        if data.get('portfolio_type') not in valid_portfolio_types:
            errors.append('Invalid portfolio type')
        
        valid_risk_levels = ['high', 'medium', 'low']
        if data.get('risk_level') not in valid_risk_levels:
            errors.append('Invalid risk level')
        
        email = data.get('contact_info', {}).get('email')
        if email and not self._is_valid_email(email):
            errors.append('Invalid email format')
        
        return errors
    
    def _is_valid_email(self, email: str) -> bool:
        """Check if email format is valid"""
        import re
        pattern = r'^[^\s@]+@[^\s@]+\.[^\s@]+$'
        return re.match(pattern, email) is not None
    
    def create_merchant_with_validation(self, merchant_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Create merchant with validation
        
        Args:
            merchant_data: Merchant data dictionary
            
        Returns:
            Created merchant data
            
        Raises:
            APIError: If validation fails or creation fails
        """
        errors = self.validate_merchant_data(merchant_data)
        if errors:
            raise APIError(f"Validation failed: {', '.join(errors)}", 400, 'VALIDATION_ERROR')
        
        return self.create_merchant(merchant_data)
    
    def batch_create_merchants(self, merchants: List[Dict[str, Any]], batch_size: int = 10) -> List[Dict[str, Any]]:
        """
        Batch create merchants
        
        Args:
            merchants: List of merchant data dictionaries
            batch_size: Batch size
            
        Returns:
            List of results
            
        Raises:
            APIError: If batch creation fails
        """
        results = []
        
        for i in range(0, len(merchants), batch_size):
            batch = merchants[i:i + batch_size]
            
            self.logger.info(f"Processing batch {i // batch_size + 1} of {len(merchants) // batch_size + 1}")
            
            batch_results = []
            for merchant in batch:
                try:
                    result = self.create_merchant(merchant)
                    batch_results.append({
                        'success': True,
                        'merchant': merchant.get('name'),
                        'result': result
                    })
                except APIError as e:
                    batch_results.append({
                        'success': False,
                        'merchant': merchant.get('name'),
                        'error': str(e)
                    })
            
            results.extend(batch_results)
            
            # Add delay between batches
            if i + batch_size < len(merchants):
                time.sleep(1)
        
        return results


class MerchantSessionManager:
    """Session Manager for handling merchant sessions"""
    
    def __init__(self, api_client: MerchantPortfolioAPI):
        """
        Initialize session manager
        
        Args:
            api_client: MerchantPortfolioAPI instance
        """
        self.api = api_client
        self.active_session = None
        self.logger = logger
    
    def start_session(self, merchant_id: str) -> Dict[str, Any]:
        """
        Start a new session (ends current session if exists)
        
        Args:
            merchant_id: Merchant ID
            
        Returns:
            Session data
            
        Raises:
            APIError: If session start fails
        """
        # End current session if exists
        if self.active_session:
            self.end_current_session()
        
        self.active_session = self.api.start_merchant_session(merchant_id)
        self.logger.info(f"Started session for merchant: {self.active_session['merchant_name']}")
        return self.active_session
    
    def end_current_session(self) -> None:
        """
        End the current session
        
        Raises:
            APIError: If session end fails
        """
        if not self.active_session:
            return
        
        self.api.end_merchant_session(self.active_session['merchant_id'])
        self.logger.info(f"Ended session for merchant: {self.active_session['merchant_name']}")
        self.active_session = None
    
    def get_active_session(self) -> Optional[Dict[str, Any]]:
        """
        Get the current active session
        
        Returns:
            Active session data or None
        """
        return self.active_session
    
    def has_active_session(self) -> bool:
        """
        Check if there's an active session
        
        Returns:
            True if session is active
        """
        return self.active_session is not None


class APICache:
    """Cache manager for API responses"""
    
    def __init__(self, ttl: int = 300):
        """
        Initialize cache
        
        Args:
            ttl: Time to live in seconds (default: 5 minutes)
        """
        self.cache = {}
        self.ttl = ttl
    
    def get(self, key: str) -> Optional[Any]:
        """
        Get cached data
        
        Args:
            key: Cache key
            
        Returns:
            Cached data or None
        """
        if key not in self.cache:
            return None
        
        item = self.cache[key]
        if datetime.now() > item['expiry']:
            del self.cache[key]
            return None
        
        return item['data']
    
    def set(self, key: str, data: Any) -> None:
        """
        Set cached data
        
        Args:
            key: Cache key
            data: Data to cache
        """
        self.cache[key] = {
            'data': data,
            'expiry': datetime.now() + timedelta(seconds=self.ttl)
        }
    
    def clear(self) -> None:
        """Clear all cached data"""
        self.cache.clear()
    
    def cleanup(self) -> None:
        """Clear expired entries"""
        now = datetime.now()
        expired_keys = [
            key for key, item in self.cache.items()
            if now > item['expiry']
        ]
        for key in expired_keys:
            del self.cache[key]


# Example usage
if __name__ == '__main__':
    # Example usage
    api = MerchantPortfolioAPI({
        'baseURL': 'https://api.kyb-platform.com/v1',
        'apiKey': 'your-api-key-here'
    })
    
    session_manager = MerchantSessionManager(api)
    
    async def example():
        try:
            # Create a new merchant
            merchant = api.create_merchant_with_validation({
                'name': 'Example Corp',
                'legal_name': 'Example Corporation LLC',
                'industry': 'Technology',
                'portfolio_type': 'prospective',
                'risk_level': 'medium',
                'address': {
                    'street1': '123 Example St',
                    'city': 'San Francisco',
                    'state': 'CA',
                    'postal_code': '94105',
                    'country': 'United States',
                    'country_code': 'US'
                },
                'contact_info': {
                    'email': 'contact@example.com',
                    'phone': '+1-555-123-4567'
                }
            })
            
            print('Created merchant:', merchant)
            
            # Start a session
            session_manager.start_session(merchant['id'])
            
            # Get analytics
            analytics = api.get_analytics()
            print('Analytics:', analytics)
            
            # End session
            session_manager.end_current_session()
            
        except APIError as e:
            print(f'Error: {e}')
    
    example()
