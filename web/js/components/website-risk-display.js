/**
 * Website Risk Display Component
 * 
 * Displays website risk findings including:
 * - Risky keywords detected in website content
 * - Sentiment analysis results
 * - Backlink analysis
 * - Content quality assessment
 * - Compliance issues
 */

class WebsiteRiskDisplay {
    
    /**
     * Render risky keywords section
     * @param {Array} keywords - Array of risky keywords
     * @returns {string} HTML string for risky keywords
     */
    static renderRiskyKeywords(keywords) {
        if (!keywords || keywords.length === 0) {
            return `
                <div class="text-center text-gray-500 py-4">
                    <i class="fas fa-check-circle text-green-500 text-2xl mb-2"></i>
                    <p>No risky keywords detected</p>
                </div>
            `;
        }
        
        // Sort by severity
        const sortedKeywords = keywords.sort((a, b) => {
            const severityOrder = { critical: 4, high: 3, medium: 2, low: 1 };
            return severityOrder[b.severity] - severityOrder[a.severity];
        });
        
        return `
            <div class="space-y-3">
                ${sortedKeywords.map(keyword => `
                    <div class="risky-keyword-badge severity-${keyword.severity} p-3 rounded-lg border-l-4">
                        <div class="flex items-start justify-between">
                            <div class="flex-1">
                                <div class="flex items-center mb-2">
                                    <i class="fas fa-exclamation-triangle mr-2 text-${this.getSeverityColor(keyword.severity)}"></i>
                                    <span class="font-medium text-gray-900">${keyword.keyword}</span>
                                    <span class="ml-2 badge-${keyword.severity} px-2 py-1 rounded text-xs font-bold">
                                        ${keyword.severity.toUpperCase()}
                                    </span>
                                </div>
                                <p class="text-sm text-gray-600 mb-2">${keyword.context || 'Found in website content'}</p>
                                <div class="flex items-center text-xs text-gray-500 space-x-4">
                                    <span>Category: ${keyword.category || 'General'}</span>
                                    <span>Confidence: ${Math.round((keyword.confidence || 0.8) * 100)}%</span>
                                    ${keyword.detectedAt ? `<span>Detected: ${new Date(keyword.detectedAt).toLocaleDateString()}</span>` : ''}
                                </div>
                            </div>
                            <div class="flex space-x-2 ml-4">
                                <button class="btn btn-sm btn-outline" onclick="WebsiteRiskDisplay.acknowledgeKeyword('${keyword.keyword}')">
                                    <i class="fas fa-check mr-1"></i>
                                    Acknowledge
                                </button>
                                <button class="btn btn-sm btn-primary" onclick="WebsiteRiskDisplay.investigateKeyword('${keyword.keyword}')">
                                    <i class="fas fa-search mr-1"></i>
                                    Investigate
                                </button>
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }
    
    /**
     * Render sentiment analysis section
     * @param {Object} sentiment - Sentiment analysis data
     * @returns {string} HTML string for sentiment analysis
     */
    static renderSentimentAnalysis(sentiment) {
        if (!sentiment) {
            return `
                <div class="text-center text-gray-500 py-4">
                    <i class="fas fa-info-circle text-blue-500 text-2xl mb-2"></i>
                    <p>Sentiment analysis not available</p>
                </div>
            `;
        }
        
        const overallSentiment = sentiment.overall_sentiment || 'neutral';
        const positivePct = sentiment.positive_pct || 0;
        const neutralPct = sentiment.neutral_pct || 0;
        const negativePct = sentiment.negative_pct || 0;
        
        return `
            <div class="sentiment-analysis">
                <!-- Overall Sentiment -->
                <div class="mb-4">
                    <div class="flex items-center justify-between mb-2">
                        <h4 class="font-medium text-gray-900">Overall Sentiment</h4>
                        <span class="sentiment-badge sentiment-${overallSentiment} px-3 py-1 rounded-full text-sm font-bold">
                            <i class="fas fa-${this.getSentimentIcon(overallSentiment)} mr-1"></i>
                            ${overallSentiment.toUpperCase()}
                        </span>
                    </div>
                    <p class="text-sm text-gray-600">${sentiment.summary || 'Sentiment analysis based on online mentions and content'}</p>
                </div>
                
                <!-- Sentiment Breakdown -->
                <div class="mb-4">
                    <h4 class="font-medium text-gray-900 mb-2">Sentiment Breakdown</h4>
                    <div class="sentiment-breakdown">
                        <div class="sentiment-bar mb-2">
                            <div class="positive" style="width: ${positivePct}%" title="Positive: ${positivePct}%"></div>
                            <div class="neutral" style="width: ${neutralPct}%" title="Neutral: ${neutralPct}%"></div>
                            <div class="negative" style="width: ${negativePct}%" title="Negative: ${negativePct}%"></div>
                        </div>
                        <div class="sentiment-labels text-xs flex justify-between">
                            <span class="positive text-green-600">${positivePct}% Positive</span>
                            <span class="neutral text-gray-600">${neutralPct}% Neutral</span>
                            <span class="negative text-red-600">${negativePct}% Negative</span>
                        </div>
                    </div>
                </div>
                
                <!-- Key Insights -->
                ${sentiment.key_insights ? `
                    <div class="mb-4">
                        <h4 class="font-medium text-gray-900 mb-2">Key Insights</h4>
                        <ul class="text-sm text-gray-600 space-y-1">
                            ${sentiment.key_insights.map(insight => `
                                <li class="flex items-start">
                                    <i class="fas fa-chevron-right text-blue-500 mr-2 mt-1"></i>
                                    ${insight}
                                </li>
                            `).join('')}
                        </ul>
                    </div>
                ` : ''}
                
                <!-- Sentiment Trends -->
                ${sentiment.trends ? `
                    <div>
                        <h4 class="font-medium text-gray-900 mb-2">Sentiment Trends</h4>
                        <div class="space-y-2">
                            ${sentiment.trends.map(trend => `
                                <div class="flex items-center justify-between text-sm">
                                    <span class="text-gray-600">${trend.period}</span>
                                    <div class="flex items-center">
                                        <span class="sentiment-badge sentiment-${trend.sentiment} px-2 py-1 rounded text-xs mr-2">
                                            ${trend.sentiment}
                                        </span>
                                        <span class="text-gray-900">${trend.score}%</span>
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                ` : ''}
            </div>
        `;
    }
    
    /**
     * Render backlink analysis section
     * @param {Object} backlinks - Backlink analysis data
     * @returns {string} HTML string for backlink analysis
     */
    static renderBacklinkAnalysis(backlinks) {
        if (!backlinks) {
            return `
                <div class="text-center text-gray-500 py-4">
                    <i class="fas fa-info-circle text-blue-500 text-2xl mb-2"></i>
                    <p>Backlink analysis not available</p>
                </div>
            `;
        }
        
        const qualityScore = backlinks.quality_score || 0;
        const totalBacklinks = backlinks.total_backlinks || 0;
        const toxicBacklinks = backlinks.toxic_backlinks || 0;
        const qualityLevel = this.getQualityLevel(qualityScore);
        
        return `
            <div class="backlink-analysis">
                <!-- Quality Score -->
                <div class="mb-4">
                    <div class="flex items-center justify-between mb-2">
                        <h4 class="font-medium text-gray-900">Backlink Quality Score</h4>
                        <span class="quality-badge quality-${qualityLevel} px-3 py-1 rounded-full text-sm font-bold">
                            ${qualityScore}/100
                        </span>
                    </div>
                    <div class="quality-progress">
                        <div class="quality-progress-fill quality-${qualityLevel}" style="width: ${qualityScore}%"></div>
                    </div>
                </div>
                
                <!-- Backlink Statistics -->
                <div class="grid grid-cols-2 gap-4 mb-4">
                    <div class="stat-card">
                        <div class="stat-value text-2xl font-bold text-blue-600">${totalBacklinks}</div>
                        <div class="stat-label text-sm text-gray-600">Total Backlinks</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value text-2xl font-bold text-red-600">${toxicBacklinks}</div>
                        <div class="stat-label text-sm text-gray-600">Toxic Backlinks</div>
                    </div>
                </div>
                
                <!-- Risk Assessment -->
                <div class="mb-4">
                    <h4 class="font-medium text-gray-900 mb-2">Risk Assessment</h4>
                    <div class="space-y-2">
                        ${this.getBacklinkRiskItems(backlinks).map(item => `
                            <div class="flex items-center justify-between text-sm">
                                <span class="text-gray-600">${item.label}</span>
                                <span class="risk-badge risk-${item.level} px-2 py-1 rounded text-xs font-bold">
                                    ${item.level.toUpperCase()}
                                </span>
                            </div>
                        `).join('')}
                    </div>
                </div>
                
                <!-- Recommendations -->
                ${backlinks.recommendations ? `
                    <div>
                        <h4 class="font-medium text-gray-900 mb-2">Recommendations</h4>
                        <ul class="text-sm text-gray-600 space-y-1">
                            ${backlinks.recommendations.map(rec => `
                                <li class="flex items-start">
                                    <i class="fas fa-chevron-right text-blue-500 mr-2 mt-1"></i>
                                    ${rec}
                                </li>
                            `).join('')}
                        </ul>
                    </div>
                ` : ''}
            </div>
        `;
    }
    
    /**
     * Render content quality assessment
     * @param {Object} contentQuality - Content quality data
     * @returns {string} HTML string for content quality
     */
    static renderContentQuality(contentQuality) {
        if (!contentQuality) {
            return `
                <div class="text-center text-gray-500 py-4">
                    <i class="fas fa-info-circle text-blue-500 text-2xl mb-2"></i>
                    <p>Content quality assessment not available</p>
                </div>
            `;
        }
        
        const score = contentQuality.score || 0;
        const level = this.getQualityLevel(score);
        
        return `
            <div class="content-quality">
                <!-- Overall Score -->
                <div class="mb-4">
                    <div class="flex items-center justify-between mb-2">
                        <h4 class="font-medium text-gray-900">Content Quality Score</h4>
                        <span class="quality-badge quality-${level} px-3 py-1 rounded-full text-sm font-bold">
                            ${score}/100
                        </span>
                    </div>
                    <div class="quality-progress">
                        <div class="quality-progress-fill quality-${level}" style="width: ${score}%"></div>
                    </div>
                </div>
                
                <!-- Quality Factors -->
                <div class="mb-4">
                    <h4 class="font-medium text-gray-900 mb-2">Quality Factors</h4>
                    <div class="space-y-2">
                        ${this.getContentQualityFactors(contentQuality).map(factor => `
                            <div class="flex items-center justify-between text-sm">
                                <span class="text-gray-600">${factor.label}</span>
                                <div class="flex items-center">
                                    <div class="w-16 bg-gray-200 rounded-full h-2 mr-2">
                                        <div class="quality-progress-fill quality-${factor.level}" style="width: ${factor.score}%"></div>
                                    </div>
                                    <span class="text-gray-900">${factor.score}%</span>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
                
                <!-- Issues -->
                ${contentQuality.issues && contentQuality.issues.length > 0 ? `
                    <div>
                        <h4 class="font-medium text-gray-900 mb-2">Content Issues</h4>
                        <ul class="text-sm text-gray-600 space-y-1">
                            ${contentQuality.issues.map(issue => `
                                <li class="flex items-start">
                                    <i class="fas fa-exclamation-triangle text-yellow-500 mr-2 mt-1"></i>
                                    ${issue}
                                </li>
                            `).join('')}
                        </ul>
                    </div>
                ` : ''}
            </div>
        `;
    }
    
    /**
     * Render compliance issues section
     * @param {Array} issues - Array of compliance issues
     * @returns {string} HTML string for compliance issues
     */
    static renderComplianceIssues(issues) {
        if (!issues || issues.length === 0) {
            return `
                <div class="text-center text-gray-500 py-4">
                    <i class="fas fa-check-circle text-green-500 text-2xl mb-2"></i>
                    <p>No compliance issues detected</p>
                </div>
            `;
        }
        
        return `
            <div class="space-y-3">
                ${issues.map(issue => `
                    <div class="compliance-issue p-3 rounded-lg border-l-4 border-red-500 bg-red-50">
                        <div class="flex items-start justify-between">
                            <div class="flex-1">
                                <div class="flex items-center mb-2">
                                    <i class="fas fa-gavel mr-2 text-red-500"></i>
                                    <span class="font-medium text-gray-900">${issue.title || 'Compliance Issue'}</span>
                                    <span class="ml-2 compliance-badge px-2 py-1 rounded text-xs font-bold bg-red-100 text-red-800">
                                        ${issue.severity || 'HIGH'}
                                    </span>
                                </div>
                                <p class="text-sm text-gray-600 mb-2">${issue.description || 'Compliance issue detected'}</p>
                                <div class="flex items-center text-xs text-gray-500 space-x-4">
                                    <span>Category: ${issue.category || 'General'}</span>
                                    <span>Regulation: ${issue.regulation || 'N/A'}</span>
                                    ${issue.detectedAt ? `<span>Detected: ${new Date(issue.detectedAt).toLocaleDateString()}</span>` : ''}
                                </div>
                            </div>
                            <div class="flex space-x-2 ml-4">
                                <button class="btn btn-sm btn-outline" onclick="WebsiteRiskDisplay.acknowledgeComplianceIssue('${issue.id}')">
                                    <i class="fas fa-check mr-1"></i>
                                    Acknowledge
                                </button>
                                <button class="btn btn-sm btn-primary" onclick="WebsiteRiskDisplay.investigateComplianceIssue('${issue.id}')">
                                    <i class="fas fa-search mr-1"></i>
                                    Investigate
                                </button>
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }
    
    // Helper methods
    
    /**
     * Get severity color for CSS classes
     * @param {string} severity - Severity level
     * @returns {string} Color class
     */
    static getSeverityColor(severity) {
        const colors = {
            critical: 'red',
            high: 'orange',
            medium: 'yellow',
            low: 'green'
        };
        return colors[severity] || 'gray';
    }
    
    /**
     * Get sentiment icon
     * @param {string} sentiment - Sentiment type
     * @returns {string} Icon class
     */
    static getSentimentIcon(sentiment) {
        const icons = {
            positive: 'smile',
            neutral: 'meh',
            negative: 'frown'
        };
        return icons[sentiment] || 'meh';
    }
    
    /**
     * Get quality level from score
     * @param {number} score - Quality score
     * @returns {string} Quality level
     */
    static getQualityLevel(score) {
        if (score >= 80) return 'excellent';
        if (score >= 60) return 'good';
        if (score >= 40) return 'fair';
        if (score >= 20) return 'poor';
        return 'critical';
    }
    
    /**
     * Get backlink risk items
     * @param {Object} backlinks - Backlink data
     * @returns {Array} Risk items
     */
    static getBacklinkRiskItems(backlinks) {
        const items = [];
        
        if (backlinks.toxic_backlinks > 0) {
            items.push({
                label: 'Toxic Backlinks',
                level: backlinks.toxic_backlinks > 10 ? 'critical' : 'high'
            });
        }
        
        if (backlinks.quality_score < 30) {
            items.push({
                label: 'Low Quality Score',
                level: 'high'
            });
        }
        
        if (backlinks.spam_score > 50) {
            items.push({
                label: 'High Spam Score',
                level: 'critical'
            });
        }
        
        return items;
    }
    
    /**
     * Get content quality factors
     * @param {Object} contentQuality - Content quality data
     * @returns {Array} Quality factors
     */
    static getContentQualityFactors(contentQuality) {
        return [
            {
                label: 'Readability',
                score: contentQuality.readability_score || 0,
                level: this.getQualityLevel(contentQuality.readability_score || 0)
            },
            {
                label: 'SEO Optimization',
                score: contentQuality.seo_score || 0,
                level: this.getQualityLevel(contentQuality.seo_score || 0)
            },
            {
                label: 'Content Freshness',
                score: contentQuality.freshness_score || 0,
                level: this.getQualityLevel(contentQuality.freshness_score || 0)
            },
            {
                label: 'Technical Quality',
                score: contentQuality.technical_score || 0,
                level: this.getQualityLevel(contentQuality.technical_score || 0)
            }
        ];
    }
    
    // Event handlers (global functions)
    
    /**
     * Acknowledge risky keyword
     * @param {string} keyword - Keyword to acknowledge
     */
    static acknowledgeKeyword(keyword) {
        console.log(`Acknowledging risky keyword: ${keyword}`);
        // TODO: Implement keyword acknowledgment
        this.showToast(`Keyword "${keyword}" acknowledged`, 'success');
    }
    
    /**
     * Investigate risky keyword
     * @param {string} keyword - Keyword to investigate
     */
    static investigateKeyword(keyword) {
        console.log(`Investigating risky keyword: ${keyword}`);
        // TODO: Implement keyword investigation
        this.showToast(`Investigation started for "${keyword}"`, 'info');
    }
    
    /**
     * Acknowledge compliance issue
     * @param {string} issueId - Issue ID to acknowledge
     */
    static acknowledgeComplianceIssue(issueId) {
        console.log(`Acknowledging compliance issue: ${issueId}`);
        // TODO: Implement compliance issue acknowledgment
        this.showToast('Compliance issue acknowledged', 'success');
    }
    
    /**
     * Investigate compliance issue
     * @param {string} issueId - Issue ID to investigate
     */
    static investigateComplianceIssue(issueId) {
        console.log(`Investigating compliance issue: ${issueId}`);
        // TODO: Implement compliance issue investigation
        this.showToast('Compliance investigation started', 'info');
    }
    
    /**
     * Show toast notification
     * @param {string} message - Toast message
     * @param {string} type - Toast type
     */
    static showToast(message, type = 'info') {
        // Create toast element
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.innerHTML = `
            <div class="flex items-center">
                <i class="fas fa-${this.getToastIcon(type)} mr-2"></i>
                <span>${message}</span>
            </div>
        `;
        toast.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${this.getToastColor(type)};
            color: white;
            padding: 12px 16px;
            border-radius: 6px;
            z-index: 10000;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
            transform: translateX(100%);
            transition: transform 0.3s ease;
        `;
        
        document.body.appendChild(toast);
        
        // Animate in
        setTimeout(() => {
            toast.style.transform = 'translateX(0)';
        }, 100);
        
        // Remove after 3 seconds
        setTimeout(() => {
            toast.style.transform = 'translateX(100%)';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        }, 3000);
    }
    
    /**
     * Get toast icon for type
     * @param {string} type - Toast type
     * @returns {string} Icon class
     */
    static getToastIcon(type) {
        const icons = {
            success: 'check-circle',
            error: 'exclamation-circle',
            info: 'info-circle',
            warning: 'exclamation-triangle'
        };
        return icons[type] || 'info-circle';
    }
    
    /**
     * Get toast color for type
     * @param {string} type - Toast type
     * @returns {string} Color
     */
    static getToastColor(type) {
        const colors = {
            success: '#10b981',
            error: '#ef4444',
            info: '#3b82f6',
            warning: '#f59e0b'
        };
        return colors[type] || '#3b82f6';
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = WebsiteRiskDisplay;
}
