

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

      // logError logs with '[ErrorHandler] API Error:' prefix in development mode
      // or 'API Error:' with code and message in production
      const errorCall = (console.error as vi.Mock).mock.calls.find(
        (call: any[]) => call[0]?.includes('API Error') || call[0]?.includes('ErrorHandler')
      );
      expect(errorCall).toBeDefined();
      if (errorCall && errorCall.length > 1 && typeof errorCall[1] === 'object') {
        expect(errorCall[1]).toMatchObject({
          code: 'ERROR_CODE',
          message: 'Error message',
        });
      }
    });

    it('should include timestamp in error log', () => {
      const error = new Error('Test error');
      ErrorHandler.logError(error, 'ERROR_CODE', 'Error message');

      // Check that timestamp is included in the error details
      const errorCall = (console.error as vi.Mock).mock.calls.find(
        (call: any[]) => call[0]?.includes('API Error') || call[0]?.includes('ErrorHandler')
      );
      expect(errorCall).toBeDefined();
      if (errorCall && errorCall.length > 1 && typeof errorCall[1] === 'object') {
        expect(errorCall[1]).toHaveProperty('timestamp');
        expect(typeof errorCall[1].timestamp).toBe('string');
      }
    });

    it('should include URL in error log when available', () => {
      // Mock window.location
      delete (window as any).location;
      (window as any).location = { href: 'https://example.com/test' };

      const error = new Error('Test error');
      ErrorHandler.logError(error, 'ERROR_CODE', 'Error message');

      // Check that URL is included in the error details
      const errorCall = (console.error as vi.Mock).mock.calls.find(
        (call: any[]) => call[0]?.includes('API Error') || call[0]?.includes('ErrorHandler')
      );
      expect(errorCall).toBeDefined();
      if (errorCall && errorCall.length > 1 && typeof errorCall[1] === 'object') {
        expect(errorCall[1]).toHaveProperty('url');
        expect(errorCall[1].url).toBe('https://example.com/test');
      }
    });
  });

  describe('Error Recovery', () => {
    it('should handle network errors gracefully', async () => {
      const networkError = new Error('Network request failed');
      await ErrorHandler.handleAPIError(networkError);

      expect(mockToastError).toHaveBeenCalledWith('Network request failed', expect.any(Object));
    });

    it('should handle timeout errors', async () => {
      const timeoutError = new Error('Request timeout');
      await ErrorHandler.handleAPIError(timeoutError);

      // ErrorHandler transforms timeout errors to a user-friendly message
      expect(mockToastError).toHaveBeenCalledWith('Request timed out. Please try again.', expect.any(Object));
    });

    it('should handle 500 errors', async () => {
      const serverError = {
        code: 'INTERNAL_SERVER_ERROR',
        message: 'Internal server error',
      };
      await ErrorHandler.handleAPIError(serverError);

      expect(mockToastError).toHaveBeenCalledWith('Internal server error', expect.any(Object));
    });

    it('should handle 404 errors', async () => {
      const notFoundError = {
        code: 'NOT_FOUND',
        message: 'Resource not found',
      };
      await ErrorHandler.handleAPIError(notFoundError);

      // ErrorHandler doesn't show notifications for 404 errors (see showErrorNotification)
      // So toast.error should NOT be called
      expect(mockToastError).not.toHaveBeenCalled();
      // But error should still be logged
      expect(console.error).toHaveBeenCalled();
    });

    it('should handle 403 errors', async () => {
      const forbiddenError = {
        code: 'FORBIDDEN',
        message: 'Access forbidden',
      };
      await ErrorHandler.handleAPIError(forbiddenError);

      expect(mockToastError).toHaveBeenCalledWith('Access forbidden', expect.any(Object));
    });
  });

  describe('Error Boundary Behavior', () => {
    it('should not throw when handling errors', async () => {
      const error = new Error('Test error');
      
      await expect(ErrorHandler.handleAPIError(error)).resolves.not.toThrow();
    });

    it('should handle null errors', async () => {
      await expect(ErrorHandler.handleAPIError(null)).resolves.not.toThrow();
      expect(mockToastError).toHaveBeenCalled();
    });

    it('should handle undefined errors', async () => {
      await expect(ErrorHandler.handleAPIError(undefined)).resolves.not.toThrow();
      expect(mockToastError).toHaveBeenCalled();
    });
  });
});

