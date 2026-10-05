/**
 * Application Configuration
 * Centralized DNS & Backend API configuration for the FinTrack Frontend.
 */

// Reads VITE_API_URL from environment variables (e.g. .env, Vercel, or Docker).
// Default is 'http://localhost:8080'.
// If set to empty string "", requests will be relative to current origin (e.g. for reverse proxies).
export const API_BASE_URL: string =
  typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_API_URL !== undefined
    ? import.meta.env.VITE_API_URL
    : 'http://localhost:8080';

/**
 * Constructs a full backend API URL from a relative path.
 * Example: getApiUrl('/api/v1/accounts') -> 'http://localhost:8080/api/v1/accounts'
 */
export function getApiUrl(path: string): string {
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  if (!API_BASE_URL) {
    return cleanPath;
  }
  return `${API_BASE_URL.replace(/\/+$/, '')}${cleanPath}`;
}

export const config = {
  api: {
    baseUrl: API_BASE_URL,
    getUrl: getApiUrl,
  },
} as const;

export default config;
