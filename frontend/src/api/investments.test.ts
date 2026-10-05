import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as investmentsApi from './investments';
import * as authApi from './auth';

describe('Investments API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('listInvestments returns portfolio summary', async () => {
    const mockSummary: investmentsApi.PortfolioSummary = {
      total_capital: 10000000,
      total_current_value: 12000000,
      total_gain_loss: 2000000,
      total_gain_loss_percentage: 20,
      investments_count: 1,
      investments: [
        {
          id: 'inv-1',
          user_id: 'u-1',
          type: 'STOCK',
          name: 'BBCA',
          capital: 10000000,
          current_value: 12000000,
          unrealized_gain: 2000000,
          unrealized_gain_percentage: 20,
          notes: '',
          created_at: '2026-10-01T00:00:00Z',
          updated_at: '2026-10-01T00:00:00Z',
        },
      ],
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockSummary }),
    });

    const res = await investmentsApi.listInvestments('STOCK');
    expect(res).toEqual(mockSummary);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/investments?type=STOCK'),
      expect.any(Object)
    );
  });

  it('createInvestment sends POST and returns created investment', async () => {
    const created: investmentsApi.Investment = {
      id: 'inv-2',
      user_id: 'u-1',
      type: 'GOLD',
      name: 'Antam 10g',
      capital: 10000000,
      current_value: 10000000,
      unrealized_gain: 0,
      unrealized_gain_percentage: 0,
      notes: '',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: created }),
    });

    const res = await investmentsApi.createInvestment({
      type: 'GOLD',
      name: 'Antam 10g',
      capital: 10000000,
    });
    expect(res).toEqual(created);
  });

  it('updateValuation sends PUT with new value', async () => {
    const updated: investmentsApi.Investment = {
      id: 'inv-2',
      user_id: 'u-1',
      type: 'GOLD',
      name: 'Antam 10g',
      capital: 10000000,
      current_value: 11000000,
      unrealized_gain: 1000000,
      unrealized_gain_percentage: 10,
      notes: 'Market gain',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-15T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: updated }),
    });

    const res = await investmentsApi.updateValuation('inv-2', {
      current_value: 11000000,
      notes: 'Market gain',
    });
    expect(res).toEqual(updated);
  });

  it('deleteInvestment sends DELETE request', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true }),
    });

    await expect(investmentsApi.deleteInvestment('inv-2')).resolves.toBeUndefined();
  });
});
