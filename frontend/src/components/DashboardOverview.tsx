import React, { useEffect, useState, useCallback } from 'react';
import { getDashboardData, type DashboardData } from '../api/dashboard';
import { usePrivacy } from '../context/PrivacyContext';
import { PrivacyToggle } from './PrivacyToggle';

interface DashboardOverviewProps {
  onNavigateTab?: (tab: string) => void;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr + 'T00:00:00Z');
  return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric', timeZone: 'UTC' });
}

export const DashboardOverview: React.FC<DashboardOverviewProps> = ({ onNavigateTab }) => {
  const { formatAmount } = usePrivacy();
  const fmtIDR = (amount: number) => formatAmount(amount);

  const [data, setData] = useState<DashboardData | null>(null);
  const [targetDate, setTargetDate] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

  const loadDashboard = useCallback(async (date?: string) => {
    setLoading(true);
    setError('');
    try {
      const res = await getDashboardData(date);
      setData(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard metrics');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadDashboard(targetDate || undefined);
  }, [loadDashboard, targetDate]);

  const handlePrevCycle = () => {
    if (!data) return;
    const start = new Date(data.cycle.start_date + 'T00:00:00Z');
    start.setUTCDate(start.getUTCDate() - 1);
    setTargetDate(start.toISOString().split('T')[0]);
  };

  const handleNextCycle = () => {
    if (!data) return;
    const end = new Date(data.cycle.end_date + 'T00:00:00Z');
    end.setUTCDate(end.getUTCDate() + 1);
    setTargetDate(end.toISOString().split('T')[0]);
  };

  const handleToday = () => {
    setTargetDate('');
  };

  if (loading && !data) {
    return (
      <section className="manager-card" id="dashboard-loading-card">
        <div className="health-loading">
          <div className="spinner" />
          <span>Synthesizing financial dashboard metrics…</span>
        </div>
      </section>
    );
  }

  if (error && !data) {
    return (
      <section className="manager-card">
        <div className="auth-alert auth-alert--error">
          <span>⚠️ {error}</span>
        </div>
        <button className="btn btn--primary btn--sm" onClick={() => loadDashboard(targetDate || undefined)}>
          Retry
        </button>
      </section>
    );
  }

  if (!data) return null;

  const cycleProgressPct = data.cycle.total_days > 0
    ? Math.min(Math.round((data.cycle.day_of_cycle / data.cycle.total_days) * 100), 100)
    : 0;

  const isGrowthPositive = data.asset_growth >= 0;

  return (
    <div className="dashboard-overview-container" id="dashboard-overview">
      {/* ── Cycle Navigation Header ───────────────────────────────── */}
      <section className="dashboard-cycle-panel" id="dashboard-cycle-header">
        <div className="dashboard-cycle-panel__left">
          <div className="dashboard-cycle-badge">
            <span className="dashboard-cycle-badge__icon">📅</span>
            <div>
              <span className="dashboard-cycle-badge__label">Financial Cycle</span>
              <h2 className="dashboard-cycle-badge__dates">
                {formatDate(data.cycle.start_date)} — {formatDate(data.cycle.end_date)}
              </h2>
            </div>
          </div>
          <div className="dashboard-cycle-track-group">
            <div className="dashboard-cycle-progress">
              <div
                className="dashboard-cycle-progress__bar"
                style={{ width: `${cycleProgressPct}%` }}
              />
            </div>
            <span className="dashboard-cycle-track-group__hint">
              Day {data.cycle.day_of_cycle} of {data.cycle.total_days} ({data.cycle.days_remaining} days remaining)
            </span>
          </div>
        </div>

        <div className="dashboard-cycle-panel__right">
          <button
            className="btn btn--outline btn--sm"
            onClick={handlePrevCycle}
            id="btn-prev-cycle"
            title="Previous cycle"
          >
            ← Previous
          </button>
          <button
            className="btn btn--outline btn--sm"
            onClick={handleToday}
            id="btn-today-cycle"
            title="Current cycle"
          >
            Current
          </button>
          <button
            className="btn btn--outline btn--sm"
            onClick={handleNextCycle}
            id="btn-next-cycle"
            title="Next cycle"
          >
            Next →
          </button>
        </div>
      </section>

      {/* ── Primary Financial KPI Grid ───────────────────────────── */}
      <section className="dashboard-kpi-grid" id="dashboard-kpi-grid">
        {/* Total Assets & Growth */}
        <div className="kpi-card kpi-card--highlight" id="kpi-total-assets">
          <div className="kpi-card__header">
            <div className="kpi-card__title-row">
              <span className="kpi-card__title">Total Net Assets</span>
              <PrivacyToggle variant="icon" id="dashboard-total-assets-privacy-btn" />
            </div>
            <span className="kpi-card__icon">🏛️</span>
          </div>
          <div className="kpi-card__value-row">
            <span className="kpi-card__main-value">{fmtIDR(data.total_assets)}</span>
            <span
              className={`kpi-badge ${isGrowthPositive ? 'kpi-badge--positive' : 'kpi-badge--negative'}`}
              id="kpi-asset-growth-badge"
              title="Asset growth compared to previous cycle end"
            >
              {isGrowthPositive ? '▲ +' : '▼ '}{fmtIDR(data.asset_growth)}
            </span>
          </div>
          <span className="kpi-card__footer-note">
            Previous Cycle Assets: <strong>{fmtIDR(data.previous_cycle_assets)}</strong>
          </span>
        </div>

        {/* Cycle Income */}
        <div className="kpi-card" id="kpi-cycle-income">
          <div className="kpi-card__header">
            <span className="kpi-card__title">Cycle Income</span>
            <span className="kpi-card__icon">📈</span>
          </div>
          <div className="kpi-card__value-row">
            <span className="kpi-card__main-value kpi-card__main-value--income">
              {fmtIDR(data.income)}
            </span>
          </div>
          <span className="kpi-card__footer-note">Total earned during cycle</span>
        </div>

        {/* Cycle Expenses */}
        <div className="kpi-card" id="kpi-cycle-expense">
          <div className="kpi-card__header">
            <span className="kpi-card__title">Cycle Spending</span>
            <span className="kpi-card__icon">📉</span>
          </div>
          <div className="kpi-card__value-row">
            <span className="kpi-card__main-value kpi-card__main-value--expense">
              {fmtIDR(data.expense)}
            </span>
          </div>
          <span className="kpi-card__footer-note">Total operational expenses</span>
        </div>

        {/* Net Savings */}
        <div className="kpi-card" id="kpi-net-savings">
          <div className="kpi-card__header">
            <span className="kpi-card__title">Net Operational Savings</span>
            <span className="kpi-card__icon">💎</span>
          </div>
          <div className="kpi-card__value-row">
            <span
              className={`kpi-card__main-value ${
                data.net_cash_flow >= 0 ? 'kpi-card__main-value--savings-pos' : 'kpi-card__main-value--savings-neg'
              }`}
            >
              {data.net_cash_flow >= 0 ? '+' : ''}{fmtIDR(data.net_cash_flow)}
            </span>
          </div>
          <span className="kpi-card__footer-note">Income minus operating expenses</span>
        </div>
      </section>

      {/* ── Two Column Layout: Budget Progress & Accounts ─────────── */}
      <div className="dashboard-content-columns">
        {/* Left Column: Budgeted vs Actual Spending */}
        <section className="dashboard-section-card" id="dashboard-budget-section">
          <div className="dashboard-section-card__header">
            <div>
              <h3 className="dashboard-section-card__title">📊 Budget Status</h3>
              <span className="dashboard-section-card__subtitle">
                Category spending limits and variances
              </span>
            </div>
            {onNavigateTab && (
              <button
                className="btn btn--outline btn--sm"
                onClick={() => onNavigateTab('budgets')}
                id="btn-goto-budgets"
              >
                View Budgets →
              </button>
            )}
          </div>

          {data.budgets.length === 0 ? (
            <div className="dashboard-empty-state">
              <span className="dashboard-empty-state__icon">🎯</span>
              <p>No budgets configured for this cycle.</p>
              {onNavigateTab && (
                <button
                  className="btn btn--primary btn--sm"
                  onClick={() => onNavigateTab('budgets')}
                >
                  + Add First Budget
                </button>
              )}
            </div>
          ) : (
            <div className="dashboard-budget-list">
              {data.budgets.map((b) => {
                const usedPct = b.planned > 0
                  ? Math.min(Math.round((b.actual / b.planned) * 100), 100)
                  : 0;

                return (
                  <div
                    key={b.category_id}
                    className={`dashboard-budget-item ${b.overspent ? 'dashboard-budget-item--over' : ''}`}
                    id={`budget-progress-${b.category_id}`}
                  >
                    <div className="dashboard-budget-item__row">
                      <div className="dashboard-budget-item__cat">
                        <span
                          className="budget-card__category-dot"
                          style={{ backgroundColor: b.category_color || '#6366F1' }}
                        />
                        <span className="dashboard-budget-item__name">{b.category_name}</span>
                      </div>
                      <div className="dashboard-budget-item__amounts">
                        <span className="dashboard-budget-item__actual">{fmtIDR(b.actual)}</span>
                        <span className="dashboard-budget-item__planned">/ {fmtIDR(b.planned)}</span>
                      </div>
                    </div>

                    <div className="dashboard-progress-track">
                      <div
                        className={`dashboard-progress-fill ${b.overspent ? 'dashboard-progress-fill--over' : ''}`}
                        style={{ width: `${usedPct}%` }}
                      />
                    </div>

                    <div className="dashboard-budget-item__footer">
                      <span className="dashboard-budget-item__pct">{usedPct}% consumed</span>
                      <span
                        className={`dashboard-budget-item__variance ${
                          b.overspent ? 'dashboard-budget-item__variance--over' : 'dashboard-budget-item__variance--under'
                        }`}
                      >
                        {b.overspent
                          ? `OVERSPENT by ${fmtIDR(Math.abs(b.variance))}`
                          : `${fmtIDR(b.variance)} remaining`}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>

        {/* Right Column: Account Balances Overview */}
        <section className="dashboard-section-card" id="dashboard-accounts-section">
          <div className="dashboard-section-card__header">
            <div>
              <h3 className="dashboard-section-card__title">🏛️ Account Overview</h3>
              <span className="dashboard-section-card__subtitle">
                Current balances across accounts
              </span>
            </div>
            {onNavigateTab && (
              <button
                className="btn btn--outline btn--sm"
                onClick={() => onNavigateTab('accounts')}
                id="btn-goto-accounts"
              >
                View Accounts →
              </button>
            )}
          </div>

          {data.accounts.length === 0 ? (
            <div className="dashboard-empty-state">
              <span className="dashboard-empty-state__icon">💳</span>
              <p>No active accounts configured.</p>
              {onNavigateTab && (
                <button
                  className="btn btn--primary btn--sm"
                  onClick={() => onNavigateTab('accounts')}
                >
                  + Add First Account
                </button>
              )}
            </div>
          ) : (
            <div className="dashboard-account-grid">
              {data.accounts.map((acc) => (
                <div key={acc.id} className="dashboard-account-tile" id={`account-tile-${acc.id}`}>
                  <div className="dashboard-account-tile__top">
                    <span className="dashboard-account-tile__name">{acc.name}</span>
                    <span className="account-type-pill">{acc.type}</span>
                  </div>
                  <div className="dashboard-account-tile__balance">
                    {fmtIDR(acc.balance)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
};
