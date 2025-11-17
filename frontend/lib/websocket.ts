/**
 * WebSocket client for real-time updates
 */

export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp?: string;
}

export interface WebSocketOptions {
  url: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  onMessage?: (message: WebSocketMessage) => void;
  onStatusChange?: (status: WebSocketStatus) => void;
  onError?: (error: Error) => void;
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private status: WebSocketStatus = 'disconnected';
  private reconnectAttempts = 0;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private options: Required<WebSocketOptions>;
  private messageQueue: WebSocketMessage[] = [];

  constructor(options: WebSocketOptions) {
    this.options = {
      reconnectInterval: 3000,
      maxReconnectAttempts: 10,
      onMessage: () => {},
      onStatusChange: () => {},
      onError: () => {},
      ...options,
    };
  }

  /**
   * Connect to WebSocket server
   */
  connect(): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return; // Already connected
    }

    if (this.ws && this.ws.readyState === WebSocket.CONNECTING) {
      return; // Already connecting
    }

    try {
      this.setStatus('connecting');
      this.ws = new WebSocket(this.options.url);

      this.ws.onopen = () => {
        this.setStatus('connected');
        this.reconnectAttempts = 0;
        this.flushMessageQueue();
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.options.onMessage(message);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        this.setStatus('error');
        this.options.onError(new Error('WebSocket error'));
      };

      this.ws.onclose = (event) => {
        this.setStatus('disconnected');
        this.ws = null;

        // Attempt to reconnect if not a normal closure
        if (event.code !== 1000 && this.reconnectAttempts < this.options.maxReconnectAttempts) {
          this.scheduleReconnect();
        }
      };
    } catch (error) {
      this.setStatus('error');
      this.options.onError(error instanceof Error ? error : new Error('Failed to connect'));
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }

    this.setStatus('disconnected');
    this.reconnectAttempts = 0;
  }

  /**
   * Send message to WebSocket server
   */
  send(message: WebSocketMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      // Queue message if not connected
      this.messageQueue.push(message);
      // Try to connect if not already connecting
      if (this.status === 'disconnected') {
        this.connect();
      }
    }
  }

  /**
   * Subscribe to a channel/topic
   */
  subscribe(channel: string, params?: Record<string, any>): void {
    this.send({
      type: 'subscribe',
      data: {
        channel,
        ...params,
      },
    });
  }

  /**
   * Unsubscribe from a channel/topic
   */
  unsubscribe(channel: string): void {
    this.send({
      type: 'unsubscribe',
      data: {
        channel,
      },
    });
  }

  /**
   * Get current connection status
   */
  getStatus(): WebSocketStatus {
    return this.status;
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.status === 'connected' && this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Set status and notify listeners
   */
  private setStatus(status: WebSocketStatus): void {
    this.status = status;
    this.options.onStatusChange(status);
  }

  /**
   * Schedule reconnection attempt
   */
  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      return;
    }

    this.reconnectAttempts++;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, this.options.reconnectInterval);
  }

  /**
   * Flush queued messages
   */
  private flushMessageQueue(): void {
    while (this.messageQueue.length > 0 && this.isConnected()) {
      const message = this.messageQueue.shift();
      if (message) {
        this.send(message);
      }
    }
  }
}

/**
 * Create a WebSocket client for risk updates
 */
export function createRiskWebSocketClient(
  merchantId?: string,
  callbacks?: {
    onRiskUpdate?: (data: any) => void;
    onRiskPrediction?: (data: any) => void;
    onRiskAlert?: (data: any) => void;
    onStatusChange?: (status: WebSocketStatus) => void;
  }
): WebSocketClient {
  const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
  const wsUrl = apiBaseUrl
    .replace('http://', 'ws://')
    .replace('https://', 'wss://')
    .replace(/\/$/, '') + '/api/v1/risk/ws';

  const client = new WebSocketClient({
    url: wsUrl,
    onMessage: (message) => {
      switch (message.type) {
        case 'riskUpdate':
          callbacks?.onRiskUpdate?.(message.data);
          break;
        case 'riskPrediction':
          callbacks?.onRiskPrediction?.(message.data);
          break;
        case 'riskAlert':
          callbacks?.onRiskAlert?.(message.data);
          break;
      }
    },
    onStatusChange: callbacks?.onStatusChange,
  });

  // Subscribe to merchant-specific updates if merchantId provided
  if (merchantId) {
    client.connect();
    // Subscribe after connection is established
    setTimeout(() => {
      if (client.isConnected()) {
        client.subscribe('risk', { merchantId });
      }
    }, 1000);
  }

  return client;
}

