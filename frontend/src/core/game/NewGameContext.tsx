import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useWebSocket } from '../websocket/WebSocketContext';
import { GameCommand } from '../../proto/client_commands';
import { GameMessage } from '../../proto/server_messages';

interface GameState {
  gameId: string | null;
  turnNumber: number;
  gameTime: number;
  isLoading: boolean;
  loadingProgress: number;
  loadingText: string;
  stateData: any; // Parsed game state
  players: GamePlayer[];
  currentPlayerId: string | null;
}

interface GamePlayer {
  id: string;
  name: string;
  color: string;
  isActive: boolean;
}

interface GameContextType {
  state: GameState;
  loading: boolean;
  error: string | null;
  // Actions
  sendGameCommand: (command: GameCommand) => void;
  loadGame: (gameId: string) => void;
  // Utils
  isCurrentPlayerTurn: () => boolean;
}

const GameContext = createContext<GameContextType | undefined>(undefined);

export const useGame = () => {
  const context = useContext(GameContext);
  if (!context) {
    throw new Error('useGame must be used within a GameProvider');
  }
  return context;
};

interface GameProviderProps {
  children: React.ReactNode;
}

export const GameProvider: React.FC<GameProviderProps> = ({ children }) => {
  const { wsService, isConnected } = useWebSocket();
  const [state, setState] = useState<GameState>({
    gameId: null,
    turnNumber: 0,
    gameTime: 0,
    isLoading: false,
    loadingProgress: 0,
    loadingText: '',
    stateData: null,
    players: [],
    currentPlayerId: null
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Subscribe to game messages
  useEffect(() => {
    if (!wsService) return;

    const unsubscribe = wsService.onGameMessage((message: GameMessage) => {
      console.log('Game message received:', message);
      
      if (message.gameState) {
        setState(prev => ({
          ...prev,
          turnNumber: message.gameState!.turnNumber,
          gameTime: message.gameState!.gameTime,
          stateData: JSON.parse(message.gameState!.stateData)
        }));
      }

      if (message.gameEvent) {
        // Handle game events
        console.log('Game event:', message.gameEvent);
        // Process events based on event type
      }

      if (message.turnUpdate) {
        setState(prev => ({
          ...prev,
          turnNumber: message.turnUpdate!.turnNumber
        }));
      }
    });

    return unsubscribe;
  }, [wsService]);

  const sendGameCommand = useCallback((command: GameCommand) => {
    if (wsService && isConnected && state.currentPlayerId) {
      wsService.sendGameCommand(state.currentPlayerId, command);
    } else {
      console.warn('Cannot send game command: WebSocket not connected or no player ID');
    }
  }, [wsService, isConnected, state.currentPlayerId]);

  const loadGame = useCallback((gameId: string) => {
    setLoading(true);
    setError(null);
    
    try {
      // Get player ID from session
      const sessionData = localStorage.getItem('currentSession');
      if (!sessionData) {
        throw new Error('No valid session found');
      }

      const session = JSON.parse(sessionData);
      setState(prev => ({
        ...prev,
        gameId: gameId,
        currentPlayerId: session.playerId,
        isLoading: true,
        loadingProgress: 0,
        loadingText: 'Loading game...'
      }));
    } catch (error: any) {
      setError(error.message);
      console.error('Failed to load game:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  const isCurrentPlayerTurn = useCallback(() => {
    // This would depend on your game's turn logic
    // For now, just return true if we have a current player
    return state.currentPlayerId !== null;
  }, [state.currentPlayerId]);

  const value: GameContextType = {
    state,
    loading,
    error,
    sendGameCommand,
    loadGame,
    isCurrentPlayerTurn
  };

  return (
    <GameContext.Provider value={value}>
      {children}
    </GameContext.Provider>
  );
};
