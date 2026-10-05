import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { CategoriesManager } from './CategoriesManager';
import * as categoriesApi from '../api/categories';

describe('CategoriesManager Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders categories and filters by tabs', async () => {
    vi.spyOn(categoriesApi, 'listCategories').mockResolvedValue([
      {
        id: 'cat-1',
        user_id: 'user-1',
        name: 'Groceries',
        type: 'EXPENSE',
        icon: 'shopping-cart',
        color: '#10B981',
        created_at: '2026-10-04T00:00:00Z',
        updated_at: '2026-10-04T00:00:00Z',
      },
      {
        id: 'cat-2',
        user_id: 'user-1',
        name: 'Freelance',
        type: 'INCOME',
        icon: 'briefcase',
        color: '#6366F1',
        created_at: '2026-10-04T00:00:00Z',
        updated_at: '2026-10-04T00:00:00Z',
      },
    ]);

    render(<CategoriesManager />);

    await waitFor(() => {
      expect(screen.getByText('Groceries')).toBeInTheDocument();
      expect(screen.getByText('Freelance')).toBeInTheDocument();
    });

    // Click Expenses tab
    fireEvent.click(screen.getByRole('button', { name: /Expenses/i }));
    expect(screen.getByText('Groceries')).toBeInTheDocument();
    expect(screen.queryByText('Freelance')).not.toBeInTheDocument();
  });

  it('creates new category successfully', async () => {
    vi.spyOn(categoriesApi, 'listCategories').mockResolvedValue([]);
    const createSpy = vi.spyOn(categoriesApi, 'createCategory').mockResolvedValue({
      id: 'cat-new',
      user_id: 'user-1',
      name: 'Coffee',
      type: 'EXPENSE',
      icon: 'tag',
      color: '#6366F1',
      created_at: '2026-10-04T00:00:00Z',
      updated_at: '2026-10-04T00:00:00Z',
    });

    render(<CategoriesManager />);

    await waitFor(() => {
      expect(screen.getByText(/No Categories Found/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /\+ Add Category/i }));

    const nameInput = screen.getByLabelText(/Category Name/i);
    const submitBtn = screen.getByRole('button', { name: /Create Category/i });

    fireEvent.change(nameInput, { target: { value: 'Coffee' } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(createSpy).toHaveBeenCalledWith({
        name: 'Coffee',
        type: 'EXPENSE',
        icon: 'tag',
        color: '#6366F1',
      });
    });
  });
});
