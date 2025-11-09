/**
 * User Registration Form Handler
 * Handles user registration form submission and validation
 */
class RegisterForm {
    constructor() {
        this.form = document.getElementById('registerForm');
        this.submitBtn = document.getElementById('submitBtn');
        this.loading = document.getElementById('loading');
        this.successMessage = document.getElementById('successMessage');
        
        this.init();
    }

    /**
     * Initialize the registration form
     */
    init() {
        if (!this.form) {
            console.error('Registration form not found');
            return;
        }

        // Bind form submission
        this.form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleSubmit();
        });

        // Bind password validation
        const passwordInput = document.getElementById('password');
        if (passwordInput) {
            passwordInput.addEventListener('input', () => {
                this.validatePassword();
            });
        }

        // Bind confirm password validation
        const confirmPasswordInput = document.getElementById('confirmPassword');
        if (confirmPasswordInput) {
            confirmPasswordInput.addEventListener('input', () => {
                this.validateConfirmPassword();
            });
        }

        // Real-time validation
        const inputs = this.form.querySelectorAll('input');
        inputs.forEach(input => {
            input.addEventListener('blur', () => {
                this.validateField(input);
            });
        });
    }

    /**
     * Handle form submission
     */
    async handleSubmit() {
        // Clear previous errors
        this.clearErrors();

        // Validate all fields
        if (!this.validateForm()) {
            return;
        }

        // Get form data
        const formData = this.getFormData();

        // Show loading state
        this.setLoading(true);

        try {
            // Submit registration
            const response = await fetch('/api/v1/auth/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || data.error || 'Registration failed');
            }

            // Show success message
            this.showSuccess();

            // Redirect to login after 3 seconds
            setTimeout(() => {
                window.location.href = '/dashboard';
            }, 3000);

        } catch (error) {
            console.error('Registration error:', error);
            this.showError(error.message || 'Registration failed. Please try again.');
        } finally {
            this.setLoading(false);
        }
    }

    /**
     * Get form data
     */
    getFormData() {
        return {
            email: document.getElementById('email').value.trim(),
            username: document.getElementById('username').value.trim(),
            password: document.getElementById('password').value,
            first_name: document.getElementById('firstName').value.trim(),
            last_name: document.getElementById('lastName').value.trim(),
            company: document.getElementById('company').value.trim()
        };
    }

    /**
     * Validate the entire form
     */
    validateForm() {
        let isValid = true;

        // Validate each field
        const fields = [
            'email', 'username', 'password', 'confirmPassword',
            'firstName', 'lastName', 'company'
        ];

        fields.forEach(fieldId => {
            const field = document.getElementById(fieldId);
            if (field && !this.validateField(field)) {
                isValid = false;
            }
        });

        return isValid;
    }

    /**
     * Validate a single field
     */
    validateField(field) {
        const fieldId = field.id;
        const value = field.value.trim();
        let isValid = true;
        let errorMessage = '';

        switch (fieldId) {
            case 'email':
                if (!value) {
                    errorMessage = 'Email is required';
                    isValid = false;
                } else if (!this.isValidEmail(value)) {
                    errorMessage = 'Please enter a valid email address';
                    isValid = false;
                }
                break;

            case 'username':
                if (!value) {
                    errorMessage = 'Username is required';
                    isValid = false;
                } else if (value.length < 3) {
                    errorMessage = 'Username must be at least 3 characters';
                    isValid = false;
                } else if (value.length > 50) {
                    errorMessage = 'Username must be less than 50 characters';
                    isValid = false;
                }
                break;

            case 'password':
                if (!value) {
                    errorMessage = 'Password is required';
                    isValid = false;
                } else if (!this.isValidPassword(value)) {
                    errorMessage = 'Password does not meet requirements';
                    isValid = false;
                }
                break;

            case 'confirmPassword':
                const password = document.getElementById('password').value;
                if (!value) {
                    errorMessage = 'Please confirm your password';
                    isValid = false;
                } else if (value !== password) {
                    errorMessage = 'Passwords do not match';
                    isValid = false;
                }
                break;

            case 'firstName':
            case 'lastName':
                if (!value) {
                    errorMessage = `${this.getFieldLabel(fieldId)} is required`;
                    isValid = false;
                }
                break;

            case 'company':
                if (!value) {
                    errorMessage = 'Company name is required';
                    isValid = false;
                }
                break;
        }

        // Show/hide error
        this.showFieldError(fieldId, errorMessage);

        // Update field styling
        if (isValid) {
            field.classList.remove('error');
        } else {
            field.classList.add('error');
        }

        return isValid;
    }

    /**
     * Validate password
     */
    validatePassword() {
        const password = document.getElementById('password').value;
        const requirements = {
            length: password.length >= 8,
            uppercase: /[A-Z]/.test(password),
            lowercase: /[a-z]/.test(password),
            number: /[0-9]/.test(password)
        };

        // Update requirement indicators
        Object.keys(requirements).forEach(req => {
            const element = document.getElementById(`req-${req}`);
            if (element) {
                if (requirements[req]) {
                    element.classList.add('valid');
                } else {
                    element.classList.remove('valid');
                }
            }
        });

        return Object.values(requirements).every(v => v);
    }

    /**
     * Validate confirm password
     */
    validateConfirmPassword() {
        const password = document.getElementById('password').value;
        const confirmPassword = document.getElementById('confirmPassword').value;

        if (confirmPassword && confirmPassword !== password) {
            this.showFieldError('confirmPassword', 'Passwords do not match');
            document.getElementById('confirmPassword').classList.add('error');
            return false;
        } else {
            this.showFieldError('confirmPassword', '');
            document.getElementById('confirmPassword').classList.remove('error');
            return true;
        }
    }

    /**
     * Check if email is valid
     */
    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    /**
     * Check if password is valid
     */
    isValidPassword(password) {
        return password.length >= 8 &&
               /[A-Z]/.test(password) &&
               /[a-z]/.test(password) &&
               /[0-9]/.test(password);
    }

    /**
     * Get field label
     */
    getFieldLabel(fieldId) {
        const labels = {
            'firstName': 'First name',
            'lastName': 'Last name'
        };
        return labels[fieldId] || fieldId;
    }

    /**
     * Show field error
     */
    showFieldError(fieldId, message) {
        const errorElement = document.getElementById(`${fieldId}Error`);
        if (errorElement) {
            if (message) {
                errorElement.textContent = message;
                errorElement.classList.add('show');
            } else {
                errorElement.textContent = '';
                errorElement.classList.remove('show');
            }
        }
    }

    /**
     * Clear all errors
     */
    clearErrors() {
        const errorElements = document.querySelectorAll('.error-message');
        errorElements.forEach(el => {
            el.textContent = '';
            el.classList.remove('show');
        });

        const inputs = this.form.querySelectorAll('input');
        inputs.forEach(input => {
            input.classList.remove('error');
        });
    }

    /**
     * Show error message
     */
    showError(message) {
        // Show error in first field or create a general error
        const emailError = document.getElementById('emailError');
        if (emailError) {
            emailError.textContent = message;
            emailError.classList.add('show');
        }
    }

    /**
     * Show success message
     */
    showSuccess() {
        if (this.successMessage) {
            this.successMessage.classList.add('show');
        }
        this.form.style.display = 'none';
    }

    /**
     * Set loading state
     */
    setLoading(loading) {
        if (this.submitBtn) {
            this.submitBtn.disabled = loading;
            if (loading) {
                this.submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Creating Account...';
            } else {
                this.submitBtn.innerHTML = '<i class="fas fa-user-plus"></i> <span>Create Account</span>';
            }
        }

        if (this.loading) {
            if (loading) {
                this.loading.classList.add('show');
            } else {
                this.loading.classList.remove('show');
            }
        }
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new RegisterForm();
});

