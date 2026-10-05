import { getToken } from './auth';

export type InvestmentType = 'STOCK' | 'MUTUAL_FUND' | 'GOLD' | 'CRYPTO' | 'OTHER';

export interface Investment {
  id: string;
  user_id: string;
  source_account_id?: string;
  source_account_name?: string;
  type: InvestmentType;
  name: string;
  capital: number;
  formatted_capital?: string;
  current_value: number;
  formatted_current_value?: string;
  unrealized_gain: number;
  formatted_unrealized_gain?: string;
  unrealized_gain_percentage: number;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface PortfolioSummary {
  total_capital: number;
  formatted_total_capital?: string;
  total_current_value: number;
  formatted_total_current_value?: string;
  total_gain_loss: number;
  formatted_total_gain_loss?: string;
  total_gain_loss_percentage: number;
  investments_count: number;
  investments: Investment[];
}

export interface CreateInvestmentInput {
  type: InvestmentType;
  name: string;
  capital: number;
  current_value?: number;
  source_account_id?: string;
  notes?: string;
}

export interface UpdateValuationInput {
  current_value: number;
  notes?: string;
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
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const res = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });

  return res;
}

export async function listInvestments(type?: InvestmentType): Promise<PortfolioSummary> {
  const query = type ? `?type=${type}` : '';
  const res = await authFetch(`/api/v1/investments${query}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to list investments');
  }
  return json.data || {
    total_capital: 0,
    total_current_value: 0,
    total_gain_loss: 0,
    total_gain_loss_percentage: 0,
    investments_count: 0,
    investments: [],
  };
}

export async function getInvestment(id: string): Promise<Investment> {
  const res = await authFetch(`/api/v1/investments/${id}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to get investment');
  }
  return json.data;
}

export async function createInvestment(input: CreateInvestmentInput): Promise<Investment> {
  const res = await authFetch('/api/v1/investments', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to create investment');
  }
  return json.data;
}

export async function updateValuation(id: string, input: UpdateValuationInput): Promise<Investment> {
  const res = await authFetch(`/api/v1/investments/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to update investment valuation');
  }
  return json.data;
}

export async function deleteInvestment(id: string): Promise<void> {
  const res = await authFetch(`/api/v1/investments/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok && res.status !== 204) {
    const json = await res.json().catch(() => ({}));
    throw new Error(json.message || 'Failed to delete investment');
  }
}
