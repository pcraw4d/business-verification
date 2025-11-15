// Request Deduplication Utility

type RequestFunction<T> = () => Promise<T>;

export class RequestDeduplicator {
  private pendingRequests: Map<string, Promise<unknown>>;

  constructor() {
    this.pendingRequests = new Map();
  }

  /**
   * Deduplicates requests - returns existing promise if request is already pending
   */
  async deduplicate<T>(key: string, requestFn: RequestFunction<T>): Promise<T> {
    // Check if request is already pending
    const existing = this.pendingRequests.get(key);
    if (existing) {
      return existing as Promise<T>;
    }

    // Create new request promise
    const promise = requestFn()
      .then((result) => {
        // Remove from pending when done
        this.pendingRequests.delete(key);
        return result;
      })
      .catch((error) => {
        // Remove from pending on error
        this.pendingRequests.delete(key);
        throw error;
      });

    // Store pending request
    this.pendingRequests.set(key, promise);

    return promise;
  }

  /**
   * Clears all pending requests
   */
  clear(): void {
    this.pendingRequests.clear();
  }
}

