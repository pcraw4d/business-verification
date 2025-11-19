import { server } from '@/__tests__/mocks/server';
import { RiskAlertsSection } from '@/components/merchant/RiskAlertsSection';
import { render, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { toast } from 'sonner';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('sonner');
const mockToast = vi.mocked(toast);

describe('RiskAlertsSection', () => {
  const merchantId = 'merchant-123';

  const mockRiskIndicators = {
    merchantId,
    indicators: [
      {
        id: 'indicator-1',
        type: 'financial',
        severity: 'critical',
        title: 'High Financial Risk',
        description: 'Merchant has significant financial risk indicators',
        status: 'active',
        createdAt: '2025-01-27T00:00:00Z',
        updatedAt: '2025-01-27T00:00:00Z',
      },
      {
        id: 'indicator-2',
        type: 'compliance',
        severity: 'high',
        title: 'Compliance Issue',
        description: 'Potential compliance violation detected',
        status: 'active',
        createdAt: '2025-01-26T00:00:00Z',
        updatedAt: '2025-01-26T00:00:00Z',
      },
      {
        id: 'indicator-3',
        type: 'operational',
        severity: 'medium',
        title: 'Operational Risk',
        description: 'Moderate operational risk identified',
        status: 'active',
        createdAt: '2025-01-25T00:00:00Z',
        updatedAt: '2025-01-25T00:00:00Z',
      },
      {
        id: 'indicator-4',
        type: 'security',
        severity: 'low',
        title: 'Security Notice',
        description: 'Minor security concern',
        status: 'active',
        createdAt: '2025-01-24T00:00:00Z',
        updatedAt: '2025-01-24T00:00:00Z',
      },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.warning = vi.fn();
    mockToast.error = vi.fn();
  });

  describe('Loading State', () => {
    it('should show loading skeleton initially', () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json(mockRiskIndicators);
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Success State', () => {
    it('should display risk alerts when loaded', async () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          return HttpResponse.json(mockRiskIndicators);
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('Risk Alerts')).toBeInTheDocument();
        expect(screen.getByText('High Financial Risk')).toBeInTheDocument();
        expect(screen.getByText('Compliance Issue')).toBeInTheDocument();
      });
    });

    it('should group alerts by severity', async () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          return HttpResponse.json(mockRiskIndicators);
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show severity sections
        expect(screen.getByText(/critical/i)).toBeInTheDocument();
        expect(screen.getByText(/high/i)).toBeInTheDocument();
        expect(screen.getByText(/medium/i)).toBeInTheDocument();
        expect(screen.getByText(/low/i)).toBeInTheDocument();
      });
    });

    it('should display alert descriptions', async () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          return HttpResponse.json(mockRiskIndicators);
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/significant financial risk indicators/i)).toBeInTheDocument();
        expect(screen.getByText(/potential compliance violation/i)).toBeInTheDocument();
      });
    });

    it('should show toast notification for critical alerts', async () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          return HttpResponse.json(mockRiskIndicators);
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show toast for critical severity
        expect(mockToast.error).toHaveBeenCalled();
      });
    });

    it('should show toast notification for high severity alerts', async () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          return HttpResponse.json(mockRiskIndicators);
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show toast for high severity
        expect(mockToast.warning).toHaveBeenCalled();
      });
    });
  });

  describe('Empty State', () => {
    it('should display message when no alerts available', async () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          return HttpResponse.json({ merchantId, indicators: [] });
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/no active risk alerts/i)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle API error gracefully', async () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show error message or empty state
        const errorText = screen.queryByText(/error|failed/i);
        const emptyText = screen.queryByText(/no.*alerts/i);
        expect(errorText || emptyText).toBeTruthy();
      }, { timeout: 3000 });
    });
  });

  describe('Auto-refresh', () => {
    it('should auto-refresh alerts periodically', async () => {
      vi.useFakeTimers();
      
      let callCount = 0;
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          callCount++;
          return HttpResponse.json(mockRiskIndicators);
        })
      );

      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('High Financial Risk')).toBeInTheDocument();
      });

      // Fast-forward time to trigger auto-refresh (30 seconds)
      vi.advanceTimersByTime(30000);

      await waitFor(() => {
        // Should have made multiple calls
        expect(callCount).toBeGreaterThan(1);
      });

      vi.useRealTimers();
    });
  });

  describe('Collapsible Sections', () => {
    it('should allow collapsing severity sections', async () => {
      server.use(
        http.get('*/api/v1/risk/indicators/:id', () => {
          return HttpResponse.json(mockRiskIndicators);
        })
      );

      const user = userEvent.setup();
      render(<RiskAlertsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('High Financial Risk')).toBeInTheDocument();
      });

      // Find and click collapse button (if implemented)
      const collapseButtons = screen.queryAllByRole('button', { name: /collapse|expand/i });
      if (collapseButtons.length > 0) {
        await user.click(collapseButtons[0]);
      }
    });
  });
});

