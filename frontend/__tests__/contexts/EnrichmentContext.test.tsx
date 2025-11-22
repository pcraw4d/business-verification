import { EnrichmentProvider, useEnrichment } from '@/contexts/EnrichmentContext';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, beforeEach, vi } from 'vitest';

// Test component that uses the enrichment context
function TestComponent({ merchantId }: { merchantId: string }) {
  const { addEnrichedFields, isFieldEnriched, getEnrichedFieldInfo, clearEnrichedFields } = useEnrichment();

  return (
    <div>
      <button
        onClick={() => {
          addEnrichedFields(merchantId, [
            { name: 'Founded Date', type: 'added', source: 'Test Source' },
            { name: 'Annual Revenue', type: 'updated', source: 'Test Source' },
          ]);
        }}
      >
        Add Fields
      </button>
      <button onClick={() => clearEnrichedFields(merchantId)}>Clear Fields</button>
      <div data-testid="founded-date-enriched">{isFieldEnriched(merchantId, 'Founded Date') ? 'Yes' : 'No'}</div>
      <div data-testid="revenue-enriched">{isFieldEnriched(merchantId, 'Annual Revenue') ? 'Yes' : 'No'}</div>
      <div data-testid="field-info">
        {getEnrichedFieldInfo(merchantId, 'Founded Date')?.source || 'None'}
      </div>
    </div>
  );
}

describe('EnrichmentContext', () => {
  const merchantId = 'merchant-123';

  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
    // Clear any timers
    vi.clearAllTimers();
  });

  afterEach(() => {
    // Clean up timers after each test
    vi.useRealTimers();
  });

  describe('EnrichmentProvider', () => {
    it('should provide enrichment context to children', () => {
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      expect(screen.getByText('Add Fields')).toBeInTheDocument();
    });

    it('should initialize with empty enriched fields', () => {
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('No');
      expect(screen.getByTestId('revenue-enriched')).toHaveTextContent('No');
    });
  });

  describe('addEnrichedFields', () => {
    it('should add enriched fields to context', async () => {
      const user = userEvent.setup();
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      const addButton = screen.getByText('Add Fields');
      await user.click(addButton);

      await waitFor(() => {
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('Yes');
        expect(screen.getByTestId('revenue-enriched')).toHaveTextContent('Yes');
      });
    });

    it('should store enriched fields in localStorage', async () => {
      const user = userEvent.setup();
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      const addButton = screen.getByText('Add Fields');
      await user.click(addButton);

      await waitFor(() => {
        const stored = localStorage.getItem('enriched_fields');
        expect(stored).toBeTruthy();
        const parsed = JSON.parse(stored!);
        expect(parsed[merchantId]).toBeDefined();
        expect(parsed[merchantId].length).toBe(2);
      });
    });

    it('should update existing field if already enriched', async () => {
      const user = userEvent.setup();
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      const addButton = screen.getByText('Add Fields');
      await user.click(addButton);

      await waitFor(() => {
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('Yes');
      });

      // Add same field again
      await user.click(addButton);

      await waitFor(() => {
        // Should still be enriched (timestamp updated)
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('Yes');
      });
    });
  });

  describe('isFieldEnriched', () => {
    it('should return true for enriched fields', async () => {
      const user = userEvent.setup();
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      const addButton = screen.getByText('Add Fields');
      await user.click(addButton);

      await waitFor(() => {
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('Yes');
      });
    });

    it('should return false for non-enriched fields', () => {
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('No');
    });

    it('should return false for expired enriched fields', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      const addButton = screen.getByText('Add Fields');
      await user.click(addButton);

      // Wait for initial state
      await waitFor(() => {
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('Yes');
      }, { timeout: 2000 });

      // Fast-forward time beyond highlight duration (5 minutes = 300000ms)
      // Advance in smaller increments to trigger the interval cleanup
      vi.advanceTimersByTime(60 * 1000); // 1 minute
      await vi.runOnlyPendingTimersAsync();
      vi.advanceTimersByTime(60 * 1000); // Another minute
      await vi.runOnlyPendingTimersAsync();
      vi.advanceTimersByTime(4 * 60 * 1000); // 4 more minutes (total 6 minutes)

      // The interval runs every 60 seconds, so we need to let it process
      await vi.runOnlyPendingTimersAsync();

      await waitFor(() => {
        // Field should no longer be enriched (expired after 5 minutes)
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('No');
      }, { timeout: 3000 });

      vi.useRealTimers();
    });
  });

  describe('getEnrichedFieldInfo', () => {
    it('should return field info for enriched fields', async () => {
      const user = userEvent.setup();
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      const addButton = screen.getByText('Add Fields');
      await user.click(addButton);

      await waitFor(() => {
        expect(screen.getByTestId('field-info')).toHaveTextContent('Test Source');
      });
    });

    it('should return null for non-enriched fields', () => {
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      expect(screen.getByTestId('field-info')).toHaveTextContent('None');
    });
  });

  describe('clearEnrichedFields', () => {
    it('should clear enriched fields for a merchant', async () => {
      const user = userEvent.setup();
      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      const addButton = screen.getByText('Add Fields');
      await user.click(addButton);

      await waitFor(() => {
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('Yes');
      });

      const clearButton = screen.getByText('Clear Fields');
      await user.click(clearButton);

      await waitFor(() => {
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('No');
        expect(screen.getByTestId('revenue-enriched')).toHaveTextContent('No');
      });
    });
  });

  describe('localStorage Persistence', () => {
    it('should load enriched fields from localStorage on mount', async () => {
      // Pre-populate localStorage
      const enrichedData = {
        [merchantId]: [
          {
            fieldName: 'Founded Date', // Note: fieldName should match what's used in the component
            enrichedAt: new Date().toISOString(),
            source: 'Test Source',
            type: 'added',
          },
        ],
      };
      localStorage.setItem('enriched_fields', JSON.stringify(enrichedData));

      render(
        <EnrichmentProvider>
          <TestComponent merchantId={merchantId} />
        </EnrichmentProvider>
      );

      // Wait for the context to load from localStorage
      await waitFor(() => {
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('Yes');
      }, { timeout: 3000 });
    });
  });
});

