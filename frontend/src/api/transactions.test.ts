import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as transactionsApi from './transactions';
import * as authApi from './auth';

describe('Transactions API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('listTransactions returns transaction array', async () => {
    const mockTxs = [
      {
        id: 'tx-1',
        user_id: 'user-1',
        account_id: 'acc-1',
        type: 'EXPENSE' as const,
        amount: 50000,
        transaction_date: '2026-10-05',
        description: 'Coffee',
        created_at: '2026-10-05T00:00:00Z',
        updated_at: '2026-10-05T00:00:00Z',
      },
    ];

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockTxs }),
    });

    const result = await transactionsApi.listTransactions();
    expect(result).toEqual(mockTxs);
  });

  it('createTransaction posts data and returns result', async () => {
    const newTx = {
      id: 'tx-2',
      user_id: 'user-1',
      account_id: 'acc-1',
      type: 'INCOME' as const,
      amount: 5000000,
      transaction_date: '2026-10-01',
      description: 'Monthly Salary',
      created_at: '2026-10-05T00:00:00Z',
      updated_at: '2026-10-05T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: newTx }),
    });

    const result = await transactionsApi.createTransaction({
      account_id: 'acc-1',
      type: 'INCOME',
      amount: 5000000,
      transaction_date: '2026-10-01',
      description: 'Monthly Salary',
    });
    expect(result).toEqual(newTx);
  });

  it('updateTransaction sends PUT and returns updated transaction', async () => {
    const updatedTx = {
      id: 'tx-2',
      user_id: 'user-1',
      account_id: 'acc-1',
      type: 'INCOME' as const,
      amount: 6000000,
      transaction_date: '2026-10-01',
      description: 'Monthly Salary + Bonus',
      created_at: '2026-10-05T00:00:00Z',
      updated_at: '2026-10-05T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: updatedTx }),
    });

    const result = await transactionsApi.updateTransaction('tx-2', {
      account_id: 'acc-1',
      type: 'INCOME',
      amount: 6000000,
      transaction_date: '2026-10-01',
      description: 'Monthly Salary + Bonus',
    });
    expect(result).toEqual(updatedTx);
  });

  it('deleteTransaction sends DELETE and resolves', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    });

    await expect(transactionsApi.deleteTransaction('tx-1')).resolves.toBeUndefined();
  });
});
