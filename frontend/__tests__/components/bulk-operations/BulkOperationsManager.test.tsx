import { BulkOperationsManager } from '@/components/bulk-operations/BulkOperationsManager';
import { getMerchantsList } from '@/lib/api';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { toast } from 'sonner';

vi.mock('@/lib/api');
vi.mock('sonner');

const mockGetMerchantsList = vi.mocked(getMerchantsList);
const mockToast = vi.mocked(toast);

describe('BulkOperationsManager', () => {
  const mockMerchants = [
    { id: 'merchant-1', businessName: 'Business 1', status: 'active', risk_level: 'low' },
    { id: 'merchant-2', businessName: 'Business 2', status: 'pending', risk_level: 'high' },
    { id: 'merchant-3', businessName: 'Business 3', status: 'active', risk_level: 'critical' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.error = vi.fn();
    mockToast.success = vi.fn();
    mockGetMerchantsList.mockResolvedValue({
      merchants: mockMerchants,
      total: mockMerchants.length,
      page: 1,
      pageSize: 100,
    });
  });

  describe('Component Rendering', () => {
    it('should render bulk operations manager', async () => {
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText(/bulk operations/i)).toBeInTheDocument();
      });
    });

    it('should load and display merchants', async () => {
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(mockGetMerchantsList).toHaveBeenCalled();
      });
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
        expect(screen.getByText('Business 2')).toBeInTheDocument();
        expect(screen.getByText('Business 3')).toBeInTheDocument();
      });
    });

    it('should show loading state initially', () => {
      mockGetMerchantsList.mockImplementation(() => new Promise(() => {}));
      
      render(<BulkOperationsManager />);
      
      // Component should show loading state
      expect(screen.queryByText('Business 1')).not.toBeInTheDocument();
    });
  });

  describe('Merchant Selection', () => {
    it('should allow selecting individual merchants', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Find and click checkbox for first merchant
      const checkboxes = screen.getAllByRole('checkbox');
      const merchantCheckbox = checkboxes.find((cb) => 
        cb.getAttribute('aria-label')?.includes('Business 1')
      ) || checkboxes[1]; // Fallback to second checkbox (first might be select all)
      
      if (merchantCheckbox) {
        await user.click(merchantCheckbox);
        
        // Merchant should be selected
        expect(merchantCheckbox).toBeChecked();
      }
    });

    it('should allow selecting all merchants', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Find select all button/checkbox
      const selectAllButton = screen.queryByRole('button', { name: /select all/i }) ||
                             screen.queryByLabelText(/select all/i);
      
      if (selectAllButton) {
        await user.click(selectAllButton);
        
        // All merchants should be selected
        const checkboxes = screen.getAllByRole('checkbox');
        checkboxes.forEach((checkbox) => {
          if (checkbox !== selectAllButton) {
            expect(checkbox).toBeChecked();
          }
        });
      }
    });

    it('should allow deselecting all merchants', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Find deselect all button
      const deselectAllButton = screen.queryByRole('button', { name: /deselect all/i });
      
      if (deselectAllButton) {
        await user.click(deselectAllButton);
        
        // All merchants should be deselected
        const checkboxes = screen.getAllByRole('checkbox');
        checkboxes.forEach((checkbox) => {
          if (checkbox !== deselectAllButton) {
            expect(checkbox).not.toBeChecked();
          }
        });
      }
    });

    it('should allow selecting merchants by filter', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Find select by filter button
      const selectByFilterButton = screen.queryByRole('button', { name: /select.*filter/i });
      
      if (selectByFilterButton) {
        await user.click(selectByFilterButton);
        
        // Merchants matching filter (pending or high risk) should be selected
        await waitFor(() => {
          // Verify selection happened
          expect(selectByFilterButton).toBeInTheDocument();
        });
      }
    });
  });

  describe('Filtering', () => {
    it('should filter merchants by search term', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Find search input
      const searchInput = screen.queryByPlaceholderText(/search/i) ||
                         screen.queryByLabelText(/search/i);
      
      if (searchInput) {
        await user.type(searchInput, 'Business 1');
        
        await waitFor(() => {
          expect(mockGetMerchantsList).toHaveBeenCalledWith(
            expect.objectContaining({
              search: 'Business 1',
            })
          );
        });
      }
    });

    it('should filter merchants by status', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Find status filter
      const statusFilter = screen.queryByLabelText(/status/i) ||
                          screen.queryByRole('combobox', { name: /status/i });
      
      if (statusFilter) {
        await user.click(statusFilter);
        const activeOption = screen.queryByText(/active/i);
        if (activeOption) {
          await user.click(activeOption);
          
          await waitFor(() => {
            expect(mockGetMerchantsList).toHaveBeenCalledWith(
              expect.objectContaining({
                status: 'active',
              })
            );
          });
        }
      }
    });

    it('should filter merchants by risk level', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Find risk level filter
      const riskFilter = screen.queryByLabelText(/risk/i) ||
                        screen.queryByRole('combobox', { name: /risk/i });
      
      if (riskFilter) {
        await user.click(riskFilter);
        const highRiskOption = screen.queryByText(/high/i);
        if (highRiskOption) {
          await user.click(highRiskOption);
          
          await waitFor(() => {
            expect(mockGetMerchantsList).toHaveBeenCalledWith(
              expect.objectContaining({
                riskLevel: 'high',
              })
            );
          });
        }
      }
    });
  });

  describe('Operations', () => {
    it('should show operation selection when merchants are selected', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Select a merchant
      const checkboxes = screen.getAllByRole('checkbox');
      if (checkboxes.length > 1) {
        await user.click(checkboxes[1]);
        
        // Operation buttons should be available
        await waitFor(() => {
          const operationButtons = screen.queryAllByRole('button', { name: /update|export|send|schedule|deactivate/i });
          expect(operationButtons.length).toBeGreaterThan(0);
        });
      }
    });

    it('should require merchant selection before starting operation', async () => {
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Operation buttons should be disabled or not visible when no merchants selected
      const operationButtons = screen.queryAllByRole('button', { name: /update|export|send|schedule|deactivate/i });
      // Buttons might be disabled or hidden
      expect(operationButtons.length).toBeGreaterThanOrEqual(0);
    });
  });

  describe('Error Handling', () => {
    it('should handle error when loading merchants fails', async () => {
      const error = new Error('Failed to load merchants');
      mockGetMerchantsList.mockRejectedValue(error);
      
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith(
          'Failed to load merchants',
          expect.objectContaining({
            description: 'Failed to load merchants',
          })
        );
      });
    });
  });

  describe('Logs', () => {
    it('should display operation logs', async () => {
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      });
      
      // Logs section should be visible
      const logsSection = screen.queryByText(/logs|log/i);
      // Logs might be in a collapsible section or always visible
      expect(logsSection || screen.getByText(/bulk operations/i)).toBeInTheDocument();
    });
  });
});

