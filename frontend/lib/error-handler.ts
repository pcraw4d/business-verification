// Error handling utility for frontend
import { toast } from 'sonner';

export interface APIErrorResponse {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export class ErrorHandler {
  /**
   * Handles API errors and displays appropriate notifications
   */
  static async handleAPIError(error: unknown): Promise<void> {
    let errorMessage = 'An unexpected error occurred';
    let errorCode = 'UNKNOWN_ERROR';

    if (error instanceof Error) {
      errorMessage = error.message;
      
      // Detect CORS errors
      if (errorMessage.includes('CORS') || errorMessage.includes('Access-Control-Allow-Origin')) {
        errorCode = 'CORS_ERROR';
        errorMessage = 'CORS policy blocked the request. Please check server configuration.';
      }
      // Detect network errors
      else if (errorMessage.includes('Failed to fetch') || errorMessage.includes('NetworkError')) {
        errorCode = 'NETWORK_ERROR';
        errorMessage = 'Network request failed. Please check your connection and try again.';
      }
      // Detect timeout errors
      else if (errorMessage.includes('timeout') || errorMessage.includes('Timeout')) {
        errorCode = 'TIMEOUT_ERROR';
        errorMessage = 'Request timed out. Please try again.';
      }
    } else if (typeof error === 'object' && error !== null) {
      const apiError = error as APIErrorResponse;
      errorMessage = apiError.message || errorMessage;
      errorCode = apiError.code || errorCode;
    }

    // Log error for debugging
    this.logError(error, errorCode, errorMessage);

    // Show error notification
    this.showErrorNotification(errorMessage, errorCode);
  }

  /**
   * Shows an error notification to the user
   */
  static showErrorNotification(message: string, code?: string): void {
    // Don't show notifications for 404s on optional endpoints
    const is404 = message.includes('404') || code === 'NOT_FOUND';
    if (!is404) {
    toast.error(message, {
      description: code ? `Error Code: ${code}` : undefined,
      duration: 5000,
    });
    }
  }

  /**
   * Shows a success notification
   */
  static showSuccessNotification(message: string): void {
    toast.success(message, {
      duration: 3000,
    });
  }

  /**
   * Shows an info notification
   */
  static showInfoNotification(message: string): void {
    toast.info(message, {
      duration: 3000,
    });
  }

  /**
   * Logs error to console and optionally to logging service
   */
  static logError(error: unknown, code?: string, message?: string): void {
    // Stringify error object for better logging
    let errorString = 'Unknown error';
    if (error instanceof Error) {
      errorString = error.message;
      if (error.stack) {
        errorString += `\nStack: ${error.stack}`;
      }
    } else if (typeof error === 'object' && error !== null) {
      try {
        errorString = JSON.stringify(error, null, 2);
      } catch {
        errorString = String(error);
      }
    } else {
      errorString = String(error);
    }

    const errorDetails = {
      code: code || 'UNKNOWN_ERROR',
      message: message || errorString,
      error: errorString,
      timestamp: new Date().toISOString(),
      url: typeof window !== 'undefined' ? window.location.href : '',
    };

    // Suppress 404 errors for optional endpoints - they're expected if endpoints aren't implemented
    const is404 = message?.includes('404') || (error instanceof Error && error.message.includes('404'));
    if (!is404) {
      if (process.env.NODE_ENV === 'development') {
        console.error('[ErrorHandler] API Error:', errorDetails);
        if (error instanceof Error && error.stack) {
          console.error('[ErrorHandler] Stack trace:', error.stack);
        }
      } else {
        console.error('API Error:', errorDetails.code, errorDetails.message);
      }
    }

    // TODO: Send to logging service (e.g., Sentry, LogRocket)
    // if (typeof window !== 'undefined' && window.Sentry) {
    //   window.Sentry.captureException(error, { extra: errorDetails });
    // }
  }

  /**
   * Parses error response from API
   */
  static parseErrorResponse(response: Response): Promise<APIErrorResponse> {
    return response.json().catch(() => ({
      code: 'PARSE_ERROR',
      message: `HTTP ${response.status}: ${response.statusText}`,
    }));
  }
}

