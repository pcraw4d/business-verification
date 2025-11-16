// Manual mock for error-handler
export const ErrorHandler = {
  handleAPIError: jest.fn().mockResolvedValue(undefined),
  showErrorNotification: jest.fn(),
  showSuccessNotification: jest.fn(),
  showInfoNotification: jest.fn(),
  parseErrorResponse: jest.fn().mockResolvedValue({ code: 'TEST_ERROR', message: 'Test error' }),
  logError: jest.fn(),
};
