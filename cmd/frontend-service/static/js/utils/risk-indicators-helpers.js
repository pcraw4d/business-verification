/**
 * Risk Indicators Helper Functions
 * 
 * Extracted from enhanced-risk-indicators.html to provide utility functions
 * for risk level calculations, badge classes, icons, and other helper methods.
 */

class RiskIndicatorsHelpers {
    
    /**
     * Get risk level from score
     * @param {number} score - Risk score (0-100)
     * @returns {string} Risk level (LOW, MEDIUM, HIGH, CRITICAL)
     */
    static getRiskLevel(score) {
        if (score <= 25) return 'LOW';
        if (score <= 50) return 'MEDIUM';
        if (score <= 75) return 'HIGH';
        return 'CRITICAL';
    }
    
    /**
     * Get risk level in lowercase
     * @param {number} score - Risk score (0-100)
     * @returns {string} Risk level (low, medium, high, critical)
     */
    static getRiskLevelLower(score) {
        return this.getRiskLevel(score).toLowerCase();
    }
    
    /**
     * Get CSS class for risk badge
     * @param {number} score - Risk score (0-100)
     * @returns {string} CSS class name
     */
    static getRiskBadgeClass(score) {
        if (score <= 25) return 'risk-low';
        if (score <= 50) return 'risk-medium';
        if (score <= 75) return 'risk-high';
        return 'risk-critical';
    }
    
    /**
     * Get CSS class for risk icon
     * @param {number} score - Risk score (0-100)
     * @returns {string} CSS class name
     */
    static getRiskIconClass(score) {
        if (score <= 25) return 'risk-icon-low';
        if (score <= 50) return 'risk-icon-medium';
        if (score <= 75) return 'risk-icon-high';
        return 'risk-icon-critical';
    }
    
    /**
     * Get FontAwesome icon for risk category
     * @param {string} category - Risk category
     * @returns {string} FontAwesome icon class
     */
    static getRiskIcon(category) {
        const icons = {
            financial: 'dollar-sign',
            operational: 'cogs',
            regulatory: 'gavel',
            reputational: 'star',
            cybersecurity: 'shield-alt',
            content: 'file-alt'
        };
        return icons[category] || 'exclamation-triangle';
    }
    
    /**
     * Get FontAwesome icon for risk category (extended version)
     * @param {string} category - Risk category
     * @returns {string} FontAwesome icon class
     */
    static getRiskCategoryIcon(category) {
        const icons = {
            financial: 'dollar-sign',
            operational: 'cogs',
            regulatory: 'gavel',
            reputational: 'star',
            cybersecurity: 'shield-alt',
            content: 'file-alt',
            market: 'chart-line',
            compliance: 'gavel',
            security: 'shield-alt',
            reputation: 'star',
            operational: 'cogs',
            credit: 'credit-card',
            liquidity: 'water',
            concentration: 'bullseye',
            country: 'globe',
            sector: 'building'
        };
        return icons[category] || 'exclamation-triangle';
    }
    
    /**
     * Format risk score for display
     * @param {number} score - Risk score
     * @param {number} decimals - Number of decimal places
     * @returns {string} Formatted score
     */
    static formatRiskScore(score, decimals = 0) {
        return Number(score).toFixed(decimals);
    }
    
    /**
     * Calculate trend direction between two scores
     * @param {number} current - Current score
     * @param {number} previous - Previous score
     * @returns {Object} Trend information
     */
    static calculateTrendDirection(current, previous) {
        if (previous === null || previous === undefined) {
            return { direction: 'stable', icon: 'minus', label: 'Stable' };
        }
        
        const diff = current - previous;
        const threshold = 5; // Minimum change to show trend
        
        if (diff > threshold) {
            return { direction: 'up', icon: 'arrow-up', label: 'Rising' };
        } else if (diff < -threshold) {
            return { direction: 'down', icon: 'arrow-down', label: 'Improving' };
        } else {
            return { direction: 'stable', icon: 'minus', label: 'Stable' };
        }
    }
    
    /**
     * Get risk color for a score
     * @param {number} score - Risk score (0-100)
     * @returns {string} Color hex code
     */
    static getRiskColor(score) {
        if (score <= 25) return '#27ae60'; // Green
        if (score <= 50) return '#f39c12'; // Orange
        if (score <= 75) return '#e74c3c'; // Red
        return '#8e44ad'; // Purple
    }
    
    /**
     * Get risk background color for a score
     * @param {number} score - Risk score (0-100)
     * @returns {string} Background color hex code
     */
    static getRiskBackgroundColor(score) {
        if (score <= 25) return '#d5f4e6'; // Light green
        if (score <= 50) return '#fef9e7'; // Light orange
        if (score <= 75) return '#fadbd8'; // Light red
        return '#f4e6f7'; // Light purple
    }
    
    /**
     * Get risk border color for a score
     * @param {number} score - Risk score (0-100)
     * @returns {string} Border color hex code
     */
    static getRiskBorderColor(score) {
        if (score <= 25) return '#27ae60'; // Green
        if (score <= 50) return '#f39c12'; // Orange
        if (score <= 75) return '#e74c3c'; // Red
        return '#8e44ad'; // Purple
    }
    
    /**
     * Get risk range for a level
     * @param {string} level - Risk level
     * @returns {string} Score range
     */
    static getRiskRange(level) {
        const ranges = {
            low: '0-25',
            medium: '26-50',
            high: '51-75',
            critical: '76-100'
        };
        return ranges[level.toLowerCase()] || '0-100';
    }
    
    /**
     * Get risk impact description for a level
     * @param {string} level - Risk level
     * @returns {string} Impact description
     */
    static getRiskImpact(level) {
        const impacts = {
            low: 'Minimal business impact expected',
            medium: 'Moderate business impact possible',
            high: 'Significant business impact likely',
            critical: 'Severe business impact expected'
        };
        return impacts[level.toLowerCase()] || 'Business impact varies';
    }
    
    /**
     * Calculate overall risk score from categories
     * @param {Object} categories - Risk categories with scores
     * @returns {number} Overall risk score
     */
    static calculateOverallRiskScore(categories) {
        const scores = Object.values(categories).map(cat => cat.score || 0);
        if (scores.length === 0) return 0;
        
        // Weighted average (can be customized)
        const weights = {
            financial: 0.25,
            operational: 0.20,
            regulatory: 0.20,
            reputational: 0.15,
            cybersecurity: 0.15,
            content: 0.05
        };
        
        let weightedSum = 0;
        let totalWeight = 0;
        
        Object.entries(categories).forEach(([category, data]) => {
            const weight = weights[category] || 0.1;
            weightedSum += (data.score || 0) * weight;
            totalWeight += weight;
        });
        
        return totalWeight > 0 ? weightedSum / totalWeight : 0;
    }
    
    /**
     * Get risk level from overall score
     * @param {number} score - Overall risk score
     * @returns {string} Risk level
     */
    static getOverallRiskLevel(score) {
        return this.getRiskLevelLower(score);
    }
    
    /**
     * Format category name for display
     * @param {string} category - Category name
     * @returns {string} Formatted category name
     */
    static formatCategoryName(category) {
        return category.charAt(0).toUpperCase() + category.slice(1);
    }
    
    /**
     * Get priority score for sorting
     * @param {string} priority - Priority level
     * @returns {number} Priority score
     */
    static getPriorityScore(priority) {
        const scores = {
            critical: 4,
            high: 3,
            medium: 2,
            low: 1
        };
        return scores[priority.toLowerCase()] || 0;
    }
    
    /**
     * Get severity score for sorting
     * @param {string} severity - Severity level
     * @returns {number} Severity score
     */
    static getSeverityScore(severity) {
        const scores = {
            critical: 4,
            high: 3,
            medium: 2,
            low: 1
        };
        return scores[severity.toLowerCase()] || 0;
    }
    
    /**
     * Generate unique ID for components
     * @param {string} prefix - ID prefix
     * @returns {string} Unique ID
     */
    static generateId(prefix = 'risk') {
        return `${prefix}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }
    
    /**
     * Debounce function for performance
     * @param {Function} func - Function to debounce
     * @param {number} wait - Wait time in ms
     * @returns {Function} Debounced function
     */
    static debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
    
    /**
     * Throttle function for performance
     * @param {Function} func - Function to throttle
     * @param {number} limit - Time limit in ms
     * @returns {Function} Throttled function
     */
    static throttle(func, limit) {
        let inThrottle;
        return function(...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }
    
    /**
     * Format timestamp for display
     * @param {string|Date} timestamp - Timestamp
     * @returns {string} Formatted timestamp
     */
    static formatTimestamp(timestamp) {
        if (!timestamp) return 'Just now';
        
        const date = new Date(timestamp);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);
        
        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
        if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
        if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
        
        return date.toLocaleDateString();
    }
    
    /**
     * Validate risk score
     * @param {number} score - Risk score
     * @returns {boolean} Is valid score
     */
    static isValidRiskScore(score) {
        return typeof score === 'number' && score >= 0 && score <= 100 && !isNaN(score);
    }
    
    /**
     * Normalize risk score to 0-100 range
     * @param {number} score - Risk score
     * @param {number} max - Maximum possible score
     * @returns {number} Normalized score
     */
    static normalizeRiskScore(score, max = 10) {
        if (!this.isValidRiskScore(score)) return 0;
        return Math.min(100, Math.max(0, (score / max) * 100));
    }
    
    /**
     * Get risk level from normalized score
     * @param {number} score - Normalized score (0-100)
     * @returns {string} Risk level
     */
    static getRiskLevelFromScore(score) {
        return this.getRiskLevelLower(score);
    }
    
    /**
     * Create risk level object
     * @param {number} score - Risk score
     * @param {string} category - Risk category
     * @returns {Object} Risk level object
     */
    static createRiskLevel(score, category) {
        const level = this.getRiskLevelLower(score);
        return {
            score: score,
            level: level,
            category: category,
            badgeClass: this.getRiskBadgeClass(score),
            iconClass: this.getRiskIconClass(score),
            icon: this.getRiskIcon(category),
            color: this.getRiskColor(score),
            backgroundColor: this.getRiskBackgroundColor(score),
            borderColor: this.getRiskBorderColor(score),
            range: this.getRiskRange(level),
            impact: this.getRiskImpact(level)
        };
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskIndicatorsHelpers;
}
