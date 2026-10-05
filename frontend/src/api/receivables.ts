import { getToken } from './auth';
import { API_BASE_URL } from '../config';

export type ReceivableStatus = 'ACTIVE' | 'PARTIALLY_PAID' | 'PAID' | 'OVERDUE' | 'WRITTEN_OFF';

export interface ReceivablePayment {
  id: string;
  receivable_id: string;
  user_id: string;
  target_account_id: string;
  target_account_name?: string;
  amount: number;
  formatted_amount?: string;
  payment_date: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface Receivable {
  id: string;
  user_id: string;
  source_account_id?: string;
  source_account_name?: string;
  counterparty: string;
  principal: number;
  formatted_principal?: string;
  total_paid: number;
  formatted_total_paid?: string;
  remaining_amount: number;
  formatted_remaining_amount?: string;
  status: ReceivableStatus;
  due_date?: string;
  notes: string;
  created_at: string;
  updated_at: string;
  payments?: ReceivablePayment[];
}

export interface CreateReceivableInput {
  counterparty: string;
  principal: number;
  source_account_id?: string;
  due_date?: string;
  notes?: string;
}

export interface RecordPaymentInput {
  amount: number;
  target_account_id: string;
  payment_date: string;
  notes?: string;
}

export interface UpdateStatusInput {
  status: ReceivableStatus;
  notes?: string;
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

  const res = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });

  return res;
}

export async function listReceivables(status?: ReceivableStatus): Promise<Receivable[]> {
  const query = status ? `?status=${status}` : '';
  const res = await authFetch(`/api/v1/receivables${query}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to list receivables');
  }
  return json.data || [];
}

export async function getReceivable(id: string): Promise<Receivable> {
  const res = await authFetch(`/api/v1/receivables/${id}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to get receivable');
  }
  return json.data;
}

export async function createReceivable(input: CreateReceivableInput): Promise<Receivable> {
  const res = await authFetch('/api/v1/receivables', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to create receivable');
  }
  return json.data;
}

export async function recordPayment(
  receivableId: string,
  input: RecordPaymentInput
): Promise<{ payment: ReceivablePayment; receivable: Receivable }> {
  const res = await authFetch(`/api/v1/receivables/${receivableId}/payments`, {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to record payment');
  }
  return json.data;
}

export async function updateReceivableStatus(id: string, input: UpdateStatusInput): Promise<Receivable> {
  const res = await authFetch(`/api/v1/receivables/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to update receivable status');
  }
  return json.data;
}

export async function deleteReceivable(id: string): Promise<void> {
  const res = await authFetch(`/api/v1/receivables/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok && res.status !== 204) {
    const json = await res.json().catch(() => ({}));
    throw new Error(json.message || 'Failed to delete receivable');
  }
}
