import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as accountsApi from './accounts';
import * as authApi from './auth';

describe('Accounts API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(authApi, 'getToken').mockReturnValue('mock-jwt-token');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('listAccounts returns accounts list', async () => {
    const mockAccounts = [
      {
        id: 'acc-1',
        user_id: 'user-1',
        name: 'BCA Savings',
        type: 'BANK',
        opening_balance: 1000000,
        current_balance: 1000000,
        status: 'ACTIVE',
        created_at: '2026-10-04T00:00:00Z',
        updated_at: '2026-10-04T00:00:00Z',
      },
    ];

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockAccounts }),
    });

    const result = await accountsApi.listAccounts();
    expect(result).toEqual(mockAccounts);
  });

  it('createAccount sends POST and returns created account', async () => {
    const newAcc = {
      id: 'acc-2',
      user_id: 'user-1',
      name: 'GoPay',
      type: 'E_WALLET' as const,
      opening_balance: 500000,
      current_balance: 500000,
      status: 'ACTIVE' as const,
      created_at: '2026-10-04T00:00:00Z',
      updated_at: '2026-10-04T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: newAcc }),
    });

    const result = await accountsApi.createAccount({
      name: 'GoPay',
      type: 'E_WALLET',
      opening_balance: 500000,
    });
    expect(result).toEqual(newAcc);
  });

  it('deleteAccount sends DELETE request', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    });

    await expect(accountsApi.deleteAccount('acc-1')).resolves.toBeUndefined();
  });
});
