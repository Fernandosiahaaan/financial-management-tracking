import { getToken } from './auth';

export type CategoryType = 'INCOME' | 'EXPENSE';

export interface Category {
  id: string;
  user_id: string;
  name: string;
  type: CategoryType;
  icon: string;
  color: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCategoryInput {
  name: string;
  type: CategoryType;
  icon?: string;
  color?: string;
}

export interface UpdateCategoryInput {
  name: string;
  type: CategoryType;
  icon?: string;
  color?: string;
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

export async function listCategories(): Promise<Category[]> {
  const res = await authFetch('/api/v1/categories');
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to list categories');
  }
  return json.data || [];
}

export async function createCategory(input: CreateCategoryInput): Promise<Category> {
  const res = await authFetch('/api/v1/categories', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to create category');
  }
  return json.data;
}

export async function updateCategory(id: string, input: UpdateCategoryInput): Promise<Category> {
  const res = await authFetch(`/api/v1/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to update category');
  }
  return json.data;
}

export async function deleteCategory(id: string): Promise<void> {
  const res = await authFetch(`/api/v1/categories/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok && res.status !== 204) {
    const json = await res.json().catch(() => ({}));
    throw new Error(json.message || 'Failed to delete category');
  }
}
