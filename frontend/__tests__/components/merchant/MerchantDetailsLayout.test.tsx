import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { render, screen, waitFor } from '@testing-library/react';
import { MerchantDetailsLayout } from '@/components/merchant/MerchantDetailsLayout';
import * as api from '@/lib/api';

// Mock API
jest.mock('@/lib/api', () => ({
  getMerchant: jest.fn(),
}));

describe('MerchantDetailsLayout', () => {
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
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render loading state initially', () => {
    (api.getMerchant as jest.Mock).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    expect(screen.getByRole('status', { hidden: true })).toBeInTheDocument();
  });

  it('should render merchant details when loaded', async () => {
    (api.getMerchant as jest.Mock).mockResolvedValue(mockMerchant);

    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText('Test Business')).toBeInTheDocument();
    });

    expect(screen.getByText('Technology')).toBeInTheDocument();
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('should render error state on API error', async () => {
    (api.getMerchant as jest.Mock).mockRejectedValue(new Error('Failed to load'));

    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load/i)).toBeInTheDocument();
    });
  });

  it('should render tabs correctly', async () => {
    (api.getMerchant as jest.Mock).mockResolvedValue(mockMerchant);

    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText('Overview')).toBeInTheDocument();
      expect(screen.getByText('Business Analytics')).toBeInTheDocument();
      expect(screen.getByText('Risk Assessment')).toBeInTheDocument();
      expect(screen.getByText('Risk Indicators')).toBeInTheDocument();
    });
  });

  it('should call getMerchant with correct merchantId', async () => {
    (api.getMerchant as jest.Mock).mockResolvedValue(mockMerchant);

    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    await waitFor(() => {
      expect(api.getMerchant).toHaveBeenCalledWith('merchant-123');
    });
  });
});

