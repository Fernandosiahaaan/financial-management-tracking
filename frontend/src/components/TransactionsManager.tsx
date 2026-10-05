import React, { useState, useEffect, useMemo } from 'react';
import {
  type Transaction,
  type TransactionType,
  listTransactions,
  createTransaction,
  updateTransaction,
  deleteTransaction,
} from '../api/transactions';
import { type Account, listAccounts } from '../api/accounts';
import { type Category, listCategories } from '../api/categories';
import { formatRupiah, formatNumberIDR, parseRupiahToCents, formatCurrencyInput } from '../utils/currency';
import { usePrivacy } from '../context/PrivacyContext';
import { PrivacyToggle } from './PrivacyToggle';

interface TransactionsManagerProps {
  onBalanceChange?: () => void;
}

export const TransactionsManager: React.FC<TransactionsManagerProps> = ({ onBalanceChange }) => {
  const { formatAmount } = usePrivacy();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [selectedType, setSelectedType] = useState<string>('ALL');
  const [selectedAccountId, setSelectedAccountId] = useState<string>('ALL');
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');

  // Modal State
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTx, setEditingTx] = useState<Transaction | null>(null);

  // Form Fields
  const [formType, setFormType] = useState<TransactionType>('EXPENSE');
  const [formAccountId, setFormAccountId] = useState('');
  const [formDestAccountId, setFormDestAccountId] = useState('');
  const [formCategoryId, setFormCategoryId] = useState('');
  const [formAmount, setFormAmount] = useState('');
  const [formDate, setFormDate] = useState(() => new Date().toISOString().split('T')[0]);
  const [formDescription, setFormDescription] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [txs, accs, cats] = await Promise.all([
        listTransactions(),
        listAccounts(),
        listCategories(),
      ]);

      setTransactions(txs);
      setAccounts(accs);
      setCategories(cats);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load transaction data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleApplyFilter = async () => {
    try {
      setLoading(true);
      setError(null);
      const filter: Parameters<typeof listTransactions>[0] = {};
      if (selectedType !== 'ALL') {
        filter.type = selectedType as TransactionType;
      }
      if (selectedAccountId !== 'ALL') {
        filter.account_id = selectedAccountId;
      }
      if (startDate) {
        filter.start_date = startDate;
      }
      if (endDate) {
        filter.end_date = endDate;
      }
      const data = await listTransactions(filter);
      setTransactions(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to filter transactions');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenCreateModal = () => {
    setEditingTx(null);
    setFormType('EXPENSE');
    setFormAccountId(accounts[0]?.id || '');
    setFormDestAccountId(accounts.length > 1 ? accounts[1]?.id : '');
    setFormCategoryId('');
    setFormAmount('');
    setFormDate(new Date().toISOString().split('T')[0]);
    setFormDescription('');
    setFormError(null);
    setIsModalOpen(true);
  };

  const handleOpenEditModal = (tx: Transaction) => {
    setEditingTx(tx);
    setFormType(tx.type);
    setFormAccountId(tx.account_id);
    setFormDestAccountId(tx.destination_account_id || '');
    setFormCategoryId(tx.category_id || '');
    setFormAmount(formatNumberIDR(tx.amount));
    setFormDate(tx.transaction_date);
    setFormDescription(tx.description || '');
    setFormError(null);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingTx(null);
    setFormError(null);
  };

  const handleFormSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    const amountNum = parseRupiahToCents(formAmount);
    if (isNaN(amountNum) || amountNum <= 0) {
      setFormError('Amount must be greater than zero');
      return;
    }

    if (!formAccountId) {
      setFormError('Source account is required');
      return;
    }

    if (formType === 'TRANSFER') {
      if (!formDestAccountId) {
        setFormError('Destination account is required for transfer');
        return;
      }
      if (formDestAccountId === formAccountId) {
        setFormError('Source and destination accounts must be different');
        return;
      }
    }

    if (!formDate) {
      setFormError('Transaction date is required');
      return;
    }

    try {
      setIsSubmitting(true);
      if (editingTx) {
        await updateTransaction(editingTx.id, {
          account_id: formAccountId,
          destination_account_id: formType === 'TRANSFER' ? formDestAccountId : undefined,
          category_id: formType !== 'TRANSFER' && formCategoryId ? formCategoryId : undefined,
          type: formType,
          amount: amountNum,
          transaction_date: formDate,
          description: formDescription,
        });
      } else {
        await createTransaction({
          account_id: formAccountId,
          destination_account_id: formType === 'TRANSFER' ? formDestAccountId : undefined,
          category_id: formType !== 'TRANSFER' && formCategoryId ? formCategoryId : undefined,
          type: formType,
          amount: amountNum,
          transaction_date: formDate,
          description: formDescription,
        });
      }

      handleCloseModal();
      await loadData();
      if (onBalanceChange) {
        onBalanceChange();
      }
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to save transaction');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this transaction? The account balance will be reversed.')) {
      return;
    }
    try {
      await deleteTransaction(id);
      await loadData();
      if (onBalanceChange) {
        onBalanceChange();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete transaction');
    }
  };

  // Summaries
  const { totalIncome, totalExpense, netFlow } = useMemo(() => {
    let inc = 0;
    let exp = 0;
    for (const tx of transactions) {
      if (tx.type === 'INCOME') inc += tx.amount;
      if (tx.type === 'EXPENSE') exp += tx.amount;
    }
    return {
      totalIncome: inc,
      totalExpense: exp,
      netFlow: inc - exp,
    };
  }, [transactions]);

  // Categories filtered for current form type
  const availableCategories = useMemo(() => {
    if (formType === 'TRANSFER') return [];
    return categories.filter((c) => c.type === formType);
  }, [categories, formType]);

  return (
    <div className="transactions-manager">
      {/* Summary KPI Banner */}
      <div className="tx-summary-banner">
        <div className="tx-stat-card income">
          <span className="stat-label">Total Income</span>
          <span className="stat-value">+{formatAmount(totalIncome)}</span>
        </div>
        <div className="tx-stat-card expense">
          <span className="stat-label">Total Expense</span>
          <span className="stat-value">-{formatAmount(totalExpense)}</span>
        </div>
        <div className={`tx-stat-card ${netFlow >= 0 ? 'net-positive' : 'net-negative'}`}>
          <div className="tx-stat-header">
            <span className="stat-label">Net Cashflow</span>
            <PrivacyToggle variant="icon" id="tx-privacy-btn" />
          </div>
          <span className="stat-value">
            {netFlow >= 0 ? '+' : ''}
            {formatAmount(netFlow)}
          </span>
        </div>
      </div>

      {/* Control Header & Filters */}
      <div className="tx-control-header">
        <div className="tx-filter-row">
          <div className="filter-group">
            <label htmlFor="filter-type">Type</label>
            <select
              id="filter-type"
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value)}
              className="filter-select"
            >
              <option value="ALL">All Types</option>
              <option value="EXPENSE">Expense</option>
              <option value="INCOME">Income</option>
              <option value="TRANSFER">Transfer</option>
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="filter-account">Account</label>
            <select
              id="filter-account"
              value={selectedAccountId}
              onChange={(e) => setSelectedAccountId(e.target.value)}
              className="filter-select"
            >
              <option value="ALL">All Accounts</option>
              {accounts.map((acc) => (
                <option key={acc.id} value={acc.id}>
                  {acc.name}
                </option>
              ))}
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="filter-start-date">From</label>
            <input
              id="filter-start-date"
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="filter-date-input"
            />
          </div>

          <div className="filter-group">
            <label htmlFor="filter-end-date">To</label>
            <input
              id="filter-end-date"
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="filter-date-input"
            />
          </div>

          <button onClick={handleApplyFilter} className="btn-filter-apply" type="button">
            Filter
          </button>
        </div>

        <button onClick={handleOpenCreateModal} className="btn-record-tx" type="button">
          + Record Transaction
        </button>
      </div>

      {error && <div className="error-banner">{error}</div>}

      {/* Transactions List */}
      {loading ? (
        <div className="tx-loading-spinner">Loading transactions...</div>
      ) : transactions.length === 0 ? (
        <div className="tx-empty-state">
          <span className="empty-icon">💸</span>
          <h3>No transactions recorded</h3>
          <p>Start recording your daily income, expenses, or transfers between accounts.</p>
          <button onClick={handleOpenCreateModal} className="btn-record-tx" type="button">
            Record First Transaction
          </button>
        </div>
      ) : (
        <div className="tx-table-container">
          <table className="tx-table">
            <thead>
              <tr>
                <th>Date</th>
                <th>Type</th>
                <th>Account</th>
                <th>Category</th>
                <th>Description</th>
                <th className="th-amount">Amount</th>
                <th className="th-actions">Actions</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map((tx) => {
                let badgeClass = 'tx-type-expense';
                let sign = '-';
                if (tx.type === 'INCOME') {
                  badgeClass = 'tx-type-income';
                  sign = '+';
                } else if (tx.type === 'TRANSFER') {
                  badgeClass = 'tx-type-transfer';
                  sign = '⇄';
                }

                return (
                  <tr key={tx.id} className="tx-row">
                    <td className="td-date">{tx.transaction_date}</td>
                    <td>
                      <span className={`tx-type-badge ${badgeClass}`}>{tx.type}</span>
                    </td>
                    <td className="td-account">
                      {tx.type === 'TRANSFER' ? (
                        <span>
                          {tx.account_name} ➔ {tx.destination_account_name}
                        </span>
                      ) : (
                        tx.account_name
                      )}
                    </td>
                    <td className="td-category">
                      {tx.category_name ? (
                        <span className="category-pill">
                          {tx.category_color && (
                            <span
                              className="cat-dot"
                              style={{ backgroundColor: tx.category_color }}
                            />
                          )}
                          {tx.category_name}
                        </span>
                      ) : (
                        <span className="text-muted">—</span>
                      )}
                    </td>
                    <td className="td-desc">{tx.description || <span className="text-muted">—</span>}</td>
                    <td className={`td-amount ${tx.type.toLowerCase()}`}>
                      {sign} {tx.formatted_amount || formatRupiah(tx.amount)}
                    </td>
                    <td className="td-actions">
                      <button
                        onClick={() => handleOpenEditModal(tx)}
                        className="btn-action edit"
                        title="Edit"
                        type="button"
                      >
                        ✏️
                      </button>
                      <button
                        onClick={() => handleDelete(tx.id)}
                        className="btn-action delete"
                        title="Delete"
                        type="button"
                      >
                        🗑️
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Record/Edit Transaction Modal */}
      {isModalOpen && (
        <div className="modal-backdrop">
          <div className="modal-card">
            <div className="modal-header">
              <h3>{editingTx ? 'Edit Transaction' : 'Record Transaction'}</h3>
              <button onClick={handleCloseModal} className="btn-close" type="button">
                &times;
              </button>
            </div>

            <form onSubmit={handleFormSubmit} className="modal-form">
              {formError && <div className="form-error-banner">{formError}</div>}

              {/* Type Switcher Tabs */}
              <div className="form-type-selector">
                <button
                  type="button"
                  className={`type-tab-btn ${formType === 'EXPENSE' ? 'active expense' : ''}`}
                  onClick={() => {
                    setFormType('EXPENSE');
                    setFormCategoryId('');
                  }}
                >
                  Expense
                </button>
                <button
                  type="button"
                  className={`type-tab-btn ${formType === 'INCOME' ? 'active income' : ''}`}
                  onClick={() => {
                    setFormType('INCOME');
                    setFormCategoryId('');
                  }}
                >
                  Income
                </button>
                <button
                  type="button"
                  className={`type-tab-btn ${formType === 'TRANSFER' ? 'active transfer' : ''}`}
                  onClick={() => {
                    setFormType('TRANSFER');
                    setFormCategoryId('');
                  }}
                >
                  Transfer
                </button>
              </div>

              {/* Source Account */}
              <div className="form-group">
                <label htmlFor="tx-account">
                  {formType === 'TRANSFER' ? 'Source Account' : 'Account'} *
                </label>
                <select
                  id="tx-account"
                  value={formAccountId}
                  onChange={(e) => setFormAccountId(e.target.value)}
                  required
                >
                  <option value="" disabled>
                    Select Account
                  </option>
                  {accounts.map((acc) => (
                    <option key={acc.id} value={acc.id}>
                      {acc.name} ({formatRupiah(acc.current_balance)})
                    </option>
                  ))}
                </select>
              </div>

              {/* Destination Account (TRANSFER only) */}
              {formType === 'TRANSFER' && (
                <div className="form-group">
                  <label htmlFor="tx-dest-account">Destination Account *</label>
                  <select
                    id="tx-dest-account"
                    value={formDestAccountId}
                    onChange={(e) => setFormDestAccountId(e.target.value)}
                    required
                  >
                    <option value="" disabled>
                      Select Destination Account
                    </option>
                    {accounts
                      .filter((acc) => acc.id !== formAccountId)
                      .map((acc) => (
                        <option key={acc.id} value={acc.id}>
                          {acc.name} ({formatRupiah(acc.current_balance)})
                        </option>
                      ))}
                  </select>
                </div>
              )}

              {/* Category (EXPENSE & INCOME) */}
              {formType !== 'TRANSFER' && (
                <div className="form-group">
                  <label htmlFor="tx-category">Category (Optional)</label>
                  <select
                    id="tx-category"
                    value={formCategoryId}
                    onChange={(e) => setFormCategoryId(e.target.value)}
                  >
                    <option value="">No Category</option>
                    {availableCategories.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* Amount */}
              <div className="form-group">
                <label htmlFor="tx-amount">Amount (IDR) *</label>
                <input
                  id="tx-amount"
                  type="text"
                  placeholder="0,00"
                  value={formAmount}
                  onChange={(e) => setFormAmount(e.target.value)}
                  onBlur={() => setFormAmount(formatCurrencyInput(formAmount))}
                  required
                />
                <small className="help-text">Format: xxx.xxx,00 (contoh: 50.000,00)</small>
                {formAmount && parseRupiahToCents(formAmount) > 0 && (
                  <span className="balance-preview">
                    {formatRupiah(parseRupiahToCents(formAmount))}
                  </span>
                )}
              </div>

              {/* Transaction Date (allows backdating) */}
              <div className="form-group">
                <label htmlFor="tx-date">Transaction Date *</label>
                <input
                  id="tx-date"
                  type="date"
                  value={formDate}
                  onChange={(e) => setFormDate(e.target.value)}
                  required
                />
                <small className="help-text">
                  You may record backdated transactions to accurately capture past expenses or income.
                </small>
              </div>

              {/* Description */}
              <div className="form-group">
                <label htmlFor="tx-description">Description / Notes</label>
                <textarea
                  id="tx-description"
                  rows={2}
                  placeholder="What was this transaction for?"
                  value={formDescription}
                  onChange={(e) => setFormDescription(e.target.value)}
                />
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  onClick={handleCloseModal}
                  className="btn-cancel"
                  disabled={isSubmitting}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-submit"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? 'Saving...' : editingTx ? 'Update Transaction' : 'Save Transaction'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
