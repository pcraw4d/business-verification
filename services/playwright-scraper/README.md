# Playwright Scraper Service

A Node.js service for scraping JavaScript-heavy websites using Playwright.

## Endpoints

### GET /health
Health check endpoint.

### POST /scrape
Scrapes a website and returns the full HTML content.

**Request Body:**
```json
{
  "url": "https://example.com"
}
```

**Response:**
```json
{
  "html": "<html>...</html>",
  "success": true
}
```

## Deployment

This service is designed to be deployed on Railway. The Dockerfile uses the official Playwright base image.

## Environment Variables

- `PORT`: Server port (default: 3000)

