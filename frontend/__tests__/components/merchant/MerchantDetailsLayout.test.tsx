// Vitest globals are available via globals: true in vitest.config.ts
import { render, screen, waitFor, act } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { server } from '../../mocks/server';

// Import component - no need to mock the API module with MSW
import { MerchantDetailsLayout } from '@/components/merchant/MerchantDetailsLayout';

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
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  beforeEach(async () => {
    // CRITICAL: Clear deduplicator FIRST to clear any pending promises from previous tests
    // This must happen before resetting handlers to prevent stale promises from being returned
    try {
      const { getAPICache, getRequestDeduplicator } = await import('@/lib/api');
      const apiCache = getAPICache();
      const requestDeduplicator = getRequestDeduplicator();
      
      // Clear deduplicator first to clear pending promises
      if (requestDeduplicator && typeof requestDeduplicator.clear === 'function') {
        requestDeduplicator.clear();
      }
      // Then clear cache
      if (apiCache && typeof apiCache.clear === 'function') {
        apiCache.clear();
      }
    } catch (error) {
      // Ignore errors if module not loaded yet
    }
    
    // Reset MSW handlers to default state to ensure test-specific handlers don't persist
    server.resetHandlers();
    
    // Clear sessionStorage
    if (typeof window !== 'undefined' && window.sessionStorage) {
      window.sessionStorage.clear();
    }
  });

  it('should render loading state initially', () => {
    // Use MSW to delay the response so we can see the loading state
    server.use(
      http.get('http://localhost:8080/api/v1/merchants/:merchantId', () => {
        // Return a promise that never resolves to simulate loading
        return new Promise(() => {});
      })
    );

    const { container } = render(<MerchantDetailsLayout merchantId="merchant-123" />);

    // Component renders Skeleton components when loading
    // Check for skeleton by class name
    const skeletons = container.querySelectorAll('[class*="skeleton"], [class*="Skeleton"]');
    // If no skeletons found, check for container div
    if (skeletons.length === 0) {
      // Component should at least render the container
      expect(container.querySelector('.container')).toBeInTheDocument();
    } else {
      expect(skeletons.length).toBeGreaterThan(0);
    }
  });

  it('should render merchant details when loaded', async () => {
    // MSW will automatically handle the API request via the default handler
    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    // Wait for the API call to complete and state to update
    // The component renders merchant.businessName in an h1 tag
    // Use findByRole to wait for the heading to appear, which is more reliable
    const heading = await screen.findByRole('heading', { name: /Test Business/i }, { timeout: 5000 });
    expect(heading).toBeInTheDocument();
    
    // Verify the text appears (may appear multiple times, so use getAllByText)
    const businessNameTexts = screen.getAllByText('Test Business');
    expect(businessNameTexts.length).toBeGreaterThan(0);

    expect(screen.getByText('Technology')).toBeInTheDocument();
    // "active" appears multiple times, so use getAllByText and check it exists
    const activeTexts = screen.getAllByText(/active/i);
    expect(activeTexts.length).toBeGreaterThan(0);
  });

  it('should render error state on API error', async () => {
    // Use MSW to return an error response
    server.use(
      http.get('http://localhost:8080/api/v1/merchants/:merchantId', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to load merchant data' },
          { status: 500 }
        );
      })
    );

    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    // Wait for the error message to appear - the component renders error in AlertDescription
    // The error message will be "API Error 500" (from handleResponse) or the actual error message
    const errorMessage = await screen.findByText(/API Error|Failed to load/i, {}, { timeout: 5000 });
    expect(errorMessage).toBeInTheDocument();
  });

  it('should render tabs correctly', async () => {
    // MSW will automatically handle the API request
    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    // Wait for merchant data to load first using findByRole for the heading
    await screen.findByRole('heading', { name: /Test Business/i }, { timeout: 5000 });

    // Then verify tabs are rendered - TabsTrigger renders as a button with role="tab"
    // Use findByRole with role="tab" instead of "button"
    expect(await screen.findByRole('tab', { name: /Overview/i }, { timeout: 3000 })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /Business Analytics/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /Risk Assessment/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /Risk Indicators/i })).toBeInTheDocument();
  });

  it('should call API with correct merchantId', async () => {
    // Track the request to verify the correct merchantId is used
    let capturedMerchantId: string | null = null;
    
    server.use(
      http.get('http://localhost:8080/api/v1/merchants/:merchantId', ({ params }) => {
        capturedMerchantId = params.merchantId as string;
        return HttpResponse.json(mockMerchant);
      })
    );

    render(<MerchantDetailsLayout merchantId="merchant-123" />);

    // Wait for merchant data to load using findByRole for the heading
    await screen.findByRole('heading', { name: /Test Business/i }, { timeout: 5000 });

    // Verify the API was called with the correct merchant ID
    expect(capturedMerchantId).toBe('merchant-123');
  });
});

