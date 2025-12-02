const express = require('express');
const { chromium } = require('playwright');

const app = express();
app.use(express.json());

// Health check
app.get('/health', (req, res) => {
    res.json({ status: 'ok', service: 'playwright-scraper' });
});

// Main scraping endpoint
app.post('/scrape', async (req, res) => {
    const { url } = req.body;
    
    if (!url) {
        return res.status(400).json({ error: 'URL is required' });
    }
    
    console.log(`Scraping: ${url}`);
    
    let browser;
    try {
        // Launch browser
        browser = await chromium.launch({
            headless: true,
            args: ['--no-sandbox', '--disable-setuid-sandbox']
        });
        
        const context = await browser.newContext({
            userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
            viewport: { width: 1920, height: 1080 }
        });
        
        const page = await context.newPage();
        
        // Set timeout
        page.setDefaultTimeout(15000);
        
        // Navigate
        await page.goto(url, { 
            waitUntil: 'networkidle',
            timeout: 15000 
        });
        
        // Wait a bit for any dynamic content
        await page.waitForTimeout(1000);
        
        // Get full HTML
        const html = await page.content();
        
        // Close
        await browser.close();
        
        console.log(`Success: ${url} (${html.length} bytes)`);
        
        res.json({
            html: html,
            success: true
        });
        
    } catch (error) {
        console.error(`Error scraping ${url}:`, error.message);
        
        if (browser) {
            await browser.close();
        }
        
        res.status(500).json({
            error: error.message,
            success: false
        });
    }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Playwright scraper listening on port ${PORT}`);
});

