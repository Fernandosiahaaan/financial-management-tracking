import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as allocationsApi from './allocations';
import * as authApi from './auth';

describe('Allocations API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('listAllocations returns allocation summary with query params', async () => {
    const mockSummary: allocationsApi.AllocationSummary = {
      allocations: [
        {
          id: 'alloc-1',
          user_id: 'user-1',
          category_id: 'cat-1',
          cycle_start: '2026-10-01',
          cycle_end: '2026-10-31',
          allocated_amount: 3000000,
          category_name: 'Living Expenses',
          created_at: '2026-10-01T00:00:00Z',
          updated_at: '2026-10-01T00:00:00Z',
        },
      ],
      total_allocated: 3000000,
      total_income: 10000000,
      remaining_income: 7000000,
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockSummary }),
    });

    const result = await allocationsApi.listAllocations('2026-10-01', '2026-10-31');
    expect(result).toEqual(mockSummary);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/allocations?cycle_start=2026-10-01&cycle_end=2026-10-31'),
      expect.any(Object)
    );
  });

  it('createAllocation sends POST and returns created allocation', async () => {
    const newAlloc: allocationsApi.Allocation = {
      id: 'alloc-2',
      user_id: 'user-1',
      category_id: 'cat-2',
      cycle_start: '2026-10-01',
      cycle_end: '2026-10-31',
      allocated_amount: 2000000,
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: newAlloc }),
    });

    const result = await allocationsApi.createAllocation({
      category_id: 'cat-2',
      cycle_start: '2026-10-01',
      cycle_end: '2026-10-31',
      allocated_amount: 2000000,
    });
    expect(result).toEqual(newAlloc);
  });

  it('updateAllocation sends PUT and returns updated allocation', async () => {
    const updated: allocationsApi.Allocation = {
      id: 'alloc-1',
      user_id: 'user-1',
      category_id: 'cat-1',
      cycle_start: '2026-10-01',
      cycle_end: '2026-10-31',
      allocated_amount: 4500000,
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: updated }),
    });

    const result = await allocationsApi.updateAllocation('alloc-1', { allocated_amount: 4500000 });
    expect(result).toEqual(updated);
  });

  it('deleteAllocation sends DELETE request', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    });

    await expect(allocationsApi.deleteAllocation('alloc-1')).resolves.toBeUndefined();
  });
});
