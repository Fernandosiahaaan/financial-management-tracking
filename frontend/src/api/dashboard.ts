import { getToken } from './auth';

export interface DashboardCycle {
  start_date: string;
  end_date: string;
  cycle_start_day: number;
  day_of_cycle: number;
  total_days: number;
  days_remaining: number;
}

export interface BudgetProgress {
  category_id: string;
  category_name: string;
  category_icon?: string;
  category_color?: string;
  planned: number;
  actual: number;
  variance: number;
  overspent: boolean;
}

export interface AccountSummary {
  id: string;
  name: string;
  type: string;
  balance: number;
}

export interface DashboardData {
  cycle: DashboardCycle;
  income: number;
  expense: number;
  net_cash_flow: number;
  total_assets: number;
  previous_cycle_assets: number;
  asset_growth: number;
  budgets: BudgetProgress[];
  accounts: AccountSummary[];
}

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const token = getToken();
  if (!token) {
    throw new Error('Not authenticated');
  }

  const headers = new Headers(options.headers || {});
  headers.set('Authorization', `Bearer ${token}`);
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json');
  }

  return fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });
}

/**
 * Fetch unified dashboard data for the given date (defaults to current date if omitted).
 */
export async function getDashboardData(date?: string): Promise<DashboardData> {
  const qs = date ? `?date=${encodeURIComponent(date)}` : '';
  const res = await authFetch(`/api/v1/dashboard${qs}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to load dashboard data');
  }
  return json.data;
}
