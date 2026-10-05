import React from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
import { AuthForms } from './components/AuthForms';
import { DashboardView } from './components/DashboardView';
import { PrivacyProvider } from './context/PrivacyContext';
import './App.css';

const MainContent: React.FC = () => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="app-loading" id="app-loading">
        <div className="spinner" />
        <span>Loading FinTrack…</span>
      </div>
    );
  }

  if (isAuthenticated) {
    return <DashboardView />;
  }

  return (
    <div className="app" id="app-unauthenticated">
      <header className="app-header">
        <div className="app-logo">💰</div>
        <h1 className="app-title">FinTrack</h1>
        <p className="app-subtitle">Personal Finance Management System</p>
      </header>

      <AuthForms />

      <footer className="app-footer">
        FinTrack v0.2.0 — Phase 02 Authentication
      </footer>
    </div>
  );
};

function App() {
  return (
    <AuthProvider>
      <PrivacyProvider defaultHidden={true}>
        <MainContent />
      </PrivacyProvider>
    </AuthProvider>
  );
}

export default App;
