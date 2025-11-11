/**
 * Comprehensive Form Flow Debugger
 * 
 * This script provides detailed debugging for the entire add-merchant form flow:
 * 1. Form submission
 * 2. Data collection
 * 3. API calls
 * 4. SessionStorage operations
 * 5. Redirect
 * 6. Merchant details page population
 */

class FormFlowDebugger {
    constructor() {
        this.logs = [];
        this.startTime = Date.now();
        this.enabled = true;
        this.init();
    }

    init() {
        if (!this.enabled) return;

        // Create debug panel
        this.createDebugPanel();
        
        // Monitor form submission
        this.monitorFormSubmission();
        
        // Monitor sessionStorage
        this.monitorSessionStorage();
        
        // Monitor redirects
        this.monitorRedirects();
        
        // Monitor DOM changes
        this.monitorDOMChanges();
        
        // Log initial state
        this.log('system', 'FormFlowDebugger initialized', {
            url: window.location.href,
            readyState: document.readyState,
            timestamp: new Date().toISOString()
        });
    }

    createDebugPanel() {
        // Wait for body to be available
        const tryCreatePanel = () => {
            if (!document.body) {
                // Body not ready yet, wait for DOMContentLoaded
                if (document.readyState === 'loading') {
                    document.addEventListener('DOMContentLoaded', tryCreatePanel);
                    return;
                }
                // If still not available, try again after a short delay
                setTimeout(tryCreatePanel, 50);
                return;
            }
            
            // Create floating debug panel
            const panel = document.createElement('div');
            panel.id = 'formFlowDebugPanel';
            panel.style.cssText = `
                position: fixed;
                top: 10px;
                right: 10px;
                width: 400px;
                max-height: 80vh;
                background: rgba(0, 0, 0, 0.9);
                color: #0f0;
                font-family: 'Courier New', monospace;
                font-size: 11px;
                padding: 10px;
                border: 2px solid #0f0;
                border-radius: 5px;
                z-index: 99999;
                overflow-y: auto;
                display: none;
            `;
            
            const header = document.createElement('div');
            header.style.cssText = 'display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; border-bottom: 1px solid #0f0; padding-bottom: 5px;';
            header.innerHTML = `
                <strong>🔍 Form Flow Debugger</strong>
                <button id="toggleDebugPanel" style="background: #0f0; color: #000; border: none; padding: 5px 10px; cursor: pointer; border-radius: 3px;">Toggle</button>
            `;
            
            const content = document.createElement('div');
            content.id = 'debugPanelContent';
            content.style.cssText = 'max-height: 70vh; overflow-y: auto;';
            
            panel.appendChild(header);
            panel.appendChild(content);
            document.body.appendChild(panel);
            
            // Toggle button
            document.getElementById('toggleDebugPanel').addEventListener('click', () => {
                panel.style.display = panel.style.display === 'none' ? 'block' : 'none';
            });
            
            // Keyboard shortcut: Ctrl+Shift+D (PC) or Cmd+Shift+D (Mac)
            document.addEventListener('keydown', (e) => {
                if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'D') {
                    panel.style.display = panel.style.display === 'none' ? 'block' : 'block';
                }
            });
        };
        
        tryCreatePanel();
    }

    log(category, message, data = null) {
        if (!this.enabled) return;
        
        const timestamp = Date.now() - this.startTime;
        const logEntry = {
            timestamp,
            category,
            message,
            data,
            time: new Date().toISOString()
        };
        
        this.logs.push(logEntry);
        console.log(`[DEBUG ${category.toUpperCase()}] ${message}`, data || '');
        
        // Update debug panel
        this.updateDebugPanel(logEntry);
        
        // Keep only last 100 logs
        if (this.logs.length > 100) {
            this.logs.shift();
        }
    }

    updateDebugPanel(logEntry) {
        const content = document.getElementById('debugPanelContent');
        if (!content) return;
        
        const entry = document.createElement('div');
        entry.style.cssText = `
            margin-bottom: 5px;
            padding: 5px;
            border-left: 3px solid ${this.getCategoryColor(logEntry.category)};
            background: rgba(0, 255, 0, 0.1);
        `;
        
        const categoryColors = {
            system: '#0f0',
            form: '#0ff',
            api: '#ff0',
            storage: '#f0f',
            redirect: '#f00',
            dom: '#0ff',
            error: '#f00'
        };
        
        entry.innerHTML = `
            <div style="color: ${categoryColors[logEntry.category] || '#0f0'}">
                <strong>[${logEntry.timestamp}ms] ${logEntry.category.toUpperCase()}:</strong> ${logEntry.message}
            </div>
            ${logEntry.data ? `<pre style="margin: 5px 0; font-size: 10px; overflow-x: auto;">${JSON.stringify(logEntry.data, null, 2)}</pre>` : ''}
        `;
        
        content.appendChild(entry);
        content.scrollTop = content.scrollHeight;
    }

    getCategoryColor(category) {
        const colors = {
            system: '#0f0',
            form: '#0ff',
            api: '#ff0',
            storage: '#f0f',
            redirect: '#f00',
            dom: '#0ff',
            error: '#f00'
        };
        return colors[category] || '#0f0';
    }

    monitorFormSubmission() {
        // Monitor form element
        const form = document.getElementById('merchantForm');
        if (form) {
            this.log('form', 'Form element found', {
                id: form.id,
                action: form.action,
                method: form.method
            });
            
            // Monitor submit events
            form.addEventListener('submit', (e) => {
                this.log('form', 'Form submit event fired', {
                    defaultPrevented: e.defaultPrevented,
                    target: e.target.id
                });
            }, true);
        } else {
            this.log('error', 'Form element not found!');
        }
        
        // Monitor submit button
        const submitBtn = document.getElementById('submitBtn');
        if (submitBtn) {
            this.log('form', 'Submit button found', {
                type: submitBtn.type,
                disabled: submitBtn.disabled,
                onclick: typeof submitBtn.onclick
            });
            
            submitBtn.addEventListener('click', (e) => {
                this.log('form', 'Submit button clicked', {
                    type: e.type,
                    target: e.target.id,
                    timestamp: Date.now()
                });
            }, true);
        }
    }

    monitorSessionStorage() {
        // Override sessionStorage.setItem
        const originalSetItem = sessionStorage.setItem.bind(sessionStorage);
        sessionStorage.setItem = (key, value) => {
            this.log('storage', `sessionStorage.setItem: ${key}`, {
                key,
                valueLength: value ? value.length : 0,
                valuePreview: value ? value.substring(0, 200) : null
            });
            return originalSetItem(key, value);
        };
        
        // Override sessionStorage.getItem
        const originalGetItem = sessionStorage.getItem.bind(sessionStorage);
        sessionStorage.getItem = (key) => {
            const value = originalGetItem(key);
            this.log('storage', `sessionStorage.getItem: ${key}`, {
                key,
                found: value !== null,
                valueLength: value ? value.length : 0
            });
            return value;
        };
        
        // Log current sessionStorage state
        this.log('storage', 'Current sessionStorage state', {
            keys: Object.keys(sessionStorage),
            merchantData: sessionStorage.getItem('merchantData') ? 'exists' : 'missing',
            merchantApiResults: sessionStorage.getItem('merchantApiResults') ? 'exists' : 'missing'
        });
    }

    monitorRedirects() {
        // Monitor location changes
        let currentUrl = window.location.href;
        
        const checkUrlChange = () => {
            if (window.location.href !== currentUrl) {
                this.log('redirect', 'URL changed', {
                    from: currentUrl,
                    to: window.location.href,
                    timestamp: Date.now()
                });
                currentUrl = window.location.href;
            }
        };
        
        // Check periodically
        setInterval(checkUrlChange, 100);
        
        // Monitor location.assign
        const originalAssign = window.location.assign.bind(window.location);
        window.location.assign = (url) => {
            this.log('redirect', 'window.location.assign called', { url });
            return originalAssign(url);
        };
        
        // Monitor location.replace
        const originalReplace = window.location.replace.bind(window.location);
        window.location.replace = (url) => {
            this.log('redirect', 'window.location.replace called', { url });
            return originalReplace(url);
        };
        
        // Monitor location.href assignment
        Object.defineProperty(window.location, 'href', {
            get: () => window.location.href,
            set: (url) => {
                this.log('redirect', 'window.location.href set', { url });
                window.location.href = url;
            }
        });
    }

    monitorDOMChanges() {
        // Monitor for merchant-details tab container
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'childList') {
                    mutation.addedNodes.forEach((node) => {
                        if (node.nodeType === 1) { // Element node
                            if (node.id === 'merchant-details' || 
                                node.classList?.contains('tab-content') ||
                                node.querySelector?.('#merchant-details')) {
                                this.log('dom', 'Merchant details tab container detected', {
                                    id: node.id,
                                    className: node.className,
                                    hasActive: node.classList?.contains('active')
                                });
                            }
                            
                            // Check for merchant detail fields
                            const fields = ['businessName', 'websiteUrl', 'streetAddress', 'city', 'state', 'country'];
                            fields.forEach(fieldId => {
                                if (node.id === fieldId || node.querySelector?.(`#${fieldId}`)) {
                                    this.log('dom', `Merchant detail field found: ${fieldId}`, {
                                        id: node.id,
                                        tagName: node.tagName
                                    });
                                }
                            });
                        }
                    });
                }
            });
        });
        
        observer.observe(document.body, {
            childList: true,
            subtree: true
        });
        
        this.log('dom', 'DOM mutation observer started');
    }

    // Export logs for analysis
    exportLogs() {
        const data = JSON.stringify(this.logs, null, 2);
        const blob = new Blob([data], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `form-flow-debug-${Date.now()}.json`;
        a.click();
        URL.revokeObjectURL(url);
    }

    // Get summary
    getSummary() {
        const categories = {};
        this.logs.forEach(log => {
            categories[log.category] = (categories[log.category] || 0) + 1;
        });
        
        return {
            totalLogs: this.logs.length,
            categories,
            duration: Date.now() - this.startTime,
            currentUrl: window.location.href,
            sessionStorage: {
                merchantData: sessionStorage.getItem('merchantData') ? 'exists' : 'missing',
                merchantApiResults: sessionStorage.getItem('merchantApiResults') ? 'exists' : 'missing'
            }
        };
    }
}

// Initialize debugger
if (typeof window !== 'undefined') {
    window.formFlowDebugger = new FormFlowDebugger();
    
    // Make it globally available
    window.debugFormFlow = {
        export: () => window.formFlowDebugger.exportLogs(),
        summary: () => window.formFlowDebugger.getSummary(),
        logs: () => window.formFlowDebugger.logs
    };
    
    console.log('✅ Form Flow Debugger initialized. Press Ctrl+Shift+D to toggle panel.');
    console.log('📊 Use window.debugFormFlow.summary() to get a summary');
    console.log('💾 Use window.debugFormFlow.export() to export logs');
}

