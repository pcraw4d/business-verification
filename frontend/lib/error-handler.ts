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
    const errorDetails = {
      code: code || 'UNKNOWN_ERROR',
      message: message || (error instanceof Error ? error.message : String(error)),
      error: error,
      timestamp: new Date().toISOString(),
      url: typeof window !== 'undefined' ? window.location.href : '',
    };

    // Suppress 404 errors for optional endpoints - they're expected if endpoints aren't implemented
    const is404 = message?.includes('404') || (error instanceof Error && error.message.includes('404'));
    if (!is404) {
      console.error('API Error:', errorDetails);
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

