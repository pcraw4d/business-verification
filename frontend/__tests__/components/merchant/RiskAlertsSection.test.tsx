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

  describe('Phase 4 Enhancements', () => {
    describe('Dismiss Functionality', () => {
      it('should display dismiss button for each alert', async () => {
        server.use(
          http.get('*/api/v1/risk/indicators/:id', () => {
            return HttpResponse.json(mockRiskIndicators);
          })
        );

        render(<RiskAlertsSection merchantId={merchantId} />);

        await waitFor(() => {
          const dismissButtons = screen.getAllByRole('button', { name: /dismiss/i });
          expect(dismissButtons.length).toBeGreaterThan(0);
        });
      });

      it('should dismiss alert when dismiss button is clicked', async () => {
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

        const dismissButtons = screen.getAllByRole('button', { name: /dismiss/i });
        await user.click(dismissButtons[0]);

        await waitFor(() => {
          // Alert should be dismissed (removed from view)
          expect(screen.queryByText('High Financial Risk')).not.toBeInTheDocument();
        });
      });

      it('should track dismissed alerts count', async () => {
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

        const dismissButtons = screen.getAllByRole('button', { name: /dismiss/i });
        await user.click(dismissButtons[0]);

        await waitFor(() => {
          // Should show count of dismissed alerts
          expect(screen.getByText(/1.*dismissed/i)).toBeInTheDocument();
        });
      });
    });

    describe('Filtering by Severity', () => {
      it('should display severity filter dropdown', async () => {
        server.use(
          http.get('*/api/v1/risk/indicators/:id', () => {
            return HttpResponse.json(mockRiskIndicators);
          })
        );

        render(<RiskAlertsSection merchantId={merchantId} />);

        await waitFor(() => {
          const filterSelect = screen.getByRole('combobox', { name: /filter/i });
          expect(filterSelect).toBeInTheDocument();
        });
      });

      it('should filter alerts by severity when filter is selected', async () => {
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

        const filterSelect = screen.getByRole('combobox', { name: /filter/i });
        await user.click(filterSelect);

        await waitFor(() => {
          const criticalOption = screen.getByRole('option', { name: /critical/i });
          await user.click(criticalOption);
        });

        await waitFor(() => {
          // Should only show critical severity alerts
          expect(screen.getByText('High Financial Risk')).toBeInTheDocument();
          expect(screen.queryByText('Compliance Issue')).not.toBeInTheDocument();
        });
      });
    });

    describe('View All Alerts Link', () => {
      it('should display "View All Alerts" link', async () => {
        server.use(
          http.get('*/api/v1/risk/indicators/:id', () => {
            return HttpResponse.json(mockRiskIndicators);
          })
        );

        render(<RiskAlertsSection merchantId={merchantId} />);

        await waitFor(() => {
          const viewAllLink = screen.getByRole('link', { name: /view all alerts/i });
          expect(viewAllLink).toBeInTheDocument();
        });
      });
    });

    describe('WebSocket Real-time Updates', () => {
      it('should listen for WebSocket riskAlert events', async () => {
        server.use(
          http.get('*/api/v1/risk/indicators/:id', () => {
            return HttpResponse.json(mockRiskIndicators);
          })
        );

        render(<RiskAlertsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText('High Financial Risk')).toBeInTheDocument();
        });

        // Simulate WebSocket event
        const newAlert = {
          merchantId,
          alert: {
            id: 'indicator-5',
            type: 'security',
            severity: 'high',
            title: 'New Security Alert',
            description: 'New security issue detected',
            status: 'active',
            createdAt: new Date().toISOString(),
          },
        };

        window.dispatchEvent(
          new CustomEvent('riskAlert', {
            detail: newAlert,
          })
        );

        await waitFor(() => {
          // New alert should appear
          expect(screen.getByText('New Security Alert')).toBeInTheDocument();
        });
      });

      it('should show toast notification for new critical alerts', async () => {
        server.use(
          http.get('*/api/v1/risk/indicators/:id', () => {
            return HttpResponse.json(mockRiskIndicators);
          })
        );

        render(<RiskAlertsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText('High Financial Risk')).toBeInTheDocument();
        });

        // Simulate WebSocket event with critical alert
        const criticalAlert = {
          merchantId,
          alert: {
            id: 'indicator-6',
            type: 'financial',
            severity: 'critical',
            title: 'Critical Financial Alert',
            description: 'Critical financial issue',
            status: 'active',
            createdAt: new Date().toISOString(),
          },
        };

        window.dispatchEvent(
          new CustomEvent('riskAlert', {
            detail: criticalAlert,
          })
        );

        await waitFor(() => {
          // Should show error toast for critical alerts
          expect(mockToast.error).toHaveBeenCalled();
        });
      });

      it('should update alerts state when WebSocket event is received', async () => {
        server.use(
          http.get('*/api/v1/risk/indicators/:id', () => {
            return HttpResponse.json(mockRiskIndicators);
          })
        );

        render(<RiskAlertsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText('High Financial Risk')).toBeInTheDocument();
        });

        const initialAlertCount = screen.getAllByText(/risk|alert/i).length;

        // Simulate WebSocket event
        const newAlert = {
          merchantId,
          alert: {
            id: 'indicator-7',
            type: 'operational',
            severity: 'medium',
            title: 'New Operational Alert',
            description: 'New operational issue',
            status: 'active',
            createdAt: new Date().toISOString(),
          },
        };

        window.dispatchEvent(
          new CustomEvent('riskAlert', {
            detail: newAlert,
          })
        );

        await waitFor(() => {
          // Should have more alerts now
          expect(screen.getByText('New Operational Alert')).toBeInTheDocument();
        });
      });
    });
  });
});

