import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { fetchHealth, type HealthResponse } from '../api/client';
import { UserSettingsCard } from './UserSettingsCard';
import { AccountsManager } from './AccountsManager';
import { CategoriesManager } from './CategoriesManager';
import { TransactionsManager } from './TransactionsManager';
import { BudgetManager } from './BudgetManager';
import { AllocationManager } from './AllocationManager';
import { DashboardOverview } from './DashboardOverview';
import { CycleBar } from './CycleBar';
import { ReceivablesManager } from './ReceivablesManager';
import { InvestmentsManager } from './InvestmentsManager';
import { ReportsManager } from './ReportsManager';
import { PrivacyToggle } from './PrivacyToggle';
import { listAccounts, type Account } from '../api/accounts';

type ActiveNavTab =
  | 'overview'
  | 'reports'
  | 'transactions'
  | 'budgets'
  | 'allocations'
  | 'receivables'
  | 'investments'
  | 'accounts'
  | 'categories'
  | 'settings';

export const DashboardView: React.FC = () => {
  const { user, logout } = useAuth();
  const [activeTab, setActiveTab] = useState<ActiveNavTab>('overview');
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [healthLoading, setHealthLoading] = useState<boolean>(true);
  const [healthError, setHealthError] = useState<string>('');

  const loadHealth = async () => {
    setHealthLoading(true);
    setHealthError('');
    try {
      const data = await fetchHealth();
      setHealth(data);
    } catch (err) {
      setHealthError(err instanceof Error ? err.message : 'System health check failed');
    } finally {
      setHealthLoading(false);
    }
  };

  const loadAccounts = async () => {
    try {
      const accs = await listAccounts();
      setAccounts(accs);
    } catch {
      // Non-blocking for header
    }
  };

  useEffect(() => {
    loadHealth();
    loadAccounts();
  }, []);

  return (
    <div className="dashboard" id="dashboard-view">
      {/* Top Navigation Bar */}
      <nav className="dashboard-nav">
        <div className="dashboard-brand">
          <span className="brand-logo">💰</span>
          <span className="brand-name">FinTrack</span>
        </div>

        <div className="nav-tab-links">
          <button
            className={`nav-tab-btn ${activeTab === 'overview' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => setActiveTab('overview')}
            id="nav-tab-overview"
          >
            📊 Dashboard
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'reports' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => setActiveTab('reports')}
            id="nav-tab-reports"
          >
            📑 Reports
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'transactions' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => setActiveTab('transactions')}
            id="nav-tab-transactions"
          >
            💸 Transactions
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'budgets' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => setActiveTab('budgets')}
            id="nav-tab-budgets"
          >
            📊 Budgets
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'allocations' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => setActiveTab('allocations')}
            id="nav-tab-allocations"
          >
            🎯 Allocations
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'receivables' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => {
              setActiveTab('receivables');
              loadAccounts();
            }}
            id="nav-tab-receivables"
          >
            🤝 Receivables
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'investments' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => {
              setActiveTab('investments');
              loadAccounts();
            }}
            id="nav-tab-investments"
          >
            📈 Investments
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'accounts' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => setActiveTab('accounts')}
            id="nav-tab-accounts"
          >
            🏛️ Accounts
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'categories' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => setActiveTab('categories')}
            id="nav-tab-categories"
          >
            🏷️ Categories
          </button>
          <button
            className={`nav-tab-btn ${activeTab === 'settings' ? 'nav-tab-btn--active' : ''}`}
            onClick={() => setActiveTab('settings')}
            id="nav-tab-settings"
          >
            ⚙️ Settings
          </button>
        </div>

        <div className="dashboard-user-actions">
          <PrivacyToggle />
          <div className="user-pill" id="user-info-pill">
            <span className="user-avatar">👤</span>
            <span className="user-email">{user?.email}</span>
          </div>
          <button
            className="btn btn--outline btn--sm"
            id="dashboard-logout-btn"
            onClick={logout}
          >
            Sign Out
          </button>
        </div>
      </nav>

      {/* Main Content Pane */}
      <main className="dashboard-content">
        {activeTab !== 'overview' && <CycleBar />}

        {activeTab === 'overview' && (
          <>
            <DashboardOverview onNavigateTab={(tab) => setActiveTab(tab as ActiveNavTab)} />

            <section className="profile-banner">
              <h2 className="banner-title">Welcome, {user?.email}!</h2>
              <p className="banner-subtitle">
                Your personal finance workspace is secured and authenticated.
              </p>
              <div className="profile-details-grid">
                <div className="detail-item">
                  <span className="detail-label">Account ID</span>
                  <code className="detail-code">{user?.id}</code>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Cycle Start Day</span>
                  <span className="detail-value">Day {user?.cycle_start_day} of each month</span>
                </div>
              </div>
            </section>

            {/* User Settings */}
            <UserSettingsCard />

            {/* System Health Status */}
            <section className="health-card" id="system-health-card">
              <div className="health-card__top">
                <h3 className="health-card__title">⚡ Backend & Database Health</h3>
                <button
                  className="btn-refresh"
                  onClick={loadHealth}
                  disabled={healthLoading}
                  title="Refresh health status"
                >
                  🔄 Refresh
                </button>
              </div>

              {healthLoading && (
                <div className="health-loading">
                  <div className="spinner" />
                  <span>Verifying service connectivity…</span>
                </div>
              )}

              {healthError && (
                <div className="auth-alert auth-alert--error">
                  <span>⚠️ {healthError}</span>
                </div>
              )}

              {!healthLoading && !healthError && health && (
                <div className="health-card__rows">
                  <div className="health-row">
                    <span className="health-row__label">API Status</span>
                    <span className="health-row__value">
                      <span className={`status-badge status-badge--${health.status}`}>
                        <span className="status-badge__dot" />
                        {health.status}
                      </span>
                    </span>
                  </div>
                  <div className="health-row">
                    <span className="health-row__label">PostgreSQL</span>
                    <span className="health-row__value">
                      <span
                        className={`status-badge ${
                          health.database === 'connected'
                            ? 'status-badge--healthy'
                            : 'status-badge--unhealthy'
                        }`}
                      >
                        <span className="status-badge__dot" />
                        {health.database}
                      </span>
                    </span>
                  </div>
                </div>
              )}
            </section>
          </>
        )}

        {activeTab === 'reports' && <ReportsManager />}
        {activeTab === 'transactions' && <TransactionsManager />}
        {activeTab === 'budgets' && <BudgetManager />}
        {activeTab === 'allocations' && <AllocationManager />}
        {activeTab === 'receivables' && <ReceivablesManager accounts={accounts} />}
        {activeTab === 'investments' && <InvestmentsManager accounts={accounts} />}
        {activeTab === 'accounts' && <AccountsManager />}
        {activeTab === 'categories' && <CategoriesManager />}
        {activeTab === 'settings' && <UserSettingsCard />}
      </main>
    </div>
  );
};
