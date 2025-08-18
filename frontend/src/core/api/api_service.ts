
// API Service for HTTP requests
export class ApiService {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request(endpoint: string, options: RequestInit = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const token = sessionStorage.getItem('gameToken');
    
    const config: RequestInit = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers,
      },
    };

    const response = await fetch(url, config);
    
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
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

  async updateProfile(data: any) {
    return this.request('/api/users/update-profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async refreshToken() {
    return this.request('/api/auth/refresh', {
      method: 'POST',
    });
  }
}