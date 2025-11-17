/**
 * Admin Dashboard Main Controller
 * Manages the admin dashboard functionality
 */
class AdminDashboard {
    constructor() {
        this.memoryMonitor = null;
        this.systemMetrics = null;
        this.isAdmin = false;
    }

    /**
     * Initialize the admin dashboard
     */
    async init() {
        // Check admin access first
        const hasAccess = await this.checkAdminAccess();
        
        if (!hasAccess) {
            this.showAccessDenied();
            return;
        }

        this.showDashboard();
        this.initializeComponents();
    }

    /**
     * Check if user has admin access
     */
    async checkAdminAccess() {
        try {
            // Try to get user info from token or API
            const token = this.getAuthToken();
            if (!token) {
                return false;
            }

            // Decode JWT token to check role (basic check)
            const payload = this.decodeJWT(token);
            if (payload && payload.role === 'admin') {
                this.isAdmin = true;
                return true;
            }

            // Alternative: Check via API endpoint if available
            const response = await fetch('/api/v1/users/profile', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (response.ok) {
                const user = await response.json();
                if (user.role === 'admin' || user.role === 'Admin') {
                    this.isAdmin = true;
                    return true;
                }
            }

            return false;
        } catch (error) {
            console.error('Error checking admin access:', error);
            return false;
        }
    }

    /**
     * Decode JWT token (basic implementation)
     */
    decodeJWT(token) {
        try {
            const base64Url = token.split('.')[1];
            const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
            const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
                return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
            }).join(''));
            return JSON.parse(jsonPayload);
        } catch (error) {
            console.error('Error decoding JWT:', error);
            return null;
        }
    }

    /**
     * Show access denied message
     */
    showAccessDenied() {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('accessDenied').style.display = 'block';
        document.getElementById('adminDashboard').style.display = 'none';
    }

    /**
     * Show admin dashboard
     */
    showDashboard() {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('accessDenied').style.display = 'none';
        document.getElementById('adminDashboard').style.display = 'block';
    }

    /**
     * Initialize dashboard components
     */
    initializeComponents() {
        // Initialize memory monitor
        if (typeof AdminMemoryMonitor !== 'undefined') {
            this.memoryMonitor = new AdminMemoryMonitor();
            this.memoryMonitor.init();
        }

        // Initialize system metrics
        if (typeof AdminSystemMetrics !== 'undefined') {
            this.systemMetrics = new AdminSystemMetrics();
            this.systemMetrics.init();
        }

        // Set up global reference for button handlers
        window.adminDashboard = this;
    }

    /**
     * Optimize memory
     * @param {HTMLElement} buttonElement - Optional button element to update
     */
    async optimizeMemory(buttonElement = null) {
        // Get button element - try parameter first, then find by ID, then fallback to event
        let button = buttonElement;
        if (!button) {
            button = document.getElementById('optimizeMemoryBtn') || 
                    document.querySelector('button[onclick*="optimizeMemory"]');
        }
        if (!button && typeof event !== 'undefined' && event.target) {
            button = event.target;
        }
        
        const originalText = button ? button.innerHTML : '';
        
        try {
            if (button) {
                button.disabled = true;
                button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Optimizing...';
            }

            const response = await fetch('/api/v1/memory/optimize', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            alert('Memory optimization completed successfully!');
            
            // Refresh metrics
            if (this.memoryMonitor) {
                this.memoryMonitor.updateChart();
                this.memoryMonitor.updateMetrics();
            }

            if (button) {
                button.disabled = false;
                button.innerHTML = originalText;
            }
        } catch (error) {
            console.error('Error optimizing memory:', error);
            alert('Failed to optimize memory: ' + error.message);
            if (button) {
                button.disabled = false;
                button.innerHTML = originalText;
            }
        }
    }

    /**
     * Refresh all metrics
     */
    refreshMetrics() {
        if (this.memoryMonitor) {
            this.memoryMonitor.updateChart();
            this.memoryMonitor.updateMetrics();
        }
        if (this.systemMetrics) {
            this.systemMetrics.updateMetrics();
        }
    }

    /**
     * Get authentication token
     */
    getAuthToken() {
        const token = localStorage.getItem('auth_token') || localStorage.getItem('access_token');
        if (token) {
            return token;
        }

        const cookies = document.cookie.split(';');
        for (let cookie of cookies) {
            const [name, value] = cookie.trim().split('=');
            if (name === 'auth_token' || name === 'access_token') {
                return value;
            }
        }

        return null;
    }

    /**
     * Cleanup on page unload
     */
    destroy() {
        if (this.memoryMonitor) {
            this.memoryMonitor.destroy();
        }
        if (this.systemMetrics) {
            this.systemMetrics.destroy();
        }
    }
}

// Initialize on page load
if (typeof window !== 'undefined') {
    window.AdminDashboard = AdminDashboard;
}

