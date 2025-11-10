/**
 * Merchant Form Component
 * Handles form validation, submission, API calls, and redirect to merchant-details page
 */
class MerchantFormComponent {
    constructor(formId = 'merchantForm') {
        this.form = document.getElementById(formId);
        this.submitBtn = document.getElementById('submitBtn');
        this.clearBtn = document.getElementById('clearFormBtn');
        this.submitLoading = document.getElementById('submitLoading');
        this.isSubmitting = false;
        
        if (!this.form) {
            console.error('Merchant form not found');
            return;
        }
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.initializeMobileOptimization();
    }

    bindEvents() {
        // Form submission
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
        
        // Clear form
        if (this.clearBtn) {
            this.clearBtn.addEventListener('click', () => this.clearForm());
        }
        
        // Real-time validation
        this.form.addEventListener('input', (e) => this.validateField(e.target));
        this.form.addEventListener('blur', (e) => this.validateField(e.target));
        
        // Phone number formatting
        const phoneInput = document.getElementById('phoneNumber');
        if (phoneInput) {
            phoneInput.addEventListener('input', (e) => this.formatPhoneNumber(e.target));
        }
    }

    initializeMobileOptimization() {
        try {
            if (typeof MobileOptimization !== 'undefined') {
                const mobileOpt = new MobileOptimization({
                    touchTargetSize: 48,
                    enableTouchOptimization: true,
                    enableProgressiveEnhancement: true
                });
                if (mobileOpt && typeof mobileOpt.optimizeForMobile === 'function') {
                    mobileOpt.optimizeForMobile();
                }
            }
        } catch (error) {
            console.log('Mobile optimization not available:', error.message);
        }
    }

    validateField(field) {
        const fieldName = field.name;
        const value = field.value.trim();
        const errorElement = document.getElementById(fieldName + 'Error');
        
        // Clear previous validation
        field.classList.remove('error', 'success');
        if (errorElement) errorElement.textContent = '';

        // Required field validation
        if (field.hasAttribute('required') && !value) {
            this.showError(field, errorElement, 'This field is required');
            return false;
        }

        // Skip validation for empty optional fields
        if (!value && !field.hasAttribute('required')) {
            this.showSuccess(field, errorElement);
            return true;
        }

        // Field-specific validation
        switch (fieldName) {
            case 'businessName':
                if (value.length < 2) {
                    this.showError(field, errorElement, 'Business name must be at least 2 characters');
                    return false;
                }
                break;
            
            case 'websiteUrl':
                if (value && !this.isValidUrl(value)) {
                    this.showError(field, errorElement, 'Please enter a valid URL (e.g., https://example.com)');
                    return false;
                }
                break;
            
            case 'email':
                if (value && !this.isValidEmail(value)) {
                    this.showError(field, errorElement, 'Please enter a valid email address');
                    return false;
                }
                break;
            
            case 'phoneNumber':
                if (value && !this.isValidPhone(value)) {
                    this.showError(field, errorElement, 'Please enter a valid phone number');
                    return false;
                }
                break;
            
            case 'country':
                if (field.hasAttribute('required') && !value) {
                    this.showError(field, errorElement, 'Please select a country');
                    return false;
                }
                break;
        }

        this.showSuccess(field, errorElement);
        return true;
    }

    isValidUrl(url) {
        try {
            new URL(url);
            return url.startsWith('http://') || url.startsWith('https://');
        } catch {
            return false;
        }
    }

    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    isValidPhone(phone) {
        const digits = phone.replace(/\D/g, '');
        return digits.length >= 10 && digits.length <= 15;
    }

    formatPhoneNumber(input) {
        let value = input.value.replace(/\D/g, '');
        
        if (value.length > 0) {
            if (value.length <= 3) {
                value = `(${value}`;
            } else if (value.length <= 6) {
                value = `(${value.slice(0, 3)}) ${value.slice(3)}`;
            } else if (value.length <= 10) {
                value = `(${value.slice(0, 3)}) ${value.slice(3, 6)}-${value.slice(6)}`;
            } else {
                value = `+${value.slice(0, -10)} (${value.slice(-10, -7)}) ${value.slice(-7, -4)}-${value.slice(-4)}`;
            }
        }
        
        input.value = value;
    }

    showError(field, errorElement, message) {
        field.classList.add('error');
        field.classList.remove('success');
        if (errorElement) {
            errorElement.innerHTML = `<i class="fas fa-exclamation-circle"></i> ${message}`;
        }
    }

    showSuccess(field, errorElement) {
        field.classList.add('success');
        field.classList.remove('error');
        if (errorElement) {
            errorElement.innerHTML = `<i class="fas fa-check-circle"></i> Valid`;
        }
    }

    validateForm() {
        const fields = this.form.querySelectorAll('input, select, textarea');
        let isValid = true;

        fields.forEach(field => {
            if (!this.validateField(field)) {
                isValid = false;
            }
        });

        return isValid;
    }

    async handleSubmit(e) {
        e.preventDefault();
        
        if (this.isSubmitting) {
            return;
        }
        
        if (!this.validateForm()) {
            this.scrollToFirstError();
            this.showNotification('Please fix the errors in the form before submitting.', 'error');
            return;
        }

        this.isSubmitting = true;
        this.setLoading(true);
        
        try {
            const formData = this.collectFormData();
            await this.processMerchantVerification(formData);
        } catch (error) {
            console.error('Error in handleSubmit:', error);
            this.showNotification('An error occurred while processing your request. Redirecting anyway...', 'error');
            
            // Store form data even on error
            try {
                const formData = this.collectFormData();
                this.storeData(formData, { errors: { general: error.message } });
            } catch (storageError) {
                console.error('Error storing data:', storageError);
            }
            
            // Redirect even on error - merchant-details can handle missing data
            this.finalizeRedirect();
        }
    }

    collectFormData() {
        const formData = new FormData(this.form);
        const data = {};
        
        for (let [key, value] of formData.entries()) {
            data[key] = value.trim();
        }
        
        // Generate business ID for API calls
        data.businessId = this.generateBusinessId(data.businessName);
        
        // Add timestamp and session info
        data.timestamp = new Date().toISOString();
        data.sessionId = this.generateSessionId();
        
        // Structure data for API calls
        data.apiData = {
            businessIntelligence: {
                business_name: data.businessName,
                geographic_region: data.country || 'us',
                website_url: data.websiteUrl || '',
                description: 'Business analysis request',
                analysis_type: data.analysisType || 'comprehensive'
            },
            riskAssessment: {
                business_id: data.businessId,
                business_name: data.businessName,
                categories: this.getSelectedCategories(data.assessmentType),
                include_history: true,
                include_predictions: true
            },
            riskIndicators: {
                business_id: data.businessId,
                business_name: data.businessName,
                merchant_data: {
                    name: data.businessName,
                    website: data.websiteUrl,
                    description: 'Business analysis request',
                    address: this.formatAddress(data),
                    phone: data.phoneNumber,
                    email: data.email,
                    registration: data.registrationNumber,
                    country: data.country
                }
            }
        };
        
        return data;
    }

    generateBusinessId(businessName) {
        const timestamp = Date.now();
        const nameHash = businessName.toLowerCase()
            .replace(/[^a-z0-9]/g, '')
            .substring(0, 8);
        return `biz_${nameHash}_${timestamp}`;
    }

    getSelectedCategories(assessmentType) {
        const categoryMap = {
            'comprehensive': ['financial', 'operational', 'regulatory', 'reputational', 'cybersecurity'],
            'financial': ['financial'],
            'operational': ['operational'],
            'regulatory': ['regulatory'],
            'reputational': ['reputational'],
            'cybersecurity': ['cybersecurity']
        };
        return categoryMap[assessmentType] || ['financial', 'operational', 'regulatory'];
    }

    formatAddress(data) {
        const addressParts = [
            data.streetAddress,
            data.city,
            data.state,
            data.postalCode,
            data.country
        ].filter(part => part && part.trim());
        
        return addressParts.join(', ');
    }

    generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    storeData(merchantData, apiResults = null) {
        try {
            sessionStorage.setItem('merchantData', JSON.stringify(merchantData));
            if (apiResults) {
                sessionStorage.setItem('merchantApiResults', JSON.stringify(apiResults));
            }
            console.log('✅ Data stored in sessionStorage');
            return true;
        } catch (error) {
            console.error('❌ Failed to store data:', error);
            return false;
        }
    }

    async processMerchantVerification(data) {
        // Store form data immediately before API calls
        this.storeData(data);
        
        // Set up fallback redirect timer (max 10 seconds)
        const FALLBACK_REDIRECT_DELAY = 10000;
        const fallbackRedirectTimer = setTimeout(() => {
            console.warn('⚠️ Fallback redirect triggered - APIs taking too long');
            this.finalizeRedirect();
        }, FALLBACK_REDIRECT_DELAY);
        
        try {
            // Make API calls in parallel
            const apiCallsPromise = Promise.allSettled([
                this.callBusinessIntelligenceAPI(data.apiData.businessIntelligence),
                this.callRiskAssessmentAPI(data.apiData.riskAssessment),
                this.callRiskIndicatorsAPI(data.apiData.riskIndicators)
            ]);
            
            // Add overall timeout (30 seconds max)
            const apiTimeoutPromise = new Promise((resolve) => {
                setTimeout(() => resolve('timeout'), 30000);
            });
            
            const result = await Promise.race([apiCallsPromise, apiTimeoutPromise]);
            clearTimeout(fallbackRedirectTimer);
            
            if (result === 'timeout') {
                console.warn('⚠️ API calls timed out, proceeding with redirect');
                this.finalizeRedirect();
                return;
            }
            
            const [businessIntelligenceResult, riskAssessmentResult, riskIndicatorsResult] = result;
            
            // Store API results
            const apiResults = {
                businessIntelligence: businessIntelligenceResult.status === 'fulfilled' ? businessIntelligenceResult.value : null,
                riskAssessment: riskAssessmentResult.status === 'fulfilled' ? riskAssessmentResult.value : null,
                riskIndicators: riskIndicatorsResult.status === 'fulfilled' ? riskIndicatorsResult.value : null,
                errors: {
                    businessIntelligence: businessIntelligenceResult.status === 'rejected' ? this.formatError(businessIntelligenceResult.reason) : null,
                    riskAssessment: riskAssessmentResult.status === 'rejected' ? this.formatError(riskAssessmentResult.reason) : null,
                    riskIndicators: riskIndicatorsResult.status === 'rejected' ? this.formatError(riskIndicatorsResult.reason) : null
                }
            };
            
            this.storeData(data, apiResults);
            this.finalizeRedirect();
            
        } catch (error) {
            clearTimeout(fallbackRedirectTimer);
            console.error('❌ Error in merchant verification process:', error);
            
            // Store error results
            const errorResults = {
                businessIntelligence: null,
                riskAssessment: null,
                riskIndicators: null,
                errors: {
                    general: error.message || 'Unknown error occurred during verification'
                }
            };
            this.storeData(data, errorResults);
            this.finalizeRedirect();
        }
    }

    formatError(error) {
        if (error instanceof Error) {
            return {
                message: error.message,
                name: error.name,
                stack: error.stack
            };
        }
        return String(error);
    }

    finalizeRedirect() {
        // Verify data is stored before redirecting
        const merchantData = sessionStorage.getItem('merchantData');
        if (!merchantData) {
            console.warn('⚠️ No merchant data in sessionStorage - redirecting anyway');
        }
        
        // Use window.location.replace() to prevent back-button issues
        // Use absolute URL to avoid hash interference
        const targetUrl = window.location.origin + '/merchant-details';
        
        try {
            // Small delay to ensure sessionStorage is written
            setTimeout(() => {
                window.location.replace(targetUrl);
            }, 50);
        } catch (error) {
            console.error('❌ Error during redirect:', error);
            // Fallback: try relative path
            try {
                window.location.replace('/merchant-details');
            } catch (fallbackError) {
                console.error('❌ Fallback redirect also failed:', fallbackError);
                this.showNotification('Failed to redirect. Please navigate to /merchant-details manually.', 'error');
            }
        }
    }

    async callBusinessIntelligenceAPI(apiData) {
        if (!window.APIConfig) {
            throw new Error('APIConfig not available');
        }
        
        const apiUrl = APIConfig.getEndpoints().classify;
        const timeout = 25000; // 25 seconds
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), timeout);
        
        try {
            const response = await fetch(apiUrl, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                },
                body: JSON.stringify(apiData),
                signal: controller.signal,
                credentials: 'omit'
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                let errorText = 'Unknown error';
                try {
                    errorText = await response.text();
                } catch (e) {
                    // Ignore
                }
                
                const error = new Error(`Business Intelligence API error: ${response.status} ${response.statusText}`);
                error.status = response.status;
                error.statusText = response.statusText;
                throw error;
            }

            const contentType = response.headers.get('content-type') || '';
            if (contentType.includes('application/json')) {
                return await response.json();
            } else {
                const text = await response.text();
                throw new Error(`Expected JSON but received ${contentType}`);
            }
        } catch (error) {
            clearTimeout(timeoutId);
            
            if (error.name === 'AbortError') {
                throw new Error('Business Intelligence API call timed out');
            }
            
            if (error.name === 'TypeError' && error.message.includes('fetch')) {
                throw new Error('Network error: Unable to reach API. Please check your connection.');
            }
            
            throw error;
        }
    }

    async callRiskAssessmentAPI(apiData) {
        // Generate risk assessment data based on business name
        return new Promise((resolve) => {
            setTimeout(() => {
                const businessName = apiData.business_name || '';
                const riskScore = this.calculateRiskScore(businessName);
                
                resolve({
                    success: true,
                    assessment: {
                        overall_risk_score: riskScore,
                        risk_level: this.getRiskLevel(riskScore),
                        confidence: 0.85,
                        categories: {
                            financial: Math.max(10, riskScore - 10),
                            operational: Math.max(15, riskScore - 5),
                            regulatory: Math.max(20, riskScore + 5),
                            reputational: Math.max(5, riskScore - 15),
                            cybersecurity: Math.max(25, riskScore + 10)
                        },
                        factors: this.generateRiskFactors(riskScore),
                        recommendations: this.generateRiskRecommendations(riskScore)
                    }
                });
            }, 800);
        });
    }

    async callRiskIndicatorsAPI(apiData) {
        // Generate risk indicators data
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve({
                    success: true,
                    risk_indicators: {
                        financial: 15,
                        operational: 35,
                        regulatory: 65,
                        cybersecurity: 85,
                        overall: 50
                    },
                    trends: {
                        financial: 'improving',
                        operational: 'stable',
                        regulatory: 'rising',
                        cybersecurity: 'escalating'
                    }
                });
            }, 1000);
        });
    }

    calculateRiskScore(businessName) {
        let riskScore = 25; // Base risk score
        const name = businessName.toLowerCase();
        
        // High-risk keywords
        if (name.includes('crypto') || name.includes('bitcoin') || name.includes('forex')) {
            riskScore += 30;
        }
        if (name.includes('gambling') || name.includes('casino') || name.includes('betting')) {
            riskScore += 25;
        }
        if (name.includes('pharmaceutical') || name.includes('medical') || name.includes('health')) {
            riskScore += 15;
        }
        if (name.includes('financial') || name.includes('investment') || name.includes('trading')) {
            riskScore += 20;
        }
        
        // Low-risk keywords
        if (name.includes('consulting') || name.includes('services') || name.includes('solutions')) {
            riskScore -= 10;
        }
        if (name.includes('technology') || name.includes('software') || name.includes('tech')) {
            riskScore -= 5;
        }
        
        return Math.max(5, Math.min(95, riskScore));
    }
    
    getRiskLevel(score) {
        if (score <= 25) return 'Low';
        if (score <= 50) return 'Medium';
        if (score <= 75) return 'High';
        return 'Critical';
    }
    
    generateRiskFactors(riskScore) {
        const factors = [];
        
        if (riskScore > 70) {
            factors.push('High regulatory compliance requirements');
            factors.push('Complex operational environment');
            factors.push('Potential cybersecurity vulnerabilities');
        } else if (riskScore > 40) {
            factors.push('Moderate compliance requirements');
            factors.push('Standard operational risks');
        } else {
            factors.push('Low regulatory requirements');
            factors.push('Simple operational model');
        }
        
        return factors;
    }
    
    generateRiskRecommendations(riskScore) {
        const recommendations = [];
        
        if (riskScore > 70) {
            recommendations.push('Implement comprehensive compliance monitoring');
            recommendations.push('Enhance cybersecurity measures');
            recommendations.push('Regular risk assessments required');
        } else if (riskScore > 40) {
            recommendations.push('Standard compliance monitoring');
            recommendations.push('Basic security measures');
        } else {
            recommendations.push('Minimal monitoring required');
            recommendations.push('Standard business practices');
        }
        
        return recommendations;
    }

    setLoading(loading) {
        if (this.submitBtn) {
            this.submitBtn.disabled = loading;
        }
        if (this.submitLoading) {
            this.submitLoading.classList.toggle('show', loading);
        }
        
        const btnText = this.submitBtn?.querySelector('.btn-text');
        if (btnText) {
            btnText.style.display = loading ? 'none' : 'inline';
        }
    }

    clearForm() {
        this.form.reset();
        
        // Clear all validation states
        const fields = this.form.querySelectorAll('input, select, textarea');
        fields.forEach(field => {
            field.classList.remove('error', 'success');
        });
        
        const errorElements = this.form.querySelectorAll('.error-message');
        errorElements.forEach(element => {
            element.textContent = '';
        });
        
        // Focus on first field
        const firstField = document.getElementById('businessName');
        if (firstField) {
            firstField.focus();
        }
    }

    scrollToFirstError() {
        const firstError = this.form.querySelector('.error');
        if (firstError) {
            firstError.scrollIntoView({ behavior: 'smooth', block: 'center' });
            firstError.focus();
        }
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.innerHTML = `
            <i class="fas fa-${type === 'error' ? 'exclamation-circle' : type === 'success' ? 'check-circle' : 'info-circle'}"></i>
            ${message}
        `;
        
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${type === 'error' ? '#e74c3c' : type === 'success' ? '#27ae60' : '#3498db'};
            color: white;
            padding: 15px 20px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            z-index: 10000;
            display: flex;
            align-items: center;
            gap: 10px;
            max-width: 400px;
            animation: slideIn 0.3s ease;
        `;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => notification.remove(), 300);
        }, 5000);
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MerchantFormComponent;
}

