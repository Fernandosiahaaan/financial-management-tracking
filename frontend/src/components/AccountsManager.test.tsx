import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { AccountsManager } from './AccountsManager';
import * as accountsApi from '../api/accounts';

describe('AccountsManager Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders accounts list and total balance', async () => {
    vi.spyOn(accountsApi, 'listAccounts').mockResolvedValue([
      {
        id: 'acc-1',
        user_id: 'user-1',
        name: 'BCA Checking',
        type: 'BANK',
        opening_balance: 500000000,
        current_balance: 500000000,
        status: 'ACTIVE',
        created_at: '2026-10-04T00:00:00Z',
        updated_at: '2026-10-04T00:00:00Z',
      },
      {
        id: 'acc-2',
        user_id: 'user-1',
        name: 'GoPay Wallet',
        type: 'E_WALLET',
        opening_balance: 25000000,
        current_balance: 25000000,
        status: 'ACTIVE',
        created_at: '2026-10-04T00:00:00Z',
        updated_at: '2026-10-04T00:00:00Z',
      },
    ]);

    render(<AccountsManager />);

    await waitFor(() => {
      expect(screen.getByText('BCA Checking')).toBeInTheDocument();
      expect(screen.getByText('GoPay Wallet')).toBeInTheDocument();
      expect(screen.getByText(/Rp\s*5\.250\.000,00/)).toBeInTheDocument();
    });
  });

  it('validates and creates a new account', async () => {
    vi.spyOn(accountsApi, 'listAccounts').mockResolvedValue([]);
    const createSpy = vi.spyOn(accountsApi, 'createAccount').mockResolvedValue({
      id: 'acc-new',
      user_id: 'user-1',
      name: 'Cash Box',
      type: 'CASH',
      opening_balance: 10000000,
      current_balance: 10000000,
      status: 'ACTIVE',
      created_at: '2026-10-04T00:00:00Z',
      updated_at: '2026-10-04T00:00:00Z',
    });

    render(<AccountsManager />);

    await waitFor(() => {
      expect(screen.getByText(/No Accounts Configured Yet/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /\+ Add Account/i }));

    const nameInput = screen.getByLabelText(/Account Name/i);
    const balanceInput = screen.getByLabelText(/Opening Balance/i);
    const submitBtn = screen.getByRole('button', { name: /Create Account/i });

    // Missing name validation
    fireEvent.click(submitBtn);
    expect(screen.getByText(/Account name is required/i)).toBeInTheDocument();

    // Fill valid data
    fireEvent.change(nameInput, { target: { value: 'Cash Box' } });
    fireEvent.change(balanceInput, { target: { value: '100000' } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(createSpy).toHaveBeenCalledWith({
        name: 'Cash Box',
        type: 'BANK',
        opening_balance: 10000000,
      });
    });
  });
});
