import React, { useEffect, useState } from 'react';
import * as accountsApi from '../api/accounts';
import type { Account, AccountType, AccountStatus } from '../api/accounts';
import { formatNumberIDR, parseRupiahToCents, formatCurrencyInput } from '../utils/currency';
import { usePrivacy } from '../context/PrivacyContext';
import { PrivacyToggle } from './PrivacyToggle';

export const AccountsManager: React.FC = () => {
  const { formatAmount } = usePrivacy();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');

  // Form states
  const [isFormOpen, setIsFormOpen] = useState<boolean>(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [name, setName] = useState<string>('');
  const [type, setType] = useState<AccountType>('BANK');
  const [openingBalance, setOpeningBalance] = useState<string>('0,00');
  const [status, setStatus] = useState<AccountStatus>('ACTIVE');
  const [formError, setFormError] = useState<string>('');

  const loadAccounts = async () => {
    setIsLoading(true);
    setError('');
    try {
      const data = await accountsApi.listAccounts();
      setAccounts(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load accounts');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadAccounts();
  }, []);

  const resetForm = () => {
    setEditingId(null);
    setName('');
    setType('BANK');
    setOpeningBalance('0,00');
    setStatus('ACTIVE');
    setFormError('');
    setIsFormOpen(false);
  };

  const handleOpenCreate = () => {
    resetForm();
    setIsFormOpen(true);
  };

  const handleOpenEdit = (acc: Account) => {
    setEditingId(acc.id);
    setName(acc.name);
    setType(acc.type);
    setOpeningBalance(formatNumberIDR(acc.opening_balance));
    setStatus(acc.status);
    setFormError('');
    setIsFormOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    const cleanName = name.trim();
    if (!cleanName) {
      setFormError('Account name is required');
      return;
    }

    const parsedBalance = parseRupiahToCents(openingBalance);
    if (isNaN(parsedBalance) || parsedBalance < 0) {
      setFormError('Opening balance must be 0 or a positive amount');
      return;
    }

    try {
      if (editingId) {
        await accountsApi.updateAccount(editingId, {
          name: cleanName,
          type,
          status,
        });
        setSuccess('Account updated successfully!');
      } else {
        await accountsApi.createAccount({
          name: cleanName,
          type,
          opening_balance: parsedBalance,
        });
        setSuccess('Account created successfully!');
      }
      resetForm();
      await loadAccounts();
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Operation failed');
    }
  };

  const handleDelete = async (id: string, accName: string) => {
    if (!window.confirm(`Are you sure you want to delete "${accName}"?`)) {
      return;
    }
    try {
      await accountsApi.deleteAccount(id);
      setSuccess(`Account "${accName}" deleted.`);
      await loadAccounts();
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete account');
    }
  };

  const getTypeIcon = (t: AccountType) => {
    switch (t) {
      case 'BANK': return '🏛️';
      case 'E_WALLET': return '📱';
      case 'CASH': return '💵';
      case 'INVESTMENT': return '📈';
      default: return '📦';
    }
  };

  const totalBalance = accounts
    .filter((a) => a.status === 'ACTIVE')
    .reduce((sum, a) => sum + Number(a.current_balance), 0);

  return (
    <div className="accounts-manager" id="accounts-manager">
      <div className="section-header">
        <div>
          <h3 className="section-title">🏛️ Financial Accounts</h3>
          <p className="section-subtitle">
            Manage your bank accounts, e-wallets, cash, and investment locations.
          </p>
        </div>
        <button
          className="btn btn--primary"
          id="btn-add-account"
          onClick={handleOpenCreate}
        >
          + Add Account
        </button>
      </div>

      {success && (
        <div className="auth-alert auth-alert--success" role="status">
          <span>✅ {success}</span>
        </div>
      )}

      {error && (
        <div className="auth-alert auth-alert--error" role="alert">
          <span>⚠️ {error}</span>
        </div>
      )}

      {/* Summary Total Banner */}
      <div className="balance-summary-banner">
        <div className="balance-summary-header">
          <span className="summary-label">Total Liquid Balance</span>
          <PrivacyToggle variant="icon" id="accounts-total-privacy-btn" />
        </div>
        <span className="summary-value" id="total-account-balance">{formatAmount(totalBalance)}</span>
        <span className="summary-subtext">{accounts.filter(a => a.status === 'ACTIVE').length} active accounts</span>
      </div>

      {/* Modal / Form */}
      {isFormOpen && (
        <div className="form-card" id="account-form-card">
          <h4 className="form-card__title">
            {editingId ? 'Edit Account' : 'Create New Account'}
          </h4>

          {formError && (
            <div className="auth-alert auth-alert--error">
              <span>⚠️ {formError}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} noValidate>
            <div className="form-row">
              <div className="form-group flex-1">
                <label htmlFor="acc-name">Account Name</label>
                <input
                  id="acc-name"
                  type="text"
                  placeholder="e.g. BCA Checking, GoPay"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                />
              </div>

              <div className="form-group flex-1">
                <label htmlFor="acc-type">Account Type</label>
                <select
                  id="acc-type"
                  className="form-select"
                  value={type}
                  onChange={(e) => setType(e.target.value as AccountType)}
                >
                  <option value="BANK">🏛️ Bank Account</option>
                  <option value="E_WALLET">📱 E-Wallet</option>
                  <option value="CASH">💵 Physical Cash</option>
                  <option value="INVESTMENT">📈 Investment Portfolio</option>
                  <option value="OTHER">📦 Other</option>
                </select>
              </div>
            </div>

            <div className="form-row">
              {!editingId && (
                <div className="form-group flex-1">
                  <label htmlFor="acc-opening-balance">Opening Balance</label>
                  <input
                    id="acc-opening-balance"
                    type="text"
                    placeholder="0,00"
                    value={openingBalance}
                    onChange={(e) => setOpeningBalance(e.target.value)}
                    onBlur={() => setOpeningBalance(formatCurrencyInput(openingBalance))}
                  />
                  <span className="form-hint">Format: xxx.xxx,00 (contoh: 100.000,00)</span>
                </div>
              )}

              {editingId && (
                <div className="form-group flex-1">
                  <label htmlFor="acc-status">Account Status</label>
                  <select
                    id="acc-status"
                    className="form-select"
                    value={status}
                    onChange={(e) => setStatus(e.target.value as AccountStatus)}
                  >
                    <option value="ACTIVE">Active</option>
                    <option value="ARCHIVED">Archived</option>
                  </select>
                </div>
              )}
            </div>

            <div className="form-actions">
              <button
                type="button"
                className="btn btn--outline"
                id="btn-cancel-account"
                onClick={resetForm}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn btn--primary"
                id="btn-submit-account"
              >
                {editingId ? 'Save Changes' : 'Create Account'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Account List */}
      {isLoading ? (
        <div className="app-loading">
          <div className="spinner" />
          <span>Loading accounts…</span>
        </div>
      ) : accounts.length === 0 ? (
        <div className="empty-state" id="accounts-empty-state">
          <span className="empty-icon">🏦</span>
          <p className="empty-title">No Accounts Configured Yet</p>
          <p className="empty-subtitle">
            Click "+ Add Account" to configure your first bank account or wallet.
          </p>
        </div>
      ) : (
        <div className="accounts-grid" id="accounts-list">
          {accounts.map((acc) => (
            <div
              key={acc.id}
              className={`account-card ${acc.status === 'ARCHIVED' ? 'account-card--archived' : ''}`}
              id={`account-card-${acc.id}`}
            >
              <div className="account-card__header">
                <div className="account-meta">
                  <span className="account-icon">{getTypeIcon(acc.type)}</span>
                  <div>
                    <h4 className="account-name">{acc.name}</h4>
                    <span className="account-type-tag">{acc.type}</span>
                  </div>
                </div>
                <span className={`status-badge status-badge--${acc.status.toLowerCase()}`}>
                  {acc.status}
                </span>
              </div>

              <div className="account-card__body">
                <div className="balance-field">
                  <span className="balance-label">Current Balance</span>
                  <span className="balance-amount">{formatAmount(acc.current_balance)}</span>
                </div>
                <div className="balance-field-small">
                  <span className="balance-sublabel">Opening: {formatAmount(acc.opening_balance)}</span>
                </div>
              </div>

              <div className="account-card__footer">
                <button
                  className="btn-action"
                  onClick={() => handleOpenEdit(acc)}
                  id={`btn-edit-account-${acc.id}`}
                >
                  ✏️ Edit
                </button>
                <button
                  className="btn-action btn-action--danger"
                  onClick={() => handleDelete(acc.id, acc.name)}
                  id={`btn-delete-account-${acc.id}`}
                >
                  🗑️ Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
