import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { render, screen, waitFor } from '@testing-library/react';
import { MerchantOverviewTab } from '@/components/merchant/MerchantOverviewTab';

describe('MerchantOverviewTab', () => {
  const mockMerchant = {
    id: 'merchant-123',
    businessName: 'Test Business',
    industry: 'Technology',
    status: 'active',
    email: 'test@example.com',
    phone: '+1-555-123-4567',
    website: 'https://test.com',
    address: {
      street: '123 Main St',
      city: 'San Francisco',
      state: 'CA',
      postalCode: '94102',
      country: 'USA',
    },
    registrationNumber: 'REG-123',
    taxId: 'TAX-456',
    foundedYear: 2020,
    employeeCount: 50,
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render merchant data when provided', () => {
    render(<MerchantOverviewTab merchant={mockMerchant} />);

    expect(screen.getByText('Test Business')).toBeInTheDocument();
    expect(screen.getByText('Technology')).toBeInTheDocument();
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('should render contact information', () => {
    render(<MerchantOverviewTab merchant={mockMerchant} />);

    expect(screen.getByText('test@example.com')).toBeInTheDocument();
    expect(screen.getByText('+1-555-123-4567')).toBeInTheDocument();
    expect(screen.getByText('https://test.com')).toBeInTheDocument();
  });

  it('should render business details', () => {
    render(<MerchantOverviewTab merchant={mockMerchant} />);

    expect(screen.getByText(/123 Main St/i)).toBeInTheDocument();
    expect(screen.getByText(/San Francisco/i)).toBeInTheDocument();
    expect(screen.getByText(/REG-123/i)).toBeInTheDocument();
  });

  it('should render status badge', () => {
    render(<MerchantOverviewTab merchant={mockMerchant} />);

    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('should handle missing optional fields', () => {
    const minimalMerchant = {
      id: 'merchant-123',
      businessName: 'Test Business',
      status: 'active',
    };

    render(<MerchantOverviewTab merchant={minimalMerchant} />);

    expect(screen.getByText('Test Business')).toBeInTheDocument();
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('should render loading state when merchant is null', () => {
    render(<MerchantOverviewTab merchant={null} loading={true} />);

    expect(screen.getByRole('status', { hidden: true })).toBeInTheDocument();
  });

  it('should render error state when error provided', () => {
    render(<MerchantOverviewTab merchant={null} error="Failed to load" />);

    expect(screen.getByText(/Failed to load/i)).toBeInTheDocument();
  });
});

