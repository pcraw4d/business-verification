import { MerchantForm } from '@/components/forms/MerchantForm';
import { createMerchant } from '@/lib/api';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';

// Mock dependencies
vi.mock('next/navigation');
vi.mock('@/lib/api');
vi.mock('sonner');

const mockRouter = {
  push: vi.fn(),
  replace: vi.fn(),
  prefetch: vi.fn(),
  back: vi.fn(),
  pathname: '/',
  query: {},
  asPath: '/',
};

const mockCreateMerchant = vi.mocked(createMerchant);
const mockToast = vi.mocked(toast);
const mockUseRouter = vi.mocked(useRouter);

describe('MerchantForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseRouter.mockReturnValue(mockRouter as any);
    mockToast.error = vi.fn();
    mockToast.success = vi.fn();
    mockToast.info = vi.fn();
    
    // Mock sessionStorage
    Object.defineProperty(window, 'sessionStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      writable: true,
    });
  });

  describe('Form Rendering', () => {
    it('should render form with all fields', () => {
      render(<MerchantForm />);
      
      expect(screen.getByLabelText(/business name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/website url/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/street address/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/city/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/state/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/postal code/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/country/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/phone number/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/registration number/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/analysis type/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/risk assessment type/i)).toBeInTheDocument();
    });

    it('should render submit button', () => {
      render(<MerchantForm />);
      
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      expect(submitButton).toBeInTheDocument();
      expect(submitButton).toHaveAttribute('type', 'submit');
    });

    it('should render clear button', () => {
      render(<MerchantForm />);
      
      const clearButton = screen.getByRole('button', { name: /clear form/i });
      expect(clearButton).toBeInTheDocument();
      expect(clearButton).toHaveAttribute('type', 'button');
    });

    it('should have default values for analysis and assessment types', () => {
      render(<MerchantForm />);
      
      // Check that select fields have default values
      const analysisSelect = screen.getByLabelText(/analysis type/i);
      const assessmentSelect = screen.getByLabelText(/risk assessment type/i);
      
      expect(analysisSelect).toBeInTheDocument();
      expect(assessmentSelect).toBeInTheDocument();
    });
  });

  describe('Form Validation', () => {
    it('should show error when submitting empty required fields', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Please fix the errors in the form');
      });
    });

    it('should validate business name is required', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      const businessNameField = screen.getByLabelText(/business name/i);
      await user.type(businessNameField, '');
      await user.tab();
      
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalled();
      });
    });

    it('should validate country is required', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      const businessNameField = screen.getByLabelText(/business name/i);
      await user.type(businessNameField, 'Test Business');
      
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalled();
      });
    });

    it('should validate email format', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      const emailField = screen.getByLabelText(/email address/i);
      await user.type(emailField, 'invalid-email');
      await user.tab();
      
      // Email validation should show error
      await waitFor(() => {
        const errorMessage = screen.queryByText(/invalid.*email/i);
        if (errorMessage) {
          expect(errorMessage).toBeInTheDocument();
        }
      });
    });

    it('should validate URL format', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      const websiteField = screen.getByLabelText(/website url/i);
      await user.type(websiteField, 'not-a-url');
      await user.tab();
      
      // URL validation should show error
      await waitFor(() => {
        const errorMessage = screen.queryByText(/invalid.*url/i);
        if (errorMessage) {
          expect(errorMessage).toBeInTheDocument();
        }
      });
    });

    it('should clear errors when user types in field', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      const businessNameField = screen.getByLabelText(/business name/i);
      
      // Trigger validation error
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      // Type in field to clear error
      await user.type(businessNameField, 'Test Business');
      
      // Error should be cleared
      await waitFor(() => {
        const errorMessage = screen.queryByText(/business name.*required/i);
        expect(errorMessage).not.toBeInTheDocument();
      });
    });
  });

  describe('Form Submission', () => {
    it('should submit form with valid data', async () => {
      const user = userEvent.setup();
      mockCreateMerchant.mockResolvedValue({ id: 'merchant-123' });
      
      render(<MerchantForm />);
      
      // Fill in required fields
      await user.type(screen.getByLabelText(/business name/i), 'Test Business');
      await user.type(screen.getByLabelText(/country/i), 'US');
      
      // Select country from dropdown
      const countrySelect = screen.getByLabelText(/country/i);
      await user.click(countrySelect);
      await user.click(screen.getByText('United States'));
      
      // Submit form
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(mockCreateMerchant).toHaveBeenCalled();
      });
    });

    it('should show success toast on successful submission', async () => {
      const user = userEvent.setup();
      mockCreateMerchant.mockResolvedValue({ id: 'merchant-123' });
      
      render(<MerchantForm />);
      
      // Fill in required fields
      await user.type(screen.getByLabelText(/business name/i), 'Test Business');
      
      // Select country
      const countrySelect = screen.getByLabelText(/country/i);
      await user.click(countrySelect);
      await user.click(screen.getByText('United States'));
      
      // Submit form
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(mockToast.success).toHaveBeenCalledWith(
          'Merchant created successfully',
          expect.objectContaining({
            description: expect.stringContaining('merchant-123'),
          })
        );
      });
    });

    it('should redirect to merchant details page on success', async () => {
      const user = userEvent.setup();
      mockCreateMerchant.mockResolvedValue({ id: 'merchant-123' });
      
      render(<MerchantForm />);
      
      // Fill in required fields
      await user.type(screen.getByLabelText(/business name/i), 'Test Business');
      
      // Select country
      const countrySelect = screen.getByLabelText(/country/i);
      await user.click(countrySelect);
      await user.click(screen.getByText('United States'));
      
      // Submit form
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(mockRouter.push).toHaveBeenCalledWith('/merchant-details/merchant-123');
      });
    });

    it('should store merchant data in sessionStorage on success', async () => {
      const user = userEvent.setup();
      mockCreateMerchant.mockResolvedValue({ id: 'merchant-123' });
      
      render(<MerchantForm />);
      
      // Fill in required fields
      await user.type(screen.getByLabelText(/business name/i), 'Test Business');
      
      // Select country
      const countrySelect = screen.getByLabelText(/country/i);
      await user.click(countrySelect);
      await user.click(screen.getByText('United States'));
      
      // Submit form
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(window.sessionStorage.setItem).toHaveBeenCalledWith(
          'merchantId',
          'merchant-123'
        );
        expect(window.sessionStorage.setItem).toHaveBeenCalledWith(
          'merchantData',
          expect.any(String)
        );
      });
    });

    it('should show error toast on submission failure', async () => {
      const user = userEvent.setup();
      const error = new Error('Failed to create merchant');
      mockCreateMerchant.mockRejectedValue(error);
      
      render(<MerchantForm />);
      
      // Fill in required fields
      await user.type(screen.getByLabelText(/business name/i), 'Test Business');
      
      // Select country
      const countrySelect = screen.getByLabelText(/country/i);
      await user.click(countrySelect);
      await user.click(screen.getByText('United States'));
      
      // Submit form
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to create merchant');
      });
    });

    it('should disable submit button while submitting', async () => {
      const user = userEvent.setup();
      let resolveCreateMerchant: (value: any) => void;
      const createMerchantPromise = new Promise((resolve) => {
        resolveCreateMerchant = resolve;
      });
      mockCreateMerchant.mockReturnValue(createMerchantPromise as any);
      
      render(<MerchantForm />);
      
      // Fill in required fields
      await user.type(screen.getByLabelText(/business name/i), 'Test Business');
      
      // Select country
      const countrySelect = screen.getByLabelText(/country/i);
      await user.click(countrySelect);
      await user.click(screen.getByText('United States'));
      
      // Submit form
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      // Button should be disabled and show processing text
      await waitFor(() => {
        expect(submitButton).toBeDisabled();
        expect(screen.getByText(/processing/i)).toBeInTheDocument();
      });
      
      // Resolve the promise
      resolveCreateMerchant!({ id: 'merchant-123' });
      
      await waitFor(() => {
        expect(submitButton).not.toBeDisabled();
      });
    });
  });

  describe('Form Clearing', () => {
    it('should clear all form fields when clear button is clicked', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      // Fill in some fields
      await user.type(screen.getByLabelText(/business name/i), 'Test Business');
      await user.type(screen.getByLabelText(/email address/i), 'test@example.com');
      
      // Click clear button
      const clearButton = screen.getByRole('button', { name: /clear form/i });
      await user.click(clearButton);
      
      // Fields should be cleared
      await waitFor(() => {
        const businessNameField = screen.getByLabelText(/business name/i) as HTMLInputElement;
        const emailField = screen.getByLabelText(/email address/i) as HTMLInputElement;
        
        expect(businessNameField.value).toBe('');
        expect(emailField.value).toBe('');
      });
    });

    it('should show info toast when form is cleared', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      // Click clear button
      const clearButton = screen.getByRole('button', { name: /clear form/i });
      await user.click(clearButton);
      
      await waitFor(() => {
        expect(mockToast.info).toHaveBeenCalledWith('Form cleared');
      });
    });

    it('should clear errors when form is cleared', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      // Trigger validation error
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      // Clear form
      const clearButton = screen.getByRole('button', { name: /clear form/i });
      await user.click(clearButton);
      
      // Errors should be cleared
      await waitFor(() => {
        const errorMessages = screen.queryAllByRole('alert');
        expect(errorMessages.length).toBe(0);
      });
    });
  });

  describe('Field Updates', () => {
    it('should update form data when fields are changed', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      const businessNameField = screen.getByLabelText(/business name/i);
      await user.type(businessNameField, 'New Business Name');
      
      expect((businessNameField as HTMLInputElement).value).toBe('New Business Name');
    });

    it('should handle select field changes', async () => {
      const user = userEvent.setup();
      render(<MerchantForm />);
      
      const analysisSelect = screen.getByLabelText(/analysis type/i);
      await user.click(analysisSelect);
      await user.click(screen.getByText('Risk Assessment'));
      
      // Value should be updated
      await waitFor(() => {
        expect(analysisSelect).toBeInTheDocument();
      });
    });
  });

  describe('Address Building', () => {
    it('should build address string from address fields', async () => {
      const user = userEvent.setup();
      mockCreateMerchant.mockResolvedValue({ id: 'merchant-123' });
      
      render(<MerchantForm />);
      
      // Fill in address fields
      await user.type(screen.getByLabelText(/business name/i), 'Test Business');
      await user.type(screen.getByLabelText(/street address/i), '123 Main St');
      await user.type(screen.getByLabelText(/city/i), 'San Francisco');
      await user.type(screen.getByLabelText(/state/i), 'CA');
      await user.type(screen.getByLabelText(/postal code/i), '94102');
      
      // Select country
      const countrySelect = screen.getByLabelText(/country/i);
      await user.click(countrySelect);
      await user.click(screen.getByText('United States'));
      
      // Submit form
      const submitButton = screen.getByRole('button', { name: /verify merchant/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(mockCreateMerchant).toHaveBeenCalledWith(
          expect.objectContaining({
            address: expect.objectContaining({
              street: '123 Main St',
              city: 'San Francisco',
              state: 'CA',
              postal_code: '94102',
              country: 'US',
            }),
          })
        );
      });
    });
  });
});

