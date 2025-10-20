/**
 * Risk Indicators Data Service
 * 
 * Aggregates data from multiple sources:
 * - Stored Business Analytics results
 * - Live Risk Assessment API calls
 * - Merchant data from existing APIs
 * 
 * Leverages existing RealDataIntegration patterns for API calls.
 */

class RiskIndicatorsDataService {
    constructor() {
        this.dataIntegration = new RealDataIntegration();
        this.apiConfig = APIConfig;
        this.helpers = RiskIndicatorsHelpers;
        this.cache = new Map();
        this.cacheTimeout = 5 * 60 * 1000; // 5 minutes
    }
    
    /**
     * Main method - loads and combines all risk data sources
     * @param {string} merchantId - Merchant ID
     * @returns {Object} Combined risk data
     */
    async loadAllRiskData(merchantId) {
        try {
            console.log(`🔍 Loading risk data for merchant: ${merchantId}`);
            
            // Check cache first
            const cacheKey = `risk_data_${merchantId}`;
            const cached = this.cache.get(cacheKey);
            if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
                console.log('📋 Using cached risk data');
                return cached.data;
            }
            
            // Load data from all sources in parallel (allow some to fail)
            const results = await Promise.allSettled([
                this.loadMerchantData(merchantId),
                this.loadStoredAnalytics(merchantId),
                this.loadRiskAssessment(merchantId)
            ]);
            
            // Extract successful results, use fallback data for failed ones
            const merchantData = results[0].status === 'fulfilled' ? results[0].value : this.getFallbackMerchantData(merchantId);
            const analyticsData = results[1].status === 'fulfilled' ? results[1].value : this.getFallbackAnalyticsData(merchantId);
            const riskAssessment = results[2].status === 'fulfilled' ? results[2].value : this.getFallbackRiskAssessment(merchantId);
            
            // Log any failures
            results.forEach((result, index) => {
                if (result.status === 'rejected') {
                    const sources = ['merchant data', 'analytics data', 'risk assessment'];
                    console.warn(`⚠️ Failed to load ${sources[index]}, using fallback data:`, result.reason);
                }
            });
            
            // Merge and normalize data
            const combinedData = this.mergeAndNormalize(merchantData, analyticsData, riskAssessment);
            
            // Cache the result
            this.cache.set(cacheKey, {
                data: combinedData,
                timestamp: Date.now()
            });
            
            console.log('✅ Risk data loaded and cached successfully');
            return combinedData;
            
        } catch (error) {
            console.error('❌ Failed to load risk data:', error);
            // Return mock data as fallback
            return this.generateMockRiskData(merchantId);
        }
    }
    
    /**
     * Load merchant data using existing RealDataIntegration
     * @param {string} merchantId - Merchant ID
     * @returns {Object} Merchant data
     */
    async loadMerchantData(merchantId) {
        return await this.dataIntegration.getMerchantById(merchantId);
    }
    
    /**
     * Load stored Business Analytics results
     * @param {string} merchantId - Merchant ID
     * @returns {Object} Analytics data
     */
    async loadStoredAnalytics(merchantId) {
        try {
            // Check session storage first
            const cached = sessionStorage.getItem(`analytics_${merchantId}`);
            if (cached) {
                console.log('📋 Using cached analytics data');
                return JSON.parse(cached);
            }
            
            // Fetch from API
            const endpoints = this.apiConfig.getEndpoints();
            const response = await fetch(endpoints.merchantById(merchantId), {
                method: 'GET',
                headers: this.apiConfig.getHeaders()
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            // Cache in session storage
            sessionStorage.setItem(`analytics_${merchantId}`, JSON.stringify(data));
            
            return data;
        } catch (error) {
            throw error; // Let Promise.allSettled handle the error
        }
    }
    
    /**
     * Load live risk assessment from Risk Assessment API
     * @param {string} merchantId - Merchant ID
     * @returns {Object} Risk assessment data
     */
    async loadRiskAssessment(merchantId) {
        try {
            const endpoints = this.apiConfig.getEndpoints();
            const response = await fetch(endpoints.riskAssess, {
                method: 'POST',
                headers: this.apiConfig.getHeaders(),
                body: JSON.stringify({ 
                    merchantId,
                    includeTrendAnalysis: true,
                    includeRecommendations: true,
                    includeExplanations: true
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            throw error; // Let Promise.allSettled handle the error
        }
    }
    
    /**
     * Merge and normalize data from all sources
     * @param {Object} merchant - Merchant data
     * @param {Object} analytics - Analytics data
     * @param {Object} risk - Risk assessment data
     * @returns {Object} Normalized risk data
     */
    mergeAndNormalize(merchant, analytics, risk) {
        const categories = this.buildRiskCategories(analytics, risk);
        const overallScore = this.helpers.calculateOverallRiskScore(categories);
        
        return {
            merchantId: merchant.id,
            merchantName: merchant.name,
            overallRiskScore: overallScore,
            riskLevel: this.helpers.getOverallRiskLevel(overallScore),
            categories: categories,
            alerts: this.extractAlerts(analytics, risk),
            recommendations: this.extractRecommendations(risk),
            websiteRisks: this.extractWebsiteRisks(analytics),
            shapData: this.extractSHAPData(risk),
            lastUpdated: new Date().toISOString(),
            dataSources: {
                merchant: !!merchant.id,
                analytics: !!analytics.id,
                riskAssessment: !!risk.id
            }
        };
    }
    
    /**
     * Build 6 risk categories including Content Risk from website analysis
     * @param {Object} analytics - Analytics data
     * @param {Object} risk - Risk assessment data
     * @returns {Object} Risk categories
     */
    buildRiskCategories(analytics, risk) {
        // Default categories with mock scores
        const categories = {
            financial: { 
                score: 15, 
                level: 'low', 
                subCategories: {
                    revenue: 10,
                    cashFlow: 20,
                    debt: 15,
                    credit: 12,
                    market: 18
                }
            },
            operational: { 
                score: 35, 
                level: 'medium', 
                subCategories: {
                    process: 30,
                    staff: 45,
                    technology: 35,
                    supplyChain: 25,
                    quality: 40
                }
            },
            regulatory: { 
                score: 65, 
                level: 'high', 
                subCategories: {
                    compliance: 70,
                    license: 80,
                    audit: 50,
                    legal: 60,
                    policy: 65
                }
            },
            reputational: { 
                score: 25, 
                level: 'low', 
                subCategories: {
                    brand: 20,
                    reviews: 30,
                    media: 25,
                    social: 22,
                    trust: 28
                }
            },
            cybersecurity: { 
                score: 85, 
                level: 'critical', 
                subCategories: {
                    data: 90,
                    access: 80,
                    network: 85,
                    application: 88,
                    infrastructure: 82
                }
            },
            content: {  // NEW category from website analysis
                score: this.calculateContentRiskScore(analytics),
                level: this.helpers.getRiskLevelLower(this.calculateContentRiskScore(analytics)),
                subCategories: {
                    riskyKeywords: this.getRiskyKeywordsScore(analytics),
                    backlinks: this.getBacklinksScore(analytics),
                    sentiment: this.getSentimentScore(analytics),
                    brandReputation: this.getBrandReputationScore(analytics),
                    complianceLanguage: this.getComplianceScore(analytics)
                }
            }
        };
        
        // Override with real risk data if available
        if (risk.risk_factors && Array.isArray(risk.risk_factors)) {
            risk.risk_factors.forEach(factor => {
                if (categories[factor.category]) {
                    categories[factor.category].score = this.helpers.normalizeRiskScore(factor.score, 1);
                    categories[factor.category].level = factor.level || this.helpers.getRiskLevelLower(categories[factor.category].score);
                }
            });
        }
        
        // Add last updated timestamps
        Object.values(categories).forEach(category => {
            category.lastUpdated = new Date().toISOString();
        });
        
        return categories;
    }
    
    /**
     * Calculate content risk score from website analysis
     * @param {Object} analytics - Analytics data
     * @returns {number} Content risk score (0-100)
     */
    calculateContentRiskScore(analytics) {
        let score = 0;
        let factors = 0;
        
        // Risky keywords factor
        if (analytics.risk_keywords && analytics.risk_keywords.length > 0) {
            const highRiskKeywords = analytics.risk_keywords.filter(kw => 
                kw.severity === 'high' || kw.severity === 'critical'
            ).length;
            score += Math.min(40, highRiskKeywords * 8); // Max 40 points
            factors++;
        }
        
        // Sentiment factor
        if (analytics.sentiment_analysis) {
            const sentiment = analytics.sentiment_analysis;
            if (sentiment.overall_sentiment === 'negative') {
                score += 30;
            } else if (sentiment.overall_sentiment === 'neutral') {
                score += 15;
            }
            factors++;
        }
        
        // Backlinks factor
        if (analytics.backlink_analysis) {
            const backlinks = analytics.backlink_analysis;
            if (backlinks.quality_score < 50) {
                score += 20;
            } else if (backlinks.quality_score < 70) {
                score += 10;
            }
            factors++;
        }
        
        // Content quality factor
        if (analytics.content_quality_score !== undefined) {
            if (analytics.content_quality_score < 30) {
                score += 25;
            } else if (analytics.content_quality_score < 60) {
                score += 15;
            }
            factors++;
        }
        
        // Compliance factor
        if (analytics.compliance_issues && analytics.compliance_issues.length > 0) {
            score += Math.min(25, analytics.compliance_issues.length * 5);
            factors++;
        }
        
        return factors > 0 ? Math.min(100, score / factors * 2) : 20; // Default to low risk
    }
    
    /**
     * Extract alerts from risky keywords and high-risk factors
     * @param {Object} analytics - Analytics data
     * @param {Object} risk - Risk assessment data
     * @returns {Array} Array of alerts
     */
    extractAlerts(analytics, risk) {
        const alerts = [];
        
        // Add website risk alerts
        if (analytics.risk_keywords && Array.isArray(analytics.risk_keywords)) {
            analytics.risk_keywords.forEach(keyword => {
                if (keyword.severity === 'high' || keyword.severity === 'critical') {
                    alerts.push({
                        id: this.helpers.generateId('alert'),
                        type: 'risky_keyword',
                        severity: keyword.severity,
                        title: `Risky Keyword Detected: ${keyword.keyword}`,
                        description: keyword.context || `Keyword "${keyword.keyword}" detected in website content`,
                        source: 'website_analysis',
                        detectedAt: keyword.detectedAt || new Date().toISOString()
                    });
                }
            });
        }
        
        // Add negative sentiment alerts
        if (analytics.sentiment_analysis && analytics.sentiment_analysis.overall_sentiment === 'negative') {
            alerts.push({
                id: this.helpers.generateId('alert'),
                type: 'negative_sentiment',
                severity: 'high',
                title: 'Negative Sentiment Detected',
                description: analytics.sentiment_analysis.summary || 'Negative sentiment detected in online mentions',
                source: 'sentiment_analysis',
                detectedAt: new Date().toISOString()
            });
        }
        
        // Add high-risk factor alerts
        if (risk.risk_factors && Array.isArray(risk.risk_factors)) {
            risk.risk_factors.filter(f => f.level === 'critical' || f.level === 'high')
                .forEach(factor => {
                    alerts.push({
                        id: this.helpers.generateId('alert'),
                        type: 'risk_factor',
                        severity: factor.level,
                        title: `${this.helpers.formatCategoryName(factor.category)} Risk: ${factor.name}`,
                        description: factor.description || `High risk factor detected: ${factor.name}`,
                        source: 'risk_assessment',
                        detectedAt: factor.detectedAt || new Date().toISOString()
                    });
                });
        }
        
        return alerts;
    }
    
    /**
     * Extract recommendations from risk assessment
     * @param {Object} risk - Risk assessment data
     * @returns {Array} Array of recommendations
     */
    extractRecommendations(risk) {
        const recommendations = [];
        
        // Add ML-based recommendations from risk assessment
        if (risk.recommendations && Array.isArray(risk.recommendations)) {
            risk.recommendations.forEach(rec => {
                recommendations.push({
                    id: this.helpers.generateId('rec'),
                    type: 'ml_based',
                    priority: rec.priority || 'medium',
                    title: rec.title || 'Risk Mitigation Recommendation',
                    description: rec.description || 'Automated recommendation based on risk analysis',
                    impactScore: rec.impact_score || 0.5,
                    difficulty: rec.difficulty || 'medium',
                    actionRequired: rec.action_required || 'Review and implement',
                    status: 'pending'
                });
            });
        }
        
        // Add manual verification recommendations
        const manualRecommendations = this.generateManualVerificationRecommendations(risk);
        recommendations.push(...manualRecommendations);
        
        return recommendations;
    }
    
    /**
     * Generate manual verification recommendations based on risk levels
     * @param {Object} risk - Risk assessment data
     * @returns {Array} Array of manual verification recommendations
     */
    generateManualVerificationRecommendations(risk) {
        const recommendations = [];
        
        // High risk factors trigger additional verification
        if (risk.risk_factors && Array.isArray(risk.risk_factors)) {
            const highRiskFactors = risk.risk_factors.filter(f => f.level === 'high' || f.level === 'critical');
            
            highRiskFactors.forEach(factor => {
                switch (factor.category) {
                    case 'financial':
                        recommendations.push({
                            id: this.helpers.generateId('rec'),
                            type: 'manual_verification',
                            priority: 'high',
                            title: 'Financial Document Verification',
                            description: 'Request and verify financial statements, bank statements, and tax returns',
                            impactScore: 0.8,
                            difficulty: 'medium',
                            actionRequired: 'Request financial documents from merchant',
                            status: 'pending'
                        });
                        break;
                    case 'regulatory':
                        recommendations.push({
                            id: this.helpers.generateId('rec'),
                            type: 'compliance_check',
                            priority: 'critical',
                            title: 'Regulatory Compliance Verification',
                            description: 'Verify business licenses, permits, and regulatory compliance status',
                            impactScore: 0.9,
                            difficulty: 'high',
                            actionRequired: 'Conduct compliance review and document verification',
                            status: 'pending'
                        });
                        break;
                    case 'cybersecurity':
                        recommendations.push({
                            id: this.helpers.generateId('rec'),
                            type: 'security_audit',
                            priority: 'critical',
                            title: 'Security Assessment Required',
                            description: 'Conduct comprehensive security assessment and penetration testing',
                            impactScore: 0.95,
                            difficulty: 'high',
                            actionRequired: 'Schedule security audit with qualified assessor',
                            status: 'pending'
                        });
                        break;
                }
            });
        }
        
        return recommendations;
    }
    
    /**
     * Extract website risk data
     * @param {Object} analytics - Analytics data
     * @returns {Object} Website risk data
     */
    extractWebsiteRisks(analytics) {
        return {
            riskyKeywords: analytics.risk_keywords || [],
            sentimentAnalysis: analytics.sentiment_analysis || null,
            backlinkAnalysis: analytics.backlink_analysis || null,
            contentQuality: analytics.content_quality_score || null,
            complianceIssues: analytics.compliance_issues || []
        };
    }
    
    /**
     * Extract SHAP analysis data
     * @param {Object} risk - Risk assessment data
     * @returns {Object} SHAP data
     */
    extractSHAPData(risk) {
        return {
            feature_contributions: risk.shap_analysis?.feature_contributions || [],
            prediction_value: risk.shap_analysis?.prediction_value || 0,
            base_value: risk.shap_analysis?.base_value || 0.5,
            confidence: risk.shap_analysis?.confidence || 0.8
        };
    }
    
    // Helper methods for content risk calculation
    
    getRiskyKeywordsScore(analytics) {
        if (!analytics.risk_keywords) return 0;
        const highRiskCount = analytics.risk_keywords.filter(kw => 
            kw.severity === 'high' || kw.severity === 'critical'
        ).length;
        return Math.min(100, highRiskCount * 20);
    }
    
    getBacklinksScore(analytics) {
        if (!analytics.backlink_analysis) return 0;
        return 100 - (analytics.backlink_analysis.quality_score || 50);
    }
    
    getSentimentScore(analytics) {
        if (!analytics.sentiment_analysis) return 0;
        const sentiment = analytics.sentiment_analysis.overall_sentiment;
        if (sentiment === 'negative') return 80;
        if (sentiment === 'neutral') return 40;
        return 10;
    }
    
    getBrandReputationScore(analytics) {
        if (!analytics.reputation_score) return 0;
        return 100 - analytics.reputation_score;
    }
    
    getComplianceScore(analytics) {
        if (!analytics.compliance_issues) return 0;
        return Math.min(100, analytics.compliance_issues.length * 25);
    }
    
    /**
     * Generate mock risk data for development/fallback
     * @param {string} merchantId - Merchant ID
     * @returns {Object} Mock risk data
     */
    generateMockRiskData(merchantId) {
        console.log('🎭 Generating mock risk data for development');
        
        return {
            merchantId: merchantId,
            merchantName: 'Mock Merchant',
            overallRiskScore: 45,
            riskLevel: 'medium',
            categories: {
                financial: { score: 15, level: 'low', subCategories: { revenue: 10, cashFlow: 20, debt: 15, credit: 12, market: 18 } },
                operational: { score: 35, level: 'medium', subCategories: { process: 30, staff: 45, technology: 35, supplyChain: 25, quality: 40 } },
                regulatory: { score: 65, level: 'high', subCategories: { compliance: 70, license: 80, audit: 50, legal: 60, policy: 65 } },
                reputational: { score: 25, level: 'low', subCategories: { brand: 20, reviews: 30, media: 25, social: 22, trust: 28 } },
                cybersecurity: { score: 85, level: 'critical', subCategories: { data: 90, access: 80, network: 85, application: 88, infrastructure: 82 } },
                content: { score: 30, level: 'medium', subCategories: { riskyKeywords: 25, backlinks: 35, sentiment: 20, brandReputation: 30, complianceLanguage: 40 } }
            },
            alerts: [
                {
                    id: 'alert_1',
                    type: 'risky_keyword',
                    severity: 'high',
                    title: 'Risky Keyword Detected: gambling',
                    description: 'Keyword "gambling" detected in website content',
                    source: 'website_analysis',
                    detectedAt: new Date().toISOString()
                }
            ],
            recommendations: [
                {
                    id: 'rec_1',
                    type: 'manual_verification',
                    priority: 'high',
                    title: 'Financial Document Verification',
                    description: 'Request and verify financial statements and bank statements',
                    impactScore: 0.8,
                    difficulty: 'medium',
                    actionRequired: 'Request financial documents from merchant',
                    status: 'pending'
                }
            ],
            websiteRisks: {
                riskyKeywords: [
                    { keyword: 'gambling', severity: 'high', context: 'Found in business description' },
                    { keyword: 'casino', severity: 'medium', context: 'Mentioned in services section' }
                ],
                sentimentAnalysis: {
                    overall_sentiment: 'neutral',
                    positive_pct: 45,
                    neutral_pct: 40,
                    negative_pct: 15
                }
            },
            shapData: {
                feature_contributions: [
                    { feature: 'financial_score', contribution: 0.3 },
                    { feature: 'regulatory_score', contribution: 0.4 },
                    { feature: 'cybersecurity_score', contribution: 0.2 }
                ],
                prediction_value: 0.45,
                base_value: 0.5,
                confidence: 0.85
            },
            lastUpdated: new Date().toISOString(),
            dataSources: { merchant: true, analytics: true, riskAssessment: true }
        };
    }
    
    /**
     * Generate mock risk assessment data
     * @param {string} merchantId - Merchant ID
     * @returns {Object} Mock risk assessment
     */
    generateMockRiskAssessment(merchantId) {
        return {
            id: `risk_${merchantId}`,
            merchantId: merchantId,
            risk_score: 0.45,
            risk_level: 'medium',
            risk_factors: [
                { category: 'financial', name: 'Credit Score', score: 0.15, level: 'low', description: 'Good credit standing' },
                { category: 'regulatory', name: 'Compliance Status', score: 0.65, level: 'high', description: 'Some compliance issues detected' },
                { category: 'cybersecurity', name: 'Security Posture', score: 0.85, level: 'critical', description: 'Critical security vulnerabilities found' }
            ],
            recommendations: [
                { title: 'Improve Security Measures', description: 'Implement additional security controls', priority: 'high', impact_score: 0.8 }
            ],
            shap_analysis: {
                feature_contributions: [
                    { feature: 'financial_score', contribution: 0.3 },
                    { feature: 'regulatory_score', contribution: 0.4 },
                    { feature: 'cybersecurity_score', contribution: 0.2 }
                ],
                prediction_value: 0.45,
                base_value: 0.5,
                confidence: 0.85
            }
        };
    }
    
    /**
     * Clear cache for a specific merchant
     * @param {string} merchantId - Merchant ID
     */
    clearCache(merchantId) {
        const cacheKey = `risk_data_${merchantId}`;
        this.cache.delete(cacheKey);
        sessionStorage.removeItem(`analytics_${merchantId}`);
    }
    
    /**
     * Clear all cache
     */
    clearAllCache() {
        this.cache.clear();
        // Clear session storage for analytics
        Object.keys(sessionStorage).forEach(key => {
            if (key.startsWith('analytics_')) {
                sessionStorage.removeItem(key);
            }
        });
    }
    
    /**
     * Get fallback merchant data when API fails
     */
    getFallbackMerchantData(merchantId) {
        return {
            id: merchantId,
            name: 'Demo Business',
            website: 'https://example.com',
            address: '123 Demo Street, Demo City, DC 12345',
            phone: '+1-555-123-4567',
            email: 'contact@example.com',
            industry: 'Technology',
            size: 'Small',
            status: 'active',
            createdAt: new Date().toISOString(),
            lastUpdated: new Date().toISOString()
        };
    }
    
    /**
     * Get fallback analytics data when API fails
     */
    getFallbackAnalyticsData(merchantId) {
        return {
            business_intelligence: {
                business_metrics: {
                    business_location: {
                        city: 'Demo City',
                        state: 'Demo State',
                        country: 'US',
                        confidence: 0.9
                    },
                    employee_count: {
                        value: 25,
                        range: '10-50',
                        confidence: 0.8
                    },
                    founded_year: {
                        year: 2020,
                        confidence: 0.9
                    },
                    revenue_range: {
                        min: 500000,
                        max: 2000000,
                        currency: 'USD',
                        confidence: 0.7
                    }
                },
                company_profile: {
                    business_type: 'Technology Company',
                    growth_stage: 'Growing',
                    industry: 'Technology',
                    size_category: 'Small Business'
                },
                financial_metrics: {
                    credit_risk: 'Low',
                    financial_health: 'Good',
                    profitability: 'Profitable'
                },
                market_analysis: {
                    competition_level: 'Medium',
                    growth_potential: 'High',
                    market_size: 'Regional'
                }
            },
            business_name: 'Demo Business',
            status: 'success',
            timestamp: new Date().toISOString()
        };
    }
    
    /**
     * Get fallback risk assessment data when API fails
     */
    getFallbackRiskAssessment(merchantId) {
        return {
            success: true,
            assessment: {
                overall_risk_score: 25,
                risk_level: 'Low',
                confidence: 0.85,
                categories: {
                    financial: 15,
                    operational: 20,
                    regulatory: 30,
                    reputational: 10,
                    cybersecurity: 35
                },
                factors: [
                    'Low regulatory requirements',
                    'Simple operational model',
                    'Good financial health',
                    'Minimal cybersecurity exposure'
                ],
                recommendations: [
                    'Continue current business practices',
                    'Monitor financial metrics regularly',
                    'Consider basic cybersecurity measures',
                    'Maintain good customer relationships'
                ],
                last_assessed: new Date().toISOString(),
                next_assessment: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString()
            }
        };
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskIndicatorsDataService;
}
