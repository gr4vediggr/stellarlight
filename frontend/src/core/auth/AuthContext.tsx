import { createContext, useContext, useEffect, useState } from "react";
import { GameWebSocketService } from "../messaging/gamewebsocketservice";
import { ApiService } from "../api/api_service";
import { getRuntimeConfig } from "../../config/runtimeConfig";
import { useUserStore, type User, type AuthState } from "../store/userStore";

// Auth Context - now integrated with Zustand store
export const AuthContext = createContext<{
  user: User | null;
  token: string | null;
  gameSession: any;
  loading: boolean;
  wsConnected: boolean;
  gameWs: GameWebSocketService;
  apiService: ApiService;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  register: (email: string, password: string, displayName: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => void;
  createGame: () => void;
  joinGame: (inviteCode: string) => void;
  leaveGame: () => void;
} | null>(null);

// Auth Provider
export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  // Get state and actions from Zustand store
  const user = useUserStore((state) => state.user);
  const auth = useUserStore((state) => state.auth);
  const isLoading = useUserStore((state) => state.isLoading);
  
  // Get individual actions to avoid object recreation
  const setUser = useUserStore((state) => state.setUser);
  const setAuth = useUserStore((state) => state.setAuth);
  
  // Local state for game-specific functionality
  const [gameSession, setGameSession] = useState<any>(null);
  const [wsConnected, setWsConnected] = useState(false);
  
  // Initialize services
  const [gameWs] = useState(() => new GameWebSocketService());
  const [apiService] = useState(() => new ApiService(getRuntimeConfig().apiUrl));  

  // Token refresh logic
  const refreshToken = async () => {
    if (!auth?.refreshToken) return false;
    try {
      const response = await apiService.refreshToken();
      const newAuthState: AuthState = {
        token: response.token,
        refreshToken: auth.refreshToken,
        expiresAt: response.expiresAt,
      };
      setAuth(newAuthState);
      return true;
    } catch (error) {
      logout();
      return false;
    }
  };

  // Initialize auth state from the store's persistence
  useEffect(() => {
    // The store handles persistence automatically, but we need to load game session
    const savedGameSession = sessionStorage.getItem('gameSession');
    if (savedGameSession) {
      setGameSession(JSON.parse(savedGameSession));
    }
  }, []);

  // Setup WebSocket connection when token changes
  useEffect(() => {
    const token = auth?.token;
    if (token) {
      gameWs.setAuthToken(token);
      gameWs.setConnectionChangeHandler(setWsConnected);
      gameWs.onTokenExpired = async () => {
        const refreshed = await refreshToken();
        if (refreshed && auth?.token) {
          gameWs.setAuthToken(auth.token);
          gameWs.connect(getRuntimeConfig().wsUrl);
        } else {
          // Optionally show a message to the user
        }
      };
      gameWs.connect(getRuntimeConfig().wsUrl);

      // Subscribe to game events
      const unsubscribeGame = gameWs.subscribeToUpdates('game_joined', (message) => {
        const session = {
          gameId: message.gameId,
          inviteCode: message.inviteCode,
          status: 'joined'
        };
        setGameSession(session);
        sessionStorage.setItem('gameSession', JSON.stringify(session));
      });

      const unsubscribeLeave = gameWs.subscribeToUpdates('game_left', () => {
        setGameSession(null);
        sessionStorage.removeItem('gameSession');
      });

      const unsubscribeError = gameWs.subscribeToUpdates('error', (message) => {
        console.error('Game error:', message.error);
      });

      return () => {
        unsubscribeGame();
        unsubscribeLeave();
        unsubscribeError();
      };
    } else {
      gameWs.disconnect();
    }
  }, [auth?.token, gameWs]);

  const login = async (email: string, password: string) => {
    try {
      const data = await apiService.login(email, password);
      
      // Create User object from response
      const userObj: User = {
        id: data.user.id,
        email: data.user.email,
        displayName: data.user.displayName,
        createdAt: data.user.createdAt,
        updatedAt: data.user.updatedAt,
        profile: data.user.profile
      };

      // Create Auth state
      const authState: AuthState = {
        token: data.token,
        refreshToken: data.refreshToken || null,
        expiresAt: data.expiresAt || (Date.now() + 24 * 60 * 60 * 1000) // Default 24h if not provided
      };
      
      setUser(userObj);
      setAuth(authState);
      
      return { success: true };
    } catch (error: any) {
      return { success: false, error: error.message };
    }
  };

  const register = async (email: string, password: string, displayName: string) => {
    try {
      const data = await apiService.register(email, password, displayName);
      
      // Create User object from response
      const userObj: User = {
        id: data.user.id,
        email: data.user.email,
        displayName: data.user.displayName,
        createdAt: data.user.createdAt,
        updatedAt: data.user.updatedAt,
        profile: data.user.profile
      };

      // Create Auth state
      const authState: AuthState = {
        token: data.token,
        refreshToken: data.refreshToken || null,
        expiresAt: data.expiresAt || (Date.now() + 24 * 60 * 60 * 1000)
      };
      
      setUser(userObj);
      setAuth(authState);
      
      return { success: true };
    } catch (error: any) {
      return { success: false, error: error.message };
    }
  };

  const logout = () => {
    gameWs.disconnect();
    setUser(null);
    setAuth(null);
    setGameSession(null);
    setWsConnected(false);
    sessionStorage.removeItem('gameToken');
    sessionStorage.removeItem('gameUser');
    sessionStorage.removeItem('gameSession');
  };

  const createGame = () => {
    gameWs.sendMessage({ type: 'create_game' });
  };

  const joinGame = (inviteCode: string) => {
    gameWs.sendMessage({ type: 'join_game', inviteCode });
  };

  const leaveGame = () => {
    gameWs.sendMessage({ type: 'leave_game' });
  };

  return (
    <AuthContext.Provider value={{
      user,
      token: auth?.token || null,
      gameSession,
      loading: isLoading,
      wsConnected,
      gameWs,
      apiService,
      login,
      register,
      logout,
      createGame,
      joinGame,
      leaveGame,
    }}>
      {children}
    </AuthContext.Provider>
  );
};

// Custom hook to use auth context
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// Additional convenience hooks that bypass context for better performance
export const useAuthUser = () => useUserStore((state) => state.user);
export const useAuthToken = () => useUserStore((state) => state.auth?.token || null);
export const useIsAuthenticated = () => useUserStore((state) => state.isAuthenticated);
export const useAuthLoading = () => useUserStore((state) => state.isLoading);