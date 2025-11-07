/**
 * Fallback Notification Component
 * 
 * Displays user-friendly notifications when fallback data is used.
 * Supports different severity levels and auto-dismiss functionality.
 */

class FallbackNotification {
    constructor(options = {}) {
        this.containerId = options.containerId || 'fallback-notifications';
        this.autoDismiss = options.autoDismiss !== false; // Default: true
        this.autoDismissDelay = options.autoDismissDelay || 5000; // 5 seconds
        this.notifications = new Map();
        this.init();
    }
    
    /**
     * Initialize the notification container
     */
    init() {
        // Create container if it doesn't exist
        let container = document.getElementById(this.containerId);
        if (!container) {
            container = document.createElement('div');
            container.id = this.containerId;
            container.className = 'fallback-notifications-container';
            container.style.cssText = `
                position: fixed;
                top: 20px;
                right: 20px;
                z-index: 10000;
                max-width: 400px;
                display: flex;
                flex-direction: column;
                gap: 10px;
            `;
            document.body.appendChild(container);
        }
        this.container = container;
    }
    
    /**
     * Show a fallback notification
     * 
     * @param {Object} options - Notification options
     * @param {string} options.severity - Severity level: 'info', 'warning', 'error'
     * @param {string} options.message - Notification message
     * @param {string} options.source - Data source that failed (e.g., 'merchant-api', 'analytics-api')
     * @param {string} options.reason - Reason for fallback (optional)
     * @param {number} options.expectedRecovery - Expected recovery time in seconds (optional)
     * @param {boolean} options.dismissible - Whether notification can be dismissed (default: true)
     * @returns {string} Notification ID
     */
    show(options) {
        const {
            severity = 'warning',
            message,
            source,
            reason,
            expectedRecovery,
            dismissible = true
        } = options;
        
        if (!message) {
            console.warn('FallbackNotification: message is required');
            return null;
        }
        
        const id = `fallback-notification-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        const notification = this.createNotificationElement({
            id,
            severity,
            message,
            source,
            reason,
            expectedRecovery,
            dismissible
        });
        
        this.container.appendChild(notification);
        this.notifications.set(id, notification);
        
        // Auto-dismiss if enabled
        if (this.autoDismiss && severity !== 'error') {
            setTimeout(() => {
                this.dismiss(id);
            }, this.autoDismissDelay);
        }
        
        return id;
    }
    
    /**
     * Create notification element
     */
    createNotificationElement({ id, severity, message, source, reason, expectedRecovery, dismissible }) {
        const notification = document.createElement('div');
        notification.id = id;
        notification.className = `fallback-notification fallback-notification-${severity}`;
        notification.style.cssText = `
            padding: 12px 16px;
            border-radius: 4px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
            background-color: ${this.getBackgroundColor(severity)};
            color: ${this.getTextColor(severity)};
            border-left: 4px solid ${this.getBorderColor(severity)};
            display: flex;
            align-items: flex-start;
            gap: 12px;
            animation: slideIn 0.3s ease-out;
        `;
        
        // Icon
        const icon = document.createElement('span');
        icon.className = 'fallback-notification-icon';
        icon.textContent = this.getIcon(severity);
        icon.style.cssText = 'font-size: 18px; flex-shrink: 0;';
        notification.appendChild(icon);
        
        // Content
        const content = document.createElement('div');
        content.style.cssText = 'flex: 1; min-width: 0;';
        
        // Message
        const messageEl = document.createElement('div');
        messageEl.className = 'fallback-notification-message';
        messageEl.textContent = message;
        messageEl.style.cssText = 'font-weight: 500; margin-bottom: 4px;';
        content.appendChild(messageEl);
        
        // Details
        const details = document.createElement('div');
        details.className = 'fallback-notification-details';
        details.style.cssText = 'font-size: 12px; opacity: 0.8;';
        
        if (source) {
            const sourceEl = document.createElement('div');
            sourceEl.textContent = `Source: ${source}`;
            details.appendChild(sourceEl);
        }
        
        if (reason) {
            const reasonEl = document.createElement('div');
            reasonEl.textContent = `Reason: ${reason}`;
            details.appendChild(reasonEl);
        }
        
        if (expectedRecovery) {
            const recoveryEl = document.createElement('div');
            recoveryEl.textContent = `Expected recovery: ${expectedRecovery} seconds`;
            details.appendChild(recoveryEl);
        }
        
        if (details.children.length > 0) {
            content.appendChild(details);
        }
        
        notification.appendChild(content);
        
        // Dismiss button
        if (dismissible) {
            const dismissBtn = document.createElement('button');
            dismissBtn.className = 'fallback-notification-dismiss';
            dismissBtn.textContent = '×';
            dismissBtn.style.cssText = `
                background: none;
                border: none;
                font-size: 20px;
                cursor: pointer;
                padding: 0;
                width: 24px;
                height: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
                opacity: 0.7;
                flex-shrink: 0;
            `;
            dismissBtn.onclick = () => this.dismiss(id);
            dismissBtn.onmouseenter = () => dismissBtn.style.opacity = '1';
            dismissBtn.onmouseleave = () => dismissBtn.style.opacity = '0.7';
            notification.appendChild(dismissBtn);
        }
        
        // Add CSS animation if not already added
        if (!document.getElementById('fallback-notification-styles')) {
            const style = document.createElement('style');
            style.id = 'fallback-notification-styles';
            style.textContent = `
                @keyframes slideIn {
                    from {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                    to {
                        transform: translateX(0);
                        opacity: 1;
                    }
                }
                @keyframes slideOut {
                    from {
                        transform: translateX(0);
                        opacity: 1;
                    }
                    to {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                }
            `;
            document.head.appendChild(style);
        }
        
        return notification;
    }
    
    /**
     * Get background color for severity level
     */
    getBackgroundColor(severity) {
        switch (severity) {
            case 'error':
                return '#fee';
            case 'warning':
                return '#fffbeb';
            case 'info':
            default:
                return '#eff6ff';
        }
    }
    
    /**
     * Get text color for severity level
     */
    getTextColor(severity) {
        switch (severity) {
            case 'error':
                return '#991b1b';
            case 'warning':
                return '#92400e';
            case 'info':
            default:
                return '#1e40af';
        }
    }
    
    /**
     * Get border color for severity level
     */
    getBorderColor(severity) {
        switch (severity) {
            case 'error':
                return '#dc2626';
            case 'warning':
                return '#f59e0b';
            case 'info':
            default:
                return '#3b82f6';
        }
    }
    
    /**
     * Get icon for severity level
     */
    getIcon(severity) {
        switch (severity) {
            case 'error':
                return '⚠️';
            case 'warning':
                return '⚠️';
            case 'info':
            default:
                return 'ℹ️';
        }
    }
    
    /**
     * Dismiss a notification
     * 
     * @param {string} id - Notification ID
     */
    dismiss(id) {
        const notification = this.notifications.get(id);
        if (!notification) {
            return;
        }
        
        // Animate out
        notification.style.animation = 'slideOut 0.3s ease-out';
        
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
            this.notifications.delete(id);
        }, 300);
    }
    
    /**
     * Dismiss all notifications
     */
    dismissAll() {
        const ids = Array.from(this.notifications.keys());
        ids.forEach(id => this.dismiss(id));
    }
    
    /**
     * Show recovery notification when data source recovers
     * 
     * @param {string} source - Data source that recovered
     */
    showRecovery(source) {
        return this.show({
            severity: 'info',
            message: 'Data source recovered',
            source: source,
            dismissible: true,
            autoDismiss: true,
            autoDismissDelay: 3000
        });
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = FallbackNotification;
}

// Make available globally
if (typeof window !== 'undefined') {
    window.FallbackNotification = FallbackNotification;
}

