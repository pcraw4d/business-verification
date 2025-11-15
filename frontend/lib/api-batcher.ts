// API Request Batching Utility
type RequestFunction<T> = () => Promise<T>;

interface PendingRequest<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
  reject: (error: Error) => void;
}

export class APIBatcher {
  private pendingRequests: Map<string, PendingRequest<unknown>>;
  private batchTimeout: number;

  constructor(batchTimeout: number = 100) {
    this.pendingRequests = new Map();
    this.batchTimeout = batchTimeout;
  }

  /**
   * Batches requests within the timeout window
   */
  async batchRequest<T>(key: string, requestFn: RequestFunction<T>): Promise<T> {
    // Check if request is already pending
    const existing = this.pendingRequests.get(key);
    if (existing) {
      return existing.promise as Promise<T>;
    }

    // Create new promise
    let resolve!: (value: T) => void;
    let reject!: (error: Error) => void;
    const promise = new Promise<T>((res, rej) => {
      resolve = res;
      reject = rej;
    });

    // Store pending request
    this.pendingRequests.set(key, {
      promise: promise as Promise<unknown>,
      resolve: resolve as (value: unknown) => void,
      reject: reject as (error: Error) => void,
    });

    // Execute request with debounce
    this.executeRequest(key, requestFn, resolve, reject);

    return promise;
  }

  /**
   * Executes request with debounce
   */
  private executeRequest<T>(
    key: string,
    requestFn: RequestFunction<T>,
    resolve: (value: T) => void,
    reject: (error: Error) => void
  ): void {
    setTimeout(async () => {
      try {
        const result = await requestFn();
        resolve(result);
      } catch (error) {
        reject(error instanceof Error ? error : new Error(String(error)));
      } finally {
        // Clean up
        this.pendingRequests.delete(key);
      }
    }, this.batchTimeout);
  }
}

