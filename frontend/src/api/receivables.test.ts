import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as receivablesApi from './receivables';
import * as authApi from './auth';

describe('Receivables API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('listReceivables returns receivable list with query param', async () => {
    const mockData: receivablesApi.Receivable[] = [
      {
        id: 'rec-1',
        user_id: 'u-1',
        counterparty: 'Budi',
        principal: 2000000,
        total_paid: 500000,
        remaining_amount: 1500000,
        status: 'PARTIALLY_PAID',
        notes: 'Personal loan',
        created_at: '2026-10-01T00:00:00Z',
        updated_at: '2026-10-01T00:00:00Z',
      },
    ];

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockData }),
    });

    const res = await receivablesApi.listReceivables('PARTIALLY_PAID');
    expect(res).toEqual(mockData);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/receivables?status=PARTIALLY_PAID'),
      expect.any(Object)
    );
  });

  it('createReceivable sends POST and returns created receivable', async () => {
    const created: receivablesApi.Receivable = {
      id: 'rec-2',
      user_id: 'u-1',
      counterparty: 'Citra',
      principal: 1000000,
      total_paid: 0,
      remaining_amount: 1000000,
      status: 'ACTIVE',
      notes: '',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: created }),
    });

    const res = await receivablesApi.createReceivable({
      counterparty: 'Citra',
      principal: 1000000,
    });
    expect(res).toEqual(created);
  });

  it('recordPayment sends POST to /payments and returns payment result', async () => {
    const payResult = {
      payment: {
        id: 'pay-1',
        receivable_id: 'rec-1',
        user_id: 'u-1',
        target_account_id: 'acc-1',
        amount: 500000,
        payment_date: '2026-10-15',
        notes: 'Installment',
        created_at: '2026-10-15T00:00:00Z',
        updated_at: '2026-10-15T00:00:00Z',
      },
      receivable: {
        id: 'rec-1',
        user_id: 'u-1',
        counterparty: 'Budi',
        principal: 2000000,
        total_paid: 500000,
        remaining_amount: 1500000,
        status: 'PARTIALLY_PAID' as const,
        notes: '',
        created_at: '2026-10-01T00:00:00Z',
        updated_at: '2026-10-15T00:00:00Z',
      },
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: payResult }),
    });

    const res = await receivablesApi.recordPayment('rec-1', {
      amount: 500000,
      target_account_id: 'acc-1',
      payment_date: '2026-10-15',
    });
    expect(res).toEqual(payResult);
  });

  it('updateReceivableStatus sends PATCH request', async () => {
    const updated = {
      id: 'rec-1',
      user_id: 'u-1',
      counterparty: 'Budi',
      principal: 2000000,
      total_paid: 0,
      remaining_amount: 2000000,
      status: 'WRITTEN_OFF' as const,
      notes: 'Bad debt',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-15T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: updated }),
    });

    const res = await receivablesApi.updateReceivableStatus('rec-1', { status: 'WRITTEN_OFF', notes: 'Bad debt' });
    expect(res).toEqual(updated);
  });

  it('deleteReceivable sends DELETE request', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true }),
    });

    await expect(receivablesApi.deleteReceivable('rec-1')).resolves.toBeUndefined();
  });
});
