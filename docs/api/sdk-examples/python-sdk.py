"""
KYB Platform Python SDK Example

This is a practical example of how to use the KYB Platform API
with Python. This example demonstrates:
- Authentication
- Business Classification
- Risk Assessment
- Compliance Checking
- Error Handling
- Rate Limiting
"""

import asyncio
import json
import logging
import time
from typing import Dict, List, Optional, Any
from dataclasses import dataclass
import aiohttp
import backoff


@dataclass
class KYBConfig:
    """Configuration for KYB Platform SDK"""
    base_url: str = "https://api.kybplatform.com"
    access_token: Optional[str] = None
    refresh_token: Optional[str] = None
    timeout: int = 30
    max_retries: int = 3


class KYBPlatformError(Exception):
    """Base exception for KYB Platform SDK"""
    def __init__(self, message: str, status_code: Optional[int] = None, data: Optional[Dict] = None):
        super().__init__(message)
        self.status_code = status_code
        self.data = data or {}


class KYBPlatformSDK:
    """KYB Platform Python SDK"""
    
    def __init__(self, config: KYBConfig):
        self.config = config
        self.logger = logging.getLogger(__name__)
        
        # Rate limiting
        self.requests = []
        self.max_requests = 100
        self.window_seconds = 60
        
        # Session for HTTP requests
        self.session: Optional[aiohttp.ClientSession] = None
    
    async def __aenter__(self):
        """Async context manager entry"""
        self.session = aiohttp.ClientSession(
            timeout=aiohttp.ClientTimeout(total=self.config.timeout)
        )
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        if self.session:
            await self.session.close()
    
    def _check_rate_limit(self):
        """Check rate limiting"""
        now = time.time()
        self.requests = [req_time for req_time in self.requests 
                        if now - req_time < self.window_seconds]
        
        if len(self.requests) >= self.max_requests:
            oldest_request = self.requests[0]
            wait_time = self.window_seconds - (now - oldest_request)
            raise KYBPlatformError(f"Rate limit exceeded. Wait {wait_time:.1f} seconds")
    
    @backoff.on_exception(
        backoff.expo,
        (aiohttp.ClientError, asyncio.TimeoutError),
        max_tries=3
    )
    async def _make_request(self, endpoint: str, method: str = "GET", 
                           data: Optional[Dict] = None) -> Dict[str, Any]:
        """Make an authenticated HTTP request"""
        self._check_rate_limit()
        
        url = f"{self.config.base_url}/v1{endpoint}"
        headers = {"Content-Type": "application/json"}
        
        if self.config.access_token:
            headers["Authorization"] = f"Bearer {self.config.access_token}"
        
        try:
            async with self.session.request(
                method=method,
                url=url,
                headers=headers,
                json=data
            ) as response:
                # Record request for rate limiting
                self.requests.append(time.time())
                
                response_data = await response.json()
                
                if not response.ok:
                    raise KYBPlatformError(
                        message=response_data.get("message", f"HTTP {response.status}"),
                        status_code=response.status,
                        data=response_data
                    )
                
                return response_data
                
        except aiohttp.ClientError as e:
            self.logger.error(f"HTTP request failed: {e}")
            raise KYBPlatformError(f"HTTP request failed: {e}")
    
    async def refresh_access_token(self) -> str:
        """Refresh access token"""
        if not self.config.refresh_token:
            raise KYBPlatformError("No refresh token available")
        
        try:
            response = await self._make_request(
                endpoint="/auth/refresh",
                method="POST",
                data={"refresh_token": self.config.refresh_token}
            )
            
            self.config.access_token = response["access_token"]
            self.config.refresh_token = response["refresh_token"]
            
            self.logger.info("Access token refreshed successfully")
            return self.config.access_token
            
        except Exception as e:
            self.logger.error(f"Failed to refresh access token: {e}")
            raise
    
    class Auth:
        """Authentication methods"""
        
        def __init__(self, sdk: 'KYBPlatformSDK'):
            self.sdk = sdk
        
        async def login(self, email: str, password: str) -> Dict[str, Any]:
            """Login with email and password"""
            response = await self.sdk._make_request(
                endpoint="/auth/login",
                method="POST",
                data={"email": email, "password": password}
            )
            
            self.sdk.config.access_token = response["access_token"]
            self.sdk.config.refresh_token = response["refresh_token"]
            
            self.sdk.logger.info("Login successful")
            return response
        
        async def register(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
            """Register a new user"""
            return await self.sdk._make_request(
                endpoint="/auth/register",
                method="POST",
                data=user_data
            )
        
        async def logout(self) -> None:
            """Logout"""
            try:
                await self.sdk._make_request(
                    endpoint="/auth/logout",
                    method="POST"
                )
            finally:
                self.sdk.config.access_token = None
                self.sdk.config.refresh_token = None
                self.sdk.logger.info("Logout successful")
    
    class Classification:
        """Business classification methods"""
        
        def __init__(self, sdk: 'KYBPlatformSDK'):
            self.sdk = sdk
        
        async def classify(self, business_data: Dict[str, Any]) -> Dict[str, Any]:
            """Classify a single business"""
            self.sdk.logger.info("Starting business classification", 
                               extra={"business_name": business_data.get("business_name")})
            
            try:
                result = await self.sdk._make_request(
                    endpoint="/classify",
                    method="POST",
                    data=business_data
                )
                
                self.sdk.logger.info("Classification completed", extra={
                    "business_name": business_data.get("business_name"),
                    "classification_id": result.get("classification_id"),
                    "confidence_score": result.get("confidence_score")
                })
                
                return result
                
            except Exception as e:
                self.sdk.logger.error("Classification failed", extra={
                    "business_name": business_data.get("business_name"),
                    "error": str(e)
                })
                raise
        
        async def batch_classify(self, businesses: List[Dict[str, Any]]) -> Dict[str, Any]:
            """Batch classify multiple businesses"""
            self.sdk.logger.info("Starting batch classification", 
                               extra={"count": len(businesses)})
            
            try:
                result = await self.sdk._make_request(
                    endpoint="/classify/batch",
                    method="POST",
                    data={"businesses": businesses}
                )
                
                successful = sum(1 for r in result.get("results", []) if r.get("success"))
                failed = len(result.get("results", [])) - successful
                
                self.sdk.logger.info("Batch classification completed", extra={
                    "count": len(businesses),
                    "successful": successful,
                    "failed": failed
                })
                
                return result
                
            except Exception as e:
                self.sdk.logger.error("Batch classification failed", extra={
                    "count": len(businesses),
                    "error": str(e)
                })
                raise
        
        async def get_history(self, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
            """Get classification history"""
            endpoint = "/classify/history"
            if params:
                query_string = "&".join(f"{k}={v}" for k, v in params.items())
                endpoint += f"?{query_string}"
            
            return await self.sdk._make_request(endpoint)
        
        async def generate_confidence_report(self, params: Dict[str, Any]) -> Dict[str, Any]:
            """Generate confidence report"""
            return await self.sdk._make_request(
                endpoint="/classify/confidence-report",
                method="POST",
                data=params
            )
    
    class Risk:
        """Risk assessment methods"""
        
        def __init__(self, sdk: 'KYBPlatformSDK'):
            self.sdk = sdk
        
        async def assess(self, risk_data: Dict[str, Any]) -> Dict[str, Any]:
            """Assess business risk"""
            self.sdk.logger.info("Starting risk assessment", extra={
                "business_id": risk_data.get("business_id"),
                "categories": risk_data.get("categories")
            })
            
            try:
                result = await self.sdk._make_request(
                    endpoint="/risk/assess",
                    method="POST",
                    data=risk_data
                )
                
                self.sdk.logger.info("Risk assessment completed", extra={
                    "business_id": risk_data.get("business_id"),
                    "overall_score": result.get("overall_score"),
                    "overall_level": result.get("overall_level")
                })
                
                return result
                
            except Exception as e:
                self.sdk.logger.error("Risk assessment failed", extra={
                    "business_id": risk_data.get("business_id"),
                    "error": str(e)
                })
                raise
        
        async def get_categories(self) -> Dict[str, Any]:
            """Get risk categories"""
            return await self.sdk._make_request("/risk/categories")
        
        async def get_factors(self, category: Optional[str] = None) -> Dict[str, Any]:
            """Get risk factors"""
            endpoint = "/risk/factors"
            if category:
                endpoint += f"?category={category}"
            
            return await self.sdk._make_request(endpoint)
        
        async def get_thresholds(self, category: Optional[str] = None) -> Dict[str, Any]:
            """Get risk thresholds"""
            endpoint = "/risk/thresholds"
            if category:
                endpoint += f"?category={category}"
            
            return await self.sdk._make_request(endpoint)
    
    class Compliance:
        """Compliance checking methods"""
        
        def __init__(self, sdk: 'KYBPlatformSDK'):
            self.sdk = sdk
        
        async def check(self, compliance_data: Dict[str, Any]) -> Dict[str, Any]:
            """Check compliance status"""
            self.sdk.logger.info("Starting compliance check", extra={
                "business_id": compliance_data.get("business_id"),
                "frameworks": compliance_data.get("frameworks")
            })
            
            try:
                result = await self.sdk._make_request(
                    endpoint="/compliance/check",
                    method="POST",
                    data=compliance_data
                )
                
                self.sdk.logger.info("Compliance check completed", extra={
                    "business_id": compliance_data.get("business_id"),
                    "frameworks_checked": len(result.get("frameworks", []))
                })
                
                return result
                
            except Exception as e:
                self.sdk.logger.error("Compliance check failed", extra={
                    "business_id": compliance_data.get("business_id"),
                    "error": str(e)
                })
                raise
        
        async def get_status(self, business_id: str) -> Dict[str, Any]:
            """Get compliance status"""
            return await self.sdk._make_request(f"/compliance/status/{business_id}")
        
        async def generate_report(self, params: Dict[str, Any]) -> Dict[str, Any]:
            """Generate compliance report"""
            return await self.sdk._make_request(
                endpoint="/compliance/report",
                method="POST",
                data=params
            )
    
    def __init__(self, config: KYBConfig):
        self.config = config
        self.logger = logging.getLogger(__name__)
        
        # Initialize sub-modules
        self.auth = self.Auth(self)
        self.classification = self.Classification(self)
        self.risk = self.Risk(self)
        self.compliance = self.Compliance(self)


async def example():
    """Example usage of the KYB Platform SDK"""
    
    # Configure logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # Initialize the SDK
    config = KYBConfig(
        base_url="https://api.kybplatform.com"
    )
    
    async with KYBPlatformSDK(config) as kyb:
        try:
            # 1. Authenticate
            print("🔐 Authenticating...")
            auth_response = await kyb.auth.login(
                email="your-email@example.com",
                password="your-password"
            )
            print("✅ Authentication successful")
            
            # 2. Classify a business
            print("\n🏢 Classifying business...")
            classification = await kyb.classification.classify({
                "business_name": "Acme Corporation",
                "business_address": "123 Main St, New York, NY 10001",
                "business_phone": "+1-555-123-4567",
                "business_website": "https://acme.com"
            })
            print("✅ Classification completed:", {
                "naics_code": classification.get("naics_code"),
                "confidence_score": classification.get("confidence_score")
            })
            
            # 3. Assess risk
            print("\n⚠️ Assessing risk...")
            risk_assessment = await kyb.risk.assess({
                "business_id": "business-123",
                "business_name": "Acme Corporation",
                "categories": ["financial", "operational"]
            })
            print("✅ Risk assessment completed:", {
                "overall_score": risk_assessment.get("overall_score"),
                "overall_level": risk_assessment.get("overall_level")
            })
            
            # 4. Check compliance
            print("\n📋 Checking compliance...")
            compliance = await kyb.compliance.check({
                "business_id": "business-123",
                "frameworks": ["SOC2", "PCI_DSS", "GDPR"]
            })
            print("✅ Compliance check completed:", {
                "frameworks_checked": len(compliance.get("frameworks", []))
            })
            
            # 5. Batch classification example
            print("\n📦 Performing batch classification...")
            batch_results = await kyb.classification.batch_classify([
                {
                    "business_name": "Tech Solutions LLC",
                    "business_address": "456 Oak Ave, San Francisco, CA"
                },
                {
                    "business_name": "Global Services Inc",
                    "business_address": "789 Pine St, Chicago, IL"
                }
            ])
            
            results = batch_results.get("results", [])
            successful = sum(1 for r in results if r.get("success"))
            
            print("✅ Batch classification completed:", {
                "total": len(results),
                "successful": successful
            })
            
            # 6. Get risk categories
            print("\n📊 Getting risk categories...")
            categories = await kyb.risk.get_categories()
            print("✅ Risk categories:", categories.get("categories", []))
            
            # 7. Get compliance status
            print("\n📈 Getting compliance status...")
            status = await kyb.compliance.get_status("business-123")
            print("✅ Compliance status:", status.get("status"))
            
        except KYBPlatformError as e:
            print(f"❌ KYB Platform Error: {e.message}")
            if e.status_code == 401:
                print("🔄 Attempting to refresh token...")
                try:
                    await kyb.refresh_access_token()
                    print("✅ Token refreshed successfully")
                except Exception as refresh_error:
                    print(f"❌ Token refresh failed: {refresh_error}")
            elif e.status_code == 429:
                print("⏳ Rate limited. Implement retry logic.")
        except Exception as e:
            print(f"❌ Unexpected error: {e}")
        finally:
            # Cleanup
            await kyb.auth.logout()


async def with_retry(func, max_retries: int = 3):
    """Retry function with exponential backoff"""
    for attempt in range(max_retries):
        try:
            return await func()
        except KYBPlatformError as e:
            if e.status_code == 429 and attempt < max_retries - 1:
                delay = 2 ** attempt
                print(f"Rate limited. Waiting {delay} seconds before retry...")
                await asyncio.sleep(delay)
                continue
            raise
        except Exception as e:
            if attempt == max_retries - 1:
                raise
            delay = 2 ** attempt
            print(f"Error occurred. Waiting {delay} seconds before retry...")
            await asyncio.sleep(delay)


# Example of using the SDK with retry logic
async def example_with_retry():
    """Example with retry logic"""
    config = KYBConfig(base_url="https://api.kybplatform.com")
    
    async with KYBPlatformSDK(config) as kyb:
        try:
            # Login with retry
            await with_retry(lambda: kyb.auth.login(
                email="your-email@example.com",
                password="your-password"
            ))
            
            # Classify with retry
            result = await with_retry(lambda: kyb.classification.classify({
                "business_name": "Acme Corporation",
                "business_address": "123 Main St, New York, NY 10001"
            }))
            
            print("✅ Classification with retry completed:", result)
            
        except Exception as e:
            print(f"❌ Error with retry: {e}")


if __name__ == "__main__":
    # Run the example
    asyncio.run(example())
    
    # Run example with retry logic
    # asyncio.run(example_with_retry())
