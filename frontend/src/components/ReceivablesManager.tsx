import React, { useState, useEffect } from 'react';
import {
  type Receivable,
  type ReceivableStatus,
  listReceivables,
  createReceivable,
  recordPayment,
  updateReceivableStatus,
  deleteReceivable,
} from '../api/receivables';
import type { Account } from '../api/accounts';
import { formatRupiah, formatNumberIDR, parseRupiahToCents, formatCurrencyInput } from '../utils/currency';
import { usePrivacy } from '../context/PrivacyContext';
import { PrivacyToggle } from './PrivacyToggle';

interface ReceivablesManagerProps {
  accounts: Account[];
}

export const ReceivablesManager: React.FC<ReceivablesManagerProps> = ({ accounts }) => {
  const { formatAmount, maskValue } = usePrivacy();
  const fmtIDR = (amount: number): string => (formatAmount ? formatAmount(amount) : formatRupiah(amount));
  const mask = (val: string): string => (maskValue ? maskValue(val) : val);
  const [receivables, setReceivables] = useState<Receivable[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<ReceivableStatus | 'ALL'>('ALL');
  const [isFormOpen, setIsFormOpen] = useState(false);

  // New Receivable Form State
  const [counterparty, setCounterparty] = useState('');
  const [principal, setPrincipal] = useState('');
  const [sourceAccountId, setSourceAccountId] = useState('');
  const [dueDate, setDueDate] = useState('');
  const [notes, setNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // Payment Modal State
  const [payingReceivable, setPayingReceivable] = useState<Receivable | null>(null);
  const [paymentAmount, setPaymentAmount] = useState('');
  const [targetAccountId, setTargetAccountId] = useState('');
  const [paymentDate, setPaymentDate] = useState(new Date().toISOString().split('T')[0]);
  const [paymentNotes, setPaymentNotes] = useState('');
  const [payingSubmitting, setPayingSubmitting] = useState(false);

  // Expanded details toggle
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const fetchReceivables = async () => {
    try {
      setLoading(true);
      setError(null);
      const filter = statusFilter === 'ALL' ? undefined : statusFilter;
      const data = await listReceivables(filter);
      setReceivables(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load receivables');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchReceivables();
  }, [statusFilter]);

  const handleCreateReceivable = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccessMsg(null);

    const principalNum = parseRupiahToCents(principal);
    if (isNaN(principalNum) || principalNum <= 0) {
      setError('Principal must be a positive amount');
      return;
    }

    try {
      setSubmitting(true);
      await createReceivable({
        counterparty: counterparty.trim(),
        principal: principalNum,
        source_account_id: sourceAccountId || undefined,
        due_date: dueDate || undefined,
        notes: notes.trim() || undefined,
      });

      setSuccessMsg(`Receivable for ${counterparty} created successfully`);
      setCounterparty('');
      setPrincipal('');
      setSourceAccountId('');
      setDueDate('');
      setNotes('');
      setIsFormOpen(false);
      await fetchReceivables();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create receivable');
    } finally {
      setSubmitting(false);
    }
  };

  const handleRecordPayment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!payingReceivable) return;

    setError(null);
    setSuccessMsg(null);

    const amountNum = parseRupiahToCents(paymentAmount);
    if (isNaN(amountNum) || amountNum <= 0) {
      setError('Payment amount must be greater than zero');
      return;
    }

    if (amountNum > payingReceivable.remaining_amount) {
      setError(`Payment cannot exceed remaining principal of ${payingReceivable.formatted_remaining_amount || formatRupiah(payingReceivable.remaining_amount)}`);
      return;
    }

    if (!targetAccountId) {
      setError('Please select a target account to deposit the repayment');
      return;
    }

    try {
      setPayingSubmitting(true);
      await recordPayment(payingReceivable.id, {
        amount: amountNum,
        target_account_id: targetAccountId,
        payment_date: paymentDate,
        notes: paymentNotes.trim() || undefined,
      });

      setSuccessMsg(`Repayment of ${formatRupiah(amountNum)} recorded successfully`);
      setPayingReceivable(null);
      setPaymentAmount('');
      setPaymentNotes('');
      await fetchReceivables();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to record payment');
    } finally {
      setPayingSubmitting(false);
    }
  };

  const handleWriteOff = async (id: string, name: string) => {
    if (!window.confirm(`Are you sure you want to write off the remaining debt from ${name}? This will mark it as a non-recoverable loss.`)) {
      return;
    }

    try {
      setError(null);
      await updateReceivableStatus(id, {
        status: 'WRITTEN_OFF',
        notes: 'Written off debt',
      });
      setSuccessMsg(`Receivable for ${name} has been written off`);
      await fetchReceivables();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to write off receivable');
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!window.confirm(`Are you sure you want to delete receivable for ${name}?`)) {
      return;
    }

    try {
      setError(null);
      await deleteReceivable(id);
      setSuccessMsg(`Receivable for ${name} deleted successfully`);
      await fetchReceivables();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete receivable');
    }
  };

  // Portfolio metrics
  const totalPrincipal = receivables.reduce((sum, r) => sum + r.principal, 0);
  const totalPaid = receivables.reduce((sum, r) => sum + r.total_paid, 0);
  const activeRemaining = receivables
    .filter((r) => r.status !== 'PAID' && r.status !== 'WRITTEN_OFF')
    .reduce((sum, r) => sum + r.remaining_amount, 0);

  const getStatusBadge = (status: ReceivableStatus) => {
    switch (status) {
      case 'ACTIVE':
        return <span className="status-badge status-active">Active</span>;
      case 'PARTIALLY_PAID':
        return <span className="status-badge status-partial">Partially Paid</span>;
      case 'PAID':
        return <span className="status-badge status-paid">Fully Settled</span>;
      case 'OVERDUE':
        return <span className="status-badge status-overdue">Overdue</span>;
      case 'WRITTEN_OFF':
        return <span className="status-badge status-written-off">Written Off</span>;
      default:
        return <span className="status-badge">{status}</span>;
    }
  };

  return (
    <div className="receivables-manager-container" data-testid="receivables-manager">
      <div className="receivables-header">
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <h2>🤝 Receivables & Loans</h2>
            <PrivacyToggle variant="icon" id="receivables-privacy-toggle" />
          </div>
          <p className="section-description">
            Track money lent to friends, family, or counterparties. Lending converts cash into a receivable asset without reducing wealth.
          </p>
        </div>
        <button
          type="button"
          className="btn-primary"
          onClick={() => setIsFormOpen(!isFormOpen)}
          aria-expanded={isFormOpen}
        >
          {isFormOpen ? '✕ Cancel' : '+ Lend Money / Add Receivable'}
        </button>
      </div>

      {error && <div className="error-banner" role="alert">{error}</div>}
      {successMsg && <div className="success-banner" role="status">{successMsg}</div>}

      {/* KPI Cards */}
      <div className="receivables-kpi-grid">
        <div className="kpi-card">
          <span className="kpi-label">Active Outstanding</span>
          <span className="kpi-value text-accent">{fmtIDR(activeRemaining)}</span>
          <span className="kpi-subtext">Collectible assets</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Total Repaid</span>
          <span className="kpi-value text-success">{fmtIDR(totalPaid)}</span>
          <span className="kpi-subtext">Collected back into accounts</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Total Lent (Filtered)</span>
          <span className="kpi-value">{fmtIDR(totalPrincipal)}</span>
          <span className="kpi-subtext">{receivables.length} total loans</span>
        </div>
      </div>

      {/* Creation Form */}
      {isFormOpen && (
        <form className="receivable-form card" onSubmit={handleCreateReceivable}>
          <h3>Record New Loan / Receivable</h3>
          <div className="form-grid">
            <div className="form-group">
              <label htmlFor="rec-counterparty">Counterparty / Borrower *</label>
              <input
                id="rec-counterparty"
                type="text"
                required
                placeholder="e.g. Budi, John Doe"
                value={counterparty}
                onChange={(e) => setCounterparty(e.target.value)}
              />
            </div>

            <div className="form-group">
              <label htmlFor="rec-principal">Principal Amount (IDR) *</label>
              <input
                id="rec-principal"
                type="text"
                required
                placeholder="0,00"
                value={principal}
                onChange={(e) => setPrincipal(e.target.value)}
                onBlur={() => setPrincipal(formatCurrencyInput(principal))}
              />
              <small className="form-hint">Format: xxx.xxx,00 (contoh: 2.000.000,00)</small>
            </div>

            <div className="form-group">
              <label htmlFor="rec-source-account">Funded From Account (Optional)</label>
              <select
                id="rec-source-account"
                value={sourceAccountId}
                onChange={(e) => setSourceAccountId(e.target.value)}
              >
                <option value="">None (Existing loan / outside tracking)</option>
                {accounts.map((acc) => (
                  <option key={acc.id} value={acc.id}>
                    {acc.name} (Balance: {formatRupiah(acc.current_balance)})
                  </option>
                ))}
              </select>
              <small className="form-hint">If selected, principal will be deducted from this account.</small>
            </div>

            <div className="form-group">
              <label htmlFor="rec-due-date">Due Date (Optional)</label>
              <input
                id="rec-due-date"
                type="date"
                value={dueDate}
                onChange={(e) => setDueDate(e.target.value)}
              />
            </div>

            <div className="form-group full-width">
              <label htmlFor="rec-notes">Notes</label>
              <input
                id="rec-notes"
                type="text"
                placeholder="e.g. Emergency medical help, due in November"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </div>
          </div>

          <div className="form-actions">
            <button type="submit" className="btn-primary" disabled={submitting}>
              {submitting ? 'Creating...' : 'Save Receivable'}
            </button>
            <button type="button" className="btn-secondary" onClick={() => setIsFormOpen(false)}>
              Cancel
            </button>
          </div>
        </form>
      )}

      {/* Filter Tabs */}
      <div className="filter-tabs">
        {(['ALL', 'ACTIVE', 'PARTIALLY_PAID', 'PAID', 'WRITTEN_OFF'] as const).map((tab) => (
          <button
            key={tab}
            type="button"
            className={`filter-tab ${statusFilter === tab ? 'active' : ''}`}
            onClick={() => setStatusFilter(tab)}
          >
            {tab.replace('_', ' ')}
          </button>
        ))}
      </div>

      {/* Receivables List */}
      {loading ? (
        <div className="loading-state">Loading receivables...</div>
      ) : receivables.length === 0 ? (
        <div className="empty-state card">
          <p>No receivables found for this filter.</p>
          <button type="button" className="btn-secondary" onClick={() => setIsFormOpen(true)}>
            Record your first loan
          </button>
        </div>
      ) : (
        <div className="receivables-list">
          {receivables.map((rec) => {
            const pctPaid = rec.principal > 0 ? Math.min(100, Math.round((rec.total_paid / rec.principal) * 100)) : 0;
            const isExpanded = expandedId === rec.id;

            return (
              <div key={rec.id} className="receivable-item card">
                <div className="receivable-main-row">
                  <div className="receivable-identity">
                    <div className="identity-header">
                      <h4>{rec.counterparty}</h4>
                      {getStatusBadge(rec.status)}
                    </div>
                    {rec.notes && <p className="receivable-notes">{rec.notes}</p>}
                    <div className="receivable-meta">
                      {rec.source_account_name && (
                        <span>Lent from: <strong>{rec.source_account_name}</strong></span>
                      )}
                      {rec.due_date && (
                        <span>Due: <strong>{rec.due_date}</strong></span>
                      )}
                    </div>
                  </div>

                  <div className="receivable-amounts">
                    <div className="amount-col">
                      <span className="amount-label">Principal</span>
                      <span className="amount-value">{mask(rec.formatted_principal || formatRupiah(rec.principal))}</span>
                    </div>
                    <div className="amount-col">
                      <span className="amount-label">Remaining</span>
                      <span className={`amount-value ${rec.remaining_amount > 0 ? 'text-accent' : 'text-muted'}`}>
                        {mask(rec.formatted_remaining_amount || formatRupiah(rec.remaining_amount))}
                      </span>
                    </div>
                  </div>

                  <div className="receivable-actions">
                    {rec.status !== 'PAID' && rec.status !== 'WRITTEN_OFF' && (
                      <>
                        <button
                          type="button"
                          className="btn-sm btn-primary"
                          onClick={() => {
                            setPayingReceivable(rec);
                            setPaymentAmount(formatNumberIDR(rec.remaining_amount));
                            if (accounts.length > 0) setTargetAccountId(accounts[0].id);
                          }}
                        >
                          Record Repayment
                        </button>
                        <button
                          type="button"
                          className="btn-sm btn-danger-outline"
                          onClick={() => handleWriteOff(rec.id, rec.counterparty)}
                        >
                          Write Off
                        </button>
                      </>
                    )}
                    <button
                      type="button"
                      className="btn-sm btn-ghost"
                      onClick={() => setExpandedId(isExpanded ? null : rec.id)}
                    >
                      {isExpanded ? 'Hide History' : `History (${rec.payments?.length || 0})`}
                    </button>
                    <button
                      type="button"
                      className="btn-sm btn-danger"
                      onClick={() => handleDelete(rec.id, rec.counterparty)}
                      title="Delete receivable"
                    >
                      🗑
                    </button>
                  </div>
                </div>

                {/* Progress bar */}
                <div className="receivable-progress-bar-container">
                  <div
                    className="receivable-progress-bar"
                    style={{ width: `${pctPaid}%` }}
                    aria-valuenow={pctPaid}
                    aria-valuemin={0}
                    aria-valuemax={100}
                  />
                  <span className="progress-text">{pctPaid}% repaid ({mask(rec.formatted_total_paid || formatRupiah(rec.total_paid))})</span>
                </div>

                {/* Expanded Payment History */}
                {isExpanded && (
                  <div className="payment-history-panel">
                    <h5>Repayment History</h5>
                    {rec.payments && rec.payments.length > 0 ? (
                      <table className="payments-table">
                        <thead>
                          <tr>
                            <th>Date</th>
                            <th>Amount</th>
                            <th>Deposited Into</th>
                            <th>Notes</th>
                          </tr>
                        </thead>
                        <tbody>
                          {rec.payments.map((p) => (
                            <tr key={p.id}>
                              <td>{p.payment_date}</td>
                              <td className="text-success">+{mask(p.formatted_amount || formatRupiah(p.amount))}</td>
                              <td>{p.target_account_name || 'Account'}</td>
                              <td>{p.notes || '-'}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    ) : (
                      <p className="empty-subtext">No repayments recorded yet.</p>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Record Repayment Modal */}
      {payingReceivable && (
        <div className="modal-backdrop" role="dialog" aria-modal="true">
          <div className="modal-card">
            <h3>Record Repayment from {payingReceivable.counterparty}</h3>
            <p className="modal-subtitle">
              Remaining debt: <strong>Rp {payingReceivable.remaining_amount.toLocaleString()}</strong>
            </p>

            <form onSubmit={handleRecordPayment}>
              <div className="form-group">
                <label htmlFor="pay-amount">Amount Repaid (IDR) *</label>
                <input
                  id="pay-amount"
                  type="text"
                  required
                  placeholder="0,00"
                  value={paymentAmount}
                  onChange={(e) => setPaymentAmount(e.target.value)}
                  onBlur={() => setPaymentAmount(formatCurrencyInput(paymentAmount))}
                />
                <small className="form-hint">Format: xxx.xxx,00 (contoh: 500.000,00)</small>
              </div>

              <div className="form-group">
                <label htmlFor="pay-target-account">Deposit Repayment Into Account *</label>
                <select
                  id="pay-target-account"
                  required
                  value={targetAccountId}
                  onChange={(e) => setTargetAccountId(e.target.value)}
                >
                  <option value="">Select Account</option>
                  {accounts.map((acc) => (
                    <option key={acc.id} value={acc.id}>
                      {acc.name} (Balance: {formatRupiah(acc.current_balance)})
                    </option>
                  ))}
                </select>
                <small className="form-hint">Cash will be added to this account without creating artificial income.</small>
              </div>

              <div className="form-group">
                <label htmlFor="pay-date">Payment Date *</label>
                <input
                  id="pay-date"
                  type="date"
                  required
                  value={paymentDate}
                  onChange={(e) => setPaymentDate(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label htmlFor="pay-notes">Notes</label>
                <input
                  id="pay-notes"
                  type="text"
                  placeholder="e.g. Bank transfer, installment #1"
                  value={paymentNotes}
                  onChange={(e) => setPaymentNotes(e.target.value)}
                />
              </div>

              <div className="form-actions">
                <button type="submit" className="btn-primary" disabled={payingSubmitting}>
                  {payingSubmitting ? 'Recording...' : 'Confirm Repayment'}
                </button>
                <button type="button" className="btn-secondary" onClick={() => setPayingReceivable(null)}>
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
