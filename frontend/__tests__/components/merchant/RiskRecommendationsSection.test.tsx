import { server } from '@/__tests__/mocks/server';
import { RiskRecommendationsSection } from '@/components/merchant/RiskRecommendationsSection';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { toast } from 'sonner';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import React from 'react';

vi.mock('sonner');
const mockToast = vi.mocked(toast);

// Mock Radix UI Select components to bypass portal rendering issues in JSDOM
vi.mock('@/components/ui/select', () => {
  const MockSelect = ({ children, value, onValueChange }: any) => {
    const [selectedValue, setSelectedValue] = React.useState(value || '');
    const [isOpen, setIsOpen] = React.useState(false);
    
    React.useEffect(() => {
      if (value !== undefined) {
        setSelectedValue(value);
      }
    }, [value]);
    
    const handleChange = (newValue: string) => {
      setSelectedValue(newValue);
      onValueChange?.(newValue);
      setIsOpen(false);
    };
    
    // Extract SelectContent and SelectTrigger from children
    const childrenArray = React.Children.toArray(children);
    let selectTrigger: any = null;
    let selectContent: any = null;
    
    childrenArray.forEach((child: any) => {
      if (child?.type?.displayName === 'SelectTrigger' || 
          child?.type?.name === 'SelectTrigger') {
        selectTrigger = child;
      } else if (child?.type?.displayName === 'SelectContent' || 
                 child?.type?.name === 'SelectContent') {
        selectContent = child;
      }
    });
    
    // Get placeholder from SelectValue (which is a child of SelectTrigger)
    const selectValue = selectTrigger?.props?.children;
    const placeholder = selectValue?.props?.placeholder || 'Select...';
    
    // Extract options from SelectContent
    const options = React.Children.toArray(selectContent?.props?.children || []);
    const selectedOption = options.find(
      (opt: any) => opt?.props?.value === selectedValue
    );
    const displayValue = selectedOption?.props?.children || placeholder;
    
    return (
      <div data-testid="select-wrapper">
        <button
          type="button"
          role="combobox"
          aria-expanded={isOpen}
          onClick={() => setIsOpen(!isOpen)}
          data-value={selectedValue}
        >
          {displayValue}
        </button>
        {isOpen && (
          <div role="listbox" data-testid="select-content">
            {options.map((option: any) => (
              <div
                key={option?.props?.value}
                role="option"
                data-value={option?.props?.value}
                data-testid={`select-item-${option?.props?.value}`}
                onClick={() => handleChange(option?.props?.value)}
              >
                {option?.props?.children}
              </div>
            ))}
          </div>
        )}
      </div>
    );
  };
  
  MockSelect.displayName = 'Select';
  
  const MockSelectTrigger = ({ children, ...props }: any) => {
    return <div {...props}>{children}</div>;
  };
  MockSelectTrigger.displayName = 'SelectTrigger';
  
  const MockSelectContent = ({ children, ...props }: any) => {
    return <div {...props}>{children}</div>;
  };
  MockSelectContent.displayName = 'SelectContent';
  
  const MockSelectItem = ({ children, value, ...props }: any) => {
    return <div {...props} data-value={value}>{children}</div>;
  };
  MockSelectItem.displayName = 'SelectItem';
  
  const MockSelectValue = ({ placeholder, ...props }: any) => {
    return <div {...props}>{placeholder}</div>;
  };
  MockSelectValue.displayName = 'SelectValue';
  
  return {
    Select: MockSelect,
    SelectTrigger: MockSelectTrigger,
    SelectContent: MockSelectContent,
    SelectItem: MockSelectItem,
    SelectValue: MockSelectValue,
  };
});

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
    vi.clearAllMocks();
    mockToast.success = vi.fn();
    mockToast.error = vi.fn();
    server.use(
      // getRiskRecommendations uses ApiEndpoints.merchants.riskRecommendations
      // Format: /api/v1/merchants/:merchantId/risk-recommendations
      http.get('*/api/v1/merchants/:merchantId/risk-recommendations', ({ params }) => {
        return HttpResponse.json(mockRecommendations);
      })
    );
  });

  describe('Loading State', () => {
    it('should show loading skeleton initially', () => {
      server.use(
        http.get('*/api/v1/merchants/:merchantId/risk-recommendations', () => {
          // Never resolve to keep in loading state
          return new Promise(() => {});
        })
      );

      const { container } = render(<RiskRecommendationsSection merchantId={merchantId} />);

      // Check for skeleton - component uses Skeleton component
      const skeleton = container.querySelector('[class*="skeleton"], [data-slot="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Success State', () => {
    it('should display risk recommendations when loaded', async () => {
      const user = userEvent.setup();
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
      }, { timeout: 5000 });

      // Recommendations are in collapsible sections - expand high priority section
      const highButton = screen.getByRole('button', { name: /high/i });
      await user.click(highButton);

      await waitFor(() => {
        expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        expect(screen.getByText('Enhance Compliance Framework')).toBeInTheDocument();
      }, { timeout: 3000 });
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
      const user = userEvent.setup();
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
      }, { timeout: 5000 });

      // Expand high priority section to see descriptions
      const highButton = screen.getByRole('button', { name: /high/i });
      await user.click(highButton);

      await waitFor(() => {
        expect(screen.getByText(/Consider implementing financial monitoring/i)).toBeInTheDocument();
        expect(screen.getByText(/Strengthen compliance processes/i)).toBeInTheDocument();
      }, { timeout: 3000 });
    });

    it('should display action items for each recommendation', async () => {
      const user = userEvent.setup();
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
      }, { timeout: 5000 });

      // Expand high priority section
      const highButton = screen.getByRole('button', { name: /high/i });
      await user.click(highButton);

      await waitFor(() => {
        expect(screen.getByText(/Set up automated financial reporting/i)).toBeInTheDocument();
        expect(screen.getByText(/Review compliance policies/i)).toBeInTheDocument();
      }, { timeout: 3000 });

      // Expand medium priority section
      const mediumButton = screen.getByRole('button', { name: /medium/i });
      await user.click(mediumButton);

      await waitFor(() => {
        expect(screen.getByText(/Review operational processes/i)).toBeInTheDocument();
      }, { timeout: 3000 });
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
        http.get('*/api/v1/merchants/:merchantId/risk-recommendations', () => {
          return HttpResponse.json({ merchantId, recommendations: [], timestamp: new Date().toISOString() });
        })
      );

      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        // Component shows "No Recommendations" when empty
        expect(screen.getByText(/no recommendations/i)).toBeInTheDocument();
      }, { timeout: 5000 });
    });
  });

  describe('Error Handling', () => {
    it('should handle API error gracefully', async () => {
      server.use(
        http.get('*/api/v1/merchants/:merchantId/risk-recommendations', () => {
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
    it('should allow collapsing priority sections', { timeout: 15000 }, async () => {
      const user = userEvent.setup();
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
      }, { timeout: 5000 });

      // Expand high priority section
      const highButton = screen.getByRole('button', { name: /high/i });
      await user.click(highButton);

      await waitFor(() => {
        expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
      }, { timeout: 5000 });

      // Click again to collapse
      await user.click(highButton);

      // Recommendation should no longer be visible (collapsed)
      await waitFor(() => {
        expect(screen.queryByText('Improve Financial Stability')).not.toBeInTheDocument();
      }, { timeout: 5000 });
    });
  });

  describe('Action Items Display', () => {
    it('should display all action items for a recommendation', async () => {
      const user = userEvent.setup();
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
      }, { timeout: 5000 });

      // Expand high priority section to see action items
      const highButton = screen.getByRole('button', { name: /high/i });
      await user.click(highButton);

      await waitFor(() => {
        // Should show all action items for high priority recommendations
        expect(screen.getByText(/Set up automated financial reporting/i)).toBeInTheDocument();
        expect(screen.getByText(/Implement cash flow monitoring/i)).toBeInTheDocument();
        expect(screen.getByText(/Establish financial reserves/i)).toBeInTheDocument();
      }, { timeout: 5000 });
    });

    it('should format action items as list items', async () => {
      const user = userEvent.setup();
      render(<RiskRecommendationsSection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
      }, { timeout: 5000 });

      // Expand high priority section to see action items
      const highButton = screen.getByRole('button', { name: /high/i });
      await user.click(highButton);

      await waitFor(() => {
        // Action items should be displayed in a list format
        const listItems = document.querySelectorAll('li');
        expect(listItems.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
    });
  });

  describe('Phase 4 Enhancements', () => {
    describe('Mark as Complete', () => {
      it('should display "Mark as Complete" button for each recommendation', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
        }, { timeout: 5000 });

        // Expand high priority section to see complete buttons
        const highButton = screen.getByRole('button', { name: /high/i });
        await user.click(highButton);

        await waitFor(() => {
          const completeButtons = screen.getAllByRole('button', { name: /mark.*complete/i });
          expect(completeButtons.length).toBeGreaterThan(0);
        }, { timeout: 5000 });
      });

      it('should mark recommendation as complete when button is clicked', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
        }, { timeout: 5000 });

        // Expand high priority section first
        const highButton = screen.getByRole('button', { name: /high/i });
        await user.click(highButton);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        }, { timeout: 5000 });

        const completeButtons = screen.getAllByRole('button', { name: /mark.*complete/i });
        await user.click(completeButtons[0]);

        await waitFor(() => {
          // Recommendation should be marked as complete (check for visual indicator or toast)
          const completedIndicator = screen.queryByText(/completed/i);
          // Toast might be called - check if available
          const toastCalled = (mockToast?.success?.mock?.calls?.length || 0) > 0;
          expect(completedIndicator || toastCalled).toBeTruthy();
        }, { timeout: 5000 });
      });

      it('should track completed recommendations count', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
        }, { timeout: 5000 });

        // Expand high priority section
        const highButton = screen.getByRole('button', { name: /high/i });
        await user.click(highButton);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        }, { timeout: 5000 });

        const completeButtons = screen.getAllByRole('button', { name: /mark.*complete/i });
        await user.click(completeButtons[0]);

        await waitFor(() => {
          // Should show count of completed recommendations (might be in description or badge)
          const completedText = screen.queryByText(/1.*completed/i) ||
                               screen.queryByText(/completed.*1/i);
          expect(completedText).toBeTruthy();
        }, { timeout: 5000 });
      });
    });

    describe('Filtering by Priority', () => {
      it('should display priority filter dropdown', async () => {
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          // Filter is a Select component - find by placeholder, label, or role
          const filterSelect = screen.queryByRole('combobox') ||
                              screen.queryByText(/all priorities|filter/i) ||
                              document.querySelector('[role="combobox"]');
          expect(filterSelect).toBeTruthy();
        }, { timeout: 5000 });
      });

      it('should filter recommendations by priority when filter is selected', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
        }, { timeout: 5000 });

        // Expand both high and medium sections to see all recommendations initially
        const highButton = screen.getByRole('button', { name: /high/i });
        await user.click(highButton);
        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        }, { timeout: 5000 });

        const mediumButton = screen.getByRole('button', { name: /medium/i });
        await user.click(mediumButton);
        await waitFor(() => {
          expect(screen.getByText('Optimize Operations')).toBeInTheDocument();
        }, { timeout: 5000 });

        // Now filter to show only high priority
        // Find the Select trigger (combobox) - using mocked Select component
        const filterTrigger = screen.getByRole('combobox');
        expect(filterTrigger).toBeInTheDocument();
        
        // Click to open the select dropdown
        await user.click(filterTrigger);
        
        // Wait for options to appear and click "High Priority"
        // Select options are: "All Priorities", "High Priority", "Medium Priority", "Low Priority"
        // Use getByTestId which is more reliable for the mocked Select component
        await waitFor(() => {
          const highOption = screen.getByTestId('select-item-high');
          expect(highOption).toBeInTheDocument();
        }, { timeout: 5000 });
        
        const highOption = screen.getByTestId('select-item-high');
        await user.click(highOption);

        // Wait for filter to be applied
        await waitFor(() => {
          // The filter should be applied - check that the combobox shows "High Priority"
          const filterTrigger = screen.getByRole('combobox');
          expect(filterTrigger).toHaveTextContent(/high priority/i);
        }, { timeout: 3000 });

        // After filtering, we need to expand the high priority section to see the recommendations
        // The sections might have been collapsed when the filter changed
        const highPriorityButton = screen.getByRole('button', { name: /high/i });
        await user.click(highPriorityButton);

        await waitFor(() => {
          // Should only show high priority recommendations
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
          // Medium priority should be filtered out
          expect(screen.queryByText('Optimize Operations')).not.toBeInTheDocument();
        }, { timeout: 5000 });
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
          expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
        }, { timeout: 5000 });

        // Expand high priority section to see recommendations
        const highButton = screen.getByRole('button', { name: /high/i });
        await user.click(highButton);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        }, { timeout: 5000 });

        const searchInput = screen.getByPlaceholderText(/search/i);
        await user.type(searchInput, 'Financial');

        await waitFor(() => {
          // Should only show recommendations matching "Financial"
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
          expect(screen.queryByText('Enhance Compliance Framework')).not.toBeInTheDocument();
        }, { timeout: 3000 });
      });

      it('should search across title, description, type, and action items', async () => {
        const user = userEvent.setup();
        render(<RiskRecommendationsSection merchantId={merchantId} />);

        await waitFor(() => {
          expect(screen.getByText(/risk recommendations/i)).toBeInTheDocument();
        }, { timeout: 5000 });

        // Expand high priority section to see recommendations
        const highButton = screen.getByRole('button', { name: /high/i });
        await user.click(highButton);

        await waitFor(() => {
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        }, { timeout: 5000 });

        const searchInput = screen.getByPlaceholderText(/search/i);
        await user.type(searchInput, 'cash flow');

        await waitFor(() => {
          // Should find recommendation with "cash flow" in action items
          expect(screen.getByText('Improve Financial Stability')).toBeInTheDocument();
        }, { timeout: 3000 });
      });
    });
  });
});

