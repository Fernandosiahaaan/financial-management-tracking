import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { InvestmentsManager } from './InvestmentsManager';
import * as investmentsApi from '../api/investments';
import { type Account } from '../api/accounts';

describe('InvestmentsManager Component', () => {
  const mockAccounts: Account[] = [
    {
      id: 'acc-1',
      user_id: 'u-1',
      name: 'Stock Account',
      type: 'INVESTMENT',
      opening_balance: 5000000000,
      current_balance: 5000000000,
      status: 'ACTIVE',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    },
  ];

  const mockPortfolio: investmentsApi.PortfolioSummary = {
    total_capital: 2000000000,
    total_current_value: 2400000000,
    total_gain_loss: 400000000,
    total_gain_loss_percentage: 20,
    investments_count: 1,
    investments: [
      {
        id: 'inv-1',
        user_id: 'u-1',
        type: 'STOCK',
        name: 'BBCA - Bank Central Asia',
        capital: 2000000000,
        current_value: 2400000000,
        unrealized_gain: 400000000,
        unrealized_gain_percentage: 20,
        notes: 'Blue chip',
        created_at: '2026-10-01T00:00:00Z',
        updated_at: '2026-10-01T00:00:00Z',
      },
    ],
  };

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(investmentsApi, 'listInvestments').mockResolvedValue(mockPortfolio);
  });

  it('renders portfolio summary and holding cards', async () => {
    render(<InvestmentsManager accounts={mockAccounts} />);

    await waitFor(() => {
      expect(screen.getByText('BBCA - Bank Central Asia')).toBeInTheDocument();
    });

    expect(screen.getByText('Total Current Valuation')).toBeInTheDocument();
    expect(screen.getByText('Invested Capital')).toBeInTheDocument();
    expect(screen.getAllByText(/Rp\s*24\.000\.000,00/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Rp\s*20\.000\.000,00/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Rp\s*4\.000\.000,00/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/20.00%/).length).toBeGreaterThan(0);
  });

  it('opens holding creation form and adds holding', async () => {
    const createSpy = vi.spyOn(investmentsApi, 'createInvestment').mockResolvedValue({
      id: 'inv-2',
      user_id: 'u-1',
      type: 'GOLD',
      name: 'Antam 10g',
      capital: 1200000000,
      current_value: 1200000000,
      unrealized_gain: 0,
      unrealized_gain_percentage: 0,
      notes: '',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    });

    render(<InvestmentsManager accounts={mockAccounts} />);

    fireEvent.click(screen.getByText('+ Add Investment Holding'));
    expect(screen.getByText('Add Investment Holding')).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/Holding Name/i), {
      target: { value: 'Antam 10g' },
    });
    fireEvent.change(screen.getByLabelText(/Capital Invested/i), {
      target: { value: '12000000' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Add Investment' }));

    await waitFor(() => {
      expect(createSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Antam 10g',
          capital: 1200000000,
        })
      );
    });
  });

  it('opens valuation modal and updates valuation', async () => {
    const updateSpy = vi.spyOn(investmentsApi, 'updateValuation').mockResolvedValue({
      ...mockPortfolio.investments[0],
      current_value: 2600000000,
      unrealized_gain: 600000000,
      unrealized_gain_percentage: 30,
    });

    render(<InvestmentsManager accounts={mockAccounts} />);

    await waitFor(() => {
      expect(screen.getByText('Update Valuation')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Update Valuation'));

    expect(screen.getByText(/Update Valuation for BBCA/i)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/Latest Market Value/i), {
      target: { value: '26000000' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Save Valuation' }));

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledWith(
        'inv-1',
        expect.objectContaining({
          current_value: 2600000000,
        })
      );
    });
  });
});
