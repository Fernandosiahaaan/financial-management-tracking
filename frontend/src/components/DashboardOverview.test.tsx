import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { DashboardOverview } from './DashboardOverview';
import * as dashboardApi from '../api/dashboard';

describe('DashboardOverview Component', () => {
  const mockDashboardData: dashboardApi.DashboardData = {
    cycle: {
      start_date: '2026-10-01',
      end_date: '2026-10-31',
      cycle_start_day: 1,
      day_of_cycle: 10,
      total_days: 31,
      days_remaining: 21,
    },
    income: 1200000000,
    expense: 400000000,
    net_cash_flow: 800000000,
    total_assets: 2500000000,
    previous_cycle_assets: 1700000000,
    asset_growth: 800000000,
    budgets: [
      {
        category_id: 'cat-1',
        category_name: 'Groceries',
        category_color: '#EF4444',
        planned: 300000000,
        actual: 350000000,
        variance: -50000000,
        overspent: true,
      },
      {
        category_id: 'cat-2',
        category_name: 'Utilities',
        category_color: '#3B82F6',
        planned: 100000000,
        actual: 50000000,
        variance: 50000000,
        overspent: false,
      },
    ],
    accounts: [
      {
        id: 'acc-1',
        name: 'Main Checking',
        type: 'BANK',
        balance: 2000000000,
      },
      {
        id: 'acc-2',
        name: 'Pocket Cash',
        type: 'CASH',
        balance: 500000000,
      },
    ],
  };

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders primary KPI cards and asset growth correctly', async () => {
    vi.spyOn(dashboardApi, 'getDashboardData').mockResolvedValue(mockDashboardData);

    render(<DashboardOverview />);

    await waitFor(() => {
      // Total assets
      expect(screen.getByText(/Total Net Assets/i)).toBeInTheDocument();
      expect(screen.getByText(/Rp\s*25\.000\.000,00/)).toBeInTheDocument();

      // Asset growth
      expect(screen.getByText(/▲ \+Rp\s*8\.000\.000,00/)).toBeInTheDocument();

      // Income and Expense
      expect(screen.getByText(/Rp\s*12\.000\.000,00/)).toBeInTheDocument();
      expect(screen.getByText(/Rp\s*4\.000\.000,00/)).toBeInTheDocument();
    });
  });

  it('renders budget progress and highlights overspent category', async () => {
    vi.spyOn(dashboardApi, 'getDashboardData').mockResolvedValue(mockDashboardData);

    render(<DashboardOverview />);

    await waitFor(() => {
      expect(screen.getByText('Groceries')).toBeInTheDocument();
      expect(screen.getByText(/OVERSPENT by/i)).toBeInTheDocument();
      expect(screen.getByText('Utilities')).toBeInTheDocument();
      expect(screen.getByText(/Rp\s*500\.000,00 remaining/i)).toBeInTheDocument();
    });
  });

  it('renders accounts list and balances', async () => {
    vi.spyOn(dashboardApi, 'getDashboardData').mockResolvedValue(mockDashboardData);

    render(<DashboardOverview />);

    await waitFor(() => {
      expect(screen.getByText('Main Checking')).toBeInTheDocument();
      expect(screen.getByText(/Rp\s*20\.000\.000,00/)).toBeInTheDocument();
      expect(screen.getByText('Pocket Cash')).toBeInTheDocument();
      expect(screen.getByText(/Rp\s*5\.000\.000,00/)).toBeInTheDocument();
    });
  });

  it('triggers navigation on cycle buttons', async () => {
    const fetchSpy = vi.spyOn(dashboardApi, 'getDashboardData').mockResolvedValue(mockDashboardData);

    render(<DashboardOverview />);

    await waitFor(() => {
      expect(screen.getByText(/Financial Cycle/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /← Previous/i }));

    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith('2026-09-30');
    });
  });

  it('hides amounts by default and reveals on privacy toggle click', async () => {
    vi.spyOn(dashboardApi, 'getDashboardData').mockResolvedValue(mockDashboardData);
    const { PrivacyProvider } = await import('../context/PrivacyContext');

    render(
      <PrivacyProvider defaultHidden={true}>
        <DashboardOverview />
      </PrivacyProvider>
    );

    await waitFor(() => {
      expect(screen.getByText(/Total Net Assets/i)).toBeInTheDocument();
    });

    // When hidden: masked values present
    expect(screen.getAllByText('Rp ••••••••').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Rp\s*25\.000\.000,00/)).not.toBeInTheDocument();

    // Find and click the toggle button
    const toggleBtn = screen.getByRole('button', { name: /Show Amounts/i });
    fireEvent.click(toggleBtn);

    // Now revealed
    await waitFor(() => {
      expect(screen.getByText(/Rp\s*25\.000\.000,00/)).toBeInTheDocument();
    });

    // Toggle back to hidden
    const hideBtn = screen.getByRole('button', { name: /Hide Amounts/i });
    fireEvent.click(hideBtn);

    await waitFor(() => {
      expect(screen.getAllByText('Rp ••••••••').length).toBeGreaterThan(0);
      expect(screen.queryByText(/Rp\s*25\.000\.000,00/)).not.toBeInTheDocument();
    });
  });
});
