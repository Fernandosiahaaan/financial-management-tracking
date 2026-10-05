import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ReceivablesManager } from './ReceivablesManager';
import * as receivablesApi from '../api/receivables';
import { type Account } from '../api/accounts';

describe('ReceivablesManager Component', () => {
  const mockAccounts: Account[] = [
    {
      id: 'acc-1',
      user_id: 'u-1',
      name: 'BCA Checking',
      type: 'BANK',
      opening_balance: 1000000000,
      current_balance: 1000000000,
      status: 'ACTIVE',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    },
  ];

  const mockReceivables: receivablesApi.Receivable[] = [
    {
      id: 'rec-1',
      user_id: 'u-1',
      counterparty: 'Budi Hartono',
      principal: 200000000,
      total_paid: 50000000,
      remaining_amount: 150000000,
      status: 'PARTIALLY_PAID',
      notes: 'Emergency loan',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
      payments: [
        {
          id: 'pay-1',
          receivable_id: 'rec-1',
          user_id: 'u-1',
          target_account_id: 'acc-1',
          target_account_name: 'BCA Checking',
          amount: 50000000,
          payment_date: '2026-10-05',
          notes: 'First installment',
          created_at: '2026-10-05T00:00:00Z',
          updated_at: '2026-10-05T00:00:00Z',
        },
      ],
    },
  ];

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(receivablesApi, 'listReceivables').mockResolvedValue(mockReceivables);
  });

  it('renders KPI summary and receivable item', async () => {
    render(<ReceivablesManager accounts={mockAccounts} />);

    await waitFor(() => {
      expect(screen.getByText('Budi Hartono')).toBeInTheDocument();
    });

    expect(screen.getByText('Active Outstanding')).toBeInTheDocument();
    expect(screen.getByText('Partially Paid')).toBeInTheDocument();
    expect(screen.getAllByText(/Rp\s*2\.000\.000,00/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Rp\s*1\.500\.000,00/).length).toBeGreaterThan(0);
  });

  it('opens creation form and creates a new receivable', async () => {
    const createSpy = vi.spyOn(receivablesApi, 'createReceivable').mockResolvedValue({
      id: 'rec-2',
      user_id: 'u-1',
      counterparty: 'Citra',
      principal: 300000000,
      total_paid: 0,
      remaining_amount: 300000000,
      status: 'ACTIVE',
      notes: '',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    });

    render(<ReceivablesManager accounts={mockAccounts} />);

    const openBtn = screen.getByText('+ Lend Money / Add Receivable');
    fireEvent.click(openBtn);

    expect(screen.getByText('Record New Loan / Receivable')).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/Counterparty \/ Borrower/i), {
      target: { value: 'Citra' },
    });
    fireEvent.change(screen.getByLabelText(/Principal Amount/i), {
      target: { value: '3000000' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Save Receivable' }));

    await waitFor(() => {
      expect(createSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          counterparty: 'Citra',
          principal: 300000000,
        })
      );
    });
  });

  it('opens repayment modal and records repayment', async () => {
    const paySpy = vi.spyOn(receivablesApi, 'recordPayment').mockResolvedValue({
      payment: {
        id: 'pay-2',
        receivable_id: 'rec-1',
        user_id: 'u-1',
        target_account_id: 'acc-1',
        amount: 50000000,
        payment_date: '2026-10-15',
        notes: '',
        created_at: '2026-10-15T00:00:00Z',
        updated_at: '2026-10-15T00:00:00Z',
      },
      receivable: {
        ...mockReceivables[0],
        total_paid: 100000000,
        remaining_amount: 100000000,
      },
    });

    render(<ReceivablesManager accounts={mockAccounts} />);

    await waitFor(() => {
      expect(screen.getByText('Record Repayment')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Record Repayment'));

    expect(screen.getByText(/Record Repayment from Budi Hartono/i)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/Amount Repaid/i), {
      target: { value: '500000' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Confirm Repayment' }));

    await waitFor(() => {
      expect(paySpy).toHaveBeenCalledWith(
        'rec-1',
        expect.objectContaining({
          amount: 50000000,
          target_account_id: 'acc-1',
        })
      );
    });
  });
});
