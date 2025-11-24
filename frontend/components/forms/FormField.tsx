'use client';

import { ReactNode } from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { cn } from '@/lib/utils';

interface FormFieldProps {
  label: string;
  name: string;
  type?: 'text' | 'email' | 'tel' | 'url' | 'textarea' | 'select';
  placeholder?: string;
  required?: boolean;
  error?: string;
  value?: string;
  onChange?: (value: string) => void;
  className?: string;
  children?: ReactNode;
  selectOptions?: { value: string; label: string }[];
  icon?: ReactNode;
}

export function FormField({
  label,
  name,
  type = 'text',
  placeholder,
  required = false,
  error,
  value,
  onChange,
  className,
  children,
  selectOptions,
  icon,
}: FormFieldProps) {
  const fieldId = `field-${name}`;
  const errorId = `error-${name}`;

  const renderInput = () => {
    switch (type) {
      case 'textarea':
        return (
          <Textarea
            id={fieldId}
            name={name}
            placeholder={placeholder}
            required={required}
            value={value}
            onChange={(e) => onChange?.(e.target.value)}
            className={cn(error && 'border-destructive')}
            aria-invalid={!!error}
            aria-describedby={error ? errorId : undefined}
          />
        );
      
      case 'select':
        return (
          <Select
            value={value}
            onValueChange={onChange}
            required={required}
          >
            <SelectTrigger
              id={fieldId}
              name={name}
              className={cn('w-full', error && 'border-destructive')}
              aria-invalid={!!error}
              aria-describedby={error ? errorId : undefined}
            >
              <SelectValue placeholder={placeholder} />
            </SelectTrigger>
            <SelectContent className="z-50" position="popper">
              {selectOptions?.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        );
      
      default:
        return (
          <Input
            id={fieldId}
            name={name}
            type={type}
            placeholder={placeholder}
            required={required}
            value={value}
            onChange={(e) => onChange?.(e.target.value)}
            className={cn(error && 'border-destructive')}
            aria-invalid={!!error}
            aria-describedby={error ? errorId : undefined}
          />
        );
    }
  };

  return (
    <div className={cn('space-y-2', className)}>
      <Label htmlFor={fieldId} className={cn('flex items-center gap-2', required && 'after:content-["*"] after:text-destructive after:ml-1')}>
        {icon}
        {label}
      </Label>
      {children || renderInput()}
      {error && (
        <p id={errorId} className="text-sm text-destructive flex items-center gap-1" role="alert">
          <span className="sr-only">Error:</span>
          {error}
        </p>
      )}
    </div>
  );
}

