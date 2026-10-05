import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { BudgetManager } from './BudgetManager';
import * as budgetsApi from '../api/budgets';
import * as categoriesApi from '../api/categories';
import * as cycleApi from '../api/cycle';

describe('BudgetManager Component', () => {
  const mockCycleSummary: cycleApi.CycleSummary = {
    cycle: {
      cycle_start_day: 1,
      start_date: '2026-10-01',
      end_date: '2026-10-31',
      day_of_cycle: 5,
      total_days: 31,
      days_remaining: 26,
    },
    total_income: 10000000,
    total_expense: 3000000,
    net_savings: 7000000,
    transaction_count: 5,
  };

  const mockCategories: categoriesApi.Category[] = [
    {
      id: 'cat-1',
      user_id: 'user-1',
      name: 'Groceries',
      type: 'EXPENSE',
      icon: 'basket',
      color: '#EF4444',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    },
    {
      id: 'cat-2',
      user_id: 'user-1',
      name: 'Dining Out',
      type: 'EXPENSE',
      icon: 'utensils',
      color: '#F59E0B',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    },
  ];

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(cycleApi, 'getCycleSummary').mockResolvedValue(mockCycleSummary);
    vi.spyOn(categoriesApi, 'listCategories').mockResolvedValue(mockCategories);
  });

  it('renders budget items and summary variance', async () => {
    vi.spyOn(budgetsApi, 'listBudgets').mockResolvedValue([
      {
        id: 'budget-1',
        user_id: 'user-1',
        category_id: 'cat-1',
        cycle_start: '2026-10-01',
        cycle_end: '2026-10-31',
        planned_amount: 3000000,
        actual_amount: 1500000,
        variance: 1500000,
        category_name: 'Groceries',
        category_color: '#EF4444',
        created_at: '2026-10-01T00:00:00Z',
        updated_at: '2026-10-01T00:00:00Z',
      },
    ]);

    render(<BudgetManager />);

    await waitFor(() => {
      expect(screen.getByText('Groceries')).toBeInTheDocument();
      expect(screen.getByText(/Total Planned/i)).toBeInTheDocument();
      expect(screen.getByText(/50% used/i)).toBeInTheDocument();
    });
  });

  it('renders overspent indicator when actual exceeds planned', async () => {
    vi.spyOn(budgetsApi, 'listBudgets').mockResolvedValue([
      {
        id: 'budget-1',
        user_id: 'user-1',
        category_id: 'cat-1',
        cycle_start: '2026-10-01',
        cycle_end: '2026-10-31',
        planned_amount: 1000000,
        actual_amount: 1500000,
        variance: -500000,
        category_name: 'Groceries',
        category_color: '#EF4444',
        created_at: '2026-10-01T00:00:00Z',
        updated_at: '2026-10-01T00:00:00Z',
      },
    ]);

    render(<BudgetManager />);

    await waitFor(() => {
      expect(screen.getByText(/OVERSPENT/i)).toBeInTheDocument();
    });
  });

  it('creates a new budget successfully', async () => {
    vi.spyOn(budgetsApi, 'listBudgets').mockResolvedValue([]);
    const createSpy = vi.spyOn(budgetsApi, 'createBudget').mockResolvedValue({
      id: 'budget-new',
      user_id: 'user-1',
      category_id: 'cat-1',
      cycle_start: '2026-10-01',
      cycle_end: '2026-10-31',
      planned_amount: 2500000,
      actual_amount: 0,
      variance: 2500000,
      category_name: 'Groceries',
      created_at: '2026-10-01T00:00:00Z',
      updated_at: '2026-10-01T00:00:00Z',
    });

    render(<BudgetManager />);

    await waitFor(() => {
      expect(screen.getByText(/No budgets set for this cycle yet/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /\+ Add Budget/i }));

    const categorySelect = screen.getByLabelText(/Category/i);
    const amountInput = screen.getByLabelText(/Planned Amount/i);
    const submitBtn = screen.getByRole('button', { name: /Create/i });

    fireEvent.change(categorySelect, { target: { value: 'cat-1' } });
    fireEvent.change(amountInput, { target: { value: '2500000' } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(createSpy).toHaveBeenCalledWith({
        category_id: 'cat-1',
        cycle_start: '2026-10-01',
        cycle_end: '2026-10-31',
        planned_amount: 250000000,
      });
    });
  });
});
