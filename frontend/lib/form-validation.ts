export interface ValidationRule {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
  custom?: (value: string) => string | null;
  email?: boolean;
  url?: boolean;
  phone?: boolean;
}

export interface ValidationErrors {
  [key: string]: string;
}

export class FormValidation {
  static validateEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  static validateURL(url: string): boolean {
    try {
      const urlObj = new URL(url);
      return urlObj.protocol === 'http:' || urlObj.protocol === 'https:';
    } catch {
      return false;
    }
  }

  static validatePhone(phone: string): boolean {
    // Basic phone validation - accepts international format
    const phoneRegex = /^[\+]?[(]?[0-9]{1,4}[)]?[-\s\.]?[(]?[0-9]{1,4}[)]?[-\s\.]?[0-9]{1,9}$/;
    return phoneRegex.test(phone.replace(/\s/g, ''));
  }

  static validateField(value: string, rules: ValidationRule): string | null {
    if (rules.required && (!value || value.trim().length === 0)) {
      return 'This field is required';
    }

    if (!value || value.trim().length === 0) {
      return null; // Optional field, no value is OK
    }

    if (rules.minLength && value.length < rules.minLength) {
      return `Must be at least ${rules.minLength} characters`;
    }

    if (rules.maxLength && value.length > rules.maxLength) {
      return `Must be no more than ${rules.maxLength} characters`;
    }

    if (rules.pattern && !rules.pattern.test(value)) {
      return 'Invalid format';
    }

    if (rules.email && !this.validateEmail(value)) {
      return 'Please enter a valid email address';
    }

    if (rules.url && !this.validateURL(value)) {
      return 'Please enter a valid URL';
    }

    if (rules.phone && !this.validatePhone(value)) {
      return 'Please enter a valid phone number';
    }

    if (rules.custom) {
      return rules.custom(value);
    }

    return null;
  }

  static validateForm(
    data: Record<string, string>,
    rules: Record<string, ValidationRule>
  ): ValidationErrors {
    const errors: ValidationErrors = {};

    for (const [field, fieldRules] of Object.entries(rules)) {
      const value = data[field] || '';
      const error = this.validateField(value, fieldRules);
      if (error) {
        errors[field] = error;
      }
    }

    return errors;
  }

  static hasErrors(errors: ValidationErrors): boolean {
    return Object.keys(errors).length > 0;
  }
}

