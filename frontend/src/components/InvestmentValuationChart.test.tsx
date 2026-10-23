import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { InvestmentValuationChart } from './InvestmentValuationChart';
import * as investmentsApi from '../api/investments';

const mockInvestments: investmentsApi.Investment[] = [
  {
    id: 'inv-1',
    user_id: 'u-1',
    type: 'STOCK',
    name: 'BBCA Stock',
    capital: 2000000000,
    current_value: 2500000000,
    unrealized_gain: 500000000,
    unrealized_gain_percentage: 25,
    notes: 'Long term',
    created_at: '2026-10-01T00:00:00Z',
    updated_at: '2026-10-01T00:00:00Z',
  },
];

const mockValuationHistory: investmentsApi.ValuationHistoryPoint[] = [
  {
    id: 'vh-1',
    investment_id: 'inv-1',
    investment_name: 'BBCA Stock',
    investment_type: 'STOCK',
    valuation: 2000000000,
    formatted_valuation: 'Rp 20.000.000,00',
    unrealized_gain: 0,
    formatted_unrealized_gain: 'Rp 0,00',
    gain_percentage: 0,
    notes: 'Initial purchase',
    recorded_at: '2026-10-01T00:00:00Z',
  },
  {
    id: 'vh-2',
    investment_id: 'inv-1',
    investment_name: 'BBCA Stock',
    investment_type: 'STOCK',
    valuation: 2500000000,
    formatted_valuation: 'Rp 25.000.000,00',
    unrealized_gain: 500000000,
    formatted_unrealized_gain: 'Rp 5.000.000,00',
    gain_percentage: 25,
    notes: 'Market rally valuation update',
    recorded_at: '2026-10-15T00:00:00Z',
  },
];

describe('InvestmentValuationChart Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders comparison mode with capital vs current valuation and KPIs', async () => {
    vi.spyOn(investmentsApi, 'fetchValuationHistory').mockResolvedValue(mockValuationHistory);

    render(<InvestmentValuationChart investments={mockInvestments} />);

    await waitFor(() => {
      expect(screen.getByText('Valuasi & Performa Investasi')).toBeInTheDocument();
      expect(screen.getByText('Total Nilai Pasar Terkini')).toBeInTheDocument();
    });

    expect(screen.getByText('Total Modal Diinvestasikan')).toBeInTheDocument();
    expect(screen.getByText('Unrealized Gain / Loss')).toBeInTheDocument();
    expect(screen.getByText(/2 Catatan/i)).toBeInTheDocument();
    expect(screen.getByText('Modal Awal (Capital)')).toBeInTheDocument();
    expect(screen.getByText('Nilai Pasar Terkini (Current Valuation)')).toBeInTheDocument();
  });

  it('switches to timeline mode and renders update history', async () => {
    vi.spyOn(investmentsApi, 'fetchValuationHistory').mockResolvedValue(mockValuationHistory);

    render(<InvestmentValuationChart investments={mockInvestments} />);

    await waitFor(() => {
      expect(screen.getByText('Valuasi & Performa Investasi')).toBeInTheDocument();
    });

    const timelineTab = screen.getByRole('button', { name: /Histori Update/i });
    fireEvent.click(timelineTab);

    await waitFor(() => {
      expect(screen.getByText('Waktu Update')).toBeInTheDocument();
      expect(screen.getByText('Market rally valuation update')).toBeInTheDocument();
    });
  });
});
