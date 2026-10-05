import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as dashboardApi from './dashboard';
import * as authApi from './auth';

describe('Dashboard API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('getDashboardData fetches and returns dashboard metrics', async () => {
    const mockData: dashboardApi.DashboardData = {
      cycle: {
        start_date: '2026-10-01',
        end_date: '2026-10-31',
        cycle_start_day: 1,
        day_of_cycle: 5,
        total_days: 31,
        days_remaining: 26,
      },
      income: 15000000,
      expense: 4000000,
      net_cash_flow: 11000000,
      total_assets: 50000000,
      previous_cycle_assets: 39000000,
      asset_growth: 11000000,
      budgets: [
        {
          category_id: 'cat-1',
          category_name: 'Groceries',
          planned: 3000000,
          actual: 2500000,
          variance: 500000,
          overspent: false,
        },
      ],
      accounts: [
        {
          id: 'acc-1',
          name: 'Main Bank',
          type: 'BANK',
          balance: 50000000,
        },
      ],
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockData }),
    });

    const result = await dashboardApi.getDashboardData('2026-10-05');
    expect(result).toEqual(mockData);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/dashboard?date=2026-10-05'),
      expect.any(Object)
    );
  });
});
