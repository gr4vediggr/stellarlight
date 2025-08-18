import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

// User types
export interface User {
  id: string;
  email: string;
  displayName: string;
  createdAt: string;
  updatedAt: string;
  profile?: {
    avatar?: string;
    bio?: string;
    preferences?: UserPreferences;
  };
}

export interface UserPreferences {
  theme: 'light' | 'dark' | 'auto';
  notifications: {
    gameUpdates: boolean;
    chatMessages: boolean;
    diplomacyEvents: boolean;
  };
  gameSettings: {
    autoSave: boolean;
    soundEnabled: boolean;
    musicVolume: number;
    sfxVolume: number;
  };
}

export interface AuthState {
  token: string | null;
  refreshToken: string | null;
  expiresAt: number | null;
}

export interface UserState {
  // User data
  user: User | null;
  auth: AuthState | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  setUser: (user: User | null) => void;
  setAuth: (auth: AuthState | null) => void;
  updateUser: (updates: Partial<User>) => void;
  updatePreferences: (preferences: Partial<UserPreferences>) => void;
  setLoading: (loading: boolean) => void;  setError: (error: string | null) => void;
  logout: () => void;
  isTokenExpired: () => boolean;
}

// Default preferences
const defaultPreferences: UserPreferences = {
  theme: 'auto',
  notifications: {
    gameUpdates: true,
    chatMessages: true,
    diplomacyEvents: true,
  },
  gameSettings: {
    autoSave: true,
    soundEnabled: true,
    musicVolume: 0.7,
    sfxVolume: 0.8,
  },
};

export const useUserStore = create<UserState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      auth: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      setUser: (user: User | null) => {
        set((state) => ({
          user,
          isAuthenticated: !!user && !!state.auth?.token,
        }));
      },

      setAuth: (auth: AuthState | null) => {
        set((state) => ({
          auth,
          isAuthenticated: !!state.user && !!auth?.token,
        }));
      },

      updateUser: (updates: Partial<User>) => {
        set((state) => ({
          user: state.user ? { ...state.user, ...updates } : null,
        }));
      },

      updatePreferences: (preferences: Partial<UserPreferences>) => {
        set((state) => ({
          user: state.user
            ? {
                ...state.user,
                profile: {
                  ...state.user.profile,
                  preferences: {
                    ...defaultPreferences,
                    ...state.user.profile?.preferences,
                    ...preferences,
                  },
                },
              }
            : null,
        }));
      },

      setLoading: (isLoading: boolean) => set({ isLoading }),

      setError: (error: string | null) => set({ error }),

      logout: () => {
        set({
          user: null,
          auth: null,
          isAuthenticated: false,
          error: null,
        });
      },      isTokenExpired: () => {
        const { auth } = get();
        if (!auth?.expiresAt) return true;
        return Date.now() >= auth.expiresAt;
      },
    }),
    {
      name: 'stellarlight-user-storage',
      storage: createJSONStorage(() => sessionStorage),
      partialize: (state) => ({
        user: state.user,
        auth: state.auth,
      }),
    }
  )
);

// Selectors for better performance
export const useUser = () => useUserStore((state) => state.user);
export const useAuth = () => useUserStore((state) => state.auth);
export const useIsAuthenticated = () => useUserStore((state) => state.isAuthenticated);
export const useUserPreferences = () => 
  useUserStore((state) => state.user?.profile?.preferences || defaultPreferences);

// Computed selectors
export const useDisplayName = () => useUserStore((state) => {
  const user = state.user;
  return user?.displayName || user?.email || 'Anonymous';
});

export const useInitials = () => useUserStore((state) => {
  const user = state.user;
  if (!user?.displayName) return 'A';
  return user.displayName
    .split(' ')
    .map((name) => name[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
});

// Action selectors
export const useUserActions = () => useUserStore((state) => ({
  setUser: state.setUser,
  setAuth: state.setAuth,
  updateUser: state.updateUser,
  updatePreferences: state.updatePreferences,
  setLoading: state.setLoading,
  setError: state.setError,
  logout: state.logout,
  isTokenExpired: state.isTokenExpired,
}));
