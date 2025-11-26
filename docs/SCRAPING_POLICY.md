# Web Scraping Policy

## Overview

The KYB Platform performs automated web scraping for the purpose of business verification and industry classification. This document outlines our compliance measures, legal basis, and practices to ensure respectful and legal web scraping.

## Compliance Measures

### 1. robots.txt Compliance

We respect the `robots.txt` standard and implement comprehensive parsing:

- **Full Parsing**: We use a proper robots.txt parser that checks both specific User-Agent rules and wildcard (`*`) rules
- **Path-Specific Checks**: We test specific paths, not just root-level disallow rules
- **Crawl-Delay Respect**: We honor `Crawl-Delay` directives specified in robots.txt
- **Graceful Degradation**: If robots.txt is unavailable or unparseable, we allow crawling (following the standard that absence of robots.txt means crawling is allowed)

**Implementation**: Our robots.txt parser uses the `github.com/temoto/robotstxt` library to ensure proper compliance with the standard.

### 2. Identifiable User-Agent

We use a clearly identifiable User-Agent string that includes:

- **Bot Identification**: `KYBPlatformBot/1.0`
- **Contact Information**: URL to our bot information page
- **Purpose Statement**: "Business Verification"

**Format**:
```
Mozilla/5.0 (compatible; KYBPlatformBot/1.0; +https://kyb-platform.com/bot-info; Business Verification)
```

The contact URL can be customized via the `SCRAPING_USER_AGENT_CONTACT_URL` environment variable.

### 3. Conservative Rate Limiting

We implement conservative rate limiting to minimize impact on target websites:

- **Default Delay**: 3 seconds minimum between requests to the same domain
- **Configurable**: Delay can be adjusted via `SCRAPING_RATE_LIMIT_DELAY` environment variable (in seconds)
- **Minimum Enforced**: Minimum delay of 2 seconds is always enforced
- **Maximum Limit**: Maximum delay of 10 seconds (configurable)
- **Robots.txt Integration**: If a website specifies a `Crawl-Delay` in robots.txt that is greater than our configured delay, we respect the robots.txt delay
- **Human-Like Timing**: We use statistical distributions (Weibull) to generate human-like timing patterns instead of fixed delays, making our requests appear more natural

### 4. Bot Evasion Techniques

While maintaining full legal compliance and identifiable User-Agent, we implement sophisticated techniques to improve scraping success rates:

#### Header Randomization

- **User-Agent**: Always remains identifiable (`KYBPlatformBot/1.0`) - **never randomized**
- **Other Headers**: We randomize Accept-Language, Accept-Encoding, Sec-Fetch-* headers, and other browser headers to appear more like a real browser
- **Purpose**: Reduces detection while maintaining transparency through our identifiable User-Agent
- **Configuration**: Can be disabled via `SCRAPING_HEADER_RANDOMIZATION_ENABLED` (default: true)

#### Human-Like Timing Patterns

- **Statistical Distributions**: We use Weibull distribution to model human reading/thinking time
- **Variable Delays**: Delays vary naturally rather than being fixed intervals
- **Occasional Pauses**: We add occasional longer pauses (simulating reading time) and shorter bursts (simulating quick navigation)
- **Per-Domain Tracking**: We track timing patterns per domain to avoid detection
- **Configuration**: Can be disabled via `SCRAPING_HUMAN_LIKE_TIMING_ENABLED` (default: true)

#### CAPTCHA Detection

- **Automatic Detection**: We detect various CAPTCHA types (reCAPTCHA, hCaptcha, Cloudflare, Turnstile, generic)
- **Immediate Stop**: When CAPTCHA is detected, we stop scraping immediately
- **No Solving**: We do **not** solve CAPTCHAs by default (requires explicit enablement and legal review)
- **Detection Methods**: We check response headers, body content, and status codes for CAPTCHA indicators
- **Configuration**: Can be disabled via `SCRAPING_CAPTCHA_DETECTION_ENABLED` (default: true)

#### Session/Cookie Management

- **Cookie Persistence**: We maintain cookies across requests to the same domain for realistic session behavior
- **Referer Tracking**: We track referer chains to simulate realistic navigation patterns
- **Session Timeout**: Sessions expire after a configurable time (default: 1 hour)
- **Configuration**: Can be disabled via `SCRAPING_SESSION_MANAGEMENT_ENABLED` (default: true)

#### Optional Proxy Integration

- **Proxy Rotation**: Optional proxy support for additional IP rotation (disabled by default)
- **Health Checking**: Proxies are health-checked and marked as healthy/unhealthy
- **Fallback**: Falls back to direct connection if proxies are unavailable
- **Configuration**: Enabled via `SCRAPING_USE_PROXIES` (default: false) and `SCRAPING_PROXY_LIST` (comma-separated proxy URLs)

**Important**: All evasion techniques respect robots.txt, rate limiting, and maintain our identifiable User-Agent. We remain transparent about our scraping activities.

**Example Configuration**:
```bash
# Set 5-second delay between requests
export SCRAPING_RATE_LIMIT_DELAY=5
```

## Legal Basis

### Public Data Only

We only scrape publicly available information from business websites:

- **Public Websites**: We only access publicly accessible web pages
- **No Authentication**: We do not attempt to access password-protected or authenticated content
- **No Personal Data**: We do not collect personal information (names, emails, phone numbers of individuals)
- **Business Information Only**: We extract keywords and business-related information for verification purposes

### Business Verification Purpose

Our scraping serves a legitimate business purpose:

- **Industry Classification**: We extract keywords to classify businesses into appropriate industry codes (MCC, SIC, NAICS)
- **Business Verification**: We verify business information for Know Your Business (KYB) compliance
- **Minimal Extraction**: We only extract keywords and business-relevant information, not full page content
- **No Competitive Intelligence**: We do not use scraped data for competitive analysis or market research

### Minimal Data Extraction

We practice minimal data extraction:

- **Keywords Only**: We extract keywords and business-related terms
- **No Full Content**: We do not store or cache full page content
- **No Media Files**: We do not download images, videos, or other media files
- **No Deep Crawling**: We limit crawling depth and page count to minimize impact

## Error Handling

### HTTP Status Code Handling

We implement specific handling for important HTTP status codes:

#### 429 (Too Many Requests)

- **Action**: Stop immediately, do not retry
- **Behavior**: Read and log the `Retry-After` header if present
- **Rationale**: The server is explicitly requesting us to slow down or stop

#### 403 (Forbidden)

- **Action**: Stop immediately, do not retry
- **Behavior**: Log the blocked access and return error
- **Rationale**: The server is explicitly denying access

#### 503 (Service Unavailable)

- **Action**: Implement exponential backoff retry
- **Behavior**: Retry up to 3 times with increasing delays (2s, 4s, 6s)
- **Rationale**: Temporary server issues may resolve, but we limit retries to avoid overwhelming the server

### Error Logging

All errors are logged with appropriate context:

- **Status Codes**: Logged with domain and URL information
- **Rate Limiting**: Logged with retry-after information
- **Access Denials**: Logged for monitoring and compliance tracking

## Contact Information

### For Website Owners

If you are a website owner and have concerns about our scraping:

1. **Email**: Contact us at [support@kyb-platform.com](mailto:support@kyb-platform.com)
2. **Bot Information Page**: Visit [https://kyb-platform.com/bot-info](https://kyb-platform.com/bot-info)
3. **Robots.txt**: Add appropriate rules to your robots.txt file - we will respect them
4. **Opt-Out**: Request exclusion via email with your domain name

### Bot Information

- **Bot Name**: KYBPlatformBot
- **Version**: 1.0
- **Purpose**: Business verification and industry classification
- **Contact URL**: [https://kyb-platform.com/bot-info](https://kyb-platform.com/bot-info)

## Data Usage

### What We Extract

- **Keywords**: Business-related keywords and terms
- **Industry Indicators**: Terms that indicate business industry or sector
- **Business Information**: Public business information for verification

### What We Do NOT Extract

- **Personal Data**: Names, emails, phone numbers of individuals
- **Full Page Content**: Complete HTML or text content
- **Media Files**: Images, videos, PDFs, or other files
- **Private Information**: Any information behind authentication

### Data Storage

- **Temporary**: Keywords are processed and used for classification
- **No Long-Term Storage**: We do not maintain long-term caches of scraped content
- **Aggregated Use**: Data is used only for business verification and classification

## Opt-Out Mechanism

### How to Request Exclusion

Website owners can request exclusion from our scraping:

1. **Email Request**: Send an email to [support@kyb-platform.com](mailto:support@kyb-platform.com) with:
   - Subject: "Scraping Opt-Out Request"
   - Domain name(s) to exclude
   - Contact information

2. **Robots.txt**: Add the following to your robots.txt:
   ```
   User-agent: KYBPlatformBot
   Disallow: /
   ```

3. **Response Time**: We will process opt-out requests within 48 hours

### Automatic Respect

We automatically respect:
- **robots.txt Disallow Rules**: If your robots.txt disallows our bot, we will not scrape
- **429 Status Codes**: If you rate-limit us, we stop immediately
- **403 Status Codes**: If you block us, we stop immediately

## Configuration

### Environment Variables

Our scraping behavior can be configured via environment variables:

#### Basic Configuration

- **`SCRAPING_RATE_LIMIT_DELAY`**: Delay between requests in seconds (default: 3, min: 2, max: 10)
- **`SCRAPING_USER_AGENT_CONTACT_URL`**: URL for bot information page (default: https://kyb-platform.com/bot-info)
- **`SCRAPING_RESPECT_ROBOTS`**: Whether to respect robots.txt (default: true)

#### Bot Evasion Configuration

- **`SCRAPING_HEADER_RANDOMIZATION_ENABLED`**: Enable header randomization (default: true)
- **`SCRAPING_HUMAN_LIKE_TIMING_ENABLED`**: Enable human-like timing patterns (default: true)
- **`SCRAPING_CAPTCHA_DETECTION_ENABLED`**: Enable CAPTCHA detection (default: true)
- **`SCRAPING_SESSION_MANAGEMENT_ENABLED`**: Enable session/cookie management (default: true)
- **`SCRAPING_USE_PROXIES`**: Enable proxy rotation (default: false)
- **`SCRAPING_PROXY_LIST`**: Comma-separated list of proxy URLs (required if proxies enabled)
- **`SCRAPING_SESSION_MAX_AGE`**: Maximum age for sessions (default: 1h, e.g., "2h", "30m")
- **`SCRAPING_INCLUDE_SEC_FETCH_HEADERS`**: Include Sec-Fetch-* headers (default: true)
- **`SCRAPING_INCLUDE_SEC_CH_UA_HEADERS`**: Include Sec-Ch-Ua-* headers (default: true)

#### CAPTCHA Solving (Advanced - Requires Legal Review)

- **`SCRAPING_CAPTCHA_SOLVING_ENABLED`**: Enable CAPTCHA solving (default: false, **requires legal review**)

**Warning**: CAPTCHA solving may violate terms of service and should only be enabled after legal review.

### Example Configuration

```bash
# Conservative scraping (5-second delay)
export SCRAPING_RATE_LIMIT_DELAY=5

# Custom contact URL
export SCRAPING_USER_AGENT_CONTACT_URL=https://example.com/bot-info

# Disable robots.txt checking (not recommended)
export SCRAPING_RESPECT_ROBOTS=false

# Enable all bot evasion techniques
export SCRAPING_HEADER_RANDOMIZATION_ENABLED=true
export SCRAPING_HUMAN_LIKE_TIMING_ENABLED=true
export SCRAPING_CAPTCHA_DETECTION_ENABLED=true
export SCRAPING_SESSION_MANAGEMENT_ENABLED=true

# Optional: Enable proxy rotation
export SCRAPING_USE_PROXIES=true
export SCRAPING_PROXY_LIST="http://proxy1.example.com:8080,http://proxy2.example.com:8080"

# Session configuration
export SCRAPING_SESSION_MAX_AGE=2h
```

## Monitoring and Compliance

### Logging

We maintain comprehensive logs of:
- **Scraping Activity**: Domains accessed and status codes received
- **Rate Limiting Events**: When and why we were rate-limited
- **Access Denials**: When and why access was forbidden
- **Robots.txt Compliance**: When robots.txt rules were respected

### Regular Review

We regularly review:
- **Compliance**: Ensure all scraping follows this policy
- **Error Rates**: Monitor 429/403/503 rates to adjust behavior
- **Opt-Out Requests**: Process and implement exclusion requests
- **Policy Updates**: Update this policy as needed

## Updates to This Policy

This policy may be updated periodically. Significant changes will be:
- **Documented**: Changes will be clearly marked in version history
- **Communicated**: Major changes will be announced
- **Effective Date**: Changes will have a clear effective date

**Last Updated**: 2025-01-27  
**Version**: 2.0

### Version History

- **v2.0** (2025-01-27): Added bot evasion techniques (header randomization, human-like timing, CAPTCHA detection, session management, optional proxy support) while maintaining identifiable User-Agent and legal compliance
- **v1.0** (2025-01-27): Initial policy with robots.txt compliance, identifiable User-Agent, and conservative rate limiting

## Questions or Concerns

If you have questions or concerns about our scraping practices:

- **Email**: [support@kyb-platform.com](mailto:support@kyb-platform.com)
- **Bot Info**: [https://kyb-platform.com/bot-info](https://kyb-platform.com/bot-info)
- **Response Time**: We aim to respond within 48 hours

---

**Note**: This policy is designed to ensure legal compliance and respectful web scraping. We are committed to being good internet citizens and respecting website owners' rights and preferences.

