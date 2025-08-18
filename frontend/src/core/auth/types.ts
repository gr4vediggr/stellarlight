// Auth-related types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  displayName: string;
}

export interface AuthResponse {
  user: {
    id: string;
    email: string;
    displayName: string;
    createdAt: string;
    updatedAt: string;
  };
  token: string;
  refreshToken?: string;
  expiresAt: number;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface RefreshTokenResponse {
  token: string;
  expiresAt: number;
}

// Error types
export interface AuthError {
  code: string;
  message: string;
  details?: Record<string, any>;
}

// Auth service interface
export interface IAuthService {
  login(request: LoginRequest): Promise<AuthResponse>;
  register(request: RegisterRequest): Promise<AuthResponse>;
  refreshToken(request: RefreshTokenRequest): Promise<RefreshTokenResponse>;
  logout(): Promise<void>;
  validateToken(token: string): Promise<boolean>;
}
