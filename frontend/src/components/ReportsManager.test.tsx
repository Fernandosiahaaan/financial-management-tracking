import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ReportsManager } from './ReportsManager';
import * as reportsApi from '../api/reports';

const mockReport: reportsApi.ReportData = {
  period: '2026-09',
  start_date: '2026-09-01',
  end_date: '2026-09-30',
  income: 13000000,
  expense: 4300000,
  transfer: 1200000,
  net_cash_flow: 8700000,
  budget_vs_actual: [
    {
      category: 'Operational',
      budget: 4000000,
      actual: 4300000,
      variance: -300000,
    },
    {
      category: 'Utilities',
      budget: 1500000,
      actual: 1200000,
      variance: 300000,
    },
  ],
  assets: {
    accounts: 9000000,
    receivables: 2000000,
    investments: 2250000,
    total: 13250000,
    account_list: [
      { id: 'acc-1', name: 'Main BCA', type: 'BANK', balance: 9000000 },
    ],
    receivable_list: [
      { id: 'rec-1', counterparty: 'John Doe', principal: 2000000, remaining_amount: 2000000, status: 'ACTIVE' },
    ],
    investment_list: [
      { id: 'inv-1', name: 'Tech ETF', type: 'STOCK', capital: 2000000, current_value: 2250000 },
    ],
  },
  liabilities: 0,
  net_worth: 13250000,
};

describe('ReportsManager Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders monthly report by default with all sections', async () => {
    vi.spyOn(reportsApi, 'fetchMonthlyReport').mockResolvedValueOnce(mockReport);

    render(<ReportsManager />);

    expect(screen.getByText(/Compiling financial statements/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('📑 Reports & Asset Snapshots')).toBeInTheDocument();
    });

    // Check KPIs
    expect(screen.getByText('Total Income')).toBeInTheDocument();
    expect(screen.getByText('Total Expenses')).toBeInTheDocument();
    expect(screen.getByText('Account Transfers')).toBeInTheDocument();
    expect(screen.getByText('Consolidated Net Worth')).toBeInTheDocument();

    // Check budget table
    expect(screen.getByText('Operational')).toBeInTheDocument();
    expect(screen.getByText('Utilities')).toBeInTheDocument();
    expect(screen.getByText('Overspent')).toBeInTheDocument();
    expect(screen.getByText('On Track')).toBeInTheDocument();

    // Check assets breakdown
    expect(screen.getByText('Main BCA')).toBeInTheDocument();
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('Tech ETF')).toBeInTheDocument();
  });

  it('switches between monthly and cycle reports', async () => {
    vi.spyOn(reportsApi, 'fetchMonthlyReport').mockResolvedValue(mockReport);
    const fetchCycleSpy = vi.spyOn(reportsApi, 'fetchCycleReport').mockResolvedValue({
      ...mockReport,
      period: '2026-08-25 to 2026-09-24',
      start_date: '2026-08-25',
      end_date: '2026-09-24',
    });

    render(<ReportsManager />);

    await waitFor(() => {
      expect(screen.getByText('Operational')).toBeInTheDocument();
    });

    // Switch to Cycle report
    const cycleTab = screen.getByTestId('report-tab-cycle');
    fireEvent.click(cycleTab);

    await waitFor(() => {
      expect(fetchCycleSpy).toHaveBeenCalled();
      expect(screen.getByText('🔄 Financial Cycle Statement')).toBeInTheDocument();
      expect(screen.getAllByText(/2026-08-25 to 2026-09-24/i).length).toBeGreaterThan(0);
    });
  });

  it('opens and closes JSON export modal', async () => {
    vi.spyOn(reportsApi, 'fetchMonthlyReport').mockResolvedValue(mockReport);

    render(<ReportsManager />);

    await waitFor(() => {
      expect(screen.getByText('Operational')).toBeInTheDocument();
    });

    const exportBtn = screen.getByRole('button', { name: /View JSON/i });
    fireEvent.click(exportBtn);

    expect(screen.getByText('📋 Report JSON Export')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Copy JSON/i })).toBeInTheDocument();

    const closeBtn = screen.getByRole('button', { name: 'Close' });
    fireEvent.click(closeBtn);

    expect(screen.queryByText('📋 Report JSON Export')).not.toBeInTheDocument();
  });

  it('displays error state when report generation fails', async () => {
    vi.spyOn(reportsApi, 'fetchMonthlyReport').mockRejectedValueOnce(
      new Error('Database unavailable')
    );

    render(<ReportsManager />);

    await waitFor(() => {
      expect(screen.getByText(/Database unavailable/i)).toBeInTheDocument();
    });

    expect(screen.getByRole('button', { name: /Retry/i })).toBeInTheDocument();
  });
});
