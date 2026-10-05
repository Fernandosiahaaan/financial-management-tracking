export interface UserProfile {
  id: string;
  email: string;
  cycle_start_day: number;
  created_at: string;
}

export interface AuthResponseData {
  token: string;
  profile: UserProfile;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
}

const API_BASE_URL = import.meta.env.VITE_API_URL || '';
const TOKEN_KEY = 'fintrack_token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

async function handleResponse<T>(res: Response): Promise<T> {
  const json: ApiResponse<T> = await res.json().catch(() => ({
    success: false,
    message: 'Invalid server response',
  }));

  if (!res.ok || !json.success) {
    throw new Error(json.message || `Request failed with status ${res.status}`);
  }

  return json.data as T;
}

export async function register(
  email: string,
  password: string,
  cycleStartDay: number = 1
): Promise<AuthResponseData> {
  const res = await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email,
      password,
      cycle_start_day: cycleStartDay,
    }),
  });

  const data = await handleResponse<AuthResponseData>(res);
  if (data.token) {
    setToken(data.token);
  }
  return data;
}

export async function login(email: string, password: string): Promise<AuthResponseData> {
  const res = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  const data = await handleResponse<AuthResponseData>(res);
  if (data.token) {
    setToken(data.token);
  }
  return data;
}

export async function logout(): Promise<void> {
  try {
    await fetch(`${API_BASE_URL}/api/v1/auth/logout`, { method: 'POST' });
  } finally {
    clearToken();
  }
}

export async function getMe(): Promise<UserProfile> {
  const token = getToken();
  if (!token) {
    throw new Error('No authentication token found');
  }

  const res = await fetch(`${API_BASE_URL}/api/v1/auth/me`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return handleResponse<UserProfile>(res);
}

export async function updateUserSettings(cycleStartDay: number): Promise<{ cycle_start_day: number }> {
  const token = getToken();
  if (!token) {
    throw new Error('No authentication token found');
  }

  const res = await fetch(`${API_BASE_URL}/api/v1/auth/settings`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ cycle_start_day: cycleStartDay }),
  });

  return handleResponse<{ cycle_start_day: number }>(res);
}
