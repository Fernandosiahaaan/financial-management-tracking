/**
 * Application Configuration
 * Centralized DNS & Backend API configuration for the FinTrack Frontend.
 */

/**
 * Normalizes backend URL to ensure it is absolute with proper protocol
 */
export function normalizeBaseUrl(rawUrl?: string): string {
  if (!rawUrl) return 'http://localhost:8080';
  const trimmed = rawUrl.trim();
  if (!trimmed) return 'http://localhost:8080';

  // If already absolute with http:// or https://
  if (/^https?:\/\//i.test(trimmed)) {
    return trimmed.replace(/\/+$/, '');
  }

  // Local development hostnames without protocol
  if (/^localhost(:\d+)?/i.test(trimmed) || /^127\.0\.0\.1(:\d+)?/.test(trimmed)) {
    return `http://${trimmed.replace(/\/+$/, '')}`;
  }

  // Domain names without protocol (e.g. *.railway.app, *.onrender.com)
  return `https://${trimmed.replace(/\/+$/, '')}`;
}

const rawEnvUrl =
  typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_API_URL !== undefined
    ? import.meta.env.VITE_API_URL
    : 'http://localhost:8080';

export const API_BASE_URL: string = normalizeBaseUrl(rawEnvUrl);

/**
 * Constructs a full backend API URL from a relative path.
 * Example: getApiUrl('/api/v1/accounts') -> 'https://your-api.railway.app/api/v1/accounts'
 */
export function getApiUrl(path: string): string {
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  return `${API_BASE_URL}${cleanPath}`;
}

export const config = {
  api: {
    baseUrl: API_BASE_URL,
    getUrl: getApiUrl,
  },
} as const;

export default config;
