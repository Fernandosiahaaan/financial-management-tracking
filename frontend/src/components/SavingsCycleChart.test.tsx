import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { SavingsCycleChart } from './SavingsCycleChart';
import * as reportsApi from '../api/reports';

const mockSavingsTrend: reportsApi.SavingsTrendResponse = {
  cycles: [
    {
      period: '2026-07-25 to 2026-08-24',
      start_date: '2026-07-25',
      end_date: '2026-08-24',
      label: 'Siklus Agt 2026',
      income: 10000000,
      formatted_income: 'Rp 10.000.000,00',
      expense: 7000000,
      formatted_expense: 'Rp 7.000.000,00',
      net_savings: 3000000,
      formatted_net_savings: 'Rp 3.000.000,00',
      saving_rate: 30,
      change_amount: 0,
      formatted_change_amount: 'Rp 0,00',
      change_percentage: 0,
      status: 'STABLE',
    },
    {
      period: '2026-08-25 to 2026-09-24',
      start_date: '2026-08-25',
      end_date: '2026-09-24',
      label: 'Siklus Sep 2026',
      income: 12000000,
      formatted_income: 'Rp 12.000.000,00',
      expense: 7000000,
      formatted_expense: 'Rp 7.000.000,00',
      net_savings: 5000000,
      formatted_net_savings: 'Rp 5.000.000,00',
      saving_rate: 41.7,
      change_amount: 2000000,
      formatted_change_amount: 'Rp 2.000.000,00',
      change_percentage: 66.7,
      status: 'INCREASED',
    },
  ],
  average_savings: 4000000,
  formatted_average_savings: 'Rp 4.000.000,00',
  highest_savings: 5000000,
  formatted_highest_savings: 'Rp 5.000.000,00',
  lowest_savings: 3000000,
  formatted_lowest_savings: 'Rp 3.000.000,00',
  current_vs_previous_diff: 2000000,
  current_status: 'INCREASED',
};

describe('SavingsCycleChart Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders savings cycle trend chart with summary metrics and cycle points', async () => {
    vi.spyOn(reportsApi, 'fetchSavingsTrend').mockResolvedValue(mockSavingsTrend);

    render(<SavingsCycleChart initialCycles={6} />);

    await waitFor(() => {
      expect(screen.getByText('Tren Tabungan Antar Siklus')).toBeInTheDocument();
      expect(screen.getByText('Performa Siklus Terakhir')).toBeInTheDocument();
    });

    expect(screen.getByText(/Naik/i)).toBeInTheDocument();
    expect(screen.getByText('Rata-rata Tabungan')).toBeInTheDocument();
    expect(screen.getByText('Tabungan Tertinggi')).toBeInTheDocument();
    expect(screen.getByText('Tabungan Terendah')).toBeInTheDocument();
    expect(screen.getAllByText('Siklus Sep 2026').length).toBeGreaterThan(0);
  });

  it('renders empty state when no cycles exist', async () => {
    vi.spyOn(reportsApi, 'fetchSavingsTrend').mockResolvedValue({
      cycles: [],
      average_savings: 0,
      formatted_average_savings: 'Rp 0,00',
      highest_savings: 0,
      formatted_highest_savings: 'Rp 0,00',
      lowest_savings: 0,
      formatted_lowest_savings: 'Rp 0,00',
      current_vs_previous_diff: 0,
      current_status: 'STABLE',
    });

    render(<SavingsCycleChart />);

    await waitFor(() => {
      expect(screen.getByText(/Belum ada data transaksi/i)).toBeInTheDocument();
    });
  });
});
