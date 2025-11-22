import { server } from '@/__tests__/mocks/server';
import { EnrichmentButton } from '@/components/merchant/EnrichmentButton';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
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

  // Helper function to select a source by clicking its Card element
  const selectSource = async (user: ReturnType<typeof userEvent.setup>, sourceName: string) => {
    await waitFor(() => {
      expect(screen.getByText(sourceName)).toBeInTheDocument();
    }, { timeout: 5000 });
    
    // Click the Card element - it has onClick handler that calls toggleSourceSelection
    const sourceText = screen.getByText(sourceName);
    const cardElement = sourceText.closest('div[class*="card"]') || 
                       sourceText.parentElement?.parentElement;
    expect(cardElement).toBeTruthy();
    await user.click(cardElement as HTMLElement);

    // Wait for selection to update
    await waitFor(() => {
      const checkbox = document.querySelector('input[type="checkbox"]') as HTMLInputElement;
      expect(checkbox?.checked).toBe(true);
    }, { timeout: 3000 });
  };

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

      expect(screen.getByRole('button', { name: /enrich merchant data/i })).toBeInTheDocument();
    });

    it('should display number of enabled sources in badge', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
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

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Data Enrichment')).toBeInTheDocument();
        expect(screen.getByText(/Select data sources to enrich merchant information/i)).toBeInTheDocument();
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

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
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
      // Use a promise that never resolves to keep component in loading state
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', async () => {
          return new Promise(() => {}); // Never resolve to keep in loading state
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
      await user.click(button);

      // Should show skeleton while loading
      await waitFor(() => {
        const skeletons = document.querySelectorAll('[class*="skeleton"], [data-slot="skeleton"]');
        expect(skeletons.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
    });

    it('should display enrichment sources when loaded', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ merchantId, sources: mockSources });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
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

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
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

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
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

      const user = userEvent.setup({ delay: null });
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
      await user.click(button);

      await waitFor(() => {
        expect(screen.getByText('Government Database')).toBeInTheDocument();
      }, { timeout: 5000 });

      // Select a source - the Card element has onClick that toggles selection
      // Find the Card containing "Government Database" and click it
      await waitFor(() => {
        const governmentCard = screen.getByText('Government Database').closest('[data-slot="card"]') ||
                              screen.getByText('Government Database').closest('div[class*="card"]') ||
                              screen.getByText('Government Database').parentElement?.parentElement;
        expect(governmentCard).toBeTruthy();
      }, { timeout: 5000 });
      
      // Click the Card element - it has onClick handler that calls toggleSourceSelection
      const governmentText = screen.getByText('Government Database');
      const cardElement = governmentText.closest('div[class*="card"]') || 
                         governmentText.parentElement?.parentElement;
      expect(cardElement).toBeTruthy();
      await user.click(cardElement as HTMLElement);

      // Wait for selection to update - component updates selectedSources state
      await waitFor(() => {
        const checkbox = document.querySelector('input[type="checkbox"]') as HTMLInputElement;
        expect(checkbox?.checked).toBe(true);
      }, { timeout: 3000 });
      
      // Verify the "Enrich Selected" button is enabled
      await waitFor(() => {
        const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
        expect(enrichButton).not.toBeDisabled();
      }, { timeout: 3000 });

      // Click the "Enrich Selected" button
      const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
      await user.click(enrichButton);

      // The component calls toast.success immediately after API response
      // No need to wait for setTimeout delays - toast is called right after triggerEnrichment resolves
      await waitFor(() => {
        expect(mockToast.success).toHaveBeenCalledWith(
          'Enrichment job started',
          expect.objectContaining({
            description: expect.stringContaining('job'),
          })
        );
      }, { timeout: 5000 });
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

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
      await user.click(button);

      // Select a source using the helper function
      await selectSource(user, 'Government Database');

      // Click the "Enrich Selected" button
      const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
      await user.click(enrichButton);

      await waitFor(() => {
        // Should show processing state - check for progress bar
        // Component shows Progress bar when status is 'processing'
        const progressBar = document.querySelector('[role="progressbar"]');
        expect(progressBar).toBeTruthy();
      }, { timeout: 5000 });
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

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
      await user.click(button);

      // Select a source using the helper function
      await selectSource(user, 'Government Database');

      // Click the "Enrich Selected" button
      const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
      await user.click(enrichButton);

      // Wait for completion (component simulates progress steps with 500ms delays, then completes)
      // Total time: 4 steps * 500ms = 2000ms, plus API call time
      await waitFor(() => {
        // Check for completion toast - component calls this after all progress steps
        expect(mockToast.success).toHaveBeenCalledWith(
          'Enrichment completed',
          expect.any(Object)
        );
      }, { timeout: 10000 });
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

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
      await user.click(button);

      // Select a source using the helper function
      await selectSource(user, 'Government Database');

      // Click the "Enrich Selected" button
      const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
      await user.click(enrichButton);

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith(
          'Enrichment failed',
          expect.objectContaining({
            description: expect.any(String),
          })
        );
      }, { timeout: 5000 });
    });
  });

  describe('Error Handling', () => {
    it('should display error message when sources fail to load', async () => {
      // Use a non-404 error to trigger error display (404s are handled silently)
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ error: 'Internal Server Error' }, { status: 500 });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
      await user.click(button);

      await waitFor(() => {
        // Component shows error in Alert when error && !sources.length
        expect(screen.getByText(/failed|error|internal server error/i)).toBeInTheDocument();
      }, { timeout: 5000 });
    });

    it('should show retry button on error', async () => {
      // Use a non-404 error to trigger error display (404s are handled silently)
      server.use(
        http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
          return HttpResponse.json({ error: 'Internal Server Error' }, { status: 500 });
        })
      );

      const user = userEvent.setup();
      render(<EnrichmentButton merchantId={merchantId} />);

      const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
      await user.click(button);

      await waitFor(() => {
        // Component shows retry button in Alert when error && !sources.length
        const retryButton = screen.getByRole('button', { name: /retry/i });
        expect(retryButton).toBeInTheDocument();
      }, { timeout: 5000 });
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
      expect(screen.getByRole('button', { name: /enrich merchant data/i })).toBeInTheDocument();

      rerender(<EnrichmentButton merchantId={merchantId} variant="default" />);
      expect(screen.getByRole('button', { name: /enrich merchant data/i })).toBeInTheDocument();
    });
  });

  describe('Phase 4 Enhancements', () => {
    describe('Multiple Vendor Selection', () => {
      it('should allow selecting multiple vendors with checkboxes', async () => {
        server.use(
          http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
            return HttpResponse.json({ merchantId, sources: mockSources });
          })
        );

        const user = userEvent.setup();
        render(<EnrichmentButton merchantId={merchantId} />);

        const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
        await user.click(button);

        await waitFor(() => {
          expect(screen.getByText('Government Database')).toBeInTheDocument();
        });

        // Should have checkboxes for vendor selection
        const checkboxes = document.querySelectorAll('input[type="checkbox"]');
        expect(checkboxes.length).toBeGreaterThan(0);

        // Select multiple vendors by clicking their Card elements
        const govText = screen.getByText('Government Database');
        const govCard = govText.closest('div[class*="card"]') || govText.parentElement?.parentElement;
        await user.click(govCard as HTMLElement);
        
        const creditText = screen.getByText('Credit Bureau');
        const creditCard = creditText.closest('div[class*="card"]') || creditText.parentElement?.parentElement;
        await user.click(creditCard as HTMLElement);

        await waitFor(() => {
          // Both should be selected
          const updatedCheckboxes = document.querySelectorAll('input[type="checkbox"]');
          expect((updatedCheckboxes[0] as HTMLInputElement).checked).toBe(true);
          expect((updatedCheckboxes[1] as HTMLInputElement).checked).toBe(true);
        }, { timeout: 3000 });
      });

      it('should trigger enrichment for all selected vendors', async () => {
        let enrichmentCalls = 0;
        server.use(
          http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
            return HttpResponse.json({ merchantId, sources: mockSources });
          }),
          http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
            enrichmentCalls++;
            return HttpResponse.json({ 
              jobId: `job-${enrichmentCalls}`, 
              merchantId, 
              source: `source-${enrichmentCalls}`, 
              status: 'pending', 
              createdAt: new Date().toISOString() 
            });
          })
        );

        const user = userEvent.setup();
        render(<EnrichmentButton merchantId={merchantId} />);

        const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
        await user.click(button);

        await waitFor(() => {
          expect(screen.getByText('Government Database')).toBeInTheDocument();
        });

        // Select multiple vendors by clicking their Card elements
        const govText = screen.getByText('Government Database');
        const govCard = govText.closest('div[class*="card"]') || govText.parentElement?.parentElement;
        await user.click(govCard as HTMLElement);
        
        const creditText = screen.getByText('Credit Bureau');
        const creditCard = creditText.closest('div[class*="card"]') || creditText.parentElement?.parentElement;
        await user.click(creditCard as HTMLElement);

        // Wait for selection to update
        await waitFor(() => {
          const checkboxes = document.querySelectorAll('input[type="checkbox"]');
          expect((checkboxes[0] as HTMLInputElement).checked).toBe(true);
          expect((checkboxes[1] as HTMLInputElement).checked).toBe(true);
        }, { timeout: 3000 });

        // Trigger enrichment using "Enrich Selected" button
        const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
        await user.click(enrichButton);

        await waitFor(() => {
          // Should have triggered enrichment for both selected vendors
          expect(enrichmentCalls).toBeGreaterThanOrEqual(2);
        }, { timeout: 5000 });
      });
    });

    describe('Job Tracking', () => {
      it('should display job status and progress', async () => {
        server.use(
          http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
            return HttpResponse.json({ merchantId, sources: mockSources });
          }),
          http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
            return HttpResponse.json({ 
              jobId: 'job-123', 
              merchantId, 
              source: 'source-1', 
              status: 'pending', 
              createdAt: new Date().toISOString() 
            });
          })
        );

        const user = userEvent.setup();
        render(<EnrichmentButton merchantId={merchantId} />);

        const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
        await user.click(button);

        await waitFor(() => {
          expect(screen.getByText('Government Database')).toBeInTheDocument();
        });

        // Select a source using the helper function
        await selectSource(user, 'Government Database');

        // Click the "Enrich Selected" button
        const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
        await user.click(enrichButton);

        await waitFor(() => {
          // Should show job status - check for progress bar
          const statusIndicator = document.querySelector('[role="progressbar"]');
          expect(statusIndicator).toBeTruthy();
        }, { timeout: 5000 });
      });

      it('should show progress indicator during enrichment', async () => {
        server.use(
          http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
            return HttpResponse.json({ merchantId, sources: mockSources });
          }),
          http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
            return HttpResponse.json({ 
              jobId: 'job-123', 
              merchantId, 
              source: 'source-1', 
              status: 'processing', 
              createdAt: new Date().toISOString() 
            });
          })
        );

        const user = userEvent.setup();
        render(<EnrichmentButton merchantId={merchantId} />);

        const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
        await user.click(button);

        await waitFor(() => {
          expect(screen.getByText('Government Database')).toBeInTheDocument();
        });

        // Select a source using the helper function
        await selectSource(user, 'Government Database');

        // Click the "Enrich Selected" button
        const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
        await user.click(enrichButton);

        await waitFor(() => {
          // Should show progress bar - component shows Progress when processing
          const progress = document.querySelector('[role="progressbar"]');
          expect(progress).toBeInTheDocument();
        }, { timeout: 5000 });
      });
    });

    describe('Enrichment History', () => {
      it('should display enrichment history tab', async () => {
        server.use(
          http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
            return HttpResponse.json({ merchantId, sources: mockSources });
          })
        );

        const user = userEvent.setup();
        render(<EnrichmentButton merchantId={merchantId} />);

        const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
        await user.click(button);

        await waitFor(() => {
          // Should have History tab
          const historyTab = screen.getByRole('tab', { name: /history/i });
          expect(historyTab).toBeInTheDocument();
        });
      });

      it('should show past enrichment jobs in history', async () => {
        server.use(
          http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
            return HttpResponse.json({ merchantId, sources: mockSources });
          }),
          http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
            return HttpResponse.json({ 
              jobId: 'job-123', 
              merchantId, 
              source: 'source-1', 
              status: 'completed', 
              createdAt: new Date().toISOString(),
              completedAt: new Date().toISOString(),
            });
          })
        );

        const user = userEvent.setup();
        render(<EnrichmentButton merchantId={merchantId} />);

        const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
        await user.click(button);

        await waitFor(() => {
          expect(screen.getByText('Government Database')).toBeInTheDocument();
        });

        // Select a source using the helper function
        await selectSource(user, 'Government Database');

        // Click the "Enrich Selected" button
        const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
        await user.click(enrichButton);

        // Wait for completion (component simulates progress with 4 steps * 500ms = 2000ms, then completes)
        await waitFor(() => {
          // Check for completion toast
          expect(mockToast.success).toHaveBeenCalledWith(
            'Enrichment completed',
            expect.any(Object)
          );
        }, { timeout: 10000 });

        // Open history tab
        const historyTab = screen.getByRole('tab', { name: /history/i });
        await user.click(historyTab);

        await waitFor(() => {
          // Should show completed job in history - check for job details
          const jobInHistory = screen.queryByText(/Government Database/i) ||
                             screen.queryByText(/completed/i) ||
                             screen.queryByText(/job/i);
          expect(jobInHistory).toBeTruthy();
        }, { timeout: 3000 });
      });
    });

    describe('Results Display', () => {
      it('should display enrichment results (added/updated/unchanged fields)', async () => {
        server.use(
          http.get('*/api/v1/merchants/:id/enrichment/sources', () => {
            return HttpResponse.json({ merchantId, sources: mockSources });
          }),
          http.post('*/api/v1/merchants/:id/enrichment/trigger', () => {
            return HttpResponse.json({ 
              jobId: 'job-123', 
              merchantId, 
              source: 'source-1', 
              status: 'completed', 
              createdAt: new Date().toISOString(),
              results: {
                added: ['Founded Date', 'Employee Count'],
                updated: ['Annual Revenue'],
                unchanged: ['Business Name'],
              },
            });
          })
        );

        const user = userEvent.setup();
        render(<EnrichmentButton merchantId={merchantId} />);

        const button = screen.getByRole('button', { name: /enrich merchant data|enrich data/i });
        await user.click(button);

        // Select a source using the helper function
        await selectSource(user, 'Government Database');

        // Click the "Enrich Selected" button
        const enrichButton = screen.getByRole('button', { name: /enrich selected/i });
        await user.click(enrichButton);

        // Wait for completion (component simulates progress, then completes with results)
        await waitFor(() => {
          // Check for completion toast first
          expect(mockToast.success).toHaveBeenCalledWith(
            'Enrichment completed',
            expect.any(Object)
          );
        }, { timeout: 10000 });

        // Open history tab to see results
        const historyTab = screen.getByRole('tab', { name: /history/i });
        await user.click(historyTab);

        await waitFor(() => {
          // Should show results in history - component displays added/updated/unchanged fields
          expect(screen.getByText(/founded date|employee count/i)).toBeInTheDocument();
          expect(screen.getByText(/annual revenue/i)).toBeInTheDocument();
        }, { timeout: 5000 });
      });
    });
  });
});

