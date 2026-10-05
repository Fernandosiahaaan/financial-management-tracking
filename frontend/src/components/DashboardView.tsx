import React, { useEffect, useState, useRef } from 'react';
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

const NAV_TABS = [
  { id: 'overview' as const, label: 'Dashboard', icon: '📊' },
  { id: 'reports' as const, label: 'Reports', icon: '📑' },
  { id: 'transactions' as const, label: 'Transactions', icon: '💸' },
  { id: 'budgets' as const, label: 'Budgets', icon: '📊' },
  { id: 'allocations' as const, label: 'Allocations', icon: '🎯' },
  { id: 'receivables' as const, label: 'Receivables', icon: '🤝' },
  { id: 'investments' as const, label: 'Investments', icon: '📈' },
  { id: 'accounts' as const, label: 'Accounts', icon: '🏛️' },
  { id: 'categories' as const, label: 'Categories', icon: '🏷️' },
  { id: 'settings' as const, label: 'Profile & Settings', icon: '⚙️' },
];

export const DashboardView: React.FC = () => {
  const { user, logout } = useAuth();
  const [activeTab, setActiveTab] = useState<ActiveNavTab>('overview');
  const [showMobileMenu, setShowMobileMenu] = useState<boolean>(false);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [healthLoading, setHealthLoading] = useState<boolean>(true);
  const [healthError, setHealthError] = useState<string>('');
  const tabLinksRef = useRef<HTMLDivElement>(null);

  // Auto-scroll active tab into view smoothly
  useEffect(() => {
    const activeEl = document.getElementById(`nav-tab-${activeTab}`);
    if (activeEl && typeof activeEl.scrollIntoView === 'function') {
      activeEl.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' });
    }
  }, [activeTab]);

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
        <div className="dashboard-nav-top">
          <div className="dashboard-brand" onClick={() => setActiveTab('overview')} style={{ cursor: 'pointer' }}>
            <span className="brand-logo">
              <img src="/icon-192.png" alt="FinTrack Logo" className="brand-logo-img" />
            </span>
            <span className="brand-name">FinTrack</span>
          </div>

          <div className="dashboard-user-actions">
            <PrivacyToggle />
            <button
              type="button"
              className="user-pill"
              id="user-info-pill"
              title="Klik untuk cek Profile & atur PIN"
              onClick={() => setActiveTab('settings')}
              style={{ cursor: 'pointer', background: activeTab === 'settings' ? 'var(--color-primary-light, rgba(16, 185, 129, 0.15))' : undefined }}
            >
              <span className="user-avatar">👤</span>
              <span className="user-email">{user?.email}</span>
            </button>
            <button
              className="btn btn--outline btn--sm dashboard-logout-btn"
              id="dashboard-logout-btn"
              onClick={logout}
            >
              Sign Out
            </button>
          </div>
        </div>

        {/* Feature Tabs Bar */}
        <div className="nav-tab-wrapper">
          <div className="nav-tab-links" ref={tabLinksRef} id="nav-tab-links">
            {NAV_TABS.map((tab) => (
              <button
                key={tab.id}
                className={`nav-tab-btn ${activeTab === tab.id ? 'nav-tab-btn--active' : ''}`}
                onClick={() => {
                  setActiveTab(tab.id);
                  if (tab.id === 'receivables' || tab.id === 'investments') {
                    loadAccounts();
                  }
                }}
                id={`nav-tab-${tab.id}`}
              >
                <span>{tab.icon}</span>
                <span>{tab.label}</span>
              </button>
            ))}
          </div>

          {/* Quick-Access Mobile All Features Button */}
          <button
            className={`nav-mobile-menu-btn ${showMobileMenu ? 'nav-mobile-menu-btn--active' : ''}`}
            id="nav-mobile-menu-toggle"
            onClick={() => setShowMobileMenu((prev) => !prev)}
            aria-label="Tampilkan Semua Fitur"
          >
            <span>{showMobileMenu ? '✕' : '🗂️'}</span>
            <span className="nav-mobile-menu-text">Semua Fitur</span>
            <span className="nav-mobile-menu-count">{NAV_TABS.length}</span>
          </button>
        </div>

        {/* Mobile All Features Dropdown/Sheet */}
        {showMobileMenu && (
          <div className="mobile-features-drawer">
            <div className="mobile-features-drawer__header">
              <span className="mobile-features-drawer__title">🗂️ Navigasi Semua Fitur</span>
              <span className="mobile-features-drawer__sub">Pilih menu untuk langsung berpindah</span>
            </div>
            <div className="mobile-features-grid">
              {NAV_TABS.map((tab) => (
                <button
                  key={tab.id}
                  className={`mobile-feature-card ${activeTab === tab.id ? 'mobile-feature-card--active' : ''}`}
                  onClick={() => {
                    setActiveTab(tab.id);
                    if (tab.id === 'receivables' || tab.id === 'investments') {
                      loadAccounts();
                    }
                    setShowMobileMenu(false);
                  }}
                >
                  <span className="mobile-feature-card__icon">{tab.icon}</span>
                  <span className="mobile-feature-card__label">{tab.label}</span>
                  {activeTab === tab.id && <span className="mobile-feature-card__badge">Aktif</span>}
                </button>
              ))}
            </div>
          </div>
        )}
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
