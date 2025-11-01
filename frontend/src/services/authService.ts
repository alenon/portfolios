import api from "./api";
import type {
  AuthResponse,
  User,
  MessageResponse,
  RefreshTokenResponse,
} from "../types/auth";

/**
 * Authentication Service
 * Handles all authentication-related API calls
 */
class AuthService {
  /**
   * Register a new user
   */
  async register(email: string, password: string): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>("/auth/register", {
      email,
      password,
    });
    return response.data;
  }

  /**
   * Login with email and password
   */
  async login(
    email: string,
    password: string,
    rememberMe: boolean,
  ): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>("/auth/login", {
      email,
      password,
      remember_me: rememberMe,
    });
    return response.data;
  }

  /**
   * Refresh access token using refresh token
   */
  async refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
    const response = await api.post<RefreshTokenResponse>("/auth/refresh", {
      refresh_token: refreshToken,
    });
    return response.data;
  }

  /**
   * Logout and revoke refresh token
   */
  async logout(refreshToken: string): Promise<void> {
    await api.post("/auth/logout", {
      refresh_token: refreshToken,
    });
  }

  /**
   * Get current authenticated user profile
   */
  async getCurrentUser(): Promise<User> {
    const response = await api.get<User>("/auth/me");
    return response.data;
  }

  /**
   * Request password reset email
   */
  async forgotPassword(email: string): Promise<MessageResponse> {
    const response = await api.post<MessageResponse>("/auth/forgot-password", {
      email,
    });
    return response.data;
  }

  /**
   * Reset password with token
   */
  async resetPassword(
    token: string,
    newPassword: string,
  ): Promise<MessageResponse> {
    const response = await api.post<MessageResponse>("/auth/reset-password", {
      token,
      new_password: newPassword,
    });
    return response.data;
  }
}

// Export singleton instance
export default new AuthService();
