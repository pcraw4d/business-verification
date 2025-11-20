import { setupServer } from 'msw/node';
import { handlers } from './handlers';

// Setup MSW server for Node.js (Jest) environment
export const server = setupServer(...handlers);

// Add error event listeners to debug MSW issues
if (process.env.NODE_ENV === 'test') {
  server.events.on('request:start', ({ request }) => {
    console.log('[MSW] Request started:', request.method, request.url);
  });

  server.events.on('request:match', ({ request }) => {
    console.log('[MSW] Request matched:', request.method, request.url);
  });

  server.events.on('request:unhandled', ({ request }) => {
    console.warn('[MSW] Unhandled request:', request.method, request.url);
  });

  // @ts-expect-error - MSW types don't include 'request:exception' event
  server.events.on('request:exception', ({ request, error }: { request: Request; error: Error }) => {
    console.error('[MSW] Request exception:', request.method, request.url);
    console.error('[MSW] Exception error:', error);
    console.error('[MSW] Exception stack:', error?.stack);
  });
  
  server.events.on('response:mocked', ({ request, response }) => {
    console.log('[MSW] Response mocked:', request.method, request.url);
    console.log('[MSW] Response status:', response.status);
    console.log('[MSW] Response ok:', response.ok);
    console.log('[MSW] Response statusText:', response.statusText);
    // Check if response has body
    if (response.body) {
      console.log('[MSW] Response has body');
    } else {
      console.log('[MSW] Response has NO body');
    }
    // Log response type to debug
    console.log('[MSW] Response type:', response.type);
    console.log('[MSW] Response constructor:', response.constructor.name);
  });
  
  server.events.on('response:bypass', ({ request }) => {
    console.log('[MSW] Response bypassed:', request.method, request.url);
  });
}

