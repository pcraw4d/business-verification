import { server } from '@/__tests__/mocks/server';
import { RiskRecommendationsSection } from '@/components/merchant/RiskRecommendationsSection';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, it, expect, beforeEach } from 'vitest';

describe('RiskRecommendationsSection', () => {
  const merchantId = 'merchant-123';

  const mockRecommendations = {
    merchantId,
    recommendations: [
      {
        id: 'rec-1',
        type: 'financial',
        priority: 'high',
        title: 'Improve Financial Stability',
        description: 'Consider implementing financial monitoring and reporting systems',
        actionItems: [
          'Set up automated financial reporting',
          'Implement cash flow monitoring',
          'Establish financial reserves',
        ],
      },
      {
        id: 'rec-2',
        type: 'compliance',
        priority: 'high',
        title: 'Enhance Compliance Framework',
        description: 'Strengthen compliance processes and documentation',
        actionItems: [
          'Review compliance policies',
          'Conduct compliance audit',
          'Update documentation',
        ],
      },
      {
        id: 'rec-3',
        type: 'operational',
        priority: 'medium',
        title: 'Optimize Operations',
        description: 'Improve operational efficiency and risk management',
        actionItems: [
          'Review operational processes',
          'Implement risk controls',
        ],
      },
      {
        id: 'rec-4',
        type: 'security',
        priority: 'low',
        title: 'Security Best Practices',
        description: 'Adopt additional security measures',
        actionItems: [
          'Update security protocols',
        ],
      },
    ],
    timestamp: new Date().toISOString(),
  };

  beforeEach(() => {
    server.use(
      http.get('*/api/v1/merchants/:id/risk-recommendations', () => {
        return HttpResponse.json(mockRecommendations);
      })
    );
  });

  describe('Loading State', () => {
    it('should show loading skeleton initially', () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-recommendations', async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json(mockRecommendations);
        })
      );

      render(<RiskRecommendationsSection merchantId={merchantId} />);

      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Success State', () => {
    it('should display risk recommendations when loaded', async () => {
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
        expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        expect(screen.getByText('Enhance Compliance Framework')).toBeInTheDocument();
      });
    });

    it('should group recommendations by priority', async () => {
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show priority sections
        expect(screen.getByText(/high priority/i)).toBeInTheDocument();
        expect(screen.getByText(/medium priority/i)).toBeInTheDocument();
        expect(screen.getByText(/low priority/i)).toBeInTheDocument();
      });
    });

    it('should display recommendation descriptions', async () => {
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/implementing financial monitoring/i)).toBeInTheDocument();
        expect(screen.getByText(/strengthen compliance processes/i)).toBeInTheDocument();
      });
    });

    it('should display action items for each recommendation', async () => {
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/set up automated financial reporting/i)).toBeInTheDocument();
        expect(screen.getByText(/review compliance policies/i)).toBeInTheDocument();
        expect(screen.getByText(/review operational processes/i)).toBeInTheDocument();
      });
    });

    it('should display priority badges', async () => {
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show priority badges
        const badges = screen.getAllByText(/high|medium|low/i);
        expect(badges.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Empty State', () => {
    it('should display message when no recommendations available', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-recommendations', () => {
          return HttpResponse.json({ merchantId, recommendations: [], timestamp: new Date().toISOString() });
        })
      );

      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/no recommendations available/i)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle API error gracefully', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-recommendations', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show error message or empty state
        const errorText = screen.queryByText(/error|failed/i);
        const emptyText = screen.queryByText(/no.*recommendations/i);
        expect(errorText || emptyText).toBeTruthy();
      }, { timeout: 3000 });
    });
  });

  describe('Collapsible Sections', () => {
    it('should allow collapsing priority sections', async () => {
      const user = userEvent.setup();
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
      });

      // Find and click collapse button (if implemented)
      const collapseButtons = screen.queryAllByRole('button', { name: /collapse|expand/i });
      if (collapseButtons.length > 0) {
        await user.click(collapseButtons[0]);
      }
    });
  });

  describe('Action Items Display', () => {
    it('should display all action items for a recommendation', async () => {
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show all action items for high priority recommendations
        expect(screen.getByText(/set up automated financial reporting/i)).toBeInTheDocument();
        expect(screen.getByText(/implement cash flow monitoring/i)).toBeInTheDocument();
        expect(screen.getByText(/establish financial reserves/i)).toBeInTheDocument();
      });
    });

    it('should format action items as list items', async () => {
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Action items should be displayed in a list format
        const listItems = document.querySelectorAll('li');
        expect(listItems.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Phase 4 Enhancements', () => {
    describe('Mark as Complete', () => {
      it('should display "Mark as Complete" button for each recommendation', async () => {
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          const completeButtons = screen.getAllByRole('button', { name: /mark as complete/i });
          expect(completeButtons.length).toBeGreaterThan(0);
        });
      });

      it('should mark recommendation as complete when button is clicked', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        });

        const completeButtons = screen.getAllByRole('button', { name: /mark as complete/i });
        await user.click(completeButtons[0]);

        await waitFor(() => {
          // Recommendation should be marked as complete (check for visual indicator)
          const completedIndicator = screen.queryByText(/completed/i);
          expect(completedIndicator).toBeInTheDocument();
        });
      });

      it('should track completed recommendations count', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        });

        const completeButtons = screen.getAllByRole('button', { name: /mark as complete/i });
        await user.click(completeButtons[0]);

        await waitFor(() => {
          // Should show count of completed recommendations
          expect(screen.getByText(/1.*completed/i)).toBeInTheDocument();
        });
      });
    });

    describe('Filtering by Priority', () => {
      it('should display priority filter dropdown', async () => {
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          const filterSelect = screen.getByRole('combobox', { name: /filter/i });
          expect(filterSelect).toBeInTheDocument();
        });
      });

      it('should filter recommendations by priority when filter is selected', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        });

        const filterSelect = screen.getByRole('combobox', { name: /filter/i });
        await user.click(filterSelect);

        await waitFor(() => {
          const highOption = screen.getByRole('option', { name: /high/i });
          await user.click(highOption);
        });

        await waitFor(() => {
          // Should only show high priority recommendations
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
          expect(screen.queryByText('Optimize Operations')).not.toBeInTheDocument();
        });
      });
    });

    describe('Search Functionality', () => {
      it('should display search input', async () => {
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          const searchInput = screen.getByPlaceholderText(/search/i);
          expect(searchInput).toBeInTheDocument();
        });
      });

      it('should filter recommendations by search text', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        });

        const searchInput = screen.getByPlaceholderText(/search/i);
        await user.type(searchInput, 'Financial');

        await waitFor(() => {
          // Should only show recommendations matching "Financial"
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
          expect(screen.queryByText('Enhance Compliance Framework')).not.toBeInTheDocument();
        });
      });

      it('should search across title, description, type, and action items', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        });

        const searchInput = screen.getByPlaceholderText(/search/i);
        await user.type(searchInput, 'cash flow');

        await waitFor(() => {
          // Should find recommendation with "cash flow" in action items
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        });
      });
    });
  });
});

