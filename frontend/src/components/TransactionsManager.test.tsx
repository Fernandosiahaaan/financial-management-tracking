import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { TransactionsManager } from './TransactionsManager';
import * as transactionsApi from '../api/transactions';
import * as accountsApi from '../api/accounts';
import * as categoriesApi from '../api/categories';

describe('TransactionsManager Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  const mockAccounts = [
    {
      id: 'acc-1',
      user_id: 'user-1',
      name: 'BCA Checking',
      type: 'BANK' as const,
      opening_balance: 1000000,
      current_balance: 1000000,
      status: 'ACTIVE' as const,
      created_at: '2026-10-04T00:00:00Z',
      updated_at: '2026-10-04T00:00:00Z',
    },
    {
      id: 'acc-2',
      user_id: 'user-1',
      name: 'Cash Wallet',
      type: 'CASH' as const,
      opening_balance: 500000,
      current_balance: 500000,
      status: 'ACTIVE' as const,
      created_at: '2026-10-04T00:00:00Z',
      updated_at: '2026-10-04T00:00:00Z',
    },
  ];

  const mockCategories = [
    {
      id: 'cat-1',
      user_id: 'user-1',
      name: 'Food & Dining',
      type: 'EXPENSE' as const,
      icon: '🍔',
      color: '#ef4444',
      created_at: '2026-10-04T00:00:00Z',
      updated_at: '2026-10-04T00:00:00Z',
    },
  ];

  it('renders transactions list and summary metrics', async () => {
    vi.spyOn(accountsApi, 'listAccounts').mockResolvedValue(mockAccounts);
    vi.spyOn(categoriesApi, 'listCategories').mockResolvedValue(mockCategories);
    vi.spyOn(transactionsApi, 'listTransactions').mockResolvedValue([
      {
        id: 'tx-1',
        user_id: 'user-1',
        account_id: 'acc-1',
        category_id: 'cat-1',
        type: 'EXPENSE',
        amount: 50000,
        transaction_date: '2026-10-05',
        description: 'Lunch at cafe',
        account_name: 'BCA Checking',
        category_name: 'Food & Dining',
        created_at: '2026-10-05T00:00:00Z',
        updated_at: '2026-10-05T00:00:00Z',
      },
    ]);

    render(<TransactionsManager />);

    await waitFor(() => {
      expect(screen.getByText('Lunch at cafe')).toBeInTheDocument();
      expect(screen.getAllByText('BCA Checking').length).toBeGreaterThan(0);
      expect(screen.getByText('Food & Dining')).toBeInTheDocument();
      expect(screen.getByText('Total Expense')).toBeInTheDocument();
    });
  });

  it('validates and records a new transfer transaction', async () => {
    vi.spyOn(accountsApi, 'listAccounts').mockResolvedValue(mockAccounts);
    vi.spyOn(categoriesApi, 'listCategories').mockResolvedValue(mockCategories);
    vi.spyOn(transactionsApi, 'listTransactions').mockResolvedValue([]);
    const createSpy = vi.spyOn(transactionsApi, 'createTransaction').mockResolvedValue({
      id: 'tx-transfer',
      user_id: 'user-1',
      account_id: 'acc-1',
      destination_account_id: 'acc-2',
      type: 'TRANSFER',
      amount: 100000,
      transaction_date: '2026-10-05',
      description: 'ATM withdrawal',
      created_at: '2026-10-05T00:00:00Z',
      updated_at: '2026-10-05T00:00:00Z',
    });

    render(<TransactionsManager />);

    await waitFor(() => {
      expect(screen.getByText(/No transactions recorded/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /\+ Record Transaction/i }));

    // Switch to Transfer tab
    fireEvent.click(screen.getByRole('button', { name: /^Transfer$/i }));

    const amountInput = screen.getByLabelText(/Amount \(IDR\)/i);
    fireEvent.change(amountInput, { target: { value: '100000' } });

    const submitBtn = screen.getByRole('button', { name: /Save Transaction/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(createSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'TRANSFER',
          amount: 10000000,
          account_id: 'acc-1',
          destination_account_id: 'acc-2',
        })
      );
    });
  });
});
