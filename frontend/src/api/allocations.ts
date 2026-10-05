import { getToken } from './auth';

export interface Allocation {
  id: string;
  user_id: string;
  category_id: string;
  cycle_start: string;
  cycle_end: string;
  allocated_amount: number;
  category_name?: string;
  category_icon?: string;
  category_color?: string;
  created_at: string;
  updated_at: string;
}

export interface AllocationSummary {
  allocations: Allocation[];
  total_allocated: number;
  total_income: number;
  remaining_income: number;
}

export interface CreateAllocationInput {
  category_id: string;
  cycle_start: string;
  cycle_end: string;
  allocated_amount: number;
}

export interface UpdateAllocationInput {
  allocated_amount: number;
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
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json');
  }

  return fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });
}

export async function createAllocation(input: CreateAllocationInput): Promise<Allocation> {
  const res = await authFetch('/api/v1/allocations', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to create allocation');
  }
  return json.data;
}

export async function listAllocations(cycleStart: string, cycleEnd: string): Promise<AllocationSummary> {
  const qs = `?cycle_start=${encodeURIComponent(cycleStart)}&cycle_end=${encodeURIComponent(cycleEnd)}`;
  const res = await authFetch(`/api/v1/allocations${qs}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to list allocations');
  }
  return json.data;
}

export async function getAllocation(id: string): Promise<Allocation> {
  const res = await authFetch(`/api/v1/allocations/${id}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to get allocation');
  }
  return json.data;
}

export async function updateAllocation(id: string, input: UpdateAllocationInput): Promise<Allocation> {
  const res = await authFetch(`/api/v1/allocations/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to update allocation');
  }
  return json.data;
}

export async function deleteAllocation(id: string): Promise<void> {
  const res = await authFetch(`/api/v1/allocations/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const json = await res.json();
    throw new Error(json.message || 'Failed to delete allocation');
  }
}
