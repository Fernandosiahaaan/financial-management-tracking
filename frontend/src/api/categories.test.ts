import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as categoriesApi from './categories';
import * as authApi from './auth';

describe('Categories API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('listCategories returns categories list', async () => {
    const mockCategories = [
      {
        id: 'cat-1',
        user_id: 'user-1',
        name: 'Dining',
        type: 'EXPENSE',
        icon: 'utensils',
        color: '#EF4444',
        created_at: '2026-10-04T00:00:00Z',
        updated_at: '2026-10-04T00:00:00Z',
      },
    ];

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockCategories }),
    });

    const result = await categoriesApi.listCategories();
    expect(result).toEqual(mockCategories);
  });

  it('createCategory sends POST and returns created category', async () => {
    const newCat = {
      id: 'cat-2',
      user_id: 'user-1',
      name: 'Salary',
      type: 'INCOME' as const,
      icon: 'briefcase',
      color: '#10B981',
      created_at: '2026-10-04T00:00:00Z',
      updated_at: '2026-10-04T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: newCat }),
    });

    const result = await categoriesApi.createCategory({
      name: 'Salary',
      type: 'INCOME',
      icon: 'briefcase',
      color: '#10B981',
    });
    expect(result).toEqual(newCat);
  });

  it('deleteCategory sends DELETE request', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    });

    await expect(categoriesApi.deleteCategory('cat-1')).resolves.toBeUndefined();
  });
});
