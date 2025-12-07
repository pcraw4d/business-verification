import express from 'express';
import { chromium } from 'playwright';
import PQueue from 'p-queue';

const app = express();
// FIX: Add request body size limit to prevent DoS attacks
app.use(express.json({ limit: '10kb' }));

// Configuration from environment variables
// FIX: Increased default browser pool size from 3 to 6 for better concurrency
let BROWSER_POOL_SIZE = parseInt(process.env.BROWSER_POOL_SIZE || '6', 10);
let MAX_CONCURRENT_REQUESTS = parseInt(process.env.MAX_CONCURRENT_REQUESTS || '8', 10);
let QUEUE_TIMEOUT_MS = parseInt(process.env.QUEUE_TIMEOUT_MS || '25000', 10);
let SCRAPE_TIMEOUT_MS = parseInt(process.env.SCRAPE_TIMEOUT_MS || '15000', 10);
const BROWSER_RECOVERY_ENABLED = process.env.BROWSER_RECOVERY_ENABLED !== 'false';

// Validate configuration
if (BROWSER_POOL_SIZE < 1 || BROWSER_POOL_SIZE > 10) {
    console.warn(`Invalid BROWSER_POOL_SIZE: ${BROWSER_POOL_SIZE}, using default: 6`);
    BROWSER_POOL_SIZE = 6;
}
if (MAX_CONCURRENT_REQUESTS < 1 || MAX_CONCURRENT_REQUESTS > 20) {
    console.warn(`Invalid MAX_CONCURRENT_REQUESTS: ${MAX_CONCURRENT_REQUESTS}, using default: 8`);
    MAX_CONCURRENT_REQUESTS = 8;
}
// FIX: Validate timeout values for reasonable ranges
if (QUEUE_TIMEOUT_MS < 1000 || QUEUE_TIMEOUT_MS > 120000) {
    console.warn(`Invalid QUEUE_TIMEOUT_MS: ${QUEUE_TIMEOUT_MS}, using default: 25000`);
    QUEUE_TIMEOUT_MS = 25000;
}
if (SCRAPE_TIMEOUT_MS < 1000 || SCRAPE_TIMEOUT_MS > 60000) {
    console.warn(`Invalid SCRAPE_TIMEOUT_MS: ${SCRAPE_TIMEOUT_MS}, using default: 15000`);
    SCRAPE_TIMEOUT_MS = 15000;
}

// Log configuration on startup
console.log(JSON.stringify({
    level: 'info',
    message: 'Playwright scraper starting',
    config: {
        browserPoolSize: BROWSER_POOL_SIZE,
        maxConcurrentRequests: MAX_CONCURRENT_REQUESTS,
        queueTimeoutMs: QUEUE_TIMEOUT_MS,
        scrapeTimeoutMs: SCRAPE_TIMEOUT_MS,
        browserRecoveryEnabled: BROWSER_RECOVERY_ENABLED
    }
}));

// Browser Pool Implementation
class BrowserPool {
    constructor(size) {
        this.pool = [];
        this.size = size;
        this.inUse = new Set();
        this.recoveryEnabled = BROWSER_RECOVERY_ENABLED;
        // FIX: Add mutex for atomic browser acquisition
        this.acquireLock = false;
    }

    async init() {
        console.log(JSON.stringify({
            level: 'info',
            message: 'Initializing browser pool',
            poolSize: this.size
        }));

        for (let i = 0; i < this.size; i++) {
            try {
                const browser = await this.createBrowser();
                this.pool.push({
                    browser,
                    inUse: false,
                    lastUsed: Date.now(),
                    id: i
                });
                console.log(JSON.stringify({
                    level: 'info',
                    message: 'Browser added to pool',
                    browserId: i,
                    poolSize: this.pool.length
                }));
            } catch (error) {
                console.error(JSON.stringify({
                    level: 'error',
                    message: 'Failed to create browser for pool',
                    browserId: i,
                    error: error.message
                }));
                // Continue with fewer browsers if some fail to initialize
            }
        }

        if (this.pool.length === 0) {
            throw new Error('Failed to initialize browser pool: no browsers available');
        }

        console.log(JSON.stringify({
            level: 'info',
            message: 'Browser pool initialized',
            poolSize: this.pool.length,
            targetSize: this.size
        }));
    }

    async createBrowser() {
        try {
            // FIX: Remove --single-process flag to improve browser stability
            // Single-process mode causes browser crashes: "Target page, context or browser has been closed"
            // Use multi-process mode for better stability (slightly more memory but much more stable)
            const launchArgs = [
                '--no-sandbox',
                '--disable-setuid-sandbox',
                '--disable-dev-shm-usage',
                '--disable-gpu'
            ];
            
            // Only use single-process if explicitly enabled via environment variable (for resource-constrained environments)
            // Default: multi-process for stability
            if (process.env.USE_SINGLE_PROCESS === 'true') {
                launchArgs.push('--single-process');
                console.log(JSON.stringify({
                    level: 'warn',
                    message: 'Using single-process mode (may cause instability)',
                    note: 'Set USE_SINGLE_PROCESS=false for better stability'
                }));
            }
            
            return await chromium.launch({
                headless: true,
                args: launchArgs
            });
        } catch (error) {
            // FIX: Add specific error handling for better debugging
            if (error.message.includes('EAGAIN') || error.message.includes('Resource')) {
                throw new Error(`Resource exhaustion: ${error.message}`);
            }
            throw error;
        }
    }

    // FIX: Increase default timeout to match queue timeout (25s) to prevent premature failures
    async acquire(timeout = 25000) {
        const startTime = Date.now();

        while (Date.now() - startTime < timeout) {
            // FIX: Use mutex to prevent race condition in browser acquisition
            if (this.acquireLock) {
                await new Promise(resolve => setTimeout(resolve, 50));
                continue;
            }

            this.acquireLock = true;
            try {
                // Find available browser
                const available = this.pool.find(item => !item.inUse && !this.inUse.has(item.browser));

                if (available) {
                    // Check if browser is still alive
                    if (available.browser.isConnected()) {
                        // FIX: Atomic operation - mark as in use immediately
                        available.inUse = true;
                        available.lastUsed = Date.now();
                        this.inUse.add(available.browser);
                        return available;
                    } else {
                        // Browser is dead, try to recover
                        console.warn(JSON.stringify({
                            level: 'warn',
                            message: 'Browser is not connected, attempting recovery',
                            browserId: available.id
                        }));
                        if (this.recoveryEnabled) {
                            await this.recover(available);
                        }
                    }
                }
            } finally {
                this.acquireLock = false;
            }

            // Wait a bit before retrying
            await new Promise(resolve => setTimeout(resolve, 100));
        }

        throw new Error('Timeout waiting for available browser');
    }

    release(poolItem) {
        if (poolItem && poolItem.browser) {
            poolItem.inUse = false;
            poolItem.lastUsed = Date.now();
            const wasInUse = this.inUse.has(poolItem.browser);
            this.inUse.delete(poolItem.browser);
            
            // FIX: Log release for debugging browser pool issues
            if (wasInUse) {
                console.log(JSON.stringify({
                    level: 'info',
                    message: 'Browser released from pool',
                    browserId: poolItem.id,
                    poolSize: this.pool.length,
                    inUseCount: this.inUse.size
                }));
            } else {
                console.warn(JSON.stringify({
                    level: 'warn',
                    message: 'Browser released but was not marked as in use',
                    browserId: poolItem.id,
                    poolSize: this.pool.length,
                    inUseCount: this.inUse.size
                }));
            }
        }
    }

    async recover(poolItem) {
        if (!this.recoveryEnabled) {
            return;
        }

        console.log(JSON.stringify({
            level: 'info',
            message: 'Recovering dead browser',
            browserId: poolItem.id
        }));

        try {
            // FIX: Store old browser reference before replacing
            const oldBrowser = poolItem.browser;
            
            // Try to close the dead browser
            try {
                await oldBrowser.close();
            } catch (e) {
                // Ignore errors when closing dead browser
            }

            // FIX: Remove old browser from inUse Set before creating new one
            this.inUse.delete(oldBrowser);

            // Create new browser
            const newBrowser = await this.createBrowser();
            poolItem.browser = newBrowser;
            poolItem.inUse = false;
            poolItem.lastUsed = Date.now();
            // Note: New browser will be added to inUse when acquired

            console.log(JSON.stringify({
                level: 'info',
                message: 'Browser recovered successfully',
                browserId: poolItem.id
            }));
        } catch (error) {
            console.error(JSON.stringify({
                level: 'error',
                message: 'Failed to recover browser',
                browserId: poolItem.id,
                error: error.message
            }));
            
            // FIX: Remove old browser from inUse Set before removing from pool
            const oldBrowser = poolItem.browser;
            this.inUse.delete(oldBrowser);
            
            // Remove dead browser from pool
            const index = this.pool.indexOf(poolItem);
            if (index > -1) {
                this.pool.splice(index, 1);
            }
            
            // FIX: Try to create replacement browser to maintain pool size
            try {
                const replacement = await this.createBrowser();
                this.pool.push({
                    browser: replacement,
                    inUse: false,
                    lastUsed: Date.now(),
                    id: poolItem.id // Reuse the same ID
                });
                console.log(JSON.stringify({
                    level: 'info',
                    message: 'Replacement browser created to maintain pool size',
                    browserId: poolItem.id,
                    poolSize: this.pool.length
                }));
            } catch (replacementError) {
                console.error(JSON.stringify({
                    level: 'error',
                    message: 'Failed to create replacement browser',
                    browserId: poolItem.id,
                    error: replacementError.message,
                    poolSize: this.pool.length
                }));
                // Pool will be smaller, but service can continue
            }
        }
    }

    async cleanup() {
        console.log(JSON.stringify({
            level: 'info',
            message: 'Cleaning up browser pool',
            poolSize: this.pool.length
        }));

        const closePromises = this.pool.map(async (item) => {
            try {
                if (item.browser && item.browser.isConnected()) {
                    await item.browser.close();
                }
            } catch (error) {
                console.error(JSON.stringify({
                    level: 'error',
                    message: 'Error closing browser during cleanup',
                    browserId: item.id,
                    error: error.message
                }));
            }
        });

        await Promise.all(closePromises);
        this.pool = [];
        this.inUse.clear();

        console.log(JSON.stringify({
            level: 'info',
            message: 'Browser pool cleaned up'
        }));
    }

    getStats() {
        // FIX: Snapshot pool state to avoid race conditions during calculation
        const poolSnapshot = [...this.pool];
        const inUseSnapshot = new Set(this.inUse);
        
        const available = poolSnapshot.filter(item => 
            !item.inUse && 
            item.browser.isConnected() && 
            !inUseSnapshot.has(item.browser)
        ).length;
        const inUse = poolSnapshot.filter(item => item.inUse).length;
        const dead = poolSnapshot.filter(item => !item.browser.isConnected()).length;

        return {
            total: poolSnapshot.length,
            available,
            inUse,
            dead,
            utilization: poolSnapshot.length > 0 ? (inUse / poolSnapshot.length) * 100 : 0
        };
    }
}

// Initialize browser pool
const browserPool = new BrowserPool(BROWSER_POOL_SIZE);

// Initialize pool on startup
browserPool.init().catch(error => {
    console.error(JSON.stringify({
        level: 'error',
        message: 'Failed to initialize browser pool',
        error: error.message
    }));
    process.exit(1);
});

// Request queue with concurrency limiting
const queue = new PQueue({
    concurrency: MAX_CONCURRENT_REQUESTS,
    timeout: QUEUE_TIMEOUT_MS,
    throwOnTimeout: true
});

// Queue metrics
// FIX: Use circular buffer for queue wait times to improve memory efficiency
const MAX_METRICS = 100;
let queueWaitTimesIndex = 0;
const queueWaitTimes = new Array(MAX_METRICS).fill(0);

const queueMetrics = {
    totalRequests: 0,
    completedRequests: 0,
    failedRequests: 0,
    timeoutRequests: 0,
    rejectedRequests: 0, // FIX: Track rejected requests (queue full)
    get queueWaitTimes() {
        // Return only non-zero values for metrics
        return queueWaitTimes.filter(t => t > 0);
    }
};

// Health check endpoint
app.get('/health', (req, res) => {
    const poolStats = browserPool.getStats();
    const queueSize = queue.size;
    const queuePending = queue.pending;

    // Determine service status
    let status = 'ok';
    if (poolStats.dead === poolStats.total && poolStats.total > 0) {
        status = 'unhealthy';
    } else if (queueSize > MAX_CONCURRENT_REQUESTS * 0.8) {
        status = 'degraded';
    }

    // Get memory usage if available
    const memoryUsage = process.memoryUsage();
    const memoryMB = Math.round(memoryUsage.heapUsed / 1024 / 1024);

    res.json({
        status,
        service: 'playwright-scraper',
        browserPool: {
            total: poolStats.total,
            available: poolStats.available,
            inUse: poolStats.inUse,
            dead: poolStats.dead,
            utilization: Math.round(poolStats.utilization * 100) / 100
        },
        queue: {
            size: queueSize,
            pending: queuePending,
            maxConcurrent: MAX_CONCURRENT_REQUESTS
        },
        metrics: {
            totalRequests: queueMetrics.totalRequests,
            completedRequests: queueMetrics.completedRequests,
            failedRequests: queueMetrics.failedRequests,
            timeoutRequests: queueMetrics.timeoutRequests,
            rejectedRequests: queueMetrics.rejectedRequests || 0,
            successRate: queueMetrics.totalRequests > 0
                ? Math.round((queueMetrics.completedRequests / queueMetrics.totalRequests) * 10000) / 100
                : 100
        },
        memory: {
            heapUsedMB: memoryMB,
            heapTotalMB: Math.round(memoryUsage.heapTotal / 1024 / 1024),
            rssMB: Math.round(memoryUsage.rss / 1024 / 1024)
        }
    });
});

// FIX: URL validation function
function isValidUrl(url) {
    if (!url || typeof url !== 'string') {
        return false;
    }
    
    // Limit URL length to prevent DoS
    if (url.length > 2048) {
        return false;
    }
    
    try {
        const parsed = new URL(url);
        // Only allow http and https protocols
        if (!['http:', 'https:'].includes(parsed.protocol)) {
            return false;
        }
        // Basic validation - URL should have a hostname
        if (!parsed.hostname || parsed.hostname.length === 0) {
            return false;
        }
        return true;
    } catch {
        return false;
    }
}

// Request counter for unique request IDs
let requestCounter = 0;

// Main scraping endpoint
app.post('/scrape', async (req, res) => {
    const { url } = req.body;
    // FIX: Replace deprecated substr() with slice() and add counter to prevent collisions
    const requestId = `req_${Date.now()}_${++requestCounter}_${Math.random().toString(36).slice(2, 11)}`;
    const requestStartTime = Date.now();

    // FIX: Add comprehensive URL validation
    if (!url) {
        return res.status(400).json({
            error: 'URL is required',
            success: false,
            requestId
        });
    }
    
    if (!isValidUrl(url)) {
        return res.status(400).json({
            error: 'Valid HTTP/HTTPS URL is required (max 2048 characters)',
            success: false,
            requestId
        });
    }

    queueMetrics.totalRequests++;

    // FIX: Fail fast if queue is too full (prevent unbounded growth)
    // Reject requests immediately if queue size exceeds reasonable limit
    const MAX_QUEUE_SIZE = MAX_CONCURRENT_REQUESTS * 10; // Allow 10x concurrent requests in queue
    if (queue.size >= MAX_QUEUE_SIZE) {
        console.warn(JSON.stringify({
            level: 'warn',
            message: 'Queue is full, rejecting request',
            requestId,
            url,
            queueSize: queue.size,
            maxQueueSize: MAX_QUEUE_SIZE
        }));

        queueMetrics.failedRequests++;
        queueMetrics.rejectedRequests = (queueMetrics.rejectedRequests || 0) + 1;

        return res.status(503).json({
            error: 'Service overloaded: queue is full',
            success: false,
            requestId,
            metrics: {
                queueSize: queue.size,
                maxQueueSize: MAX_QUEUE_SIZE
            }
        });
    }

    console.log(JSON.stringify({
        level: 'info',
        message: 'Scrape request received',
        requestId,
        url,
        queueSize: queue.size,
        queuePending: queue.pending
    }));

    let poolItem = null; // FIX: Declare outside try-catch to ensure release in all paths
    let context = null;
    let page = null;
    
    try {
        const queueWaitStart = Date.now();
        
        await queue.add(async () => {
            const queueWaitTime = Date.now() - queueWaitStart;
            // FIX: Use circular buffer instead of array with shift() (O(n) -> O(1))
            queueWaitTimes[queueWaitTimesIndex] = queueWaitTime;
            queueWaitTimesIndex = (queueWaitTimesIndex + 1) % MAX_METRICS;

            console.log(JSON.stringify({
                level: 'info',
                message: 'Request dequeued, starting scrape',
                requestId,
                url,
                queueWaitTimeMs: queueWaitTime
            }));

            const scrapeStartTime = Date.now();

            try {
                // FIX: Increase browser acquisition timeout to match queue timeout (25s)
                // This prevents premature timeout when browsers are busy
                const BROWSER_ACQUISITION_TIMEOUT = parseInt(process.env.BROWSER_ACQUISITION_TIMEOUT_MS || '25000');
                poolItem = await browserPool.acquire(BROWSER_ACQUISITION_TIMEOUT);

                if (!poolItem || !poolItem.browser) {
                    throw new Error('Failed to acquire browser from pool');
                }

                // Create new context for this request (prevents state pollution)
                context = await poolItem.browser.newContext({
                    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
                    viewport: { width: 1920, height: 1080 }
                });

                // Create new page for this request
                page = await context.newPage();

                // Set timeout
                page.setDefaultTimeout(SCRAPE_TIMEOUT_MS);

                // Navigate
                await page.goto(url, {
                    waitUntil: 'networkidle',
                    timeout: SCRAPE_TIMEOUT_MS
                });

                // Wait a bit for any dynamic content
                await page.waitForTimeout(1000);

                // Get full HTML
                const html = await page.content();

                const scrapeDuration = Date.now() - scrapeStartTime;
                const totalDuration = Date.now() - requestStartTime;

                console.log(JSON.stringify({
                    level: 'info',
                    message: 'Scrape completed successfully',
                    requestId,
                    url,
                    htmlLength: html.length,
                    scrapeDurationMs: scrapeDuration,
                    totalDurationMs: totalDuration,
                    queueWaitTimeMs: queueWaitTime
                }));

                queueMetrics.completedRequests++;

                res.json({
                    html: html,
                    success: true,
                    requestId,
                    metrics: {
                        scrapeDurationMs: scrapeDuration,
                        totalDurationMs: totalDuration,
                        queueWaitTimeMs: queueWaitTime
                    }
                });

            } catch (error) {
                const scrapeDuration = Date.now() - scrapeStartTime;
                const totalDuration = Date.now() - requestStartTime;

                console.error(JSON.stringify({
                    level: 'error',
                    message: 'Scrape failed',
                    requestId,
                    url,
                    error: error.message,
                    errorStack: error.stack,
                    scrapeDurationMs: scrapeDuration,
                    totalDurationMs: totalDuration,
                    queueWaitTimeMs: queueWaitTime
                }));

                queueMetrics.failedRequests++;

                // Determine appropriate HTTP status code
                let statusCode = 500;
                if (error.message.includes('timeout') || error.message.includes('Timeout')) {
                    statusCode = 408;
                    queueMetrics.timeoutRequests++;
                } else if (error.message.includes('browser') || error.message.includes('Browser')) {
                    statusCode = 503;
                    // Try to recover browser if it's dead
                    if (poolItem && BROWSER_RECOVERY_ENABLED) {
                        await browserPool.recover(poolItem);
                    }
                }

                res.status(statusCode).json({
                    error: error.message,
                    success: false,
                    requestId,
                    metrics: {
                        scrapeDurationMs: scrapeDuration,
                        totalDurationMs: totalDuration,
                        queueWaitTimeMs: queueWaitTime
                    }
                });
            } finally {
                // Always cleanup: close page and context
                if (page) {
                    try {
                        await page.close();
                    } catch (e) {
                        // Ignore errors when closing page
                    }
                }

                if (context) {
                    try {
                        await context.close();
                    } catch (e) {
                        // Ignore errors when closing context
                    }
                }

                // Release browser back to pool (inside queue.add finally block)
                if (poolItem) {
                    browserPool.release(poolItem);
                    console.log(JSON.stringify({
                        level: 'info',
                        message: 'Browser released to pool',
                        requestId,
                        browserId: poolItem.id
                    }));
                    poolItem = null; // Clear reference after release
                }
            }
        }, { throwOnTimeout: true });

    } catch (error) {
        const totalDuration = Date.now() - requestStartTime;

        // FIX: Ensure browser is released even if queue.add() throws (e.g., queue timeout)
        // This is critical to prevent browser pool exhaustion
        if (poolItem) {
            try {
                browserPool.release(poolItem);
                console.log(JSON.stringify({
                    level: 'info',
                    message: 'Browser released to pool after queue error',
                    requestId,
                    browserId: poolItem.id,
                    error: error.message
                }));
                poolItem = null;
            } catch (releaseError) {
                console.error(JSON.stringify({
                    level: 'error',
                    message: 'Failed to release browser after queue error',
                    requestId,
                    browserId: poolItem ? poolItem.id : 'unknown',
                    error: releaseError.message
                }));
            }
        }

        // Handle queue timeout
        if (error.message.includes('timeout') || error.name === 'TimeoutError') {
            console.warn(JSON.stringify({
                level: 'warn',
                message: 'Request timed out in queue',
                requestId,
                url,
                totalDurationMs: totalDuration
            }));

            queueMetrics.timeoutRequests++;
            queueMetrics.failedRequests++;

            return res.status(429).json({
                error: 'Request timed out in queue',
                success: false,
                requestId,
                metrics: {
                    totalDurationMs: totalDuration
                }
            });
        }

        // Handle other queue errors
        console.error(JSON.stringify({
            level: 'error',
            message: 'Queue error',
            requestId,
            url,
            error: error.message,
            errorStack: error.stack
        }));

        queueMetrics.failedRequests++;

        res.status(503).json({
            error: 'Service temporarily unavailable',
            success: false,
            requestId
        });
    }
});

// Graceful shutdown
// FIX: Add timeout to shutdown to prevent hanging in containerized environments
const SHUTDOWN_TIMEOUT_MS = 30000; // 30 seconds

async function gracefulShutdown(signal) {
    console.log(JSON.stringify({
        level: 'info',
        message: `${signal} received, shutting down gracefully`
    }));

    // Stop accepting new requests
    // Wait for queue to drain with timeout
    const shutdownPromise = queue.onIdle();
    const timeoutPromise = new Promise(resolve => setTimeout(() => {
        console.warn(JSON.stringify({
            level: 'warn',
            message: 'Shutdown timeout exceeded, forcing exit'
        }));
        resolve();
    }, SHUTDOWN_TIMEOUT_MS));
    
    await Promise.race([shutdownPromise, timeoutPromise]).catch(() => {
        console.warn(JSON.stringify({
            level: 'warn',
            message: 'Queue did not drain within timeout'
        }));
    });

    // Cleanup browser pool
    await browserPool.cleanup();

    process.exit(0);
}

process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
process.on('SIGINT', () => gracefulShutdown('SIGINT'));

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(JSON.stringify({
        level: 'info',
        message: 'Playwright scraper listening',
        port: PORT,
        config: {
            browserPoolSize: BROWSER_POOL_SIZE,
            maxConcurrentRequests: MAX_CONCURRENT_REQUESTS,
            queueTimeoutMs: QUEUE_TIMEOUT_MS,
            scrapeTimeoutMs: SCRAPE_TIMEOUT_MS
        }
    }));
});
