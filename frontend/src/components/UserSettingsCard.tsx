import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { PinSettingsCard } from './PinSettingsCard';
import { fetchHealth, type HealthResponse } from '../api/client';

export const UserSettingsCard: React.FC = () => {
  const { user, updateCycleStartDay } = useAuth();
  const [cycleDay, setCycleDay] = useState<number>(user?.cycle_start_day || 1);
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [successMessage, setSuccessMessage] = useState<string>('');
  const [errorMessage, setErrorMessage] = useState<string>('');

  // Database & Backend Health
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
      setHealthError(err instanceof Error ? err.message : 'Gagal memeriksa status database');
    } finally {
      setHealthLoading(false);
    }
  };

  useEffect(() => {
    loadHealth();
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSuccessMessage('');
    setErrorMessage('');

    if (cycleDay < 1 || cycleDay > 31) {
      setErrorMessage('Cycle start day must be between 1 and 31');
      return;
    }

    setIsSaving(true);
    try {
      await updateCycleStartDay(cycleDay);
      setSuccessMessage('Financial cycle start day updated successfully!');
      setTimeout(() => setSuccessMessage(''), 3000);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to update settings';
      setErrorMessage(msg);
    } finally {
      setIsSaving(false);
    }
  };

  const formattedDate = user?.created_at
    ? new Date(user.created_at).toLocaleDateString('id-ID', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    : 'Aktif';

  return (
    <div className="profile-settings-container">
      {/* 1. User Profile Information Card */}
      <div className="settings-card user-profile-card" id="user-profile-info-card">
        <div className="user-profile-card__header">
          <div className="user-profile-avatar-wrapper">
            <span className="user-profile-avatar">👤</span>
          </div>
          <div className="user-profile-heading">
            <h3 className="user-profile-title">Profil Pengguna</h3>
            <p className="user-profile-email">{user?.email}</p>
            <span className="badge badge--success" style={{ marginTop: '4px' }}>
              🛡️ Akun Aktif & Terverifikasi
            </span>
          </div>
        </div>

        <div className="user-profile-info-grid">
          <div className="profile-info-box">
            <span className="profile-info-label">Alamat Email</span>
            <span className="profile-info-value">{user?.email}</span>
          </div>
          <div className="profile-info-box">
            <span className="profile-info-label">Tanggal Bergabung</span>
            <span className="profile-info-value">{formattedDate}</span>
          </div>
          <div className="profile-info-box">
            <span className="profile-info-label">Siklus Keuangan</span>
            <span className="profile-info-value">Tanggal {user?.cycle_start_day || 1} setiap bulan</span>
          </div>
          <div className="profile-info-box">
            <span className="profile-info-label">Status Keamanan PIN</span>
            <span className={`profile-info-value ${user?.has_pin ? 'text-success' : 'text-muted'}`}>
              {user?.has_pin ? '🟢 PIN 6-Digit Aktif' : '⚪ PIN Belum Diatur'}
            </span>
          </div>
        </div>
      </div>

      {/* 2. PIN Security Settings Card */}
      <PinSettingsCard />

      {/* 3. Financial Cycle Settings Card */}
      <div className="settings-card" id="user-settings-card" style={{ marginTop: 'var(--space-4)' }}>
        <div className="settings-card__header">
          <h3 className="settings-card__title">⚙️ Financial Cycle Settings</h3>
          <p className="settings-card__subtitle">
            Configure the start day of your monthly accounting period. All budgeting and cashflow cycles align with this date.
          </p>
        </div>

        {successMessage && (
          <div className="auth-alert auth-alert--success" role="status" id="settings-success-banner">
            <span className="auth-alert__icon">✅</span>
            <span>{successMessage}</span>
          </div>
        )}

        {errorMessage && (
          <div className="auth-alert auth-alert--error" role="alert" id="settings-error-banner">
            <span className="auth-alert__icon">⚠️</span>
            <span>{errorMessage}</span>
          </div>
        )}

        <form className="settings-form" onSubmit={handleSave}>
          <div className="form-group">
            <label htmlFor="settings-cycle-day">Cycle Start Day (1 - 31)</label>
            <div className="settings-input-group">
              <input
                id="settings-cycle-day"
                type="number"
                min={1}
                max={31}
                value={cycleDay}
                onChange={(e) => setCycleDay(parseInt(e.target.value, 10) || 1)}
                disabled={isSaving}
              />
              <button
                type="submit"
                className="btn btn--primary"
                id="settings-save-btn"
                disabled={isSaving || cycleDay === user?.cycle_start_day}
              >
                {isSaving ? 'Saving…' : 'Save Changes'}
              </button>
            </div>
            <span className="form-hint">
              Current active cycle start day: <strong>Day {user?.cycle_start_day || 1}</strong> of each month.
            </span>
          </div>
        </form>
      </div>

      {/* 4. Backend & Database Health Diagnostic Card */}
      <section className="health-card" id="system-health-card" style={{ marginTop: 'var(--space-4)' }}>
        <div className="health-card__top">
          <div>
            <h3 className="health-card__title">⚡ Status Sistem & Koneksi Database</h3>
            <p className="health-card__subtitle" style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)', marginTop: '2px' }}>
              Pemeriksaan konektivitas real-time ke backend API dan database PostgreSQL Supabase
            </p>
          </div>
          <button
            className="btn-refresh"
            onClick={loadHealth}
            disabled={healthLoading}
            title="Perbarui status koneksi"
          >
            🔄 Refresh
          </button>
        </div>

        {healthLoading && (
          <div className="health-loading">
            <div className="spinner" />
            <span>Memeriksa koneksi layanan…</span>
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
              <span className="health-row__label">Backend API</span>
              <span className="health-row__value">
                <span className={`status-badge status-badge--${health.status}`}>
                  <span className="status-badge__dot" />
                  {health.status === 'healthy' ? 'Normal (Healthy)' : health.status}
                </span>
              </span>
            </div>
            <div className="health-row">
              <span className="health-row__label">Database PostgreSQL (Supabase)</span>
              <span className="health-row__value">
                <span
                  className={`status-badge ${
                    health.database === 'connected'
                      ? 'status-badge--healthy'
                      : 'status-badge--unhealthy'
                  }`}
                >
                  <span className="status-badge__dot" />
                  {health.database === 'connected' ? 'Terhubung (Connected)' : health.database}
                </span>
              </span>
            </div>
          </div>
        )}
      </section>
    </div>
  );
};

