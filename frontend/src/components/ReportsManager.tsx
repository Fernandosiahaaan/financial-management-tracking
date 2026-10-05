import React, { useEffect, useState, useCallback } from 'react';
import {
  fetchMonthlyReport,
  fetchCycleReport,
  type ReportData,
} from '../api/reports';
import { formatRupiah } from '../utils/currency';
import { usePrivacy } from '../context/PrivacyContext';
import { PrivacyToggle } from './PrivacyToggle';

export const ReportsManager: React.FC = () => {
  const { formatAmount } = usePrivacy();
  const [reportType, setReportType] = useState<'monthly' | 'cycle'>('monthly');

  // Month selector defaults to current month (YYYY-MM)
  const now = new Date();
  const currentMonthStr = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
  const currentDateStr = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;

  const [selectedMonth, setSelectedMonth] = useState<string>(currentMonthStr);
  const [selectedCycleDate, setSelectedCycleDate] = useState<string>(currentDateStr);

  const [report, setReport] = useState<ReportData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

  // JSON Export modal
  const [showJsonModal, setShowJsonModal] = useState<boolean>(false);
  const [copySuccess, setCopySuccess] = useState<boolean>(false);

  const loadReport = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      if (reportType === 'monthly') {
        const [yearStr, monthStr] = selectedMonth.split('-');
        const data = await fetchMonthlyReport(parseInt(yearStr, 10), parseInt(monthStr, 10));
        setReport(data);
      } else {
        const data = await fetchCycleReport(selectedCycleDate);
        setReport(data);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate financial report');
    } finally {
      setLoading(false);
    }
  }, [reportType, selectedMonth, selectedCycleDate]);

  useEffect(() => {
    loadReport();
  }, [loadReport]);

  const handleCopyJson = async () => {
    if (!report) return;
    try {
      await navigator.clipboard.writeText(JSON.stringify(report, null, 2));
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch {
      // Fallback
    }
  };

  const formatCurrency = (val: number): string => {
    return formatAmount ? formatAmount(val) : formatRupiah(val);
  };

  return (
    <div className="reports-manager" id="reports-manager-view">
      {/* Top Header & Type Switcher */}
      <div className="manager-header">
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <h2 className="manager-title">📑 Reports & Asset Snapshots</h2>
            <PrivacyToggle variant="icon" id="reports-privacy-toggle" />
          </div>
          <p className="manager-subtitle">
            Comprehensive financial statements and consolidated asset valuations
          </p>
        </div>

        <div className="reports-top-actions">
          <div className="segmented-control" role="tablist">
            <button
              type="button"
              id="report-tab-monthly"
              role="tab"
              aria-selected={reportType === 'monthly'}
              className={`segmented-btn ${reportType === 'monthly' ? 'segmented-btn--active' : ''}`}
              onClick={() => setReportType('monthly')}
            >
              📅 Monthly Report
            </button>
            <button
              type="button"
              id="report-tab-cycle"
              data-testid="report-tab-cycle"
              role="tab"
              aria-selected={reportType === 'cycle'}
              className={`segmented-btn ${reportType === 'cycle' ? 'segmented-btn--active' : ''}`}
              onClick={() => setReportType('cycle')}
            >
              🔄 Cycle Report
            </button>
          </div>

          <button
            id="report-export-json-btn"
            className="btn btn--outline btn--sm"
            onClick={() => setShowJsonModal(true)}
            disabled={!report || loading}
          >
            ⬇️ View JSON
          </button>
        </div>
      </div>

      {/* Date Filter Bar */}
      <div className="reports-filter-bar card">
        <div className="filter-group">
          {reportType === 'monthly' ? (
            <>
              <label htmlFor="report-month-input" className="form-label">
                Select Calendar Month
              </label>
              <input
                id="report-month-input"
                type="month"
                className="form-input"
                value={selectedMonth}
                onChange={(e) => setSelectedMonth(e.target.value)}
                min="2000-01"
                max="2100-12"
              />
            </>
          ) : (
            <>
              <label htmlFor="report-cycle-date-input" className="form-label">
                Select Date in Cycle
              </label>
              <input
                id="report-cycle-date-input"
                type="date"
                className="form-input"
                value={selectedCycleDate}
                onChange={(e) => setSelectedCycleDate(e.target.value)}
              />
            </>
          )}
        </div>

        <button
          id="report-generate-btn"
          className="btn btn--primary"
          onClick={loadReport}
          disabled={loading}
        >
          {loading ? 'Generating…' : '🔄 Refresh Report'}
        </button>
      </div>

      {/* Error state */}
      {error && (
        <div className="auth-alert auth-alert--error" role="alert">
          <span>⚠️ {error}</span>
          <button className="btn btn--outline btn--sm" onClick={loadReport} style={{ marginLeft: '1rem' }}>
            Retry
          </button>
        </div>
      )}

      {/* Loading state */}
      {loading && (
        <div className="report-loading-state">
          <div className="spinner" />
          <p>Compiling financial statements and asset snapshot…</p>
        </div>
      )}

      {/* Main Report View */}
      {!loading && !error && report && (
        <div className="report-content">
          {/* Period Banner */}
          <div className="report-banner card">
            <div className="report-banner__header">
              <span className="badge badge--primary">
                {reportType === 'monthly' ? '📅 Monthly Statement' : '🔄 Financial Cycle Statement'}
              </span>
              <span className="report-banner__dates">
                Period: <strong>{report.period}</strong> ({report.start_date} to {report.end_date})
              </span>
            </div>
            <div className="report-banner__networth">
              <div>
                <span className="report-stat-label">Consolidated Net Worth</span>
                <h3 className="report-stat-networth">{formatCurrency(report.net_worth)}</h3>
              </div>
              <div className="report-banner__cashflow">
                <span className="report-stat-label">Period Net Cash Flow</span>
                <span
                  className={`report-cashflow-tag ${
                    report.net_cash_flow >= 0 ? 'cashflow--positive' : 'cashflow--negative'
                  }`}
                >
                  {report.net_cash_flow >= 0 ? '+' : ''}
                  {formatCurrency(report.net_cash_flow)}
                </span>
              </div>
            </div>
          </div>

          {/* 1. Cash Flow Statement */}
          <section className="report-section">
            <h3 className="section-title">💵 Cash Flow Statement</h3>
            <div className="kpi-grid">
              <div className="kpi-card" id="kpi-report-income">
                <span className="kpi-card__label">Total Income</span>
                <div className="kpi-card__value text-success">{formatCurrency(report.income)}</div>
                <span className="kpi-card__subtext">Cash inflows in period</span>
              </div>

              <div className="kpi-card" id="kpi-report-expense">
                <span className="kpi-card__label">Total Expenses</span>
                <div className="kpi-card__value text-danger">{formatCurrency(report.expense)}</div>
                <span className="kpi-card__subtext">Operational spending</span>
              </div>

              <div className="kpi-card" id="kpi-report-transfer">
                <span className="kpi-card__label">Account Transfers</span>
                <div className="kpi-card__value text-muted">{formatCurrency(report.transfer)}</div>
                <span className="kpi-card__subtext">Asset shifts (non-cashflow)</span>
              </div>

              <div className="kpi-card" id="kpi-report-net-cashflow">
                <span className="kpi-card__label">Net Cash Flow</span>
                <div
                  className={`kpi-card__value ${
                    report.net_cash_flow >= 0 ? 'text-success' : 'text-danger'
                  }`}
                >
                  {report.net_cash_flow >= 0 ? '+' : ''}
                  {formatCurrency(report.net_cash_flow)}
                </div>
                <span className="kpi-card__subtext">Income minus Expense</span>
              </div>
            </div>
          </section>

          {/* 2. Budget vs Actual Comparison */}
          <section className="report-section">
            <h3 className="section-title">📊 Budget vs. Actual Spending</h3>
            {report.budget_vs_actual.length === 0 ? (
              <div className="empty-card card">
                <p>No budgets or expense records recorded for this period.</p>
              </div>
            ) : (
              <div className="table-responsive card">
                <table className="report-table" id="report-budget-table">
                  <thead>
                    <tr>
                      <th>Category</th>
                      <th className="text-right">Planned Budget</th>
                      <th className="text-right">Actual Spending</th>
                      <th className="text-right">Variance</th>
                      <th className="text-center">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {report.budget_vs_actual.map((item, idx) => {
                      const isOverspent = item.variance < 0;
                      return (
                        <tr key={item.category_id || idx}>
                          <td className="table-cell-category">
                            {item.category_icon && <span className="cat-icon">{item.category_icon}</span>}
                            <strong>{item.category}</strong>
                          </td>
                          <td className="text-right">{formatCurrency(item.budget)}</td>
                          <td className="text-right">{formatCurrency(item.actual)}</td>
                          <td
                            className={`text-right font-medium ${
                              isOverspent ? 'text-danger' : 'text-success'
                            }`}
                          >
                            {item.variance >= 0 ? '+' : ''}
                            {formatCurrency(item.variance)}
                          </td>
                          <td className="text-center">
                            <span
                              className={`badge ${
                                isOverspent ? 'badge--danger' : 'badge--success'
                              }`}
                            >
                              {isOverspent ? 'Overspent' : 'On Track'}
                            </span>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          {/* 3. Asset Snapshot & Net Worth */}
          <section className="report-section">
            <h3 className="section-title">🏛️ Asset Snapshot & Net Worth</h3>
            <div className="kpi-grid">
              <div className="kpi-card" id="kpi-report-asset-accounts">
                <span className="kpi-card__label">Cash & Bank Accounts</span>
                <div className="kpi-card__value">{formatCurrency(report.assets.accounts)}</div>
                <span className="kpi-card__subtext">
                  {report.assets.account_list?.length || 0} active accounts
                </span>
              </div>

              <div className="kpi-card" id="kpi-report-asset-receivables">
                <span className="kpi-card__label">Outstanding Receivables</span>
                <div className="kpi-card__value">{formatCurrency(report.assets.receivables)}</div>
                <span className="kpi-card__subtext">
                  {report.assets.receivable_list?.length || 0} active loans
                </span>
              </div>

              <div className="kpi-card" id="kpi-report-asset-investments">
                <span className="kpi-card__label">Investment Portfolio</span>
                <div className="kpi-card__value">{formatCurrency(report.assets.investments)}</div>
                <span className="kpi-card__subtext">
                  {report.assets.investment_list?.length || 0} holdings
                </span>
              </div>

              <div className="kpi-card" id="kpi-report-liabilities">
                <span className="kpi-card__label">Liabilities</span>
                <div className="kpi-card__value text-muted">{formatCurrency(report.liabilities)}</div>
                <span className="kpi-card__subtext">Debt tracking placeholder</span>
              </div>
            </div>

            {/* Asset Breakdown Detail Tables */}
            <div className="asset-breakdown-grid">
              {/* Accounts Breakdown */}
              <div className="card breakdown-card">
                <h4 className="breakdown-card__title">🏦 Cash Accounts</h4>
                {!report.assets.account_list || report.assets.account_list.length === 0 ? (
                  <p className="text-muted text-sm">No cash accounts available.</p>
                ) : (
                  <ul className="breakdown-list">
                    {report.assets.account_list.map((acc) => (
                      <li key={acc.id} className="breakdown-item">
                        <span>
                          <strong>{acc.name}</strong> <span className="badge badge--neutral">{acc.type}</span>
                        </span>
                        <span className="font-mono">{formatCurrency(acc.balance)}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              {/* Receivables Breakdown */}
              <div className="card breakdown-card">
                <h4 className="breakdown-card__title">🤝 Active Receivables</h4>
                {!report.assets.receivable_list || report.assets.receivable_list.length === 0 ? (
                  <p className="text-muted text-sm">No active receivables.</p>
                ) : (
                  <ul className="breakdown-list">
                    {report.assets.receivable_list.map((rec) => (
                      <li key={rec.id} className="breakdown-item">
                        <span>
                          <strong>{rec.counterparty}</strong>{' '}
                          <span className="badge badge--primary">{rec.status}</span>
                        </span>
                        <span className="font-mono">{formatCurrency(rec.remaining_amount)}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              {/* Investments Breakdown */}
              <div className="card breakdown-card">
                <h4 className="breakdown-card__title">📈 Investment Holdings</h4>
                {!report.assets.investment_list || report.assets.investment_list.length === 0 ? (
                  <p className="text-muted text-sm">No investment holdings.</p>
                ) : (
                  <ul className="breakdown-list">
                    {report.assets.investment_list.map((inv) => (
                      <li key={inv.id} className="breakdown-item">
                        <span>
                          <strong>{inv.name}</strong> <span className="badge badge--neutral">{inv.type}</span>
                        </span>
                        <span className="font-mono">{formatCurrency(inv.current_value)}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          </section>
        </div>
      )}

      {/* JSON Export Modal */}
      {showJsonModal && report && (
        <div className="modal-backdrop" onClick={() => setShowJsonModal(false)}>
          <div
            className="modal-card modal-card--lg"
            onClick={(e) => e.stopPropagation()}
            role="dialog"
            aria-labelledby="modal-json-title"
          >
            <div className="modal-header">
              <h3 id="modal-json-title" className="modal-title">
                📋 Report JSON Export
              </h3>
              <button
                className="btn-close"
                onClick={() => setShowJsonModal(false)}
                aria-label="Dismiss"
              >
                ✕
              </button>
            </div>

            <div className="modal-body">
              <pre className="json-viewer" id="report-json-content">
                {JSON.stringify(report, null, 2)}
              </pre>
            </div>

            <div className="modal-footer">
              <button
                id="btn-copy-report-json"
                className="btn btn--primary"
                onClick={handleCopyJson}
              >
                {copySuccess ? '✅ Copied to Clipboard!' : '📋 Copy JSON'}
              </button>
              <button
                id="btn-close-json-modal"
                className="btn btn--outline"
                onClick={() => setShowJsonModal(false)}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
