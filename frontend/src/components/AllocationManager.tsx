import React, { useEffect, useState, useCallback } from 'react';
import { getCycleSummary, type CycleSummary } from '../api/cycle';
import {
  listAllocations,
  createAllocation,
  updateAllocation,
  deleteAllocation,
  type Allocation,
  type AllocationSummary,
  type CreateAllocationInput,
} from '../api/allocations';
import { listCategories, type Category } from '../api/categories';
import { listBudgets, type Budget } from '../api/budgets';
import { formatRupiah, formatNumberIDR, parseRupiahToCents, formatCurrencyInput } from '../utils/currency';
import { Plus, Pencil, Trash2 } from 'lucide-react';

/** Format IDR amount */
function fmtIDR(amount: number): string {
  return formatRupiah(amount);
}

export const AllocationManager: React.FC = () => {
  const [cycleSummary, setCycleSummary] = useState<CycleSummary | null>(null);
  const [summaryData, setSummaryData] = useState<AllocationSummary | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [budgets, setBudgets] = useState<Budget[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [formError, setFormError] = useState('');

  // Form state
  const [showForm, setShowForm] = useState(false);
  const [editingAllocation, setEditingAllocation] = useState<Allocation | null>(null);
  const [selectedCategoryID, setSelectedCategoryID] = useState('');
  const [allocatedAmount, setAllocatedAmount] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [summary, cats] = await Promise.all([getCycleSummary(), listCategories()]);
      setCycleSummary(summary);
      setCategories(cats);

      const [allocSummary, bList] = await Promise.all([
        listAllocations(summary.cycle.start_date, summary.cycle.end_date),
        listBudgets(summary.cycle.start_date, summary.cycle.end_date).catch(() => []),
      ]);
      setSummaryData(allocSummary);
      setBudgets(bList);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load allocation data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const allocations = summaryData?.allocations || [];
  const totalIncome = summaryData?.total_income || 0;
  const totalAllocated = summaryData?.total_allocated || 0;
  const remainingIncome = summaryData?.remaining_income || 0;

  // Buying Power & Budget Savings Intelligence
  const budgetBuyingPower = Math.max(0, remainingIncome);
  const totalPlannedBudget = budgets.reduce((sum, b) => sum + b.planned_amount, 0);
  const projectedCycleSavings = budgetBuyingPower > totalPlannedBudget ? budgetBuyingPower - totalPlannedBudget : 0;
  const totalCycleSavings = totalAllocated + projectedCycleSavings;

  // Categories already allocated for this cycle
  const allocatedCategoryIDs = new Set(allocations.map((a) => a.category_id));
  const availableCategories = categories.filter((c) => !allocatedCategoryIDs.has(c.id));

  const resetForm = () => {
    setShowForm(false);
    setEditingAllocation(null);
    setSelectedCategoryID('');
    setAllocatedAmount('');
    setFormError('');
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    if (!cycleSummary) return;

    const amount = parseRupiahToCents(allocatedAmount);
    if (isNaN(amount) || amount < 0) {
      setFormError('Allocated amount must be a non-negative number');
      return;
    }

    if (!selectedCategoryID) {
      setFormError('Please select a category');
      return;
    }

    if (amount > remainingIncome) {
      setFormError(`Amount exceeds available unallocated income (${fmtIDR(remainingIncome)})`);
      return;
    }

    setSubmitting(true);
    try {
      const input: CreateAllocationInput = {
        category_id: selectedCategoryID,
        cycle_start: cycleSummary.cycle.start_date,
        cycle_end: cycleSummary.cycle.end_date,
        allocated_amount: amount,
      };
      await createAllocation(input);
      resetForm();
      await loadData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to create allocation');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    if (!editingAllocation) return;

    const amount = parseRupiahToCents(allocatedAmount);
    if (isNaN(amount) || amount < 0) {
      setFormError('Allocated amount must be a non-negative number');
      return;
    }

    // Maximum allowed is other allocations + this one <= totalIncome, so remaining + editingAllocation.allocated_amount
    const maxAllowed = remainingIncome + editingAllocation.allocated_amount;
    if (amount > maxAllowed) {
      setFormError(`Amount exceeds total available cycle income (${fmtIDR(maxAllowed)})`);
      return;
    }

    setSubmitting(true);
    try {
      await updateAllocation(editingAllocation.id, { allocated_amount: amount });
      resetForm();
      await loadData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to update allocation');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to remove this allocation?')) return;
    try {
      await deleteAllocation(id);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete allocation');
    }
  };

  const startEdit = (a: Allocation) => {
    setEditingAllocation(a);
    setAllocatedAmount(formatNumberIDR(a.allocated_amount));
    setShowForm(true);
    setFormError('');
  };

  const pctAllocated = totalIncome > 0
    ? Math.min(Math.round((totalAllocated / totalIncome) * 100), 100)
    : 0;

  if (loading) {
    return (
      <section className="manager-card" id="allocation-manager">
        <div className="manager-card__header">
          <h2 className="manager-card__title">🎯 Income Allocation</h2>
        </div>
        <div className="health-loading">
          <div className="spinner" />
          <span>Loading allocations…</span>
        </div>
      </section>
    );
  }

  return (
    <section className="manager-card" id="allocation-manager">
      <div className="manager-card__header">
        <div>
          <h2 className="section-title manager-card__title">🎯 Income Allocation</h2>
          <p className="section-subtitle manager-card__subtitle">
            Distribute earned cycle income across designated purposes. Total allocations cannot exceed total income.
          </p>
        </div>
        {!showForm && (
          <button
            className="btn btn--primary btn-action-add btn--sm"
            onClick={() => { resetForm(); setShowForm(true); }}
            disabled={availableCategories.length === 0 || remainingIncome <= 0}
            id="allocation-add-btn"
            title={remainingIncome <= 0 ? 'No unallocated income remaining' : 'Add allocation'}
          >
            <Plus size={16} className="btn-icon-svg" />
            <span>Add Allocation</span>
          </button>
        )}
      </div>

      {error && (
        <div className="auth-alert auth-alert--error">
          <span>⚠️ {error}</span>
        </div>
      )}

      {/* Allocation Summary Bar */}
      <div className="allocation-summary-bar" id="allocation-summary">
        <div className="allocation-summary-chip">
          <span className="allocation-summary-chip__label">Total Income</span>
          <span className="allocation-summary-chip__value income">{fmtIDR(totalIncome)}</span>
        </div>
        <div className="allocation-summary-chip">
          <span className="allocation-summary-chip__label">Total Allocated</span>
          <span className="allocation-summary-chip__value allocated">{fmtIDR(totalAllocated)}</span>
        </div>
        <div className={`allocation-summary-chip ${remainingIncome < 0 ? 'remaining-negative' : 'remaining-positive'}`}>
          <span className="allocation-summary-chip__label">Remaining Unallocated</span>
          <span className="allocation-summary-chip__value">{fmtIDR(remainingIncome)}</span>
        </div>
        {budgets.length > 0 && (
          <>
            <div className="allocation-summary-chip">
              <span className="allocation-summary-chip__label">Budget Planned</span>
              <span className="allocation-summary-chip__value">{fmtIDR(totalPlannedBudget)}</span>
            </div>
            <div className="allocation-summary-chip">
              <span className="allocation-summary-chip__label">Projected Savings</span>
              <span className="allocation-summary-chip__value text-success">+{fmtIDR(projectedCycleSavings)}</span>
            </div>
            <div className="allocation-summary-chip">
              <span className="allocation-summary-chip__label">Total Cycle Savings</span>
              <span className="allocation-summary-chip__value text-success">+{fmtIDR(totalCycleSavings)}</span>
            </div>
          </>
        )}
      </div>

      {/* Overall Distribution Progress Track */}
      {totalIncome > 0 && (
        <div className="allocation-overall-progress">
          <div className="allocation-progress-bar">
            <div
              className="allocation-progress-bar__fill"
              style={{ width: `${pctAllocated}%` }}
            />
          </div>
          <div className="allocation-progress-labels">
            <span>{pctAllocated}% of income allocated</span>
            <span>{fmtIDR(remainingIncome)} unallocated</span>
          </div>
        </div>
      )}

      {/* Create / Edit Form */}
      {showForm && (
        <form
          className="manager-form"
          onSubmit={editingAllocation ? handleUpdate : handleCreate}
          id="allocation-form"
        >
          <h3 className="manager-form__title">
            {editingAllocation ? `Edit Allocation — ${editingAllocation.category_name}` : 'New Allocation'}
          </h3>

          {formError && (
            <div className="auth-alert auth-alert--error">
              <span>⚠️ {formError}</span>
            </div>
          )}

          {!editingAllocation && (
            <div className="form-group">
              <label htmlFor="allocation-category">Category</label>
              <select
                id="allocation-category"
                className="form-select"
                value={selectedCategoryID}
                onChange={(e) => setSelectedCategoryID(e.target.value)}
                required
              >
                <option value="">Select a category…</option>
                {availableCategories.map((cat) => (
                  <option key={cat.id} value={cat.id}>
                    {cat.name} ({cat.type === 'EXPENSE' ? 'Expense' : 'Income'})
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="form-group">
            <label htmlFor="allocation-amount">Allocated Amount (IDR)</label>
            <input
              id="allocation-amount"
              type="text"
              value={allocatedAmount}
              onChange={(e) => setAllocatedAmount(e.target.value)}
              onBlur={() => setAllocatedAmount(formatCurrencyInput(allocatedAmount))}
              placeholder="0,00"
              required
            />
            <span className="form-hint">Format: xxx.xxx,00 (contoh: 3.000.000,00)</span>
            <span className="form-hint">
              Available to allocate: <strong>{fmtIDR(editingAllocation ? remainingIncome + editingAllocation.allocated_amount : remainingIncome)}</strong>
            </span>
          </div>

          <div className="form-actions">
            <button type="submit" className="btn btn--primary btn--sm" disabled={submitting}>
              {submitting ? 'Saving…' : editingAllocation ? 'Update' : 'Allocate'}
            </button>
            <button type="button" className="btn btn--outline btn--sm" onClick={resetForm}>
              Cancel
            </button>
          </div>
        </form>
      )}

      {/* Allocation List */}
      {allocations.length === 0 && !showForm ? (
        <div className="manager-empty" id="allocation-empty">
          <p>No income allocations for this cycle yet.</p>
          <p className="manager-empty__hint">
            {totalIncome > 0
              ? 'Click "+ Add Allocation" to assign parts of your earned cycle income to specific categories.'
              : 'Log an income transaction first to unlock income allocations for this cycle.'}
          </p>
        </div>
      ) : (
        <div className="allocation-grid" id="allocation-list">
          {allocations.map((a) => {
            const share = totalIncome > 0
              ? Math.round((a.allocated_amount / totalIncome) * 100)
              : 0;

            return (
              <div
                key={a.id}
                className="allocation-card"
                id={`allocation-${a.id}`}
              >
                <div className="allocation-card__header">
                  <span
                    className="allocation-card__category-dot"
                    style={{ backgroundColor: a.category_color || '#6366F1' }}
                  />
                  <div className="allocation-card__title-group">
                    <span className="allocation-card__category-name">
                      {a.category_name || 'Unknown Category'}
                    </span>
                    <span className="allocation-card__share-badge">
                      {share}% of cycle income
                    </span>
                  </div>
                  <div className="allocation-card__actions">
                    <button
                      className="btn-action btn-action--edit"
                      onClick={() => startEdit(a)}
                      title="Edit allocation"
                      aria-label="Edit allocation"
                    >
                      <Pencil size={13} />
                    </button>
                    <button
                      className="btn-action btn-action--danger"
                      onClick={() => handleDelete(a.id)}
                      title="Remove allocation"
                      aria-label="Remove allocation"
                    >
                      <Trash2 size={13} />
                    </button>
                  </div>
                </div>

                <div className="allocation-card__amount">
                  <span className="allocation-card__amount-label">Allocated Amount</span>
                  <span className="allocation-card__amount-value">{fmtIDR(a.allocated_amount)}</span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
};
