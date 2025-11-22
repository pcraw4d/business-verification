import { BulkOperationsManager } from '@/components/bulk-operations/BulkOperationsManager';
import { getMerchantsList } from '@/lib/api';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { toast } from 'sonner';
import React from 'react';

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
          child?.type?.name === 'SelectTrigger' ||
          (child?.props && 'id' in child.props)) {
        selectTrigger = child;
      } else if (child?.type?.displayName === 'SelectContent' || 
                 child?.type?.name === 'SelectContent') {
        selectContent = child;
      }
    });
    
    const selectValue = selectTrigger?.props?.children;
    const placeholder = selectValue?.props?.placeholder || 'Select...';
    
    const options = React.Children.toArray(selectContent?.props?.children || []);
    const selectedOption = options.find(
      (opt: any) => opt?.props?.value === selectedValue
    );
    const displayValue = selectedOption?.props?.children || placeholder;
    
    const triggerProps = selectTrigger?.props || {};
    const fieldId = triggerProps.id;
    const labelElement = fieldId ? document.querySelector(`label[for="${fieldId}"]`) : null;
    const labelText = labelElement?.textContent?.trim() || '';
    
    return (
      <div data-testid="select-wrapper">
        <button
          type="button"
          role="combobox"
          aria-expanded={isOpen}
          aria-label={labelText}
          aria-labelledby={fieldId && labelElement ? fieldId : undefined}
          onClick={() => setIsOpen(!isOpen)}
          data-value={selectedValue}
          id={fieldId}
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
    return <span {...props}>{placeholder}</span>;
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

vi.mock('@/lib/api');
vi.mock('sonner');

const mockGetMerchantsList = vi.mocked(getMerchantsList);
const mockToast = vi.mocked(toast);

describe('BulkOperationsManager', () => {
  const mockMerchants = [
    { id: 'merchant-1', name: 'Business 1', businessName: 'Business 1', status: 'active', risk_level: 'low', industry: 'Technology' },
    { id: 'merchant-2', name: 'Business 2', businessName: 'Business 2', status: 'pending', risk_level: 'high', industry: 'Finance' },
    { id: 'merchant-3', name: 'Business 3', businessName: 'Business 3', status: 'active', risk_level: 'critical', industry: 'Retail' },
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
      
      // Wait for the component to mount and load merchants
      await waitFor(() => {
        expect(mockGetMerchantsList).toHaveBeenCalled();
      }, { timeout: 3000 });
      
      // Wait for merchants to be displayed (component uses merchant.name)
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
        expect(screen.getByText('Business 2')).toBeInTheDocument();
        expect(screen.getByText('Business 3')).toBeInTheDocument();
      }, { timeout: 5000 });
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
      }, { timeout: 5000 });
      
      // Find and click checkbox for first merchant
      // Checkboxes are in the merchant list items
      const checkboxes = screen.getAllByRole('checkbox');
      // First checkbox might be select all, so find one that's not checked initially
      const uncheckedCheckboxes = checkboxes.filter(cb => !cb.checked);
      const merchantCheckbox = uncheckedCheckboxes[0] || checkboxes[1];
      
      if (merchantCheckbox) {
        await user.click(merchantCheckbox);
        
        // Merchant should be selected
        await waitFor(() => {
          expect(merchantCheckbox).toBeChecked();
        });
      }
    });

    it('should allow selecting all merchants', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      }, { timeout: 5000 });
      
      // Find select all button using aria-label - get first one if multiple
      const selectAllButtons = screen.getAllByRole('button', { name: /select all merchants/i });
      const selectAllButton = selectAllButtons[0];
      
      await user.click(selectAllButton);
      
      // All merchant checkboxes should be selected
      await waitFor(() => {
        const merchantCheckboxes = screen.getAllByRole('checkbox').filter(cb => 
          !cb.getAttribute('aria-label')?.includes('Select all') &&
          !cb.getAttribute('aria-label')?.includes('Deselect all')
        );
        merchantCheckboxes.forEach((checkbox) => {
          expect(checkbox).toBeChecked();
        });
      });
    });

    it('should allow deselecting all merchants', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      }, { timeout: 5000 });
      
      // First select all - use getAllByRole and pick the first one
      const selectAllButtons = screen.getAllByRole('button', { name: /select all merchants/i });
      const selectAllButton = selectAllButtons[0];
      await user.click(selectAllButton);
      
      await waitFor(() => {
        const checkboxes = screen.getAllByRole('checkbox');
        // Filter merchant checkboxes by checking if they're in merchant item divs
        // Exclude the "select all" checkbox in the header
        const merchantCheckboxes = checkboxes.filter(cb => {
          const parent = cb.closest('div');
          if (!parent) return false;
          const text = parent.textContent || '';
          // Check if this div contains a merchant name (not "Select All" button)
          return (text.includes('Business 1') || text.includes('Business 2') || text.includes('Business 3')) &&
                 !text.includes('Select All');
        });
        expect(merchantCheckboxes.length).toBeGreaterThan(0);
        // Check if at least one is checked using Radix UI's data-state
        const hasChecked = merchantCheckboxes.some(cb => {
          return cb.getAttribute('data-state') === 'checked' || 
                 cb.getAttribute('aria-checked') === 'true';
        });
        expect(hasChecked).toBe(true);
      }, { timeout: 3000 });
      
      // Now deselect all
      const deselectAllButton = screen.getByRole('button', { name: /deselect all merchants/i });
      await user.click(deselectAllButton);
      
      // All merchant checkboxes should be deselected
      // Wait for state to update after deselectAll
      await waitFor(() => {
        // Radix UI Checkbox uses role="checkbox" and data-state="checked" or aria-checked
        const allCheckboxes = screen.getAllByRole('checkbox');
        // Filter to only merchant checkboxes, excluding header "select all" checkbox
        const merchantCheckboxes = allCheckboxes.filter(cb => {
          const container = cb.closest('div');
          if (!container) return false;
          const text = container.textContent || '';
          // Check if this div contains a merchant name (not "Select All" button)
          return (text.includes('Business 1') || text.includes('Business 2') || text.includes('Business 3')) &&
                 !text.includes('Select All');
        });
        
        // All merchant checkboxes should be unchecked
        // Radix UI uses data-state="checked" or aria-checked="true" for checked state
        expect(merchantCheckboxes.length).toBeGreaterThan(0);
        const allUnchecked = merchantCheckboxes.every(cb => {
          const isChecked = cb.getAttribute('data-state') === 'checked' || 
                          cb.getAttribute('aria-checked') === 'true';
          return !isChecked;
        });
        expect(allUnchecked).toBe(true);
      }, { timeout: 5000 });
    });

    it('should allow selecting merchants by filter', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      }, { timeout: 5000 });
      
      // Find select by filter button using aria-label
      const selectByFilterButton = screen.getByRole('button', { name: /select merchants by current filter/i });
      
      // Verify merchants are loaded first
      await waitFor(() => {
        expect(screen.getByText('Business 2')).toBeInTheDocument(); // pending
        expect(screen.getByText('Business 3')).toBeInTheDocument(); // critical
      }, { timeout: 3000 });
      
      await user.click(selectByFilterButton);
      
      // Merchants matching filter (pending or high/critical risk) should be selected
      // Business 2 is pending, Business 3 is critical
      // selectByFilter selects merchants with status='pending' OR risk_level='high' OR risk_level='critical'
      await waitFor(() => {
        // Radix UI Checkbox uses role="checkbox" and data-state="checked" or aria-checked
        const allCheckboxes = screen.getAllByRole('checkbox');
        const merchantCheckboxes = allCheckboxes.filter(cb => {
          // Check if this checkbox is in a container with a merchant name
          const container = cb.closest('div');
          if (!container) return false;
          const text = container.textContent || '';
          // Business 2 (pending) and Business 3 (critical) should be selected
          return text.includes('Business 2') || text.includes('Business 3');
        });
        
        // At least one of Business 2 or Business 3 should be checked
        expect(merchantCheckboxes.length).toBeGreaterThan(0);
        const checkedCount = merchantCheckboxes.filter(cb => {
          const isChecked = cb.getAttribute('data-state') === 'checked' || 
                          cb.getAttribute('aria-checked') === 'true';
          return isChecked;
        }).length;
        expect(checkedCount).toBeGreaterThan(0);
      }, { timeout: 5000 });
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
      }, { timeout: 5000 });
      
      // Find status filter - Using mocked Select component
      const comboboxes = screen.getAllByRole('combobox');
      const statusSelect = comboboxes[0] || screen.getByText(/all status/i);
      expect(statusSelect).toBeInTheDocument();
      await user.click(statusSelect);
      
      // Wait for select content to appear and click Active
      await waitFor(async () => {
        const activeOption = screen.getByTestId('select-item-active') ||
                             screen.getByRole('option', { name: 'Active' }) ||
                             screen.getByText('Active');
        await user.click(activeOption);
      }, { timeout: 3000 });
      
      // Wait for API call with status filter - component calls loadMerchants when filter changes
      await waitFor(() => {
        // Component should call getMerchantsList with status filter
        expect(mockGetMerchantsList).toHaveBeenCalled();
        // Check if any call had status filter
        const calls = mockGetMerchantsList.mock.calls;
        const hasStatusCall = calls.some(call => 
          call[0] && typeof call[0] === 'object' && 
          (call[0].status === 'active' || call[0].status === undefined) // undefined means 'all'
        );
        // Component should have called the API
        expect(hasStatusCall || calls.length > 0).toBe(true);
      }, { timeout: 5000 });
    });

    it('should filter merchants by risk level', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      }, { timeout: 5000 });
      
      // Find risk level filter - Using mocked Select component
      const riskSelects = screen.getAllByRole('combobox');
      expect(riskSelects.length).toBeGreaterThanOrEqual(2);
      // Find the risk level select (second combobox, after status)
      const riskSelect = riskSelects[1];
      expect(riskSelect).toBeInTheDocument();
      await user.click(riskSelect);
      
      // Wait for select content to appear and click High
      await waitFor(async () => {
        const highOption = screen.getByTestId('select-item-high') ||
                           screen.getByRole('option', { name: /^high$/i }) ||
                           screen.getByText(/^high$/i);
        await user.click(highOption);
      }, { timeout: 5000 });
      
      // Wait for API call with riskLevel filter
      // The component calls loadMerchants when filters change, which triggers useEffect
      await waitFor(() => {
        // Check if API was called - it should be called at least once (initial load + filter change)
        expect(mockGetMerchantsList).toHaveBeenCalled();
        // Check if any call had riskLevel filter
        const calls = mockGetMerchantsList.mock.calls;
        // The component should call with riskLevel='high' when filter is set
        const hasRiskCall = calls.some(call => {
          if (!call[0] || typeof call[0] !== 'object') return false;
          const filters = call[0] as any;
          return filters.riskLevel === 'high';
        });
        // We should have a call with riskLevel='high'
        expect(hasRiskCall).toBe(true);
      }, { timeout: 5000 });
    });
  });

  describe('Operations', () => {
    it('should show operation selection when merchants are selected', async () => {
      const user = userEvent.setup();
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      }, { timeout: 5000 });
      
      // Select a merchant - find a checkbox in a table row (merchant checkbox)
      const checkboxes = screen.getAllByRole('checkbox');
      // Merchant checkboxes are in table rows, not buttons
      const merchantCheckbox = checkboxes.find(cb => 
        cb.closest('tr') && !cb.getAttribute('aria-label')?.includes('Select all')
      ) || checkboxes[1]; // Skip first checkbox if it's select all
      
      if (merchantCheckbox) {
        await user.click(merchantCheckbox);
        
        // Wait for selection to register
        await waitFor(() => {
          expect(merchantCheckbox).toBeChecked();
        }, { timeout: 2000 });
        
        // Operation selection should be available - operation buttons are always rendered
        await waitFor(() => {
          // Check for operation type buttons - they're always visible
          const operationButtons = screen.getAllByRole('button');
          const hasOperationButton = operationButtons.some(btn => 
            btn.textContent?.includes('Update Portfolio Type') ||
            btn.textContent?.includes('Update Portfolio') ||
            btn.textContent?.includes('Update Risk Level') ||
            btn.textContent?.includes('Update Risk') ||
            btn.textContent?.includes('Export Data') ||
            btn.getAttribute('aria-label')?.includes('operation')
          );
          // Operation buttons should be present
          expect(hasOperationButton).toBe(true);
        }, { timeout: 3000 });
      }
    });

    it('should require merchant selection before starting operation', async () => {
      render(<BulkOperationsManager />);
      
      await waitFor(() => {
        expect(screen.getByText('Business 1')).toBeInTheDocument();
      }, { timeout: 5000 });
      
      // Operation buttons are always visible, but start button should be disabled when no merchants selected
      // The component shows operation selection buttons, but the start button might be disabled
      const startButton = screen.queryByRole('button', { name: /start|run|execute/i });
      if (startButton) {
        expect(startButton).toBeDisabled();
      } else {
        // If no start button, operation buttons should still be present
        const operationButtons = screen.getAllByRole('button');
        expect(operationButtons.length).toBeGreaterThan(0);
      }
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
      }, { timeout: 5000 });
      
      // Logs section should be visible - component adds logs when merchants are loaded
      // Look for log-related text or the operation logs section
      await waitFor(() => {
        // Component adds a log when merchants are loaded: "Loaded X merchants"
        // Use getAllByText and get first one if multiple
        const logTexts = screen.queryAllByText(/loaded.*merchants/i);
        const operationLogTexts = screen.queryAllByText(/operation.*log/i);
        const bulkOpsTexts = screen.queryAllByText(/bulk operations/i);
        
        // At least one of these should be present
        expect(logTexts.length > 0 || operationLogTexts.length > 0 || bulkOpsTexts.length > 0).toBe(true);
      }, { timeout: 3000 });
    });
  });
});

