

// Mock sonner using manual mock
vi.mock('sonner');

// Import after mock setup
import { ErrorHandler } from '@/lib/error-handler';
import { toast } from 'sonner';

// Get the mocked functions
const mockToastError = toast.error as vi.Mock;
const mockToastSuccess = toast.success as vi.Mock;
const mockToastInfo = toast.info as vi.Mock;

describe('ErrorHandler', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockToastError.mockClear();
    mockToastSuccess.mockClear();
    mockToastInfo.mockClear();
    vi.spyOn(console, 'error').mockImplementation();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('handleAPIError', () => {
    it('should handle Error objects', async () => {
      const error = new Error('Test error');
      await ErrorHandler.handleAPIError(error);

      expect(console.error).toHaveBeenCalled();
      expect(mockToastError).toHaveBeenCalledWith('Test error', expect.any(Object));
    });

    it('should handle APIErrorResponse objects', async () => {
      const error = {
        code: 'API_ERROR',
        message: 'API error message',
        details: {},
      };
      await ErrorHandler.handleAPIError(error);

      expect(mockToastError).toHaveBeenCalledWith('API error message', expect.any(Object));
    });

    it('should handle unknown error types', async () => {
      const error = 'String error';
      await ErrorHandler.handleAPIError(error);

      expect(mockToastError).toHaveBeenCalledWith('An unexpected error occurred', expect.any(Object));
    });
  });

  describe('showErrorNotification', () => {
    it('should show error toast', () => {
      ErrorHandler.showErrorNotification('Error message', 'ERROR_CODE');
      expect(mockToastError).toHaveBeenCalledWith('Error message', {
        description: 'Error Code: ERROR_CODE',
        duration: 5000,
      });
    });

    it('should show error toast without code', () => {
      ErrorHandler.showErrorNotification('Error message');
      expect(mockToastError).toHaveBeenCalledWith('Error message', {
        description: undefined,
        duration: 5000,
      });
    });
  });

  describe('showSuccessNotification', () => {
    it('should show success toast', () => {
      ErrorHandler.showSuccessNotification('Success message');
      expect(mockToastSuccess).toHaveBeenCalledWith('Success message', {
        duration: 3000,
      });
    });
  });

  describe('showInfoNotification', () => {
    it('should show info toast', () => {
      ErrorHandler.showInfoNotification('Info message');
      expect(mockToastInfo).toHaveBeenCalledWith('Info message', {
        duration: 3000,
      });
    });
  });

  describe('parseErrorResponse', () => {
    it('should parse JSON error response', async () => {
      // Create a mock Response object
      const mockJson = vi.fn().mockResolvedValue({
        code: 'ERROR_CODE',
        message: 'Error message',
      });
      const response = {
        json: mockJson,
        status: 400,
        statusText: 'Bad Request',
      } as unknown as Response;

      const result = await ErrorHandler.parseErrorResponse(response);
      expect(result).toEqual({
        code: 'ERROR_CODE',
        message: 'Error message',
      });
      expect(mockJson).toHaveBeenCalled();
    });

    it('should handle parse errors', async () => {
      // Create a mock Response that throws on json()
      const mockJson = vi.fn().mockRejectedValue(new Error('Invalid JSON'));
      const response = {
        json: mockJson,
        status: 500,
        statusText: 'Internal Server Error',
      } as unknown as Response;

      const result = await ErrorHandler.parseErrorResponse(response);
      expect(result).toEqual({
        code: 'PARSE_ERROR',
        message: 'HTTP 500: Internal Server Error',
      });
    });
  });

  describe('logError', () => {
    it('should log error details', () => {
      const error = new Error('Test error');
      ErrorHandler.logError(error, 'ERROR_CODE', 'Error message');

      expect(console.error).toHaveBeenCalledWith('API Error:', expect.objectContaining({
        code: 'ERROR_CODE',
        message: 'Error message',
        error: error,
      }));
    });
  });
});

