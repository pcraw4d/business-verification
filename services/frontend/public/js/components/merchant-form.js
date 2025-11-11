/**
 * Merchant Form Component
 * Handles form validation, submission, API calls, and redirect to merchant-details page
 */
class MerchantFormComponent {
    constructor(formId = 'merchantForm') {
        console.log('🔍 [DEBUG] MerchantFormComponent constructor called with formId:', formId);
        console.log('🔍 [DEBUG] Document ready state:', document.readyState);
        console.log('🔍 [DEBUG] Window location:', window.location.href);
        
        this.form = document.getElementById(formId);
        this.submitBtn = document.getElementById('submitBtn');
        this.clearBtn = document.getElementById('clearFormBtn');
        this.submitLoading = document.getElementById('submitLoading');
        this.isSubmitting = false;
        
        console.log('🔍 [DEBUG] Form element found:', !!this.form);
        console.log('🔍 [DEBUG] Submit button found:', !!this.submitBtn);
        console.log('🔍 [DEBUG] Clear button found:', !!this.clearBtn);
        console.log('🔍 [DEBUG] Submit loading element found:', !!this.submitLoading);
        
        if (!this.form) {
            console.error('❌ [ERROR] Merchant form not found with ID:', formId);
            console.error('❌ [ERROR] Available form elements:', Array.from(document.querySelectorAll('form')).map(f => f.id || '(no id)'));
            return;
        }
        
        console.log('✅ [DEBUG] Form found, calling init()...');
        this.init();
        console.log('✅ [DEBUG] MerchantFormComponent initialization complete');
    }

    init() {
        console.log('🔍 [DEBUG] init() called');
        try {
            this.bindEvents();
            console.log('✅ [DEBUG] bindEvents() completed');
            this.initializeMobileOptimization();
            console.log('✅ [DEBUG] initializeMobileOptimization() completed');
        } catch (error) {
            console.error('❌ [ERROR] Error in init():', error);
            console.error('❌ [ERROR] Error stack:', error.stack);
            throw error;
        }
    }

    bindEvents() {
        console.log('🔍 [DEBUG] bindEvents() called');
        
        try {
            // Form submission - use capture phase to catch early, non-passive
            const self = this;
            const handleFormSubmit = function(e) {
                console.log('🔍 [DEBUG] ========== FORM SUBMIT HANDLER FIRED ==========');
                console.log('🔍 [DEBUG] Form submit event triggered - handler FIRED!');
                console.log('🔍 [DEBUG] Event type:', e.type);
                console.log('🔍 [DEBUG] Event target:', e.target);
                console.log('🔍 [DEBUG] Handler context (self):', self);
                
                // CRITICAL: Prevent default FIRST
                if (e && typeof e.preventDefault === 'function') {
                    e.preventDefault();
                }
                if (e && typeof e.stopPropagation === 'function') {
                    e.stopPropagation();
                }
                if (e && typeof e.stopImmediatePropagation === 'function') {
                    e.stopImmediatePropagation();
                }
                
                console.log('✅ [DEBUG] Default form submission prevented and propagation stopped');
                console.log('🔍 [DEBUG] Event defaultPrevented after preventDefault:', e ? e.defaultPrevented : 'N/A');
                
                try {
                    self.handleSubmit(e);
                } catch (error) {
                    console.error('❌ [ERROR] Error in handleFormSubmit calling handleSubmit:', error);
                    console.error('❌ [ERROR] Error stack:', error.stack);
                }
            };
            
            // Store handler reference for debugging
            this._handleFormSubmit = handleFormSubmit;
            
            console.log('🔍 [DEBUG] Attaching submit event listener to form (capture phase, non-passive)');
            this.form.addEventListener('submit', handleFormSubmit, { capture: true, passive: false });
            console.log('✅ [DEBUG] Submit event listener attached to form (capture, non-passive)');
            
            // Also add in bubble phase as backup
            console.log('🔍 [DEBUG] Attaching submit event listener to form (bubble phase, non-passive)');
            this.form.addEventListener('submit', handleFormSubmit, { capture: false, passive: false });
            console.log('✅ [DEBUG] Submit event listener attached to form (bubble, non-passive)');
            
            // Also listen to button click - use capture phase and handle immediately
            if (this.submitBtn) {
                console.log('🔍 [DEBUG] Attaching click event listener to submit button (capture phase)');
                
                // Store reference to this for use in handlers
                const self = this;
                
                const handleButtonClick = function(e) {
                    console.log('🔍 [DEBUG] ========== COMPONENT BUTTON CLICK HANDLER FIRED ==========');
                    console.log('🔍 [DEBUG] Submit button clicked - handler FIRED!');
                    console.log('🔍 [DEBUG] Handler function reference:', handleButtonClick);
                    console.log('🔍 [DEBUG] Handler attached to element:', self.submitBtn);
                    console.log('🔍 [DEBUG] Event type:', e ? e.type : 'NO EVENT');
                    console.log('🔍 [DEBUG] Event target:', e ? e.target : 'NO EVENT');
                    console.log('🔍 [DEBUG] Event currentTarget:', e ? e.currentTarget : 'NO EVENT');
                    console.log('🔍 [DEBUG] Handler context (self):', self);
                    console.log('🔍 [DEBUG] Form element:', self ? self.form : 'NO SELF');
                    
                    // CRITICAL: Prevent default FIRST
                    if (e && typeof e.preventDefault === 'function') {
                        e.preventDefault();
                    }
                    if (e && typeof e.stopPropagation === 'function') {
                        e.stopPropagation();
                    }
                    if (e && typeof e.stopImmediatePropagation === 'function') {
                        e.stopImmediatePropagation();
                    }
                    
                    console.log('✅ [DEBUG] Default prevented, propagation stopped');
                    console.log('🔍 [DEBUG] Event defaultPrevented after preventDefault:', e ? e.defaultPrevented : 'N/A');
                    
                    // Process immediately - no delay
                    console.log('🔍 [DEBUG] Processing button click immediately');
                    
                    // Trigger form validation and submission
                    console.log('🔍 [DEBUG] Checking form validity...');
                    try {
                        const isValid = self.form ? self.form.checkValidity() : false;
                        console.log('🔍 [DEBUG] Form validity result:', isValid);
                        
                        if (isValid) {
                            console.log('🔍 [DEBUG] Form is valid, calling handleSubmit');
                            // Call handleSubmit and ensure it completes
                            const submitPromise = self.handleSubmit(e);
                            if (submitPromise && typeof submitPromise.then === 'function') {
                                submitPromise.catch(err => {
                                    console.error('❌ [ERROR] handleSubmit promise rejected:', err);
                                    console.error('❌ [ERROR] Error stack:', err.stack);
                                });
                            }
                        } else {
                            console.warn('⚠️ [DEBUG] Form validation failed on button click');
                            // Get invalid fields for debugging
                            const invalidFields = Array.from(self.form.elements).filter(el => !el.validity.valid && el.required);
                            console.warn('⚠️ [DEBUG] Invalid required fields:', invalidFields.map(f => ({ 
                                name: f.name || f.id, 
                                value: f.value, 
                                validationMessage: f.validationMessage 
                            })));
                            if (self.form) {
                                self.form.reportValidity();
                            }
                        }
                    } catch (error) {
                        console.error('❌ [ERROR] Error in handleButtonClick:', error);
                        console.error('❌ [ERROR] Error stack:', error.stack);
                    }
                    
                    return false;
                };
                
                // Store handler reference for debugging
                this._handleButtonClick = handleButtonClick;
                
                // Capture phase - highest priority - MUST be first
                // Use once: false to ensure it stays attached
                this.submitBtn.addEventListener('click', handleButtonClick, { capture: true, passive: false, once: false });
                console.log('✅ [DEBUG] Click event listener attached to submit button (capture, non-passive, once: false)');
                
                // Verify the listener was actually attached
                console.log('🔍 [DEBUG] Verifying listener attachment...');
                // Note: getEventListeners is Chrome DevTools only, not available in production
                // But we can verify by checking if the button has onclick or by trying to remove it
                if (this.submitBtn.onclick) {
                    console.log('✅ [DEBUG] Button onclick property is set');
                }
                
                // Also attach to mousedown as backup (fires before click)
                this.submitBtn.addEventListener('mousedown', function(e) {
                    console.log('🔍 [DEBUG] Submit button mousedown - handler FIRED!');
                    // Don't prevent default on mousedown, just log
                }, { capture: true, passive: true });
                console.log('✅ [DEBUG] Mousedown event listener attached (capture, passive)');
                
                // Bubble phase - backup
                this.submitBtn.addEventListener('click', handleButtonClick, { capture: false, passive: false });
                console.log('✅ [DEBUG] Click event listener attached to submit button (bubble, non-passive)');
                
                // onclick handler - set early and don't let it be overwritten
                // Use Object.defineProperty to make it non-configurable
                const originalOnclick = this.submitBtn.onclick;
                const onclickHandler = function(e) {
                    console.log('🔍 [DEBUG] ========== ONCLICK HANDLER FIRED ==========');
                    console.log('🔍 [DEBUG] Submit button onclick handler triggered');
                    if (e) {
                        e.preventDefault();
                        e.stopPropagation();
                    }
                    if (self.form.checkValidity()) {
                        console.log('🔍 [DEBUG] Form is valid (onclick), calling handleSubmit');
                        self.handleSubmit(e);
                    } else {
                        console.warn('⚠️ [DEBUG] Form validation failed (onclick)');
                        self.form.reportValidity();
                    }
                    return false;
                };
                
                // Try to make onclick non-configurable
                try {
                    Object.defineProperty(this.submitBtn, 'onclick', {
                        value: onclickHandler,
                        writable: false,
                        configurable: false
                    });
                    console.log('✅ [DEBUG] onclick handler attached (non-configurable)');
                } catch (e) {
                    // Fallback if defineProperty fails
                    this.submitBtn.onclick = onclickHandler;
                    console.log('✅ [DEBUG] onclick handler attached (fallback method)');
                    console.warn('⚠️ [DEBUG] Could not make onclick non-configurable:', e.message);
                }
                
                // Final verification - check if onclick is actually set
                setTimeout(() => {
                    if (this.submitBtn.onclick === onclickHandler) {
                        console.log('✅ [DEBUG] onclick handler verified - still attached after 100ms');
                    } else {
                        console.error('❌ [ERROR] onclick handler was removed or changed!');
                        console.error('❌ [ERROR] Expected:', onclickHandler);
                        console.error('❌ [ERROR] Got:', this.submitBtn.onclick);
                        // Re-attach it
                        this.submitBtn.onclick = onclickHandler;
                        console.log('✅ [DEBUG] onclick handler re-attached');
                    }
                }, 100);
                
                // Make button type="button" to prevent form submission
                if (this.submitBtn.type === 'submit') {
                    console.log('🔍 [DEBUG] Changing button type from submit to button');
                    this.submitBtn.type = 'button';
                    console.log('✅ [DEBUG] Button type changed to prevent default form submission');
                }
            } else {
                console.warn('⚠️ [DEBUG] Submit button not found, skipping click listener');
            }
            
            // Clear form
            if (this.clearBtn) {
                console.log('🔍 [DEBUG] Attaching click event listener to clear button');
                this.clearBtn.addEventListener('click', () => this.clearForm());
                console.log('✅ [DEBUG] Clear button event listener attached');
            } else {
                console.warn('⚠️ [DEBUG] Clear button not found, skipping event listener');
            }
            
            // Real-time validation
            console.log('🔍 [DEBUG] Attaching input and blur event listeners for validation');
            this.form.addEventListener('input', (e) => this.validateField(e.target));
            this.form.addEventListener('blur', (e) => this.validateField(e.target));
            console.log('✅ [DEBUG] Validation event listeners attached');
            
            // Phone number formatting
            const phoneInput = document.getElementById('phoneNumber');
            if (phoneInput) {
                console.log('🔍 [DEBUG] Attaching phone number formatting listener');
                phoneInput.addEventListener('input', (e) => this.formatPhoneNumber(e.target));
                console.log('✅ [DEBUG] Phone number formatting listener attached');
            } else {
                console.warn('⚠️ [DEBUG] Phone input not found, skipping formatting listener');
            }
            
            console.log('✅ [DEBUG] All event listeners attached successfully');
        } catch (error) {
            console.error('❌ [ERROR] Error in bindEvents():', error);
            console.error('❌ [ERROR] Error stack:', error.stack);
            throw error;
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
        console.log('🔍 [DEBUG] handleSubmit() called');
        console.log('🔍 [DEBUG] Event object:', e);
        console.log('🔍 [DEBUG] Event type:', e?.type);
        console.log('🔍 [DEBUG] Event target:', e?.target);
        console.log('🔍 [DEBUG] Current isSubmitting state:', this.isSubmitting);
        
        // Prevent default again just in case
        if (e) {
            e.preventDefault();
            e.stopPropagation();
            e.stopImmediatePropagation();
        }
        
        try {
            console.log('✅ [DEBUG] Default form submission prevented (in handleSubmit)');
            
            if (this.isSubmitting) {
                console.warn('⚠️ [DEBUG] Form is already submitting, ignoring duplicate submission');
                return;
            }
            
            console.log('🔍 [DEBUG] Starting form validation...');
            const isValid = this.validateForm();
            console.log('🔍 [DEBUG] Form validation result:', isValid);
            
            if (!isValid) {
                console.warn('⚠️ [DEBUG] Form validation failed, showing errors');
                this.scrollToFirstError();
                this.showNotification('Please fix the errors in the form before submitting.', 'error');
                return;
            }

            console.log('✅ [DEBUG] Form validation passed, proceeding with submission');
            this.isSubmitting = true;
            this.setLoading(true);
            console.log('🔍 [DEBUG] Loading state set to true');
            
            console.log('🔍 [DEBUG] Collecting form data...');
            const formData = this.collectFormData();
            console.log('✅ [DEBUG] Form data collected:', Object.keys(formData));
            
            console.log('🔍 [DEBUG] Starting processMerchantVerification...');
            await this.processMerchantVerification(formData);
            console.log('✅ [DEBUG] processMerchantVerification completed');
        } catch (error) {
            console.error('❌ [ERROR] Error in handleSubmit:', error);
            console.error('❌ [ERROR] Error name:', error.name);
            console.error('❌ [ERROR] Error message:', error.message);
            console.error('❌ [ERROR] Error stack:', error.stack);
            this.showNotification('An error occurred while processing your request. Redirecting anyway...', 'error');
            
            // Store form data even on error
            try {
                console.log('🔍 [DEBUG] Attempting to store form data on error...');
                const formData = this.collectFormData();
                this.storeData(formData, { errors: { general: error.message } });
                // Try to extract merchant ID from stored data if available
                const storedData = sessionStorage.getItem('merchantData');
                let merchantId = null;
                if (storedData) {
                    try {
                        const parsed = JSON.parse(storedData);
                        merchantId = parsed.merchantId || parsed.id || null;
                        console.log('🔍 [DEBUG] Extracted merchant ID from stored data:', merchantId);
                    } catch (e) {
                        console.warn('⚠️ [DEBUG] Failed to parse stored merchant data:', e);
                    }
                }
                console.log('🔍 [DEBUG] Redirecting after error, merchantId:', merchantId);
                this.finalizeRedirect(merchantId);
            } catch (storageError) {
                console.error('❌ [ERROR] Error storing data:', storageError);
                console.error('❌ [ERROR] Storage error stack:', storageError.stack);
                this.finalizeRedirect(null);
            }
        }
    }

    collectFormData() {
        console.log('🔍 [DEBUG] collectFormData() called');
        try {
            const formData = new FormData(this.form);
            const data = {};
            
            console.log('🔍 [DEBUG] Collecting form field values...');
            for (let [key, value] of formData.entries()) {
                data[key] = value.trim();
            }
            console.log('🔍 [DEBUG] Form fields collected:', Object.keys(data));
            
            // Generate business ID for API calls
            console.log('🔍 [DEBUG] Generating business ID...');
            data.businessId = this.generateBusinessId(data.businessName);
            console.log('🔍 [DEBUG] Generated business ID:', data.businessId);
            
            // Add timestamp and session info
            data.timestamp = new Date().toISOString();
            data.sessionId = this.generateSessionId();
            console.log('🔍 [DEBUG] Added timestamp and session ID');
            
            // Structure data for API calls
            console.log('🔍 [DEBUG] Structuring data for API calls...');
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
            
            console.log('✅ [DEBUG] Form data collection complete');
            console.log('🔍 [DEBUG] Final data structure keys:', Object.keys(data));
            return data;
        } catch (error) {
            console.error('❌ [ERROR] Error in collectFormData():', error);
            console.error('❌ [ERROR] Error stack:', error.stack);
            throw error;
        }
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
        console.log('🔍 [DEBUG] processMerchantVerification() called');
        console.log('🔍 [DEBUG] Data keys:', Object.keys(data));
        
        // Store form data immediately before API calls
        console.log('🔍 [DEBUG] Storing form data in sessionStorage...');
        this.storeData(data);
        console.log('✅ [DEBUG] Form data stored');
        
        let savedMerchantId = null;
        
        // First, save the merchant to the portfolio
        try {
            console.log('💾 [DEBUG] Starting merchant save to portfolio...');
            console.log('💾 [DEBUG] Calling saveMerchantToPortfolio()...');
            const merchantResponse = await this.saveMerchantToPortfolio(data);
            console.log('🔍 [DEBUG] saveMerchantToPortfolio() returned:', merchantResponse);
            
            if (merchantResponse && merchantResponse.id) {
                savedMerchantId = merchantResponse.id;
                console.log('✅ [DEBUG] Merchant saved to portfolio with ID:', savedMerchantId);
                // Update the data with the saved merchant ID
                data.merchantId = savedMerchantId;
                data.id = savedMerchantId;
                console.log('🔍 [DEBUG] Updating stored data with merchant ID...');
                this.storeData(data);
                console.log('✅ [DEBUG] Data updated with merchant ID');
            } else {
                console.warn('⚠️ [DEBUG] Merchant response missing ID:', merchantResponse);
            }
        } catch (error) {
            console.error('❌ [ERROR] Failed to save merchant to portfolio:', error);
            console.error('❌ [ERROR] Error name:', error.name);
            console.error('❌ [ERROR] Error message:', error.message);
            console.error('❌ [ERROR] Error stack:', error.stack);
            // Continue anyway - we'll use the generated businessId
            this.showNotification('Warning: Merchant may not be saved to portfolio. Continuing with verification...', 'error');
        }
        
        // Set up fallback redirect timer (max 10 seconds)
        const FALLBACK_REDIRECT_DELAY = 10000;
        console.log('🔍 [DEBUG] Setting up fallback redirect timer:', FALLBACK_REDIRECT_DELAY, 'ms');
        const fallbackRedirectTimer = setTimeout(() => {
            console.warn('⚠️ [DEBUG] Fallback redirect triggered - APIs taking too long');
            this.finalizeRedirect(savedMerchantId);
        }, FALLBACK_REDIRECT_DELAY);
        
        try {
            console.log('🔍 [DEBUG] Starting parallel API calls...');
            // Make API calls in parallel
            const apiCallsPromise = Promise.allSettled([
                this.callBusinessIntelligenceAPI(data.apiData.businessIntelligence),
                this.callRiskAssessmentAPI(data.apiData.riskAssessment),
                this.callRiskIndicatorsAPI(data.apiData.riskIndicators)
            ]);
            console.log('✅ [DEBUG] API calls promise created');
            
            // Add overall timeout (30 seconds max)
            console.log('🔍 [DEBUG] Setting up API timeout (30 seconds)...');
            const apiTimeoutPromise = new Promise((resolve) => {
                setTimeout(() => {
                    console.warn('⚠️ [DEBUG] API timeout promise resolved');
                    resolve('timeout');
                }, 30000);
            });
            
            console.log('🔍 [DEBUG] Racing API calls against timeout...');
            const result = await Promise.race([apiCallsPromise, apiTimeoutPromise]);
            clearTimeout(fallbackRedirectTimer);
            console.log('🔍 [DEBUG] Promise race completed, result type:', typeof result);
            
            if (result === 'timeout') {
                console.warn('⚠️ [DEBUG] API calls timed out, proceeding with redirect');
                this.finalizeRedirect(savedMerchantId);
                return;
            }
            
            console.log('🔍 [DEBUG] API calls completed, processing results...');
            const [businessIntelligenceResult, riskAssessmentResult, riskIndicatorsResult] = result;
            console.log('🔍 [DEBUG] Business Intelligence status:', businessIntelligenceResult.status);
            console.log('🔍 [DEBUG] Risk Assessment status:', riskAssessmentResult.status);
            console.log('🔍 [DEBUG] Risk Indicators status:', riskIndicatorsResult.status);
            
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
            
            console.log('🔍 [DEBUG] Storing API results...');
            this.storeData(data, apiResults);
            console.log('✅ [DEBUG] API results stored, finalizing redirect...');
            this.finalizeRedirect(savedMerchantId);
            
        } catch (error) {
            clearTimeout(fallbackRedirectTimer);
            console.error('❌ [ERROR] Error in merchant verification process:', error);
            console.error('❌ [ERROR] Error name:', error.name);
            console.error('❌ [ERROR] Error message:', error.message);
            console.error('❌ [ERROR] Error stack:', error.stack);
            
            // Store error results
            const errorResults = {
                businessIntelligence: null,
                riskAssessment: null,
                riskIndicators: null,
                errors: {
                    general: error.message || 'Unknown error occurred during verification'
                }
            };
            console.log('🔍 [DEBUG] Storing error results...');
            this.storeData(data, errorResults);
            console.log('🔍 [DEBUG] Finalizing redirect after error...');
            this.finalizeRedirect(savedMerchantId);
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

    finalizeRedirect(merchantId = null) {
        // Verify data is stored before redirecting
        const merchantData = sessionStorage.getItem('merchantData');
        if (!merchantData) {
            console.warn('⚠️ No merchant data in sessionStorage - redirecting anyway');
        } else {
            console.log('✅ Merchant data confirmed in sessionStorage');
        }
        
        // Build target URL with merchant ID if available
        // Use .html extension for compatibility with both route patterns
        let targetUrl = '/merchant-details.html';
        if (merchantId) {
            // Try both 'id' and 'merchantId' parameters for maximum compatibility
            targetUrl += `?id=${encodeURIComponent(merchantId)}&merchantId=${encodeURIComponent(merchantId)}`;
            console.log('🔀 Redirecting with merchant ID:', merchantId);
        } else {
            // Try to get merchant ID from sessionStorage as fallback
            try {
                const storedData = sessionStorage.getItem('merchantData');
                if (storedData) {
                    const parsed = JSON.parse(storedData);
                    const fallbackId = parsed.id || parsed.merchantId || parsed.businessId;
                    if (fallbackId) {
                        targetUrl += `?id=${encodeURIComponent(fallbackId)}&merchantId=${encodeURIComponent(fallbackId)}`;
                        console.log('🔀 Using merchant ID from sessionStorage:', fallbackId);
                    }
                }
            } catch (e) {
                console.warn('⚠️ [DEBUG] Could not extract merchant ID from sessionStorage:', e);
            }
            if (!targetUrl.includes('?')) {
                console.log('🔀 Redirecting without merchant ID (will use sessionStorage data)');
            }
        }
        
        console.log('🔀 Full redirect URL:', targetUrl);
        console.log('🔀 Current URL:', window.location.href);
        console.log('🔀 Current origin:', window.location.origin);
        console.log('🔀 Absolute redirect URL:', window.location.origin + targetUrl);
        
        // Add a small delay to ensure sessionStorage writes are flushed before redirect
        // This prevents race conditions where navigation happens before data persistence completes
        setTimeout(() => {
            try {
                console.log('🔀 Executing redirect after sessionStorage flush delay...');
                console.log('🔀 Attempting navigation to:', targetUrl);
                
                // Force navigation - try multiple methods
                // Method 1: window.location.href (most compatible)
                try {
                    window.location.href = targetUrl;
                    console.log('✅ [DEBUG] window.location.href set successfully');
                    // Give it a moment, then check if navigation started
                    setTimeout(() => {
                        if (window.location.href.includes('merchant-details')) {
                            console.log('✅ [DEBUG] Navigation successful!');
                        } else {
                            console.warn('⚠️ [DEBUG] Navigation may not have started, trying alternative...');
                            // Try method 2
                            window.location.assign(targetUrl);
                        }
                    }, 50);
                } catch (hrefError) {
                    console.error('❌ [ERROR] window.location.href failed:', hrefError);
                    throw hrefError;
                }
            } catch (error) {
                console.error('❌ [ERROR] Error during redirect:', error);
                console.error('❌ [ERROR] Error name:', error.name);
                console.error('❌ [ERROR] Error message:', error.message);
                console.error('❌ [ERROR] Error stack:', error.stack);
                
                // Fallback: try absolute URL
                try {
                    const absoluteUrl = window.location.origin + targetUrl;
                    console.log('🔀 Trying absolute URL:', absoluteUrl);
                    window.location.href = absoluteUrl;
                } catch (fallbackError) {
                    console.error('❌ [ERROR] Fallback redirect also failed:', fallbackError);
                    
                    // Try window.location.assign as last resort
                    try {
                        console.log('🔀 Trying window.location.assign...');
                        window.location.assign(window.location.origin + targetUrl);
                    } catch (assignError) {
                        console.error('❌ [ERROR] window.location.assign also failed:', assignError);
                        // Last resort: show notification with manual link
                        const absoluteUrl = window.location.origin + targetUrl;
                        this.showNotification(
                            `Redirect failed. Please <a href="${absoluteUrl}" style="color: white; text-decoration: underline; font-weight: bold;">click here</a> to view merchant details.`, 
                            'error'
                        );
                    }
                }
            }
        }, 100); // 100ms delay to ensure sessionStorage writes complete
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
            console.log('🔍 [DEBUG] Business Intelligence API response received');
            console.log('🔍 [DEBUG] Response status:', response.status);

            if (!response.ok) {
                console.error('❌ [ERROR] Business Intelligence API error:', response.status, response.statusText);
                let errorText = 'Unknown error';
                try {
                    errorText = await response.text();
                    console.error('❌ [ERROR] Error response text:', errorText);
                } catch (e) {
                    console.warn('⚠️ [DEBUG] Could not read error response text');
                }
                
                const error = new Error(`Business Intelligence API error: ${response.status} ${response.statusText}`);
                error.status = response.status;
                error.statusText = response.statusText;
                throw error;
            }

            const contentType = response.headers.get('content-type') || '';
            console.log('🔍 [DEBUG] Response content-type:', contentType);
            if (contentType.includes('application/json')) {
                console.log('🔍 [DEBUG] Parsing JSON response...');
                const result = await response.json();
                console.log('✅ [DEBUG] Business Intelligence API call successful');
                return result;
            } else {
                console.warn('⚠️ [DEBUG] Unexpected content-type, reading as text');
                const text = await response.text();
                console.error('❌ [ERROR] Response text:', text);
                throw new Error(`Expected JSON but received ${contentType}`);
            }
        } catch (error) {
            clearTimeout(timeoutId);
            console.error('❌ [ERROR] Error in callBusinessIntelligenceAPI:', error);
            console.error('❌ [ERROR] Error name:', error.name);
            console.error('❌ [ERROR] Error message:', error.message);
            console.error('❌ [ERROR] Error stack:', error.stack);
            
            if (error.name === 'AbortError') {
                console.error('❌ [ERROR] Request was aborted (timeout)');
                throw new Error('Business Intelligence API call timed out');
            }
            
            if (error.name === 'TypeError' && error.message.includes('fetch')) {
                console.error('❌ [ERROR] Network error detected');
                throw new Error('Network error: Unable to reach API. Please check your connection.');
            }
            
            throw error;
        }
    }

    async callRiskAssessmentAPI(apiData) {
        console.log('🔍 [DEBUG] callRiskAssessmentAPI() called');
        console.log('🔍 [DEBUG] API data:', apiData);
        // Generate risk assessment data based on business name
        return new Promise((resolve) => {
            console.log('🔍 [DEBUG] Creating risk assessment promise...');
            setTimeout(() => {
                console.log('🔍 [DEBUG] Risk assessment timeout completed, generating data...');
                const businessName = apiData.business_name || '';
                const riskScore = this.calculateRiskScore(businessName);
                console.log('🔍 [DEBUG] Calculated risk score:', riskScore);
                
                const result = {
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
                };
                console.log('✅ [DEBUG] Risk assessment data generated');
                resolve(result);
            }, 800);
        });
    }

    async callRiskIndicatorsAPI(apiData) {
        console.log('🔍 [DEBUG] callRiskIndicatorsAPI() called');
        console.log('🔍 [DEBUG] API data:', apiData);
        // Generate risk indicators data
        return new Promise((resolve) => {
            console.log('🔍 [DEBUG] Creating risk indicators promise...');
            setTimeout(() => {
                console.log('🔍 [DEBUG] Risk indicators timeout completed, generating data...');
                const result = {
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
                };
                console.log('✅ [DEBUG] Risk indicators data generated');
                resolve(result);
            }, 1000);
        });
    }

    async saveMerchantToPortfolio(formData) {
        console.log('🔍 [DEBUG] saveMerchantToPortfolio() called');
        console.log('🔍 [DEBUG] Form data received:', Object.keys(formData));
        
        // Check APIConfig availability
        console.log('🔍 [DEBUG] Checking APIConfig availability...');
        console.log('🔍 [DEBUG] window.APIConfig exists:', typeof window.APIConfig !== 'undefined');
        console.log('🔍 [DEBUG] window.APIConfig value:', window.APIConfig);
        
        if (!window.APIConfig) {
            console.error('❌ [ERROR] APIConfig not available');
            console.error('❌ [ERROR] window object keys:', Object.keys(window).filter(k => k.includes('API') || k.includes('Config')));
            throw new Error('APIConfig not available');
        }
        
        console.log('✅ [DEBUG] APIConfig is available');
        
        try {
            const endpoints = APIConfig.getEndpoints();
            console.log('🔍 [DEBUG] APIConfig.getEndpoints() result:', endpoints);
            console.log('🔍 [DEBUG] Available endpoints:', Object.keys(endpoints));
            
            const apiUrl = APIConfig.getEndpoints().merchants;
            console.log('🔍 [DEBUG] Merchant API URL:', apiUrl);
            
            if (!apiUrl) {
                throw new Error('Merchant API endpoint not found in APIConfig');
            }
            
            const timeout = 15000; // 15 seconds
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), timeout);
            console.log('🔍 [DEBUG] Request timeout set to 15000ms');
            
            // Prepare merchant data for API
            console.log('🔍 [DEBUG] Preparing merchant request data...');
            const merchantRequest = {
            name: formData.businessName,
            legal_name: formData.businessName, // Use business name as legal name if not provided
            registration_number: formData.registrationNumber || '',
            tax_id: formData.registrationNumber || '', // Use registration number as tax ID if available
            industry: '', // Will be populated from business intelligence results if available
            industry_code: '',
            business_type: '',
            portfolio_type: 'prospective', // Default to prospective
            risk_level: 'medium', // Default risk level
            status: 'active',
            address: {
                street1: formData.streetAddress || '',
                street2: '',
                city: formData.city || '',
                state: formData.state || '',
                postal_code: formData.postalCode || '',
                country: formData.country || '',
                country_code: formData.country || ''
            },
            contact_info: {
                phone: formData.phoneNumber || '',
                email: formData.email || '',
                website: formData.websiteUrl || '',
                primary_contact: ''
            }
        };
        
            console.log('🔍 [DEBUG] Merchant request data prepared:', merchantRequest);
            
            try {
                console.log('💾 [DEBUG] Sending merchant data to API...');
                console.log('💾 [DEBUG] API URL:', apiUrl);
                console.log('💾 [DEBUG] Request method: POST');
                console.log('💾 [DEBUG] Request payload:', JSON.stringify(merchantRequest, null, 2));
                
                const response = await fetch(apiUrl, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Accept': 'application/json',
                    },
                    body: JSON.stringify(merchantRequest),
                    signal: controller.signal,
                    credentials: 'omit'
                });

                clearTimeout(timeoutId);
                console.log('🔍 [DEBUG] Fetch request completed');
                console.log('🔍 [DEBUG] Response status:', response.status);
                console.log('🔍 [DEBUG] Response statusText:', response.statusText);
                console.log('🔍 [DEBUG] Response headers:', Object.fromEntries(response.headers.entries()));

                if (!response.ok) {
                let errorText = 'Unknown error';
                try {
                    errorText = await response.text();
                } catch (e) {
                    // Ignore
                }
                
                const error = new Error(`Failed to save merchant: ${response.status} ${response.statusText}`);
                error.status = response.status;
                error.statusText = response.statusText;
                error.details = errorText;
                throw error;
            }

                const contentType = response.headers.get('content-type') || '';
                console.log('🔍 [DEBUG] Response content-type:', contentType);
                
                if (contentType.includes('application/json')) {
                    console.log('🔍 [DEBUG] Parsing JSON response...');
                    const merchant = await response.json();
                    console.log('✅ [DEBUG] Merchant saved successfully:', merchant);
                    console.log('✅ [DEBUG] Merchant ID:', merchant.id);
                    return merchant;
                } else {
                    console.warn('⚠️ [DEBUG] Unexpected content-type, attempting to read as text');
                    const text = await response.text();
                    console.error('❌ [ERROR] Response text:', text);
                    throw new Error(`Expected JSON but received ${contentType}`);
                }
            } catch (fetchError) {
                clearTimeout(timeoutId);
                console.error('❌ [ERROR] Error in fetch request:', fetchError);
                console.error('❌ [ERROR] Error name:', fetchError.name);
                console.error('❌ [ERROR] Error message:', fetchError.message);
                console.error('❌ [ERROR] Error stack:', fetchError.stack);
                
                if (fetchError.name === 'AbortError') {
                    console.error('❌ [ERROR] Request was aborted (timeout)');
                    throw new Error('Merchant save request timed out');
                }
                
                if (fetchError.name === 'TypeError' && fetchError.message.includes('fetch')) {
                    console.error('❌ [ERROR] Network error detected');
                    throw new Error('Network error: Unable to reach merchant service. Please check your connection.');
                }
                
                throw fetchError;
            }
        } catch (error) {
            console.error('❌ [ERROR] Error in saveMerchantToPortfolio:', error);
            console.error('❌ [ERROR] Error name:', error.name);
            console.error('❌ [ERROR] Error message:', error.message);
            console.error('❌ [ERROR] Error stack:', error.stack);
            throw error;
        }
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

