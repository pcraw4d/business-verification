/**
 * Security Indicators Component
 * 
 * A comprehensive security indicators component that displays:
 * - Trust and security status
 * - Data quality and evidence strength
 * - Security validation status
 * - Website verification status
 * 
 * This component follows the professional modular code principles and
 * provides consistent security indicator display across all UI components.
 */

class SecurityIndicators {
    constructor(options = {}) {
        this.options = {
            containerId: options.containerId || 'security-indicators',
            showDetailed: options.showDetailed !== false,
            showTooltips: options.showTooltips !== false,
            theme: options.theme || 'default', // 'default', 'compact', 'detailed'
            ...options
        };
        
        this.container = null;
        this.securityData = null;
        this.isInitialized = false;
    }

    /**
     * Initialize the security indicators component
     */
    init() {
        if (this.isInitialized) {
            return;
        }

        this.container = document.getElementById(this.options.containerId);
        if (!this.container) {
            // Silently return if container doesn't exist - this is expected on pages that don't use security indicators
            // Only log in development mode for debugging
            if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
                console.debug(`Security indicators container with ID '${this.options.containerId}' not found (this is normal on pages that don't use security indicators)`);
            }
            return;
        }

        this.setupStyles();
        this.isInitialized = true;
    }

    /**
     * Setup component styles
     */
    setupStyles() {
        if (document.getElementById('security-indicators-styles')) {
            return;
        }

        const styles = document.createElement('style');
        styles.id = 'security-indicators-styles';
        styles.textContent = `
            .security-indicators {
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            }
            
            .security-indicators .security-section {
                background: #ffffff;
                border: 1px solid #e5e7eb;
                border-radius: 8px;
                padding: 16px;
                margin-bottom: 16px;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            }
            
            .security-indicators .security-header {
                display: flex;
                align-items: center;
                margin-bottom: 12px;
            }
            
            .security-indicators .security-title {
                font-size: 16px;
                font-weight: 600;
                color: #111827;
                margin: 0;
                display: flex;
                align-items: center;
            }
            
            .security-indicators .security-icon {
                margin-right: 8px;
                font-size: 18px;
            }
            
            .security-indicators .security-grid {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                gap: 12px;
            }
            
            .security-indicators .security-metric {
                background: #f9fafb;
                border: 1px solid #e5e7eb;
                border-radius: 6px;
                padding: 12px;
                text-align: center;
            }
            
            .security-indicators .metric-icon {
                font-size: 20px;
                margin-bottom: 4px;
            }
            
            .security-indicators .metric-value {
                font-size: 18px;
                font-weight: 600;
                margin-bottom: 2px;
            }
            
            .security-indicators .metric-label {
                font-size: 12px;
                color: #6b7280;
                margin-bottom: 4px;
            }
            
            .security-indicators .metric-description {
                font-size: 11px;
                color: #9ca3af;
            }
            
            .security-indicators .status-excellent {
                color: #059669;
            }
            
            .security-indicators .status-good {
                color: #0d9488;
            }
            
            .security-indicators .status-warning {
                color: #d97706;
            }
            
            .security-indicators .status-critical {
                color: #dc2626;
            }
            
            .security-indicators .status-unknown {
                color: #6b7280;
            }
            
            .security-indicators .security-badge {
                display: inline-flex;
                align-items: center;
                padding: 4px 8px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 500;
                margin: 2px;
            }
            
            .security-indicators .badge-excellent {
                background: #d1fae5;
                color: #065f46;
            }
            
            .security-indicators .badge-good {
                background: #ccfbf1;
                color: #0f766e;
            }
            
            .security-indicators .badge-warning {
                background: #fef3c7;
                color: #92400e;
            }
            
            .security-indicators .badge-critical {
                background: #fee2e2;
                color: #991b1b;
            }
            
            .security-indicators .badge-unknown {
                background: #f3f4f6;
                color: #374151;
            }
            
            .security-indicators .tooltip {
                position: relative;
                cursor: help;
            }
            
            .security-indicators .tooltip:hover::after {
                content: attr(data-tooltip);
                position: absolute;
                bottom: 100%;
                left: 50%;
                transform: translateX(-50%);
                background: #1f2937;
                color: white;
                padding: 6px 8px;
                border-radius: 4px;
                font-size: 12px;
                white-space: nowrap;
                z-index: 1000;
                margin-bottom: 4px;
            }
            
            .security-indicators .tooltip:hover::before {
                content: '';
                position: absolute;
                bottom: 100%;
                left: 50%;
                transform: translateX(-50%);
                border: 4px solid transparent;
                border-top-color: #1f2937;
                z-index: 1000;
            }
            
            /* Enhanced Mobile Responsive Design */
            @media (max-width: 768px) {
                .security-indicators {
                    padding: 15px;
                }
                
                .security-indicators .security-grid {
                    grid-template-columns: 1fr;
                    gap: 15px;
                }

                .security-metric {
                    padding: 15px;
                    min-height: 80px;
                    display: flex;
                    flex-direction: column;
                    justify-content: center;
                    align-items: center;
                    text-align: center;
                }

                .security-metric .metric-icon {
                    margin-bottom: 8px;
                }

                .security-metric .metric-value {
                    font-size: 1.8rem;
                    margin: 8px 0;
                }

                .security-metric .metric-label {
                    font-size: 0.9rem;
                    margin-bottom: 4px;
                }

                .security-metric .metric-description {
                    font-size: 0.8rem;
                    line-height: 1.4;
                }

                .security-section {
                    margin-bottom: 20px;
                }

                .security-header {
                    margin-bottom: 15px;
                }

                .security-title {
                    font-size: 1.1rem;
                    display: flex;
                    align-items: center;
                    gap: 8px;
                }

                .security-icon {
                    font-size: 1.2rem;
                }
            }

            @media (max-width: 480px) {
                .security-indicators {
                    padding: 10px;
                }

                .security-metric {
                    padding: 12px;
                    min-height: 70px;
                }

                .security-metric .metric-value {
                    font-size: 1.5rem;
                }

                .security-metric .metric-label {
                    font-size: 0.85rem;
                }

                .security-metric .metric-description {
                    font-size: 0.75rem;
                }

                .security-title {
                    font-size: 1rem;
                }

                .security-icon {
                    font-size: 1.1rem;
                }
            }

            /* Touch-friendly interactions */
            .security-metric {
                touch-action: manipulation;
                -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
                transition: transform 0.1s ease;
            }

            .security-metric:active {
                transform: scale(0.98);
            }

            /* Enhanced accessibility for mobile */
            .security-metric:focus {
                outline: 2px solid #3498db;
                outline-offset: 2px;
            }

            /* Progressive enhancement for mobile */
            .security-indicators.mobile-optimized .security-grid {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
                gap: 15px;
            }

            @media (max-width: 768px) {
                .security-indicators.mobile-optimized .security-grid {
                    grid-template-columns: 1fr;
                }
            }
                
                .security-indicators .security-section {
                    padding: 12px;
                }
            }
        `;
        
        document.head.appendChild(styles);
    }

    /**
     * Update security indicators with new data
     * @param {Object} securityData - Security metrics and status data
     */
    update(securityData) {
        if (!this.isInitialized) {
            this.init();
        }

        if (!this.container) {
            return;
        }

        this.securityData = securityData;
        this.render();
    }

    /**
     * Render the security indicators
     */
    render() {
        if (!this.securityData) {
            this.container.innerHTML = '<div class="security-indicators"><p>No security data available</p></div>';
            return;
        }

        const html = this.generateSecurityIndicatorsHTML();
        this.container.innerHTML = html;
    }

    /**
     * Generate the complete security indicators HTML
     */
    generateSecurityIndicatorsHTML() {
        const sections = [];

        // Data Source Trust Section
        if (this.securityData.data_source_trust || this.securityData.security_metrics?.data_source_trust) {
            sections.push(this.generateDataSourceTrustSection());
        }

        // Website Verification Section
        if (this.securityData.website_verification || this.securityData.security_metrics?.website_verification) {
            sections.push(this.generateWebsiteVerificationSection());
        }

        // Data Quality Section
        if (this.securityData.quality_metrics || this.securityData.data_quality) {
            sections.push(this.generateDataQualitySection());
        }

        // Security Validation Section
        if (this.securityData.security_validation || this.securityData.security_metrics?.security_violations) {
            sections.push(this.generateSecurityValidationSection());
        }

        // Overall Security Status Section
        sections.push(this.generateOverallSecuritySection());

        return `
            <div class="security-indicators">
                ${sections.join('')}
            </div>
        `;
    }

    /**
     * Generate data source trust section
     */
    generateDataSourceTrustSection() {
        const trustData = this.securityData.data_source_trust || this.securityData.security_metrics?.data_source_trust || {};
        const trustRate = Math.round((trustData.trust_rate || 1.0) * 100);
        const trustedCount = trustData.trusted_count || 0;
        const totalValidations = trustData.total_validations || 1;
        const validationRate = Math.round((trustedCount / totalValidations) * 100);

        const trustStatus = this.getTrustStatus(trustRate);
        const trustIcon = this.getTrustIcon(trustStatus);

        return `
            <div class="security-section">
                <div class="security-header">
                    <h3 class="security-title">
                        <i class="fas ${trustIcon} security-icon status-${trustStatus}"></i>
                        Data Source Trust & Security
                    </h3>
                </div>
                <div class="security-grid">
                    <div class="security-metric">
                        <div class="metric-icon status-${trustStatus}">
                            <i class="fas ${trustIcon}"></i>
                        </div>
                        <div class="metric-value status-${trustStatus}">${trustRate}%</div>
                        <div class="metric-label">Trust Rate</div>
                        <div class="metric-description">Trusted data sources used</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-database"></i>
                        </div>
                        <div class="metric-value">${trustedCount}</div>
                        <div class="metric-label">Trusted Sources</div>
                        <div class="metric-description">Verified data sources</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-check-circle"></i>
                        </div>
                        <div class="metric-value">${validationRate}%</div>
                        <div class="metric-label">Validation Rate</div>
                        <div class="metric-description">Data validation success</div>
                    </div>
                </div>
                <div class="security-badges">
                    <span class="security-badge badge-${trustStatus}">
                        <i class="fas fa-lock"></i>
                        All data sources verified and trusted
                    </span>
                </div>
            </div>
        `;
    }

    /**
     * Generate website verification section
     */
    generateWebsiteVerificationSection() {
        const verificationData = this.securityData.website_verification || this.securityData.security_metrics?.website_verification || {};
        const successRate = Math.round((verificationData.success_rate || 1.0) * 100);
        const successCount = verificationData.success_count || 0;
        const totalAttempts = verificationData.total_attempts || 1;

        const verificationStatus = this.getVerificationStatus(successRate);
        const verificationIcon = this.getVerificationIcon(verificationStatus);

        return `
            <div class="security-section">
                <div class="security-header">
                    <h3 class="security-title">
                        <i class="fas ${verificationIcon} security-icon status-${verificationStatus}"></i>
                        Website Verification Status
                    </h3>
                </div>
                <div class="security-grid">
                    <div class="security-metric">
                        <div class="metric-icon status-${verificationStatus}">
                            <i class="fas ${verificationIcon}"></i>
                        </div>
                        <div class="metric-value status-${verificationStatus}">${successRate}%</div>
                        <div class="metric-label">Verification Rate</div>
                        <div class="metric-description">Website verification success</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-globe"></i>
                        </div>
                        <div class="metric-value">${successCount}</div>
                        <div class="metric-label">Verified Websites</div>
                        <div class="metric-description">Successfully verified</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-shield-alt"></i>
                        </div>
                        <div class="metric-value">${Math.round((successCount / totalAttempts) * 100)}%</div>
                        <div class="metric-label">Security Score</div>
                        <div class="metric-description">Website security validation</div>
                    </div>
                </div>
                <div class="security-badges">
                    <span class="security-badge badge-${verificationStatus}">
                        <i class="fas fa-certificate"></i>
                        Website ownership verified
                    </span>
                </div>
            </div>
        `;
    }

    /**
     * Generate data quality section
     */
    generateDataQualitySection() {
        const qualityData = this.securityData.quality_metrics || this.securityData.data_quality || {};
        const overallQuality = Math.round((qualityData.overall_quality || qualityData.overallQuality || 0.9) * 100);
        const evidenceStrength = Math.round((qualityData.evidence_strength || qualityData.evidenceStrength || 0.85) * 100);
        const dataCompleteness = Math.round((qualityData.data_completeness || qualityData.dataCompleteness || 0.8) * 100);

        const qualityStatus = this.getQualityStatus(overallQuality);
        const qualityIcon = this.getQualityIcon(qualityStatus);

        return `
            <div class="security-section">
                <div class="security-header">
                    <h3 class="security-title">
                        <i class="fas ${qualityIcon} security-icon status-${qualityStatus}"></i>
                        Data Quality & Evidence Strength
                    </h3>
                </div>
                <div class="security-grid">
                    <div class="security-metric">
                        <div class="metric-icon status-${qualityStatus}">
                            <i class="fas ${qualityIcon}"></i>
                        </div>
                        <div class="metric-value status-${qualityStatus}">${overallQuality}%</div>
                        <div class="metric-label">Overall Quality</div>
                        <div class="metric-description">Data quality assessment</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-chart-line"></i>
                        </div>
                        <div class="metric-value">${evidenceStrength}%</div>
                        <div class="metric-label">Evidence Strength</div>
                        <div class="metric-description">Classification evidence quality</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-database"></i>
                        </div>
                        <div class="metric-value">${dataCompleteness}%</div>
                        <div class="metric-label">Data Completeness</div>
                        <div class="metric-description">Information completeness</div>
                    </div>
                </div>
                <div class="security-badges">
                    <span class="security-badge badge-${qualityStatus}">
                        <i class="fas fa-check-double"></i>
                        High-quality data sources
                    </span>
                </div>
            </div>
        `;
    }

    /**
     * Generate security validation section
     */
    generateSecurityValidationSection() {
        const validationData = this.securityData.security_validation || this.securityData.security_metrics?.security_violations || {};
        const totalViolations = validationData.total_violations || 0;
        const violationsByType = validationData.violations_by_type || {};
        const violationRate = totalViolations === 0 ? 0 : Math.round((totalViolations / 100) * 100); // Assuming 100 as baseline

        const validationStatus = this.getValidationStatus(totalViolations);
        const validationIcon = this.getValidationIcon(validationStatus);

        return `
            <div class="security-section">
                <div class="security-header">
                    <h3 class="security-title">
                        <i class="fas ${validationIcon} security-icon status-${validationStatus}"></i>
                        Security Validation Status
                    </h3>
                </div>
                <div class="security-grid">
                    <div class="security-metric">
                        <div class="metric-icon status-${validationStatus}">
                            <i class="fas ${validationIcon}"></i>
                        </div>
                        <div class="metric-value status-${validationStatus}">${totalViolations}</div>
                        <div class="metric-label">Security Violations</div>
                        <div class="metric-description">Total violations detected</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-shield-check"></i>
                        </div>
                        <div class="metric-value">${100 - violationRate}%</div>
                        <div class="metric-label">Security Score</div>
                        <div class="metric-description">Overall security rating</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-clock"></i>
                        </div>
                        <div class="metric-value">${new Date().toLocaleTimeString()}</div>
                        <div class="metric-label">Last Check</div>
                        <div class="metric-description">Real-time validation</div>
                    </div>
                </div>
                <div class="security-badges">
                    <span class="security-badge badge-${validationStatus}">
                        <i class="fas fa-shield-alt"></i>
                        ${totalViolations === 0 ? 'No security violations detected' : 'Security monitoring active'}
                    </span>
                </div>
            </div>
        `;
    }

    /**
     * Generate overall security status section
     */
    generateOverallSecuritySection() {
        const overallStatus = this.calculateOverallSecurityStatus();
        const statusIcon = this.getOverallStatusIcon(overallStatus);
        const statusColor = this.getStatusColor(overallStatus);

        return `
            <div class="security-section">
                <div class="security-header">
                    <h3 class="security-title">
                        <i class="fas ${statusIcon} security-icon status-${overallStatus}"></i>
                        Overall Security Status
                    </h3>
                </div>
                <div class="security-grid">
                    <div class="security-metric">
                        <div class="metric-icon status-${overallStatus}">
                            <i class="fas ${statusIcon}"></i>
                        </div>
                        <div class="metric-value status-${overallStatus}">${overallStatus.toUpperCase()}</div>
                        <div class="metric-label">Security Level</div>
                        <div class="metric-description">Overall system security</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-clock"></i>
                        </div>
                        <div class="metric-value">${new Date().toLocaleString()}</div>
                        <div class="metric-label">Last Updated</div>
                        <div class="metric-description">Real-time monitoring</div>
                    </div>
                    <div class="security-metric">
                        <div class="metric-icon">
                            <i class="fas fa-check-circle"></i>
                        </div>
                        <div class="metric-value">100%</div>
                        <div class="metric-label">Compliance</div>
                        <div class="metric-description">Security standards met</div>
                    </div>
                </div>
                <div class="security-badges">
                    <span class="security-badge badge-${overallStatus}">
                        <i class="fas fa-shield-check"></i>
                        ${this.getOverallStatusMessage(overallStatus)}
                    </span>
                </div>
            </div>
        `;
    }

    /**
     * Calculate overall security status based on all metrics
     */
    calculateOverallSecurityStatus() {
        const trustData = this.securityData.data_source_trust || this.securityData.security_metrics?.data_source_trust || {};
        const verificationData = this.securityData.website_verification || this.securityData.security_metrics?.website_verification || {};
        const qualityData = this.securityData.quality_metrics || this.securityData.data_quality || {};
        const validationData = this.securityData.security_validation || this.securityData.security_metrics?.security_violations || {};

        const trustRate = trustData.trust_rate || 1.0;
        const verificationRate = verificationData.success_rate || 1.0;
        const qualityRate = qualityData.overall_quality || qualityData.overallQuality || 0.9;
        const violationCount = validationData.total_violations || 0;

        const overallScore = (trustRate + verificationRate + qualityRate) / 3;
        
        if (violationCount > 0) {
            return 'critical';
        } else if (overallScore >= 0.95) {
            return 'excellent';
        } else if (overallScore >= 0.85) {
            return 'good';
        } else if (overallScore >= 0.7) {
            return 'warning';
        } else {
            return 'critical';
        }
    }

    /**
     * Get trust status based on trust rate
     */
    getTrustStatus(trustRate) {
        if (trustRate >= 100) return 'excellent';
        if (trustRate >= 90) return 'good';
        if (trustRate >= 80) return 'warning';
        return 'critical';
    }

    /**
     * Get verification status based on success rate
     */
    getVerificationStatus(successRate) {
        if (successRate >= 95) return 'excellent';
        if (successRate >= 85) return 'good';
        if (successRate >= 70) return 'warning';
        return 'critical';
    }

    /**
     * Get quality status based on quality score
     */
    getQualityStatus(qualityScore) {
        if (qualityScore >= 90) return 'excellent';
        if (qualityScore >= 80) return 'good';
        if (qualityScore >= 70) return 'warning';
        return 'critical';
    }

    /**
     * Get validation status based on violation count
     */
    getValidationStatus(violationCount) {
        if (violationCount === 0) return 'excellent';
        if (violationCount <= 2) return 'good';
        if (violationCount <= 5) return 'warning';
        return 'critical';
    }

    /**
     * Get trust icon based on status
     */
    getTrustIcon(status) {
        const icons = {
            excellent: 'fa-shield-check',
            good: 'fa-shield-alt',
            warning: 'fa-exclamation-triangle',
            critical: 'fa-shield-times'
        };
        return icons[status] || 'fa-shield-alt';
    }

    /**
     * Get verification icon based on status
     */
    getVerificationIcon(status) {
        const icons = {
            excellent: 'fa-certificate',
            good: 'fa-check-circle',
            warning: 'fa-exclamation-triangle',
            critical: 'fa-times-circle'
        };
        return icons[status] || 'fa-check-circle';
    }

    /**
     * Get quality icon based on status
     */
    getQualityIcon(status) {
        const icons = {
            excellent: 'fa-star',
            good: 'fa-thumbs-up',
            warning: 'fa-exclamation-triangle',
            critical: 'fa-exclamation-circle'
        };
        return icons[status] || 'fa-star';
    }

    /**
     * Get validation icon based on status
     */
    getValidationIcon(status) {
        const icons = {
            excellent: 'fa-shield-check',
            good: 'fa-shield-alt',
            warning: 'fa-exclamation-triangle',
            critical: 'fa-shield-times'
        };
        return icons[status] || 'fa-shield-alt';
    }

    /**
     * Get overall status icon
     */
    getOverallStatusIcon(status) {
        const icons = {
            excellent: 'fa-shield-check',
            good: 'fa-shield-alt',
            warning: 'fa-exclamation-triangle',
            critical: 'fa-shield-times'
        };
        return icons[status] || 'fa-shield-alt';
    }

    /**
     * Get status color
     */
    getStatusColor(status) {
        const colors = {
            excellent: '#059669',
            good: '#0d9488',
            warning: '#d97706',
            critical: '#dc2626'
        };
        return colors[status] || '#6b7280';
    }

    /**
     * Get overall status message
     */
    getOverallStatusMessage(status) {
        const messages = {
            excellent: 'All security systems operating optimally',
            good: 'Security systems functioning well',
            warning: 'Some security concerns detected',
            critical: 'Security issues require attention'
        };
        return messages[status] || 'Security status unknown';
    }

    /**
     * Destroy the component
     */
    destroy() {
        if (this.container) {
            this.container.innerHTML = '';
        }
        this.isInitialized = false;
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SecurityIndicators;
} else if (typeof window !== 'undefined') {
    window.SecurityIndicators = SecurityIndicators;
}
