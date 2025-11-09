/**
 * Event Listener Component
 * Provides UI for displaying and filtering events
 */
class EventListener {
    constructor(containerId, options = {}) {
        this.containerId = containerId;
        this.container = document.getElementById(containerId);
        this.options = {
            maxEvents: options.maxEvents || 100,
            autoScroll: options.autoScroll !== false,
            ...options
        };
        this.events = [];
        this.filters = {
            type: null,
            level: null
        };
        this.eventStream = null;
    }

    /**
     * Initialize the event listener
     */
    async init() {
        if (!this.container) {
            console.error(`Event listener container ${this.containerId} not found`);
            return;
        }

        this.render();
        this.initEventStream();
    }

    /**
     * Initialize event stream connection
     */
    initEventStream() {
        if (typeof EventStream !== 'undefined') {
            this.eventStream = new EventStream({
                url: '/api/v1/events/stream',
                useWebSocket: true
            });

            this.eventStream.on('*', (data) => {
                this.addEvent(data);
            });

            this.eventStream.init();
        }
    }

    /**
     * Render the event listener UI
     */
    render() {
        this.container.innerHTML = `
            <div class="event-listener">
                <div class="event-listener-header">
                    <h3><i class="fas fa-stream"></i> Event Stream</h3>
                    <div class="event-filters">
                        <select id="eventTypeFilter" onchange="eventListener.filterEvents()">
                            <option value="">All Types</option>
                            <option value="merchant">Merchant</option>
                            <option value="risk">Risk</option>
                            <option value="compliance">Compliance</option>
                        </select>
                        <select id="eventLevelFilter" onchange="eventListener.filterEvents()">
                            <option value="">All Levels</option>
                            <option value="info">Info</option>
                            <option value="warning">Warning</option>
                            <option value="error">Error</option>
                        </select>
                        <button class="btn btn-sm" onclick="eventListener.clearEvents()">
                            <i class="fas fa-trash"></i> Clear
                        </button>
                    </div>
                </div>
                <div class="event-list" id="eventList"></div>
            </div>
        `;

        this.addStyles();
    }

    /**
     * Add event to the list
     */
    addEvent(eventData) {
        this.events.unshift(eventData);
        
        if (this.events.length > this.options.maxEvents) {
            this.events = this.events.slice(0, this.options.maxEvents);
        }

        this.renderEvents();
    }

    /**
     * Render events list
     */
    renderEvents() {
        const container = document.getElementById('eventList');
        if (!container) return;

        const filteredEvents = this.filterEventsData();

        container.innerHTML = filteredEvents.map(event => `
            <div class="event-item event-${event.level || 'info'}">
                <div class="event-time">${this.formatTime(event.timestamp)}</div>
                <div class="event-type">${event.type || 'event'}</div>
                <div class="event-message">${event.message || JSON.stringify(event.data)}</div>
            </div>
        `).join('');

        if (this.options.autoScroll) {
            container.scrollTop = 0;
        }
    }

    /**
     * Filter events data
     */
    filterEventsData() {
        return this.events.filter(event => {
            if (this.filters.type && event.type !== this.filters.type) {
                return false;
            }
            if (this.filters.level && event.level !== this.filters.level) {
                return false;
            }
            return true;
        });
    }

    /**
     * Apply filters
     */
    filterEvents() {
        const typeFilter = document.getElementById('eventTypeFilter');
        const levelFilter = document.getElementById('eventLevelFilter');

        this.filters.type = typeFilter ? typeFilter.value : null;
        this.filters.level = levelFilter ? levelFilter.value : null;

        this.renderEvents();
    }

    /**
     * Clear all events
     */
    clearEvents() {
        this.events = [];
        this.renderEvents();
    }

    /**
     * Format time
     */
    formatTime(timestamp) {
        if (!timestamp) return '';
        const date = new Date(timestamp);
        return date.toLocaleTimeString();
    }

    /**
     * Add styles
     */
    addStyles() {
        if (document.getElementById('eventListenerStyles')) {
            return;
        }

        const style = document.createElement('style');
        style.id = 'eventListenerStyles';
        style.textContent = `
            .event-listener {
                background: white;
                border-radius: 8px;
                padding: 1rem;
            }

            .event-listener-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 1rem;
            }

            .event-filters {
                display: flex;
                gap: 0.5rem;
            }

            .event-list {
                max-height: 400px;
                overflow-y: auto;
            }

            .event-item {
                padding: 0.75rem;
                border-left: 3px solid #4a90e2;
                margin-bottom: 0.5rem;
                background: #f8f9fa;
                border-radius: 4px;
            }

            .event-item.event-error {
                border-left-color: #dc3545;
            }

            .event-item.event-warning {
                border-left-color: #ffc107;
            }

            .event-time {
                font-size: 0.75rem;
                color: #666;
                margin-bottom: 0.25rem;
            }

            .event-type {
                font-weight: 600;
                margin-bottom: 0.25rem;
            }

            .event-message {
                font-size: 0.9rem;
                color: #333;
            }
        `;
        document.head.appendChild(style);
    }

    /**
     * Cleanup
     */
    destroy() {
        if (this.eventStream) {
            this.eventStream.close();
        }
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.EventListener = EventListener;
    window.eventListener = null;
}

