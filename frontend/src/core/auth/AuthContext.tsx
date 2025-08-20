import { createContext, useContext } from "react";
import { ApiService } from "../api/api_service";
import { getRuntimeConfig } from "../../config/runtimeConfig";
import { useUserStore, type User, type AuthState } from "../store/userStore";

// Auth Context - focused only on authentication
export const AuthContext = createContext<{
  user: User | null;
  token: string | null;
  loading: boolean;
  apiService: ApiService;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  register: (email: string, password: string, displayName: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => void;
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
  
  // Initialize API service with token getter
  const apiService = new ApiService(
    getRuntimeConfig().apiUrl, 
    () => useUserStore.getState().auth?.token || null
  );  

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
    setUser(null);
    setAuth(null);
    sessionStorage.removeItem('gameToken');
    sessionStorage.removeItem('gameUser');
  };

  return (
    <AuthContext.Provider value={{
      user,
      token: auth?.token || null,
      loading: isLoading,
      apiService,
      login,
      register,
      logout,
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