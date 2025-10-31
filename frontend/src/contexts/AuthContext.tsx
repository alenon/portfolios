import React, { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import authService from '../services/authService';
import type { User } from '../types/auth';
import { setTokens, getRefreshToken, clearTokens, isAuthenticated as checkAuth } from '../utils/tokenStorage';

// Auth context interface
interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string, rememberMe: boolean) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshAuth: () => Promise<void>;
}

// Create context with undefined default
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider props
interface AuthProviderProps {
  children: ReactNode;
}

/**
 * AuthProvider component
 * Manages authentication state and provides auth functions to the app
 */
export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  /**
   * Load user on mount if tokens exist
   */
  useEffect(() => {
    const loadUser = async () => {
      if (checkAuth()) {
        try {
          const userData = await authService.getCurrentUser();
          setUser(userData);
          setIsAuthenticated(true);
        } catch (error) {
          console.error('Failed to load user:', error);
          clearTokens();
          setUser(null);
          setIsAuthenticated(false);
        }
      }
      setIsLoading(false);
    };

    loadUser();
  }, []);

  /**
   * Login function
   */
  const login = async (email: string, password: string, rememberMe: boolean): Promise<void> => {
    setIsLoading(true);
    try {
      const response = await authService.login(email, password, rememberMe);

      // Store tokens
      setTokens(response.access_token, response.refresh_token, response.expires_in);

      // Update state
      setUser(response.user);
      setIsAuthenticated(true);
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Register function
   */
  const register = async (email: string, password: string): Promise<void> => {
    setIsLoading(true);
    try {
      const response = await authService.register(email, password);

      // Store tokens
      setTokens(response.access_token, response.refresh_token, response.expires_in);

      // Update state
      setUser(response.user);
      setIsAuthenticated(true);
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Logout function
   */
  const logout = async (): Promise<void> => {
    setIsLoading(true);
    try {
      const refreshToken = getRefreshToken();

      if (refreshToken) {
        await authService.logout(refreshToken);
      }
    } catch (error) {
      console.error('Logout failed:', error);
      // Continue with logout even if API call fails
    } finally {
      // Clear tokens and state
      clearTokens();
      setUser(null);
      setIsAuthenticated(false);
      setIsLoading(false);
    }
  };

  /**
   * Refresh auth state (reload user data)
   */
  const refreshAuth = async (): Promise<void> => {
    if (checkAuth()) {
      try {
        const userData = await authService.getCurrentUser();
        setUser(userData);
        setIsAuthenticated(true);
      } catch (error) {
        console.error('Failed to refresh auth:', error);
        clearTokens();
        setUser(null);
        setIsAuthenticated(false);
      }
    }
  };

  const value: AuthContextType = {
    user,
    isAuthenticated,
    isLoading,
    login,
    register,
    logout,
    refreshAuth,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

/**
 * Custom hook to use auth context
 */
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);

  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return context;
};
