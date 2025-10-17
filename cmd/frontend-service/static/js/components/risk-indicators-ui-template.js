/**
 * Risk Indicators UI Template Component
 * 
 * Extracted from enhanced-risk-indicators.html to create reusable HTML templates
 * for the Risk Indicators tab. Provides static methods to generate HTML for
 * all risk indicator sections.
 */

class RiskIndicatorsUITemplate {
    
    /**
     * Generate HTML for risk level badges section
     * @param {Object} riskData - Risk data containing categories and scores
     * @returns {string} HTML string for risk badges
     */
    static getRiskBadgesHTML(riskData) {
        const categories = riskData.categories || {};
        const overallScore = riskData.overallRiskScore || 0;
        const riskLevel = riskData.riskLevel || 'medium';
        
        // Get the top 4 risk categories for badges
        const topCategories = Object.entries(categories)
            .sort(([,a], [,b]) => b.score - a.score)
            .slice(0, 4);
        
        return `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-bold text-gray-900 mb-6">Enhanced Risk Level Badges</h2>
                
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                    ${topCategories.map(([category, data], index) => {
                        const level = data.level || 'medium';
                        const score = data.score || 0;
                        const trend = this.getTrendDirection(score, data.previousScore);
                        
                        return `
                            <div class="risk-indicator text-center">
                                <div class="risk-badge risk-${level} px-4 py-2 rounded-lg text-sm font-bold mb-2 inline-block">
                                    <i class="fas fa-${this.getRiskIcon(category)} mr-2"></i>
                                    ${level.toUpperCase()} RISK
                                </div>
                                <div class="risk-tooltip">
                                    Risk Score: ${this.getRiskRange(level)} | ${this.getRiskImpact(level)}
                                </div>
                                <p class="text-sm text-gray-600">Score: ${Math.round(score)}</p>
                                <div class="risk-trend trend-${trend.direction} mt-2">
                                    <i class="fas fa-${trend.icon} mr-1"></i>
                                    ${trend.label}
                                </div>
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>
        `;
    }
    
    /**
     * Generate HTML for risk heat map section
     * @param {Object} categories - Risk categories with scores and subcategories
     * @returns {string} HTML string for heat map
     */
    static getHeatMapHTML(categories) {
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        
        return `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-bold text-gray-900 mb-6">Risk Heat Map</h2>
                
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    <div>
                        <h3 class="font-medium text-gray-900 mb-4">Risk Intensity by Category</h3>
                        <div class="space-y-4">
                            ${categoryOrder.map(category => {
                                const data = categories[category];
                                if (!data) return '';
                                
                                const subCategories = data.subCategories || {};
                                const heatmapCells = this.generateHeatmapCells(subCategories, category);
                                
                                return `
                                    <div class="flex items-center justify-between">
                                        <div class="flex items-center">
                                            <div class="risk-icon risk-icon-${data.level || 'medium'}">
                                                <i class="fas fa-${this.getRiskIcon(category)} text-xs"></i>
                                            </div>
                                            <span class="text-sm font-medium">${this.formatCategoryName(category)}</span>
                                        </div>
                                        <div class="risk-heatmap">
                                            ${heatmapCells}
                                        </div>
                                    </div>
                                `;
                            }).join('')}
                        </div>
                    </div>
                    
                    <div>
                        <h3 class="font-medium text-gray-900 mb-4">Risk Level Legend</h3>
                        <div class="space-y-3">
                            <div class="flex items-center">
                                <div class="w-4 h-4 bg-green-200 rounded mr-3"></div>
                                <span class="text-sm text-gray-600">Low Risk (0-25)</span>
                            </div>
                            <div class="flex items-center">
                                <div class="w-4 h-4 bg-yellow-200 rounded mr-3"></div>
                                <span class="text-sm text-gray-600">Medium Risk (26-50)</span>
                            </div>
                            <div class="flex items-center">
                                <div class="w-4 h-4 bg-orange-200 rounded mr-3"></div>
                                <span class="text-sm text-gray-600">High Risk (51-75)</span>
                            </div>
                            <div class="flex items-center">
                                <div class="w-4 h-4 bg-red-200 rounded mr-3"></div>
                                <span class="text-sm text-gray-600">Critical Risk (76-100)</span>
                            </div>
                        </div>
                        
                        <div class="mt-6">
                            <h4 class="font-medium text-gray-900 mb-2">Risk Trend Indicators</h4>
                            <div class="space-y-2">
                                <div class="flex items-center">
                                    <div class="risk-trend trend-down mr-3">
                                        <i class="fas fa-arrow-down mr-1"></i>
                                        Improving
                                    </div>
                                    <span class="text-sm text-gray-600">Risk level decreasing</span>
                                </div>
                                <div class="flex items-center">
                                    <div class="risk-trend trend-stable mr-3">
                                        <i class="fas fa-minus mr-1"></i>
                                        Stable
                                    </div>
                                    <span class="text-sm text-gray-600">Risk level unchanged</span>
                                </div>
                                <div class="flex items-center">
                                    <div class="risk-trend trend-up mr-3">
                                        <i class="fas fa-arrow-up mr-1"></i>
                                        Rising
                                    </div>
                                    <span class="text-sm text-gray-600">Risk level increasing</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    /**
     * Generate HTML for progress bars section
     * @param {Object} categories - Risk categories with scores
     * @returns {string} HTML string for progress bars
     */
    static getProgressBarsHTML(categories) {
        const categoryOrder = ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity', 'content'];
        
        return `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-bold text-gray-900 mb-6">Enhanced Risk Progress Indicators</h2>
                
                <div class="space-y-6">
                    ${categoryOrder.map(category => {
                        const data = categories[category];
                        if (!data) return '';
                        
                        const score = Math.round(data.score || 0);
                        const level = data.level || 'medium';
                        const description = this.getRiskDescription(category, level);
                        const lastUpdated = data.lastUpdated || 'Just now';
                        
                        return `
                            <div>
                                <div class="flex items-center justify-between mb-2">
                                    <div class="flex items-center">
                                        <div class="risk-icon risk-icon-${level}">
                                            <i class="fas fa-${this.getRiskIcon(category)} text-xs"></i>
                                        </div>
                                        <span class="font-medium text-gray-900">${this.formatCategoryName(category)} Risk</span>
                                    </div>
                                    <div class="flex items-center space-x-2">
                                        <span class="text-sm font-medium text-gray-900">${score}/100</span>
                                        <span class="risk-badge risk-${level} px-2 py-1 rounded text-xs font-bold">${level.toUpperCase()}</span>
                                    </div>
                                </div>
                                <div class="enhanced-progress">
                                    <div class="enhanced-progress-fill risk-${level}" style="width: ${score}%"></div>
                                </div>
                                <div class="flex justify-between text-xs text-gray-500 mt-1">
                                    <span>${description}</span>
                                    <span>Last updated: ${lastUpdated}</span>
                                </div>
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>
        `;
    }
    
    /**
     * Generate HTML for radar chart section
     * @returns {string} HTML string for radar chart container
     */
    static getRadarChartHTML() {
        return `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-bold text-gray-900 mb-6">Risk Radar Chart</h2>
                
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    <div class="radar-container">
                        <canvas id="riskRadarChart"></canvas>
                    </div>
                    
                    <div>
                        <h3 class="font-medium text-gray-900 mb-4">Risk Category Analysis</h3>
                        <div id="riskCategoryAnalysis" class="space-y-4">
                            <!-- Populated by JavaScript -->
                        </div>
                        
                        <div class="mt-6 p-4 bg-gray-50 rounded-lg">
                            <h4 class="font-medium text-gray-900 mb-2">Risk Summary</h4>
                            <p class="text-sm text-gray-600" id="riskSummary">
                                Overall risk profile analysis will be displayed here.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    /**
     * Generate HTML for alert cards section (NEW)
     * @param {Array} alerts - Array of risk alerts
     * @returns {string} HTML string for alert cards
     */
    static getAlertsHTML(alerts) {
        if (!alerts || alerts.length === 0) {
            return '';
        }
        
        // Sort alerts by severity
        const sortedAlerts = alerts.sort((a, b) => {
            const severityOrder = { critical: 4, high: 3, medium: 2, low: 1 };
            return severityOrder[b.severity] - severityOrder[a.severity];
        });
        
        return `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-bold text-gray-900 mb-6">
                    <i class="fas fa-exclamation-triangle mr-2"></i>
                    Risk Alerts
                </h2>
                
                <div class="space-y-4">
                    ${sortedAlerts.map(alert => `
                        <div class="alert-card alert-${alert.severity} p-4 rounded-lg border-l-4">
                            <div class="flex items-start justify-between">
                                <div class="flex-1">
                                    <div class="flex items-center mb-2">
                                        <i class="fas fa-${this.getAlertIcon(alert.type)} mr-2"></i>
                                        <h3 class="font-semibold text-gray-900">${alert.title}</h3>
                                        <span class="ml-2 alert-badge alert-${alert.severity} px-2 py-1 rounded text-xs font-bold">
                                            ${alert.severity.toUpperCase()}
                                        </span>
                                    </div>
                                    <p class="text-sm text-gray-600 mb-2">${alert.description}</p>
                                    <div class="flex items-center text-xs text-gray-500">
                                        <span class="mr-4">Source: ${this.formatSource(alert.source)}</span>
                                        <span>Detected: ${new Date(alert.detectedAt || Date.now()).toLocaleString()}</span>
                                    </div>
                                </div>
                                <div class="flex space-x-2 ml-4">
                                    <button class="btn btn-sm btn-outline" onclick="acknowledgeAlert('${alert.id}')">
                                        <i class="fas fa-check mr-1"></i>
                                        Acknowledge
                                    </button>
                                    <button class="btn btn-sm btn-primary" onclick="investigateAlert('${alert.id}')">
                                        <i class="fas fa-search mr-1"></i>
                                        Investigate
                                    </button>
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }
    
    /**
     * Generate HTML for recommendations section (NEW)
     * @param {Array} recommendations - Array of recommendations
     * @returns {string} HTML string for recommendations
     */
    static getRecommendationsHTML(recommendations) {
        if (!recommendations || recommendations.length === 0) {
            return `
                <div class="bg-white rounded-lg shadow-lg p-6">
                    <h2 class="text-xl font-bold text-gray-900 mb-6">
                        <i class="fas fa-lightbulb mr-2"></i>
                        Recommended Actions
                    </h2>
                    <div class="text-center text-gray-500 py-8">
                        <i class="fas fa-info-circle text-4xl mb-4"></i>
                        <p>No specific recommendations available at this time.</p>
                    </div>
                </div>
            `;
        }
        
        // Sort by priority
        const sortedRecommendations = recommendations.sort((a, b) => {
            const priorityOrder = { critical: 4, high: 3, medium: 2, low: 1 };
            return priorityOrder[b.priority] - priorityOrder[a.priority];
        });
        
        return `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-bold text-gray-900 mb-6">
                    <i class="fas fa-lightbulb mr-2"></i>
                    Recommended Actions
                </h2>
                
                <div class="space-y-4">
                    ${sortedRecommendations.map(rec => `
                        <div class="recommendation-card recommendation-${rec.priority} p-4 rounded-lg border">
                            <div class="flex items-start justify-between">
                                <div class="flex-1">
                                    <div class="flex items-center mb-2">
                                        <i class="fas fa-${this.getRecommendationIcon(rec.type)} mr-2"></i>
                                        <h3 class="font-semibold text-gray-900">${rec.title}</h3>
                                        <span class="ml-2 recommendation-badge recommendation-${rec.priority} px-2 py-1 rounded text-xs font-bold">
                                            ${rec.priority.toUpperCase()}
                                        </span>
                                    </div>
                                    <p class="text-sm text-gray-600 mb-3">${rec.description}</p>
                                    <div class="flex items-center text-xs text-gray-500 space-x-4">
                                        <span>Impact: ${rec.impactScore || 'N/A'}</span>
                                        <span>Difficulty: ${rec.difficulty || 'Medium'}</span>
                                        <span>Type: ${rec.type || 'General'}</span>
                                    </div>
                                </div>
                                <div class="flex space-x-2 ml-4">
                                    <button class="btn btn-sm btn-outline" onclick="dismissRecommendation('${rec.id}')">
                                        <i class="fas fa-times mr-1"></i>
                                        Dismiss
                                    </button>
                                    <button class="btn btn-sm btn-primary" onclick="implementRecommendation('${rec.id}')">
                                        <i class="fas fa-play mr-1"></i>
                                        Implement
                                    </button>
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }
    
    /**
     * Generate HTML for website risk findings section (NEW)
     * @param {Object} websiteRisks - Website risk data
     * @returns {string} HTML string for website risk findings
     */
    static getWebsiteRiskFindingsHTML(websiteRisks) {
        if (!websiteRisks || Object.keys(websiteRisks).length === 0) {
            return '';
        }
        
        return `
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-bold text-gray-900 mb-6">
                    <i class="fas fa-globe mr-2"></i>
                    Website Risk Findings
                </h2>
                
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <!-- Risky Keywords -->
                    ${websiteRisks.riskyKeywords ? `
                        <div>
                            <h3 class="font-medium text-gray-900 mb-3">Risky Keywords Detected</h3>
                            <div class="space-y-2">
                                ${websiteRisks.riskyKeywords.slice(0, 5).map(keyword => `
                                    <div class="risky-keyword-badge severity-${keyword.severity} p-2 rounded border-l-4">
                                        <div class="flex items-center justify-between">
                                            <span class="font-medium">${keyword.keyword}</span>
                                            <span class="text-xs badge-${keyword.severity}">${keyword.severity.toUpperCase()}</span>
                                        </div>
                                        <p class="text-xs text-gray-600 mt-1">${keyword.context}</p>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                    ` : ''}
                    
                    <!-- Sentiment Analysis -->
                    ${websiteRisks.sentimentAnalysis ? `
                        <div>
                            <h3 class="font-medium text-gray-900 mb-3">Sentiment Analysis</h3>
                            <div class="sentiment-breakdown">
                                <div class="sentiment-bar mb-2">
                                    <div class="positive" style="width: ${websiteRisks.sentimentAnalysis.positive_pct || 0}%"></div>
                                    <div class="neutral" style="width: ${websiteRisks.sentimentAnalysis.neutral_pct || 0}%"></div>
                                    <div class="negative" style="width: ${websiteRisks.sentimentAnalysis.negative_pct || 0}%"></div>
                                </div>
                                <div class="sentiment-labels text-xs">
                                    <span class="positive">${websiteRisks.sentimentAnalysis.positive_pct || 0}% Positive</span>
                                    <span class="neutral">${websiteRisks.sentimentAnalysis.neutral_pct || 0}% Neutral</span>
                                    <span class="negative">${websiteRisks.sentimentAnalysis.negative_pct || 0}% Negative</span>
                                </div>
                            </div>
                        </div>
                    ` : ''}
                </div>
            </div>
        `;
    }
    
    // Helper methods
    
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
    
    static getAlertIcon(type) {
        const icons = {
            risky_keyword: 'exclamation-triangle',
            negative_sentiment: 'frown',
            risk_factor: 'exclamation-circle',
            security: 'shield-alt',
            compliance: 'gavel'
        };
        return icons[type] || 'info-circle';
    }
    
    static getRecommendationIcon(type) {
        const icons = {
            ml_based: 'brain',
            manual_verification: 'user-check',
            document_verification: 'file-check',
            compliance_check: 'gavel',
            security_audit: 'shield-alt'
        };
        return icons[type] || 'lightbulb';
    }
    
    static formatCategoryName(category) {
        return category.charAt(0).toUpperCase() + category.slice(1);
    }
    
    static formatSource(source) {
        return source.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
    }
    
    static getRiskRange(level) {
        const ranges = {
            low: '0-25',
            medium: '26-50',
            high: '51-75',
            critical: '76-100'
        };
        return ranges[level] || '0-100';
    }
    
    static getRiskImpact(level) {
        const impacts = {
            low: 'Minimal business impact expected',
            medium: 'Moderate business impact possible',
            high: 'Significant business impact likely',
            critical: 'Severe business impact expected'
        };
        return impacts[level] || 'Business impact varies';
    }
    
    static getRiskDescription(category, level) {
        const descriptions = {
            financial: {
                low: 'Excellent financial health',
                medium: 'Some financial concerns',
                high: 'Significant financial risks',
                critical: 'Critical financial issues'
            },
            operational: {
                low: 'Smooth operations',
                medium: 'Some operational challenges',
                high: 'Operational difficulties',
                critical: 'Critical operational issues'
            },
            regulatory: {
                low: 'Good compliance standing',
                medium: 'Minor compliance issues',
                high: 'Compliance issues detected',
                critical: 'Critical compliance violations'
            },
            reputational: {
                low: 'Strong reputation',
                medium: 'Mixed reputation signals',
                high: 'Reputation concerns',
                critical: 'Severe reputation damage'
            },
            cybersecurity: {
                low: 'Strong security posture',
                medium: 'Some security concerns',
                high: 'Security vulnerabilities found',
                critical: 'Critical security breaches'
            },
            content: {
                low: 'Clean content profile',
                medium: 'Some content concerns',
                high: 'Content risks detected',
                critical: 'High-risk content found'
            }
        };
        return descriptions[category]?.[level] || 'Risk assessment in progress';
    }
    
    static getTrendDirection(current, previous) {
        if (!previous) {
            return { direction: 'stable', icon: 'minus', label: 'Stable' };
        }
        
        const diff = current - previous;
        if (diff > 5) {
            return { direction: 'up', icon: 'arrow-up', label: 'Rising' };
        } else if (diff < -5) {
            return { direction: 'down', icon: 'arrow-down', label: 'Improving' };
        } else {
            return { direction: 'stable', icon: 'minus', label: 'Stable' };
        }
    }
    
    static generateHeatmapCells(subCategories, category) {
        const cellTitles = {
            financial: ['Revenue Risk', 'Cash Flow Risk', 'Debt Risk', 'Credit Risk', 'Market Risk'],
            operational: ['Process Risk', 'Staff Risk', 'Technology Risk', 'Supply Chain Risk', 'Quality Risk'],
            regulatory: ['Compliance Risk', 'License Risk', 'Audit Risk', 'Legal Risk', 'Policy Risk'],
            reputational: ['Brand Risk', 'Review Risk', 'Media Risk', 'Social Risk', 'Trust Risk'],
            cybersecurity: ['Data Risk', 'Access Risk', 'Network Risk', 'Application Risk', 'Infrastructure Risk'],
            content: ['Keyword Risk', 'Backlink Risk', 'Sentiment Risk', 'Brand Risk', 'Compliance Risk']
        };
        
        const titles = cellTitles[category] || ['Risk 1', 'Risk 2', 'Risk 3', 'Risk 4', 'Risk 5'];
        const values = Object.values(subCategories).slice(0, 5);
        
        return values.map((value, index) => {
            const level = this.getRiskLevelFromScore(value);
            return `<div class="heatmap-cell heatmap-${level}" title="${titles[index]}: ${level}"></div>`;
        }).join('');
    }
    
    static getRiskLevelFromScore(score) {
        if (score <= 25) return 'low';
        if (score <= 50) return 'medium';
        if (score <= 75) return 'high';
        return 'critical';
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskIndicatorsUITemplate;
}
