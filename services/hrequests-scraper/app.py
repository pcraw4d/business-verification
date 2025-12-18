"""
hrequests-scraper service
A lightweight Python service for fast website scraping using hrequests library.
"""

import os
import time
import json
import logging
from typing import Optional, Dict, Any
from flask import Flask, request, jsonify
from bs4 import BeautifulSoup

# Import hrequests library
import hrequests

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Configuration
TIMEOUT = int(os.getenv("SCRAPE_TIMEOUT", "5"))  # 5 seconds default
MAX_CONTENT_SIZE = int(os.getenv("MAX_CONTENT_SIZE", "10485760"))  # 10MB default


def extract_structured_content(html: str, url: str) -> Dict[str, Any]:
    """
    Extract structured content from HTML.
    Returns a dictionary matching the ScrapedContent structure.
    """
    soup = BeautifulSoup(html, 'html.parser')
    
    # Extract title
    title = ""
    if soup.title:
        title = soup.title.string.strip() if soup.title.string else ""
    
    # Extract meta description
    meta_desc = ""
    meta_desc_tag = soup.find('meta', attrs={'name': 'description'})
    if not meta_desc_tag:
        meta_desc_tag = soup.find('meta', attrs={'property': 'og:description'})
    if meta_desc_tag and meta_desc_tag.get('content'):
        meta_desc = meta_desc_tag['content'].strip()
    
    # Extract headings (H1, H2, H3)
    headings = []
    for tag in soup.find_all(['h1', 'h2', 'h3']):
        text = tag.get_text(strip=True)
        if text:
            headings.append(text)
    
    # Extract navigation menu items
    nav_menu = []
    nav_tags = soup.find_all('nav')
    for nav in nav_tags:
        for link in nav.find_all('a', href=True):
            text = link.get_text(strip=True)
            if text and len(text) < 100:  # Reasonable nav item length
                nav_menu.append(text)
    
    # Extract about/company section
    about_text = ""
    about_keywords = ['about', 'company', 'who we are', 'our story']
    for keyword in about_keywords:
        section = soup.find(string=lambda text: text and keyword.lower() in text.lower())
        if section:
            parent = section.find_parent()
            if parent:
                about_text = parent.get_text(strip=True)[:1000]  # Limit length
                break
    
    # Extract product/service list
    products = []
    product_keywords = ['product', 'service', 'solution', 'offering']
    for keyword in product_keywords:
        sections = soup.find_all(string=lambda text: text and keyword.lower() in text.lower())
        for section in sections[:5]:  # Limit to 5 matches
            parent = section.find_parent()
            if parent:
                text = parent.get_text(strip=True)
                if len(text) < 200:  # Reasonable product description length
                    products.append(text)
    
    # Extract contact information
    contact_info = ""
    contact_keywords = ['contact', 'phone', 'email', 'address']
    for keyword in contact_keywords:
        section = soup.find(string=lambda text: text and keyword.lower() in text.lower())
        if section:
            parent = section.find_parent()
            if parent:
                contact_info = parent.get_text(strip=True)[:500]  # Limit length
                break
    
    # Extract main content (paragraphs)
    main_content = ""
    paragraphs = soup.find_all('p')
    for p in paragraphs[:10]:  # Limit to first 10 paragraphs
        text = p.get_text(strip=True)
        if len(text) > 50:  # Only substantial paragraphs
            main_content += text + " "
    main_content = main_content.strip()[:2000]  # Limit total length
    
    # Calculate word count
    plain_text = soup.get_text()
    word_count = len(plain_text.split())
    
    # Calculate quality score
    quality_score = 0.0
    if word_count >= 200:
        quality_score = 0.7
    if word_count >= 500:
        quality_score = 0.8
    if title and meta_desc and word_count >= 200:
        quality_score = min(quality_score + 0.1, 1.0)
    if len(headings) >= 3:
        quality_score = min(quality_score + 0.05, 1.0)
    
    # Extract domain from URL
    from urllib.parse import urlparse
    parsed_url = urlparse(url)
    domain = parsed_url.netloc or url
    
    return {
        "raw_html": html[:MAX_CONTENT_SIZE],  # Limit HTML size
        "plain_text": plain_text[:MAX_CONTENT_SIZE],
        "title": title,
        "meta_description": meta_desc,
        "headings": headings[:20],  # Limit to 20 headings
        "navigation": nav_menu[:30],  # Limit to 30 nav items
        "about_text": about_text,
        "products": products[:20],  # Limit to 20 products
        "contact": contact_info,
        "main_content": main_content,
        "word_count": word_count,
        "language": "en",  # Default to English
        "has_logo": bool(soup.find('img', alt=lambda x: x and 'logo' in x.lower() if x else False)),
        "quality_score": round(quality_score, 2),
        "domain": domain,
        "scraped_at": time.time()
    }


@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "service": "hrequests-scraper",
        "timestamp": time.time()
    }), 200


@app.route('/scrape', methods=['POST'])
def scrape():
    """
    Scrape a website using hrequests.
    
    Request body:
    {
        "url": "https://example.com"
    }
    
    Response:
    {
        "success": true,
        "content": {
            "raw_html": "...",
            "plain_text": "...",
            "title": "...",
            ...
        },
        "method": "hrequests",
        "latency_ms": 1234
    }
    """
    start_time = time.time()
    
    try:
        # Parse request
        data = request.get_json()
        if not data or 'url' not in data:
            return jsonify({
                "success": False,
                "error": "Missing 'url' in request body"
            }), 400
        
        url = data['url']
        logger.info(f"Scraping URL: {url}")
        
        # Scrape with hrequests
        try:
            response = hrequests.get(url, timeout=TIMEOUT)
            
            # Check if response is successful
            if response.status_code >= 400:
                logger.error(f"HTTP error for {url}: status {response.status_code}")
                return jsonify({
                    "success": False,
                    "error": f"HTTP error: status {response.status_code}",
                    "latency_ms": int((time.time() - start_time) * 1000)
                }), 500
            
            # Check content size
            content_length = len(response.content) if hasattr(response, 'content') else len(response.text)
            if content_length > MAX_CONTENT_SIZE:
                logger.warning(f"Content too large: {content_length} bytes, limiting to {MAX_CONTENT_SIZE}")
                html = response.content[:MAX_CONTENT_SIZE].decode('utf-8', errors='ignore') if hasattr(response, 'content') else response.text[:MAX_CONTENT_SIZE]
            else:
                html = response.text if hasattr(response, 'text') else str(response.content.decode('utf-8', errors='ignore'))
            
            # Extract structured content
            structured_content = extract_structured_content(html, url)
            
            latency_ms = int((time.time() - start_time) * 1000)
            
            logger.info(f"Successfully scraped {url}: {structured_content['word_count']} words, "
                       f"quality={structured_content['quality_score']}, latency={latency_ms}ms")
            
            return jsonify({
                "success": True,
                "content": structured_content,
                "method": "hrequests",
                "latency_ms": latency_ms
            }), 200
            
        except Exception as e:
            # Catch all exceptions from hrequests (timeout, connection errors, etc.)
            error_type = type(e).__name__
            logger.error(f"hrequests error for {url} ({error_type}): {str(e)}")
            return jsonify({
                "success": False,
                "error": f"Scraping failed: {str(e)}",
                "latency_ms": int((time.time() - start_time) * 1000)
            }), 500
            
    except Exception as e:
        logger.error(f"Request processing error: {str(e)}")
        return jsonify({
            "success": False,
            "error": f"Request processing failed: {str(e)}",
            "latency_ms": int((time.time() - start_time) * 1000)
        }), 500


if __name__ == '__main__':
    port = int(os.getenv("PORT", "8080"))
    logger.info(f"Starting hrequests-scraper service on port {port}")
    app.run(host='0.0.0.0', port=port, debug=False)



