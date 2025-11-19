'use client';

import { createContext, useContext, useEffect, useState, useRef, ReactNode } from 'react';
import { createRiskWebSocketClient, type WebSocketStatus } from '@/lib/websocket';
import { Badge } from '@/components/ui/badge';
import { AlertCircle, CheckCircle2, XCircle, Loader2 } from 'lucide-react';

interface RiskWebSocketContextType {
  status: WebSocketStatus;
  isConnected: boolean;
  subscribe: (merchantId: string) => void;
  unsubscribe: (merchantId: string) => void;
  sendMessage: (type: string, data: any) => void;
}

const RiskWebSocketContext = createContext<RiskWebSocketContextType | null>(null);

interface RiskWebSocketProviderProps {
  children: ReactNode;
  merchantId?: string;
}

export function RiskWebSocketProvider({ children, merchantId }: RiskWebSocketProviderProps) {
  const [status, setStatus] = useState<WebSocketStatus>('disconnected');
  const wsClientRef = useRef<ReturnType<typeof createRiskWebSocketClient> | null>(null);

  useEffect(() => {
    // Only create WebSocket if merchantId is provided
    if (!merchantId) {
      return;
    }

    // Add a small delay to ensure component is fully mounted
    const connectTimeout = setTimeout(() => {
      try {
        const client = createRiskWebSocketClient(merchantId, {
          onStatusChange: (newStatus) => {
            setStatus(newStatus);
          },
          onRiskUpdate: (data) => {
            // Handle risk update - could dispatch event or update context
            window.dispatchEvent(
              new CustomEvent('riskUpdate', {
                detail: data,
              })
            );
          },
          onRiskPrediction: (data) => {
            window.dispatchEvent(
              new CustomEvent('riskPrediction', {
                detail: data,
              })
            );
          },
          onRiskAlert: (data) => {
            window.dispatchEvent(
              new CustomEvent('riskAlert', {
                detail: data,
              })
            );
          },
        });

        wsClientRef.current = client;
        client.connect();
      } catch (error) {
        console.error('Failed to create WebSocket client:', error);
        setStatus('error');
      }
    }, 100);

    return () => {
      clearTimeout(connectTimeout);
      if (wsClientRef.current) {
        wsClientRef.current.disconnect();
      }
    };
  }, [merchantId]);

  const subscribe = (id: string) => {
    if (wsClientRef.current && wsClientRef.current.isConnected()) {
      wsClientRef.current.subscribe('risk', { merchantId: id });
    }
  };

  const unsubscribe = (id: string) => {
    if (wsClientRef.current && wsClientRef.current.isConnected()) {
      wsClientRef.current.unsubscribe('risk');
    }
  };

  const sendMessage = (type: string, data: any) => {
    if (wsClientRef.current && wsClientRef.current.isConnected()) {
      wsClientRef.current.send({ type, data });
    }
  };

  return (
    <RiskWebSocketContext.Provider
      value={{
        status,
        isConnected: status === 'connected',
        subscribe,
        unsubscribe,
        sendMessage,
      }}
    >
      {children}
    </RiskWebSocketContext.Provider>
  );
}

export function useRiskWebSocket() {
  const context = useContext(RiskWebSocketContext);
  if (!context) {
    throw new Error('useRiskWebSocket must be used within RiskWebSocketProvider');
  }
  return context;
}

/**
 * WebSocket Status Indicator Component
 */
export function WebSocketStatusIndicator() {
  const { status, isConnected } = useRiskWebSocket();

  const getStatusBadge = () => {
    switch (status) {
      case 'connected':
        return (
          <Badge variant="default" className="gap-1">
            <CheckCircle2 className="h-3 w-3" />
            Connected
          </Badge>
        );
      case 'connecting':
        return (
          <Badge variant="secondary" className="gap-1">
            <Loader2 className="h-3 w-3 animate-spin" />
            Connecting
          </Badge>
        );
      case 'error':
        return (
          <Badge variant="destructive" className="gap-1">
            <XCircle className="h-3 w-3" />
            Error
          </Badge>
        );
      default:
        return (
          <Badge variant="outline" className="gap-1">
            <AlertCircle className="h-3 w-3" />
            Disconnected
          </Badge>
        );
    }
  };

  return <div>{getStatusBadge()}</div>;
}

