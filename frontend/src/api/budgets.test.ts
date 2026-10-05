import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as budgetsApi from './budgets';
import * as authApi from './auth';

describe('Budgets API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('listBudgets returns budget list with query params', async () => {
    const mockBudgets: budgetsApi.Budget[] = [
      {
        id: 'budget-1',
        user_id: 'user-1',
        category_id: 'cat-1',
        cycle_start: '2026-10-01',
        cycle_end: '2026-10-31',
        planned_amount: 3000000,
        actual_amount: 1500000,
        variance: 1500000,
        category_name: 'Food & Dining',
        category_icon: 'utensils',
        category_color: '#EF4444',
        created_at: '2026-10-01T00:00:00Z',
        updated_at: '2026-10-01T00:00:00Z',
      },
    ];

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockBudgets }),
    });

    const result = await budgetsApi.listBudgets('2026-10-01', '2026-10-31');
    expect(result).toEqual(mockBudgets);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/budgets?cycle_start=2026-10-01&cycle_end=2026-10-31'),
      expect.any(Object)
    );
  });

  it('createBudget sends POST and returns created budget', async () => {
    const newBudget: budgetsApi.Budget = {
      id: 'budget-2',
      user_id: 'user-1',
      category_id: 'cat-2',
      cycle_start: '2026-10-01',
      cycle_end: '2026-10-31',
      planned_amount: 2000000,
      actual_amount: 0,
      variance: 2000000,
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: newBudget }),
    });

    const result = await budgetsApi.createBudget({
      category_id: 'cat-2',
      cycle_start: '2026-10-01',
      cycle_end: '2026-10-31',
      planned_amount: 2000000,
    });
    expect(result).toEqual(newBudget);
  });

  it('updateBudget sends PUT and returns updated budget', async () => {
    const updated: budgetsApi.Budget = {
      id: 'budget-1',
      user_id: 'user-1',
      category_id: 'cat-1',
      cycle_start: '2026-10-01',
      cycle_end: '2026-10-31',
      planned_amount: 4000000,
      actual_amount: 1500000,
      variance: 2500000,
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: updated }),
    });

    const result = await budgetsApi.updateBudget('budget-1', { planned_amount: 4000000 });
    expect(result).toEqual(updated);
  });

  it('deleteBudget sends DELETE request', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    });

    await expect(budgetsApi.deleteBudget('budget-1')).resolves.toBeUndefined();
  });
});
