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
      <div data-testid="founded-date-enriched">{isFieldEnriched(merchantId, 'foundedDate') ? 'Yes' : 'No'}</div>
      <div data-testid="revenue-enriched">{isFieldEnriched(merchantId, 'annualRevenue') ? 'Yes' : 'No'}</div>
      <div data-testid="field-info">
        {getEnrichedFieldInfo(merchantId, 'foundedDate')?.source || 'None'}
      </div>
    </div>
  );
}

describe('EnrichmentContext', () => {
  const merchantId = 'merchant-123';

  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
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
      vi.useFakeTimers();
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      
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

      // Fast-forward time beyond highlight duration (5 minutes)
      vi.advanceTimersByTime(6 * 60 * 1000);

      await waitFor(() => {
        // Field should no longer be enriched (expired)
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('No');
      });

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
            fieldName: 'foundedDate',
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

      await waitFor(() => {
        expect(screen.getByTestId('founded-date-enriched')).toHaveTextContent('Yes');
      });
    });
  });
});

