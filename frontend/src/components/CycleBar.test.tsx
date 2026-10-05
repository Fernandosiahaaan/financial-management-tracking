import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { CycleBar } from './CycleBar';
import * as cycleApi from '../api/cycle';

describe('CycleBar Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  const mockSummary: cycleApi.CycleSummary = {
    cycle: {
      cycle_start_day: 25,
      start_date: '2026-08-25',
      end_date: '2026-09-24',
      total_days: 31,
      day_of_cycle: 6,
      days_remaining: 25,
    },
    total_income: 1000000000,
    total_expense: 200000000,
    net_savings: 800000000,
    transaction_count: 3,
  };

  it('renders financial cycle range, progress, and aggregates', async () => {
    vi.spyOn(cycleApi, 'getCycleSummary').mockResolvedValue(mockSummary);

    render(<CycleBar />);

    await waitFor(() => {
      expect(screen.getByText(/25 Aug 2026 — 24 Sep 2026/i)).toBeInTheDocument();
      expect(screen.getByText(/Starts Day 25/i)).toBeInTheDocument();
      expect(screen.getByText(/Day 6 of 31 • 25 days remaining/i)).toBeInTheDocument();
      expect(screen.getByText(/Cycle Income/i)).toBeInTheDocument();
      expect(screen.getByText(/\+\s*Rp\s*10\.000\.000,00/i)).toBeInTheDocument();
      expect(screen.getByText(/-\s*Rp\s*2\.000\.000,00/i)).toBeInTheDocument();
      expect(screen.getByText(/\+\s*Rp\s*8\.000\.000,00/i)).toBeInTheDocument();
    });
  });

  it('navigates to previous cycle on clicking Prev Cycle', async () => {
    const fetchSpy = vi.spyOn(cycleApi, 'getCycleSummary').mockResolvedValue(mockSummary);

    render(<CycleBar />);

    await waitFor(() => {
      expect(screen.getByText(/25 Aug 2026 — 24 Sep 2026/i)).toBeInTheDocument();
    });

    const prevBtn = screen.getByRole('button', { name: /◀ Prev Cycle/i });
    fireEvent.click(prevBtn);

    await waitFor(() => {
      // 1 day before 2026-08-25 is 2026-08-24
      expect(fetchSpy).toHaveBeenCalledWith('2026-08-24');
    });
  });

  it('navigates to next cycle on clicking Next Cycle', async () => {
    const fetchSpy = vi.spyOn(cycleApi, 'getCycleSummary').mockResolvedValue(mockSummary);

    render(<CycleBar />);

    await waitFor(() => {
      expect(screen.getByText(/25 Aug 2026 — 24 Sep 2026/i)).toBeInTheDocument();
    });

    const nextBtn = screen.getByRole('button', { name: /Next Cycle ▶/i });
    fireEvent.click(nextBtn);

    await waitFor(() => {
      // 1 day after 2026-09-24 is 2026-09-25
      expect(fetchSpy).toHaveBeenCalledWith('2026-09-25');
    });
  });
});
