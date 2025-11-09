/**
 * Risk Drag and Drop Handlers
 * Provides drag-and-drop functionality for risk configuration
 */
class RiskDragDrop {
    constructor(containerId, options = {}) {
        this.containerId = containerId;
        this.container = document.getElementById(containerId);
        this.options = {
            onDragStart: null,
            onDrag: null,
            onDragEnd: null,
            ...options
        };
        this.draggedElement = null;
        this.dragOffset = { x: 0, y: 0 };
    }

    /**
     * Initialize drag and drop
     */
    init() {
        if (!this.container) {
            console.error(`Drag-drop container ${this.containerId} not found`);
            return;
        }

        this.bindEvents();
    }

    /**
     * Bind drag events
     */
    bindEvents() {
        // Make draggable elements
        const draggableElements = this.container.querySelectorAll('[data-draggable="true"]');
        draggableElements.forEach(element => {
            element.addEventListener('mousedown', (e) => this.handleDragStart(e));
        });

        // Global mouse events
        document.addEventListener('mousemove', (e) => this.handleDrag(e));
        document.addEventListener('mouseup', (e) => this.handleDragEnd(e));
    }

    /**
     * Handle drag start
     */
    handleDragStart(event) {
        const element = event.target.closest('[data-draggable="true"]');
        if (!element) return;

        this.draggedElement = element;
        const rect = element.getBoundingClientRect();
        this.dragOffset = {
            x: event.clientX - rect.left,
            y: event.clientY - rect.top
        };

        element.style.opacity = '0.5';
        element.style.cursor = 'grabbing';

        // Call custom handler
        if (this.options.onDragStart) {
            this.options.onDragStart(element, event);
        }

        // Call global function for backward compatibility
        if (typeof dragstarted === 'function') {
            dragstarted(event, element);
        }

        event.preventDefault();
    }

    /**
     * Handle drag
     */
    handleDrag(event) {
        if (!this.draggedElement) return;

        const x = event.clientX - this.dragOffset.x;
        const y = event.clientY - this.dragOffset.y;

        this.draggedElement.style.position = 'absolute';
        this.draggedElement.style.left = `${x}px`;
        this.draggedElement.style.top = `${y}px`;
        this.draggedElement.style.zIndex = '1000';

        // Call custom handler
        if (this.options.onDrag) {
            this.options.onDrag(this.draggedElement, event, { x, y });
        }

        // Call global function for backward compatibility
        if (typeof dragged === 'function') {
            dragged(event, this.draggedElement, { x, y });
        }
    }

    /**
     * Handle drag end
     */
    handleDragEnd(event) {
        if (!this.draggedElement) return;

        const element = this.draggedElement;
        element.style.opacity = '1';
        element.style.cursor = 'grab';

        // Call custom handler
        if (this.options.onDragEnd) {
            this.options.onDragEnd(element, event);
        }

        // Call global function for backward compatibility
        if (typeof dragended === 'function') {
            dragended(event, element);
        }

        this.draggedElement = null;
    }

    /**
     * Make element draggable
     */
    makeDraggable(element) {
        if (!element) return;

        element.setAttribute('data-draggable', 'true');
        element.style.cursor = 'grab';
        element.addEventListener('mousedown', (e) => this.handleDragStart(e));
    }

    /**
     * Make element not draggable
     */
    makeNotDraggable(element) {
        if (!element) return;

        element.removeAttribute('data-draggable');
        element.style.cursor = 'default';
    }
}

// Global functions for backward compatibility
function dragstarted(event, element) {
    if (window.riskDragDrop) {
        window.riskDragDrop.handleDragStart(event);
    }
}

function dragged(event, element, position) {
    if (window.riskDragDrop) {
        window.riskDragDrop.handleDrag(event);
    }
}

function dragended(event, element) {
    if (window.riskDragDrop) {
        window.riskDragDrop.handleDragEnd(event);
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.RiskDragDrop = RiskDragDrop;
}

