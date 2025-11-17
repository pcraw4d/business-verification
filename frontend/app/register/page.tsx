'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { FormField } from '@/components/forms/FormField';
import { FormValidation } from '@/lib/form-validation';
import { ApiEndpoints } from '@/lib/api-config';
import { toast } from 'sonner';
import { UserPlus, Mail, Lock, User } from 'lucide-react';
import Link from 'next/link';

interface RegisterFormData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

const validationRules = {
  username: { required: true, minLength: 3, maxLength: 50 },
  email: { required: true, email: true },
  password: { required: true, minLength: 8 },
  confirmPassword: { required: true },
};

export default function RegisterPage() {
  const router = useRouter();
  const [formData, setFormData] = useState<RegisterFormData>({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const updateField = (field: keyof RegisterFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Validate all fields
    const formDataRecord: Record<string, string> = {
      username: formData.username,
      email: formData.email,
      password: formData.password,
      confirmPassword: formData.confirmPassword,
    };
    const baseErrors = FormValidation.validateForm(formDataRecord, validationRules);
    Object.assign(newErrors, baseErrors);

    // Custom validation for password confirmation
    if (formData.password && formData.confirmPassword && formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      toast.error('Please fix the errors in the form');
      return;
    }

    setIsSubmitting(true);

    try {
      const response = await fetch(ApiEndpoints.auth.register(), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: formData.username,
          email: formData.email,
          password: formData.password,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || 'Registration failed');
      }

      const data = await response.json();
      toast.success('Account created successfully!');
      
      // Redirect to login or dashboard
      router.push('/');
    } catch (error) {
      console.error('Registration error:', error);
      toast.error(error instanceof Error ? error.message : 'Registration failed');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AppLayout
      title="Create Account"
      description="Join KYB Platform to get started"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Register' },
      ]}
    >
      <div className="flex items-center justify-center min-h-[60vh]">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="flex items-center justify-center mb-2">
              <UserPlus className="h-8 w-8 text-primary" />
            </div>
            <CardTitle>Create Account</CardTitle>
            <CardDescription>Join KYB Platform to get started</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <FormField
                label="Username"
                name="username"
                type="text"
                placeholder="Enter username"
                required
                value={formData.username}
                onChange={(value) => updateField('username', value)}
                error={errors.username}
                icon={<User className="h-4 w-4" />}
              />

              <FormField
                label="Email"
                name="email"
                type="email"
                placeholder="Enter email address"
                required
                value={formData.email}
                onChange={(value) => updateField('email', value)}
                error={errors.email}
                icon={<Mail className="h-4 w-4" />}
              />

              <FormField
                label="Password"
                name="password"
                type="text"
                placeholder="Enter password"
                required
                value={formData.password}
                onChange={(value) => updateField('password', value)}
                error={errors.password}
                icon={<Lock className="h-4 w-4" />}
              />

              <FormField
                label="Confirm Password"
                name="confirmPassword"
                type="text"
                placeholder="Confirm password"
                required
                value={formData.confirmPassword}
                onChange={(value) => updateField('confirmPassword', value)}
                error={errors.confirmPassword}
                icon={<Lock className="h-4 w-4" />}
              />

              <Button type="submit" className="w-full" disabled={isSubmitting} aria-label={isSubmitting ? 'Creating account' : 'Create new account'}>
                {isSubmitting ? 'Creating Account...' : 'Create Account'}
              </Button>

              <div className="text-center text-sm text-muted-foreground pt-4 border-t">
                Already have an account?{' '}
                <Link href="/" className="text-primary hover:underline">
                  Sign in
                </Link>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

