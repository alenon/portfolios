// Token storage keys
const ACCESS_TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';
const TOKEN_EXPIRY_KEY = 'token_expiry';

/**
 * Store access and refresh tokens in localStorage
 */
export const setTokens = (accessToken: string, refreshToken: string, expiresIn?: number): void => {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);

  // Store expiration time if provided (current time + expires_in seconds)
  if (expiresIn) {
    const expiryTime = Date.now() + (expiresIn * 1000);
    localStorage.setItem(TOKEN_EXPIRY_KEY, expiryTime.toString());
  }
};

/**
 * Get access token from localStorage
 */
export const getAccessToken = (): string | null => {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
};

/**
 * Get refresh token from localStorage
 */
export const getRefreshToken = (): string | null => {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
};

/**
 * Clear all tokens from localStorage
 */
export const clearTokens = (): void => {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(TOKEN_EXPIRY_KEY);
};

/**
 * Check if user is authenticated (has valid tokens)
 */
export const isAuthenticated = (): boolean => {
  const accessToken = getAccessToken();
  const refreshToken = getRefreshToken();
  return !!(accessToken && refreshToken);
};

/**
 * Check if access token is expired or about to expire
 * Returns true if token will expire in less than 5 minutes
 */
export const isTokenExpired = (): boolean => {
  const expiryTime = localStorage.getItem(TOKEN_EXPIRY_KEY);

  if (!expiryTime) {
    return false; // No expiry data, assume not expired
  }

  const expiry = parseInt(expiryTime, 10);
  const now = Date.now();
  const fiveMinutes = 5 * 60 * 1000; // 5 minutes in milliseconds

  // Return true if token expires in less than 5 minutes
  return expiry - now < fiveMinutes;
};

/**
 * Get time until token expiration in milliseconds
 */
export const getTimeUntilExpiry = (): number | null => {
  const expiryTime = localStorage.getItem(TOKEN_EXPIRY_KEY);

  if (!expiryTime) {
    return null;
  }

  const expiry = parseInt(expiryTime, 10);
  const now = Date.now();

  return Math.max(0, expiry - now);
};
