import { ProtobufWebSocketService, generateMessageId } from '../messaging/ProtobufWebSocketService';
import { 
  ClientCommand, 
  LobbyCommand, 
  SetReadyCommand, 
  SetColorCommand, 
  UpdateSettingsCommand,
  GalaxyGenerateSettings 
} from '../../proto/client_commands';
import { 
  ServerMessage, 
  LobbyMessage, 
  LobbyStateMessage, 
  LobbyPlayer,
  PlayerJoinedMessage,
  PlayerLeftMessage,
  PlayerUpdatedMessage,
  GameLoadingMessage 
} from '../../proto/server_messages';

export interface LobbyState {
  sessionId: string;
  inviteCode: string;
  hostPlayerId: string;
  status: 'WAITING' | 'STARTING' | 'IN_GAME';
  players: LobbyPlayer[];
  settings?: GalaxyGenerateSettings;
}

export interface LobbyEventData {
  type: 'lobby_state' | 'player_joined' | 'player_left' | 'player_updated' | 'settings_updated' | 'game_starting' | 'game_loading' | 'error';
  data: any;
}

export type LobbyEventHandler = (event: LobbyEventData) => void;

export class LobbyService {
  private websocket: ProtobufWebSocketService;
  private currentLobbyState: LobbyState | null = null;
  private eventHandlers = new Set<LobbyEventHandler>();
  private currentPlayerId: string = '';

  constructor(websocket: ProtobufWebSocketService) {
    console.log('🏗️ LobbyService: Constructor called, setting up handlers');
    this.websocket = websocket;
    this.setupWebSocketHandlers();
  }

  private setupWebSocketHandlers() {
    console.log('🔧 LobbyService: Setting up WebSocket handlers');
    // Handle server messages
    this.websocket.onServerMessage((message: ServerMessage) => {
      console.log('LobbyService: Received server message:', message);
      if (message.lobbyMessage) {
        console.log('LobbyService: Found lobby message in server message');
        this.handleLobbyMessage(message.lobbyMessage);
      } else if (message.errorMessage) {
        console.log('LobbyService: Found error message in server message');
        this.emitEvent({
          type: 'error',
          data: {
            code: message.errorMessage.errorCode,
            message: message.errorMessage.errorMessage,
            context: message.errorMessage.context
          }
        });
      } else {
        console.log('LobbyService: Server message contains no lobby or error message');
      }
    });
  }

  private handleLobbyMessage(lobbyMessage: LobbyMessage) {
    console.log('LobbyService: Handling lobby message:', lobbyMessage);
    if (lobbyMessage.lobbyState) {
      console.log('LobbyService: Found lobby state in message');
      this.handleLobbyState(lobbyMessage.lobbyState);
    } else if (lobbyMessage.playerJoined) {
      this.handlePlayerJoined(lobbyMessage.playerJoined);
    } else if (lobbyMessage.playerLeft) {
      this.handlePlayerLeft(lobbyMessage.playerLeft);
    } else if (lobbyMessage.playerUpdated) {
      this.handlePlayerUpdated(lobbyMessage.playerUpdated);
    } else if (lobbyMessage.settingsUpdated) {
      this.handleSettingsUpdated(lobbyMessage.settingsUpdated);
    } else if (lobbyMessage.gameStarting) {
      this.handleGameStarting();
    } else if (lobbyMessage.gameLoading) {
      this.handleGameLoading(lobbyMessage.gameLoading);
    } else {
      console.warn('LobbyService: Unknown lobby message content:', lobbyMessage);
    }
  }

  private handleLobbyState(lobbyState: LobbyStateMessage) {
    console.log('LobbyService: Processing lobby state message:', lobbyState);
    this.currentLobbyState = {
      sessionId: lobbyState.sessionId,
      inviteCode: lobbyState.inviteCode,
      hostPlayerId: lobbyState.hostPlayerId,
      status: this.convertLobbyStatus(lobbyState.status),
      players: [...lobbyState.players],
      settings: lobbyState.settings
    };

    console.log('LobbyService: Emitting lobby_state event with data:', this.currentLobbyState);
    this.emitEvent({
      type: 'lobby_state',
      data: this.currentLobbyState
    });
  }

  private handlePlayerJoined(playerJoined: PlayerJoinedMessage) {
    if (this.currentLobbyState && playerJoined.player) {
      this.currentLobbyState.players.push(playerJoined.player);
      
      this.emitEvent({
        type: 'player_joined',
        data: playerJoined.player
      });
    }
  }

  private handlePlayerLeft(playerLeft: PlayerLeftMessage) {
    if (this.currentLobbyState) {
      this.currentLobbyState.players = this.currentLobbyState.players.filter(
        p => p.playerId !== playerLeft.playerId
      );
      
      this.emitEvent({
        type: 'player_left',
        data: {
          playerId: playerLeft.playerId,
          displayName: playerLeft.displayName
        }
      });
    }
  }

  private handlePlayerUpdated(playerUpdated: PlayerUpdatedMessage) {
    if (this.currentLobbyState && playerUpdated.player) {
      const playerIndex = this.currentLobbyState.players.findIndex(
        p => p.playerId === playerUpdated.player?.playerId
      );
      
      if (playerIndex >= 0) {
        this.currentLobbyState.players[playerIndex] = playerUpdated.player;
        
        this.emitEvent({
          type: 'player_updated',
          data: playerUpdated.player
        });
      }
    }
  }

  private handleSettingsUpdated(settingsUpdated: any) {
    if (this.currentLobbyState && settingsUpdated.settings) {
      this.currentLobbyState.settings = settingsUpdated.settings;
      
      this.emitEvent({
        type: 'settings_updated',
        data: {
          settings: settingsUpdated.settings,
          updatedBy: settingsUpdated.updatedByPlayerId
        }
      });
    }
  }

  private handleGameStarting() {
    if (this.currentLobbyState) {
      this.currentLobbyState.status = 'STARTING';
      
      this.emitEvent({
        type: 'game_starting',
        data: this.currentLobbyState
      });
    }
  }

  private handleGameLoading(gameLoading: GameLoadingMessage) {
    this.emitEvent({
      type: 'game_loading',
      data: {
        progress: gameLoading.progress,
        statusText: gameLoading.statusText,
        phase: gameLoading.phase
      }
    });
  }

  private convertLobbyStatus(status: number): 'WAITING' | 'STARTING' | 'IN_GAME' {
    switch (status) {
      case 0: return 'WAITING';
      case 1: return 'STARTING';
      case 2: return 'IN_GAME';
      default: return 'WAITING';
    }
  }

  private emitEvent(event: LobbyEventData) {
    console.log('📡 LobbyService: Emitting event to', this.eventHandlers.size, 'handlers:', event);
    this.eventHandlers.forEach(handler => handler(event));
  }

  // Public API methods
  setCurrentPlayerId(playerId: string) {
    this.currentPlayerId = playerId;
  }

  getCurrentLobbyState(): LobbyState | null {
    return this.currentLobbyState;
  }

  getCurrentPlayer(): LobbyPlayer | null {
    if (!this.currentLobbyState || !this.currentPlayerId) return null;
    
    return this.currentLobbyState.players.find(
      p => p.playerId === this.currentPlayerId
    ) || null;
  }

  isHost(): boolean {
    const currentPlayer = this.getCurrentPlayer();
    return currentPlayer?.isHost || false;
  }

  canStartGame(): boolean {
    if (!this.currentLobbyState || !this.isHost()) return false;
    
    return this.currentLobbyState.players.length > 0 && 
           this.currentLobbyState.players.every(p => p.isReady);
  }

  // Command methods
  joinLobby(inviteCode: string) {
    console.log('LobbyService: Sending join lobby command with invite code:', inviteCode);
    const joinLobbyCommand = { inviteCode };
    const lobbyCommand: LobbyCommand = { joinLobby: joinLobbyCommand };
    
    this.sendCommand(lobbyCommand);
  }

  setReady(ready: boolean) {
    const setReadyCommand: SetReadyCommand = { ready };
    const lobbyCommand: LobbyCommand = { setReady: setReadyCommand };
    
    this.sendCommand(lobbyCommand);
  }

  setColor(color: string) {
    const setColorCommand: SetColorCommand = { color };
    const lobbyCommand: LobbyCommand = { setColor: setColorCommand };
    
    this.sendCommand(lobbyCommand);
  }

  updateSettings(settings: GalaxyGenerateSettings) {
    if (!this.isHost()) {
      console.warn('Only host can update settings');
      return;
    }
    
    const updateSettingsCommand: UpdateSettingsCommand = { settings };
    const lobbyCommand: LobbyCommand = { updateSettings: updateSettingsCommand };
    
    this.sendCommand(lobbyCommand);
  }

  startGame() {
    if (!this.canStartGame()) {
      console.warn('Cannot start game - not all players ready or not host');
      return;
    }
    
    const lobbyCommand: LobbyCommand = { startGame: {} };
    this.sendCommand(lobbyCommand);
  }

  leaveLobby() {
    const lobbyCommand: LobbyCommand = { leaveLobby: {} };
    this.sendCommand(lobbyCommand);
  }

  private sendCommand(lobbyCommand: LobbyCommand) {
    console.log('LobbyService: Sending command with playerId:', this.currentPlayerId, 'command:', lobbyCommand);
    const clientCommand: ClientCommand = {
      playerId: this.currentPlayerId,
      timestamp: Date.now(),
      lobbyCommand
    };
    
    this.websocket.sendCommand(clientCommand, generateMessageId());
  }

  // Event subscription
  onLobbyEvent(handler: LobbyEventHandler): () => void {
    console.log('🎯 LobbyService: Adding event handler, total handlers:', this.eventHandlers.size + 1);
    this.eventHandlers.add(handler);
    return () => {
      console.log('🗑️ LobbyService: Removing event handler, total handlers:', this.eventHandlers.size - 1);
      this.eventHandlers.delete(handler);
    };
  }

  // Cleanup
  destroy() {
    this.eventHandlers.clear();
    this.currentLobbyState = null;
  }
}
