import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { ProtobufWebSocketService } from '../messaging/ProtobufWebSocketService';

interface WebSocketContextType {
  wsService: ProtobufWebSocketService | null;
  isConnected: boolean;
  connectionError: string | null;
  connect: () => Promise<void>;
  disconnect: () => void;
  subscribe: (callback: (message: any) => void) => () => void;
  send: (message: any) => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

interface WebSocketProviderProps {
  children: React.ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
  const [wsService, setWsService] = useState<ProtobufWebSocketService | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const [subscribers, setSubscribers] = useState<Set<(message: any) => void>>(new Set());

  const connect = useCallback(async () => {
    if (wsService && isConnected) return;

    try {
      setConnectionError(null);
      const service = new ProtobufWebSocketService();
      
      // Get auth token from localStorage
      const authToken = localStorage.getItem('authToken');
      if (!authToken) {
        throw new Error('No auth token available');
      }
      service.setAuthToken(authToken);
      
      // Set up connection change handler
      service.setConnectionChangeHandler((connected: boolean) => {
        setIsConnected(connected);
        if (!connected) {
          setConnectionError('Connection lost');
        }
      });

      // Set up error handler
      service.onError((error) => {
        setConnectionError(error.message);
        setIsConnected(false);
      });

      // Set up server message handler to broadcast to all subscribers
      service.onServerMessage((message) => {
        // Broadcast to all subscribers
        subscribers.forEach(callback => {
          try {
            callback(message);
          } catch (error) {
            console.error('Error in message subscriber:', error);
          }
        });
      });

      // Get base URL from config
      const baseUrl = 'wss://localhost:8443/ws';
      service.connect(baseUrl);
      setWsService(service);
      setIsConnected(true);
      setConnectionError(null);
    } catch (error: any) {
      setConnectionError(error.message || 'Failed to connect');
      console.error('WebSocket connection failed:', error);
    }
  }, [wsService, isConnected, subscribers]);

  const disconnect = useCallback(() => {
    if (wsService) {
      wsService.disconnect();
      setWsService(null);
      setIsConnected(false);
      setConnectionError(null);
    }
  }, [wsService]);

  const subscribe = useCallback((callback: (message: any) => void) => {
    setSubscribers(prev => new Set([...prev, callback]));
    
    // Return unsubscribe function
    return () => {
      setSubscribers(prev => {
        const newSet = new Set(prev);
        newSet.delete(callback);
        return newSet;
      });
    };
  }, []);

  const send = useCallback((command: any) => {
    if (wsService && isConnected) {
      wsService.sendCommand(command);
    } else {
      console.warn('Cannot send message: WebSocket not connected');
    }
  }, [wsService, isConnected]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  const value: WebSocketContextType = {
    wsService,
    isConnected,
    connectionError,
    connect,
    disconnect,
    subscribe,
    send
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
};
