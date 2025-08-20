
// API Service for HTTP requests
export class ApiService {
  private baseUrl: string;
  private getToken?: () => string | null;

  constructor(baseUrl: string, getToken?: () => string | null) {
    this.baseUrl = baseUrl;
    this.getToken = getToken;
  }

  private async request(endpoint: string, options: RequestInit = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    
    // Get token from provided getter function or fallback to storage parsing
    let token = null;
    if (this.getToken) {
      token = this.getToken();
    } else {
      // Fallback: parse from sessionStorage directly
      const persistedData = sessionStorage.getItem('stellarlight-user-storage');
      if (persistedData) {
        try {
          const parsed = JSON.parse(persistedData);
          token = parsed.state?.auth?.token;
        } catch (error) {
          console.error('Failed to parse auth data:', error);
        }
      }
    }
    
    const config: RequestInit = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers,
      },
    };

    let response = await fetch(url, config);
    
    // If we get 401 and have a refresh token, try to refresh
    if (response.status === 401 && !endpoint.includes('/refresh')) {
      console.log('🔄 ApiService: Token expired (401), attempting refresh...');
      const refreshed = await this.attemptTokenRefresh();
      if (refreshed) {
        // Retry with new token
        const newToken = this.getToken?.();
        config.headers = {
          ...config.headers,
          ...(newToken && { 'Authorization': `Bearer ${newToken}` }),
        };
        response = await fetch(url, config);
      }
    }
    
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }

  private async attemptTokenRefresh(): Promise<boolean> {
    try {
      const persistedData = sessionStorage.getItem('stellarlight-user-storage');
      if (!persistedData) return false;
      
      const parsed = JSON.parse(persistedData);
      const refreshToken = parsed.state?.auth?.refreshToken;
      if (!refreshToken) return false;

      console.log('🔄 ApiService: Using refresh token...');
      const response = await fetch(`${this.baseUrl}/api/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken }),
      });

      if (!response.ok) {
        console.log('🔄 ApiService: Refresh failed, user needs to re-login');
        return false;
      }

      const data = await response.json();
      
      // Update tokens in store
      const currentState = JSON.parse(persistedData);
      currentState.state.auth.token = data.token;
      if (data.refreshToken) {
        currentState.state.auth.refreshToken = data.refreshToken;
      }
      currentState.state.auth.expiresAt = data.expiresAt || (Date.now() + 15 * 60 * 1000);
      
      sessionStorage.setItem('stellarlight-user-storage', JSON.stringify(currentState));
      
      console.log('✅ ApiService: Token refreshed successfully');
      return true;
    } catch (error) {
      console.error('🔄 ApiService: Token refresh failed:', error);
      return false;
    }
  }

  async login(email: string, password: string) {
    return this.request('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  }

  async register(email: string, password: string, displayName: string) {
    return this.request('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, displayName }),
    });
  }

  async refreshToken(refreshToken: string) {
    return this.request('/api/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    });
  }

  async updateProfile(data: any) {
    return this.request('/api/users/update-profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  // Game-related API methods
  async createGame() {
    return this.request('/api/game/create', {
      method: 'POST',
    });
  }

  async joinGame(inviteCode: string) {
    return this.request('/api/game/join', {
      method: 'POST',
      body: JSON.stringify({ invite_code: inviteCode }),
    });
  }

  async leaveGame() {
    return this.request('/api/game/leave', {
      method: 'POST',
    });
  }

  async getCurrentGame() {
    return this.request('/api/game/current', {
      method: 'GET',
    });
  }

  async startGame() {
    return this.request('/api/game/start', {
      method: 'POST',
    });
  }
}