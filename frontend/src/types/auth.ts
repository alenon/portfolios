// User data structure
export interface User {
  id: string;
  email: string;
  created_at: string;
  last_login_at?: string;
}

// Authentication response from API
export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

// Registration request payload
export interface RegisterRequest {
  email: string;
  password: string;
}

// Login request payload
export interface LoginRequest {
  email: string;
  password: string;
  remember_me: boolean;
}

// Password reset request payload
export interface ResetPasswordRequest {
  token: string;
  new_password: string;
}

// Forgot password request payload
export interface ForgotPasswordRequest {
  email: string;
}

// Refresh token request payload
export interface RefreshTokenRequest {
  refresh_token: string;
}

// Logout request payload
export interface LogoutRequest {
  refresh_token: string;
}

// Generic message response
export interface MessageResponse {
  message: string;
}

// Token refresh response
export interface RefreshTokenResponse {
  access_token: string;
  expires_in: number;
}
