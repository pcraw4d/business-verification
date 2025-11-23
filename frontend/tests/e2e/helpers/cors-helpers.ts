/**
 * CORS helper functions for E2E tests
 * Provides consistent CORS handling across all test files
 */

/**
 * Helper function to handle CORS OPTIONS preflight requests
 * Returns true if OPTIONS was handled, false otherwise
 */
export async function handleCorsOptions(route: any): Promise<boolean> {
  if (route.request().method() === 'OPTIONS') {
    await route.fulfill({
      status: 200,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      },
    });
    return true;
  }
  return false;
}

/**
 * Helper function to get CORS headers for API responses
 */
export function getCorsHeaders(): Record<string, string> {
  return {
    'Access-Control-Allow-Origin': '*',
  };
}

