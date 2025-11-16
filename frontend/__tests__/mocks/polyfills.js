// Polyfill fetch and Response APIs for MSW in Node.js/Jest environment
// This must be loaded BEFORE any MSW imports
// Based on MSW documentation: https://mswjs.io/docs/getting-started/integrate/node

// Set up TextEncoder/TextDecoder first (needed by undici)
const { TextEncoder, TextDecoder } = require('util');
if (typeof globalThis.TextEncoder === 'undefined') {
  globalThis.TextEncoder = TextEncoder;
}
if (typeof globalThis.TextDecoder === 'undefined') {
  globalThis.TextDecoder = TextDecoder;
}

// Set up ReadableStream and WritableStream (needed by undici and MSW)
const { ReadableStream, WritableStream } = require('stream/web');
if (typeof globalThis.ReadableStream === 'undefined') {
  globalThis.ReadableStream = ReadableStream;
}
if (typeof globalThis.WritableStream === 'undefined') {
  globalThis.WritableStream = WritableStream;
}

// Set up TransformStream (needed by MSW) - may not be in stream/web
try {
  const { TransformStream } = require('stream/web');
  if (typeof globalThis.TransformStream === 'undefined') {
    globalThis.TransformStream = TransformStream;
  }
} catch (e) {
  // TransformStream not available, create minimal polyfill
  if (typeof globalThis.TransformStream === 'undefined') {
    globalThis.TransformStream = class TransformStream {
      constructor() {
        this.readable = new ReadableStream();
        this.writable = new ReadableStream();
      }
    };
  }
}

// Set up MessagePort (needed by undici)
if (typeof globalThis.MessagePort === 'undefined') {
  // Create a minimal MessagePort polyfill
  globalThis.MessagePort = class MessagePort {
    postMessage() {}
    start() {}
    close() {}
    addEventListener() {}
    removeEventListener() {}
    dispatchEvent() { return true; }
  };
  globalThis.MessageChannel = class MessageChannel {
    constructor() {
      this.port1 = new globalThis.MessagePort();
      this.port2 = new globalThis.MessagePort();
    }
  };
}

// Set up BroadcastChannel (needed by MSW)
if (typeof globalThis.BroadcastChannel === 'undefined') {
  globalThis.BroadcastChannel = class BroadcastChannel {
    constructor(name) {
      this.name = name;
      this._listeners = [];
    }
    postMessage(message) {
      // In test environment, just store the message
      this._messages = this._messages || [];
      this._messages.push(message);
    }
    addEventListener(type, listener) {
      if (type === 'message') {
        this._listeners.push(listener);
      }
    }
    removeEventListener(type, listener) {
      if (type === 'message') {
        this._listeners = this._listeners.filter(l => l !== listener);
      }
    }
    close() {
      this._listeners = [];
      this._messages = [];
    }
  };
}

// CRITICAL: MSW cannot intercept requests made via undici
// See: https://mswjs.io/docs/limitations/
// 
// With jest-fixed-jsdom, Node.js globals are restored, so we should use
// Node's native fetch directly without polyfills that might conflict
// 
// jest-fixed-jsdom restores Node.js globals (fetch, Request, Response, etc.)
// so we don't need to polyfill them here - doing so might cause conflicts
// 
// Only polyfill if Node's native fetch is not available (shouldn't happen with Node 18+)
if (typeof globalThis.fetch === 'undefined') {
  // This shouldn't happen with Node 18+ and jest-fixed-jsdom
  // But if it does, use whatwg-fetch as fallback
  require('whatwg-fetch');
}

