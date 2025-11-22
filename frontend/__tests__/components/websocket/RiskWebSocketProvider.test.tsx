import { render, screen, waitFor } from '@testing-library/react';
import { RiskWebSocketProvider, useRiskWebSocket, WebSocketStatusIndicator } from '@/components/websocket/RiskWebSocketProvider';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ReactNode } from 'react';

// Mock the websocket client
const mockWebSocketClient = {
  connect: vi.fn(),
  disconnect: vi.fn(),
  isConnected: vi.fn(() => false),
  subscribe: vi.fn(),
  unsubscribe: vi.fn(),
  send: vi.fn(),
};

vi.mock('@/lib/websocket', () => ({
  createRiskWebSocketClient: vi.fn(() => mockWebSocketClient),
  WebSocketStatus: {
    disconnected: 'disconnected',
    connecting: 'connecting',
    connected: 'connected',
    error: 'error',
  },
}));

// Test component that uses the hook
function TestComponent() {
  const { status, isConnected, subscribe, unsubscribe, sendMessage } = useRiskWebSocket();
  
  return (
    <div>
      <div data-testid="status">{status}</div>
      <div data-testid="is-connected">{isConnected ? 'true' : 'false'}</div>
      <button onClick={() => subscribe('merchant-1')}>Subscribe</button>
      <button onClick={() => unsubscribe('merchant-1')}>Unsubscribe</button>
      <button onClick={() => sendMessage('test', { data: 'test' })}>Send</button>
    </div>
  );
}

describe('RiskWebSocketProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockWebSocketClient.isConnected.mockReturnValue(false);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should provide context to children', () => {
    render(
      <RiskWebSocketProvider>
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    expect(screen.getByTestId('status')).toBeInTheDocument();
    expect(screen.getByTestId('is-connected')).toBeInTheDocument();
  });

  it('should not create WebSocket when merchantId is not provided', async () => {
    const { createRiskWebSocketClient } = await import('@/lib/websocket');
    
    render(
      <RiskWebSocketProvider>
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    // Wait a bit to ensure useEffect has run
    await waitFor(() => {
      expect(createRiskWebSocketClient).not.toHaveBeenCalled();
    }, { timeout: 200 });
  });

  it('should create and connect WebSocket when merchantId is provided', async () => {
    const { createRiskWebSocketClient } = await import('@/lib/websocket');
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    // Wait for the setTimeout in useEffect to complete (100ms delay)
    await waitFor(() => {
      expect(createRiskWebSocketClient).toHaveBeenCalledWith('merchant-1', expect.any(Object));
      expect(mockWebSocketClient.connect).toHaveBeenCalled();
    }, { timeout: 200 });
  });

  it('should disconnect WebSocket on unmount', async () => {
    const { unmount } = render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    // Wait for connection to be established
    await waitFor(() => {
      expect(mockWebSocketClient.connect).toHaveBeenCalled();
    }, { timeout: 200 });
    
    unmount();
    
    expect(mockWebSocketClient.disconnect).toHaveBeenCalled();
  });

  it('should handle subscribe when connected', async () => {
    mockWebSocketClient.isConnected.mockReturnValue(true);
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    // Wait for connection
    await waitFor(() => {
      expect(mockWebSocketClient.connect).toHaveBeenCalled();
    }, { timeout: 200 });
    
    const subscribeButton = screen.getByText('Subscribe');
    subscribeButton.click();
    
    // Component's subscribe function calls: client.subscribe('risk', { merchantId: id })
    expect(mockWebSocketClient.subscribe).toHaveBeenCalledWith('risk', { merchantId: 'merchant-1' });
  });

  it('should handle unsubscribe when connected', async () => {
    mockWebSocketClient.isConnected.mockReturnValue(true);
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    // Wait for connection
    await waitFor(() => {
      expect(mockWebSocketClient.connect).toHaveBeenCalled();
    }, { timeout: 200 });
    
    const unsubscribeButton = screen.getByText('Unsubscribe');
    unsubscribeButton.click();
    
    // Component's unsubscribe function calls: client.unsubscribe('risk')
    expect(mockWebSocketClient.unsubscribe).toHaveBeenCalledWith('risk');
  });

  it('should handle sendMessage when connected', async () => {
    mockWebSocketClient.isConnected.mockReturnValue(true);
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    // Wait for connection
    await waitFor(() => {
      expect(mockWebSocketClient.connect).toHaveBeenCalled();
    }, { timeout: 200 });
    
    const sendButton = screen.getByText('Send');
    sendButton.click();
    
    // Component's sendMessage function wraps: client.send({ type, data })
    expect(mockWebSocketClient.send).toHaveBeenCalledWith({ type: 'test', data: { data: 'test' } });
  });

  it('should throw error when useRiskWebSocket is used outside provider', () => {
    // Suppress console.error for this test
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
    
    expect(() => {
      render(<TestComponent />);
    }).toThrow('useRiskWebSocket must be used within RiskWebSocketProvider');
    
    consoleError.mockRestore();
  });
});

describe('WebSocketStatusIndicator', () => {
  it('should display connected status', async () => {
    // Mock the client to return connected status
    mockWebSocketClient.isConnected.mockReturnValue(true);
    
    // Mock onStatusChange callback to set status to connected
    const { createRiskWebSocketClient } = await import('@/lib/websocket');
    const mockCreate = vi.mocked(createRiskWebSocketClient);
    
    let statusCallback: ((status: string) => void) | undefined;
    mockCreate.mockImplementation((merchantId, callbacks) => {
      statusCallback = callbacks?.onStatusChange;
      // Immediately call onStatusChange with 'connected'
      setTimeout(() => {
        statusCallback?.('connected');
      }, 0);
      return mockWebSocketClient;
    });
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <WebSocketStatusIndicator />
      </RiskWebSocketProvider>
    );
    
    await waitFor(() => {
      expect(screen.getByText('Connected')).toBeInTheDocument();
    }, { timeout: 500 });
  });

  it('should display connecting status', async () => {
    // Mock the client to return connecting status
    mockWebSocketClient.isConnected.mockReturnValue(false);
    
    const { createRiskWebSocketClient } = await import('@/lib/websocket');
    const mockCreate = vi.mocked(createRiskWebSocketClient);
    
    let statusCallback: ((status: string) => void) | undefined;
    mockCreate.mockImplementation((merchantId, callbacks) => {
      statusCallback = callbacks?.onStatusChange;
      // Immediately call onStatusChange with 'connecting'
      setTimeout(() => {
        statusCallback?.('connecting');
      }, 0);
      return mockWebSocketClient;
    });
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <WebSocketStatusIndicator />
      </RiskWebSocketProvider>
    );
    
    await waitFor(() => {
      expect(screen.getByText('Connecting')).toBeInTheDocument();
    }, { timeout: 500 });
  });

  it('should display error status', async () => {
    const { createRiskWebSocketClient } = await import('@/lib/websocket');
    const mockCreate = vi.mocked(createRiskWebSocketClient);
    
    let statusCallback: ((status: string) => void) | undefined;
    mockCreate.mockImplementation((merchantId, callbacks) => {
      statusCallback = callbacks?.onStatusChange;
      // Immediately call onStatusChange with 'error'
      setTimeout(() => {
        statusCallback?.('error');
      }, 0);
      return mockWebSocketClient;
    });
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <WebSocketStatusIndicator />
      </RiskWebSocketProvider>
    );
    
    await waitFor(() => {
      expect(screen.getByText('Error')).toBeInTheDocument();
    }, { timeout: 500 });
  });

  it('should display disconnected status', async () => {
    // Default status is 'disconnected', so we don't need to set it
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <WebSocketStatusIndicator />
      </RiskWebSocketProvider>
    );
    
    // Wait a bit for the component to render
    await waitFor(() => {
      expect(screen.getByText('Disconnected')).toBeInTheDocument();
    }, { timeout: 500 });
  });
});

