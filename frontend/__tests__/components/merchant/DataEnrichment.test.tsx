import { server } from '@/__tests__/mocks/server';
import { DataEnrichment } from '@/components/merchant/DataEnrichment';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { toast } from 'sonner';

vi.mock('sonner');

const mockToast = vi.mocked(toast);

describe('DataEnrichment', () => {
  const merchantId = 'merchant-123';

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.success = vi.fn();
    mockToast.error = vi.fn();
    
    // Mock sessionStorage
    Object.defineProperty(window, 'sessionStorage', {
      value: {
        getItem: vi.fn(() => null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      writable: true,
    });
  });

  describe('Component Rendering', () => {
    it('should render enrichment button', () => {
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      expect(button).toBeInTheDocument();
    });

    it('should open dialog when button is clicked', async () => {
      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('Data Enrichment')).toBeInTheDocument();
      });
    });
  });

  describe('Loading Sources', () => {
    it('should load enrichment sources on mount', async () => {
      const mockSources = [
        { id: 'source-1', name: 'Source 1', description: 'Description 1' },
        { id: 'source-2', name: 'Source 2', description: 'Description 2' },
      ];

      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
          return HttpResponse.json({ sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('Source 1')).toBeInTheDocument();
        expect(screen.getByText('Source 2')).toBeInTheDocument();
      });
    });

    it('should show loading skeleton while loading sources', async () => {
      let resolveSources: (value: any) => void;
      const sourcesPromise = new Promise((resolve) => {
        resolveSources = resolve;
      });

      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', async () => {
          await sourcesPromise;
          return HttpResponse.json({ sources: [] });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      // Should show skeleton while loading
      await waitFor(() => {
        const skeletons = document.querySelectorAll('[class*="skeleton"]');
        expect(skeletons.length).toBeGreaterThan(0);
      });
      
      // Resolve the promise
      resolveSources!({ sources: [] });
    });

    it('should show empty state when no sources available', async () => {
      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
          return HttpResponse.json({ sources: [] });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText(/no enrichment sources available/i)).toBeInTheDocument();
      });
    });

    it('should handle error when loading sources fails', async () => {
      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
          return HttpResponse.json({ error: 'Failed to load' }, { status: 500 });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      // Should show empty state or handle error gracefully
      await waitFor(() => {
        const emptyState = screen.queryByText(/no enrichment sources available/i);
        expect(emptyState).toBeInTheDocument();
      });
    });
  });

  describe('Triggering Enrichment', () => {
    it('should trigger enrichment when source is clicked', async () => {
      const mockSources = [
        { id: 'source-1', name: 'Source 1', description: 'Description 1' },
      ];

      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
          return HttpResponse.json({ sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:merchantId/enrichment/trigger', () => {
          return HttpResponse.json({ jobId: 'job-123', status: 'pending' });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('Source 1')).toBeInTheDocument();
      });
      
      const sourceButton = screen.getByRole('button', { name: /enrich data from source 1/i });
      await user.click(sourceButton);
      
      await waitFor(() => {
        expect(mockToast.success).toHaveBeenCalledWith('Enrichment job started successfully');
      });
    });

    it('should close dialog after successful enrichment trigger', async () => {
      const mockSources = [
        { id: 'source-1', name: 'Source 1', description: 'Description 1' },
      ];

      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
          return HttpResponse.json({ sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:merchantId/enrichment/trigger', () => {
          return HttpResponse.json({ jobId: 'job-123', status: 'pending' });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('Source 1')).toBeInTheDocument();
      });
      
      const sourceButton = screen.getByRole('button', { name: /enrich data from source 1/i });
      await user.click(sourceButton);
      
      await waitFor(() => {
        expect(screen.queryByText('Data Enrichment')).not.toBeInTheDocument();
      });
    });

    it('should disable source buttons while enriching', async () => {
      let resolveTrigger: (value: any) => void;
      const triggerPromise = new Promise((resolve) => {
        resolveTrigger = resolve;
      });

      const mockSources = [
        { id: 'source-1', name: 'Source 1', description: 'Description 1' },
      ];

      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
          return HttpResponse.json({ sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:merchantId/enrichment/trigger', async () => {
          await triggerPromise;
          return HttpResponse.json({ jobId: 'job-123', status: 'pending' });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('Source 1')).toBeInTheDocument();
      });
      
      const sourceButton = screen.getByRole('button', { name: /enrich data from source 1/i });
      await user.click(sourceButton);
      
      // Button should be disabled while enriching
      await waitFor(() => {
        expect(sourceButton).toBeDisabled();
      });
      
      // Resolve the promise
      resolveTrigger!({ jobId: 'job-123', status: 'pending' });
    });

    it('should show error toast when enrichment trigger fails', async () => {
      const mockSources = [
        { id: 'source-1', name: 'Source 1', description: 'Description 1' },
      ];

      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
          return HttpResponse.json({ sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:merchantId/enrichment/trigger', () => {
          return HttpResponse.json({ error: 'Failed' }, { status: 500 });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('Source 1')).toBeInTheDocument();
      });
      
      const sourceButton = screen.getByRole('button', { name: /enrich data from source 1/i });
      await user.click(sourceButton);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to trigger enrichment');
      });
    });

    it('should handle network errors when triggering enrichment', async () => {
      const mockSources = [
        { id: 'source-1', name: 'Source 1', description: 'Description 1' },
      ];

      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
          return HttpResponse.json({ sources: mockSources });
        }),
        http.post('*/api/v1/merchants/:merchantId/enrichment/trigger', () => {
          return HttpResponse.error();
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('Source 1')).toBeInTheDocument();
      });
      
      const sourceButton = screen.getByRole('button', { name: /enrich data from source 1/i });
      await user.click(sourceButton);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to trigger enrichment');
      });
    });
  });

  describe('Authentication', () => {
    it('should include auth token in headers when available', async () => {
      const getItemSpy = vi.fn(() => 'test-token');
      Object.defineProperty(window, 'sessionStorage', {
        value: {
          getItem: getItemSpy,
          setItem: vi.fn(),
          removeItem: vi.fn(),
          clear: vi.fn(),
        },
        writable: true,
      });

      const mockSources = [
        { id: 'source-1', name: 'Source 1', description: 'Description 1' },
      ];

      server.use(
        http.get('*/api/v1/merchants/:merchantId/enrichment/sources', ({ request }) => {
          const authHeader = request.headers.get('Authorization');
          expect(authHeader).toBe('Bearer test-token');
          return HttpResponse.json({ sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(getItemSpy).toHaveBeenCalledWith('authToken');
      });
    });
  });

  describe('Dialog Management', () => {
    it('should close dialog when clicking outside', async () => {
      const user = userEvent.setup();
      render(<DataEnrichment merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /enrich data/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('Data Enrichment')).toBeInTheDocument();
      });
      
      // Press Escape to close dialog
      await user.keyboard('{Escape}');
      
      await waitFor(() => {
        expect(screen.queryByText('Data Enrichment')).not.toBeInTheDocument();
      });
    });
  });
});

