import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
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
import {
  LayoutDashboard,
  FileText,
  ArrowLeftRight,
  PieChart,
  Target,
  HandCoins,
  TrendingUp,
  Landmark,
  Tags,
  Settings,
  LogOut,
  Sun,
  Moon,
  Menu,
  X,
  Sparkles,
  Wallet,
} from 'lucide-react';

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

interface NavItemConfig {
  id: ActiveNavTab;
  label: string;
  badge?: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  elementId: string;
}

const NAV_GROUPS: { group: string; items: NavItemConfig[] }[] = [
  {
    group: 'Ringkasan & Analisis',
    items: [
      { id: 'overview', label: 'Dashboard', icon: LayoutDashboard, elementId: 'nav-tab-overview' },
      { id: 'reports', label: 'Reports', icon: FileText, elementId: 'nav-tab-reports' },
    ],
  },
  {
    group: 'Arus Kas & Anggaran',
    items: [
      { id: 'transactions', label: 'Transactions', icon: ArrowLeftRight, elementId: 'nav-tab-transactions' },
      { id: 'budgets', label: 'Budgets', icon: PieChart, elementId: 'nav-tab-budgets' },
      { id: 'allocations', label: 'Allocations', icon: Target, elementId: 'nav-tab-allocations' },
    ],
  },
  {
    group: 'Aset & Portofolio',
    items: [
      { id: 'receivables', label: 'Receivables', icon: HandCoins, elementId: 'nav-tab-receivables' },
      { id: 'investments', label: 'Investments', icon: TrendingUp, elementId: 'nav-tab-investments' },
      { id: 'accounts', label: 'Accounts', icon: Landmark, elementId: 'nav-tab-accounts' },
      { id: 'categories', label: 'Categories', icon: Tags, elementId: 'nav-tab-categories' },
    ],
  },
  {
    group: 'Sistem',
    items: [
      { id: 'settings', label: 'Profile & Settings', icon: Settings, elementId: 'nav-tab-settings' },
    ],
  },
];

export const DashboardView: React.FC = () => {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const [activeTab, setActiveTab] = useState<ActiveNavTab>('overview');
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [healthLoading, setHealthLoading] = useState<boolean>(true);
  const [healthError, setHealthError] = useState<string>('');
  const [mobileMenuOpen, setMobileMenuOpen] = useState<boolean>(false);

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

  useEffect(() => {
    if (mobileMenuOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [mobileMenuOpen]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && mobileMenuOpen) {
        setMobileMenuOpen(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [mobileMenuOpen]);

  const handleNavClick = (tabId: ActiveNavTab) => {
    setActiveTab(tabId);
    setMobileMenuOpen(false);
    if (tabId === 'receivables' || tabId === 'investments') {
      loadAccounts();
    }
  };

  // Human-readable titles for the top bar
  const getTabTitle = (tab: ActiveNavTab) => {
    switch (tab) {
      case 'overview': return 'Financial Overview';
      case 'reports': return 'Financial Reports & Snapshots';
      case 'transactions': return 'Transaction Ledger';
      case 'budgets': return 'Budget Planning & Control';
      case 'allocations': return 'Income Allocation Strategy';
      case 'receivables': return 'Receivables & Loan Tracker';
      case 'investments': return 'Investment Portfolio';
      case 'accounts': return 'Account & Liquidity Balances';
      case 'categories': return 'Transaction Categories';
      case 'settings': return 'User Profile & Preferences';
      default: return 'FinTrack Workspace';
    }
  };

  return (
    <div className="dashboard-app-layout" id="dashboard-view">
      {/* Mobile Backdrop */}
      {mobileMenuOpen && (
        <div
          className="sidebar-backdrop"
          onClick={() => setMobileMenuOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* ── Modern Collapsible Sidebar ─────────────────────────────── */}
      <aside className={`dashboard-sidebar ${mobileMenuOpen ? 'dashboard-sidebar--open' : ''}`}>
        {/* Brand Header */}
        <div className="sidebar-brand-wrapper">
          <div className="sidebar-brand">
            <div className="brand-icon-box">
              <Wallet size={20} className="brand-icon-svg" />
              <div className="brand-glow-ring" />
            </div>
            <div className="brand-text-block">
              <div className="brand-title-row">
                <span className="brand-title">FinTrack</span>
                <span className="brand-pill">v0.2.0</span>
              </div>
              <span className="brand-tagline">Financial Management</span>
            </div>
          </div>
          <button
            className="sidebar-mobile-close"
            onClick={() => setMobileMenuOpen(false)}
            aria-label="Close menu"
          >
            <X size={18} />
          </button>
        </div>

        {/* Navigation Groups */}
        <div className="sidebar-scroll-area">
          <nav className="sidebar-nav-groups dashboard-nav">
            {NAV_GROUPS.map((grp) => (
              <div key={grp.group} className="sidebar-group">
                <span className="sidebar-group-label">{grp.group}</span>
                <div className="sidebar-group-items nav-tab-links">
                  {grp.items.map((item) => {
                    const IconComponent = item.icon;
                    const isActive = activeTab === item.id;
                    return (
                      <button
                        key={item.id}
                        id={item.elementId}
                        className={`nav-tab-btn ${isActive ? 'nav-tab-btn--active' : ''}`}
                        onClick={() => handleNavClick(item.id)}
                      >
                        <span className="nav-tab-icon-wrapper">
                          <IconComponent size={17} className="nav-tab-lucide-icon" />
                        </span>
                        <span className="nav-tab-label-text">{item.label}</span>
                        {isActive && <span className="nav-tab-active-dot" />}
                      </button>
                    );
                  })}
                </div>
              </div>
            ))}
          </nav>
        </div>

        {/* Sidebar Footer User Card */}
        <div className="sidebar-footer">
          <div className="sidebar-user-card" id="user-info-pill">
            <div className="user-avatar-circle">
              <span className="user-avatar-initials">
                {user?.email ? user.email.slice(0, 2).toUpperCase() : 'FT'}
              </span>
              <span className="user-online-dot" />
            </div>
            <div className="user-details-column">
              <span className="user-email-text user-email" title={user?.email}>
                {user?.email}
              </span>
              <span className="user-cycle-caption">
                Cycle Day {user?.cycle_start_day || 1}
              </span>
            </div>
          </div>

          <div className="sidebar-action-bar dashboard-user-actions">
            <button
              className="sidebar-action-btn"
              onClick={toggleTheme}
              title={`Switch to ${theme === 'dark' ? 'Light' : 'Dark'} mode`}
              aria-label="Toggle theme"
            >
              {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
              <span className="sidebar-action-label">
                {theme === 'dark' ? 'Light Mode' : 'Dark Mode'}
              </span>
            </button>

            <button
              className="btn btn--outline btn--sm sidebar-logout-btn"
              id="dashboard-logout-btn"
              onClick={logout}
              title="Sign Out"
            >
              <LogOut size={15} />
              <span>Sign Out</span>
            </button>
          </div>
        </div>
      </aside>

      {/* ── Main Workspace Body ───────────────────────────────────── */}
      <div className="dashboard-main-wrapper">
        {/* Top Header Bar */}
        <header className="dashboard-top-bar">
          <div className="top-bar-left">
            <button
              className="mobile-hamburger-btn"
              onClick={() => setMobileMenuOpen(true)}
              aria-label="Open navigation menu"
            >
              <Menu size={20} />
            </button>
            <div className="top-bar-heading">
              <div className="top-bar-breadcrumbs">
                <span className="top-bar-breadcrumb-item">Workspace</span>
                <span className="top-bar-breadcrumb-sep">/</span>
                <span className="top-bar-breadcrumb-active">{getTabTitle(activeTab)}</span>
              </div>
              <h1 className="top-bar-title">{getTabTitle(activeTab)}</h1>
            </div>
          </div>

          <div className="top-bar-right">
            {/* Live System Health Badge */}
            <div
              className={`top-bar-health-pill ${
                health?.status === 'healthy' ? 'top-bar-health-pill--healthy' : 'top-bar-health-pill--warning'
              }`}
              title={healthError ? `Error: ${healthError}` : `API: ${health?.status || 'checking'}, DB: ${health?.database || 'checking'}`}
            >
              <span className="health-pulse-dot" />
              <span className="health-status-text">
                {healthLoading ? 'Connecting…' : health?.status === 'healthy' ? 'System Online' : 'Check API'}
              </span>
            </div>

            {/* Privacy Toggle Button */}
            <PrivacyToggle />
          </div>
        </header>

        {/* Content Pane */}
        <main className="dashboard-content">
          {activeTab !== 'overview' && <CycleBar />}

          {activeTab === 'overview' && (
            <>
              {/* Profile & Cycle Welcome Banner at the very top */}
              <section className="profile-banner">
                <div className="profile-banner__glow" />
                <div className="profile-banner__header">
                  <div className="profile-banner__badge">
                    <Sparkles size={16} />
                    <span>Active Session</span>
                  </div>
                  <h2 className="banner-title">Welcome, {user?.email}!</h2>
                  <p className="banner-subtitle">
                    Your personal finance workspace is secured and authenticated.
                  </p>
                </div>
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

              <DashboardOverview onNavigateTab={(tab) => handleNavClick(tab as ActiveNavTab)} />
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
