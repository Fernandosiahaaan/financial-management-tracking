import { getToken } from './auth';
import { API_BASE_URL } from '../config';

export type TransactionType = 'INCOME' | 'EXPENSE' | 'TRANSFER';

export interface Transaction {
  id: string;
  user_id: string;
  account_id: string;
  destination_account_id?: string;
  category_id?: string;
  type: TransactionType;
  amount: number;
  formatted_amount?: string;
  transaction_date: string; // YYYY-MM-DD
  description?: string;
  created_at: string;
  updated_at: string;

  // Joined display attributes
  account_name?: string;
  destination_account_name?: string;
  category_name?: string;
  category_color?: string;
  category_icon?: string;
}

export interface CreateTransactionInput {
  account_id: string;
  destination_account_id?: string;
  category_id?: string;
  type: TransactionType;
  amount: number;
  transaction_date: string;
  description?: string;
}

export interface UpdateTransactionInput {
  account_id: string;
  destination_account_id?: string;
  category_id?: string;
  type: TransactionType;
  amount: number;
  transaction_date: string;
  description?: string;
}

export interface TransactionFilter {
  start_date?: string;
  end_date?: string;
  type?: TransactionType;
  account_id?: string;
  limit?: number;
  offset?: number;
}

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

  return fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });
}

export async function listTransactions(filter?: TransactionFilter): Promise<Transaction[]> {
  const params = new URLSearchParams();
  if (filter?.start_date) params.append('start_date', filter.start_date);
  if (filter?.end_date) params.append('end_date', filter.end_date);
  if (filter?.type) params.append('type', filter.type);
  if (filter?.account_id) params.append('account_id', filter.account_id);
  if (filter?.limit) params.append('limit', filter.limit.toString());
  if (filter?.offset) params.append('offset', filter.offset.toString());

  const qs = params.toString() ? `?${params.toString()}` : '';
  const res = await authFetch(`/api/v1/transactions${qs}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || json.error || 'Failed to list transactions');
  }
  return json.data || [];
}

export async function getTransaction(id: string): Promise<Transaction> {
  const res = await authFetch(`/api/v1/transactions/${id}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || json.error || 'Failed to get transaction');
  }
  return json.data;
}

export async function createTransaction(input: CreateTransactionInput): Promise<Transaction> {
  const res = await authFetch('/api/v1/transactions', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || json.error || 'Failed to create transaction');
  }
  return json.data;
}

export async function updateTransaction(id: string, input: UpdateTransactionInput): Promise<Transaction> {
  const res = await authFetch(`/api/v1/transactions/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || json.error || 'Failed to update transaction');
  }
  return json.data;
}

export async function deleteTransaction(id: string): Promise<void> {
  const res = await authFetch(`/api/v1/transactions/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok && res.status !== 204) {
    const json = await res.json().catch(() => ({}));
    throw new Error(json.message || json.error || 'Failed to delete transaction');
  }
}
