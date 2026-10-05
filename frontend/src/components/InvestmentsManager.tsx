import React, { useState, useEffect } from 'react';
import {
  type Investment,
  type InvestmentType,
  type PortfolioSummary,
  listInvestments,
  createInvestment,
  updateValuation,
  deleteInvestment,
} from '../api/investments';
import type { Account } from '../api/accounts';
import { formatRupiah, formatNumberIDR, parseRupiahToCents, formatCurrencyInput } from '../utils/currency';
import { usePrivacy } from '../context/PrivacyContext';
import { PrivacyToggle } from './PrivacyToggle';

interface InvestmentsManagerProps {
  accounts: Account[];
}

export const InvestmentsManager: React.FC<InvestmentsManagerProps> = ({ accounts }) => {
  const { maskValue } = usePrivacy();
  const mask = (val: string): string => (maskValue ? maskValue(val) : val);
  const [portfolio, setPortfolio] = useState<PortfolioSummary>({
    total_capital: 0,
    total_current_value: 0,
    total_gain_loss: 0,
    total_gain_loss_percentage: 0,
    investments_count: 0,
    investments: [],
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);
  const [typeFilter, setTypeFilter] = useState<InvestmentType | 'ALL'>('ALL');
  const [isFormOpen, setIsFormOpen] = useState(false);

  // New Investment Form State
  const [invType, setInvType] = useState<InvestmentType>('STOCK');
  const [invName, setInvName] = useState('');
  const [invCapital, setInvCapital] = useState('');
  const [invCurrentValue, setInvCurrentValue] = useState('');
  const [sourceAccountId, setSourceAccountId] = useState('');
  const [invNotes, setInvNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // Valuation Modal State
  const [editingInv, setEditingInv] = useState<Investment | null>(null);
  const [newValuation, setNewValuation] = useState('');
  const [valNotes, setValNotes] = useState('');
  const [valSubmitting, setValSubmitting] = useState(false);

  const fetchPortfolio = async () => {
    try {
      setLoading(true);
      setError(null);
      const filter = typeFilter === 'ALL' ? undefined : typeFilter;
      const data = await listInvestments(filter);
      setPortfolio(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load investments');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPortfolio();
  }, [typeFilter]);

  const handleCreateInvestment = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccessMsg(null);

    const capitalNum = parseRupiahToCents(invCapital);
    if (isNaN(capitalNum) || capitalNum < 0) {
      setError('Capital must be a non-negative amount');
      return;
    }

    let currentValNum: number | undefined;
    if (invCurrentValue.trim() !== '') {
      currentValNum = parseRupiahToCents(invCurrentValue);
      if (isNaN(currentValNum) || currentValNum < 0) {
        setError('Current value must be a non-negative amount');
        return;
      }
    }

    try {
      setSubmitting(true);
      await createInvestment({
        type: invType,
        name: invName.trim(),
        capital: capitalNum,
        current_value: currentValNum,
        source_account_id: sourceAccountId || undefined,
        notes: invNotes.trim() || undefined,
      });

      setSuccessMsg(`Investment ${invName} created successfully`);
      setInvName('');
      setInvCapital('');
      setInvCurrentValue('');
      setSourceAccountId('');
      setInvNotes('');
      setIsFormOpen(false);
      await fetchPortfolio();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create investment');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdateValuation = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingInv) return;

    setError(null);
    setSuccessMsg(null);

    const valNum = parseRupiahToCents(newValuation);
    if (isNaN(valNum) || valNum < 0) {
      setError('Valuation must be a non-negative amount');
      return;
    }

    try {
      setValSubmitting(true);
      await updateValuation(editingInv.id, {
        current_value: valNum,
        notes: valNotes.trim() || undefined,
      });

      setSuccessMsg(`Valuation for ${editingInv.name} updated to ${formatRupiah(valNum)}`);
      setEditingInv(null);
      setNewValuation('');
      setValNotes('');
      await fetchPortfolio();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update valuation');
    } finally {
      setValSubmitting(false);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!window.confirm(`Are you sure you want to remove ${name} from your portfolio?`)) {
      return;
    }

    try {
      setError(null);
      await deleteInvestment(id);
      setSuccessMsg(`Investment ${name} deleted successfully`);
      await fetchPortfolio();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete investment');
    }
  };

  const getTypeIcon = (type: InvestmentType) => {
    switch (type) {
      case 'STOCK':
        return '📊 Stock';
      case 'MUTUAL_FUND':
        return '🧺 Mutual Fund';
      case 'GOLD':
        return '🪙 Gold';
      case 'CRYPTO':
        return '⚡ Crypto';
      case 'OTHER':
        return '💼 Other Asset';
      default:
        return type;
    }
  };

  return (
    <div className="investments-manager-container" data-testid="investments-manager">
      <div className="investments-header">
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <h2>📈 Investment Portfolio</h2>
            <PrivacyToggle variant="icon" id="investments-privacy-toggle" />
          </div>
          <p className="section-description">
            Track investments across stocks, mutual funds, gold, and crypto. Market performance affects total wealth without being treated as operational income.
          </p>
        </div>
        <button
          type="button"
          className="btn-primary"
          onClick={() => setIsFormOpen(!isFormOpen)}
          aria-expanded={isFormOpen}
        >
          {isFormOpen ? '✕ Cancel' : '+ Add Investment Holding'}
        </button>
      </div>

      {error && <div className="error-banner" role="alert">{error}</div>}
      {successMsg && <div className="success-banner" role="status">{successMsg}</div>}

      {/* Portfolio KPI Grid */}
      <div className="investments-kpi-grid">
        <div className="kpi-card">
          <span className="kpi-label">Total Current Valuation</span>
          <span className="kpi-value text-accent">{mask(portfolio.formatted_total_current_value || formatRupiah(portfolio.total_current_value))}</span>
          <span className="kpi-subtext">Current market value</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Invested Capital</span>
          <span className="kpi-value">{mask(portfolio.formatted_total_capital || formatRupiah(portfolio.total_capital))}</span>
          <span className="kpi-subtext">Total cost basis</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Unrealized Gain / Loss</span>
          <div className="gain-loss-row">
            <span
              className={`kpi-value ${
                portfolio.total_gain_loss > 0
                  ? 'text-success'
                  : portfolio.total_gain_loss < 0
                  ? 'text-danger'
                  : ''
              }`}
            >
              {mask(portfolio.formatted_total_gain_loss || formatRupiah(portfolio.total_gain_loss))}
            </span>
            <span
              className={`gain-chip ${
                portfolio.total_gain_loss >= 0 ? 'gain-positive' : 'gain-negative'
              }`}
            >
              {portfolio.total_gain_loss >= 0 ? '▲' : '▼'}{' '}
              {portfolio.total_gain_loss_percentage.toFixed(2)}%
            </span>
          </div>
          <span className="kpi-subtext">{portfolio.investments_count} active holdings</span>
        </div>
      </div>

      {/* Creation Form */}
      {isFormOpen && (
        <form className="investment-form card" onSubmit={handleCreateInvestment}>
          <h3>Add Investment Holding</h3>
          <div className="form-grid">
            <div className="form-group">
              <label htmlFor="inv-type">Asset Class *</label>
              <select
                id="inv-type"
                value={invType}
                onChange={(e) => setInvType(e.target.value as InvestmentType)}
              >
                <option value="STOCK">Stock (Saham)</option>
                <option value="MUTUAL_FUND">Mutual Fund (Reksadana)</option>
                <option value="GOLD">Gold (Emas)</option>
                <option value="CRYPTO">Crypto</option>
                <option value="OTHER">Other</option>
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="inv-name">Holding Name *</label>
              <input
                id="inv-name"
                type="text"
                required
                placeholder="e.g. BBCA, Bibit Reksadana Pasar Uang"
                value={invName}
                onChange={(e) => setInvName(e.target.value)}
              />
            </div>

            <div className="form-group">
              <label htmlFor="inv-capital">Capital Invested (IDR) *</label>
              <input
                id="inv-capital"
                type="text"
                required
                placeholder="0,00"
                value={invCapital}
                onChange={(e) => setInvCapital(e.target.value)}
                onBlur={() => setInvCapital(formatCurrencyInput(invCapital))}
              />
              <small className="form-hint">Format: xxx.xxx,00 (contoh: 10.000.000,00)</small>
            </div>

            <div className="form-group">
              <label htmlFor="inv-current-value">Current Valuation (IDR)</label>
              <input
                id="inv-current-value"
                type="text"
                placeholder="0,00"
                value={invCurrentValue}
                onChange={(e) => setInvCurrentValue(e.target.value)}
                onBlur={() => setInvCurrentValue(formatCurrencyInput(invCurrentValue))}
              />
              <small className="form-hint">Leave blank to initialize at capital cost basis.</small>
            </div>

            <div className="form-group">
              <label htmlFor="inv-source-account">Funded From Account (Optional)</label>
              <select
                id="inv-source-account"
                value={sourceAccountId}
                onChange={(e) => setSourceAccountId(e.target.value)}
              >
                <option value="">None (Existing portfolio / outside cash)</option>
                {accounts.map((acc) => (
                  <option key={acc.id} value={acc.id}>
                    {acc.name} (Balance: {formatRupiah(acc.current_balance)})
                  </option>
                ))}
              </select>
              <small className="form-hint">If selected, capital will be deducted from this account.</small>
            </div>

            <div className="form-group">
              <label htmlFor="inv-notes">Notes</label>
              <input
                id="inv-notes"
                type="text"
                placeholder="e.g. Stockbit, Bibit, Pluang portfolio"
                value={invNotes}
                onChange={(e) => setInvNotes(e.target.value)}
              />
            </div>
          </div>

          <div className="form-actions">
            <button type="submit" className="btn-primary" disabled={submitting}>
              {submitting ? 'Saving...' : 'Add Investment'}
            </button>
            <button type="button" className="btn-secondary" onClick={() => setIsFormOpen(false)}>
              Cancel
            </button>
          </div>
        </form>
      )}

      {/* Filter Tabs */}
      <div className="filter-tabs">
        {(['ALL', 'STOCK', 'MUTUAL_FUND', 'GOLD', 'CRYPTO', 'OTHER'] as const).map((tab) => (
          <button
            key={tab}
            type="button"
            className={`filter-tab ${typeFilter === tab ? 'active' : ''}`}
            onClick={() => setTypeFilter(tab)}
          >
            {tab.replace('_', ' ')}
          </button>
        ))}
      </div>

      {/* Investments List */}
      {loading ? (
        <div className="loading-state">Loading portfolio...</div>
      ) : portfolio.investments.length === 0 ? (
        <div className="empty-state card">
          <p>No investments found for this category.</p>
          <button type="button" className="btn-secondary" onClick={() => setIsFormOpen(true)}>
            Add your first holding
          </button>
        </div>
      ) : (
        <div className="investments-list">
          {portfolio.investments.map((inv) => {
            const isGain = inv.unrealized_gain >= 0;

            return (
              <div key={inv.id} className="investment-item card">
                <div className="investment-identity">
                  <span className="asset-type-badge">{getTypeIcon(inv.type)}</span>
                  <h4>{inv.name}</h4>
                  {inv.notes && <p className="investment-notes">{inv.notes}</p>}
                  {inv.source_account_name && (
                    <span className="funded-from">Funded from: {inv.source_account_name}</span>
                  )}
                </div>

                <div className="investment-metrics">
                  <div className="metric-box">
                    <span className="metric-label">Capital</span>
                    <span className="metric-value">{mask(inv.formatted_capital || formatRupiah(inv.capital))}</span>
                  </div>

                  <div className="metric-box">
                    <span className="metric-label">Current Value</span>
                    <span className="metric-value font-bold">{mask(inv.formatted_current_value || formatRupiah(inv.current_value))}</span>
                  </div>

                  <div className="metric-box">
                    <span className="metric-label">Gain / Loss</span>
                    <span className={`metric-value ${isGain ? 'text-success' : 'text-danger'}`}>
                      {mask(inv.formatted_unrealized_gain || formatRupiah(inv.unrealized_gain))}
                    </span>
                    <span className={`gain-chip-sm ${isGain ? 'gain-positive' : 'gain-negative'}`}>
                      {isGain ? '▲' : '▼'} {inv.unrealized_gain_percentage.toFixed(2)}%
                    </span>
                  </div>
                </div>

                <div className="investment-actions">
                  <button
                    type="button"
                    className="btn-sm btn-primary"
                    onClick={() => {
                      setEditingInv(inv);
                      setNewValuation(formatNumberIDR(inv.current_value));
                      setValNotes(inv.notes || '');
                    }}
                  >
                    Update Valuation
                  </button>
                  <button
                    type="button"
                    className="btn-sm btn-danger"
                    onClick={() => handleDelete(inv.id, inv.name)}
                    title="Remove from portfolio"
                  >
                    🗑
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Valuation Modal */}
      {editingInv && (
        <div className="modal-backdrop" role="dialog" aria-modal="true">
          <div className="modal-card">
            <h3>Update Valuation for {editingInv.name}</h3>
            <p className="modal-subtitle">
              Capital Invested: <strong>{editingInv.formatted_capital || formatRupiah(editingInv.capital)}</strong>
            </p>

            <form onSubmit={handleUpdateValuation}>
              <div className="form-group">
                <label htmlFor="val-current-value">Latest Market Value (IDR) *</label>
                <input
                  id="val-current-value"
                  type="text"
                  required
                  placeholder="0,00"
                  value={newValuation}
                  onChange={(e) => setNewValuation(e.target.value)}
                  onBlur={() => setNewValuation(formatCurrencyInput(newValuation))}
                />
                <small className="form-hint">Format: xxx.xxx,00 (contoh: 12.000.000,00)</small>
              </div>

              {/* Dynamic gain preview */}
              {newValuation && !isNaN(parseRupiahToCents(newValuation)) && (
                <div className="valuation-preview">
                  <span>Estimated Unrealized Gain/Loss:</span>
                  <strong
                    className={
                      parseRupiahToCents(newValuation) - editingInv.capital >= 0
                        ? 'text-success'
                        : 'text-danger'
                    }
                  >
                    {formatRupiah(parseRupiahToCents(newValuation) - editingInv.capital)} (
                    {editingInv.capital > 0
                      ? (
                          ((parseRupiahToCents(newValuation) - editingInv.capital) /
                            editingInv.capital) *
                          100
                        ).toFixed(2)
                      : '0.00'}
                    %)
                  </strong>
                </div>
              )}

              <div className="form-group">
                <label htmlFor="val-notes">Valuation Notes</label>
                <input
                  id="val-notes"
                  type="text"
                  placeholder="e.g. October market valuation update"
                  value={valNotes}
                  onChange={(e) => setValNotes(e.target.value)}
                />
              </div>

              <div className="form-actions">
                <button type="submit" className="btn-primary" disabled={valSubmitting}>
                  {valSubmitting ? 'Saving...' : 'Save Valuation'}
                </button>
                <button type="button" className="btn-secondary" onClick={() => setEditingInv(null)}>
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
