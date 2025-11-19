import { server } from '@/__tests__/mocks/server';
import { EnrichmentButton } from '@/components/merchant/EnrichmentButton';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { toast } from 'sonner';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('sonner');
const mockToast = vi.mocked(toast);

describe('EnrichmentButton', () => {
  const merchantId = 'merchant-123';

  const mockSources = [
    { id: 'source-1', name: 'Government Database', description: 'Official business registry data', enabled: true },
    { id: 'source-2', name: 'Credit Bureau', description: 'Credit and financial data', enabled: true },
    { id: 'source-3', name: 'Third Party API', description: 'External data provider', enabled: false },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.success = vi.fn();
    mockToast.error = vi.fn();
  });

  describe('Component Rendering', () => {
    it('should render enrichment button', () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: [] });
        })
      );

      render(<EnrichmentButton merchantId={merchantId} />);

      expect(screen.getByRole('button', { name: /enrich data/i })).toBeInTheDocument();
    });

    it('should display number of enabled sources in badge', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        // Should show badge with count of enabled sources (2 enabled, 1 disabled)
        expect(screen.getByText('2')).toBeInTheDocument();
      });
    });
  });

  describe('Dialog Interaction', () => {
    it('should open dialog when button is clicked', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Data Enrichment')).toBeInTheDocument();
        expect(screen.getByText('Select a data source to enrich merchant information')).toBeInTheDocument();
      });
    });

    it('should close dialog when clicking outside', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Data Enrichment')).toBeInTheDocument();
      });

      // Press Escape to close
      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.queryByText('Data Enrichment')).not.toBeInTheDocument();
      });
    });
  });

  describe('Loading Sources', () => {
    it('should show loading skeleton while fetching sources', async () => {
      let resolveSources: (value: any) => void;
      const sourcesPromise = new Promise((resolve) => {
        resolveSources = resolve;
      });

      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', async () => {
          await sourcesPromise;
          return HttpResponse.json({ merchantId, sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      // Should show skeleton while loading
      await waitFor(() => {
        const skeletons = document.querySelectorAll('[class*="skeleton"]');
        expect(skeletons.length).toBeGreaterThan(0);
      });

      // Resolve the promise
      resolveSources!({ merchantId, sources: mockSources });
    });

    it('should display enrichment sources when loaded', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Government Database')).toBeInTheDocument();
        expect(screen.getByText('Credit Bureau')).toBeInTheDocument();
        expect(screen.getByText('Official business registry data')).toBeInTheDocument();
      });
    });

    it('should only show enabled sources', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Government Database')).toBeInTheDocument();
        expect(screen.getByText('Credit Bureau')).toBeInTheDocument();
        // Disabled source should not be shown
        expect(screen.queryByText('Third Party API')).not.toBeInTheDocument();
      });
    });

    it('should show message when no sources available', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: [] });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText(/no enrichment sources available/i)).toBeInTheDocument();
      });
    });
  });

  describe('Triggering Enrichment', () => {
    it('should trigger enrichment when source is clicked', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
          return HttpResponse.json({ jobId: 'job-123', merchantId, source: 'source-1', status: 'pending', createdAt: new Date().toISOString() });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Government Database')).toBeInTheDocument();
      });

      const enrichButton = screen.getByRole('button', { name: /enrich/i });
      await user.click(enrichButton);

      await waitFor(() => {
        expect(mockToast.success).toHaveBeenCalledWith(
          'Enrichment job started',
          expect.objectContaining({
            description: expect.stringContaining('job-123'),
          })
        );
      });
    });

    it('should show processing state after triggering enrichment', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
          return HttpResponse.json({ jobId: 'job-123', merchantId, source: 'source-1', status: 'pending', createdAt: new Date().toISOString() });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Government Database')).toBeInTheDocument();
      });

      const enrichButton = screen.getByRole('button', { name: /enrich/i });
      await user.click(enrichButton);

      await waitFor(() => {
        // Should show processing state
        expect(screen.getByText(/processing/i)).toBeInTheDocument();
      });
    });

    it('should show completed state after enrichment finishes', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
          return HttpResponse.json({ jobId: 'job-123', merchantId, source: 'source-1', status: 'pending', createdAt: new Date().toISOString() });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Government Database')).toBeInTheDocument();
      });

      const enrichButton = screen.getByRole('button', { name: /enrich/i });
      await user.click(enrichButton);

      // Wait for completion (simulated after 2 seconds in component)
      await waitFor(() => {
        expect(screen.getByText(/done|completed/i)).toBeInTheDocument();
      }, { timeout: 3000 });
    });

    it('should handle enrichment failure', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
          return HttpResponse.json({ error: 'Failed to trigger enrichment' }, { status: 500 });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Government Database')).toBeInTheDocument();
      });

      const enrichButton = screen.getByRole('button', { name: /enrich/i });
      await user.click(enrichButton);

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith(
          'Enrichment failed',
          expect.objectContaining({
            description: expect.any(String),
          })
        );
      });
    });
  });

  describe('Error Handling', () => {
    it('should display error message when sources fail to load', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText(/failed|error/i)).toBeInTheDocument();
      });
    });

    it('should show retry button on error', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);

      await waitFor(() => {
        const retryButton = screen.getByRole('button', { name: /retry/i });
        expect(retryButton).toBeInTheDocument();
      });
    });
  });

  describe('Button Variants', () => {
    it('should support different button variants', () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: [] });
        })
      );

      const { rerender } = render(<EnrichmentButton merchantId={merchantId} variant="outline" />);
      expect(screen.getByRole('button', { name: /enrich data/i })).toBeInTheDocument();

      rerender(<EnrichmentButton merchantId={merchantId} variant="default" />);
      expect(screen.getByRole('button', { name: /enrich data/i })).toBeInTheDocument();
    });
  });
});

