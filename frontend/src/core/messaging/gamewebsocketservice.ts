import { GameMessage } from "../../proto/messages";

// Game WebSocket Service Integration
export class GameWebSocketService {
  private ws: WebSocket | null = null;
  private messageHandlers = new Map<string, Array<(message: any) => void>>();
  private responseWaiters = new Map<string, (message: any) => void>();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private authToken: string | null = null;
  private onConnectionChange?: (connected: boolean) => void;

  // Callback for token expiration
  public onTokenExpired?: () => void;

  setAuthToken(token: string) {
    this.authToken = token;
  }

  setConnectionChangeHandler(handler: (connected: boolean) => void) {
    this.onConnectionChange = handler;
  }

  connect(baseUrl: string) {
    if (!this.authToken) {
      console.error('Cannot connect without auth token');
      return;
    }

    const wsUrl = `${baseUrl}/ws?token=${this.authToken}`;
    this.ws = new WebSocket(wsUrl);
    
    this.ws.onopen = () => {
      console.log('Connected to game server');
      this.reconnectAttempts = 0;
      this.onConnectionChange?.(true);
    };

    this.ws.onmessage = (event) => {
      try {
        // For demo purposes using JSON, in production use protobuf
        const message = JSON.parse(event.data);
        // Detect token expiration error from backend
        if (message.code === "TOKEN_EXPIRED") {
          if (this.onTokenExpired) this.onTokenExpired();
          return;
        }
        this.handleMessage(message);
      } catch (error) {
        console.error('Failed to parse message:', error);
      }
    };

    this.ws.onclose = (event) => {
      console.log('Disconnected from game server', event.code);
      this.onConnectionChange?.(false);
      
      // Only attempt reconnect if we have a token and it wasn't a clean close
      if (this.authToken && event.code !== 1000) {
        this.attemptReconnect(baseUrl);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  private handleMessage(message: GameMessage) {
    // Handle response to specific message
    if (message.messageId) {
      const responseWaiter = this.responseWaiters.get(message.messageId);
      if (responseWaiter) {
        responseWaiter(message);
        this.responseWaiters.delete(message.messageId);
        return;
      }
    }

  

    // Handle game-specific updates
    if (message?.gameStateUpdate) {
      this.handleGameStateUpdate(message);
    } else if (message?.chatMessage) {
      this.handleChatMessage(message);
    }
  }

  private handleGameStateUpdate(message: any) {
    const update = message?.gameStateUpdate;
    if (!update) return;

    // Dispatch to appropriate handlers
    if (update?.fleetUpdate) {
      this.notifyHandlers('fleet_update', message);
    } else if (update?.resourceUpdate) {
      this.notifyHandlers('resource_update', message);
    } else if (update?.systemUpdate) {
      this.notifyHandlers('system_update', message);
    } else if (update?.constructionUpdate) {
      this.notifyHandlers('construction_update', message);
    } else if (update?.diplomacyUpdate) {
      this.notifyHandlers('diplomacy_update', message);
    }
  }

  private handleChatMessage(message: any) {
    this.notifyHandlers('chat_message', message);
  }

  private notifyHandlers(type: string, message: any) {
    const handlers = this.messageHandlers.get(type) || [];
    handlers.forEach(handler => handler(message));
  }

  sendMessage(message: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      // In production, encode as protobuf binary
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket not connected, cannot send message');
    }
  }

  subscribeToMessage(messageId: string, handler: (message: any) => void): () => void {
    this.responseWaiters.set(messageId, handler);
    return () => this.responseWaiters.delete(messageId);
  }

  subscribeToUpdates(updateType: string, handler: (message: any) => void): () => void {
    if (!this.messageHandlers.has(updateType)) {
      this.messageHandlers.set(updateType, []);
    }
    this.messageHandlers.get(updateType)!.push(handler);

    return () => {
      const handlers = this.messageHandlers.get(updateType);
      if (handlers) {
        const index = handlers.indexOf(handler);
        if (index > -1) handlers.splice(index, 1);
      }
    };
  }

  private attemptReconnect(baseUrl: string) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
      console.log(`Reconnecting in ${delay}ms... (attempt ${this.reconnectAttempts})`);
      setTimeout(() => this.connect(baseUrl), delay);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close(1000); // Clean close
      this.ws = null;
    }
    this.authToken = null;
    this.messageHandlers.clear();
    this.responseWaiters.clear();
    this.reconnectAttempts = 0;
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export function generateMessageId(): string {
  return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}