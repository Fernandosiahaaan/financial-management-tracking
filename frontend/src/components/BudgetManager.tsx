import React, { useEffect, useState, useCallback } from 'react';
import { getCycleSummary, type CycleSummary } from '../api/cycle';
import {
  listBudgets,
  createBudget,
  updateBudget,
  deleteBudget,
  type Budget,
  type CreateBudgetInput,
} from '../api/budgets';
import { listCategories, type Category } from '../api/categories';
import { formatRupiah, formatNumberIDR, parseRupiahToCents, formatCurrencyInput } from '../utils/currency';

/** Format IDR amount */
function fmtIDR(amount: number): string {
  return formatRupiah(amount);
}

export const BudgetManager: React.FC = () => {
  const [cycleSummary, setCycleSummary] = useState<CycleSummary | null>(null);
  const [budgets, setBudgets] = useState<Budget[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [formError, setFormError] = useState('');

  // Form state
  const [showForm, setShowForm] = useState(false);
  const [editingBudget, setEditingBudget] = useState<Budget | null>(null);
  const [selectedCategoryID, setSelectedCategoryID] = useState('');
  const [plannedAmount, setPlannedAmount] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [summary, cats] = await Promise.all([getCycleSummary(), listCategories()]);
      setCycleSummary(summary);
      setCategories(cats);

      const bList = await listBudgets(summary.cycle.start_date, summary.cycle.end_date);
      setBudgets(bList);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load budget data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const expenseCategories = categories.filter((c) => c.type === 'EXPENSE');

  // Categories that already have a budget for this cycle (for filtering the dropdown)
  const budgetedCategoryIDs = new Set(budgets.map((b) => b.category_id));
  const availableCategories = expenseCategories.filter((c) => !budgetedCategoryIDs.has(c.id));

  const resetForm = () => {
    setShowForm(false);
    setEditingBudget(null);
    setSelectedCategoryID('');
    setPlannedAmount('');
    setFormError('');
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    if (!cycleSummary) return;

    const amount = parseRupiahToCents(plannedAmount);
    if (isNaN(amount) || amount < 0) {
      setFormError('Planned amount must be a non-negative number');
      return;
    }

    if (!selectedCategoryID) {
      setFormError('Please select a category');
      return;
    }

    setSubmitting(true);
    try {
      const input: CreateBudgetInput = {
        category_id: selectedCategoryID,
        cycle_start: cycleSummary.cycle.start_date,
        cycle_end: cycleSummary.cycle.end_date,
        planned_amount: amount,
      };
      await createBudget(input);
      resetForm();
      await loadData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to create budget');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    if (!editingBudget) return;

    const amount = parseRupiahToCents(plannedAmount);
    if (isNaN(amount) || amount < 0) {
      setFormError('Planned amount must be a non-negative number');
      return;
    }

    setSubmitting(true);
    try {
      await updateBudget(editingBudget.id, { planned_amount: amount });
      resetForm();
      await loadData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to update budget');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to remove this budget?')) return;
    try {
      await deleteBudget(id);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete budget');
    }
  };

  const startEdit = (b: Budget) => {
    setEditingBudget(b);
    setPlannedAmount(formatNumberIDR(b.planned_amount));
    setShowForm(true);
    setFormError('');
  };

  // Compute totals
  const totalPlanned = budgets.reduce((sum, b) => sum + b.planned_amount, 0);
  const totalActual = budgets.reduce((sum, b) => sum + b.actual_amount, 0);
  const totalVariance = totalPlanned - totalActual;

  if (loading) {
    return (
      <section className="manager-card" id="budget-manager">
        <div className="manager-card__header">
          <h2 className="manager-card__title">📊 Budget</h2>
        </div>
        <div className="health-loading">
          <div className="spinner" />
          <span>Loading budgets…</span>
        </div>
      </section>
    );
  }

  return (
    <section className="manager-card" id="budget-manager">
      <div className="manager-card__header">
        <h2 className="manager-card__title">📊 Budget</h2>
        {!showForm && (
          <button
            className="btn btn--primary btn--sm"
            onClick={() => { resetForm(); setShowForm(true); }}
            disabled={availableCategories.length === 0}
            id="budget-add-btn"
          >
            + Add Budget
          </button>
        )}
      </div>

      {error && (
        <div className="auth-alert auth-alert--error">
          <span>⚠️ {error}</span>
        </div>
      )}

      {/* Summary bar */}
      {budgets.length > 0 && (
        <div className="budget-summary-bar" id="budget-summary">
          <div className="budget-summary-chip">
            <span className="budget-summary-chip__label">Total Planned</span>
            <span className="budget-summary-chip__value">{fmtIDR(totalPlanned)}</span>
          </div>
          <div className="budget-summary-chip">
            <span className="budget-summary-chip__label">Total Actual</span>
            <span className="budget-summary-chip__value">{fmtIDR(totalActual)}</span>
          </div>
          <div className={`budget-summary-chip ${totalVariance < 0 ? 'budget-summary-chip--over' : 'budget-summary-chip--under'}`}>
            <span className="budget-summary-chip__label">Variance</span>
            <span className="budget-summary-chip__value">
              {totalVariance >= 0 ? '+' : ''}{fmtIDR(totalVariance)}
            </span>
          </div>
        </div>
      )}

      {/* Create / Edit Form */}
      {showForm && (
        <form
          className="manager-form"
          onSubmit={editingBudget ? handleUpdate : handleCreate}
          id="budget-form"
        >
          <h3 className="manager-form__title">
            {editingBudget ? `Edit Budget — ${editingBudget.category_name}` : 'New Budget'}
          </h3>

          {formError && (
            <div className="auth-alert auth-alert--error">
              <span>⚠️ {formError}</span>
            </div>
          )}

          {!editingBudget && (
            <div className="form-group">
              <label htmlFor="budget-category">Category</label>
              <select
                id="budget-category"
                value={selectedCategoryID}
                onChange={(e) => setSelectedCategoryID(e.target.value)}
                required
              >
                <option value="">Select a category…</option>
                {availableCategories.map((cat) => (
                  <option key={cat.id} value={cat.id}>
                    {cat.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="form-group">
            <label htmlFor="budget-amount">Planned Amount (IDR)</label>
            <input
              id="budget-amount"
              type="text"
              value={plannedAmount}
              onChange={(e) => setPlannedAmount(e.target.value)}
              onBlur={() => setPlannedAmount(formatCurrencyInput(plannedAmount))}
              placeholder="0,00"
              required
            />
            <span className="form-hint">Format: xxx.xxx,00 (contoh: 4.000.000,00)</span>
          </div>

          <div className="form-actions">
            <button type="submit" className="btn btn--primary btn--sm" disabled={submitting}>
              {submitting ? 'Saving…' : editingBudget ? 'Update' : 'Create'}
            </button>
            <button type="button" className="btn btn--outline btn--sm" onClick={resetForm}>
              Cancel
            </button>
          </div>
        </form>
      )}

      {/* Budget List */}
      {budgets.length === 0 && !showForm ? (
        <div className="manager-empty" id="budget-empty">
          <p>No budgets set for this cycle yet.</p>
          <p className="manager-empty__hint">Click "Add Budget" to set spending limits by category.</p>
        </div>
      ) : (
        <div className="budget-grid" id="budget-list">
          {budgets.map((b) => {
            const pct = b.planned_amount > 0
              ? Math.min(Math.round((b.actual_amount / b.planned_amount) * 100), 100)
              : 0;
            const isOver = b.variance < 0;

            return (
              <div
                key={b.id}
                className={`budget-card ${isOver ? 'budget-card--overspent' : ''}`}
                id={`budget-${b.id}`}
              >
                <div className="budget-card__header">
                  <span
                    className="budget-card__category-dot"
                    style={{ backgroundColor: b.category_color || '#6366F1' }}
                  />
                  <span className="budget-card__category-name">
                    {b.category_name || 'Unknown'}
                  </span>
                  <div className="budget-card__actions">
                    <button
                      className="btn-icon"
                      onClick={() => startEdit(b)}
                      title="Edit budget"
                    >
                      ✏️
                    </button>
                    <button
                      className="btn-icon"
                      onClick={() => handleDelete(b.id)}
                      title="Remove budget"
                    >
                      🗑️
                    </button>
                  </div>
                </div>

                <div className="budget-card__amounts">
                  <div className="budget-card__row">
                    <span className="budget-card__label">Planned</span>
                    <span className="budget-card__value">{fmtIDR(b.planned_amount)}</span>
                  </div>
                  <div className="budget-card__row">
                    <span className="budget-card__label">Actual</span>
                    <span className="budget-card__value">{fmtIDR(b.actual_amount)}</span>
                  </div>
                  <div className={`budget-card__row budget-card__row--variance ${isOver ? 'budget-card__row--negative' : 'budget-card__row--positive'}`}>
                    <span className="budget-card__label">Variance</span>
                    <span className="budget-card__value">
                      {b.variance >= 0 ? '+' : ''}{fmtIDR(b.variance)}
                    </span>
                  </div>
                </div>

                <div className="budget-card__progress">
                  <div className="budget-progress-bar">
                    <div
                      className={`budget-progress-bar__fill ${isOver ? 'budget-progress-bar__fill--over' : ''}`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <span className="budget-progress-bar__label">
                    {pct}% used{isOver ? ' — OVERSPENT' : ''}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
};
