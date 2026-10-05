import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as cycleApi from './cycle';
import * as authApi from './auth';

describe('Cycle API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('getCurrentCycle fetches current financial cycle info', async () => {
    const mockCycle = {
      cycle_start_day: 25,
      start_date: '2026-08-25',
      end_date: '2026-09-24',
      total_days: 31,
      day_of_cycle: 6,
      days_remaining: 25,
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockCycle }),
    });

    const result = await cycleApi.getCurrentCycle('2026-08-30');
    expect(result).toEqual(mockCycle);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/cycle/current?date=2026-08-30'),
      expect.any(Object)
    );
  });

  it('getCycleSummary fetches cycle metrics', async () => {
    const mockSummary = {
      cycle: {
        cycle_start_day: 25,
        start_date: '2026-08-25',
        end_date: '2026-09-24',
        total_days: 31,
        day_of_cycle: 6,
        days_remaining: 25,
      },
      total_income: 10000000,
      total_expense: 2000000,
      net_savings: 8000000,
      transaction_count: 3,
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockSummary }),
    });

    const result = await cycleApi.getCycleSummary();
    expect(result).toEqual(mockSummary);
  });
});
