/**
 * API Mocking Blueprint
 * 
 * Simple fetch patching for the few cases where we need to mock external calls.
 * Currently GameViewerPage uses WASM for validation, so minimal mocking needed.
 */

import { Page } from '@playwright/test';

/**
 * Install basic API mocking pattern
 * 
 * Usage: Add more endpoints as needed when we discover them during testing
 */
export async function installBasicApiMocking(page: Page): Promise<void> {
  await page.addInitScript(() => {
    const originalFetch = window.fetch;
    
    window.fetch = function(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
      const url = typeof input === 'string' ? input : input.toString();
      
      // Example: Mock save-game endpoint (add more as discovered)
      if (url.includes('/api/save-game')) {
        console.log('🎭 Mocked save-game call');
        return Promise.resolve(new Response(JSON.stringify({ saved: true }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' }
        }));
      }
      
      // TODO: Add more mocks here as we discover external API calls during testing
      // Example patterns:
      // if (url.includes('/api/analytics')) { ... }
      // if (url.includes('/external-service')) { ... }
      
      // Default: Use real fetch (keeps WASM and internal calls real)
      return originalFetch.call(this, input, init);
    };
  });
}

/**
 * Instructions for adding new mocks:
 * 
 * 1. Run tests and watch console for network calls
 * 2. Identify which calls are external/problematic
 * 3. Add pattern matching in the fetch interceptor above
 * 4. Return appropriate mock response
 * 
 * Example:
 * if (url.includes('/api/new-endpoint')) {
 *   return Promise.resolve(new Response(JSON.stringify({ data: 'mock' }), {
 *     status: 200,
 *     headers: { 'Content-Type': 'application/json' }
 *   }));
 * }
 */