const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface HealthResponse {
  success: boolean;
  status: string;
  database: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
}

/**
 * Fetch health status from the backend API.
 */
export async function fetchHealth(): Promise<HealthResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/health`);
  if (!response.ok) {
    throw new Error(`Health check failed: ${response.status}`);
  }
  return response.json();
}
