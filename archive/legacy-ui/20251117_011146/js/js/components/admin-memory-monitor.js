/**
 * Admin Memory Monitor Component
 * Displays memory profile and history charts
 */
class AdminMemoryMonitor {
    constructor() {
        this.memoryChart = null;
        this.historyData = [];
        this.updateInterval = null;
        this.isUpdating = false;
        this.cache = new Map();
        this.cacheTimeout = 5000; // 5 second cache
    }

    /**
     * Initialize the memory monitor
     */
    init(containerId = 'memoryMonitor') {
        this.container = document.getElementById(containerId);
        if (!this.container) {
            console.error('Memory monitor container not found');
            return;
        }

        this.createChart();
        this.startAutoRefresh();
    }

    /**
     * Create the memory history chart
     */
    createChart() {
        const canvas = document.createElement('canvas');
        this.container.innerHTML = '';
        this.container.appendChild(canvas);

        const ctx = canvas.getContext('2d');
        
        this.memoryChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [
                    {
                        label: 'Heap Alloc (MB)',
                        data: [],
                        borderColor: '#4a90e2',
                        backgroundColor: 'rgba(74, 144, 226, 0.1)',
                        tension: 0.4
                    },
                    {
                        label: 'Heap Sys (MB)',
                        data: [],
                        borderColor: '#28a745',
                        backgroundColor: 'rgba(40, 167, 69, 0.1)',
                        tension: 0.4
                    },
                    {
                        label: 'Heap Inuse (MB)',
                        data: [],
                        borderColor: '#ffc107',
                        backgroundColor: 'rgba(255, 193, 7, 0.1)',
                        tension: 0.4
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Memory (MB)'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                }
            }
        });
    }

    /**
     * Fetch memory profile history
     */
    async fetchMemoryHistory(limit = 50) {
        // Check cache first
        const cacheKey = `memory-history-${limit}`;
        const cached = this.cache.get(cacheKey);
        if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
            return cached.data;
        }

        try {
            const response = await fetch(`/api/v1/memory/profile/history?limit=${limit}`, {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            
            // Cache the result
            this.cache.set(cacheKey, {
                data,
                timestamp: Date.now()
            });
            
            return data;
        } catch (error) {
            console.error('Error fetching memory history:', error);
            return null;
        }
    }

    /**
     * Fetch current memory profile
     */
    async fetchMemoryProfile() {
        try {
            const response = await fetch('/api/v1/memory/profile', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Error fetching memory profile:', error);
            return null;
        }
    }

    /**
     * Update the chart with new data
     */
    async updateChart() {
        const history = await this.fetchMemoryHistory();
        if (!history || !history.profiles) {
            return;
        }

        const profiles = history.profiles.slice(-20); // Last 20 data points
        const labels = profiles.map((p, i) => {
            const date = new Date(p.timestamp);
            return date.toLocaleTimeString();
        });

        const heapAlloc = profiles.map(p => (p.heap?.alloc || 0) / 1024 / 1024);
        const heapSys = profiles.map(p => (p.heap?.sys || 0) / 1024 / 1024);
        const heapInuse = profiles.map(p => (p.heap?.inuse || 0) / 1024 / 1024);

        if (this.memoryChart) {
            this.memoryChart.data.labels = labels;
            this.memoryChart.data.datasets[0].data = heapAlloc;
            this.memoryChart.data.datasets[1].data = heapSys;
            this.memoryChart.data.datasets[2].data = heapInuse;
            this.memoryChart.update('none');
        }
    }

    /**
     * Update current memory metrics
     */
    async updateMetrics() {
        const profile = await this.fetchMemoryProfile();
        if (!profile) {
            return;
        }

        // Update memory usage display
        const heapAllocMB = (profile.heap?.alloc || 0) / 1024 / 1024;
        const element = document.getElementById('memoryUsage');
        if (element) {
            element.textContent = `${heapAllocMB.toFixed(2)} MB`;
        }

        // Update GC cycles
        const gcCycles = profile.gc?.num_gc || 0;
        const gcElement = document.getElementById('gcCycles');
        if (gcElement) {
            gcElement.textContent = gcCycles.toLocaleString();
        }
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh(interval = 30000) {
        // Use requestAnimationFrame for smoother updates
        this.updateChart();
        this.updateMetrics();
        
        this.updateInterval = setInterval(() => {
            // Throttle updates to prevent excessive API calls
            if (this.isUpdating) return;
            this.isUpdating = true;
            
            requestAnimationFrame(() => {
                this.updateChart();
                this.updateMetrics();
                this.isUpdating = false;
            });
        }, interval);
    }

    /**
     * Stop auto-refresh
     */
    stopAutoRefresh() {
        if (this.updateInterval) {
            clearInterval(this.updateInterval);
            this.updateInterval = null;
        }
    }

    /**
     * Get authentication token from localStorage or cookie
     */
    getAuthToken() {
        // Try localStorage first
        const token = localStorage.getItem('auth_token') || localStorage.getItem('access_token');
        if (token) {
            return token;
        }

        // Try to get from cookie
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
     * Destroy the component
     */
    destroy() {
        this.stopAutoRefresh();
        if (this.memoryChart) {
            this.memoryChart.destroy();
            this.memoryChart = null;
        }
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AdminMemoryMonitor;
}

