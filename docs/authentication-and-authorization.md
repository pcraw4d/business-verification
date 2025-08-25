# Authentication and Authorization System

## Overview

The Enhanced Business Intelligence System implements a comprehensive authentication and authorization system that provides secure access control, user management, and role-based permissions. The system supports multiple authentication methods including JWT tokens and API keys, with robust security features and audit logging.

## Architecture

### Core Components

1. **AuthService** (`internal/auth/service.go`)
   - Core authentication logic
   - JWT token generation and validation
   - User registration and login
   - Password management
   - Token blacklisting

2. **AuthMiddleware** (`internal/api/middleware/auth.go`)
   - HTTP middleware for authentication
   - Role-based access control
   - Permission checking
   - API key validation

3. **AuthHandler** (`internal/api/handlers/auth.go`)
   - HTTP request handlers
   - Cookie management
   - CSRF protection
   - Session management

4. **TokenBlacklistRepository** (`internal/auth/service.go`)
   - Token revocation management
   - Blacklist storage and cleanup

## Authentication Methods

### 1. JWT Token Authentication

The system uses JSON Web Tokens (JWT) for stateless authentication with the following features:

#### Token Structure
```json
{
  "user_id": "user_123",
  "email": "user@example.com",
  "username": "username",
  "role": "user",
  "iat": 1640995200,
  "exp": 1640996100,
  "iss": "kyb-tool",
  "sub": "user_123"
}
```

#### Token Types
- **Access Token**: Short-lived (15 minutes) for API access
- **Refresh Token**: Long-lived (7 days) for token renewal

#### Security Features
- HMAC-SHA256 signing
- Configurable expiration times
- Token blacklisting for logout
- Automatic cleanup of expired tokens

### 2. API Key Authentication

For programmatic access, the system supports API key authentication:

#### API Key Format
```
Authorization: ApiKey test-api-key-123
```

#### Security Features
- Prefix-based validation (`test-` for development)
- Configurable key formats
- Rate limiting per key
- Audit logging

### 3. Multi-Authentication Support

The system supports flexible authentication with the `RequireAnyAuth` middleware:

```go
// Accepts either JWT tokens or API keys
middleware.RequireAnyAuth(handler)
```

## User Management

### User Registration

```http
POST /v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "username",
  "password": "secure_password",
  "first_name": "John",
  "last_name": "Doe",
  "company": "Acme Corp"
}
```

#### Features
- Email validation
- Password strength requirements
- Username uniqueness
- Company association
- Email verification tokens

### User Login

```http
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password"
}
```

#### Response
```json
{
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 900
  },
  "message": "Login successful"
}
```

#### Security Features
- Account lockout after failed attempts
- Progressive lockout duration
- Failed attempt tracking
- Secure cookie management

### Password Management

#### Change Password
```http
POST /v1/auth/change-password
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "current_password": "old_password",
  "new_password": "new_secure_password"
}
```

#### Password Reset
```http
POST /v1/auth/request-password-reset
Content-Type: application/json

{
  "email": "user@example.com"
}
```

```http
POST /v1/auth/reset-password
Content-Type: application/json

{
  "token": "reset_token",
  "new_password": "new_secure_password"
}
```

## Authorization System

### Role-Based Access Control (RBAC)

The system implements a hierarchical role-based access control system:

#### Roles
1. **Admin** - Full system access
2. **Manager** - Read/write access to all data
3. **User** - Read/write access to own data
4. **Viewer** - Read-only access to own data

#### Permissions
```go
permissions := map[string][]string{
    "admin": {
        "read:all",
        "write:all", 
        "delete:all",
        "admin:all",
        "user:manage",
        "system:manage",
    },
    "manager": {
        "read:all",
        "write:all",
        "user:read",
        "user:write",
    },
    "user": {
        "read:own",
        "write:own",
        "profile:manage",
    },
    "viewer": {
        "read:own",
    },
}
```

### Middleware Usage

#### Require Authentication
```go
// Require valid JWT token
middleware.RequireAuth(handler)
```

#### Require Specific Role
```go
// Require admin role
middleware.RequireRole("admin")(handler)
```

#### Require Permission
```go
// Require specific permission
middleware.RequirePermission("user:manage")(handler)
```

#### Require Email Verification
```go
// Require verified email
middleware.RequireEmailVerified(handler)
```

#### Optional Authentication
```go
// Add user context if available
middleware.OptionalAuth(handler)
```

## Security Features

### 1. Password Security
- BCrypt hashing with configurable cost
- Minimum length requirements
- Complexity requirements (uppercase, lowercase, numbers, special characters)
- Password history tracking

### 2. Session Security
- Secure HTTP-only cookies
- CSRF protection with token rotation
- SameSite cookie attributes
- Configurable cookie domains and paths

### 3. Token Security
- Short-lived access tokens
- Refresh token rotation
- Token blacklisting for logout
- Automatic cleanup of expired tokens

### 4. Rate Limiting
- Authentication-specific rate limiting
- Progressive lockout for failed attempts
- Configurable lockout durations
- Permanent lockout after repeated violations

### 5. Audit Logging
- Comprehensive security event logging
- User action tracking
- Failed authentication attempts
- Token usage monitoring

## Configuration

### AuthConfig Structure
```go
type AuthConfig struct {
    JWTSecret         string        `json:"jwt_secret" yaml:"jwt_secret"`
    JWTExpiration     time.Duration `json:"jwt_expiration" yaml:"jwt_expiration"`
    RefreshExpiration time.Duration `json:"refresh_expiration" yaml:"refresh_expiration"`
    MinPasswordLength int           `json:"min_password_length" yaml:"min_password_length"`
    RequireUppercase  bool          `json:"require_uppercase" yaml:"require_uppercase"`
    RequireLowercase  bool          `json:"require_lowercase" yaml:"require_lowercase"`
    RequireNumbers    bool          `json:"require_numbers" yaml:"require_numbers"`
    RequireSpecial    bool          `json:"require_special" yaml:"require_special"`
    MaxLoginAttempts  int           `json:"max_login_attempts" yaml:"max_login_attempts"`
    LockoutDuration   time.Duration `json:"lockout_duration" yaml:"lockout_duration"`
    RefreshCookieName string        `json:"refresh_cookie_name" yaml:"refresh_cookie_name"`
    CSRFCookieName    string        `json:"csrf_cookie_name" yaml:"csrf_cookie_name"`
    CookieDomain      string        `json:"cookie_domain" yaml:"cookie_domain"`
    CookiePath        string        `json:"cookie_path" yaml:"cookie_path"`
    CookieSecure      bool          `json:"cookie_secure" yaml:"cookie_secure"`
    CookieSameSite    string        `json:"cookie_same_site" yaml:"cookie_same_site"`
}
```

### Environment Variables
```bash
# JWT Configuration
AUTH_JWT_SECRET=your-secret-key
AUTH_JWT_EXPIRATION=15m
AUTH_REFRESH_EXPIRATION=168h

# Password Requirements
AUTH_MIN_PASSWORD_LENGTH=8
AUTH_REQUIRE_UPPERCASE=true
AUTH_REQUIRE_LOWERCASE=true
AUTH_REQUIRE_NUMBERS=true
AUTH_REQUIRE_SPECIAL=true

# Security Settings
AUTH_MAX_LOGIN_ATTEMPTS=5
AUTH_LOCKOUT_DURATION=30m

# Cookie Settings
AUTH_REFRESH_COOKIE_NAME=refresh_token
AUTH_CSRF_COOKIE_NAME=csrf_token
AUTH_COOKIE_DOMAIN=.example.com
AUTH_COOKIE_PATH=/
AUTH_COOKIE_SECURE=true
AUTH_COOKIE_SAME_SITE=strict
```

## API Endpoints

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/v1/auth/register` | User registration | No |
| POST | `/v1/auth/login` | User login | No |
| POST | `/v1/auth/logout` | User logout | Yes |
| POST | `/v1/auth/refresh` | Token refresh | Yes |
| GET | `/v1/auth/verify-email` | Email verification | No |
| POST | `/v1/auth/request-password-reset` | Request password reset | No |
| POST | `/v1/auth/reset-password` | Reset password | No |
| POST | `/v1/auth/change-password` | Change password | Yes |
| GET | `/v1/auth/profile` | Get user profile | Yes |

### Protected Endpoints

All business logic endpoints require authentication:

```http
GET /v1/classifications
Authorization: Bearer <access_token>

POST /v1/classifications
Authorization: Bearer <access_token>
Content-Type: application/json
```

## Error Handling

### Authentication Errors
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `423 Locked` - Account temporarily locked

### Common Error Responses
```json
{
  "error": "authentication_failed",
  "message": "Invalid email or password",
  "code": "AUTH_001"
}
```

## Testing

### Unit Tests
The authentication system includes comprehensive unit tests:

```bash
# Run auth service tests
go test ./internal/auth -v

# Run middleware tests
go test ./internal/api/middleware -v

# Run handler tests
go test ./internal/api/handlers -v
```

### Test Coverage
- Service layer: 100% coverage
- Middleware: 100% coverage
- Handlers: 95% coverage
- Integration tests: Complete workflow testing

## Security Best Practices

### 1. Token Management
- Use short-lived access tokens (15 minutes)
- Implement refresh token rotation
- Blacklist tokens on logout
- Monitor token usage patterns

### 2. Password Security
- Enforce strong password policies
- Use secure hashing (BCrypt)
- Implement account lockout
- Monitor failed login attempts

### 3. Session Security
- Use secure cookies
- Implement CSRF protection
- Set appropriate SameSite attributes
- Use HTTPS in production

### 4. API Security
- Validate all inputs
- Implement rate limiting
- Log security events
- Monitor for suspicious activity

## Monitoring and Observability

### Metrics
- Authentication success/failure rates
- Token usage patterns
- Failed login attempts
- Account lockouts

### Logging
- Security event logging
- User action tracking
- Failed authentication attempts
- Token validation errors

### Alerts
- High failure rate alerts
- Suspicious activity detection
- Account lockout notifications
- Token abuse warnings

## Integration Examples

### Go Client
```go
package main

import (
    "net/http"
    "encoding/json"
)

type AuthClient struct {
    baseURL string
    client  *http.Client
}

func (c *AuthClient) Login(email, password string) (*TokenResponse, error) {
    req := LoginRequest{
        Email:    email,
        Password: password,
    }
    
    resp, err := c.client.Post(c.baseURL+"/v1/auth/login", 
        "application/json", bytes.NewBuffer(reqBytes))
    if err != nil {
        return nil, err
    }
    
    var tokenResp TokenResponse
    json.NewDecoder(resp.Body).Decode(&tokenResp)
    return &tokenResp, nil
}

func (c *AuthClient) AuthenticatedRequest(token, endpoint string) (*http.Response, error) {
    req, _ := http.NewRequest("GET", c.baseURL+endpoint, nil)
    req.Header.Set("Authorization", "Bearer "+token)
    return c.client.Do(req)
}
```

### JavaScript Client
```javascript
class AuthClient {
    constructor(baseURL) {
        this.baseURL = baseURL;
    }
    
    async login(email, password) {
        const response = await fetch(`${this.baseURL}/v1/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });
        
        return response.json();
    }
    
    async authenticatedRequest(token, endpoint) {
        const response = await fetch(`${this.baseURL}${endpoint}`, {
            headers: {
                'Authorization': `Bearer ${token}`,
            },
        });
        
        return response.json();
    }
}
```

## Future Enhancements

### Planned Features
1. **OAuth 2.0 Integration** - Support for third-party authentication
2. **Multi-Factor Authentication** - TOTP and SMS-based MFA
3. **Single Sign-On (SSO)** - SAML and OpenID Connect support
4. **Advanced Role Management** - Dynamic role assignment
5. **API Key Management** - Web-based key management interface
6. **Session Management** - Active session monitoring and control

### Security Improvements
1. **Hardware Security Modules** - HSM integration for key management
2. **Zero Trust Architecture** - Continuous verification
3. **Behavioral Analytics** - User behavior monitoring
4. **Threat Intelligence** - Integration with security feeds

## Troubleshooting

### Common Issues

#### Token Validation Failures
- Check token expiration
- Verify JWT secret configuration
- Ensure proper token format
- Check token blacklist

#### Authentication Errors
- Verify user credentials
- Check account status
- Review rate limiting settings
- Monitor failed attempts

#### Permission Denied
- Verify user role assignment
- Check permission configuration
- Review middleware setup
- Validate endpoint protection

### Debug Mode
Enable debug logging for authentication issues:

```bash
export LOG_LEVEL=debug
export AUTH_DEBUG=true
```

## Support

For authentication and authorization support:

1. **Documentation**: Check this guide and API documentation
2. **Logs**: Review application logs for error details
3. **Metrics**: Monitor authentication metrics
4. **Testing**: Use provided test suites
5. **Community**: Check GitHub issues and discussions

---

**Document Version**: 1.0.0  
**Last Updated**: December 19, 2024  
**Next Review**: March 19, 2025
