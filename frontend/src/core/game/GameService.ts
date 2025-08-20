import { ProtobufWebSocketService } from '../messaging/ProtobufWebSocketService';
import { ApiService } from '../api/api_service';
import { LobbyMessage, GameMessage } from '../../proto/server_messages';
import { GalaxyGenerateSettings } from '../../proto/client_commands';

export interface GameSession {
  session_id: string;
  invite_code: string;
  state: 'waiting' | 'active' | 'paused' | 'ended';
}

export interface Player {
  user_id: string;
  display_name: string;
  color?: string;
  ready: boolean;
  is_host: boolean;
}

export interface LobbyState {
  session: GameSession;
  players: Player[];
  settings?: {
    numStars?: number;
    shape?: string;
    maxHyperlanes?: number;
    hyperlaneConnectivity?: number;
  };
}

export type GameEventType = 
  | 'lobby_updated'
  | 'player_joined'
  | 'player_left'
  | 'player_ready_changed'
  | 'player_color_changed'
  | 'game_started'
  | 'game_loading'
  | 'error';

export interface GameEvent {
  type: GameEventType;
  data: any;
}

export class GameService {
  private websocket: ProtobufWebSocketService;
  private apiService: ApiService;
  private listeners: Map<GameEventType, Set<(data: any) => void>> = new Map();
  private currentSession: GameSession | null = null;
  private currentLobbyState: LobbyState | null = null;

  constructor(apiService: ApiService) {
    this.apiService = apiService;
    this.websocket = new ProtobufWebSocketService();
    this.setupWebSocketHandlers();
  }

  private setupWebSocketHandlers() {
    // Handle lobby messages
    this.websocket.onLobbyMessage((message: LobbyMessage) => {
      if (message.lobbyState) {
        // Update local lobby state
        this.updateLocalLobbyState(message.lobbyState);
        this.emit('lobby_updated', this.currentLobbyState);
      }
      
      if (message.playerJoined) {
        this.emit('player_joined', {
          playerId: message.playerJoined.player?.playerId,
          displayName: message.playerJoined.player?.displayName
        });
      }
      
      if (message.playerLeft) {
        this.emit('player_left', {
          playerId: message.playerLeft.playerId,
          displayName: message.playerLeft.displayName
        });
      }
      
      if (message.playerUpdated) {
        this.emit('player_ready_changed', {
          playerId: message.playerUpdated.player?.playerId,
          ready: message.playerUpdated.player?.isReady
        });
      }
      
      if (message.gameStarting) {
        this.emit('game_started', {
          settings: message.gameStarting.finalSettings
        });
      }
      
      if (message.gameLoading) {
        this.emit('game_loading', {
          progress: message.gameLoading.progress
        });
      }
    });

    // Handle game messages
    this.websocket.onGameMessage((message: GameMessage) => {
      // Handle game state updates, chat messages, etc.
      console.log('Received game message:', message);
    });

    // Handle errors
    this.websocket.onError((error) => {
      this.emit('error', error);
    });

    // Handle connection changes
    this.websocket.setConnectionChangeHandler((connected: boolean) => {
      if (connected) {
        console.log('WebSocket connected');
      } else {
        console.log('WebSocket disconnected');
      }
    });

    // Handle token expiration
    this.websocket.onTokenExpired = () => {
      console.log('WebSocket token expired');
      this.emit('error', { 
        code: 'TOKEN_EXPIRED', 
        message: 'Authentication token expired' 
      });
    };
  }

  private updateLocalLobbyState(lobbyState: any) {
    // Convert the protobuf lobby state to our local format
    this.currentLobbyState = {
      session: {
        session_id: lobbyState.sessionId,
        invite_code: lobbyState.inviteCode,
        state: this.mapLobbyStatus(lobbyState.status)
      },
      players: lobbyState.players?.map((p: any) => ({
        user_id: p.playerId,
        display_name: p.displayName,
        color: p.color,
        ready: p.isReady,
        is_host: p.isHost
      })) || [],
      settings: lobbyState.settings ? {
        numStars: lobbyState.settings.numStars,
        shape: lobbyState.settings.shape,
        maxHyperlanes: lobbyState.settings.maxHyperlanes,
        hyperlaneConnectivity: lobbyState.settings.hyperlaneConnectivity
      } : undefined
    };
  }

  private mapLobbyStatus(status: number): 'waiting' | 'active' | 'paused' | 'ended' {
    // Map protobuf enum to our local enum
    switch (status) {
      case 0: return 'waiting'; // WAITING
      case 1: return 'active';  // STARTING
      case 2: return 'active';  // IN_GAME
      default: return 'waiting';
    }
  }

  // Event listener management
  on(event: GameEventType, callback: (data: any) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);

    // Return unsubscribe function
    return () => {
      const eventListeners = this.listeners.get(event);
      if (eventListeners) {
        eventListeners.delete(callback);
      }
    };
  }

  private emit(event: GameEventType, data: any) {
    const eventListeners = this.listeners.get(event);
    if (eventListeners) {
      eventListeners.forEach(callback => callback(data));
    }
  }

  // Game session management
  async createGame(): Promise<GameSession> {
    const response = await this.apiService.createGame();
    this.currentSession = response;
    return response;
  }

  async joinGame(inviteCode: string): Promise<GameSession> {
    const response = await this.apiService.joinGame(inviteCode);
    this.currentSession = response;
    return response;
  }

  async leaveGame(): Promise<void> {
    await this.apiService.leaveGame();
    this.disconnect();
    this.currentSession = null;
    this.currentLobbyState = null;
  }

  async getCurrentGame(): Promise<GameSession | null> {
    try {
      const response = await this.apiService.getCurrentGame();
      this.currentSession = response;
      return response;
    } catch (error) {
      this.currentSession = null;
      return null;
    }
  }

  // WebSocket connection management
  connect(wsUrl: string, token: string) {
    this.websocket.setAuthToken(token);
    // Add protocol parameter for protobuf communication
    const protobufUrl = wsUrl;
    this.websocket.connect(protobufUrl);
  }

  disconnect() {
    this.websocket.disconnect();
  }

  isConnected(): boolean {
    return this.websocket.isConnected();
  }

  // Lobby actions using protobuf messages
  // Note: joinLobby is not needed - backend automatically sends lobby state on WebSocket connection

  setPlayerReady(userId: string, ready: boolean) {
    console.log('GameService: Setting player ready:', userId, ready);
    const lobbyCommand = { setReady: { ready } };
    this.websocket.sendLobbyCommand(userId, lobbyCommand);
  }

  setPlayerColor(userId: string, color: string) {
    console.log('GameService: Setting player color:', userId, color);
    const lobbyCommand = { setColor: { color } };
    this.websocket.sendLobbyCommand(userId, lobbyCommand);
  }

  updateGameSettings(settings: GalaxyGenerateSettings, playerId: string = 'current-user') {
    console.log('GameService: Updating game settings:', settings);
    // Store settings locally
    if (this.currentLobbyState) {
      this.currentLobbyState.settings = {
        numStars: settings.numStars,
        shape: settings.shape,
        maxHyperlanes: settings.maxHyperlanes,
        hyperlaneConnectivity: settings.hyperlaneConnectivity
      };
    }
    
    const lobbyCommand = { updateSettings: { settings } };
    this.websocket.sendLobbyCommand(playerId, lobbyCommand);
  }

  startGame(playerId: string = 'current-user') {
    console.log('GameService: Starting game');
    const lobbyCommand = { startGame: {} };
    this.websocket.sendLobbyCommand(playerId, lobbyCommand);
  }

  // Getters
  getCurrentSession(): GameSession | null {
    return this.currentSession;
  }

  getCurrentLobbyState(): LobbyState | null {
    return this.currentLobbyState;
  }
}
