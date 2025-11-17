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

  it('should not create WebSocket when merchantId is not provided', () => {
    const { createRiskWebSocketClient } = require('@/lib/websocket');
    
    render(
      <RiskWebSocketProvider>
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    expect(createRiskWebSocketClient).not.toHaveBeenCalled();
  });

  it('should create and connect WebSocket when merchantId is provided', () => {
    const { createRiskWebSocketClient } = require('@/lib/websocket');
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    expect(createRiskWebSocketClient).toHaveBeenCalledWith('merchant-1', expect.any(Object));
    expect(mockWebSocketClient.connect).toHaveBeenCalled();
  });

  it('should disconnect WebSocket on unmount', () => {
    const { unmount } = render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    unmount();
    
    expect(mockWebSocketClient.disconnect).toHaveBeenCalled();
  });

  it('should handle subscribe when connected', () => {
    mockWebSocketClient.isConnected.mockReturnValue(true);
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    const subscribeButton = screen.getByText('Subscribe');
    subscribeButton.click();
    
    expect(mockWebSocketClient.subscribe).toHaveBeenCalledWith('risk', { merchantId: 'merchant-1' });
  });

  it('should handle unsubscribe when connected', () => {
    mockWebSocketClient.isConnected.mockReturnValue(true);
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    const unsubscribeButton = screen.getByText('Unsubscribe');
    unsubscribeButton.click();
    
    expect(mockWebSocketClient.unsubscribe).toHaveBeenCalledWith('risk');
  });

  it('should handle sendMessage when connected', () => {
    mockWebSocketClient.isConnected.mockReturnValue(true);
    
    render(
      <RiskWebSocketProvider merchantId="merchant-1">
        <TestComponent />
      </RiskWebSocketProvider>
    );
    
    const sendButton = screen.getByText('Send');
    sendButton.click();
    
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
  it('should display connected status', () => {
    const mockContext = {
      status: 'connected' as const,
      isConnected: true,
      subscribe: vi.fn(),
      unsubscribe: vi.fn(),
      sendMessage: vi.fn(),
    };
    
    // Mock the context
    vi.spyOn(require('react'), 'useContext').mockReturnValue(mockContext);
    
    render(<WebSocketStatusIndicator />);
    
    expect(screen.getByText('Connected')).toBeInTheDocument();
  });

  it('should display connecting status', () => {
    const mockContext = {
      status: 'connecting' as const,
      isConnected: false,
      subscribe: vi.fn(),
      unsubscribe: vi.fn(),
      sendMessage: vi.fn(),
    };
    
    vi.spyOn(require('react'), 'useContext').mockReturnValue(mockContext);
    
    render(<WebSocketStatusIndicator />);
    
    expect(screen.getByText('Connecting')).toBeInTheDocument();
  });

  it('should display error status', () => {
    const mockContext = {
      status: 'error' as const,
      isConnected: false,
      subscribe: vi.fn(),
      unsubscribe: vi.fn(),
      sendMessage: vi.fn(),
    };
    
    vi.spyOn(require('react'), 'useContext').mockReturnValue(mockContext);
    
    render(<WebSocketStatusIndicator />);
    
    expect(screen.getByText('Error')).toBeInTheDocument();
  });

  it('should display disconnected status', () => {
    const mockContext = {
      status: 'disconnected' as const,
      isConnected: false,
      subscribe: vi.fn(),
      unsubscribe: vi.fn(),
      sendMessage: vi.fn(),
    };
    
    vi.spyOn(require('react'), 'useContext').mockReturnValue(mockContext);
    
    render(<WebSocketStatusIndicator />);
    
    expect(screen.getByText('Disconnected')).toBeInTheDocument();
  });
});

