import { getToken } from './auth';

export interface CycleInfo {
  cycle_start_day: number;
  start_date: string; // YYYY-MM-DD
  end_date: string;   // YYYY-MM-DD
  total_days: number;
  day_of_cycle: number;
  days_remaining: number;
}

export interface CycleSummary {
  cycle: CycleInfo;
  total_income: number;
  total_expense: number;
  net_savings: number;
  transaction_count: number;
}

const API_BASE_URL =
  import.meta.env.VITE_API_URL ||
  (typeof window !== 'undefined' && window.location?.origin?.startsWith('http') ? '' : 'http://localhost:8080');

async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const token = getToken();
  if (!token) {
    throw new Error('Not authenticated');
  }

  const headers = new Headers(options.headers || {});
  headers.set('Authorization', `Bearer ${token}`);

  return fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });
}

export async function getCurrentCycle(date?: string): Promise<CycleInfo> {
  const qs = date ? `?date=${encodeURIComponent(date)}` : '';
  const res = await authFetch(`/api/v1/cycle/current${qs}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || json.error || 'Failed to retrieve current financial cycle');
  }
  return json.data;
}

export async function getCycleSummary(date?: string): Promise<CycleSummary> {
  const qs = date ? `?date=${encodeURIComponent(date)}` : '';
  const res = await authFetch(`/api/v1/cycle/summary${qs}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || json.error || 'Failed to retrieve financial cycle summary');
  }
  return json.data;
}
