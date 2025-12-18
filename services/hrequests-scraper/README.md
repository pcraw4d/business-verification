# hrequests-scraper Service

A lightweight Python microservice for fast website scraping using the `hrequests` library.

## Features

- Fast HTTP scraping with browser-like behavior
- Structured content extraction (title, meta description, headings, navigation, etc.)
- Quality scoring and word count calculation
- Timeout handling (5s default)
- Content size limits (10MB default)

## API Endpoints

### POST /scrape

Scrape a website and return structured content.

**Request:**
```json
{
  "url": "https://example.com"
}
```

**Response:**
```json
{
  "success": true,
  "content": {
    "raw_html": "...",
    "plain_text": "...",
    "title": "Example",
    "meta_description": "...",
    "headings": ["Heading 1", "Heading 2"],
    "navigation": ["Home", "About"],
    "about_text": "...",
    "products": ["Product 1"],
    "contact": "...",
    "main_content": "...",
    "word_count": 500,
    "quality_score": 0.85,
    "domain": "example.com",
    "scraped_at": 1234567890.0
  },
  "method": "hrequests",
  "latency_ms": 1234
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "service": "hrequests-scraper",
  "timestamp": 1234567890.0
}
```

## Environment Variables

- `PORT`: Server port (default: 8080)
- `SCRAPE_TIMEOUT`: Scraping timeout in seconds (default: 5)
- `MAX_CONTENT_SIZE`: Maximum content size in bytes (default: 10485760 = 10MB)

## Deployment

### Docker

```bash
docker build -t hrequests-scraper .
docker run -p 8080:8080 hrequests-scraper
```

### Local Development

```bash
pip install -r requirements.txt
python app.py
```

## Integration

This service is called by the Go `classification-service` via the `HrequestsClient`.
The service URL is configured via the `HREQUESTS_SERVICE_URL` environment variable.



