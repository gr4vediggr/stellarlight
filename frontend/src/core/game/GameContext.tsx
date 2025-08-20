import React, { createContext, useContext, useState, useEffect } from 'react';
import { GameService } from './GameService';
import type { GameSession, GameEventType, LobbyState } from './GameService';
import { getRuntimeConfig } from '../../config/runtimeConfig';
import { useAuth } from '../auth/AuthContext';

interface GameContextType {
  gameService: GameService;
  currentSession: GameSession | null;
  lobbyState: LobbyState | null; // Using GameService's state format
  isConnected: boolean;
  loading: boolean;
  error: string | null;
  
  // Game management
  createGame: () => Promise<void>;
  joinGame: (inviteCode: string) => Promise<void>;
  leaveGame: () => Promise<void>;
  reconnectToGame: () => Promise<void>;
  
  // Lobby actions (delegated to LobbyService)
  setReady: (ready: boolean) => void;
  setColor: (color: string) => void;
  updateSettings: (settings: any) => void;
  startGame: () => void;
  
  // Event subscription
  onGameEvent: (event: GameEventType, callback: (data: any) => void) => () => void;
  onLobbyEvent: (callback: (event: any) => void) => () => void;
}

const GameContext = createContext<GameContextType | null>(null);

export const GameProvider = ({ children }: { children: React.ReactNode }) => {
  const { token, apiService, user } = useAuth();
  const [gameService] = useState(() => new GameService(apiService));
  const [currentSession, setCurrentSession] = useState<GameSession | null>(null);
  const [lobbyState, setLobbyState] = useState<any | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [loading, setLoading] = useState(true); // Start as loading
  const [error, setError] = useState<string | null>(null);

  // Initialize game service and check for existing session
  useEffect(() => {
    const initializeGame = async () => {
      if (!token) {
        setLoading(false);
        return;
      }
      
      setLoading(true); // Set loading when starting initialization
      try {
        console.log('🔍 GameContext: Starting initialization...');
        // Check if user is already in a game
        const existingSession = await gameService.getCurrentGame();
        if (existingSession) {
          console.log('🔍 GameContext: Found existing session:', existingSession);
          setCurrentSession(existingSession);
          
          // Auto-connect to lobby for existing sessions
          console.log('🔍 GameContext: Auto-connecting to existing session lobby...');
          await autoConnectToLobby(existingSession);
        }
        console.log('🔍 GameContext: Initialization complete');
      } catch (error) {
        console.log('🔍 GameContext: No existing game session');
      } finally {
        setLoading(false); // Always clear loading after initialization
      }
    };

    if (token) {
      initializeGame();
    } else {
      setCurrentSession(null);
      setLobbyState(null);
      setIsConnected(false);
      setLoading(false);
    }
  }, [token, gameService]);

  // Create game
  const createGame = async () => {
    try {
      setError(null);
      const session = await gameService.createGame();
      setCurrentSession(session);
      console.log('🎮 GameContext: Game created:', session);
      
      // Connect to lobby immediately after creation
      await autoConnectToLobby(session);
    } catch (error) {
      console.error('Failed to create game:', error);
      setError(error instanceof Error ? error.message : 'Failed to create game');
      throw error;
    }
  };

  // Join game
  const joinGame = async (inviteCode: string) => {
    try {
      setError(null);
      const session = await gameService.joinGame(inviteCode);
      setCurrentSession(session);
      console.log('🎮 GameContext: Joined game:', session);
      
      // Connect to lobby immediately after joining
      await autoConnectToLobby(session);
    } catch (error) {
      console.error('Failed to join game:', error);
      setError(error instanceof Error ? error.message : 'Failed to join game');
      throw error;
    }
  };

  // Leave game
  const leaveGame = async () => {
    try {
      setError(null);
      
      // Disconnect and cleanup
      gameService.disconnect();
      setLobbyState(null);
      setIsConnected(false);
      
      // Leave the game session
      await gameService.leaveGame();
      
      setCurrentSession(null);
      console.log('🎮 GameContext: Left game');
    } catch (error) {
      console.error('Failed to leave game:', error);
      setError(error instanceof Error ? error.message : 'Failed to leave game');
      throw error;
    }
  };

  // Reconnect to existing game
  const reconnectToGame = async () => {
    try {
      setError(null);
      const session = await gameService.getCurrentGame();
      if (session) {
        setCurrentSession(session);
        await autoConnectToLobby(session);
      }
    } catch (error) {
      console.error('Failed to reconnect to game:', error);
      setError(error instanceof Error ? error.message : 'Failed to reconnect to game');
      throw error;
    }
  };

  // Connect to lobby
  const autoConnectToLobby = async (session?: GameSession) => {
    const targetSession = session || currentSession;
    console.log('🔌 GameContext: autoConnectToLobby called with session:', targetSession);
    
    if (!targetSession) {
      console.log('🔌 GameContext: No session provided, cannot connect to lobby');
      throw new Error('No session provided');
    }

    if (!token || !user) {
      console.log('🔌 GameContext: No token or user, cannot connect to lobby');
      throw new Error('Not authenticated');
    }

    try {
      setError(null);
      setLoading(true);
      
      console.log('🔌 GameContext: Setting up WebSocket connection...');
      
      // Configure WebSocket service
      const config = getRuntimeConfig();
      const wsUrl = config.wsUrl || 'wss://localhost:8080/ws';
      console.log('🔌 GameContext: Using WebSocket URL:', wsUrl);
      
      // Connect WebSocket through GameService
      gameService.connect(wsUrl, token);
      console.log('🔌 GameContext: WebSocket connected via GameService');
      
      // Set up GameService event handlers for lobby updates
      gameService.on('lobby_updated', (lobbyState: any) => {
        console.log('🏛️ GameContext: Lobby state updated via GameService:', lobbyState);
        setLobbyState(lobbyState);
        setIsConnected(true);
      });
      
      // No need to send join lobby command - backend will automatically send lobby state
      console.log('🏛️ GameContext: WebSocket connected, waiting for automatic lobby state from backend...');
      
      setIsConnected(true);
      console.log('🔌 GameContext: Successfully connected to lobby');
      
    } catch (error) {
      console.error('🔌 GameContext: Failed to connect to lobby:', error);
      setError(error instanceof Error ? error.message : 'Failed to connect to lobby');
      throw error;
    } finally {
      setLoading(false);
    }
  };

  // Lobby actions
  const setReady = (ready: boolean) => {
    if (user && gameService.isConnected()) {
      gameService.setPlayerReady(user.id, ready);
    }
  };

  const setColor = (color: string) => {
    if (user && gameService.isConnected()) {
      gameService.setPlayerColor(user.id, color);
    }
  };

  const updateSettings = (settings: any) => {
    if (gameService.isConnected()) {
      gameService.updateGameSettings(settings);
    }
  };

  const startGame = () => {
    if (gameService.isConnected()) {
      gameService.startGame();
    }
  };

  // Event subscriptions
  const onGameEvent = (event: GameEventType, callback: (data: any) => void) => {
    return gameService.on(event, callback);
  };

  const onLobbyEvent = (callback: (event: any) => void) => {
    return gameService.on('lobby_updated', callback);
  };

  const contextValue: GameContextType = {
    gameService,
    currentSession,
    lobbyState,
    isConnected,
    loading,
    error,
    createGame,
    joinGame,
    leaveGame,
    reconnectToGame,
    setReady,
    setColor,
    updateSettings,
    startGame,
    onGameEvent,
    onLobbyEvent,
  };

  return (
    <GameContext.Provider value={contextValue}>
      {children}
    </GameContext.Provider>
  );
};

export const useGame = () => {
  const context = useContext(GameContext);
  if (!context) {
    throw new Error('useGame must be used within a GameProvider');
  }
  return context;
};

export default GameContext;
