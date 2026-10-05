import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { AllocationManager } from './AllocationManager';
import * as allocationsApi from '../api/allocations';
import * as categoriesApi from '../api/categories';
import * as cycleApi from '../api/cycle';

describe('AllocationManager Component', () => {
  const mockCycleSummary: cycleApi.CycleSummary = {
    cycle: {
      cycle_start_day: 1,
      start_date: '2026-10-01',
      end_date: '2026-10-31',
      day_of_cycle: 5,
      total_days: 31,
      days_remaining: 26,
    },
    total_income: 1000000000,
    total_expense: 300000000,
    net_savings: 700000000,
    transaction_count: 5,
  };

  const mockCategories: categoriesApi.Category[] = [
    {
      id: 'cat-1',
      user_id: 'user-1',
      name: 'Essential Living',
      type: 'EXPENSE',
      icon: 'home',
      color: '#3B82F6',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    },
    {
      id: 'cat-2',
      user_id: 'user-1',
      name: 'Emergency Fund',
      type: 'EXPENSE',
      icon: 'shield',
      color: '#10B981',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    },
  ];

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(cycleApi, 'getCycleSummary').mockResolvedValue(mockCycleSummary);
    vi.spyOn(categoriesApi, 'listCategories').mockResolvedValue(mockCategories);
  });

  it('renders allocation summary bar and cards', async () => {
    vi.spyOn(allocationsApi, 'listAllocations').mockResolvedValue({
      allocations: [
        {
          id: 'alloc-1',
          user_id: 'user-1',
          category_id: 'cat-1',
          cycle_start: '2026-10-01',
          cycle_end: '2026-10-31',
          allocated_amount: 500000000,
          category_name: 'Essential Living',
          category_color: '#3B82F6',
          created_at: '2026-10-01T00:00:00Z',
          updated_at: '2026-10-01T00:00:00Z',
        },
      ],
      total_allocated: 500000000,
      total_income: 1000000000,
      remaining_income: 500000000,
    });

    render(<AllocationManager />);

    await waitFor(() => {
      expect(screen.getByText('Essential Living')).toBeInTheDocument();
      expect(screen.getByText(/Total Allocated/i)).toBeInTheDocument();
      expect(screen.getByText(/Remaining Unallocated/i)).toBeInTheDocument();
      expect(screen.getByText(/50% of cycle income/i)).toBeInTheDocument();
    });
  });

  it('prevents allocating more than available unallocated income', async () => {
    vi.spyOn(allocationsApi, 'listAllocations').mockResolvedValue({
      allocations: [
        {
          id: 'alloc-1',
          user_id: 'user-1',
          category_id: 'cat-1',
          cycle_start: '2026-10-01',
          cycle_end: '2026-10-31',
          allocated_amount: 800000000,
          category_name: 'Essential Living',
          created_at: '2026-10-01T00:00:00Z',
          updated_at: '2026-10-01T00:00:00Z',
        },
      ],
      total_allocated: 800000000,
      total_income: 1000000000,
      remaining_income: 200000000,
    });

    render(<AllocationManager />);

    await waitFor(() => {
      expect(screen.getByText('Essential Living')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /\+ Add Allocation/i }));

    const categorySelect = screen.getByLabelText(/Category/i);
    const amountInput = screen.getByLabelText(/Allocated Amount/i);
    const submitBtn = screen.getByRole('button', { name: /Allocate/i });

    fireEvent.change(categorySelect, { target: { value: 'cat-2' } });
    fireEvent.change(amountInput, { target: { value: '3000000' } }); // Exceeds 2M remaining
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText(/Amount exceeds available unallocated income/i)).toBeInTheDocument();
    });
  });

  it('creates allocation within available income', async () => {
    vi.spyOn(allocationsApi, 'listAllocations').mockResolvedValue({
      allocations: [],
      total_allocated: 0,
      total_income: 1000000000,
      remaining_income: 1000000000,
    });
    const createSpy = vi.spyOn(allocationsApi, 'createAllocation').mockResolvedValue({
      id: 'alloc-new',
      user_id: 'user-1',
      category_id: 'cat-1',
      cycle_start: '2026-10-01',
      cycle_end: '2026-10-31',
      allocated_amount: 400000000,
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    });

    render(<AllocationManager />);

    await waitFor(() => {
      expect(screen.getByText(/No income allocations for this cycle yet/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /\+ Add Allocation/i }));

    const categorySelect = screen.getByLabelText(/Category/i);
    const amountInput = screen.getByLabelText(/Allocated Amount/i);
    const submitBtn = screen.getByRole('button', { name: /Allocate/i });

    fireEvent.change(categorySelect, { target: { value: 'cat-1' } });
    fireEvent.change(amountInput, { target: { value: '4000000' } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(createSpy).toHaveBeenCalledWith({
        category_id: 'cat-1',
        cycle_start: '2026-10-01',
        cycle_end: '2026-10-31',
        allocated_amount: 400000000,
      });
    });
  });
});
