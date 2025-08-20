import { ClientCommand, LobbyCommand, GameCommand, ChatCommand } from "../../proto/client_commands";
import { ServerMessage, LobbyMessage, GameMessage as ServerGameMessage, ChatMessage, SystemMessage } from "../../proto/server_messages";

// Local error message structure
interface ErrorMessage {
  code: string;
  message: string;
  details?: any;
}

export type ServerMessageHandler = (message: ServerMessage) => void;
export type LobbyMessageHandler = (message: LobbyMessage) => void;
export type GameMessageHandler = (message: ServerGameMessage) => void;
export type ChatMessageHandler = (message: ChatMessage) => void;
export type SystemMessageHandler = (message: SystemMessage) => void;
export type ErrorHandler = (error: ErrorMessage) => void;

// Simplified Protobuf WebSocket Service - sends protobuf directly over wire
export class ProtobufWebSocketService {
  private ws: WebSocket | null = null;
  private serverMessageHandlers = new Set<ServerMessageHandler>();
  private lobbyMessageHandlers = new Set<LobbyMessageHandler>();
  private gameMessageHandlers = new Set<GameMessageHandler>();
  private chatMessageHandlers = new Set<ChatMessageHandler>();
  private systemMessageHandlers = new Set<SystemMessageHandler>();
  private errorHandlers = new Set<ErrorHandler>();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private authToken: string | null = null;
  private onConnectionChange?: (connected: boolean) => void;
  private pingInterval?: NodeJS.Timeout;

  // Callback for token expiration
  public onTokenExpired?: () => void;

  setAuthToken(token: string) {
    this.authToken = token;
  }

  setConnectionChangeHandler(handler: (connected: boolean) => void) {
    this.onConnectionChange = handler;
  }

  connect(baseUrl: string) {
    console.log('🚀 PROTOBUF SERVICE: STARTING CONNECTION TO', baseUrl);
    if (!this.authToken) {
      console.error('Cannot connect without auth token');
      return;
    }

    const wsUrl = baseUrl + `?token=${this.authToken}`;
    console.log('Connecting to:', wsUrl);
    
    this.ws = new WebSocket(wsUrl);
    
    // Set binary type for protobuf
    this.ws.binaryType = 'arraybuffer';

    this.ws.onopen = () => {
      console.log('🟢 Connected to game server via protobuf WebSocket');
      this.reconnectAttempts = 0;
      this.onConnectionChange?.(true);
      this.startPing();
    };

    this.ws.onmessage = (event) => {
      console.log('📨 RAW MESSAGE RECEIVED:', event.data);
      try {
        if (event.data instanceof ArrayBuffer) {
          this.handleBinaryMessage(new Uint8Array(event.data));
        } else if (typeof event.data === 'string') {
          // Fallback for JSON messages (errors, etc.)
          this.handleJsonMessage(JSON.parse(event.data));
        }
      } catch (error) {
        console.error('Failed to parse message:', error);
      }
    };

    this.ws.onclose = (event) => {
      console.log('Disconnected from game server', event.code);
      this.onConnectionChange?.(false);
      this.stopPing();
      
      // Only attempt reconnect if we have a token and it wasn't a clean close
      if (this.authToken && event.code !== 1000) {
        this.attemptReconnect(baseUrl);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  private handleBinaryMessage(data: Uint8Array) {
    try {
      // Directly decode the ServerMessage protobuf
      const serverMessage = ServerMessage.decode(data);
      console.log('🎯 ProtobufWebSocketService: Received server message:', serverMessage);
      console.log('📊 ProtobufWebSocketService: Handler counts:', {
        serverMessageHandlers: this.serverMessageHandlers.size,
        lobbyMessageHandlers: this.lobbyMessageHandlers.size,
        gameMessageHandlers: this.gameMessageHandlers.size,
        chatMessageHandlers: this.chatMessageHandlers.size,
        systemMessageHandlers: this.systemMessageHandlers.size,
        errorHandlers: this.errorHandlers.size
      });
      
      // Notify all server message handlers
      console.log('📤 ProtobufWebSocketService: Calling', this.serverMessageHandlers.size, 'server message handlers');
      this.serverMessageHandlers.forEach(handler => {
        console.log('📞 ProtobufWebSocketService: Calling server message handler');
        handler(serverMessage);
      });
      
      // Handle specific message types
      if (serverMessage.lobbyMessage) {
        console.log('🏛️ ProtobufWebSocketService: Found lobby message, calling', this.lobbyMessageHandlers.size, 'lobby handlers');
        this.lobbyMessageHandlers.forEach(handler => handler(serverMessage.lobbyMessage!));
      }
      
      if (serverMessage.gameMessage) {
        console.log('🎮 ProtobufWebSocketService: Found game message, calling', this.gameMessageHandlers.size, 'game handlers');
        this.gameMessageHandlers.forEach(handler => handler(serverMessage.gameMessage!));
      }
      
      if (serverMessage.chatMessage) {
        console.log('💬 ProtobufWebSocketService: Found chat message, calling', this.chatMessageHandlers.size, 'chat handlers');
        this.chatMessageHandlers.forEach(handler => handler(serverMessage.chatMessage!));
      }
      
      if (serverMessage.systemMessage) {
        console.log('⚙️ ProtobufWebSocketService: Found system message, calling', this.systemMessageHandlers.size, 'system handlers');
        this.systemMessageHandlers.forEach(handler => handler(serverMessage.systemMessage!));
      }
      
      if (serverMessage.errorMessage) {
        console.log('❌ ProtobufWebSocketService: Found error message, calling', this.errorHandlers.size, 'error handlers');
        const errorMsg: ErrorMessage = {
          code: serverMessage.errorMessage.errorCode,
          message: serverMessage.errorMessage.errorMessage,
          details: serverMessage.errorMessage.details
        };
        this.errorHandlers.forEach(handler => handler(errorMsg));
      }
    } catch (error) {
      console.error('Failed to decode server message:', error);
    }
  }

  private handleJsonMessage(message: any) {
    // Handle JSON error messages
    if (message.code === "TOKEN_EXPIRED") {
      if (this.onTokenExpired) this.onTokenExpired();
      return;
    }
    
    if (message.error) {
      const errorMsg: ErrorMessage = {
        code: message.code || 'UNKNOWN_ERROR',
        message: message.error || message.message || 'Unknown error',
        details: message.details
      };
      this.errorHandlers.forEach(handler => handler(errorMsg));
    }
  }

  // Send protobuf commands to server
  sendCommand(command: ClientCommand, messageId?: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const data = ClientCommand.encode(command).finish();
      console.log('Sending client command:', command);
      this.ws.send(data);
    } else {
      console.warn('WebSocket not connected, cannot send command');
    }
  }

  // Convenience methods for specific commands
  sendLobbyCommand(playerId: string, lobbyCommand: LobbyCommand, messageId?: string) {
    const command: ClientCommand = {
      playerId,
      timestamp: Date.now(),
      lobbyCommand
    };
    this.sendCommand(command, messageId);
  }

  sendGameCommand(playerId: string, gameCommand: GameCommand, messageId?: string) {
    const command: ClientCommand = {
      playerId,
      timestamp: Date.now(),
      gameCommand
    };
    this.sendCommand(command, messageId);
  }

  sendChatCommand(playerId: string, chatCommand: ChatCommand, messageId?: string) {
    const command: ClientCommand = {
      playerId,
      timestamp: Date.now(),
      chatCommand
    };
    this.sendCommand(command, messageId);
  }

  // Subscription methods
  onServerMessage(handler: ServerMessageHandler): () => void {
    console.log('🎧 ProtobufWebSocketService: Adding server message handler, total:', this.serverMessageHandlers.size + 1);
    this.serverMessageHandlers.add(handler);
    return () => {
      console.log('🗑️ ProtobufWebSocketService: Removing server message handler, total:', this.serverMessageHandlers.size - 1);
      this.serverMessageHandlers.delete(handler);
    };
  }

  onLobbyMessage(handler: LobbyMessageHandler): () => void {
    this.lobbyMessageHandlers.add(handler);
    return () => this.lobbyMessageHandlers.delete(handler);
  }

  onGameMessage(handler: GameMessageHandler): () => void {
    this.gameMessageHandlers.add(handler);
    return () => this.gameMessageHandlers.delete(handler);
  }

  onChatMessage(handler: ChatMessageHandler): () => void {
    this.chatMessageHandlers.add(handler);
    return () => this.chatMessageHandlers.delete(handler);
  }

  onSystemMessage(handler: SystemMessageHandler): () => void {
    this.systemMessageHandlers.add(handler);
    return () => this.systemMessageHandlers.delete(handler);
  }

  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler);
    return () => this.errorHandlers.delete(handler);
  }

  // Connection management
  private startPing() {
    // Disable ping for now - WebSocket will handle keep-alive
    // this.pingInterval = setInterval(() => {
    //   if (this.isConnected()) {
    //     console.log('Ping disabled for debugging');
    //   }
    // }, 30000);
  }

  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = undefined;
    }
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
    this.stopPing();
    if (this.ws) {
      this.ws.close(1000); // Clean close
      this.ws = null;
    }
    this.authToken = null;
    this.serverMessageHandlers.clear();
    this.lobbyMessageHandlers.clear();
    this.gameMessageHandlers.clear();
    this.chatMessageHandlers.clear();
    this.systemMessageHandlers.clear();
    this.errorHandlers.clear();
    this.reconnectAttempts = 0;
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export function generateMessageId(): string {
  return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}
