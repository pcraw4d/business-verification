'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FormField } from './FormField';
import { FormValidation, type ValidationErrors } from '@/lib/form-validation';
import { Building, Globe, MapPin, Phone, Mail, FileText, ChartLine, AlertTriangle } from 'lucide-react';
import { toast } from 'sonner';
import { createMerchant, type CreateMerchantRequest } from '@/lib/api';

interface MerchantFormData {
  businessName: string;
  websiteUrl: string;
  streetAddress: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
  phoneNumber: string;
  email: string;
  registrationNumber: string;
  analysisType: string;
  assessmentType: string;
}

const COUNTRIES = [
  { value: 'US', label: 'United States' },
  { value: 'CA', label: 'Canada' },
  { value: 'GB', label: 'United Kingdom' },
  { value: 'AU', label: 'Australia' },
  { value: 'DE', label: 'Germany' },
  { value: 'FR', label: 'France' },
  { value: 'IT', label: 'Italy' },
  { value: 'ES', label: 'Spain' },
  { value: 'NL', label: 'Netherlands' },
  { value: 'BE', label: 'Belgium' },
  { value: 'CH', label: 'Switzerland' },
  { value: 'AT', label: 'Austria' },
  { value: 'SE', label: 'Sweden' },
  { value: 'NO', label: 'Norway' },
  { value: 'DK', label: 'Denmark' },
  { value: 'FI', label: 'Finland' },
  { value: 'IE', label: 'Ireland' },
  { value: 'PT', label: 'Portugal' },
  { value: 'LU', label: 'Luxembourg' },
  { value: 'JP', label: 'Japan' },
  { value: 'KR', label: 'South Korea' },
  { value: 'SG', label: 'Singapore' },
  { value: 'HK', label: 'Hong Kong' },
  { value: 'NZ', label: 'New Zealand' },
  { value: 'BR', label: 'Brazil' },
  { value: 'MX', label: 'Mexico' },
  { value: 'AR', label: 'Argentina' },
  { value: 'CL', label: 'Chile' },
  { value: 'CO', label: 'Colombia' },
  { value: 'PE', label: 'Peru' },
  { value: 'ZA', label: 'South Africa' },
  { value: 'EG', label: 'Egypt' },
  { value: 'NG', label: 'Nigeria' },
  { value: 'KE', label: 'Kenya' },
  { value: 'MA', label: 'Morocco' },
  { value: 'TN', label: 'Tunisia' },
  { value: 'IN', label: 'India' },
  { value: 'CN', label: 'China' },
  { value: 'TH', label: 'Thailand' },
  { value: 'MY', label: 'Malaysia' },
  { value: 'ID', label: 'Indonesia' },
  { value: 'PH', label: 'Philippines' },
  { value: 'VN', label: 'Vietnam' },
  { value: 'TW', label: 'Taiwan' },
  { value: 'IL', label: 'Israel' },
  { value: 'AE', label: 'United Arab Emirates' },
  { value: 'SA', label: 'Saudi Arabia' },
  { value: 'TR', label: 'Turkey' },
  { value: 'RU', label: 'Russia' },
  { value: 'PL', label: 'Poland' },
  { value: 'CZ', label: 'Czech Republic' },
  { value: 'HU', label: 'Hungary' },
  { value: 'RO', label: 'Romania' },
  { value: 'BG', label: 'Bulgaria' },
  { value: 'HR', label: 'Croatia' },
  { value: 'SI', label: 'Slovenia' },
  { value: 'SK', label: 'Slovakia' },
  { value: 'LT', label: 'Lithuania' },
  { value: 'LV', label: 'Latvia' },
  { value: 'EE', label: 'Estonia' },
  { value: 'GR', label: 'Greece' },
  { value: 'CY', label: 'Cyprus' },
  { value: 'MT', label: 'Malta' },
];

const ANALYSIS_TYPES = [
  { value: 'comprehensive', label: 'Comprehensive Analysis' },
  { value: 'basic', label: 'Basic Classification' },
  { value: 'risk', label: 'Risk Assessment' },
  { value: 'compliance', label: 'Compliance Check' },
];

const ASSESSMENT_TYPES = [
  { value: 'comprehensive', label: 'Comprehensive Assessment' },
  { value: 'financial', label: 'Financial Risk Only' },
  { value: 'operational', label: 'Operational Risk Only' },
  { value: 'regulatory', label: 'Regulatory Risk Only' },
  { value: 'reputational', label: 'Reputational Risk Only' },
  { value: 'cybersecurity', label: 'Cybersecurity Risk Only' },
];

const validationRules = {
  businessName: { required: true, minLength: 2, maxLength: 255 },
  websiteUrl: { url: true },
  streetAddress: { maxLength: 255 },
  city: { maxLength: 100 },
  state: { maxLength: 100 },
  postalCode: { maxLength: 20 },
  country: { required: true },
  phoneNumber: { phone: true },
  email: { email: true },
  registrationNumber: { maxLength: 100 },
  analysisType: {},
  assessmentType: {},
};

export function MerchantForm() {
  const router = useRouter();
  const [formData, setFormData] = useState<MerchantFormData>({
    businessName: '',
    websiteUrl: '',
    streetAddress: '',
    city: '',
    state: '',
    postalCode: '',
    country: '',
    phoneNumber: '',
    email: '',
    registrationNumber: '',
    analysisType: 'comprehensive',
    assessmentType: 'comprehensive',
  });
  const [errors, setErrors] = useState<ValidationErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Persist form data to sessionStorage to prevent data loss
  useEffect(() => {
    const savedFormData = sessionStorage.getItem('merchantFormData');
    if (savedFormData) {
      try {
        const parsed = JSON.parse(savedFormData);
        setFormData((prev) => ({ ...prev, ...parsed }));
      } catch (error) {
        console.error('Error restoring form data:', error);
      }
    }
  }, []);

  // Save form data to sessionStorage whenever it changes
  useEffect(() => {
    if (typeof window !== 'undefined' && Object.keys(formData).length > 0) {
      sessionStorage.setItem('merchantFormData', JSON.stringify(formData));
    }
  }, [formData]);

  const updateField = (field: keyof MerchantFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    
    // Re-validate the field immediately when it's updated
    const fieldRule = validationRules[field];
    if (fieldRule) {
      const formDataRecord: Record<string, string> = {
        [field]: value,
      };
      const fieldErrors = FormValidation.validateForm(formDataRecord, { [field]: fieldRule });
      
      setErrors((prev) => {
        const newErrors = { ...prev };
        if (fieldErrors[field]) {
          newErrors[field] = fieldErrors[field];
        } else {
          // Clear error if field is now valid
          delete newErrors[field];
        }
        return newErrors;
      });
    } else {
      // If no validation rule, just clear any existing error
      if (errors[field]) {
        setErrors((prev) => {
          const newErrors = { ...prev };
          delete newErrors[field];
          return newErrors;
        });
      }
    }
  };

  const validateForm = (): boolean => {
    const formDataRecord: Record<string, string> = {
      businessName: formData.businessName,
      websiteUrl: formData.websiteUrl || '',
      streetAddress: formData.streetAddress || '',
      city: formData.city || '',
      state: formData.state || '',
      postalCode: formData.postalCode || '',
      country: formData.country,
      phoneNumber: formData.phoneNumber || '',
      email: formData.email || '',
      registrationNumber: formData.registrationNumber || '',
      analysisType: formData.analysisType || '',
      assessmentType: formData.assessmentType || '',
    };
    const newErrors = FormValidation.validateForm(formDataRecord, validationRules);
    setErrors(newErrors);
    return !FormValidation.hasErrors(newErrors);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      toast.error('Please fix the errors in the form');
      return;
    }

    setIsSubmitting(true);

    try {
      // Build address string
      const addressParts = [
        formData.streetAddress,
        formData.city,
        formData.state,
        formData.postalCode,
        formData.country,
      ].filter(Boolean);
      const address = addressParts.join(', ');

      // Prepare merchant data
      const merchantData: CreateMerchantRequest = {
        name: formData.businessName,
        legal_name: formData.businessName, // Use business name as legal name if not provided
        website: formData.websiteUrl || undefined,
        address: {
          street: formData.streetAddress || undefined,
          city: formData.city || undefined,
          state: formData.state || undefined,
          postal_code: formData.postalCode || undefined,
          country: formData.country,
        },
        contact_info: {
          phone: formData.phoneNumber || undefined,
          email: formData.email || undefined,
        },
        registration_number: formData.registrationNumber || undefined,
        country: formData.country,
      };

      // Create merchant via API
      const result = await createMerchant(merchantData);

      // Store data in sessionStorage for merchant-details page (for backward compatibility)
      if (typeof window !== 'undefined') {
        sessionStorage.setItem('merchantData', JSON.stringify(merchantData));
        sessionStorage.setItem('merchantId', result.id);
        // Clear form data after successful submission
        sessionStorage.removeItem('merchantFormData');
      }

      toast.success('Merchant created successfully', {
        description: `Merchant ID: ${result.id}`,
      });
      
      // Redirect to merchant details page with the actual merchant ID
      router.push(`/merchant-details/${result.id}`);
    } catch (error) {
      console.error('Error submitting merchant form:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to verify merchant');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleClear = () => {
    setFormData({
      businessName: '',
      websiteUrl: '',
      streetAddress: '',
      city: '',
      state: '',
      postalCode: '',
      country: '',
      phoneNumber: '',
      email: '',
      registrationNumber: '',
      analysisType: 'comprehensive',
      assessmentType: 'comprehensive',
    });
    setErrors({});
    // Clear persisted form data
    if (typeof window !== 'undefined') {
      sessionStorage.removeItem('merchantFormData');
    }
    toast.info('Form cleared');
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Building className="h-5 w-5" />
          Merchant Information
        </CardTitle>
        <CardDescription>
          Enter merchant information to perform comprehensive business verification, risk assessment, and analytics analysis.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Information */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <FormField
              label="Business Name"
              name="businessName"
              type="text"
              placeholder="Enter business name..."
              required
              value={formData.businessName}
              onChange={(value) => updateField('businessName', value)}
              error={errors.businessName}
              icon={<Building className="h-4 w-4" />}
            />
            
            <FormField
              label="Website URL"
              name="websiteUrl"
              type="url"
              placeholder="https://example.com"
              value={formData.websiteUrl}
              onChange={(value) => updateField('websiteUrl', value)}
              error={errors.websiteUrl}
              icon={<Globe className="h-4 w-4" />}
            />
          </div>

          {/* Address Information */}
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <FormField
                label="Street Address"
                name="streetAddress"
                type="text"
                placeholder="123 Main Street"
                value={formData.streetAddress}
                onChange={(value) => updateField('streetAddress', value)}
                error={errors.streetAddress}
              />
              
              <FormField
                label="City"
                name="city"
                type="text"
                placeholder="New York"
                value={formData.city}
                onChange={(value) => updateField('city', value)}
                error={errors.city}
              />
              
              <FormField
                label="State/Province"
                name="state"
                type="text"
                placeholder="NY"
                value={formData.state}
                onChange={(value) => updateField('state', value)}
                error={errors.state}
              />
              
              <FormField
                label="Postal Code"
                name="postalCode"
                type="text"
                placeholder="10001"
                value={formData.postalCode}
                onChange={(value) => updateField('postalCode', value)}
                error={errors.postalCode}
              />
              
              <FormField
                label="Country"
                name="country"
                type="select"
                required
                value={formData.country}
                onChange={(value) => updateField('country', value)}
                error={errors.country}
                selectOptions={COUNTRIES}
              />
            </div>
          </div>

          {/* Contact Information */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <FormField
              label="Phone Number"
              name="phoneNumber"
              type="tel"
              placeholder="+1 (555) 123-4567"
              value={formData.phoneNumber}
              onChange={(value) => updateField('phoneNumber', value)}
              error={errors.phoneNumber}
              icon={<Phone className="h-4 w-4" />}
            />
            
            <FormField
              label="Email Address"
              name="email"
              type="email"
              placeholder="contact@business.com"
              value={formData.email}
              onChange={(value) => updateField('email', value)}
              error={errors.email}
              icon={<Mail className="h-4 w-4" />}
            />
          </div>

          {/* Business Registration */}
          <FormField
            label="Business Registration Number"
            name="registrationNumber"
            type="text"
            placeholder="Enter business registration number (EIN, VAT, etc.)"
            value={formData.registrationNumber}
            onChange={(value) => updateField('registrationNumber', value)}
            error={errors.registrationNumber}
            icon={<FileText className="h-4 w-4" />}
          />

          {/* Analysis Configuration */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <FormField
              label="Analysis Type"
              name="analysisType"
              type="select"
              value={formData.analysisType}
              onChange={(value) => updateField('analysisType', value)}
              error={errors.analysisType}
              selectOptions={ANALYSIS_TYPES}
              icon={<ChartLine className="h-4 w-4" />}
            />
            
            <FormField
              label="Risk Assessment Type"
              name="assessmentType"
              type="select"
              value={formData.assessmentType}
              onChange={(value) => updateField('assessmentType', value)}
              error={errors.assessmentType}
              selectOptions={ASSESSMENT_TYPES}
              icon={<AlertTriangle className="h-4 w-4" />}
            />
          </div>

          {/* Form Actions */}
          <div className="flex justify-end gap-4 pt-4 border-t">
            <Button
              type="button"
              variant="outline"
              onClick={handleClear}
              disabled={isSubmitting}
              aria-label="Clear all form fields"
            >
              Clear Form
            </Button>
            <Button type="submit" disabled={isSubmitting} aria-label={isSubmitting ? 'Processing merchant verification' : 'Submit merchant verification'}>
              {isSubmitting ? 'Processing...' : 'Verify Merchant'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

