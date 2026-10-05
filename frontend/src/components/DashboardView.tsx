import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
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
  const [sidebarOpen, setSidebarOpen] = useState<boolean>(false);
  const [accounts, setAccounts] = useState<Account[]>([]);

  // Lock body scrolling on mobile when sidebar drawer is open
  useEffect(() => {
    if (sidebarOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [sidebarOpen]);

  const loadAccounts = async () => {
    try {
      const accs = await listAccounts();
      setAccounts(accs);
    } catch {
      // Non-blocking for header
    }
  };

  useEffect(() => {
    loadAccounts();
  }, []);

  const handleSelectTab = (tabId: ActiveNavTab) => {
    setActiveTab(tabId);
    setSidebarOpen(false);
    if (tabId === 'receivables' || tabId === 'investments') {
      loadAccounts();
    }
  };

  return (
    <div className="dashboard" id="dashboard-view">
      {/* Mobile Top Bar (Clean & Safe from Notches) */}
      <header className="mobile-header" id="mobile-top-header">
        <button
          type="button"
          className="mobile-hamburger-btn"
          id="mobile-sidebar-toggle"
          onClick={() => setSidebarOpen(true)}
          aria-label="Buka Menu Navigasi"
        >
          <span className="hamburger-icon">☰</span>
        </button>

        <div
          className="mobile-header-brand"
          onClick={() => handleSelectTab('overview')}
          style={{ cursor: 'pointer' }}
        >
          <img src="/icon-192.png" alt="FinTrack Logo" className="brand-logo-img" />
          <span className="brand-name">FinTrack</span>
        </div>

        <div className="mobile-header-actions">
          <PrivacyToggle />
        </div>
      </header>

      {/* Mobile Drawer Backdrop */}
      {sidebarOpen && (
        <div
          className="sidebar-backdrop"
          onClick={() => setSidebarOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* Responsive Sidebar (Desktop Permanent / Mobile Slide-Over Drawer) */}
      <aside
        className={`dashboard-sidebar ${sidebarOpen ? 'dashboard-sidebar--open' : ''}`}
        id="dashboard-sidebar"
      >
        {/* Sidebar Header */}
        <div className="sidebar-header">
          <div
            className="dashboard-brand"
            onClick={() => handleSelectTab('overview')}
            style={{ cursor: 'pointer' }}
          >
            <span className="brand-logo">
              <img src="/icon-192.png" alt="FinTrack Logo" className="brand-logo-img" />
            </span>
            <div className="brand-info">
              <span className="brand-name">FinTrack</span>
              <span className="brand-tag">Finance Hub</span>
            </div>
          </div>

          <button
            type="button"
            className="sidebar-close-btn"
            onClick={() => setSidebarOpen(false)}
            aria-label="Tutup Menu"
          >
            ✕
          </button>
        </div>

        {/* Sidebar Navigation Menu */}
        <nav className="sidebar-nav">
          <div className="sidebar-nav-title">MENU UTAMA</div>
          <div className="sidebar-nav-links" id="nav-tab-links">
            {NAV_TABS.map((tab) => (
              <button
                key={tab.id}
                className={`sidebar-nav-item ${activeTab === tab.id ? 'sidebar-nav-item--active' : ''}`}
                onClick={() => handleSelectTab(tab.id)}
                id={`nav-tab-${tab.id}`}
              >
                <span className="sidebar-nav-icon">{tab.icon}</span>
                <span className="sidebar-nav-label">{tab.label}</span>
                {activeTab === tab.id && <span className="sidebar-active-indicator" />}
              </button>
            ))}
          </div>
        </nav>

        {/* Sidebar Footer with Profile & Sign Out (Comfortable & Easily Reached on Mobile) */}
        <div className="sidebar-footer">
          <div className="sidebar-privacy-row">
            <span className="sidebar-privacy-label">Mode Privasi</span>
            <PrivacyToggle />
          </div>

          {/* User Profile Card */}
          <button
            type="button"
            className={`sidebar-user-card ${activeTab === 'settings' ? 'sidebar-user-card--active' : ''}`}
            id="user-info-pill"
            title="Klik untuk cek Profil & Atur PIN"
            onClick={() => handleSelectTab('settings')}
          >
            <div className="sidebar-user-avatar">👤</div>
            <div className="sidebar-user-meta">
              <span className="sidebar-user-email">{user?.email}</span>
              <span className="sidebar-user-badge">
                {user?.has_pin ? '🟢 PIN Aktif' : '⚪ Atur PIN Akun'}
              </span>
            </div>
            <span className="sidebar-user-action-arrow">⚙️</span>
          </button>

          {/* Large, Accessible Sign Out Button */}
          <button
            type="button"
            className="btn btn--outline dashboard-logout-btn sidebar-logout-btn"
            id="dashboard-logout-btn"
            onClick={() => {
              setSidebarOpen(false);
              logout();
            }}
          >
            <span className="logout-icon">🚪</span>
            <span>Sign Out</span>
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="dashboard-main-area">

      {/* Main Content Pane */}
      <main className="dashboard-content">
        {activeTab !== 'overview' && <CycleBar />}

        {activeTab === 'overview' && (
          <>
            <section className="profile-banner welcome-banner">
              <div className="welcome-banner-header">
                <div className="welcome-banner-text">
                  <div className="welcome-greeting-row">
                    <h2 className="banner-title">Welcome, {user?.email}!</h2>
                    <span className="badge badge--success welcome-badge">
                      🛡️ Akun Aktif
                    </span>
                  </div>
                  <p className="banner-subtitle">
                    Workspace finansial pribadi Anda untuk memantau arus kas, alokasi anggaran, dan portofolio aset secara real-time.
                  </p>
                </div>
                <button
                  type="button"
                  className="btn btn--outline btn--sm welcome-profile-btn"
                  onClick={() => handleSelectTab('settings')}
                  title="Buka Profil & Pengaturan Lengkap"
                >
                  ⚙️ Kelola Profil & PIN
                </button>
              </div>

              <div className="welcome-meta-chips">
                <div className="welcome-chip">
                  <span className="welcome-chip-icon">📅</span>
                  <span className="welcome-chip-label">Siklus Periode:</span>
                  <strong className="welcome-chip-val">Day {user?.cycle_start_day || 1} of each month</strong>
                </div>
                <div className="welcome-chip">
                  <span className="welcome-chip-icon">🔐</span>
                  <span className="welcome-chip-label">Keamanan PIN:</span>
                  <span className={user?.has_pin ? 'text-success' : 'text-muted'}>
                    {user?.has_pin ? '🟢 Aktif di Database' : '⚪ Belum Dikonfigurasi'}
                  </span>
                </div>
              </div>
            </section>

            <DashboardOverview onNavigateTab={(tab) => setActiveTab(tab as ActiveNavTab)} />
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
    </div>
  );
};

