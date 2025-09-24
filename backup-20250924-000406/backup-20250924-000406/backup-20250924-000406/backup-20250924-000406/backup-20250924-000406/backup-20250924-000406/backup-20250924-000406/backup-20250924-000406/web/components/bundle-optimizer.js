/**
 * Bundle Optimizer
 * Provides code splitting, tree shaking, and bundle optimization for merchant UI components
 * Implements dynamic imports and module federation for better performance
 */

class BundleOptimizer {
    constructor(options = {}) {
        this.options = {
            baseUrl: '/',
            cachePrefix: 'kyb-bundle-',
            cacheVersion: '1.0.0',
            maxCacheSize: 50 * 1024 * 1024, // 50MB
            compressionEnabled: true,
            ...options
        };

        this.modules = new Map();
        this.loadedModules = new Set();
        this.loadingPromises = new Map();
        this.cache = new Map();
        this.bundleStats = {
            totalSize: 0,
            loadedSize: 0,
            moduleCount: 0,
            loadTime: 0
        };

        this.init();
    }

    /**
     * Initialize the bundle optimizer
     */
    init() {
        this.setupCache();
        this.registerCoreModules();
        this.setupPerformanceMonitoring();
    }

    /**
     * Setup cache management
     */
    setupCache() {
        // Check if localStorage is available
        if (typeof Storage !== 'undefined') {
            this.cacheEnabled = true;
            this.loadCacheFromStorage();
        } else {
            this.cacheEnabled = false;
        }
    }

    /**
     * Register core modules
     */
    registerCoreModules() {
        // Core merchant modules
        this.registerModule('merchant-core', {
            path: '/components/merchant-core.js',
            dependencies: [],
            priority: 'high',
            size: 0 // Will be calculated on load
        });

        this.registerModule('merchant-portfolio', {
            path: '/components/merchant-portfolio.js',
            dependencies: ['merchant-core'],
            priority: 'high',
            size: 0
        });

        this.registerModule('merchant-dashboard', {
            path: '/components/merchant-dashboard.js',
            dependencies: ['merchant-core'],
            priority: 'high',
            size: 0
        });

        this.registerModule('merchant-search', {
            path: '/components/merchant-search.js',
            dependencies: ['merchant-core'],
            priority: 'medium',
            size: 0
        });

        this.registerModule('merchant-comparison', {
            path: '/components/merchant-comparison.js',
            dependencies: ['merchant-core', 'merchant-portfolio'],
            priority: 'medium',
            size: 0
        });

        this.registerModule('bulk-operations', {
            path: '/components/bulk-operations.js',
            dependencies: ['merchant-core', 'merchant-portfolio'],
            priority: 'low',
            size: 0
        });

        this.registerModule('lazy-loader', {
            path: '/components/lazy-loader.js',
            dependencies: [],
            priority: 'high',
            size: 0
        });

        this.registerModule('virtual-scroller', {
            path: '/components/virtual-scroller.js',
            dependencies: [],
            priority: 'medium',
            size: 0
        });

        // Utility modules
        this.registerModule('utils', {
            path: '/components/utils.js',
            dependencies: [],
            priority: 'high',
            size: 0
        });

        this.registerModule('api-client', {
            path: '/components/api-client.js',
            dependencies: ['utils'],
            priority: 'high',
            size: 0
        });
    }

    /**
     * Register a module
     * @param {string} name - Module name
     * @param {Object} config - Module configuration
     */
    registerModule(name, config) {
        this.modules.set(name, {
            name,
            path: config.path,
            dependencies: config.dependencies || [],
            priority: config.priority || 'medium',
            size: config.size || 0,
            loaded: false,
            loading: false,
            error: null,
            loadTime: 0,
            ...config
        });
    }

    /**
     * Load a module with dependencies
     * @param {string} moduleName - Module name to load
     * @param {Object} options - Loading options
     */
    async loadModule(moduleName, options = {}) {
        const startTime = performance.now();
        
        if (this.loadedModules.has(moduleName)) {
            return this.modules.get(moduleName);
        }

        if (this.loadingPromises.has(moduleName)) {
            return this.loadingPromises.get(moduleName);
        }

        const module = this.modules.get(moduleName);
        if (!module) {
            throw new Error(`Module '${moduleName}' not found`);
        }

        const loadPromise = this.loadModuleWithDependencies(module, options);
        this.loadingPromises.set(moduleName, loadPromise);

        try {
            const result = await loadPromise;
            const loadTime = performance.now() - startTime;
            
            module.loaded = true;
            module.loadTime = loadTime;
            module.loading = false;
            this.loadedModules.add(moduleName);
            this.bundleStats.loadedSize += module.size;
            this.bundleStats.loadTime += loadTime;
            
            this.loadingPromises.delete(moduleName);
            
            // Cache the module
            if (this.cacheEnabled) {
                this.cacheModule(moduleName, result);
            }
            
            // Trigger module loaded event
            this.triggerEvent('moduleLoaded', { module: moduleName, loadTime });
            
            return result;
        } catch (error) {
            module.error = error;
            module.loading = false;
            this.loadingPromises.delete(moduleName);
            
            this.triggerEvent('moduleError', { module: moduleName, error });
            throw error;
        }
    }

    /**
     * Load module with its dependencies
     * @param {Object} module - Module configuration
     * @param {Object} options - Loading options
     */
    async loadModuleWithDependencies(module, options) {
        // Load dependencies first
        const dependencyPromises = module.dependencies.map(dep => 
            this.loadModule(dep, options)
        );
        
        await Promise.all(dependencyPromises);
        
        // Load the module itself
        return this.loadModuleContent(module, options);
    }

    /**
     * Load module content
     * @param {Object} module - Module configuration
     * @param {Object} options - Loading options
     */
    async loadModuleContent(module, options) {
        // Check cache first
        if (this.cacheEnabled && this.cache.has(module.name)) {
            const cached = this.cache.get(module.name);
            if (this.isCacheValid(cached)) {
                return cached.content;
            }
        }

        // Load from server
        const url = this.options.baseUrl + module.path;
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Accept': 'application/javascript',
                'Cache-Control': 'max-age=3600'
            }
        });

        if (!response.ok) {
            throw new Error(`Failed to load module '${module.name}': ${response.statusText}`);
        }

        const content = await response.text();
        const size = new Blob([content]).size;
        
        // Update module size
        module.size = size;
        this.bundleStats.totalSize += size;

        // Execute the module content
        const moduleExports = this.executeModuleContent(content, module.name);
        
        return moduleExports;
    }

    /**
     * Execute module content
     * @param {string} content - Module content
     * @param {string} moduleName - Module name
     */
    executeModuleContent(content, moduleName) {
        try {
            // Create a module context
            const moduleContext = {
                exports: {},
                module: { exports: {} },
                require: this.createRequireFunction(moduleName),
                __filename: moduleName,
                __dirname: '/'
            };

            // Wrap the content in a function
            const wrappedContent = `
                (function(exports, module, require, __filename, __dirname) {
                    ${content}
                    return module.exports;
                })
            `;

            const moduleFunction = eval(wrappedContent);
            return moduleFunction(
                moduleContext.exports,
                moduleContext.module,
                moduleContext.require,
                moduleContext.__filename,
                moduleContext.__dirname
            );
        } catch (error) {
            throw new Error(`Failed to execute module '${moduleName}': ${error.message}`);
        }
    }

    /**
     * Create require function for module
     * @param {string} moduleName - Module name
     */
    createRequireFunction(moduleName) {
        return (name) => {
            if (this.loadedModules.has(name)) {
                return this.modules.get(name).exports;
            }
            throw new Error(`Module '${name}' not loaded. Use loadModule() first.`);
        };
    }

    /**
     * Preload modules by priority
     * @param {Array} priorities - Priority levels to preload
     */
    async preloadModules(priorities = ['high']) {
        const modulesToLoad = Array.from(this.modules.values())
            .filter(module => priorities.includes(module.priority) && !module.loaded)
            .sort((a, b) => this.getPriorityWeight(b.priority) - this.getPriorityWeight(a.priority));

        const loadPromises = modulesToLoad.map(module => 
            this.loadModule(module.name).catch(error => {
                console.warn(`Failed to preload module '${module.name}':`, error);
            })
        );

        await Promise.all(loadPromises);
    }

    /**
     * Get priority weight
     * @param {string} priority - Priority level
     */
    getPriorityWeight(priority) {
        const weights = { high: 3, medium: 2, low: 1 };
        return weights[priority] || 1;
    }

    /**
     * Load modules for a specific route
     * @param {string} route - Route name
     */
    async loadRouteModules(route) {
        const routeModules = this.getRouteModules(route);
        const loadPromises = routeModules.map(moduleName => 
            this.loadModule(moduleName)
        );

        await Promise.all(loadPromises);
    }

    /**
     * Get modules required for a route
     * @param {string} route - Route name
     */
    getRouteModules(route) {
        const routeModuleMap = {
            'merchant-portfolio': ['merchant-core', 'merchant-portfolio', 'merchant-search', 'lazy-loader', 'virtual-scroller'],
            'merchant-dashboard': ['merchant-core', 'merchant-dashboard', 'lazy-loader'],
            'merchant-comparison': ['merchant-core', 'merchant-comparison', 'merchant-portfolio'],
            'bulk-operations': ['merchant-core', 'bulk-operations', 'merchant-portfolio'],
            'default': ['merchant-core', 'utils', 'api-client']
        };

        return routeModuleMap[route] || routeModuleMap['default'];
    }

    /**
     * Cache module content
     * @param {string} moduleName - Module name
     * @param {*} content - Module content
     */
    cacheModule(moduleName, content) {
        if (!this.cacheEnabled) return;

        const cacheKey = `${this.options.cachePrefix}${moduleName}`;
        const cacheData = {
            content,
            timestamp: Date.now(),
            version: this.options.cacheVersion
        };

        this.cache.set(moduleName, cacheData);

        try {
            localStorage.setItem(cacheKey, JSON.stringify(cacheData));
        } catch (error) {
            console.warn('Failed to cache module to localStorage:', error);
        }
    }

    /**
     * Check if cache is valid
     * @param {Object} cached - Cached data
     */
    isCacheValid(cached) {
        if (!cached) return false;
        
        const maxAge = 24 * 60 * 60 * 1000; // 24 hours
        const isNotExpired = Date.now() - cached.timestamp < maxAge;
        const isCorrectVersion = cached.version === this.options.cacheVersion;
        
        return isNotExpired && isCorrectVersion;
    }

    /**
     * Load cache from storage
     */
    loadCacheFromStorage() {
        try {
            for (let i = 0; i < localStorage.length; i++) {
                const key = localStorage.key(i);
                if (key && key.startsWith(this.options.cachePrefix)) {
                    const moduleName = key.replace(this.options.cachePrefix, '');
                    const cacheData = JSON.parse(localStorage.getItem(key));
                    
                    if (this.isCacheValid(cacheData)) {
                        this.cache.set(moduleName, cacheData);
                    } else {
                        localStorage.removeItem(key);
                    }
                }
            }
        } catch (error) {
            console.warn('Failed to load cache from storage:', error);
        }
    }

    /**
     * Clear cache
     */
    clearCache() {
        this.cache.clear();
        
        if (this.cacheEnabled) {
            try {
                for (let i = localStorage.length - 1; i >= 0; i--) {
                    const key = localStorage.key(i);
                    if (key && key.startsWith(this.options.cachePrefix)) {
                        localStorage.removeItem(key);
                    }
                }
            } catch (error) {
                console.warn('Failed to clear cache from storage:', error);
            }
        }
    }

    /**
     * Setup performance monitoring
     */
    setupPerformanceMonitoring() {
        // Monitor bundle loading performance
        this.performanceObserver = new PerformanceObserver((list) => {
            const entries = list.getEntries();
            entries.forEach(entry => {
                if (entry.name.includes('bundle') || entry.name.includes('module')) {
                    this.bundleStats.loadTime += entry.duration;
                }
            });
        });

        try {
            this.performanceObserver.observe({ entryTypes: ['measure', 'navigation'] });
        } catch (error) {
            console.warn('Performance Observer not supported:', error);
        }
    }

    /**
     * Get bundle statistics
     */
    getBundleStats() {
        return {
            ...this.bundleStats,
            moduleCount: this.modules.size,
            loadedModuleCount: this.loadedModules.size,
            cacheSize: this.cache.size,
            cacheEnabled: this.cacheEnabled,
            averageLoadTime: this.bundleStats.loadTime / Math.max(this.loadedModules.size, 1)
        };
    }

    /**
     * Trigger custom event
     * @param {string} eventName - Event name
     * @param {Object} detail - Event detail
     */
    triggerEvent(eventName, detail) {
        const event = new CustomEvent(`bundle:${eventName}`, { detail });
        window.dispatchEvent(event);
    }

    /**
     * Destroy the bundle optimizer
     */
    destroy() {
        if (this.performanceObserver) {
            this.performanceObserver.disconnect();
        }
        
        this.loadingPromises.clear();
        this.cache.clear();
        this.modules.clear();
        this.loadedModules.clear();
    }
}

/**
 * Dynamic Import Manager
 * Manages dynamic imports with fallbacks and error handling
 */
class DynamicImportManager {
    constructor(bundleOptimizer) {
        this.bundleOptimizer = bundleOptimizer;
        this.importCache = new Map();
        this.fallbackModules = new Map();
    }

    /**
     * Dynamic import with fallback
     * @param {string} moduleName - Module name
     * @param {string} fallbackModule - Fallback module name
     */
    async importWithFallback(moduleName, fallbackModule = null) {
        try {
            return await this.bundleOptimizer.loadModule(moduleName);
        } catch (error) {
            console.warn(`Failed to load module '${moduleName}':`, error);
            
            if (fallbackModule) {
                try {
                    return await this.bundleOptimizer.loadModule(fallbackModule);
                } catch (fallbackError) {
                    console.error(`Failed to load fallback module '${fallbackModule}':`, fallbackError);
                    throw fallbackError;
                }
            }
            
            throw error;
        }
    }

    /**
     * Conditional import
     * @param {string} moduleName - Module name
     * @param {Function} condition - Condition function
     */
    async conditionalImport(moduleName, condition) {
        if (condition()) {
            return await this.bundleOptimizer.loadModule(moduleName);
        }
        return null;
    }

    /**
     * Lazy import with timeout
     * @param {string} moduleName - Module name
     * @param {number} timeout - Timeout in milliseconds
     */
    async lazyImportWithTimeout(moduleName, timeout = 5000) {
        const timeoutPromise = new Promise((_, reject) => {
            setTimeout(() => reject(new Error(`Module '${moduleName}' load timeout`)), timeout);
        });

        return Promise.race([
            this.bundleOptimizer.loadModule(moduleName),
            timeoutPromise
        ]);
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { BundleOptimizer, DynamicImportManager };
}
