import { getToken } from './auth';

export type AccountType = 'BANK' | 'E_WALLET' | 'CASH' | 'INVESTMENT' | 'OTHER';
export type AccountStatus = 'ACTIVE' | 'ARCHIVED';

export interface Account {
  id: string;
  user_id: string;
  name: string;
  type: AccountType;
  opening_balance: number;
  current_balance: number;
  formatted_balance?: string;
  status: AccountStatus;
  created_at: string;
  updated_at: string;
}

export interface CreateAccountInput {
  name: string;
  type: AccountType;
  opening_balance: number;
}

export interface UpdateAccountInput {
  name: string;
  type: AccountType;
  status: AccountStatus;
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

export async function listAccounts(): Promise<Account[]> {
  const res = await authFetch('/api/v1/accounts');
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to list accounts');
  }
  return json.data || [];
}

export async function createAccount(input: CreateAccountInput): Promise<Account> {
  const res = await authFetch('/api/v1/accounts', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to create account');
  }
  return json.data;
}

export async function updateAccount(id: string, input: UpdateAccountInput): Promise<Account> {
  const res = await authFetch(`/api/v1/accounts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to update account');
  }
  return json.data;
}

export async function deleteAccount(id: string): Promise<void> {
  const res = await authFetch(`/api/v1/accounts/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok && res.status !== 204) {
    const json = await res.json().catch(() => ({}));
    throw new Error(json.message || 'Failed to delete account');
  }
}
