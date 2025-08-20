import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useWebSocket } from '../websocket/WebSocketContext';
import { LobbyCommand } from '../../proto/client_commands';
import { LobbyMessage } from '../../proto/server_messages';

interface Player {
  id: string;
  name: string;
  isReady: boolean;
  isHost: boolean;
}

interface LobbyState {
  gameId: string | null;
  players: Player[];
  gameSettings: {
    maxPlayers: number;
    galaxySize: string;
    gameSpeed: string;
  };
  isHost: boolean;
  isReady: boolean;
  isGameStarting: boolean;
}

interface LobbyContextType {
  state: LobbyState;
  loading: boolean;
  error: string | null;
  // Actions
  joinLobby: (gameId: string) => Promise<void>;
  leaveLobby: () => void;
  toggleReady: () => void;
  updateGameSettings: (settings: Partial<LobbyState['gameSettings']>) => void;
  startGame: () => void;
  // Utils
  canStartGame: () => boolean;
  getCurrentPlayer: () => Player | null;
}

const LobbyContext = createContext<LobbyContextType | undefined>(undefined);

export const useLobby = () => {
  const context = useContext(LobbyContext);
  if (!context) {
    throw new Error('useLobby must be used within a LobbyProvider');
  }
  return context;
};

interface LobbyProviderProps {
  children: React.ReactNode;
}

export const LobbyProvider: React.FC<LobbyProviderProps> = ({ children }) => {
  const { wsService, isConnected, subscribe } = useWebSocket();
  const [state, setState] = useState<LobbyState>({
    gameId: null,
    players: [],
    gameSettings: {
      maxPlayers: 4,
      galaxySize: 'medium',
      gameSpeed: 'normal'
    },
    isHost: false,
    isReady: false,
    isGameStarting: false
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Subscribe to lobby messages
  useEffect(() => {
    if (!wsService) return;

    const unsubscribe = wsService.onLobbyMessage((message: LobbyMessage) => {
      console.log('Lobby message received:', message);
      
      if (message.lobbyState) {
        setState(prev => ({
          ...prev,
          gameId: message.lobbyState!.sessionId,
          players: message.lobbyState!.players.map(p => ({
            id: p.playerId,
            name: p.displayName,
            isReady: p.isReady,
            isHost: p.isHost
          })),
          gameSettings: {
            maxPlayers: prev.gameSettings.maxPlayers, // Use current settings for now
            galaxySize: prev.gameSettings.galaxySize,
            gameSpeed: prev.gameSettings.gameSpeed
          }
        }));
      }

      if (message.playerJoined) {
        const player = message.playerJoined.player;
        if (player) {
          setState(prev => ({
            ...prev,
            players: [...prev.players.filter(p => p.id !== player.playerId), {
              id: player.playerId,
              name: player.displayName,
              isReady: player.isReady,
              isHost: player.isHost
            }]
          }));
        }
      }

      if (message.playerLeft) {
        setState(prev => ({
          ...prev,
          players: prev.players.filter(p => p.id !== message.playerLeft!.playerId)
        }));
      }

      if (message.playerUpdated) {
        const player = message.playerUpdated.player;
        if (player) {
          setState(prev => ({
            ...prev,
            players: prev.players.map(p => 
              p.id === player.playerId 
                ? { ...p, isReady: player.isReady }
                : p
            )
          }));
        }
      }

      if (message.settingsUpdated) {
        if (message.settingsUpdated.settings) {
          setState(prev => ({
            ...prev,
            gameSettings: {
              maxPlayers: prev.gameSettings.maxPlayers,
              galaxySize: prev.gameSettings.galaxySize,
              gameSpeed: prev.gameSettings.gameSpeed
            }
          }));
        }
      }

      if (message.gameStarting) {
        setState(prev => ({ ...prev, isGameStarting: true }));
      }
    });

    return unsubscribe;
  }, [wsService]);

  const joinLobby = useCallback(async (gameId: string) => {
    setLoading(true);
    setError(null);
    
    try {
      // Get player ID from current session
      const sessionData = localStorage.getItem('currentSession');
      if (!sessionData) {
        throw new Error('No valid session found');
      }

      const session = JSON.parse(sessionData);
      if (!session.playerId) {
        throw new Error('No player ID in session');
      }

      // Send join lobby command via WebSocket
      if (wsService && isConnected) {
        const joinCommand: LobbyCommand = {
          joinLobby: {
            inviteCode: gameId
          }
        };
        
        wsService.sendLobbyCommand(session.playerId, joinCommand);
        setState(prev => ({ 
          ...prev, 
          gameId: gameId,
          isHost: false // We'll get the real value from the server
        }));
      } else {
        throw new Error('WebSocket not connected');
      }
    } catch (error: any) {
      setError(error.message);
      console.error('Failed to join lobby:', error);
    } finally {
      setLoading(false);
    }
  }, [wsService, isConnected]);

  const leaveLobby = useCallback(() => {
    if (wsService && isConnected && state.gameId) {
      // Get player ID from current user
      const session = localStorage.getItem('currentSession');
      if (session) {
        const sessionData = JSON.parse(session);
        const leaveCommand: LobbyCommand = {
          leaveLobby: {}
        };
        
        wsService.sendLobbyCommand(sessionData.playerId, leaveCommand);
      }
    }
    
    // Reset state
    setState({
      gameId: null,
      players: [],
      gameSettings: {
        maxPlayers: 4,
        galaxySize: 'medium',
        gameSpeed: 'normal'
      },
      isHost: false,
      isReady: false,
      isGameStarting: false
    });
  }, [wsService, isConnected, state.gameId]);

  const toggleReady = useCallback(() => {
    if (wsService && isConnected) {
      const session = localStorage.getItem('currentSession');
      if (session) {
        const sessionData = JSON.parse(session);
        const readyCommand: LobbyCommand = {
          setReady: {
            ready: !state.isReady
          }
        };
        
        wsService.sendLobbyCommand(sessionData.playerId, readyCommand);
        setState(prev => ({ ...prev, isReady: !prev.isReady }));
      }
    }
  }, [wsService, isConnected, state.isReady]);

  const updateGameSettings = useCallback((settings: Partial<LobbyState['gameSettings']>) => {
    if (wsService && isConnected && state.isHost) {
      const session = localStorage.getItem('currentSession');
      if (session) {
        const sessionData = JSON.parse(session);
        // Note: This needs the actual GalaxyGenerateSettings structure
        const settingsCommand: LobbyCommand = {
          updateSettings: {
            settings: {
              // Add proper settings structure here based on proto definition
            }
          }
        };
        
        wsService.sendLobbyCommand(sessionData.playerId, settingsCommand);
        setState(prev => ({
          ...prev,
          gameSettings: { ...prev.gameSettings, ...settings }
        }));
      }
    }
  }, [wsService, isConnected, state.isHost, state.gameSettings]);

  const startGame = useCallback(() => {
    if (wsService && isConnected && state.isHost && canStartGame()) {
      const session = localStorage.getItem('currentSession');
      if (session) {
        const sessionData = JSON.parse(session);
        const startCommand: LobbyCommand = {
          startGame: {}
        };
        
        wsService.sendLobbyCommand(sessionData.playerId, startCommand);
      }
    }
  }, [wsService, isConnected, state.isHost]);

  const canStartGame = useCallback(() => {
    return state.isHost && 
           state.players.length >= 2 && 
           state.players.every(p => p.isReady || p.isHost) &&
           !state.isGameStarting;
  }, [state.isHost, state.players, state.isGameStarting]);

  const getCurrentPlayer = useCallback(() => {
    const session = localStorage.getItem('currentSession');
    if (session) {
      const sessionData = JSON.parse(session);
      return state.players.find(p => p.id === sessionData.playerId) || null;
    }
    return null;
  }, [state.players]);

  const value: LobbyContextType = {
    state,
    loading,
    error,
    joinLobby,
    leaveLobby,
    toggleReady,
    updateGameSettings,
    startGame,
    canStartGame,
    getCurrentPlayer
  };

  return (
    <LobbyContext.Provider value={value}>
      {children}
    </LobbyContext.Provider>
  );
};
