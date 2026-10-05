import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { fetchMonthlyReport, fetchCycleReport, type ReportData } from './reports';
import * as authApi from './auth';

describe('Reports API Client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  const mockReport: ReportData = {
    period: '2026-09',
    start_date: '2026-09-01',
    end_date: '2026-09-30',
    income: 13000000,
    expense: 4300000,
    transfer: 1200000,
    net_cash_flow: 8700000,
    budget_vs_actual: [
      { category: 'Operational', budget: 4000000, actual: 4300000, variance: -300000 },
    ],
    assets: { accounts: 9000000, receivables: 2000000, investments: 2250000, total: 13250000 },
    liabilities: 0,
    net_worth: 13250000,
  };

  it('fetchMonthlyReport requests correct endpoint and returns data', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockReport }),
    });

    const res = await fetchMonthlyReport(2026, 9);
    expect(res.period).toBe('2026-09');
    expect(res.net_worth).toBe(13250000);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/reports/monthly?year=2026&month=09'),
      expect.any(Object)
    );
  });

  it('fetchCycleReport requests correct endpoint with date', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockReport }),
    });

    const res = await fetchCycleReport('2026-09-10');
    expect(res.net_cash_flow).toBe(8700000);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/reports/cycle?date=2026-09-10'),
      expect.any(Object)
    );
  });

  it('fetchCycleReport without date requests base cycle endpoint', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockReport }),
    });

    await fetchCycleReport();
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/reports/cycle'),
      expect.any(Object)
    );
  });

  it('throws error when API response is not ok', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      json: async () => ({ success: false, message: 'Invalid month parameter' }),
    });

    await expect(fetchMonthlyReport(2026, 13)).rejects.toThrow('Invalid month parameter');
  });
});
