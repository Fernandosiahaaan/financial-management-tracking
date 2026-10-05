import { getToken } from './auth';
import { API_BASE_URL } from '../config';

export interface Budget {
  id: string;
  user_id: string;
  category_id: string;
  cycle_start: string;
  cycle_end: string;
  planned_amount: number;
  actual_amount: number;
  variance: number;
  category_name?: string;
  category_icon?: string;
  category_color?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateBudgetInput {
  category_id: string;
  cycle_start: string;
  cycle_end: string;
  planned_amount: number;
}

export interface UpdateBudgetInput {
  planned_amount: number;
}

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

export async function createBudget(input: CreateBudgetInput): Promise<Budget> {
  const res = await authFetch('/api/v1/budgets', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to create budget');
  }
  return json.data;
}

export async function listBudgets(cycleStart: string, cycleEnd: string): Promise<Budget[]> {
  const qs = `?cycle_start=${encodeURIComponent(cycleStart)}&cycle_end=${encodeURIComponent(cycleEnd)}`;
  const res = await authFetch(`/api/v1/budgets${qs}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to list budgets');
  }
  return json.data;
}

export async function getBudget(id: string): Promise<Budget> {
  const res = await authFetch(`/api/v1/budgets/${id}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to get budget');
  }
  return json.data;
}

export async function updateBudget(id: string, input: UpdateBudgetInput): Promise<Budget> {
  const res = await authFetch(`/api/v1/budgets/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to update budget');
  }
  return json.data;
}

export async function deleteBudget(id: string): Promise<void> {
  const res = await authFetch(`/api/v1/budgets/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const json = await res.json();
    throw new Error(json.message || 'Failed to delete budget');
  }
}
